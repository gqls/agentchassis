# Session Handoff — Vertical Knowledge Architecture

## Date: 2026-03-21

---

## Where We Are — The Big Picture

We have a working agent orchestration framework that builds multipage websites from domain names. The framework uses Kubernetes, Kafka, and Postgres with a hierarchical agent architecture where orchestrators spawn and call specialised child agents. The agent chassis is a single Go binary driven by SQL workflow definitions.

The strategic direction is to build websites that compete on genuine domain expertise by leveraging deep, vertical-specific knowledge bases. Rather than producing generic AI content, each domain is routed to a specialised vertical (veterinary, energy, mortgage, etc.) that maintains its own deep knowledge, research strategy, and content patterns. Knowledge compounds across domains within a vertical — the tenth site benefits from everything the first nine taught the system.

The architecture for this is fully designed. The RAG pipeline that underpins it has been patched and is ready for deployment. We are at the boundary between design and implementation.

---

## Current Implementation Status

### DONE — Patched and Ready to Deploy

**Go files patched (ready to commit):**

- **`platform/aiservice/anthropic.go`** — PATCHED. Added `Usage` struct to response parsing, writes `__usage_input_tokens` and `__usage_output_tokens` back into the options map after each API call. This enables the LLM call logger to capture token usage.

- **`platform/orchestration/actions/ai_actions.go`** — PATCHED. Four changes:
  1. Model alias resolution now preserves both the original alias and resolved name as function-scoped variables for logging
  2. `llmCallStart := time.Now()` added before GenerateText call
  3. `LogLLMCall()` added in the error path (before retry logic)
  4. `LogLLMCall()` added in the success path (after response received, reads token counts from options)
  5. `case "ollama":` added to `createAIClient` switch

**New Go files (ready to add):**

- **`platform/aiservice/ollama.go`** — Ollama provider implementing AIService interface. GenerateText (via /api/chat) and GenerateEmbedding (via /api/embeddings). Writes usage tokens back to options map.
- **`platform/orchestration/actions/llm_call_logger.go`** — Fire-and-forget LLM call logging. Runs in a goroutine with 5s timeout. Never blocks the workflow.
- **`platform/orchestration/actions/rag_actions.go`** — `RAGLookupAction` (vector search with trigram fallback) and `RAGIndexAction` (chunk, embed, store with SHA256 dedup). Both use Ollama for embeddings.
- **`platform/orchestration/datahelpers/nullable_helpers.go`** — NullableInt, NullableInt64 for SQL parameter building.

**Still needs patching (not yet done):**
- **`platform/orchestration/actions/registry.go`** — Add `rag_lookup` and `rag_index` entries to GlobalActionRegistry. Two lines to add in the storage section.

**SQL migrations (ready to run):**

- **`081_llm_model_upgrades_and_logging_v2.sql`** — Idempotent. Upgrades model references:
  - `chief-strategist` → `claude-opus-4-6`
  - `site-planner`, `build-site-planner` → `claude-sonnet-4-6`
  - `domain-research-classifier` → `claude-sonnet-4-6`
  - `domain-strategist` → `claude-sonnet-4-6`
  - `site-classifier` → `claude-sonnet-4-6` (was haiku — upgraded because vertical classification is more complex)
  - All stale claude-3.x refs → `claude-sonnet-4-6`
  - Creates `llm_call_log` table with indexes, cleanup function, stats view

- **`082_rag_knowledge_base_fixed.sql`** — Idempotent (fixed from partial run that failed on index creation). Drops existing indexes before recreating. Creates `knowledge_base` table with pgvector (768-dim for nomic-embed-text), trigram fallback index, dedup on collection+content_hash, stats view. The table already exists from the partial run; the fixed migration handles this via CREATE TABLE IF NOT EXISTS plus DROP INDEX IF EXISTS.

**Kustomize manifests (ready to apply):**

- **Ollama adapter** — Full kustomize structure following the same pattern as web-scrape-adapter, git-adapter, image-generator-adapter:
  - `deployments/kustomize/services/ollama-adapter/base/` — deployment.yaml (with init container that pulls nomic-embed-text), service.yaml (ClusterIP on 11434), pvc.yaml (10Gi for models), kustomization.yaml
  - `deployments/kustomize/services/ollama-adapter/overlays/production/uk_001/kustomization.yaml`
  - Init container starts Ollama in background, pulls models, then exits. Idempotent on restart.
  - Makefile target `deploy-ollama-adapter` needs adding, plus entries in `deploy-agents` and `redeploy-agents`

**Model aliases** — `model_aliases.go` already has entries for `claude-sonnet-4-6` and `claude-opus-4-6` mapping correctly. No changes needed.

