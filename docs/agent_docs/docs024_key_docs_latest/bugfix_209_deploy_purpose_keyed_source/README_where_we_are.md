# Where we are — bug 209, the image deploy that picks its source by "purpose"

Plain-prose log, append-only, newest at the bottom.

---

## 2026-08-08, late evening — first look, and it turned out better and worse than expected

Picked this up because the previous session's handoff said 209 was the only bug in
the open pile that nobody owned. Checked that myself first, since the ownership
script reads commit history and can't see a session that's mid-fix. Three other
sessions were live at the time; none of them was working on this. One was my own
predecessor, which is why the script pointed at me.

The bug is about how we deploy an image. When the system deploys, say, a hero
image, it has to find where that image is stored. One of the ways it looks is by
asking "what's the URL for the *hero* image in this run?" — and that's the
problem: "hero" is a category, not a name. If a run produced two hero images, the
second one quietly overwrites the first, and a deploy asking for "the hero" gets
whichever came last. You'd deploy the wrong picture and nothing would complain.

**The good news: I can't find any way this actually happens today.** I went
through every live workflow that can deploy an image — there are three, plus one
more that was the prime suspect. The two workflows that do handle several images
in one go handle a *hero* and a *logo*, which are different categories, so they
never tread on each other. The suspect workflow does have two steps that both
claim the "hero" slot, but they sit on opposite sides of a fork in the road —
only one can run in any given job. And it doesn't deploy anything itself; it hands
off to a separate deployer, passing the exact image it just saved, by name rather
than by category. I also checked what that deployer actually has in memory when it
runs: the category-keyed values simply aren't there.

So the fault is real in the code but currently unreachable. Latent, not live.
That's a downgrade in urgency, not a clean bill of health — the door is still
open for the next workflow anyone writes that makes two of the same kind of image.

**The bad news, and it's the useful part: the fix the bug file recommends would
cause the very bug it's trying to prevent.** The file's first suggestion is to
delete the category lookup and rely on the image's ID instead. But the two older
workflows don't tell the deploy step which fields to use, so the system falls back
to rummaging through everything in memory looking for anything called "asset_id" —
and the order it rummages in is deliberately randomised by Go. I ran the real code
400 times on identical input: **the step deploying the logo picked up the hero's ID
344 times out of 400.** So the category lookup, which looks like legacy cruft, is
actually the thing keeping those two workflows correct — because for them, the
category *is* the difference between the two images.

Given that, I've deliberately not written a fix tonight. Changing how a shared
piece of machinery resolves things, to fix something that can't currently happen,
on the evening I just discovered the obvious fix is the harmful one, is how you
turn a latent bug into a live one. What I've done instead is lock the behaviour
down with tests that record *why* the lookup was kept, so the next person who
reads the bug file's ranking and reaches for the obvious fix hits the 86% figure
before they ship it.

**One thing I got wrong and want on the record.** I first wrote that those two
older workflows "haven't run in 26 days", because the run table has rows going
back to mid-July. Then I checked what those old rows actually are — and they're
all cancelled or stuck jobs. Completed runs get cleared out after about a day. So
the honest statement is "they didn't run today", which is a much weaker claim, and
I'd have been quoting a made-up month if I hadn't looked. I also tried a second,
longer-lived log to settle it properly, and threw the answer away: it reported
zero runs for a service I'd just watched run sixteen times, so it can't see these
agents at all. A measurement that can't come out the other way isn't evidence.

What I'd want a decision on, when there's appetite: whether those two older
workflows are actually dead. Nothing schedules them any more as far as I can see.
If they're genuinely retired we can delete them, and then the clean fix for this
bug becomes available with no awkward legacy constraint. While they're merely
dormant, we have to keep supporting them.

## 2026-08-09, morning — checked everything again after the new build, still holds

A fresh build went out this morning. Nothing of ours is in it — we shipped tests
and documents, not product code — but the deployment also re-applies the workflow
configurations, and our whole "this can't currently happen" conclusion rests on
those. So I re-read them, comparing content rather than trusting timestamps
(the timestamps all changed, because re-applying identical config still stamps
it). Every fact the conclusion depends on is unchanged, character for character.
The tests still pass. Nobody touched the relevant code overnight. The conclusion
stands on the new build.

