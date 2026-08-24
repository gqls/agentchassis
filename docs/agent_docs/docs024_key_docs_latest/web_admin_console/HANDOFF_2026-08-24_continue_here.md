# HANDOFF 2026-08-24 — web admin console + the exposure work around it

**Read order, cold:** this file · `PLAN_2026-08-24_build_steps_screen.md` (the next build) ·
`RUNBOOK_deploy_admin_apis_uk.md` (how the console got public) · `NOTES_web_admin_console.md` ·
the two decision docs in `../webdesign_uk_build_service/` dated 2026-08-22 (D-A) and the
D-B SQL pair dated 2026-08-22. The WireGuard-client saga: `RUNBOOK_web_admin_console.md`
(superseded as the access route, kept for the diagnosis).

## 0. State in one paragraph

**The owner has a working web admin console: `https://admin.apis.uk`, behind Cloudflare
Access (One-time PIN, emails `info@designconsultancy.co.uk` + `uk@websy.uk`), through the
webdesign box's cloudflared tunnel, over WireGuard, to `admin-dashboard` in the cluster. He
has logged in (as `uk@websy.uk` — so an admin account EXISTS; the old `[UNVERIFIED]` in the
runbooks is resolved) and can see every site.** That answered **D-A** (narrowly — see the
decision doc; it does NOT license `/c/`//`/d/`). **D-B** is applied and verified at the bot
("three or four days, usually sooner"). The `/c/` prefetch guard is **LIVE** in
`v1.0.1332`. The `/c/` route's **move off the shopfront is written but NOT applied to the
box** (§2 — the owner does that; sessions cannot SSH). The **build-steps screen is planned,
not built** (§4). Everything below has its proof inline or one pointer away.

## 1. What is LIVE, with proofs (all measured 2026-08-24 unless dated)

| thing | state | proof |
|---|---|---|
| `admin.apis.uk` end to end | **LIVE, owner logged in** | screenshot 2026-08-24; unauthenticated curl → 302 to `billowing-smoke-5ed4.cloudflareaccess.com` |
| Cloudflare Access app | **LIVE**, OTP-only (no IdP configured, so OTP is the only method — that is why no "choose PIN" step ever appeared) | the 302 above; policy `Owner only`, emails ×2 |
| DNS `admin.apis.uk` | CNAME → tunnel `81f59f78…` (webdesign-box), **created by hand in the dashboard** | owner's zone listing. ⚠ `cloudflared tunnel route dns` put the record in the WRONG ZONE and reported success — LANDMINES 2026-08-23 |
| Box nginx `:8083` vhost | **LIVE**, fail-closed on missing `Cf-Access-Jwt-Assertion` (403/200 both arms proven on the box) | owner's terminal, 2026-08-23 |
| `/c/` prefetch guard | **LIVE** — `v1.0.1332` carries `0e9cb31ee`, proven by `git merge-base --is-ancestor` against the pod's provenance stamp `0b262ed5e` | this session. Closes the ANNOUNCED vector only; mail-scanner residual stands (owner's open call) |
| D-B (three-or-four-days) | **LIVE at the bot** | bot replied with the new figure post-roll |
| wireguard egress fence | **LIVE** — peers reach kube-dns, core-manager:8088, auth-service:8081, admin-dashboard:8080, nothing else (postgres proven blocked, with control) | `networkpolicy-wireguard-egress.yaml`, applied 2026-08-22 |
| `peer_laptop` keys | **ROTATED** in place 2026-08-22 (PSK had reached a chat transcript); box peer undisturbed | NOTES + runbook "Rotating a peer's keys" |
| laptop VPN | **UNRESOLVED and deliberately abandoned** as the access route: client exonerated at packet level (148-byte initiations out, zero back; server rx=0), network exonerated by STUN. Loss is in transit/at the node | runbook troubleshooting section. Do not resurrect without reading it |

## 2. IMMEDIATE — the `/c/` move, written this session, owner applies on the box

Owner ruled 2026-08-24: move `/c/` off the shopfront. Files are committed in
`../webdesign_uk_build_service/box/`:

- **`links.webdesign.uk.nginx`** — new vhost, loopback `:8084`, `location /c/` only,
  everything else 404. Deliberately public once DNS exists (customers click these from
  email); safe now: guard live + `customer_access_tokens` = 0 rows (re-check before
  trusting). `/d/` belongs here when built.
- **`webdesign.uk.nginx`** — `/c/` block and its `confirm_rl` zone REMOVED, pointer comment
  left. **The box still runs the OLD file** until the owner applies both.
