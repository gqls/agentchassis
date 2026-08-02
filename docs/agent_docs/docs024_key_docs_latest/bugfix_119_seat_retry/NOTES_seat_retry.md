# NOTES — `bugs_open/119` seat retry (append-only, newest at the bottom)

## 2026-08-01 — session 1

### Picking the bug, and the two I put back

Swept `bugs_open/` against every session transcript modified in the last 5 hours,
grepping each for `bugs_open/NNN` to find what is genuinely in flight. That found 37
numbers under active work. Two candidates looked good and both were **put back after
reading**, which is worth recording because neither was obvious from its title:

- **071** (validate gate discards broken-link findings) — its own triage section says it
  now "bundles one **closed** mechanism with three open ones, which is the shape that
  makes a bug file un-closable forever", and recommends a split that the *owning lane*
  should pick. Taking it would have meant either closing it dishonestly or re-filing
  someone else's numbers.
- **093** (stat audit has one guarded call site) — reads as a perfect framework-scope
  bug, and is: but its last triage says *"093 is not a code task any more. It is blocked
  on `bugs_open/083`… Do not spend a chassis roll on it."* 083 is under active work by
  two other sessions.

**Lesson worth keeping: the title and the STATUS line are not enough — the last dated
section of a bug file is where its real state lives.** Both of these read as OPEN and
tractable at the top and were neither, 400 lines down.

### The measurement that reframed the bug — and nearly did not happen

119's filed mechanism is a seat emitting a *complete but malformed* review (a stray `]`).
I was one step from implementing against exactly that when I checked whether it still
reproduces. It does not, and the check took four queries:

```sql
-- every unreadable seat, joined to the orchestration that produced it
WITH rpt AS (SELECT da.orchestration_id::uuid AS oid,
                    jsonb_array_elements_text(da.body::jsonb->'unreadable') AS seat_key
             FROM diagnosis_artifacts da
             WHERE da.kind='council_report' AND jsonb_typeof(da.body::jsonb->'unreadable')='array')
SELECT CASE WHEN seat IS NULL THEN 'orchestration pruned'
            WHEN (seat->>'__truncated')='true' THEN 'TRUNCATED'
            WHEN COALESCE(seat->>'result','')='' THEN 'empty, not marked'
            ELSE 'COMPLETE but unparsable (the filed class)' END, count(*)
FROM ( ... ) GROUP BY 1;
-- orchestration pruned 36 · TRUNCATED 3 · the filed class 0
```

Then the denominator, because "0 occurrences" is meaningless without one
(`a-count-you-kept-is-not-a-census`): across **785** JSON-declared step outputs in the
live orchestration window, **782 parsed, 2 unparsable (both truncated), 0 malformed**.

> **So the filed MECHANISM is unmeasurable and the filed DEFECT is entirely real.** The
> retry is written against the defect ("nothing re-asks an unusable answer"), not against
> the bracket slip, which is why its corrective text branches on which failure occurred.
> Had I not measured, I would have shipped a mechanism that fires on a class with zero
> live instances and reported it as a fix for 15 voided rounds a week.

**Also caught here:** 119's own §"How to verify a fix" step 1 says to reproduce from the
stored artefact and "capture it before ~2026-08-09" because "`orchestration_states`
retains 13 days". That retention claim is **wrong** — rows exist back to 07-13, but the
specific orchestrations were pruned within days. The instruction was already
unsatisfiable when I read it.

### Misstep: I was about to add a config key that already existed

My first design keyed the retry on a **new** step-config key, `retry_on_unparsable_json`.
Before writing it I grepped whether the key I planned to *reuse* for "this step needs
JSON" was live — per the `grep-the-config-key-before-calling-it-a-win` landmine — and
found the real defect underneath:

```
$ grep -rn '"output_format"' --include=*.go .
platform/orchestration/actions/database_actions.go:26   <- a DIFFERENT action
scripts/goscripts/workflow_validator/main.go:126        <- listed as an optional key
```

`execute_llm_prompt` never reads it. `getOutputType` reads **`output_type`**. Live census:
`output_type` on 6 steps, `output_format` on 100. **A new key would have added a second
inert key beside an existing one** — precisely `bugs_open/134`'s class, which I would
have created while fixing 119.

### The trap in the fix for that

`output_format` is **not this action's name to claim**: `query_database` reads the same
key with `array`/`object`. So the fallback is allow-listed to `json|text|html|markdown`
and an unrecognised value falls through to the default instructions — which is exactly
what it gets today, so the failure mode of the guard is "no change".

### The design decision I reversed, and what reversed it

