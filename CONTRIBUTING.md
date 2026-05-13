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
but available locally either manually or as an opt-in pre-commit hook.

### Manual flow (default)

```bash
git add <files>
adr-lint                          # check staged diff against all applicable ADRs
git commit -m "feat: ..."         # lefthook then runs gofmt/golangci-lint/gitleaks
```

### Opt-in pre-commit hook

If you'd rather not remember to run `adr-lint` between `add` and
`commit`, enable the hook by exporting one env var in your shell rc:

```bash
# ~/.zshrc or ~/.bashrc
export ADR_LINT_HOOK=1
```

After that, `git commit` will run `adr-lint` automatically as part of
the pre-commit checks. To temporarily skip a commit's ADR check
without unsetting the var: `LEFTHOOK_EXCLUDE=adr-lint git commit ...`.

The hook is off by default because every run consumes Claude Code
subscription quota — opting in is a personal choice about how much of
that quota you want to spend on commit-time feedback.

`adr-lint` operates on the staged diff by default. It picks up only the
ADRs whose `applies_to` globs match the staged files, then for each one
either short-circuits via `pre_filter` (zero cost) or calls Claude to
evaluate. Output is file:line for each violation, with a concrete fix.

If a check fails: edit the file, `git add` again, re-run `adr-lint`,
then `git commit`. Lefthook's gofmt/golangci/gitleaks pre-commit hooks
are independent and run on every commit — they're free, they always run.

### Other modes

```bash
adr-lint -v                       # verbose: show every applicable ADR
                                  # (passed + failed), with pre-filter reasons
adr-lint --branch origin/main     # lint everything that has diverged from main
                                  # — useful before pushing a long-running branch
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
