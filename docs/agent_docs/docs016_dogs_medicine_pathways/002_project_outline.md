Let me think through the decomposition, agent types, LLM allocation, and practical architecture for this.

**The knowledge tree**

The domain is everything known about Canis lupus familiaris, with a Labrador Retriever as the reference breed. The tree decomposes naturally:

```
Level 0: Root (1 agent)
  "Comprehensive biology of the Labrador Retriever"

Level 1: Major domains (~20 agents)
  Body systems: musculoskeletal, cardiovascular, respiratory, digestive,
  nervous, endocrine, immune/lymphatic, integumentary, urinary, reproductive,
  sensory
  Cross-cutting: nutrition, behaviour, genetics/genomics, breed history,
  aging/geriatrics, exercise physiology, common diseases, veterinary 
  procedures, pharmacology

Level 2: Aspects per domain (~200 agents)
  e.g. Cardiovascular → anatomy, physiology, pathology, pharmacology,
  biochemistry, development/embryology, diagnostics, genetics

Level 3: Sub-topics (~2,000 agents)
  e.g. Cardiovascular > Pathology → dilated cardiomyopathy, valvular disease,
  heartworm, arrhythmias, congestive heart failure, pericardial effusion,
  aortic stenosis, patent ductus arteriosus...

Level 4: Specific topics (~20,000 agents)
  e.g. Dilated cardiomyopathy → genetic basis, molecular mechanism, clinical
  presentation, echocardiographic findings, diagnosis protocol, treatment
  protocol, prognosis, breed predisposition data, diet-associated DCM

Level 5: Detailed mechanisms (~200,000 agents)
  e.g. Genetic basis of DCM → LMNA mutations, titin (TTN) variants,
  phospholamban (PLN) mutations, troponin mutations, myosin binding protein C,
  genome-wide association studies, heritability estimates

Level 6: Molecular/atomic detail (~1,000,000 agents)
  e.g. LMNA mutations → specific variants identified in Labradors, lamin A/C
  protein structure, nuclear envelope disruption mechanism, downstream effects
  on cardiomyocyte integrity, comparison with human LMNA cardiomyopathy,
  relevant published papers with findings
```

The branching factor averages about 8-12 at each level. 20 × 10 × 10 × 10 × 10 × 5 gets you to 1M leaf agents. The tree wouldn't be uniform — some branches are deep and narrow (biochemical pathways have long chains), others are broad and shallow (breed history has many breeds to compare but not much depth per breed). The decomposition agents at levels 0-3 would decide how to split based on how much published knowledge exists for each sub-topic.

**Agent types and what each does**

There are roughly seven distinct agent roles in this system:

**1. Decomposer agents (Levels 0-3)**

These receive a topic and decide how to split it into subtopics. They need genuine reasoning about the domain — understanding that "cardiovascular pathology" should split by condition type, not alphabetically, and that some conditions warrant deeper decomposition than others. They consult the research literature to understand what's out there before deciding the tree structure.

Work: receive topic → search for overview papers → reason about how to decompose → spawn children for each subtopic → collect children's results → synthesise into a coherent entry for their level.

LLM: Anthropic Opus for levels 0-1 (these make the structural decisions that shape the entire tree). Shared 7B for levels 2-3 (the decomposition is narrower in scope).

**2. Research agents (Levels 4-5)**

These investigate a specific topic in depth. They search PubMed, veterinary journals, textbooks. They don't decompose further — they do the actual knowledge work.

Work: receive specific topic → search PubMed/Google Scholar → fetch and parse papers → extract key findings → structure into a knowledge entry with citations.

LLM: Shared 7B model for the paper analysis and summarisation. The 7B handles focused extraction well — "read this abstract about canine troponin T mutations and extract the key findings" is a narrow task.

**3. Leaf extractors (Level 6)**

The most numerous. These handle a very specific piece of knowledge — a single gene, a single chemical pathway, a single drug mechanism. They fetch one or two papers and extract structured data.

Work: receive very narrow query → fetch 1-3 papers → extract structured facts (gene name, variant, effect, evidence strength, species, breed if specified) → return structured JSON.

LLM: Embedded 3B model in the worker Pod. The task is simple enough that a small model handles it. These are the agents that make up the bulk of the 1M count.

**4. Paper fetcher agents (non-LLM)**

Pure data retrieval. Given a search query, they hit PubMed API, Google Scholar, or a pre-downloaded paper corpus, and return abstracts and full texts.

