# NOTES — bugfix 129, retry replay

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-28, ~19:40 BST — picking the bug

Swept all 47 files in `bugs_open/` through `scripts/who-owns.py`. Most are OWNED
by a live workstream. The genuinely unowned set came out as 080, 081, 091, 097,
104, 106, 111, 114, 115, 120, 122, 129, 132.

Discarded, with reasons, so nobody re-derives them:

- **132** (every B2 site serves a raw JSON error blob instead of a 404). Excellent
  bug, and its own fix candidate 1 is a **Cloudflare Worker**. Grepped all of
  `~/projects` for `objectKey` and `"B2 returned error"` — **the emitter is not in
  any local repo**, and there is no `wrangler*` or worker source anywhere on this
  machine. So the fix is edge configuration with no code in this tree to commit.
  Left for a thread with Cloudflare access.
- **120** (merge commit skips the deploy). Also good, also unowned — but the file
  to change is `.github/workflows/deploy-to-b2.yml` in **`~/projects/sites`**
  (`git@github.com:gqls/sites.git`), a different repo. Out of scope for a chassis
  build; a separate task.
- **091, 097** — `bugs_sweep_2026_07`'s own handoff names 091 as its next item and
  it already shipped 097's repair half. Not unowned in practice, only in the
  commit record. **`who-owns` reads commits, so a session mid-fix is invisible.**

Took **129**. Chassis Go, framework-level, diagnosis already CONFIRMED, and the
citing dirs cite it only as a cross-reference.

## 2026-07-28, ~19:50 — the first wrong turn: I fixed the file the log named

129's evidence points at `coordinator.go:615` (`getOrCreateState`) and
`coordinator.go:780` (`handleOrchestrationStatus`), and its three fix candidates
are all child-side. I started by designing a change to `getOrCreateState` so the
child could not resolve the parent's row.

**That would have treated a symptom.** The question I had not asked was *why the
child was given the parent's orchestration_id in the first place*. Nothing in
the child does that — the child reads what it is sent.

Chased the sender instead. `spawn_actions.go:370` sets
`message.Headers.OrchestrationID = agentID` — the **child's** own id — and
`call_agent.go:573` sets `= childOrchID`, likewise correct. So the *original*
messages are fine, and the defect had to be in something that sends a **second**
message for the same request. That is the retry.

## 2026-07-28, ~19:55 — the second wrong turn: I fixed the DEAD retry path first

Found `TimeoutMonitor.retryTimedOutRequest` in `helpers.go:260`:

```go
OrchestrationID:   state.OrchestrationID,   // the parent's
```

and wrote it up as the root cause. It is the right *shape* — but it is **not the
path that fired**. The error string in 129's evidence is
`Request <id> timed out after 3 retries`, and grepping for it lands in
`coordinator.go:3148/3155/3252`, not in `helpers.go`. The live path is
`retryExpiredAwaitedRequest → handleRecoverableError`, which carries the identical
defect at `coordinator.go:2929` **plus two more** the helpers version does not
make as obvious: `Action: "execute"` and a stub body.

*The cheap check that caught it:* grep the **error string from the evidence**,
not the concept. I had matched on "this looks like a retry that sends the wrong
id" and stopped at the first hit. Both paths are now fixed — the dormant one
because leaving a known-defective twin is how it comes back.

> **Both of these are the same misstep twice:** I matched on shape and stopped
> looking. The first time it put me in the wrong *file family* (child instead of
> sender), the second time in the wrong *function* within the right family.

## 2026-07-28, ~20:05 — what the measurement changed

Before measuring I would have written "this fires occasionally, on spawn". The
queries said something much stronger, and also killed one thing I was about to
claim.

```
agent: synthesised retry (poisoned) | 430 retried in 14d | 294 exhausted
adapter: re-executes step           |   0                |   0
```

**Every retry the fleet sent in fourteen days took the poisoned path.** The
adapter re-execute branch — the one whose comment already says "adapters need the
full payload" — has not been taken once. I had assumed it carried a meaningful
share and would be the natural model for the fix.

