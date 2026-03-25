# RAG Best Practices for Site Building Pipeline

## Date: 2026-03-24

---

## Core Principle: Filter First, Then Rank by Similarity

The most common RAG failure is retrieving semantically similar but contextually wrong content. A veterinary hero section example retrieved when writing a gas wholesale hero section is worse than no example at all — it actively misleads the model.

**Always filter by structured metadata before doing embedding similarity search:**

```sql
-- WRONG: nearest-embedding across entire knowledge base
SELECT content FROM knowledge_base
ORDER BY embedding <=> query_embedding
LIMIT 3;

-- RIGHT: filter by vertical and component, then rank by similarity
SELECT content, 1 - (embedding <=> query_embedding) as similarity
FROM knowledge_base
WHERE metadata->>'vertical' = 'veterinary'
  AND metadata->>'component_type' = 'hero'
  AND metadata->>'source_quality' IN ('high', 'verified')
ORDER BY embedding <=> query_embedding
LIMIT 3;
```

This means every knowledge base entry needs structured metadata — not just an embedding and a blob of text.

---

## Knowledge Base Entry Structure

Each entry in the `knowledge_base` table should have rich metadata for filtering:

```json
{
  "content": "The actual text, design spec, or structured example",
  "metadata": {
    "vertical": "veterinary",
    "sub_vertical": "practice_comparison",
    "component_type": "hero",
    "content_type": "example",
    "source": "scraped",
    "source_url": "https://example-vet-site.co.uk",
    "source_quality": "high",
    "quality_signals": {
      "domain_age_years": 8,
      "has_real_details": true,
      "ai_content_probability": 0.1
    },
    "created_at": "2026-03-24",
    "prompt_version": "content_writer_v3"
  }
}
```

**Required metadata fields for filtering:**

| Field | Purpose | Values |
|---|---|---|
| `vertical` | Industry category | veterinary, gas_wholesale, consulting, mortgage, etc. |
| `component_type` | What UI component | hero, features_grid, faq, comparison_table, nav, etc. |
| `content_type` | What kind of knowledge | example, design_spec, audit_insight, terminology, fact |
| `source` | Where it came from | scraped, claude_output, human_curated, audit_feedback |
| `source_quality` | Quality gate | high, medium, low, unverified |

---

## How Much RAG Context to Inject

**Rule of thumb: RAG context should be 20-30% of available context window.**

| Context Window | RAG Budget | Roughly |
|---|---|---|
| 4K tokens | 800-1200 tokens | 2-3 short examples |
| 8K tokens | 1600-2400 tokens | 3-5 short examples or 1-2 detailed ones |
| 32K tokens | 6000-9000 tokens | 5-8 detailed examples |
| 131K tokens | Don't fill it | Still 5-8 examples — more isn't better |

**More examples is not always better.** Beyond 3-5 relevant examples, additional context adds noise. The model has to figure out which examples matter. If 3 of 8 injected examples are relevant, the model might pick up patterns from the wrong 5.

**Quality over quantity:** 2 highly relevant examples beat 8 loosely relevant ones.

---

## Embedding Model Choice

Currently using: **nomic-embed-text** (v1, 768 dimensions)

Recommended upgrade: **nomic-embed-text-v2-moe** (same dimensions, better multilingual, MoE architecture, runs in Ollama)

The upgrade is a drop-in replacement:
```bash
# On Ollama
ollama pull nomic-embed-text-v2-moe
```

The `knowledge_base` table already uses 768 dimensions — no schema change needed. Re-embed existing entries with the new model (batch job, one-time).

**When to consider a bigger model:**

| Scenario | Model | Why |
|---|---|---|
| Short entries (<500 tokens), English | nomic-embed-text-v2-moe | Fast, accurate, runs on CPU |
| Long documents (>2000 tokens) | BGE-M3 (568M) | Better long-document retrieval |
| Multilingual content | nomic-v2-moe or BGE-M3 | Both support 100+ languages |
| Code/technical content | Qwen3-Embedding-0.6B | Strong on code retrieval |

For the site building pipeline, nomic-v2-moe is sufficient. Most knowledge base entries are short structured examples (hero text, design specs, component schemas).

**Task prefixes matter.** Nomic models perform significantly better when you prepend the correct prefix:

```
Embedding a query:     "search_query: veterinary hero section example"
Embedding a document:  "search_document: Find the Right Vet for Your Pet..."
```

The embedding action (`rag_index`) should prepend `search_document:` when storing. The retrieval action (`rag_lookup`) should prepend `search_query:` when searching. This is a one-line change in each action.

---

## Knowledge Sources and Quality

### Source 1: Scraped Competitor Sites

The highest-value training data. Real human-written content from live business sites.

**Pipeline:**
```
Scraping adapter crawls top 20-30 sites per vertical
  → LLM extracts structured data (hero text, colors, fonts, CTAs)
  → Quality assessment scores each site (domain age, real details, AI probability)
  → High-quality sites → knowledge_base with source: "scraped", quality: "high"
  → Low-quality/AI sites → stored for competitive intel, NOT for training
```

**AI slop detection:** Score scraped content for authenticity:
- Domain age >3 years: likely human-written
- Specific details (real names, addresses, dates): human signal
- Template/stock photo design: low-quality signal
- Generic phrasing without industry specifics: AI signal

Sites scoring below quality threshold are excluded from training data but can still be stored for competitive analysis.

**HITL option:** With HITL flag on, a human reviews quality assessment before content enters the training set. With HITL off, the system uses the quality threshold automatically.