Work: receive search query → call PubMed E-utilities API → parse XML response → return list of papers with abstracts.

LLM: None. This is a web scraping / API call action. Fast, cheap, no inference needed.

**5. Synthesis agents (at each level, after children complete)**

Once a decomposer's children all return, the synthesis step combines their findings into a coherent entry. At higher levels this is a substantial writing task — weaving together the results from 10 children into a section that reads as a unified piece of knowledge.

Work: receive collected results from children → write a coherent section that integrates all findings → identify gaps or contradictions → flag for review if conflicting data found.

LLM: Anthropic Opus for levels 0-2 synthesis (these produce the top-level sections that need good writing quality). Shared 7B for levels 3-4. Embedded 3B for level 5 (just combining a few leaf results).

**6. Diagram/illustration agents**

These generate visual content. There are several sub-types:

- **Anatomical illustrations**: "Generate a labelled cross-section of the canine heart showing the four chambers, valves, and major vessels." → Image generation model (FLUX, Stable Diffusion XL with a medical/anatomical LoRA if available)
- **Biochemical pathway diagrams**: "Show the cyclic AMP signaling pathway in canine cardiomyocytes" → LLM generates Mermaid or SVG code, rendered to image. The 7B model is good at generating structured diagram code.
- **Flowcharts**: "Diagnostic algorithm for canine dilated cardiomyopathy" → LLM generates Mermaid flowchart code
- **Comparative diagrams**: "Skeletal comparison: Labrador vs Greyhound vs Dachshund" → Image generation model
- **Microscopy-style**: "Histological appearance of canine liver tissue" → Image generation model with appropriate style prompt

Work: receive description of needed visual → choose generation method (image model vs diagram code) → generate → return image URL or SVG.

LLM: Mixed. Some use image generation models directly. Others use the 7B to write Mermaid/SVG code which is then rendered (no GPU needed for rendering). Flowcharts and pathway diagrams are better as Mermaid because they're editable, searchable, and precise. Anatomical illustrations need actual image generation.

**7. Validator/cross-reference agents**

These check consistency between sibling branches. If the cardiovascular pathology branch says "Labradors have a 10% incidence of DCM" and the genetics branch says "5%", a validator flags the contradiction.

Work: receive two or more entries → compare factual claims → identify contradictions → return list of conflicts with citations from each source.

LLM: Shared 7B. This is a comparison and extraction task, not creative writing.

**LLM allocation summary**

```
Agent type          Count        LLM              Where it runs
─────────────────────────────────────────────────────────────────
Decomposer L0-1     ~20         Anthropic Opus    API call
Decomposer L2-3     ~2,200      Shared 7B         Inference server
Synthesis L0-2      ~220        Anthropic Opus    API call  
Synthesis L3-4      ~22,000     Shared 7B         Inference server
Research L4-5       ~220,000    Shared 7B         Inference server
Leaf extractor L6   ~750,000    Embedded 3B       Worker pod (local)
Paper fetcher       ~200,000    None              Worker pod (code only)
Diagram (image)     ~10,000     FLUX/SDXL         GPU inference server
Diagram (mermaid)   ~30,000     Shared 7B         Inference server
Validator           ~50,000     Shared 7B         Inference server
─────────────────────────────────────────────────────────────────
Total              ~1,284,000
```

The Opus calls are small in number — maybe 250 total. At ~$15/M output tokens with average 2K output tokens each, that's about $7.50 total. Negligible.

The shared 7B handles ~325,000 agents. With average 1.5 calls per agent (some agents make multiple passes), that's ~500K inference calls. On 10 GPUs running vLLM, each doing ~100 calls/minute with continuous batching, that's 60K calls/hour, so ~8 hours for the 7B workload.

The embedded 3B handles ~750,000 agents. Each makes 1-2 short calls (extracting from a paper abstract). With 100 worker Pods each running a local 3B model doing ~6 calls/minute on CPU, that's 600 calls/minute across the fleet, so ~2,500 hours of serial inference spread across 100 Pods = ~25 hours.

The paper fetchers are fast — no LLM, just API calls. Limited by PubMed's rate limiting (3 requests/second without API key, 10/second with). With 200K paper fetch agents and 10 req/second, that's ~5.5 hours for the fetching alone. You could cache aggressively — many agents will reference the same papers.

