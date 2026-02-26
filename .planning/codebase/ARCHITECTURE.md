# Architecture

**Analysis Date:** 2026-02-26

## Pattern Overview

**Overall:** Modular CLI with Command Pattern

**Key Characteristics:**
- Self-contained command packages in `cmd/` - each command is isolated and independently testable
- Central command registration in `cmd/cli.go` - single wiring point for all commands
- Internal packages separate business logic from command handling - enables reuse across commands
- Interactive prompts unified in `internal/prompt/` - consistent UX across all commands

## Layers

**Entry Layer:**
- Purpose: Application bootstrap
- Location: `main.go`
- Contains: Minimal entry point that delegates to cmd package
- Depends on: cmd package only

**Command Registration Layer:**
- Purpose: Central wiring of all CLI commands
- Location: `cmd/cli.go`
- Contains: App setup, command registration, error handling, interactive command selection
- Depends on: All command packages, prompt package
- Key function: `Execute()` - registers commands and runs the app

**Command Layer:**
- Purpose: CLI command implementations
- Location: `cmd/*/command.go` (version, ztag, git, gencmd, prj)
- Contains: Command definitions using urfave/cli/v2, subcommand definitions, flag handling
- Depends on: internal/pkg packages for business logic
- Pattern: Each package has `Command()` function returning `*cli.Command`

**Business Logic Layer:**
- Purpose: Core domain functionality
- Location: `internal/pkg/git/`, `internal/pkg/project/`
- Contains: Git operations, project management, data persistence
- Depends on: Standard library, external commands (git)

**Utility Layer:**
- Purpose: Shared functionality across commands
- Location: `internal/prompt/`, `internal/cmd/`
- Contains: Interactive prompts (survey), subcommand validation
- Depends on: urfave/cli, survey

## Data Flow

**CLI Execution Flow:**

1. User runs `cli-aio [command] [args]`
2. `main.go` calls `cmd.Execute()`
3. `cmd/cli.go` initializes urfave/cli App with registered commands
4. urfave/cli routes to appropriate command based on args
5. Command Action executes:
   - Parses flags/args
   - Calls internal package functions for business logic
   - May invoke prompt package for user interaction
   - Returns result or error
6. Error handling in `cmd/cli.go` formats and displays errors

**Interactive Command Selection Flow:**

1. No command provided (`cli-aio` with no args)
2. `cmd/cli.go` Action handler detects empty args
3. Calls `prompt.SelectCommand()` with available commands
4. User selects command from TTY
5. Selected command's Action is executed directly

## Key Abstractions

**Command Package:**
- Purpose: Encapsulates a CLI command with all its subcommands
- Examples: `cmd/prj/`, `cmd/git/`, `cmd/ztag/`
- Pattern: `func Command() *cli.Command` returns urfave/cli command struct

**Internal Package:**
- Purpose: Reusable business logic independent of CLI framework
- Examples: `internal/pkg/git/git.go`, `internal/pkg/project/project.go`
- Pattern: Pure functions with clear inputs/outputs, no CLI dependencies

**Prompt Utilities:**
- Purpose: Unified interactive user experience
- Location: `internal/prompt/prompt.go`
- Key functions: `Select()`, `Input()`, `Confirm()`, `SelectCommand()`
- Handles TTY detection, fuzzy search, /dev/tty fallback

## Entry Points

**Primary Entry:**
- Location: `main.go`
- Triggers: Running the compiled binary
- Responsibilities: Minimal - calls `cmd.Execute()` only

**Command Wiring:**
- Location: `cmd/cli.go`
- Triggers: Called from main.go
- Responsibilities:
  - Register all commands
  - Configure app (name, usage, flags)
  - Handle unknown commands
  - Error formatting and exit handling

**Individual Commands:**
- Location: `cmd/*/command.go`
- Triggers: User runs specific command
- Responsibilities:
  - Define command structure (name, usage, flags, subcommands)
  - Implement Action handler
  - Coordinate business logic and prompts

## Error Handling

**Strategy:** Centralized in cmd/cli.go with command-specific validation

**Patterns:**
- Command validation: `internal/cmd/ValidateSubcommand()` checks subcommand existence
- Business logic errors: Returned as error values from internal packages
- User-friendly messages: Commands format errors with context (e.g., `[!] Warning: ...`)
- Exit handling: `ExitErrHandler` in `cmd/cli.go` catches all errors

## Cross-Cutting Concerns

**Logging:** Printf-style to stdout/stderr (no structured logging)

**Validation:**
- Command arguments validated in command Action handlers
- Path validation in project commands
- Git repo detection in git commands

**Authentication:** Not applicable - CLI operates on local git/project data

---

*Architecture analysis: 2026-02-26*
