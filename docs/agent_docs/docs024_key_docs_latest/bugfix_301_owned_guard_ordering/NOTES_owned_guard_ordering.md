# NOTES — bugfix 301 (owned-guard ordering) — append-only, newest at the bottom

## 2026-08-19 — session 1: validity, ownership, design

**Ownership check** (`scripts/who-owns.py 301`, ~11:50 UTC): verdict "OWNED or recently
active", but reading the two named lanes' handoffs settles it — `bugfix_277_required_fields_repair`
shipped Tier 1 (status change, live on v1.0.1314) and explicitly does not claim the ordering
fix; `vigilant_designer_offer_analysis` (the filer) offers it: "Take `bugs_open/301`" in
HANDOFF_2026-08-18. Taking it. `needs_diagnosis` queue: 0 rows (checked before filing 090).

**Bug still valid** [MEASURED 2026-08-19 ~12:00 UTC]:
- 18 `page-build-handler`/`page-content-writer` orchestrations since the roll (07:52Z);
  the 10:18 cluster shows the shape exactly: handler created 10:18:27 ends `complete_error`,
  its writer child 10:18:42 COMPLETED, resolver child 10:19:16 COMPLETED — full chain run,
  then refused.
- 146 open findings on owned pages routed at `handler_agent='page-build-handler'`
  (live table only — archive not needed for "queued now"; statuses
  detected/needs_human_review/unresolved/failed).
- Tier 1 IS behaviourally live now: 2 `wont_fix` rows since roll with
  `result ? 'owned_page_refusal'`, 1 new `owned_page_review`. (The 277 lane's handoff
  asked for this check with both controls — I have only run the positive half; noted for
  them rather than claimed as their verification.)

**Design facts established first-hand** (file:line in PLAN):
- Live workflow graph read from `agent_definitions` (dump in scratchpad, steps listed in
  PLAN): every write path ends at `save_sections`; no-op arms park without touching the page.
- `save_page_sections_action.go` guard: unconditional refuse on owned, error LEADS with
  `ownedPageSkipReasonPrefix`; emit added by 295.
- Tier 1 mechanism: `routeToErrorStep` copies the error verbatim to `__step_error.message`
  (`coordinator.go:3697`); `UpdateWorkItemStatusAction` matches the prefix and stamps the
  configured `owned_page_refusal_status` (`v3_site_actions.go:5527-5545`). An early refusal
  that errors with the same prefix through `error_step → mark_item_failed` inherits all of it.
- Carrier census: `load_page_record` is carried by exactly 2 live agents:
  `page-build-handler`, `tool-recreation-handler`. The latter is the tool pipeline —
  must never refuse owned pages — which forces the opt-in default-OFF shape.
- Order tolerance: `LoadPageRecordInputSpec` has `CheckConfig: true`, no `StrictConfig` →
  `platform/validation/workflow.go:185-195` warns once on an unknown key, never rejects.
  So config-first is safe; binary-first is dormant.
- RFC_022 budget: `cmd/config-key-audit/optionalbudget.go:14-21` counts `Optional` only,
  deliberately not `ConfigKeys`. New key goes in `ConfigKeys` (it is a setting); trade-off
  disclosed in the council submission. Budget run 2026-08-19: 0 shared actions over N=10;
  `load_page_record` not in the widest list (3 optional keys).

**090 filed** (~12:10 UTC): intake `7281193f-59c2-489a-a9f2-fd4d58408cf5`, run
`dd61df1b-0d93-46e6-9065-1e0b9623379a`. Question put to it: does any branch of
`page-build-handler` write page content other than through `save_page_sections`?
Budget ~30 min per CLAUDE.md; do not retry on a missing orchestration row.

