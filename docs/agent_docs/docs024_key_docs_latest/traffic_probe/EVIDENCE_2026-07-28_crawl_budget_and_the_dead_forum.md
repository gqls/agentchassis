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


---

# CORRECTION + the actual opportunity (2026-07-28, third pass)

Owner asked three things: what to set in Cloudflare, what "crawl budget" actually means here,
and whether seeded Q&A text ("what is the best watch forum? relojistas.com") would help.

## First, a correction: I overstated "crawl budget"

**I used a large-site concept on an 18-page site.** Crawl budget is a real constraint for sites
with tens of thousands of URLs or more; Google's own guidance is that it is not a limiting
factor for small sites. **Google will crawl 18 pages whether or not the dead forum is there.**

What I measured is true — 2,942 crawler 404s against 38 200s — but **the causation I implied is
not established.** The 404s are not obviously *crowding out* the live pages.

What the dead surface actually costs, stated properly:

| effect | real? |
|---|---|
| Server load | **Yes — but it is the scraper fleet (184k/day), not search engines (~3.7k/33h).** |
| Crawlers retry dead URLs indefinitely | **Yes.** 404 means "gone for now"; only 410 makes them drop it. |
| A domain where nearly every known URL 404s may read as abandoned | **Plausible, [UNMEASURED].** |
| Live pages can't get crawled because the budget is spent | **Overstated. Withdraw.** |

**So the ranking changes: the sitemap was the bigger fix, not the 410.** Discovery was the
binding constraint — 18 pages reachable only by walking the homepage, and no sitemap at all.
That is now fixed. **The 410 work is housekeeping** (load, hygiene, stopping the retry loop),
not the unlock I implied. Worth doing; not first.

## What to set in Cloudflare

The served `robots.txt` is Cloudflare's **Managed** file — we do not control it from the repo,
so this is a dashboard change. **I cannot see the dashboard**, so this is the *outcome* to aim
for plus a way to verify it, not a menu path.

**Target state:**

```
User-agent: *
Content-Signal: search=yes, ai-input=yes, ai-train=no
Allow: /
```

…and remove the blanket `Disallow: /` for the agents below.

