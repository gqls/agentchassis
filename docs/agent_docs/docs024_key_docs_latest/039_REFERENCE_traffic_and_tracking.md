# 039 — REFERENCE: traffic and tracking across the estate

**Written 2026-08-25.** Sources: the `dartsonline_traffic` lane did the Cloudflare method and the
self-traffic discovery; the `apis_uk_bees_homepage` lane rolled out Tag Manager and verified the
Cloudflare numbers independently on a second site. **Every figure below is marked with how it was
obtained** — `[MEASURED <date>]` means run here and the output seen, `[RELAYED]` means another
lane measured it, `[UNTESTED]` means nobody has run it.

---

## 1. Read this first: our own tooling is a large slice of "traffic"

**The single most important fact in this document.** Sessions hit these sites constantly with
`curl` and headless Chrome — verification probes, cache-busted fetches, render checks. Cloudflare
counts every one as a page view.

| site | our share of page views | window | source |
|---|---|---|---|
| dartsonline.com | **28.5%** | 30d | `[RELAYED]` dartsonline_traffic |
| apis.uk | **27.1%** | 7d | `[MEASURED 2026-08-25]` |
| fleet | **10.8%** (8,300 of 77,132) | 7d | `[RELAYED]` |

**It concentrates on whichever site a lane is actively working**, which is exactly the site someone
is about to ask about. The fleet average understates it by roughly 3× for the site that matters
this week. **The cleanest demonstration is one query across two sites** `[MEASURED 2026-08-25]`:

| | apis.uk (worked all week) | noted.co.uk (not worked) |
|---|---|---|
| our tooling share | **27.1%** | **2.4%** |
| page views 7d | 395 | 17,037 |

Same method, same window, an order of magnitude apart — and the difference is not the sites, it is
**which one had a session pointed at it**.

⚠ **And it moves, which is the dangerous part.** On dartsonline, splitting a window in half:
human traffic **+19%** while ours went **4.8×**. Total page views roughly doubled and **none of it
was growth** `[RELAYED]`. That lane nearly reported the rise to the owner as the work paying off.

**So: never quote a raw Cloudflare page-view figure. Classify by `uaBrowserFamily` and exclude
`Curl`, `ChromeHeadless`, `Wget`, `PythonRequests`, `Go-http-client` first.** A traffic number that
has not had this done to it is not a traffic number.

---

## 2. Cloudflare — the only source with history, and it needs no per-site setup

**Status: live and working on all 40 zones** `[RELAYED 40/40]`, `[MEASURED 2026-08-25: 5/5 on a
sample]`. Scope `Zone → Analytics → Read` was added to the token on 2026-08-23.

Tokens that carry it: `~/.config/cloudflare/portfoliotoken` **and**
`~/.config/cloudflare/token.expired-2026-08` (⚠ **the filename is a lie — that one works**). Two
other token files are denied.

### The working query

```bash
T=$(tr -d '\n' < ~/.config/cloudflare/portfoliotoken)
curl -s -X POST https://api.cloudflare.com/client/v4/graphql \
  -H "Authorization: Bearer $T" -H "Content-Type: application/json" --data '{
  "query":"query($zones:[String!],$since:Time!,$until:Time!){viewer{zones(filter:{zoneTag_in:$zones}){zoneTag httpRequests1dGroups(limit:100,filter:{date_geq:$since,date_lt:$until}){sum{pageViews requests browserMap{uaBrowserFamily pageViews}}}}}}",
  "variables":{"zones":["<zone-id>"],"since":"2026-08-18","until":"2026-08-25"}}'
```

### Limits, all paid for by someone

- **`zoneTag_in` is capped at 8 zones per query.** 40 returns `too many zones requested`. Batch it.
- **Free plan cannot use `httpRequestsAdaptiveGroups` beyond a 1-day window** — so **no per-path
  breakdown and no bot-score splits** over a useful period. `browserMap` is the best proxy.
- **`browserMap` is a User-Agent string classification, not a bot score.** It is trivially
  spoofable and it leaves an `Unknown` bucket — **20.3% on apis.uk** `[MEASURED 2026-08-25]`,
  **36% on dartsonline** `[RELAYED]`. **Do not fold `Unknown` into "human".**
- **`pageViews` is roughly 6% of `requests`** — requests count every asset. **Always quote page
  views.** On apis.uk: 395 page views against 38,830 requests over 7 days
  `[MEASURED 2026-08-25]`, and that zone also carries a live API, which inflates requests further.
- **A `10000 Authentication error` means EITHER a missing scope OR an IP restriction** — Cloudflare
  returns the same code for both. `LANDMINES.md` records someone re-issuing a token over that
  confusion for nothing. Distinguish by calling a *different* endpoint from the same machine.
