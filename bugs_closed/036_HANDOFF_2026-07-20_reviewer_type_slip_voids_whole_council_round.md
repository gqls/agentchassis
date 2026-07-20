# 036 — A reviewer emitting `"edit": "3"` instead of `3` voids the entire council round

**Filed:** 2026-07-20, from the bugfix-028 thread, when a council-gate submission
died at `council_decide` after all seats had reviewed.
**Severity:** Medium-high. No crash, no bad verdict — the round is simply *lost*,
after every seat has already been paid for. The submitter gets no verdict and no
objections, and the failure names a reviewer, which reads like the reviewer's fault.
**Status:** CLOSED 2026-07-20 — fixed (58f5a6bb6 + ab158c32a) and VERIFIED LIVE on v1.0.1140 by
forced reproduction; see §8. Originally filed as: OPEN, diagnosed from the live failure + code, no fix applied because the
fix is in `diagnose_council_decide_action.go`, which another thread had open at the
time (the bugs_open/019 work), and racing it in the same file is how same-file
passengers get committed by whoever commits first.

---

## 1. What happened

Council-gate submission `d35844da-f533-42da-b096-4f82cc2839bc`
(orchestration `5438f851-db9d-4f6a-bb32-3633e6793071`), a real bugs_open/028 fix.
The run reached `complete_invalid` with `error` **empty** and no `council_report`
row. The reason was in `collected_data.__step_error`:

```
step council_decide failed: failed to execute action diagnose_council_decide:
reviewer output at "review_debug_historian.result" does not match the review
schema: json: cannot unmarshal string into Go struct field .objections.edit of
type int
```

Every seat ran. `select_panel`, `review_editquality`, `review_debug_historian`,
`review_guardian` and the gates all executed. The credits were spent. One seat
wrote `"edit": "<something>"` where the struct wants an int, and **the whole round
was discarded** — including every other seat's valid review.

## 2. Mechanism

`platform/orchestration/actions/diagnose_council_decide_action.go:52`:

```go
type councilReview struct {
	Reviewer   string `json:"reviewer"`
	Verdict    string `json:"verdict"` // approve | object | veto
	Objections []struct {
		Edit     int    `json:"edit"` // 1-based index into the plan's edits; 0 = plan-wide
		...
```

`Edit` is a plain `int` and the reviewer output is strictly unmarshalled. An LLM
asked to identify *which edit* it objects to has three natural ways to answer —
`3`, `"3"`, or `"provider.go"` — and exactly one of them parses. The other two
take down the round.

**This is NOT bugs_open/019.** 019 is the *truncation* variant: the JSON is
incomplete and invalid, and 019's salvage path closes open brackets to recover the
verdict. Here the JSON is **complete and valid** — it simply has a string where an
int was expected. 019's salvage cannot help, because there is nothing to salvage;
the parse fails on a well-formed document. Same blast radius, different cause, and
a fix for one does not fix the other.

## 3. Why this is worse than it looks

- **It fails at the LAST step**, after every seat has been run and paid for. A
  truncation void at least fails a seat at a time; this one wastes the entire
  council.
- **The error names a reviewer**, so it reads as "the debug-historian is broken"
  rather than "the schema is brittle". The first instinct is to look at the seat.
- **`error` on `orchestration_states` is EMPTY** and `status` is `COMPLETED`. Only
  `collected_data.__step_error` and `current_step = 'complete_invalid'` say
  anything is wrong. A run that lost its entire output is recorded as completed —
  the "trust the artefact, not the status" invariant again.
- **It is nondeterministic**, so it will be intermittent and will look like flake.

## 4. Suspected aggravating factor from the submitter's side

The submission that triggered this had **five edits, three of which named the same
file** (`banana/provider.go`, for three distinct changes). A reviewer asked for a
1-based index into a list where three entries look identical has an obvious pull
toward disambiguating by name — i.e. toward a string. Unproven, but cheap to
respect: **one edit entry per file** in a submission removes the ambiguity. The
resubmission consolidated to three edits for exactly this reason.

That is a workaround, not the fix. The schema should not be able to lose a round.

## 5. Fix candidates