| allow — these cite and send a reader | keep blocking — these absorb, nothing returns |
|---|---|
| `ClaudeBot`, `Claude-User` | `GPTBot` (OpenAI's *training* crawler) |
| `OAI-SearchBot`, `ChatGPT-User` (ChatGPT search + live fetch) | `Google-Extended` (Gemini training) |
| `PerplexityBot` | `Applebot-Extended` (Apple Intelligence training) |
| `CCBot` — **owner's call**, see below | `Bytespider`, `meta-externalagent`, `Amazonbot` |

**The distinction that makes this coherent:** the same companies run *separate* crawlers for
training and for answering. `GPTBot` trains; `OAI-SearchBot`/`ChatGPT-User` fetch to answer and
cite. Allowing the second while refusing the first is a supported, standard position — it is
exactly what `search=yes, ai-train=no` says in words.

**`CCBot` (Common Crawl) is the genuine judgement call.** It is the public corpus behind a great
deal of downstream tooling, research and "does this domain exist / what is it about" lookups —
being in it compounds. It is also a primary training input. Refusing training *and* allowing
CCBot is not fully consistent; the owner should pick which he cares about more.

**Verify after changing it** — this is the check, and it must be run per-agent because the file
is served conditionally:

```bash
for ua in ClaudeBot GPTBot OAI-SearchBot PerplexityBot CCBot Googlebot; do
  printf "%-16s " "$ua"
  curl -s -A "Mozilla/5.0 (compatible; $ua/1.0)" https://relojistas.com/robots.txt \
   | awk -v n="$ua" 'BEGIN{IGNORECASE=1} $0~"^User-agent:[ ]*"n"$"{f=1;next} \
       f&&/^Disallow:[ ]*\/$/{print "DISALLOWED";x=1;exit} f&&/^User-agent:/{f=0} \
       END{if(!x)print "allowed"}'
done
```

**Caveat:** all of this is voluntary compliance, and the crawler taxonomy changes every few
months. The 184k/day sweep ignores robots.txt entirely and will not stop because of any setting
here.

## "Feed the trainers our URL and some content"

**Training gives nothing back, by construction.** Content absorbed into weights carries no link,
no attribution and no referral, and one small site is a vanishing fraction of a training corpus.
The theory that a model would thereby "know" the domain is weak and unmeasurable.

**Retrieval and citation is where the value is** — the model fetches at answer time and cites
with a link a reader can click. That is a real channel, and it is the one currently blocked.

So: allow the citers. Whether the trainers are allowed barely matters either way.

## The seeded-Q&A idea — this one will not work, and it fights the site's own rules

*"What is the best watch forum? relojistas.com"* — three separate reasons:

1. **Models do not accept a site's self-assessment as fact.** Training aggregates billions of
   documents; a page asserting its own superiority carries no more weight than any other. Answer
   engines rank on relevance and corroboration across sources, not self-declaration.
2. **It is an unsupported superlative, which this site's own style rules ban by name.**
   `content_direction` lists *"Superlativos vacíos sin respaldo"* under things to avoid, and
   *"El reloj que todo coleccionista DEBE tener"* under would-never-say. The platform's
   claims-verification layer exists to stop exactly this and would flag it.
3. **It is not true.** relojistas is not a forum any more — the site's own about page says the
   forum is gone. The first thing an answer engine would find is a contradiction.

And the risk is asymmetric: it cannot really succeed, and if it reads as manipulation the
downside is a de-ranked domain.

## What actually works — and relojistas is unusually well placed for it

**You do not get cited by claiming to be the best. You get cited by being the clearest answer to
a question, in a language where few good answers exist.**

The asset already exists. `/glosario/tourbillon.html` opens:

> **Tourbillon: qué es y cómo funciona esta complicación**
> *El tourbillon es un mecanismo ideado para contrarrestar los efectos de la gravedad sobre el
> escapamento…*

That is a question-shaped heading with a one-sentence definition and a real example
(Breguet; Richard Mille RM 64-01). It is precisely what an answer engine wants for
*"¿qué es un tourbillon?"* — and **Spanish-language horology is a far thinner corpus than
English**, so the marginal value of being *the* clear Spanish source is much higher than the
equivalent English page would be.

Three concrete moves, in order:

1. **Allow the citing crawlers** (above). Nothing else matters while ClaudeBot is turned away at
   the door.
2. **Add structured data — the site emits ZERO JSON-LD today** (verified across the homepage,
   a glossary entry and the news index; only 5 Open Graph tags). This is the standard,
   legitimate, machine-readable way to be quotable:
   - glossary → `DefinedTerm` within a `DefinedTermSet`
   - guides → `Article`, news → `NewsArticle` with `datePublished`
   - `Organization` + `WebSite` on the homepage
   This is a platform-wide gap, not a relojistas one — **worth checking whether any site in the
   fleet emits JSON-LD before building it here.**
3. **Optionally `llms.txt`** (`/llms.txt`, currently 404) — an emerging convention offering a
   plain-markdown summary of a site for LLMs. Cheap, honest, **and unproven; adoption is not
   broad.** Mentioned for completeness, not recommended as a priority.

**Expected effect, in advance and falsifiable:** if the crawlers are unblocked and structured
data added, the check in ~2 weeks is whether `ClaudeBot`/`OAI-SearchBot`/`PerplexityBot` appear
in the access log fetching `/glosario/*` and `/guias/*` rather than only `robots.txt`. That is
observable in the log we already have. **[UNMEASURED]** until then.


---

# MEASURED 2026-07-28 (fourth pass) — Cloudflare MERGES, it does not yield

Owner asked: *"should I just disable robots.txt configuration and we handle it ourselves?"*
**Yes — and it is now proven that self-serving alone is not enough.**

We shipped our own `robots.txt` in `vm-sites` (`00bc72f`). What is now served:

```
lines 27-60   # BEGIN Cloudflare Managed content …  Disallow: / for ClaudeBot, GPTBot,
              CCBot, Amazonbot, Applebot-Extended, Bytespider, Google-Extended,
              meta-externalagent          ← STILL THERE, and FIRST
lines 64-82   our file, appended after    ← Content-Signal: ai-train=yes, Sitemap:
```

**Cloudflare PREPENDS its managed block to the origin's file.** The served result is 82 lines
with **two `User-agent: *` groups carrying contradictory Content-Signal values**:

```
line 30:  Content-Signal: search=yes,ai-train=no,use=reference     (Cloudflare's)
line 75:  Content-Signal: search=yes, ai-input=yes, ai-train=yes   (ours)
```

Per-agent check after shipping ours — **unchanged**:

```
ClaudeBot DISALLOWED · GPTBot DISALLOWED · CCBot DISALLOWED · Applebot-Extended DISALLOWED
OAI-SearchBot allowed · PerplexityBot allowed · Googlebot allowed
```

**What DID take effect:** the `Sitemap:` line, because `Sitemap` is group-independent — so
crawlers now get pointed at the sitemap regardless. That half was worth shipping on its own.

## So the dashboard change is required, and it is two settings not one

From the owner's screenshot (`Manage AI bot access`):

1. **"Block AI training bots"** → set to **"Do not block (allow crawlers)"**.
   This is a Cloudflare-managed WAF rule that *actually blocks requests at the edge* — it is
   enforcement, not advice, and it is separate from robots.txt.
2. **"Set your preference to block training in robots.txt"** → set to whatever means *no
   preference / do not manage*. This is what injects lines 27-60.

Only (2) stops the merge. Only (1) stops the enforcement. **Both, or the effect is partial** —
and the failure mode is silent, because a blocked crawler simply never appears in the log.

**Then re-run the per-agent check.** Not a single `curl`: the managed file is served
conditionally, so one fetch proves nothing about any particular agent.

*Also on that panel: "Markdown for Agents" (Pro plan) auto-converts HTML to Markdown for
requests sending `Accept: text/markdown`. That is a genuine convenience for agent consumers and
is the same intent as `llms.txt`. Not worth a plan upgrade on its own — noted so the option is
known.*

## Shipped alongside

- **`/llms.txt`** (live, 200): built **from** the live pages rather than written **about** them —
  every entry is the page's own `<h1>` and its own first sentence, 18 pages across glosario,
  guías and noticias. The convention is emerging and **unproven**; it is cheap and honest, which
  is the entire case for it.
- **`/robots.txt`** deliberately does **not** disallow the dead vBulletin paths. Blocking them
  would stop crawlers fetching them and therefore stop them ever seeing the `410` that removes
  them from an index. Crawl-then-410 first; disallow later if ever.

## Open Graph — the owner was right that it exists, and it is half-broken

He recalled OG work. It is real: `render_site_components_action.go:417-448` emits `og:type`,
`og:site_name`, `og:title`, `og:image`, `og:url`. But on relojistas:

```
og:image → https://relojistas.com/assets/images/og-card.png   →  404
og:title → "relojistas.com"        (the domain, not the page title)
og:description → ABSENT
```

**`og-card.png` is 404 on all five sites checked** (fundamentallyai, webdesign, oufe,
relojistas, idea.uk). So every social/WhatsApp/Slack share of every site in the fleet renders
with no preview image and, here, a bare domain as its title. **Fleet-wide, [UNMEASURED] beyond
those five.** Worth its own bug — it is not a relojistas defect.

**Structured data is dormant, not missing.** `process_html` IS registered
(`registry.go:1042`) and referenced by 2 agent definitions, and it calls
`datahelpers.AddStructuredData`. But that function only emits when
`businessInfo["business_name"]` is present, and **zero of seven live sites carry any
`application/ld+json`**. Registered, reachable, and silently producing nothing — the same
dormant-machinery class as `bugs_open/117`. And what it would emit is `Organization` only,
which is not what a glossary needs (`DefinedTerm`).
