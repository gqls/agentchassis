# HANDOFF 2026-08-03 — P1 shopfront: the page is built and proven, and two Cloudflare clicks are all that stand between it and being live

**Read `PLAN §6a` (wildcard, RESOLVED), `§7a` (offer), `§7d` (price), `§9` (phasing),
and `RUNBOOK` "Cloudflare over the API".** Chassis live at **v1.0.1237** (both
replicas, deployed 2026-08-03 08:47).

---

## 1. State in one paragraph

The webdesign.uk shopfront **exists, is uploaded, and is proven to render** — but it
is **not yet reachable at webdesign.uk**, because the Cloudflare API token stopped
accepting my IP mid-task. `webdesign.uk` still serves last night's holding 302 to
`webdesign.co.uk`. Two changes in the Cloudflare dashboard (§3) make it live; both
take about a minute and neither needs code.

## 2. What is done, with evidence

| thing | state | evidence |
|---|---|---|
| Shopfront page written | **done** | `portfolio-sites/webdesign.uk/index.html`, 12,973 B |
| Page renders through the real Worker→B2 path | **PROVEN** | `https://preview.ugg2.com/` → **200**, 12,973 B byte-identical, `<title>`, `£1,200`, guarantee and contact all present |
| Price £1,200 | **owner-confirmed 2026-08-03** | chose "Full fake door with the price" |
| Contact details | **real, not invented** | `webdesign@contactforsales.com` + `+44 (0) 7934 524 911` — the estate convention, `sites` rows for 12 domains |
| ugg2.com wildcard | live, now proven with a **200** | previously only 404-with-`objectKey` |
| webdesign.uk live | **BLOCKED** | still `302 → webdesign.co.uk`; apex A is `192.0.2.1` |

**The page deliberately does NOT take money.** There is no checkout and no chat —
both need the box (§4). Intent is captured by a `mailto:` with a pre-filled body
asking for domain + what the business does. That is weaker than a form but it is
honest, has zero dependencies, and a reply is a real lead. **Replace it with a real
form the moment the box exists** — see §6a.i's warning that a `mailto:` contact path
is a known free-repair risk under §7a.

## 3. THE BLOCKER — and it is not subtle

```
code 9109  Cannot use the access token from location: 5.65.164.9
```

**The token at `~/.config/cloudflare/token` has an IP allow-list, and this machine's
public address is not on it.** It worked for the whole of last night's DNS work and
began refusing partway through today. The error is **location-based, not
permission-based**, but a second call in the same breath returned the *generic*
`code 10000 Authentication error` — so **the same root cause reports itself two
different ways**, and 10000 reads like a scope problem you will waste time chasing.
[LANDMINE — recorded in `LANDMINES.md`.]

**A green health check does NOT mean the token will work.** Measured the same
minute: `/user/tokens/verify` returned `active` over **both** families while a real
zone call from that same IPv4 was refused. The verify endpoint is **exempt from the
filter**, so it can never detect this. Test with a real (read) call against the zone
you intend to write to.

### How to stop it happening — ranked

Token id `f0089a62ce6ea218b8c8137956d28297`.
**Cloudflare → My Profile → API Tokens → the token → Edit → Client IP Address
Filtering.**

1. **Allow the ISP prefix, not the single address — recommended.** The address that
   worked last night had rotated by this afternoon, so listing `5.65.164.9` alone
   buys days, not weeks. Enter a CIDR (e.g. `5.65.0.0/16`) so ordinary DHCP
   rotation stops breaking it. Keeps the filter meaningful — it still excludes the
   whole internet outside one ISP.
2. **Also pin the address family in the client.** Both families are filtered, so a
   dual-stack machine flips between two different addresses and fails
   *intermittently* — the worst kind. Use `curl -4` for every API call (the RUNBOOK
   snippets now do) and only the IPv4 side needs listing.
3. **Long-term: call the API from a stable address.** Once P1's box exists (§5), run
   Cloudflare changes from there and lock the token to that one IP. That is the
   version where the filter is genuinely tight *and* never in the way.
4. **Removing the filter entirely is defensible but is the weakest option here** —
   this token reaches **36 zones** with DNS and Page Rule write. The filter is the
   main thing limiting blast radius if it leaks. If you do remove it, shorten the
   expiry and rotate on a schedule instead.

**Or sidestep it entirely:** §4 is two dashboard changes and needs no token at all.

## 4. To make webdesign.uk live (2 changes, ~1 minute)

Both in the `webdesign.uk` zone. **Order does not matter here** — the page is
already in the bucket, so neither step can expose anything.

1. **Rules → Page Rules → delete** the rule `*webdesign.uk/*` → `302
   https://webdesign.co.uk/` (id `b8e08b35028315a274b2f5c7fea9154d`). It is the only
   rule on the zone.
2. **DNS → edit the apex `A` record** (id `3f0570fb2f0f45b9979b61779745e8fa`):
   `192.0.2.1` → **`199.59.243.228`**, keep **Proxied (orange)**.
   *That IP is a placeholder and is never contacted* — the `portfolio-sites-router`
   Worker intercepts first. It only has to match what `ugg2.com` uses.
   Do the same for `www` if you want www to serve rather than fail.

