# RAG Pipeline Deployment Bundle

## File Placement

```
YOUR_REPO/
├── platform/
│   ├── aiservice/
│   │   └── ollama.go                          ← NEW FILE (from go/platform/aiservice/)
│   ├── orchestration/
│   │   ├── actions/
│   │   │   ├── llm_call_logger.go             ← NEW FILE (from go/platform/orchestration/actions/)
│   │   │   ├── rag_actions.go                 ← NEW FILE (from go/platform/orchestration/actions/)
│   │   │   ├── ai_actions.go                  ← PATCH (see patches/PATCHES.md)
│   │   │   └── registry.go                    ← PATCH (see patches/PATCHES.md)
│   │   └── datahelpers/
│   │       └── nullable_helpers.go            ← NEW FILE (from go/platform/orchestration/datahelpers/)
│   └── database/
│       └── migrations/
│           ├── 081_llm_model_upgrades_and_logging.sql  ← from sql/
│           └── 082_rag_knowledge_base.sql              ← from sql/
├── deployments/
│   └── kustomize/
│       └── services/
│           └── ollama-adapter/                ← ENTIRE DIRECTORY (from kustomize/ollama-adapter/)
│               ├── base/
│               │   ├── deployment.yaml
│               │   ├── service.yaml
│               │   ├── pvc.yaml
│               │   └── kustomization.yaml
│               └── overlays/
│                   └── production/
│                       └── uk_001/
│                           └── kustomization.yaml
└── Makefile                                   ← PATCH (add deploy-ollama-adapter target)
```

## Deployment Steps

### Step 1: Run SQL Migrations

```bash
kubectl -n ai-persona-system exec -it deploy/postgres-clients -- \
    psql -U clients_user -d clients_db

\i platform/database/migrations/081_llm_model_upgrades_and_logging.sql
\i platform/database/migrations/082_rag_knowledge_base.sql
```

Verify:
```sql
-- Check tables exist
SELECT count(*) FROM llm_call_log;
SELECT count(*) FROM knowledge_base;

-- Check model upgrades
SELECT type, default_config->'ai_service'->>'model' as model
FROM agent_definitions
WHERE type IN ('chief-strategist', 'site-planner', 'domain-research-classifier')
ORDER BY type;

-- Check stats views
SELECT * FROM llm_call_stats;
SELECT * FROM knowledge_base_stats;
```

### Step 2: Add New Go Files + Apply Patches

1. Copy the 4 new Go files to their locations (see file placement above)
2. Apply the 3 patches described in patches/PATCHES.md:
   - PATCH 01: anthropic.go — add Usage struct + token write-back
   - PATCH 02: ai_actions.go — add timing/logging + ollama case in createAIClient
   - PATCH 03: registry.go — register rag_lookup and rag_index

3. Build and push:
```bash
docker build -t docker.io/aqls/agent-chassis:v1.0.XXX .
docker push docker.io/aqls/agent-chassis:v1.0.XXX
```

4. Update image tag:
```bash
make update-kustomization-images IMAGE_TAG=v1.0.XXX
make deploy-agents IMAGE_TAG=v1.0.XXX
```

### Step 3: Verify LLM Logging

Run any site build, then:
```sql
SELECT agent_type, step_name, model,
       input_tokens, output_tokens, latency_ms,
       LEFT(response_text, 80) as preview
FROM llm_call_log ORDER BY created_at DESC LIMIT 20;
```

### Step 4: Deploy Ollama Adapter

Add the Makefile target (see patches/PATCHES.md), then:
```bash
make deploy-ollama-adapter ENVIRONMENT=production REGION=uk001
```

Watch the init container pull the model:
```bash
kubectl -n ai-persona-system get pods -l app=ollama-adapter -w
```

Verify:
```bash
# Check model is loaded
kubectl -n ai-persona-system exec -it deploy/ollama-adapter -- ollama list

# Test embedding
kubectl -n ai-persona-system exec -it deploy/ollama-adapter -- \
    curl -s http://localhost:11434/api/embeddings -d '{
      "model": "nomic-embed-text",
      "prompt": "test embedding"
    }' | head -c 200
```

### Step 5: Test RAG End-to-End

Insert test data:
```sql
INSERT INTO knowledge_base (collection, content, metadata)
VALUES ('test', 'French Bulldogs are brachycephalic breeds prone to airway obstruction.
They require specialist anaesthesia protocols and monitoring.',
'{"source": "test", "topic": "breed_health"}');
```

Test rag_index via workflow (indexes with embedding):
```json
{
    "action": "rag_index",
    "config": {
        "content_field": "input_data.content",
        "collection": "test",
        "chunk_size": 500,
        "embedding_service": {
            "provider": "ollama",
            "model": "nomic-embed-text",
            "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
        }
    }
}
```

Test rag_lookup via workflow:
```json
{
    "action": "rag_lookup",
    "config": {
        "query_field": "input_data.query",
        "collection": "test",
        "top_k": 3,
        "embedding_service": {
            "provider": "ollama",
            "model": "nomic-embed-text",
            "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
        }
    }
}
```

Verify embedding was stored:
```sql
SELECT id, collection, LEFT(content, 80) as preview,
       embedding IS NOT NULL as has_embedding,
       embedding_model
FROM knowledge_base
WHERE collection = 'test';
```

### Makefile Additions

Add to your Makefile:

```makefile
.PHONY: deploy-ollama-adapter
deploy-ollama-adapter: ## Deploy ollama-adapter using kustomize
	@echo "$(YELLOW)Deploying ollama-adapter...$(NC)"
	KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k \
	    $(KUSTOMIZE_DIR)/services/ollama-adapter/overlays/$(OVERLAY_PATH)
```

Add to `deploy-agents` target:
```makefile
	# Deploy ollama-adapter
	@echo "Deploying ollama-adapter..."
	@if [ -d "$(KUSTOMIZE_DIR)/services/ollama-adapter/overlays/$(OVERLAY_PATH)" ]; then \
	    KUBECONFIG=$(KUBECONFIG_PATH) kubectl apply -k \
	        $(KUSTOMIZE_DIR)/services/ollama-adapter/overlays/$(OVERLAY_PATH); \
	fi
```

Add to `redeploy-agents` target:
```makefile
	kubectl rollout restart deployment ollama-adapter -n ai-persona-system
```

## Notes

- The ollama-adapter uses `ollama/ollama` image, NOT `docker.io/aqls/ollama-adapter`.
  Do NOT add it to `update-kustomization-images` (that targets aqls/* images).
- The init container is idempotent — if models are already on the PVC, pull completes instantly.
- First deploy will take a few minutes while nomic-embed-text (~300MB) downloads.
- Memory limit of 8Gi accommodates nomic-embed-text (~1GB) plus future 7B models (~4GB).
- To add models later, update the init container args in deployment.yaml.
