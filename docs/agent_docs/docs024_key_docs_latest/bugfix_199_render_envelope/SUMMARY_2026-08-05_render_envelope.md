# SUMMARY — 2026-08-05 — the render-seam envelope guard (`bugs_open/199`)

First in this lane's series. Written to be read aloud.

## What we're trying to do

Stop a known piece of broken data — an "LLM transport envelope" — from reaching the point where
a web page section is drawn. The envelope is what the system produces when a model's answer
can't be read as JSON: a small wrapper carrying the raw reply. It is never content. When it
reaches a template, the template renders blanks, page assembly drops the empty section, and the
section silently vanishes from a live site.

## Where we've come from

Yesterday's lane closed the **storage** half of this. Its guard sits at the two places that
write page content into the database and either recovers the envelope losslessly or refuses the
save outright. That work is live and proven.

While doing it, that lane noticed a third path — from a model step's output straight into
render — and chose not to change it, on the grounds that altering what components render is a
different decision from what a database will store. The review council agreed the deferral was
properly declared, but one seat objected that a note in a rationale field is not a tracked item.
This bug file is that objection's disposition, and this lane is its answer.

The bug file was scrupulous about what it did not know. It marked its own central claim
`[UNMEASURED]` and said the census should decide whether this was a bug or merely a note.

## What we've done

**Measured it first.** A quarter of the estate is exposed: 315 of 1212 live page sections sit on
components whose declared schema gives the existing safety check nothing to test. The one bad
row still in the database, on gaswholesalers.com, is on exactly such a component — so this is a
mechanism that has fired, not one that might.

**Measured the other direction too, and it cuts against urgency.** Over the last day, 62
page-writing runs went through this path and none produced an envelope. The check that says so
could have come out differently — the same query finds 111 instances elsewhere in the same
window. So: the door is open and nothing is currently walking through it. That caveat is written
beside every count someone might later read, because after the next deploy the honest reading of
"zero firings" is "as expected", not "it works" and not "it's broken".

**Fixed it by reuse, not reinvention.** The new guard calls yesterday's function. Same payload,
same policy, one seam earlier: recover the content when that loses nothing, refuse the render
when it doesn't. Because rendering happens *before* saving, this also means the recoverable case
now produces a working section — the storage guard could repair a database column but could
never go back and fix a page already drawn wrong.

**Corrected the brief.** The bug file names the wrong branch of the resolver as the leak. The
real one has no content check at all, and the branch it names is a *second*, different leak. That
is why the guard sits at the caller, where one call closes both doors. The correction is now in
the bug file, the debugging guide, the concept register and the landmines file.

**Wrote down two of our own errors.** A census query that produced exactly the answer we wanted
from a join that could never have resolved; and a test whose named sabotage didn't break it,
because the check it guards is duplicated one layer down. Both are in `WRONG_CALLS.md`, and the
first is now a landmine so the next person meets it before they have a symptom.

## Where we are now

Written, tested with all four sabotage checks actually run, submitted to the council
(`dfb87f5e`), registered as the third seam of PBP-032, documented, and committed.

**It is inert.** Go changes do nothing until an image is rebuilt and rolled, so the ticket stays
open. Not because work remains — because the bar for closing is "fixed *and* live", and a defect
committed but unrolled is still reproducible in production.

## Where we're going

One step: after the next chassis roll, grep both running pods for the new symbol, read the pod
start time in the same breath, then close the ticket to `bugs_closed/`.

Beyond that, one thing stays deliberately open. This closes the *envelope* class only. A
component that declares no required fields will still render whatever else it is handed,
envelope or not. Making that check speak for schema-less components at all is a much bigger
decision — probably an architecture question rather than a bug fix — and it stays on the file
rather than being quietly folded in here.
