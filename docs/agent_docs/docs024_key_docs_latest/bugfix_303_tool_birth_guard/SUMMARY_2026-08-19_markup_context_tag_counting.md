# SUMMARY — bugfix 303, 2026-08-19 (lane closed)

**What we're trying to do.** The platform refuses to save a generated tool that looks like a
cut-off AI generation — a good safety check with real damage behind it. But it decided "cut off"
by counting how often text like `<script` appears anywhere in the file, so a tool whose own code
merely *talks about* those tags — an HTML minifier with a comment, a regex, a snippet generator —
was refused for ever, with an error blaming a truncation that never happened. We set out to make
the check count actual markup the way a browser reads it, everywhere the platform runs it.

**Where we've come from.** The bug was filed on 18 August by the lane building HTML tools, which
could only ship by phrasing its code to dodge the counter. The same counting idiom turned out to
live in five places, not one — including the fleet sweep that files "truncated component" tickets
and the verifier that closes them, which meant false tickets that could never be resolved. Two
such false tickets were already sitting in the human review queue; one of them advised restoring
an old version over a perfectly good current one.

**What we've done.** Built one shared, browser-faithful counter and pointed all five places at it;
made the refusal message state what was actually measured instead of asserting a cause. Before
shipping, replayed old-vs-new over every component the platform has ever stored: every genuine
casualty is still caught (zero disagreements), nothing new is flagged, and the only three
components that stop being flagged were read by hand and are fine. The change went through the
review council — one revise round whose objections produced real improvements (a checked answer on
why we didn't reuse the existing HTML parser, a pod-verification recipe, first-ever tests for one
of the five call sites) — then approved. A same-file collision with another session's in-flight
fix was resolved by explicitly sequencing our commits so the shared build never broke.

**Where we are now.** Closed. The fix is live on both server replicas — verified by probing the
running binaries, since the deploy log had already rotated — and since the deploy two tools have
been born with zero false refusals. The old workaround (avoid angle brackets when describing a
tool) is retired everywhere it was written down. The two false-alarm tickets are labelled so
nobody follows their harmful advice; they clear when processed normally.

**Where we're going.** Nothing further on this lane. Two things happen elsewhere on their own:
the next HTML-manipulating tool built *without* the old defensive phrasing is the first natural
live exercise of the exact case the bug was about (the test suite already pins it), and the two
labelled tickets resolve at their normal completion. If either misbehaves, the close-out in
`bugs_closed/303` has the evidence trail to reopen from.
