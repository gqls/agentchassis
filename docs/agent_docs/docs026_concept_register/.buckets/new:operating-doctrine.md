
<!-- SOURCE: U14_docs019_runbooks.md -->
### Standing evidence rules (the working-method contract)
- **category:** NEW:operating-doctrine
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) header "Standing rules: user runs all SQL/kubectl/builds; read outcomes by correlation_id only; snapshot_agent before agent_definitions UPDATEs; schema before SQL; a 0-rows result is not decisive until the query itself is checked."
- **what:** The recurring operating contract of every runbook in this unit: the human runs all mutations/builds; outcomes are read by correlation_id, never `ORDER BY … LIMIT 1` (twice a red herring); `\d <table>` before every query; every agent_definitions change snapshots the row first (`snapshot_agent` = byte-exact revert path); a 0-rows result proves nothing until the query/selector is validated (wrong key, wrong label, wrong nesting all produced false zeros); migrations are self-guarded (UPDATE 0 = assumption wrong, nothing changed) and carry REVERT blocks.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#header; docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_gamesdesign_index_rebuild.md#7 (0-rows reminder)
- **relations:** instrument skepticism; repo-label bug (LIMIT-1 lesson); diagnostician seed→fix (snapshot rule)
- **verify-later:** snapshot_agent function; snapshots table growth

<!-- SOURCE: U14_docs019_runbooks.md -->
### Parallel-thread boundary and handoff convention
- **category:** NEW:operating-doctrine
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "THREAD HANDED OFF 2026-07-06 → HANDOFF_builder_thread.md"; §B5 "BOUNDARY (adopted): the other chat owns everything INSIDE the tool pipeline …; This chat owns the RELAY …; The §B5 interface … is a JOINT decision, not taken unilaterally"; fix_loop "RULE: any fix-loop change to diagnose workflows is fetch-first against the CURRENT JSON and coordinated".
- **what:** Multiple concurrent working threads (builder, tools, quality, fix-loop, imagery) each own declared surfaces; runbooks record explicit boundaries, joint-decision seams, collision surfaces, and fetch-first rules for shared state; work moves between threads via handoff documents and "this item retains / that thread owns" dispositions. This is how the runbook families themselves relate — each family is one thread's travelling state.
- **sources:** docs019/RUNBOOK_builder_route(21).md#handoff-banner; docs019/RUNBOOK_builder_route(21).md#B5; docs019/RUNBOOK_diagnosis_fix_loop(9).md#boundaries; docs019/RUNBOOK_site_quality(1).md#boundaries
- **relations:** doc_plans/doc_notes; per-task notes; documentation-system travelling docs
- **verify-later:** HANDOFF_builder_thread.md; boundary sections across sibling runbooks

<!-- SOURCE: U14_docs019_runbooks.md -->
### Standing evidence rules (the working-method contract)
- **category:** NEW:operating-doctrine
- **status-signal:** deployed
- **status-evidence:** code_retrieval_route(21) header "Standing rules: user runs all SQL/kubectl/builds; read outcomes by correlation_id only; snapshot_agent before agent_definitions UPDATEs; schema before SQL; a 0-rows result is not decisive until the query itself is checked."
- **what:** The recurring operating contract of every runbook in this unit: the human runs all mutations/builds; outcomes are read by correlation_id, never `ORDER BY … LIMIT 1` (twice a red herring); `\d <table>` before every query; every agent_definitions change snapshots the row first (`snapshot_agent` = byte-exact revert path); a 0-rows result proves nothing until the query/selector is validated (wrong key, wrong label, wrong nesting all produced false zeros); migrations are self-guarded (UPDATE 0 = assumption wrong, nothing changed) and carry REVERT blocks.
- **sources:** docs019/RUNBOOK_code_retrieval_route(21).md#header; docs019/RUNBOOK(31)_diagnosis_loop.md#6C; docs019/RUNBOOK_gamesdesign_index_rebuild.md#7 (0-rows reminder)
- **relations:** instrument skepticism; repo-label bug (LIMIT-1 lesson); diagnostician seed→fix (snapshot rule)
- **verify-later:** snapshot_agent function; snapshots table growth

<!-- SOURCE: U14_docs019_runbooks.md -->
### Parallel-thread boundary and handoff convention
- **category:** NEW:operating-doctrine
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "THREAD HANDED OFF 2026-07-06 → HANDOFF_builder_thread.md"; §B5 "BOUNDARY (adopted): the other chat owns everything INSIDE the tool pipeline …; This chat owns the RELAY …; The §B5 interface … is a JOINT decision, not taken unilaterally"; fix_loop "RULE: any fix-loop change to diagnose workflows is fetch-first against the CURRENT JSON and coordinated".
- **what:** Multiple concurrent working threads (builder, tools, quality, fix-loop, imagery) each own declared surfaces; runbooks record explicit boundaries, joint-decision seams, collision surfaces, and fetch-first rules for shared state; work moves between threads via handoff documents and "this item retains / that thread owns" dispositions. This is how the runbook families themselves relate — each family is one thread's travelling state.
- **sources:** docs019/RUNBOOK_builder_route(21).md#handoff-banner; docs019/RUNBOOK_builder_route(21).md#B5; docs019/RUNBOOK_diagnosis_fix_loop(9).md#boundaries; docs019/RUNBOOK_site_quality(1).md#boundaries
- **relations:** doc_plans/doc_notes; per-task notes; documentation-system travelling docs
- **verify-later:** HANDOFF_builder_thread.md; boundary sections across sibling runbooks
