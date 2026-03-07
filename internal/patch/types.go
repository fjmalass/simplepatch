package patch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Mapping struct {
	Patched  string // must be abolute
	Original string // must be absolute
}

func (m *Mapping) IsValid() error {
	var errs []error
	if !filepath.IsAbs(m.Patched) {
		errs = append(errs, fmt.Errorf("patched path is not absolute: %q", m.Patched))
	}

	info, err := os.Stat(m.Patched)
	if err != nil {
		errs = append(errs, fmt.Errorf("patched file does not exist %q: %w", m.Patched, err))
	} else if info.IsDir() {
		errs = append(errs, fmt.Errorf("patched file is a directory %q", m.Patched))
	}

	if !filepath.IsAbs(m.Original) {
		errs = append(errs, fmt.Errorf("original path is not absolute: %q", m.Original))
	}

	return errors.Join(errs...)
}
