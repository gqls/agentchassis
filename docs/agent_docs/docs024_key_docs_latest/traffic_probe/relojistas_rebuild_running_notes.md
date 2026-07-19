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

## 2026-07-16 (d) — routing fix BUILT, SHIPPED (v1.0.1126), PROVEN LIVE; news serving on relojistas.com

- **Code:** `resolveGitRepoNameDB` in helpers.go (explicit config → collected site_record →
  `SELECT github_repo FROM sites WHERE domain=$1` → "sites"); wired into `git_commit`
  (domain extracted first, then resolve) and `deploy_image_asset`'s `sendGitCommitRequest`.
  7-subtest table in `git_repo_resolution_test.go` — all PASS (sqlmock; covers precedence,
  NULL repo, unknown domain, DB error → safe default, nil DB/empty domain).
- **Multi-session moment:** the working tree shifted mid-build (directory-export refactor
  transiently broke compile). Discovered `multi_session_coordination/` + new repo-root
  CLAUDE.md ruling: explicit-pathspec commits, **ref-builds** (`build-agent-chassis-ref` =
  git-archive, cannot bundle WIP). My changes were swept into another session's commit
  `87d13b864` — HEAD-archive verified compiling standalone → ref-built v1.0.1126.
  **Operator shipped it** (push/kustomize/rollout theirs).
- **Verified against the pod** (doctrine: pod, never tag): image v1.0.1126, `strings`
  finds resolveGitRepoNameDB ×3.
- **Allowlist (rec 2) already applied by another session/operator:** vm-sites commit
  `d0fb7a1` — deploy-targets.json = {"relojistas.com": "167.233.33.159"}. Confirmed on
  origin/main.
- **rsync --delete trap disarmed BEFORE first pipeline deploy:** vm-sites' relojistas.com/
  contained only assets/js/snippets.js (a 12:44 commit that never deployed — flaky runner),
  while the box webroot held index+gracias from the June manual rsync. First Action run
  would have wiped them. Mirrored the LIVE box files into the repo (checksums verified),
  explicit-pathspec commit `19debed`, push → Action ran → box intact (checksums unchanged),
  assets/ + data/ arrived.
- **ROUTING FIX PROVEN LIVE:** feed pass 3 → orchestration COMPLETED, `repo=vm-sites`,
  6 items rendered → Action → box → **https://relojistas.com/data/latest-news.json serving
  real Spanish watch news** ("Últimas noticias de relojería"; Patek Milán etc., link-outs
  to sources). News layer live end to end.
- **Page cutover fired:** 4 built pages (index/historia/contacto/sobre-nosotros) flagged
  needs_rebuild + page-rebuild orchestration dispatched — commits should now route to
  vm-sites and replace the probe page. NOTE: new index = hero/latest-news/info-card-grid/CTA
  — **no search form** until P5 (engine search-that-answers), so intent capture pauses at
  cutover; accepted per plan sequencing.

## 2026-07-16/17 — CUTOVER COMPLETE: the Spanish news portal is LIVE on relojistas.com

- **The misroute had a third layer.** After the DB fallback shipped (v1.0.1126): the first
  page-rebuild dispatch died silently (spawn near the chassis rollout — the CLAUDE.md ~300s
  warning, no logs for the correlation); the retry FAILED structurally (page-rebuild →
  page-content-writer `resolve_links` contract violation — page-rebuild regenerates content,
  wrong tool anyway). Switched to **page-rerender** (re-render existing content + commit).
  Round 1 STILL went to `repo=sites`: **three agent definitions hardcode `repo_name:"sites"`
  in their git_commit step config** — page-rerender(deploy_page), site-deployer(deploy_to_git),
  deployer-agent(commit_to_git) — and explicit config rightly outranks the DB fallback.
  Removed the three pins via jsonb `#-` (data-only, immediate; diagnosis/fix agents pinning
  `agentchassis` left untouched; med-json-exporter left — Class A anyway). Verified 0
  git_commit steps pin repo_name.
- **Round 2: all four rerenders COMPLETED, `repo=vm-sites`** (index files=2 → html +
  latest-news component JS). Action deployed within ~1 min. Box now: index.html 29,772 B
  (was the 2,337 B probe page), + tools/ (component JS), + data/latest-news.json.
