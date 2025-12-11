package example

import (
	"cli-aio/internal/prompt"
	"fmt"

	"github.com/urfave/cli/v2"
)

// createSubcommand demonstrates a subcommand with flags and interactive mode.
// This pattern keeps each subcommand isolated and testable.
// Interactive mode prompts for missing required flags instead of failing.
func createSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new resource",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "name",
				Aliases: []string{"n"},
				Usage:   "Name of the resource",
				// Not required - will prompt if missing in interactive mode
			},
			&cli.StringFlag{
				Name:    "type",
				Aliases: []string{"t"},
				Usage:   "Type of the resource",
				Value:   "default",
			},
		},
		Action: func(c *cli.Context) error {
			interactive := c.Bool("interactive")
			name := c.String("name")
			resourceType := c.String("type")

			// Auto-enable interactive mode if parameters are missing
			hasMissingParams := name == ""
			useInteractive := prompt.ShouldUseInteractive(interactive, hasMissingParams)

			// Interactive mode: prompt for missing values
			if useInteractive {
				if name == "" {
					var err error
					name, err = prompt.Input("Enter resource name:", "", true)
					if err != nil {
						return fmt.Errorf("failed to get resource name: %w", err)
					}
				}

				// Prompt for type with selection
				if resourceType == "default" {
					typeOptions := []string{"default", "custom", "premium", "enterprise"}
					_, selectedType, err := prompt.Select("Select resource type:", typeOptions, "default")
					if err != nil {
						return fmt.Errorf("failed to select resource type: %w", err)
					}
					resourceType = selectedType
				}
			} else {
				// Non-interactive mode: validate required fields
				if name == "" {
					return fmt.Errorf("name is required (use --name flag or run in interactive mode)")
				}
			}

			fmt.Printf("Creating resource: %s (type: %s)\n", name, resourceType)
			// Add your implementation here
			return nil
		},
	}
}

// listSubcommand demonstrates a simple subcommand with optional flags and interactive mode.
func listSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List all resources",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Show all resources, including hidden ones",
			},
			&cli.StringFlag{
				Name:    "filter",
				Aliases: []string{"f"},
				Usage:   "Filter resources by pattern",
			},
		},
		Action: func(c *cli.Context) error {
			interactive := c.Bool("interactive")
			showAll := c.Bool("all")
			filter := c.String("filter")

			// Auto-enable interactive mode if we want to prompt for optional values
			// (list command doesn't require params, but can benefit from interactive prompts)
			useInteractive := prompt.ShouldUseInteractive(interactive, false) || interactive

			// Interactive mode: prompt for options
			if useInteractive {
				if !c.IsSet("all") {
					var err error
					showAll, err = prompt.Confirm("Show all resources (including hidden)?", false)
					if err != nil {
						return fmt.Errorf("failed to get confirmation: %w", err)
					}
				}

				if filter == "" && !c.IsSet("filter") {
					var err error
					filter, err = prompt.Input("Enter filter pattern (leave empty for no filter):", "", false)
					if err != nil {
						return fmt.Errorf("failed to get filter: %w", err)
					}
				}
			}

			fmt.Printf("Listing resources (all: %v, filter: %s)\n", showAll, filter)
			// Add your implementation here
			return nil
		},
	}
}

// deleteSubcommand demonstrates a subcommand with confirmation and interactive mode.
func deleteSubcommand() *cli.Command {
	return &cli.Command{
		Name:  "delete",
		Usage: "Delete a resource",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "id",
				Aliases: []string{"i"},
				Usage:   "ID of the resource to delete",
				// Not required - will prompt if missing in interactive mode
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Force deletion without confirmation",
			},
		},
		Action: func(c *cli.Context) error {
			interactive := c.Bool("interactive")
			id := c.String("id")
			force := c.Bool("force")

			// Auto-enable interactive mode if parameters are missing
			hasMissingParams := id == ""
			useInteractive := prompt.ShouldUseInteractive(interactive, hasMissingParams)

			// Interactive mode: prompt for missing values
			if useInteractive {
				if id == "" {
					// In a real scenario, you might fetch available resources and let user select
					availableIDs := []string{"res-001", "res-002", "res-003", "res-004"}
					_, selectedID, err := prompt.Select("Select resource to delete:", availableIDs, "")
					if err != nil {
						return fmt.Errorf("failed to select resource: %w", err)
					}
					id = selectedID
				}

				if !force {
					confirmed, err := prompt.Confirm(fmt.Sprintf("Are you sure you want to delete resource '%s'?", id), false)
					if err != nil {
						return fmt.Errorf("failed to get confirmation: %w", err)
					}
					if !confirmed {
						fmt.Println("Deletion cancelled.")
						return nil
					}
				}
			} else {
				// Non-interactive mode: validate required fields
				if id == "" {
					return fmt.Errorf("id is required (use --id flag or run in interactive mode)")
				}
				if !force {
					fmt.Printf("Would delete resource: %s (use --force to confirm or run in interactive mode)\n", id)
					return nil
				}
			}

			fmt.Printf("Deleting resource: %s\n", id)
			// Add your implementation here
			return nil
		},
	}
}
