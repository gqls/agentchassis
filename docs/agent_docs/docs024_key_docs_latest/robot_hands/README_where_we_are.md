# robot-hands.com — where we are

Plain-prose running log, newest at the bottom. Append only.

---

**2026-07-19 (session: robot-hands.com site fixes R1–R6)**

You looked at the site on the 17th and found six things wrong. All six are now
fixed, and five of them I've checked on the live site rather than just in the
database. The dark look is back, the learning centre has one address instead of
three, the dead "Load More" button is gone, and the article list now shows only
the three guides that actually exist instead of nine, six of which were dead
links.

The interesting part is that the handoff I started from had the cause wrong, in
a way worth knowing about. It said a regeneration run on the 16th had wrecked
the header. Actually the blue header had been there since the 9th — I found
that by walking back through the deployed page history — and the run on the
16th only spread it to every page. Two other things in that handoff were also
wrong: it told me to restore the header from saved snapshots, and there are no
snapshots; and it blamed a colour-fixing agent that, it turns out, has never
worked at all. Every time it runs it fails, and the platform records it as
successful. I've written that up as a bug.

What was really wrong was that when the site was switched to the dark design
back on the 10th, only one of four places colour is stored actually changed.
Three stayed on the old light palette, and the header and footer were still
being built from old components that write colours directly into the page. So
every time anything regenerated, the blue came back.

Worse, a broken health check was telling the platform every few minutes that
this site had no design at all — which made the design agent invent a brand new
colour scheme, from scratch, over and over. On the 17th alone the stylesheet
was rewritten four times. One of those rewrites put a white background on a
site whose entire design is dark, and that went live before I caught it.

I've fixed that health check in the platform code, not just on this site. It
was looking for a marker that nothing in the entire system has ever written, so
it was firing on every properly-designed site, forever. That fix is now live —
another session's image roll picked it up last night — and I've confirmed it
worked: the false alarm hasn't fired once anywhere since. I also rebuilt this
site's header and footer so they carry no colours at all, only references to
the stylesheet, which means a future regeneration can't reintroduce this.

I put the platform change through the reviewer council, and I'm glad I did. It
took three rounds and finished with seven of the eight reviewers approving. It
caught a genuine mistake in my fix — I was checking whether a value existed
rather than whether it had anything in it, which would have created the exact
opposite problem — and it pushed back, correctly, on my trying to bundle a
bigger change in alongside a small one. I've split that bigger piece out as its
own job with the groundwork done.

**What still needs a decision from you.** Two tool pages — MatchMatrix and the
robot payload budget calculator — were planned but never actually built, so
five buttons on the homepage lead to "page not found". That's now the most
visible thing wrong with the site. The experience-loop work is already tackling
exactly this problem across all the sites, so my recommendation is that the
next session joins that rather than building these two by hand. If you'd rather
have it look right in the meantime, the quick fix is to point those five
buttons at the MatchMatrix page that does exist.

One caution I want to flag rather than bury. The project guidance changed
yesterday: claims about how the platform works are now meant to go through the
automated diagnosis loop before being asserted, because a session with full
context recently filed a confident claim that was refuted in ten minutes. Two
of my write-ups were done before that change. The evidence in them is cited and
checkable, but if either becomes the basis for someone else's work, it should
go through the loop first.
