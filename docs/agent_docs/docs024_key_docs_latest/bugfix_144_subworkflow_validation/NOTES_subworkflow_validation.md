# NOTES — bugs_open/144 sub-workflow validation

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-29 — picking it up

Checked ownership before starting: `scripts/who-owns.py 144` names only the filing
commit (`cbb08bfb8`) and a docs commit from the filing thread (`396b92a5f`), no owning
workstream. Bug file says "OPEN, unowned". `git log --since=5.days -- platform/validation/`
shows one commit, from the 100/101 lane, not this. Working tree has no edits under
`platform/validation/`. Taken on.

## 2026-07-29 — the measurement, before the design

Exported all 178 live workflow definitions and walked them in python. Numbers in the
PLAN. Two results decided the design before a line was written:

* **All 20 distinct nested actions are `IsLocal: true`, and 0 of the 85 nested steps
  carries a topic.** So the "remote action needs a topic" rule costs nothing today —
  and would have been an instant fleet-wide outage if even one nested action were
  remote. This was the single most load-bearing check of the day and it was one grep.
* **0 nested `(action, key)` pairs trip the strict-config rule.** So the hard-error
  path for a strict action is free today too.

## 2026-07-29 — two things the bug file's fix candidate would have got wrong

Not criticism of the filing thread — both live in the executor, which the bug file
correctly points at but does not quote.

1. **`next_step` out of a sub-workflow is LEGITIMATE.** `coordinator.go:4009-4014`
   passes an external reference through untouched, and
   `loop_expansion_handler.go:192` *injects* `<loop>_complete`, which exists in no
   definition. "Validate each sub-workflow as a workflow in its own right" would make
   both a hard error. Live count of nested steps whose `next_step` leaves the
   sub-workflow: 0 today — so this would have passed the fleet measurement and still
   been wrong, breaking the first workflow anybody wrote that way. **A measurement
   that comes back zero does not tell you the rule is right; it tells you the rule is
   not currently firing.**
2. **`parseSubsteps` drops fields.** `dependencies` is read by the top-level decoder
   and NOT by the nested one. Validating a nested `depends_on` for existence would
   have been enforcing a field the executor discards. Reported as dropped instead.

## 2026-07-29 — misstep: I nearly measured the wrong flag

For "would any nested step hard-fail the strict-config rule?" I measured against
`opted_in` from `config-key-audit --specs`, which is `CheckConfig || len(ConfigKeys)>0`
— NOT `StrictConfig`, which is what `IsStrictConfigAction` actually gates on. I only
noticed when a test failed for registering `CheckConfig: true` and expecting a hard
error.

The measurement survives because the direction is conservative: `opted_in` is a
superset of strict, and the answer was 0 under the superset. **But I had written "0 of
the 66 nested pairs trips the strict rule" before checking which flag I had measured**
— the right answer for the wrong reason, which is indistinguishable from the wrong
answer until someone checks. Corrected in the PLAN to state the population precisely.

## 2026-07-29 — misstep: two tests that could not fail, caught by the compiler and by a control

* `dropped_fields` came back empty because zaptest's `ContextMap()` renders
  `zap.Strings` as `[]interface{}`, and my `.([]string)` assertion silently yielded
  nil — a type assertion with the comma-ok form swallows the failure. The test then
  asserted `len(got) != len(want)` and failed for the right reason by luck. Fixed to
  `[]interface{}`.
* Both new warning tests were written with a positive-control half (a step using only
  honoured fields must report NOTHING). Without it, a warning that fired on every step
  would have passed the first half exactly as well.

## 2026-07-29 — the instrument found something that is not mine

The dry-run reported 3 live definitions rejected. **All three pre-existing** —
`html-developer-chunked`, `multipage-wrapper`, `html-assembler`, each with a top-level
step naming an action that is in **no registry** (`assemble_html_parts`,
`wrap_multipage`, `assemble_full_page`) and carrying no topic, so `ValidateWorkflow`
rejects them on every message.

`bugs_closed/044` already records this family as "retired code still flagged
`is_active=true`" and explicitly scopes the `is_active` hygiene half out as an owner
decision. What 044 does NOT say, and what I measured:

