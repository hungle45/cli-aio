package git

import (
	"cli-aio/internal/cmd"
	"cli-aio/internal/pkg/git"
	"cli-aio/internal/prompt"
	"fmt"

	"github.com/urfave/cli/v2"
)

func Command() *cli.Command {
	subcommands := []*cli.Command{
		extractProjectFullName(),
		reversedMergeBranch(),
	}

	return &cli.Command{
		Name:        "git",
		Usage:       "Git commands",
		Subcommands: subcommands,
		Action: func(c *cli.Context) error {
			if c.Args().Len() > 0 {
				// Validate subcommand exists
				if !cmd.ValidateSubcommand(c, subcommands) {
					return fmt.Errorf("unknown subcommand: %s", c.Args().First())
				}
				// Valid subcommand, let cli handle it
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

func reversedMergeBranch() *cli.Command {
	return &cli.Command{
		Name:  "rmerge",
		Usage: "Reverse merge current branch into target branch (checkout to target, then merge current into it)",
		Action: func(c *cli.Context) error {
			// Get current branch (A)
			currentBranch, err := git.GetCurrentBranch()
			if err != nil {
				return fmt.Errorf("failed to get current branch: %w", err)
			}
			fmt.Printf("Current branch: %s\n", currentBranch)

			// Get target branch (B) from args
			if c.Args().Len() == 0 {
				return fmt.Errorf("target branch is required")
			}
			targetBranch := c.Args().First()
			fmt.Printf("Target branch: %s\n", targetBranch)

			// Check if target branch exists
			branchExists, err := git.BranchExists(targetBranch)
			if err != nil {
				return fmt.Errorf("failed to check if branch exists: %w", err)
			}
			if !branchExists {
				return fmt.Errorf("branch '%s' does not exist", targetBranch)
			}

			// Check if we're already on the target branch
			if currentBranch == targetBranch {
				return fmt.Errorf("already on target branch '%s'", targetBranch)
			}

			// Fetch the target branch to make sure we have latest info
			fmt.Printf("Fetching branch '%s'...\n", targetBranch)
			if err := git.FetchBranch(targetBranch); err != nil {
				fmt.Printf("⚠️  Warning: Failed to fetch branch: %v\n", err)
				// Continue anyway, might be a local branch
			}

			// Checkout to target branch
			fmt.Printf("Checking out to branch '%s'...\n", targetBranch)
			if err := git.CheckoutBranch(targetBranch); err != nil {
				return fmt.Errorf("failed to checkout branch: %w", err)
			}

			// Pull latest changes
			fmt.Printf("Pulling latest changes for '%s'...\n", targetBranch)
			if err := git.PullBranch(); err != nil {
				return fmt.Errorf("failed to pull branch: %w", err)
			}

			// Check for merge conflicts before merging
			fmt.Printf("Checking for potential merge conflicts...\n")
			hasConflicts, err := git.CheckMergeConflicts(currentBranch)
			if err != nil {
				return fmt.Errorf("failed to check merge conflicts: %w", err)
			}

			if hasConflicts {
				fmt.Printf("❌ Merge conflicts detected! Cannot merge '%s' into '%s'\n", currentBranch, targetBranch)
				fmt.Printf("Please resolve conflicts manually or use 'git merge %s' after resolving conflicts.\n", currentBranch)
				return fmt.Errorf("merge conflicts detected")
			}

			// Merge current branch into target branch
			fmt.Printf("Merging '%s' into '%s'...\n", currentBranch, targetBranch)
			if err := git.MergeBranch(currentBranch, false); err != nil {
				return fmt.Errorf("failed to merge branch: %w", err)
			}

			// Show success result
			fmt.Printf("✅ Successfully merged '%s' into '%s'\n", currentBranch, targetBranch)
			fmt.Printf("Current branch: %s\n", targetBranch)

			return nil
		},
	}
}
