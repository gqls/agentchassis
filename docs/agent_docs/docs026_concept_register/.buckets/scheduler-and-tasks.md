
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Build pipeline trigger: 30s heartbeat, fire-and-forget, one item per dispatch orchestration
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 002(4) resolved decisions 20–21; 010 seed schedules
- **what:** build-pipeline-trigger fires via kafka-scheduler, seeds queue, picks one dispatchable site (skipping sites with claimed items via NOT EXISTS), spawns build-dispatch-loop with await_response:false. Loop claims atomically, processes one item, completes — parallel sites, no batch accumulation, no OOM.
- **sources:** 002(4)#Dispatch Loop and Pipeline Trigger; 004#Entry Points
- **relations:** site-excluded-by-stuck-claim failure; scheduler concurrency groups
- **verify-later:** build-pipeline-trigger pre_query; find_dispatchable_site SQL

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Kafka scheduler (DB-driven heartbeat service)
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 010 full deployment reference (migration 066, kustomize, terraform paths)
- **what:** Single-replica Go producer-only service ticking 30s over scheduled_tasks: interval elapsed + concurrency-group capacity + pre_query gating → publish standard orchestrate message (from kafka-scheduler identity, responses to system.scheduler.responses — currently unconsumed). Adding a schedule is an INSERT. Pre-queries provide dynamic input (first row merged into input_data) and gating (no rows = skip). timeout_seconds is the in-flight safety valve; double-fire tolerated via idempotent work-item dedup.
- **sources:** 010 full
- **relations:** build-pipeline-trigger; improvement-sweep; med tasks; batch submitter/retriever placement
- **verify-later:** scheduled_tasks rows; cmd/scheduler/main.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### content-feed-trigger workflow shape bug (array vs object count)
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** "Fix applied … output_format = 'object' ✓ items_field = 'news_sites.rows' ✓ … Pending verification on next fire" (2026-04-20)
- **what:** The scheduled news trigger was "broken for weeks" not because of routing (generic-agent routing works as designed) but because find_news_sites returned a bare array: check_has_sites read `.count` off an array (empty string → default branch), and the loop crashed on nil when no sites existed. Fixed by output_format object + items_field .rows. General lesson: condition fields need the object {rows,count} shape.
- **sources:** HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md#7
- **relations:** owner_agent_type observability gap (why it was misdiagnosed)
- **verify-later:** content-feed-trigger definition current shape

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Work-item claim/retry behaviour and the claim-timeout class
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** W6 FINAL VERIFY: "3.1 failure class: `Claim timed out — handler pod likely died` on all three retried items — dispatch infrastructure, not the template changes; retries recovered."
- **what:** Build items are claimed by the dispatch loop and retried on claim timeout; heavy page builds (19:18–22:45 for six pages) collide with claim durations, producing retried-then-complete items whose error text is retained — read the error class before calling retries healthy. Observed hygiene gaps: `site_work_items.updated_at` stays frozen at insert through claim/retry/completion (same family as the pre-trigger layouts.updated_at); a deploy can release claims mid-flight (claimed→triaged). All parked on the hygiene list, not actioned in-thread.
- **sources:** RUNBOOK_scheme_to_components(50).md#W6-FINAL-VERIFY; w6_03_final_verify.sql; running_notes_scheme_to_components(55).md#Te #Tf #Tp
- **relations:** work-item crafting conventions; debugging (pod health).
- **verify-later:** build dispatch loop claim timeout vs typical build durations; updated_at handling on site_work_items.

<!-- SOURCE: U05_content_quality_linking.md -->
### Dispatch throughput constraints (one-site-per-tick, NOT-EXISTS freeze)
- **category:** scheduler-and-tasks
- **status-signal:** unknown
- **status-evidence:** running_notes_14(26) Part 9 confirms the mechanism; HANDOFF_2026-06-15(2) §5: "Rebuild pipeline takes MANY HOURS … NOT investigated".
- **what:** The build-dispatch-loop is one-site-per-tick (LIMIT 1, spawned per scheduler tick, ~5 items then exits) and excludes a site entirely while ANY of its items is claimed — so items serialise within a site and a dead handler freezes the whole site for the claim-timeout window. Catalogued as Family J with candidate levers (per-site bounded concurrency, per-item exclusion, shorter reaper window, trigger cadence) plus the standing speed-up TODO (batches take hours; single index rebuild ~610–770s). Parked, never closed in these docs.
- **sources:** running_notes_14(26).md#part-9; HANDOFF_2026-06-15(2).md#5; running_notes_17(21).md#missing-game
- **relations:** claimed-item-timeout reaper; operational rule "don't roll the chassis image while a batch drains".
- **verify-later:** build-dispatch-loop pre_query/LIMIT + NOT-EXISTS clause; scheduled_tasks build-pipeline-trigger cadence.

<!-- SOURCE: U11_traffic_probe.md -->
### Scheduler fires one message per tick — pre_query is a gate, not fan-out
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-13(c): "DESIGN CORRECTED by a real finding: the scheduler fires ONE message per tick — it does NOT fan out pre_query rows (ctx line 5236; thunder-monitor does in-agent loop fan-out, not scheduler fan-out)."
- **what:** A platform fact established from live source and used to correct the collector design: scheduled_tasks.pre_query does not produce per-row dispatch; the live improvement-sweep/thunder-monitor pattern is a count>0 GATE with the fired agent doing in-agent loop fan-out. The intent collector was rewritten from "collect one site from input" to a single self-querying loop-all action accordingly (complexity in Go, one-step workflow); the migration's per-row pre_query was superseded. Also the thunder-monitor convention: INSERT scheduled tasks DISABLED until the action is deployed.
- **sources:** traffic_probe_running_notes(28).md#2026-06-13-c, intent_events_migration(1).sql#scheduled-collector (gate form), deploy_setup/working_dir/intent_events_migration.sql (family-delta: superseded fan-out form)
- **relations:** intent collection topology, scheduler-and-tasks doc 010
- **verify-later:** kafka-scheduler dispatch code path (one fire per tick)

<!-- SOURCE: U12_docs024_archives.md -->
### CTE-only scheduled tasks pattern ("Always Return a Row" rule)
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive `011b_scheduler_and_tasks_guide.md` (a later revision than `011_kafka_scheduler_guide.md`, which is byte-identical to live) has a full section on this; none of it appears in live `010_scheduler_and_tasks.md`.
- **what:** Some scheduled tasks do their real work directly inside the pre_query's CTEs rather than triggering an agent — but the scheduler still requires the SELECT to return at least one row, or `last_triggered_at`/`last_completed_at` never advance, silently breaking firing cadence and concurrency-group accounting. This is a documented, previously-hit production bug pattern completely absent from the current live scheduler doc.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"Pre-Queries", #"The fire_message Column"; docs024_key_docs_latest/010_scheduler_and_tasks.md (confirmed absent)
- **relations:** concurrency-group starvation; last_completed_at ownership
- **verify-later:** `SELECT name, pre_query FROM scheduled_tasks WHERE fire_message = false` for current CTE-only tasks.

