---
description: Create a new Architecture Decision Record (ADR)
argument-hint: <title or topic of the architectural decision>
---

# Create ADR: Architecture Decision Record Creator

Create a new ADR to document an architectural decision. ADRs capture the "why" behind technical choices, helping future developers understand constraints and tradeoffs.

> **Format**: ADRs follow the structure described below — frontmatter with `status`, `date`, `applies_to`, plus Context / Decision / Consequences / References sections.

## Input

$ARGUMENTS

(If no input provided, ask user for the architectural decision topic)

## Process

### Step 1: Detect ADR Directory and Determine Number

Find existing ADR location and calculate next number:

```bash
# Check common ADR directories (in order of preference)
for dir in doc/adr docs/adr decisions doc/architecture/decisions; do
  if [ -d "$dir" ]; then
    echo "Found: $dir"
    ls "$dir"/*.md 2>/dev/null | grep -E '/[0-9]{4}-' | sort | tail -1
    break
  fi
done
```

If no ADR directory exists:
1. Ask user which location to use (default: `doc/adr/`)
2. Create the directory

Calculate next number:
1. Extract highest existing number
2. Increment by 1
3. Format as 4-digit zero-padded (e.g., `0001`, `0012`)

### Step 2: Discovery Questions

Gather context through conversation (use AskUserQuestion for structured choices):

**Context & Problem**
- What forces are at play? (technological, social, project constraints)
- What problem, pattern, or situation prompted this decision?
- What triggered the need to decide now? (bug, confusion, inconsistency, new requirement)
- Are there related PRs, issues, or prior discussions to reference?

**The Decision**
- What are we deciding to do (or not do)?
- What alternatives were considered?
- Why was this approach chosen over alternatives?

**Consequences**
- What becomes easier or more consistent with this decision?
- What becomes harder, more constrained, or riskier?
- What tradeoffs are we explicitly accepting?

**Scope**
- Which parts of the codebase does this apply to?
- Are there exceptions or areas where this doesn't apply?

### Step 3: Generate ADR File

Create `{adr_directory}/NNNN-title-slug.md`:
- Convert title to kebab-case slug (lowercase, hyphens, no special chars)
- Use today's date for the `date` field
- Default status to `accepted` (most ADRs are written after the decision is made)

**ADR Template:**

```markdown
---
status: accepted
date: YYYY-MM-DD
applies_to:
  - "**/*.ts"
  - "**/*.tsx"
---

# N. Title

## Context

[Forces at play - technological, social, project constraints.
What problem prompted this? Value-neutral description of the situation.]

## Decision

We will [decision statement in active voice].

[If the decision involves code patterns, include concrete examples:]

**Forbidden pattern:**
\`\`\`typescript
// ❌ BAD - [explanation]
[example of what NOT to do]
\`\`\`

**Required pattern:**
\`\`\`typescript
// ✅ GOOD - [explanation]
[example of what TO do]
\`\`\`

## Consequences

**Positive:**
- [What becomes easier]
- [What becomes more consistent]

**Negative:**
- [What becomes harder]
- [What constraints we accept]

**Neutral:**
- [Other impacts worth noting]

## References

- [Related PRs, issues, or documentation]
```

### Step 4: Refine applies_to Scope

Help user define which files this decision applies to using glob patterns:

Common patterns:
- **/*.ts - All TypeScript files
- **/*.tsx - All React component files
- src/components/** - Only files in components directory
- Prefix with ! to exclude (e.g., test files, type definitions)
- packages/api/** - Specific package only

If the decision applies broadly, use ["**/*"].

**Note**: `applies_to` is required. The linter uses these patterns to determine which files to check against this ADR.

### Step 5: Confirm and Write

Show the complete ADR content and ask user to confirm before writing.

After creation, suggest:
- Review the ADR for completeness
- Commit with `/commit`

## Tips for Good ADRs

1. **Focus on the "why"** - The decision itself may be obvious; the reasoning often isn't
2. **Keep it concise** - 1-2 pages maximum; should be readable in 5 minutes
3. **Use active voice** - "We will use X" not "X will be used"
4. **Include concrete examples** - Code examples make abstract decisions tangible
5. **Document tradeoffs honestly** - Every decision has costs; be explicit about them
6. **Link to context** - Reference PRs, issues, or discussions where the decision was made
7. **Be specific about scope** - Use `applies_to` patterns to clarify affected code

## Status Values

| Status | When to Use |
|--------|-------------|
| `proposed` | Under discussion, not yet agreed |
| `accepted` | Agreed upon and should be followed |
| `deprecated` | No longer relevant (context changed) |
| `superseded` | Replaced by another ADR (link to it) |

To supersede an existing ADR:
1. Create new ADR with the updated decision
2. Update old ADR's status to `superseded by ADR-NNNN`

## Optional Frontmatter (Advanced)

Most ADRs don't need these fields - only add them when necessary:

```yaml
complexity: standard      # ultralite | lite | standard | complex (linter model selection)
pre_filter: "keyword"     # Skip LLM check if keyword not in diff (optimization)
enforced_by: eslint       # When rule is enforced by other tooling
```

## Related Commands

- After creating: Commit with `/commit`

## References

- [`doc/adr/templates/template.md`](../../doc/adr/templates/template.md) - The canonical ADR template
- [`src/`](../../src) - The ADR linting implementation
