# Design Note: Recommendation Specialist Architecture

**Status:** Proposed — not yet implemented
**Context:** Discovered while triaging content_rewrite failures (April 2026)
**Related:** P9 in pipeline-failures-report.md (partial fix already deployed)

---

## The Problem

LLM auditors (`content-quality-auditor`, `visual-design-auditor`) produce findings that mix two fundamentally different kinds of observation:

1. **Bugs** — factually broken content. Placeholder text, broken links, unrendered templates, empty sections, cross-site contamination. A reasonable human would agree these are wrong.

2. **Recommendations** — business or design opinions. Tone preferences, branding strategy ("use a branded email for trust"), differentiation advice, CTA strategy. A reasonable business owner might disagree.

The current pipeline treats both as `content_rewrite` items to be auto-fixed. This causes:

- **False-positive rewrites** — the pipeline rewrites things that weren't broken, producing content that then fails validation (e.g. inventing `hello@leopardessconsulting.co.uk` because the audit said the real `leopardess@contactforsales.com` should "match the company domain")
- **Wasted LLM cycles** on subjective findings the pipeline can't verify
- **Human review queue pollution** — HITL ends up being the default fallback for every recommendation, so humans get flooded with minor tone/branding opinions they shouldn't need to review

## Current State (Post-P9)

P9 fixed routing for `gap` findings (empty sections on existing pages → `needs_content_page` rebuild, not `content_rewrite`). This addresses the structural subset but not the opinion subset.

Still broken: categories like `content`, `differentiation`, `tone` on existing pages go to `content_rewrite` regardless of whether they describe a bug or a preference.

## Proposed Architecture

### Three-way classification

Auditor prompts emit an explicit `finding_type` field alongside `category`:

```json
{
  "finding_type": "bug" | "recommendation" | "gap",
  "category": "content" | "tone" | "differentiation" | ...,
  ...
}
```

Definitions for the LLM:
- **bug**: Content is factually broken. Placeholder text, broken link, unrendered template, empty section, inconsistent data, cross-site contamination. Deterministically verifiable.
- **recommendation**: Business or design opinion. Tone preference, branding strategy, differentiation approach, conversion optimisation. Subjective.
- **gap**: Missing content on a page that exists, or a page that should exist. Routed to build/rebuild.

### Routing by finding_type

| finding_type | Route to | Rationale |
|--------------|----------|-----------|
| `bug` | `content_rewrite` → page-build-handler | Factual fix, validator catches regressions |
| `gap` | `needs_content_page` → page-build-handler | Rebuild, not rewrite (P9 already handles this) |
| `recommendation` | Specialist agent (see below) | Not a bug — needs judgement |

### Recommendation Specialist Agents

Rather than defaulting all recommendations to `needs_human_review`, route by category to a specialist that can reason about the domain and decide:

| Category | Specialist | Decisions it can make |
|----------|-----------|----------------------|
| email/contact | `identity-advisor` (new) | Accept third-party emails if real; flag only true placeholders |
| tone | `tone-shift-agent` (new or extend `content-writer`) | Apply tone adjustment, or defer if conflicts with brand DNA |
| differentiation | `content-strategist` (new or extend `site-strategist`) | Rewrite with specific differentiators from site_specs, or defer if none exist |
| cta | `component-template-fixer` (exists) | Already handles CTA fixes |
| branding | `brand-designer` (exists) | Could extend to evaluate branding recommendations |

Each specialist returns one of:
- `apply` — make the change
- `dismiss` — finding is incorrect for this site, mark `wont_fix` with reasoning
- `escalate` — uncertain, route to `needs_human_review` with context

### Per-Site Approval Mode

Add `sites.approval_mode` column: `'auto' | 'review'`

- `auto` (default for most sites): specialists can apply recommendations directly
- `review` (for high-stakes / client-sensitive sites): all recommendations go to `needs_human_review` regardless of specialist decision

This is the escape hatch — a user wanting human oversight can opt in per-site without disabling the specialist routing globally.

The existing `site_work_items.approval_mode` column (`'auto'`) already exists at the item level. The site-level setting would be the default; individual items could override.

## Implementation Plan (Rough)

1. **Phase 1 — Classification**
   - Update both auditor prompts to emit `finding_type`
   - Update `auditFinding` struct in `write_audit_findings_action.go` to parse it
   - Add `finding_type` to routing logic (before category switch)

2. **Phase 2 — Identity Advisor (highest-impact specialist)**
   - New agent `identity-advisor` for contact/email/company-identity recommendations
   - Handles: third-party email validation, company name consistency, contact info choices
   - Returns apply/dismiss/escalate

3. **Phase 3 — Approval Mode**
   - Add `sites.approval_mode` column + migration
   - Respect it in specialist routing (if `review`, all recommendations → HITL)

4. **Phase 4 — Other Specialists**
   - Tone specialist, differentiation specialist, etc.
   - Each replaces a currently-broken `content_rewrite` path

5. **Phase 5 — Learning Loop**
   - Log specialist decisions with reasoning
   - Periodically review dismissed findings to improve auditor prompts
   - Findings the specialists consistently dismiss indicate prompt patterns to remove

## Why This Matters

The current model treats the LLM auditor as oracle — every finding is actionable. The proposed model treats it as collaborator — findings are suggestions that specialists evaluate in context.

This aligns with the existing pattern (handlers own their domain, orchestrators dispatch) and reduces the HITL burden from "review every recommendation" to "review only when the specialist can't decide."

## Deferred Until

This is a ~1 week project (agent definitions, prompt updates, routing changes, testing). Deferred while we ship the smaller pipeline fixes. Revisit when:

- HITL queue becomes a bottleneck
- A site needs `approval_mode: review` for compliance reasons
- Content rewrite failures from recommendation-as-bug cases become frequent enough to justify

## Interim Workaround

For now, recommendations that cause validation failures can be closed manually:

```sql
UPDATE site_work_items
SET status = 'wont_fix',
    error = 'Audit recommendation, not a bug — see design note on recommendation specialists'
WHERE item_type = 'content_rewrite'
  AND status = 'needs_human_review'
  AND id = '<id>';
```