I intended the re-ask to **raise `max_tokens`** when the first attempt truncated —
reasoning that an identical re-ask reproduces a truncation, which correlation `c5219a69`
proves (three consecutive rounds, same seat, same cap). Reading the file I was about to
lean on stopped it:

> `platform/aiservice/truncation.go:26-29` — *"NOTE this is NOT a reason to raise output
> caps and call the class fixed: experience-planner/compose truncated at a 32,000-token
> cap, 4x the council seats' 8,000. Whatever the number, the step that writes most
> approaches it on the work most worth doing."*

So the re-ask asks for the same judgement **shorter** instead of buying headroom the next
long review eats anyway. *What caught it:* reading the header of the file whose function
I was calling, rather than only its signature.

### Tests: passing proved nothing until I broke them

Per `a-quiet-test-passes-when-the-rule-is-gone` and `mutate-the-code-to-prove-the-guard`,
both behaviours were verified by mutation:

| mutation | expected | observed |
|---|---|---|
| remove the `output_format` fallback | 4 tests fail | 4 failed, restore → pass |
| remove the vocabulary allow-list | the gate test fails | failed, restore → pass |

One test also failed **honestly on first run** and it was right to: my truncation re-ask
wrapped "cut words, never findings" across a newline, so the phrase was split. I reflowed
the *prompt* rather than weakening the assertion — the phrase is the instruction's point.

### Build hygiene

`go build ./...` is green in the working tree, but the tree carries four other sessions'
uncommitted edits in this same package, so a green tree is not a green HEAD
(`a-shared-tree-commit-can-break-head`). Verified properly: `git archive HEAD` into a
clean dir, overlay **only** my two files, `go build ./platform/... ./internal/...` → exit
0, and the package tests pass there too. (The `cmd/` link steps failed on `no space left
on device` — a full `/tmp`, not a compile error. Cleaned up after.)

---

## 2026-08-01 — council round 1: REVISE, and the seats caught a real hole

`576832f3` → **REVISE**, gating objection from `debug_historian` (high). 14 reviewers.
**They were right and I had not checked the thing they gated on.**

### The gating objection

Four seats converged (debug_historian high, guardian medium, editquality medium): my
"90 steps across 32 agents" census **assumed** `output_format` sits at a depth
`getOutputType` actually reads, citing the house landmine that a step's prompt and its
token cap sit at *different* depths in `default_config`. I had asserted the depth and
never measured it — the same species of error as this file's first entry, one layer down.

**Measured properly** (RUNBOOK R5, extended to walk `sub_workflow` too):

| depth | LLM steps | `output_format` at `config` | under `config.ai_service` | at step root |
|---|---|---|---|---|
| top-level | 134 | 100 | **0** | **0** |
| nested in `sub_workflow` | 1 | 1 | **0** | **0** |

So the assumption held — **and the re-measurement corrected my numbers anyway**, because
my first census walked only top-level steps: **135 LLM steps, 101 `output_format`, 91 of
them json across 33 agents** — not 90/32. The missing one is
`page-content-writer` → `process_sections_loop` → `generate_content` (verified by query,
not inferred; I had guessed that name and then checked it).

> **The transferable bit: a census over `agent_definitions` that walks only
> `workflow.steps` is blind to every step inside a loop's `sub_workflow`.** That is what
> WFA-003 exported `WalkSteps` for. My SQL did not use it, and a one-step miss here would
> have been a 91-step miss on a different agent.

### The objection I answered with code rather than argument

`bug_historian` (medium): the re-ask lowers the **frequency** of the silent success but
not its **shape** — the step still returns `{result: text, type: text}` and still
SUCCEEDS. That is `bugs_closed/076`'s title and the `missingkey=zero` family, and it was
a fair hit: I had reduced a rate and called it a fix.

Answered with `__json_contract_unmet`, stamped only when json was **declared** and, after
the re-ask, still not delivered. **A marker, not a hard error** — making those 91 steps
fail loud would convert currently-succeeding steps into failing ones over content they
did not author, which is `bugs_closed/073`'s defect and the exact reason 119's own
candidate 2 was declined. Keep the behaviour, end the silence.

### A fourth specimen, delivered by the review itself

Round 1's own council report carried `unreadable: ["review_llm_reliability.result"]`. I
captured it before pruning:

```json
{"type":"text","result":"","__truncated":true,"__truncated_output_tokens":8000}
```

**Four in-retention specimens now, four truncations, zero bracket slips.** It also shows
the treadmill in one picture — the per-seat caps today:

```
16000: review_editquality, review_guidelines, review_prior_art, review_architecture
 8000: the other THIRTEEN, including review_llm_reliability
```

