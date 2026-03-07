package operations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fjmalass/simplepatch/internal/log"
)

// backupPattern is the substring that identifies backup files
const backupPattern = ".backup_"

// FindBackupFiles finds all .backup_* files in a directory recursively
func FindBackupFiles(dir string, logger log.Logger) ([]string, error) {
	logger.Debug("searching for backup files", "dir", dir)

	var backups []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Warn("error accessing path", "path", path, "err", err)
			return nil // Continue walking
		}

		if info.IsDir() {
			return nil
		}

		// Check if filename contains .backup_
		if strings.Contains(info.Name(), backupPattern) {
			logger.Debug("found backup file", "path", path)
			backups = append(backups, path)
		}

		return nil
	})

	if err != nil {
		logger.Error("failed to walk directory", "dir", dir, "err", err)
		return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	logger.Info("found backup files", "count", len(backups))
	return backups, nil
}

// GetBackupFileSize returns the total size of backup files
func GetBackupFileSize(paths []string, logger log.Logger) int64 {
	var total int64
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			logger.Warn("failed to stat file", "path", path, "err", err)
			continue
		}
		total += info.Size()
	}
	return total
}

// DeleteBackupFile removes a backup file
func DeleteBackupFile(path string, dryRun bool, logger log.Logger) error {
	logger.Debug("deleting backup file", "path", path, "dryRun", dryRun)

	if dryRun {
		logger.Debug("dry-run: would delete", "path", path)
		return nil
	}

	if err := os.Remove(path); err != nil {
		logger.Error("failed to delete backup file", "path", path, "err", err)
		return fmt.Errorf("failed to delete %s: %w", path, err)
	}

	logger.Debug("deleted backup file", "path", path)
	return nil
}

// FormatBytes formats bytes to human readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
