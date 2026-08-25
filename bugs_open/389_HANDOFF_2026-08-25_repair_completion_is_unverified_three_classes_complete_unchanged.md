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
