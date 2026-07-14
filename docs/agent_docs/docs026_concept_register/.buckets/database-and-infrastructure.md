
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Backup discipline: never drop or overwrite an existing backup
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 009 rules section with naming convention `agent_definitions_backup_YYYYMMDD_pre<NNN>`
- **what:** Backups are named for the migration they guard; DROP TABLE IF EXISTS before CREATE destroys the recovery path exactly when needed (failed-and-retried migrations); name collision is the safety mechanism working — pick a new suffix. Nuclear full-table restore procedure retained.
- **sources:** 009#Operational Reference: Backup, Swap, and Revert
- **relations:** snapshot_agent; re-running-migrations hazard
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Three databases and the pgbouncer transaction-mode constraints
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 011 full operational reference
- **what:** clients_db + templates_db (in-cluster PG16 via pgbouncer:6432 transaction mode) and external MySQL auth DB (Clook cPanel, Remote-MySQL IP whitelist 134.213.168.%). Transaction mode forbids prepared statements/session state — conn strings need simple_protocol/cache_describe; pg_dump and LISTEN/NOTIFY must bypass pgbouncer. Go driver split: chassis still pgxpool, core-manager on database/sql (conversion cheat sheet). Auth-to-Postgres migration sketched, not urgent. All credentials in personae-platform-secrets (Terraform 047-base-configs).
- **sources:** 011 full
- **relations:** admin auth flow; backup cronjob
- **verify-later:** which binaries still pgxpool

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Client schema isolation (create_client_schema)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 011 runbook with current clients (demo_client, vetcomparison, test_client, system)
- **what:** Each client gets schema client_<id> with agent_instances/agent_spawn_history/projects (+optional website_projects/agent_memory/workflow_executions); spawn_agent resolves the schema from the client_id Kafka header (validator checks presence only, no DB lookup) and inserts exact columns — manual table creation must match spawn_actions.go's INSERT or spawning fails.
- **sources:** 011#Creating a New Client Schema
- **relations:** multicluster/multitenancy; scheduler client_id header requirement
- **verify-later:** create_client_schema function source

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### sites.build_status vestigial column
- **category:** database-and-infrastructure
- **status-signal:** unknown
- **status-evidence:** "defaulted to 'pending' at insert, never advanced by any code path … Decide whether to maintain or drop the column" (2026-05-26)
- **what:** Site-level build_status is dead; real state lives in last_built_at/last_deployed_at/last_reconciled_at and per-page/per-component build_status. A schema-hygiene decision waiting.
- **sources:** HANDOFF_2026-05-26…md#other-open-items
- **relations:** mark_site_deployed (which flips sites.status, not build_status)
- **verify-later:** any writer of sites.build_status

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### API key rotation flag (STABILITY_API_KEY, BANANA_API_KEY)
- **category:** database-and-infrastructure
- **status-signal:** unknown
- **status-evidence:** "SECURITY — rotate STABILITY_API_KEY and BANANA_API_KEY (plaintext exposure flagged in the imagery handoff). Ops-only action; not addressed" (2026-05-26)
- **what:** Two image-provider API keys were exposed in plaintext and flagged for rotation; still open at the last dated mention.
- **sources:** HANDOFF_2026-05-26…md#other-open-items
- **relations:** imagery pipeline providers
- **verify-later:** whether keys were rotated (ops)

<!-- SOURCE: U03_idea_uk_section_data.md -->
### layouts.updated_at trigger and the reuse-before-create gate
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Notes (Ss): "CREATE FUNCTION errored = gate firing (shared `set_updated_at` already exists, used by site_specs/site_plans/content_feed_items/training_runs); CREATE TRIGGER bound to the EXISTING function; bump proved it. Complete — the reuse path"; (Su) observed it firing on W1b.
- **what:** Small but doctrine-carrying: layouts.updated_at gained a BEFORE UPDATE trigger, written with a deliberate collision gate — plain CREATE FUNCTION (not OR REPLACE) so a name collision errors rather than silently overwriting different semantics, which fired and routed the change onto the shared existing `set_updated_at` function. Notes the codebase convention of explicit `updated_at = now()` in UPDATEs, with which the trigger coexists harmlessly. The same "columns that never move" hygiene family was later observed on site_work_items.updated_at (listed, not actioned).
- **sources:** w2b_00_trigger_check.sql; w2b_01_layouts_updated_at_trigger.sql; running_notes_scheme_to_components(55).md#Sr #Ss #Su
- **relations:** work-item claim/retry hygiene; SQL change-management.
- **verify-later:** pg_trigger rows for layouts; set_updated_at consumers.

## Proposed NEW categories

- **NEW:page-build-pipeline** — the plan_sections → page-content-writer → compile → deploy build path and its semantics: field resolution/deferral (on_missing/skip_field, needs_section_data escalation), LLM routing (needs_llm), array item-key contracts and the render-time reconciler, build-vs-rerender distinction and fossilisation, rerender-pages workflow, the no-component-level-regen limitation, the de-tool hazard. Nine concepts in this unit alone land here; no existing spine slug owns the build path itself (styling-render-pipeline owns CSS/render, site-plan-and-reconciler owns the plan domain).
- **NEW:sql-change-management** — the needle-gate surgery pattern, idempotent sentinel-guarded prompt migrations with paired down-migrations, backup/rollback discipline, run-as-files convention. A coherent expert competence distinct from debugging (which owns the pitfall catalogue) — the council agent for "how production data is changed safely".

## Cross-unit notes for consolidation

