package main

import (
	"testing"
)

func TestGetBranchPriority(t *testing.T) {
	tests := []struct {
		name          string
		branchName    string
		currentBranch string
		expected      BranchPriority
	}{
		{
			name:          "main branch has highest priority",
			branchName:    "main",
			currentBranch: "feature/test",
			expected:      PriorityMain,
		},
		{
			name:          "current branch has second priority",
			branchName:    "feature/test",
			currentBranch: "feature/test",
			expected:      PriorityCurrent,
		},
		{
			name:          "other branch has lowest priority",
			branchName:    "feature/other",
			currentBranch: "feature/test",
			expected:      PriorityOther,
		},
		{
			name:          "main is always priority even if it's current",
			branchName:    "main",
			currentBranch: "main",
			expected:      PriorityMain,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBranchPriority(tt.branchName, tt.currentBranch)
			if result != tt.expected {
				t.Errorf("Expected priority %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestSortBranchesByPriority(t *testing.T) {
	loader := &MultiBranchLoader{CurrentBranch: "feat/test"}

	branches := []string{"feat/other", "main", "feat/another", "feat/test"}

	sorted := loader.sortBranchesByPriority(branches)

	// Expected order: main, feat/test (current), feat/other, feat/another
	if len(sorted) != 4 {
		t.Errorf("Expected 4 branches, got %d", len(sorted))
	}

	// First should be main
	if sorted[0] != "main" {
		t.Errorf("Expected first branch to be 'main', got '%s'", sorted[0])
	}

	// Second should be current branch
	if sorted[1] != "feat/test" {
		t.Errorf("Expected second branch to be 'feat/test', got '%s'", sorted[1])
	}

	// Remaining branches should be feat/other and feat/another (order doesn't matter)
	remaining := map[string]bool{
		sorted[2]: true,
		sorted[3]: true,
	}
	if !remaining["feat/other"] || !remaining["feat/another"] {
		t.Errorf("Expected remaining branches to be 'feat/other' and 'feat/another', got '%s' and '%s'", sorted[2], sorted[3])
	}
}

func TestSortBranchesByPriorityWithMainAsCurrent(t *testing.T) {
	loader := &MultiBranchLoader{CurrentBranch: "main"}

	branches := []string{"feat/test", "main", "feat/other"}

	sorted := loader.sortBranchesByPriority(branches)

	// main should still be first (PriorityMain takes precedence over PriorityCurrent)
	if sorted[0] != "main" {
		t.Errorf("Expected first branch to be 'main', got '%s'", sorted[0])
	}
}