The four at 16000 are exactly the four that had previously failed. `138`'s lane raised
each cap *after* its seat broke, and round 1's failure landed on one of the thirteen left
behind. That is `truncation.go:26-29`'s warning happening in front of me, and it is the
strongest argument for the re-ask asking for **brevity** rather than headroom.

### One objection I pushed back on, with a reason

`architecture` (medium): a new retry contract on shared plumbing, same shape as
`bugs_closed/129` shipping without an RFC. Answered rather than conceded, because
`ExecuteLLMPromptAction` **already** makes up to five calls via the 500/502/503/529
backoff ladder ~100 lines above the parse. "One step, one call" is not a guarantee that
exists today, so this extends an existing ladder rather than introducing retries to a
mechanism that had none. The seam is registered (WFA-005) per condition (2) of the
ordering exemption, and I claim no ordering constraint.

### Two things I got wrong in the SUBMISSION rather than the code

- **The pod-grep was planned and I never said so.** `debug_historian` (medium) marked
  verification as "unit tests + mutation only". RUNBOOK R10 had the pod-grep with a
  negative control from the start; the plan block did not mention it. A reporting gap
  reads exactly like a missing step, and a reviewer cannot tell them apart.
- **Round 1 named 32 agents; there are 33.** I named them in full (which the owner ruling
  requires) off a census that was one step short. Naming consumers is only as good as the
  walk that enumerated them.

---

## 2026-08-01 — council round 2: APPROVED, and six objections that were my own handwriting

`576832f3` round 2 → **APPROVED**, 14 reviewers, **0 unreadable**, 8 advisory objections,
none high-severity. Verdict read and dispositioned rather than filed.

### The one that was a real catch — acted on

`diagnosis_guardian` (medium): my truncation re-ask says *"cut words, never findings"*,
which protects the **conclusions** and says nothing about what **supports** them. A
verdict-emitting step re-asked for brevity could keep its findings and shed the citations
that make them checkable — a shorter answer becoming a **less evidenced** one, which is
worse than the lost round it was meant to prevent. Added *"Keep every citation, quote, id
and file:line that supports a finding"* and *"Cut explanation, never evidence"*, with two
assertions pinning both. I would rather have thought of this myself; "be brief" is exactly
the instruction that goes wrong at the evidence.

### The one I refuted, by reading rather than arguing

`llm_reliability` (medium): claimed `GenerateText` does not decode `stop_reason`, so a
truncated re-ask would come back as a normal 200 and slip past my `if reaskErr != nil`
guard. **Checkable, and false:** `anthropic.go:217-227` decodes it and returns a
`*TruncatedError`; `gemini.go` and `ollama.go` do the same. The guard is correct, and it is
the identical shape the first attempt already uses — had the premise been true, the primary
path would have had the same hole since 019.

### The artefact: six seats objected to a landmine I wrote an hour earlier

`editquality`, `reuse_agent`, `guidelines`, `debug_historian`, `prior_art_librarian` and
`architecture` — **six of the eight objections** — all said the plan "fails to cite or
reconcile a live landmine" keyed to `getOutputType`/`output_format`/`llmOutputVocabulary`,
warning that `output_format` *"now BUYS A SECOND LLM CALL"*.

**That landmine is mine.** I appended it in commit `3af110705` and ran
`landmines-sync.py --apply`, which publishes to `doc_notes` immediately — and the seats read
`doc_notes` at review time. So documenting the seam, exactly as CLAUDE.md requires, armed
six reviewers against the submission that created it.