All-history by `retry_version`: `1 → 93`, `2 → 45`, `3 → 294`. A retry that
recovers *decays*; this one **accumulates at the cap**.

**[CORRECTED before it was written down]** I nearly wrote "retries never work".
The 108 rows that ended `processed` at `retry_version ≥ 1` are indistinguishable
in this table from a **late original response** arriving after the retry was sent.
So the durable claim is *"68% of retried requests exhausted the budget"*, which is
measured, and nothing stronger. `[UNMEASURED]` how many of the 108 the retry
actually rescued.

## 2026-07-28, ~20:10 — the design decision, and the one I rejected

The obvious cheap fix is to generalise the adapter branch: on timeout,
**re-execute the step**. Correct by construction (the same code rebuilds the
request, so the payload cannot be wrong) and needs no storage at all.

**Rejected**, and the measurement is why: `spawn_dispatch` is 154 of the 430
retried requests, and re-executing a `spawn_agent` step **spawns a second pod**
and orphans the first. That is the `bugs_open/124` double-dispatch class. The
adapter branch is left exactly as it is.

So: store what was sent, replay it. The three things that made this cheaper than
it looked:

1. `handleRecoverableError` **already** calls `repo.GetAwaitedRequest` to read the
   authoritative `retry_version`, so the payload arrives with no extra query.
2. Actions already pass retry-routing facts to the coordinator through the result
   map (`requests_topic`, `responses_topic`, `target_agent_type`). One more key is
   the existing convention, not a new seam.
3. Only **two** actions can reach this path fleet-wide.

On (3) — I checked rather than assumed, because "which actions send awaited
requests" read off the *code* gives ten-plus (`hitl_*`, `thunder_*`,
`spawn_group`, `SpawnAgentActionOld`…). Asked of `agent_definitions` instead, the
answer for the whole live fleet is **`call_agent` and `spawn_agent`**. Everything
else is unreachable by configuration. That is the difference between a fix with
eight capture sites and a fix with two.

> **CORRECTED 2026-07-28 ~21:30 — "everything else is unreachable" is FALSE, and
> the check above is the reason.** It asked *which spawn/call actions are seeded*;
> coverage turns on *which actions await a response*. Six of 428 retried requests
> in 14 days come from `scrape_web`/`web_search`, which are seeded and do await.
> See the entry at the bottom of this file for the split and why those six are a
> different defect rather than a coverage miss. Caught by the council seats
> pressing on the coverage claim.

## 2026-07-28, ~20:20 — where the payload must NOT live

First draft put `RequestPayload` on the `AwaitedRequest` struct with a normal JSON
tag. That is wrong and would have been invisible: the same struct is serialised
into `orchestration_states.awaited_requests`, a JSONB column **rewritten on every
state update**. A `call_agent` body (prompt + input_data + context) would have
been re-serialised many times per orchestration.

`json:"-"` + a column on the per-request table. There is now a test that fails if
anyone removes the tag (`TestRequestPayloadStaysOutOfTheStateJSONB`), because
nothing else enforces it and the cost is silent.

Second thing nearly broken: `InsertAwaitedRequest`'s `ON CONFLICT DO NOTHING`. I
started to make it `DO UPDATE` so a pre-registered row would get its payload. But
the `rows == 0 → "already exists"` error is **load-bearing** — its caller uses it
to decide whether to arm a timeout handler. `DO UPDATE` always affects a row, so
that signal would have vanished silently. Kept `DO NOTHING`; back-fill is a
separate guarded `UPDATE … WHERE request_payload IS NULL`.

## 2026-07-28, ~20:35 — honest limits of the tests

`platform/orchestration/types/retry_payload_test.go` pins the invariant, and it is
worth being clear about what it does **not** do: it exercises `ReplayRequest`,
which did not exist before this change, so it **cannot fail against the old
code**. It is a pin, not a regression test in the strict sense. The genuinely
discriminating artefact is the binary grep — `is_retry` is a string this change
*deleted*, present in v1.0.1192 and absent from v1.0.1193.

