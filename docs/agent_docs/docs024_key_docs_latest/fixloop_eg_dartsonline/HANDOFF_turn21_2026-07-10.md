# Handoff — Diagnosis→Fix Loop Workstream, Turn 21 (2026-07-10)

*For restarting a new chat from the end of turn 20. Complete state summary, pending actions, gotchas, and what to verify first.*

---

## Session 20 Summary (turns 18–20)

**What was built:** F2.2 revise loop — complete, tested, deployed.

**Key accomplishments:**
1. **Revise loop wired and tested** (Go + SQL + workflow v3 live)
   - `diagnose_council_decide` now counts rounds **per orchestration_id** (NOT per correlation) — critical fix that prevents prior reports from inflating round counts on fresh runs.
   - Revise decision logic: `should_revise = (decision=="revise" AND round < maxRounds)`.
   - A revise that exhausts its cap becomes `decision="exhausted"` (terminal, not silent approval).
   - Round count sourced from durable council_reports in diagnosis_artifacts, no workflow loop-state threading.

2. **Fixed design flaw caught mid-implementation:**
   - Originally scoped round count per correlation, but correlation belongs to the diagnosis and accumulates reports across proposer re-runs. Would have started the demo run at round 2 of 2, exhausting it without a repropose.
   - Fixed to count per orchestration_id (per proposer run, the right semantics).

3. **Deployed and running:**
   - v1.0.1107 shipped on commit 9c083493 with all council Go code.
   - fix-proposer v3 workflow live in the database with 10 steps: load_diagnosis → check_confirmed → load_last_bundle → propose → persist_plan → review_editquality → review_guardian → council_decide → check_revise → {repropose or complete}.
   - Snapshot taken before every agent update (`snapshot_agent` called; snapshot IDs recorded).

4. **Revise-loop demo fired:**
   - Fired against run-5 correlation `e08c5b01-01ef-42ad-80d0-b77c50ec9e84` (the CONFIRMED diagnosis from benchmark run 5).
   - Set `max_rounds=3` live (demo run is exempt from the default cap-2, so it gets genuine revise rounds on the deployed binary despite round-count ambiguity).
   - Settled 300s before firing (rebalance window risk from run-2, gotcha earned live).
   - **Status at end of turn 20:** run in flight, watcher polling for completion.

---

## Current State (as of turn 20 end)

### Code
**Tracked (in git commit 9c083493):**
- Go: tier guard, closure gate + citation-backed coverage, data_request persistence, plan validator + no-op rejection, council decision action (all in v1.0.1107).
- SQL: fix-proposer seed v3 (10 steps, repropose loop, max_rounds config).
- Runbook/Plan/Notes: updated through turn 16.

**Untracked (in working tree, need commit):**
- `platform/orchestration/actions/diagnose_council_decide_action.go` (new, 209 lines, F2.1 + F2.2)
- `platform/orchestration/actions/diagnose_council_test.go` (new, 5-case cap + decision tests)
- `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/MILESTONE_diagnosis_fix_loop_2026-07-10.md` (new, shareable narrative)
- `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/NOTES_running_fixloop(10).md` (modified, turns 16–20 added)
- `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/PLAN_fixloop_pilot.md` (modified, F2.2 status)
- `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md` (modified, CURRENT POSITION refreshed)
- `docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_fix_proposer.sql` (modified, v3 seed)
- FYI addendum to travelling_docs about max_tokens dead config.

### Live Artifacts
- `diagnosis_artifacts` (clients_db): 
  - Run 5 (e08c5b01): 2 bundles, 2 fix_plans (v1 + v2), 1 council_report (from first plan review), awaiting revise-loop results.
  - Revise-loop demo awaiting completion; will persist round 2, 3 reports + plans when it fires.
  
### Agents
- `diagnose-agent` v1.0.1107: verdict step prompt has rules 6–7 (cite every mechanism or say why not in risks; every edit changes something). Snapshot ID recorded.
- `fix-proposer` v1.0.1107: v3 workflow live, 12 steps, `max_rounds=3` set for demo. Snapshot ID recorded.

### Memory (persist across chats)
- Updated `fixloop-workstream.md` with turn-20 status, gotchas (max_tokens placement, rebalance window).

---

## Pending Actions (for next chat)

**Immediate (blocking):**
1. **Check revise-loop demo completion:**
   - Query `diagnosis_artifacts` WHERE `correlation_id='e08c5b01-01ef-42ad-80d0-b77c50ec9e84' AND kind='council_report'` — count should be 2+ if run looped.
   - If completed: pull plan v3 (first round after revise), council reports, check decisions (approved/revise/exhausted).
   - If still running: wait and re-check; if timeout, check orchestration_states for `__step_error`.

