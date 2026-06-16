# 019 — Admin Dashboard & API Gateway

How the admin dashboard, API gateway, and HITL checkpoint system work together.

---

## Architecture

The admin dashboard is a React SPA served by an nginx container that also acts as an API gateway. This keeps both core-manager and auth-service as internal ClusterIP services with no external exposure.

```
Browser
  │
  ▼
┌──────────────────────────────┐
│  admin-dashboard (nginx)     │  ← serves SPA static files
│  port 8080                   │  ← proxies API calls
│                              │
│  /                → SPA      │
│  /api/v1/auth/*   → auth-service:8081
│  /api/v1/*        → core-manager:8088
│  /health          → 200 OK   │
└──────────────────────────────┘
         │                │
         ▼                ▼
   auth-service     core-manager
   (ClusterIP)      (ClusterIP)
     :8081             :8088
```

The nginx ordering matters: `/api/v1/auth/` is matched before `/api/v1/`, so login and register requests go to auth-service while all other API calls go to core-manager.

---

## File Layout

```
frontends/admin-dashboard/
├── Dockerfile              # Multi-stage: Node builds SPA, nginx serves it
├── nginx.conf              # Gateway routing + rate limiting + security headers
├── package.json            # React 18, Vite 6, TypeScript
├── vite.config.js          # Dev server config with API proxy
├── tsconfig.json           # TypeScript config (relaxed, allowJs)
├── index.html              # SPA entry point
├── env.example             # Local dev reference
└── src/
    ├── main.tsx            # React mount point
    ├── App.tsx             # Dashboard application (login, sites, work items, approvals)
    ├── components/         # Shared components (future)
    └── pages/              # Page components (future)

deployments/kustomize/services/admin-dashboard/
├── base/
│   ├── kustomization.yaml
│   ├── deployment.yaml     # 2 replicas, 32Mi memory, health probes
│   └── service.yaml        # ClusterIP on port 8080
└── overlays/production/uk_001/
    └── kustomization.yaml  # Image tag override
```

---

## Dashboard Views

**Login** — posts credentials to `/api/v1/auth/login` (proxied to auth-service). Validates that the user has `role: admin`. Stores the JWT in sessionStorage.

**Sites Overview** — fetches `/api/v1/admin/sites`, displays cards for each site with work item counts: review, failed, active, ready, done. Click a site to drill into its work items.

**Work Items** — fetches `/api/v1/admin/work-items` with filters for status, site, and item type. Split-pane layout: item list on the left, detail panel on the right when selected. Checkpoint items are tagged with a purple "checkpoint" badge in the list.

**Standard Item Actions:**

| Action | What it does | Available when |
|--------|-------------|----------------|
| Retry | Resets to `triaged`, clears error and attempts | `needs_human_review`, `failed`, `blocked` |
| Resolve | Marks `complete` with optional resolution note | `needs_human_review`, `failed`, `blocked` |
| Dismiss | Marks `complete` as "Dismissed by admin" | `triaged` |
| Unblock | Sets status to `triaged` | `blocked` |

**Checkpoint Item Actions:**

| Action | What it does | Available when |
|--------|-------------|----------------|
| Approve & Continue | Updates spec, creates follow-on work item | `needs_human_review` (checkpoint items) |
| Reject / Skip | Marks complete without creating follow-on | `needs_human_review` (checkpoint items) |

Checkpoint items show an editable form for the review data. Field types are detected automatically: long strings become textareas, arrays of objects use JSON editors, booleans become checkboxes, and short strings become text inputs.

---

## Checkpoint Pattern

### The Problem

Previously, agents needing human input used `request_human_input` which suspended the running orchestration. This created fragile long-lived orchestrations waiting on humans, manual Kafka message construction (kcat) with exact headers, no audit trail of human decisions, and timeout management issues.

### The Solution: checkpoint_for_review

Any agent can add a `checkpoint_for_review` step to its workflow. The agent saves its work to the database and creates a review work item, then completes normally. No suspended orchestrations.

```
Agent workflow:
  step_1 → step_2 → checkpoint_for_review → complete
                            │
                     saves to site_specs (versioned)
                     creates work item (needs_human_review)
                            │
                      [admin reviews in dashboard]
                            │
                     POST /work-items/:id/approve
                       ├── updates site_specs with corrections
                       └── creates follow-on work item
                            │
                     dispatch loop picks up → next agent
```

### Workflow Configuration

