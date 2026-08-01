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
