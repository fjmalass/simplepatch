package main

import (
	"github.com/alecthomas/kong"
	"github.com/fjmalass/simplepatch/internal/cli"
	"github.com/fjmalass/simplepatch/internal/format"
	"github.com/fjmalass/simplepatch/internal/log"
)

func main() {
	// Instantiate CLI struct in main with global flags
	var app struct {
		Verbose bool `name:"verbose" short:"v" help:"Enable verbose/debug logging"`

		Patch  cli.PatchCmd  `cmd:"" help:"Apply patches to original files"`
		Clean  cli.CleanCmd  `cmd:"" help:"Remove backup files"`
		Revert cli.RevertCmd `cmd:"" help:"Restore original files from backups"`
	}

	// Create logger early so we can bind it
	logger := log.NewIconLogger()

	ctx := kong.Parse(&app,
		kong.Name("simplepatch"),
		kong.Description("Copy patched files to original locations with backup support"),
		kong.UsageOnError(),
		format.KongOption(format.DefaultTheme()),
		kong.BindTo(logger, (*log.Logger)(nil)), // Bind concrete logger to Logger interface
	)

	// Set log level based on global verbose flag
	if app.Verbose {
		logger.SetLevel(log.DebugLevel)
	}

	// Run command with logger injected via kong.Bind
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
