# Implementation Todo List — Vertical Knowledge Architecture

## Overview

This is the ordered implementation plan from current state to vertical-specialised domain authority sites. Each phase builds on the previous. Items within a phase can often be parallelised. Dependencies are noted where they exist.

---

## Phase 0: RAG Pipeline Foundation

Everything downstream depends on this. No vertical architecture works without the knowledge base and retrieval mechanism.

### 0.1 Database Migrations

- [ ] **Run migration 081_llm_model_upgrades_and_logging.sql** against clients database. Creates `llm_call_log` table, updates model aliases, adds cleanup function and stats view.
  ```bash
  kubectl -n ai-persona-system exec -it deploy/postgres-clients -- \
      psql -U clients_user -d clients_db
  ```
- [ ] **Run migration 082_rag_knowledge_base.sql** against clients database. Creates `knowledge_base` table with pgvector extension, trigram indexes, and stats view.
- [ ] **Verify model aliases updated** — check that agent definitions reference current model strings:
  ```sql
  SELECT type, default_config->'ai_service'->>'model' as model
  FROM agent_definitions
  WHERE type IN ('chief-strategist', 'site-planner', 'domain-research-classifier')
  ORDER BY type;
  ```

### 0.2 Go Code Changes

- [ ] **Add new files to codebase:**
  - `platform/aiservice/ollama.go` — Ollama provider implementing AIService interface
  - `platform/orchestration/actions/llm_call_logger.go` — fire-and-forget LLM call logging
  - `platform/orchestration/actions/rag_actions.go` — `rag_lookup` and `rag_index` actions
  - `platform/orchestration/datahelpers/nullable_helpers.go` — NullableInt, NullableInt64

- [ ] **Apply PATCH_01 to `platform/aiservice/anthropic.go`** — Usage struct in response, write-back to options map for token tracking.

- [ ] **Apply PATCH_02 to `platform/orchestration/actions/ai_actions.go`** — 4 insertions for LLM call logging after each `execute_llm_prompt` call, plus `ollama` case in `createAIClient` switch.

- [ ] **Apply PATCH_03 to `platform/orchestration/actions/registry.go`** — Register `rag_lookup` and `rag_index` in GlobalActionRegistry.

- [ ] **Rebuild and push agent-chassis image:**
  ```bash
  docker build -t docker.io/aqls/agent-chassis:v1.0.XXX .
  docker push docker.io/aqls/agent-chassis:v1.0.XXX
  ```

- [ ] **Update image tag in agent_definitions** — bump all definitions to the new image version.

### 0.3 Deploy Ollama Adapter

Ollama is a permanent adapter (like web-scrape-adapter, git-adapter, image-generator-adapter). It must follow the same kustomize/Makefile patterns as the other adapters. It differs in that it uses a third-party image (`ollama/ollama`) rather than being built from the repo, and it needs a PVC for model persistence plus an init container to pull models on startup.

#### 0.3.1 Create Kustomize Structure

- [ ] **Create base manifests** at `deployments/kustomize/services/ollama-adapter/base/`:

  - [ ] `deployment.yaml` — Deployment using `ollama/ollama` image. Key differences from other adapters:
    - **Init container** that pulls required models before the main container starts. The init container uses the same `ollama/ollama` image, mounts the same PVC, starts the Ollama server in the background, runs `ollama pull nomic-embed-text`, then exits. This is idempotent — if the model is already on the PVC from a previous run, the pull completes instantly.
    - **PVC volume mount** at `/root/.ollama` for model persistence across pod restarts.
    - **Resource limits** appropriate for CPU inference: requests `cpu: 500m, memory: 2Gi`, limits `cpu: 4, memory: 8Gi`. Memory is higher than other adapters because models load into RAM (~1GB for nomic-embed-text, ~4GB for a quantized 7B if used later).
    - **No liveness/readiness probes using HTTP** initially — Ollama's `/api/tags` endpoint can serve as readiness check once the server is up, but the model loading time means the initial delay needs to be generous (60-90s).
    - **Single replica** — CPU inference is sequential within a model, so horizontal scaling doesn't help much. Add a second replica only if request queuing becomes a bottleneck.
    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: ollama-adapter
      labels:
        app: ollama-adapter
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: ollama-adapter
      template:
        metadata:
          labels:
            app: ollama-adapter
        spec:
          initContainers:
            - name: model-pull
              image: ollama/ollama:latest
              command: ["/bin/sh", "-c"]
              args:
                - |
                  # Start Ollama server in background
                  ollama serve &
                  SERVER_PID=$!
                  # Wait for server to be ready
                  until curl -s http://localhost:11434/api/tags > /dev/null 2>&1; do
                    sleep 1
                  done
                  # Pull required models (idempotent — skips if already present)
                  ollama pull nomic-embed-text
                  # Stop server
                  kill $SERVER_PID
                  wait $SERVER_PID 2>/dev/null
                  echo "Models ready"
              volumeMounts:
                - name: ollama-models
                  mountPath: /root/.ollama
              resources:
                requests:
                  cpu: "500m"
                  memory: "2Gi"
                limits:
                  cpu: "2"
                  memory: "4Gi"
          containers:
            - name: ollama
              image: ollama/ollama:latest
              ports:
                - containerPort: 11434
                  name: http
              volumeMounts:
                - name: ollama-models
                  mountPath: /root/.ollama
              resources:
                requests:
                  cpu: "500m"
                  memory: "2Gi"
                limits:
                  cpu: "4"
                  memory: "8Gi"
              readinessProbe:
                httpGet:
                  path: /api/tags
                  port: 11434
                initialDelaySeconds: 10
                periodSeconds: 10
              livenessProbe:
                httpGet:
                  path: /api/tags
                  port: 11434
                initialDelaySeconds: 30
                periodSeconds: 30
          volumes:
            - name: ollama-models
              persistentVolumeClaim:
                claimName: ollama-models-pvc
    ```

  - [ ] `service.yaml` — ClusterIP service exposing port 11434. Other agents reference it as `http://ollama-adapter.ai-persona-system.svc.cluster.local:11434`.
    ```yaml
    apiVersion: v1
    kind: Service
    metadata:
      name: ollama-adapter
      labels:
        app: ollama-adapter
    spec:
      type: ClusterIP
      selector:
        app: ollama-adapter
      ports:
        - port: 11434
          targetPort: 11434
          name: http
    ```

  - [ ] `pvc.yaml` — PersistentVolumeClaim for model storage. Models persist across pod restarts. Size depends on what models you'll load: nomic-embed-text is ~300MB, a quantized 7B is ~4GB. Start with 10Gi to leave room.
    ```yaml
    apiVersion: v1
    kind: PersistentVolumeClaim
    metadata:
      name: ollama-models-pvc
      labels:
        app: ollama-adapter
    spec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: 10Gi
    ```

  - [ ] `kustomization.yaml` — base kustomization:
    ```yaml
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    resources:
      - deployment.yaml
      - service.yaml
      - pvc.yaml
    ```

