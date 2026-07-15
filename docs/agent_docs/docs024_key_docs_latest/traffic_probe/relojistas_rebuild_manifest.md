# relojistas.com — Rebuild Manifest (HANDOFF Thread A, concrete)

Status: **DRAFT for approval** (2026-07-15). Sister docs: `relojistas_notes(8).md`
(provenance/decisions), `relojistas_golive(3).md` (go-live), `HANDOFF_vm_sites_permanent_thread.md`
(Thread A brief), and the news-feed pipeline (concept register `news-feed-pipeline.md`,
docs 006). This manifest is the Thread-A deliverable: package what we know and hand it
to the chassis to build a multi-page Spanish static site + an outbound RSS feed that
reactivates the domain's still-live subscribers.

---

## 0. Last traffic read (2026-07-15, live off the box, 9–14 Jul window)

- **405,701 requests / 5.5 days**, but ~99.7% bot/crawler noise: **83.5% are 404s**
  on dead vBulletin paths, **15.8%** are www/https 301s, **only 1,367 (0.34%) are 200**
  — of which just **89 were the real homepage** and 8 were beacon hits.
- **ASSET 1 — the RSS feed still has real subscribers, all getting 404.**
  `/external.php?type=RSS2` was pulled **749×** (~136/day, steady) from 31 distinct
  CF-masked sources. Pullers include **FeedFetcher-Google** (Google only polls feeds a
  real user subscribed to), **meta-webindexer** (Facebook), **Applebot**, DotBot, and
  many feed-reader UAs. Subscribed variants seen: `forumids=2,4,13,44,58,78,145,288`
  and per-user `cat=…&ppuser=…`. **Every request 404s today.**
- **ASSET 2 — small but real organic intent (the June "human ≈ 0" verdict predates it).**
  Search box captured **9 events, ~8 organic**, over 3.5 weeks from **ES / CL / MX**:
  `Omega seamaster` (CL), `Certina 919` (ES), `Casio shock ga2100` (MX), `Aguja 0.18`
  (ES), `president 22` (ES), `Rolen date` (ES), `Omega 285` (ES), `Casio` (MX).
- **Verdict shift:** reactivating the RSS feed with live Spanish watch news is the
  highest-leverage move — it converts ~136 daily 404s into 136 daily touches on real
  subscribers, and (per the brief) the feed can link out to sources now, pulling traffic
  back later.

## 1. Confirmed decisions (this session)

1. **Host = the existing box.** Keep the Hetzner CPX22 (167.233.33.159, nbg1,
   CF-proxied) as the static host. nginx serves the built static site; a rewrite/handler
   answers the legacy query-string feed URL. (Pure object storage was rejected — it
   cannot answer `/external.php?type=RSS2`.)
2. **Keep the engine + a "clever" backend feature** alongside the static pages
   (see §6). Not pure-static; the engine is already live and now yielding real intent.
3. **Manifest first, then build** (this document).

## 2. Site identity / classification

- **Domain:** relojistas.com · **Language:** Spanish (`es`) · **Audience:** Spanish-speaking
  watch enthusiasts (ES/CL/MX evidenced) · **Vertical:** relojería — brands, repair,
  collecting, marketplace-heritage, news.
- **Framing (CONFIRMED 2026-07-15):** a **Spanish-language watch NEWS portal built on the
  forum's heritage** — it acknowledges the old *Foro de relojes*, honours the old feed
  subscriptions at the same URL, and gives returning visitors current watch news + a search
  that now answers.
- **`sites` row:** `github_repo='vm-sites'`, `deploy_config.target='vm'`,
  `deploy_config.engine.{base_url,stats_key}` set (onboarding UPDATE already designed in
  P4), capabilities include `backend` (engine retained).
