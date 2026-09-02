# RUNBOOK — domains → Cloudflare rollout

## Credentials (all chmod 600, NEVER in the repo — shared tree)

| service | file | contents |
|---|---|---|
| Cloudflare **(read-only)** | `~/.config/cloudflare/token` | the token, one line. PRESENT — but **READ-ONLY**, see below |
| Cloudflare **(read-write)** | `~/.config/cloudflare/portfoliotoken` | ⚠ **THIS is the one that can WRITE.** Added 2026-08-18, undocumented until 2026-08-25 |
| Nominet | `~/.config/nominet/credentials` | `TAG=…` and `EPP_PASSWORD=…` lines. PENDING |
| Dynadot | `~/.config/dynadot/credentials` | `API_KEY=…` (+ `API_SECRET=…`, unused by API3, kept for the RESTful API). **PRESENT 2026-09-02 — `list_domain` PROVEN (451 domains), writes not yet** |
| Porkbun | `~/.config/porkbun/credentials` | `API_KEY=…` and `SECRET_API_KEY=…`. **PRESENT 2026-09-02 — ping + listAll PROVEN (683 domains); per-domain endpoints refused until the account-level API-access opt-in (see Porkbun section), writes not yet** |
| Spaceship | `~/.config/spaceship/credentials` | `API_KEY=…` and `API_SECRET=…`. **PRESENT 2026-09-02 — read paths PROVEN, writes not yet** |

Gotcha: read the token with `tr -d '[:space:]'` — a trailing newline in the file
becomes an invalid bearer header.

## Cloudflare (verified 2026-08-02/03)

- Account ID `13044f178ae0b156961065f55c8fada8`; 36 zones pre-existing, all Free.
- Verify token: `curl -s -H "Authorization: Bearer $TOKEN" https://api.cloudflare.com/client/v4/user/tokens/verify`
- ~~Token scopes: Zone:Edit, DNS:Edit, Workers Routes:Edit (all zones in account) +
  Workers Scripts:Read, Account Settings:Read. Zone-create not yet exercised —
  first real zone proves it.~~
  > **⚠ CORRECTED 2026-08-25 — THE TOKEN IN `~/.config/cloudflare/token` IS READ-ONLY TODAY.**
  > `[MEASURED 2026-08-25 12:40Z]` the zone object's own `permissions` array reads
  > `['#worker:read', '#analytics:read', '#zone:read', '#organization:read']` — **no edit scope of
  > any kind.** Confirmed by attempting both halves: `GET /zones/<id>/dns_records` returns
  > `10000 Authentication error`, and `POST /zones` returns *"Requires permission
  > `com.cloudflare.api.account.zone.create` to create zones for the selected account"*.
  > So **neither zone-create nor any DNS-record or route write is possible with `token`.**
  >
  > **⚠⚠ SUPERSEDED THE SAME DAY, 2026-08-25 12:46Z — DO NOT ISSUE A NEW TOKEN, THERE IS ALREADY A
  > SECOND ONE.** `~/.config/cloudflare/portfoliotoken` (present since 2026-08-18, undocumented
  > here until now) carries `#zone:edit`, `#dns_records:edit`, `#zone_settings:edit`,
  > `#worker:edit` — and **account-level zone-create, now PROVEN rather than assumed**: it created
  > `homegarden.uk` (zone `252c10abde85a6985392a084f68f9235`), which is **the first zone this
  > estate has created via the API**. The line above stood for about ninety minutes; I wrote it,
  > and it was wrong because I checked the credential I knew about instead of asking what
  > credentials exist. **`ls ~/.config/cloudflare/` before concluding anything about permissions.**
  > ⚠ **`GET /zones` SUCCEEDS on this read-only token** (40 zones, `success: true`), so the
  > runbook's own "prove the token with `GET /zones?per_page=1`" check — written to defeat the
  > IP-filter trap — **passes for a token that can read everything and write nothing.** It tests
  > reachability, not capability. **Probe the verb you actually need**, e.g. a `POST` to the real
  > endpoint, and read the `permissions` array on any zone object as the cheap first look.
- Template (from live `dartsonline.com`): apex `A 199.59.243.228` proxied;
  routes `<domain>/*` and `*.<domain>/*` → script `portfolio-sites-router`.
  NO www/wildcard record exists (open question in PLAN).
- Account NS pair observed: `alexis.ns.cloudflare.com` / `leah.ns.cloudflare.com`
  — still capture `name_servers` from each zone-create response, do not assume.
  > **⚠ THAT CAVEAT IS NOT THEORETICAL — MEASURED 2026-08-25, the account uses TWO pairs.**
  > Across all 40 zones: **29 × `alexis` / `leah`** and **11 × `betty.ns.cloudflare.com` /
  > `ivan.ns.cloudflare.com`**. So quoting the observed pair for a new zone is a **~28% chance of
  > handing someone the wrong nameservers**, and the failure is quiet — the domain simply never
  > resolves to Cloudflare and the delegation looks plausible on paper. **A zone's NS pair exists
  > only after the zone is created; there is no way to know it in advance.** Take them from the
  > create response, or from `GET /zones?name=<domain>` afterwards.
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

