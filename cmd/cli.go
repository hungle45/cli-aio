package cmd

import (
	"cli-aio/cmd/gencmd"
	"cli-aio/cmd/git"
	"cli-aio/cmd/version"
	"cli-aio/cmd/ztag"
	"cli-aio/internal/prompt"
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"
)

// findCommand recursively searches for a command in the command tree
func findCommand(commands []*cli.Command, path []string) (*cli.Command, []string) {
	if len(path) == 0 {
		return nil, nil
	}

	target := path[0]
	for _, cmd := range commands {
		if cmd.Name == target {
			if len(path) == 1 {
				// Found the target command
				return cmd, nil
			}
			// Continue searching in subcommands
			if len(cmd.Subcommands) > 0 {
				return findCommand(cmd.Subcommands, path[1:])
			}
			// Command found but no subcommands, return it with remaining path
			return cmd, path[1:]
		}
	}
	// Command not found, return available commands
	return nil, path
}

// getCommandPath extracts the command path from context
func getCommandPath(c *cli.Context) []string {
	var path []string
	// Get all args (command + subcommands)
	args := c.Args().Slice()
	if len(args) > 0 {
		path = args
	}
	return path
}

// showUnknownCommandWarning displays a warning for unknown commands/subcommands
func showUnknownCommandWarning(c *cli.Context, commands []*cli.Command, isSubcommand bool) {
	path := getCommandPath(c)
	if len(path) == 0 {
		return
	}

	var commandPath string
	var availableCommands []*cli.Command
	var parentCommand *cli.Command
	var parentPath []string
	actualIsSubcommand := isSubcommand

	// Try to find where the command path breaks
	for i := len(path); i > 0; i-- {
		parentCommand, remaining := findCommand(commands, path[:i])
		if parentCommand != nil {
			if len(remaining) > 0 {
				// Found a valid parent, but remaining path is invalid
				commandPath = strings.Join(path, " ")
				availableCommands = parentCommand.Subcommands
				parentPath = path[:i]
				actualIsSubcommand = true
				break
			} else if i < len(path) {
				// Found a valid command, but there's more in the path
				commandPath = strings.Join(path, " ")
				availableCommands = parentCommand.Subcommands
				parentPath = path[:i]
				actualIsSubcommand = true
				break
			}
		}
	}

	// If no parent found, it's a top-level unknown command
	if parentCommand == nil {
		commandPath = path[0]
		availableCommands = commands
		actualIsSubcommand = false
	}

	fmt.Fprintf(os.Stderr, "[!] Unknown %s '%s'\n",
		map[bool]string{true: "subcommand", false: "command"}[actualIsSubcommand],
		commandPath)

	if len(availableCommands) > 0 {
		if actualIsSubcommand {
			fmt.Fprintf(os.Stderr, "\nAvailable subcommands:\n")
		} else {
			fmt.Fprintf(os.Stderr, "\nAvailable commands:\n")
		}
		for _, cmd := range availableCommands {
			fmt.Fprintf(os.Stderr, "  %s - %s\n", cmd.Name, cmd.Usage)
		}
	}

	if actualIsSubcommand && len(parentPath) > 0 {
		fmt.Fprintf(os.Stderr, "\nUse 'cli-aio %s --help' for more information.\n", strings.Join(parentPath, " "))
	} else {
		fmt.Fprintf(os.Stderr, "\nUse 'cli-aio --help' for more information.\n")
	}
}

// Execute initializes and runs the CLI application.
// This is the central wiring point where all commands are registered.
// To add a new command:
//  1. Create a new package under cmd/ (e.g., cmd/mycommand/)
//  2. Implement a Command() function that returns *cli.Command
//  3. Import the package here and add it to the Commands slice
func Execute() error {
	commands := []*cli.Command{
		version.Command(),
		ztag.Command(),
		git.Command(),
		gencmd.Command(),
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
			// If command is provided via args, check if it's a valid command path
			if c.Args().Len() > 0 {
				path := getCommandPath(c)
				foundCmd, remaining := findCommand(commands, path)

				// If command found and no remaining path, it's valid
				if foundCmd != nil && len(remaining) == 0 {
					// Valid command, let cli handle it
					return nil
				}

				// Unknown command - show warning
				showUnknownCommandWarning(c, commands, false)
				return fmt.Errorf("unknown command: %s", strings.Join(path, " "))
			}

			// Use the helper function - it automatically handles interactive mode detection
			// and extracts command names from the commands slice
			return prompt.SelectCommand(c, commands, "Select a command:", cli.ShowAppHelp)
		},
		// OnUsageError is called when an unknown command or flag is used
		// This handles both top-level commands and subcommands automatically
		OnUsageError: func(c *cli.Context, err error, isSubcommand bool) error {
			showUnknownCommandWarning(c, commands, isSubcommand)
			return err
		},
		ExitErrHandler: func(c *cli.Context, err error) {
			if err == nil {
				return
			}

			// Check if this is an unknown command error (in case it wasn't caught by Action)
			errMsg := err.Error()
			if strings.Contains(errMsg, "unknown command") {
				// Warning already shown by Action handler, just exit
				os.Exit(1)
			}

			// For other errors, show the error message
			fmt.Fprintf(os.Stderr, "[-] Error: %v\n", err)
			os.Exit(1)
		},
	}

	return app.Run(os.Args)
}
