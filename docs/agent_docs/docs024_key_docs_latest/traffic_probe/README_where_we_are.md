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

---

Report complete — full version in traffic_probe/REPORT_vm_sites_repo_architecture.md. Here's the substance.

The answer: keep vm-sites as a separate repo — the decision is already settled and sound. What's broken is how commits find it.

What the previous discussions decided

- June (probe go-live): gqls/vm-sites was created by hand, private, deliberately — the git-adapter auto-creates repos public. Same root layout as sites so the chassis could target it unchanged. "github_repo selects
  the target" was specified then but never wired.
- idea.uk workstream (14–16 Jul — the deepest treatment, with your constraint on record: thousands of domains ⇒ per-site repos ungainly): a three-class fleet model was agreed. Class A static→B2 for the thousands;
  Class B static→VM+backend for the handful that sell something; Class C (dynamic rendering on the box) rejected — the VM is a second sink, not a second renderer. Delivery converges on pull: each box sparse-checkouts
  its own folder from vm-sites with a read-only key (the push Action was rejected for fleet use — one runner SSH key authorized on every box). The "four dead wires" for per-site targeting shipped in v1.0.1123.
- Your premise is exactly the settled doctrine: static→dynamic is primarily a deploy-stage difference — same artefact, different sink, plus backend components opted in at plan level.

Why the separate repo is right

1. Deploy keys are repo-scoped; GitHub has no path-scoped keys. Under pull, if Class-B sites lived inside gqls/sites, every box's key could read the entire portfolio — thousands of domains including unlaunched
   staging. vm-sites caps a compromised box's reach at the Class-B handful. This can't be engineered away inside one repo.
2. Sink separation by construction. The sites Action B2-syncs everything it sees; VM sites inside it would forever double-deploy to B2 (the "elaborate staging copy" pathology idea.uk suffered) or force exclusion
   lists into the fleet Action.
3. The repo flip is the migration primitive. Class A→B = flip one column — as close to "only the deploy stage differs" as it gets.

The live finding that sharpens it (worse than yesterday's feed bug)

I verified against today's running cluster (v1.0.1125): relojistas' page deploys also committed to gqls/sites — not just the news JSON. page-rerender and build-dispatch-loop orchestrations carry no site_record at
all; only planner-tier workflows run ensure_site_record. So despite correct data and shipped wiring, every relojistas artefact is sitting invisible in the B2 repo and the box still serves the old probe page. The
shipped wiring resolves the repo from workflow state, when it should resolve from the site row.

Recommendations (in the report, priority order)

1. Workflow-independent repo resolution — git_commit/deploy_image_asset fall back to SELECT github_repo FROM sites WHERE domain=$1 before defaulting to "sites". One query; kills the whole misroute class for all ~40
   workflows. (Interim data-only fix for news: add a load-site step to the feed orchestrator's workflow JSON — immediate, no image rebuild.)
2. Apply the vm-sites Action allowlist now (deploy-targets.json: {"relojistas.com": "167.233.33.159"}) — today that Action rsyncs every changed folder to relojistas' box; the moment idea.uk lands in vm-sites it
   deploys to the wrong machine.
3. One source of truth: deploy_config.target='vm' is semantic; github_repo is the consequence — derive or consistency-check it.
4. Script the A→B migration (flip, re-render, allowlist/pull provision, and — critically — delete the stale folder from sites/B2).
5. Converge push→pull (relojistas migrates when convenient), and keep Class B the exception.

Want me to implement recommendation 1 (the Go fallback + the interim feed-workflow fix) so relojistas' pages and news can land on the box?


---
The cutover is complete — relojistas.com is live as the Spanish watch-news portal. https://relojistas.com/ now serves "Relojistas — Portal de noticias de relojería en español" with the news card pulling six curated
Spanish stories from /data/latest-news.json, and historia, sobre-nosotros, and contacto all return 200. Better still, the 6-hourly heartbeat has adopted the site on its own — three autonomous "Update latest news
feed" commits landed in vm-sites after ours, so news now refreshes without any manual triggering.

The misroute turned out to be a three-layer onion, and all three layers are now fixed:

1. Workflow-state dependence — fixed by resolveGitRepoNameDB in v1.0.1126 (your ship), which falls back to sites.github_repo by domain.
2. Hardcoded pins — the reason round-1 rerenders still went to sites on the new image: three agent definitions (page-rerender, site-deployer, deployer-agent) had repo_name: "sites" baked into their git_commit step
   config, and explicit config rightly outranks the fallback. I removed the three pins (data-only jsonb edit, immediate; the diagnosis agents pinning agentchassis were left alone). Round 2: all four pages committed to
   repo=vm-sites and the Action deployed within a minute.
3. The rsync --delete trap — disarmed before any of this could bite, by mirroring the box's live webroot into the repo first (checksums verified; the box page was never disturbed until its replacement arrived).

Two operational gotchas worth keeping (both in the notes): the first rebuild dispatch vanished silently because it was fired too close to the chassis rollout (the ~300s warning is real), and page-rebuild was the
wrong tool anyway — it regenerates content via LLM and failed at resolve_links; page-rerender is the correct redeploy path.

What remains, in priority order: P4 render_rss_feed → /feed.xml; P5 the engine's /external.php legacy handler — flipping the ~136 daily subscriber 404s to 200s is the mission metric — plus search-that-answers
(intent capture is paused since cutover, as the new homepage has no form yet); composing /noticias/index.html (still planned); the scaffold pages and contacto details; and cleaning the stale relojistas copy out of
gqls/sites/B2.

Shall I proceed with P4 — building the render_rss_feed action?

--
