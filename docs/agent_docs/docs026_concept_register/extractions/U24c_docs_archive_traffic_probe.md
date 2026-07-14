# EXTRACTION U24c — docs/_archive/agent_docs/docs024_key_docs_latest/traffic_probe/
Extracted 2026-07-13 (recovered from a sub-agent of U24 that completed before the parent hit the session limit). Part of U24 (docs/_archive/, ~372 files) — covers the archived copy of the traffic_probe project (site-engine capture backend, VM-hosted backend sites class).

## Coverage
| file | treatment |
|---|---|
| docs/_archive/.../traffic_probe/deploy_setup/site-engine/env.go | header-scan |
| docs/_archive/.../traffic_probe/deploy_setup/working_dir/accessdigest(1).go | header-scan |
| docs/_archive/.../traffic_probe/deploy_setup/working_dir/intent_collector_agents(1).sql | family-delta |
| docs/_archive/.../traffic_probe/deploy_setup/working_dir/intent_collector_agents(3).sql | family-delta |
| docs/_archive/.../traffic_probe/deploy_setup/working_dir/main(15).go | header-scan |
| docs/_archive/.../traffic_probe/deploy_setup/working_dir/main(17).go | header-scan |
| docs/_archive/.../traffic_probe/deploy_setup/working_dir/main(19).go | header-scan |
| docs/_archive/.../traffic_probe/deploy_setup/working_dir/main.go | header-scan |
| docs/_archive/.../traffic_probe/deploy_setup/working_dir/service(24).go | header-scan |
| docs/_archive/.../traffic_probe/traffic_probe_plan.md ... (1)-(6) | family-delta |
| docs/_archive/.../traffic_probe/traffic_probe_plan(11).md | family-latest |
| docs/_archive/.../traffic_probe/traffic_probe_runbook.md ... (1)-(7) | family-delta |
| docs/_archive/.../traffic_probe/traffic_probe_runbook(12).md | family-latest |
| docs/_archive/.../traffic_probe/traffic_probe_running_notes.md, (1),(4)-(22) | family-delta |
| docs/_archive/.../traffic_probe/traffic_probe_running_notes(27).md | family-latest |

Notes on families: archived plan(11) ≡ live plan(12), archived runbook(12) ≡ live runbook(13), archived running_notes(27) ≡ live running_notes(28) (byte-identical; the live "latest" only bumps the numeric suffix). So the archived latest of each family carries no dropped concepts vs live; the delta value is entirely in the EARLIER archived versions. SQL (1)/(3) ≡ live intent_collector_agents(2).sql.

## Concepts

### Traffic-probe program (residual-traffic intent capture)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** "FIRST LIVE CAPTURE" 2026-06-12 13:03:44 UTC (kind=search, "correa Omega Seamaster"); relojistas.com live behind Cloudflare 2026-06-13.
- **what:** Put dormant-but-still-visited domains on the chassis as first-class sites whose page plausibly reflects the old vertical and offers ONE invited action ("what are you looking for?"). Captured intent ranks which domains are worth building out. End-to-end: VM (nginx + site-engine) serves + captures, cluster pulls data on schedule, framework treats each as a normal `sites` row.
- **sources:** traffic_probe_plan(11).md#how-it-all-fits, traffic_probe_running_notes(27).md#2026-06-12-first-live-capture, traffic_probe_runbook(12).md#0
- **relations:** parent of site-engine, intent-probe component, P4 collection, VM-hosted backend sites class
- **verify-later:** `sites` rows with deploy_config.target='vm'; intent_events table

### site-engine (API-only capture backend)
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** service(24).go header "site-engine: the capture backend for VM-hosted backend sites (API only)"; builds/tests pass; running live on relojistas box.
- **what:** A stdlib-only Go binary that does the one thing a static page cannot: record a structured intent event server-side keyed by Host into a file store. Endpoints: `POST /intent` (capture then 303), `GET /api/hit` (1×1 beacon), `GET /stats` (key-gated), `GET /health`, plus later `GET /events` and `GET /access-digest`. No page rendering or content registry (the chassis owns both).
- **sources:** deploy_setup/working_dir/service(24).go#header, deploy_setup/working_dir/main.go#header, traffic_probe_runbook(12).md#1
- **relations:** replaces the abandoned standalone probe-go fork; page content owned by chassis intent-probe component
- **verify-later:** site-engine repo (`$OWNER/site-engine`); go.mod `module site-engine`

### Standalone "probe-go" service (abandoned first cut)
- **category:** system-architecture
- **status-signal:** abandoned
- **status-evidence:** Session 1 "Forked idea.uk's Go service into probe-go … Caveat raised next session: this drifted into a separate project"; Session 2 reframed it as not-a-separate-project.
- **what:** The original framing forked idea.uk's multi-vhost Go service (page-by-Host-header, page.go + domains.json in Go) into a self-contained project. Rejected because it sat too far from the website-building chassis; page.go and domains.json were removed and the engine was trimmed to an API-only backend with content moved to chassis build outputs.
- **sources:** traffic_probe_running_notes(27).md#session-1, traffic_probe_running_notes(27).md#session-2, traffic_probe_running_notes(27).md#session-3
- **relations:** superseded by site-engine + Layer-4/thin-Layer-5 framing
- **verify-later:** n/a (removed page.go/domains.json)

