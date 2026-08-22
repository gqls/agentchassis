# README — where we are: the release coverage gate (`bugs_open/318`)

Plain-prose log for the owner. Append; never rewrite or reorder.

---

## 2026-08-22, morning — what I was asked, and what I found in the first hour

I was pointed at `bugs_open/153` and told to fall back to `bugs_open/318` if 153 was
already being worked. 153 is being worked — a lane has owned it since 10 August, its
main fix is live, and what is left there is a short residual list rather than an open
defect. So this is 318.

**What 318 is about, in one sentence.** The release has a gate whose job is to shout
when one of our services is going to be left behind, and the gate cannot see the very
services most likely to be left behind, because it only looks at services somebody
already remembered to add to a list.

That is not a figure of speech. The check walks the deployment folder, and for each
service it asks "is the image this service uses one that the release builds?" — and if
the answer is no, it *skips the service entirely*. But "the release does not build it"
is exactly what being left behind means. So the one case it exists to catch is the one
case it steps over, and it prints "Release coverage OK" while doing it. Eight services
have fallen into this hole so far. Six were found and fixed on 18 August. Two more fell
in on 21 and 22 August — after the ruling that was meant to close it, by sessions that
had that ruling in front of them.

**And then I found something live and rather more urgent.**

There is a companion rule that nobody polices: the list of images a release *ships*
must match the list of images a release *builds*. The bug file records that this was
checked by hand on 18 August and found to agree. Four days later it does not agree. The
two new check services created this week — and one more — were added to the ship list
and not to the build list.

The consequence is not subtle. When you next run `make release`, it will build
twenty-two images, take about six minutes doing it, and then stop dead the first time
it tries to upload an image that was never built. It stops *before* the deploy step, so
**nothing reaches the cluster at all**. I confirmed this at the artefact rather than by
reading the makefile: one of the three services has an image only at `v1.0.1321` and
the current tag is `v1.0.1324`; another has never had an image built, at any tag, ever.
Two services that *are* in the build list carry an image at every recent tag, which is
what tells me the difference is real and not an accident of how I looked.

There is also a service — `capped-schedule-ordering-check` — that has a folder, build
commands, and membership of the release lists, and no job running in the cluster at
all. It was scaffolded and never switched on.

**One thing I expected to be able to do, and cannot.** The owner's stated intent on
18 August was to make the gate fire "when a service's pinned version is older than the
last change to that service's own code". The natural way to work out when a service's
version was last moved is to look at the history of the file that records it. That does
not work here: the release edits those files directly and nobody commits the result.
Twenty-six of them are sitting uncommitted right now, and the chassis's file says
`v1.0.1239` in git while actually saying `v1.0.1323` on disk — eighty-four versions
apart. So the honest way to ask "what code is this thing actually running" is to ask
the running thing, which the platform can already do since the build-provenance work in
August. That is a re-aim of the owner's intent rather than a rejection of it, and I will
put it to him as such rather than quietly substituting my own.

Next: a design plan, then the council, then the fix — narrow first for the thing that
breaks the next release, then the structural change that stops the hole reopening.

---

## 2026-08-22, afternoon — the gate is built, reviewed and in place, and there is one decision I need from you

**The urgent thing first, and it is fixed.** The next `make release` would have failed.
Three of the newest check services had been added to the list of images the release
*ships* and to none of the lists that say how to *build* them, so the release would have
built twenty-two images, taken about six minutes doing it, and then stopped dead trying
to upload an image that had never been built — before the deploy step, so nothing would
have reached the cluster.

I did not fix that by adding three names, because that is the same shape as the bug: two
hand-written lists that have to agree, with nothing keeping them in agreement. The build
list is now **derived** from the ship list. There is one list, and the other cannot drift
from it. I proved it reacts by putting a made-up service into a *copy* of the file and
watching the build refuse; the real file is never edited for a test, because a session
did exactly that this morning and another session committed their half-finished edit
before they could put it back.

**Then the actual bug.** The gate's job is to shout when one of our services is going to
be left behind. It asked "is this service's image one the release builds?" and skipped
the service when the answer was no — which is what being left behind *means*. So the one
case it existed to catch was the one case it stepped over. That is now inverted: an image
of ours that no release builds is a failure, and the only way out is to say so explicitly
in a new list, naming what does move that service instead.

The code moved out of the makefile into Go. That is not tidiness. Our review council only
looks at code in certain directories, and the makefile is not one of them — so the old
gate was never reviewed by anyone, and could not have been. It also means the thing can
now be tested properly, which the shell version could not be without editing a file forty
sessions share.

**The council approved it, first time round**, with three advisory notes. Two of them
were good enough to change the code rather than argue with:

- One reviewer noticed that my new scanner read only the *first* image in a file and
  silently ignored any second one — which is precisely the disease this whole change
  exists to cure, reproduced in miniature inside the cure. I checked: no file on the
  estate has two images today, so it was a trap waiting rather than a live fault. I
  removed the limit rather than adding a warning about it.
- Another pointed out an inconsistency in which lists I insisted must exist. Fixed, with
  a test that pins the reason it is now safe.

**Something rather satisfying happened while I was writing it up.** Another team created
a brand-new check service, and did it correctly — and the new gate said so, silently and
without being asked. That is a pass that could have failed, which is the only kind worth
much.

**The decision I need from you.** On 18 August you said the gate should fire "when a
service's version is older than the last change to that service's own code". I tried to
build that and could not, for a reason worth knowing: the file that records a service's
version is edited by the release itself and never committed, so its history is a work of
fiction — twenty-six of them are sitting uncommitted right now, and the chassis reads
"v1.0.1239" in the record while actually running v1.0.1323. Eighty-four versions apart.

The honest version of your rule asks the running thing what code it is made of, which the
platform has been able to answer since the build-provenance work earlier this month. But
there is a second fact that changes the shape of the question: now that every service of
ours is either in the release or explicitly excused, that rule could only ever apply to
the excused ones — and there is exactly one of those. So I would rather put it to you than
quietly build a different thing than you asked for. It is your call whether that is worth
building at all yet.

**One more thing to know before your next release.** Use a version number of v1.0.1325 or
higher, not v1.0.1324. Someone hand-built one image at 1324 already, and re-using a
version number serves the old cached copy rather than the new one. Also: the next release
will *create* a check service that has been sitting scaffolded but switched off
(`capped-schedule-ordering-check`) — that is fine and intended, but it will appear in the
cluster and I would rather you heard it from me than found it.
