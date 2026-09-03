# SUMMARY — 2026-09-03 — the audit now sees the whole site

## What we were trying to do

Our sites are checked by a robot that opens each page in a real browser and looks at it the way a
visitor would — unreadable text, images that failed to load, content spilling off the side. It is
the only check we have that sees a page rather than reasoning about it, and it is the only thing
that can confirm a repair actually worked.

It had a limit on how many pages it would do in one go, because each page is a real browser opening
a real page and they are slow. The task was to stop that limit costing us coverage.

## Where we came from

The limit was not the problem. The problem was that the robot took the **same** pages every time.
It sorted the site and did the first sixty, then stopped — and the next run did the same first
sixty. Pages past that line were not "checked less often"; they had never been checked once, and on
that behaviour never would be.

On our largest site that was ninety-one pages, including **every one** of the forty-five guide
pages. Those sat in a block that no realistic limit could have reached — you would have needed to
raise it by more than half again, on a site that grew twenty pages in ten days. So raising the
limit was not a fix, it was a postponement, and we had already postponed once.

We also knew about this. A previous piece of work had made the robot announce every time it ran out
of room. Nobody was reading the announcements.

## What we've done

Two things, which is what had been asked for.

**The robot now remembers where it stopped** and carries on from there next time, so the limit costs
time-to-full-coverage rather than coverage. There was a trap in that, and it is worth knowing
because it nearly cost us something you had already ruled on: if the robot simply worked through the
site in order, a page with a *known* problem would wait a full lap to be re-checked — and you had
specifically instructed, in August, that waiting a week to confirm a repair was too slow. So pages
with an open problem now ride along in **every** run, and the rotation fills the seats that are
left. On our largest site that costs three seats out of sixty.

**And the announcements now have a reader** — a watchdog that runs every morning and can tell the
difference between "the robot is working through the site" and "the robot has gone back to
repeating itself".

## Where we are now

Finished, and checked rather than assumed.

Over three runs on its own schedule, with nobody watching, the robot covered **151 pages out of the
151 live on the site — nothing missed.** I checked that the strict way: not "did it visit 151
things" but "is there a live page it never visited". There is not. A second site did the same
unaided.

The watchdog started by itself yesterday morning at 07:50 and filed its report: nothing wrong.

Three things went wrong along the way and all three were caught before they could do damage. The
review council found a real fault in my first attempt that would have quietly blamed one page's
problems on a different page. Running the watchdog against live data before shipping it revealed
that it would have raised a false alarm every single morning. And a small correction to the robot
only showed up when it ran on its own schedule rather than when I ran it by hand.

## Where we're going

Nothing is left to build. Three questions are left for you, all in the handoff and none urgent:
whether the design-critique tool should also rotate or keep looking at the same pages (I would leave
it — it is a taste instrument, not a coverage one); whether we accept that a *new* problem on our
largest site's main pages may now take a lap to be spotted, in exchange for ninety-one pages going
from never to once a lap (I would accept it); and whether the watchdog's fortnight of patience
before it ignores a quiet site is the right number (I would leave it, but you should know it is a
judgement and not a measurement).
