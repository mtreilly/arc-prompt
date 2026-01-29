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

func newPathCmd() *cobra.Command {
	var outputOpts output.OutputOptions

	cmd := &cobra.Command{
		Use:   "path",
		Short: "Show prompts directory path",
		Long:  "Display the path to the prompts directory (~/.config/arc/prompts).",
		Example: `  arc-prompt path
  arc-prompt path --output json
  cd $(arc-prompt path)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := outputOpts.Resolve(); err != nil {
				return err
			}

			dir, err := prompt.Dir()
			if err != nil {
				return fmt.Errorf("get prompts directory: %w", err)
			}

			switch {
			case outputOpts.Is(output.OutputJSON):
				result := struct {
					Path string `json:"path"`
				}{Path: dir}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)

			case outputOpts.Is(output.OutputYAML):
				result := struct {
					Path string `yaml:"path"`
				}{Path: dir}
				data, err := yaml.Marshal(result)
				if err != nil {
					return err
				}
				_, err = cmd.OutOrStdout().Write(data)
				return err

			default:
				fmt.Fprintln(cmd.OutOrStdout(), dir)
				return nil
			}
		},
	}

	outputOpts.AddOutputFlags(cmd, output.OutputTable)
	return cmd
}
