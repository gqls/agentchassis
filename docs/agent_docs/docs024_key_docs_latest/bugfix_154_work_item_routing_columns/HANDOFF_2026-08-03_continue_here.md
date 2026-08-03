# HANDOFF 2026-08-03 — shrink-guard thread CLOSED end to end; what remains is 178's root cause

**Read this first; then `NOTES_…` (2026-08-03 entries) for evidence.** The
2026-08-02 handoff in this directory is SUPERSEDED — its items 1–3 are done and
live; its items 4–6 are carried forward below unchanged.
`SUMMARY_2026-08-03_work_item_routing_columns.md` is the read-aloud milestone
account of the whole lane.

## State: DONE, LIVE (v1.0.1238, pod-proven both replicas), and APPROVED

- `bugs_closed/154` (routing columns) + `bugs_closed/176` (selector↔loader,
  FIFO fairness) — closed earlier; unchanged.
- **Per-slot shrink guard** (`2da3e08e5`) — live since 1233, **proven by live
  induction** (refused twice, honestly FAILED, refusal item emitted+deduped,
  zero bytes written; prediction recorded before outcome — NOTES 08-03).
  Council `e64f8576`: REVISE → **APPROVED r2** (08-02 23:11Z).
- **Locked-slot exclusion** (`5f00dcba9`) — live in 1238, proven by ANCESTRY
  (no rodata of its own; the 0913d5754 literal postdates it).
- **Refusal wording** (`77b58fd4d` + advisory `0913d5754`) — Summary/Fix are
  required per-call-site params; shrink + measurement-error paths each state
  their own true sentence; completeness wording byte-identical. Council
  `98aa9103`: **APPROVED** (08-02 23:26Z). Live in 1238: pod-grep
  `shrank past the floor` = 1, `could not measure the page's existing
  sections` = 1, marker 2, control 1, nonsense-control 0, both replicas.
- Restores (`287`) artefact-proven earlier; robot-hands rendered slot carries
  both `ISO 9409-1` and the calculator anchor after a full rerender cycle.

**Nothing on this thread is owed. No post-roll checks outstanding.**

## OPEN — in priority order for whoever continues

1. **`bugs_open/178` root cause — run `090`.** Why does a link-insertion item
   regenerate the WHOLE section? Undiagnosed. Suggested symptom shape (state
   the mechanism, point at evidence, no counts): *"A tool_crosslink work item
   whose summary is 'Add … tool reference' rewrites entire page_components
   slots instead of editing them — history rows show whole-slot regeneration
   (7 paragraphs → 4, heading changed) on an item scoped to one anchor. The
   handler path and its prompt/action seam are pointed at by
   `bugs_open/178`; evidence in `page_component_history` around 2026-08-02
   10:41 for page 5a385981."* The 090 trigger self-checks the queue; grep
   `/bugs_open/` first per CLAUDE.md. Fix candidates 1 (edit-not-regenerate)
   and 3 (emit the delta) are open; the guard makes the failure LOUD now, so
   the class will surface as `save_refused_incomplete` items rather than
   silent losses — check that queue when picking this up:
   `SELECT * FROM site_work_items WHERE item_type='save_refused_incomplete'
   AND status='needs_human_review';`
2. **Sibling writers unguarded** (`ApplySectionEditAction`,
   `rebuild_blog_listing_action.go`, `apply_gap_plan_action.go`,
   `deploy_tool_action.go`) — named in 178; the council's tracked rule: a
   FOURTH floor on `save_page_sections` is the trigger for a unified
   content-loss detector as its own design, NOT another bespoke guard.
3. **`bugs_open/177`** — tool-generator raises spurious tool_content items
   (8/8 failed identically). Unstarted; 7 remaining rows sweep with its fix.
4. **relojistas' deleted DefinedTermSet slot** (history `b0e119a4`, 2,816
   chars, no slot_name recorded) — needs an owner call or the 178 lane;
   blind INSERT was refused deliberately.
5. Watch list (unchanged): `bugs_open/169` part A (spawn hang) · scheduler
   pre_query vs selector asymmetry [UNMEASURED] · loader's dependency
   subquery is SITE-SCOPED (fix = Go + both queries together).

## Landmines specific to this lane (carry-forward + new)

- **The refusal queue is live now** — `save_refused_incomplete` items are the
  guard WORKING, not a new bug. Read `spec.reason` (names slot + before/after
  sizes) before routing anything at one. A legitimate large cut is resolved by
  `section_shrink_floor` on the step config (0 disables); a measurement-error
  refusal is transient — retry, do NOT tune floors.
- 284+285 only safe TOGETHER; never revert one alone.
- A dependency releases ONLY on complete/verified — wont_fix blocks for ever.
- Dispatch quiet spells: read `collected_data->'load_items'`
  (`item_count:0` + `rows_dropped:0` = the 176 signature), never
  time-since-last-claim.
- Induction recipe (if ever needed again): inflate the STORED
  `rendered_html` of one slot (md5-guarded, back up first), queue a
  `page_rerender` item with `reason='section_data_resolved'` +
  `page_name` from `pages.name` (NOT derived from a subdirectory URL — the
  287 formula produces `blog/x` where pages.name says `x`), wait out the
  ~300s post-restart window, and mind idx_swi_dedup: an existing OPEN
  `save_refused_incomplete:<page>` item swallows the new emit.
- `orchestration_states` retention is ~24h — capture evidence WHEN it
  happens; the 08-02 induction rows are already gone, the captures live in
  NOTES and `bugs_open/178`.

## Cold-start pointers

- This file → `NOTES_…` (08-03 tail) → `SUMMARY_2026-08-03_…` for the story.
- `bugs_open/178` carries the class state + the tracked consolidation rule.
- Council trail: correlations `e64f8576` (guard, 2 rounds) and `98aa9103`
  (wording, 1 round), both APPROVED; reports in `diagnosis_artifacts`,
  commits carry `Council-Submitted:` trailers and auto-credit via 098.
