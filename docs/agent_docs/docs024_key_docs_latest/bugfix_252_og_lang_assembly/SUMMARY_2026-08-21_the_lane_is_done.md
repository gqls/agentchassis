# SUMMARY — 2026-08-21 · the og/lang lane is done

The milestone read-out for bug 252 (og/lang slug) and everything it pulled with it. Current state
only; the chronology is in `NOTES_og_lang_assembly.md` and `README_where_we_are.md`. First summary in
this lane's series.

---

## What we were trying to do

Make an assembled page able to say **who it is** when someone shares it, and **what language it is
in** — two things a page built by the framework could not do, on a British estate where every page
declared American English.

## Where we came from

Filed on 11 August as a modest, accepted loss: assembled pages carried no per-page share tags, and
`<html lang="en">` was hardcoded in Go. It was escalated from one lane's private note to platform work
because the scale had grown from 2 pages to 503.

**By the time we picked it up, it had become a different and worse bug.** A well-meant change in
between — the share-card work — started writing share tags into each site's shared page-header. Those
tags include "the address of this page is…", and at that moment there is no page to ask, so it filled
in the **homepage**. So pages had stopped merely lacking their own identity and had started actively
asserting the wrong one, on 700 pages across 26 sites, beside a canonical tag that named the page
correctly. The page contradicted itself. Four sites were worse still, serving the same tag twice —
once blank, once wrong.

That mattered beyond bookkeeping: **a missing tag is silence and a scraper falls back sensibly; a
wrong tag is followed.** It also killed the fix the bug file proposed — leave blanks and fill them —
because the sites with blanks already had them overridden a few lines later.

## What we've done

**The fix.** A page now removes whatever the shared header claims about *that page's* identity and
states its own — title, description, and address. The address goes through the same helper that
produces the canonical tag, so the two cannot drift apart; before, they were two separate calculations
kept in step by a comment. Pages with no description get no description tag rather than an empty one.

**The language.** It moved out of Go and into each site's page-header, per the owner's ruling, with
the value in site configuration. Twenty-five sites declare British English; **relojistas.com declares
Spanish** — it is a Spanish-language publication, and a blanket "British" would have been false
metadata stated more confidently than the "English" it replaced. The owner ruled that this
generalises: non-English sites must never be defaulted to British English.

**Three things this uncovered, all fixed.** Our largest site, webdesign.co.uk, had **no page-header
element at all** — the opening and closing tags were simply missing from its component, which is why
that site alone silently received no share image and no favicon tags, on 117 pages, while every tool
involved reported success. New sites had no way to get a language, so the next one created would
default to English with nothing to notice. And the guard behind all of it — one hand-authored tag
switching off an entire block of tags — is now per-tag, so an authored tag is preserved and only the
missing ones are added.

**Everything went through review.** The main fix was approved first time. The guard fix came back
**REVISE**, and that round was worth more than it cost: it found that I had written a careful comment
explaining a *silent failure* in the middle of a fix for silent failures, and that a comment of mine
contradicted the code it described in a way that would have led a later reader to delete a working
fallback. Both fixed; approved on the second round.

## Where we are now

**Everything upstream of the pages is repaired and proven on the running system** — not inferred from
a version number, but read out of the binary on both machines with controls, and then read off live
pages.

- 24 stored page-headers: **23 declare a language**, **none** still bakes the homepage address,
  **none** carries a duplicate blank tag, **none** unlocked is missing brand tags. The 24th is
  deliberately locked hand-authored chrome and was correctly refused.
- **Every page that rebuilds from here is correct.** Proven on four real pages across three sites.
- Bug **252 is closed**. So is **347** (the missing page-header element), filed and fixed the same day.
- **322 item 4** — the underlying guard — is fixed, live and approved.
- A **daily check** now reports any site with no language, or with a language its header cannot
  render. It caught a real case in its dry run before it was even switched on.

**What remains is not work, it is time.** 487 of 727 pages still serve their old header, because a
page only picks it up when it next rebuilds. The owner ruled that we do not force those rebuilds —
about 500 queued items would be other lanes' risk — so they heal as sites are worked on. That is
already happening: **ten sites at zero this evening, down from twelve this morning, with nothing
dispatched at them.** It is tracked in `bugs_open/346` as a tick-list, and the honest expectation is
that busy sites drain in days and quiet ones may sit for months.

## Where we're going

Nothing on this lane needs a next session. Three things are written down and unclaimed:

- **`bugs_open/322` items 2, 3 and 5** — the share-image emitted whether or not the file exists, the
  fallback copy quality, and wide logos making illegible favicons. Item 3 carries a landmine against
  the obvious fix.
- **The other page-header producer.** Pages built through it get no brand tags at all and none of this
  reaches them. We could not size that population honestly and said so rather than guessing.
- **The convergence question** the register has held open since July. This was the **fourth** fix to
  land on one of the two page-header producers, and the review council's architecture seat set a
  threshold: **a fifth raises an RFC rather than taking a fifth patch.**

Plus six incidental findings in `FINDINGS_2026-08-21_errors_caught.md` — including a diagnosis tool
that cannot see what a site actually serves, and migration numbering with no allocator, which
collided twice during this work alone.
