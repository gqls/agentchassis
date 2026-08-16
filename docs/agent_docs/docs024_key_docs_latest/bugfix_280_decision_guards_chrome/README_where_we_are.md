# README — where we are (bugfix 280, plain prose, newest at the bottom)

## 2026-08-16

You asked me to pick up "bug 180." I checked, and that one's already done —
fixed and confirmed working back on 2 August. What you were probably
thinking of is bug 280, a sibling issue that turned up while a different
session was fixing bug 270 a couple of days ago, and their handoff note
specifically said "you probably mean 280, not 180" — which is exactly what
happened. I asked you rather than guessing, and you confirmed 280.

Quick recap of what 280 actually is, since it's a bit dry as written in the
bug tracker: there's a feature where you (or whoever's steering a site) can
write down a decision — "this page must always link to the audience-check
tool" — and the system automatically re-checks that decision is still true
every time it scans a site, flagging it if something breaks the promise.
That re-check works by reassembling "the page as it's stored" and searching
it for the text the decision cares about. The bug is that the reassembly was
quietly leaving out the header and footer — the shared bits at the top and
bottom of every page — because it was reading from a database column that
stopped being used a while back (this is the same root mistake as bug 270,
just showing up in a different place). So any decision about header or
footer content was either always wrongly flagged as broken, or never
flagged even when it genuinely broke, depending on which way the check was
written. Nobody has actually hit this yet — none of the five decisions on
record today happen to be about header or footer content — but it was a
silent trap waiting for the first one that is.

I checked first that nobody else was already working on this — the
ownership tool said "maybe," but that was just because the session that
found it two days ago is still around working on other things; nobody has
actually touched the file since, and I didn't find any other live session
talking about the specific bit of code involved. Clear to go ahead.

The fix itself is small: point the same lookup at the right table for the
header and footer, the same fix bug 270 used for a different check. Wrote it,
tested it (including a test that proves the old, broken version would fail
the new test — so the test is actually catching the right thing, not just
passing by luck), and the whole codebase still builds cleanly.

Next: commit it, send it through the standing review process (advisory, not
a blocker), and then it's a question for you whether to build and roll it out
now or leave it for later — same as 270, where you said you'd handle the
shipping step yourself.
