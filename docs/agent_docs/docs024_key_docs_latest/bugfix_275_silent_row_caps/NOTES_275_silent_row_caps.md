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

## 2026-08-17 — a roll happened and LCO-009 is STILL NOT LIVE (verified two ways)

Chassis pods restarted 14:42–14:43Z. **They are serving the same cached image.** Verified twice,
independently, because "pods look new" is exactly the surface this trap fools:

1. **Binary probe with controls.** `query_database result count EQUALS` (added by `eb137faed`) →
   **ABSENT**; `load page slot identities` (added by `bugs_open/257`, shipped at v1.0.1305 yesterday) →
   **present**; a plausible fake (`zzz-fabricated-never-in-any-build`) → **absent**. So the probe
   discriminates and the binary is yesterday's. ⚠ The fake matters — LANDMINES records a run whose
   negative control was 40 zeros, which **matches every Go binary**, so that run could not discriminate
   and its "absent" readings were worthless.
2. **Image digest**, which is the landmine's own one-command proof: running
   `sha256:f90a7e88…` vs locally built `sha256:6039e19c…`. **Different digests under the same tag.**
   Byte-identical to the worked case already recorded there — the situation has not moved.

`IMAGE_TAG ?= v1.0.1305` is unbumped, and with `imagePullPolicy: IfNotPresent` the node keeps serving
its cached layer for ever.

**Scope, because this is not my lane's problem alone:** `git rev-list 6a782274b..HEAD` = **238
commits** not in the running binary, **26 of them touching Go**, across at least eight lanes (275, 283,
285, 289, 291, 292, 293, 295, 299).

**Already a landmine** (`A same-tag rebuild leaves the OLD binary running under new pods`), so nothing
new to file — grepped first, found it, used its check. The fix is one line and owner-run:
`make release IMAGE_TAG=v1.0.<next>` (it is `?=`, so no file edit is needed).

**Consequences for this lane, stated so nothing over-reads:** migrations 445 and 446 are config and
**are live** — the tool-suggester fix is working. LCO-009's detector is **not**, so the live census of
which caps actually fire (and `bugs_open/298`'s reachability question, which the detector would answer
by itself) is still owed and cannot be done yet.

## 2026-08-17 (evening) — LCO-009 LIVE at v1.0.1307; the "before" proof done; a time-confound caught

**The detector shipped this time.** v1.0.1307, pods 17:05Z, verified at the binary with controls (added
string present, known-present control present, plausible fake absent). The 14:42Z "fresh build" had
shipped nothing — same tag, cached image — so this is the second attempt and the first real one.

### The before-half of the bug's own proof, done from stored data

No new run needed: the most recent pre-fix `suggest_tools` prompt is in `llm_call_log`. Ranked against
the library **as it stood then** (71 masters): **29 tools in the prompt, 0 past rank 30, highest rank
exactly 30.** The model saw the first 30 alphabetically and nothing else — 41 unreachable in a real,
specific run. Config-level evidence ("the query says LIMIT 30") became artefact-level evidence.

### MISSTEP 6: I ranked a two-day-old artefact against today's state

The first version of that query ranked **today's 74** tools and reported *"1 tool past rank 30 appeared
in the prompt"* — which would have undermined the claim (the cut looking not-quite-alphabetical) and
was purely an artefact of tools added after the prompt rendered. Constraining the CTE with
`created_at <= <prompt timestamp>` gives the clean 0.

**The check: when you measure a stored artefact, every population you compare it against must be
constrained to that artefact's timestamp.** This is the same family as today's earlier census errors —
the query answered a question about *now*, and I was asking one about *then* — and it is the fourth
time today the shape has appeared. Caught before publishing, again by running the disconfirming version
rather than the confirming one.

### The live census is not answerable yet, and I am not reporting a zero as a result

`EQUALS the query's LIMIT` → 0 firings. **Demand control: only 5 `query_database` completions since the
roll, all `agent-landmine-verifier` (`load_entry`, `LIMIT 1`, correctly excluded). No capped step has
executed.** So the zero is uninformative, and saying "the detector has not fired" without that control
would be exactly the blind-pass this estate keeps logging.

`content-feed-refresh` (cap 5, population 9) is 6-hourly and last fired 14:31Z pre-roll — the first
expected positive. Command recorded in the bug file.

## 2026-08-18 — ran the census; the answer is "nothing it watches has run", plus one real limitation

Detector re-verified at **v1.0.1309** (pods 15:45Z; added string present, known-present control
present, plausible fake absent). Swept **all 41** chassis-image pods over 24h.

**0 WARNs. 21 `query_database` completions.** And — this is the part that makes the zero readable —
the log line carries `step_name`, so all 21 are attributed: `find_dispatchable_site` ×9,
`notify_scheduler` ×6, `load_entry` ×3, `notify_scheduler_idle` ×3. **Not one capped step.**

