# Register — flows-and-narrative

3 concepts, consolidated from 6 raw extractions (3 unique blocks, each present
twice in the source cluster file due to mechanical duplication in the input),
all from unit U21.

### FLW-001 — Multi-track flows (journeys, narrative arcs, layered context)
- **status:** abandoned
- **status-evidence:** Full schema (site_flows, flow_pages, page_transitions, site_brand_dna) in docs010/002; docs010/005 "Configuration: Single-flow (production)... build for complexity, configure for simplicity"; docs012/007 MVP migration re-lists site_flows as still-to-create; no later doc shows flows populated.
- **what:** Model a site as choreographed audience journeys rather than a flat page list: each flow has an audience segment, entry points, a narrative arc of stages with per-stage voice parameters, and ordered pages with context_overrides; context inherits hierarchically SITE (immutable brand DNA) → FLOW (narrative) → PAGE (objective/overrides) → COMPONENT (paragraph tactics); navigation becomes flow-aware (different next-step CTAs per track); shared pages get per-flow variants; page_transitions support A/B weighting. "Stop thinking pages → start thinking journeys."
- **sources:** docs009_site_interrogation_and_solutions/004_multitrack_sitemap_architecture_different_flows.md; docs010_multitrack_flows_persona_architecture/002_multi_track_schema.sql; docs010_multitrack_flows_persona_architecture/005_implementation_summary.md
- **relations:** brand DNA invariants; voice parameters; persona assignment per stage (persona-architecture register); pattern library flow-stage tagging; conceptual ancestor of content strategy in site plans
- **verify-later:** site_flows/flow_pages/page_transitions/site_brand_dna tables — created? populated?

### FLW-002 — Brand DNA invariants with bounded variance
- **status:** abandoned
- **status-evidence:** docs009/004 "brand_dna.invariants: core_message, forbidden_phrases, required_elements; variance_allowed: voice_formality [0.4,1.0]"; site_brand_dna table in docs010/002; later brand data lives in sites.content_data.brand_spec instead (docs017/019b).
- **what:** A site-level immutable identity layer — core message, values, visual system, forbidden phrases, required elements — plus explicit allowed ranges for voice variance, enforced by an evaluator check before content is accepted (vocabulary, contradiction, variance bounds, visual consistency). Solves coherence-vs-variation across multiple flows/voices.
- **sources:** docs009_site_interrogation_and_solutions/004_multitrack_sitemap_architecture_different_flows.md#Q3; docs010_multitrack_flows_persona_architecture/002_multi_track_schema.sql#BRAND-DNA
- **relations:** brand_spec in sites.content_data (descendant); content-reviewer coherence checks (hitl register); design-composition brand decisions
- **verify-later:** site_brand_dna table vs sites.content_data.brand_spec usage

### FLW-003 — Voice parameters (numeric stage-tuned voice)
- **status:** abandoned
- **status-evidence:** docs010/019 Week 2 plan (get_voice_for_page SQL, formality 0.5 home / 0.7 elsewhere); docs010/007: "Instead of trying to tune voice parameters numerically (formality 0.7 → 0.8), we select the right copywriter persona."
- **stage2-verified (2026-07-14):** superseded → abandoned — Old mechanism (get_voice_for_page, numeric voice dials) only in docs010/019 planning docs, zero hits in platform/ code. Claimed replacement PERS-001 (copywriter persona roster) is itself classified 'abandoned' in register/persona-architecture.md:8 ('no later builder references personas'; personas/specialized_agents/...
- **what:** Continuous voice dials (formality, technical_depth, sales_pressure, urgency, data_density, emotional_appeal 0–1) attached to flow stages and page context_overrides, injected into content prompts so voice progresses through the journey (awareness casual → conversion formal). Explicitly superseded within its own directory by persona selection, which embodies the parameters naturally.
- **sources:** docs010_multitrack_flows_persona_architecture/019_start_here_document.md#Week-2; docs010_multitrack_flows_persona_architecture/007_personas_discussion.md#The-Key-Insight
- **relations:** copywriter persona roster (successor, persona-architecture register); multi-track flows
- **verify-later:** get_voice_for_page function existence
