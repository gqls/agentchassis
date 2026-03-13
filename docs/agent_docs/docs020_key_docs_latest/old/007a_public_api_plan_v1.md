# 007 — Public API Implementation Plan

How to expose the site-building pipeline via HTTP. Covers the admin API current state, public API design, ownership model, implementation blocks, and build order.

---

## Current Admin API State

Two services, one gateway proxy.

### Auth Service (handles directly)

| Group | Endpoints | Status |
|---|---|---|
| `/api/v1/auth/*` | register, login, refresh, validate, logout | Working |
| `/api/v1/user/*` | profile CRUD, password change, delete | Working |
| `/api/v1/subscription/*` | get subscription, usage stats, quota check | Working |
| `/api/v1/projects/*` | CRUD scoped by user+client | Working |
| `/api/v1/admin/users/*` | user management, permissions | Working |
| `/api/v1/admin/subscriptions/*` | list, create, update | Working |

### Core Manager (proxied via gateway)

| Group | Endpoints | Status |
|---|---|---|
| `/api/v1/templates/*` | persona template CRUD | Working, admin only |
| `/api/v1/personas/instances/*` | instance CRUD | Working, tenant-scoped |
| `/api/v1/admin/clients/*` | create, list, usage | Working |
| `/api/v1/admin/system/*` | status, Kafka topics | Working, health endpoint returns hardcoded Kafka |
| `/api/v1/admin/workflows/*` | list, get, resume/terminate | Working, queries orchestration_states |
| `/api/v1/admin/agent-definitions/*` | list, create, update, topic verify/recreate | Working |
| `/api/v1/admin/agent-instances/*` | list, get, toggle, restart | Working, health returns hardcoded values |
| `/api/v1/admin/dashboard` | aggregated metrics | Working, some MySQL syntax in queries needs fixing |
| `/ws` | WebSocket events | Working |

### Gateway pattern

Auth-service receives all HTTP requests. For paths it owns (auth, user, subscription, projects, admin/users, admin/subscriptions) it handles directly. For everything else it proxies to core-manager via `gateway/handlers.go`, enriching with `X-User-ID`, `X-Client-ID`, `X-User-Role`, `X-User-Tier`, `X-User-Email` headers.

Routes use `.Any()` wildcard matching — core-manager defines actual method handling:
```go
adminGroup.Any("/clients", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/clients/*path", gatewayHandler.HandleAdminRoutes)
```

### What the admin API does NOT cover

The entire site-building domain is invisible to HTTP:

- `sites`, `pages`, `page_components` — no endpoints
- `build_queue` — no HTTP write path, only CLI → Kafka or direct DB
- `site_work_items` — no visibility, no HITL approval via HTTP
- `site_specs` — no read or write endpoints
- `assets` — no listing
- Briefing questionnaire HITL — Kafka only
- Site ownership — `sites` table has no `user_id` or `client_id`

---

## Ownership Model

### The problem

The `sites` table has no ownership columns. Sites are created by agents (via `ensure_site_record`) or by `seed_build_queue` reading from `build_queue`. Neither path carries user identity. The public API needs to know which user owns which sites.

### Solution: `site_ownership` table

A junction table rather than adding columns to `sites`, because sites can be shared (team members, admin access) and because `sites` is heavily referenced and we want to avoid schema changes to a table that 15+ foreign keys point to.

```sql
CREATE TABLE site_ownership (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id     uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    client_id   text NOT NULL,
    user_id     text NOT NULL,
    role        text NOT NULL DEFAULT 'owner',  -- owner, editor, viewer
    created_at  timestamptz DEFAULT now(),
    
    UNIQUE(site_id, user_id)
);

CREATE INDEX idx_site_ownership_user ON site_ownership (client_id, user_id);
CREATE INDEX idx_site_ownership_site ON site_ownership (site_id);
```

### How it gets populated

- `POST /api/v1/sites` (public API) → writes to `build_queue` AND inserts a `site_ownership` row with `role = 'owner'`
- `seed_build_queue` action → creates the site record, links to the ownership row via site_id
- Admin can assign sites to users via admin endpoint

### Scoping queries

All public API queries filter through `site_ownership`:
```sql
SELECT s.* FROM sites s
JOIN site_ownership so ON s.id = so.site_id
WHERE so.client_id = $1 AND so.user_id = $2
```

---

## Public API Endpoints

All under `/api/v1/sites`, protected by `RequireAuth` + `TenantMiddleware`. Scoped to the user's sites via `site_ownership`.

### Sites

