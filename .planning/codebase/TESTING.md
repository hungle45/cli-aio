# Testing Patterns

**Analysis Date:** 2026-02-26

## Test Framework

**Status:** No tests currently exist in the codebase

**Go Version:** 1.21

**Test Discovery:** No `*_test.go` files found in the project

### Recommended Framework

**Test Runner:**
- `testing` - Go's built-in testing package
- `go test ./...` - Run all tests

**Assertion Library (recommended for future):**
- `testify/assert` - For more readable assertions
- Alternative: Use standard Go comparisons with custom helpers

## Test File Organization

**Current State:**
- No test files exist
- No test directory structure defined

**Recommended Structure:**

```
cmd/
├── ztag/
│   ├── command.go
│   ├── command_test.go      # Tests for ztag command
│   └── tag_test.go          # Tests for tag generation logic
├── git/
│   ├── command.go
│   └── command_test.go
internal/
├── pkg/
│   ├── git/
│   │   ├── git.go
│   │   └── git_test.go      # Tests for git operations
│   └── project/
│       ├── project.go
│       └── project_test.go  # Tests for project management
```

### Naming Conventions
- Test files: `*_test.go` (Go standard)
- Test functions: `Test<FunctionName>_<Scenario>` or `Test<UnitName>`
  - Examples: `TestCheckIfGitRepo`, `TestGenerateNextTag_BugLevel`
- Test files co-located with implementation: Same directory as source files

## Test Structure

### Basic Test Structure (Go standard)

```go
package ztag

import (
    "testing"
)

func TestGenerateNextTag(t *testing.T) {
    tests := []struct {
        name     string
        oldTag   string
        level    Level
        env      Env
        want     string
        wantErr  bool
    }{
        {
            name:   "bug level increments patch",
            oldTag: "qc-v1.0.0",
            level:  LevelBug,
            env:    EnvQC,
            want:   "qc-v1.0.1",
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := GenerateNextTag(tt.oldTag, tt.level, tt.env)
            if (err != nil) != tt.wantErr {
                t.Errorf("GenerateNextTag() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("GenerateNextTag() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Table-Driven Tests

The codebase should use table-driven tests for functions with multiple test cases:

```go
func TestTagComponents_Next(t *testing.T) {
    tests := []struct {
        name   string
        c      TagComponents
        level  Level
        want   TagComponents
    }{
        {
            name:   "major increments major and resets minor/patch",
            c:      TagComponents{Major: 1, Minor: 2, Patch: 3},
            level:  LevelMajor,
            want:   TagComponents{Major: 2, Minor: 0, Patch: 0},
        },
        {
            name:   "minor increments minor and resets patch",
            c:      TagComponents{Major: 1, Minor: 2, Patch: 3},
            level:  LevelMinor,
            want:   TagComponents{Major: 1, Minor: 3, Patch: 0},
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.c.Next(tt.level); got != tt.want {
                t.Errorf("Next() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Mocking

**Framework:** No mocking framework currently in use

### Recommended Patterns for Future Tests

**1. Interface-based dependencies:**
```go
// Define interface for git operations
type GitRunner interface {
    CheckIfGitRepo() (bool, error)
    GetCurrentBranch() (string, error)
    // ...
}

// Use interface in commands
type GitCommand struct {
    runner GitRunner
}

func NewGitCommand(runner GitRunner) *GitCommand {
    return &GitCommand{runner: runner}
}
```

**2. For external commands (os/exec):**
```go
// Wrap exec.Command in a function for testing
type CommandRunner func(name string, args ...string) *exec.Cmd

func runGitCommand(runner CommandRunner, args ...string) error {
    cmd := runner("git", args...)
    // ...
}

// In tests, provide mock runner
func TestGitOperations(t *testing.T) {
    mockRunner := func(name string, args ...string) *exec.Cmd {
        // Return mocked command
    }
    // Use mockRunner
}
```

**3. File system operations:**
```go
// For project.go file operations
type FileSystem interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    // ...
}

type RealFileSystem struct{}

func (rfs *RealFileSystem) ReadFile(path string) ([]byte, error) {
    return os.ReadFile(path)
}

// Default implementation
var fs FileSystem = &RealFileSystem{}

// In tests, replace with mock
```

### What to Mock
- **External commands:** `git` CLI calls via `os/exec`
- **File system operations:** Reading/writing config files
- **Environment variables:** `os.Getenv` calls
- **Network calls:** HTTP requests (e.g., GitLab API)

### What NOT to Mock
- **Pure business logic:** Tag generation, version incrementing
- **Standard library:** String operations, simple type conversions
- **Built-in functionality:** Internal package functions that have no dependencies

## Fixtures and Factories

### Test Data Location
- Define test fixtures in `_test.go` files within the same package
- For complex fixtures, consider a `fixtures.go` file in test directory

### Example Fixture Pattern
```go
// Test data for project tests
var testProjects = []Project{
    {
        Name: "project-a",
        Path: "/home/user/projects/project-a",
    },
    {
        Name: "project-b",
        Path: "/home/user/projects/project-b",
    },
}

func newTestProject(name, path string) Project {
    return Project{
        Name: name,
        Path: path,
    }
}
```

### Tag Test Fixtures
```go
var tagTestCases = []struct {
    tag    string
    valid  bool
}{
    {"qc-v1.0.0", true},
    {"stg-v2.3.4", true},
    {"v1.0.0", true},
    {"invalid", false},
}
```

## Coverage

**Current State:** No coverage requirements enforced

### Recommendations

**Target Coverage:**
- Business logic (tag generation, version incrementing): 80%+
- Package utilities: 70%+
- Command handlers: 60%+ (may have integration tests instead)

**View Coverage:**
```bash
# Run tests with coverage
go test -cover ./...

# View coverage for specific package
go test -coverprofile=coverage.out ./cmd/ztag/
go tool cover -html=coverage.out
```

**Enforce Coverage (optional):**
```bash
# Add to Makefile or CI
go test -cover ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -1
```

## Test Types

### Unit Tests
- **Scope:** Individual functions and methods
- **Focus:** Tag generation, version components, project management logic
- **Examples:**
  - `TestGenerateNextTag` - Tag generation with various levels
  - `TestTagComponents_Next` - Version incrementing
  - `TestProject_Add` - Project list manipulation

### Integration Tests
- **Scope:** Command execution with real dependencies
- **Focus:** CLI commands, file system, git operations
- **Examples:**
  - Full command execution tests
  - Project config file read/write tests
  - Interactive prompt flow tests

### E2E Tests
- **Not currently implemented**
- **Recommendation:** Use Go's integration test patterns with `go test -run` for CLI

## Common Patterns

### Async Testing
- Not applicable - CLI is synchronous
- For async operations, use goroutine + `t.Run` with subtests

### Error Testing
```go
func TestGenerateNextTag_InvalidTag(t *testing.T) {
    _, err := GenerateNextTag("invalid-tag", LevelBug, EnvQC)
    if err == nil {
        t.Error("GenerateNextTag() expected error for invalid tag")
    }
}
```

### Testing Private Functions
- **Option 1:** Test via exported functions that use them
- **Option 2:** Place tests in same package (common Go practice)
  ```go
  // In same package
  func TestMustAtoi_Panic(t *testing.T) {
      defer func() {
          if r := recover(); r == nil {
              t.Error("Expected panic for non-numeric string")
          }
      }()
      mustAtoi("not-a-number")
  }
  ```

---

*Testing analysis: 2026-02-26*
