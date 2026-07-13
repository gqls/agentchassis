# Framework Reframed — Substrate, Not Product

Experimental alternative framing where the framework itself is the headline and the website builder is one demonstration of it among several. Honest about what's implemented at every step. To compare against, not necessarily replace, the original pitch doc.

---

## 1. The Reframed Headline

> I built a domain-agnostic distributed agent orchestration substrate. The thing in the middle is a chassis — a single Go binary — that can be any agent on the system depending on which agent definition it loads from the database. The agent definitions live in PostgreSQL. The chassis communicates only via Kafka. Pods are spawned as Kubernetes Jobs, one per agent. Every agent on the system is itself an orchestrator that can spawn further agents recursively, all using the same chassis, same topics, same state model.
>
> This produces a fractal architecture: behaviour at any depth of the agent tree looks the same as the level above and below it — same spawn / call / claim / complete primitives, same fault tolerance, same observability hooks, same Kafka topic conventions. There is no fundamental difference between a "top-level orchestrator" and a "leaf specialist agent." Both are agents. The leaf becomes a sub-orchestrator the moment it needs to spawn anything.
>
> The substrate is what I built. The applications running on it are what prove it. The most ambitious is an autonomous website-building and maintenance pipeline — five live sites today, the most complex demonstration of what the substrate can do. Alongside it, the same substrate runs a Companies House business-intelligence pipeline (5,780 companies collected, 23.2% match rate against a target vertical, succession-risk signals derived), a UK vet-practice discovery sweep (postcode-by-postcode, ~5,000 practices captured), a medicine-price scraping pipeline (URL discovery via Firecrawl, daily exports), and a multi-source news feed pipeline (RSS + LLM-based news search + Grok/xAI Responses API + scrape) with LLM-based relevance and credibility triage.
>
> The substrate is what scales — horizontally within a cluster today, and across clusters in principle (Kafka, Postgres, and stateless pods admit it without redesign; the operational glue is not yet built). The applications are how I've stress-tested it.
>
> Mainly: Go, Kubernetes (jobs, deployments, services, RBAC, kustomize), Kafka, PostgreSQL with pgvector + pgbouncer, MySQL, Terraform, Docker, Anthropic Claude (Opus / Sonnet / Haiku) via direct API and Batch API, Ollama (CPU + GPU), Firecrawl, GitHub Actions, Backblaze S3, Cloudflare DNS.

---

## 2. The Fractal Architecture — what makes "every agent is an orchestrator" technically meaningful

This is the part of the architecture that's genuinely novel and worth being able to explain in 90 seconds.

### 2.1 Same primitives at every depth

- Every agent runs the same Go binary (`agent-chassis`). The differentiator at startup is the `AGENT_TYPE` env var, which the chassis uses to load its agent definition (workflow, model config, contracts, idle timeout) from PostgreSQL.
- Every agent communicates only via Kafka topics named `system.agent.{type}.requests` and `system.agent.{type}.responses`.
- Every agent has an `orchestration_states` row per running orchestration, with `collected_data` (JSONB accumulator), `current_step`, `status`, and `workflow_plan` columns. State is in the database, not in process memory.
- Every agent uses the same orchestration primitives — `spawn_agent`, `call_agent`, `claim_work_item`, `complete_work_item`, `conditional`, `loop`, `execute_llm_prompt`, `query_database`, `write_site_spec`, etc. The action registry is global; any agent can invoke any registered action.
- Every agent benefits from the same fault tolerance: `awaited_requests` for timeout tracking, the stale-orchestration sweeper for lost responses, `idle_timeout_seconds` for pod self-shutdown, ephemeral topic cleanup, and the dispatch loop's `continue_on_error` semantics.

### 2.2 Self-similarity in practice

A real call chain from the production website pipeline:

```
intake-orchestrator
  ↓ spawns
  site-work-orchestrator
    ↓ spawns
    build-dispatch-loop
      ↓ spawns (per work item, dynamic type)
      page-build-handler
        ↓ spawns
        page-content-writer
          ↓ spawns
          research-agent
            ↓ might spawn
            web-fetcher (sub-orchestrator within a sub-orchestrator)
```

Seven levels deep. Each spawn point uses identical code paths. Each level has its own `orchestration_states` row, its own Kafka topics, its own pod, its own consumer group. Each level is queryable by the same SQL diagnostic templates. Each level is independently restartable, killable, and inspectable.

