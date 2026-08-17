# SUMMARY 2026-08-17 — bugs_closed/285 and its delivery half (read-out at the second close)

**What we're trying to do.** Make sure a fix aimed at one page cannot damage others, and that the
checks meant to catch such damage actually can see it.

**Where we've come from.** Yesterday the write half closed: a fence now refuses a page-scoped fix
that would rewrite a template many pages share, it is live, and it has been seen refusing. One page
had already been emptied by the incident and was restored from the platform's own archive.

**What we've done since.** Asked why nothing stopped that page being emptied, given that two safety
checks sit on the exact write path it came through. Both passed, and the reason is measurable: the
text check counts stylesheet and script source as "text", so a replacement that deleted the whole
article and added a bigger stylesheet measured as 262% of the original. Visible text had gone from
2,143 characters to 16. That check now measures what a reader sees — using the extractor the
estate already had for its claims scans, after the review council pointed out I had written a new
one. Reusing it forced a re-measurement, which improved the evidence: across all 117 recorded edits
the corrected check refuses three, and two of them are a *different* lane's hollowed-out tool
pages, found from the other end. The old measure refuses exactly one thing: my own repair.

**Where we are now.** Both halves of 285 are closed and council-approved (two rounds, both APPROVED
first time, five advisory objections acted on rather than argued). The corrected check is committed
and starts working at the next build. The equivalent blindness on the whole-page rebuild path is
filed as bug 293 — deliberately not fixed, because the archive cannot reconstruct before/after
pairs for rebuilds and changing a safety rule on evidence that does not cover the path is how
guards start refusing good work. It carries a priority note: it is the higher-volume path (3,603
writes against 281 in the same eight days).

**Where we're going.** Nothing here needs an owner decision. The imported tools are being rebuilt
natively by their own lane, including the rich apps now that the owner has ruled on them; 293 needs
either a way to pair rebuild writes or a week of shadow-mode logging, and whoever takes it should
re-run the calibration harness first — it is in this lane's NOTES.
