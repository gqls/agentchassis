# HANDOFF — news_feed_ingestion — continue here (b)

Written 2026-09-02, end of session, at the owner's request to switch
sessions after he deployed a fresh build himself. **Read this file first**,
then `PLAN_2026-09-02_news_feed_ingestion.md` (design + decisions, has a
STATUS block at the top of the UK-news section), `NOTES_news_feed_ingestion.md`
(full chronological account — this session's entries are at the very
bottom), `RUNBOOK_news_feed_ingestion.md` (commands), `README_where_we_are.md`
(plain-prose account for the owner). Supersedes
`HANDOFF_2026-09-02_continue_here.md` (still readable for candidate #1's and
316's fuller history, but everything actionable from it is folded in below).

## 0. What this lane is

Charter: the raw news-ingestion pipeline — source enrollment, fetch
scheduling/cadence/queue fairness, and normalising/extracting structure out
of what arrives. Concretely: `content-feed-orchestrator`,
`find_news_sites`/`content-feed-trigger`, `feed-triage`, `content-feed-refresh`,
`content_feed_items` as a data asset. Full charter + exclusions in PLAN's
opening section.

## 1. DONE — bugs_open/427 fix candidate #1 (LIVE, PROVEN — do not redo)

Extraction step turning a confirmed news item into a dated, cited
`evidence_base` fact. Shipped and tested end-to-end against boxingonline.com
earlier this session. Nothing left to do here. Full detail in the prior
HANDOFF and PLAN §"427 — design for fix candidate #1". Peer status
(candidates #2/#3) may have moved since — check `bugs_open/427` §9 before
assuming anything.

## 2. DONE — bugs_open/316: core already fixed, one rider refuted

Ordering + capacity already fixed and live (migrations 554, 556). The
"unenrolled sites never reach the seeder" rider was investigated and
refuted this session (full evidence in NOTES and in a RESPONSE section
appended to the CONTRIB file). Do not re-investigate.

## 3. IN PROGRESS — UK-news region default — RESUME HERE FIRST

Owner ask: UK news, not American, for `.co.uk`/`.uk` sites, as a default
with room to override. **Root cause found and fixed**: Firecrawl (the live
`PRIMARY_SEARCH_PROVIDER`) defaults its `country` geo-targeting param to
`"US"` when absent — verified directly against Firecrawl's and ScrapingBee's
own API docs, not assumed. Wired a new opt-in `region` config key end to
end: `content_sources.config` → `FetchNewsSearchAction` →
`WebSearchAction` → `web-search-adapter` → `providers.SearchOptions.Region`
(a struct field that already existed but was never populated) →
Firecrawl's `country` / ScrapingBee's `country_code`. `SeedContentSourcesAction`
now derives `region="uk"` for new `.uk`/`.co.uk` sites at seed time.
Migration `691_uk_news_search_region_default.sql` (+`_ROLLBACK`+`_VERIFY`)
backfills the same key onto the 26 already-provisioned rows across 6 sites
(measured 2026-09-02) that the seeder's `ON CONFLICT DO NOTHING` would
otherwise never reach. DuckDuckGo deliberately untouched — it already
declines `search_type=news` outright, so it never serves this pipeline
regardless of region (proven by an existing test, not assumed).

**Status, precisely:**
- Code + tests + migration: committed (`0a408f8db`), council-APPROVED round 1
  (`8842fe96-9a71-4ea5-9993-2483f10712cb`; `Council-Reviewed:` trailer on
  `2f8411b7e`). All 3 advisory objections were checked against the live
  system, not waved through — none needed a design change. Worth reading in
  NOTES: one check turned up a real, separate finding — `web_search_action.go`
  has no `ActionInputSpec` registered at all (predates this change), so
  the RFC_022 optional-key-budget audit lists it as **"NOT COUNTED —
  unknowable"**, not "under budget". Flagged for whoever next touches that
  file or the RFC_022 tooling; not this fix's job to build.
- Images built + pushed by this session at `v1.0.1353` (`agent-chassis`,
  `web-search-adapter`), from committed HEAD, tag confirmed unused first.
  IMAGE_TAG + both overlays bumped and committed (`886d693bc`).
- **This session asked the owner before rolling** rather than doing a
  single-service `kubectl apply -k` unilaterally — a standing memory
  ([[releases-are-whole-fleet-make-release]]) records the owner previously
  blocked exactly that and asked to run deploys himself. **The owner then
  deployed a fresh build directly**: both overlays now show
  `newTag: v1.0.1355` (higher than this session's `v1.0.1353` — HEAD had
  moved further by the time he built). Confirmed via `docker manifest
  inspect` (doesn't need kubectl) that `v1.0.1355` genuinely exists in the
  registry for both services — a real build, not a bare tag edit. The
  overlay files carrying `v1.0.1355` are **UNCOMMITTED** in the working
  tree as of this note (last commit on those files is still this lane's own
  `886d693bc` at `v1.0.1353`) — not this lane's file to commit on the
  owner's behalf without knowing his intended message; just noted.

**NOT YET VERIFIED — kubectl access is down fleet-wide** (confirmed as the
known 3-day token expiry — decoded the JWT `exp` claim directly, expired
2026-09-02 22:08Z; only the owner can refresh it, see
[[kubeconfig-token-expires-every-3-days]] in memory). **Do this first, in
this order, once kubectl works again:**

1. **Pod rollout status** — confirm both deployments actually rolled and
   pods are `Running`, not stuck:
   ```bash
   kubectl -n ai-persona-system get pods -l app=agent-chassis -o wide
   kubectl -n ai-persona-system get pods -l app=web-search-adapter -o wide
   kubectl -n ai-persona-system get deploy agent-chassis web-search-adapter \
     -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.spec.template.spec.containers[0].image}{"\n"}{end}'
   ```
2. **Binary verification, not the tag** (CLAUDE.md's own standing rule —
   never trust the roll status). Positive control: this lane's own commit.
   Negative control: a sha that should be absent (NOT "40 zeros" — that
   matches Go's binary padding and hides a stale build, LANDMINES.md).
   ```bash
   git log --oneline -3 -- platform/orchestration/actions/web_search_action.go
   # take the short sha of 0a408f8db (or whatever this file's HEAD commit now reads)
   kubectl -n ai-persona-system exec <agent-chassis-pod> -- grep -aq "<that-sha>" /proc/1/exe && echo PRESENT
   kubectl -n ai-persona-system exec <web-search-adapter-pod> -- grep -aq "<that-sha>" /proc/1/exe && echo PRESENT
   ```
   Also check the startup log line for provenance (may have scrolled out of
   range on a busy pod — the binary grep above is the one with no shelf
   life):
   ```bash
   kubectl -n ai-persona-system logs -l app=web-search-adapter --tail=300 | grep -m1 'build provenance'
   ```
3. **Migration 691's application status.** May already be applied (the
   owner's build/deploy step doesn't apply migrations, so this is
   independent of the image roll — check don't assume either way):
   ```bash
   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA -c "
     SELECT count(*) FROM content_sources cs JOIN sites s ON s.id=cs.site_id
     WHERE cs.source_type='news_search' AND lower(s.domain) LIKE '%.uk'
       AND NOT (cs.config ? 'region');"
   # 0 = already applied. 26 = still pending, re-check the count matches the
   # migration's own guard before applying (population may have moved).
   ```
   If still pending:
   ```bash
   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
     -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/691_uk_news_search_region_default.sql
   ./scripts/migration/run-migrations.sh --record-only 691_uk_news_search_region_default.sql \
     --note "applied by hand 2026-09-0X, verified via 691_..._VERIFY.sql"
   ```
   Then run the VERIFY sidecar:
   ```bash
   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
     -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/691_uk_news_search_region_default_VERIFY.sql
   ```
4. **Live verification against a real `.uk` site** — this fix's own PLAN
   calls for this and it hasn't happened yet. Dispatch a real
   `fetch_news_search` for `webdesign.co.uk`, `idea.uk`, or
   `farmerinsurance.uk` (NOT `mortgagecalculator.co.uk` — flagged elsewhere
   in the fleet's own comments as parked/404-on-the-wire). Confirm the
   adapter's logs show `region: "uk"` reaching it and, ideally, that
   returned result URLs/publishers skew UK (BBC/Sky/Reuters-UK-bureau etc.)
   rather than trusting a 200 status alone. `RUNBOOK` has the dispatch
   pattern used for candidate #1's own live test (`kafka-publish-lib.sh`,
   `agent_type=feed-triage`) — same shape applies here via
   `fetch_news_search`.

Only once all four pass should this be called done in PLAN/README.

## 4. NEW — three pending cross-session routes, all genuinely this lane's

Arrived today, none acted on yet, all fall in this lane's own priority-3
gap (source enablement for sites with zero `content_sources` rows):

1. **`portfolio_positioning`**: owner liked `webpronews.com`'s RSS feed
   (`https://www.webpronews.com/feed/`, verified 200/~1.08MB/100
   items/multiple-per-hour) as a candidate news source. Write-up already
   committed here as
   `CONTRIB_2026-09-02_from_portfolio_positioning_webpronews_feed_candidate.md`
   (commit `ebc050732`, not by this session — read it before acting).
   Caution stated in the CONTRIB: the owner endorsed the feed's *content*,
   not the old `advertise.co.uk` Drupal consumer's *wholesale-import
   pattern* — the classifier read that old consumer as "no original
   content", so whatever consumes this feed needs this pipeline's normal
   editorial treatment (triage/scoring/extraction), not a straight copy.
2. **`designblog.co.uk`**: `/the-design-feed/` serves zero items — 0
   `content_sources` rows for that site. Wants a DESIGN-vertical source
   (explicitly NOT WebProNews, which reads as marketing/business-adjacent).
   ACK sent this session. **RE-SCOPED (owner ruling 2026-09-03, relayed by
   `designblog.co.uk` after this handoff was first written) — read this
   before doing anything for this site.** The page KEEPS `page_type
   section-index` and is meant to fill via CHILD PAGES under the
   `/the-design-feed/` prefix, NOT a replan to a news-index page type. The
   `bugs_open/444` resolver confirms a section-index page resolves by child
   pages, not by `content_sources` directly — **so wiring a design-vertical
   source alone, the way this lane wired candidate #1/§3 above, will NOT
   fill this page.** The design-vertical source is still wanted, but as the
   INPUT that generates child pages/articles under the feed section (the
   editorial shape the page promises: foundry releases, studio rebrands,
   tooling notes) — not as a directly-bound feed. **How source→child-pages
   composes mechanically is explicitly left as this lane's own call, to be
   worked out WITH `portfolio_positioning` (owns the site's plan shape) and
   the `bugs_open/444` session (whose gate holds the page until children
   exist)** — do not build this one solo the way `advertise`'s plain
   news-index page can be. Coordinate before writing any code or spec for
   designblog.co.uk specifically; `advertise`/WebProNews (item 3 below) is
   the more direct, page_type-uncomplicated case and can proceed first.
3. **`bugs_open/444`** (empty listing pages): diagnosed that `advertise`'s
   empty `/news/` page needs BOTH: `content_features.news_feed` authored in
   its classification spec (no key exists there at all today —
   `idea.uk`'s 2026-08-25 spec entry is the worked example: `recommended:
   true`, `source_types`, `vertical_keywords`) AND a `content_sources` row
   (`source_type='rss'`, the WebProNews URL, with the editorial caution
   above). **UPDATE (received after this handoff was first written, via
   `Portfolio positioning`): the plan-time gate is now council-APPROVED
   (round 3, 2026-09-02 20:53Z) and migration 720 is applied — the
   planner's prompt half is LIVE, not pending.** Their `capability_gap`
   named `news_source_enablement` will hold un-enabled news pages until
   this lane's enablement work lands. Confirmed live and empty right now:
   `https://advertise.co.uk/news/index.html` serves 200 at ~61 KB, zero
   items — a directory-of-prose until a source exists; nothing on their
   side blocks starting once cluster access is back. Recipe: their own
   `RUNBOOK_remake_release.md` §6. Their plan:
   `bugfix_444_empty_listing_pages/PLAN_2026-09-02_listing_source_gate.md`.

**Suggested order once §3 is verified**: read `idea.uk`'s 2026-08-25
classification-spec entry as the template. **`advertise` first** — a plain
`news-index`-shaped page, WebProNews as the source, author the spec entry +
seed the `content_sources` row, confirm `/news/index.html` fills on the
next `feed-triage` cycle. **`designblog.co.uk` NOT solo** — per item 2's
re-scope, message `portfolio_positioning` and the `bugs_open/444` session
to agree the source→child-pages mechanism before writing anything for that
site; do not just reuse the `advertise` recipe there, it's a different
page_type with a different fill mechanism. Grep `bugs_open/444` in full
before starting either — it may have moved since this note.

## 5. Standing practices this lane has been following (keep doing these)

- Verify a peer's or a CONTRIB's claim against the live system before
  acting on it.
- Test end-to-end against real data before declaring anything done — not
  yet possible for §3's build; that's exactly why it isn't marked done.
- Pathspec every commit, explicitly; `git status --short` before any commit
  with no explicit file list.
- Never reuse a working tree's uncommitted `IMAGE_TAG`/kustomize tag bump —
  check `docker manifest inspect` for whether it's already pushed/live
  first.
- Migrations: apply your OWN migration by hand + `run-migrations.sh
  --record-only`, never the bulk `--apply` on a shared queue.
- Verify a deploy at the binary (`/proc/1/exe` sha grep, positive AND
  negative control), never at the roll status alone.
- **New this session: a single-service cluster deploy is the owner's call
  unless he's said so for THIS specific task** — build and push freely
  (an unused pushed tag is harmless), but ask before `kubectl apply
  -k`/rollout on a change he hasn't already authorised, even if an earlier
  task in the same session got a "please go ahead". One authorisation does
  not carry to the next change.
- **New this session: a fleet-wide `Unauthorized` from every `kubectl` call
  at once is the known 3-day token expiry, not a bug** — decode the JWT
  `exp` claim directly before assuming anything else is wrong
  ([[kubeconfig-token-expires-every-3-days]]); only the owner can refresh
  it, and `docker manifest inspect` still works without it for checking
  whether an image was actually pushed.