```json
"request_review": {
    "action": "checkpoint_for_review",
    "config": {
        "item_type": "needs_brief_review",
        "summary_from": "Review brief for {{.site_record.domain}}",
        "severity": "high",
        "review_fields_from": "generated_brief",
        "save_spec_aspect": "reviewed_brief",
        "site_id_from": "site_record.site_id",
        "page_id_from": "current_page.id",
        "on_approve": {
            "item_type": "ready_to_build",
            "handler_agent": "pageflow-builder",
            "include_fields": ["reviewed_brief", "site_record"]
        }
    },
    "next_step": "complete"
}
```

### Config Reference

| Field | Required | Description |
|-------|----------|-------------|
| `site_id_from` | Yes | Path in collected_data to site_id. Defaults to `site_record.site_id` |
| `item_type` | No | Work item type. Defaults to `needs_human_review` |
| `summary_from` | No | Template string for summary. Supports `{{.path}}` syntax |
| `severity` | No | `high`, `medium`, or `low`. Defaults to `medium` |
| `review_fields_from` | No | Path to the data to be reviewed. Falls back to all collected_data |
| `save_spec_aspect` | No | If set, saves review data to `site_specs` with this aspect name |
| `page_id_from` | No | Path to page_id, if this checkpoint relates to a specific page |
| `on_approve` | No | Instructions for what happens after admin approval |
| `on_approve.item_type` | No | Work item type for the follow-on item |
| `on_approve.handler_agent` | No | Which agent handles the follow-on item |
| `on_approve.include_fields` | No | Fields from spec to copy into the follow-on item |

### What the Action Does (checkpoint_for_review_action.go)

1. Extracts `site_id` from collected_data at the configured path
2. Extracts review data from the configured `review_fields_from` path
3. If `save_spec_aspect` is set, saves the review data to `site_specs` (versioned — previous version marked `is_current = false`)
4. Creates a `site_work_item` with `status = needs_human_review` and `handler_agent = human-review`
5. Embeds `on_approve` instructions in the work item spec
6. Returns success — the workflow continues to the next step (usually `complete`)

### Approval Flow (HandleApproveWorkItem)

When the admin approves a checkpoint item via `POST /work-items/:id/approve`:

1. Validates the item is a checkpoint (`spec.checkpoint = true`) and status is `needs_human_review`
2. If `spec_aspect` is set: updates `site_specs` with the admin's corrected `review_data` (versioned)
3. If `on_approve` is set: creates a follow-on work item with the configured `item_type` and `handler_agent`
4. Marks the review work item complete with audit trail (`approved_by: admin`)

### Example: Briefing Agent

```
briefing-agent workflow:
  research_domain
    → classify_site_type
    → generate_questionnaire
    → fill_brief_with_llm
    → checkpoint_for_review          ← saves brief, creates review item
    → complete

Admin reviews in dashboard:
  - Corrects company name, services, tagline
  - Clicks "Approve & Continue"
  - Spec 'reviewed_brief' updated, "ready_to_build" item created

dispatch loop picks up "ready_to_build":
  → pageflow-builder reads approved spec → builds site pages
```

### Example: Content Quality Audit

```
content-quality-auditor workflow:
  load_page → audit_content → checkpoint_for_review → complete

Config:
  review_fields_from: "audit_result"
  save_spec_aspect: "content_audit"
  on_approve:
    item_type: "content_rewrite"
    handler_agent: "page-build-handler"
```

---

## API Endpoints

All admin endpoints require `Authorization: Bearer <JWT>` with `role: admin`.

### Sites

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/admin/sites` | List sites with work item counts |
| GET | `/api/v1/admin/sites/:id` | Site detail with specs |
| PATCH | `/api/v1/admin/sites/:id` | Update site fields (email, phone, company_name, tagline, contact_address, logo_text) |
| PATCH | `/api/v1/admin/sites/:id/specs/:aspect` | Update a site spec (versioned) |

### Work Items

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/admin/work-items` | Create a work item |
| GET | `/api/v1/admin/work-items` | List work items (filterable) |
| GET | `/api/v1/admin/work-items/:id` | Work item detail |
| PATCH | `/api/v1/admin/work-items/:id` | Update status, spec, handler |
| POST | `/api/v1/admin/work-items/:id/retry` | Reset to triaged for retry |
| POST | `/api/v1/admin/work-items/:id/resolve` | Mark complete with resolution |
| POST | `/api/v1/admin/work-items/:id/approve` | Approve checkpoint item (updates spec, creates follow-on) |