There is no architectural difference between level 1 and level 7. There is no "leaf agent" type — `research-agent` could itself spawn three sub-researchers if its workflow asked it to, with no changes to its definition shape. This is what makes the architecture fractal: the agent abstraction is recursive, and the recursion is not bounded by the framework.

### 2.3 The contrast with how other frameworks do it

- **LangGraph / LlamaIndex / CrewAI / AutoGen / DSPy.** "Sub-agents" are nested function calls inside one Python process. State is in-memory. Failure of any task fails the parent. Restart loses everything in flight. "Distributed" means stateless API calls to a remote LLM, not distributed agents.
- **Most "production" agent frameworks** add Celery, RQ, or a hosted task queue as a layer on top of in-process Python. This gives you task queueing but not hierarchical orchestration; sub-agents are still in-process, the queue just defers when work runs.
- **This substrate** treats every agent as an independent process with its own pod, topics, DB row, and consumer group, communicating only via messages on a durable log. The recursion is real recursion in the deployment sense, not just in the function-call sense.

The cost is operational complexity (Kafka, Postgres, K8s instead of `pip install` and `python main.py`). The payoff is genuine survivability — kill any pod at any time, the work continues from the database state.

### 2.4 Tiers emerge naturally; they're not enforced

There is no rule in the framework that says "this agent is tier 1, this is tier 3." Tiers emerge from how work decomposes. In practice the production system has roughly five tiers:

| Tier | Examples | What it does |
|---|---|---|
| Strategic orchestrator | `intake-orchestrator`, `site-work-orchestrator`, `vet-pipeline-orchestrator`, `improvement-loop` | Coordinates a whole pipeline from a single trigger |
| Domain orchestrator | `build-dispatch-loop`, `content-feed-orchestrator`, `med-price-scrape-orchestrator`, `area-sweep-orchestrator` | Iterates over a queue or collection, dispatches per-item work |
| Handler | `page-build-handler`, `image-build-handler`, `tool-recreation-handler`, `vet-batch-processor` | Owns persistence and side-effects for one item type |
| Specialist | `page-content-writer`, `webdesign-agent`, `site-classifier`, `feed-ingester`, `feed-triage`, `vet-practice-verifier`, `ch-collector` | Produces structured output for its caller |
| Tool / utility | `research-agent`, `endpoint-health-checker`, `work-item-archiver` | Single-purpose, often spawned by specialists |

But — and this is the important part — any tier can spawn any other tier. A specialist that needs complex sub-work spawns its own domain orchestrator. A handler that finds it needs another handler can spawn one. The tiering is a description of common patterns, not a rule. The framework only enforces the orchestrator abstraction; the tiers are emergent.

---

## 3. What the Substrate Provides (domain-agnostic capabilities)

Everything here works the same regardless of whether the application is "build a website" or "monitor regulatory changes" or "scrape a marketplace."

### 3.1 Work-item lifecycle infrastructure

- `site_work_items` (despite the name, generic) — every piece of work is a row with `source`, `pipeline`, `item_type`, `severity`, `priority`, `depends_on UUID[]`, `parent_item_id`, `item_key` (deterministic dedup), `handler_agent`, `status`, `spec JSONB`, `result JSONB`, `attempt_count`, `max_attempts`. The table name predates the realisation that the same table serves any application; the rename is on the list.
- 12 lifecycle states including `needs_human_review` and `unresolved`. The `unresolved` state in particular is interesting: if the same item is rediscovered twice in 7 days, the third instance is born as `unresolved`, stops being claimed, and surfaces for investigation. Prevents fix-rediscovery loops in any application.
- Atomic single-execution claiming via `UPDATE ... WHERE status = 'triaged' RETURNING ...`. No Redis, no broker locks. Postgres MVCC plus a conditional UPDATE is sufficient.
- Dependency graph via `depends_on UUID[]`. The dispatch query excludes any item whose dependencies aren't in `(complete, verified)` status. Means complex multi-stage pipelines (research → analyse → write → review → deploy) express as DAGs of work items, not as workflow branches.

### 3.2 Hierarchical orchestration

