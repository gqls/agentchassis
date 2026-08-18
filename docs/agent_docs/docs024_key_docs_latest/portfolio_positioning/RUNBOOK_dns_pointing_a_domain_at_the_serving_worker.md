# RUNBOOK — pointing a portfolio domain at the serving worker (Cloudflare + Nominet)

Written 2026-08-18 for the Phase C pilot (`remortgagecalculator.uk`) and the Phase E waves
after it. **Owner ruling 2026-08-18: domains go to Cloudflare in bulk groups; nmsvr.uk is
parked** (own nameservers would take the worker out of the request path).

## How serving actually works (measured, not assumed)

A site is served by the Cloudflare Worker **`portfolio-sites-router`**, which signs a request
to Backblaze B2 and returns the object at key **`<hostname><path>`** from the bucket
**`portfolio-sites`** (`scripts/cloudflare/worker.js`). Two things must therefore be true for
`https://example.uk/page.html` to serve:

1. `example.uk` is a **Cloudflare zone** (its nameservers point at Cloudflare), with a
   **proxied** DNS record at the apex — measured: working sites resolve to Cloudflare proxy
   IPs (`104.21.x` / `172.67.x`, `NetName: CLOUDFLARENET`).
2. A **worker route** `example.uk/*` → `portfolio-sites-router` exists on that zone —
   measured on the working zone via the API.

Account: `13044f178ae0b156961065f55c8fada8`. 36 zones already work this way.

## The check that tells you a domain is NOT pointed yet

```sh
dig +short NS example.uk
```
- `alexis.ns.cloudflare.com` / `leah.ns.cloudflare.com` (or another Cloudflare pair) → pointed.
- `ns1.dan.com` / `ns2.dan.com` → **still at the registrar's parking service.**

> ⚠ **DO NOT use an HTTP status code for this check.** A parked domain returns **200 on every
> path** with a lander body (`<script>window.onload=…"/lander"</script>`). `curl -o /dev/null
> -w '%{http_code}'` against a parked domain is a check that cannot fail. **Read the body.**

## Steps

**1. Create the zone.** The existing session token (`~/.cloudflare/404-token.env`,
`CLOUDFLARE_API_404_TOKEN`) is **Workers-scoped**: it verifies, reads zones and reads/writes
worker routes, but **cannot read DNS records and cannot create zones** — measured 2026-08-18:
`Requires permission "com.cloudflare.api.account.zone.create"`.

> **CORRECTED 2026-08-18 — the permission is NOT under the "Account" group.** I first told the
> owner to add *Account → Zone → Edit*; he checked and there is no such entry (the Account
> group lists Workers Scripts, Workers Tail, Zero Trust…). **The zone-creation right comes from
> the ZONE group, scoped to the whole account:**
>
> - set the first dropdown to **Zone** (not Account), then **Zone → Edit**
> - add **Zone → DNS → Edit**
> - add **Zone → Workers Routes → Edit**
> - **under *Zone Resources*, choose `Include → All zones from an account → <the account>`.**
>   This is the load-bearing part: a token scoped to *specific zones* cannot create a zone that
>   does not exist yet. The account-wide scope is what grants
>   `com.cloudflare.api.account.zone.create` despite the permission living under "Zone".
>
> Account: `13044f178ae0b156961065f55c8fada8`. Save as `CLOUDFLARE_API_TOKEN=…` in
> `~/.cloudflare/dns-token.env` (mode 600).

With that token I can script steps 1, 4 and 5 for a whole batch; **step 3 (Nominet) stays
manual** — it is registrar-side and there is no access from here.

**2. Note the two assigned nameservers** shown after zone creation.

**3. Set those nameservers at Nominet (OWNER — no access from here).** This is the bulk step:
Nominet allows changing NS for a group of domains, so do a category at a time (the finance
group first: `remortgagecalculator.uk` + the M-family siblings).

**4. Add the proxied apex record** (owner, or me with a DNS-scoped token): an `A` record at the
apex, **proxy enabled** (orange cloud). The address barely matters — the worker intercepts
before origin — but it must be proxied, or the worker never sees the request.

