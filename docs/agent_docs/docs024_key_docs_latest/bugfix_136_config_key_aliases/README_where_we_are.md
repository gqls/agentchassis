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

---

## 2026-08-08 (evening) — it shipped, and the proof I had lined up turned out to prove nothing

A fresh chassis build went out (v1.0.1267) and the change is genuinely in it. I checked the
running binaries rather than trusting the version tag, and I checked them the corrected way —
by finding every pod running that image rather than the two that carry the obvious label, and
by including a control string that has been live for ages so that a "not found" can be told
apart from "I looked in the wrong place". Before the release the marker was absent and the
control present; after it, both present. That is a real before-and-after, not a hopeful
reading.

So the code is live. What I cannot yet tell you is whether it is *doing* anything.

**The witness I had planned does not work, and the reason is uncomfortable.** My plan was:
one of the four places I fixed runs on live lanes every day, so after the release it would
exercise the shared piece of code and I could point at that. When I went to collect it, I
found that all nine live configurations using the old key set it to `build` — which is
exactly the value the code falls back to when it finds nothing. So the result is `build`
either way. There is no reading of that run that could have told me the fix was working, and
equally none that could have told me it wasn't.

That is precisely the fault this whole bug is about. The bug exists because three pieces of
configuration were being ignored, and nobody noticed for months because the ignored value
happened to match the built-in default. I then proposed to prove the fix by watching a case
where the aliased value happens to match the default. I reproduced the bug's own blind spot
inside its own verification, which is worth writing down plainly rather than tidying away.

The one thing that would settle it is a log line the new code emits whenever it honours an
old key. I searched every pod for it and found none — but the same search also found no trace
of the action that should have produced it, which means I was searching the wrong logs, not
that the line never appeared. A zero from a search that cannot find the thing it is standing
next to is not a negative result.

**So the honest position:** the fix is committed, reviewed, approved, and demonstrably present
in the running binaries; the audit report against the live database confirms the declarations
match real configuration; and whether the aliasing actually fires at runtime is still unproven.
I have written down the two observations that would settle it — find where that action logs,
or change one configuration value to something other than the default and watch what happens.
The second is a one-line change to another lane's agent, so I have left it for whoever owns
that, rather than reaching into it.

Separately, the item still most worth your attention is unchanged and is a decision rather
than a bug: two entries in the human-review queue are labelled with an internal type name
instead of a description, because the configuration asks for a templated label and the code
that writes it does no templating. Renaming the key would put a raw `{{...}}` string in front
of a reviewer, so it needs a call on whether templated labels should exist at all.

### Correction to this morning's entry, which I should have added earlier today

The afternoon entry above says of the four wrongly-filed items: *"It matters because the
system counts outstanding work per pipeline to decide whether a site is finished — so an item
under the wrong heading is invisible to the count that should find it."* **That is wrong and
I corrected it hours later in the bug file, but not here, which is the file you actually
read.**

The count I cited is real, but it has exactly one caller and that caller always asks about
the `build` pipeline. An item filed under `design` and one filed under `content` are equally
invisible to it — it cannot tell those two apart, so it cannot be evidence that confusing
them costs anything. I checked every other consumer afterwards, in the code and in the live
configuration: they all name `build`, `reports` or `diagnose` specifically. **Nothing today
treats `design` and `content` differently.**

So the accurate version: the four items are genuinely mislabelled, and nothing currently
reads that label in a way that makes the mislabelling cost anything. The exposure is that
the labels are right-or-wrong by luck, and the first thing that ever does distinguish them
inherits whatever the default happened to write — which is the same fault as the original
bug, one level out. That is a good reason to fix it and a poor reason to have called it
urgent, and I called it urgent.

Nothing about the fix changes. What was wrong was my argument for it.
