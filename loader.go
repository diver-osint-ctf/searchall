package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ChallengeLoader is an interface for loading challenges from various sources
type ChallengeLoader interface {
	LoadChallenges(genres []string) ([]ChallengeResult, error)
}

// FileSystemLoader loads challenges from the current working directory (file system)
type FileSystemLoader struct {
	BranchName string // Optional: branch name to display (empty string means no branch display)
}

// LoadChallenges loads all challenges from the file system (existing logic)
func (f *FileSystemLoader) LoadChallenges(genres []string) ([]ChallengeResult, error) {
	var allChallenges []ChallengeResult

	for _, genre := range genres {
		if _, err := os.Stat(genre); os.IsNotExist(err) {
			continue // Skip non-existent genre directories
		}

		err := filepath.WalkDir(genre, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || d.Name() != "challenge.yml" {
				return nil
			}

			challenge, err := loadChallenge(path)
			if err != nil {
				fmt.Printf("Warning: Failed to load %s: %v\n", path, err)
				return nil
			}

			allChallenges = append(allChallenges, ChallengeResult{
				Name:       challenge.Name,
				Tags:       challenge.Tags,
				FilePath:   path,
				BranchName: f.BranchName,
			})

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", genre, err)
		}
	}

	return allChallenges, nil
}

// GitBranchLoader loads challenges from a specific Git branch
type GitBranchLoader struct {
	BranchName string
}

// LoadChallenges loads all challenges from the specified Git branch
func (g *GitBranchLoader) LoadChallenges(genres []string) ([]ChallengeResult, error) {
	var challenges []ChallengeResult

	for _, genre := range genres {
		files, err := listChallengeFilesInBranch(g.BranchName, genre)
		if err != nil {
			// Genre might not exist in this branch, skip
			continue
		}

		for _, file := range files {
			content, err := getFileContentFromBranch(g.BranchName, file)
			if err != nil {
				// File might not exist or be readable, skip
				continue
			}

			var challenge Challenge
			if err := yaml.Unmarshal(content, &challenge); err != nil {
				fmt.Printf("Warning: Failed to parse %s in branch %s: %v\n", file, g.BranchName, err)
				continue
			}

			challenges = append(challenges, ChallengeResult{
				Name:       challenge.Name,
				Tags:       challenge.Tags,
				FilePath:   file,
				BranchName: g.BranchName,
			})
		}
	}

	return challenges, nil
}
