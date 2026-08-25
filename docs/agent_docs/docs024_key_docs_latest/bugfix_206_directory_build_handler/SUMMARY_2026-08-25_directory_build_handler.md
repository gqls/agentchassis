# SUMMARY 2026-08-25 — finished the part we were blocked on, and found that our own proof did not prove anything

## What we're trying to do

Make sure that when the system plans a website, every page it plans gets sent to a builder that can
actually build it. Some kinds of page — a directory of things, an index listing other pages — need
a builder that knows how to give a brand-new page a starting layout. Only one of our builders can
do that. If such a page is sent to the ordinary builder instead, the build quietly does nothing,
the page never appears, and the error it leaves behind reads like a content problem on somebody
else's site.

## Where we've come from

We fixed this once, on 8 August, and believed it was done. It was not: there are two separate
routines in the system that hand out this work, and we had only taught one of them. The other had
the answer written into it as a fixed string and never asked. That was found on 24 August, after a
directory page had sat broken for fifteen days while the builder that could have built it was
running the whole time.

The fix for that was approved on 24 August after six rounds of review, and is live. But it could
only be **half** done. The second routine's file was in a state where committing it would have
broken everybody else's build, so we could not touch it. That forced us to leave the two routines
carrying **identical** copies of the same decision — and it forced us to hold back one page type
(`section-index`, the "list of other pages" kind) that we knew needed fixing, because adding it to
one copy and not the other would have made them disagree, and when they disagree one silently wins.

We left a note saying: check whether that blocker is still real before assuming it.

## What we've done

Checked. It was not still real — other people had tidied that file overnight. So today the second
copy is **deleted**. There is now exactly one place in the system that answers "which builder
builds this kind of page", and both routines ask it. The page type we had held back went in at the
same moment, which is exactly what the reviewers had asked for: one line now moves both doors.

The review **approved it first time — thirteen reviewers, no vetoes, nothing above "low"**. The
equivalent yesterday took six rounds.

We also wrote the first tests this routine has ever had. It had none. That mattered more than it
sounds: our first attempt at those tests **passed even when we deliberately broke the code**, for
two entirely separate reasons, and we only found the second because fixing the first still did not
make it fail. We logged both, because a test that cannot fail is worse than no test — it is a
false assurance somebody will quote later.

And we corrected two statements we ourselves had written into the code the day before, inside the
change the reviewers had approved. One of them claimed the two routines could not collide. They
can, and that collision was the whole reason for the reviewers' objection that shaped the change.
One search of the other routine disproved it.

## Where we are now

**The design work is finished.** There is one authority, both routes use it, every page type we
know about is routed, and it has tests that fail when the code is wrong. That part of this bug is
closed in substance.

**But the more valuable thing we found today is that our proof was not a proof.** We had a query
ready to demonstrate the fix works: find a build job created automatically, pointing at the right
builder. If that exists, the fix works.

It does not follow. That field can be edited by hand, and editing it by hand is our own documented
repair procedure for a stuck page — we have used it at least twice. So a page a person rescued in
August is indistinguishable from a page the fix routed correctly. We checked how bad that was:
**every single record in the entire database that would have passed our test is one of those hand
repairs.** All three of them. We would have declared the fix proven using records created by the
very bug it fixes.

There is a clean way to tell them apart and it now gates the query. It is clean *right now* in a
way it will not stay: the marker it looks for exists on no record at all yet, out of five hundred
and eight, so the first record that has it was necessarily made by the new code.

Separately, three reviewers asked whether other parts of the system also skip the shared decision.
They do — three more places, one of them the busiest in the fleet. The obvious conclusion is that
the bug is still alive through them, and we were one sentence from writing that down. We counted
instead: across every job those places have ever created for these page types, **not one** shows
this failure, and there is a reason — this bug is about pages that have no layout *yet*, and those
routines only ever touch pages that already have one.

## Where we're going

The bug stays open, and for the same reason as yesterday: **nobody has watched it work on a real
page.** Today's code is committed but not yet shipped to the running fleet, and no new website has
been built since the previous half went live, so there is nothing to read yet.

The proof still costs nothing and needs no site disturbed — it arrives the next time anyone builds
a new site with one of these page types on it. The difference today is that when that happens, we
now have a test that can actually tell us the answer, which yesterday we did not have and did not
know we did not have.

Three things are named as somebody's next job rather than quietly left: a mismatch between two
rules about which jobs count as "still open" (which has one real casualty we can point at, and
which today's fix cannot reach); a label the system writes and nothing ever reads; and the larger
version of this problem, which is that the ordinary builder cannot give *any* page a starting
layout — that is bigger than this fix and wants its own review, not an extension of an approved
one.
