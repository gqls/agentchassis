# HANDOFF — router engine lane, cold start (created 2026-08-15 by the bugfix_277 session)

**Owner ruling to implement:** RFC_030 — build ONE work-item router engine that runs
per-type classifiers, each item type keeping its own classifier + route table as its own
modular unit; migrate the three existing bespoke routers onto it (410 first, then 397's two).
Ruling text and the two design constraints it adds: `architecture_review/RFC_030_….md`
(status block at the top).

## Read first, in this order
1. `PLAN_2026-08-15_router_engine.md` here — the two candidate shapes (A config-driven agent,
   B Go action), the eight non-negotiable engine guarantees, and the phasing.
2. `sql_for_agents/410_required_fields_missing_router.sql` — the hardened router; its header
   is the best single account of WHY each guarantee exists (four council rounds' worth).
3. `sql_for_agents/397_image_flag_only_routers.sql` — the two earlier routers (simpler shape,
   none of 410's hardening — they are what migrating onto the engine will FIX).
4. Register: CQ-023 (410), IMG-071 (397's two), SCH-026 (the promoter — conversions should
   be born `detected` now), RFC_022 (accumulation counter — the engine should count its types).
5. `bugfix_277_required_fields_repair/` — the census + five-canary evidence for 410; reuse it
   as the regression fixture when migrating 410 onto the engine.

## Session-start checklist
- `git log --oneline -10`; `scripts/who-owns.py 277` and grep live `.jsonl` transcripts for
  `router_engine` / `RFC_030` — this lane was CREATED but not started; check nobody else picked
  it up.
- Re-measure: `SELECT type FROM agent_definitions WHERE type IN ('image-url-404-handler',
  'image-source-unsatisfiable-handler','required-fields-missing-handler') AND is_active …` and
  the open item counts per routed type — the numbers in the 277 docs are 2026-08-15's.
- The design choice (A vs B) goes to the council as an RFC-shaped submission BEFORE building —
  this is a shared mechanism (architecture scope, 2026-07-28 ruling); one round, then build.

## What is NOT this lane's
- What any type's classifier decides — each type's lane owns that.
- The auto-fix half of `improvement-sweep` (still paused, IMP-016) — separate owner decision.
