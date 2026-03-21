Good, let me walk through how the pieces connect now that they're running.

**What's sitting in the cluster right now**

You have three new things deployed:

The **Ollama adapter** is a pod running the Ollama inference server with `nomic-embed-text` loaded (the init container pulled it on first boot). It's accessible at `http://ollama-adapter.ai-persona-system.svc.cluster.local:11434` from any other pod in the namespace. It does one thing for now: takes text in, returns a 768-dimensional vector out. That vector is a numerical representation of the meaning of the text — similar texts produce similar vectors.

The **knowledge_base table** is sitting empty in Postgres, waiting for content. It has a `vector(768)` column that stores those embeddings alongside the text chunks, and an ivfflat index that makes similarity searches fast.

The **llm_call_log table** is capturing every LLM call from every agent — prompt, response, token counts, latency. This started logging the moment you deployed the new chassis image.

**How rag_index works (writing knowledge)**

When an agent workflow includes a `rag_index` step, it takes a block of text (from a scrape, from research, from manual input), and does this:

1. Splits the text into chunks (default 1000 characters with 200 character overlap, breaking at sentence boundaries where possible)
2. For each chunk, computes a SHA256 hash for deduplication
3. Sends each chunk to Ollama at `/api/embeddings` — Ollama returns a 768-float vector
4. Inserts the chunk + embedding + metadata into the `knowledge_base` table
5. If a chunk with the same hash already exists in that collection, the insert is silently skipped (dedup via the unique index on `collection + content_hash`)

The workflow config looks like this — say you've scraped a vet website and want to index the content:

```json
"index_scraped": {
    "action": "rag_index",
    "config": {
        "content_field": "scraped_data.text",
        "collection": "veterinary",
        "industry_field": "input_data.industry",
        "domain_field": "input_data.domain",
        "source_url_field": "scraped_data.url",
        "chunk_size": 1000,
        "chunk_overlap": 200,
        "embedding_service": {
            "provider": "ollama",
            "model": "nomic-embed-text",
            "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
        }
    },
    "next_step": "next_thing",
    "output_field": "index_result"
}
```

The action reads the text from `scraped_data.text` in the workflow's collected data, chunks it, embeds it, and stores it tagged with `collection: "veterinary"`. If Ollama is down or embedding fails, the chunks still get stored — just without embeddings. They're still searchable via the trigram text index as a fallback.

**How rag_lookup works (reading knowledge)**

When an agent workflow includes a `rag_lookup` step before generating content, it does this:

1. Takes the query text (from a field in collected data — could be the page topic, the domain name, a specific question)
2. Sends it to Ollama to get a query embedding (same 768-float vector)
3. Runs a vector similarity search against the `knowledge_base` table: "find the 5 chunks in this collection whose embeddings are closest to my query embedding"
4. Returns the matching chunks as both structured results and a combined `rag_context` string ready for prompt injection

The workflow config:

```json
"lookup_knowledge": {
    "action": "rag_lookup",
    "config": {
        "query_field": "current_page.rag_query",
        "collection": "veterinary",
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

The output looks like:

```json
{
    "rag_results": [
        {"content": "French Bulldogs are brachycephalic...", "title": "...", "similarity": 0.87},
        {"content": "Soft palate resection is the most common...", "similarity": 0.82}
    ],
    "rag_context": "[Source 1: Breed Health]\nFrench Bulldogs are brachycephalic...\n\n---\n\n[Source 2]...",
    "result_count": 5,
    "search_method": "vector"
}
```

If Ollama is down, it falls back to trigram text search (Postgres `%` operator) — less precise but still functional. The `search_method` field tells you which path it took.

**How it feeds into content generation**

The content writer's prompt template includes a conditional block:

```
{{if .knowledge_context.rag_context}}
## Domain Knowledge
Use the following verified knowledge to inform your content.
This is authoritative information — prefer it over general knowledge.

