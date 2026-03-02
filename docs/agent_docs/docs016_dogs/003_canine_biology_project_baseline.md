# Canine Biology Knowledge Tree — Baseline Discussion Document
## Date: 2026-03-02
## Status: Working draft for further iteration

## Project Concept

Build a comprehensive, structured knowledge base covering everything known about Canis lupus familiaris, using the Labrador Retriever as the primary reference breed. The system uses hierarchical agent orchestration to decompose the domain, research the literature, extract findings, synthesise knowledge, generate diagrams, and validate consistency.

The knowledge tree would be populated by ~800,000-1,000,000 autonomous agents, each responsible for a specific piece of the knowledge. Once built, the tree is incrementally updated as new research is published.

This serves as both a genuinely useful output (a navigable veterinary/biological reference) and a demonstration of the agent framework at scale.

## Knowledge Tree Structure

The domain decomposes naturally into a tree with branching factor ~8-12 at each level:

```
Level 0: Root (1 agent)
  "Comprehensive biology of the Labrador Retriever"

Level 1: Major domains (~20 agents)
  Body systems: musculoskeletal, cardiovascular, respiratory, digestive,
  nervous, endocrine, immune/lymphatic, integumentary, urinary, reproductive,
  sensory (organ of corti etc)
  Cross-cutting: nutrition, behaviour, genetics/genomics, breed history,
  aging/geriatrics, exercise physiology, common diseases, veterinary
  procedures, pharmacology

Level 2: Aspects per domain (~200 agents)
  e.g. Cardiovascular → anatomy, physiology, pathology, pharmacology,
  biochemistry, development/embryology, diagnostics, genetics

Level 3: Sub-topics (~2,000 agents)
  e.g. Cardiovascular > Pathology → dilated cardiomyopathy, valvular
  disease, heartworm, arrhythmias, congestive heart failure, pericardial
  effusion, aortic stenosis, patent ductus arteriosus...

Level 4: Specific topics (~20,000 agents)
  e.g. Dilated cardiomyopathy → genetic basis, molecular mechanism,
  clinical presentation, echocardiographic findings, diagnosis protocol,
  treatment protocol, prognosis, breed predisposition data, diet-associated
  DCM

Level 5: Detailed mechanisms (~200,000 agents)
  e.g. Genetic basis of DCM → LMNA mutations, titin (TTN) variants,
  phospholamban (PLN) mutations, troponin mutations, myosin binding
  protein C, genome-wide association studies, heritability estimates

Level 6: Molecular/atomic detail (up to ~1,000,000 agents)
  e.g. LMNA mutations → specific variants identified in Labradors,
  lamin A/C protein structure, nuclear envelope disruption mechanism,
  downstream effects on cardiomyocyte integrity, comparison with human
  LMNA cardiomyopathy, relevant published papers with findings
```

The tree would not be uniform. Some branches are deep and narrow (biochemical pathways), others broad and shallow (breed history). Decomposer agents at each level decide the branching structure based on how much published knowledge exists.

## Reaching 1M Agents

The core Labrador biology tree produces ~800K agents. To reach 1M, options include: covering the top 20 sporting/working dog breeds in parallel (each with breed-specific sub-trees), extending into veterinary practice (treatment protocols, drug monographs), or broadening into adjacent domains (canine nutrition across commercial diets, breed-specific exercise physiology). The knowledge base becomes "comprehensive veterinary reference for sporting dog breeds" rather than strictly one breed.

## Agent Types

Seven distinct agent roles, each with different characteristics:

### 1. Decomposer Agents (Levels 0-3, ~2,020 agents)

Receive a topic and decide how to split it into subtopics. Need genuine reasoning about domain structure — understanding that "cardiovascular pathology" should split by condition type, that some conditions warrant deeper decomposition than others, and consulting the literature to know what's out there before deciding.

Work: receive topic → search for overview papers → reason about decomposition → spawn children → collect results → synthesise.

### 2. Research Agents (Level 4, ~20,000 agents)

Investigate a specific topic in depth. Each handles a topic cluster (a group of related papers) rather than a single paper. Do the core knowledge work.

