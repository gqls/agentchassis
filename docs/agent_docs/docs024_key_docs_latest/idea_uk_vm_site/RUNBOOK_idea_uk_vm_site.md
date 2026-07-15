# RUNBOOK — idea.uk onto the VM (keep the £29 tool live)

Operational record. Newest "▶ WHERE WE ARE" section supersedes earlier ones.
Read `PLAN_idea_uk_vm_site.md` for *why*; this is *how*.

**Live, earning service. The money path is the Stripe webhook. Nothing here runs until you choose.**

```
BOX      Hetzner (Nuremberg)  116.203.204.115        ssh root@116.203.204.115
TOOL     systemd service `idea`, single Go binary, 127.0.0.1:8080
ENV      /etc/idea/idea.env          ORDERS  /var/lib/idea/orders.json   (a FILE, no DB)
FRONT    nginx + Let's Encrypt, DNS (Cloudflare) → the VM
SITE     idea.uk   site_id 1244516d-014d-421c-88c6-090bb1e9552a
PSQL     kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

**Not this box:** `167.233.33.159` is relojistas.com's site-engine box. Different machine. `setup.sh`
has box-takeover semantics (`ufw --force reset`) — never point it at the live idea.uk box.

---

## ▶ WHERE WE ARE (2026-07-14)

Design agreed, nothing executed. Phase 0 is urgent and independent; Phase 1 blocks Phase 3.

---

## Phase 0 — Leaked credentials (do this first, it is independent)

`gqls/agentchassis` is **public**. `docs/.../idea.uk/golang_files/idea.env.example` has been on
`origin/main` since **2026-06-04** with two **real** secrets. Verified by length, not by name:

| Key | Length | Verdict |
|---|---|---|
| `SMTP_USER` | 20 | **REAL** — exact AWS `AKIA…` access-key-id length |
| `SMTP_PASS` | 44 | **REAL** — exact AWS SES SMTP password length |
| `INTERNAL_API_KEY` | 64 | **REAL** — exactly `openssl rand -hex 32` |
| `ANTHROPIC_API_KEY` | 25 | placeholder (real ≈108) — not exposed |
| `STRIPE_SECRET_KEY` / `STRIPE_WEBHOOK_SECRET` | 16 / 14 | placeholders (real ≈107 / ≈38) — not exposed |

Blast radius: anyone can **send email as idea.uk** (SES reputation, phishing, your AWS bill), and
anyone can hit `/confirm` `/decline` `/approve` `/internal/run` — approve or decline orders, and
trigger Claude-billing runs. Stripe and Anthropic are **safe**.

**Owner rotates** (the repo scrub does not close this — the values are in pushed history):
```bash
# 1. AWS: delete the SES IAM user's access key, create a new one, generate new SMTP credentials.
# 2. New internal key:
openssl rand -hex 32
# 3. On the box: update both in /etc/idea/idea.env, then
ssh root@116.203.204.115 'systemctl restart idea && systemctl is-active idea'
# 4. Re-issue any operator /op links — they are signed with INTERNAL_API_KEY.
```

---

## Phase 1 — Build the three uncomposed pages (blocks Phase 3)

Confirmed state — 9 pages, 6 deployed, 3 planned with **zero sections**, which is why their nav links
404:

```
 url                              | page_type     | build_status | n_sections
 /index.html                      | landing       | deployed     | 6
 /tools.html                      | content       | deployed     | 3
 /report.html                     | landing       | deployed     | 4
 /guides/index.html               | section-index | planned      | 0   ← 404
 /news/index.html                 | section-index | planned      | 0   ← 404
 /about.html                      | content       | deployed     | 4
 /contact.html                    | content       | deployed     | 3
 /tools/audience-check/index.html | tool          | planned      | 0   ← 404
 /privacy.html                    | content       | deployed     | 1