`platform/orchestration/retry_payload_capture_test.go` is closer to a regression
test: `TestRequestPayloadStaysOutOfTheStateJSONB` fails on a future edit that
removes `json:"-"`, and that edit is otherwise silent.

## 2026-07-28, ~20:45 — build and migration

Migration 263 applied by hand. **Not `--apply`** — the dry run showed ~20 pending
files belonging to other threads, and `--apply` takes every one of them.
Recorded with `--record-only` afterwards.

Image `v1.0.1193` built from committed `HEAD` and pushed. Distinct image id
(`17b2e8ebe040`) and timestamp — checked, because a retag is not a rebuild and
1188/1189 once shared one image id built 56 minutes before the fix in it.

**Deliberately NOT deployed yet.** A chassis roll kills an in-flight council
round and `EXECUTING_STEP` hides it for an hour. The council round for this
change is live, so the deploy waits for the verdict.

## 2026-07-28, ~21:10 — the council came back REJECTED, and it was right about the venue

`75cb2fdc-e74c-4d3d-99b7-9264548e65d6` → `rejected`, `decided_by: "hard veto from
guardian"`, `unreadable: 0`, `abstained: 6`. **`unreadable: 0` matters** — this was
a judgement, not the ~11% of rounds that die on one seat's unparseable JSON.

Six approvals, no seat disputing the diagnosis, and a veto purely on **scope**:
a new shared contract + a schema column + two changed signatures + three
coordinator edits, arriving as one bug patch. Same finding as 124's `$ctx.` veto.
Full record and the three costed options: `REVIEW_2026-07-28_council_scope_veto.md`.

Not resubmitting. The owner ruling is explicit that a SCOPE veto is a judgement
about *how* a capability reached production, not a measurement to be improved —
and the seats **disagree with each other** on the remedy here, which the ruling
names as the case that needs a human.

## 2026-07-28, ~21:20 — three checkable objections, and one of them made me wrong

The scope veto is for a human, but the other seats raised checkable things.

- **bug_historian, medium — real, fixed.** My back-fill
  `UPDATE … WHERE request_payload IS NULL` discarded rows-affected. It succeeds
  with `err == nil` when it matches nothing, so checking only the error could not
  tell "backed-fill" from "did nothing", and the sole other surface was
  `RETRY_PAYLOAD_UNAVAILABLE` minutes later with the write-time cause long gone.
  Now checks rows-affected and, on zero, distinguishes the benign duplicate from a
  real gap (`RETRY_PAYLOAD_BACKFILL_MISSED`).
- **editquality, high — right about the submission, not the change.** "No edit
  creates migration 263." The migration existed, was committed, and was already
  applied when the round returned — but it was in my *rationale* and not in the
  eight-edit list. **Reviewers cannot open files.** An edit list that omits a file
  is, to them, a plan that omits it. Cheap to avoid, so it goes in the runbook.
- **prior_art / guidelines, low — measured, and better than I claimed.**
  `SpawnAgentActionOld`/`Old2` are **not in the action registry at all** — dead Go
  functions, not an unpatched router branch, so not the 093 class. `spawn_agent_k8s`
  is an alias whose Handler is the wired `SpawnAgentAction`.

## 2026-07-28, ~21:30 — [CORRECTED] my "complete coverage" claim was false

> **CORRECTED — I wrote "complete coverage" in the submission and it was wrong.**

The seats' pressure on coverage made me re-ask the question, and my check had
silently encoded a different one. I ran *"which spawn/call actions are seeded"*.
The question that decides coverage is *"which actions await a response"*. Those are
not the same question, and the first one cannot falsify the second.

Split on whether `call_agent`/`spawn_agent` actually produced the request:

```
call_agent/spawn_agent (WIRED)                      | 422 | 289 exhausted
adapter (re-executes step, untouched)               |   0 |   0
OTHER awaited sender (NOT wired — would now refuse) |   6 |   4
```

