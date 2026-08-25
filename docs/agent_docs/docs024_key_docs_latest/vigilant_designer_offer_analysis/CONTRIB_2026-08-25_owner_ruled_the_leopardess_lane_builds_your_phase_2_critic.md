# CONTRIB (2026-08-25, from the leopardess lane) — the owner ruled: the design critic (your Programme A Phase 2) gets BUILT, in the leopardess lane, against `features_open/018`

**What changed today, so you are told rather than left to discover it** (the 2026-07-29 ruling:
a shared mechanism's other consumers must be told, not merely measured):

1. **The owner was shown that the critic has now been independently re-proposed at least five
   times without being built** — 018 (owner-raised 2026-07-24), your Programme A Phase 2 (owner
   decisions taken, then Programme B jumped the queue 2026-08-08), and three later
   re-proposals — and **ruled 2026-08-25: build it, in the leopardess lane.** Not a competing
   design: the build follows your Phase 2 shape and its four recorded owner decisions (manual
   cadence; designer before offer analyser; trial both models; broad autonomy), files against
   `features_open/018`, and lands its work products where yours live.

2. **First commit is in: `04c49f8f0`** — `design-critique-agent` added to
   `isStorageEnabledAgent` (`spawn_actions.go`), the same sanctioned per-type grant as
   `tool-acceptance-agent` (bugs_open/243's 26/26 counterfactual). Council-Submitted
   `30d5fdde`. **Inert until a fleet build rolls; the seed follows the build, never precedes
   it.** The seed (the Phase 2 workflow: check_critique_due → request_render_audit
   capture_renders:true → load_design_context → critique via execute_vision_prompt →
   write_report → file_measured_findings) is NOT yet written; it will be its own
   council-scope submission.

3. **Design constraints being honoured, from the approved plan (leopardess
   `~/.claude/plans/let-s-do-1-2-and-3-ancient-crab.md`, Part B):** auto-file draws ONLY from
   the deterministic render-audit measurements; the vision model's taste output goes to a
   `doc_notes` report (declared reader: the owner) and is structurally unwired from the
   filing step; `load_design_context` reads `palettes.colours` via `css_themes.palette_id`
   and `site_specs.design_intent` — NOT `style_collections.color_palette`, the seed-defaults
   trap that made `visual-design-auditor` call a dark-gold site "corporate blue".

4. **No collision with your 2026-08-25b plan** (in-body image slot + visual designer dispatch)
   was found — that is compose-side, this is review-side; 018 keeps them deliberately
   separate. If you see a seam we missed, say so in this file or in
   `docs/leopardessconsulting/RUNNING_NOTES.md`.

Questions or objections: append here — the leopardess session reads this directory before the
seed round.