```

Fix via the tested route — re-run `build-site-planner`. `normaliseRealisedToPlanPage`
(`v3_site_actions.go:~4383`) loads the realised pages and **unions** them with the LLM proposal, so
the six built pages are carried forward, not clobbered. **Do not hand-write `site_plan_sections`.**

1. Run `stepF_replan_read.sql` (in `../idea_uk_section_data_missing/`) and read F.1–F.4.
2. Emit a `needs_site_plan` work item for idea.uk matching the historical shape from F.3.
3. Watch the plan supersede and `site_plan_sections` populate **for all nine pages**.
4. Let the cascade run: `needs_composition` → site-design-planner → `needs_design` → webdesign-agent
   → `needs_page` → build + deploy.
5. Verify the three pages render and their nav links resolve.

**Cutover blocker, check here:** the site's primary CTA must point at the tool's real entry
(`/request`). If the CTA is unresolved, the funnel breaks the moment the static homepage goes live.

---

## Phase 2 — Take idea.uk off B2 (per-site deploy target)

### 2a. The four dead wires

`resolveGitRepoName` (`platform/orchestration/actions/helpers.go:206`) has the right logic and
**nothing calls it**. Wire all four or you get a split brain — pages to the new target, images still
to the old one.

| # | File | What |
|---|---|---|
| 1 | `site_db_actions.go:1016-1020` | `upsertSite`'s `RETURNING` clause does not select `github_repo` |
| 2 | `site_db_actions.go:193-205` | `EnsureSiteRecordAction`'s return map has no `github_repo` key, so `site_record.github_repo` never reaches `CollectedData` |
| 3 | `git_deployer_actions.go:95-98` | `GitCommitAction` reads `config["repo_name"]` directly — call `resolveGitRepoName` instead |
| 4 | `deploy_image_asset_action.go:496` | hardcodes `"repo_name": "sites"` — **this is the split-brain one** |

Then: `go build ./...`, and **rebuild the chassis image** (`make quick-agent-update IMAGE_TAG=…`).
Go changes are inert until the image is rebuilt — unlike DB config, which takes effect immediately.

### 2b. ⚠️ Guard the `vm-sites` Action BEFORE pointing idea.uk at it

`gqls/vm-sites` currently carries an Action that rsyncs **every changed domain folder** to a single
`VM_HOST` repo secret — which is **relojistas' box, `167.233.33.159`**. The moment idea.uk lands in
that repo, the Action would push idea.uk's files onto the **wrong machine**.

Add an explicit allowlist so the Action only deploys domains it is told to, and idea.uk (which pulls
itself, §3a) is simply absent from it:

```yaml
# gqls/vm-sites — deploy-targets.json at the repo root
{ "relojistas.com": "167.233.33.159" }
```
```yaml
      - name: Rsync changed domains to their mapped host
        run: |
          for domain in ${{ steps.changed.outputs.domains }}; do
            host=$(jq -r --arg d "$domain" '.[$d] // empty' deploy-targets.json)
            if [ -z "$host" ]; then echo "Skipping $domain (no mapped host)."; continue; fi
            rsync -az --delete -e "ssh -i ~/.ssh/id_ed25519 -o StrictHostKeyChecking=yes" \
              "$domain/" "$VM_USER@$host:/var/www/vm-sites/$domain/"
          done
```
Hosts become **data, not secrets**. Verify with a no-op commit that the Action skips idea.uk *before*
2c.

### 2c. Point idea.uk at the VM class

```sql
\d sites                                        -- schema first
SELECT domain, github_repo FROM sites WHERE domain = 'idea.uk';   -- before
UPDATE sites SET github_repo = 'vm-sites' WHERE domain = 'idea.uk';
SELECT domain, github_repo FROM sites WHERE domain = 'idea.uk';   -- after: vm-sites
```
Rollback: `UPDATE sites SET github_repo = NULL WHERE domain = 'idea.uk';` (falls back to `"sites"`).

**Never let the git-adapter create a target repo** — `createOrGetRepo`
(`internal/adapters/git/github_client.go:307`) creates repos **public**. `gqls/vm-sites` exists and is
private because it was made by hand. Keep it that way.

---

## Phase 3 — Serve it from the box

### 3a. Provision the pull sync (the box syncs itself)

Chosen over the push Action: a compromised runner holding one fleet-wide SSH key can reach every box;
a compromised box holding a read-only repo deploy key can reach nothing.

```bash
ssh root@116.203.204.115
# read-only deploy key for gqls/vm-sites, THIS BOX ONLY (add as a repo Deploy Key, no write access)
sudo -u www-data ssh-keygen -t ed25519 -f /var/lib/sitesync/id_ed25519 -N ''

