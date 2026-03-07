# Integrating simplepatch into Your Project

This guide shows how to integrate `simplepatch` into another Git repository using `cargo-make`.

## Prerequisites

- Go toolchain installed (1.22+)
- cargo-make installed (`cargo install cargo-make`)

## Installation

simplepatch can be installed via `go install`:

```bash
go install github.com/fjmalass/simplepatch/cmd/simplepatch@v0.1.0
```

This installs the binary to your `$GOPATH/bin` directory.

## Makefile.toml Integration

Add the following to your project's `Makefile.toml` to integrate simplepatch:

```toml
# =============================================================================
# simplepatch integration
# =============================================================================

[env]
# Pin to a specific version
SIMPLEPATCH_VERSION = "v0.1.0"

# Configure paths for your project
PATCHES_ROOT = "patches"          # Directory containing your patched files
ORIGINAL_ROOT = "../source"       # Directory containing original files to patch
MAPPING_FILE = "patched_files.map"

# =============================================================================
# Installation
# =============================================================================

[tasks.install-simplepatch]
description = "Install simplepatch CLI tool"
command = "go"
args = ["install", "github.com/fjmalass/simplepatch/cmd/simplepatch@${SIMPLEPATCH_VERSION}"]

# =============================================================================
# Patch Operations
# =============================================================================

[tasks.patch]
description = "Apply patches to source files"
dependencies = ["install-simplepatch"]
command = "simplepatch"
args = ["patch", "-p", "${PATCHES_ROOT}", "-o", "${ORIGINAL_ROOT}", "-m", "${MAPPING_FILE}"]

[tasks.patch-dry-run]
description = "Preview patch operations (dry-run)"
dependencies = ["install-simplepatch"]
command = "simplepatch"
args = ["-v", "patch", "--dry-run", "-p", "${PATCHES_ROOT}", "-o", "${ORIGINAL_ROOT}", "-m", "${MAPPING_FILE}"]

# =============================================================================
# Backup Management
# =============================================================================

[tasks.clean-backups]
description = "Remove backup files from patches directory"
dependencies = ["install-simplepatch"]
command = "simplepatch"
args = ["clean", "-p", "${PATCHES_ROOT}"]

[tasks.clean-backups-dry-run]
description = "Preview backup cleanup (dry-run)"
dependencies = ["install-simplepatch"]
command = "simplepatch"
args = ["clean", "--dry-run", "-p", "${PATCHES_ROOT}"]

# =============================================================================
# Revert Operations
# =============================================================================

[tasks.revert]
description = "Revert to original files from oldest backups"
dependencies = ["install-simplepatch"]
command = "simplepatch"
args = ["revert", "-p", "${PATCHES_ROOT}", "-o", "${ORIGINAL_ROOT}", "-m", "${MAPPING_FILE}"]

[tasks.revert-latest]
description = "Revert to original files from newest backups"
dependencies = ["install-simplepatch"]
command = "simplepatch"
args = ["revert", "--use-latest", "-p", "${PATCHES_ROOT}", "-o", "${ORIGINAL_ROOT}", "-m", "${MAPPING_FILE}"]

[tasks.revert-dry-run]
description = "Preview revert operations (dry-run)"
dependencies = ["install-simplepatch"]
command = "simplepatch"
args = ["-v", "revert", "--dry-run", "-p", "${PATCHES_ROOT}", "-o", "${ORIGINAL_ROOT}", "-m", "${MAPPING_FILE}"]
```

## Mapping File Format

Create a `patched_files.map` file in your patches directory:

```
#patched->original
pyproject.toml -> pyproject.toml
src/module.py -> lib/module.py
config.yaml
```

Format:
- First line must be `#patched->original` (header)
- Each line: `patched_path -> original_path`
- Single path (e.g., `config.yaml`) means same path for both
- Lines starting with `#` are comments
- Inline comments supported: `file.py -> dest.py # comment`

## Usage Examples

```bash
# Preview what will be patched
cargo make patch-dry-run

# Apply patches (with automatic backup of originals)
cargo make patch

# See what backups exist
cargo make clean-backups-dry-run

# Remove all backup files
cargo make clean-backups

# Revert to original state using oldest backups
cargo make revert

# Revert using most recent backups
cargo make revert-latest
```

## Directory Structure Example

```
your-project/
  patches/
    patched_files.map       # Mapping file
    pyproject.toml          # Your patched version
    src/
      module.py             # Your patched version
  source/                   # Original source (e.g., git submodule)
    pyproject.toml          # Will be overwritten
    lib/
      module.py             # Will be overwritten
  Makefile.toml
```

## Workflow

1. **Initial setup**: Create patches directory with your modified files
2. **Create mapping**: Define `patched_files.map` with source->dest mappings
3. **Preview**: Run `cargo make patch-dry-run` to verify
4. **Apply**: Run `cargo make patch` to copy patched files to source
5. **Work**: Edit files in patches directory as needed
6. **Re-apply**: Run `cargo make patch` again after edits
7. **Revert**: If needed, run `cargo make revert` to restore originals

## CLI Reference

```
simplepatch - Copy patched files to original locations with backup support

Commands:
  patch   Apply patches to original files
  clean   Remove backup files
  revert  Restore original files from backups

Global Flags:
  -v, --verbose    Enable debug logging
  -h, --help       Show help

Patch Flags:
  -p, --patches-root    Directory containing patched files (default: .)
  -o, --original-root   Directory containing original files
  -m, --map             Mapping file (default: patched_files.map)
  -d, --dry-run         Preview without making changes
  -b, --backup          Backup originals before overwriting (default: true)

Clean Flags:
  -p, --patches-root    Directory to search for backups (default: .)
  -d, --dry-run         Preview without deleting
  -a, --all             Remove all backups recursively

Revert Flags:
  -p, --patches-root    Directory containing backups (default: .)
  -o, --original-root   Directory to restore files to
  -m, --map             Mapping file (default: patched_files.map)
  -d, --dry-run         Preview without making changes
  --use-latest          Use newest backup instead of oldest
  -c, --cleanup         Delete backup after successful revert
```
