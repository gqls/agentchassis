# Summary — the chrome pin (bug 170), 2026-08-01

## What we're trying to do

Make one rule decide which component supplies a site's header and footer, and make it
impossible for anything to get round that rule. Not for these three sites — for the
mechanism, so the next component switched off cannot quietly become somebody's header.

## Where we've come from

Over the last fortnight this one question — "which component serves this site's chrome?" —
produced three separate bugs, all now fixed and live on `v1.0.1225`. **118** found that the
question had three different implementations giving three different wrong answers, and gave
it one predicate. **166** found that the platform's repair for the resulting mess re-rendered
the broken component instead of replacing it, and made it actually repoint. **167** found
that the page-build path could pick a page-section component to serve as a site header.

Each of those concerns the same store: the per-site record of which component a slot uses.
On the last of them, the 167 lane noticed a *fourth* way in — the style collection, a shared
look-and-feel record that several sites point at, which may also name a header and footer
directly and was checked by nothing at all. They filed it as bug 170 and deliberately did
not fix it, because fixing it changes how live sites look and that felt like a decision to
put to the owner rather than take.

## What we've done

Confirmed the bug is real and found two things the filing does not say.

It **undercounts**: three sites are on a switched-off header, but **four** are on a
switched-off footer — including the one site whose header pin is correct.

And, materially, **the pin is not only read, it is written from**. The routine job that
decides which component each site's header slot should point at takes whatever the style
collection names, unchecked, and writes it into the per-site record — the record 118 and 166
just repaired. Right now every one of those sites' per-site records is correct and every
style collection is wrong, so the next run of that job would put the broken value back and
blank the stored header. A third route copies the pins into every new collection, which is
how the problem spreads. So the fix that shipped on Thursday was not durable while this
stood.

Fixed all three routes through one predicate and one dereference function, plus the
detector — which had been raising this defect class since 17 July while joining only the
guarded store, and so had never once mentioned these seven rows. Two static guards were
added, both proven by deliberately breaking them: a scan that fails if anyone writes a
fourth unguarded consumer, and a lockstep that fails if the detector's copy of the rule
drifts from the shared one.

The predicate is deliberately **not** a copy of the existing one. It omits the clause that
excludes forks, because a fork is wrong as a shared default and is the entire point of a
per-site pin. Measured across all four live pins, the two rules disagree on exactly one row,
and copying the existing one would have deleted the fleet's single correct pin while
catching the three wrong ones.

## Where we are now

Committed (`e44e6dd06`, docs in `e8f8bd504`), all tests green against a clean checkout,
submitted to the review council, and **inert until the next chassis image rolls**. The
ticket stays open until then, because a fix that hasn't shipped is still reproducible in
production — the closing condition and its verification commands are written into the ticket.

We took the decision the ticket was holding for the owner, and said so plainly rather than
quietly. The reason: when the pin is ignored, the site falls back to the very component the
earlier repair already moved those same sites to, with council approval. It makes the second
record agree with the answer we have already given, rather than deciding anything new.

Two caveats stated up front rather than found later: the four style-collection rows are
still wrong (now ignored rather than obeyed, which is what the detector now surfaces for a
human, since repointing a shared collection moves three sites at once); and no live page
changes until site chrome is next rebuilt, which is a separate open bug.

Two mistakes are on the record. Both of the guards described above passed for the wrong
reason on their first version — one couldn't distinguish a correct fix from an unguarded
one, the other matched a word that appeared elsewhere in the file for an unrelated reason.
Same error twice in one session: asking about *the file* when the question was about *the
thing*.

Separately: the automatic diagnosis service was run, as the rules now require for a claim
this structural, and produced no conclusion at all — the file at the centre of the question
is 94KB and the service can only show its reader 60KB, so the one thing it needed was
omitted every iteration. That is not specific to this bug and is now written up where the
next person will hit it.

## Where we're going

Three things, in order. **Roll and verify** — the fix does nothing until it ships, and the
ticket names exactly what to check on the pods, including the negative control that catches
the one way this fix could go wrong. **Read the council verdict** and act on it; the code is
already on the shared branch, so a REVISE is a follow-up commit, not a hold. **Then the
owner call that remains**: whether to repoint the four style-collection rows or to
reactivate the components they name. That question is now visible in the work queue instead
of invisible in the schema, which was the point.
