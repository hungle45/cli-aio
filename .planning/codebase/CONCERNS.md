# Codebase Concerns

**Analysis Date:** 2026-02-26

## Tech Debt

### Hardcoded Project Configuration

**Issue:** Project environment mapping is hardcoded in `cmd/ztag/command.go`
- Files: `cmd/ztag/command.go`
- Impact: Only works for specific internal projects; cannot be extended without code changes
- Fix approach: Move to configuration file (e.g., `~/.config/cli-aio/projects.yaml`)

### Unsafe Error Handling with Panic

**Issue:** Uses `panic()` for error handling in tag parsing
- Files: `cmd/ztag/tag.go` (line 117-125)
- Impact: CLI crashes instead of graceful error handling
- Fix approach: Return error instead of panicking

```go
// Current (unsafe):
func mustAtoi(s string) int {
    // ...
    panic(fmt.Sprintf("Regex matched non-integer value: %s", s))
}

// Better: Return error
```

### Incomplete Error Handling in ztag

**Issue:** QC environment silently ignores potential errors
- Files: `cmd/ztag/command.go` (lines 124-126)
- Impact: Release may be incomplete but user not warned
- Fix approach: Add proper error handling or user confirmation

### Shell Command Mixing

**Issue:** `cmd/git/command.go` uses raw `exec.Command` instead of git package functions
- Files: `cmd/git/command.go` (line 234)
- Impact: Inconsistent error handling, harder to test
- Fix approach: Add `CheckoutTrackingBranch` function to `internal/pkg/git/git.go`

---

## Security Considerations

### Unchecked API Response

**Issue:** GitLab release creation doesn't verify API response
- Files: `internal/pkg/git/git.go` (lines 126-133)
- Risk: Release could fail silently; no validation of HTTP response
- Current mitigation: None
- Recommendations: Parse curl output, check HTTP status code

### Token Handling in Command

**Issue:** GitLab token passed via command line (visible in process list)
- Files: `internal/pkg/git/git.go` (lines 126-129)
- Risk: Token visible in process arguments
- Current mitigation: Uses environment variable (good)
- Recommendations: Use GitLab Go client library instead of curl

### No Input Validation

**Issue:** Jira ticket input not validated
- Files: `cmd/ztag/command.go` (line 128)
- Risk: Arbitrary strings passed to release API
- Recommendations: Add regex validation for ticket format

### Shell Wrapper Injection

**Issue:** Shell config modification writes to user RC files
- Files: `cmd/prj/install.go`
- Risk: Malformed config could break shell startup
- Current mitigation: Uses markers to detect existing installation
- Recommendations: Add backup before modification

---

## Performance Bottlenecks

### Tag Retrieval Fetches All Tags

**Issue:** Gets all remote tags before limiting
- Files: `internal/pkg/git/git.go` (lines 79-108)
- Problem: Slow with many tags; fetches entire tag list
- Improvement path: Use `--limit` flag with git ls-remote

### Directory Walking Not Parallelized

**Issue:** `FindGitRepos` walks sequentially
- Files: `internal/pkg/project/project.go` (lines 89-117)
- Problem: Slow on large directory trees
- Improvement path: Use goroutines for parallel directory scanning

### Repeated Git Process Spawning

**Issue:** Multiple individual git commands instead of batching
- Files: `internal/pkg/git/git.go`
- Problem: Process spawn overhead for each operation
- Improvement path: Consider git2go library for embedded git operations

---

## Fragile Areas

### Shell Detection Reliability

**Issue:** Relies on `$SHELL` environment variable which may not match actual shell
- Files: `cmd/prj/install.go` (line 51)
- Why fragile: Users may run different shell than login shell
- Safe modification: Always allow shell flag override
- Test coverage: Limited - no test for shell detection

### Git Remote Assumptions

**Issue:** Assumes `origin` remote always exists
- Files: `internal/pkg/git/git.go` (multiple functions)
- Why fragile: Commands fail without clear error if no remote
- Safe modification: Add remote name parameter with default

### Interactive Mode Inconsistency

**Issue:** Error handling differs between interactive functions
- Files: `internal/prompt/prompt.go`
- Why fragile: Some errors show help, others return silently
- Safe modification: Standardize error handling in SelectCommand

### Fish Shell Function Overwrite

**Issue:** Writes directly to Fish functions directory
- Files: `cmd/prj/install.go` (line 78)
- Why fragile: Could overwrite existing user function
- Safe modification: Check if function exists before writing

---

## Test Coverage Gaps

### No Test Files

**Issue:** No test files found in the codebase
- What's not tested: All command logic, git operations, project management
- Files: Entire codebase
- Risk: Silent failures in edge cases
- Priority: High

### Untested Edge Cases

**What's not tested:**
- Empty git repositories
- Repositories with no remote
- Network failures during git operations
- Invalid project paths
- Concurrent project file access
- Shell detection on different platforms

---

## Dependencies at Risk

### survey/v2 (v2.3.7)

**Issue:** Last updated 2021, potential compatibility issues
- Risk: May not work with newer Go versions
- Impact: Interactive prompts may fail
- Migration plan: Consider migrate to newer alternative or maintain fork

### urfave/cli/v2 (v2.27.1)

**Status:** Actively maintained, but v3 is available
- Risk: Using v2 limits access to v3 features
- Migration plan: Update to v3 with minimal code changes

---

## Missing Critical Features

### No Configuration System

**Problem:** No centralized configuration
- Blocks: Customization of project mappings, default behaviors
- Recommendation: Add YAML/JSON config file support

### No Logging System

**Problem:** No structured logging
- Blocks: Debugging production issues
- Recommendation: Add structured logger (zap, zerolog)

### No Undo/Rollback

**Problem:** Destructive operations cannot be undone
- Blocks: Safe experimentation with git operations
- Recommendation: Add transaction-like behavior or clear warnings

### No Dry-Run Mode

**Problem:** Cannot preview destructive operations
- Blocks: Understanding what commands will do
- Recommendation: Add `--dry-run` flag to git commands

---

## Code Quality Issues

### Duplicate Code

**Issue:** Branch filtering logic duplicated in multiple places
- Files: `cmd/git/command.go` (lines 76-82, 107-113)
- Pattern: Same "filter out current branch" logic repeated

### Magic Strings

**Issue:** Hardcoded strings throughout codebase
- Examples: "origin", ".git", marker strings
- Recommendation: Extract to constants

### Inconsistent Error Messages

**Issue:** Error formats vary between commands
- Examples: Different prefixes ([!], [-], error:)
- Recommendation: Standardize error formatting

---

*Concerns audit: 2026-02-26*
