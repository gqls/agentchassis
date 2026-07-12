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

**In flight:** the build-gate action (k8s Job spawn + wait) — task 9.
**Then:** fix-implementer seed (task 10) → chassis + git-adapter image
rebuilds → seed the small bug → first end-to-end run.

**Build-gate implementation notes (earned this repo):** `go build ./...` at
repo root FAILS today on pre-existing docs-dir package clashes, and gofmt -l
repo-wide flags pre-existing unformatted files — the gate must build TARGETED
paths (./platform/... ./internal/... ./pkg/... ./cmd/...) and gofmt ONLY the
implementation's changed files, else every gate run fails on inherited mess.

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
