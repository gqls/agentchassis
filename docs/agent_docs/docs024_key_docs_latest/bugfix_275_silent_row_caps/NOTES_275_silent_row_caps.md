# NOTES — bugs_open/275, the tool-suggester cap and the silent-cap class

Append-only, newest at the bottom. Missteps are the point, not an appendix.

## 2026-08-16 — picking it up

Taken after `bugs_open/257` went live. Ownership checked two ways, because one is not enough:
`who-owns.py 275` names `webdesign_uk_build_service` — but that is the lane that FILED it, not a fixer;
and a grep of ~25 live session transcripts (the only instrument that sees uncommitted work) shows no
session on 275. Actively worked elsewhere at the time: 251, 270, 153, 253, 286, 284, 287, 083, 277,
072, 145, 283, 285, 289, 225, 288, 271, 268, 278, 149, 177, 098, 248, 113, 122.

## 2026-08-16 — validity, and two refinements the bug did not make

Still valid and **worse**: 74 masters now (68 at filing), so **44 hidden, not 38**. The bug predicted
exactly this decay and it took two days to show.

**Refinement 1 — the 406 gate is not the cause and nobody had checked.** The bug reasons about
`LIMIT 30` against the whole library, but 406 added a `requires-backend` filter that runs *before* the
cap. Measured: exactly **1 of 74** masters carries the tag, and **3 of 40** sites have the capability.
So the gate narrows almost nothing and the cap does all the hiding. Worth checking rather than
assuming, because if the gate HAD been narrowing heavily, the fix would have been a different one.

**Refinement 2 — the cap is not arbitrary, and "just remove it" would have been the wrong fix.**
74 rows is 37,209 chars (~9-10k tokens) of prompt. The bug's own candidate 1 anticipates this: *"if the
real constraint is prompt size, cap by TOKENS at the prompt assembly, not by row count in the dark."*
The data names the knob:

| column | chars across 74 rows |
|---|---|
| **`description`** | **29,832 — 80% of the payload** |
| `id` | 2,664 |
| `display_name` | 2,100 |
| `function` | 1,828 |
| `category` | 785 |

`description`: median 374, mean 403, max **2,526**; 50 of 74 exceed 200.

So: bound `description`, not coverage.

| variant | rows | chars |
|---|---|---|
| TODAY — 30 rows, description uncapped | 30 (41%) | 16,421 |
| **THIS — all 74, description ≤ 200** | **74 (100%)** | **20,376** |
| all 74, description ≤ 300 | 74 | 25,146 |

**The whole library for +24%.** I checked 200 for MEANING as well as size by reading the first 200
chars of the longest descriptions — they still say what the tool IS, which is what a relevance
judgement needs. And I read the prompt template rather than assuming its shape:
`- {{.display_name}} ({{.function}}, id: {{.id}}): {{.description}}`.

⚠ **`category` is selected and never rendered** — 785 chars of dead payload. Left alone deliberately:
dropping a column a future consumer might read is scope creep with non-zero risk for a 2% saving.

## 2026-08-16 — the class, which is where the framework leverage is

275 is one instance of: **a row-count `LIMIT` feeding an LLM prompt is silent by construction.** The
model returns plausible output whether it saw 30 rows or 74, so there is nothing to notice.

Census over live `agent_definitions`: **26** literal LIMITs in query-shaped step configs. **19 are
`LIMIT 1`** — the fetch-one/claim-one idiom, by design. The **seven multi-row caps** that can bite:
30 (this bug), 15, 12, 10, 5, 5, 2.

**All 26 run through one function**, `QueryDatabaseAction`. So the detector goes there and the whole
class becomes visible at once, including caps nobody has written yet. That is Part A.

## 2026-08-16 — MISSTEP 1: a mutation survived because the two forms were equivalent

