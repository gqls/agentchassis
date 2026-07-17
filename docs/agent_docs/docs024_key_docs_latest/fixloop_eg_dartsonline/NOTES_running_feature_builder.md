# RUNNING NOTES — feature builder (fixloop delta, docs024/fixloop_eg_dartsonline)

**What this thread is.** Builds the multi-step feature-construction capability
on top of the proven fix loop — design §1 of
`DESIGN_feature_builder_and_council_gate.md`, booted from
`HANDOFF_2026-07-17_feature_builder_thread.md`. Session name: "fixloop feature
builder". Coordinates with the concept register workstream (docs026), whose
bug-historian council seat serves this chain too.

Newest entries at the bottom. Update every turn.

---

## Turn 1 — 2026-07-17 — Boot; staged-plan schema drafted for sign-off

Read (per the handoff): design §1, `0NN_fix_proposer.sql` (v6 live shape),
`0NN_fix_implementer.sql`, `0NN_fix_implementer_orchestrator.sql`, and the
three Go seams (`diagnose_persist_fix_plan_action.go`,
`diagnose_prepare_fix_commit_action.go`, `diagnose_read_repo_files_action.go`),
plus the concept register's running notes (thread-boot instruction).

Drafted and committed `SCHEMA_staged_plan_v1.md` (`cc136c902`) — the stages[]
schema, cross-stage file discipline, encoded image-before-seed checklist,
compatibility map, the F1.2 pilot as a worked instance, and six owner
decisions D1–D6. **Key structural finding:** F1.2 (implementer ref/base as
per-run input) is a PRECONDITION of the stage loop, not a cleanup — stage 2+
must read the fix branch, and `ref` is a config literal at
`diagnose_read_repo_files_action.go:101`. No code written; sign-off requested.

## Turn 2 — 2026-07-17 — Owner signed off (all recommendations); delta 1 built

Owner: "go with all recommendations" — D1 (reuse `kind='fix_plan'` +
`plan_format` discriminator), D2 (`feat/<short-corr>` branches), D3 (caps
6/8/24/128KB), D4 (seed/checklist rules as hard validation), D5 (F1.2 pilot,
self-hosted), D6 (derived go-test end gate).

Built delta 1:

1. **Staged validation** in `diagnose_persist_fix_plan_action.go` (commit
   `4b3d50f4c`): discriminated by `plan_format`/`stages`; legacy single-plan
   path behaviourally unchanged (per-edit rules factored into shared
   `editProblems`); staged path adds stage shape rules, add-once /
   no-modify-before-create / no-create-then-delete, the gate.build⇒seed/doc
   implication, and the checklist contract (seed⇒exactly-one seed_apply;
   image_deploy strictly before any seed_apply — the wrong order is
   unexpressible). 25 new test cases + probe discriminator tests, full
   package green.
2. **Feature-designer seed** drafted: `0NN_feature_designer.sql` — a DRAFT
   FILE, deliberately not applied (the seed discipline this feature encodes,
   applied to itself). Reuses the proposer's chain wholesale: staged persist →
   3-seat council (prompts adapted for staged judging; bug-historian digest
   verbatim) → same deterministic router/verify/reframe/escalate. Intake gate:
   `site_work_items` capability_gap items whose `spec` carries BOTH
   `owner_approval` and `code_pointers` (the designer runs in-chassis with no
   repo read token, so specs must carry curated paths; no new status values —
   the dedup-index/status contract stands). Deliberate delta from v6:
   run_checks answers all THREE reviewers' checks. Models: sonnet-5 for
   design/repropose/reframe, sonnet-4-6 for reviewers. MDL-039 guard: no root
   ai_service key.

Context that moved under us mid-turn: the makefile build default inverted
(`make build-<service>` now = committed HEAD; `-tree` is the WIP escape
hatch) — schema doc references updated to the new default form.

**Where this leaves the loop:** delta 1 complete on the code side. Next:
delta 2 (implementer stage-loop — per-stage allowlist/read/commit/gate, `feat/*`
branch, derived go-test end gate, new deterministic stage router action), then
the F1.2 pilot through the full chain. The owner's future acts (not now, in
order): merge/build image carrying `4b3d50f4c`+delta 2 → apply
`0NN_feature_designer.sql` → approve a pilot spec work item.