<!-- SOURCE: U12_docs024_archives.md -->
### Concurrency group starvation problem and prevention rules
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive documents a real incident ("the original maintenance group had both claimed-item-timeout and database-cleanup. When database-cleanup stalled, it blocked claim resets, which blocked the entire pipeline") and gives four prevention rules; entirely absent from live doc.
- **what:** Tasks sharing a `concurrency_group` can starve each other if one never updates `last_completed_at`, permanently occupying the group's `max_concurrent` slot. Prevention: set `timeout_seconds < interval_seconds`, never group unrelated tasks together, ensure every completion path updates `last_completed_at`.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"The Group Starvation Problem", #"Known Issues & Future Work"
- **relations:** CTE-only scheduled tasks pattern; last_completed_at ownership
- **verify-later:** query current `scheduled_tasks` group assignments against the archive's "Recommended Group Assignments" table.

<!-- SOURCE: U12_docs024_archives.md -->
### last_completed_at ownership contract and fire_message known-gap
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive explicitly documents: "The scheduler Go code does not currently read this column [fire_message]. It always sends a Kafka message"; none of these operational caveats appear in live doc.
- **what:** Agent-triggered scheduled tasks must include an explicit `notify_scheduler` step on every completion path to set `last_completed_at`; the scheduler itself never sets this column and never reads `fire_message`, flagged as a known low-priority gap.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"last_completed_at — Who Updates It?", #"Known Issues & Future Work"
- **relations:** CTE-only scheduled tasks pattern; concurrency group starvation
- **verify-later:** `grep -rn "fire_message" cmd/scheduler/` to check if the Go scheduler now reads this column.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Private inert pipeline statuses pattern
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** "inertness matrix scores 0 against all six sweeps in both states" (PLAN_fixloop_pilot.md §F0.1d)
- **what:** A reusable pattern for giving a new pipeline namespace statuses that no existing sweep or claim path recognizes, so it is inert "by construction" rather than by luck of anchor-site choice. The diagnose pipeline uses `awaiting_diagnosis` (queued) → `diagnosing` (in-flight), claimed atomically via `UPDATE ... FOR UPDATE SKIP LOCKED ... RETURNING` rather than the shared `claim_work_item` (which only claims `triaged|approved`). Because opting out of shared sweeps also opts out of their cleanup, the private-status loop must reap its own dead runs.
- **sources:** fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql#header, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1d
- **relations:** diagnose-dispatch-loop; pipeline-blind dispatch surfaces (discovered platform gap)
- **verify-later:** site_work_items.status values in the diagnose pipeline; reap_stuck step logic

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Pipeline-blind dispatch surfaces (discovered platform defect)
- **category:** scheduler-and-tasks
- **status-signal:** deployed (documented, not fixed — routed elsewhere)
- **status-evidence:** "Nothing in the relay filters work items by pipeline where it matters" (RUNBOOK(10)#Inherited gotchas); "Routed to the builder thread, not fixed here" (0NN_diagnose_dispatch_loop.sql#header)
- **what:** `build-dispatch-loop`'s `load_items` step and `build-pipeline-trigger`'s `find_dispatchable_site` query both lack any `item_pipeline`/pipeline filter, so any item of any pipeline on a claimable site gets dispatched to whatever handler_agent it names — this is the only reason the `maintenance` pipeline gets dispatched at all. `triage_detect_items` compounds this: it claims on `status='detected'` with no pipeline filter and rewrites `pipeline` to `'build'`, while its own comment falsely claims a filter exists. Fixing `build-dispatch-loop` naively would orphan the maintenance pipeline, so this was reported to the builder thread rather than fixed by the fix-loop team.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql#header, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 5
- **relations:** private inert pipeline statuses pattern; two intake paths disagreement
- **verify-later:** build-dispatch-loop.load_items config; build-pipeline-trigger.find_dispatchable_site query; triage_detect_items query

<!-- SOURCE: U13_docs024_small_dirs.md -->
### diagnose-dispatch-loop (automatic dispatch)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** "F0.1d — ✅ LANDED 2026-07-09, SHIPPED DISABLED" (PLAN_fixloop_pilot.md §F0.1d); "ships enabled=false on purpose"
- **what:** An `agent_definitions` orchestrator agent that claims one `awaiting_diagnosis` item on a 60s tick (via `diagnose-pipeline-trigger` scheduled task, `max_concurrent=1`), atomically moves it to `diagnosing`, spawns `diagnose-orchestrator`, and reaps its own runs older than 75 minutes as `failed`. Deliberately shipped with the scheduled task disabled until the chassis image is live and the benchmark's blinding is confirmed, since enabling it would let the loop claim and consume the benchmark item before blinding could be verified.
- **sources:** fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1d, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#CURRENT POSITION history
- **relations:** private inert pipeline statuses pattern; needs_diagnosis intake route
- **verify-later:** `scheduled_tasks.enabled` for name='diagnose-pipeline-trigger' (should still be false unless deliberately turned on)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Reaper mechanisms and the work-item-claim reaper gap
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** "Correction (2026-05-21). An earlier draft ... assumed the reapers were Go code ... They are not: the reapers are SQL pre_query entries in the scheduled_tasks table"
- **what:** Three/four reaper-like mechanisms recover stuck state at different layers: stuck-orchestration reaper (backed by scheduled_tasks SQL entries), `FailWorkItemAction`'s three retry paths, and `agent-job-cleanup` CronJob (k8s housekeeping only). The gap: no periodic sweep exists for work items stuck at `status='claimed'` when a pod dies uncleanly — `idx_swi_claimed` index exists for exactly this query but nothing uses it.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_platform_reliability_oom_and_reapers.md#Part-2, js_snippets_news_gaswholesalers/old/reapers_and_stuck_state_recovery.md
- **relations:** collected_data/OOM bloat; two rerender trigger paths
- **verify-later:** scheduled_tasks table pre_query entries, idx_swi_claimed index

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Reaper-location framing correction (superseded documentary claim)
- **category:** scheduler-and-tasks
- **status-signal:** superseded
- **status-evidence:** explicit dated correction: "An earlier draft of this section assumed the reapers were Go code in the coordinator. They are not"
- **what:** The original analysis framed all reaper logic as Go code in the chassis coordinator. A 2026-05-21 follow-up established the confirmed scheduled reapers are actually SQL `pre_query` entries in `scheduled_tasks`, with the Go on-access check being secondary.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_platform_reliability_oom_and_reapers.md, js_snippets_news_gaswholesalers/old/TODO_remaining_work_2026-05-21.md
- **relations:** Reaper mechanisms and the work-item-claim reaper gap
- **verify-later:** grep/inspect `pre_query`; `scheduled_tasks`

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Stale orchestration sweeper/reaper
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 001(0) "Timeout handling uses in-process goroutines. These die when pods restart … This is the #1 cause of pipeline stalls"
- **what:** In-process timeout goroutines die on pod restart, stranding orchestrations in AWAITING_RESPONSES. A periodic DB sweep (every 60s, `FOR UPDATE SKIP LOCKED`) classifies each expired awaited request: child completed (synthesize the lost response), child failed (forward), or no child/still-running (retry up to 3 then fail parent). The `stale-orchestration-reaper` scheduled task also fails 24h-stale orchestrations.
- **sources:** WM/001_development_guide(0).md#stale-orchestration-sweeper, WM/016_debugging_guide_v2_44.md#4, WM/007_adoption_pipeline_v3.md#known-issue-zombie-dispatch-loop-pods
- **relations:** timeout chain; work item lifecycle; awaited_requests
- **verify-later:** orchestration_states; awaited_requests; scheduled_tasks stale-orchestration-reaper

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### claimed-item-timeout & timeout chain
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** 016 v2_44 §7 "Three timeouts interact and must be ordered correctly … claim_timeout > call_handler timeout > workflow timeout"
- **what:** A `claimed-item-timeout` scheduled task resets long-claimed items; three timeouts must stay ordered, else two handlers run one item. A two-phase reset (15-min evidence-based, 40-min blind) is used; the evidence check can produce false-positive completions.
- **sources:** WM/016_debugging_guide_v2_44.md#7, WM/016_debugging_guide_v2_44.md#9, WM/007_adoption_pipeline_v3.md#implementation-fixes-schema-notes-from-028j-handoff
- **relations:** dispatch loop; stale orchestration sweeper; work item lifecycle
- **verify-later:** claimed-item-timeout pre_query; scheduled_tasks; idle_timeout_seconds

<!-- SOURCE: U19_sql_tables_components.md -->
### Kafka scheduler and scheduled_tasks
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** Table DDL (066_kafka_scheduler) plus a long operational history of seeded tasks: build-pipeline-trigger, improvement-sweep, claimed-item-timeout, feasibility-recheck, content-feed-refresh, database-cleanup, vet-*, med-*, ch-enrichment, health checks, archiver.
- **what:** Interval-based scheduling in Postgres: each row names a target agent/topic, input_data, interval_seconds, timeout, and concurrency_group/max_concurrent (group-wide in-flight cap). The scheduler publishes Kafka trigger messages; last_triggered_at/last_completed_at implement a no-refire guard (with known operational pitfalls when nothing sets last_completed_at for fire-and-forget tasks — mitigated by shorter timeout windows).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql
- **relations:** pre_query SQL-worker pattern; every pipeline's periodic trigger.
- **verify-later:** kafka-scheduler service; fire_message column semantics.

<!-- SOURCE: U19_sql_tables_components.md -->
### pre_query SQL-worker pattern and self-healing tasks
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** Iterated in place: pre_query CTE UPDATEs doing the work, then `WHERE 1=0` / `HAVING COUNT(*) > 0` variants to control whether a Kafka message fires; vet-cleanup broadened to fail stuck AWAITING_RESPONSES orchestrations and reset orphaned collection tasks.
- **what:** scheduled_tasks.pre_query is a full worker channel, not just a gate: SQL that returns rows merges into input_data and fires the message; returning zero rows skips the tick. Maintenance tasks exploit this to run entire cleanup UPDATEs inside the pre_query (claimed-item reset, blocked-item promotion, orchestration failing, database cleanup) with row-suppression idioms deciding whether anything downstream is triggered (fire_message=false for pure-SQL tasks).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-lifecycle and #vet-cleanup and #database-cleanup; docs/agent_docs/sql_for_tables/024_database_cleanup.sql
- **relations:** scheduler; claimed-item timeout; database cleanup.
- **verify-later:** scheduler's pre_query evaluation code; fire_message flag.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### thunder-training-monitor + worker (probe/classify/reconcile/decommission)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** NOTES(39) §9 "Training-monitor: BUILT + VERIFIED live 2026-06-04 (both paths) … Terminal/decommission branch still never run live … Not enabled"
- **what:** A periodic orchestrator (`thunder-training-monitor`, migration 108) that runs `find_active_training_instances → loop(spawn_worker → call_worker)` every 5 min (scheduled_tasks row, inserted DISABLED, gated pre_query). Each `thunder-training-monitor-worker` (migration 107) probes a box via the adapter's `ssh_get_status`, classifies run.sh markers (ALIVE/DONE_OK/DONE_FAIL/GONE_UNKNOWN) via `classify_training_probe`, reconciles `training_runs` via `mark_training_run_terminal`, and decommissions on terminal verdicts. Deliberately separate from the reaper (different dependencies); closes the running→complete/failed reconcile gap. Enabling it is gated on the upload path proving DONE⟹durable.
- **sources:** phase5/108_thunder_training_monitor_orchestrator.sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150, #update-2026-06-04-1x; docubundle/.../STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** reuses ssh_get_status + dispatch_thunder_decommission; depends on unreachable counter; gated by RUNBOOK step 6
- **verify-later:** agent_definitions thunder-training-monitor (c3b4c052) / -worker (470c6b3f); 5 actions incl find_active_training_instances; scheduled_tasks

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Thunder unreachable-probe counter
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 106 SQL header "counts CONSECUTIVE unreachable probes and only treats the instance as 'lost' … once the count crosses a threshold"
- **what:** Migration 106 adds `consecutive_unreachable_probes` + `last_probe_at` to `thunder_instances` so the monitor can distinguish a transient SSH blip from a truly-lost box. Each scheduler tick is a fresh sub-agent that can't hold count in memory, so the streak lives on the row: the `record_probe_streak` action bumps on unreachable (route to lost/decommission at threshold, default 3) and resets to 0 on any reachable probe.
- **sources:** phase5/106_thunder_unreachable_counter.sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150 (Counter step)
- **relations:** part of thunder-training-monitor; keeps the classifier action pure
- **verify-later:** thunder_instances.consecutive_unreachable_probes; record_probe_streak_action.go

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### P4 off-box collection (intent_events + CollectIntentEventsAction)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(d) "Migration applied (CREATE TABLE + 3 indexes + INSERT 0 1 task)"; action "VERIFIED against live source + registered"; but agent deploy fields still to confirm and enable order pending.
- **what:** The cluster pulls intent over key-gated HTTPS with NO adapter and NO SSH. `intent_events` table (engine_event_id UNIQUE = structural idempotency, CHECK on kind/value len, host→site_id resolve, checkpoint = max(event_created_at) with no extra storage). `collect_intent_events` is a SINGLE Go action that self-queries all VM backend sites and loops (parameterised upserts), registered in GlobalActionRegistry (Category "data", IsLocal). Ingest contract: parameterised SQL only, per-line shape checks, burst dedupe, NFC normalisation + lowercasing here.
- **sources:** traffic_probe_plan(11).md#p4, traffic_probe_running_notes(27).md#2026-06-13-b, traffic_probe_running_notes(27).md#2026-06-13-c
- **relations:** driven by intent-collection-orchestrator/intent-collector agents; extended with collectSiteStats + access-digest pull
- **verify-later:** intent_events_migration.sql; intent_collector_actions.go; registry.go DATA region

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Superseded checkpoint-JSON / events-per-1k ranking (early P4)
- **category:** scheduler-and-tasks
- **status-signal:** superseded
- **status-evidence:** plan(1)/(4)/(5) P4 "checkpoint JSON, compute events-per-1k, rank domains"; plan(11) now "idempotent via unique engine_event_id; no extra checkpoint storage — since=max(event_created_at)".
- **what:** Early P4 phrasing planned an explicit checkpoint-JSON file to track collection progress and a direct events-per-1k rank. Dropped in favour of structural idempotency (unique engine_event_id) with the checkpoint derived as since=max(event_created_at) — no extra checkpoint storage. Ranking became a set of read-only SQL queries.
- **sources:** traffic_probe_plan(4).md#phases, traffic_probe_plan(1).md#phases, traffic_probe_plan(11).md#p4
- **relations:** replaced by intent_events unique-id design + intent_ranking_queries
- **verify-later:** intent_events.engine_event_id UNIQUE

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### intent-collection-orchestrator + intent-collector agents
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** intent_collector_agents SQL headers "intent-collection-orchestrator + intent-collector (P4) … mirror the LIVE med-export-orchestrator / med-json-exporter pair verbatim"; running_notes 2026-06-13(g) INSERT bug fixed.
- **what:** A thin wrapper-orchestrator (spawn_collector → call_collector → complete, no substantive in-chassis work) that spawns the `intent-collector` task worker (one step: collect_intent_events, processing_mode "task"). Infra fields (image docker.io/aqls/agent-chassis v1.0.1063, resources, health_config, business-intel topics, delegation) copied verbatim from the med-export pair. Reached by the scheduler via target_topic=system.agent.generic.requests by agent_type. Idempotency uses `ON CONFLICT (type, version)`.
- **sources:** deploy_setup/working_dir/intent_collector_agents(3).sql#header, deploy_setup/working_dir/intent_collector_agents(1).sql, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** identical to live intent_collector_agents(2).sql; wrapper-orchestrator requirement; replaces a single in-pod collector
- **verify-later:** agent_definitions rows intent-collection-orchestrator / intent-collector

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Retention prune timer
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "Added to setup.sh: RETENTION_DAYS param (default 90) + site-engine-prune.service/.timer (daily find-delete of old events-*.jsonl)".
- **what:** Because daily JSONL IS the rotation, logrotate on engine files would race the open handle; instead setup.sh installs a `site-engine-prune` systemd service+timer that daily find-deletes `events-*.jsonl` older than RETENTION_DAYS (default 90). nginx logs keep their existing size-based logrotate.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_runbook(12).md#3.5
- **relations:** part of the privacy posture
- **verify-later:** setup.sh site-engine-prune.timer; RETENTION_DAYS

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### claimed-item-timeout evidence-gated completion + reset (Lever A/C) — avoided building a duplicate watchdog
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running_notes_14(20) Part 12: "'Auto-completed...' is set by the **`claimed-item-timeout` scheduled task's SQL `pre_query`**, NOT a Go reaper... `migration_claimed_item_timeout_evidence_check.sql` ... is essentially Option A + Lever C, already authored... the FOCUS_dispatch `reset_stale_claims` watchdog is redundant; do NOT build it." Part 12 addenda confirm the v2 migration (page_components-based evidence, not the untrustworthy `build_status='deployed'` flag) applied and verified live, plus the companion `pageHasComponents` deploy-guard (Option B) delivered.
- **what:** A `claimed-item-timeout` scheduled task's `pre_query` already implements both (a) evidence-gated auto-completion of stuck claims (only complete if the specific artefact shows positive evidence) and (b) a stale-claim reset-to-`triaged`/`failed` after a timeout with attempt counting. Mid-investigation, an agent nearly built a brand-new "reset stale claims" watchdog before discovering this — a documented reuse-over-build catch. The evidence signal itself evolved further: from trusting `pages.deployed_at`/`build_status='deployed'` (provably lying, per the homepage case) to checking `page_components.updated_at > claimed_at` directly.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 12–12 addenda
- **relations:** A4 homepage root cause (above); sectionless-page durability stack (below); dispatch throughput (Family J)
- **verify-later:** current `claimed-item-timeout` pre_query SQL in `scheduled_tasks`.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Dispatch throughput bottleneck (Family J) — one-site-per-tick, NOT-EXISTS-blocked
- **category:** scheduler-and-tasks
- **status-signal:** unknown
- **status-evidence:** CATALOGUE_gamesdesign_post_sync_fix_defects(4).md Family J: "the dispatcher is **one-site-per-tick** (selection `LIMIT 1`... processes ~5 items then exits) and **NOT-EXISTS-blocked** (a `NOT EXISTS` clause excludes a site *entirely* while any of its items is `status='claimed'`... line 276)... Standard manual unstick for now... **To investigate in the separate thread.**"
- **what:** Multi-tool/multi-game adoption sites drain over hours, appearing stalled, because the build-dispatch mechanism processes one site per scheduler tick and blocks an entire site's queue while any single item on it is claimed (no bounded concurrency, no per-item exclusion). A dead handler leaving a stale claim freezes the whole site until a reaper resets it. Explicitly spun out as a separate, not-yet-investigated thread rather than fixed within this arc; running_notes_17(16) later notes it's still an open TODO ("SPEED UP the rebuild pipeline... Not yet investigated").
- **sources:** adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md Family J; running_notes_14(20) Part 9; running_notes_17(16) "Missing-game... + speed TODO"
- **relations:** claimed-item-timeout evidence-check reliability mini-project; A1 tool/game deploy gap
- **verify-later:** `build-pipeline-trigger` dispatcher current selection logic (`LIMIT 1`, NOT EXISTS clause, line ~276 at time of writing).

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Build pipeline trigger: 30s heartbeat, fire-and-forget, one item per dispatch orchestration
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 002(4) resolved decisions 20–21; 010 seed schedules
- **what:** build-pipeline-trigger fires via kafka-scheduler, seeds queue, picks one dispatchable site (skipping sites with claimed items via NOT EXISTS), spawns build-dispatch-loop with await_response:false. Loop claims atomically, processes one item, completes — parallel sites, no batch accumulation, no OOM.
- **sources:** 002(4)#Dispatch Loop and Pipeline Trigger; 004#Entry Points
- **relations:** site-excluded-by-stuck-claim failure; scheduler concurrency groups
- **verify-later:** build-pipeline-trigger pre_query; find_dispatchable_site SQL

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Kafka scheduler (DB-driven heartbeat service)
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 010 full deployment reference (migration 066, kustomize, terraform paths)
- **what:** Single-replica Go producer-only service ticking 30s over scheduled_tasks: interval elapsed + concurrency-group capacity + pre_query gating → publish standard orchestrate message (from kafka-scheduler identity, responses to system.scheduler.responses — currently unconsumed). Adding a schedule is an INSERT. Pre-queries provide dynamic input (first row merged into input_data) and gating (no rows = skip). timeout_seconds is the in-flight safety valve; double-fire tolerated via idempotent work-item dedup.
- **sources:** 010 full
- **relations:** build-pipeline-trigger; improvement-sweep; med tasks; batch submitter/retriever placement
- **verify-later:** scheduled_tasks rows; cmd/scheduler/main.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### content-feed-trigger workflow shape bug (array vs object count)
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** "Fix applied … output_format = 'object' ✓ items_field = 'news_sites.rows' ✓ … Pending verification on next fire" (2026-04-20)
- **what:** The scheduled news trigger was "broken for weeks" not because of routing (generic-agent routing works as designed) but because find_news_sites returned a bare array: check_has_sites read `.count` off an array (empty string → default branch), and the loop crashed on nil when no sites existed. Fixed by output_format object + items_field .rows. General lesson: condition fields need the object {rows,count} shape.
- **sources:** HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md#7
- **relations:** owner_agent_type observability gap (why it was misdiagnosed)
- **verify-later:** content-feed-trigger definition current shape

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Work-item claim/retry behaviour and the claim-timeout class
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** W6 FINAL VERIFY: "3.1 failure class: `Claim timed out — handler pod likely died` on all three retried items — dispatch infrastructure, not the template changes; retries recovered."
- **what:** Build items are claimed by the dispatch loop and retried on claim timeout; heavy page builds (19:18–22:45 for six pages) collide with claim durations, producing retried-then-complete items whose error text is retained — read the error class before calling retries healthy. Observed hygiene gaps: `site_work_items.updated_at` stays frozen at insert through claim/retry/completion (same family as the pre-trigger layouts.updated_at); a deploy can release claims mid-flight (claimed→triaged). All parked on the hygiene list, not actioned in-thread.
- **sources:** RUNBOOK_scheme_to_components(50).md#W6-FINAL-VERIFY; w6_03_final_verify.sql; running_notes_scheme_to_components(55).md#Te #Tf #Tp
- **relations:** work-item crafting conventions; debugging (pod health).
- **verify-later:** build dispatch loop claim timeout vs typical build durations; updated_at handling on site_work_items.

<!-- SOURCE: U05_content_quality_linking.md -->
### Dispatch throughput constraints (one-site-per-tick, NOT-EXISTS freeze)
- **category:** scheduler-and-tasks
- **status-signal:** unknown
- **status-evidence:** running_notes_14(26) Part 9 confirms the mechanism; HANDOFF_2026-06-15(2) §5: "Rebuild pipeline takes MANY HOURS … NOT investigated".
- **what:** The build-dispatch-loop is one-site-per-tick (LIMIT 1, spawned per scheduler tick, ~5 items then exits) and excludes a site entirely while ANY of its items is claimed — so items serialise within a site and a dead handler freezes the whole site for the claim-timeout window. Catalogued as Family J with candidate levers (per-site bounded concurrency, per-item exclusion, shorter reaper window, trigger cadence) plus the standing speed-up TODO (batches take hours; single index rebuild ~610–770s). Parked, never closed in these docs.
- **sources:** running_notes_14(26).md#part-9; HANDOFF_2026-06-15(2).md#5; running_notes_17(21).md#missing-game
- **relations:** claimed-item-timeout reaper; operational rule "don't roll the chassis image while a batch drains".
- **verify-later:** build-dispatch-loop pre_query/LIMIT + NOT-EXISTS clause; scheduled_tasks build-pipeline-trigger cadence.

<!-- SOURCE: U11_traffic_probe.md -->
### Scheduler fires one message per tick — pre_query is a gate, not fan-out
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running notes 2026-06-13(c): "DESIGN CORRECTED by a real finding: the scheduler fires ONE message per tick — it does NOT fan out pre_query rows (ctx line 5236; thunder-monitor does in-agent loop fan-out, not scheduler fan-out)."
- **what:** A platform fact established from live source and used to correct the collector design: scheduled_tasks.pre_query does not produce per-row dispatch; the live improvement-sweep/thunder-monitor pattern is a count>0 GATE with the fired agent doing in-agent loop fan-out. The intent collector was rewritten from "collect one site from input" to a single self-querying loop-all action accordingly (complexity in Go, one-step workflow); the migration's per-row pre_query was superseded. Also the thunder-monitor convention: INSERT scheduled tasks DISABLED until the action is deployed.
- **sources:** traffic_probe_running_notes(28).md#2026-06-13-c, intent_events_migration(1).sql#scheduled-collector (gate form), deploy_setup/working_dir/intent_events_migration.sql (family-delta: superseded fan-out form)
- **relations:** intent collection topology, scheduler-and-tasks doc 010
- **verify-later:** kafka-scheduler dispatch code path (one fire per tick)

<!-- SOURCE: U12_docs024_archives.md -->
### CTE-only scheduled tasks pattern ("Always Return a Row" rule)
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive `011b_scheduler_and_tasks_guide.md` (a later revision than `011_kafka_scheduler_guide.md`, which is byte-identical to live) has a full section on this; none of it appears in live `010_scheduler_and_tasks.md`.
- **what:** Some scheduled tasks do their real work directly inside the pre_query's CTEs rather than triggering an agent — but the scheduler still requires the SELECT to return at least one row, or `last_triggered_at`/`last_completed_at` never advance, silently breaking firing cadence and concurrency-group accounting. This is a documented, previously-hit production bug pattern completely absent from the current live scheduler doc.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"Pre-Queries", #"The fire_message Column"; docs024_key_docs_latest/010_scheduler_and_tasks.md (confirmed absent)
- **relations:** concurrency-group starvation; last_completed_at ownership
- **verify-later:** `SELECT name, pre_query FROM scheduled_tasks WHERE fire_message = false` for current CTE-only tasks.

<!-- SOURCE: U12_docs024_archives.md -->
### Concurrency group starvation problem and prevention rules
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive documents a real incident ("the original maintenance group had both claimed-item-timeout and database-cleanup. When database-cleanup stalled, it blocked claim resets, which blocked the entire pipeline") and gives four prevention rules; entirely absent from live doc.
- **what:** Tasks sharing a `concurrency_group` can starve each other if one never updates `last_completed_at`, permanently occupying the group's `max_concurrent` slot. Prevention: set `timeout_seconds < interval_seconds`, never group unrelated tasks together, ensure every completion path updates `last_completed_at`.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"The Group Starvation Problem", #"Known Issues & Future Work"
- **relations:** CTE-only scheduled tasks pattern; last_completed_at ownership
- **verify-later:** query current `scheduled_tasks` group assignments against the archive's "Recommended Group Assignments" table.

<!-- SOURCE: U12_docs024_archives.md -->
### last_completed_at ownership contract and fire_message known-gap
- **category:** scheduler-and-tasks
- **status-signal:** abandoned
- **status-evidence:** Archive explicitly documents: "The scheduler Go code does not currently read this column [fire_message]. It always sends a Kafka message"; none of these operational caveats appear in live doc.
- **what:** Agent-triggered scheduled tasks must include an explicit `notify_scheduler` step on every completion path to set `last_completed_at`; the scheduler itself never sets this column and never reads `fire_message`, flagged as a known low-priority gap.
- **sources:** archive_april_26/011b_scheduler_and_tasks_guide.md#"last_completed_at — Who Updates It?", #"Known Issues & Future Work"
- **relations:** CTE-only scheduled tasks pattern; concurrency group starvation
- **verify-later:** `grep -rn "fire_message" cmd/scheduler/` to check if the Go scheduler now reads this column.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Private inert pipeline statuses pattern
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** "inertness matrix scores 0 against all six sweeps in both states" (PLAN_fixloop_pilot.md §F0.1d)
- **what:** A reusable pattern for giving a new pipeline namespace statuses that no existing sweep or claim path recognizes, so it is inert "by construction" rather than by luck of anchor-site choice. The diagnose pipeline uses `awaiting_diagnosis` (queued) → `diagnosing` (in-flight), claimed atomically via `UPDATE ... FOR UPDATE SKIP LOCKED ... RETURNING` rather than the shared `claim_work_item` (which only claims `triaged|approved`). Because opting out of shared sweeps also opts out of their cleanup, the private-status loop must reap its own dead runs.
- **sources:** fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql#header, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1d
- **relations:** diagnose-dispatch-loop; pipeline-blind dispatch surfaces (discovered platform gap)
- **verify-later:** site_work_items.status values in the diagnose pipeline; reap_stuck step logic

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Pipeline-blind dispatch surfaces (discovered platform defect)
- **category:** scheduler-and-tasks
- **status-signal:** deployed (documented, not fixed — routed elsewhere)
- **status-evidence:** "Nothing in the relay filters work items by pipeline where it matters" (RUNBOOK(10)#Inherited gotchas); "Routed to the builder thread, not fixed here" (0NN_diagnose_dispatch_loop.sql#header)
- **what:** `build-dispatch-loop`'s `load_items` step and `build-pipeline-trigger`'s `find_dispatchable_site` query both lack any `item_pipeline`/pipeline filter, so any item of any pipeline on a claimable site gets dispatched to whatever handler_agent it names — this is the only reason the `maintenance` pipeline gets dispatched at all. `triage_detect_items` compounds this: it claims on `status='detected'` with no pipeline filter and rewrites `pipeline` to `'build'`, while its own comment falsely claims a filter exists. Fixing `build-dispatch-loop` naively would orphan the maintenance pipeline, so this was reported to the builder thread rather than fixed by the fix-loop team.
- **sources:** fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#Inherited gotchas, fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql#header, fixloop_eg_dartsonline/NOTES_running_fixloop(10).md#Turn 5
- **relations:** private inert pipeline statuses pattern; two intake paths disagreement
- **verify-later:** build-dispatch-loop.load_items config; build-pipeline-trigger.find_dispatchable_site query; triage_detect_items query

<!-- SOURCE: U13_docs024_small_dirs.md -->
### diagnose-dispatch-loop (automatic dispatch)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** "F0.1d — ✅ LANDED 2026-07-09, SHIPPED DISABLED" (PLAN_fixloop_pilot.md §F0.1d); "ships enabled=false on purpose"
- **what:** An `agent_definitions` orchestrator agent that claims one `awaiting_diagnosis` item on a 60s tick (via `diagnose-pipeline-trigger` scheduled task, `max_concurrent=1`), atomically moves it to `diagnosing`, spawns `diagnose-orchestrator`, and reaps its own runs older than 75 minutes as `failed`. Deliberately shipped with the scheduled task disabled until the chassis image is live and the benchmark's blinding is confirmed, since enabling it would let the loop claim and consume the benchmark item before blinding could be verified.
- **sources:** fixloop_eg_dartsonline/0NN_diagnose_dispatch_loop.sql, fixloop_eg_dartsonline/PLAN_fixloop_pilot.md §F0.1d, fixloop_eg_dartsonline/RUNBOOK_diagnosis_fix_loop(10).md#CURRENT POSITION history
- **relations:** private inert pipeline statuses pattern; needs_diagnosis intake route
- **verify-later:** `scheduled_tasks.enabled` for name='diagnose-pipeline-trigger' (should still be false unless deliberately turned on)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Reaper mechanisms and the work-item-claim reaper gap
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** "Correction (2026-05-21). An earlier draft ... assumed the reapers were Go code ... They are not: the reapers are SQL pre_query entries in the scheduled_tasks table"
- **what:** Three/four reaper-like mechanisms recover stuck state at different layers: stuck-orchestration reaper (backed by scheduled_tasks SQL entries), `FailWorkItemAction`'s three retry paths, and `agent-job-cleanup` CronJob (k8s housekeeping only). The gap: no periodic sweep exists for work items stuck at `status='claimed'` when a pod dies uncleanly — `idx_swi_claimed` index exists for exactly this query but nothing uses it.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_platform_reliability_oom_and_reapers.md#Part-2, js_snippets_news_gaswholesalers/old/reapers_and_stuck_state_recovery.md
- **relations:** collected_data/OOM bloat; two rerender trigger paths
- **verify-later:** scheduled_tasks table pre_query entries, idx_swi_claimed index

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Reaper-location framing correction (superseded documentary claim)
- **category:** scheduler-and-tasks
- **status-signal:** superseded
- **status-evidence:** explicit dated correction: "An earlier draft of this section assumed the reapers were Go code in the coordinator. They are not"
- **what:** The original analysis framed all reaper logic as Go code in the chassis coordinator. A 2026-05-21 follow-up established the confirmed scheduled reapers are actually SQL `pre_query` entries in `scheduled_tasks`, with the Go on-access check being secondary.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_platform_reliability_oom_and_reapers.md, js_snippets_news_gaswholesalers/old/TODO_remaining_work_2026-05-21.md
- **relations:** Reaper mechanisms and the work-item-claim reaper gap
- **verify-later:** grep/inspect `pre_query`; `scheduled_tasks`

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Stale orchestration sweeper/reaper
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 001(0) "Timeout handling uses in-process goroutines. These die when pods restart … This is the #1 cause of pipeline stalls"
- **what:** In-process timeout goroutines die on pod restart, stranding orchestrations in AWAITING_RESPONSES. A periodic DB sweep (every 60s, `FOR UPDATE SKIP LOCKED`) classifies each expired awaited request: child completed (synthesize the lost response), child failed (forward), or no child/still-running (retry up to 3 then fail parent). The `stale-orchestration-reaper` scheduled task also fails 24h-stale orchestrations.
- **sources:** WM/001_development_guide(0).md#stale-orchestration-sweeper, WM/016_debugging_guide_v2_44.md#4, WM/007_adoption_pipeline_v3.md#known-issue-zombie-dispatch-loop-pods
- **relations:** timeout chain; work item lifecycle; awaited_requests
- **verify-later:** orchestration_states; awaited_requests; scheduled_tasks stale-orchestration-reaper

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### claimed-item-timeout & timeout chain
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** 016 v2_44 §7 "Three timeouts interact and must be ordered correctly … claim_timeout > call_handler timeout > workflow timeout"
- **what:** A `claimed-item-timeout` scheduled task resets long-claimed items; three timeouts must stay ordered, else two handlers run one item. A two-phase reset (15-min evidence-based, 40-min blind) is used; the evidence check can produce false-positive completions.
- **sources:** WM/016_debugging_guide_v2_44.md#7, WM/016_debugging_guide_v2_44.md#9, WM/007_adoption_pipeline_v3.md#implementation-fixes-schema-notes-from-028j-handoff
- **relations:** dispatch loop; stale orchestration sweeper; work item lifecycle
- **verify-later:** claimed-item-timeout pre_query; scheduled_tasks; idle_timeout_seconds

<!-- SOURCE: U19_sql_tables_components.md -->
### Kafka scheduler and scheduled_tasks
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** Table DDL (066_kafka_scheduler) plus a long operational history of seeded tasks: build-pipeline-trigger, improvement-sweep, claimed-item-timeout, feasibility-recheck, content-feed-refresh, database-cleanup, vet-*, med-*, ch-enrichment, health checks, archiver.
- **what:** Interval-based scheduling in Postgres: each row names a target agent/topic, input_data, interval_seconds, timeout, and concurrency_group/max_concurrent (group-wide in-flight cap). The scheduler publishes Kafka trigger messages; last_triggered_at/last_completed_at implement a no-refire guard (with known operational pitfalls when nothing sets last_completed_at for fire-and-forget tasks — mitigated by shorter timeout windows).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql
- **relations:** pre_query SQL-worker pattern; every pipeline's periodic trigger.
- **verify-later:** kafka-scheduler service; fire_message column semantics.

<!-- SOURCE: U19_sql_tables_components.md -->
### pre_query SQL-worker pattern and self-healing tasks
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** Iterated in place: pre_query CTE UPDATEs doing the work, then `WHERE 1=0` / `HAVING COUNT(*) > 0` variants to control whether a Kafka message fires; vet-cleanup broadened to fail stuck AWAITING_RESPONSES orchestrations and reset orphaned collection tasks.
- **what:** scheduled_tasks.pre_query is a full worker channel, not just a gate: SQL that returns rows merges into input_data and fires the message; returning zero rows skips the tick. Maintenance tasks exploit this to run entire cleanup UPDATEs inside the pre_query (claimed-item reset, blocked-item promotion, orchestration failing, database cleanup) with row-suppression idioms deciding whether anything downstream is triggered (fire_message=false for pure-SQL tasks).
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#work-item-lifecycle and #vet-cleanup and #database-cleanup; docs/agent_docs/sql_for_tables/024_database_cleanup.sql
- **relations:** scheduler; claimed-item timeout; database cleanup.
- **verify-later:** scheduler's pre_query evaluation code; fire_message flag.

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### thunder-training-monitor + worker (probe/classify/reconcile/decommission)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** NOTES(39) §9 "Training-monitor: BUILT + VERIFIED live 2026-06-04 (both paths) … Terminal/decommission branch still never run live … Not enabled"
- **what:** A periodic orchestrator (`thunder-training-monitor`, migration 108) that runs `find_active_training_instances → loop(spawn_worker → call_worker)` every 5 min (scheduled_tasks row, inserted DISABLED, gated pre_query). Each `thunder-training-monitor-worker` (migration 107) probes a box via the adapter's `ssh_get_status`, classifies run.sh markers (ALIVE/DONE_OK/DONE_FAIL/GONE_UNKNOWN) via `classify_training_probe`, reconciles `training_runs` via `mark_training_run_terminal`, and decommissions on terminal verdicts. Deliberately separate from the reaper (different dependencies); closes the running→complete/failed reconcile gap. Enabling it is gated on the upload path proving DONE⟹durable.
- **sources:** phase5/108_thunder_training_monitor_orchestrator.sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150, #update-2026-06-04-1x; docubundle/.../STATUS_thunder_adapter_2026-06_04.md#update-2026-06-04
- **relations:** reuses ssh_get_status + dispatch_thunder_decommission; depends on unreachable counter; gated by RUNBOOK step 6
- **verify-later:** agent_definitions thunder-training-monitor (c3b4c052) / -worker (470c6b3f); 5 actions incl find_active_training_instances; scheduled_tasks

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Thunder unreachable-probe counter
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** 106 SQL header "counts CONSECUTIVE unreachable probes and only treats the instance as 'lost' … once the count crosses a threshold"
- **what:** Migration 106 adds `consecutive_unreachable_probes` + `last_probe_at` to `thunder_instances` so the monitor can distinguish a transient SSH blip from a truly-lost box. Each scheduler tick is a fresh sub-agent that can't hold count in memory, so the streak lives on the row: the `record_probe_streak` action bumps on unreachable (route to lost/decommission at threshold, default 3) and resets to 0 on any reachable probe.
- **sources:** phase5/106_thunder_unreachable_counter.sql; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-04-1150 (Counter step)
- **relations:** part of thunder-training-monitor; keeps the classifier action pure
- **verify-later:** thunder_instances.consecutive_unreachable_probes; record_probe_streak_action.go

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### P4 off-box collection (intent_events + CollectIntentEventsAction)
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(d) "Migration applied (CREATE TABLE + 3 indexes + INSERT 0 1 task)"; action "VERIFIED against live source + registered"; but agent deploy fields still to confirm and enable order pending.
- **what:** The cluster pulls intent over key-gated HTTPS with NO adapter and NO SSH. `intent_events` table (engine_event_id UNIQUE = structural idempotency, CHECK on kind/value len, host→site_id resolve, checkpoint = max(event_created_at) with no extra storage). `collect_intent_events` is a SINGLE Go action that self-queries all VM backend sites and loops (parameterised upserts), registered in GlobalActionRegistry (Category "data", IsLocal). Ingest contract: parameterised SQL only, per-line shape checks, burst dedupe, NFC normalisation + lowercasing here.
- **sources:** traffic_probe_plan(11).md#p4, traffic_probe_running_notes(27).md#2026-06-13-b, traffic_probe_running_notes(27).md#2026-06-13-c
- **relations:** driven by intent-collection-orchestrator/intent-collector agents; extended with collectSiteStats + access-digest pull
- **verify-later:** intent_events_migration.sql; intent_collector_actions.go; registry.go DATA region

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Superseded checkpoint-JSON / events-per-1k ranking (early P4)
- **category:** scheduler-and-tasks
- **status-signal:** superseded
- **status-evidence:** plan(1)/(4)/(5) P4 "checkpoint JSON, compute events-per-1k, rank domains"; plan(11) now "idempotent via unique engine_event_id; no extra checkpoint storage — since=max(event_created_at)".
- **what:** Early P4 phrasing planned an explicit checkpoint-JSON file to track collection progress and a direct events-per-1k rank. Dropped in favour of structural idempotency (unique engine_event_id) with the checkpoint derived as since=max(event_created_at) — no extra checkpoint storage. Ranking became a set of read-only SQL queries.
- **sources:** traffic_probe_plan(4).md#phases, traffic_probe_plan(1).md#phases, traffic_probe_plan(11).md#p4
- **relations:** replaced by intent_events unique-id design + intent_ranking_queries
- **verify-later:** intent_events.engine_event_id UNIQUE

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### intent-collection-orchestrator + intent-collector agents
- **category:** scheduler-and-tasks
- **status-signal:** partial
- **status-evidence:** intent_collector_agents SQL headers "intent-collection-orchestrator + intent-collector (P4) … mirror the LIVE med-export-orchestrator / med-json-exporter pair verbatim"; running_notes 2026-06-13(g) INSERT bug fixed.
- **what:** A thin wrapper-orchestrator (spawn_collector → call_collector → complete, no substantive in-chassis work) that spawns the `intent-collector` task worker (one step: collect_intent_events, processing_mode "task"). Infra fields (image docker.io/aqls/agent-chassis v1.0.1063, resources, health_config, business-intel topics, delegation) copied verbatim from the med-export pair. Reached by the scheduler via target_topic=system.agent.generic.requests by agent_type. Idempotency uses `ON CONFLICT (type, version)`.
- **sources:** deploy_setup/working_dir/intent_collector_agents(3).sql#header, deploy_setup/working_dir/intent_collector_agents(1).sql, traffic_probe_running_notes(27).md#2026-06-13-f
- **relations:** identical to live intent_collector_agents(2).sql; wrapper-orchestrator requirement; replaces a single in-pod collector
- **verify-later:** agent_definitions rows intent-collection-orchestrator / intent-collector

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Retention prune timer
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "Added to setup.sh: RETENTION_DAYS param (default 90) + site-engine-prune.service/.timer (daily find-delete of old events-*.jsonl)".
- **what:** Because daily JSONL IS the rotation, logrotate on engine files would race the open handle; instead setup.sh installs a `site-engine-prune` systemd service+timer that daily find-deletes `events-*.jsonl` older than RETENTION_DAYS (default 90). nginx logs keep their existing size-based logrotate.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-retention-timer, traffic_probe_runbook(12).md#3.5
- **relations:** part of the privacy posture
- **verify-later:** setup.sh site-engine-prune.timer; RETENTION_DAYS

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### claimed-item-timeout evidence-gated completion + reset (Lever A/C) — avoided building a duplicate watchdog
- **category:** scheduler-and-tasks
- **status-signal:** deployed
- **status-evidence:** running_notes_14(20) Part 12: "'Auto-completed...' is set by the **`claimed-item-timeout` scheduled task's SQL `pre_query`**, NOT a Go reaper... `migration_claimed_item_timeout_evidence_check.sql` ... is essentially Option A + Lever C, already authored... the FOCUS_dispatch `reset_stale_claims` watchdog is redundant; do NOT build it." Part 12 addenda confirm the v2 migration (page_components-based evidence, not the untrustworthy `build_status='deployed'` flag) applied and verified live, plus the companion `pageHasComponents` deploy-guard (Option B) delivered.
- **what:** A `claimed-item-timeout` scheduled task's `pre_query` already implements both (a) evidence-gated auto-completion of stuck claims (only complete if the specific artefact shows positive evidence) and (b) a stale-claim reset-to-`triaged`/`failed` after a timeout with attempt counting. Mid-investigation, an agent nearly built a brand-new "reset stale claims" watchdog before discovering this — a documented reuse-over-build catch. The evidence signal itself evolved further: from trusting `pages.deployed_at`/`build_status='deployed'` (provably lying, per the homepage case) to checking `page_components.updated_at > claimed_at` directly.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Parts 12–12 addenda
- **relations:** A4 homepage root cause (above); sectionless-page durability stack (below); dispatch throughput (Family J)
- **verify-later:** current `claimed-item-timeout` pre_query SQL in `scheduled_tasks`.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Dispatch throughput bottleneck (Family J) — one-site-per-tick, NOT-EXISTS-blocked
- **category:** scheduler-and-tasks
- **status-signal:** unknown
- **status-evidence:** CATALOGUE_gamesdesign_post_sync_fix_defects(4).md Family J: "the dispatcher is **one-site-per-tick** (selection `LIMIT 1`... processes ~5 items then exits) and **NOT-EXISTS-blocked** (a `NOT EXISTS` clause excludes a site *entirely* while any of its items is `status='claimed'`... line 276)... Standard manual unstick for now... **To investigate in the separate thread.**"
- **what:** Multi-tool/multi-game adoption sites drain over hours, appearing stalled, because the build-dispatch mechanism processes one site per scheduler tick and blocks an entire site's queue while any single item on it is claimed (no bounded concurrency, no per-item exclusion). A dead handler leaving a stale claim freezes the whole site until a reaper resets it. Explicitly spun out as a separate, not-yet-investigated thread rather than fixed within this arc; running_notes_17(16) later notes it's still an open TODO ("SPEED UP the rebuild pipeline... Not yet investigated").
- **sources:** adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md Family J; running_notes_14(20) Part 9; running_notes_17(16) "Missing-game... + speed TODO"
- **relations:** claimed-item-timeout evidence-check reliability mini-project; A1 tool/game deploy gap
- **verify-later:** `build-pipeline-trigger` dispatcher current selection logic (`LIMIT 1`, NOT EXISTS clause, line ~276 at time of writing).
