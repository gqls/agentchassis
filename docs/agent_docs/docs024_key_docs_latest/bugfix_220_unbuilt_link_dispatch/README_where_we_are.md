# Where we are — bugfix 220 (plain prose, append-only, newest at the bottom)

## 2026-08-08 evening — picked up, plan settled

The improvement loop had a nasty habit: when it found a link on a live page pointing
at a page that was never built, it "fixed" it by rebuilding the page the link sits on
— which was never broken — and then marked the job done. The missing page stayed
missing, the link stayed a 404, and the next sweep found it again and did the same
thing. Every cycle looked green and cost a full page build.

The team that found this (while closing bug 206) wrote it up but didn't take it, so
this lane has. Three things need doing and two of them are small config/code changes:
tell the dispatcher to hand the worker the ID of the page the job is actually about
(it already stores it — it just never passed it on; the sibling dispatcher already
does exactly this); teach the page loader that when that ID is handed over
explicitly, it wins over the page name (today the name wins, which is the quiet root
of the whole thing); and register a completion check so a job of this kind can only
be marked done when the missing page is genuinely live, or the link genuinely gone.
That last piece slots into the verification framework another lane made strict
today, so the timing is good.

One deliberate deferral: pointing directory-type pages at the new directory builder
(rather than the generic page builder) is left as a recorded follow-up — with the
completion check in place, that case now fails loudly instead of lying, and the
directory team's own machinery is the proper fix for it.

## 2026-08-08 late evening — built, committed, live config applied; waiting on review + the next release

All three pieces are done and committed. The database side is already applied and
proven by reading it back: the dispatcher now hands workers the right page's ID, and
the sweep that used to auto-complete these jobs on circumstantial evidence now leaves
them to the real check. The code side (the loader honouring that ID, and the
completion check itself) is in the shared branch and takes effect when the owner next
rolls the fleet — until then everything behaves exactly as before, which was measured
rather than assumed. The reviewer council is looking at the change now; the commit
carries the pending-review marker so it gets credited automatically on approval. One
honest wrinkle for the record: two other teams' index rows travelled in my commit
(unavoidable when several people edit one file), and my own log entry travelled out in
someone else's — both noted where they happened, nothing lost either way.

## 2026-08-08 night — reviewers caught two real gaps (fixed), then the review system ran out of API credit

