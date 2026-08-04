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
