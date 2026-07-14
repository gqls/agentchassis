
<!-- SOURCE: U19_sql_tables_components.md -->
### site_chat_turns per-domain chatbot logging
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Migration drafted with "NOTE ON NUMBERING: this snapshot only shows migrations up to 085. Confirm the next free migration number... before applying" — written against a snapshot, application unconfirmed in this unit.
- **what:** End-user chatbot turns from the site chatbot edge worker: one row per prompt/answer (PII), populated by a Layer-1 puller draining the edge sink with idempotent ingest via edge-supplied uuid PK; bounding outcomes (refused off-topic, capped), provenance for "why did it say that" (model, context pack_version, grounding_ids chunk list), token/latency columns name-aligned to llm_call_log, GDPR-conscious salted client_ip_hash instead of raw IPs, edge vs ingest timestamps, per-site cascade delete. Explicitly distinct from llm_call_log (build-time flywheel vs end-user data with its own retention/access profile).
- **sources:** docs/agent_docs/sql_for_tables/046_site_chat_turns.sql
- **relations:** llm_call_log; rag-retrieval (context packs / grounding chunks); edge workers.
- **verify-later:** table existence in production; edge worker + Layer-1 puller implementations.

<!-- SOURCE: U22_recent_small_docs.md -->
### Site chatbot edge worker (synchronous, not an orchestrated agent)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Design doc with "Suggested build order (structural first)" and "Open decisions" — canonical design, nothing deployed.
- **what:** The canonical design for a per-domain chatbot on static-S3 sites: a synchronous request/response handler on a provider-agnostic serverless edge worker (Cloudflare first), NOT run through Kafka/the chassis. Deliberate documented exception to "every agent is an orchestrator" — Kafka's async failure modes (offset replay, phantom-complete, no streaming) are wrong for live chat, and a central nginx VM would drag static traffic behind a hackable box and lose S3's hack-resistance. Worker: resolve domain → load context pack → guard limits → compose bounded prompt → stream LLM tokens (SSE) → fire-and-forget record turn.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md
- **relations:** context pack, provider-agnostic deps adapters, site_chat_turns, isolated chat environment
- **verify-later:** any edge worker deploy; /api/chat route registration

<!-- SOURCE: U22_recent_small_docs.md -->
### Build-time context pack (per-domain bounded context)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 7 defines the JSON shape and versioning; produced by an unbuilt `chat-context-builder` agent.
- **what:** One per-domain JSON document published to static storage at install time, carrying identity, scope (instructions/refusal message/banned topics), build-time-selected grounding chunks (bounded by token budget), suggested model, and operational limits. The worker holds no per-site logic — the pack is the entire bounded context. Grounding is selected on Layer 1 via Ollama embeddings + pgvector; v2 optionally ships chunk vectors for in-worker per-question retrieval plus a narrow stateless embedding endpoint.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#7
- **relations:** chat-context-builder, RAG knowledge_base (install-time reuse), three-layer bounding
- **verify-later:** context-pack schema; chat-context-builder agent; pack publish-to-S3 step

<!-- SOURCE: U22_recent_small_docs.md -->
### site-chat-installer orchestration (install_chat maintenance task)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Resolved: Install is a separate orchestration, triggered via a maintenance_queue install_chat task — not a build-pipeline stage." Not built.
- **what:** Chat install is its own orchestration (triggered by a `maintenance_queue` `install_chat` task, build pipeline untouched), spawning three sub-agents: `chat-context-builder` (build+publish the pack via Ollama+pgvector), `chat-widget-installer` (fork the chat widget through the existing component/tool pipeline; only difference is it POSTs to /api/chat), and `chat-route-registrar` (record the route + mark chat installed on the site, reversible via uninstall_chat). Supersedes the older `chat-suggester` gating agent from the FOCUS base version.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack.md (delta: chat-suggester)
- **relations:** maintenance_queue, context pack, component/tool pipeline, chat-suggester (superseded)
- **verify-later:** site-chat-installer + sub-agent definitions; install_chat maintenance task_type

<!-- SOURCE: U22_recent_small_docs.md -->
### Provider-agnostic worker (deps adapters)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 5 defines interfaces and a Cloudflare shim; "Best practice" reference impls listed, no code shipped.
- **what:** The portability strategy: a Web-platform-only core `handleChat(request, deps)` plus a ~20-line per-platform shim. Three (v2: four) small adapters — ContextStore (HTTP GET of static pack), LLMClient (Anthropic Messages over fetch, swappable to self-hosted), TurnSink (queue/D1, fire-and-forget), and v2 Embedder — each with a Cloudflare and a portable HTTP impl. Nothing vendor-specific in the core; Cloudflare/Deno/Fastly/Vercel/self-host are drop-in. Rate limiting is the least-portable concern (WAF + in-pack per-session cap floor).
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#5, #6
- **relations:** edge worker, context pack, pluggable billing/LLM/storage adapter discipline
- **verify-later:** handleChat core + adapter interfaces if implemented

