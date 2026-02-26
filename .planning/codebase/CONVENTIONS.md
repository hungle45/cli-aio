# Coding Conventions

**Analysis Date:** 2026-02-26

## Language

**Primary:**
- Go 1.21 - All application code

## Naming Patterns

### Files
- **Commands:** `command.go` - Each command package contains a `command.go` file that exports a `Command()` function returning `*cli.Command`
  - Examples: `cmd/ztag/command.go`, `cmd/git/command.go`, `cmd/prj/command.go`
- **Multiple files per command:** When a command grows large, split into additional files in the same package
  - Example: `cmd/prj/command.go` and `cmd/prj/install.go`
- **Internal packages:** `internal/pkg/*/` for shared logic packages
  - Example: `internal/pkg/git/git.go`, `internal/pkg/project/project.go`
- **Utilities:** `internal/prompt/prompt.go` for interactive prompt utilities

### Functions
- **Exported functions:** PascalCase - Start with uppercase letter
  - Examples: `Execute()`, `Command()`, `CheckIfGitRepo()`, `SelectCommand()`
- **Unexported functions:** camelCase - Start with lowercase letter
  - Examples: `findCommand()`, `getCommandPath()`, `showUnknownCommandWarning()`
- **Constructor-like functions:** Often use the type name as function name for clarity
  - Example: `func Command() *cli.Command` - returns a new CLI command

### Variables
- **Local variables:** camelCase
  - Examples: `currentBranch`, `targetPath`, `commandNames`
- **Package-level variables:** PascalCase for exported, camelCase for unexported
  - Examples: `Version`, `BuildTime` (exported), `defaultEnvMap` (unexported)
- **Constants:** PascalCase for exported const groups, camelCase for unexported
  - Examples: `EnvQC`, `EnvStg`, `EnvProd` (const group), `markerBegin`, `markerEnd` (unexported)
- **Struct fields:** PascalCase for exported, camelCase for unexported
  - Example: `Project.Name`, `Project.Path`

### Types
- **Structs:** PascalCase, singular nouns
  - Examples: `Project`, `VersionInfo`, `TagComponents`, `shellConfig`
- **Interfaces:** PascalCase, often end with "er" suffix for implementations
  - Example: `TagTemplate` interface in `cmd/ztag/tag.go`
- **Custom types (type aliases):** PascalCase
  - Examples: `type Env string`, `type Level string`

## Code Style

### Formatting
- **Tool:** Go's built-in `gofmt` - No external formatter configured
- **Tab indentation:** Use tabs for indentation (Go standard)
- **Line length:** No strict line length limit, but keep functions readable
- **Vertical spacing:** Single blank line between top-level declarations
- **Import grouping:** Standard library imports first, then third-party imports
  ```go
  import (
      "fmt"
      "os"
      "path/filepath"
      
      "github.com/urfave/cli/v2"
      "golang.org/x/term"
  )
  ```

### Linting
- **Tool:** No explicit linter configured (no `.golangci.yml`)
- **Implicit conventions:** Relies on `go vet` and IDE defaults
- **Recommendations for new code:**
  - Run `go vet ./...` before commits
  - Consider adding `golangci-lint` for comprehensive linting

## Import Organization

### Order (Standard Go import grouping):
1. **Standard library:** `fmt`, `os`, `strings`, `path/filepath`, etc.
2. **External packages (third-party):** `github.com/urfave/cli/v2`, `github.com/AlecAivazis/survey/v2`
3. **Internal packages:** `cli-aio/internal/...`, `cli-aio/cmd/...`

### Path Aliases
- **Module name:** `cli-aio` - Used as base for internal imports
  - Examples: `"cli-aio/cmd"`, `"cli-aio/internal/prompt"`, `"cli-aio/internal/pkg/git"`

## Error Handling

### Patterns

**1. Return errors explicitly:**
```go
func CheckIfGitRepo() (bool, error) {
    cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
    output, err := cmd.Output()
    if err != nil {
        return false, fmt.Errorf("error running git command to check if git repository: %w", err)
    }
    return strings.TrimSpace(string(output)) == "true", nil
}
```

**2. Use wrapped errors with `%w`:**
- Always wrap errors with `fmt.Errorf("...: %w", err)` for proper error chaining

**3. Error messages:**
- Start with lowercase: `"error running git command..."` not `"Error running..."`
- Be descriptive: Include context about what operation failed

