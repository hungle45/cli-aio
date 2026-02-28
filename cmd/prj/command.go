package prj

import (
	"cli-aio/internal/cmd"
	"cli-aio/internal/pkg/project"
	"cli-aio/internal/prompt"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

// expandPath replaces a leading ~ with the user's home directory.
func expandPath(p string) (string, error) {
	if !strings.HasPrefix(p, "~") {
		return p, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, strings.TrimPrefix(p, "~")), nil
}

func Command() *cli.Command {
	subcommands := []*cli.Command{
		cdCmd(),
		addCmd(),
		gitAddCmd(),
		gitRefreshCmd(),
		editConfigCmd(),
		installCmd(),
	}

	return &cli.Command{
		Name:        "prj",
		Usage:       "Manage projects on your laptop",
		Subcommands: subcommands,
		Action: func(c *cli.Context) error {
			if c.Args().Len() > 0 {
				if !cmd.ValidateSubcommand(c, subcommands) {
					return fmt.Errorf("unknown subcommand: %s", c.Args().First())
				}
				return nil
			}
			return prompt.SelectCommand(c, subcommands, "Select a subcommand:", cli.ShowSubcommandHelp)
		},
	}
}

// cdCmd lists all saved projects and lets the user select one to cd into.
// Because a child process cannot change the parent shell's working directory,
// this command prints the selected path to stdout.
// Wrap it in a shell function to get the actual cd behaviour:
//
//	prj() { local p; p=$(cli-aio prj cd) && cd "$p"; }
func cdCmd() *cli.Command {
	return &cli.Command{
		Name:  "cd",
		Usage: "List projects and print the selected project's path (use with shell wrapper to cd)",
		Action: func(c *cli.Context) error {
			if term.IsTerminal(int(os.Stdout.Fd())) {
				fmt.Fprintln(os.Stderr, "[!] 'aio prj cd' is meant to be called via the 'prj' shell wrapper, not directly.")
				fmt.Fprintln(os.Stderr, "    Run 'aio prj install' to set it up, then reload your shell and use 'prj'.")
				return fmt.Errorf("direct invocation not supported")
			}

			store, err := project.Load()
			if err != nil {
				return err
			}
			if len(store.Projects) == 0 {
				fmt.Fprintln(os.Stderr, "[!] No projects saved. Use 'prj add' or 'prj git-add' to add projects.")
				return nil
			}

			home, _ := os.UserHomeDir()

			// Find max name length for alignment
			maxName := 0
			for _, p := range store.Projects {
				if len(p.Name) > maxName {
					maxName = len(p.Name)
				}
			}

			// Build pretty labels: "name (padded)  ~/short/path"
			labels := make([]string, len(store.Projects))
			pathByLabel := make(map[string]string, len(store.Projects))
			for i, p := range store.Projects {
				shortPath := p.Path
				if home != "" && strings.HasPrefix(p.Path, home) {
					shortPath = "~" + p.Path[len(home):]
				}
				label := fmt.Sprintf("%-*s  %s", maxName, p.Name, shortPath)
				labels[i] = label
				pathByLabel[label] = p.Path
			}

			// SelectOnTTY renders on /dev/tty directly so ANSI escape codes
			// don't leak into the $(...) capture in the shell wrapper.
			_, selected, err := prompt.SelectOnTTY("Select a project:", labels, "")
			if err != nil {
				return fmt.Errorf("selection cancelled: %w", err)
			}

			targetPath, ok := pathByLabel[selected]
			if !ok {
				return fmt.Errorf("selected project not found")
			}
			// Print path to stdout so the shell wrapper can cd to it
			fmt.Print(targetPath)
			return nil
		},
	}
}

// addCmd adds a single folder path to the project list.
func addCmd() *cli.Command {
	return &cli.Command{
		Name:      "add",
		Usage:     "Add a folder as a project",
		ArgsUsage: "[path]",
		Action: func(c *cli.Context) error {
			var folderPath string

			if c.Args().Len() > 0 {
				folderPath = c.Args().First()
			} else {
				// Interactive input
				var err error
				folderPath, err = prompt.Input("Enter folder path:", "", true)
				if err != nil {
					return fmt.Errorf("input cancelled: %w", err)
				}
			}

			// Resolve absolute path (expand ~ first)
			expanded, err := expandPath(folderPath)
			if err != nil {
				return err
			}
			absPath, err := filepath.Abs(expanded)
			if err != nil {
				return fmt.Errorf("invalid path: %w", err)
			}

			// Validate the path exists and is a directory
			info, err := os.Stat(absPath)
			if err != nil {
				return fmt.Errorf("path does not exist: %s", absPath)
			}
			if !info.IsDir() {
				return fmt.Errorf("path is not a directory: %s", absPath)
			}

			store, err := project.Load()
			if err != nil {
				return err
			}

			p := project.Project{
				Name: filepath.Base(absPath),
				Path: absPath,
			}

			added := project.Add(store, p)
			if !added {
				fmt.Printf("[!] Project already exists: %s\n", absPath)
				return nil
			}

			if err := project.Save(store); err != nil {
				return err
			}

			fmt.Printf("[+] Added project: %s (%s)\n", p.Name, p.Path)
			return nil
		},
	}
}

// gitAddCmd scans a folder for git repositories, adds them to the project list,
// and saves the folder path as a git root for future refreshes.
func gitAddCmd() *cli.Command {
	return &cli.Command{
		Name:      "git-add",
		Usage:     "Scan a folder for git repos, add them, and save the folder path for refreshing",
		ArgsUsage: "[path]",
		Aliases:   []string{"add-git"},
		Action: func(c *cli.Context) error {
			var folderPath string

			if c.Args().Len() > 0 {
				folderPath = c.Args().First()
			} else {
				var err error
				folderPath, err = prompt.Input("Enter folder path to scan:", "", true)
				if err != nil {
					return fmt.Errorf("input cancelled: %w", err)
				}
			}

			expanded, err := expandPath(folderPath)
			if err != nil {
				return err
			}
			absPath, err := filepath.Abs(expanded)
			if err != nil {
				return fmt.Errorf("invalid path: %w", err)
			}

			info, err := os.Stat(absPath)
			if err != nil {
				return fmt.Errorf("path does not exist: %s", absPath)
			}
			if !info.IsDir() {
				return fmt.Errorf("path is not a directory: %s", absPath)
			}

			fmt.Printf("Scanning %s for git repositories...\n", absPath)
			repos, err := project.FindGitRepos(absPath)
			if err != nil {
				return err
			}

			store, err := project.Load()
			if err != nil {
				return err
			}

			// Add the root itself to GitRoots
			if addedRoot := project.AddGitRoot(store, absPath); addedRoot {
				fmt.Printf("[+] Saved git root: %s\n", absPath)
			}

			addedProjects := 0
			skippedProjects := 0
			for _, repoPath := range repos {
				p := project.Project{
					Name: filepath.Base(repoPath),
					Path: repoPath,
				}
				if wasAdded := project.Add(store, p); wasAdded {
					addedProjects++
					fmt.Printf("  [+] %s (%s)\n", p.Name, p.Path)
				} else {
					skippedProjects++
					fmt.Printf("  [-] already exists: %s\n", p.Path)
				}
			}

			if err := project.Save(store); err != nil {
				return err
			}

			fmt.Printf("\nDone. Added: %d, Skipped: %d\n", addedProjects, skippedProjects)
			return nil
		},
	}
}

// gitRefreshCmd re-scans all saved git roots for new repositories.
func gitRefreshCmd() *cli.Command {
	return &cli.Command{
		Name:  "git-refresh",
		Usage: "Re-scan all saved git roots for new repositories",
		Action: func(c *cli.Context) error {
			store, err := project.Load()
			if err != nil {
				return err
			}

			if len(store.GitRoots) == 0 {
				fmt.Println("[!] No git roots saved. Use 'prj git-add' to save a git root.")
				return nil
			}

			totalAdded := 0
			totalSkipped := 0

			for _, root := range store.GitRoots {
				fmt.Printf("Refreshing root: %s\n", root)
				repos, err := project.FindGitRepos(root)
				if err != nil {
					fmt.Printf("  [!] Error scanning %s: %v\n", root, err)
					continue
				}

				for _, repoPath := range repos {
					p := project.Project{
						Name: filepath.Base(repoPath),
						Path: repoPath,
					}
					if wasAdded := project.Add(store, p); wasAdded {
						totalAdded++
						fmt.Printf("  [+] %s (%s)\n", p.Name, p.Path)
					} else {
						totalSkipped++
					}
				}
			}

			if totalAdded > 0 {
				if err := project.Save(store); err != nil {
					return err
				}
			}

			fmt.Printf("\nDone. Total added: %d, Total already exist: %d\n", totalAdded, totalSkipped)
			return nil
		},
	}
}

// editConfigCmd opens the projects config file in the user's preferred editor.
func editConfigCmd() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Open the projects config file in $EDITOR (fallback: nvim)",
		Action: func(c *cli.Context) error {
			configPath, err := project.ConfigPath()
			if err != nil {
				return err
			}

			// Ensure the file exists so the editor doesn't open a blank buffer
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				if err := project.Save(&project.Store{Projects: []project.Project{}, GitRoots: []string{}}); err != nil {
					return fmt.Errorf("failed to initialise config file: %w", err)
				}
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				// Try common editors in order of preference
				for _, candidate := range []string{"nvim", "vim", "nano", "vi", "notepad"} {
					if _, err := exec.LookPath(candidate); err == nil {
						editor = candidate
						break
					}
				}
			}
			if editor == "" {
				return fmt.Errorf("no editor found; set the $EDITOR environment variable")
			}

			cmdExec := exec.Command(editor, configPath)
			cmdExec.Stdin = os.Stdin
			cmdExec.Stdout = os.Stdout
			cmdExec.Stderr = os.Stderr
			if err := cmdExec.Run(); err != nil {
				return fmt.Errorf("editor exited with error: %w", err)
			}
			return nil
		},
	}
}
