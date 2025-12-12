package ztag

import (
	"cli-aio/internal/pkg/git"
	"cli-aio/internal/prompt"
	"fmt"

	"github.com/urfave/cli/v2"
)

type Env string

const (
	EnvQC   Env = "qc"
	EnvStg  Env = "stg"
	EnvProd Env = "prod"
)

type Level string

const (
	LevelBug   Level = "b"
	LevelMinor Level = "m"
	LevelMajor Level = "M"
)

// map between project path and env to indicate which env the project will be deployed to when no env is provided
var defaultEnvMap = map[string][]Env{
	"bank/operation/bank-config-fe-v2": {EnvQC, EnvStg},
}

type VersionInfo struct {
	Major int
	Minor int
	Patch int
}

func Command() *cli.Command {
	subcommands := []*cli.Command{
		createGenerateTagCommand(EnvQC),
		createGenerateTagCommand(EnvStg),
		createGenerateTagCommand(EnvProd),
	}

	return &cli.Command{
		Name:  "ztag",
		Usage: "Generate a new tag for a specific environment",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "level",
				Aliases: []string{"l"},
				Usage:   "Level of the tag: b (default) for bug, m for minor and M for major",
				Value:   "b",
			},
		},
		Subcommands: subcommands,
		Action: func(c *cli.Context) error {
			if isGitRepo, err := git.CheckIfGitRepo(); err != nil || !isGitRepo {
				return fmt.Errorf("not a git repository")
			}

			if c.Args().Len() > 0 {
				return nil
			}

			projectID, err := git.ExtractProjectID()
			if err != nil {
				return err
			}
			fmt.Printf("Project ID: %s\n", projectID)

			envs, ok := defaultEnvMap[projectID]
			if ok {
				for _, env := range envs {
					err = createGenerateTagCommand(env).Action(c)
					if err != nil {
						return err
					}
				}
				return nil
			}

			return prompt.SelectCommand(c, subcommands, "Select a Environment:", cli.ShowSubcommandHelp)
		},
	}
}

func createGenerateTagCommand(env Env) *cli.Command {
	return &cli.Command{
		Name:  string(env),
		Usage: fmt.Sprintf("Generate a new tag for %s environment", string(env)),
		Action: func(c *cli.Context) error {
			currentBranch, err := git.GetCurrentBranch()
			if err != nil {
				return err
			}
			if env == EnvProd && currentBranch != "main" && currentBranch != "master" {
				return fmt.Errorf("only main/master branches are allowed to be deployed to %s environment", string(env))
			}

			latestTags, err := git.GetLatestTags(1)
			if err != nil {
				return err
			}

			nextTag, err := GenerateNextTag(latestTags[0], Level(c.String("level")), env)
			if err != nil {
				return err
			}

			fmt.Printf("Latest tag: %s, Next tag: %s\n", latestTags[0], nextTag)
			err = git.CreateAndPushTag(nextTag, fmt.Sprintf("Release %s", nextTag))
			if err != nil {
				return err
			}

			// require user input jira ticket
			if env == EnvQC {
				return nil
			}

			jiraTicket, err := prompt.Input("Enter Jira ticket (required):", "", true)
			if err != nil {
				return err
			}

			projectID, err := git.ExtractProjectID()
			if err != nil {
				return err
			}

			fmt.Printf("Release project with tag %s and Jira ticket %s\n", nextTag, jiraTicket)
			err = git.CreateZalopayRelease(projectID, nextTag, jiraTicket)
			if err != nil {
				return err
			}
			fmt.Printf("Released %s successfully\n", nextTag)

			return nil
		},
	}
}
