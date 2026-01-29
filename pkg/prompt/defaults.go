// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package prompt

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed defaults/commit-message.yaml
var defaultCommitMessage string

//go:embed defaults/code-review.yaml
var defaultCodeReview string

//go:embed defaults/annotate-commit.yaml
var defaultAnnotateCommit string

// defaultPrompts maps prompt names to their embedded content.
var defaultPrompts = map[string]string{
	"commit-message":  defaultCommitMessage,
	"code-review":     defaultCodeReview,
	"annotate-commit": defaultAnnotateCommit,
}

// EnsureDefaults checks if the prompts directory exists and has prompts.
// If not, it creates the directory and installs default prompts.
// Returns true if defaults were installed, false if already present.
func EnsureDefaults() (bool, error) {
	dir, err := Dir()
	if err != nil {
		return false, fmt.Errorf("failed to get prompts directory: %w", err)
	}

	// Create directory if needed
	if err := os.MkdirAll(dir, 0755); err != nil {
		return false, fmt.Errorf("failed to create prompts directory: %w", err)
	}

	// Check which prompts need to be installed
	needsInstall := false
	for name := range defaultPrompts {
		path := filepath.Join(dir, name+".yaml")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			needsInstall = true
			break
		}
	}

	if !needsInstall {
		return false, nil // All prompts already exist
	}

	fmt.Printf("Installing default prompts to %s...\n", dir)

	// Install each default prompt
	installed := 0
	for name, content := range defaultPrompts {
		if content == "" {
			continue // Skip if embed failed
		}

		path := filepath.Join(dir, name+".yaml")

		// Check if already exists
		if _, err := os.Stat(path); err == nil {
			continue // Already exists, skip
		}

		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return false, fmt.Errorf("failed to write %s: %w", name, err)
		}

		fmt.Printf("  Installed %s.yaml\n", name)
		installed++
	}

	if installed == 0 {
		return false, fmt.Errorf("no default prompts were installed (embed may have failed)")
	}

	fmt.Printf("\nInstalled %d default prompt(s).\n", installed)
	return true, nil
}

// LoadWithDefaults is like Load but ensures defaults are installed first.
func LoadWithDefaults(name string) (*Prompt, error) {
	// First try to load normally
	p, err := Load(name)
	if err == nil {
		return p, nil // Success
	}

	// If loading failed, try to install defaults and load again
	if _, installErr := EnsureDefaults(); installErr != nil {
		return nil, err
	}

	// Try loading again after installing defaults
	return Load(name)
}
