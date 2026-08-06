# NOTES — `bugs_open/173` per-substep `continue_on_error`

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-04 — picking the bug, and a measurement that could not have come out otherwise

**Task:** take the next `bugs_open/` case not already being worked, fix it robustly at the
framework level rather than the individual case.

### MISSTEP 1 — my first ownership measurement was non-disconfirmable

To find an unclaimed bug I grepped every live session transcript for `bugs_open/NNN` and
counted **mentions**. It returned "mentioned in 4–16 active transcripts" for *every*
candidate, including ones nobody was touching.

**Why it was worthless:** every session runs `ls bugs_open/` at some point, so all 54
numbers appear in most transcripts. The measurement **could not have returned a low number
for any open bug**, which is the disqualifying property — dated and run is not the same as
disconfirmable (`a-count-you-kept-is-not-a-census`, and the 2026-08-03 pair in
`WRONG_CALLS.md`).

**What replaced it, and it worked:** grep the transcripts for tool-call *file paths* —
`"file_path":"…/bugs_open/NNN…"` — i.e. which bug files a session actually **read or
edited**. That returns 26 numbers out of 54 and discriminates sharply. Logged in
`WRONG_CALLS.md`.

### MISSTEP 2 — `who-owns.py`'s verdict is dominated by the FILING lane

