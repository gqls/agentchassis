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

~~**Everything below is COMMITTED and NONE of it is LIVE.**~~ **SUPERSEDED 2026-08-17 18:2xZ — IT
SHIPPED.** Build `v1.0.1307` (pods started 17:05Z, digest `sha256:8339bdbd…`, OCI revision label
`a6d1c53c0`) carries **`509e01e6a` and `cf970b009`**, confirmed by `git merge-base --is-ancestor`
against that revision — not by the tag, which had lied twice. **The doubling is GONE and measured:**
`build-dispatch-loop`'s per-lap `_done` keys were 201 kB → 6,247 kB and are now a flat **77 bytes**
(ratio 1.00; ~2.0 would have disconfirmed it), and multi-lap runs went from 2,575 kB avg / 14 MB max
to **104 kB avg / 229 kB max**. `a436d898f` (the council's `GetIntField` reuse swap) postdates the
build, is behaviourally identical, and rides the next one.

## State, all `[MEASURED 2026-08-17]`

| item | state |
|---|---|
| `bugs_open/289` — loop `collected_data` doubles per lap (2^N) | **OPEN.** Fix `509e01e6a`, inert until the roll. Residual (5) closed by the tripwire; residual (4) still open |
| `bugs_open/294` — a run stalled in `RUNNING` is unreachable by every recovery path | **OPEN, UNOWNED, NOT APPLIED.** Filed with the exact SQL and a verify that has a negative control |
| Corpse sweep | **DONE.** 49 rows → `FAILED`; 0 `RUNNING` fleet-wide, 98 Kafka topics released. Ids in the scratchpad `289_corpse_rows_before.txt`; rows were failed, not deleted |
| `WFA-015` `loop_iteration_terminal` | built + tested, **not live** |
| `WFA-016` `collected_data` size tripwire (WARN 8 MiB / ERROR 24 MiB) | built + tested, **not live** |
| Council | **APPROVED**, round 2, corr `7a3c4fb7-e8c1-4b5f-950e-7a826d5bebbe` (~16:50Z) — 4 advisory objections, none high-severity, **all answered with a change or a measurement** (see 289's council section). `Council-Reviewed:` trailer on `a436d898f`. Note `cf970b009` still carries no trailer and forward-only forbids an amend, so `098` cannot auto-credit that one |

**Commits, in order:** `820230756` `a6312cb21` `969cea2ae` `03cfab45a` (diagnosis, 08-16) ·
`509e01e6a` (the fix + WFA-015) · `ab2b4bdd3` (docs) · `c2f66d9ff` (the quota finding) ·
`d0e104057` (sweep + bug 294) · `839312eb0` (the 274 correction) · `7d832ebc8` (2 landmines) ·
`cf970b009` (the tripwire + WFA-016).

## The blocker that IS real, and the one I got wrong

**REAL — the build is not reaching the cluster, and this is the live blocker.** `[MEASURED
2026-08-17 16:30Z]` The pods look new (replicaset `5bd56bdd9b`, started 14:43Z) and carry tag
`v1.0.1305`, but their image digest is `sha256:f90a7e88…`, whose own OCI revision label is
**`6a782274b`, built 2026-08-16 21:53Z**. The locally built `v1.0.1305` is a *different image* —
`sha256:6039e19c…`, revision `89a0cbeb7`, created 2026-08-17 14:30Z — and it DOES contain both of
this lane's commits. Same tag, different content: the build worked, the delivery did not.
**252 commits in HEAD are absent from the running binary, 24 of them touching
`platform/`/`internal/`/`pkg/`** across bugs 275/283/285/287/289/291/292/293/295/299, several
already council-APPROVED. A same-tag re-release re-serves the node's cache, so restarting pods
cannot fix it — **only a new tag can**. `makefile` line 17 is now `v1.0.1306` (commit
`aa9c7b74f`); the owner runs the release. One-command proof, no exec:
`docker inspect aqls/<svc>@sha256:<imageID from kubectl> --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'`.

**NOT REAL — the "Anthropic quota exhausted until 2026-09-01" claim was mine and it was WRONG.**
The outage lasted **~3 minutes**: 4 failed calls between 11:08:37Z and 11:09:53Z, successes
resuming 11:13:02Z, and 462 successful calls in the five hours after. I measured inside the window
and took the duration from the API's 400 body. **The date in that message is not predictive** —
`bugs_open/243` records the identical text on 08-10 against a 3 h 20 m outage that ended when the
owner added credit, 21 days early, and says so explicitly. I did not grep for it. Council round 2
went out normally at 16:38Z. Do not plan around a September date.

## What is owed, in the order it becomes possible

0. ~~**THE ROLL ITSELF IS BLOCKED ON A TAG.**~~ **DONE — the release landed at `v1.0.1307`.** Kept below because the check is the reusable part: `makefile` line 17 is now
   `v1.0.1306` (`aa9c7b74f`). Until a release goes out at a NEW tag, items 1 and 2 below cannot be
   done at all — and a same-tag re-release will look like it worked. Confirm before believing any
   roll: `docker inspect aqls/agent-chassis@sha256:<imageID from kubectl> --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'`
   then `git merge-base --is-ancestor 509e01e6a <that revision>`.
1. ~~After the next roll — verify 289 at the artefact.~~ **DONE on the shared engine, PENDING on the motivating case.** Proven on `build-dispatch-loop` with real demand (see 289's "LIVE AND PROVEN" section). **Still owed: a `tool-auditor` run reaching `COMPLETED`** — none has run since the roll because its queue holds nothing runnable, so that zero is absence of demand, not evidence. One bounded probe is in flight: item `12836a25-8266-46fb-bba1-2e8635ef9cc0`, `max_attempts=1` so it costs at most ONE Sonnet audit. Read its outcome, then 289 can move to `bugs_closed/`. The original instruction, still the right check:
   ```sql
   SELECT key, length(value::text) FROM orchestration_states, jsonb_each(collected_data)
   WHERE orchestration_id='<a new tool-auditor run>' AND key LIKE '%\_done' ORDER BY key;
   ```
   **The disconfirming result is a ratio near 2.0 between consecutive laps.** It proves nothing
   unless a multi-lap loop actually ran, so check `loop_metadata.total_iterations > 1` on the run
   you read. The durable pass is a `tool-auditor` orchestration reaching `COMPLETED` — the
   population held exactly 1 in 63 before this.
2. ~~After the next roll — verify WFA-016 fired.~~ **CHECKED: 0 lines, and that is the expected reading.** Nothing persists an oversized state now, so the instrument has no demand; per its own verify-later note a zero means "no oversized state", not "broken". Exercising it now needs a fixture. Original note kept:
   `kubectl -n ai-persona-system logs -l app=agent-chassis --tail=100000 | grep -c 'collected_data is unusually large'`.
   Once 289's fix is ALSO live the expected count falls toward zero, so **a zero means "no
   oversized state persisted", not "the tripwire is broken"**. To prove the instrument itself,
   check `build-dispatch-loop`, which was still legitimately reaching 14 MB when this was written.
3. ~~When the quota returns — submit both code commits to the council.~~ **DONE 2026-08-17
   16:38Z.** Round 2 dispatched on the **original** correlation
   `7a3c4fb7-e8c1-4b5f-950e-7a826d5bebbe` (run orch `ec91ab5a-4548-4327-8f3d-9dd012005bbc`), so
   `509e01e6a`'s existing `Council-Submitted:` trailer resolves and `098` credits it automatically
   on approval. **Both changes are in that ONE round** — the loop fix and the tripwire — because
   `bugs_open/244` measures `council-gate` at 87.8% of the fleet's August LLM spend and ~790k input
   tokens per round, so a second round for four lines of logging is a poor trade. It cleared
   `review_constitution` (where round 1 died) and was at `review_tooling_provenance` at 16:44Z.
   **Still owed: READ THE VERDICT** and act on a REVISE/REJECTED — the code is already on the
   shared branch. `SELECT metadata->>'decision' FROM diagnosis_artifacts WHERE
   correlation_id='7a3c4fb7-…' AND kind='council_report';`
   **Never write `Council-Reviewed:` on a verdict you have not read.** Note `cf970b009` carries no
   trailer at all and forward-only forbids an amend, so `098` cannot auto-credit it whatever the
   verdict — the correlation is recorded in the bug file, WFA-016 and here instead.
4. ~~When the quota returns — re-arm four landmine entries.~~ **DONE.** The verifier is running
   normally and verdicts are arriving. Mine so far: the one-sha binary stamp → **STILL_VALID**; the
   same-tag rebuild → **NEEDS_HUMAN_REVIEW**, which is the verifier's `.go`-only index refusing a
   makefile/shell footprint rather than a doubt about the entry — the human half was measured by
   hand and is recorded in the entry. **Read any `NEEDS_HUMAN_REVIEW` that way: an entry footprinted
   on shell, SQL, YAML or a makefile can never come back `STILL_VALID`.**
5. **`bugs_open/294`, whenever the owner rules.** Live config, immediate, no roll to gate it, on a
   fleet-wide reaper — which is why it was filed rather than applied. (The earlier note that the
   council gate was down was **wrong** and was never the real reason; the three above are.) **Re-run its age census
   immediately before applying**: the census (0 `RUNNING` rows under 4 h anywhere) is what
   licenses the 4 h threshold, and if that has changed the proposed arm is the wrong fix.
6. **289 residual (6), from the council's architecture seat** — the `loop_iteration`-presence
   fallback in `isLoopIterationTerminal` is a permanent SECOND discriminator once every persisted
   plan carries the explicit flag. Delete it once no in-flight plan predates `509e01e6a`. The
   seat's wider point is the RFC signal and is NOT this fix: the step model overloads `action` to
   mean both "what to run" and "role in the workflow", which is why a flag was needed at all.
7. **289 residual (4)** — `LoopCompleteAction` still lets a step lacking its own
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
