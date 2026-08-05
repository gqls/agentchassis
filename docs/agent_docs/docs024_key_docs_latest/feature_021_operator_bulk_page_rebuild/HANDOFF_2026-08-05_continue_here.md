# HANDOFF 2026-08-05 — continue here (`features_open/021`, operator bulk page rebuild)

**Read `SUMMARY_2026-08-05_operator_bulk_page_rebuild.md` first** for the
five-part overview, then this file for the concrete next steps. `NOTES` has
the full trail including two mistakes caught by testing; `PLAN` has the
design and the correction to the original feature file's own reasoning.

## State in one line

The operator entry-point script (`scripts/rebuild_pages.sh`) is built and
its safe (`DRY_RUN=1`) path is tested clean. **No real dispatch has been
fired yet.** That is the next step, not a loose end — it needs a deliberate
target choice, explained below.

## What's proven vs. what isn't

**Proven (2026-08-05):**
- The mechanism (`maintenance_queue` → `maintenance-triage` →
  `page-rebuild`) is live, correctly configured, and undriven — confirmed by
  reading the full workflow JSON from `agent_definitions`, not just the
  feature file's summary of it.
- `bugs_closed/070`'s reaper fix is genuinely live (`pre_query` keys on
  `updated_at`, checked directly).
- The reaper cannot reach this path at all (different table) — a correction
  to the original feature file, recorded in `PLAN`.
- The script's `DRY_RUN=1` path: zero DB writes, zero dispatches, an honest
  local report. Tested against `gaswholesalers.com`, re-confirmed via a
  follow-up count query.

**Not proven — this is the actual next step:**
- A real (`DRY_RUN=0`) dispatch, end to end: does a named page actually
  reach `page-rebuild`, get rewritten, redeploy, and show the requested
  change on the live site?

## Immediate next step — choose a target deliberately, then fire for real

1. **Do not use `gaswholesalers.com`** for the first real test. It carries 7
   pre-existing `needs_rebuild` pages (checked live 2026-08-05) whose history
   this session doesn't know, and a real dispatch sweeps all of them in
   alongside whatever you name. Pick a site with either zero pre-existing
   `needs_rebuild` pages, or where you specifically know and accept what else
   would ride along:
   ```sql
   SELECT s.domain, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
   WHERE p.build_status='needs_rebuild' GROUP BY 1 ORDER BY 2 DESC;
   ```
2. **Pick one real, small, low-stakes page rebuild** — ideally something you
   already wanted done, so a successful test also does useful work. Run the
   script with `DRY_RUN=1` first (default) and read its report, including the
   sweep-in warning for your chosen site.
3. **Fire for real**: `DRY_RUN=0 ./scripts/rebuild_pages.sh <domain>
   <page(s)> "<real reason>"`. Save the printed `CORRELATION_ID` and
   `task_id`.
4. **Verify end to end**, not just "it dispatched cleanly" — `kcat -P` can
   silently drop a message (this repo's own well-documented landmine, hit
   twice already in the `bugs_open/178` lane on 2026-08-04). Use
   `RUNBOOK_operator_bulk_page_rebuild.md`'s verification queries: confirm an
   `orchestration_states` row appears within a few minutes, watch it run
   (page-rebuild's own step timeout is 5400s, so this can take a while),
   confirm `build_status` actually changed, and confirm the deployed page
   reflects the requested change — not just that the pipeline reported
   success (this repo's own recurring lesson: `complete` is not proof the
   work happened; check the artefact).
5. **Write up the result** in `NOTES` regardless of outcome — a failure here
   is exactly as valuable to record as a success, per this repo's standing
   practice, and it's the FIRST time any of these downstream mechanisms
   (`page-rebuild`, `rebuild_loop`'s sequencing, `mark_maintenance_complete`)
   will have run for real via this path.

## After a clean first real run

- Decide whether to register this mechanism in the concept register
  (`docs026_concept_register/register/`) — the bar per `CLAUDE.md` is
  "another workstream could call this and would not know it exists," which
  is exactly true here once it's proven live.
- Update `features_open/021`'s own status line (currently "specified, not
  built... unblocked") to reflect what actually shipped.
- Revisit the two explicitly-deferred questions in `PLAN` (intent wiring,
  dedicated Kafka topic) only if real usage makes either one matter — don't
  build ahead of a demonstrated need.

## Landmines specific to this lane

- **`psql -t -A` does not suppress a non-SELECT command tag** — wrap any
  `INSERT`/`UPDATE ... RETURNING` you capture in a CTE + `SELECT`, and assert
  the shape of what you captured before trusting it. Already fixed in the
  script; a reminder for anyone extending it.
- **The existing `check_dry_run` branch on `maintenance-triage` previews the
  AUTOMATED scanner's findings, not an operator-supplied page list.** Do not
  reach for `input_data.dry_run=true` expecting it to preview a specific
  request — it can't; `prepare_rebuild_dispatches` (the only step that reads
  a `maintenance_queue` row) is skipped entirely on that branch.
- **`page-rebuild` sweeps in every `needs_rebuild` page on the target site**,
  not only the ones a given `maintenance_queue` row named. Always run the
  pre-flight sweep-in check before a real dispatch.
- **A council/diagnosis trigger's clean-looking exit is not proof of
  delivery** — carried forward from the `bugfix_154`/`178` lane, applies
  identically to this script's own kcat dispatch.
