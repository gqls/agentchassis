# Register — site-chatbot

9 concepts, consolidated from 22 raw extractions (11 unique blocks, each
appearing twice due to exact whole-block duplication in the cluster input file)
across units U19, U22, U24f.

### CHAT-001 — site_chat_turns table (per-domain chatbot turn logging)
- **status:** aspirational
- **status-evidence:** Two extraction units cite two different migration numbers for what is otherwise the same table — `046_site_chat_turns.sql` (U19) and `086_site_chat_turns.sql` (U22) — both carrying the same "NOTE ON NUMBERING: this snapshot only shows migrations up to N; confirm the next free migration number before applying" caveat, i.e. neither confirms the table has landed at a fixed migration number in production.
- **what:** A `site_chat_turns` table logging one row per end-user chatbot prompt/answer turn from the site chatbot edge worker (PII: question/answer), populated by a Layer-1 puller draining the edge sink via idempotent ingest (edge-supplied turn uuid as PK, `ON CONFLICT DO NOTHING`). Records bounding outcomes (refused off-topic, capped), provenance for "why did it say that" (model, context pack_version, grounding_ids chunk list), token/latency columns name-aligned to `llm_call_log`, and a GDPR-conscious salted `client_ip_hash` instead of raw IPs, plus per-site cascade delete. Deliberately kept separate from the build-time `llm_call_log` — different owner, privacy profile, retention, and access pattern.
- **sources:** docs/agent_docs/sql_for_tables/046_site_chat_turns.sql; docs025.../086_site_chat_turns.sql; docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#8
- **relations:** llm_call_log (kept separate); rag-retrieval (context packs/grounding chunks, CHAT-003); edge workers (CHAT-002); TurnSink (CHAT-005); Isolated chat environment (CHAT-007, isolated-DB variant drops the FK)
- **verify-later:** table existence in production and its actual migration number (046 vs 086 discrepancy unresolved); edge worker + Layer-1 puller implementations

### CHAT-002 — Site chatbot edge worker (synchronous, not an orchestrated agent)
- **status:** aspirational
- **status-evidence:** Design doc with "Suggested build order (structural first)" and "Open decisions" — canonical design, nothing deployed.
- **what:** The canonical design for a per-domain chatbot on static-S3 sites: a synchronous request/response handler on a provider-agnostic serverless edge worker (Cloudflare first), NOT run through Kafka/the chassis. Deliberate documented exception to "every agent is an orchestrator" — Kafka's async failure modes (offset replay, phantom-complete, no streaming) are wrong for live chat, and a central nginx VM would drag static traffic behind a hackable box and lose S3's hack-resistance. Worker: resolve domain → load context pack → guard limits → compose bounded prompt → stream LLM tokens (SSE) → fire-and-forget record turn.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md
- **relations:** context pack (CHAT-003); provider-agnostic deps adapters (CHAT-005); site_chat_turns (CHAT-001); isolated chat environment (CHAT-007)
- **verify-later:** any edge worker deploy; /api/chat route registration

### CHAT-003 — Build-time context pack (per-domain bounded context)
- **status:** aspirational
- **status-evidence:** Section 7 defines the JSON shape and versioning; produced by an unbuilt `chat-context-builder` agent.
- **what:** One per-domain JSON document published to static storage at install time, carrying identity, scope (instructions/refusal message/banned topics), build-time-selected grounding chunks (bounded by token budget), suggested model, and operational limits. The worker holds no per-site logic — the pack is the entire bounded context. Grounding is selected on Layer 1 via Ollama embeddings + pgvector; v2 optionally ships chunk vectors for in-worker per-question retrieval plus a narrow stateless embedding endpoint.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#7
- **relations:** chat-context-builder (CHAT-004); RAG knowledge_base (install-time reuse); three-layer bounding (CHAT-006)
- **verify-later:** context-pack schema; chat-context-builder agent; pack publish-to-S3 step

### CHAT-004 — site-chat-installer orchestration (install_chat maintenance task)
- **status:** aspirational
- **status-evidence:** "Resolved: Install is a separate orchestration, triggered via a maintenance_queue install_chat task — not a build-pipeline stage." Not built.
- **what:** Chat install is its own orchestration (triggered by a `maintenance_queue` `install_chat` task, build pipeline untouched), spawning three sub-agents: `chat-context-builder` (build+publish the pack via Ollama+pgvector), `chat-widget-installer` (fork the chat widget through the existing component/tool pipeline; only difference is it POSTs to /api/chat), and `chat-route-registrar` (record the route + mark chat installed on the site, reversible via uninstall_chat). Supersedes the older `chat-suggester` gating agent from the FOCUS base version.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#install-path; docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack.md (delta: chat-suggester)
- **relations:** maintenance_queue; context pack (CHAT-003); component/tool pipeline; chat-suggester (superseded)
- **verify-later:** site-chat-installer + sub-agent definitions; install_chat maintenance task_type

