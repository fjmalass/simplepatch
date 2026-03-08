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
	logger.Info("=== Clean Backup Files ===")

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
		logger.Info("[DRY-RUN] no files will be deleted")
	}

	// Find backup files
	logger.Info("searching for backup files", "dir", patchesRoot)
	backups, err := operations.FindBackupFiles(patchesRoot, logger)
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		logger.Info("no backup files found")
		return nil
	}

	// Calculate total size
	totalSize := operations.GetBackupFileSize(backups, logger)

	if cfg.DryRun {
		logger.Info("would delete backup files", "count", len(backups))
	} else {
		logger.Info("deleting backup files")
	}

	deletedCount := 0
	var errors []error

	for _, backup := range backups {
		if cfg.DryRun {
			logger.Info("[DRY-RUN] would delete", "path", backup)
		} else {
			err := operations.DeleteBackupFile(backup, false, logger)
			if err != nil {
				logger.Error("failed to delete", "path", backup, "err", err)
				errors = append(errors, err)
			} else {
				logger.Info("deleted", "path", backup)
				deletedCount++
			}
		}
	}

	// Summary
	if cfg.DryRun {
		logger.Info("dry-run complete", "count", len(backups), "size", operations.FormatBytes(totalSize))
	} else {
		if len(errors) > 0 {
			logger.Error("clean completed with errors", "deleted", deletedCount, "errors", len(errors))
		} else {
			logger.Info("clean completed", "deleted", deletedCount, "freed", operations.FormatBytes(totalSize))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to delete %d files", len(errors))
	}
	return nil
}
