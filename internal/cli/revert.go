package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fjmalass/simplepatch/internal/log"
	"github.com/fjmalass/simplepatch/internal/operations"
	"github.com/fjmalass/simplepatch/internal/patch"
)

// RevertCmd represents the revert subcommand
type RevertCmd struct {
	MappingPath  string `name:"map" short:"m" default:"patched_files.map" help:"Mappings file"`
	PatchesRoot  string `name:"patches-root" short:"p" default:"." help:"Where backup files are located"`
	OriginalRoot string `name:"original-root" short:"o" required:"" help:"Where to restore original files"`
	DryRun       bool   `name:"dry-run" short:"d" help:"Show what would be reverted without making changes"`
	UseLatest    bool   `name:"use-latest" help:"Use newest backup instead of oldest"`
	Cleanup      bool   `name:"cleanup" short:"c" help:"Delete backup files after successful revert"`
}

// Run executes the revert command
func (c *RevertCmd) Run(logger log.Logger) error {
	// Resolve defaults similar to config.go
	if err := c.resolveDefaults(); err != nil {
		return err
	}
	return PerformRevert(c, logger)
}

// resolveDefaults resolves relative paths to absolute
func (c *RevertCmd) resolveDefaults() error {
	var err error

	if c.PatchesRoot == "" || c.PatchesRoot == "." {
		c.PatchesRoot, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	c.PatchesRoot, err = filepath.Abs(c.PatchesRoot)
	if err != nil {
		return err
	}

	c.OriginalRoot, err = filepath.Abs(c.OriginalRoot)
	if err != nil {
		return err
	}

	// Make absolute mapping path
	// If it's the default value, look in patches root
	// Otherwise, resolve relative to current working directory
	if !filepath.IsAbs(c.MappingPath) {
		if c.MappingPath == "patched_files.map" {
			// Default value - look in patches root
			c.MappingPath = filepath.Join(c.PatchesRoot, c.MappingPath)
		} else {
			// User provided explicit path - resolve from cwd
			c.MappingPath, err = filepath.Abs(c.MappingPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// PerformRevert restores original files from backups
func PerformRevert(cfg *RevertCmd, logger log.Logger) error {
	logger.Debug("starting revert", "config", fmt.Sprintf("%+v", cfg))

	styles := newStyles()

	// Print banner
	fmt.Println(styles.Title.Render("  Revert from Backups"))
	fmt.Println(styles.Dim.Render("=================================================="))
	fmt.Println()

	// Validate patches root exists
	logger.Debug("validating patches root", "path", cfg.PatchesRoot)
	if info, err := os.Stat(cfg.PatchesRoot); err != nil || !info.IsDir() {
		logger.Error("patches root not found", "path", cfg.PatchesRoot, "err", err)
		return fmt.Errorf("patches root not found at %s", cfg.PatchesRoot)
	}

	// Validate original root exists
	logger.Debug("validating original root", "path", cfg.OriginalRoot)
	if info, err := os.Stat(cfg.OriginalRoot); err != nil || !info.IsDir() {
		logger.Error("original root not found", "path", cfg.OriginalRoot, "err", err)
		return fmt.Errorf("original root not found at %s", cfg.OriginalRoot)
	}

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

	if cfg.DryRun {
		logger.Info("dry-run mode enabled")
		fmt.Println(styles.Header.Render("[DRY-RUN] No files will be modified"))
		fmt.Println()
	}

	// Process each mapping
	if cfg.UseLatest {
		fmt.Println(styles.Header.Render("Reverting from newest backups..."))
	} else {
		fmt.Println(styles.Header.Render("Reverting from oldest backups..."))
	}

	revertedCount := 0
	skippedCount := 0
	var errors []error

	for _, m := range mappings {
		relPatched, _ := filepath.Rel(cfg.PatchesRoot, m.Patched)
		relOriginal, _ := filepath.Rel(cfg.OriginalRoot, m.Original)

		// Find backups for this file
		backups, err := operations.FindBackupsForFile(m.Patched, logger)
		if err != nil {
			logger.Warn("failed to find backups", "patched", m.Patched, "err", err)
			fmt.Printf("  %s: %s - %v\n",
				styles.Error.Render("Error"),
				styles.Path.Render(relPatched),
				err)
			errors = append(errors, err)
			continue
		}

		if len(backups) == 0 {
			logger.Warn("no backup found", "patched", m.Patched)
			fmt.Printf("  %s: No backup found for %s\n",
				styles.Dim.Render("Warning"),
				styles.Path.Render(relPatched))
			skippedCount++
			continue
		}

		// Select backup (oldest or newest)
		backup := operations.SelectBackup(backups, cfg.UseLatest)
		if backup == nil {
			skippedCount++
			continue
		}

		relBackup, _ := filepath.Rel(cfg.PatchesRoot, backup.Path)

		if cfg.DryRun {
			fmt.Printf("  %s: %s -> %s\n",
				styles.Dim.Render("[DRY-RUN]"),
				styles.Path.Render(relBackup),
				styles.Path.Render(relOriginal))

			if cfg.Cleanup {
				fmt.Printf("           %s would delete: %s\n",
					styles.Dim.Render("(cleanup)"),
					styles.Path.Render(relBackup))
			}
			revertedCount++
		} else {
			// Perform the revert
			err := operations.RevertFile(*backup, m.Original, false, logger)
			if err != nil {
				fmt.Printf("  %s: %s - %v\n",
					styles.Error.Render("Failed"),
					styles.Path.Render(relPatched),
					err)
				errors = append(errors, err)
				continue
			}

			fmt.Printf("  %s: %s (backup: %s)\n",
				styles.Success.Render("Reverted"),
				styles.Path.Render(relOriginal),
				backup.Time.Format("2006-01-02 15:04:05"))

			revertedCount++

			// Cleanup if requested
			if cfg.Cleanup {
				err := operations.DeleteBackup(*backup, false, logger)
				if err != nil {
					fmt.Printf("    %s: Failed to delete backup - %v\n",
						styles.Error.Render("Warning"),
						err)
				} else {
					fmt.Printf("    %s: %s\n",
						styles.Dim.Render("Cleaned"),
						styles.Path.Render(relBackup))
				}
			}
		}
	}

	// Summary
	fmt.Println()
	if cfg.DryRun {
		fmt.Printf("Would revert %s files",
			styles.Header.Render(fmt.Sprintf("%d", revertedCount)))
	} else {
		fmt.Printf("%s reverted %s files",
			styles.Success.Render("Successfully"),
			styles.Header.Render(fmt.Sprintf("%d", revertedCount)))
	}

	if skippedCount > 0 {
		fmt.Printf(" (%s skipped - no backup)",
			styles.Dim.Render(fmt.Sprintf("%d", skippedCount)))
	}

	if len(errors) > 0 {
		fmt.Printf(" (%s errors)",
			styles.Error.Render(fmt.Sprintf("%d", len(errors))))
	}

	fmt.Println()
	fmt.Println()

	if len(errors) > 0 {
		return fmt.Errorf("revert completed with %d errors", len(errors))
	}
	return nil
}