- **Matching a token file to its dashboard entry:** the **actor id in the denial message** is the
  token id, and it appears in the dashboard URL when you open that token. That is the only reliable
  mapping found; names do not match.

---

## 3. Google Tag Manager / GA4 — live on the sites, and it started recording tonight

**Container `GTM-PQ3WCTBD` is in the `head` component of all 27 sites** that have one — 14 already
had it, 13 backfilled 2026-08-24 `[MEASURED]`. It reaches visitors only after a page re-render;
that roll-out covered 24 sites / ~680 pages.

⚠ **A container with no tags records NOTHING.** As of 2026-08-24 21:30, Version 2 of that container
had **0 tags, 0 triggers** — loading on every site and firing nothing, which is why GA4 Realtime
showed 0. **There is no backfill: GA4 history begins when a GA4 tag is published**, so for any
question about the past, Cloudflare is the only source.

### Setting it up (the two things people get wrong)

1. **Tag type must be `Google Tag`** (older UI: *GA4 Configuration*), **not `GA4 Event`**. A GA4
   Event tag needs an Event Name and does not send page views.
2. **The Measurement ID is on the data stream, not the property.** GA4 → **Admin → Data streams →**
   click the web stream → **Measurement ID**, format `G-XXXXXXXXXX`. The numbers in the property
   picker are property IDs and will not work.
3. **Save is not enough — Submit → Publish.** A saved tag in an unpublished container does nothing.
4. Trigger: **All Pages**.

### All sites share one container

So every site reports into whichever GA4 property that container's tag names. **Break reports down
by `Hostname`** or it is one merged number. Per-site properties are possible but need a
lookup-table variable keyed on hostname in GTM — materially more work, and splitting later is easy.

---

## 4. The two sources have OPPOSITE blind spots — a gap between them is not a bug

This is the part that will otherwise be "discovered" as a discrepancy every few weeks.

| | Cloudflare | GA4 |
|---|---|---|
| our `curl` probes | **counted** | **invisible** — curl runs no JavaScript |
| our headless Chrome | counted | **counted** — it does run JS |
| bots / crawlers | counted, and useful | invisible |
| ad-blocked humans | counted | **lost** |
| history before tonight | **yes** | none, no backfill |
| referrers, events, behaviour | none | **yes** |
| search queries / position | none | none — see §5 |
| needs cookie consent | no | **yes** (UK PECR) — see §4a |

### 4a. Turning GA4 on is a change of compliance position, not a continuation

**Measured 2026-08-25, before any GA4 tag was published:** `apis.uk`, `vonc.com`, `dartsonline.com`,
`oufe.com` and `noted.co.uk` all returned **zero `Set-Cookie` headers** on a first visit, and
`gtm.js` for `GTM-PQ3WCTBD` sets none either. With **0 tags in the container**, nothing was firing,
so **the estate was setting no cookies at all.**

⚠ *Method limit:* `curl` sees server-set cookies; GA4 sets `_ga` from JavaScript, which `curl` would
not catch. The zero-tag state is the stronger evidence, not the header count.

**So publishing a GA4 tag starts setting cookies on ~24 sites simultaneously.** That is a change of
position, not a continuation of the status quo, and it is worth deciding deliberately rather than
discovering. Context `[RELAYED, dartsonline_traffic 2026-08-16]`: 11 live sites carried the
container, 8 had no privacy policy at all; dartsonline published one on 08-20, driven partly by
affiliate-network requirements. **No site carries a consent banner today.**

**The asymmetry is the point:** the Cloudflare route is server-side and brings no consent
obligation, so the two options are not equivalent-with-different-numbers — one of them changes what
the estate owes its visitors.

**Consequence worth internalising: GA4 is not a cross-check on Cloudflare's human figure.** They
measure different populations. On apis.uk, GA4 would have excluded the *entire* 27.1% `curl` slice
automatically, because `ChromeHeadless` does not appear there at all — so on that site GA4 is
structurally cleaner. On a site where a lane drives headless Chrome, it would not be.

---

## 5. Search Console — the only source for the questions people actually ask about SEO

Neither Cloudflare nor GA4 gives **search queries, impressions, average position, or index
coverage**. Only Search Console does, and it is **per-property**, which is why it has not been done.

Cloudflare *can* tell you crawling is thin, and **crawler traffic IS readable from `browserMap`** —
no extra scope, same query as everything else `[MEASURED 2026-08-25]`:

| site | GoogleBot | BingBot | window |
|---|---|---|---|
| noted.co.uk (our busiest) | **25** | 41 | 7d |
| apis.uk (3 days old) | 0 | 0 | 7d |
| dartsonline.com | 54 across 24 URLs | – | 30d `[RELAYED]` |

