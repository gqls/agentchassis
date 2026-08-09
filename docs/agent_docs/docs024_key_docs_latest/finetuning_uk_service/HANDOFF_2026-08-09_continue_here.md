# HANDOFF 2026-08-09 — finetuning.uk service: 186 closed, orphan sweep LIVE, Phase 0 is next

**This is the COLD-START document for the lane** (supersedes
`HANDOFF_2026-08-08_continue_here.md`, which carries the fuller Phase −1
background — token rotation mechanics, pricing measurements, model menu; read
it second, not first). RUNBOOK has the commands; PLAN_2026-07-31 has the
approved design; NOTES has the evidence trail and every misstep; README is
the owner's plain-prose log. Everything below verified 2026-08-09 midday
unless dated otherwise.

## What this lane is

finetuning.uk (live Class A site) is to offer a real, paid, demo fine-tuning
service: a few pounds via Stripe Payment Link, one small-model QLoRA
fine-tune on a Thunder Compute GPU, before/after eval + adapter/GGUF download
+ a booked GPU playground hour (a6000, $0.35/hr — measured, there is NO
`l40s`). Concierge first, automate later. Front door = a `finetune` route
group in tools-api on the island; cluster only ever pulls; Thunder strictly
stop/start with artefacts in B2.

## ⚠ Lane boundary (unchanged)

The site's FRONT END belongs to the `finetuning_uk_repair` thread
(`7b4e88a8-…`). This lane is service backend only. Phase 1 (offer page) is
BLOCKED on coordinating with that thread — do not author page content or
fire rerenders at finetuning.uk without checking `MEMORY_workstreams.md`.

## State of the world

### bugs_open/186 — CLOSED IN SUBSTANCE (fixed AND live, verified 08-08)
NULL-`instance_ip` no longer kills the decommission: verified by the NULL-IP
re-drill on `v1.0.1267` (row `999999` → `decommissioned` in ~30s, real
Thunder 404≡success, cost stamped, table back to baseline). File stays in
`bugs_open/` per owner ruling 08-06. Nothing further owed. NOTE for any
similar future fix: the diff was type-only — **no greppable string** — so
the behavioural drill IS the proof; the 114 template is that recipe.

### FTW-042 orphan sweep — LIVE AND VERIFIED (08-09), council round 3 in flight
The reconcile the reaper structurally cannot do: every 6h,
`thunder-orphan-scan` asks Thunder for its own account view
(`list_instances` on the adapter, read-only) and compares against
`thunder_instances`. Orphans (billing with no live row) file as
`thunder_orphan` work items on `system.internal` (severity high, via the
shared `insertWorkItem` — dedup + two-strike); ghosts are reported, not
filed. 30-min grace absorbs the provision INSERT-after-up window; unknown
`createdAt` cannot hide. **No remediation authority, by design.**

- **Live state:** fleet `v1.0.1273` (pod-grepped, both chassis replicas +
  adapter, positive + spelling-negative controls). `sql_for_agents/342`
  applied AND **recorded in the schema_migrations ledger** (`--record-only`
  — the ledger DOES cover `sql_for_agents/`; run-migrations.sh header line
  5). First verified run COMPLETED 11:42:06Z: `vendor_billing:0` (= manual
  `instances/list` `{}` same session), `db_rows:23/db_live:0` (= table
  read), `clean:true`.
- **Council trail** (one correlation: `7ffecfa2-ff96-4d73-be0d-25eb9589c6df`):
  r1 REVISE (gating: hand-rolled work-item INSERT → adopted the shared
  `insertWorkItem`; see WRONG_CALLS 08-08) · r2 REVISE (gating: my "no
  ledger" claim was FALSE — see WRONG_CALLS 08-09; plus system.internal
  assert, ROLLBACK sidecar, category fix, all shipped `cfaa93126`) ·
  r3 REVISE 12:00Z (all verification requests — dispositioned by
  measurement, including proving the seed's DO/RAISE fail-closed by
  INDUCING the config-nested regression on the live row, rolled back) ·
  **r4 submitted ~12:15Z 08-09, EVIDENCE-ONLY (no code/config changes) —
  READ THE VERDICT**: `SELECT created_at, metadata->>'decision' FROM
  diagnosis_artifacts WHERE correlation_id='7ffecfa2-…' AND
  kind='council_report' ORDER BY created_at;` (4th row = r4). If APPROVED:
  the `Council-Submitted:` trailers on `81484df8a`/`ecbb0f362`/`cfaa93126`
  are credited automatically by 098 — nothing to amend. If REVISE again:
  iterate on `council_r4_submission_ftw042.json` in this directory; note
  the trail is now 4 rounds on ONE coherent task — if the round is again
  only verification-requests a council tier cannot itself satisfy, consider
  the RUNBOOK_council_gate guidance about when to stop iterating and let
  the 098 report + a human carry it.
