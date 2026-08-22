# HANDOFF — 2026-08-21 — bugs_open/343 + bugs_open/040, continue here

**Session brief was:** research both bugs, validate they are still real, plan with Fable, council-review,
commit for the next chassis build, keep the docs current, log missteps.

**State: five coherent tasks shipped, all four council correlations APPROVED, three of them LIVE.
NEITHER BUG CLOSES, and §"What is still open" says exactly why.**

> ## ⚠ OVERTAKEN 2026-08-22 — read this before acting on §3 or §4
>
> Appended by the `bugfix_029` lane, which put this file's open items to the owner and heard the
> answers first-hand. **This file's own record above is unrewritten and was correct on 08-21.**
>
> - **§3 "THE FIRST THING TO DO NEXT — verify the second roll" is DONE.** `agent-chassis` is on
>   **v1.0.1323** (both pods, 2026-08-22 08:36Z). `[VERIFIED at the binary]` `kafka produce succeeded
>   after retry` **PRESENT** — it was §1's *absent* control on v1.0.1322, so the probe is shown to
>   discriminate — plus `client_no_leader` PRESENT, `ai_persona_kafka_produce_total` PRESENT,
>   `Reply beat the park - consumer owns the continuation` PRESENT, `zzz_nonexistent_marker_qqq`
>   absent. The `build provenance` line had already scrolled at ~2h uptime, as its landmine predicts.
> - **§4 "343 — the first death is untouched": still true, and 343 is CLOSED ANYWAY.** Owner ruling,
>   an explicit override of the fixed-AND-live bar. Now at `bugs_closed/343_…`. RSH-011 stays armed
>   and broad — **do not decommission it on the grounds that the bug is closed.**
> - **§4 "040 — three things": items 1 and 3 are PARKED by the same ruling; item 2 is not.** 040
>   itself **stays OPEN**. The park is recorded inside
>   `bugs_open/040_HANDOFF_2026-07-20_kafka_dial_timeouts_fleetwide_intermittent.md` **§12 — by the
>   040 lane, which owns that file**; grep it for `PARKED` rather than for a section number, because
>   they are striking §12.6's items 1 and 3 in place and the exact shape is theirs to choose.
> - **Still live and now armed by this roll:** the `empty_host` disconfirming test (item 2's cheap
>   answer — no `090` needed any more) and the `DUPLICATE_SKIPPED` watch.

---

## 1. What is LIVE right now

`agent-chassis` **v1.0.1322**, both pods, started **2026-08-21 16:54Z**, built from `bac189921`.

Proven at the binary, not inferred from the tag — and with a control that could have failed:

| probe | result | why it matters |
|---|---|---|
| `Reply beat the park - consumer owns the continuation` | PRESENT | 343 task A |
| `Optimistic lock failure persisting loop skip` | PRESENT | 343 task B (P2) |
| `ai_persona_kafka_produce_total` | PRESENT | 040 task C |
| `refusing dial to structurally invalid address` | PRESENT | 040 empty-host guard |
| `kafka produce succeeded after retry` | **absent** | task D, committed AFTER the build — **this is the control that shows the probe discriminates** |
| `ai_persona_kafka_dial_total` | PRESENT | pre-existing positive control |
| `zzz_nonexistent_marker_qqq` | absent | negative control |

**Committed but NOT yet live — rides the next roll:** `9b93af8a0` (040 round-2 cardinality + label
split) and the opt-in produce retry + classifier needles.

## 2. The commits

| commit | what |
|---|---|
| `ca5e41122` | 343: id-keyed arrival check, `parkOutcome`, table cross-check (RSH-012, WFA-021) |
| `7f3875d3c` | 343 round 2: the test for the guard-in-series gap the council exposed |
| `bf1fbc5b7` | 343 P2: loop-skip retry discipline + mark-after-persist (RSH-013) |
| `e4ce7073b` | 040: produce counter, empty-host dial refusal, PodMonitor wired into kustomize (SYS-092) |
| `9b93af8a0` | 040 round 2: closed `system.*` family set, `client_no_leader`/`broker_no_leader` split |
| *(retry commit)* | 040: opt-in bounded produce retry, 4 adopters, 2 needles (SYS-093) |
| `340f9e218`, `5192e1e23`, `7e269d4ab` | docs, landmine provenance, council trail |

## 3. ~~THE FIRST THING TO DO NEXT — verify the second roll~~ **DONE 2026-08-22 — and it passed**

> **The second roll landed: `v1.0.1323`, both pods, 08:36:48Z / 08:37:14Z.** Binary-probed with a
> discriminating control — `kafka produce succeeded after retry`, `client_no_leader`, `system.other`
> and `ai_persona_kafka_produce_retry_recoveries_total` all PRESENT, nonsense control ABSENT. **Every
> change from the 08-21 session is now live.** Live figures and the roll-spanning-window trap that
> nearly had me file a regression are in `bugs_open/040` §12.7. Owner parks (040 residuals 1 and 3,
> and the close of 343) are in §12.6 and in the 029 lane. The recipe below is kept for the NEXT roll.

### The recipe, for next time

