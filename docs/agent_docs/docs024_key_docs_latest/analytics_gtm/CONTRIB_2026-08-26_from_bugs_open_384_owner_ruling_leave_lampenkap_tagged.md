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

---

## ADDENDUM 2026-08-26 — the second fact, answered by the lane that owns it

The section above deliberately left one thing unasserted: whether the applied key results in GA4
reporting. The `analytics_gtm` lane has now supplied the measured answer (container re-read
10:50Z, not from memory), and it is recorded here so a reader of this file gets the whole
picture rather than half of it:

- **`GTM-PQ3WCTBD` is a Tag Manager CONTAINER** — a socket now fitted on every estate site,
  lampenkap included. **GA4 is the appliance**, identified by a `G-` prefixed measurement id on
  the Agent Chassis property. Different things, and the `GTM-`/`G-` prefixes are how to tell
  them apart at a glance.
- **The live container is at version 2 with ZERO tags and no `G-` id anywhere in its `gtm.js`.**
  So today it fires nothing, and **no estate site reports into GA4 yet** — lampenkap or any
  other.

So the owner's ruling settles the first fact only, exactly as this CONTRIB said before the
answer was available: lampenkap stays tagged. The second fact — anything actually reporting —
is not true for ANY site today.

It becomes true for all ~30 sites at once when the pending step is completed: a Google Tag in
that container carrying the Agent Chassis measurement id, then Submit → Publish. That is an
owner action in the Google console, with a consent decision first (`039` §4a); the walkthrough
is in the `analytics_gtm` handoff §2 and apis.uk's `HANDOFF_2026-08-25` §4a. One command
verifies it the moment it happens — `analytics_gtm/scripts/check_gtm_state.sh`, whose verdict
line flips to `PUBLISHED`.

**Why this matters beyond lampenkap:** "the sites are tagged" and "we are collecting analytics"
read as the same statement and are not. Anyone reasoning from the first to the second — for a
report, a client answer, or a decision about traffic — would be wrong today, for every site.

— `bugs_open/384` lane, 2026-08-26, from `analytics_gtm`'s measurement
