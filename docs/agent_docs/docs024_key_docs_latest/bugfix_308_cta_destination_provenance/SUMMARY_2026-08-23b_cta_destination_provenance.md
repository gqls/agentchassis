# SUMMARY — 2026-08-23b — CTA destination provenance (`bugs_open/308`)

*(The earlier summary today, `SUMMARY_2026-08-23_…`, said Phase B was blocked until 1 September and
should not be started. That was wrong within two hours of being written; this one supersedes it and
explains why. Both are kept — the pair is the record.)*

## What we are trying to do

Fix a bug where the platform spots a broken button, correctly works out where it should point,
files a repair — and the repair cannot possibly perform it. A button reading "Contact our supply
team" links to a break-even calculator. The check sees it, names the contact page as the fix, the
repair runs, reports success, and changes nothing. Next pass it finds it again.

## Where we have come from

The blocker was never the repair. It was that the platform decided whether it could rewrite a
stored link by *reasoning*: "we could never have produced a link to the contact page, so if one is
there, a person put it there." Sound today, and exactly what had to stop being true. The owner
ruled on 18 August: **record the provenance properly, then widen.** Phase A — the record — shipped
on 22 August and was proven working on live data the next morning.

## What we have done

**Phase B is written, measured, mutation-tested, committed and submitted for review.** The
detector and the two repair writers now share one answer to "which pages may this button's words
name?", which is the thing bug 308 is actually about: they had been answering it from different
lists, so the check could name a page the repairer's list did not contain.

**But the measurement came first, and it changed the design twice.** Before writing any of it we
replayed the platform's own matching code over a frozen snapshot of the live fleet — 829 pages,
1,266 buttons with words on them — with a control to prove the harness agreed with the shipping
code (1,266 comparisons, no disagreements).

- **The change is much bigger than the bug.** It takes fleet-wide button rewrites from about 32 to
  **428**, and roughly two thirds of that is nothing to do with bug 308.
- **A third of those rewrites were decided by alphabetical order** — 263 of 1,146 matches. Not a
  tail: on finetuning.uk it would have moved a button reading "how we work" *off* the "how we work"
  page, because the About page's title happens to contain that phrase and "about" wins the
  alphabet. Thirteen live findings say to do exactly that.

Two attempts to break the tie more cleverly were measured over the whole fleet and both rejected:
each fixed some cases and broke others, for the same reason a third such rule was thrown away back
on 11 August. **So the platform now says "I don't know" and leaves the button alone.** The repairs
bug 308 is actually about all survive that.

And one counter-intuitive result worth carrying: the *cautious* version of this change — add only
contact and about pages to the candidate list — rewrites a third as much and gets **more** of it
wrong. A short list does not make the matcher careful; it makes it certain for the wrong reason.

## Where we are now

Committed and inert until the next fleet build. Seven deliberate sabotages of the change were each
caught by a test. The one piece that changes what a shared component *promises* is written up as a
formal architecture note (RFC 047) with a live question for the owner in it. Two of the bug file's
own standing suggestions are now answered by measurement rather than opinion, both in the negative.

**Bug 308 stays open, for three reasons that are worth stating separately.** Nothing has touched a
served page yet — the bar is a button moving on a site you can load in a browser. **41 of its 188
findings can never be closed by this change at all**, because they are links inside prose rather
than button fields. And one class of wrong answer survives: a confident false match, where the copy
genuinely shares a distinctive word with the wrong page's title.

The estate-wide AI outage that stopped this lane yesterday turned out not to be an outage: the
billing account being read was not the account the fleet's key belongs to. It was resolved this
morning, the stalled review round restarted, and it returned a verdict within the hour.

## Where we are going

Two things, in order. **Verify at the artefact after the next fleet roll** — induce a discovery run
and a repair on finetuning.uk, which carries 55 of the 188 findings, and load the page; watch both
directions, because a fall in findings alone could just mean the detector stopped running. Then
**close the loop the bug file actually asks for**: the repairer still re-derives the answer the
detector already computed and wrote down, rather than consuming it, and there is still no
completion check that refuses to mark a repair done when the button did not move. That is what
turns "complete and unchanged" from a silent outcome into a refusal.
