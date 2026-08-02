# SUMMARY — 2026-08-02 — the third copy of a page's links

Written to be read aloud.

---

## What we're trying to do

Stop our sites publishing links that go nowhere, and — the part that had been
missed — stop us *storing* them, so that the same broken link does not come back
every time a page is rebuilt.

## Where we've come from

A bug was filed on 26 July after oufe.com went live with six dead links on its
home page and a top-menu item that 404'd. Over the following fortnight several
threads did solid work on it, and by the end of July the *publishing* side was
covered: whenever a page is built, saved or re-rendered, we now walk the finished
markup, check every link against the list of pages that really exist, and either
correct it or turn it back into ordinary text. Four separate points in the
pipeline do this. What ships is clean.

The bug stayed open, and its own final note said why, in a sentence worth keeping:
*"the repair now deletes a phantom link silently, and the authoring defect that
wrote it goes unreported."* Fixing the finished page is proofreading the printout.
The manuscript was untouched.

## What we've done

We found that a page keeps **three** copies of its links, not two — the deployed
file, the stored markup, and the structured data each section is rebuilt from. The
first two had a resolver. The third had nothing, so a bad link sat in the source,
was regenerated on every rebuild, quietly corrected again on the way out, and was
never once reported to anybody.

We also found why nothing had caught it, and it was not carelessness. Every check
we had asked its question of the wrong *level*: they held a list of field names,
or read the top level of a component's schema. Sections that hold several links —
a grid of cards, each with its own "read more" — keep those one level down, inside
the card, where a top-level check cannot reach at all. Twenty-five of our
component types are built that way. Adding two of them to a list would have fixed
two of them.

So we built the check to ask a different pair of questions. *Might this field hold
a web address?* is answered by the field's **name**, at any depth. *Is this
actually a link to one of our own pages?* is answered by the **value**, using the
classifier the rest of the platform already shares. That split is what removes the
need for a list of exceptions — a picture's address ends in `.jpg` and an outside
link starts with `https://`, and the existing code already knows both.

Then we measured it properly: not with a database query approximating the rule,
but by running the real, shipping function over all 885 stored sections on every
live site. **Fifty-two stored links point at a page that does not exist**, across
seven sites and four kinds of section — and not one of those four was covered by
either of the checks we already had. Nineteen of the fifty-two are near misses
where the real page is right there under a slightly different address; those are
now corrected in the stored data. Thirty-three point at pages that were never
built; those are recorded and left alone deliberately, because emptying them would
throw away what the writer intended in the one copy where it cannot be recovered,
and there is an unresolved disagreement between two of our own reviewers about
whether we should be doing that even on the finished page.

The reassuring number is the other one: 872 of the 885 sections came through
completely untouched, and nothing that was not a genuine internal link was ever
picked up.

## Where we are now

The change is committed, registered, documented, and with the review council. It
is **not yet running** — it takes effect at the next rebuild and rollout of the
software, and until then the bug stays open, which is the right rule: until it is
actually live the problem is still reproducible.

There is no clean-up operation to schedule. The fifty-two sort themselves out as
each page is next saved in the ordinary course of work.

Two things went wrong on the way and both are written down. I had built the same
safety guarantee twice without noticing, which meant either copy could be deleted
with every test still passing — so neither was ever really tested. And two of my
deliberate break-it-to-check-it attempts failed to compile, which prints the word
"FAIL" and looks exactly like proof when it is nothing of the kind.

## Where we're going

Three things, in order. When the software next rolls out, confirm it is really
running — by looking inside the running binary on every replica, not at the
version number — and then re-measure to check the count of fifty-two is *falling*
rather than sitting still, because a static count means the new pass is not being
reached.

After that, two extensions that were deliberately left out rather than forgotten:
a handful of writers that save a single section on their own route, bypassing the
place this check sits; and a decision, which is genuinely a judgement call and not
a coding one, about whether a link to a page that was never built should be
escalated to a person rather than merely recorded. That last question is the same
one already open about the published pages, and the two should be settled
together rather than separately.
