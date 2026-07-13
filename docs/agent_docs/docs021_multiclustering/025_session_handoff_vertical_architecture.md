# Session Handoff — Vertical Cluster Architecture and Domain Strategy

## Date: 2026-03-05

---

## Where We Are — The Big Picture

We have a working agent orchestration framework that builds multipage websites from domain names. The framework uses Kubernetes, Kafka, and Postgres with a hierarchical agent architecture where orchestrators spawn and call specialised child agents. The agent chassis is a single Go binary driven by SQL workflow definitions — different agent types are configuration, not code.

Over the past two weeks, the strategic direction has evolved from "build generic websites" to something more ambitious: **build websites that compete on genuine domain expertise by leveraging deep, vertical-specific knowledge bases.** The insight is that AI content sites all produce the same surface-level synthesis. What ranks — and what creates real asset value — is content containing knowledge the reader cannot easily find elsewhere.

The architecture to achieve this is now designed but not yet fully implemented. The implementation has a clear dependency chain with the RAG pipeline as the foundational piece.

---

## Documents Produced This Session

Three documents were created, building on each other:

### 1. Domain Content Strategy Framework
**File**: `domain_content_strategy_framework.md` (created in previous session, referenced throughout)
**Contains**: 15-question framework for analysing any domain's content potential. Three layers — what is this site, what content does it need, what makes each page compete. Worked examples for gaswholesalers.com and vetcomparison.uk with verified revenue data (energy broker commissions, pet insurance affiliate rates, content site valuation multiples).

### 2. Deep Research Strategy for Domain Authority
**File**: `deep_research_domain_authority.md`
**Contains**: How multi-cluster knowledge bases create content that outranks competitors. Detailed analysis of three domains:

- **vetcomparison.uk** — Four knowledge clusters needed (treatment/procedures, breed health profiles, cost structures, practice quality indicators). Leverages existing canine biology research. Revenue projection: £1,960-7,875/month at months 12-18.
- **gaswholesalers.com** — Five knowledge clusters (gas market structure/pricing, contract analysis, supplier differentiation, regulatory/compliance, consumption benchmarking). Revenue projection: £1,250-5,350/month at months 12-18.
- **mortgagecalculator.co.uk** — Four knowledge clusters (swap rate dynamics, lender affordability models, product structures, specialist situations). Long-tail keyword strategy against major competitors. Five unique calculator concepts. Revenue projection: £16,500-44,000/month at months 12-18 (mortgage is highest-value vertical).
- **xmaspresents.com** — Seasonal challenge analysed. Recommendation: extend to year-round gift content. Revenue projection: £500-1,665/month averaged annually.
- **design.co.uk** — Recommendation: sell as premium domain (£20,000-100,000+) rather than develop. Domain ambiguity reduces content site value but increases name-value sale price.

General principle extracted: the knowledge base pipeline (identify primary sources → build clusters → synthesise → find gaps → produce content filling those gaps) applies to any vertical.

### 3. Vertical Cluster Architecture
**File**: `vertical_cluster_architecture.md`
**Contains**: The full architectural design for vertical-specialised knowledge systems. This is the main strategic document. Covers:

- What a vertical contains (knowledge base collection, research strategy, content patterns, monetisation config)
- All domain strategies with specific knowledge clusters, content page lists, and revenue projections
- The routing architecture (domain intake → classification → vertical orchestrator)
- The research/build cluster separation (why research is messy and should be separated, the two-cluster model, how they interact via shared knowledge base + Kafka coordination)
- Knowledge accumulation loop (compounding advantage — each domain enriches the vertical for all future domains)
- Full hierarchy diagram showing all verticals with their research orchestrators, knowledge bases, build orchestrators
- What already exists vs what needs building
- Schema additions (source_authority on knowledge_base, vertical_registry table)
- Implementation order (10 steps)

---

## The Dependency Chain

This is the most important thing to understand for continuing work. The vertical architecture depends on foundational pieces that are designed but not deployed:

### Layer 1: RAG Pipeline (blocking everything else)

**Status**: Code written, not deployed.

Files ready to deploy (from development guide):
- `081_llm_model_upgrades_and_logging.sql` — model alias updates + llm_call_log table
- `082_rag_knowledge_base.sql` — knowledge_base table with pgvector + trigram indexes
- `platform/aiservice/ollama.go` — Ollama provider
- `platform/orchestration/actions/llm_call_logger.go` — fire-and-forget logging
- `platform/orchestration/actions/rag_actions.go` — rag_lookup and rag_index actions
- `platform/orchestration/datahelpers/nullable_helpers.go`
- Patches to `anthropic.go`, `ai_actions.go`, `registry.go`
- `ollama-adapter-k8s.yaml` — Kubernetes deployment for Ollama

