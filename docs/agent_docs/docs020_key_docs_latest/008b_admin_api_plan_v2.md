# 008 — Admin API: Current State, Issues, and Implementation Plan

Full audit of the existing admin API, known bugs and gaps, and a plan to bring it to a consistent working state alongside the new site-domain endpoints.

---

## 1. Architecture Overview

### Two services, one gateway

**Auth Service** (`:8081`) — handles auth, user management, subscriptions, projects directly. Proxies everything else to core-manager via the gateway.

**Core Manager** (`:8082`) — handles templates, persona instances, and all admin domain logic (clients, system, workflows, agents). Receives proxied requests from auth-service gateway.

### Authentication flow

1. Client sends `Authorization: Bearer <JWT>` to auth-service
2. Auth-service `RequireAuth` middleware validates JWT, sets `user_id`, `client_id`, `user_role`, `user_tier`, `user_email`, `user_permissions` in Gin context
3. For admin routes: `RequireRole("admin")` middleware checks `user_role == "admin"`
4. For proxied routes: gateway copies context values into `X-User-ID`, `X-Client-ID`, `X-User-Role`, `X-User-Tier`, `X-User-Email`, `X-User-Permissions` headers
5. Core-manager `AuthMiddleware` re-validates the JWT itself (shared secret) OR falls back to calling auth-service `/api/v1/auth/validate`
6. Core-manager `AdminOnly()` middleware checks `user_role == "admin"`
7. Core-manager `TenantMiddleware` extracts `client_id` and `user_id` from claims for scoped queries

### Dual auth validation

Core-manager validates JWTs independently using a shared `JWT_SECRET_KEY`. If that fails, it falls back to an HTTP call to auth-service. This means core-manager doesn't depend on auth-service being up for already-issued tokens, but can still validate edge cases.

---

## 2. Current Admin Endpoints — Full Inventory

### Auth Service: Handled Directly

#### Auth (Public — no auth required)

| Method | Path | Handler | Status |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | `authHandlers.HandleRegister` | Working |
| `POST` | `/api/v1/auth/login` | `authHandlers.HandleLogin` | Working |
| `POST` | `/api/v1/auth/refresh` | `authHandlers.HandleRefresh` | Working |
| `POST` | `/api/v1/auth/validate` | `authHandlers.HandleValidate` | Working |
| `POST` | `/api/v1/auth/logout` | `authHandlers.HandleLogout` | Working (stateless — logs only, no token blacklist) |

Request/Response: Registration requires `email`, `password`, `client_id`, optional `first_name`, `last_name`, `company`. Login returns `access_token`, `refresh_token`, `expires_in` (3600s), `user` object with id, email, client_id, role, tier, permissions.

Issues:
- Logout is a no-op (stateless JWT, `TODO: Send verification email` in Register)
- No rate limiting on login attempts
- No email verification flow

#### User (Protected — auth required)

| Method | Path | Handler | Status |
|---|---|---|---|
| `GET` | `/api/v1/user/profile` | `userHandlers.HandleGetCurrentUser` | Working |
| `PUT` | `/api/v1/user/profile` | `userHandlers.HandleUpdateCurrentUser` | Working |
| `POST` | `/api/v1/user/password` | `userHandlers.HandleChangePassword` | Working |
| `DELETE` | `/api/v1/user/delete` | `userHandlers.HandleDeleteAccount` | Working (requires password + "DELETE MY ACCOUNT" confirmation) |

#### Subscription (Protected — auth required)

| Method | Path | Handler | Status |
|---|---|---|---|
| `GET` | `/api/v1/subscription` | `subscriptionHandlers.HandleGetSubscription` | Working |
| `GET` | `/api/v1/subscription/usage` | `subscriptionHandlers.HandleGetUsageStats` | Working |
| `GET` | `/api/v1/subscription/check-quota` | `subscriptionHandlers.HandleCheckQuota` | Working (requires `?resource=` param) |

#### Projects (Protected — auth required)

| Method | Path | Handler | Status |
|---|---|---|---|
| `GET` | `/api/v1/projects` | `projectHandler.ListProjects` | Working |
| `POST` | `/api/v1/projects` | `projectHandler.CreateProject` | Working |
| `GET` | `/api/v1/projects/:id` | `projectHandler.GetProject` | Working |
| `PUT` | `/api/v1/projects/:id` | `projectHandler.UpdateProject` | Working |
| `DELETE` | `/api/v1/projects/:id` | `projectHandler.DeleteProject` | Working |

