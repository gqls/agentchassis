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
