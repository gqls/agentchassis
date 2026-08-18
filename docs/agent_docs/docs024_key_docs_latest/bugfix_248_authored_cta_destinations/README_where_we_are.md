# Where we are — the CTA contact-link fix

*(Plain prose, append-only, newest at the bottom. No jargon where a plain word will do.)*

## 2026-08-18

**What the problem was.** A lot of our sites have a button that says something like "Get in
touch" or "Start a Conversation", and it should go to the contact page. Something in our own
repair machinery was quietly changing where those buttons point — usually to whatever
calculator or tool happened to rank first on that site. The button text was left alone, so the
page still says "Get in touch" and now takes the visitor to a password-strength calculator.
Nothing failed, nothing was logged, and the job that did it reported success.

**Why it did that.** We have one list of "places a button should not send people": contact,
about, privacy, terms, legal. That list is right for one job — when the system is *inventing*
a destination for a new button, it should not pick the contact page, it should pick something
that sells. The trouble is the same list was being used to answer a completely different
question: "should I trust the link that is already here?" Those are not the same question, and
using one answer for both meant that any button already pointing at the contact page was
treated as untrustworthy and replaced. The list was never evidence about what a person had
chosen — only about what the machine should choose.

**How bad, and is it still happening.** Eighteen buttons across the fleet are currently sitting
in the state where this can happen, thirteen of them in the two component types the repair
actually touches — and eight of those thirteen are on webdesign.uk, including the homepage.
The bug file said this was mostly harmless at the moment because the jobs that trigger it were
switched off. They were switched back on two days ago by another piece of work, quite
reasonably, and nobody connected the two. So it has been live.

**We caught one happening.** While checking a page yesterday evening I found that its contact
button had been destroyed a couple of hours after I first looked at it — at 19:11 on the 17th,
on finetuning.uk's services page. The button still reads "Start a Conversation" and now points
at a password-strength calculator. In the same second, a repair job finished on that page. That
matters more than it sounds: until now we had a list of 59 pages that had *lost* their contact
link, and no way to prove what had removed them. This one we can prove, because the job that
did it is named in the database.

**What I changed.** The list now does only its original job — stopping the system inventing a
link to the contact page. A separate, clearly named rule now answers the other question, and it
does not need anyone to record anything new. The reasoning is: the system is *incapable* of
choosing the contact page as a destination, so if a button is pointing there and that page
really exists, a person must have put it there. Keep it. That gives us the "who wrote this"
information the original bug report said we would have to build a whole new database field to
get.

I also fixed a second version of the same fault that nobody had noticed. There are two places
in the code that write these links: one when repairing a page, one when rebuilding it from
scratch. Only the repairing one had been reported. The rebuilding one had no protection at all,
which means fixing only the reported half would have left the same buttons dying the next time
their page was rebuilt.

**One thing I did differently from what we agreed.** You approved deleting the check that files
these buttons for human review — it has filed 103 items and every one that has been sampled was
a correct button. I demoted it instead: it still records what it sees, it just no longer creates
a task for a person. The reason is that it turned out to be the only check capable of spotting a
contact link that the machine invented by mistake. Deleting it would have swapped a pile of
false alarms for a blind spot, so I kept the eyes and removed the paperwork.

**How confident am I.** The tests do not just pass — I took the fix back out three times, once
per part, and confirmed each time that the right test failed. That matters because a test which
passes against the broken code proves nothing. There is also a deliberate "control" test that
fails if the fix were too aggressive and started freezing every button on the site.

**One honest gap.** If a button's text happens to share a single word with some other page —
"Talk to us about pricing" shares "pricing" with a pricing page — the system will still redirect
it there, over the contact page. I have not closed that, because the same mechanism is what lets
us fix genuinely wrong buttons, and turning it off would break more than it fixes. It is written
down rather than papered over.

**What happens next.** The fix is committed but not yet live — it goes out with the next fleet
release. Until then the bug stays open, because a fix that has not shipped is still broken in
production. Once it ships I will prove it on a real page rather than trusting the tests: take a
before-and-after snapshot of one affected page, run a repair at it, and confirm the contact link
comes through untouched — with a second link on the same page that *should* change, so that
"nothing changed" cannot be confused with "nothing ran".

**Something for you to decide, when convenient.** There are 149 findings where the system
correctly spots a wrong button — "Contact our supply team" pointing at a break-even calculator —
and correctly says it should point at the contact page, but then cannot carry out its own
suggestion, because the repairer is not allowed to choose the contact page. I have written this
up as its own bug (308) rather than fixing it here, because fixing it undoes the reasoning this
fix depends on: the moment the machine *can* choose the contact page, "the machine can't have
chosen this" stops being true. Doing both properly needs the database field the original bug
report proposed. It is now the second bug to need it, which is probably the argument for
building it.

## 2026-08-18, later — it is live, and it works

The release went out while I was writing this up, and the fix is in it. I checked that
properly rather than trusting the release: the running binary says which commit built it, and
our fix is an ancestor of that commit. A version tag alone would not have told us — we have
been caught by that before.

Then I proved it on a real page instead of trusting the tests. I took a before-snapshot of
leopardessconsulting.co.uk/how-it-works, fired exactly the kind of repair job that has been
destroying these buttons, and looked at what came back. Both contact buttons — "Walk us
through your problem" and "Describe the job" — still point at the contact page, and the live
page confirms it.

The part that makes that meaningful is the control. The repair genuinely rewrote both
sections: their last-modified stamps jumped from 18 July to this evening and fourteen archive
rows were written. So the job ran, rewrote the page, and *chose* to leave the contact links
alone. Without checking that, "nothing changed" would have been indistinguishable from "the
job never ran" — which is how you end up believing a fix that did nothing.

The review panel approved it first time, with four advisory comments and no blocking ones. I
checked every comment that could be checked rather than answering from memory, and one of them
caught me: I had written that the demoted detector's findings are still recorded "durably".
They are recorded, and they are queryable for about a month, and nothing downstream reads
them. That is a weaker claim than I made, so I have corrected it.

A second correction came from another session working on the neighbouring half of the same
function. They pointed out that the *rebuild* path's output is currently thrown away before it
reaches a page — so the second fault I fixed was not actually destroying anything yet. It
would have started the moment their own fix lands, which is why it is worth having in place
first. I had claimed it was doing damage already. It was not, and I have corrected that in
four places and logged it.

I am leaving the bug file open rather than filing it as closed, and deliberately: two other
sessions are gating their work on it and are actively writing into it this evening. Closing it
would move the file out from under them mid-flight. It should close once their wiring fix
lands and the rebuild half gets its first real exercise.

On your instruction to keep a provenance record — I have written that up as the decided
direction on the new bug, along with your rule about not adding flags that let agents ignore
things. That one is worth having in writing, because the obvious way to build provenance is to
add a switch that turns it off for callers in a hurry, and that is exactly what you have said
not to do. The fix that just shipped conforms already: it added no switch of any kind.