#### Admin — User Management (Admin only)

| Method | Path | Handler | Status |
|---|---|---|---|
| `GET` | `/api/v1/admin/users` | `adminHandlers.HandleListUsers` | Working (paginated, filterable by email/client_id/role/tier/is_active) |
| `GET` | `/api/v1/admin/users/:user_id` | `adminHandlers.HandleGetUser` | Working (returns user + stats) |
| `PUT` | `/api/v1/admin/users/:user_id` | `adminHandlers.HandleUpdateUser` | Working (role, tier, is_active, email_verified) |
| `DELETE` | `/api/v1/admin/users/:user_id` | `adminHandlers.HandleDeleteUser` | Working (prevents self-deletion) |
| `GET` | `/api/v1/admin/users/:user_id/activity` | `adminHandlers.HandleGetUserActivity` | Working (offset/limit) |
| `POST` | `/api/v1/admin/users/:user_id/permissions` | `adminHandlers.HandleGrantPermission` | Working |
| `DELETE` | `/api/v1/admin/users/:user_id/permissions/:permission_name` | `adminHandlers.HandleRevokePermission` | Working |

#### Admin — Subscription Management (Admin only)

| Method | Path | Handler | Status |
|---|---|---|---|
| `GET` | `/api/v1/admin/subscriptions` | `subscriptionAdminHandlers.HandleListSubscriptions` | Working (paginated, filterable by status/tier) |
| `POST` | `/api/v1/admin/subscriptions` | `subscriptionAdminHandlers.HandleCreateSubscription` | Working |
| `PUT` | `/api/v1/admin/subscriptions/:user_id` | `subscriptionAdminHandlers.HandleUpdateSubscription` | Working |

### Auth Service: Proxied to Core Manager

These routes use `.Any()` wildcards in auth-service. The gateway enriches headers and forwards to core-manager. Core-manager defines the actual method handlers.

#### Auth-service gateway proxy lines (router setup)

```go
// Admin routes → core-manager
adminGroup.Any("/clients", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/clients/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/system/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/workflows/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/agent-definitions/*path", gatewayHandler.HandleAdminRoutes)

// Non-admin routes → core-manager
gatewayGroup.Any("/templates", gatewayHandler.HandleTemplateRoutes)        // admin only (separate middleware)
gatewayGroup.Any("/templates/*path", gatewayHandler.HandleTemplateRoutes)
gatewayGroup.Any("/personas/instances", gatewayHandler.HandleInstanceRoutes)
gatewayGroup.Any("/personas/instances/*path", gatewayHandler.HandleInstanceRoutes)
```