Work: receive topic cluster → search PubMed/Scholar → fetch papers → analyse findings → structure into knowledge entry with citations.

### 3. Paper Fetcher Agents (non-LLM, ~200,000 agents)

Pure data retrieval. Hit PubMed E-utilities API, Google Scholar, or a pre-downloaded paper corpus. Return abstracts and metadata as structured data.

Work: receive search query → call PubMed API → parse XML → return papers with abstracts and MeSH terms.

No LLM needed. Fast, cheap. Rate-limited by PubMed (10 req/sec with API key).

### 4. Entity Extractor Agents (non-LLM or tiny model, ~200,000 agents)

Structured extraction from paper abstracts using specialised biomedical NER models (SciSpacy, PubMedBERT, ~110M parameters). Run on CPU in milliseconds. Very accurate for biomedical entities.

Work: receive abstract text → run NER → extract genes, proteins, drugs, diseases, species, dosages → extract study type from MeSH terms → return structured entities.

No general LLM needed. Specialised tiny models are better at this than general-purpose 7B models.

### 5. Relevance Filter Agents (embedded 3B, ~200,000 agents)

Classification task: given extracted entities and the query topic, is this paper relevant? Binary yes/no with confidence score. Filters out the 60-70% of search results that aren't actually relevant to the specific question.

This is a task the embedded 3B model handles well — it's classification, not extraction or synthesis.

### 6. Finding Synthesiser Agents (7B biomedical, ~100,000 agents)

The actual analysis step. Given an abstract, pre-extracted entities, study type, and the specific question being answered, write a structured finding: what was found, confidence level, applicability to Labradors specifically, evidence strength.

This is where model quality matters most at the leaf level. The 7B gets a focused, context-rich prompt because the structured extraction and filtering already happened upstream.

### 7. Synthesis Agents (at each tree level, ~20,200 agents)

Combine children's results into coherent entries. Higher levels need better writing quality because these become the readable sections of the knowledge base.

Work: receive collected results from children → write coherent section → identify gaps or contradictions → flag for review.

### 8. Diagram/Illustration Agents (~40,000 agents)

Two sub-types:

**Image generation** (~10,000): Anatomical illustrations, microscopy-style images, comparative diagrams. Use FLUX/SDXL with appropriate style prompts. Some may benefit from medical/anatomical LoRA fine-tunes if available.

**Code-based diagrams** (~30,000): Biochemical pathway diagrams, flowcharts, diagnostic algorithms. LLM generates Mermaid or SVG code, rendered to image. These are editable, searchable, and precise — better than image generation for structured/logical diagrams.

### 9. Validator/Cross-Reference Agents (7B, ~20,000 agents)

Check consistency between sibling branches. Flag contradictions in factual claims, incidence rates, or terminology.

Work: receive entries from related branches → compare claims → identify contradictions → return conflicts with citations.

## LLM Allocation

### The 3B Problem

The original design placed leaf extractors (750K agents) on embedded 3B quantised models. This was questioned and found to be the weak point. Paper analysis is one of the hardest tasks in the pipeline:

- **Domain-specific terminology**: "Canine phospholamban R14del mutation" must be parsed correctly. A 3B model frequently confuses gene names, misattributes variants, or loses species qualifiers.
- **Hedging and evidence strength**: "Our results suggest that taurine supplementation may reverse DCM in a subset of affected dogs" contains hedged findings that a 3B model will flatten to "taurine cures DCM in dogs."
- **Numerical precision**: Incidence rates, p-values, confidence intervals, dosages. Small models are notoriously bad at faithfully reproducing numbers.
- **Specificity**: Distinguishing Labrador-specific vs general canine vs general mammalian findings.
- **Study quality recognition**: RCT with 500 dogs vs case report of 3 dogs.

A 3B model gets ~60-70% of this right. Errors at the leaf level propagate upward through synthesis, contaminating higher-level entries with subtly wrong information — worse than being obviously wrong.

### Current Allocation

