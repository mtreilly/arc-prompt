// Copyright (c) 2025 Arc Engineering
// SPDX-License-Identifier: MIT

package prompt

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// Prompt represents a prompt template with system and user messages.
type Prompt struct {
	Name     string            `yaml:"name"`
	System   string            `yaml:"system,omitempty"`
	User     string            `yaml:"user,omitempty"`
	Model    string            `yaml:"model,omitempty"`
	Defaults map[string]string `yaml:"defaults,omitempty"`
	Options  map[string]any    `yaml:"options,omitempty"`
	Metadata map[string]any    `yaml:"metadata,omitempty"`

	// Private fields for compiled templates
	systemTmpl *template.Template
	userTmpl   *template.Template
}

// Load reads a prompt from a YAML file.
// The name can be:
// - A basename (e.g., "code-review") - looks in Dir()
// - A relative path (e.g., "./custom-prompts/review.yaml")
// - An absolute path (e.g., "/path/to/prompt.yaml")
func Load(name string) (*Prompt, error) {
	path, err := resolvePath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read prompt %q: %w", name, err)
	}

	var p Prompt
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parse prompt %q: %w", name, err)
	}

	// Set name from argument if not in YAML
	if p.Name == "" {
		p.Name = name
	}

	// Precompile templates with strict error handling on missing keys
	if p.System != "" {
		p.systemTmpl, err = template.New("system").Option("missingkey=error").Parse(p.System)
		if err != nil {
			return nil, fmt.Errorf("parse system template in %q: %w", name, err)
		}
	}

	if p.User != "" {
		p.userTmpl, err = template.New("user").Option("missingkey=error").Parse(p.User)
		if err != nil {
			return nil, fmt.Errorf("parse user template in %q: %w", name, err)
		}
	}

	return &p, nil
}

// Execute renders the prompt with given parameters.
// Parameters are merged with defaults (params override defaults).
// Returns (system, user, error).
func (p *Prompt) Execute(params map[string]string) (system, user string, err error) {
	// Merge defaults with params (params take precedence)
	data := make(map[string]string)
	for k, v := range p.Defaults {
		data[k] = v
	}
	for k, v := range params {
		data[k] = v
	}

	// Execute system template
	if p.systemTmpl != nil {
		var buf bytes.Buffer
		if err := p.systemTmpl.Execute(&buf, data); err != nil {
			return "", "", fmt.Errorf("execute system template: %w", err)
		}
		system = buf.String()
	}

	// Execute user template
	if p.userTmpl != nil {
		var buf bytes.Buffer
		if err := p.userTmpl.Execute(&buf, data); err != nil {
			return "", "", fmt.Errorf("execute user template: %w", err)
		}
		user = buf.String()
	}

	return system, user, nil
}

// resolvePath resolves a prompt name to a file path.
func resolvePath(name string) (string, error) {
	// Absolute path
	if filepath.IsAbs(name) {
		if _, err := os.Stat(name); err != nil {
			return "", fmt.Errorf("prompt file not found: %s", name)
		}
		return name, nil
	}

	// Relative path (contains /)
	if strings.Contains(name, string(filepath.Separator)) {
		if _, err := os.Stat(name); err != nil {
			return "", fmt.Errorf("prompt file not found: %s", name)
		}
		return name, nil
	}

	// Basename - look in prompts directory
	dir, err := Dir()
	if err != nil {
		return "", err
	}

	// Try with .yaml extension
	path := filepath.Join(dir, name)
	if !strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml") {
		path += ".yaml"
	}

	if _, err := os.Stat(path); err != nil {
		// Try .yml as fallback
		if strings.HasSuffix(path, ".yaml") {
			altPath := strings.TrimSuffix(path, ".yaml") + ".yml"
			if _, err := os.Stat(altPath); err == nil {
				return altPath, nil
			}
		}
		return "", fmt.Errorf("prompt %q not found in %s", name, dir)
	}

	return path, nil
}

// Dir returns the user's prompt directory (~/.config/arc/prompts).
// Creates the directory if it doesn't exist.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("get home directory: %w", err)
	}

	dir := filepath.Join(home, ".config", "arc", "prompts")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create prompts directory: %w", err)
	}

	return dir, nil
}

// List returns all available prompts in the prompts directory.
func List() ([]string, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read prompts directory: %w", err)
	}

	var prompts []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		// Strip extension
		if strings.HasSuffix(name, ".yaml") {
			name = strings.TrimSuffix(name, ".yaml")
		} else {
			name = strings.TrimSuffix(name, ".yml")
		}

		prompts = append(prompts, name)
	}

	return prompts, nil
}

// Exists checks if a prompt with the given name exists.
func Exists(name string) bool {
	_, err := resolvePath(name)
	return err == nil
}
