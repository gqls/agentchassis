# Plan — build the "amend an image asset" path (operator-supplied bytes → S3, through the chassis)

## Context

Three sites (relojistas, gaswholesalers, idea.uk) have a stored `logo` asset that is not a
logo — a spec sheet, a nine-up contact sheet, and a broken row respectively — and the platform
has **no path for a human to supply corrected image bytes**. Every image ever stored arrived
via generation. The admin API edits metadata only; the admin SPA has no upload; the operator
must not read the S3 secret (classifier-refused, owner confirmed the platform should do its own
writes "through the chassis"). This plan builds that missing path as a reusable platform
mechanism, then hands off a runbook to use it for the logo swaps.

Key facts from exploration (all verified in tree / live DB):

- `S3Client.Upload(ctx, key, contentType, body)` exists (`platform/storage/s3.go:114`) —
  plain PutObject; every chassis pod builds the client when `IMAGE_BUCKET` is set
  (`platform/agentbase/agent.go:305-334`) and actions receive it as `params.StorageClient`.
- `asset-deployer` is the storage-enabled agent with a **mode-conditional chain**, live row
  verified: `check_mode`(brand_head) → `check_sprite_mode`(sprite_css) →
  `check_card_mode`(content_card) → else `deploy_asset`. Three prior migrations added modes
  the same way (`SQL_2026-07-11/13/16_asset_deployer_*_mode.sql`) — this is an established
  pattern, not a new seam shape.
- **Bytes cannot ride Kafka**: kafka-go writer default caps messages at 1 MiB
  (`producer.go:44-56` never sets BatchBytes) and the recorded doctrine is "heavy artifacts
  live in the DB, retrievable by id" (`docs026 register/system-architecture.md:157`).
  Precedent for BYTEA in DB: `chassis_intake_events.payload`
  (`sql_for_agents/249_chassis_intake_events.sql:70`).
- The `assets` upsert convention with lock-refusal exists in `StoreAssetAction`
  (`v3_site_actions.go:2630-2669`): `ON CONFLICT (site_id, asset_key) WHERE asset_key IS NOT
  NULL AND status='active' DO UPDATE … WHERE assets.locked_at IS NULL RETURNING id`; no row
  returned ⇒ locked ⇒ refusal, not error.
