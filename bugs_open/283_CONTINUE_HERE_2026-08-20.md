# 283 — CONTINUE HERE (2026-08-20 night). THE PLANNED CONVERSION PROGRAMME IS COMPLETE: 24 judged conversions live (canary + 20 LMC + 3 generic), oracle PASS 170 at baseline, click-throughs pass. 6 rows parked for humans; ONE OWNER DECISION outstanding.

**Supersedes `283_CONTINUE_HERE_2026-08-19.md`.** Round 9 APPROVED (correlation `07635a2f…`,
verdict read 2026-08-19 15:53 — `Council-Reviewed: 07635a2f` is legitimate on this lane's
commits). Full session record:
`docs/agent_docs/docs024_key_docs_latest/bugfix_283_component_instance_scope/NOTES_component_instance_scope.md`
(2026-08-19/20 session-7 entries).

## State (all verified at the artefact, not the status)

- **486 applied** (judged branch live on component-template-fixer; snapshot 076455bf) and
  **487 applied** (67-item repair batch), both `--record-only`'d; renames committed `88ad91433`.
- **Repair batch drained**: 28 fixed (27 mechanical + 1 judged-LLM success) + 35 no-ops +
  4 refused→needs_human_review. `instanceaudit --bindings` on a fresh export: dangling rows =
  exactly the 4 parked. 51 serving placements, 0 bare literal lookups.
- **Canary loans-standard-calc PROVEN end-to-end**: judged write clean (audit exit 0), owned
  delivery via empty-field_updates section_edit, live page 6/6 prefixed ids / 0 tokens /
  0 bare lookups, **oracle PASS 170 / FAIL 0 / CONV 6 / N/A 0 before AND after** the selector
  move (block-scoped, commit `d5cd2dbb0`), `--mutate expectation` control failing correctly
  both sides. LMC lane told in NOTES (their phase-4 hold can lift).
- **22-calculator judged batch seeded** (`created_by='283-judged-batch'`, item_key
  `instance-scope:<8hex>`) ~17:05Z, draining under monitor.

## Do next, in order — **steps 1–4 DONE 2026-08-20 night (see NOTES session-7 entries; batch: 20 of 22 + 2 legitimate refusals; generics: 3 of 3 + click-through proven)**. What remains is step 5's LMC recapture (theirs), the owner decision below, and the follow-ons in NOTES (18 new unconverted arrivals need a birth-gate-or-sweep decision).

1. **Wait for the batch drain** (conversions + their section_edit deliveries; progress:
   `SELECT status,count(*) FROM site_work_items WHERE created_by='283-judged-batch' GROUP BY 1;`
   plus section_edit items created_by='component-template-fixer' after 17:00Z). Gate refusals
   → needs_human_review are DESIGNED, not failures.
2. **Verify at the artefact**: fresh converted-export → `go run ./cmd/instanceaudit <export>
   --bindings` (dangling must be only the 4 parked + any new refusals); live-fetch 2–3 pages
   incl. **mortgages-repayment (2 pages — deliberate look, PLAN §5.3)**; binding check = every
   getElementById literal resolves to a declared id.
3. **Oracle lockstep for the delivered tools**: move selectors `#id` → `#c-<function>-id`
   in `loanandmortgagecalculator_couk/oracle.py`, **block-scoped per tool** (other tools reuse
   `#total-interest` etc. — a file-wide replace corrupts them; see canary edit in
   `d5cd2dbb0`), then FULL oracle → PASS 170 restored, then `--mutate expectation` control.
   One commit per wave: oracle.py + NOTES verification, trailer `Council-Reviewed: 07635a2f`.
4. **Then the generic pool**: `tool-bayesian-ranking` + `tool-archetype-clash-calculator`
   (×2 rows — the "2 generic tools" drifted to 3 rows). No oracle: verify by instanceaudit on
   written rows + served-page checks + manual click-through, after the 22 have proven passes.
5. **b2_verify red-by-design**: it covers only the 7 `b2_seeds/` pages (consolidation,
   damage-checker, bridging-loan, equity-release, fee-analyser, rate-forecaster, repayment).
   On those, post-conversion red = "conversion happened". Recapture is the LMC lane's step;
   do not restate its pin (import from decompose_lmc — their poisoned-ref history).

## Owner decision outstanding (unchanged)

**3 automation-savings placements live-broken** (ai-agent-orchestration.com, finetuning.uk,
fundamentallyai.com; rows 795c34e6, c243e0e0): judged gate refused the LLM rewrite twice
(composition hazards). Options: snapshot rollback of the rows + rerender (contained, named in
324), or human-guided script fix. The other 2 parked rows (fuel-budget, loot-table) serve
WORKING pre-conversion bytes — no urgency.

## ⚠ Traps hit this session (all now in LANDMINES.md)

- **A status-only requeue of a failed work item is permanently unclaimable** when
  attempt_count ≥ max_attempts (claim_work_item_action.go:103). Reset attempt_count AND
  claimed_by with status. This stalled the canary delivery 9.5 h.
- **`kubectl exec … psql > file` truncates intermittently** — parse-validate every export
  (another lane's entry; it bit our first export same-day).
- **A config key declared on the WRONG action's spec** passes binary probes and git -S while
  a StrictConfig action refuses it at dispatch (bugs_closed/336 — fixed+re-armed same day;
  our delivery failures 07:07–07:10 were its collateral).
- 484/485-style migration number races: re-check the directory at apply time.
