# HANDOFF 2026-08-02 — bugfix-19 session: dispatch fixed end-to-end, content restored, guard pending roll

**Read this first; then `NOTES_…` (2026-08-02 entries) for evidence.** The
2026-07-31 handoff in this directory is SUPERSEDED (154 closed).

## State: what is DONE and LIVE
- `bugs_closed/154` — routing columns; witnessed end-to-end; CLOSED.
- `bugs_closed/176` — selector↔loader disagreement; `284`+`285` applied; fleet
  dispatching fairly (verified: claims in exact FIFO order across 6 sites).
- `286` — triage: `0733a7a4` wont_fix (spurious), `93f2a3b7` dependency cleared,
  crosslink ran (link correct on all 3 slots).
- `287` — restores applied byte-exact from history: robot-hands 4655,
  fundamentallyai 32444, vetcomparison 7206+3043. Dispositions + not-restored
  list in `bugs_open/178`'s update block.

## OPEN — in priority order for whoever continues
1. **Verify the 2 rerender items** `item_key LIKE 'restore_287:%'` (robot-hands
   how-to-specify-a-gripper, vetcomparison /about) complete AND check the
   ARTEFACT: robot-hands rendered generic-text-block must contain BOTH
   `ISO 9409-1` AND `gripper-safety-factor-calculator`; vet faq ~3× its 2197.
   ⚠ fundamentallyai `/tools/review-council-simulator.html` must NOT be
   rerendered until the restored content_data shape is checked vs the renderer
   — its live render is the only known-good copy.
2. **The shrink guard is INERT until an image roll** (commit `2da3e08e5`).
   After the next roll, pod-grep BOTH replicas in one exec:
   `strings /app/agent-chassis | grep -c "SECTION SHRINK"` (expect ≥1) +
   control `grep -c "CONTENT REGRESSION BLOCKED"` (expect ≥1). Then INDUCE:
   fire a save that shrinks a big slot >50% and watch it refuse + emit the
   refusal work item.
3. **Council verdict `e64f8576`** (shrink guard). Find by payload:
   `SELECT current_step,status FROM orchestration_states WHERE
   collected_data->'input_data'->>'fix_correlation_id'='e64f8576-e056-4c2d-87f9-4e27c65aee08';`
   REVISE → answer into the code, resubmit with RESUBMIT_CORR.
4. **`bugs_open/178` root cause** — why does a link-insertion item regenerate
   the whole section? Undiagnosed; run `090`. Candidates 1 (edit-not-regenerate)
   and 3 (emit the delta) open. Also: relojistas' deleted DefinedTermSet slot
   (snapshot `b0e119a4`, needs a slot_name decision); vetcomparison's discarded
   same-day edits (re-detect via discovery).
5. **`bugs_open/177`** — tool-generator raises spurious tool_content items
   (8/8 failed identically). Unstarted. The 7 remaining rows sweep with its fix.
6. Watch list: `bugs_open/169` part A (spawn hang) untouched · scheduler
   pre_query vs selector asymmetry (`triaged`-only + locked_at) [UNMEASURED] ·
   loader's dependency subquery is SITE-SCOPED (cross-site depends_on never
   resolves; 285 copies it deliberately — fix means Go + both queries together).

## Landmines specific to this work
- 284+285 are only safe TOGETHER (FIFO + selector/loader agreement) — never
  revert one alone.
- A dependency releases ONLY on complete/verified — wont_fix blocks for ever;
  that is why 286 cleared the array instead of faking complete.
- Dispatch "quiet spells" are never baseline: read
  `collected_data->'load_items'` (`item_count:0` + `rows_dropped:0` = the 176
  signature), not time-since-last-claim.

## UPDATE 2026-08-02 evening — grep run, guard NOT in v1.0.1229; council REVISE

- **Pod-grep both replicas of v1.0.1229**: marker `SECTION SHRINK` **0**, control
  `CONTENT REGRESSION BLOCKED` **1**, `section_shrink_floor` **0**. The image was
  built without `2da3e08e5` (which IS on HEAD — verified merge-base). IMAGE_TAG
  bumped to **v1.0.1230** (`21defe33d`); the owner runs the build/roll. After the
  next roll re-run the same three greps expecting ≥1/≥1/≥1.
