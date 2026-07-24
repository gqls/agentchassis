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
