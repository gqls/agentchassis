# NOTES — bug 124, double dispatch. Append-only, newest at the bottom.

## 2026-07-28 — picking the bug up

`scripts/who-owns.py` across the `bugs_open/` list. Most numbers came back
"OWNED or recently active" with a named workstream. 124 came back with one
commit (its own filing) and no owning workstream; the filing thread
(work-item-parallelisation) wrote *"Status OPEN, unowned. Separable from 029 and
cheaper"* and MEMORY.md now records that thread's programme as complete. Taken up.

## The filed mechanism did not survive contact

The bug offered two mechanisms. I checked both before building anything.

**Filed item 1, `[VERIFIED]`: "Nothing marks a needs_diagnosis item complete on
success."** REFUTED.

```
SELECT jsonb_object_keys(default_config->'workflow'->'steps') FROM agent_definitions
WHERE type='diagnose-dispatch-loop' AND is_active AND COALESCE(is_snapshot,false)=false;
 → complete, claim_item, reap_stuck, mark_failed, call_handler,
   check_claimed, mark_complete, spawn_handler, notify_scheduler
```

`mark_complete` is a `complete_work_item` step and it works — every 090-filed item
sits at `complete` or `failed`, `claimed_by='diagnose-dispatch-loop'`. The
`[VERIFIED]` marker was attached to an inference drawn from a **print statement in
the shell script** ("closing it by hand until a diagnose dispatch loop exists").
That is the failure mode the marker exists to prevent, wearing the marker.

**My own near-miss, recorded because it is the same shape:** I nearly built the
fix off `0NN_diagnose_dispatch_loop.sql`, the seed file, which is a beautifully
documented 250-line account of a config that has since drifted. I read the live
row instead because the standing rule says to. The rule is normally stated about
seeds; it applies just as hard to *comments and headers*.

**Filed item 2, `[UNVERIFIED]`: "What started chain B is not established… could be
the retry driver, a second claim, or a separate direct dispatch."** It is the
third: **the 090 trigger's own `kcat -P` publish**, at the bottom of
`090_TRIGGER_needs_diagnosis_v1.sh`. The script writes the item at
`awaiting_diagnosis` *and* publishes the envelope. The loop claims
`awaiting_diagnosis` items. Two dispatchers, one queue.

The script's own header names the invariant it is breaking:

> *"The task ships DISABLED — enable it deliberately. Until then, and for any
> ad-hoc run, THIS SCRIPT is the dispatcher."*

`scheduled_tasks.enabled = t` for `diagnose-pipeline-trigger`. Someone enabled it.
The script was never told.

## The shape, in the data

```
corr 914dc844  (= item's spec.correlation_id)  orchestrator 07:03:04, agent 07:03:20   ← the script's publish
corr c89c718c  (fresh)      dispatch-loop 06:59:39, orchestrator 06:59:53, agent 07:00:08   ← the loop
```

Chain **without** a `diagnose-dispatch-loop` row = the script. Chain **with** one
= the loop. 6 of 6 items inside the retention window have both. Each duplicate is
a 12–14 minute `diagnose-agent` run (longest observed 31m).

## A number I refused to write down

I started to write "every `needs_diagnosis` item ever filed ran twice". The query
that would have supported it counts orchestrations over all history — and
`orchestration_states` is on a retention clock, so the 15 items older than a few
days show **0** chains, not 2. Zero there means "the rows were reaped", not "it
did not happen". The honest claim is the windowed one: 6 of 6 since 2026-07-27
19:04, which is as far back as the evidence reaches. (Standing landmine: every
history table is on a retention clock — record a rate, not a count.)

## Correcting the neighbouring bug's evidence

`029` §6 and `124` both cite orchestration `41d64b75` as the loop *"re-dispatching
an already-diagnosed item, 43 minutes after that diagnosis finished"*.

```
SELECT orchestration_id, correlation_id, owner_agent_type, created_at, updated_at
FROM orchestration_states WHERE orchestration_id::text LIKE '41d64b75%';
 → 7803075d… | diagnose-dispatch-loop | 2026-07-27 20:08:16 | 2026-07-27 20:52:32
```

Created **20:08:16** — 91 seconds after the item was created at 20:06:45. It is
the concurrent duplicate, not a later re-dispatch. What was sent at 20:49:31 was a
**retry** of its `call_handler` request, which is `029` proper. The *conclusion*
survives intact and gets stronger: the duplicate chain failed and wrote `failed`
over an item whose other chain had returned a REFUTED verdict, so counting
`failed` needs_diagnosis rows over-counts `029`. Only the story of *how* is wrong.

