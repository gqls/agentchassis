# 019 — Admin Dashboard & API Gateway

How the admin dashboard and its nginx gateway work, how to develop locally, and how to build and deploy.

Covers the architecture diagram, 
file layout, 
what each piece does, 
all the API endpoints, 
nginx routing details, 
local dev workflow, 
build/deploy steps, 
admin user setup, 
VPN integration notes, 
core-manager changes, 
and troubleshooting.

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
    ├── App.tsx             # Dashboard application (login, sites, work items)
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

## What the Dashboard Does

The SPA provides three views:

**Login** — posts credentials to `/api/v1/auth/login` (proxied to auth-service). Validates that the user has `role: admin`. Stores the JWT in sessionStorage.

**Sites Overview** — fetches `/api/v1/admin/sites`, displays cards for each site with work item counts: review, failed, active, ready, done. Click a site to drill into its work items.

**Work Items** — fetches `/api/v1/admin/work-items` with filters for status, site, and item type. Split-pane layout: item list on the left, detail panel on the right when selected. The detail panel shows summary, spec (expandable JSON), status, severity, handler agent, attempt count, and error message.

HITL actions available in the detail panel:

| Action | What it does | Available when |
|--------|-------------|----------------|
| Retry | Resets to `triaged`, clears error and attempts | `needs_human_review`, `failed`, `blocked` |
| Resolve | Marks `complete` with optional resolution note | `needs_human_review`, `failed`, `blocked` |
| Dismiss | Marks `complete` as "Dismissed by admin" | `triaged` |
| Unblock | Sets status to `triaged` | `blocked` |

---

## API Endpoints Used

All endpoints require `Authorization: Bearer <JWT>` with `role: admin`.

| Method | Path | Handler | Purpose |
|--------|------|---------|---------|
| POST | `/api/v1/auth/login` | auth-service | Get JWT token |
| GET | `/api/v1/admin/sites` | core-manager | List sites with work item counts |
| GET | `/api/v1/admin/sites/:id` | core-manager | Site detail with specs |
| GET | `/api/v1/admin/work-items` | core-manager | List work items (filterable) |
| GET | `/api/v1/admin/work-items/:id` | core-manager | Work item detail |
| PATCH | `/api/v1/admin/work-items/:id` | core-manager | Update status, spec, handler |
| POST | `/api/v1/admin/work-items/:id/retry` | core-manager | Reset to triaged for retry |
| POST | `/api/v1/admin/work-items/:id/resolve` | core-manager | Mark complete with resolution |
| PATCH | `/api/v1/admin/sites/:id/specs/:aspect` | core-manager | Update site spec (versioned) |

Query parameters for `GET /work-items`: `status`, `site_id`, `item_type`, `domain` (default: `build`).

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

**SPA routing:** `try_files $uri $uri/ /index.html` ensures client-side routing works — all unmatched paths return index.html and React handles the route.

---

## Local Development

Local dev uses the Vite dev server which provides hot module replacement (instant updates on save) and proxies API calls to a port-forwarded core-manager.

**Prerequisites:** Node.js 20+, npm, kubectl access to the cluster.

**Setup:**

```bash
# Terminal 1: port-forward core-manager for API access
kubectl -n ai-persona-system port-forward svc/core-manager 8088:8088

# Terminal 2: port-forward auth-service for login
# (needed because Vite only proxies /api to one target)
# Alternative: use the deployed dashboard gateway for login, dev for UI work

# Terminal 3: start Vite dev server
cd frontends/admin-dashboard
npm install
npm run dev
# Opens http://localhost:5173
```

The `vite.config.js` proxies all `/api` requests to `localhost:8088`. For login to work locally, core-manager needs the auth proxy route (or you port-forward auth-service on a different port and adjust the login URL temporarily).

The simpler local dev flow is:
1. Deploy the dashboard to the cluster
2. Port-forward the dashboard: `make dashboard-port-forward`
3. Edit App.tsx locally, rebuild and redeploy when ready

---

## Building and Deploying

**Build the Docker image:**

```bash
make build-dashboard DASHBOARD_TAG=v1.0.1
```

This runs a multi-stage Docker build:
1. Node 20 installs deps and runs `vite build`, producing optimised static files in `dist/`
2. nginx 1.27-alpine copies the built files and the nginx.conf
3. Final image is ~25MB

**Push to registry:**

```bash
make push-dashboard DASHBOARD_TAG=v1.0.1
```

**Deploy to cluster:**

```bash
make deploy-dashboard
```

This applies the kustomize manifests: 2-replica Deployment + ClusterIP Service on port 8080.

**Full release (build + push + deploy):**

```bash
make release-dashboard DASHBOARD_TAG=v1.0.1
```

**Verify:**

```bash
# Port-forward and test
make dashboard-port-forward
# In another terminal:
curl -s http://localhost:8080/health
# Open http://localhost:8080 in browser
```

---

## Admin User Setup

The dashboard requires an admin user in the auth database.

**Register (from inside the cluster):**

```bash
kubectl -n ai-persona-system exec -it $(kubectl -n ai-persona-system get pod -l app=auth-service -o jsonpath='{.items[0].metadata.name}') -- wget -qO- \
  http://localhost:8081/api/v1/auth/register \
  --post-data='{"email":"admin@example.com","password":"YourPassword123","client_id":"demo_client"}' \
  --header='Content-Type: application/json' 2>&1
```

**Promote to admin (MySQL):**

```bash
mysql -h rs17.uk-noc.com -u catalogu_personae -p"PASSWORD" --skip-ssl catalogu_vectordb_chassis \
  -e "UPDATE users SET role = 'admin', subscription_tier = 'enterprise' WHERE email = 'admin@example.com';"
```

The auth service issues tokens with `role: user` on registration. The MySQL update changes the role so subsequent logins produce admin JWTs.

---

## VPN Access (Future)

The dashboard is currently accessible via port-forward only. When WireGuard is deployed (see 016_admin_access_infrastructure.md), the dashboard service will be reachable via the VPN without port-forwarding. The gateway stays as a ClusterIP service — WireGuard provides the external entry point.

---

## Relationship to Core-Manager

Core-manager was previously configured to serve the SPA from `/admin` and proxy auth requests. Both responsibilities have moved to the nginx gateway:

- `server.go` no longer has `StaticFS("/admin", ...)` 
- `server.go` no longer calls `setupAuthProxy()`
- `auth_proxy.go` is not needed

Core-manager is now a pure API service. The nginx gateway is the only thing that needs to know about both auth-service and core-manager's addresses.

---

## Troubleshooting

**Login returns 502 Bad Gateway:** auth-service is down or the upstream address in nginx.conf is wrong. Check `kubectl -n ai-persona-system get pods | grep auth` and the nginx error log: `make dashboard-logs`.

**Login returns 401 but credentials are correct:** the auth-service MySQL connection may be stale (cPanel kills idle connections). Restart the auth-service pod: `kubectl -n ai-persona-system delete pod -l app=auth-service`.

**API calls return 502:** core-manager is down or crash-looping. Check `kubectl -n ai-persona-system get pods | grep core-manager`.

**SPA shows blank page:** the Vite build may have failed. Check Docker build logs. Verify `dist/index.html` exists in the built image.

**Token expires after 1 hour:** this is the default JWT TTL from auth-service. The dashboard currently redirects to login when it gets a 401. A future improvement would be to use the refresh token to get a new access token transparently.