**Missteps this session:** none yet worth a WRONG_CALLS row. (Watching for: quoting the
bug file's 39/74 figures as live — both are dated aggregates; the 74 is policy-join
contaminated per the 277 lane's §8 correction.)

## 2026-08-19 — session 1 (continued): built, tested, submitted

- Go change: `refuse_owned_page` in `LoadPageRecordInputSpec.ConfigKeys` +
  `refuseOwnedPageIfConfigured` on BOTH return paths (authoritative-id and name/id);
  marker-emitters comment added to `owned_page_guard.go`. Tests: 4 cases, incl. the
  key-absent control (mock proves the default path never reads `rebuild_policy`).
- **Proven against HEAD, not the dirty tree**: another session has dirty files in the SAME
  package (`check_endpoint_health_action.go`, `conditional_branch_action.go`, two test
  files), so the in-tree green proved nothing — re-ran `git archive HEAD` + only my 3 files:
  build + full `actions` suite green.
- Migration `488` + `_ROLLBACK`: dry-run apply+rollback in ONE rolled-back transaction
  (both NOTICEs fired); double-apply probe ABORTED on the already-set guard ("could the
  control have fired" — it did). 486 collides already; 488 was free.
- Parity test run (RFC_022 obligation): **FAILS, pre-existing, NOT mine** —
  `save_page_meta_description` (commit `aeccfc595`, the 320 lane) declares 5 optional keys;
  `deployments/kustomize/services/optional-key-budget-check/base/check.py` literal says 0.
  The daily budget cron is BLIND to that action. My key is in ConfigKeys (not counted), so
  my change cannot fix or worsen it. TO DO this session: regenerate the literal per
  check.py's own comment as a separate narrow commit + re-apply the overlay (CLAUDE.md's
  documented procedure), and tell the 320 lane.
- Council submitted ~12:35Z: **corr `c7bc1b9e-97c8-4f3e-8a4f-b3a7029505ee`** (save this;
  budget ~30 min; find the run by payload not printout). Committing with
  `Council-Submitted:` per the norm.
- Register: **PBP-045** appended (page-build-pipeline.md + index row), additions-only
  verified by `--numstat` (9/0, 1 long row in index).
- Noted in passing, post-roll: migration `473` is now APPLIED (`162ab424a`) and its
  pre-registered prediction was WRONG in a fourth-outcome way — the 184/277 lanes' story,
  not mine, but the re-route "created a pair OUR gate holds"; my fix does not change that
  either way (early refusal lands `wont_fix`, which the gate ignores).

> **CORRECTED 2026-08-19:** the timestamps in this file reading "~12:00 UTC"/"~12:10 UTC"/
> "~12:35Z" above are BST misread as UTC — the real times are ~1h earlier (the 488 ledger
> row, written immediately after the council submission, reads 11:05:25+00). The measured
> DATA is unaffected (queries carried their own timestamps); only my narration drifted.
> Caught by the schema_migrations `applied_at`.

## 2026-08-19 — 488 APPLIED and ledger-recorded (11:05:25Z)

`488 OK` NOTICE fired; ledger row `applied_by='bugfix_301_owned_guard_ordering lane'`.
The config half is now LIVE (old binary: warns once + ignores the key; the error-routing
alignment for genuine load errors is active now). The refusal activates at the next
chassis roll carrying commit `6be66bceb`.

## 2026-08-19 — checkpoint (session pausing on usage limit)

DONE: fix committed `6be66bceb` (Go + tests + 488 + PBP-045 + bug-file note + standing
five); 488 APPLIED + ledger-recorded 11:05:25Z; parity drift fixed separately
(`1c16eb692`, one line, overlay re-applied — cronjob `configured`).

STILL OPEN, in order:
1. **READ the council verdict** — corr `c7bc1b9e-97c8-4f3e-8a4f-b3a7029505ee`. Query in
   CLAUDE.md/097 output; act on REVISE/REJECTED (code is already on the shared branch).
2. **READ the 090 verdict** — run `dd61df1b`, was at call_diagnoser ~11:10Z. If it refutes
   the "no write path bypasses save_sections" premise, act BEFORE the next roll (rollback
   file exists and is exercised).
3. **Post-roll behavioural verification** — RUNBOOK, BOTH controls (owned → wont_fix +
   refused_by='load_page_record' + no writer child; generic → writer runs and saves).
   Then, and only then, 301 moves toward closed (fixed-AND-live bar).
