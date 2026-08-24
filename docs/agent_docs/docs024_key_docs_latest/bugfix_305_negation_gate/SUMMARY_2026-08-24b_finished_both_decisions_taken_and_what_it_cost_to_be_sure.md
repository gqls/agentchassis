# SUMMARY 2026-08-24b — finished: both decisions taken, and what it cost to be sure

*(Fifth and final in the series: `2026-08-20`, `08-21`, `08-22`, `08-24`, this one. Each is a new
file and none replaces another — read in order they show how the understanding moved, including the
places it was wrong. This one supersedes nothing; it closes the account.)*

## What we were trying to do

Someone looked at three live pages on one of our sites and said the writing "looks like it didn't go
through the framework". The specific fault was a habit of **saying what something is not, in order to
say what it is** — "shows you what's possible, not what survives production" — repeated until it read
as a verbal tic rather than a sentence. Two things were asked for: stop the platform producing it, and
fix the pages that already had it.

## Where we've come from

The first useful discovery was that this was not three bad pages. The construction was written into
the site's own **brief** — the instructions the page-writer reads before it writes anything — and the
same shape was in 24 of the 25 briefs across the estate. So no amount of editing pages would have
fixed it; they would simply have been written that way again.

What we built instead was a **gate in the writing pipeline**. As each section is written it is
scanned, the offending sentences are sent back to be rewritten, the rewrite is checked for whether it
is genuinely better and does not smuggle in a new claim, and it is put back before the page is
rendered. It took four rounds of review to be accepted, including one outright rejection that was
right and changed the design.

Then it went live, and running it in production found four things that no test had: it was detecting
but never repairing, because no model was configured. The repairs were being made and thrown away,
because the renderer read the wrong field. Several rewrites to one paragraph were overwriting each
other. And the log that recorded what it had done was quietly losing entries. Every one of those was
found by looking at the actual result rather than at the status that said "success".

## What we've done since the last read-out

**Two more defects, which turned out to be one and a half.** The repair log was dropping records when
the model ignored a sentence we'd asked about. Separately, the repair had a **size limit far too
small** — a page with one phrase to fix was fine, but a page with nine or ten needed the model to quote
each original *and* its replacement, and it ran off the end. When that happened the entire answer was
discarded: not "some fixes land", none did.

Measuring the two fixes separately is the part worth keeping. Before either, 40% of repair runs didn't
add up. After raising the size limit **alone** — with the accounting code untouched — 15%. After both,
zero. So most of what I had confidently described as "the model ignored some sentences" was really "the
model ran out of room". I had diagnosed the symptom correctly and pinned nearly all of it on the wrong
cause. Both fixes were needed; but if I'd found the size limit first, the other bug would have looked
far smaller than I said it was.

**Then both of the decisions that were waiting on you.**

*The tagline.* It now says "in days" rather than "in days, not months". The work was larger than it
sounds because the sentence was written into the brief in **five places**, and correcting fewer would
have looked finished and changed nothing — the gate checks the whole brief before deciding what to
leave alone, so one surviving copy keeps the old wording protected. It also wasn't where any of our own
notes said it was, and it is not in the fields literally named "tagline", which hold something else.
The satisfying part is what follows: correcting the instruction **removed the protection** from the old
wording automatically, so the pages carrying it became ordinary things for the gate to repair on their
next rebuild. Nobody edits a page by hand — which was the entire point of building the machinery.

*"A little bit of a tic."* Your ruling on "rather than" was neither of the two answers the question had
been framed with, and finding its home meant re-reading the code and correcting myself in front of you.
I had described the gate as having a limited number of repairs to hand out. It hasn't. It has a limited
amount of **forgiveness** — it lets a page keep two of these constructions and repairs the rest. And
until today, *which* two it kept was decided by nothing but where they happened to sit on the page. A
page could keep both of its worst lines because they were near the top, while the gate spent its effort
rewriting two mild ones further down. Once that was clear, your ruling had an obvious implementation:
the mild one is what gets forgiven, and the sharp ones always get fixed.

## Where we are now

**Finished.** The bug is closed. Every fix is live and demonstrated on real traffic rather than
inferred from a deployment — including the last one, which proved itself within twenty minutes: across
38 pages, not once did a sharp construction take forgiveness it was no longer entitled to. All eight
reviews came back approved.

On the three pages that started it: the one carrying both quoted sentences is completely clean, the
second has nothing left to repair, and the third is waiting on a rebuild in another team's queue.

One caveat worth repeating because it will look like a failure and isn't: **the tagline change will not
be visible on the site yet.** That site's rebuilds are stuck behind about thirty unrelated items
belonging to another team. The brief is correct; the pages haven't caught up.

## Where we're going

Nothing on this. What remains is one page in someone else's queue, and two smaller questions about
alert thresholds that are blocked on a different piece of work entirely.

The thing I'd carry forward is not a fix, it's a habit. **Four separate claims of mine were wrong this
week, and every one was caught by measuring rather than by anyone arguing with me**: the cause of the
log defect, the shape of the budget, whether a fault was newly reachable or had been live a fortnight,
and one case where I repeated another team's unmeasured assumption in four documents until it looked
corroborated. None of them changed what got built, because each was caught before it hardened. But all
four would have entered the record as fact, and two would have been quoted back at us later as
established. They are written down where the claims were made, which is the only reason that isn't
what happened.