| Method | Path | Action |
|---|---|---|
| `POST` | `/sites` | Submit domain for building |
| `GET` | `/sites` | List user's sites |
| `GET` | `/sites/:site_id` | Get site detail |
| `PATCH` | `/sites/:site_id` | Update settings |
| `GET` | `/sites/:site_id/status` | Lightweight build progress |

**POST /sites** — the main entry point. Creates a `build_queue` row + `site_ownership` row. Does NOT trigger Kafka directly; the `seed_build_queue` action picks it up on the next cycle. If the domain already exists as a site, returns 409 Conflict.

Request:
```json
{
    "domain": "finetuning.uk",
    "objective": "AI model fine-tuning consultancy",
    "direction": {}
}
```

Response (201):
```json
{
    "queue_id": "uuid",
    "domain": "finetuning.uk",
    "status": "queued",
    "message": "Domain queued for building"
}
```

Once seeded, subsequent `GET /sites` will show the site with its build status.

**GET /sites/:site_id/status** — lightweight progress view:
```json
{
    "site_id": "uuid",
    "domain": "finetuning.uk",
    "build_status": "building",
    "work_items": {
        "total": 12,
        "complete": 5,
        "in_progress": 2,
        "pending": 4,
        "failed": 1
    },
    "last_activity": "2026-02-24T10:30:00Z",
    "current_phase": "content_writing"
}
```

### Pages

| Method | Path | Action |
|---|---|---|
| `GET` | `/sites/:site_id/pages` | List pages |
| `GET` | `/sites/:site_id/pages/:page_id` | Get page detail |
| `PATCH` | `/sites/:site_id/pages/:page_id` | Update content_direction, trigger rebuild |

### Work Items (build progress + HITL)

| Method | Path | Action |
|---|---|---|
| `GET` | `/sites/:site_id/work-items` | List, filterable by status/domain/item_type |
| `GET` | `/sites/:site_id/work-items/:item_id` | Get detail |
| `POST` | `/sites/:site_id/work-items/:item_id/approve` | HITL: pending_review → approved |
| `POST` | `/sites/:site_id/work-items/:item_id/reject` | HITL: → rejected, with reason |
| `POST` | `/sites/:site_id/work-items/:item_id/retry` | Failed → triaged, increment attempt |

`GET /sites/:site_id/work-items` supports query params: `?status=in_progress&domain=build&item_type=needs_content_page&limit=20&offset=0`

### Site Specs (read-only initially)

| Method | Path | Action |
|---|---|---|
| `GET` | `/sites/:site_id/specs` | All current spec aspects |
| `GET` | `/sites/:site_id/specs/:aspect` | One aspect + recent history |

Later: `PATCH /sites/:site_id/specs/:aspect` for manual overrides (writes via `write_site_spec` with `source: 'manual'`).

### Assets

| Method | Path | Action |
|---|---|---|
| `GET` | `/sites/:site_id/assets` | List site assets (logos, hero images) |

### Briefing

| Method | Path | Action |
|---|---|---|
| `GET` | `/sites/:site_id/briefing` | Current briefing state: questions asked/pending/answered |
| `POST` | `/sites/:site_id/briefing/respond` | Submit answers to pending questions |

The briefing response writes a HITL response message. Two approaches, decide during implementation:

**Option A** — HTTP handler writes directly to Kafka responses topic (using the existing kafkaProducer in core-manager handlers). Matches the pattern in `HandleResumeWorkflow`.

**Option B** — HTTP handler updates `awaited_requests` in DB and lets the stale orchestration sweeper synthesize the Kafka message. Simpler but adds latency (up to 60s sweeper interval).

Option A is the likely choice — it's what `HandleResumeWorkflow` already does, and we have the Kafka producer available.

### WebSocket Additions

New event types for the existing `/ws` infrastructure:

```
site.build.progress         — work item status changes
site.build.complete         — all items done for a site
site.briefing.question      — new question needs answering
site.work_item.approval     — item moved to pending_review
```

These can be emitted by a Kafka consumer that watches orchestration state changes, or by the handler actions themselves posting to a notifications topic.

---

## Admin API Additions

Alongside the public API, the admin group needs site management endpoints for operations support.

### Admin Site Management

Under `/api/v1/admin/sites`, admin-only:

| Method | Path | Action |
|---|---|---|
| `GET` | `/admin/sites` | List all sites (across all users/clients), filterable |
| `GET` | `/admin/sites/:site_id` | Full detail including ownership |
| `POST` | `/admin/sites/:site_id/assign` | Assign site to a user |
| `POST` | `/admin/sites/:site_id/trigger-build` | Manually trigger dispatch loop for a site |
| `DELETE` | `/admin/sites/:site_id` | Soft delete |

