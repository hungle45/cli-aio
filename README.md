# aio — All-in-One CLI

A Go CLI tool with useful commands for daily development.

## Install

```sh
make build-cmd
```

This installs `aio` to `/usr/local/bin/aio`.

---

## Build Info

```sh
aio version
```

Shows version, build time, git commit, Go version, and OS/arch.

---

## Git Helpers

### Get project name
```sh
aio git fname
```
Prints the full project name from your git remote URL.

### Reverse merge
```sh
aio git rmerge main
```
Checks out the target branch, pulls it, then merges your current branch into it.

### Checkout branch
```sh
aio git ckl
```
Fuzzy-select from all local and remote branches and check it out.

---

## Tagging

```sh
aio ztag qc      # Tag for QC
aio ztag stg    # Tag for Staging
aio ztag prod   # Tag for Production (must be on main)
```

Flags: `-l b|m|M` — bump level (b=patch, m=minor, M=major)

---

## Create Commands

```sh
aio gencmd mytool -s list -s create -u "Manage my tools"
```

Scaffolds a new command under `cmd/mytool/` and registers it automatically.

---

## Project Navigation

### First-time setup

```sh
aio prj install
exec zsh   # restart shell
```

### Navigate to a project

```sh
prj
```

Fuzzy-search your project list and jump to it.

### Add projects

```sh
aio prj add ~/path/to/project
aio prj add-git ~/folder    # Scan folder for git repos
```

### Edit project list

```sh
aio prj config
```

---

## Global Flags

```sh
aio --interactive   # Force interactive mode
aio -i
```
