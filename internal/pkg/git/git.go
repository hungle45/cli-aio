package git

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// CheckIfGitRepo checks if the current directory is a git repository.
func CheckIfGitRepo() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("error running git command to check if git repository: %w", err)
	}
	return strings.TrimSpace(string(output)) == "true", nil
}

// GetCurrentBranch gets the current branch name using the git command.
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running git command to get current branch: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// ExtractProjectFullName extracts the project full name from the remote origin URL
// eg: https://gitlab.zalopay.vn/bank/operation/bank-config-fe-v2.git -> bank/operation/bank-config-fe-v2
func ExtractProjectFullName() (string, error) {
	url, err := GetRemoteOriginURL()
	if err != nil {
		return "", err
	}
	pattern := `(?:.*:?\/\/|.*@.*?[:/])(.*)\.git$`
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(url)

	if len(matches) > 1 {
		projectFullName := matches[1]
		return projectFullName, nil
	}

	return "", fmt.Errorf("could not extract project full name from URL: %s", url)
}

// ExtractProjectID extracts the project ID from the remote origin URL.
// eg: https://gitlab.zalopay.vn/bank/operation/bank-config-fe-v2.git -> bank/operation/bank-config-fe-v2.git
func ExtractProjectID() (string, error) {
	fullName, err := ExtractProjectFullName()
	if err != nil {
		return "", err
	}
	parts := strings.Split(fullName, "/")
	projectID := strings.Join(parts[1:], "/")
	return projectID, nil

}

// GetRemoteOriginURL gets the remote origin URL using the git command.
func GetRemoteOriginURL() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error running git command to get remote origin URL: %w", err)
	}

	url := strings.TrimSpace(string(output))
	if url == "" {
		return "", fmt.Errorf("git remote 'origin' URL not found")
	}
	return url, nil
}

// GetLatestTags gets the latest tags from the git repository.
func GetLatestTags(limit int) ([]string, error) {
	// git tag --sort=version:refname | tail -n {limit}
	cmd := exec.Command("git", "tag", "--sort=version:refname")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running git command to get latest tags: %w", err)
	}
	tags := strings.Split(strings.TrimSpace(string(output)), "\n")

	if len(tags) == 0 || len(tags) == 1 && tags[0] == "" {
		return []string{"v0.0.0"}, nil
	}
	if len(tags) > limit {
		return tags[len(tags)-limit:], nil
	}
	return tags, nil
}

func CreateAndPushTag(tag string, message string) error {
	if err := exec.Command("git", "tag", tag, "-m", message).Run(); err != nil {
		return fmt.Errorf("error running git command to create tag: %w", err)
	}
	if err := exec.Command("git", "push", "origin", tag).Run(); err != nil {
		return fmt.Errorf("error running git command to push tag: %w", err)
	}
	return nil
}

func CreateZalopayRelease(projectID string, tag string, message string) error {
	gitlabToken := os.Getenv("GITLAB_PRIVATE_TOKEN")
	if gitlabToken == "" {
		return fmt.Errorf("GITLAB_PRIVATE_TOKEN is not set")
	}
	_, err := exec.Command("curl", "--header", "Content-Type: application/json", "--header",
		fmt.Sprintf("PRIVATE-TOKEN: %s", gitlabToken),
		"--data", fmt.Sprintf("{ \"name\": \"%s\", \"tag_name\": \"%s\", \"description\": \"%s\" }", tag, tag, message),
		"--request", "POST", fmt.Sprintf("https://gitlab.zalopay.vn/api/v4/projects/%s/releases", projectID)).Output()
	if err != nil {
		return fmt.Errorf("error running git command to create release: %w", err)
	}
	return nil
}