Both bug files get this correction written back into them.

## The throughput objection — checked, not assumed

Making the loop the sole dispatcher puts every diagnosis behind a task with
`max_concurrent = 1`. That reads like fleet-wide serialisation behind one 13-minute
run, which would have been a real reason not to do it.

It is not what happens. `cmd/scheduler/main.go:287` calls `stampCompleted`
**immediately after publishing** — fire-and-forget, both timestamps advanced — and
`countInFlight` counts *rows in `scheduled_tasks`*, not running orchestrations. The
slot frees on the next tick. Observed overlap in the live data: dispatch-loop runs
`04920015` (11:25:51→11:38:57) and `2184add8` (11:33:22→11:46:35) ran concurrently.

Dispatch latency *improves*, too: loop start was +55s / +40s / +79s after intake,
against +4m19s for the script's own publish on the same item — the direct publish
queues behind the shared generic requests lane, the scheduler tick does not.

## The half of the bug nobody had noticed

`diagnosis_artifacts` rows are keyed on `params.ExecutionContext.CorrelationID` —
the **envelope** correlation — not on the `correlation_id` the loop passes down
through `input_mapping`. So for an item the loop dispatched, `spec.correlation_id`
names **nothing**. It resolves today only because the script's duplicate chain
happens to run under it.

Both `diagnosis-triage`-created items have a `spec.correlation_id` minted by
`triageSpecJSON` that no run ever used. Those two were never double-dispatched —
and they are the ones whose trail is completely unjoinable. The two halves of the
fix are one change: kill the direct publish without fixing the join and the
printed `SAVE: CORRELATION_ID=…` stops resolving for everybody.

## 2026-07-28 later — build

Three parts, ordered by what closes the door: the claim becomes the ticket to
dispatch (P1), the script asks the DB who the live dispatcher is rather than
assuming (P2), and `query_database` gains a generic `$ctx.` parameter namespace so
the loop's claim can stamp the run's own correlation onto the item (P3). Details
and the declined alternative (correlation override on `call_agent`) in the PLAN.

## 2026-07-28 16:39 — caught in the act, in real time

The 097 submission's own queue report listed two diagnose chains in flight, so I
looked. Another session filed `needs_diagnosis:validate-page-content-repairs-dead-in-bo`
at 16:37:51 and both dispatchers took it:

```
16:37:58  diagnose-orchestrator   corr 954d8da9   ← = the item's spec.correlation_id: the SCRIPT's publish
16:38:12  diagnose-agent          corr 954d8da9
16:38:28  diagnose-dispatch-loop  corr 2a656f25   ← fresh: the LOOP. claimed_at 16:38:28
16:38:45  diagnose-orchestrator   corr 2a656f25
16:38:59  diagnose-agent          corr 2a656f25
```

Two `diagnose-agent` pods, both at step `verdict`, on the same symptom, 47 seconds
apart. Whatever those two return, one of them is money we did not need to spend —
and the item's `spec.correlation_id` names only the script's chain, so the loop's
chain is invisible from the item.

This is the specimen the bug file asked for and it arrived unprompted while the
council was reviewing the fix. Not touched: the standing landmine says the
duplicate chains ARE the evidence — do not cancel them to tidy the queue.

Also worth recording: it is *someone else's* diagnosis. This is not a cost we pay
when we go looking for it; it is a cost the fleet pays on every intake.

## 2026-07-28 17:15 — I killed my own council round with my own deploy

Council round 1 (`90361922-e4c4-482e-a0b7-b1a49640265a`) stopped at
`review_diagnosis_guardian`, `updated_at 16:55:46.936`. The replacement chassis
pod for v1.0.1191 started at **16:55:47**. Same second.

`awaited_requests` is `{}` — the council was executing a seat INLINE on the
chassis, not awaiting a child, so there is no outstanding request for a retry
driver to recover. The round died with the pod and nothing will resurrect it.

**The lesson, which I should have seen coming and did not:** the council gate
runs its seats inline on the chassis request lane (that is the whole premise of
`bugs_open/096`, which I had read). So `kubectl rollout restart deployment/agent-chassis`
kills any council round in flight — including, with a certain symmetry, the one
reviewing the change you are deploying. There is no warning and the orchestration
does not go FAILED; it sits at `EXECUTING_STEP` looking alive.

