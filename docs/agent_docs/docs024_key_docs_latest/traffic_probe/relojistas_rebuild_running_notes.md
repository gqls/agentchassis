# relojistas.com — Rebuild Running Notes (dated log)

Append-only log for the relojistas rebuild (HANDOFF Thread A). Companions:
`relojistas_rebuild_manifest.md`, `_plan.md`, `_runbook.md`. Umbrella workstream log
remains `traffic_probe_running_notes(28).md`; this file is the domain-scoped Thread-A log.

---

## 2026-07-15 — last traffic read + rebuild kicked off (manifest/plan/runbook drafted)

**Context.** Operator asked for a final read of relojistas traffic, then to submit the
domain for a complete rebuild onto static hosting including an outbound newsfeed sent via
RSS to the *same* legacy RSS link that is still being hammered, linking out to sources for
now, content + news in Spanish.

**Traffic read (live off the box, 9–14 Jul window, 5.5 days).**
- 405,701 requests; **83.5% 404** (dead vBulletin paths), **15.8% 301** (www/https),
  **only 0.34% 200** (1,367) — of which **89 were the real homepage**, 8 beacon. Still
  ~99.7% bot/crawler noise at the request level.
- **RSS is the asset:** `/external.php?type=RSS2` pulled **749×** (~136/day, steady:
  87/145/172/144/112/89), 31 distinct CF-masked sources, UAs incl. **FeedFetcher-Google**
  (real Google subscriptions), **meta-webindexer**, **Applebot**, DotBot, feed readers —
  **all 404**. Subscribed variants: `forumids=2,4,13,44,58,78,145,288`, per-user `cat/ppuser`.
- **Organic intent now exists:** 9 captured events, **~8 organic**, over 3.5 weeks from
  **ES/CL/MX** (`Omega seamaster` CL, `Certina 919` ES, `Casio shock ga2100` MX, `Aguja 0.18`,
  `president 22`, `Rolen date`, `Omega 285`, `Casio`). This **updates** the day-one
  "human ≈ 0" verdict — small but real demand.
- **Interpretation:** reactivating the feed with live Spanish watch news is the
  highest-leverage move (136 daily 404s → 136 daily subscriber touches). Rebuild justified.

**Capability check (what already exists vs net-new).**
- News **ingest→triage→curate is live** (`content-feed-trigger`→`orchestrator`→`ingester`→
  `triage`→`render_news_section`→`/data/latest-news.json`); sources are pure data rows.
- **VM deploy plumbing has landed** since the June handoff: `resolveGitRepoName`,
  `deploy_config.target='vm'`, `check_backend_unreachable` all present in the Go source.
- **Net-new = outbound RSS.** The pipeline emits JSON, not RSS XML; `feed-publisher` is
  unbuilt. So we add one thin `render_rss_feed` step + a legacy-URL handler.

**Decisions taken (via operator Q&A).**
- **D1 Host = existing CF-proxied box.** Pure object storage can't answer a query-string
  URL (`/external.php?type=RSS2`); the box + nginx can rewrite/serve it. (chosen over
  object-storage + CF-edge migration.)