### NOT YET DONE — Next Deployment Steps

1. **Patch registry.go** — Add rag_lookup and rag_index to GlobalActionRegistry (2 lines)
2. **Commit all Go changes** — 4 new files + 2 patched files + 1 registry patch
3. **Build and push new chassis image** — `docker build/push docker.io/aqls/agent-chassis:v1.0.XXX`
4. **Run SQL migrations** — 081 then 082 against clients database
5. **Update image tag** — `make update-kustomization-images IMAGE_TAG=v1.0.XXX && make deploy-agents`
6. **Add Makefile targets** for ollama-adapter, copy kustomize directory
7. **Deploy Ollama adapter** — `make deploy-ollama-adapter`
8. **Verify** — check llm_call_log populates on next site build, check Ollama responds to embedding requests

---

## Documents Produced

### Strategic Documents

| Document | File | Content |
|---|---|---|
| Domain Content Strategy Framework | `domain_content_strategy_framework.md` | 15-question framework for analysing any domain's content potential. Three layers of questions. Worked examples for gaswholesalers.com and vetcomparison.uk with verified revenue data. |
| Deep Research Strategy | `deep_research_domain_authority.md` | How multi-cluster knowledge bases create content that outranks competitors. Detailed domain analysis: vetcomparison.uk (4 knowledge clusters), gaswholesalers.com (5 clusters), mortgagecalculator.co.uk (4 clusters + 5 calculator concepts), xmaspresents.com (seasonal strategy), design.co.uk (sell don't develop). Revenue projections per domain. |
| Vertical Cluster Architecture | `vertical_cluster_architecture.md` | Full architectural design for vertical-specialised knowledge systems. Routing, vertical registry, research/build cluster separation, knowledge accumulation loop, full hierarchy diagram. Schema additions. Implementation order. |
| Implementation Todo List | `implementation_todo_vertical_architecture.md` | 80-item phased todo across 8 phases. Covers RAG pipeline deployment, knowledge base content, schema extensions, classifier/planner prompt updates, research handler, vertical orchestrators, maintenance, fine-tuning. Includes kustomize structure for Ollama adapter with init container. |

### Deployment Artefacts

| File | Content |
|---|---|
| `rag_deployment_bundle.tar.gz` | Complete bundle: SQL migrations, Go files, patches, kustomize manifests, README |
| `rag_deployment_README.md` | File placement map, deployment steps, verification commands, Makefile additions |
| `anthropic_patched.go` | Patched anthropic.go with Usage capture — drop-in replacement |
| `ai_actions_patched.go` | Patched ai_actions.go with LLM logging + ollama case — drop-in replacement |
| `081_llm_model_upgrades_and_logging_v2.sql` | Model upgrades to 4.6 + llm_call_log table (idempotent) |
| `082_rag_knowledge_base_fixed.sql` | Knowledge_base table + indexes (idempotent, fixes partial run failure) |

---

## The Dependency Chain

```
DONE (this session)                          NEXT (deploy)
─────────────────                            ──────────────
✓ anthropic.go patched                       → Patch registry.go (2 lines)
✓ ai_actions.go patched                      → Commit all Go changes
✓ ollama.go written                          → Build + push chassis image
✓ llm_call_logger.go written                 → Run SQL migrations (081, 082)
✓ rag_actions.go written                     → Deploy new image to cluster
✓ nullable_helpers.go written                → Add Makefile targets for ollama
✓ 081 migration updated (4.6 models)         → Copy kustomize directory for ollama
✓ 082 migration fixed (idempotent)           → Deploy Ollama adapter
✓ Ollama kustomize structure designed        → Verify LLM logging works
✓ Model aliases already have 4.6 entries     → Verify Ollama serves embeddings
                                             → Test rag_index + rag_lookup end-to-end

AFTER DEPLOY VERIFIED                        THEN (Phase 1-3)
────────────────────                         ─────────────────
                                             → Index canine biology into knowledge_base
                                             → Wire rag_lookup into content writer workflow
                                             → A/B test content quality with/without RAG
                                             → Add vertical_registry table + schema extensions
                                             → Update classifier prompts (vertical_slug output)
                                             → Update planner prompts (vertical page types)
                                             → Create vertical-specific planner variants
                                             → Build research handler agent
```

---

## Architecture Decisions Made

1. **Verticals are logical, not physical** (initially). They run on shared infrastructure using knowledge_base collections, agent definitions with vertical-specific workflows, and standard spawn/call. Physical separation via dispatch_agent when workload justifies it.

2. **Research and build should separate.** Research is messy, slow, shared across domains. Build is structured, fast, per-site. They communicate through the shared knowledge base (Postgres reads/writes) and Kafka for orchestration.

3. **Knowledge accumulates across domains.** First domain in a vertical bears the research cost. Subsequent domains benefit. This is the compounding advantage and the defensible moat.

4. **Research fits the work items model.** `needs_vertical_research` items run at priority 1-4, content items depend on them at priority 10-17. The dispatch loop doesn't change — research is just another handler agent type.

5. **Vertical intelligence lives in the planner and specs, not the handlers.** Content writers receive RAG context and content patterns via the work item spec. The handler code is the same across verticals. What changes is what the planner creates and what knowledge the spec tells the handler to retrieve.

6. **Ollama deploys as a standard adapter** — kustomize base + overlay, Makefile targets, init container for model pulling. Follows the same patterns as web-scrape-adapter, git-adapter, etc.

7. **Models upgraded to 4.6 family.** Chief-strategist gets opus-4-6 (most important structural decisions). Classifiers and planners get sonnet-4-6. Content generation stays on haiku-4-5 (short constrained outputs where cost matters more than reasoning depth).

8. **"Don't develop, sell the domain" is a valid pipeline output.** Premium generic domains like design.co.uk route to a broker listing pathway, not content generation.

---

## Verticals Designed

| Vertical | Slug | Example Domain | Knowledge Clusters | Revenue Model | Monthly Potential (12-18mo) |
|---|---|---|---|---|---|
| Veterinary | veterinary | vetcomparison.uk | Treatment/procedures, breed health, cost structures, practice quality | Insurance affiliate (£15-35/signup) + practice listings | £1,960-7,875 |
| Energy | energy_wholesale | gaswholesalers.com | Market structure, contract analysis, supplier data, regulatory, benchmarking | Qualified leads (£30-60/lead) | £1,250-5,350 |
| Finance/Mortgage | finance_mortgage | mortgagecalculator.co.uk | Swap rates, lender criteria, product structures, specialist situations | Broker leads (£50-150/lead) | £16,500-44,000 |
| Seasonal/Gifts | seasonal_gifts | xmaspresents.com | Gift category intelligence, affiliate programme data | Product affiliate (3-17%) | £500-1,665 (avg) |
| Generic | generic | unclassified | Competitor analysis, keyword research | AdSense + generic affiliate | Varies |

---

## Key Files in the Codebase

| What | Where | Status |
|---|---|---|
| Model aliases | `platform/aiservice/model_aliases.go` | Has 4.6 entries, no changes needed |
| Anthropic client | `platform/aiservice/anthropic.go` | PATCHED (usage capture) — ready to commit |
| AI actions | `platform/orchestration/actions/ai_actions.go` | PATCHED (logging + ollama) — ready to commit |
| Action registry | `platform/orchestration/actions/registry.go` | NEEDS PATCH (add 2 rag entries) |
| Ollama client | `platform/aiservice/ollama.go` | NEW FILE ready |
| LLM logger | `platform/orchestration/actions/llm_call_logger.go` | NEW FILE ready |
| RAG actions | `platform/orchestration/actions/rag_actions.go` | NEW FILE ready |
| Nullable helpers | `platform/orchestration/datahelpers/nullable_helpers.go` | NEW FILE ready |
| Ollama kustomize | `deployments/kustomize/services/ollama-adapter/` | NEW DIRECTORY ready |

---

## Previous Session Context

| Session | Content |
|---|---|
| 2026-02-27 | Million-agent scaling plan, multi-cluster architecture |
| 2026-02-28 | Multi-cluster dispatch design |
| 2026-03-02 (×4) | Dispatch implementation, remote-job-spawner, cluster C AWS, handoff |
| 2026-03-03 (×4) | Kafka scaling, canine biology project, commercialisation strategy, pipeline quality |
| 2026-03-03 | Domain content strategy framework (15-question methodology) |
| 2026-03-05 | Domain analysis, deep research strategy, vertical cluster architecture |
| 2026-03-21 (this) | Work items integration, RAG deployment files, Go patches applied, migrations fixed, Ollama adapter kustomize, model upgrades to 4.6 |

---

## Standing Rules

- Keep workflows simple, put complexity in Go action code
- Keep workflow variable names in sync with what actions expect
- Spawn sub-agents rather than creating sub-workflows in SQL
- Every agent is an orchestrator
- Reuse existing functions before creating new ones
- Check database schemas before writing SQL
- Don't change variable names without explicit note
- K8s namespace: `kubectl -n ai-persona-system` (or `-n kafka`)
- Deployment: GitHub → GitHub Actions → Backblaze S3
- Don't use logger.Debug (won't show in logs)
- Don't create summary documents unless asked
- Don't say "final fix" or "excellent choice"
