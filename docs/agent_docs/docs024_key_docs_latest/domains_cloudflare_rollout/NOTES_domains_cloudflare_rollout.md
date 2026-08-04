# NOTES — domains → Cloudflare rollout (append-only, newest at the bottom)

## 2026-08-02

- Owner asked whether we can drive Cloudflare + Nominet for thousands of domains.
  Verified from this machine: `api.cloudflare.com` reachable (403 unauth = good);
  `epp.nominet.org.uk:700` TCP-reachable. No credentials existed anywhere
  (checked env + `~/.cloudflare` + `~/.wrangler` + `~/.config/cloudflare`).
- Egress: IPv4 `151.226.83.138`, IPv6 `2a02:c7c:f61f:ac00::/64`.
- Advised a custom Cloudflare token (the dashboard's pre-configured templates
  cannot create zones — "Edit zone DNS" is edit-only on existing zones).
- Owner created the token, stored at `~/.config/cloudflare/token` (was 664,
  tightened to 600). Verified: active. Account `13044f178ae0b156961065f55c8fada8`,
  36 zones, all Free plan.
- Template zone read (`dartsonline.com`): ONE DNS record only (apex
  `A 199.59.243.228` proxied) + routes `dartsonline.com/*` and
  `*.dartsonline.com/*` → `portfolio-sites-router` (the account's only Worker).
  Noted to owner: the wildcard route is dead today — no `www`/`*` DNS record, so
  nothing but the apex resolves. Open question.

## 2026-08-03

- Owner: also wire up Dynadot, Porkbun, Spaceship. All three API hosts reachable
  from here (`api.dynadot.com` 200; `api.porkbun.com` 404-on-root = up;
  `spaceship.dev` 401 = up, wants auth; `api.spaceship.dev` does not resolve —
  the API host IS `spaceship.dev`).
- Docs read (fetched live): auth models + rate limits per registrar → RUNBOOK.
  Notable: Dynadot `set_ns` requires the nameservers pre-registered in the
  account; Spaceship NS updates limited to 5/domain/300s; Porkbun full reference
  lives at `/llms-full.txt`.
- [ASSUMED] Porkbun per-domain "API Access" toggle requirement — docs page did
  not confirm or deny; flagged in RUNBOOK rather than stated as fact.
- Still waiting on: Nominet TAG+EPP password & IP allowlisting, three registrar
  keys, skip-list, www decision.
- MISSTEP (08-03): appended the coverage-ratchet line via `cat >>` at a guessed
  path — the shell created a stray file instead of erroring (the real one is
  `docs026_concept_register/102_coverage_ratchet.txt`). Caught by the pathspec
  commit refusing an untracked file. The check: `git ls-files --error-unmatch`
  the target BEFORE a shell append; or use the Write/Edit tools, which refuse
  unread files.

## 2026-08-04

- Owner answered all three open decisions → PLAN "Owner decisions": skip-list
  (relojistas.com, finetuning.uk, webdesign.uk, idea.uk), the static-first rule,
  and www = proxied CNAME to apex.
- Owner added `151.226.83.138` to the Nominet EPP allowlist; EPP password landed
  at `~/.config/nominet/epp-password` (single value — TAG still missing).
- EPP greeting read cleanly over IPv4 (2,527 bytes, svID "Nominet EPP server");
  IPv6 got a 94-byte reject → pin EPP to IPv4.
- > **CORRECTED 2026-08-04 (two of my own claims, same session):** (1) "the
  > greeting proves the allowlist works" — FALSE, the full greeting later arrived
  > from `5.65.164.9`, never allowlisted; only login tests it. (2) 08-02's "the
  > IPv6 /64 is stable, pin to it" — FALSE, the line rotates both families
  > wholesale. Both in WRONG_CALLS.md 2026-08-04; distilled into a LANDMINES.md
  > entry (health checks that don't exercise the allowlist).
- Zone audit attempted: `audit_zones.py` (copied into this dir) — read-only
  audit of all 36 zones vs template. BLOCKED: every real endpoint returns 9109
  ("Cannot use the access token from location …") — the token's IP filter names
  the rotated-away office IPs. `/user/tokens/verify` still says 200 — it does
  not enforce the filter. A wasted hour chasing a phantom User-Agent theory
  before reading the 403 BODY, which named the cause in one line.
- Cluster egress measured for a stable anchor: five node IPs
  `134.213.168.26/.37/.44/.54/.56` (per-node egress, no shared NAT;
  `postgres-clients-0` on the `.26` node, has OpenSSL 3.0.20 for the EPP pipe).
  Pod creation is permission-blocked; `kubectl exec` into existing pods works.
