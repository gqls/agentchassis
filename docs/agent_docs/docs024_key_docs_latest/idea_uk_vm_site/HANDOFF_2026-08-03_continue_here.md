# HANDOFF 2026-08-03 — the ingress arc is DONE; one small deletion is parked behind a self-inflicted token lockout

**Written by the "ideauk sec" session (2026-08-01 → 08-03), which took
`HANDOFF_2026-07-31_cloudflare_decision.md` from "two open owner decisions" to
Option B fully live.** Read with: RUNBOOK §4a + §4a-bis (the Progress blocks
are the authoritative state), RUNNING_NOTES §X.34–§X.42 (the evidence),
LANDMINES entries of 08-01 and 08-03 (the two traps now guarding this box).

## State per domain — all verified live, none inferred

| domain | state | evidence |
|---|---|---|
| **idea.uk** | **Option B COMPLETE.** Orange behind Cloudflare, SSL **Full (strict)**, real-IP restoring (`conf.d/cloudflare-realip.conf`), catch-all vhost live, ufw allows 80/443 **only** from the 22 CF ranges, SSH open. NS = alexis/leah via Nominet EPP. | NOTES §X.40: cf-ray 200s, 16-route loop identical to the pre-CF baseline, two-network proof (5.65.164.9 + 116.203.204.115 in access.log, zero CF-range client IPs), direct :80/:443/v6 all TIME OUT from outside CF, `ssh` fine |
| **loanzy.uk** | Delegated to CF via EPP (VMB-015 run 2), zone ACTIVE (60s), edge cert ISSUED (`CN=loanzy.uk` + wildcard). Serves a **correct 522** (~20s): pattern record `A → 199.59.243.228 proxied`, **no Worker route yet**. | NOTES §X.41 + addendum/final |
| **webzy.uk** | **NOT the owner's domain** (GODADDY tag). The CF zone created 08-02 (`aeddc60d81acba81e87d39488011fcd8`) **must be DELETED — owner instruction, NOT yet executed** (see open item 1). No NS change ever happened (impossible: wrong tag) — nothing outside the CF account references it. | NOTES §X.38, §X.42 |

## Open items, in order

1. **Delete the webzy.uk CF zone.** Blocked at session end by a Cloudflare
   auth-failure lockout (`10502`, then zone endpoints refusing with the
   MISLEADING "Cannot use the access token from location: <your ip>" — that
   message appears mid-lockout from BOTH address families and does not prove
   an IP filter). **Cause was self-inflicted**: a burst of refused calls
   (IPv6-sourced + scope probes at cert-packs/universal-settings), and then —
   the key insight — **my recovery watcher polled the locked endpoint with
   the locked token once a minute, feeding the failure counter that caused
   the lockout. Do not poll a failure-counter lockout.** The watcher is
   killed. Next session: **make ONE attempt after ≥1h of API silence** —
   `GET zones/aeddc60d81acba81e87d39488011fcd8` (check `.success` first!),
   then DELETE, then confirm with a list that is **successful AND empty**
   (`{success:true, matches:0}` — a failing list also shapes as empty;
   §X.42 item 1 nearly reported the zone deleted on exactly that).
   Dashboard fallback is ten seconds: webzy.uk → Overview → Delete zone.
   If a clean, post-silence DELETE still 9109s while GETs work, only then
   conclude the token lacks zone-delete permission.
2. **idea.uk residual: the first ORGANIC signed Stripe webhook.** Plumbing is
   proven through the proxy (`/stripe/webhook → 400` = reaches the binary's
   signature check via CF); a synthetically SIGNED event was deliberately not
   fired. After the next real order: check `/var/lib/idea/orders.json` moves
   to paid.
3. **loanzy.uk content**: needs a Worker route / B2 wiring — **webdesign
   lane's machinery**, they are pointed at it (their NOTES, 08-03 correction:
   loanzy IN, webzy OUT).
4. Optional cleanup on the box: `/root/nominet-epp-ns-change.py` was scp'd to
   the VM (never used there; no credential accompanies it). Harmless; remove
   if tidying.

## Credentials & access facts (paths only, no secrets here)

- CF token: `~/.config/cloudflare/token`. Scope: zones read/create/edit-DNS +
  Zone Settings. NOT: certificate_packs, universal SSL settings, and possibly
  zone-delete (unproven — see open item 1). **Pin `curl -4`**: at least one
  v6-sourced call was refused naming the v6 temp address (SLAAC addresses
  rotate daily, so v6 can never be stably allow-listed anyway).
- Nominet EPP: password `~/.config/nominet/epp-password`; tag
  `DESIGNCONSULT`; allow-list has workstation `5.65.164.9` + VM
  `116.203.204.115`. Client: `box/nominet-epp-ns-change.py` (VMB-015,
  2 clean production runs). Dry-run default; **pins IPv4** (a dual-stack
  connect arrives from an address a v4 allow-list never heard of); refusals
  are unframed text; **one login failure = stop** (tag lockout risk).
  **Check the domain's TAG before planning EPP work** — webzy.uk (GODADDY)
  proved the portfolio spans registrars.

## The two rollback landmines (fleet LANDMINES, 08-03 — do not relearn)

- **Grey is NO LONGER a safe rollback for idea.uk.** With the firewall on,
  DNS-only records point visitors at a sealed origin: site down AND certbot
  renewal silently dead. Order: `ufw allow 80/tcp && ufw allow 443/tcp`
  FIRST, grey second.
- **A timeout curling `116.203.204.115` is the FIREWALL working**, not an
  outage. Liveness = `https://idea.uk` (expect cf-ray) or ssh.
- (And from §X.40/§X.41: a proxy-flip probe without `--resolve` can pass
  against the resolver's cached origin; a fresh zone's cert question is
  answered by `openssl s_client`, not by curl's HTTP timeout — the 522 from a
  placeholder origin takes ~20s.)

## Chassis note

The 08-03 chassis roll is IRRELEVANT to this lane: nothing here ships in a
chassis image (box nginx/ufw config, Cloudflare API state, Nominet registry
state, and docs only). No pod-grep owed by this lane.

## Where the full story lives

RUNBOOK §4a Progress blocks (authoritative sequence + evidence) ·
RUNNING_NOTES §X.34–§X.42 (including every misstep: the false-negative TLS
watcher, the poisoned empty list, the v6 traps ×3) · README_where_we_are
(owner-voice history, 1–3 Aug) · WRONG_CALLS 08-03 (the layered-probe lesson)
· concept register VMB-015 (the reusable EPP client) · MEMORY
`idea-uk-vm-site-workstream.md` (cold-start pointers).
