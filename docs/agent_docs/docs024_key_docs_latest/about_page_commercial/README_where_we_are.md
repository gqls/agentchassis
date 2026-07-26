# Where we are — about-page commercial elements

*(Owner's plain-prose log. Append below, never rewrite what's above.)*

## 2026-07-24 — the design is settled, starting the smallest real test

We worked out what the about pages should carry on the sites we build to sell.
Three things, all quiet and tasteful, at the foot of the about page: a line
saying the domain name can be acquired (with the enquiry going to Afternic so
their minimum-offer floor filters out the £50 chancers — we never print a price
on the page); a line offering ad space on the site (pointing at advertise.co.uk,
which is being built separately — flat monthly rate, honest that these are small
sites); and a line saying who built the site (fundamentallyai.com), which quietly
markets the platform itself. The footer on every page gets just the built-by
line, nothing else, so the sites never look like parked junk.

Wording matters here: we dropped "premium" and "serious offers" as worn-out
domain-trade talk and settled on "available to acquire … enquiries via our
domain team", which sounds represented rather than desperate. Three tiers of
name get three strengths of wording; the best names route to Afternic's
brokerage.

Two safety rules are baked in from the start. Nothing shows unless a site is
explicitly marked as portfolio stock — so a client's site can never accidentally
carry "buy this domain". And a site that currently has a paying advertiser never
shows the for-sale line (though the quiet Afternic listing stays live, so a
serious buyer can always find us).

The switch that turns for-sale on and off will be an API the admin page drives
now and advertise.co.uk can drive later, automatically.

We chose to build incrementally: first a database-only pilot on one site to see
the block rendering live, then the platform code (footer line, admin switch)
which needs a proper review and an image release. We're currently checking one
technical question: whether we can attach the new block to one about page and
re-render just that page without a full rebuild — the platform has burned us
before on "the fix is in but the page never re-rendered", so we're verifying
rather than assuming.

## 2026-07-24, evening — pilot built and fired; it tripped over an older bug

The checking paid off: there is a safe way to add the block to one page, and we
built it. The new block is in the component library, finetuning.uk was chosen as
the guinea pig (quiet site, nobody else working on it), and its settings were
written so that only the "Built by fundamentallyai.com" line would show — the
for-sale and advertise lines stay hidden until there's a real Afternic listing
and a real advertise.co.uk to point at. We deliberately started with the honest
minimum.

When we pressed the button, the rebuild failed — but not because of anything we
added. It hit a fault that's been sitting in the platform since at least the
16th: the "rebuild one page" machinery hands its content writer less information
than the writer's link-checking step insists on, so every single-page rebuild
dies at that step no matter whose page it is. Two earlier failures in the logs
show someone hit this eight days ago and nobody filed it. We've now put it
through the formal diagnosis loop (so the cause gets verified and fixed
properly, not patched twice by two threads), and our pilot is still armed: the
page is untouched, still flagged, and the moment the rebuild path is fixed it
will pick our block up with no further work.

So: design done, block built, pilot armed, one pre-existing platform bug found
and filed — waiting on the diagnosis verdict before the last step.

## 2026-07-26 — the platform bug is fixed, the pilot had already landed, and there is one more bug behind it

Written by the thread that took `bugs_open/068` (the platform fault this pilot tripped over),
not by this workstream — appended here because it changes what you'd find if you came back to it.

Three things, in order of how much they matter to you.

**The block is already live, and has been since the evening of the 24th.** The pilot did not
stay stuck. About an hour after the failed rebuild, the ordinary build pipeline picked the page
up and rebuilt it the normal way — all seven of its parts were written in the same second, the
commercial block among them. So finetuning.uk's about page carries "Built by
fundamentallyai.com" today, with the for-sale and advertise lines correctly hidden. The last
entry here says we were waiting on the fix before the block could appear; that was true when it
was written and stopped being true a few hours later, by a route nobody was watching.

**The platform fault you filed is fixed and now proven.** The contract change that was applied
on the 24th was never actually exercised, because nothing ran the rebuild path again. Today I
re-armed your page, fired your trigger script, and watched it: the step that used to kill every
rebuild now passes. Your page was not harmed — same content, same block, and I put its status
back the way I found it. `068` is closed.

**But the rebuild path still does not finish.** One step further on, it dies again, for an
unrelated reason: the "rebuild one page" machinery hands the writer no plan of what sections to
write, and nothing in the writer can work that out for itself. Filed as `bugs_open/087` with
three ways to fix it, one of which is a small config change. Until someone takes that, the way
to rebuild a single page is the ordinary pipeline — flag the page and let the dispatcher do it,
which is exactly what happened to your page on the 24th.

So: nothing is owed to this workstream to get the block on that page — it is there. What is
still open is the rebuild *trigger* you scripted against, and that now has its own bug file.