**M3**: changing `rowCount == n` to `rowCount >= n` did not fail a single test. A driver cannot return
more rows than its own `LIMIT`, so for every input reachable from `QueryDatabaseAction` the two are
identical — a genuine **equivalent mutant**, not a test gap.

I did not simply record it and move on, because "equivalent today" is a statement about the current
caller, not about the function. Switched to the defensive `>=`, put the reasoning in the source, and
added a test pinning the arm (a count *beyond* the cap must still report). Without that test the choice
between `==` and `>=` is untested and reads as arbitrary to the next person.

## 2026-08-16 — MISSTEP 2: the helper was right and nothing proved it was USED

**M7**: disabling the branch in `QueryDatabaseAction` — so the warning could never fire — left the
**entire suite green**. Seven tests proving `resultHitItsRowCap` computes the right answer, and not one
of them noticed the action never calls it.

This is the estate's own lesson, which the provocation lane already wrote down for a different helper:
*"the helper being correct and the helper being USED are independent facts."* I had read that sentence
this week, in `provocation_generator_action_test.go`, while working 257. It did not stop me.

**Fixed with a CALL-SITE test**, and deliberately a behavioural one — sqlmock drives the real action, a
`zap` observer asserts exactly one warning naming the step, plus a negative control that an under-full
result stays silent. Not a source scan: register **OPP-003** records a source-scanning detector that
examined zero files and printed a clean result, and *"a clean result and an unrun check are
byte-identical output."*

Final tally: nine mutations, all killed. Two of them changed the code.

## 2026-08-16 — reuse checked BEFORE writing, this time

The `reuse_agent` seat gated `bugs_open/257` round 1 on exactly this — a helper written without
checking what already existed. So before writing a line: `grep` over `platform/` for LIMIT parsing or
result-cap detection. Nothing. The `Truncate*` family is string truncation, `markTruncated` is LLM
`stop_reason` handling, `LimitedRead` is HTTP body bounding, `isTruncatedJSON` is JSON well-formedness.
Checked, not assumed — and recorded so the seat does not have to ask.

## 2026-08-16 — migration 445 APPLIED and verified live

DB config is **live on apply** — no roll gates it, no redeploy undoes it — so the file went into git
first (`eb137faed`) and the rollback sidecar was written before the apply, not after something looked
wrong.

Applied cleanly: snapshot captured, pre-state gate passed, `UPDATE 1`, post-state verify passed,
`COMMIT`. Scoped by id with a pre-state gate that refuses unless the row still carries **both** 406's
requires-backend gate **and** the `LIMIT 30` — so it could not have clobbered a concurrent change. That
shape is what 406's own council round asked future `agent_definitions` migrations to adopt.

**Verified live:**

| check | result |
|---|---|
| `LIMIT` gone from the query | **true** |
| `left(description, 200)` present | **true** |
| 406's requires-backend gate intact | **true** |
| tools that sort past position 30 — previously unreachable | **44 now visible**, first is *Early Settlement Estimator* |
| real payload on the no-backend path | **73 rows, 20,101 chars — +22.4%** vs 30 rows / 16,421 |

73 rather than 74 because the no-backend path correctly excludes the single `requires-backend` tool —
the gate still working, which is the thing the bug warned most about losing. Predicted +24%, measured
+22.4%.

## 2026-08-16 — adjacent findings, recorded not acted on

