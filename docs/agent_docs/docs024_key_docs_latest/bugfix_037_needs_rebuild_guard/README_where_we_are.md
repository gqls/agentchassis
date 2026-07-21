# Where we are — bug 037 (needs_rebuild pages unprotected by the re-plan guard)

## 2026-07-21 (Claude)

You filed 037 as a deliberate decision, not a bug report: when the site planner re-plans a site, a
page marked `needs_rebuild` loses its previously-built layout to whatever the LLM proposes that run —
and you wanted a considered choice on whether that's wanted or wrong, rather than it being an accident
of how bug 001 was written.

I went and looked at every place in the code that puts a page into `needs_rebuild`. There are four,
and every single one keeps the page's existing layout and means "render this page again as planned" —
not "throw the layout away and start over". Two of them are the giveaway: they flag a page for rebuild
*because a component the page already lists has just become available*, so the layout is the whole
point of the rebuild. Re-planning that page and swapping its layout would defeat the very reason it
was flagged. So the "maybe this is what we want" reading doesn't survive contact with the code — it's
a genuine defect, silent loss of a real layout. Across the fleet right now 19 live pages are sitting
in exactly this exposed state.

So I made the fix: a `needs_rebuild` page's layout is now protected on a re-plan, the same way a fully
deployed page's is. I had to be careful, because another session was in the middle of editing the same
file for three other bugs (including bug 050, which is closely related — it's about what an *empty*
layout means). I did it in a way that slots in alongside their work rather than fighting it, and I
proved with tests that the careful version is necessary: the obvious shortcut would have broken a
different case (a page that genuinely is waiting to be laid out for the first time).

**It's already live.** While I was working, you ran one of your sweep commits and it carried my change
into the v1.0.1146 build, which is now running across the whole fleet — I checked the actual running
binary, not just the version tag. I committed my tests separately on top.

**Two things I've left for you to decide, neither of which is holding anything up:**

1. **Do you want a proper "redesign this one page" button?** After my fix, you can't accidentally
   redesign a built page by re-planning — which is the point — but if you *deliberately* want to
   recompose one, the way to do it is to clear its layout first and then re-plan. That works, but it's
   a bit of a dance. Bug 001 always intended a cleaner explicit signal (a "rebuild: true" flag) and
   deferred it as a product decision. It's still deferred. If you want it, it's a small feature to
   build; if the "clear the layout first" route is fine, we leave it.

2. **Do you want me to prove it on a real site?** The tests and the live-binary check are solid, but
   the gold-standard would be to actually re-plan a live site that has a `needs_rebuild` page and
   watch its layout survive. That mutates a real site and takes ~half an hour of the build queue, so I
   didn't fire it without asking.

I've left the bug in `/bugs_open/` (not closed) until you've had a look at those two.
