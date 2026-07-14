
<!-- SOURCE: U21_legacy_docs_b.md -->
### Canine biology knowledge tree (1M-agent demo)
- **category:** canine-biology
- **status-signal:** aspirational
- **status-evidence:** docs016/003c dated 2026-03-02, "Status: Working draft for further iteration"; docs016/004 (2026-03-03) demotes it: "best treated as marketing spend, not as a product... Build one branch (cardiovascular) as a polished showcase."
- **what:** A hierarchical agent swarm building a citable Labrador-reference knowledge tree: 7 levels of decomposition (root → body systems → aspects → subtopics → specific topics → mechanisms → molecular detail, branching 8–12), ~800K–1M agents across nine roles (Opus decomposers/synthesisers at top levels; BioMistral 7B research and finding-synthesis; non-LLM paper fetchers hitting PubMed; SciSpacy NER entity extractors; embedded-3B relevance filters; mermaid/FLUX diagram agents; 7B validators flagging cross-branch contradictions). Design priorities: accuracy over completeness; no reader-visible text from 3B models; phased rollout (125K live agents on five priority branches, background fill, then continuous PubMed-monitoring updates ~500-1000 agents/week); every node auditable (agent, prompt, sources, model); correction/discussion layer with versioning; pathway/mechanism cross-layer. Honest-risk section: credibility vs Plumb's/Merck, theatrical agent count, hallucination persistence, front-end decisive, costs 2-3x estimates ($2.2K-8.5K full run).
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md; docs016_dogs_medicine_pathways/002_project_outline.md; docs016_dogs_medicine_pathways/004_medical_business_reality_assessment.md
- **relations:** canine-biology category (docs018 feature plans); multicluster worker pools; model-tiering; business strategy demotion.
- **verify-later:** any decomposer/leaf agent definitions; knowledge tree tables (expected absent).

<!-- SOURCE: U22_recent_small_docs.md -->
### Canine biology knowledge base (veterinary seeding)
- **category:** canine-biology
- **status-signal:** aspirational
- **status-evidence:** "The canine biology project stops being aspirational and becomes the working proof..." — future tense; "knowledge base is empty" in the RAG explainer.
- **what:** The first real RAG content and proof-of-concept for the veterinary vertical: structured LLM extraction (breed health profiles for top 20 UK breeds, 30-40 procedures, top 30 conditions, nutrition/vaccination/behaviour) into ~300-500 self-contained 200-500-word chunks, validated (self-consistency, cross-reference, structural), embedded via Ollama, and indexed into `collection: "veterinary"`. Structured JSON with confidence markers, not prose.
- **sources:** docs023.../018_canine_biology.md, docs023.../001_canine_biology_grok_plan.md
- **relations:** RAG knowledge_base, vertical knowledge architecture, text LoRA (vet extractor), deep research domain authority
- **verify-later:** knowledge_base rows collection='veterinary'; knowledge-extractor agent

<!-- SOURCE: U22_recent_small_docs.md -->
### Interactive Biological Explorer + experiment engine (aspirational vision)
- **category:** canine-biology
- **status-signal:** abandoned
- **status-evidence:** The grandiose Grok "Final Consolidated Plan" (multi-scale explorer, knowledge graph, experiment engine, 14-week timeline) is explicitly downgraded in the later doc: "The original 1M-agent design was aspirational. This plan is practical."
- **what:** An early, much larger vision: a public Next.js/Three.js/Cytoscape web app allowing drill-down from a pseudo-photographic Labrador image → organ systems → cells → biochemical pathways → genes, backed by a PostgreSQL/Neo4j knowledge graph (Gene/Protein/Metabolite/Reaction/Organ nodes), plus an agent-driven "theoretical experiment engine" running SciPy ODE simulations. Superseded by the practical RAG-seeding plan; the explorer/graph/experiment layers were dropped.
- **sources:** docs023.../001_canine_biology_grok_plan.md, docs023.../018_canine_biology.md#1
- **relations:** canine biology knowledge base (the practical replacement), image LoRA
- **verify-later:** n/a (not built; abandoned scope)

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Canine-biology per-vertical knowledge + LoRA project
- **category:** canine-biology
- **status-signal:** partial
- **status-evidence:** 018_canine_biology (identical to live docs023): §5 RAG in content generation "Authoritative Domain Knowledge"; §6 Text LoRA (4-bit QLoRA→GGUF Q4_K_M→Ollama), §7 Image LoRA (SDXL, ~£35-95 first pass); FOCUS_imagery_assessment §8 "Status: planned, not started."
- **what:** The reference per-vertical knowledge/fine-tuning project: research → chunk+embed into the RAG knowledge base → text LoRA fine-tune (Unsloth QLoRA, deployed via Ollama Modelfile) → image LoRA (60-90 curated images, SDXL/PixArt) for consistent per-vertical visual style. The image LoRA presupposes an adapter that accepts a `model:"vet-diagram-v1"` field — which the current Stability-only adapter cannot, so it is blocked on the provider-router work.
- **sources:** 018_canine_biology.md#6-text-lora-fine-tuning, #7-image-lora-fine-tuning; FOCUS_imagery_assessment(1).md#8-per-vertical-training-infrastructure
- **relations:** quality flywheel; RAG best practices; image provider router; vision auditor
- **verify-later:** knowledge_base pgvector rows; training_runs; Ollama LoRA Modelfiles

