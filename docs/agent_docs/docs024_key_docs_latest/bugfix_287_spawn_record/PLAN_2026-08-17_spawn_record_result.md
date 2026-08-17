# PLAN — fix `bugs_open/287` (spawn_record slug): dispatch loops complete work items with the spawn record as `result`

**Lane opened 2026-08-17.** Bug file (resolve by SLUG — the other 287 was renumbered 290 and closed):
`bugs_open/287_HANDOFF_2026-08-16_dispatch_loop_completes_items_with_the_spawn_record_not_the_handler_reply_since_the_08-15_roll.md`

## Pre-work checks (done 2026-08-17, evidence in NOTES)

- **Still valid:** spawn-record completions + `RESOLVER_%` rows for `build-dispatch-loop`
  within the last ~2–4 h (live DB).
- **Unowned:** filing lane handed it off; RFC_029 lane disclaimed it (§10); no live session
  editing the target files (transcript grep); `git status` clean on them.
- **RFC_029 Phase 2 is waiting on this fix** (their §10.6 — 86% of the conflict population).

## The mechanism (as verified at HEAD this session — sharper than the bug file)

Two independent doors, and the first is the one that fires:

1. **The resolver bypass (primary).** `mark_complete` = `complete_work_item` with
   `"result": "handler_result"` — a DOTLESS mapping. In `ExtractActionInputs`
   (`datahelpers/action_inputs.go:646`) Strategy 0 resolves only DOTTED paths, so the
   whole-tree search (Strategy 1/2) runs first, searches for the FIELD NAME `result`, finds
   `handler_spawned.result` (the spawn-ack payload) and wins; Strategy 4 — which would read
   `collectedData["handler_result"]` — is skipped as already-resolved. Measured: 176
   conflict + 279 bypass rows/day, winner always `handler_spawned.result`.
2. **The loop-expansion allow-list gap (secondary, the CLASS).**
   `prefixConfigStepReferences` (`coordinator.go:4443`) suffixes only an allow-list of config
   keys + `input_mapping`; `complete_work_item`'s `result`/`work_item_id`/`commit_sha` and
   `mark_maintenance_complete`'s `result_field` are missing, despite the comment "Any config
   key that references step outputs must be listed here". (The `dataRefKeys` loop is also
   literally duplicated.)

**CORRECTION to 287 §9 fact 2 (and RFC_029 §10.6a's premise), evidence-cited:**
`setLoopVariable` → `propagateIterationOutputs` runs **before every substep**
(`executeLocalAction` step 5a, `coordinator.go:1355`), not once at iteration start. So at
`mark_complete` the base `handler_result` holds the CURRENT iteration's reply. The 279
`RESOLVER_MAPPING_BYPASSED` rows prove it (they fire only when the mapped key EXISTS and
differs). Consequences: (a) the defect is purely search-beats-mapping, not a missing key;
(b) 287 §10's warning that arming `!` before suffixing "would hard-fail every loop-dispatched
completion" is wrong at HEAD — the key is present; (c) suffixing ALONE zeroes nothing — the
search still runs first on a dotless mapping. The `!` is the closer; suffixing is the
robustness prerequisite that makes strict resolution read the reply's own suffixed key.

**diagnose-dispatch-loop and report-dispatch-loop are NOT loops** (top-level steps, no
sub_workflow) and have the identical defect (287 §6a's own 090 item was an instance) —
independent proof the primary door is the resolver bypass.

## The fix (two halves, both framework-preferring)

### Half 1 — Go: stop enumerating in `prefixConfigStepReferences`

- Keep `stepRefKeys` pass + existing `dataRefKeys`/`input_mapping` handling; delete the
  duplicated second `dataRefKeys` loop.
- Add a generic pass: any OTHER top-level string config value that is *reference-shaped*
  (`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z0-9_]+)*$`) goes through the existing
  `prefixDataReference` (rewrites only when the first segment is a sibling `output_field`).
  Covers `!`-suffixed keys automatically (keyed on the value). Conditions are excluded by
  the shape gate (spaces/operators) — deliberate: one census condition is an `OR` with two
  references; whole-string prefixing would half-rewrite it.
- Census (2026-08-17 re-run, live): 22 sites / 7 agents, all read-references, zero
  literals/write-destinations; nested sibling-output strings appear only inside
  `input_fields` ARRAYS (Strategy-1 field names — must NOT be rewritten; test pins this);
  zero nested loops; zero `!`/`?` keys outside input_mapping. Effect today ≈ path change,
  not value change (given per-substep propagation).
- Tests: new `platform/orchestration/loop_config_reference_suffixing_test.go` driving
  `handleLoopExpansion` on a bare coordinator (pattern: `error_step_loop_expansion_test.go:22`).

### Half 2 — Config: the `!` strict marker (RFC_029/CTS-060, live in fleet since v1.0.1303)

- **Migration 448 (apply immediately):** diagnose-dispatch-loop + report-dispatch-loop
  `mark_complete.config`: `result` → `result!`, `work_item_id` → `work_item_id!`
  (`claimed.work_item_id`, dotted → Strategy 0), + add `"error_step": "mark_failed"`.
- **Migration 450 `_HOLD` (lift after Half 1 is live on agent-chassis):** same on
  `build-dispatch-loop.process_item.sub_workflow.steps.mark_complete`
  (`result!: handler_result`, `work_item_id!: current_item.id`, `error_step: mark_failed`).
  Template: `417_image_build_handler_asset_id_goes_strict_HOLD.sql` (binary merge-base gate,
  measured preamble). Held so strict resolves the suffixed `handler_result_N` directly —
  no reliance on the propagation side-channel — per the RFC_029 lane's fix-then-ratchet
  sequencing.
- `error_step: mark_failed` (adversarial-review finding G1): without it a strict miss fails
  the WHOLE loop orchestration (`routeToErrorStepOrFail` → `failWorkflow`,
  `continue_on_error: false`) and strands the item in `claimed`, which nothing in code
  reclaims. With it: one failed item, loop continues.
- `work_item_id!` closes the wrong-item-completion door (search winner
  `claim_result.work_item_id`, 453 conflict rows/day); it resolves explicitly today, so
  strict only forbids the silent fallback. Separately objectionable at council.

### Deliberately NOT doing

- No resolver precedence change (RFC_029's seam — "stop searching, not search better").
- No change to `propagateIterationOutputs` / `applyResponseToState` (RFC_012's seam).
- No rewriting of `condition` expressions.

## Verification

1. `RESOLVER_%` rows with `context->>'field'='result'` → 0 per agent after its migration,
   WHILE the agent has traffic (demand control: orchestration_states count). NOTE: Half 1
   alone zeroes nothing; `work_item_id`/`current_page` conflict rows are NOT expected to
   zero (search noise for Strategy-0-resolved fields / another extraction — RFC_029 triage).
2. Item census: spawn-record-shaped `complete` rows → 0 while own-envelope > 0 (287 §8).
3. One item end-to-end at the artefact.
4. Unit tests + build against `git archive HEAD`.

## Historical repair (measured, after live)

Spawn-record results embed corr id + iteration in `topics` → join back to the parent's
`collected_data.process_item_iter_N_call_handler.response` where the parent survives.
Measure the joinable population first; migration only if meaningful; preserve the original
under a `_replaced_spawn_record` key; report counts to the owner before applying.

## Process

- Council gate submission (both halves, one coherent task) via 097; commit with
  `Council-Submitted: <corr>` trailer. Register the mechanism change (new WFA entry) in the
  SAME commit as the code (ordering-exemption condition 2). 287 stays OPEN until the
  build-dispatch-loop half is live-verified (fixed AND live bar).
