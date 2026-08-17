# SUMMARY — 2026-08-17 · the calculators are back in the plan

## What we're trying to do

loancalculator.co.uk was a hand-built site with eleven working calculators on it.
The goal has been to bring it fully inside the framework — so the site is planned,
built, checked and improved by the same machinery as every other site we run —
**without losing the calculators**, which are the product. They compute real money
figures, they were written by hand, and they are the one thing on the site that
cannot be regenerated from a description.

## Where we've come from

The calculators were protected by pinning them: twelve component rows locked so no
rebuild can overwrite them. On that footing the site was rebuilt through the
framework on 15 August. Most of it worked — twenty-eight of twenty-nine pages
rebuilt and served, the locked rows came through byte-identical.

But the tool pages came out of the rebuild with plans that described everything
except the tool. The first read of that was the worst possible one: that the planner,
even when shown the calculators, was choosing not to place them — which would have
meant the framework fundamentally could not host this kind of page.

That read was wrong, and finding out it was wrong is the most useful thing that
happened in the last three days. The diagnosis loop refuted it, and reading the
planner's raw response showed it had placed **every** calculator correctly. A
checking step downstream was silently deleting them: it validated the planner's
proposed sections against a list that only contained ordinary page sections, so
every tool-level name failed and was dropped before anything was written. The
planner had been right all along; the checker had never been taught what the planner
was now allowed to offer. That became bug 282.

## What we've done

Another thread fixed 282 yesterday — the checker now accepts exactly what the
planner was offered instead of re-deciding for itself — and today we confirmed the
fix is genuinely in the running system and then proved it on this site.

Confirming it mattered more than it sounds. The first check we reached for asked
whether the running program contained the fix's own commit, which it never would:
a binary carries the one commit it was built from, not its history. The honest
question is whether the fix is an ancestor of that stamp, and it is.

Then we re-ran the planning step over the twelve tool-carrying pages and counted.
Before the fix, none of the eleven pages that own a calculator had its calculator in
the plan. After: **all eleven, each with its own, in the right place.** Verified by
joining the locked components' function names against the written plan, page by
page — not by reading a list and nodding. Nothing was recomposed into a no-op, and
the twelve locks were still intact afterwards.

We also finished the outstanding arithmetic check on the calculators themselves.
The comparison tool reported "eleven of eleven diverged", which reads as a disaster
and is not one: of 1,340 recorded values, not a single one changed. Every difference
is the new page furniture arriving and the old hand-built navigation leaving. We
re-recorded the baseline against today's pages so the next run compares like with
like — and discovered on the way that the handoff had been pointing at the wrong
baseline file for two days: a discarded, known-broken one from 31 July.

## Where we are now

**The blocker that has defined this lane is discharged, and the plan for the site is
correct.** That is a plan-level result, and it is complete and checked.

The pages have not yet been rebuilt from that plan. The rebuild jobs were queued
correctly and then nothing moved, because the estate's entire job queue is shut: the
Anthropic account's spend limit was touched briefly this morning, a health record
flipped to "unavailable", and everything that claims a job consults that record
first. The API recovered within minutes; the record is only re-checked once an hour.
So the estate is idle behind a stale flag rather than a real outage. Three other
threads hit the same wall this morning; it is written up as bugs 243 and 244.

Two things are worth the owner's attention, neither of them faults:

The rebuilt site has lost its **Tools menu**. The old header listed nine
calculators; the new one has Home and About. This is the framework behaving as
designed — it deliberately keeps individual tool pages out of the top menu and
expects a single parent "Tools" listing page to stand for them. This site has such a
page for Guides and none for Tools. The calculators are still reachable from links
in the body of every page, so nothing is stranded, but the navigation is poorer than
what it replaced.

> **CORRECTED later the same day (~16:40Z) — the paragraph above overstates the loss.**
> The framework did put all eleven calculators into the site's navigation, in the
> **footer** (`utility` group), which is exactly what it does with a page it bars from
> the top menu. Verified at three layers: 11 `utility` rows in `site_nav_items`, all 11
> hrefs in the regenerated footer component, and a guide page republished at 16:21
> serving a footer containing every calculator.
> **What misled me:** I sampled pages published at 13:44:08, and the shared chrome
> regenerated at 13:47:45. A page carries the chrome it had when it was last rendered,
> so the change was invisible on anything not yet re-rendered — and the pending
> re-renders were failing, which made the stale state look settled.
> So the true change is **header dropdown → footer list**, not lost navigation. Only the
> header question remains, and it is a preference rather than a repair.

And the **Guides page** is still the site's one missing page. Its menu entry is
already recorded and it is explicitly not barred from the menu, so building the page
should restore it.

## Where we're going

Confirm the queue re-opens, then let the wave run and check the homepage rebuild at
the artefact — it is the one page in that wave carrying a locked calculator, and the
protection has held once already under worse conditions. Then the eleven tool-page
rebuilds, which sit behind a deliberate human gate and are the owner's to release
now that the reason for holding them is gone.

After that the remaining questions are choices rather than problems: whether to
build a Tools listing page, whether to un-park the design and imagery work that is
currently holding the visual refresh, and whether the temporarily released locks on
the chrome and stylesheets should be re-armed at all now that both have been rebuilt.
