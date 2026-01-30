# arc-prompt

Manage prompt templates for AI-powered CLI tools.

## Features

- **list** - List all available prompts
- **show** - Display prompt details
- **path** - Show prompts directory
- **edit** - Open prompt in $EDITOR
- **stats** - Show usage statistics
- **history** - View prompt usage history

## Installation

```bash
go install github.com/mtreilly/arc-prompt@latest
```

## Usage

```bash
# List all prompts
arc-prompt list

# Show a specific prompt
arc-prompt show code-review

# Show prompts directory
arc-prompt path

# Edit a prompt
arc-prompt edit analyze-repo

# View usage statistics
arc-prompt stats

# View history
arc-prompt history
```

## Prompt Format

Prompts are YAML files stored in `~/.config/arc/prompts/` and use `text/template` syntax for variable interpolation.

## License

MIT
