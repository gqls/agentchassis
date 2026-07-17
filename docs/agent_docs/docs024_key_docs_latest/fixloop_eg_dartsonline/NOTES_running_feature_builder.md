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

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->
