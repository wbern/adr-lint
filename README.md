# ADR Lint

Automatically validates code changes against Architecture Decision Records (ADRs) using Claude as the analysis backend.

## How It Works

This tool checks staged or branch changes against relevant ADRs. If your code
violates an architectural decision, you get feedback with explanations and
suggestions.

**Suggested workflow:**

1. Install a pre-commit hook so staged files are checked against applicable
   ADRs before each commit. A working sample lives at `scripts/pre-commit`:

   ```bash
   ln -s ../../scripts/pre-commit .git/hooks/pre-commit
   ```

2. Wire `adr-lint --branch` into CI to check PR changes against `main`. The
   tool depends on the Claude Code CLI for analysis, so your runner needs to
   have it installed and authenticated — typically a self-hosted runner.
   `.github/workflows/test.yml` in this repo runs the Go test suite; it is
   not (yet) a reference for running `adr-lint` itself in CI.

## Managing ADRs

The CLI has bd-style subcommands for managing the ADRs in `doc/adr/`:

```bash
adr-lint create "Use Testify for tests"   # scaffold doc/adr/NNNN-use-testify-for-tests.md
adr-lint list                              # one line per ADR (id, status, title)
adr-lint show 1                            # raw file contents
adr-lint accept 1                          # flip frontmatter status: accepted
adr-lint reject 1                          # status: rejected
adr-lint withdraw 1                        # status: withdrawn
adr-lint deprecate 1                       # status: deprecated
adr-lint supersede 1 2                     # status: superseded + superseded_by: "0002"
adr-lint validate                          # check cross-refs, IDs, status invariants
adr-lint version                           # print binary version
adr-lint help                              # subcommand reference
adr-lint <sub> --help                      # per-subcommand usage
```

`adr-lint create` scaffolds an ADR from
[`doc/adr/templates/template.md`](doc/adr/templates/template.md) when present
(substituting `{{number}}` and `{{title}}`), falling back to a built-in
minimum frontmatter (`status: proposed`, `applies_to: ["**/*"]`) otherwise.
The template documents every field the parser understands. For a guided
discovery flow that also drafts the body, the `/create-adr` Claude Code slash
command remains available.

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
