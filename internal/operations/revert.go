package operations

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/fjmalass/simplepatch/internal/log"
)

// BackupInfo holds parsed backup file information
type BackupInfo struct {
	Path      string    // Full path to backup file
	BaseName  string    // Original filename without backup suffix
	Timestamp string    // YYYYMMDD_HHMMSS
	Time      time.Time // Parsed timestamp
}

// backupRegex matches: <name>.backup_<YYYYMMDD_HHMMSS><ext>
// Example: dpt.backup_20260306_120000.py
var backupRegex = regexp.MustCompile(`^(.+)\.backup_(\d{8}_\d{6})(\.[^.]*)?$`)

// ParseBackupFilename parses a backup filename and returns BackupInfo
func ParseBackupFilename(path string) (*BackupInfo, error) {
	filename := filepath.Base(path)
	matches := backupRegex.FindStringSubmatch(filename)

	if matches == nil {
		return nil, fmt.Errorf("not a valid backup filename: %s", filename)
	}

	// matches[1] = base name (without .backup_timestamp.ext)
	// matches[2] = timestamp (YYYYMMDD_HHMMSS)
	// matches[3] = extension (including dot, may be empty)

	baseName := matches[1]
	if matches[3] != "" {
		baseName += matches[3] // Add extension back
	}

	timestamp := matches[2]
	t, err := time.ParseInLocation("20060102_150405", timestamp, time.Local)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp %s: %w", timestamp, err)
	}

	return &BackupInfo{
		Path:      path,
		BaseName:  baseName,
		Timestamp: timestamp,
		Time:      t,
	}, nil
}

// FindBackupsForFile finds all backups for a given patched file
func FindBackupsForFile(patchedPath string, logger log.Logger) ([]BackupInfo, error) {
	logger.Debug("finding backups for file", "patched", patchedPath)

	dir := filepath.Dir(patchedPath)
	filename := filepath.Base(patchedPath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	// Pattern: <name>.backup_*<ext>
	pattern := filepath.Join(dir, nameWithoutExt+".backup_*"+ext)
	logger.Debug("searching with pattern", "pattern", pattern)

	matches, err := filepath.Glob(pattern)
	if err != nil {
		logger.Error("glob failed", "pattern", pattern, "err", err)
		return nil, fmt.Errorf("glob failed: %w", err)
	}

	var backups []BackupInfo
	for _, match := range matches {
		info, err := ParseBackupFilename(match)
		if err != nil {
			logger.Warn("skipping invalid backup", "path", match, "err", err)
			continue
		}
		backups = append(backups, *info)
		logger.Debug("found backup", "path", match, "timestamp", info.Timestamp)
	}

	logger.Debug("found backups for file", "patched", patchedPath, "count", len(backups))
	return backups, nil
}

// SelectBackup returns the oldest or newest backup
func SelectBackup(backups []BackupInfo, useLatest bool) *BackupInfo {
	if len(backups) == 0 {
		return nil
	}

	// Sort by time
	sorted := make([]BackupInfo, len(backups))
	copy(sorted, backups)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Time.Before(sorted[j].Time)
	})

	if useLatest {
		return &sorted[len(sorted)-1]
	}
	return &sorted[0]
}

// RevertFile copies backup to original location
func RevertFile(backup BackupInfo, originalPath string, dryRun bool, logger log.Logger) error {
	logger.Debug("reverting file", "backup", backup.Path, "original", originalPath, "dryRun", dryRun)

	// Check backup exists
	if _, err := os.Stat(backup.Path); os.IsNotExist(err) {
		logger.Error("backup file not found", "path", backup.Path)
		return fmt.Errorf("backup file not found: %s", backup.Path)
	}

	if dryRun {
		logger.Debug("dry-run: would revert", "backup", backup.Path, "original", originalPath)
		return nil
	}

	// Copy backup to original
	logger.Info("reverting file", "backup", backup.Path, "original", originalPath)
	return CopyFile(backup.Path, originalPath, logger)
}

// DeleteBackup removes a backup file after successful revert
func DeleteBackup(backup BackupInfo, dryRun bool, logger log.Logger) error {
	logger.Debug("deleting backup after revert", "path", backup.Path, "dryRun", dryRun)

	if dryRun {
		logger.Debug("dry-run: would delete backup", "path", backup.Path)
		return nil
	}

	if err := os.Remove(backup.Path); err != nil {
		logger.Error("failed to delete backup", "path", backup.Path, "err", err)
		return fmt.Errorf("failed to delete backup %s: %w", backup.Path, err)
	}

	logger.Debug("deleted backup", "path", backup.Path)
	return nil
}