### Admin Build Queue

Under `/api/v1/admin/build-queue`, admin-only:

| Method | Path | Action |
|---|---|---|
| `GET` | `/admin/build-queue` | List queue entries with status |
| `POST` | `/admin/build-queue` | Bulk insert domains |
| `POST` | `/admin/build-queue/seed` | Trigger seed_build_queue manually |

### Admin Work Items

Under `/api/v1/admin/work-items`, admin-only:

| Method | Path | Action |
|---|---|---|
| `GET` | `/admin/work-items` | Cross-site work item view, filterable |
| `POST` | `/admin/work-items/:item_id/reassign` | Change handler_agent |
| `POST` | `/admin/work-items/:item_id/force-complete` | Admin override to mark complete |

---

## Implementation Blocks

Ordered by dependency. Each block can be tested independently before moving to the next.

### Block 0 — DB Migration: site_ownership

No Go changes. Prerequisite for everything else.

1. Create `site_ownership` table
2. Backfill: for existing sites, create ownership rows linked to a default admin user (or leave unowned — admin can assign later)

Files changed:
- New migration SQL file

### Block 1 — Core Manager: Site Handlers (Public)

The main implementation work. New handler package in core-manager.

**New files:**
- `internal/core-manager/sites/handlers.go` — main handler struct, site CRUD
- `internal/core-manager/sites/work_items.go` — work item listing + HITL actions
- `internal/core-manager/sites/specs.go` — spec reading
- `internal/core-manager/sites/briefing.go` — briefing state + response
- `internal/core-manager/sites/repository.go` — DB queries, ownership checks

**Handler struct:**
```go
type SiteHandlers struct {
    db            *pgxpool.Pool
    kafkaProducer kafka.Producer
    logger        *zap.Logger
}
```

Uses the same `pgxpool.Pool` (clientsDB) that the admin handlers use — it's the same database that has `sites`, `pages`, `site_work_items`, etc.

**Ownership check helper** — called by every handler before accessing a site:
```go
func (h *SiteHandlers) verifySiteAccess(ctx context.Context, siteID, clientID, userID string) (*Site, error) {
    // JOIN sites + site_ownership, return site or 403/404
}
```

**Route registration** in `server.go`:
```go
siteHandlers := sites.NewSiteHandlers(personaRepoImpl.ClientsDB(), s.kafkaProducer, s.logger)

siteGroup := apiV1.Group("/sites")
siteGroup.Use(middleware.TenantMiddleware(s.logger))
{
    siteGroup.POST("", siteHandlers.HandleCreateSite)
    siteGroup.GET("", siteHandlers.HandleListSites)
    siteGroup.GET("/:site_id", siteHandlers.HandleGetSite)
    siteGroup.PATCH("/:site_id", siteHandlers.HandleUpdateSite)
    siteGroup.GET("/:site_id/status", siteHandlers.HandleGetSiteStatus)
    
    siteGroup.GET("/:site_id/pages", siteHandlers.HandleListPages)
    siteGroup.GET("/:site_id/pages/:page_id", siteHandlers.HandleGetPage)
    siteGroup.PATCH("/:site_id/pages/:page_id", siteHandlers.HandleUpdatePage)
    
    siteGroup.GET("/:site_id/work-items", siteHandlers.HandleListWorkItems)
    siteGroup.GET("/:site_id/work-items/:item_id", siteHandlers.HandleGetWorkItem)
    siteGroup.POST("/:site_id/work-items/:item_id/approve", siteHandlers.HandleApproveWorkItem)
    siteGroup.POST("/:site_id/work-items/:item_id/reject", siteHandlers.HandleRejectWorkItem)
    siteGroup.POST("/:site_id/work-items/:item_id/retry", siteHandlers.HandleRetryWorkItem)
    
    siteGroup.GET("/:site_id/specs", siteHandlers.HandleListSpecs)
    siteGroup.GET("/:site_id/specs/:aspect", siteHandlers.HandleGetSpec)
    
    siteGroup.GET("/:site_id/assets", siteHandlers.HandleListAssets)
    
    siteGroup.GET("/:site_id/briefing", siteHandlers.HandleGetBriefing)
    siteGroup.POST("/:site_id/briefing/respond", siteHandlers.HandleBriefingRespond)
}
```

### Block 2 — Auth Service: Gateway Proxy Lines

