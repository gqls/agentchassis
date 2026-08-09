# 031 FEATURE — pages carry a "last checked" date

**Raised:** 2026-08-09, by the owner, while ruling on the `claims_unverified`
retraction gate (`bugfix_168_deployed_asset_path`).
**Status:** idea captured. Not designed, not scheduled.
**Related:** `docs024_key_docs_latest/bugfix_168_deployed_asset_path/` (the ruling
that prompted it), concept register CQ-020 / CQ-021 (the two shared scanners that
would be the natural writers), `bugs_open/033` and `bugs_open/083` (the review
queue this would give a freshness axis to).

## The idea, in the owner's framing

> "Maybe we should date the pages when last checked?"

Raised immediately after choosing option (b) for the claims retraction — the gate
that refuses to close a factual-claim finding unless the page's copy has actually
changed since the finding was filed.

## Why it came up there, and why it is a different instrument

Option (b) proves *the page moved*. It does that by comparing
`page_components.updated_at` against the work item's `created_at`. That is a
**proxy, and it only answers one direction**: it can tell you the copy changed,
and it can tell you nothing at all about whether anyone ever looked.

Today the platform has no answer to any of these:

- **Which pages have never been audited by check X?** Not "which pages have
  findings" — which pages were never *examined*. A page with no findings and a
  page nothing ever scanned are indistinguishable in every table we keep.
- **Which pages were cleared by a version of a check that we have since fixed?**
  This is a live, recurring cost. When a detector is corrected, the pages the
  blind version passed keep their clean bill of health for ever — nothing records
  which reading cleared them, so nobody knows what to re-run. It has already been
  written up as a standing trap ("a PASS from a BLIND check outlives the
  blindness"), and every occurrence has been handled by hand.
- **Is the review queue's silence about a page trustworthy, or just old?** The
  sweep reports what it judged. It has no way to say "and these 300 pages have not
  been looked at since June".

A last-checked date turns all three from archaeology into a query.

## What it is not

Not a content-freshness feature. `features_open/004` and `/006` both touch
"is this *content* stale" — a 2024 statistic ageing badly, a shared topic needing
a refresh policy. This is the orthogonal axis: **is our KNOWLEDGE of this page
stale**, regardless of whether the page itself has changed.

The two interact usefully but must not be conflated: a page can be freshly
checked and badly out of date, or recently rewritten and never examined.

## Rough shape (not a design — first thoughts only)

The natural writers already exist and already run over every page: the shared
scanners extracted this week for the voice and claims audits (`ScanVoiceTells`,
`ScanDeployedClaims`, register CQ-020 / CQ-021). Both walk every deployed
component of every page on a site and already know exactly what they examined and
what they skipped as human-locked. Neither currently records that it was there.

Open questions, all genuinely open:

- **Per page, or per page-and-check?** One date per page is cheap and answers
  "has anyone looked lately". A date per (page, check) answers the re-run question
  above, which is the more valuable one and considerably more storage and
  bookkeeping. Probably the deciding question for the whole feature.
- **Does it record the check's VERSION, not just the time?** The blind-detector
  problem is only solved if you can ask "which pages were last cleared by a
  reading older than the fix". A bare timestamp does not answer that; a timestamp
  plus something identifying the reading does.
- **Where does it live?** A column on `pages` is simplest and immediately
  queryable. A separate table is the honest shape if it becomes per-check, and
  avoids widening a hot row for bookkeeping.
- **Does "checked" mean examined, or examined-and-clean?** These must not collapse.
  This week's work established the same distinction inside the scanners — a scan
  that read nothing returns the same empty findings list as a page that is
  genuinely clean, and conflating them is how a no-op reads as a success. Whatever
  is stored has to preserve it: "examined 0 components" is not a check.
- **Cost.** The scanners run over every component of every page already, so the
  read is free; the write is the question, and a naive per-component write on
  every pass would be the expensive way to do a cheap thing.

## Why it is worth doing eventually

Every mechanism this estate has built for retraction so far answers *"is this
finding still true?"*. None answers *"how confident are we that we looked?"* —
and the second question is what makes an empty queue mean something. An empty
review queue currently has two indistinguishable causes: everything is fine, or
nothing was examined. That is the same ambiguity, one level up, that the
`ComponentsExamined` counters were added to kill inside a single page scan.
