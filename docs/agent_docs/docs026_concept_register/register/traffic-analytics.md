# Register — traffic-analytics

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

21 concepts, consolidated from 58 raw extractions (29 unique blocks, each duplicated
once in the source cluster file) across units U05, U11, U18, U24c.

### TRF-001 — Traffic-probe mission/program (residual-traffic intent capture on parked domains)
- **status:** deployed
- **status-evidence:** "FIRST LIVE CAPTURE" 2026-06-12 13:03:44 UTC (kind=search, "correa Omega Seamaster"); relojistas.com live behind Cloudflare 2026-06-13; HANDOFF (~2026-06-13): "Engine (site-engine, stdlib Go) live on a dedicated EU box for relojistas.com … HTTPS, capturing"; plan P0-P2 done, P3/P4 in progress, P5 not started.
- **what:** Domains that still receive residual visitors but serve only a parking lander are put on the chassis as first-class `sites` rows carrying a minimal "probe" page that plausibly reflects the old vertical and invites ONE action (search box / category links / free-text). The stated intent is captured server-side by a stdlib-only Go engine (site-engine, forked from the idea.uk store pattern); after 2-4 weeks the terms rank which domains have real demand worth building an idea.uk-style site for. Explicit scope boundary: capture what visitors *say they want* on our own page, never recover anyone's old gated content. End-to-end: VM (nginx + site-engine) serves and captures; the cluster pulls data on schedule; the framework treats each as a normal site.
- **sources:** TASK_traffic_probe_brief.md#1-2; traffic_probe_plan(11).md#how-it-all-fits / (12).md#how-it-all-fits; traffic_probe_runbook(12).md#0 / (13).md#0; traffic_probe_running_notes(27).md#2026-06-12-first-live-capture; golang_code/service.go, store.go, main.go (headers, filed under a different unit by accident)
- **relations:** probe page pattern (TRF-003); ranking queries + graduation criteria (TRF-014); VM-hosted backend sites class; intent-probe component (TRF-015)
- **verify-later:** live relojistas.com/stats; intent_events table row counts; sites rows with deploy_config.target='vm'

### TRF-002 — Wayback grounding of probe pages
- **status:** deployed
- **status-evidence:** running notes 2026-06-11: "Relojistas grounded in the snapshot: it was a Spanish watch FORUM"; 2026-06-13(b): grounding constraint recorded.
- **what:** Before building a probe page, look up what the domain used to be via archive.org (CDX path list, availability API, snapshot view). The old path list (/login, /members, /forum…) signals what was gated and what visitors still want; the snapshot fixes language, vertical, and the invited action. Operational constraint discovered: Claude can web_fetch archive pages only when a search surfaces the exact URL and cannot enumerate CDX on demand — so the operator supplies Wayback URLs/snapshots, or grounding falls back to web search + the domain name.
- **sources:** TASK_traffic_probe_brief.md#2-method, traffic_probe_running_notes(28).md#2026-06-13-b, HANDOFF#thread-c
- **relations:** per-domain notes convention (TRF-021); adoption-pipeline (site recreation from crawl) is the platform cousin
- **verify-later:** archive.org.results/ snapshots exist for both live domains (they do, in this unit)

### TRF-003 — Probe page pattern — one invited action, plausible framing
- **status:** deployed
- **status-evidence:** relojistas live 2026-06-12 (first capture 13:03:44 UTC); wayfaringlondoner page "built + grounded … not yet deployed" (HANDOFF).
- **what:** The page looks intentional, not parked: a one-line tagline matching the old vertical, exactly one invited action (v1: a single text input, kind=search or freetext), a plain privacy line, a 1x1 beacon, no JS, no cookies. Framing follows the domain's heritage — relojistas is a Spanish marketplace/search posture (marca/modelo/reparación/compraventa, thanks at /gracias.html); wayfaringlondoner is an English BLOG posture asking for a destination/story. Hand-made pages for the first domains were explicitly a go-live unblocker; chassis-built pages take over under P3.
- **sources:** TASK_traffic_probe_brief.md#2, relojistas_notes(8).md#decisions, wayfaringlondoner_notes.md#decisions, relojistas_golive/index.html
- **relations:** intent-probe component (the library form of the same pattern, TRF-015); probe content restraint (TRF-016)
- **verify-later:** live page HTML vs intent-probe component render

