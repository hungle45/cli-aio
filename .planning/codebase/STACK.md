# Technology Stack

**Analysis Date:** 2026-02-26

## Languages

**Primary:**
- Go 1.21 - Core CLI application language

**Secondary:**
- Not applicable - Single-language project

## Runtime

**Environment:**
- Go runtime (compiled binary)
- Operating System: macOS/Linux (Unix-like systems)

**Package Manager:**
- Go modules (go.mod/go.sum)

## Frameworks

**Core:**
- [urfave/cli/v2](https://github.com/urfave/cli) v2.27.1 - CLI application framework
  - Used in: `cmd/cli.go`, `cmd/*/command.go`, `cmd/*/command.go`
  - Provides command/subcommand registration, flag parsing, help generation

**Interactive UI:**
- [AlecAivazis/survey/v2](https://github.com/AlecAivazis/survey) v2.3.7 - Interactive prompts
  - Used in: `internal/prompt/prompt.go`
  - Provides: Select prompts, Input prompts, Confirmation prompts

**Terminal Utilities:**
- [golang.org/x/term](https://pkg.go.dev/golang.org/x/term) v0.15.0 - Terminal manipulation
  - Used for terminal detection and interactive mode handling

## Key Dependencies

**Critical:**
- `github.com/urfave/cli/v2` v2.27.1 - CLI framework, command routing
- `github.com/AlecAivazis/survey/v2` v2.3.7 - Interactive CLI prompts

**Infrastructure (Transitive):**
- `github.com/mattn/go-colorable` v0.1.13 - ANSI color output
- `github.com/mattn/go-isatty` v0.0.19 - TTY detection
- `github.com/mgutz/ansi` v0.0.0-20200706080929-d51e80ef957d - ANSI color conversion
- `golang.org/x/sys` v0.15.0 - System calls
- `golang.org/x/text` v0.14.0 - Text processing

## Configuration

**Build Configuration:**
- `go.mod` - Module definition with Go version 1.21
- `Makefile` - Build script with version injection via ldflags

**Build Flags (from Makefile):**
```
-X 'cli-aio/cmd/version.Version=$(git describe --tags --always --dirty)'
-X 'cli-aio/cmd/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)'
-X 'cli-aio/cmd/version.GitCommit=$(git rev-parse --short HEAD)'
```

**Application Configuration:**
- Projects stored in: `~/.config/cli-aio/projects.json`
- Shell wrapper configs: `~/.zshrc`, `~/.bashrc`, `~/.profile`, `~/.config/fish/functions/`

**Environment Variables:**
- `SHELL` - Used for shell detection (zsh, bash, fish, ksh)
- `GITLAB_PRIVATE_TOKEN` - Required for GitLab release creation

## Platform Requirements

**Development:**
- Go 1.21+
- git (for version info during build)
- Standard Unix tools (mkdir, cat, etc.)

**Production:**
- Compiled binary (no runtime dependencies)
- Target: `/usr/local/bin/aio` (from Makefile)

---

*Stack analysis: 2026-02-26*
