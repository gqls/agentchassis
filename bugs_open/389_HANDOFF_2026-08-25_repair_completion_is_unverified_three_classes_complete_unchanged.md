# 389 — a CTA repair that changes nothing still completes green: no completion verifier, and three proven "complete and unchanged" classes

**Filed 2026-08-25 by the bugfix_308 lane, as the re-file of `bugs_closed/308`'s Phase C —
the one half of that bug's ask that was never built.** Owner chose close-and-refile over
keeping 308 open (2026-08-25). Lane dir (history, evidence, runbook):
`docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/` — read
`HANDOFF_2026-08-25_continue_here.md` first, then NOTES §§8-15.

## The defect

A `cta_links_stale` page_rerender reports `complete` whether or not any CTA moved.
`suggested_target` is written by the detector and read by nothing
(4 grep hits, all detectors, re-verified 2026-08-25). So "repaired" is asserted by the
handler, never checked against the page — the exact shape `bugs_closed/308` was filed
about, one level up.

## Three classes where `complete` provably meant "unchanged" [all MEASURED 2026-08-24/25]

1. **Uncovered component** — the finding sits in a component not in `ctaFieldNames`
   (**124** of the fleet's 135 findings on 2026-08-25: article-body 36, ported-page 31,
   info-card-grid 14, tool-cta 12, ported-prose 9, generic-text-block 8, tool-list 5, …).
   The detector still files ONE `page_rerender` per page whatever the slot, so a page with
   zero covered findings gets a no-op rerender that completes, strikes, and manufactures
   `unresolved` stock (this made the 215-row backlog 308's lane spent a day unpicking).
2. **Owned page** — handler refuses `rebuild_policy=owned`; since `333`'s door (v1.0.1335+)
   these park at `deferred` with `builder_needed`, which is visible but still not a repair.
   > **CORRECTED 2026-08-25 (by the `bugfix_333_owned_page_door` lane): they do NOT park, and the door
   > structurally cannot cover them.** The door parks only when the TARGET HANDLER declares
   > `refuse_owned_page` (mig 488); `cta_links_stale` rerenders are filed at `page-rerender`, which must
   > never declare it (per-agent/per-branch ruling, register WII-028, with the 384 lane).
   > [MEASURED 2026-08-25 ~10:50Z, live+archive] owned-page `spec.reason='cta_links_stale'` rows:
   > **0 `deferred`, ever** — 135 complete / 108 unresolved / 96 failed / 22 cancelled / 1 triaged. The
   > `save_sections` refusal also has no `wont_fix` terminal (mig 480 covers `load_page_record` only), so
   > these loop `failed`→`triaged`. Consequences for the fix candidates — including that candidate 1
   > would refuse-loop on owned pages unless they are excluded upstream — in
   > `docs/agent_docs/docs024_key_docs_latest/bugfix_308_cta_destination_provenance/CONTRIB_2026-08-25_owned_page_cta_rows_do_not_park_under_333s_door.md`.
3. **Data-less legacy component** — empty `content_data`, `rendered_html` frozen
   (ai-agent-orchestration.com `/blog` hero + call-to-action, frozen 2026-04-14): the
   recompute has nothing to write into; the rerender carries stored HTML byte-identical.

## Fix candidates, ordered by what closes the door

1. **`VerifyMisdirectedCTAResolved`** — before a `cta_links_stale` rerender may complete,
   re-run the detector's own predicate (`ctaClassifyAnchor` — shared already, the lockstep
   discipline) on the post-render page; unresolved findings → the item does NOT complete
   (refused/deferred with the residue named). Turns "complete and unchanged" into a refusal.
2. **Stop filing rerenders for pages with ZERO covered findings** — the detector knows the
   slot and can consult `ctaFieldNames`' membership (exported or mirrored per the package
   rule in check_misdirected_cta.go); route those pages' findings straight to review/park.
   Kills the no-op → two-strike → `unresolved` loop at the source.
3. **Optional widening**: `tool-cta` (12), `tool-list` (5), `case-studies-grid` (1) carry
   `cta_url`-shaped schema fields — extending `ctaFieldNames` converts ~18 findings from
   human to machine. Architecture note: ctaFieldNames is a shared seam; council-gate it.
   `article-body`/`ported-*` stay human by design (the framework writes content).
4. `RFC_047` §10 residue — page-level `offer-analyser` output + a route back from a refused
   match (the `Talk to us about your setup` → `/about.html` family, 6 wrong writes of 256).

## Do NOT

- Have the repair execute the stored `suggested_target` — a work item's spec is data
  written by an earlier binary (`bugs_closed/308` said this; it still holds).
- Re-derive the backlog premise: the two-strike arithmetic and the census method live in
  the lane NOTES §§8, 14; counts are dated, re-derive before quoting.

## Verification bar

A `cta_links_stale` item on a page whose finding cannot be repaired must FAIL to complete
(named residue), and a fresh fleet census after one full rolling-sweep cycle must show
zero items completing with their finding still present under the live predicate.

## Lead on class 3, from the 277 lane (2026-08-25) — [UNVERIFIED, not measured by either lane]

The two frozen ai-agent-orchestration.com `/blog` components (hero + call-to-action) also sit in
`bugs_closed/277`'s residual — the `no_content_data` parked set, **12 rows across four pages**
(count as of 2026-08-25, theirs). 277 §9 records a measured cause for that population: template
drift, with `component_versions` holding zero rows for the components involved, and
`cmd/content-data-recover` already refuses exactly those rows for a stated reason, gating on a
byte-identical re-render. Whether class 3 here IS that defect or merely overlaps it on one site
with a bad early build has not been measured. Whoever takes class 3: start from 277 §9 and
`cmd/content-data-recover`'s refusal reason before designing anything new.
