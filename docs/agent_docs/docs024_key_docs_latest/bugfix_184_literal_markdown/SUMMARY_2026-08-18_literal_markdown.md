# Summary — bug 184, literal markdown on live pages — 2026-08-18

**What we're trying to do.** Our AI content writers sometimes type markdown formatting
symbols — `**bold**`, `# headings`, backticked code, `[link text](web address)` — into
page fields that the site shows as plain text, so visitors see the raw symbols instead
of formatted words. We want those symbols off the live pages, and we want the mistake to
stop recurring, across every site the framework builds.

**Where we've come from.** The bug was filed on the 3rd of August on three rows. The
detector built then works well and has been finding real defects ever since. The repair
never worked: it asked an AI to rebuild each affected page, and the rebuilding AI has
the same habit — on the 7th of August a full rebuild typed the markdown straight back
into the very field it was sent to clean, three days after a prompt rule forbidding it
went live. That repair route succeeded 3 times in 39 attempts, the queue machinery
sensibly stopped feeding it, and by tonight 71 defect items sat open across 6 sites with
new ones arriving daily and raw symbols confirmed on live pages.

**What we've done.** Tonight the repair went mechanical. One small piece of ordinary
code deletes the markdown symbols from plain text — it can only remove characters, never
add any, so it cannot inject anything — and it is shared by the detector, the repair,
and the completion checker, with a test guaranteeing they can never disagree. The repair
is now: re-run the existing, highly reliable page re-render (99% success over ~14,000
runs) with the cleaner switched on for exactly those repair jobs and nothing else. The
same cleaner also guards the two places where AI-written content is born, so new pages
come out clean. The detector was widened to catch markdown links, today's commonest
form. Everything is committed and rides the next fleet release; the two database
switches that arm it are written, reviewed, and waiting for the release. The review
council took this through four rounds — the first three asked for revisions, two of
which caught genuine design flaws (the cleaner would have run on far more pages than
intended, and its actions weren't durably recorded); both are fixed — and round four
approved it.

**Where we are now.** Code complete, tested, committed, council-approved. Not yet live:
the fix ships with the next whole-fleet release (the image tag is bumped ready), after
which the two switches get applied and a two-page canary proves the repair on real
pages before the remaining ~70 items are fed through it.

**Where we're going.** After the release: canary, then batch repair of the backlog, then
the detector's next sweep should confirm the sites clean and close the items on its own.
The bug file closes when the founding pages and the widened-symptom pages are verified
clean at the served page, not just in the database. One named follow-up: the report
pipeline is the one remaining AI writer without the birth-time guard — it has no live
defects today and the detector covers it; if that changes, the same switch pattern
covers it with one small migration.
