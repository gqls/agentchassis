# box/ — ready-to-paste artifacts for the idea.uk cutover (owner runs these)

Everything here executes ON the box (`ssh root@116.203.204.115`) — the chat
sandbox has no box SSH. Prepared 2026-07-17 against RUNBOOK §3a–§3e; the
RUNBOOK stays the narrative authority, these files are the exact payloads.

**Not this box:** `167.233.33.159` is relojistas'. Nothing here touches it.

## Order of operations

### §3a — pull-sync (safe any time; nginx untouched)
```bash
# rm -rf FIRST: `scp -r box dest` NESTS as dest/box/ when dest already exists, so a
# re-copy silently leaves the OLD script at dest/ and you re-run the bug you just fixed.
ssh root@116.203.204.115 'rm -rf /root/idea-uk-box'
scp -r box root@116.203.204.115:/root/idea-uk-box
ssh root@116.203.204.115 'cd /root/idea-uk-box && bash provision-pullsync.sh'

# Confirm you are running the copy you think you are:
ssh root@116.203.204.115 'grep -c pre-flight /root/idea-uk-box/provision-pullsync.sh'   # ≥1
```
The script needs **no TTY** once the deploy key is registered (it tests first and only
prompts if the key is missing). On the very first run, if it reports it cannot pause, add
the printed key and re-run — or use `ssh -t`.
The script pauses to let you add the printed public key as a **read-only**
Deploy Key on `gqls/vm-sites`, then sparse-clones only `idea.uk/`, installs the
5-minute timer, runs one sync, and checks all 8 pages are in `/var/www/idea.uk`.

### §3b–§3c — stage nginx (still serving nothing new)
```bash
cp proxy_tool.conf /etc/nginx/snippets/proxy_tool.conf
cp idea.uk.nginx  /etc/nginx/sites-available/idea.uk.new
# copy the real ssl_certificate lines + port-80 block from the live config in,
nginx -t
```

### §3d — prove it BEFORE cutting over
Every reserved path must reach the tool — expect the TOOL's codes (200/400/401/405),
never a static 404:
```bash
for p in /health /capacity /audience-check /subscribe /request /confirm /approve /decline \
         /op /stripe/webhook /internal/run /order/success /order/cancel /terms /refund-policy /privacy; do
  printf '%-16s -> ' "$p"; curl -sk -o /dev/null -w '%{http_code}\n' https://idea.uk$p
done
```
**THE MONEY PATH:** send a Stripe test event through the staged config and see
the order move in `orders.json` before proceeding.

### §3e — cutover (one swap, DNS unchanged) + rollback
```bash
cp /etc/nginx/sites-enabled/idea.uk /root/idea.uk.nginx.bak.$(date +%Y%m%d-%H%M%S)
cp /etc/nginx/sites-available/idea.uk.new /etc/nginx/sites-available/idea.uk
ln -sf /etc/nginx/sites-available/idea.uk /etc/nginx/sites-enabled/idea.uk
nginx -t && systemctl reload nginx
```
Re-run §3d in full, then a real end-to-end purchase. Purge Cloudflare cache.
Rollback = copy the `.bak` file back over `sites-enabled/idea.uk`, `nginx -t`,
reload. The tool binary and systemd service are never touched.

## Files

| File | Installs to | What |
|---|---|---|
| `provision-pullsync.sh` | (run once from this dir) | §3a end to end, idempotent |
| `sitesync` | `/usr/local/bin/sitesync` | the 5-minutely pull (fetch → reset → rsync) |
| `sitesync.service` / `sitesync.timer` | `/etc/systemd/system/` | runs it as `www-data` |
| `proxy_tool.conf` | `/etc/nginx/snippets/` | shared proxy block (raw body — Stripe) |
| `idea.uk.nginx` | `/etc/nginx/sites-available/idea.uk.new` | 16 tool paths + 3 legal 301s + static root |
