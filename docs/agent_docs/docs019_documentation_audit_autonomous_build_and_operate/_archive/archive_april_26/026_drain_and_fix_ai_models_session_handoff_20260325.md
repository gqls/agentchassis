# Session Handoff — 2026-03-25

## What This Session Covered

Triage drain fix and model infrastructure deployment. All items from "Priority 1" through "Priority 5" of the previous handoff are now either applied or ready with code written.

---

## Applied This Session

### SQL Migrations (all applied and verified)

| Migration | What | Status |
|---|---|---|
| 083 | `snapshot_agent()`, `swap_agent_model()`, `revert_agent()` functions + `agent_snapshots` view | Applied |
| 084 | Audit prompt restructure: TOP 5 cap, `acceptance_test`, `current_value`, `max_fix_attempts` fields. Updated `visual-design-auditor`, `content-quality-auditor`, `site-review-agent` | Applied |
| 085 | `ai_endpoint_health` table (3 seed rows), `update_endpoint_health()` function, `ai_endpoint_status` view, `llm_call_log` flywheel columns (`work_item_id`, `prompt_variant`, `vertical`, `rag_context_used`), `endpoint-health-checker` agent definition, `ai-endpoint-health-check` scheduled task | Applied |
| 086 | Section locking in auditors (data-loading queries exclude `locked_at IS NOT NULL`), `get_audit_pass_count()`, `increment_audit_pass()`, `reset_audit_passes()` functions, `site_locking_progress` view | Applied |
| 087 | Improvement loop audit pass guard: `load_pass_count` → `check_audit_pass_limit` steps inserted. Sites skip all audits after 3 passes. `increment_audit_pass` after each triage. | Applied |

### Go Files Deployed (chassis rebuild)

| File | Type | What |
|---|---|---|
| `write_audit_findings_action.go` | Full replacement | Added `CurrentValue`, `AcceptanceTest`, `MaxFixAttempts` to struct + parsing + spec building + `getIntFromMap` helper |
| `ai_errors.go` | New file | `AIUnavailableError` type + `isAIUnavailable()` detection helper |
| `check_endpoint_health_action.go` | New file | Periodic health pinger for Ollama (GET /api/tags) and Claude (1-token haiku) |
| `registry_patch.go` | Patch | Registered `check_endpoint_health` in `GlobalActionRegistry` + `LocalActions` |

### Go Patches Written (for next deploy)

| File | What | Where to apply |
|---|---|---|
| `ai_actions_patch.go` | Fast-fail on `isAIUnavailable` before retry loop, reactive health table update | Modify `ai_actions.go` |
| `fail_work_item_patch.go` | Release to triaged without incrementing `attempt_count` when error is AI unavailability | Modify `FailWorkItemAction` |
| `claim_work_item_patch.go` | Check `ai_endpoint_health` before claiming + `extractAIEndpointFromHandler()` helper | Modify `claim_work_item_action.go` |

---

## Post-Deploy Verification (2026-03-25 ~18:10 UTC)

All checks passed:

- LLM call log: 9 calls, all successful, populating correctly
- Endpoint health: claude=UP, cpu-ollama=UP, gpu-ollama=DOWN (as expected)
- Flywheel columns: all 4 present on `llm_call_log`
- Swap functions: all 3 exist (`snapshot_agent`, `swap_agent_model`, `revert_agent`)
- Locking functions: all 3 exist
- Audit prompt caps: all 3 agents show `TOP 5`
- Work items: massively reduced — only 1 pending per domain vs 845 from previous session

**Known minor issue**: `agent_type` column is empty in `llm_call_log` entries. The `params.AgentType` isn't being passed through to the logging call. Low priority — fix in next chassis iteration.

---

## Ready to Run (Not Yet Executed)

### Briefing-Agent Model Swap

File: `swap_briefing_agent_to_mistral.sql`

Swaps `briefing-agent` step `infer_via_llm` from `claude-haiku-4-5` to `mistral-small3.1` on CPU Ollama. Uses `swap_agent_model()` which takes a snapshot first. Revert with `SELECT revert_agent('briefing-agent')`.

Run after confirming a test build works on the current deploy.

---

## Current Database State

### ai_endpoint_health

