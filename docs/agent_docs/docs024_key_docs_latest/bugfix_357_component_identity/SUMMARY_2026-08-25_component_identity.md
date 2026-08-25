# SUMMARY — 2026-08-25, component identity (`bugs_open/357`) — PROVEN, AND A BIGGER FAULT FOUND

## What we are trying to do

Stop the platform lying about what it has stored. Every page section is kept as a row saying
which component built it, next to the markup it actually serves. On twenty-two live pages
those two things disagree completely: the row says "I am the shared hero banner" and the
markup is an entire interactive tool. Nothing errors, because nothing compares them — so a
checker complains a page is missing its headline while the page serves its own headline, and
no repair dares touch the row, because the only repair available would regenerate a banner
over the tool.

The owner's ruling was for the durable answer: record which component actually produced a
section's bytes, at the moment it is produced, and only then fix the mislabelled pages.

## Where we have come from

The recording half was built and proven at volume, and its most dangerous failure — claiming
a false origin for a tool — was shown not to happen. The producer half, the part that stops
new pages being mislabelled at birth, was built, reviewed, approved and switched on across
all six pipelines that can reach it on 24 August.

Then it sat there. For a day and a half it did nothing at all, and nobody could say whether
that was because it worked or because it had never been asked. The count of pages arriving
through the route it watches was zero, so the count of pages it had fixed was also zero, and
neither number meant anything. The repair of the twenty-two was written, and deliberately
refused to run until that changed.

## What we have done

We ran two site adoptions — cv1.co.uk and lampenkap.com — to create, on purpose, the kind of
page the mechanism is for.

They produced something better than the expected answer. The route **is** reached: the tool
rebuilder ran twice and produced two complete tool pages whose every property was exactly what
the mechanism needs. And then both saves were thrown away by a *different* safety rule, two
hundred lines further down, which compares how many sections a rebuild produced against how
many the page is planned to have. A tool page is one block by nature, so it produced one. The
plans said four and three.

**The same piece of software makes both decisions, moments apart.** It decides "this page is a
tool, send it to the tool rebuilder" and it writes "this page has four sections", and those
cannot both be true. Its own analysis of one page reads "self_contained: true" while it writes
that page a three-section plan.

The owner chose to correct the two plans rather than relax the safety rule or wait for the
proper fix. Both pages then rebuilt, and **both were recorded correctly**.

## Where we are now

**The mechanism has been watched doing its job.** Two adopted rows, live and serving. The
important one is cv1.co.uk's front page: a seventeen-and-a-half-thousand-character interactive
tool, in a slot named "hero" — precisely the situation the bug was filed about — and the row
now says "these bytes are my content", proves it, and carries a proper record of where it came
from. Checked at the served page, not just in the database.

**And the fault we found on the way is the more valuable finding.** It explains the original
bug rather than sitting beside it: of the twenty-two mislabelled rows, **twenty-one** are on
pages planned with two sections or fewer — the only ones whose single-section save could ever
have got past the rule. The rule has quietly been deciding which tool pages exist at all. Plan
two and it saves, badly labelled. Plan three and it is refused outright and the page stays
empty. **Thirty-two pages across fourteen sites** have been refused this way since the end of
July and are sitting waiting for a human.

**The repair of the twenty-two has not been run.** Its main condition is now genuinely met. A
second condition — that an adopted page survives being rebuilt without losing what it gained —
is written as a sentence rather than enforced by the code, and is still untested: the test was
killed mid-flight by a chassis roll, and is running again.

## Where we are going

Finish that one test. If it passes, run the repair and verify it at the pages themselves. If
it fails, the repair should not run at all, and the carrying half needs work first.

Separately, and probably mattering more: the software that writes the contradiction is
untouched, and will keep writing it. Correcting two page plans fixed two pages. The thirty-two
refused pages, and every one that follows, need the route and the plan to stop disagreeing —
which is a change to shared machinery and belongs in front of the reviewers rather than in a
bug patch.
