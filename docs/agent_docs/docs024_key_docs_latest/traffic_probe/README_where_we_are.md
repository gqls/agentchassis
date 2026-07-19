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
The summary document is written and committed at docs/agent_docs/docs024_key_docs_latest/traffic_probe/SUMMARY_relojistas_rebuild.md — plain language, structured exactly as you asked (what we set out to achieve →
where we are now → what we aim to do), and written to be read aloud rather than skimmed.

The shape of it:

What we set out to achieve — the dead forum, the discouraging headline traffic number (400k requests, ~99% crawler noise, only 89 real homepage visits), and the two things buried in it: a feed still pulled
~136×/day by real subscription services all getting error pages, and genuine human watch searches from Spain, Chile and Mexico. The goal: rebuild as a Spanish watch news portal and reactivate that feed at its
original address.

Where we are now — the site is live and rebuilt; the news updates itself every six hours from five verified Spanish watch magazines plus an AI news search; and the headline result, the dormant feed is answering
again with 30 current stories at the exact address subscribers have always used. It also credits the platform fixes that came out of the work, and states honestly what isn't finished (three dead nav links, the
invented homepage links, the temporarily-absent search box).

What we aim to do — finish the navigation, give Guides and Glossary real Spanish content, bring the search box back improved so it answers rather than just records, then map the old per-board subscriptions to
matching topic feeds and measure the reactivation directly.

It closes with a one-sentence version you can lead with: a dead forum domain turned out to still have a live audience nobody was serving.

Meanwhile the Noticias build is moving — the page is now correctly typed as a news archive with its three sections, and the build item is queued for the builder to pick up. I'll carry on with that and report when
it's live.


---

Verified fixed, and already filed by another thread as bugs_open/018_..._coverage_report_hides_90pc_of_commits.md — so I won't duplicate it (that's the related-bug rule working as intended). The report now
correctly finds 29 in-scope commits, not 4.

Here's the honest summary of this re-read.

CLAUDE.md changed again in the minutes between my two reads — three deltas, all in the council section: the Council-Reviewed: trailer now accepts a fix-proposer RUN_ORCH_ID as well as a gate SUBMISSION_CORR; the
relevance-gated seat count moved from seven to "11 as of 2026-07-18, and growing"; and roster changes are now mechanical via 099_SYNC_gate_roster.py rather than hand-patching both councils. The file is under active
edit by the council thread while I work, which is precisely the concurrency hazard its own opening paragraph describes.

Acting on it surfaced a real bug, independently confirmed. I ran the coverage report, noticed it claimed 4 in-scope commits where git log showed 28, and traced it to kubectl exec -i draining the while read loop's
stdin — truncating the report at the first commit bearing a trailer. Another session had found the same thing ("4 of 41") and shipped the fix while I was mid-diagnosis. I checked before filing and found
bugs_open/018 already covers it, so I've added nothing.

Where my own work actually stands on coverage: my platform/ commit 9f515b0d7 (the render_rss_feed action) is genuinely unreviewed — it was never hidden by the bug, it simply predates my knowing the council existed.
It's now one of 28 unreviewed commits fleet-wide, so this isn't a me-specific gap; the practice is young and the whole branch is behind. I can't retrofit a trailer (forward-only, no amends), so the useful move is
to route the next platform change through the council — and bugs_open/015 gives me a natural candidate, since its fix candidates (teaching the planner the typed page vocabulary, or making MissingNewsPageCheck adopt
rather than create) are exactly the fleet-wide behaviour change the council is for.

Earlier in this same re-read cycle I also filed the two durable diagnoses I'd been missing: bugs_open/014 (the VM-site repo misroute, both causes, fixed) and bugs_open/015 (the mistyped page_type orphan, open
fleet-wide), with both transferable patterns added to 016b §9 and indexed.

Relojistas itself is untouched by all this and still where it was: Noticias re-typed to news-index with its sections, build item queued. Want me to carry on with that build, or first put the 015 planner fix through
the council now that I know how?


----

Both documents are written and committed, and memory is updated so a new chat picks this up cleanly.

Best news first: /noticias went live while we were in the council loop. The targeted fix worked exactly as diagnosed — re-typing section-index → news-index plus the three sections was enough, and render_news_section immediately
produced the archive it had been withholding. Verified live, not assumed: the page returns 200, carries the news-listing component, and news-archive.json renders 20 items. That was the most important dead link on the site.

