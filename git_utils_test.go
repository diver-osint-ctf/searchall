package main

import (
	"testing"
)

func TestListLocalBranches(t *testing.T) {
	branches, err := listLocalBranches()
	if err != nil {
		t.Fatalf("Failed to list branches: %v", err)
	}

	if len(branches) == 0 {
		t.Error("Expected at least one branch, got none")
	}

	// Should have at least the current branch
	currentBranch, err := getCurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	found := false
	for _, branch := range branches {
		if branch == currentBranch {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Current branch '%s' not found in branch list: %v", currentBranch, branches)
	}
}

func TestGetCurrentBranch(t *testing.T) {
	branch, err := getCurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	if branch == "" {
		t.Error("Current branch name is empty")
	}
}

func TestGetFileContentFromBranch(t *testing.T) {
	// Get current branch
	currentBranch, err := getCurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	// Try to read README.md from current branch
	content, err := getFileContentFromBranch(currentBranch, "README.md")
	if err != nil {
		t.Fatalf("Failed to get file content: %v", err)
	}

	if len(content) == 0 {
		t.Error("File content is empty")
	}
}

func TestListChallengeFilesInBranch(t *testing.T) {
	// Get current branch
	currentBranch, err := getCurrentBranch()
	if err != nil {
		t.Fatalf("Failed to get current branch: %v", err)
	}

	// List challenge files in osint genre
	files, err := listChallengeFilesInBranch(currentBranch, "osint")
	if err != nil {
		t.Fatalf("Failed to list challenge files: %v", err)
	}

	// We know osint directory exists with challenges
	if len(files) == 0 {
		t.Error("Expected to find challenge files in osint genre")
	}

	// All files should end with challenge.yml
	for _, file := range files {
		if len(file) < 13 || file[len(file)-13:] != "challenge.yml" {
			t.Errorf("File '%s' does not end with challenge.yml", file)
		}
	}
}
