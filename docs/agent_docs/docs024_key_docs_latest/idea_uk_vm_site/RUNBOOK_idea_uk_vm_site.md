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

## ▶ WHERE WE ARE (2026-07-16, supersedes below)

**Phases 1 and 2 are COMPLETE** (see RUNNING_NOTES §I–§N). Site is 9 coherent pages; per-site
deploy target is wired (v1.0.1123+), guarded (§2b: `deploy-targets.json` allowlist, verified live),
activated (§2c: `github_repo='vm-sites'`), and `gqls/vm-sites` is seeded with the built artefact.
En route we found the vm-sites Action had NEVER run — no runner on the repo and the runner image
lacked ssh/rsync — fixed: image `v1.0.1126` + `github-actions-runner-vmsites` deployment; relojistas
deploys green through the allowlist, idea.uk skip proven live 3×. **Phase 0 CLOSED 2026-07-17**:
old SES user deleted, new SMTP user live (email verified), INTERNAL_API_KEY rotated + restart —
the leaked history values are dead; /op links re-issue on next use.
**§3a DONE 2026-07-18** — pull-sync live on the box (8 pages, 5-min timer, `OnBootSec` proven across a
reboot). Traps fixed en route: `ssh` ignores `$HOME` (`/bugs_open/016`); `scp -r` nests onto an
existing destination.
**🎉 §3b–3e CUTOVER DONE 2026-07-18 14:55 UTC** — idea.uk now serves the chassis static site with all
16 tool paths proxied; verified live from outside (homepage is the chassis build; `/op`'s 404 is the
tool's own branded page, not a static miss; legal `.html` 301s work; `/nonexistent` 404s loudly;
security headers intact). The live file is **`idea.conf`**; the staged config was rebuilt as a
superset of setup.sh's template first — it had been dropping `proxy_read_timeout 930s`, `limit_req`,
the port-80 block, IPv6, ssl_protocols and the security headers. RUNNING_NOTES §T.
**2026-07-19 — post-cutover reality.** Two things the cutover exposed, both now understood:
- **Funnel was orphaned** (`/bugs_open/017`): the tool's forms lived on the `/` page it lost.
  **FIXED** — `audience-check-form` (tools.html) and `report-request-form` (report.html) authored as
  chassis sections and verified live, including the honeypot, `_elapsed`, and the JS asset at
  `/tools/assets/report-request-form.js` (200). `sql/p2_01` + `p2_02`.
- **Chrome broken site-wide** (`/bugs_open/018`): **31 of 33 homepage links are `href=""`** (whole
  nav, every CTA) + empty logo `src`. Pre-dates the cutover; the cutover only made it visible.
  **OPEN — this is now the top job.**
- The chassis→vm-sites→box pipeline is **proven end-to-end** by a real build (was only code-verified).

**Next: (1) fix the chrome — `/bugs_open/018`; (2) prove the money path (Stripe test event);
(3) confirm `proxy_read_timeout` landed on the box; (4) purge Cloudflare; (5) Phase 4 tool deploy;
(6) SES bounce records.**
⚠️ **`setup.sh` must not be re-run unmodified** — it rewrites `idea.conf` back to tool-only.
⚠️ **Verify against the deployed artefact, never the work-item status** — this site has produced a
`failed` item for a page that rendered and deployed correctly (claim-timeout churn), and a
`missing_structure` finding that is false because the footer is a `<section class="footer-…">`, so
`grep '<footer'` "confirms" it.
§3b correction: static `terms.html`/`refund-policy.html` DO exist and footers link all three legal
pages with `.html` — add 301s (`/terms.html → /terms` etc.) at cutover so the tool stays canonical.

## ▶ WHERE WE WERE (2026-07-14)

Design agreed, nothing executed. Phase 0 is urgent and independent; Phase 1 blocks Phase 3.

---

## Phase 0 — Leaked credentials — ✅ CLOSED 2026-07-17

**Executed by the owner 2026-07-17:** new SES SMTP IAM user created and verified sending; the old
user (`ses-smtp-user.20260611-195505`, whose key was the leaked one) **deleted**; `INTERNAL_API_KEY`
regenerated; `/etc/idea/idea.env` updated; `idea` restarted and healthy. The values in public git
history are now dead. Operator `/op` links: old ones no longer verify — issue fresh ones on next use.
Original finding kept below for the record.

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
a compromised box holding a read-only repo deploy key can reach nothing. Full rationale: `BRIEFING §7`.

