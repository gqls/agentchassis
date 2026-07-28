# EVIDENCE — what is crawling relojistas.com, what it wants, and what it is costing us

**Measured 2026-07-28 from the origin's own nginx log**, window `27/Jul 00:32 → 28/Jul 09:20`
(~33 hours, 243,551 requests). **This was impossible before 2026-07-27**: until Cloudflare
real-ip landed, every client IP was a CF edge address, so per-source analysis could not be done
at all. This is the first look.

Owner's question: *"what are they doing, what content are they looking for, what will they do
with that information — can we feed them good content to our benefit?"*

## The headline

| | | |
|---|---|---|
| total requests | **243,551** | |
| **404** | **225,244** | **92.5%** |
| 301 | 17,181 | mostly http→https |
| **200** | **946** | **0.4%** |

**Nine of every ten requests to this domain are a 404 for a forum that no longer exists.**

## Who they are — two completely different populations

### 1. A scraper fleet sweeping the dead forum's images (~85% of ALL traffic)

```
/attachment.php   208,529 requests   (184,238 on 27 Jul alone)
                  1,409 distinct source IPs
                  25,030 distinct attachmentid values, range 9 … 40,756
```

Three signatures pin what this is:

- **98% request the literal `&amp;`** — `attachment.php?s=<32-hex>&amp;attachmentid=12599&amp;d=1338327221`.
  204,470 of 208,529. A browser would have decoded that entity. **These URLs were lifted from
  raw HTML by something that never parsed it.**
- **The user-agents are spoofed desktop browsers** — Windows Chrome (129,881), Mac Safari
  (61,880), Linux Chrome (13,368). Only 1,971 admit to being a crawler.
- **The `s=` session tokens are dead vBulletin session ids**, preserved from whatever page the
  URLs were harvested off.

**What it is:** a distributed image harvest of the old forum's attachment ID space, run from a
list of ~25k URLs scraped out of archived or third-party HTML. Most likely an image/dataset
collector or a mirror.

**What it will do with the information: nothing.** Every one of those requests is a 404. It has
been getting 404s for months and has not stopped, because **a 404 means "not now" — only a 410
means "never".**

**Can we feed it good content to our benefit? No.** It is not an audience: it fetches images,
renders nothing, refers no one, and indexes nothing you can be found in. Serving it content
would be giving away bandwidth for no return.

### 2. Real search and AI crawlers (a rounding error by volume, and the ones that matter)

```
Applebot 3,274 · Bingbot 298 · Googlebot 99 · YandexBot 75 · ClaudeBot 20 · DuckDuckBot 8
plus facebookexternalhit 23 · Twitterbot 8
```

**And here is the finding that matters.** Their status split over the same 33 hours:

```
404: 2,942     301: 808     200: 38     304: 7
```

**78% of real crawler budget is spent on the dead forum.** They fetched `/faq.php` 480 times
and got 404 every time. The 38 successes were:

```
8 /              6 /data/latest-news.json    4 /assets/js/snippets.js
2 /historia.html 1 /sobre-nosotros.html      1 /noticias/index.html   1 /noticias/
```

**Seven requests in 33 hours touched the feed or any live article** across `/noticias/`,
`/guias/` and `/glosario/`.

So: a site publishing curated Spanish watch news every day, with 19 live pages and a 30-item
feed, is **effectively invisible to search** — not because the content is weak, but because the
crawlers are spending their visits on a corpse.

## Two things nobody had noticed

**1. `robots.txt` is Cloudflare's, not ours — and it blocks every AI crawler.**

```
User-agent: ClaudeBot | GPTBot | CCBot | Google-Extended | Amazonbot
            Applebot-Extended | Bytespider | meta-externalagent   →  Disallow: /
Content-Signal: search=yes, ai-train=no, use=reference
```

That is Cloudflare **Managed** content, served at the edge. It is a platform default, not a
decision anyone here made. `search=yes, ai-train=no` is a defensible stance — but the blanket
`Disallow: /` also blocks **ai-input** (being read to answer a question and cited), which is
where referral value would come from.

