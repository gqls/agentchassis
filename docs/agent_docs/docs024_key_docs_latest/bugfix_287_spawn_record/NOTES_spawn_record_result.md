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

## 2026-08-17 — session 1 (execution)

- Half 1 committed in `0ed96c7eb` (Council-Submitted: cba35b35). Mutation proofs run and
  recorded: MUT-1 (generic pass disabled) fails the 3 suffixing assertions; MUT-2 (shape gate
  removed) fails the condition assertion — the prose assertion did NOT fire under MUT-2
  because `prefixDataReference`'s first-segment check guards it in series (a mutation that
  partially passes hit a guard in series — the condition case is the gate's unique pin).
- Full `./platform/orchestration/...` suite green; **archive build green at the new HEAD**
  (`git archive HEAD` → `go build ./...` rc=0 + new tests pass) — load-bearing today because
  the 289 and 291 lanes landed platform commits underneath us during the session.
- **Migration 448 APPLIED** ~13:15 BST: UPDATE 2, DO verify passed, doc_note inserted,
  ledger record-only with note. Pre-checks: binary probe v1.0.1305 stamp `6a782274b` present
  + HEAD-sha control absent + `53edef286` (strict parser) ancestor ✓; one active row per
  type ✓. Pre-apply txn test passed and the DO verify was INDUCED (re-adding `result`
  alongside `result!` raised as designed).
- **Same-file passenger, inbound direction:** my WFA-017 index row (uncommitted ~15 min) was
  swept into the 291 lane's `c8400e452` along with their WDS-018 row. Entry file
  (workflow-authoring.md) was untouched and landed with my code in `0ed96c7eb`, so the
  register-with-code rule holds; provenance named in my commit message. The index count
  narrative (theirs, 1,886) is consistent with HEAD after both rows — measured 1,886 rows /
  0 dups / 1,886 entries post-commit.
- **Council run** `445fbbb7` dispatched within a minute of submission (12:16:53Z — no queue
  today), seats executing at time of writing.
- **Phase D measurement (historical repair):** spawn-record `complete` rows since 08-15
  10:00Z = **2,259** (kept accumulating since the bug file's ~270). Parent row survives
  (corr8-prefix join) for 1,689, but the reply is actually still present for only **244**
  (`process_item_iter_N_call_handler.response` 238 / `handler_result_N` 240 / non-loop
  `handler_result` 4; union 244 ≈ 11%). The rest lost `collected_data` (reaping/aggregation
  blowups — see 289) or use loop names other than process_item (minor). OWNER DECISION
  QUEUED: repair the 244 by migration (counted needles, preserve the spawn record under a
  `_replaced_spawn_record` key) and accept the ~2,015 as marked-untrustworthy (the LANDMINE
  + 287 §6a reader guidance already cover them), or skip repair entirely.
- **Council verdict (12:31Z): REJECTED — guardian veto, round 1, on the Go half's SCOPE**
  (fleet-wide resolver-behaviour change riding a bug fix); architecture seat `needs_rfc`;
  guardian endorsed the config half explicitly. Actions taken: RFC_035 filed (ratify /
  narrow / revert — owner's call); WFA-017 entry + index row record the veto with the
  BLD-019-style "Live ≠ approved" line; 452_HOLD hardened (pod-probe gate, max-version pin,
  catchment note); 448 NOT edited (ledger checksum) — its prospective version-pin objection
  is noted here instead: at apply time each type had exactly one active row (measured), so
  the applied UPDATE hit the loaded rows. Seat objections that were sketch-vs-file
  misreadings (error_step nesting, is_snapshot spelling, unverified parser liveness) are
  answered in 287 §11a with pointers to the files/NOTES.
