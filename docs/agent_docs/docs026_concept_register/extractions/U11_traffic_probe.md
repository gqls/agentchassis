# EXTRACTION U11 — docs/agent_docs/docs024_key_docs_latest/traffic_probe/
Extracted 2026-07-13. Files in scope: 558. Concepts found: 39.

Chronology note: the project runs 2026-06-10 → 2026-06-13 in four living docs
(plan / runbook / running-notes / per-domain notes, all version families).
Highest-numbered variants are authoritative; earlier variants were diffed
(`family-delta`) and contain only superseded designs already narrated in the
running notes (store v1, suitable_site_types gate, per-row scheduler fan-out,
`sites/**` layout, ON CONFLICT (type)). ~460 of the 558 files are archive.org
site captures and generated context bundles — accounted for below as grouped
rows with exact counts (individually enumerable via the find command in the
unit brief); no concept content lives in them beyond what the notes already
cite (the Wayback snapshots are provenance evidence for the two domains).

## Coverage

Root documents and SQL:

| file | treatment |
|---|---|
| HANDOFF_vm_sites_permanent_thread.md | full |
| README_stats_internal_key.md | full |
| TASK_traffic_probe_brief.md | full |
| intent_collector_agents(2).sql | family-latest |
| intent_collector_agents.sql | family-delta (only diff: ON CONFLICT (type) → (type,version)) |
| intent_collector_registration.sql | full |
| intent_events_migration(1).sql | family-latest (base in deploy_setup/working_dir/) |
| intent_probe_component(1).sql | family-latest |
| intent_probe_component.sql | family-delta (superseded suitable_site_types gate) |
| intent_ranking_queries(1).sql | family-latest |
| intent_ranking_queries.sql | family-delta (pre-intent_site_stats join) |
| intent_site_stats_migration.sql | full |
| package_traffic_probe.sh | header-scan |
| passive_harvest_spec(2).md | family-latest |
| passive_harvest_spec(1).md | family-delta |
| passive_harvest_spec.md | family-delta |
| relojistas_golive(3).md | family-latest |
| relojistas_golive(2).md | family-delta |
| relojistas_golive(1).md | family-delta |
| relojistas_golive.md | family-delta (open layout item, since resolved) |
| relojistas_notes(8).md | family-latest |
| relojistas_notes.md, (1)–(7) | family-delta (8 files; store-v1 jq commands superseded) |
| traffic_probe_domains.tsv | header-scan (388-line marketplace export; header comments read) |
| traffic_probe_plan(12).md | family-latest |
| traffic_probe_plan(7)–(10) | family-delta (4 files; no dropped concepts) |
| traffic_probe_runbook(13).md | family-latest |
| traffic_probe_runbook(8)–(11) | family-delta (4 files; no dropped concepts) |
| traffic_probe_running_notes(28).md | family-latest |
| traffic_probe_running_notes(23)–(26) | family-delta (4 files; verified strict subsets of (28)) |
| wayfaringlondoner_notes.md | full |

Hand-made probe pages (HTML instances of the intent-probe pattern):

| file | treatment |
|---|---|
| relojistas_golive/index.html | header-scan |
| relojistas_golive/gracias.html | header-scan |
| wayfaringlondoner/index(2).html | family-delta |
| wayfaringlondoner/index(3).html | header-scan (family-latest) |
| wayfaringlondoner/thanks.html | header-scan |
| deploy_setup/relojistas-site/index.html | header-scan (duplicate of relojistas_golive/) |
| deploy_setup/relojistas-site/gracias.html | header-scan |

deploy_setup (engine source, provisioning, workflows — stage-2 code; headers read):

| file | treatment |
|---|---|
| deploy_setup/site-engine/go.mod | header-scan |
| deploy_setup/site-engine/main.go | header-scan |
| deploy_setup/site-engine/service.go | header-scan |
| deploy_setup/site-engine/store.go | header-scan |
| deploy_setup/site-engine/site-engine.env | full (config comments) |
| deploy_setup/vm-deploy/setup.sh | header-scan |
| deploy_setup/vm-deploy/deploy-to-vm.yml | header-scan |
| deploy_setup/vm-deploy/deploy-engine-to-vm.yml | header-scan |
| deploy_setup/working_dir/README_setup.md | full (one line) |
| deploy_setup/working_dir/README_setup.sh | full (two lines) |
| deploy_setup/working_dir/accessdigest.go | header-scan |
| deploy_setup/working_dir/env.go | header-scan |
| deploy_setup/working_dir/intent_events_migration.sql | family-delta (superseded per-row fan-out pre_query) |
| deploy_setup/working_dir/main(13).go, main(20).go, main(11).go.orig11 | family-delta (3 files) |
| deploy_setup/working_dir/service.go, service(11)(13)(14)(25).go, service(20).go.orig20, service(23).go.orig23_wayfaringondoner | family-delta / header-scan of latest (7 files) |
| deploy_setup/working_dir/store.go, store.go.orig1, store(1)(2)(4).go.orig*, store(5).go.orig5_wayfaringlondoner | family-delta / header-scan of latest (6 files) |
| deploy_setup/working_dir/setup(3)(5)(6)(7)(8)(9)(10)(11)(12)(13)(14)(16)(17)(18)(19).sh | family-delta (15 files; latest matches vm-deploy/setup.sh) |
| deploy_setup/working_dir/deploy-to-vm(0)–(5).yml | family-delta (6 files) |
| deploy_setup/working_dir/deploy-engine-to-vm.yml, (1)–(3) | family-delta (4 files) |
| deploy_setup/working_dir/go(1).mod | header-scan |
| deploy_setup/working_dir/probe.env(0).example, probe.env(2).example | family-delta (pre-rename names) |
| deploy_setup/working_dir/site-engine.env, site-engine.env(1)–(4).example, site-engine.env.example.orig1 | family-delta (6 files) |

Site captures and generated context bundles (all listed per unit brief without reading):