- **LIVE:** https://relojistas.com/ → 200, `<title>Relojistas — Portal de noticias de
  relojería en español</title>`, news section wired (fetches /data/latest-news.json —
  serving 6 curated Spanish items). historia/sobre-nosotros/contacto all 200.
- **The 6-hourly heartbeat has adopted the site on its own** — three autonomous "Update
  latest news feed" commits in vm-sites after ours (content-feed-trigger keying on
  news_feed.recommended=true). News now refreshes without manual triggering.
- **Fix summary (the whole onion):** (1) v1.0.1126 `resolveGitRepoNameDB` DB fallback
  [workflows without site_record]; (2) three agent-definition repo_name pins removed
  [explicit config overrode everything]; (3) rsync --delete disarmed by mirroring the box
  webroot into the repo BEFORE first pipeline deploy; (4) allowlist already in place.
- **Remaining (unchanged plan):** P4 `render_rss_feed` → /feed.xml; P5 engine
  `/external.php` legacy handler (subscriber 404→200 flip = the mission metric) + B
  search-that-answers (restores intent capture, paused since cutover); compose
  /noticias/index.html (still 'planned'; the news-listing page + archive JSON); scaffold
  pages (guias/glosario/articulo) + contacto business details; stale relojistas copy in
  gqls/sites + B2 to clean per migration script (report rec 4); intent_events collector
  still disabled (P4-June).

## 2026-07-17 — P4 BUILT (render_rss_feed); Action-127 diagnosed; awaiting v1.0.1130

- **Deploy-failure report (operator, mid-turn) DIAGNOSED:** vm-sites run 29522203773
  (the webroot-mirror commit, 18:01Z) died exit **127 in "Set up SSH"** — the new
  dedicated runner pod `github-actions-runner-vmsites-…` came up at exactly that time and
  my run raced its first-boot tool install; ssh-keyscan's "not found" was swallowed by
  `2>/dev/null`, hence the bare 127. **No harm:** the next run (18:10Z news feed) rsync'd
  the whole relojistas.com/ folder incl. the mirrored files; 7+ consecutive greens since,
  including the autonomous 6-hourly news deploys (19:42/01:35/07:41 — heartbeat healthy).
  Runner pod verified: jq/ssh-keyscan/rsync all present now. **Hardened the workflow**
  (vm-sites `516b8e9`): loud tool guards for jq/ssh-keyscan/rsync/sort + WARN on keyscan
  failure. (Cosmetic: GitHub masks "deploy" in logs because VM_USER's value is "deploy" —
  that's why "deploy-targets.json" renders as "***-targets.json".)
- **P4 CODE COMPLETE** (committed, explicit pathspec): `render_rss_feed_action.go` —
  outbound RSS 2.0 (atom:self = the legacy `/external.php?type=RSS2` URL; items link OUT
  to sources; dedupe by URL; RFC1123Z dates; created_at fallback), **per-site gated by
  `deploy_config.rss_feed.enabled`** so the shared workflow is fleet-safe (non-opted-in
  sites return item_count=0 → conditional skips commit). Registered in registry.go
  (registry was clean vs HEAD — directory-export session had committed). Tests PASS
  (TestLoadRSSItems: escaping/atom:self/dedupe/date-fallback/round-trip;
  TestRenderRSSFeedGate: skip shape carries no files).
- **Activation artifact ready:** `relojistas_rss_feed_apply.sql` — adds
  commit_news→render_rss_xml→check_has_rss→commit_rss to content-feed-orchestrator +
  enables the relojistas gate (Spanish channel metadata). **Apply only after the image is
  live** (image first, then seeds). Caveat noted: a no-news cycle also skips the RSS
  refresh (check_has_news short-circuits) — acceptable.
- **Ship:** operator building **v1.0.1130** (includes render_rss_feed via HEAD). Watcher
  armed for rollout + pod symbol check; then: apply SQL → trigger feed pass → verify
  feed.xml on box + well-formed over HTTPS.

## 2026-07-17 — P4 PROVEN LIVE; v1.0.1130→1134 survived; images migrated; **P5.1 DONE**

- **P4 feed.xml LIVE:** after v1.0.1130 (then 1134) rollout + pod-verify, applied the
  RSS workflow SQL, ran a feed pass (rss_items=30, repo=vm-sites), and
  **https://relojistas.com/feed.xml serves valid RSS 2.0** (30 items, atom:self = legacy
  URL, items link out to sources; Go test round-trips the XML — Python check was redundant
  per CLAUDE.md Go-not-Python). Kafka dispatch vanished once silently (no orchestration
  row) — refired, fine. **v1.0.1134 redeploy verified:** both symbols in pod; all live DB
  config survived (0 re-pinned git_commit repos, RSS steps intact, gate+target+forced
  news_feed intact); all surfaces 200.
- **Missing-images fixed (operator report):** logo + 4 heroes + 4 icons + styles.css had
  misrouted to gqls/sites pre-routing-fix; migrated into vm-sites (commit cfd787c), all
  200 now. Gap: homepage references `/assets/images/favicon.png` — never generated.
- **P5.1 LIVE — the mission metric:** legacy `/external.php?type=RSS2` (and all
  `forumids`/`cat` variants) now **200 → serves feed.xml** (~136/day subscriber 404s
  reactivated). Bare `/external.php` still 404 (correct). Done as a surgical nginx location
  (`if ($arg_type = RSS2) { rewrite ^ /feed.xml last; }`, the nginx-safe if-use) applied
  live: backed up conf, `nginx -t` OK, reload. Master-feed-to-all; per-forumid category
  feeds = later engine step (needs category feeds rendered first). Served statically so it
  survives engine downtime. **CAVEAT:** the vm-sites.conf is setup.sh-managed and the box
  is drifted from every repo setup.sh version (box predates the /events endpoint) — my
  location must be reconciled into the setup.sh generator before any future re-run, else
  it's lost. Backup: /root/vm-sites.conf.bak-p5-*.

## 2026-07-17 — NAV DEAD-ENDS diagnosed (operator report): two classes, one root-cause

- **Class 1 — header nav to planned pages (404):** Noticias, Guías, Glosario are all
  `build_status='planned'` with 0 sections. **Root cause for Noticias (the core one):** the
  planner typed it `page_type='section-index'`, NOT `'news-index'` — so render_news_section
  (which only produces news-archive.json + attaches news-listing for `page_type='news-index'`)
  never composed it, and the `missing_news_page` discovery check (fires on separate_page=true
  AND no news-index page) would route it to content-gap-planner to build a proper
  news-listing page. Guías (section-index) + Glosario (entity-directory) have NO child
  content (no guides, no glossary terms) → would be empty listings.
- **Class 2 — invented CTA links in the homepage info-card-grid (404):** the LLM-written
  `info-card-grid` cards link to `/ferias`, `/archivo`, `/guias/mantenimiento` — pages that
  were never planned. Baked into the homepage's rendered component, not the nav.
- **Fix forks on the deferred scaffold-pages decision** (build content vs trim) + needs the
  news-index re-typing for Noticias + card-link repair. Risk: composing empty pages via a
  full re-plan is the known clobber landmine — must use targeted routes
  (missing_news_page→content-gap-planner for news; nav-updater/link repair for cards), not
  build-site-planner. Presented to operator for direction.

## 2026-07-17 — NAV-BUILD plan (operator: build Noticias + author Guías/Glosario content)

Doc-searched (006_news_feed_pipeline_v2, FOCUS_navigation) + DB-verified. Findings:
- **Header dead-ends are the DOCUMENTED nav bug — already code-fixed.** `GetHeaderNavFromPages`
  / `GetFooterNavFromPages` (component_library.go:1731/1792) already filter
  `build_status='deployed'`. So Noticias/Guías/Glosario show as dead links only because
  they're `planned`. **Building+deploying them IS the fix** — they resolve and the filter
  keeps them honest. (The live rendered nav still lists them because the pages were rendered
  earlier; a rerender after they deploy refreshes it.)
- **Exact page records:** `noticias-index` (/noticias/index.html, section-index),
  `guias-index` (/guias/index.html, section-index), `glosario-index` (/glosario/index.html,
  entity-directory), + templates `articulo` (blog-post) and `glosario-entrada` (entity-page).
  Sections live in the **`resolved_composition`** aspect (there is NO `site_plan` aspect).
- **news-listing + latest-news components EXIST** (content_components, 1 each) — Noticias is
  buildable. Reference: gaswholesalers/robot-hands `/news.html` news-index =
  `[hero, news-listing, call-to-action]`, archive JSON rendered because page_type='news-index'.
- **Proven Noticias chain (006 §Ongoing enrichment):** missing_news_page check →
  content-gap-planner (approach=new_page) → apply_gap_plan (page+nav+build item) →
  page-build-handler (composes news-listing) → rerender → content-feed-refresh renders
  news-archive.json. CAVEAT: gap-planner CREATES a page; we already have an empty
  section-index `noticias-index`. Plan = **re-type noticias-index → news-index** and drive
  the build onto it (keep the Spanish /noticias URL + nav), rather than let gap-planner mint
  an English /news.html. render_news_section then produces /data/news-archive.json (it gates
  on page_type='news-index').
- **Guías/Glosario need CHILD content** (no guides, no glossary terms) → author starter
  Spanish content: guide articles (blog-post instances under /guias) + glossary term
  entities (entity-page instances under /glosario), then the index pages list them.
- **Invented card links** (/ferias, /archivo, /guias/mantenimiento) live in the homepage
  `info-card-grid` component's content_data (LLM-fabricated) → repoint archivo/ferias →
  /noticias, drop/repoint mantenimiento; re-render index.
- **Task order:** (4) Noticias news-index build → (7) card-link repair + index rerender →
  (5) Guías articles → (6) Glosario terms. Tasks tracked in the session task list.

## 2026-07-18/19 — /noticias LIVE; council caught a duplicate; handoff written

- **NOTICIAS BUILT + DEPLOYED.** The targeted fix worked exactly as diagnosed: re-type
  `noticias-index` section-index→**news-index** + `sections=[hero,news-listing,call-to-action]`
  (copied from the gaswholesalers reference) + re-queue the `needs_page` item. No re-plan, so
  no clobber. Verified live: `/noticias/` 200, `<title>Noticias de relojería | Relojistas</title>`,
  carries `data-component="news-listing"`, and `/data/news-archive.json` now renders **20
  items** (render_news_section gates the archive on page_type='news-index' — which is exactly
  why it produced nothing before). Work item `complete`.
- **COUNCIL GATE used for the first time** on the bugs_open/015 fix (a proposed
  `unresolvable_internal_links` check). `SUBMISSION_CORR=8a1f7b4f`, run `745e1913`.
  **Outcome: withdrawn as a DUPLICATE — the council was right.**
  - The reuse seat's own suggested check ("search for an existing link-resolution helper
    rather than reimplementing href extraction") surfaced
    **`discovery_checks/check_phantom_internal_links.go`**, which already does it, reusing the
    shared `ExtractHrefs`/`ClassifyLinkScope`/`PageURLSet` datahelpers that the deploy gate
    uses — my draft would have reimplemented extraction.
  - A second seat found the real flaw: I routed BOTH defect classes to `page-rerender`, but a
    rerender **re-emits an LLM-invented href unchanged** from component content_data and then
    marks the item resolved — silent-success. It routes by surface instead
    (`nav-link-fixer` vs `page-build-handler`).
  - Decisive nuance: the phantom check treats "real page" as *a pages row exists*, so
    **planned-but-unbuilt pages are deliberately NOT flagged** → it covers /ferias, /archivo,
    /guias/mantenimiento (no rows) but not Noticias/Guías/Glosario (rows, unbuilt).
    It IS enabled in `completeness-discovery-agent` but has 0 work items fleet-wide — it has
    simply never swept this site. **So task 7 needs no code, just the sweep.**
  - Correction recorded in `bugs_open/015`.
- **INFRA FINDING (unfiled, operator to decide):** the council run produced **no verdict
  artifact** — `council_decide` died on `review_editquality.result` being *"invalid JSON —
  likely truncated at max_tokens"* (the 005/008/012 truncation family). Reviews survived only
  in `orchestration_states.collected_data`; the coverage report will score such runs MISMATCH
  rather than reviewed. The council thread owns that code.
- **Handoff + summary written:** `HANDOFF_RESUME_relojistas_rebuild.md` (fresh-chat entry
  point: coordinates, live-verified state, remaining tasks, traps) and a rewritten
  `SUMMARY_relojistas_rebuild.md` (read-aloud, full arc).
- **Remaining:** Guías/Glosario content (operator chose author-starter-content); phantom
  sweep for invented links; P5.2 search-that-answers (intent capture paused since cutover);
  per-forumid category feeds; measure the 404→200 reactivation; housekeeping (stale
  gqls/sites copy + B2, favicon.png, CF real-ip re-run, intent_events collector).

## 2026-07-19 (session 2) — TASK 5 DONE: reactivation measured; CF real-ip is load-bearing

Handoff `HANDOFF_RESUME_relojistas_rebuild.md` re-verified end to end before acting — every
claim in its §3 table held (all 8 URLs returned the stated codes; the 9 `pages` rows match
§4 exactly: `guias-index`/`glosario-index` `planned` with `sections=0`, `noticias-index`
`news-index`/`deployed`/3 sections). No corrections needed to the handoff.

**Mission metric — legacy feed 404→200.** Counted across both live logs (`access.log.1`
covers 9–15 Jul, `access.log` 15–19 Jul):

```
ssh root@167.233.33.159 'cat /var/log/nginx/access.log.1 /var/log/nginx/access.log \
  | grep -i "external\.php" \
  | awk "{split(\$4,a,\":\"); print substr(a[1],2), \$9}" | sort | uniq -c'
