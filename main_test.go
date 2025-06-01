package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `genre:
  - web
  - osint
  - crypto`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Test loading config
	config, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	expected := []string{"web", "osint", "crypto"}
	if !reflect.DeepEqual(config.Genre, expected) {
		t.Errorf("Expected genres %v, got %v", expected, config.Genre)
	}
}

func TestLoadChallenge(t *testing.T) {
	// Create temporary challenge file
	tmpDir := t.TempDir()
	challengePath := filepath.Join(tmpDir, "challenge.yml")

	challengeContent := `name: "Test Challenge"
tags:
  - web
  - sql-injection
  - beginner`

	err := os.WriteFile(challengePath, []byte(challengeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test challenge: %v", err)
	}

	// Test loading challenge
	challenge, err := loadChallenge(challengePath)
	if err != nil {
		t.Fatalf("Failed to load challenge: %v", err)
	}

	if challenge.Name != "Test Challenge" {
		t.Errorf("Expected name 'Test Challenge', got '%s'", challenge.Name)
	}

	expectedTags := []string{"web", "sql-injection", "beginner"}
	if !reflect.DeepEqual(challenge.Tags, expectedTags) {
		t.Errorf("Expected tags %v, got %v", expectedTags, challenge.Tags)
	}
}

func TestHasMatchingTag(t *testing.T) {
	tests := []struct {
		name          string
		challengeTags []string
		searchTags    []string
		expected      bool
	}{
		{
			name:          "Exact match",
			challengeTags: []string{"web", "crypto", "beginner"},
			searchTags:    []string{"web"},
			expected:      true,
		},
		{
			name:          "Partial match",
			challengeTags: []string{"web-security", "crypto", "beginner"},
			searchTags:    []string{"web"},
			expected:      true,
		},
		{
			name:          "Case insensitive match",
			challengeTags: []string{"WEB", "CRYPTO", "BEGINNER"},
			searchTags:    []string{"web"},
			expected:      true,
		},
		{
			name:          "Multiple search tags, one matches",
			challengeTags: []string{"web", "crypto", "beginner"},
			searchTags:    []string{"forensics", "web"},
			expected:      true,
		},
		{
			name:          "No match",
			challengeTags: []string{"web", "crypto", "beginner"},
			searchTags:    []string{"forensics", "reverse"},
			expected:      false,
		},
		{
			name:          "Empty challenge tags",
			challengeTags: []string{},
			searchTags:    []string{"web"},
			expected:      false,
		},
		{
			name:          "Empty search tags",
			challengeTags: []string{"web"},
			searchTags:    []string{},
			expected:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasMatchingTag(tt.challengeTags, tt.searchTags)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFindMatchingChallenges(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create web genre directory with challenges
	webDir := filepath.Join(tmpDir, "web")
	chall1Dir := filepath.Join(webDir, "chall_1")
	chall2Dir := filepath.Join(webDir, "chall_2")

	err := os.MkdirAll(chall1Dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	err = os.MkdirAll(chall2Dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create osint genre directory with challenge
	osintDir := filepath.Join(tmpDir, "osint")
	chall3Dir := filepath.Join(osintDir, "chall_3")

	err = os.MkdirAll(chall3Dir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create challenge files
	challenge1Content := `name: "Web Challenge 1"
tags:
  - web
  - sql-injection`

	challenge2Content := `name: "Web Challenge 2"
tags:
  - web
  - xss`

	challenge3Content := `name: "OSINT Challenge 1"
tags:
  - osint
  - social-media`

	err = os.WriteFile(filepath.Join(chall1Dir, "challenge.yml"), []byte(challenge1Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test challenge: %v", err)
	}

	err = os.WriteFile(filepath.Join(chall2Dir, "challenge.yml"), []byte(challenge2Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test challenge: %v", err)
	}

	err = os.WriteFile(filepath.Join(chall3Dir, "challenge.yml"), []byte(challenge3Content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test challenge: %v", err)
	}

	// Change to temporary directory to test
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tmpDir)

	// Test finding challenges
	genres := []string{"web", "osint"}
	searchTags := []string{"web"}

	results, err := findMatchingChallenges(genres, searchTags)
	if err != nil {
		t.Fatalf("Failed to find challenges: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Check if correct challenges were found
	foundNames := make(map[string]bool)
	for _, result := range results {
		foundNames[result.Name] = true
	}

	if !foundNames["Web Challenge 1"] {
		t.Error("Expected to find 'Web Challenge 1'")
	}
	if !foundNames["Web Challenge 2"] {
		t.Error("Expected to find 'Web Challenge 2'")
	}
	if foundNames["OSINT Challenge 1"] {
		t.Error("Should not find 'OSINT Challenge 1' when searching for 'web'")
	}
}
