# SUMMARY — bug 020 data-integrity fix (2026-07-21)

*Milestone read-out, written to be read aloud. First summary for this workstream.*

## What we're trying to do

Stop the platform from inventing data. When we adopt a site with an interactive
tool whose behaviour depends on a real dataset (a searchable directory, a
data-backed chart), the tool-recreation path must rebuild the tool against the
*same* data source — never fabricate a realistic-looking dataset to make the widget
appear to work. Bug 020 is the case where it did exactly that and shipped fake
veterinary practices and postcodes to a live public site.

## Where we've come from

The vetcomparison thread found it on 18 July, contained their own site (restored
the real file, rewrote the source component to fetch the real data, locked the
components), and filed a precise diagnosis in `/bugs_open/020` with two root causes
and four ranked fix candidates — but flagged the platform defect as unfixed and
fleet-wide. The imagery workstream put a hold on tool imagery until 020 is fixed.
That is where this thread picked it up.

## What we've done

Two halves, matching the case file's own ranking (structural prompt fix + the cheap
mechanical net).

- **The prompt contract is live** (a database-only change, migration 183, no
  deploy needed). The recreation instructions now carry a prominent data-integrity
  rule — never invent records, "self-contained" means the code not the data,
  preserve the original data source, show an honest empty state if you can't reach
  it — and the analysis step now records where the original tool got its data so
  the contract is explicit.

- **The mechanical gate is built, tested, and committed** (Go, so inert until the
  next platform build). It inspects the finished tool and, on the fingerprint of
  invented data, holds the tool for human review instead of publishing it. It is
  deliberately precise — it does not flag the ordinary randomness that every game
  uses; it only trips on the actual bug-020 signature (the original was data-backed
  and the rebuild dropped the fetch and manufactured a corpus). Eleven unit tests
  pin that precision. It has gone to the reviewer council, and the final wiring is
  staged to switch on the moment the new build ships.

## Where we are now

Half the fix is live and reduces the chance of invention happening at all. The
other half — the part that makes the platform genuinely *unable* to publish
invented data — is written and proven in tests but needs the next chassis image
roll before it does anything. The bug stays **OPEN** by our own bar: fixed means
fixed AND live, and the gate isn't live yet.

## Where we're going

One owner decision and three mechanical steps: roll a chassis image carrying the
new check; confirm it's in the running pod; apply the staged wiring; then verify by
re-running a recreation of a data-backed tool and confirming it is held, not
deployed (checking the rendered page, never the "complete" status). When that
lands, 020 closes and the tool-imagery hold can lift. A couple of adjacent things
we noticed but deliberately did not fold in — a machine-readable no-fabrication
site flag, and the fact that tool-recreation swallows *all* validation errors and
deploys anyway — are noted for follow-ups.