**5. Add the worker route (I CAN DO THIS with the current token):**
```sh
. ~/.cloudflare/404-token.env
ZID=$(curl -s "https://api.cloudflare.com/client/v4/zones?name=<domain>" \
  -H "Authorization: Bearer $CLOUDFLARE_API_404_TOKEN" \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['result'][0]['id'])")
curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$ZID/workers/routes" \
  -H "Authorization: Bearer $CLOUDFLARE_API_404_TOKEN" -H "Content-Type: application/json" \
  --data '{"pattern":"<domain>/*","script":"portfolio-sites-router"}'
```

**6. Verify at the BODY, never the status:**
```sh
curl -s "https://<domain>/<a-deployed-page>.html" | head -c 200
```
Expect the page's real markup. If you see the `/lander` redirect, DNS has not propagated (or
step 3 has not happened). **Token expiry: 2026-09-01** — re-mint before the Phase E waves.

## State at time of writing

`remortgagecalculator.uk`: pages `mortgage-lenders`, `next-steps`, `about` are **deployed to
the bucket**; the domain is still on `ns1/ns2.dan.com`, so nothing is reachable. Steps 1–3 are
the owner's; ping me and I will do 5 and verify 6.


---

# PROVEN RECIPE — executed 2026-08-18 on `remortgagecalculator.uk`

Token: `~/.config/cloudflare/portfoliotoken` (bare token, one line — read it with
`T=$(tr -d ' \r\n' < ~/.config/cloudflare/portfoliotoken)`). Scope: All zones —
Zone:Edit, DNS:Edit, Workers Routes:Edit. **Active, no expiry.** It does NOT carry
Account:Read, so `GET /accounts` returns empty — that is expected, not a fault; use the
account id directly.

Account: `13044f178ae0b156961065f55c8fada8`.

```sh
T=$(tr -d ' \r\n' < ~/.config/cloudflare/portfoliotoken); ACC=13044f178ae0b156961065f55c8fada8; D=<domain>

# 1. zone  -> returns the two nameservers for Nominet
ZID=$(curl -s -X POST "https://api.cloudflare.com/client/v4/zones" -H "Authorization: Bearer $T" \
  -H "Content-Type: application/json" --data "{\"name\":\"$D\",\"account\":{\"id\":\"$ACC\"},\"type\":\"full\"}" \
  | python3 -c "import sys,json;d=json.load(sys.stdin);print(d['result']['id']) if d['success'] else print(d['errors'])")

# 2. ONE proxied apex A record. The address is irrelevant — the worker answers before origin.
curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$ZID/dns_records" -H "Authorization: Bearer $T" \
  -H "Content-Type: application/json" \
  --data "{\"type\":\"A\",\"name\":\"$D\",\"content\":\"192.0.2.1\",\"proxied\":true,\"ttl\":1}"

# 3. ONE worker route
curl -s -X POST "https://api.cloudflare.com/client/v4/zones/$ZID/workers/routes" -H "Authorization: Bearer $T" \
  -H "Content-Type: application/json" --data "{\"pattern\":\"$D/*\",\"script\":\"portfolio-sites-router\"}"
```

**Match the reference config exactly — one A record, one route.** I initially also added a
proxied `www` CNAME; then read the working zone (now possible with DNS:Read) and found it has
**one apex A record and nothing else**. A `www` CNAME with no matching `www.<domain>/*` worker
route proxies to the placeholder origin and fails, which is worse than not existing. Removed.

**`192.0.2.1` is deliberate** — TEST-NET-1 (RFC 5737), which can never route anywhere. The
reference zone happens to use a parking IP (`199.59.243.228`); either works because the worker
intercepts first, but a documentation-reserved address cannot accidentally reach a live host.

## ⚠ FLEET-WIDE FINDING: `www.` resolves NOWHERE

Measured 2026-08-18: `www.ai-agent-orchestration.com` **does not resolve at all** while the
apex returns 200. There is no `www` record and no `www` route on the reference zone — and by
implication on **any of the 36**. Anyone typing `www.<domain>` gets a DNS failure, not a
redirect.

**Not fixed here, deliberately** — it is a fleet-wide convention, and changing it on one zone
would diverge that zone from the other 35. **It is a decision for the owner:** add a proxied
`www` CNAME plus a `www.<domain>/*` worker route to every zone (the worker keys on
`<hostname><path>`, so `www.` would need its own bucket prefix or a redirect worker), or
accept that these sites are apex-only.