### Layer-4-build + thin-Layer-5-VM-deploy framing
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** Session 2 conclusion "the probe is Layer 4 (build a targeted site) + a thin slice of Layer 5 (deploy a tiny backend to a VM instead of B2)"; decision to keep git→Actions and only swap the target.
- **what:** Rather than a side project, the probe reuses the existing build pipeline (Layer 4) and the git→self-hosted-Actions deploy seam (Layer 5), swapping only the destination from B2 to VM. The heavier chassis service-deployer adapter is the eventual move, not now.
- **sources:** traffic_probe_running_notes(27).md#session-2, traffic_probe_plan(11).md#where-we-are
- **relations:** underlies "commit is deploy" seam swap; defers P5 vmhost adapter
- **verify-later:** CONSOLIDATION_where_it_all_fits.md, PARALLEL_engine_deployment_and_layer5.md

### Phased plan P0–P5
- **category:** system-architecture
- **status-signal:** partial
- **status-evidence:** plan(11) Phases: P0/P1/P2 done, P3 in progress ("Remaining for P3: land the chassis patch…"), "P4 … IN PROGRESS (this chat)", P5 not started.
- **what:** P0 structural decisions; P1 manual go-live (Path A); P2 wire deploy-on-update (two Actions); P3 make the probe a normal pipeline output (github_repo target selection + capture component + capability gate); P4 off-box collection + ranking; P5 registry + provisioning adapter.
- **sources:** traffic_probe_plan(11).md#phases, traffic_probe_running_notes(27).md#open-threads
- **relations:** contains most other concepts; earlier plan versions phrased P4/P5 differently
- **verify-later:** n/a

### VM-Hosted Backend Sites class (proposed doc 024)
- **category:** system-architecture
- **status-signal:** aspirational
- **status-evidence:** plan(11) "Genuinely new (proposed doc 024 'VM-Hosted Backend Sites (site-engine)', Infrastructure Reference; numbering operator's)".
- **what:** The genuinely-new infrastructure: a persistent, non-reaped, internet-facing VM class and its lifecycle; DNS + public TLS as managed state outside k8s; a data-RETURN path from off-cluster; the off-cluster "commit is deploy" seam and where its credential lives (repo secrets now, adapter later); capability-gate semantics. The traffic probe is instance #1 of this class; future chat/board sections join it.
- **sources:** traffic_probe_plan(11).md#framework-integration, traffic_probe_running_notes(27).md#2026-06-11-relojistas-go-live-bundle
- **relations:** class parent of intent-probe; ties to D5 requires-backend gate
- **verify-later:** doc 024 existence; service_instances table

### Capability gate D5 — requires-backend semantic tag
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** plan(11) D5 "Planner gate (to apply): load_components gains AND NOT (… ? 'requires-backend')"; running_notes SQL revised; marked "Outstanding: apply the planner query change".
- **what:** Gates backend-requiring components off static sites by CLASS, not site-type. Component carries `semantic_tags:["requires-backend"]`; planner load_components excludes them unless opted in via roadmap section_types; site side sets `deploy_config || {"target":"vm","capabilities":["backend"]}`; a later audit check compares placed sections' requires-* tags to site capabilities.
- **sources:** traffic_probe_plan(11).md#decision-5, traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight, traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection
- **relations:** supersedes the intent-probe site-type gate
- **verify-later:** build-site-planner workflow JSON load_components query

### Superseded D5 — suitable_site_types / "intent-probe" site type gate
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** plan(4) "Decision 5 (OPEN) … component carries suitable_site_types:['intent-probe'] … planner's load_components gains AND suitable_site_types = '[]'::jsonb"; running_notes "the invented site type is GONE (suitable_site_types: [])".
- **what:** The earlier D5 formulation invented an `intent-probe` site type and gated via `suitable_site_types`. Dropped after operator feedback that "intent-probe is the wrong label" — the distinguishing feature is the class (has a backend), so the tag-based `requires-backend` gate replaced it.
- **sources:** traffic_probe_plan(4).md#decision-5-open, traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight
- **relations:** replaced by requires-backend semantic tag gate (D5)
- **verify-later:** component semantic_tags vs suitable_site_types columns

### intent-probe capture component
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "intent-probe INSERTED into the live library (INSERT 0 1) … second run's INSERT 0 0 is the ON CONFLICT idempotency".
- **what:** A NEW content-library section (after STEP-ZERO found nothing reusable among 83 sections) rendering the invited-action page: no-JS HTML `<form>` POST + 1×1 beacon `<img>`, CSS-var theming, Component Input Schema v2. v1 limit: single text-input action (search/freetext kinds); the {{range}}-based categories variant deferred until the renderer's array handling is verified.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection, traffic_probe_plan(11).md#p3, traffic_probe_running_notes(27).md#2026-06-11-component-live
- **relations:** carries requires-backend tag; hand-instanced for relojistas/wayfaringlondoner
- **verify-later:** content_components row `intent-probe`; intent_probe_component.sql

### github_repo target selection + resolveGitRepoName patch
- **category:** deployment-github
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-10 "sites.github_repo is DORMANT end-to-end … The patch (guide's 'small patch' pattern)"; plan(11) P3 "Remaining … land the chassis patch (resolveGitRepoName helper …)".
- **what:** A site's `sites.github_repo` selects deploy target (vm-sites repo vs default "sites"), but was dormant (upsertSite didn't SELECT it, nothing read it). The fix: one `resolveGitRepoName(config, collected)` helper (config repo_name → site_record.github_repo → "sites") used by both `git_commit` and `deploy_image_asset`, plus upsertSite RETURNING + ensure_site_record map additions. `deploy_image_asset` hardcoded "sites" and would split-brain a probe site (pages→VM, logo/hero→sites) without the same fallback.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection, traffic_probe_running_notes(27).md#2026-06-10-p3-pre-flight, traffic_probe_plan(11).md#p3
- **relations:** enables P3 pipeline wiring; deploy_image_asset split-brain risk
- **verify-later:** git_deployer_actions.go, site_db_actions.go, upsertSite, EnsureSiteRecordAction, deploy_image_asset:463

