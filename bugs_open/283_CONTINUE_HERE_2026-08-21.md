# 283 — CONTINUE HERE (2026-08-21 night). Stock CONVERTED, flow GUARDED (both guards LIVE on v1.0.1322), rebuilds 6/8 done, sweep self-healing. What remains is ONE untangle (owner call), the LMC pair (veto window), and close-out verification.

**Supersedes `283_CONTINUE_HERE_2026-08-20.md`.** Session record:
`docs/agent_docs/docs024_key_docs_latest/bugfix_283_component_instance_scope/NOTES_component_instance_scope.md`
(sessions 7–8). Council: stock programme `07635a2f` APPROVED (r9); flow half `6acf8e4e`
APPROVED (r2) — both verdicts READ, trailers legitimate.

## What is DONE and LIVE (verify nothing here, it is all evidence-backed in NOTES)

- **Stock**: 24 judged conversions live; oracle PASS 170 = pre-conversion baseline; repair
  batch done; fleet serving-check clean except the shared-row pair below.
- **Flow**: generator prompt rules (mig 520); birth guard + fork guard in
  `create_tool_component`/`deploy_tool_to_site` — **ARMED AND LIVE since v1.0.1322**
  (rev `bac189921`, digest-verified, both commits ancestor-checked). Post-roll births are
  converted-or-refused at the seam.
- **Sweep** (`instance-scope-sweep`, daily 07:40): detects, files conversions, ESCALATES
  ≥2-failed rows to full-pipeline rebuilds (14-day rate limit); parked rows never
  auto-escalate. First two waves: 25+? converted (wave 2 of 14 draining at handoff time).
- **Rebuilds (owner-ruled)**: 6 of 8 regenerated clean — `chartTitle` dup and hex-id defects
  are GONE by rebuild; finetuning.uk's broken savings page fixed.

## What REMAINS, in order

1. **Wave-2 drain + live verification** (mechanical): when the 14 conversions + their
   rerenders/deliveries finish, live-fetch: finetuning.uk savings page (was broken),
   gamesdesign spawn-rate + gaswholesalers fuel-budget (rebuilt+converted), one new arrival.
   Binding check = prefixed ids, 0 tokens, 0 bare lookups.
2. **THE OWNER CALL — the shared row `795c34e6`** (automation-savings on
   ai-agent-orchestration.com + fundamentallyai.com, BOTH STILL SERVING THE BROKEN TOOL).
   Its rebuild is double-refused: the visible-text shrink floor (fresh gen kept 46/48% vs
   floor 50%) and, behind that, the cross-site regen refusal (a site-scoped regen must not
   rewrite a row another site serves). The fork route is also (correctly) blocked: the armed
   fork guard refuses forking a composition-broken template. Framework-honest options:
   a) **Retire the shared row + per-site fresh births** (recommended): deactivate 795c34e6
      (curated SQL + snapshot, precedent SEED_*), then re-seed the two `add_tool` items
      (WITHOUT replace_existing — no incumbent then exists) with `adopt_existing_page: true`
      on the save step (the bugs_open/286 ported-tool route; STEP-level key — set, run the
      two items, unset, or accept it standing). Also enrich spec.description to ask for the
      incumbent's explanatory prose (the floor's 1,114-char bar falls away with no incumbent,
      but the prose SHOULD survive — say so in the spec).
   b) Interim snapshot rollback of the row (pages work in an hour, unconverted), untangle at
      leisure — the owner previously chose rebuild over rollback; re-offer only because two
      pages stay broken meanwhile.
3. **LMC pair** (application-tracker + consolidation): veto window in their dir expires
   ~2026-08-22 mid-day (CONTRIB_2026-08-21_from_283_lane_owner_ruled_rebuild...). Silence ⇒
   seed the two rebuild items (same shape as the 8, `283-owner-rebuilds`). ⚠ consolidation's
   oracle block (oracle.py:489) drives `addDebtRow` via inline onclick — the rebuild removes
   inline handlers, so the block needs a REWRITE (row-adding by its new button id), with
   PASS 170 before+after and the mutate control, same commit. application-tracker is not
   oracle-covered: instanceaudit + served-page + click-through instead. These two run
   POST-roll ⇒ they are also the birth guard's natural demand checks (result must carry
   instance_scope fields; born converted).
4. **Tomorrow's 07:40 sweep row** in doc_notes: corpus should be ~2 (the shared pair);
   filed 0 (dedup) — reading it IS the loop's steady-state check.
5. **Close-out** (after 1–3): move `bugs_open/324` → closed (all remedies live+proven);
   `bugs_open/283` → closed ONCE the shared pair serves converted working bytes (the bar is
   fixed AND live; two serving-broken pages fail it today). Cancel the automation-savings
   parked items when the untangle lands. Residuals that OUTLIVE the lane (register/RFC, not
   blockers): RFC_032 (three render context-builders disagree on ComponentID),
   generic-text-block's 13 already-colliding pages (blocks fleet-wide `enforce_instance_scope`
   on the RERENDER path — per-workflow arming remains the rule), LMC b2_verify recapture
   (theirs), deploy_tool fork guard's first natural live exercise.

## ⚠ Session-8 traps (all in NOTES/LANDMINES; the new ones)

- **`complete` ≠ succeeded on tool-generator items**: the workflow's `complete_error` path
  completes the ITEM with empty `create_result` — read the ORCHESTRATION error.
- The shrink floor refuses intentionally-terser rebuilds; its override is STEP-level (would
  run every concurrent add_tool floor-less) — enrich the spec, don't lower the guard.
- Migration numbers: FILENAME uniqueness is the directory's real invariant now (three lanes
  share 530).
- Council gate refuses `deployments/`-only submissions on SCOPE (owner ruling 2026-07-17) —
  not on merits; and the commit-msg hook rightly refuses placeholder trailers: submit first.
