# Register — storage-architecture

9 concepts, consolidated from 22 raw extractions (11 unique blocks, each duplicated
once in the source cluster file) across units U01_docs024_numbered_core,
U06_finetuning, U18_sql_for_agents, U24b_docs_archive_finetuning,
U24c_docs_archive_traffic_probe, U26_misc_dirs. Two concepts additionally absorbed
raw blocks the source cluster file had tagged under other categories because they
describe the same underlying mechanism from a different project's angle: STG-007
absorbed vm-backend-sites' "Engine store v2" (site-engine's own JSONL store
evolution), and STG-008 absorbed new:backend-service-deployment's "B2 dead-drop
persistence" (idea.uk's version of the same tiered box→B2→chassis pattern) — see
register/vm-backend-sites.md for the surrounding project narratives.

### STG-001 — Storage: per-call S3 client construction is canonical (params.StorageClient deprecated)
- **status:** deployed
- **status-evidence:** 032 TL;DR + deprecation rationale (nil-at-startup pods).
- **what:** All storage-touching actions construct their own client via `storage.NewS3Client` with env-var names in `ObjectStorageConfig` (B2_APPLICATION_KEY_ID/KEY from `personae-platform-secrets`); the injected `params.StorageClient` is unreliable (nil when IMAGE_BUCKET is absent at startup). Spawn-time env forwarding (Path C) is gated by `isStorageEnabledAgent`/orchestrator/code-driven — the gate must be kept maintained; storage workers must be spawn-and-called, not direct-triggered.
- **sources:** 032 (full)
- **relations:** spawn env propagation; thunder presigned URLs; Storage credential architecture decision (STG-004)
- **verify-later:** isStorageEnabledAgent list; remaining Path-B users

### STG-002 — Hostile-VM threat model for the training data plane
- **status:** deployed
- **status-evidence:** PLAN(7): "Threat model: assume the Thunder VM is hostile"; Phase-4 FOCUS §14: "the VM holds no B2 credentials, just a time-limited URL."
- **what:** The GPU box is treated as untrusted: it holds no B2 key, no DB access, no inbound endpoint — only single-object presigned URLs (write-only PUTs, plus one GET on resume). Rejected alternatives were a standing scoped B2 key on the box (prefix-wide bearer leak risk) and a per-save callback endpoint (attack surface + a mintable token on the box). A compromised box can at most overwrite its own checkpoint objects within expiry, bounded by versioning; artefact *integrity* is explicitly the eval gate's job, not the URL's. The adapter is the sole credential boundary and mints all URLs.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#chosen-approach,#net-security-position; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow
- **relations:** Presigned-URL data plane (STG-003); Checkpoint & final-adapter upload to B2 (STG-006); Storage credential architecture decision (STG-004)
- **verify-later:** adapter presign code paths; B2 bucket versioning setting