Sequence it the other way: **get the verdict, then roll.** Or, when the image
genuinely must go first (as here — migration 258 could not be applied against an
older binary without stopping the diagnose lane), submit AFTER the roll.

Resubmitting on the same correlation with `RESUBMIT_CORR` so the trail
accumulates rather than starting a fresh, disconnected round.

## 2026-07-28 17:11 — verified, end to end, against a real run

`SLUG=scheduler-stamps-completed-at-publish`, fired 17:04 on chassis v1.0.1191
with migration 258 live. Every clause of the bug file's "How to verify a fix":

```
site_work_items: status=complete, completed_at=17:11:20, claimed_by=diagnose-dispatch-loop
                 intake_corr ae10e615…   run_corr 66a65287…
orchestrations under ae10e615 (intake corr) : 0      ← the script published NOTHING
orchestrations under 66a65287 (run corr)    : diagnose-dispatch-loop 1
                                              diagnose-orchestrator  1
                                              diagnose-agent         1   ← ONE, was two
diagnosis_artifacts under 66a65287          : bundle 2
```

- exactly one `diagnose-agent` orchestration ✓
- item reached a terminal status with no hand-written `UPDATE` ✓
- `spec.dispatch_correlation_id` equals the correlation the chain actually ran
  under, and the artifacts are under it ✓ — so the item→run→artifacts join now
  holds for the automatic path, which it never did before, not even in principle

Deliberately verified against a **real** diagnosis, not an idle window (029's
note: the fleet reports zero when nothing is running). The symptom chosen was a
genuine open question rather than a throwaway — whether
`scheduled_tasks.max_concurrent` governs anything for workflows whose work
outlives the publish — so the credits bought a finding as well as a proof.

Cost note for the record: this run cost ONE diagnosis. Before today it would have
cost two, and one of the two would have been invisible.

## 2026-07-28 17:2x — what the verification run actually found (REFUTED, and that is a success)

The symptom I paid for came back **REFUTED**, with a cited trail. It independently
confirms the throughput reading I had made by hand — which is worth more than my
own reading of the same file, because I wrote both.

> *"There is no `notify_scheduler` contract and no race … the scheduler's own
> comments state the actual, deliberate contract is fire-and-forget — 'we don't
> wait for the orchestration to finish' — and `stampCompleted` synchronously
> advances both `last_triggered_at` and `last_completed_at` BY DESIGN, precisely
> so `countInFlight`'s window closes immediately and a fired task doesn't pin its
> group."*

Two things it adds that I did not have:

1. **`max_concurrent`'s only read sites are inside `cmd/scheduler/main.go` itself**
   (`loadDueTasks`' SELECT and `runTick`'s comparison), plus
   `internal/core-manager/admin/pipeline_admin_handlers.go:HandleListPipelines`
   which only *displays* the column. **No orchestrator workflow consults it.** So
   `max_concurrent` means "do not fire this task twice inside one scheduler tick"
   and nothing more — it is not, and has never been, a limit on how many runs of a
   lane are in flight. Anyone reading `max_concurrent=1` as "one diagnosis at a
   time" is reading a guarantee that does not exist.