### Source 2: Successful Claude Outputs

Content produced by Claude that passed audit verification without rewrites.

**Pipeline:**
```
LLM call log captures prompt + response
  → Join to work item outcomes
  → Filter: success=true AND no subsequent rewrite for same section
  → Store in knowledge_base with source: "claude_output", quality: "verified"
```

These are valuable because they're examples of exactly the task you're training for, in exactly the format your pipeline produces. But they cap at Claude's quality level — they can't teach the model to be better than Claude.

### Source 3: Human-Curated Examples

Manually selected or edited examples that represent the ideal output.

Best used for: web design (where both Claude and Llama produce generic results) and content where industry-specific voice matters.

### Source 4: Audit Insights

When the audit agent identifies what makes content good or bad, those insights are themselves valuable knowledge:

```json
{
  "vertical": "veterinary",
  "component_type": "hero",
  "content_type": "audit_insight",
  "content": "Subheadlines that reference specific data types (costs, services, locations) consistently pass audit. Generic benefit claims ('we care about your pet') consistently fail.",
  "source": "audit_feedback"
}
```

These insights can be injected into content writer prompts: "Based on previous audits for this industry, here's what works and what doesn't."

---

## Retrieval Strategy

### At Content Writing Time

```
1. Identify: vertical = "veterinary", component = "hero"
2. Retrieve from knowledge_base:
   - 2 scraped examples of real veterinary hero sections (source: scraped)
   - 1 successful Claude output for same vertical+component (source: claude_output)
   - 1 audit insight about what works (source: audit_feedback)
3. Inject into prompt as context (within 20-30% token budget)
4. Content writer produces output informed by real examples
```

### At Audit Time

```
1. Retrieve audit insights for this vertical+component
2. Inject as context: "Previous audits for veterinary hero sections found these patterns..."
3. Audit agent produces more consistent, industry-aware findings
```

### At Design Time

```
1. Retrieve scraped design specs for this vertical
   - Color schemes from top 5 veterinary sites
   - Typography choices
   - Layout patterns
2. Inject as context: "Here are design choices from successful sites in this industry..."
3. Design agent produces industry-appropriate choices instead of generic defaults
```

---

## Avoiding Common RAG Failures

### Failure 1: Wrong vertical contamination

**Problem:** A gas wholesale example retrieved when writing veterinary content because the embedding similarity was high on "professional comparison" semantics.

**Fix:** Always filter by `vertical` before similarity ranking. The metadata filter is not optional.

### Failure 2: Stale examples

**Problem:** Design examples from 2023 producing outdated aesthetics.

**Fix:** Include `created_at` in metadata. Weight recent examples higher, or filter to last 12 months for design-related content.

### Failure 3: Too much context

**Problem:** Injecting 8 examples consuming 5000 tokens, leaving insufficient space for the actual instruction and output.

**Fix:** Hard cap on RAG token budget. Count tokens before injection. Truncate or reduce example count if over budget.

### Failure 4: Low-quality examples

**Problem:** AI-generated scraped content entering the knowledge base, teaching the model to produce more AI slop.

**Fix:** Quality gate on all entries. `source_quality` must be "high" or "verified" for training use. Unverified entries are searchable for reference but excluded from prompt injection.

### Failure 5: Embedding mismatch

**Problem:** Query embedded with one model, documents embedded with a different model or version. Similarity scores are meaningless across different embedding spaces.

**Fix:** Track `embedding_model` in metadata. When upgrading models, re-embed all entries. Never mix embedding spaces.

```sql
-- Add to knowledge_base if not already present
ALTER TABLE knowledge_base ADD COLUMN IF NOT EXISTS embedding_model TEXT DEFAULT 'nomic-embed-text';
```

---

## Practical Schema for knowledge_base

The existing table has pgvector and basic fields. Ensure these are present:

```sql
-- Check current schema
\d knowledge_base

-- Add metadata fields if missing
ALTER TABLE knowledge_base 
ADD COLUMN IF NOT EXISTS source TEXT DEFAULT 'unknown',
ADD COLUMN IF NOT EXISTS source_quality TEXT DEFAULT 'unverified',
ADD COLUMN IF NOT EXISTS vertical TEXT,
ADD COLUMN IF NOT EXISTS component_type TEXT,
ADD COLUMN IF NOT EXISTS content_type TEXT DEFAULT 'example',
ADD COLUMN IF NOT EXISTS embedding_model TEXT DEFAULT 'nomic-embed-text',
ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT '{}';

-- Indexes for the filter-then-rank pattern
CREATE INDEX IF NOT EXISTS idx_kb_vertical_component 
ON knowledge_base(vertical, component_type) 
WHERE source_quality IN ('high', 'verified');

CREATE INDEX IF NOT EXISTS idx_kb_source_quality 
ON knowledge_base(source_quality);
```

---

## Implementation Priority

1. **Add metadata columns to knowledge_base** (if not present)
2. **Update rag_index action** to require vertical + component_type + source on every insert
3. **Update rag_lookup action** to filter by vertical + component_type before similarity search, and prepend task prefix to query
4. **Upgrade embedding model** to nomic-embed-text-v2-moe (drop-in replacement)
5. **Build scraping extraction pipeline** — scrape top sites per vertical, extract structured examples, quality-score, store
6. **Connect to content writer** — inject 3-5 relevant examples into prompts
7. **Connect to audit agent** — inject audit insights for consistency
8. **Connect to design agent** — inject real design specs for distinctiveness