{{.knowledge_context.rag_context}}
{{end}}
```

So when knowledge exists, the LLM gets it as context alongside the normal page spec. When knowledge doesn't exist (empty collection, no relevant matches), the conditional block is skipped and the content writer works exactly as it does now — from general LLM knowledge.

**The complete flow for a page build with RAG**

```
Work item: needs_content_page
  spec: {
    "name": "french-bulldog-health",
    "rag_collection": "veterinary",
    "rag_query": "french bulldog health brachycephalic conditions screening"
  }

Dispatch loop claims item → spawns page-content-writer handler

Handler workflow:
  1. load_page_components (existing)
  2. build_render_context (existing)
  3. lookup_knowledge ← NEW STEP (rag_lookup)
       → sends "french bulldog health brachycephalic..." to Ollama
       → Ollama returns embedding vector
       → Postgres finds 5 closest chunks in veterinary collection
       → returns rag_context with breed health data, procedure info, cost ranges
  4. generate_content (existing, but now has knowledge_context in template data)
       → LLM receives the page spec PLUS authoritative vet knowledge
       → generates content that references specific conditions, costs, screening schedules
  5. compile_page (existing)
  6. deploy (existing)
```

Steps 1, 2, 4, 5, 6 are unchanged. Step 3 is new. If the rag_collection or rag_query fields are absent from the spec (which they will be for all existing work items), step 3 returns empty context and everything works exactly as before.

**What you can test right now**

The infrastructure is running but the knowledge base is empty. You can verify the chain works:

```sql
-- Insert a test chunk directly
INSERT INTO knowledge_base (collection, content, content_hash, metadata)
VALUES (
    'test',
    'French Bulldogs are brachycephalic breeds with shortened skulls that predispose them to airway obstruction. BOAS (Brachycephalic Obstructive Airway Syndrome) affects up to 50% of the breed. Surgical intervention via soft palate resection has a success rate of approximately 85-90%.',
    encode(sha256('French Bulldogs are brachycephalic...'::bytea), 'hex'),
    '{"source": "manual_test", "topic": "breed_health"}'
);
```

Then embed it via Ollama directly to test the embedding path:

```bash
# Test Ollama embedding from inside the cluster
kubectl -n ai-persona-system exec -it deploy/ollama-adapter -- \
    curl -s http://localhost:11434/api/embeddings -d '{
      "model": "nomic-embed-text",
      "prompt": "French Bulldog health problems"
    }' | python3 -c "import sys,json; d=json.load(sys.stdin); print(f'Got {len(d[\"embedding\"])} dimensions')"
```

To test the full rag_index → rag_lookup cycle through the agent system, you'd create a simple test agent definition with a two-step workflow: first step does rag_index with some test content, second step does rag_lookup querying for related content. But that can wait until you're ready to wire it into the actual content writer.

**What the LLM logging gives you right now**

Every site build that runs is already logging to `llm_call_log`. You can see what's happening:

```sql
-- What's been logged since deployment
SELECT agent_type, step_name, model,
       input_tokens, output_tokens, latency_ms,
       success, LEFT(response_text, 60) as preview
FROM llm_call_log
ORDER BY created_at DESC
LIMIT 20;

-- Cost and speed by agent
SELECT agent_type, model, COUNT(*) as calls,
       ROUND(AVG(latency_ms)) as avg_ms,
       ROUND(AVG(input_tokens)) as avg_in,
       ROUND(AVG(output_tokens)) as avg_out
FROM llm_call_log
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY agent_type, model
ORDER BY calls DESC;
```

This data accumulates toward the fine-tuning threshold (200+ examples per agent type) and gives you immediate visibility into what's happening, what's slow, and what's expensive.

**The path from here to vertical-powered content**

Right now: infrastructure is running, knowledge base is empty, content writers don't have the rag_lookup step yet.

Next steps in order:
1. Get some real knowledge into the base (index the canine biology material into `collection: "veterinary"`)
2. Add the `lookup_knowledge` step to the page-content-writer workflow (one workflow SQL change)
3. Update the planner to include `rag_collection` and `rag_query` in page specs for domains in known verticals
4. Build a page and compare the output with and without RAG context