**Deployment order**:
1. Run SQL migrations (081, 082) against clients database
2. Apply Go code patches + new files → rebuild agent-chassis image
3. Verify LLM call logging works
4. Deploy Ollama adapter (for embeddings)
5. Test `rag_index` and `rag_lookup` in a simple workflow

### Layer 2: First Knowledge Base Content

**Status**: Canine biology research exists conceptually (from the multi-cluster canine biology project design session) but has not been processed into the knowledge_base table.

**What needs to happen**:
- The canine biology material needs to be chunked into knowledge pieces
- Each chunk embedded via Ollama (nomic-embed-text)
- Stored in knowledge_base with `collection: "veterinary"`
- This is the first real data in the RAG system

### Layer 3: RAG-Augmented Content Generation

**Status**: Not yet wired.

**What needs to happen**:
- Add `rag_lookup` step before content generation in at least one workflow
- Confirm that content produced with RAG context is measurably better than content without
- This validates the entire vertical concept

### Layer 4: Vertical Orchestrators

**Status**: Designed, not built.

**What needs to happen** (only after layers 1-3 are working):
- Extend site classifier to output `vertical_slug` and `disposition`
- Create `vertical_registry` table
- Create vertical research orchestrator definitions (agent definitions with workflows)
- Create vertical build orchestrator definitions
- Add routing to main pipeline orchestrator
- Seed first vertical knowledge base (veterinary, using existing canine biology material)
- Test end-to-end: domain → classify → vertical orch → rag_lookup → content with depth

### Layer 5: Research/Build Cluster Separation

**Status**: Designed, implementation deferred.

**What needs to happen** (only when research workload justifies it):
- Research orchestrators move to separate cluster via `dispatch_agent`
- Build orchestrators stay on main cluster
- Shared knowledge base via Postgres, coordination via Kafka

---

## Schema Additions Designed

Two additions to the database, not yet applied:

```sql
-- Extend knowledge_base with source provenance
ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS source_authority integer DEFAULT 3;
ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS source_url text;
ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS source_date timestamp with time zone;
ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS vertical_slug text;
ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS knowledge_type text;

CREATE INDEX IF NOT EXISTS idx_kb_vertical ON knowledge_base(vertical_slug);
CREATE INDEX IF NOT EXISTS idx_kb_authority ON knowledge_base(source_authority DESC);

-- Vertical registry
CREATE TABLE IF NOT EXISTS vertical_registry (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    vertical_slug text UNIQUE NOT NULL,
    display_name text NOT NULL,
    description text,
    research_orch_type text NOT NULL,
    build_orch_type text NOT NULL,
    knowledge_collection text NOT NULL,
    research_sources jsonb DEFAULT '[]',
    content_patterns jsonb DEFAULT '[]',
    monetisation_config jsonb DEFAULT '{}',
    refresh_schedule jsonb DEFAULT '{}',
    maturity_stage text DEFAULT 'seeding',
    domain_count integer DEFAULT 0,
    knowledge_chunk_count integer DEFAULT 0,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);
```

These depend on migration 082 (knowledge_base table) being deployed first.

---

## Verticals Designed

Five verticals have been analysed with specific knowledge clusters, content strategies, and revenue projections:

| Vertical | Slug | Example Domain | Primary Revenue | Monthly Potential (12-18mo) |
|---|---|---|---|---|
| Veterinary | veterinary | vetcomparison.uk | Insurance affiliate + practice listings | £1,960-7,875 |
| Energy/Utilities | energy_wholesale | gaswholesalers.com | Qualified lead gen (£30-60/lead) | £1,250-5,350 |
| Finance/Mortgage | finance_mortgage | mortgagecalculator.co.uk | Broker leads (£50-150/lead) | £16,500-44,000 |
| Seasonal/Gifts | seasonal_gifts | xmaspresents.com | Product affiliate (3-17%) | £500-1,665 (avg) |
| Generic | generic | unclassified domains | AdSense + generic affiliate | Varies |

Plus a "premium domain" pathway for domains like design.co.uk that are more valuable as name sales than content sites.

Each vertical has detailed knowledge cluster specifications, authoritative source lists, specific content pages to create, and monetisation configurations. These are fully documented in the vertical cluster architecture document.

---

## Market Data Verified Through Research

Revenue figures used in projections are grounded in verified data, not estimates:

- **Content site valuation multiples 2025**: 24-32x monthly profit (Empire Flippers averaging 24x, premium 30-35x)
- **Energy broker commissions**: 0.05p-2p per kWh (typical 0.5p = £1,000/year on 200,000 kWh business)
- **Energy lead values UK**: £10-25 raw, £30-60 qualified
- **Pet insurance affiliate rates UK**: £15-35 per signup (Petplan £35, PDSA £15)
- **Mortgage lead values UK**: £50-150 per lead (Sendlead industry benchmark data)
- **UK web design market**: ~£640 million (IBISWorld)
- **Gift affiliate commissions**: Amazon 3-4%, John Lewis 5-7%, experience companies 10-17%
- **Seasonal display RPMs**: £15-30 in Nov/Dec for gift content

