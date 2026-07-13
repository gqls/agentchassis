# 012 — Admin Dashboard & API Gateway

How the admin dashboard, API gateway, and content management system work together.

---

## Architecture

The admin dashboard is a React SPA served by an nginx container that also acts as an API gateway. Core-manager and auth-service stay internal as ClusterIP services.

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

---

## Dashboard Views

**Sites Overview** — cards for each site with work item counts, lock badge, last deployed date. Five buttons per site: Work Items, Pages, Direction, Media, Lock/Unlock Site. Locked sites show purple badge and reduced opacity.

**Work Items** — loads all non-complete items and filters client-side. Status counts in dropdown. Split-pane: item list with error previews, detail panel with review forms. Three review flows: placeholder content (page-type-aware forms), checkpoint (editable + approve), standard (retry/resolve). Bulk "Retry All Failed" button. "All Items" nav tab shows cross-site view with domain badges.

**Pages** — page browser with three-level structure:
- Left panel: "Site-Wide" entry (header/footer/CSS) + page list with section/locked/empty counts
- Right panel: component cards with text preview, lock/empty badges, Edit/Lock/Unlock/Remove buttons
- Edit panel: three tabs — Fields (structured form), HTML (textarea), Brief (content instructions)
- Page purpose bar (from `page_spec`) with edit and "Regenerate Page" button
- "Regenerate" on Brief tab queues LLM rewrite with updated instructions
- Suppressed sections with "Restore" button
- Site-wide components show size in kb, edit triggers full-site rebuild

**Direction** — spec editor showing all current specs as cards with data preview. Edit opens full form. Pin/unpin per aspect. "Propagate" creates work items for affected pages.

**Media** — asset browser with deployed/not-deployed/deleted groupings. Cards show purpose, type, URL, size, dimensions, origin, reference count, image thumbnail. Detail panel: larger preview, metadata, reference list (page → slot with lock indicator), delete.

---

## Three Edit Paths

| Need | Tab | Button | What happens | Speed |
|------|-----|--------|-------------|-------|
| Fix a typo | HTML | Save & Deploy | Direct edit, auto-lock, rerender | Seconds |
| Change field values | Fields | Save & Deploy | Direct edit, auto-lock, rerender | Seconds |
| Change section direction | Brief | Regenerate | Brief saved, content_rewrite queued, LLM rewrites | Minutes |
| Change whole page | Page Purpose | Regenerate Page | All unlocked sections rewritten | Minutes |
| Change site direction | Direction | Propagate | Work items created per page | Minutes-hours |

---

## Nginx Gateway

**Rate limiting:** Auth: 10/min/IP burst 5. API: 60/min/IP burst 20.
**Timeouts:** Auth: 5s connect, 15s read. API: 5s connect, 120s read.
**Caching:** JS/CSS/images/fonts get 7-day immutable (Vite content-hashed filenames).
**Security:** X-Frame-Options DENY, X-Content-Type-Options nosniff, X-XSS-Protection, strict Referrer-Policy.
**SPA routing:** `try_files $uri $uri/ /index.html`

---

## Building and Deploying

```bash
# Build (Docker multi-stage: Node builds SPA, nginx serves it)
make build-dashboard IMAGE_TAG=v1.0.886

# Push + deploy
make push-dashboard IMAGE_TAG=v1.0.886
make deploy-dashboard IMAGE_TAG=v1.0.886

# Force rebuild if Docker cache is stale
docker build --no-cache -t docker.io/aqls/admin-dashboard:v1.0.886 \
  -f frontends/admin-dashboard/Dockerfile frontends/admin-dashboard/

# Local dev (no Docker)
kubectl -n ai-persona-system port-forward svc/core-manager 8088:8088
cd frontends/admin-dashboard && npm run dev
# http://localhost:5173, proxies /api to localhost:8088
```

Uses `IMAGE_TAG` like all services. The Dockerfile handles `npm install && npm run build` — the Makefile just runs `docker build`.

---

## API Endpoints

All admin endpoints require `Authorization: Bearer <JWT>` with `role: admin`. All paths are under `/api/v1/admin/`.

