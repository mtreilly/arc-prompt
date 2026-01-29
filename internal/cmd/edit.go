// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourorg/arc-prompt/pkg/prompt"
)

func newEditCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit <name>",
		Short: "Open prompt in $EDITOR",
		Long:  "Open a prompt template in your default editor ($EDITOR).",
		Example: `  arc-prompt edit code-review
  arc-prompt edit analyze-repo`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}

			dir, err := prompt.Dir()
			if err != nil {
				return fmt.Errorf("get prompts directory: %w", err)
			}

			path := fmt.Sprintf("%s/%s", dir, name)
			if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
				path += ".yaml"
			}

			if _, err := os.Stat(path); os.IsNotExist(err) {
				fmt.Printf("Prompt %q does not exist.\n", name)
				fmt.Printf("Create new prompt at %s? [y/N]: ", path)

				var response string
				fmt.Scanln(&response)
				response = strings.ToLower(strings.TrimSpace(response))

				if response != "y" && response != "yes" {
					return fmt.Errorf("prompt not found: %s", name)
				}

				template := `name: ` + name + `

system: |
  You are a helpful assistant.

user: |
  {{.Input}}

defaults:
  Input: ""

model: claude-sonnet-4-5

metadata:
  description: "Description of this prompt"
  version: "1.0"
`
				if err := os.WriteFile(path, []byte(template), 0644); err != nil {
					return fmt.Errorf("create prompt file: %w", err)
				}
				fmt.Printf("Created new prompt: %s\n", path)
			}

			editorCmd := exec.Command(editor, path)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr

			if err := editorCmd.Run(); err != nil {
				return fmt.Errorf("run editor: %w", err)
			}

			return nil
		},
	}

	return cmd
}
