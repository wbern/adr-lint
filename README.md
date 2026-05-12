# ADR Lint

Automatically validates code changes against Architecture Decision Records (ADRs) using Claude as the analysis backend.

## How It Works

When you commit code or open a PR, this tool checks your changes against relevant ADRs. If your code violates an architectural decision, you'll get immediate feedback with explanations and suggestions.

**The flow is automatic:**

1. Pre-commit hook checks staged files against applicable ADRs
2. CI checks all PR changes against main branch
3. Results appear in your terminal (local) or as PR comments (CI)

## Creating ADRs

Use the `/create-adr` command in Claude Code:

```
/create-adr <topic>
```

This walks you through the discovery process and generates a properly formatted ADR with the correct frontmatter fields.

See [`doc/adr/templates/template.md`](doc/adr/templates/template.md) for the format and the `/create-adr` command documentation for the full field reference.

## Setup

The tool uses the Claude Code CLI as its analysis backend. Install it and log in once; no API key is required afterward — analysis runs against your existing Claude Code subscription.

## Manual Usage

Usually the hooks handle everything. For debugging or one-off runs:

```bash
# Check staged files
adr-lint

# Check all changes against main
adr-lint --branch

# Check specific files
adr-lint --files pkg/foo.go pkg/bar.go

# Preview what would be checked
adr-lint --dry-run

# Additional options
adr-lint --no-cache         # Skip cache
adr-lint --verbose          # Detailed output
adr-lint --per-file         # One chunk per file (slower, more precise)
```