Tiny change. Add proxy routes in auth-service's router setup.

**File changed:** `cmd/auth-service/main.go` (or wherever the router is set up — around line 897)

```go
// Site management (tenant-scoped, proxied to core-manager)
gatewayGroup.Any("/sites", gatewayHandler.HandleSiteRoutes)
gatewayGroup.Any("/sites/*path", gatewayHandler.HandleSiteRoutes)
```

Plus the thin handler method in `gateway/handlers.go`:
```go
func (h *HTTPHandler) HandleSiteRoutes(c *gin.Context) {
    h.ProxyToCoreManager(c)
}
```

This is the same pattern used for `/personas/instances`.

### Block 3 — Core Manager: Admin Site Handlers

New admin endpoints. Follows the same pattern as existing admin handlers.

**New files:**
- `internal/core-manager/admin/site_admin_handlers.go`

**Route registration** — added to the existing admin group in `server.go`:
```go
// Site administration
adminGroup.GET("/sites", siteAdminHandlers.HandleListAllSites)
adminGroup.GET("/sites/:site_id", siteAdminHandlers.HandleGetSiteAdmin)
adminGroup.POST("/sites/:site_id/assign", siteAdminHandlers.HandleAssignSite)
adminGroup.POST("/sites/:site_id/trigger-build", siteAdminHandlers.HandleTriggerBuild)
adminGroup.DELETE("/sites/:site_id", siteAdminHandlers.HandleDeleteSite)

// Build queue administration
adminGroup.GET("/build-queue", siteAdminHandlers.HandleListBuildQueue)
adminGroup.POST("/build-queue", siteAdminHandlers.HandleBulkQueueDomains)
adminGroup.POST("/build-queue/seed", siteAdminHandlers.HandleTriggerSeed)

// Work item administration
adminGroup.GET("/work-items", siteAdminHandlers.HandleListAllWorkItems)
adminGroup.POST("/work-items/:item_id/reassign", siteAdminHandlers.HandleReassignWorkItem)
adminGroup.POST("/work-items/:item_id/force-complete", siteAdminHandlers.HandleForceComplete)
```

Auth-service already has `adminGroup.Any("/sites/*path", gatewayHandler.HandleAdminRoutes)` pattern for proxying — we just need to add the line if it's not already there alongside the other admin routes.

### Block 4 — Auth Service: Admin Gateway Lines

Same pattern as Block 2, but for admin routes:

```go
adminGroup.Any("/sites", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/sites/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/build-queue", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/build-queue/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/work-items", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/work-items/*path", gatewayHandler.HandleAdminRoutes)
```

### Block 5 — WebSocket Events

Extend the existing WebSocket infrastructure to emit site-scoped events.

**Approach:** A thin Kafka consumer in core-manager that watches `orchestration.state-changes` and `system.events` topics, filters for site-relevant changes, and pushes events to connected WebSocket clients scoped by site ownership.

This is a later concern — the REST endpoints work independently.

### Block 6 — Briefing HTTP-to-Kafka Bridge

The `POST /sites/:site_id/briefing/respond` handler needs to:

1. Look up the active orchestration for the site (find the briefing-agent's orchestration that's in `AWAITING_RESPONSES`)
2. Find the `awaited_requests` row for the pending briefing question
3. Produce a Kafka message to the agent's responses topic with the HITL response, matching the pattern in `HandleResumeWorkflow`:

```go
responsePayload := map[string]interface{}{
    "answers":  req.Answers,
    "approved": true,
}
payloadBytes, _ := json.Marshal(responsePayload)

headers := map[string]string{
    "correlation_id":         correlationID,
    "orchestration_id":       orchestrationID,
    "request_id":             newRequestID,
    "message_type":           "response",
    "in_response_to_request_id": awaitedRequestID,
    "status":                 "complete",
    "sender_agent_type":      "human",
}

h.kafkaProducer.Produce(ctx, responsesTopic, headers, key, payloadBytes)
```

This is the trickiest part of the implementation because it needs to correctly interact with the orchestration state machine. Can be deferred to after the basic CRUD endpoints work.

---

## Build Order

Dependency-driven. Each step is testable before moving to the next.

