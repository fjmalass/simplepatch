package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/alecthomas/kong"
	"github.com/fjmalass/simplepatch/internal/cli"
	"github.com/fjmalass/simplepatch/internal/format"
	"github.com/fjmalass/simplepatch/internal/log"
)

// version is set via ldflags at build time
var version = ""

func getVersion() string {
	// Prefer ldflags-injected version
	if version != "" {
		return version
	}
	// Fall back to Go module version (works with go install ...@tag)
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return "dev"
}

func main() {
	// Instantiate CLI struct in main with global flags
	var app struct {
		Version kong.VersionFlag `name:"version" short:"V" help:"Show version"`
		Verbose bool             `name:"verbose" short:"v" help:"Enable verbose/debug logging"`
		LogFile string           `name:"log" short:"l" help:"Write logs to file"`

		Patch  cli.PatchCmd  `cmd:"" help:"Apply patches to original files"`
		Clean  cli.CleanCmd  `cmd:"" help:"Remove backup files"`
		Revert cli.RevertCmd `cmd:"" help:"Restore original files from backups"`
	}

	// Build logger options based on flags
	// Note: We parse flags first with a temporary parse to get LogFile/Verbose,
	// then create the logger. However, kong.Parse consumes args, so we need
	// to create the logger after parsing but before Run().

	ctx := kong.Parse(&app,
		kong.Name("simplepatch"),
		kong.Description("Copy patched files to original locations with backup support"),
		kong.UsageOnError(),
		kong.Vars{"version": getVersion()},
		format.KongOption(format.DefaultTheme()),
	)

	// Build logger options from parsed flags
	var opts []log.Option
	if app.Verbose {
		opts = append(opts, log.WithLevel(log.DebugLevel))
	}
	if app.LogFile != "" {
		opts = append(opts, log.WithFile(app.LogFile))
	}

	// Create logger with options
	logger, err := log.NewIconLogger(opts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	if app.LogFile != "" {
		logger.Info("logging to file", "path", app.LogFile)
	}
	logger.Info("version", "version", getVersion())

	// Bind logger for dependency injection into commands
	ctx.BindTo(logger, (*log.Logger)(nil))

	// Run command with logger injected via kong.Bind
	err = ctx.Run()
	ctx.FatalIfErrorf(err)
}
