# PLAN — 117: the chrome staleness reference points at the wrong table

**Started** 2026-08-07. **Lane owner:** this thread (session `521bfaa9`).
**Bug:** `bugs_open/117_HANDOFF_2026-07-27_site_chrome_is_a_stored_artefact_no_page_rerender_regenerates.md`

## Why this bug, and why now

Picked from `bugs_open/` as the next case with **no active thread and no owning
workstream**. Checked three ways on 2026-08-07:

- `scripts/who-owns.py 117` → no owning workstream; last commit touching the bug
  file was `db14421e7`, 2026-07-27.
- Grep of all 27 session transcripts modified in the previous 5 hours: only this
  session mentions `bugs_open/117`. Four other sessions mention `site_components`
  incidentally; **none** mentions `check_integrity`, `StaleSiteComponents`,
  `stale_site_components`, `deactivated_site_components` or
  `render_site_components_action`.
- Re-checked 30 minutes later before starting work — same answer.

## What the bug says, and what I found instead

117 as filed is a **coupling gap**: chrome is pre-rendered into
`site_components.rendered_html` and served verbatim, and nothing causes a chrome
rebuild when the thing it renders from changes. That framing is still correct.

Its fix candidate 2 ("stamp provenance and detect drift") is marked
`[UNMEASURED] fleet-wide — run it before designing anything`. I ran it, and the
answer changed the shape of the fix:

**A drift detector already exists, is live, is firing — and compares the wrong
two timestamps.** `StaleSiteComponentsCheck`
(`platform/orchestration/actions/discovery_checks/check_integrity.go:306-375`,
check name `stale_site_components`) compares `site_components.updated_at`
against `MAX(page_components.updated_at)` for the site, threshold 24h. Chrome is
not rendered from `page_components`. The reference point is independent of the
subject, so the check is wrong in **both** directions at once.

So this is not "build a detector that does not exist". It is "**an existing,
draining detector answers a different question from the one its name and its
work items claim**" — which is worse, because its output is trusted.

## Decisions

**D1 — the deliverable is a correct staleness REFERENCE, not a new queue.**
The detect→rebuild path already works end to end (`needs_rerender` →
`rerender-pages`, 7 complete per slot, most recent 2026-08-06). Adding a second
producer or a second item_type would hit the `UNIQUE (site_id, item_key)` dedup
trap that `check_integrity.go`'s own `deactivated_pin_` comment documents, and
would reproduce the shape of the open `bugs_open/213`. Fix the predicate; keep
the pipe.

**D2 — a wider timestamp is NOT a better signal.** [MEASURED, see NOTES]
`GREATEST(content_components.updated_at, site_nav_items.updated_at,
sites.updated_at)` marks essentially every row stale, because `sites.updated_at`
churns for reasons unrelated to chrome. Rejected before it was written up.

**D3 — timestamps cannot be the whole answer, because one live writer does not
set one.** `fixTemplateColors`
(`platform/orchestration/actions/fix_harcoded_colours_action.go:180`) does
`UPDATE content_components SET html_template = $1 WHERE id = $2` — no
`updated_at` — and its selection query (same file, ~:145-160) explicitly targets
chrome via `EXISTS (SELECT 1 FROM site_components sc WHERE sc.site_id=$1 AND
sc.component_id=cc.id)`. A chrome template edit by that writer is invisible to
every timestamp-based detector, including a corrected one.

**D4 — therefore the primary mechanism should answer "would a re-render change
anything?", not "is one timestamp older than another".** That points at a
render-input fingerprint stamped on the `site_components` row at render time and
recomputed by the check. It is immune to (a) writers that forget `updated_at`,
(b) timestamp bumps that change no output, and (c) unrelated churn.
**NOT YET DESIGNED IN DETAIL — this is where the work stopped.** See the handoff.

**D5 — this is council-gate scope, not RFC scope [ASSUMED, wants a second
opinion].** Under the owner ruling of 2026-07-29 (§1), an RFC is needed when a
change alters what a *shared mechanism guarantees*. Correcting a check's
predicate changes which rows it emits, not the contract of the work-item seam.
But a new provenance column on `site_components` is a shared-schema addition and
may pull it over the line. Decide before submitting.

## Phasing

1. ~~Confirm the bug is still live~~ — DONE, see NOTES.
2. ~~Measure the detector against the real signal~~ — DONE, cross-tab in NOTES.
3. ~~File the 090 diagnosis~~ — DONE, ran to completion; verdict not retrievable.
4. **Design the fingerprint** — the exact render inputs per slot, where the
   fingerprint lives, what it hashes. **NOT DONE.**
5. Decide the fate of the existing `stale_site_components` check, with an
   induction test that proves it can still fire.
6. Council gate, then commit, then build.

## Corrections to this plan

None yet.
