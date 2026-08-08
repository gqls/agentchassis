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