- **Council `e64f8576` → REVISE** (gating: debug_historian; 1 high, 5 medium).
  Resubmit on the SAME correlation (`RESUBMIT_CORR=e64f8576-…`). Answers to embed:
  1. **HIGH — complete_error masks the refusal as green**: yes it may, exactly like
     the sibling page-total guard (016b back-catalogue), and that is WHY the guard
     emits a durable refusal WORK ITEM via emitPruneRefusalWorkItem — the
     already-council-approved mitigation from the 165 round (approved 07-31
     19:27). The refusal protects content unconditionally (nothing written);
     the work item makes it visible regardless of orchestration status. Verify
     the claim by inducing a refusal post-roll and checking BOTH the step error
     AND the emitted item.
  2. **fundamentallyai NULL shape**: cited as class motivation, not a case this
     guard closes. If the NULLing slot arrives PRESENT-with-empty → this guard
     fires (0 < floor). If ABSENT → the completeness floor's territory; note the
     residual gap: dropping 1 of 3 slots = 0.67 cohort ratio, above the 0.5
     default floor — that gap belongs to the floor's ratio, not to this guard.
  3. **Other writers unguarded (ApplySectionEditAction etc.)**: true and known —
     178 stays open for the class; this guards the chokepoint the measured
     incidents ALL went through (`save_page_sections_overwrite` history rows).
  4. **Blast radius**: measure before resubmitting — count agent_definitions whose
     config reaches save_page_sections (one query; the reviewer is right that it
     was asserted not measured).
  5. **Pod-grep step**: name the three-grep verification above in the plan.

## UPDATE 2026-08-02 ~22:15Z — GUARD LIVE on v1.0.1233; restores artefact-proven

- Pod-grep both replicas of **v1.0.1233**: `SECTION SHRINK` **2**,
  `section_shrink_floor` **1**, control **1**, negative control **0**. Guard is
  in the running binary. (1230 was skipped; the owner's roll landed as 1233.)
- Both `restore_287:%` rerenders **complete**; robot-hands rendered
  generic-text-block verified at the artefact: `ISO 9409-1` AND the
  `gripper-safety-factor-calculator` anchor both present (4,831 chars). The
  restore+merge survived a full rerender cycle.
- STILL OPEN (unchanged): induce a refusal (proves the block fires AND the
  refusal work item emits — the empirical answer to the council's HIGH
  objection), then resubmit on `RESUBMIT_CORR=e64f8576-…` with the answers
  sketched above + the blast-radius measurement.

## UPDATE 2026-08-03 ~00:15Z — refusal INDUCED and proven; round 2 resubmitted on e64f8576

- **The induction is done and the HIGH objection is answered empirically.**
  dartsonline `/blog/beginners.html`: stored article-body inflated to 15,443
  stripped chars, platform's own dispatcher fired the rerender, and the guard
  refused TWICE — both orchestrations **FAILED** (not masked) with
  `article-body 15443→4644 chars (30% kept, floor 50%)` in `error`; refusal item
  `ebc1dda8` emitted `needs_human_review` (second emit deduped, by design);
  **zero** writes (md5s, timestamps, 0 history rows, no deploy). Prediction was
  recorded in NOTES before the outcome. Everything cleaned up and verified
  (marker grep 0, byte-exact restore, backup table dropped, both items
  cancelled). Full evidence in NOTES + `scratchpad/induce178/`.
- **Locked-slot exclusion shipped** (`5f00dcba9`, Council-Submitted trailer):
  answers prior_art_librarian — locked slots are unreachable by this save
  (bugs_open/058 discards the incoming copy), so the guard no longer measures
  them (`pageComponentAgentWritableSQL`). INERT until the next roll (v1.0.1234
  predates it); the guard itself is live and proven without it.
- **Blast radius measured**: 6 live agent types invoke `save_page_sections`
  (page-build-handler, page-rerender, tool-recreation-handler top-level;
  pageflow-builder, page-rebuild, site-work-orchestrator nested). Zero configs
  set `section_shrink_floor` — the 0.5 default governs all six.
- **Round 2 submitted** (`SUBMISSION_2026-08-03_shrink_guard_round2.json`,
  `RESUBMIT_CORR=e64f8576-…`) answering all 9 objections. **NEXT: read the
  verdict** (watch armed; find by payload:
  `SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts WHERE
  correlation_id LIKE 'e64f8576%' AND kind='council_report' ORDER BY created_at;`).
  If APPROVED nothing further is owed on this thread of 178's prevention half —
  the class root-cause (item 4 of the OPEN list above) is still the 178 lane's.
- After the NEXT roll, verifying `5f00dcba9` is in the pod needs the 165
  runbook's ancestor method — the commit adds NO new rodata literal (the
  comment is stripped; the lock predicate string predates it in
  `lock_helpers.go`, and Sprintf assembles the query at runtime). So: find a
  commit made AFTER `5f00dcba9` that DID add a grep-able literal, confirm that
  literal in the pod, and since builds come from committed HEAD, ancestry
  proves mine is in. A tag postdating the commit proves nothing
  (`bugs_open/153`).