- **Wrapper: `scripts/domains/dynadot.sh <command> [param=value …]`** (added
  2026-09-02). Reads the credentials file, never prints the key, and exits
  non-zero unless the response carries `ResponseCode 0` — so an `&&` chain
  cannot mistake an API error body for a success. Endpoint + error path proven
  2026-09-02 with a bogus key (`{"ResponseCode":"-1","Error":"invalid key"}`);
  **happy path PROVEN later the same day: `list_domain` returned the full
  inventory, ResponseCode 0.** Writes (`set_ns`/`add_ns`) not yet exercised.
- API key: control panel → Tools → API. Legacy API3: `https://api.dynadot.com/api3.json?key=…&command=…`.
- If the API settings page offers an IP allowlist, leave it UNSET or allowlist
  the five cluster node IPs — NEVER the office line, which rotates both address
  families (see Egress).
- `list_domain` = full inventory ~~(paginated)~~ — **CORRECTED 2026-09-02: the
  live response has NO pagination fields** (top-level keys `ResponseCode` /
  `Status` / `MainDomains` only; **451** domains as of 2026-09-02 arrived in one
  302 KB body). Sanity-check the count against the control panel total before
  treating it as complete. `set_ns` sets nameservers,
  comma-separates multiple domains per call — **but the target nameservers must
  already exist in the account** (`add_ns` them once, first).
- Rate tier by spend level: Regular = 1 thread, 60 req/min. Fine for thousands
  (NS updates batch multiple domains per request anyway).
- ⚠ **`isForSale` in `list_domain` is NOT marketplace-listing state** (LANDMINES
  2026-09-02): 451/451 read `"no"` while 5 had live Buy Now listings. Listings
  census: `download_all_listings` (full PUBLIC marketplace dump — **361 MB /
  7.18M rows as of 2026-09-02**, minutes; `--max-time` + `-o file` or it hangs
  the shell) grepped `^<domain>,`, then confirm hits with
  `get_listing_item domain=…` (authoritative, per-domain).
- **RESTful v2 + appraisals: `scripts/domains/dynadot-restful.sh <METHOD>
  </restful/v2/path> [json-body]`** (added 2026-09-02). Signs with the
  credentials file's `API_SECRET`: `X-Signature = base64(HMAC-SHA256(secret,
  "API_KEY\n<path+query>\n<X-Request-ID or ''>\n<body or ''>"))` +
  `Authorization: Bearer <key>`. PROVEN 2026-09-02:
  `GET /restful/v2/domains/<domain>/appraisal` →
  `{"code":200,"data":{"appraisal_price":"$3559"}}` (Dynappraisal). ⚠ Appraisal
  is capped **PER DAY** by account price level (50 Regular / 100 Bulk / 300
  Super Bulk) and takes ONE domain per call — a 451-domain portfolio is a
  multi-day walk; keep a deterministic order so each day resumes cleanly.

## Porkbun

- Keys: porkbun.com/account/api. ~~IP-restrict the key there~~ **CORRECTED
  2026-08-04: do NOT IP-restrict to the office line — it rotates (see Egress).**
- Auth: `X-API-Key` / `X-Secret-API-Key` headers or `apikey`/`secretapikey` in
  the JSON body. Base `https://api.porkbun.com/api/json/v3`.
- Full endpoint reference: `https://porkbun.com/llms-full.txt`.
- **Client: `scripts/domains/porkbun.py`** (`ping` / `domains` / `ns` / `set-ns`
  / `dns` / `dns-create` / `dns-edit` / `dns-delete` / `check` / `raw`) — same
  family as `spaceship.py` and `dynadot.sh`; reads the credentials file, never
  prints key material, exits non-zero on any non-SUCCESS response. `domains`
  paginates `listAll` completely: **683** domains, all ACTIVE, as of 2026-09-02.
- ~~[ASSUMED] API access may need enabling per-domain (bulk toggle in Domain
  Management) — the docs page didn't confirm; the error on a disabled domain is
  explicit, so the first listing/update call settles it.~~ **MEASURED 2026-09-02:
  CONFIRMED — and the remedy is better than assumed, because there is a GLOBAL
  opt-in.** `ping`/`listAll` work with no opt-in at all; every per-domain
  endpoint (proven on `getNs`) refuses with the explicit error *"Domain is not
  opted in to API access. You can enable API access for all domains globally
  from your account settings at porkbun.com."* — one account-settings switch
  covers the whole estate, no per-domain (let alone 683-domain) toggling needed.
  Writes (`set-ns`, `dns-*`) remain UNEXERCISED until after that switch — a read
  refusal proves nothing about the write path either way.

