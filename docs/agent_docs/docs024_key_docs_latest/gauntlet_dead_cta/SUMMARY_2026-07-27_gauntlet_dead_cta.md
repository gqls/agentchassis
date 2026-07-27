# SUMMARY — 2026-07-27 · nothing on vonc.com is invented any more

## What we're trying to do

vonc.com is a site built around a daily argument. It shows a provocation, invites
you to take a position on it, and puts you against an AI opponent that argues
back and then judges you. The work of this stream has been to make that promise
literally true — and, just as importantly, to remove everything on the site that
was pretending to be true already. The site had been generated with the texture
of a busy place: counts of how many people had taken part, percentages who
disagreed, leaderboards, named users with opinions. None of it had ever happened.

## Where we've come from

The debate engine was built, reviewed and deployed to its own small server, and
proven with a real round-trip. The Gauntlet page was then rebuilt against it: the
clock now starts because a round really started, the objectives complete only when
the AI actually replies, and the invented statistics are gone. The archive page
was rebuilt at the same time, so each past provocation opens its real written case
at an address you can share. Both went live on the 26th and were verified on the
deployed pages rather than assumed — seventy-two of seventy-three checks passing,
the single failure being the engine's own intermittent fault.

That left one page untouched: the Arena. We had it on record as a page that never
finished loading.

## What we've done

The record was wrong, and the truth was worse. The Arena loaded perfectly well
and almost everything on it was fabricated: twenty-six invented users with
handles, posting opinions they never wrote, each carrying an invented tally of how
many people had voted it "Genius" or "Delusional"; a "remix chain" crediting
invented contributors; and a hardcoded list of five provocations rotating by day
of the year, which had silently drifted away from the real provocation the rest of
the site was showing. The box inviting you to file your take wrote it to your own
browser and nowhere else.

Given the choice between building a real backend for it and scoping it down to
what is honestly true, the owner chose to scope it down. So the invented people,
their votes and the remix chains are deleted. The provocation is now read from the
same file the homepage, the Gauntlet and the archive share, so it cannot drift
again. The take box is replaced by a real route into the Gauntlet, where the
opponent actually exists. In the space the fabrication occupied, the page now
lists the six real provocations that do exist, each linking to its own case.

It was driven ninety times in a real browser — desktop, phone, and once with the
data file deliberately broken to confirm it says so plainly instead of sitting on
"Loading" forever. Ninety of ninety pass, both before delivery and against the
live page afterwards.

Along the way we found a fault that belongs to the platform rather than to this
site: sixteen tool pages across six sites publish their own build instructions as
the description search engines display. The Arena's was the worst — it announced
to Google that the page had "no fetch calls, no backend". That one page is
corrected; the rest is filed as `bugs_open/103` with the root cause cited.

## Where we are now

**There is no fabricated content left anywhere on vonc.com.** The platform's own
claims scanner returns zero findings across all forty-nine components. Every
destination the Arena offers resolves. The site's four public surfaces — homepage,
Arena, Gauntlet, archive — now all read from one file, so they cannot disagree
with each other about what today's argument is.

We also confirmed something about the new chassis build that shipped this morning:
it changed nothing for this site. The acceptance harness still waits three-tenths
of a second for answers that take ten seconds, because nobody had yet written that
fix — a new build cannot ship a change that does not exist. And the engine's
intermittent failure lives on a separate server the cluster does not touch.

Two faults remain, both known and neither hidden. The engine still fails now and
then and still discards the reason, so nobody can say why. And the acceptance
harness would fail a correct page, so the formal acceptance run cannot yet be
trusted.

## Where we're going

Three things, in order. Make the engine's failures diagnosable — one line of
logging — and then fix them, since that is now the only thing between the Gauntlet
and a page that simply works every time. Fix the acceptance harness so it can wait
for an answer, rather than teaching the page to print reassuring text that would
also appear with the engine switched off. And decide what the Arena eventually
becomes: it is honest now, but it is a lobby, and the option of giving it a real
backend — real takes from real people — remains open and is now a clean build on
top of something true rather than a rescue of something invented.