- `orchestration_states` table tracks every running workflow, every step's accumulated data, every transition. The `parent_orchestration_id` column ties spawned agents back to their caller — the full agent tree is reconstructable from a single SQL query.
- `awaited_requests` tracks every pending parent-child call with a timeout. When the timeout expires, the stale-orchestration sweeper classifies why (lost response, dead child, child running long) and synthesises recovery — without requiring application code to handle each case.
- `agent_spawn_history` is the audit log of every spawn — parent, child, time, reason, config.

### 3.3 Heartbeat scheduling

- `scheduled_tasks` table — single-row entries with `interval_seconds`, `target_agent_type`, `target_topic`, `input_data`, optional `pre_query`, `concurrency_group`, `max_concurrent`, `enabled`.
- Adding a new scheduled job is an `INSERT`. No code change, no redeploy. The kafka-scheduler service polls the table every 30s.
- `pre_query` is a SQL `SELECT` that runs before each trigger — both as dynamic input (column values merged into the message body) and as a gate (no rows returned → task skipped this tick).
- `concurrency_group` + `max_concurrent` give per-group concurrency control across the cluster, not just per-task.

### 3.4 LLM cost engineering as substrate, not application

- `llm_call_log` captures every LLM call (agent, step, model, prompt rendered, response, tokens in/out, latency, success, error, work_item_id, prompt_variant, vertical, rag_context_used). Universal — every application contributes to the same log.
- `ai_endpoint_health` per-endpoint up/down state plus reactive (failure-driven) and active (haiku ping) checks. Items waiting for a model don't get claimed if their endpoint is unhealthy — they stay triaged. Universal across all applications.
- `swap_agent_model()` SQL function with auto-snapshot. Per-agent, per-step. Means cost optimisation is a SQL call, not a code change. Universal.
- `llm_batch_queue` + universal LLM work queue — can route any LLM call to Anthropic Batch (50% off), Anthropic direct, GPU Ollama, CPU Ollama, image generation, or a future provider. Universal.

### 3.5 Knowledge accumulation

- `knowledge_base` with `vector(768)` (pgvector + nomic-embed-text) — any agent can `rag_index` to it, any agent can `rag_lookup` from it. Metadata filtering by `vertical`, `component_type`, `content_type`, `source`, `source_quality` so retrieval is scoped, not global.
- Trigram fallback when Ollama is down — RAG never hard-fails, it degrades gracefully.
- `training_exports.runs` and `training_exports.rows` for fine-tuning data export, versioned by UUID. Any agent's prompts and responses are training material for any future model.

### 3.6 Human-in-the-loop as a built-in, not bolted on

- HITL responses flow through the same Kafka response topic as automated responses, distinguished only by `sender_agent_type: "human"`. The orchestrator doesn't know it was a human; the response handling path is identical.
- `awaited_requests` rows pause workflows waiting for human input. The dashboard surfaces them. The same admin UI works for any application.
- Lock types (permanent / timed / review) on any record that has the lock columns. The component-locking semantics generalise to any "this is human-owned, don't touch" use case.

### 3.7 Snapshots and revert

- `site_snapshots` (again, name predates generalisation) — JSONB blobs of full state at a point in time, surviving row deletions and schema changes. Triggered on deploy, manually, pre-edit, scheduled, pre-revert. SQL functions to capture and restore.
- Same pattern works for any record set you want point-in-time recoverable.

### 3.8 Deployment adapter abstraction

- The `git-adapter` is one of a planned family. `cpanel-adapter`, `wordpress-adapter`, `laravel-forge-adapter`, `vercel-adapter`, `shopify-adapter`, `cloudflare-workers-adapter` share the same interface (auth → file push → build trigger → DNS → SSL → health check).
- This isn't substrate-level today, but it's the design for it.

---

## 4. Demonstrations (applications running on the substrate)

These all share the same chassis, same Kafka, same Postgres, same scheduling, same observability. Adding a new one is mostly SQL inserts (agent definitions, scheduled tasks, prompts) plus a few new Go actions for genuinely new capabilities.

### 4.1 Website builder (most developed application — ~80% of total agent count)

- Full lifecycle: research → classify → brief → plan → compose design → render CSS → write content per page → audit → deploy → maintain.
- 50+ agent definitions specific to this application sit in `agent_definitions` alongside the other application's agents.
- Five live production sites. Self-maintaining via the improvement loop. Detailed in the original pitch document — won't repeat here.

