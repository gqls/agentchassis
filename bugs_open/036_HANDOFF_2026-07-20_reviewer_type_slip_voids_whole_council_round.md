# 036 — A reviewer emitting `"edit": "3"` instead of `3` voids the entire council round

**Filed:** 2026-07-20, from the bugfix-028 thread, when a council-gate submission
died at `council_decide` after all seats had reviewed.
**Severity:** Medium-high. No crash, no bad verdict — the round is simply *lost*,
after every seat has already been paid for. The submitter gets no verdict and no
objections, and the failure names a reviewer, which reads like the reviewer's fault.
**Status:** OPEN. Diagnosed from the live failure + code. **No fix applied** — the
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
