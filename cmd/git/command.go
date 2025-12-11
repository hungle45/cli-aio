package git

import (
	"cli-aio/internal/pkg/git"
	"cli-aio/internal/prompt"
	"fmt"

	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	subcommands := []*cli.Command{
		extractProjectFullName(),
	}

	return &cli.Command{
		Name:        "git",
		Usage:       "Git commands",
		Subcommands: subcommands,
		Action: func(c *cli.Context) error {
			if c.Args().Len() > 0 {
				return nil
			}
			return prompt.SelectCommand(c, subcommands, "Select a subcommand:", cli.ShowSubcommandHelp)
		},
	}
}

func extractProjectFullName() *cli.Command {
	return &cli.Command{
		Name:  "fname",
		Usage: "Extract project full name from git repository",
		Action: func(c *cli.Context) error {
			projectFullName, err := git.ExtractProjectFullName()
			if err != nil {
				return err
			}
			fmt.Printf("Project full name: %s\n", projectFullName)
			return nil
		},
	}
}
