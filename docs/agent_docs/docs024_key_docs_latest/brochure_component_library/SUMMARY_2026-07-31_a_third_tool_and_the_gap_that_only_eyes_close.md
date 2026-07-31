# A third tool, and the gap that only eyes close

**2026-07-31, brochure component library / fundamentallyai.com.** Written to be read
aloud. Supersedes nothing: the previous read-out
(`SUMMARY_2026-07-30_the_panel_is_finished_and_two_new_fronts_open.md`) is still the
record of where we were yesterday morning, and this is the record of what happened next.

---

## What we're trying to do

Build fundamentallyai.com into a site that demonstrates the platform rather than
describing it. The company's pitch is that it builds AI systems which do a defined job,
run unsupervised, and keep a record of every decision they make. A brochure page can
claim that. A working tool built on the platform's own operating data shows it.

Underneath that, a second aim that matters more in the long run: every component we
hand-build here should teach us something about how to make the framework build the next
one. The site is the visible half; the method is the half we keep.

## Where we've come from

The site was rebuilt over the last fortnight from something unreadable into something
that reads properly. Along the way we built a carousel component (`teaser-reveal-panel`)
that took five rounds of owner feedback to get right, and the final round found a bug
that had been live since the first version: the JavaScript had never run at all, because
the script loaded in the page head before the markup it needed existed. Every check that
component had ever passed had exercised static markup or forced the browser's state by
hand. **Nothing had ever clicked anything.** That finding is the reason everything below
is shaped the way it is.

Two tools already existed on the site: an LLM cost calculator and a model-approach
selector. Both are self-contained interactive pages. Neither had been examined closely
enough for us to know how they were actually built.

## What we've done

**Built and shipped a third tool: the AI Review Council Simulator**, live at
`/tools/review-council-simulator.html`. You set which reviewers sit on a panel, how
serious an objection has to be before it blocks a change, and how many revisions you will
run; it tells you how often sound work gets through, which reviewers stop you most, and
what the review costs in rounds. Every number in it is measured from 362 real council runs
on our own platform, with the individual objection rate of all 26 reviewer seats.

The measurement did more than fill the tool in. It **killed two things before they
shipped**. A planned feature modelling how many rounds a change really takes turned out to
have no data behind it at all — every one of the 266 verdict records says "round 1", so the
distribution I intended to use does not exist. And a label reading "medium and high
objections block: this is what we run" was simply false: of 110 approvals, 99 contained a
medium objection and passed anyway. Medium is advice; only high blocks. What caught that
was not a review, it was the tool disagreeing with reality by a factor of ten.

**Then fixed the carousel on the owner's report.** The cards' text was touching their
edges, and the "Read the rest" links sat at different heights. The second was a layout
choice, quickly corrected. The first was more interesting: the padding was not too small,
it was **zero**. The component asked for spacing using variable names the theme does not
define, and when CSS asks for a name that does not exist without giving a fallback, the
browser discards the whole instruction rather than choosing something sensible. Eight
declarations in that component were dead the same way. The file looks perfectly correct;
only the browser knows.

**And built one thing worth more than any of it: a test that drives the real controls in a
real browser and has been proved able to fail.** Six deliberately broken copies of the
simulator, each caught. One of those mutants is the exact head-loading bug that cost five
rounds last week; it now fails in seconds.

## Where we are now

The tool is live and verified — 47 assertions, run against the served page, not a local
copy. The carousel is fixed on all four pages that carry it. Both changes are recorded in
the lane's log, the owner's log, two new fleet-wide landmine entries, and their commits.

**Three findings from this stretch are worth more than the deliverables**, and all three
are about the difference between checking something and knowing it.

The first: **a test that has never failed is not evidence.** The simulator's own test
reported, on its first run, that the component was completely dead — the precise signature
of last week's bug. The component was fine; the test was inspecting the page a fraction too
early. Had it been trusted, it would have caused someone to break working code. The same
thing happened again hours later on the carousel, in a different form. Both times what
saved it was running the test against the *unchanged* version as a control. A failure you
cannot reproduce on known-good code tells you nothing about the code.

The second: **agreement is not confirmation.** I measured our council approval rate,
found the figure in our own standing documentation looked wrong, and published a
correction. It was not wrong; it was counted on a different basis, and another session had
measured exactly that two days earlier and written it down. My number reproduced theirs
precisely — and I read the agreement as a discovery instead of as corroboration. Retracted,
and the tool now shows both figures, because confusing them is how an ordinary revision
starts to look like a failing plan.

The third, and the one that is genuinely a gap in the framework rather than in a person:
**nothing in the system ever looks at a page unless something has already failed.** The
framework can screenshot a page, store it and hand back a link — but only when a declared
check has failed, as evidence explaining the failure. There is no path that renders a page
and puts it in front of a human or a vision check for its own sake. Every defect in this
stretch that reached the owner — text against the card edge, links off the line,
overlapping labels on the simulator — occurred on pages where **every automated check
passed**. There was no assertion about spacing or alignment, so nothing failed, so nothing
was captured, so nobody would have looked. All three were found by a person looking.

There is also a smaller, sharper gap: **the carousel changes are not recorded in any
travelling document belonging to the carousel.** They are in the lane's log and its commit,
which the next person to touch that component may never open. The tool we built *does*
carry its own plan and notes, because the platform's document tables accept a subject of
type "tool". They still refuse type "component" — the migration that would allow it is
written but not applied — so the component that took five rounds of feedback has no
travelling record of its own, and the one built in a day does.

## Where we're going

Closing those last two gaps is the immediate work, in that order: give the carousel a
travelling record that lives with the component, and then make the framework capable of
showing a page to someone without waiting for a failure to justify it. The second is the
larger prize, because it is the difference between a system that can prove what it was
told to check and one that can be looked at.

Beyond that, the staged build ladder — the proposal that would make all of this the normal
path rather than something hand-rolled per component — remains a separate thread by the
owner's instruction. This stretch has given it three concrete pieces of evidence it did not
have: a mutation-proved interaction test, two worked examples of a test lying convincingly,
and a measured account of exactly which defect class the current tooling cannot see.
