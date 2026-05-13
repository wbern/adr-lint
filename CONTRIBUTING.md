# Contributing to adr-lint

## Local setup

Two things to install on top of the Go toolchain:

```bash
brew install lefthook gitleaks
go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.0

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
[goreleaser](https://goreleaser.com/):

1. Conventional Commits land on `main`.
2. release-please maintains a "Release PR" with the proposed next version
   and an auto-generated changelog.
3. Merging the Release PR tags `vX.Y.Z` and creates a GitHub Release.
4. goreleaser builds cross-platform binaries and updates the Homebrew tap.

You never tag manually.

## Tests

```bash
cd go && go test ./...
```

The `pre-push` hook runs the full suite before letting you push.
