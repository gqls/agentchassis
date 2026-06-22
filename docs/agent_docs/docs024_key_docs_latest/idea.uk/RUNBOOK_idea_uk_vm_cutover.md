# RUNBOOK — idea.uk front-site cutover onto the VM (keep the £29 tool live)

What this does: puts the chassis-built static front site live at **idea.uk** on the existing VM, while the
£29 Go tool keeps running and keeps owning its own paths. nginx becomes the front door — **static** for
general pages, **proxy to the Go tool** (`127.0.0.1:8080`) for the tool's reserved paths. **DNS does not
change.** The tool binary/service is untouched, so **rollback is nginx-only**.

Safety frame: the chassis build (→ GitHub Actions → Backblaze B2) is already invisible to the live site
because DNS points at the VM, not B2. Nothing in this runbook runs until you choose to. The money path is
the Stripe webhook — do not cut over until it is proven still working through the new config.

Confirmed live facts (re-confirm the ones marked *confirm*):
- Box: Hetzner (Nuremberg), `116.203.204.115` *(confirm)*. `ssh root@116.203.204.115`.
- Tool: systemd service `idea`, single Go binary, listens on `127.0.0.1:8080`. Env `/etc/idea/idea.env`.
  Order store `/var/lib/idea/orders.json` (file-based, no DB).
- Front door: nginx + Let's Encrypt. DNS (Cloudflare) → VM.
- Stripe **live** webhook destination: `https://idea.uk/stripe/webhook` (source of truth, idempotent).
- Reserved tool paths (the set nginx must proxy): `/request`, `/confirm`, `/approve`, `/decline`,
  `/stripe/webhook`, `/internal/*`, `/order/*`, plus the buyer success page and any assets the tool serves
  — **confirm the complete set in Step 0**.

---

## Prerequisite — the built site is reviewed and cutover-ready

Before any of this, the chassis build of idea.uk (the run you just triggered) must be reviewed and these
two cutover blockers fixed, because nginx routing depends on them:
1. **The site's primary CTA(s) point at the tool's real entry** (`/request`, or wherever the tool's funnel
   starts). The framework's unresolved-CTA item must be resolved to the tool before cutover, or the funnel
   breaks.
2. **The static site has no page at a reserved tool path** (no generated `/request`, `/order/...`, etc.).
   nginx will give the tool path precedence anyway, but a collision is a smell — check for it.

---

## Step 0 — Inspect (change nothing yet)

```bash
# a) the tool's FULL route list — the reserved paths nginx must proxy.
#    If the Go source is on the box, list its routes; otherwise rely on the known set + test each path.
ssh root@116.203.204.115 "grep -rEn 'HandleFunc|mux.Handle|\.(GET|POST|Handle)\(' /opt/idea 2>/dev/null | head -50"

# b) the CURRENT nginx config for idea.uk (it currently proxies everything to :8080).
#    Note the TLS cert paths, the ACME challenge location, and the 80->443 redirect — all preserved below.
ssh root@116.203.204.115 "nginx -T 2>/dev/null | awk '/server_name (www\.)?idea\.uk/,/^}/'"

# c) where the chassis build landed (B2 bucket path or the sites repo + idea.uk subdir).
```

Write down the exact reserved-path set from (a). **Anything the tool must own that is missing from the
nginx list below will be served as a static 404 and break that tool function** — completeness here is the
one thing that must be right.

---

## Step 1 — Put the built static site on the VM

```bash
ssh root@116.203.204.115 'mkdir -p /var/www/idea.uk'

# Pull the build — pick the source that matches your pipeline:
#   git:  clone/pull the sites repo, copy its idea.uk/ subdirectory into /var/www/idea.uk
#   B2:   rclone/aws-cli sync the idea.uk path from the bucket into /var/www/idea.uk
# example (B2 via rclone, adjust remote+path):
#   ssh root@116.203.204.115 'rclone sync b2:‹bucket›/idea.uk /var/www/idea.uk --fast-list'

ssh root@116.203.204.115 'chown -R www-data:www-data /var/www/idea.uk && find /var/www/idea.uk -type f | head'
```

Confirm `index.html` and the key pages (`/tools`, `/guides`, `/news`, …) are present locally before going on.

---

## Step 2 — Write the new nginx server block (stage it, don't enable yet)

Reusable proxy snippet `/etc/nginx/snippets/proxy_tool.conf`:

```nginx
proxy_pass              http://127.0.0.1:8080;
proxy_http_version      1.1;
proxy_set_header Host               $host;
proxy_set_header X-Real-IP          $remote_addr;
proxy_set_header X-Forwarded-For    $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto  $scheme;
# Stripe signature verification needs the body unmodified — do NOT add sub_filter or any body rewrite here.
```

The server block (`/etc/nginx/sites-available/idea.uk.new` — keep the real cert/ACME/redirect lines from
Step 0(b)):

