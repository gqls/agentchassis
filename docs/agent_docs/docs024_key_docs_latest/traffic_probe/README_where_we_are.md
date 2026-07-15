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

