# SUMMARY — 2026-08-19 · fixed, live and verified

Current state only. Written to be read aloud.

---

## What we're trying to do

Make it impossible for a service to be running in production and be invisible to
the release tooling — not by fixing each service that turns out to be invisible,
but by removing the condition that lets one hide.

## Where we've come from

The release tooling held four hand-written lists of the services a release deploys,
and nothing checked them against what was on disk. `render-audit-adapter` was in
none of them: it sat 86 image tags behind the fleet for months, serving a binary
with a known credential leak, invisible because the obvious check looks at the
other service sharing that image, which was up to date.

That case was fixed on 10 August and the class in the days after: one declaration
per set, the lists became loops, and a gate that enumerates the filesystem and
fails a release when a service exists that no release path would move. Two things
were left over. Four old deploy rules that shadowed the generic one and skipped its
safety check — settled on 17 August. And six services the gate deliberately
ignored, because the release did not build their images at all; four of them the
estate's own daily health checks.

On 18 August we costed those six and the answer was sharper than expected. Two of
the four checks audit the estate by walking a list of every action the platform
knows about, and that list is compiled into the binary. A frozen image therefore
has a frozen *inventory*: one could see 165 of 169 registered actions, the other
161, and a skipped action produces no output at all, so a blind check and a healthy
estate file the same empty report. The owner ruled all six into the fleet release,
with a further trigger to follow.

## What we've done

Implemented the ruling the same day, and the first release after it carried
everything with no manual step.

All four check CronJobs and both GitHub runners now serve the current build and ran
this morning on it. The runners are the striking part: one had not moved since
mid-July and the other not since April, and the April one was missing `rsync` and
`ssh`. Both moved on the same roll, and the second runner — which the handoff
described as having no way to move it at all — needed no new machinery, because
both runners pin the same image and it simply needed declaring in a list we already
had.

The acceptance test passed and it was a test that could have failed: the inventory
compiled into the new build matches the platform exactly, 170 of 170, including all
four actions that were invisible on Monday. The platform added an action overnight,
169 to 170 — the exact churn that caused the freeze — and it is now covered without
anyone remembering anything.

Folding the six in had a second effect worth stating, because it is the more
general lesson. The coverage gate only polices services whose image the release
builds, so all six sat outside its predicate; no amount of tightening the gate
would have found them. The fix was to move them to the other side of it. We proved
that both ways: under the old declarations none of the six appears however hard you
probe, under the new ones a mutated list names all six.

We also stated clearly what we could not show. We tried to prove the change at the
check's own output rather than at its version number and could not: the report is
identical to the day before, because none of the newly-visible actions is of the
kind that report covers. So the evidence is that the right code is demonstrably in
the image, not that we watched behaviour change — and that distinction is recorded
rather than glossed.

## Where we are now

The bug is fixed, live and verified. Everything `bugs_open/237` describes is done,
including both owner decisions and the first release proving them.

Two items remain and neither belongs to this bug. The content-change trigger — fail
a release when a pinned image is older than the code it is built from — is the
follow-on the owner asked for when ruling, and it is what would catch a *seventh*
service without anyone remembering; it is the only thing standing between "these
six are fixed" and "this cannot recur". Separately, the five newly-added release
images build without the provenance stamp every other backend service carries, so
proving one of them moved means reasoning about the release rather than asking the
binary. Small, and now worth doing that they are release images.

One question is deliberately left unanswered: whether the render check was ever
functionally wrong. It is unknowable cheaply, moot going forward, and we have
declined to assert it in either direction.

## Where we're going

Close `bugs_open/237` and carry the two residuals as their own items, so the
trigger does not evaporate when the bug does. Build the trigger as its own change
with its own review — noting that whether it is *reviewable* depends on where it
lands: a makefile-only implementation is refused by the council gate on scope, as
everything in this lane was, whereas the service→sources map it needs may well
want a small Go helper, which would be in scope. **That is a design question, not
a settled fact — do not plan on the review being available until the shape is
decided.**
