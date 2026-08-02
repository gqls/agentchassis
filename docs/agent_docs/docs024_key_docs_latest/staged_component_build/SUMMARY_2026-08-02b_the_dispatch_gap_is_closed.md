# SUMMARY — the S6 dispatch gap for components is closed, proven live in the cluster, 2026-08-02

## What we're trying to do

Give a *component* — a piece of a page shared across many sites — the same machinery a
*tool* already has, including the last and hardest piece: an automatic test that drives the
real page in a real browser, in the cluster, the way `tool-acceptance-agent` already does for
tools. The wider goal is `features_open/027`'s deliberately small three-gate build ladder.

## Where we've come from

Earlier the same day, `teaser-reveal-panel` got a real, persisted, mutation-proven criteria
fence in the database — proof that a component's *contract* can exist and be trusted. What
remained open was narrower but load-bearing: nothing could get that fence in front of a real
browser in the cluster, because the action that dispatches a browser check resolves its
target page by name (`pages.name`), which only works when one function maps to exactly one
page. A tool is built that way. `teaser-reveal-panel` is placed on five pages across two
sites, because that is the whole point of a shared component, so the existing dispatch path
could not even ask "which page" for it. Two ways to close that gap were written down as a
real, undecided tradeoff and handed off deliberately unpicked.

## What we've done

Picked up the handoff, read the standing docs it pointed at, and made the decision it left
open: added a new, separate dispatch action for components rather than teaching the existing
tool action a second way to find a page. The reasoning, written into the plan before any code
was touched: the existing action is what every one of the fleet's tool tests already depends
on, so a new action next to it costs nothing to that working mechanism, where a branch inside
it would have to be trusted not to. Everywhere the two actions do the same work — building
the request and sending it to the browser service — the new one reuses the old one's code
rather than copying it, so they cannot quietly drift apart later.

Built it, checked it compiles and every existing test still passes, and confirmed by a
line-by-line diff that the existing tool path behaves exactly as before. Registered the new
mechanism in the project's own index of what exists, submitted the change for the platform's
advisory review, and committed it — catching and openly recording one small mistake along
the way, where a placeholder got written into a commit message instead of a real tracking
number.

Then the two halves that can only be proven live, not on a laptop: an owner-triggered rebuild
of the platform, checked rather than trusted (the first report of one turned out not to have
reached this particular deployment yet; a later, real one did, confirmed by finding the new
code's own fingerprints inside the running program on both machines). And the dispatch
itself — sent a real request into the cluster asking the panel's actual page to be checked
for real. All fifteen checks that could run passed cleanly; the rest correctly sat out because
this was a desktop-only pass. Alongside it, in the same request, deliberately pointed the
same machinery at the wrong page on the same site — a real page, just not one with this panel
on it — and it correctly refused, with exactly the message written for that situation, which
is what makes the refusal mean something rather than being a lucky accident.

## Where we are now

The gap is closed, and closed with evidence rather than an unexercised code path. A
component's fence can now be dispatched through the same browser-testing service a tool's
fence already goes through, it produces the same honest pass/fail/skip result shape, and the
one genuinely new piece of logic — deciding which of a component's several pages to test —
has been shown, live, to fail safely rather than silently testing the wrong thing. What this
does *not* newly prove is that each individual check in the fence can fail on its own merits;
that was already established earlier the same day, offline, against the same underlying
checking code, and re-running it through the cluster again would have cost real time for no
new information.

## Where we're going

Nothing is blocking. The one open thread is procedural: reading the advisory review's verdict
once it lands and acting on it if it comes back asking for changes — the code is already on
the shared branch either way, so this is about responding, not about permission to proceed.
Behind that, unchanged from before today: a backlog of components with no written test at
all, a filed-but-unowned question about documents left behind by renamed things, and a
visibility check owed to every existing fence now that the bug blocking it elsewhere has
closed.
