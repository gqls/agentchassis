# SUMMARY 2026-08-18 — webdesign.uk build service

## What we're trying to do

Make webdesign.uk sell and deliver a £149 website without a person in the loop, and do it
using the framework itself rather than by hand. Two halves: the site has to explain an
unusual offer honestly, and the chat box on it has to answer real questions from real
visitors without ever inventing a price, a promise or a timescale.

## Where we've come from

The chat box began as a one-off: a hand-built widget spliced into one page, backed by a Go
service on a rented box with the business's facts typed into its source code. It worked, and
it was a dead end — a second site meant doing the whole thing again by hand, and the facts in
the code had already drifted from the facts in the database once.

The last few weeks turned that one-off into a capability. The widget became a registered
library tool. The bot stopped carrying its own copy of the facts and started reading them
live from the database. And a plan was written to close the remaining gap: letting the
framework decide, deploy and maintain a chat box on any site, the way it already does for
calculators.

## What we've done

**That plan is finished — all six steps.** The last two landed this week. The deployer now
refuses to put a backend-needing tool on a site that cannot run one, and files a plainly
worded request for a human when it needs a backend provisioned. The chat program is no longer
webdesign's: one binary now serves any site given its own identity and facts, and it refuses
to start if it is handed the wrong site's facts by mistake.

**A defect that had been quietly eating the chat box is fixed and proven.** Every rebuild of
the contact page had been proposing a version of the page with the chat box missing; only a
lock stood between that and deletion, and before the lock existed it had already been deleted
once. Another thread fixed the cause; we ran the acceptance on the real page and all five
checks passed. The lock is now off — safely, because we first put the chat box into the
page's plan, so it is held there by design rather than by exception.

**The commercial terms changed at the root, not on the page.** The owner ruled that payment
comes before the build, that customers do not see the site first, and that what they get is a
ZIP to keep plus a preview link for about a month. Rather than editing pages, we changed the
register of attested facts — which is what the writers are allowed to say and what the checker
enforces — and moved the billing system's own payment switch in the same breath, so the copy
and the software cannot disagree. Within minutes the live chat bot stopped telling customers
they could see the site before paying. It had been saying that all morning.

**And the chat box is now on the home page**, where the owner wanted it: straight after the
price block. Getting there took two refused rebuilds, and the refusals were the system working
properly — the writer kept turning "usually ready the next day" into a hard statistic, "1 day",
and the claims checker would not publish a number nobody had attested. We fixed that by
telling the writers that hedged facts never become figures, rather than by attesting "1 day"
as true, which would have quietly converted the owner's hedge into a promise.

## Where we are now

The site states the new terms, the bot speaks them, the chat box is live on both the home and
contact pages, and no link was lost in any of it. The owner's voice brief is in the spec: write
like a helpful assistant, not a marketing bot — say the thing, then give the next step, or the
order to do it in, or the name of someone who can help.

Two things are open and both are honest gaps rather than surprises. The site promises a ZIP to
keep and a month-long preview; the delivery lane's download link currently lasts seven days
and we could find no month-long preview mechanism at all — so a promise is running ahead of
the machinery, and we have written to that lane saying so. And the home page still carries a
button whose words offer a helpful questionnaire while its link dials the phone; it is filed,
and deliberately not patched, because the section is about to be rewritten anyway.

## Where we're going

The owner is settling how the site should lead. The proposal is to show the work rather than
promise anything: real sites built by the system, each shown with the exact prompt that made
it. That suits a business where every build differs, it survives the cheaper second brand he
plans to launch, and it is the honest answer to the one hard question the new terms raise —
if I cannot see my own site before paying, what can I see? The examples come from his own
domains, once he is using the system in earnest.

After that: the site rewrite in the new voice, the delivery lane's handover and email, and a
prompt maker so a visitor can arrive with a rough idea and leave with a brief good enough to
build from. The owner's instinct there is to teach the chat box we already have rather than
build a second one, which is also the cheaper answer.
