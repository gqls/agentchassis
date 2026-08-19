# SUMMARY — 2026-08-19 — bugs_open/131 (og-card slug): CLOSED

*Third summary in the series (07-29, 07-29b, now 08-19). Written because the lane's state
genuinely changed: the bug is closed. Current state only; the chronology is in NOTES and
README_where_we_are.*

**What we were trying to do.** Make every site's social share preview and favicon actually
exist — and make the platform unable to believe it had done that work when it had not.

**Where we came from.** On 28 July, eleven of fourteen live sites advertised a share card that
returned 404; the generator turned out to exist and had simply never been run. July fixed the
generator's worst habits (aspect-preserving favicons, lock honouring) and repaired sites by
hand. Then the sites kept regressing: a second mechanism was hiding behind the first. Found on
17 August by another thread, fixed at source on 18 August by this one: the automatic detector
filed its "please generate" notes without the one field that routes them, so they reached a
branch whose only possible answer was a polite refusal, and that refusal was recorded as
success — twenty-one times.

**What we did.** Closed that at all three doors: the detector now includes the routing field;
the router gained a last-resort rule that sends obviously-brand-head items to the generator
rather than to a refusal; and the completion gate now checks the artefact record before
accepting "done" for this item type, so a refusal that changed nothing can never read as
success again. The change went through the review council (one revision round, which caught
two real sloppinesses of mine; then approved by all twelve seats). Six sites were repaired the
same day and checked by eye, and the routing fallback was exercised deliberately with an item
carrying the exact broken shape.

**Where we are now.** Closed and live, each claim verified at the artefact: both of today's
chassis builds carry all three commits (proven in the running binary, not the tag); the new
completion check fired on a real item and graded it correctly; a fleet census of every site
with deployed pages shows all eighteen non-loan public sites serving both files. The five
loan-family sites that still 404 are other bugs' mechanisms and are routed there: two just
needed their first-ever generation run (redriven; one already verified 200/200), three have no
logo at all, which is a logo problem before it is a card problem. Deliberately not touched:
the other bug numbered 131 (vonc gauntlet), a different case.

**Where we're going.** The remaining defects in this area — the head block being blind to
which page it is on (every inner page shares as the homepage), the share title falling back
to the bare domain, the image tag emitted whether or not the file exists, and wide logos
making unreadable tab icons — are spun out to `bugs_open/322` with their evidence re-checked
against current code. One housekeeping follow-up for whoever next touches the asset
deployer: its input contract is empty and should declare the two routing fields. This lane's
directory stays as the record; 322 is the successor work.
