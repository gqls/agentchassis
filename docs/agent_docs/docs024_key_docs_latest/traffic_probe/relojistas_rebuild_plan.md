# relojistas.com — Rebuild Plan (phased)

Living plan for the relojistas.com static-rebuild + RSS-reactivation workstream
(HANDOFF Thread A). Companion to `relojistas_rebuild_manifest.md` (the what),
`relojistas_rebuild_runbook.md` (the exact commands), and
`relojistas_rebuild_running_notes.md` (the dated log). Grounding docs:
`relojistas_notes(8).md`, `relojistas_golive(3).md`, `HANDOFF_vm_sites_permanent_thread.md`,
concept register `news-feed-pipeline.md`.

## Goal

Turn relojistas.com from a hand-made intent probe (index + gracias, ~all bot traffic)
into a **Spanish-language watch news portal on static hosting**, with an **outbound RSS
feed served at the legacy URL** that reactivates the domain's still-live feed subscribers,
while **retaining the engine** for a value-adding backend feature. News links out to
sources now (rights-safe; also seeds "traffic back" later).

## Why now (evidence, 2026-07-15 read)

- The legacy feed `/external.php?type=RSS2` is still pulled **~136×/day** (749 in 5.5 days,
  31 sources incl. FeedFetcher-Google/Applebot/meta-webindexer) — **all 404**. Live demand
  for a feed we're not serving.
- **~8 organic watch searches** (ES/CL/MX) over 3.5 weeks — real human intent that the
  day-one "human ≈ 0" verdict predates.
- Everything but the outbound-RSS emitter already exists in the chassis (news pipeline is
  live; VM deploy plumbing has landed).

## Confirmed decisions

| # | Decision | Rationale |
|---|----------|-----------|
| D1 | Host on the **existing CF-proxied box** (nginx static + legacy-URL handler) | Pure object storage can't answer `/external.php?type=RSS2`; box is already provisioned/proxied |
| D2 | **Keep the engine** + add a clever backend feature (§ manifest 6: A+B recommended) | Real intent now exists; engine enables dynamic per-category feeds + search-that-answers |
| D3 | **Manifest/plan first**, then build | Complete rebuild is multi-step; approve before spend |
| D4 | News **links out** to sources; title+summary only | Matches brief; rights-safe; feeds traffic-back goal |
| D5 | **Spanish** throughout (`es`); vertical = relojería | Audience evidenced ES/CL/MX; forum heritage |

## Phases

### P0 — Prereqs (infra hygiene)
- CF real-ip on the box (`CLOUDFLARE=true` setup.sh re-run / `cloudflare-realip.conf`) so
  logs + `country` are accurate under the proxy. **Exit:** access log shows real client IPs
  + `CF-IPCountry` reaching the engine.

### P1 — Site record + classification
- Create/patch the `sites` row: `github_repo='vm-sites'`, `deploy_config.target='vm'`,
  `deploy_config.engine.{base_url,stats_key}`, capability `backend`.
- Classification `content_features.news_feed.recommended=true`.
- **Exit:** the 6-hourly `content-feed-trigger` picks up the site.

### P2 — News ingest live (data rows only)
- Seed `content_sources`: `api_news` (Grok web-search, Spanish watch prompt) primary +
  verified `rss` supplements + optional `news_search`.
- Run pipeline once; confirm `content_feed_items` ingested + triaged (relevance/credibility/
  diversity). **Exit:** `/data/latest-news.json` renders with real Spanish items; 0 sources
  erroring.

### P3 — Static site build (chassis pipeline → vm-sites)
- planner → content → design → assemble → deploy: home + `/noticias` + evergreen Spanish
  pages + latest-news card + search box.
- **Exit:** site live at relojistas.com, news card populated, pages in Spanish.

### P4 — Outbound RSS (net-new)
- Build `render_rss_feed` (cousin of `render_news_section`): curated items → RSS 2.0 XML →
  commit `/feed.xml`; channel self-link = legacy URL; item links = sources.
- nginx `/feed.xml` location. **Exit:** `curl .../feed.xml` → valid RSS 2.0.

### P5 — Legacy-URL honouring + clever engine (D2)
- Engine serves `/external.php`: `type=RSS2` → master feed; `forumids/cat` → mapped feed
  (v1: master-to-all; v2: per-category). Option B: search returns results while still
  capturing intent.
- nginx `/external.php` → engine; deploy engine via Action.
- **Exit:** `curl '.../external.php?type=RSS2'` → 200 valid RSS; search returns results;
  intent event still recorded.

### P6 — Verify reactivation
- Watch the access log: subscriber pulls on `/external.php?type=RSS2` flip **404 → 200**.
- Confirm a real feed reader renders the feed; confirm outbound links resolve.
- **Exit:** measurable reactivation (subscriber 200s), news refreshing on the 6h cycle.

## Open items / risks
- ✅ Framing confirmed (news portal); ✅ clever-engine A+B chosen; ✅ 5 Spanish RSS feeds
  verified live (Grok api_news primary, Gemini later) — see runbook P2 / running notes.
