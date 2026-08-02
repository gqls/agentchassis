# HANDOFF — vigilant designer + offer analyser (2026-08-02, session 1)

**COLD-START = this file + PLAN_2026-08-02 (the approved programme + owner decisions).**
Read NOTES for the missteps; they are load-bearing.

## State at handoff

**LIVE now (DB config, no roll needed):**
- Migration **290**: improvement-sweep → scheduler lane (`enabled` still false — G1 is the
  owner's separate go). Applied + recorded.
- Migration **291**: the convergence gate replaced the 3-pass cap (bugs_open/171 both
  halves): promotion runs on every path, skipped ≠ clean (`audit_state` in outputs),
  non-convergence files one `capability_gap` roadmap row, `site_audit_fingerprint(uuid)`
  live. Applied + recorded. **OWED: one witnessed improvement-loop run through the gate**
  (guard proved the SQL, not the engine's parse) — that is A0.4's first job; 171 closure
  waits on it.

**COMMITTED, INERT UNTIL IMAGE ROLLS (the next big step):**
- `write_render_audit_findings` action (A0.3, commits f2a222964/0b112fda4) — chassis image.
- Whole-site renders `capture_renders` (A1.1, cc328626f + b814a3d83) — browser-runner image
  AND chassis image (request_render_audit pass-through).
- Vision seam `aiservice.VisionCapable` + `execute_vision_prompt` (A1.2, 9f8f377b7) —
  chassis image.

**Council:** A1.1 **APPROVED r1** (46640fe2; both mediums answered in b814a3d83 —
`renders_failed` counter added; the runDeadline concern's premise does not hold: no
deadline exists on the render_audit path, 120s is run_checks-internal). A0.3 (e49f5935)
and A1.2 (fee9d810) verdicts **pending — read them first thing** (task #16).

## Next session, in order

1. **Read the two pending verdicts**; act on any REVISE (code is already on shared HEAD).
2. **Image rolls** (commit-built, per CLAUDE.md): bump IMAGE_TAG, `make build-agent-chassis`
   + `make build-browser-runner-adapter`, push, deploy. Verify at the pod with positive AND
   negative greps: chassis `write_render_audit_findings` + `execute_vision_prompt` +
   `GenerateWithImages`; browser-runner `renderSweepKey` + `capture_renders`. Every replica.
   No dispatch within ~300s of pod (re)start.
3. **A0.3b** (config tail): add `write_findings` step to render-audit-agent between `audit`
   and `complete` (`action: write_render_audit_findings`; output_field `render_audit` is
   the audit step's — the action unwraps `.response` itself). Seed migration with VERIFY
   reading `start_step` (the 256 lesson). THEN witness one real run and confirm the
   `.response` nesting live (the action fails loud on mismatch — that is the test).
4. **A0.4** (the drain proof): hand-fire one improvement sweep on a specimen site
   (`run_improvement_sweep_once.sh`, read its blast-radius header). Watch: the 291 gate
   branch taken (`audit_state` in collected_data), one item `detected→triaged→claimed→
   complete` with a visible page change. Then CLOSE bugs_open/171 citing the orchestration
   id. Cancel provably-stale rerender rows first; hold bug 115's two rows for A3.2.
5. **A2** (the critic): seed `design-critique-agent` per PLAN §Phase 2. LANDMINES for the
   seed: `start_step` never `initial_step` (VIZ-012); **root ai_service SHADOWS step-level**
   (MDL-039) — put the trial's provider/model where it will actually apply; the critic must
   read `renders_failed` and refuse to critique a partial sweep (the 46640fe2 objection our
   own counter now answers). Extend `write_audit_findings` maps for `design_css_fix` +
   `unknown_category_policy` (Go — needs the NEXT roll; sequence it with A3.1's Go).
   Then the **Gemini-vs-Claude trial** over 2–3 sites; record the ruling in PLAN.
6. Then A3 (recompose handler + 016/017), A4 (anti-brochure), per PLAN.

## Landmines specific to what's in flight

- The dedup key for contrast items is `contrast_failure:<page-path>#<selector>` —
  prefix==item_type is an invariant (work_items_common.go), don't "improve" it.
- `undeployed_asset` items from the render tail CO-DEDUP with check_undeployed_assets
  deliberately — an open check row suppresses the render sighting; that is correct.
- render-sweep/ bucket keys have no GC (same standing gap as acceptance-evidence/).
- `page_components.content_hash` is a dead column (0/1,183) — the fingerprint hashes
  rendered_html; don't switch it back.
- `generic.requests` is NOT a dead topic (chassis main lane) — the checker-gaps NOTES
  carry the correction; don't re-import the old claim.
- Migration runner: `--apply` takes EVERY pending file and the tree carries other
  sessions' pending migrations — apply alone via `psql -f` + `--record-only` (both 290
  and 291 were done this way; RUNBOOK has the pattern).

## Who owns what nearby

Do not collide: portfolio_positioning owns premise→writer wiring (lendzy shadow build);
brochure_component_library owns 016's first-user relationship; bugfix_149 owns checker-layer
plumbing; the 151/156 lanes own content-duplication detection (deliberately inert — do not
route work at it). This lane owns: the drain, the critic, the recompose handler, the
anti-brochure compose-time work, and (after A) the offer analyser.