```

| day | 200 | 404 |
|---|---|---|
| 09–16 Jul | 0 | 49–122/day |
| 17 Jul (cutover) | 29 | 91 |
| **18 Jul (first full day)** | **122** | **3** |
| 19 Jul (to ~11:00Z) | 50 | 3 |

100% failure → ~97.6% success. (The 301s in the raw counts are the http→https redirect and
are re-counted as a 200/404 on the follow-up line — not failures, don't add them in.)

**Who is actually fetching it — the number flatters us.** Of ~201 successful fetches:
Googlebot 56, meta-webindexer ~79, Applebot 4, `curl/8.14.1` 6 (almost certainly our own
verification). Excluding known crawlers leaves **55 non-crawler fetches**. The strongest
genuine-subscriber signal is `Apache-HttpClient/4.5.5 (Java/1.8.0_181)` — a scheduled
server-side poller, the exact class the mission targeted. Crawler re-discovery is a real
benefit (the feed gets indexed) but it is NOT a reactivated subscriber; do not report the
raw 200 count as subscribers.

> **Trap for anyone repeating this: subscriber counts are currently IMPOSSIBLE.** Every
> client IP in the log is a Cloudflare edge address (172.70.x, 104.22.x). "86 distinct IPs
> on 200s" = 86 CF nodes, not 86 subscribers. The outstanding **P0 CF real-ip `setup.sh`
> re-run** (filed as housekeeping in the handoff §6) is therefore load-bearing for
> measurement, not cosmetic — promote it if a real subscriber count is wanted. Note
> setup.sh is no longer on the box; re-scp first.

**Residual 404s are three named variants, not a long tail.** Post-cutover only (18–19 Jul):
`/external.php?type=rss2` (lowercase) ×3, `/ventas/external.php?type=RSS2&cat=…&ppuser=…`
×2, bare `/external.php` (no params) ×1. Pre-cutover the 404 list was dominated by
`forumids=` variants — those now correctly fall through to the master feed, so the
handoff's "all `type=RSS2` variants get the master feed" is confirmed *for the uppercase
spelling only*; lowercase is a genuine miss.

Deliberately NOT fixed by a fourth surgical nginx edit: the conf is setup.sh-managed and
already drifted (handoff §6 trap), and each hand-edit deepens the debt that one re-run
erases. Reconcile the legacy-feed location + these three variants into the generator
together.

## 2026-07-19 (session 2) — TASK 2 detection DONE; a handoff premise CORRECTED

**Phantom sweep run — the council was right, no code needed.** Dispatched
`completeness-discovery-agent` (pattern from `FOCUS_dispatch_diagnostic(4).md:130-159`;
orchestration `cb80a881-9ac4-4d86-861d-86fb54129b7e`). It found all three invented links
plus three defects nobody was looking for:

| item_type | summary | handler |
|---|---|---|
| `phantom_internal_link` ×3 | `/archivo`, `/ferias`, `/guias/mantenimiento` in `index:info-card-grid` | `page-build-handler` |
| `empty_section` | `news-listing` on `noticias-index` | `page-build-handler` |
| `needs_rerender` | 9 pages missing header/footer — need reassembly | `rerender-pages` |
| `page_rerender` | 4 misdirected CTAs on `index` — copy names a different page than the link target | `page-rerender` |

Routing confirmed correct: the phantom items go to **`page-build-handler`, not
`page-rerender`** — so the silent-success flaw the council caught in my withdrawn draft
(a rerender re-emits the invented href unchanged and marks the item resolved) does not
apply to the shipped check.

> **TRAP — discovery does not dispatch itself.** All six landed at `status='detected'`, and
> per `FOCUS_dispatch_diagnostic(4).md:110-116` there is **no automated coupling between
> discovery and triage**. `detected` items sit indefinitely; the build-dispatch loop only
> claims `triaged`/`approved`. Running the sweep is therefore read-only and safe, but
> nothing happens until something promotes them.

Two of these are probably NOT what they look like — do not action them blind:
- `empty_section: news-listing` on a page that demonstrably renders 20 items live. Almost
  certainly the runtime-fill-template class (the section HTML is filled from
  `news-archive.json` at render time), which is a known false-positive shape.
- `misdirected_cta` ×4 on index overlaps the CTA/link-integrity workstream's own class
  (`bugs_open/023`) — check there before treating it as a relojistas defect.

### > **CORRECTED 2026-07-19:** `articulo` and `glosario-entrada` are NOT templates.

The handoff (§4) and this file (~line 365) both call them "templates" that spawn
instances. **There is no template mechanism in this codebase at all** — grep for
`is_template` / `template_page` / `instance_of` / `page_template_id` / `from_template`
across `platform/`, `internal/`, `pkg/`, `sql_for_tables/` returns zero. The `pages` table
has no template column.

They are ordinary one-off pages, exactly what `datahelpers.CanonicalisePage`
(`page_canonical.go:190-213`) emits for a leaf page the planner LLM happened to name in
Spanish: `articulo` → role blog-post, no ParentSection → dir defaults to `blog` →
`/blog/articulo.html`; `glosario-entrada` → role entity-page → dir defaults to `entities`
→ `/entities/glosario-entrada.html`. The "template" reading was operator shorthand in
these notes that hardened into an assumed mechanism. Caught by reading
`page_canonical.go` rather than trusting the doc.

**Consequence — they are a latent defect, not a resource.** If built as-is they become two
real published pages titled "Artículo" and "Glosario Entrada", and they will then *list
themselves* inside the Guías/Glosario indexes, because the listing query filters on
`page_type` + `status` only and ignores URL (`queryresolve.go` `resolvePagesWhereType`).
`glosario-entrada` also sits at `/entities/…`, orphaned from its own hub. Disposition
needed before either index is built.

### Route findings for authoring the child content

- **Listings are a COMPONENT behaviour, not a page-type behaviour.** There is no
  section-index builder. A component declares `"source": "query.pages_where_type:<type>"`
  in its `input_schema`; `PlanSectionsAction` (`plan_sections_action.go:1175-1207`)
  resolves it via `queryresolve.Resolve`. Selection is by **`page_type`**, ordered
  `nav_order, name`, default limit 12 / cap 24. Not by URL prefix, not by topics.
- **An empty listing renders empty — it does not fail or defer.** `resolvePagesWhereType`
  returns a non-nil empty slice, which passes the `value != nil` check, so the section is
  marked ready with zero items. `min_items` is parsed but **never compared against a
  resolved array anywhere**. So children are not a precondition for the index to build.
- **What actually blocks `guias-index`/`glosario-index` is `pages.sections = []`** →
  `plan_sections_action.go:673-680` returns "no sections to plan" → `needs_human_review`.
  That is precisely the state both are in, and it is the same state `/noticias` was in
  before the fix that worked. **The index pages need SECTIONS, not children, to build.**
- **`entity-directory` has a builder capability gap.**
  `load_work_item_actions.go:217-238` has `entity-directory`/`entity-page` **commented out**
  of `availableBuilders` and listed in `unavailableBuilders` → `write_build_items` emits a
  `capability_gap` item, never a build item. `section-index` is in neither map so it falls
  through to the `page-build-handler` default — which is why section-index builds and
  entity-directory does not on that path. **`reconcile_site_plan` ignores those maps
  entirely** (`reconcile_site_plan_action.go:200-268`) and routes everything to
  `page-build-handler`, so the reconcile route does build them.
  (Caution: `load_work_item_actions.go` is under concurrent edit by another session as of
  this writing — do not plan a change to that file without re-checking.)
- **Only ONE route yields `/guias/<slug>.html`.** `apply_gap_plan` hardcodes
  `"/" + pageName + ".html"` (`:351`, flat only) and `create_blog_posts` hardcodes
  `/blog/%s.html` (`:204`). Arbitrary paths come only from inserting `site_plan_pages` +
  `pages` rows and running **`reconcile_site_plan`**, which skips anything already
  `deployed` at the current `plan_id` (`skip_built`, `:341-355`) — that is the property
  that makes it non-clobbering. Worked example in-repo:
  `docs/social001_vonc_tiktok_social/minilobby_task/088_archetype_entity_pages.sql`.
- **`page-build-handler` does NOT create missing pages** (`load_page_record_action.go:184-193`
  returns `{found:false}` → `complete_error`). The `pages` row must pre-exist.
- **Nothing pushes work items to Kafka on insert** — `build-dispatch-loop` pulls
  `status IN ('triaged','approved')` by `pipeline`, so one kick of that loop with
  `input_data.site_id` is the only dispatch needed.
- Incidental bug spotted, not filed: `defaultSectionsForPage(pageName, pageType)`
  (`apply_gap_plan_action.go:459`) declares `pageType` but never reads it — the switch is on
  `pageName` alone, contradicting its own doc comment at `:454-458`. A new blog-post gets
  `[hero, generic-text-block, call-to-action]` rather than an article shape.

### Plan-row state behind the two stray pages (checked before touching anything)

Site has exactly one plan, `f12ab433-f7c4-4209-8c21-dcfdaed43078` (`is_current`), created
2026-07-16. `articulo` and `glosario-entrada` are **in `site_plan_pages`**, both with
`parent_section` NULL and `nav_order` 100:

```
articulo         | blog-post   | /blog/articulo.html
glosario-entrada | entity-page | /entities/glosario-entrada.html
```

The NULL `parent_section` is the whole story — that is exactly the input under which
`CanonicalisePage` defaults the directory to `blog`/`entities`. Nothing malformed happened;
the planner simply emitted two leaf pages without attaching them to a section.

**Consequence for disposition:** deactivating or deleting the `pages` rows alone would NOT
work — reconcile diffs plan-vs-realised, so the plan rows would re-emit them as `missing`
on the next run. Plan row and page row must move together, matched on `name`
(`ON CONFLICT (site_id, name)`).

**Preferred disposition — repurpose, don't delete.** Both already carry the correct roles
for what we need (`blog-post`, `entity-page`). Giving them a real `parent_section`, slug,
url and title turns a latent defect into the first guide and the first glossary term, with
no DELETE against a live plan. The pleasing case: `articulo` → `/guias/mantenimiento.html`,
which is one of the three phantom homepage links — so the invented link is satisfied by the
stray page rather than by deleting either. Requires updating plan row and `pages` row in
lockstep or reconcile will orphan one against the other.

## 2026-07-19 (session 2) — cite-or-omit is ALREADY BUILT; feed-grounding is not

Traced how page content is actually generated, to find the lever for the operator's
cite-or-omit decision. Three results, one of them a filed bug.

**1. `pages.content_direction` is a DEAD COLUMN — filed as `bugs_open/025`.** Its own SQL
comment says "Passed to content-writer prompt", and nothing reads it. The trap is a name
collision: `content_direction` appears all over the Go source, but every hit is the
*site-level* `site_specs` aspect, not the column. I planned against the column before
checking. Caught by grepping the two page loaders
(`get_pages_to_build_actions.go:98-104`, `load_page_record_action.go:167`) for the column
name and finding it absent from both SELECTs.

**2. The V2 claims-verification `writer_block` IS cite-or-omit, and it is live.** Verified
against prod, not the repo:

```sql
SELECT default_config::text ~ 'writer_block', default_config::text ~ 'Verified Facts'
  FROM agent_definitions WHERE type='page-content-writer' AND deleted_at IS NULL;