### 4.2 Business intelligence: Companies House enrichment (deployed, real numbers)

- Bulk collect 5,780 active companies in SIC 75000 (veterinary services) → two-pass local matching (postcode+name then trigram name-only) against a 2,730-business target list → LLM disambiguation of ambiguous matches via Claude Haiku (~$0.05 per review run) → fetch officers, PSC (persons with significant control), and accounts (iXBRL extraction) for confirmed matches → derive succession-risk signals.
- 634 confirmed matches (23.2%); 83 pending/uncertain for human review; ~5,100 unmatched as discovery candidates for net-new lead generation.
- Runs on a single `business-intel` pod via Kafka routing — five different agent types share one image, with the `agent_type` in the message body selecting the workflow.
- Same substrate, same chassis, same observability as the website builder. Different `agent_definitions` rows, different `scheduled_tasks` entries.
- **Generalisable to any vertical.** SIC code, name patterns, and disambiguation prompt are configuration. Replacing 75000 with 96020 (hairdressing) or 47990 (other retail) is a SQL update.

### 4.3 UK vet-practice discovery (deployed)

- Postcode-district-by-district sweep dispatcher (`area-sweep-orchestrator`) loads UK postcodes, spawns one `area-sweep-discoverer` per district.
- Discoverers issue web searches scoped to "veterinary practice in BT4 Belfast UK"-style queries, parse results, deduplicate.
- Verifier agent (`vet-practice-verifier`) checks discovered candidates, enriches with public data, writes to a database. Output exported as JSON consumed by vetcomparison.uk's client-side search.
- ~5,000 practices captured this way. Demonstrates: the same substrate is doing geographic-grid search, not just per-domain build work.
- The pattern is reusable for any "find entities of type X across geographic regions" application.

### 4.4 Medicine pricing pipeline (deployed)

- URL discovery via Firecrawl `/map` (`med-url-mapper`) and via category-page scraping (`med-url-discoverer`) — two independent discovery strategies.
- Per-URL price collector (`med-price-collector`) scrapes pharmacy product pages and extracts structured price data.
- JSON exporter (`med-json-exporter`) commits current prices to the site's git repo for client-side rendering.
- Demonstrates: continuous data refresh against external sources, with the substrate handling scheduling, dedup, fault tolerance, and structured output.
- Each `med-*` agent has an orchestrator wrapper pairing — the wrapper spawns the worker in a dedicated pod. This is now a documented pattern ("one-step orchestrator → real worker") used wherever an agent needs its own pod without a parent in the same workflow.

### 4.5 Multi-source news feed pipeline (deployed for sites, generalisable beyond)

- Four ingestion modes: RSS (`fetch_rss`), LLM-based news search (`fetch_news_search` via the web-search adapter), structured-news API (`fetch_llm_news` via xAI Responses API with web + X search), and direct scraping (`fetch_scrape`).
- Two-pass design: Run N triages items ingested by Run N-1 and dispatches new ingesters; new items are scored by Run N+1. No synchronous wait on async ingestion.
- Triage by relevance and credibility against the consumer's spec (currently a site spec; could be any taxonomy / watchlist / persona).
- Source-diversity render: `ROW_NUMBER() OVER (PARTITION BY source_id)` interleaves sources so a top-N render isn't dominated by one prolific RSS feed.
- Renders to JSON files attached to the relevant site. Currently the terminal step is "commit JSON to git for S3 serving" — the same substrate could trivially route to:
  - A daily-briefing email to a subscriber list (new Go action: `send_briefing_email`).
  - A Slack channel post (`send_to_slack`).
  - An RSS feed of curated digest items (`render_rss_feed`).
  - A standalone news dashboard with no associated website (terminal step writes to a different bucket).
  - A pre-LLM training corpus for vertical fine-tuning.
- The orchestration, the ingestion, the triage, the dedup, the credibility scoring — all of that is substrate. Only the terminal "what do you do with the curated items" varies by application.

### 4.6 Internal training-data flywheel (substrate as application of itself)