**Provenance note (same turn):** this turn's three doc artifacts
(`0NN_feature_designer.sql`, this file, the schema doc's sign-off edit) were
swept into a concurrent session's bulk commit `cf3803b49` ("product specs
additions and miscellaneous runbooks and handoffs") between our add and
commit — the exact index race `CLAUDE.md` documents. Content verified intact
in that commit; per the forward-only rule, no corrective action — this note is
the record. The delta-1 Go change committed cleanly under its own message
(`4b3d50f4c`).

## Turn 3 — 2026-07-17 — E1–E5 signed off; delta 2 BUILT; standing docs established

User: keep running notes + runbook + plan at all times, write a read-aloud
summary, carry on with all recommendations. E1–E5 thereby approved.

**Delta 2 built** (commit `c19b5d097`, full package tests green):

1. `feature_stage_route` — the loop's only new control machinery. Emits each
   stage as a SINGLE-PLAN shape so the proven read/prepare actions loop
   unchanged; per-stage read ref (base for stage 1, the feat branch after);
   per-stage commit message; terminal emission carries the PR payload
   (checklist rendered as an owner task list) + go-test packages DERIVED from
   the plan's .go edits (D6). E4 enforced at seed time: a pre-existing
   `feat/*` branch is a hard refusal via a GitHub API existence check.
2. Seams, all optional-field additions with single-plan behaviour untouched:
   `diagnose_read_repo_files` gains `ref_field`; `diagnose_prepare_fix_commit`
   gains `branch_field`/`commit_message_field`/`expected_symbols_field`
   (symbols checked against produced bodies); `diagnose_build_gate` gains
   `test_packages_field` (E2); `feature-implementer` joins the
   isRepoCloningAgent spawn gate (E1).
3. Seeds DRAFTED as files (`5b131b88a`): `0NN_feature_implementer.sql` (22
   steps, graph-validated), `0NN_feature_implementer_orchestrator.sql`
   (dedicated-pod wrapper — the read-token lesson), and two trigger drafts on
   the proven 092 kcat envelope.

**Council roster moved under us mid-turn:** the concept-register thread
shipped `review_reuse_agent` (fix-proposer v7, 4 seats). Extended the
feature-designer seed to mirror it — chain editquality → bug_historian →
reuse_agent → guardian, all four seats' checks answered. Reuse is this
builder's hard rule 1, so the new seat bites hardest here. RUNBOOK A3 covers
future roster drift.

**Standing docs established** (user request): `PLAN_feature_builder.md`,
`RUNBOOK_feature_builder.md` (A1–A7: image → seeds → roster check → pilot
spec SQL → fire designer → fire implementer → close out),
`SUMMARY_feature_builder_2026-07-17.md` (read-aloud), this file continuing as
the running record. RUNBOOK A5 evolves E5 slightly, flagged as the owner's
choice: with the designer built, prefer firing the full chain on the pilot
spec and GRADING its plan against the hand-written §6 reference, over
hand-injecting the plan.

Another index-race sweep this turn: registry.go's new entry rode into
`aabd38161` (experience-loop's commit). Content verified intact; noted here,
forward-only.

**State: all delta-1+2 code committed and inert. Everything from here is
owner acts (RUNBOOK A1–A7) — nothing further to build until the pilot's
grades come back.**

## Turn 4 — 2026-07-17 — Image built+pushed; rollout HELD on live traffic (user correction); designer synced to v8

User said "let's carry on"; began RUNBOOK A1. Bumped IMAGE_TAG → v1.0.1131
(commit `202019e6c`), built agent-chassis from committed HEAD, pushed to
docker.io/aqls. **User stopped the rollout step and said: read CLAUDE.md and
follow it.** They were right, twice over: (1) acting on the stale
session-start git status — the chassis kustomization was already dirty with
another session's uncommitted newTag; (2) restarting the shared chassis
without checking what's in flight.

Checks then run (the CLAUDE.md way): the dirty kustomization diff is ONLY
newTag 1128→1130 residue matching the live deployment (v1.0.1130, 1/1
ready) — safe. But orchestration_states shows a LIVE pipeline: 1
EXECUTING_STEP (spawn_link_resolver, seconds old) + 8 AWAITING_RESPONSES,
new arrivals every 1–2 minutes since 17:28. A restart now kills the
executing step (never reaped — bugs_open/003) and drops spawns for ~300s.
**Rollout held; owner picks the moment.** v1.0.1131 sits pushed and ready.

Council roster moved AGAIN mid-day (v8, 5 seats: + review_guidelines).
Synced the designer seed: chain editquality → bug_historian → reuse_agent →
guidelines → guardian; all five seats' checks answered; guidelines prompt
carries the wrapper-orchestrator / dedup-contract / declared-contracts /
schema-tier rules + the approve-don't-object meta-rule for stale guidelines.
Graph re-validated (22 steps, all targets defined).

## Turn 5 — 2026-07-17 — LIVE: image verified (v1.0.1132), three seeds applied, pilot spec created; approval is the owner's

Owner answered the two questions: (1) another thread owns the rollout —
hands off; (2) yes, apply the seeds after the pod verifies.

The other thread's **v1.0.1132** landed while we watched — running pod
binary verified (`feature_stage_route` count 3), so A1 closed without us
touching the deployment. Applied the three seeds to
postgres-clients-0/clients_db in order (`INSERT 0 1` × 3) and verified:
step counts 22/22/3, designer chain editquality → bug_historian →
reuse_agent → guidelines → guardian, 5 review_fields, no root `ai_service`
anywhere. **The feature builder is now REGISTERED and inert in production.**

A4: created the pilot spec work item
`db066cac-c647-44bf-a3ca-e04416405b28` (capability_gap,
needs_human_review, item_key f12-ref-input-pilot, goal + 3 code_pointers).
Caught a runbook bug on the way: the site anchor is named `System
(internal)`, not `system.internal` — the drafted INSERT would have silently
inserted nothing; fixed in the runbook.

**Remaining, both the owner's: the named approval UPDATE (runbook A4) and
the credit-spending go to fire the designer (A5).** Mind the ~300s
no-dispatch window if the chassis pod has just restarted.

## Turn 6 — 2026-07-17 — New chassis image v1.0.1134; machinery re-verified live

Owner: a new chassis image was deployed. Verified the running pod
(`agent-chassis-6d85fff446-54jzc`, `v1.0.1134`, started 19:34:11Z): binary
carries `feature_stage_route` (×3) AND the delta-2 expected-symbols guard
("expected symbols not present"). All three feature agent rows still active
(an image swap doesn't touch the DB — confirmed, not assumed). At 19:43 the
pod is ~9 min old, past the ~300s dispatch-drop window, so a fire would be
safe timing-wise. Pilot work item `db066cac` still `needs_human_review`,
`owner_approval` absent.

**Nothing new to build or apply. The two remaining acts are the owner's and
cannot be delegated: the NAMED approval UPDATE (A4 — it records who
approved) and the credit-spending GO to fire the designer (A5).**

## Turn 7 — 2026-07-17 — FIRST FIRE: feature-designer live on the F1.2 pilot

Owner approved (aaa) and gave the fire go. Confirmed pod age 1530s (past the
300s window), fired `0NN_TRIGGER_feature_designer_v1.sh db066cac`.

- FEATURE correlation: `fc2cf851-ab2d-4303-a8e4-a545275f6ee3`
- Run orchestration: `69def2d7-8932-461b-8258-93c4444ce3ab`

Confirmed live it passed the spec gate: load_spec → check_spec_approved →
load_schema_hint → design — i.e. the owner_approval + code_pointers gate
computed `approved` on real data (first end-to-end proof of the intake gate).
**RESULT — PARTIAL SUCCESS, validation did its job.** The run went
load_spec → check_spec_approved → load_schema_hint → design → persist_plan →
`complete_refused` (COMPLETED, no artifact written), in ~90s. First proof the
intake gate works on real data, and the designer produced a well-formed
4-stage staged plan (s1/s2 code, s3 seed, s4 config_change) with a checklist —
but staged VALIDATION refused it, all three reasons on stage s4 (a
config_change edit):
1. config_change `file` was prose with spaces/parens (`fixloop….fix_implementer
   (live workflow definition row…)`) → tripped the repo-relative/no-whitespace
   rule (shared with single plans, where config_change targets look like
   `agent_definitions:page-build-handler`).
2. `artifact_role: "doc"` on a config_change → my contradiction rule fired.
3. `config_change` used as a post_merge_checklist act → not in the allowlist
   (image_deploy|seed_apply|verify).

Diagnosis: two are model-steer gaps (config_change file format; role), one is
the model over-decomposing — it added a config_change "activate v2" stage that
is redundant because the seed's `ON CONFLICT DO UPDATE` already activates. NOT
a schema gap: config_change edits are already surfaced in the PR body's own
section (buildStagedPRPayload), so they were never meant to be checklist acts.
No council credits spent (refused before review).

**Fix applied (prompt-steer, contract preserved):** designer seed rule 5
now states a seed upsert IS the activation (no separate activate stage);
new rule 5b constrains config_change — `file` = compact `agent_definitions:<type>`,
role code/omit never seed/doc, appears in PR body not the checklist. Seed
re-validated (balanced, graph intact). This is a seed-FILE edit — the LIVE
designer row still has the old prompt until the seed is re-applied (owner act).

**Next (owner calls, both stated):** re-apply the improved designer seed
(prod DB write) + re-fire on db066cac (credits). Optional spec sharpening:
the Go seams (ref_field etc.) already exist since c19b5d097, so F1.2's real
remaining work is the fix_implementer WORKFLOW seed — the spec's goal could
say so, to stop the designer proposing redundant s1/s2 Go edits.

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->
