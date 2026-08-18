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
