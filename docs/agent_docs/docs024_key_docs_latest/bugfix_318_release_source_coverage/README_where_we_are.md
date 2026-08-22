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
