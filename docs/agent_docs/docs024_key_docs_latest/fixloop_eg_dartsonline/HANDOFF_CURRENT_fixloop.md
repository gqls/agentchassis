# HANDOFF — Diagnosis→Fix Loop (LIVING DOCUMENT — update every turn)

> ➡️ **FRESH CHAT ("diagnosis fixloop 2")? Read `HANDOFF_diagnosis_fixloop_2.md`
> FIRST** — it is the clean, self-sufficient cold-start (immediate next action =
> take triage live once the shipping chassis image lands). This file below is the
> turn-by-turn running history that fed it.

## ★ MILESTONE 2026-07-13: THE LOOP OPENED ITS FIRST PULL REQUEST ★

**https://github.com/gqls/agentchassis/pull/1** — end-to-end, every gate live:
hand-authored CONFIRMED diagnosis of a REAL defect (generate_image_actions.go
raw fmt.Printf naming the wrong function) → proposer plan → council APPROVED
(round 1, both reviewers — the council's first approval) → implementer in a
DEDICATED pod (read token via spawn gate; writes via git-adapter only) → read
file at the correct ref → 41KB whole-file rewrite → deterministic allowlist
PASS → fix/11111111 branch → commit → golang-Job build gate GREEN → PR #1
with the Q-H package in its body. **The diff: 1 file, 2 deletions, zero
drive-bys — exactly the plan.** Human review terminal: the PR awaits the owner.

Also earned en route: the gate's FIRST RED correctly blocked a PR when
./cmd/... carried pre-existing breakage (cmd/test-spawning stale 3-arg
NewSagaCoordinator call — real bug, fixed + pushed 9f29efb9, found by the
gate). Gotcha recorded: a stale fix/* branch from a failed run must be deleted
before re-firing (create_branch is idempotent and would reuse the old base).
Pilot config note: implementer ref/base/from_branch point at
084_site_improvements_local_ai via live jsonb_set (origin/main is stale; the
committed seed still says main — align when main catches up).

*Convention: this file is rewritten/amended at the end of every working turn so
a fresh chat can resume from exactly here. Superseded point-in-time handoffs:
HANDOFF_turn21_2026-07-10.md (historical). Last updated: 2026-07-12, turn 25.*

---

## TRIAGE LIVE in DRY-RUN (turn 34, v1.0.1116) — first sweep reveals a needed filter

First dry-run sweep (6ae98f10) COMPLETED cleanly, 0 work items written
(confirmed), report in doc_notes (categories triage). Found 9 real loud-failure
patterns. THE DRY RUN EARNED ITS KEEP: several patterns are OPERATIONAL, not
code bugs — "Claim timed out (attempts exhausted)", "Claim timed out — handler
pod likely died" (dispatch/infra failures), and "(no error text)" (no signal).
Only some are genuine code bugs (component-creator "store_component failed:
new row violates ..." constraint; template "rejected by pre-store validation").
With cap=3 ordered by count, flipping dry_run→false then would have escalated 2
operational patterns + 1 real bug — sending the loop to diagnose "pod died".

**Phase 1.1 loop-worthiness FILTER BUILT (owner chose (a)); committed, needs
next image.** `triageRoute` classifier (pure, 6 tests green): "" → hold(human);
transient/infra signature (claim timed out / pod likely died / handler pod /
consumer rebalance — precise to the dispatch/pod layer, tunable via
transient_signatures) → requeue (surfaced, NOT escalated); real error → loop.
ONLY code-bug patterns escalate (capped). Report now has 4 routes: code bugs→
loop | transient→re-queue | no-signal→hold | capability gaps→roadmap. On the
same live data, this would escalate the component-creator "store_component
failed" constraint bug and drop the claim-timeout/no-signal noise. (Code landed
in concurrent commit ab5dee1dc; verified in HEAD.)
**NEXT:** next chassis image (verify `grep -ac triageRoute /proc/1/exe`) →
re-run dry-run (095) → confirm only real bugs in the "code bugs → fix loop"
group → flip dry_run→false (jsonb_set in seed footer) → fire again to escalate
for real → then the escalated needs_diagnosis items are ready for the diagnose
loop (manual dispatch; the dispatch loop is shipped-disabled).

## TRIAGE ROUTER (Phase 1) BUILT (turn 33) — awaiting next image; PR #1 fix now on main

**`main` fixed:** cherry-picked the stranded PR #1 fix commit (670d6dd2) onto
main via a throwaway worktree and pushed (218e3b52..998c0b31). Verified: main no
longer has the DEBUG defect. The PR-ordering wrinkle is resolved.

**Triage Phase 1 built + committed (NOT deployed — needs next chassis image):**
- `diagnose_triage` action — deterministic (no LLM). Scans site_work_items:
  LOUD failures (`status='failed'`) → deduped by (item_type, handler, error
  signature) → escalate PATTERN to needs_diagnosis (090 contract:
  system.internal anchor eac60db8, pipeline='diagnose', awaiting_diagnosis,
  ON CONFLICT dedup on item_key, capped at max_escalations=3); CAPABILITY GAPS
  (`capability_gap`/`deferred`) → surfaced to the roadmap in the report, NEVER
  escalated. One doc_note per sweep (categories triage+fixloop) — always writes
  (the visibility artifact). Pure helpers (item_key/symptom/spec/report) tested,
  5 cases green.
- **SHIPS dry_run=true** — previews escalations + writes the report, creates NO
  work items until the owner flips dry_run→false (jsonb_set in the seed footer).
- Seed `0NN_diagnosis_triage.sql` (dry-run validated on live DB, graph checked);
  trigger `095_TRIGGER_diagnosis_triage_v1.sh`.
- LIVE DATA CONFIRMS REAL WORK: the loud-failure query already finds ~8 genuine
  patterns (e.g. needs_new_component via component-creator, "store_component
  failed", 4 items; needs_page "Claim timed out", 6 items).
- **Deploy checklist:** next chassis image (verify `grep -ac diagnose_triage
  /proc/1/exe`) → apply seed → fire 095 (dry_run) → read the report
  (`SELECT body FROM doc_notes WHERE categories ? 'triage' ORDER BY created_at
  DESC LIMIT 1;`) → if good, flip dry_run→false → fire again to escalate for
  real. All manual (Fable credits low).

## TRIAGE + ESCALATION DESIGNED (turn 32) — next build; owner choices recorded

`DESIGN_triage_and_escalation.md` is the next slice (NOT yet built). Core idea:
the fix loop is fed by a handler FAILING, not a checker DETECTING. THREE failure
flavours, all already in the data:
- **loud** (`status='failed'`, attempts exhausted) → fix loop if code cause;
- **silent** (handler "completed" but problem persists — the darts bug) → needs
  a verification checker (THIS thread owns it) → fix loop;
- **no handler yet** (`item_type='capability_gap'`, `status='deferred'`,
  `builder_needed` — ALREADY emitted by load_work_item_actions.go:245-280) →
  roadmap/builder queue, NEVER the fix loop.
`diagnosis-triage` = thin router: scan site_work_items → loop-worthiness filter
→ DEDUPE by pattern (50 same-cause failures = 1 escalation) → route to
needs_diagnosis / roadmap / re-queue / human. Closed feedback: re-verify after a
fix deploys.
OWNER DECISIONS (2026-07-14): cadence hourly-for-now (slower later); this thread
owns verification checkers; MANUAL enablement for now. Phase 1 = loud failures +
capability gaps (both already in data). Not implemented yet.

**OPERATING-CONTEXT (2026-07-14): Fable credits running low.** Keep everything
MANUAL (no unattended auto-cadence consuming model calls). Docs are written to
be self-sufficient so the workstream survives a model change — the design is
gates + deterministic routing + human decisions, not model-dependent.

**GIT (turn 32): merge error fixed.** It was a dirty working tree (6 uncommitted
multi-session files), not a content conflict — checkpointed them (commit;
forward-only), `git merge origin/main` now clean. WRINKLE: `main` is missing PR
#1's fix — PR #2 merged 084→main at 13:44, PR #1 merged the fix→084 at 13:49
(5 min later), so main is 1 commit behind 084's fix. Clean 1-commit merge to
put it on main; awaiting owner okay to push to main.

## AWARENESS SURFACE: LIVE (turn 30, v1.0.1114) — first digest delivered

First digest ran 2026-07-13 21:00 (doc_notes, categories digest+fixloop) and
immediately earned its keep BOTH ways: the config-change ledger caught two
changes from OTHER workstreams (tool-acceptance-agent, asset-deployer) —
exactly the "machine changing without me knowing" coverage — AND its own first
output exposed its own gap: spawned children (the dedicated implementer pods,
carrying gate verdicts + the PR url) were missing from Runs because their type
lives at __execution_context__.sender.agent_type, not agent_group. FIXED
(commit 17b81535, COALESCE both locations; verified live, all 9 runs match) —
rides the NEXT chassis image; the deployed v1.0.1114 digest still misses
spawned children until then. Decisions(0) in the first digest was honest: the
council approval sat 25h back, outside the 24h window.
NOTE: working branch moved to 085_debug_and_feature_loops (owner); commits now
land there. NEXT: owner feedback on digest content → scheduled cadence
enablement → council-widening per the F2 roster → the real-bug run.

## PRIOR: IN FLIGHT (turn 29): the AWARENESS SURFACE (owner rule: before wider autonomy)

Built + committed, awaiting the next chassis image:
- `fixloop_digest` action — DETERMINISTIC digest (no LLM in the path): loop
  runs (status/terminal/gate/PR), decisions per correlation (kinds + latest
  council decision & why), and agent_definitions_backup snapshots in-window
  (the "what changed about the machine itself" ledger). Persists to doc_notes
  (pipeline/diagnose, categories ["digest","fixloop"]). Rendering pure +
  tested (empty sections read as "no activity", never "not checked").
- Seed `0NN_fixloop_digest.sql` (dry-run passed; apply AFTER the image) +
  trigger `093_TRIGGER_fixloop_digest_v1.sh`. v1 = manual trigger; a daily
  scheduled cadence is a deliberate later enablement once the owner likes the
  content.
- Read the latest digest:
  `SELECT body FROM doc_notes WHERE categories ? 'digest' ORDER BY created_at DESC LIMIT 1;`
- NEXT after image: apply seed → fire 093 → show the owner their first digest
  (it should feature PR #1's whole story). Then: council-widening per the F2
  roster, and the real-bug run.

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

**FIRST END-TO-END RUN — blocked at read_current_files (turn 27→28).**
- Proposer APPROVED (round 1, both reviewers) a clean single-file plan for the
  seeded defect — the council's FIRST approval. Correlation
  `11111111-e2e2-4a1b-9c3d-000000000001`.
- Implementer run `29bd92c8` reached read_current_files → complete_refused:
  `GITHUB_READ_TOKEN not in env`. ROOT CAUSE: fix-implementer runs IN-CHASSIS
  via the generic orchestrate path (sender pod agent-chassis, type generic),
  NOT as a dedicated spawned pod — so isRepoCloningAgent (which injects the
  read token at Job spawn) never fires for it. The write-isolation is fine
  (adapter-only); it is the READ token that has nowhere to land, and an
  explicit prior decision says "the chassis pod never holds the GitHub token".
- RESOLVED (owner 2026-07-13: "dedicated implementer pod that uses the
  git-adapter") = option B. Built `fix-implementer-orchestrator`
  (0NN_fix_implementer_orchestrator.sql): spawn_agent(fix-implementer) →
  call_agent(forward fix_correlation_id) → complete, mirroring
  diagnose-orchestrator→diagnose-agent. The spawned implementer pod gets
  GITHUB_READ_TOKEN via the already-deployed isRepoCloningAgent gate (reads
  in-pod, reaped after); WRITES still via the git-adapter; the chassis holds
  no token. NO image rebuild — SQL wrapper + retarget 092 to
  fix-implementer-orchestrator.
- Everything else in the write path is UNTESTED-BUT-READY downstream of the
  read: allowlist, branch, commit, build gate, PR.

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
  > **CORRECTED 2026-07-20 — this line is INVERTED; see `bugs_open/009`.** The
  > ROOT block won (first-found-wins); the step block was dead whenever a root
  > block existed. The rule above only held for agents with no root block.
  > Fixed by the step-wins overlay (`resolveAIServiceConfig`, ai_actions.go);
  > from that image on, the step block overrides root key-by-key.
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
