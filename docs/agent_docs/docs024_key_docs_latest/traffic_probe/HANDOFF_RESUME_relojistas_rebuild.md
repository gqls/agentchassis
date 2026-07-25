# HANDOFF — relojistas.com rebuild (fresh-chat entry point)

**Written 2026-07-19.** Start a new chat from this file. Everything below was verified
live at the time of writing, not assumed. Read `CLAUDE.md` at the repo root first — it is
edited frequently by other threads and its rules bind (pathspec commits, build-from-HEAD,
image→seeds, pod-verify, council gate, bugs_open filing).

Companions in this directory: `SUMMARY_relojistas_rebuild.md` (plain-language, for reading
out), `relojistas_rebuild_running_notes.md` (full dated log — the detail), `_plan.md`
(phases P0–P6), `_runbook.md` (commands), `_manifest.md` (spec),
`REPORT_vm_sites_repo_architecture.md`, `news_vertical_autodetect_design.md`.

---

## 1. What this is

relojistas.com — a dead Spanish watch forum — rebuilt as a **Spanish-language watch news
portal**, whose point is to **reactivate the forum's still-live RSS subscribers** at their
original feed URL. Traffic probe found ~99.7% bot noise but two real assets: a legacy feed
pulled **~136×/day** by real subscription services (all 404ing), and a trickle of genuine
human watch searches (ES/CL/MX).

## 2. Coordinates

| Thing | Value |
|---|---|
| site_id | `ecf15e75-a966-4900-bcb0-1c85f689dbfd` |
| Box | Hetzner CPX22 **167.233.33.159** (nbg1), CF-proxied, root SSH works |
| Webroot | `/var/www/vm-sites/relojistas.com` |
| Content repo | `gqls/vm-sites` (local checkout `~/projects/vm-sites`) → Action rsyncs to box |
| Engine repo | `gqls/site-engine` (local `~/projects/site-engine`), binary on `127.0.0.1:8080` |
| nginx | `/etc/nginx/sites-enabled/vm-sites.conf` (setup.sh-managed; see §6 caveat) |
| DB | `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db` |

## 3. LIVE and verified (2026-07-19)

| Surface | Status |
|---|---|
| `https://relojistas.com/` | 200 — Spanish news portal, homepage news card |
| `https://relojistas.com/noticias/` | 200 — `news-index`, `news-listing` component, **20-item archive** |
| `/historia.html`, `/sobre-nosotros.html`, `/contacto.html` | 200 |
| `/feed.xml` | 200 — valid RSS 2.0, 30 items, links out to sources |
| **`/external.php?type=RSS2`** | **200 — THE MISSION METRIC**: legacy subscriber URL reactivated |
| `/data/latest-news.json`, `/data/news-archive.json` | 200 (6 / 20 items) |
| Imagery + CSS | 200 (logo, 4 heroes, 4 icons, styles.css) |

**News runs itself:** 5 verified Spanish RSS magazines + Grok `api_news`, ingested/triaged/
rendered every 6h by `content-feed-trigger`; autonomous commits observed. Feed publishing
(`render_rss_feed`) fires in the same chain, per-site gated by `deploy_config.rss_feed`.

## 4. What remains (the live task list)

1. **Guías + Glosario content** (operator decision: *author starter content*). `guias-index`
   (section-index) and `glosario-index` (entity-directory) are `planned`, `sections=0`, and
   **their nav links 404**. They need child content authored — guide articles (there is a
   `articulo` blog-post template page) and glossary term entities (`glosario-entrada`
   entity-page template) — then the index pages compose and build.
2. **Invented CTA links** — `/ferias`, `/archivo`, `/guias/mantenimiento` 404 from the
   homepage `info-card-grid`. **No new code needed**: `check_phantom_internal_links` already
   covers exactly this class (invented links with no `pages` row) and is enabled in
   `completeness-discovery-agent` — it has simply never swept this site. Run that sweep.
3. **P5.2 — search-that-answers.** The rebuilt homepage has **no search box**, so intent
   capture has been paused since cutover. Plan: engine returns results matched against the
   curated news while still recording the intent event; then re-add the box.
