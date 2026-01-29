// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourorg/arc-prompt/pkg/prompt"
	"github.com/yourorg/arc-sdk/output"
	"gopkg.in/yaml.v3"
)

func newListCmd() *cobra.Command {
	var outputOpts output.OutputOptions

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all available prompts",
		Long:  "List all YAML prompt templates in the prompts directory.",
		Example: `  arc-prompt list
  arc-prompt list --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := outputOpts.Resolve(); err != nil {
				return err
			}

			prompts, err := prompt.List()
			if err != nil {
				return fmt.Errorf("list prompts: %w", err)
			}

			switch {
			case outputOpts.Is(output.OutputJSON):
				result := struct {
					Prompts []string `json:"prompts"`
					Count   int      `json:"count"`
				}{
					Prompts: prompts,
					Count:   len(prompts),
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)

			case outputOpts.Is(output.OutputYAML):
				result := struct {
					Prompts []string `yaml:"prompts"`
					Count   int      `yaml:"count"`
				}{
					Prompts: prompts,
					Count:   len(prompts),
				}
				data, err := yaml.Marshal(result)
				if err != nil {
					return err
				}
				_, err = cmd.OutOrStdout().Write(data)
				return err

			case outputOpts.Is(output.OutputQuiet):
				for _, p := range prompts {
					fmt.Fprintln(cmd.OutOrStdout(), p)
				}
				return nil

			default:
				if len(prompts) == 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "No prompts found.")
					dir, _ := prompt.Dir()
					fmt.Fprintf(cmd.OutOrStdout(), "\nCreate prompts in: %s\n", dir)
					return nil
				}

				fmt.Fprintf(cmd.OutOrStdout(), "Available prompts (%d):\n", len(prompts))
				for _, p := range prompts {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s\n", p)
				}
				return nil
			}
		},
	}

	outputOpts.AddOutputFlags(cmd, output.OutputTable)
	return cmd
}
