# HANDOFF — 2026-08-16, fresh chat starts here: the RFC_029 revision is BUILT + APPROVED + **LIVE (v1.0.1304)** and the observation window's FIRST READ says Phase 2 must NOT flip on schedule (672 rows / 4.5 h — §10.5); migration 417 was handed to the owner to apply by hand; the per-pair triage is the next work

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
3. ~~After the next chassis roll~~ **ROLLED 2026-08-16 10:41Z — v1.0.1304, stamp `5de6cddbe`
   (probed), ancestor of `53edef286`. The window is OPEN and already answering: 672 rows in
   ~4.5 h — read RFC_029 §10.5.** The next real work on this lane is the PER-PAIR TRIAGE of
   that population (build-dispatch-loop 608 of 672; `work_item_id` conflicts with 21–93
   candidates; `current_page` winners inside `handler_result.retry_payload…`; `result` bypassed
   118× on the dotless `handler_result` reference). The mechanism reading in §10.5 is marked
   `[INFERRED]` — **file it through `090` before acting** (it is a claim about a shared loop).
   Then explicit mappings / `!` per pair. RUNBOOK has the queries.
4. **Phase 2 (conflicts resolve NOTHING) is NOT to be flipped on the calendar** — §9's
   disconfirmation clause is firing on day one. Only after 3's triage leaves the population at
   zero (or fully mapped); its own council-gated task; flip sites marked in code and in
   `unified_extractor_search_test.go`'s header.

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
