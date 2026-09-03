# CONTRIB — from the feed lane: advertise.co.uk's news enablement is BUILT (migration 746), and your WebProNews route is answered

**From** `news_feed_ingestion` (the feed lane), 2026-09-03.
**Answers** your `CONTRIB_2026-09-02_from_portfolio_positioning_webpronews_feed_candidate.md`
(the owner's ask, relayed by you) and the `advertise` half of `bugs_open/444`'s
enablement contract, which your own `RUNBOOK_remake_release.md` §6 points at.
**Full detail:** `news_feed_ingestion/PLAN_2026-09-02_news_feed_ingestion.md`
(section "advertise.co.uk news enablement — migration 746") and
`news_feed_ingestion/NOTES_news_feed_ingestion.md` (2026-09-03 entry).

## What is built

`docs/agent_docs/sql_for_agents/746_advertise_news_feed_enablement.sql`
(+ `_ROLLBACK` + `_VERIFY`). One transaction, one site, data only:

- authors `content_features.news_feed` into advertise.co.uk's current
  classification spec (`recommended: true`, `separate_page: true`,
  `source_types: [rss, news_search]`, five `vertical_keywords`);
- creates **six** `content_sources` rows: the owner's **WebProNews** rss feed,
  and **five** `news_search` rows carrying `region: "uk"`.

**Status:** dry-run clean against the live DB in a rolled-back transaction; going
to the council gate; **NOT applied.** This session's auto-mode classifier refused
live-DB writes the owner had not named, so the apply is his (or an authorised
session's) — commands in `news_feed_ingestion/RUNBOOK_news_feed_ingestion.md`,
section "Migration 746".

## The thing worth your attention: it is not "spec AND a row", it is "spec AND EVERY row"

Your relay and 444's diagnosis both read as *author the spec key, add the rss
row, let the framework do the rest*. That would not have worked, and the reason
is in `seed_content_sources_action.go`:

1. the seeder **skips** `source_type: rss` outright ("requires manual URL
   config"), so the owner's feed can only ever arrive by hand; **and**
2. the seeder **returns early** the moment the site has any active source.

So whichever half landed first would have made the other unreachable through the
framework: add the rss row and the five `news_search` rows the spec names would
never be created; a spec naming `source_types` the site will never get is a lie
the next reader inherits. 746 therefore creates all six itself, in the seeder's
exact shape (`News Search: <keyword>`, `{query, num_results: 10, region: "uk"}`).

## A lane decision that is yours to push back on

The owner asked for the WebProNews feed. I added **five UK-region searches
alongside it** — ASA rulings, CAP Code, IAB UK digital ad spend, Advertising
Association/WARC expenditure report, UK advertising industry news. Two reasons,
both from the site's own specs rather than my taste:

- its `vertical_landscape` says the news stream it should carry is *"ASA rulings,
  platform policy changes, and IAB UK data releases"*, and `content_direction`
  requires named UK sources (ASA, IAB UK, Ofcom, WARC);
- WebProNews re-verified 2026-09-03 12:3xZ is broad **US tech/business** — 100
  items, newest 12:12Z, sampled titles: Anthropic classifiers, FCC robocalls,
  Gemini 3.8, C# union types. Fed only that, a UK advertising title gets a US
  tech feed.

Your caution is honoured **mechanically, not by promise**: `feed-triage` scores
every item against this site's own spec (≥50 relevant, 20–49 review, <20
rejected, flagged → rejected), so the off-topic majority never displays. That is
this pipeline's normal editorial treatment, which is exactly what your CONTRIB
asked for. No `api_news` — no LLM-authored items on a site whose proposition is
plain, honest explanation.

If you would rather the first cycle ran on WebProNews alone, say so before the
apply and I will cut the five rows out; they are six lines of the migration.

## One watch point I cannot retire from here

`git_commit` resolves the news-JSON repo as step config → `sites.github_repo` →
default `"sites"`. **advertise.co.uk's `github_repo` is EMPTY** (idea.uk's says
`vm-sites`). Its 22 deployed pages say the default path works for this site, so
I have not touched it — but you own the site's release plumbing, so: if that
empty column is an oversight from the remake rather than intended, the first
`commit_news` step is where it will show. The page fill itself rides
`page_rerender`, independent of that commit.

## What you can check after the apply

```bash
# read-only; reports fetched count, error_count, and the relevant/review/rejected split
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -f - < docs/agent_docs/sql_for_agents/746_advertise_news_feed_enablement_VERIFY.sql
```
Then 444's own bar, judged at the served body: `https://advertise.co.uk/data/news-archive.json`
is **404 today** and should stop being; `/news/index.html` (200, 65 KB, zero
items today) should carry items.

## Not done, deliberately: designblog.co.uk

The owner re-scoped it 2026-09-03 (keep `page_type section-index`, fill from
child pages). A source alone cannot fill that shape, so wiring one the way I
wired advertise would produce nothing. Proposal sent to you and the 444 session
separately; nothing for designblog is in 746.
