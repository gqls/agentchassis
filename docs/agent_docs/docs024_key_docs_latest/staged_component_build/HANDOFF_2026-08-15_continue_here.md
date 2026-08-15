# HANDOFF — 2026-08-15, fresh chat starts here: 248 is CLOSED — the drain executed as pure bookkeeping (84/84 already served, zero redeploys), and both resolver RFCs are RULED

**Supersedes `HANDOFF_2026-08-14c_continue_here.md`.** Everything below was measured by this
session 2026-08-15 ~09:00–10:30Z, with owner rulings received in chat the same morning.

## 1. What happened, and the number that inverted

The owner approved the full drain (with a concurrency check), authorised one-off **manual**
bookkeeping for the stale-row-only class, and answered the site-lock question (a lock system
exists — `sites.locked_at`, honoured by `find_dispatchable_site` — and none of the 12 sites
needs locking during heavy development).

The wire check then inverted the whole design: **all 84 bucket D+E rows already served a
genuine image at the reader-derived path** (curl per row; per-domain must-be-absent controls
404 on all 12 domains; content-types `image/*`). The pilot's 2-of-13 skip rate was actually
100% at D+E scale — organic rerenders/regenerations had repaired every artefact since the fix
went live; only the rows lagged. So the drain fired **zero redeploys**: one guarded
transaction corrected 85 rows (`UPDATE 85`; pre-image in
`DATA_2026-08-15_bookkeeping_preimage.tsv` beside this file) and cancelled 3 dormant
`unresolved` items whose promotion would have overwritten served files (incl. leopardess's
RETIRED logo). **Census now: 0 active marker rows** (11 superseded/retired remain by design —
every future marker census MUST carry `AND status='active'`).

**`bugs_closed/248_HANDOFF_2026-08-10_undeployed_asset_repair_…placeholder_name.md`** — moved
this session; fixed AND live both hold (binary-probed `v1.0.1300` with two-way controls, both
symptom sites 200, backlog zero). Full evidence: its CLOSING CONTRIBUTION + NOTES `## 2026-08-15`.

## 2. The two RFCs are ruled — implementation is the remaining platform work

- **RFC_028: RULED and already implemented** by the 231 lane (`260cb2393`, Council-Submitted
  `5d491545`): contract of record on `ExtractActionInputs`, `IsDottedPathReference` predicate,
  enforced arm budget (AST-walk test, floor 10 / ceiling 15), alias-collision guard. This
  lane's only involvement was correcting the RFC file's stale OPEN status.
- **RFC_029: RULED 2026-08-15 (owner-delegated), implementation NOT started.** §9 of the RFC
  is the ruling: **unique-or-nothing** (`findFieldRecursive` collects all matches; unique
  value → resolve deterministically shallowest-first; conflicting values → resolve nothing +
  WARN), phased **instrument-first-then-flip** (two chassis builds, observation window
  between), an opt-in `!` strict marker in `input_mapping` (mirror of `?`) with **`asset_id`
  on the two 401/402 callers as named first adopter**, and the arm budget extended to the
  inner chain (floor 5 / ceiling 8, same mutation-proven AST pattern) with descriptive arm
  names replacing the colliding "Strategy N" numbering. **Whoever implements: repair
  `TestDefaultBeatsTheRecursiveSearch` FIRST** (fails on pristine HEAD, stale vacuity
  control — noted in `260cb2393`'s message), then one coherent council-gated task.

## 3. Owner rulings routed elsewhere this morning

- **`tool-gas-unit-converter`: ruled — build a fleet-wide repair handler for
  `required_fields_missing`.** Filed as **`bugs_open/277`** (this session). Not this lane's
  build; it needs a platform session (new handler + registration + dispatch wiring).
- **Tracker feeds**: owner will wake the `model_directory_pipeline` lane; the ready-to-run
  fix is `model_directory_pipeline/FINDING_2026-08-10_the_tracker_publisher_was_reverted_and_never_re_extended.md`.
- **gaswholesalers stray `/assets/images/logo.jpg`**: owner said leave it for now. Unchanged.

## 4. Observations recorded, not acted on

- **~49 dormant off-census `unresolved` `undeployed_asset` items** (robot-hands ~44,
  July-dated; gaswholesalers 3; ai-agent-orchestration 1; others). Never dispatched, pointing
  at assets with NO placeholder marker. Unowned discovery-hygiene question: should ancient
  `unresolved` items expire? Do not promote any of them without the full wire-check ritual.
- The 268 CTA fleet batch was mid-flight during this session (pages, not assets — no
  conflict; their pre-flight measured 248-exposure zero).

## 5. Session-start checklist (much shorter than 14c's)

1. `git log --oneline -10`; re-read this file FROM DISK.
2. If touching the resolver work: read RFC_029 §9 + RFC_028's status block first, then the
   code — both files' rulings are newer than most prose about the resolver.
3. If anything asset-shaped resurfaces: census WITH `AND status='active'`, wire-check with a
   must-be-absent control per domain, and grep the closed 248 file for the site before
   trusting any residual row. The lock check is `sites.locked_at IS NULL`.
4. `scripts/who-owns.py 277` before starting the repair-handler build — a platform session
   may have claimed it.
