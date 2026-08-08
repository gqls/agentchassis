# HANDOFF — 117 chrome staleness reference — cold start for a fresh chat

**Rewritten 2026-08-08 evening.** State: **FIX LIVE on chassis v1.0.1266,
pod-verified on both replicas. Council APPROVED r1. Migration applied. What
remains is a WATCH, not a build: observe the first stamp, then the baseline
wave firing once per site and going quiet.**

## Read these first, in this order

1. `bugs_open/117_HANDOFF_2026-07-27_...md` — bug + three dated contributions
   (measurement → fix built → LIVE update)
2. `PLAN_2026-08-07_chrome_staleness_reference.md` — D1–D8 (D6 render inputs,
   D7 mechanism, D8 pre-proposal live validation)
3. `NOTES_chrome_staleness_reference.md` — append-only; 4 missteps, each with
   its cheap check; the council-objection triage; the oufe finding
4. `RUNBOOK_chrome_staleness_reference.md` — **R10 step 4 is the watch recipe**
5. Register **IMP-052** (`docs026_concept_register/register/improvement-loop.md`)

## What is DONE and verified (do not redo)

- Commit `998bf4c9f` (code + register + migration file), `ecb5ce91f` +
  follow-ups (docs). Council `f62e20ae` APPROVED r1; 6 advisory objections
  triaged in NOTES ("content_hash exists" → refuted: 0/1,884 rows populated).
- Migration 334 applied + ledger-recorded BEFORE the roll
  (`schema_migrations`, `bugfix_117_lane_hand_applied`).
- v1.0.1266 binary carries the change — both replicas, one exec each:
  `render_inputs` 6 / `stale_sc_` 0 / `stale_render` 0 /
  `chrome_dead_control` 5 (positive / two negatives / pipeline control).
- Pre-proposal live validation: fingerprint deterministic, healthy variance,
  in-transaction stamp→check→converge round trip (NOTES + scratchpad
  `fingerprint_candidate.sql`, reproduced in RUNBOOK R9).

## THE WATCH — what the next session actually does

As of 2026-08-08 ~16:30Z: **57/57 `site_components` rows unstamped; 0
`stale_chrome` items.** Nothing has rendered chrome and no discovery pass has
run post-roll. Two observations close this lane's loop:

1. **First stamp.** Any chrome render (site build, nav_drift, drain of the old
   `stale_sc_*` items below) should write `render_inputs`:
   `SELECT count(*) FILTER (WHERE render_inputs IS NOT NULL) FROM site_components;`
   If a post-roll chrome render leaves its row unstamped → the UPDATE is not
   reaching the new code — check which POD ran it (spawned agents pin their
   image at spawn; pre-roll spawns on v1.0.1264 still run old code).
2. **The wave.** The first discovery pass per site fires ONE `stale_chrome`
   item (~19 total across the fleet, priority 8, handler rerender-pages), each
   rebuild restamps, items do NOT re-fire. Loud-forever and silent-forever are
   the two opposite failure modes — check both:
   `SELECT s.domain, swi.status, swi.created_at FROM site_work_items swi JOIN
   sites s ON s.id=swi.site_id WHERE swi.item_key='stale_chrome' ORDER BY 3;`
   NOTE: nothing fires discovery on a clock (improvement-sweep has been
   disabled for months — IMP-016/IMP-050); a pass runs when a build pipeline
   or a manual 294 trigger does. Do not read "no items yet" as breakage
   without first checking whether ANY discovery ran.
3. When both are observed: record the close in the bug file (**it STAYS in
   `bugs_open/`** — owner ruling 2026-08-06), update IMP-052's status line,
   and update this handoff.

Old-predicate queue debt, deliberate: 6 `detected` `stale_sc_*` items
(dartsonline ×3, leopardessconsulting ×3) + 1 `deferred` (idea.uk footer).
When dispatch drains them, those rebuilds STAMP those sites — free
convergence. Leave them.

## The oufe finding — handed over, but check it moved

While canary-hunting, found **oufe.com's footer honesty note (mig 268,
stored-artefact-only) has been GONE from stored chrome AND the wire since a
chrome re-render 2026-07-31 19:21** — trigger unidentified (the 19:18
`stale_sc_footer` item is idea.uk's; attributing it to oufe was this lane's
misstep 4, logged in WRONG_CALLS). Contribution appended to
`docs024_key_docs_latest/oufe/HANDOFF_2026-07-30_continue_here.md`; the
workstreams memory line updated. oufe's slots are UNLOCKED, so the baseline
wave will rebuild its chrome (harmless now — the note is already gone — but a
bare re-restore without a 069 lock or templating will be lost again). If the
oufe lane has not acted by the time the wave reaches it, nothing breaks
further; the note's restoration is their call.

## Standing constraints (do not rediscover)

- Deploys are owner-run whole-fleet (`make release redeploy-agents`); a lane
  never builds/rolls a single service at its own tag.
- The fingerprint fragment correlates on an `sc` alias; both callers and four
  mutation-proven pinning tests guard the seam
  (`chrome_render_inputs_contract_test.go`).
- NULL stamp = stale is DESIGNED (backfilling would have declared oufe's
  stale footer fresh). The wave is not a bug; do not let anyone file it.
- Locked slots (6/57) are invisible to this check on purpose — 069 owns them.
- Hand-patched chrome fleet-wide is one legitimate rebuild from reset: the
  durable protections are 069 locks or template/data carriage, never a stored
  artefact alone.

## Ownership

This lane owns 117. Re-run RUNBOOK R8 (who-owns + live-transcript symbol
grep) before resuming — both checks lag.

---

# CONTRIBUTION 2026-08-08 ~17:40Z, from the oufe rerender-safety lane — WATCH OBSERVATION 1 IS IN: the first post-roll chrome render stamped, all three slots

Your oufe finding was acted on today (owner directive: "any oufe rerender
should not break the site"), and the fix's propagation gave your watch its
first observation for free:

- **First stamp: POSITIVE.** A `needs_rerender` item (key
  `oufe-mig339-chrome-carriers-2026-08-08`, `refresh_site_components: true`)
  re-rendered oufe.com's three chrome slots at **2026-08-08 17:36:08–09Z** —
  the fleet's first post-roll chrome render — and **all three rows came back
  `render_inputs IS NOT NULL`**. The UPDATE reaches the new code; the writer
  works. (Run by the current v1.0.1266 pods, both started 16:26–16:27Z, so no
  pre-roll-spawn caveat applies.) Fleet count is now 3/57 stamped.
- **The oufe rebuild hazard your handoff flagged is CLOSED, the durable way.**
  The honesty note and the header CTA rewrite (a second artefact-only patch
  your 07-31 finding implied and we confirmed live — "Get Started" was back on
  the wire) are now carried as gated config on the shared templates
  (STY-052/053, `sql_for_agents/339`, commit `efc879d92`): `chrome.footer_note`
  + `chrome.header_cta_url/_label` in oufe's `site_specs`. A
  `refresh_site_components` run now REPRODUCES both — verified in the stored
  artefact and on the wire (index.html: note 1, "Read the cases" 1,
  "Get Started" 0). So when your baseline wave reaches oufe, it is licensed to
  rebuild and will change nothing a visitor sees.
- **One knock-on for your wave arithmetic:** mig 339 edited both shared
  templates (footer-theme-chrome, header-theme-chrome), which changes those
  slots' render inputs fleet-wide. All rows were unstamped anyway, so the wave
  you already expected covers it — but if any site had been stamped between
  your handoff and now, this edit would legitimately re-fire it once. Nothing
  fired today: 0 `stale_chrome` items exist as of this writing.
