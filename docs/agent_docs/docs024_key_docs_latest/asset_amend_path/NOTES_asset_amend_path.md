# NOTES — asset-amend path

Append-only, newest at the bottom.

---

## 2026-07-29 (1) — designed, built, tested locally; the parallel fix landed mid-plan

Design settled by exploration (two Explore agents over the storage/deploy path and the ingress
options; full findings summarised in PLAN):

- **The decisive constraint is Kafka, not S3.** kafka-go's writer defaults to a 1 MiB cap
  (`producer.go` never sets `BatchBytes`) — the broker's 5 MiB is not the binding limit. So the
  bytes go through the DB (BYTEA staging, precedent `chassis_intake_events.payload`) and the
  work item carries only the row id. This also matches the recorded doctrine: "heavy artifacts
  live in the DB, retrievable by id".
- **`GetPresignedPutURL` already existed** (`s3.go:217`, thunder-adapter) — the presigned-PUT
  alternative was real, rejected on operator choreography (two ~2min dispatch round-trips, a
  write-capable URL parked in DB state). Owner chose the staging design.
- **`assets.alterations` finally gets its first writer.** Documented as a history array since
  the schema landed; no Go code ever wrote it. The amend records
  `{type:'bytes_replaced', at, by, note, previous:{url, storage_path}}`.
- **`storage_path` is always populated** — the url-flip defect class (idea.uk: `assets.url`
  holding a web path, derivation dead) cannot recur for amended assets.

Built: `ingest_staged_asset_action.go` (+registry), migrations 265/266, `scripts/amend-asset.sh`.
Registered IMG-065. Six unit tests pass (`TestIngestStagedAsset_*`, `TestFormatToExtAndMIME`).

Two things worth recording from the build itself:

- **Deadlock found by reading, not by running:** the in-tx "mark staging ingested" failure path
  called `refuse()`, which writes the SAME staging row on a fresh connection while the tx still
  holds its lock — the two would block until timeout. Fixed with an explicit rollback before the
  refusal. The staging claim being a separate autocommitted statement is what confines the
  hazard to that one path.
- **The lock check moved BEFORE the S3 upload** during test-writing: originally the locked-row
  refusal sat inside the tx (after upload), which both wrote an orphan object for a refused
  amend and made the branch untestable without a storage fake. The pre-check refuses early; the
  in-tx `FOR UPDATE` re-check stays as the TOCTOU-safe enforcement. The test asserts the
  ordering by passing NO storage client — reaching the upload would fail the type assertion.

**Mid-plan collision, benign:** the deriving-code fix this workstream's sequencing depended on
(favicon aspect + locks honoured before the git commit) was committed by a parallel session as
`e9e345464` while the plan was in review — discovered when a test I was modelling on named a
`composeFavicon` that hadn't existed when I read the file an hour earlier. Shared-tree rule
held: check `git log` on the file, not your memory of it. It rides the same build as the new
action, so one roll satisfies both.
