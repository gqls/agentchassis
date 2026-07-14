# Register — public-api

1 concept, consolidated from 2 raw extractions (1 unique block, appearing twice
due to exact whole-block duplication in the cluster input file) across unit
U01.

### PUB-001 — Public API plan: site_ownership junction + user-facing build/HITL endpoints
- **status:** aspirational
- **status-evidence:** P2 is an implementation plan (blocks 0–6, build order); Block 3 admin subset "implemented" per its own notes — the user-facing half was never built.
- **what:** `site_ownership` junction table (site/client/user/role) rather than columns on `sites` (shared sites; 15+ FKs untouched); all public queries scope through it. `POST /sites` writes build_queue + ownership (seed picks it up; 409 on existing). Endpoints for sites/status (work-item progress rollup), pages, work items with the HITL review flow (needs_human_review → provide-data-and-retry / retry / dismiss; retry converts to content_rewrite), specs read+write, assets, briefing HTTP-to-Kafka bridge, WebSocket build events.
- **sources:** P2_public_api_plan.md (full, docs024_key_docs_latest)
- **relations:** admin API (admin-dashboard-and-api.md ADM-002); needs_human_review status; build_queue. This is very likely the same underlying plan as admin-dashboard-and-api.md's ADM-007 ("Public REST API for the site-building pipeline", from an earlier archived doc version `007b_public_api_plan_v2.md`) and ADM-008 ("site_ownership table / ownership model", same archived doc) — P2 appears to be the live/current version of the plan that 007b was an earlier snapshot of. Kept as a separate entry here because it was tagged with this distinct assigned category (`new:public-api`) in extraction rather than `admin-dashboard-and-api`; recommend reconciling all three into one entry in stage 2.
- **verify-later:** site_ownership table; which blocks landed
