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

// Store holds the overall project state.
type Store struct {
	Projects []Project `json:"projects"`
	GitRoots []string  `json:"git_roots"`
}

// ConfigPath returns the path to the projects config file.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", "cli-aio", "projects.json"), nil
}

// Load reads the store from disk.
func Load() (*Store, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Store{
			Projects: []Project{},
			GitRoots: []string{},
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read projects file: %w", err)
	}

	// Treat an empty file the same as an absent one
	if len(bytes.TrimSpace(data)) == 0 {
		return &Store{
			Projects: []Project{},
			GitRoots: []string{},
		}, nil
	}

	// Try parsing as the new Store format
	var store Store
	if err := json.Unmarshal(data, &store); err == nil && (len(store.Projects) > 0 || len(store.GitRoots) > 0) {
		// New format successfully parsed (and not just an empty object)
		if store.Projects == nil {
			store.Projects = []Project{}
		}
		if store.GitRoots == nil {
			store.GitRoots = []string{}
		}
		return &store, nil
	}

	// Fallback: parse as the old []Project format
	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("failed to parse projects file: %w", err)
	}

	// Return a new Store containing the old projects
	return &Store{
		Projects: projects,
		GitRoots: []string{},
	}, nil
}

// Save writes the store to disk.
func Save(store *Store) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal store: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write projects file: %w", err)
	}
	return nil
}

// Add appends a project to the project list if it doesn't already exist (by path).
// Returns true if the project was newly added, false if it already existed.
func Add(store *Store, p Project) bool {
	for _, existing := range store.Projects {
		if existing.Path == p.Path {
			return false
		}
	}
	store.Projects = append(store.Projects, p)
	return true
}

// AddGitRoot appends a git root to the list if it doesn't already exist.
// Returns true if the root was newly added, false if it already existed.
func AddGitRoot(store *Store, gitRoot string) bool {
	for _, existing := range store.GitRoots {
		if existing == gitRoot {
			return false
		}
	}
	store.GitRoots = append(store.GitRoots, gitRoot)
	return true
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
