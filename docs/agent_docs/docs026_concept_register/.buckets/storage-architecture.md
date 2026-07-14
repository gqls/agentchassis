
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Storage: per-call S3 client construction is canonical (params.StorageClient deprecated)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** 032 TL;DR + deprecation rationale (nil-at-startup pods)
- **what:** All storage-touching actions construct their own client via storage.NewS3Client with env-var names in ObjectStorageConfig (B2_APPLICATION_KEY_ID/KEY from personae-platform-secrets); injected params.StorageClient is unreliable (nil when IMAGE_BUCKET absent at startup). Spawn-time env forwarding (Path C) is gated by isStorageEnabledAgent/orchestrator/code-driven — keep the gate maintained; storage workers must be spawn-and-called, not direct-triggered.
- **sources:** 032 full
- **relations:** spawn env propagation; thunder presigned URLs
- **verify-later:** isStorageEnabledAgent list; remaining Path-B users

<!-- SOURCE: U06_finetuning.md -->
### Hostile-VM threat model for the training data plane
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** PLAN(7): "Threat model: assume the Thunder VM is hostile"; Phase-4 FOCUS §14 "the VM holds no B2 credentials, just a time-limited URL".
- **what:** The GPU box is treated as untrusted: it holds no B2 key, no DB access, no inbound endpoint — only single-object presigned URLs (write-only PUTs, plus one GET on resume). Rejected alternatives: standing scoped B2 key on the box (prefix-wide bearer leak risk) and per-save callback endpoint (attack surface + a mintable token on the box). A compromised box can at most overwrite its own checkpoint objects within expiry, bounded by versioning; artefact *integrity* is explicitly the eval gate's job, not the URL's. The adapter is the sole credential boundary and mints all URLs.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#chosen-approach,#net-security-position; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow
- **relations:** presigned data plane; eval gate; storage credential architecture decision
- **verify-later:** adapter presign code paths; B2 bucket versioning setting

