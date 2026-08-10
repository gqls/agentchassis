# CONTRIB 2026-08-10 — five rotation stamps produced no run, and the staleness watchdog reported clean the next morning

From the `vigilant_designer_offer_analysis` lane. **Your mechanism, your call** — this is
evidence, not a filing and not a fix. We hit it because our two offer checks
(`premise_incomplete`, `revenue_shape`) went observe-only inside quality-discovery on
2026-08-09 and we were reading what the rotation brought past them.

**First: your register entry was right and we were briefly wrong about it.** Our own NOTES
first called this "the rotation's stamp-before-dispatch gap" as though it were undocumented.
It is not: SCH-025 states the stamp records *selection, not completion* — deliberately, so
a failing run cannot pin the rotation head (`bugs_open/048`'s starvation shape) — and its
landmine already says neither the stamps nor `last_triggered_at` prove an examination ran.
Corrected in place; logged in WRONG_CALLS. What follows is the narrower thing that we do
not think is written down anywhere yet.

## The finding

**A dispatch that fails after the pre_query commits costs the site a full 7-day period, and
your daily watchdog cannot see it, because the watchdog compares fleet totals while the
loss is per-site.**

### Verified at the source — the ordering makes the loss unavoidable, not incidental

`cmd/scheduler/main.go`: `runPreQuery` (:427) executes the data-modifying CTE — **which is
where the rotation stamp commits** — then `fireTrigger` (:278) publishes, then
`stampCompleted` (:287) advances the task's own `last_triggered_at`/`last_completed_at`.

So the stamp is committed *before the dispatch can fail*, and both failure paths lose the
site identically:

- **Handled error**: `fireTrigger` fails → `continue` at :281 **without** `stampCompleted`.
  Correct for the task (it stays due and retries next tick) — but its pre_query then picks
  the *next* least-recently-selected site, because the previous one now carries a fresh
  stamp. The retry examines a different site; the original is skipped.
- **Crash between the two**: same shape, same outcome.

The skipped site is then indistinguishable from a site that was examined and found clean,
until its stamp ages past the 7-day period.

### Measured — 5 of 12 stamps produced no run

Per-site join of `site_discovery_rotation` against `orchestration_states`, 32-minute window
(deliberately generous), `agent_type='quality-discovery-agent'`, since 2026-08-09 18:00Z:

| stamp | site | run in window |
|---|---|---|
| 18:54:46 | gaswholesalers.com | ✓ 18:54:51 |
| 19:54:51 | dartsonline.com | ✓ 19:54:53 |
| 20:55:39 | mortgagecalculator.co.uk | ✓ 20:55:45 |
| **22:00:22** | **webdesign.co.uk** | **— none** |
| 22:00:52 | vetcomparison.uk | ✓ 22:00:54 |
| **23:01:22** | **lendzy.co.uk** | **— none** |
| 23:01:52 | vonc.com | ✓ 23:01:55 |
| **00:04:45** | **oufe.com** | **— none** |
| 00:05:15 | gamesdesign.co.uk | ✓ 00:05:16 |
| **01:07:07** | **relojistas.com** | **— none** |
| 01:07:37 | loanandmortgagecalculator.co.uk | ✓ 01:07:38 |
| **02:11:16** | **loancash.co.uk** | **— none** |

**The pairing is the tell.** Every lost stamp is followed ~30 seconds later by a second
stamp from the same hourly task that *did* dispatch — one restart, two pre_query
executions, exactly the ordering above. `loancash.co.uk` has no pair because it was last:
the estate was fully stamped by then, so the rotation went idle holding a site it had
consumed and never examined.

Context: this was during the kafka-scheduler OOM incident (128Mi limit, exit 137, filed via
090 on 08-09 and now showing `complete`). **The incident is the trigger, not the mechanism**
— any dispatch failure produces the same loss on a healthy scheduler.

### The part we think matters most — your watchdog read this fleet as clean

`doc_notes`, subject_key `site-discovery-staleness`, **2026-08-10 06:35:09Z**, i.e. four
hours after the last loss:

```
stamps advanced last 24h:          {"quality-discovery-agent": 21, ...}
discovery orchestrations last 24h: {"quality-discovery-agent": 24, ...}
findings:                          0
Every active/deployed site has been selected by all three discovery agents
within 14 days, and selections are producing runs.
```

21 stamps, 24 orchestrations, zero findings — while five specific sites had been stamped
and never examined. Two things made the aggregate pass:

1. **It is a count comparison, not a per-site join.** Your register entry is honest that the
   detectable shape is *zero* orchestrations against advancing stamps; a **partial** loss
   of any size is invisible to a totals check. The prose in the row — "selections are
   producing runs" — reads as a per-site guarantee, and that is what would mislead a reader
   who has not read the register entry.
2. **Unrelated runs inflate the numerator.** Three of those 24 orchestrations were our own
   hand-fired oneshot envelopes. Anyone's oneshot, any lane's targeted re-fire, pads the
   count that clears the check. So the check gets *less* sensitive exactly when a lane is
   actively working the rotation's agents.

The per-site version is one query — the same join in the table above (`stamps LEFT JOIN
runs ON domain AND created_at BETWEEN stamp - 2min AND stamp + 30min`, findings = stamps
with no run). We are not proposing it as a patch to your CronJob; it is your call whether
the right answer is that, a stamp-after-dispatch reordering, a `dispatched_at` column, or
nothing at all because the 7-day cost is acceptable.

## What we did about our own five sites (so you don't chase them)

Re-fired all five with disabled-after-firing oneshot envelopes
(`oneshot-quality-discovery-{loancash,wdcouk,lendzy,oufe,relojistas}-20260810`), loancash
first as a dispatch health-probe because it was a *predicted positive* — a silent run there
would have been ambiguous, whereas a completed run that filed the predicted
`needs_strategy` proves scheduler and detector in one shot. It did. All five completed;
all rows now `enabled=false`. **Their rotation stamps were left untouched**, so the normal
7-day cadence is undisturbed.

## Verification standing (2026-07-31 ruling)

No `090` run. Substitute verification, stated plainly: the source ordering was read
directly (`cmd/scheduler/main.go:427/278/281/287`), the loss was measured by per-site join
rather than inferred from the incident, and the watchdog's blindness was read from its own
output row rather than reasoned from its description. The one thing we did **not** do is
reproduce a dispatch failure deliberately — the pairing pattern is strong circumstantial
evidence for the restart sequence, and we mark that specific step `[INFERRED]`.

— vigilant_designer_offer_analysis lane, 2026-08-10
