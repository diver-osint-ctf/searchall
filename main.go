package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// Config represents the config.yaml structure
type Config struct {
	Genre []string `yaml:"genre"`
}

// Challenge represents the challenge.yml structure
type Challenge struct {
	Name string   `yaml:"name"`
	Tags []string `yaml:"tags"`
}

// ChallengeResult holds challenge information with its file path
type ChallengeResult struct {
	Name       string
	Tags       []string
	FilePath   string
	BranchName string // Branch name where the challenge was found
}

func main() {
	// Parse flags
	allBranches := flag.Bool("all-branches", false, "Search challenges across all local branches")
	flag.Parse()

	// Load config.yaml
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config.yaml: %v", err)
	}

	// Select appropriate loader
	var loader ChallengeLoader
	if *allBranches {
		currentBranch, err := getCurrentBranch()
		if err != nil {
			log.Fatalf("Failed to get current branch: %v", err)
		}
		loader = &MultiBranchLoader{CurrentBranch: currentBranch}
	} else {
		// Use file system loader for backward compatibility
		loader = &FileSystemLoader{BranchName: ""}
	}

	// Load all challenges once
	allChallenges, err := loader.LoadChallenges(config.Genre)
	if err != nil {
		log.Fatalf("Failed to load challenges: %v", err)
	}

	// Get non-flag arguments (tags)
	searchTags := flag.Args()

	if len(searchTags) == 0 {
		// Interactive dynamic search mode
		if err := interactiveSearch(allChallenges); err != nil {
			log.Fatalf("Interactive search failed: %v", err)
		}
	} else {
		// Static search mode with provided tags
		results := filterChallengesByTags(allChallenges, searchTags)

		if len(results) == 0 {
			fmt.Printf("No challenges found with tags: %s\n", strings.Join(searchTags, ", "))
			return
		}

		// Display results in markdown format
		displayMarkdownResults(results)
	}
}

// interactiveSearch provides real-time interactive search
func interactiveSearch(allChallenges []ChallengeResult) error {
	// Save the original terminal state
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	// Clear screen and hide cursor
	clearScreen()
	defer fmt.Print("\033[?25h") // Show cursor on exit

	var input []rune
	var cursorPos int

	// Display initial state
	displaySearchUIWithCursor(string(input), cursorPos, allChallenges)

	// Buffer for reading characters
	buf := make([]byte, 3) // Support for escape sequences

	for {
		// Read input
		n, err := os.Stdin.Read(buf[:1])
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}
		if n == 0 {
			continue
		}

		char := buf[0]

		switch char {
		case 3: // Ctrl+C
			clearScreen()
			fmt.Print("\033[?25h") // Show cursor
			fmt.Println("Goodbye!")
			return nil
		case 27: // ESC - start of escape sequence
			// Read the next two bytes for arrow keys
			n, err := os.Stdin.Read(buf[1:3])
			if err != nil || n < 2 {
				continue
			}

			if buf[1] == '[' {
				switch buf[2] {
				case 'D': // Left arrow
					if cursorPos > 0 {
						cursorPos--
						clearScreen()
						displaySearchUIWithCursor(string(input), cursorPos, filterChallengesByInput(allChallenges, string(input)))
					}
				case 'C': // Right arrow
					if cursorPos < len(input) {
						cursorPos++
						clearScreen()
						displaySearchUIWithCursor(string(input), cursorPos, filterChallengesByInput(allChallenges, string(input)))
					}
				}
			}
		case 127, 8: // Backspace or Delete
			if cursorPos > 0 && len(input) > 0 {
				// Remove character before cursor
				input = append(input[:cursorPos-1], input[cursorPos:]...)
				cursorPos--

				// Update display
				clearScreen()
				results := filterChallengesByInput(allChallenges, string(input))
				displaySearchUIWithCursor(string(input), cursorPos, results)
			}
		case 13: // Enter
			// Select first result if available
			results := filterChallengesByInput(allChallenges, string(input))
			if len(results) > 0 {
				clearScreen()
				fmt.Print("\033[?25h") // Show cursor
				selected := results[0]
				fmt.Printf("Selected: %s\n", selected.Name)
				fmt.Printf("Tags: %s\n", strings.Join(selected.Tags, ", "))
				fmt.Printf("Path: %s\n", selected.FilePath)
				return nil
			}
		default:
			// Add printable characters at cursor position
			if char >= 32 && char <= 126 {
				// Insert character at cursor position
				input = append(input[:cursorPos], append([]rune{rune(char)}, input[cursorPos:]...)...)
				cursorPos++

				// Update display in real-time
				clearScreen()
				results := filterChallengesByInput(allChallenges, string(input))
				displaySearchUIWithCursor(string(input), cursorPos, results)
			}
		}
	}
}