- [ ] **Create production overlay** at `deployments/kustomize/services/ollama-adapter/overlays/production/uk_001/`:

  - [ ] `kustomization.yaml`:
    ```yaml
    apiVersion: kustomize.config.k8s.io/v1beta1
    kind: Kustomization
    namespace: ai-persona-system
    resources:
      - ../../../base
    images:
      - name: ollama/ollama
        newTag: latest
    ```
    Note: unlike other adapters where `newTag` is the agent-chassis build version, here it's the Ollama release tag. You may want to pin to a specific version (e.g., `0.5.4`) rather than `latest` for reproducibility.

#### 0.3.2 Update Makefile

- [ ] **Add `deploy-ollama-adapter` target** following the same pattern as other adapters:
  ```makefile
  .PHONY: deploy-ollama-adapter
  deploy-ollama-adapter: ## Deploy ollama-adapter using kustomize
  	@echo "$(YELLOW)Deploying ollama-adapter...$(NC)"
  	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k \
  	    $(KUSTOMIZE_DIR)/services/ollama-adapter/overlays/$(OVERLAY_PATH)
  ```
  Note: no `sed` for image tag update here since Ollama uses its own image, not `docker.io/aqls/ollama-adapter`. If you want to pin versions, update the overlay's `kustomization.yaml` directly.

- [ ] **Add ollama-adapter to `deploy-agents` target** — add after the image-generator-adapter block:
  ```makefile
  # Deploy ollama-adapter
  @echo "Deploying ollama-adapter..."
  @if [ -d "$(KUSTOMIZE_DIR)/services/ollama-adapter/overlays/$(OVERLAY_PATH)" ]; then \
      KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k \
          $(KUSTOMIZE_DIR)/services/ollama-adapter/overlays/$(OVERLAY_PATH); \
  fi
  ```

- [ ] **Add ollama-adapter to `redeploy-agents` target**:
  ```makefile
  kubectl rollout restart deployment ollama-adapter -n ai-persona-system
  ```

- [ ] **Do NOT add ollama-adapter to `update-kustomization-images`** — that target updates `docker.io/aqls/*` image tags. Ollama uses `ollama/ollama` which has its own versioning.

#### 0.3.3 Deploy and Verify

- [ ] **Deploy using the new Makefile target:**
  ```bash
  make deploy-ollama-adapter ENVIRONMENT=production REGION=uk001
  ```

- [ ] **Wait for init container to complete** (first deploy pulls the model, may take a few minutes depending on network speed):
  ```bash
  kubectl -n ai-persona-system get pods -l app=ollama-adapter -w
  ```
  You should see the pod go through `Init:0/1` → `Init:0/1` (pulling) → `PodInitializing` → `Running`.

- [ ] **Verify model is loaded:**
  ```bash
  kubectl -n ai-persona-system exec -it deploy/ollama-adapter -- \
      ollama list
  ```
  Should show `nomic-embed-text` with its size and modification date.