-- t | t
```

The whole claims layer is **opt-in per site purely by the presence of a `site_specs` row
with `aspect='evidence_base'`** — `loadEvidenceBase` returns nil and every check silently
skips (`validate_page_content.go:683-686`). When present it turns on three things with no
code change: the writer prompt gains a bounded "state only these" block that **explicitly
overrides STRICT RULE 14**'s unbounded "don't invent"; `validate_page_content` check 8 runs
`ScanBannedClaims` (blocker) + `ScanUnregisteredNumbers` (error) *between writer and save*;
and the post-deploy `check_unverified_claims` sweep catches drift.

> **CORRECTION to a figure I was given:** the claims docs say leopardessconsulting is the
> only site with an evidence base. **Stale — `vonc.com` has had one since 2026-07-17.** Two
> sites, not one. Grounded against the live table, per the standing rule.

Live schema, read off `vonc.com`'s row (`site_specs.data`, not `.specs` — the column is
`data`):
```
{ facts[]:          {id, kind: capability|metric, claim, value?, source:{sql|artifact}, tolerance?, verified_at},
  banned_claims[]:  {pattern (regex over LOWERCASED assertion text), reason},
  allowed_entities[]: nouns that are NOT claims,
  governing_rule:   the one-line rule the reviewers judge against,
  audit_doc, schema_notes }
