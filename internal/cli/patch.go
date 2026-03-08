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
	logger.Info("=== Copy Patched Files ===")

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
		logger.Info("[DRY-RUN] no files will be modified")
	}

	// Phase 1: Backup existing originals
	if cfg.Backup {
		logger.Info("backing up existing files")
		for _, m := range mappings {
			backupDir := filepath.Dir(m.Patched)
			err := operations.BackupOriginal(m.Original, backupDir, timestamp, cfg.DryRun, logger)
			if err != nil {
				logger.Error("failed to backup", "path", m.Original, "err", err)
			}
		}
	}

	// Phase 2: Copy patched -> original
	logger.Info("copying patched files", "to", cfg.OriginalRoot)

	successCount := 0
	errorCount := 0
	for _, m := range mappings {
		err := operations.CopyPatchedToOriginal(m.Patched, m.Original, cfg.DryRun, logger)
		if err != nil {
			logger.Error("failed to copy", "err", err)
			errorCount++
		} else if !cfg.DryRun {
			logger.Info("copied", "from", m.Patched, "to", m.Original)
			successCount++
		}
	}

	// Summary
	if errorCount == 0 {
		logger.Info("patch completed successfully", "count", successCount)
		for i, m := range mappings {
			logger.Info("Patched", "index", i, "path", m.Patched)
		}
	} else {
		logger.Error("patch completed with errors", "errors", errorCount, "total", errorCount+successCount)
	}

	return nil
}
