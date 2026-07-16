● Done. Here's the write-up in full, matching this directory's existing plan/runbook/running-notes conventions.

What I wrote (4 new docs in docs024_key_docs_latest/traffic_probe/)

┌─────────────────────────────────────┬──────────────────────────────────────────────────────────────────────────────────────────────────────┐
│                File                 │                                               Purpose                                                │
├─────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ relojistas_rebuild_manifest.md      │ The full spec — the Thread-A deliverable (identity, IA, content, RSS-out, the clever-engine options) │
├─────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ relojistas_rebuild_plan.md          │ Phased plan P0–P6 with entry/exit criteria, a decisions table, and sequencing                        │
├─────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ relojistas_rebuild_runbook.md       │ Exact operator commands; §0 (traffic read) is proven, P0–P6 are templated for build time             │
├─────────────────────────────────────┼──────────────────────────────────────────────────────────────────────────────────────────────────────┤
│ relojistas_rebuild_running_notes.md │ Dated log of this session's read, decisions, and next steps                                          │
└─────────────────────────────────────┴──────────────────────────────────────────────────────────────────────────────────────────────────────┘

I also added a memory entry so the workstream carries across sessions.

The read, and the decisions captured in them

Traffic (9–14 Jul, live off the box): 405,701 requests but ~99.7% bot noise — 83.5% are 404s on dead forum paths, only 89 hits reached the real homepage. The two real assets: the legacy feed
/external.php?type=RSS2 is still pulled ~136×/day (749 in the window, 31 sources including FeedFetcher-Google/Applebot) and every one gets a 404; and the search box has logged ~8 genuine organic watch
searches from ES/CL/MX over 3.5 weeks — real demand that the June "human ≈ 0" verdict predates.

Your choices, recorded: host on the existing CF-proxied box (object storage can't answer a query-string feed URL); keep the engine and pair a clever backend feature with the static pages; manifest-first
before building.

The one net-new build item: the chassis already ingests, triages and curates news, and VM deploy has landed — but it emits JSON, not RSS. So the only new code is a thin render_rss_feed step plus a
legacy-URL handler. For the "clever engine" you asked me to think up, my recommendation is in the manifest (§6): (A) the engine serves the legacy query-string feeds dynamically, mapping old forumids/cat
to our topic feeds so each returning subscriber gets the closest modern feed — the one thing static hosting genuinely can't do — and (B) the search box now answers (from curated news + link-outs) while
still capturing intent.

Three things I flagged for you to confirm on the manifest before I start P0: the news-portal framing, which clever-engine option(s) to build (I recommend A+B), and that we verify real RSS source URLs
before seeding (leaning on Grok api_news as the fabrication-safe primary). Say the word and I'll begin at P0 (the Cloudflare real-ip re-run).


---

The P1/P2 artifacts are drafted and grounded in the real schema. Here's what I produced and the one design finding that shaped it.

New artifact: relojistas_rebuild_seed.sql

I read the actual content_sources DDL, the site_specs versioning, and the two feed actions before writing a line, so this matches the live schema rather than guessing. Three sections:

- P1a — onboarding UPDATE: sets github_repo='vm-sites', deploy_config.target='vm', capabilities=['backend'], and engine.{base_url,stats_key} (the documented shape from the P4 migration). The stats key is
  left as <INTERNAL_API_KEY> — read from the box env at apply time, not committed.
- P1b — relojistas_set_news_feed(): a versioned site_specs merge that sets content_features.news_feed.recommended=true with source_types=[rss, api_news, news_search] and Spanish vertical_keywords. It
  deep-merges so it's safe whether or not the build pipeline has already written a classification spec.
- P2 — seed_relojistas_sources(): inserts the 5 verified Spanish RSS feeds plus the Spanish Grok api_news row, then marks them due immediately.

The finding that changed the approach

The auto-seeder (SeedContentSourcesAction) deliberately skips rss and scrape — "requires manual URL config." It only creates news_search rows (one per keyword) and one generic api_news named LLM News:
<domain>. Two consequences I handled:

1. The 5 verified RSS feeds would never be added automatically → they're inserted explicitly.
2. To stop a generic English Grok source from being what runs, I pre-insert our Grok row under that exact canonical name (LLM News: relojistas.com) with a Spanish prompt — so the auto-seeder's ON CONFLICT
   (site_id,name) DO NOTHING no-ops and our Spanish config wins.

Gemini is included as a commented-out second api_news row, flagged blocked until the ingester's provider routing supports it (today it routes xai/openai/perplexity only) — so we don't seed a source that
silently errors.

Two things now waiting on you

Everything above was safe to prepare without touching production. Before the seed SQL can be applied, two operator steps remain:
1. P0 — the Cloudflare real-ip setup.sh re-run on the box (a production change with box-takeover semantics; the runbook has the exact command).
2. Confirm the sites row exists for relojistas.com — the seed presumes one. The SQL opens with the SELECT ... FROM sites WHERE domain='relojistas.com' check; if it returns nothing, the row needs creating
   via the normal site-creation path first.

Next artifact in the queue is P4's render_rss_feed Go action (the one genuinely new piece) — I can draft that against render_news_section_action.go whenever you want, or hold until P0/P1/P2 are applied
and ingesting. Which would you prefer?

