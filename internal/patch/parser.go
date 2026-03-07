package patch

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/fjmalass/simplepatch/internal/log"
)

const (
	mapHeader = "#patched->original"
	sep       = "->"
)

func Load(path string, logger log.Logger) ([]Mapping, error) {
	logger.Debug("opening mappings file", "path", path)

	f, err := os.Open(path)
	if err != nil {
		logger.Error("cannot open mappings file", "path", path, "err", err)
		return nil, fmt.Errorf("cannot open mappings file %q: %w", path, err)
	}
	defer f.Close()

	var mappings []Mapping
	scanner := bufio.NewScanner(f)
	lineNum := 0

	var errs []error
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		logger.Debug("parsing line", "lineNum", lineNum, "content", line)

		if line == "" || strings.HasPrefix(line, "#") {
			// check the header
			if lineNum == 1 && strings.ToLower(strings.ReplaceAll(line, " ", "")) != mapHeader {
				logger.Error("invalid header", "expected", mapHeader, "got", line)
				return nil, fmt.Errorf("invalid header in %q (expected %q)", path, mapHeader)
			}
			logger.Debug("skipping comment/empty line", "lineNum", lineNum)
			continue
		}

		// remove inline comments
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}

		var patched, original string

		// Check if separator exists - support single-path format
		if strings.Contains(line, sep) {
			parts := strings.SplitN(line, sep, 2)
			if len(parts) != 2 {
				logger.Warn("invalid mapping format", "lineNum", lineNum, "line", line)
				errs = append(errs, fmt.Errorf("invalid mapping on line %d: %q", lineNum, line))
				continue // FIX: prevent index out of bounds
			}
			patched = strings.TrimSpace(parts[0])
			original = strings.TrimSpace(parts[1])
		} else {
			// Single path - same for both patched and original
			patched = line
			original = line
			logger.Debug("single-path mapping", "path", line)
		}

		if patched == "" || original == "" {
			logger.Warn("empty path in mapping", "lineNum", lineNum, "patched", patched, "original", original)
			errs = append(errs, fmt.Errorf("empty patched or original path on line %d", lineNum))
			continue
		}

		logger.Debug("parsed mapping", "patched", patched, "original", original)
		mappings = append(mappings, Mapping{Patched: patched, Original: original})
	}

	if err := scanner.Err(); err != nil {
		logger.Error("scanner error", "err", err)
		return nil, err
	}
	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	if len(mappings) == 0 {
		logger.Error("no valid mappings found", "path", path)
		return nil, fmt.Errorf("no valid mappings found in %q", path)
	}

	logger.Info("mappings loaded successfully", "count", len(mappings))
	return mappings, nil
}