- **`links.webdesign.uk.cloudflared-ingress.yml`** — the tunnel rule fragment.

Box steps (owner, ~5 min): copy both nginx files into `/etc/nginx/sites-available/`,
symlink links.webdesign.uk into sites-enabled, `nginx -t && systemctl reload nginx`; add the
ingress rule ABOVE the 404 catch-all, `cloudflared tunnel ingress validate`,
`systemctl restart cloudflared` (blips webdesign.uk seconds); DNS **in the dashboard**
(never `route dns` — landmine): webdesign.uk zone → CNAME `links` →
`81f59f78-dda8-40a0-984b-cfadb36bc891.cfargotunnel.com`, Proxied. Verify:
`curl https://links.webdesign.uk/c/x` → core-manager's "no longer active" page;
`curl https://links.webdesign.uk/other` → 404; apex `/c/x` still 302-parked.
**The delivery-email builder must mint links on `links.webdesign.uk` — that is now the
canonical emailed-links host** (recorded in the lane NOTE).

## 3. IMMEDIATE — `www.apis.uk` (walkthrough given 2026-08-24, owner applies)

`www.apis.uk` is `A 192.0.2.1 Proxied` — TEST-NET-1, a dead address; proxied visitors get a
522. Owner wants www → the bees page at the apex (verified live: *"apis.uk — A page about
bees"*). Fix is a **redirect rule, keeping the dummy record as the proxied anchor**:
apis.uk zone → Rules → Create → Redirect Rule: custom filter `Hostname equals
www.apis.uk` → Dynamic redirect `concat("https://apis.uk", http.request.uri.path)`,
**301**, preserve query string. Verify `curl -sI https://www.apis.uk/` → `301` +
`location: https://apis.uk/`. (DNS-only alternatives fail here: a CNAME to the apex lands
on the island tunnel's `*.apis.uk` probe → 404, because cloudflared matches by hostname.)

## 4. NEXT BUILD — the build-steps screen

`PLAN_2026-08-24_build_steps_screen.md`. Short version: a Builds tab + per-site button over
`GET /admin/workflows` / `:correlation_id` / `resume` (all live, **0 SPA references
today**). One backend addition first (a `site_id` filter — orchestration_states has no site
column). ⚠ Build around `bugs_open/099`: a FAILED step can read COMPLETED with `error`
NULL — truth is in `collected_data.__step_error`; a screen rendering `status` verbatim will
show green builds that discarded their design. Verify JSON key paths against live rows
before coding anything.

## 5. OWED / OPEN

1. **Architecture boundary review** — `links.webdesign.uk` will be the SECOND deliberately
   public cluster route (after `admin.apis.uk`). The architecture seat's approval condition
   ("a second and third publicly-proxied prefix should trigger a boundary review") is now
   met. Advisory, not blocking — but it was the seat's stated condition; run one council
   round over the exposure posture when the links host goes live.
2. **Council verdict** on the prefetch guard: `Council-Submitted:
   6b1726ab-35fd-4541-8439-ffb3d699ba6b` (commit `0e9cb31ee`). Read it; act on REVISE.
3. **Mail-scanner residual on `/c/`** — unchanged by everything above; second click vs
   accepted risk, owner's call, becomes live at the first delivery email.
4. **The HOLD ban** (`SQL_2026-08-22b_…_HOLD.sql`) still must NOT run until the 4 pages
   (10 components — census in the file) are re-rendered off "two or three days".
5. **webdesign lane items** (delivery email, terms page, Stripe, etc.):
   `../webdesign_uk_build_service/HANDOFF_2026-08-21_continue_here.md` §6 — still current
   EXCEPT its §2: both owner decisions there are now ANSWERED (D-A 2026-08-22 doc, D-B
   applied). A NOTE in that dir says so.
6. **VPN endpoint fragility + laptop mystery** — parked; runbook has the full diagnosis.
7. `ADM-002` staleness — re-verify any admin route before building UI on it.

## 6. Falsifiers

- Image tags roll daily — `v1.0.1332` and the provenance stamp are today's; re-ask the pod.
- `customer_access_tokens` = 0 and handed_over/confirmed = 0/0 — the moment any is non-zero,
  every "nothing at risk" claim above expires.
- Whether the box has applied §2 (check: `curl https://links.webdesign.uk/c/x`).
- Whether the www redirect rule exists (curl it).
- The council verdict on `6b1726ab` may have landed.
- A newer handoff in this dir or `../webdesign_uk_build_service/`.
