# RUNBOOK — deploying the admin console at `admin.apis.uk`

**Owner rulings, 2026-08-22:** the admin console goes on `admin.apis.uk`, behind **Cloudflare
Access**. This supersedes the VPN route as the primary way in (that path is unresolved — see
`RUNBOOK_web_admin_console.md` — and needs no longer block anything).

**This also settles standing decision D-A**, and it is worth stating exactly how, because the
council gated a submission on it: core-manager becomes reachable from the internet **only
through `admin-dashboard`'s own gateway, only on this one hostname, and only after Cloudflare
Access has authenticated the request at the edge.** It does not become a public service, and
`sitefacts.go`'s "ClusterIP only, no Ingress" remains literally true — there is still no
Ingress object anywhere in the cluster.

## The path

```
browser
  └─ https://admin.apis.uk
       └─ Cloudflare: TLS, WAF, and ACCESS (auth happens here, at the edge)
            └─ cloudflared tunnel, dialling OUT from the webdesign box (no inbound ports)
                 └─ nginx on the box, 127.0.0.1:8083, fail-closed on the Access header
                      └─ WireGuard wg0
                           └─ admin-dashboard.ai-persona-system.svc.cluster.local:8080
                                ├─ the SPA
                                ├─ /api/v1/auth/ → auth-service
                                └─ /api/v1/     → core-manager
```

## What is already true, verified 2026-08-22 — do not re-derive it

- The box runs cloudflared already, tunnel-only posture: ufw default-deny inbound, nginx binds
  loopback only (`setup-webdesignbox.sh` §1).
- The box holds a **working** WireGuard tunnel into the cluster, handshaking continuously.
- The egress fence **already permits** wireguard → `admin-dashboard:8080`
  (`networkpolicy-wireguard-egress.yaml`); no cluster change is needed.
- The full upstream leg was pre-flighted from the pod the box's traffic emerges from:
  DNS → `10.21.171.225`; `GET /` with `Host: admin.apis.uk` → **200**; `/health` → healthy;
  `POST /api/v1/auth/login` with bad credentials → **401**, not a 5xx.
- `apis.uk` is live on Cloudflare with a proxied `*.apis.uk` wildcard pointing at the **island's**
  tunnel (the 020 traffic probe). A specific `admin.apis.uk` record beats the wildcard, so the
  probe is undisturbed. **Do not delete the wildcard.**

## Files (in `webdesign_uk_build_service/box/`)

| file | goes to | what |
|---|---|---|
| `admin.apis.uk.nginx` | `/etc/nginx/sites-available/admin.apis.uk` | the vhost |
| `admin.apis.uk.map.conf` | `/etc/nginx/conf.d/upgrade-map.conf` | http-level `$connection_upgrade` map the vhost needs |
| `admin.apis.uk.cloudflared-ingress.yml` | merge into `/etc/cloudflared/config.yml` | the hostname rule |

## Steps

Ordered so that **nothing is reachable until the last step**, and each step is verified before
the next.

### 1. nginx on the box

```bash
scp admin.apis.uk.nginx     root@webdesign.vs.mythic-beasts.com:/etc/nginx/sites-available/admin.apis.uk
scp admin.apis.uk.map.conf  root@webdesign.vs.mythic-beasts.com:/etc/nginx/conf.d/upgrade-map.conf
ssh root@webdesign.vs.mythic-beasts.com '
  ln -sf /etc/nginx/sites-available/admin.apis.uk /etc/nginx/sites-enabled/admin.apis.uk
  nginx -t && systemctl reload nginx'
```

`nginx -t` must pass before the reload. If it complains `unknown "connection_upgrade" variable`,
the map file did not land — that is step 1's own failure mode, not a vhost bug.

### 2. Prove the fail-closed guard, ON THE BOX, before anything is exposed

**Both arms, because a single request cannot tell a working guard from a broken proxy:**

```bash
ssh root@webdesign.vs.mythic-beasts.com '
  echo -n "no Access header  -> "; curl -s -o /dev/null -w "%{http_code}\n" \
    -H "Host: admin.apis.uk" http://127.0.0.1:8083/
  echo -n "with Access header -> "; curl -s -o /dev/null -w "%{http_code}\n" \
    -H "Host: admin.apis.uk" -H "Cf-Access-Jwt-Assertion: dummy" http://127.0.0.1:8083/'
```

