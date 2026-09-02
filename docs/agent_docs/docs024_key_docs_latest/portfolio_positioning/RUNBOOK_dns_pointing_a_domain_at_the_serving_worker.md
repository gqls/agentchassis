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

## ✅ DONE 2026-08-18 20:07Z — `www.` now 301s to the apex on every applicable zone

**36 of 36 applicable zones verified redirecting; 3 deliberately skipped; 0 failures.**
Owner ruling: redirect, do not serve. Two moving parts:

1. **`scripts/cloudflare/worker.js`** carries the redirect (`hostname.startsWith('www.')`
   → 301 to the same URL with `www.` stripped, above the object-key construction).
   Deployed by the owner 20:02:37Z via `scripts/cloudflare/deploy_worker.sh`.
2. **`scripts/cloudflare/add_www_redirect.sh --apply`** adds, per zone, a proxied `www`
   A → 192.0.2.1 and — only where no wildcard route already covers it — a
   `www.<domain>/*` route. **28 DNS records + 7 routes added, 0 failed.**

**It classifies each zone live rather than looping, and that is load-bearing.** Skipped:
`idea.uk` (no route to the worker at all — a proxied A there is a 522 black hole),
`relojistas.com` (www already serves a real page off another host), `webdesign.uk`
(deliberate 302 to webdesign.co.uk). Left alone as already correct: `cookly.uk`,
`dartsonline.com`. Fixed as a side effect: `robot-hands.com` and
`leopardessconsulting.co.uk`, whose `www` records had existed with nothing serving them
and simply hung.

### ⚠ Two things that make a WORKING change look broken — do not undo it on either

- **A newly created worker route 522s for the first few requests.** `www.remortgagecalculator.uk/`
  returned **522** immediately after its route was created, while `…/mortgage-lenders.html`
  on the same host already 301'd. 522 is exactly the signature of "no worker, dead origin"
  (192.0.2.1 is TEST-NET-1 and can never answer), so it reads as *"the redirect is not
  working and the A record is pointing at nothing"*. It settled within a minute: 5/5
  clean on retry. **Re-test before diagnosing; never delete the record on one reading.**
- **Your own resolver's NEGATIVE cache outlives the record you just created.**
  `www.vonc.com` and `www.webdesign.co.uk` returned `curl: (6) Could not resolve host`
  for minutes after their records existed — because the dry run had queried those names
  while they were NXDOMAIN and the negative answer was still cached locally. **"Could not
  resolve" is indistinguishable from "the record was never created."** Ask authoritative
  DNS instead of your resolver, and prove the redirect by pinning the IP:
  ```sh
  curl -s -H 'accept: application/dns-json' "https://1.1.1.1/dns-query?name=www.<domain>&type=A"
  curl -s -o /dev/null -w '%{http_code} -> %{redirect_url}\n' \
       --resolve "www.<domain>:443:<ip-from-above>" "https://www.<domain>/"
  ```
  Both zones were correct all along and read as failures for ~4 minutes.

### Verify (read the redirect, never the API response)

```sh
curl -s -o /dev/null -w '%{http_code} -> %{redirect_url}\n' https://www.<domain>/
# expect: 301 -> https://<domain>/   — path and query string are preserved
```

## ⚠ SUPERSEDED FINDING (kept for the history): `www.` resolves NOWHERE

Measured 2026-08-18: `www.ai-agent-orchestration.com` **does not resolve at all** while the
apex returns 200. There is no `www` record and no `www` route on the reference zone — and by
implication on **any of the 36**. Anyone typing `www.<domain>` gets a DNS failure, not a
redirect.

> **⚠ CORRECTED 2026-08-18: "on any of the 36" was an INFERENCE from one zone, and it was
> wrong in both directions.** Measured across all 39: **8 zones already carried a `www`
> record**, in FOUR different states — two redirecting correctly (`cookly.uk`,
> `dartsonline.com`), one deliberately pointing elsewhere (`webdesign.uk` → webdesign.co.uk),
> one serving a real page off another host (`relojistas.com`), and two simply hanging
> (`robot-hands.com`, `leopardessconsulting.co.uk`). A uniform claim about a fleet, drawn
> from the reference zone, is the shape to distrust — the exceptions are exactly the zones a
> blind fan-out would have broken.

**Resolved by the owner 2026-08-18: redirect www to the apex.** Implemented in the worker
rather than per-zone redirect rules, because the portfolio token has no ruleset permission
(`http_request_dynamic_redirect` → `Authentication error`) and no account scope. See the
DONE section at the top of this file.

## ⚠ Addendum 2026-09-02 — two field lessons from the first four-domain cutover

1. **Order bit us exactly as §Steps warns:** the NS change went in at Nominet BEFORE zones
   existed → all four domains lame-delegated (registry SERVFAIL "no reachable authority"),
   old sites dark, for the hours until the owner created the zones. Zone FIRST, always.
2. **The 404-token's zone list is not evidence.** With all four domains serving through the
   worker, `GET /zones?name=<domain>` under `CLOUDFLARE_API_404_TOKEN` returned NO ZONE for
   every one — the token's zone visibility is scoped (or the zones live under a different
   view). Judge a cutover at the served body (proxy IPs + the page title), never at that
   token's empty list.
