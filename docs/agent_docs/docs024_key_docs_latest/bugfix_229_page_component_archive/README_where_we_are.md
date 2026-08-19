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

## 2026-08-09, late evening — released, rehearsed, done

Your release went out, and the fire-drill for pages ran exactly as the chrome
one did this morning: we saved a test page properly (its sections got their
fingerprints), broke it the wrong way on purpose (a direct database edit —
which the safety net immediately archived, unprompted), then rebuilt it. The
alarm rang once in the right place, a review ticket appeared naming exactly
which section lost what, the erased bytes are recoverable to the byte, and a
final untouched rebuild stayed silent — no false alarms. We cancelled our own
test ticket so nobody chases a drill.

One detail worth knowing for pages specifically: rebuilds here work by
delete-and-replace, so every rebuild files away copies of what it removes as
a matter of course. That is deliberate — deletion is how page content
actually gets destroyed on this platform, so deletion is exactly where the
copies matter.

So both halves of what you decided this morning now exist and are proven in
production: chrome and page content each keep a copy of everything they
destroy, and destroying hand-made work raises a ticket. The two loose
threads, both parked with named owners: reviewers would eventually like
losses from *unlisted* writers to raise tickets too (the copies already
exist; the query is written; it needs the discovery plumbing from bugs 83/230
to run it), and nothing prunes the archive yet — we sized the growth and it
is modest, but it is not zero.

## 2026-08-19 — ten days on: everything still working, the ticket retired, and the archive needs a diet

A fresh session picked this up, checked everything again from scratch before
believing the 9th-of-August close, and it all holds: the safety net is in the
running software, both database triggers are on, nothing has ever been
wrongly blocked, and the alarm has raised twenty real tickets since. On your
12th-of-August direction that finished-and-live bugs move to the closed list,
the bug file has now moved there, with today's evidence written inside it.

Two things from the routine check-up are worth your attention. First, the
archive is growing about four times faster than we projected — 63MB now
against 30MB nine days ago. Nothing is wrong; pages simply get rebuilt more
often than the estimate assumed, and every rebuild files copies. But "decide
pruning once we have real numbers" was the deal, and we have them, so a
retention proposal is going through the reviewers next: throw away only the
old copies that the machine can recreate anyway, keep every copy of hand-made
work indefinitely, and keep the record of who-destroyed-what forever either
way. Second, those twenty tickets the alarm raised are sitting unread —
that's the old "findings never reach a handler" problem (bug 83), not this
one, but you should know the queue exists.

One small human note: the copies also caught someone (one of us, via the
database prompt) polishing the webdesign.uk hero three times over two
evenings. The net archived each draft faithfully and said nothing — exactly
as designed for self-edits, and proof the "unlisted writers" thread above is
about real behaviour, not theory.