**2. There is no `sitemap.xml`** (404). Something is already looking for one — 16 requests
carry `http://www.relojistas.com/sitemap.xml` as a *referer*.

## Also observed: genuine third-party hotlinks

`foroderelojes.es` — a **live** Spanish watch forum — appears as a referer on 149 requests,
all for `/attachment.php`. Old threads over there embed images that used to live here. Those
are real pages real enthusiasts still read, and every one shows a broken image.

Small in volume, and it cannot be converted by serving content (a broken `<img>` does not carry
a click). Noted because it is evidence the domain's images were woven across the Spanish watch
web — which speaks to the domain's standing, not to a traffic tactic.

## What follows — the answer to "feed them to our benefit"

The instinct is right; the target is wrong. **You cannot feed the 85%** — it is a scraper fleet
with no return path. **What you can do is stop it eating the budget of the ones you want.**

1. **`410 Gone` for the dead vBulletin surface** (`attachment.php`, `showthread.php`,
   `newreply.php`, `faq.php`, `printthread.php`, `private.php`, `sendmessage.php`, `search.php`).
   410 is the only status that tells a crawler *never come back*. This is the same
   generator-owned nginx move already proven on the legacy feed URLs — a `setup.sh` edit and a
   re-run. Kills ~184k/day of pointless work and removes the "92% 404" deadness signal.
2. **Publish `sitemap.xml`** — 19 live pages plus the feed. The counterpart to (1): stop
   advertising the corpse, start advertising the site.
3. **Decide the AI-crawler policy deliberately** rather than inheriting Cloudflare's default.
   If the strategy is to build the domain's value, being *readable and citable* by AI answers
   (`ai-input`) is exposure; being *training fodder* (`ai-train`) is not. Those are separable and
   are currently both refused. **Owner call — it is a business decision, not a technical one.**

**Expected effect, stated honestly and in advance so it can be checked:** (1) and (2) should
move real-crawler 200s from **38 per 33 hours** toward the live page count, and the 404 share
well below 92%. That is the measurement to re-run in a week. It is a hypothesis, not a promise —
**[UNMEASURED]** until it is.

## Caveat on any traffic figure quoted for this domain

**Any headline traffic number for relojistas.com is 92.5% 404s.** If a figure ever reaches a
listing, a pitch or a valuation, it must separate: **~184k/day scraper 404s**, versus **946
successful responses in 33 hours**, versus the human-scale number, which is smaller again. This
workstream has already logged one instance of counting requests as people
(`WRONG_CALLS.md`, 2026-07-25) — this is the same trap at ten times the scale.


---

# REVISITED, same day — the owner pushed back, and he was right

He asked again: *"What will the crawlers do with my images, text etc. They may use them and that
can be to my advantage."* Asked me to look once more and be willing to reach a different answer.
I have.

## Where the first answer was wrong

**I answered a different question from the one asked.** I measured what the crawlers do with
**404s** — nothing, obviously — and presented that as what they would do with **content**. Those
are not the same question, and the evidence I gathered cannot settle the second one. The whole
site is returning errors to them; of course nothing comes back.

**And I called the Cloudflare default "a defensible stance" without measuring what it costs.**
It is measurable, and it is not free.

## The evidence I did not have when I answered

Tested directly — what relojistas actually tells each crawler:

```
ClaudeBot → DISALLOWED     Googlebot → allowed
GPTBot    → DISALLOWED     Applebot  → allowed
CCBot     → DISALLOWED
```

And what they did about it, from the origin log:

```
ClaudeBot   24 requests — ALL of them /robots.txt. Zero pages.
Bytespider   8 requests — ALL of them /robots.txt. Zero pages.
Applebot 3,964 requests — allowed, and spent 599 of them on /faq.php 404s.
```

**ClaudeBot came, asked permission, was refused, and left.** That is not hypothetical exclusion
I inferred from a config file — it is an observed round trip.

## The finding: the managed robots.txt contradicts itself

```
User-agent: *
Content-Signal: search=yes, ai-train=no, use=reference     ← "you MAY read and cite me"
Allow: /

User-agent: ClaudeBot | GPTBot | CCBot | Google-Extended | …
Disallow: /                                                ← "you may not fetch me at all"
```

