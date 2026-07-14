# Register — admin-dashboard-and-api

10 concepts, consolidated from 22 raw extractions (11 unique blocks, each
appearing twice due to exact whole-block duplication in the cluster input file)
across units U01, U09, U12, U24a, U26.

### ADM-001 — Admin dashboard + nginx gateway architecture
- **status:** deployed
- **status-evidence:** 012 full; 013 phases 1-11 ✅ except user portal.
- **what:** React SPA served by nginx that also gateways /api/v1/auth→auth-service and /api/v1→core-manager (rate limits, timeouts, immutable asset caching, security headers). Views: Sites (lock badges), Work Items (three review flows: placeholder/checkpoint/standard; bulk retry; cross-site tab), Pages (three-level browser, Fields/HTML/Brief tabs, page-purpose bar, suppressed-section restore), Direction (spec cards, pin/propagate), Media (assets + references). Access via port-forward today; WireGuard/bastion planned.
- **sources:** 012 full; 013 status table
- **relations:** content governance edit paths; admin API endpoints (ADM-002); WireGuard VPN admin-access (ADM-006, the planned access upgrade)
- **verify-later:** frontends/admin-dashboard; nginx conf

### ADM-002 — Admin API current state: dual-auth gateway, inventory, and fix blocks
- **status:** partial
- **status-evidence:** P3_admin_api_plan.md (live) headings show known issues (bugs, hardcoded/mock data, missing wiring) with blocks A–F and a target route map; its predecessor 008b_admin_api_plan_v2.md (archived) itemises the same audit with specific bug IDs (B1 MySQL-syntax `CURDATE()` in dashboard, B2 `orchestrator_state` vs `orchestration_states` naming mismatch, B3 cartesian join, B4 missing agent-instances proxy) and H1-H7 hardcoded/mock values (Kafka topics, agent health, usage metrics); 008b's build order marked blocks A-D "Not started" and E3 "Implemented" — live 012_admin_dashboard.md's status (013 phases 1-11 done except user portal) suggests the site/work-item HITL subset (E3) has since shipped.
- **what:** Two services share one gateway: auth-service handles auth/user/subscription/projects directly and proxies admin site routes to core-manager, with dual JWT auth validation on both sides. The admin API plan doc (evolved from archived v2 `008b` to live `P3`) inventories every current route, catalogues concrete bugs and hardcoded/mock values, and sequences fixes into blocks: A fix bugs, B wire handlers, C replace hardcoded data, D performance, E new site-domain endpoints (its HITL subset E3 already implemented), F agent-definition admin improvements.
- **sources:** P3_admin_api_plan.md (header-scan); archive_april_26/008b_admin_api_plan_v2.md#current-admin-endpoints,#known-issues,#implementation-plan; docs024_key_docs_latest/012_admin_dashboard.md
- **relations:** admin dashboard (ADM-001); public API plan (public-api.md PUB-001); gateway proxy pattern; HITL review flow; site admin handlers
- **verify-later:** internal/core-manager/admin/{dashboard_handlers,system_handlers,agent_handlers,site_admin_handlers}.go; which blocks A-F completed (012 suggests site-domain endpoints largely live)

### ADM-003 — Core-manager API server surface (spec pin/unpin among admin routes)
- **status:** deployed
- **status-evidence:** old2/server.go (Core Manager API server, gin router, persona repo, kafka producer); PLAN_lock_coherence correction: "server.go routes POST /sites/:site_id/specs/:aspect/pin and .../unpin to specAdminHandlers.HandlePinSpec / HandleUnpinSpec (Phase 4 'Spec Direction Control')".
- **what:** The core-manager HTTP API (internal/core-manager) is a separate reader/writer surface from the chassis — notably exposing spec pin/unpin endpoints that keep Pattern B semantics alive even though chassis code has zero `pinned` references. Any lock-model retirement must account for admin-API consumers, not just chassis greps.
- **sources:** old2/server.go; PLAN_lock_coherence.md#pinning-verification
- **relations:** lock-model coherence step 5; site_specs supersede-then-insert writes
- **verify-later:** specAdminHandlers read/write of site_specs.pinned

### ADM-004 — Work-item HITL model: approve/reject endpoints on pending_review status
- **status:** superseded
- **status-evidence:** `007a_public_api_plan_v1.md`: `POST /work-items/:item_id/approve|reject`; live `P2_public_api_plan.md`/`P3_admin_api_plan.md` have no approve/reject endpoints or `pending_review`/`rejected` statuses anywhere.
- **what:** The original API plan modelled human review as a binary approval gate on work items, with specs read-only initially. Replaced end-to-end by `needs_human_review` items with three resolution paths (provide missing spec data + retry, retry unchanged, or dismiss with a resolution note), and `PATCH /specs/:aspect` as a first-class, versioned write path feeding that retry flow.
- **sources:** old/older1/007a_public_api_plan_v1.md#"Work Items (build progress + HITL)"; docs024_key_docs_latest/P2_public_api_plan.md#"HITL Review Flow"
- **relations:** content-governance (locks, HITL)
- **verify-later:** grep core-manager handlers for any surviving `pending_review`/`HandleApproveWorkItem`/`HandleRejectWorkItem`

