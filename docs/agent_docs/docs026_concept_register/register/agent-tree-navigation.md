# Register — agent-tree-navigation

1 concept, consolidated from 2 raw extractions (1 unique block, present twice
in the source cluster file due to mechanical duplication in the input), from
unit U22.

### ATN-001 — Agent hierarchy tree navigation (ltree paths + subtree summaries + live viewer)
- **status:** aspirational
- **status-evidence:** Raw design-session transcript only ("The data model changes are small... The bigger piece of work is the API endpoints and the tree viewer UI"); no implementation claimed, buried inside a 273KB chat-transcript file the rest of the extraction treats as header-scan.
- **what:** A proposal for navigating the `orchestration_states` parent/child tree at massive scale (millions of rows, 8-10 levels deep) without recursive-CTE cost: add an `ltree`-typed `tree_path` column (materialised ancestry path, set cheaply at spawn time by prepending the parent's own path), enrich the existing `subtree_agents` jsonb with rolling status/type/failure counts so a UI can show summaries and only fetch detail on expand, add a `tags` jsonb column (GIN-indexed) for semantic queries ("find all bankrupt fast-food agents" rather than tree position), and a lightweight `agent_tree_index` table (~200 bytes/row, no heavy jsonb blobs) so a million-row tree fits comfortably in cache. Proposed REST API (`/trees/{correlation_id}`, `/agents/{id}/children`, `/agents/{id}/subtree`, `/trees/{id}/search?agent_type=...&status=...`) plus a WebSocket live tree viewer fed from existing Kafka response topics, giving filesystem-like drill-down ("root > uk-economy > fast-food-sector > dominos-agent-47").
- **sources:** docs021_multiclustering/021_2026-02-28-20-03-32-multi-cluster-dispatch-design.txt (sections "The fundamental query patterns" through "The user experience")
- **relations:** Multi-cluster scaling tiers; orchestration_states schema; Agent swarm simulation ideas (agent-swarm-simulations register — this viewer was requested specifically to make the swarm-simulation ideas practically navigable)
- **verify-later:** orchestration_states.tree_path / tags columns; any agent_tree_index table; core-manager tree API endpoints
