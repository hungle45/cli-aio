package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// ValidateSubcommand checks if a subcommand exists and shows a warning if not.
// This can be used in command Action handlers to validate subcommands.
// Returns true if subcommand is valid or no subcommand provided, false if invalid.
func ValidateSubcommand(c *cli.Context, subcommands []*cli.Command) bool {
	if c.Args().Len() == 0 {
		return true
	}

	subcmdName := c.Args().First()
	for _, subcmd := range subcommands {
		if subcmd.Name == subcmdName {
			return true
		}
	}

	// Unknown subcommand
	fmt.Fprintf(os.Stderr, "⚠️  Warning: Unknown subcommand '%s'\n", subcmdName)
	fmt.Fprintf(os.Stderr, "\nAvailable subcommands:\n")
	for _, subcmd := range subcommands {
		fmt.Fprintf(os.Stderr, "  %s - %s\n", subcmd.Name, subcmd.Usage)
	}
	fmt.Fprintf(os.Stderr, "\nUse 'cli-aio %s --help' for more information.\n", c.Command.Name)
	return false
}

