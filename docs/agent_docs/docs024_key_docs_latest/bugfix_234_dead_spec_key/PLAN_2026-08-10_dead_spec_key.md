# PLAN — bugfix 234: the dead `spec` config key, and closing its class

**Lane opened 2026-08-10.** Case file: `bugs_open/234_HANDOFF_2026-08-09_create_work_item_spec_key_is_dead_so_refresh_site_components_never_reaches_the_rerender_gate.md`.
Full approved plan (owner, 2026-08-10): this file is its working copy; deviations get recorded here as corrections.

## What we are doing, and why this shape

Three live `create_work_item` steps set a config key `spec` the action has never read, so
every item they file carries `spec = '{}'`. The improvement-loop consequence is that
`refresh_site_components: true` has never reached the rerender gate — 16/16
`improvement_rerender_*` rows empty (re-measured 2026-08-09; positive control 5,040
non-empty fleet-wide).

Two owner decisions taken at plan time (2026-08-09/10 session, recorded in the case file):

1. **RESTORE the flag** on `improvement-loop.insert_rerender_item` via `spec_literal` —
   not delete it. Grounds: bug 226's divergence guard is LIVE (pod-verified both replicas,
   v1.0.1274) so a chrome overwrite is now recorded and recoverable; and the flag is filed
   daily by 8 other producers (~5–15 rows/day), so this is restoring one lost path to the
   fleet norm, not switching on a dormant behaviour.
2. **Ship BOTH class fixes**: `StrictConfig: true` on `create_work_item` (existing
   machinery; its stated precondition — recognised set complete against every live step,
   all depths — verified met on 2026-08-09) **and** a new opt-in
   `ActionInputSpec.RemovedConfigKeys` field (retired key → message naming the correct
   spelling; hard validation error; empty map = OFF, per the owner ruling of 2026-08-02 #2).

Why both: StrictConfig catches the NOVEL unread key (what `spec`, `spec_fields`,
`domain`, `commit_from` all were); RemovedConfigKeys catches the REINTRODUCED retired key
with an error that names the replacement, and works on actions not yet ready for strict.

## Phases

| phase | what | state |
|---|---|---|
| 0 | this directory + dated corrections into the case file | **DONE** (`90f7e8b3b`) |
| 1 | data migration translating the three carriers; seeds 054/269/291 same commit; DO/RAISE guards proven disconfirmable by mutation; applied by hand + `--record-only` | **DONE** — landed as **364** (363 was taken between check and write), applied + verified 0 carriers at all depths (`b0d6f6ce0`) |
| 2 | Go: `RemovedConfigKeys` field + helper, removed-key hard error in `checkStepConfigKeys` (BEFORE strict/unknown), `create_work_item` declares `spec` removed + flips strict, 9 tests, audit surfaces, register SCR-007 same commit | **DONE** (`d278d7b25`), every guard mutation-proven |
| 3 | council submission before the code commit; `Council-Submitted:` trailer | **DONE** — corr `3eb0d1f1-6929-4131-bbef-c636256aa667`, verdict pending at close of session |
| 4 | IMAGE_TAG bump + `make build-agent-chassis`; rides the next fleet roll | **DONE** — v1.0.1278 built from HEAD, strings-verified, pushed (`0631d6996`); NOT deployed (whole-fleet releases are the owner's) |
| 5 | verification: data half at a FILED ROW; code half post-roll (pod-grep + strict canary); close-out bookkeeping | **bookkeeping DONE; two proofs PENDING** — see NOTES 08-10 final entry |

## Ordering constraint (load-bearing)

The migration is live immediately; the Go hard-error is inert until an image rolls — and on
this tree **committing is shipping** (any session's roll carries HEAD). So the migration
must be APPLIED before the Go change is COMMITTED. Phase 2's commit gate: re-run the
all-depths carrier query and see 0 rows first.

## Decisions & corrections log

- 2026-08-10: case file's two stale premises corrected (226 guard now live; flag not
  fleet-dormant). Owner chose RESTORE + BOTH mechanisms. See case file addendum.
