# FOCUS — Finetuning Flywheel & finetuning.uk Service

Living document. Consolidates what we have planned for the AI training flywheel
(internal: our own models get better over time as we build sites) and opens the
separate question of whether finetuning.uk becomes a **self-service fine-tuning
product** for external users.

Last touched: 2026-04-21

---

## 1. Two separate concerns, one doc

| Concern | What it is | Owner | State |
|---|---|---|---|
| **A. Internal flywheel** | Our pipeline produces training data as a byproduct of building sites. We periodically fine-tune local models on that data, swap them in for Claude calls where quality holds, and drop API costs. | Agent orchestration system | Infra mostly in place, training loop not yet run |
| **B. finetuning.uk product** | External users upload their own data and fine-tune a model through a UI on finetuning.uk. Likely monetised. | New product surface | Not started. Questions to answer before scoping. |

A and B share plumbing (Ollama, GPU scheduling, training data export, Unsloth)
but have very different surfaces, failure modes, and economics. Keep them
distinct in our heads.

---

## 2. The internal flywheel — what exists

### 2.1 Data capture

Every LLM call in the system writes to `llm_call_log` (migration 081, flywheel
columns added in 085):

| Column | Purpose |
|---|---|
| `agent_type`, `step_name` | Which agent, which prompt — the join key for "what to train" |
| `model`, `model_resolved`, `provider` | What was actually called |
| `prompt_template`, `prompt_rendered`, `response_text` | Raw training pair |
| `input_tokens`, `output_tokens`, `latency_ms` | Cost/perf signal |
| `success`, `error_message`, `retry_count` | Quality filter |
| `work_item_id` | Link call → outcome (did the fix work?) |
| `prompt_variant` | A/B on prompt versions |
| `vertical` | Per-industry slicing |
| `rag_context_used` | Was RAG injected into this prompt? |

Write path lives in `ai_actions.go` → `LogLLMCall` (fire-and-forget goroutine,
5s timeout, never blocks the workflow). Flywheel context is extracted from
`CollectedData` inside `ExecuteLLMPromptAction`.

Retention: 90 days for successful calls, 180 days for errors. Cleanup function
exists (`cleanup_old_llm_logs`) but nothing schedules it yet.

### 2.2 Knowledge base (RAG — the other half of the flywheel)

`knowledge_base` table (migration 082) — pgvector(768) for nomic-embed-text.
Shared resource, any agent can read via `rag_lookup`, any can write via
`rag_index`. Trigram fallback when Ollama is down.

Best-practices doc says: **filter first by metadata, then rank by similarity.**
Metadata fields that should be on every row: `vertical`, `component_type`,
`content_type`, `source`, `source_quality`. Without these, a vet example gets
retrieved when writing gas-wholesale copy.

RAG knowledge is the **short-term** lever (useful immediately, no training
needed). Fine-tuning is the **long-term** lever (needs 200+ examples, a GPU,
and evaluation).

### 2.3 Inference routing

`ai_endpoint_health` table (migration 085) tracks three endpoints:

| Name | URL | Role |
|---|---|---|
| claude | `https://api.anthropic.com/v1/messages` | Default for high-quality, low-volume |
| cpu-ollama | `http://ollama-adapter.ai-persona-system.svc.cluster.local:11434` | Embeddings + small models (mistral-small3.1, nomic-embed-text) |
| gpu-ollama | `http://ollama-gpu.ai-persona-system.svc.cluster.local:11434` | Llama 3.3 70B, future LoRAs — currently DOWN, not always-on |

Healthy endpoint → claim flows. Unhealthy → items wait (back-to-triage). GPU is
either healthy (work flows) or not — no separate batch scheduler needed.

### 2.4 Model swap / revert

`snapshot_agent()`, `swap_agent_model()`, `revert_agent()` (migration 083) —
per-agent per-step, safely snapshot before swapping an `ai_service` block in
`agent_definitions.default_config`. The nuclear option (full-table backup)
still exists alongside.

### 2.5 Fine-tuning path (not yet executed)

