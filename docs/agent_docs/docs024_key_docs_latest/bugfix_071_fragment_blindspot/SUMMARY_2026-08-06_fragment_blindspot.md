# SUMMARY — 2026-08-06 — the fragment blind spot is closed at the detector

Written to be read aloud. Current state only; the chronology is in NOTES and
README_where_we_are.

## What we're trying to do

Make the platform notice when a link promises to jump to a section of a page and
that section does not exist. These are the links that end in `#something`. A
visitor clicks "see the pricing table", the page loads, and nothing moves — the
link is not broken in the way a 404 is broken, it just quietly does nothing.

## Where we've come from

This was the last piece of a bug filed on 25 July, which was originally about
something bigger and worse: the platform detected broken links on every build,
named them, and shipped the page anyway. Most of that got fixed through July. What
nobody picked up was the section-link half, and the reason it survived is
interesting — every one of our checkers skips it *by name*. The shared piece of
code that decides "what kind of link is this?" files anything starting with `#`
into a category that the pre-publish gate and the post-publish audit both
explicitly ignore, and the function that compares a link against real pages throws
the `#section` part away before it compares anything. So the blind spot was not an
oversight in one checker; it was a hole in the one definition all of them share.

When the bug was filed, 24 of the estate's 25 section links pointed at sections
that did not exist.

## What we've done

Re-measured first, and the picture had changed: today's 67 section links all work.
That happened through hand repairs and by giving the writing agent a list of real
pages, not through any check — and the same bug file records the writer inventing
dead section links on three consecutive rebuilds of one site. So the damage had
healed and nothing was guarding it.

We taught the existing link checker a fourth thing to look for, rather than
building a new checker. That choice is the important one: a new checker needs
somebody to switch it on in a configuration file, and we have a bug on the shelf
right now whose fix is correct, deployed, and has never once executed because that
step was never taken. Riding an already-running check means ours started working
the moment the new build rolled.

We also reused rather than rewrote. Another check already knows how to answer "does
this page contain an element with this name?", and it learned that the hard way —
its first version accused a working tool. We extracted that knowledge into a shared
piece both checks now use, so they cannot drift apart and we do not re-buy the
lesson. And we told the writing agent to stop inventing section links at all, since
nobody ever gives it a list of the real ones.

Before shipping we ran the new code over every page of every site: no complaints,
on all 67 links. Then we planted two broken links into a copy of the same data and
it caught both — because a "nothing wrong here" from something never shown capable
of finding a problem is worth nothing.

The review council approved it first time. One reviewer asked a question we had not
asked ourselves: the shared code we rearranged — who else uses it? It turned out a
third place does, the gate that refuses to publish a broken tool. We proved we had
changed nothing by running the old and new versions side by side over 4,036 real
documents.

## Where we are now

Live, on the build that rolled this afternoon, and proven on the real thing rather
than in a test. We built a page with four section links — two broken, two working —
ran the actual checker against it, and got exactly the two complaints we predicted
and silence on the other two. Then we repaired one of the two broken ones and ran
it again: two complaints became one, and the right one survived. That is what shows
the deployed code is genuinely looking at the page rather than guessing from the
link text. The test page has been deleted and we confirmed the site it sat on is
back exactly as we found it.

One small thing is not yet exercised: the piece that double-checks a fix before
marking the problem solved. Its database queries are verified, but the code around
them only runs when a real one of these problems gets fixed, and reaching it
deliberately would have meant starting a page build we did not want. It is written
down as owed.

## Where we're going

Three things are deliberately not done, and each is a decision rather than a
leftover.

The pre-publish gate still cannot check section links, because at that moment it
cannot see the site's header and footer and would raise false alarms about links
that are perfectly fine. Nothing yet repairs a dead section link, and the obvious
repair — deleting the link — leaves the words stranded in the middle of the page,
which we have been bitten by before. And most significantly: **no part of a page
currently publishes a stable name for a link to jump to.** Until that changes,
these links can only be avoided, never made to work on purpose. That last one
touches every page on every site, so it should be a deliberate decision, not
something slipped in with a bug fix — two of the reviewers made the same point,
observing that we now have three separate places reasoning about where a link
goes, and that the split itself may be the real thing to fix.
