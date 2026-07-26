# HANDOFF — 2026-07-26 — chrome lock gate: both bugs closed, one advisory loop owed

**Cold start:** read this file, then `README_where_we_are.md` (plain prose), then
`NOTES_chrome_lock_gate.md` (the missteps are the point). The full fix records live in the bug files
themselves: `/bugs_closed/069_HANDOFF_2026-07-24_site_components_writers_ignore_chrome_locks.md` and
`/bugs_closed/088_HANDOFF_2026-07-26_snapshot_revert_destroys_component_locks.md`.

## State: DONE. Nothing is blocked and nothing is inert.

| | what | where | proven |
|---|---|---|---|
| **069** | chrome (`site_components`) writers honour human locks | `05bcb3586` + `d9e7ef7cb`, live **v1.0.1170**, re-verified **v1.0.1171** | induced fault on the deployed binary |
| **088** | a snapshot revert no longer destroys the lock columns | migration **219**, applied + `--record-only` recorded | induced fault against the live functions |

Both moved to `/bugs_closed/`. `016b` §10 rows updated for both. A `WRONG_CALLS.md` row filed.

## The single owed action (advisory, does not reopen anything)

Council **round 2** was not submitted. Round 1 (`75dff4cd-e822-4b88-bd98-d989ef32bc90`) came back
REVISE, but `decided_by` was *"unreadable reviewer(s): review_editquality.result"* — a harness fault:
**10 of 12 seats approved, no veto**. Every objection is answered verbatim in the "Council" section
of the closed 069 file, and the only one that earned code (debug_historian: the hand-copied predicate
had nothing pinning it to the shared helper) is fixed in `d9e7ef7cb`.

To close the loop: rebuild the submission JSON (the scratchpad copy was cleared), paste those answers
in as the added rationale, add the lockstep test as an eighth edit, and fire
`RESUBMIT_CORR=75dff4cd-e822-4b88-bd98-d989ef32bc90 ./…/097_TRIGGER_council_review_v1.sh <file>`.
**A trailer is earned by APPROVED only**, so the commits correctly carry none and `098` will list
them as un-reviewed by design.

## Residuals, deliberately named rather than left silent

1. **Other chrome detectors have no lock filter** — `check_broken_nav_links`, `check_generic_theme`,
   `check_phantom_internal_links`, `check_integrity`, and the sibling sub-checks inside
   `check_component_standards`. A locked stale header can still raise findings its fixer now
   declines: the `bugs_open/077` shape, bounded by the two-strike rule. Only
   `checkUnlinkedSiteComponents` (this change) and `check_unverified_claims` filter today. The clean
   fix is **one exported canonical predicate**, which first needs the import direction resolved —
   `actions` imports `discovery_checks`, so the helper cannot be imported back and would have to move
   to a third package (six call sites in concurrently-edited files).
2. **`site_plan_directives` and `assets`** carry the same four lock columns from `115_locks.sql`.
   Nobody has audited whether *their* writers honour them. That is the obvious next instance of this
   class.
3. **The `high`-severity chrome item branch was never driven live** — it keys on the literal slot
   names `header`/`footer`, and the probe used scratch names on purpose so no real chrome could be
   touched. Unit-tested and read, not exercised. Same for `link_site_components` and the four
   `fix_component_template` paths: they rest on the shared predicate plus their sqlmock cases.

## Landmines this workstream paid for — read before touching any of it

- **The gate only bites when a caller passes `force_rerender: true`** (4 of the 6 live agents). The
  pre-check sits BELOW the `!force` idempotence exit on purpose: above it, every ordinary build of a
  site with a locked slot would file a `lock_blocked_change` item whose text claims a writer "wanted
  to change this", for a call that was never going to write. **So an unforced probe passes
  vacuously.**
- **"Queued, not lost" needs a second question: has the consumer RESTARTED since I published?** My
  first probe sat 20 minutes behind a 47-minute council run; another session rolled the chassis; the
  message vanished with no `orchestration_states` row, ever. `dispatch-queue-depth.sh` gave correct
  advice ("do NOT re-fire") and the message still died, because it cannot see a later roll.
- The **~300s post-roll drop window is about the spawn path**: a generic-orchestrate publish 253s
  after pod start ran in under five seconds.
- **Never cite `HEAD~1` in a durable artefact.** It became my own commit mid-task and nearly shipped
  post-fix code as evidence of the pre-fix defect, in a submission whose reviewers cannot open files.
  Resolve the SHA once (`git rev-parse <mycommit>^`) and cite the literal SHA.
- **A tag that exists is not a tag that contains your change.** v1.0.1169 was built from my commit's
  parent. `git log -1 -- makefile` and `docker images` first.
- **Read `pg_get_functiondef`, not `docs/agent_docs/sql_for_tables/*.sql`** — the copies have drifted,
  and a subagent audit got the snapshot capture set wrong by trusting them.
- `run-migrations.sh --apply` runs **every** pending file (8 other threads' were waiting). Apply
  yours with `psql -f`, then `--record-only` with a note.
- A DB-side fix can be proven inside one transaction ending in **ROLLBACK**: real deployed functions,
  real schema, no fixture ever committed, nothing to clean up. Add a **control** (an old row lacking
  the new key) or the test cannot discriminate.
- A lockstep test nobody has watched fail is a claim, not a guard —
  `TestDiscoveryChromeLockFilterMatchesSharedPredicate` was falsified on purpose (flip `'timed'` →
  `'review'` ⇒ fail; restore ⇒ pass) before being trusted.

## Verification recipes, if you need to re-run either proof

Both are written out step by step in the closed bug files, and the commands (with their gotchas
attached) are in `RUNBOOK_chrome_lock_gate.md`, which also carries the **pre-219 function definitions
verbatim** as the rollback artefact for migration 219.
