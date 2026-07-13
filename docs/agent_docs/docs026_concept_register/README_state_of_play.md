Stage 1 of the concept register is complete. Here's the full picture:

What was built

docs/agent_docs/docs026_concept_register/ — a new directory, nothing else touched:

- 1,627 concepts across 107 category register files, each entry with status (deployed/partial/aspirational/superseded/abandoned/unknown), dated evidence, sources, relations, and stage-2 verify-later pointers
- Consolidated from 2,185 raw concept blocks, extracted by 34 agent runs sweeping essentially all ~4,111 files under docs/
- Master index: register/000_concept_index.md
- Final taxonomy comparison: 005_TAXONOMY_final.md

How the taxonomy landed

Started from the docs024 spine (~30 categories) but stayed open by design — extraction agents proposed 65 distinct NEW: categories, which consolidation collapsed into a settled 107. The biggest surprises versus the
seed: the diagnosis-loop/fix-loop machinery, imagery, and the site-build/work-item pipeline all turned out far richer than the docs024 index suggested; finetuning-flywheel and a long tail of abandoned early-era
agent-org experiments weren't anticipated at all.

Directly useful to your fix-loop project

- FIX-036 (in the register) is explicitly the vision of a wider council roster — flagged by a consolidator as "the seam this concept register itself is meant to help fill." That's stage 3.
- Several concepts were independently rediscovered 4-6 times across different documentation eras (e.g. "adoption writes first, classifier consumes," the wrapper-orchestrator pattern) — strong signals for which
  concepts deserve a dedicated council seat first.

Loose ends worth your attention

- Two credential leaks found in docs: a Thunder API bearer token and what looks like a real AWS password, both noted in memory for rotation.
- A couple of evidence tensions flagged for stage 2: whether multi-cluster dispatch is actually wired into any live workflow, and whether the trained LoRA adapter was ever wired into production inference.

Stage 2 (verifying each concept against real code/DB) and stage 3 (building council agents per concept area) are next, whenever you want to pick them up.

--