### TRF-004 — Minimal-data privacy posture (UK GDPR/PECR)
- **status:** deployed
- **status-evidence:** running notes "Standing observations": "no cookies, no JS, no IP stored, referer reduced to host, country only from a coarse CDN header"; holds "regardless of volume" (relojistas_notes).
- **what:** Server-side-only logging, no third-party trackers, no non-essential cookies (nothing stored on the device → no consent banner needed), no names/emails collected, free-text treated as potentially personal and not retained longer than needed, plain privacy line on every page. Explicitly declared load-invariant: under traffic pressure the project will not add client-side JS, third-party analytics, or IP logging. Open choice: redact email/phone patterns at ingest vs rely on the 90-day prune. (The same posture is also described independently under content-governance — see CGV-027.)
- **sources:** TASK_traffic_probe_brief.md#4, relojistas_notes(8).md#traffic-handling, traffic_probe_running_notes(28).md#standing-observations
- **relations:** intent event record (TRF-005); ingest validation contract (TRF-012); privacy posture (content-governance CGV-027)
- **verify-later:** intent-probe component privacy_text fallback; engine code stores no IP/UA

### TRF-005 — Intent event record (fields, deliberate omissions, landing_query enrichment)
- **status:** deployed
- **status-evidence:** relojistas_notes "What we record" + FIRST LIVE CAPTURE log entry 2026-06-12 13:03:44 UTC; running_notes 2026-06-13 "Small legitimate engine enrichment shipped: landing_query field on IntentEvent … Tested … Additive, no breaking change".
- **what:** One event per submission: id, host, kind (search|categories|freetext), value (typed text ≤500 runes), ref_host (referer reduced to bare host, blank if same-site), country (coarse CDN header or empty), created_at (UTC), plus landing_query (inbound ?q=/?utm= that survived to the form page — added 2026-06-13 so the structured export carries it without a log join). Deliberately NOT recorded: IP addresses, user agents, cookies, full referer URLs, names/emails. There is no results page: the probe performs no search; the submission itself is the product (303 → thanks page).
- **sources:** relojistas_notes(8).md#what-we-record, traffic_probe_running_notes(28).md#2026-06-13 (landing_query), intent_events_migration(1).sql, traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict
- **relations:** minimal-data privacy posture (TRF-004); /events export (TRF-008); intent_events table (TRF-010); access-log passive harvest (TRF-009, complements referer host)
- **verify-later:** IntentEvent struct in site-engine repo; events-*.jsonl line shape on box; service.go IntentEvent.LandingQuery / landingQuery() helper

### TRF-006 — Visit beacon and events-per-1k metric
- **status:** deployed
- **status-evidence:** runbook §2 "the page must include the visit beacon"; running notes self-correction 2026-06-11 (beacon removed from gracias page); ranking query 1 LEFT JOINs for events-per-1k.
- **what:** A no-JS, no-cookie 1x1 `<img src="/api/hit">` counts human-with-browser visits per host — the denominator for the project's core metric, intent events per 1,000 visits. The thanks page deliberately carries no beacon so submissions don't inflate the denominator. Because the beacon counts humans only, nginx access logs remain the bot-inclusive ground truth for traffic-claim comparisons. Visits live in the engine's counters.json (/stats), not in intent_events, so the rate metric requires joining the intent_site_stats snapshot.
- **sources:** traffic_probe_runbook(13).md#2, relojistas_notes(8).md#what-we-record, traffic_probe_running_notes(28).md#2026-06-11; deploy_setup/working_dir/service(24).go#header, traffic_probe_runbook(12).md#6
- **relations:** intent_site_stats snapshot (TRF-013); traffic-claim verification (TRF-017); access-log passive harvest (TRF-009)
- **verify-later:** /api/hit handler in service.go; counters.json per-host visits

### TRF-007 — Capture-side input sanitisation with deferred normalisation
- **status:** deployed
- **status-evidence:** running notes 2026-06-12 "Sanitisation v2 … Tests green"; a real bug (tab both Cc and whitespace silently joining words, `gmt\t\tmaster` → `gmtmaster`) found and fixed.
- **what:** The engine's sanitizeValue strips Unicode Cc AND Cf (zero-widths, bidi overrides incl. U+202E, BOM, soft hyphen), collapses whitespace runs (IsSpace checked FIRST), caps values by runes not bytes (multibyte-safe), drops junk-only submissions. Deliberate division of labour: NFC normalisation + lowercasing happen at the P4 collector, not the engine — the engine is stdlib-only (no x/text), so NFD combining marks pass through and two byte-forms of "ñ" count as separate terms until ingest normalises. (Same underlying mechanism also described independently under content-quality — see CQ-013.)
- **sources:** traffic_probe_running_notes(28).md#2026-06-12 (sanitisation v2), traffic_probe_plan(12).md#P4 ingest contract, relojistas_notes(8).md#decisions
- **relations:** ingest validation contract (TRF-012); ranking queries (lower() caveat, TRF-014); input sanitisation (content-quality CQ-013)
- **verify-later:** sanitizeValue in site-engine service.go; NFC step in collector action

