package gencmd

import (
	"cli-aio/internal/prompt"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "gencmd",
		Usage: "Generate a new command or subcommand",
		Flags: []cli.Flag{
			&cli.StringSliceFlag{
				Name:    "subcommand",
				Aliases: []string{"s"},
				Usage:   "Subcommand names to generate (can be used multiple times)",
			},
			&cli.StringFlag{
				Name:    "usage",
				Aliases: []string{"u"},
				Usage:   "Usage description for the command",
			},
		},
		Action: func(c *cli.Context) error {
			var cmdName string
			var subcommands []string
			var usage string
			var err error

			// Prompt for command name if not provided
			if c.Args().Len() == 0 {
				cmdName, err = prompt.Input("Enter command name:", "", true)
				if err != nil {
					return fmt.Errorf("command name is required")
				}
			} else {
				cmdName = c.Args().First()
			}

			// Validate command name
			if !isValidCommandName(cmdName) {
				return fmt.Errorf("invalid command name: %s (must contain only alphanumeric characters, hyphens, or underscores)", cmdName)
			}

			// Get subcommands from flags or prompt
			subcommands = c.StringSlice("subcommand")
			if len(subcommands) == 0 {
				// Ask if user wants to add subcommands
				wantsSubcommands, err := prompt.Confirm("Do you want to add subcommands?", false)
				if err != nil {
					// If not in interactive mode, skip subcommands
					wantsSubcommands = false
				}

				if wantsSubcommands {
					fmt.Println("Enter subcommand names (press Enter with empty name to finish):")
					// Prompt for subcommands until user is done
					for i := 1; ; i++ {
						subcmd, err := prompt.Input(fmt.Sprintf("Subcommand %d:", i), "", false)
						if err != nil {
							// If error (e.g., not in TTY), break
							break
						}
						if subcmd == "" {
							break
						}
						// Validate subcommand name
						if !isValidCommandName(subcmd) {
							fmt.Printf("⚠️  Invalid subcommand name: %s (skipping)\n", subcmd)
							continue
						}
						// Check for duplicates
						duplicate := false
						for _, existing := range subcommands {
							if existing == subcmd {
								fmt.Printf("⚠️  Subcommand '%s' already added (skipping)\n", subcmd)
								duplicate = true
								break
							}
						}
						if duplicate {
							continue
						}
						subcommands = append(subcommands, subcmd)
						fmt.Printf("✅ Added subcommand: %s\n", subcmd)
					}
				}
			}

			// Get usage from flag or prompt
			usage = c.String("usage")
			if usage == "" {
				defaultUsage := fmt.Sprintf("%s commands", strings.Title(cmdName))
				usage, err = prompt.Input("Enter usage description:", defaultUsage, false)
				if err != nil {
					// If not in interactive mode, use default
					usage = defaultUsage
				}
				if usage == "" {
					usage = defaultUsage
				}
			}

			return generateCommand(cmdName, subcommands, usage)
		},
	}
}

func generateCommand(cmdName string, subcommands []string, usage string) error {
	// Validate command name (allow alphanumeric, hyphens, underscores)
	if !isValidCommandName(cmdName) {
		return fmt.Errorf("invalid command name: %s (must contain only alphanumeric characters, hyphens, or underscores)", cmdName)
	}

	// Get the workspace root (assuming we're in cmd/generate)
	workspaceRoot := findWorkspaceRoot()
	if workspaceRoot == "" {
		return fmt.Errorf("could not find workspace root")
	}

	cmdDir := filepath.Join(workspaceRoot, "cmd", cmdName)
	cmdFile := filepath.Join(cmdDir, "command.go")

	// Check if command already exists
	if _, err := os.Stat(cmdDir); err == nil {
		return fmt.Errorf("command '%s' already exists at %s", cmdName, cmdDir)
	}

	// Create directory
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate command.go content
	content := generateCommandFile(cmdName, subcommands, usage)
	if err := os.WriteFile(cmdFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write command file: %w", err)
	}

	fmt.Printf("✅ Generated command '%s' at %s\n", cmdName, cmdDir)

	// Update cmd/cli.go to register the new command
	if err := registerCommandInCLI(workspaceRoot, cmdName); err != nil {
		fmt.Printf("⚠️  Warning: Failed to auto-register command in cmd/cli.go: %v\n", err)
		fmt.Printf("   Please manually add: %s.Command() to the commands slice\n", cmdName)
	} else {
		fmt.Printf("✅ Auto-registered command in cmd/cli.go\n")
	}

	return nil
}