1. **Make `Edit` tolerant.** A custom `UnmarshalJSON` accepting `3`, `"3"`, and
   falling back to `0` (plan-wide) for anything unparseable. Smallest change,
   removes the whole class. `json.Number` would take `3` and `"3"` but still
   reject `"provider.go"`.
2. **Never let one seat's parse failure void the round.** Parse reviews
   individually; on failure record that seat as unusable (with the raw output kept
   for the report) and decide on the seats that DID parse, marking the round
   degraded. This is the structural fix and it also covers 019's class — the
   council's whole value is that it is many independent opinions, so one seat's
   malformed output should cost one seat, never all of them.
3. **Say so where it can be seen.** `error` must not be empty when a round is
   voided; `complete_invalid` with an empty error is indistinguishable from a
   clean finish at a glance.

(2) is the one worth doing; (1) is worth doing anyway because it is three lines
and stops the commonest instance immediately.

## 6. How to verify a fix

- Unit: feed `councilReview` a payload with `"edit": "2"` and one with
  `"edit": "somefile.go"`; both must parse, the second landing as plan-wide (0).
- Unit: feed a two-seat round where one seat's output is unparseable; assert a
  verdict is still produced from the other seat and the round is marked degraded.
- Live: submit anything through 097 and confirm a `council_report` row exists for
  the correlation. **The absence of a report is the symptom** — check for the row,
  not for an error.

## 7. Related

- `bugs_open/019` — one *truncated* reviewer voids a whole round. Same blast
  radius, different cause (invalid vs. well-formed-but-wrongly-typed JSON). Fix
  candidate (2) above would close both; 019's salvage path closes only 019.
- `bugs_open/017` (unregistered action, workflow invalid, marked complete) — the
  same "recorded as complete while having produced nothing" shape.
- The submission this blocked is the `bugs_open/028` fix.

---

## 7. NOTE 2026-07-20 12:05 — still not started, and the seam for candidate (2) now exists

Checked while clearing the backlog; **deliberately not fixed, for the reason §
Status already gives.** `diagnose_council_decide_action.go` is once again dirty in
the shared tree — a session is mid-work extending the 019 truncation path. Two
edits to one file cannot be split by a pathspec commit, so whoever commits first
takes both, and this file has already produced that damage once.

**What that in-flight work gives the fixing thread, and why it changes the shape
of candidate (2):** it introduces a `Degraded` flag on `councilReview`, set when
a review parses cleanly but the step recorded a truncation. So the "round is
degraded, decide on the seats that did parse" concept is no longer something
candidate (2) must invent — it is being built, for the adjacent cause. Candidate
(2) should extend that flag to cover a seat lost to a SCHEMA failure, not add a
parallel mechanism beside it. Two independently-invented degraded-round paths in
one function is the split-contract-drift shape this repo files bugs about.

Sequencing that follows: let the truncation work land, then do (1) and (2)
together on top of it. Doing (1) alone first is still safe — a custom
`UnmarshalJSON` on `Edit` touches only the type declaration — but it is three
lines and not worth a same-file race on its own.

**Unchanged and worth restating:** §4's "one edit entry per file" is a submission
habit that reduces the trigger rate. It is not a fix, and adopting it as one is
how this class stays open — the schema should not be able to lose a round.

---

## 8. FIX APPLIED 2026-07-20 (bugfix-036 thread) — INERT until the next image roll

**Status: still OPEN.** The bar for `/bugs_closed/` is fixed AND live; this is a Go
change, so the defect stays reproducible in prod until a chassis image ships.

**Sequencing gate from §7 is cleared.** The 019 truncation work landed
(`a3b606798`, then `11a72dc31` round 2) and `diagnose_council_decide_action.go` was
clean in the shared tree, so the same-file race §7 was avoiding no longer applied.

### What the live evidence changed about the diagnosis

The handoff guessed the reviewer would emit `"3"` — a malformed index. **It does
not.** All three voided rounds in the last 14 days carried a *plan-level
description* in `edit`:

| orchestration | `edit` value |
|---|---|
| `5438f851…` (this case) | `"plan-level (deploy verification)"` |
| `32a379cb…` | `"risks note on the 54 mis-stamped rows"` |
| `9f9970b8…` | `"risks/summary (item 5)"` |