**Missing proxy lines** (core-manager has the routes but auth-service doesn't proxy them):
- `/api/v1/admin/agent-instances/*` — core-manager registers these routes but auth-service only proxies `/agent-definitions/*path`. The `.Any()` on `agent-definitions` won't match `agent-instances`.
- `/api/v1/admin/dashboard` — core-manager's `DashboardHandlers` exists but has no proxy line and isn't registered in core-manager's `setupRoutes` either.
- `/api/v1/admin/system/logs` — `HandleGetSystemLogs` exists but no route registered.
- `/api/v1/admin/sites/*` — site admin handlers written (Block E), proxy lines documented but not yet added.
- `/api/v1/admin/work-items/*` — work item admin handlers written (Block E), proxy lines documented but not yet added.

### Core Manager: Route Registration (server.go setupRoutes)

| Method | Path | Handler | Notes |
|---|---|---|---|
| `GET` | `/health` | `healthHandler.HandleHealth` | No auth |
| `POST` | `/api/v1/agents/bootstrap` | `bootstrapHandler.HandleAgentBootstrap` | Bootstrap key auth, not JWT |
| | | | |
| **Templates (admin only)** | | | |
| `POST` | `/api/v1/templates` | `templateHandler.HandleCreateTemplate` | Working |
| `GET` | `/api/v1/templates` | `templateHandler.HandleListTemplates` | Working |
| `GET` | `/api/v1/templates/:id` | `templateHandler.HandleGetTemplate` | Working |
| `PUT` | `/api/v1/templates/:id` | `templateHandler.HandleUpdateTemplate` | Working |
| `DELETE` | `/api/v1/templates/:id` | `templateHandler.HandleDeleteTemplate` | Working |
| | | | |
| **Persona Instances (tenant-scoped)** | | | |
| `POST` | `/api/v1/personas/instances` | `instanceHandler.HandleCreateInstance` | Working |
| `GET` | `/api/v1/personas/instances` | `instanceHandler.HandleListInstances` | Working |
| `GET` | `/api/v1/personas/instances/:id` | `instanceHandler.HandleGetInstance` | Working |
| `PATCH` | `/api/v1/personas/instances/:id` | `instanceHandler.HandleUpdateInstance` | Working |
| `DELETE` | `/api/v1/personas/instances/:id` | `instanceHandler.HandleDeleteInstance` | Working |
| | | | |
| **Admin — Clients** | | | |
| `POST` | `/api/v1/admin/clients` | `clientHandlers.HandleCreateClient` | Working (creates schema + clients_info row) |
| `GET` | `/api/v1/admin/clients` | `clientHandlers.HandleListClients` | Working (combines clients_info + orphan schemas) |
| `GET` | `/api/v1/admin/clients/:client_id/usage` | `clientHandlers.HandleGetClientUsage` | Working (instances, workflows, memory, fuel) |
| | | | |
| **Admin — System** | | | |
| `GET` | `/api/v1/admin/system/status` | `systemHandlers.HandleGetSystemStatus` | Working (DB health real, Kafka health hardcoded) |
| `GET` | `/api/v1/admin/system/kafka/topics` | `systemHandlers.HandleListKafkaTopics` | **Hardcoded list** — not querying Kafka admin API |
| | | | |
| **Admin — Workflows** | | | |
| `GET` | `/api/v1/admin/workflows` | `systemHandlers.HandleListWorkflows` | Working (queries orchestration_states, filterable) |
| `GET` | `/api/v1/admin/workflows/:correlation_id` | `systemHandlers.HandleGetWorkflow` | Working (full state detail) |
| `POST` | `/api/v1/admin/workflows/:correlation_id/resume` | `systemHandlers.HandleResumeWorkflow` | Working (resume via Kafka or terminate) |
| | | | |
| **Admin — Agent Definitions** | | | |
| `GET` | `/api/v1/admin/agent-definitions` | `systemHandlers.HandleListAgentDefinitions` | Working |
| `POST` | `/api/v1/admin/agent-definitions` | `agentAdminHandlers.HandleCreateAgentDefinition` | Working (async Kafka topic creation) |
| `PUT` | `/api/v1/admin/agent-definitions/:type_name` | `systemHandlers.HandleUpdateAgentDefinition` | Working (partial update) |
| `GET` | `/api/v1/admin/agent-definitions/:type/topics/verify` | `agentAdminHandlers.HandleVerifyAgentTopics` | Working |
| `POST` | `/api/v1/admin/agent-definitions/:type/topics/recreate` | `agentAdminHandlers.HandleRecreateAgentTopics` | Working |
| | | | |
| **Admin — Agent Instances** | | | |
| `GET` | `/api/v1/admin/agent-instances` | `agentAdminHandlers.HandleListAgentInstances` | Working (cross-schema scan) |
| `GET` | `/api/v1/admin/agent-instances/:agent_id` | `agentAdminHandlers.HandleGetAgentInstance` | Working (usage + health + executions) |
| `PUT` | `/api/v1/admin/agent-instances/:agent_id/status` | `agentAdminHandlers.HandleToggleAgentStatus` | Working |
| `POST` | `/api/v1/admin/agent-instances/:agent_id/restart` | `agentAdminHandlers.HandleRestartAgent` | Working (Kafka command) |
| `PUT` | `/api/v1/admin/clients/:client_id/instances/:instance_id/config` | `agentAdminHandlers.HandleUpdateInstanceConfig` | Working |

### WebSocket

| Method | Path | Handler | Notes |
|---|---|---|---|
| `GET` | `/ws` | `gatewayHandler.HandleWebSocket` | Auth required, proxied to core-manager |

The gateway establishes a WebSocket to core-manager and bridges messages bidirectionally.

---

## 3. Known Issues

### Bugs (code won't work as written)

| # | Location | Issue | Severity |
|---|---|---|---|
| B1 | `dashboard_handlers.go` — `getUserMetrics` | Uses MySQL syntax: `CURDATE()`, `INTERVAL 7 DAY`, `INTERVAL 30 DAY`, `DATE_SUB()`. PostgreSQL requires `CURRENT_DATE`, `INTERVAL '7 days'`, `NOW() - INTERVAL '1 month'`. | High — will error on every dashboard call |
| B2 | `system_handlers.go` — `updateWorkflowStatus` | Queries `orchestrator_state` (singular) but table is `orchestration_states` (plural). | High — terminate action will fail |
| B3 | `dashboard_handlers.go` — `getAgentMetrics` | The "most used agents" query joins `orchestration_states` to `agent_definitions` with `ON true` (cartesian product). This returns nonsense data. | Medium — returns wrong metrics |
| B4 | Auth-service gateway | No proxy line for `/api/v1/admin/agent-instances/*`. Core-manager registers these routes but requests from auth-service will 404. | High — agent instance admin endpoints unreachable via auth-service gateway |

### Hardcoded / Mock Data

| # | Location | Issue |
|---|---|---|
| H1 | `agent_handlers.go` — `getAgentHealth` | Returns hardcoded values (status: "healthy", error_rate: 2.5%, response_time: 145.3ms, queue_depth: 3) |
| H2 | `system_handlers.go` — `getKafkaStatus` | Returns hardcoded values (status: "healthy", broker_count: 3, topic_count: 20) |
| H3 | `system_handlers.go` — `HandleListKafkaTopics` | Returns a hardcoded list of topic names instead of querying Kafka admin API |
| H4 | `dashboard_handlers.go` — `getUsageMetrics` | All values are hardcoded mock data (fuel: 125000, API calls: 8543, storage: 45.7GB) |
| H5 | `dashboard_handlers.go` — `getSystemHealth` | Kafka status, average latency, error rate, queue depth are all hardcoded |
| H6 | `dashboard_handlers.go` — `getAgentMetrics` | Average response time is hardcoded (245.7ms) |
| H7 | `dashboard_handlers.go` — `HandleGetSystemLogs` | Returns a single hardcoded mock log entry |

### Missing Registration / Wiring

| # | Issue |
|---|---|
| M1 | `DashboardHandlers` exists but `HandleGetDashboard` is never registered in `setupRoutes` |
| M2 | `HandleGetSystemLogs` exists but no route registered |
| M3 | No auth-service proxy lines for `agent-instances` (only `agent-definitions` proxied) |
| M4 | No auth-service proxy lines for `dashboard` |
| M5 | `DashboardHandlers` constructor requires `authDB *sql.DB` but core-manager server doesn't have an auth DB connection — this struct can't be constructed as-is |

### Design Concerns

| # | Issue |
|---|---|
| D1 | `clientHandlers.storeClientInfo` runs `CREATE TABLE IF NOT EXISTS` on every call. Should be a migration. |
| D2 | `clientHandlers.listClients` calls `storeClientInfo` with an empty request to ensure the table exists. |
| D3 | Agent instance listing scans all `client_*` schemas with individual queries — O(N) schemas × O(M) instances per schema. No pagination. |
| D4 | `findClientForAgent` also scans all schemas sequentially. |
| D5 | `HandleCreateAgentDefinition` only writes to `agent_definitions` — doesn't write the full schema (workflow, topics, contracts, etc). The admin SQL files in the project are much more comprehensive. |
| D6 | SQL injection surface: `fmt.Sprintf` for schema names in client/agent queries. Validation exists (`isValidClientID`, `isValidAgentType`) but pattern is fragile. |

---

## 4. Implementation Plan

### Block A — Fix Bugs in Existing Code

No new endpoints. Fix what's broken.

**A1: Dashboard MySQL → PostgreSQL syntax**

File: `internal/core-manager/admin/dashboard_handlers.go`

```go
// BEFORE (MySQL)
WHERE DATE(created_at) = CURDATE()
WHERE created_at > NOW() - INTERVAL 7 DAY
SELECT COUNT(*) FROM users WHERE created_at < DATE_SUB(NOW(), INTERVAL 1 MONTH)

// AFTER (PostgreSQL)
WHERE created_at::date = CURRENT_DATE
WHERE created_at > NOW() - INTERVAL '7 days'
SELECT COUNT(*) FROM users WHERE created_at < NOW() - INTERVAL '1 month'
```

Also: the `INTERVAL 30 DAY` → `INTERVAL '30 days'` in `getOverviewMetrics`.

**A2: Fix table name in updateWorkflowStatus**

File: `internal/core-manager/admin/system_handlers.go`

```go
// BEFORE
UPDATE orchestrator_state SET ...
// AFTER
UPDATE orchestration_states SET ...
```

**A3: Fix agent metrics cartesian join**

File: `internal/core-manager/admin/dashboard_handlers.go`

The query needs to join on `agent_type` from the orchestration's initial request data or use a different approach entirely. Simplest fix for now: query `orchestration_states` grouped by the agent_type stored in the state's `initial_request_data`, or just count from `agent_definitions` usage_count column.

**A4: Add missing gateway proxy line for agent-instances**

File: `cmd/auth-service/main.go` (router setup, admin group)

```go
adminGroup.Any("/agent-instances", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/agent-instances/*path", gatewayHandler.HandleAdminRoutes)
```

**Effort:** Half day. Test by calling each fixed endpoint.

### Block B — Wire Up Unregistered Handlers

Handlers exist but aren't reachable.

**B1: Register dashboard route in core-manager**

File: `internal/core-manager/api/server.go`

The `DashboardHandlers` constructor requires `authDB *sql.DB`. Two options:

- **Option 1** (simpler): Remove auth DB dependency. Dashboard queries about users go through a Kafka request to auth-service, or core-manager gets a read-only connection to the auth DB.
- **Option 2** (pragmatic): Add an auth DB read-only connection to core-manager config. Core-manager already knows the auth-service URL; adding a DB connection string is reasonable for aggregated metrics.

Once wired:
```go
dashboardHandlers := admin.NewDashboardHandlers(
    personaRepoImpl.ClientsDB(),
    personaRepoImpl.TemplatesDB(),
    authDBConn,  // new
    s.logger,
)

adminGroup.GET("/dashboard", dashboardHandlers.HandleGetDashboard)
adminGroup.GET("/system/logs", dashboardHandlers.HandleGetSystemLogs)
```

**B2: Add gateway proxy lines for dashboard and logs**

File: `cmd/auth-service/main.go`

```go
adminGroup.Any("/dashboard", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/dashboard/*path", gatewayHandler.HandleAdminRoutes)
```

The `system/logs` route is already covered by the existing `adminGroup.Any("/system/*path", ...)` wildcard.

**Effort:** 1 day (mostly the auth DB connection decision and wiring).

### Block C — Replace Hardcoded Data

Progressive — can be done endpoint by endpoint. Ordered by value.

**C1: Kafka topic listing — use admin API**

Replace the hardcoded topic list in `HandleListKafkaTopics` with actual Kafka admin client query. The `TopicManager` already exists in `platform/kafka` and has `TopicExists()` — extend it with `ListTopics()`.

**C2: Kafka health — real broker check**

Replace hardcoded `getKafkaStatus` with actual broker metadata query. The Kafka admin client can return broker count and cluster status.

**C3: Agent health — query from orchestration data**

Replace hardcoded `getAgentHealth` with real metrics derived from `orchestration_states`:
- Error rate: count failed vs total orchestrations for that agent type in the last hour
- Response time: average duration of completed orchestrations
- Queue depth: count of running/awaiting orchestrations

**C4: Dashboard usage — real aggregation**

Replace mock fuel consumption, API calls, storage with actual queries:
- Fuel: aggregate from `usage_analytics` across client schemas
- Storage: query PostgreSQL `pg_database_size` + any file storage metrics
- API calls: would need request logging or a counter (defer if no metrics infrastructure)

**C5: System logs — real log source**

Replace mock log entry with actual log retrieval. Options: query from PostgreSQL (if logs are stored there), or integrate with whatever centralized logging exists (Loki, ELK, etc). If no centralized logging, this endpoint can be deferred.

**Effort:** 2-3 days total, spread across sprints. C1-C3 are most valuable.

### Block D — Performance Improvements

**D1: Client/agent schema scanning**

The pattern of iterating all `client_*` schemas for every agent instance query is expensive. Options:

- **Short term:** Add a `client_agent_index` table in the public schema that maps agent_instance_id → client_id. Updated on instance create/delete.
- **Long term:** Move agent instances to the public schema with a client_id column, or use a view.

**D2: Replace CREATE TABLE IF NOT EXISTS in handlers**

Move `clients_info` table creation to a proper migration. Remove the defensive creation from `storeClientInfo` and `listClients`.

**D3: Add pagination to agent instance listing**

Add `limit`, `offset`, `client_id` query params to `HandleListAgentInstances` and stop the full-schema scan when a specific client_id is provided.

**Effort:** 1-2 days. D1 is the most impactful.

### Block E — New Admin Endpoints (Site Domain)

Admin site management endpoints. The HITL review workflow subset (E1 partial + E3 expanded) is implemented. Build queue and site lifecycle management are not yet built.

**E1: Site administration — PARTIALLY IMPLEMENTED**

File: `internal/core-manager/admin/site_admin_handlers.go`

| Method | Path | Handler | Purpose | Status |
|---|---|---|---|---|
| `GET` | `/admin/sites` | `HandleListSites` | Sites with work item count breakdown | **Implemented** |
| `GET` | `/admin/sites/:site_id` | `HandleGetSite` | Site detail with all current specs | **Implemented** |
| `PATCH` | `/admin/sites/:site_id/specs/:aspect` | `HandleUpdateSiteSpec` | Supersede spec (versioned, preserves history) | **Implemented** |
| `POST` | `/admin/sites/:site_id/assign` | `HandleAssignSite` | Create/update site_ownership row | Not started (needs site_ownership table) |
| `POST` | `/admin/sites/:site_id/trigger-build` | `HandleTriggerBuild` | Produce Kafka message to start dispatch loop | Not started |
| `DELETE` | `/admin/sites/:site_id` | `HandleDeleteSite` | Soft delete | Not started |

**E2: Build queue administration — NOT STARTED**

| Method | Path | Handler | Purpose |
|---|---|---|---|
| `GET` | `/admin/build-queue` | `HandleListBuildQueue` | List entries with status, filterable |
| `POST` | `/admin/build-queue` | `HandleBulkQueueDomains` | Insert multiple domains at once (batch_id grouping) |
| `POST` | `/admin/build-queue/seed` | `HandleTriggerSeed` | Manually invoke seed_build_queue via Kafka |

**E3: Work item administration + HITL review — IMPLEMENTED**

| Method | Path | Handler | Purpose | Status |
|---|---|---|---|---|
| `GET` | `/admin/work-items` | `HandleListWorkItems` | Cross-site, filterable (?status, ?domain, ?site_id, ?item_type) | **Implemented** |
| `GET` | `/admin/work-items/:item_id` | `HandleGetWorkItem` | Full detail with spec | **Implemented** |
| `PATCH` | `/admin/work-items/:item_id` | `HandleUpdateWorkItem` | Update status, spec, handler_agent, item_type | **Implemented** |
| `POST` | `/admin/work-items/:item_id/retry` | `HandleRetryWorkItem` | Convert to content_rewrite, reset attempts, re-triage | **Implemented** |
| `POST` | `/admin/work-items/:item_id/resolve` | `HandleResolveWorkItem` | Mark complete with resolution note | **Implemented** |

The retry endpoint converts any `needs_human_review`, `failed`, or `blocked` item to `content_rewrite` with `page-build-handler`, resets `attempt_count = 0`, and sets `status = 'triaged'`. The dispatch loop picks it up on the next cycle.

The resolve endpoint marks the item `complete` with a `resolution` note stored in the `error` column. Used when the section should stay hidden (awaiting real data from client).

The PATCH endpoint allows updating individual fields (status, spec, handler_agent, item_type). When status is changed to `triaged`, it also resets `attempt_count`, `claimed_by`, and `claimed_at`.

**HITL review workflow through the API:**

```
1. GET  /admin/work-items?status=needs_human_review
   → See which pages have placeholder content

2. GET  /admin/sites/:site_id
   → See the identity spec (company name, email, team data)

3. PATCH /admin/sites/:site_id/specs/identity
   → Add missing data (e.g. leadership_team array)

4. POST /admin/work-items/:item_id/retry
   → Page gets regenerated with the new data
```

**Route registration** in core-manager `setupRoutes`:
```go
siteAdminHandlers := admin.NewSiteAdminHandlers(s.clientsDB, s.logger)

siteAdmin := adminGroup.Group("/sites")
{
    siteAdmin.GET("", siteAdminHandlers.HandleListSites)
    siteAdmin.GET("/:site_id", siteAdminHandlers.HandleGetSite)
    siteAdmin.PATCH("/:site_id/specs/:aspect", siteAdminHandlers.HandleUpdateSiteSpec)
}

workItemAdmin := adminGroup.Group("/work-items")
{
    workItemAdmin.GET("", siteAdminHandlers.HandleListWorkItems)
    workItemAdmin.GET("/:item_id", siteAdminHandlers.HandleGetWorkItem)
    workItemAdmin.PATCH("/:item_id", siteAdminHandlers.HandleUpdateWorkItem)
    workItemAdmin.POST("/:item_id/retry", siteAdminHandlers.HandleRetryWorkItem)
    workItemAdmin.POST("/:item_id/resolve", siteAdminHandlers.HandleResolveWorkItem)
}
```

**Gateway proxy lines** in auth-service:
```go
adminGroup.Any("/sites", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/sites/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/work-items", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/work-items/*path", gatewayHandler.HandleAdminRoutes)
```

**Effort remaining:** E2 (build queue) + E1 remainder (assign, trigger-build, delete) = ~1 day.
E3 is complete.

### Block F — Agent Definition Admin Improvements

The current `HandleCreateAgentDefinition` only writes basic fields (type, display_name, description, category, default_config). The actual agent definitions in the SQL files have many more fields: capabilities, image_repository, image_tag, resources, topics, health_config, env_vars, delegation_preferences, domain_tags, input_contract, output_contract, briefing_questionnaire, workflow.

**F1: Full agent definition CRUD**

Extend `AgentDefinitionRequest` to accept all fields:
```go
type AgentDefinitionRequest struct {
    Type                  string                 `json:"type" binding:"required"`
    DisplayName           string                 `json:"display_name" binding:"required"`
    Description           string                 `json:"description"`
    Category              string                 `json:"category" binding:"required"`
    DefaultConfig         map[string]interface{} `json:"default_config"`
    Capabilities          map[string]interface{} `json:"capabilities"`
    ImageRepository       string                 `json:"image_repository"`
    ImageTag              string                 `json:"image_tag"`
    Resources             map[string]interface{} `json:"resources"`
    Topics                map[string]interface{} `json:"topics"`
    HealthConfig          map[string]interface{} `json:"health_config"`
    EnvVars               map[string]interface{} `json:"env_vars"`
    DelegationPreferences map[string]interface{} `json:"delegation_preferences"`
    DomainTags            []string               `json:"domain_tags"`
    InputContract         map[string]interface{} `json:"input_contract"`
    OutputContract        map[string]interface{} `json:"output_contract"`
    BriefingQuestionnaire map[string]interface{} `json:"briefing_questionnaire"`
}
```

**F2: GET single agent definition**

Currently missing — you can list all or update by type, but can't GET a single definition. Add:
```go
adminGroup.GET("/agent-definitions/:type_name", systemHandlers.HandleGetAgentDefinition)
```

**F3: DELETE (soft) agent definition**

Currently missing:
```go
adminGroup.DELETE("/agent-definitions/:type_name", systemHandlers.HandleDeleteAgentDefinition)
```

Sets `deleted_at = NOW()` and `is_active = false`.

**Effort:** 1 day.

---

## 5. Build Order

| Order | Block | Depends on | Effort | Status |
|---|---|---|---|---|
| 1 | **A: Fix bugs** | Nothing | Half day | Not started |
| 2 | **B: Wire unregistered handlers** | A (optional) | 1 day | Not started |
| 3 | **D2: Migration for clients_info** | Nothing | 1 hour | Not started |
| 4 | **C1-C3: Real Kafka + agent health** | Nothing | 1.5 days | Not started |
| 5 | **E: Site admin endpoints** | Nothing (site_ownership no longer required for admin) | 1.5 days | **E1 partial + E3: IMPLEMENTED** |
| 6 | **F: Agent definition improvements** | A (optional) | 1 day | Not started |
| 7 | **D1, D3: Performance** | Nothing | 1-2 days | Not started |
| 8 | **C4-C5: Dashboard real data** | B | 1 day | Not started |

Block E was prioritised ahead of Blocks A-D because the HITL review workflow was needed immediately — placeholder content was appearing on live sites. The admin site endpoints (E1 partial + E3) work without the site_ownership migration because they're admin-only (no user scoping needed).

Block A should still be done next — the MySQL syntax bugs and wrong table name in existing admin endpoints remain broken.

---

## 6. Files Summary

### Modified files

| File | Block | Changes | Status |
|---|---|---|---|
| `internal/core-manager/admin/dashboard_handlers.go` | A1, A3 | MySQL → PostgreSQL syntax, fix cartesian join | Not started |
| `internal/core-manager/admin/system_handlers.go` | A2 | Fix table name `orchestrator_state` → `orchestration_states` | Not started |
| `cmd/auth-service/main.go` (router setup) | A4, B2, E | Add proxy lines for agent-instances, dashboard, sites, work-items | **E portion ready** |
| `internal/core-manager/api/server.go` | B1, E, F | Register dashboard, system logs, site admin routes | **E portion ready** |
| `platform/kafka/topic_manager.go` | C1, C2 | Add `ListTopics()` and `GetClusterMetadata()` methods | Not started |
| `internal/core-manager/admin/agent_handlers.go` | C3, F1 | Replace hardcoded health with real queries | Not started |

### New files

| File | Block | Purpose | Status |
|---|---|---|---|
| `internal/core-manager/admin/site_admin_handlers.go` | E | Site admin + HITL review endpoints | **Written** |
| `site-admin-dashboard.jsx` | E | React admin UI for review workflow | **Written** (mock data) |
| `patch_site_admin_routes.go` | E | Route + gateway wiring documentation | **Written** |
| Migration: `clients_info` table | D2 | Move table creation from handler to migration | Not started |

### Not modified

All agent definitions, workflows, Go actions, and the agent-chassis codebase remain untouched. The admin API changes are additive or fix existing code — no changes to how agents operate.

---

## 7. Combined Route Map (Target State)

After all blocks, the complete admin API:

### Auth-service gateway proxy lines (target)

```go
// Currently proxied
adminGroup.Any("/clients", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/clients/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/system/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/workflows/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/agent-definitions/*path", gatewayHandler.HandleAdminRoutes)

// Missing — to add
adminGroup.Any("/agent-instances", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/agent-instances/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/dashboard", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/sites", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/sites/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/build-queue", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/build-queue/*path", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/work-items", gatewayHandler.HandleAdminRoutes)
adminGroup.Any("/work-items/*path", gatewayHandler.HandleAdminRoutes)
```

### Core-manager admin routes (target)

```
Clients:
  POST   /admin/clients
  GET    /admin/clients
  GET    /admin/clients/:client_id/usage

System:
  GET    /admin/system/status
  GET    /admin/system/kafka/topics
  GET    /admin/system/logs                    ← Block B

Dashboard:
  GET    /admin/dashboard                      ← Block B

Workflows:
  GET    /admin/workflows
  GET    /admin/workflows/:correlation_id
  POST   /admin/workflows/:correlation_id/resume

Agent Definitions:
  GET    /admin/agent-definitions
  POST   /admin/agent-definitions
  GET    /admin/agent-definitions/:type_name   ← Block F
  PUT    /admin/agent-definitions/:type_name
  DELETE /admin/agent-definitions/:type_name   ← Block F
  GET    /admin/agent-definitions/:type/topics/verify
  POST   /admin/agent-definitions/:type/topics/recreate

Agent Instances:
  GET    /admin/agent-instances
  GET    /admin/agent-instances/:agent_id
  PUT    /admin/agent-instances/:agent_id/status
  POST   /admin/agent-instances/:agent_id/restart
  PUT    /admin/clients/:client_id/instances/:instance_id/config

Sites:                                         ← Block E (PARTIAL)
  GET    /admin/sites                            ✓ Implemented
  GET    /admin/sites/:site_id                   ✓ Implemented
  PATCH  /admin/sites/:site_id/specs/:aspect     ✓ Implemented (HITL spec editing)
  POST   /admin/sites/:site_id/assign            ○ Planned
  POST   /admin/sites/:site_id/trigger-build     ○ Planned
  DELETE /admin/sites/:site_id                   ○ Planned

Build Queue:                                   ← Block E (NOT STARTED)
  GET    /admin/build-queue                      ○ Planned
  POST   /admin/build-queue                      ○ Planned
  POST   /admin/build-queue/seed                 ○ Planned

Work Items:                                    ← Block E (IMPLEMENTED)
  GET    /admin/work-items                       ✓ Implemented (filterable)
  GET    /admin/work-items/:item_id              ✓ Implemented
  PATCH  /admin/work-items/:item_id              ✓ Implemented (update fields)
  POST   /admin/work-items/:item_id/retry        ✓ Implemented (HITL retry)
  POST   /admin/work-items/:item_id/resolve      ✓ Implemented (HITL dismiss)
```

