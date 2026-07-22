# SUMMARY — the approval-rate fix is live, and the council approved itself

**2026-07-22 (second summary of the day — a distinct milestone: the fix shipped,
was confirmed working in the wild, and passed the gate under its own new rule).**
New file per the standing-five rule; it does not overwrite
`SUMMARY_2026-07-22_council_approval_rate.md` (which stopped at "fix chosen").
Written to be read aloud.

## What we're trying to do

Get every platform bugfix reviewed by the council, so an approval means something
and — eventually — the council can block a bad change (PR-mode). The measure was a
coverage report counting commits that carry a "reviewed by council" stamp, which is
earned by an *approved* verdict.

## Where we've come from

The gate was live and being used, but it almost never *approved* — about 4.5% of
submissions over a week, everything else "revise". We diagnosed that: the code that
tallies the reviewers' votes sent a change back on any single objection and never
looked at how serious the objection was, so a trivial nit blocked a change exactly
as hard as a real flaw. Two-thirds of the "revise" verdicts had no serious
objection in them at all. You chose the fix — only a serious objection (or a formal
veto) blocks; minor ones are recorded and returned to the author but don't block.

## What we've done since

1. **Built, tested and committed the fix**, then it shipped in the new chassis
   build.
2. **Brought the guidelines reviewer into the council for this class of change** —
   widened the rule that decides which reviewers wake up so that changes to the
   council's own decision code always draw the guidelines seat. While doing that we
   found and fixed a real bug in the tool that keeps our two councils identical: it
   was silently failing to copy exactly this kind of change while reporting "in
   sync".
3. **Confirmed the fix is genuinely live** — by reading the running program, not
   trusting the version number.
4. **Caught it working on someone else's change**: an organic review came back
   "approved with 4 advisory objections, none serious" — a change that four
   reviewers had minor notes on now passes instead of being sent back.
5. **Dogfooded the fix through the gate itself**, under the new rule.

## Where we are now — the fix approved itself

The dogfood came back **APPROVED** — the first change approved through this gate.
Nine reviewers ran (seven correctly sat it out as irrelevant), the guidelines seat
was among them and approved, and — the telling part — **the guardian, our
always-on veto seat, raised a medium and a low objection and said in as many words
"the plan is right, so I'm not vetoing."** Under yesterday's rule that objection
forces a "revise"; under the new rule it's advisory, so the change passes. That is
the exact behaviour we set out to create, demonstrated on a real submission by the
strictest seat on the panel.

Both halves of the new rule are now witnessed live: minor objections no longer
block (two approvals now, one organic and one the dogfood), and the safety floor —
a serious objection, a veto, or a truncated review still blocks — is covered by the
tests baked into the running binary.

The reviewers' notes were fair and useful. The sharpest, from the guardian: there's
no explicit test for one corner of the floor (a review that was cut off mid-stream
*and* carried no objections at all). It behaves correctly, but it deserves a test —
we're adding it, which is exactly the point of harvesting a review's objections
even when it approves.

## Where we're going

- **Harvest the objections**: add the missing floor test; note that low/medium
  objections are already returned to the author in full (the persisted review
  record), not merely counted — a fair point one reviewer raised from the sketch
  alone.
- **Watch the approval rate** over the coming days. If it climbs as the diagnosis
  predicted (67% of the old "revise" rounds had no serious objection), the coverage
  measure becomes meaningful again.
- **Reopen the PR-mode conversation** — it was unbuildable when approval was
  unreachable; now that a well-made change can actually pass, blocking a bad one
  becomes a real option rather than a wall in front of every bugfix.
- **One honest limit to carry forward**: the "reviewed by council" commit stamp is
  designed for review *before* committing. We reviewed a change that was already
  committed and live, so it cannot retroactively carry the stamp — the approval is
  recorded against the submission, and the change is council-approved, but the
  coverage report's automatic join won't see it. That's a property of dogfooding a
  live change, not a gap in the fix.
