# PLAN — bugfix 291: tool-auditor files dispatchable work at `hitl-review`, an agent that has never existed

Created 2026-08-17. Owning thread: this lane (session "bugfix 291"). Bug file:
`bugs_open/291_HANDOFF_2026-08-17_tool_auditor_files_dispatchable_work_at_hitl_review_an_agent_that_has_never_existed.md`.
090 diagnosis run: `RUN_CORRELATION_ID=3555b514-ca8f-4f31-9f55-e105ce73e961` (filed 2026-08-17 ~12:03Z).

## Root cause (first-hand, pre-090)

- tool-auditor's `create_review_item` step (`create_work_item` action) files
  `item_type='needs_human_review'`, `handler_agent='hitl-review'`, **no `status` key**.
- `create_work_item_action.go:208-211` defaults status to `'triaged'` → born dispatchable.
- `hitl-review` has never existed. It was documented 2026-04-19 as "a convention, not a
  registered agent"; the roadmap row "build hitl-review agent" was never done
  (`old_design_and_styling/HANDOFF_2026-04-19_…update4(3).md:136`).
- Claim's handler-not-registered branch (`claim_work_item_action.go:182-215`) flips each
  row to `blocked`; `feasibility-recheck` can never promote (its predicate requires the
  handler to exist). Each row holds a `(site_id,item_key)` dedup slot `audit_review_<page_id>`
  → the auditor's LATER findings for the same page are silently dropped.

## Decisions and their reasons

1. **Candidate 1 of the bug file, decided: `hitl-review` was the never-built half of the
   HITL parking idiom, not a real agent.** The canonical spelling is `handler_agent=''` +
   `status='needs_human_review'` (migration 217; fleet census 544 `''` vs 22 `'human-review'`;
   the council-driven correction at `refresh_evidence_fact_drift.go:694-706`). The designed
   consumer already exists: admin confirm (`confirm_work_item_handler.go:77,95-117`) turns a
   confirmed `spec.check='tool_auditor'` review item into an `improve_tool` follow-up.
2. **Three phases, forced by a live-binary constraint.** The LIVE binary hard-errors on an
   empty `handler_agent` in `create_work_item` config (`create_work_item_action.go:184-187`).
   Flipping the handler to `''` by migration NOW would, under the loop's
   `continue_on_error:true`, silently lose every finding (no row at all — worse than the bug).
   So: **Phase 1** (live today): migration adds `status: needs_human_review` ONLY + repair
   migration parks the blocked rows. **Phase 2** (rides next roll): write-door guard + relaxed
   validation + resolve_composition fix. **Phase 3** (staged, post-roll-verification): flip
   tool-auditor's handler to `''`.
3. **Framework fix = demote at the shared write door, never refuse.** In `writeWorkItem`:
   born-dispatchable + non-empty handler + not registered (probe reuses
   `workItemHandlerRegisteredSQL`) → born `blocked` with the same error text claim writes.
   Refusal would lose the finding to a pod log (continue_on_error, discovery sweeps
   log-and-continue); a blocked row is durable, visible, and feasibility-recheck (600s)
   promotes it the moment the handler is registered — self-healing the image-before-seeds
   ordering this platform lives with. Claim's branch stays as the universal backstop
   (41 raw `INSERT INTO site_work_items` sites bypass the door).
4. **The guard's trigger set is exactly CHECK 443's statuses (triaged/approved/claimed),
   nothing wider.** Widening to parked/deferred statuses would be WRONG, not just redundant:
   `capability_gap` rows deliberately name unbuilt builders at `deferred`
   (`load_work_item_actions.go:278-292`), and five producers park `needs_human_review` at the
   unregistered pseudo-handler `'human-review'`. Policing those would demote them all to
   blocked — recreating bug 284 inside 291's fix.
5. **Scope kept tight**: the five `'human-review'` Go producers (`evidence_citations.go:438`,
   `refresh_evidence_base_action.go:1191,:1266`, `directory_claims.go:432,:729`) are
   inert-but-wrong the same way `resolve_composition_layout` was; they are RECORDED as a
   residual (register entry + bug file), not cleaned here. Only the `hitl-review` producers
   change in this lane.
6. **279's candidate-3 ratchet is discharged, not built**: the write-door probe against
   `agent_definitions` IS the write-time registry check the 279 ratchet header said did not
   exist (`work_item_type_minting_ratchet_test.go:36-43`), for the handler half.

## Phasing

- Phase 0: 090 run (done, awaiting verdict); standing five (this dir); fresh census.
- Phase 1: migrations 447 (config, status-only) + 448 (repair, stamped `result.repair_291`),
  hand-applied per file + `--record-only`; snapshot verified in `agent_definitions_backup`.
- Phase 2: Go commit (guard + relaxed validation + resolve_composition + tests + register
  entry, one pathspec commit), council gate before/alongside.
- Phase 3: `STAGED_tool_auditor_review_handler_to_empty.sql` in this dir; moves into
  `sql_for_agents/` at the next free number ONLY after the rolled binary is
  provenance-verified (`git merge-base --is-ancestor <guard-commit> <stamp>`).

## Known residuals (stated, not hidden)

- Admin Retry (`site_admin_handlers.go:852,877`) promotes parked rows to `triaged` without
  touching handler: until Phase 3, a retried review item recycles to `blocked` (non-corrupting
  loop); after Phase 3, CHECK 443 refuses the Retry (HTTP 500, integrity kept). The designed
  path is the confirm endpoint. Out of scope here.
- The 41 raw-INSERT sites bypass the door guard; claim remains their backstop.
- The five `'human-review'` producers (above).
- bugs_open/033: the human-review queue still has no working surface; parked review items
  join the same queue (that is where they were always meant to go).
