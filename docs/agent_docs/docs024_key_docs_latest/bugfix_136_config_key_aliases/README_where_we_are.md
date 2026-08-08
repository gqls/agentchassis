# Where we are — config key aliases (bug 136)

Plain prose, append-only, newest at the bottom.

---

## 2026-08-08 (afternoon) — what this bug actually was, and why it got worse while nobody was looking

Someone renamed a setting. Three different parts of the system used to be told which
"pipeline" they were working on using a config key called `domain`; the code was changed to
call it `pipeline` instead, and the configuration in the database was never changed to match.
So for months, three places have been reading a key that nothing writes, and quietly using
their built-in default instead.

Nothing broke, because in all three cases the built-in default happened to be the same as
what the config was asking for. That coincidence is the whole story. The person who filed
this bug back in July saw it clearly and said so: the config "reads as a specification of
behaviour while being evidence of nothing". They also wrote down exactly what would break the
coincidence — if a check that inherits the pipeline setting were ever added to one of the
agents asking for a different value, its findings would be filed under the wrong heading.

**That is what has since happened.** Two such checks were added on the 4th of August. There
are now four items sitting in the work queue filed under "design" that should say "content",
two of them still open. It matters because the system counts outstanding work *per pipeline*
to decide whether a site is finished — so an item under the wrong heading is invisible to the
count that should find it, and inflates one that should not.

So the bug went from "correct by luck, worth fixing eventually" to "actively producing wrong
data", and it did so by the exact route the original filing predicted.

## The bit I did not expect: why two of the three places never got fixed

One of the three actions *had* been patched, by hand, with three lines of code that say "if
the new key is missing, look for the old one". The bug file spotted this and asked the
obvious question — why did whoever wrote that patch it in one place and not the other two?

The answer turned out to be that the framework gave them no way to do it properly. There is
a field in the system for declaring "this old config key is an alias for that new one" — it
looks like exactly the right tool. It is not. It only works for keys whose value is a
*pointer* to data stored elsewhere, not for keys whose value is simply the setting itself.
Used on the wrong kind of key it does something worse than nothing: it silently fails to find
anything, falls back to the default, *and* switches off the warning that would otherwise have
told you the key was being ignored.

So the honest options available to those authors were "write the fallback by hand" or "leave
it". Two of them left it. That is not carelessness — it is a missing tool.

## What I have done

I have added the missing tool: a way for any action to declare that an old setting name is
still honoured, in three lines, with the old name kept working, a warning logged when it is
used, and — importantly — the config audit report now listing every place still using the old
spelling, so it is a visible migration list rather than a silence.

Then I used it on all three places, plus a fourth small thing: the audit was reporting one
key as "unknown" that the code genuinely does read, just via a shared helper. A report with a
known-false line in it is a report people stop reading, so I fixed that too.

The measurable result, checked against the live database rather than argued: the audit's
"unknown keys" list went from four entries to one. The one left is a genuinely dead key on a
part of the system several other people are working in today, so I have left it reported
rather than reaching into their files.

## What I cannot prove yet, and I would rather say so

Two of the three places I fixed belong to the improvement loop, which you stopped
deliberately during this development phase. I have not restarted it and I do not think I
should — a check that requires switching something back on that you switched off on purpose
is not worth the evidence it produces.

So what I can offer is: the shared piece of code is exercised in production by the third
action, which does run on live lanes, so after the next release it will genuinely be
running. The wiring of each individual action is proven by tests, and I broke each rule
deliberately to confirm the test actually notices — including one that catches the specific
failure of someone reverting the code while leaving the declaration in place, which is this
bug's own shape. And after the release, a one-line check against the running pod will confirm
the code is really in there.

For the two quiesced actions specifically: no live proof until the improvement loop runs
again. That is written on the bug file rather than dressed up.

## What I have not touched, on purpose

The four wrongly-filed items are still wrong — repairing them means editing another lane's
queue. There is a separate, smaller problem in the same area where two items in the
human-review queue are labelled with their own internal type name instead of a description;
that one is not an alias problem and needs a decision about whether the system should support
templated labels at all, so I have recorded it rather than guessed. And I have changed no
configuration in the database — the whole point of the fix is that the old spelling now
works, so nobody has to.

The change is committed and has gone to the review council. It will not be live until the
next chassis release.
