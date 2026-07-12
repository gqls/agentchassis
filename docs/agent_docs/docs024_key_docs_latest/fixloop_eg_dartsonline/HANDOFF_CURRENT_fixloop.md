# HANDOFF — Diagnosis→Fix Loop (LIVING DOCUMENT — update every turn)

*Convention: this file is rewritten/amended at the end of every working turn so
a fresh chat can resume from exactly here. Superseded point-in-time handoffs:
HANDOFF_turn21_2026-07-10.md (historical). Last updated: 2026-07-12, turn 25.*

---

## Read first (fresh chat bootstrap)

1. This file, top to bottom.
2. `SUMMARY_write_step_position_2026-07-12.md` — plain-language position + the
   open build-gate decision.
3. `RUNBOOK_diagnosis_fix_loop(10).md` §CURRENT POSITION + §gotchas.
4. `NOTES_running_fixloop(10).md` turns 22–25 (the F2.3→F1.1b(c) arc).

## State snapshot

**Deployed & proven (v1.0.1108 chassis + fix-proposer workflow v5):**
- Diagnosis loop (F0.x): gated CONFIRMED with symptom coverage, durable bundles.
- Fix-proposer + council (F1.1a, F2.1–F2.3b): constrained plans; 2-reviewer
  council; decision router (approved | revise→verify→repropose |
  first-veto→reframe-once | rejected/exhausted→escalate); verify step (8/8
  reviewer SQL checks answered after the v5 live schema hint); escalation
  artifact (kind='escalation'). Round counting orchestration-scoped, proven.
- Reframe path is unit-tested but has never fired live (no veto since v4).

**Built, committed, NOT yet live:**
- git-adapter: create_branch / create_pull_request / branch-aware commit
  (commit 89175383) — needs a **git-adapter image rebuild**.
- chassis: diagnose_prepare_fix_commit allowlist safety core (commit a4c6cc63)
  — rides the **next chassis image**.

**Decisions CLOSED (owner, 2026-07-12 turn 26):**
1. **Build gate = B**: pre-PR golang k8s Job (clone fix branch, gofmt changed
   files + targeted go build); broken implementations never become PRs.
2. **First write-step run = seeded small bug** (single-file, contained, a plan
   the council can genuinely approve); a real bug after.
3. **Awareness = standing rule**: more awareness BEFORE wider autonomy; the
   digest surface is the slice after F1.1b(c), before council-widening.
