# SUMMARY — 2026-08-11 — the index can now answer about a declaration, not only its use sites

Second and final milestone read-out for `bugs_open/223`. The previous one
(`SUMMARY_2026-08-10_the_index_now_says_what_it_cannot_see.md`) was written when half the
work was live. Both halves are now live and proven, so this is the read-out for the finished
piece — and for the one thing it turned up on the way out.

## What we're trying to do

We have a robot that checks our own written warnings — the file of traps called `LANDMINES.md`
— against the actual code, and marks each one still-valid or stale. It answers by searching an
index of our codebase.

The point of this piece of work was narrow and important: **make that index able to say "I
cannot see this" instead of guessing.** When a search came back empty, the robot had no way to
tell "this code was deleted" from "this kind of thing was never in my index in the first
place" — so it did what anyone does with a gap, and filled it with a plausible story. Those
stories then got written into the permanent record as verdicts, where the next reader has no
way to tell a finding from a guess.

## Where we've come from

The bug was filed on 8 August by a different workstream doing the routine thing: it added two
warnings and let the robot check them. Both verdicts were wrong in the same direction, and one
of them denied the existence of the very database category it had just been written into.

The investigation found the cause was wider than the filing. The index holds Go code only, but
**284 of 288 warnings mention something that isn't Go** — a database table, a script, a
command. And within Go itself there was a second hole: the index held functions and types but
not *named values*, the constants and lookup tables a programmer writes once at the top of a
file. So when the robot was asked about one of those, it found nothing and reported that it
had "possibly been inlined or renamed". Nothing had been renamed. The answer was structurally
guaranteed before anyone asked.

## What we've done

Two pieces, each built, reviewed by the council, and shipped separately.

**The first** made the index state its own limits. Every answer now carries a census of what
the index actually contains, computed live rather than written down, so it can never drift out
of date. When a search returns nothing, the reply says which kinds of thing the index holds no
record of at all — turning a gap into a stated fact rather than something to explain away.
That went live on 10 August and was proven on the exact warning that produced the original bad
verdict.

**The second** filled the hole itself: the index now records named values, with their contents,
so a question about a declaration returns the declaration. That went live yesterday evening and
was proven this morning.

The proof is the same warning, asked the same question, three days apart, with nothing changed
but what the index can see:

> **Friday:** the pattern lists "no longer resolve as standalone symbols, **possibly inlined or
> renamed**" — flagged for a human.
>
> **Today:** all six named things "**confirmed present at expected line ranges**" — still
> valid, nothing owed.

And a detail worth more than it looks: the warning the first piece added about missing kinds
has now **switched itself off** for those kinds, because the census it reads is live. It was
built to become untrue, and it did.

## Where we are now

Finished, and checked rather than assumed. Every category of thing in the index was counted
before and after. The new ones appeared; not one of the old ones moved. That is the check that
matters, because of the way the index stores things a new entry can silently overwrite an old
one of the same name — so we also looked directly for that overlap, and there was none.

Two things are worth saying plainly, because both are the kind of mistake that looks like
success.

**One is ours.** The handoff said to expect about 1,371 new entries; we got 1,204. The gap is
not a fault — the 1,371 was wrong. It came from running the real tool, which was itself the
correction to two earlier bad guesses, but it was run without the one setting the live system
passes it. So it was a proxy for the third time, and this time it had been written down as the
pass mark for a deploy. Trusting it would have turned a perfectly healthy result into "12% of
the data has gone missing" — which is precisely the symptom we are told to stop and investigate.
It was caught by deciding "about 1,371" was too vague to be a test at all, and computing the
exact number the live system should produce *before* looking at what it did. Every category
then matched to the unit.

**The other is a gift from a neighbouring workstream.** The handoff warned this change would be
impossible to confirm in the running system, because it adds no new text to search the program
for. True when written — but somebody has since made every service announce which version of
the source it was built from as it starts. So you ask it instead of hunting for a fingerprint,
and the answer is exact. That retires a standing workaround across the whole estate, not just
here.

## Where we're going

Nothing technical remains owed on this bug. What is left is two things, neither of them ours to
decide alone.

**A question for the owner, written up as RFC 022.** The careful way we are told to extend
shared machinery — add an opt-in switch, default it to off — is also the shape our own reviewer
is required to flag as needing extra scrutiny. So following the rule and ignoring it currently
draw the same warning, which is what makes a warning stop being worth reading. Three options
are costed there with a recommendation, but it is a judgement about how the estate should be
governed.

**And one new finding, deliberately left as a finding rather than a fix.** Verifying our own
new warnings turned up the same disease one level up. The index is only ever as current as the
last *push*, and this tree commits far more often than it pushes — this morning the index was
246 commits and 88 changed files behind the working copy. Asked about something added
yesterday, the robot again explained the absence with a manufactured reason. It is the same
shape as the bug we just fixed: we taught it to say which *kinds* it cannot see, and it now
does; we never taught it to say which *commits* it has not seen, so that gap gets filled with a
story about kinds instead. The mechanism is verified first-hand and written into `LANDMINES.md`
so nobody is caught by it. What the remedy should be is a change to shared machinery, and that
belongs in a proper diagnosis round rather than tacked onto the end of an acceptance.
