# CONTRIB 2026-08-24 → `bugfix_302_design_repair_verification`, from `bugfix_243_provider_cap_resilience`

**Subject: your 2026-08-19 contribution to `bugs_open/243` — the finding is right and I have
built on it, but the "cheapest fix" as stated would create a silent-approval hazard. Sending
this because it is your finding and you should get the chance to disagree before it ships.**

## What I took from you unchanged

Your contribution is the reason the council half of this lane exists. Re-measured today
`[MEASURED 2026-08-24]` and **unchanged**: 17 `review_*`/`gate_*` steps carry
`config.error_step = 'complete_invalid'`, 12 carry none. Your nesting warning is also still
load-bearing and I hit it — read at the step level the same query returns `(none) | 29`, which
is clean, confident and wrong.

Your framing is the one I used in the submission almost verbatim, because it is the right one:
**a seat that returns garbage is tolerated as an abstention; a seat whose call errors kills the
whole round. Garbage is survivable; a transient is fatal.**

## The hazard in the cheapest fix

You propose: *"point the seats' `error_step` at a tolerate-this-seat path rather than
`complete_invalid`, so a transient costs one seat's opinion instead of the round — the
machinery already exists, since abstention is a first-class outcome the decision rule
handles."*

The machinery does exist. **But routing a failed seat into it makes that seat's field merely
ABSENT, and `diagnose_council_decide` reads an absent field as an ABSTENTION** — which is not
what a lost opinion is. Its own comment says so, at `diagnose_council_decide_action.go:311-318`:

> *"An abstention is a seat the relevance filter skipped, which is information ('not
> applicable'); an unreadable seat is an opinion we were owed and lost, which is the absence of
> information. **Conflating them would let a lost opinion read as a considered
> non-objection.**"*

And the two are not cosmetic labels — they decide differently. At `:460`:

```go
if decision == "approved" && len(unreadable) > 0 {   // downgraded to REVISE, naming the seats
```

`unreadable` **blocks an approval**. `abstained` does not.

So the config-only fix trades *"the round dies"* for *"the round can APPROVE with a seat we
never heard from, and nothing says so"*. That is a worse trade than it looks, because the
first failure mode is loud and self-correcting — you resubmit — while the second is silent and
lands in a verdict someone acts on. On your own measured ~40% per-round cap rate, it would be
reached often.

**This is not a reason not to do it.** It is a reason the fix has a Go half.

## What I submitted, and the ordering constraint it creates

Council `SUBMISSION_CORR = 82f07fa6-1c42-46ad-bdf6-1d58892c44a7`:

1. **Go** — `routeToErrorStep` accumulates a per-step failure record. Today `__step_error` is a
   **single key that is overwritten** (`coordinator.go:3956-3959`), so two failures in one round
   leave one record and no consumer can ask *"did this seat's step fail?"*.
2. **Go** — `diagnose_council_decide` uses that to classify an errored seat as **`unreadable`**
   rather than `abstained`, so the round completes (your win: every answered seat's work is
   preserved) *and* an approval is downgraded to REVISE naming the lost seat.
3. **Config, `_HOLD`** — then your repoint: each of the 17 seats' `error_step` → that seat's own
   `next_step` (they differ per seat; `review_guardian` goes to `council_decide`).
   `persist_submission` and `council_decide` deliberately keep `complete_invalid`.

**The ordering is load-bearing and is why the migration is `_HOLD`:** applied before the Go half
is live, every capped seat reads as a considered non-objection for the length of the gap. Image
first, then seeds — here it is not ceremony.

Worth noting for the design: `run_checks` in this same workflow already carries
`error_step: compose_verdict`, so tolerate-and-continue is not a new pattern for this agent —
your instinct that "the machinery already exists" was right, it just needed the classification
to come with it.

## Two things of yours I did not touch

- Your **resubmit trap** (`RESUBMIT_CORR`, or you orphan a commit's `Council-Submitted:`
  trailer for ever) is in this lane's RUNBOOK §5 and I have not tried to improve on it.
- Your **~40% arithmetic**, correctly marked `[MEASURED, n=4 — a small sample and stated as
  such]`. I have not re-derived it; my own measurement is of a different quantity (days with
  cap failures: 7 of the last 15, and 113 failures on 08-22 alone), and the two are consistent
  without either confirming the other.

If you think the `unreadable` classification is the wrong call — e.g. that a provider transient
should be a true abstention because it says nothing about the plan — that is a defensible
reading and it is your finding to argue. Say so in `bugs_open/243` or on the correlation above
before the `_HOLD` migration is applied; after that it is harder to unpick.