Query parameters for `GET /work-items`: `status`, `site_id`, `item_type`, `domain` (default: `build`).

### Auth (proxied to auth-service)

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/auth/login` | Get JWT token |
| POST | `/api/v1/auth/register` | Register new user |

---

## Nginx Gateway Details

**Rate limiting:**
- Auth endpoints (`/api/v1/auth/`): 10 requests/minute per IP, burst 5
- API endpoints (`/api/v1/`): 60 requests/minute per IP, burst 20

**Timeouts:**
- Auth proxy: 5s connect, 15s read
- API proxy: 5s connect, 120s read (longer for LLM-backed operations)

**Static asset caching:** JS, CSS, images, fonts get `Cache-Control: public, immutable` with 7-day expiry. Vite's content-hashed filenames ensure cache busting on deploys.

**Security headers:** X-Frame-Options DENY, X-Content-Type-Options nosniff, X-XSS-Protection, strict Referrer-Policy.

**SPA routing:** `try_files $uri $uri/ /index.html` ensures client-side routing works.

---

## Local Development

```bash
# Terminal 1: port-forward core-manager
kubectl -n ai-persona-system port-forward svc/core-manager 8088:8088

# Terminal 2: Vite dev server with hot reload
cd frontends/admin-dashboard
npm install
npm run dev
# Opens http://localhost:5173, proxies /api to localhost:8088
```

---

## Building and Deploying

```bash
# Build
make build-dashboard DASHBOARD_TAG=v1.0.2

# Push
make push-dashboard DASHBOARD_TAG=v1.0.2

# Deploy
make deploy-dashboard

# Full release
make release-dashboard DASHBOARD_TAG=v1.0.2

# Verify
make dashboard-port-forward
# Open http://localhost:8080
```

---

## Admin User Setup

```bash
# Register (from inside cluster)
kubectl -n ai-persona-system exec -it $(kubectl get pod -l app=auth-service \
  -o jsonpath='{.items[0].metadata.name}' -n ai-persona-system) -- wget -qO- \
  http://localhost:8081/api/v1/auth/register \
  --post-data='{"email":"admin@example.com","password":"YourPassword","client_id":"demo_client"}' \
  --header='Content-Type: application/json' 2>&1

# Promote to admin (MySQL — auth DB on cPanel)
mysql -h rs17.uk-noc.com -u catalogu_personae -p"PASSWORD" --skip-ssl catalogu_vectordb_chassis \
  -e "UPDATE users SET role = 'admin', subscription_tier = 'enterprise' WHERE email = 'admin@example.com';"
```

Note: the auth service's MySQL connection goes stale if idle. If login fails after the pod has been running a while, restart it: `kubectl -n ai-persona-system delete pod -l app=auth-service`.

---

## Relationship to Core-Manager

Core-manager is a pure API service. It does not serve static files or proxy auth requests — both responsibilities belong to the nginx gateway.

- `server.go` has no `StaticFS` or `setupAuthProxy` calls
- `auth_proxy.go` is not needed
- The nginx gateway is the only service that needs to know about both auth-service and core-manager

---

## Go Files Reference

| File | Location | Purpose |
|------|----------|---------|
| `checkpoint_for_review_action.go` | `platform/orchestration/actions/` | Generic checkpoint action for any agent workflow |
| `site_admin_handlers.go` | `internal/core-manager/admin/` | Site and work item API handlers including approve |
| `server.go` | `internal/core-manager/api/` | Route registration |

---

## Troubleshooting

**Login returns 502:** auth-service is down. Check `kubectl -n ai-persona-system get pods | grep auth`.

**Login returns 401 with correct credentials:** auth-service MySQL connection is stale (cPanel kills idle connections). Restart the auth pod.

**API calls return 502:** core-manager is down. Check pod status and logs.

**SPA blank page:** Vite build may have failed. Check Docker build output.

**Token expires after 1 hour:** default JWT TTL. Dashboard redirects to login on 401. Future improvement: use refresh token.

**Checkpoint item shows no editable form:** the source agent's `review_fields_from` path didn't resolve, so `review_data` is null in the spec. Check the agent's collected_data at the time the checkpoint was created.

**Approve fails with "not a checkpoint":** the work item's spec doesn't have `checkpoint: true`. Only items created by `checkpoint_for_review` action can be approved — other items use retry or resolve.