<!-- SOURCE: U22_recent_small_docs.md -->
### site_chat_turns table (turn recording)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Migration 086 written and "schema-checked" against live `sites`, but header notes "this snapshot only shows migrations up to 085. Confirm the next free migration number ... before applying."
- **what:** A `site_chat_turns` table logging each end-user prompt/answer turn per domain (question/answer as PII, refused/capped flags, model, pack_version, grounding_ids, tokens/latency named to match llm_call_log, salted client_ip_hash never raw IP). Deliberately separate from the build-time `llm_call_log` (different owner, privacy profile, and access pattern). Edge-supplied turn uuid is the PK for idempotent ingest (ON CONFLICT DO NOTHING); populated by a Layer-1 puller draining the edge sink.
- **sources:** docs025.../086_site_chat_turns.sql, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#8
- **relations:** llm_call_log (kept separate), TurnSink, isolated chat environment (isolated-DB variant drops the FK)
- **verify-later:** site_chat_turns migration number; Layer-1 turn puller

<!-- SOURCE: U22_recent_small_docs.md -->
### Three-layer bounding (retrieval / prompt / operational)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 9 "Bounded context — the three layers"; part of the design.
- **what:** A precise decomposition of "bounded" to stop chatbot drift: retrieval bounding (only this site's grounding is in the pack, frozen at build time), prompt bounding (system prompt pins identity and emits an exact refusal message for out-of-scope questions), and operational bounding (input length, output tokens, turns/session, history window, rate limiting from pack.limits). Conflating the three is where bots go off-topic.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#9
- **relations:** context pack, edge worker composeSystemPrompt
- **verify-later:** composeSystemPrompt refusal enforcement; pack.limits guards

<!-- SOURCE: U22_recent_small_docs.md -->
### Isolated chat environment (satellite; load/hack/bug vectors)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Current lean (not committed — kept open): Option Y-copy ... experiment in a sandbox." Explicitly undecided.
- **what:** A plan to run the chatbot's server-side pieces (turn store, drain, analytics, optionally chat workflow code) on infrastructure separate from the core build cluster, severing three blast vectors — load (turn write-load), hack (compromised edge worker's reachable radius), bug (chat code faulting the shared chassis). Deliberately does NOT reuse the coupled multi-cluster dispatch (which shares core Kafka/Postgres). Option X = minimal satellite (maybe no chassis at all); Option Y = full cut-down chassis (Y-copy config-only vs Y-slim purpose-built image). Boundary is one-directional, async, egress-from-core only.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md, docs025.../PLAN_isolated_chat_environment(1).md
- **relations:** remote-job-spawner (NOT reused), site_chat_turns, boundary contract, building-as-a-service
- **verify-later:** any separate chat cluster/DB; isolated-DB variant of migration 086

<!-- SOURCE: U22_recent_small_docs.md -->
### Simple paid multi-domain chat (freemium + day-pass)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Status: discussion draft — direction firming."
- **what:** A deliberately simple "fast lane" route: the FOCUS edge worker + a light paywall, multi-tenant-by-Host, add a domain by publishing config + DNS (no chassis/Kafka/satellite). Monetisation is freemium + a flat day-pass (£2-5) rather than counted credits, because card processing's fixed ~20-30p fee makes sub-£5 one-off charges poor. Entitlement is a stateless signed `{domain, expiry}` token issued via a synchronous Stripe guest-checkout `redeem` (no accounts, no webhook on the critical path, no edge KV). The real cost driver is the free taster + abuse, not paying users.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md, docs025.../PLAN_simple_paid_multidomain_chat.md
- **relations:** edge worker, context pack, chat lanes (fast lane), commercial model/billing adapter
- **verify-later:** paywall gate + redeem endpoint; day-pass token signing/validation

