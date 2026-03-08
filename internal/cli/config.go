package cli

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/fjmalass/simplepatch/internal/log"
)

type Config struct {
	MappingPath  string `name:"map" short:"m" default:"patched_files.map" help:"Mappings file" type:"existingfile"`
	PatchesRoot  string `name:"patches-root" short:"p" default:"." help:"Patched files root (copy from)"`
	OriginalRoot string `name:"original-root" short:"o" help:"Patched files root (copy to)"`
	DryRun       bool   `name:"dry-run" short:"d" help:"Dry-Run (no execution)"`
	Backup       bool   `name:"backup" short:"b" default:"true" help:"Backup copied files so we can revert"`
}

const DefaultPatchPath = "Patches"

type PatchCmd struct {
	Config
}

func (c *PatchCmd) Run(logger log.Logger) error {
	cfg := &c.Config
	resolveDefaults(cfg)
	return PerformPatch(cfg, logger)
}

func resolveDefaults(cfg *Config) error {
	var errs []error
	var err error
	if cfg.PatchesRoot == "" || cfg.PatchesRoot == "." {
		wd, err := os.Getwd()
		if err != nil {
			errs = append(errs, err)
		}
		cfg.PatchesRoot = wd
	}

	cfg.PatchesRoot, err = filepath.Abs(cfg.PatchesRoot)
	if err != nil {
		errs = append(errs, err)
	}

	if cfg.OriginalRoot == "" {
		cfg.OriginalRoot = filepath.Join(filepath.Dir(cfg.PatchesRoot), DefaultPatchPath)
	}
	cfg.OriginalRoot, err = filepath.Abs(cfg.OriginalRoot)
	if err != nil {
		errs = append(errs, err)
	}

	// Make absolute mappingPath
	if !filepath.IsAbs(cfg.MappingPath) {
		cfg.MappingPath = filepath.Join(cfg.PatchesRoot, cfg.MappingPath)
	}

	return errors.Join(errs...)
}