2. It flags the loop's own `notify_scheduler` step (`UPDATE scheduled_tasks SET
   last_completed_at = NOW()`) as writing a value the scheduler already wrote at
   publish. Harmless, but it is dead config that *reads* like a completion
   handshake — and reading it that way is exactly how I nearly talked myself out
   of this fix. Not touched here: out of scope, and worth its own look.

Recorded rather than acted on. The verdict is REFUTED, which is the cheapest place
to be wrong and the reason the symptom was worth a real run.

## 2026-07-28 17:18 — P1 was BROKEN, and only testing the failing branch found it

I had verified the default path (loop live → intake only) against a real run and
it was clean. That proved nothing about **P1**, the guard that is supposed to stop
a direct publish when someone else already holds the claim — and P1 is the part
that actually closes the door. So I staged the failure: inserted a
`needs_diagnosis` row already at `diagnosing`, `claimed_by='pretend-other-dispatcher'`,
and ran the script with `DISPATCH=1`.

**It dispatched.** A real orchestration went out
(`902e981a-d3d5-4219-9c9c-f8685d1cf992`) against an item another dispatcher owned
— the exact thing the fix exists to prevent.

The cause, and it is a good one:

```
$ printf '%s' "UPDATE … WHERE status='awaiting_diagnosis' RETURNING id::text;" | psql -t -A
UPDATE 0
```

**`psql -t -A` still prints the command tag on a non-SELECT.** `-t` suppresses the
header and footer *of a result set*; it does nothing to the status line of an
`UPDATE`. So a claim that matched **zero rows** returned the eight-character
string `UPDATE 0`, my `if [ -z "$CLAIMED_ID" ]` test saw a non-empty value, and
the guard concluded it held the claim. The happy path had worked perfectly and
concealed it, because there the RETURNING row *does* come back and the tag is
suppressed by `-t` for the SELECT-shaped output.

Two fixes, because one of them is not enough on its own:

1. **Wrap the UPDATE in a CTE and SELECT from it.** Then it is a SELECT, and no
   rows genuinely means no output. Verified both ways against the live DB: `[]`
   length 0 on no match, `[a24c692c-…]` length 36 on a match.
2. **Assert the SHAPE, not presence.** The guard now requires a uuid. Controls
   run: uuid ACCEPT, `UPDATE 0` REFUSE, empty REFUSE, and `UPDATE 1` REFUSE — that
   last one matters, because it is what the tag would be if a row *had* matched,
   and a presence test would have accepted it as an id.

Re-ran the staged failure afterwards: `NOT DISPATCHING: could not claim the
intake`, exit 1, nothing published.

**What this cost, honestly:** one junk diagnosis run
(`902e981a`, symptom "claim branch test") which I left to finish rather than
cancel — cancelling an in-flight orchestration has wedged this lane before, and a
wasted 13 minutes of LLM is the cheaper of the two risks. Test row `a24c692c`
cancelled with its purpose recorded in `error`.

**The lesson is the standing one, and I nearly skipped it.** A green happy path
proves deployment, not correctness — induce the fault. I had *already written*
"verify the failing branch" into the plan for the council, and still nearly shipped
on the happy path alone because the happy path had been so convincing. Two hours
of live verification on the good case would not have found this; five minutes of
staging the bad case did.

Generalisable beyond this script: **`psql -t -A` into a shell variable is safe for
SELECT and unsafe for anything else.** Any `UPDATE/INSERT/DELETE … RETURNING`
whose result is tested for emptiness has this bug. Worth a grep across the trigger
scripts — not done here.

## 2026-07-28 18:2x — v1.0.1192 rolled by another session; the landmine held

Someone deployed a fresh chassis. This is the first live test of 124's standing
invariant, so I checked it rather than assumed:

```
pods    agent-chassis-f757fcf65-{bg9t7,kctnr}  v1.0.1192  digest sha256:4bd9a111…
grep -c "unknown execution-context field"      → 1 on BOTH pods
replicas                                        2/2      (the overlay fix held)
agent_definitions.image_tag                     186 rows @ v1.0.1192
needs_diagnosis awaiting_diagnosis|diagnosing   0
```

Both pods carry the `$ctx.` binding, so migration 258 is still satisfied and the
diagnose lane is safe. The build came from committed HEAD, which now includes the
fix — which is exactly why the committed-ref build rule exists.

Worth noting what was actually at risk: had that build come from a ref predating
`af0cde87d`, `claim_item` would have failed on the next tick and the diagnose lane
would have stopped dispatching, silently, with no failed row to look at. **A
pod-grep after every roll is not ceremony for this lane; it is the only thing
standing between a routine deploy and a dead queue.**

Also confirmed the 2-replica overlay fix survived a deploy I did not run — which
was the point of moving it out of `kubectl scale` and into the overlay.

## The psql trap had been found TWICE before, and written down nowhere findable

Ran the audit I said was worth doing:

```
docs024/idea_uk_vm_site/054_chrome_verify/02_verify_054_induced_fault.sh:54
  "INSERT then SELECT (not RETURNING): psql -tA prints the 'INSERT 0 1' command tag
   on its own line alongside a RETURNING value, which pollutes a captured id."

docs024/webdesign_couk/scripts/watch_park_webdesign.sh:90-93
  "psql prints its command tag ('UPDATE 0') alongside the RETURNING rows even
   under -t -A, so filter it out. Logging it made every idle cycle look like a
   park and would have hidden a real one in the noise."
