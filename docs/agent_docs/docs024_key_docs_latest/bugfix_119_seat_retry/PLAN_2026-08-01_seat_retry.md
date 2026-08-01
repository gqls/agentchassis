# PLAN — `bugs_open/119`: one seat's unreadable output voids a whole council round

**Started** 2026-08-01. **Bug:** `bugs_open/119_HANDOFF_2026-07-27_one_seat_emitting_malformed_json_costs_a_whole_council_round.md`
**Ownership checked before starting:** `who-owns.py 119` names no owning workstream —
the two ACTIVE lanes it lists (`bugfix_100_101_scrape_provenance`, `architecture_review`)
cite 119 in passing, and the only two commits ever to touch the file are its own filing
(`1d241db79`) and one contributed sighting (`556e39a13`). Live `.jsonl` transcripts of all
sessions active in the last 5 hours mention 119 **zero** times. The bug file says
"OPEN, unowned" and that still holds.

---

## 1. What the bug claims, and what I measured

119's filed claim: a seat emits a **complete but structurally malformed** review (an
object closed one bracket early, leaving a stray `]`), `ParseLLMJSON` correctly refuses
it, the decider correctly records the seat `unreadable`, and an approve is correctly
downgraded to revise — **but nothing retries the seat**, so a recoverable emission slip
costs a whole round.

Every step of that is still true of the code. But my own measurement **reframes which
mechanism is actually biting**, and that changes the fix.

### 1a. The damage is real, recurring, and worse than filed

`diagnosis_artifacts` where `kind='council_report'`, all history (424 rounds,
2026-07-10 → 2026-08-01):

| week | rounds | ≥1 unreadable seat | **voided by an unreadable seat** |
|---|---|---|---|
| 2026-07-06 | 10 | 0 | 0 |
| 2026-07-13 | 85 | 0 | 0 |
| 2026-07-20 | 162 | 17 | **8** |
| 2026-07-27 | 167 | 21 | **15** |

**23 rounds all-time were decided by a parse failure rather than by a judgement.** The
filing measured "2 of the last 18"; the true all-time figure is 23, and the worst day was
2026-07-28 (8 of 48 rounds, 17%). Seven distinct seats have done it, so it is roster-wide
emission fragility, not one bad prompt:

```
review_editquality 25 · review_guidelines 4 · review_prior_art 3 · review_guardian 3
review_deferral_honesty 2 · review_checkability 1 · review_improvement_guardian 1
```

The sharpest instance is 2026-07-31: correlation `c5219a69` burned **three consecutive
rounds** (15:32, 15:37, 15:42), same seat, same orchestration. That is the cost shape —
not one wasted round but a submission that cannot get a verdict at all.

### 1b. [MEASURED] The live trigger is NOT the bracket slip 119 documents

I classified every unreadable seat against the surviving orchestration rows. Of 39
seat-instances, 36 orchestrations are already pruned; **all 3 survivors are total
truncations with an EMPTY partial**:

```json
{"type":"text","result":"","__truncated":true,"__truncated_output_tokens":8000}
```

And fleet-wide, across every step that declares it needs JSON (785 outputs in the live
orchestration window):