- [ ] **Verify service is accessible from other pods:**
  ```bash
  kubectl -n ai-persona-system run --rm -it test-ollama --image=curlimages/curl -- \
      curl -s http://ollama-adapter.ai-persona-system.svc.cluster.local:11434/api/tags
  ```

- [ ] **Test an embedding call:**
  ```bash
  kubectl -n ai-persona-system exec -it deploy/ollama-adapter -- \
      curl -s http://localhost:11434/api/embeddings -d '{
        "model": "nomic-embed-text",
        "prompt": "French Bulldog brachycephalic health"
      }' | head -c 200
  ```
  Should return a JSON response with an `embedding` array of 768 floats.

#### 0.3.4 Adding Models Later

When you need additional models (e.g., a fine-tuned 7B for classification), update the init container's `args` to pull them:
```yaml
args:
  - |
    ollama serve &
    SERVER_PID=$!
    until curl -s http://localhost:11434/api/tags > /dev/null 2>&1; do sleep 1; done
    ollama pull nomic-embed-text
    ollama pull site-classifier-v1    # <-- add new models here
    kill $SERVER_PID
    wait $SERVER_PID 2>/dev/null
```
For custom fine-tuned models that aren't in the Ollama registry, you'll need to copy the GGUF file to the PVC and create the model via `ollama create`. That's a Phase 7 concern — for now, only `nomic-embed-text` is needed.

### 0.4 Verify LLM Logging

- [ ] **Run a site build** (any domain) with the new image.
- [ ] **Check logging is working:**
  ```sql
  SELECT agent_type, step_name, model,
         input_tokens, output_tokens, latency_ms,
         LEFT(response_text, 80) as preview
  FROM llm_call_log ORDER BY created_at DESC LIMIT 20;
  ```
- [ ] **Check stats view:**
  ```sql
  SELECT * FROM llm_call_stats;
  ```

### 0.5 Test RAG Actions

- [ ] **Test rag_index manually** — insert some test knowledge into the knowledge_base:
  ```sql
  -- Insert a test chunk directly to verify the table works
  INSERT INTO knowledge_base (collection, content, metadata)
  VALUES ('test', 'French Bulldogs are brachycephalic breeds prone to airway obstruction.',
          '{"source": "test", "topic": "breed_health"}');
  ```

- [ ] **Test rag_lookup via a simple workflow** — create a minimal test agent definition that does a `rag_lookup` step and returns the results. Confirm it retrieves the test chunk.

- [ ] **Test rag_index via workflow** — create a test workflow step that indexes content through the action (with Ollama embedding). Verify the embedding column is populated.

- [ ] **Test embedding search** — confirm vector similarity search returns relevant results when querying with related terms (e.g., "brachycephalic dog health" should find the French Bulldog chunk).

---

## Phase 1: First Knowledge Base Content

The RAG pipeline is plumbing. This phase puts real data through it.

### 1.1 Index Canine Biology Knowledge

- [ ] **Prepare canine biology material for ingestion.** The multi-cluster canine biology project produced research output covering anatomy, physiology, genetics, nutrition, and behaviour. This needs to be structured into chunks suitable for indexing.

- [ ] **Create a knowledge seeding script or workflow** that processes the canine biology material through `rag_index` with `collection: "veterinary"`. Each chunk needs meaningful metadata (topic, sub-topic, source type).

- [ ] **Verify knowledge base population:**
  ```sql
  SELECT collection, COUNT(*) as chunks,
         COUNT(embedding) as embedded,
         COUNT(*) - COUNT(embedding) as missing_embeddings
  FROM knowledge_base
  WHERE collection = 'veterinary'
  GROUP BY collection;
  ```

- [ ] **Test retrieval quality** — run `rag_lookup` queries against the veterinary collection and assess whether returned chunks are relevant and useful:
  - Query: "french bulldog health problems" — should return brachycephalic content
  - Query: "labrador hip dysplasia" — should return relevant orthopaedic content
  - Query: "dog vaccination schedule" — should return immunology content

### 1.2 Wire RAG into Content Generation

- [ ] **Add rag_lookup step to page-content-writer workflow.** Insert before the LLM content generation step. The lookup should use fields from the work item spec:
  ```json
  "lookup_knowledge": {
      "action": "rag_lookup",
      "config": {
          "query_field": "current_page.rag_query",
          "collection_field": "current_page.rag_collection",
          "top_k": 5,
          "embedding_service": {
              "provider": "ollama",
              "model": "nomic-embed-text",
              "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
          }
      },
      "next_step": "generate_content",
      "output_field": "knowledge_context"
  }
  ```
  When `rag_collection` or `rag_query` are absent from the spec, the action should return empty context (no failure). This keeps backwards compatibility.

- [ ] **Update page-content-writer prompt template** to inject RAG context when available:
  ```
  {{if .knowledge_context.rag_context}}
  ## Domain Knowledge
  Use the following verified knowledge to inform your content. 
  This is authoritative information — prefer it over general knowledge.
  
  {{.knowledge_context.rag_context}}
  {{end}}
  ```

