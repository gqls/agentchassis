# Register — entity-data

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

2 concepts, consolidated from 4 raw extractions (2 unique blocks, each appearing
twice due to exact whole-block duplication in the cluster input file) across
unit U21.

### ENT-001 — Entity data agent family (structured data drives pages)
- **status:** aspirational
- **status-evidence:** docs017/019b: site_entities/site_entity_relationships "(exist)", entity_sources/entity_sync_log "(planned)"; "First implementation target: boxing ticket/events site, then football tickets, then finance"; no later doc confirms the sync pipeline ran.
- **stage2-verified (2026-07-14):** partial → aspirational — site_entities/site_entity_relationships/entity_sources/entity_sync_log: 0 CREATE TABLE hits in any .sql file repo-wide; only non-doc hit is a code comment in docs/agent_docs/sql_for_tables/027_content_sources_table.sql:24,211 referencing site_entities.entity_type, but that table itself is never defined anywhere. Lif...
- **what:** Real-world entities (events, performers, venues, ticket tiers, products, articles) stored as typed JSONB rows with relationships, synced from configured sources (API/scrape/feed with field_mapping, poll intervals, rate limits), change-logged, and driving template-rendered pages with minimal LLM. Entity lifecycle is state-based, not time-based (announced → on_sale → selling_fast → sold_out → event_day → past → historical/cancelled) with per-state page and nav behaviour; status transitions auto-detected from source data. entity_sources.news_triggers defines which changes are newsworthy, bridging to the feed pipeline. Three of four stress-tested site types need it.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#5-Entity-Data-Agent-Family,#Entity-Lifecycle; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Entity-Data
- **relations:** news feed pipeline (entity_event triggers); Events/tickets vertical (ENT-002); products tables (superseded by entities); dogs-medicine entities unrelated
- **verify-later:** site_entities/site_entity_relationships rows; entity_sources/entity_sync_log existence; entity-data-agent definition

### ENT-002 — Events/tickets vertical (boxing first target)
- **status:** abandoned
- **status-evidence:** docs017/019b "Events / Tickets Site (first target — boxing, then football)... API sources: Ticketmaster, SeatGeek, BoxRec"; entity examples "Fury vs Joshua"; no boxing/tickets site appears in later portfolio lists.
- **what:** The planned first entity-driven site type: dense entity relationships (event↔performer↔venue↔ticket_tier), frequently-updating ticket tier data (price/availability) flowing to pages quickly, state-transition-driven news (fight announced, on sale, sold out, results), contextual per-event/per-performer navigation, and past events retained as permanent SEO assets.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Entity-Types-for-Events-Tickets,#Site-Type-Stress-Tests
- **relations:** Entity data agent family (ENT-001); news feed pipeline
- **verify-later:** any boxing/tickets site records