One sentence from yesterday deserved tightening, and I've tightened it in all the
records. I said removing the category lookup "swaps a correct lookup for a
mostly-wrong one" — true, but it's the *backup* route that gets swapped, not the
main one. The older workflows normally find their image through a direct pointer
that the proposed fix wouldn't touch; the category lookup is what catches them
when that pointer is missing. So the danger is conditional: the bad fix only
misfires on the day something else has already gone slightly wrong. That doesn't
change the advice — a safety net that grabs the wrong thing exactly when you fall
is worse than useless — but the record should say precisely where the danger
lives, and now it does.

## 2026-08-09, mid-morning — owner's answer received; the problem restated plainly before any fix

**The owner's ruling:** pageflow-builder and site-work-orchestrator are not dead,
but not being worked on. If we need to diverge from them, we can use new actions
and workflows as necessary. Explain the problem clearly before moving further.

So here is the problem, stated to be read aloud.

**What the deploy step has to do.** When a workflow says "deploy this image", the
deploy step must first find where the image's bytes actually live in storage.
It has three ways of finding out, tried in order: first, an address handed to it
directly by the workflow; second, a set of lookups in the run's shared scratchpad,
all keyed by the image's *category* — hero, logo, and so on; third, the image's
own database record, looked up by its unique identity.

**The defect.** The scratchpad has one slot per category, and the last writer
wins. If one run ever saves two images of the same category and then deploys them
through the scratchpad route, both deploys get the second image's bytes. The
first image's address is simply gone — overwritten — and nothing errors, because
a perfectly valid address came back. It is just the wrong image's address.

**Why nothing is broken today.** Measured twice now (and re-checked after this
morning's deployment): no live workflow saves two same-category images and then
deploys through the scratchpad in one run. The two older workflows handle one
hero and one logo each — different categories, different slots, no collision.
The modern path hands the address over directly, by identity, so the scratchpad
is never consulted. The fault is a loaded gun in a drawer nobody currently opens.

**Why the obvious fix is the dangerous one.** The natural instinct — and the bug
file's original recommendation — is to delete the category route and rely on
identity. But identity is only trustworthy when a workflow *names its inputs*,
and the two older workflows don't. For them, "find the identity" means rummaging
through the whole scratchpad for anything called asset_id, in an order the
language deliberately randomises — and with two images saved, that rummage picks
the wrong one 86 times in 100 (measured on the real code, four hundred runs). So
for precisely those two workflows, category is the correct key and identity is
the broken one. Deleting the category route wouldn't hurt them on a normal day —
their direct pointer still works — but it would replace their correct safety net
with one that grabs the wrong image exactly when the direct pointer fails.

**What your ruling means for the fix.** Keeping the older pair alive but
untouched fits the right fix perfectly, because the right fix is divergence, not
surgery. The shape I propose, for approval before anything is built:

1. **An opt-in switch on the deploy step, default off.** Off means today's
   behaviour, bit for bit — the older workflows never set it and never notice.
   On means: resolve by explicit address or identity only, never by category, and
   fail loudly rather than guess. The modern workflows — which already pass
   identity — switch it on. This follows the estate's own 2026-08-02 ruling that
   new authority on shared machinery ships as an opt-in field with the unsafe
   default off. The characterisation tests we shipped yesterday exist precisely
   to prove the off position stays identical.
2. **Alternatively, an entirely new action**, leaving the old one untouched. Zero
   risk to the frozen pair, but it duplicates the download-optimise-commit logic
   and becomes a second thing to maintain. I'd only choose this if you prefer the
   old code literally unedited.
3. Either way, the writers additionally record each image's address under its
   *identity* in the scratchpad, so an in-run identity route exists at all. That
   part touches shared machinery and goes through the register and the council,
   with the other consumers told.

No code has been written. Awaiting your go-ahead on shape 1 versus shape 2.

**Separately, the question you asked about bug 064.** Nobody is working on it —
and strictly, 064 itself is closed (July 24th). What's failing today is a
*recurrence* of its shape, the second one: on the evening of August 8th the
**idea_uk_vm_site** thread shipped its "decision records" change (RFC_015), which
taught the database a new document type — "decision" — but not the Go code's list
of valid types, and a test that exists precisely to catch that split has been red
ever since. Two other threads (this one, and the chrome-divergence one) have both
tripped over the red test and correctly recorded it as not theirs. The
idea_uk_vm_site thread's own notes don't mention it, so they likely don't know.
The fix is one word in one list, plus the closed bug's own checklist — it belongs
to that thread, and they should be told.