Image generation: 10,000 images on FLUX. A single GPU generates an image in 5-15 seconds. With 5 GPUs dedicated to image generation, that's ~60-200 images/minute, so 50-170 minutes. This runs in parallel with everything else.

**Total timeline estimate for the full run:**

The tree builds depth-first. Level 0-1 agents decompose first (Opus, minutes), then level 2-3 (7B, an hour), then the massive level 4-6 expansion. The bottleneck is the leaf extraction phase — 750K agents running through 100 worker Pods takes ~25 hours. But it's not purely serial — as early branches complete, their synthesis agents start working while later branches are still extracting.

Realistic total: 30-48 hours for the full tree, with most of the time spent in the leaf extraction phase.

**Data storage and the knowledge graph**

Each agent produces a structured entry:

```json
{
  "node_id": "uuid",
  "parent_id": "uuid",
  "path": "cardiovascular.pathology.dcm.genetics.lmna",
  "title": "LMNA Gene Mutations in Canine Dilated Cardiomyopathy",
  "level": 6,
  "content": {
    "summary": "...",
    "key_findings": [...],
    "citations": [...],
    "confidence": "moderate",
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

1M of these entries at ~2-5KB each = 2-5GB of structured knowledge data. Plus ~10K images at ~500KB each = 5GB of media. Total dataset: ~10GB. Easily stored in Postgres or YugabyteDB. The tree structure with `parent_id` and `path` enables both hierarchical navigation (drill down from root) and search (find all entries mentioning "troponin").

**The incremental update mechanism**

This is where it gets interesting long-term. Once the tree exists, you set up a monitoring layer:

- A scheduled agent checks PubMed for new papers matching key terms from each branch
- When a relevant new paper appears, it spawns a leaf agent to extract findings
- The leaf's parent re-runs its synthesis with the new data included
- If the new data contradicts existing entries, a validator flags it
- Changes propagate up the tree — a new finding at level 6 triggers re-synthesis at levels 5, 4, 3

This is a trickle, not a flood. Veterinary research publishes maybe 50-100 relevant papers per week across all of canine biology. Each paper might touch 5-10 nodes in the tree. So the update workload is maybe 500-1000 agent activations per week — easily handled by a couple of worker Pods running continuously.

**Infrastructure for the full run**

```
Primary cluster (Rackspace):
  Kafka: 15 brokers (handles message throughput)
  Postgres: 1 instance with Redis/Valkey for hot state
  agent-chassis: 3 pods (handles top-level orchestration)
  PubMed cache: Redis instance holding fetched abstracts

Worker cluster(s) (2-3 clusters, Rackspace or cloud):
  Worker pool pods: 100 total (each with embedded 3B model, 8 cores, 8GB)
  remote-job-spawner: 1 per cluster

Inference cluster:
  7B inference (vLLM): 10 GPUs (A100 or equivalent)
  Image generation (FLUX): 5 GPUs
  
External APIs:
  Anthropic Opus: ~250 calls over the run
  PubMed E-utilities: sustained 10 req/sec
```

**Estimated cost for the full run**:

| Component | Duration | Cost |
|-----------|----------|------|
| 100 worker Pods (8-core VMs) × 48hrs | 4,800 VM-hours | $500-1,500 |
| 15 GPUs (10 inference + 5 image) × 48hrs | 720 GPU-hours | $1,500-3,000 |
| Kafka brokers (15) × 48hrs | 720 VM-hours | $200-500 |
| Anthropic Opus (~250 calls) | - | ~$8 |
| PubMed API | - | Free |
| Persistent infrastructure (already running) | - | Existing |
| **Total** | | **$2,200-5,000** |

**The UI/UX question you raised**

The output is a navigable knowledge tree. The most natural presentation is a web UI where you start at the root ("The Labrador Retriever") and click to drill down. Each node shows its synthesised content, citations, diagrams, and links to children. A search bar lets you jump to any node. A "recent updates" panel shows which branches have been refreshed with new research.

The tree is also the data — it could be exported as a structured JSON dataset, a set of markdown files, or even a printable reference. The mermaid diagrams render in any markdown viewer. The images are standard PNGs.

For the demo recording specifically, the compelling visual is the tree expanding in real-time: starting from the root, watching branches grow as agents decompose and populate, seeing images appear as diagram agents complete. A dashboard showing the agent count climbing toward 1M while the knowledge tree fills in would be strong promotional material.

Want me to start on the agent definitions for this — the decomposer, researcher, leaf extractor, and diagram agents — or keep working through architecture questions first?

