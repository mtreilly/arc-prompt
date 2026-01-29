// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/yourorg/arc-prompt/store"
	"github.com/yourorg/arc-sdk/output"
	"gopkg.in/yaml.v3"
)

func newStatsCmd(promptStore *store.PromptsStore) *cobra.Command {
	var outputOpts output.OutputOptions

	cmd := &cobra.Command{
		Use:   "stats [name]",
		Short: "Show prompt usage statistics",
		Long:  "Display usage statistics for a specific prompt or all prompts.",
		Example: `  arc-prompt stats                     # Show most used prompts
  arc-prompt stats code-review         # Stats for specific prompt
  arc-prompt stats code-review --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := outputOpts.Resolve(); err != nil {
				return err
			}

			ctx := context.Background()
			out := cmd.OutOrStdout()

			if len(args) > 0 {
				return showPromptStats(ctx, promptStore, args[0], outputOpts, out)
			}
			return showMostUsed(ctx, promptStore, outputOpts, out)
		},
	}

	outputOpts.AddOutputFlags(cmd, output.OutputTable)
	return cmd
}

func showPromptStats(ctx context.Context, ps *store.PromptsStore, name string, out output.OutputOptions, w io.Writer) error {
	stats, err := ps.Stats(ctx, name)
	if err != nil {
		return fmt.Errorf("get stats: %w", err)
	}

	switch {
	case out.Is(output.OutputJSON):
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(stats)

	case out.Is(output.OutputYAML):
		enc := yaml.NewEncoder(w)
		defer enc.Close()
		return enc.Encode(stats)

	case out.Is(output.OutputQuiet):
		fmt.Fprintln(w, stats.Total)
		return nil

	default:
		fmt.Fprintf(w, "Prompt: %s\n", stats.PromptName)
		fmt.Fprintf(w, "Total executions: %d\n", stats.Total)
		if stats.Total > 0 {
			fmt.Fprintf(w, "Success rate: %s\n", calculateSuccessRate(stats.Successful, stats.Total))
			fmt.Fprintf(w, "Avg tokens: %.0f\n", stats.AvgTokens)
			fmt.Fprintf(w, "Total tokens: %d\n", stats.TotalTokens)
		} else {
			fmt.Fprintln(w, "Success rate: n/a (no executions yet)")
		}
		return nil
	}
}

func showMostUsed(ctx context.Context, ps *store.PromptsStore, out output.OutputOptions, w io.Writer) error {
	prompts, err := ps.MostUsed(ctx, 10)
	if err != nil {
		return fmt.Errorf("get most used: %w", err)
	}

	type promptUsage struct {
		Name             string
		Count            int
		SuccessRate      float64
		SuccessRateLabel string
		TotalTokens      int64
	}

	statsList := make([]promptUsage, 0, len(prompts))
	for _, p := range prompts {
		stats, err := ps.Stats(ctx, p.Name)
		if err != nil {
			return fmt.Errorf("get stats for %s: %w", p.Name, err)
		}
		statsList = append(statsList, promptUsage{
			Name:             p.Name,
			Count:            stats.Total,
			SuccessRate:      successRateValue(stats.Successful, stats.Total),
			SuccessRateLabel: calculateSuccessRate(stats.Successful, stats.Total),
			TotalTokens:      stats.TotalTokens,
		})
	}

	if len(statsList) == 0 {
		switch {
		case out.Is(output.OutputJSON):
			enc := json.NewEncoder(w)
			enc.SetIndent("", "  ")
			return enc.Encode(struct {
				Prompts []any `json:"prompts"`
				Count   int   `json:"count"`
			}{Prompts: nil, Count: 0})

		case out.Is(output.OutputYAML):
			enc := yaml.NewEncoder(w)
			defer enc.Close()
			return enc.Encode(struct {
				Prompts []any `yaml:"prompts"`
				Count   int   `yaml:"count"`
			}{Prompts: nil, Count: 0})

		case out.Is(output.OutputQuiet):
			return nil

		default:
			fmt.Fprintln(w, "No prompt usage recorded yet.")
			return nil
		}
	}

	switch {
	case out.Is(output.OutputJSON):
		payload := struct {
			Prompts []struct {
				Name        string  `json:"name"`
				Count       int     `json:"count"`
				SuccessRate float64 `json:"success_rate"`
				TotalTokens int64   `json:"total_tokens"`
			} `json:"prompts"`
		}{}
		for _, s := range statsList {
			payload.Prompts = append(payload.Prompts, struct {
				Name        string  `json:"name"`
				Count       int     `json:"count"`
				SuccessRate float64 `json:"success_rate"`
				TotalTokens int64   `json:"total_tokens"`
			}{
				Name:        s.Name,
				Count:       s.Count,
				SuccessRate: s.SuccessRate,
				TotalTokens: s.TotalTokens,
			})
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(payload)

	case out.Is(output.OutputYAML):
		payload := struct {
			Prompts []struct {
				Name        string  `yaml:"name"`
				Count       int     `yaml:"count"`
				SuccessRate float64 `yaml:"success_rate"`
				TotalTokens int64   `yaml:"total_tokens"`
			} `yaml:"prompts"`
		}{}
		for _, s := range statsList {
			payload.Prompts = append(payload.Prompts, struct {
				Name        string  `yaml:"name"`
				Count       int     `yaml:"count"`
				SuccessRate float64 `yaml:"success_rate"`
				TotalTokens int64   `yaml:"total_tokens"`
			}{
				Name:        s.Name,
				Count:       s.Count,
				SuccessRate: s.SuccessRate,
				TotalTokens: s.TotalTokens,
			})
		}
		enc := yaml.NewEncoder(w)
		defer enc.Close()
		return enc.Encode(payload)

	case out.Is(output.OutputQuiet):
		for _, s := range statsList {
			fmt.Fprintf(w, "%s\t%d\n", s.Name, s.Count)
		}
		return nil

	default:
		fmt.Fprintln(w, "Most used prompts (top 10):")
		fmt.Fprintf(w, "%-3s %-30s %8s %12s %12s\n", "#", "Prompt", "Uses", "Success", "Tokens")
		for i, s := range statsList {
			fmt.Fprintf(w, "%2d. %-30s %8d %12s %12d\n",
				i+1, s.Name, s.Count, s.SuccessRateLabel, s.TotalTokens)
		}
		return nil
	}
}

func calculateSuccessRate(successful, total int) string {
	if total <= 0 {
		return "n/a"
	}
	rate := successRateValue(successful, total)
	return fmt.Sprintf("%.1f%% (%d/%d)", rate, successful, total)
}

func successRateValue(successful, total int) float64 {
	if total <= 0 {
		return 0
	}
	return (float64(successful) / float64(total)) * 100
}
