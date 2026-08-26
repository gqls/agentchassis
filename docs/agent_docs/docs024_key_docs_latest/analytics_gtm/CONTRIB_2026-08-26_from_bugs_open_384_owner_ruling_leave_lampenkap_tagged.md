# CONTRIB 2026-08-26 — OWNER RULING: leave lampenkap.com tagged. No retraction.

Filed by the `bugs_open/384` (page-list invalidation) lane, at the `analytics_gtm` lane's
request, so the ruling lives in the directory that is the single record for everything Google
rather than in a chat thread or in my own lane's notes.

## The ruling

> **"leave lampenkap google tag"** — owner, 2026-08-26.

**No supersede. No retraction. lampenkap.com keeps the key applied in the 2026-08-26 morning
apply and reports like any other estate site.**

## How the question arose, and why it was asked at all

The `analytics_gtm` lane flagged (per `bugs_open/397` §9) that lampenkap.com was included in the
spec-key apply, noted it was born 2026-08-25 with no `site_config` row and had never been
tagged, and asked whether it was a throwaway test site that should NOT report into the estate's
GA4 — offering to retract its key if so.

I was asked because lampenkap had turned up in my lane's sweep readings. I declined to answer
it: the question is a judgement about what the site IS FOR, and the data underdetermines it.
The facts I supplied, `[MEASURED 2026-08-25]`, were:

- site row created 2026-08-25 11:30Z, status `deployed`; its single page (`index`, page_type
  `landing`) deployed 12:06Z and has not moved since;
- **one active page, and that is all** — no tool pages, no blog, nothing else;
- its `index` carries a `tool-list` component declaring `query.pages_where_type:tool` against
  **zero** `tool` pages, so that listing resolves empty.

That shape fits a throwaway test site and a real site caught one day into its build equally
well, which is precisely why it went to the owner rather than being inferred here.

By the time the ruling came, the key had already been applied, so the live question had narrowed
from "should we tag it" to "should we retract" — one supersede. The answer is no.

## Not a symptom, and please do not read it as one

lampenkap appears in `page_list_stale` sweep findings as
`consumer_pages: 1, stale: 0, current: 0, unknown: 1`. **That is correct and expected**, not a
fault and not related to the tag: the site has one page and zero `tool` pages, so its `tool-list`
array is legitimately empty, and the sweep classifies an empty resolve as UNKNOWN by design
(never as "current"). The GTM re-render will not change that reading, and nobody should expect it
to. The `analytics_gtm` lane has this noted at NOTES §17.

## What this CONTRIB does not decide

Whether the applied key results in GA4 reporting for this site depends on what the container
fires, which is the `analytics_gtm` lane's territory and not asserted here. The ruling recorded
above is about the tag staying on the site; the wiring behind it is theirs.

— `bugs_open/384` lane, 2026-08-26
