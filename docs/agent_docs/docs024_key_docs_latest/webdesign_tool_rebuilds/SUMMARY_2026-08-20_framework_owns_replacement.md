# SUMMARY — 2026-08-20: the framework now owns tool replacement (written by the 286/331 session; figures measured live today ~17:35Z)

**What we're trying to do.** Replace all 63 of webdesign.co.uk's imported ("ported") tools with
framework-built ones at the same web addresses — and, per the owner's ruling of 19 August, make every
fix along the way a *framework* fix, so the next site gets it for free rather than this lane doing it
by hand 63 times.

**Where we've come from.** The 18 August summary said five tools were rebuilt and the recipe was
proven but manual: read the live tool, write a brief, build, grade, retire the old copy by hand, race
the re-render. Since then the lane found that five of the first nine ported tools were measurably
broken in production without looking broken, and that the platform itself blocked parts of the job:
the generator could not build at an existing address (bug 286), could not build a tool whose name the
shared library already claimed (RFC 036), and could not ever build the same tool twice for one site
(bug 331) — so it could ship a tool but never fix one.

**What we've done.** All three platform walls are down or falling. 286 is CLOSED — fixed, live since
16 August, proven by seven successful builds onto existing pages and zero recurrences of the original
error. The library-claim fix (RFC 036 §9.3) is live in today's fleet build, which unblocks the two
tools that were parked. The "can never rebuild" gap is bug 331: its fix is written and tested — a
work item can now say "replace the existing one" and the generator regenerates the tool in place,
same identity, old version archived automatically for a one-statement revert, page never showing two
tools, no race. The reviewer council sent round one back with a correct objection — nothing stopped a
regeneration from replacing a working tool with an empty shell — so the fix now refuses any
regeneration whose visible text vanishes or collapses, using the same guard the rest of the estate
uses. Round two is under review now. Separately, six recurring quality defects were promoted into the
generator's own contract (rules 15–20), so new briefs no longer repeat them.

**Where we are now.** Measured against the live database today: 41 of the original 63 ported tools
still serve; 24 native tool components are deployed on the site. The rebuild recipe is routine; the
generator's replacement path is code-complete and inert behind a held config switch, awaiting the
council verdict and the next fleet build. Nothing about the current recipe changes until then — the
runbook's manual steps stand.

**Where we're going.** When round two is approved and the next build rolls: apply the held switch,
prove one live re-fix end-to-end at the served page, and close 331 — from then on a re-fix is one
filing instead of three hand-run database edits and a race. Then the remaining ~41 tools proceed
through the proven phases (simple ones first, the two formerly-parked tools via the new library-claim
path, the rich apps last and one at a time, each seen in a browser). The honest end state remains:
every tool replaced, each graded at the served bytes, with the framework — not this lane's hands —
carrying every rule we learned.
