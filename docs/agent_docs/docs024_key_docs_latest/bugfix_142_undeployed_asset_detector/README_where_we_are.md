# Where we are — the undeployed-asset detector (bugs_open/142)

Plain prose, append-only, newest at the bottom.

---

**2026-07-31, evening.**

I picked up bug 142 off the open pile. The short version of what it says: we
have a check whose job is to notice when we've generated an image for a site and
then never actually put it on the site. It has been getting the answer backwards
for the two "brand head" images — the little favicon in the browser tab, and the
social card that shows up when someone pastes your link into Slack or LinkedIn.

Two things were wrong with it, and they're both worth understanding because
they're the same kind of mistake in different places.

The first is that the check starts by asking the database "list me this site's
images". Then, for each one, it asks "is this one on the site yet?" That sounds
sensible until you notice what it can never find: a site that has **no** image at
all. There's no row to start from, so there's nothing to ask about. The check was
structurally incapable of reporting the exact problem it exists to report. I
think that's the transferable lesson here — if you want to detect that something
is *missing*, you can't start your search from a list of the things that exist.

The second is that it looked for the image in the wrong place. It searched the
HTML of the site's pages. But the favicon and the social card aren't referenced
from a page — they're in the site's shared `<head>`, which we store separately.
So the check would search every page, find nothing, and conclude the image had
never been deployed. Every time. For ever.

I measured what that actually costs us before changing anything. Run against
production today, the check as shipped would raise **96 findings across all 14
live sites** — and I then went and fetched the actual files over the internet.
Twelve of those sites serve both images perfectly well. So 24 of the 96 findings
were flatly wrong. Meanwhile idea.uk and webdesign.co.uk genuinely serve nothing
— you get a 404 — and the check has never once mentioned them in its entire
history. It was crying wolf about the healthy sites and silent about the two
broken ones.

Worth noting: idea.uk's pages *advertise* that social card. If you paste an
idea.uk link anywhere, the preview is fetching an image that isn't there.

**What I've changed.** The brand-head half now starts from the two images we
know every live site should have, and asks each site "do you have this one?" —
so absence is now something it can actually say. And it looks for the evidence
in the right place.

**One thing I got wrong along the way, and how it got caught.** I ran this past
our diagnosis loop before committing to the explanation. The loop came back
"unverifiable" — it couldn't reach the evidence it wanted, partly because I'd
written the symptom badly (I bundled two separate faults into one question, which
its own instructions tell you not to do). So it didn't confirm anything. But
buried in the data it pulled was one row I hadn't seen: a site whose image record
contained a literal unfilled template placeholder instead of a real filename.

That mattered, because my fix was about to treat "there's a record" as proof the
image was deployed. Two sites — gamesdesign and robot-hands — have records like
that, and both serve their images fine. So my check would have been right about
them by accident. Worse, when I tightened it to be strictly correct, it started
saying "this was never generated" about sites that were serving the file. That's
a *false statement*, in a fix whose entire purpose is stopping a check from making
false statements.

So it now recognises three states rather than two: deployed, missing, and "the
record is wrong so I can't tell". The third gets written down where a human can
see it and deliberately raises no work item. I'd rather it say "I don't know"
than say something untrue.

**Where it lands.** After the change: the 24 false alarms are gone, the two
genuinely broken sites get flagged for the first time, and the two odd-record
sites get an honest note instead of a wrong accusation.

**One caveat I want to be straight about.** These findings go into a queue that
currently has nothing draining it — that's bug 083, it's a known problem, and the
owner has a decision pending on it. So this fix makes the detector *correct*; it
doesn't make anything act on what it finds. I deliberately haven't tried to route
around that, because the obvious shortcut would auto-dispatch every finding the
platform raises without anyone looking at it first, and 083 explicitly warns
against exactly that.

It's with the review council now. Once that's back and the next chassis image
rolls, I'll verify it on the running pods and close the ticket.
