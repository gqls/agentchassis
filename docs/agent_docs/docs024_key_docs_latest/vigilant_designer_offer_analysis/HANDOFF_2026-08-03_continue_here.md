# HANDOFF — vigilant designer + offer analyser (2026-08-03, session 2)

**COLD-START = this file + PLAN_2026-08-02 (programme + owner decisions). NOTES has the
missteps; the 08-03 entries are load-bearing. This supersedes HANDOFF_2026-08-02.**

## State at handoff

**ALL PHASE 0 + PHASE 1 CODE IS COUNCIL-APPROVED AND LIVE AT THE POD:**
- A0.3 `write_render_audit_findings` **r2 APPROVED** (e49f5935, 2026-08-03) — live in
  chassis ≥ v1.0.1244, pod-verified both replicas (positives + the r2-only phrase +
  negative "on the next render audit" = 0).
- A1.1 `capture_renders` APPROVED r1 — live in browser-runner v1.0.1241 (my build),
  verified with `grep -ac` on the binary (**the image has NO `strings` binary — clean
  zeros from `strings | grep -c` are a lie**; WRONG_CALLS 08-03, LANDMINES 07-30 entry).
- A1.2 `VisionCapable`/`execute_vision_prompt` **r2 APPROVED** (fee9d810) — live in
  chassis. The r1 gating objection was a STALE MDL-038 register entry; corrected, and a
  new LANDMINE (register status lines outlive their truth) is synced to doc_notes.
- **A0.3b LIVE: migration 301** applied + recorded (render-audit-agent:
  site → audit → write_findings → complete; step-level error edge; verify DO/RAISE
  passed; live-probed). 299/300 were taken by other lanes mid-session — my file is 301.

**A0.4 (drain proof) — IN FLIGHT, next action below:**
- Specimen chosen: **relojistas.com** (rebuild lane COMPLETE 07-28, nothing in-flight).
- Its detected queue reviewed row-by-row: **5 provably-stale rows CANCELLED with
  evidence in `result`** (2 rerender rows: live site now serves header/footer 5/5 pages,
  index CTAs coherent; 2 phantom links + 1 empty href: rendered_html greps = 0). ONE
  detected row deliberately left (`empty_section` noticias news-listing) for the sweep's
  own check to adjudicate.
- **`run_improvement_sweep_once.sh` now EXISTS at repo root** (committed; the 08-02
  handoff referenced it before it was written). Blast-radius header is real — read it.
- **HELD: another session is building + deploying a fresh chassis** (user notice,
  ~2026-08-03 late). A spawn within ~300s of a chassis restart is silently dropped, and
  a roll kills the orchestration. A background watcher waits for the roll; then:
  1. Re-grep the new pods (positives `write_render_audit_findings` /
     `execute_vision_prompt`; negative `"on the next render audit"` = 0) — a fresh build
     from HEAD carries this lane's code, but VERIFY, don't assume (bugs_open/153).
  2. Wait 300s+ after the newest pod start.
  3. `./run_improvement_sweep_once.sh relojistas.com` — save SWEEP_CORR.
  4. Watch: `audit_state` lands in collected_data (291 gate — audit_due should be TRUE,
     site never audited); the audit chain runs; write_findings fires (FIRST witnessed run
     — it fails loud on .response mismatch, and that loud failure would be a REAL
     finding, not a nuisance); triage promotes; one item travels
     detected→triaged→claimed→complete with a visible page change.
  5. THEN close bugs_open/171 citing the orchestration id (the 291 gate's witnessed run
     is what 171's closure has been waiting on). 016b §10 index update on close.
- Hold bugs_open/115's two rows for A3.2's acceptance run (unchanged).

## Watch-out list for the sweep run

- css-patch-agent has NEVER received a work item — if the audit files contrast items and
  the handler misbehaves, that is the FIRST exercise of that workflow, expected to be
  rough (improvement_guardian advisory on e49f5935 r2 said exactly this).
- The 090-family sweep dedup: `empty_section` may be RESOLVED by the sweep's check via
  RFC_010 retraction (positively-observed-healthy) rather than dispatched — either
  outcome is a valid drain-proof leg, name which one happened.
- `scheduled_tasks.last_triggered_at` advances on fire-and-forget — measure at
  `orchestration_states`, never the task row.
- A FAILED step can show COMPLETED with error NULL — read `__step_error`.

## After A0.4, in order (unchanged from PLAN)

A2 critic (seed `design-critique-agent`, LANDMINES: start_step never initial_step
VIZ-012; root ai_service SHADOWS step-level MDL-039; critic must read `renders_failed`
and refuse a partial sweep) + Gemini-vs-Claude trial → A3 recompose + 016/017 → A4
anti-brochure → then the offer analyser (B track).

## Who owns what nearby (unchanged)

portfolio_positioning owns premise→writer wiring; brochure_component_library owns 016's
first-user relationship; bugfix_149 owns checker-layer plumbing; 151/156 own
content-duplication detection (inert, do not route work at it). This lane owns: the
drain, the critic, the recompose handler, anti-brochure compose-time work, and (after A)
the offer analyser.

> **Cross-lane pointer added 2026-08-04 by `staged_component_build` (not this lane):**
> when you reach A2's finding contract or B4's vocabulary, read
> `CONTRIB_2026-08-04_your_decisions_could_be_fence_checkable_when_your_vocabulary_settles.md`
> in this directory first — a one-field data-shape suggestion (populate `acceptance_test`
> in the browser-runner's check vocabulary for page-shaped findings), parked now because
> it is cheap at vocabulary-authoring time and a retrofit after. Nothing asked until then.