### "Commit is deploy" seam swapped B2→VM + two GitHub Actions
- **category:** deployment-github
- **status-signal:** deployed
- **status-evidence:** runbook(12) §5 "Two separate workflows"; running_notes 2026-06-11 both siblings "rewritten as faithful siblings … Validated".
- **what:** The static "commit is deploy" seam is preserved, only the destination moves. Content Action (`deploy-to-vm.yml` in vm-sites repo): on push, rsync -az --delete each changed root-level `<domain>/` over SSH to `/var/www/vm-sites/<domain>`; self-hosted runner, no CF purge. Engine Action (`deploy-engine-to-vm.yml` in site-engine repo): on push to `**.go`/go.mod, build static stripped linux/amd64, scp, run the narrow `site-engine-deploy` sudo hook (atomic swap + restart). Secrets VM_HOST/VM_USER/VM_SSH_KEY.
- **sources:** traffic_probe_runbook(12).md#5, traffic_probe_running_notes(27).md#2026-06-10-vm-deploy-action, traffic_probe_running_notes(27).md#2026-06-11-live-b2-action
- **relations:** mirrors live deploy-to-b2.yml + Cloudflare Worker; target-agnostic terminal build item
- **verify-later:** vm-sites/.github/workflows/deploy-to-vm.yml, site-engine/.github/workflows/deploy-engine-to-vm.yml

### setup.sh box provisioning script
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "Provisioning succeeded end-to-end … engine ACTIVE, unit + deploy hook + prune timer installed, nginx OK"; runbook(12) §3.5.
- **what:** Multi-vhost box installer: per-domain nginx server_name blocks (serve static webroot + proxy the four API paths), per-domain webroot certbot (graceful HTTP fallback, idempotent re-run upgrades to HTTPS), systemd unit, ufw/fail2ban/logrotate/unattended-upgrades/ssh-hardening, the deploy sudo hook, and `site-engine-prune.timer`. Params are env-vars: DOMAINS, LETSENCRYPT_EMAIL, DEPLOY_USER, ENGINE_BINARY_PATH, WEBROOT_OWNER, WWW_ALIAS, RETENTION_DAYS, CLOUDFLARE, MODE=full/update. Add a domain = extend DOMAINS + re-run (idempotent).
- **sources:** traffic_probe_runbook(12).md#3.5, traffic_probe_running_notes(27).md#2026-06-10-box-setup-artifact, traffic_probe_running_notes(27).md#2026-06-13-g
- **relations:** origin of the deploy privilege model; WWW_ALIAS/CLOUDFLARE/RETENTION shipped incrementally
- **verify-later:** deploy_setup/vm-deploy/setup.sh (live tree)

### Deploy privilege model (site-engine-deploy sudo hook)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-10 "Privilege model (low-risk): no root key in CI … a sudoers rule scoped to ONLY that script".
- **what:** No root key in CI. When `DEPLOY_USER` is set, setup.sh installs `/usr/local/sbin/site-engine-deploy` (root-owned: atomic binary swap + restart) and a sudoers rule scoped to only that script. The deploy user can swap the engine and nothing else; the swapped binary runs as the unprivileged `site-engine` user.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-engine-deploy-workflow, traffic_probe_runbook(12).md#5
- **relations:** part of engine Action; retired later by P5 adapter holding the SSH credential
- **verify-later:** setup.sh DEPLOY_USER branch; sudoers site-engine-deploy

### JSON store scaling evolution (whole-file → dirty-flusher → daily JSONL)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "Store scaling fix (structural, pre-launch)" then "Store v2 (JSONL) … v2: events append to daily JSONL … O(1) at any volume"; burst-tested.
- **what:** v0 rewrote the entire ever-growing JSON file on every beacon hit (linear write cliff). v1 added a dirty-flag + 5s background flusher (AddVisit no longer persists per call; AddEvent still immediate). v2 replaced the monolithic file: events append to daily `events-YYYYMMDD.jsonl` (one line per submission, bounded RAM), /stats counters live in a small `counters.json`; SIGTERM fsyncs. Removed the in-RAM `Store.Events` map and uncalled `Store.Snapshot()`.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-live-b2-action, traffic_probe_running_notes(27).md#2026-06-11-store-v2, deploy_setup/working_dir/main.go#header
- **relations:** abandoned Store.Events/Snapshot; drove the ENGINE_DATA_DIR rename
- **verify-later:** store.go Flush/flushLoop/EventCounts/openEventsFileLocked; /var/lib/site-engine/{events-*.jsonl,counters.json}