The solution: split paper analysis into LLM and non-LLM steps. Use the 3B model only for classification (relevance filtering), where it performs well. Use 7B or biomedical-fine-tuned 7B for actual analysis.

```
Agent type              Count      LLM                Where it runs
────────────────────────────────────────────────────────────────────
Decomposer L0-1         ~20        Anthropic Opus     API call
Decomposer L2-3         ~2,000     General 7B         Inference server
Research agents L4      ~20,000    BioMistral 7B      Inference server
Paper fetchers          ~200,000   None               Worker pod (API only)
Entity extractors       ~200,000   SciSpacy NER       Worker pod (CPU, ms)
Relevance filters       ~200,000   Embedded 3B        Worker pod (local)
Finding synthesisers    ~100,000   BioMistral 7B      Inference server
Synthesis L0-2          ~200       Anthropic Opus     API call
Synthesis L3-4          ~20,000    General 7B         Inference server
Diagrams (image)        ~10,000    FLUX/SDXL          GPU inference server
Diagrams (mermaid)      ~30,000    General 7B         Inference server
Validators              ~20,000    General 7B         Inference server
────────────────────────────────────────────────────────────────────
Total:                  ~802,000
```

### Biomedical Models

Models fine-tuned on PubMed/biomedical text (BioMistral, Med-Llama, Meditron) outperform general-purpose models of the same size on paper analysis, terminology handling, and evidence extraction. These would be hosted on the inference cluster alongside the general 7B, with leaf-level analysis routed to the biomedical model and decomposition/synthesis routed to the general model.

Trade-off: biomedical fine-tunes sometimes lose general reasoning ability. Good at "extract finding X with p-value Y" but worse at "how should I decompose this topic."

## The Paper Analysis Pipeline

Rather than asking one agent to read a paper and extract everything, the pipeline separates structured extraction (cheap, accurate, no LLM) from semantic interpretation (needs a capable model):

```
Paper Fetcher (no LLM)
  → PubMed API, get abstract + full metadata
  → Return structured XML with MeSH terms, species, publication type

Entity Extractor (SciSpacy/PubMedBERT, ~110M params, CPU)
  → NER on abstract text
  → Extract: genes, proteins, drugs, diseases, species, dosages
  → Extract: study type from MeSH "Publication Type" field
  → Return structured entity list

Relevance Filter (embedded 3B, local in worker pod)
  → Given extracted entities + query topic
  → Is this paper relevant? Yes/no with confidence
  → Filters out ~60-70% of search results that aren't relevant
  → Classification task — 3B is good at this

Finding Synthesiser (BioMistral 7B, inference server)
  → Given: abstract, pre-extracted entities, study type, specific question
  → Write structured finding with evidence strength and breed applicability
  → Focused prompt because upstream extraction already happened
  → This is the only step that needs a strong model
```

This design reduces the 7B inference load substantially — each leaf agent makes 1 call to the 7B instead of 2-3, and the call is more focused.

## Infrastructure Design

### Shared Topic Pools (Kafka)

All agents use shared topic pools instead of per-agent topics:

```
system.work.pool.{00-63}  — 64 topics × 16 partitions = 1,024 partitions
```

At 1M agents, ~1,000 agents hash to each partition. At peak concurrency of 20K active agents, ~20 agents active per partition. Blast radius of a single bad agent is limited to ~1,000 co-located agents on one partition. Debugging is manageable — inspect a single partition with kcat.

Messages carry `target_agent_id` in headers. Workers consume their assigned partitions and filter by agent ID.

### Worker Pools

Long-running Pods that execute multiple agent workflows as goroutines. Each Pod loads an embedded 3B model for relevance filtering and runs SciSpacy for entity extraction.

```
Worker Pod (8 CPU cores, 8GB RAM)
├── Embedded 3B model (llama.cpp, ~2GB)
├── SciSpacy NER model (~200MB)
├── Goroutine pool: 5,000-10,000 agent workflows
├── Kafka consumer: reads from assigned pool partitions
└── Routes LLM calls: local 3B for classification,
    remote 7B/Opus for analysis/synthesis
```

### Inference Servers

Shared 7B (general + BioMistral) and image generation on GPU:

