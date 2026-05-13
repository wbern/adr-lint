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

adr-lint is the project's own architectural linter. Running it on this
repo's pull requests would close the loop — but every invocation issues
real Claude API calls, which cost real money. CI cost was the explicit
constraint when this decision was made.

Two ways to enforce ADRs:

- **In CI** — every PR triggers `adr-lint` against the diff. Catches
  violations from any contributor automatically. Costs money on every
  push, including no-op pushes (docs typos, CI tweaks, dependabot).
- **Locally** — the maintainer runs `adr-lint` against staged changes
  before pushing. Zero CI spend. Relies on discipline, but works fine
  for a single-maintainer project.

## Decision

We will run **adr-lint locally only**, not in CI. The maintainer is
expected to run it against staged changes before pushing significant
work. The exact local invocation is documented in `CONTRIBUTING.md`.

**Forbidden:**

```yaml
# Do not add adr-lint as a CI job.
# Do not wire adr-lint into lefthook pre-commit / pre-push either —
# pre-push fires on every push and would surprise-bill the maintainer.
jobs:
  adr-lint:
    steps:
      - run: adr-lint check
```

**Acceptable today:**

```bash
# Run on demand before pushing a feature commit.
git add <files>
adr-lint          # no subcommand = lint the staged diff
```

If this project gains contributors or moves to a model where CI spend
is acceptable, supersede this ADR — do not silently flip the switch by
adding a workflow.

## Consequences

**Positive:**
- Zero ongoing CI cost for ADR enforcement.
- Lefthook + golangci-lint + gitleaks still gate every commit; the
  ADR check is the only intentionally manual gate.

**Negative:**
- A violation introduced in a commit can land on `main` if the
  maintainer forgets to run the check.
- New contributors won't get automatic ADR feedback on their PRs.

**Neutral:**
- The CI infrastructure (`ci.yml`) is unchanged; the absence of an
  `adr-lint` job is the decision itself.

## References

- ADR-0001 (Claude is the only LLM provider) — establishes that every
  run is a Claude Code subscription call, which is the cost being
  avoided here.