| file | treatment |
|---|---|
| archive.org.results/relojistas/Relojistas - Foro de relojes.html | skipped-generated (Wayback capture; provenance evidence) |
| archive.org.results/relojistas/Relojistas - Foro de relojes_files/** (105 files: gif/png/jpg/js/css/txt assets) | skipped-binary / skipped-generated |
| archive.org.results/wayfaringlondoner/** (111 files: 2 Wayback capture pages + asset dirs) | skipped-binary / skipped-generated |
| docubundle/output_contexts/traffic-probe_context.txt (445 KB packager output) | skipped-generated |
| docubundle/output_contexts/relojistas/outputtotext.sh | header-scan |
| docubundle/output_contexts/relojistas/reduce_output_dir.sh | header-scan |
| docubundle/output_contexts/relojistas/repo_summary.txt (1.1 MB) | skipped-generated (>1 MB) |
| docubundle/output_contexts/relojistas/** remaining (105 files: duplicate of the archive.org capture) | skipped-binary / skipped-generated |
| docubundle/output_contexts/wayfaringlondoner/outputtotext.sh | header-scan |
| docubundle/output_contexts/wayfaringlondoner/reduce_output_dir.sh | header-scan |
| docubundle/output_contexts/wayfaringlondoner/repo_summary.txt (2.3 MB) | skipped-generated (>1 MB) |
| docubundle/output_contexts/wayfaringlondoner/resulttextoutputwayfaringlondoner.txt (7.2 MB) | skipped-generated (>1 MB) |
| docubundle/output_contexts/wayfaringlondoner/** remaining (109 files: duplicate capture assets) | skipped-binary / skipped-generated |

File accounting: 116 individual files outside the capture areas + 217
(archive.org.results) + 225 (docubundle/output_contexts) = 558.

## Concepts

### Traffic-probe mission — intent discovery on parked domains
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** HANDOFF (≈2026-06-13): "Engine (site-engine, stdlib Go) live on a dedicated EU box for relojistas.com … HTTPS, capturing"; plan: P0–P2 done, P3/P4 in progress, P5 not started.
- **what:** Domains that still receive residual visitors but serve only a parking lander are put on a minimal "probe" page that plausibly reflects the old vertical and invites ONE action (search box / category links / free-text). The stated intent is captured server-side; after 2–4 weeks the terms rank which domains have real demand worth building an idea.uk-style site for. Explicit scope boundary: capture what visitors *say they want* on our own page, never recover anyone's old gated content.
- **sources:** TASK_traffic_probe_brief.md#1-2, traffic_probe_plan(12).md#how-it-all-fits, traffic_probe_runbook(13).md#0
- **relations:** probe page pattern, ranking queries + graduation criteria, VM-hosted backend sites class
- **verify-later:** live relojistas.com/stats; intent_events table row counts; sites rows with deploy_config.target='vm'

### Wayback grounding of probe pages
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-11: "Relojistas grounded in the snapshot: it was a Spanish watch FORUM"; 2026-06-13(b): grounding constraint recorded.
- **what:** Before building a probe page, look up what the domain used to be via archive.org (CDX path list, availability API, snapshot view). The old path list (/login, /members, /forum…) signals what was gated and what visitors still want; the snapshot fixes language, vertical, and the invited action. Operational constraint discovered: Claude can web_fetch archive pages only when a search surfaces the exact URL and cannot enumerate CDX on demand — so the operator supplies Wayback URLs/snapshots, or grounding falls back to web search + the domain name.
- **sources:** TASK_traffic_probe_brief.md#2-method, traffic_probe_running_notes(28).md#2026-06-13-b, HANDOFF#thread-c
- **relations:** per-domain notes convention; adoption-pipeline (site recreation from crawl) is the platform cousin
- **verify-later:** archive.org.results/ snapshots exist for both live domains (they do, in this unit)

### Probe page pattern — one invited action, plausible framing
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas live 2026-06-12 (first capture 13:03:44 UTC); wayfaringlondoner page "built + grounded … not yet deployed" (HANDOFF).
- **what:** The page looks intentional, not parked: a one-line tagline matching the old vertical, exactly one invited action (v1: a single text input, kind=search or freetext), a plain privacy line, a 1×1 beacon, no JS, no cookies. Framing follows the domain's heritage — relojistas is a Spanish marketplace/search posture (marca/modelo/reparación/compraventa, thanks at /gracias.html); wayfaringlondoner is an English BLOG posture asking for a destination/story. Hand-made pages for the first domains were explicitly a go-live unblocker; chassis-built pages take over under P3.
- **sources:** TASK_traffic_probe_brief.md#2, relojistas_notes(8).md#decisions, wayfaringlondoner_notes.md#decisions, relojistas_golive/index.html
- **relations:** intent-probe component (the library form of the same pattern), probe content restraint
- **verify-later:** live page HTML vs intent-probe component render

### Minimal-data privacy posture (UK GDPR/PECR)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes "Standing observations": "no cookies, no JS, no IP stored, referer reduced to host, country only from a coarse CDN header"; holds "regardless of volume" (relojistas_notes).
- **what:** Server-side-only logging, no third-party trackers, no non-essential cookies (nothing stored on the device → no consent banner needed), no names/emails collected, free-text treated as potentially personal and not retained longer than needed, plain privacy line on every page. Explicitly declared load-invariant: under traffic pressure the project will not add client-side JS, third-party analytics, or IP logging. Open choice: redact email/phone patterns at ingest vs rely on the 90-day prune.
- **sources:** TASK_traffic_probe_brief.md#4, relojistas_notes(8).md#traffic-handling, traffic_probe_running_notes(28).md#standing-observations
- **relations:** intent event record, ingest validation contract, content-governance (platform-wide posture cousin)
- **verify-later:** intent-probe component privacy_text fallback; engine code stores no IP/UA

### Intent event record (fields and deliberate omissions)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas_notes "What we record" + FIRST LIVE CAPTURE log entry 2026-06-12 13:03:44 UTC.
- **what:** One event per submission: id, host, kind (search|categories|freetext), value (typed text ≤500 runes), ref_host (referer reduced to bare host, blank if same-site), country (coarse CDN header or empty), created_at (UTC), plus landing_query (inbound ?q=/?utm= that survived to the form page — added 2026-06-13 so the structured export carries it without a log join). Deliberately NOT recorded: IP addresses, user agents, cookies, full referer URLs, names/emails. There is no results page: the probe performs no search; the submission itself is the product (303 → thanks page).
- **sources:** relojistas_notes(8).md#what-we-record, traffic_probe_running_notes(28).md#2026-06-13 (landing_query), intent_events_migration(1).sql
- **relations:** minimal-data privacy posture, /events export, intent_events table
- **verify-later:** IntentEvent struct in site-engine repo; events-*.jsonl line shape on box

### Visit beacon and events-per-1k metric
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** runbook §2 "the page must include the visit beacon"; running notes self-correction 2026-06-11 (beacon removed from gracias page).
- **what:** A no-JS, no-cookie 1×1 `<img src="/api/hit">` counts human-with-browser visits per host — the denominator for the project's core metric, intent events per 1,000 visits. The thanks page deliberately carries no beacon so submissions don't inflate the denominator. Because the beacon counts humans only, nginx access logs remain the bot-inclusive ground truth for traffic-claim comparisons.
- **sources:** traffic_probe_runbook(13).md#2, relojistas_notes(8).md#what-we-record, traffic_probe_running_notes(28).md#2026-06-11
- **relations:** intent_site_stats snapshot, traffic-claim verification, access-log passive harvest
- **verify-later:** /api/hit handler in service.go; counters.json per-host visits

### Capture-side input sanitisation with deferred normalisation
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-12 "Sanitisation v2 … Tests green"; a real bug (tab both Cc and whitespace silently joining words, `gmt\t\tmaster` → `gmtmaster`) found and fixed.
- **what:** The engine's sanitizeValue strips Unicode Cc AND Cf (zero-widths, bidi overrides incl. U+202E, BOM, soft hyphen), collapses whitespace runs (IsSpace checked FIRST), caps values by runes not bytes (multibyte-safe), drops junk-only submissions. Deliberate division of labour: NFC normalisation + lowercasing happen at the P4 collector, not the engine — the engine is stdlib-only (no x/text), so NFD combining marks pass through and two byte-forms of "ñ" count as separate terms until ingest normalises.
- **sources:** traffic_probe_running_notes(28).md#2026-06-12 (sanitisation v2), traffic_probe_plan(12).md#P4 ingest contract, relojistas_notes(8).md#decisions
- **relations:** ingest validation contract, ranking queries (lower() caveat)
- **verify-later:** sanitizeValue in site-engine service.go; NFC step in collector action

### /events export endpoint and checkpoint contract
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-12: "GET /events built + tested … Tests green ×6"; HANDOFF lists "/events export endpoint" among live capabilities.
- **what:** Key-gated NDJSON stream of stored events, oldest first, original line bytes preserved; params since (RFC3339, strictly-after), host, limit (default 5000); final `_meta` line {count, truncated, server_time}. Checkpoint contract: collector stores max created_at received; strictly-after semantics + the engine event id make pulls duplicate-free. Lock-free by design so a large export can never block live captures — a torn mid-append tail line is skipped and arrives next pull. Day-file skip by filename date.
- **sources:** traffic_probe_runbook(13).md#6, traffic_probe_running_notes(28).md#2026-06-12 (events built), relojistas_notes(8).md#how-we-see
- **relations:** intent_events table (consumer), pull-not-push collection topology
- **verify-later:** Store.StreamEvents + App.events in site-engine; nginx /events location on box

### Access-log passive harvest and /access-digest
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** passive_harvest_spec(2) 2026-06-13: "Part 2 — access-log digest: DONE" (endpoint built + tested) but "STILL TO DO on the collector side: pull /access-digest per site into a rollup table".
- **what:** The signals the structured event stream can never see on a static page load — external referer, landing path+query, the dead-forum 404 paths (themselves an intent signal: what surviving inbound links point at), and user-agent for bot classification — already sit in nginx's combined access log. Option A (chosen over B: defer to P5 ssh; C: Cloudflare analytics): the engine reads its own box's per-domain log and exposes key-gated `GET /access-digest?host=&since=&top=` returning status mix, top referers (canonicalHost-reduced, self excluded), top paths, top 404 paths, UA buckets (known_search_bot / seo_or_scraper_bot / other_bot / browser_like / empty / other), top real client IPs. Requires setup.sh support: per-domain access_log files, engine user in adm group, CF real_ip conf when proxied.
- **sources:** passive_harvest_spec(2).md, traffic_probe_running_notes(28).md#2026-06-13-g, deploy_setup/working_dir/accessdigest.go (header)
- **relations:** global bot-IP blocklist (same rollup source), traffic-claim verification, Cloudflare-proxied option
- **verify-later:** accessdigest.go in site-engine repo; whether the collector rollup table was ever built

### intent_events table with structural idempotency
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-13(d): "Migration applied (operator: CREATE TABLE + 3 indexes + INSERT 0 1 task)".
- **what:** Cluster-side landing table for pulled events: engine_event_id UNIQUE makes re-pulling overlapping windows a no-op via ON CONFLICT DO NOTHING, so the collector can use a safely-overlapping since. Checkpoint needs no extra storage — next since = max(event_created_at) per host. CHECK constraints on kind enum and value length; host resolved to site_id (nullable FK to sites). Collected_at vs event_created_at kept separate.
- **sources:** intent_events_migration(1).sql, traffic_probe_running_notes(28).md#2026-06-13-b/d
- **relations:** /events checkpoint contract, intent collection topology, ranking queries
- **verify-later:** \d intent_events in clients_db; uq_intent_events_engine_id

### Intent collection topology — collector action under a wrapper-orchestrator
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** intent_collector_registration.sql enable-order: migration applied (done), action/agents/enable steps still pending; scheduled_tasks row "INSERTED DISABLED".
- **what:** Collection needs NO adapter and NO SSH: one Go action (`collect_intent_events`, Category "data", IsLocal) self-queries all sites with deploy_config.target='vm', pulls /events + /stats per box over key-gated HTTPS, and upserts; per-site failures caught and skipped. Because it is scheduler-reached AND does substantive unbounded work, guideline 001's wrapper rule applies: a thin `intent-collection-orchestrator` (spawn→call→complete, med-export pair mirrored verbatim incl. image v1.0.1063) spawns the `intent-collector` task worker in its own pod. The box's INTERNAL_API_KEY lives in sites.deploy_config.engine.stats_key (low-sensitivity read-only export key; movable to a secrets table later). agent_definitions is UNIQUE(type,version), so idempotency uses ON CONFLICT (type, version).
- **sources:** intent_collector_registration.sql, intent_collector_agents(2).sql, intent_events_migration(1).sql#scheduled-collector, traffic_probe_running_notes(28).md#2026-06-13-c/d
- **relations:** scheduler single-fire semantics (design correction), pull-not-push topology, scheduler-and-tasks, development-guide wrapper rule
- **verify-later:** collect_intent_events in GlobalActionRegistry; agent_definitions rows intent-collection-orchestrator/intent-collector; scheduled_tasks 'intent-collection' enabled flag and target_agent_type

### Ingest validation contract
- **category:** traffic-analytics
- **status-signal:** aspirational
- **status-evidence:** plan P4 section (2026-06-12/13 additions) specifies the contract; collector enablement itself still pending.
- **what:** Everything the collector must enforce when pulling engine lines into the DB: parameterised SQL only (values are data, never concatenated — injection structurally impossible per house rule); per-line shape checks (JSON parses, kind ∈ enum, value ≤500 runes, host ∈ accepted set, timestamp sane); burst dedupe of identical (host,value) within a minute as bot noise (raw JSONL stays source of truth); Unicode NFC normalisation + lowercasing HERE (deferred from the stdlib-only engine); DB CHECK constraints; values escaped at every display surface. Open choice: redact email/phone patterns at ingest vs rely on the 90-day prune.
- **sources:** traffic_probe_plan(12).md#P4, relojistas_notes(8).md#decisions (input hygiene), passive_harvest_spec(2).md
- **relations:** capture-side sanitisation (the other half), intent_events table, minimal-data privacy posture
- **verify-later:** validation body of collect_intent_events action

### intent_site_stats visit-count snapshot
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** passive_harvest_spec(2): "Part 1 — visit counts: DONE" (built/validated); enablement rides on the disabled collector.
- **what:** The events-per-1k denominator (visits) lives only in the engine's counters.json exposed at /stats — not in intent_events. A one-row-per-host table holds the latest cumulative /stats snapshot (visits, events, observed_at); the collector's collectSiteStats pulls it non-fatally each run; ranking query 1 LEFT JOINs it for the true rate. History table explicitly deferred until a visits-over-time trend is wanted.
- **sources:** intent_site_stats_migration.sql, passive_harvest_spec(2).md#part-1, intent_ranking_queries(1).sql#1
- **relations:** visit beacon, ranking queries, intent collection topology
- **verify-later:** \d intent_site_stats; collectSiteStats in collector action

### Ranking queries and graduation criteria
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** running notes 2026-06-13(e): "ranking ✓ … Works TODAY on absolute signal"; graduation numbers are an explicit "proposal — tune once data exists" (relojistas_notes).
- **what:** Six read-only queries over intent_events answer "is there demand here?": per-domain summary (with events-per-1k via intent_site_stats), top terms, dominant-cluster share (crude single-term proxy; real clustering a later refinement), referer breakdown, landing-query breakdown, recent raw submissions. Proposed graduation criterion (probe → real build): sustained events-per-1k ≥ 20 AND a dominant intent cluster covering ≥ 30% of terms over 2–4 weeks.
- **sources:** intent_ranking_queries(1).sql, relojistas_notes(8).md#open-choices, passive_harvest_spec(2).md#whats-not-blocked
- **relations:** intent_site_stats, traffic-probe mission (the ranking is the mission's output)
- **verify-later:** whether any report/dashboard consumes these queries

### intent-probe content component
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-11: "intent-probe INSERTED into the live library (INSERT 0 1 …); the second run's INSERT 0 0 is the ON CONFLICT idempotency working."
- **what:** New `content_components` section (STEP ZERO verdict: nothing in the 83-section library captures anonymous intent server-side; contact-form collects PII — the opposite posture). Kebab function `intent-probe`, v2 input schema (tagline/action_label/placeholder/submit_label llm-sourced; probe_kind and privacy_text from config with fallbacks; contact_email from site_specs.identity, skip-if-missing), plain HTML form POST to /intent + beacon img (js_content NULL — JS Content Separation trivially satisfied), CSS-var theming scoped to .intent-probe-section. Deliberate v1 limit: single text-input action only; the {{range}}-based category-buttons variant is deferred until the renderer's array handling is verified ("arrays are where templates fail").
- **sources:** intent_probe_component(1).sql, traffic_probe_running_notes(28).md#2026-06-10 (STEP ZERO) and #2026-06-11
- **relations:** requires-backend capability gate (carries the tag), probe page pattern, contracts-and-standards, tool-library
- **verify-later:** SELECT … FROM content_components WHERE name='intent-probe'; renderer array handling for the categories variant

### Probe content restraint — no results, no imagery, no anchoring
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas_notes decisions dated 2026-06-11 ("No results page in v1", "Imagery: v1 ships text-only").
- **what:** Three linked restraint decisions that protect the signal: (1) no results page — the probe performs no search and returns nothing; revisit only if repeated same-term re-submissions show visitors expect an answer; (2) v1 text-only — no manufacturer/press photos (rights, shop-implication, and any displayed list ANCHORS what visitors search for); v1.1 option is ONE brand-free generated hero via the chassis image pipeline; (3) the "novedades" category-buttons idea would turn the latest-models display into measurement itself (kind=categories) but must run as an A/B against the plain box, with top-terms read before choosing the button set. Status of (2)-hero and (3): aspirational.
- **sources:** relojistas_notes(8).md#decisions (imagery, no-results), traffic_probe_running_notes(28).md#2026-06-11 (imagery)
- **relations:** intent-probe component (deferred categories variant), imagery (platform pipeline)
- **verify-later:** whether any probe page ever gained a hero image or category buttons

### Traffic-claim verification and the bot-vs-human verdict method
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** relojistas_notes Log 2026-06-12: "VERDICT (access log, 14,961 reqs): overwhelmingly bots/ghosts, human intent ≈ 0 … a clean probe result, not a measurement failure."
- **what:** Marketplace visit estimates are treated as unverified relative rankings (relojistas' claimed ~1.2M/mo was the outlier test case). Method: convert the claim to expected visits/hour; compare beacon (humans-only) vs nginx access log (bot-inclusive ground truth); enumerate confounds before concluding (DNS propagation window, humans-only beacon, the invisible www gap); set a dated verdict criterion (48h, UA-split requests/day, www share). Relojistas outcome: 83% 404s on dead vBulletin paths; UA mix Chrome-spoof crawler / Claude-SearchBot / SemrushBot / YandexBot; Cloudflare's "unique visitors" an upper bound dominated by bots. A negative verdict is a successful probe result. By-product: the 404 paths ARE intent and feed the passive harvest.
- **sources:** relojistas_notes(8).md#log (verdict + traffic-claim assessment), traffic_probe_running_notes(28).md#2026-06-13, README_stats_internal_key.md (the settle-it commands)
- **relations:** visit beacon, access-log passive harvest, WWW_ALIAS (closes the www confound), debugging (don't-jump-to-conclusions rule applied)
- **verify-later:** relojistas access-log digests over a longer window; whether any other domain got the same treatment

### Global bot-IP blocklist (Thread D)
- **category:** traffic-analytics
- **status-signal:** aspirational
- **status-evidence:** HANDOFF Thread D: "Design sketch for this thread … This is separate from intent capture but shares the log source" — no build claimed anywhere.
- **what:** Operator idea: relojistas' bot storm makes it a harvesting ground for illegitimate-crawler IPs (high-volume, spoofed-UA, 404-storming, robots.txt-ignoring) to block GLOBALLY across all boxes/sites via a shared denylist applied at the edge (nginx geo/map deny, or Cloudflare where proxied), with legitimate crawlers (Googlebot, Bing, real Claude-SearchBot) allow-listed. Consumes the same UA/IP rollup the access-digest produces.
- **sources:** HANDOFF_vm_sites_permanent_thread.md#thread-d, passive_harvest_spec(2).md#if-option-a
- **relations:** access-log passive harvest (shared source), Cloudflare-proxied option
- **verify-later:** any denylist artifact on the boxes or in vm-sites/site-engine repos

### Relojistas static-rebuild manifest (Thread A)
- **category:** traffic-analytics
- **status-signal:** aspirational
- **status-evidence:** HANDOFF Thread A: "do first; concrete … Open: build now from heritage alone, or wait ~1–2 weeks for P4 intent data? (Lean: scaffold now, enrich from data.)"
- **what:** Despite the bot verdict, relojistas keeps value: an RSS feed real aggregators still pull (populate with OUR content), heavy crawler presence already indexing the domain, and the 404/referer log revealing what inbound links want. Plan: package provenance (Spanish watch forum, boards), language, vertical, an RSS/news section (news-feed pipeline), top inbound 404 paths + referer clusters, and roadmap-pinned section_types into a manifest handed to the framework for a multi-page static build deployed via the same vm-sites Action — optionally retaining intent-probe (capability=backend) or going pure-static.
- **sources:** HANDOFF_vm_sites_permanent_thread.md#thread-a, traffic_probe_running_notes(28).md#2026-06-13-b
- **relations:** news-feed-pipeline, site-plan-and-reconciler (roadmap section_types pinning), VM-hosted backend sites class
- **verify-later:** any relojistas manifest/site_specs/roadmap rows; whether the static build happened

### Domain shortlist and selection policy
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** TASK brief §5 table + §4 "Start with 3–5 high-traffic, clearly generic domains you fully control"; two domains actioned by 2026-06-13.
- **what:** A parked-marketplace export (traffic_probe_domains.tsv, 388 lines) ranked by the marketplace's own estimated visits, with name-based vertical guesses and per-domain probe ideas. Policy: eligibility statuses concern the parking program's monetisation, NOT DNS control; repointing DNS stops parking revenue — choose deliberately; start with a few controlled generic domains; health-adjacent names (healthscare.*, overpronation.com…) need careful non-clinical framing; verify estimates against own logs before committing effort.
- **sources:** TASK_traffic_probe_brief.md#5-7, traffic_probe_domains.tsv (header), traffic_probe_plan(12).md#risks
- **relations:** traffic-claim verification, Wayback grounding
- **verify-later:** which domains beyond relojistas/wayfaringlondoner were ever probed

### Per-domain notes and living-docs convention
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** HANDOFF Thread C: "Each domain gets its own <domain>_notes.md … per the relojistas/wayfaringlondoner template"; cross-thread rule "append, don't fork".
- **what:** Every probe domain gets a living `<domain>_notes.md` holding provenance (what it was, evidence snapshot), dated decisions, open choices, coordinates (box/IP/repos/paths/key location), and a dated log. Project-level knowledge lives in three living docs (plan = decisions + phases; runbook = operational how-to; running notes = per-session reasoning journal with a rename map and "new names per the standing rule" discipline). These are the single source of truth across parallel chats.
- **sources:** relojistas_notes(8).md (the template instance), wayfaringlondoner_notes.md, HANDOFF#cross-thread, traffic_probe_running_notes(28).md#conventions
- **relations:** documentation-system (travelling/living doc conventions)
- **verify-later:** n/a (documentary convention)

### VM-hosted backend sites — a new infrastructure class (proposed doc 024)
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** plan "Genuinely new (proposed doc 024 'VM-Hosted Backend Sites (site-engine)', Infrastructure Reference)" — the class runs live for one domain; the reference doc itself was only proposed ("Draft it in this thread once the shape is agreed", HANDOFF).
- **what:** The genuinely-new platform material the probe project surfaced: a persistent, non-reaped, internet-facing VM class and its lifecycle; DNS + public TLS as managed state outside k8s; a data-RETURN path from off-cluster; the off-cluster "commit is deploy" seam and where its credential lives (repo secrets now, adapter later); capability-gate semantics. Everything else was deliberately mapped onto existing mechanisms (adapter skeleton, thunder ssh, thunder_instances→service_instances, scheduled tasks, discovery checks, in-cluster Actions runner). Probe sites remain first-class `sites` rows so the maintenance/improvement loop covers them automatically — the discovery agents scan live sites over HTTP regardless of hosting.
- **sources:** traffic_probe_plan(12).md#framework-integration, HANDOFF#thread-b, traffic_probe_running_notes(28).md#2026-06-11 (integration mapping)
- **relations:** every concept below in this category; improvement-loop; adapters
- **verify-later:** whether docs024 doc "VM-Hosted Backend Sites" was ever written; sites rows with github_repo='vm-sites'

### site-engine — API-only capture backend for the class
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** HANDOFF: "Engine (site-engine, stdlib Go) live on a dedicated EU box for relojistas.com (CPX22, 167.233.33.159)".
- **what:** A single stdlib-only Go binary (zero deps, no go.sum by design) forked from idea.uk's service (kept: App/routes/cors shape, writeJSON, store pattern; dropped: engine/prompts/audience_check/billing). It does only what static files cannot: POST /intent (capture + 303 to THANKS_PATH), GET /api/hit (visit beacon), GET /stats (key-gated summary), GET /health, GET /events (export), GET /access-digest (log digest). nginx serves the chassis-built static site and proxies only these paths; the engine is never exposed directly, keyed by canonical Host, with ACCEPT_HOSTS as optional defence-in-depth. Explicitly class-level: "First feature: visitor-intent capture … the engine … grows by feature (e.g. chat, boards) later." Superseded first cut: a standalone "probe-go" multi-vhost page-serving service (session 1) — page rendering and per-domain content registry removed once the chassis owned the page.
- **sources:** deploy_setup/site-engine/service.go (header), traffic_probe_runbook(13).md#1-2, traffic_probe_running_notes(28).md#session-1-3, deploy_setup/site-engine/site-engine.env
- **relations:** engine store v2, /events endpoint, access-digest, setup.sh provisioning, idea.uk (fork origin)
- **verify-later:** gqls/site-engine repo contents; systemctl status site-engine on 167.233.33.159

### Engine store v2 — daily JSONL events + debounced counters + on-box retention
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-11: "Store v2 (JSONL) … Burst-tested: 300 events + 100 visits"; prune timer installed at go-live (relojistas_notes 2026-06-12 log).
- **what:** Two pre-launch storage cliffs were fixed structurally: v1 rewrote one ever-growing JSON file on every persist and held all events in RAM (superseded). v2 splits by access pattern — events append to daily JSONL (events-YYYYMMDD.jsonl, one line per submission, O(1) at any volume, rotation = the date, retention = delete old files); /stats counters live in a small counters.json flushed by a dirty-flag 5s debounced flusher (crash window ≤5s of visit counts, never events); SIGTERM/SIGINT flush+fsync. Retention enforced on-box by site-engine-prune.timer (daily delete of events files older than RETENTION_DAYS, default 90); explicitly NO logrotate on engine files (move/truncate would race the open handle) — nginx logs keep their own logrotate.
- **sources:** deploy_setup/site-engine/store.go (header), traffic_probe_running_notes(28).md#2026-06-11 (store fix + store v2 + retention), relojistas_notes(8).md#decisions
- **relations:** /events export (tails these lines), intent event record, minimal-data privacy (90-day prune)
- **verify-later:** store.go in site-engine repo; systemctl list-timers site-engine-prune.timer on box

### Probe as Layer 4 build + thin Layer 5 VM deploy (decisions D1–D4)
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan "Decisions — RESOLVED 2026-06-10" summary block.
- **what:** The structural framing that killed the standalone-project drift: a probe is a normal chassis-built site whose only differences are the deploy target and one capture component. D1: reuse the modern build-dispatch-loop pipeline (no separate probe workflow; pageflow-builder deprecation is a separate call). D2: a second shared repo for VM sites with the identical domain-folders-at-root layout; sites.github_repo selects the target; the static portfolio-sites repo + B2 Action stay untouched. D3: light per-repo Action now ("commit is deploy", target swapped); the heavier chassis-driven service-deployer is the eventual move. D4 moot: no needs_vm_deploy terminal item — the terminal build item stays target-agnostic (assemble + commit); the one-time per-domain VM setup is a separate provisioning step. Deferred: multi-box routing via deploy_config/service_instances only when relocation matters.
- **sources:** traffic_probe_plan(12).md#decisions-resolved + #decision-1-4 analysis, traffic_probe_running_notes(28).md#2026-06-10 (decisions resolved)
- **relations:** vm-sites repo + Action, github_repo target selector, vmhost adapter (the later heavy path), development-guide (build pipeline)
- **verify-later:** build-dispatch-loop handling a vm-sites-designated site end-to-end

### vm-sites content repo and deploy-to-vm Action
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan P2 "*Done: content Action deploy-to-vm.yml + engine Action deploy-engine-to-vm.yml … both validated*"; HANDOFF: "Deploy is 'commit is deploy' via two GitHub Actions … self-hosted runner."
- **what:** A standalone private repo (gqls/vm-sites; created BY HAND because the git-adapter auto-creates repos as PUBLIC; working checkout a sibling of agentchassis, never nested; the docs-tree copy is a reference snapshot only, contextkit pattern). Domain folders at repo ROOT — an assumption bug resolved 2026-06-11: the live sites repo keeps domain folders at root (the `sites/**` variant was a stale copy inside agentchassis/.git/workflows/, which GitHub never reads). The Action is a faithful sibling of the live B2 action: self-hosted runner, dotted-first-segment regex for changed-domain detection (structurally excludes .github/LICENSE/unknown-domain), full-sync fallback on empty diff, secret-presence checks, rsync -az --delete over SSH into /var/www/vm-sites/<domain>; no CF purge; deploys content only for already-provisioned domains. Deletion-propagation gap shared with the B2 action — noted, not fixed.
- **sources:** deploy_setup/vm-deploy/deploy-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-11 (layout resolved; live b2 action learned), traffic_probe_runbook(13).md#3.1+5
- **relations:** deployment-github (B2 action sibling), setup.sh WEBROOT_OWNER, debugging lesson #24
- **verify-later:** gqls/vm-sites .github/workflows; Action run history

### site-engine deploy Action and the narrow-sudo privilege model
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** plan P2 done note; runbook §5; running notes 2026-06-12 "the 3.9 engine-seam test now SHIPS the endpoint".
- **what:** On push of **.go/go.mod to the engine repo: build linux/amd64 (static, stripped) → scp to box → run the root-owned hook /usr/local/sbin/site-engine-deploy which atomically swaps the binary and restarts. Privilege model: no root key in CI; setup.sh (when DEPLOY_USER set) installs the hook plus a sudoers rule scoped to ONLY that script — the deploy user can swap the engine and nothing else; the binary runs as the unprivileged site-engine user. Engine and content deploys are deliberately separate workflows so neither touches the other. x86-only constraint: the Action builds GOARCH=amd64 (Arm boxes would need a build-matrix change).
- **sources:** deploy_setup/vm-deploy/deploy-engine-to-vm.yml (header), traffic_probe_running_notes(28).md#2026-06-10 (engine-deploy workflow + privilege model), traffic_probe_runbook(13).md#5
- **relations:** setup.sh (installs the hook), dedicated-vs-shared box policy (x86 constraint)
- **verify-later:** sudoers rule + hook on box; Action run history in gqls/site-engine

### setup.sh — idempotent multi-vhost box provisioning
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** relojistas_notes log 2026-06-12 12:32: "Box provisioned (setup.sh full run)"; cert issued on idempotent re-run at 13:02.
- **what:** Adapted from idea.uk's authoritative setup.sh into the class-level provisioner: non-interactive (env-var params, positional domains fallback), idempotent (re-run IS rebuild; add a domain = extend DOMAINS + re-run; existing domains untouched), parameterised, self-contained (inline nginx conf + systemd unit). Installs per-domain vhosts serving /var/www/vm-sites/<domain> and proxying only the API paths; webroot certbot per domain with graceful HTTP degradation when DNS lags (re-run upgrades to HTTPS); ufw/fail2ban/logrotate/unattended-upgrades/ssh hardening; deploy sudo hook; prune timer; MODE=full|update. Grown options: WEBROOT_OWNER (deploy-user rsync rights), WWW_ALIAS (opt-in www server_name + cert SAN with getent pre-flight; v1 is apex-only), CLOUDFLARE=true (CF real_ip conf), per-domain access logs + adm group for the digest, version-neutral `listen 443 ssl` (nginx ≥1.25 http2 deprecation found in the field). Warning captured: box-takeover semantics (ufw --force reset, removes nginx default site) — why sharing the idea.uk box was declined.
- **sources:** deploy_setup/vm-deploy/setup.sh (header), traffic_probe_running_notes(28).md#2026-06-10 (box artifact) + 2026-06-12 entries, traffic_probe_runbook(13).md#3.5+4
- **relations:** site-engine deploy hook, multi-domain multiplexing, vmhost adapter (automates this later)
- **verify-later:** setup.sh in site-engine or vm-sites repo vs the docs-tree snapshot

### Multi-domain single-binary hosting and domain onboarding/relocation
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** runbook §4 documented + relojistas live; the shared multi-vhost box itself not yet provisioned (wayfaringlondoner "Awaiting a shared box + DNS").
- **what:** One engine binary per box behind many domains: per-domain nginx server_name blocks each serving that domain's web root and proxying the four API paths; the store keys events by host. Onboarding a new domain = DNS first, extend DOMAINS + re-run setup.sh (vhost + cert), deploy content, verify — the one-time step the content Action never does. Relocation = move web root + add to new box's DOMAINS + repoint DNS (instant if CF-proxied) + drop from old box. Design constraint discovered: THANKS_PATH is engine-wide (one env var per box), so all domains on a shared box must share a thanks filename — standard /thanks.html, each domain shipping its own; relojistas keeps /gracias.html on its dedicated box.
- **sources:** traffic_probe_runbook(13).md#4, wayfaringlondoner_notes.md#decisions, traffic_probe_running_notes(28).md#2026-06-13 (THANKS_PATH design point)
- **relations:** setup.sh, dedicated-vs-shared box policy, vmhost adapter (onboard-domain automation)
- **verify-later:** whether the shared box exists; wayfaringlondoner.com DNS/deployment state

### Dedicated vs shared box policy and VM sizing
- **category:** NEW:vm-backend-sites
- **status-signal:** deployed
- **status-evidence:** relojistas_notes decisions 2026-06-11 (dedicated VM, hosting); HANDOFF: "no new boxes for now" (2026-06-13).
- **what:** Unknown-traffic experiments get their own box (relojistas: Hetzner CPX22, nbg1, IP 167.233.33.159 — sized by disk/log headroom, not CPU; even the claimed 1.2M visits/mo ≈ 0.5 req/s avg is far inside a small box); low-traffic domains share one multi-vhost box; the live idea.uk box is NOT reused (setup.sh box-takeover semantics + product coupling for a ~€3.49/mo saving). Bandwidth analysis: Hetzner EU cloud includes 20 TB/mo (avoid US/Singapore — slashed allowances); 1.2M visits ≈ 360 GB ≈ 2% of allowance. Stay on x86 (amd64 build). Policy hardened 2026-06-13: use EXISTING boxes only for new domains.
- **sources:** relojistas_notes(8).md#decisions+provenance (coordinates), traffic_probe_running_notes(28).md#2026-06-11 (sizing, bandwidth, box question), HANDOFF#where-things-stand
- **relations:** setup.sh takeover semantics, engine deploy Action x86 constraint
- **verify-later:** Hetzner project inventory; whether a shared box was later provisioned

### Pull-not-push off-cluster data return
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** relojistas_notes decision 2026-06-11 "No third 'collector' VM"; the pulling collector itself still disabled.
- **what:** The serving box only buffers (daily JSONL); the CLUSTER pulls over key-gated HTTPS on a schedule into clients_db. Rationale: pull keeps every credential in the cluster — boxes never hold DB or cluster secrets; a push model or middle VM inverts that, adds an attack surface and a hop for no gain. B2 remains optional cold backup. Collection therefore needs no adapter and no SSH — the engine already speaks key-gated HTTPS through nginx (the "key simplification" of P4). SSH is reserved for provisioning (P5).
- **sources:** relojistas_notes(8).md#decisions, traffic_probe_plan(12).md#P4, traffic_probe_running_notes(28).md#2026-06-11 (no collector VM; integration mapping)
- **relations:** /events endpoint, intent collection topology, vmhost adapter (the SSH half)
- **verify-later:** no box-side push cron/credentials exist; collector egress path

### requires-backend capability gate (Decision 5)
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** plan D5 "Outstanding: apply the planner query change"; component-side tag live (component inserted 2026-06-11); planner gate and audit check not applied.
- **what:** Gating backend-requiring sections off static sites keys on the CLASS (site has a server-side backend), not an instance label or site type. Component side: semantic tag `requires-backend` (on intent-probe; future chat/board sections carry the same). Planner side (to apply): load_components gains `AND NOT (COALESCE(semantic_tags,'[]'::jsonb) ? 'requires-backend')` so such components are opt-in via roadmap section_types only. Site side: deploy_config || {"target":"vm","capabilities":["backend"]} at onboarding. Later: an audit check comparing placed sections' requires-* tags against site capabilities → site_work_items findings. Supersedes the first design (an invented `intent-probe` site type in suitable_site_types + a suitable_site_types='[]' planner gate), corrected on operator feedback: "has a backend" is a property of the deploy target, not a site type.
- **sources:** traffic_probe_plan(12).md#decision-5, intent_probe_component(1).sql#gating, intent_probe_component.sql (family-delta: the superseded layer-1 gate), traffic_probe_running_notes(28).md#2026-06-10 (naming correction)
- **relations:** intent-probe component, site-plan-and-reconciler (build-site-planner load_components), design-composition
- **verify-later:** build-site-planner default_config load_components query; sites.deploy_config on any vm site

### sites.github_repo as deploy-target selector (resolveGitRepoName patch)
- **category:** deployment-github
- **status-signal:** aspirational
- **status-evidence:** HANDOFF Thread B: "Chassis patch (P3, still pending) … Land this so the pipeline can target the VM repo at all"; plan P3 "Remaining: land the chassis patch".
- **what:** Tracing showed sites.github_repo is DORMANT end-to-end: git_commit reads config repo_name → default "sites"; upsertSite doesn't SELECT it; ensure_site_record doesn't return it. Specified 3-touch patch: (1) upsertSite RETURNING += COALESCE(github_repo,''), (2) EnsureSiteRecordAction return map += github_repo, (3) a private resolveGitRepoName(config, collected) helper (config repo_name → site_record.github_repo → "sites") used by BOTH git_commit and deploy_image_asset — the latter hardcodes "sites" at line 463 and would otherwise split-brain a probe site (pages → VM repo, images → sites). vet_med_export left alone (dedicated pipeline). Pre-flight confirmed github_repo empty on all 8 sites, so the fallback is safe. CommitToRepo already prefixes <domain>/ for any repo (shared root layout confirmed); createOrGetRepo auto-creates missing repos as PUBLIC — a deliberate-visibility trap.
- **sources:** traffic_probe_running_notes(28).md#2026-06-10 (P3 traced; repo surface complete), traffic_probe_plan(12).md#P3, HANDOFF#thread-b
- **relations:** vm-sites repo, D1–D4, requires-backend gate (the other half of onboarding)
- **verify-later:** grep resolveGitRepoName in platform/orchestration/actions/; deploy_image_asset repo_name resolution; whether the patch ever landed

### backend_unreachable discovery check
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** running notes 2026-06-13(f): "backend_unreachable REWRITTEN against the real DiscoveryCheck interface … gofmt-clean. Enable by adding … to the discovery agent's default_config checks array" — written and interface-reconciled, enablement not claimed.
- **what:** A discovery_checks/ check giving the improvement loop eyes on the VM class: per-site, NOOPs unless deploy_config.target='vm'; probes the public https://<domain>/health; on failure returns a WorkItemSpec (source='discovery', item_type='backend_unreachable', item_key dedup against idx_swi_dedup's partial unique index → one open alert per site, no spam). ALERT not auto-fix: handler_agent empty because a down VM isn't chassis-fixable — sits visible at 'detected'; the P5 vmhost adapter becomes the handler later. SELF-CLEARING: resolves its own open item on recovery using the runner's transaction. A companion `missing_beacon` check (rendered index lacks the /api/hit img) was floated and not built.
- **sources:** traffic_probe_running_notes(28).md#2026-06-13-e/f, traffic_probe_plan(12).md#P4, HANDOFF#cross-thread
- **relations:** VM-hosted backend sites class (first-class sites coverage), scheduler-and-tasks, vmhost adapter
- **verify-later:** discovery_checks/check_backend_unreachable.go in chassis; discovery agent checks array

### P5 vmhost provisioning adapter and service_instances registry
- **category:** NEW:vm-backend-sites
- **status-signal:** aspirational
- **status-evidence:** plan P5 is entirely future-tense; HANDOFF Thread B lists it as pending; "P5 — registry + provisioning adapter" never marked started.
- **what:** The SSH half of the class, automating what runbook §3 does by hand: a `vmhost` adapter (analyser-adapter README skeleton: cmd/vmhost-adapter, internal/adapters/vmhost/ reusing thunder's ssh package via the shared/ precedent, configs, dockerfile, kustomize overlays, Makefile ×4, KafkaTopic system.adapter.vmhost.requests, 003 envelope contract) for provision-box / run setup.sh / onboard-domain (extend DOMAINS + re-run) / ship engine / decommission. Tracked in a `service_instances` table modelled on thunder_instances MINUS the reaper/uptime cap (persistent boxes are never reaped). Thin request actions + a deployer-family agent. Long-term the adapter holds the deploy SSH credential, retiring the repo-secrets copy.
- **sources:** traffic_probe_plan(12).md#P5 + #framework-integration, HANDOFF#thread-b, traffic_probe_running_notes(28).md#2026-06-11 (integration mapping)
- **relations:** adapters (thunder precedent, 003 envelope), setup.sh (what it automates), backend_unreachable (future handler)
- **verify-later:** any vmhost-adapter code/kustomize; service_instances table existence

### Cloudflare-proxied-in-front option
- **category:** NEW:vm-backend-sites
- **status-signal:** partial
- **status-evidence:** running notes 2026-06-13(f): "Cloudflare: relojistas now PROXIED (operator data: 22,046 SSL reqs/24h …)"; the real_ip conf ("set CLOUDFLARE=true on its next setup.sh re-run") still pending at last entry.
- **what:** Optional per-domain posture: keep DNS on Cloudflare with a proxied record → VM origin. Explicitly NOT a second Worker and not a second content copy (a Worker serving a copy would reintroduce the sync problem — avoid); the VM stays the single source of truth, CF just caches. Adjustments: cache-bypass the API paths; nginx set_real_ip_from CF ranges + real_ip_header CF-Connecting-IP (else rate-limiting throttles all of CF as one client and logs/digest/fail2ban see CF IPs); TLS Full (strict). Bonuses: CF-IPCountry populates the country field for free (engine default GeoHeader), and relocation becomes instant (change the origin IP) instead of DNS-TTL-bound.
- **sources:** traffic_probe_runbook(13).md#8, traffic_probe_running_notes(28).md#2026-06-10 (CF clarification) + 2026-06-13-f, passive_harvest_spec(2).md#cloudflare-note
- **relations:** access-digest (real-IP dependency), setup.sh CLOUDFLARE param, multi-domain relocation
- **verify-later:** relojistas CF zone config; cloudflare-realip.conf on box

### Scheduler fires one message per tick — pre_query is a gate, not fan-out
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-13(c): "DESIGN CORRECTED by a real finding: the scheduler fires ONE message per tick — it does NOT fan out pre_query rows (ctx line 5236; thunder-monitor does in-agent loop fan-out, not scheduler fan-out)."
- **what:** A platform fact established from live source and used to correct the collector design: scheduled_tasks.pre_query does not produce per-row dispatch; the live improvement-sweep/thunder-monitor pattern is a count>0 GATE with the fired agent doing in-agent loop fan-out. The intent collector was rewritten from "collect one site from input" to a single self-querying loop-all action accordingly (complexity in Go, one-step workflow); the migration's per-row pre_query was superseded. Also the thunder-monitor convention: INSERT scheduled tasks DISABLED until the action is deployed.
- **sources:** traffic_probe_running_notes(28).md#2026-06-13-c, intent_events_migration(1).sql#scheduled-collector (gate form), deploy_setup/working_dir/intent_events_migration.sql (family-delta: superseded fan-out form)
- **relations:** intent collection topology, scheduler-and-tasks doc 010
- **verify-later:** kafka-scheduler dispatch code path (one fire per tick)

### Traffic-probe field lessons absorbed into the debug guide (#24–#28)
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-12 "Debug guide updated … 016_debugging_guide_v2_46" and 2026-06-13(g) "Debug guide v2_48".
- **what:** Five checklist entries earned in this project's field work, each rule + dated instance: #24 a config/workflow file is only authoritative at its runtime read-path (the stale agentchassis/.git/workflows/deploy-to-b2.yml nearly produced a never-firing Action); #25 prove the harness delivered the intended input before debugging the system (dash not expanding $'…' made the field literally "$value"); #26 shell variables never reach child processes without export, die with the session, and error-text-vs-source mismatch means a stale deployed artifact — read state back from the artifact, not `echo $KEY`; #27 never invent an interface — compiling standalone ≠ satisfying the real DiscoveryCheck signature; #28 agent_definitions is UNIQUE(type,version) with two similar category columns. Plus operator-handover lessons: explicit file manifests + a loud go vet/build check, flat-shipped workflows (delivery channel rejects dot-dirs), git branch -M main before first push.
- **sources:** traffic_probe_running_notes(28).md#2026-06-12 (debug guide v2_46, operator execution) + #2026-06-13-g (v2_48), traffic_probe_runbook(13).md#3.5-3.6 (traps in place)
- **relations:** debugging (016 guide family), per-domain notes convention
- **verify-later:** 016_debugging_guide latest version contains #24–#28

### Traffic-probe context packaging (docubundle)
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** package_traffic_probe.sh header + its output traffic-probe_context.txt (445 KB) and repo_summary bundles present in the unit.
- **what:** A self-contained packager bundles the task brief, domain list, reusable Go service, and deploy/persistence/VM docs into one context file so a new chat can start cold — coping with the messy versioned folder by resolving each doc to the newest (N) variant by mtime and dropping *.orig* backups. Companion scripts (outputtotext.sh, reduce_output_dir.sh) flatten captured site directories into repo_summary.txt bundles. The same cold-start pattern produced the HANDOFF file for the permanent thread.
- **sources:** package_traffic_probe.sh (header), docubundle/output_contexts/relojistas/outputtotext.sh, HANDOFF_vm_sites_permanent_thread.md (the product of the pattern)
- **relations:** documentation-system (context packaging, travelling docs), per-domain notes convention
- **verify-later:** n/a (tooling snapshot)
