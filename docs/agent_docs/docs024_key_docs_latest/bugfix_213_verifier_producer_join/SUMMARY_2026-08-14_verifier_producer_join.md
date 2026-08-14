# SUMMARY 2026-08-14 — bugfix 213: the gate is proven in production, and the sweep is back on

*Current state only. The chronology is in `README_where_we_are.md` and `NOTES…md`; earlier
summaries (`SUMMARY_2026-08-11`, `SUMMARY_2026-08-11b`) are the record of what we believed
then and are not superseded — they are the trail.*

---

## What we are trying to do

Stop the platform from ticking work off as "done" when nothing was done.

A work item is a fault the system found on one of our sites, handed to an agent to fix. The
question this workstream exists to answer is the boring one nobody was asking: **when that
agent says it has finished, did anything actually change?** For most kinds of fault there is a
checker that re-examines the page afterwards. For some there is not, and for one kind we found
the checker was answering a different question from the one the fault asked — correctly, and
about something else. An item like that closes green, the fault stays on the site, and the
system's own record says it was handled.

## Where we have come from

The original bug was a mix-up of identity: two different producers were filing faults under the
same label, and the checker registered against that label only understood one of them. That was
fixed and shipped, and a second, daily check was built to catch the same mix-up happening again
anywhere else in the fleet.

What that fix exposed was worse than the mix-up. The other producer's faults were moved to a
label of their own — and that label has **no checker at all**, so its items were completing
entirely ungraded. Fourteen of them closed in a single day carrying no verification whatsoever.

Then, measuring rather than assuming, two things turned out to be true. First, the agent those
faults are sent to **cannot fix them** — not "usually fails", cannot: its repair vocabulary is
two words wide and every one of these faults asks for something else. We proved that by running
the agent's own repair function over the real page content from the live database: it changes
nothing, on any of them. Second, we caught one closing **falsely**: the agent reported in its
own record that it had changed nothing, the item was marked done anyway, nothing on that page
had moved, and the next audit found the identical fault the following day.

## What we have done

**Built a gate, and today watched it work on a real site.**

The gate is small and its rule is one sentence: *a fix that changed nothing is not a fix.* When
the agent's own report says it altered no page and no template, the item is refused instead of
being ticked off. It needs no web browser and no page fetch, which matters because the estate
has repeatedly refused to put either on this path.

It went through the review council twice. The first round sent it back, correctly: to add the
gate I had moved some existing code, and I had proved my new code harmless to the things it does
not touch while proving nothing about the code I had moved — which every other fault type in the
fleet depends on. That is now covered by tests, and the review approved it on the second pass.

**Today it was proven in production, on `mortgagecalculator.co.uk`.** One waiting item was
dispatched deliberately. The agent ran, reported zero changes on both of its repair steps, and
the gate refused the completion — recording, in the item's own error field, exactly why and the
measurement that licenses it. The item then cycled through its three permitted attempts and
came to rest as `failed`, which is the honest description of a fault nobody can currently
repair. Three dispatches, about three minutes, and no language-model cost at all. **Before this
gate, that item would have been marked `complete`.**

We also stopped short of the second half of the work, deliberately, and that turned out to be
the right call twice over. The plan had been to grade these faults at the next audit instead —
if the audit stops reporting a fault, close it. Measuring first showed the audit is reliable
enough for that (it re-reported a known fault on seven of seven return visits) but not reliable
enough to act on a single silence. And then a neighbouring team shipped their own version of
exactly that mechanism while we were working, including a note from their reviewers that a third
hand-built copy of the pattern should become shared code. Ours would be the third. So that piece
now has a settled shape and one clear precondition: talk to them first.

## Where we are now

**The gate is live, proven, and from today it is doing real work rather than standing by.**

Until this afternoon it had never once fired, and the reason mattered: the mechanism that
dispatches these faults had been switched off two days earlier, by another team, on cost
grounds. So nothing was completing, nothing false was being minted, and the gate had nothing to
catch. It is worth being blunt about that, because it would have been easy to claim credit for a
leak that had actually stopped for an unrelated reason.

At your instruction that mechanism is now **back on**. Its first run in two days behaved
exactly as designed — it examined a single site, filed thirteen new findings, and cost nothing
measurable in language-model spend. One of those thirteen is another fault of the kind the gate
guards, so it will meet the gate on a later run. We should expect to see a small population of
these items settling as `failed` from here; that is the gate working, not a new problem.

Two things remain open and both are yours to call. The bug **cannot close on its own recorded
terms** — the fix removed the very traffic that would have demonstrated it, so the criterion
written into the file is now unsatisfiable. And the second half of the work is designed but
unbuilt, waiting on a conversation with the neighbouring team rather than on any technical
blocker.

## Where we are going

Three things, in order of how much they are worth.

**First, the second half: grading these faults at the next audit.** The gate refuses a bad
"done"; it cannot confirm a real one. Closing that gap means retracting a fault when the audit
comes round and no longer sees it — site by site, and only after three consecutive silences,
because the arithmetic of what we measured does not support acting on one. It must be built on
the shared helper the neighbouring team's reviewers asked for, not beside it.

**Second, deciding how this bug closes.** Three options, all written up: accept the proof we
have and close it, recording the one branch never exercised in production; stage a synthetic
case to exercise that branch for real; or leave it open and accept that the file no longer
describes anything reproducible. The proof we gained today is real but it is about the *new*
gate, not about the original branch, so this question is genuinely still open.

**Third, the thing nobody has explained.** Ten of the fourteen items that closed ungraded carry
a record that does not belong to the agent that was supposed to have handled them — a design
specification for nine of them, and an unrelated decision about other pages for the tenth. We
have not guessed at why. The gate now records every case it cannot read, so the next few weeks
of that record will tell us, without anyone having to theorise.