- `training-data-export-orchestrator` spawns `training-data-exporter` as a dedicated pod. The exporter reads from `llm_call_log` and writes versioned datasets to `training_exports.runs` and `training_exports.rows`.
- First real export run produced 1,949 training rows from one step (`page-content-writer / iter_0_generate_content`) — clean enough for LoRA fine-tuning, with stripped markdown fences, validated JSON structure, dominant schema (hero-with-CTAs 68%, minimal hero 18%, header/nav 9%).
- The same substrate's observability surface (`llm_call_log`) is producing the training data for its own future improvement. This is the most direct demonstration of substrate value: a feature added to power one pipeline (cost observability) became the input to another pipeline (model training) without redesign.

### 4.7 Applications the substrate admits without redesign

These aren't built; they're examples of where the same plumbing extends. Honest list — what would need to be added is mostly SQL plus a small Go action surface, not architectural change.

- **Continuous watchlists.** Watchlist becomes a row; trigger pre-query selects due watchlists; existing news pipeline runs against the watchlist's taxonomy; terminal step delivers to subscriber. **What's new:** the watchlist table and a few terminal-step actions.
- **Compliance / regulatory monitoring.** Source feeds for regulator publications (FCA / SEC / ICO / Ofcom etc.) ingested through the existing feed pipeline; LLM triage scores items against tracked client industries; impact briefs generated. **What's new:** source seeding for regulator feeds, an impact-brief generator agent.
- **Lead generation pipelines.** "Find businesses in vertical X within region Y, enrich with public data, score by fit, output qualified leads." Combines existing CH enrichment, area-sweep discovery, and a new scoring agent. **What's new:** a scoring agent definition.
- **Marketplace data collection.** Pattern proven by medicine pricing and vetcomparison. Generalises to "watch prices / availability / reviews for a set of products across a set of retailers, produce comparison data, refresh on a cadence." **What's new:** retailer-specific scrapers (one Go action each), product taxonomy.
- **Document processing at scale.** Ingest PDFs / iXBRL / scraped pages → extract structured data via LLM → validate → deduplicate → export. iXBRL extraction already done for Companies House accounts. **What's new:** more extractors per document type.
- **Cross-vertical research digests.** Continuous research across an industry, summarised as periodic briefings. The research-agent already does single-shot vertical research as a sub-spawn for the classifier. **What's new:** scheduling and a synthesis agent.
- **Mini admin / SaaS apps for clients.** The dynamic-applications doc lays out the tiers; the framework can in principle generate backend code as artefacts in the same way it generates HTML and CSS today. **What's substantial:** the framework specs for target stacks (Next.js, Laravel, WordPress).

The honest claim is: each of these would be days-to-weeks of new work, not months. Most of the work is "what does the output look like" and "where does it go." The fetching, ranking, scheduling, retry, fault tolerance, cost engineering, HITL, observability, snapshotting are all substrate.

---

## 5. Scalability — Including the Cross-Cluster Story

### 5.1 Horizontal scaling within one cluster (proven)

- Add nodes, K8s scheduler picks them up, more concurrent agents, no architectural change.
- Current cluster: three nodes, several hundred concurrent agents at peak, headroom is large.
- Bottleneck is LLM API spend, not compute. Compute is cheap; tokens are not.

### 5.2 Vertical scaling per service (proven)

- pgbouncer in transaction mode means each Go service runs with 10 application connections, pooled to 15 server-side at most. The Postgres pod has not been close to capacity.
- Postgres pod can scale vertically several orders of magnitude before sharding becomes necessary.
- Kafka is a 3-broker Strimzi cluster on the same K8s — proven, room to grow horizontally.

### 5.3 Cross-cluster scaling — what the architecture admits, what isn't built

This is the honest version of the "fractal at any scale" claim. The architectural admission is real; the operational glue isn't done.

**What the architecture admits:**

- All agent state lives in PostgreSQL. Multiple K8s clusters can read and write the same Postgres (via service mesh, public ingress, or VPC peering) and run agents that participate in the same orchestration tree.
- All inter-agent messaging is Kafka. Multi-cluster Kafka is a solved problem (Strimzi supports it; Kafka Connect / MirrorMaker / cluster linking are mature). Federated topics across regions are a standard pattern.
- Agent pods don't share filesystem. They don't co-locate. They share durable state (Kafka, Postgres) and that's it.
- Agent definitions are rows in a database. A cluster in a new region that can read `agent_definitions` can spawn any agent type already defined.
- The chassis image is one binary. The same image runs in any cluster, differentiated only by `AGENT_TYPE` env var at startup.
- Pod-level idle timeout means spawned agents are ephemeral by design — a cluster can scale up for a workload and back down without manual intervention.