The six are `scrape_pages` (3, all exhausted) and `search_web` (3, 1 exhausted).

**But they are not a coverage miss — they are a different defect.**
`web_search_action.go:139` puts `params.ExecutionContext.OrchestrationID` — the
**caller's own** id — on the *original* outbound message, not just on a retry. So
there is no child identity to replay; wiring them would only make
`RETRY_SELF_ADDRESSED` fire on their originals. Own mechanism, own diagnosis, and
bundling it here would be precisely what the guardian vetoed.

Honest behaviour change: those six now fail fast instead of retrying. Four
exhausted anyway ⇒ **≤2 requests a fortnight** lose a retry that might have worked,
in exchange for a named error instead of silence.

*The cheap check that would have caught it first time:* split the retried
population on `target_agent_id` — `call_agent` and `spawn_agent` both always set
it, so its absence names the senders neither of them produced. One query, and it
is the query that answers the question rather than a neighbour of it.

## 2026-07-28, ~21:40 — deliberately NOT deployed

v1.0.1193 is built, pushed and pod-grep-discriminated in both directions.
Migration 263 is applied and recorded (additive, nullable, inert against the
current binary). **The fleet has not been rolled.**

A chassis roll is fleet-wide and outward-facing, and the 07-28 owner ruling says a
SCOPE veto gets broken by a human, not by the thread that drew it. So `bugs_open/129`
**stays open** — the bar for `bugs_closed` is fixed AND live, and this is fixed and
built, not live. Everything needed to make it live is one `make deploy-agent-chassis`
once someone rules; everything needed to reverse it is not doing that.

## 2026-07-28, ~22:00 — it went live anyway, and THAT is the finding

I deliberately did not deploy: the council vetoed on scope, and the owner ruling
says a human breaks a SCOPE veto. **It shipped regardless.**

The fleet is on **v1.0.1194**, rolled at **20:48:11Z** by another session. My
commits landed at 19:57Z (`eb70c3dd3`) and 20:16Z (`17bd675dd`) — both *before*
that build. `make build-<service>` builds from committed `HEAD`, so whoever rolled
next carried my change out with theirs, exactly as designed.

Verified rather than inferred, pod-grep on **both** running pods:

```
is_retry                       0     <- the string this change DELETED
RETRY_PAYLOAD_UNAVAILABLE      1
RETRY_SELF_ADDRESSED           1
MISROUTED_REQUEST              1
RETRY_PAYLOAD_BACKFILL_MISSED  1     <- from the LATER commit, so both are in
```

> **The finding, and it is bigger than this bug: on this tree, "hold the deploy
> pending review" is not a strategy that exists.** The platform-seam ruling assumes
> the committing thread controls when its seam reaches production. It does not.
> HEAD is shared, `build-<service>` builds from committed HEAD *by design* (so that
> a build cannot bundle WIP), and the fleet rolled ~8 times on 2026-07-28. So the
> moment a seam is committed it is in the next roll, whoever fires it and whatever
> verdict is outstanding. **Committing IS shipping, on someone else's schedule.**
>
> The two live options for a thread that wants to hold a seam back are therefore
> (a) don't commit it — which CLAUDE.md forbids, because a long-lived dirty tree is
> shared mutable state that another session will sweep, and (b) commit it behind a
> flag or a config switch that defaults off. **(b) is the only one that actually
> works**, and nothing in the current ruling asks for it. Written up in the REVIEW
> doc as the thing the owner may want to change about the ruling itself.

## 2026-07-28, ~22:05 — the capture half is PROVEN on live traffic

Two awaited requests since the roll, one from each wired action, both recorded:

| step | awaiting orch | target orch | parent orch | action | body | payload |
|---|---|---|---|---|---|---|
| `call_scraper` (call_agent) | `b89b6e5e` | **`ef7e2ddb`** | `b89b6e5e` | `process` | 210 B | 1186 B |
| `spawn_scraper` (spawn_agent) | `b89b6e5e` | **`8b5dc669`** | `b89b6e5e` | `initialize` | 447 B | 1124 B |