<!-- SOURCE: U06_finetuning.md -->
### Presigned-URL data plane (adapter mints URLs; bytes never transit Kafka)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) Phase-4 section: "PHASE 4 STATUS: COMPLETE & DEPLOYED (verified in production 2026-05-24)"; bucket/key convention "VERIFIED end-to-end 2026-05-23".
- **what:** The adapter presigns; it never moves data. Only URLs (hundreds of bytes) travel over Kafka; dataset/artefact bytes go directly VM↔B2 over HTTPS. Canonical layout: bucket `personae-model-training`, keys `finetuning/datasets/{export_id}/training.jsonl`, `finetuning/scripts/bundle.tar.gz`, `finetuning/checkpoints/{run_id}/ckpt-N.tar.gz`, `finetuning/artefacts/{run_id}/adapter.tar.gz` (note: `finetuning/` is a folder prefix, not a bucket; the preparer agent-def's `s3_bucket=finetuning` is stale/logical and cost a 403 debugging cycle). The presign primitive evolved: DatasetURL/ArtefactURL → generic `prepare_object_url` (they now delegate to ObjectURL — one signing path) → batch `prepare_object_urls` → `prepare_resume_url`. Verification gotchas: presigned GETs 403 on HEAD (`curl -I`) because SigV4 signs the method; kcat escapes `&` as &; use the b2 CLI, not aws (and not the snap b2, which is a BBC Micro emulator).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow; working/phase5/NOTES_phase5_training_launcher_running(45).md#5a,#update-2026-06-05; working/phase5/UPLOAD_bundle.sh
- **relations:** hostile-VM threat model; storage credential decision; batch presign; docubundle B2 notes
- **verify-later:** data_url_actions.go; TRAINING_BUCKET env; personae-storage-secrets wiring

<!-- SOURCE: U06_finetuning.md -->
### Storage credential architecture decision (no storage-adapter service)
- **category:** storage-architecture
- **status-signal:** aspirational
- **status-evidence:** FOCUS(25) §14 "Decision (2026-05-22): hardcode the adapter's storage env for now; adopt centralised credential sourcing later; do NOT build a storage-adapter service… Deferred to a dedicated platform pass; not built yet."
- **what:** A storage-adapter (service owning creds that others message) was rejected because it would route multi-MB dataset/artefact bytes through Kafka (max.message.bytes ~1MB; raised limits wreck brokers) — the presign pattern moves only URLs. The acknowledged mess: the same B2 creds are sourced four different ways across services (webscrape B2_* env; image-generator AWS_* + configmap; preparer spawn-injection; thunder hardcoded env). Eventual fix: one shared constructor (`storage.NewDefaultClient`) reading `personae-storage-secrets` uniformly. Related blast-radius lesson: adding `GetPresignedPutURL` to the shared storage.Client interface forced rebuilding every binary importing platform/storage.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#storage-credential-architecture
- **relations:** presigned data plane; adapters category; debugging guide item 18
- **verify-later:** whether NewDefaultClient exists; secret sourcing per service

<!-- SOURCE: U18_sql_for_agents.md -->
### asset-deployer (S3 → optimize-by-purpose → git)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** 044 definition; idle timeout in 075; called from image-build-handler flows.
- **what:** Single-purpose specialist wrapping deploy_image_asset: downloads an asset from S3, optimizes it according to purpose (logo vs hero), commits to git. Reusable for any image deploy task.
- **sources:** 044_asset_deployer.sql; 057_image_build_handler.sql
- **relations:** image-build-handler, undeployed_assets discovery check
- **verify-later:** deploy_image_asset action; optimization rules per purpose

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Presigned-URL storage boundary (prepare_object_url / dataset / artefact)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** NOTES(39) §5a "Adapter prepare_object_url is live on thunder-adapter:v1.0.1049 … signing against personae-model-training" (2026-06-02)
- **what:** The adapter mints presigned B2 URLs but never moves data — only the few-hundred-byte URL travels over Kafka, the actual bytes go directly VM↔B2 over HTTPS. `prepare_dataset_url`/`prepare_artefact_url` presign by ID; the general `prepare_object_url` presigns any key (GET default 60m, PUT 24h; DatasetURL/ArtefactURL delegate to it). Bucket is `personae-model-training` (the preparer's `s3_bucket=finetuning` agent_def value is stale/logical); B2 keys live in `personae-storage-secrets`.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#Phase-4-data-flow-actions; phase5/NOTES_phase5_training_launcher_running(39).md#5a
- **relations:** enables the checkpoint/artefact upload; contrasted with rejected storage-adapter; verification gotcha: presigned GET fails HEAD (curl -I) with 403
- **verify-later:** internal/adapters/thunder/data_url_actions.go (ObjectURL/handlePrepareObjectURL); storage.Client.GetPresignedPutURL

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Storage-credential architecture decision (no storage-adapter; presign not blob-pipe)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §14 "Storage credential architecture — decision (2026-05-22) … do NOT build a storage-adapter service"
- **what:** A decision that routing multi-MB datasets / hundreds-of-MB artefacts through a storage-adapter over Kafka is a real failure (max.message.bytes ~1MB), so a storage-adapter is only safe for minting URLs, dangerous for moving data. Interim: hardcode the adapter's B2 env. Eventual (deferred, not built): one shared `storage.NewDefaultClient` reading `personae-storage-secrets` everywhere to kill the four-conventions credential mess; untrusted agents get presigned URLs, never a blob pipe.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Storage credential architecture)
- **relations:** justifies the presigned-URL boundary; GetPresignedPutURL added to storage.Client forced rebuild of every binary importing platform/storage
- **verify-later:** platform/storage/interface.go; personae-storage-secrets

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Checkpoint & final-adapter upload to B2 (upload manifest, save-index keying, resume)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** PLAN(4) "Phases A, B, C BUILT and audited 2026-06-05. Phase A Tier-1 … PASSED … Phase D adapter side BUILT … its launcher wiring … is the only code left"
- **what:** The design solving three coupled gaps (final adapter not durable, monitor decommissions on completion, no checkpoints). The launcher pre-mints presigned single-object write-only PUT URLs into a `/workspace/upload_manifest.json`; `02_train`'s `CheckpointUploader` callback tars+PUTs each save keyed by save-INDEX (not the Trainer's global_step); the final adapter upload is a hard gate (raises → non-zero exit → no RUN_SH_DONE). Threat model assumes the VM is hostile (holds no B2 key, only single-object URLs); a standing scoped key and per-save callback endpoint were both rejected. Resume reuses `storage.Client.ListObjects` (not a new method) to pick the highest `ckpt-<N>` and presign a GET.
- **sources:** docubundle/.../PLAN_checkpoint_and_artefact_upload_b2(4).md; phase5/PLAN_checkpoint_and_artefact_upload_b2(6).md; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-05-upload-path
- **relations:** makes the monitor's DONE_OK→decommission safe; still-LEFT Phase D3 dispatch_thunder_prepare_resume_url + migration for check_resume; corrects the "list-keys is the ONE adapter gap" claim
- **verify-later:** 02_train_llama_3_3_70b.py CheckpointUploader; adapter prepare_resume_url; migration 109; storage.Client.ListObjects

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### JSON store scaling evolution (whole-file → dirty-flusher → daily JSONL)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "Store scaling fix (structural, pre-launch)" then "Store v2 (JSONL) … v2: events append to daily JSONL … O(1) at any volume"; burst-tested.
- **what:** v0 rewrote the entire ever-growing JSON file on every beacon hit (linear write cliff). v1 added a dirty-flag + 5s background flusher (AddVisit no longer persists per call; AddEvent still immediate). v2 replaced the monolithic file: events append to daily `events-YYYYMMDD.jsonl` (one line per submission, bounded RAM), /stats counters live in a small `counters.json`; SIGTERM fsyncs. Removed the in-RAM `Store.Events` map and uncalled `Store.Snapshot()`.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-live-b2-action, traffic_probe_running_notes(27).md#2026-06-11-store-v2, deploy_setup/working_dir/main.go#header
- **relations:** abandoned Store.Events/Snapshot; drove the ENGINE_DATA_DIR rename
- **verify-later:** store.go Flush/flushLoop/EventCounts/openEventsFileLocked; /var/lib/site-engine/{events-*.jsonl,counters.json}

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Persistence design — tiered one-way data flow for exposed services (box → B2 → chassis)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** `running_notes(44).md` "Persistence decisions LOCKED"; "Phased: Phase 1 (now, box): keep local store + add B2 record-writing... Phase 2 (when ready, framework): create table + idea-ingest scheduled task."
- **what:** A security-motivated pattern for any internet-facing satellite service (idea.uk being the first case) to get its data into the core chassis DB without opening an inbound path: (1) local operational store on the exposed box (kept as JSON, explicitly rejecting SQLite to preserve the stdlib-only/`GOPROXY=off` build); (2) a one-way B2 "dead-drop" channel (box writes immutable per-event records via a write-only-scoped/presigned URL — reuses the same pattern Thunder adapter already uses for artefact transfer); (3) a `scheduled_tasks`-driven ingest agent on the chassis side that *pulls* new B2 records and upserts into a restricted-role schema (`business_intel`/`ecommerce`), "chassis PULLS; box never connects in." Explicit worst-case analysis: a compromised box can write junk into one B2 prefix, no more. Table design (`ecommerce.orders`, `ecommerce.taster_events`, `clients_db.idea_reports`) deliberately keeps no card data (Stripe opaque refs only).
- **sources:** `running_notes(44).md` (`PERSISTENCE_design.md` summary, two checkpoints on 2026-06-04)
- **relations:** service-deployer pattern; Thunder adapter (B2 presigned-URL precedent); storage-architecture (032, S3/B2)
- **verify-later:** whether `business_intel.idea_orders` / `ecommerce.orders` / an `idea-ingest` scheduled task exist

<!-- SOURCE: U26_misc_dirs.md -->
### Result storage split (DB paper-trail + S3 artefacts)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** basic_usage/002 states it as fact: final_result column in orchestrator_state, a website_projects table per client schema with preview/live URLs, and site-publisher's s3_upload of files.
- **what:** The record of a build lives in PostgreSQL (full workflow history + consolidated final_result JSON + website_projects metadata with URLs) while the tangible outputs (HTML/CSS/JS files, generated images/logos) live in S3-compatible object storage, referenced by URI from workflow results — "the database holds the record of what happened... the object storage holds the actual product".
- **sources:** docs/basic_usage/002storage_of_results; docs/architecture/027-create-website-creation-system (site-publisher s3_upload)
- **relations:** website-builder group; storage-architecture spine (032, S3/B2)
- **verify-later:** website_projects table; s3_upload action; current B2 storage docs

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Storage: per-call S3 client construction is canonical (params.StorageClient deprecated)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** 032 TL;DR + deprecation rationale (nil-at-startup pods)
- **what:** All storage-touching actions construct their own client via storage.NewS3Client with env-var names in ObjectStorageConfig (B2_APPLICATION_KEY_ID/KEY from personae-platform-secrets); injected params.StorageClient is unreliable (nil when IMAGE_BUCKET absent at startup). Spawn-time env forwarding (Path C) is gated by isStorageEnabledAgent/orchestrator/code-driven — keep the gate maintained; storage workers must be spawn-and-called, not direct-triggered.
- **sources:** 032 full
- **relations:** spawn env propagation; thunder presigned URLs
- **verify-later:** isStorageEnabledAgent list; remaining Path-B users

<!-- SOURCE: U06_finetuning.md -->
### Hostile-VM threat model for the training data plane
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** PLAN(7): "Threat model: assume the Thunder VM is hostile"; Phase-4 FOCUS §14 "the VM holds no B2 credentials, just a time-limited URL".
- **what:** The GPU box is treated as untrusted: it holds no B2 key, no DB access, no inbound endpoint — only single-object presigned URLs (write-only PUTs, plus one GET on resume). Rejected alternatives: standing scoped B2 key on the box (prefix-wide bearer leak risk) and per-save callback endpoint (attack surface + a mintable token on the box). A compromised box can at most overwrite its own checkpoint objects within expiry, bounded by versioning; artefact *integrity* is explicitly the eval gate's job, not the URL's. The adapter is the sole credential boundary and mints all URLs.
- **sources:** working/phase5/PLAN_checkpoint_and_artefact_upload_b2(7).md#chosen-approach,#net-security-position; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow
- **relations:** presigned data plane; eval gate; storage credential architecture decision
- **verify-later:** adapter presign code paths; B2 bucket versioning setting

<!-- SOURCE: U06_finetuning.md -->
### Presigned-URL data plane (adapter mints URLs; bytes never transit Kafka)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** FOCUS(25) Phase-4 section: "PHASE 4 STATUS: COMPLETE & DEPLOYED (verified in production 2026-05-24)"; bucket/key convention "VERIFIED end-to-end 2026-05-23".
- **what:** The adapter presigns; it never moves data. Only URLs (hundreds of bytes) travel over Kafka; dataset/artefact bytes go directly VM↔B2 over HTTPS. Canonical layout: bucket `personae-model-training`, keys `finetuning/datasets/{export_id}/training.jsonl`, `finetuning/scripts/bundle.tar.gz`, `finetuning/checkpoints/{run_id}/ckpt-N.tar.gz`, `finetuning/artefacts/{run_id}/adapter.tar.gz` (note: `finetuning/` is a folder prefix, not a bucket; the preparer agent-def's `s3_bucket=finetuning` is stale/logical and cost a 403 debugging cycle). The presign primitive evolved: DatasetURL/ArtefactURL → generic `prepare_object_url` (they now delegate to ObjectURL — one signing path) → batch `prepare_object_urls` → `prepare_resume_url`. Verification gotchas: presigned GETs 403 on HEAD (`curl -I`) because SigV4 signs the method; kcat escapes `&` as &; use the b2 CLI, not aws (and not the snap b2, which is a BBC Micro emulator).
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#phase-4-data-flow; working/phase5/NOTES_phase5_training_launcher_running(45).md#5a,#update-2026-06-05; working/phase5/UPLOAD_bundle.sh
- **relations:** hostile-VM threat model; storage credential decision; batch presign; docubundle B2 notes
- **verify-later:** data_url_actions.go; TRAINING_BUCKET env; personae-storage-secrets wiring

<!-- SOURCE: U06_finetuning.md -->
### Storage credential architecture decision (no storage-adapter service)
- **category:** storage-architecture
- **status-signal:** aspirational
- **status-evidence:** FOCUS(25) §14 "Decision (2026-05-22): hardcode the adapter's storage env for now; adopt centralised credential sourcing later; do NOT build a storage-adapter service… Deferred to a dedicated platform pass; not built yet."
- **what:** A storage-adapter (service owning creds that others message) was rejected because it would route multi-MB dataset/artefact bytes through Kafka (max.message.bytes ~1MB; raised limits wreck brokers) — the presign pattern moves only URLs. The acknowledged mess: the same B2 creds are sourced four different ways across services (webscrape B2_* env; image-generator AWS_* + configmap; preparer spawn-injection; thunder hardcoded env). Eventual fix: one shared constructor (`storage.NewDefaultClient`) reading `personae-storage-secrets` uniformly. Related blast-radius lesson: adding `GetPresignedPutURL` to the shared storage.Client interface forced rebuilding every binary importing platform/storage.
- **sources:** working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#storage-credential-architecture
- **relations:** presigned data plane; adapters category; debugging guide item 18
- **verify-later:** whether NewDefaultClient exists; secret sourcing per service

<!-- SOURCE: U18_sql_for_agents.md -->
### asset-deployer (S3 → optimize-by-purpose → git)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** 044 definition; idle timeout in 075; called from image-build-handler flows.
- **what:** Single-purpose specialist wrapping deploy_image_asset: downloads an asset from S3, optimizes it according to purpose (logo vs hero), commits to git. Reusable for any image deploy task.
- **sources:** 044_asset_deployer.sql; 057_image_build_handler.sql
- **relations:** image-build-handler, undeployed_assets discovery check
- **verify-later:** deploy_image_asset action; optimization rules per purpose

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Presigned-URL storage boundary (prepare_object_url / dataset / artefact)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** NOTES(39) §5a "Adapter prepare_object_url is live on thunder-adapter:v1.0.1049 … signing against personae-model-training" (2026-06-02)
- **what:** The adapter mints presigned B2 URLs but never moves data — only the few-hundred-byte URL travels over Kafka, the actual bytes go directly VM↔B2 over HTTPS. `prepare_dataset_url`/`prepare_artefact_url` presign by ID; the general `prepare_object_url` presigns any key (GET default 60m, PUT 24h; DatasetURL/ArtefactURL delegate to it). Bucket is `personae-model-training` (the preparer's `s3_bucket=finetuning` agent_def value is stale/logical); B2 keys live in `personae-storage-secrets`.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#Phase-4-data-flow-actions; phase5/NOTES_phase5_training_launcher_running(39).md#5a
- **relations:** enables the checkpoint/artefact upload; contrasted with rejected storage-adapter; verification gotcha: presigned GET fails HEAD (curl -I) with 403
- **verify-later:** internal/adapters/thunder/data_url_actions.go (ObjectURL/handlePrepareObjectURL); storage.Client.GetPresignedPutURL

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Storage-credential architecture decision (no storage-adapter; presign not blob-pipe)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** FOCUS(21) §14 "Storage credential architecture — decision (2026-05-22) … do NOT build a storage-adapter service"
- **what:** A decision that routing multi-MB datasets / hundreds-of-MB artefacts through a storage-adapter over Kafka is a real failure (max.message.bytes ~1MB), so a storage-adapter is only safe for minting URLs, dangerous for moving data. Interim: hardcode the adapter's B2 env. Eventual (deferred, not built): one shared `storage.NewDefaultClient` reading `personae-storage-secrets` everywhere to kill the four-conventions credential mess; untrusted agents get presigned URLs, never a blob pipe.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#14 (Storage credential architecture)
- **relations:** justifies the presigned-URL boundary; GetPresignedPutURL added to storage.Client forced rebuild of every binary importing platform/storage
- **verify-later:** platform/storage/interface.go; personae-storage-secrets

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Checkpoint & final-adapter upload to B2 (upload manifest, save-index keying, resume)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** PLAN(4) "Phases A, B, C BUILT and audited 2026-06-05. Phase A Tier-1 … PASSED … Phase D adapter side BUILT … its launcher wiring … is the only code left"
- **what:** The design solving three coupled gaps (final adapter not durable, monitor decommissions on completion, no checkpoints). The launcher pre-mints presigned single-object write-only PUT URLs into a `/workspace/upload_manifest.json`; `02_train`'s `CheckpointUploader` callback tars+PUTs each save keyed by save-INDEX (not the Trainer's global_step); the final adapter upload is a hard gate (raises → non-zero exit → no RUN_SH_DONE). Threat model assumes the VM is hostile (holds no B2 key, only single-object URLs); a standing scoped key and per-save callback endpoint were both rejected. Resume reuses `storage.Client.ListObjects` (not a new method) to pick the highest `ckpt-<N>` and presign a GET.
- **sources:** docubundle/.../PLAN_checkpoint_and_artefact_upload_b2(4).md; phase5/PLAN_checkpoint_and_artefact_upload_b2(6).md; phase5/NOTES_phase5_training_launcher_running(39).md#update-2026-06-05-upload-path
- **relations:** makes the monitor's DONE_OK→decommission safe; still-LEFT Phase D3 dispatch_thunder_prepare_resume_url + migration for check_resume; corrects the "list-keys is the ONE adapter gap" claim
- **verify-later:** 02_train_llama_3_3_70b.py CheckpointUploader; adapter prepare_resume_url; migration 109; storage.Client.ListObjects

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### JSON store scaling evolution (whole-file → dirty-flusher → daily JSONL)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "Store scaling fix (structural, pre-launch)" then "Store v2 (JSONL) … v2: events append to daily JSONL … O(1) at any volume"; burst-tested.
- **what:** v0 rewrote the entire ever-growing JSON file on every beacon hit (linear write cliff). v1 added a dirty-flag + 5s background flusher (AddVisit no longer persists per call; AddEvent still immediate). v2 replaced the monolithic file: events append to daily `events-YYYYMMDD.jsonl` (one line per submission, bounded RAM), /stats counters live in a small `counters.json`; SIGTERM fsyncs. Removed the in-RAM `Store.Events` map and uncalled `Store.Snapshot()`.
- **sources:** traffic_probe_running_notes(27).md#2026-06-11-live-b2-action, traffic_probe_running_notes(27).md#2026-06-11-store-v2, deploy_setup/working_dir/main.go#header
- **relations:** abandoned Store.Events/Snapshot; drove the ENGINE_DATA_DIR rename
- **verify-later:** store.go Flush/flushLoop/EventCounts/openEventsFileLocked; /var/lib/site-engine/{events-*.jsonl,counters.json}

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Persistence design — tiered one-way data flow for exposed services (box → B2 → chassis)
- **category:** storage-architecture
- **status-signal:** partial
- **status-evidence:** `running_notes(44).md` "Persistence decisions LOCKED"; "Phased: Phase 1 (now, box): keep local store + add B2 record-writing... Phase 2 (when ready, framework): create table + idea-ingest scheduled task."
- **what:** A security-motivated pattern for any internet-facing satellite service (idea.uk being the first case) to get its data into the core chassis DB without opening an inbound path: (1) local operational store on the exposed box (kept as JSON, explicitly rejecting SQLite to preserve the stdlib-only/`GOPROXY=off` build); (2) a one-way B2 "dead-drop" channel (box writes immutable per-event records via a write-only-scoped/presigned URL — reuses the same pattern Thunder adapter already uses for artefact transfer); (3) a `scheduled_tasks`-driven ingest agent on the chassis side that *pulls* new B2 records and upserts into a restricted-role schema (`business_intel`/`ecommerce`), "chassis PULLS; box never connects in." Explicit worst-case analysis: a compromised box can write junk into one B2 prefix, no more. Table design (`ecommerce.orders`, `ecommerce.taster_events`, `clients_db.idea_reports`) deliberately keeps no card data (Stripe opaque refs only).
- **sources:** `running_notes(44).md` (`PERSISTENCE_design.md` summary, two checkpoints on 2026-06-04)
- **relations:** service-deployer pattern; Thunder adapter (B2 presigned-URL precedent); storage-architecture (032, S3/B2)
- **verify-later:** whether `business_intel.idea_orders` / `ecommerce.orders` / an `idea-ingest` scheduled task exist

<!-- SOURCE: U26_misc_dirs.md -->
### Result storage split (DB paper-trail + S3 artefacts)
- **category:** storage-architecture
- **status-signal:** deployed
- **status-evidence:** basic_usage/002 states it as fact: final_result column in orchestrator_state, a website_projects table per client schema with preview/live URLs, and site-publisher's s3_upload of files.
- **what:** The record of a build lives in PostgreSQL (full workflow history + consolidated final_result JSON + website_projects metadata with URLs) while the tangible outputs (HTML/CSS/JS files, generated images/logos) live in S3-compatible object storage, referenced by URI from workflow results — "the database holds the record of what happened... the object storage holds the actual product".
- **sources:** docs/basic_usage/002storage_of_results; docs/architecture/027-create-website-creation-system (site-publisher s3_upload)
- **relations:** website-builder group; storage-architecture spine (032, S3/B2)
- **verify-later:** website_projects table; s3_upload action; current B2 storage docs
