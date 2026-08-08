# SUMMARY 2026-08-08b — the first owner review, and the paragraph that was never ours

Second summary today, and it earns its place: `SUMMARY_2026-08-08` reported the
rewrite landing. This one reports the **owner reading it**, which is a different
event and changed what we think the problem is.

---

## What we're trying to do

Take a site that was hand-built outside the framework, bring it fully inside, and have
the framework write its words — so that everything the platform checks for gets applied
to them, and so the site can be improved by the pipeline rather than by hand. Then have
the copy actually be good: readable, written for the person reading it, offering what we
can genuinely do without either overselling or constantly narrating our own limits.

## Where we've come from

The site was adopted from a hand-built original at the end of July. Since then it has
been decomposed into editable components, its eleven calculators locked and their
arithmetic frozen against a golden baseline, an orphan page retired, four calculator
defects fixed, and twenty-odd pages brought up to date with platform improvements they
had been missing.

Two platform bugs had to be fixed before the framework could write anything here at all.
The build path could not resolve this site's section names, so a rebuild silently
declined to happen — that was `204`. Fixing it exposed `189` underneath, where a rebuild
would duplicate a locked calculator on the page. Both were fixed, reviewed and proven by
another lane on the 6th.

With those out of the way the voice rollout ran on the 7th and 8th: **23 of 26 pages
rewritten through the framework**, calculators untouched and still exact, every page
still serving.

## What we've done

Three pages did not rebuild — the homepage among them. That is `bugs_open/219`, and it
is a genuinely tidy piece of diagnosis: the content validator scans the whole assembled
page including HTML comments, and three of the twelve calculator templates carry a
developer changelog comment containing a string the validator treats as meta-commentary.
Exactly those three failed, all three of them, and the nine without the string passed.
Nothing is wrong with the pages; they serve fine, in their original voice.

Then the owner read the homepage and gave the first real verdict on the work: he likes
it, except the opening paragraph, which is too strong. It calls the site *"mathematically
rigorous"*, promises *"exact"* repayments and the *"true cost of credit"*. His point was
not that any of that is false. It is that it positions us as the authority on accuracy,
and accuracy is not what a borrower arrives wanting reassurance about. They already
assume our calculator works, and everyone else's.

Two things came out of chasing that down, and both were surprises.

**The paragraph he was reviewing is not ours.** It is original hand-built copy, one of
the three pages `219` blocked. So his praise for the rest of the page and his complaint
about the opening are both about writing the framework never touched.

**And a voice fix would not have prevented it.** I put the block back through the live
writer twice — once with only the platform default, once with this site's own voice spec
— and both removed "mathematically rigorous" and "exact" without being asked. The plain
default was the softer of the two. So the over-claiming is not a voice defect. What no
voice can judge is whether the sentence is worth having at all, and that is a positioning
question, not a writing one. It has gone to the offer and benefit analysis thread as its
first worked example.

## Where we are now

The site is healthy: 26 of 26 pages serving, all twelve locked calculator rows
byte-identical, the golden arithmetic still reproducing exactly. Twenty-three pages speak
in the new voice; three, including the homepage, still speak in the old one and are
waiting on a platform fix that is filed and understood.

What has changed is our model of the problem. We had been treating copy quality as a
voice-tuning exercise. The clearest evidence yet that it is not sits in this week's own
results: the default voice already strips the language the owner objected to, and still
nobody would call the result good. The judgement he applied — *this is true, and it is
the wrong thing to lead with* — is not available to any rule we could write, because it
depends on knowing what the reader wanted, and nothing in the platform represents that.

## Where we're going

Three things, in order.

Ship `219` and rebuild the last three pages, which is about three minutes each with the
tooling this lane already has, then re-baseline the golden set. That closes the rollout.

Rewrite the homepage opening. Two candidates are drafted and with the owner. Both lead
with what the reader is trying to work out rather than with how good our tools are, and
let the specifics carry the authority instead of claiming it. This is small and it is
also the first piece of copy on this site written to the new understanding, so it is
worth getting right rather than fast.

And the larger one, which is no longer this lane's: give the platform some way to reason
about what a reader wants, not only about what we are permitted to say. Everything we own
today — banned claims, the claims floor, the evidence register — filters what is
*allowed*. Nothing scores what is *wanted*. A page can pass every gate we have and still
open by answering a question nobody asked. That is what the homepage does, and it is why
the finding went to the offer and benefit thread rather than staying here.