func generateCommandFile(cmdName string, subcommands []string, usage string) string {
	packageName := toPackageName(cmdName)
	var imports string
	var subcommandList string
	var actionCode string

	var subcommandFuncs strings.Builder

	if len(subcommands) > 0 {
		imports = `	"cli-aio/internal/cmd"
	"cli-aio/internal/prompt"
	"fmt"

	"github.com/urfave/cli/v2"`
		// Generate subcommand functions
		for _, subcmd := range subcommands {
			funcName := toCamelCase(subcmd)
			subcommandFuncs.WriteString(fmt.Sprintf(`
func create%sCommand() *cli.Command {
	return &cli.Command{
		Name:  "%s",
		Usage: "%s command",
		Action: func(c *cli.Context) error {
			// TODO: Implement your logic here
			fmt.Printf("Executing %s subcommand\n", c.Command.Name)
			return nil
		},
	}
}`, funcName, subcmd, strings.Title(subcmd), subcmd))
		}

		// Generate subcommand list
		subcommandList = "\tsubcommands := []*cli.Command{\n"
		for _, subcmd := range subcommands {
			funcName := toCamelCase(subcmd)
			subcommandList += fmt.Sprintf("\t\tcreate%sCommand(),\n", funcName)
		}
		subcommandList += "\t}\n\n"

		actionCode = `		Action: func(c *cli.Context) error {
			if c.Args().Len() > 0 {
				// Validate subcommand exists
				if !cmd.ValidateSubcommand(c, subcommands) {
					return fmt.Errorf("unknown subcommand: %s", c.Args().First())
				}
				// Valid subcommand, let cli handle it
				return nil
			}
			return prompt.SelectCommand(c, subcommands, "Select a subcommand:", cli.ShowSubcommandHelp)
		},`
	} else {
		imports = `	"fmt"

	"github.com/urfave/cli/v2"`
		subcommandList = ""
		actionCode = `		Action: func(c *cli.Context) error {
			// TODO: Implement your logic here
			fmt.Printf("Executing %s command\n", c.Command.Name)
			return nil
		},`
	}

	var subcommandsField string
	if len(subcommands) > 0 {
		subcommandsField = "\t\tSubcommands: subcommands,\n"
	}

	template := fmt.Sprintf(`package %s

import (
%s
)

func Command() *cli.Command {%s
	return &cli.Command{
		Name:  "%s",
		Usage: "%s",%s
%s
	}
}%s
`, packageName, imports, subcommandList, cmdName, usage, subcommandsField, actionCode, subcommandFuncs.String())

	return template
}

func registerCommandInCLI(workspaceRoot, cmdName string) error {
	cliFile := filepath.Join(workspaceRoot, "cmd", "cli.go")
	content, err := os.ReadFile(cliFile)
	if err != nil {
		return fmt.Errorf("failed to read cmd/cli.go: %w", err)
	}

	contentStr := string(content)

	// Check if already registered
	if strings.Contains(contentStr, fmt.Sprintf("%s.Command()", cmdName)) {
		return fmt.Errorf("command already registered")
	}

	// Find the import section and add the import
	importLine := fmt.Sprintf(`	"cli-aio/cmd/%s"`, cmdName)
	if !strings.Contains(contentStr, importLine) {
		// Find the last import before the closing parenthesis
		importEnd := strings.LastIndex(contentStr, `	"github.com/urfave/cli/v2"`)
		if importEnd == -1 {
			return fmt.Errorf("could not find import section")
		}
		// Insert new import before the closing quote
		newImport := fmt.Sprintf("\t\"cli-aio/cmd/%s\"\n", cmdName)
		contentStr = contentStr[:importEnd] + newImport + contentStr[importEnd:]
	}

	// Find the commands slice and add the command
	commandsStart := strings.Index(contentStr, "commands := []*cli.Command{")
	if commandsStart == -1 {
		return fmt.Errorf("could not find commands slice")
	}

	// Find the end of the commands slice (before the closing brace)
	commandsEnd := strings.Index(contentStr[commandsStart:], "\n\t}")
	if commandsEnd == -1 {
		return fmt.Errorf("could not find end of commands slice")
	}
	commandsEnd += commandsStart

	// Insert the new command before the closing brace
	newCommand := fmt.Sprintf("\t\t%s.Command(),\n", cmdName)
	contentStr = contentStr[:commandsEnd] + newCommand + contentStr[commandsEnd:]

	// Write back
	if err := os.WriteFile(cliFile, []byte(contentStr), 0644); err != nil {
		return fmt.Errorf("failed to write cmd/cli.go: %w", err)
	}

	return nil
}

func findWorkspaceRoot() string {
	// Start from current directory and go up until we find go.mod
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func isValidCommandName(name string) bool {
	if len(name) == 0 {
		return false
	}
	// Allow alphanumeric, hyphens, and underscores
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_') {
			return false
		}
	}
	return true
}

// toPackageName converts a command name to a valid Go package name
func toPackageName(cmdName string) string {
	// Replace hyphens with underscores for package name
	return strings.ReplaceAll(cmdName, "-", "_")
}

func toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}
	parts := strings.Split(s, "-")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]) + part[1:])
		}
	}
	return result.String()
}
