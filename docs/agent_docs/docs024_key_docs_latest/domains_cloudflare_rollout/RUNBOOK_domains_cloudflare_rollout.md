# RUNBOOK — domains → Cloudflare rollout

## Credentials (all chmod 600, NEVER in the repo — shared tree)

| service | file | contents |
|---|---|---|
| Cloudflare | `~/.config/cloudflare/token` | the token, one line. PRESENT + verified |
| Nominet | `~/.config/nominet/credentials` | `TAG=…` and `EPP_PASSWORD=…` lines. PENDING |
| Dynadot | `~/.config/dynadot/credentials` | `API_KEY=…`. PENDING |
| Porkbun | `~/.config/porkbun/credentials` | `API_KEY=…` and `SECRET_API_KEY=…`. PENDING |
| Spaceship | `~/.config/spaceship/credentials` | `API_KEY=…` and `API_SECRET=…`. PENDING |

Gotcha: read the token with `tr -d '[:space:]'` — a trailing newline in the file
becomes an invalid bearer header.

## Cloudflare (verified 2026-08-02/03)

- Account ID `13044f178ae0b156961065f55c8fada8`; 36 zones pre-existing, all Free.
- Verify token: `curl -s -H "Authorization: Bearer $TOKEN" https://api.cloudflare.com/client/v4/user/tokens/verify`
- Token scopes: Zone:Edit, DNS:Edit, Workers Routes:Edit (all zones in account) +
  Workers Scripts:Read, Account Settings:Read. Zone-create not yet exercised —
  first real zone proves it.
- Template (from live `dartsonline.com`): apex `A 199.59.243.228` proxied;
  routes `<domain>/*` and `*.<domain>/*` → script `portfolio-sites-router`.
  NO www/wildcard record exists (open question in PLAN).
- Account NS pair observed: `alexis.ns.cloudflare.com` / `leah.ns.cloudflare.com`
  — still capture `name_servers` from each zone-create response, do not assume.
- Rate limit 1,200 req/5 min account-wide.

## Egress IPs of this machine (for IP allowlists)

- IPv4 `151.226.83.138` · IPv6 `2a02:c7c:f61f:ac00::/64` (low half rotates — use the /64).
- Re-check: `curl -s https://www.cloudflare.com/cdn-cgi/trace | grep ^ip=` (and `curl -4 …`).
- Nominet EPP allowlist is IPv4 — register `151.226.83.138`.

## Nominet (member tag, EPP)

- `epp.nominet.org.uk:700`, TLS, XML EPP. Port reachable from here (tested 2026-08-02).
- Login refused until the egress IP is allowlisted in Online Services, regardless
  of credentials.
- Domain inventory: Nominet's EPP list extension enumerates by registration or
  expiry month — walk 12 expiry months for full coverage — OR the owner exports a
  CSV from Online Services (preferred: gives a checkable total).
- NS repoint: EPP `domain:update` add/rem ns per domain.

## Dynadot

- API key: control panel → Tools → API. Legacy API3: `https://api.dynadot.com/api3.json?key=…&command=…`.
- `list_domain` = full inventory (paginated). `set_ns` sets nameservers,
  comma-separates multiple domains per call — **but the target nameservers must
  already exist in the account** (`add_ns` them once, first).
- Rate tier by spend level: Regular = 1 thread, 60 req/min. Fine for thousands
  (NS updates batch multiple domains per request anyway).

## Porkbun

- Keys: porkbun.com/account/api (can also IP-restrict the key there — do it).
- Auth: `X-API-Key` / `X-Secret-API-Key` headers or `apikey`/`secretapikey` in
  the JSON body. Base `https://api.porkbun.com/api/json/v3`.
- Full endpoint reference: `https://porkbun.com/llms-full.txt`.
- [ASSUMED] API access may need enabling per-domain (bulk toggle in Domain
  Management) — the docs page didn't confirm; the error on a disabled domain is
  explicit, so the first listing/update call settles it.

## Spaceship

- Keys: dashboard → API Manager → "New API key". Grant `domains:read` +
  `domains:write` only.
- Base `https://spaceship.dev/api/v1`; auth headers `X-Api-Key` / `X-Api-Secret`.
- `GET /v1/domains?take=100&skip=N` lists; `PUT /v1/domains/{domain}/nameservers`
  repoints. Rate limits: list 300/300s; NS updates **5 per domain**/300s; most
  ops 30/30s per user.

## Verification pattern (end of run)

Census, not spot-check: every in-scope domain must appear as an `active` zone
with the A record, both routes, and registrar/EPP NS matching the pair Cloudflare
assigned. A `pending` zone after NS repoint → trigger the activation check
(`PUT /zones/{id}/activation_check`), then re-poll.