<!-- SOURCE: U22_recent_small_docs.md -->
### Chat lanes (fast/slow/job) + warm-adapter maturation
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 11 "What the agent framework buys chat (lanes and maturation)"; "This is still open by design; needs further conversation."
- **what:** A model splitting chat by what it does: fast lane (bounded Q&A, synchronous/streamed, no framework — ships independently); slow lane (turns needing work — live research, structured-data queries, running a site's tool, in-answer charts, multi-step tasks — routed by a cheap intent classifier, user warned it's slower); job lane (long-running submissions like "build me a site", ack + status + deliver). Maturation path: prove a slow-lane capability as a spawned agent (~12s cold), promote popular ones to warm adapters, end-state a resident chat-orchestrator adapter that fans out without spawning per turn.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#11
- **relations:** simple paid multi-domain chat (fast lane), building-as-a-service (job lane), warm adapters
- **verify-later:** intent classifier; any resident chat-orchestrator adapter

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### isolated chat satellite architecture (three blast vectors: load/hack/bug)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Current lean (not committed — kept open)." Explicitly a plan document with an open central decision (Option X vs Y).
- **what:** A plan to run the site-chatbot's server-side pieces (turn storage, drain/analytics, and any chat workflow code) on infrastructure **separate** from the core build cluster, so that live chat traffic, a compromise of the internet-facing edge worker, or a chat-code bug cannot degrade or reach the webdesign/build system. Deliberately **not** built on the existing multi-cluster dispatch (Phase 4a, `remote-job-spawner`), which shares cluster A's Kafka/Postgres by design — the chat satellite instead reuses only the chassis *binaries* and action code, deployed against its own Kafka/DB/storage, with a one-directional async boundary (core publishes install triggers and content; nothing on the chat side has synchronous or write access back into core). Two options are weighed: **Option X (minimal, recommended MVP)** — pack-building stays on core, the satellite is just a turn store + puller + analytics, possibly needing no Kafka/chassis at all; **Option Y (full satellite chassis)** — the whole chat pipeline including install/pack-building moves to a cut-down copy of the chassis on the satellite. A worked "building-and-hosting-as-a-service via chat" example (a customer types a domain into another site's chatbot and gets a fully built, hosted site with its own chatbot) reframes the satellite as a second, customer-facing instance of the whole platform and pushes the design toward Option Y for that specific use case.
- **sources:** docs/_archive/agent_docs/docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(3).md
- **relations:** multicluster (Phase 4a, the pattern explicitly rejected as a template); SaaS commercial model (below, same document, §13)
- **verify-later:** whether any satellite cluster / separate chat Postgres exists; site_chat_turns table; remote-job-spawner (the Phase 4a mechanism used as contrast)

<!-- SOURCE: U19_sql_tables_components.md -->
### site_chat_turns per-domain chatbot logging
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Migration drafted with "NOTE ON NUMBERING: this snapshot only shows migrations up to 085. Confirm the next free migration number... before applying" — written against a snapshot, application unconfirmed in this unit.
- **what:** End-user chatbot turns from the site chatbot edge worker: one row per prompt/answer (PII), populated by a Layer-1 puller draining the edge sink with idempotent ingest via edge-supplied uuid PK; bounding outcomes (refused off-topic, capped), provenance for "why did it say that" (model, context pack_version, grounding_ids chunk list), token/latency columns name-aligned to llm_call_log, GDPR-conscious salted client_ip_hash instead of raw IPs, edge vs ingest timestamps, per-site cascade delete. Explicitly distinct from llm_call_log (build-time flywheel vs end-user data with its own retention/access profile).
- **sources:** docs/agent_docs/sql_for_tables/046_site_chat_turns.sql
- **relations:** llm_call_log; rag-retrieval (context packs / grounding chunks); edge workers.
- **verify-later:** table existence in production; edge worker + Layer-1 puller implementations.

<!-- SOURCE: U22_recent_small_docs.md -->
### Site chatbot edge worker (synchronous, not an orchestrated agent)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Design doc with "Suggested build order (structural first)" and "Open decisions" — canonical design, nothing deployed.
- **what:** The canonical design for a per-domain chatbot on static-S3 sites: a synchronous request/response handler on a provider-agnostic serverless edge worker (Cloudflare first), NOT run through Kafka/the chassis. Deliberate documented exception to "every agent is an orchestrator" — Kafka's async failure modes (offset replay, phantom-complete, no streaming) are wrong for live chat, and a central nginx VM would drag static traffic behind a hackable box and lose S3's hack-resistance. Worker: resolve domain → load context pack → guard limits → compose bounded prompt → stream LLM tokens (SSE) → fire-and-forget record turn.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md
- **relations:** context pack, provider-agnostic deps adapters, site_chat_turns, isolated chat environment
- **verify-later:** any edge worker deploy; /api/chat route registration

<!-- SOURCE: U22_recent_small_docs.md -->
### Build-time context pack (per-domain bounded context)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 7 defines the JSON shape and versioning; produced by an unbuilt `chat-context-builder` agent.
- **what:** One per-domain JSON document published to static storage at install time, carrying identity, scope (instructions/refusal message/banned topics), build-time-selected grounding chunks (bounded by token budget), suggested model, and operational limits. The worker holds no per-site logic — the pack is the entire bounded context. Grounding is selected on Layer 1 via Ollama embeddings + pgvector; v2 optionally ships chunk vectors for in-worker per-question retrieval plus a narrow stateless embedding endpoint.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#7
- **relations:** chat-context-builder, RAG knowledge_base (install-time reuse), three-layer bounding
- **verify-later:** context-pack schema; chat-context-builder agent; pack publish-to-S3 step

<!-- SOURCE: U22_recent_small_docs.md -->
### site-chat-installer orchestration (install_chat maintenance task)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Resolved: Install is a separate orchestration, triggered via a maintenance_queue install_chat task — not a build-pipeline stage." Not built.
- **what:** Chat install is its own orchestration (triggered by a `maintenance_queue` `install_chat` task, build pipeline untouched), spawning three sub-agents: `chat-context-builder` (build+publish the pack via Ollama+pgvector), `chat-widget-installer` (fork the chat widget through the existing component/tool pipeline; only difference is it POSTs to /api/chat), and `chat-route-registrar` (record the route + mark chat installed on the site, reversible via uninstall_chat). Supersedes the older `chat-suggester` gating agent from the FOCUS base version.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack.md (delta: chat-suggester)
- **relations:** maintenance_queue, context pack, component/tool pipeline, chat-suggester (superseded)
- **verify-later:** site-chat-installer + sub-agent definitions; install_chat maintenance task_type

<!-- SOURCE: U22_recent_small_docs.md -->
### Provider-agnostic worker (deps adapters)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 5 defines interfaces and a Cloudflare shim; "Best practice" reference impls listed, no code shipped.
- **what:** The portability strategy: a Web-platform-only core `handleChat(request, deps)` plus a ~20-line per-platform shim. Three (v2: four) small adapters — ContextStore (HTTP GET of static pack), LLMClient (Anthropic Messages over fetch, swappable to self-hosted), TurnSink (queue/D1, fire-and-forget), and v2 Embedder — each with a Cloudflare and a portable HTTP impl. Nothing vendor-specific in the core; Cloudflare/Deno/Fastly/Vercel/self-host are drop-in. Rate limiting is the least-portable concern (WAF + in-pack per-session cap floor).
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#5, #6
- **relations:** edge worker, context pack, pluggable billing/LLM/storage adapter discipline
- **verify-later:** handleChat core + adapter interfaces if implemented

<!-- SOURCE: U22_recent_small_docs.md -->
### site_chat_turns table (turn recording)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Migration 086 written and "schema-checked" against live `sites`, but header notes "this snapshot only shows migrations up to 085. Confirm the next free migration number ... before applying."
- **what:** A `site_chat_turns` table logging each end-user prompt/answer turn per domain (question/answer as PII, refused/capped flags, model, pack_version, grounding_ids, tokens/latency named to match llm_call_log, salted client_ip_hash never raw IP). Deliberately separate from the build-time `llm_call_log` (different owner, privacy profile, and access pattern). Edge-supplied turn uuid is the PK for idempotent ingest (ON CONFLICT DO NOTHING); populated by a Layer-1 puller draining the edge sink.
- **sources:** docs025.../086_site_chat_turns.sql, docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#8
- **relations:** llm_call_log (kept separate), TurnSink, isolated chat environment (isolated-DB variant drops the FK)
- **verify-later:** site_chat_turns migration number; Layer-1 turn puller

<!-- SOURCE: U22_recent_small_docs.md -->
### Three-layer bounding (retrieval / prompt / operational)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 9 "Bounded context — the three layers"; part of the design.
- **what:** A precise decomposition of "bounded" to stop chatbot drift: retrieval bounding (only this site's grounding is in the pack, frozen at build time), prompt bounding (system prompt pins identity and emits an exact refusal message for out-of-scope questions), and operational bounding (input length, output tokens, turns/session, history window, rate limiting from pack.limits). Conflating the three is where bots go off-topic.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#9
- **relations:** context pack, edge worker composeSystemPrompt
- **verify-later:** composeSystemPrompt refusal enforcement; pack.limits guards

<!-- SOURCE: U22_recent_small_docs.md -->
### Isolated chat environment (satellite; load/hack/bug vectors)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Current lean (not committed — kept open): Option Y-copy ... experiment in a sandbox." Explicitly undecided.
- **what:** A plan to run the chatbot's server-side pieces (turn store, drain, analytics, optionally chat workflow code) on infrastructure separate from the core build cluster, severing three blast vectors — load (turn write-load), hack (compromised edge worker's reachable radius), bug (chat code faulting the shared chassis). Deliberately does NOT reuse the coupled multi-cluster dispatch (which shares core Kafka/Postgres). Option X = minimal satellite (maybe no chassis at all); Option Y = full cut-down chassis (Y-copy config-only vs Y-slim purpose-built image). Boundary is one-directional, async, egress-from-core only.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md, docs025.../PLAN_isolated_chat_environment(1).md
- **relations:** remote-job-spawner (NOT reused), site_chat_turns, boundary contract, building-as-a-service
- **verify-later:** any separate chat cluster/DB; isolated-DB variant of migration 086

<!-- SOURCE: U22_recent_small_docs.md -->
### Simple paid multi-domain chat (freemium + day-pass)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Status: discussion draft — direction firming."
- **what:** A deliberately simple "fast lane" route: the FOCUS edge worker + a light paywall, multi-tenant-by-Host, add a domain by publishing config + DNS (no chassis/Kafka/satellite). Monetisation is freemium + a flat day-pass (£2-5) rather than counted credits, because card processing's fixed ~20-30p fee makes sub-£5 one-off charges poor. Entitlement is a stateless signed `{domain, expiry}` token issued via a synchronous Stripe guest-checkout `redeem` (no accounts, no webhook on the critical path, no edge KV). The real cost driver is the free taster + abuse, not paying users.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md, docs025.../PLAN_simple_paid_multidomain_chat.md
- **relations:** edge worker, context pack, chat lanes (fast lane), commercial model/billing adapter
- **verify-later:** paywall gate + redeem endpoint; day-pass token signing/validation

<!-- SOURCE: U22_recent_small_docs.md -->
### Chat lanes (fast/slow/job) + warm-adapter maturation
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** Section 11 "What the agent framework buys chat (lanes and maturation)"; "This is still open by design; needs further conversation."
- **what:** A model splitting chat by what it does: fast lane (bounded Q&A, synchronous/streamed, no framework — ships independently); slow lane (turns needing work — live research, structured-data queries, running a site's tool, in-answer charts, multi-step tasks — routed by a cheap intent classifier, user warned it's slower); job lane (long-running submissions like "build me a site", ack + status + deliver). Maturation path: prove a slow-lane capability as a spawned agent (~12s cold), promote popular ones to warm adapters, end-state a resident chat-orchestrator adapter that fans out without spawning per turn.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#11
- **relations:** simple paid multi-domain chat (fast lane), building-as-a-service (job lane), warm adapters
- **verify-later:** intent classifier; any resident chat-orchestrator adapter

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### isolated chat satellite architecture (three blast vectors: load/hack/bug)
- **category:** NEW:site-chatbot
- **status-signal:** aspirational
- **status-evidence:** "Current lean (not committed — kept open)." Explicitly a plan document with an open central decision (Option X vs Y).
- **what:** A plan to run the site-chatbot's server-side pieces (turn storage, drain/analytics, and any chat workflow code) on infrastructure **separate** from the core build cluster, so that live chat traffic, a compromise of the internet-facing edge worker, or a chat-code bug cannot degrade or reach the webdesign/build system. Deliberately **not** built on the existing multi-cluster dispatch (Phase 4a, `remote-job-spawner`), which shares cluster A's Kafka/Postgres by design — the chat satellite instead reuses only the chassis *binaries* and action code, deployed against its own Kafka/DB/storage, with a one-directional async boundary (core publishes install triggers and content; nothing on the chat side has synchronous or write access back into core). Two options are weighed: **Option X (minimal, recommended MVP)** — pack-building stays on core, the satellite is just a turn store + puller + analytics, possibly needing no Kafka/chassis at all; **Option Y (full satellite chassis)** — the whole chat pipeline including install/pack-building moves to a cut-down copy of the chassis on the satellite. A worked "building-and-hosting-as-a-service via chat" example (a customer types a domain into another site's chatbot and gets a fully built, hosted site with its own chatbot) reframes the satellite as a second, customer-facing instance of the whole platform and pushes the design toward Option Y for that specific use case.
- **sources:** docs/_archive/agent_docs/docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(3).md
- **relations:** multicluster (Phase 4a, the pattern explicitly rejected as a template); SaaS commercial model (below, same document, §13)
- **verify-later:** whether any satellite cluster / separate chat Postgres exists; site_chat_turns table; remote-job-spawner (the Phase 4a mechanism used as contrast)
