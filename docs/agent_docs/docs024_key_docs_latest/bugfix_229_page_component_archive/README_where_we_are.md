# README — where we are (bug 229: the same safety net, for page content)

## 2026-08-09, evening — you chose option 1; the net is half-live already

You ruled that ordinary page content should get the same protection we just
built for headers and footers, done as an extension of that work rather than
a separate invention. Done as follows.

The recording half is LIVE now: any change or deletion that would destroy a
page section's stored HTML keeps a byte-exact copy first, no matter who does
it — the pipeline, an admin screen, or someone typing into the database. This
matters more for pages than it did for chrome because pages are mostly
rebuilt by delete-and-replace: the numbers show deletions outnumber edits
four to one, and delete-and-replace is exactly how the two recorded losses of
interactive tools happened. One honest gap, forced by how the tables
reference each other: deliberately deleting a whole page keeps no copies —
only sections destroyed while their page lives are archived.

The alarm half is written and tested but waits for the next software release:
rebuilds will fingerprint what they write, and the two rebuild paths with a
track record of destroying hand-made work will raise a review ticket naming
exactly what they erased. Ported-in pages are deliberately left out of
fingerprinting — their content can't be regenerated, so pretending it could
be would defeat the alarm.

A wrinkle worth telling: our own verification caught a bug in ITSELF on the
first run — the database applies one timestamp to everything in a
transaction, so "the newest row" meant nothing and the check was reading the
wrong row. Fixed by checking rows by their content. The safety change itself
was fine; the checker was wrong; this is why we make checkers that can fail.

The reviewers are looking at it now. Next: their verdict, the release, and
the same fire-drill we ran for chrome, one table over.
