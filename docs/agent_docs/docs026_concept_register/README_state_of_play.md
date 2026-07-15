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

● Stage 2 is now fully complete. Summary of what landed:

The final sweep (100 agents, 174 concepts checked): 18 corrections confirmed, 9 overturned.

- A scheduler cluster (SCH-007/008/009): three concepts, same shape — real fixes landed in cmd/scheduler/main.go (ownership gap, starvation prevention, fire_message handling) but the docs never caught up.
  Three independent confirmations of the same drift.
- A new failure-mode class: DOC-064 was tagged abandoned because extraction's search was scoped to one doc subtree and never reached a sibling folder holding a byte-identical live copy — evidence that was
  never found, distinct from the present-tense-plan misreading batches 1-2 found.
- Several "bundled" concepts were half-superseded — a claimed clean replacement, but part of the old mechanism was still live and wired.

Stage 2 grand total, all 1,627 concepts now checked at least once: 124 corrections confirmed, 106 overturned by the adversarial pass (~7.6% error rate). Final distribution: 853 deployed / 257 partial /
290 aspirational / 90 superseded / 67 abandoned / 21 unknown / 49 convention.

All three running docs, the master index, and memory are updated to reflect completion. Two items remain, both explicitly yours to call: credential rotation (RUNBOOK B1) and whether/when to implement the
stage-3 council design against the live fix-loop workflow (RUNBOOK B4) — that's a cross-workstream production change I've deliberately left for your sign-off rather than proceeding on unilaterally.


---

