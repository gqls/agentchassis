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

**Fix, either way:**
- **Owner:** Cloudflare → My Profile → API Tokens → this token → edit **Client IP
  Address Filtering**: add `5.65.164.9`, or remove the restriction. Note the machine
  is dual-stack — an IPv6 egress (`2a02:c7e:3066:5400::/64`) hit the same wall, so
  allow-list both families or neither.
- **Or just do §4 in the dashboard by hand** — it is two changes.

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
- **VAT.** The page says only "Prices shown in pounds sterling" and deliberately
  avoids "inc/ex VAT", because the position is undecided. **Settle it before step 5**
  — a consumer-facing price is normally VAT-inclusive, and changing £1,200 from
  inclusive to +VAT after people have seen it is a bad look.

## 8. Two things I got wrong today, for the next session's benefit

- I re-ran a `curl` against `idea.uk` and reported it **down**. It was serving 200
  throughout; my stub resolver held the pre-migration address. `dig @1.1.1.1`
  bypasses the system resolver and *agreed with the wrong conclusion*. Full entry in
  `WRONG_CALLS.md`; the check is `getent ahosts <host>` then
  `curl --resolve host:443:<ip>`.
- I claimed both halves of the ugg2 wildcard were missing. Only DNS was — the Worker
  route already existed. An empty `dig` cannot distinguish the two.