Every column is the fix: the target orchestration id is the **child's**, not the
awaiting one; the parent link is set (the old synthesised message had none); the
action is the real one (`process`/`initialize`) rather than the hardcoded
`"execute"`; the body is present rather than a stub.

The invariant query returns **0**, as it must:

```sql
SELECT count(*) FROM awaited_requests WHERE request_payload IS NOT NULL
  AND request_payload->'message'->'headers'->>'orchestration_id' = orchestration_id::text;
```

**And it answers the council's size risk with a number instead of a worry:**
`pg_column_size(request_payload)` is **~1.1–1.2 KB**, not the tens of KB
`bug_historian` and I both guessed at. `[UNMEASURED]` a `call_agent` carrying a
full LLM prompt will be larger — these two are a scraper's payloads. Re-measure
the p95 once a day of mixed traffic has accumulated.

## 2026-07-28, ~22:10 — the REPLAY half is NOT yet witnessed

`Replaying original request` count over 90 minutes of live logs: **0**. So is
`Retrying request`, and so is every one of the four new error markers — i.e. this
is *no timeouts have happened*, not *replays are failing*. The fleet is quiet
tonight; the capture half had only 2 rows to work with.

**So the fix is half-proven and I am not going to call it more than that.** Capture
is proven on live traffic. Replay is proven only in unit tests, which exercise a
function that did not exist before and therefore cannot fail against the old code.
Induced a real one via `TRIGGER_code_indexer_v2.sh` (correlation
`c54b3fdf-b556-45f1-8e59-8237bec64d2a`) — the lane bugs_open/129 records as failing
2 of 3 on v1.0.1184.

## 2026-07-28, ~22:30 — the REPLAY half IS witnessed. Bug closed. (thread "bugsearch 5")

The wait was about 20 minutes. A real `call_scraper` request sent at 20:55:17Z with
`timeout_seconds: 1800` expired at 21:25:17Z, and the coordinator replayed it at
21:25:18.5 — a **natural** timeout, not an induced one, which is better evidence
than the induced run I was chasing.

```
coordinator.go:2984  "Replaying original request to target agent requests topic"
  request_id             30585e6d-ade5-406b-a90b-b64222de0853
  child_orchestration_id ef7e2ddb-d58e-4e5f-a391-421f9fadbc3f   ← awaiting orch is b89b6e5e
  action                 "process"                              ← old code forced "execute"
  retry_version          1
  pod                    business-intel-765566d57-wxpsr
```

`processed_at 21:27:05.999`; child orch `ef7e2ddb` → `complete/COMPLETED` at
21:27:06.938, parent `b89b6e5e` at 21:27:07.104. Every assertion the handoff named
as closing evidence is met, on the one request class the old code destroyed.

Detail worth keeping: the stored `request_payload` still reads `retry_version: 0`
and the original 20:55:17 timestamp. That is **correct** — it is the captured
original; only the wire copy is bumped. Seeing `0` in the stored blob next to `1`
on the row briefly looked like a bug and is not.

### Misstep 1 — I grepped the wrong pod label and nearly recorded a false negative

My first log grep was the handoff's own command, `-l app=agent-chassis`. **Zero
lines.** The retry runs in whichever service hosts the *awaiting orchestration*,
which here was `business-intel`. Every service runs the same image
(`agent-chassis:v1.0.1194` — one binary, many app labels), so the code is present
everywhere; only the *execution* is localised. Had I stopped at that grep I would
have written "no replay has occurred" while the replay sat in the next pod's log.
The handoff's §2 command has this defect; fixed in the RUNBOOK and flagged in the
bug file.

Corollary: the log line is **stronger** evidence than the pod-grep the handoff
leaned on. `strings | grep` proves the binary *contains* the code; the log line
proves it *ran*.

### Misstep 2 — I read a live request as expired