Expected: **403** then **200**.

- Both 403 → the proxy or the WireGuard leg is broken, not the guard.
- Both 200 → **the guard is not firing. Stop.** Do not proceed to step 4; you would be putting
  an unauthenticated admin API on the internet.

### 3. Cloudflare Access — create the application FIRST

In the Cloudflare dashboard, Zero Trust → Access → Applications → Add:

- Type **Self-hosted**, application domain **`admin.apis.uk`**
- Policy: **Allow**, rule `Emails` = your address (`aaa@designconsultancy.co.uk`)
- Identity: **One-time PIN** is enough and needs no IdP setup; Google/GitHub also fine
- Session duration: 24h is a reasonable default

Doing this **before** the DNS route means there is never a window where the hostname resolves
without Access in front. The nginx guard in step 2 is the belt; this is the braces.

### 4. Tunnel rule + DNS

Merge the `ingress:` entry from `admin.apis.uk.cloudflared-ingress.yml` into the box's
`/etc/cloudflared/config.yml` — **above** any wildcard or catch-all, because cloudflared takes
the first match — then:

```bash
ssh root@webdesign.vs.mythic-beasts.com '
  cloudflared tunnel ingress validate           # parses the file, catches ordering mistakes
  systemctl restart cloudflared
  cloudflared tunnel route dns <TUNNEL-NAME> admin.apis.uk'
```

`<TUNNEL-NAME>` is the box's existing tunnel — `cloudflared tunnel list` on the box. The
`route dns` call creates the `admin.apis.uk` CNAME to `<UUID>.cfargotunnel.com`.

### 5. Verify from a browser, and verify the gate

1. Open `https://admin.apis.uk` → you should get **Cloudflare Access**, not the app.
2. Complete the one-time PIN → the admin dashboard's own login screen appears.
3. Log in with your admin account (`role` must be `admin`).

Then verify the gate actually gates, from a machine that is **not** signed in — a phone on
mobile data, or a private window:

```bash
curl -s -o /dev/null -w '%{http_code}\n' https://admin.apis.uk/api/v1/admin/sites
```

Expected **302** (redirect to the Access login) or **403** — **never 200, and never 401**.
A 401 would mean the request reached the application and only the app's own auth stopped it,
which is exactly the posture Access exists to prevent.

## Rollback

Any one of these cuts access immediately, in decreasing order of bluntness:

```bash
# 1. remove the DNS record          (Cloudflare dashboard, admin.apis.uk)
# 2. remove the tunnel rule         ssh box: edit config.yml; systemctl restart cloudflared
# 3. disable the nginx vhost        ssh box: rm /etc/nginx/sites-enabled/admin.apis.uk; nginx -t && systemctl reload nginx
```

The cluster is untouched by all three — no Ingress, no service change, no image. The only
cluster-side dependency is the egress allowlist entry that already existed.

## Known gaps, recorded rather than hidden

- **`AdminOnly()` is still a bare role-string check** with no second factor, no rate limit and
  no lockout (`internal/core-manager/middleware/auth.go:181`). Access is what makes that
  acceptable; if Access is ever removed, the app alone is not sufficient for a public URL.
- **The nginx guard proves the Access header is PRESENT, not that its JWT is VALID.** Full
  verification needs a JWKS check stock nginx cannot do. It guards against the Access
  application being deleted or misconfigured — the realistic failure — not against a forged
  header, which would have to arrive through a tunnel only Cloudflare can dial.
- **`ADM-002` records known bugs and hardcoded/mock data in parts of the admin API**, and it
  predates the 2026-07-13 extraction freeze, so it is stale in an unknown direction. Expect
  rough edges on screens beyond sites/work-items/pages.
- **The build-steps screen still does not exist** — `/admin/workflows/:correlation_id` is served
  but has 0 references in the SPA. That remains the actual feature ask.
- **Static `proxy_pass` resolves once at nginx start.** If `admin-dashboard`'s ClusterIP
  changes, the vhost serves a dead address until nginx is reloaded. Same gotcha as the `/c/`
  and `/stripe/webhook` blocks beside it.
