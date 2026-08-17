# NOTES — bugfix 287 (spawn_record) — append-only, newest at the bottom

## 2026-08-17 — session 1 (lane opened)

**Ownership/validity checks:**
- `who-owns.py 287` → filing lane = mortgagecalculator adoption (handed off); RFC_029 lane
  contributed §10 and disclaimed ownership. Transcript grep across live `.jsonl` sessions:
  hits are memory-index lines / citations; the one session touching
  `loop_expansion_handler.go` is the 289 lane (committed `509e01e6a`). No session editing
  `coordinator.go`/`load_work_item_actions.go` for this bug. Target files clean in `git status`.
- Validity (live DB, ~12:30Z): resolver rows for `build-dispatch-loop` `field=result` at
  10:00Z (2), 08:00Z (6); item census 08:00Z: 4 spawn-record vs 1 own-envelope completions.
  Bug still firing.

**First-hand reads at HEAD (declared substitution per the 2026-07-31 ruling — the bug's own
090 ran against a tree 881 commits stale, §6):**
- `ExtractActionInputs` (`action_inputs.go:646-1160`): Strategy 0 = dotted-only; whole-tree
  search runs for ALL non-strict fields and its winner blocks Strategy 4 ("already
  resolved"); the OBSERVATION block (~:794) documents exactly the dotless-mapping bypass and
  persists `RESOLVER_MAPPING_BYPASSED` rows.
- **`setLoopVariable` runs before EVERY local action** — single call site
  `coordinator.go:1355` (`executeLocalAction` step 5a), calling `propagateIterationOutputs`
  with the CURRENT step's `loop_iteration`. This CORRECTS 287 §9 fact 2 ("once, at the START
  of each iteration") and the §10 warning that `!` today would hard-fail (base key is
  present at mark_complete — the 279 bypass rows require mapped-key-EXISTS to fire, so the
  instrument itself proves it).
- `UnknownConfigKeys` strips `!` (`action_inputs.go:284`) — `result!` is recognised when
  `result` is; `CheckConfig: true` on `CompleteWorkItemInputSpec` is safe.
- `diagnose-dispatch-loop` / `report-dispatch-loop`: NOT loops — top-level steps. Same
  dotless `result: handler_result` on their mark_complete. 3/3 recent diagnose COMPLETED
  orchestrations hold the envelope at exactly `collected_data.handler_result`;
  report-dispatch-loop had zero recent rows (demand caveat).

**Census re-run** (query in RUNBOOK): 22 sites / 7 agents (was 15/6 on 08-16 — grew back as
§9b predicted). All read-references; conditions are the only expression-shaped values.

**Adversarial review (fork, ~21 tool calls, read-only) — 2 plan gaps found, both adopted:**
- G1: a strict miss on build-dispatch-loop's mark_complete (no error_step,
  `continue_on_error: false`) fails the WHOLE loop via `routeToErrorStepOrFail` →
  `failWorkflow` and strands the item in `claimed` (claim predicate only reclaims
  `triaged`/`approved`). Remedy: migrations also add `error_step: mark_failed`.
- G2: migration must gate on the binary carrying the `!` parser (v1.0.1303+); precedent/
  template = `sql_for_agents/417_image_build_handler_asset_id_goes_strict_HOLD.sql`.
- Also: full walk of all 80 live loop substeps — zero literal/write-dest collisions for the
  generic suffixing pass; nested sibling-output strings only inside `input_fields` ARRAYS
  (field names — must not be rewritten; test pins it); `sql_for_agents` numbering: 443-447,
  449 taken; 448 free. `commit_sha` ride-along checked: 0 of 1,000+ recent complete items
  carry `result.commit_sha` — latent, noted in §11, out of scope.

**Missteps this session:** my first live-config query used the wrong JSON path
(`s.value->'sub_workflow'` instead of `s.value->'config'->'sub_workflow'`) and returned 0
rows — a well-formed answer from interrogating what doesn't exist. Caught within minutes by
`jsonb_pretty` on the actual row. It never left the session or became a recorded claim, so
not a WRONG_CALLS row; noted here as the check ("pretty-print one row before trusting a
zero-row LATERAL walk").
