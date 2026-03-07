package operations

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fjmalass/simplepatch/internal/log"
)

const postfix = ".backup_"

// CurrentTimestamp generates YYYYMMDD_HHMMSS format
func CurrentTimestamp() string {
	return time.Now().Format("20060102_150405")
}

// backupBaseFilename: generate Base.backup_YYYYMMDD_HHMMSS.ext
func backupBaseFilename(base, timestamp string) string {
	_, file := filepath.Split(base)
	ext := filepath.Ext(file)
	name := strings.TrimSuffix(file, ext)
	return name + postfix + timestamp + ext
}

// BackupFilename: generate Target/Base.backup_YYYYMMDD_HHMMSS.ext
func BackupFilename(originPath, targetDirPath, timestamp string) string {
	return filepath.Join(targetDirPath, backupBaseFilename(originPath, timestamp))
}

// CopyFile copies src to dst, creating parent directories if needed
func CopyFile(src, dst string, logger log.Logger) error {
	logger.Debug("copying file", "src", src, "dst", dst)

	sourceFile, err := os.Open(src)
	if err != nil {
		logger.Error("failed to open source file", "src", src, "err", err)
		return err
	}
	defer sourceFile.Close()

	// Get source file info for permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		logger.Error("failed to stat source file", "src", src, "err", err)
		return err
	}

	// Create parent directory if needed
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		logger.Error("failed to create destination directory", "dir", dstDir, "err", err)
		return err
	}

	destFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sourceInfo.Mode())
	if err != nil {
		logger.Error("failed to create destination file", "dst", dst, "err", err)
		return err
	}
	defer destFile.Close()

	bytes, err := io.Copy(destFile, sourceFile)
	if err != nil {
		logger.Error("failed to copy file contents", "err", err)
		return err
	}

	logger.Debug("file copied successfully", "bytes", bytes)
	return nil
}

// BackupOriginal backs up the original file to backupDir with timestamp
// Returns nil if original doesn't exist (no backup needed)
func BackupOriginal(originalPath, backupDir, timestamp string, dryRun bool, logger log.Logger) error {
	logger.Debug("checking if backup needed", "original", originalPath)

	// Check if original exists
	if _, err := os.Stat(originalPath); os.IsNotExist(err) {
		logger.Debug("original file does not exist, skipping backup", "path", originalPath)
		fmt.Printf("  No existing file: %s (no backup needed)\n", originalPath)
		return nil
	}

	backupPath := BackupFilename(originalPath, backupDir, timestamp)

	if dryRun {
		logger.Debug("dry-run: would backup", "from", originalPath, "to", backupPath)
		fmt.Printf("  [DRY-RUN] Would backup: %s -> %s\n", originalPath, backupPath)
		return nil
	}

	logger.Info("backing up file", "from", originalPath, "to", backupPath)
	if err := CopyFile(originalPath, backupPath, logger); err != nil {
		return fmt.Errorf("backup failed: %w", err)
	}

	fmt.Printf("  Backed up: %s -> %s\n", originalPath, backupPath)
	return nil
}

// CopyPatchedToOriginal copies the patched file to the original location
func CopyPatchedToOriginal(patchedPath, originalPath string, dryRun bool, logger log.Logger) error {
	logger.Debug("preparing to copy patched to original", "patched", patchedPath, "original", originalPath)

	// Check patched file exists
	if _, err := os.Stat(patchedPath); os.IsNotExist(err) {
		logger.Error("patched file missing", "path", patchedPath)
		return fmt.Errorf("patched file missing: %s", patchedPath)
	}

	if dryRun {
		logger.Debug("dry-run: would copy", "from", patchedPath, "to", originalPath)
		fmt.Printf("  [DRY-RUN] Would copy: %s -> %s\n", patchedPath, originalPath)
		return nil
	}

	logger.Info("copying patched to original", "from", patchedPath, "to", originalPath)
	return CopyFile(patchedPath, originalPath, logger)
}
