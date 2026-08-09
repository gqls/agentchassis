# SUMMARY — chrome divergence guard (bug 226), 2026-08-09

*The lane's first summary. Written the morning the rehearsal passed; current
state only — the chronology lives in NOTES and README_where_we_are.*

> **CORRECTED 2026-08-09 13:10Z — the last section of this file was wrong about
> the wave, and the prose below is left standing so the error is visible.** It
> says the fingerprinting "fills in as the wave reaches them, at roughly a site
> an hour". There is **no fleet-wide wave**, no scheduled task that would run
> one, and nothing is currently pulling sites through. Measured 2.5h after this
> file was committed: 6/57 slots on 2/19 sites, and one of those two
> (`mortgagecalculator.co.uk`) was stamped by an unrelated `nav-updater` run,
> not by anything wave-shaped. The two newest detection items have sat
> **untriaged for 3h18m and 17m** against three siblings that each drained in
> under five minutes, because the new hourly discovery rotation that files them
> (shipped the same morning by the `bugfix_230_discovery_driver` lane) is
> **observe-only by design** and the drain half is `bugs_open/083`, open.
> What caught it: the owner asking whether the wave had finished. The reasoning
> error — counting arrivals and calling it throughput — is logged in
> `WRONG_CALLS.md`. Convergence is real but ambient and unscheduled; see the
> corrected close criterion 3 in `bugs_open/226`.

## What we're trying to do

Every site's header, footer and head are stored as ready-made HTML that the
platform rebuilds from templates whenever it needs to. Until this week, a
rebuild simply replaced what was there — silently. If a person had fixed
something directly in that stored copy (which is sometimes the fastest or only
way), the next routine rebuild destroyed the fix without a warning and without
keeping a copy. That happened twice on one site and went unnoticed for eight
days. The job: make it impossible for the platform to destroy hand-made
content silently — every overwrite keeps a copy, and one that destroys
hand-made work raises its hand.

## Where we've come from

The bug was filed at a reviewer's direction — an earlier fix had "re-armoured
the symptom rather than closing the mechanism". The bug's own first fix idea
turned out to be unbuildable as written (it wanted to reproduce content from a
fingerprint that can only tell you *whether* things changed, not *what* they
were), so the design became two halves. A database-side safety net: before any
overwrite of the stored HTML, from any writer — including a person typing SQL,
the very route that caused the original loss — the outgoing copy is archived,
and if the archive fails the overwrite is refused. And a rendering-side
detector: each rebuild stamps a fingerprint of what it wrote, so the next
rebuild can tell "I am replacing my own work" from "I am replacing something
someone changed by hand" — the second gets a warning and a review ticket
naming exactly what was lost and where the copy is. The change went through
three rounds of council review (two revise-and-resubmit rounds, both of which
genuinely improved it) and was approved on 2026-08-09 with only
keep-an-eye-on-this notes.

## What we've done

Everything is live and everything has been proven against the running system,
not the paperwork. The safety net went live on the evening of 2026-08-08 and
saved four pieces of real production chrome, unprompted, within hours. The
detector shipped in the next image and was verified in the running binaries
both positively and negatively. Then the full fire-drill, this morning, on the
darts demo site: we hand-edited the footer the "wrong" way, asked for a
rebuild, and watched every promised thing happen in order — the warning fired
once in the right place, the hand-edit was archived byte-for-byte before being
replaced, a review ticket appeared carrying the edit's fingerprint, and the
rebuild re-stamped cleanly. Just as important, the control run: rebuilding
with nothing touched produced no warning, no copies, no tickets — the alarm
does not cry wolf. Even our own test edit left a tidy archive of what *it*
replaced, which proves the net catches the direct-database route, the one no
code review could ever guard.

## Where we are now

Done in substance. The one open checklist item is watching the big
re-fingerprinting wave (a sibling fix's work) roll across the fleet: it has
started — three sites through, a fourth queued as this was written — with zero
errors and the fingerprint count climbing exactly as designed (3 of 57 slots
so far; the rest fill in as the wave reaches them, at roughly a site an hour).
Until a site's slots are fingerprinted, a hand-edit there is archived at its
destruction but not flagged; that window closes site by site as the wave
proceeds. Per the owner's ruling the bug file stays in `bugs_open/` even when
the watch completes. One decision remains with the owner, deliberately not
taken by this lane: whether ordinary page content deserves the same net —
two reviewers disagreed on the record, and both positions are written up in
bug 229.

## Where we're going

Let the wave finish and note it in the bug file — the checking queries are
written down and take a minute to run. After that, this protection simply
runs itself: copies on every overwrite, tickets on every destroyed hand-edit,
and a recovery path (the archive) for anything that does slip through. The
page-content question (bug 229) waits on the owner's call, and the two
council advisories are recorded where future editors of the trigger and the
work-item flow will trip over them.
