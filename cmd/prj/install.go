package prj

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

const (
	markerBegin = "# >>> prj wrapper (added by aio prj install) >>>"
	markerEnd   = "# <<< prj wrapper <<<"
)

// shellConfig describes how to install the wrapper for a particular shell.
type shellConfig struct {
	// configFile is the rc file to append to (absolute path).
	configFile string
	// snippet is the text to inject (between the markers).
	snippet string
	// reload is the human-readable command to reload the shell.
	reload string
}

// posixSnippet returns the POSIX-compatible wrapper for bash/zsh/ksh.
func posixSnippet() string {
	return `function prj() {
  local target
  target=$(aio prj cd 2>/dev/tty) && [ -n "$target" ] && cd "$target"
}`
}

// fishSnippet returns the Fish shell wrapper.
func fishSnippet() string {
	return `function prj
  set target (aio prj cd 2>/dev/tty)
  and test -n "$target"
  and cd $target
end`
}

// detectShellConfig reads $SHELL and returns the appropriate shellConfig.
func detectShellConfig() (*shellConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}

	shell := os.Getenv("SHELL")
	base := filepath.Base(shell)

	switch base {
	case "zsh":
		return &shellConfig{
			configFile: filepath.Join(home, ".zshrc"),
			snippet:    posixSnippet(),
			reload:     "exec zsh",
		}, nil

	case "bash":
		// Prefer .bashrc; fall back to .bash_profile (macOS default login shell)
		rc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(rc); os.IsNotExist(err) {
			rc = filepath.Join(home, ".bash_profile")
		}
		return &shellConfig{
			configFile: rc,
			snippet:    posixSnippet(),
			reload:     "source " + rc,
		}, nil

	case "fish":
		funcDir := filepath.Join(home, ".config", "fish", "functions")
		return &shellConfig{
			// Fish loads every file in functions/ automatically
			configFile: filepath.Join(funcDir, "prj.fish"),
			snippet:    fishSnippet(),
			reload:     "source ~/.config/fish/functions/prj.fish",
		}, nil

	case "ksh", "ksh93", "mksh":
		return &shellConfig{
			configFile: filepath.Join(home, ".kshrc"),
			snippet:    posixSnippet(),
			reload:     "source ~/.kshrc",
		}, nil

	default:
		// Unknown shell — fall back to ~/.profile (POSIX lowest-common-denominator)
		return &shellConfig{
			configFile: filepath.Join(home, ".profile"),
			snippet:    posixSnippet(),
			reload:     "source ~/.profile",
		}, nil
	}
}

// isAlreadyInstalled checks whether the markers are present in the config file.
func isAlreadyInstalled(configFile string) (bool, error) {
	data, err := os.ReadFile(configFile)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return strings.Contains(string(data), markerBegin), nil
}

// writeWrapper appends the marked wrapper block to the config file.
func writeWrapper(cfg *shellConfig) error {
	// Ensure parent directory exists (e.g. fish functions/)
	if err := os.MkdirAll(filepath.Dir(cfg.configFile), 0755); err != nil {
		return fmt.Errorf("cannot create directory: %w", err)
	}

	f, err := os.OpenFile(cfg.configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("cannot open %s: %w", cfg.configFile, err)
	}
	defer f.Close()

	block := fmt.Sprintf("\n%s\n%s\n%s\n", markerBegin, cfg.snippet, markerEnd)
	if _, err := f.WriteString(block); err != nil {
		return fmt.Errorf("cannot write to %s: %w", cfg.configFile, err)
	}
	return nil
}

func installCmd() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "Install the prj shell wrapper so 'prj' cd's your terminal into the selected project",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "shell",
				Aliases: []string{"s"},
				Usage:   "Override shell detection (zsh, bash, fish, ksh)",
			},
		},
		Action: func(c *cli.Context) error {
			cfg, err := detectShellConfig()
			if err != nil {
				return err
			}

			// Allow manual shell override
			if override := c.String("shell"); override != "" {
				home, _ := os.UserHomeDir()
				switch override {
				case "zsh":
					cfg = &shellConfig{filepath.Join(home, ".zshrc"), posixSnippet(), "exec zsh"}
				case "bash":
					rc := filepath.Join(home, ".bashrc")
					if _, err := os.Stat(rc); os.IsNotExist(err) {
						rc = filepath.Join(home, ".bash_profile")
					}
					cfg = &shellConfig{rc, posixSnippet(), "source " + rc}
				case "fish":
					cfg = &shellConfig{
						filepath.Join(home, ".config", "fish", "functions", "prj.fish"),
						fishSnippet(),
						"source ~/.config/fish/functions/prj.fish",
					}
				case "ksh":
					cfg = &shellConfig{filepath.Join(home, ".kshrc"), posixSnippet(), "source ~/.kshrc"}
				default:
					return fmt.Errorf("unsupported shell: %s (supported: zsh, bash, fish, ksh)", override)
				}
			}

			// Check if already installed
			installed, err := isAlreadyInstalled(cfg.configFile)
			if err != nil {
				return fmt.Errorf("cannot check %s: %w", cfg.configFile, err)
			}
			if installed {
				fmt.Printf("[!] prj wrapper is already installed in %s\n", cfg.configFile)
				fmt.Printf("    To reinstall, remove the block between:\n")
				fmt.Printf("      %s\n", markerBegin)
				fmt.Printf("      %s\n", markerEnd)
				return nil
			}

			if err := writeWrapper(cfg); err != nil {
				return err
			}

			fmt.Printf("[+] Installed prj wrapper into %s\n\n", cfg.configFile)
			fmt.Printf("    Reload your shell to activate:\n")
			fmt.Printf("      %s\n\n", cfg.reload)
			fmt.Printf("    Then just type 'prj' to navigate to any project.\n")
			return nil
		},
	}
}
