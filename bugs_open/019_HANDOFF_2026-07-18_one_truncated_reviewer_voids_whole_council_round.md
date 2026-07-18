# 019 — One truncated reviewer voids the entire council round

*Found 2026-07-18 by the claims-verification thread, submitting its own V4
change through the council gate. Affects `diagnose_council_decide`, so it hits
**every** council: `fix-proposer`, `feature-designer`, and `council-gate`.
Not fixed here — the action belongs to the fixloop/council threads, and the
council-gate runbook warns to diff any seed against the LIVE row first.*

## The defect

If a single reviewer's LLM output is cut at `max_tokens`, its JSON is invalid,
and `diagnose_council_decide` returns a hard error. The orchestration lands on
`complete_invalid` and **every other reviewer's completed, well-formed review
is discarded**. No verdict, no council report, no revise round — the whole run
is void, and the credits for all seats are spent.

The failure is silent about its own cause at the point a human looks: the run
shows `COMPLETED @ complete_invalid`, which is the same terminal step used for
a malformed *submission* — a completely different problem (see §"Don't confuse
these two" below).

## Evidence (this run)

Submission correlation `c9ca40d5-73c2-4fcf-a6e2-d8ee12e7bf60`,
orchestration `f1b6a9a9-a069-4b7b-8e01-cac280ec6213`.

```
step council_decide failed: failed to execute action diagnose_council_decide:
reviewer output at "review_guidelines.result" is invalid JSON —
likely truncated at max_tokens: unexpected end of JSON input
```

`llm_call_log` for that round — the CLAUDE.md rule (`output_tokens ==
max_tokens` means CUT) identifies the culprit exactly:

| step | out_tokens | max_tokens | cut |
|---|---|---|---|
| review_editquality | 4596 | 8000 | no |
| review_reuse_agent | 3220 | 8000 | no |
| **review_guidelines** | **8000** | **8000** | **YES** |
| review_tooling_provenance | 2738 | 8000 | no |
| review_compliance | 4229 | 8000 | no |
| review_debug_historian | 5548 | 8000 | no |
| review_guardian | 4856 | 8000 | no |

Six complete reviews (including the guardian's 4,856 tokens and the
bug-historian's 5,548) were thrown away because a seventh overran.

**This is not a rare tail.** Four of the seven seats used more than half the
ceiling on a single submission. The margin is thin and shrinks as plans get
richer — the same round for a smaller submission earlier that day ran
441–2,040 tokens per seat, so the ceiling only bites on substantial changes,
which are exactly the ones most worth reviewing.

## Root cause

`diagnose_council_decide_action.go` ~line 99–126. The loop already has a
principled abstention path for a reviewer whose field is **absent** — the
stage-3 relevance filter skips irrelevant seats, and the code reasons
(correctly) that a skipped seat did not object, so it must not gate:

```go
if raw == nil {
    // ... a skipped seat is an ABSTENTION, not a failure ...
    abstained++
    continue
}
```

But a reviewer whose field is **present and unparseable** falls through to:

```go
if !json.Valid(rb) {
    return nil, fmt.Errorf("reviewer output at %q is invalid JSON — likely truncated at max_tokens: %w", field, err)
}
```

An all-or-nothing hard error. The action distinguishes "seat didn't run" from
"seat ran and produced garbage", which is right — but it handles the second
case by destroying the work of every other seat.

## Fix candidates (in preference order)

1. **Treat an unreadable reviewer as a LOUD abstention, and never approve on
   incomplete information.** Count it separately from relevance-filter
   abstentions (`unreadable`, not `abstained`), log at Warn with the field
   name, carry it into the decision object and the council report so it is
   auditable — and if the surviving reviews would produce `approve` while any
   reviewer was unreadable, **downgrade to `revise`**. This preserves the
   safety property the current hard error is protecting ("a council that cannot
   read its reviewers must not wave a plan through") without discarding six
   valid opinions, and it keeps an objection from a *readable* seat decisive.
   An unreadable seat must never be able to turn an objection into an approval.

2. **Salvage the truncated review before giving up.** The platform already has
   `repairTruncatedJSON` (`apply_adoption_plan_action.go`) for exactly this
   shape. A truncated review usually has its `verdict` and `reviewer` fields
   intact — those come early in the object — so a best-effort repair often
   recovers the load-bearing part and only loses trailing prose. Combine with
   (1) as the fallback when repair fails.

3. **Raise the reviewer ceiling and/or cap reviewer prose.** `max_tokens` 8000
   is set per reviewer step; the reviewers are also not told to be terse.
   Either lift the ceiling or instruct seats to bound `notes`/`objections`
   length. This alone only moves the cliff — do it *with* (1), not instead.

## Don't confuse these two

`complete_invalid` is reached from two very different failures, and they need
different responses:

| Reached from | Meaning | What to do |
|---|---|---|
| `persist_submission` | **Your submission** was malformed (`diagnose_persist_fix_plan` validation: operation not in `modify\|add\|remove\|config_change`, a path appearing in two edits of one stage, >8 edits, bad `artifact_role`) | Fix the submission JSON and resubmit |
| `council_decide` | **A reviewer's output** was truncated — your submission was fine and was fully reviewed | Nothing to fix in the plan; this bug |

Read `collected_data->'__step_error'->>'message'` to tell them apart. The
`failed_step` field names which one.

## How to verify a fix

1. Reproduce cheaply: temporarily set one reviewer step's `max_tokens` to ~200
   in a scratch copy of a council definition and submit any valid plan. Current
   behaviour: `complete_invalid`, no verdict. Fixed behaviour: a verdict that
   records the unreadable seat and does not approve.
2. Confirm the surviving reviews are honoured — an `object` from a readable
   seat must still produce `revise`.
3. Confirm the count appears in the decision object and the council report:
   `SELECT collected_data->'council_decide' FROM orchestration_states WHERE ...`
4. Re-run this thread's actual submission (correlation above) and expect a real
   verdict rather than `complete_invalid`.

## Related

- `bugs_open/016` — a different council defect (revise/reframe prompts render
  reviewer output as `<no value>` via `UnwrapDeep`). Same subsystem, unrelated
  mechanism. **A run hitting both would revise blind AND lose a round.**
- `bugs_open/005`, `008`, `012` — the truncation family. This is the same root
  hazard (`output_tokens == max_tokens` is a cut, not a completion) surfacing in
  a fourth place, which strengthens the case for a shared guard rather than
  another local check: the platform detects truncation *after* persisting or
  acting on the fragment, over and over.
- `CLAUDE.md` § Debugging — the rule that found this in one query.
