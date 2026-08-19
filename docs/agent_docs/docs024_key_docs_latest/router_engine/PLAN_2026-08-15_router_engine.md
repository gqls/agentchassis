# PLAN — the work-item router engine (RFC_030, owner-ruled 2026-08-15)

**Owner ruling (2026-08-15, evening):** *"I want each handler to be quite modular and
responsible for its own specific thing"* → *"Do as your recommendation recommends."* The
recommendation: one shared engine that RUNS per-type classifiers, each type keeping its own
classifier + route table as its own reviewable unit; the three existing routers are the
engine's first migrations. This lane builds it. Created by the bugfix_277 session at
hand-off time; the first working session should read `HANDOFF_2026-08-15_continue_here.md`.

## What exists today (the three routers — read all three seeds before designing)

| router | seed | routes | hardening it has |
|---|---|---|---|
| `image-url-404-handler` | 397 | 2 (convert `needs_imagery` / escalate via `checkpoint_for_review`) | none of 410's |
| `image-source-unsatisfiable-handler` | 397 | 2 (same shape) | none of 410's |
| `required-fields-missing-handler` | 410 v3 (CQ-023) | 8 (close×2 / park×4 / convert×2 / malformed→fail) | park-in-place holds dedup key; single-active-row assert; conversions born at a dispatchable status; component-keyed conversion item_keys; verify block asserting every branch pair |

All three are pure DB config: `query_database` classifier → `conditional_branch` cascade →
arms of `update_work_item_status` / `create_work_item` / `checkpoint_for_review` →
`complete_workflow`. **The classifier SQL is the only genuinely type-specific part.**

## The design question (for the lane's own council round — do NOT decide it in this file)

Two shapes are viable; the lane's first job is to pick one WITH the council:

**A. Config-driven generic agent.** One `agent_definitions` row (`work-item-router`) whose
workflow reads a per-type config row (`router_types` table or `agent_default_configs`
entry): `{item_type: {classifier_sql, routes: {label: {kind: close|park|convert, ...}}}}`.
Cheap to add a type (one config row), no image roll for a new type. Costs: the branch cascade
becomes data-driven, which today's `conditional_branch` cannot express without a loop step or
a Go action that evaluates a route table; per-type verify blocks become a generic validator.

**B. One Go action, `route_work_item`,** taking `classifier_sql` + `routes` from step config,
executing the classification and the chosen arm's effect in Go (reusing `insertWorkItem` for
converts, the same status UPDATE the park arms use). One agent per type still exists (a
thin seed naming the action + its config) — or one agent for all types keyed on
`current_item.item_type`. Costs: an image roll to change engine behaviour; gains: the
park/convert/close mechanics and the RFC_022-style accumulation counter live in ONE tested
place, and a type's config is validated by `RegisterActionInputSpec` (StrictConfig) at seed
time rather than by a hand-written verify block.

Constraint from the ruling that bears on the choice: *modularity is per item type* — a
type's classifier and route table must read as one unit. Both shapes can satisfy it; B
satisfies it with compile-time-checkable config, A with a table row.

## Non-negotiable engine guarantees (what the routers kept re-arguing; make them engine behaviour)

1. **Park-in-place holds the dedup key** (never checkpoint-and-complete for a human class —
   `checkpoint_for_review` writes no item_key and hardcodes an unregistered handler).
2. **Conversions born at the status the promoter contract expects** — `detected` now that
   SCH-026 (`detected-item-promoter`) is live; the engine should not need to know the
   promoter's state.
3. **Conversion item_keys are stable across failed repair cycles** (410's component-keyed
   rule) so the two-strike brake works.
4. **Unknown/malformed route fails LOUDLY** (never closes an item nothing acted on).
5. **Close-evidence goes to the orchestration trail; park-evidence goes on the row** — the
   loop's `mark_complete` REPLACES `result` on completed rows (LANDMINES 2026-08-15).
6. **One active definition row per engine agent** (the two-active-rows landmine).
7. **A verifier registered later for a routed type fail-closes the converted arm** — the
   engine should surface, at seed/validation time, when a routed type has a verifier.
8. **The type-count itself is a signal** — ~~RFC_022's unbuilt accumulation counter applies:~~
   the engine should be able to say how many types it routes, so the estate notices growth.

   > **⚠ CORRECTED 2026-08-19 — "unbuilt" IS FALSE, and it was already false when this was
   > written.** `RFC_022` is **CLOSED** and the whole mechanism is **LIVE**, not hypothetical:
   > the counter was built **2026-08-13** (`cmd/config-key-audit --optional-key-budget`, and
   > `scripts/audit-optional-key-budget.sh [--json] [N]`; concept register **WFA-013**); the owner
   > **ruled N = 10 on 2026-08-14**; and the automatic half runs daily (`50 6 * * *` UTC, CronJob
   > `optional-key-budget-check`, live since 2026-08-14). It writes **one `doc_notes` row per run,
   > including on clean results** — so a MISSING row means the job did not run and must **not** be
   > read as "nothing is wrong". An action past N owes one review of its accumulated surface,
   > recorded in `architecture_review/optional_key_budget_acks.json`.
   >
   > **Why this bears on the A-vs-B choice and is not just a citation fix.** Written as "unbuilt",
   > guarantee 8 reads as *"the engine should volunteer a count nobody yet consumes"* — a
   > nice-to-have, and cheap to satisfy in either shape. In fact there is a **live budget with a
   > ruled threshold and a daily job already enforcing it**, so the question becomes a real design
   > constraint: does the chosen shape make each routed type accumulate **optional keys on one
   > shared action** — in which case the engine walks toward N = 10 as it succeeds, and the budget
   > is a designed-in ceiling to argue about — or does it keep per-type surface off that action?
   > **[UNVERIFIED] which of A or B has that property is exactly what the design round must
   > establish; do not assume either way here.**
   >
   > ⚠ One trap to carry into the round: two actions (`retract_asset_files`, `publish_site`)
   > entered the registry **counted as ZERO** and were invisible to the check until 2026-08-17,
   > because the cron's literal is hand-maintained. The parity test
   > (`cmd/config-key-audit/optional_budget_cron_parity_test.go`) catches it — **run it** — and after
   > editing `check.py` the kustomize overlay must be re-applied or the cluster keeps the old
   > literal. An engine that adds routed types over time is precisely the shape that goes stale here.
   >
   > Source: `CLAUDE.md` § RFC_022 (itself corrected 2026-08-17 after telling every session the
   > opposite for three days) and this lane's `NOTES_router_engine.md`, which flagged the staleness
   > on 2026-08-18 without fixing it here.

## Phasing

1. Read the three seeds + CQ-023 + IMG-071 + this file; measure the live population per type.
2. Design round: A vs B, submitted to the council as an RFC-shaped design (this IS
   architecture scope — a new shared mechanism); record the verdict in RFC_030.
3. Build the engine; migrate **410 first** (its 8 routes define the contract; its 44-item
   history is the regression fixture — the census + canary evidence in
   `bugfix_277_required_fields_repair/`), then 397's two.
4. Retire the three bespoke seeds (rollback files exist for all three); update CQ-023 and
   IMG-071 to point at the engine; register the engine.
5. Standing five here as you go; RFC_030 status → DELIVERED when the third migration is live.

## Out of scope

Deciding what any specific type's classifier should DO — that stays with each type's lane.
The engine runs classifiers; it does not write them.