**The scripts are written and committed — do not hand-type the commands.** They live in the `box/`
folder beside this runbook: `provision-pullsync.sh` (the one-shot installer), `sitesync` (the sync
itself), `sitesync.service`, `sitesync.timer`. This section walks through what the installer does so
you can run it with your eyes open. It is idempotent — safe to re-run if a step fails.

**Safe to run any time.** It stages files onto the box and installs a timer; it does **not** touch
nginx, so nothing the public sees changes. The cutover is §3b–§3e, a separate deliberate step.

#### Get the scripts onto the box and run the installer
```bash
# from this folder on your workstation. rm -rf FIRST — see the trap below.
ssh root@116.203.204.115 'rm -rf /root/idea-uk-box'
scp -r box root@116.203.204.115:/root/idea-uk-box
ssh root@116.203.204.115 'grep -c pre-flight /root/idea-uk-box/provision-pullsync.sh'  # ≥1 = fresh copy
ssh root@116.203.204.115 'cd /root/idea-uk-box && bash provision-pullsync.sh'
```
The installer must run **as root** (it writes to `/usr/local/bin`, `/etc/systemd/system`, `/var/www`)
and **from its own directory** (it installs `sitesync` + the units that sit beside it).

> ⚠️ **`scp -r` nests when the destination already exists.** `scp -r box host:/root/idea-uk-box`
> creates `/root/idea-uk-box/` on the first run, but on the *second* run copies into it —
> `/root/idea-uk-box/box/` — leaving the **old** script at the path you then execute. This cost a
> real cycle on 2026-07-18: a fixed script was copied, the stale one ran, and the identical failure
> reappeared, which reads exactly like "the fix didn't work". Always `rm -rf` the destination first,
> and grep the copy for a string only the new version contains before running it.
> The installer needs **no TTY** once the deploy key is registered — it tests authentication first and
> prompts only when the key is missing (a `read` pause cannot work under `ssh host 'bash script'`).

#### What each stage does — and what to watch for