install -d -o www-data -g www-data /var/www/idea.uk /var/lib/sitesync
cd /var/lib/sitesync
# partial + sparse clone: this box fetches ONLY its own domain folder, so a repo with
# thousands of domains stays cheap.
git clone --filter=blob:none --no-checkout git@github.com:gqls/vm-sites.git repo
cd repo && git sparse-checkout set idea.uk && git checkout main
```

`/usr/local/bin/sitesync` (idempotent; safe to run by hand):
```bash
#!/usr/bin/env bash
set -euo pipefail
cd /var/lib/sitesync/repo
git fetch --quiet origin main
git reset --hard --quiet origin/main
rsync -a --delete idea.uk/ /var/www/idea.uk/
chown -R www-data:www-data /var/www/idea.uk
```

systemd timer (`sitesync.timer`, `OnUnitActiveSec=5min`) running `sitesync.service` as `www-data`.
Run it once by hand and confirm `index.html` and the nine pages are on disk **before** touching nginx.

### 3b. The reserved-path set — this is the whole risk

The tool serves **16 routes** (`service.go:527-543`). The older cutover runbook's example block lists
**7**. Anything missing becomes a static 404 and silently breaks that function.

```
/health  /capacity  /audience-check  /subscribe  /request  /confirm  /approve  /decline
/op  /stripe/webhook  /internal/run  /order/success  /order/cancel  /terms  /refund-policy  /privacy
                                                                                    …and "/" (the landing page it loses)
```
Missing `/audience-check` → the free taster dies. Missing `/op` `/confirm` `/approve` `/decline` →
the operator flow dies. Re-confirm the list against the running binary before writing the config:
```bash
ssh root@116.203.204.115 "grep -n 'HandleFunc' /opt/idea/*.go 2>/dev/null || echo 'source not on box — use service.go:527-543'"
```

**DECISION — `/privacy` collides.** The static build generates `/privacy.html`; the tool serves
`/privacy`. `try_files $uri $uri.html` hands `/privacy` to the static page unless a `location =
/privacy` block wins. There is **no** static `/terms` or `/refund-policy`, so those two are
uncontested and stay with the tool.
*Default (unless overridden): the tool keeps all three legal pages* — they are embedded in the binary
and tied to the £29 purchase terms, so keeping them beside the money path stops the published terms
drifting from the terms the buyer agreed to.

### 3c. nginx (stage it; do not enable yet)

`/etc/nginx/snippets/proxy_tool.conf`:
```nginx
proxy_pass              http://127.0.0.1:8080;
proxy_http_version      1.1;
proxy_set_header Host               $host;
proxy_set_header X-Real-IP          $remote_addr;
proxy_set_header X-Forwarded-For    $proxy_add_x_forwarded_for;
proxy_set_header X-Forwarded-Proto  $scheme;
# Stripe verifies a signature over the RAW body. No sub_filter, no body rewrite. Ever.
```

`/etc/nginx/sites-available/idea.uk.new` — keep the real cert/ACME/redirect lines from the current
config:
```nginx
server {
    listen 443 ssl http2;
    server_name idea.uk www.idea.uk;

    ssl_certificate     /etc/letsencrypt/live/idea.uk/fullchain.pem;   # confirm
    ssl_certificate_key /etc/letsencrypt/live/idea.uk/privkey.pem;     # confirm
    location ^~ /.well-known/acme-challenge/ { root /var/www/html; }   # keep renewal working

    # ---- reserved tool paths → the Go service (these WIN over the static root) ----
    location = /health         { include snippets/proxy_tool.conf; }
    location = /capacity       { include snippets/proxy_tool.conf; }
    location = /audience-check { include snippets/proxy_tool.conf; }
    location = /subscribe      { include snippets/proxy_tool.conf; }
    location = /request        { include snippets/proxy_tool.conf; }
    location = /confirm        { include snippets/proxy_tool.conf; }
    location = /approve        { include snippets/proxy_tool.conf; }
    location = /decline        { include snippets/proxy_tool.conf; }
    location = /op             { include snippets/proxy_tool.conf; }
    location = /terms          { include snippets/proxy_tool.conf; }
    location = /refund-policy  { include snippets/proxy_tool.conf; }
    location = /privacy        { include snippets/proxy_tool.conf; }   # DECISION 3b
    location ^~ /stripe/       { include snippets/proxy_tool.conf; }
    location ^~ /internal/     { include snippets/proxy_tool.conf; }
    location ^~ /order/        { include snippets/proxy_tool.conf; }

    # ---- everything else → the static site ----
    root  /var/www/idea.uk;
    index index.html;
    location / {
        try_files $uri $uri/ $uri.html =404;   # =404, NOT /index.html — a missed tool path must fail loudly
    }
}
```
```bash
ssh root@116.203.204.115 'nginx -t'      # validate without enabling
```

### 3d. Prove it BEFORE cutting over

```bash
# every reserved path must reach the tool — expect ITS codes (405/400/401), never a static 404
for p in /health /capacity /audience-check /subscribe /request /confirm /approve /decline \
         /op /stripe/webhook /internal/run /order/success /order/cancel /terms /refund-policy /privacy; do
  printf '%-16s -> ' "$p"; curl -s -o /dev/null -w '%{http_code}\n' https://idea.uk$p
