# External Integrations

**Analysis Date:** 2026-02-26

## APIs & External Services

**Git Operations:**
- Git (local) - All git operations via `exec.Command`
  - Implementation: `internal/pkg/git/git.go`
  - Commands used: `git rev-parse`, `git config`, `git ls-remote`, `git tag`, `git push`, `git checkout`, `git pull`, `git merge`, `git fetch`, `git branch`, `git show-ref`, `git merge-base`

**GitLab API (Zalopay Internal):**
- GitLab instance: `https://gitlab.zalopay.vn`
  - Purpose: Create releases for projects
  - Implementation: `internal/pkg/git/git.go` - `CreateZalopayRelease()` function
  - Endpoint: `POST /api/v4/projects/{project_id}//releases`
  - Auth: `GITLAB_PRIVATE_TOKEN` environment variable (PRIVATE-TOKEN header)
  - Used by: `cmd/ztag/command.go` for stg/prod environment releases

**Jira Integration:**
- Jira tickets - User input for release documentation
  - Purpose: Associate releases with Jira tickets
  - Implementation: Interactive prompt in `cmd/ztag/command.go`
  - Required for: stg and prod environment releases (not qc)

## Data Storage

**Local File Storage:**
- Projects configuration: JSON file at `~/.config/cli-aio/projects.json`
  - Schema: Array of `{name, path}` objects
  - Implementation: `internal/pkg/project/project.go`
  - Functions: `Load()`, `Save()`, `Add()`

**Shell Configuration:**
- Shell rc files modified for wrapper installation:
  - `~/.zshrc` - Zsh
  - `~/.bashrc` or `~/.bash_profile` - Bash
  - `~/.profile` - POSIX fallback
  - `~/.config/fish/functions/prj.fish` - Fish shell
  - Implementation: `cmd/prj/install.go`

## Authentication & Identity

**GitLab Token:**
- Type: Personal Access Token (PRIVATE_TOKEN)
- Environment Variable: `GITLAB_PRIVATE_TOKEN`
- Required for: Creating releases via GitLab API
- Scope: Must have API or write access to projects

## Monitoring & Observability

**Error Handling:**
- CLI errors displayed to stderr via `fmt.Fprintf(os.Stderr, ...)`
- Exit codes: 1 for errors, 0 for success

**Logging:**
- Standard output for informational messages (`fmt.Printf`)
- Standard error for errors (`fmt.Fprintf(os.Stderr, ...)`)

## CI/CD & Deployment

**Build System:**
- Makefile-based build
- Binary output: `aio` (moves to `/usr/local/bin/aio`)
- Version info injected at build time via ldflags

**Hosting:**
- Self-hosted CLI tool
- Installation via Makefile to `/usr/local/bin/`

## Environment Configuration

**Required env vars:**
- `GITLAB_PRIVATE_TOKEN` - GitLab API token for release creation
- `SHELL` - Auto-detected, determines which shell config to modify

**Optional env vars:**
- Auto-detected based on terminal capabilities

## Shell Integration

**Shell Wrappers:**
- `prj` function - Quick navigation to saved projects
  - Supports: zsh, bash, fish, ksh
  - Installation: `aio prj install` command
  - Implementation: `cmd/prj/install.go`

**Git Completion:**
- Interactive command/branch selection via survey prompts
- Terminal detection for auto-enabling interactive mode

---

*Integration audit: 2026-02-26*