<!-- SOURCE: U21_legacy_docs_b.md -->
### Canine biology knowledge tree (1M-agent demo)
- **category:** canine-biology
- **status-signal:** aspirational
- **status-evidence:** docs016/003c dated 2026-03-02, "Status: Working draft for further iteration"; docs016/004 (2026-03-03) demotes it: "best treated as marketing spend, not as a product... Build one branch (cardiovascular) as a polished showcase."
- **what:** A hierarchical agent swarm building a citable Labrador-reference knowledge tree: 7 levels of decomposition (root → body systems → aspects → subtopics → specific topics → mechanisms → molecular detail, branching 8–12), ~800K–1M agents across nine roles (Opus decomposers/synthesisers at top levels; BioMistral 7B research and finding-synthesis; non-LLM paper fetchers hitting PubMed; SciSpacy NER entity extractors; embedded-3B relevance filters; mermaid/FLUX diagram agents; 7B validators flagging cross-branch contradictions). Design priorities: accuracy over completeness; no reader-visible text from 3B models; phased rollout (125K live agents on five priority branches, background fill, then continuous PubMed-monitoring updates ~500-1000 agents/week); every node auditable (agent, prompt, sources, model); correction/discussion layer with versioning; pathway/mechanism cross-layer. Honest-risk section: credibility vs Plumb's/Merck, theatrical agent count, hallucination persistence, front-end decisive, costs 2-3x estimates ($2.2K-8.5K full run).
- **sources:** docs016_dogs_medicine_pathways/003c_canine_biology_project_baseline_v3.md; docs016_dogs_medicine_pathways/002_project_outline.md; docs016_dogs_medicine_pathways/004_medical_business_reality_assessment.md
- **relations:** canine-biology category (docs018 feature plans); multicluster worker pools; model-tiering; business strategy demotion.
- **verify-later:** any decomposer/leaf agent definitions; knowledge tree tables (expected absent).

<!-- SOURCE: U22_recent_small_docs.md -->
### Canine biology knowledge base (veterinary seeding)
- **category:** canine-biology
- **status-signal:** aspirational
- **status-evidence:** "The canine biology project stops being aspirational and becomes the working proof..." — future tense; "knowledge base is empty" in the RAG explainer.
- **what:** The first real RAG content and proof-of-concept for the veterinary vertical: structured LLM extraction (breed health profiles for top 20 UK breeds, 30-40 procedures, top 30 conditions, nutrition/vaccination/behaviour) into ~300-500 self-contained 200-500-word chunks, validated (self-consistency, cross-reference, structural), embedded via Ollama, and indexed into `collection: "veterinary"`. Structured JSON with confidence markers, not prose.
- **sources:** docs023.../018_canine_biology.md, docs023.../001_canine_biology_grok_plan.md
- **relations:** RAG knowledge_base, vertical knowledge architecture, text LoRA (vet extractor), deep research domain authority
- **verify-later:** knowledge_base rows collection='veterinary'; knowledge-extractor agent

<!-- SOURCE: U22_recent_small_docs.md -->
### Interactive Biological Explorer + experiment engine (aspirational vision)
- **category:** canine-biology
- **status-signal:** abandoned
- **status-evidence:** The grandiose Grok "Final Consolidated Plan" (multi-scale explorer, knowledge graph, experiment engine, 14-week timeline) is explicitly downgraded in the later doc: "The original 1M-agent design was aspirational. This plan is practical."
- **what:** An early, much larger vision: a public Next.js/Three.js/Cytoscape web app allowing drill-down from a pseudo-photographic Labrador image → organ systems → cells → biochemical pathways → genes, backed by a PostgreSQL/Neo4j knowledge graph (Gene/Protein/Metabolite/Reaction/Organ nodes), plus an agent-driven "theoretical experiment engine" running SciPy ODE simulations. Superseded by the practical RAG-seeding plan; the explorer/graph/experiment layers were dropped.
- **sources:** docs023.../001_canine_biology_grok_plan.md, docs023.../018_canine_biology.md#1
- **relations:** canine biology knowledge base (the practical replacement), image LoRA
- **verify-later:** n/a (not built; abandoned scope)

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Canine-biology per-vertical knowledge + LoRA project
- **category:** canine-biology
- **status-signal:** partial
- **status-evidence:** 018_canine_biology (identical to live docs023): §5 RAG in content generation "Authoritative Domain Knowledge"; §6 Text LoRA (4-bit QLoRA→GGUF Q4_K_M→Ollama), §7 Image LoRA (SDXL, ~£35-95 first pass); FOCUS_imagery_assessment §8 "Status: planned, not started."
- **what:** The reference per-vertical knowledge/fine-tuning project: research → chunk+embed into the RAG knowledge base → text LoRA fine-tune (Unsloth QLoRA, deployed via Ollama Modelfile) → image LoRA (60-90 curated images, SDXL/PixArt) for consistent per-vertical visual style. The image LoRA presupposes an adapter that accepts a `model:"vet-diagram-v1"` field — which the current Stability-only adapter cannot, so it is blocked on the provider-router work.
- **sources:** 018_canine_biology.md#6-text-lora-fine-tuning, #7-image-lora-fine-tuning; FOCUS_imagery_assessment(1).md#8-per-vertical-training-infrastructure
- **relations:** quality flywheel; RAG best practices; image provider router; vision auditor
- **verify-later:** knowledge_base pgvector rows; training_runs; Ollama LoRA Modelfiles