### CHAT-005 — Provider-agnostic worker (deps adapters)
- **status:** aspirational
- **status-evidence:** Section 5 defines interfaces and a Cloudflare shim; "Best practice" reference impls listed, no code shipped.
- **what:** The portability strategy: a Web-platform-only core `handleChat(request, deps)` plus a ~20-line per-platform shim. Three (v2: four) small adapters — ContextStore (HTTP GET of static pack), LLMClient (Anthropic Messages over fetch, swappable to self-hosted), TurnSink (queue/D1, fire-and-forget), and v2 Embedder — each with a Cloudflare and a portable HTTP impl. Nothing vendor-specific in the core; Cloudflare/Deno/Fastly/Vercel/self-host are drop-in. Rate limiting is the least-portable concern (WAF + in-pack per-session cap floor).
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#5,#6
- **relations:** edge worker (CHAT-002); context pack (CHAT-003); pluggable billing/LLM/storage adapter discipline
- **verify-later:** handleChat core + adapter interfaces if implemented

### CHAT-006 — Three-layer bounding (retrieval / prompt / operational)
- **status:** aspirational
- **status-evidence:** Section 9 "Bounded context — the three layers"; part of the design.
- **what:** A precise decomposition of "bounded" to stop chatbot drift: retrieval bounding (only this site's grounding is in the pack, frozen at build time), prompt bounding (system prompt pins identity and emits an exact refusal message for out-of-scope questions), and operational bounding (input length, output tokens, turns/session, history window, rate limiting from pack.limits). Conflating the three is where bots go off-topic.
- **sources:** docs025.../FOCUS_site_chatbot_edge_worker_and_context_pack(1).md#9
- **relations:** context pack (CHAT-003); edge worker composeSystemPrompt (CHAT-002)
- **verify-later:** composeSystemPrompt refusal enforcement; pack.limits guards

### CHAT-007 — Isolated chat environment (satellite; load/hack/bug vectors)
- **status:** aspirational
- **status-evidence:** "Current lean (not committed — kept open): Option Y-copy ... experiment in a sandbox" — explicitly undecided across both document versions cited by the two extraction units (PLAN_isolated_chat_environment(4)/(1) and (3)).
- **what:** A plan to run the site-chatbot's server-side pieces (turn store, drain/analytics, optionally chat workflow code) on infrastructure separate from the core build cluster, severing three blast vectors: load (turn write-load competing with builds), hack (a compromised internet-facing edge worker's reachable radius), and bug (chat code faulting the shared chassis). Deliberately does NOT reuse the coupled multi-cluster dispatch (Phase 4a, `remote-job-spawner`), which shares core Kafka/Postgres by design — the satellite instead reuses only chassis binaries/action code against its own Kafka/DB/storage, with a strictly one-directional, async, egress-from-core-only boundary. Two sizing options are weighed: Option X (minimal — turn store + puller + analytics only, maybe no chassis at all) vs Option Y (a full cut-down copy of the whole chassis, Y-copy config-only vs Y-slim purpose-built image); a worked "building-and-hosting-as-a-service via chat" example reframes the satellite as a second customer-facing platform instance and pushes toward Option Y for that case.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md; docs025.../PLAN_isolated_chat_environment(1).md; docs/_archive/.../PLAN_isolated_chat_environment(3).md
- **relations:** remote-job-spawner (NOT reused); site_chat_turns (CHAT-001); boundary contract; building-as-a-service; multicluster dispatch (Phase 4a, the rejected template). Closely related to Isolated chat/satellite architecture for SaaS build isolation (saas-isolation-architecture.md SAAS-001) — a near-duplicate framing of the same architecture applied specifically to a build-as-a-service escalation, registered separately there because it was tagged as its own assigned category; worth reconciling in stage 2.
- **verify-later:** any separate chat cluster/DB; isolated-DB variant of migration 086; remote-job-spawner (the Phase 4a mechanism used as contrast)

### CHAT-008 — Simple paid multi-domain chat (freemium + day-pass)
- **status:** aspirational
- **status-evidence:** "Status: discussion draft — direction firming."
- **what:** A deliberately simple "fast lane" route: the FOCUS edge worker + a light paywall, multi-tenant-by-Host, add a domain by publishing config + DNS (no chassis/Kafka/satellite). Monetisation is freemium + a flat day-pass (£2-5) rather than counted credits, because card processing's fixed ~20-30p fee makes sub-£5 one-off charges poor. Entitlement is a stateless signed `{domain, expiry}` token issued via a synchronous Stripe guest-checkout `redeem` (no accounts, no webhook on the critical path, no edge KV). The real cost driver is the free taster + abuse, not paying users.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md; docs025.../PLAN_simple_paid_multidomain_chat.md
- **relations:** edge worker (CHAT-002); context pack (CHAT-003); chat lanes (CHAT-009, fast lane); commercial model/billing adapter
- **verify-later:** paywall gate + redeem endpoint; day-pass token signing/validation

### CHAT-009 — Chat lanes (fast/slow/job) + warm-adapter maturation
- **status:** aspirational
- **status-evidence:** Section 11 "What the agent framework buys chat (lanes and maturation)"; "This is still open by design; needs further conversation."
- **what:** A model splitting chat by what it does: fast lane (bounded Q&A, synchronous/streamed, no framework — ships independently); slow lane (turns needing work — live research, structured-data queries, running a site's tool, in-answer charts, multi-step tasks — routed by a cheap intent classifier, user warned it's slower); job lane (long-running submissions like "build me a site", ack + status + deliver). Maturation path: prove a slow-lane capability as a spawned agent (~12s cold), promote popular ones to warm adapters, end-state a resident chat-orchestrator adapter that fans out without spawning per turn.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#11
- **relations:** simple paid multi-domain chat (CHAT-008, fast lane); building-as-a-service (job lane); warm adapters
- **verify-later:** intent classifier; any resident chat-orchestrator adapter
