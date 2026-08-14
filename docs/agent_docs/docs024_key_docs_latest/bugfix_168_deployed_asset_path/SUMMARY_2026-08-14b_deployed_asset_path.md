# SUMMARY — 2026-08-14b — a gate has now been reached and passed, and the decision that unblocked it was the owner's

*Current state only. Chronology lives in `NOTES_deployed_asset_path.md` and `README_where_we_are.md`.
The previous read-out was `SUMMARY_2026-08-14_deployed_asset_path.md`, written this morning; this
supersedes nothing — the series is the record. It exists because that one ended on a decision that has
since been taken, and taking it changed the answer to its last two headings.*

---

## What we're trying to do

The platform raises a finding when a page says something the site cannot support — a statistic with no
source, a number its own evidence register does not carry. Those findings are deliberately for a human
to judge, because deciding whether a sentence is *true* is not a machine's job.

But findings go stale. Somebody fixes the page and the finding sits in the queue for ever, because
nothing goes back to look. So there is a daily sweep that re-examines parked findings and withdraws the
ones that genuinely no longer hold.

The whole difficulty is one asymmetry. Leaving a finding open costs a human a glance. Closing one
wrongly means the platform has quietly decided a factual complaint about a live page was settled when
it wasn't. Every piece of work on this lane has been about the same question: **what evidence is good
enough to let a machine close one of these?**

## Where we've come from

The answer was tightened three times, each time because a reviewer refused to accept the previous one:
the page must be proved to have *moved*, not merely to have stopped tripping the check; the specific
sentences the finding quoted must be proved *gone*; and the page must have been *published* since that
edit, because the finding is about a live website and not a database. Seven council rounds, six of them
"revise", and every one of those six caught something real.

Through all of it one number stayed uninformative: **no gate had ever refused anything.** Zero — and we
could not tell what the zero meant, because a gate that looked and approved and a gate that was never
asked report the same nothing.

Yesterday and this morning fixed the *measurement*. Each decision now records which of eighteen steps
decided it, with the gate steps marked. Approved by the council first time, live, and run. It showed
the answer at once: the gates were **never reached**. Every finding stopped earlier, at "the page still
carries these claims" — correctly, because the claims really were still there.

That left exactly one thing unresolved, and this morning's summary ended on it: **an instrument cannot
tell "unreachable" from "not yet reached".** Only a genuinely cleaned page could. Creating one meant
editing live site content, so it was left as the owner's call.

## What we've done

**The owner said to do it, and it is done.** One page, cleaned properly, published, and re-examined.

The page was Leopardess Consulting's case-studies page, which claimed **"75,061 orchestration state
records"** about our own platform. The register puts that number at **2,578**, under a rule that copy
may state no more than the live count — so the page overstated it about twenty-nine times. That was not
our judgement: two separate automatic checks had already flagged it independently, the claims audit and
the freshness check that watches registered facts drift away from published copy.

The correction was **one sentence deleted**. Not rewritten, not rephrased, not given a fresh number —
removed. That is the narrow act the owner's ruling of the 6th of August permits, and it is what the
check's own instructions ask a human to do with an unsupported figure: register it, or take it out. The
sentences either side stand alone, so nothing needed patching. There is an accident worth enjoying: the
next line of that same paragraph reads *"we would rather say so than let the number do work it has not
earned"*.

Then the finding was re-examined, against a prediction written down **before** the run so it could come
out wrong. It came out as predicted: the finding passed all three gates and closed.

**One mistake, caught before it did any damage, and it is the useful part of the day.** The plan was a
single edit to the page's stored content, on the belief that re-rendering rebuilds the published HTML
from it. It does not. That re-render is a straight assembly job — it glues together HTML rendered and
saved earlier — so the edit would have been ignored, the page would have republished the claim word for
word, and the next audit would have reported the same finding. That failure would have looked like the
audit being broken rather than like nothing having happened. It was found by reading the code before
firing it, rather than by anything going wrong, and it is now written up where the next person will hit
it.

## Where we are now

**Three gates that had never run have now run, and passed.** For the first time that is a query — one
finding recorded as having passed all three — rather than an argument from prose. The record states
what each gate actually checked: the quoted text gone from the component it was quoted from, the copy
edited after the finding was raised, and the page published after that edit.

**And here is the limit, which matters more than the result.** All three gates *passed*. Not one has
ever *refused*. So the standing instruction is unchanged and should be repeated rather than quietly
dropped: **nobody may describe any gate as having prevented anything.** What changed is narrower than
it sounds — the gates have gone from *never reached* to *reached once, and passed*. A pass is not proof
of a refusal, and the two look alike from a counter, which is the same trap this lane spent a week
climbing out of, one rung higher.

**A refusal was within about three minutes' reach and we missed it.** Had the sweep run in the gap
between the page being edited and the page being republished, the third gate would have refused with
"the correction is sitting unpublished" — precisely the case it was built for. The re-render closed
that gap by publishing immediately. It costs nothing to wait: it will happen on its own the first time
the daily sweep lands in the middle of somebody else's edit.

**The counts.** Of thirty findings: one closed through all three gates, seventeen still correctly
refused because the claims remain on their pages, three cannot be judged, and nine were closed before
any of this was measured. Every closure that has ever happened shows the copy genuinely moved — ten out
of ten, no exceptions.

**A side effect we caused and should own.** The daily sweep is fleet-wide, not per-page, so running it
by hand also closed four other findings that other people had genuinely fixed earlier the same day —
two tool pages, a directory page and a front-page link across three other sites. All would have closed
the next morning anyway. Nothing closed that shouldn't have. The daily schedule itself was deliberately
left alone, and it still runs at its usual time.

**One piece of luck worth recording.** The whole job needed no AI at all, which was just as well: the
fleet's AI capability went down that afternoon on a monthly spending cap and is out until the 1st of
September unless the owner raises it. The re-render is pure assembly and the re-examination is
deterministic. It is also a second, independent reason not to have reached for the content-rewriting
route — which this finding type forbids anyway, because truth decisions are human.

## Where we're going

**Nothing is blocked, and nothing needs a decision.** The question that stood open for two days has
been answered by doing it.

**The next observation is free and needs no intervention**: the first refusal. Wait for a sweep to land
between somebody's edit and their deploy. Do not manufacture one — a second live-content change to
chase an observation would be a worse trade than patience.

**What remains is tidying, all small and all recorded.** Only one of the five re-checkers names its
steps, so the others belong on the same shared machinery rather than each growing their own. A minor
point from the approving council round is still owed a test. Five older loose ends from earlier in the
lane are still listed in the handoff.

**The transferable lesson has moved on by one step, and it is worth stating in its finished form.**
This morning's read-out said a zero is not a measurement, and that instrumenting *where a process
stopped* is what separates the meanings a count collapses. That is still true, and it turned out to be
only two-thirds of the job. Instrumenting told us the gates were never reached; it could not tell us
whether they *could* be. **The third step was to construct the input that reaches them** — and no
amount of querying would ever have substituted for it. A measurement can tell you a mechanism is
silent. Only a positive case can tell you it works.