4. ~~**Per-category legacy feeds (engine).**~~ **DEFERRED 2026-07-25 — surveyed, and
   the premise did not survive the survey.** Board-param feed requests are ~88%
   self-identified crawlers with **zero** conditional GETs; the one client that behaves
   like a real subscriber (42 × 304) polls the **bare** feed. The old boards people
   supposedly "subscribed to" were an unchecked sample of 8 from a real set of 123, and
   most are forum-social boards a news portal cannot fill (Seiko 1 matching item, Louis
   Erard 0, Sorteos 0). Today all `type=RSS2` variants correctly get the master feed and
   that is the right answer. Decision, reversal trigger and ready-to-build design:
   plan §P8; evidence: `EVIDENCE_2026-07-25_legacy_board_feed_demand.md`.
5. **Measure reactivation** — ~~count subscriber 404→200~~ **DONE 2026-07-25 for the
   part that is countable.** The legacy feed went 404→200 on **17 July** (0 × 200 before,
   200s dominant since; 25 Jul: 36 × 200, 0 × 404). Residual failures are ~5/day in
   exactly three shapes — bare `/external.php` (25), lowercase `?type=rss2` (11, the live
   conf's hand-edit is case-sensitive), `/ventas/external.php` (4) — **all three already
   fixed in the pending `setup.sh`**, so the owner's box run has a measurable before/after.
   Evidence of real subscribers: `FeedFetcher-Google`, `Apache-HttpClient`, empty-UA
   pollers, and one client doing 42 conditional GETs. **What is still NOT countable:
   distinct people** — every logged IP is a Cloudflare edge address until CF real-ip
   lands. That is now the measured blocker on the headline number, not a preference.
6. **Housekeeping:** stale relojistas copy still in `gqls/sites` + B2 (delete per
   `REPORT_vm_sites_repo_architecture.md` rec 4); `favicon.png` referenced but never
   generated; P0 CF real-ip `setup.sh` re-run still outstanding (setup.sh is no longer on
   the box — re-scp first); `intent_events` collector still disabled.

## 5. Platform work delivered here (benefits the whole fleet)

- **`resolveGitRepoNameDB`** (helpers.go) — deploy repo now resolves from the **site row**,
  not workflow state. Shipped v1.0.1126. Tests in `git_repo_resolution_test.go`.
- **Three `repo_name:"sites"` pins removed** from `page-rerender` / `site-deployer` /
  `deployer-agent` git_commit steps (DB config, live).
- **`render_rss_feed` action** — the platform could not publish outbound RSS before.
  Registered; gated by `deploy_config.rss_feed`; tests in `render_rss_feed_test.go`.
- **watch/horology vertical** added to `verticalNewsMap`.
- **Filed:** `bugs_open/014` (repo misroute, FIXED), `bugs_open/015` (mistyped `page_type`
  orphan, **open fleet-wide**), both patterns in `016b §9`.
- **Design:** `news_vertical_autodetect_design.md` (phased plan to make news-vertical
  detection + verified RSS sourcing automatic).

## 6. Traps — read before touching anything

- **nginx conf is setup.sh-managed and the box is DRIFTED** from every setup.sh in the repo
  (it predates `/events`). The P5.1 legacy-feed location was added **surgically** to the
  live conf (backup: `/root/vm-sites.conf.bak-p5-*`). **A setup.sh re-run will delete it** —
  reconcile into the generator first.
- **`page-rebuild` is the wrong tool for redeploys** — it regenerates content and dies at
  `resolve_links`. Use **`page-rerender`** (page_id-driven).
- **A full re-plan clobbers built pages** (`bugs_open/001`). To fix one page, edit that page
  targetedly — that is how `/noticias` was fixed (re-type + sections + re-queue).
- **Don't dispatch within ~300s of a chassis pod restart** — the spawn is silently dropped.
- **kcat dispatches can vanish** with no orchestration row. Verify a row appears; refire.
- **`rsync --delete`**: the vm-sites Action mirrors repo→box. Commit the box's current state
  before a first pipeline deploy or it deletes live files.
- **Verify against the live URL / DB row, never a `complete` status.**

## 7. Open question for the operator

The council gate produced **no verdict artifact** for submission `8a1f7b4f`: its
`council_decide` step died on `review_editquality.result` being *"invalid JSON — likely
truncated at max_tokens"* (the `bugs_open` 005/008/012 truncation family). The reviews
survived only inside `orchestration_states.collected_data`. Any submission with verbose
reviewers will hit this, and the coverage report will score it MISMATCH rather than
reviewed. **Not filed** — the council thread owns that code and may know. Decide whether to
file it.