So the zero says nothing about the detector. Reporting "the WARN has not fired" without that table
would be precisely the blind pass this estate keeps logging — the traffic control passing while blind.

### The `LIMIT 1` exclusion, vindicated by measurement rather than by argument

`find_dispatchable_site` returned **exactly 1 row on 5 of its 9 runs** — sitting on its own `LIMIT 1`
five times in 24 hours, on a quiet fleet, from a dispatch-loop step that runs continuously. **Without
the `n >= 2` exclusion that is five false warnings from one step in one quiet day.**

I made that decision on reasoning alone (19 of 26 live hits are fetch-one; a channel that always fires
is a channel nobody reads) and wrote it into the source as the arm most likely to be "simplified" away
later. It is satisfying to find the margin is larger than the argument claimed — but the useful lesson
is the reverse: **the design argument was checkable all along and I did not check it until now.** The
same query would have run before I shipped.

### ⚠ MISSTEP-ADJACENT: a limitation the design review never surfaced, that running it did

**The WARN is a log line, so its history dies with the pod.** The observable window is "time since the
last pod restart", and I never considered that when choosing log-only.

Measured today: pods restarted **15:45Z**; `content-feed-refresh` (cap 5, population 9) last fired
**14:32Z**; `model-directory-publish` **12:15Z**. **Both fired before the restart**, so both are
invisible, and their next runs are ~20:32Z and ~18:15Z.

Rolls land roughly daily on this tree; the capped steps are 6-hourly. **A cap that fires shortly before
a roll is invisible for ever.** The detector is correct and live — the unit tests and the binary probe
both say so — but *whether it will actually catch the caps in practice* is a race between roll
frequency and schedule period that nobody has characterised, and I certainly did not.

**This is the strongest argument yet for the follow-up I recorded as out of scope** (the `LIMIT n+1`
probe): a definitive result written somewhere durable — a `doc_notes` row, a column — survives a roll.
A log line does not. I chose log-only for good reasons (observational, no authority on a shared seam,
smallest possible intervention) and those reasons still hold; what I got wrong was not noticing that
the *medium* has a retention property, and that the thing being watched runs less often than the medium
is wiped.

**Recorded as LCO-009's `verify-later`, not fixed here** — widening the seam is a separate change with
its own review, and the council was explicit that the observational scope was the right call for THIS
one.

### Still owed

**0 `suggest_tools` runs since migration 445 (11:22Z on 08-17)**, so 275's "after" half remains
undone. The "before" half is proven at the artefact (29 tools, 0 past rank 30, highest exactly 30).

---

## 2026-08-18 (evening) — the channel the detector logs into cannot retain what it is meant to census; a durable channel already holds it, and answers §3b and §3c

Picked up from `HANDOFF_2026-08-18_continue_here.md` to run the three owed verifications. Two of the
three are now answered — but not by the instrument the handoff specified, because that instrument
cannot work. The correction below **supersedes the "⚠ MISSTEP-ADJACENT" entry above**, which
understated the problem by about three orders of magnitude.

### First, the artefact check, because everything below depends on it

Pods `agent-chassis-85b844f547-{l8r76,xdsz6}`, image `v1.0.1310`, both started **18:00Z**, digest
`sha256:9ca35bac…` identical on both.

`build provenance` was **absent from `--tail=3000` on both pods twenty minutes after they started** —
so the startup-line recipe in CLAUDE.md failed here, exactly as its own caveat says it can. Fell back
to the binary probe, run as one `grep -aoFf -` over 122 candidate shas (120 recent commits + two
controls) so the answer could come out wrong:

| probe | result |
|---|---|
| fabricated sha `deadbeefcafe…` (must be ABSENT) | absent ✓ |
| real sha from 3,000 commits back (must be ABSENT) | absent ✓ |
| exactly one of the 120 recent shas matched | `0b185bad2a49c6e032352fa9e7d0b429f0a95104` |

`git merge-base --is-ancestor eb137faed 0b185bad2` → **yes**. [MEASURED] The detector is in the
running binary.

### ⚠ CORRECTION — the observable window is SECONDS, not "time since the last pod restart"

The entry above says the WARN's history "dies with the pod" and reasons about daily rolls against
6-hourly schedules. That framing is wrong, and it is wrong in the optimistic direction. The container
log rotates **on size**, and the coordinator emits whole-state dumps, so the log is wiped continuously
while the pods keep running.

[MEASURED] 2026-08-18, both pods up since 18:00Z with **0 restarts**:

