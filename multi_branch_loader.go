package main

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v3"
)

// BranchPriority represents the priority of a branch for deduplication
type BranchPriority int

const (
	PriorityMain    BranchPriority = 1
	PriorityCurrent BranchPriority = 2
	PriorityOther   BranchPriority = 3
)

// MultiBranchLoader loads challenges from all local branches and deduplicates them
type MultiBranchLoader struct {
	CurrentBranch string
}

// LoadChallenges loads challenges from all local branches and returns deduplicated results
// Optimized to parse only files from the highest priority branch
func (m *MultiBranchLoader) LoadChallenges(genres []string) ([]ChallengeResult, error) {
	branches, err := listLocalBranches()
	if err != nil {
		return nil, err
	}

	// Sort branches by priority (main -> current -> others)
	sortedBranches := m.sortBranchesByPriority(branches)

	// Track which file paths we've already processed
	processedFiles := make(map[string]bool)
	var results []ChallengeResult

	// Process branches in priority order
	for _, branch := range sortedBranches {
		for _, genre := range genres {
			files, err := listChallengeFilesInBranch(branch, genre)
			if err != nil {
				// Genre might not exist in this branch, skip
				continue
			}

			for _, filePath := range files {
				// Skip if we've already processed this file from a higher-priority branch
				if processedFiles[filePath] {
					continue
				}

				// Mark as processed
				processedFiles[filePath] = true

				// Parse the challenge file
				content, err := getFileContentFromBranch(branch, filePath)
				if err != nil {
					// File might not exist or be readable, skip
					continue
				}

				var challenge Challenge
				if err := yaml.Unmarshal(content, &challenge); err != nil {
					fmt.Printf("Warning: Failed to parse %s in branch %s: %v\n", filePath, branch, err)
					continue
				}

				results = append(results, ChallengeResult{
					Name:       challenge.Name,
					Tags:       challenge.Tags,
					FilePath:   filePath,
					BranchName: branch,
				})
			}
		}
	}

	return results, nil
}

// sortBranchesByPriority sorts branches by priority: main -> current -> others
func (m *MultiBranchLoader) sortBranchesByPriority(branches []string) []string {
	sorted := make([]string, len(branches))
	copy(sorted, branches)

	sort.Slice(sorted, func(i, j int) bool {
		priorityI := getBranchPriority(sorted[i], m.CurrentBranch)
		priorityJ := getBranchPriority(sorted[j], m.CurrentBranch)
		return priorityI < priorityJ
	})

	return sorted
}

// getBranchPriority returns the priority of a branch
func getBranchPriority(branchName, currentBranch string) BranchPriority {
	if branchName == "main" {
		return PriorityMain
	}
	if branchName == currentBranch {
		return PriorityCurrent
	}
	return PriorityOther
}