The reviewer council sent the change back once, and two of its objections were
genuinely right — a missing-page case that would have failed silently, and a text
match that could misread links containing underscores. Both are fixed, tested and
committed. The resubmission then died for a reason nothing to do with the work:
**the platform's Anthropic API account is out of credit, and every AI-driven process
fleet-wide has been failing since about half past six this evening — that needs a
billing top-up from you.** The moment credit is back, one saved command re-runs the
review (it's in the notes file). The fix itself is unaffected: code is on the shared
branch and goes live with the next release; the database side is already applied and
verified.

## 2026-08-08 late night — review approved

Your credit top-up went through, the saved command re-ran the review, and it came
back approved (a few advisory notes, all answered in the technical log — none
needing action). So this piece of work is now finished except for the proof that
only the next release can provide: once you roll the fleet, the notes file has the
two checks that confirm the fix is really running — a binary check on every pod, and
one live run of the improvement loop against a site with a missing page, which
should now build the missing page rather than pointlessly rebuilding the page that
links to it.

## 2026-08-09 morning — the new build proved the fix, caught a deeper problem, and we closed that too

Your fresh build carries everything; both running copies were checked directly. We
then ran the real test: pointed the improvement loop at dartsonline, which has a
blog post that five live pages link to but that was never built. The good news —
the dispatcher now aims at the right page (the old behaviour shipped the wrong
page's file; the new one didn't), and the new completion check works. The catch —
the test exposed one last inconsistency inside the page builder itself: the step
that SAVES written content was still using the old way of naming its page, so it
saved the missing page's freshly-written copy onto the page that links to it. One
live page (the darts "beginners" post) had its stored draft overwritten with the
wrong article; its published page is still correct because we cancelled the two
jobs that would have republished it, a repair job is queued to rewrite it
properly, and the one-line configuration fix that closes the whole class is
applied and verified. A second page was protected automatically by an existing
safety floor, which refused a suspiciously small overwrite — nice to see.

Still to watch: the repair landing on the beginners post, and then one clean
end-to-end run where the missing page actually gets built and published. The next
session picks up from the handoff file in this folder.

## 2026-08-09 midday — the beginners post is repaired, and it is repaired on the real website

The repair job finished, and I checked it the way we say to check things: not by
looking at the job saying "complete", but by looking at the page. The stored draft
of the beginners post is beginners copy again, the site republished it just after
midday, and fetching the live page returns the right article. So the damage from
this morning is fully undone, and the two jobs we cancelled to prevent it spreading
can stay cancelled — nothing is waiting to republish the wrong thing.

One neat detail that tells us the test is still a fair test: the repaired beginners
post links to the grip-styles guide, and that guide is still missing, so the link
still leads nowhere. That is exactly the fault this whole piece of work is about,
sitting there in the open, ready to be found again. Which means we can now run the
one thing we have never actually watched happen from end to end: the system
noticing the dead link, building the missing page, publishing it, and only then
declaring the job done.

I have started that run. What I am watching for is the missing grip-styles page
going from "planned" to actually published, and the job that fixes it closing with
its own verification saying the target page has shipped — rather than the weaker
reason it gave this morning, which was really just a side effect of the damage.

## 2026-08-09 half past two — it worked. The missing page built itself, and the link is alive

The run I started before lunch reached our jobs at twenty past two, and the thing we
have been trying to watch happen for two days happened twice in six minutes.

The grip-styles guide — the page that has been "planned but never built" all week,
the one every dead link on the site was pointing at — got built. Three sections
written, published to the site at 14:28, and the address that has been returning
"page not found" all morning now returns the page. I fetched it to be sure: it is
there, it has its own title, it is the grip-styles article and not a copy of
something else. And the page that contained the dead link, the barrel-weight guide,
still says "Barrel Weight Guide" and still contains its own writing. That last part
is the whole bug. The old behaviour was that the system would rebuild the page that
*contained* the broken link instead of the page the link *pointed at*, overwrite it
with the wrong content, and then announce success. It did none of that. It built the
right page, left the other one alone, and only then closed the job — and the reason
it gave for closing was the strong one: "the target page has shipped, the link now
resolves."

So bug 220 is finished. Fixed, live, and proven from the dead link all the way to a
working web page.

Two things I want to flag, because both mean somebody was wrong and it is better
said out loud.

The first is mine. Before I started watching, I ran the check this folder tells you
to run — a known-bad job from this morning, which the notes say should come out
looking wrong in a particular way. It did, so I said the check was sound and could
be trusted. It was not sound. One of the columns it looks at was reading blank for
every job on the system, because the instructions had the wrong address for that
piece of data. I could not tell, because on the known-bad job that column was
*supposed* to be blank. A blank that means "nothing was published" and a blank that
means "you are looking in the wrong place" are the same blank. If the first success
had not landed in front of me while my check was still calling it a failure, I would
have told you the fix had not worked. I have corrected the instructions, added a
second check that reads a job which definitely did publish, and written the whole
thing up as a wrong call. The lesson is small and sharp: a check only tests the
things it expects to *find*. Anything it expects to be empty, it is not testing at
all.

The second is this folder's. We had written down, twice, that four of the ten jobs
were expected to fail loudly, because they pointed at a different kind of page —
section listings rather than articles — that we believed the builder could not make.
One of those four is what succeeded first. It built the brands listing page and
published it, no complaints. So that belief was wrong, and it was never actually
tested before we wrote it down; we had extrapolated it from a different kind of page
failing earlier in the day. It matters because that supposed failure was being used
as the argument for a further piece of work we deferred. On this evidence the
argument is weaker than we thought and possibly gone. Three of those four jobs are
still working through the queue as I write; their results are the honest answer, and
nobody should pick that deferred work up without looking at them first.
