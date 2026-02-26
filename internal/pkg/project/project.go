package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Project represents a saved project entry.
type Project struct {
	Name string `json:"name"` // folder base name
	Path string `json:"path"` // absolute path
}

// ConfigPath returns the path to the projects config file.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "cli-aio", "projects.json"), nil
}

// Load reads all saved projects from disk.
func Load() ([]Project, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []Project{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read projects file: %w", err)
	}

	// Treat an empty file the same as an absent one
	if len(bytes.TrimSpace(data)) == 0 {
		return []Project{}, nil
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse projects file: %w", err)
	}
	return projects, nil
}

// Save writes the project list to disk.
func Save(projects []Project) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(projects, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal projects: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write projects file: %w", err)
	}
	return nil
}

// Add appends a project to the list if it doesn't already exist (by path).
// Returns true if the project was newly added, false if it already existed.
func Add(projects []Project, p Project) ([]Project, bool) {
	for _, existing := range projects {
		if existing.Path == p.Path {
			return projects, false
		}
	}
	return append(projects, p), true
}

// FindGitRepos recursively walks root and returns every directory that
// contains a .git entry. It does not descend further into a found repo
// (avoids counting submodules / nested repos separately).
func FindGitRepos(root string) ([]string, error) {
	var repos []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip directories we can't read (permissions, etc.)
			return filepath.SkipDir
		}
		if !d.IsDir() {
			return nil
		}
		// Skip hidden directories (e.g. .git itself, .cache, ...)
		if path != root && d.Name() != "." && len(d.Name()) > 0 && d.Name()[0] == '.' {
			return filepath.SkipDir
		}

		gitPath := filepath.Join(path, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			repos = append(repos, path)
			// Don't recurse into the repo itself
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to scan %s: %w", root, err)
	}
	return repos, nil
}

// DisplayLabel returns the label shown in the selection list: "name#path".
func (p Project) DisplayLabel() string {
	return fmt.Sprintf("%s#%s", p.Name, p.Path)
}
