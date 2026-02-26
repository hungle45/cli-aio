# Codebase Structure

**Analysis Date:** 2026-02-26

## Directory Layout

```
cli-aio/
├── main.go              # Entry point
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── Makefile             # Build commands
├── README.md            # Documentation
├── cmd/                 # Command implementations
│   ├── cli.go          # Central command registration
│   ├── version/         # Version command
│   ├── ztag/           # Tag generation command
│   ├── git/            # Git helper commands
│   ├── gencmd/         # Command generator
│   └── prj/            # Project management
└── internal/            # Private packages
    ├── cmd/            # Shared CLI utilities
    ├── pkg/            # Business logic packages
    │   ├── git/        # Git operations
    │   └── project/    # Project management
    └── prompt/        # Interactive prompts
```

## Directory Purposes

**`cmd/`:**
- Purpose: All CLI command implementations
- Contains: Self-contained command packages, central CLI wiring
- Key files: `cmd/cli.go`, `cmd/prj/command.go`, `cmd/git/command.go`

**`cmd/cli.go`:**
- Purpose: Central command registration and app configuration
- Contains: App setup, command registration, error handling, interactive mode

**`cmd/prj/`:**
- Purpose: Project management commands (add, cd, config)
- Contains: `command.go` with all project subcommands

**`cmd/git/`:**
- Purpose: Git helper commands (fname, rmerge, ckl)
- Contains: `command.go` with git utility subcommands

**`cmd/ztag/`:**
- Purpose: Tag generation for Zalopay deployments
- Contains: `command.go`, `tag.go`

**`cmd/gencmd/`:**
- Purpose: Generate new CLI commands
- Contains: `command.go` - code generation logic

**`cmd/version/`:**
- Purpose: Display version information
- Contains: `command.go` with version display

**`internal/`:**
- Purpose: Private packages not exposed as commands
- Contains: Business logic, utilities (not part of public API)

**`internal/pkg/git/`:**
- Purpose: Git operations (branch management, tags, merge)
- Contains: `git.go` - wrapper around git CLI commands

**`internal/pkg/project/`:**
- Purpose: Project persistence and discovery
- Contains: `project.go` - JSON file storage, git repo scanning

**`internal/prompt/`:**
- Purpose: Interactive terminal prompts
- Contains: `prompt.go` - Select, Input, Confirm, SelectCommand

**`internal/cmd/`:**
- Purpose: Shared CLI utilities
- Contains: `validate.go` - subcommand validation

## Key File Locations

**Entry Points:**
- `main.go`: Application entry point, calls `cmd.Execute()`

**Configuration:**
- `go.mod`: Go module definition (module cli-aio, go 1.21)
- `Makefile`: Build targets

**Command Wiring:**
- `cmd/cli.go`: Central command registration (`Execute()` function)

**Command Implementations:**
- `cmd/version/command.go`: Version display command
- `cmd/ztag/command.go`: Tag generation for qc/stg/prod
- `cmd/ztag/tag.go`: Tag version calculation
- `cmd/git/command.go`: Git helpers (fname, rmerge, ckl)
- `cmd/prj/command.go`: Project management (cd, add, add-git, config)
- `cmd/prj/install.go`: Shell wrapper installation
- `cmd/gencmd/command.go`: Command generator

**Business Logic:**
- `internal/pkg/git/git.go`: Git CLI wrappers
- `internal/pkg/project/project.go`: Project storage and discovery

**Utilities:**
- `internal/prompt/prompt.go`: Interactive prompt functions
- `internal/cmd/validate.go`: Subcommand validation

## Naming Conventions

**Files:**
- Go source: `snake_case.go` (e.g., `command.go`, `git.go`)
- Command packages: `kebab-case` directory with `command.go` inside

**Functions:**
- Exported: `PascalCase` (e.g., `Command()`, `Execute()`, `Load()`)
- Unexported: `camelCase` (e.g., `cdCmd()`, `addCmd()`)

**Packages:**
- Command packages: `kebab-case` (e.g., `cmd/gencmd`, `cmd/prj`)
- Internal packages: `snake_case` (e.g., `internal/pkg/git`)

**Types:**
- Structs: `PascalCase` (e.g., `Project`, `VersionInfo`)
- Interfaces: `PascalCase` with `er` suffix (none currently)

**Constants:**
- Grouped with `var` blocks, `PascalCase` (e.g., `EnvQC`, `LevelBug`)

## Where to Add New Code

**New Top-Level Command:**
- Implementation: Create `cmd/<command-name>/` directory
- Add `command.go` with `Command() *cli.Command` function
- Import in `cmd/cli.go` and add to commands slice
- Tests: Create `command_test.go` in same directory

**New Subcommand:**
- Implementation: Add function returning `*cli.Command` in existing command package
- Register in parent's `subcommands` slice

**New Business Logic (reusable):**
- Implementation: Create `internal/pkg/<package-name>/`
- Export functions with clear interfaces
- Import in command packages as needed

**New Prompt Utility:**
- Implementation: Add to `internal/prompt/prompt.go`
- Export function for use by commands

## Special Directories

**`cmd/`:**
- Purpose: All command implementations
- Generated: No (hand-written)
- Committed: Yes

**`internal/`:**
- Purpose: Private packages (not importable by external packages)
- Generated: No
- Committed: Yes

**`.git/`:**
- Purpose: Git repository metadata
- Generated: Yes (by git)
- Committed: Yes (partial - hooks are samples)

---

*Structure analysis: 2026-02-26*
