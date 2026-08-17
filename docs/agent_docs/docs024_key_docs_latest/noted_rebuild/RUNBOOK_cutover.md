# noted.co.uk cutover — runbook and ROLLBACK

**Written 2026-08-16 during the cutover itself.** The lane had no cutover runbook;
this is it.

## The one thing to understand

**The Cloudflare Worker route is the switch.** `portfolio-sites-router` carries B2
key bindings (`B2_APP_KEY`, `B2_KEY_ID`) and fetches the bucket **directly** — it
never consults the zone origin. So while a worker owns `noted.co.uk/*`, the apex
serves the legacy app from B2 **regardless of DNS**. That is why DNS and tunnel
ingress could be prepared with zero user impact, and why rollback is one API call.

## ROLLBACK — if anything is wrong after the flip

Restores the legacy app at the apex within seconds:

```bash
set -a; source ~/.cloudflare/404-token.env; set +a
curl -s -X PUT \
  -H "Authorization: Bearer $CLOUDFLARE_API_404_TOKEN" -H "Content-Type: application/json" \
  "https://api.cloudflare.com/client/v4/zones/843436eae8f533e726bd20f0ba4537d4/workers/routes/aa774ef73ff44d17a996f8caa67ea1e4" \
  --data '{"pattern":"noted.co.uk/*","script":"portfolio-sites-router"}'
# verify: apex should return ~4393 bytes with x-amz-* headers and "being refreshed"
curl -s -o /dev/null -D - https://noted.co.uk/ | grep -ci '^x-amz-'
```

Zone `843436eae8f533e726bd20f0ba4537d4`, route id `aa774ef73ff44d17a996f8caa67ea1e4`,
original script `portfolio-sites-router`. Token: `~/.cloudflare/404-token.env`
(var `CLOUDFLARE_API_404_TOKEN`; **workers-routes scope only — it cannot read or
write DNS**).

## What was prepared BEFORE the flip (all zero-impact, all verified)

1. **Legacy app preserved at `https://noted.co.uk/legacy-app/`** — byte-identical
   to what B2 was serving (md5-compared on 5 files). Placed at
   `/var/www/noted-legacy/` on the box, **outside `/var/www/noted.co.uk`** because
   sitesync rsyncs `--delete` into that root every 5 minutes and would erase it.
   nginx `location /legacy-app/` with `alias`.
   **It MUST stay on this origin**: IndexedDB is keyed by origin, so the old app
   only sees anyone's notes at `https://noted.co.uk`. On a subdomain it would show
   every visitor an empty database and look exactly like data loss.
2. **Tunnel ingress** for `noted.co.uk` and `www.noted.co.uk` → `127.0.0.1:8082`
   (`/etc/cloudflared/config.yml`; `cloudflared tunnel ingress validate` OK;
   applied with `systemctl restart` — **never `kill -s HUP`, which TERMINATES it
   and the unit has no `Restart=`**).
3. **Apex DNS** → CNAME to tunnel `81f59f78-dda8-40a0-984b-cfadb36bc891`
   (`cloudflared tunnel route dns --overwrite-dns webdesign-box noted.co.uk`).
   Verified inert: apex still served B2 afterwards.

Backups on the box: `/root/nginx-backups/` (nginx vhost + cloudflared config).
⚠ **Never leave a backup in `/etc/nginx/sites-enabled/`** — nginx glob-includes
that directory, so a `.bak` file is parsed as a second config and `nginx -t` fails
on duplicate directives. Happened here; moved to `/root/nginx-backups/`.

## THE FLIP

```bash
curl -s -X PUT -H "Authorization: Bearer $CLOUDFLARE_API_404_TOKEN" \
  -H "Content-Type: application/json" \
  ".../workers/routes/aa774ef73ff44d17a996f8caa67ea1e4" \
  --data '{"pattern":"noted.co.uk/*","script":null}'
```

## Verify AFTER the flip — at the artefact, not the status

- apex serves the framework build (not 4393 bytes, no `x-amz-*`)
- `/legacy-app/` still serves the old app on the same origin
- `/tools/legacy-rescue/` and `/tools/write/` respond
- **re-run the two probes on the NEW origin** — this is the §6 requirement, and
  the whole migration premise rides on it:
  `python3 …/editor_tool/smoke_live_editor.py https://noted.co.uk`
- shopfront control: `curl -H "Host: webdesign.uk" http://127.0.0.1:8080/`

---

## ADDENDUM 2026-08-17 — legacy app retired

`/legacy-app/` no longer serves the old app; it **302s (relative,
`absolute_redirect off`) to `/tools/legacy-rescue/`**. Decision basis: zero
non-probe traffic in the grace period, zero inbound links, notice only existed
inside the app.

**Nothing deleted**: copies remain at `/var/www/noted-legacy/` (box), `gqls/sites`
master, and B2. **Rollback**: restore the alias block from
`/root/nginx-backups/noted.co.uk.pre-retire-20260817`, `nginx -t`, reload.
⚠ keep backups OUT of `sites-enabled/` (glob-included; breaks `nginx -t`).

The origin probe (`legacy_tool/probe_origin_after_cutover.py`) now seeds from the
rescue page itself and passes 9/9 post-retirement.
