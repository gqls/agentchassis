# Final taxonomy — settled categories vs. the seed spine

Stage 1 (extraction + consolidation) complete: 2026-07-13. This note is C3 —
the closing deliverable of the extraction/consolidation programme, comparing
what the taxonomy turned out to be against the docs024-spine seed in
`003_TAXONOMY_seed.md`.

## Headline numbers

- **1,627 concepts**, **107 category register files**, in `register/`.
- Consolidated from **2,185 raw concept blocks** across **32 extraction-unit
  files** (26 planned units + 8 session-limit recovery/gap-fill units), which
  themselves swept **~4,111 files** under `docs/`.
- Status distribution: 871 deployed · 274 partial · 271 aspirational ·
  99 superseded · 72 abandoned · 40 unknown.
- Master index: `register/000_concept_index.md`.

## How the taxonomy actually settled

The seed spine (`003_TAXONOMY_seed.md`) proposed ~30 categories from the
docs024 documentation index, plus ~17 provisional additions guessed from
directory names. Extraction agents were told the taxonomy was open and to
propose `NEW:<slug>` freely. They did — 274 raw blocks arrived tagged with a
`NEW:` category, spanning **65 distinct proposed slugs**. Consolidation then
collapsed obvious synonyms (see "Merges" below), landing on **107 final
category files** — roughly double the seed's ~30, reflecting how much
territory the seed spine (built for a narrower "current documentation
structure" purpose) didn't anticipate: the diagnosis/fix-loop machinery,
imagery as a rich multi-phase subsystem, the finetuning flywheel, several
distinct site-build "generations," and a long tail of early-era abandoned
agent-org experiments.

### Seed categories that held up largely as-is
development-guide, system-architecture, contracts-and-standards,
improvement-loop, adoption-pipeline, model-infrastructure, scheduler-and-tasks,
database-and-infrastructure, admin-dashboard-and-api, content-governance,
site-snapshots-and-revert, locks, storage-architecture, tool-lifecycle,
site-spec-and-classifier, dynamic-applications, llm-quality-testing,
navigation, batch-processing.

### Seed categories that split or were absorbed
- `debugging` (016) stayed one category but absorbed three sibling proposals
  (`resilience-self-heal`, `migration-governance`, `sql-change-management`) as
  separate-but-related files rather than folding into it.
- `design-composition` (025/026/027) held as one large file (80 concepts) —
  the seed correctly predicted this as one area, but it turned out to need a
  dedicated lineage entry (DES-056) tracing a 4-generation mechanism arc.
- `styling-render-pipeline` (036) split three ways in practice
  (styling-render-pipeline / navigation / link-management) — related enough to
  cluster together for consolidation, distinct enough to stay separate files.
- `site-plan-and-reconciler` (029/030, a seed provisional addition) turned out
  to be the seed's most under-scoped guess — it fanned out into 7 final
  categories (site-plan-and-reconciler, page-build-pipeline, build-pipeline,
  rebuild-cascade, work-dispatch, work-item-integrity, action-build-pipeline)
  covering the work-item/build-pipeline machinery in far more depth than the
  seed anticipated.
- `imagery` (a seed provisional addition) proved right in scope but needed two
  small siblings (data-charts, component-asset-pipeline) split out.
- `diagnosis-loop` and `fix-loop` (seed provisional additions) were correct
  calls — together they absorbed the largest share of NEW: proposals
  (context-assembly, contextkit-toolchain, context-pack-tooling,
  context-engineering-principles under diagnosis; autonomy-governance,
  autonomous-build-operate, autonomy-trust-model, reasoning,
  investigation-discipline, operating-doctrine, operator-practice under
  fix-loop) — this whole area is genuinely as rich as the user's original
  framing (docs019) suggested.

### Categories the seed missed entirely (now real, substantial register files)
- **finetuning-flywheel** and its cluster (model-infrastructure,
  rag-knowledge-base, llm-call-observability) — a whole subsystem the seed's
  docs024-only view never saw (it lives mostly in docs019/docs024-finetuning).
- **vm-backend-sites** — a genuinely new infrastructure class (persistent,
  non-reaped, internet-facing VM-hosted backends) that two independent
  pioneer projects (idea.uk, traffic_probe) converged on describing, absorbing
  the seed's guessed `backend-service-deployment`/`persistent-service-
  deployment` proposals.
- **business-strategy-products** cluster (idea-product, business-
  intelligence-platform, conversion-playbooks, portfolio-evolution, vertical-
  knowledge-architecture, payments, legal-and-compliance, marketing, seo,
  affiliate-commerce) — the seed's `business-strategy` guess was far too
  narrow for how much product/commercial material the corpus actually holds.
- **hitl-onboarding-agentorg** long tail (agent-spawning-and-groups, agent-
  memory-and-evolution, persona-architecture, flows-and-narrative, org-
  framework, agent-tree-navigation, agent-swarm-simulations, agent-definition-
  registry) — mostly early-era (2025-era) abandoned agent-organization
  experiments the seed had no way to anticipate.
- **site-chatbot**, **vonc**, **social-media** — site-specific/product-specific
  material that legitimately needed platform-level category status rather
  than folding into a generic "site-case-studies" bucket.

### Merges performed during consolidation (collapsing near-synonym proposals)
- `backend-service-deployment` + `persistent-service-deployment` → `vm-backend-sites`
- `work-item-system` + `dispatch-pipeline` → `work-dispatch`
- `site-build-pipeline` + `site-build-orchestration-generations` → `build-pipeline`
- `legal-and-compliance` + `legal-liability` → `legal-and-compliance`
- `affiliate-and-products` + `affiliate-commerce` → `affiliate-commerce`
- (tool-lifecycle's "fork-on-deploy" concept merged into tool-library's
  near-identical entry, without a full category merge)

## Known limitations carried into stage 2

- **Status tags are documentary signals, not verified fact.** Several
  consolidator agents independently upgraded a status (aspirational→partial)
  after spot-checking the live repo for a named file/directory
  (`platform/discovery/`, `platform/evolution/` both confirmed to exist) —
  this is the kind of check stage 2 should do systematically, not just where
  an agent happened to have a moment to look.
- **A few likely-duplicate concepts were deliberately left unmerged** across
  category boundaries because the instructions scoped merging to within a
  consolidator's own assigned categories. Flagged pairs worth a stage-2 look:
  `site-chatbot` CHAT-007 vs `saas-isolation-architecture` SAAS-001 (same
  "isolated chat satellite" design, two framings); `public-api` PUB-001 vs
  `admin-dashboard-and-api` ADM-007/008 (same plan, two doc versions);
  `vet-med-pricing` VET-011 (thunder-reaper cost gate) is very likely
  mistagged and belongs under model-infrastructure/finetuning.
- **Two credential-hygiene flags surfaced during extraction, unresolved:** a
  hardcoded Thunder API bearer token in
  `docs024_key_docs_latest/finetuning/working/flywheel_docs/ssh_probe.sh`, and
  what appears to be a real AWS console password + account ID in
  `docs024_key_docs_latest/idea.uk/golang_files/README_email.md`. Worth
  rotating independently of this project.
- **Two evidence tensions flagged for stage-2 verification:** whether the
  multi-cluster `dispatch_agent`/`remote-job-spawner` mechanism is actually
  wired into any live workflow (sources disagree), and whether the trained
  Llama 3.3 70B LoRA adapter has ever been wired into production inference
  (as of the latest dated source, 2026-07-10, it had not).

## What's next (stage 2 and 3, per the original charter)

- **Stage 2:** analyse the agent-chassis code, workflows, and DB to determine
  the true state of each concept — starting with the `partial`/`unknown`
  entries, since `deployed`/`abandoned` claims are lower-risk to leave as-is
  pending spot-checks, and the flagged tensions/duplicates above.
- **Stage 3:** build an expert council agent per concept-area (roughly one
  per register file, or per closely-related cluster) to join the diagnosis/
  fix-loop's council — the register's category structure is designed to map
  onto council seats directly.
