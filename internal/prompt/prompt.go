package prompt

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

// IsInteractive checks if the command should run in interactive mode.
// Interactive mode is enabled when:
//   - The interactive flag is explicitly set to true, OR
//   - The interactive flag is not set and we're in a TTY (terminal)
func IsInteractive(interactiveFlag bool) bool {
	if interactiveFlag {
		return true
	}
	// Check if we're in a TTY (terminal)
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Select prompts the user to select from a list of options.
// Returns the selected option index and value.
// If defaultOption is empty, the first option will be used as default.
func Select(message string, options []string, defaultOption string) (int, string, error) {
	return SelectWithFuzzy(message, options, defaultOption, true)
}

// SelectWithFuzzy prompts the user to select from a list of options with optional fuzzy search.
// If fuzzy is true, enables fuzzy search filtering.
func SelectWithFuzzy(message string, options []string, defaultOption string, fuzzy bool) (int, string, error) {
	if len(options) == 0 {
		return -1, "", fmt.Errorf("no options to select from")
	}

	var selected string
	prompt := &survey.Select{
		Message: message,
		Options: options,
	}

	// Only set default if it's not empty and exists in options
	if defaultOption != "" {
		for _, opt := range options {
			if opt == defaultOption {
				prompt.Default = defaultOption
				break
			}
		}
	}

	var err error
	if fuzzy {
		// Enable fuzzy search with a custom filter
		err = survey.AskOne(prompt, &selected, survey.WithFilter(fuzzyFilter))
	} else {
		err = survey.AskOne(prompt, &selected)
	}

	if err != nil {
		return -1, "", err
	}

	// Find the index of the selected option
	for i, opt := range options {
		if opt == selected {
			return i, selected, nil
		}
	}
	return -1, selected, nil
}

// fuzzyFilter implements fuzzy matching for survey prompts.
// It matches if all characters in the filter appear in order in the option.
func fuzzyFilter(filter string, option string, index int) bool {
	if filter == "" {
		return true
	}

	filter = strings.ToLower(filter)
	option = strings.ToLower(option)

	// Simple fuzzy matching: all characters in filter must appear in order in option
	filterIdx := 0
	for i := 0; i < len(option) && filterIdx < len(filter); i++ {
		if option[i] == filter[filterIdx] {
			filterIdx++
		}
	}

	return filterIdx == len(filter)
}

// Input prompts the user for text input.
func Input(message string, defaultVal string, required bool) (string, error) {
	var result string
	prompt := &survey.Input{
		Message: message,
		Default: defaultVal,
	}
	var err error
	if required {
		err = survey.AskOne(prompt, &result, survey.WithValidator(survey.Required))
	} else {
		err = survey.AskOne(prompt, &result)
	}
	return result, err
}

// Confirm prompts the user for a yes/no confirmation.
func Confirm(message string, defaultVal bool) (bool, error) {
	var result bool
	prompt := &survey.Confirm{
		Message: message,
		Default: defaultVal,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// MultiSelect prompts the user to select multiple options from a list.
func MultiSelect(message string, options []string, defaults []string) ([]string, error) {
	var result []string
	prompt := &survey.MultiSelect{
		Message: message,
		Options: options,
		Default: defaults,
	}
	err := survey.AskOne(prompt, &result)
	return result, err
}

// ShouldUseInteractive checks if interactive mode should be used.
// Returns true if:
//   - We're in a TTY (terminal), AND
//   - Any required parameters are missing
//
// This enables interactive mode automatically when needed.
func ShouldUseInteractive(interactiveFlag bool, hasMissingParams bool) bool {
	// If explicitly disabled, don't use interactive
	if !interactiveFlag && !term.IsTerminal(int(os.Stdin.Fd())) {
		return false
	}
	// If explicitly enabled, use interactive
	if interactiveFlag {
		return true
	}
	// Auto-enable if in TTY and params are missing
	return term.IsTerminal(int(os.Stdin.Fd())) && hasMissingParams
}

// SelectCommand is a helper function that prompts the user to select a command/subcommand
// from a list of cli.Command. It automatically extracts command names and handles execution.
// This makes it easy to add interactive selection without manually creating name arrays and maps.
//
// Usage:
//
//	subcommands := []*cli.Command{createCmd(), listCmd(), deleteCmd()}
//	return prompt.SelectCommand(c, subcommands, "Select a subcommand:", cli.ShowSubcommandHelp)
func SelectCommand(c *cli.Context, commands []*cli.Command, message string, onCancel func(*cli.Context) error) error {
	if len(commands) == 0 {
		return fmt.Errorf("no commands available to select")
	}

	// Auto-extract command names from the commands slice
	commandNames := make([]string, len(commands))
	commandMap := make(map[string]*cli.Command, len(commands))
	for i, cmd := range commands {
		commandNames[i] = cmd.Name
		commandMap[cmd.Name] = cmd
	}

	// Check if we're in a TTY - if not, show help
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		if onCancel != nil {
			return onCancel(c)
		}
		return nil
	}

	// We're in a TTY - prompt user to select
	_, selected, err := Select(message, commandNames, "")
	if err != nil {
		// If user cancels (Ctrl+C) or stdin is closed, show help instead of error
		if err.Error() == "interrupt" || err.Error() == "EOF" {
			if onCancel != nil {
				return onCancel(c)
			}
			return nil
		}
		// For other errors, show help with a message
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		if onCancel != nil {
			return onCancel(c)
		}
		return err
	}

	// Execute the selected command
	selectedCmd := commandMap[selected]
	if selectedCmd == nil {
		return fmt.Errorf("selected command not found: %s", selected)
	}

	if selectedCmd.Action != nil {
		return selectedCmd.Action(c)
	}

	// If no Action, show help for the command
	if onCancel != nil {
		return onCancel(c)
	}
	return nil
}
