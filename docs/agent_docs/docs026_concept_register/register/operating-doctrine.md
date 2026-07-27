# Register — operating-doctrine

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

2 concepts, consolidated from 4 raw extractions across unit U14 (each of the 2 distinct blocks
appeared byte-identically twice within the cluster input file — treated as duplicate copies of
one extraction, not independent corroboration).

### OPD-001 — Standing evidence rules (the working-method contract)
- **status:** deployed
- **status-evidence:** code_retrieval_route(21) header "Standing rules: user runs all SQL/kubectl/builds; read outcomes by correlation_id only; snapshot_agent before agent_definitions UPDATEs; schema before SQL; a 0-rows result is not decisive until the query itself is checked."
- **what:** The recurring operating contract of every runbook in the docs019 family: the human runs all mutations/builds; outcomes are read by correlation_id, never ORDER BY … LIMIT 1 (twice a red herring); \d <table> before every query; every agent_definitions change snapshots the row first (snapshot_agent = byte-exact revert path); a 0-rows result proves nothing until the query/selector is validated (wrong key, wrong label, wrong nesting all produced false zeros); migrations are self-guarded (UPDATE 0 = assumption wrong, nothing changed) and carry REVERT blocks.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#header; docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_gamesdesign_index_rebuild.md#7 (0-rows reminder)
- **relations:** instrument skepticism; repo-label bug (LIMIT-1 lesson); diagnostician seed→fix (snapshot rule); operator discipline: verify-by-artifact (operator-practice register, OPP-002)
- **verify-later:** snapshot_agent function; snapshots table growth

### OPD-002 — Parallel-thread boundary and handoff convention
- **status:** convention
- **status-evidence:** builder_route(21) "THREAD HANDED OFF 2026-07-06 → HANDOFF_builder_thread.md"; §B5 "BOUNDARY (adopted): the other chat owns everything INSIDE the tool pipeline …; This chat owns the RELAY …; The §B5 interface … is a JOINT decision, not taken unilaterally"; fix_loop "RULE: any fix-loop change to diagnose workflows is fetch-first against the CURRENT JSON and coordinated".
- **stage2-verified (2026-07-14):** deployed → convention — Parallel-thread boundary/handoff convention (builder/tools/quality/fix-loop own declared surfaces) — a documented working-agreement, not code/infra. Only doc quotes cited (builder_route handoff banner, §B5); no code artifact claimed.
- **what:** Multiple concurrent working threads (builder, tools, quality, fix-loop, imagery) each own declared surfaces; runbooks record explicit boundaries, joint-decision seams, collision surfaces, and fetch-first rules for shared state; work moves between threads via handoff documents and "this item retains / that thread owns" dispositions. This is how the runbook families themselves relate — each family is one thread's travelling state.
- **sources:** docs019/RUNBOOK_builder_route(21).md#handoff-banner; docs019/RUNBOOK_builder_route(21).md#B5; docs019/RUNBOOK_diagnosis_fix_loop(9).md#boundaries; docs019/RUNBOOK_site_quality(1).md#boundaries
- **relations:** doc_plans/doc_notes; per-task notes; documentation-system travelling docs
- **verify-later:** HANDOFF_builder_thread.md; boundary sections across sibling runbooks