4. **NO FORK**: isolation = fix/* branches + owner-gated merges on this repo.

**F1.1b(c) CODE COMPLETE (turn 26, all committed, none deployed):**
- `diagnose_build_gate` — golang k8s Job: gofmt CHANGED FILES ONLY + TARGETED
  go build (repo-wide fails on pre-existing docs-dir clashes / unformatted
  legacy — encoded in tests); red = result routed to no-PR terminal;
  BackoffLimit 0; TTL 1h; RBAC pods/log added to rbac-job-spawner.yaml.
- `diagnose_read_repo_files` — plan's modify/add files via GitHub contents API
  (raw media type; read token from spawn gate; modify-404 = hard error).
- `git_adapter_request` — ONE generic adapter caller (allowlisted verbs:
  commit/create_branch/create_pull_request; delete_repo unreachable), data
  from config paths/literals, awaits adapter response.
- `fix-implementer` added to isRepoCloningAgent (read token only).
- Seed `0NN_fix_implementer.sql` (15 steps, graph verified, dry-run passed):
  load plan/council/diagnosis → check_approved gate → read_current_files →
  sketch_to_files (whole files, no drive-bys) → prepare (allowlist) →
  create_branch → commit_files → build_gate → check_gate →
  {create_pr→complete | complete_gate_failed (NO PR, branch+log left)}.
- Trigger: `092_TRIGGER_fix_implementer_v1.sh [fix_correlation_id]`.

**DEPLOY CHECKLIST — ALL DONE 2026-07-12 (turn 27):**
1. ✅ Chassis v1.0.1110 verified in-pod (binary Jul 12 18:11; all four new
   action strings present).
2. ✅ git-adapter rebuilt + verified in-pod (create_branch/create_pull_request).
3. ✅ RBAC applied (pods/log on agent-job-spawner).
4. ✅ 0NN_fix_implementer.sql applied — fix-implementer live, 15 steps,
   v1.0.1110.
5. ✅ Write-scope smoke PASSED: adapter created smoke/write-scope-test from
   main (sha 4c2c172b) — token writes to gqls/agentchassis; branch deleted
   after.
6. ⏳ First end-to-end target — see "first-run design decision" below.

**First-run design decision (owner input wanted):** literally PLANTING a bug
can't honestly feed the diagnose loop — a planted bug that never ran has no
agent_error_log/DB rows, so the tier guard would (correctly) refuse CONFIRMED;
fabricating evidence rows is off the table. Recommended instead: pick a REAL,
tiny, zero-risk defect (candidate: the raw `fmt.Printf("DEBUG: ...")` lines in
ai_actions.go:722 / generate_image_actions.go:594 — the latter even logs the
WRONG function name, a genuine copy-paste bug; violates the logging
constitution; observable in pod stdout), HAND-AUTHOR its diagnosis (true,
cited — the workstream's known-answer tradition) as a CONFIRMED
orchestration_states row, then run the REAL proposer→council (should approve a
minimal single-file fix) → REAL implementer → branch → gate → PR. Everything
new gets exercised honestly; the planted-bug idea inverts to "real small bug
first, planted bug never needed".

## Key artifacts & tools

- Trigger: `091_TRIGGER_fix_proposer_v1.sh [fix_correlation_id]` (defaults to
  the benchmark correlation `e08c5b01-01ef-42ad-80d0-b77c50ec9e84`).
- DB access: `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U
  clients_user -d clients_db`.
- Checkpoint queries: council reports/plans/escalations by
  `correlation_id + orchestration_id` in `diagnosis_artifacts`
  (kinds: bundle | fix_plan | council_report | escalation).
- Seeds: `0NN_fix_proposer.sql` (v5, applied); fix-implementer seed not yet
  written.

## Live gotchas (cost hours each — do not relearn)

- **Same-tag deploys ship stale binaries**: bump IMAGE_TAG (makefile line 16)
  and verify IN THE POD: `grep -ac <new-symbol> /proc/1/exe` (control-string a
  known-present symbol to validate the method).
- **Deploy order for workflow seeds**: image first, seed second — a seed
  naming unregistered actions fails at runtime.
- **Rebalance window**: never fire within ~300s of a chassis pod (re)start.
- **max_tokens** lives INSIDE a step's `ai_service` block; root is dead config
  (client defaults to 2048 and truncates JSON mid-plan).
- **Round counting on <v1.0.1108** was per-correlation (stale reports inflate
  rounds) — fixed, but relevant when reading old run data.
- **Git: forward-only.** Other chats commit to the same branch concurrently —
  no resets/amends; check `git log` before assuming your commit is HEAD.
- `snapshot_agent` writes to `agent_definitions_backup`
  (`snapshot_taken_at`/`snapshot_reason` columns).
- BST-vs-UTC: dev host +0100, DB UTC; `orchestration_states.last_activity` is
  timestamp WITHOUT tz.

## The benchmark bug (still live, still unfixed — deliberately)

dartsonline guides defect: known answer = `complete_error` success-terminal
should be `fail_workflow` (cause B) + nav filters on `status` not
`build_status` (cause C) + sections=[] partition (cause A). The loop has
CONFIRMED it (run 5) and the council consistently — and correctly — judges the
full fix architecture-level: every clean proposer run ends
`exhausted/rejected → escalation`. That is the honest terminal, not a defect.
The hand fix can be applied any time; doing so retires the benchmark.

## Direction (owner intent, stated turn 25)

Owner wants: multi-perspective councils on EVERY task (hallucination
minimisation; guideline/mission/structural-decision conformance, changing only
when completely right); simultaneous legacy-migration review; feature work
from specs/mission docs eventually — all PR-gated; an awareness surface so the
owner stays informed as autonomy grows; and the next phase in a SEPARATE
FORKED REPO. Treat these as the design brief for what follows F1.1b(c).
