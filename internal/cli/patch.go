package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fjmalass/simplepatch/internal/log"
	"github.com/fjmalass/simplepatch/internal/operations"
	"github.com/fjmalass/simplepatch/internal/patch"
)

func PerformPatch(cfg *Config, logger log.Logger) error {
	logger.Debug("starting patch", "config", fmt.Sprintf("%+v", cfg))

	styles := newStyles()

	// Print banner
	fmt.Println(styles.Title.Render("  Copy Patched Files (Cross-Platform)"))
	fmt.Println(styles.Dim.Render("=================================================="))
	fmt.Println()

	// Validate original root exists
	logger.Debug("validating original root", "path", cfg.OriginalRoot)
	if info, err := os.Stat(cfg.OriginalRoot); err != nil || !info.IsDir() {
		logger.Error("original root not found", "path", cfg.OriginalRoot, "err", err)
		return fmt.Errorf("original root not found at %s", cfg.OriginalRoot)
	}
	logger.Debug("original root validated")

	// Validate patched root exists
	logger.Debug("validating patched root", "path", cfg.PatchesRoot)
	if info, err := os.Stat(cfg.PatchesRoot); err != nil || !info.IsDir() {
		logger.Error("patched root not found", "path", cfg.PatchesRoot, "err", err)
		return fmt.Errorf("patched root not found at %s", cfg.PatchesRoot)
	}
	logger.Debug("patched root validated")

	// Load mappings
	logger.Info("loading mappings", "path", cfg.MappingPath)
	mappings, err := patch.Load(cfg.MappingPath, logger)
	if err != nil {
		logger.Error("failed to load mappings", "err", err)
		return err
	}
	logger.Info("loaded mappings", "count", len(mappings))

	// Resolve relative paths to absolute
	for i := range mappings {
		mappings[i].Patched = filepath.Join(cfg.PatchesRoot, mappings[i].Patched)
		mappings[i].Original = filepath.Join(cfg.OriginalRoot, mappings[i].Original)
		logger.Debug("resolved mapping",
			"patched", mappings[i].Patched,
			"original", mappings[i].Original)
	}

	// Generate timestamp for backups
	timestamp := operations.CurrentTimestamp()
	logger.Debug("generated timestamp", "timestamp", timestamp)

	if cfg.DryRun {
		logger.Info("dry-run mode enabled")
		fmt.Println(styles.Header.Render("[DRY-RUN] No files will be modified"))
		fmt.Println()
	}

	// Phase 1: Backup existing originals
	if cfg.Backup {
		logger.Info("starting backup phase")
		fmt.Println(styles.Header.Render("Backing up existing files (if they exist)..."))

		for _, m := range mappings {
			backupDir := filepath.Dir(m.Patched)
			err := operations.BackupOriginal(m.Original, backupDir, timestamp, cfg.DryRun, logger)
			if err != nil {
				fmt.Println(styles.Error.Render(fmt.Sprintf("  Failed to backup %s: %v", m.Original, err)))
			}
		}
		fmt.Println()
	}

	// Phase 2: Copy patched -> original
	logger.Info("starting copy phase")
	fmt.Println(styles.Header.Render("Copying patched files to original locations..."))

	success := true
	for _, m := range mappings {
		err := operations.CopyPatchedToOriginal(m.Patched, m.Original, cfg.DryRun, logger)
		if err != nil {
			fmt.Println(styles.Error.Render(fmt.Sprintf("  Failed: %v", err)))
			success = false
		} else if !cfg.DryRun {
			fmt.Printf("  %s: %s -> %s\n",
				styles.Success.Render("Copied"),
				styles.Path.Render(m.Patched),
				styles.Path.Render(m.Original))
		}
	}

	// Summary
	fmt.Println()
	if success {
		logger.Info("patch completed successfully")
		fmt.Println(styles.Success.Render("Files patched successfully."))
		fmt.Println("\nYou can now edit:")
		for _, m := range mappings {
			fmt.Printf("  %s\n", styles.Path.Render(m.Patched))
		}
	} else {
		logger.Error("patch completed with errors")
		fmt.Println(styles.Error.Render("Some errors occurred - check above."))
	}
	fmt.Println()

	return nil
}