**What isn't built:**

- Multi-cluster Kafka mirror topology. Currently single Strimzi cluster.
- Postgres read replicas in target clusters or shared via service. Currently single Postgres pair.
- A cluster-tag column on `agent_definitions` (or K8s `nodeSelector` patterns) to control which agents run in which cluster. Currently every agent type is eligible everywhere.
- Coordinated secrets management across clusters. Currently single K8s secret store per cluster.
- Latency-aware routing — if an agent in cluster A needs to spawn an agent in cluster B, the spawn happens via Kafka, which works, but optimisation for keeping a workflow's hot path within one region isn't designed.

**The honest claim:** a satellite cluster in another region could be brought up in a sprint of operational work, not a redesign. The agent code doesn't change. The Postgres connection string changes. The Kafka bootstrap servers change. The chassis image deploys unchanged.

**What would change:** the database connectivity layer (it reads from env vars today — straightforward to point at a federated endpoint), the Kafka producer/consumer config (similarly env-var driven), and the chassis would learn to filter `agent_definitions` by a cluster tag if you wanted to restrict which agents a region can run.

The reason this isn't done is that the workload doesn't need it yet. Three nodes in one cluster have plenty of headroom for current and projected near-term traffic. Cross-cluster is a future option the design protects, not a feature ready to ship.

### 5.4 Cost as the real ceiling

- The genuine scale-blocker is not compute, not Kafka, not Postgres — it's LLM API spend.
- Mitigations all in flight (batch API integration, fine-tuned local models for high-volume agents, RAG to shrink prompt sizes, audit caps to bound re-spending, prompt caching).
- The substrate is built to make these mitigations operational rather than rewrites: `swap_agent_model()`, `llm_batch_queue`, `ai_endpoint_health`, model routing per agent per step.
- 2,000 domains per cycle projected at $15-30k without optimisation. Same number with batch API at 50% off the eligible call mix: ~$8-15k. With fine-tuned local models taking page-content-writer (the largest token consumer) off Claude: substantially less. None of these require touching the framework — they're configuration.

---

## 6. Why This Framing Matters

### 6.1 What it changes about the pitch

- "I built a website builder" describes an application. Many people have built website builders. The market for them is crowded.
- "I built a distributed agent orchestration substrate, and the website builder is one demonstration of it" describes infrastructure. The market for production-grade agent infrastructure is small and the gap between Python single-process toys and production-credible orchestration is wide.
- The website builder is impressive evidence; it's not the asset.

### 6.2 What it changes about positioning

- Roles where the website framing fits: senior backend engineer at a marketing-tech or content company; founding engineer of a marketing-automation startup.
- Roles where the substrate framing fits: infrastructure engineer for an AI-tooling company; AI platform engineer at a large enterprise building internal agent platforms; founding engineer / CTO at an AI infrastructure startup; consulting / advisory for organisations building production agent systems.
- The audience for "I can build production websites with AI" is different from the audience for "I can build production agent systems and here's the substrate I've built and stress-tested across several different problem domains."

### 6.3 What it doesn't change

- The system is the system. Reframing doesn't add features. Everything claimed has to be backed by what's actually in the codebase.
- The honest weak points carry over: solo development, no formal CI test suite, schema drift history, race conditions found and fixed, multi-tenant work undone. These are properties of what was built, not properties of the framing.

### 6.4 Risks of this framing

- **Sounds grander than it is.** A skilled interviewer will probe and find that, e.g., multi-cluster is admitted but not built; that the news pipeline currently writes to site-attached JSON not standalone briefings; that the universal LLM work queue is rolled out for two agent types not all of them.
- **Defensible if you're honest.** Each probe has an honest answer: "admitted by the architecture, not built, here's roughly what it would take." The substrate framing only collapses if you over-claim what's in production.
- **The fractal claim needs to land in a sentence.** Most interviewers haven't heard "fractal" used architecturally. Practice the 90-second version in section 2 above. If it doesn't land in 90 seconds, the framing is hurting rather than helping.

---

## 7. Honest Delta — What's Built vs What the Architecture Admits