* three LIVE builders still point at two of them — `website-builder`,
  `landing-page-builder`, `content-site-builder`, via `spawn_agent`/`call_agent` with
  `config.agent_type` = `multipage-wrapper` / `html-assembler`;
* zero orchestration rows for those three agent types in the retention window
  (oldest row 2026-07-13, 1,952 rows total). So either those dispatch paths are never
  taken, or they are taken and die — **[UNMEASURED] which**, and the distinction
  matters.

Recorded here and in the bug I filed rather than fixed in this change: it is a
different defect (an action name that exists in no registry is undetectable until a
message arrives) and fixing it inside a validation patch is exactly the scope mistake
the guardian seat vetoes.

## 2026-07-29 — HEAD does not compile, and it is not mine

`go build ./cmd/...` fails at HEAD: `cmd/reasoningset/main.go:504: declared and not
used: planJoined, planMissing, provJoined`. Confirmed against a clean
`git archive HEAD` extract, and `git status` shows no session holding a fix in the
tree. **It does not block image builds** — `build/docker/backend/agent-chassis.dockerfile`
compiles `./cmd/agent-chassis` only — but it does break `go build ./...` and `go vet
./...` for every session, which is the check we all use to answer "did I break HEAD?".
Left alone (another session's in-flight tool, forward-only, not mine to rewrite) and
reported to the owner. My own HEAD check was scoped to `./platform/... ./cmd/config-key-audit`
accordingly, and that is stated rather than implied.

## 2026-07-29 — submitted, and committing rather than waiting

Council submission corr `9194bc97-8475-4022-b658-2ac64f06dd63`. Committed the code
immediately without a trailer, per the standing practice: waiting for a verdict does
not protect anything on a shared tree — it just exposes finished work to the next
`git add -A` while forfeiting attribution. Trailer is earned by APPROVED only.

## 2026-07-29 — council: APPROVED round 1, and two of the four objections were right

Corr `9194bc97-8475-4022-b658-2ac64f06dd63`. 4 objections, none high-severity, 5 seats
abstained. Taken one at a time, because "approved" is not a reason to skip reading them.

**1. `reuse_agent` (medium) — RIGHT, and answered with code.** *"DecodeSubWorkflowStep
is a hand-written second implementation of parseSubsteps' decode contract, pinned
together only by a lockstep test rather than by sharing one function. The plan's own
rationale for the founding incident is exactly this pattern."* It also noted, fairly,
that I asserted a package-boundary justification without showing the alternative had
been tried.

It had been considered and dropped for the wrong reason — I did not want to touch the
runtime decoder inside a validation fix. But extracting an identical field-by-field
decode is not a behaviour change, and "a test proves the two copies agree" is precisely
the shape of guarantee this bug exists to distrust. So: the decode moved to
`pkg/models/substep_decode.go`, `parseSubsteps` now calls it, the validator calls it,
and the lockstep test is honestly relabelled as a **re-inlining guard that cannot fail
today**. The guarantee that CAN fail moved to `pkg/models/substep_decode_test.go`: add
a field to `models.Step` and it fails until you say, in code, whether the loop reads it
or knowingly drops it.

**2. `debug_historian` (medium) — RIGHT to ask, and the answer holds.** *"Nothing
enumerates whether `is_active` actually gates which definitions the processor dispatches
against — this is exactly the `sites.status='active'` shape the lore warns about."* I
had scoped the whole measurement by `is_active` without checking. Checked both ways:

* `processor.go:357-360` filters on `is_active = true AND deleted_at IS NULL AND
  (is_snapshot IS NULL OR is_snapshot = false)` — the same three predicates the export
  used, so the population is exactly what can be dispatched;
* and re-ran the dry-run over **all 183** non-deleted, non-snapshot definitions
  regardless of `is_active`: same answer, 0 newly rejected, same 3 pre-existing.

Answering it produced a **correction to my own submission**: I named
`platform/agentbase/agent.go` as a consumer of the changed guarantee. It is not.
`ValidateWorkflow` has exactly ONE production call site — `processor.go:276` — plus the
`validation.Validator` wrapper; agentbase constructs that wrapper for
`ValidateIncomingMessage` and never calls `ValidateWorkflow`. Blast radius is narrower
than I claimed, and I claimed it in the direction that flatters the submission.

**3. `editquality` (medium ×2) — right about the SUBMISSION, wrong about the code.**
It objected that nested cycle detection is "asserted in a trailing comment with no
corresponding code shown" and that no test exercises a cycle among nested steps. Both
were already in the change before submission — `validateSubWorkflow` calls
`checkForCycles` on the nested plan, and `TestNestedCycleDetected` pins A→B→A inside a
sub-workflow. What was missing was in the **sketch**: I summarised six checks in prose
and showed three in code. A reviewer can only review what is in front of it, and a
sketch that lists what it omits is not the same as showing it. No code change; recorded
because the lesson is about the submission, not the fix.

**4. `guardian` (medium) — OPEN, and it is not mine to close.** Two asks: (a) a human
should settle the RFC-vs-bug-patch venue question that my own risk #4 raised, and (b)
confirmation that no consumer relies on today's silent pass-through as a feature. It
also floated a warn-first period instead of shipping straight to hard-reject. Recorded
as open and put to the owner; per CLAUDE.md a scope/venue judgement is broken by a
human, not by a thread resubmitting with better measurements. Note the shape: the
guardian is not disputing the measurement, it is disputing whether a measurement is the
right kind of answer to the question.

`prior_art_librarian` and `architecture` approved with low-severity notes (verify the
"second binary" precedent — it is recorded in `cmd/config-key-audit/main.go`'s own
header; and note `WalkSteps` in the register so a third consumer does not hand-roll a
third traversal — done, WFA-003).

## 2026-07-29 — LIVE on v1.0.1203, closed — and my delete-marker was not a delete-marker

Rolled and verified on **both** replicas (`sha256:afd2a4683362…`):

```
"uses fan_out, which cannot work inside a sub-workflow"   1   (new)
"Substep declares fields"                                 1   (new, round 2 specifically)
"Checking disconnected step: "                            0   (DELETED — the real marker)
"Checking disconnected step for cycles"                   1   (its replacement)
"they are silently ignored at execution"                  1   (control: untouched 101 string)
```

**MISSTEP, and it escaped this workstream before I caught it.** The marker I wrote into
the bug file, the runbook and the register was `"Checking disconnected step"` with an
expected count of **0**. On a correctly deployed pod it returns **1** — because the
replacement Debug message *contains the old phrase as a prefix*
(`"Checking disconnected step for cycles"`). Following my own runbook would have read a
correctly-deployed image as not deployed.

It did not stay mine. Another session's `bugs_open/153` picked it up as a **positive
control** and drew the inference *"⇒ 144's pre-fix code IS what is running"* — which the
grep cannot support in either direction. Chasing that back produced a larger finding for
their lane, contributed to their file rather than filed separately: **all five of 153's
marker strings are structurally incapable of appearing in any binary** — four exist in no
`.go` file in any commit (they are phrases from the 138 and 104 workstreams' own README
and RUNBOOK prose, plus one line of site copy), and the fifth sits inside a Go comment.
Their `validation.WalkSteps` marker is a symbol name, which greps 0 on an image I had
just proven contains it.

**The rule, stated so it survives:** a marker must be a string the binary **EMITS** —
not a symbol name, not a comment, not a sentence from your own workstream docs. A
*delete*-marker must additionally be a string the new code cannot contain **as a
substring**. Cheapest possible check, before running anything against a pod:
`git grep -c "<marker>" -- '*.go'` at the commit you expect to be running. Mine would
have returned 1 for a phrase I believed I had deleted; theirs would have returned 0 for
four phrases that were never code.

**Functional proof, which no binary grep can give:** 22 orchestration runs whose
`workflow_plan` carries a `sub_workflow` between 18:01Z and 18:49Z — 21 COMPLETED, one
mid-flight at `process_item_iter_2_spawn_handler`, a **loop-expanded substep**, and **0
validation errors** fleet-wide. The nested traversal ran in production against real
definitions and passed them, which is the thing the 0-newly-rejected dry-run predicted.

Bug moved to `/bugs_closed/`. Register WFA-003 → deployed. The RFC-vs-bug-patch venue
question stays open for the owner; it is not a defect and does not hold the bug open.