In robots.txt the **most specific matching group wins**, so a named agent obeys its own
`Disallow: /` and never sees the `*` group. **The permission to be referenced is a dead letter** —
granted in one breath to agents that are denied access in the next. Nobody chose that; it is
what Cloudflare's managed file plus its block-AI toggle produce together.

## The argument that actually decides it

**A robots.txt block is voluntary compliance. It only binds the crawlers that read it.**

- The **184,000/day scraper fleet ignores it entirely** — spoofed browser UAs, no self-identification, and it has been hammering `/attachment.php` regardless.
- **ClaudeBot read it and left.**

So the block is doing the precise opposite of what a publisher wants: **it stops the crawlers
that would attribute, cite and send traffic, and does nothing whatever to the ones that would
take content without credit.** It filters for good behaviour and penalises it.

## What they would actually do with the content — split by kind

| | what happens | worth it? |
|---|---|---|
| **Text → answer engines** (ClaudeBot, GPTBot, PerplexityBot) | fetched at answer time, quoted, **cited with a link** | **Yes, and disproportionately here.** Spanish-language horology is far thinner than English. Being *the* Spanish source is a much larger share of a much smaller pool. |
| **Text → Common Crawl** (CCBot) | enters the public corpus that feeds downstream datasets, research and "what is this domain" tooling | **Yes, and it compounds.** Free, permanent, and the substrate everything else is built on. Currently blocked. |
| **Text → training** (`ai-train`) | absorbed into weights. No link, no attribution, no referral | **Legitimately refusable.** Nothing comes back. Refusing this is a real choice, not squeamishness. |
| **Images → image search** (Googlebot, Applebot) | indexed, surfaced, clickable | Yes — and **already allowed**. It just cannot reach them past the 404 wall. |
| **Images → AI datasets** | absorbed, no attribution | Low value, same class as `ai-train`. |
| **Images → the 184k sweep** | fetched, 404, retried forever | **No. Unchanged from my first answer** — and this is the one part of it that survives. |

The key distinction, which Cloudflare's own Content-Signal already models and its blanket
`Disallow` then discards: **`ai-input`/reference (read → cite → send a reader) is advertising.
`ai-train` (absorb → no attribution) is donation.** They are separable. Right now both are
refused.

## The concession I should have offered the first time

I considered a **branded placeholder image** at `/attachment.php` and dropped it without saying
so. It deserves airing, because there is a version that works:

- 98% of those requests are the sweep, which renders nothing → a placeholder is wasted on them.
- But **149 requests carry a `foroderelojes.es` referer** — a *live* Spanish watch forum whose
  old threads still embed images that used to live here. Real enthusiasts are reading those pages
  and seeing a broken image.
- **Serve the placeholder only when a genuine third-party referer is present, and 410 to
  everything else** (`map $http_referer` in nginx). The sweep gets told to stop; the humans get a
  small "Relojistas — relojistas.com" mark on a forum full of the exact audience this site wants.

Modest — the human slice is ~150 requests in 33h, **[UNMEASURED]** how many are real page views
versus the sweep spoofing a referer. Worth doing as a rider on the 410 work, not as a project.

## Revised recommendation

1. **410 + sitemap** — unchanged, still first. Nothing else matters while 78% of real crawler
   budget hits a corpse. *(Sitemap shipped 2026-07-28.)*
2. **Unblock the reference/answer crawlers, and CCBot.** Keep refusing `ai-train` if he wants —
   Content-Signal already expresses exactly that, and it is the honest line. This is a
   **Cloudflare dashboard setting, not a file we control** — the managed robots.txt is served at
   the edge, so it is an owner action.
3. **Referer-gated placeholder** as a rider on (1).

**What I got right and am keeping:** the 184k/day sweep is not an audience, and any traffic
figure for this domain is 92.5% 404s. **What I got wrong: I let "these particular scrapers are
worthless" stand in for "crawlers have nothing to give us", and those are very different
claims.**
