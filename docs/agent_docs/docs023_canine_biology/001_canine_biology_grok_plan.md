Here is the complete, polished, and self-contained final plan exported as a clean Markdown document. You can copy and paste it directly into a `.md` file.

```markdown
# 018 — Canine Biology: Comprehensive Multi-Scale Model and Interactive Explorer  
**Labrador Retriever Exemplar**

**Version:** Final Consolidated Plan  
**Date:** April 2026  
**Goal:** Build the most complete publicly accessible digital model of canine biology, using the Labrador Retriever as the primary exemplar. Deliver a practical veterinary knowledge base for content sites while creating a deeply navigable, visually rich interactive Biological Explorer that allows users and agents to drill from a pseudo-photographic Labrador image down to individual biochemical reactions, neurological pathways, genetic variants, and fluid flows. Support agent-driven theoretical experiments on the model.

---

## 1. Project Vision and Deliverables

This project integrates and significantly extends the original practical veterinary RAG pipeline into a full multi-scale biological model.

**Five Core Deliverables:**

1. **Populated Knowledge Bases**  
   - `veterinary` collection: 300–500 high-value chunks for vet content sites.  
   - `biology_model` collection: 2,000–3,000 structured entries covering anatomy, pathways, biochemistry, genetics, and fluid dynamics.

2. **Fine-Tuned LoRAs**  
   - Text LoRA: Veterinary extractor + multi-scale biology simulator.  
   - Enhanced Image LoRA: Consistent scientific, infographic, and pseudo-photographic canine biology illustrations.

3. **Interactive Biological Explorer Website**  
   - Public-facing web application.  
   - Start with a high-resolution pseudo-photographic Labrador image.  
   - Clickable drill-down through layers: whole animal → organ systems → tissues → cells → biochemical pathways → individual molecules → genes.  
   - Every element supported by structured RAG data, knowledge graph links, and on-demand visuals.

4. **Theoretical Experiment Engine**  
   - Agents translate natural-language experiment requests into graph queries and lightweight simulations.  
   - Return structured predictions with confidence scores and source traceability.

5. **Reusable Pipeline**  
   - Proven research → extraction → validation → indexing → visualisation → simulation workflow that can be applied to other species or verticals.

All outputs are structured, confidence-tagged, source-referenced, and designed for both human browsing and agent consumption.

---

## 2. Scope and Priorities

### Priority 1–3 (Weeks 1–3): Veterinary Foundation (Original Scope)
- Breed health profiles (top 20 UK breeds, with deep Labrador focus)  
- Common procedures, costs, and recovery  
- Top 30 conditions and diseases  
- Nutrition, vaccination, preventive care, and behaviour basics  
- Output: ~400 structured chunks ready for RAG and content generation

### Priority 4 (Weeks 4–8): Deep Multi-Scale Labrador Biology
| Topic Area                          | Labrador Relevance                          | Key Outputs |
|-------------------------------------|---------------------------------------------|-----------|
| Organ-system anatomy (cellular level) | Base navigation layer                      | Layered cross-sections for 11 major systems |
| Biochemical & metabolic pathways    | Obesity, exercise-induced collapse         | 80+ pathways with canine-specific notes |
| Neurological & endocrine cascades   | Behaviour, hunger signalling, pain         | Synaptic, hormonal, and ion-flow maps |
| Genomics & breed-specific variants  | Well-studied breed                         | Catalogue of key Labrador variants (DNM1/EIC, PTPLA/CNM, prcd/PRA, DENND1B/obesity, etc.) |
| Fluid dynamics & transport pathways | Liquid flow requirement                    | Blood, lymph, CSF, bile, urine, synovial fluid pathways with transporters |
| Integrative multi-scale models      | Enables theoretical experiments            | Knowledge graph linking gene → protein → reaction → organ → phenotype |

### Priority 5 (Weeks 9–14): Explorer and Simulation Layer
- Fully interactive web UI with layered visuals and graph navigation  
- Agent-driven experiment engine supporting realistic “what-if” queries

---

## 3. Research Phase

### 3.1 Structured Extraction at Scale
Use parallel “Biology Researcher” agents (50–100 concurrently) with tightly specified JSON prompts.

**Example Pathway Extraction Prompt:**
```json
You are a canine molecular biologist. Extract a structured model for the {pathway_name} in Labrador Retrievers.

