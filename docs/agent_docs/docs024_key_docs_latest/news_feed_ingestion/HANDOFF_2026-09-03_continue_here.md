# HANDOFF — news_feed_ingestion — continue here (2026-09-03)

Written 2026-09-03. **Read this first**, then `PLAN_2026-09-02_news_feed_ingestion.md`
(design + decisions; STATUS blocks at the top of the UK-news and advertise
sections), `NOTES_news_feed_ingestion.md` (chronological, newest at the bottom),
`RUNBOOK_news_feed_ingestion.md` (every command below, in full),
`README_where_we_are.md` (plain prose for the owner).

**Supersedes `HANDOFF_2026-09-02b_continue_here.md`** — still readable for the
fuller history of candidate #1 and 316, but everything actionable from it is
folded in here, and two of its statements are now stale (the fleet is on
`v1.0.1358`, not the `v1.0.1355` it names; the overlay/`IMAGE_TAG` bumps it
records as uncommitted are committed).

## 0. What this lane is

Charter: the raw news-ingestion pipeline — source enrollment, fetch
scheduling/cadence/queue fairness, and normalising/extracting structure out of
what arrives. Concretely `content-feed-orchestrator`,
`find_news_sites`/`content-feed-trigger`, `feed-triage`, `content-feed-refresh`,
and `content_feed_items` as a data asset. Full charter and exclusions in PLAN's
opening section.

## 1. THE ONLY THING BLOCKING THIS LANE: two migrations need an authorised apply

Both are written, guarded, dry-run clean, and **not applied**. This session's
auto-mode permission classifier refused writes to the live `clients_db` that the
user had not named in his own message — correctly, and it was not worked around.
Anyone resuming with the owner's say-so (or the owner himself) can clear this in
about five minutes. **Exact commands: `RUNBOOK_news_feed_ingestion.md`, sections
"Migration 691" and "Migration 746".**

| file | what it does | pre-check today |
|---|---|---|
| `691_uk_news_search_region_default.sql` | stamps `region:'uk'` onto the 26 existing `.uk` `news_search` sources (6 sites) | **26** pending = exactly the guard's expectation |
| `746_advertise_news_feed_enablement.sql` | enables news for advertise.co.uk: spec flag + 6 `content_sources` rows | site has 0 sources, spec has no `content_features` |

⚠ **The number 691 is now shared** with another lane's
`691_per_site_palettes_for_three_sites_on_a_shared_library_row.sql` (applied
2026-09-02 21:26Z). The ledger keys on `filename` so nothing collides, but
**refer to this one by slug**, and record it by its full filename.

## 2. DONE and PROVEN — the UK-news region fix is live in the running binaries

The owner's ask ("the news is from America; I'd like UK news for all .co.uk and
.uk sites, perhaps as a flag with a UK default") is fixed in code and the code is
running. Verified this session, not assumed:

- fleet on `v1.0.1358` for both `agent-chassis` and `web-search-adapter`; pods
  `Running`;
- the adapter's own startup line gives `git_commit d0252fd4dab2a3a583d1cc8eb8e1b26e9c422d85`,
  and `git merge-base --is-ancestor 0a408f8db d0252fd4d` → **YES**, so this lane's
  commit is in the build;
- the stamp sha is **PRESENT** in `/proc/1/exe` on both chassis pods and the
  adapter, with a post-build sha (HEAD, 147 commits later) **ABSENT** as the
  negative control.

Mechanism, for anyone picking this up cold: Firecrawl (the live
`PRIMARY_SEARCH_PROVIDER`) defaults its `country` param to `"US"` when absent.
A new opt-in `region` key now flows `content_sources.config` →
`FetchNewsSearchAction` → `WebSearchAction` → `web-search-adapter` →
`providers.SearchOptions.Region` → Firecrawl `country` / ScrapingBee
`country_code`. `SeedContentSourcesAction.regionForDomain()` sets `region:"uk"`
for any `.uk` suffix at seed time. Council-APPROVED round 1
(`8842fe96-9a71-4ea5-9993-2483f10712cb`, trailer on `2f8411b7e`). DuckDuckGo is
untouched on purpose — it declines `search_type=news` outright, proven by an
existing test.

**What is NOT yet proven, and cannot be until 691 lands:** that a real `.uk`
search comes back British. The 26 existing sources carry no `region` key, so a
dispatch today exercises the absent-key path and proves nothing about the fix.

## 3. RESUME HERE — step 4, the live `.uk` dispatch (after 691)

1. Dispatch: `idea_uk_vm_site/scripts/dispatch_content_feed_orchestrator.sh`
   as-is (idea.uk, receipt + landing check built in). Its header's precondition
   holds: no chassis pod restarted in the last ~300 s.
2. Read the fix **at the adapter**, which logs `region` on every search:
   ```bash
   kubectl -n ai-persona-system logs -l app=web-search-adapter --since=30m \
     | grep -E '"msg":"Executing search"' | grep -oE '"(query|region|provider)":"[^"]*"' | paste - - -
   ```
3. ⚠ **Do not judge "results skew UK" on `content_feed_items.source_url` hosts.**
   `[MEASURED 2026-09-03]` 41 of 73 idea.uk `news_search` URLs are
   `www.google.com` — Google News redirects, so a host census measures the
   redirect, not the publisher. Judge on `source_title`/`source_summary`
   publisher names, or on the adapter's `region` field above.

Baseline for the comparison is in NOTES (per-source fetch times, item counts,
full host mix, all `[MEASURED 2026-09-03]`).

Only when step 4 passes should the UK-news work be called done in PLAN/README.

## 4. DONE (pending apply) — advertise.co.uk, migration 746