### Sites

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites` | List with work item counts and lock status |
| GET | `/sites/:id` | Detail with specs |
| PATCH | `/sites/:id` | Update fields (company_name, email, phone, etc.) |
| POST | `/sites/:id/lock` | Lock site — stops all automation |
| POST | `/sites/:id/unlock` | Unlock site |

### Pages & Components

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites/:id/pages` | Page list with counts |
| GET | `/sites/:id/pages/:name/components` | Components with content, brief, locks, suppressed |
| PATCH | `/sites/:id/pages/:name/components/:id` | Edit, auto-lock, rerender |
| POST | `.../components/:id/lock` | Lock |
| POST | `.../components/:id/unlock` | Unlock |
| DELETE | `.../components/:id` | Remove + suppress |
| POST | `.../components/:id/regenerate` | Queue LLM rewrite with updated brief |
| POST | `/sites/:id/pages/:name/regenerate` | Regenerate all unlocked sections |
| PATCH | `/sites/:id/pages/:name/spec` | Update page purpose/direction |
| POST | `/sites/:id/pages/:name/restore-section` | Restore suppressed section |

### Site-Wide Components

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites/:id/site-components` | List header/footer/head |
| PATCH | `/sites/:id/site-components/:slot` | Edit, auto-lock, full-site rerender |
| POST | `.../site-components/:slot/lock` | Lock |
| POST | `.../site-components/:slot/unlock` | Unlock |

### Specs (Direction)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites/:id/specs` | All current specs |
| PATCH | `/sites/:id/specs/:aspect` | Update (versioned) |
| POST | `/sites/:id/specs/:aspect/pin` | Pin |
| POST | `/sites/:id/specs/:aspect/unpin` | Unpin |
| POST | `/sites/:id/specs/:aspect/propagate` | Create propagation items |

### Assets (Media)

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/sites/:id/assets` | List with deploy status and references |
| GET | `/sites/:id/assets/:id/references` | Component references |
| PATCH | `/sites/:id/assets/:id` | Update metadata |
| DELETE | `/sites/:id/assets/:id` | Soft-delete |

### Work Items

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/work-items` | Create |
| GET | `/work-items` | List (filterable) |
| GET | `/work-items/:id` | Detail |
| PATCH | `/work-items/:id` | Update |
| POST | `/work-items/:id/retry` | Reset to triaged (dedup-safe) |
| POST | `/work-items/:id/resolve` | Mark complete |
| POST | `/work-items/:id/approve` | Approve checkpoint |

### Auth (proxied to auth-service)

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/auth/login` | Get JWT |
| POST | `/api/v1/auth/register` | Register |

---

## Go Files

| File | Location | Purpose |
|------|----------|---------|
| `site_admin_handlers.go` | `internal/core-manager/admin/` | Sites, work items, site lock, retry, approve |
| `page_admin_handlers.go` | `internal/core-manager/admin/` | Pages, components, site-components, suppress/restore, regenerate |
| `spec_admin_handlers.go` | `internal/core-manager/admin/` | Spec list, pin/unpin, propagate |
| `asset_admin_handlers.go` | `internal/core-manager/admin/` | Asset list, references, update, delete |
| `server.go` | `internal/core-manager/api/` | Route registration |

---

## Troubleshooting

**Login 502:** auth-service down. **Login 401:** MySQL stale, restart auth pod. **API 502:** core-manager down. **Blank page:** Vite build failed. **Old UI:** Docker cache — use `--no-cache`, hard-refresh browser. **0 items:** check API with curl. **Retry 500:** dedup conflict — current code auto-resolves. **Site-wide edit:** reassembly only, no section content modified.
-e 

---

## Network Access Options (from 019_admin_access_infrastructure)

### Option A: WireGuard in the Cluster

Deploy WireGuard as a pod. Laptop connects via VPN, gets routed to cluster services. Simplest. Requires LoadBalancer service for UDP port 51820.

### Option B: External VM Bastion

Small VM (DigitalOcean/Hetzner) running WireGuard + nginx. Static IP for DNS, Let's Encrypt TLS, rate limiting. More flexible for team access.

### Option C: Port-Forward (current)

```bash
kubectl -n ai-persona-system port-forward svc/core-manager 8088:8088
# Access: http://localhost:8082/api/v1/admin/sites
```

Works now, needs no infrastructure, secure (only your machine). Move to A or B when remote/team access needed.

### Recommended Path

1. **Now:** port-forward while building
2. **This week:** WireGuard in cluster for stable access
3. **When team grows:** bastion VM for TLS + stable domain + multi-user

### Security Layers

| Layer | What | Status |
|-------|------|--------|
| Network | VPN or port-forward | Choose above |
| Auth | JWT via auth-service | Built |
| Role | RequireRole("admin") | Built |
| Audit | Admin actions logged | Built |