Return ONLY valid JSON with these fields:
- pathway_id, name, external_id (KEGG/Reactome)
- compartments
- enzymes: [{gene_symbol, canine_ortholog, EC_number, labrador_notes}]
- metabolites: [{name, formula, role, concentration_range_if_known}]
- reactions: [{step, reactants, products, enzyme, reversible, regulation}]
- labrador_specific_notes and phenotype_links
- visual_caption (detailed prompt for image LoRA)
- simulation_defaults (Km, Vmax where supported)
Sources: KEGG cfa, Reactome dog pathways, Dog10K, AMDB, 2025–2026 Labrador studies. Use [UNCERTAIN] for low-confidence data.
```

Similar structured prompts for genetics, anatomy layers, and fluid pathways.

### 3.2 Literature and Data Augmentation
- Dedicated Literature Miner agent scrapes and extracts from: Dog10K, KEGG *Canis lupus familiaris*, Reactome, AMDB metabolomics database, PubMed (canine + Labrador filters), BSAVA manuals.
- Cross-reference all claims.

### 3.3 Validation
- Triple extraction with majority vote for critical entries.
- Model Integrator agent builds temporary knowledge graph and flags inconsistencies.
- Every entry receives a confidence score (1–5) and source references.

---

## 4. Indexing and Storage

- **Collections**:
    - `veterinary` (original)
    - `biology_model` (structured JSON + nomic-embed-text embeddings)
    - `visual_assets` (LoRA generation prompts + metadata)

- **Knowledge Graph** (PostgreSQL with pg_graphql or dedicated Neo4j):
    - Nodes: Gene, Protein, Metabolite, Reaction, Pathway, Organ, CellType, FluidCompartment, Phenotype
    - Edges: encodes, catalyzes, transports, regulates, located_in, linked_to_phenotype, etc.

- Chunking rules: Self-contained 200–500 word chunks tagged with node IDs for direct explorer linking.

---

## 5. Interactive Biological Explorer

**Technical Approach:**
- Frontend: Next.js + Tailwind + Three.js (for pseudo-3D layered model) + Cytoscape.js (pathway and network visualisation)
- Navigation: Clickable hotspots on pseudo-photographic image → SVG overlays → dynamic loading of RAG text, knowledge graph sub-network, and on-demand LoRA-generated visuals
- “Run Experiment” button triggers agent workflow

**Experiment Engine Workflow:**
1. Natural-language query parsed into structured simulation spec
2. Relevant sub-graph retrieved
3. Simulation code generated (SciPy ODE solver, simple kinetic models, or flux analysis)
4. Code executed safely in agent environment
5. Results returned with confidence, limitations, and sources

---

## 6. LoRA Fine-Tuning

### Text LoRA
- Veterinary Knowledge Extractor (original)
- Multi-Scale Biology Simulator (new – outputs simulation-ready JSON)

### Image LoRA (Critical Deliverable)
- Training set: 200 carefully curated images (pseudo-photographic cross-sections, clean pathway diagrams, fluid flow animations, genetic maps, layered anatomy)
- Style trigger phrase: “veterinary scientific multi-layer canine biology illustration, Labrador Retriever, clean labels, pseudo-photographic where appropriate, consistent colour coding, professional medical style”
- Base model: SDXL or FLUX.2 Dev with LoRA
- Integration: Explorer requests visuals by node ID using pre-stored captions

---

## 7. Implementation Timeline (14 Weeks)

**Phase A–C (Weeks 1–3)**  
Veterinary foundation, RAG population, initial content testing.

**Phase D (Weeks 4–8)**  
Deep biology research, structured extraction, knowledge graph construction.

**Phase E (Weeks 9–11)**  
Image LoRA training, bulk visual asset generation.

**Phase F (Weeks 12–14)**  
Interactive Explorer website build, experiment engine implementation, end-to-end testing.

**Phase G (Week 15+)**  
Expert review, scaling to additional breeds/species, pipeline documentation.

---

## 8. Costs and Resources (April 2026 Estimate)

| Activity                          | Estimated Cost | Notes |
|-----------------------------------|----------------|-------|
| LLM extractions (Claude API)      | £180–320      | ~2,500 structured calls |
| GPU rental (LoRA training)        | £50–100       | 6–10 hours total |
| Graph DB and simulation compute   | £15–30        | Ongoing costs near zero |
| **Total Project**                 | **£250–450**  | Extremely cost-effective for scope |

**Risks and Mitigations:**
- Canine-specific data gaps → explicit confidence tagging and ortholog fallbacks
- Visual style consistency → strict LoRA triggers + quality spot-checks
- Simulation accuracy → restrict to well-characterised pathways with clear limitations displayed

---

## 9. Success Criteria

- Veterinary RAG improves content quality on test sites (specific figures, authority tone).
- Biological Explorer website is live and publicly accessible, supporting seamless drill-down from whole-animal image to molecular level.
- Image LoRA produces consistent, professional visuals across all layers.
- Experiment engine successfully runs and explains at least 10 meaningful Labrador-specific theoretical scenarios.
- The entire pipeline is documented and reusable for future biological or scientific verticals.

This plan provides a complete, executable roadmap that respects the original practical veterinary goals while fully realising the deep biological, visual, and experimental vision requested.

---

**End of Document**
```

You can now save this as `018-canine-biology-final-plan.md` and use it directly in your project documentation or agent workflows. Let me know if you need any section expanded, agent prompt libraries added, or a version with more technical implementation details.