```bash
$ git log --oneline -S'output_format` means two different things' -- .../LANDMINES.md
3af110705 landmine(output_format): one key name, two vocabularies, and it now buys a second LLM call
```

Nothing was blocked — the round approved. The cost is **signal dilution**: three-quarters
of the review's advisory capacity spent restating my own documentation back at me, which is
capacity not spent on the design. And it is **systematic, not bad luck** — the more
faithfully a thread follows "register the seam in the same commit that ships it", the more
objections it draws.

Written up as a transferable pattern in `016b` §9, with the `git log -S` check (verified,
not asserted — the command above is the run) and the unbuilt candidate fix: carry the
introducing commit on a landmine so a seat can suppress one authored by the submission under
review. Explicitly **not** fixed by delaying the landmine until after the verdict, which
would trade a review artefact for an undocumented seam.

> The general shape is one this fleet already knows in its inverse form. MEMORY has
> `prompt-text-poisons-its-own-detector` and `declaring-a-key-silences-your-own-detector` —
> both "your own writing enters the corpus your detector reads", and in both the author's
> text **silenced** a check. Here it **armed** one. I did not recognise it until the sixth
> seat said the same sentence.

### Status at end of session

Committed `059ae2acd` with `Council-Reviewed: 576832f3-…` — the trailer is earned, the
verdict was read. **The bug stays OPEN.** Go changes are inert until an image is built and
rolled, and CLAUDE.md's bar is *fixed AND live*. RUNBOOK R10 has the pod-grep with its
negative control and the behavioural `llm_call_log` check; neither can run until the roll.

---

## 2026-08-02 — post-roll: LIVE, and two measurement errors caught on the way to proving it

**Roll.** Chassis `v1.0.1228`, both replicas started 08:47 UTC, carrying `4c1a97874`,
`085160a6c`, `059ae2acd`. Confirmed the commits are ancestors of HEAD first
(`git merge-base --is-ancestor 059ae2acd HEAD`) — HEAD had moved 8 commits ahead under other
lanes, and my three files were untouched by them.

**Artefact proof, both pods** (RUNBOOK R10, now rewritten). All five added markers present
1–3×, pipeline control 1, spelling control 0. `Keep every citation` is the load-bearing one:
it entered the tree only in `059ae2acd`, the last commit of the series, so its presence dates
the image past the whole series.

**MISSTEP 1 — my own negative control could never have failed.** R10 as written grepped the
binary for `adds format-specific instructions based on output_type`, "a string this change
REMOVED". It is a **Go comment**. Comments are not compiled, so it returns 0 against every
binary ever built and reads as a pass. Checking properly
(`git diff 4c1a97874^..059ae2acd -- <file> | grep '^-'`) showed this change removed **no
string literal at all** — only comments and code structure — so the removal-based negative
control `bugs_open/153` asks for was never available here. Substituted the last-commit marker
above, which answers the same question (is the image new enough) by a route that can fail.
**Check a control CAN fail before trusting that it didn't.**

**MISSTEP 2 — I measured my own submission text and nearly filed it as the fleet's
behaviour.** To show edit A's defect was real I counted, over all calls to
`output_format: json` steps, how many prompts contained the JSON instruction:

```sql
count(*) FILTER (WHERE prompt_rendered LIKE '%Ensure valid JSON syntax%')  -- 31 of 9,063
```

31 looked like a plausible trickle of correctly-instructed calls. It was not: all 31 were
dated to my two council rounds, at **exactly 2 per seat** across 16 seats, and the phrase sat
inside my own prose — the submission rationale quotes the instruction it is about, and a
submission is rendered into every seat's prompt. `prompt_template` matched **0** times, which
alone falsifies the reading; and the alternative explanation (an `ai_service`-level
`output_type` the old code could already see) is refuted at **0 rows**.

Re-measured on the block **header**, which no rationale quotes in passing:

| era | calls | `CRITICAL OUTPUT FORMAT - JSON:` | `CRITICAL OUTPUT FORMAT:` (says nothing about JSON) |
|---|---|---|---|
| before roll (2026-03-31 → 08-01, 25 agents) | 9,063 | **2** | **9,061** |

Positive control, and it sharpens the finding: steps declaring `output_type: json` — the
spelling the *old* code could read — have had **0 calls** in the entire retained window. So
the JSON block was essentially never appended by anything, ever. The corrected figure is more
damning than the wrong one, which is why it was worth re-doing rather than shipping the 31.

**What the after-column is not.** Empty. `[UNMEASURED — 0 denominator]`: in the 32 minutes
after the roll the fleet made **4** LLM calls, all `med-price-collector.scrape_prices`, which
runs `med_scrape_prices` (not `execute_llm_prompt`) and declares no contract — so it is
correctly absent, not a counter-example. Fleet is idle at single-digit calls/hour and the
last council round was 2026-08-01 08:59:45, my own round 2. Hence the re-ask has fired 0
times out of ~0 opportunities, and the damage figure is unchanged at **23 voided of 429**
(weekly 0/0/8/15). None of that is evidence either way and it is recorded as such.

**A close condition I set too strictly, recorded rather than dropped.** The handoff's step 4
said the damage figure falling "is the close condition". That is an *outcome* metric gated on
other sessions convening councils — unsatisfiable by anything this lane does, possibly for
days — whereas CLAUDE.md's bar (*fixed AND live*) is met and artefact-proven. Closed on the
stated bar with the outcome measurement written up as an owed follow-up, both queries inline
in the bug file.

**Landmine verifier still not fired**, same reasoning as 2026-08-01: `bugs_open/163` (OPEN,
unowned) records that the verifier's symbol lookup cannot answer a path-bearing query and
invents a stale-index cause for its own blindness. This entry's footprint is path- and
table-bearing, so a false negative is the likely outcome and banking one would degrade the
corpus. Synced to `doc_notes` (700 owned rows); flagged `NEEDS_VERIFICATION` and left so.
