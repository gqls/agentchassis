# NOTES — architecture seat / council memory

Running technical record. **Append-only, newest at the bottom.** Evidence,
commands, what the system actually said, and every misstep — the missteps are
the point, not an appendix.

> **Started late (2026-07-27).** This workstream ran for three days on
> `DECISIONS`, `SUMMARY`, `RUNBOOK` and `HANDOFF` without a NOTES or a
> `README_where_we_are`, which is two of the standing five missing. That is
> itself a misstep worth recording: the wrong turns from 07-25 and 07-26 survive
> only as the corrections embedded in the SUMMARY series and in `WRONG_CALLS.md`,
> not as a log. This file starts from the 07-27 late session; earlier missteps
> are cited where known rather than reconstructed.

---

## 2026-07-27 (late) — closing the feature-designer gap

**Done: `/tmp/acm/APPLY_gap.sh`.** The one thing the handoff owed. It gives
`feature-designer`'s guardian its minutes + the deflection check, its
bug_historian the case index, and replaces `review_architecture`'s prompt so the
routing signal lands in the first line of `notes`.

Pre-flight checks, all of which I would want run again before any config push:

```
live updated_at = 2026-07-27 13:44:56 on all three council agents (unchanged
since the cutover) — so nothing had been written by another session in the
1h15m since CONTEXT.json was generated at 13:57.
```

Structural diff of live → `CONTEXT.json`, canonically loaded, was **exactly
three differences**, all `prompt_template` strings:

| path | chars |
|---|---|
| `review_architecture` | 6627 → 7614 |
| `review_bug_historian` | 3773 → 29200 |
| `review_guardian` | 2456 → 4390 |

and `live == SEATED.json` was `True`, which is what makes `ROLLBACK=1`
trustworthy — the rollback target was verified to be the actual current state,
not an assumed one.

Post-apply invariants, all held: `live == CONTEXT.json` byte-exact after
canonical load; step set unchanged; `review_fields` still 6; `hard_veto_from`
still `["guardian"]`; `max_rounds` still 3.

**MISSTEP (caught before acting, cost nothing, would have cost a wasted council
round): I nearly fired a probe run at another thread's live ticket.**

The architecture seat had 0 reviews and I wanted to exercise it. The only two
`capability_gap` specs carrying both `owner_approval` and `code_pointers` are
`9ed684bc` (tools-api — visibly owned by the Gauntlet workstream) and
`7b89fb35` (the colour-fixer remit gap). `7b89fb35` looked safe on two signals
that were both misleading:

- `status = 'deferred'` — which I read as idle. It is not; it is where this lane
  parks an item between council rounds.
- `grep -rln hardcoded_section_colors docs/ bugs_open/ features_open/` returned
  only generic guides and unrelated `idea_uk` running notes — **no owning
  workstream doc at all.**

Both said unowned. Then I opened the spec body, and its `capability` field
contains `=== REVISION REQUIRED (round 2)`, `=== ROUND 3`, and
`=== ROUND 4 — ONE CHANGE ONLY, owner-directed`, with three prior council
correlations (`c91bb061` → `1a9feed2` → `b604f92d`, all APPROVED, 07-26 21:45 →
07-27 11:21). An active four-round design iteration with owner-directed
instructions already written for its next run.

**The check that caught it:** opening the row I was about to act on. Not a
grep, not a status column. Same shape as the 2026-07-19 refutation that
corrected the diagnosis section of `CLAUDE.md` — the failure mode was not
missing information, it was not looking.

**Transferable, and not currently covered by anything:** `scripts/who-owns.py`
resolves ownership for a **bug number or slug**. There is no equivalent for a
`site_work_items` row, and for work items the ownership evidence often lives
*inside the spec jsonb*, where no repo grep can reach it. A work item's `status`
is not an ownership signal and a docs grep is not a coverage check.

**Consequence for the measurement, and a correction to the handoff.** The
handoff's §6 item 2 said to let councils run and re-read the report, "it cannot
be hurried". That understated it: `review_architecture` exists **only** on
`feature-designer`, `feature-designer` refuses anything without an
owner-approved spec (`check_spec_approved`), and there are 5 `capability_gap`
items in total, 2 approved, **both owned by other threads**. So waiting produces
nothing on its own — the seat's first review has to arrive on the colour
thread's round 4, or from a newly owner-approved spec. Not a defect; a rate
limit worth stating so the next session does not read 0 as breakage.

**Reachability, checked rather than assumed.** Before waiting on the seat I
verified it can actually speak. BFS over the workflow graph including
conditional branches (`then_step`/`else_step`, not just `next_step`):

```
reachable from load_spec: 24 of 24 steps
review_architecture REACHABLE: True
orphans: none
chain: ... review_guidelines -> review_architecture -> review_guardian -> council_decide
```

It has **no footprint/relevance gate** — unlike the 16-seat gate's seats it is
an unconditional step, so it fires on every design run that clears the approval
gate. Worth noting the first walk I ran looked alarming (`reached: False`)
because it followed `next_step` only and stopped dead at the first `conditional`
step; a linear walk over a branching graph proves nothing.

**First post-cutover evidence — and it is good.** Cutover is 13:44:56. The
council at **14:18:19** (`b64141e5`, the 109 render-context fix on the fix lane)
is the first past it, and all three payloads landed:

- **debug_historian cited `WRONG_CALLS.md` by date** to reject an absence
  claim: *"This is exactly the shape of WRONG_CALLS 2026-07-21 ('no existing
  loop-controller action' — an absence claim shipped without the search) and
  2026-07-24"*, against the plan's *"That is the complete set — no sixth path
  exists"*. That is the case index doing the one thing the workstream was built
  for: our own logged mistakes, applied to a new submission.
- **bug_historian cited `016b §9`, `bugs_open/034`, `bugs_open/109`** and named
  the symptom-vs-mechanism pattern across rounds.
- **The guardian invoked the stability preference and reasoned its way OUT of
  deflecting**: *"The recurrence across three rounds is evidence of a genuinely
  scattered defect (multiple independent RenderContext producers), not evidence
  that this fix belongs at a higher layer."*

**A measurement-fidelity problem this exposes, stated against my own metric.**
The adoption report scored that guardian review `invoked_stability=1,
cited_precedent=0`, which reads as a miss. Qualitatively it is the D5 payload
working: the seat engaged with recurrence explicitly instead of reflexively
sending the fix upward. The report counts a *precedent citation*, and reasoning
correctly about recurrence without quoting a past report does not match. So
**`cited_precedent` undercounts correct behaviour**, and "6 of 90 → n" is not by
itself the verdict on D5. n=1; no conclusion either way yet. Recording it now,
before the numbers grow, so the metric is not later read as cleaner than it is.

**An unplanned finding that bears on D7(b) and on D1/D2/D3.** On that same
council, `bug_historian` opened its note with *"Architecture-level concern for a
human"* and recommended a human decide between a shared render-context-builder
refactor and continuing to fix drop points one live test at a time. That is an
architecture judgement, raised by a seat not commissioned to make one, on the
**fix lane** — which has no `review_architecture` seat. D1/D2/D3 deliberately
placed the seat on `feature-designer` only. This is the first live evidence that
the fix lane also generates forward-fitness questions and currently routes them
to "a human" because nothing there owns them. Not enough to reopen the decision;
enough that it should be on the table when D7(b) is answered. `[UNMEASURED]`
how often this happens — one instance, noticed by reading, not counted.
