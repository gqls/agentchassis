# Where we are — the render-side half of the envelope bug (199)

Plain prose, append-only, newest at the bottom.

---

**2026-08-05, morning.**

This is the follow-up to the bug we closed yesterday. Yesterday's fix stopped a particular
piece of broken data being *saved*. This one stops it being *drawn on the page*.

The broken data is easy to describe. When we ask a model for a page section we ask for its
answer as JSON. Occasionally it replies with something we can't read as JSON — it waffles, or
it gets cut off. When that happens the system wraps the raw reply in a little envelope and
passes it along, and that envelope is not content. Yesterday's work made sure an envelope can
never be written into the database column that every future rebuild of a page reads from.

What nobody had closed is the step just before that. The renderer asks "what content do I have
for this section?", and on the failure path the thing it gets handed is the envelope itself.
There is a safety check meant to catch this, but it only works for components that have
declared, in advance, which fields they require. A component that declares nothing has nothing
for the check to find missing, so the envelope sails past and the section renders blank.

The bug file was honest that nobody had measured how many components are in that unprotected
group, and said outright that the measurement should decide whether this was worth fixing at
all. So that was the first job.

**It was worth fixing.** A quarter of the live page sections on the estate — 315 out of 1212 —
sit on components the check cannot speak for. And this isn't theory: the single remaining bad
row in the database, on gaswholesalers.com, is on exactly such a component.

There is one honest caveat and I've put it everywhere the fix is mentioned. **It isn't happening
right now.** Over the last day, sixty-two page-writing runs went through this path and not one
of them produced an envelope. So the door is open and nothing is currently walking through it.
That matters practically: when we check after the next deploy whether the new guard has fired,
the answer will be "no", and that's the expected answer, not evidence the guard is broken. I've
written that warning next to every place someone might read the count.

**What we've built.** Rather than invent a new rule, the fix calls yesterday's function. Same
policy, one step earlier: if the envelope's contents can be recovered without losing a single
byte, recover them and render the real content; if they can't, refuse to render and say so
loudly. The nice consequence is that the recoverable case now produces a *working section*
where today it produces a blank one — yesterday's guard could clean up the database column, but
it could never go back and fix a page that had already been drawn wrong.

**Two things I got wrong along the way**, both written up properly.

The first was a measurement. I asked which components the historical bad rows belonged to, got
a clean answer of "67 rows, all in the unprotected group" — exactly what I wanted to hear — and
it was worthless. The table I was querying doesn't link to components the way its column name
suggests, and the link is wiped every time a page is saved. Every one of those 67 rows had no
link at all, and the way I'd written the query quietly relabelled "no link" as "unprotected
component". Two mistakes stacked, and the second one hid the first.

The second was in my own tests. The convention here is that each test names the change to the
code that ought to break it, and I ran all four rather than just claiming them. Three broke as
predicted. One didn't. It turned out my new check is redundant — the function it calls does the
same check again — so sabotaging mine changes nothing. The tempting move is to rename the test's
claim to something that does fail. I didn't; I wrote down why it passed, because the next person
to harden that file needs to know which of the two checks is actually load-bearing.

**Where it stands.** The code is written, tested, submitted to the review council, registered,
and committed. It does nothing yet — Go changes only take effect when a new image is built and
rolled out, which happens on someone else's schedule. So the ticket stays open until we can
prove the guard is live in the running pods. That's the one remaining step, and it's a check
rather than a piece of work.

One thing I should flag: another session had stopped an hour before I started, with a note
saying its next job was this exact bug. It had stopped, and you pointed me here, so I took it —
but if that session wakes up and resumes, two of us are on one ticket. Worth knowing.