Answers **both** the owner's WebProNews ask and the `advertise` half of
`bugs_open/444`. Built, dry-run clean, council submission written
(`COUNCIL_SUBMISSION_746.json` in this dir).

One transaction: authors `content_features.news_feed` into the current
classification spec (`recommended`, `separate_page`, `source_types:[rss,
news_search]`, 5 keywords) **and** creates six `content_sources` rows — the
owner's WebProNews rss feed plus five `news_search` rows with `region:"uk"`
anchored on ASA rulings / CAP Code / IAB UK ad spend / AA-WARC expenditure /
UK advertising industry news, all taken from the site's own `vertical_landscape`
and `content_direction` specs.

**The finding worth carrying to any future feed enablement** (now a LANDMINES
entry): it is not "author the spec key **and/or** seed a source", as
`bugs_open/444` and the remake runbook both say. `seed_content_sources_action.go`
**skips `rss` outright** and **returns early on any existing active source**, so
the two halves block each other — whichever lands first makes the other
unreachable, and the site reads as enabled while producing nothing. Author the
key AND create every row it names, in one transaction, by hand.

Post-apply verification, in order: `_VERIFY.sql` (reports fetched count,
`error_count`, and the relevant/review/rejected split) → wait for
`content-feed-refresh` (6-hourly; the trigger selects the site because every new
source has `next_fetch_at` NULL) or dispatch the orchestrator directly →
`_VERIFY.sql` again → `https://advertise.co.uk/data/news-archive.json` stops
404-ing → served `/news/index.html` item count above zero (444's own bar: judge
at the served body, never at page status or byte size).

**Watch point I could not retire:** `git_commit` resolves the news-JSON repo as
step config → `sites.github_repo` → default `"sites"`, and advertise's
`github_repo` is **empty** (idea.uk's says `vm-sites`). Its 22 deployed pages say
the default works; flagged to `portfolio_positioning`, who own that plumbing.

## 5. NEXT, and it is a DECISION not a build — designblog.co.uk `/the-design-feed/`

**Do not wire this one solo, and do not reuse the advertise recipe.** Owner
ruling 2026-09-03: the page keeps `page_type section-index` and fills from
**child pages** under the prefix. A bound source cannot fill that shape, so the
missing link is a producer that writes child pages from feed items.

`[MEASURED 2026-09-03, first-hand]` **no Go code writes
`content_feed_items.published_page_id`** (repo-wide non-test grep, zero hits);
the column is set on **15 of 14,194** rows. So no live feed-item→page producer
exists. The one candidate is `create_blog_posts` via `blog-content-planner` —
real, wired, and **dormant**: `llm_call_log` shows **10 calls all-history,
2026-04-03 → 2026-04-24**, none since, cause unestablished. Its workflow plans
posts from an LLM prompt over a site spec, so even revived it would need
`content_feed_items` wired in as an input.

Two routes, written up for the three lanes that must agree:
`designblog_couk/CONTRIB_2026-09-03_from_feed_lane_the_design_feed_needs_a_child_page_producer_not_just_a_source.md`
— (1) revive `blog-content-planner` and feed it triaged items, or (2) build a
feed-item→article producer here. This lane's preference is (1) first, (2) as
fallback; route (2) is a new shared mechanism, so it is architecture-scope and
owes a concept-register entry in the same commit. The design-vertical source set
is ready to author (explicitly **not** WebProNews) the moment the mechanism is
settled.

## 6. Earlier work — do not redo

- **`bugs_open/427` fix candidate #1** — extraction step turning a confirmed news
  item into a dated, cited `evidence_base` fact. LIVE and PROVEN end-to-end
  against boxingonline.com. Peer status on candidates #2/#3 may have moved —
  check `bugs_open/427` §9 before assuming anything.
- **`bugs_open/316`** — ordering + capacity fixed and live (migrations 554, 556).
  The "unenrolled sites never reach the seeder" rider was investigated and
  **refuted**; evidence in NOTES and in a RESPONSE section on the CONTRIB file.
  Do not re-investigate.

## 7. Standing practices this lane follows

- Verify a peer's or a CONTRIB's claim against the live system before acting on
  it. Two peer statements were corrected this way today.
- Test end-to-end against real data before declaring anything done.
- `\d <table>` before writing SQL. This session guessed `site_specs.version` and
  `scheduled_tasks.is_active`; neither column exists, and it cost a batch.
- Pathspec every commit explicitly; `git status --short` first.
- Apply your OWN migration by hand plus `run-migrations.sh --record-only`; never
  the bulk `--apply` on a shared queue.
- **Dry-run a migration by hand** (`sed 's/^COMMIT;$/ROLLBACK;/' | psql`). The
  runner's probe silently declines any file whose text contains the word
  "rollback" — including a header naming its own sidecar — and prints "not
  probed", which reads as a property of your SQL. The hand run caught `min(uuid)`
  in 746's guard, which would have failed at apply. LANDMINES entry added.
- Verify a deploy at the **binary**, with positive and negative controls, never
  at the roll status. Grep the exact JSON key `"msg":"build provenance"` — a
  loose `grep 'build provenance'` on `agent-chassis` matches a 5 MB debug
  payload. LANDMINES entry added.
- **A single-service cluster deploy is the owner's call** unless he has said so
  for that specific task; one authorisation does not carry to the next change.
  Build and push freely — an unused pushed tag is harmless.
- **A live-DB write the owner has not named may be refused by the session's own
  permission layer.** That is the harness working. Write the file, guard it,
  dry-run it, put the exact commands in the RUNBOOK, and hand it over — do not
  look for a way around it.