The round-2 fixes and the retry are inert until the next build. When it lands:

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <the retry commit> <that stamp>     # per SERVICE, not per fleet
```
Greppable literals for the binary probe: `kafka produce succeeded after retry`,
`client_no_leader`. **Always run a control that must be ABSENT in the same breath.**

Then, and only with its demand control:

```promql
sum by (outcome, topic_class) (max_over_time(ai_persona_kafka_produce_total[24h]))
```
**`ok` ≫ 0 IS the control.** A zero there means the instrument is broken, not the fleet clean.
⚠ `max_over_time`, never `increase()` — these counters live on ephemeral Jobs and are frequently born
at their final value (040 §10).

## 4. What is STILL OPEN — do not read any of the above as a close

**343** — the **first death** is untouched and unexplained: why a child stopped answering, and why the
parent never registered iteration N+1's `call_handler` at all. Everything shipped explains why a job
*stayed* stuck. RSH-011 `wedge-evidence-capture` is armed and its trigger is **deliberately broad** —
do not narrow it. ⚠ The entry condition has been ~0 since 08-18 and **six of the eight days around the
08-17 burst were also zero before anything changed**, so quiet is not evidence. The informative read is
`error_code='await_divergence'` in `agent_error_log` **paired with a non-zero await volume**.

**040** — three things:
1. **The `timeout` residual: 146 in 7 days, prod-0 dominant, undiagnosed.** §4.2's node-pinned `nc`
   probes are **still unexercised** and still the best untried lead, now with a named broker. §7's
   traps apply: normalise by pod uptime, brokers are in namespace `kafka`, busybox `date +%s%N`
   returns 0.
2. **The `refused` mechanism is `[INFERRED]`.** The candidate is kafka-go's own consumer-group
   coordinator lookup (`consumergroup.go`, `net.JoinHostPort` from an unvalidated FindCoordinator
   response). **The cheap confirmation is the new `empty_host` label on the next burst** — a `090` run
   is only worth firing if a burst arrives before that label is live, and then hand it the Prometheus
   evidence inline (the loop cannot reach Prometheus; that is what made §11.4's verdict UNVERIFIABLE).
   **Named disconfirming result: `refused` with an empty broker label must now be structurally zero. A
   non-zero means a THIRD producer exists outside the instrumented dial path.**
3. **13 adapter/service Deployments still serve no `/metrics`** (§8b). Every figure covers the chassis
   and spawned agents only. Its own round; do not bundle it.

**Also watch, because the retry accepts it:** a duplicate reply after a lost ack is possible (no
idempotent producer in kafka-go 0.4.47) and is *assumed* absorbed by the parent's two-phase
`ClaimAwaitedRequest`. **Named disconfirming signal: `DUPLICATE_SKIPPED` volume rising after the roll.**

## 5. Traps this session hit, so you do not

- **`097` prints `SAVE: SUBMISSION_CORR=…` BEFORE it publishes.** Read the LAST line. A failed publish
  (`AlreadyExists` on the epoch-second `kcat` pod name) leaves you holding a correlation for a dispatch
  that never happened — and CLAUDE.md's "a missing row is latency, do not retry" is exactly wrong there.
- **`RESUBMIT_CORR` is an ENV VAR**, or a bare uuid as arg 2. Passing `RESUBMIT_CORR=<uuid>`
  positionally makes the literal string the trail id; the `commit-msg` trailer gate then refuses the
  commit, correctly.
- **The council submission schema is in the SCRIPT header, not CLAUDE.md's summary** — `plan` is an
  object with `summary`/`edits[]`/`grounded_in`/`risks`, and `symbol` is per-edit. `DRY_RUN=1` is free.
- **`UpdateStateWithRetry` discards your mutations** (`*state = *reloaded`). LANDMINES entry filed.
- **Grep the right pods.** The `refused` counters are on spawned `app=dynamic-agent` Jobs, not
  `app=agent-chassis`. A selector that cannot hold the line returns a zero that means nothing.
- All in `WRONG_CALLS.md` (2026-08-21 entry) with the cheap checks that would have caught each.

## 6. Where everything is

- Bugs: `bugs_open/343_HANDOFF_2026-08-20_…md` (08-21 block at the top),
  `bugs_open/040_HANDOFF_2026-07-20_…md` (§12).
- Lanes: `docs024_key_docs_latest/bugfix_029_retry_kills_live_child/` (NOTES §28, README) and
  `docs024_key_docs_latest/bugfix_040_kafka_dial/` (NOTES, README, `COUNCIL_TRAIL_2026-08-21.md`).
- Register: **RSH-012**, **RSH-013**, **WFA-021**, **SYS-092**, **SYS-093**.
- LANDMINES: the legacy no-id marker branch, `UpdateStateWithRetry`, kafka-go's `MaxAttempts` mirage,
  and the per-job-topic cardinality rule.
- Submissions (reusable as templates): `submission_343_park_identity.json`,
  `submission_343_p2_loop_skip_persist.json`, `submission_040_produce_counter_and_empty_host.json`,
  `submission_040_optin_produce_retry.json`.
