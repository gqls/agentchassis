# SUMMARY 2026-08-19 — the fault was never in the writer

Fifth in the series (08-12 *why the copy reads wrong* → 08-14 *the house voice ships* → 08-15
*plan complete, four decisions* → 08-17 *the editorial pass exists and showed restraint* →
this). Written because the previous entry's headline turned out to be true only of an easy
case, and because the lane now knows something it did not know two days ago about where the
problem actually lives.

## What we're trying to do

Make the copy on our sites read like a person wrote it. The machinery that builds a page is
graded on facts, coverage, structure, links and styling — never on whether the result is worth
reading, whether the most useful thing is at the top, or whether the page talks about itself
instead of to the reader. Something in the estate already noticed those faults; nothing fixed
them. Design faults had a fixer; copy faults had none.

## Where we've come from

A house voice that lived in seven drifting copies, now one row every writer reads. Two sites'
briefs corrected where they were teaching the register the owner had rejected. Then, on the
17th, the missing fixer itself: a second, editorial pass that reads a whole page and proposes
changes. It needed no new code — everything it required already existed, and the one genuinely
missing piece was a database query in configuration.

## What we've done

**The editorial pass is built, live, and has changed two live sites** — both times with the
owner's approval, and both times verified at the served page rather than at a status.

On the loan-and-mortgage homepage it restored six guide links that had been missing since the
12th, adding six lines and touching nothing else. On the AI-orchestration homepage it removed a
sales pitch masquerading as a product feature, cut a hundred-word tangent, and standardised a
name that was being written four different ways — 250 words shorter, with every number and
every link on the page still present and nothing invented.

**We also found out what it is not.** Given a page whose fault was spread thinly through eight
sections, it tried to rewrite the whole thing at once and was cut off mid-answer. So the
restraint we praised on the 17th was a property of an obvious defect, not of the design. It now
works to a limit of three changes per run, chosen by what a reader actually loses — a bounded
job cannot overrun, whereas a bigger allowance only moves the wall.

**And the checker that grades its work was wrong three times**, each found by using it on
something harder rather than by reading it, each the same fault in different clothes: it
reported "checked" for something it had not looked at. It could not tell deliberate
de-duplication from gutting a section; it silently skipped list-type content; and it could not
see a web address written as ordinary prose. All three are fixed, and each fix was followed by
proving the checker can still catch real damage — because "it complained, so I changed it" is
how a check quietly becomes decorative.

## Where we are now

**The important thing we learned this week is that the fault was never in the writer.**

The owner reported three directory pages whose copy "looks like it didn't go through the
framework", quoting sentences built on a mannerism — saying what a thing *isn't* in order to
say what it is. We spent a day looking for the cause in the writing machinery and the house
voice. It was in neither.

**It is in the brief.** The instructions handed to the writer for that site are themselves
written in that mannerism, seven times over — and one of them supplies, word for word, the
sentence the owner objected to. That phrase reaches the writer in 1,348 prompts and comes back
in 408 pieces of copy. The machine is doing as it is told.

**And it is not one site.** Twenty-four of twenty-five briefs across the estate are written the
same way, most of them far more heavily than the site that prompted the complaint — which turns
out to be among the mildest. The worst is the pilot page of another team, who had offered to
rerun it as soon as a fix existed; they have been warned that rerunning it against that brief
would reproduce the fault from the worst source we have, and read as the fix failing.

**We were wrong in public twice this week and corrected both promptly.** A before-and-after
measurement I published — and sent to two other teams — reversed its conclusion the moment I
controlled it properly, and the underlying figure turns out too noisy to answer the question at
all. It was withdrawn within the hour, in five places, naming the sentence each team should stop
repeating. A diagnosis run we commissioned came back "not confirmed" and refuted our leading
theory, which is the cheapest possible place to be wrong.

## Where we're going

**First, a measurement designed so it can come out either way** — we know the supplied phrases
travel from brief to page, because we can trace one word for word. We do not know whether the
*instructional* parts of a brief shape the writing in the same way, and after this week's
retraction that question gets a test with its refutation condition written down before we look.

**Then a check on the briefs themselves**, flagging the construction where a brief hands the
writer a phrase to reuse, and merely reporting it where a brief is giving guidance. That is the
owner's "make sure this never leaves the framework again", aimed at the place that causes it —
before a single page is written, rather than after twenty-five sites have been.

**Then the briefs get fixed** — which is site configuration belonging to other teams, so it is
a conversation to have rather than an edit to make.

The editorial pass stays what it has proved to be: good at what a section-by-section writer
structurally cannot see — the same argument made five times on one page, one thing under four
names — and not the answer to a fault that was written into the instructions before any page
existed.
