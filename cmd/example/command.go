package example

import (
	"cli-aio/internal/prompt"

	"github.com/urfave/cli/v2"
)

// Command returns the example command with its subcommands.
// This demonstrates how to structure a command with nested subcommands.
// Each subcommand is defined in a separate file for better organization.
//
// To add a new subcommand:
//  1. Create a new function in this package that returns *cli.Command
//  2. Add it to the Subcommands slice below
//     That's it! The SelectCommand helper automatically handles the rest.
func Command() *cli.Command {
	subcommands := []*cli.Command{
		createSubcommand(),
		listSubcommand(),
		deleteSubcommand(),
	}

	return &cli.Command{
		Name:  "example",
		Usage: "Example command demonstrating subcommands",
		// Subcommands are registered here. Each subcommand is self-contained
		// and can be easily added or removed without affecting others.
		Subcommands: subcommands,
		// Action is called when no subcommand is provided.
		// It allows interactive selection of subcommands.
		// Uses the SelectCommand helper which automatically extracts subcommand names.
		Action: func(c *cli.Context) error {
			// If subcommand is provided via args, let cli handle it
			if c.Args().Len() > 0 {
				return nil
			}

			// Use the helper function - it automatically handles interactive mode detection
			// and extracts subcommand names from the subcommands slice
			return prompt.SelectCommand(c, subcommands, "Select a subcommand:", cli.ShowSubcommandHelp)
		},
	}
}
