# Contributing to adr-lint

## Local setup

Two things to install on top of the Go toolchain:

```bash
brew install lefthook gitleaks golangci-lint

# Wire the git hooks into this clone (idempotent)
lefthook install
```

That's it. The hooks live in [`lefthook.yml`](lefthook.yml) and run on commit
and push.

## Commit messages

This repo uses [Conventional Commits](https://www.conventionalcommits.org/).
The `commit-msg` hook rejects anything else; CI re-validates PR titles since
squash-merges adopt the PR title as the commit message.

Format:

```
<type>(<optional scope>)<optional !>: <subject>
```

Allowed types:

| Type       | Use for                                                  | Triggers release? |
| ---------- | -------------------------------------------------------- | ----------------- |
| `feat`     | New feature                                              | minor             |
| `fix`      | Bug fix                                                  | patch             |
| `perf`     | Performance improvement                                  | patch             |
| `refactor` | Internal change, no behavior change                      | no                |
| `docs`     | Documentation only                                       | no                |
| `test`     | Test-only changes                                        | no                |
| `build`    | Build system, dependencies                               | no                |
| `ci`       | CI configuration                                         | no                |
| `chore`    | Tooling, repo housekeeping                               | no                |
| `style`    | Whitespace, formatting (no logic change)                 | no                |
| `revert`   | Revert a previous commit                                 | varies            |

Breaking changes: append `!` after the type/scope or add a `BREAKING CHANGE:`
footer. Either bumps the major version.

Examples:

```
feat(create): add --template flag
fix(runner): handle empty diff without panicking
feat(api)!: rename --files to --paths
docs: clarify pre_filter semantics in README
chore(deps): bump golangci-lint to v1.62
```

## Releases

Driven by [release-please](https://github.com/googleapis/release-please) +
[goreleaser](https://goreleaser.com/). Trunk-based — no manual tagging,
no review gate.

1. Conventional Commits land on `main`.
2. release-please opens a "Release PR" with the proposed next version and
   an auto-generated changelog.
3. The `auto-merge` job in `release-please.yml` squash-merges that PR
   immediately. release-please then tags `vX.Y.Z` and creates the GitHub
   Release.
4. The tag push triggers `release.yml`, which runs goreleaser to build
   cross-platform binaries and update the Homebrew tap.

See [ADR-0002](doc/adr/0002-trunk-based-releases-via-release-please-auto-merge.md)
for the design rationale.

## Tests

```bash
cd go && go test ./...
```

The `pre-push` hook runs the full suite before letting you push.

## Dogfooding adr-lint on adr-lint

This repo is its own first user. The check is **off in CI by design**
(see [ADR-0003](doc/adr/0003-dogfood-adr-lint-locally-not-in-ci.md))
but **on by default locally** as a lefthook pre-commit step.

### Default flow

```bash
git add <files>
git commit -m "feat: ..."         # lefthook runs adr-lint + gofmt/golangci-lint/gitleaks
```

`adr-lint` operates on the staged diff. It picks up only the ADRs
whose `applies_to` globs match the staged files, then for each one
either short-circuits via `pre_filter` (zero cost) or calls Claude to
evaluate. Output is file:line for each violation, with a concrete fix.

If a check fails: edit the file, `git add` again, then `git commit`
again — the hook re-runs against the new staged diff.

### Skipping the ADR check

For a single commit (formatting-only, docs typo, etc.):

```bash
ADR_LINT_SKIP=1 git commit -m "..."
# or, equivalently, the lefthook built-in:
LEFTHOOK_EXCLUDE=adr-lint git commit -m "..."
```

Both leave the other pre-commit hooks (gofmt, golangci, gitleaks)
running — they're free and always worth it.

### Other modes

```bash
adr-lint -v                       # verbose: show every applicable ADR
                                  # (passed + failed), with pre-filter reasons
adr-lint --branch                 # lint everything on the current branch that
                                  # has diverged from main (merge-base..HEAD)
                                  # — useful before pushing a long-running branch
adr-lint --branch feat/other      # same, but for an explicit ref instead of HEAD
adr-lint --files path/to/file.go  # lint a specific file even if not staged
```

### When to skip

`adr-lint` makes a real Claude call per applicable ADR (minus pre-filter
hits). Skip it for changes that obviously can't violate an architectural
constraint:

- Pure formatting/whitespace
- README/docs typo fixes
- Comment-only edits

Use judgement. The ADRs themselves are short — if you've read them
recently and the change clearly doesn't touch the surface they govern,
just commit.
