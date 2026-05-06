# Distributed AI Agent Framework — Pitch & CV Reference

Working document for verbal pitches, written CV blurbs, and deeper conversations. Each section starts with a tight summary and then drills into points you can pull out individually depending on the audience.

---

## 1. The 60-Second Headline (improved version of your current paragraph)

> Following my time leading the internal AI community for management at Bumble, I wanted to build an agent framework that could genuinely scale rather than the single-process Python notebook setup that most popular frameworks (LangGraph, CrewAI, AutoGen) collapse to. I designed a natively distributed, hierarchical architecture in Go, where every agent is itself an orchestrator that can dynamically spawn sub-agents (a content agent spawning its own researchers and tone-checkers, a build orchestrator spawning a planner, a writer, a designer, and an auditor). Each agent runs in its own Kubernetes pod with its own Kafka topics and consumer group, so death of any single pod is a local event, not a system event.
>
> To prove the framework end-to-end I built a fully autonomous website-building and maintenance pipeline. Given a domain name, the system researches the vertical, decides what the site *should be* (revenue model included — not defaulted to a brochure), writes a versioned aspirational spec, builds the multi-page site, deploys it via git → GitHub Actions → Backblaze S3, then keeps running a discover → audit → triage → fix → rerender loop against a "site spec" that is the ground truth. Several real sites are running on this today (vetcomparison.uk, gaswholesalers.com, finetuning.uk, leopardessconsulting.co.uk, ai-agent-orchestration.com), each self-maintaining, with several hundred agents running concurrently across just three cheap servers.
>
> Mainly: Go, Kubernetes (jobs, deployments, services, RBAC, kustomize), Kafka, PostgreSQL with pgvector + pgbouncer, MySQL, Terraform, Docker, Anthropic Claude API, Ollama (CPU + GPU), Firecrawl, GitHub Actions, Backblaze S3, Cloudflare DNS.

---

## 2. The Core Concept (one-liner versions for different audiences)

- **Engineering audience:** "An orchestration framework where every agent is itself an orchestrator, runs as its own Kubernetes pod, and communicates only via Kafka — natively horizontal-scalable rather than retrofitted."
- **Product / business audience:** "An autonomous website factory: hand it a domain name, it researches the vertical, builds a multi-page site, deploys it, and keeps improving it with no human in the loop unless you want one."
- **AI / research audience:** "A real-world test bed for taming LLM agents — every prompt, response, token count, work item, and outcome is logged and feeds a fine-tuning flywheel; quality gates run before, during and after generation; auditors enforce upstream specs rather than overriding them."
- **Strategic / commercial audience:** "An always-on revenue-aware site builder. The classifier picks the revenue model that fits the domain — affiliate, ad-supported tools, SaaS marketing, services — instead of defaulting every domain to a consultancy brochure."

---

## 3. Architectural Strengths

### 3.1 Genuine distribution (vs. single-process frameworks)

- **Each agent is its own Kubernetes Job pod**, with its own image entry, its own Kafka consumer group, its own PostgreSQL connection pool. A misbehaving agent cannot starve the rest of the system.
- **Hierarchical spawning is real, not simulated.** When an orchestrator spawns a sub-agent, a new K8s Job is created, registered in `agent_instances`, and given its own request and response Kafka topics. This is unlike Python frameworks where "sub-agents" are just nested function calls in one process.
- **Spawn → call separation.** Agents must be `spawn_agent`'d before they can be `call_agent`'d. The spawn step writes the registration row and creates topics; the call step then publishes a request. This keeps orchestration tracking honest — every parent-child relationship is in the database, queryable, replayable.
- **Ephemeral topics per spawn** (`EPHEMERAL_TOPICS=true`). A future migration to shared topics is already designed; the agent code reads from env vars either way, so the cutover is purely infrastructural.
- **Idle timeout on spawned agents** — `agent_definitions.idle_timeout_seconds` per agent type (default 120s for Job-spawned). Pods exit cleanly after inactivity; K8s TTL reaps the Jobs. Without this, pods would sit in Kafka consumer loops forever and the cluster would saturate.
- **Three nodes (~£cheap-spot-tier), several hundred concurrent agents.** Headroom is large; the bottleneck is API spend, not compute.

### 3.2 "Every agent is an orchestrator"

- A first-class architectural rule, not just a slogan: any check that lives as an action step inside an agent can be promoted to its own agent with one line of workflow change. The check moves from `query_database` step to `call_agent` step; the output field stays the same; nothing downstream breaks.
- This means **growth is incremental, not big-bang**. New auditors, fixers, discovery checks, even whole pipelines (news feeds, Companies House enrichment, vet practice discovery, medicine price scraping) bolt onto the same agent graph without touching core code.
- **Group agents own shared context.** Checks that need the same data (CSS theme, palette, page samples, brief) are grouped into one agent that loads context once and runs all checks in one LLM call — one round-trip for many subjective judgements, not one per check.

### 3.3 Database-driven configuration

- **Agent definitions live in PostgreSQL** (`agent_definitions` table), keyed by `(type, version)`. The workflow itself is JSONB. Adding a new agent is a SQL `INSERT`, not a redeploy.
- **Workflows are declarative**; complexity stays in Go actions. The rule "keep workflows simple, complexity in Go" is enforced socially via the development guide and tested against constantly. Workflow steps are: `action`, `config`, `next_step`, `output_field`, `error_step` — no branching except via `conditional` actions.
- **Hot-swappable models per agent per step** via `swap_agent_model()` SQL function (with auto-snapshot). The config-driven `ai_service` block on each LLM step (`provider`, `model`, `api_key_env_var`, `max_tokens`, optional `budget_tokens` for extended thinking) means swapping briefing-agent from Claude Sonnet to Mistral-Small-3 on CPU Ollama is one SQL call. `revert_agent()` restores from the most recent snapshot.
- **Dynamic dispatch.** The build dispatch loop spawns whatever handler the work item asks for via `agent_type_field: "current_item.handler_agent"`. The loop itself is generic — it does not know about specific item types.