### STG-003 — Presigned-URL data plane / storage boundary (adapter mints URLs; bytes never transit Kafka)
- **status:** deployed
- **status-evidence:** FOCUS(25) Phase-4 section: "PHASE 4 STATUS: COMPLETE & DEPLOYED (verified in production 2026-05-24)"; bucket/key convention "VERIFIED end-to-end 2026-05-23"; NOTES(39) §5a confirms "Adapter prepare_object_url is live on thunder-adapter:v1.0.1049… signing against personae-model-training" (2026-06-02).
- **what:** The adapter presigns; it never moves data. Only URLs (hundreds of bytes) travel over Kafka; dataset/artefact bytes go directly VM↔B2 over HTTPS. `prepare_dataset_url`/`prepare_artefact_url` presign by ID; the general `prepare_object_url` presigns any key (GET default 60m, PUT 24h). Canonical layout: bucket `personae-model-training`, keys `finetuning/datasets/{export_id}/training.jsonl`, `finetuning/scripts/bundle.tar.gz`, `finetuning/checkpoints/{run_id}/ckpt-N.tar.gz`, `finetuning/artefacts/{run_id}/adapter.tar.gz` (`finetuning/` is a folder prefix, not a bucket; the preparer agent-def's stale `s3_bucket=finetuning` value cost a 403 debugging cycle). The presign primitive evolved: DatasetURL/ArtefactURL → generic `prepare_object_url` (delegating to one `ObjectURL` signing path) → batch `prepare_object_urls` → `prepare_resume_url`. Verification gotchas: presigned GETs 403 on HEAD (`curl -I`) because SigV4 signs the method; kcat escapes `&`; use the b2 CLI, not aws (and not the snap b2, which is a BBC Micro emulator).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow; working/phase5/NOTES_phase5_training_launcher_running(45).md#5a,#update-2026-06-05; working/phase5/UPLOAD_bundle.sh; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#Phase-4-data-flow-actions; phase5/NOTES_phase5_training_launcher_running(39).md#5a
- **relations:** Hostile-VM threat model (STG-002); Storage credential architecture decision (STG-004); Checkpoint & final-adapter upload to B2 (STG-006)
- **verify-later:** data_url_actions.go (ObjectURL/handlePrepareObjectURL); TRAINING_BUCKET env; personae-storage-secrets wiring; storage.Client.GetPresignedPutURL

### STG-004 — Storage credential architecture decision (no storage-adapter service; presign not blob-pipe)
- **status:** partial
- **status-evidence:** FOCUS(25)/(21) §14, both dated 2026-05-22: "Decision: hardcode the adapter's storage env for now; adopt centralised credential sourcing later; do NOT build a storage-adapter service… Deferred to a dedicated platform pass; not built yet."
- **what:** A storage-adapter (a service owning creds that others message) was rejected because it would route multi-MB dataset/artefact bytes through Kafka (`max.message.bytes` ~1MB; raised limits wreck brokers) — the presign pattern moves only URLs, so a storage-adapter is only safe for minting URLs, dangerous for moving data. The acknowledged mess as of the decision: the same B2 creds are sourced four different ways across services (webscrape B2_* env; image-generator AWS_* + configmap; preparer spawn-injection; thunder hardcoded env). Eventual fix (deferred, not built): one shared constructor (`storage.NewDefaultClient`) reading `personae-storage-secrets` uniformly, so untrusted agents get presigned URLs, never a blob pipe. Related blast-radius lesson: adding `GetPresignedPutURL` to the shared `storage.Client` interface forced rebuilding every binary importing `platform/storage`.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#storage-credential-architecture; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14
- **relations:** Presigned-URL data plane (STG-003); Storage: per-call S3 client construction (STG-001); adapters category
- **verify-later:** whether storage.NewDefaultClient exists; secret sourcing per service; platform/storage/interface.go

### STG-005 — asset-deployer (S3 → optimize-by-purpose → git)
- **status:** deployed
- **status-evidence:** 044 definition; idle timeout in 075; called from image-build-handler flows.
- **what:** Single-purpose specialist wrapping `deploy_image_asset`: downloads an asset from S3, optimizes it according to purpose (logo vs hero), commits to git. Reusable for any image deploy task.
- **sources:** 044_asset_deployer.sql; 057_image_build_handler.sql
- **relations:** image-build-handler; undeployed_assets discovery check
- **verify-later:** deploy_image_asset action; optimization rules per purpose

### STG-006 — Checkpoint & final-adapter upload to B2 (upload manifest, save-index keying, resume)
- **status:** partial
- **status-evidence:** PLAN(4): "Phases A, B, C BUILT and audited 2026-06-05. Phase A Tier-1 … PASSED … Phase D adapter side BUILT … its launcher wiring … is the only code left."
- **what:** The design solving three coupled gaps (final adapter not durable, monitor decommissions on completion, no checkpoints). The launcher pre-mints presigned single-object write-only PUT URLs into a `/workspace/upload_manifest.json`; `02_train`'s `CheckpointUploader` callback tars+PUTs each save keyed by save-INDEX (not the Trainer's global_step); the final adapter upload is a hard gate (raises → non-zero exit → no RUN_SH_DONE). Resume reuses `storage.Client.ListObjects` (not a new method) to pick the highest `ckpt-<N>` and presign a GET.
- **sources:** docubundle/.../PLAN_checkpoint_and_artefact_upload_b2(4).md; phase5/PLAN_checkpoint_and_artefact_upload_b2(6).md; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-05-upload-path
- **relations:** Hostile-VM threat model (STG-002); Presigned-URL data plane (STG-003); makes the monitor's DONE_OK→decommission safe
- **verify-later:** 02_train_llama_3_3_70b.py CheckpointUploader; adapter prepare_resume_url; migration 109; storage.Client.ListObjects