- **Classification:** `content_features.news_feed.recommended=true` (this is the flag the
  6-hourly `content-feed-trigger` keys on — it's what makes the pipeline adopt the site).

## 3. Information architecture (pages)

- **`/` home:** hero (Spanish) → **latest-news card grid** (top ~6, client-fetched from
  `/data/latest-news.json`) → **search box** (Spanish; now returns results, §6) →
  brand/category tiles (marcas / reparación / compraventa / ferias — mirrors the old
  boards).
- **`/noticias` news listing** (news-index page → also produces `/data/news-archive.json`,
  ~20 items).
- **Evergreen static content** so the site is never empty between news cycles: short
  original Spanish primers (brand guides, repair basics, a glossary). All original text.
- **`/gracias.html`** thanks page — retained (engine `THANKS_PATH=/gracias.html`, its own
  box so no shared-filename constraint).
- Optional v2: per-category news pages (marcas / reparación / …) that double as the
  targets for the mapped legacy feeds (§5).

## 4. News ingest (existing pipeline — data rows, no new code)

Chain is live and proven elsewhere (gaswholesalers, robot-hands, vonc):
`content-feed-trigger` (6h) → `content-feed-orchestrator` → `feed-ingester` →
`feed-triage` (relevance + credibility + source diversity) → `render_news_section`
→ commits `/data/latest-news.json` (+ archive). Seed `content_sources` rows:

- **Primary — `api_news`** (xAI Grok-4-1-fast Responses API, web_search + x_search) with a
  Spanish watch-news prompt (últimas 24–72h; marcas, novedades, ferias, subastas,
  reparación). **No URL fabrication risk** — real-time search, and triage already rejects
  fabricated URLs.
- **Second `api_news` provider — Gemini**, added later (operator request) alongside Grok.
- **Supplement — `rss`** — 5 feeds **VERIFIED live + on-vertical Spanish (2026-07-15):**
  Debajo del Reloj, Tiempo de Relojes, TR Magazine, Máquinas del Tiempo, Relojes y Estilo
  (URLs in the runbook P2 / running notes). Monochrome Watches (EN) optional for a v2
  "internacional" category. Triage's source-diversity interleaving caps any one source at
  ~2 of 6 slots — 5 independent magazines + Grok is an ideal spread.
- **Optional — `news_search`** (web-search adapter) as a third angle.
- **Cadence:** existing 6h heartbeat. **Rights posture:** items carry the **source URL**;
  we store title + short summary + link-out only — never full-text republication. This is
  exactly the "link out to sources rather than us" brief, and it's the rights-safe design.

## 5. Outbound RSS (the one net-new piece) + honouring the legacy URL

The pipeline *ingests* news but has **no outbound RSS publisher** today (`feed-publisher`
is unbuilt; `render_news_section` emits JSON, not RSS XML). Two parts:

- **`render_rss_feed`** — new render step, close cousin of `render_news_section`: reads the
  same curated `content_feed_items` (top ~30), emits **RSS 2.0 XML**, commits `/feed.xml`.
  Channel metadata in Spanish; channel `<link>`/`atom:self` = the **legacy URL** so
  subscribers keep the same address; each `<item><link>` = the original source.
- **Legacy-URL honouring** — the engine answers `/external.php` (a query-string path pure
  static can't serve; see §6). `type=RSS2` with no board id → master feed. `forumids=N` /
  `cat=N` → the nearest mapped category feed. Aggressively cached at nginx/CF.
  - **Board→topic map (best-effort, from the subscribed ids we observed):** v1 may serve
    the master feed to **all** `type=RSS2` requests (captures 100% of the hammering with
    zero mapping risk); v2 maps `forumids=44/288/78/145/13/58/4/2` to category feeds once
    per-category news pages exist. The old board titles are recoverable from the Wayback
    snapshot if we want an exact map.
- **nginx:** add `location = /external.php` (proxy to engine) and a static `/feed.xml`
  location; idempotent `setup.sh` re-run installs them.

## 6. The engine's "clever" delivery (pick one; A+B recommended)

Retaining the engine buys capabilities a static host can't. Candidates, all reusing the
curated feed data (no new data source):

- **(A) Dynamic legacy feeds — RECOMMENDED.** The engine serves the query-string feed URLs
  and maps old `forumids`/`cat` to our topic-filtered feeds, so each returning subscriber
  gets the closest modern feed instead of a generic one. This is *the* thing pure static
  cannot do, and it directly maximises reactivation of the 31 live subscriptions.
- **(B) Search that answers — RECOMMENDED.** Upgrade the v1 "no-results" box: on submit the
  engine (i) records the intent event as today AND (ii) returns real results — matching the
  query against our news items + a small watch reference set, linking out. Converts captured
  intent into delivered value and feeds the "direct traffic back" goal, while still
  measuring demand. Low effort (engine reads the same `/data/*.json`).
- **(C) "Búsquedas recientes" wall.** A moderated, anonymised live strip of recent searches
  as social proof + evergreen content. Needs a moderation gate (raw searches can be junk/PII)
  — hold as v2.

**CHOSEN: A + B** (operator-confirmed 2026-07-15). A honours the RSS-out ask at subscriber
granularity; B turns the retained intent probe from a silent sensor into a visible feature.
Both reuse existing data. (C — recent-searches wall — stays v2.)

## 7. Hosting / deploy mechanics

- **Box:** Hetzner CPX22, CF-proxied. Webroot `/var/www/vm-sites/relojistas.com`.
- **Content deploy:** the `vm-sites` GitHub Action (git_deployer → `resolveGitRepoName`
  resolves `vm-sites` from the site's `github_repo`; this plumbing has **landed** since the
  June handoff — verified in `helpers.go`/`git_deployer_actions.go`).
- **Engine deploy:** the `site-engine` Action (build amd64 → ship → sudo-hook swap).
- **Prereq:** finish CF real-ip (`cloudflare-realip.conf` / `CLOUDFLARE=true` setup.sh
  re-run — still pending per running notes) so logs + `country` are accurate under proxy.
- Retention/prune timers stay; privacy posture unchanged (no IP/UA/cookie logging in engine).

## 8. Build sequence (once approved)

1. `sites` row + classification (`news_feed.recommended=true`, `target=vm`, engine config).
2. Seed `content_sources` (Grok `api_news` primary + verified `rss` supplements).
3. Run the news pipeline once; confirm items ingested + triaged (don't build pages empty).
4. Chassis build of the Spanish static site (planner → content → design → assemble → deploy
   to `vm-sites`): home + `/noticias` + evergreen pages + latest-news + search.
5. Build `render_rss_feed`; commit `/feed.xml`; verify valid RSS 2.0.
6. Engine: add `/external.php` handler (master + mapped feeds, option A) + search-with-results
   (option B); deploy via site-engine Action.
7. nginx: add `/external.php` + `/feed.xml` locations; idempotent `setup.sh` re-run.
8. **Verify end to end:** `curl 'https://relojistas.com/external.php?type=RSS2'` → 200 valid
   RSS; a real feed reader subscribes; search returns results; intent still captured; watch
   a subscriber's 404s turn into 200s in the access log.

## 9. Open items / risks

- ✅ **Framing** confirmed (news portal); ✅ **clever-engine** A+B chosen; ✅ **RSS sources**
  verified (5 live Spanish feeds; Grok primary + Gemini later).
- **Rights:** link-out + short summary only; never full-text republication.
- **forumid→category map** is best-effort; v1 master-feed-to-all is the safe default.
- **CF real-ip** re-run (P0) is a prerequisite for accurate country/logs — a production-box
  change the operator runs before the build proper.
- Re-check feed liveness at seed time; keep leaning on `api_news` (fabrication-safe).