| sample | pod | oldest retrievable line | window |
|---|---|---|---|
| 18:27:09Z | l8r76 | 18:27:06Z | **3 s** |
| 18:27:24Z | xdsz6 | 18:26:50Z | **34 s** |
| 18:28:37Z | l8r76 | 18:27:06Z | **91 s** |
| 18:28:45Z | xdsz6 | 18:28:30Z | **15 s** |
| 18:29:11Z | xdsz6 | 18:28:50Z | **21 s** |

Cause, by bytes, from one snapshot (448 KB in 205 lines — mean **2.2 KB/line**, worst single line
**183 KB**):

```
68389 B   5 lines  Executing local action in executelocalaction
65419 B   5 lines  Executing local action - result back is: look for request id
64626 B   5 lines  just into executeLocalAction look for execCtx before it gets changed
52266 B   4 lines  Transitioning to next step
22227 B   5 lines  CollectedData structure
```

And there is nowhere else to read it: no aggregator runs in the cluster (`kubectl get pods -A` matches
nothing for loki/fluent/vector/elastic/promtail; nothing in `deployments/`), and
`platform/logger/logger.go:37` sets `OutputPaths: []string{"stdout"}` — one sink, rotating.

**This is not a new discovery and I should have grepped for it before measuring it.** `LANDMINES.md`
has carried it since 2026-08-08 ("A chassis pod's retrievable log holds LESS THAN A SECOND of history
under load"), measured at 0.4 s, with the two remedies I ended up rediscovering: attach `kubectl logs
-f` *before* the event, or arrange a DB-visible observable. The lane built a log-only detector with
that entry already on file.

### ⚠ MY OWN MISSTEP, and it is the same one the handoff warns about

My first action was the RUNBOOK's census over 48 pods with `--since=45m`. It returned **0 WARNs and 0
`query_database` completions**, and the base rate in the handoff (21 completions/24 h) makes zero look
entirely reasonable. I was one step from recording "nothing has fired" a second time.

What stopped it was checking whether the thing I was censusing had *run*: the 18:15Z
`model-directory-trigger` is in `orchestration_states` as COMPLETED, on pod `l8r76`, whose captured log
began at **18:23:44Z** — after the event. The zero was not a negative result. It was a blind pass, and
the same blind pass the handoff's own 24 h census recorded.

### The detector was live for almost none of the window that census covered

[MEASURED] Oldest surviving replicaset carrying `v1.0.1309` (the release that first contains the
detector) was created **2026-08-18 15:45:31Z**; `v1.0.1310` at 17:58–18:00Z. Older replicasets are
pruned, so 15:45Z is the earliest *evidenced* roll, not provably the first.

So the 24 h census in the handoff spanned a period in which the detector was mostly **not running at
all**. Two independent reasons for the same zero, neither of them "no caps fired".

### The durable channel — `collected_data` already holds every fact the WARN would report

`QueryDatabaseAction` writes its result to the step's `output_field`, and that lands in
`orchestration_states.collected_data`, which survives rolls. So the census does not need the log at
all: **array** output → `jsonb_array_length`, **object** output → `->>'count'`, compared against the
step's own cap. Retroactive, and it produces negatives.

[MEASURED] 2026-08-18, all three live capped steps, over the whole retained table:

| agent | step | cap | runs measurable | max rows | **runs that HIT the cap** |
|---|---|---|---|---|---|
| `content-feed-trigger` | `find_news_sites` | 5 | 4 | 5 | **3** |
| `internal-linker` | `load_candidate_pages` | 15 | 2 | 15 | **1** |
| `model-directory-trigger` | `find_directory_sites` | 12 | 5 | 4 | **0** |

Itemised, because an aggregate cannot show you a control:

```
2026-08-17 20:31Z  content-feed-trigger   5 of cap 5   HIT
2026-08-17 22:22Z  internal-linker        7 of cap 15  under
2026-08-18 01:01Z  internal-linker       15 of cap 15  HIT
2026-08-18 02:32Z  content-feed-trigger   4 of cap 5   under      <- same agent, same query, not a constant
2026-08-18 08:32Z  content-feed-trigger   5 of cap 5   HIT
2026-08-18 14:32Z  content-feed-trigger   5 of cap 5   HIT
```

Two independent negative arms, which is what makes the positives worth anything: `model-directory-trigger`
never exceeded 4 against a cap of 12 across 5 runs, and `content-feed-trigger` itself returned 4 on one
of its four runs.

⚠ **The durable channel has a horizon too, and it is ~2 days.** `orchestration_states` holds 5,701 rows
back to 2026-07-13, but only **25** are older than two days. So this is a far better instrument than the
log (~3,000× the window) and still not "all history" — do not read "3 hits" as a lifetime total.

### What this means for the three owed verifications

- **§3b is answered, by the better instrument.** Four cap hits in the retained window, with working
  controls. What is NOT yet shown is the WARN itself firing — because every one of those six runs
  predates 15:45Z, when the detector went live.
- **Since the detector went live, exactly ONE capped step has executed**: `model-directory-trigger` at
  18:15Z, **4 rows against a cap of 12**, so no warning was due. That is the negative control the
  handoff predicted, and it came out as predicted — though note the WARN's *absence* was not observed
  either (the log had already rotated); the durable row is what says no warning was due.
- **First genuine positive opportunity: `content-feed-trigger` at ~20:32Z tonight.** [MEASURED] its
  eligible population right now is **6 against a cap of 5**, and three of its last four runs returned
  exactly 5. Not certain — the predicate includes `next_fetch_at <= NOW()`, so the population moves.
- A streaming capture is armed for it (`kubectl logs -f` since 18:34Z, deadline 20:45Z), because with a
  15–90 s window streaming is the only way to read a WARN.
- **§3a is still blocked.** [MEASURED] `suggest_tools` has run **0 times** since migration 445; last run
  in all history **2026-08-15 20:29Z**, and the agent has been quiet on 08-16, 08-17 and 08-18. Its
  history is 1–9 runs on roughly half of days, so this may clear on its own or may not.

### §3a, as far as it CAN be taken without a run

The prompt artefact still needs a real run, but the mechanism is verifiable now, at the live row rather
than at the migration file. [MEASURED] the live `load_library_tools` config carries **no `LIMIT`** and
bounds `description` to 200 chars with the ` […truncated]` marker; executing that exact query today
returns **76 tools, of which 46 rank past position 30** — 46 the pre-fix query could not have shown —
with **54 descriptions marked truncated** and a longest description of 213 chars (200 + the 13-char
marker). What remains unproven is only that a *prompt* carried them.

### §3c is answered, and the answer is not the one the ticket expected — see `bugs_open/298`

The handoff says `internal-linker` has "zero `llm_call_log` rows in all history, so whether its
`LIMIT 15` has ever shaped a link decision is unmeasured", and that the detector would answer it. Both
halves turned out differently.

[MEASURED] the agent is live and has run: **13 `orchestration_states` rows**, 82 `site_work_items` naming
it as `handler_agent`, most recent run **2026-08-18 07:33Z**. The "zero rows" was true of `llm_call_log`
only — which is the single-channel reading the handoff's own landmine #2 warns against.

Two of those runs reached `load_candidate_pages`; one returned **exactly 15** — the cap. But **both runs
ended at `current_step = complete_no_candidates`**, including the one holding 15 candidates.

The reason is readable in the config and the code, and it is not the cap:

- `load_candidate_pages` declares `"output_format": "array"`, and `QueryDatabaseAction` returns the bare
  slice for that format (`database_actions.go:129`) — **no `count` key exists**; `count` is only set in
  the `object` branch.
- `check_candidates` tests `"candidate_pages.count > 0"`.
- `resolveFieldValue` (`conditional_branch_action.go:320`) tries five strategies; strategy 5, the
  recursive one, requires the base to be a `map[string]interface{}`, and an array is not, so all five
  fail and it returns `nil`.
- The numeric arm then fails `datahelpers.ToFloat64(nil)`, logs `Numeric comparison: left side is not
  numeric`, and **returns false** — the `else_step`, `complete_no_candidates`.

So `plan_links` cannot run, which is exactly why there are no `llm_call_log` rows. **The cap has never
shaped a link decision because the linker has never made one.** The condition is unsatisfiable as
configured, on every run, for as long as the config has had this shape.

**Filed to the diagnosis loop before asserting it anywhere durable** (owner ruling 2026-07-31 — this is
a root cause outside the symptom, on a shared conditional mechanism):
`RUN_CORRELATION_ID=c4aa3559-86b1-4356-a28b-c71dfa661465`.

### My own missteps this session, in the order I made them

1. **Ran the log census first and nearly recorded its zero** (above). The check that saved it — "did the
   thing I am censusing actually run?" — cost one query and belongs *before* the census, not after.
2. **`grep -m1 'build provenance'` returned a 1.6 MB line.** On these logs a substring can match inside a
   183 KB state dump, so `-m1` is no protection. `grep -ao '"msg":"build provenance","git_commit":"[0-9a-f]*"'`
   is the form that cannot run away.
3. **`date -u -d '2026-08-18 20:45:00'` parsed as LOCAL time** and set the watcher's deadline to 19:45Z —
   an hour before the event it exists to catch. The `-u` flag formats output in UTC; it does not make the
   *input* UTC. `date -u -d '... 20:45:00Z'` does. It printed the deadline back, which is the only reason
   I caught it.
4. **`pkill -f warn_watch.sh` killed the shell running it**, because that shell's own command line
   contained the pattern. `pkill -f 'warn_[w]atch'` is the form that excludes itself.
5. **`cut -c8-30` on `"ts":"2026-…"` chopped the leading `2`**, so `date -d` produced epoch-63-billion
   nonsense in the window sampler. The raw timestamps were still in the output, so the measurement
   survived; had I only kept the computed field, the whole sample would have been silently garbage.

### 2026-08-18 18:56Z — the diagnosis came back CONFIRMED, and `bugs_open/313` is filed

**CONFIRMED, first iteration**, `c4aa3559`. It re-read the same config and code independently, cited
`check_candidates`, `load_candidate_pages`, the surviving `orchestration_states` row (with the actual
page list in it), and the `resolveFieldValue` type assertion — and added the one link I had left
implicit: **`check_candidates`' `then_step` is `load_specs`, and `load_specs` is the only step that
chains to `plan_links`**, so the else branch alone is sufficient to make the LLM step unreachable. That
is the difference between "the condition is false" and "the agent cannot work", and I had asserted the
second while only having read the first.

Filed as `bugs_open/313`; `bugs_open/298` now carries the verdict and the fix ORDER (313 first — fixing
the cap alone changes nothing observable, and fixing the branch alone makes the cap live on 8 sites).

**Three things the filing turned up that the analysis had not:**

1. **The mismatch dates to the agent's creation.** `sql_for_agents/101_internal_linker.sql` (added
   2026-04-13; row created 2026-04-12) carries the identical pairing, and live config agrees with the
   seed. So there is no drift and no regression to hunt — it has never worked. **[MEASURED]** 57
   `needs_internal_links` items read `complete` between 2026-07-24 and 2026-08-18 while the workflow
   step itself declares `"status": "skipped"`.
2. **The blast radius is exactly one.** Censused across live config, **7** conditionals test some
   `X.count`; six read a producer that really emits it (four `query_database` steps with
   `output_format: object`, plus `load_unswept_areas` and `load_ch_enrichment_batch`, whose Go actions
   set `"count": len(...)`). Worth measuring before writing the fix, because "the shared resolver
   should synthesise `.count` for arrays" is the tempting root fix and it would change what a shared
   conditional guarantees for every always-false condition of this shape. One instance does not justify
   that; the bug file argues against it explicitly.
3. **The obvious one-word fix is a trap.** Flipping `load_candidate_pages` to `output_format: object`
   makes the condition resolve — and breaks `plan_links`, whose template does `{{range .candidate_pages}}`
   over the bare array. Ranging a map yields `rows`, `count`, `columns`. **Both halves in one
   migration, or you trade a dead branch for a broken prompt.** I only found this by reading the prompt
   template, which is the same lesson as 445: read what actually consumes the column.

### ⚠ MISSTEP: I read a bulk timestamp as another session's edit

`agent_definitions.updated_at` for `internal-linker` read **2026-08-18 17:59:19Z** — an hour before I
filed against it, and a minute before the pods rolled. That looks exactly like another thread editing
the agent I am about to write a bug about, and the correct response to that is to stop.

**[MEASURED] 200 of 201 live definitions share that minute** — it is the fleet release syncing
`image_tag` (all 201 rows now read `v1.0.1310`), which the makefile documents at length under
`bugs_open/066`. One `GROUP BY date_trunc('minute', updated_at)` separates "someone edited this row"
from "everything was touched at once".

**And this is already in `LANDMINES.md` twice** — "*`agent_definitions.updated_at` is bumped by BULK
SWEEPS — a fresh timestamp is not another session working on your agent*" and a second entry on the
missing trigger. That is twice in one session that I measured my way to something the file already
knew (the log window was the first). **The pattern is mine, not the tool's: I grep LANDMINES by path,
and both of these are keyed on a TABLE and a COMMAND, which the SessionStart hook cannot match.**
Grepping the table name I am about to trust would have cost one command, and CLAUDE.md says to do
exactly that.

One useful side-effect of the scare: it forced a check that mattered for this lane — migration 445/446
edits `agent_definitions`, so a fleet-wide write to that table an hour earlier could in principle have
reverted them. It did not; the live `load_library_tools` config verified at 18:38Z (after the 17:59Z
sweep) still has no `LIMIT` and still carries the truncation marker.

### 2026-08-18 20:32Z — the cap FIRED, and the streamed capture still missed the WARN

**The durable channel delivered.** `content-feed-trigger-orchestrate-0818-2032`, created **20:32:54Z**,
`news_sites.count = 5` against a cap of **5** — `hit_the_cap = true`, with five real domains in the
payload. **This is the first cap hit since the detector went live at 15:45Z**, and it was read straight
out of `collected_data` while the orchestration was still `AWAITING_RESPONSES`.

**The WARN was not captured.** 0 anchored WARN lines, against a liveness control of **133 anchored
`QueryDatabaseAction: Complete` lines** in the same capture, with the watcher still running.

**Do NOT read that as "the detector did not fire."** The step's *unconditional* completion line
(`database_actions.go:111`, logged on every `query_database` step regardless of any cap) is **also
absent**. Both lines from that one step are missing together, which locates the failure at the
observation layer, not in `resultHitItsRowCap`.

What was ruled out, in order:

| hypothesis | verdict |
|---|---|
| ran on a pod outside the watcher's selector | **refuted** — `processing_node` is `l8r76`, one of the two followed pods, and all 8 recent trigger runs ran on `agent-chassis-*` deployment pods, never on a dynamic-agent pod |
| the stream had disconnected | **refuted** — no `[stream ended - reconnecting]` marker at all, and captured lines bracket the event at 20:32:25Z, 20:33:02Z and 20:33:55Z from both pods |
| the pod ran an older binary without the detector | **refuted** — all **62** chassis-image pods carry `v1.0.1310` |
| `logs -f` is a lossy instrument in general | **not supported** — fidelity test on the same busy pod: the stream saw **561** lines over 60 s where the pod's own retained log held **291** for the same window. The stream is nearly 2× richer than a retrospective pull; it is rotation that eats the pull |

**So the miss is unattributed, and I am leaving it that way rather than picking the tidiest story.**
The remaining candidates are a gap in `kubectl logs -f` across a rotation boundary (which the fidelity
test above cannot isolate, because the pull is the weaker instrument on both sides) and something about
that step's logging I have not thought of.

**⚠ One real instrument defect found regardless, and it would have mattered on a different night:** the
watcher followed `-l app=agent-chassis`, which matches **2 of the 62 pods running the chassis image**.
The other 60 are ephemeral agents labelled **`app=dynamic-agent`** (`agent-feed-ingester`,
`agent-page-rerender`, `agent-content-feed-orchestrator`, …), and `agent-job-cleanup` deletes them
within minutes of completion. It did not cause tonight's miss — these triggers run on the deployment
pods — but any capped step executing inside a dynamic agent is invisible to that selector *and* to any
later log pull, because the pod itself is gone.

### What this settles for the lane

The point of tonight was to witness the WARN once, end to end. **It did not happen, and the attempt is
more informative than the success would have been:** the medium was chosen for a detector whose events
are 6-hourly, the retention is seconds, and even doing the prescribed thing — attaching `logs -f`
*before* the event, on the right pod, with the stream verified live and the pattern verified anchored —
still produced nothing to show. Three instrument bugs of my own had to be fixed first (block-buffering
`cut`, unanchored patterns, a local-time deadline), and the fourth (the label selector) is still there.

**The durable channel answered the same question in one query, retroactively, with controls.** That is
the recommendation: LCO-009's WARN is a hint for someone already watching; `collected_data` is the
record. LCO-009's `verify-later` should now read "witness the WARN once" as an open item that has been
*attempted and failed*, not as untried.

## 2026-08-19 09:00Z — new build verified, and the census turns up a harm the class analysis had explicitly waved through

### The build, at the artefact

`v1.0.1314`, rolled **07:50–07:52Z**; 44 of 45 chassis-image pods carry it (one v1.0.1310 straggler).
Probed `/proc/1/exe` with **three** controls this time — a fabricated sha, a real-but-ancient sha, and
**yesterday's stamp `0b185bad2`**. All three ABSENT; one match:
`d3590ca4638d49bb6a3874db681814c4b0a99bbe` (08-18 22:17). Adding yesterday's stamp as a control is the
cheap way to prove a roll shipped **new code** rather than serving the node's cached image — the
"a fresh build can ship no new code" trap, answered with a query instead of trust. 158 commits between
the two stamps; `git merge-base --is-ancestor` confirms the detector is still in the binary, and none of
those 158 touched `query_row_cap.go`, `database_actions.go` or `conditional_branch_action.go`. **313 is
still live and unfixed**; the live config still reads `array / candidate_pages.count > 0`.

### The census, now with five runs of history

| agent | cap | runs retained | hit the cap |
|---|---|---|---|
| `content-feed-trigger` | 5 | 5 | **5 — every single run** |
| `model-directory-trigger` | 12 | 4 | **0** (4 rows each time) |

Yesterday's "3 of 4" has become **5 of 5**: the run that came in under (08-18 02:32Z, 4 rows) has aged
out of the ~2-day window. **Do not read the improvement as a change in the fleet** — it is the same
queue, measured over a window that now excludes its one quiet run. The negative control is unchanged
and still discriminating.

### ⚠ Which prompted the question the cap census does not ask: WHO does the cap cut?

`ORDER BY s.domain` is alphabetical and stable, so I checked whether the tail is starved. It is, and
the boundary is exact:

```
rank 1-5  (ai-agent-orchestration … mortgagecalculator)  overdue by 0        <- served 08:34-08:42
rank 6-9  (relojistas, robot-hands, vetcomparison, webdesign.co.uk)  all overdue <- served 02:36-02:42
```

Measured against each site's **own** `content_sources.fetch_interval`, `relojistas.com` is **117% of its
own cycle late** — and it is worst-hit precisely *because* it asked for the most (3-hourly, so it comes
due twice per 6-hourly window while sitting one place past the cut).

**And the queue is 2.10× oversubscribed** — 42 site-fetches demanded per day by the configured cadences
against 20 slots supplied (4 runs × cap 5). **Removing the cap does not close that**: 4 × 9 = 36 vs 42.
So ordering fixes *who* absorbs the shortfall; only capacity fixes the shortfall.

**Filed as `bugs_open/316`.** It explicitly narrows a gloss THIS LANE wrote into register LCO-009 — that
a work-queue cap means "coverage is eventual, not a defect". Coverage is eventual; it is also
deterministically unfair, and nobody had looked. The register note was reasoning, not measurement, and
it was written by the same session that built the detector — the "read the WARN" advice told a future
reader to dismiss exactly this case.

**The transferable lesson:** the cap census answers *whether a result was truncated*. It cannot answer
*who was cut*, and for a work queue that second question is the whole defect. One `ORDER BY` inspection
per capped step is the cheap check, and I only ran it because five identical HITs in a row looked too
tidy.

### §3a, day 4

**0 `suggest_tools` runs since migration 445**; last in all history 2026-08-15 20:29Z. Four days quiet
against a historical cadence of 1–9 runs on roughly half of days. `plan_links` likewise still **0 rows
all history**, which is 313's disconfirming arm staying clean.

## 2026-08-19 09:35Z — OWNER RULINGS: stop chasing the WARN; dispatch the suggester deliberately

Two decisions taken, both recorded here because the reasoning matters more than the outcome.

### 1. Witnessing the WARN in the logs is CLOSED — by decision, not by success

It was attempted properly once (08-18 20:32Z, cap fired 5 of 5, capture attached beforehand on the
right pod, stream verified live, patterns verified anchored) and produced nothing, with the miss
unattributed. **The owner ruled to stop.** Not "unresolved and someone should try again" — closed.

The justification, so nobody reopens it out of tidiness: the WARN's only job is to make a suspicion
visible to someone already looking, and **`collected_data` answers the same question better on every
axis** — retroactively, with controls, over ~2 days rather than ~60 seconds, and without needing anyone
present at the moment it fires. Witnessing the log line would have proved the code path executes; the
unit tests and the binary probe already establish that, and no decision depends on the third proof.

### 2. §3a: the last owed proof is DISPATCHED, not waited for

**Why waiting was never going to work — measured, and this is the part I had wrong.** I had reported
§3a as "wait and it may clear on its own". Reading `check_missing_tools.go` shows it cannot: the
producer of `evaluate_tools` items applies a **tiered cooldown** — 7 days for a site with no tools,
**30 days** for a site with tools that is not behind its content-to-tools ratio. Every candidate site
was evaluated on 08-10..08-15, so the next natural run was **mid-September**. "It runs 1–9 times on
roughly half of days" was true of its history and false about its future, because the history included
the initial sweep across sites that had never been evaluated.

**What was dispatched, and why it is not hand-rolling.** The item filed is byte-for-byte the shape the
discovery check itself files (`check_missing_tools.go:225-238`): `source=discovery`, `pipeline=build`,
`item_type=evaluate_tools`, `severity=low`, `priority=130`, `handler_agent=tool-suggester`,
`status=detected`, `item_key=evaluate_tools:<site_id>`. The only difference is that it arrives ahead of
the cooldown — which is exactly what "dispatch one deliberately" means. `created_by` is set to
`bugfix_275_silent_row_caps` rather than an agent name, so the row is honest about who filed it.

**Blast radius, established BEFORE firing rather than discovered after:**

- The workflow ends in `create_items_loop`, which creates one **`add_tool`** work item per suggestion —
  library matches to `tool-deployer`, novel ones to `tool-generator`. Those carry `approval_mode: auto`
  (38 complete, 2 deferred historically), so **they build without a further gate**. That is the real
  cost of this dispatch and it is on a live site.
- **No outward-facing side effect:** no `send_email`/`notify`/`deliver`/`publish` action exists in any
  of the three workflows. A regex over the step configs matched `tool-suggester` and `tool-generator`,
  and both matches were prose in prompt text ("...a suggested tool would deliver", "delivery cost
  estimator"). Checked because CLAUDE.md says to grep the handler before firing an operator action.

**Target chosen twice, and the first choice was wrong.** First filed against `webdesign.co.uk` (ours,
12 tools, previously evaluated). Then checked the queue rather than assuming: that site has **115
`triaged` page_rerenders and one `claimed`**, and the dispatcher takes one item per site at a time, so
the run would have sat behind all of them. **Withdrawn** (`status=cancelled`, never claimed, nothing
half-done) and re-filed against **`gamesdesign.co.uk`** — ours, **queue empty**, and the most tools
already deployed (7) among the idle candidates, which is the axis that minimises new suggestions.

**Verified the route end to end before waiting on it**, because a filed item that no mechanism collects
looks exactly like a slow one:

| gate | requirement | this item |
|---|---|---|
| `detected-item-promoter` (15 min, pure SQL) | pipeline in build/content/design | `build` ✓ |
| | handler is a live agent | `tool-suggester` active ✓ |
| | pair has ≥1 lifetime complete | 18 completes ✓ |
| | pair ≥25% success over ≥5 terminal | 18 complete / 0 failed ✓ |
| | 20 per tick, oldest first | 10 eligible ahead of it ✓ |
| `build-pipeline-trigger` (60 s) | status in triaged/approved | after promotion ✓ |
| | site not locked, no claimed item | both clear ✓ |

**The proof it must produce** (the "after" half of this bug's own disconfirming pair): the rendered
prompt must contain tools ranked **past 30** by `display_name`, ranked against the library **as it stood
at the prompt's timestamp**. For this site the eligible library is **79 tools, 49 of them past rank 30**,
so a post-fix prompt cannot pass by accident and a pre-fix one could not have passed at all.

### 2026-08-19 09:44Z — the dispatch: proof obtained, regression exposed, and three dead ends before the message landed

**Three failed dispatch attempts, all instructive:**

1. **The work-item route was a dead end on the clock.** Filed the item correctly (framework shape,
   `status=detected`), verified every gate — promoter doors all pass, 10 eligible ahead of it against a
   20-per-tick limit — and then read the *dispatcher's* ordering: `ORDER BY wi.created_at ASC … LIMIT 1`,
   **one site per 60 s tick, oldest work item first**. My item was the newest of **158**. ETA ~9 hours at
   the observed drain rate (18/hour) — and worse than that, `build-pipeline-trigger` was returning
   `complete_idle` while those 158 waited, because the two sites holding them were excluded (one had a
   live claim). **Checking that a queue will reach you is a different question from checking you are
   validly in it, and I only asked the second.**
2. **The documented direct-dispatch script is stale.** `140_tool_suggester/076_tool_suggester.sh` (March)
   uses `correlation_id:tool-eval-gas-001` — a non-UUID. `orchestration_states.correlation_id` is a `uuid`
   column now, and the message was dropped.
3. **A UUID was not enough either.** The chassis logged
   `validation/validator.go:81 "Incoming message missing required fields"` and discarded it.
   `ValidateIncomingMessage` requires **`client_id`, `correlation_id` AND `orchestration_id`**; the old
   script supplies only the first two. Working header set copied from `082_submit_domain_unified.sh`,
   which is current. **The tell was cheap and I nearly missed it: grep the chassis log for your own
   correlation id within ~60 s of publishing — 5 lines on one pod, and the warn was one of them.**

**Then it ran, and both results matter.**

**✅ The proof.** Prompt at 09:44:26Z: **80 of 81** eligible library tools present, **51 past rank 30**,
highest rank **81** (`Your notes from the old Noted`), truncation marker present. Against the recorded
"before" half — 29 tools, 0 past rank 30, highest exactly 30 — the disconfirming pair is complete and
275 is proven at the artefact.

**❌ The regression.** The same call died on `stop_reason=max_tokens`: output hit the step's own 3,000
cap, 11,090 chars recovered and discarded, step FAILED. **275's fix caused it** — 2.7× the menu means a
longer answer. Filed as `bugs_open/319`.

**The number that should have been read before migration 445 shipped:** across 59 historical
`suggest_tools` calls, output tokens ran 1,178–2,921 against a 3,000 cap — an all-time high-water mark of
**97.4%**, with 5 calls within 15% of the ceiling, *before* the library was widened. One query. This lane
measured the prompt per column to find that `description` was 80% of the payload, and never once asked
what the answer costs. **Bounding an input can move the failure to the output budget.**

**Blast radius of the dispatch itself: nil.** The step failed before `create_items_loop`, so **zero
`add_tool` items** were created on `gamesdesign.co.uk`. Both work items I filed are `cancelled`. The
platform failed closed — the `output_tokens == max_tokens` rule is why one run was enough to find this
rather than weeks of quietly odd suggestions.
