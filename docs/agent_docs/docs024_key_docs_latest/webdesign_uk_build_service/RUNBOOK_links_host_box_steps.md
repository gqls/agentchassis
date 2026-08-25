# RUNBOOK — bringing up links.webdesign.uk (the emailed-links host)

Owner-executed (sessions cannot SSH). Written 2026-08-24, matching the HARDENED vhost
committed that evening (token-shape regex — expectations differ from the morning
handoff's originals; see `../web_admin_console/HANDOFF_2026-08-24b_continue_here.md`
§3.9). Each step verifies before the next; nothing is publicly reachable until step 5.
Box: `root@webdesign.vs.mythic-beasts.com`. Files: `box/` beside this runbook.

## 0. Snapshot (makes rollback trivial; shows where the 404 catch-all sits)
```bash
ssh root@webdesign.vs.mythic-beasts.com '
  cp /etc/nginx/sites-available/webdesign.uk /root/webdesign.uk.nginx.bak-2026-08-24
  ls /etc/nginx/sites-enabled/
  grep -n "hostname:\|http_status:404" /etc/cloudflared/config.yml'
```

## 1. Copy both nginx files (from the repo checkout)
```bash
cd ~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/box
scp links.webdesign.uk.nginx root@webdesign.vs.mythic-beasts.com:/etc/nginx/sites-available/links.webdesign.uk
scp webdesign.uk.nginx       root@webdesign.vs.mythic-beasts.com:/etc/nginx/sites-available/webdesign.uk
```
The second REPLACES the shopfront vhost with the /c/-less copy — changes nothing live
(the apex is parked; the box sees no shopfront traffic). Backup from step 0.

## 2. Enable + test + reload
```bash
ssh root@webdesign.vs.mythic-beasts.com '
  ln -sf /etc/nginx/sites-available/links.webdesign.uk /etc/nginx/sites-enabled/links.webdesign.uk
  nginx -t && systemctl reload nginx'
```
`nginx -t` failing = nothing changed; the error names file:line.

## 3. Prove the vhost ON THE BOX before exposure — three arms
```bash
ssh root@webdesign.vs.mythic-beasts.com '
  echo -n "token-shaped -> "; curl -s -o /dev/null -w "%{http_code}\n" -H "Host: links.webdesign.uk" http://127.0.0.1:8084/c/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
  echo -n "junk /c/x    -> "; curl -s -o /dev/null -w "%{http_code}\n" -H "Host: links.webdesign.uk" http://127.0.0.1:8084/c/x
  echo -n "other path   -> "; curl -s -o /dev/null -w "%{http_code}\n" -H "Host: links.webdesign.uk" http://127.0.0.1:8084/other'
```
Expected **200 / 404 / 404**. ~~The 200 proves box → WireGuard → core-manager:8088.~~
> **CORRECTED 2026-08-25 (web_admin_console lane): the upstream is now `:8090`,** the
> delivery-only listener (`d30917150`, RFC_054 Q2, register SYS-095). The 200 proves
> box → WireGuard → **core-manager:8090**.
- 502/504 on the first arm → the WireGuard leg or the egress fence
  (`networkpolicy-wireguard-egress.yaml` must allowlist core-manager:**8090**) — STOP.
  ⚠ **This bit me on 2026-08-25 and would have bitten whoever ran step 2 next.** `d30917150`
  added `8090` to that file **and nobody applied it** — the live policy still allowed only
  `8088`, so the repo read correct while the cluster refused the very port the new vhost
  proxies to. **Committed and applied are independent facts.** Check the LIVE policy, and
  test from the pod with controls rather than reading either file:
  ```bash
  WG=$(kubectl -n ai-persona-system get pods -l app=wireguard -o jsonpath='{.items[0].metadata.name}')
  kubectl -n ai-persona-system exec $WG -- sh -c '
    for p in 8090 8088 9999; do timeout 4 nc -z core-manager.ai-persona-system.svc.cluster.local $p \
      && echo "$p OPEN" || echo "$p blocked"; done'
  ```
  8090 OPEN, 8088 OPEN (the control that proves the probe works), 9999 blocked (the control
  that proves the fence still discriminates). **Applied here 2026-08-25** with the owner's
  go-ahead: one added line, `configured` not `unchanged`, no pod restarted, and postgres
  re-verified still blocked from that pod afterwards.
- 404 on the first arm → the location regex — STOP.

## 4. Tunnel rule
Insert into `/etc/cloudflared/config.yml`, ABOVE the `http_status:404` catch-all,
indented like its siblings — **two lines only**:
```yaml
  - hostname: links.webdesign.uk
    service: http://127.0.0.1:8084
```
⚠ The fragment file (`links.webdesign.uk.cloudflared-ingress.yml`) carries an
`ingress:` header for context — do NOT paste that line; the config already has one,
and a duplicate key breaks it.
```bash
ssh root@webdesign.vs.mythic-beasts.com 'cloudflared tunnel ingress validate && systemctl restart cloudflared'
```
The restart blips admin.apis.uk + webdesign.uk for seconds — expected.

## 5. DNS — LAST, in the DASHBOARD
webdesign.uk zone → DNS → Records → Add: **CNAME** `links` →
`81f59f78-dda8-40a0-984b-cfadb36bc891.cfargotunnel.com`, **Proxied**.
Never `cloudflared tunnel route dns` — it put admin.apis.uk's record in the
noted.co.uk zone on 2026-08-23 and reported success (LANDMINES).

## 6. End-to-end from a laptop
```bash
curl -s -o /dev/null -w "%{http_code}\n" https://links.webdesign.uk/other      # 404
curl -s -o /dev/null -w "%{http_code}\n" https://links.webdesign.uk/c/x       # 404 — hardening, by design
curl -s -o /dev/null -w "%{http_code}\n" "https://links.webdesign.uk/c/$(printf 'a%.0s' {1..43})"  # 200 "no longer active" page
for i in $(seq 1 40); do curl -s -o /dev/null -w "%{http_code}\n" https://links.webdesign.uk/c/x; done  # 404s -> 429s (edge rate limit fires)
```
40× `000` = the DNS record does not exist (NXDOMAIN) — that is step 5, not a fault
(measured 2026-08-24 when the loop ran pre-DNS).

## Rollback (any ONE severs it)
1. Delete the DNS record (dashboard — instant).
2. Remove the two ingress lines; `systemctl restart cloudflared`.
3. `rm /etc/nginx/sites-enabled/links.webdesign.uk; nginx -t && systemctl reload nginx`.
Restoring the OLD shopfront vhost (re-adding /c/ there) is step 0's backup — only
needed if the move itself is being abandoned.

## After it is live
- The architecture boundary review condition is met (handoff §1.3) — run one council
  round over the exposure posture.
- The delivery-email builder mints links on THIS host — and NO email goes out before
  the second-click page ships (DECISION_2026-08-24_confirmation_needs_a_second_click.md).