### STG-007 — JSON store scaling evolution (whole-file → dirty-flusher → daily JSONL)
- **status:** deployed
- **status-evidence:** running_notes 2026-06-11: "Store scaling fix (structural, pre-launch)" then "Store v2 (JSONL) … Burst-tested: 300 events + 100 visits"; prune timer installed at go-live (2026-06-12).
- **what:** site-engine's on-box storage went through three shapes as traffic volume was reasoned about ahead of launch. v0 rewrote the entire ever-growing JSON file on every beacon hit (linear write cliff) and held all events in RAM. v1 added a dirty-flag + 5s background flusher (`AddVisit` no longer persists per call; `AddEvent` still immediate) — still a single ever-growing file. v2 replaced the monolithic file entirely: events append to daily `events-YYYYMMDD.jsonl` (one line per submission, O(1) at any volume, rotation = the date, bounded RAM); `/stats` counters live in a small `counters.json` flushed by the same dirty-flag debounced flusher (crash window ≤5s of visit counts, never events); SIGTERM/SIGINT flush+fsync. The in-RAM `Store.Events` map and the uncalled `Store.Snapshot()` were removed. Retention is enforced on-box by `site-engine-prune.timer` (daily delete of events files older than RETENTION_DAYS, default 90); explicitly NO logrotate on engine files (move/truncate would race the open handle) — nginx logs keep their own logrotate.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-live-b2-action,#2026-06-11-store-v2; deploy_setup/working_dir/main.go#header; deploy_setup/site-engine/store.go (header); relojistas_notes(8).md#decisions
- **relations:** site-engine (register/vm-backend-sites.md, VMB-002); /events export; minimal-data privacy (90-day prune)
- **verify-later:** store.go Flush/flushLoop/EventCounts/openEventsFileLocked; /var/lib/site-engine/{events-*.jsonl,counters.json}; systemctl list-timers site-engine-prune.timer

### STG-008 — Persistence design: tiered one-way data flow for exposed services (box → B2 → chassis)
- **status:** partial
- **status-evidence:** `running_notes(44).md` "Persistence decisions LOCKED"; "Phased: Phase 1 (now, box): keep local store + add B2 record-writing... Phase 2 (when ready, framework): create table + idea-ingest scheduled task" — design settled, but as of this documentation the live idea.uk service still ran on `orders.json` only, with no idea-ingest agent or `idea_orders` table evidenced. Independently, traffic_probe's relojistas_notes record the same principle as a decision ("No third 'collector' VM") with its own pulling collector still disabled.
- **what:** A security-motivated pattern for any internet-facing satellite service to get its data into the core chassis DB without opening an inbound path, arrived at independently by two projects. idea.uk's version (the first case, PERSISTENCE_design.md): (1) local operational store on the exposed box (kept as JSON, explicitly rejecting SQLite to preserve the stdlib-only/`GOPROXY=off` build); (2) a one-way B2 "dead-drop" channel (box writes immutable per-event records via a write-only-scoped/presigned URL — reusing the same pattern Thunder adapter already uses for artefact transfer); (3) a `scheduled_tasks`-driven `idea-ingest` agent on the chassis side that *pulls* new B2 records and upserts into a restricted-role schema (`business_intel`/`ecommerce`) — "chassis PULLS; box never connects in." Kafka topic / narrow HTTPS ingest / direct PG were all rejected as inbound paths. Table design (`ecommerce.orders`, `ecommerce.taster_events`, `clients_db.idea_reports`) deliberately keeps no card data (Stripe opaque refs only). Explicit worst-case analysis: a compromised box can write junk into one B2 prefix, no more. traffic_probe's own variant of the same principle (serving box only buffers daily JSONL; the CLUSTER pulls over key-gated HTTPS on a schedule into `clients_db`, B2 optional cold backup, no adapter/SSH needed for collection) is tracked separately as its own concept — see register/vm-backend-sites.md "Pull-not-push off-cluster data return" (VMB-009) — since it is a genuinely different mechanism (HTTPS-pull vs B2-dead-drop-then-poll) solving the same design goal.
- **sources:** idea.uk/nginx/PERSISTENCE_design(1).md; idea.uk/running_notes(63).md; running_notes(44).md (PERSISTENCE_design.md summary, two checkpoints 2026-06-04)
- **relations:** service-deployer / vmhost adapter (register/adapters.md, ADP-016); Thunder adapter (B2 presigned-URL precedent, STG-002/STG-003); Pull-not-push off-cluster data return (register/vm-backend-sites.md, VMB-009); scheduler-and-tasks
- **verify-later:** whether business_intel.idea_orders / ecommerce.orders / an idea-ingest scheduled task exist

### STG-009 — Result storage split (DB paper-trail + S3 artefacts)
- **status:** deployed
- **status-evidence:** basic_usage/002 states it as fact: `final_result` column in `orchestrator_state`, a `website_projects` table per client schema with preview/live URLs, and site-publisher's `s3_upload` of files.
- **what:** The record of a build lives in PostgreSQL (full workflow history + consolidated `final_result` JSON + `website_projects` metadata with URLs) while the tangible outputs (HTML/CSS/JS files, generated images/logos) live in S3-compatible object storage, referenced by URI from workflow results — "the database holds the record of what happened... the object storage holds the actual product."
- **sources:** docs/basic_usage/002storage_of_results; docs/architecture/027-create-website-creation-system
- **relations:** website-builder group; storage-architecture spine (032, S3/B2)
- **verify-later:** website_projects table; s3_upload action; current B2 storage docs
