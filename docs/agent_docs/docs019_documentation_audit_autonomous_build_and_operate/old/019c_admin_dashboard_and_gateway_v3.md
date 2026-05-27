# 019 — Admin Dashboard & API Gateway

How the admin dashboard, API gateway, and HITL review system work together.

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
    ├── App.tsx             # Dashboard application
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

## Dashboard Navigation

The top bar has two nav tabs:

- **Sites** — overview cards for each site, click to drill into that site's work items
- **All Items** — work items across all sites in one view, with domain badges to identify which site each belongs to

Both views share the same work items UI with filtering, detail panels, and actions.

---

## Dashboard Views

**Login** — posts credentials to `/api/v1/auth/login` (proxied to auth-service). Validates that the user has `role: admin`. Stores the JWT in sessionStorage.

**Sites Overview** — fetches `/api/v1/admin/sites`, displays cards for each site with work item counts: review, failed, active, ready, done. Click a site to drill into its work items.

**Work Items** — loads all non-complete items for the scope (single site or all sites) and filters client-side. Split-pane layout: item list on the left (40%), detail panel on the right (58%) when selected.

### Item List Features

- **Status filter with counts** — dropdown shows counts per status: "Failed (12)", "Triaged (45)", etc. Default shows all non-complete items.
- **Type filter** — filter by item_type (content_rewrite, empty_section, etc.)
- **Retry All Failed button** — appears when there are failed items. Bulk retries all failed items with confirmation prompt.
- **Item cards show:**
  - Summary (first 100 chars)
  - Status badge
  - Item type, severity, handler agent, attempt count
  - Purple "checkpoint" badge for checkpoint items
  - Pink "needs input" badge for human review items
  - Blue domain badge when viewing all sites
  - Error preview (first 120 chars) for failed items

### Site Edit Panel

When viewing a site's work items, the "Edit Site" button opens an inline form for:
- Company Name, Tagline, Email, Phone, Address

This calls `PATCH /api/v1/admin/sites/:id` to update the sites table directly.

---

## Three Review Flows

The dashboard handles three types of review items, each with its own UI and action flow:

### 1. Placeholder Content Items

Items created by the validation system when pages contain placeholder data (e.g. "Team member names, titles, photos, bios or other section-specific data").

**What the admin sees:**
- Yellow context box showing: page name, fix_guidance, and what data is missing
- Heading: "Provide Real Data"
- Purpose-built input form based on page type:

| Page Type | Fields Shown |
|-----------|-------------|
| About | company_description, team_members (name/title/bio), company_values, founded_year |
| Contact | email, phone, contact_address, opening_hours, contact_form_note |
| Services | services array (name/description/features per service) |
| Pricing | pricing_intro, plans array (name/price/features/cta per plan) |
| Other | content, notes (generic fallback) |

- Array fields (team_members, services, plans) render as structured cards with per-field inputs, remove buttons, and an "+ Add" button
- Button: **Save & Rebuild** (green)

**What Save & Rebuild does:**
1. Updates site-level fields (email, phone, etc.) on the `sites` table if present
2. Saves page-specific content to `site_specs` as `page_content_{page_name}`
3. Creates a `content_rewrite` work item for the page (picked up by dispatch loop → page-build-handler)
4. Resolves the placeholder review item with audit trail

### 2. Checkpoint Items

Items created by agents using the `checkpoint_for_review` action (see below). The agent saves its work and creates a review item, then completes normally.

**What the admin sees:**
- Heading: "Review & Approve"
- Editable form populated with the agent's `review_data`
- Info about what happens on approve (e.g. "On approve: creates ready_to_build item → pageflow-builder")
- Optional notes field
- Button: **Approve & Continue** (green)
- Button: **Reject / Skip** (to dismiss without creating follow-on)

**What Approve & Continue does:**
1. Updates `site_specs` with the corrected review_data (versioned)
2. Creates a follow-on work item from the `on_approve` config
3. Marks the review item complete

### 3. Standard Items

All other items that aren't checkpoints or placeholder content.

**What the admin sees:**
- Heading: "Item Detail"
- Metadata grid (status, type, severity, handler, attempts, error)
- Read-only spec in a collapsible JSON view
- Action buttons based on status

### Actions by Status

| Action | What it does | Available when |
|--------|-------------|----------------|
| Retry | Resets to `triaged`, clears error and attempts | `failed`, `blocked`, `needs_human_review` (non-checkpoint) |
| Resolve | Marks `complete` with optional note | `needs_human_review`, `failed`, `blocked` |
| Dismiss | Marks `complete` as "Dismissed by admin" | `triaged` |
| Unblock | Sets status to `triaged` | `blocked` |
| Save & Rebuild | Updates data, creates rebuild item, resolves | `needs_human_review` (placeholder content) |
| Approve & Continue | Updates spec, creates follow-on | `needs_human_review` (checkpoint) |
| Reject / Skip | Resolves without follow-on | `needs_human_review` (checkpoint) |
| Retry All Failed | Bulk retries all failed items | When failed items exist |

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
5. Embeds `on_approve` instructions and `checkpoint: true` flag in the work item spec
6. Returns success — the workflow continues to the next step (usually `complete`)