All three name the **same seat** (`review_debug_historian`). This reframes the fix
rather than just confirming it: those reviewers were not emitting garbage — they
were saying *the objection is about the plan, not any single edit*, which the
contract **already spells `0`**. So the tolerant reading is not a leniency hack; it
recovers the meaning the strict one discarded. It also kills §5's `json.Number`
suggestion outright: `json.Number` would have parsed **none** of the three.

Measured cause split over 14 days of `complete_invalid` runs — worth having,
because 016b §9's own correction says to count the layer before fixing it:

| failed step | cause | n |
|---|---|---|
| `review_editquality` (upstream) | truncation, 019's dominant path | 9 |
| `council_decide` | **schema slip (this bug)** | **3** |
| `council_decide` | truncation | 2 |
| `persist_submission` / `review_guidelines` | other | 2 |

### The fix

Both candidates, as §5 and §7 sequenced them, in
`platform/orchestration/actions/diagnose_council_decide_action.go`:

1. **(1) Tolerant pointer.** `Edit int` → `objectionEdit{Index int; Raw
   json.RawMessage}` with an `UnmarshalJSON` that *never returns an error*:
   accepts `3`, `3.0`, `"3"`, `" 4 "`, and lands free text / `null` / an object on
   `0` (plan-wide). `MarshalJSON` round-trips the reviewer's **own token**, so the
   persisted `council_report` shows what was actually written instead of a
   laundered `0` — the report is read by humans judging whether the council was fair.
2. **(2) Structural.** The three remaining `return nil, err` exits in the per-seat
   loop — `planBytes` failure, schema mismatch on **valid** JSON, and an
   **unrecognised verdict** — now route into the `unreadable` mechanism 019 already
   built, via `salvageMistypedReview`, a sibling of `salvageTruncatedReview`. Per
   §7 this **extends** that seam rather than adding a parallel degraded-round path.
   An unrecognised verdict is deliberately **not** normalised to the nearest legal
   one: guessing what a seat meant is how a veto becomes an approval.

The contract is now uniform across every per-seat failure mode:

```
recover the opinion if it is there   → count it, marked Degraded
no opinion recoverable               → count the seat UNREADABLE
zero readable opinions in the round  → THEN fail — the one state that must
                                       not become a verdict
```

The safety property the old hard error protected is kept, enforced where it bites:
an unreadable seat downgrades `approved` → `revise`, so the change can only ever
make an outcome **more conservative, never less**.

### Verification (§6's two tests, plus what testing taught)

`go test ./platform/orchestration/actions/` — all green, five pre-existing council
tests unchanged. New: `TestObjectionEditTolerance` (built on the three **real**
payloads above, asserting `Index` *and* that `Raw` round-trips),
`TestMistypedEditDoesNotVoidTheReview`, `TestSalvageMistypedReview`,
`TestOneBadSeatDoesNotVoidTheRound`.

> **Wrong turn, recorded:** the first `TestSalvageMistypedReview` asserted that a
> mistyped `severity` deep inside an objection loses the whole objection. It
> failed — `encoding/json` **continues past a TYPE error** (unlike a syntax error)
> and keeps everything that did decode, so the objection survives with only the
> bad field zeroed. Better than assumed, and load-bearing: it is why the
> per-field salvage retains as much as it does. The test now pins the real
> behaviour and the code says why the ignored errors are doing more work than
> they look.

### Deliberately NOT done — §5 candidate (3)

`error` staying empty on a voided round is **not** a property of this action. The
workflow routes a step failure to the terminal `complete_invalid` step and the
orchestration then genuinely *completes*, so `error` is empty by engine design
(`coordinator.go` `routeToErrorStep` stores the real reason at
`collected_data.__step_error`). Fixing it is a workflow-seed / engine question with
a fleet-wide blast radius, not a council one — it wants its own bug, and it now
matters less because this class should no longer void rounds. Until then, the
diagnostic remains: **absence of a `council_report` row is the symptom, not `error`.**

### Council gate: corr `80cdd428`, round 1 — REVISE (7 approve / 2 object)

The round **completed without voiding** (`reviewers 9, abstained 7, unreadable 0`)
— the gate reviewing a fix for its own worst habit, through the code that still
has the bug. All four "confirm X" objections were checkable and are answered with
evidence in the 019 workstream NOTES: the three struct fields pre-exist
(`:73,74,79` — my sketch abbreviated the diff), 019's machinery is real
(`markerFieldFor:151`, `salvageTruncatedReview:173`), there is **no** downstream
consumer of `objections[].edit`, and the three types are unexported so the blast
radius is this step alone.