### 3.4 The unified work-item queue

- Every piece of work — building a page, fixing a stale date, deploying an asset, publishing a news article, replanning a tool, regenerating CSS — is a row in `site_work_items`.
- Rich metadata on every item: `source` (planner / discovery / content_feed / manual / improvement / side_effect), `pipeline` (build / content / links / seo / compliance / structural / design / navigation / entity / tools), `severity`, `priority`, `depends_on UUID[]`, `parent_item_id`, `item_key` (deterministic dedup key), `status` (12 states including `needs_human_review` and `unresolved`).
- **Build and maintenance share the same queue.** A new site is just "site with a lot of pending items." An old site is "site that occasionally generates new items via discovery." Same dispatch loop processes both.
- **Atomic claim** via PostgreSQL `UPDATE ... WHERE status = 'triaged' RETURNING ...` — guarantees single-execution even with concurrent dispatchers. Cheap; no Redis, no broker locks.
- **Two-strike unresolved mechanism.** If a discovery check creates the same item twice within 7 days, the third instance is born with status `unresolved` and a `[unresolved after N attempts]` summary prefix. Stops fix-redetect loops in their tracks.

### 3.5 Kafka-based messaging discipline

- **Strict message structure**: headers (correlation_id, orchestration_id, request_id, message_type, action, responses_topic, sender), config (workflow), input_data. Three top-level keys, separation enforced.
- **Agents respond to the caller's responses topic, not their own.** This is how the parent's pending-request map matches replies to outbound calls. Caller-supplied response topic is a small, important detail that prevents an entire class of orchestration ambiguity.
- **Headers route, not paths.** The `responses_topic` header tells the agent where to publish; no service discovery, no DNS-aware routing.
- **Topics named deterministically** (`system.agent.{type}.requests` / `.responses`). Easy to reason about in logs, easy to filter in `kafkacat`/`kcat`.

### 3.6 Spec-driven authority chain

- **One spec, not two.** The classifier produces an aspirational spec including features the system can't yet build; those items have status `blocked`. The "build" is the non-blocked subset; the "dream" is the whole spec. Gap analysis is `WHERE status = 'blocked'`.
- **`feasibility-recheck` scheduled task** promotes blocked items to `planned`/`triaged` when the agent or component required becomes available — the system catches up to the spec's ambition without anyone re-running the classifier.
- **Per-aspect versioning** (`identity` / `classification` / `design_intent` / `content_direction` / `site_plan` / `seo` / `maintenance` / `direction` / `mission` / `roadmap` / `design_reference` / `resolved_composition`). Changing the design intent doesn't churn the site plan version. `is_current` / `superseded_at` columns give full audit trail.
- **Authority chain enforced socially.** Classifier writes intent. Planner writes implementation detail. Designer reads intent and produces CSS. Auditors read the spec and flag deviations — they don't override decisions. Documented as "audit enforces, doesn't override."
- **The classifier is the strategic brain — never shortcut.** Even when adoption has crawled an existing live site, the classifier still does vertical research; the adopted state describes where the site IS, competitor research informs where it COULD GO. The fidelity dial (`locked` / `high` / `medium` / `low` / `aspirational`) controls how aggressively first build extends the adopted baseline.

---

## 4. Engineering & Resilience Strengths

### 4.1 Designed for cheap ephemeral hardware

- Built on the assumption that pods will die. The whole stale-orchestration sweeper exists because that assumption is true.
- Spot-tier servers, three nodes, no premium Kubernetes managed offering. Whole stack runs on Civo / OVH-style commodity hosting.
- **pgbouncer in transaction mode** between every Go service and PostgreSQL. Application code uses `simple_protocol` or `cache_describe` to side-step prepared-statement-per-connection issues. Max 10 conns per service, pgbouncer pools 15 server-side per (user, db). Connection storms during pod restarts don't take the database down.

### 4.2 Stale orchestration sweeper (the thing that keeps it all running)

- Periodic (60s) DB sweep on every chassis pod, using `FOR UPDATE SKIP LOCKED` so multiple pods can run the sweep concurrently without coordination.
- For each expired `awaited_request`, classifies the failure into one of three modes:
  - **A. Child completed, response was lost** (most common — Kafka transient or pod restart between completion and produce). Synthesises a completion response from `final_result`, publishes to parent's responses topic.
  - **B. Child failed.** Forwards the failure; parent's error handling takes over.
  - **C. No child found / child still running.** Increments `retry_version`. Re-sends if < 3, marks expired and fails parent if ≥ 3.
- **Cascading stalls** are processed deepest-first (oldest `timeout_at`). One stall in a leaf agent does not propagate up the tree before being resolved at its source.
- **Job topics expired (1hr Kafka retention)** are handled by directly updating the parent's `orchestration_state` to advance past the awaited step — even when the messaging layer has lost evidence the child existed.

### 4.3 Race condition defence