### Class-level rename (probe → site-engine) and env-var churn
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** running_notes 2026-06-11 "RENAME MAP (every changed name)"; env var `PROBE_DB_PATH → ENGINE_DB_PATH`; then 2026-06-11 store-v2 "env var ENGINE_DB_PATH → ENGINE_DATA_DIR".
- **what:** When the box became the home of the whole backend-site class (not just probes), "probe" defaults were neutralised to class-level names across engine + deploy artifacts: service/user `site-engine`, `/opt/site-engine`, `/var/lib/site-engine`, `/etc/site-engine/site-engine.env`, webroots `/var/www/vm-sites/<d>`, rate zone `engine_rl`, hook `site-engine-deploy`. The DB-path env var was renamed twice: `PROBE_DB_PATH` → `ENGINE_DB_PATH` → `ENGINE_DATA_DIR`; store file `probe_events.json` → `intent_events.json` → (dropped for JSONL). ProbeSearch/ProbeCategory/ProbeFreeText kind constants kept (they name the feature, not the class).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-component-live, traffic_probe_running_notes(27).md#2026-06-11-store-v2, traffic_probe_runbook(12).md#changelog
- **relations:** supersedes probe-go naming
- **verify-later:** grep for stale PROBE_/probe_events across artifacts

### Visit beacon + events-per-1k ranking metric
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** service(24).go / main.go headers describe `GET /api/hit` 1×1 beacon; runbook(12) §6 "Metric: intent events per 1,000 visits"; ranking query 1 LEFT JOINs for events-per-1k.
- **what:** A no-JS/no-cookie 1×1 image (`<img src="/api/hit">`) on the page counts human visits as the denominator for an "intent events per 1,000 visits" ranking metric. Visits live in the engine's counters.json (/stats), not in intent_events, so the rate metric requires joining the intent_site_stats snapshot. The gracias/thanks page deliberately omits the beacon (would inflate the denominator).
- **sources:** deploy_setup/working_dir/service(24).go#header, traffic_probe_runbook(12).md#6, traffic_probe_running_notes(27).md#2026-06-13-e
- **relations:** feeds intent_ranking_queries; depends on intent_site_stats
- **verify-later:** counters.json; /stats visit counter; intent_ranking_queries.sql query 1

### Input sanitisation (sanitizeValue, Cc/Cf stripping, NFD survives)
- **category:** content-quality
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "Sanitisation v2 … now strips Cc AND Cf … Real bug found by the new tests: checking IsControl before IsSpace silently JOINED words".
- **what:** The engine's `sanitizeValue()` strips control (Cc) and format (Cf: zero-widths, bidi overrides incl. U+202E, BOM, soft hyphen) chars, collapses whitespace runs (IsSpace checked FIRST to avoid joining words like `gmt\t\tmaster`→`gmtmaster`), and caps by RUNES not bytes (MaxValueLen semantic changed). NFD combining marks deliberately survive; NFC normalisation + lowercasing are deferred to the P4 collector (needs x/text; engine stays stdlib-only).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_running_notes(27).md#2026-06-12-debug-guide, traffic_probe_plan(11).md#p4
- **relations:** pairs with P4 ingest validation contract (NFC there)
- **verify-later:** service.go sanitizeValue; MAX_VALUE_LEN handling

### /stats endpoint + INTERNAL_API_KEY (stats internal key)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** service(24).go "GET /stats key-gated per-host summary"; runbook(12) env table "INTERNAL_API_KEY gates /stats (sent as X-Internal-Key) … unset → /stats 401"; verified over HTTPS 2026-06-12.
- **what:** `/stats` returns a key-gated per-host summary (visits/events counters), gated by `INTERNAL_API_KEY` sent as header `X-Internal-Key`. Unset key → /stats returns 401. The same key doubles as the read-only capture-export key for /events and /access-digest; on the collector side it is stored in `deploy_config.engine.stats_key` (low sensitivity, one accessor, movable to a secrets table later). The env file (not a shell variable) is the source of truth.
- **sources:** deploy_setup/working_dir/service(24).go#header, traffic_probe_runbook(12).md#2, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** read by CollectIntentEventsAction
- **verify-later:** /etc/site-engine/site-engine.env INTERNAL_API_KEY; deploy_config.engine.stats_key

### /events export endpoint (P4 collector interface)
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "GET /events built + tested … Tests green ×6"; runbook(12) §6 "Export endpoint (P4 collector interface)".
- **what:** `GET /events` streams stored events as key-gated NDJSON oldest-first (original line bytes preserved); params `since` (RFC3339, strictly-after), `host`, `limit` (default 5000). Final line `{"_meta":{count,truncated,server_time}}` aids the collector checkpoint (store max created_at → duplicate-free pulls). Lock-free by design so a big export never blocks live captures; a torn mid-append tail line is skipped and arrives next pull.
- **sources:** traffic_probe_running_notes(27).md#2026-06-12-events-export, traffic_probe_runbook(12).md#6
- **relations:** consumed by CollectIntentEventsAction; the pull architecture
- **verify-later:** store.go StreamEvents; nginx /events location

### /access-digest endpoint (passive nginx-log harvest)
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** accessdigest(1).go header "parse this box's nginx combined access log into a compact, key-gated digest"; running_notes 2026-06-13(g) "/access-digest endpoint BUILT + tested … Builds clean".
- **what:** `GET /access-digest?host=&since=&top=` returns a key-gated JSON rollup of one domain's nginx combined access log: status mix, top referers (canonicalHost-reduced), top paths, top 404 paths, UA buckets, top REAL client IPs. Captures the referer/landing-path/404-intent/UA signals the engine can't see on a static page load. Needs per-domain logs + engine in the `adm` group (both from setup.sh); needs `CLOUDFLARE=true` (nginx real_ip) on proxied boxes so IPs are the real client, not Cloudflare's.
- **sources:** deploy_setup/working_dir/accessdigest(1).go#header, traffic_probe_running_notes(27).md#2026-06-13-g, traffic_probe_runbook(12).md#6
- **relations:** implements passive_harvest_spec Option A part 2; shares source with Thread-D bot blocklist
- **verify-later:** accessdigest.go buildAccessDigest/classifyUA/safeHost; NGINX_LOG_DIR config (main(19).go)