2. **Commit untracked workstream files:**
   - Snapshot `diagnose_council_decide_action.go` + test before committing (already snapshotted during build, but good hygiene).
   - Commit message should reference: "F2.2 revise loop built and tested; council round count scoped per orchestration_id (not correlation)."

**Next (after demo result):**
3. **If demo approved an updated plan:**
   - Plan converged → proceed to F1.1b(c) design review (branch+PR behind write token, isolated credential, gofmt+build gate in spawned Job).

4. **If demo returned exhausted or rejected:**
   - Diagnose why; expected path is: plan → revised plan → approved (max 2 revises).
   - If exhausted at round 2: looks like the rounds are being counted correctly but the model is stuck. Escalate to F1 conversation.

5. **If demo is still running:**
   - Schedule a checkpoint in ~10 min; revise loops should complete in ~2 min per round (propose ~60s, review ~30s each, decide ~10s).

---

## Gotchas & Landmines (live on this platform)

**Max_tokens placement (CRITICAL):**
- `execute_llm_prompt` reads `max_tokens` **ONLY** from agent-top-level config or **INSIDE the step's `ai_service` block** (ai_actions.go:252-256).
- Root-level step-config `max_tokens` is dead; Anthropic client defaults to 2048 output tokens.
- Diagnose-agent verdict step + fix-proposer propose step were both silently capped at 2048 until fixed.
- Any new step using `execute_llm_prompt`: place `max_tokens` inside `ai_service`, not at root.

**Kafka consumer-rebalance window (gotcha from run-2):**
- Never fire an orchestration within ~5 min of a pod rollout/restart.
- Rebalance silently drops the spawn into the void; pod logs show nothing.
- Workaround: always wait 300s after a deploy before firing a run.
- Run-2 (turn 6) was lost to this; cost 8 hours of debugging.

**Round counting (just fixed):**
- Council reports accumulate on the correlation_id (the diagnosis).
- Proposer runs are separate orchestrations on the same correlation.
- If round count is scoped by correlation alone, run-N would see run-(N-1)'s reports and start at round 2+ instantly.
- **FIX (applied):** count per orchestration_id + correlation_id.
- If demo run looks like it exhausted at round 1, this wasn't applied. Verify Go deployed.

**Snapshot discipline (from turn 20 feedback):**
- Always call `snapshot_agent()` before updating an agent definition.
- Snapshots go to a `_backup` table, named by snapshot ID and timestamp.
- Snapshots are how you revert a live agent that broke: load the snapshot as the live definition.

**Platform bugs affecting the pilot:**
- None currently known. Runs 1–5 were all defects in the loop itself, not the platform.

---

## The Test Bed (still live, still broken)

