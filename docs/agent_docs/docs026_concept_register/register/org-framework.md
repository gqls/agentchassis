# Register — org-framework

1 concept, consolidated from 2 raw extractions (1 unique block, present twice
in the source cluster file due to mechanical duplication in the input), from
unit U21.

### ORG-001 — Organizational framework (roles, listeners, policy-as-filters)
- **status:** abandoned
- **status-evidence:** Extended thought experiment across docs006/005, /006, /006c ("Acme Corp", "Sarah, Marketing Content Writer"); "Open Items for Later" never picked up; no later doc builds on roles/listeners.
- **what:** A design showing the framework is domain-agnostic by modelling a whole company: roles typed as identity/function/composite/position (only identity roles get schemas, like client_X); employees as clients with personal agent_instances; always-on shared listeners (like adapters) that spawn discrete orchestrations per task ("Sarah isn't running, she's ready"); authority as conditional filters (policy-owner agents like legal-review-agent injected by trigger conditions) rather than hierarchy; strategy flowing down as intake, decomposing at each level. Cross-cutting agents concluded to be "just agents many workflows call."
- **sources:** docs006_workflow_builder/005_acme_corp_org_chart.md; docs006_workflow_builder/006_conclude_role_entity_strategy.md#Role-vs-Agent; docs006_workflow_builder/006c_org_framework_discussion.md
- **relations:** entity_state_log (agent-memory-and-evolution register); relationships; policy filters prefigure legal-content-agent constraints; "strategy as intake" prefigures autonomous mission decomposition
- **verify-later:** roles/role_assignments tables (expected absent)
