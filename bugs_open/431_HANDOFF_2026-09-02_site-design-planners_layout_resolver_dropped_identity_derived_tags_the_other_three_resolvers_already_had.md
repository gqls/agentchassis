# 431 — site-design-planner's layout resolver was the one of four composition resolvers that never got the identity-derived tag fallback

**Filed 2026-09-02** by the `site_design_planner` thread (opened same day, no prior
owner — see `docs/agent_docs/docs024_key_docs_latest/site_design_planner/`).
**Status: FIXED IN CODE (`bd8e45aba`), submitted to council
(`Council-Submitted: bd469ba1-228e-443e-a04d-6a577a210e5d`), INERT UNTIL THE NEXT
CHASSIS ROLL.** Per CLAUDE.md's bar, stays in `/bugs_open/` until verified live.

Grepped `/bugs_open/` and `/bugs_closed/` for `resolve_composition_layout`,
`extractClassificationTags` and `needs_new_layout_candidate` before filing — zero
hits, not a duplicate. Adjacent to and cites `bugs_open/113` (same mechanism
family, different defect).

## What was found

`site-design-planner` (`agent_definitions.type='site-design-planner'`, DES-003/006
in the concept register) resolves a site's layout/typography/palette in three
sibling actions that all read classification+identity data through one shared
helper, `readClassificationFromContext` (`resolve_composition_helpers.go:113`):

- `install_site_composition_action.go:295`
- `resolve_composition_typography_action.go:121`
- `resolve_composition_pallette_action.go:153`

That helper's own doc comment states the fallback explicitly: if
`classification.category`/`classification.industry_tags` are absent, derive both
from `identity.industry` + `identity.sub_industry`, folding in
`classification.site_type` as an extra tag.

**The fourth — `resolve_composition_layout_action.go`, the action that actually
picks which of the ~18 library layouts a site gets — did not call this helper.**
It had its own private `extractClassificationTags`, which read only
`classData["category"]` and `classData["industry_tags"]` with no fallback. Any
site whose classification spec carried neither field resolved with **zero**
category and **zero** tags, which `resolveLayoutByTags` (`fork_theme_composition.go:149-152`)
short-circuits straight to `fallbackLayout` — `brochure-formal`, plus a
`needs_new_layout_candidate` HITL work item, per DES-037's honest-signal design.

## Live evidence

`[MEASURED 2026-09-02]` Four sites currently have a **current** `classification`
spec (`site_specs`, `aspect='classification'`, `is_current=true`) with no
`industry_tags` key at all:

| domain | `created_by` | shape |
|---|---|---|
| `finetuning.uk` | `domain-research-classifier` (2026-04-18) | legacy classifier output, predates the current schema |
| `leopardessconsulting.co.uk` | `domain-research-classifier` (2026-04-18) | same |
| `gaswholesalers.com` | `evaluate_news_feed` (2026-09-02 00:20) | re-stamped today by an unrelated enrichment producer that deep-merges onto whatever shape was already current — confirmed by reading `feed_news_recommendation_action.go:334` (`deepMergeNewsFeed`), it does not clobber, it just never adds what wasn't there |
| `ai-agent-orchestration.com` | `evaluate_news_feed` (2026-09-02 04:59) | same mechanism, hours later |

Query used (34 current classification rows fleet-wide checked):
```sql
SELECT s.domain, ss.created_by, ss.created_at, (ss.data ? 'industry_tags') has_tags
FROM site_specs ss JOIN sites s ON s.id=ss.site_id
WHERE ss.aspect='classification' AND ss.is_current AND NOT (ss.data ? 'industry_tags')
ORDER BY ss.created_at;
```

**`ai-agent-orchestration.com` is the concrete, reproducible case.** Its `identity`
spec (`site_specs`, present since 2026-05-01, most recently re-stamped
2026-08-24) carries real, usable data:

```
industry     = "Technology Services"
sub_industry = "AI Infrastructure & Enterprise Software Engineering"
```

— data the layout resolver never looked at. On 2026-08-12 13:50:13Z it resolved
with `site_tags: []`, `reason: "fallback — no classification tags"`, applied
`brochure-formal`, and queued `needs_new_layout_candidate`
(`7b0420b9-9e32-4067-be81-b3510f41bafc`, `needs_human_review`). **That item has
sat untouched for three weeks** — nobody had connected "the layout resolver got
nothing" to "identity data existed the whole time and a sibling function already
knew how to use it."