### ADM-005 — Admin work-item reassign + force-complete override endpoints
- **status:** superseded
- **status-evidence:** `008a_admin_api_plan_v1.md` E3 table has `reassign`/`force-complete`; live `P3_admin_api_plan.md`'s equivalent table has neither, only generic `PATCH`, `retry`, `resolve` (all Implemented).
- **what:** The original admin plan gave two narrow, single-purpose override endpoints for stuck work items: reassign the handler agent, or force-mark-complete with an arbitrary result. Generalised instead into one `PATCH` endpoint plus the shared `retry`/`resolve` pair — reassign and force-complete as distinct named actions never shipped.
- **sources:** old/older1/008a_admin_api_plan_v1.md#"E3: Work item administration"; docs024_key_docs_latest/P3_admin_api_plan.md#"E3: Work item administration + HITL review — IMPLEMENTED"
- **relations:** work-item HITL model (ADM-004)
- **verify-later:** confirm `site_admin_handlers.go` has no `HandleReassignWorkItem`/`HandleForceComplete`

### ADM-006 — WireGuard VPN admin-access implementation detail
- **status:** superseded
- **status-evidence:** Archive contains full runnable K8s manifests and nginx configs; live `012_admin_dashboard.md`'s condensed section keeps only one-line summaries, drops every YAML/config block.
- **what:** Three documented approaches to securely expose the admin dashboard without public ingress: (A) WireGuard-in-cluster with full K8s manifests, (B) external VM bastion with WireGuard + nginx + TLS + rate limiting, (C) plain `kubectl port-forward`. The live doc retains only the decision framework, not the deployable configuration.
- **sources:** archive_april_26/019_admin_access_infrastructure.md (whole file); docs024_key_docs_latest/012_admin_dashboard.md#"Network Access Options"
- **relations:** admin dashboard (ADM-001); auth-service JWT/RequireRole security layer
- **verify-later:** check whether WireGuard was ever actually deployed or whether the system is still on Option C

### ADM-007 — Public REST API for the site-building pipeline
- **status:** aspirational
- **status-evidence:** 007b build-order table (2026-04) lists Blocks 0-6 (site_ownership, HandleCreateSite, pages, work-items, specs, briefing bridge, websockets) all "Not started" except the admin subset; only admin `site_admin_handlers.go` is "Written — ready to deploy".
- **what:** Plan to expose `sites`, `pages`, `site_work_items`, `site_specs`, `assets`, and briefing over `/api/v1/sites/*`, tenant-scoped via a new `site_ownership` junction table, so users can submit domains, watch build progress, and resolve HITL reviews over HTTP. Reads/writes the same DB the agents use (build_queue, dispatch loop pick changes up); Kafka only touched for the briefing HTTP→Kafka bridge. The user-scoped public half was never built.
- **sources:** archive_april_26/007b_public_api_plan_v2.md#public-api-endpoints,#ownership-model,#build-order
- **relations:** depends on site_ownership (ADM-008); complements Admin API (ADM-002); superseded reference by live docs/api/reference.html (ADM-010). Likely the same underlying plan as public-api.md PUB-001 (a later/live doc version, P2); see PUB-001 for the reconciliation note.
- **verify-later:** internal/core-manager/sites/*.go (planned, may not exist); site_ownership migration

### ADM-008 — site_ownership table / ownership model
- **status:** abandoned
- **status-evidence:** 007b: "The `sites` table has no ownership columns"; site_ownership migration listed Block 0 "Not started"; 008b Block E notes admin endpoints "work without the site_ownership migration because they're admin-only".
- **what:** Proposed `site_ownership(site_id, client_id, user_id, role)` junction table to attach user identity to agent-created sites (which carry none), enabling per-user scoping of the public API. Chosen as a junction (not columns on `sites`) because sites can be shared and `sites` has 15+ FKs. Never created; admin API sidesteps it.
- **sources:** archive_april_26/007b_public_api_plan_v2.md#ownership-model; 008b_admin_api_plan_v2.md#block-e
- **relations:** blocks the public API (ADM-007); `assign`/`trigger-build` admin endpoints
- **verify-later:** grep for site_ownership in migrations and core-manager

### ADM-009 — React admin dashboard for build review
- **status:** partial
- **status-evidence:** 007b: `site-admin-dashboard.jsx` "Written — uses mock data until API connected (toggle `useMock=false`)", Tailwind utility classes; three views (Dashboard, Review Queue, Review Detail).
- **what:** A React frontend rendering site cards with progress bars, a `needs_human_review` queue with Review/Retry/Dismiss, and a review-detail view with an editable identity-spec JSON + "Save Spec & Retry". Runs on mock data pending API wiring.
- **sources:** archive_april_26/007b_public_api_plan_v2.md#react-admin-dashboard; 008b#files-summary
- **relations:** Admin API Block E (ADM-002); HITL review flow
- **verify-later:** site-admin-dashboard.jsx location

### ADM-010 — AI Persona Platform public API
- **status:** superseded
- **status-evidence:** docs/api/reference.html is a generated Redoc bundle titled "AI Persona Platform API" covering the persona-era surface; the current API surface is the admin dashboard/nginx gateway (spine 012) and the persona-instance concepts do not appear in current docs.
- **what:** The v1 REST surface of the persona era: JWT auth (register/login/refresh/validate/logout), user profile/password/delete, projects CRUD, subscription with usage stats and quota checks, persona template listing, persona instance list/create, health check, and a WebSocket connection endpoint. Documents the original productisation of the platform as "AI personas" for end users.
- **sources:** docs/api/reference.html (tags: Authentication, Users, Projects, Subscriptions, Templates, Instances, System, WebSocket; paths /api/v1/auth/*, /api/v1/projects, /api/v1/subscription/*, /api/v1/personas/instances)
- **relations:** three-database architecture (auth DB backs these endpoints); superseded by current admin-dashboard-and-api (ADM-001)
- **verify-later:** which endpoints survive in core-manager/api-gateway code
