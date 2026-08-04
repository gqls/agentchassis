# SUMMARY — 2026-08-04 — the page writer now plans its own work

## What we're trying to do

Close `bugs_open/087`, which had been open since 26 July. In plain terms: when the system is
asked to rebuild a single page, it hands the job to a content-writing agent — and it was
handing that job over without saying which parts of the page to write. The writer had no way
to work it out for itself, so the job died. Not cleanly: it fell over two steps later on an
error naming a missing key rather than the missing instruction, which is why it took a while
to understand what had gone wrong.

## Where we've come from

Someone had already fixed part of this. On 27 July a bug-sweep thread added a planning step to
the rebuild agent, and on 28 July they proved it worked — the writer received a real plan and
got through the step that used to kill it.

The ticket stayed open anyway, for a slightly awkward reason. The page they chose to prove it
on was one the system is not allowed to overwrite, so a guard quite correctly refused, and the
test could not finish. Worse, the attempt exposed a more serious bug underneath: the rebuild
was publishing pages to the wrong web address, creating live duplicates. That became its own
ticket, 125, and was closed on 31 July. So 087 sat waiting for a clean run on a page we *are*
allowed to rebuild.

## What we've done

We started by checking the ticket was still real, and found something the file did not say.
The July fix had been applied to **one** of the callers. Four different agents ask this writer
to do work; two were fine, and two others had exactly the same hole — no planning step, no
plan passed across — and nobody had noticed, because the ticket was written about a third
agent. One of those two is the builder our own site classifier recommends by default.

That changed the shape of the answer. The obvious move — copy the July fix into the other two
— would have left four hand-maintained copies of the same instruction in four places. We have
been bitten by exactly that before: the wrong-web-address bug turned out to be five copies of
one piece of logic, four agreeing and one not, and the odd one out was the one the build
agents actually reached. Copies drift.

So instead we made the **writer** able to plan its own work. If the agent asking for the page
supplies a plan, it is used exactly as before, untouched. If not, the writer builds one
itself. No caller can get this wrong any more — including callers nobody has written yet. We
added one thing the ticket did not ask for: if the writer plans its own work and finds there
is nothing it can write, it now stops and says so, rather than quietly publishing an empty
page over a real one. The main build pipeline already refused that case; the writer simply had
no equivalent.

All of it is configuration rather than code, so it went live the moment it was applied. That
mattered today, because another session is mid-flight on a related code fix and this change
did not have to queue behind it.

We then ran the ticket's own acceptance test, twice, on a guide page on vetcomparison.uk —
chosen because we are allowed to rewrite it, no other session was touching it, and its file
name differs from its web address, so the wrong-address fix from 31 July was genuinely tested
rather than vacuously passed. Everything completed, the page republished at the right address,
the wrong address stayed a 404, and the check that mattered most came out right: the writer
recognised the plan it was given and did **not** re-plan. That is what tells us the change is
inert for everything already working.

## Where we are now

087 is closed — fixed, live, and proven. The class behind it is closed too, not just the
instance, which was the point of doing it the longer way.

The test also found a second, unrelated defect and we filed it as 194. Each section of a page
is stored twice — as finished HTML, and as the structured content it was built from — and the
rebuild was writing the first and discarding the second. Nothing visible breaks, but that
structured version is what we need to re-render a page later, so pages were quietly losing the
ability to be regenerated. The cause is a setting the save step accepts and **four of the six
agents that call it never set**, with the data sitting right there in the writer's reply. We
fixed it for the rebuild agent and re-ran the page, because our own test is what caused the
loss; we deliberately left the other three, two being dormant and the third running a
different flow we have not actually read.

Two missteps are recorded. The more useful one: a check we wrote into the migration to prove
we had not damaged the loop **could never have failed** — it named a key that does not exist,
and in SQL a comparison against a missing value is neither true nor false, so the alarm could
not fire. It read as coverage. That is now a landmine entry and a WRONG_CALLS row, along with
the habit that catches it: run the verification alone against the unchanged system first, and
require it to fail.

## Where we're going

One piece of tidying is deliberately deferred rather than forgotten. On the new
self-planning path the internal-link resolver is handed nothing, because of how a lookup path
falls back — so links in those sections will not be resolved. It is a one-line change to fix,
but it touches something another session has deliberately pinned while their code fix is
unrolled, and it would trade an exact lookup for a fallback chain on the one path that
currently works. It is written down with the exact edit and the test, to be done after the
next release.

Beyond that: 194 has three callers left, and the original ticket's boldest option — retiring
the single-page rebuild agent altogether, which two live paths already cover — is now
decoupled. The writer is safe either way, so that decision can be taken on its own merits
whenever the owner wants to take it.