- After deploy, `assets.url` is flipped to a web path and `storage_path` keeps the unsigned
  S3 path (`deploy_image_asset_action.go:254-261`) — `derive_brand_head_assets` currently
  breaks on flipped rows (that is idea.uk's live failure).
- `assets.alterations JSONB DEFAULT '[]'` is documented as a history array and **never
  written by any Go code** — purpose-built for recording an amendment.
- Work-item dispatch: `status='triaged'` + `pipeline='build'`, handler routing is generic,
  ~120s/site tick. Item `result` jsonb carries the handler response back to the operator.

## Design (recommended): DB staging row + one new action + one new asset-deployer mode

Bytes flow: operator file → base64 → **psql stdin** → `asset_ingest_staging` (BYTEA) →
work item (carries only the staging id) → asset-deployer (`mode=ingest_upload`) →
validate → `S3Client.Upload` at a **new** key → `assets` row upsert (lock-honouring, id
stable, history appended to `alterations`) → staging row marked consumed.

Rejected alternatives, recorded for the council submission: base64 in the work-item spec
(1 MiB Kafka cap + state bloat, against the artifacts-in-DB doctrine); commit-to-deploy-repo
then HTTP-fetch (leaves S3 stale on partial failure, conflates deploy artefact with source of
truth); presigned PUT URL via `GetPresignedPutURL` (exists at `s3.go:217`, but needs two
dispatch round-trips ≈4+ min, exposes a write-capable URL into DB state, more operator
choreography).

## Changes

### 1. Staging table — `docs/agent_docs/sql_for_agents/265_asset_ingest_staging.sql`

House style (`\set ON_ERROR_STOP on`, guard DO block, verify DO block, commented rollback).

```sql
CREATE TABLE asset_ingest_staging (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  site_id     uuid NOT NULL REFERENCES sites(id),
  asset_key   text NOT NULL,
  purpose     text,                    -- optional; defaults to existing row's purpose
  content     bytea NOT NULL,
  sha256      text NOT NULL,          -- computed by the loader; action re-verifies
  note        text,                   -- why this amendment (goes into alterations history)
  created_by  text NOT NULL,
  status      text NOT NULL DEFAULT 'pending',  -- pending|ingested|refused|failed
  error       text,
  consumed_at timestamptz,
  created_at  timestamptz DEFAULT now()
);
```

Retention note in-file: rows are small in number (operator-driven); purge `status != 'pending'`
older than 7 days, same policy family as `chassis_intake_events`.

### 2. New action — `platform/orchestration/actions/ingest_staged_asset_action.go`

Registered in `registry.go` (Category `"site"`, IsLocal) + `RegisterActionInputSpec` init,
mirroring `derive_brand_head_assets_action.go`'s shape. Input: `staging_id` (required),
`site_id` (required — cross-checked against the staging row; mismatch = refuse).

Steps, each with the refusal path returning a result (not an error) per house style:

1. Load staging row `FOR UPDATE` where `status='pending'`; missing/consumed ⇒ refuse.
2. Re-compute sha256 over `content`; mismatch with stored `sha256` ⇒ `status='failed'`, refuse.
3. `image.Decode` to validate it is a real image; capture dimensions + mime. Not an image ⇒
   `status='refused'`. Sanity cap ~10 MB.
4. Upload via `params.StorageClient.(*storage.S3Client).Upload` at a **new key**
   `images/uploads/<site_id>/<YYYYMMDD>/<uuid>.<ext>` — never overwrites; the old object
   survives (matches the regenerate path, which never deletes either).
5. Upsert the `assets` row using the exact `StoreAssetAction` convention
   (`v3_site_actions.go:2630`): same `ON CONFLICT … WHERE assets.locked_at IS NULL`;
   locked ⇒ mark staging `refused` with the store_asset refusal wording ("approved assets are
   never overwritten — clear the lock deliberately first"). Sets: `url` = fresh presigned URL
   (`GetPresignedURL`), `storage_path` = unsigned object path (so the row survives a later
   url-flip — the defect that broke idea.uk), `origin_type='uploaded'`,
   `origin_model='operator-supplied'`, `mime_type`, `file_size`, `dimensions`, and **appends
   to `alterations`**: `{type:'bytes_replaced', at, by, note, previous:{url, storage_path}}` —
   first real writer of the column, used exactly as documented. Row **id stays stable** (update
   in place) so existing references hold.
6. Mark staging row `ingested`, `consumed_at=now()`. Return a summary (never the bytes):
   `{ingested, asset_id, s3_uri, presigned_url, width, height, bytes}`.

Unit test `ingest_staged_asset_action_test.go` for: sha mismatch, non-image, locked row
(following `derive_brand_head_assets_test.go`'s use of injected fakes where possible).

### 3. asset-deployer mode — `docs/agent_docs/sql_for_agents/266_asset_deployer_ingest_mode.sql`

Mirror of `SQL_2026-07-11_asset_deployer_brand_head_mode.sql` exactly (snapshot_agent, backup
table, verify DO block): rewire `check_card_mode.else_step` → new `check_ingest_mode`
(`input_data.spec.mode == "ingest_upload" OR input_data.mode == "ingest_upload"`), then
`ingest_staged_asset_step` → `complete`; `check_ingest_mode.else_step` = `deploy_asset` so the
default path is unchanged. Config passes `staging_id`/`site_id` from
`input_data.spec.staging_id` / `input_data.site_id` (the dispatch contract shapes).

**Ordering: image before seed** — the migration is applied only after the new binary is rolled
and pod-grepped (a seed naming an unregistered action fails at runtime).

### 4. Operator loader — `scripts/amend-asset.sh <domain> <asset_key> <file> [--note "…"]`

One command, one transaction, no cluster credentials beyond the standard psql route:
- sha256 + base64 the file **via stdin** (the ARG_MAX lesson from
  `webdesign_publish_assets.sh:97-100` is in-repo; cite it);
- single psql heredoc: INSERT the staging row (`decode('<b64>','base64')`) **and** the
  `site_work_items` row (`item_type='amend_asset'`, `handler_agent='asset-deployer'`,
  `spec={"mode":"ingest_upload","staging_id":…}`, `status='triaged'`, `pipeline='build'`,
  `item_key='amend_asset:<asset_key>'`, `created_by='operator'`) in one transaction;
- print the staging id, the work-item id, and the watch query.

`--dry-run` prints the SQL without executing (house norm for operator scripts).

### 5. Concept register + docs — same commit as the seam

- Register the mechanism in `docs/agent_docs/docs026_concept_register/register/` (imagery or
  contracts bucket): status/what/sources/relations; per the 2026-07-28 platform-seam ruling,
  registration rides the shipping commit.
- New workstream dir `docs/agent_docs/docs024_key_docs_latest/asset_amend_path/` with the
  standing five started: PLAN (this design), RUNBOOK (the loader + watch queries + verify
  sequence), NOTES, README_where_we_are, and **HANDOFF_2026-07-29_logo_swaps.md** — the
  handoff the owner asked for, covering how the three swaps use the path (below).

### 6. Council review (before/alongside commit)

Blast radius, measured for the submission (per the seam ruling: measure BEFORE submitting):
- live conditional chain read from `agent_definitions` (done — three modes, disjoint string
  equality, the rewire touches only the default fall-through);
- `ingest_upload` mode string: 0 uses fleet-wide; `asset_ingest_staging`: no name collision;
- `amend_asset` item_type: 0 rows; the action name: 0 registry entries.
Present as the **fourth instance of the established mode pattern**, citing the three prior
mode migrations. One run, `Council-Reviewed:` trailer on APPROVED.

## Sequencing

1. Go action + test + registry (commit after council verdict; trailer if APPROVED).
2. `make build-agent-chassis` (bump IMAGE_TAG), push, deploy; **pod-grep a string the new
   action logs** (positive control: an existing symbol).
3. Apply migration 265 (table), then 266 (mode) — mode only after the pod-grep passes.
4. **End-to-end verify on relojistas** (the corrected crop already exists):
   `./scripts/amend-asset.sh relojistas.com logo relojistas-logo-transparent.png` → watch item
   complete → `curl` the fresh presigned URL from the assets row (operator-readable proof the
   object landed, no S3 creds needed) → **Read the downloaded image** (the session's own
   lesson: every mechanical signal can be green while the picture is wrong).
5. Then the swaps, per the handoff (§ below).

## The handoff content (written as HANDOFF_2026-07-29_logo_swaps.md in step 5)

- **relojistas**: amend `logo` with the knocked-out crop → queue `deploy_asset` item (writes
  `/assets/images/logo.png` to the site header) → queue `brand_head` re-derivation **only
  after the favicon-aspect/approved-artefact code fix is live** (the README 07-29 sequencing:
  re-derive after the squash fix, not before) → Read card + favicon + header on the wire.
- **gaswholesalers + idea.uk**: fresh logo **generation** through the existing imagery
  pipeline (not the amend path), owner eyeballs before anything deploys; idea.uk's malformed
  row (web path, no storage_path) is healed as a side effect of its new asset landing.
- **leopardess**: untouched. Its protection is the deriving-code fix, not a row lock (README
  course-correction 07-29); its malformed row must not be "tidied".
- Landmines carried forward: triaged+build or never dispatched; Read every produced PNG;
  locked rows refuse by design — clearing a lock is a deliberate documented step.

## Follow-on task (this thread, after the amend path ships — owner decision 2026-07-29)

**The deriving-code fix, as its own coherent task and its own council run** (not folded into
the amend-path submission):
- `derive_brand_head_assets_action.go`: favicon via `resize.Thumbnail` (aspect-preserving,
  matching `composeOGCard`) composited on the brand background at 64×64 — not the current
  non-proportional `resize.Resize(64,64)`;
- the approved-artefact guard in the same action: refuse to overwrite a favicon/og-card whose
  assets row (or predecessor) is human-locked/approved — the README 07-29 course-correction
  established a row lock alone cannot protect the committed file, the deriving code must
  refuse.
- Sequencing consequence: **relojistas' brand_head re-derivation runs only after this fix is
  live**; the amend of its logo asset (step 4 E2E) can and should happen before.

## Verification

- Unit: `go test ./platform/orchestration/actions/ -run TestIngestStagedAsset`.
- Deploy: pod-grep the new action's registration string on both pods.
- Live E2E (relojistas, step 4 above): staging row → `ingested`; assets row `url` curl → 200
  with `content-type: image/png`; `storage_path` populated; `alterations` has one
  `bytes_replaced` entry with the previous URL; **Read the image**; old S3 object untouched
  (previous presigned URL still 200 until expiry).
- Failure branch (the lesson from "verify the failing branch"): run the loader once with a
  deliberately corrupted sha → staging `failed`, assets row untouched.

---

> **CORRECTED 2026-07-29, before implementation started:** the "Follow-on task"
> section above is already discharged — the deriving-code fix (favicon aspect via
> `composeFavicon` + lock honoured before the git commit) was committed by a
> parallel session as `e9e345464` while this plan was being approved. It rides
> the same build this workstream needs anyway, so the sequencing constraint
> ("re-derive relojistas only after that fix is live") is satisfied by the one
> roll. Nothing else in the plan changed.

> **CORRECTED 2026-07-29 (mid-morning):** two things moved under this plan while it was being
> executed. (1) The E2E target is now **gaswholesalers**, not relojistas — the owner split the
> sites between sessions (`bugfix_131_og_card/COORDINATION_2026-07-29_who_does_what.md`);
> relojistas' logo row is already amended AND LOCKED by relojistas-5, so an amend against it
> now exercises the refusal branch, not the success path. (2) The action's presign-failure
> fallback originally wrote a bare `s3://` URI into `assets.url` — relojistas-5's landmine 1
> (measured live) shows that form BREAKS `derive_brand_head_assets`. Fixed before first run:
> `assets.url` now always gets the path-style HTTPS form (query-stripped presign), and a
> presign failure refuses rather than storing a poisoned row.