```
General 7B (vLLM, continuous batching): ~10 GPUs
BioMistral 7B (vLLM): ~10-20 GPUs
Image generation (FLUX/SDXL): ~5 GPUs
```

### Database

Postgres with Redis/Valkey for hot orchestration state:

- Redis: active workflow state (current step, collected data, status). 100K+ writes/sec.
- Postgres: persistent storage — agent definitions, completed orchestrations, knowledge entries, citations. Written to on completion or periodically, not on every step.
- 1M knowledge entries at ~2-5KB each = 2-5GB structured data
- ~10K images at ~500KB each = 5GB media

### Full Infrastructure Summary

```
Primary cluster (Rackspace):
  Kafka: 15 brokers
  Postgres: 1 instance + Redis/Valkey for hot state
  agent-chassis: 3 pods (top-level orchestration)
  PubMed abstract cache: Redis instance

Worker cluster(s) (2-3 clusters):
  Worker pool pods: 100 total (each 8 cores, 8GB, embedded 3B + SciSpacy)
  remote-job-spawner: 1 per cluster

Inference cluster:
  General 7B (vLLM): 10 GPUs
  BioMistral 7B (vLLM): 10-20 GPUs
  FLUX/SDXL image gen: 5 GPUs

External APIs:
  Anthropic Opus: ~250 calls total (~$8)
  PubMed E-utilities: sustained 10 req/sec
```

## Timeline Estimate

The tree builds depth-first. Early branches complete and begin synthesis while later branches are still in research/extraction.

| Phase | Duration | Notes |
|-------|----------|-------|
| L0-1 decomposition (Opus) | Minutes | ~20 agents, fast |
| L2-3 decomposition (7B) | ~1 hour | ~2,000 agents |
| L4 research + paper fetching | ~6-8 hours | 20K research agents, 200K fetches, PubMed rate-limited |
| Entity extraction + filtering | ~4-6 hours | 400K agents, CPU-only, fast per agent |
| Finding synthesis (7B) | ~8-15 hours | 100K agents on 20-30 GPUs |
| Tree synthesis (bottom-up) | ~6-10 hours | Starts as soon as first branches complete |
| Diagram generation | ~3-5 hours | 40K agents, runs in parallel with synthesis |
| Validation pass | ~4-6 hours | 20K agents, runs after synthesis |
| **Total (with parallelism)** | **30-48 hours** | Bottleneck: finding synthesis phase |

## Cost Estimate

| Component | Duration | Cost |
|-----------|----------|------|
| 100 worker Pods (8-core VMs) × 48hrs | 4,800 VM-hours | $500-1,500 |
| 25-35 GPUs (inference + image) × 48hrs | 1,200-1,700 GPU-hours | $2,500-5,000 |
| Kafka brokers (15) × 48hrs | 720 VM-hours | $200-500 |
| Anthropic Opus (~250 calls) | — | ~$8 |
| PubMed API | — | Free |
| Persistent infrastructure | — | Existing |
| **Total** | | **$3,200-7,000** |

## Output Format

Each agent produces a structured knowledge entry:

```json
{
  "node_id": "uuid",
  "parent_id": "uuid",
  "path": "cardiovascular.pathology.dcm.genetics.lmna",
  "title": "LMNA Gene Mutations in Canine Dilated Cardiomyopathy",
  "level": 6,
  "breed_specificity": "labrador_specific | general_canine | general_mammalian",
  "content": {
    "summary": "...",
    "key_findings": [...],
    "citations": [...],
    "confidence": "high | moderate | low",
    "evidence_base": "strong | moderate | limited | single_study",
    "last_updated": "2026-03-02",
    "contradictions_flagged": [...]
  },
  "media": [
    {"type": "mermaid", "id": "uuid", "caption": "LMNA signaling pathway"},
    {"type": "image", "id": "uuid", "caption": "Nuclear envelope disruption"}
  ],
  "children": ["uuid", "uuid", ...],
  "cross_references": ["uuid", ...]
}
```

Total dataset: ~10GB (structured data + images).

## Incremental Update Mechanism