**Dartsonline guides defect:** a navigation link to a blank page. Known answer (hand-diagnosed turn 1):
- Root cause: `page-build-handler` has a success-terminal `complete_error` step that should be `fail_workflow` (mark_no_sections).
- Secondary: nav-generation code (`loadPagesForNav`) filters on `status` not `build_status`.
- Benchmark is repeatable because the page is still blank (could be fixed by hand any time, but we're using it as a training target).

**Run 5 benchmark result:** CONFIRMED under the strictest gates, citations include nav-generation code for the first time, [context] marks on control clauses working, fair-share sibling budget fix proven end-to-end.

**Run 5 → plan v2 → council result:** plan was much better than v1 (cited `complete_error`/`fail_workflow` directly), but council sent it back as `revise` with correct objections (edit 1 targets wrong causal path, edit 2 doesn't name owning pipeline, safety question on unbounded retry).

**Demo run (current):** intended to show the loop converging across 2–3 rounds on run 5's CONFIRMED. If it converges to `approved`, F1.1b(c) is unblocked. If it exhausts, we debug convergence. If it's still running, wait for results.

---

## Files to Read First (new chat)

1. **MILESTONE_diagnosis_fix_loop_2026-07-10.md** — what was built and why it matters (shareable, no code required).
2. **NOTES_running_fixloop(10).md § turns 16–20** — the detailed arc of F2.2 from design to deployment.
3. **PLAN_fixloop_pilot.md § F2.2 REVISE LOOP** — technical summary of what's live.
4. **diagnose_council_decide_action.go § lines 145–170** — the round-counting fix (orchestration_id scoping).
5. **0NN_fix_proposer.sql § v3 seed line 1** — the full 10-step workflow with repropose loop.

---

## Database Queries (for checkpoint)

```sql
-- Check revise-loop demo status
SELECT COUNT(*) as council_reports, COUNT(DISTINCT orchestration_id) as rounds
FROM diagnosis_artifacts
WHERE correlation_id='e08c5b01-01ef-42ad-80d0-b77c50ec9e84' AND kind='council_report';

-- Fetch the latest decision
SELECT metadata->>'decision' as decision, metadata->>'reviewers' as reviewers
FROM diagnosis_artifacts
WHERE correlation_id='e08c5b01-01ef-42ad-80d0-b77c50ec9e84' AND kind='council_report'
ORDER BY created_at DESC LIMIT 1;

-- Count plans (should be 2+ if revise loop fired)
SELECT COUNT(*) as plans, MAX(created_at) as latest
FROM diagnosis_artifacts
WHERE correlation_id='e08c5b01-01ef-42ad-80d0-b77c50ec9e84' AND kind='fix_plan';

-- Check orchestration_states for demo run
SELECT status, current_step, string_agg(DISTINCT status, '|')
FROM orchestration_states
WHERE correlation_id LIKE 'demo-run-corr%' LIMIT 5;
```

---

## Next Phase (F1.1b(c)) — Design Summary

**The write step:** if the revise loop produces an `approved` plan, turn it into a code branch and PR.

**Constraints:**
- Write token is isolated (injected only into fix-implementer pods, never shared chassis).
- Gated on code compiling (gofmt + go build in a spawned golang-image Job — chassis image has no toolchain).
- Human review terminal (PR created, but no auto-merge).
- Plan's file list is a hard allowlist (edit operations apply only to named files).

**Design (not yet built):**
1. `fix-implementer` agent (new).
2. Input: approved plan + diagnosis + last bundle.
3. Steps: load_approved_plan → sketch_to_diffs (LLM, turn sketches into concrete diffs) → apply_diffs (action, file-allowlisted) → branch (new branch, commit diffs) → gofmt+build (Job) → create_pr (GitHub).
4. Output: PR URL, metadata (branch name, author, checks status).
5. Spawn gate for GITHUB_WRITE_TOKEN (new, mirrors isRepoCloningAgent pattern).
6. PR body carries: diagnosis (conclusion), coverage (symptom_check entries), plan (edits + grounded_in), council report (decision + objections).

---

## Open Design Questions (F2.2+)

**Q-D (hard-veto placement):** currently in step config (`hard_veto_from: [guardian]`). Should it live in agent-definition column (reviewer roles)? Deferred to F2.3.

**Q-C (write token scope):** GITHUB_WRITE_TOKEN per project or per user? Current design assumes per-user (injected at pod spawn). Check with your token strategy.

**F0.3 (iteration_note rows):** the table exists, the kind is defined, nothing writes them yet. Would provide per-iteration summaries. Low priority, bundled into a later F0 refinement.

**F3 (learning record):** categorize confirmed bugs into recurring classes for earlier detection. Post-MVP, design TBD.

---

## Commits Needed (before or after handoff)

One commit covering:
- `diagnose_council_decide_action.go` + test (new, 209 + test lines).
- Updated `NOTES_running_fixloop(10).md`, `PLAN_fixloop_pilot.md`, `RUNBOOK_diagnosis_fix_loop(10).md`.
- New `MILESTONE_diagnosis_fix_loop_2026-07-10.md`.
- Updated `0NN_fix_proposer.sql` (v3 seed).
- FYI addendum to travelling_docs.

**Message template:**
```
F2.2 revise loop: built, tested, deployed on v1.0.1107.

- diagnose_council_decide action (deterministic decision aggregation, orchestration_id-scoped round counting)
- revise loop wired into fix-proposer workflow v3 (10 steps, repropose on revise)
- per-orchestration round cap prevents prior-run report inflation (critical fix from design review)
- milestone doc added (shareable narrative of what was built and why)
- runbook/plan/notes current through turn 20, revise-loop demo in flight

Co-Authored-By: Claude Haiku 4.5 <noreply@anthropic.com>
```

---

## Resume Checklist

- [ ] Read MILESTONE doc (5 min, context for any stakeholders).
- [ ] Check revise-loop demo status (SQL queries above, 2 min).
- [ ] If complete: pull results, grade against expectations (10 min).
- [ ] If incomplete: wait or debug (per completion time estimate, 5–30 min).
- [ ] Verify v1.0.1107 deployed and council code reached pods (check pod logs for "diagnose_council_decide", 2 min).
- [ ] Commit untracked files if not done (5 min).
- [ ] Next decision: approve demo result → F1.1b(c); revise needed → iterate; timeout → escalate.

---

## Why This Matters (for the record)

By the end of this session, the loop has four organs working together under load:
1. **Diagnosis** — reads-only, grounded, refuses to guess, explains the whole symptom.
2. **Planning** — constrained, evidence-cited, minimal, no-ops rejected.
3. **Review** — two independent specialists (quality + safety), deterministic aggregate.
4. **Iteration** — revise loop feeds objections back to the model, converges before human approval.

The revise loop is the keystone: it means a plan doesn't ship because one LLM said "yes", it ships because the plan survived objections and improved. The demo run measures whether that actually works.