// clearScreen clears the entire screen and moves cursor to top-left
func clearScreen() {
	fmt.Print("\033[2J")   // Clear entire screen
	fmt.Print("\033[H")    // Move cursor to home position (1,1)
	fmt.Print("\033[?25l") // Hide cursor
}

// displaySearchUIWithCursor displays the search interface with cursor position
func displaySearchUIWithCursor(input string, cursorPos int, challenges []ChallengeResult) {
	// Display input line with cursor visualization
	if cursorPos >= len(input) {
		fmt.Printf("input: %s█\r\n", input)
	} else {
		before := input[:cursorPos]
		at := input[cursorPos : cursorPos+1]
		after := input[cursorPos+1:]
		fmt.Printf("input: %s\033[7m%s\033[0m%s\r\n", before, at, after)
	}
	fmt.Print("\r\n") // Empty line

	// Display results
	if len(challenges) == 0 {
		fmt.Print("No challenges found\r\n")
	} else {
		for _, challenge := range challenges {
			if challenge.BranchName != "" {
				fmt.Printf("- [%s] %s (tags: %s)\r\n",
					challenge.BranchName,
					challenge.Name,
					strings.Join(challenge.Tags, ", "))
			} else {
				fmt.Printf("- %s (tags: %s)\r\n",
					challenge.Name,
					strings.Join(challenge.Tags, ", "))
			}
		}
	}
}

// filterChallengesByInput filters challenges by input string contained in tags
func filterChallengesByInput(allChallenges []ChallengeResult, input string) []ChallengeResult {
	if input == "" {
		return allChallenges
	}

	var results []ChallengeResult
	searchTerm := strings.ToLower(strings.TrimSpace(input))

	for _, challenge := range allChallenges {
		// Check if the input string is contained in any of the challenge's tags
		for _, tag := range challenge.Tags {
			if strings.Contains(strings.ToLower(tag), searchTerm) {
				results = append(results, challenge)
				break // Found a match, no need to check other tags for this challenge
			}
		}
	}

	return results
}

// loadAllChallenges loads all challenges from all genres
func loadAllChallenges(genres []string) ([]ChallengeResult, error) {
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
				BranchName: "",
			})

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %w", genre, err)
		}
	}

	return allChallenges, nil
}

// filterChallengesByTags filters challenges by the given search tags
func filterChallengesByTags(allChallenges []ChallengeResult, searchTags []string) []ChallengeResult {
	var results []ChallengeResult

	for _, challenge := range allChallenges {
		if hasMatchingTag(challenge.Tags, searchTags) {
			results = append(results, challenge)
		}
	}

	return results
}

// findMatchingChallenges searches for challenges with matching tags (kept for backward compatibility)
func findMatchingChallenges(genres []string, searchTags []string) ([]ChallengeResult, error) {
	allChallenges, err := loadAllChallenges(genres)
	if err != nil {
		return nil, err
	}

	return filterChallengesByTags(allChallenges, searchTags), nil
}

// loadConfig loads and parses config.yaml
func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// loadChallenge loads and parses a challenge.yml file
func loadChallenge(challengePath string) (*Challenge, error) {
	data, err := os.ReadFile(challengePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read challenge file: %w", err)
	}

	var challenge Challenge
	if err := yaml.Unmarshal(data, &challenge); err != nil {
		return nil, fmt.Errorf("failed to parse challenge file: %w", err)
	}

	return &challenge, nil
}

// hasMatchingTag checks if any of the search tags match the challenge tags
func hasMatchingTag(challengeTags []string, searchTags []string) bool {
	for _, searchTag := range searchTags {
		for _, challengeTag := range challengeTags {
			if strings.Contains(strings.ToLower(challengeTag), strings.ToLower(searchTag)) {
				return true
			}
		}
	}
	return false
}

// displayMarkdownResults displays the results in markdown list format
func displayMarkdownResults(results []ChallengeResult) {
	for _, result := range results {
		if result.BranchName != "" {
			fmt.Printf("- [%s] \"%s\"\n", result.BranchName, result.Name)
		} else {
			fmt.Printf("- \"%s\"\n", result.Name)
		}
	}
}