```

Two threads, two independent discoveries, two different local workarounds (avoid
`RETURNING` entirely; filter the tag out of the stream). **Neither reached 016b or
`WRONG_CALLS`.** So I hit it a third time — and my instance was the dangerous one:
theirs polluted *visible output*, mine made a *guard silently pass* and dispatched
a paid job against another dispatcher's item while reporting success.

Now in 016b §9 with both halves of the fix. The lesson is not about psql. It is
that a fix recorded in the script where it was needed is invisible to the next
thread, and the cost of that invisibility scales with how much worse the third
instance happens to be. Same class as `bugs_open/106` (the register is
two-thirds uncovered), reached from a different direction.

No live instances of the bug found elsewhere: no other script captures a non-SELECT
into a variable and tests it for emptiness.

## 2026-07-28 evening — the fleet-wide audit (handoff item 2c) — RUN. 0 live, 1 latent.

The 016b §9 entry claimed this shape was fleet-wide. That was an inference from
how the lanes grew, not a measurement, so I measured it.

**Method.** The 124 signature is precise: a script that **inserts a
`site_work_items` row AND publishes an orchestrate envelope**, where something
else can claim that row. Both halves matter — a script that only pokes a *loop*
is benign (the loop's claim is atomic, so two ticks race harmlessly), and a
script that only inserts is the correct shape.

```
23 enabled scheduled_tasks;  scripts inserting a work item AND publishing:  3
```

**The three, judged individually:**

| script | verdict |
|---|---|
| `090_TRIGGER_needs_diagnosis_v1.sh` | the known bug — **fixed today** |
| `180_adoption/080_trigger_adoption.sh` | **NOT an instance.** It inserts an `evaluate_tools`/`tool-suggester` row and separately publishes to `site-adoption-agent` with a `{domain,url}` payload. Two different pieces of work in one script; the inserted row's dispatch by the build loop is the designed path. |
| `092_TRIGGER_experience_plan.sh` | **LATENT instance.** Inserts `needs_experience_plan` at the private status `awaiting_experience_plan`, then publishes to `experience-planner` itself. Safe *today*: no `agent_definitions` row references that status and no scheduled task targets an experience agent — both checked, both empty. |

**So: zero live instances beyond 124.** The §9 entry's "fleet-wide" framing was
right about the *shape* and wrong to imply live prevalence — corrected by this
measurement rather than left as an unchecked claim.

**But the pattern is a repeated design, not a one-off**, which is the finding that
justifies the entry. `report-dispatch-loop` exists in `agent_definitions` as a
**direct clone of the diagnose loop** — same `claim_item` shape, same
`FOR UPDATE SKIP LOCKED`, same private-status pair (`awaiting_report` →
`reporting`) — and its scheduled task `report-dispatch` ships **DISABLED**. That is
precisely the configuration the diagnose lane was in on 2026-07-09. It is *not* at
risk, because `report_request` rows are created by a Go action
(`report_request_pull_action.go`), not by a publishing script — so that lane has
one dispatcher by construction. Worth knowing that the safe lane is safe for a
structural reason, not by luck.

**Action taken where it will actually be read.** A warning block now sits in
`092_TRIGGER_experience_plan.sh` immediately above its INSERT — naming both
remedies, the 016b §9 pointer, and the `psql -t -A` trap for whoever implements
the claim. Standing practice: put the check where the error is MADE, not only in a
doc nobody opens at the moment of the mistake. That is the same discoverability
failure that let the psql trap bite three separate threads today.

## 2026-07-28 20:5x — third roll (v1.0.1194), invariant still holds

Another session rolled again. Same check, same result:

```
pods    agent-chassis-74dbd9c9f4-{7p6d8,rxb52}  v1.0.1194  digest sha256:8013878b…
grep -c "unknown execution-context field"       → 1 on both
replicas 2/2      needs_diagnosis stuck >90m: 0
```

Three consecutive rolls by three different sessions, none of whom knew about
migration 258, and the lane survived all three. That is not luck — it is the
committed-HEAD build rule doing its job: every build since `af0cde87d` structurally
contains the fix, so the only way to break this is a deliberate `REF=` to an older
commit or a rollback. Worth stating plainly, because "it held three times" could
otherwise be read as evidence the landmine is theoretical. It is not; the
protection is upstream of the check, and the check is what tells you when the
protection stopped applying.

Replicas held at 2/2 for the third time as well — the overlay fix continues to
survive deploys done by others, which was the entire point of moving it out of
`kubectl scale`.
