# SUMMARY 2026-09-03 — the prompt was the bug

## What we're trying to do

Stop a class of failure where a page is planned, approved, and then never appears — while
other live pages go on linking to it. Six sites had pages in that state, some for weeks.
The immediate goal was to find out why 119 builds in a fortnight failed identically, and
to close the door on it rather than patch the symptom.

## Where we've come from

The failures all carried the same message: the writer had produced a sentence where the
page design called for a list of decision points. That reads as an unreliable AI writer,
and the bug was filed with two candidate explanations along those lines — either the
design had changed underneath the writer, or the writer's prompt had never learned the
shape.

There was good reason to expect the writer to be at fault. An earlier bug in the same
family had already been split into a "renderer half" and a "writer-output half", and the
lane holding the writer half had ruled, on solid evidence, that you cannot fix this class
by instruction — it needs a mechanical check plus a repair step. That check had been built
and armed, and it was doing its job: it caught all 119.

## What we've done

We read the instruction the model was actually sent, rather than the design we assumed it
was sent. They were not the same thing, and that is the whole finding.

The prompt does not contain the page design. It contains a worked example, generated from
that design, for the writer to copy. The generator only ever carried field *names*, not
their shapes — so where the design said "a list of outcomes, each with a label and a body",
the example we sent said `"branches": "..."`, which is a sentence. **The prompt declared
the wrong type.** The writer copied the demonstration it was given, and our own check
correctly refused the result. That is why the failure was perfectly consistent: we were
asking for the wrong thing perfectly consistently. The evidence is one database row
holding the instruction and the obedient reply side by side.

We fixed the generator: it now renders the real nested shape and carries the designer's own
explanatory note, which the old projection also discarded. It sits in the same package as
the check that judges the result, so the prompt now promises exactly what the check
enforces — and a test asserts that by feeding the demonstrated shape to the real checker.
Exactly one component on the estate is affected today, measured across all three ways
components can describe their fields; every other prompt is unchanged byte for byte, and a
test proves that rather than a claim asserting it.

The database half is applied and verified by reading the live row back. The code half is
committed and takes effect at the next chassis release. The two are safe in either order,
proven by rendering both states through the real template engine. The live contract is
declared to the daily drift auditor, with the old broken spelling marked forbidden, so a
revert is caught by a scheduled check rather than by six sites failing again.

## Where we are now

The fix is in and inert on the code side, live on the config side, and with the reviewer
council. It stops new occurrences and it rebuilds nothing. Most of the already-stuck pages
will heal themselves once the code ships, because the routine sweep re-mints work for a
page whose item was marked failed; the exception is a page branded "unresolved after two
attempts", a state deliberately kept open, which therefore blocks re-minting until a person
closes it. Another lane is holding one site ready as a live test and has been told which
situation it is in and to wait for the release.

One thing went wrong on the way and is worth stating because the lesson generalises: the
safety guard I wrote onto the database change had its arithmetic wrong and would have
refused my own correct edit. Rehearsing the change twice before running it — once as
arithmetic, once as the real commands in a transaction thrown away afterwards — caught it,
and also proved the undo script restores the original exactly. When you assert what an edit
will change, you have to count what it removes as well as what it adds.

## Where we're going

Two things remain open, and the second is the more valuable:

1. **Nothing repairs a build that already failed this way.** It fails, retries identically,
   and is marked terminal. There is no path that regenerates the one bad field with the
   error in hand. A design for a narrow repair agent already exists in another lane,
   unbuilt and unclaimed; if anyone takes it, these should be one piece of work.
2. **Nothing notices an active page that other live pages link to and that has never been
   built.** That silence is why these sat for weeks — the failure was loud in the queue and
   completely invisible everywhere a person would look.

After the next release: confirm at the artefact, not the status — the prompt shows the
nested shape, and a previously failing page builds with real structured content stored.
The one measurement to watch afterwards is over-production: demonstrations govern, so a
writer now shown a filled example may fill it more often than the source material warrants.
The instruction says it may be left out; whether that holds is a question for a census in a
few days, not an assumption today.