HANDOFF_RESUME_relojistas_rebuild.md — the fresh-chat entry point. It carries the coordinates (site id, box, repos, DB access), a live-verified table of what's actually serving, the remaining task list, the platform work
delivered, and a traps section that will save the next session real time: nginx is setup.sh-managed and the box is drifted, so a re-run will silently delete the legacy-feed handler; page-rebuild is the wrong tool for redeploys
(use page-rerender); a full re-plan clobbers built pages; don't dispatch within ~300s of a pod restart; kcat dispatches can vanish silently.

SUMMARY_relojistas_rebuild.md — rewritten for reading aloud, with the arc you asked for: what we're doing → where we've come from → what we've done → where we are now → where we're going. I included the council episode, because
it's the most interesting thing that happened: the review didn't rubber-stamp our fix, it sent us looking for existing machinery, we found a component that already did the job better, and we withdrew the proposal. Two minutes
of review saved shipping a duplicate with a hidden flaw.

One thing I've deliberately left for you rather than filing myself: the council run produced no verdict artifact, because council_decide died on a reviewer's output being truncated at max_tokens — the same truncation family as
bugs_open 005/008/012. The reviews survived only inside collected_data. Every verbose submission will hit this, and the coverage report will score it as a mismatch rather than a review. The council thread owns that code and may
already know, so it's flagged in both the handoff and the notes for you to decide.

The remaining work is well-scoped: Guías/Glosario content (your author-starter-content decision), the phantom sweep for invented links (no code needed), search-that-answers, per-board category feeds, then measuring the 404→200
reactivation — which is the number that tells you whether this domain earns more investment.

✻ Churned for 3m 16s · 1 shell still running

※ recap: We're rebuilding relojistas.com as a Spanish watch news portal to reactivate its dormant RSS subscribers; that feed and the news archive are now live. Next: author starter Spanish content for the Guías and Glosario
sections, whose nav links still 404.

---

## 2026-07-19 (new session) — the reactivation number is in

I took the measurement you were left waiting for, and it's decisive. I did it first because it's
the number that says whether this domain earns any more investment, and it turned out to cost one
command.

Before the cutover, the old feed URL failed for everybody, every time. For as long as the logs go
back — 9 July onward — it was between fifty and ninety requests a day and a 404 for every single
one of them. After the cutover on the 17th it flips clean: the 18th was the first full day and it
served 122 successes against 3 failures. The 19th, to mid-morning, is 50 against 3. So the URL
those subscribers have had in their readers for years went from never working to working about
97% of the time, and it did it the same day we pointed it at real content.

Now the honest part, because the raw number flatters us. Most of that traffic is crawlers. Google,
Meta and Apple all re-discovered the URL the moment it started returning content instead of an
error, and between them they account for well over half the successful fetches. That is not
nothing — it means the feed is being indexed — but it is not a reactivated subscriber. Stripping
the known crawlers out leaves about 55 fetches from ordinary clients, and the most interesting of
those is a Java HTTP client polling on a schedule, which is exactly the signature of the
server-side feed reader we were hoping to wake up.

There's a limit on how much further I can take that, and it's worth knowing about because it
changes the priority of something we'd filed as tidying. Every request reaches us from a
Cloudflare address rather than the visitor's own, so I cannot count distinct subscribers at all.
When I say I see 86 distinct addresses, that means 86 Cloudflare edge nodes — it tells us nothing
about how many actual readers are behind them. Fixing that is the "CF real-ip" item that has been
sitting in housekeeping since P0, and it turns out to be load-bearing for measurement rather than
cosmetic. If you want a real subscriber count, that's the thing to do, and I'd promote it.

The three failures a day that remain are not mysterious. They're three specific variants we didn't
catch: the URL spelled with a lowercase type=rss2, the bare URL with no parameters at all, and the
old /ventas/ classifieds path. All three are cheap to add — but they live in the nginx config,
which is the drifted, setup.sh-managed file the handoff warns about, so I'd rather reconcile that
into the generator than pile a fourth surgical edit onto a file that a single re-run will wipe.

Carrying on now with the Guías and Glosario content, which is your author-starter-content decision
and the most visible thing still broken on the site.

---