- **Commits:** `81484df8a` (build) · `2ef4ab581` (gofmt) · `ecbb0f362`
  (r1 revision: shared door) · `95a455d35` (output_field correction, LIVE
  proof) · `cfaa93126` (r2 revisions) · plus NOTES/WRONG_CALLS/LANDMINES
  docs commits.
- **Two open caveats, both disclosed:** the FILING path has never fired
  against a real orphan (first real one is its live proof); a one-off
  double-fire at a hand-kicked first tick (2 runs 30s apart, stamp race,
  no recurrence — dedup absorbs it, not chased).

### Landmines MINTED by this lane (all in LANDMINES.md, synced, verifier ran)
- **`output_field` nested in a step's `config` is INERT** — must be a
  STEP-LEVEL sibling of `action` (`processor.go:434`). The reaper's own
  seed (028/114) models the WRONG form and works only because nothing reads
  its response. A specimen whose step name and output_field are
  word-reversals cannot verify the storage rule — read
  `applyResponseToState` (coordinator.go:2636) instead.
- The scan's earlier lane landmines stand: drill-row hygiene, numeric drill
  ids only when `instances/list` is `{}`, `owner_agent_type` not
  `agent_type`, never `2>/dev/null` a psql check, terraform owns the
  Thunder secret, `reaped_at` is a dead column.

## Next steps, in order

1. **Read the r3 council verdict** (query above). APPROVED → record in
   NOTES, done. REVISE → objections come back with checks answered;
   iterate on the saved r3 JSON, resubmit with `RESUBMIT_CORR`.
2. **Phase 0 — the measured rehearsal (~$1–2, NEEDS OWNER-ISH SUPERVISION).**
   Unchanged from the 08-08 handoff, which has the full checklist: correct
   `thunder_config` rates just before (RUNBOOK §1 step 5); re-tar + upload
   the scripts bundle (RUNBOOK §2, md5 round-trip — git and B2 deliberately
   differ until this); one small-model run end to end (SmolLM2-1.7B or
   Mistral 7B, timed and priced per stage, `RUN_SH_DONE` reached, adapter
   genuinely in B2); merge → GGUF; **playground rehearsal on an a6000**
   (provision→first-token, tok/s — decides the booking shape); then enable
   `thunder-training-monitor` (stays `enabled=f` until `DONE⟹durable`
   proven). Closes flywheel gates FTW-032/035.
   **The orphan scan is now live during Phase 0** — a provision that
   crashes mid-flight will be caught within 6h + grace; the §1b manual
   check remains the same-day cross-check habit.
3. **Phase 1 — page + payment link:** BLOCKED on the front-end thread
   (boundary above). Owner calls also pending: final price (after Phase 0),
   playground booking shape, sample datasets, registering the island's
   widened Stripe posture when Phase 2 ships.

## Where everything is

- **This dir:** PLAN (design) · RUNBOOK (commands: §1 token/§1b billing
  check/§1c rotation/§2 bundle/§3 queries) · NOTES (evidence + missteps) ·
  README (owner log) · council r2+r3 JSONs · SUMMARY_2026-07-31(+b).
- **Code:** `internal/adapters/thunder/` (adapter: provision, decommission,
  ssh, data URLs, list_instances) · `platform/orchestration/actions/
  thunder_*.go` (dispatches + reconcile) · `docs/agent_docs/sql_for_agents/
  342*` (scan seed + rollback sidecar).
- **Tasks in the harness:** #2 (orphan sweep) — completed once the r3
  verdict is read and recorded.
- **MEMORY:** workstream line + `finetuning-uk-service-workstream.md`
  updated 08-09.
