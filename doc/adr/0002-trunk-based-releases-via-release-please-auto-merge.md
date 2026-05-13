---
status: accepted
date: 2026-05-13
applies_to:
  - ".github/workflows/release-please.yml"
  - ".github/workflows/release.yml"
  - ".goreleaser.yaml"
  - "release-please-config.json"
  - ".release-please-manifest.json"
pre_filter:
  - "release-please"
  - "goreleaser"
---

# 2. Trunk-based releases via release-please auto-merge

## Context

The release pipeline needs to turn Conventional Commits on `main` into
tagged GitHub Releases with cross-platform binaries and an updated
Homebrew formula, with no manual tagging step.

release-please is the standard tool in this space for Go, but its model
is PR-based: it opens a Release PR on every push to `main`, accumulates
commits into it, and only releases when the PR is merged. That review
gate is a poor fit for a single-maintainer trunk-based workflow, where
PRs are friction rather than a quality signal.

Alternatives considered:

- **svu + goreleaser changelog** — pure tag-based, no PR. Loses the
  in-repo `CHANGELOG.md`, the rich per-type section grouping, and
  introduces a new tool. No batching mechanism, so every feat/fix
  commit becomes its own release.
- **Manual tagging** — predictable but requires remembering to do it,
  and the maintainer explicitly rejected this.
- **release-please without auto-merge** — keeps the friction we want
  to remove.

## Decision

We will use **release-please** with an **auto-merge job** that squashes
the Release PR immediately after release-please opens or updates it.
The PR exists only as a one-step staging area, not a review gate.

Required structure:

- `release-please.yml` runs on push to `main`, uses a PAT
  (`RELEASE_PLEASE_TOKEN`) instead of `GITHUB_TOKEN` so that the tag
  push it creates actually triggers downstream workflows. Tags pushed
  by `GITHUB_TOKEN` are deliberately ignored by Actions.
- An `auto-merge` job in the same workflow finds the open Release PR
  by the `autorelease: pending` label and squash-merges it via
  `gh pr merge --squash`. Looking up by label (rather than relying on
  release-please's `pr` output) handles the case where the action
  reconciles an existing PR without updating it.
- `release.yml` runs on tag push `v*` and invokes goreleaser, which
  appends artifacts to the GitHub Release that release-please already
  created (`release.mode: append` in `.goreleaser.yaml`).

**Required:** the release-please action uses the PAT.

```yaml
- uses: googleapis/release-please-action@v4
  with:
    token: ${{ secrets.RELEASE_PLEASE_TOKEN }}
```

**Forbidden:** falling back to `GITHUB_TOKEN`, which silently breaks
the chain because its tag push won't fire `release.yml`.

```yaml
- uses: googleapis/release-please-action@v4
  # ❌ tag will be created but goreleaser will never run
```

## Consequences

**Positive:**
- Every `feat:` / `fix:` push to `main` releases automatically.
- No human in the release loop; no forgotten tags.
- `CHANGELOG.md` lives in the repo with full grouping by commit type.
- Goreleaser builds and the Homebrew formula update happen on the
  same trigger, so `brew install` always tracks the latest tag.

**Negative:**
- A PAT is required and must be rotated when it expires. The token
  needs `Contents: read/write` + `Pull requests: read/write` on this
  repo only.
- No CI gate between release-please opening the PR and the merge. The
  Release PR only modifies `CHANGELOG.md` and the manifest, so this is
  acceptable, but a regression in those files would land directly.
- Rapid bursts of commits produce one release each (no batching).

**Neutral:**
- The PR briefly appears in the PR list before auto-merge clears it.

## References

- googleapis/release-please-action#1110 (tag push from GITHUB_TOKEN
  does not trigger workflows)
- Goreleaser docs: `release.mode: append`
