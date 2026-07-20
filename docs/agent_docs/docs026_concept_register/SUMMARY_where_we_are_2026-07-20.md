# The council — where we are, 20 July 2026

*The read-aloud version, following the 2026-07-16/17/17b/18 series. Current
state only; the story of how is in `README_where_we_are.md` and the notes.*

---

## What we're trying to do

This platform fixes and improves itself: it diagnoses its own bugs with
evidence a human can audit, drafts constrained fixes, and puts every fix — and
now every direction-setting decision — in front of a council of reviewers
before anything real happens. We set out to grow that council from two seats
into a full bench chosen from evidence, and then to point it at the thing that
matters most: keeping the platform on its owner's fixed direction.

## Where we've come from

A fortnight ago the review panel had two members. We distilled sixteen hundred
verified lessons from four thousand documents, used them to choose and charter
each new reviewer, and installed them one at a time — each grounded in real
incidents, each applied surgically to the live system with a backup, because
other teams edit the same machinery daily. Along the way the council proved
itself on a real case end to end: a genuine bug diagnosed, a fix drafted,
reviewed through three revision rounds — the council actually *improved* the
fix, spotting a second affected component — approved, machine-implemented,
caught by the build gate on a formatting slip, hand-finished, and finally
shipped in tonight's release.

## What we've done (this weekend)

Three things. First, we finished the bench: sixteen reviewers, eleven of them
specialists woken only when a change enters their territory. Second — the
owner's direction, made enforceable. The platform always *had* a written
constitution ("fix the cause, not the symptom; reuse before recreating") and a
written mission ("build the best possible site each domain can carry; the
revenue model shapes the site"), but nothing enforced either. Now three
always-on reviewers hold every change to them: a constitution seat, a mission
seat, and a librarian who fact-checks any claim that something "doesn't exist
and needs building" — because we watched plans sail past every reviewer on
exactly that false premise. Third, the direction is now physically guarded:
the constitution and mission files have one blessed copy each, a fingerprint
ledger, a commit gate that blocks any edit lacking the owner's recorded
sign-off, and a checker that watches every place the words live — including
the reviewers' own instructions in the database.

## Where we are now

Everything above is live in production and survived tonight's deploy intact.
The site-building brain now reads the mission as part of its brief, and a new
observe-only reviewer judges each of its decisions against it — it can't block
anything yet; it files findings to a log with a weekly report, deliberately on
probation until the numbers show it finds real drift rather than crying wolf.
Tonight's release also closed the platform's oldest silent failure: a model
response cut off mid-sentence can no longer masquerade as a success anywhere
in the system, including inside the council itself.

## Where we're going

Four decisions ahead, all the owner's: whether the mission reviewer graduates
from observing to enforcing, once a week of findings is in; extending the
direction seats to the two remaining councils (feature design and experience
planning); a fleet audit of already-built sites against the mission; and the
constitution's long-promised move from a file into proper database rows. And
one new build waiting on a green light: the multi-model gauntlet — a panel of
different AI models set loose on bugs our own loop gives up on — whose one
technical prerequisite shipped tonight.