- **`ErrLoopExpansionHandled` sentinel.** When a loop expands and immediately calls a fast-responding child, the response handler goroutine can advance state in the database before the recursive `continueExecution` returns. Without the sentinel, the outer loop reads stale step config and skips the rest of the iterations. The sentinel is caught by the outer loop and returns nil cleanly. Discovered the hard way; documented at length.
- **Atomic work-item claim.** No optimistic locking, no application-level mutexes — Postgres' MVCC plus a conditional `UPDATE ... WHERE status = 'triaged'` is enough.
- **Site-composition install in one transaction.** Four writes (css_themes / style_collections / sites.style_collection_id / site_specs.resolved_composition) commit together or not at all. A guarded UPDATE on `sites.style_collection_id` only commits if the value was NULL when the transaction started — handles two concurrent composition items for the same site without producing a half-built theme.

### 4.4 Bug discipline embedded in documentation

- **Every architectural rule is a scar.** The development guide contains 20+ documented post-mortems with rules attached. Examples:
  - `query_database` must use `$1` placeholders with `params` array — never Go template interpolation, SQL injection risk;
  - `api_key_env_var` is mandatory in every `ai_service` config — silent failure mode otherwise;
  - column names drift (`domain` → `pipeline`, `version_note` → `change_description`) — always run `\d table_name` against the live DB, never trust cached schema dumps;
  - LLM JSON output must be markdown-stripped before parsing — they wrap JSON in ` ```json ` fences even when told not to;
  - field name collisions in `ExtractActionInputs` nested lookups — optional fields named `content_data` or `domain` will silently match `site_record.content_data` even if the caller sent nothing.
- **18 different functions resolved dot-paths in nested maps** before being canonicalised into the `datahelpers` package. The "before adding a new utility, grep `datahelpers`" rule is enforced.
- **Backup naming convention** that prevents catastrophic loss: `agent_definitions_backup_YYYYMMDD_pre<NNN>` tied to the migration the backup is guarding. **Never `DROP TABLE IF EXISTS`** before re-creating — if a migration is re-run after partial failure, dropping replaces a known-good snapshot with a partially-modified one.

### 4.5 Observability built in, not bolted on

- **`llm_call_log`** captures every LLM call: agent_type, step_name, model, model_resolved, provider, prompt_template, prompt_rendered, response_text, input_tokens, output_tokens, latency_ms, success, error_message, retry_count, work_item_id, prompt_variant, vertical, rag_context_used. 90-day retention for successful, 180-day for errors.
- **`ai_endpoint_health`** table tracks up/down state per endpoint (Claude / CPU Ollama / GPU Ollama). Reactive (failure → mark unhealthy) plus active (hourly haiku ping ~$0.002/month) checks. Items don't get claimed if their handler's endpoint is unhealthy — they stay triaged and wait.
- **Site snapshots** (`site_snapshots`) capture the full state of a site (sites row + all current `site_specs` rows + pages + page_components + nav groups + nav items + site_components + git_commit_sha) as JSONB blobs. Survive row deletions and schema changes. Triggered on deploy, manually, pre-edit, scheduled, pre-revert.
- **Diagnostic queries baked in.** Every doc has SQL templates for "is this stuck?", "which step is awaiting?", "what does the workflow plan look like for this orchestration?", "which agents use this action?", "which model swaps are pending?"

### 4.6 Deployment pipeline

- **"Commit IS deploy."** Per-page git commits via the `git-adapter`. GitHub Actions fires on commit and pushes to Backblaze S3. Cloudflare for DNS only — no vendor lock-in.
- **Per-page atomic deploys.** A failure on one page doesn't block the others — each page has its own commit, its own GitHub Action run, its own S3 push. Per-work-item commit means per-work-item revert via `git revert`.
- **Multi-file commits** for component-with-JS pairs: HTML page + `/tools/assets/{function}.js` go in one commit so the asset is live the moment the page references it.
- **Deployment adapters as a future-proofing layer.** Current adapter: `git-adapter` → GitHub → Cloudflare Pages. Designed as a pluggable layer; future adapters for cPanel/FTP, WordPress REST API, Laravel Forge, Vercel, Shopify, Cloudflare Workers all share the same interface (auth, file push, build trigger, DNS, SSL, health check).

---

## 5. AI / LLM Quality Control Strengths

### 5.1 Three-layer quality gating, not one

- **Layer 1: Pre-generation (gates before tokens are spent).** `plan_sections` checks data availability for every section before the content writer runs. Sections with missing required data are deferred to `needs_human_review`, not fabricated. Pages deploy with whatever sections are ready.
- **Layer 2: Inline validation during build.** `validate_page_content` runs after every page generation: placeholder detection (NEEDS HUMAN REVIEW, Lorem ipsum, [INSERT], TODO), unrendered template variables (`{{.field}}`), cross-site contamination (wrong company name leaking from prior context), invented contact info, fabricated testimonials and statistics, links to non-existent pages. Algorithmic, deterministic, cheap.
- **Layer 3: Post-build audits (LLM judgement).** Three LLM auditors (visual-design / content-quality / strategic-review) run on a schedule, each capped at five findings per pass. The prompts forbid flagging issues already caught algorithmically and forbid removing features marked `must_have` in the human direction spec.

### 5.2 Audit pass cap and section locking

- **Three audit passes per site, then quiet.** `audit_pass_count` lives in `sites.settings.maintenance_profile`. Pass 1: top-5 per auditor → fix → verify → increment. Pass 2: same. Pass 3: fix or accept → increment. Pass 4+: improvement-loop exits in <1 second.
- **Auto-reset every 60 days** (or on direction change, or on N pages rebuilt in one cycle, or manually via dashboard "re-audit"). Sites breathe.
- **Three lock types** — `permanent` (manual unlock only — brand elements, legal disclaimers, hand-crafted content), `timed` (expires after N days, default 90), `review` (creates HITL review item on expiry). Discovery agents query `AND (locked_at IS NULL OR (lock_expires_at IS NOT NULL AND lock_expires_at < NOW()))` so locked components are skipped.
- **Audit token cost reduced ~65-70%** by these caps, measured against the pre-cap baseline.

### 5.3 Validation: catches the failure modes most LLM apps ignore

- **Cross-site contamination** — when the LLM borrows the wrong company name from earlier context.
- **Fabricated testimonials, statistics, percentages, case studies, named businesses** — explicit prompt rules ("NEVER invent fake people, NEVER invent specific statistics, NEVER claim to be the largest/best/number one") plus post-generation pattern matches.
- **Cliché bingo cards in prompts** ("Avoid: your pet deserves the best, we care about your pet, one-stop-shop") — pre-empting the LLM's tendency to lean on filler.
- **Schema-level commercial bias** — recognised explicitly as a failure mode. A schema that always contains `services`, `cta_style.verb_choices`, `persuasion_approach`, `social_proof_style` forces the LLM to populate them even for sites where they make no sense. Fields conditional on the determined commercial shape.

### 5.4 The fine-tuning flywheel (genuine differentiator)

- **Every LLM call captured** with full provenance. `llm_call_log` holds prompt, response, tokens, latency, success, work_item_id (so we can join "did the fix actually work?" back to the call that produced it), prompt_variant (A/B), vertical, rag_context_used.
- **First export run produced 1,949 training rows for `page-content-writer / iter_0` alone.** Versioned in `training_exports.runs` (UUID-named datasets supporting A/B comparison across retraining cycles).
- **Pure manual batch chosen over real-time streaming.** Snapshot boundaries matter — "the dataset we trained model X on" must be replayable.
- **RAG knowledge base on pgvector(768)** with `nomic-embed-text` embeddings via Ollama. Trigram fallback when Ollama is down. **Task prefixes (`search_document:` / `search_query:`) verified empirically as load-bearing** — without them, a Labrador chunk narrowly beat a French Bulldog chunk on a BOAS-specific query (wrong); with them, French Bulldog won by a 5× wider gap. Documented as mandatory in production code.
- **Three improvement channels operating independently**: (a) RAG knowledge base for short-term retrieval-augmented improvements (no training needed); (b) prompt evolution via the prompt_variant column for A/B; (c) LoRA fine-tuning for long-term cost reduction. Each is valuable alone; they compound together.
- **Cost story.** Current observed: ~$120 per 4 domains (~19.6M input tokens, 4.2M output) over 2-3 weeks. Anthropic Batch API: 50% off input AND output, stacks with prompt caching (up to 95% combined). Universal LLM work queue (`llm_batch_queue`, `llm_batch_config`, `llm_batch_agent_config`) routes audits and maintenance through batch, leaves user-facing builds on direct API. Three-gate check (global / agent / provider) means every batch-eligible step has a sync-fallback path so the same workflow runs whether batch is on or off.

### 5.5 Taming hallucination through architectural discipline

- **Strict reuse-before-create rule** (Step Zero of the development guide). Every new agent, every new action, every new utility starts with grep across `agent_definitions`, `registry.go`, the actions package, and gate functions like `isStorageEnabledAgent`. Documented examples of three-hour rebuilds of things that already existed under slightly different names.
- **Modular relentless documentation** as guardrails inside prompts. The development guide and contracts doc are loaded into context for every code change. Rules are framed as "check this before writing SQL" / "check this before adding a utility" — actionable, not aspirational.
- **Algorithmic checkers paired with LLM checkers** — for any quality concern, the algorithmic check runs first (cheap, deterministic, catches obvious cases) and feeds its findings into the LLM prompt as context, so the LLM doesn't re-flag what's already known and uses the algorithmic context to inform its subjective judgement.
- **"LLM for reasoning, Go for extraction."** Don't ask an LLM to read hex values when a regex can do it. Don't pipe content through an LLM just to get it back unchanged. The adoption pipeline is the canonical example: Go extracts the design fingerprint (colours, fonts, CSS variables, layout patterns), the LLM produces a semantic design brief from that fingerprint plus identity, then Go renders CSS from the brief.

---

## 6. Cost & Scale Discipline

- **Audit pass caps + finding caps + section locking** cut audit token spend ~65-70%.
- **Anthropic Batch API integration** (50% discount) for non-blocking calls (audits, maintenance, content rewrites). Universal queue with provider-agnostic routing — same queue can dispatch to Anthropic batch, Anthropic sync, GPU Ollama, CPU Ollama, image generation services.
- **GPU on-demand** rather than always-on. ThunderCompute single-H100 ($1.38/hr) for Llama 3.3 70B. CPU Ollama for embeddings (`nomic-embed-text`) and small-model tasks (`mistral-small3.1`). GPU endpoint health is the de-facto scheduler — when up, items flow; when down, items wait without retrying.
- **Model assignment matched to task value.** Claude Opus for chief-strategist (one call per domain). Claude Sonnet for site-classifier and page-content-writer (high-volume, quality matters). Claude Haiku for ambiguous Companies House review (cheap, structured). Mistral-Small-3 on CPU Ollama for briefing-agent (lowest stakes, structured Q&A, swap-tested).
- **`llm_batch_queue.sync_executed` rows** record cost data even when batch is off — so you can see exactly what the rollout would save before flipping the gate.
- **Operational thrift.** External MySQL on Clook shared hosting (£~5/mo) for auth. Backblaze S3 for static delivery. Cloudflare DNS only. Spot-tier instances. The whole production stack costs in the order of low tens of pounds per month plus LLM spend.

---

## 7. Pipeline Capabilities (what it actually does end-to-end)

### 7.1 Domain → live site

- Intake orchestrator → classifier (with research) → briefing agent (HITL questionnaire if needed) → site-work-orchestrator → planner writes work items → image-build-handler generates logo and hero → site-design-planner resolves palette+layout+typography composition → webdesign-agent renders CSS → page-build-handler writes each page (one per work item) → rerender-pages assembles → git commit → GitHub Actions → S3 → live.
- **Composition resolution is its own handler** (`site-design-planner`). Six deterministic Go steps, no LLM. Owns palette/layout/typography selection; webdesign-agent owns CSS rendering and the design overlay. Separation prevents the bug where the renderer used a default theme literal regardless of what was chosen.
- **Composable themes.** `palettes` + `layouts` + `typography_sets` are independently versioned; a `css_themes` row composes one of each. The library can grow in three orthogonal directions instead of being a flat catalogue of "theme variants."

### 7.2 Adoption (takeover of an existing live site)

- Crawl source URL via Firecrawl → extract design fingerprint with Go (colours, fonts, CSS variables, layout patterns; external CSS fetched and merged) → produce semantic `design_intent` from fingerprint + identity via LLM → write specs (`identity`, `classification`, `archetype`, `design_intent`, `design_reference`, `content_direction`) → queue `needs_composition` and content-page work items → site rebuilds to match adopted character.
- **`design_reference` is history, `design_intent` is direction.** Reference records concrete extracted values; intent describes character with reference values as guidance. Webdesign-agent reads intent (creative freedom); audit checks against intent (not reference). Evolution happens by updating intent.
- **Fidelity dial** (`locked` / `high` / `medium` / `low` / `aspirational`) — controls how aggressively first build extends adopted baseline. `locked` reproduces faithfully; `aspirational` lets the classifier's research push the design toward where the niche is going.

### 7.3 Improvement loop (post-build maintenance)

- Heartbeat scheduler picks the least-recently-audited deployed site → discovery agents (algorithmic) flag obvious problems (broken nav links, placeholder contact, undeployed assets, missing CSS, hardcoded section colours, empty sections, orphan pages, empty blog) → LLM auditors (top-5 findings each) flag subjective issues (tone, content gaps, design coherence, strategic alignment) → triage promotes findings to triaged work items → dispatch loop processes them → rerender-pages re-assembles touched pages → git commit → S3.
- All findings include `current_value`, `acceptance_test`, `suggestion`, `max_fix_attempts`. Verification is cheap: a structured criterion to check, not "is the audit happy now."
- **Pass-count cap + auto-reset** means sites cycle between active improvement and quiet stasis on a 60-day rhythm.

### 7.4 News and content feed

- Per-site ingestion from RSS, news search, scrape, and LLM-as-source (xAI Responses API for web+X search). Items land in `content_feed_items`.
- **Two-pass design.** Run N triages items from Run N-1 and dispatches new ingesters; new items are scored by Run N+1. No synchronous wait on async ingestion.
- **Source-diversity render.** `ROW_NUMBER() OVER (PARTITION BY source_id)` interleaves sources so the homepage is not an OilPrice monoculture.
- **Triage by relevance + credibility.** LLM scores each item against the site's identity/content_direction; low-relevance items are dropped silently.
- Renders to `/data/latest-news.json` (homepage carousel) and `/data/news-archive.json` (dedicated `/news.html` page). Client-side JS reads JSON; no backend.

### 7.5 Tools (interactive components)

- **Tool suggester** evaluates what tools would benefit a site (LLM with site context) → routes library tools to `tool-deployer` (fork-on-deploy) or novel tools to `tool-generator` (LLM-created HTML+JS+CSS).
- **Cross-linking.** Suggester returns 1-3 `related_pages` per tool; `create_tool_cross_link_items` queues `content_rewrite` work items for those pages with natural-language guidance for the content writer ("mention this calculator in context, link to it").
- **Tool health audit** (Tier 1 structural + Tier 2 LLM code review) catches degradation; deployed tools are queueable for improvement.
- **`forked_from` lineage** — bad library change can't break ten sites. Each site owns its fork.
- **JS extracted to asset files** at `/tools/assets/{function}.js`. `<script>` tags in component templates are mechanically pulled out by `separateInlineJS()` before storage; pages serve the asset, not inline JS.

### 7.6 Companies House enrichment (vet vertical so far)

- Bulk collect SIC 75000 companies (5,780 active) → two-pass local matching (postcode+name then trigram name-only) → LLM disambiguation of ambiguous matches (Claude Haiku, ~$0.05 per run) → fetch officers / PSC / accounts (iXBRL) for confirmed matches → derive succession risk signals.
- **23.2% match rate** (634 confirmed of 2,730 businesses) plus 83 pending/uncertain for human review, plus ~5,100 unmatched CH companies as discovery candidates.
- Runs on a single `business-intel` static pod sharing one image across all CH agent types — Kafka routing keys do the dispatch.

### 7.7 Vet practice discovery

- Postcode-by-postcode sweep → discovers practices via web search → dispatches verifier per practice → enriches each → results land in a JSON export consumed by vetcomparison.uk client-side.
- **5,000 practices at ~500 bytes each ≈ 2.5MB JSON.** Client-side filter and search; no backend at all. Demonstrates the "Tier 1 — Static + client-side JS" pattern.

### 7.8 Medicine pricing pipeline

- URL discovery via Firecrawl `/map` and via category-page scraping → price scraping per pharmacy → JSON export → git commit → S3.
- Same orchestration shape as everything else (orchestrator wrapper spawns specialist worker in dedicated pod). The pipeline is generic; the verticals are configuration.

---

## 8. Human-in-the-Loop / Governance

- **Three direction channels** with different persistence models: work item request (one specific thing, until completed), direction update (permanent until human changes it, influences classifier and auditors), reference suggestion (feeds next planning cycle).
- **Direction spec is pinned.** Agents read it but cannot supersede it. Feature marked `must_have` will not be flagged for removal by content auditor; strategic review will not suggest dropping it.
- **Direct admin edit + auto-lock + rerender** = typo to live in seconds, no LLM round trip.
- **Three-tier component lock** (page_components / site_components / sites) — the first locks one section on one page; the second locks header / footer / CSS site-wide; the third freezes the entire site (all automated activity stops, including discovery sweep).
- **Edit doesn't require unlock.** Lock semantics: "human controls this," not "read-only." Locked component still gets edited; lock stays on; `locked_at` refreshes.
- **Spec propagation is explicit.** Admin clicks Propagate, sees what work items will be created across pages, confirms.
- **Page growth budget** prevents bloat: weekly rolling windows for content pages (default 3/wk) and blog posts (default 2/wk), with a hard absolute_max ceiling. Rate-limited additions don't fail — they're marked `blocked` (retryable next week).
- **Content briefs control regeneration.** Every page_component carries a `content_brief` JSONB with purpose / tone / guidance. Edit the brief, click Regenerate, the page-content-writer rewrites with the new brief as instructions.
- **Admin dashboard already shipped** for: site list, work item triage, direct edit (HTML + fields + brief), spec editor with pin/unpin/propagate, suppression and restore, media browser with reference counts, page purpose editor with regen, growth budget config.

---

## 9. Production Evidence (what's actually live, with what)

| Site | Pattern | Notable |
|---|---|---|
| `vetcomparison.uk` | Tier-1 static + client-side JSON search | ~5,000 vets, ~2.5MB JSON index, lazy-loaded medicine data by letter chunk |
| `gaswholesalers.com` | Brochure + tools + news feed | Gas Unit Converter (novel LLM-generated tool), Fuel Cost Estimator, news pipeline with cross-linked articles |
| `finetuning.uk` | Marketing + tools | AI Readiness Quiz (novel), AI Time Savings Estimator (library fork) |
| `leopardessconsulting.co.uk` | Brochure | Two deployed tools, professional-dark theme |
| `ai-agent-orchestration.com` | Tools + content | AI Agent ROI Estimator (novel), 9 cross-link items |

- All running concurrently on the same three-node cluster, all self-maintaining, all generating LLM call logs that feed the training flywheel.
- **57+ rows in `llm_call_log` from the first verification run** of the deployed call-logging code; production volumes are now in the multi-thousands.
- **Every site has site snapshots** taken on each deploy, retainable for point-in-time revert.

---

## 10. Honest Weak Points / Trade-Offs (have ready for sceptical questions)

These are the things to be straight about — both because they're real and because volunteering them in a pitch is far stronger than being caught by them.

### 10.1 Solo development with AI assistance

- One developer plus LLM-aided coding. No external code review, no second pair of eyes on architectural decisions. Every "rule" in the docs is a scar — found through pain rather than anticipated through team review.
- LLMs hallucinate variables and overwrite core architecture. Mitigated by relentless modular codebase summaries acting as guardrails in prompts, and by separate algorithmic and LLM-based checker agents that verify outputs before integration. But this is mitigation, not elimination — the silent column-rename bugs (`domain` → `pipeline`, `version_note` → `change_description`) sat undetected for weeks.
- **Honest framing:** "I built this knowing the LLM would lie to me about the codebase, so the discipline is everywhere — but the discipline is mine, not a team's."

### 10.2 Documentation drift

- 28+ numbered docs plus FOCUS / HANDOFF / FUTURE / PATCH series. They're consolidated periodically (the index document tracks what's absorbed where) but in between consolidations, multiple docs can describe the same area with subtly different states.
- Living documents — not all flagged as deprecated when superseded. Future sessions check docs against current behaviour because drift is real.
- **Honest framing:** "The docs are the project's memory. I'd rather have them messy and over-detailed than tidy and incomplete."

### 10.3 Schema drift

- Live PostgreSQL has been the source of truth for column names because cached dumps go stale silently. The development guide enforces `\d table_name` before writing any SQL or Go. Two real cases (years' worth of writes silently failing in best-effort code paths) happened before the rule was fully internalised.
- **Honest framing:** "The lesson cost me weeks of invisible failure. Now I treat any 'best-effort' operation as 'silent no-effort until proven otherwise' and there's a smoke-test query attached."

### 10.4 Race conditions are real, even where defended

- The `ErrLoopExpansionHandled` sentinel was introduced because the original loop-expansion code raced with fast-responding child agents. Multiple sequential loops in one workflow needed the `loop_metadata` shared-key caveat documented. Zombie dispatch loop pods (loop-expanded steps lost from `workflow_plan` during concurrent state updates) needed a 30-minute reaper threshold as mitigation rather than a full fix.
- **Honest framing:** "Concurrency in distributed orchestration is hard. I've found and fixed a class of these; I'd budget more for the next class as new patterns emerge."

### 10.5 Migration stories are partly incomplete

- `pageflow-builder` predates the unified `site_specs` design and writes to `sites.content_data`. `intake-orchestrator-v2`, `site-classifier-v2`, `site-planner-v2` exist alongside their v1 predecessors, with `read_site_spec` providing a fallback. Backfill scripts exist but haven't been run on every legacy site.
- **Honest framing:** "Migration is incremental. I keep both pipelines working with fallback reads rather than forcing a flag day. The cost is more code surface; the benefit is no production outage during transitions."

### 10.6 Multi-tenancy is shared-resource

- All sites in one `clients_db`, scoped by `site_id`. Per-tenant compute or data isolation is listed as Phase 4 work, not done. A bad agent definition affecting one site could in principle affect runtime resources that other sites share.
- **Honest framing:** "This is a single-customer architecture today, plus internal sites. Multi-tenant isolation is a known phase ahead, with the dispatch loop generic and the agent boundaries already clean enough to support it."

### 10.7 GPU endpoint and on-prem inference are partial

- Llama 3.3 70B on H100 tested and quality-assessed (8/10 vs Claude's 9/10 for content; 7/10 vs 9/10 for design); ThunderCompute platform issues (2-GPU bug, single-H100 only) and on-demand spin-up logistics mean it's not always-on.
- CPU Ollama is up and doing useful work (embeddings, briefing extraction with Mistral-Small-3) but most production calls still route to Claude.
- Fine-tuning pipeline has data export and RAG done; LoRA training scripted but awaiting first GPU run; eval harness paused.
- **Honest framing:** "The path is laid down. The cost-shifting hasn't fully landed yet — but every component along the path has been built and tested in isolation."

### 10.8 Some pipelines are layered defence rather than clean

- LLM JSON output stripping is defensive layer-after-layer (markdown fences, escaped quotes in SVG paths, occasional fully malformed responses). Template rendering fails silently on nil values — every `{{range}}` needs a `{{if}}` guard. Handler vs specialist boundary (handler must persist its outputs) was discovered the hard way after `page-content-writer` returned HTML that was never saved.
- **Honest framing:** "These are LLM-orchestration realities. The defences are documented, the post-mortems are searchable, and the rules survive across sessions."

### 10.9 No formal automated test suite

- Verification is via SQL queries after deploy, manual smoke-test scripts, and the audit loop itself catching regressions. No mention of a Go test suite for actions or a CI gate that runs the full pipeline against a sandbox site.
- **Honest framing:** "The system is its own test harness — running improvement loops on every deployed site every few days catches regressions in production. A formal CI gate is on the list, not done."

### 10.10 Cost ceiling unproven at scale

- Projected $15-30k per cycle at 2,000 domains without optimisations. Batch API rollout is incremental; gates are flipped on a few agents only.
- **Honest framing:** "I've measured the unit economics on small numbers. Scaling is a financial as well as engineering problem; the routes to dropping per-domain cost — batch API, fine-tuned local models for the high-volume agents, RAG instead of long prompts — are all in motion, just not all on."

### 10.11 Limited dynamic-application support today

- Tier 1 (static + client-side JS) is fully supported. Tier 2 (forms, simple APIs, agent-managed backends) is planned. Tier 3+ (full apps with persistent state, real-time, multi-tenant SaaS scaffolding) is roadmap.
- WordPress / Laravel / Next.js / Vercel / Cloudflare-Workers adapters are designed (the publishing-adapter abstraction exists) but only `git-adapter` is implemented.
- **Honest framing:** "The framework can build any site that fits Tier 1 today. The architecture admits the higher tiers without rework. I'd want a paying use case before building Tier 2 in detail."

---

## 11. Pitch Talking Points (anticipated questions)

### "Why not just use LangGraph / CrewAI / AutoGen?"

- Those are single-process Python frameworks. The agents are object instances in one Python interpreter. Hierarchical sub-agent spawning is nested function calls in the same process; one stuck task kills the whole graph; restart loses all in-flight state.
- This framework treats every agent as a Kubernetes Job pod with its own Kafka topics and Postgres connection. Death of any agent is a local event. Spawning is real provisioning. Orchestration state lives in a database, survives restart, is queryable.
- The trade-off is honest: LangGraph is faster to prototype with for one developer on one laptop. This framework is built for "I want hundreds of these running and I don't want to babysit them."

### "Why Go and not Python?"

- The orchestration layer needs to be a server you trust. Go's static typing, single-binary deployment, low memory footprint per agent (each pod is ~50MB resident), and goroutine-based concurrency match the architecture.
- The LLM calls themselves go to whatever API; the agent is the harness, not the model.
- Python is fine for the LLM calls and would have been fine for the chassis too. Go was the right call for predictability under load and for K8s native deployment.

### "How do you stop runaway agent loops?"

- Audit pass cap (3 per site).
- Finding cap per auditor (5 per pass).
- Two-strike unresolved mechanism — same item rediscovered twice in 7 days becomes `unresolved` status, stops being claimed.
- Section locking with timed/permanent lock types.
- Stale orchestration sweeper kills anything awaiting > 30s past timeout.
- Site lock — admin can freeze all activity on a site instantly.
- Page growth budget — weekly content/blog quotas with absolute_max ceiling.
- All of these were responses to actual loop failures, not theoretical ones.

### "What's the most painful bug you've fixed?"

- Pick one of: silent INSERT failures from column renames (`version_note` → `change_description` cost two years of `component_versions` history); the loop expansion race condition that needed `ErrLoopExpansionHandled`; the page-content-writer specialist-vs-handler boundary that produced "deployed" pages with no content; the dispatch loop's input_mapping `?`-suffix discipline that took several debug sessions to crystallise.
- All documented in the development guide. Each one earned a rule.

### "How does this make money?"

- Two layers, distinct economics.
- **Layer A (internal flywheel).** The factory builds and maintains a portfolio of sites we own — vetcomparison, gaswholesalers, finetuning, etc. Each earns through the revenue model the classifier picked (affiliate, advertising, lead generation, services). The factory is the asset; the sites are the cashflow.
- **Layer B (productised framework).** The framework can build factories for other people. Concrete first product: finetuning.uk as a hosted RAG / fine-tuning service for technical-adjacent SMEs that don't want to run their own infrastructure. Reuses the agent framework, the LLM call logging, the model swap functions, the universal queue.
- Layer A pays the bills now (or will, soon). Layer B is the upside.

### "What would scaling look like?"

- Compute: trivially horizontal — add nodes, K8s scheduler picks them up, more concurrent agents, no architectural change. The current cluster is at <20% utilisation across the three nodes most of the time.
- Database: pgbouncer is the bottleneck for small-conn-count Go services; vertical scale of the Postgres pod gets us several orders of magnitude before sharding by site_id becomes necessary.
- LLM cost: the dominant cost and the scale-blocker. Mitigations are in flight (batch API, fine-tuned local models for high-volume agents, RAG to shrink prompts).
- Operationally: most painful would be onboarding a second engineer. The codebase is single-author and the development guide is the only onboarding artefact today.

### "What did you learn that you'd do differently?"

- Build the canonical utility set (`datahelpers`) first, before any agent. The 17-different-functions-resolving-the-same-dot-path mess was the cost of growing the helpers organically alongside the agents.
- Standardise the schema source of truth from day one. Cached dumps are dangerous; live `\d table_name` is not optional.
- Build the snapshot/revert path before the first model swap, not after. Snapshots are cheap; learning you needed them after losing data is expensive.
- Treat "best-effort" as "silent no-effort until proven otherwise" — every best-effort write needs a smoke test.
- Write the "every agent is an orchestrator" rule into the very first code, not after retrofitting two non-orchestrator agents.

---

## 12. Tech Stack — Detailed

| Area | What I'm using | Why |
|---|---|---|
| Orchestration language | Go (1.22+) | Single binary deploy, static types, goroutines, low per-pod memory |
| Container runtime | Docker | Standard |
| Orchestrator | Kubernetes (jobs, deployments, services, RBAC, kustomize, port-forwards) | Native distribution, mature primitives, cheap on commodity hosting |
| Inter-agent messaging | Kafka (Strimzi cluster, 3-broker combined-pool) | Durable, partition-ordered, decoupled |
| Primary DB | PostgreSQL 16 (in-cluster) with pgbouncer (transaction mode) | JSONB for spec/workflow/result storage; pgvector for embeddings |
| Auth DB | MySQL 8 (Clook shared hosting) | Pragmatic — auth is a small surface, external host keeps it isolated |
| LLM (cloud) | Anthropic Claude (Opus 4.6, Sonnet 4.6, Haiku 4.5), with Batch API integration | Highest quality for design, content, and reasoning |
| LLM (on-prem) | Ollama on CPU (mistral-small3.1, nomic-embed-text), GPU on demand (llama 3.3 70B) | Cost shifting, training-data destination |
| Embeddings | nomic-embed-text via Ollama, 768-dim, with task prefixes | Verified empirically; trigram fallback when Ollama unavailable |
| Crawling / scraping | Firecrawl (via adapter), webscrape adapter for direct URLs | Adoption pipeline, news ingestion |
| Deployment | git-adapter → GitHub Actions → Backblaze S3 | Commit-is-deploy; per-page atomic; vendor-neutral |
| DNS | Cloudflare (DNS only, no Pages lock-in) | Cheap; no vendor coupling on the application layer |
| Provisioning | Terraform (cluster, nodes, storage), Kustomize (k8s manifests) | Reproducible, auditable |
| Observability | `llm_call_log`, `ai_endpoint_health`, `agent_error_log`, `orchestration_states`, site snapshots, kubectl logs | Not glamorous but it's all in the database — queryable, joinable |

---

## 13. Stuff to Avoid Saying

- "It works perfectly." It doesn't, no system does, and the framework's documented bug history is one of its strengths.
- "It scales infinitely." The architecture admits horizontal compute scaling; the LLM cost ceiling is the real constraint, and it's honest to say that.
- "The agents reason." The agents orchestrate. Reasoning happens in the LLM calls; the framework's value is in everything around the LLM call — what to call, when, with what context, what to do with the response, how to recover when things go wrong.
- "Every agent is autonomous." Every agent is an orchestrator. Autonomy is governed by specs, locks, growth budgets, and audit caps — deliberately bounded, not absent.
- Marketing phrases like "AI-powered," "next-gen," "revolutionary." The system speaks for itself in technical and product terms; salesy language undermines the engineering credibility.

---

## 14. Quick Reference Quotes (for use in writing)

- "Every agent is an orchestrator."
- "One spec, not two — the dream is the full spec, the build is the non-blocked subset."
- "Audit enforces, doesn't override."
- "LLM for reasoning, Go for extraction."
- "Design reference is history, design intent is direction."
- "The sites improve themselves. The framework improves itself by training on what the sites generated."
- "Every rule in the development guide started life as a bug."
- "Built for cheap ephemeral hardware on the assumption that pods will die."
- "Commit is deploy."
- "Adoption is an input to strategy, not a substitute for it."
- "Default-to-brochure bias is a failure mode, not a safe fallback."