### TRF-008 — /events export endpoint and checkpoint contract
- **status:** deployed
- **status-evidence:** running notes 2026-06-12: "GET /events built + tested … Tests green x6"; HANDOFF lists "/events export endpoint" among live capabilities.
- **what:** Key-gated NDJSON stream of stored events, oldest first, original line bytes preserved; params since (RFC3339, strictly-after), host, limit (default 5000); final `_meta` line {count, truncated, server_time}. Checkpoint contract: collector stores max created_at received; strictly-after semantics + the engine event id make pulls duplicate-free. Lock-free by design so a large export can never block live captures — a torn mid-append tail line is skipped and arrives next pull. Day-file skip by filename date.
- **sources:** traffic_probe_runbook(13).md#6, traffic_probe_running_notes(28).md#2026-06-12 (events built), relojistas_notes(8).md#how-we-see
- **relations:** intent_events table (consumer, TRF-010); pull-not-push collection topology (TRF-011)
- **verify-later:** Store.StreamEvents + App.events in site-engine; nginx /events location on box

### TRF-009 — Access-log passive harvest and /access-digest
- **status:** deployed
- **status-evidence:** passive_harvest_spec(2) 2026-06-13: "Part 2 — access-log digest: DONE" (endpoint built + tested); accessdigest(1).go header "parse this box's nginx combined access log into a compact, key-gated digest"; running_notes 2026-06-13(g) "/access-digest endpoint BUILT + tested … Builds clean". Collector-side rollup ("pull /access-digest per site into a rollup table") remained outstanding at last check.
- **what:** The signals the structured event stream can never see on a static page load — external referer, landing path+query, dead-forum 404 paths (themselves an intent signal: what surviving inbound links point at), and user-agent for bot classification — already sit in nginx's combined access log. Of three options considered (A: engine reads its own log; B: defer to the P5 SSH vmhost adapter; C: Cloudflare analytics if proxied), Option A was chosen and built: the engine exposes key-gated `GET /access-digest?host=&since=&top=` returning status mix, top referers (canonicalHost-reduced, self excluded), top paths, top 404 paths, UA buckets (known_search_bot / seo_or_scraper_bot / other_bot / browser_like / empty / other), top real client IPs. Requires setup.sh support: per-domain access_log files, engine user in adm group, CF real_ip conf when proxied.
- **sources:** passive_harvest_spec(2).md, traffic_probe_running_notes(28).md#2026-06-13-g, deploy_setup/working_dir/accessdigest.go / accessdigest(1).go (header), traffic_probe_runbook(12).md#6
- **relations:** global bot-IP blocklist (same rollup source, TRF-018); traffic-claim verification (TRF-017); visit beacon (TRF-006)
- **verify-later:** accessdigest.go buildAccessDigest/classifyUA/safeHost; NGINX_LOG_DIR config; whether the collector rollup table was ever built

### TRF-010 — intent_events table with structural idempotency
- **status:** deployed
- **status-evidence:** running notes 2026-06-13(d): "Migration applied (operator: CREATE TABLE + 3 indexes + INSERT 0 1 task)".
- **what:** Cluster-side landing table for pulled events: engine_event_id UNIQUE makes re-pulling overlapping windows a no-op via ON CONFLICT DO NOTHING, so the collector can use a safely-overlapping since. Checkpoint needs no extra storage — next since = max(event_created_at) per host. CHECK constraints on kind enum and value length; host resolved to site_id (nullable FK to sites). Collected_at vs event_created_at kept separate.
- **sources:** intent_events_migration(1).sql, traffic_probe_running_notes(28).md#2026-06-13-b/d
- **relations:** /events checkpoint contract (TRF-008); intent collection topology (TRF-011); ranking queries (TRF-014)
- **verify-later:** \d intent_events in clients_db; uq_intent_events_engine_id