So ~3.5 GoogleBot page views/day on the busiest site in the estate. **That is the number to put in
front of anyone asking "are these sites being found", and GA4 will never show it** — a crawler runs
no analytics JavaScript. Cloudflare can say crawling is thin; it cannot say *why*.

### The automation shape — `[UNTESTED]`, offered as a plan to verify, not as fact

The hard half is already held: **the same Cloudflare token has DNS write on all 40 zones**
(`scripts/cloudflare/add_www_redirect.sh` POSTs `dns_records` with it), so the DNS-TXT step needs
no human.

1. `siteVerification.webResource.getToken` (method `DNS_TXT`) → the TXT value
2. `POST /zones/{id}/dns_records` with our token → live in seconds; Cloudflare is authoritative
3. `siteVerification.webResource.insert` → verifies
4. Search Console `sites.add` → property exists
5. `searchanalytics.query` → the data

Steps 1, 3, 4, 5 need a **Google Cloud service account** with the Site Verification and Search
Console APIs enabled — **one owner action, once, not per site.** After that a new site could
verify and register itself as part of the build pipeline.

⚠ **If this is ever automated across zones, scope it to the record types it needs.** The `apis.uk`
zone carries a **live API on `tools.apis.uk`**; `LANDMINES.md` (2026-08-23) records that a wildcard
worker route there would kill it silently. TXT records are harmless, but a per-zone routine that
writes anything broader is not.

---

## 6. How to answer "how much traffic do we get?" without being wrong

1. **Cloudflare, page views, not requests**, batched 8 zones at a time.
2. **Subtract our tooling** by `uaBrowserFamily`. Report it as its own line rather than deleting it
   — it is the only measure of how much we are hammering a site.
3. **Report `Unknown` separately.** Do not call it human.
4. **State the window and never derive a duration from a stale row.** A neighbouring lane published
   "about two days, not progressing" by subtracting a stale `updated_at` from today — arithmetic on
   an unmeasured row, in the same voice as the measurement.
5. **If a rise looks like success, split the window in half and check ours separately before saying
   so.** That is the check that caught the 4.8× above.
6. **For search performance, say plainly that we cannot answer it yet** and point at §5.
7. **Before believing a share, measure a case where it should NOT hold.** The self-traffic finding
   was first supported by **28.5%** on dartsonline and **27.1%** on apis.uk — but both are sites a
   lane was actively working, so that pair is equally consistent with *"every site reads ~27%
   curl"* and **could not have come out otherwise**. The number that made it a finding was
   **2.4% on noted.co.uk, which nobody was working** — the measurement that could have refuted it
   and didn't. Two agreeing figures from the same condition are one figure measured twice.
   *(Both lanes recorded this against themselves; it was caught by the second lane running the
   query on a site chosen because it should look different.)*

## 7. Current state, 2026-08-25

- Cloudflare analytics: **working, all zones, no setup** — the source of record for history.
- GTM: **on 27 sites**; GA4 tag being configured by the owner 2026-08-24 evening; **history starts
  from publication**.
- Search Console: **nothing set up.** Needs one owner action (service account) before any
  automation.
- Cookie consent: **not implemented on any site**, deliberately parked.

---

## Addendum 2026-09-02 (analytics_gtm lane) — GA4 is LIVE; §3/§4a/§7 above are historical from this line

- **GA4 published 2026-09-02 ~20:11Z**: container `GTM-PQ3WCTBD` version 3 carries one **Google
  Tag → `G-Y26N29T4KH`** (verified at gtm.js, both directions, `analytics_gtm/scripts/check_gtm_state.sh`).
  **GA4 history begins at that instant. There is no backfill.** Any GA4 number quoted for a period
  starting earlier is comparing a source to its own absence.
- **The consent position changed in fact at that instant** (§4a's "change of compliance position"
  happened): `_ga` cookies now set across ~30 estate sites, no consent banner anywhere. Owner's
  standing decision; consent is the open compliance item.
- **Customer sites are insulated by design**: hosted customer builds default to the separate,
  empty container `GTM-TH5XGNQ4` (owner ruling 2026-08-26, created 2026-09-02) — zero cookies by
  construction; a tag ever appearing in it is a re-ruling trigger, not progress.
- Break GA4 reports down by **Hostname** or the estate is one merged number (§3 above stands).

- **Addendum 2026-09-02 (late): consent is SHIPPING.** Owner chose the banner + Consent Mode v2
  route; live in the three head templates at 20:55:43Z (STY-060), reaching each site as it
  re-renders. From each site's convergence: **no cookies before opt-in** (GA4 falls back to
  cookieless pings), banner with equal Accept/No-thanks, withdrawal wipes `_ga*`. GA4 numbers will
  therefore UNDERCOUNT relative to the 09-02→convergence window — that step-change is the consent
  gate arriving, not lost traffic. Cookie policy PAGES still owed per site.