---

## Key Architectural Decisions Made

1. **Verticals are logical, not physical** (initially). They run on shared infrastructure using existing mechanisms (knowledge_base collections, agent definitions with vertical-specific workflows, standard spawn/call). Physical separation via dispatch_agent when workload justifies it.

2. **Research and build should separate.** Research is messy, slow, shared across domains, and needs external network access. Build is structured, fast, per-site, and mostly internal. They communicate through the shared knowledge base (Postgres reads/writes) and Kafka for orchestration coordination.

3. **Knowledge accumulates across domains.** The first domain in a vertical bears the research cost. Subsequent domains benefit from existing knowledge. This is the compounding advantage and the defensible moat.

4. **No new Go code needed for the vertical concept.** It's all achievable through agent definition SQL, workflow configuration, and prompt engineering. The RAG pipeline (which does need Go code) is already written.

5. **The site classifier needs extending** to output vertical_slug and disposition (develop/sell/hold), but this is a prompt change to an existing agent, not a new agent.

6. **"Don't develop, sell the domain" is a valid pipeline output.** Premium generic domains like design.co.uk should route to a broker listing pathway, not the content generation pipeline.

---

## What Exists in the Codebase (Relevant to This Work)

Already built and deployed:
- Agent chassis with spawn/call orchestration
- Multi-cluster dispatch via dispatch_agent and remote-job-spawner
- Site classifier agent
- Content writer agents
- Site work orchestrator with dispatch loop
- Web scrape adapter
- Git adapter for deployment
- vertical_slug concept in vet collection pipeline

Written but not deployed:
- RAG pipeline (migrations 081/082, rag_actions.go, ollama adapter)
- LLM call logging
- Ollama provider (for embeddings and local model inference)

Not yet written:
- Vertical research orchestrator agent definitions
- Vertical build orchestrator agent definitions
- Vertical registry table
- Knowledge seeding workflows
- Knowledge refresh scheduler
- Source authority extensions to knowledge_base schema

---

## Immediate Next Steps (In Order)

1. **Deploy RAG pipeline** — this unblocks everything else
2. **Test rag_index and rag_lookup** with simple data
3. **Index canine biology material** into knowledge_base (collection: "veterinary")
4. **Wire rag_lookup into a content writer workflow** and validate quality improvement
5. **Extend site classifier** to output vertical_slug
6. **Create vertical_registry table** and seed initial verticals
7. **Build first vertical research orchestrator** (veterinary — has existing knowledge)
8. **Build first vertical build orchestrator** (veterinary)
9. **Test end-to-end** with vetcomparison.uk
10. **Expand to energy and mortgage verticals**

---

## Previous Session Context

This session builds on work from multiple previous sessions. Key transcripts:

| Session | Content |
|---|---|
| 2026-02-27 | Million-agent scaling plan, multi-cluster architecture |
| 2026-02-28 | Multi-cluster dispatch design |
| 2026-03-02 (×4) | Dispatch implementation, remote-job-spawner, cluster C AWS, handoff |
| 2026-03-03 (×4) | Kafka scaling, canine biology project design, commercialisation strategy, pipeline quality assessment |
| 2026-03-03 | Domain content strategy framework (15-question methodology, worked examples) |
| 2026-03-05 (this) | Domain analysis (mortgage, xmas, design), deep research strategy, vertical cluster architecture, research/build separation |

The commercialisation strategy session established the two-tier revenue model: service tier (build sites for clients) leading to setup sales (sell the pipeline capability). The £25K+ annual revenue target. The domain portfolio as the testing ground.

The canine biology session designed a multi-cluster knowledge base project that became the proof-of-concept for the vertical knowledge approach. That design exists but the knowledge hasn't been processed into the system yet.

---

## Standing Rules for This Project

- Keep workflows simple, put complexity in Go action code
- Keep workflow variable names in sync with what actions expect
- Spawn sub-agents rather than creating sub-workflows in SQL
- Every agent is an orchestrator
- Reuse existing functions before creating new ones
- Check database schemas before creating SQL
- Don't change variable names from current code without explicit note
- K8s namespace: `kubectl -n ai-persona-system` (or `-n kafka` for Kafka operations)
- Deployment is to GitHub, GitHub Actions triggers write to Backblaze S3
- Don't create summary documents unless asked
- Keep responses pragmatic
- Don't say "final fix" or "excellent choice"