- [ ] **A/B test content quality** — build two versions of the same page (one with RAG context, one without) and compare depth, specificity, and accuracy. This validates the entire vertical concept.

---

## Phase 2: Knowledge Base Schema Extensions

Before building verticals, extend the knowledge base to support source provenance and vertical tagging.

### 2.1 Schema Additions

- [ ] **Add source provenance columns to knowledge_base:**
  ```sql
  ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS source_authority integer DEFAULT 3;
  ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS source_url text;
  ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS source_date timestamptz;
  ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS vertical_slug text;
  ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS knowledge_type text;
  
  COMMENT ON COLUMN knowledge_base.source_authority IS '1=blog/forum, 2=trade publication, 3=industry body, 4=regulatory/clinical guideline, 5=primary research/statute';
  COMMENT ON COLUMN knowledge_base.knowledge_type IS 'factual, procedural, pricing, regulatory, comparative';
  
  CREATE INDEX IF NOT EXISTS idx_kb_vertical ON knowledge_base(vertical_slug);
  CREATE INDEX IF NOT EXISTS idx_kb_authority ON knowledge_base(source_authority DESC);
  ```

- [ ] **Create vertical_registry table:**
  ```sql
  CREATE TABLE IF NOT EXISTS vertical_registry (
      id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
      vertical_slug text UNIQUE NOT NULL,
      display_name text NOT NULL,
      description text,
      research_orch_type text,
      build_orch_type text,
      knowledge_collection text NOT NULL,
      research_sources jsonb DEFAULT '[]',
      content_patterns jsonb DEFAULT '[]',
      page_type_library jsonb DEFAULT '[]',
      monetisation_config jsonb DEFAULT '{}',
      refresh_schedule jsonb DEFAULT '{}',
      maturity_stage text DEFAULT 'seeding',
      domain_count integer DEFAULT 0,
      knowledge_chunk_count integer DEFAULT 0,
      created_at timestamptz DEFAULT now(),
      updated_at timestamptz DEFAULT now()
  );
  ```

- [ ] **Update rag_lookup action** to support optional `min_authority` filter so build agents can request only high-confidence knowledge.

- [ ] **Update rag_index action** to accept `source_authority`, `source_url`, `source_date`, `vertical_slug`, `knowledge_type` in its config and write them to the new columns.

### 2.2 Seed Vertical Registry

- [ ] **Insert initial vertical definitions:**
  ```sql
  INSERT INTO vertical_registry (vertical_slug, display_name, description, knowledge_collection,
      research_sources, content_patterns, page_type_library, monetisation_config)
  VALUES
  ('veterinary', 'Veterinary & Pet Health', 'Breed health, vet comparison, pet insurance',
   'veterinary',
   '["bsava_guidelines", "rcvs_practice_standards", "kennel_club_surveys", "breed_club_data", "insurance_claims_aggregated"]',
   '["breed_health_profile", "procedure_guide", "practice_directory", "insurance_comparison", "cost_transparency", "vet_chooser"]',
   '[{"type": "breed_health_profile", "description": "Year-by-year health guide for a specific breed"},
     {"type": "procedure_guide", "description": "What a procedure involves, costs, recovery, vet selection criteria"},
     {"type": "cost_guide", "description": "Pricing transparency for common procedures with range explanations"},
     {"type": "practice_comparison", "description": "How to compare vets on quality indicators beyond reviews"},
     {"type": "insurance_comparison", "description": "Pet insurance comparison with affiliate integration"}]',
   '{"primary": "insurance_affiliate", "secondary": ["practice_listings", "specialist_lead_gen"], "insurance_payout_gbp": "15-35"}'),
   
  ('energy_wholesale', 'Energy & Utilities', 'Gas/electricity wholesale, business energy comparison',
   'energy_wholesale',
   '["ofgem_reports", "national_grid_data", "ice_futures", "cornwall_insight", "beis_statistics", "cibse_benchmarks"]',
   '["market_analysis", "contract_guide", "supplier_directory", "benchmarking_tool", "regulatory_guide", "price_tracker"]',
   '[{"type": "market_analysis", "description": "NBP pricing, forward curves, procurement timing guidance"},
     {"type": "contract_comparison", "description": "Contract term analysis — tolerance bands, exit clauses, deemed rates"},
     {"type": "supplier_directory", "description": "Filterable directory with complaint data and financial stability"},
     {"type": "benchmarking_tool", "description": "Consumption comparison by building type and size"},
     {"type": "regulatory_guide", "description": "CCL, ESOS, SECR compliance guidance"}]',
   '{"primary": "qualified_lead_gen", "secondary": ["b2b_display", "supplier_directory_fees"], "lead_value_gbp": "30-60"}'),
   
  ('finance_mortgage', 'Finance & Mortgage', 'Mortgage calculators, rate analysis, lender comparison',
   'finance_mortgage',
   '["pra_rules", "fca_mcob", "boe_data", "ice_swap_rates", "moneyfacts", "lender_criteria_docs"]',
   '["calculator_tool", "affordability_guide", "rate_analysis", "specialist_situation", "product_comparison", "lender_comparison"]',
   '[{"type": "calculator_tool", "description": "Interactive calculator — repayment, affordability, overpayment, scenario"},
     {"type": "rate_analysis", "description": "Swap rate explanation, monthly outlook, fix vs tracker guidance"},
     {"type": "affordability_guide", "description": "How lenders assess income — self-employed, bonus, complex"},
     {"type": "specialist_situation", "description": "Non-standard construction, short leases, adverse credit, new build"},
     {"type": "product_comparison", "description": "Total cost analysis beyond headline rate — fees, ERCs, overpayment terms"}]',
   '{"primary": "broker_lead_gen", "secondary": ["comparison_affiliate", "financial_display"], "lead_value_gbp": "50-150"}'),
   
  ('seasonal_gifts', 'Seasonal & Gifts', 'Gift guides, seasonal commerce, affiliate content',
   'seasonal_gifts',
   '["retail_trend_reports", "google_shopping_trends", "affiliate_programme_data"]',
   '["gift_guide", "product_roundup", "seasonal_landing", "gift_finder_tool"]',
   '[{"type": "gift_guide", "description": "Curated gift list by recipient/price/interest with affiliate links"},
     {"type": "seasonal_landing", "description": "Seasonal hub page linking to all relevant guides"},
     {"type": "gift_finder", "description": "Interactive quiz/filter tool for gift selection"}]',
   '{"primary": "product_affiliate", "secondary": ["seasonal_display"], "commission_range": "3-17%"}'),
   
  ('generic', 'Generic / Unclassified', 'Standard business sites without vertical specialisation',
   'generic',
   '["competitor_analysis", "keyword_research"]',
   '["standard_brochure", "standard_landing"]',
   '[]',
   '{"primary": "adsense", "secondary": ["generic_affiliate"]}');
  ```

