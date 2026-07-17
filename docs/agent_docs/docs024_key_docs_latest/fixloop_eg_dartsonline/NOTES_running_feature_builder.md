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

<!-- Append new turns below this line. Format: ## Turn N — date — one-line summary -->