| Order | Block | Depends on | Effort | What to test |
|---|---|---|---|---|
| 1 | Block 0: site_ownership migration | Nothing | Half day | Query returns ownership correctly, backfill works |
| 2 | Block 1a: HandleCreateSite + HandleListSites | Block 0 | 1 day | POST creates build_queue row + ownership row. GET returns user's sites. |
| 3 | Block 2: Auth gateway proxy | Block 1a | 1 hour | Request reaches core-manager, headers present |
| 4 | Block 1b: Site detail + status + pages | Block 1a | 1 day | GET returns correct data. Status shows work item counts. |
| 5 | Block 1c: Work items list + detail | Block 1b | Half day | Filtering works. Detail shows spec and result. |
| 6 | Block 1d: Work item approve/reject/retry | Block 1c | Half day | Status transitions work. Dispatch loop picks up approved items. |
| 7 | Block 1e: Specs + assets | Block 1b | Half day | Read from site_specs and assets tables. |
| 8 | Block 3: Admin site handlers | Block 1a | 1 day | Admin can list all sites, assign to users, trigger builds. |
| 9 | Block 4: Admin gateway lines | Block 3 | 1 hour | Admin routes proxied correctly. |
| 10 | Block 6: Briefing bridge | Block 1d | 1-2 days | Question visible via GET, response triggers orchestration continuation. |
| 11 | Block 5: WebSocket events | Block 1c | 1-2 days | Client receives real-time updates on site progress. |

Roughly 7-9 days of focused work. Blocks 1a through 1e + Block 2 get a usable public API. Blocks 3-4 add admin oversight. Blocks 5-6 add real-time and briefing interaction.

---

## Implementation Notes

### Database access pattern

Core-manager already has `clientsDB` and `templatesDB` pools. The `sites`, `pages`, `site_work_items`, `build_queue`, `site_specs`, `assets`, and `orchestration_states` tables are all in the `clientsDB` database. The site handlers use the same pool.

### No new services

Everything goes into the existing core-manager service. No new Kubernetes deployments, no new Docker images. The site handlers are just new route handlers in the existing Gin router.

### No Kafka for basic CRUD

The public API reads from and writes to the database only for most operations. The agent orchestration system picks up changes on its own schedule (dispatch loop checks `site_work_items`, seed action checks `build_queue`). The only place the API touches Kafka is the briefing response bridge (Block 6) and potentially the build trigger (admin).

### Error responses

Follow the existing pattern in core-manager:
```json
{"error": "Site not found"}
{"error": "Access denied"}
{"error": "Work item is not in pending_review status"}
```

### Pagination

Use offset/limit for now (same as existing `WorkflowListRequest`):
```
?limit=20&offset=0&status=triaged&domain=build
```

### build_queue write path

`POST /sites` writes to `build_queue` with `status = 'queued'`. The site record itself is created later by `seed_build_queue` → `upsert_site`. The API needs to handle the transitional state where the queue entry exists but the site record doesn't yet:

- If site already exists for domain → return the existing site with its current status
- If only queue entry exists → return queue status (queued/seeded)
- If neither exists → create queue entry + ownership record

The ownership record initially links to the queue entry's domain. Once the site is created (by seeding), a background step or the seed action itself populates the `site_id` on the ownership row. Alternatively, the `GET /sites` handler joins through domain as a fallback when `site_id` is null.

### HandleTriggerBuild (admin)

Produces a Kafka message to the `system.agent.generic.requests` topic with an inline workflow that spawns and calls `build-dispatch-loop` for the given site_id. Same pattern used by CLI triggers today.

---

## Files Summary

### New files

| File | Block | Purpose |
|---|---|---|
| Migration: `site_ownership` table | 0 | Ownership junction table |
| `internal/core-manager/sites/handlers.go` | 1 | Site CRUD handlers |
| `internal/core-manager/sites/work_items.go` | 1 | Work item handlers |
| `internal/core-manager/sites/specs.go` | 1 | Spec read handlers |
| `internal/core-manager/sites/briefing.go` | 1, 6 | Briefing state + HITL bridge |
| `internal/core-manager/sites/repository.go` | 1 | DB queries, ownership verification |
| `internal/core-manager/admin/site_admin_handlers.go` | 3 | Admin site/queue/work-item management |

### Modified files

| File | Block | Change |
|---|---|---|
| `internal/core-manager/api/server.go` | 1, 3 | Add site route groups to setupRoutes |
| `cmd/auth-service/main.go` (router setup) | 2, 4 | Add gateway proxy lines for `/sites` and admin routes |
| `internal/auth-service/gateway/handlers.go` | 2 | Add HandleSiteRoutes method |

### Not modified

All existing agent definitions, workflows, Go actions, and the agent-chassis codebase remain untouched. The public API is purely additive — it reads from and writes to the same tables the agents already use, without changing how the agents work.