---

## Phase 3: Classifier and Planner Prompt Updates

This is where the vertical intelligence enters the existing pipeline. The classifiers and planners need to understand verticals, recommend research, and produce richer work item specs.

### 3.1 Site Classifier Updates

The existing `site-classifier` currently outputs: `site_type`, `confidence`, `reasoning`, `recommended_builder`, `detected_industry`, `detected_signals`. It needs to also output vertical classification and disposition.

- [ ] **Update site-classifier prompt** to add vertical classification output. Add to the return JSON spec:
  ```
  "vertical_slug": "veterinary|energy_wholesale|finance_mortgage|seasonal_gifts|generic",
  "vertical_confidence": 0.0-1.0,
  "disposition": "develop|sell_as_domain|hold_for_review",
  "disposition_reasoning": "why this recommendation",
  "domain_interpretation": {
    "primary": "what this domain most likely represents",
    "alternatives": ["other possible interpretations"],
    "chosen_reasoning": "why the primary interpretation was selected"
  }
  ```
  The prompt should include guidance on domain ambiguity (as per the design.co.uk analysis) and the principle that expectation match × revenue potential × achievable traffic determines the best interpretation.

- [ ] **Update site-classifier prompt to include the domain interpretation framework** from the content strategy work. Add the three scoring dimensions: visitor expectation match (1-10), revenue per converting visitor, and addressable traffic volume. The classifier should reason through these for ambiguous domains.

### 3.2 Domain Research Classifier Updates

The `domain-research-classifier` currently scrapes the domain and produces identity + classification. It needs to also assess vertical fit and produce richer strategy output.

- [ ] **Update domain-research-classifier prompt** to add:
  - Vertical assessment: which vertical this domain belongs to, with confidence
  - Knowledge gap identification: what the vertical knowledge base would need to contain for this domain (this feeds into research work item creation later)
  - Revenue model assessment: primary and secondary monetisation strategies for this specific domain (pulled from the 15-question framework — who visits, what do they want, how does this niche make money)
  - Competitive landscape: what currently ranks for the domain's keywords, what gap exists
  - Content depth requirements: what E-E-A-T level the niche demands (health/finance need deeper authority than general business)

- [ ] **Add vertical_slug to the write_site_spec output** so downstream agents (planner, content writers) can look up vertical configuration.

### 3.3 Domain Strategist — New Agent or Expanded Classifier

Currently the domain-research-classifier creates a `needs_strategy` work item pointing to a `domain-strategist` handler. This strategist needs vertical awareness.

- [ ] **Create or update domain-strategist agent definition.** The strategist should:
  - Read vertical_registry for the classified vertical
  - Query `rag_lookup` against the vertical's collection to see what knowledge exists
  - Identify what research is needed for this specific domain
  - Output a domain strategy spec that includes:
    - Confirmed vertical_slug
    - Page type recommendations from the vertical's page_type_library
    - Specific research needs (what knowledge clusters need filling)
    - Monetisation hooks to include
    - Competitive positioning notes
    - Content depth and authority requirements

