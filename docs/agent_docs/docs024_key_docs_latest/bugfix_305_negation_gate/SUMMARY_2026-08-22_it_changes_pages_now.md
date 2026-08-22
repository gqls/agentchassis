# SUMMARY 2026-08-22 — it changes pages now

Third in this lane's series (08-20 *the rule becomes a check* → 08-21 *it went live, and running it
found what the tests could not* → this). Written because the previous entry ended with a check that
was live, was rewriting well, and — as we discovered that evening — **was not actually changing any
pages**. That is fixed, and this morning a real page proved it. That is a different read-out, so it
gets its own file.

## What we're trying to do

Stop a habit of machine writing reaching our pages: saying what a thing *isn't* in order to say what
it is. The owner read three of our pages, quoted two sentences of exactly that shape, and asked for
two things — fix those pages, and make sure that sort of copy never leaves the framework again.

## Where we've come from

We built a mechanical check rather than another written rule, because two attempts to fix this with
words had already failed. It counts the mannerism everywhere, and on the writer that produces almost
all our page copy it sends the offending sentences back once and pastes the answers in. It went
through review four times, was refused once, and was approved on the fourth.

It went live on the 21st and immediately taught us three things no test had. First it was live but
blind — the repair step had no model attached, so it found everything and could rewrite nothing.
Then, once it could rewrite, several corrections to the same paragraph overwrote each other, so six
accepted rewrites arrived as one. Both were fixed the same day.

**The third was the serious one, and it was only visible by reading the stored page.** Pages were
finishing successfully and reporting "repaired, nothing left to fix" while the stored page was
byte-for-byte what it had been before. The correction was being made in memory and thrown away at the
boundary between two steps — an honest status over a page that never changed. That is precisely the
failure this design was supposed to be immune to.

## What we've done

**Both halves of the fix for that are now live, and this morning a page proved the whole thing end to
end.**

The code change (the step now hands its corrected copy onward as its own output, instead of editing
someone else's in place) shipped in this morning's release. I confirmed that against the running
software on both servers rather than the release notes, using a check that could have failed and
didn't. Then I switched the page builder over to read the corrected copy, and confirmed the change
took on the live configuration.

Then I waited for ordinary traffic, and it came. A page called *interest rate stress test* was
rebuilt at four minutes to ten. The check found eight of these constructions, rewrote six, and left
two alone because that site's own instructions had supplied them — which is the deliberate behaviour,
not a miss. The page saved two minutes later.

**And the stored page is the proof.** None of the six phrases it removed are in it; all six of its
replacements are. I checked both directions on purpose, because a test that can only come out one way
proves nothing. The same query run ninety minutes earlier, on a page built before the switch whose
save also succeeded, gave exactly the opposite answer. Same site, same morning, same instrument.

The rewrites are the sort of thing you would do with a pencil: *"…rather than paying down the loan
itself"* cut so the sentence ends at *"interest"*; *"based on your budget, rather than theirs"*
shortened to *"based on your budget"*. Nothing else on the page moved.

That same page also settled the earlier splice problem for good: all six corrections were in one
paragraph, and all six landed.

## Where we are now

**The check detects, chooses, rewrites, and changes pages. All four, proven on real copy rather than
in a test.** The remaining caution is narrow and worth keeping: if a page's save is refused for some
unrelated reason, the check's own status will still say "repaired" while the page has not changed —
so a status is never the last word, the stored page is.

Two things I got wrong today and corrected. An earlier attempt at that same page had its save refused
by a different safety guard, and I told the team who own that guard it would block every rebuild of
our older pages. The very next attempt disproved it — that was one bad roll of the dice, the guard did
its job, and I had the disconfirming evidence in my own output when I wrote the claim. I have withdrawn
the question and logged the mistake. Separately, a correction I made to our shared traps file was
committed by another session before I got to it; nothing was lost, but the trail now credits the wrong
thread, and my first check of "nothing was lost" was too weak to have detected a partial loss.

**What has not changed is the thing the owner originally complained about.** The three pages he read
still carry that copy, because it comes from that site's own written instructions and the check
deliberately does not overrule a site's own voice. Nine of our twenty-five sites are in the same
position. That is a decision about positioning for whoever owns those sites, and no amount of
machinery will make it for them.

## Where we're going

For this lane, the engineering is done. What remains sits with other people: the site brief that
mandates the tagline onto four page types, a question about whether *rather than* is a tic or ordinary
English (it appears in nearly half our sections, and the answer is taste, not measurement), and a
proposal to run the counting half across the whole fleet, which the reviewers were right to say does
not belong inside a bug fix.

The bug itself stays open, and correctly so: the defect is fixed, live and proven, but the damage the
owner pointed at is still on the pages.