- **One library master has an empty `display_name`** (and one an empty `category`). An empty string
  sorts FIRST, so that row has *always* occupied a slot in the visible 30 while telling the model
  nothing — its description is developer-facing implementation notes (*"Parameterised calculator
  component (Track B2, bugs_open/263): panel, ids and scripts live in this template…"*). Content
  quality, a different lane's call.
- `category` selected but never rendered (above).
- **`bugs_open/242` is the same class in another subsystem** — *"a capped render audit is
  indistinguishable from a complete one"* — and is still open. Cross-referenced in LCO-009.

## 2026-08-16 — I checked the other six, and the census was wrong twice

Having written *"nobody has checked whether the other six bite"* into the bug file, I checked. Each was
one query. **Both corrections came from checking rather than reasoning.**

**Correction 1 — only FIVE of the "seven" are whole-result caps.** `fix-proposer.load_last_bundle` puts
its `LIMIT 2` inside a subquery and `string_agg`s the result into ONE row; `visual-design-auditor` does
the same in a correlated subquery. My end-anchored regex ignores both — **correctly**, and flagging
them would have been a false positive on a real config. That is the anchoring decision vindicated by a
live case rather than by my own reasoning about it, which is the only kind of vindication worth much.

**Correction 2 — NOT EVERY MULTI-ROW CAP IS A DEFECT, and I had implicitly assumed they were.** The
distinguishing question is not the cap's size but what the rows ARE:

| step | cap | population | bites | kind |
|---|---|---|---|---|
| `tool-suggester.load_library_tools` | 30 | 74 | fixed | **corpus** |
| `tool-recreation-handler.load_related_context` | 10 | **107** | YES | **corpus** |
| `internal-linker.load_candidate_pages` | 15 | **68** | YES | **corpus** |
| `content-feed-trigger.find_news_sites` | 5 | 9 | yes, but | work queue |
| `model-directory-trigger.find_directory_sites` | 12 | 3 | no | work queue |

A **work queue** takes N per run and the rest arrive next run — coverage is eventual and the cap is a
batch size. A **corpus shown to a model** takes N and the rest are never seen on that run. Only the
second is this bug's defect.

**The SQL cannot distinguish them**, so LCO-009's warning has to stay dumb and the reader makes the
call. I have written that into the register entry, because a detector whose false-positive class is
undocumented gets switched off by the third person who meets it.

**Two new confirmed instances, filed nowhere** (grepped both bug dirs first): `load_related_context`
sees 10 of up to 107 — **worse in ratio than 275 itself** — and `load_candidate_pages` picks internal
link targets from at most 15 of 68, alphabetically. Recorded in 275's file and the register rather than
filed as two new bugs: that is the owner's call, but the measurement should not have to be redone.

**What this says about the original framing.** I wrote "seven multi-row caps that can bite" into the
commit message, the submission and the register on the strength of a `substring` census — without
checking either whether the LIMIT bounded the whole result or whether the population exceeded it. Two
of the seven were not whole-result caps at all and two more were harmless queues. **The census answered
the question I encoded ("does this query text contain a multi-row LIMIT?") and I reported it as the one
I meant ("how many silent caps are there?")** — the same shape as the `bugs_open/257` census error
earlier today, on a different object, four hours apart.

## 2026-08-16/17 — council ROUND 1: REVISE, gated by debug_historian, and it found real things

Five seats abstained; `llm_reliability` approved as out-of-scope. `gated_by_truncation: false`.

### MISSTEP 3: my sketch omitted the safety-critical line, so the gating objection was fair

`debug_historian` (HIGH): *"never calls snapshot_agent() before the fenced UPDATE"*. **The file always
did** — it is the second statement, the apply printed `NOTICE: Snapshot captured`, and
`agent_definitions_backup` holds the row at 11:22:23Z. **But my SKETCH showed the DO blocks and left
the snapshot out**, and a seat can only review what it is shown. That is my error, not a misread.

**The lesson, and it generalises past the council: a sketch must show the SAFETY-CRITICAL lines, not
the interesting ones.** I had chosen what to include by what was novel about the migration, which is
exactly backwards — the reviewer's question is "is this safe", not "what is new".

### The duplicate-active-row objection: I was right and my gate was still wrong

Raised independently by `debug_historian` (HIGH) and `architecture` (MEDIUM): four agent types have
TWO active rows, only the higher version loads, and an id-scoped UPDATE on the shadowed one silently
no-ops while the verify blocks pass.

Measured: **tool-suggester has exactly ONE row** (id `c0756913…`, version 1). Not one of the four.
**But the seats were right about the GATE**: 445 counts rows matching the pinned *id*, which is 1 by
definition and structurally blind to a sibling. Being correct by luck is not the same as being
guarded. Migration 446 gates on `count(*) WHERE type='tool-suggester' <> 1` instead.

### The ledger hole was real, and not the one the seat guessed

`architecture` (LOW) asked me to confirm 445 was free. It was — 442/443/444 were applied, 444 at
11:11:48 the same morning. **But because I applied by hand via psql rather than through the runner, it
was recorded NOWHERE**, which is precisely how two sessions collide on a number. Both 445 and 446 are
now `INSERT`ed into `schema_migrations` with notes. A low-severity objection that found a live process
hole.

### The regex objection, answered by language semantics

`architecture` (MEDIUM): confirm the regex cannot pathologically backtrack, since it now runs on every
`query_database` call fleet-wide. **It cannot, by construction — Go's `regexp` is RE2, linear-time with
no backtracking at all.** The failure mode is not available. Input is agent-authored SQL from config,
never user input; the pattern is end-anchored with no nested quantifier.

### MISSTEP 4: I nearly published a compelling, false, second-order damage claim

`editquality` (MEDIUM) asked whether a downstream filter silently drops suggestions — the *"widening a
planner's MENU changes nothing"* landmine. **Checked: there is none.** The loop routes on
`current_suggestion.tool_component_id != null` — library id → deployer, otherwise → `tool-generator`
**builds it from scratch**. Nothing is dropped, so widening the menu changes the outcome and not just
the prompt. Good news for the fix.

**Then I over-reached.** If an invisible tool can only be suggested as "novel", the cap should have
been causing the fleet to REBUILD tools it already had — the same shape as `bugs_open/204`'s junk work
items. I measured: **18 of 19 novel build-requests name a tool the library already holds.** A striking
number, and I was one paragraph from writing it into the bug file.

**The disconfirming check refuted it.** Comparing timestamps: **all 18 library masters were created AT
OR AFTER their work item** — so they were created *by* those novel builds, not duplicated by them. The
match is the pipeline working normally, not waste. **There is no evidence the cap caused duplicate
builds and I am not claiming it.**

That is three times today my inference has outrun my measurement (257's census, 275's census, this).
The difference here is only that I ran the check *before* publishing — which is the entire process
working, and worth noting as such rather than as a near-miss.

### bug_historian was right: I replaced a silent ROW cap with a silent COLUMN cap

*"`left(description, 200)` with no truncation marker — 50 of 74 cut with nothing telling the LLM or a
future reader. Same failure mode as the bug being fixed."* **Entirely correct**, and the sharpest
objection of the round: the estate already marks truncation explicitly (`TruncateString`'s tested
ellipsis contract; webscrape's literal banner) and 445 followed neither.

**Migration 446, applied and verified**: `CASE WHEN length(description) > 200 THEN left(description,200)
|| ' […truncated]' ELSE description END`. **49 of 73 rows now carry the marker**; payload 20,738 chars
against 445's 20,101 — **the signal costs ~3%**.

**The marker rather than a bigger cap is the deliberate choice**: 300 chars would cost +53% against
+26% and *still* cut 24 descriptions silently. Loss without a signal is a defect; loss with one is a
budget.

Its sibling objection — the row-cap detector cannot see column truncation — is true and now a stated
scope limit in LCO-009 rather than something for the next reader to discover. A generic
column-truncation detector is not attempted: `left(...)` in agent SQL is indistinguishable from a
legitimate projection without knowing the consumer.

**Round 2 resubmitted on the same trail** (`b684a399`, run orch `517928d9`).

## 2026-08-17 — council ROUND 2: APPROVED, and MISSTEP 5 is the same one as misstep 3

`decision: approved`, *"3 advisory objections, none high-severity"*, `gated_by_truncation: false`,
4 abstained.

### MISSTEP 5: I wrote the lesson in round 2 and broke it in the same submission

Round 1's gating objection was that 445 never snapshots — false about the file, true about my sketch.
I wrote the general lesson into the round-2 rationale: *"a sketch must show the SAFETY-CRITICAL lines,
not the interesting ones."*

**Then the new sketch I wrote for 446, in that same submission, omitted its pre-state gate**, and
round 2's `editquality` seat objected (medium) to exactly that. The file has the gate at lines 52-56
and always did.

**I fixed the instance and not the class.** I added `snapshot_agent` to the two sketches the seat had
NAMED, and wrote a third sketch that dropped a different safety line. The remediation *felt* complete
because the objection was answered — which is the whole failure mode, and it is distinct from not
having learned the lesson at all.

The concrete fix, since a rule I have now broken twice needs a mechanism and not another resolution:
**for a migration, sketch the WHOLE file.** They are ~40 lines. The summarising was never buying
anything, and it has now cost two review rounds' attention on artefacts of my own description.

### The one real code gap, and it was in the worst possible place

`editquality` (low): `LIMIT 30 -- widened 08-14` is a **false negative** — a genuinely capped query
going undetected *by the mechanism built to end undetected caps*. Fixed: the regex tolerates trailing
`--` and `/* */`. Tested **both directions**, because widening a pattern is how you turn a false
negative into a false positive: a comment must not HIDE a cap, and prose mentioning one
(`-- we removed the LIMIT 30 here`) must not INVENT one. Mutation-proven — M10 (comment-blind regex)
killed by the first, M11 (no end anchor) by the second.

That objection is worth more than its severity label. An operator who annotates a cap is the operator
who *thought about it*, and that is precisely the cap most worth a second look.

### `bug_historian`'s observation, carried rather than closed

Approved, with the note that Part A is WARN-only while this lane's own census found two live analogous
caps. **Observation is not remediation.** The warning makes them visible; it does not fix
`tool-recreation-handler.load_related_context` (10 of up to 107) or
`internal-linker.load_candidate_pages` (15 of up to 68). Both remain open, recorded in the bug file,
and are an owner call.

## 2026-08-17 — the two instances filed as tickets (owner directed), and re-measuring changed both

Filed as `bugs_open/297` and `bugs_open/298`. **Re-measured from scratch rather than copying my own
figures**, which was right twice over:

- **298's population was over-counted.** My census query omitted the step's own
  `HAVING COUNT(pc.id) > 0`. With the real predicate it bites at **8 of 24 sites**, not the "worst site
  has 68" framing I had been carrying — and the **median site (12) is UNDER the cap**, so most sites
  are unaffected. That is a materially weaker bug than I had been describing.
- **298's reachability collapsed under checking.** `llm_call_log` has **zero** rows for
  `internal-linker` in all history, so the `plan_links` step that consumes those candidates has no
  logged call, ever. The cap is structurally real and the step is reachable — 69 work items name the
  agent (newest today) and 13 of 38 completed runs passed `check_target_found` — but **whether it has
  shaped a single link decision is UNMEASURED**, and the ticket says so instead of guessing.

⚠ **`owner_agent_type` is the WRONG instrument for "does this agent run".** It returned 0 for
`tool-recreation-handler` — which has **290 `llm_call_log` rows**. It also revealed a near-miss:
`internal-link-resolver` (54 orchestrations) is a *different* agent from `internal-linker`, and has no
capped query at all. Two similar names, one real. Use `llm_call_log` and `site_work_items.handler_agent`.

**297 is the stronger ticket by a distance**: live, 19 of 24 sites over cap, and at the median site the
model sees 10 of 26 — 38%, against 275's 41%. The bug I fixed was not the worst instance of its own class.

**Adjacent finding recorded in 298, not chased:** 15 of 38 completed `internal-linker` items found no
target page and completed anyway — a run that links nothing is indistinguishable from one that worked,
by `status` alone.
