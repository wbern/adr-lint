# ADR Lint

Automatically validates code changes against Architecture Decision Records (ADRs) using Claude as the analysis backend.

![demo](docs/demo-all.gif)

(Fresh repo to a caught violation in one take: `adr-lint create` scaffolds the
ADR, frontmatter narrows what it applies to, then staging a `fmt.Println` that
violates the decision lights up `❌` with a concrete fix. Reproduce with
[`./scripts/demo/record_demo.sh`](scripts/demo/record_demo.sh) — needs
`asciinema`, `agg`, and the `claude` CLI on PATH.)

## What it solves

ADRs are how teams record architectural decisions, but they drift out of
sync with the code as soon as the ink dries — nothing checks new diffs
against the rules. `adr-lint` closes that loop: every commit (or PR) is
read against the ADRs in `doc/adr/`, and violations come back with the
file, line, and a suggested fix.

## How it works

For each lint run:

1. **Collect a diff** — staged files by default, the full branch diff vs
   `main` with `--branch`, or specific paths with `--files`.
2. **Match `applies_to`** — each ADR's glob list decides whether it cares
   about any of the changed paths. Non-matching ADRs are skipped.
3. **Apply `pre_filter`** — a substring shortcut. If none of the ADR's
   pre-filter strings appear anywhere in the diff, the LLM call is skipped
   entirely and the ADR passes. This is the difference between a free
   re-run and a paid one.
4. **Ask Claude** — surviving ADRs are sent to the Claude Code CLI with
   the diff. The model returns pass/fail + location + a fix.

The Claude CLI is the analysis backend — there's no API key plumbing, the
tool shells out to `claude` and inherits whatever auth your account
already has.

## Quickstart

```bash
# 1. Install + log in to the Claude Code CLI (one-time, no API key needed)
#    https://claude.com/claude-code

# 2. Install adr-lint
go install github.com/wbern/adr-lint/go/cmd/adr-lint@latest

# 3. Write your first ADR
adr-lint create "Use the logger package instead of fmt.Println"
# ...edit doc/adr/0001-*.md: tighten applies_to + pre_filter, write the decision
adr-lint accept 1

# 4. Wire it into git
ln -s ../../scripts/pre-commit .git/hooks/pre-commit

# Every commit from here on checks staged files against accepted ADRs.
```

## ADR file format

ADRs live in `doc/adr/NNNN-slug.md` with YAML frontmatter that controls
how the linter treats them. The full annotated template is at
[`doc/adr/templates/template.md`](doc/adr/templates/template.md) — copy
it into your project to customize the scaffold.

| Field           | Purpose                                                                                  |
| --------------- | ---------------------------------------------------------------------------------------- |
| `status`        | `proposed` / `accepted` / `rejected` / `withdrawn` / `deprecated` / `superseded`         |
| `applies_to`    | Doublestar globs; `!`-prefix negates. Defaults to `["**/*"]`.                            |
| `pre_filter`    | Substrings that must appear in the diff for the LLM to be invoked. Free pass otherwise.  |
| `complexity`    | `lite` / `standard` / `complex` — controls chunking and how much context Claude sees.    |
| `enforced_by`   | Marks the ADR as covered by external tooling (eslint rule, type check). LLM skips it.    |
| `diff_context`  | `false` evaluates each file in isolation. Defaults to `true`.                            |
| `superseded_by` | Set automatically by `adr-lint supersede`; points at the replacement.                    |

Only `status` is required. Everything else has sensible defaults — the
goal is that an ADR with no frontmatter still works, and you add fields
only when you want to tighten scope or speed things up.

## Commands

### Managing ADRs

```bash
adr-lint create "Use Testify for tests"  # scaffold doc/adr/NNNN-*.md from template
adr-lint list                            # id, status, title (one per line)
adr-lint show 1                          # raw file contents
adr-lint accept 1                        # flip status: accepted
adr-lint reject 1                        # status: rejected
adr-lint withdraw 1                      # status: withdrawn
adr-lint deprecate 1                     # status: deprecated
adr-lint supersede 1 2                   # 0001 → superseded; writes superseded_by: "0002"
adr-lint validate                        # cross-refs, IDs, status invariants
adr-lint version                         # print binary version
adr-lint help                            # subcommand reference
```

`supersede` writes both halves of the link so `list` can surface the
relationship without you opening the file.

### Running the lint

```bash
adr-lint                       # check staged files (default; matches the pre-commit hook)
adr-lint --branch              # check the full diff vs main (PR-review mode)
adr-lint --files pkg/foo.go    # check specific paths
adr-lint --dry-run             # show which ADRs would run; skip LLM calls
adr-lint --verbose             # print provider, mode, and applicable ADRs
adr-lint --no-cache            # bypass the result cache
adr-lint --per-file            # one chunk per file (slower, more precise)
```

## Integration

### Pre-commit hook

The sample at [`scripts/pre-commit`](scripts/pre-commit) runs `adr-lint`
on staged files and exits cleanly if the binary isn't on PATH (so
collaborators without it aren't blocked).

```bash
ln -s ../../scripts/pre-commit .git/hooks/pre-commit
```

### CI (PR review)

`adr-lint --branch` is designed for CI: it lints the entire diff that
would land in the PR, no staging required. The runner needs the Claude
Code CLI installed and authenticated, which in practice means a
self-hosted runner. The `.github/workflows/test.yml` in this repo only
runs the Go test suite — it's not a reference for running `adr-lint`
itself in CI.

## Focused demos

The hero GIF above is the full tour. The per-section ones under
[`docs/`](docs/) are shorter and useful for pointing a colleague at a
single slice:

- [`demo-create.gif`](docs/demo-create.gif) — author your first ADR
- [`demo-lint.gif`](docs/demo-lint.gif) — violation caught, fix, re-run hits the pre-filter shortcut
- [`demo-branch.gif`](docs/demo-branch.gif) — PR review with `--branch`
- [`demo-lifecycle.gif`](docs/demo-lifecycle.gif) — supersede a decision

For a guided discovery flow that also drafts the ADR body, the
`/create-adr` Claude Code slash command remains available alongside the
CLI scaffold.
