# CONTRIB — WebProNews RSS feed: owner-flagged candidate news source (2026-09-02)

**From the portfolio_positioning lane, at the owner's direction** (message 2026-09-02 evening):
*"Please can you record the webpronews.com rss feed details currently being consumed by
advertise.co.uk and we can use this on another site as it looks quite good. we could add it to
the news sources."* Routing to this lane because you own news sources; adding it (and choosing
which site consumes it) is yours.

## The feed, measured 2026-09-02 ~17:5xZ

- **URL:** `https://www.webpronews.com/feed/` (200, follows from webpronews.com/feed)
- **Size/shape:** ~1.08 MB, **100 items** in one fetch, RSS 2.0, channel title "WebProNews"
- **Cadence:** very active — two freshest pubDates 16:12:18Z and 16:02:16Z the same afternoon
  (multiple items/hour)
- **Breadth:** general business/tech/finance (sampled titles: Apple Vision Pro in surgery, AI
  layoffs, battlefield robots, German auto industry) — broad, not vertical-specific; which site
  it suits is a positioning question, happy to advise if wanted
- **Not yet in the machinery:** `content_sources` has **0** rows matching webpronews as of
  2026-09-02.

## Provenance and the one caution that travels with it

The consumer the owner refers to is the **OLD advertise.co.uk** — a Drupal 7 auto-aggregator
that imports WebProNews articles **wholesale** (the domain still serves it; the new framework
site deployed 2026-09-02 16:23Z awaits DNS cutover, after which this old consumer disappears).
The classifier's own reading of the old site: *"Articles imported wholesale from WebProNews —
no original content."* So the owner's endorsement is of the FEED's content quality, not of the
wholesale-import pattern — through your pipeline it should get the same editorial treatment as
any other source, not verbatim republication.

The old site's rendering of it is snapshotted at
`portfolio_positioning/salvage/advertise.co.uk/` if you want to see what the owner liked.
