# vetcomparison.uk — milestone read-out, 26 July 2026

Written to be read aloud. Figures re-checked against the live system on the day of writing; where
a figure comes from another session's work the same day, it says so. For the story in order see
`README_where_we_are.md`; for evidence and commands, `NOTES_` and `RUNBOOK_`. Previous read-out:
`SUMMARY_2026-07-19_readout.md`.

## What we're trying to do

Make vetcomparison.uk the place UK pet owners go to compare what veterinary care costs, and the
place practices go to publish those costs. The timing is the opportunity: the Competition and
Markets Authority has finished its investigation into the vet market and has published a draft
Order that will require every practice to publish a standardised, comparable price list.
Comparison becomes possible for the first time, and we intend to be where it happens.

Two rules constrain everything. Anything built on the platform must be generic, so the next
comparison site reuses the machinery rather than having it written again. And we never publish a
figure we cannot show the source for — either the practice gave it to us, or we can link to where
they published it, with the date we saw it.

## Where we've come from

The site went live in February publishing invented prices for roughly 3,100 named real practices,
under a false claim that the data was ours, next to a quotation attributed to the CMA that the CMA
never wrote. In mid-July we stripped every price, quarantined the fabricated data, removed the
guides, and wrote a dated factual record of what had been published — the document a solicitor
would be briefed from if a practice or the regulator ever asks.

What survived was the good half: a verified directory and a working collection pipeline. By the
last read-out on 19 July we had rebuilt on that base — the directory cleaned to 2,109 genuinely
verified practices, three CMA guides rewritten with every figure sourced, the database unified so
services and medicines share one structure, and a generic exporter publishing the directory and
consented price lists for any vertical and domain.

## What we've done

**We found the site's own machinery inventing prices again, and stopped it.** The medicine price
scraper reads retailer pages; when its ordinary parser finds no prices it hands the page text to a
small AI model running on our own hardware and asks it to extract them. Shown a page with no
prices on it, that model does not answer "none" — it invents a plausible table. Two of those
invented prices reached the live site. Following the evidence found **212** fabricated rows across
two eras, including seventy-nine copies of a price the model had simply read back out of the
example in its own instructions. All 212 were quarantined and removed, and the poisoned summary
values reset.

The fix is a check at the moment of writing: a price is stored only if that exact figure appears in
the page text we keep as proof. If it doesn't, the price is dropped and the refusal is logged and
counted. AI-extracted rows also now carry their own label, so they can never again be mistaken for
parsed ones, and the model is only consulted about text that actually contains prices.

**Today we proved it works, and very nearly didn't.** The fix had been running in production for
two days. Every check was green — all 2,583 prices we hold can be found in their own evidence, the
published file had been rebuilt clean, and the AI fallback had been asked for prices sixteen times
since the fix and returned "nothing found" every single time. I concluded from that the model had
stopped inventing things, so the check was really just a backstop, and wrote that down as a reason
there was nothing left to test. **I tested it anyway and the model invented three prices on the
first attempt** — £19.25, £34.99 and £68.75, plus a pack size that does not exist on that page. The
check caught all three and stored nothing. So the honest finding is the opposite of what I was
about to record: the check is not a backstop, it is the thing doing the work, and the sixteen quiet
answers were simply other pages. That is now written into the case file, the debugging guide and the
fleet-wide log of things we asserted that turned out to be false.

**Separately, a second session diagnosed why the site still feels thin** and the owner has set the
direction from it. That work is today's `PLAN_2026-07-26_site_strength.md`.

## Where we are now

The live site is honest and narrow. It publishes 2,109 practice names with postcode and an outbound
link, three CMA guides with sourced figures, and a working news feed that picked up the draft Order
by itself. Medicine prices are the one price dataset we can publish: **2,583 snapshots across 306
active listings from three retailers**, every one of them verifiable against the page text captured
at the same moment, with 76 products and 99 variants in the file published on 25 July.

The thinness has a specific cause, and it is not missing pages. **We are under-publishing, not
mis-publishing.** Nearly everything else we hold is barred by our own provenance rule: 762 practice
price rows with no source URL, ownership data for 1,102 practices that has no evidence trail and
contradicts our own independence flag on 870 of them, and practice detail last touched in March
with nothing recording where it came from. (Those four figures are from today's plan, verified by
its author against the live database.) None of it is live, and the fix is to earn better data
rather than to publish what we have.

The CMA consultation responses have been **dropped** on the owner's decision — the funding deadline
of 30 July and the substantive one of 20 August are no longer being worked.

One risk sits underneath all of it. The shared clients database is running with no resource
guarantee and a one-second health check, so when the machine gets busy it is killed for being slow
and every service loses its database until it restarts. That was filed today by another session as
an open bug. My own finding is that the medicine scraper is a likely trigger: its AI fallback runs
on the same machine as the database and occupied it for eight and a quarter minutes today, and the
database failures start two minutes into that call and stop before it ends. I have not touched that
bug or its fix — it is another session's, and it deliberately has not been patched in production.

## Where we're going

The next move is already chosen and is the cheapest one with the largest unlock: **collect
practices' Companies House registration numbers from their own websites**, which UK companies are
legally required to display. It is self-declared by the business, carries its own source URL, and
uses pattern-matching rather than an AI model — so it introduces no new invention risk, which
matters more on this site than anywhere. The machinery is already built and live; it simply was
never switched on. It is database configuration, so it takes effect immediately with no rebuild.
The plan is to pilot on about 25 practices, report the real hit rate, and only then ask the owner
whether to run it across all 2,109.

That one crawl feeds the two things that follow: ownership that is **derived from evidence rather
than asserted**, which should also resolve the 870 contradictions without editing a single row by
hand; and per-practice pages that finally have something on them. In parallel, the compliance
deadline calculator is now a build rather than a cancel, with one firm constraint — every figure and
date in the draft Order is still a bracketed placeholder, so it can honestly say "six months after
the Order is made" and must never print a calendar date. The homepage's dead links get cleaned up
last.

Three decisions are still yours: the five standing review items, whether to scale that crawl beyond
the pilot once we know the hit rate, and what `/claim-listing` and `/search` should actually point
at. On my own work the only thing outstanding is whether to write the database-outage trigger I
found into the other session's bug file; say the word and I will.
