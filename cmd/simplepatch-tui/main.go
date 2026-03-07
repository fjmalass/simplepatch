package main

import (
	"fmt"
	"os"
)

const helpText = `
simplepatch-tui - Interactive TUI for simplepatch

DESCRIPTION:
    This is the interactive terminal user interface (TUI) for simplepatch.
    It provides a visual interface for managing patched files.

STATUS:
    TUI mode is not yet implemented.

ALTERNATIVE:
    Use the CLI version instead:

        simplepatch patch [flags]

    Common flags:
        -m, --map=FILE          Mappings file (default: patched_files.map)
        -p, --patches-root=DIR  Patched files root (copy from)
        -o, --original-root=DIR Original files root (copy to)
        -d, --dry-run           Show what would happen without making changes
        -b, --backup            Backup original files before overwriting (default: true)
        -v, --verbose           Enable debug logging

    Example:
        simplepatch patch -p ./patches -o ./source -m mappings.map --dry-run

For more information, run: simplepatch --help
`

func main() {
	fmt.Println(helpText)
	os.Exit(1)
}
