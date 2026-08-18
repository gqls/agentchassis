# HANDOFF — 2026-08-18 (rolling file, started 08-16), fresh chat starts here: **RFC_029 Phase 1 is COMPLETE and both its promises are discharged** — 417 PROVEN (26/26) and the window produced a real result (287 fixed, `field=result` → 0 across 193 runs). ONE thing is open: triage pass 2, and its question is written below. A rate I published was corrected today (−53%, not −73%)

> **⚠ SUPERSEDED 2026-08-18 by `HANDOFF_2026-08-18_continue_here.md`.** Read that first. Two things in this file are now WRONG: its triage method ("write the explicit mapping, then `!`") — an explicit mapping does NOT stop the search, only `!` does (RFC_029 §10.10); and its "−73% / 3.4 rows per run" figure, corrected to **−53% / 8.4** (§10.9). Everything else here is history and still accurate.

**Supersedes `HANDOFF_2026-08-15c_continue_here.md`** (whose §1 verdict summary and §4 traps
still hold). Everything 15c §2 asked for is done; this file records what changed, what was
measured, and the short list that remains. Session record with the missteps: NOTES
`## 2026-08-16`; the implementation record is RFC_029 **§10.4** (read it before touching
anything resolver-shaped — it supersedes §10.2).

## 1. What is now true (all measured this session, ~09:45–10:30Z)

- **The revision is on the shared branch: commit `53edef286`** (`Council-Submitted:
  75091072-9d65-433e-8a30-84719dc3f30f`). Both Phase 1 WARNs now ALSO write an
  `agent_error_log` row on every occurrence — `RESOLVER_CONFLICTING_CANDIDATES` /
  `RESOLVER_MAPPING_BYPASSED`, severity `warning`, action `input-resolver`, pod-level
  attribution only (`orchestration_id` empty BY DESIGN; each row's `context.identity_scope`
  says so). Mechanism: `datahelpers.SetResolverFindingRecorder` (nil = log-only), registered
  by the chassis in `agentbase.initializeComponents` as a thin wrapper over
  `orchestration.LogAgentError`. Tests green from `git archive HEAD` + the task's files; both
  call sites mutation-proven; arm budgets unmoved. **Inert until a chassis image ≥ `53edef286`
  rolls.**