If you adopt this framing, you have to be ready to draw this distinction quickly and without defensiveness. Here is a candid version.

| Capability | Built and proven | Built and partial | Admitted by the architecture, not built |
|---|---|---|---|
| Domain-agnostic chassis (one binary, AGENT_TYPE differentiates) | ✓ | | |
| Agent definitions in DB, hot-swappable | ✓ | | |
| Hierarchical spawning across many levels | ✓ (production traces show 6+ levels) | | |
| Same primitives at every depth | ✓ | | |
| Fault tolerance via stale-orchestration sweeper | ✓ | | |
| Atomic single-execution work-item claiming | ✓ | | |
| Universal LLM work queue with sync-fallback | | ✓ (queue and Go action live; batch enabled for two agents; rollout incremental) | |
| Multi-provider model routing (Claude / CPU Ollama / GPU Ollama) | | ✓ (Claude and CPU Ollama in steady production; GPU on-demand only) | |
| Fine-tuning data export pipeline | | ✓ (one export run done; LoRA training scripted, awaiting GPU) | |
| RAG with pgvector | | ✓ (verified end-to-end; integrated into one test agent; production workflows not yet using rag_lookup) | |
| HITL via awaited_requests with dashboard | ✓ | | |
| Snapshot / revert | ✓ | | |
| Multiple application domains (website, CH BI, vet discovery, med pricing, news, fine-tuning) | ✓ | | |
| Cross-cluster operation | | | ✓ (all state is in Postgres + Kafka; chassis is one binary; operational glue not built) |
| Multi-tenant isolation (per-tenant compute / data) | | | ✓ (clean enough boundaries; Phase 4 of roadmap) |
| Deployment adapter family beyond git | | ✓ (git-adapter only; abstraction designed) | |
| Standalone news briefings (not site-attached) | | | ✓ (substrate exists; terminal-step actions to deliver elsewhere are not built) |
| Watchlist-as-first-class (vs site-as-first-class) | | | ✓ (would be a SQL inserts + small action surface change) |
| Dynamic backend application generation | | | ✓ (Tier 2-3 of dynamic-apps roadmap; substrate ready, framework specs needed per target stack) |

What this table protects you from: over-claiming. What this table allows you to claim: that the architecture admits a much wider class of applications than has been built, and that adding new applications is mostly SQL + small action surfaces rather than redesign. Both are defensible.

---

## 8. How to Use Both Pitches

The two framings are not mutually exclusive. They serve different audiences and different moments.

| Use the original (website-pitch) framing when | Use this (substrate-pitch) framing when |
|---|---|
| Audience expects a concrete product story | Audience is technical infrastructure / platform engineering |
| Talking to commercial / marketing-tech roles | Talking to AI-infrastructure / agent-platform roles |
| The conversation is about revenue and applied AI | The conversation is about engineering depth and scalability |
| You're describing what's currently earning | You're describing what could earn next |
| Five minutes or less, need a vivid demonstration | Twenty minutes or more, room for an architectural walk-through |

The original pitch grounds the abstract claim. The substrate pitch raises the ceiling on what someone might hire you to do. Both true; emphasis differs.

---

## 9. Words to Use / Words to Avoid in This Framing

**Use:**
- "Substrate" — captures the foundational, application-agnostic positioning.
- "Fractal" — but only if you can defend it in 90 seconds (see Section 2). Otherwise skip.
- "Recursive composition" — engineers know what this means and it accurately describes the spawn pattern.
- "Demonstrations" or "applications" rather than "use cases" — "use case" sounds like a sales deck.
- "Same primitives at every depth" — specific, falsifiable, true.
- "Admits without redesign" — honest about what's structural vs operational.

**Avoid:**
- "Platform" — over-used, vague, sounds like vapourware unless you have customers.
- "AI-powered substrate" — substrate is the right word, "AI-powered" is the wrong modifier.
- "Generic agent framework" — accurate but flat. "Domain-agnostic distributed agent orchestration substrate" is a mouthful but it tells someone exactly what you mean in one phrase.
- "Multi-cloud" / "multi-region" / "global scale" — until those are built, these phrases are landmines.
- "Solves the problem of X" — the framework doesn't solve any single problem; it makes a class of problems addressable. Better to say "makes X tractable" or "the substrate handles X so the application doesn't have to."
