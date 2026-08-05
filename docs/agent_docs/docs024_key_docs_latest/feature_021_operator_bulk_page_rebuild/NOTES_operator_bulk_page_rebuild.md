# NOTES — operator bulk page rebuild (`features_open/021`)

## 2026-08-05 — picked up, scoped, built, tested; one live-fire test still owed

Picked this up after a `features_open/` survey (done while `bugs_open/178`'s
workstream ran dry) found #021's blocker (`bugs_open/070`) had closed
2026-07-27 with nobody following up. User confirmed: go ahead.

**Re-verified every claim in the feature file against the LIVE system before
building anything** (11 days since the filing, ~1500 commits/wk on this repo —
per this repo's own memory practice, an unrefreshed prior-art claim is not
evidence):
- `maintenance_queue`: still exactly 2 rows, both `complete`, max `created_at`
  2026-02-18 — unchanged, confirms nobody has used this since the filing.
- `maintenance-triage` agent: still `is_active=true`.
- `scheduled_tasks` targeting it: still 0 rows.
- `stale-work-item-reaper`'s `pre_query`: **now keys on `updated_at`**, not
  `created_at` — confirms `bugs_closed/070`'s fix is genuinely live, not just
  claimed in its own file.
- `build-pipeline-trigger`: still 120s, enabled.

**Read the actual mechanism in full before designing anything** —
`maintenance_actions.go`, both agents' full workflow JSON from
`agent_definitions`. Found the correction recorded in `PLAN`: this path never
touches `site_work_items`, only `pages.build_status` and `maintenance_queue`
directly. `stale-work-item-reaper` cannot reach it. `bugs_open/070` is not
actually a prerequisite for this path (it is for the operator's ORIGINAL
by-hand workaround, which this feature makes obsolete). This is a genuine
correction to the feature file's own stated reasoning, not just new
information — recorded in `PLAN`, not silently.

Also read `recompose_pages` (`features_open/012`) to check whether its intent
vocabulary already applies here. **It does not** — it lives in
`v3_site_actions.go`'s site-PLAN validation pipeline, a different mechanism
entirely. `page-rebuild`'s own `build_pages_loop` has no re-render-only branch;
it always calls `plan_sections` + a content-writer step. So "intent" (point 4
of the feature file) is currently un-wireable without new Go code in
`page-rebuild` itself — recorded as explicitly deferred in `PLAN`, not silently
dropped.

**Built `scripts/rebuild_pages.sh`** — INSERT a `maintenance_queue` row +
direct kcat dispatch to `maintenance-triage`, modelled on the existing
`090_TRIGGER_needs_diagnosis_v1.sh` conventions (envelope shape, correlation
handling, "don't trust a clean dispatch" caveat).

**Tested it — twice, and it was wrong both times before it was right:**

1. **First run (DRY_RUN=1, but the code path inserted a row regardless):**
   `TASK_ID` captured as `6f454f1c-...\nINSERT 0 1` — the INSERT's own command
   tag leaked into the captured value, because `psql -t` suppresses a SELECT's
   header/footer but NOT a non-SELECT's completion tag. **This is the exact
   landmine `090_TRIGGER_needs_diagnosis_v1.sh` already documents** (its own
   header explains wrapping an UPDATE in a CTE + SELECT for precisely this
   reason) — I wrote a bare `INSERT ... RETURNING` anyway and only caught it by
   actually running the script, not by reading 090's script first as closely
   as I should have. Fixed: wrapped in `WITH ins AS (INSERT ... RETURNING id)
   SELECT id::text FROM ins`, plus a UUID-shape assertion on the captured value
   before trusting it (same discipline 090 uses for its claim check). Cleaned
   up the stray test row by hand (`DELETE FROM maintenance_queue WHERE
   id='6f454f1c-...'`) since nothing else would have.

2. **Second issue, found by the same test run, not by re-reading code:** the
   dispatch used `dry_run:true` in `input_data`, matching what I'd assumed the
   safe/preview path meant. But `maintenance-triage`'s `check_dry_run` step's
   dry branch (`complete_dry_run`) skips straight past
   `prepare_rebuild_dispatches` — the ONLY step that ever reads a
   `maintenance_queue` row. So the "dry run" I dispatched never looked at my
   inserted row at all; it only ran `scan_and_queue`'s own independent
   stale/missing/orphan scan. A `DRY_RUN=1` invocation of my script therefore
   inserted a real row that would sit unpreviewed and unclaimed until some
   LATER real dispatch happened to claim it — a genuine design gap, not just a
   labelling problem. **Corrected: `DRY_RUN=1` now does no DB write and no
   dispatch at all — it prints a local report of what WOULD happen and exits.**
   Only `DRY_RUN=0` touches the database or Kafka.

**Re-tested after both fixes**: `DRY_RUN=1` against `gaswholesalers.com`
(chosen because it already carries 7 pre-existing `needs_rebuild` pages, a good
case to prove the sweep-in warning fires) — output showed the correct 7-page
warning, then the local dry-run report, and a follow-up query confirmed **0
rows** in `maintenance_queue` for that site afterward. Clean.

**Did not fire a real (`DRY_RUN=0`) dispatch this session.** A real dispatch on
`gaswholesalers.com` right now would sweep in those same 7 pre-existing
`needs_rebuild` pages, whose history I do not know, into a 90-minute,
content-regenerating, real-deploy run — the kind of hard-to-reverse,
production-affecting action this repo's own operating norms say to slow down
for, not a unilateral call for a first proof-of-mechanism run. Left as the
explicit next step; see `HANDOFF`.

**Not yet done, not attempted, correctly deferred** (see `PLAN` for the full
reasoning): intent (recompose vs re-render) wiring; a dedicated Kafka topic
for this dispatch type (page-rebuild runs are long, same shape of concern that
got council-gate its own topic — worth measuring after real usage, not
designing for zero data points); the first real live-fire test itself.
