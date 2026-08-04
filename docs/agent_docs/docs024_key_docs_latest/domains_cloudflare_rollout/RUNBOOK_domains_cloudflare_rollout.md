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

## Egress IPs (for IP allowlists)

> **CORRECTED 2026-08-04:** this section said the IPv6 /64 was stable and named
> `151.226.83.138` as "the" IPv4. **The office line rotates BOTH families
> wholesale** (by 08-04: `5.65.164.9` and `2a02:c7e:3066:5400::/64`) — never pin
> an allowlist or token filter to this line. WRONG_CALLS 2026-08-04 has the row.

- Re-check current egress: `curl -4 -s https://www.cloudflare.com/cdn-cgi/trace | grep ^ip=` (and `-6`).
- **Stable addresses live in the k8s cluster**: each node egresses via its own
  public IP — `kubectl get nodes -o wide` → `134.213.168.26/.37/.44/.54/.56`
  (measured 08-04; `postgres-clients-0` sits on the `.26` node). Allowlist the
  node set at Nominet and run EPP from inside the cluster (pipe local framing
  through `kubectl exec -i postgres-clients-0 -- openssl s_client -connect
  epp.nominet.org.uk:700 -quiet` — OpenSSL 3.0.20 confirmed present there).
- **LANDMINE (also in LANDMINES.md):** the two obvious health checks do NOT
  exercise the allowlists. Cloudflare `/user/tokens/verify` returns 200 for a
  token whose IP filter 403s (`code 9109`) every real endpoint — prove the token
  with `GET /zones?per_page=1`. Nominet serves the EPP **greeting to any IP** —
  only LOGIN proves allowlisting. The 9109 message names the address you are
  actually egressing from; read it. Dual-stack makes failures intermittent
  (family chosen per connection): pin scripts to one family. Python's default
  urllib UA also draws a WAF 403 on some endpoints — send a Mozilla-prefixed UA.

## Nominet (member tag, EPP)

- `epp.nominet.org.uk:700`, TLS, XML EPP (RFC 5734 framing: 4-byte big-endian
  length prefix incl. itself, then the XML). Full greeting confirmed 2026-08-04.
- **Pin to IPv4** — the two families get different treatment (IPv6 got a 94-byte
  brush-off where IPv4 got the 2,527-byte greeting).
- Login refused until the egress IP is allowlisted in Online Services, regardless
  of credentials — **and the greeting is served to ANY IP, so only a login tests
  the allowlist** (see Egress section).
- Credential state: password at `~/.config/nominet/epp-password` (single line);
  **TAG still needed** as of 2026-08-04.
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

- Keys: porkbun.com/account/api. ~~IP-restrict the key there~~ **CORRECTED
  2026-08-04: do NOT IP-restrict to the office line — it rotates (see Egress).**
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