### P4 off-box collection (intent_events + CollectIntentEventsAction)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(d) "Migration applied (CREATE TABLE + 3 indexes + INSERT 0 1 task)"; action "VERIFIED against live source + registered"; but agent deploy fields still to confirm and enable order pending.
- **what:** The cluster pulls intent over key-gated HTTPS with NO adapter and NO SSH. `intent_events` table (engine_event_id UNIQUE = structural idempotency, CHECK on kind/value len, host→site_id resolve, checkpoint = max(event_created_at) with no extra storage). `collect_intent_events` is a SINGLE Go action that self-queries all VM backend sites and loops (parameterised upserts), registered in GlobalActionRegistry (Category "data", IsLocal). Ingest contract: parameterised SQL only, per-line shape checks, burst dedupe, NFC normalisation + lowercasing here.
- **sources:** traffic_probe_plan(11).md#p4, traffic_probe_running_notes(27).md#2026-06-13-b, traffic_probe_running_notes(27).md#2026-06-13-c
- **relations:** driven by intent-collection-orchestrator/intent-collector agents; extended with collectSiteStats + access-digest pull
- **verify-later:** intent_events_migration.sql; intent_collector_actions.go; registry.go DATA region

### Superseded checkpoint-JSON / events-per-1k ranking (early P4)
- **category:** scheduler-and-tasks
- **status-signal:** superseded
- **status-evidence:** plan(1)/(4)/(5) P4 "checkpoint JSON, compute events-per-1k, rank domains"; plan(11) now "idempotent via unique engine_event_id; no extra checkpoint storage — since=max(event_created_at)".
- **what:** Early P4 phrasing planned an explicit checkpoint-JSON file to track collection progress and a direct events-per-1k rank. Dropped in favour of structural idempotency (unique engine_event_id) with the checkpoint derived as since=max(event_created_at) — no extra checkpoint storage. Ranking became a set of read-only SQL queries.
- **sources:** traffic_probe_plan(4).md#phases, traffic_probe_plan(1).md#phases, traffic_probe_plan(11).md#p4
- **relations:** replaced by intent_events unique-id design + intent_ranking_queries
- **verify-later:** intent_events.engine_event_id UNIQUE

### intent-collection-orchestrator + intent-collector agents
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** intent_collector_agents SQL headers "intent-collection-orchestrator + intent-collector (P4) … mirror the LIVE med-export-orchestrator / med-json-exporter pair verbatim"; running_notes 2026-06-13(g) INSERT bug fixed.
- **what:** A thin wrapper-orchestrator (spawn_collector → call_collector → complete, no substantive in-chassis work) that spawns the `intent-collector` task worker (one step: collect_intent_events, processing_mode "task"). Infra fields (image docker.io/aqls/agent-chassis v1.0.1063, resources, health_config, business-intel topics, delegation) copied verbatim from the med-export pair. Reached by the scheduler via target_topic=system.agent.generic.requests by agent_type. Idempotency uses `ON CONFLICT (type, version)`.
- **sources:** deploy_setup/working_dir/intent_collector_agents(3).sql#header, deploy_setup/working_dir/intent_collector_agents(1).sql, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** identical to live intent_collector_agents(2).sql; wrapper-orchestrator requirement; replaces a single in-pod collector
- **verify-later:** agent_definitions rows intent-collection-orchestrator / intent-collector

### Wrapper-orchestrator requirement finding (001:405-462)
- **category:** contracts-and-standards
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(d) "STRUCTURAL finding … wrapper-orchestrator REQUIRED (001:405-462): the collector is reached via the SCHEDULER … AND does substantive in-chassis work → must NOT run in a shared agent-chassis pod".
- **what:** Because the collector is reached via the generic scheduler entry point AND does substantive work (HTTP to N boxes + multi-row upserts, unbounded as boxes grow), it must not run in a shared agent-chassis pod. Fix = a thin orchestrator that spawns a worker child into its own pod. Also corrected: the scheduler fires ONE message per tick and does NOT fan out pre_query rows, so the collector self-queries and loops, and the scheduled_tasks pre_query is a count>0 GATE.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-d, traffic_probe_running_notes(27).md#2026-06-13-c, traffic_probe_running_notes(27).md#standing-observations
- **relations:** shaped the two-agent topology; scheduled_tasks intent-collection target corrected to orchestrator
- **verify-later:** 001 guide §405-462; scheduled_tasks intent-collection target_topic

### backend_unreachable discovery check
- **category:** diagnosis-loop
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(f) "backend_unreachable REWRITTEN against the real DiscoveryCheck interface … Run(dctx DiscoveryCheckContext)(*CheckResult,error) … gofmt-clean"; enable pending.
- **what:** A discovery_checks check that NOOPs unless deploy_config.target='vm', GETs each backend site's public `/health`, and on failure emits a site_work_items row (source='discovery', item_type='backend_unreachable', item_key for dedup). Self-clearing. Alert-only: HandlerAgent "" because a down VM isn't chassis-fixable (the P5 vmhost adapter becomes the handler later). A `missing_beacon` check was floated too.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-e, traffic_probe_running_notes(27).md#2026-06-13-f, traffic_probe_plan(11).md#p4
- **relations:** ties to P5 vmhost adapter as future handler
- **verify-later:** discovery_checks/check_backend_unreachable.go; site_work_items idx_swi_dedup

