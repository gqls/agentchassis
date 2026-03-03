# Canine Biology Knowledge Tree — Baseline Discussion Document
## Date: 2026-03-02
## Status: Working draft for further iteration

## Project Vision

Build a world-leading, regularly-updated biological reference for Canis lupus familiaris, using the Labrador Retriever as the primary reference breed. The system uses hierarchical agent orchestration to decompose the domain, research the literature, extract findings, synthesise knowledge, generate diagrams, validate consistency, and accept corrections.

The knowledge tree would be populated by ~800,000-1,000,000 autonomous agents, each responsible for a specific piece of the knowledge. Once built, the tree is incrementally updated as new research is published and corrections are submitted.

This serves three purposes:
1. **A genuinely useful output** — a navigable, citable veterinary/biological reference that is accurate where it has coverage, even if not initially complete.
2. **A demonstration of the agent framework at scale** — 1M agents running over days, producing real output.
3. **A selling tool for the framework itself** — every entry is traceable to a specific agent with a visible prompt, and any piece can be regenerated or improved without rebuilding the whole thing. The machinery is as much the product as the output.

## Design Priorities

**Accuracy over completeness.** The first run does not need to cover every branch. What it does cover must be defensible — errors in a resource people cite or rely on destroy trust permanently. Better to have 50,000 deeply accurate entries than 1M shallow ones. The 1M agent count is the infrastructure story; the output quality is what makes it a resource rather than a stunt.

**Nothing that produces text a reader will see should come from a 3B model.** The 3B embedded models are used only for classification (relevance filtering). All factual extraction, synthesis, and writing uses 7B or larger models.

**Phased coverage, not big-bang.** The first release covers a subset of branches in depth (e.g. cardiovascular, musculoskeletal, nervous system). Remaining branches are structurally present in the tree but marked as "identified, not yet researched." These fill in over subsequent weeks as background processing. This is not a limitation — it's a feature of a living, regularly-updated resource.

**Every entry is auditable.** Each node stores which agent produced it, what prompt was used, what sources were consulted, and what model generated the text. This traceability is both a quality control mechanism and a selling point for the framework.

## Honest Assessment of Risks

Before the technical detail, some candid concerns about the project as a whole.

