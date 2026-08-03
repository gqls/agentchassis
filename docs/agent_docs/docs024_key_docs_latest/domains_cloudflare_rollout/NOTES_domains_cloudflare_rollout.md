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