### intent_site_stats + intent_ranking_queries
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(f) "Option A part 1 (visits) BUILT: intent_site_stats table … ranking query 1 LEFT JOINs for events-per-1k"; 2026-06-13(e) "intent_ranking_queries.sql — 6 read-only queries".
- **what:** `intent_site_stats` stores the latest /stats snapshot per host (PK host); the collector's collectSiteStats pulls /stats and upserts (non-fatal). `intent_ranking_queries.sql` is 6 read-only queries over intent_events: per-domain summary, top terms, dominant-cluster share (the graduation signal), referer breakdown, landing-query breakdown, recent raw submissions — working today on absolute signal, with events-per-1k once visits join.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-e, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** consumes /stats; ranking is the domain-graduation decision input
- **verify-later:** intent_site_stats_migration.sql; intent_ranking_queries.sql

### passive_harvest_spec (3 options, A recommended)
- **category:** traffic-analytics
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(e) "passive_harvest_spec.md lays out 3 options … RECOMMENDS A … DECISION NEEDED from operator before building"; parts built in (f)/(g).
- **what:** Spec for getting the visit rate + passive signals (referer/404/UA, which live in nginx's combined log, not visible to the engine on static loads). Option A: engine reads its own box's nginx log + /stats → key-gated digest, preserving the pull model (new intent_site_stats table + /access-digest). Option B: defer to the P5 vmhost SSH adapter. Option C: Cloudflare analytics if proxied. A was chosen and built.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-e, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** realised by /access-digest + intent_site_stats
- **verify-later:** passive_harvest_spec.md options A/B/C

### landing_query enrichment on IntentEvent
- **category:** traffic-analytics
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-13 "Small legitimate engine enrichment shipped: landing_query field on IntentEvent … Tested … Additive, no breaking change".
- **what:** IntentEvent gained a `landing_query` field populated from the submission's Referer query (the inbound ?q=/?utm= that survives into the form page), so the structured /events export carries inbound-query intent without a log-join. omitempty when absent; external ref_host still recorded separately.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict
- **relations:** complements the access-log harvest (referer host)
- **verify-later:** service.go IntentEvent.LandingQuery / landingQuery() helper

### Pull architecture / no collector VM
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "No third 'collector' VM: the serving box buffers (JSONL); the CLUSTER pulls over key-gated HTTPS … Pull keeps every credential in the cluster — boxes never hold DB/cluster secrets".
- **what:** Collection is pull-only: the serving box buffers events in JSONL and exposes them via key-gated HTTPS (/events, /access-digest, /stats); the cluster's scheduled collector pulls. No third collector VM and no push, because push or a middle VM would put DB/cluster secrets on the box and add attack surface + a hop for no gain. B2 remains an optional cold backup.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-store-v2, traffic_probe_runbook(12).md#6
- **relations:** rationale for /events + collector design; boxes hold only the read-only stats_key
- **verify-later:** n/a

### VM sizing / Hetzner box selection
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "VM sizing (relojistas, its own box): Hetzner CX22-class"; 2026-06-12 "Box provisioned: Hetzner CPX22 #140056673, nbg1 … IP 167.233.33.159, €11.39/mo".
- **what:** Boxes are sized by disk/log headroom not CPU (static nginx + O(1) JSONL appends are far inside a small box even at claimed 1.2M visits/mo). EU-only Hetzner (20 TB/mo included) — runbook standardises on CX23 (~€3.49/mo). x86-only caveat: the engine Action builds GOARCH=amd64, so Arm (CAX) would need a build-matrix change. relojistas has its own box; small domains share a multi-vhost box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-store-v2, traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_running_notes(27).md#2026-06-12-operator-execution, traffic_probe_runbook(12).md#3.2
- **relations:** dedicated-box vs shared-box decision
- **verify-later:** relojistas_notes coordinates; Hetzner CPX22 #140056673

### relojistas.com go-live + bot verdict
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-13 "Relojistas VERDICT from the access log: 14,961 reqs, 83% 404s … Human intent ≈ 0. Clean probe result (domain not worth building), not a measurement failure".
- **what:** First live domain: a Spanish watch FORUM (grounded in the Wayback snapshot), hand-made `relojistas-site/` (index + gracias, kind=search, THANKS_PATH=/gracias.html) to unblock go-live. After going live (and later Cloudflare-proxied), the access log showed overwhelmingly bot/crawler traffic (Chrome-spoof crawler, Claude-SearchBot, Semrush, Yandex) with ~0 human intent — a clean negative probe result. Later decided to static-build it anyway (RSS + crawler presence + 404/referer signal are assets).
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-relojistas-go-live-bundle, traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** hand-instance of intent-probe component; drove the passive access-log harvest decision
- **verify-later:** deploy_setup/relojistas-site/{index,gracias}.html; relojistas_notes/relojistas_golive.md

### wayfaringlondoner.com page + THANKS_PATH-is-engine-wide
- **category:** site-case-studies
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13 "wayfaringlondoner.com page built … a 2015–16 travel blog … BLOG framing"; "Design point — THANKS_PATH is engine-wide".
- **what:** Second hand-made page: a 2015–16 travel blog (Csilla; London + Bangkok/Transylvania/Jersey), BLOG framing asking for a destination/London spot/story, tagline gained "and under new ownership". Targets the SHARED multi-vhost box. Surfaced the constraint that THANKS_PATH is one engine-wide env var, so domains on a shared box must share a thanks filename (standard `/thanks.html`); relojistas keeps `/gracias.html` on its own box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict, traffic_probe_running_notes(27).md#2026-06-13-b
- **relations:** shared-box multi-domain onboarding
- **verify-later:** wayfaringlondoner-site/; wayfaringlondoner_notes.md (live)

### Original first-domain set (dropped surgerylight + finance/retail)
- **category:** site-case-studies
- **status-signal:** abandoned
- **status-evidence:** runbook base §3 "Suggested first set: relojistas.com, wayfaringlondoner.com, surgerylight.com, plus one finance tool and one clear retail" — absent from runbook(12) §3 which names only relojistas.
- **what:** The earliest runbook proposed a concrete 3–5 domain starter set (relojistas.com, wayfaringlondoner.com, surgerylight.com, plus a finance tool and a clear retail), each grounded via Wayback. Later versions dropped the named list down to relojistas + wayfaringlondoner; surgerylight and the finance/retail candidates silently vanished.
- **sources:** traffic_probe_runbook.md#3, traffic_probe_runbook(2).md#3, traffic_probe_plan(11).md#risks
- **relations:** relates to Wayback grounding method
- **verify-later:** n/a

### Cloudflare-in-front option
- **category:** multicluster
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-13(f) "Cloudflare: relojistas now PROXIED (operator data: 22,046 SSL reqs/24h, 4,416 attacks blocked)"; runbook(12) §8.
- **what:** Optional proxied (orange-cloud) Cloudflare record → VM origin (a reverse proxy, NOT a second Worker/copy). Adjustments: cache-bypass the API paths; set nginx `set_real_ip_from`/`real_ip_header CF-Connecting-IP` (else rate-limit throttles all CF IPs as one, and logs/digest show CF IPs); TLS Full(strict); bonus CF-IPCountry + instant relocation. setup.sh `CLOUDFLARE=true` writes cloudflare-realip.conf.
- **sources:** traffic_probe_runbook(12).md#8, traffic_probe_running_notes(27).md#2026-06-10-engine-deploy-workflow, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** required for real client IPs in /access-digest + Thread-D blocklist
- **verify-later:** /etc/nginx/conf.d/cloudflare-realip.conf; setup.sh CLOUDFLARE branch

### Privacy posture (no cookies/JS/IP; UK GDPR/PECR)
- **category:** content-governance
- **status-signal:** deployed
- **status-evidence:** running_notes standing observations "Privacy posture (UK GDPR/PECR, low risk appetite): no cookies, no JS, no IP stored, referer reduced to host, country only from a coarse CDN header".
- **what:** A deliberate low-risk privacy stance baked into the engine and page: no cookies, no JavaScript, no IP stored, referer reduced to host only, country only from a coarse CDN header (CF-IPCountry). Open ingest choice: redact email/phone patterns at ingest vs rely on the 90-day prune.
- **sources:** traffic_probe_running_notes(27).md#standing-observations, traffic_probe_plan(11).md#p4, traffic_probe_runbook(12).md#6
- **relations:** shapes intent-probe (no-JS form), sanitisation, retention timer
- **verify-later:** engine handlers (no cookie/IP); site-engine-prune.timer RETENTION_DAYS=90

### Retention prune timer
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "Added to setup.sh: RETENTION_DAYS param (default 90) + site-engine-prune.service/.timer (daily find-delete of old events-*.jsonl)".
- **what:** Because daily JSONL IS the rotation, logrotate on engine files would race the open handle; instead setup.sh installs a `site-engine-prune` systemd service+timer that daily find-deletes `events-*.jsonl` older than RETENTION_DAYS (default 90). nginx logs keep their existing size-based logrotate.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_runbook(12).md#3.5
- **relations:** part of the privacy posture
- **verify-later:** setup.sh site-engine-prune.timer; RETENTION_DAYS

### P5 vmhost provisioning adapter (superseded service-deployer)
- **category:** adapters
- **status-signal:** aspirational
- **status-evidence:** plan(11) P5 "A vmhost adapter for what DOES need SSH … built to the analyser-adapter README skeleton"; earlier plan(1)/(4)/(5) P5 "registry + relocation (service_instances) and, eventually, the chassis service-deployer adapter".
- **what:** The eventual automation for what genuinely needs SSH: provision box, run setup.sh, onboard domain, ship engine, decommission — built as a `vmhost` adapter (cmd/vmhost-adapter, internal/adapters/vmhost, reuse thunder's ssh via shared/, kustomize, KafkaTopic system.adapter.vmhost.requests, 003 envelope), with a `service_instances` registry modelled on thunder_instances minus the reaper. Long-term it holds the deploy SSH credential, retiring the repo-secrets copy. Earlier versions named this the "service-deployer" adapter.
- **sources:** traffic_probe_plan(11).md#p5, traffic_probe_plan(1).md#phases, traffic_probe_running_notes(27).md#open-threads
- **relations:** future handler for backend_unreachable; supersedes "service-deployer" naming
- **verify-later:** analyser-adapter README; thunder_instances → service_instances

### Adapter Response Envelope Contract (003) — conditional, later dropped from plan
- **category:** contracts-and-standards
- **status-signal:** superseded
- **status-evidence:** plan(4)/(5)/(6) P4/P5 "If collection runs as a chassis adapter, it MUST follow the Adapter Response Envelope Contract (typed-struct bool headers, reuse request_id, message_id, ProduceWithValidation)"; plan(11) reduces this to a one-line parenthetical.
- **what:** The guidelines-audit flagged that IF P4 collection or the P5 deployer were built as chassis adapters, replies must use a typed-struct envelope — getting it wrong = silent drop until timeout (the documented multi-day thunder fault). Once P4 was redesigned to need no adapter (key-gated HTTPS pull + one local action), the prominent P4/P5 envelope warnings were demoted.
- **sources:** traffic_probe_plan(6).md#phases, traffic_probe_running_notes(27).md#2026-06-10-guidelines-audit, traffic_probe_plan(11).md#p4
- **relations:** applies only if P5 vmhost adapter is built
- **verify-later:** 003 contracts doc; ProduceWithValidation

### Guidelines audit (001/002/003 compliance)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-10 "Read the dev guide, architecture, and contracts. Existing code: no violations"; 2026-06-13(d) action "audited against 001/002/003 — code is COMPLIANT".
- **what:** Recurring audits confirming the engine and collector honour the house rules: engine is standalone package main; the no-JS HTML form satisfies JS Content Separation; parameterised SQL only; no logger.Debug; kebab-case/snake_case names; private same-file helpers are allowed; stats_key never logged.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-guidelines-audit, traffic_probe_running_notes(27).md#2026-06-13-d
- **relations:** produced the wrapper-orchestrator finding and the envelope-contract flag
- **verify-later:** 001 dev guide; 002 architecture; 003 contracts

### Probe debugging-guide entries #24–#28
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "Debug guide updated … 016_debugging_guide_v2_46.md"; 2026-06-13(g) "Debug guide v2_48 … #27 invented-interface … #28 agent_definitions UNIQUE(type,version)".
- **what:** Reusable pitfalls harvested from probe execution: #24 a config/workflow file is only authoritative at its runtime read-path; #25 prove the test harness delivered the intended input before debugging; #26 shell vars need export/prefix and die with the session; #27 invented interface (compiles standalone ≠ satisfies interface — wire to registry early); #28 agent_definitions UNIQUE(type,version) + two look-alike category columns.
- **sources:** traffic_probe_running_notes(27).md#2026-06-12-debug-guide, traffic_probe_running_notes(27).md#2026-06-12-provisioning-ran, traffic_probe_running_notes(27).md#2026-06-13-g
- **relations:** #24 is the stale-artifact class; #27 fixed backend_unreachable; #28 fixed the agent INSERT
- **verify-later:** 016_debugging_guide_v2_48.md entries #24–#28

### Wayback/archive.org grounding method + limitation
- **category:** research-agents
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(b) "archive.org: Claude CAN web_fetch archive pages but ONLY when a search surfaces the exact URL; canNOT enumerate CDX on demand and the sandbox can't reach archive.org directly".
- **what:** Each probe page is grounded in the domain's old vertical via a Wayback snapshot. Constraint: the sandbox can't reach archive.org directly and can't enumerate CDX on demand, so grounding a NEW domain requires the operator to supply the Wayback URL/snapshot, or Claude uses web search + the domain name.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-b, traffic_probe_runbook.md#3
- **relations:** feeds intent-probe page content
- **verify-later:** archive.org.results/{relojistas,wayfaringlondoner}

### HANDOFF permanent-thread scope split (Threads A–D)
- **category:** documentation-system
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(b) "everything else → HANDOFF_vm_sites_permanent_thread.md (Threads A manifest / B framework integration / C more domains / D global bot blocklist)".
- **what:** Work was split so P4 collection stayed active while the rest handed off to a permanent thread: Thread A = static-build relojistas as a manifest→framework build; Thread B = framework integration (a backend site becomes a normal multi-page chassis build); Thread C = more domains on existing boxes; Thread D = a global bot-IP blocklist sharing the access-log digest source.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-b, traffic_probe_running_notes(27).md#2026-06-13-relojistas-verdict
- **relations:** Thread D shares /access-digest source
- **verify-later:** HANDOFF_vm_sites_permanent_thread.md (live)

### http2 deprecation fix at the generator
- **category:** styling-render-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "nginx 1.28.3 warns on `listen ... http2` (deprecated since 1.25) … the generator now emits version-neutral `listen 443 ssl;`".
- **what:** A field finding: newer nginx deprecates `listen ... http2` while the local container lacks the replacement `http2 on;`, so setup.sh's conf generator emits version-neutral `listen 443 ssl;` (with an opt-in comment for ≥1.25.1). Caught by fixing at the generator rather than per-box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-12-cert-issued
- **relations:** part of setup.sh nginx conf generation
- **verify-later:** setup.sh nginx server block listen directive

## Additional notes
The four `.go` entrypoint files are a version family: `main.go` ≡ `main(15).go` ≡ `main(17).go` (no NginxLogDir); `main(19).go` adds `NginxLogDir` config — the change that enabled the /access-digest endpoint. `env.go` header states its tiny env/envInt helpers are "copied verbatim from idea.uk's engine.go" — the trace of the idea.uk fork lineage. The two archived SQL files are byte-identical to each other and to live `intent_collector_agents(2).sql` (only cross-version change was `ON CONFLICT (type)` → `ON CONFLICT (type, version)`, i.e. debug-guide #28). This unit overlaps U11 (traffic_probe, the live tree) — consolidation should de-duplicate against U11.
