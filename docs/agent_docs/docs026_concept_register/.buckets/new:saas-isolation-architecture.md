
<!-- SOURCE: U13_docs024_small_dirs.md -->
### Isolated chat/satellite architecture ("Y-copy") for SaaS build isolation
- **category:** NEW:saas-isolation-architecture
- **status-signal:** aspirational
- **status-evidence:** "Current lean (open): Option Y-copy... not committed — kept open" (PLAN_isolated_chat_environment(5).md §5); "Nothing in this plan has been deployed or applied yet" (stripe/001commentary.md worked-example thread)
- **what:** A plan to run the site chatbot's server-side pieces (turn storage, drain/analytics, chat workflow code) on infrastructure separate from the core build cluster, decomposing "don't let chat interfere with build" into three distinct threats — load (turn-ingestion write-load competing with builds), hack (a compromised internet-facing edge worker reaching core data), bug (chat code faults degrading shared chassis/Kafka/DB) — via a strictly one-directional, async, egress-from-core-only boundary. Two sizing options were weighed: minimal Option X (turn store + puller + analytics only) vs full Option Y (a cut-down copy of the whole chassis on a separate cluster); the "current lean" is Y-copy (deploy the existing monolithic chassis image against new Kafka/Postgres/storage, curate the agent_definitions seed) as an experimentation sandbox. The plan escalated once chat was reframed as the intake front-end to a full build-as-a-service product: an anonymous, internet-triggered, token-spending build pipeline must not run on core, which rules out minimal Option X and pushes toward full-chassis Option Y as a second, customer-facing instance of the whole platform.
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#1-5,9,11-13, tools/tool_widget_clobber/PLAN_isolated_chat_environment(2).md#13, stripe/001commentary.md#worked-example section, stripe/001commentary.md#§13
- **relations:** Multi-cluster dispatch (Phase 4a) — the coupled model explicitly NOT reused; Agent-to-adapter capability maturation path; Conversational build-intake via briefing-agent chat; Operator-vs-vendor business model fork; Entitlement gate architecture
- **verify-later:** whether any satellite infrastructure has actually been stood up; `086_site_chat_turns.sql` isolated-DB variant; `TurnSink` implementation

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Conversational build-intake via briefing-agent chat
- **category:** NEW:saas-isolation-architecture
- **status-signal:** aspirational
- **status-evidence:** Described purely as a designed flow: "instead of a static briefing form, a briefing-agent conducts the intake as a conversation"
- **what:** Reframes an existing chatbot feature as the entry point to the whole build pipeline: a customer types a domain + rough spec into a chat, a `briefing-agent` conducts the intake as dialogue rather than a static form, then hands off to `intake-orchestrator` on the satellite to create the site row and kick the build. Reuses the existing build pipeline unchanged once the brief is solid; the chat drops into a "job lane" for the build duration.
- **sources:** stripe/001commentary.md#worked example section
- **relations:** Isolated chat/satellite architecture (Y-copy); New-domain build pipeline stage chain
- **verify-later:** 018_briefing_questionnaire fields; 002_intake_orchestrator entry contract

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Isolated chat/satellite architecture ("Y-copy") for SaaS build isolation
- **category:** NEW:saas-isolation-architecture
- **status-signal:** aspirational
- **status-evidence:** "Current lean (open): Option Y-copy... not committed — kept open" (PLAN_isolated_chat_environment(5).md §5); "Nothing in this plan has been deployed or applied yet" (stripe/001commentary.md worked-example thread)
- **what:** A plan to run the site chatbot's server-side pieces (turn storage, drain/analytics, chat workflow code) on infrastructure separate from the core build cluster, decomposing "don't let chat interfere with build" into three distinct threats — load (turn-ingestion write-load competing with builds), hack (a compromised internet-facing edge worker reaching core data), bug (chat code faults degrading shared chassis/Kafka/DB) — via a strictly one-directional, async, egress-from-core-only boundary. Two sizing options were weighed: minimal Option X (turn store + puller + analytics only) vs full Option Y (a cut-down copy of the whole chassis on a separate cluster); the "current lean" is Y-copy (deploy the existing monolithic chassis image against new Kafka/Postgres/storage, curate the agent_definitions seed) as an experimentation sandbox. The plan escalated once chat was reframed as the intake front-end to a full build-as-a-service product: an anonymous, internet-triggered, token-spending build pipeline must not run on core, which rules out minimal Option X and pushes toward full-chassis Option Y as a second, customer-facing instance of the whole platform.
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#1-5,9,11-13, tools/tool_widget_clobber/PLAN_isolated_chat_environment(2).md#13, stripe/001commentary.md#worked-example section, stripe/001commentary.md#§13
- **relations:** Multi-cluster dispatch (Phase 4a) — the coupled model explicitly NOT reused; Agent-to-adapter capability maturation path; Conversational build-intake via briefing-agent chat; Operator-vs-vendor business model fork; Entitlement gate architecture
- **verify-later:** whether any satellite infrastructure has actually been stood up; `086_site_chat_turns.sql` isolated-DB variant; `TurnSink` implementation

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Conversational build-intake via briefing-agent chat
- **category:** NEW:saas-isolation-architecture
- **status-signal:** aspirational
- **status-evidence:** Described purely as a designed flow: "instead of a static briefing form, a briefing-agent conducts the intake as a conversation"
- **what:** Reframes an existing chatbot feature as the entry point to the whole build pipeline: a customer types a domain + rough spec into a chat, a `briefing-agent` conducts the intake as dialogue rather than a static form, then hands off to `intake-orchestrator` on the satellite to create the site row and kick the build. Reuses the existing build pipeline unchanged once the brief is solid; the chat drops into a "job lane" for the build duration.
- **sources:** stripe/001commentary.md#worked example section
- **relations:** Isolated chat/satellite architecture (Y-copy); New-domain build pipeline stage chain
- **verify-later:** 018_briefing_questionnaire fields; 002_intake_orchestrator entry contract