done
```
**THE MONEY PATH.** Send a Stripe test event through the new config and confirm the tool verifies and
processes it (order moves to paid in `orders.json`). **Do not proceed until `/stripe/webhook` is proven
through the new nginx** and the CTA funnel link resolves to `/request`.

### 3e. Cut over (one nginx swap; DNS unchanged)

```bash
ssh root@116.203.204.115 '
  cp /etc/nginx/sites-enabled/idea.uk /root/idea.uk.nginx.bak.$(date +%Y%m%d-%H%M%S) &&
  cp /etc/nginx/sites-available/idea.uk.new /etc/nginx/sites-available/idea.uk &&
  ln -sf /etc/nginx/sites-available/idea.uk /etc/nginx/sites-enabled/idea.uk &&
  nginx -t && systemctl reload nginx && echo CUTOVER_RELOADED'
# purge the Cloudflare cache for idea.uk afterwards
```
Then re-run 3d in full, plus a **real end-to-end purchase**:
`/request` → operator `/confirm` → `/approve` → buyer pays → `/stripe/webhook` → order = paid.

### Rollback (instant, nginx-only)
```bash
ssh root@116.203.204.115 '
  cp /root/idea.uk.nginx.bak.<timestamp> /etc/nginx/sites-enabled/idea.uk &&
  nginx -t && systemctl reload nginx && echo ROLLED_BACK'
```
The tool's binary and systemd service are never touched by any of this, so rollback reverts only the
front door.

---

## Phase 4 — Spam

**Read RUNNING_NOTES §4 first.** The existing `HANDOFF_spam_and_ip_blocklist.md` names the wrong
process and the wrong datastore; `spam_read.sql` searches Postgres for orders that live in a JSON file
on the VM and will find nothing. Order of work:

**4a. Restore the real client IP — nothing else works until this is done.** idea.uk is behind
Cloudflare, but `setup.sh` never sets `set_real_ip_from <CF ranges>` + `real_ip_header
CF-Connecting-IP`. So nginx's `$binary_remote_addr` is a **Cloudflare edge IP**: the existing
`limit_req` zone (`setup.sh:86,226,299`) buckets all Cloudflare traffic as one, and any nginx `geo`
deny or fail2ban jail would ban **Cloudflare, not the spammer**. First confirm whether the DNS record
is actually proxied (orange) or DNS-only (grey) — that decides whether this is live, and whether
Cloudflare WAF/Turnstile is even reachable as the blocking layer.
*(The Go app is unaffected: `clientIP()` takes the first XFF entry, which Cloudflare sets correctly.)*

**4b. Harden `/request`.** `handleRequest` (`service.go:301-310`) has no rate limit, no honeypot, no
validation beyond presence, and discards `ParseForm`'s error — which is exactly why
`test/test/test/test@test.com` sails through. The ingredients already exist and are simply unwired:
the per-IP sliding-window `rateLimiter` (3/hr + 20/day, `audience_check.go:31-95`) and `clientIP()`
(`:100-113`) are used **only** by the free taster. Add a honeypot field, a minimum time-to-submit,
an email-format check and length caps; reuse the rate limiter; add an `IP` field to `Order`
(`store.go:17-30` has no IP, UserAgent or Referer). Honeypot + timing **beats** an IP list — spammers
rotate IPs.

**4c. Remove the existing spam.** Signature: every field literally `test`, `test@test.com`, e.g.
`ord_1783948426211007948`. It is a **JSON file**, so there is no guarded `DELETE`, and `Store` has no
`Delete` method. Stop the service, back up `orders.json`, filter, restart — reading out exactly what
would be removed *first*. **The existing spam rows carry no IP**, so a blocklist cannot be seeded from
them; the only historical source is `grep 'POST /request' /var/log/nginx/access.log*` correlated on
timestamp against `created_at`.

Ship: the tool has no CI. `GOOS=linux GOARCH=amd64 go build -o idea .` in
`docs/.../idea.uk/golang_files/`, `scp` to `/opt/idea/idea.new`, `mv`, `systemctl restart idea`.
