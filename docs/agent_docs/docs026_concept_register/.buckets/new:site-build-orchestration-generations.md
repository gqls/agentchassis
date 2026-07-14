
<!-- SOURCE: U15_docs019_running_notes.md -->
### Work-item relay / three-generation builder architecture
- **category:** NEW:site-build-orchestration-generations
- **status-signal:** partial
- **status-evidence:** "THREE generations coexist... GEN-3 component/spec/DB era = pageflow-builder v20 ACTIVE + site-work-orchestrator (queue-native sibling)... §B3 CLOSED: spine = the work-item relay" (NOTES_running_synthesis_v4(39).md, 2026-07-04).
- **what:** A builder-thread inventory found three coexisting generations of "build a site" orchestration on the platform (GEN-1 template era; GEN-2 in-memory multipage v1≈v2; GEN-3 component/spec/DB era) with ~8 overlapping top-level "build the site" orchestrators, only one of which (`pageflow-builder`) is the active monolith. Separately, a queue-native work-item relay (`domain-submitter → needs_domain_research → build-dispatch-loop → domain-research-classifier → needs_strategy → domain-strategist → needs_briefing → build-briefing-agent → needs_site_plan → build-site-planner → needs_page/needs_content_page → page-build-handler`) was traced end-to-end via `reconcile_site_plan`'s routing table and confirmed to reach the builder NATIVELY — established as the real spine, with `pageflow-builder` demoted to "intake convenience." A commented-out `"tool"` route in the same routing table is the mechanism gap blocking tool/infographics pages from the relay.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 "§B0" through "§B3 CLOSED" entries.
- **relations:** Roadmap-phase enforcement gap; adoption pipeline; vertical-exemplar-researcher hop; site-quality programme handoff.
- **verify-later:** `RUNBOOK_builder_route.md`, `load_work_item_actions.go` routing table, the un-consolidated GEN-1/2 legacy orchestrators (Q1/Q5 consolidation candidates, left open).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Work-item relay / three-generation builder architecture
- **category:** NEW:site-build-orchestration-generations
- **status-signal:** partial
- **status-evidence:** "THREE generations coexist... GEN-3 component/spec/DB era = pageflow-builder v20 ACTIVE + site-work-orchestrator (queue-native sibling)... §B3 CLOSED: spine = the work-item relay" (NOTES_running_synthesis_v4(39).md, 2026-07-04).
- **what:** A builder-thread inventory found three coexisting generations of "build a site" orchestration on the platform (GEN-1 template era; GEN-2 in-memory multipage v1≈v2; GEN-3 component/spec/DB era) with ~8 overlapping top-level "build the site" orchestrators, only one of which (`pageflow-builder`) is the active monolith. Separately, a queue-native work-item relay (`domain-submitter → needs_domain_research → build-dispatch-loop → domain-research-classifier → needs_strategy → domain-strategist → needs_briefing → build-briefing-agent → needs_site_plan → build-site-planner → needs_page/needs_content_page → page-build-handler`) was traced end-to-end via `reconcile_site_plan`'s routing table and confirmed to reach the builder NATIVELY — established as the real spine, with `pageflow-builder` demoted to "intake convenience." A commented-out `"tool"` route in the same routing table is the mechanism gap blocking tool/infographics pages from the relay.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 "§B0" through "§B3 CLOSED" entries.
- **relations:** Roadmap-phase enforcement gap; adoption pipeline; vertical-exemplar-researcher hop; site-quality programme handoff.
- **verify-later:** `RUNBOOK_builder_route.md`, `load_work_item_actions.go` routing table, the un-consolidated GEN-1/2 legacy orchestrators (Q1/Q5 consolidation candidates, left open).
