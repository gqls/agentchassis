# SUMMARY — mortgagecalculator.co.uk: complete adoption underway; calculators rebuilt, not yet verified

**2026-08-06. The lane's first summary** — nothing before this point was a milestone;
this is one: the owner widened the brief to complete adoption, the arithmetic
safety-net he asked for is built and proven, and the first full rebuild of the
calculators has run and been honestly measured. Written to be read aloud.

## What we're trying to do

Bring mortgagecalculator.co.uk — a live, hand-built site whose product is twelve
working mortgage calculators — fully under the platform's management, without ever
breaking the live site. The owner's two standing constraints: nothing regresses in
production, and the calculators' arithmetic must be provably right before a rebuilt
version replaces a working one. His explicit sequencing: build the arithmetic checker
first, then do the rewrite — and check for prior art before building anything.

## Where we've come from

The site was adopted editable in late July: 26 pages planned, the original 29 files
mirrored into a safety repo and verified byte-by-byte at the wire. Early August proved
the build ordering (structure and styling first, then pages) with a single test guide.
On 4 August the platform rebuilt the homepage over the live original while the site
lock was held — the lock turned out to govern only the work queue, not direct
dispatches, a correction now recorded everywhere it matters. The owner chose to
restore his original homepage; it was back within a minute of the decision and is now
protected structurally rather than by anyone's memory. Three more guides were then
rebuilt cleanly on his go-ahead.

## What we've done

The prior-art search paid twice. First, the arithmetic checker already existed —
another lane built it a week earlier: a browser check that asserts the exact text of
every calculator output, plus a harness that records what the original tools answer
for deterministic inputs. Nothing new was built where the platform already had it.
Second, running that harness against our tools exposed a real bug in the checker
itself: it varied test inputs by scaling every field uniformly, so calculators that
compute ratios — yield, loan-to-value — looked "input-independent" and were wrongly
refused. We fixed the harness (a fourth, asymmetric vector), proved the fix changed
nothing for its home corpus (all eleven of their tools still match their recorded
answers exactly), and then certified and recorded the answers of all twelve of our
originals. Those recordings are the standard every rebuild must meet.

All twelve calculator recreations then ran through the platform's own pipeline to new
addresses — the originals untouched and still serving throughout, verified after every
step.

## Where we are now

Nine of the twelve rebuilt calculators are live at their new addresses. Three
recreations reported success but produced nothing — their pages 404 and that is an
open defect to chase. The number-for-number comparison ran against all nine live
rebuilds and every one diverged — but the divergence is dominated by the rebuilds
renaming nearly every element, which destroys the comparison's unit of measurement
rather than proving wrong arithmetic. One genuine worry survives the noise: the
rebuilt stamp-duty calculator reads £0 even after being driven, and must not be
trusted until that is explained. So the honest position is: zero of twelve rebuilds
are verified yet, the machinery to verify them is in place and proven, and the gap is
a naming contract the rebuilds didn't know about, not a broken method. The live site
the public sees is exactly the owner's original everywhere it matters, plus five
clean new pages. The site is locked, nothing is queued, and the homepage cannot be
rebuilt by accident.

## Where we're going

Align the nine rebuilds to the original element names (that is also what the
permanent enforcement fences need), re-run the comparison, and adopt each calculator
only when its numbers match its original exactly. Re-run the three failed
recreations. Explain or fix stamp-duty's £0. Then hand the twelve enforcement fences
to the lane that owns acceptance contracts, so the platform re-checks the arithmetic
on a schedule forever. After the tools: the remaining planned pages, and last of all —
by explicit owner decision only — the homepage.
