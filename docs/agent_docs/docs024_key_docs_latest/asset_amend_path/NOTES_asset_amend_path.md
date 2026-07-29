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

---

## 2026-07-29 (2) — council round 1: REVISE, 8 advisory objections, two of them real

Corr `0237eb64-fc03-4faa-a15a-439df6b12555`. **13 reviewers, 4 abstained, 0 unreadable, 0
high-severity.** Round 1 took ~13 minutes end to end (08:38 submit → 08:50 `complete_revise`),
well under the ~30 min the runbook budgets — the queue was empty.

Worth naming: **a REVISE here was not a rejection of the design.** Six of the eight objecting
seats spent most of their text saying the plan was well-armoured (editquality: "no audit-only
edits… minimal — six edits, each load-bearing"; bug_historian: "unusually well-armored against
this council's own recurring pattern"; guardian: "that's the plan doing the council's job for
it, not evading it"). [[a-revise-is-the-median-path]] — per round, approval is ~51%.

### The two REAL defects (fixed in code, commit `048dbd96b`)

1. **constitution — I built SQL by shell interpolation.** The loader inlined `$DOMAIN`,
   `$ASSET_KEY`, `$SHA`, the note and the summary into the statement text, defended by a
   charset guard on identifiers and quote-doubling on free text. The seat's point is that the
   ALWAYS-ON parameterisation rule prohibits interpolation *regardless of escaping
   discipline* — and it is right. **My submission actually cited the escaping as a feature**
   ("charset-guarded identifiers, quote-doubled free text"), which is the tell: I was
   defending a practice instead of not doing it. Now every value is a psql variable
   (`:'var'`), including the server-side guard's domain via `set_config`/`current_setting`.
   The base64 blob is the one unavoidable inline literal (psql `-v` is argv; a megabyte blows
   ARG_MAX) and is made inert by **construction** — refused unless it matches
   `^[A-Za-z0-9+/=]+$`, an alphabet with no quote, backslash or colon. Validation, not escaping.
2. **guidelines — `source='operator'` is not a sanctioned value.** `018_site_work_items.sql:18`
   names `'manual'`, `'improvement'`, `'side_effect'`. I invented one. Fixed to `'manual'`.
   *Interesting wrinkle:* the live table already has 8 rows at `'operator'` and 26 at
   `'operator:brochure_component_library'` from other lanes — so the drift is real and I could
   have pointed at precedent. The seat's own advice was to conform absent a ratified guideline
   change, which is the right call: **precedent in the data is not permission in the rules.**

### The other six, answered with evidence rather than code

Per [[answer-review-objections-with-evidence]] — measured over 8 rounds, answering by building
never converged; answering with queries cleared every objecting seat in one.

- **guardian + debug_historian, "verify the url-format claim against real consumers":** the
  ONLY parser of `assets.url` is `resolveReferenceAssetURIs` → `presignedURLToS3URI`
  (`deploy_image_asset_action.go:355-369`). It reads `u.Path` only, strips the slash, splits
  bucket/key. My unsigned path-style HTTPS form round-trips correctly (the query is never
  read); a bare `s3://bucket/prefix/key` yields `s3://prefix/key` — the **wrong bucket**,
  exactly the live breakage relojistas-5 measured. Every other reader avoids `assets.url` by
  design, with in-code comments saying so.
- **guardian, "refresh the blast radius":** re-run at resubmission, unchanged — and this time
  **with its denominator**, which round 1 lacked: asset-deployer is 1 active non-snapshot row
  **of 1 total rows including snapshots**, so the migration's WHERE clause cannot miss a
  sibling. Round 1's bare "1 active row" was the empty-denominator shape this fleet keeps
  hitting.
- **prior_art_librarian, "a content-search negative is not proof":** re-ran the SPA search **by
  shape, not by word** — `FormData|multipart|input type=file|FileReader|readAsDataURL|.files[|
  enctype|XMLHttpRequest|Blob(|arrayBuffer(` → zero hits, and all three `Content-Type` literals
  in the SPA are `application/json`. There is no byte-carrying surface under any name. On "the
  Job leaves nothing reusable": relojistas-5's own NOTES say "Job + ConfigMap deleted after",
  `kubectl` confirms neither exists, and no upload script landed in `scripts/`.
- **tooling_provenance, "a fourth mode migration with no doc_notes interaction":** fair, and
  the mechanism being bypassed was the point. Wrote `doc_notes`
  `24ff1ea0-fdec-413d-95e3-cd44b09a71c9` (`subject_type='pipeline'`, key `asset-deployer`).
  **Trap met:** `subject_type` has a CHECK constraint — `'agent'` is rejected; the vocabulary
  is `tool|pipeline|experience|action|experience-pattern`.
- **editquality, "the IMG-065 register claim has no covering edit":** the *commit* had it
  (`f2c9bd2cc` touched the register and the index); the **plan block** didn't. Plan and commit
  disagreeing is itself the defect — added as an explicit edit.
- **bug_historian, "correcting the row does not fix three live pages":** correct, and now said
  out loud instead of left implicit. This change fixes the **source of truth** only;
  propagation is two existing work items (`deploy_image_asset`, `needs_brand_head_assets`),
  deliberately separate so the operator eyeballs the amended source first — which is the
  taste-gate this whole bug exists to serve.

---

## 2026-07-29 (3) — council round 2: APPROVED

`0237eb64-fc03-4faa-a15a-439df6b12555`, second report 12:33:38Z. 13 reviewers, 4 abstained,
**0 unreadable**, decision `approved`, `complete_approved/COMPLETED`. Round 2 took ~11 minutes.

**Two rounds, ~24 minutes of council time total, and the shape held:** the round that changed
CODE changed it for exactly two objections (the SQL construction and the `source` vocabulary);
the other six were cleared by **queries**, not edits. That is
[[answer-review-objections-with-evidence]] reproduced — answering by building never converged
over the 8 rounds that memory records; answering with evidence cleared every objecting seat in
one. Worth noting the counterfactual: had I "fixed" the six by writing more code (a url-format
abstraction, an upload-surface audit, a propagation step folded in), the plan would have grown
and drawn fresh objections on the new surface.

Per trailer discipline, `Council-Reviewed:` goes on a commit only now that the verdict is
APPROVED — and it carries the **submission correlation**, which is the key the artifacts are
written under.
