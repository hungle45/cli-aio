package version

import (
	"fmt"
	"runtime"

	"github.com/urfave/cli/v2"
)

var (
	// Version can be set at build time using:
	// go build -ldflags "-X cli-aio/cmd/version.Version=1.0.0"
	Version = "dev"
	// BuildTime can be set at build time using:
	// go build -ldflags "-X cli-aio/cmd/version.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
	BuildTime = "unknown"
	// GitCommit can be set at build time using:
	// go build -ldflags "-X cli-aio/cmd/version.GitCommit=$(git rev-parse --short HEAD)"
	GitCommit = "unknown"
)

// Command returns a simple version command.
// This demonstrates a minimal command without subcommands.
// Each command package is self-contained and can be easily
// added or removed from the main CLI without affecting others.
func Command() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Show version information",
		Action: func(c *cli.Context) error {
			fmt.Printf("Version: %s\n", Version)
			fmt.Printf("Build Time: %s\n", BuildTime)
			fmt.Printf("Git Commit: %s\n", GitCommit)
			fmt.Printf("Go Version: %s\n", runtime.Version())
			fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return nil
		},
	}
}
