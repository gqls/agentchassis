# RUNNING NOTES — Empty sections & fix-loop completion integrity

**What this project is about (read this first).** agentchassis autonomously
builds and operates a fleet of content websites. Discovery checks find defects
(e.g. `empty_section`), create `site_work_items`, and dispatch loops send them
to handler agents. This project fixes the discovery→fix loop's integrity after
finding that robot-hands.com served product pages with every value empty while
the platform had already marked the matching work items **complete** — a fix
loop that closes without fixing. Plan: `PLAN_empty_sections_loop_integrity.md`.
Operator tasks: `RUNBOOK_empty_sections_loop_integrity.md`. Origin handoff:
`../HANDOFF_2026-07-14_empty_product_sections.md`.

Newest entries at the bottom. Update every session.

---

## Decision log

| Date | Decision | Status |
|---|---|---|
| 2026-07-14 | Verification lives in `CompleteWorkItemAction` (single choke point both dispatch loops share), via a per-item-type verifier registry in `discovery_checks` — not in workflow JSON | done |
| 2026-07-14 | Verifier reuses the SAME predicate as the detection check so detection and verification cannot drift | done |
| 2026-07-14 | Gate fails OPEN on verifier internal error (records under `result._verification`); re-detection + two-strike is the backstop. Fails CLOSED only on a positive "defect persists" verdict | done |
| 2026-07-14 | Blocked completion routes into the EXISTING attempt machinery (`attempt_count+1` → triaged/failed), claim released — no new status vocabulary | done |
| 2026-07-14 | `required_fields_missing` scope: LLM-sourced (or sourceless) value fields only; `query.*`/assets/pages sources and image fields excluded (owned elsewhere; dartsonline must not flag) | done |
| 2026-07-14 | New check emits flag-only at `needs_human_review` — no handler can honestly fix a missing data source | done |
| 2026-07-14 | Handler no-op exits flag via `update_work_item_status` (skip_if_missing) rather than `fail_work_item`, so non-work-item invocations of page-build-handler pass through unchanged | in SQL 149, unapplied |
| 2026-07-14 | robot-hands product pages: recommend remove/replace (Option B/C) — spec site, cart furniture category-wrong | open (owner) |

---

## 2026-07-14 — Session 1 (chat: "product data missing robot-hands")

### Root cause established (handoff §5.1 answered)

The false completions are three stacked holes, all proven from live config/DB:

1. **`build-dispatch-loop` completes unconditionally.** Its sub-workflow is
   `claim → spawn_handler → call_handler → mark_complete(complete_work_item)`;
   the only error routing is on `call_agent` itself failing. Any saga that
   returns success gets its item stamped complete. (Same shape in
   site-work-orchestrator's `fix_items_loop` — which additionally doesn't even
   pass `work_item_id` to the handler.)
2. **page-build-handler no-ops are success-labelled.** `check_has_ready_sections
   → else → complete_error` and `check_content_produced → else → complete_error`;
   `complete_error` is a `complete_workflow` ("Content writer skipped — page has
   no sections defined"). Only real step errors flag the item
   (`mark_item_failed` / `mark_needs_review`).
3. **gripper-detail no-ops deterministically.** `pages.sections = []`, site-plan
   `sections: null` (entity page). `load_spec_sections` → empty →
   `plan_sections` → `ready_count 0` → `complete_error`. The handler never
   looks at the slot/component in the item spec.

Smoking-gun evidence:
- The four 2026-07-10 "complete" items share an **identical 19,364-byte result
  payload**: the coordinator's response wrapper containing ONLY `site_record`
  (`complete_error` outputs `page_content` + `site_record`; `page_content` was
  never produced → the saga exited before the writer).
- Completions clocked 23:54–23:59, ~1–4 min apart — far too fast for a real
  content-writer run (1200 s timeout budget).
- No `needs_section_data` siblings were created on 07-10 → `plan_sections`
  never evaluated any fields (empty sections list, not deferred fields).
- Two-strike (`insertWorkItem`) then converted re-detections into
  `[unresolved after N attempts]` **non-dispatchable zombies** — the ~36-item
  robot-hands backlog. So the loop's failure mode was: 2 wasted no-op runs per
  item_key per week, then permanent invisible parking.

### §5.2 answered — why the source:llm fields were never filled

`_built_at: 2026-05-02`, `_sources_merged` present; components re-touched
2026-07-10 by a fleet rerender but content never regenerated. The current fill
machinery (plan_sections → writer → on_missing) is driven by **spec sections**;
this page's list is empty, so the `source: llm, required: true` fields have no
owner. Not "skipped" — orphaned. (Handoff trap confirmed: `input_schema` uses a
`fields` wrapper, not JSON-Schema `properties`.)

### Code shipped (branch 085_debug_and_feature_loops, working tree)

| Change | Files |
|---|---|
| Verifier registry | `platform/orchestration/actions/discovery_checks/verifiers.go` (new) |
| `empty_section` verifier + shared predicate + tests | `discovery_checks/check_empty_sections.go`, `check_empty_sections_test.go` (new) |
| Completion gate | `actions/complete_work_item_verification.go` (new), wired in `actions/load_work_item_actions.go` (`CompleteWorkItemAction`) |
| `update_work_item_status`: allow `needs_human_review`/`unresolved`, add `error_message` | `actions/v3_site_actions.go` |
| `required_fields_missing` discovery check + tests | `discovery_checks/check_required_fields_missing.go`, `_test.go` (new) |
| Meta-commentary guard (check 7) + tests | `actions/validate_page_content.go`, `validate_page_content_meta_test.go` (new) |
| Handler no-op flags (workflow JSON) | `docs/agent_docs/sql_for_agents/149_page_build_handler_noop_flags.sql` (new, **unapplied**) |
| Enable new check | `docs/agent_docs/sql_for_agents/150_enable_required_fields_missing_check.sql` (new, **unapplied**) |

All builds green; `go test ./platform/orchestration/actions/...` green. The two
pre-existing test failures (`platform/orchestration/orchestration_test.go`
NewSagaCoordinator signature, thunder `client_test.go` Identifier field) are
stale test files unrelated to this work.

### Deploy state (verified against the pod, 2026-07-14 ~14:00 UTC)

Owner built + deployed chassis `v1.0.1116` (pod `agent-chassis-859f7df957-kgnmg`,
started 13:48Z; all 155 agent_definitions rows → v1.0.1116). Binary grep results:

- ✅ completion gate, ✅ empty_section verifier, ✅ update_work_item_status
  extension, ✅ required_fields_missing check
- ❌ **meta-commentary guard NOT in the image** — `validate_page_content.go`
  finished 14:37:40 local; the image's `COPY . .` snapshot ran ≥14:35:31 but
  before that; image created 14:38:06. One rebuild picks it up.

### Open at end of session

1. Rebuild/redeploy chassis for the meta-commentary guard.
2. Apply SQL 149 + 150 (owner gets psql prompts).
3. Live re-drive of one gripper-detail `empty_section` item (procedure in
   RUNBOOK) — expected `needs_human_review` (149) or gate-blocked
   `triaged/failed`, never `complete`.
4. Phase 4 decision: robot-hands product pages (recommend B/C).
5. Triage the existing zombie backlog once the loop is honest.
