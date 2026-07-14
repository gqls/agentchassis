
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Public API plan: site_ownership junction + user-facing build/HITL endpoints
- **category:** NEW:public-api
- **status-signal:** aspirational
- **status-evidence:** P2 is an implementation plan (blocks 0–6, build order); Block 3 admin subset "implemented" per its own notes
- **what:** site_ownership junction table (site/client/user/role) rather than columns on sites (shared sites; 15+ FKs untouched); all public queries scope through it. POST /sites writes build_queue + ownership (seed picks it up; 409 on existing). Endpoints for sites/status (work-item progress rollup), pages, work items with the HITL review flow (needs_human_review → provide-data-and-retry / retry / dismiss; retry converts to content_rewrite), specs read+write, assets, briefing HTTP-to-Kafka bridge, WebSocket build events.
- **sources:** P2 full
- **relations:** admin API; needs_human_review status; build_queue
- **verify-later:** site_ownership table; which blocks landed

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Public API plan: site_ownership junction + user-facing build/HITL endpoints
- **category:** NEW:public-api
- **status-signal:** aspirational
- **status-evidence:** P2 is an implementation plan (blocks 0–6, build order); Block 3 admin subset "implemented" per its own notes
- **what:** site_ownership junction table (site/client/user/role) rather than columns on sites (shared sites; 15+ FKs untouched); all public queries scope through it. POST /sites writes build_queue + ownership (seed picks it up; 409 on existing). Endpoints for sites/status (work-item progress rollup), pages, work items with the HITL review flow (needs_human_review → provide-data-and-retry / retry / dismiss; retry converts to content_rewrite), specs read+write, assets, briefing HTTP-to-Kafka bridge, WebSocket build events.
- **sources:** P2 full
- **relations:** admin API; needs_human_review status; build_queue
- **verify-later:** site_ownership table; which blocks landed