Saw `status='waiting'` with `timeout_at 21:28:18` and called it timed-out. It was
21:27:45. The row was **still in flight** and reached `processed` moments later.
Cheap check I skipped: `SELECT now()` in the same query as the row. Doing that from
the start would also have saved the wrong inference in Misstep 1, because the
sequence would have been obviously mid-flight rather than dead.

### The §5 rule in the handoff is too strict — corrected

§5 says the NULL-payload population should be only `scrape_pages`/`search_web` and
"anything else is a genuine gap". The live census says `deploy_page`/`unknown` ×4.
Not a gap: `coordinator.go:2809` branches
`TargetAgentID == "" && RequestsTopic LIKE 'system.adapter%'` into step
**re-execution** and returns at :2881, never reaching `DecodeRetryPayload` at :2947.
Adapter actions re-run the step with full context instead of replaying a message,
so they have no use for a stored payload. All 4 rows match the predicate exactly.

Census split by the path each row would actually take, since the roll:
**adapter re-execution 4 · replay path 18, of which 18 carry a payload · genuine
gaps 0.** The corrected query is in the bug file and the RUNBOOK.

Counter-markers `RETRY_SELF_ADDRESSED` / `MISROUTED_REQUEST` /
`RETRY_PAYLOAD_UNAVAILABLE` / `RETRY_PAYLOAD_BACKFILL_MISSED`: **0** across five
services over 3 hours.

### What is NOT closed by this

The council's SCOPE veto and the "committing is shipping" clause are owner calls
about *how* the change shipped; they are untouched by this verification and live on
in `REVIEW_2026-07-28_council_scope_veto.md`. `scrape_web`/`web_search` still record
no payload by design and need their own diagnosis — deliberately not bundled here,
that being the exact thing the guardian vetoed.

## 2026-07-29 — re-verified on v1.0.1196, and §5 corrected: the defect is per-ACTION, not per-step-name

Fresh build rolled 22:37:49Z. Pod-grep re-run on both pods per this lane's own
landmine: `is_retry` **0**, all four new markers **1**, plus 124's
`unknown execution-context field` **1**. Behavioural evidence is better: a **second
natural replay** ran on the new image (`call_scraper` retried 03:25:19 →
`processed` 03:26:08).

### Misstep — I removed a time filter and manufactured a catastrophe

Chasing 6 unrecorded payloads I ran the capture census with **no `sent_at` bound**
and got ~7,000 NULL-payload rows across ~100 step names, `spawn_dispatch` at 720.
For a few seconds that read as a fleet-wide regression of the fix.

It is **pre-fix history**: `request_payload` has only been written since the
07-28 20:48 roll, so every row before it is necessarily NULL. Correctly scoped:
**145 replay-path requests, 139 recorded (95.9%)**, and the only unrecorded step is
`search_news`.

The cheap check I skipped: **when a column is new, its "missing" count is dominated
by the period before it existed** — bound every census by the migration/roll
boundary, or the denominator is a different population than the one you mean. This
is [[narrow-filter-defines-the-conclusion]] running backwards: usually a *narrow*
filter flatters you, and here *removing* one did the opposite. Same root discipline:
state the population, then check the query describes it.

### §5's enumeration was by step name; the defect is per action

`search_news` looked like a genuine new gap by §5's wording. It is not —
`FetchNewsSearchAction` **delegates to `WebSearchAction`**
(`feed_fetch_async_actions.go:177`), the same sender §5 already names. Measured
scope: **10 step instances across 9 agents** over the two actions, under 7 distinct
step names (`search_web` ×5, `search_area`, `search_domain`, `search_practice`,
`search_news`, `scrape_source`). §5 listed two.

The cost figure survives (`search_news`: 193 rows/14d, retried **0** times, so
nothing is losing a retry today), but the exposure is 5× wider than documented.
Corrected in the handoff §5 with the query that enumerates by action.

**Transferable:** a doc that lists *instances* of a defect goes stale the moment
someone adds another caller; a doc that names the *shared function* does not.
Whenever a "known exceptions" list is written, write the predicate that generates
it, not the list.
