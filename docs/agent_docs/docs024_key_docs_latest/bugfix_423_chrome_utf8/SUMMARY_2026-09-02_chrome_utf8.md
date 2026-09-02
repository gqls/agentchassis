# SUMMARY 2026-09-02 — bugs_open/423: a byte-sliced em-dash, and ten days of silence

## What we're trying to do

Make the site chrome pipeline unable to lose a footer quietly. Two things had to be true:
the specific corruption had to stop, and a database refusal had to stop being reported as
success. The second matters more than the first.

## Where we've come from

The bug was filed on 31 August by the delivery lane, working on boxingonline.com. They did
the expensive part: they caught the mechanism **live**, with a repro dispatch and a log
monitor on every chassis pod, and saw Postgres refuse the footer with
`invalid byte sequence for encoding "UTF8": 0x80` while the step went on to report success.
They split the bug into two halves — an observability defect they had settled at the code,
and a data defect whose source they could not find — and left two "graders": tests any
proposed cause had to beat, including one that killed their own best hypothesis. Those
graders are why this session did not re-derive a dead end.

## What we've done

Found the cutter, and it was not where the bug file was looking. `buildServicesHTML` builds
the footer's services column and title-cases each word with `strings.ToUpper(w[:1]) + w[1:]`
— a **byte** slice. `strings.Fields` makes a standalone em-dash its own word, so it cuts a
three-byte character after one byte. Running it produces `ef bf bd 80 94`: the `0x80` from
the live capture, exactly. The site's trigger is a page titled *"Boxing Quiz — Test Your
Knowledge"*.

It passed the filer's own grader — this **cuts** at a byte offset rather than merely
containing a multi-byte character — and then a two-way census settled it. Sites with such a
label in the footer query's window: two. Sites whose footer will not store: the same two.
No false positives, no false negatives.

The second site is **garden-tools.uk**, broken since **23 August** and unknown to anyone.
That is the observability defect's cost, measured rather than asserted.

Then it turned out to be a class, not a case: the same byte-sliced shortcut was hand-written
at **eight** call sites. And the estate had already fixed the *truncation* shape of this
class on 20 July, built the rune-safe primitive for it, and never gone looking for the
*casing* shape. So the fix is one shared primitive with all eight sites converted; the
failure branch wired into the reporting surface that already existed; the three byte-sliced
summaries in the reporting path itself made safe in the same pass (one of them interpolates
an arbitrary error string, so the surface that reports a chrome failure could have died of
the disease it reports); and a gate before the store that refuses invalid bytes and — unlike
Postgres — says **where** they are.

Five tests, each proven to go red when its fix is reverted.

## Where we are now

Committed, registered as STY-059, submitted to the review council. **Not live.** Go changes
are inert until an image is built and rolled, so both footers stay broken until then and
boxingonline keeps serving its hand-patch.

One accepted behaviour change, flagged rather than buried: a slot that fails to store **and
has nothing stored to serve** now fails the build. Garden Tools is exactly that case, so
this lands on a live site. We hold it is right — a site must not go live with a missing
footer — and we have asked the council to argue it rather than nod it through.

## Where we're going

The roll, then verification at the artefact: both footers stored, digests matching, the
damage census empty, and boxingonline's pre-delivery check (no contact block, because its
email is empty) still passing. After that the bug closes on the estate's bar — fixed **and**
live. One sibling branch was deliberately scoped out and named in the file, so the next
reader can see it was considered rather than missed.
