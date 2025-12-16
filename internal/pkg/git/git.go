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

// GetLatestTags gets the latest tags from the remote git repository using creatordate order.
func GetLatestTags(limit int) ([]string, error) {
	// git ls-remote --tags --refs --sort=-creatordate | head -n {limit}
	cmd := exec.Command("git", "ls-remote", "--tags", "--refs", "--sort=-creatordate")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running git command to get latest tags: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var tags []string
	for _, line := range lines {
		parts := strings.Split(line, "\t")
		if len(parts) == 2 {
			ref := parts[1]
			const prefix = "refs/tags/"
			if strings.HasPrefix(ref, prefix) {
				tag := strings.TrimPrefix(ref, prefix)
				tags = append(tags, tag)
			}
		}
	}

	if len(tags) == 0 {
		return []string{"v0.0.0"}, nil
	}

	if len(tags) > limit {
		return tags[:limit], nil
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

// CheckoutBranch checks out to the specified branch.
func CheckoutBranch(branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error checking out branch %s: %w\nOutput: %s", branch, err, string(output))
	}
	return nil
}

// PullBranch pulls the latest changes from remote for the current branch.
func PullBranch() error {
	cmd := exec.Command("git", "pull")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error pulling branch: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// CheckMergeConflicts checks if merging sourceBranch into current branch would cause conflicts.
// Returns true if there would be conflicts, false otherwise.
// Uses a test merge approach: attempts merge with --no-commit and --no-ff, then aborts.
func CheckMergeConflicts(sourceBranch string) (bool, error) {
	// Ensure we clean up any merge state on exit
	defer func() {
		// Try to abort any ongoing merge
		abortCmd := exec.Command("git", "merge", "--abort")
		_ = abortCmd.Run() // Ignore errors, just try to clean up
	}()

	// First, check if branches are already merged
	cmd := exec.Command("git", "merge-base", "--is-ancestor", sourceBranch, "HEAD")
	err := cmd.Run()
	if err == nil {
		// sourceBranch is already an ancestor of HEAD, so it's already merged
		return false, nil
	}

	// Try to do a test merge with --no-commit to check for conflicts
	// This will not actually commit the merge, allowing us to check for conflicts
	cmd = exec.Command("git", "merge", "--no-commit", "--no-ff", sourceBranch)
	output, err := cmd.CombinedOutput()

	// Check if merge was successful (no conflicts)
	if err == nil {
		// Merge succeeded, abort it since we're just testing
		abortCmd := exec.Command("git", "merge", "--abort")
		_ = abortCmd.Run() // Ignore abort errors
		return false, nil
	}

	// Merge failed, check if it's due to conflicts
	outputStr := string(output)
	hasConflicts := strings.Contains(outputStr, "CONFLICT") ||
		strings.Contains(outputStr, "conflict") ||
		strings.Contains(outputStr, "Automatic merge failed")

	if hasConflicts {
		// Abort the merge attempt
		abortCmd := exec.Command("git", "merge", "--abort")
		_ = abortCmd.Run() // Ignore abort errors
		return true, nil
	}

	// Some other error occurred - abort and return error
	abortCmd := exec.Command("git", "merge", "--abort")
	_ = abortCmd.Run() // Try to clean up anyway
	return false, fmt.Errorf("error checking merge conflicts: %w\nOutput: %s", err, outputStr)
}

// MergeBranch merges sourceBranch into the current branch.
func MergeBranch(sourceBranch string, noFF bool) error {
	args := []string{"merge", sourceBranch}
	if noFF {
		args = append(args, "--no-ff")
	}
	cmd := exec.Command("git", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error merging branch %s: %w\nOutput: %s", sourceBranch, err, string(output))
	}
	return nil
}

// FetchBranch fetches the specified branch from remote.
func FetchBranch(branch string) error {
	cmd := exec.Command("git", "fetch", "origin", branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error fetching branch %s: %w\nOutput: %s", branch, err, string(output))
	}
	return nil
}

// BranchExists checks if a branch exists (local or remote).
func BranchExists(branch string) (bool, error) {
	// Check local branches
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", "refs/heads/"+branch)
	err := cmd.Run()
	if err == nil {
		return true, nil
	}

	// Check remote branches
	cmd = exec.Command("git", "show-ref", "--verify", "--quiet", "refs/remotes/origin/"+branch)
	err = cmd.Run()
	if err == nil {
		return true, nil
	}

	return false, nil
}

// GetLocalBranches gets a list of all local branch names.
func GetLocalBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--format", "%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error getting local branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	for _, line := range lines {
		branch := strings.TrimSpace(line)
		if branch != "" {
			branches = append(branches, branch)
		}
	}

	return branches, nil
}

// GetRemoteBranches gets a list of all remote branch names (without remote prefix).
func GetRemoteBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-r", "--format", "%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error getting remote branches: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var branches []string
	seen := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Remove remote prefix (e.g., "origin/branch-name" -> "branch-name")
		parts := strings.Split(line, "/")
		if len(parts) > 1 {
			branch := strings.Join(parts[1:], "/")
			// Skip HEAD reference
			if branch != "HEAD" && !seen[branch] {
				branches = append(branches, branch)
				seen[branch] = true
			}
		}
	}

	return branches, nil
}

// GetAllAvailableBranches gets a combined list of local and remote branches.
// Remote branches are only included if they don't exist locally.
func GetAllAvailableBranches() ([]string, error) {
	localBranches, err := GetLocalBranches()
	if err != nil {
		return nil, err
	}

	remoteBranches, err := GetRemoteBranches()
	if err != nil {
		// If we can't get remote branches, just return local ones
		return localBranches, nil
	}

	// Create a map of local branches for quick lookup
	localMap := make(map[string]bool)
	for _, branch := range localBranches {
		localMap[branch] = true
	}

	// Combine local branches with remote branches that don't exist locally
	allBranches := make([]string, len(localBranches))
	copy(allBranches, localBranches)

	for _, remoteBranch := range remoteBranches {
		if !localMap[remoteBranch] {
			allBranches = append(allBranches, remoteBranch)
		}
	}

	return allBranches, nil
}
