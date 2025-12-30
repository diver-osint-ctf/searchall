package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// listLocalBranches returns a list of all local branch names
func listLocalBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "--format=%(refname:short)")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	branches := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(branches) == 1 && branches[0] == "" {
		return []string{}, nil
	}

	return branches, nil
}

// getCurrentBranch returns the name of the current branch
func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// getFileContentFromBranch retrieves the content of a file from a specific branch
func getFileContentFromBranch(branch, path string) ([]byte, error) {
	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", branch, path))
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get file content from branch %s: %w", branch, err)
	}

	return output, nil
}

// listChallengeFilesInBranch lists all challenge.yml files in a specific branch under a genre directory
func listChallengeFilesInBranch(branch, genre string) ([]string, error) {
	// List all files in the genre directory tree
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", fmt.Sprintf("%s:%s", branch, genre))
	output, err := cmd.Output()
	if err != nil {
		// Genre directory might not exist in this branch
		return []string{}, nil
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var challengeFiles []string

	for _, line := range lines {
		if strings.HasSuffix(line, "challenge.yml") {
			// Construct full path: genre/...path.../challenge.yml
			fullPath := fmt.Sprintf("%s/%s", genre, line)
			challengeFiles = append(challengeFiles, fullPath)
		}
	}

	return challengeFiles, nil
}
