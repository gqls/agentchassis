# README — news_feed_ingestion, where we are

2026-09-02. Another session working on a boxing-betting-style news site
(boxingonline.com) found that the site's "Fight Calendar" tool page shipped with
a big headline and a paragraph of text about itself, but no actual fights listed
— no dates, no venues, nobody in the ring. They traced it back to something
structural: the platform has a shared store called `evidence_base` where each
site is supposed to keep verified facts it can build pages from, and separately
a news-ingestion pipeline that pulls in articles from around the web and scores
them for relevance. Both halves exist and work fine for what they were built for
— but nothing in between ever takes "this article confirms a specific fight is
happening on this date" and turns it into a fact the calendar tool could actually
render. They checked: across the whole fleet, 14,013 news items have come
through, and not one has ever had that kind of structured detail extracted from
it. This isn't a boxingonline-only bug — it's a missing capability, and
boxingonline is just the site that hit it hardest because a customer is paying
for a working fight calendar.

I picked this up because I'd just been named "feed lane" for this kind of work,
and because the piece of the fleet that would need to build the fix — the news
feed pipeline itself — didn't have anyone actively working it. I checked the
other session's findings myself before taking their word for it (re-read the
code, re-ran their database counts, checked nobody else had quietly started the
same fix) and it held up, so I've taken it on.

The plan: rather than inventing something new, reuse a pattern the platform
already has twice over — a step that verifies a claimed fact against its source
before it's allowed to go live. I'm adding a third use of that same pattern: pull
in news items that have already been scored as relevant, ask an AI step whether
any of them confirm a specific dated event (and if so, what the date/venue/who's
involved actually is), verify that against the source article, and only then
write it down as a fact. Nothing gets invented — if the article doesn't say
something clearly, it doesn't get written down as if it does.

This is the first, most urgent piece of a longer-term job — the owner has ruled
this needs to be fixed before boxingonline.com's site is delivered to the
customer, so it's not just a nice-to-have. There's a second piece (keeping dates
correct as they change — fight dates move) and a third (a proper page to render
the fixtures onto) that come after, plus an older, separate feed-scheduling bug
(some sites' news updates were consistently arriving late) that I'll pick up once
this is done.