## UPDATE 2026-08-03 ~00:45Z — round 2 APPROVED; wording fix shipped under its own submission

- **Council `e64f8576` round 2: APPROVED** (23:11:38Z, 4 advisory objections,
  none high; editquality/prior_art/guidelines/provenance and all four guardians
  clean). The two guard commits carry `Council-Submitted:` and 098 credits them
  automatically. Advisory threads: (1) three bespoke floors → consolidation
  deferral now TRACKED in `bugs_open/178` (the fourth-floor trigger rule);
  (2) debug_historian: pod-grep counts should use anchored matches (`grep -cx`
  or exact-line), since the Go linker packs literals — adopt in future verifies.
- **The wrong-summary finding is FIXED and submitted on its own correlation**:
  `77b58fd4d` — Summary/Fix are now required parameters on
  `savePageSectionsRefusal` (aftermath-clause design, same cure); completeness
  wording byte-identical, shrink guard states its own case and names
  `section_shrink_floor`. Council `98aa9103-05d4-4239-b116-330167bbcaf8`
  (Council-Submitted trailer). INERT until the next roll. **NEXT SESSION: read
  that verdict** (payload query as above with the new correlation) and act on a
  REVISE; after the next roll, pod-grep the NEW literal
  `shrank past the floor` (expect ≥1, both replicas) — this commit DOES add
  rodata, unlike `5f00dcba9`.
- Remaining OPEN on this handoff's list: items 4–6 unchanged (178 root cause /
  177 tool-generator / watch list). The shrink-guard thread of item 2–3 is DONE.

## UPDATE 2026-08-03 ~01:15Z — 98aa9103 APPROVED same night; its advisory implemented, not shelved

- **Council `98aa9103` (refusal wording): APPROVED** first round, 23:26:46Z,
  2 advisory objections — the queue was empty at this hour, so the 30-min
  budget was 6 minutes for once. `77b58fd4d` auto-credits via its trailer.
- **The advisory three seats converged on is IMPLEMENTED** (`0913d5754`, no
  trailer — it executes the approved round's own instruction): the fail-closed
  measurement-error path now has `shrinkMeasurementErrorFix`, its own true
  sentence (nothing shrank; the guard could not measure; floor-tuning is not a
  remedy; `=0` is an escape hatch, not a fix). Test pins distinctness and the
  wrong-remedy exclusion. Listed un-reviewed by 098, deliberately — the message
  says why.
- tooling_provenance's ask (record the principle where the next consumer finds
  it): the principle — **a shared refusal helper's Summary/Fix are per-call-site
  REQUIRED params, never inherited** — lives in `savePageSectionsRefusal`'s doc
  comment, which IS where the next consumer looks, and the compiler enforces it
  harder than any note. Declined to hand-write a `doc_notes` row (LANDMINES
  rule: never hand-write rows the sync owns).
- **Post-roll greps for the wording commits** (both inert until the next roll):
  `shrank past the floor` ≥1 (77b58fd4d) and `could not measure the page's
  existing sections` ≥1 (0913d5754), both replicas, anchored per
  debug_historian's `grep -c` caution. In-tree build confirmation still owed by
  whoever lands the other session's `load_work_item_actions.go` edit.
- **This thread of 178's prevention is now fully closed**: guard live + proven +
  approved; wording true on all four refusal paths + approved; consolidation
  deferral tracked. What remains on 178 is the handler root cause (090) and the
  unguarded sibling writers — the class, not this chokepoint.
