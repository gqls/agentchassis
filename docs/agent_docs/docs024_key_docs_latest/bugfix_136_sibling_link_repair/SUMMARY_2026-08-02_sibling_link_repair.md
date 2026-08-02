# Summary — 2026-08-02 — the sibling link-repair bug (`bugs_open/136`, section-editor slug)

Written to be read aloud. First in the series for this workstream.

## What we're trying to do

Stop the platform publishing links to pages that do not exist. When a writer produces page
copy it sometimes invents a plausible destination — "see our pricing" pointing at a
`/pricing` page the site has never had. A visitor clicks it and gets a 404 on a site we
built for a client.

We already have the repair for that. If the target really exists under a slightly different
address, we correct the address; if it does not exist at all, we keep the words and remove
the link. The job here was to make sure that repair actually runs everywhere a page's HTML
is saved, rather than at the one place somebody happened to wire it.

## Where we've come from

The repair was built for an earlier bug and wired into the code that saves a whole page's
sections — the normal route for generated copy. When that fix went through our review
council, one seat objected that nobody had shown this was the *only* place page HTML gets
written; it had only been shown to be the only place with a particular step name. The check
it asked for found four more writers, none of which repaired anything. That objection became
this bug, and it sat unowned for five days.

The awkward part is what the gap meant in practice: the more carefully a session followed
our own written guidance — which points at the targeted "edit one section" path — the more
reliably it bypassed the protection.

## What we've done

Two of those writers now repair their links: the targeted section editor and the report
generator. Both call one new shared function, which is a thin wrapper around the repair that
already existed, so there is one set of rules rather than two copies that drift.

Doing it turned up the same trap one level down. The section editor saved its work in two
different places depending on which kind of edit was requested, so a guard in front of one
would have been silently skipped by the other. Those two saves are now one save with the
repair in front of it.

We also added a check that runs when anyone commits: if a future change writes page HTML and
does not repair links, the author is told at the moment they do it. That matters more than
either fix, because the set of writers is not stable — a new one appeared two days after
this bug was filed and nobody noticed. A list in a document cannot keep up; a check at the
moment of the edit can.

The change went through the review council and was **approved at the first round**, with
five advisory objections and none of them serious. Three led to real work. The sharpest came
from a seat that read only the plan and never saw the code: it predicted that one of our
tests proved nothing, and it was right — the test would have passed even with the feature
deleted. That is fixed and the fix is proven by deliberately breaking the code to watch the
test fail. Two others asked us to actually run lookups we had only claimed to have run; both
took one command and both confirmed what we had assumed, which is not an argument for
skipping them.

Four seats independently said the same thing about the commit-time check: advisory is the
wrong ceiling, because an advisory warning can be ignored. They are probably right, and the
proper answer is a bigger structural change that our own rules say must not be smuggled into
a bug fix. So it is now written up as a formal proposal, including the strongest argument
against it and the one measurement that would settle the question — whether anyone actually
reads advisory warnings at all. Nobody has ever checked.

## Where we are now

The fix is committed and approved. **It is not live**, and that is the honest status: our Go
code does nothing until a new image is built and rolled out, and rolling one right now would
kill another session's review that is mid-flight. The exact before-and-after checks are
recorded and the "before" half is already measured, so proving it shipped will be three
commands rather than a judgement call.

I also measured the real damage, having found the bug file explicitly refused to guess at it:
**35 links in stored page HTML that point at pages which do not exist**, across 13 components
on 6 sites. Seventeen are dead and would be removed; eighteen are near-misses we can repair
automatically. Two honest caveats. That stock cannot be blamed on the writers I fixed — a
stored link does not record who wrote it. And neither of the two paths I guarded has actually
run in the twenty days of history we keep. This is prevention on live, documented, reachable
paths; it is not a bleed I have stopped, and I would rather say so than let it sound bigger.

I got those numbers wrong the first time, and the wrong ones are in a commit message that
cannot be edited. I had read them off a results table on screen instead of asking the
database to count. Writing the runbook caught it. The correction is recorded everywhere the
claim will next be read, and in the fleet-wide log of wrong calls.

## Where we're going

1. **The roll.** Whoever next builds the chassis picks this up; then three greps prove it
   live and the ticket closes. Until then it stays open, because the defect is still
   reproducible on the running system.
2. **The two tool-page writers.** They genuinely have this problem and hold the largest share
   of the live damage. They were left out because several sessions are editing those exact
   files right now and our commit rules cannot stop two sessions' changes to one file being
   mixed together. The new commit-time check will keep pointing at them until somebody does
   it — deliberately, because silencing it would turn a real debt into a false all-clear.
3. **The structural question**, now a filed proposal rather than a line in a bug file.
4. **The existing damage** is untouched by all of this. Those seventeen dead links are still
   live for visitors, and the check that would find them fleet-wide has never run on any
   site — which is a different, already-filed bug.
