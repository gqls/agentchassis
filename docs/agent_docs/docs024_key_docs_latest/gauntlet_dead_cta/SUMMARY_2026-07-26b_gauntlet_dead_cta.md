# SUMMARY — 2026-07-26b · the Gauntlet is live and genuinely works

## What we're trying to do

vonc.com has a page called the Gauntlet. Until today it was a convincing-looking
shell: a countdown you started yourself, three objectives you ticked off by
clicking them, a progress bar that filled because you had clicked, and a set of
statistics that were invented. Nothing was sent anywhere and nothing answered
back. The aim of this workstream has been to make the promise on that page true
— you take a position on a daily provocation, something argues back, and you get
judged — and to do it without any of the fabrication that made the original look
busier and more popular than it was.

## Where we've come from

The backend was built, reviewed, merged and deployed to a standalone VM, and
proven with a real round-trip over the public internet. A five-seat review
council then approved a concrete build plan for the front end — the first full
approval this workstream has ever had. That left exactly one thing outstanding:
the page itself, which had never been touched.

## What we've done

The data file came first, because everything else depended on it. It still
carried invented numbers from June — "1,284 positions filed", "62% disagree", a
countdown frozen at "3h 12m" — which the backend was serving through to the page
verbatim. We settled a rule and applied it everywhere: no participation metric
exists in this system, so none appears. Where a number now shows, it is true by
construction — the clock is twenty minutes, there are three objectives, there is
one verdict. The archive entries gained real written cases and their own web
addresses, and the Arena copy was rewritten because it had been advertising six
live "rooms" that have never existed.

Then the page. The Gauntlet now makes three real calls to the debate engine.
Pressing the button starts an actual round and an actual clock. Typing a position
gets a written counter-argument and a pointed question back from an AI opponent,
usually in about ten seconds. Answering that gets a judge's verdict with its
reasons. The three objectives can no longer be clicked at all — each one
completes only as a consequence of a real reply arriving, and the third is
awarded only if the verdict lands with time still on the clock. When the engine
is unavailable, the page says so plainly and starts nothing.

The archive page changed too: click a provocation and the full case opens in
place at an address you can share, and the one entry with no case written yet is
visibly not clickable rather than being a button that goes nowhere.

Before any of it was delivered we drove both components in a real browser against
the real backend — sixty-five checks on the Gauntlet, thirty-one on the archive,
desktop and phone, no failures. That caught two defects that a check for "does
the element exist" would have sailed past: a detail panel that read as full while
it was empty, and a dead link left on a hidden template row.

Delivery was interrupted by the production database crash-looping for about
twenty minutes. That turned out to be a real and separate fault — the live
database has quietly drifted from its own configuration file and has no
guaranteed share of its machine, so when a neighbouring process took all eight
cores, a one-second health check timed out and Kubernetes killed a perfectly
healthy database, repeatedly. We did not touch it; it is written up with the
evidence and the fix as `bugs_open/082` for the owner to decide on.

## Where we are now

It is live, and it has been verified live rather than assumed. Seventy-two of
seventy-three checks pass against the deployed pages on both desktop and mobile,
including the full journey through to a real verdict, on both profiles. The
homepage buttons lead where they say. The archive deep links work, including on a
cold load of a shared URL. The lobby cards now land on individual provocations
instead of all pointing back at the same index page.

Two things are honestly imperfect and neither is hidden. First, the debate engine
fails intermittently — a request comes back "unavailable" after about twenty-five
seconds and the same request usually succeeds on a retry. The page handles it
correctly, which is why the run still passed, but a visitor will sometimes have
to press the button twice. The cause cannot be determined from outside because
the backend throws the underlying error away without logging it, so the first fix
is to make it say what went wrong. Second, the single failing check is "no console
errors", and it fails *because* of those upstream failures — the browser logs
every one. That is the check doing its job.

## Where we're going

Three things, in order. Make the engine's failures diagnosable and then fix them,
since that is now the only thing standing between this and a page that just works
every time. Fix the acceptance harness, which waits three-tenths of a second for
answers that take ten to eighteen seconds and would therefore fail a correct
page — it must be fixed properly rather than by teaching the page to print
reassuring text that would also appear with the engine switched off. And finish
the Arena page, which we deliberately left alone: its source is now pulled and
read, its mount points identified, and the one visual bug in it located
precisely, so the next round starts from a known position rather than an
investigation.
