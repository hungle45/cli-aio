package cmd

import (
	"cli-aio/cmd/git"
	"cli-aio/cmd/version"
	"cli-aio/cmd/ztag"
	"cli-aio/internal/prompt"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

// Execute initializes and runs the CLI application.
// This is the central wiring point where all commands are registered.
// To add a new command:
//  1. Create a new package under cmd/ (e.g., cmd/mycommand/)
//  2. Implement a Command() function that returns *cli.Command
//  3. Import the package here and add it to the Commands slice
func Execute() error {
	commands := []*cli.Command{
		// example.Command(),
		version.Command(),
		ztag.Command(),
		git.Command(),
	}

	app := &cli.App{
		Name:  "cli-aio",
		Usage: "A modular CLI application built with urfave/cli",
		// Commands are registered here. Each command is self-contained
		// in its own package, preventing tight coupling.
		Commands: commands,
		// Global flags can be added here if needed
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "interactive",
				Aliases: []string{"i"},
				Usage:   "Force enable interactive mode (auto-enabled when params missing)",
				Value:   false,
			},
		},
		// Action is called when no command is provided.
		// It allows interactive selection of commands.
		// Uses the SelectCommand helper which automatically extracts command names.
		Action: func(c *cli.Context) error {
			// If command is provided via args, let cli handle it
			if c.Args().Len() > 0 {
				return nil
			}

			// Use the helper function - it automatically handles interactive mode detection
			// and extracts command names from the commands slice
			return prompt.SelectCommand(c, commands, "Select a command:", cli.ShowAppHelp)
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			if err == nil {
				return
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		},
	}

	return app.Run(os.Args)
}
