# SUMMARY — 2026-08-07 · both checks standing, both green, nothing owed

*Fourth in the series. Predecessors: `SUMMARY_2026-08-02_…closed.md`,
`SUMMARY_2026-08-03_…complete.md`, `SUMMARY_2026-08-04_the_element_level_check_exists.md`.
Each is what we believed at the time; none has been edited.*

---

## What we're trying to do

Stop the platform from publishing things nobody told it. That started as a narrow
problem about invented phone numbers and opening hours, and turned into a broader one:
a page should never quietly assert a fact we were not given, and it should never quietly
render an empty space where a fact was supposed to be. Those are two different failures
and they need two different checks — one that reads the recipe, and one that reads the
finished dish. Neither can see what the other sees.

## Where we've come from

A component called `contact-info` had, from the day it was written, a line saying
"if we have a phone number show it, otherwise show +1234567890". The same for opening
hours: "Monday – Friday, 9am – 6pm". Nobody chose those; they were placeholders that
became live content. Eight commercial sites served the invented hours, and one served
the fake telephone number as a clickable link.

The component's own specification already said those fields must be *skipped* when we
have no value. The specification was right and nothing enforced it. That gap is the
whole story: a rule written down in one place and read by nothing.

So we fixed the component, then closed the door behind it — the platform now refuses,
at the moment of writing, a component template that substitutes a business fact for a
missing one. Then we built a daily check that reads the whole component library and
reports any new instance, because a rule enforced only at one door is a rule that leaks
through the other nine.

Then we found the blind spot. A template can satisfy that check and still be wrong: wrap
the field in "if we have it, show it", inside a table cell that is always drawn, and the
check goes quiet while the page renders an empty cell. Same hole, no warning. That is
what the second check exists for — it renders each component twice, once with data and
once without, and reports the empty elements that appear. It reads the output, so no
amount of cleverness in the recipe can hide from it.

That is where the last summary left things, three days ago: the second check existed and
was proven, but it had never once run on its own, and one decision was sitting with you.

## What we've done since

**Your decision landed and is live.** The older daily check now *fails* rather than
merely reporting when a component ignores its own specification. It was safe to do
because the backlog it would have complained about had already been cleared to zero — so
it starts green and only ever speaks up about something that arrives from now on. It has
run green every morning since.

**The new check woke up on its own — and its first act was to catch its own fault.**
It fired unattended at 06:55 and reported thirteen problems. All thirteen were in a
component created at 1:19 that morning, which was a character-for-character copy of one
we have had since February, whose same thirteen problems were already on the accepted
list. They weren't new faults; they were old faults under a new name, because the check
recognised a problem by which component it sat in. Copies like that are routine here —
three were made in a single week — so left alone the check would have gone red roughly
weekly over things nobody could act on, and everybody would have learned to ignore it.
That is precisely the failure the design existed to prevent, so the check itself was at
fault. You ruled that an identical copy is the same problem, and that is now how it
works: a copy inherits the original's accepted problems, while a copy that is later
*edited* stops matching and gets judged on its own.

**It was also reporting a number without saying what it was a number of.** It said
"139 components analysed" and never "of how many". I chased that as a possible blind
spot and it turned out to be nothing — of 184 components, exactly 139 contain anything
the check could test, and the other 45 are tools whose content the browser fills in. But
a count with no denominator is how a library quietly drifts away from what covers it, so
it now says "139 of 184, and 45 have nothing to probe".

**The repeated manual verification became a tool.** Every time the platform is rebuilt —
several times a day — the proof that our fix is in the running system expires, because a
proof belongs to the version it was taken on. We were re-running the same four checks by
hand across two machines; that had happened four times in three days, and that is exactly
the kind of repetition that gets quietly shortened on a busy day. It is now a single
command, it checks every machine rather than one, it refuses to accept a badly-chosen
probe, and it distinguishes "I could not look" from "I looked and it is not there". It is
registered where other teams will find it.

**And one correction.** A commit message of mine claimed a discipline that the commit did
not actually follow — another session had swept part of my change into theirs a few
minutes earlier. Nothing was lost and nothing broke, but the record was wrong, and in the
one place whose job is to enforce that discipline. It is written up with the near-free
check that catches it.

## Where we are now

Both checks run every morning without anyone touching them, and both passed today:
the recipe-level check clean across 184 components, the output-level check reporting no
new problems, no disappeared ones, and nothing it failed to examine. Both verified by
reading what the machine actually returned rather than trusting that it ran.

There is nothing outstanding on this piece of work. The fix is live, the door is closed
at the write path, both checks are standing, and the traps we fell into on the way are
written down where the next person will meet them rather than in anyone's memory.

## Where we're going

The honest remaining item is the size of what we already accept. The output-level check
records about a thousand places where a component *could* render an empty space, and it
is deliberately quiet about all of them — it only speaks when the number grows. That
list is not noise; it is a real inventory of small holes, mostly in tool components,
that predate any of this work. We have stopped the problem getting worse and we have not
yet made it better. Whether to work through that inventory, and in what order, is a
decision worth taking deliberately rather than drifting into — it is a body of work, not
a tidy-up.

The other thing worth watching is whether the verification tool gets used by anyone
outside this lane. It was built from one lane's repetition, and a tool that only its
author uses is a tool that quietly rots. Its register entry says so in as many words, so
the question is on the record rather than left to be noticed later.
