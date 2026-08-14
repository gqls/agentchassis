# HANDOFF — 2026-08-14c. Candidate 2 **APPROVED** and **live on `v1.0.1299`, all 17 pods**. Nothing is owed. COLD-START HERE.

Supersedes `HANDOFF_2026-08-14b_live_on_1298_one_check_owed_verdict_pending.md` — its
title is now wrong in both halves: the owed check is closed, the verdict is read, and the
tag it certifies has been rolled past. Its §5 (what else is open on 231) and §7 (traps)
still hold. Full evidence: `bugs_open/231` §POST-ROLL, which now carries the verdict and
a disposition for every objection. Missteps: `NOTES_209…`. Owner prose:
`README_where_we_are.md`. Architecture question: **RFC_028**.

## 1. State in one table

| thing | state |
|---|---|
| Seam (Strategy 6) | **LIVE**, `v1.0.1299`, **all 17 pods running this binary**, stamp `6f8efa158`, ancestry-proven both sides |
| Behavioural proof that Strategy 6 FIRES | **DONE** — 6 lines, both replicas, 13:56–14:30Z, with Strategy 0 as liveness control |
| Council | **APPROVED** 17:29:51Z — 13 seats, 4 abstained, no truncation, 5 advisory objections, none HIGH |
| Trailer | `Council-Reviewed: 41a01378-1211-4987-966d-f8b6e2fddce1` — **written, verdict read in full** |
| `bug_historian`'s blast-radius ask | **CLOSED by measurement** — 0 of the 21 touch a render/rebuild write path |
| `architecture` + `guardian` objections | **OPEN BY REQUEST**, routed to `RFC_028` — a human's call, not a session's |
| `--report` mode | built, has written a real row, **still UNDRIVEN** (no CronJob) |

Commits: `d3edb5b89` seam · `14e4333f7` revise round · `bc39e7bf5` RFC_028 · `220dff6ad`
round-3 record · `b9b937ed0` restore + refire · `77d738a97` landmine · this one, carrying
the trailer.

## 2. What the round-3 story is actually about, because it is not the code

Round 3 died at `complete_invalid` on `review_editquality` at 16:38Z and **the submission
was fine** — it had already passed `persist_submission`. The Anthropic account hit its
usage cap (`bugs_open/243-anthropic-cap`, third exhaustion in 15 days). Refired unchanged
at 17:10:27Z after the owner restored service; approved 19 minutes later.

**If you hit this: the table that answers "is it back" is `llm_call_log`, and the two
obvious alternatives actively mislead.** `orchestration_states` showed **63 COMPLETED / 0
FAILED** straight through the outage, because those completions are `build-dispatch-loop`,
`endpoint-health-checker`, `build-pipeline-trigger`, `page-rerender` — plumbing that never
calls a model. `agent_error_log` goes quiet both when calls succeed and when nothing calls.
Only a `success` column on the provider call itself discriminates, and only with the
failing minutes shown beside the passing one:

```sql
SELECT date_trunc('minute',created_at) AS minute, provider,
       count(*) FILTER (WHERE success) AS ok, count(*) FILTER (WHERE NOT success) AS failed
FROM llm_call_log WHERE created_at > now() - interval '75 minutes' GROUP BY 1,2 ORDER BY 1 DESC;
```
Measured: 15 failures 16:05→16:42Z, zero successes; first success **17:08:40Z**. Written
up into the existing usage-limit entry in `LANDMINES.md` (appended, not a new entry).

## 3. Two objections were about THIS LANE'S EVIDENCE — one was right

Recorded in full in `bugs_open/231`; the transferable halves:

- **`/proc/1/exe` was cited correctly and the seat was wrong**, but only by an arm: the
  landmine it quoted condemns probing with **your own commit's sha** (the binary carries
  exactly one commit — the build point). This lane probed the build point itself and ran
  the mandated `git merge-base --is-ancestor` both ways. **The seat read the entry's
  title, not its arms** — and the entry's title had itself been refined to "TIME-LIMITED,
  INOPERATIVE is too strong" the day before.
- **"Both replicas verified" was a 2-pod sample of 17 and the seat was right.** Answering
  it produced a better fact than the one objected to: all 17 pods running this binary are
  on one tag, and the seam is an ancestor of its stamp. **Do not verify a chassis change
  by the two `-l app=agent-chassis` replicas** — enumerate by image:
  ```bash
  kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{" "}{.spec.containers[0].image}{" "}{.status.startTime}{"\n"}{end}' | grep agent-chassis:
  ```
  The ephemeral `agent-build-dispatch-loop` pods are the trick for reading the stamp:
  they respawn every ~90 seconds, so one is always young enough that its log still reaches
  back to startup. Prove that with `kubectl logs <pod> | head -1` **before** trusting the
  `build provenance` line — the discriminator is the landmine's own.

## 4. What remains — none of it blocking, none of it this lane's to decide

1. **`RFC_028`** — the `architecture` seat recorded `needs_rfc` with **8 deflections**, and
   `guardian` asked that its own objection stay open rather than be absorbed into the
   approve. Both are honoured in `bugs_open/231`. The owner's three questions are unchanged.
   **Guardian's LOW objection is the one worth carrying into it:** the Strategy-3 bridge
   precedence rests on a point-in-time census, so a future `agent_definitions` row carrying
   both a deprecated alias and its canonical key on one defaulted field gets the new
   precedence silently, with a unit test as the only backstop.
2. **The `--report` CronJob** — unchanged from `14b` §4: build/push the image FIRST, then
   apply the overlay, or this fleet reports `ImagePullBackOff` as a Job still RUNNING.
   Copy `removed-config-keys-check` line-for-line; schedule 06:30.
3. **A finding worth someone's time:** `live_override` counts what the RESOLVER would
   honour, and it **over-counts behaviour change** — the three render-adjacent entries of
   the 21 read `params.StepConfig.Config` directly (`GetIntField(config,…)`), so they were
   never dead and Strategy 6 does not touch them. Whoever next quotes the 99/21 figures
   should quote that limit with them.
4. The 96 `dotted_conditional` entries and CTS-059's open review question — unchanged,
   `14b` §5.