```nginx
server {
    listen 443 ssl http2;
    server_name idea.uk www.idea.uk;

    ssl_certificate     /etc/letsencrypt/live/idea.uk/fullchain.pem;   # confirm path from Step 0(b)
    ssl_certificate_key /etc/letsencrypt/live/idea.uk/privkey.pem;     # confirm path

    # keep cert renewal working
    location ^~ /.well-known/acme-challenge/ { root /var/www/html; }   # confirm webroot

    # ---- reserved tool paths -> Go service (these WIN over the static root) ----
    location = /request        { include snippets/proxy_tool.conf; }
    location = /confirm        { include snippets/proxy_tool.conf; }
    location = /approve        { include snippets/proxy_tool.conf; }
    location = /decline        { include snippets/proxy_tool.conf; }
    location = /stripe/webhook { include snippets/proxy_tool.conf; }
    location ^~ /internal/     { include snippets/proxy_tool.conf; }
    location ^~ /order/        { include snippets/proxy_tool.conf; }
    # >>> add any other reserved paths/assets confirmed in Step 0(a) <<<

    # ---- everything else -> static front site ----
    root  /var/www/idea.uk;
    index index.html;
    location / {
        try_files $uri $uri/ $uri.html =404;   # =404 (not /index.html) so a missed tool path fails loudly
    }
}
```

Validate the config without enabling it:

```bash
ssh root@116.203.204.115 'nginx -t'
```

---

## Step 3 — Prove it BEFORE cutover (especially the webhook)

```bash
# reserved paths must reach the Go tool, not static. Expect the tool's codes (e.g. 405/400/401), not 404-static:
for p in /request /confirm /approve /decline /stripe/webhook /internal/run; do
  echo -n "$p -> "; curl -s -o /dev/null -w '%{http_code}\n' https://idea.uk$p
done

# THE MONEY PATH: send a Stripe test event and confirm the Go tool verifies + processes it.
#   stripe listen --forward-to https://idea.uk/stripe/webhook   (or trigger a sandbox event)
# then confirm in the tool: order moves to paid / webhook event recorded in orders.json.
```

Do not proceed until `/stripe/webhook` is proven working through nginx and the CTA funnel link resolves to
the tool.

---

## Step 4 — Cut over (one nginx swap; DNS unchanged)

```bash
ssh root@116.203.204.115 '
  cp /etc/nginx/sites-enabled/idea.uk /root/idea.uk.nginx.bak.$(date +%Y%m%d-%H%M%S) &&   # backup current
  cp /etc/nginx/sites-available/idea.uk.new /etc/nginx/sites-available/idea.uk &&
  ln -sf /etc/nginx/sites-available/idea.uk /etc/nginx/sites-enabled/idea.uk &&
  nginx -t && systemctl reload nginx && echo CUTOVER_RELOADED'
# If Cloudflare is in front, purge cache for idea.uk afterwards.
```

---

## Step 5 — Verify after cutover

```bash
curl -s -o /dev/null -w 'home %{http_code}\n' https://idea.uk/         # new static site
curl -s -o /dev/null -w 'tools %{http_code}\n' https://idea.uk/tools   # a new page
# Full tool funnel, end-to-end (a real test purchase, as done 2026-06-14):
#   /request -> operator /confirm -> /approve -> buyer pays -> /stripe/webhook delivers -> order = paid.
```

Check: the front pages load from the static site; every reserved path returns the tool's response (not a
static 404); a real purchase completes.

---

## Rollback (instant, nginx-only)

```bash
ssh root@116.203.204.115 '
  cp /root/idea.uk.nginx.bak.‹timestamp› /etc/nginx/sites-enabled/idea.uk &&
  nginx -t && systemctl reload nginx && echo ROLLED_BACK'
# purge Cloudflare cache again if used.
```

The tool's systemd service and binary are never touched, so rollback only reverts the front door.

---

## Gotchas

- **Reserved-path completeness is the whole risk.** Any tool path missing from Step 2's list is served as a
  static 404 → that tool function breaks. Confirm the full set in Step 0(a).
- **Webhook body integrity.** `/stripe/webhook` must receive the raw, unmodified body for signature
  verification — keep the proxy plain; no `sub_filter` or body rewrites on that location.
- **Never static-serve** `/stripe/webhook` or the operator paths (`/confirm`, `/approve`, `/decline`,
  `/internal/*`) under any circumstances.
- **ACME renewal.** Keep the `/.well-known/acme-challenge/` location so certbot can still renew the cert.
- **Tool-served assets.** If the Go binary serves its own assets or the buyer success page templates, those
  paths must be in the reserved list too (Step 0).
- **Sync cadence.** This deploys a snapshot. To keep the static site current with future chassis builds, add
  a pull-based sync (cron/systemd-timer `git pull`/`rclone sync`) into `/var/www/idea.uk` — separate, later.