`./scripts/who-owns.py 173` returns **OWNED or recently active**, naming
`bugfix_165_reconciliation_deletes`. That lane **filed** 173 as a follow-on *for someone
else* and has since closed (`bugs_closed/165`; its own summary is titled *"all four guarded
and the lane closes"*). The tool reads commits, so "the lane that wrote the bug file" and
"the lane fixing it" look identical.

Not a defect in the tool — its own output separates "likely OWNING workstream" from "also
cites it", and its verdict is deliberately conservative. But **its verdict line alone is not
an answer**; read the named lane's state before believing it.

### First pick, abandoned: 093

I first picked `bugs_open/093` (the stat audit's one guarded call site). It is a genuinely
good structural bug — but reading it to the end showed its own 2026-07-27 triage says the
fix **is already built and live** and *"093 is not a code task any more. It is blocked on
`bugs_open/083`"* (the check ships, but `improvement-sweep` is disabled so nothing ever runs
it). And 083 **is** being actively read by a live session.

**Cost:** ~5 minutes. **What would have avoided it:** reading the bug file's *last* dated
section before its first. On this estate a bug file's header is frequently stale relative to
its own updates — 093's header still says "OPEN, not started" while four later sections
record it built, shipped and pod-verified.

### Settled on 173

- Its own header says **OPEN, UNOWNED**.
- No live session has read the file.
- The filing lane is closed.
- It is a missing degree of freedom in the orchestration engine — framework-level by
  construction, which is what the task asked for.
- Filed at the direction of the council's `architecture` and `constitution` seats, so the
  shape wanted is already on record.

---

## 2026-08-04 — verifying the bug is still valid, and one finding the file does not contain

The tree moves under us, so validity first. `loop_expansion_handler.go:157` still reads
`injectedStep.Config["continue_on_error"] = continueOnError`, unconditional. **Still valid.**

**New finding, not in `173`:** line 104 deep-clones the substep's config into the injected
step, so a substep's own `continue_on_error` **is already present** in `clonedConfig` — and
line 157 then overwrites it. The key is therefore **declared-and-inert today**: an author can
write it, the config-key audit will not object (it is in `frameworkStepConfigKeys`), nothing
warns, and the value is silently discarded. Registered as a landmine.

### Four checks done by reading, so the fix does not guess

1. `models.DecodeSubWorkflowStep` does `step.Config = config` verbatim → a substep's config
   key survives the decoder; **no** `SubWorkflowStepFields`/`models.Step` change (which
   WFA-003's registered landmine explicitly warns against).
2. `continue_on_error` is already in `datahelpers.frameworkStepConfigKeys:144` → no audit
   change, and no audit silenced.
3. All three coordinator decision sites (`coordinator.go:907`, `:3239`, `:3345`) go through
   the single `shouldContinueLoopOnError` → resolving once at expansion covers all three.
   **Checked rather than assumed**, because "one guarded call site" is exactly the family
   `bugs_open/093` exists for.
4. `loop_metadata["continue_on_error"]` (line 85) is loop-level and must stay so —
   `skipToNextLoopIteration` deliberately does not rely on that shared key.

### The blast-radius measurement, with its positive control

`0` live substeps declare `config.continue_on_error`. **[MEASURED 2026-08-04]**

The control matters more than the result: the same query, grouped instead of filtered,
returns **18 loop steps / 79 substeps** (9 `true`, 1 `false`, 8 unset). So the predicate
demonstrably finds loops and substeps and a non-zero answer was reachable. Without that, "0"
is equally consistent with "my `s.value->>'action' = 'loop'` predicate matches nothing".

> **CORRECTION to `173`'s own contribution figures.** It recorded 20 loops (10 unset / 9 true
> / 1 false) on 2026-08-01. Today: **18** (8 unset / 9 true / 1 false). Two unset loops
> retired in between, consistent with `7a15c3a47` "retire(agents): three unused builders are
> out". The bug's conclusion is unaffected — all nine `true` loops are still fan-out/dispatch
> loops and none wraps a destructive reconciliation. Recorded because CLAUDE.md requires
> figures to be grounded before being repeated, and this one had drifted in three days.

### Residual, deliberately not fixed here

`loop_actions.go:66` reads the **loop-level** flag with a bare `.(bool)` assertion, so a
loop-level `"true"` (string) is silently ignored — the same declared-and-inert shape the fix
makes loud at substep level. Left alone: it is a separate contract on a separate line, and
widening it changes how 18 live loops parse their own config. **[UNMEASURED]** whether any
live loop declares it as a non-bool — worth a query before anyone touches it.

---

## 2026-08-04 — the bug's MOTIVATING loop no longer exists; the defect and its blast radius do

Owner ruling 2026-07-29 §3 says a shared mechanism's other consumers must be **told**, not
merely measured. So I went to name them — and found the headline one gone.

`173` opens on `multipage-website-builder` → `generate_pages_loop` → `extract_links`. That
agent is **deleted and inactive** as of today:

```
 type                      | is_active | snap | deleted | loop_steps
 multipage-website-builder | f         | f    | t       |          1   (x2 rows)
```

Consistent with `7a15c3a47` *"retire(agents): three unused builders are out"* — and it is the
same retirement that took the loop census from 20 (2026-08-01) to 18 (today).

**This does not weaken the bug; it relocates it.** The three loops named in `173`'s own
contribution section are all still live, all still `continue_on_error` **(unset)**, and all
still carry `save_page_sections` — the action that gained a completeness floor and can now
refuse:

| agent | loop | substep | action | loop `continue_on_error` |
|---|---|---|---|---|
| `pageflow-builder` | `build_pages_loop` | `save_sections` | `save_page_sections` | (unset) |
| `page-rebuild` | `build_pages_loop` | `save_sections` | `save_page_sections` | (unset) |
| `site-work-orchestrator` | `build_items_loop` | `save_sections` | `save_page_sections` | (unset) |

So the disproportion `173` describes — **one page's refusal fails an entire multi-page
build** — is live in three agents today, and these are the consumers to tell. The one lane
that would have been hardest to reach (the retired builder) needs telling by nobody.

> **CORRECTION, and it matters for anyone quoting this bug.** `173`'s § "Why it matters —
> the concrete case that produced this file" is now **historical**: the case it describes
> cannot recur, because the agent was retired. Quote the three rows above instead. The
> reasoning in the file is unaffected — it never depended on that agent, only illustrated
> itself with it. **[MEASURED 2026-08-04]**, query in the runbook.

**Why I checked at all**, since the temptation was to copy the four rows out of the bug file
into the council submission: a closed bug's scope-out expires
(`a-model-upgrade-can-invert-a-closed-bugs-premise`), and so does an open one's example. The
submission would have named a deleted agent as a live consumer, and the guardian seat would
have been right to ask whether I had looked at anything at all.

---

## 2026-08-04 — council APPROVED round 1, and the three objections answered with checks

Correlation `549e25fb-acc1-4806-a2a7-95bf73cca806`. **APPROVED, round 1**, 8 seats approving,
3 advisory objections, **none high-severity**, 7 abstained (relevance-gated).

The norm here is that objections come back with the reviewers' own read-only checks already
answered, so answering them with a nod would be the wrong shape. All three are answered below
with a check.

### `guardian` (medium) — "the fallback_step/retry interaction is UNVERIFIED"

Fair: `173` listed it as not looked at, and so did my submission. **Now looked at, and the
answer is that there is no interaction, because there is no mechanism.**

`fallback_step` and `retry_step` appear in the **entire** Go codebase exactly twice, and both
are name-prefixing lists — `loop_expansion_handler.go:528` (`stepRefFields`) and
`coordinator.go:4243` (`stepRefKeys`). There is no `FallbackStep`/`RetryStep` field on
`models.Step`, and **nothing reads either key to make a routing decision**. Fleet-wide, **0
live agent definitions declare either key** `[MEASURED 2026-08-04]`.

So the two keys are rewritten as if they were live routing and consumed by nothing — which is
*the same declared-and-inert class as the bug this lane just fixed*, one level out. Not filed:
unlike `continue_on_error` these are declared by nobody, so there is no author to mislead. Worth
knowing before anyone "restores" them.

### `bug_historian` (medium) — "does a tolerance-skip leave a durable trace, or is it a silent loss?"

The sharpest objection of the round, because the seat is guarding the exact pattern this
platform keeps re-finding. **Answer: three traces, and it is not the silent-drop shape.**

1. `skipToNextLoopIteration` (`loop_error_handler.go:139-158`) writes
   `{loop}_iter_{N}_error` into `CollectedData` — `status:"error"`, the message, the failing
   **step name**, the iteration index, `skipped:true`, a timestamp — and increments
   `{loop}_error_count`. It then **persists state to the DB** (`repo.UpdateState`, :185).
2. It logs a `Warn` — *"Skipping failed loop iteration, advancing to next"* — naming
   `failed_step`, `error`, `failed_iteration`, `total_iterations` and `total_errors`.
3. `LoopCompleteAction` (`loop_actions.go:466-472`) reads that key back per iteration and
   stamps `status:"error"` plus the error payload and the item's page name into the loop's
   aggregated result — **deliberately distinguished from `status:"missing"`**, which is what a
   genuinely empty iteration gets.

**The honest residual, which the seat is entitled to:** none of that raises a *work item*, and
`orchestration_states` is retention-clocked, so the trace expires with the row. Whether any
downstream consumer acts on the aggregate's `status:"error"` is per-workflow and
**[UNMEASURED]**. So: durable within the run and typed in the aggregate, but not escalated. A
consumer that ignores the aggregate's status would still lose the page quietly — and that is a
property of the consumer, not of this knob.

### `bug_historian` (medium) — "file the sibling rather than leaving it as prose"

**Done: `bugs_open/193`.** The seat's reasoning is the transferable part — a deferral recorded
only in the deferring document *"will not surface to anyone auditing the mechanism later"*,
because the next reader reads the mechanism, not the document that once touched it. The sibling
is `loop_actions.go:66`: the loop-level read of the very same key silently ignores a non-bool,
and after today's fix its substep-level twin **warns** — so the mechanism is now loud on one
side and silent on the other, which is worse than uniform silence.

Measured while filing, rather than asserted: all **10** live loops declaring `continue_on_error`
declare it as `boolean` (9 `true`, 1 `false`), so the trap is latent, not live.

### `editquality` (low) + a MISSING item — and this one is a genuine submission-craft error

The seat flagged: *"the plan's own rationale requires a concept-register edit alongside the code
change, but the edits array contains only the production file and the test file… Either the
registration edit is missing from the plan, or the plan's own compliance claim is false."*

**The seat is right about the submission and wrong about the commit, and that is my fault, not
its.** WFA-008 *is* in the same commit as the code (`2e497e846`) — but I described the
registration in the rationale and left it out of the `edits` array, so from the reviewer's seat
the compliance claim was unevidenced. The reviewers can only judge what the plan shows them.

**MISSTEP, logged.** The lesson is specific and reusable: **when a submission's rationale claims
compliance with the ordering exemption, put the register edit IN the edits array**, even though
a docs-only submission would be refused for scope. Scope is judged on the whole submission, and
mine already qualified on `platform/`. Adding the docs edit costs nothing and converts an
unverifiable claim into a checkable one.

Its `low` objection — that the warning-and-fallback branch is "slightly beyond minimal", since
`173` only asked for a substep able to declare its own tolerance — is noted and **stood by**:
reproducing the declared-and-inert failure mode inside the fix for that mode would have been
indefensible, and the seat explicitly framed it as "flagging so the extra branch is deliberate,
not accidental scope creep." It is deliberate.

### One thing the seats asked a HUMAN to confirm, recorded so it is not lost

Both `reuse_agent` and `architecture` noted that owner ruling 2026-07-29 §1 is being
**self-applied by the plan's author**: I argued my own change does not cross the RFC threshold.
Both agreed with the reading — `architecture` at length (*"this is the shape the 2026-07-29
owner ruling explicitly carved out of the RFC gate"*) — but flagged that the author is not the
ruling's owner. **Left standing for the owner, not resolved here.** It is in the README.

---

## 2026-08-04 — LIVE on v1.0.1250, pod-verified with three controls; induction seeded and queued

A fresh chassis build rolled. **The fix is live**, and this is the measurement rather than the
inference — `bugs_open/153`'s rule is that a roll is not evidence your fix shipped, so the
image was grepped, on **both** replicas (`agent-chassis-88cf8787-4dzzx`, `-5z5sn`, both
`v1.0.1250`, started 10:29:20Z/10:29:40Z):

```
NEW  resolveSubstepContinueOnError                        = 2   (definition + call site)
NEW  "Substep declares continue_on_error with a non-bool" = 1
CTRL "continue_on_error is true for this loop iteration"  = 1   (pre-existing; the probe works)
NEG  resolveSubstepContinueOnErrorXYZZY                   = 0   (the probe CAN return 0)
```

The negative control is the one that is usually skipped and is the reason to trust the rest: a
`grep -c` that never returns 0 is not measuring anything. All four values are as predicted
before the roll, and the pre-roll reading of the same probes was `0 / 0 / 1 / —`.

### The induction, designed so that a pass cannot be a coincidence

Two throwaway agents (`SEED_2026-08-04_induction_agents.sql`), each setting the **loop**-level
flag to the **opposite** of the substep's:

| agent | loop | substep `boom` | expect |
|---|---|---|---|
| `test-173-tolerant-substep` | (unset) — strict | `true` | COMPLETED, both iterations skipped |
| `test-173-strict-substep` | `true` — tolerant | `false` | FAILED |

That opposition is deliberate: if expansion still clobbered the substep's value, **each run
would produce the other run's outcome**, so a stamp-regression cannot pass as a success.

Two design checks done **before** dispatching, because a fault that does not fire makes the
tolerant branch read COMPLETED for entirely the wrong reason:

1. `QueryDatabaseAction` returns `return nil, fmt.Errorf("query failed: %w", err)` on a SQL
   error (`database_actions.go`), i.e. a genuine Go error that fails the step and reaches the
   coordinator's error path — not a soft error stuffed into a result.
2. `SELECT 1/0` really does raise (`ERROR: division by zero`) and `SELECT 1 AS n UNION ALL
   SELECT 2` really does return 2 rows. Both run against the live DB, not assumed.

### MISSTEP — I suppressed kcat's output on the first two dispatches

I sent the first two runs with `>/dev/null 2>&1` on the `kcat` invocation, to keep the shell
output tidy. **`kcat -P` can send nothing while exiting 0** — that is a recorded landmine on
this estate, and suppressing the output is precisely how you fail to see it. So for those two
orchestration ids, "published" is **not established**, and their absence from
`orchestration_states` cannot be attributed to queue latency with any confidence.

Re-dispatched with output visible; exit 0 and no error, though that is still not a delivery
receipt. **The rule for next time is simple: never redirect kcat's output on this estate.**

### Status at the end of this session — queued, not witnessed

No `orchestration_states` row for any `ind173-%` run yet. That is **not** being read as a
dropped dispatch: CLAUDE.md records publish→start at **29 minutes** under normal load, and the
generic lane is the one `bugs_open/096` is about. The fleet is demonstrably alive (orchestration
updates in most minutes through 10:47Z), so the lane is draining, just not to me yet.

**`bugs_open/173` therefore stays OPEN**, with the fix live and the induction owed. Everything
needed to finish it — the expected results, what would REFUTE the fix, and the cleanup
obligation for the two seeded agents — is in `HANDOFF_2026-08-04_continue_here.md`.

---

## 2026-08-04 — the induction RAN, both branches, and 173 is CLOSED

### MISSTEP 3 — my first two dispatches died on a validation error I never looked for

The first three dispatches produced **no `orchestration_states` row at all**, and I spent a
while treating that as queue latency, which CLAUDE.md explicitly warns is the usual cause and
tells you not to retry on. It was reasonable and it was wrong.

**What actually settled it:** grepping the chassis logs for the orchestration name. The
messages had arrived and been processed within *milliseconds* of publishing — so kcat had
published fine and the lane was not slow at all. Filtering the same correlation to
`level=error` gave the answer in one line:

```
Invalid workflow configuration ... error: step 'done' with action 'complete' requires a topic
```

My seed used `{"action": "complete"}` as a terminal step. The convention is a step **named**
`complete` whose **action** is `complete_workflow` (172 live agents have exactly that). The
validator was right and my definition was malformed.

**The transferable bit — and it is the reverse of the rule I was applying.** "A missing
orchestration row is latency, not a drop" is sound advice *for a valid definition*. It gave me
a ready-made explanation that fit the evidence and cost three dispatches and ~15 minutes. The
cheap check that beats it: **grep the chassis log for your `orchestration_name` before
theorising about the queue.** If the message arrived, latency is refuted outright; if it
arrived and died, the error is right there. A queue theory explains an absent row; it does not
explain an absent row *plus* a log line showing the message was processed — and I never looked
for the second thing.

Also worth recording: I had suppressed kcat's output on the first two sends, so I had
independently made the silent-publish landmine unfalsifiable for myself. Both errors point the
same way — **I removed my own evidence and then reasoned about the gap.**

### The induction, and why its pass is discriminating

Corrected seed re-applied, both agents re-dispatched. Results:

| run | loop | substep | status | current_step | mechanism |
|---|---|---|---|---|---|
| `35acb827…` | (unset) strict | `true` | **COMPLETED** | `complete` | `iter_0_error` + `iter_1_error`, `skipped=true`, `error_count=2` |
| `982bf0ce…` | `true` tolerant | `false` | **FAILED** | `run_loop_iter_0_boom` | died at the failing substep |

Logs confirm both skips by name (`step=run_loop_iter_0_boom iter=0 total_errors=1`, then
`iter=1 total_errors=2`).

**The design property that makes this a proof rather than a green light:** each run's
loop-level flag is the OPPOSITE of its substep's, so if expansion still clobbered the substep
value, **each run would have produced the other run's outcome** — tolerant FAILED, strict
COMPLETED. It is not possible to pass both branches by accident.

And the status alone would not have been enough: the tolerant run reads COMPLETED just as
happily if the fault never fired. The discriminating evidence is the two `iter_N_error` records
with `skipped=true`, which is why I checked the action returns a real Go error and that
`SELECT 1/0` really raises **before** dispatching.

**Cleanup done** — both `test-173-*` definitions deleted, 0 remaining. **173 moved to
`bugs_closed/`.** WFA-008 promoted from *built (inert until roll)* to *deployed*, with the
induction as its status-evidence.

---

## NOTICE 2026-08-06 from the `bugfix_197_transient_classifier` lane — retry semantics of failing substeps change on the next roll

Told rather than merely measured (owner ruling 2026-07-29 §3). You filed `bugs_open/195`
and its sibling `197`; 197 is now fixed in code (commit `1e349d046`, inert until the roll).

**What changes for your lane:** the retryable-side classifier stops being three
case-sensitive needles and becomes typed-first with a case-folded, census-derived fallback.
Concretely: an agent-level processing failure whose prose contains `deadline exceeded` (885
of the 2,996 measured), a capitalised `Timeout`/`Connection`/`Temporary` (882 more), or
rate-limit/5xx prose now classifies `error_recoverable` where it was terminal — so on the
paths where the agentbase response decides (spawned children; no-ResponsesTopic; errors
bypassing `handleError`), **a failing step gets up to 3 redispatches before terminal**
instead of failing on the first attempt. Your `continue_on_error` semantics are unchanged;
what changes is how many attempts precede the failure your flag then handles.

Bounds: `retry_version >= 3` (coordinator) + the 075 adapter cap; no backoff exists on the
recoverable arm (named as a known gap in RSH-006, not fixed). On the orchestrated main path
the processor's typed-only sender still pre-empts — see the 196 lane's notes for the
scheduled convergence.