- **D2 Keep the engine + a clever backend feature.** Operator: "we can still have the engine
  and think of something clever to deliver with it along with the static pages." Proposed
  A+B: (A) dynamic per-category legacy feeds mapping old `forumids`/`cat` → our topic feeds
  (the thing static can't do; maximises subscriber reactivation); (B) search that *answers*
  from curated news while still capturing intent. C (live "búsquedas recientes" wall) held
  to v2 (needs moderation).
- **D3 Manifest/plan first, then build** (operator chose review-before-build).
- **D4** News links out to sources, title+summary only (brief + rights-safe).
- **D5** Spanish throughout; vertical = relojería; framing = watch news portal on forum
  heritage (to confirm).

**Deliverables this session.** `relojistas_rebuild_manifest.md` (full spec),
`relojistas_rebuild_plan.md` (phases P0–P6), `relojistas_rebuild_runbook.md` (exact
commands; §0 traffic read proven), and this log.

**Next.** On approval: P0 CF real-ip re-run → P1 sites row + `news_feed.recommended` →
P2 seed sources + first ingest → P3 chassis build → P4 `render_rss_feed`/`/feed.xml` →
P5 engine legacy handler + search-answers → P6 verify subscriber 404→200 flip.

**Open.** Confirm framing (§ manifest 2); pick clever-engine option(s) (A+B recommended);
verify real RSS source URLs before seeding (prefer api_news); CF real-ip is a P0 prereq.

## 2026-07-15 (b) — operator ACCEPTED all three open items

- **Framing CONFIRMED:** news portal on forum heritage (D5 locked).
- **Clever-engine CONFIRMED:** build **A + B** — (A) dynamic per-category legacy feeds
  mapping old `forumids`/`cat` → our topic feeds; (B) search that answers from curated news
  while still capturing intent. (C "búsquedas recientes" wall stays v2.)
- **Sourcing CONFIRMED:** verify real RSS source URLs before seeding; **Grok `api_news`
  primary**, and **Gemini as an additional `api_news` provider later** (operator: "maybe
  later also via Gemini"). No blind/fabricated feed URLs.
- All open items cleared → path to P0 is open. P0 (CF real-ip setup.sh re-run) is a
  production-box change the operator runs; the safe unblocked prep is RSS-source
  verification (below) + drafting the P2 seed artifact.

### RSS source verification (for P2 seed) — done 2026-07-15
Each candidate feed fetched + inspected (format, language, recency). **Verified live +
on-vertical Spanish (seed these):**

| Source | Feed URL | Format | Last item (checked 07-15) |
|--------|----------|--------|---------------------------|
| Debajo del Reloj | https://www.debajodelreloj.com/feed/ | RSS 2.0 es | 15 Jul 2026 (today) |
| Tiempo de Relojes | https://tiempoderelojes.com/feed/ | RSS 2.0 es | 13 Jul 2026 |
| TR Magazine | https://trmagazine.es/feed/ | RSS 2.0 es | 13 Jul 2026 |
| Máquinas del Tiempo | https://www.maquinasdeltiempo.com/feed/ | RSS 2.0 es | 08 Jul 2026 |
| Relojes y Estilo | https://relojesyestilo.es/feed/ | RSS 2.0 es | 23 Jun 2026 |

**International (English, optional — v2 with translation / an "internacional" category):**
Monochrome Watches `https://monochrome-watches.com/feed/` (RSS 2.0, updated today).

**Rejected:** El Cronómetro (`elcronometro.com/feed/` — stale, last Jul 2025; likely a
retailer blog), Relojesmanía (`relojesmania.com/feed/` — stale 2024-25 + smartwatch/
off-vertical), Hora Latina / Europa Star ES (`horalatina.com/feed/` → 404, no feed there).

**Primary generative source:** Grok `api_news` (grok-4-1-fast Responses API, web_search+
x_search) with a Spanish watch-news prompt — fabrication-safe. **Gemini** to be added as a
second `api_news` provider later (operator). 5 verified RSS + Grok gives strong source
diversity (triage caps each source at ~2 of 6 slots). Re-check feed liveness at seed time.

### P1/P2 SQL artifacts drafted — `relojistas_rebuild_seed.sql`
Schema verified against the repo before writing (`content_sources` DDL 027; `site_specs`
versioning + `content_features.news_feed` structure from `feed_news_recommendation_action.go`
/ `seed_content_sources_action.go`; onboarding UPDATE from `intent_events_migration(1).sql`).
Contents: **P1a** onboarding UPDATE (github_repo=vm-sites, deploy_config target=vm +
capabilities=[backend] + engine.{base_url,stats_key}); **P1b** `relojistas_set_news_feed()`
(versioned site_specs merge, `recommended=true`, `source_types=[rss,api_news,news_search]`,
Spanish `vertical_keywords`); **P2** `seed_relojistas_sources()` (the 5 verified RSS rows +
the Spanish Grok row).
- **Key design finding:** `SeedContentSourcesAction` **skips `rss`/`scrape`** ("requires
  manual URL config") — only auto-seeds `news_search` (per keyword) + one generic
  `api_news` named `LLM News: <domain>`. So (a) the 5 RSS rows MUST be inserted explicitly,
  and (b) our Grok row is pre-inserted under that exact canonical name with a **Spanish**
  prompt, so the auto-seeder's `ON CONFLICT DO NOTHING` no-ops and our config wins.
- **Gemini** left as a commented row — BLOCKED until the ingester's api_news provider
  routing supports it (today: xai/openai/perplexity). Flagged in the SQL.
- All non-destructive to prepare; applied by the operator once the `sites` row exists.
  **Still pending operator:** P0 CF real-ip box re-run; confirm/create the `sites` row.

## 2026-07-16 — SUBMITTED into the framework; P1a/P1b/P2 APPLIED live; build on the news-portal track

- **No sites row existed** (`SELECT … WHERE domain='relojistas.com'` → 0 rows). Submitted
  FRESH via `082_submit_domain_unified.sh relojistas.com --mission-file
  relojistas_mission_brief.txt` (new brief pins Spanish + news-portal + forum heritage +
  link-out-to-sources). Correlation `6cc3a05c…`; submitter orchestration COMPLETED.
- **Site row created:** `ecf15e75-a966-4900-bcb0-1c85f689dbfd` (status active, build pending).
- **Critical finding — no watch vertical:** `verticalNewsMap` has energy/gas/finance/boxing/
  tech/vet/legal… but **nothing for watches/horology**, so `evaluate_news_feed` would return
  `recommended:false` and build the site with NO news feed. On no-match it returns early
  WITHOUT writing site_specs — so a forced recommendation is safe from clobber.
- **P1a APPLIED (before deploy):** onboarding UPDATE via psql — github_repo=vm-sites,
  deploy_config target=vm + capabilities=[backend] + engine.{base_url, stats_key}. Key read
  live from the box env (`INTERNAL_API_KEY`, 48 chars, `819419…`), never printed/committed.
- **P1b + P2 APPLIED (after classifier, before planner — the exact window):** created
  `relojistas_set_news_feed()` + `seed_relojistas_sources()` and ran them.
  Classification current spec now `content_features.news_feed.recommended=true,
  separate_page=true` (source_agent `relojistas-rebuild`; classifier's version superseded).
  6 content_sources seeded (5 verified Spanish RSS + Spanish Grok api_news).
- **VALIDATED:** cascade ran classifier → vertical-exemplar → strategist → briefing →
  site-design-planner; forced classification stayed `cur=true` (not clobbered). Planner
  produced **27 work items incl. a `noticias-index` page** + index, guias-index, articulo,
  glosario-index/entrada, historia (forum heritage), sobre-nosotros, contacto — the Spanish
  news portal, as specified. Pages `not_built`; build now proceeding (pages → design →
  deploy to vm-sites).
- **Code (future domains):** added watch/horology to `verticalNewsMap`
  (`watchHorologyNews`, aliased watch/watches/horology/watchmaking/reloj/relojes; gofmt +
  build OK). Effective only after a chassis rebuild+redeploy; did NOT affect this build.
- **Design written:** `news_vertical_autodetect_design.md` — phased plan to make
  news-vertical detection + RSS sourcing automatic (Phase 1 map entry done; Phase 2 LLM
  fallback; Phase 3 DB-backed self-learning verticals; Phase 4 VERIFIED feed-discovery — the
  safe sourcing automation). Careful: never seed unverified feed URLs; LLM only on miss;
  discovery async off the hot path.
- **Next:** watch build → deploy; then P4 `render_rss_feed` + P5 engine legacy
  `/external.php` handler + search-answers. P0 CF real-ip box re-run still operator's.

## 2026-07-16 (b) — news pipeline PROVEN live; build partial; a real repo-routing bug found

- **News pipeline WORKS end to end.** Triggered `content-feed-orchestrator` for the site
  twice (kcat, adapted from `080_test_content_feed_orchestrator.sh`). Pass 1: 4/5 RSS feeds
  fetched, **33 items ingested**, rendered 3. Pass 2 (two-pass triage): **15 items scored
  `relevant`** (relevance 83–85), rendered a full **6-item** `latest-news.json`. Real,
  current Spanish headlines (Vacheron Overseas Tourbillon, Richard Mille RM 64-01, AP Neo
  Frame, Patek boutique España, Panerai PAM01756, Grand Seiko 62GS, Zenith DEFY). The 5
  verified feeds + Grok were the right call.
- **BUILD partial:** pages `index` (home), `historia`, `contacto` built; `sobre-nosotros`
  building; `noticias-index` + 4 scaffold pages (`guias-index`, `articulo`,
  `glosario-index/entrada`) in `needs_human_review` — "no sections ready to build": they
  have no content behind them yet (news data/components arrive via the feed enrichment; the
  guide/glossary pages need seed articles/entries; `contacto` wanted business contact info).
  Site `build_status=pending` — **nothing deployed yet.**
- **BUG FOUND — feed commits misroute for VM sites.** The feed render committed
  `/data/latest-news.json` to **`gqls/sites`** (→ B2), NOT `gqls/vm-sites` (→ the box),
  despite `github_repo='vm-sites'`. Root cause: `resolveGitRepoName` (helpers.go:207) reads
  `site_record.github_repo` from the workflow's collected_data → the **content-feed-
  orchestrator never loads the site record**, so it defaults to `"sites"`. So news for any
  VM-hosted site perpetually lands in the wrong repo. Fix options: (a) load
  `site_record.github_repo` in the feed orchestrator's workflow so resolveGitRepoName sees
  it (correct, general); (b) explicit `repo_name` on the commit step (per-site/global, less
  clean). **VERIFY before trusting P3 deploy:** confirm the main BUILD's deploy step carries
  `site_record.github_repo` (idea.uk→vm-sites reportedly works, so the build path likely
  does — but confirm, or pages misroute to `sites` too).
- **Decisions for operator:** (1) fix the feed repo-routing before/with deploy; (2) scaffold
  pages — author seed guide/glossary content, or trim those pages from the plan (not core to
  a news portal); (3) `contacto` details; (4) then P4 `render_rss_feed` + P5 engine handler.

## 2026-07-16 (c) — vm-sites repo architecture reviewed; misroute is WIDER than the feed

Operator asked for a hard think on vm-sites as a separate repo (several sites will move
static→dynamic; difference is primarily the deploy stage) + prior-discussion archaeology.
**Full report: `REPORT_vm_sites_repo_architecture.md`.** Key facts established:
- Prior decisions found: June probe notes (repo created by hand-private; `github_repo`
  selects target — designed, unwired) and the **idea.uk workstream 14–16 Jul** (Class A
  static→B2 / B static→VM+backend / C dynamic-render REJECTED; **pull-not-push** with
  per-box sparse-checkout of vm-sites; `deploy-targets.json` allowlist for the legacy push
  Action; four dead wires shipped v1.0.1123). Owner constraint on record: thousands of
  domains ⇒ per-site repos ungainly.
- **WIDER BUG CONFIRMED (live, v1.0.1125):** not just the feed — relojistas **page deploys
  also committed to `gqls/sites`** (14:17, repo_name='sites'). `page-rerender` /
  `build-dispatch-loop` orchestrations carry **no site_record at all**; only planner-tier
  workflows run `ensure_site_record`. So ALL relojistas artefacts currently land in the B2
  repo, invisible; the box webroot still has the probe page. The shipped wiring resolves
  from workflow state when it should resolve from the site row.
- **Verdict:** keep vm-sites separate (deploy keys are repo-scoped → box blast-radius cap;
  sink separation by construction; repo flip = the A→B migration primitive). Fix forward:
  (1) `git_commit`/`deploy_image_asset` fall back to `SELECT github_repo FROM sites WHERE
  domain=$1` — workflow-independent routing, kills the class; interim data-only fix = add a
  load-site step to content-feed-orchestrator; (2) apply the vm-sites Action **allowlist
  now** (relojistas-only) before any second domain lands; (3) `deploy_config.target` =
  semantic truth, consistency-check `github_repo`; (4) script the A→B migration incl. stale
  `sites`/B2 cleanup; (5) converge push→pull; (6) Class B stays the exception.
