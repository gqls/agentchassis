# relojistas.com — milestone summary, 26 July 2026

*Plain language, written to be read aloud. Previous read-out:
`SUMMARY_2026-07-25_relojistas_rebuild.md`. Detailed technical handover for a fresh
session: `HANDOFF_2026-07-26_continue_here.md`.*

*Honest note on why this one exists: the site itself has not gained a capability since
yesterday's read-out. What has changed is that everything in it has now been
**re-verified live** rather than repeated, the fix that mattered most has a proper
window behind it instead of a single cycle, two of my own claims turned out to be
wrong and are corrected, and the question of who configures the server has become its
own piece of work.*

## What we're trying to do

Take a dead Spanish watch forum whose audience never fully left, and turn it into a
Spanish-language watch news portal that serves that audience automatically — the old
RSS subscribers at their original address, the searchers who still type watch queries
into it, and the crawlers that decide whether anyone else ever finds it.

## Where we've come from

The feed was reactivated and its return measured; the site gained reference content
behind a fabrication fence; the news became readable by machines as well as browsers;
a routing mistake of ours that had the homepage rewriting itself every six hours was
found and fixed; search capture returned and the engine was taught to answer. Then the
build list was finished by *deciding not to build* the last item on it — per-board
feeds — because the survey showed there was no real subscriber to serve.

## What we've done since the last read-out

**We proved the important fix, rather than merely believing it.** When the homepage-
rewriting fix went live it had exactly one refresh cycle behind it. It now has two days
and twenty completed jobs, with not one of the bad kind since the fix rolled. That is
the difference between "it worked once" and "it holds".

**We caught a false alarm before it reached you.** A quick count showed twelve jobs of
the very type that was supposed to be extinct, dated after the fix. Read as a total, it
looked like a regression. Read row by row, it was two different things: another session
deliberately repairing the glossary pages, and old jobs from before the fix. Their
repair is live and good — the glossary pages now carry their own headlines instead of
a borrowed generic one. The lesson recorded is that a total is not evidence about a
cause; the column that explained it was one query away.

**We corrected a second claim of mine.** Two safety rows I said had "vanished
unexplained" had done nothing of the kind — I had looked for them under the wrong
status. They are present, inert, and now *provably* inert rather than presumed so.

**And the server itself became a proper piece of work.** Walking the provisioning
script end to end turned up a real bug — one that cannot bite your pending session, but
would break the very thing the script's own instructions tell you to do next: add a
domain and re-run. It is fixed. More importantly it exposed the wider problem: three
machines, three ways of configuring them, two copies of one script that now agree on 61
lines and disagree on 614. That is now its own workstream with the design settled and
your ruling recorded — the island pulls its configuration outward rather than being
pushed to, so it keeps the isolation it was built for.

## Where we are now

The site runs itself and the build list is empty. Verified today, not assumed: the feed
rebuilt this afternoon unattended with thirty items; the homepage carries no invented
links, no fabricated address, its search box, and twelve news items written into the
page itself; the corpus is still growing; eighteen pages are deployed and healthy.

One item remains and it is yours: a single session on the server that switches on
search answers, starts counting real visitors instead of Cloudflare's machines, enables
the collector, and fixes the last three failing feed addresses. Those three are still
failing today — which is the correct and expected answer until you run it.

## Where we're going

After your session: confirm those addresses now answer, and re-check the one condition
that would revive per-board feeds now that real visitor addresses are visible. Then the
measurement the whole project exists for — how many real people came back — becomes
answerable for the first time.

Alongside it, the machine work proceeds on its own track, beginning with the step that
costs nothing: prove the framework can reproduce this box's live configuration exactly,
before it is ever allowed to change it.

## The one-sentence version

The portal is finished, self-running and re-verified rather than remembered; two of my
own claims about it turned out wrong and are corrected on the record; and one server
session of yours still stands between it and the only number that ever mattered.
