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
	logger.Info("=== Revert from Backups ===")

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
		logger.Info("[DRY-RUN] no files will be modified")
	}

	if cfg.UseLatest {
		logger.Info("reverting from newest backups")
	} else {
		logger.Info("reverting from oldest backups")
	}

	revertedCount := 0
	skippedCount := 0
	var errors []error

	for _, m := range mappings {
		// Find backups for this file
		backups, err := operations.FindBackupsForFile(m.Patched, logger)
		if err != nil {
			logger.Warn("failed to find backups", "patched", m.Patched, "err", err)
			errors = append(errors, err)
			continue
		}

		if len(backups) == 0 {
			logger.Warn("no backup found", "patched", m.Patched)
			skippedCount++
			continue
		}

		// Select backup (oldest or newest)
		backup := operations.SelectBackup(backups, cfg.UseLatest)
		if backup == nil {
			skippedCount++
			continue
		}

		if cfg.DryRun {
			logger.Info("[DRY-RUN] would revert", "from", backup.Path, "to", m.Original)
			if cfg.Cleanup {
				logger.Info("[DRY-RUN] would delete backup", "path", backup.Path)
			}
			revertedCount++
		} else {
			// Perform the revert
			err := operations.RevertFile(*backup, m.Original, false, logger)
			if err != nil {
				logger.Error("failed to revert", "patched", m.Patched, "err", err)
				errors = append(errors, err)
				continue
			}

			logger.Info("reverted", "to", m.Original, "backup_time", backup.Time.Format("2006-01-02 15:04:05"))
			revertedCount++

			// Cleanup if requested
			if cfg.Cleanup {
				err := operations.DeleteBackup(*backup, false, logger)
				if err != nil {
					logger.Warn("failed to delete backup", "path", backup.Path, "err", err)
				} else {
					logger.Info("cleaned backup", "path", backup.Path)
				}
			}
		}
	}

	// Summary
	if cfg.DryRun {
		logger.Info("dry-run complete", "would_revert", revertedCount, "skipped", skippedCount)
	} else {
		if len(errors) > 0 {
			logger.Error("revert completed with errors", "reverted", revertedCount, "skipped", skippedCount, "errors", len(errors))
		} else {
			logger.Info("revert completed", "reverted", revertedCount, "skipped", skippedCount)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("revert completed with %d errors", len(errors))
	}
	return nil
}
