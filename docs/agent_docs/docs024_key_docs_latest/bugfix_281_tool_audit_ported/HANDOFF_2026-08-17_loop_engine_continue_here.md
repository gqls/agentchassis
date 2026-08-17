# HANDOFF 2026-08-17 (evening) — loop engine: 289 CLOSED and PROVEN LIVE; what is left

**Supersedes the earlier version of this file and `HANDOFF_2026-08-16_continue_here.md` for
everything loop-engine.** The 08-16 file remains the cold-start for the *tool-audit-by-instance*
lane (bug 281, closed); read it only for that.

**Status in one line: the loop-engine defect is fixed, live on `v1.0.1307`, and proven at the
artefact on the case that motivated it. Nothing here is blocked on the owner except `bugs_open/294`.**

---

## 1. What shipped, and the proof

`bugs_closed/289` — a loop sub-workflow's terminal substep is declared `action: loop_complete`,
which is *also* the action of the loop's own end-step, so expansion injected the whole-loop
aggregator once **per iteration** and each lap swallowed every earlier lap's aggregate.
`collected_data` doubled per lap (2^N).

**Live on `v1.0.1307`** (pods started 17:05Z; image digest `sha256:8339bdbd…`, OCI revision label
`a6d1c53c0`, built 16:50Z). Verified with `git merge-base --is-ancestor`, **not** by the tag —
which had reported a "fresh build" twice while shipping nothing.

**Proof, two independent consumers:**

| | before | after |
|---|---|---|
| `build-dispatch-loop` per-lap `_done` | 201 kB → 6,247 kB (ratio ~2.1–3.1) | **77 B, flat (1.00)** |
| `build-dispatch-loop` multi-lap runs | 156 runs, 2,575 kB avg, 14 MB max | 10 runs, **104 kB avg, 229 kB max** |
| `tool-auditor` per-lap `_done` (10 laps) | 70 kB → 9,076 kB | **82 B each, all ten** |
| `tool-auditor` outcome | **1 COMPLETED in 63** | **COMPLETED on attempt 0** |

The `tool-auditor` run (`be4cf3a5-f3cc-4208-9ec8-0d17d27e421d`) carried **18 findings against a cap
of 10**, i.e. it hit the exact truncation condition that correlated with every stall, and finished.
Aggregation still works: the end-step returned `items_created = {count: 10, iterations: 10}`.

**Fleet health since the roll, with a demand control:** 0 `RUNNING` rows fleet-wide, 0 stuck >4 h,
0 runs at any `*_loop_complete` step — across a window containing 11 real multi-lap runs, so the
zeros are not silence.

**Council: APPROVED**, corr `7a3c4fb7-e8c1-4b5f-950e-7a826d5bebbe`, 4 advisory objections, all
answered with a change or a measurement (see the council section in `bugs_closed/289`).

## 2. What is still open, in the order I would take it

1. **`bugs_open/294` — OWNER CALL. The only thing genuinely waiting on a human.** A run stalled in
   `RUNNING` is unreachable by every recovery path: the reaper has arms for `AWAITING_RESPONSES`
   (30 m / 90 m) and `EXECUTING_STEP` (4 h) and **none for `RUNNING`**, while `TimeoutMonitor` keys
   entirely on `awaited_requests`, which these rows have empty. 289 was the producer; **294 is why
   the corpses were immortal, and it is untouched.** The file carries the exact SQL, the census that
   licenses its 4 h threshold, and an induced-repro verify with a negative control.
   **Why it was filed rather than applied:** live config, effective the instant it is saved, no
   build step to catch a mistake, on a fleet-wide reaper. **Re-run its age census immediately
   before applying** — that census is the licence, and if healthy work has since started living in
   `RUNNING` the proposed arm is the wrong fix.
2. **`a436d898f` needs a build.** The council's `reuse_agent` objection is applied (my bespoke
   `loopConfigInt` swapped for the existing `datahelpers.GetIntField`). It **postdates `v1.0.1307`**,
   is behaviourally identical, and rides the next roll. Nothing depends on it.
3. **Residual (6)** — the `loop_iteration`-presence fallback in `isLoopIterationTerminal` is a
   permanent *second* discriminator once every persisted plan carries the explicit flag. Delete it
   then. (Council architecture seat.) Its wider observation is the RFC signal and is **not** this
   fix: the step model overloads `action` to mean both "what to run" and "role in the workflow".