- Rights: link-out + summary only.
- forumid→category map is best-effort (v1 master-to-all).
- render_rss_feed is the only new Go action — keep it a thin variant of render_news_section.
- P0 (CF real-ip setup.sh re-run) is a production-box change the operator runs before P1+.

## Sequencing note
P2 must precede P3 (don't build pages against an empty feed). P4/P5 can follow P3 in
parallel-ish (both consume curated items). P0 should land before P6's log reading is trusted.

---

## P6 outcome (2026-07-19) — met, with the caveat this plan predicted

Reactivation measured: legacy feed 404→200, 100% failure → ~97.6% success (122/3 on the
first full day). Evidence and query in the running notes.

**The plan's own sequencing note was right:** "P0 should land before P6's log reading is
trusted." It is exactly the binding constraint. Every client IP arrives as a Cloudflare
edge address, so the 200-count is measurable but *subscribers* are not countable at all.
P0 (CF real-ip `setup.sh` re-run) is therefore **promoted from housekeeping to a
measurement prerequisite** — not cosmetic. Second correction to its framing: `setup.sh` is
no longer present on the box; re-scp before re-running, and reconcile the drifted nginx
conf first (see traps).

Also worth recording against P6's exit criterion: the raw 200 count flatters. Crawlers
(Google/Meta/Apple) re-discovered the URL once it stopped erroring and account for over
half the successes. Non-crawler fetches are ~55, the strongest genuine signal being a
scheduled `Apache-HttpClient` poller. "Measurable reactivation" is met; "how many real
subscribers" remains unanswerable until P0.

---

## P7 — Guías + Glosario starter content (operator decisions, 2026-07-19)

### Decisions and their reasons

**D1. Sourcing: cite-or-omit, grounded in our own corpus.** Content may assert only what
traces to an ingested sourced item we already hold, or to a manufacturer's own published
specification. Anything else is omitted rather than guessed.

*Why:* relojistas.com is a live public site and watch guides carry checkable factual
claims — service intervals, water-resistance ratings, movement mechanics. This platform
has published LLM fabrications on real sites twice (vetcomparison's invented prices;
leopardess marketing claims), and both were caught after publication. The corpus supports
this choice: 50 relevant items, **34 at `credibility='high'`**, with `source_url` and
`source_attribution` already populated per item.

*How the rule is applied (my reading, recorded so it can be argued with):*
- **Glossary definitions are vocabulary, not world-claims.** "Un tourbillon aloja el escape
  en una jaula giratoria" is definitional and needs no per-sentence citation. But the term
  must be one the corpus actually uses — we document the vocabulary our own readers meet,
  we don't invent a dictionary — and any *attribution* (who invented it, when, which brand
  first used it) must cite.
- **Guides are advice and carry the real risk.** Anything numeric or brand-specific must
  cite. General principles may be stated non-numerically.
- **Never, under any framing:** invented prices, service intervals, water-resistance
  figures, or dates. The maintenance guide in particular must say "consult your
  manufacturer's published interval", never name one.

**D2. Volume: ~4 guides, ~12 glossary terms.** Enough that both indexes list something
credible and neither reads as abandoned, without committing to a content operation we
cannot sustain. Grows from the news feed later.

**D3. Guide subjects are corpus-led, not invented.** Chosen against live topic counts:
buceo/dive watch (4 items), alta relojería (8) + tourbillon (3) + cronógrafo (3) + reserva
de marcha (3), edición limitada (6) + coleccionismo (4). The fourth is **mantenimiento** —
deliberately, because `/guias/mantenimiento` is one of the three phantom links on the
homepage, so authoring it converts an invented link into a real one instead of deleting it.

**D4. Glossary terms drawn from corpus topics:** tourbillon, cronógrafo, reserva de marcha,
movimiento automático, calibre, complicación, edición limitada, hermeticidad/buceo, esfera,
bisel, corona, acero inoxidable.

### Route (forced by the code, not chosen)

Only `reconcile_site_plan` yields `/guias/<slug>.html` — `apply_gap_plan` hardcodes a flat
`/<name>.html` and `create_blog_posts` hardcodes `/blog/<name>.html`. So: insert
`site_plan_pages` (+ sections) and `pages` rows at the current `plan_id`, then reconcile.
Reconcile returns `skip_built` for anything already deployed at the current plan version,
which is the property that makes it non-clobbering — the same reason the targeted
`/noticias` fix was safe. Full mechanism and citations in the running notes.

### Ordering correction to the handoff

The handoff implies children must exist before the index pages can build. **They need not.**
An empty listing renders as an empty list — it does not fail or defer (`resolvePagesWhereType`
returns a non-nil empty slice; `min_items` is parsed but never compared). What actually
blocks `guias-index`/`glosario-index` is `pages.sections = []`. Children are still authored
first here, but for an editorial reason, not a technical one: building an index with zero
children would immediately trip the `empty_sections` discovery check.

### Open risk carried into P7

`articulo` and `glosario-entrada` are **not** templates (see the correction in the running
notes) — they are two ordinary unbuilt pages that will publish as "Artículo" and "Glosario
Entrada" and then list *themselves* inside these very indexes. Their disposition must be
settled before either index is built, or the first thing a visitor sees in the Glosario is
a page called "Glosario Entrada".
