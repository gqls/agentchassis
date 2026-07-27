# Register — public-api

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

1 concept, consolidated from 2 raw extractions (1 unique block, appearing twice
due to exact whole-block duplication in the cluster input file) across unit
U01.

**Stage-2 duplicate resolution (2026-07-14):** PUB-001 is the same underlying
plan as `admin-dashboard-and-api.md`'s ADM-007 + ADM-008, extracted under a
distinct category tag (`new:public-api`) because P2 (its source doc) is a
later/live version of the plan that ADM-007/008's source (`007b_public_api_plan_v2.md`)
was an earlier snapshot of. Full entry retained below for its distinct source
citation (P2 vs 007b); treat ADM-007 (endpoint plan) + ADM-008 (site_ownership
table) as the canonical pair — this entry is the pointer.

### PUB-001 — Public API plan: site_ownership junction + user-facing build/HITL endpoints (duplicate — see ADM-007 + ADM-008)
- **status:** aspirational
- **status-evidence:** P2 is an implementation plan (blocks 0–6, build order); Block 3 admin subset "implemented" per its own notes — the user-facing half was never built.
- **stage2-verified (2026-07-14):** confirmed unbuilt (0 hits for site_ownership or /api/v1/sites in any .go/.sql file) and confirmed duplicate of ADM-007 + ADM-008 — see those entries for the canonical write-up.
- **what:** `site_ownership` junction table (site/client/user/role) rather than columns on `sites` (shared sites; 15+ FKs untouched); all public queries scope through it. `POST /sites` writes build_queue + ownership (seed picks it up; 409 on existing). Endpoints for sites/status (work-item progress rollup), pages, work items with the HITL review flow (needs_human_review → provide-data-and-retry / retry / dismiss; retry converts to content_rewrite), specs read+write, assets, briefing HTTP-to-Kafka bridge, WebSocket build events.
- **sources:** P2_public_api_plan.md (full, docs024_key_docs_latest)
- **relations:** duplicate-of: ADM-007 (admin-dashboard-and-api.md, public REST API plan), ADM-008 (admin-dashboard-and-api.md, site_ownership table). admin API (ADM-002); needs_human_review status; build_queue.
- **verify-later:** site_ownership table; which blocks landed