- **Phase 1 (log-only) IS LIVE** — not "inert until roll" as 15b/15c said. Chassis deploy is
  `v1.0.1303`; the running binary is stamped `5e075a6f9` (probed `/proc/1/exe` with HEAD
  `bc4cd65e7` as the must-be-absent control), and `1806371ef` is its ancestor. So the WARNs
  are firing now into a ~90s log (unreadable — the seat's point), and **the `!` parser is
  live**, which is migration 417's precondition.
- **Round 2 submitted** under the same trail (`RESUBMIT_CORR=75091072-…`), JSON committed at
  `COUNCIL_SUBMISSION_2026-08-16_rfc029_phase1_revision_round2.json` (7 edits; the 417 header
  change is comment-only, so it is evidence not an edit — the server refuses comment-only
  sketches). ⚠ **It was published TWICE by mistake** (I re-ran the trigger to re-read its
  output; WRONG_CALLS entry written): orchestrations `b5678c3a-02af-4150-831a-cc4abec15a27`
  and `d1a20669-fd87-49a2-ab2c-8177fc963044`, both `fix_correlation_id=75091072-…`. Both judge
  the identical plan; **the first-completed verdict is the verdict of record**, the second is a
  free consistency check. Not cancelled — both were already into review seats.
- **In-DB evidence for seats exists now**: `doc_notes` `decision`/`RFC_029` (the ruling's key
  lines, what shipped, the codes, the window query) and `decision`/`council-submission-75091072`
  (this round's checks with the queries). Applied via `SQL_2026-08-16_doc_notes_…sql`, replay
  proven `INSERT 0 0`. `doc_notes.subject_type` has a CHECK constraint — `rfc` is not a value;
  `decision` is the fit.
- **Migration 417's two pre-apply checks are MEASURED and in its header**: 1 active
  image-build-handler row (trap N/A); `snapshot_agent(text,text)` writes
  `agent_definitions_backup` (pg_get_functiondef); ledger: this filename unclaimed. Baseline
  re-measured: **29/29 asset-deployer spawns in the last 48h carry `asset_id`**; the live
  mapping still reads `asset_id?`.
- Docs: RFC_029 §10.4; CTS-060 (status + the sink); RUNBOOK (window query + gotchas); NOTES;
  README; WRONG_CALLS. My `000_concept_index.md` CTS-060 row and my WRONG_CALLS entry both rode
  to HEAD inside other lanes' commits as stated same-file passengers (`67996ebf1`, `17afc9324`)
  — nothing lost, nothing to redo.

## 2. OWED, in order

1. ~~READ the round-2 verdict~~ **DONE 2026-08-16 10:2xZ: APPROVED** (run `b5678c3a`, 3 advisory
   objections, none high; answers measured and recorded in RFC_029 §10.4). `53edef286` is
   credited via its `Council-Submitted:` trailer — nothing to amend. The duplicate run
   (`d1a20669`) was still in seats; **read it once, as a consistency check only** — if it
   somehow returns REVISE, the first-completed verdict stands and the disagreement is worth a
   NOTES line, not a rework. Find by payload (narrowed — the bare filter seq-scans):
   ```sql
   SELECT orchestration_id, current_step, status, updated_at FROM orchestration_states
   WHERE owner_agent_type='council-gate' AND created_at > now()-interval '12 hours'
     AND collected_data->'input_data'->>'fix_correlation_id'='75091072-9d65-433e-8a30-84719dc3f30f'
   ORDER BY created_at;
   SELECT created_at, metadata->>'decision', left(body,600) FROM diagnosis_artifacts
   WHERE correlation_id='75091072-9d65-433e-8a30-84719dc3f30f' AND kind='council_report' ORDER BY created_at;
   ```
   (Round 1's report is the oldest row; the two round-2 reports follow.)
2. ~~Apply migration 417 BY HAND~~ **DONE — the owner applied it 2026-08-16 15:58:18Z on
   v1.0.1304.** `UPDATE 1`, verify DO-block passed, `COMMIT`; live mapping now carries
   `asset_id!` and no `asset_id?`; ledger row recorded (record-only). **It was run TWICE**
   (15:58:43Z): the UPDATE fence held (`UPDATE 0`) but the two statements OUTSIDE it fired
   again — a second `snapshot_agent` row carrying the POST-change config under a `pre-update`
   reason (so "the latest snapshot" is the WRONG pre-image; the true one is 15:58:18Z — new
   LANDMINE + WRONG_CALLS, and 417's header is corrected), and a duplicate `doc_notes` row
   (deleted under a guard, keeping the first). **STILL OWED: the live proof — and it is DEMAND-BOUND, not blocked.**
   `[MEASURED 2026-08-17 09:2xZ, 17.5 h after the apply]` image-build-handler runs since the
   apply: **0**. That zero is ambiguous on its own, so it carries a demand control:
   image-build-handler ran **3 times in the last 8 DAYS** — all of them 10:09–10:13Z on 08-16,
   i.e. BEFORE the apply — while the fleet around it is busy (240 build-pipeline-trigger and
   240 endpoint-health-checker runs in the last 6 h alone, build-dispatch-loop 13). So the
   agent is rare and bursty, nothing is stuck, and the proof simply has not had an occasion.
   Queued image work exists but in non-runnable statuses (`image_url_404` 40 blocked + 1
   detected, `image_source_unsatisfiable` 33 needs_human_review, `needs_imagery` 15 deferred,
   `needs_hero_image` 1 failed) — none of those drive a spawn as they stand, and the last 3
   runs had a NULL parent (dispatched, not spawned by a parent orchestration), so there is no
   parent pipeline to nudge. **Do not read a continuing 0 as either success or failure.**
   Safety sweep the same morning: image-build-handler is still the ONLY live definition
   fleet-wide carrying a `!` key (1 of 1), and zero strict/asset_id errors have appeared in
   `agent_error_log` since the apply — so the blast radius of a wrong marker stays one agent.
   The first post-apply asset-deployer child of an image-build-handler parent must carry a
   bare `asset_id`:
   ```sql
   SELECT c.created_at, c.status, c.collected_data->'input_data'->>'asset_id' AS asset_id,
          (c.collected_data->'input_data') ? 'asset_id!' AS has_suffixed_key_BAD
   FROM orchestration_states c
   WHERE c.owner_agent_type='asset-deployer' AND c.created_at > '2026-08-16 15:58:18+00'
     AND EXISTS (SELECT 1 FROM orchestration_states p
                  WHERE p.orchestration_id=c.parent_orchestration_id AND p.owner_agent_type='image-build-handler')
   ORDER BY c.created_at LIMIT 3;
   ```
   `has_suffixed_key_BAD = t` (a child receiving a key literally named `asset_id!`) means the
   binary does not parse the marker — roll back at once:
   `… -f - < docs/agent_docs/sql_for_agents/417_image_build_handler_asset_id_goes_strict_HOLD_ROLLBACK.sql`
   (a forward jsonb transform fenced on `asset_id!`; it does NOT depend on the snapshot rows).
   Two asset-deployer spawns since the roll were build-dispatch-loop children
   (`needs_brand_head_assets`, no asset in spec) — the 402 `?` doing its job, not 417's path.
3. ~~After the next chassis roll~~ **DONE. Two rolls have happened; the recorder is live and
   verified at the artefact both times** (v1.0.1304 stamp `5de6cddbe`, then **v1.0.1305 stamp
   `6a782274b`** probed 2026-08-17 with HEAD `896c5aeeb` absent as the control; `53edef286` is
   an ancestor of both). Rows are cumulative in the DB — a roll does not reset the window.
   **SECOND READ at +24 h: 1,571 rows, 7 agents, steady ~65/h — read RFC_029 §10.6.**
4. ~~THE FINDING: 86% of the population is bugs_open/287~~ **287 IS FIXED AND PROVEN, 2026-08-17
   (v1.0.1307).** That lane shipped WFA-017 (loop expansion stops enumerating) + the `!` flip on
   three dispatch agents' `mark_complete` (migs 448/452). **We graded it through our own
   instrument and filed the verification as 287 §11d** — `field=result` **805 → 0** across 11
   loop runs with the demand control holding (9.7 → 8.1 runs/h); ballots collapsed **190 → 22**
   max (WFA-017 made visible); all build-dispatch-loop rows **14.6 → 3.4 per run, −73%**.
   Our §10 warning was followed: the `!` marker went on AFTER the reference resolved, so it
   closed the field instead of hard-failing the fleet. **Do not reopen 287.**
5. **THE NEXT WORK — the residual triage, now fully enumerated and one session's job.**
   Post-fix population (rates from RFC_029 §10.7):
   - `build-dispatch-loop` `current_page` → `handler_result.retry_payload.message.body.~unwrap.current_page` (~18/h)
   - `build-dispatch-loop` `work_item_id` → `claim_result.work_item_id` (~10/h)
   - `page-content-writer` `current_page` → `~unwrap.current_page`
   - `page-build-handler` `sections` / `page_type` / `current_page` → `load_page_record.*`
   - `tool-generator` `description`/`function`/`reason`/`related_pages`, `generic` `summary` — ≤3 each
   Method per pair: **is the winner the value the step needs?** Yes → write the explicit mapping,
   then `!` (in that order — never `!` first). No → the pipeline was living on the search and
   needs the mapping more.
   **PASS 1 IS DONE (RFC_029 §10.7a) — start pass 2 from its one open question, do not redo it:**
   - the ballot-collapse claim now has its confound control (iterations/run 6.3 → 5.5, max 10
     both sides — comparable workload, so the collapse is the fix);
   - **the dangerous consumer is already safe**: live `mark_complete` is
     `{"result!": …, "work_item_id!": "current_item.id"}` — BOTH strict, so completing the wrong
     item is closed, and the residual is not urgent;
   - **but nothing in the loop's sub-workflow explains the rows**: `claim`, `mark_failed` and
     `call_handler` all map these fields explicitly (dotted `current_item.id` / `current_item.spec`),
     which should take the explicit arm — yet ~10–18/h keep arriving.
   **PASS 2's question, with the two candidate answers ranked:** (i) an action whose
   `ActionInputSpec` DECLARES `work_item_id`/`current_page` while its step config does not map it,
   so `ExtractFields` searches — *this is a grep over `RegisterActionInputSpec` and is much the
   likelier*; or (ii) `current_item` is transiently unresolvable so the explicit arm fails and the
   chain falls through. Settle (i) first. If it is inconclusive, §10.7a(d) proposes the cheap
   instrument fix (the action name IS reachable at the bypass site — add it to the row's context;
   `step_name` still is not, and is not worth threading a ctx for).
6. **Phase 2 (conflicts resolve NOTHING): gated on item 5 reaching zero or fully mapped — NOT on
   a date.** No longer "off the calendar" (§10.5's reading): the worst case, a fleet living on
   luck, is now positively excluded. It remains its own council-gated task; flip sites are marked
   in code and in `unified_extractor_search_test.go`'s header.
7. ~~417's live proof is STILL owed~~ **DONE 2026-08-18: PROVEN.** Demand arrived (26
   `image-build-handler` runs, all COMPLETED). **26/26** asset-deployer children carry a bare
   `asset_id`; **0** carry a literal `asset_id!` (the control that would mean the binary did not
   parse the marker); **0** strict errors fleet-wide. Recorded in 417's header, RFC_029 §10.8,
   and CTS-060's verify-later (now discharged). ⚠ **Do not quote "26/26" without the statuses:**
   14 of those 26 children later FAILED on `failed to get latest commit/base tree for branch
   "master"` — a git-adapter error well after input resolution, so not ours, but a **54% failure
   rate on asset deploys is somebody's bug** and nobody appears to be looking at it. Worth a
   pointer to whoever owns the deploy path.
8. ⚠ **A FIGURE I PUBLISHED WAS WRONG AND IS NOW CORRECTED — do not re-quote the old one.**
   §10.7 and `bugs_open/287` §11d said the fix gave "−73%, 3.4 rows per run". That was an
   11-run, 1.3-hour sample. Matched-window on 193 runs: **17.7 → 8.4 rows per run, −53%.**
   Corrected at RFC_029 §10.9, 287 §11e, and logged in WRONG_CALLS. **287's own claim
   (`field=result` → 0) is untouched and is now confirmed on 17× the demand.**
   Consequence for item 5: the pair COUNT is right, the VOLUME is not — over 12 h,
   `current_page` 1,124 and `work_item_id` 503 from build-dispatch-loop, plus
   `page-content-writer` `current_page` 111 and small change. It is a louder list than
   yesterday's text implies, which makes pass 2 more clearly the next work, not less.

## 3. Traps found this session (cheap, easy to lose)

- **A TRIGGER script publishes on every run — capture its output once (`| tee file`) and read
  the file.** Re-running to re-read cost a duplicate council round (WRONG_CALLS).
- **`097` refuses a comment-only sketch** ("a fix plan proposes changes, not observations") —
  a header-only migration edit goes in `grounded_in`, not `edits`.
- **`kubectl exec -i` with nothing on stdin hangs** (15c §4 still true) — pipe a heredoc.
- **A restore from the wrong backup path fails silently under `cp` if the `||` fallback never
  fired** — verify a restore by `diff`, not by exit code (my clean-room mutation harness; the
  full-suite rerun caught it).
- **The working tree still does not compile** (other lanes' WIP) — build/test from `git
  archive HEAD` + your files.

## 4. Session-start checklist

1. `git log --oneline -10`; re-read THIS file from disk.
2. §2 item 2 (apply 417) if you have the permission — otherwise tell the owner it is ready and
   why it was not applied; glance at the duplicate run's verdict (item 1).
3. Nothing else in this lane is open. RFC_029 §10.4 + CTS-060 + the RUNBOOK are current.