**One objection found a real gap and it is fixed** (edit-quality, low):
`salvageMistypedReview` dropped the entire objections list when `objections`
arrived as a single object instead of an array — the same wrong-register slip one
level up. I had asserted that was acceptable in a test. It is not: the objection
list is what goes back to the proposer in a revise round, so dropping it costs the
round its content even though the verdict survives. Now coerced to a one-element
list, with tests for the coercion and for a genuinely uncoercible payload.

No round 2 (the gate has no reviser loop; objections go to the human, which is
that record), and **no `Council-Reviewed` trailer** — earned by APPROVED only.

### VERIFIED LIVE 2026-07-20 on v1.0.1140 — reproduced, and the round survived

**Status: CLOSED.** The defect is no longer reproducible in production. I tried to
reproduce it deliberately and got a completed round instead of a void.

**1. The binary carries the change** (discriminating — symbols this change
*created*, plus a negative control, on pod `agent-chassis-5567d99bd6-5snzn`):

```
salvageMistypedReview        2
objectionEdit                7
ZZZ_definitely_not_a_symbol  0     <- the grep discriminates
```

**2. Forced reproduction.** A scratch copy of the council
(`council-gate-036scratch`; live councils untouched, verified `patched=f` on the
real `council-gate`) with the always-on `review_editquality` seat instructed to
emit `edit` as free text. Probe correlation `4031d8b4-0e90-4d39-b69d-9cd9ce332bc3`,
orchestration `11013b21-2915-449d-b741-e6326b88a312`. The seat complied and emitted:

```json
"edit": "edit 1 (comment-only change to diagnose_council_decide_action.go)"
```

That is the exact shape that produced `json: cannot unmarshal string into Go
struct field .objections.edit of type int` and destroyed three rounds. Result:

| | before (3 live cases) | now |
|---|---|---|
| terminal | `COMPLETED @ complete_invalid` | **`COMPLETED @ complete_revise`** |
| `__step_error` | the unmarshal error | **none** |
| `council_report` | absent — the symptom | **written** |
| seats counted | none | `reviewers 8, abstained 8, unreadable 0` |

**3. The report round-trips both registers.** Persisted `edit` tokens in that one
report: `["edit 1 (comment-only change to diagnose_council_decide_action.go)", 1]`
— a string and an int side by side, each in its original form. That is
`objectionEdit.UnmarshalJSON` tolerating the string *and* `MarshalJSON` preserving
the reviewer's own token rather than laundering it to `0`.

### Residual: fix candidate (2)'s salvage path is NOT proven live

`salvageMistypedReview` did **not** execute — the report shows no degraded seat
(`$.reviews[*] ? (@.degraded == true)` → `[]`). The probe also patched the
`review_guardian` seat to emit `objections` as a single object and `missing` as a
string, and **guardian refused to comply**, in terms worth quoting:

> "a reviewer that can be talked into corrupting its own output schema by text
> embedded in the thing it's reviewing is a bigger hole than the one bugs_open/036
> is patching"

It had misread the source — the instruction was in its own patched prompt, not in
the submission — but the refusal is the right instinct and it makes this harness
unreliable for seats with a strong security posture. So candidate (2) rests on unit
tests (`TestSalvageMistypedReview`, `TestOneBadSeatDoesNotVoidTheRound`) and symbol
presence, not on a live occurrence. That is a **defence-in-depth generalisation**,
not the filed defect: 036 as filed is the tolerant-pointer case, and that is proven.
Anyone who exercises the salvage path in the wild should record it here.

### Blast radius (corrected)

`diagnose_council_decide` is shared by **five active pipelines** — `council-gate`,
`fix-proposer`, `experience-planner`, `feature-designer` (+ scratch copies):

```sql
SELECT type, is_active FROM agent_definitions
WHERE default_config->'workflow'->'steps' ? 'council_decide';
```

An earlier answer of mine implied one pipeline. The types are unexported so there
is no cross-package consumer, but every council on the platform shares this
decider — the fix is broader, and uniformly conservative.
