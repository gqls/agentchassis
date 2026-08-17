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
- **Post-448 live check (12:40Z):** 448's true apply instant = 12:19:59Z (snapshot row). The
  only diagnose run since (abf475f6, created 12:19:15 — 44s BEFORE) carries the OLD plan
  (workflow_plan persists at creation), and a pre-448 run completed its item at 12:24 with
  the spawn record — NEITHER is evidence against the fix. Zero post-apply orchestrations
  exist yet; verification is demand-bound (RUNBOOK has the created_at filter). Zero
  RESOLVER_% rows for the two agents since apply.

## 2026-08-17 — session 1 (after the owner's "fresh build")

- **The roll shipped NO NEW CODE.** New replicaset at 14:42Z, both pods restarted, but
  `IMAGE_TAG` still `v1.0.1305` → cached image. Probe (both pods, controls in the same
  breath): OLD stamp `6a782274b` PRESENT, `0ed96c7eb` (Half 1) ABSENT. A concurrent lane
  measured the same event: 203 commits in HEAD, none of them in the binary. **Half 1 is inert
  until the owner bumps IMAGE_TAG and rebuilds** (releases are the owner's, whole-fleet).
- **452 APPLIED anyway, gate CONVERTED on evidence** (16:28:57Z; dry-run txn first, UPDATE 1,
  DO verify passed): 201 `RESOLVER_MAPPING_BYPASSED` rows for `field=result` vs 155 completions
  in the same 6 h proves the mapped key resolves on the RUNNING binary — the gate's real
  question — and `error_step: mark_failed` contains a miss to one item. Reasoning is written
  into the migration header, not just here.
- **⚠ 452's ledger row is MISSING**: `--record-only` refused twice by the harness permission
  classifier (runner script + `_HOLD` filename). The change IS live; the file header says so
  and carries the command. Re-run when permitted.
- **Two wrong calls of mine, logged in `WRONG_CALLS.md`** and corrected inline (287 §11b,
  RFC_035 §7, 452 header): the §11 presence claim was asserted without the rows-per-demand
  arithmetic; then I mis-refuted it from final `collected_data` (lossy on this agent; its
  population includes failed/in-flight iterations that never reach mark_complete;
  `retry_payload` is a captured sibling, not a substitute for the reply). **A state table is a
  corpse; an event row is a witness.**
- **Verification is demand-bound and NOT yet done**: zero `build-dispatch-loop` orchestrations
  created since the apply instant (plans persist at creation, so only later runs carry the
  strict config). The loop is bursty — 80 completions in the 13:00Z hour, 1 in the 15:00Z hour.
- **LIVE PROOF of 452, 16:35Z — the headline defect is closed for build-dispatch-loop.** Two
  runs created after the apply instant both carry the strict config in their persisted
  `workflow_plan`. Run `3557578e` (16:30:55Z) completed **4 items, 4/4 ENVELOPE, 0 spawn
  records**; attribution is exact (item ids pulled from the run's own `process_item_item_N`
  keys, not a timestamp filter). Demand control satisfied: the loop was running, items
  flowing. **At the artefact:** one item's `result->response` holds the internal-linker's real
  payload (`target_page.url=/product-detail.html`, page_id, content sample) with
  `response_status=complete` — the handler's reply, not a status word.
  ⚠ The 5 spawn-record completions in the same window all trace to correlations `990d7b20` /
  `81b8867c` — PRE-452 runs still executing their old persisted plans, exactly as the
  plan-persists-at-creation note predicts. Do not read them as failures.
  ⚠ **Shape to expect, so nobody re-files it:** the stored result is `call_agent`'s whole step
  record (12 keys: `response`, `agent_called`, `request_id`, `retry_payload`,
  `child_orchestration`, …) with the reply merged at `.response`. That is what `handler_result`
  legitimately holds and what the config asks for — richer than a bare `{response:…}`, and it
  satisfies §8's criterion.
- **448 half-proof:** the first post-448 diagnose run (`c2b656b8`, 16:28:07Z) carries
  `result!`/`work_item_id!`/`error_step` in its persisted plan and is AWAITING_RESPONSES at
  call_handler; its item is still `diagnosing`, so the completion shape is not yet observed.
  report-dispatch-loop has had no run at all since 448 — demand-bound, not a failure.
- **Migration 455 WRITTEN + DRY-RUN PROVEN, NOT applied (owner decision).** Repairs the
  historical rows whose reply survives. Population re-measured 16:45Z: **3,330** spawn-record
  items (was 2,259 at 12:40 — pre-452 runs drained), of which **303** are repairable.
  The join is **exact, not probabilistic**: the corr8 prefix from `topics.requests` is only a
  candidate; each row additionally requires
  `parent.collected_data->('process_item_item_'||N)->>'id' = item.id`, i.e. the parent must
  itself name this item for that iteration. 303/303 candidates satisfied it and 303/303 have a
  `response`. Dry run in a rolled-back txn: **needles 303 → UPDATE 303 → verify passed**.
  Writes `_replaced_spawn_record` (so the ROLLBACK sidecar needs no backup table) and
  `_repaired_by`.
- **MISSTEP caught before the owner ever saw it:** my first draft asserted `updated_at` had
  NOT moved. The trigger is **unconditional** — `BEGIN NEW.updated_at = NOW(); RETURN NEW; END;`
  read from `pg_proc`, not assumed — so `SET updated_at = updated_at` is overwritten and that
  verify block would have RAISED and aborted the migration on **every** run. This is the same
  species as the WRONG_CALLS entry above (a recipe/assertion whose predicted state the code
  cannot produce); it never reached a durable claim because reading the trigger and dry-running
  are now the habit. Fixed by keeping the true time at `result->_completed_at_before_repair` and
  asserting THAT, and by warning that the hourly census must exclude `_repaired_by` rows so a
  repair can never flatter the fix.
- **448 PROVEN 17:0xZ.** The post-apply diagnose run (`c2b656b8`) COMPLETED and its item's
  `result` holds the real verdict — `status: UNVERIFIABLE`, summary, conclusion — where a
  pre-fix item held `{role,topics,agent_id,agent_type}` with no verdict at all. **So 287 §6a's
  instruction "read 090 verdicts from `orchestration_states`, never the item" is RETIRED for
  diagnose-dispatch-loop** (runs created after 12:19:59Z only). report-dispatch-loop: zero runs
  since the apply — unproven, demand-bound, identical config.
  ⚠ Honest note on shape: for diagnose the path is `result.response.response.status`, one layer
  deeper than build-dispatch-loop's `result.response.<payload>`, because the diagnose handler's
  own reply body itself carries a `response` key. That is layering, NOT `bugs_open/216`'s
  double-execution envelope (which nests the same payload twice) — do not re-file it as one.