```
llm_call_log (200+ success rows for a given agent_type + step_name)
    │
    ▼
training_data_export.sql  (Alpaca or ChatML JSONL)
    │
    ▼  scp to GPU host
GPU host (RunPod / Lambda / ThunderCompute H100)
    │
    ▼  unsloth: QLoRA on llama-3.1-8b-instruct-bnb-4bit
LoRA adapter (~3-50MB)
    │
    ▼  save_pretrained_gguf with q4_k_m quantisation
GGUF file (~4GB for 7-8B Q4)
    │
    ▼  kubectl cp → ollama-gpu PVC + ollama create
New Ollama model, e.g. site-classifier-v1
    │
    ▼  swap_agent_model('site-classifier', 'classify_site', {provider: ollama, ...})
Live. A/B by running both and comparing, revert_agent() if worse.
```

Cost estimate from `018_canine_biology.md`: **£35-95** for first full pass
(research → text LoRA → image LoRA).

### 2.6 Fine-tuning candidates (priority order)

| Agent | Output type | Why local wins |
|---|---|---|
| knowledge-extractor | Structured JSON from prose | Runs repeatedly, well-defined schema, data accumulates during research |
| site-classifier | JSON classification | Every build hits it, output small |
| vet-practice-verifier | JSON boolean + evidence | High volume, yes/no + short extract |
| briefing-agent | JSON questionnaire from raw input | Lowest risk — swap to Mistral Small 3 already queued |
| domain-analyst | JSON metadata | Structured extraction |
| content-researcher | Search queries | Short query strings |

**Not good candidates**: page-content-writer (long creative output, quality
matters), visual-design-auditor (judgement), chief-strategist (structural
reasoning, one call per domain — worth the Claude cost).

---

## 3. Three improvement channels — they compound

From `009_model_infrastructure.md`, decision 10:

1. **RAG** — inject verified knowledge into prompts. Immediate effect. No training.
2. **LoRA** — train a local model to replicate an agent's task. Medium-term. Reduces per-call cost.
3. **Prompt evolution** — `prompt_variant` column lets us A/B prompts and pick winners. Ongoing.

Each is useful alone. Together they compound: good prompts with good RAG give
you the best training data, which produces the best fine-tuned model, which
needs a good prompt and good RAG to perform well.

---

## 4. Current state — what to do next on the internal flywheel

### 4.1 Ready-to-go
- [x] llm_call_log populating (flywheel columns wired through `ai_actions.go`)
- [x] Swap / revert functions deployed
- [x] Endpoint health routing deployed
- [x] Ollama CPU adapter up (nomic-embed-text + mistral-small3.1)

### 4.2 Open tasks (small, near-term)
- [ ] `agent_type` column empty in `llm_call_log` — `params.AgentType` not
      reaching `LogLLMCall`. Low priority but blocks clean training-data slicing.
- [ ] Schedule `cleanup_old_llm_logs()` (pg_cron or maintenance agent) —
      table will grow to ~1GB/month at current rates.
- [ ] Ensure `work_item_id` flows through orchestration context so "did the fix
      work" joins become possible.
- [ ] Periodic REINDEX of the ivfflat vector index on `knowledge_base`.

### 4.3 Open tasks (larger, flywheel completion)
- [ ] First real export: pick the agent with the most successful rows in
      `llm_call_log` — probably `page-content-writer` or `site-classifier` —
      and produce a JSONL file. Just the export, no training yet.