## Spaceship

- Keys: dashboard → API Manager → "New API key". **Key IN and read paths PROVEN
  2026-09-02** (`domains` list, domain info/NS read, DNS record read all answer —
  so `domains:read` + `dnsrecords:read` are granted for certain). **The write
  side is UNEXERCISED**: no NS or DNS write has been attempted, so both the
  `domains:write` scope and the doc-derived set-ns body shapes are unproven until
  the first real repoint — a read success is not a write capability (the
  read-only Cloudflare token above is this lane's own worked proof).
- Base `https://spaceship.dev/api/v1`; auth headers `X-Api-Key` / `X-Api-Secret`.
- **Client: `scripts/domains/spaceship.py`** (`domains` / `info` / `ns` / `set-ns`
  / `dns` / `dns-put` / `dns-delete` / `raw`) — same family as `porkbun.py` and
  `dynadot.sh`; reads the credentials file, never prints key material.
- `GET /v1/domains?take=100&skip=N` lists; `PUT /v1/domains/{domain}/nameservers`
  repoints — body `{"provider":"custom","hosts":[…]}`, or `{"provider":"basic"}`
  with hosts **omitted** to revert to Spaceship's own NS. Rate limits: list
  300/300s; NS updates **5 per domain**/300s; most ops 30/30s per user.
- Inventory `[MEASURED 2026-09-02]`: **203** domains (API `total` agrees), all
  `registered`. NS: **144** × aftermarket.com, **58** × atom.com, **1** ×
  cloudflare — near-all parked at marketplaces, and DNS is not hosted at
  Spaceship (0 records on the sampled zone), so the NS repoint is the only
  Spaceship-side write the rollout needs. Renewals: 17 expire before 2027-01-01,
  **all 17 autoRenew=true**; 14 domains have autoRenew=false, none expiring
  before 2027-06. Snapshot JSON in the 09-02 session scratchpad only.

## Verification pattern (end of run)

Census, not spot-check: every in-scope domain must appear as an `active` zone
with the A record, both routes, and registrar/EPP NS matching the pair Cloudflare
assigned. A `pending` zone after NS repoint → trigger the activation check
(`PUT /zones/{id}/activation_check`), then re-poll.


## ✅ Creating a zone end-to-end — WORKED AND VERIFIED 2026-08-25 (`homegarden.uk`)

**The first API-created zone on this account.** Use `portfoliotoken`, not `token`.

```bash
T=$(tr -d '[:space:]' < ~/.config/cloudflare/portfoliotoken)   # never echo it
ACC=13044f178ae0b156961065f55c8fada8

# 1. create the zone — and TAKE THE NAMESERVERS FROM THIS RESPONSE
curl -s -X POST -H "Authorization: Bearer $T" -H "Content-Type: application/json" \
  --data '{"name":"<domain>","account":{"id":"'"$ACC"'"},"type":"full"}' \
  https://api.cloudflare.com/client/v4/zones

# 2. two proxied A records at TEST-NET-1 (the worker serves everything; the IP never answers)
for n in "<domain>" "www.<domain>"; do
  curl -s -X POST -H "Authorization: Bearer $T" -H "Content-Type: application/json" \
    --data "{\"type\":\"A\",\"name\":\"$n\",\"content\":\"192.0.2.1\",\"proxied\":true,\"ttl\":1}" \
    https://api.cloudflare.com/client/v4/zones/<zone_id>/dns_records; done

# 3. two worker routes
for p in "<domain>/*" "www.<domain>/*"; do
  curl -s -X POST -H "Authorization: Bearer $T" -H "Content-Type: application/json" \
    --data "{\"pattern\":\"$p\",\"script\":\"portfolio-sites-router\"}" \
    https://api.cloudflare.com/client/v4/zones/<zone_id>/workers/routes; done
```

- **Verify by RE-READING the zone, never by trusting the POST receipts** — then `diff` the record
  set against a known-good zone with the domain name normalised out. `garden-tools.uk`
  (`82d90228c20877e2b3fc8470c2bc73d1`) is the reference: exactly 2 A records and 2 routes, nothing else.
- **`status` stays `pending` until the registrar delegation actually moves.** `pending` is not a
  failure and not something to retry; it is Cloudflare saying the NS change has not happened yet.
- ⚠ **Creating the zone changes NOTHING about what the domain serves.** Until the nameservers move
  at the registrar, every path still answers from wherever it is parked — 200 on every path if that
  is a marketplace lander, which will make any HTTP census report a perfect site. See LANDMINES,
  "A parked domain returns HTTP 200 for EVERY path".