### TRF-011 — Intent collection topology (collector under a wrapper-orchestrator, P4 VM collection)
- **status:** partial
- **status-evidence:** intent_collector_registration.sql enable-order: migration applied (done), action/agents/enable steps still pending; scheduled_tasks row "INSERTED DISABLED"; 119 "Pattern... mirrors... the thunder-training-monitor convention of INSERTING DISABLED until the action is deployed".
- **what:** Collection needs NO adapter and NO SSH: one Go action (`collect_intent_events`, Category "data", IsLocal) self-queries all sites with deploy_config.target='vm', pulls /events + /stats per box over key-gated HTTPS, and upserts into intent_events (engine_event_id UNIQUE for idempotency) and intent_site_stats (cumulative visit counters, one row per host, so ranking can compute true events-per-1k-visits); per-site failures caught and skipped. Because it is scheduler-reached AND does substantive unbounded work, guideline 001's wrapper rule applies: a thin `intent-collection-orchestrator` (spawn→call→complete, med-export pair mirrored verbatim incl. image v1.0.1063) spawns the `intent-collector` task worker in its own pod. The box's INTERNAL_API_KEY lives in sites.deploy_config.engine.stats_key (low-sensitivity read-only export key). agent_definitions is UNIQUE(type,version), so idempotency uses ON CONFLICT (type, version). kind constrained to search/categories/freetext.
- **sources:** intent_collector_registration.sql, intent_collector_agents(2).sql, intent_events_migration(1).sql#scheduled-collector, traffic_probe_running_notes(28).md#2026-06-13-c/d; 119_intent_events_for_vms.sql; 120_intent_site_stats.sql; 121_intent_collector_agents.sql
- **relations:** scheduler single-fire semantics (design correction); pull-not-push topology; scheduler-and-tasks; development-guide wrapper rule; intent capture engine on the VM side
- **verify-later:** collect_intent_events in GlobalActionRegistry; agent_definitions rows intent-collection-orchestrator/intent-collector; scheduled_tasks 'intent-collection' enabled flag and target_agent_type

### TRF-012 — Ingest validation contract
- **status:** aspirational
- **status-evidence:** plan P4 section (2026-06-12/13 additions) specifies the contract; collector enablement itself still pending.
- **what:** Everything the collector must enforce when pulling engine lines into the DB: parameterised SQL only (values are data, never concatenated — injection structurally impossible per house rule); per-line shape checks (JSON parses, kind ∈ enum, value ≤500 runes, host ∈ accepted set, timestamp sane); burst dedupe of identical (host,value) within a minute as bot noise (raw JSONL stays source of truth); Unicode NFC normalisation + lowercasing HERE (deferred from the stdlib-only engine); DB CHECK constraints; values escaped at every display surface. Open choice: redact email/phone patterns at ingest vs rely on the 90-day prune.
- **sources:** traffic_probe_plan(12).md#P4, relojistas_notes(8).md#decisions (input hygiene), passive_harvest_spec(2).md
- **relations:** capture-side sanitisation (the other half, TRF-007); intent_events table (TRF-010); minimal-data privacy posture (TRF-004)
- **verify-later:** validation body of collect_intent_events action

### TRF-013 — intent_site_stats visit-count snapshot
- **status:** partial
- **status-evidence:** passive_harvest_spec(2): "Part 1 — visit counts: DONE" (built/validated); enablement rides on the disabled collector.
- **what:** The events-per-1k denominator (visits) lives only in the engine's counters.json exposed at /stats — not in intent_events. A one-row-per-host table (PK host) holds the latest cumulative /stats snapshot (visits, events, observed_at); the collector's collectSiteStats pulls it non-fatally each run and upserts; ranking query 1 LEFT JOINs it for the true rate. History table explicitly deferred until a visits-over-time trend is wanted.
- **sources:** intent_site_stats_migration.sql, passive_harvest_spec(2).md#part-1, intent_ranking_queries(1).sql#1, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** visit beacon (TRF-006); ranking queries (TRF-014); intent collection topology (TRF-011)
- **verify-later:** \d intent_site_stats; collectSiteStats in collector action