**4. Early returns:**
```go
func ExtractProjectID() (string, error) {
    fullName, err := ExtractProjectFullName()
    if err != nil {
        return "", err  // Early return on error
    }
    parts := strings.Split(fullName, "/")
    projectID := strings.Join(parts[1:], "/")
    return projectID, nil
}
```

**5. CLI error handling:**
- Commands return errors that are handled by `cmd/cli.go`'s `ExitErrHandler`
- Use `fmt.Errorf` to create user-friendly error messages
- Print errors to `os.Stderr` using `fmt.Fprintf(os.Stderr, "[-] Error: %v\n", err)`

## Logging

**Framework:** Standard library `fmt` package

**Patterns:**
- **Info messages:** `fmt.Printf()` for standard output
  ```go
  fmt.Printf("Project ID: %s\n", projectID)
  fmt.Printf("[+] Successfully merged '%s' into '%s'\n", currentBranch, targetBranch)
  ```
- **Warnings:** Print to `os.Stderr` with `[!]` prefix
  ```go
  fmt.Fprintf(os.Stderr, "[!] Warning: Failed to fetch branch: %v\n", err)
  ```
- **Errors:** Print to `os.Stderr` with `[-]` prefix
  ```go
  fmt.Fprintf(os.Stderr, "[-] Error: %v\n", err)
  ```
- **Success:** Print to stdout with `[+]` prefix
  ```go
  fmt.Printf("[+] Added project: %s (%s)\n", p.Name, p.Path)
  ```

**Console UI Patterns:**
- Use ASCII prefixes for status: `[!]` warning, `[-]` error, `[+]` success
- Use arrow notation for actions: `->` for navigation, `>>` for selection
- Keep messages concise and actionable

## Comments

### When to Comment
- **Public API documentation:** All exported functions should have comments
  ```go
  // Command returns a simple version command.
  // This demonstrates a minimal command without subcommands.
  func Command() *cli.Command
  ```
- **Complex logic:** Explain non-obvious implementation details
  ```go
  // CheckMergeConflicts checks if merging sourceBranch into current branch would cause conflicts.
  // Returns true if there would be conflicts, false otherwise.
  // Uses a test merge approach: attempts merge with --no-commit and --no-ff, then aborts.
  ```
- **Workarounds:** Document why a particular approach was taken
  ```go
  // SelectOnTTY is like Select but forces all survey I/O through /dev/tty.
  // Use this when stdout is captured (e.g. inside $(...)) so that the
  // interactive UI is shown on the terminal instead of being swallowed.
  ```

### Style
- **Sentence case:** Start comments with capital letter
- **Imperative mood:** "Check if..." not "Checking if..."
- **Complete sentences:** Include period at end for multi-sentence comments
- **Go doc conventions:** First sentence becomes the summary in generated docs

### TODO Comments
- Currently no TODO/FIXME comments found in codebase
- Use `// TODO:` prefix for future improvements

## Function Design

### Size Guidelines
- **Keep functions focused:** Each function should do one thing well
- **Average function length:** 20-50 lines for most functions
- **Large functions:** Split into helper functions when > 100 lines
  - Example: `cmd/cli.go` splits logic into `findCommand()`, `getCommandPath()`, `showUnknownCommandWarning()`

### Parameters
- **Limit parameters:** Maximum 4-5 parameters
- **Use structs for many parameters:** Group related params into structs
  ```go
  type VersionInfo struct {
      Major int
      Minor int
      Patch int
  }
  ```
- **Use options pattern for extensibility:** Not currently used, but good pattern for future

### Return Values
- **Multiple returns:** Use named returns when clarity benefits
  ```go
  func CheckIfGitRepo() (bool, error)
  ```
- **Error as last return:** Convention is `(value, error)` pattern
- **Nil slices:** Return empty slice `[]string{}` instead of `nil` for slices that may be empty

## Module Design

### Package Structure

**1. Command packages (`cmd/*/`):**
- Self-contained command implementations
- Export a single `Command() *cli.Command` function
- May have multiple files for organization

**2. Internal packages (`internal/*/`):**
- `internal/cmd/` - Shared command utilities (e.g., `ValidateSubcommand`)
- `internal/prompt/` - Interactive prompt helpers
- `internal/pkg/*/` - Shared domain logic (git, project)

### Exports
- **Minimal exports:** Only export what's needed
- **Command pattern:** Each command package exports only `Command()` function
- **Internal packages:** Use unexported functions freely within package

### Barrel Files
- Not used - each package exposes only what's needed
- Avoid circular imports by keeping packages focused

---

*Convention analysis: 2026-02-26*
