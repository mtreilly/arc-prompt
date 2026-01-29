// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package store

import (
	"context"
	"database/sql"
	"strings"
	"time"
)

// PromptUsage represents a single prompt execution record.
type PromptUsage struct {
	ID         int64  `json:"id"`
	PromptName string `json:"prompt_name"`
	Model      string `json:"model"`
	Params     string `json:"params"` // JSON-encoded parameters
	Timestamp  int64  `json:"timestamp"`
	TokensUsed int    `json:"tokens_used"`
	Success    bool   `json:"success"`
	SessionID  string `json:"session_id,omitempty"`
}

// PromptStats holds aggregate statistics for a prompt.
type PromptStats struct {
	PromptName  string  `json:"prompt_name"`
	Total       int     `json:"total"`
	Successful  int     `json:"successful"`
	Failed      int     `json:"failed"`
	AvgTokens   float64 `json:"avg_tokens"`
	TotalTokens int64   `json:"total_tokens"`
}

// PromptsStore provides access to prompt usage tracking and metadata.
type PromptsStore struct {
	DB *sql.DB
}

// NewPromptsStore creates a new PromptsStore.
func NewPromptsStore(db *sql.DB) *PromptsStore {
	return &PromptsStore{DB: db}
}

// RecordUsage tracks a prompt execution.
func (s *PromptsStore) RecordUsage(ctx context.Context, usage PromptUsage) (int64, error) {
	sessionID := any(nil)
	if strings.TrimSpace(usage.SessionID) != "" {
		sessionID = usage.SessionID
	}
	result, err := s.DB.ExecContext(ctx, `
		INSERT INTO prompt_usage (prompt_name, model, params, timestamp, tokens_used, success, session_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, usage.PromptName, usage.Model, usage.Params, usage.Timestamp, usage.TokensUsed, boolToInt(usage.Success), sessionID)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// RecentUsage returns recent prompt executions.
func (s *PromptsStore) RecentUsage(ctx context.Context, limit int) ([]PromptUsage, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.DB.QueryContext(ctx, `
		SELECT id, prompt_name, model, params, timestamp, tokens_used, success, session_id
		FROM prompt_usage
		ORDER BY timestamp DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PromptUsage
	for rows.Next() {
		var u PromptUsage
		var successInt int
		var sessionID sql.NullString

		err := rows.Scan(&u.ID, &u.PromptName, &u.Model, &u.Params, &u.Timestamp, &u.TokensUsed, &successInt, &sessionID)
		if err != nil {
			return nil, err
		}

		u.Success = successInt != 0
		u.SessionID = sessionID.String
		results = append(results, u)
	}

	return results, rows.Err()
}

// Stats returns usage statistics for a prompt.
func (s *PromptsStore) Stats(ctx context.Context, promptName string) (*PromptStats, error) {
	var stats PromptStats
	var avgTokens sql.NullFloat64
	var totalTokens sql.NullInt64

	err := s.DB.QueryRowContext(ctx, `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN success = 1 THEN 1 ELSE 0 END) as successful,
			SUM(CASE WHEN success = 0 THEN 1 ELSE 0 END) as failed,
			AVG(tokens_used) as avg_tokens,
			SUM(tokens_used) as total_tokens
		FROM prompt_usage
		WHERE prompt_name = ?
	`, promptName).Scan(&stats.Total, &stats.Successful, &stats.Failed, &avgTokens, &totalTokens)

	if err != nil {
		return nil, err
	}

	stats.PromptName = promptName
	stats.AvgTokens = avgTokens.Float64
	stats.TotalTokens = totalTokens.Int64

	return &stats, nil
}

// MostUsed returns the most frequently used prompts.
func (s *PromptsStore) MostUsed(ctx context.Context, limit int) ([]struct {
	Name  string
	Count int
}, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.DB.QueryContext(ctx, `
		SELECT prompt_name, COUNT(*) as usage_count
		FROM prompt_usage
		GROUP BY prompt_name
		ORDER BY usage_count DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Name  string
		Count int
	}
	for rows.Next() {
		var item struct {
			Name  string
			Count int
		}
		if err := rows.Scan(&item.Name, &item.Count); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	return results, rows.Err()
}

// NewPromptUsage creates a PromptUsage record with current timestamp.
func NewPromptUsage(promptName, model, params string, tokensUsed int, success bool, sessionID string) PromptUsage {
	return PromptUsage{
		PromptName: promptName,
		Model:      model,
		Params:     params,
		Timestamp:  time.Now().Unix(),
		TokensUsed: tokensUsed,
		Success:    success,
		SessionID:  sessionID,
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
