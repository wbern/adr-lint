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

This repo is its own first user. After a non-trivial change, run the
linter against staged work before pushing:

```bash
git add <files>
adr-lint            # no subcommand = lint the staged diff
```

We deliberately do **not** run this in CI — every check costs real
Claude API spend, and the trunk-based release flow ships fast enough
that bloating CI with paid steps isn't worth it. The check is the
maintainer's responsibility. See
[ADR-0003](doc/adr/0003-dogfood-adr-lint-locally-not-in-ci.md).