Note: uses `params.AgentType` (not `params.ExecutionContext.AgentType` which doesn't exist on `ActionParams`).

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

### Testing Checkpoint Items

To test the checkpoint approval flow without waiting for an agent, insert a test work item:

```sql
INSERT INTO site_work_items (
    site_id, source, domain, item_type, severity, summary,
    spec, handler_agent, status, created_by, priority
) VALUES (
    '1368e337-dd1d-4799-bbb3-8221a1b79bcc',
    'checkpoint', 'build', 'needs_brief_review', 'high',
    'Review generated brief for finetuning.uk',
    '{"checkpoint": true, "review_data": {"company_name": "FineTuning", "tagline": "AI for the Rest of Us", "services": [{"name": "AI Consulting", "description": "Help SMEs adopt AI"}], "about_us": "We help small businesses..."}, "spec_aspect": "reviewed_brief", "source_agent": "briefing-agent", "on_approve": {"item_type": "ready_to_build", "handler_agent": "pageflow-builder"}}'::jsonb,
    'human-review', 'needs_human_review', 'test', 10
);
```

---

## Editable Form Component

The `EditableReviewForm` component auto-detects field types and renders appropriate inputs:

| Data Type | Rendered As |
|-----------|------------|
| Short string (< 100 chars) | Text input |
| Long string (≥ 100 chars) | Textarea |
| Boolean | Checkbox |
| Array of strings | Text input (comma-separated) |
| Array of objects | Structured cards with per-field inputs, remove/add buttons |
| Nested object | JSON textarea editor |

For arrays of objects (team_members, services, plans), each entry renders as a card with individual field inputs. The admin can add or remove entries without touching JSON.

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

Query parameters for `GET /work-items`: `status`, `site_id`, `item_type`, `domain` (default: `build`). When no `status` is specified, returns all non-complete items.

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
- API proxy: 5s connect, 120s read

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

The dashboard uses `IMAGE_TAG` like all other services. The deploy step automatically updates the kustomize overlay tag via `sed`.

```bash
# Build (Docker multi-stage: Node → nginx)
make build-dashboard

# Or with explicit tag
make build-dashboard IMAGE_TAG=v1.0.880

# Push
make push-dashboard IMAGE_TAG=v1.0.880

# Deploy (updates kustomize tag, applies, waits for rollout)
make deploy-dashboard IMAGE_TAG=v1.0.880

# Full release (build + push + deploy)
make release-dashboard IMAGE_TAG=v1.0.880

# Verify
make dashboard-port-forward
# Open http://localhost:8080

# Force rebuild if Docker cache is stale
docker build --no-cache -t docker.io/aqls/admin-dashboard:v1.0.880 \
  -f frontends/admin-dashboard/Dockerfile frontends/admin-dashboard/
```

**Note:** if the UI doesn't update after deploy, the Docker build may have cached an old App.tsx layer. Use `--no-cache` to force a clean build. After deploy, hard-refresh the browser (`Ctrl+Shift+R`).

**Makefile targets:**

| Target | Purpose |
|--------|---------|
| `build-dashboard` | Build Docker image |
| `push-dashboard` | Push to registry |
| `deploy-dashboard` | Update kustomize tag and apply |
| `release-dashboard` | Build + push + deploy |
| `dashboard-logs` | Tail nginx logs |
| `dashboard-port-forward` | Port-forward to localhost:8080 |
| `dev-dashboard` | Run Vite dev server locally |

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

**Login returns 401 with correct credentials:** auth-service MySQL connection is stale. Restart the auth pod.

**API calls return 502:** core-manager is down. Check pod status and logs.

**SPA blank page:** Vite build failed. Check Docker build output.

**UI shows old version after deploy:** Docker cached an old layer. Rebuild with `--no-cache`. Hard-refresh browser with `Ctrl+Shift+R`. Verify with: `kubectl exec -it <pod> -- sh -c 'grep -c "someNewFunction" /usr/share/nginx/html/assets/*.js'`

**Token expires after 1 hour:** default JWT TTL. Dashboard redirects to login on 401.

**Work items view shows "0 items":** the default filter no longer pre-selects a status. If it still shows empty, check that the API is responding: `curl -s http://localhost:8080/api/v1/admin/work-items?domain=build -H "Authorization: Bearer $TOKEN" | python3 -m json.tool | head`.

**Checkpoint item shows no editable form:** the source agent's `review_fields_from` path didn't resolve. Check the agent's collected_data.

**Approve fails with "not a checkpoint":** the work item's spec doesn't have `checkpoint: true`. Only items created by `checkpoint_for_review` can be approved — other items use retry, resolve, or save & rebuild.

**Placeholder form shows wrong fields:** the `buildPlaceholderForm` function in App.tsx matches on page_name and missing_data keywords. Add new page type patterns there for pages not yet covered.

**"Retry All Failed" retries items that shouldn't be retried:** the bulk retry resets all failed items regardless of error type. Items that failed due to structural issues (missing pages, bad specs) will fail again. Check error messages before bulk retrying.