- [ ] **Write domain-strategist prompt** that incorporates the 15-question content strategy framework. The prompt should walk through the three layers:
  - Layer 1: Who visits this domain? What do they want? What makes them leave satisfied? How does money flow?
  - Layer 2: What do top-ranking sites have? What's the buying journey? What questions do people ask? What makes it bookmarkable?
  - Layer 3: What's the best page on the internet for this topic and why? What can this page have that others don't? What format serves the intent? What should the visitor do next?

### 3.4 Site Planner Updates

The site-planner currently produces pages with component sections. It needs to produce richer output that includes vertical-specific page types, research work items, and RAG-aware content specs.

- [ ] **Update site-planner prompt** to accept vertical configuration. When `vertical_slug` is present in site_specs, the planner should:
  - Load the vertical's page_type_library from vertical_registry (via a `query_database` step before the LLM call)
  - Use the vertical's page types in its plan instead of only generic brochure pages
  - Include `rag_collection` and `rag_queries` in each page's spec so content writers know what knowledge to retrieve
  - Include monetisation hooks from the vertical config in page specs (e.g., which pages get insurance affiliate CTAs, which get lead gen forms)

- [ ] **Update site-planner prompt to produce research work items.** When the domain strategist has identified knowledge gaps, the planner should create `needs_vertical_research` work items at priority 1-4, with content page items depending on them. The planner prompt needs to understand the dependency chain:
  ```
  Research items (priority 1-4) → Content items depend on research (priority 10-17) → Assembly (priority 20)
  ```

- [ ] **Update WriteBuildItemsAction** (or create a new action) to handle the new item types:
  - `needs_vertical_research` items with spec containing: research_type, collection, targets, source_types, knowledge_types
  - Content page items with enriched spec containing: rag_collection, rag_queries, content_pattern (from vertical), monetisation_hooks
  - Set `depends_on` correctly — content items that need research should reference the research item UUIDs

- [ ] **Add page_type awareness to the planner.** Currently the planner outputs pages with sections arrays of component names. For vertical page types (breed_health_profile, calculator_tool, market_analysis), the planner should output the `page_type` field so the dispatch loop can route to the appropriate handler. Some page types need specialist handlers (calculator builder, directory builder) rather than the standard page-content-writer.

### 3.5 Vertical-Specific Planner Variants

For domains where the vertical is well-established, a vertical-specific planner produces better results than a generic planner with vertical config injected. These are separate agent definitions using the same Go code but different prompt templates.

- [ ] **Create veterinary site planner prompt variant.** Knows about breed health profiles, procedure guides, practice directories, insurance comparison pages, cost transparency guides. Understands the conversion funnel: informational content (breed health, procedure info) → decision content (vet comparison, quality indicators) → action content (find a vet, get insurance quote). Includes specific guidance on:
  - Every breed health page should link to "find a vet for this breed"
  - Every procedure guide should include cost ranges and insurance relevance
  - Directory pages need structured data for local search
  - Insurance comparison pages need affiliate integration points

- [ ] **Create energy site planner prompt variant.** Knows about market analysis pages, contract guides, supplier directories, benchmarking tools. Understands the B2B buyer journey: education (how gas pricing works) → evaluation (contract comparison, supplier analysis) → action (get quotes, request audit). Includes specific guidance on:
  - Price tracker page as recurring-value hook (updated monthly)
  - Contract guides should explain what to negotiate, not just what terms mean
  - Supplier directory should include complaint data and financial stability
  - Benchmarking tool captures consumption data that qualifies leads

- [ ] **Create mortgage site planner prompt variant.** Knows about calculator tools, affordability guides, rate analysis, specialist situations. Understands the long-tail strategy: specific situation pages (mortgage on 30k salary, mortgage after CCJ) capture high-intent searchers that big sites ignore. Includes specific guidance on:
  - Every calculator should have lead capture below results
  - Rate analysis page updated monthly with swap rate data
  - Specialist situation pages need specific lender criteria from knowledge base
  - Affordability pages should show how different lenders assess differently

