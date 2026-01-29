// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package cmd

import (
	"database/sql"

	"github.com/spf13/cobra"
	"github.com/yourorg/arc-prompt/store"
)

// NewRootCmd creates the root command for arc-prompt.
func NewRootCmd(db *sql.DB) *cobra.Command {
	promptStore := store.NewPromptsStore(db)

	root := &cobra.Command{
		Use:   "arc-prompt",
		Short: "Manage prompt templates",
		Long: `Manage YAML prompt templates for AI-powered CLI tools.

Prompts are stored in ~/.config/arc/prompts/ and use text/template syntax
for variable interpolation.`,
		Example: `  arc-prompt list                    # List all available prompts
  arc-prompt show code-review        # Display prompt details
  arc-prompt path                    # Show prompts directory
  arc-prompt edit analyze-repo       # Open prompt in $EDITOR`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newListCmd(),
		newShowCmd(),
		newPathCmd(),
		newEditCmd(),
		newStatsCmd(promptStore),
		newHistoryCmd(promptStore),
	)

	return root
}
