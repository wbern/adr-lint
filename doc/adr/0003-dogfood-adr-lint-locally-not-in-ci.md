---
status: accepted
date: 2026-05-13
applies_to:
  - ".github/workflows/**/*.yml"
  - "lefthook.yml"
pre_filter:
  - "adr-lint"
  - "adr_lint"
---

# 3. Dogfood adr-lint locally, not in CI

## Context

adr-lint is the project's own architectural linter. Each run shells
out to the Claude Code CLI, which authenticates against the user's
Claude Code subscription. The subscription is **flat-rate** — there's
no per-call dollar charge — but every run does consume that account's
rate-limit quota.

Two enforcement loci to consider:

- **In CI** — every PR triggers `adr-lint`. To work, CI needs an auth
  token for *someone's* Claude Code subscription stashed as a repo
  secret. That account's quota is then burned by every push: feature
  branches, dependabot bumps, docs typo fixes, force-pushes during PR
  iteration. PR latency goes up by however long the Claude calls take.
  Also: external contributors' PRs would consume the maintainer's
  quota, with no rate limiting beyond what the subscription enforces.
- **Locally (manual)** — the maintainer runs `adr-lint` against
  staged changes before committing. Uses their own logged-in Claude
  Code session. Tight feedback loop, no shared secrets, but relies on
  discipline.
- **Locally (lefthook pre-commit hook)** — same as manual but
  automatic. Same auth model. Every commit pays a quota cost, but
  only for the committer's own work.

## Decision

We will run **adr-lint locally only**, never in CI.

The local hook integration is **opt-in**, not enabled by default. The
canonical local invocation is a manual `adr-lint` between `git add`
and `git commit`, documented in `CONTRIBUTING.md`. Contributors who
want it wired into `pre-commit` can flip the lefthook command on by
setting `ADR_LINT_HOOK=1` in their shell environment.

**Forbidden:**

```yaml
# Do not add adr-lint as a CI job.
jobs:
  adr-lint:
    steps:
      - env:
          ANTHROPIC_API_KEY: ${{ secrets.MAINTAINER_CLAUDE_TOKEN }}
        run: adr-lint
```

**Acceptable:**

```bash
# Manual flow (default).
git add <files>
adr-lint
git commit -m "feat: ..."

# Or opt-in to the pre-commit hook for the same flow without the manual step:
export ADR_LINT_HOOK=1
# adr-lint now runs automatically as part of `git commit`.
```

If this project ever gains real contributors, supersede this ADR
rather than silently adding a CI workflow — the auth-token-in-secret
question deserves a fresh decision.

## Consequences

**Positive:**
- No shared Claude Code auth token in CI secrets.
- No PR latency from Claude calls.
- Dependabot / docs / chore pushes don't burn quota.
- Maintainer keeps full control of when their subscription is consumed.

**Negative:**
- A violation introduced in a commit can land on `main` if the
  maintainer forgets to run the check (manual flow) or hasn't enabled
  the hook (`ADR_LINT_HOOK=1`).
- New contributors won't get automatic ADR feedback on their PRs.

**Neutral:**
- The opt-in env-var pattern means the hook is documented and ready,
  but disabled by default — flipping it on is a one-line shell rc edit.

## References

- ADR-0001 (Claude is the only LLM provider) — establishes that every
  run is a Claude Code subscription call.