**Stage 1 — `== dirs ==`.** Creates `/var/www/idea.uk` (the web root nginx will serve) and
`/var/lib/sitesync` (the sync's private home), both owned `www-data:www-data`, and locks
`/var/lib/sitesync/.ssh` to `0700`. Nothing to check; it is `install -d`, so re-runs are no-ops.

**Stage 2 — `== deploy key ==`.** Generates an ed25519 keypair **as `www-data`** at
`/var/lib/sitesync/.ssh/id_ed25519`. This is the box's read-only identity to GitHub. Generated once;
on a re-run the `if [ ! -f … ]` guard skips it so the key is stable. It also adds GitHub to
`known_hosts` so the first fetch isn't blocked on a host-key prompt.

**Stage 3 — the PAUSE (`read -rp`).** The script prints the **public** key and stops. Go to
`gqls/vm-sites → Settings → Deploy keys → Add deploy key`, paste it, title it `idea.uk-box sitesync`,
and **leave "Allow write access" UNTICKED** — read-only is the whole security property (§E; a box that
can only read cannot poison other sites or reach sibling boxes). Press Enter only after it's added.
- ⚠️ This is a **repo Deploy Key**, not an account SSH key and not the `vm-sites` runner's key. It is
  specific to `gqls/vm-sites` and read-only. Do not reuse the runner's key here.

**Stage 3b — `== pre-flight ==`.** Before cloning, authenticates to GitHub as `www-data` and matches
the greeting text (`ssh -T` exits 1 even on success, so exit status is useless here). Success reads
`Hi gqls/vm-sites! You've successfully authenticated…`. It distinguishes a host-key failure, a
refused key, and the `/var/www/.ssh` fallback below — so you get the actual problem, not a clone error.

**Stage 4 — `== sparse clone ==`.** Clones `gqls/vm-sites` into `/var/lib/sitesync/repo` with
`--filter=blob:none --no-checkout`, then `sparse-checkout set idea.uk` and `checkout main`. Net: the
box materialises **only** `idea.uk/`, so a repo of thousands of domains stays cheap. Guarded by
`if [ ! -d …/repo/.git ]`, so a re-run won't re-clone.

> ⚠️ **`HOME` does not point ssh at the key** (hit for real 2026-07-18, `/bugs_open/016`). ssh expands
> `~` from the **passwd entry**, not `$HOME` — and `www-data`'s passwd home is `/var/www`, which it
> cannot write. `env HOME=…` configures git but leaves ssh hunting in `/var/www/.ssh`, producing
> *"Host key verification failed"* and then *"Permission denied (publickey)"* while the GitHub deploy
> key sits there reading **"Never used"**. The scripts therefore name the identity and known_hosts
> explicitly via `GIT_SSH_COMMAND` (`-i … -o IdentitiesOnly=yes -o UserKnownHostsFile=…`) — in the
> provisioner **and** in `sitesync`, since the 5-minutely `git fetch` runs as the same account and
> would otherwise fail on every tick. To see which files ssh really opens:
> `sudo -u www-data ssh -v -T git@github.com 2>&1 | grep -iE 'identity file|known hosts'`

**Stage 5 — `== install script + units ==`.** Installs `sitesync` to `/usr/local/bin` (0755) and the
two unit files to `/etc/systemd/system`, `daemon-reload`, then `enable --now sitesync.timer`. The timer
(`OnBootSec=1min`, `OnUnitActiveSec=5min`) is now live and will re-sync every 5 minutes.

**Stage 6 — `== first sync + verify ==`.** Runs one sync immediately and lists the web root, printing
`OK`/`MISSING` for each of the eight built pages. **This is your gate: all eight must read `OK` before
you go anywhere near nginx.** (The Free Audience Check is intentionally absent — it's a live-tool
pointer with no static file.)

#### What `sitesync` itself does (runs every 5 min, and by hand)
```bash
cd /var/lib/sitesync/repo
git fetch --quiet origin main          # pull new commits; working tree untouched
git reset --hard --quiet origin/main   # force the tree to EXACTLY match origin — never merge
rsync -a --delete idea.uk/ /var/www/idea.uk/   # mirror into the web root; --delete removes orphans
```
`reset --hard` because the box is a read-only mirror: any local drift is obliterated, not merged, since
the repo is the single source of truth. It runs as `www-data` (via `sitesync.service`), so everything
rsync writes is already correctly owned — no `chown` step, no root in the sync path.

#### Verify by hand, any time
```bash
ssh root@116.203.204.115 '
  systemctl start sitesync.service                       # force a sync now
  systemctl list-timers sitesync.timer --no-pager        # confirm it is scheduled
  systemctl status sitesync.service --no-pager | tail -5 # last run result
  ls -R /var/www/idea.uk | head'
```
`journalctl -u sitesync.service --no-pager | tail -20` shows the history if a sync ever fails.

**Do not proceed to §3b until the eight pages are on disk and a hand-run sync is clean.**

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

**✅ ROUTE COVERAGE VERIFIED 2026-07-18.** The box has no source (`/opt/idea/*.go` → "source not on
box"), so re-confirmed against `service.go:596-612` in the repo: **16 tool routes** + `/` (the landing
page, which the static site takes over). `box/idea.uk.nginx` covers all 16 with **15 location blocks**
— 12 exact + 3 prefix (`^~ /stripe/`, `^~ /internal/`, `^~ /order/`; the last covers both
`/order/success` and `/order/cancel`), 15 `proxy_tool.conf` includes. No gaps.

**DECISION CONFIRMED BY OWNER 2026-07-18 — the tool keeps all three legal pages** (`/terms`,
`/refund-policy`, `/privacy`), and the static `.html` copies 301 to them. This was previously only
the recommended default; it is now settled. Rationale below stands.

**DECISION — the legal pages collide (all three).** The static build generates `/privacy.html` —
and, **correction 2026-07-16**: `terms.html` and `refund-policy.html` too (this runbook previously
said it didn't), with footers linking all three *with the `.html` extension*. Exact-match proxy
locations only catch the extension-less paths, so add 301s at cutover to keep one canonical set:
```nginx
location = /terms.html         { return 301 /terms; }
location = /refund-policy.html { return 301 /refund-policy; }
location = /privacy.html       { return 301 /privacy; }   # if DECISION stays "tool keeps /privacy"
```
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
**§3d(ii) — THE FUNNEL-ENTRY DIFF (`bugs_open/017`, prevention `features_open/011`).** The loop above
proves the routes *reach* the tool; it does NOT prove the new static `/` still contains the **forms
that drive them**. That is exactly what 017 lost: the cutover kept `/audience-check` and `/request`
reachable but dropped the tool's own landing-page forms, leaving GET links to POST-only handlers
("405 POST only") on a live earning site. **Every POST-only funnel route must have a live `<form>`
posting to it, or the cutover orphans the funnel.** Run this against the staged config (the port-8443
rehearsal, README §4c) BEFORE the public swap — a block here is free, nothing public has changed:
```bash
NEW=https://127.0.0.1:8443          # staged config; Host header carries idea.uk
missing=0
for route in /audience-check /request; do    # POST-only handlers that carried the entry forms
  code=$(curl -sk -o /dev/null -w '%{http_code}' -H 'Host: idea.uk' "$NEW$route")   # GET -> 405 = POST-only
  hasform=$(for pg in / /tools.html /report.html /index.html; do
              curl -sk -H 'Host: idea.uk' "$NEW$pg"; done | grep -c "action=\"$route\"")
  printf '%-16s GET=%s  forms-posting-to-it=%s\n' "$route" "$code" "$hasform"
  [ "$code" = 405 ] && [ "$hasform" -eq 0 ] && { echo "  ✗ ORPHANED: $route is POST-only, no page posts to it"; missing=1; }
done
[ "$missing" -eq 0 ] || echo "STOP — cutover would orphan the funnel (bugs_open/017). Author the entry form into the static build first."
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

---

## Phase 5 — Adding a GUIDE page (the ideas-pipeline build path) — proven 2026-07-25

This is the repeatable recipe for every guide in `features_open/014`. It bypasses
`build-site-planner` entirely: **never re-plan the site to add one page** — the re-plan path
recomposes everything and is how built pages get clobbered (`bugs_open/001`, `050`).

### The mechanism, so you can reason about it rather than copy blindly

- A guide is a normal `pages` row: `url='/guides/<slug>/index.html'`, `page_type='guide'`.
  Fleet-verified shape (gamesdesign.co.uk ×5, relojistas ×4, vetcomparison ×3).
- **The guides hub populates itself.** `guide-list_pre_037`
  (`9d5e461a-8981-4ecc-b236-05895edfc15d`) sources `items` from
  `query.pages_where_type:guide` — `queryresolve.go:81` → `resolvePagesWhereType`. Add a guide
  page, and the hub lists it on its next render. No edit to the hub.
- **Eligibility is the gate.** That resolver applies `FetchablePageEligibilitySQL`:
  `deployed_at IS NOT NULL OR build_status='deployed'`. So a guide only appears once it has
  actually shipped. Build the guide FIRST, then render the hub.
- `Generic Text Block` renders `{{.content}}` **unescaped** — authored HTML in `content_data`
  is safe (verified against the live `/guides/rng-design/index.html` artefact).

### The recipe

```bash
# 1. Create the page + sections with authored content_data. Template: sql/p4_01_guide_patents.sql
#    MUST set page_components.slot_name — see the trap below.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  < docs/.../idea_uk_vm_site/sql/p4_0N_guide_<slug>.sql

# 2. Direct-fire a section_data_resolved rerender (the queue is unreliable — bugs_open/029/030).
#    NOT the plain 049b assemble-only call: a new page has no rendered_html to assemble.
#    check_rerender_mode reads input_data.spec.reason, NOT input_data.reason.
kcat -P -c 1 -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests  ... (headers per sql/../scratchpad fire_patents_guide.sh)
{"action":"orchestrate","config":{"agent_type":"page-rerender"},
 "input_data":{"page_id":"<id>","site_id":"1244516d-014d-421c-88c6-090bb1e9552a","domain":"idea.uk",
   "spec":{"domain":"idea.uk","page_id":"<id>","filename":"guides/<slug>/index.html",
           "page_name":"guide-<slug>","reason":"section_data_resolved"}}}

# 3. Confirm the sections actually RENDERED (not carried) before believing the job:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT collected_data->'rerender_sections'->>'rerendered' AS rendered,
       collected_data->'rerender_sections'->>'carried'    AS carried
FROM orchestration_states WHERE correlation_id='<CORR>';"
#    rendered must equal the section count and carried must be 0.

# 4. Verify LIVE, not the DB (VM sitesync lag ~5 min after the git commit):
curl -s -o /dev/null -w '%{http_code}\n' https://idea.uk/guides/<slug>/index.html

# 5. Backfill pages.sections (the rerender path does NOT write it) — sql/p4_01c.
# 6. Lock the authored sections — sql/p4_03.
```

### ⚠ TRAP 1 — `slot_name` NULL renders NOTHING and the job still reports COMPLETED

`rerender_page_sections_action.go:249` looks the component up as **`schemas[s.slotName]`** — keyed
on `page_components.slot_name`, *not* `component_id` (`loadStoredSections` reads
`COALESCE(slot_name,'')`). A NULL `slot_name` misses the map and takes the *"component not found,
carrying stored HTML"* branch at `:251-257`. On a brand-new page the stored HTML is empty, so:

```
rerender_sections -> {"section_count":3,"rerendered":0,"carried":3}
render_page       -> {"skipped":true,"reason":"no components found for page"}
workflow          -> complete_skipped, status COMPLETED     ← green, and the page does not exist
```

**The key is the component's `function` column, not its `name`.** Verified across the fleet:
`Generic Text Block` → `generic-text-block`; `hero` → `hero`; `call-to-action` → `call-to-action`.
Set it from the component itself so it cannot drift:

```sql
UPDATE page_components pc SET slot_name = cc.function
FROM content_components cc WHERE cc.id = pc.component_id AND pc.page_id = '<id>';
```

This cost one wasted round on 2026-07-25 (`sql/p4_01b`). It is the CLAUDE.md/016b
"trust the rendered artefact, not the status" rule in its purest form: **nothing failed.**

### ⚠ TRAP 2 — the hub's own listing component may be the static one

idea.uk's `/guides/index.html` shipped with `content-listing`
(`aa3e4b68-bcea-49ca-890a-c111acefa551`), whose `articles` is a **static array** with no query
source — it had always been empty and would never have picked up a guide page. Swap it to
`guide-list_pre_037` once (done: `sql/p4_02`), moving `slot_name` **and** `pages.sections` with it.
Check before assuming a hub self-populates:

```sql
SELECT pc.slot_name, cc.name, cc.input_schema->'fields'->'items'->>'source'
FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
JOIN pages p ON p.id=pc.page_id WHERE p.url='/guides/index.html' AND p.site_id='<site>';
```

### ⚠ TRAP 3 — "no orchestration row" means QUEUED, not dropped. Check before you re-fire.

The first hub rerender fire produced no orchestration row for ~4 minutes and I re-fired it as a
lost spawn (`bugs_open/003`). **That was wrong**, and the check that proves it takes one query:

```sql
-- Is it MY message, or is the whole consumer stalled? Look at everyone's orchestrations.
SELECT correlation_id, current_step, status, created_at FROM orchestration_states
WHERE created_at > now() - interval '10 minutes' ORDER BY created_at DESC LIMIT 10;
```

Zero rows **fleet-wide** since the fire ⇒ the generic-requests consumer is stalled/backed up
(`bugs_open/029/030`), and every thread's dispatch is waiting, not just yours. Only if *other*
orchestrations are starting normally while yours never appears do you have a dropped spawn.

Confirm the message actually reached the topic before suspecting anything at all — it is on the
log whether or not a consumer has taken it:

```bash
kubectl -n kafka run -i --rm kcat-read-$(date +%s) --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -C -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests -o -6 -e -q -f '%o %T | %s\n'
```

Cost of getting this wrong: a duplicate rerender. Harmless here (idempotent — the same
`section_data_resolved` render twice), but the same reflex against the council gate costs a whole
duplicate review round. This is `memory/council-queue-latency-trap` and it did not fire in time.

`agent-page-rerender` pods are one-shot Jobs that **idle-shut-down after ~3 minutes**
(`agentbase/agent.go:1541`, observed `idle_duration 184s`), so "no page-rerender pod running" is
also not evidence of anything — a new Job is created per dispatch.

Note the *deploy* step also retries: the patents guide was committed to `vm-sites` twice
(`b253d868`, `b78f70c4`) because the first git-adapter success response was not consumed. Harmless
— identical content — but it means "two commits" is not evidence of two edits.
