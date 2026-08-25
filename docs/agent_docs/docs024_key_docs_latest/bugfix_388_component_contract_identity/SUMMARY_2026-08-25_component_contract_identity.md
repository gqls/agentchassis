# SUMMARY — 2026-08-25 — bug 388: the component writer and the component store, made to agree

Written to be read aloud. Five parts, in order.

---

## What we're trying to do

When one of our sites needs a section it hasn't got a design for — a pricing block, a hero banner, a
calculator — the platform asks an AI to write one, and then files it in a shared library that every
site draws on. Because the library is shared, rewriting a component is delicate: dozens of pages have
their content stored under that component's field names, and renaming a field strands all of it.

So the platform does two things before it lets a rewrite through. Before the AI writes anything, it
looks up the existing component and says *these are the field names you must keep*. After the AI
finishes, it checks the result against the component it's about to overwrite and refuses if anything
was dropped.

**Bug 388 is that those two steps were not always talking about the same component.**

## Where we've come from

The bug was filed the day before by another workstream, which found it sideways while answering a
review question about something else. It was filed carefully — it measured the problem, and it
flagged the one part of its own reasoning it hadn't verified.

That flag turned out to be the important part. The filing said the store works out which component
to overwrite by transforming the section's name. In fact it uses **whatever name the AI wrote in its
own output**, and three days earlier a line had been added to the AI's instructions telling it to
copy the advised name across. So the two halves were joined after all.

**Joined by a sentence in a prompt.** Nothing checked that the AI complied. Nothing recorded it when
it didn't. And the sentence only appeared when there were field names to show, so a component with an
empty schema got picked and never named to the writer at all.

The bug also turned out to have a history. Buried in our concept register is an entry from an earlier
investigation that describes this exact defect, marks it "partial", and says the store-side half was
*"flagged as follow-up, not built"*. It even lists two checks someone should run to see whether it had
started causing damage. **Nobody had run them. Both came back bad.**

## What we've done

Diagnosed it properly, then built the fix, then had it reviewed twice.

The decisive finding was that a component's *name* doesn't identify it. The store's lookup filters on
the name and nothing else — not whether the component is live, not what kind of component it is — and
takes the most recently touched match. **Twenty-five names in the library are held by more than one
component**; two are held by five each. So the answer could change with nobody changing any code.

The fix: the step that looks up the contract now hands the store the component's **row id**, and the
store honours it instead of re-deriving anything from the AI's output. An AI no longer decides which
shared component gets overwritten. It also fixes a second, quieter defect found on the way — the
lookup step wasn't checking whether other sites depend on the component, so it could advise a contract
the store was about to steer away from.

Alongside the fix, four new warnings so the remaining risk is counted rather than guessed at. That
matters because the old evidence was eleven observations with no failures, which sounds reassuring and
statistically rules out almost nothing.

The review board rejected the first submission on the grounds that a function I was reusing writes to
the database. It doesn't — it builds a name — and one command over the function's body settled it. I
answered with the evidence and a test that would fail if it ever stopped being true, and the second
round approved. The board also spotted something I'd genuinely missed: I'd instrumented the case where
a stray duplicate gets created and not the case where an existing component gets overwritten by a
guess. The second is worse. That's now covered too.

**Two mistakes of mine are on the record, both caught before they mattered.** I attributed eight
historical failures to this bug on the strength of a census taken today, for events that happened last
week — the component that made them look like this bug was created four days *after* they failed. And
I wrote a test to prove a warning stays quiet that could not have failed, because the code that records
warnings forgives its own errors. I found that one by deliberately breaking the rule and expecting red;
I got green. Both are written up where the next person will hit them.

## Where we are now

The code is committed. The database change that switches it on is written and reviewed but **not
applied** — that's a live change to how the fleet builds components, and it's the owner's call. The
code itself does nothing until the next chassis build.

The honest severity: **this bug has caused no measurable damage that we can attribute to it.** The
failure class it belongs to stops dead on 19 August, when two other fixes closed the routes that were
producing it. What it is, is a door standing open on 27 of our 120 section types — 15 of which would
fail loudly and blame the wrong thing, and 12 of which would silently produce a duplicate — and a door
that widens by itself every time a component is named to a slightly different convention.

## Where we're going

Three things, in order.

**Apply the migration, or decide not to.** It's a one-line decision and safe in either order relative
to the code roll.

**Then prove it on a real case, not a clean sweep.** Rebuild one component whose name and section label
disagree, and check the rewrite landed where the writer was told it would, that no second copy appeared,
and that the new warning recorded what the AI *would* have chosen instead. A report of zero problems on
a system nobody has exercised is not evidence of anything, and this estate has been caught by that
before.

**Then watch the four warnings.** They're designed so that a row in the log always means something
real. If the "no identity was supplied" ones keep firing, it means some caller is bypassing the advice
step in practice, and the pin should stop being optional. That would be the next round — and we'd know
it from data rather than from argument, which is the whole point of having added them.
