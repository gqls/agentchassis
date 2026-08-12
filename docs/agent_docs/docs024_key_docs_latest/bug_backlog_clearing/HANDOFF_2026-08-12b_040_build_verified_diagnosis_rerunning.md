# HANDOFF 2026-08-12b — 040-kafka-dial: fix proven live, a follow-up diagnosis run is in flight

Written to hand off from a very long single session (picked up bug 040, found
and fixed the `refused` burst's `getController` mechanism, got council
approval, then a fresh chassis roll landed mid-session and killed a follow-up
diagnosis run). Continues `HANDOFF_2026-08-12_040_council_approved_next_pickup.md`
in this same series — read that one first for the full arc; this file is the
narrow "what changed since, and what's still running" continuation.

## State, precisely

1. **The fix is committed, council-APPROVED, and now PROVEN LIVE.**
   `platform/kafka/topic_manager.go`'s `getController` rejects an empty
   `controller.Host` instead of building a dial target that resolves to the
   pod's own loopback (`bugs_open/040` §11.4). Commit `e1f960ac2`. Council
   verdict APPROVED, correlation `af5f74bc-5e6c-4a6c-a3fc-7ac27eab4b6f` (2
   advisory objections, none high-severity — full JSON in the workstream
   NOTES). **Live on `v1.0.1291`**, proven via OCI revision label
   (`da5a7eb8f`) + `git merge-base --is-ancestor` + a pod-grep with controls
   (new log line and error text both present; nonsense negative control
   absent). §11.6 of the bug file has the exact commands.

2. **A follow-up diagnosis run is IN FLIGHT, not completed.** Question: does
   the same unvalidated-Host-from-kafka-go-response pattern this fix guards
   against exist anywhere else in `platform/kafka` or its callers — this is
   a direct answer to two of the council's advisory objections (`guardian`:
   name `getController`'s callers exactly; `prior_art_librarian`: reuse vs.
   duplicate the existing `spawn_actions.go` `:9092` filter). A first attempt
   at this (correlation `e91d71d6-058a-4902-a852-c6d54bc7411c`) was cycling
   normally through several verdict rounds, then got orphaned by the
   `v1.0.1291` chassis roll — stuck at `load_runtime` with zero
   `last_activity` movement for 100+ minutes, its work item now
   `status='failed'` (the documented 40-minute claim-timeout landing
   terminal, `max_attempts=1`). **Re-fired** (`FORCE=1`, since the coverage
   probe's terminal-status list doesn't include `failed`) as correlation
   **`58a0390c-33ec-4580-9697-3320b280475d`** — left running, not polled to
   completion. Confirmed before re-firing that no new commits had landed on
   `platform/kafka/` or `spawn_actions.go` since the push its clone reads
   from, so its view of the code is current.

## What to do next

1. **Check the run first, before anything else:**
   ```sql
   SELECT status, current_step, EXTRACT(EPOCH FROM (NOW()-last_activity))::int AS since_s,
          substring(COALESCE(error,''),1,200) AS err
   FROM orchestration_states WHERE correlation_id = '58a0390c-33ec-4580-9697-3320b280475d'::uuid
   ORDER BY created_at DESC LIMIT 1;
   ```
   - `COMPLETED` → read the verdict (below), write it into `bugs_open/040`
     §11 (new subsection) and the workstream NOTES, close the loop on the
     two council objections it was meant to answer.
   - `EXECUTING_STEP` with `since_s` climbing normally (tens of seconds,
     cycling through `route`/`load_runtime`/`assemble_bundle`/`verdict`) →
     still genuinely running, keep waiting.
   - `EXECUTING_STEP` with `since_s` in the thousands and not moving →
     wedged again (another roll, or something else). Do not re-poll
     forever — check `kubectl -n ai-persona-system get pods -l
     app=agent-chassis` for a recent restart to confirm, then treat as dead
     and decide whether a third attempt is worth it (it may not be — see
     "if it dies again" below).
   - `FAILED` → read `error`, decide whether to re-fire again.

   Fetch the verdict once `COMPLETED` (write to a file first — piping a
   large `kubectl exec -i` payload straight into `python3 -c` has truncated
   mid-stream at least once this session):
   ```bash
   kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c "
     SELECT collected_data->'verdict' FROM orchestration_states
     WHERE correlation_id = '58a0390c-33ec-4580-9697-3320b280475d'::uuid
     ORDER BY created_at DESC LIMIT 1;" > /tmp/verdict2.json
   python3 -m json.tool /tmp/verdict2.json
   ```

2. **If it dies again to a roll:** this is now the second time in one
   session a chassis roll has orphaned an in-flight diagnosis run. That
   matches the standing landmine ("no orchestration dispatch within ~300s of
   a chassis pod (re)start — the spawn is silently dropped", and "a roll
   KILLS an in-flight council") but this is the first time this session has
   seen a roll kill a run that was already well underway (not just at
   dispatch). If you see it happen a THIRD time, that's worth a `LANDMINES.md`
   entry of its own — "090/council runs have no resilience to a mid-run
   chassis roll, and there is no visible retry" — with this correlation pair
   as the worked example. Check `docs/agent_docs/docs024_key_docs_latest/LANDMINES.md`
   first in case someone already added it.

3. **Whether or not the audit run lands cleanly, the bug itself (`040-kafka-dial`)
   stays OPEN.** The shipped fix hardens one candidate mechanism and is
   proven live; it does not claim to be the confirmed cause of the measured
   71,832-event `refused` burst. Nothing about this run changes that — it is
   scoped to the narrower "are there other instances of the same code
   pattern" question, not to root-causing the burst itself.

4. **Standing four before touching anything else in the backlog:**
   `scripts/who-owns.py`, `git log` on the bug's own named file, live
   `.jsonl` transcript grep, `site_work_items` queue check — as in every
   prior handoff in this series. The `029`–`255` range was fully saturated
   as of the last handoff in this series; that will only be more true now.

## Files touched this session (all committed, pathspec per task)

- `platform/kafka/topic_manager.go`, `platform/kafka/topic_manager_controller_test.go`
  — the fix (`e1f960ac2`).
- `bugs_open/040_HANDOFF_2026-07-20_kafka_dial_timeouts_fleetwide_intermittent.md`
  — §11 (the whole arc: residual found, diagnosed, fixed), §11.5 (MetadataTopics
  cross-check), §11.6 (build proof + the killed/re-fired audit run).
- `bugs_open/240_HANDOFF_2026-08-10_kafka_scheduler_oom_every_client_fetches_metadata_for_all_25k_topics.md`
  — a contribution note (timing cross-check), not a competing fix; that bug
  stays owned by the `bugfix_209_deploy_purpose_keyed_source` lane.
- `docs/agent_docs/docs024_key_docs_latest/bugfix_040_kafka_dial/` — all five
  standing docs current (PLAN §8, RUNBOOK §10–12, NOTES through this run,
  README, plus a new `SUMMARY_2026-08-12`).
- `docs/agent_docs/docs024_key_docs_latest/WRONG_CALLS.md` — the
  `count`-vs-`sum` PromQL misstep.
- `docs/agent_docs/docs024_key_docs_latest/bug_backlog_clearing/` — this file
  and `HANDOFF_2026-08-12_040_council_approved_next_pickup.md`.

No bugs_closed move — 040 stays open per the owner's standing direction
(closure evidence lives inline in `bugs_open/`, per CLAUDE.md's 2026-08-06
override of the fixed-AND-live bar).