### TRF-014 — Ranking queries and graduation criteria
- **status:** partial
- **status-evidence:** running notes 2026-06-13(e): "ranking ✓ … Works TODAY on absolute signal"; graduation numbers are an explicit "proposal — tune once data exists" (relojistas_notes).
- **what:** Six read-only queries over intent_events answer "is there demand here?": per-domain summary (with events-per-1k via intent_site_stats), top terms, dominant-cluster share (crude single-term proxy; real clustering a later refinement), referer breakdown, landing-query breakdown, recent raw submissions. Proposed graduation criterion (probe → real build): sustained events-per-1k ≥ 20 AND a dominant intent cluster covering ≥ 30% of terms over 2-4 weeks.
- **sources:** intent_ranking_queries(1).sql, relojistas_notes(8).md#open-choices, passive_harvest_spec(2).md#whats-not-blocked, traffic_probe_running_notes(27).md#2026-06-13-e
- **relations:** intent_site_stats (TRF-013); traffic-probe mission (the ranking is the mission's output, TRF-001)
- **verify-later:** whether any report/dashboard consumes these queries

### TRF-015 — intent-probe content component
- **status:** deployed
- **status-evidence:** running notes 2026-06-11: "intent-probe INSERTED into the live library (INSERT 0 1 …); the second run's INSERT 0 0 is the ON CONFLICT idempotency working."
- **what:** New `content_components` section (STEP ZERO verdict: nothing in the 83-section library captures anonymous intent server-side; contact-form collects PII — the opposite posture). Kebab function `intent-probe`, v2 input schema (tagline/action_label/placeholder/submit_label llm-sourced; probe_kind and privacy_text from config with fallbacks; contact_email from site_specs.identity, skip-if-missing), plain HTML form POST to /intent + beacon img (js_content NULL — JS Content Separation trivially satisfied), CSS-var theming scoped to .intent-probe-section. Deliberate v1 limit: single text-input action only; the {{range}}-based category-buttons variant is deferred until the renderer's array handling is verified ("arrays are where templates fail").
- **sources:** intent_probe_component(1).sql, traffic_probe_running_notes(28).md#2026-06-10 (STEP ZERO) and #2026-06-11
- **relations:** requires-backend capability gate (carries the tag); probe page pattern (TRF-003); tool-library
- **verify-later:** SELECT … FROM content_components WHERE name='intent-probe'; renderer array handling for the categories variant

### TRF-016 — Probe content restraint — no results, no imagery, no anchoring
- **status:** deployed
- **status-evidence:** relojistas_notes decisions dated 2026-06-11 ("No results page in v1", "Imagery: v1 ships text-only").
- **what:** Three linked restraint decisions that protect the signal: (1) no results page — the probe performs no search and returns nothing; revisit only if repeated same-term re-submissions show visitors expect an answer; (2) v1 text-only — no manufacturer/press photos (rights, shop-implication, and any displayed list ANCHORS what visitors search for); v1.1 option is ONE brand-free generated hero via the chassis image pipeline; (3) the "novedades" category-buttons idea would turn the latest-models display into measurement itself (kind=categories) but must run as an A/B against the plain box, with top-terms read before choosing the button set. Status of (2)-hero and (3): aspirational.
- **sources:** relojistas_notes(8).md#decisions (imagery, no-results), traffic_probe_running_notes(28).md#2026-06-11 (imagery)
- **relations:** intent-probe component (deferred categories variant, TRF-015); imagery (platform pipeline)
- **verify-later:** whether any probe page ever gained a hero image or category buttons

### TRF-017 — Traffic-claim verification and the bot-vs-human verdict method
- **status:** deployed
- **status-evidence:** relojistas_notes Log 2026-06-12: "VERDICT (access log, 14,961 reqs): overwhelmingly bots/ghosts, human intent ≈ 0 … a clean probe result, not a measurement failure."
- **what:** Marketplace visit estimates are treated as unverified relative rankings (relojistas' claimed ~1.2M/mo was the outlier test case). Method: convert the claim to expected visits/hour; compare beacon (humans-only) vs nginx access log (bot-inclusive ground truth); enumerate confounds before concluding (DNS propagation window, humans-only beacon, the invisible www gap); set a dated verdict criterion (48h, UA-split requests/day, www share). Relojistas outcome: 83% 404s on dead vBulletin paths; UA mix Chrome-spoof crawler / Claude-SearchBot / SemrushBot / YandexBot; Cloudflare's "unique visitors" an upper bound dominated by bots. A negative verdict is a successful probe result. By-product: the 404 paths ARE intent and feed the passive harvest.
- **sources:** relojistas_notes(8).md#log (verdict + traffic-claim assessment), traffic_probe_running_notes(28).md#2026-06-13, README_stats_internal_key.md (the settle-it commands)
- **relations:** visit beacon (TRF-006); access-log passive harvest (TRF-009); debugging (don't-jump-to-conclusions rule applied)
- **verify-later:** relojistas access-log digests over a longer window; whether any other domain got the same treatment

### TRF-018 — Global bot-IP blocklist (Thread D)
- **status:** aspirational
- **status-evidence:** HANDOFF Thread D: "Design sketch for this thread … This is separate from intent capture but shares the log source" — no build claimed anywhere.
- **what:** Operator idea: relojistas' bot storm makes it a harvesting ground for illegitimate-crawler IPs (high-volume, spoofed-UA, 404-storming, robots.txt-ignoring) to block GLOBALLY across all boxes/sites via a shared denylist applied at the edge (nginx geo/map deny, or Cloudflare where proxied), with legitimate crawlers (Googlebot, Bing, real Claude-SearchBot) allow-listed. Consumes the same UA/IP rollup the access-digest produces.
- **sources:** HANDOFF_vm_sites_permanent_thread.md#thread-d, passive_harvest_spec(2).md#if-option-a
- **relations:** access-log passive harvest (shared source, TRF-009); Cloudflare-proxied option
- **verify-later:** any denylist artifact on the boxes or in vm-sites/site-engine repos

### TRF-019 — Relojistas static-rebuild manifest (Thread A)
- **status:** aspirational
- **status-evidence:** HANDOFF Thread A: "do first; concrete … Open: build now from heritage alone, or wait ~1-2 weeks for P4 intent data? (Lean: scaffold now, enrich from data.)"
- **what:** Despite the bot verdict, relojistas keeps value: an RSS feed real aggregators still pull (populate with OUR content), heavy crawler presence already indexing the domain, and the 404/referer log revealing what inbound links want. Plan: package provenance (Spanish watch forum, boards), language, vertical, an RSS/news section (news-feed pipeline), top inbound 404 paths + referer clusters, and roadmap-pinned section_types into a manifest handed to the framework for a multi-page static build deployed via the same vm-sites Action — optionally retaining intent-probe (capability=backend) or going pure-static.
- **sources:** HANDOFF_vm_sites_permanent_thread.md#thread-a, traffic_probe_running_notes(28).md#2026-06-13-b
- **relations:** news-feed-pipeline; site-plan-and-reconciler (roadmap section_types pinning); VM-hosted backend sites class
- **verify-later:** any relojistas manifest/site_specs/roadmap rows; whether the static build happened

### TRF-020 — Domain shortlist and selection policy
- **status:** deployed
- **status-evidence:** TASK brief §5 table + §4 "Start with 3-5 high-traffic, clearly generic domains you fully control"; two domains actioned by 2026-06-13.
- **what:** A parked-marketplace export (traffic_probe_domains.tsv, 388 lines) ranked by the marketplace's own estimated visits, with name-based vertical guesses and per-domain probe ideas. Policy: eligibility statuses concern the parking program's monetisation, NOT DNS control; repointing DNS stops parking revenue — choose deliberately; start with a few controlled generic domains; health-adjacent names (healthscare.*, overpronation.com…) need careful non-clinical framing; verify estimates against own logs before committing effort.
- **sources:** TASK_traffic_probe_brief.md#5-7, traffic_probe_domains.tsv (header), traffic_probe_plan(12).md#risks
- **relations:** traffic-claim verification (TRF-017); Wayback grounding (TRF-002)
- **verify-later:** which domains beyond relojistas/wayfaringlondoner were ever probed

### TRF-021 — Per-domain notes and living-docs convention
- **status:** deployed
- **status-evidence:** HANDOFF Thread C: "Each domain gets its own <domain>_notes.md … per the relojistas/wayfaringlondoner template"; cross-thread rule "append, don't fork".
- **what:** Every probe domain gets a living `<domain>_notes.md` holding provenance (what it was, evidence snapshot), dated decisions, open choices, coordinates (box/IP/repos/paths/key location), and a dated log. Project-level knowledge lives in three living docs (plan = decisions + phases; runbook = operational how-to; running notes = per-session reasoning journal with a rename map and "new names per the standing rule" discipline). These are the single source of truth across parallel chats.
- **sources:** relojistas_notes(8).md (the template instance), wayfaringlondoner_notes.md, HANDOFF#cross-thread, traffic_probe_running_notes(28).md#conventions
- **relations:** documentation-system (travelling/living doc conventions)
- **verify-later:** n/a (documentary convention)