- The 016b debugging-guide lessons extracted here (SQL pitfalls, fossil tells, status vocabulary, needle-gate rules) will also surface from the debugging docs unit — merge there, keep this unit as provenance.
- Doc 025 (Phase 4.5), 003 (contracts), 026 (regeneration), 029/030 (site plans) concepts referenced here are anchored in their own units; this unit contributes status evidence (e.g. 003 item 6 rewritten, Phase 4.5 deferred, 026's rerender claim contradicted by direct evidence).
- The two idea.uk site_ids (97ed2f64 June vs 1244516d July) should be reconciled in stage 2.

<!-- SOURCE: U04_idea_uk.md -->
### Prompt/agent-definition migration discipline (snapshot, anchor, sentinel, file-not-paste)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Practised across three real migrations in this unit; the classifier one applied 2026-06-20; the paste failure and jsonb_set no-op both bitten and documented.
- **what:** The safe way to edit live agent prompts and specs in SQL: snapshot_agent() backup inside the same transaction (is_snapshot at version+1000; runtime selection excludes snapshots); UPDATE guarded to the live row only; exact-anchor replace() with a self-check that RAISEs and rolls back if an anchor is missing; idempotency sentinels so re-runs no-op (blind replace would double-expand); single-line anchors (multi-line anchors broke on whitespace); run migration FILES via `kubectl … < file` — pasting into psql mangles \set/\echo/blank lines and once left an open transaction. Companion jsonb facts: jsonb_set into a missing parent silently no-ops — use `||` to add top-level keys; site_specs jsonb column is `data`; partial UNIQUE (site_id,aspect) WHERE is_current.
- **sources:** idea.uk/migration_domain_research_classifier_structured_design_intent.sql (header); idea.uk/019_pcw_prompt_item_fields.sql (idempotency note); idea.uk/HANDOFF(13).md (schemas + operating rules)
- **relations:** build-standard migration (anchor-bug case); snapshot/revert machinery (docs014/016 6.1).
- **verify-later:** snapshot_agent function; bak_* conventions.

<!-- SOURCE: U08_travelling_docs.md -->
### Migrations system — schema_migrations ledger + guarded runner
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Migrations system live (2026-07-10)"; 140 was "the system's first real apply"; applied through 146 at unit close.
- **what:** `schema_migrations` table (filename PK, applied_at, md5 checksum, applied_by, notes) + `scripts/migration/run-migrations.sh` (dry-run default, `--apply`, per-file re-check, stops-without-recording on failure, LOUD warning for near-miss filenames since the repo really uses `NNNb_`/hyphenated names). Home `docs/agent_docs/sql_for_agents/`, baseline 124 (001–123 = pre-system history, never auto-applied). The travelling-docs arc was renumbered 125–139 in applied order and backfilled with dates from the runbook; 128 is an honest reconstruction stub (original lost with an old workspace; effect verified live; NULL checksum). Every migration carries its own guard DO block; parking a re-enable migration outside sql_for_agents/ (141) was used as a deliberate safety gate against a stray --apply.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#2026-07-10-migrations,#fix-140-141; RUNBOOK_travelling_docs(38).md#task-5-closed; HANDOFF_2026-07-10…md#§0.3
- **relations:** migration write-hook to pipeline NOTES; snapshot rule; guard-design patterns (016b).
- **verify-later:** schema_migrations rows; run-migrations.sh behaviour; the 128 stub.

<!-- SOURCE: U10_imagery.md -->
### API keys logged in plaintext (scrub + rotate)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "SECURITY — STILL OPEN, STILL HIGHEST PRIORITY" (2026-05-20) → "✅ DONE 2026-07-08. Keys rotated (user confirmed)."
- **what:** STABILITY_API_KEY and BANANA_API_KEY were logged in plaintext at info level (adapter init env-dump zap.Any, zap.String("apiKey"), plus B2 keys in NewS3Client debug logs) — with billing exposure on the paid Banana tier. Carried as highest-priority for seven weeks across handoffs; resolved by scrubbing to scoped fields and rotating both keys.
- **sources:** TODO_imagery_followups.md#SECURITY, HANDOFF_robot_hands_rebuild.md#Carried-forward, RUNBOOK_imagery_best_in_class.md#B1
- **relations:** provider abstraction (the code carrying the logging); credentials handling conventions.
- **verify-later:** dynamic_adapter.go logging fields; no raw keys in current logs.

<!-- SOURCE: U12_docs024_archives.md -->
### Client schema manual-creation column drift (agent_instances)
- **category:** database-and-infrastructure
- **status-signal:** superseded
- **status-evidence:** Live `011_database_and_infrastructure.md` §"Method 3" explicitly warns "Do not invent column names" and includes a troubleshooting entry for `column "template_id" of relation "agent_instances" does not exist` matching the archive's error-prone DDL.
- **what:** The archive's Method-3 fallback DDL for `agent_instances` used columns that don't match what `spawn_actions.go` actually inserts. Live doc corrects the column list, adds an FK to `projects`, and instructs checking `create_client_schema()`'s source before hand-writing DDL.
- **sources:** old/older1/017_creating_new_client_schemas.md#"Method 3: Manual table creation"; docs024_key_docs_latest/011_database_and_infrastructure.md#"Method 3: Manual table creation"
- **relations:** `create_client_schema()` function; spawn_agent action contract
- **verify-later:** current `agent_instances` schema in a live `client_*` schema.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### system.internal pseudo-site anchor pattern
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "reuse the existing `system.internal` pseudo-site... Every `needs_diagnosis` item anchors there" (090_TRIGGER_needs_diagnosis_v1.sh header; RUNBOOK(10)#Q-B CORRECTION)
- **what:** Platform-wide work items (that have no natural single site) anchor to an existing pseudo-site `system.internal` (id `eac60db8-b032-432b-b36d-76f37632045d`, `sites.status='system'`) rather than inventing a null-site mechanism. Discovered because `site_work_items.site_id` is `NOT NULL` and `LoadWorkItemsAction` requires a uuid and filters `WHERE wi.site_id = $1`, so a genuinely null-site item could never be loaded. The real site under diagnosis travels in `spec.site_id`/`spec.runtime_site` instead of the item's own site_id column, keeping diagnose items off every per-site dispatch loop.
- **sources:** fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh#header points 1-3, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Q-B CORRECTION, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 4
- **relations:** needs_diagnosis intake route; superseded "null-site allowed" design
- **verify-later:** `SELECT id FROM sites WHERE domain='system.internal'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Ownership hierarchy reuse for entitlement scoping (clients → networks → sites)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Ownership already exists. Sites sit in a clients → networks → sites hierarchy (sites.network_id → networks.client_id → clients.id, all foreign-keyed)... clients.external_id is a unique external identifier." (stripe/001commentary.md#ownership discovery turn; PLAN_isolated_chat_environment(5).md §13, correcting PLAN_isolated_chat_environment(2).md §13's original claim that ownership was "the genuinely new layer")
- **what:** The platform already has a `clients → networks → sites` ownership hierarchy (all FK'd), with `clients.external_id` as the natural hook to an external billing/identity id and `clients.settings`/`networks.settings` as jsonb rooms for plan/entitlement state — discovered mid-session correcting an earlier assumption (recorded independently in two documents) that ownership/entitlement was a wholly new layer. This means `owner_id` on `sites` is unnecessary, and per-domain sell-on is simply re-parenting a site's `network_id` to the buyer's network/client. What is genuinely new is narrower than first thought: entitlement/subscription state itself, the billing-adapter mapping, and the two gate checks.
- **sources:** stripe/001commentary.md#ownership discovery turn, stripe/PLAN_stripe_billing_integration.md#§1,§5, tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#13, tools/tool_widget_clobber/PLAN_isolated_chat_environment(2).md#13 (superseded prior claim)
- **relations:** Operator-vs-vendor business model fork; Existing but non-functional auth-service subscription scaffold; Entitlement gate architecture; Two-plane billing architecture
- **verify-later:** `sites.network_id`, `networks.client_id`, `clients.external_id`, `clients.settings`/`networks.settings` schema

<!-- SOURCE: U15_docs019_running_notes.md -->
### Model-written SQL under a three-guard read-only substrate
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** v3(32) DECISIONS: "EXPLAIN-estimate pre-flight guard on data_requests"; principles(59)/v2(36) "three-guard model-SQL... Guard 3 CONFIRMED — pgbouncer pool_mode = transaction."
- **what:** Rather than a vetted-query-catalogue-only approach (rejected as too limited for open-ended diagnosis) or an unsafe SQL-string-filter (rejected: cannot be made safe against statement stacking, data-modifying CTEs, `COPY ... TO PROGRAM`), the design lets the verdict model emit arbitrary SQL under three layered guards: (1) the prompt instructs SELECT-only; (2) a parse-lint (`sqlguard.go`, word-boundary token checks) drops anything unsafe before it can be issued, recording the drop in the evidence trail; (3) the real safety boundary — execution under a read-only DB transaction (`BeginTx(ctx,&sql.TxOptions{ReadOnly:true})`), confirmed transaction-scoped-safe under pgbouncer's `pool_mode=transaction` (session-level `SET`/`ALTER ROLE` was found NOT reliable under pgbouncer pooling). A schema digest (denylist-based `## Schema` bundle section + `EXPLAIN`-estimate size guard) is fed to the model so it writes SQL against confirmed columns.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 DB-evidence entries; NOTES_running_synthesis_principles(59) DB discipline; NOTES_running_synthesis_v3(32).md DECISIONS (turn #2 EXPLAIN guard).
- **relations:** Diagnosis loop; DB discipline / snapshot_agent convention; pgvector/rag hybrid retrieval reuse.
- **verify-later:** `diagnose_load_runtime` execution wiring (flagged as the one remaining piece not testable outside the chassis).

<!-- SOURCE: U15_docs019_running_notes.md -->
### DB discipline / snapshot_agent convention
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Snapshot before any `agent_definitions` change; check `\df` for a helper before hand-writing SQL" (principles(59)); "Backup before any agent_definitions change (standing rule, turn 11)" (v3(32) DECISIONS).
- **what:** A standing rule, reinforced repeatedly across the notes, to call `SELECT snapshot_agent('<type>','<reason>')` (with a paired `revert_agent`) before any change to `agent_definitions`, and to check `\df`/`\dx` for an existing DB helper or extension (pgvector, pgcrypto, pg_trgm all confirmed present) before hand-writing SQL — reuse-before-recreate applied at the database layer.
- **sources:** NOTES_running_synthesis_principles(59) DB discipline section; NOTES_running_synthesis_v3(32).md DECISIONS.
- **relations:** Model-written SQL guard model; schema-before-SQL discipline.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### QueryDatabaseAction parameterised queries & schema-drift discipline
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 001(0) Appendix A #1 "New query_database usage MUST use $1 placeholders with 'params' array. Never embed values via {{.field}} in SQL — SQL injection risk"; #18 "Schema column renames — always check the live schema"
- **what:** `QueryDatabaseAction` supports `$1` placeholders via a `params` config array (never Go-template interpolation into SQL); the live DB is the source of truth for column names (dumps drift), and best-effort/fire-and-forget writes silently no-op on schema mismatch. Includes the `to_jsonb('...'::text)` cast rule for updating prompt templates.
- **sources:** WM/001_development_guide(0).md#1-querydatabaseaction-doesnt-support-parameterised-queries, WM/001_development_guide(0).md#18-schema-column-renames-always-check-the-live-schema, WM/001_development_guide(0).md#15-postgresql-to_jsonb-fails-with-could-not-determine-polymorphic-type
- **relations:** debugging guide schema reminders; snapshots/revert; LLM call logging
- **verify-later:** QueryDatabaseAction; site_specs/site_work_items/component_versions schemas

## Scope-handling notes
excellent_discussions families are purely additive — latest versions contain every earlier section; no abandoned-idea deltas there (they are exploratory/aspirational by nature, "nothing decided", "synthesis spine"). The one genuine family-delta dropped concept found across both sub-trees is the "Adapter & Service Deployment Issues" section in base `016_debugging_guide.md`, absent from `016_debugging_guide_v2_44.md` (captured above). The working/main docs are archive copies whose live successors sit in `docs024_key_docs_latest` (001/016/007/028/029/030/031/033 all tagged superseded/partial per their in-doc phase status vs the live docs already captured in earlier units). This unit overlaps material also touched by U01/U02/U09/U10/U16 (live docs024 + docs019 design/plans) — consolidation should de-duplicate accordingly, retaining this unit's unique value: the MASTER/FOCUS reasoning-architecture concepts (salience, mediator, standards curation, authored/derived context) which have no other extraction unit covering them, since the rest of docs019/_archive/excellent_discussions was not otherwise in scope.

<!-- SOURCE: U17b_docs019_gofiles.md -->
### snapshot-before-mutate practice (snapshot_agent / take_site_snapshot)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** pasted `\df`-style listing shows both functions already exist: `snapshot_agent(uuid, p_agent_type text[, p_reason text])` and `take_site_snapshot(uuid, p_site_id, p_trigger, p_git_sha, p_label, p_created_by)`; also invoked directly in the code-indexer SQL comment ("Snapshot before re-applying to an existing row: SELECT snapshot_agent('code-indexer');")
- **what:** A working convention of snapshotting an agent_definitions row (or a site) via a DB function before applying a migration that touches it, so changes are reversible without relying on git history alone. Paired with a documentation-discipline request to "start or update a runbook, a running notes and a plan" alongside any such migration.
- **sources:** contextkit/001_more_potential_thin_slice_prompt.md, NNN_create_code_indexer_agent.sql
- **relations:** code-indexer agent (applies this practice), site-snapshots-and-revert (014)
- **verify-later:** `\df snapshot_agent` / `\df take_site_snapshot` against the live clients_db schema to confirm current signatures

<!-- SOURCE: U18_sql_for_agents.md -->
### Migration discipline: schema_migrations ledger, snapshot_agent, migration_backups
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 124 "APPLIED 2026-07-10... From this file onward, every numbered file... is applied via scripts/migration/run-migrations.sh... Files 001–123... are HISTORY, not pending work; the runner's baseline (124) excludes them." Backfill rows include a lost-file reconstruction (128).
- **what:** The operational regime for this directory: schema_migrations records WHAT ran WHEN (filename PK, checksum, notes); snapshot_agent(type, reason) is the standing rule opening every agent-updating transaction (MVCC before-image); migration_backups holds manual before-values; 107's backup preamble adds the no-DROP rule ("The collision IS the safety net"). Workflow-altering migrations must leave a pipeline doc_note (runbook §3, seen in 141/142/144/146).
- **sources:** 124_schema_migrations.sql; 107_image_build_handler.sql; 131_tool_generator_plan_writing.sql; 128_fix_load_runtime_error_step_target.sql
- **relations:** travelling docs; versioned agent_definitions (UNIQUE on type+version, 121)
- **verify-later:** run-migrations.sh; snapshot_agent function; schema_migrations contents

<!-- SOURCE: U18_sql_for_agents.md -->
### client_system schema for agent instances
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 073 drops wrong tables and recreates agent_instances/agent_spawn_history "matching what create_client_schema and spawn_agent expect".
- **what:** Per-client schema tables backing spawn_agent: agent_instances (template_id → agent_definitions, project FK, config) and agent_spawn_history (parent/spawned lineage). Documents the column contract spawn_agent expects.
- **sources:** 073_create_new_client_schema.sql
- **relations:** spawn_agent action; create_client_schema; multitenancy/client schemas (docs 011)
- **verify-later:** create_client_schema function vs 073 shapes

## Proposed NEW categories
- **NEW:build-pipeline** — the site_work_items work-item build pipeline and its builder/handler agents (pageflow-builder, site-work-orchestrator, dispatch loop, page-build-handler, page-content-writer, page-rebuild). Distinct from improvement-loop (post-build) and site-plan-and-reconciler (planning domain); large enough to back a council agent.
- **NEW:rag-retrieval** — shared knowledge_base pgvector store, rag_index/rag_lookup actions, collections (tool_docs, industry_sites), embedding-model management. Not covered by model-infrastructure (endpoints/GPUs) or documentation-system.

<!-- SOURCE: U19_sql_tables_components.md -->
### Database cleanup and log retention policy
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** database-cleanup task (agent_error_log 14/30 days, audit last 100k, orchestrations 7 days/24h stuck, orchestration_requests FK made CASCADE) plus per-table cleanup functions (llm 90/180, http 90/180, awaited 7 days) and the always-return-a-row HAVING fix.
- **what:** A uniform retention discipline: every high-churn operational table has an explicit cleanup function or scheduled CTE with distinct retention for successes vs errors, and the cleanup task itself is written to always mark itself executed. Includes the FK CASCADE fix required so orchestration deletion cascades to requests.
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#database-cleanup; docs/agent_docs/sql_for_tables/024_database_cleanup.sql; docs/agent_docs/sql_for_tables/026_http_request_log.sql#cleanup
- **relations:** scheduler pre_query pattern; agent_error_log; llm/http logs.
- **verify-later:** database-cleanup enabled and returning rows.

<!-- SOURCE: U19_sql_tables_components.md -->
### Migration discipline: pre-change snapshots, renumbering, footguns
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Recurring practice across files: CREATE TABLE ... AS SELECT snapshots inside the txn (pages_bak_retype_guides, assets_backup_20260508...), NNN placeholders with "confirm the next free migration number" notes (048, 046, 047), deliberate plain CREATE TABLE to error on shape mismatch ("the migration-110 trap, §6.1"), and "code shipped but migration unapplied has bitten this project repeatedly" (042 ssh_port).
- **what:** The project's migration conventions as embodied in the files: snapshot rows before destructive UPDATEs with pasted rollback SQL; verify blocks (DO $$ ... RAISE EXCEPTION) inside transactions; idempotence via IF NOT EXISTS except where silent no-op would hide a shape conflict; migration numbers confirmed against the live runner before applying; migrations applied separately from code deploys.
- **sources:** docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql#BEFORE-APPLYING; docs/agent_docs/sql_for_tables/003_pages.sql#snapshots; docs/agent_docs/sql_for_tables/042_thunder.sql#ssh_port
- **relations:** debugging guide §6.1 / item 17; agent snapshot/revert.
- **verify-later:** migration runner and numbering source of truth.

<!-- SOURCE: U19_sql_tables_components.md -->
### Early schema inventory and since-dropped tables
- **category:** database-and-infrastructure
- **status-signal:** superseded
- **status-evidence:** 006 is a psql \dt+ snapshot listing tables absent from later docs: flow_pages, site_flows, navigation_structures, pending_requests, improvement_proposals, approval_requests, agent_groups/agent_group_definitions/agent_group_members, agent_metrics, theme_tags, system_events, event_statistics matview; 016 explicitly replaces "the navigation_structures cache table".
- **what:** A point-in-time inventory of clients_db that preserves abandoned concepts: site flows/flow pages (a flow-based site model), a navigation cache table, standalone improvement_proposals and approval_requests tables (roles later absorbed by site_work_items and input_requests), and an agent-groups mechanism. Valuable as the "what silently vanished" record.
- **sources:** docs/agent_docs/sql_for_components/006_old_summary_table_descriptions.sql; docs/agent_docs/sql_for_tables/016_nav_tables.sql
- **relations:** superseded by site_work_items, input_requests, site_nav_* tables.
- **verify-later:** whether these tables still exist in production (dead weight) or were dropped.

<!-- SOURCE: U19_sql_tables_components.md -->
### Auth database provisioning
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Raw CREATE DATABASE auth_db / CREATE USER auth_user with a subsequent password ALTER (credentials visible in file).
- **what:** A separate auth_db with its own user for the authentication service, provisioned by hand. The file preserves a real credential — a hygiene finding for stage 2 (secret in docs).
- **sources:** docs/agent_docs/sql_for_tables/021_auth_db.sql
- **relations:** database-and-infrastructure credentials; admin dashboard auth.
- **verify-later:** whether that password is still live (rotate); auth service consumer.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Database password rotation runbook (Postgres → platform secrets → PgBouncer)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Step-by-step live commands with make targets (deploy-065-pgbouncer, pgbouncer-restart/test) and the caution about preserving other secret keys.
- **what:** The password chain has three holders: PostgreSQL users, the `personae-platform-secrets` K8s secret (read by agents), and the `pgbouncer-userlist` secret. Safe rotation order: ALTER USER in PG (existing conns keep working) → update platform secret → rebuild+restart PgBouncer userlist → test → rollout-restart agent pods.
- **sources:** docs001a_password_changing/001_changing_passwords.md
- **relations:** pgbouncer; clients_db/templates_db users; credentials handling (database-and-infrastructure docs 011).
- **verify-later:** make targets still exist; secret key inventory.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Clients → networks → sites hierarchy
- **category:** database-and-infrastructure
- **status-signal:** partial
- **status-evidence:** docs012/006 and /007 CREATE TABLE clients/networks/sites "designed for 1000s of sites, 10000s+ pages... networks of sites"; sites is heavily used later, networks/clients rarely referenced again.
- **what:** The multi-tenancy spine: clients (linked to auth-service external_id) own networks (with network-wide settings such as affiliate config), networks own sites (domain, brand_dna, github repo/branch, settings, build/deploy timestamps). Motivated by cross-site linking within a client's networks and component-level bulk updates across many pages.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.1; docs012_site_maps_and_components/004_more_on_links.md#Part-1
- **relations:** cross-site link scope; multicluster scaling; client schemas in database-and-infrastructure.
- **verify-later:** networks/clients tables — created and populated?

<!-- SOURCE: U23_docs_root_vonc.md -->
### sites.status vocabulary and the blast-radius filter trap
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 016b merged entry: UpdateSiteStatusAction (v3_site_actions.go:323) validated vocabulary read from code; the wrong 'active' filter incident recorded.
- **what:** `sites.status` vocabulary is draft/building/review/published/deployed/archived/error ('active' is a legacy hand-written value); no code filters on it — it is an informational lifecycle label, and build dispatch keys on site_work_items (a deployed site is still rebuildable). Heuristic: never scope blast-radius or "live sites" queries with status='active' (it silently dropped the site under investigation); enumerate GROUP BY status first. Companion reuse-gate lesson: a shared set_updated_at() trigger function already exists — check pg_proc/pg_trigger before creating.
- **sources:** docs/016b_debugging_guide_merged(3).md#sites.status
- **relations:** debugging doctrine (0-rows discipline); shared library blast-radius checks
- **verify-later:** UpdateSiteStatusAction vocabulary; pg_trigger set_updated_at users

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### training_exports Postgres schema (versioned dataset snapshots)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4g schema SQL + §2.4i "1,958 rows landed" with export_id `fef7be6b-…`; RUNBOOK bcd(7) §4b uses `training_exports.runs`/`rows` live (2026-06)
- **what:** Two-table schema (`runs` metadata + `rows` ChatML JSONB) chosen over S3 because 21MB–2GB fits Postgres TOAST and avoids a second storage system. A unique index on `(export_id, metadata->>'source_log_id')` blocks duplicate source rows; real-time streaming into it was considered and rejected in favour of named batch snapshots for A/B-comparable training sets. A load-bearing gotcha surfaced later: `runs.rows_exported` can disagree with the real `rows` count (export `a8484922` had rows_exported=1957 but 0 actual rows).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4g; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-06-2; phase5/RUNBOOK_phase_b_c_d_deploy(7).md#step-4
- **relations:** output of Flywheel A; consumed by training-data-preparer / model-trainer
- **verify-later:** training_exports.runs, training_exports.rows; flywheel_A_v3/001_training_exports_schema.sql

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### clients_db vs templates_db agent_definitions source-of-truth
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-03 ~17:3x: CORRECTION — chassis reads clients_db, NOT templates_db … templates_db.agent_definitions has the OLD schema (NO version column) … only the 8 original website-builder agents"
- **what:** A multi-session saga establishing that the flywheel-C/rich-schema `agent_definitions` (model-trainer, gpu-provisioner, training-launcher) live and are loaded from `clients_db`, not `templates_db`. Migration 103 first mis-applied to clients_db, then "corrected" to templates_db, then re-corrected: the chassis loader query filters `is_snapshot`/`version` columns that exist only in clients_db's rich schema. The 002 architecture doc's "source of truth is templates_db" refers to the old website-builder catalog.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-03-15:31, #update-2026-06-03-17:3x
- **relations:** every Phase 5 migration targets clients_db; contradicts the frozen 002_system_architecture pack copy
- **verify-later:** clients_db.agent_definitions vs templates_db.agent_definitions; migration 103

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### setup.sh box provisioning script
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "Provisioning succeeded end-to-end … engine ACTIVE, unit + deploy hook + prune timer installed, nginx OK"; runbook(12) §3.5.
- **what:** Multi-vhost box installer: per-domain nginx server_name blocks (serve static webroot + proxy the four API paths), per-domain webroot certbot (graceful HTTP fallback, idempotent re-run upgrades to HTTPS), systemd unit, ufw/fail2ban/logrotate/unattended-upgrades/ssh-hardening, the deploy sudo hook, and `site-engine-prune.timer`. Params are env-vars: DOMAINS, LETSENCRYPT_EMAIL, DEPLOY_USER, ENGINE_BINARY_PATH, WEBROOT_OWNER, WWW_ALIAS, RETENTION_DAYS, CLOUDFLARE, MODE=full/update. Add a domain = extend DOMAINS + re-run (idempotent).
- **sources:** traffic_probe_runbook(12).md#3.5, traffic_probe_running_notes(27).md#2026-06-10-box-setup-artifact, traffic_probe_running_notes(27).md#2026-06-13-g
- **relations:** origin of the deploy privilege model; WWW_ALIAS/CLOUDFLARE/RETENTION shipped incrementally
- **verify-later:** deploy_setup/vm-deploy/setup.sh (live tree)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Deploy privilege model (site-engine-deploy sudo hook)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-10 "Privilege model (low-risk): no root key in CI … a sudoers rule scoped to ONLY that script".
- **what:** No root key in CI. When `DEPLOY_USER` is set, setup.sh installs `/usr/local/sbin/site-engine-deploy` (root-owned: atomic binary swap + restart) and a sudoers rule scoped to only that script. The deploy user can swap the engine and nothing else; the swapped binary runs as the unprivileged `site-engine` user.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-engine-deploy-workflow, traffic_probe_runbook(12).md#5
- **relations:** part of engine Action; retired later by P5 adapter holding the SSH credential
- **verify-later:** setup.sh DEPLOY_USER branch; sudoers site-engine-deploy

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### VM sizing / Hetzner box selection
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "VM sizing (relojistas, its own box): Hetzner CX22-class"; 2026-06-12 "Box provisioned: Hetzner CPX22 #140056673, nbg1 … IP 167.233.33.159, €11.39/mo".
- **what:** Boxes are sized by disk/log headroom not CPU (static nginx + O(1) JSONL appends are far inside a small box even at claimed 1.2M visits/mo). EU-only Hetzner (20 TB/mo included) — runbook standardises on CX23 (~€3.49/mo). x86-only caveat: the engine Action builds GOARCH=amd64, so Arm (CAX) would need a build-matrix change. relojistas has its own box; small domains share a multi-vhost box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-store-v2, traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_running_notes(27).md#2026-06-12-operator-execution, traffic_probe_runbook(12).md#3.2
- **relations:** dedicated-box vs shared-box decision
- **verify-later:** relojistas_notes coordinates; Hetzner CPX22 #140056673

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### core client→network→site→page content hierarchy (early MVP schema)
- **category:** database-and-infrastructure
- **status-signal:** superseded
- **status-evidence:** Filed as an "MVP Migration" ("Designed for patch-only updates from the start") establishing `clients`/`networks`/`sites`/`site_flows`/`pages`/`flow_pages`/`page_components` with a minimal column set (e.g. `pages` here has no `sections`, `build_status` as later became standard, no `site_plan_*` linkage) — an early snapshot of a hierarchy that has since grown substantially richer elsewhere in the live schema.
- **what:** The foundational multi-tenant hierarchy this platform is built on: `clients` (external_id linking to auth-service) → `networks` (affiliate/network-wide settings) → `sites` (domain, brand_dna, github_repo/branch) → `site_flows` (multi-track audience journeys with a narrative_arc) → `pages` (page_type, nav ordering, content_hash for change detection) → `page_components` (template instances with rendered_html, content_data, and a semantic `data_path`/`data_uuid` addressing scheme intended for future granular editing).
- **sources:** docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql
- **relations:** link registry + navigation cache (below, same migration file); system-architecture; site-plan-and-reconciler (the later, richer plan/reconciler layer this hierarchy predates)
- **verify-later:** the current `clients`/`networks`/`sites`/`pages` schema shape vs. this early version — confirm which columns/tables here are still live as originally designed vs. superseded by site_plans/site_work_items

<!-- SOURCE: U25_leopardess_social.md -->
### No tenant isolation today; dedicated-cluster-per-client as offered capability
- **category:** database-and-infrastructure
- **status-signal:** partial
- **status-evidence:** AUDIT P5: "Single shared Postgres (no row-level security anywhere in the schema), single shared Kafka, single shared ollama-adapter pod — separated only by a site_id column"; cross-cluster scaffolding "exists (remote-job-spawner, DispatchAgentAction), just not exercised in production".
- **what:** The platform has no per-client isolation — a due-diligence-relevant fact recorded as a landmine so copy never implies otherwise. Real isolation is positioned as buildable: a dedicated cluster per client, with existing but unexercised cross-cluster dispatch scaffolding. UK/EU residency is also not true end-to-end today (compute UK/Rackspace; storage Backblaze us-east-005; cloud models US).
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P5/P6; docs/leopardessconsulting/RUNBOOK.md#landmine-11
- **relations:** multicluster; data-sovereignty positioning; UK-sovereign stack exploration
- **verify-later:** RLS absence in schema; remote-job-spawner usage history

<!-- SOURCE: U26_misc_dirs.md -->
### Three-database architecture (MySQL auth + PG clients + PG templates)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** databases.md is a factual "Database Architecture Summary" (not a proposal); basic_usage/001 contains live credentials/connection commands for both MySQL (rs17.uk-noc.com) and postgres-clients; current taxonomy (011) still lists MySQL + client schemas.
- **what:** Authentication isolated in MySQL (users, JWT refresh tokens, profiles, projects, subscriptions/tiers, permissions, activity logs; BINARY(16) UUIDs); agent/AI runtime in PostgreSQL clients DB (global agent_definitions, orchestrator_state, clients_info + per-client schemas); shared persona templates in a second PostgreSQL DB. Core Manager owns clients/templates access with read-only auth access.
- **sources:** docs/architecture/databases.md; docs/basic_usage/001basic_usage.txt
- **relations:** schema-per-client multi-tenancy; AI Persona Platform API (auth endpoints)
- **verify-later:** current DB inventory; whether templates DB still exists separately

<!-- SOURCE: U26_misc_dirs.md -->
### Schema-per-client multi-tenancy
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** databases.md documents `create_client_schema()` and per-client tables as implemented; every operational query in basic_usage targets `client_demo_client.*` schemas.
- **what:** Each client gets an isolated PostgreSQL schema (client_{id}) containing agent_instances, agent_memory (pgvector embeddings), projects, workflow_executions and usage_analytics, created by a SQL function; global resources (agent_definitions, templates, orchestrator_state) are shared. Strong tenant isolation on shared infrastructure.
- **sources:** docs/architecture/databases.md#2-postgresql-database-1; docs/basic_usage/001basic_usage.txt
- **relations:** three-database architecture; agent spawning (instances live per-schema)
- **verify-later:** create_client_schema function; pgvector agent_memory usage in current code

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Backup discipline: never drop or overwrite an existing backup
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 009 rules section with naming convention `agent_definitions_backup_YYYYMMDD_pre<NNN>`
- **what:** Backups are named for the migration they guard; DROP TABLE IF EXISTS before CREATE destroys the recovery path exactly when needed (failed-and-retried migrations); name collision is the safety mechanism working — pick a new suffix. Nuclear full-table restore procedure retained.
- **sources:** 009#Operational Reference: Backup, Swap, and Revert
- **relations:** snapshot_agent; re-running-migrations hazard
- **verify-later:** —

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Three databases and the pgbouncer transaction-mode constraints
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 011 full operational reference
- **what:** clients_db + templates_db (in-cluster PG16 via pgbouncer:6432 transaction mode) and external MySQL auth DB (Clook cPanel, Remote-MySQL IP whitelist 134.213.168.%). Transaction mode forbids prepared statements/session state — conn strings need simple_protocol/cache_describe; pg_dump and LISTEN/NOTIFY must bypass pgbouncer. Go driver split: chassis still pgxpool, core-manager on database/sql (conversion cheat sheet). Auth-to-Postgres migration sketched, not urgent. All credentials in personae-platform-secrets (Terraform 047-base-configs).
- **sources:** 011 full
- **relations:** admin auth flow; backup cronjob
- **verify-later:** which binaries still pgxpool

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Client schema isolation (create_client_schema)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 011 runbook with current clients (demo_client, vetcomparison, test_client, system)
- **what:** Each client gets schema client_<id> with agent_instances/agent_spawn_history/projects (+optional website_projects/agent_memory/workflow_executions); spawn_agent resolves the schema from the client_id Kafka header (validator checks presence only, no DB lookup) and inserts exact columns — manual table creation must match spawn_actions.go's INSERT or spawning fails.
- **sources:** 011#Creating a New Client Schema
- **relations:** multicluster/multitenancy; scheduler client_id header requirement
- **verify-later:** create_client_schema function source

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### sites.build_status vestigial column
- **category:** database-and-infrastructure
- **status-signal:** unknown
- **status-evidence:** "defaulted to 'pending' at insert, never advanced by any code path … Decide whether to maintain or drop the column" (2026-05-26)
- **what:** Site-level build_status is dead; real state lives in last_built_at/last_deployed_at/last_reconciled_at and per-page/per-component build_status. A schema-hygiene decision waiting.
- **sources:** HANDOFF_2026-05-26…md#other-open-items
- **relations:** mark_site_deployed (which flips sites.status, not build_status)
- **verify-later:** any writer of sites.build_status

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### API key rotation flag (STABILITY_API_KEY, BANANA_API_KEY)
- **category:** database-and-infrastructure
- **status-signal:** unknown
- **status-evidence:** "SECURITY — rotate STABILITY_API_KEY and BANANA_API_KEY (plaintext exposure flagged in the imagery handoff). Ops-only action; not addressed" (2026-05-26)
- **what:** Two image-provider API keys were exposed in plaintext and flagged for rotation; still open at the last dated mention.
- **sources:** HANDOFF_2026-05-26…md#other-open-items
- **relations:** imagery pipeline providers
- **verify-later:** whether keys were rotated (ops)

<!-- SOURCE: U03_idea_uk_section_data.md -->
### layouts.updated_at trigger and the reuse-before-create gate
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Notes (Ss): "CREATE FUNCTION errored = gate firing (shared `set_updated_at` already exists, used by site_specs/site_plans/content_feed_items/training_runs); CREATE TRIGGER bound to the EXISTING function; bump proved it. Complete — the reuse path"; (Su) observed it firing on W1b.
- **what:** Small but doctrine-carrying: layouts.updated_at gained a BEFORE UPDATE trigger, written with a deliberate collision gate — plain CREATE FUNCTION (not OR REPLACE) so a name collision errors rather than silently overwriting different semantics, which fired and routed the change onto the shared existing `set_updated_at` function. Notes the codebase convention of explicit `updated_at = now()` in UPDATEs, with which the trigger coexists harmlessly. The same "columns that never move" hygiene family was later observed on site_work_items.updated_at (listed, not actioned).
- **sources:** w2b_00_trigger_check.sql; w2b_01_layouts_updated_at_trigger.sql; running_notes_scheme_to_components(55).md#Sr #Ss #Su
- **relations:** work-item claim/retry hygiene; SQL change-management.
- **verify-later:** pg_trigger rows for layouts; set_updated_at consumers.

## Proposed NEW categories

- **NEW:page-build-pipeline** — the plan_sections → page-content-writer → compile → deploy build path and its semantics: field resolution/deferral (on_missing/skip_field, needs_section_data escalation), LLM routing (needs_llm), array item-key contracts and the render-time reconciler, build-vs-rerender distinction and fossilisation, rerender-pages workflow, the no-component-level-regen limitation, the de-tool hazard. Nine concepts in this unit alone land here; no existing spine slug owns the build path itself (styling-render-pipeline owns CSS/render, site-plan-and-reconciler owns the plan domain).
- **NEW:sql-change-management** — the needle-gate surgery pattern, idempotent sentinel-guarded prompt migrations with paired down-migrations, backup/rollback discipline, run-as-files convention. A coherent expert competence distinct from debugging (which owns the pitfall catalogue) — the council agent for "how production data is changed safely".

## Cross-unit notes for consolidation

- The 016b debugging-guide lessons extracted here (SQL pitfalls, fossil tells, status vocabulary, needle-gate rules) will also surface from the debugging docs unit — merge there, keep this unit as provenance.
- Doc 025 (Phase 4.5), 003 (contracts), 026 (regeneration), 029/030 (site plans) concepts referenced here are anchored in their own units; this unit contributes status evidence (e.g. 003 item 6 rewritten, Phase 4.5 deferred, 026's rerender claim contradicted by direct evidence).
- The two idea.uk site_ids (97ed2f64 June vs 1244516d July) should be reconciled in stage 2.

<!-- SOURCE: U04_idea_uk.md -->
### Prompt/agent-definition migration discipline (snapshot, anchor, sentinel, file-not-paste)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Practised across three real migrations in this unit; the classifier one applied 2026-06-20; the paste failure and jsonb_set no-op both bitten and documented.
- **what:** The safe way to edit live agent prompts and specs in SQL: snapshot_agent() backup inside the same transaction (is_snapshot at version+1000; runtime selection excludes snapshots); UPDATE guarded to the live row only; exact-anchor replace() with a self-check that RAISEs and rolls back if an anchor is missing; idempotency sentinels so re-runs no-op (blind replace would double-expand); single-line anchors (multi-line anchors broke on whitespace); run migration FILES via `kubectl … < file` — pasting into psql mangles \set/\echo/blank lines and once left an open transaction. Companion jsonb facts: jsonb_set into a missing parent silently no-ops — use `||` to add top-level keys; site_specs jsonb column is `data`; partial UNIQUE (site_id,aspect) WHERE is_current.
- **sources:** idea.uk/migration_domain_research_classifier_structured_design_intent.sql (header); idea.uk/019_pcw_prompt_item_fields.sql (idempotency note); idea.uk/HANDOFF(13).md (schemas + operating rules)
- **relations:** build-standard migration (anchor-bug case); snapshot/revert machinery (docs014/016 6.1).
- **verify-later:** snapshot_agent function; bak_* conventions.

<!-- SOURCE: U08_travelling_docs.md -->
### Migrations system — schema_migrations ledger + guarded runner
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Migrations system live (2026-07-10)"; 140 was "the system's first real apply"; applied through 146 at unit close.
- **what:** `schema_migrations` table (filename PK, applied_at, md5 checksum, applied_by, notes) + `scripts/migration/run-migrations.sh` (dry-run default, `--apply`, per-file re-check, stops-without-recording on failure, LOUD warning for near-miss filenames since the repo really uses `NNNb_`/hyphenated names). Home `docs/agent_docs/sql_for_agents/`, baseline 124 (001–123 = pre-system history, never auto-applied). The travelling-docs arc was renumbered 125–139 in applied order and backfilled with dates from the runbook; 128 is an honest reconstruction stub (original lost with an old workspace; effect verified live; NULL checksum). Every migration carries its own guard DO block; parking a re-enable migration outside sql_for_agents/ (141) was used as a deliberate safety gate against a stray --apply.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#2026-07-10-migrations,#fix-140-141; RUNBOOK_travelling_docs(38).md#task-5-closed; HANDOFF_2026-07-10…md#§0.3
- **relations:** migration write-hook to pipeline NOTES; snapshot rule; guard-design patterns (016b).
- **verify-later:** schema_migrations rows; run-migrations.sh behaviour; the 128 stub.

<!-- SOURCE: U10_imagery.md -->
### API keys logged in plaintext (scrub + rotate)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "SECURITY — STILL OPEN, STILL HIGHEST PRIORITY" (2026-05-20) → "✅ DONE 2026-07-08. Keys rotated (user confirmed)."
- **what:** STABILITY_API_KEY and BANANA_API_KEY were logged in plaintext at info level (adapter init env-dump zap.Any, zap.String("apiKey"), plus B2 keys in NewS3Client debug logs) — with billing exposure on the paid Banana tier. Carried as highest-priority for seven weeks across handoffs; resolved by scrubbing to scoped fields and rotating both keys.
- **sources:** TODO_imagery_followups.md#SECURITY, HANDOFF_robot_hands_rebuild.md#Carried-forward, RUNBOOK_imagery_best_in_class.md#B1
- **relations:** provider abstraction (the code carrying the logging); credentials handling conventions.
- **verify-later:** dynamic_adapter.go logging fields; no raw keys in current logs.

<!-- SOURCE: U12_docs024_archives.md -->
### Client schema manual-creation column drift (agent_instances)
- **category:** database-and-infrastructure
- **status-signal:** superseded
- **status-evidence:** Live `011_database_and_infrastructure.md` §"Method 3" explicitly warns "Do not invent column names" and includes a troubleshooting entry for `column "template_id" of relation "agent_instances" does not exist` matching the archive's error-prone DDL.
- **what:** The archive's Method-3 fallback DDL for `agent_instances` used columns that don't match what `spawn_actions.go` actually inserts. Live doc corrects the column list, adds an FK to `projects`, and instructs checking `create_client_schema()`'s source before hand-writing DDL.
- **sources:** old/older1/017_creating_new_client_schemas.md#"Method 3: Manual table creation"; docs024_key_docs_latest/011_database_and_infrastructure.md#"Method 3: Manual table creation"
- **relations:** `create_client_schema()` function; spawn_agent action contract
- **verify-later:** current `agent_instances` schema in a live `client_*` schema.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### system.internal pseudo-site anchor pattern
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "reuse the existing `system.internal` pseudo-site... Every `needs_diagnosis` item anchors there" (090_TRIGGER_needs_diagnosis_v1.sh header; RUNBOOK(10)#Q-B CORRECTION)
- **what:** Platform-wide work items (that have no natural single site) anchor to an existing pseudo-site `system.internal` (id `eac60db8-b032-432b-b36d-76f37632045d`, `sites.status='system'`) rather than inventing a null-site mechanism. Discovered because `site_work_items.site_id` is `NOT NULL` and `LoadWorkItemsAction` requires a uuid and filters `WHERE wi.site_id = $1`, so a genuinely null-site item could never be loaded. The real site under diagnosis travels in `spec.site_id`/`spec.runtime_site` instead of the item's own site_id column, keeping diagnose items off every per-site dispatch loop.
- **sources:** fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh#header points 1-3, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Q-B CORRECTION, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 4
- **relations:** needs_diagnosis intake route; superseded "null-site allowed" design
- **verify-later:** `SELECT id FROM sites WHERE domain='system.internal'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Ownership hierarchy reuse for entitlement scoping (clients → networks → sites)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Ownership already exists. Sites sit in a clients → networks → sites hierarchy (sites.network_id → networks.client_id → clients.id, all foreign-keyed)... clients.external_id is a unique external identifier." (stripe/001commentary.md#ownership discovery turn; PLAN_isolated_chat_environment(5).md §13, correcting PLAN_isolated_chat_environment(2).md §13's original claim that ownership was "the genuinely new layer")
- **what:** The platform already has a `clients → networks → sites` ownership hierarchy (all FK'd), with `clients.external_id` as the natural hook to an external billing/identity id and `clients.settings`/`networks.settings` as jsonb rooms for plan/entitlement state — discovered mid-session correcting an earlier assumption (recorded independently in two documents) that ownership/entitlement was a wholly new layer. This means `owner_id` on `sites` is unnecessary, and per-domain sell-on is simply re-parenting a site's `network_id` to the buyer's network/client. What is genuinely new is narrower than first thought: entitlement/subscription state itself, the billing-adapter mapping, and the two gate checks.
- **sources:** stripe/001commentary.md#ownership discovery turn, stripe/PLAN_stripe_billing_integration.md#§1,§5, tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#13, tools/tool_widget_clobber/PLAN_isolated_chat_environment(2).md#13 (superseded prior claim)
- **relations:** Operator-vs-vendor business model fork; Existing but non-functional auth-service subscription scaffold; Entitlement gate architecture; Two-plane billing architecture
- **verify-later:** `sites.network_id`, `networks.client_id`, `clients.external_id`, `clients.settings`/`networks.settings` schema

<!-- SOURCE: U15_docs019_running_notes.md -->
### Model-written SQL under a three-guard read-only substrate
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** v3(32) DECISIONS: "EXPLAIN-estimate pre-flight guard on data_requests"; principles(59)/v2(36) "three-guard model-SQL... Guard 3 CONFIRMED — pgbouncer pool_mode = transaction."
- **what:** Rather than a vetted-query-catalogue-only approach (rejected as too limited for open-ended diagnosis) or an unsafe SQL-string-filter (rejected: cannot be made safe against statement stacking, data-modifying CTEs, `COPY ... TO PROGRAM`), the design lets the verdict model emit arbitrary SQL under three layered guards: (1) the prompt instructs SELECT-only; (2) a parse-lint (`sqlguard.go`, word-boundary token checks) drops anything unsafe before it can be issued, recording the drop in the evidence trail; (3) the real safety boundary — execution under a read-only DB transaction (`BeginTx(ctx,&sql.TxOptions{ReadOnly:true})`), confirmed transaction-scoped-safe under pgbouncer's `pool_mode=transaction` (session-level `SET`/`ALTER ROLE` was found NOT reliable under pgbouncer pooling). A schema digest (denylist-based `## Schema` bundle section + `EXPLAIN`-estimate size guard) is fed to the model so it writes SQL against confirmed columns.
- **sources:** NOTES_running_synthesis_v2(36).md 2026-06-17 DB-evidence entries; NOTES_running_synthesis_principles(59) DB discipline; NOTES_running_synthesis_v3(32).md DECISIONS (turn #2 EXPLAIN guard).
- **relations:** Diagnosis loop; DB discipline / snapshot_agent convention; pgvector/rag hybrid retrieval reuse.
- **verify-later:** `diagnose_load_runtime` execution wiring (flagged as the one remaining piece not testable outside the chassis).

<!-- SOURCE: U15_docs019_running_notes.md -->
### DB discipline / snapshot_agent convention
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "Snapshot before any `agent_definitions` change; check `\df` for a helper before hand-writing SQL" (principles(59)); "Backup before any agent_definitions change (standing rule, turn 11)" (v3(32) DECISIONS).
- **what:** A standing rule, reinforced repeatedly across the notes, to call `SELECT snapshot_agent('<type>','<reason>')` (with a paired `revert_agent`) before any change to `agent_definitions`, and to check `\df`/`\dx` for an existing DB helper or extension (pgvector, pgcrypto, pg_trgm all confirmed present) before hand-writing SQL — reuse-before-recreate applied at the database layer.
- **sources:** NOTES_running_synthesis_principles(59) DB discipline section; NOTES_running_synthesis_v3(32).md DECISIONS.
- **relations:** Model-written SQL guard model; schema-before-SQL discipline.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### QueryDatabaseAction parameterised queries & schema-drift discipline
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 001(0) Appendix A #1 "New query_database usage MUST use $1 placeholders with 'params' array. Never embed values via {{.field}} in SQL — SQL injection risk"; #18 "Schema column renames — always check the live schema"
- **what:** `QueryDatabaseAction` supports `$1` placeholders via a `params` config array (never Go-template interpolation into SQL); the live DB is the source of truth for column names (dumps drift), and best-effort/fire-and-forget writes silently no-op on schema mismatch. Includes the `to_jsonb('...'::text)` cast rule for updating prompt templates.
- **sources:** WM/001_development_guide(0).md#1-querydatabaseaction-doesnt-support-parameterised-queries, WM/001_development_guide(0).md#18-schema-column-renames-always-check-the-live-schema, WM/001_development_guide(0).md#15-postgresql-to_jsonb-fails-with-could-not-determine-polymorphic-type
- **relations:** debugging guide schema reminders; snapshots/revert; LLM call logging
- **verify-later:** QueryDatabaseAction; site_specs/site_work_items/component_versions schemas

## Scope-handling notes
excellent_discussions families are purely additive — latest versions contain every earlier section; no abandoned-idea deltas there (they are exploratory/aspirational by nature, "nothing decided", "synthesis spine"). The one genuine family-delta dropped concept found across both sub-trees is the "Adapter & Service Deployment Issues" section in base `016_debugging_guide.md`, absent from `016_debugging_guide_v2_44.md` (captured above). The working/main docs are archive copies whose live successors sit in `docs024_key_docs_latest` (001/016/007/028/029/030/031/033 all tagged superseded/partial per their in-doc phase status vs the live docs already captured in earlier units). This unit overlaps material also touched by U01/U02/U09/U10/U16 (live docs024 + docs019 design/plans) — consolidation should de-duplicate accordingly, retaining this unit's unique value: the MASTER/FOCUS reasoning-architecture concepts (salience, mediator, standards curation, authored/derived context) which have no other extraction unit covering them, since the rest of docs019/_archive/excellent_discussions was not otherwise in scope.

<!-- SOURCE: U17b_docs019_gofiles.md -->
### snapshot-before-mutate practice (snapshot_agent / take_site_snapshot)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** pasted `\df`-style listing shows both functions already exist: `snapshot_agent(uuid, p_agent_type text[, p_reason text])` and `take_site_snapshot(uuid, p_site_id, p_trigger, p_git_sha, p_label, p_created_by)`; also invoked directly in the code-indexer SQL comment ("Snapshot before re-applying to an existing row: SELECT snapshot_agent('code-indexer');")
- **what:** A working convention of snapshotting an agent_definitions row (or a site) via a DB function before applying a migration that touches it, so changes are reversible without relying on git history alone. Paired with a documentation-discipline request to "start or update a runbook, a running notes and a plan" alongside any such migration.
- **sources:** contextkit/001_more_potential_thin_slice_prompt.md, NNN_create_code_indexer_agent.sql
- **relations:** code-indexer agent (applies this practice), site-snapshots-and-revert (014)
- **verify-later:** `\df snapshot_agent` / `\df take_site_snapshot` against the live clients_db schema to confirm current signatures

<!-- SOURCE: U18_sql_for_agents.md -->
### Migration discipline: schema_migrations ledger, snapshot_agent, migration_backups
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 124 "APPLIED 2026-07-10... From this file onward, every numbered file... is applied via scripts/migration/run-migrations.sh... Files 001–123... are HISTORY, not pending work; the runner's baseline (124) excludes them." Backfill rows include a lost-file reconstruction (128).
- **what:** The operational regime for this directory: schema_migrations records WHAT ran WHEN (filename PK, checksum, notes); snapshot_agent(type, reason) is the standing rule opening every agent-updating transaction (MVCC before-image); migration_backups holds manual before-values; 107's backup preamble adds the no-DROP rule ("The collision IS the safety net"). Workflow-altering migrations must leave a pipeline doc_note (runbook §3, seen in 141/142/144/146).
- **sources:** 124_schema_migrations.sql; 107_image_build_handler.sql; 131_tool_generator_plan_writing.sql; 128_fix_load_runtime_error_step_target.sql
- **relations:** travelling docs; versioned agent_definitions (UNIQUE on type+version, 121)
- **verify-later:** run-migrations.sh; snapshot_agent function; schema_migrations contents

<!-- SOURCE: U18_sql_for_agents.md -->
### client_system schema for agent instances
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 073 drops wrong tables and recreates agent_instances/agent_spawn_history "matching what create_client_schema and spawn_agent expect".
- **what:** Per-client schema tables backing spawn_agent: agent_instances (template_id → agent_definitions, project FK, config) and agent_spawn_history (parent/spawned lineage). Documents the column contract spawn_agent expects.
- **sources:** 073_create_new_client_schema.sql
- **relations:** spawn_agent action; create_client_schema; multitenancy/client schemas (docs 011)
- **verify-later:** create_client_schema function vs 073 shapes

## Proposed NEW categories
- **NEW:build-pipeline** — the site_work_items work-item build pipeline and its builder/handler agents (pageflow-builder, site-work-orchestrator, dispatch loop, page-build-handler, page-content-writer, page-rebuild). Distinct from improvement-loop (post-build) and site-plan-and-reconciler (planning domain); large enough to back a council agent.
- **NEW:rag-retrieval** — shared knowledge_base pgvector store, rag_index/rag_lookup actions, collections (tool_docs, industry_sites), embedding-model management. Not covered by model-infrastructure (endpoints/GPUs) or documentation-system.

<!-- SOURCE: U19_sql_tables_components.md -->
### Database cleanup and log retention policy
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** database-cleanup task (agent_error_log 14/30 days, audit last 100k, orchestrations 7 days/24h stuck, orchestration_requests FK made CASCADE) plus per-table cleanup functions (llm 90/180, http 90/180, awaited 7 days) and the always-return-a-row HAVING fix.
- **what:** A uniform retention discipline: every high-churn operational table has an explicit cleanup function or scheduled CTE with distinct retention for successes vs errors, and the cleanup task itself is written to always mark itself executed. Includes the FK CASCADE fix required so orchestration deletion cascades to requests.
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#database-cleanup; docs/agent_docs/sql_for_tables/024_database_cleanup.sql; docs/agent_docs/sql_for_tables/026_http_request_log.sql#cleanup
- **relations:** scheduler pre_query pattern; agent_error_log; llm/http logs.
- **verify-later:** database-cleanup enabled and returning rows.

<!-- SOURCE: U19_sql_tables_components.md -->
### Migration discipline: pre-change snapshots, renumbering, footguns
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Recurring practice across files: CREATE TABLE ... AS SELECT snapshots inside the txn (pages_bak_retype_guides, assets_backup_20260508...), NNN placeholders with "confirm the next free migration number" notes (048, 046, 047), deliberate plain CREATE TABLE to error on shape mismatch ("the migration-110 trap, §6.1"), and "code shipped but migration unapplied has bitten this project repeatedly" (042 ssh_port).
- **what:** The project's migration conventions as embodied in the files: snapshot rows before destructive UPDATEs with pasted rollback SQL; verify blocks (DO $$ ... RAISE EXCEPTION) inside transactions; idempotence via IF NOT EXISTS except where silent no-op would hide a shape conflict; migration numbers confirmed against the live runner before applying; migrations applied separately from code deploys.
- **sources:** docs/agent_docs/sql_for_tables/048_NNN_create_code_symbols_index.sql#BEFORE-APPLYING; docs/agent_docs/sql_for_tables/003_pages.sql#snapshots; docs/agent_docs/sql_for_tables/042_thunder.sql#ssh_port
- **relations:** debugging guide §6.1 / item 17; agent snapshot/revert.
- **verify-later:** migration runner and numbering source of truth.

<!-- SOURCE: U19_sql_tables_components.md -->
### Early schema inventory and since-dropped tables
- **category:** database-and-infrastructure
- **status-signal:** superseded
- **status-evidence:** 006 is a psql \dt+ snapshot listing tables absent from later docs: flow_pages, site_flows, navigation_structures, pending_requests, improvement_proposals, approval_requests, agent_groups/agent_group_definitions/agent_group_members, agent_metrics, theme_tags, system_events, event_statistics matview; 016 explicitly replaces "the navigation_structures cache table".
- **what:** A point-in-time inventory of clients_db that preserves abandoned concepts: site flows/flow pages (a flow-based site model), a navigation cache table, standalone improvement_proposals and approval_requests tables (roles later absorbed by site_work_items and input_requests), and an agent-groups mechanism. Valuable as the "what silently vanished" record.
- **sources:** docs/agent_docs/sql_for_components/006_old_summary_table_descriptions.sql; docs/agent_docs/sql_for_tables/016_nav_tables.sql
- **relations:** superseded by site_work_items, input_requests, site_nav_* tables.
- **verify-later:** whether these tables still exist in production (dead weight) or were dropped.

<!-- SOURCE: U19_sql_tables_components.md -->
### Auth database provisioning
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Raw CREATE DATABASE auth_db / CREATE USER auth_user with a subsequent password ALTER (credentials visible in file).
- **what:** A separate auth_db with its own user for the authentication service, provisioned by hand. The file preserves a real credential — a hygiene finding for stage 2 (secret in docs).
- **sources:** docs/agent_docs/sql_for_tables/021_auth_db.sql
- **relations:** database-and-infrastructure credentials; admin dashboard auth.
- **verify-later:** whether that password is still live (rotate); auth service consumer.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Database password rotation runbook (Postgres → platform secrets → PgBouncer)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** Step-by-step live commands with make targets (deploy-065-pgbouncer, pgbouncer-restart/test) and the caution about preserving other secret keys.
- **what:** The password chain has three holders: PostgreSQL users, the `personae-platform-secrets` K8s secret (read by agents), and the `pgbouncer-userlist` secret. Safe rotation order: ALTER USER in PG (existing conns keep working) → update platform secret → rebuild+restart PgBouncer userlist → test → rollout-restart agent pods.
- **sources:** docs001a_password_changing/001_changing_passwords.md
- **relations:** pgbouncer; clients_db/templates_db users; credentials handling (database-and-infrastructure docs 011).
- **verify-later:** make targets still exist; secret key inventory.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Clients → networks → sites hierarchy
- **category:** database-and-infrastructure
- **status-signal:** partial
- **status-evidence:** docs012/006 and /007 CREATE TABLE clients/networks/sites "designed for 1000s of sites, 10000s+ pages... networks of sites"; sites is heavily used later, networks/clients rarely referenced again.
- **what:** The multi-tenancy spine: clients (linked to auth-service external_id) own networks (with network-wide settings such as affiliate config), networks own sites (domain, brand_dna, github repo/branch, settings, build/deploy timestamps). Motivated by cross-site linking within a client's networks and component-level bulk updates across many pages.
- **sources:** docs012_site_maps_and_components/006_start_concluding_links.md#2.1; docs012_site_maps_and_components/004_more_on_links.md#Part-1
- **relations:** cross-site link scope; multicluster scaling; client schemas in database-and-infrastructure.
- **verify-later:** networks/clients tables — created and populated?

<!-- SOURCE: U23_docs_root_vonc.md -->
### sites.status vocabulary and the blast-radius filter trap
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** 016b merged entry: UpdateSiteStatusAction (v3_site_actions.go:323) validated vocabulary read from code; the wrong 'active' filter incident recorded.
- **what:** `sites.status` vocabulary is draft/building/review/published/deployed/archived/error ('active' is a legacy hand-written value); no code filters on it — it is an informational lifecycle label, and build dispatch keys on site_work_items (a deployed site is still rebuildable). Heuristic: never scope blast-radius or "live sites" queries with status='active' (it silently dropped the site under investigation); enumerate GROUP BY status first. Companion reuse-gate lesson: a shared set_updated_at() trigger function already exists — check pg_proc/pg_trigger before creating.
- **sources:** docs/016b_debugging_guide_merged(3).md#sites.status
- **relations:** debugging doctrine (0-rows discipline); shared library blast-radius checks
- **verify-later:** UpdateSiteStatusAction vocabulary; pg_trigger set_updated_at users

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### training_exports Postgres schema (versioned dataset snapshots)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4g schema SQL + §2.4i "1,958 rows landed" with export_id `fef7be6b-…`; RUNBOOK bcd(7) §4b uses `training_exports.runs`/`rows` live (2026-06)
- **what:** Two-table schema (`runs` metadata + `rows` ChatML JSONB) chosen over S3 because 21MB–2GB fits Postgres TOAST and avoids a second storage system. A unique index on `(export_id, metadata->>'source_log_id')` blocks duplicate source rows; real-time streaming into it was considered and rejected in favour of named batch snapshots for A/B-comparable training sets. A load-bearing gotcha surfaced later: `runs.rows_exported` can disagree with the real `rows` count (export `a8484922` had rows_exported=1957 but 0 actual rows).
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.4g; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-06-2; phase5/RUNBOOK_phase_b_c_d_deploy(7).md#step-4
- **relations:** output of Flywheel A; consumed by training-data-preparer / model-trainer
- **verify-later:** training_exports.runs, training_exports.rows; flywheel_A_v3/001_training_exports_schema.sql

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### clients_db vs templates_db agent_definitions source-of-truth
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** NOTES(39) "Update — 2026-06-03 ~17:3x: CORRECTION — chassis reads clients_db, NOT templates_db … templates_db.agent_definitions has the OLD schema (NO version column) … only the 8 original website-builder agents"
- **what:** A multi-session saga establishing that the flywheel-C/rich-schema `agent_definitions` (model-trainer, gpu-provisioner, training-launcher) live and are loaded from `clients_db`, not `templates_db`. Migration 103 first mis-applied to clients_db, then "corrected" to templates_db, then re-corrected: the chassis loader query filters `is_snapshot`/`version` columns that exist only in clients_db's rich schema. The 002 architecture doc's "source of truth is templates_db" refers to the old website-builder catalog.
- **sources:** phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-03-15:31, #update-2026-06-03-17:3x
- **relations:** every Phase 5 migration targets clients_db; contradicts the frozen 002_system_architecture pack copy
- **verify-later:** clients_db.agent_definitions vs templates_db.agent_definitions; migration 103

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### setup.sh box provisioning script
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-12 "Provisioning succeeded end-to-end … engine ACTIVE, unit + deploy hook + prune timer installed, nginx OK"; runbook(12) §3.5.
- **what:** Multi-vhost box installer: per-domain nginx server_name blocks (serve static webroot + proxy the four API paths), per-domain webroot certbot (graceful HTTP fallback, idempotent re-run upgrades to HTTPS), systemd unit, ufw/fail2ban/logrotate/unattended-upgrades/ssh-hardening, the deploy sudo hook, and `site-engine-prune.timer`. Params are env-vars: DOMAINS, LETSENCRYPT_EMAIL, DEPLOY_USER, ENGINE_BINARY_PATH, WEBROOT_OWNER, WWW_ALIAS, RETENTION_DAYS, CLOUDFLARE, MODE=full/update. Add a domain = extend DOMAINS + re-run (idempotent).
- **sources:** traffic_probe_runbook(12).md#3.5, traffic_probe_running_notes(27).md#2026-06-10-box-setup-artifact, traffic_probe_running_notes(27).md#2026-06-13-g
- **relations:** origin of the deploy privilege model; WWW_ALIAS/CLOUDFLARE/RETENTION shipped incrementally
- **verify-later:** deploy_setup/vm-deploy/setup.sh (live tree)

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Deploy privilege model (site-engine-deploy sudo hook)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-10 "Privilege model (low-risk): no root key in CI … a sudoers rule scoped to ONLY that script".
- **what:** No root key in CI. When `DEPLOY_USER` is set, setup.sh installs `/usr/local/sbin/site-engine-deploy` (root-owned: atomic binary swap + restart) and a sudoers rule scoped to only that script. The deploy user can swap the engine and nothing else; the swapped binary runs as the unprivileged `site-engine` user.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-engine-deploy-workflow, traffic_probe_runbook(12).md#5
- **relations:** part of engine Action; retired later by P5 adapter holding the SSH credential
- **verify-later:** setup.sh DEPLOY_USER branch; sudoers site-engine-deploy

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### VM sizing / Hetzner box selection
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "VM sizing (relojistas, its own box): Hetzner CX22-class"; 2026-06-12 "Box provisioned: Hetzner CPX22 #140056673, nbg1 … IP 167.233.33.159, €11.39/mo".
- **what:** Boxes are sized by disk/log headroom not CPU (static nginx + O(1) JSONL appends are far inside a small box even at claimed 1.2M visits/mo). EU-only Hetzner (20 TB/mo included) — runbook standardises on CX23 (~€3.49/mo). x86-only caveat: the engine Action builds GOARCH=amd64, so Arm (CAX) would need a build-matrix change. relojistas has its own box; small domains share a multi-vhost box.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-store-v2, traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_running_notes(27).md#2026-06-12-operator-execution, traffic_probe_runbook(12).md#3.2
- **relations:** dedicated-box vs shared-box decision
- **verify-later:** relojistas_notes coordinates; Hetzner CPX22 #140056673

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### core client→network→site→page content hierarchy (early MVP schema)
- **category:** database-and-infrastructure
- **status-signal:** superseded
- **status-evidence:** Filed as an "MVP Migration" ("Designed for patch-only updates from the start") establishing `clients`/`networks`/`sites`/`site_flows`/`pages`/`flow_pages`/`page_components` with a minimal column set (e.g. `pages` here has no `sections`, `build_status` as later became standard, no `site_plan_*` linkage) — an early snapshot of a hierarchy that has since grown substantially richer elsewhere in the live schema.
- **what:** The foundational multi-tenant hierarchy this platform is built on: `clients` (external_id linking to auth-service) → `networks` (affiliate/network-wide settings) → `sites` (domain, brand_dna, github_repo/branch) → `site_flows` (multi-track audience journeys with a narrative_arc) → `pages` (page_type, nav ordering, content_hash for change detection) → `page_components` (template instances with rendered_html, content_data, and a semantic `data_path`/`data_uuid` addressing scheme intended for future granular editing).
- **sources:** docs/_archive/agent_docs/sql_for_tables/002_links_clients_networks_etc_tables.sql
- **relations:** link registry + navigation cache (below, same migration file); system-architecture; site-plan-and-reconciler (the later, richer plan/reconciler layer this hierarchy predates)
- **verify-later:** the current `clients`/`networks`/`sites`/`pages` schema shape vs. this early version — confirm which columns/tables here are still live as originally designed vs. superseded by site_plans/site_work_items

<!-- SOURCE: U25_leopardess_social.md -->
### No tenant isolation today; dedicated-cluster-per-client as offered capability
- **category:** database-and-infrastructure
- **status-signal:** partial
- **status-evidence:** AUDIT P5: "Single shared Postgres (no row-level security anywhere in the schema), single shared Kafka, single shared ollama-adapter pod — separated only by a site_id column"; cross-cluster scaffolding "exists (remote-job-spawner, DispatchAgentAction), just not exercised in production".
- **what:** The platform has no per-client isolation — a due-diligence-relevant fact recorded as a landmine so copy never implies otherwise. Real isolation is positioned as buildable: a dedicated cluster per client, with existing but unexercised cross-cluster dispatch scaffolding. UK/EU residency is also not true end-to-end today (compute UK/Rackspace; storage Backblaze us-east-005; cloud models US).
- **sources:** docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P5/P6; docs/leopardessconsulting/RUNBOOK.md#landmine-11
- **relations:** multicluster; data-sovereignty positioning; UK-sovereign stack exploration
- **verify-later:** RLS absence in schema; remote-job-spawner usage history

<!-- SOURCE: U26_misc_dirs.md -->
### Three-database architecture (MySQL auth + PG clients + PG templates)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** databases.md is a factual "Database Architecture Summary" (not a proposal); basic_usage/001 contains live credentials/connection commands for both MySQL (rs17.uk-noc.com) and postgres-clients; current taxonomy (011) still lists MySQL + client schemas.
- **what:** Authentication isolated in MySQL (users, JWT refresh tokens, profiles, projects, subscriptions/tiers, permissions, activity logs; BINARY(16) UUIDs); agent/AI runtime in PostgreSQL clients DB (global agent_definitions, orchestrator_state, clients_info + per-client schemas); shared persona templates in a second PostgreSQL DB. Core Manager owns clients/templates access with read-only auth access.
- **sources:** docs/architecture/databases.md; docs/basic_usage/001basic_usage.txt
- **relations:** schema-per-client multi-tenancy; AI Persona Platform API (auth endpoints)
- **verify-later:** current DB inventory; whether templates DB still exists separately

<!-- SOURCE: U26_misc_dirs.md -->
### Schema-per-client multi-tenancy
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** databases.md documents `create_client_schema()` and per-client tables as implemented; every operational query in basic_usage targets `client_demo_client.*` schemas.
- **what:** Each client gets an isolated PostgreSQL schema (client_{id}) containing agent_instances, agent_memory (pgvector embeddings), projects, workflow_executions and usage_analytics, created by a SQL function; global resources (agent_definitions, templates, orchestrator_state) are shared. Strong tenant isolation on shared infrastructure.
- **sources:** docs/architecture/databases.md#2-postgresql-database-1; docs/basic_usage/001basic_usage.txt
- **relations:** three-database architecture; agent spawning (instances live per-schema)
- **verify-later:** create_client_schema function; pgvector agent_memory usage in current code