| result | occurrences |
|---|---|
| parsed as JSON | 782 |
| unparsable — **truncated** | 2 |
| unparsable — **complete but malformed (119's filed class)** | **0** |

> **So: 119's filed MECHANISM has zero live occurrences I can measure, while 119's filed
> DEFECT — nothing re-asks a seat — is 100% true of the code and is costing rounds every
> week.** Both statements are needed. A fix aimed only at the bracket slip would be a
> mechanism rotting unexercised; a fix aimed at the *defect* covers both.
>
> The one bracket-slip specimen that exists is quoted verbatim in the bug file itself.
> The orchestration it came from is long pruned, so the filing's own §"How to verify"
> step 1 ("reproduce from the stored artefact… capture it before ~2026-08-09") is
> **already unsatisfiable** — the retention assumption in that sentence was wrong.

### 1c. The boundary with `bugs_open/138`, which is actively worked

138 is *salvaged* truncation → `degraded: true` → gates unconditionally. 119 is
*unsalvageable* output → `unreadable` → voids the round's approval. Different branches of
`diagnose_council_decide_action.go`, different fixes. My change touches neither
`salvageTruncatedReview` nor the `Degraded` gating rule, so it cannot collide.

---

## 2. Root cause, read first-hand in the code

`ExecuteLLMPromptAction` (`platform/orchestration/actions/ai_actions.go`):

- **`:598-619`** — a transient HTTP error (500/502/503/529) gets a **four-attempt retry
  ladder** with backoff.
- **`:698-711`** — a response that comes back and **will not parse** gets **no retry at
  all**. It is packaged as `{"result": <raw text>, "type": "text"}` and the step
  **succeeds**.

That asymmetry is the defect, stated structurally: *the platform retries a call that never
arrived, and accepts without question a call that arrived unusable.* The council decider
then does the only thing it can with a text blob — records the seat unreadable — and 019's
downgrade contract (correctly) refuses to let an approve stand beside it. The round is
spent.

### 2a. A second, upstream cause found while reading — a dead config key

The step config that declares intent is `output_format`. **Nothing in the LLM path reads
it.** `appendOutputInstructions` → `getOutputType` reads **`output_type`** (`:891-896`).

[MEASURED] across all 134 active `execute_llm_prompt` steps:

| key | value | steps | agents |
|---|---|---|---|
| `output_format` (**never read**) | json | **90** | 32 |
| `output_format` (**never read**) | text | 10 | 5 |
| `output_type` (read) | json | 5 | 5 |
| `output_type` (read) | text | 1 | 1 |

So **90 steps across 32 agents — including every council seat — ask for JSON and are
never told to produce it.** They fall through to `getDefaultOutputInstructions()`, which
says nothing about JSON. The instruction they are missing is:

> `- Ensure valid JSON syntax (proper quotes, commas, brackets)`

which is precisely the failure 119 documents. [INFERRED — I cannot prove this instruction
would have prevented the observed stray `]`; what is measured is that the seats never
receive it.] This is the same class as `bugs_open/134` (a config key that is declared and
inert) and the `grep-the-config-key-before-calling-it-a-win` landmine.

---

## 3. The fix — prevention upstream, recovery downstream

Two edits. Both are inside the existing machinery; neither adds a reserved key or a new
namespace, so neither is a platform seam under the 2026-07-28/29 rulings.

**Edit A — prevention. Make the declared key live.** `getOutputType` falls back to
`output_format` when `output_type` is absent, honouring **only** the LLM output
vocabulary (`json`/`text`/`html`/`markdown`). The allow-list is load-bearing, not
decoration: `query_database` uses the *same key name* with a *different vocabulary*
(`array`/`object`, `database_actions.go:26`), so a blind pass-through could one day select
a wrong instruction set. Under this, the 90 JSON-declaring steps finally receive the JSON
instructions.

*Blast radius, measured:* 782 of 785 (99.6%) of those steps' outputs **already** parse as
JSON, because their prompts already say "Output — ONLY this JSON". So this reinforces
what they already do rather than changing it.

**Edit B — recovery. Extend the existing retry ladder to cover an unusable answer.**
When the response fails to parse **and** the step declared it needs JSON, re-ask once.
Not a new mechanism: the same function already retries four times on transient errors.
On a truncated first attempt the re-ask is given headroom, because an identical re-ask
reproduces a truncation — proven by `c5219a69`, which failed three times running. If the
retry also fails, behaviour is **byte-for-byte what it is today**, so the 019 downgrade
contract and 138's salvage path are untouched.

*Cost, measured:* the retry would have fired **twice** across 785 outputs in the live
window. It can only fire on a path that is already producing an unusable result.

### Why not the other candidates
- **119 candidate 2 (provider-enforced structured output)** — the real door-closer, but it
  touches every seat and every provider adapter. Edit A is the cheap 80% of it.
- **119 candidate 3 (loosen the Tier 3 parser rule)** — the file recommends against it and
  is right: that rule is backed by a 5,844-response measurement. Not touched.

## 4. How this gets verified (NOT by a green round)

A green council round proves nothing — the healthy path already works. Required:
1. Retry **fires and recovers** on an unparsable-then-parsable response.
2. Retry **does not fire** for a step that did not declare JSON (no cost regression).
3. **The 019 contract survives** — a still-unreadable seat STILL downgrades approve→revise.
   Proven by mutation, not by a passing test (`mutate-the-code-to-prove-the-guard`).
4. Pod-grep a string this change ADDs *and* one it does not, on every replica.

## 5. Decisions and corrections

- **2026-08-01, initial sizing was wrong and I caught it before writing code.** I planned
  to key the retry on a **new** config key (`retry_on_unparsable_json`). Reading the config
  first showed `output_format` already exists, is already declared on exactly the right 90
  steps, and is already dead — so a new key would have added a second inert key beside an
  existing one. *What caught it:* grepping whether the key I was about to reuse was read
  by anything, per the `grep-the-config-key-before-calling-it-a-win` landmine.
- **2026-08-01, the bug's headline mechanism is not the live one, and I nearly shipped
  against the headline.** Only measuring the fleet-wide population (§1b) showed the
  bracket slip at zero and truncation at two. Recorded rather than quietly re-aimed,
  because the bug file will be read again by someone else.
