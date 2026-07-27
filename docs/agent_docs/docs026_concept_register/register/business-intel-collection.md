# Register — business-intel-collection

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 2 raw extractions (1 unique block, appearing twice
due to exact whole-block duplication in the cluster input file) across unit U19.

Note: this category was seeded with only one distinct concept in this cluster.
It is closely related to, but not the same real-world thing as, vet-med-pricing
(medicine PRICE collection from retailers) — this pipeline instead discovers and
verifies vet PRACTICES as businesses, which companies-house-enrichment.md then
deepens. The two were considered for a category merge per the brief but kept
separate since they describe distinct pipeline stages on different entities
(businesses vs. medicine products); see cross-references below.

### BIC-001 — Business-intel sweep/verify collection pipeline (vet-intel)
- **status:** deployed
- **status-evidence:** Operational scheduled tasks: vet-batch-verify (claims pending collection_tasks), vet-task-reset → broadened vet-cleanup self-healer (fails orchestrations stuck AWAITING_RESPONSES >20 min, resets stuck collection_tasks, "breaks the stall chain"), vet-sweep-continue (batches of 200 unswept areas); later re-pointed at a dedicated vet-intel pod on system.agent.vet-intel.requests.
- **what:** The area-sweep → collection_tasks → batch-verify pipeline that builds the verified business directory (vertical: veterinary) which CH enrichment then deepens. Includes the operational self-healing pattern and the dedicated-pod routing decision (vet-intel instead of the generic agent). The same underlying agents (area-sweep-orchestrator, area-sweep-discoverer, vet-batch-processor, vet-practice-verifier) are also named, with SQL migration references, in the vet-med-pricing.md VET-001 entry's source material (U18) — that entry treats them as a companion sub-pipeline rather than duplicating them here.
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#vet-tasks,#vet-cleanup,#vet-intel-setup; docs/agent_docs/sql_for_tables/023_companies_house_data.sql#pre-query; 037_area_sweep_discoverer.sql; 038_area_sweep_orchestrator.sql; 063_vet_batch_processor.sql; 063b_vet_practice_verifier.sql
- **relations:** companies-house-enrichment.md CH-001 (consumes verified businesses); Vet med pricing pipeline (vet-med-pricing.md VET-001, sibling pipeline on the same business-intel pod); batch-processing; scheduler self-healing
- **verify-later:** business_intel.businesses / collection_tasks schemas (defined elsewhere); vet-intel agent definition
