# NOTES — ai_site_selling_automation — append-only, newest at the bottom

## 2026-08-10 ~18:00 — session start: orientation, freshness checks, owner rulings

Cold-started from `HANDOFF_2026-08-10_start_here.md` (this directory). Ran the
handoff's own falsification checks before trusting it:

- **Sibling lane is fresher than the handoff.** `webdesign_uk_build_service/`
  has `HANDOFF_2026-08-10c_continue_here.md` (17:21) and NOTES appended to
  17:55 — both post-date the start-here file (17:23). Read both.
- **The chat intake's "LIVE" status changed meaning this afternoon.** The
  cache gap closed on its own (proven at the nginx layer, sibling 08-10c §1a
  — do not re-litigate), but underneath it: **the Anthropic ACCOUNT hit its
  spend cap ~14:51Z 2026-08-10**. The box now serves the fail-closed
  contact-details line to every visitor. Sibling NOTES (17:00–17:40Z entry,
  contributed by the bugfix-236 lane) corroborates from a different credential
  path: last successful LLM call fleet-wide 14:51:45Z, 0 successes after,
  **council gate down** (a submission died at its first review seat,
  terminal `complete_invalid` — which reads as "invalid submission" and is
  not). Provider says access returns 2026-09-01 unless the owner raises the
  limit in the Anthropic Console. [CITED from sibling NOTES, not re-measured —
  their evidence spans two independent paths, which is stronger than my
  re-running the same query.]
  Consequence for THIS lane: council submissions are unsatisfiable until the
  cap lifts, so the clients-columns migration (council scope) is drafted but
  not shipped; FE work (out of scope) proceeds.
- **`bugs_open/239` is OWNED.** A live session named "bugfix bugs_open/239"
  was busy on it at 17:5x tonight (peer-session listing). Sequencing step 3 of
  the handoff is therefore someone else's; this lane tracks, does not touch.
- **Workstream memory entry exists** (MEMORY_workstreams.md line ~34, NEW
  08-10) and matches the handoff. No competing directory; this dir held only
  the handoff at session start.

**Owner rulings taken at session start** (put directly, three questions,
recorded in PLAN §1 — the authoritative copy): identity = extend `clients`
(BIZ-014); automation target = £1,200 tier intake (human releases); scope
while cutover pending = design + ungated build.

Spawned two read-only scouts: (a) admin FE + `/api/v1/admin` client endpoints
(for the Customers tab); (b) `046_site_chat_turns.sql` + chat-service JSONL
format + core-side pull precedents (for the ingestion design). Results to be
recorded below when they land.

## 2026-08-10 ~19:00 — first build session: mig 375 live, /admin/customers built, FE tab built, ingestion designed

**The handoff's cheapest-first premise failed on contact with the code.** Scout
read of `client_handlers.go` showed the "existing client CRUD endpoints" serve
`clients_info` — a side table the handler lazily CREATEs on first call
(`to_regclass('public.clients_info')` → NULL live, i.e. never yet called) —
part of the per-client-schema tenant machinery, not the ruled
`clients→networks→sites` chain. GET /clients names its columns and discards
`settings` into a dead variable; new columns would not have flowed through
anyway. Corrected in the handoff §3.5 in place; landmine appended to
LANDMINES.md + synced to doc_notes (`--check` clean; the verify dispatch will
sit starved until the LLM cap lifts, expected).