- [ ] **Create seasonal/gift site planner prompt variant.** Knows about gift guide structures, recipient/price/interest segmentation, affiliate link density, seasonal content timing. Understands that the site needs year-round content (birthday, anniversary, Mother's Day, Father's Day) alongside the Christmas peak. Includes guidance on:
  - Each guide should link to 10-20 products from high-commission retailers first (John Lewis 5-7%, experience companies 10-17%), Amazon as fallback (3-4%)
  - Guides need annual refresh — date them and plan for updates
  - Category hub pages (gifts for him, gifts for her) are evergreen anchors
  - Seasonal landing pages (Christmas, Valentine's) are time-limited traffic magnets

### 3.6 Content Writer Prompt Updates

The page-content-writer prompt needs to use RAG context effectively and follow vertical content patterns.

- [ ] **Update page-content-writer prompt** to:
  - Inject RAG context as authoritative domain knowledge (already covered in Phase 1.2)
  - Follow the `content_pattern` from the page spec when present (e.g., breed_health_profile pattern structures content as: overview → conditions → screening timeline → choosing a vet → costs)
  - Include `monetisation_hooks` from the spec as natural calls-to-action within content (not bolted-on ads, but genuinely helpful next steps — "compare pet insurance for this breed" after discussing common health costs)
  - Adjust depth and authority level based on the vertical's E-E-A-T requirements (health/finance content needs more careful sourcing and hedging than general business content)

- [ ] **Create vertical-specific content writer prompt variants** for verticals where content patterns differ significantly:
  - Veterinary: clinical accuracy matters, use proper medical terminology with lay explanations, cite knowledge base sources, include cost ranges with explanations of variance
  - Energy: B2B tone, procurement-focused language, include regulatory references, explain market mechanisms without jargon
  - Mortgage: careful with financial advice disclaimers, explain calculations step-by-step, reference specific lender policies from knowledge base, address reader anxiety (mortgage decisions are stressful)
  - Gift guides: conversational tone, persuasive product descriptions, clear price points, natural affiliate link placement, seasonal urgency without being pushy

---

## Phase 4: Research Handler Agent

This creates the agent that processes `needs_vertical_research` work items.

### 4.1 Research Handler Definition

- [ ] **Create `vertical-research-handler` agent definition.** This handler receives a research work item spec and populates the knowledge base. Workflow:
  ```
  receive_spec
    → identify_sources       (from spec.source_types + vertical_registry.research_sources)
    → scrape_sources          (spawn webscrape agents for each source URL)
    → parse_content           (extract text from HTML/PDF responses)
    → extract_knowledge       (LLM: raw content → structured knowledge chunks with metadata)
    → validate_quality        (LLM: check coherence, flag low-confidence extractions)
    → index_knowledge         (rag_index with collection, source_authority, vertical_slug)
    → complete                (report what was indexed)
  ```

- [ ] **Write knowledge extraction prompt.** This is the LLM step that converts raw scraped content into structured knowledge chunks. The prompt needs to:
  - Extract factual claims with source attribution
  - Identify numerical data (costs, statistics, rates) and tag as `knowledge_type: "pricing"` or `"factual"`
  - Distinguish regulatory requirements from general guidance
  - Flag uncertain or contradictory information rather than silently choosing one version
  - Output chunks of 200-500 words each, suitable for RAG retrieval

- [ ] **Write quality validation prompt.** Checks extracted knowledge against the vertical's existing knowledge base for contradictions. Scores each chunk on confidence (1-5). Chunks below threshold get flagged rather than indexed.

- [ ] **Register the handler** so the dispatch loop can route `needs_vertical_research` items to it:
  ```sql
  -- The handler_agent field on work items will be 'vertical-research-handler'
  -- The dispatch loop's dynamic agent_type_field resolution handles the rest
  ```

### 4.2 Research Source Configuration

- [ ] **Create source URL lists per vertical** (stored in vertical_registry.research_sources or a separate table). Each source entry needs:
  - URL pattern or search query to find the source
  - Expected content type (HTML page, PDF document, data table)
  - Source authority level (1-5)
  - Refresh frequency (how often to re-scrape)
  - Extraction hints (what parts of the page contain useful content vs navigation/boilerplate)

- [ ] **Extend webscrape action** if needed to handle PDF extraction and more robust retry logic for research use cases. The existing webscrape adapter may need PDF support added.

---

## Phase 5: Vertical Build Orchestrators

These orchestrate the full build for a domain within a specific vertical.

### 5.1 Routing from Main Pipeline

- [ ] **Add vertical routing to the main pipeline.** After classification, look up the vertical:
  ```json
  "lookup_vertical": {
      "action": "query_database",
      "config": {
          "query": "SELECT * FROM vertical_registry WHERE vertical_slug = $1",
          "params": ["classification.vertical_slug"]
      },
      "next_step": "check_vertical_exists",
      "output_field": "vertical_config"
  }
  ```
  If a vertical exists and has a `build_orch_type`, spawn it. Otherwise fall through to the generic pipeline.

- [ ] **Create vertical-build-orchestrator agent definitions** — one per vertical. Each is an orchestrator with a workflow that:
  1. Loads vertical config from vertical_registry
  2. Runs the vertical's domain strategist (or the generic one with vertical context)
  3. Checks knowledge base coverage via rag_lookup
  4. Creates research work items for gaps
  5. Runs the vertical's planner variant
  6. Dispatches build via the standard dispatch loop

  Initially these can be thin wrappers that set vertical context and delegate to existing agents. They grow into fully specialised orchestrators as each vertical matures.

### 5.2 Knowledge Gap Detection

- [ ] **Create a `check_knowledge_coverage` action or workflow step.** Given a domain and vertical, this queries the knowledge base to determine what knowledge is available and what gaps exist. Uses the vertical's page_type_library to determine what knowledge is needed:
  - For a veterinary domain about French Bulldogs: do we have breed health data for French Bulldogs? Do we have brachycephalic procedure information? Do we have vet cost ranges?
  - For an energy domain: do we have current NBP pricing analysis? Do we have contract term comparisons? Do we have supplier complaint data?

  Gaps become `needs_vertical_research` work items.

---

## Phase 6: Knowledge Maintenance

### 6.1 Refresh Scheduling

- [ ] **Create a knowledge-refresh-scheduler agent** (or extend the existing maintenance-catch-all). Reads refresh_schedule from vertical_registry and creates research work items for stale knowledge:
  - Gas pricing data: monthly refresh
  - Swap rate analysis: weekly refresh
  - Regulatory changes: quarterly check
  - Breed health surveys: annual refresh
  - Supplier complaint data: quarterly refresh

- [ ] **Add `last_refreshed` tracking per knowledge cluster** — either a column on knowledge_base or a separate tracking table.

### 6.2 Knowledge Quality Monitoring

- [ ] **Create monitoring queries** for knowledge base health:
  ```sql
  -- Chunks per vertical, with staleness
  SELECT vertical_slug, knowledge_type,
         COUNT(*) as chunks,
         AVG(source_authority) as avg_authority,
         MIN(created_at) as oldest,
         MAX(created_at) as newest
  FROM knowledge_base
  GROUP BY vertical_slug, knowledge_type;
  
  -- Missing embeddings (need re-embedding)
  SELECT collection, COUNT(*) as missing
  FROM knowledge_base WHERE embedding IS NULL
  GROUP BY collection;
  ```

---

## Phase 7: Fine-Tuning Pipeline (Deferred)

This becomes relevant after 200+ successful examples per agent type accumulate in llm_call_log.

- [ ] **Monitor training data accumulation:**
  ```sql
  SELECT agent_type, COUNT(*) as examples
  FROM llm_call_log WHERE success = true
  GROUP BY agent_type ORDER BY examples DESC;
  ```

- [ ] **First fine-tuning candidate: site-classifier.** Short structured output, runs every build, high volume. Export training data, LoRA fine-tune on GPU, GGUF export, copy GGUF to ollama-models PVC, create model via `ollama create`, add to init container model list (see 0.3.4), update agent definition to `provider: ollama`.

- [ ] **Second candidate: domain-research-classifier.** Structured extraction from domain research.

- [ ] **Third candidate: vertical-specific knowledge extractor.** The LLM step in the research handler that converts raw scraped content into structured knowledge chunks — this is a good fine-tuning target because the task is well-defined and repeated.

---

## Phase 8: Physical Cluster Separation (Deferred)

Only when research workload justifies separate infrastructure.

- [ ] **Deploy research cluster** (second K8s cluster or dedicated node pool).
- [ ] **Move research orchestrators** to research cluster via `dispatch_agent`.
- [ ] **Ensure shared Postgres access** for knowledge_base reads/writes.
- [ ] **Verify cross-cluster Kafka** for orchestration coordination.

---

## Dependency Summary

```
Phase 0 (RAG Pipeline)
  ├── 0.1 DB Migrations
  ├── 0.2 Go Code ──────────────────┐
  ├── 0.3 Ollama Adapter            │ kustomize base + overlay + Makefile targets
  │     (init container pulls       │ deploy via: make deploy-ollama-adapter
  │      nomic-embed-text)          │
  ├── 0.4 Verify Logging            │ all depend on 0.1 + 0.2
  └── 0.5 Test RAG ─────────────────┘ also depends on 0.3
          │
Phase 1 (First Knowledge Content)
  ├── 1.1 Index Canine Biology ──── depends on 0.5
  └── 1.2 Wire RAG to Content ──── depends on 0.5
          │
Phase 2 (Schema Extensions)
  ├── 2.1 Schema Additions ──────── depends on 0.1 (knowledge_base must exist)
  └── 2.2 Seed Vertical Registry ── depends on 2.1
          │
Phase 3 (Prompt Updates) ──────────── depends on 2.2 (needs vertical_registry)
  ├── 3.1 Site Classifier
  ├── 3.2 Domain Research Classifier
  ├── 3.3 Domain Strategist
  ├── 3.4 Site Planner
  ├── 3.5 Vertical Planner Variants
  └── 3.6 Content Writer Updates
          │
Phase 4 (Research Handler) ────────── depends on 0.5 + 2.1
  ├── 4.1 Handler Definition
  └── 4.2 Source Configuration
          │
Phase 5 (Vertical Orchestrators) ──── depends on 3.x + 4.x
  ├── 5.1 Routing
  └── 5.2 Knowledge Gap Detection
          │
Phase 6 (Maintenance) ────────────── depends on 5.x (needs verticals running)
Phase 7 (Fine-tuning) ────────────── depends on 0.4 (needs logging data accumulated)
Phase 8 (Cluster Separation) ─────── depends on 5.x (needs verticals running at scale)
```

Phases 1 and 2 can run in parallel. Phase 3 items are largely independent of each other (classifier updates don't depend on planner updates). Phase 4 can start as soon as 0.5 and 2.1 are done, in parallel with Phase 3 work.

The fastest path to a working vertical: 0.1 → 0.2 → 0.3 → 0.5 → 1.1 → 1.2 → 2.1 → 3.1 → 3.4 → test with vetcomparison.uk.