This is a **different mechanism** from `bugs_open/113` (the palette
merge/fall-through bug on the same site) — 113 is fixed and closed at the
artefact; this is upstream of it, in the layout pick, not the palette merge.

## Cause

Code duplication drift: `resolve_composition_layout_action.go` was written with
its own inline extraction (`extractClassificationTags`) instead of the shared
helper the other three resolvers use — no evidence either was written after the
other; they simply diverged and nobody noticed because 3 of 4 kept working.

## Fix

`bd8e45aba` (this thread):

1. `ResolveCompositionLayoutAction` now calls `readClassificationFromContext`
   directly, matching its three siblings.
2. Deleted `extractClassificationTags` (58 lines, one caller, no other
   references — confirmed by grep and a clean `go build ./...` after removal).
   Its documented `classification_source` step-config override was never wired
   into any seed or workflow config (grepped `docs/agent_docs/sql_for_agents/`
   and `platform/`); the sibling typography action carries the identical stale
   doc-comment mention while already calling the shared helper directly, so
   dropping it brings this file in line with that precedent too.
3. New test: `resolve_composition_layout_action_test.go`,
   `TestResolveCompositionLayout_DerivesTagsFromIdentityWhenClassificationIsBare`
   — sqlmock, reproduces `ai-agent-orchestration.com`'s exact spec shapes
   (classification with `site_type` only, identity with real `industry`/
   `sub_industry`), asserts `site_tags` is non-empty and `is_fallback` is false
   against a mocked two-row `layouts` table. `go build ./...` clean;
   `go test ./platform/orchestration/actions/...` passes (one unrelated
   pre-existing failure, `TestNoNewSilentScanLoss` against another session's
   untracked `page_archetypes_resolver.go` — confirmed via `git status` as not
   mine and unrelated to this change).

**Council submitted, not yet reviewed**: `bd469ba1-228e-443e-a04d-6a577a210e5d`.
Full submission and rationale in the commit message; re-check verdict via
`098_REPORT` or the correlation query in `097_TRIGGER`'s own output.

⚠ **`scripts/verify-head-builds.sh ./platform/orchestration/...` FAILS against
committed HEAD right now** — not from this fix. Another concurrent session
("theme kits") is editing the same file in the working tree
(`loadSiteThemeKitDefaults`, an `apply_theme_kit` short-circuit inserted just
below this fix's call site) and has not committed yet. Notified them directly.
This fix's own code is logically independent of theirs — my regression test
passes against their uncommitted addition — but HEAD won't build until they
commit. Re-run the build check once that lands before trusting a fleet build.

## What this does NOT do

- Does **not** touch `ai-agent-orchestration.com`'s stuck `needs_new_layout_candidate`
  item. That is a per-site decision (does the library need a new layout, or is a
  re-resolve with real tags now good enough?) belonging to that site's owner/thread
  — a session named exactly `ai-agent-orchestration` appears (offline) in this
  estate's session list. Per this repo's own precedent (`bugs_open/113`: "do not
  repaint them from here"), not acted on unilaterally. Once this fix rolls, a
  re-resolve (`RUNBOOK_site_design_planner.md` §3) would let the matcher see real
  tags for the first time — whether that resolves to a real layout match or a
  second, better-informed `needs_new_layout_candidate` is an honest open question,
  not something to predict here.
- Does **not** touch `finetuning.uk`, `leopardessconsulting.co.uk`, or
  `gaswholesalers.com` — none currently have an open `needs_composition`/
  `needs_new_layout_candidate` item, so nothing is stuck for them today; they are
  named here only so the next thread that re-runs the query above doesn't
  re-derive this same list from scratch.

## Verification owed at the next roll

1. Pod-grep both replicas for a symbol this change removed, with the mirror
   control: `extractClassificationTags` should be **absent** post-roll (it's
   gone from source), `readClassificationFromContext` should show a call-count
   increase consistent with 4 callers instead of 3 (not independently countable
   from strings alone — better: watch `resolve_composition_layout`'s own log line
   `"readClassificationFromContext: resolved"` start appearing where
   `"extractClassificationTags"` used to).
2. Confirm `git merge-base --is-ancestor bd8e45aba <deployed sha>` once a build
   lands, per `service_binary_capabilities`.
3. If/when `ai-agent-orchestration.com` (or another affected site) gets a
   deliberate re-resolve, record the before/after `site_tags` and chosen layout
   here or in a follow-up file — this file predicts non-empty tags but does not
   predict which layout, and does not stand in for that measurement.
