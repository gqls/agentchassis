# HANDOFF 2026-08-17 — the loop-engine work (bugs 289 / 294 / WFA-015 / WFA-016)

**This supersedes `HANDOFF_2026-08-16_continue_here.md` for everything loop-engine.** That file
is still the cold-start for the *tool-audit-by-instance* lane (bug 281, CLOSED); read it for
that, and read this for the loop engine, which grew out of its one open thread.

**Read in this order:** this file → `bugs_open/289` (the case, with three addenda) →
`bugs_open/294` (why the wreckage was immortal) → `NOTES_tool_audit_by_instance.md` (evidence,
newest at bottom) → `README_where_we_are.md` (the owner's plain-prose log) → register
**WFA-015** and **WFA-016** in `docs026_concept_register/register/workflow-authoring.md`.

---

## The one thing to know first

**Everything below is COMMITTED and NONE of it is LIVE.** All three code changes are Go, so they
are inert until the chassis image is rebuilt and rolled. Releases are whole-fleet and the owner
runs them. Do not report any of this as fixed in production.

## State, all `[MEASURED 2026-08-17]`

| item | state |
|---|---|
| `bugs_open/289` — loop `collected_data` doubles per lap (2^N) | **OPEN.** Fix `509e01e6a`, inert until the roll. Residual (5) closed by the tripwire; residual (4) still open |
| `bugs_open/294` — a run stalled in `RUNNING` is unreachable by every recovery path | **OPEN, UNOWNED, NOT APPLIED.** Filed with the exact SQL and a verify that has a negative control |
| Corpse sweep | **DONE.** 49 rows → `FAILED`; 0 `RUNNING` fleet-wide, 98 Kafka topics released. Ids in the scratchpad `289_corpse_rows_before.txt`; rows were failed, not deleted |
| `WFA-015` `loop_iteration_terminal` | built + tested, **not live** |
| `WFA-016` `collected_data` size tripwire (WARN 8 MiB / ERROR 24 MiB) | built + tested, **not live** |
| Council | **BLOCKED.** `Council-Submitted: 7a3c4fb7-…` on `509e01e6a` never rendered a verdict; the tripwire commit `cf970b009` has no trailer at all |

**Commits, in order:** `820230756` `a6312cb21` `969cea2ae` `03cfab45a` (diagnosis, 08-16) ·
`509e01e6a` (the fix + WFA-015) · `ab2b4bdd3` (docs) · `c2f66d9ff` (the quota finding) ·
`d0e104057` (sweep + bug 294) · `839312eb0` (the 274 correction) · `7d832ebc8` (2 landmines) ·
`cf970b009` (the tripwire + WFA-016).

## The blocker that is not ours to clear

**The Anthropic account quota is exhausted — 2026-08-17 11:08:03Z was the last successful LLM
call fleet-wide, and the API says access returns 2026-09-01 00:00 UTC.** Failures span two agent
types and two models, so it is account-level; over 24 h the fleet ran 478 Anthropic calls against
2 on `mistral-small3.1`, so there is no meaningful fallback. **Every LLM-driven agent is stopped**
— auditors, writers, the diagnosis loop, the fix loop, the council gate, the landmine verifier.
Owner decision: raise it or wait. Nothing in this lane can proceed past it.

## What is owed, in the order it becomes possible

1. **After the next roll — verify 289 at the artefact, WITH the demand control.**
   ```sql
   SELECT key, length(value::text) FROM orchestration_states, jsonb_each(collected_data)
   WHERE orchestration_id='<a new tool-auditor run>' AND key LIKE '%\_done' ORDER BY key;
   ```
   **The disconfirming result is a ratio near 2.0 between consecutive laps.** It proves nothing
   unless a multi-lap loop actually ran, so check `loop_metadata.total_iterations > 1` on the run
   you read. The durable pass is a `tool-auditor` orchestration reaching `COMPLETED` — the
   population held exactly 1 in 63 before this.
2. **After the next roll — verify WFA-016 fired, and mind which way the zero cuts.**
   `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=100000 | grep -c 'collected_data is unusually large'`.
   Once 289's fix is ALSO live the expected count falls toward zero, so **a zero means "no
   oversized state persisted", not "the tripwire is broken"**. To prove the instrument itself,
   check `build-dispatch-loop`, which was still legitimately reaching 14 MB when this was written.
3. **When the quota returns — submit both code commits to the council.** `509e01e6a` carries
   `Council-Submitted: 7a3c4fb7-…` whose round died on the quota, not on its content (it reached
   `review_constitution`); `cf970b009` was never submitted. **Never write `Council-Reviewed:` on a
   verdict you have not read.**
4. **When the quota returns — re-arm four landmine entries.** All four are synced to `doc_notes`
   but their verifier runs failed on the quota:
   `./scripts/trigger-landmine-verifier.sh 'LANDMINES.md#<slug>'` per entry.
5. **`bugs_open/294`, whenever the owner rules.** Live config, immediate, no roll to gate it, on a
   fleet-wide reaper — which is why it was filed rather than applied. **Re-run its age census
   immediately before applying**: the census (0 `RUNNING` rows under 4 h anywhere) is what
   licenses the 4 h threshold, and if that has changed the proposed arm is the wrong fix.
6. **289 residual (4)** — `LoopCompleteAction` still lets a step lacking its own
   `total_iterations` inherit the whole loop's. Latent now that the early return is in, and the
   fallback is deliberate backward-compat for pre-expansion plans, so weigh that before "fixing" it.

## Traps this lane paid for — do not re-walk them

- **The `090` diagnosis loop CANNOT diagnose an oversized-state bug.** Two runs (`815322b9`,
  `12ffad7c`) returned REFUTED on **false absences** — `12ffad7c` asserted "no `_done` key appears
  at all" about a row holding ten. The bundle truncates a 20 MB `collected_data` and the diagnoser
  reads truncation as absence. **A REFUTED whose citation is an ABSENCE, on a target with large
  rows, is bundle starvation** — re-ask it as a `count(*)`, never a dump. In LANDMINES.
- **Never `SELECT collected_data`, `jsonb_pretty(...)` or an unfiltered `jsonb_each` on these
  rows** — a single row reaches 29 MB and the query times out at 120 s, which reads as a busy
  cluster rather than an enormous row. Size first, then filter with `key LIKE`.
- **Match a loop's terminal substep on its ACTION, never its name.** It is called `done` in only
  9 of the 15 live loops; the rest are `complete_page` (×4), `complete_dispatch`, `task_complete`.
- **Grep BOTH bug directories, not just the source.** I re-derived a fault from the code, called
  it "worth a separate file", and it was `bugs_closed/274` — fixed two days earlier. Grepping the
  code for the mechanism is not grepping the queue (WRONG_CALLS 2026-08-17).
- **Date-bound a symptom before calling it current.** `agent_error_log` retains ~30 days, so rows
  spanning 08-03→08-15 read exactly like a live fault. `GROUP BY occurred_at::date` shows the cliff.
- **A relayed figure is not a measurement.** The 281 handoff's "a Sonnet audit every ~40 min until
  `failed`" was one probe item generalised; measured, attempts **cap at 3** at 43–147 min gaps.
- **Direct-call tests cannot prove a guard is WIRED.** Unwiring the tripwire's single call site
  left every direct test passing. Drive the real path (sqlmock) and mutate to prove it.