- [ ] Dry-run Unsloth end-to-end on a borrowed GPU to validate the flow
      (we've never actually run it).
- [ ] Write an A/B scoring script — produce outputs from Claude and the local
      model for N examples, let a reviewer (human or Claude judge) rank pairs.
- [ ] Bring the GPU Ollama endpoint up reliably (`ollama-gpu` currently DOWN).
      ThunderCompute H100 works but is stopped and manual to start.

### 4.4 Decisions to make
- **Which model as the fine-tune base?** Llama 3.1 8B is what the canine-biology doc assumes. Qwen/DeepSeek/Mistral are alternatives. Pick one to standardise on so we accumulate ops experience with a single stack.
- **Where does training happen?** ThunderCompute single-H100 worked in testing. Other options: RunPod, Lambda, local 3090/4090. Trade-off: spot cost vs queue time vs ops complexity.
- **On-cluster training pod or manual?** Short-term: manual on rented GPU. Long-term: a `training-job` agent that does export → upload → train → pull GGUF → install on ollama-gpu automatically. That's after A/B confirms the path works.

---

## 5. Product pivot: finetuning.uk as a self-service platform

**This is a product decision, not a technical one.** The technical bits are
shared with the internal flywheel, but building an external product is a
different company's worth of surface area.

### 5.1 What would "users can fine-tune from finetuning.uk" require?

| Layer | What's needed |
|---|---|
| Identity | Multi-tenant auth, org/team model, API keys, billing identity |
| Data | Upload (CSV/JSONL), preview, validation, PII scanning, quota, retention, deletion (GDPR) |
| Training | Base model picker, hyperparameter picker (or opinionated defaults), job queue, GPU scheduling across tenants, eval metric display, checkpointing |
| Inference | Per-tenant endpoints, cold-start handling, rate limits, usage metering |
| Billing | Stripe, VAT, invoices, usage reconciliation |
| Ops | Support, docs, status page, abuse prevention, content moderation |
| Legal | ToS, privacy policy, DPA, UK/EU data residency commitments, base model licence compliance |
| UI | The actual finetuning.uk frontend — dashboards, job monitoring, chat playground |

### 5.2 Competitive context (April 2026)

| Player | Angle |
|---|---|
| Fireworks AI | Cheap (~$3 minimum training fee), fast LoRA serving, strong API DX |
| Together AI | 100B+ models, LoRA serving, checkpoint resume, enterprise trust |
| Nscale | SFT-first, speed, UK-presence |
| OpenAI / Vertex / Azure OpenAI | Managed on proprietary models, enterprise default |
| H2O LLM Studio | No-code UI, self-hostable, includes distillation |
| Predibase / Modal / Replicate / RunPod | Various tiers |
| Unsloth Studio (+ Rafay) | Self-host pattern — this is the engine most serverless services wrap |
| Azumo / Accenture / Deloitte | Consulting model — done-for-you |

### 5.3 Where differentiation could come from

Honest read: it is **not** a technically-novel product. Anyone can wrap Unsloth.
Differentiation has to come from positioning, not engineering.

Plausible angles:
- **UK / EU data residency** — easy story, legally meaningful, hard for US competitors to match without setting up local infra.
- **Opinionated simplicity** — 3-click fine-tuning for non-ML users. One base model choice, one dataset format. "If you know how to use Excel, you can fine-tune."
- **Vertical templates** — pre-built datasets and prompts for legal, medical, accounting, customer support. Buy a fine-tune recipe, add your data, go.
- **Self-improvement loop as a feature** — fine-tuned models whose quality keeps improving as users interact with them. This is **exactly** the flywheel we've built internally. It's our edge.
- **Free tier funded by data contribution** — users get free training in exchange for (opted-in) contribution to a shared knowledge base. Only works ethically with genuine opt-in and clear terms.

### 5.4 Concrete risks / things that will bite

1. **GPU supply is lumpy.** At the scale of "dozens of concurrent training jobs", ThunderCompute-on-demand is not reliable enough. Reserved capacity commits you to cost whether or not users arrive.
2. **Multi-tenant isolation is not optional.** A training job reading another tenant's data is an extinction-level bug.
3. **Users upload bad data → bad models → blame you.** Needs strong pre-flight validation and clear expectation-setting.
4. **Base model licences are hairy.** Llama has MAU thresholds; some bases forbid commercial fine-tuning outputs; Mistral varies by model.
5. **Inference hosting is its own product.** Selling fine-tuning without hosting means "download a GGUF and good luck". Selling hosting means running a GPU fleet 24/7 with cold starts solved.
6. **Billing against GPU seconds is error-prone.** Users will dispute. Need clear usage meters and generous grace.
7. **Support load.** People break things in creative ways. A support tier must exist from day one.
8. **Opportunity cost.** Every engineering week on finetuning.uk is a week not on the agent orchestration system (the primary product). Unless the sites-as-a-service business funds it, the pivot should be justified on its own economics.
9. **Claude/OpenAI get cheaper every quarter.** The "fine-tune your own to save money" pitch weakens as frontier API prices drop.

### 5.5 Staging — plausible phases if we do pursue this

Not committing; sketch only.

**Phase 0 — the site tells a story.**
Make finetuning.uk a credible industry site: clear positioning, guides,
comparisons, opinionated advice. No product yet. Works whether we ship a
product or not. Fits the existing agent-site-builder pipeline.

**Phase 1 — managed service, concierge.**
Offer fine-tuning as a bespoke service. Users talk to us, we run the job on
our infra, hand them a GGUF or host it. No self-serve UI. Validates demand
and pricing without building the product.

**Phase 2 — narrow self-serve.**
Pick one vertical (e.g., "fine-tune a customer-service model from your Zendesk
exports"). Build the minimum UI that does **just that** well. Prove
acquisition and retention economics on a narrow wedge before going broad.

**Phase 3 — general self-serve.**
Only after phase 2 has paying users. By this point we know what the UI
actually needs to be.

---

## 6. Overlap with the internal flywheel — what's reusable

| Piece | Internal flywheel | finetuning.uk product |
|---|---|---|
| Ollama GPU adapter | Serves local models | Serves tenant models (with isolation) |
| Unsloth script | Trains our LoRAs | Runs tenant training jobs |
| GPU provisioning | On-demand, manual | Automated, queued, multi-tenant |
| Training data export | `llm_call_log` → JSONL | Tenant upload → JSONL |
| Evaluation | Manual A/B | Per-tenant dashboards |
| llm_call_log schema | Our ops + training | Irrelevant to tenants (they have their own logs) |
| knowledge_base | Our RAG | Optional add-on: tenant RAG |

Most of the infra maps. What's entirely new is: **multi-tenancy, billing,
UI, support, legal.** That's where the product work lives.

---

## 7. Decisions taken (2026-04-21)

1. **Goal = both layers.** finetuning.uk is a generous, credible knowledge
   site **and** a revenue-generating service. Site links outward to credible
   third-party sources without gatekeeping. Generosity is positioning.
2. **Target user = technical-adjacent SMEs / agencies / ops leads.**
   Refined from earlier "non-technical business owners" — that audience has
   the worst economics for a solo operator (highest support load, lowest
   willingness to pay). One level up has budget and understands value.
3. **Hosting = full managed service.** We host the resulting model.
   No "here's a GGUF, good luck" tier.
4. **Infrastructure approach = separate cluster, reuse the framework.**
   Shared GPU pool with LoRA stacking, not per-tenant clusters. Training
   jobs queue onto shared training GPU(s). Tenant isolation via namespace
   + data encryption, not physical separation.
5. **Operator shape = solo, £9-12k/month target, full runway, appetite
   for immediate shipping.** Interim contract gigs acceptable at <50%
   of time to generate market signal and cash flow.
6. **Acquisition = cold only, so content-led.** No warm network to
   activate. finetuning.uk itself is the acquisition engine. The framework
   becomes the content engine that outpaces competitors who have to write
   manually.
7. **Positioning = "AI tailored to your business."** "Fine-tuning" read as
   the verb, not strictly LoRA. The site says what we actually do; the word
   redefines itself through the content.
8. **Posture on multi-agent = capability, not headline.** "18 agents" is
   marketing theatre. The product is the outcome; multi-agent is how we
   deliver it.
9. **Partner directory = strategic asset.** Positions us as honest broker.
   Protects against accepting work we can't deliver. Built from real
   relationships, not a scraped list.
10. **UI-first product build (revised).** Earlier hypothesis was "concierge
    first, UI later". Revised: build the UI as our own operational tooling
    for delivering the work. We're not building a SaaS speculatively —
    we're building our own cockpit that happens to also be a sellable
    product once it's good enough. If an SME joins early to shape it, great;
    if not, we use it ourselves on test data and first customers come to a
    working tool.
11. **Data curation is a first-class product feature, not a service
    add-on.** Most competitors treat bad data as the customer's problem.
    We use the framework to filter, deduplicate, quality-score, classify,
    redact PII, and flag inconsistencies automatically. This is the
    differentiator — most SMEs have tried and failed with RAG *because* of
    their data, not despite good data.
12. **Product-first flagship = RAG automation with built-in data
    curation.** RAG is chosen ahead of text LoRA and image LoRA because:
    (a) the user shows up with a folder of docs, not curated training
    pairs; (b) the infrastructure (knowledge_base, rag_index, rag_lookup,
    Ollama, embeddings) is all built; (c) the data-curation pipeline
    aligns naturally with RAG ingestion; (d) it's the broadest market.
    Text and image LoRA become later tiers once RAG is live and proven.

## 8. Offer structure (revised — product at the centre)

| Tier | Offer | Price | Notes |
|---|---|---|---|
| **Flagship product** | finetuning.uk RAG platform — upload docs, auto-curate, chat over them, API access | £199-1,499/mo subscription + £2-5k setup for concierge onboarding | The main product. Self-serve tier after month 3, concierge-onboarded until then. |
| Mid-range | Custom AI assistant over business knowledge — goes beyond platform defaults into bespoke integration | £5-10k setup + £800-1.5k/mo | For clients whose needs outgrow the platform defaults |
| High | Bespoke multi-agent workflow / fine-tuning / image LoRA | £15-30k project + £500-1.5k/mo | Enters after RAG platform is proven; includes text LoRA and image LoRA offerings as they come online |
| Legacy (existing) | AI-built industry sites + ongoing content ops | £3-5k setup + £1.5-3k/mo | Already demonstrated. Keep as offer but not the lead. |

Ratio at steady state: RAG platform carries ~60% of revenue once there are
5-10 subscribers plus occasional setup fees. Bespoke work tops up.

Plus entry-level productised items (diagnostic call, audit, ebook, tools)
that generate market signal without locking in direction.

### 8a. Data curation as a named feature

Data curation lives inside the RAG product as a visible feature, not a hidden
detail. The ingestion pipeline runs through framework-driven agents:

| Stage | What it does | Framework pieces reused |
|---|---|---|
| Parse | PDF/DOCX/HTML/CSV/URL → text | web-scrape adapter + doc parsers |
| Classify | Tag each chunk by topic, doc type, likely usefulness | agent with content-quality LLM call |
| Deduplicate | SHA256 + near-duplicate detection | existing dedup in rag_index |
| Quality-score | Flag low-value content (boilerplate, navigation, stale) | auditor-pattern agent |
| PII scan | Detect + optionally redact names, emails, IDs | new adapter (pattern-match + LLM review) |
| Inconsistency flag | Surface contradictions across docs | multi-agent comparison |
| Structure extract | Build outlines + topic maps | classification-extraction agent |

Each stage produces a visible report the user can review before the final
index commits. That transparency is part of the sell: they see what
they're getting, including what we chose to exclude and why.

## 9. Content pillars

Five pillars, one voice. Each earns SEO independently and funnels into
one of the offer tiers.

| Pillar | Funnel to |
|---|---|
| AI advisory (fine-tune vs RAG, vendor comparisons, honest guides) | Diagnostic calls + lead offer |
| Sites + content operations | Lead offer |
| Custom AI assistants / RAG over business data | Middle offer |
| Multi-agent workflows | High offer |
| Teaching / newsletter / courses | Products + general inbound |

Compound effect: the framework produces content, content grows the
`knowledge_base`, better knowledge produces better content. This loop is
our structural edge against competitors who produce content manually.

## 10. Shipping ladder (revised — UI-first for RAG platform)

Aspirational dates, not promises. The UI is being built as our own
operational cockpit — every feature has to justify itself by making
concierge delivery faster or better.

**Week 1**
- Reposition finetuning.uk via `site_specs` rewrite; regenerate core pages.
- Publish "Should you fine-tune?" decision guide on the repositioned site.
- "Book a diagnostic call" page live at £250/hr.
- **New:** decide auth stack (Clerk vs Supabase vs Kinde) and scaffold.
- **New:** add `tenant_id` to `knowledge_base` + enforce in `rag_lookup`/`rag_index`.

**Month 1 — MVP RAG platform, used internally only**
- Upload pipeline (PDF/DOCX/HTML/CSV/URL → `rag_index`).
- Chat UI with citations (calls `rag_lookup` → LLM).
- Source management (list, preview, remove).
- Deploy behind auth at `finetuning.uk/app`.
- Use it ourselves on test data (our own docs, public data).
- Still no paying customers yet. Framework first.

**Month 2 — Data curation pipeline becomes visible**
- Parse + classify + deduplicate + quality-score agents wired in.
- Curation report UI — user sees what was kept, dropped, flagged.
- PII detection (pattern-match first, LLM review later).
- Stripe integration + billing meter.
- First concierge customer onboarded (manually, we do the work, they use the UI).
- Free tools shipped (decision tools, cost estimators) for acquisition.

**Month 3 — Semi-self-serve**
- Integrations: Google Drive pull, Slack connector.
- Inconsistency flagging agent.
- API access for customers who want programmatic query.
- Second concierge customer.
- Case study from first customer.

**Month 4-6**
- Self-serve signup open (with generous trial).
- Text LoRA feature added as paid upgrade (for customers whose RAG plateau suggests LoRA would help).
- 3-5 paying customers.
- Product-page content engine running continuously.

**Month 6-12**
- Image LoRA feature added.
- Multi-agent workflow offering for customers outgrowing the defaults.
- Steady retainer book of 5-8 customers.
- Revenue approaching £9-12k target.

## 11. What NOT to ship immediately

Things that look like products but lock us into directions we'll regret:

- Self-serve multi-tenant fine-tuning SaaS (6+ month project; wrong audience)
- Public fine-tuning API
- Subscription "AI assistant for £99" product
- Any infrastructure without a paying customer asking for it

Things to ship immediately produce **market signal** without committing
direction: diagnostic calls, audits, tools, articles, ebooks, case studies.
Rule of thumb: if it doesn't teach us what the market wants, don't ship it
yet.

## 12. Interim gig discipline

Two rules on interim work:

1. Every interim gig should teach us something about the main thing.
   "Can I build an AI assistant for this law firm in 6 weeks?" is a gig
   that pays *and* validates the middle offer.
2. Cap interim work at 50% of available time. Rest goes to finetuning.uk.
   If interim expands to fill time, the main thing never compounds.

## 13. Still-open questions

1. Which base model(s) do we commit to when fine-tuning use cases arrive?
   Llama 3.x default; Qwen and Mistral as plausible alternatives.
2. Pricing model for the middle tier — setup + monthly, or pure monthly
   with setup amortised? Affects cash flow shape.
3. Data residency promise — "UK/EU only"? Legal promises become infra
   constraints.
4. How much of finetuning.uk itself can our agent pipeline build and
   maintain end-to-end — including content updates, new tools, case studies?
5. Which interim gig channels to target first (Upwork, Contra, direct
   agency outreach, AI-specific communities)?
6. Which free tool ships first — the one with best intent-qualification
   for our lead offer?
7. When does the FOCUS doc split — site content plan vs service delivery
   vs internal flywheel? Likely after we've shipped the first iteration.

## 14. Changelog

- 2026-04-21 (fourth pass) — UI-first pivot. Revised from "concierge first,
  UI later" to "build the UI as our own operational cockpit, use it
  ourselves, then bring customers onto it". Data curation becomes a named
  product feature (not hidden in delivery). RAG platform chosen as flagship
  ahead of text LoRA and image LoRA — those become later tiers once RAG is
  live. Offer structure and shipping ladder revised accordingly.
  Business plan spun out to separate doc.
- 2026-04-21 (third pass) — Operator shape nailed: solo, £9-12k target,
  full runway, cold acquisition only, appetite for immediate ship, tolerant
  of interim gigs. Target user refined up from "non-technical owners" to
  "technical-adjacent SMEs / agencies / ops leads". Three-tier offer
  structure locked. Five content pillars defined. Shipping ladder laid
  out through month 12. Clear list of what not to ship. Interim-gig
  discipline captured.
- 2026-04-21 (second pass) — Direction set: both layers (credible site
  + product), hosted service, separate cluster reusing the framework.
- 2026-04-21 — Initial consolidation. Pulled internal flywheel material
  from 018_canine_biology, 009_model_infrastructure, 004_improvement_loop,
  022_ai_endpoint_health.sql, 021_model_swap_and_rollback.sql,
  012b_rag_best_practices_v2. Framed the finetuning.uk product question.
