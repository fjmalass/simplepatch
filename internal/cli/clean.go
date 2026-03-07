package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fjmalass/simplepatch/internal/log"
	"github.com/fjmalass/simplepatch/internal/operations"
)

// CleanCmd represents the clean subcommand
type CleanCmd struct {
	PatchesRoot string `name:"patches-root" short:"p" default:"." help:"Directory containing backup files"`
	DryRun      bool   `name:"dry-run" short:"d" help:"Show what would be deleted without deleting"`
	All         bool   `name:"all" short:"a" help:"Remove all backups (searches recursively)"`
}

// Run executes the clean command
func (c *CleanCmd) Run(logger log.Logger) error {
	return PerformClean(c, logger)
}

// PerformClean removes backup files from the patched directory
func PerformClean(cfg *CleanCmd, logger log.Logger) error {
	logger.Debug("starting clean", "config", fmt.Sprintf("%+v", cfg))

	styles := newStyles()

	// Print banner
	fmt.Println(styles.Title.Render("  Clean Backup Files"))
	fmt.Println(styles.Dim.Render("=================================================="))
	fmt.Println()

	// Resolve patches root to absolute path
	patchesRoot, err := filepath.Abs(cfg.PatchesRoot)
	if err != nil {
		logger.Error("failed to resolve patches root", "path", cfg.PatchesRoot, "err", err)
		return fmt.Errorf("failed to resolve patches root: %w", err)
	}

	// Validate patches root exists
	logger.Debug("validating patches root", "path", patchesRoot)
	if info, err := os.Stat(patchesRoot); err != nil || !info.IsDir() {
		logger.Error("patches root not found", "path", patchesRoot, "err", err)
		return fmt.Errorf("patches root not found at %s", patchesRoot)
	}

	if cfg.DryRun {
		logger.Info("dry-run mode enabled")
		fmt.Println(styles.Header.Render("[DRY-RUN] No files will be deleted"))
		fmt.Println()
	}

	// Find backup files
	logger.Info("searching for backup files", "dir", patchesRoot)
	backups, err := operations.FindBackupFiles(patchesRoot, logger)
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		fmt.Println(styles.Dim.Render("No backup files found."))
		fmt.Println()
		return nil
	}

	// Calculate total size
	totalSize := operations.GetBackupFileSize(backups, logger)

	// Delete or report
	if cfg.DryRun {
		fmt.Println(styles.Header.Render("Would delete the following backup files:"))
	} else {
		fmt.Println(styles.Header.Render("Deleting backup files..."))
	}

	deletedCount := 0
	var errors []error

	for _, backup := range backups {
		relPath, _ := filepath.Rel(patchesRoot, backup)
		if relPath == "" {
			relPath = backup
		}

		if cfg.DryRun {
			fmt.Printf("  %s %s\n",
				styles.Dim.Render("[DRY-RUN]"),
				styles.Path.Render(relPath))
		} else {
			err := operations.DeleteBackupFile(backup, false, logger)
			if err != nil {
				fmt.Printf("  %s: %s\n",
					styles.Error.Render("Failed"),
					styles.Path.Render(relPath))
				errors = append(errors, err)
			} else {
				fmt.Printf("  %s: %s\n",
					styles.Success.Render("Deleted"),
					styles.Path.Render(relPath))
				deletedCount++
			}
		}
	}

	// Summary
	fmt.Println()
	if cfg.DryRun {
		fmt.Printf("Found %s backup files (%s would be freed)\n",
			styles.Header.Render(fmt.Sprintf("%d", len(backups))),
			styles.Header.Render(operations.FormatBytes(totalSize)))
	} else {
		if len(errors) > 0 {
			fmt.Printf("%s: Cleaned %d files, %d errors\n",
				styles.Error.Render("Completed with errors"),
				deletedCount,
				len(errors))
		} else {
			fmt.Printf("%s: Cleaned %d backup files (%s freed)\n",
				styles.Success.Render("Done"),
				deletedCount,
				operations.FormatBytes(totalSize))
		}
	}
	fmt.Println()

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete %d files", len(errors))
	}
	return nil
}