```

**3. Grounding generation in `content_feed_items` does NOT exist — it is new code.** The
feed is a *link-list pipeline, not a generation corpus*: `render_news_section_action.go:341-363`
and `render_rss_feed_action.go:232` read it deterministically into JSON/RSS with no LLM,
and the only LLM that touches feed rows is triage, which writes `relevance_score`,
`credibility` and `source_attribution` back. **`source_attribution` is written by triage and
read by nothing.** No prompt template anywhere references `content_feed_items`, and
`create_article_from_feed` does not exist (zero hits across `*.go` and `*.sql`).

### What this means for P7 — the approach changes, the decision does not

The operator's cite-or-omit choice is achievable **with existing machinery and no image
roll**, but not the way the question implied. "Generate guides *from* the corpus" is new
code. "Generate guides constrained *to facts curated from* the corpus" is a config change.

So: read the 50 relevant items (34 high-credibility), curate what they actually support
into an `evidence_base` row for relojistas — `governing_rule` forbidding any unsourced
interval/rating/price/date, `banned_claims` regexes for exactly those shapes,
`allowed_entities` for the horological vocabulary that is *definitional and therefore not a
claim* — and the writer is then bounded by it automatically, with a build-time gate that
blocks rather than warns. relojistas becomes the **third** site on the claims machinery.

> **Caveat carried forward:** the repo's `sql_for_agents/` copy of the writer prompt is
> already stale relative to prod. Live prompts live in `agent_definitions.default_config`.
> Check prod before believing any prompt in the repo.

## 2026-07-19 (session 2) — P7 build: I got the sections wrong, and a no-op deployed anyway

Applied the evidence_base fence (13 facts / 9 bans / 3453-char writer_block, verified live),
created 19 pages, queued 12 child builds, dispatched. Then read the first deployed page
instead of trusting its status — which is the only reason the next three findings exist.

### MISTAKE (mine): `pages.sections` is NOT what page-build-handler reads

I set `pages.sections` on all 12 new pages and assumed that composed them. It does not.
`page-build-handler` reads its **spec sections from `site_plan_sections`**
`(plan_id, page_name, ordering, component_name)` — a table I never populated. Every build
logged, in `site_work_items.error`:

> `page-build-handler no-op: no sections ready to build (empty spec sections, or all sections deferred for missing data)`

Caught by comparing against gamesdesign.co.uk, which runs this exact shape in production and
**does** have `site_plan_sections` rows (2 per guide: `hero`, `generic-text-block`).
relojistas had them only for the 4 originally-planned pages. Fixed by inserting 28 rows.

> **Durable, and not obvious:** `pages.sections` and `site_plan_sections` are two different
> section lists and the build reads the second. The first is what the *listing/query* layer
> and rerender use. Setting one without the other yields a silent no-op, not an error.

### THE BAD ONE: a no-op build was marked `complete` and the page DEPLOYED

`glosario-tourbillon` recorded that same "no sections ready to build" no-op **and** came out
`status='complete'`, `build_status='deployed'`, live at `/glosario/tourbillon.html`. The
handler built nothing, and the page still shipped — carrying two components that are not
its own:

- `hero` with the **site homepage's** headline ("Relojería en español: noticias, guías y
  glosario") rather than anything about tourbillons;
- `content-block-about` with generic about-us copy.

So a page titled "Tourbillon" published saying nothing whatsoever about a tourbillon, and
every status field said success. This is the `016b` invariant in its purest form — *trust
the rendered artefact, not the status* — and it is worth filing separately: a no-op must not
report `complete`, and must certainly not deploy. Reset to `planned`, components deleted,
re-queued.

### MY SECOND MISTAKE: `content-block-about` on a glossary page

I copied vonc.com's entity-page shape `["hero","content-block-about","call-to-action"]`
without checking what that middle component is for. `content-block-about` writes
*about-the-company* copy — it will do that on any page, whatever the page is about. Moved
all 8 glossary terms to `["hero","generic-text-block"]`, the gamesdesign guide shape, which
demonstrably produces page-specific prose (its "The Skinner Box in Game Design" hero +
substantive body is the proof).

### What the fence DID and DID NOT do — both worth knowing

**It reached the writer.** The generated about-copy contained "No vendemos relojes, no
representamos marcas y no cobramos por recomendar nada" — that is fact R13 and the
governing_rule coming back out in the model's own words, unprompted by anything else. The
writer_block is genuinely steering.

**It does not cover links.** The same component emitted `"cta_url": "/sobre-relojistas"` —
a page that does not exist (ours is `/sobre-nosotros.html`). An invented internal link
sailed straight through, because the claims layer scans *assertion text* for banned patterns
and unregistered numbers; it has no opinion about hrefs. That is exactly the
`check_phantom_internal_links` / `bugs_open/023` class, and it means **the two layers are
complementary and neither substitutes for the other** — a page can be claim-clean and still
full of dead links. Re-run the phantom sweep after this content lands.

**It also did not stop invented stat fields.** The same block produced
`"stat_2_value": "100%"` (labelled "Independencia editorial") and `"stat_3_value": "0€"`
("Relojes vendidos"). Neither number traces to a fact. They survived because
`ScanUnregisteredNumbers` is inert here (English business-context gate — see the
evidence_base header) and because no banned pattern covers a bare "100%". Arguably both are
defensible as editorial rather than factual claims, but they are exactly the shape the gate
was supposed to catch, and on this site it cannot.

### How well the cite-or-omit fence actually held — measured, not assumed

Read both completed guides in full rather than sampling. Verdict: **the fence works, and it
is not airtight.** Both halves matter.

**What it got right.**
- `guia-mantenimiento` — the highest-risk page, because maintenance advice is where an
  invented service interval would live — contains **no numbers at all**. The model refused
  the figure in its own words, twice, and explained why: *"cuál es el intervalo de revisión
  que recomienda la marca … nosotros no publicamos intervalos genéricos porque cada calibre
  es distinto"*, and *"La frecuencia exacta depende de la reserva de marcha … que el
  fabricante especifica para cada calibre. Consulta esa cifra en la documentación oficial de
  tu modelo."* That is `governing_rule` reproduced as editorial voice, unprompted.
- `guia-complicaciones` used **four registered facts and attributed every one to the right
  model**: RM 64-01 Tourbillon Colnago as a limited edition with the cycle maker (R4), Zenith
  DEFY Extreme Ultraviolet measuring to the hundredth (R6), AP Neo Frame Horas Saltantes
  reading the time without hands (R5), Longines Conquest Heritage recovering a historic
  central power-reserve indicator (R3). No model was given a specification belonging to
  another — the failure mode I most expected.
- Zero prices, zero water-resistance figures, zero invented references across both.

**What got through.** `guia-complicaciones` asserts two unregistered facts:
`Breguet … en 1801` (the tourbillon's invention) and the perpetual calendar running
`hasta 2100`. Both are true and are standard horological reference, but neither traces to
our corpus, so under a strict reading of D1 they should have been omitted or cited. They
survived for a reason worth writing down: **years are structurally exempt from the number
scan** — `isExcludedNumber` (claims.go) drops any standalone 1900–2099 token as a date — and
`ScanUnregisteredNumbers` is inert here anyway (English business-context gate). So the only
thing standing between us and an unsourced date is the prompt, and the prompt is
persuasion, not enforcement.

> **The honest characterisation, for the owner:** the fence removed the *dangerous* class
> (invented specs, prices, service intervals, mis-attributed models) and left a residue of
> *encyclopedic* claims that happen to be correct. That is a large improvement, not a
> guarantee. Anything numeric and historical still wants a human eye before it ships.

**Language defect, both pages:** the writer produced **"escapamento"**, which is not a
Spanish word (Spanish for the escapement is *escape*; *escapamento* is Portuguese). It
appears once per guide, so it is systematic rather than a slip. Worth a corrective line in
the site's content_direction or a banned_claims entry — it is a credibility tell on a
Spanish-language specialist site, and exactly the sort of thing no automated gate here
checks.

### Refinement after reading the REBUILT tourbillon page — the fence's exact shape

The rebuilt `glosario-tourbillon` (correct `site_plan_sections`, `generic-text-block`
instead of `content-block-about`) is genuinely about tourbillons, which closes the section
mistake. More interesting is what it reveals about the fence's precise behaviour.

It **cited, unprompted by anything except the writer_block's own phrasing convention**:

> "Richard Mille presentó el RM 64-01 Tourbillon Colnago, una edición limitada desarrollada
> junto al fabricante de bicicletas Colnago … **según informó TR Magazine**."

The `(reported by TR Magazine)` suffix I wrote into each `writer_line` came back out as real
in-copy attribution. That is a cheap and effective trick worth reusing: **phrase the
writer_line the way you want the sentence to read, attribution included.**

But the same page also asserts `menos de medio gramo` for a tourbillon cage — a specific
quantitative claim supported by nothing we hold — alongside the Breguet/1801 pattern seen in
`guia-complicaciones`.

> **The precise characterisation, and the one to carry forward:
> the fence makes the model CITE what it has; it does not stop it ADDING what it knows.**
> Sourced facts come through correctly attributed to the right model. Unsourced general
> knowledge still arrives, unmarked, sitting flush against the cited material — which is
> the harder review problem, because the cited sentences lend the uncited ones their
> credibility.

Practical consequence for anyone extending this: the residue is *encyclopedic* claims, not
*invented* ones, so the review question is "is this true?" rather than "did it make this
up?" — a much cheaper question, but not a free one. If it must be airtight, the missing
piece is enforcement, not prompting: `ScanUnregisteredNumbers` would have to be taught a
non-English/product-spec context gate (see the evidence_base header for why it is inert
here), or banned_claims extended with unit-shaped patterns (`\d+\s*gramos?`, `\d+\s*mm`) —
accepting that those will also block legitimate cited specs unless the fact list carries
them first.