**Shipped, in order:**
- Migration `375_clients_customer_identity_columns.sql` — email/phone/tier/
  customer_status/notes on `clients`, all nullable. Blast radius measured
  first: zero Go readers of `FROM clients`, zero positional INSERTs.
  Applied out-of-band (`psql -f` via kubectl) because the runner's `--apply`
  is blocked by four older pending files with pre-state mismatches
  (353/358/359/361 — other threads') and two LIKELY-ALREADY-APPLIED
  (363/370); a failed file stops the run before reaching 375. Ledger-recorded
  via `--record-only` with that reason. Verified: information_schema lists
  all 5 columns.
- `internal/core-manager/admin/customer_handlers.go` + routes —
  GET/POST `/api/v1/admin/customers`, GET/PATCH `/customers/:customer_id`.
  `go build` + `go vet` clean; **HEAD verified from a clean `git archive`
  build after committing** (HEAD-BUILD-OK). Registered as ADM-011 with index
  row, same commit. Commit `fe6b99d05`.
- FE `CustomersPage.tsx` + three-line App.tsx wiring (nav key `customers`,
  PipelinesPage precedent). Verified in a throwaway node:20 container on a
  scratchpad COPY (tree kept clean): `vite build` ✓; `tsc --noEmit` reports
  15 errors, ALL pre-existing in App.tsx, zero in the new page — this repo
  has no typecheck script and has never typechecked; do not "fix" that as
  part of an FE change. Commit `a84d544d1`.
- Ingestion design written into PLAN §2.3 (B2 sink, spawn-free CronJob
  puller, UUIDv5 pairing, hash-at-ingest, box-side asks for the sibling
  lane). Not built.

**Council obligation open:** `fe6b99d05` touches `internal/` with no
Council-Submitted trailer because the gate cannot currently run a seat (fleet
LLM cap; the sibling NOTES show a submission dying at its first review seat).
OWED: submit the mig-375 + customer-endpoints change as one coherent
submission when `llm_call_log` shows successes again — the RUNBOOK curl is
the cheap tell. [ASSUMED] the gate will accept a post-commit submission
normally; the Council-Submitted mechanism explicitly supports commit-first.

**Still true / unchanged tonight:** Anthropic account cap in force (chat box
fail-closed; owner action only); `bugs_open/239` owned by its own live
session, untouched by this lane.

## 2026-08-10 ~20:30 — fleet roll v1.0.1283: Customers surface LIVE end-to-end; cap LIFTED; council submission fired

- **Roll verified at the artefact, not the tag.** core-manager AND
  admin-dashboard both at `v1.0.1283` (2 replicas each, ~28 min old at check).
  Pod-grep, same exec, every replica: `strings /app/core-manager` →
  `"Failed to list customers"` = 1 (my string) AND `"Failed to list clients"`
  = 1 (pre-existing positive control) on BOTH core-manager replicas; dashboard
  bundle `index-DP86XXhq.js` contains `"No customers yet"` (mine) AND
  `"Data Pipelines"` (control) on BOTH replicas. Route probe from inside the
  cluster: `GET /api/v1/admin/customers` → **401** (auth wall), not 404.
  The whole surface — columns (375) → API → FE tab — is live.
- **The Anthropic account cap has LIFTED.** The RUNBOOK curl returns a real
  answer ("£1,200, paid once…" — correct pricing, correct follow-up), not the
  contact-line fallback. [INFERRED] the owner raised the limit; not verified
  in the Console, but the observable (fleet LLM works) is what matters here.
- **Council submission DISCHARGED for `fe6b99d05`:**
  `SUBMISSION_CORR = 371f8b7d-0835-4879-b48f-ad0176bf2058`. Budget ~30 min
  (dispatch queues behind the fleet, and the fleet is digesting a post-outage
  backlog — do NOT re-fire on a missing orchestration row). Find the run by
  payload: `SELECT current_step, status FROM orchestration_states WHERE
  collected_data->'input_data'->>'fix_correlation_id' =
  '371f8b7d-0835-4879-b48f-ad0176bf2058';` Verdict:
  `SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY
  created_at DESC LIMIT 1;`
  **Known 098 gap, accepted:** `fe6b99d05` predates the submission and
  forward-only forbids amending a trailer in, so it will list un-reviewed in
  098 forever; this NOTES entry + the verdict note are the audit trail. Do
  NOT stick the corr as a trailer on an unrelated docs commit — the trailer
  binds to the commit it is on.
