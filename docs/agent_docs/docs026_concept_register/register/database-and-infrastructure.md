# Register — database-and-infrastructure

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

24 concepts, consolidated from 34 raw extractions across units U01, U02, U03, U04,
U08, U10, U12, U13, U15, U17a, U17b, U18, U19, U20, U21, U23, U24b, U24c, U24f, U25,
U26.
(Note: as with the system-architecture half of this cluster, the assigned input
file's entire content appeared mechanically duplicated verbatim, twice, back-to-back
— a bucketing artifact, not independent evidence. The 34 counted here are the unique
raw blocks after collapsing that duplication.)

### DBI-001 — Snapshot-before-mutate discipline (backup naming, snapshot_agent/take_site_snapshot, revert)
- **status:** deployed
- **status-evidence:** Practised and re-stated across every era of the doc history — the 009 backup/swap/revert reference, idea.uk's three real migrations (classifier applied 2026-06-20), and a standing rule reinforced repeatedly in the diagnosis-loop running notes ("Snapshot before any agent_definitions change... standing rule, turn 11").
- **what:** The safe way to touch live `agent_definitions` rows (and sites) in SQL: call `snapshot_agent('<type>','<reason>')` (paired with `revert_agent`) inside the same transaction as the change — snapshots insert at version+1000 with `is_snapshot` true, excluded from runtime selection — before any UPDATE; name backups for the migration they guard (`agent_definitions_backup_YYYYMMDD_pre<NNN>`) and never `DROP TABLE IF EXISTS` before `CREATE` (a name collision on re-run is the safety mechanism working — pick a new suffix, don't destroy the old backup). Companion techniques: exact-anchor `replace()` with a self-check that RAISEs and rolls back if the anchor is missing; idempotency sentinels so re-runs no-op; single-line anchors only (multi-line anchors break on whitespace); run migration FILES via `kubectl … < file` (pasting into psql mangles `\set`/`\echo`/blank lines); check `\df`/`\dx` for an existing DB helper or extension before hand-writing SQL. A parallel site-level function, `take_site_snapshot()`, applies the same doctrine to site rows.
- **sources:** 009#Operational Reference: Backup, Swap, and Revert; idea.uk/migration_domain_research_classifier_structured_design_intent.sql; idea.uk/HANDOFF(13).md; NOTES_running_synthesis_principles(59); contextkit/001_more_potential_thin_slice_prompt.md; docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql#BEFORE-APPLYING
- **relations:** migrations ledger system; re-running-migrations hazard; SQL change-management
- **verify-later:** snapshot_agent / take_site_snapshot function signatures against the live clients_db schema; bak_* conventions

### DBI-002 — Migrations ledger system (schema_migrations table + guarded run-migrations.sh runner)
- **status:** deployed
- **status-evidence:** "Migrations system live (2026-07-10)"; migration 140 was "the system's first real apply"; applied through 146 at unit close; the same system is independently described from the sql_for_agents directory's own numbering.
- **what:** A `schema_migrations` table (filename PK, applied_at, md5 checksum, applied_by, notes) plus `scripts/migration/run-migrations.sh` (dry-run default, `--apply`, per-file re-check, stops-without-recording on failure, loud warning for near-miss filenames). Baseline migration 124 marks files 001–123 as pre-system history, never auto-applied. `snapshot_agent(type, reason)` is the standing rule opening every agent-updating transaction inside this regime; `migration_backups` holds manual before-values; the no-DROP rule from DBI-001 is restated as "the collision IS the safety net." Workflow-altering migrations must leave a pipeline doc_note. One migration (128) is an honest reconstruction stub for a lost file, verified live with a NULL checksum.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#2026-07-10-migrations; RUNBOOK_travelling_docs(38).md#task-5-closed; 124_schema_migrations.sql; 107_image_build_handler.sql; 131_tool_generator_plan_writing.sql
- **relations:** snapshot-before-mutate discipline; versioned agent_definitions (UNIQUE on type+version)
- **verify-later:** schema_migrations rows; run-migrations.sh behaviour; the 128 stub

### DBI-003 — API keys logged in plaintext — exposure & rotation (STABILITY_API_KEY, BANANA_API_KEY)
- **status:** deployed
- **status-evidence:** Flagged "SECURITY — STILL OPEN, STILL HIGHEST PRIORITY" (2026-05-20), carried across handoffs for seven weeks, then "✅ DONE 2026-07-08. Keys rotated (user confirmed)."
- **what:** STABILITY_API_KEY and BANANA_API_KEY were logged in plaintext at info level (adapter init env-dump `zap.Any`/`zap.String("apiKey")`, plus B2 keys in NewS3Client debug logs), with billing exposure on the paid Banana tier. Resolved by scrubbing to scoped log fields and rotating both keys.
- **sources:** TODO_imagery_followups.md#SECURITY; HANDOFF_robot_hands_rebuild.md#Carried-forward; RUNBOOK_imagery_best_in_class.md#B1; HANDOFF_2026-05-26…md#other-open-items
- **relations:** provider abstraction (the code carrying the logging); credentials handling conventions
- **verify-later:** dynamic_adapter.go logging fields; confirm no raw keys in current logs

### DBI-004 — Three-database architecture (MySQL auth + PostgreSQL clients + PostgreSQL templates)
- **status:** deployed
- **status-evidence:** Documented consistently from the earliest architecture summary through to the current docs024 operational reference (011), which adds live pgbouncer/connection detail.
- **what:** Authentication is isolated in an external MySQL DB (Clook cPanel, Remote-MySQL IP whitelist) holding users, JWT refresh tokens, profiles, subscriptions/tiers, permissions, activity logs (BINARY(16) UUIDs); agent/AI runtime lives in an in-cluster PostgreSQL 16 `clients_db` (global agent_definitions, orchestrator_state, clients_info + per-client schemas) reached via pgbouncer:6432 in transaction mode; a second PostgreSQL `templates_db` holds shared persona templates. Transaction-mode pgbouncer forbids prepared statements/session state — connection strings need simple_protocol/cache_describe, and `pg_dump`/`LISTEN`/`NOTIFY` must bypass pgbouncer entirely. Go driver split: chassis still uses pgxpool, core-manager uses database/sql. All credentials live in the `personae-platform-secrets` K8s secret (Terraform-managed).
- **sources:** docs024_key_docs_latest/011_database_and_infrastructure.md (full); docs/architecture/databases.md; docs/basic_usage/001basic_usage.txt
- **relations:** schema-per-client multi-tenancy; database password rotation runbook; admin auth flow
- **verify-later:** which binaries still use pgxpool; current DB inventory; whether templates_db still exists separately

### DBI-005 — Schema-per-client multi-tenancy (create_client_schema)
- **status:** deployed
- **status-evidence:** Documented as implemented from the earliest `databases.md` summary through to the current docs024 runbook (011), which lists live clients (demo_client, vetcomparison, test_client, system).
- **what:** Each client gets an isolated PostgreSQL schema `client_<id>` containing agent_instances, agent_spawn_history, projects (+optionally website_projects/agent_memory pgvector embeddings/workflow_executions/usage_analytics), created by the `create_client_schema()` SQL function; global resources (agent_definitions, templates, orchestrator_state) stay shared. `spawn_agent` resolves the target schema from the `client_id` Kafka header (the validator only checks presence, no DB lookup) and inserts exact columns — manual table creation must match spawn_actions.go's INSERT or spawning fails.
- **sources:** docs024_key_docs_latest/011_database_and_infrastructure.md#Creating a New Client Schema; docs/architecture/databases.md#2-postgresql-database-1; docs/basic_usage/001basic_usage.txt
- **relations:** three-database architecture; multicluster/multitenancy; agent_instances column-shape drift; scheduler client_id header requirement
- **verify-later:** create_client_schema function source; pgvector agent_memory usage in current code

### DBI-006 — agent_instances / agent_spawn_history column-shape drift & correction
- **status:** superseded (as a documented drift; the corrected shape is deployed)
- **status-evidence:** The live `011_database_and_infrastructure.md` §"Method 3" explicitly warns "Do not invent column names" and includes a troubleshooting entry for `column "template_id" of relation "agent_instances" does not exist" matching an archived doc's error-prone DDL; migration 073 independently "drops wrong tables and recreates agent_instances/agent_spawn_history matching what create_client_schema and spawn_agent expect."
- **what:** An archived Method-3 fallback DDL for manually creating `agent_instances` used columns that don't match what `spawn_actions.go` actually inserts. The corrected shape (agent_instances: template_id → agent_definitions FK, project FK, config; agent_spawn_history: parent/spawned lineage) was fixed twice independently — once in the live dev-guide's corrected column list + instruction to check `create_client_schema()`'s source first, and once via migration 073's drop-and-recreate.
- **sources:** old/older1/017_creating_new_client_schemas.md#"Method 3"; docs024_key_docs_latest/011_database_and_infrastructure.md#"Method 3"; 073_create_new_client_schema.sql
- **relations:** schema-per-client multi-tenancy; spawn_agent action contract
- **verify-later:** current agent_instances schema in a live client_* schema; create_client_schema function vs 073 shapes

### DBI-007 — Clients → networks → sites hierarchy (early multi-tenancy spine)
- **status:** superseded (as the primary multi-tenancy mechanism; sites itself remains heavily used)
- **status-evidence:** Original CREATE TABLE migration explicitly "designed for 1000s of sites, 10000s+ pages... networks of sites"; a later era notes sites is heavily used while networks/clients are rarely referenced again.
- **what:** The multi-tenancy spine as originally designed: `clients` (linked to auth-service external_id) own `networks` (network-wide settings such as affiliate config), networks own `sites` (domain, brand_dna, github repo/branch, settings, build/deploy timestamps), motivated by cross-site linking within a client's networks and component-level bulk updates across many pages. An early MVP variant of the same migration also carried `site_flows`/`flow_pages` (a flow-based site model, since abandoned) alongside `page_components` with a semantic `data_path`/`data_uuid` addressing scheme intended for future granular editing.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.1; 004_more_on_links.md#Part-1; docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql
- **relations:** ownership hierarchy reuse for entitlement scoping; schema-per-client multi-tenancy; site-plan-and-reconciler (the later, richer layer this hierarchy predates)
- **verify-later:** networks/clients tables — created and populated?; confirm which columns/tables here are still live vs superseded by site_plans/site_work_items

### DBI-008 — sites.build_status vestigial column
- **status:** aspirational
- **status-evidence:** "defaulted to 'pending' at insert, never advanced by any code path … Decide whether to maintain or drop the column" (2026-05-26).
- **stage2-verified (2026-07-14):** unknown → aspirational — sql_for_tables/005_content_components.sql:9921-9931 INSERT INTO sites sets build_status='pending' at insert (matches doc). grep -rn for sites.build_status writers across .go/.sql: 0 hits of any UPDATE/SET on sites.build_status — only site_db_actions.go:1012 INSERT INTO sites(domain,name,network_id,status) which omit...
- **what:** Site-level build_status is dead; real build state lives in last_built_at/last_deployed_at/last_reconciled_at and per-page/per-component build_status. A schema-hygiene decision left open.
- **sources:** HANDOFF_2026-05-26…md#other-open-items
- **relations:** mark_site_deployed (which flips sites.status, not build_status); work-site-orchestrator vs build-site-planner (system-architecture)
- **verify-later:** any writer of sites.build_status

### DBI-009 — layouts.updated_at trigger and the reuse-before-create gate
- **status:** deployed
- **status-evidence:** "CREATE FUNCTION errored = gate firing (shared set_updated_at already exists, used by site_specs/site_plans/content_feed_items/training_runs); CREATE TRIGGER bound to the EXISTING function; bump proved it."
- **what:** Small but doctrine-carrying: layouts.updated_at gained a BEFORE UPDATE trigger, written with a deliberate collision gate — plain `CREATE FUNCTION` (not OR REPLACE) so a name collision errors rather than silently overwriting different semantics, which fired and routed the change onto the shared existing `set_updated_at` function. The codebase convention of explicit `updated_at = now()` in UPDATEs coexists harmlessly with the trigger. The same "columns that never move" hygiene family was later observed on site_work_items.updated_at (listed, not actioned).
- **sources:** w2b_00_trigger_check.sql; w2b_01_layouts_updated_at_trigger.sql; running_notes_scheme_to_components(55).md#Sr #Ss #Su
- **relations:** work-item claim/retry hygiene; SQL change-management
- **verify-later:** pg_trigger rows for layouts; set_updated_at consumers

### DBI-010 — system.internal pseudo-site anchor pattern
- **status:** deployed
- **status-evidence:** "reuse the existing system.internal pseudo-site... Every needs_diagnosis item anchors there."
- **what:** Platform-wide work items that have no natural single site anchor to an existing pseudo-site `system.internal` (id `eac60db8-b032-432b-b36d-76f37632045d`, `sites.status='system'`) rather than inventing a null-site mechanism — discovered because `site_work_items.site_id` is `NOT NULL` and `LoadWorkItemsAction` requires a uuid and filters on it, so a genuinely null-site item could never be loaded. The real site under diagnosis travels in `spec.site_id`/`spec.runtime_site` instead of the item's own site_id column, keeping diagnose items off every per-site dispatch loop.
- **sources:** fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh#header; RUNBOOK_diagnosis_fix_loop(10).md#Q-B CORRECTION; NOTES_running_fixloop(10).md#Turn 4
- **relations:** needs_diagnosis intake route; superseded "null-site allowed" design
- **verify-later:** `SELECT id FROM sites WHERE domain='system.internal'`

### DBI-011 — Ownership hierarchy reuse for entitlement scoping (clients → networks → sites)
- **status:** deployed
- **status-evidence:** "Ownership already exists. Sites sit in a clients → networks → sites hierarchy (sites.network_id → networks.client_id → clients.id, all foreign-keyed)... clients.external_id is a unique external identifier" — a mid-session correction of an earlier assumption that ownership was a wholly new layer, recorded independently in two documents.
- **what:** The platform already has a `clients → networks → sites` ownership hierarchy (all FK'd), with `clients.external_id` as the natural hook to an external billing/identity id and `clients.settings`/`networks.settings` as jsonb rooms for plan/entitlement state. This means `owner_id` on `sites` is unnecessary, and per-domain sell-on is simply re-parenting a site's `network_id` to the buyer's network/client. What is genuinely new is narrower than first thought: entitlement/subscription state itself, the billing-adapter mapping, and the two gate checks.
- **sources:** stripe/001commentary.md#ownership discovery turn; stripe/PLAN_stripe_billing_integration.md#§1,§5; tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#13
- **relations:** clients→networks→sites hierarchy (early spine); operator-vs-vendor business model fork; entitlement gate architecture; two-plane billing architecture
- **verify-later:** sites.network_id, networks.client_id, clients.external_id, clients.settings/networks.settings schema

### DBI-012 — Model-written SQL under a three-guard read-only substrate
- **status:** deployed
- **status-evidence:** "three-guard model-SQL... Guard 3 CONFIRMED — pgbouncer pool_mode = transaction"; an EXPLAIN-estimate pre-flight guard decision recorded in the same running-notes stream.
- **what:** Rather than a vetted-query-catalogue-only approach (too limited for open-ended diagnosis) or an unsafe SQL-string-filter (cannot be made safe against statement stacking, data-modifying CTEs, `COPY ... TO PROGRAM`), the design lets a verdict model emit arbitrary SQL under three layered guards: (1) the prompt instructs SELECT-only; (2) a parse-lint (`sqlguard.go`, word-boundary token checks) drops anything unsafe before it can be issued, recording the drop in the evidence trail; (3) the real safety boundary — execution under a read-only DB transaction (`BeginTx(ctx,&sql.TxOptions{ReadOnly:true})`), confirmed transaction-scoped-safe under pgbouncer's `pool_mode=transaction` (session-level `SET`/`ALTER ROLE` was found NOT reliable under pgbouncer pooling). A schema digest (denylist-based bundle section + EXPLAIN-estimate size guard) is fed to the model so it writes SQL against confirmed columns.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 DB-evidence entries; NOTES_running_synthesis_principles(59); NOTES_running_synthesis_v3(32).md DECISIONS
- **relations:** DB discipline / snapshot_agent convention; pgvector/rag hybrid retrieval reuse; diagnosis loop
- **verify-later:** diagnose_load_runtime execution wiring — the one remaining piece not testable outside the chassis

### DBI-013 — QueryDatabaseAction parameterised queries & schema-drift discipline
- **status:** deployed
- **status-evidence:** "New query_database usage MUST use $1 placeholders with 'params' array. Never embed values via {{.field}} in SQL — SQL injection risk"; "Schema column renames — always check the live schema."
- **what:** `QueryDatabaseAction` supports `$1` placeholders via a `params` config array (never Go-template interpolation into SQL); the live DB is the source of truth for column names (dumps drift), and best-effort/fire-and-forget writes silently no-op on schema mismatch. Includes the `to_jsonb('...'::text)` cast rule for updating prompt templates.
- **sources:** WM/001_development_guide(0).md#1-querydatabaseaction-doesnt-support-parameterised-queries, #18-schema-column-renames-always-check-the-live-schema, #15-postgresql-to_jsonb-fails
- **relations:** debugging guide schema reminders; snapshots/revert; LLM call logging
- **verify-later:** QueryDatabaseAction; site_specs/site_work_items/component_versions schemas

### DBI-014 — Database cleanup and log retention policy
- **status:** deployed
- **status-evidence:** A live `database-cleanup` scheduled task (agent_error_log 14/30 days, audit last 100k, orchestrations 7 days/24h stuck, orchestration_requests FK made CASCADE) plus per-table cleanup functions.
- **what:** A uniform retention discipline: every high-churn operational table has an explicit cleanup function or scheduled CTE with distinct retention for successes vs errors (llm 90/180 days, http 90/180 days, awaited_requests 7 days), and the cleanup task itself is written to always mark itself executed (the "always-return-a-row HAVING fix"). Includes the FK CASCADE fix required so orchestration deletion cascades to requests.
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#database-cleanup; 024_database_cleanup.sql; 026_http_request_log.sql#cleanup
- **relations:** scheduler pre_query pattern; agent_error_log; awaited_requests registry (system-architecture)
- **verify-later:** database-cleanup task enabled and returning rows

### DBI-015 — Early schema inventory and since-dropped tables
- **status:** superseded
- **status-evidence:** A point-in-time `\dt+` snapshot listing tables absent from later docs: flow_pages, site_flows, navigation_structures, pending_requests, improvement_proposals, approval_requests, agent_groups/agent_group_definitions/agent_group_members, agent_metrics, theme_tags, system_events, event_statistics matview; a later doc explicitly replaces "the navigation_structures cache table."
- **what:** A point-in-time inventory of clients_db preserving abandoned concepts: site flows/flow pages (a flow-based site model), a navigation cache table, standalone improvement_proposals and approval_requests tables (roles later absorbed by site_work_items and input_requests), and the agent-groups mechanism (superseded by the agent_group_definitions elimination). Valuable as the "what silently vanished" record.
- **sources:** docs/agent_docs/sql_for_components/006_old_summary_table_descriptions.sql; docs/agent_docs/sql_for_tables/016_nav_tables.sql
- **relations:** superseded by site_work_items, input_requests, site_nav_* tables; "every agent is an orchestrator" elimination of agent_group_definitions (system-architecture)
- **verify-later:** whether these tables still exist in production (dead weight) or were dropped

### DBI-016 — Auth database provisioning
- **status:** deployed
- **status-evidence:** Raw CREATE DATABASE auth_db / CREATE USER auth_user with a subsequent password ALTER, credentials visible in the file.
- **what:** A separate auth_db with its own user for the authentication service, provisioned by hand. The file preserves a real credential — a hygiene finding for stage 2 (secret in docs).
- **sources:** docs/agent_docs/sql_for_tables/021_auth_db.sql
- **relations:** database credentials handling; admin dashboard auth
- **verify-later:** whether that password is still live (rotate); auth service consumer

### DBI-017 — Database password rotation runbook (Postgres → platform secrets → PgBouncer)
- **status:** deployed
- **status-evidence:** Step-by-step live commands with make targets (deploy-065-pgbouncer, pgbouncer-restart/test) and the caution about preserving other secret keys.
- **what:** The password chain has three holders: PostgreSQL users, the `personae-platform-secrets` K8s secret (read by agents), and the `pgbouncer-userlist` secret. Safe rotation order: ALTER USER in PG (existing connections keep working) → update platform secret → rebuild+restart PgBouncer userlist → test → rollout-restart agent pods.
- **sources:** docs001a_password_changing/001_changing_passwords.md
- **relations:** three-database architecture; pgbouncer; API key rotation
- **verify-later:** make targets still exist; secret key inventory

### DBI-018 — sites.status vocabulary and the blast-radius filter trap
- **status:** deployed
- **status-evidence:** UpdateSiteStatusAction vocabulary validated directly from code (v3_site_actions.go:323); the wrong 'active' filter incident recorded as a real production-scoping mistake.
- **what:** `sites.status` vocabulary is draft/building/review/published/deployed/archived/error ('active' is a legacy hand-written value); no code filters on it at build time — it is an informational lifecycle label, and build dispatch keys on `site_work_items` instead (a deployed site is still rebuildable). Heuristic: never scope blast-radius or "live sites" queries with `status='active'` — it silently dropped the site under investigation in a real incident; enumerate `GROUP BY status` first. Companion reuse-gate lesson: check `pg_proc`/`pg_trigger` for an existing shared `set_updated_at()` function before creating a new one.
- **sources:** docs/016b_debugging_guide_merged(3).md#sites.status; RUNBOOK_scheme_to_components(18).md §sites.status RESOLVED
- **relations:** debugging doctrine (0-rows discipline); sites.status informational lifecycle label (system-architecture, SYS-019); shared library blast-radius checks
- **verify-later:** UpdateSiteStatusAction vocabulary; pg_trigger set_updated_at users

### DBI-019 — training_exports Postgres schema (versioned dataset snapshots)
- **status:** deployed
- **status-evidence:** Schema applied with "1,958 rows landed" for a named export_id; used live in a later phase's deploy runbook (2026-06).
- **what:** A two-table schema (`runs` metadata + `rows` ChatML JSONB) chosen over S3 because 21MB–2GB fits Postgres TOAST and avoids a second storage system. A unique index on `(export_id, metadata->>'source_log_id')` blocks duplicate source rows; real-time streaming into it was considered and rejected in favour of named batch snapshots for A/B-comparable training sets. A load-bearing gotcha: `runs.rows_exported` can disagree with the real `rows` count (one export had rows_exported=1957 but 0 actual rows).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4g; phase5/NOTES_phase5_training_launcher_running(39).md; phase5/RUNBOOK_phase_b_c_d_deploy(7).md#step-4
- **relations:** output of Flywheel A; consumed by training-data-preparer / model-trainer
- **verify-later:** training_exports.runs, training_exports.rows; flywheel_A_v3/001_training_exports_schema.sql

### DBI-020 — clients_db vs templates_db agent_definitions source-of-truth
- **status:** deployed
- **status-evidence:** "CORRECTION — chassis reads clients_db, NOT templates_db … templates_db.agent_definitions has the OLD schema (NO version column) … only the 8 original website-builder agents" — resolved after a migration was first mis-applied to clients_db, "corrected" to templates_db, then re-corrected back.
- **what:** The flywheel-C/rich-schema `agent_definitions` (model-trainer, gpu-provisioner, training-launcher) is loaded from `clients_db`, not `templates_db` — the chassis loader query filters `is_snapshot`/`version` columns that exist only in clients_db's rich schema. An older architecture doc's "source of truth is templates_db" statement refers only to the old website-builder catalog and is now misleading.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-03-15:31, #update-2026-06-03-17:3x
- **relations:** every Phase 5 migration targets clients_db; snapshot-shadowing agent-definition loader defect (system-architecture, SYS-024)
- **verify-later:** clients_db.agent_definitions vs templates_db.agent_definitions; migration 103

### DBI-021 — setup.sh box provisioning script (VM-hosted backend sites)
- **status:** deployed
- **status-evidence:** "Provisioning succeeded end-to-end … engine ACTIVE, unit + deploy hook + prune timer installed, nginx OK."
- **what:** A multi-vhost box installer: per-domain nginx server_name blocks (serve static webroot + proxy the API paths), per-domain webroot certbot (graceful HTTP fallback, idempotent re-run upgrades to HTTPS), systemd unit, ufw/fail2ban/logrotate/unattended-upgrades/ssh-hardening, the deploy sudo hook, and a `site-engine-prune.timer`. Parameterised entirely by env-vars (DOMAINS, LETSENCRYPT_EMAIL, DEPLOY_USER, ENGINE_BINARY_PATH, WEBROOT_OWNER, WWW_ALIAS, RETENTION_DAYS, CLOUDFLARE, MODE=full/update); adding a domain is just extending DOMAINS and re-running (idempotent).
- **sources:** traffic_probe_runbook(12).md#3.5; traffic_probe_running_notes(27).md#2026-06-10-box-setup-artifact, #2026-06-13-g
- **relations:** deploy privilege model; VM sizing / Hetzner box selection
- **verify-later:** deploy_setup/vm-deploy/setup.sh (live tree)

### DBI-022 — Deploy privilege model (site-engine-deploy sudo hook)
- **status:** deployed
- **status-evidence:** "Privilege model (low-risk): no root key in CI … a sudoers rule scoped to ONLY that script."
- **what:** No root key lives in CI. When `DEPLOY_USER` is set, setup.sh installs `/usr/local/sbin/site-engine-deploy` (root-owned: atomic binary swap + restart) and a sudoers rule scoped to only that script. The deploy user can swap the engine and nothing else; the swapped binary itself runs as the unprivileged `site-engine` user.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-engine-deploy-workflow; traffic_probe_runbook(12).md#5
- **relations:** setup.sh box provisioning script; later retired by a P5 adapter holding the SSH credential
- **verify-later:** setup.sh DEPLOY_USER branch; sudoers site-engine-deploy

### DBI-023 — VM sizing / Hetzner box selection
- **status:** deployed
- **status-evidence:** "VM sizing (relojistas, its own box): Hetzner CX22-class"; a specific box was provisioned (Hetzner CPX22 #140056673, nbg1, €11.39/mo).
- **what:** Boxes are sized by disk/log headroom, not CPU (static nginx + O(1) JSONL appends stay far inside a small box even at claimed 1.2M visits/mo). EU-only Hetzner (20 TB/mo included) — the runbook standardises on CX23 (~€3.49/mo). x86-only caveat: the engine builds GOARCH=amd64, so Arm (CAX) would need a build-matrix change. Larger domains (relojistas) get their own box; small domains share a multi-vhost box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-store-v2, #2026-06-11-retention-timer, #2026-06-12-operator-execution; traffic_probe_runbook(12).md#3.2
- **relations:** dedicated-box vs shared-box decision; setup.sh box provisioning script
- **verify-later:** relojistas_notes coordinates; Hetzner CPX22 #140056673

### DBI-024 — No tenant isolation today; dedicated-cluster-per-client as offered capability
- **status:** partial
- **status-evidence:** "Single shared Postgres (no row-level security anywhere in the schema), single shared Kafka, single shared ollama-adapter pod — separated only by a site_id column"; cross-cluster scaffolding "exists (remote-job-spawner, DispatchAgentAction), just not exercised in production."
- **what:** The platform has no per-client isolation today — a due-diligence-relevant fact recorded as a landmine so copy never implies otherwise. Real isolation is positioned as buildable: a dedicated cluster per client, with existing but unexercised cross-cluster dispatch scaffolding. UK/EU residency is also not true end-to-end today (compute UK/Rackspace; storage Backblaze us-east-005; cloud models US).
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P5/P6; docs/leopardessconsulting/RUNBOOK.md#landmine-11
- **relations:** multicluster; data-sovereignty positioning; early long-term platform ambitions (system-architecture, SYS-063)
- **verify-later:** RLS absence in schema; remote-job-spawner usage history

### DBI-025 — `platform/fetchguard`: the outbound-fetch (SSRF) guard
- **status:** deployed and exercised — one live call site (`internal/adapters/webscrape/adapter.go`'s `downloadImage`), own test suite proving the properties against real listeners
- **status-evidence:** `go test ./platform/fetchguard/...` — a request to a loopback `httptest.Server` is refused (`TestGuardedClient_RefusesPrivateTarget`); a redirect from a throwaway server to a loopback one is refused at the SAME check, no separate redirect-target inspection (`TestGuardedClient_RedirectToPrivateTargetIsRefused`); a literal metadata-shaped IP is refused (`TestGuardedClient_RefusesLiteralPrivateIP`); the IPv4-in-IPv6-mapped form of the metadata address classifies identically to the bare form (`TestIsPubliclyRoutable_IPv4Mapped`). `internal/adapters/webscrape`'s existing test suite passes unchanged after adoption.
- **what:** The mirror of `httpguard` (DBI-family sibling, inbound-abuse) — a guard for the OUTBOUND direction: any code fetching a URL the platform did not itself choose (scraped content, a discovered image, a customer-typed domain). `NewClient(cfg)` returns an `*http.Client` whose `Transport.DialContext` resolves the target and refuses to dial any address that is not publicly routable (private/loopback/link-local — where cloud metadata endpoints live — multicast/unspecified/`0.0.0.0/8`), checked at the SPECIFIC address about to be dialed rather than a pre-resolved hostname, closing the DNS-rebinding TOCTOU gap a check-then-connect design leaves open. Because redirects re-dial through the same transport, a redirect to a private target is caught by the identical check. `LimitedRead` caps response size and reports truncation explicitly (never a silently-partial body). Built to fix `bugs_open/159`: `downloadImage` fetched image URLs taken from scraped page content — attacker-influenced by construction — with a bare client, no scheme/address/size checks.
- **the landmine it fixes, and the one it could itself become:** `platform/httpguard`'s own package doc scopes itself to inbound abuse only; this package is deliberately a SIBLING, not an addition to that package, so its doc comment stays true. Do not fold outbound-fetch logic into `httpguard` later without renaming its header, or the exact trap this entry exists to close reappears one level up.
- **what it does NOT cover, on purpose:** a headless browser navigating a URL (`internal/adapters/browserrunner/run_checks_action.go`'s `page.Goto`) is a different fetch surface — Playwright does its own DNS and connections, invisible to a Go `http.Transport`. Needs network-layer interception inside the browser or an egress firewall; not built here, flagged rather than silently absent.
- **sources:** `platform/fetchguard/fetchguard.go` + `fetchguard_test.go`; `bugs_open/159`; `webdesign_uk_build_service/PLAN` §8
- **relations:** `httpguard` (DBI sibling, inbound direction); `bugs_open/159`; `LANDMINES.md` "platform/httpguard is INBOUND-abuse only"
- **verify-later:** adoption beyond the one call site — `browser-runner-adapter`'s Playwright navigation, `analyser-adapter`, any future domain-intake flow (webdesign.uk's own P2 teaser will need this the moment it fetches a customer-typed domain directly)
