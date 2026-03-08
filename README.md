# simplepatch

A CLI tool to copy patched files to original locations with backup support.

## Installation

```bash
go install github.com/fjmalass/simplepatch/cmd/simplepatch@latest
```

## Usage

```bash
simplepatch <command> [flags]
```

**Commands:**
- `patch` - Copy patched files to original locations using a mapping file
- `clean` - Remove backup files
- `revert` - Restore original files from backups

The `patch` command reads a mapping file that defines source -> destination pairs, then copies files from the patches directory to the original directory.

**Mapping file example:**
```
# patched -> original
pyproject.toml -> pyproject.toml
dpt.py -> depth_anything_v2/dpt.py
```

**Global flags:**
- `-v, --verbose` - Enable debug logging
- `-l, --log FILE` - Write logs to file

Use `simplepatch --help` or `simplepatch <command> --help` to see all available flags.

## Example

```bash
simplepatch patch -p ./patches -o ./source -m mappings.map --dry-run
```

## TUI

The interactive TUI (`simplepatch-tui`) is planned but not yet implemented. Use the CLI version for now.

## License

Apache 2.0 with patent grant.