4. **Residual (4)** — `LoopCompleteAction` still lets a step lacking its own `total_iterations`
   inherit the whole loop's. Latent now, and the fallback is **deliberate** backward-compat for
   pre-expansion plans, so weigh that before "fixing" it. I recommend leaving it.
5. **The per-page `audit_review` key** (the 281 lane's item, not mine) — demonstrated live by the
   verification run: 18 findings → 10 create attempts → **0 rows inserted**, all deduped against one
   pre-existing row. `items_created.count` counts **step results, not rows written**. Anyone
   measuring this machinery's yield must read `inserted`, not the count.

## 3. Traps this lane paid for — do not re-walk them

- **A "fresh build" can ship no new code.** Two of the three reported to me had not. **Check the
  image's own label, never the tag, the pod age or the deploy event:**
  `docker inspect aqls/<svc>@sha256:<imageID from kubectl> --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'`
  then `git merge-base --is-ancestor <your commit> <that revision>`. Different digests under one tag
  = not live; the remedy is a **tag bump**, never a redeploy.
- **The `090` diagnosis loop CANNOT diagnose an oversized-state bug.** Both runs returned REFUTED on
  **false absences** — one asserted "no `_done` key appears at all" about a row holding ten. The
  bundle truncates a 20 MB `collected_data` and the diagnoser reads truncation as absence. **A
  REFUTED whose citation is an ABSENCE, on a target with large rows, is bundle starvation.**
- **Never `SELECT collected_data` / `jsonb_pretty` / unfiltered `jsonb_each`** on these rows — 29 MB
  singles time out at 120 s, and a timeout reads as a busy cluster.
- **Match a loop's terminal substep on its ACTION, never its name** — `done` in only 9 of 15 loops.
- **Grep BOTH bug directories, and grep them for the noun in your error.** I missed this twice in
  one day: `bugs_closed/274` (a fault I re-derived and called unfiled) and `bugs_open/243` (an API
  error whose message was literally that file's title). Searching the code is not searching the queue.
- **A duration cannot be read off a single moment's data**, and a vendor's error text is the worst
  place to get one. I reported a 15-day outage that lasted 3 minutes. `bugs_open/243` §third
  occurrence has the histogram query that settles it in one shot.
- **Direct-call tests cannot prove a guard is WIRED** — unwiring the tripwire's call site left all
  nine direct tests green. Drive the real path (sqlmock) and mutate.
- **When you mutate to prove a test, prove the mutant still COMPILES first** — a build-breaking
  mutant prints the same `FAIL` as a caught one. (Council `debug_historian` caught this in my
  evidence; the corrected three-step form is in 289's council section.)
- **No backticks inside `git commit -m`** — they execute. One ate a word out of `fcf436b87`.

## 4. Commits (all on `087_towards_multiple_domains`)

Diagnosis 08-16: `820230756` `a6312cb21` `969cea2ae` `03cfab45a`.
Fix + register WFA-015: `509e01e6a`. Docs: `ab2b4bdd3`. Quota claim (later corrected): `c2f66d9ff`.
Sweep + bug 294: `d0e104057`. 274 correction: `839312eb0`. Landmines: `7d832ebc8`.
Tripwire + WFA-016: `cf970b009`. Tag bump: `aa9c7b74f`. Quota correction: `30d0d421c`.
Landmine verification: `5aee74d01`. Council APPROVED + reuse swap: `a436d898f`.
Live proof: `d81969ca2`. **Close: `fcf436b87`.**

## 5. Useful queries

```sql
-- is the doubling back? a ratio near 2.0 between laps is the disconfirming result
SELECT key, length(value::text) FROM orchestration_states, jsonb_each(collected_data)
WHERE orchestration_id='<run>' AND key LIKE '%\_done' ORDER BY key;

-- 294's census: the licence for a RUNNING reaper arm. Re-run before applying it.
SELECT CASE WHEN last_activity > now()-interval '4 hours' THEN 'live' ELSE 'dead' END,
       count(*), count(DISTINCT owner_agent_type)
FROM orchestration_states WHERE status='RUNNING' GROUP BY 1;

-- fleet loop health
SELECT owner_agent_type, count(*), pg_size_pretty(max(length(collected_data::text))::bigint)
FROM orchestration_states WHERE created_at > now()-interval '6 hours' GROUP BY 1 ORDER BY 2 DESC;
```
