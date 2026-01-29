// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourorg/arc-prompt/store"
	"github.com/yourorg/arc-sdk/output"
	"gopkg.in/yaml.v3"
)

func newHistoryCmd(promptStore *store.PromptsStore) *cobra.Command {
	var limit int
	var outputOpts output.OutputOptions

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show recent prompt executions",
		Long:  "Display recent prompt execution history with timestamps and results.",
		Example: `  arc-prompt history
  arc-prompt history --limit 50
  arc-prompt history --output json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := outputOpts.Resolve(); err != nil {
				return err
			}

			ctx := context.Background()
			out := cmd.OutOrStdout()

			history, err := promptStore.RecentUsage(ctx, limit)
			if err != nil {
				return fmt.Errorf("get history: %w", err)
			}

			switch {
			case outputOpts.Is(output.OutputJSON):
				result := struct {
					History []store.PromptUsage `json:"history"`
					Count   int                 `json:"count"`
				}{
					History: history,
					Count:   len(history),
				}
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(result)

			case outputOpts.Is(output.OutputYAML):
				result := struct {
					History []store.PromptUsage `yaml:"history"`
					Count   int                 `yaml:"count"`
				}{
					History: history,
					Count:   len(history),
				}
				enc := yaml.NewEncoder(out)
				defer enc.Close()
				return enc.Encode(result)

			case outputOpts.Is(output.OutputQuiet):
				for _, h := range history {
					ts := time.Unix(h.Timestamp, 0).Format(time.RFC3339)
					fmt.Fprintf(out, "%s\t%s\t%t\n", ts, h.PromptName, h.Success)
				}
				return nil

			default:
				if len(history) == 0 {
					fmt.Fprintln(out, "No prompt execution history found.")
					return nil
				}

				fmt.Fprintf(out, "Recent prompt executions (%d):\n\n", len(history))
				for _, h := range history {
					ts := time.Unix(h.Timestamp, 0)
					status := "OK"
					if !h.Success {
						status = "FAIL"
					}

					fmt.Fprintf(out, "[%s] %s\n", ts.Format("2006-01-02 15:04:05"), h.PromptName)
					fmt.Fprintf(out, "  Model: %s | Tokens: %d | Status: %s\n", h.Model, h.TokensUsed, status)
					if h.SessionID != "" {
						fmt.Fprintf(out, "  Session: %s\n", h.SessionID)
					}
					if paramsLine := formatPromptParams(h.Params); paramsLine != "" {
						fmt.Fprintf(out, "  Params: %s\n", paramsLine)
					}
					fmt.Fprintln(out)
				}
				return nil
			}
		},
	}

	outputOpts.AddOutputFlags(cmd, output.OutputTable)
	cmd.Flags().IntVar(&limit, "limit", 20, "number of recent executions to show")

	return cmd
}

func formatPromptParams(raw string) string {
	if strings.TrimSpace(raw) == "" {
		return ""
	}
	params := make(map[string]string)
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		return ""
	}
	if len(params) == 0 {
		return ""
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	return strings.Join(parts, ", ")
}
