// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yourorg/arc-prompt/pkg/prompt"
	"github.com/yourorg/arc-sdk/output"
	"gopkg.in/yaml.v3"
)

func newShowCmd() *cobra.Command {
	var outputOpts output.OutputOptions

	cmd := &cobra.Command{
		Use:   "show <name>",
		Short: "Display prompt details",
		Long:  "Display the contents and metadata of a prompt template.",
		Example: `  arc-prompt show code-review
  arc-prompt show analyze-repo --output json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			p, err := prompt.LoadWithDefaults(name)
			if err != nil {
				return fmt.Errorf("load prompt %q: %w", name, err)
			}

			if err := outputOpts.Resolve(); err != nil {
				return err
			}

			switch {
			case outputOpts.Is(output.OutputJSON):
				result := struct {
					Name     string            `json:"name"`
					Model    string            `json:"model,omitempty"`
					System   string            `json:"system,omitempty"`
					User     string            `json:"user,omitempty"`
					Defaults map[string]string `json:"defaults,omitempty"`
					Options  map[string]any    `json:"options,omitempty"`
					Metadata map[string]any    `json:"metadata,omitempty"`
				}{
					Name:     p.Name,
					Model:    p.Model,
					System:   p.System,
					User:     p.User,
					Defaults: p.Defaults,
					Options:  p.Options,
					Metadata: p.Metadata,
				}
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)

			case outputOpts.Is(output.OutputYAML):
				result := struct {
					Name     string            `yaml:"name"`
					Model    string            `yaml:"model,omitempty"`
					System   string            `yaml:"system,omitempty"`
					User     string            `yaml:"user,omitempty"`
					Defaults map[string]string `yaml:"defaults,omitempty"`
					Options  map[string]any    `yaml:"options,omitempty"`
					Metadata map[string]any    `yaml:"metadata,omitempty"`
				}{
					Name:     p.Name,
					Model:    p.Model,
					System:   p.System,
					User:     p.User,
					Defaults: p.Defaults,
					Options:  p.Options,
					Metadata: p.Metadata,
				}
				data, err := yaml.Marshal(result)
				if err != nil {
					return err
				}
				_, err = cmd.OutOrStdout().Write(data)
				return err

			case outputOpts.Is(output.OutputQuiet):
				fmt.Fprintln(cmd.OutOrStdout(), p.Name)
				return nil

			default:
				out := cmd.OutOrStdout()
				fmt.Fprintf(out, "Prompt: %s\n", p.Name)
				if p.Model != "" {
					fmt.Fprintf(out, "Model: %s\n", p.Model)
				}

				if len(p.Defaults) > 0 {
					fmt.Fprintln(out, "\nDefaults:")
					for k, v := range p.Defaults {
						fmt.Fprintf(out, "  %s: %s\n", k, v)
					}
				}

				if len(p.Options) > 0 {
					fmt.Fprintln(out, "\nOptions:")
					for k, v := range p.Options {
						fmt.Fprintf(out, "  %s: %v\n", k, v)
					}
				}

				if len(p.Metadata) > 0 {
					fmt.Fprintln(out, "\nMetadata:")
					for k, v := range p.Metadata {
						fmt.Fprintf(out, "  %s: %v\n", k, v)
					}
				}

				if p.System != "" {
					fmt.Fprintln(out, "\nSystem Template:")
					fmt.Fprintln(out, strings.Repeat("-", 60))
					fmt.Fprintln(out, p.System)
				}

				if p.User != "" {
					fmt.Fprintln(out, "\nUser Template:")
					fmt.Fprintln(out, strings.Repeat("-", 60))
					fmt.Fprintln(out, p.User)
				}

				return nil
			}
		},
	}

	outputOpts.AddOutputFlags(cmd, output.OutputTable)
	return cmd
}