**The "world-leading resource" claim is aspirational, not assured.** Existing veterinary references (Plumb's, BSAVA Formulary, VIN, Merck Veterinary Manual) have decades of expert editorial oversight. An AI-generated resource, however well-structured, starts at zero credibility with the veterinary community. The system can produce comprehensive coverage faster than any human team, but coverage is not the same as accuracy, and accuracy is not the same as trust. Building trust requires expert review, clinical validation over time, and demonstrated reliability. The first version is more honestly described as "a large-scale AI-generated draft for expert review" than "a world-leading reference."

**The 1M agent count is partly theatrical.** The biology tree produces ~800K agents, but a significant fraction of those are paper fetchers, entity extractors, and relevance filters — agents that do simple, fast work. The actually hard work (decomposition, finding synthesis, postulation) is done by ~150K agents. The 1M number is real for the infrastructure demo but shouldn't obscure the fact that most agents are doing lightweight work. This is fine for demonstrating the framework, but the marketing should be honest about what "1M agents" actually means in terms of cognitive work.

**LLM-generated scientific content is inherently risky.** Even the best models hallucinate, misattribute findings, and produce confident-sounding errors. A single hallucinated gene name or misquoted incidence rate, buried in a 1M-node tree, could persist for months before anyone notices. The mitigation (structured extraction pipelines, validation agents, expert review) reduces this risk but doesn't eliminate it. Any public-facing version needs prominent disclaimers about AI generation and should not position itself as a clinical decision-making tool until it has undergone external validation.

**The front end determines whether this is useful or just impressive.** A million nodes of perfectly accurate content are worthless if the UI makes them unfindable. And a beautiful UI showing slightly wrong content is actively harmful. The front end is not a nice-to-have — it's what determines whether the project succeeds as a resource (not just as a demo). It deserves serious design investment, probably with input from actual veterinarians and researchers, not just engineers.

**Cost estimates may be optimistic.** The $3,700-8,500 estimate for the full build assumes efficient GPU utilisation, minimal re-runs, and no major pipeline failures. In practice, first runs of complex multi-agent systems produce a lot of errors that require debugging, prompt iteration, and re-execution. Budget for 2-3x the estimated cost for the first successful end-to-end run.

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

## Phased Rollout

The first run does not require all 1M agents to hit live LLMs.

### Phase 1: Structure + Deep Coverage (~125,000 agents, live LLMs)

Run the full tree decomposition (levels 0-4, ~25,000 agents) with real LLMs, producing the complete structure and all level 4 entries. Then for levels 5-6, run a subset live (~100,000 leaf agents) covering the most important branches — those with the most published research and highest clinical relevance. Priority branches for phase 1:

- Cardiovascular system (top to bottom)
- Musculoskeletal system (top to bottom)
- Nervous system (top to bottom)
- Genetics/genomics (top to bottom)
- Common diseases (top to bottom)

These five domains likely account for the majority of published canine research and produce the most visually interesting content (cardiac anatomy diagrams, skeletal illustrations, neural pathway flowcharts, genetic pathway maps).

Remaining branches get placeholder entries: "This topic has been identified and placed in the knowledge tree but has not yet been researched by the agent system. It is queued for a future research pass." The tree has 1M nodes but ~125K have content.

### Phase 2: Background Fill (~200,000 agents/week)

After phase 1, the remaining branches fill in as background processing over subsequent weeks. Each week, the system activates agents for the next set of branches, prioritised by research volume and user interest. At 200K agents/week, full coverage takes 4-5 weeks after launch.

### Phase 3: Continuous Updates (ongoing, ~500-1,000 agents/week)

PubMed monitoring agents check for new publications. New papers trigger re-extraction, re-synthesis, and validation. Corrections submitted by users trigger re-analysis of affected nodes.

### Cost Implication of Phased Approach

Phase 1 with 125K live agents costs roughly 15-20% of the full 1M run. The demo still shows 1M agents expanding the tree (most of them doing the decomposition, paper fetching, entity extraction, and placeholder creation), but the expensive inference calls are concentrated on the branches that will have the best content.

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

### Phase 1: Structure + Priority Branches (~125K live LLM agents)

| Component | Duration | Cost |
|-----------|----------|------|
| 100 worker Pods (8-core VMs) × 48hrs | 4,800 VM-hours | $500-1,500 |
| 15-25 GPUs (inference + image) × 48hrs | 720-1,200 GPU-hours | $1,500-3,500 |
| Kafka brokers (15) × 48hrs | 720 VM-hours | $200-500 |
| Anthropic Opus (~250 calls) | — | ~$8 |
| PubMed API | — | Free |
| Persistent infrastructure | — | Existing |
| **Phase 1 Total** | | **$2,200-5,500** |

### Full 1M Run (all branches, over weeks)

| Component | Duration | Cost |
|-----------|----------|------|
| Phase 1 (above) | 48 hours | $2,200-5,500 |
| Phase 2 fill (200K agents/week × 4 weeks) | Background | $1,500-3,000 |
| **Full build total** | | **$3,700-8,500** |

### Ongoing Monthly (continuous updates)

| Component | Cost/month |
|-----------|-----------|
| Worker Pods (2-3 always-on for updates) | $100-300 |
| GPU inference (occasional re-synthesis) | $50-200 |
| Kafka + Postgres + Redis (persistent) | Already running |
| PubMed monitoring | Free |
| **Ongoing total** | **$150-500/month** |

## Output Format

Each agent produces a structured knowledge entry:

```json
{
  "node_id": "uuid",
  "parent_id": "uuid",
  "path": "cardiovascular.pathology.dcm.genetics.lmna",
  "title": "LMNA Gene Mutations in Canine Dilated Cardiomyopathy",
  "level": 6,
  "status": "researched | placeholder | under_review | disputed",
  "breed_specificity": "labrador_specific | general_canine | general_mammalian",
  "content": {
    "summary": "...",
    "key_findings": [...],
    "citations": [...],
    "confidence": "high | moderate | low",
    "evidence_base": "strong | moderate | limited | single_study",
    "discrepancies": [
      {
        "claim": "DCM incidence rate",
        "positions": [
          {"value": "10%", "source": "pmid:12345", "context": "UK referral population"},
          {"value": "4-6%", "source": "pmid:67890", "context": "North American consensus"}
        ],
        "explanation": "Likely reflects population selection bias"
      }
    ],
    "last_updated": "2026-03-02",
    "contradictions_flagged": [...]
  },
  "agent_provenance": {
    "agent_id": "uuid",
    "agent_type": "finding-synthesiser",
    "model_used": "biomistral-7b",
    "prompt_template": "template_id",
    "prompt_hash": "sha256",
    "input_data_hash": "sha256",
    "sources_consulted": ["pmid:12345", "pmid:67890"],
    "run_timestamp": "2026-03-15T14:30:00Z",
    "inference_time_ms": 4200,
    "rerunnable": true
  },
  "media": [
    {"type": "mermaid", "id": "uuid", "caption": "LMNA signaling pathway"},
    {"type": "image", "id": "uuid", "caption": "Nuclear envelope disruption"}
  ],
  "children": ["uuid", "uuid", ...],
  "cross_references": ["uuid", ...],
  "discussion_thread_id": "uuid",
  "version": 2,
  "version_history": ["uuid_v1", "uuid_v2"]
}
```

Total dataset: ~10GB (structured data + images). The `agent_provenance` block enables the re-run capability — given the same prompt template and input data, the agent can be replayed to regenerate the entry.

## Incremental Update Mechanism

Once built, a monitoring layer checks for new publications:

- Scheduled agent queries PubMed for new papers matching branch keywords
- New relevant papers trigger leaf-level re-extraction
- Changed findings propagate upward through re-synthesis
- Validators flag new contradictions

Veterinary research publishes ~50-100 relevant papers per week. Update workload: ~500-1,000 agent activations per week — a couple of worker Pods running continuously.

## Error Correction and Discussion Layer

The resource sits between a wiki and a knowledge graph. Entries are generated by agents but can be challenged by humans or by later validation passes.

### Correction Workflow

When someone flags an error — e.g. "this entry says Labradors have a 10% DCM incidence but the 2024 ACVIM consensus paper says 4-6%":

1. The correction goes into a review queue with the cited source.
2. A validator agent re-runs the analysis for that node, now including the correction source alongside the original sources.
3. If the correction checks out (the new source is credible and contradicts the existing entry), the entry is updated.
4. The parent synthesis re-runs to incorporate the change.
5. The original entry is versioned, not deleted — visitors can see what it said before, what the correction was, and what it says now.

### Handling Genuine Discrepancies

Where the literature itself disagrees, the entry surfaces both positions with evidence rather than picking one:

> "Smith et al. (2022) report 10% incidence in a UK population of 2,000 dogs; the ACVIM consensus (2024) cites 4-6% across North American referral populations. The difference may reflect population selection bias in referral vs primary care settings."

This is more useful than either number alone, and is the kind of nuance the synthesis agents must preserve rather than flatten. Synthesis prompts should explicitly instruct the model to present conflicting evidence as a discrepancy with possible explanations, not to pick the "most likely" value.

### Discussion Threads

Each node can have a discussion thread where users can:
- Flag potential errors with citations
- Suggest additional sources the agents missed
- Ask questions about methodology or interpretation
- Propose alternative decomposition of a sub-topic

Discussions that result in corrections trigger the correction workflow above. Discussions that suggest new sources trigger a targeted re-research pass. This creates a community layer that improves the resource over time without requiring manual editing of every entry.

### Versioning

Every entry stores its full history:

```json
{
  "versions": [
    {
      "version": 1,
      "date": "2026-03-15",
      "agent_id": "uuid",
      "prompt_hash": "sha256",
      "model": "biomistral-7b",
      "sources_consulted": ["pmid:12345", "pmid:67890"],
      "content": { "..." }
    },
    {
      "version": 2,
      "date": "2026-04-02",
      "trigger": "user_correction",
      "correction_source": "pmid:11111",
      "agent_id": "uuid",
      "prompt_hash": "sha256",
      "model": "biomistral-7b",
      "sources_consulted": ["pmid:12345", "pmid:67890", "pmid:11111"],
      "content": { "..." },
      "diff_summary": "DCM incidence updated from 10% to 4-6% based on ACVIM 2024 consensus"
    }
  ]
}
```

## Pathway and Mechanism Layer

### Concept

The biology tree describes what's known about canine biology, organised by body system. The pathway layer is a cross-cutting overlay that traces how drugs, treatments, and interventions act through the biology — which receptors they target, which tissues are affected, and what downstream consequences follow.

This is a graph laid on top of the hierarchy. Each drug has a pathway that touches multiple branches of the biology tree. At each point where it contacts a cell type, receptor, or signalling molecule, it links to everything else in the tree that uses the same component.

```
The biology tree (hierarchical):
  cardiovascular > physiology > signalling > prostaglandin synthesis
  musculoskeletal > pathology > inflammation > COX-2 mediated pain
  digestive > physiology > mucosal protection > prostaglandin-dependent
  renal > physiology > blood flow regulation > prostaglandin-dependent

The pathway graph (cross-cutting):
  Meloxicam (NSAID)
    → targets COX-2 (selective)
      → reduces prostaglandin E2 synthesis
        → in joints: reduces inflammation, pain [INTENDED]
        → in GI mucosa: reduces protective mucus [SIDE EFFECT]
        → in kidneys: reduces renal blood flow [SIDE EFFECT]
```

Each pathway node links back to the biology tree node describing that tissue or mechanism. Each biology tree node gains a list of pathways that affect it.

### Agent Types for Pathway Layer

**Drug profiler agents** (~1,000): One per drug/treatment. Build mechanism profiles from pharmacology databases (Plumb's, FARAD, published PK/PD studies). Extract: targets, binding affinities, downstream effects, species-specific dosing.

**Pathway tracer agents** (~5,000): Given a drug's primary target, trace through the biology tree to find all tissues where that target is expressed and what role it plays in each.

**Postulation agents** (~100,000): Given a drug-tissue pair where the drug's target is expressed, reason about the likely physiological consequence of inhibiting/activating that target in that tissue. Tag as known_side_effect (literature support exists) or postulated (mechanistic reasoning only).

**Interaction agents** (~20,000): Given two or more drugs, trace overlapping pathways and reason about combined effects.

**Behavioural pathway agents** (~5,000): Trace non-drug interventions (exercise, diet, training, environment) through their biological mechanisms.

**Pathway diagram agents** (~20,000): Generate flowcharts and mechanism diagrams for each pathway.

**Pathway synthesis agents** (~10,000): Combine per-drug results into coherent pathway summaries.

### Pathway Data Structure

```json
{
  "pathway_id": "uuid",
  "type": "drug | treatment | behavioural | dietary | environmental",
  "name": "Meloxicam mechanism of action in canine osteoarthritis",
  "agent": {
    "drug_name": "Meloxicam",
    "drug_class": "NSAID, preferential COX-2 inhibitor",
    "indication": "Osteoarthritis pain and inflammation",
    "canine_dose_range": "0.1 mg/kg q24h"
  },
  "tissue_effects": [
    {
      "tissue": "Synovial joint lining",
      "biology_tree_node": "uuid",
      "effect_type": "intended_therapeutic",
      "mechanism": "Reduced PGE2 → reduced inflammatory cascade",
      "evidence_base": "strong",
      "confidence": "high",
      "citations": ["pmid:12345"]
    },
    {
      "tissue": "Hepatic parenchyma",
      "biology_tree_node": "uuid",
      "effect_type": "postulated",
      "mechanism": "COX-2 expressed in hepatocytes; reduced PGE2 may affect blood flow",
      "evidence_base": "limited",
      "confidence": "low",
      "postulation_basis": "COX-2 expression in canine hepatocytes (pmid:45678); PGE2 role in hepatic flow from rodent models (pmid:56789); extrapolation speculative",
      "species_gap": "Rodent to canine extrapolation; no direct canine studies",
      "citations": ["pmid:45678", "pmid:56789"]
    }
  ],
  "interactions": [
    {
      "interacting_agent": "Furosemide",
      "mechanism": "Additive reduction of renal perfusion",
      "clinical_significance": "high"
    }
  ]
}
```

### Dependency on the Biology Tree

The pathway agents need the biology tree to exist first — at least the physiology and biochemistry branches covering receptor expression and signalling pathways. This makes the pathway layer a phase 2 activity:

1. Build biology tree physiology/biochemistry to level 5-6
2. Launch drug profiler agents (can run in parallel — they use pharmacology databases)
3. Launch pathway tracers and postulation agents (need both drug profiles and biology tree)
4. Launch interaction agents (need pathway results from multiple drugs)

### Agent Count

```
Drug profiler agents:         ~1,000
Pathway tracer agents:        ~5,000
Postulation agents:           ~100,000
Interaction agents:           ~20,000
Behavioural pathway agents:   ~5,000
Pathway synthesis agents:     ~10,000
Pathway diagram agents:       ~20,000
─────────────────────────────────────
Pathway layer total:          ~161,000 agents
```

Combined with the biology tree (~800K), this approaches 1M total.

### Serious Concerns and Limitations

This section must be honest about what's problematic. The pathway concept is appealing but has real weaknesses that need to be addressed or accepted.

**1. LLM pharmacological reasoning is unreliable.**

The postulation agents are doing pharmacological reasoning from first principles using a 7B language model. Pharmacologists spend years learning this and still get it wrong regularly. The model is pattern-matching on training data, not doing actual pharmacokinetic/pharmacodynamic modelling. It will produce plausible-sounding reasoning chains that are confidently wrong — and those are more dangerous than obviously wrong statements because they're harder to catch.

A reasoning chain like "COX-2 is expressed in tissue X → PGE2 has role Y in tissue X → therefore meloxicam inhibits role Y" looks rigorous but may be wrong at any step. The expression level might be too low to matter clinically. The PGE2 role might be compensated by other pathways. The drug concentration at therapeutic doses might not reach that tissue. A 7B model won't reliably account for these nuances.

Mitigation: Postulated effects must be very prominently flagged as mechanistic speculation, not clinical findings. The UI should visually distinguish postulated from evidence-based effects so strongly that a casual reader cannot mistake one for the other. Even so, the risk of generating plausible-but-wrong medical information is real.

**2. The species extrapolation problem.**

Most receptor expression and signalling pathway data comes from rodent or human studies. The biology tree will contain a lot of "established in rats, presumed similar in dogs." Extrapolating from rat hepatocyte COX-2 expression to canine clinical effects crosses two species boundaries and an in-vitro-to-clinical gap. The pathway agents need to tag every species extrapolation explicitly, and the confidence should drop significantly for cross-species inferences.

The canine-specific primary literature is much thinner than human or rodent. For many receptor-tissue combinations, there simply won't be canine data, and the postulation will rest entirely on cross-species inference. This should be stated clearly rather than hidden in a confidence score.

**3. Quantitative expression data is patchy.**

The postulation agents assume the biology tree contains receptor expression levels for each tissue. In practice, this data is often unavailable, especially canine-specific. Gene Expression Omnibus (GEO) and proteomics databases have reasonable coverage for human and mouse, much less for dog. Without quantitative expression data, the postulation can only say "this receptor is present in this tissue" but not "it's present at levels that would be clinically affected by therapeutic drug doses." That's a large gap.

**4. Behavioural pathways are largely speculative.**

Chains like "exercise → serotonin → reduced anxiety → reduced cortisol → improved immune function → reduced atopic dermatitis" sound coherent but each link has varying evidence strength. The cumulative uncertainty compounds — by the end of a 5-step chain the confidence should be very low. But a coherent narrative reads as more certain than it is. These pathways need aggressive uncertainty communication.

**5. Drug interaction combinatorics don't scale.**

300 commonly used drugs produce ~45,000 pairwise combinations. Triple combinations produce ~4.5 million. The 20,000 interaction agents won't cover this comprehensively. The system should focus on clinically common combinations (NSAIDs + corticosteroids, cardiac drugs + anaesthetics, etc.) and be explicit about what it hasn't assessed rather than implying completeness.

**6. Credibility problem with veterinary professionals.**

Existing resources — Plumb's Veterinary Drug Handbook, BSAVA Formulary, VIN (Veterinary Information Network) — have decades of editorial oversight, expert peer review, and clinical validation behind them. An AI-generated pharmacological resource starts at zero credibility with the veterinary community. The "postulated effects" section, however well-reasoned, will be met with justified scepticism by clinicians who know that mechanistic reasoning frequently doesn't predict clinical reality.

Getting from zero to trusted takes years and requires expert involvement — veterinary pharmacologists reviewing and endorsing entries. The system can generate draft content at scale, but without expert validation it's a research tool or hypothesis generator, not a clinical reference. The framing should be honest about this.

**7. Clinical liability.**

If a veterinarian reads a postulated effect and changes their prescribing based on it, and the postulation was wrong, there's a liability question. "The AI said meloxicam might affect liver function so I switched to a drug with its own risks" is a real scenario. The disclaimers need to be strong, and the system should probably frame postulations as "research hypotheses for expert evaluation" rather than "clinical guidance."

### What the Pathway Layer Is Actually Good For

Despite the above, there are genuine use cases where this adds value:

**Hypothesis generation for researchers.** "Here are 47 tissues where your drug's target is expressed, ranked by expression level and potential clinical significance, with the mechanistic reasoning for each." A pharmacologist can quickly scan this and say "yes, no, interesting, wrong, already known" much faster than doing the literature search themselves. The value is in the systematic enumeration, not in the individual postulations being correct.

**Education.** Veterinary students learning pharmacology benefit from seeing the mechanism of action traced through the whole body, even if the postulated effects are approximate. The visual pathway diagrams linking drug → receptor → tissue → effect are genuinely helpful for understanding.

**Drug safety signal detection.** If the system postulates an effect that later shows up in adverse event reports, that's a useful early signal. "We predicted this side effect from mechanism; now clinical reports are confirming it" is a legitimate contribution. But this requires tracking postulations against emerging clinical data over time.

**Identifying knowledge gaps.** "COX-2 is expressed in canine intestinal epithelium but we found zero studies examining NSAID effects on canine intestinal barrier function" is useful information for the research community even without a postulation about the consequences.

### Recommended Approach

Given the concerns above, the pathway layer should be framed as:

- A **research and hypothesis generation tool**, not a clinical reference
- Clearly separated from the biology tree's evidence-based content
- With postulated effects visually and structurally distinct from established findings
- With every species extrapolation explicitly tagged
- With aggressive uncertainty communication — confidence scores, evidence gaps, chain-of-reasoning exposure
- With an invitation for expert review rather than a claim of authority
- Initially covering only the ~200-300 most commonly used veterinary drugs, not attempting completeness

The pathway layer should probably launch after the biology tree has been reviewed by at least a few domain experts, so the foundation it builds on has been sanity-checked.

## UI/UX Considerations

### The Two Audiences

**Content consumers** (veterinarians, researchers, students): They want a beautiful, navigable reference. They care about the biology, not the agents. The surface should be a polished knowledge base with good typography, clear diagrams, and reliable citations.

**Framework buyers** (CTOs, engineering leads, potential customers): They want to see the machinery. How was this built? How does it work? Can I do this for my domain? The agent layer needs to be visible one click below the surface.

### Navigation Approaches for a 1M-Node Tree

A million nodes is not navigable as a traditional tree view. Two approaches that work for large knowledge structures:

**Zoomable map**: Start at the highest level showing ~20 major domains as tiles. Click cardiovascular and it expands to show its ~10 aspects. Keep drilling into specific conditions, mechanisms, molecular detail. Each zoom level shows the synthesis for that node plus children as clickable tiles. Similar to genome browsers — handles depth well.

**Search-first**: Type "labrador hip dysplasia" and jump directly to that node. Shows content, position in tree (breadcrumb: musculoskeletal > pathology > hip dysplasia), children, and sources. Tree structure exists as context but search is the primary navigation.

Likely both — search for targeted access, map for exploration and serendipity.

### Agent Transparency Panel

Each node has an "agent" tab or slide-out panel showing:

- Which agent produced this entry (type, ID, timestamp)
- The prompt that was used (viewable, editable for re-run)
- The model that generated it (biomistral-7b, opus, etc.)
- The sources consulted (linked to PubMed)
- The raw input data the agent received from its children
- Version history with diffs
- A "re-run" button that spawns a new agent with the same (or edited) prompt

When a user edits a prompt and hits re-run:
1. System spawns a new agent with the modified prompt
2. Agent runs, produces new output
3. New output shown alongside the old for comparison
4. User can accept (replaces entry, triggers upstream re-synthesis) or discard
5. If accepted, the change is versioned and attributed to the user

**Quality control concern**: This feature is good for the framework demo but creates a tension with the accuracy-first principle. If any user can re-run an agent with a modified prompt and accept the output into the resource, quality control is lost. Options: (a) re-runs produce a "proposed revision" that goes into a review queue rather than immediately replacing the entry, (b) re-runs are unrestricted but the output is shown as a user variant alongside the canonical entry, (c) re-run acceptance requires a certain trust level or expert role. For a clinical/scientific resource, option (a) is probably necessary. For a framework demo, option (b) shows the capability without the risk. This needs a decision.

This is the framework's selling point made tangible. Every piece of output is traceable, replayable, and improvable. Enterprises care about this — auditability, traceability, iterability.

### Discussion/Correction UI

Each node has a discussion tab:
- Flag errors with citations
- Suggest additional sources
- Ask questions about methodology
- Propose alternative decomposition
- See resolution history (which corrections were accepted, which were rejected and why)

### Status Indicators

Each node shows its status visually:
- **Researched**: Full content, sources, confidence level shown
- **Placeholder**: Topic identified, queued for research. Shows estimated research date.
- **Under review**: A correction has been submitted and is being re-analysed
- **Disputed**: The literature disagrees and both positions are presented
- **Recently updated**: New research incorporated in the last 30 days

### Demo Recording View

For the promotional recording, a dashboard showing:
- Real-time tree expansion — branches growing as agents decompose and populate
- Agent count climbing toward 1M
- Images appearing as diagram agents complete
- Message throughput (Kafka messages/sec)
- Inference utilisation (GPU cluster load)
- A ticker of recent completions: "Agent #847,291 completed: 'Troponin T mutations in Labrador cardiac muscle'"

### Front End Timing

The front end should not be designed in detail yet. Get the backend right first, produce real data for one complete branch (cardiovascular system top to bottom), then design the UI around actual content. Designing before data exists usually leads to redesigning afterward.

## Framework Selling Strategy

### What This Demo Proves

The dog biology project demonstrates capabilities that are hard to show with simpler examples:

1. **Hierarchical decomposition**: The system doesn't need a human to plan the tree. Give it a domain and it figures out the structure. This generalises to any knowledge domain — regulatory compliance, patent landscapes, API documentation, market research.

2. **Mixed model orchestration**: Different agents use different models based on the task. Opus for high-level reasoning, 7B for focused extraction, 3B for classification, specialised biomedical models for domain work, image models for illustrations. The framework handles routing transparently.

3. **Scale with auditability**: 1M agents, but every output is traceable to a specific prompt, model, and input. You can replay any piece. This is the opposite of "black box AI" — it's fully transparent AI at scale.

4. **Living output**: The knowledge base updates itself. New research is incorporated automatically. Errors are correctable. The system improves over time without manual intervention.

5. **Agent accessibility**: Non-technical users can inspect and modify agent behaviour through the UI. This is framework-as-product, not framework-as-infrastructure.

### Target Audiences for the Framework

**Pharmaceutical/biotech**: The same architecture maps directly to drug research literature review, clinical trial analysis, regulatory submission preparation. Replace "canine biology" with "oncology drug interactions" and the tree, agents, and pipeline are the same.

**Legal/compliance**: Decompose a regulatory framework (GDPR, SOX, HIPAA) into requirements, map requirements to business processes, identify gaps. Each agent analyses a section of regulation. The correction layer handles regulatory updates.

**Consulting/research firms**: Any domain where "gather everything known about X, structure it, keep it current" is valuable. Market research, competitive intelligence, technology landscaping.

**The pitch**: "We built a million-agent system that produced a world-class biological reference. Here's the reference — you can use it. Here's the framework — you can build the same thing for your domain. Here's every agent, every prompt, every source. Nothing is hidden."

### What the Demo Doesn't Prove

Honesty about limitations strengthens the pitch more than hiding them:

- **It doesn't prove the output is correct.** The demo proves the framework can produce structured, cited, well-formatted content at scale. It doesn't prove the content is free of errors. Any buyer who cares about accuracy (which is most of them) will ask "how do you validate this?" The answer needs to be "expert review, correction workflows, and continuous improvement" — not "the AI got it right."

- **It doesn't prove the framework works for any domain.** Canine biology has a well-structured corpus (PubMed, textbooks, established taxonomies). Other domains may not decompose as cleanly, may have messier source material, or may require different agent patterns. The demo shows one successful application, not universal applicability.

- **The 1M agent count is attention-grabbing but the per-agent sophistication varies enormously.** A paper fetcher that calls an API is not the same kind of "agent" as a postulation agent reasoning about drug mechanisms. Potential buyers will see through inflated numbers quickly. Better to lead with "200,000 agents doing substantive analytical work, supported by 800,000 agents doing data processing" than to imply all 1M are doing complex reasoning.

- **The ongoing cost model matters.** A potential customer's first question after "how much to build this for my domain" will be "what does it cost per month to keep it running." The $150-500/month ongoing cost for the biology project is low because veterinary publishing is slow. A fast-moving domain (regulatory changes, market data, news) would cost more. The pricing model needs to account for this.

## Open Questions

### Content and Quality

1. **Scope to reach 1M**: Core Labrador biology produces ~800K agents. What's the right extension — more breeds, veterinary practice, nutrition databases?

2. **3B model adequacy for relevance filtering**: Is classification genuinely safe at 3B, or should we benchmark against 7B on a test set before committing?

3. **BioMistral vs general 7B**: Need to evaluate BioMistral on actual veterinary paper abstracts. It's trained on human biomedical literature — canine-specific terminology coverage may vary.

4. **PubMed rate limiting**: 200K paper fetches at 10/sec = 5.5 hours. Should we pre-download and cache the canine corpus? There are services that provide bulk PubMed access.

5. **Full text access**: Many papers are behind paywalls. The pipeline works on abstracts, but full text would produce richer findings. Options: PubMed Central (open access subset), institutional access, or limit to open-access papers and note where full-text analysis would improve confidence.

6. **Image generation quality**: FLUX/SDXL for anatomical illustrations may need medical/anatomical LoRA fine-tunes. Quality of generated anatomical diagrams needs evaluation — might be better to use diagram code (SVG/Mermaid) for anything that needs precision.

7. **Synthesis prompt design for discrepancies**: The synthesis agents need explicit instruction to preserve conflicting evidence as presented discrepancies rather than flattening to a single "best" answer. This prompt engineering is important and needs iteration.

8. **Which branches first**: The phase 1 priority branches (cardiovascular, musculoskeletal, nervous, genetics, common diseases) are a starting proposal. Should this be driven by research volume, clinical relevance, or visual interest for the demo?

### Architecture and Implementation

9. **Worker pool implementation**: The pool_agent action and worker pool binary don't exist yet. This is the biggest code deliverable before the project can run at scale.

10. **Redis/Valkey for hot state**: The orchestration engine currently writes to Postgres on every step. The Redis caching layer needs implementation in the coordinator.

11. **Knowledge entry storage schema**: New tables for tree nodes, versions, discussion threads, correction queue. Needs design against the output format above.

12. **PubMed API integration**: New action or adapter for the paper fetching pipeline.

13. **SciSpacy NER integration**: Python sidecar in worker Pods, or a Go NER library, or a separate NER microservice. Needs design decision.

14. **Correction workflow implementation**: Queue, validator re-run, versioning, upstream re-synthesis propagation. This is a new pattern the chassis doesn't currently support.

### Front End and Demo

15. **UI/UX design**: How to present a 1M-node knowledge tree. Zoomable map vs search-first vs hybrid. Should wait until real data exists for one branch.

16. **Agent transparency panel**: Prompt viewing, editing, and re-run capability. How much of the agent internals to expose.

17. **Recording the demo**: How to capture the tree building in real-time. Dashboard design, metrics to show, narrative structure for the promotional video.

18. **Is this the right demo domain?** Dog biology is deep, emotionally engaging, and visually interesting. But the audience for canine biology is niche. Should the demo be in a domain with broader appeal (API documentation, patent analysis, regulatory compliance) or does the dog angle actually help because it's accessible and people connect with it?

### Commercial

19. **Target market for framework sales**: Technical buyers (want to see infrastructure) vs business buyers (want to see output). The demo needs to serve both.

20. **Pricing model**: Is the framework sold as infrastructure (deploy it yourself), as a service (we run it for your domain), or as finished knowledge bases (we ran it, here's the output)?

21. **Ongoing cost model**: The continuous update mechanism has ongoing infrastructure costs. How does this factor into pricing — subscription, per-update, or included?

## Architecture Dependencies

Items that need to exist before this project can run:

| Component | Status | Notes |
|-----------|--------|-------|
| Multi-cluster dispatch | Proven working | remote-job-spawner tested 2026-03-02 |
| dispatch_agent action | Code written, not committed | In dispatch_actions.go |
| Registry entry for dispatch_agent | Not done | One line in registry.go |
| pool_agent action | Not started | New action type — biggest code deliverable |
| Worker pool binary/mode | Not started | Consumes from shared topics, runs goroutines |
| Shared topic pool creation | Not started | Pre-create 64 topics × 16 partitions |
| Embedded 3B model in worker | Not started | llama.cpp or llama-cpp-go integration |
| SciSpacy NER in worker | Not started | Python sidecar or Go NER library |
| BioMistral 7B deployment | Not started | vLLM config for biomedical model |
| Redis/Valkey hot state layer | Not started | Coordinator change |
| Knowledge entry storage schema | Not started | Tree nodes, versions, provenance |
| Correction queue + workflow | Not started | New pattern for user-submitted corrections |
| Discussion thread storage | Not started | Per-node discussion threads |
| Version history storage | Not started | Entry versioning with diffs |
| Image generation pipeline | Not started | FLUX/SDXL inference server |
| PubMed API integration | Not started | New action or adapter |
| Incremental update scheduler | Not started | Kafka scheduler pattern exists |
| Re-run capability (agent replay) | Not started | Spawn agent from stored provenance |
| Front end (knowledge tree UI) | Not started | Wait for real data first |
| Front end (agent transparency) | Not started | Prompt viewing, editing, re-run |
| Front end (discussion/correction) | Not started | User-facing correction workflow |
| Demo dashboard (live metrics) | Not started | For promotional recording |

## Related Documents

- 006c: Multi-cluster dispatch handoff summary
- 013: Scaling analysis (10K through 1M)
- 001e: Development guide for new agents
- 002c: System architecture
- 010: Tool library guide
- 011: Kafka scheduler guide