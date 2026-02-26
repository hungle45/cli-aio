# aio — All-in-One CLI

A modular command-line toolbox built with [urfave/cli](https://github.com/urfave/cli).

## Installation

```sh
make build-cmd   # builds and installs the binary to /usr/local/bin/aio
```

---

## Commands

### `version` — Show build info

```sh
aio version
```

Prints version, build time, git commit, Go version, and OS/arch. Values are
injected at build time via `ldflags`.

---

### `git` — Git helpers

```sh
aio git [subcommand]
```

| Subcommand | Description |
|---|---|
| `fname` | Print the full project name extracted from the git remote URL |
| `rmerge` | Reverse-merge: checkout target branch, pull, then merge current branch into it |
| `ckl` | Checkout list — fuzzy-select from all local + remote branches and check it out |

#### Examples

```sh
aio git fname
aio git rmerge main        # or omit arg to pick branch interactively
aio git ckl
```

---

### `ztag` — Git tag generator (ZaloPay convention)

```sh
aio ztag [subcommand] [--level b|m|M]
```

Generates and pushes the next semver-style tag for a deployment environment.
Run without a subcommand to auto-detect environment from the project path, or
pick it interactively.

| Subcommand | Description |
|---|---|
| `qc` | Tag for QC environment |
| `stg` | Tag for Staging environment (requires Jira ticket) |
| `prod` | Tag for Production environment — must be on `main`/`master` (requires Jira ticket) |

#### Flags

| Flag | Default | Description |
|---|---|---|
| `--level, -l` | `b` | Bump level: `b` = bug/patch, `m` = minor, `M` = major |

#### Examples

```sh
aio ztag qc
aio ztag stg --level m
aio ztag prod --level M
```

---

### `gencmd` — Command scaffolder

```sh
aio gencmd [command-name] [--subcommand sub1] [--subcommand sub2] [--usage "description"]
```

Scaffolds a new command package under `cmd/<name>/command.go` and
auto-registers it in `cmd/cli.go`. Runs interactively if arguments are omitted.

#### Flags

| Flag | Description |
|---|---|
| `--subcommand, -s` | Subcommand names to generate (repeatable) |
| `--usage, -u` | Usage description string |

#### Examples

```sh
aio gencmd myfeature
aio gencmd myfeature -s list -s create -u "Manage my features"
```

---

### `prj` — Project manager

Manage a list of projects on your laptop. Projects are stored in
`~/.config/cli-aio/projects.json`.

```sh
aio prj [subcommand]
```

| Subcommand | Description |
|---|---|
| `cd` | Fuzzy-select a project and navigate to it (via shell wrapper) |
| `add [path]` | Add a folder as a project |
| `add-git [path]` | Recursively scan a folder for git repos and add them all |
| `config` | Open `projects.json` in your editor |
| `install` | Install the `prj` shell wrapper into your shell's rc file |

---

#### `prj install` — First-time setup

Run this once after installing the binary:

```sh
aio prj install
```

It detects your shell and appends the `prj` wrapper function to the correct
config file:

| Shell | Config file |
|---|---|
| zsh | `~/.zshrc` |
| bash | `~/.bashrc` or `~/.bash_profile` |
| fish | `~/.config/fish/functions/prj.fish` |
| ksh | `~/.kshrc` |
| other | `~/.profile` |

Override shell detection with `--shell`:

```sh
aio prj install --shell bash
```

Then reload your shell:

```sh
exec zsh   # or exec bash, etc.
```

---

#### `prj` — Navigate to a project

After installing the wrapper, just type:

```sh
prj
```

A fuzzy-searchable list appears:

```
? Select a project:  [Use arrows to move, type to filter]
  bank-config-fe-v2        ~/Workspace/ZaloPay/Bank/operation/bank-config-fe-v2
  cli-aio                  ~/Workspace/nothing/cli-aio
  blog                     ~/Workspace/nothing/blog
```

Select a project and your terminal changes to that directory.

> **Note:** `aio prj cd` must be called via the `prj` wrapper — calling it
> directly will show an error and hint to run `aio prj install`.

---

#### `prj add` — Add a single project

```sh
aio prj add ~/Workspace/myproject
aio prj add   # prompts interactively
```

Supports `~` expansion. Errors if the path doesn't exist or is not a directory.
Skips silently if the path is already in the list.

---

#### `prj add-git` — Bulk-add git repos

```sh
aio prj add-git ~/Workspace/ZaloPay
aio prj add-git   # prompts interactively
```

Recursively walks the given folder and adds every directory that contains a
`.git` entry. Does **not** descend into a found repo (submodules are skipped).
Hidden directories (`.cache`, `.git`, etc.) are skipped.

---

#### `prj config` — Edit project list directly

```sh
aio prj config
```

Opens `~/.config/cli-aio/projects.json` in `$EDITOR`. Falls back to the first
available editor in: `nvim` → `vim` → `nano` → `vi` → `notepad`.

---

## Storage

| Path | Purpose |
|---|---|
| `~/.config/cli-aio/projects.json` | Project list for `prj` |

---

## Global flags

```sh
aio --interactive   # -i  force interactive mode for any command
```

---

## Adding a new command

```sh
aio gencmd mycommand -s list -s create -u "My command description"
```

Or manually:
1. Create `cmd/mycommand/command.go` with a `Command() *cli.Command` function
2. Import it in `cmd/cli.go` and add `mycommand.Command()` to the `commands` slice
3. Rebuild: `make build-cmd`