**Verify:**
```bash
curl -4 -sS -o /dev/null -D- https://webdesign.uk/ | grep -iE '^HTTP|^content-type'
# want: HTTP/2 200 + text/html   (NOT 302, and NOT a JSON 404)
```
A JSON body with `"objectKey":"webdesign.uk/index.html"` means routing works and the
object is missing — different failure, different fix. It is present; I checked.

## 5. The Mythic Beasts box — spec, and why

**Do not copy the idea.uk box.** Copy **the tools-api island**, which is already a
Mythic Beasts VDS and already runs the profile §6c wants:
containers + Postgres, Caddy, **`cloudflared` tunnel with no inbound**, nightly
`pg_dump` + off-box rsync (`vm_estate/PLAN_2026-07-25...:184`).

| | spec | why this number |
|---|---|---|
| **CPU** | **2 cores** | The chat is I/O-bound — it spends its life waiting on the Anthropic API, not computing. Concurrency is capped by §5.1's spend ceiling long before CPU matters. |
| **RAM** | **4 GB** | Postgres + app + Caddy + `cloudflared` fit in 2 GB, but 4 GB is the difference between "runs" and "runs while you also build/deploy on it". This is the one to not skimp on. |
| **Disk** | **40–60 GB SSD** | Transcripts and orders are text and stay tiny. Container images are what actually eat the disk. |
| **OS** | **Ubuntu 24.04 LTS** | Matches the estate (idea.uk runs nginx on Ubuntu); keeps `setup.sh` lineage usable. |
| **IPv4** | **not required** | ⭐ With a `cloudflared` tunnel there is **no inbound at all**, so you do not need a public IPv4 — and Mythic Beasts charge for one. IPv6-only is genuinely fine here and is cheaper. |
| **Backups** | nightly `pg_dump` + off-box copy | Copy the island's, verbatim. |

**Two properties matter more than the numbers.** (1) **No inbound ports** — the
tunnel makes `CF-Connecting-IP` unforgeable rather than merely conventional, and
removes the origin-firewall step that idea.uk needed. (2) **This box faces strangers
and spends money on every visitor**, so §5.1's control table — per-IP limit, turn
cap, per-day spend ceiling, request log, transcript-as-data — **ships with P1 or P1
does not ship**. It is not P2 polish.

> ⚠ **The `bugs_open/139` landmine applies the moment the tunnel is up.** Behind
> Cloudflare the true client address is in **`CF-Connecting-IP` only**. Get this
> wrong and your per-IP limiter silently becomes **one global bucket that still
> looks like it is working**. The discriminating check is
> `count(DISTINCT ip) > 1` **from two different networks** — one test machine
> cannot tell a constant from a working key (139: 83/83 identical rows).
> Reuse the estate's `cloudflare_realip` module; do not hand-roll it.

## 6. What is left for P1, in order

1. **Make the page live** (§4) — unblocks everything, costs a minute.
2. **Provision the box** (§5). No `hcloud`/`cloudflared` credentials exist on this
   machine, so this is an owner action or needs credentials supplied.
3. **The LLM chat**, with §5.1's controls written *first*, not retro-fitted.
   Intake model is the Haiku 4.5 tier per §7b — the chat is intake, not the product.
4. **Replace the `mailto:` with a real form** posting to the box.
5. **Stripe in test mode**, orders stored on the box.
6. **Terms** — §7a requires them before the first *sale*. The current page takes no
   money, so it is not yet blocking, but it becomes blocking at step 5.

## 7. Still owned by the owner

- ~~The price~~ — **settled 2026-08-03 at £1,200.**
- **The correction fee** (§7d suggested £150/change or £600/day). The page says
  changes after acceptance are charged and our mistakes are free — it does **not**
  quote a number, so this is not urgent, but a caller will ask.
- ~~**VAT.**~~ **SETTLED by the owner, 2026-08-03: not VAT registered, so there is
  no VAT on the price.** The page now says so in three places — the price card
  (*"£1,200 is the total — there's no VAT to add"*), a FAQ entry (*"Is there VAT on
  top?"*), and the footer. Live and verified at 13,254 B.
  > **Why it is stated rather than left silent:** a business buyer assumes a quoted
  > price is ex-VAT unless told otherwise, so silence reads as "£1,440 really". It
  > also removes the single most common pre-purchase question.
  > **Do not write "not VAT registered" on the page** — the price statement is what
  > the buyer needs, and registration status is a fact about turnover that need not
  > be published. **If registration ever happens this page must change**, and
  > £1,200 should stay the total (absorb it) rather than becoming £1,440 for
  > anyone who saw the old page.

## 8. Two things I got wrong today, for the next session's benefit

- I re-ran a `curl` against `idea.uk` and reported it **down**. It was serving 200
  throughout; my stub resolver held the pre-migration address. `dig @1.1.1.1`
  bypasses the system resolver and *agreed with the wrong conclusion*. Full entry in
  `WRONG_CALLS.md`; the check is `getent ahosts <host>` then
  `curl --resolve host:443:<ip>`.
- I claimed both halves of the ugg2 wildcard were missing. Only DNS was — the Worker
  route already existed. An empty `dig` cannot distinguish the two.