Once built, a monitoring layer checks for new publications:

- Scheduled agent queries PubMed for new papers matching branch keywords
- New relevant papers trigger leaf-level re-extraction
- Changed findings propagate upward through re-synthesis
- Validators flag new contradictions

Veterinary research publishes ~50-100 relevant papers per week. Update workload: ~500-1,000 agent activations per week — a couple of worker Pods running continuously.

## UI/UX (Not Yet Designed)

The output is a navigable knowledge tree. Possible presentations:

- Web UI: start at root, click to drill down. Each node shows synthesised content, citations, diagrams, children links. Search bar for jumping to any node.
- "Recent updates" panel showing branches refreshed with new research.
- Export as structured JSON, markdown files, or printable reference.
- Mermaid diagrams render in any markdown viewer. Images as standard PNGs.

For the demo recording: real-time tree expansion visualisation — watching branches grow as agents decompose and populate, images appearing as diagram agents complete, agent count climbing toward 1M.

## Open Questions

1. **Scope to reach 1M**: Core Labrador biology produces ~800K agents. What's the right extension — more breeds, veterinary practice, nutrition databases?

2. **3B model adequacy for relevance filtering**: Is classification genuinely safe at 3B, or should we benchmark against 7B on a test set before committing?

3. **BioMistral vs general 7B**: Need to evaluate BioMistral on actual veterinary paper abstracts. It's trained on human biomedical literature — canine-specific terminology coverage may vary.

4. **PubMed rate limiting**: 200K paper fetches at 10/sec = 5.5 hours. Should we pre-download and cache the canine corpus? There are services that provide bulk PubMed access.

5. **Full text access**: Many papers are behind paywalls. The pipeline works on abstracts, but full text would produce richer findings. Options: PubMed Central (open access subset), institutional access, or limit to open-access papers and note where full-text analysis would improve confidence.

6. **Image generation quality**: FLUX/SDXL for anatomical illustrations may need medical/anatomical LoRA fine-tunes. Quality of generated anatomical diagrams needs evaluation — might be better to use diagram code (SVG/Mermaid) for anything that needs precision.

7. **Worker pool implementation**: The pool_agent action and worker pool binary don't exist yet. This is the biggest code deliverable before the project can run at scale.

8. **Redis/Valkey for hot state**: The orchestration engine currently writes to Postgres on every step. The Redis caching layer needs implementation in the coordinator.

9. **UI/UX design**: How to present a 1M-node knowledge tree in a way that's genuinely useful, not just impressive.

10. **Recording the demo**: How to capture the tree building in real-time for promotional material. Dashboard design, metrics to show, narrative structure.

## Architecture Dependencies

Items that need to exist before this project can run:

| Component | Status | Notes |
|-----------|--------|-------|
| Multi-cluster dispatch | Proven working | remote-job-spawner tested 2026-03-02 |
| dispatch_agent action | Code written, not committed | In dispatch_actions.go |
| Registry entry for dispatch_agent | Not done | One line in registry.go |
| pool_agent action | Not started | New action type |
| Worker pool binary/mode | Not started | Biggest code deliverable |
| Shared topic pool creation | Not started | Pre-create 64 topics × 16 partitions |
| Embedded 3B model in worker | Not started | llama.cpp or llama-cpp-go integration |
| SciSpacy NER in worker | Not started | Python sidecar or Go NER library |
| BioMistral 7B deployment | Not started | vLLM config for biomedical model |
| Redis/Valkey hot state layer | Not started | Coordinator change |
| Knowledge entry storage schema | Not started | New table for tree nodes |
| Image generation pipeline | Not started | FLUX/SDXL inference server |
| PubMed API integration | Not started | New action or adapter |
| Incremental update scheduler | Not started | Kafka scheduler pattern exists |
| UI for knowledge tree | Not started | — |

## Related Documents

- 006c: Multi-cluster dispatch handoff summary
- 013: Scaling analysis (10K through 1M)
- 001e: Development guide for new agents
- 002c: System architecture
- 010: Tool library guide
- 011: Kafka scheduler guide