| Name | Status | Check Interval |
|---|---|---|
| claude | UP | 3600s (hourly) |
| cpu-ollama | UP | 60s |
| gpu-ollama | DOWN (never healthy) | 30s |

Health check agent (`endpoint-health-checker`) is registered but needs the `check_endpoint_health` action in the chassis to actually ping. Currently seeded values only.

### Improvement Loop Flow (post-087)

```
ensure_site_record → load_pass_count → check_audit_pass_limit
    ├── pass_count >= 3 → notify_scheduler_clean → complete_clean
    └── pass_count < 3  → spawn_quality_discovery → ... normal flow ...
                          → triage_findings → increment_audit_pass → check_has_findings → ...
```

### Site Locking Progress

```
gaswholesalers.com: 79 components, 0 locked, 0 audit passes → in_progress
```

Other sites either have no recent work items or are already mostly complete.

### Audit Findings Output (post-084)

Each auditor now produces UP TO 5 findings with:
- `current_value` — what's there now
- `acceptance_test` — concrete verification criterion
- `max_fix_attempts` — default 2
- `suggestion` — specific fix

The Go struct now captures these and writes them to `site_work_items.spec` JSONB.

---

## Architecture Changes

### Triage Drain Fix (complete)

The triage drain had 845 design-audit items across 4 domains in ~10 days. Five fixes applied:

1. **Finding cap at 5** (migration 084) — each auditor produces max 5 findings per pass
2. **Structured findings** (migration 084 + Go) — acceptance criteria enable cheap verification instead of full re-audit
3. **Section locking exclusion** (migration 086) — auditor data-loading queries skip `locked_at IS NOT NULL` components
4. **Audit pass cap at 3** (migrations 086 + 087) — improvement-loop checks pass count and skips after 3 passes
5. **Pass tracking** (migration 086) — `increment_audit_pass()` called after each triage

Estimated token reduction: ~65-70% on audit work items alone.

### Model Infrastructure (foundation laid)

- `ai_endpoint_health` table is the GPU scheduler — no separate mechanism needed
- `swap_agent_model()` / `revert_agent()` make model swaps safe and reversible
- `agent_snapshots` view shows all snapshots for easy operator visibility
- Back-to-triage error handling code is written (Go patches), awaiting next chassis deploy
- Health check action code is written and registered, pinging starts after deploy

---

## Key Files Produced

| File | What |
|---|---|
| `083_model_swap_and_rollback.sql` | snapshot/swap/revert functions |
| `084_audit_prompt_structured_findings_cap.sql` | Audit prompt updates (3 agents) |
| `085_ai_endpoint_health_and_flywheel.sql` | Health table + flywheel columns + scheduled task + agent def |
| `086_section_locking_and_audit_cap.sql` | Locking exclusion queries + pass tracking functions + progress view |
| `087_improvement_loop_audit_pass_guard.sql` | Pass count guard in improvement-loop |
| `write_audit_findings_action.go` | Structured findings Go code (deployed) |
| `ai_errors.go` | AIUnavailableError type (deployed) |
| `check_endpoint_health_action.go` | Health check pinger (deployed) |
| `ai_actions_patch.go` | Back-to-triage fast-fail (for next deploy) |
| `fail_work_item_patch.go` | Release on AI unavailability (for next deploy) |
| `claim_work_item_patch.go` | Health check before claim (for next deploy) |
| `registry_patch.go` | Action registration (deployed) |
| `swap_briefing_agent_to_mistral.sql` | Ready-to-run briefing agent swap |
| `post_deploy_verification.sql` | Verification checklist |

---

## Suggested Order for Next Session

1. Apply the three remaining Go patches (ai_actions, fail_work_item, claim_work_item) and rebuild chassis
2. Run the briefing-agent swap to Mistral Small 3, test with a build
3. Trigger an improvement-loop on a domain and verify: max 5 findings per auditor, pass count increments
4. Fix the empty `agent_type` in `llm_call_log` entries
5. Build the fix verification step (cheap LLM call checking acceptance_test after a fix)
6. Consider swapping site-classifier to Llama 70B on GPU (when ThunderCompute is next needed)
7. Begin wiring `work_item_id` through orchestration context to `LogLLMCall` for flywheel linkage
