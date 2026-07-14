
<!-- SOURCE: U19_sql_tables_components.md -->
### Business-intel sweep/verify collection pipeline (vet-intel)
- **category:** NEW:business-intel-collection
- **status-signal:** deployed
- **status-evidence:** Operational scheduled tasks: vet-batch-verify (claims pending collection_tasks), vet-task-reset → broadened vet-cleanup self-healer (fails orchestrations stuck AWAITING_RESPONSES >20 min, resets stuck collection_tasks, "breaks the stall chain"), vet-sweep-continue (batches of 200 unswept areas); later re-pointed at a dedicated vet-intel pod on system.agent.vet-intel.requests.
- **what:** The area-sweep → collection_tasks → batch-verify pipeline that builds the verified business directory (vertical: veterinary) which CH enrichment then deepens. Includes the operational self-healing pattern and the dedicated-pod routing decision (vet-intel instead of the generic agent).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#vet-tasks and #vet-cleanup and #vet-intel-setup; docs/agent_docs/sql_for_tables/023_companies_house_data.sql#pre-query
- **relations:** companies-house enrichment; batch-processing; scheduler self-healing.
- **verify-later:** business_intel.businesses / collection_tasks schemas (defined elsewhere); vet-intel agent definition.

<!-- SOURCE: U19_sql_tables_components.md -->
### Business-intel sweep/verify collection pipeline (vet-intel)
- **category:** NEW:business-intel-collection
- **status-signal:** deployed
- **status-evidence:** Operational scheduled tasks: vet-batch-verify (claims pending collection_tasks), vet-task-reset → broadened vet-cleanup self-healer (fails orchestrations stuck AWAITING_RESPONSES >20 min, resets stuck collection_tasks, "breaks the stall chain"), vet-sweep-continue (batches of 200 unswept areas); later re-pointed at a dedicated vet-intel pod on system.agent.vet-intel.requests.
- **what:** The area-sweep → collection_tasks → batch-verify pipeline that builds the verified business directory (vertical: veterinary) which CH enrichment then deepens. Includes the operational self-healing pattern and the dedicated-pod routing decision (vet-intel instead of the generic agent).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#vet-tasks and #vet-cleanup and #vet-intel-setup; docs/agent_docs/sql_for_tables/023_companies_house_data.sql#pre-query
- **relations:** companies-house enrichment; batch-processing; scheduler self-healing.
- **verify-later:** business_intel.businesses / collection_tasks schemas (defined elsewhere); vet-intel agent definition.
