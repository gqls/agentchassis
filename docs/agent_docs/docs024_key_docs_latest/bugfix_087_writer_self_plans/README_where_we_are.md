# Where we are — the page writer that could not plan its own work

Plain prose, append-only, newest at the bottom.

---

## 2026-08-04, late morning — what this was, and what it turned out to be

The bug, 087, was filed back on 26 July. In plain terms: when we ask the system to rebuild
a single page, it hands the job to a "page writer" agent — and it was handing it over
**without telling it which sections of the page to write**. The writer had no way of
working that out for itself, so it died. Not gracefully: it fell over two steps later, on
an error that named a missing key rather than the missing instruction, which is why it took
a while to understand.

Somebody had already fixed part of this on 27 July, and verified it worked on the 28th. The
ticket stayed open for a slightly awkward reason: the test they chose to prove it landed on
a page the system is not allowed to overwrite (a tool page, which is "owned" by a different
pipeline), so a guard quite correctly refused. And in the process that attempt exposed a
*worse* bug — the rebuild was publishing pages to the wrong web address — which had to be
fixed first. That one, 125, was closed on the 31st.

So my job started as: find a page we *are* allowed to rebuild, run the test, close the
ticket.

## What I found instead

The fix from 27 July was applied to **one** of the callers. There are four different agents
in the system that ask this page writer to do work. Two of them were fine. The other two —
one of which is the builder our own site classifier recommends by default — had exactly the
same hole and nobody had noticed, because the ticket was written about the third one.

So the ticket's own title was still literally true, just not for the agent it named.

At that point the obvious move — copy the 27 July fix into the other two — is the wrong
move. That would leave four hand-maintained copies of the same instruction in four
different places, and we have been bitten by precisely that before: the page-address bug I
mentioned above turned out to be *five* copies of the same logic, four of which agreed and
one of which did not. Copies drift. The one that drifts is the one you find in production.

## What I did

I made the writer able to plan its own work. If the agent asking for the page supplies a
plan, it is used exactly as before, untouched. If it does not, the writer now builds one
itself. That means **no caller, present or future, can get this wrong** — including callers
nobody has written yet. It closes the hole rather than patching the three places it
currently shows.

I added one more thing that was not in the original ticket. If the writer plans its own work
and concludes there is *nothing it can write*, it now stops and says so. Before, it would
have quietly produced an empty page and published it over a real one. That is the kind of
failure nobody notices for weeks. The main build pipeline already refused this case; the
writer simply had no equivalent, so I gave it the same judgement in the one place every
caller passes through.

All of this is configuration, not code, which means it went live the moment I applied it —
no waiting for the next software release. That mattered today, because another session is
mid-flight on a related software fix and this change does not have to queue behind it.

## The test, and the thing the test found

I picked a guide page on vetcomparison.uk — one we are allowed to rewrite, untouched for two
days, and deliberately one whose file name differs from its web address so that the
page-address fix from the 31st would be genuinely tested rather than vacuously passed. I
armed exactly that one page, inside a database transaction that would abort if it found more
than one, because on a system this busy "I checked a moment ago" is not the same as "it is
still true".

The rebuild ran end to end. Every stage completed, the page republished correctly at the
right address, the wrong address stayed a 404, and — the part I most wanted to see — the
writer recognised the plan it was given and did *not* re-plan. That is the check that could
have come out the other way, and it is what tells me the change is inert for everything
already working.

**But it also left the page slightly worse off in a way I did not expect.** Each section of
a page is stored twice: as finished HTML, and as the structured content it was built from.
The rebuild wrote the HTML and threw away the structured version. Nothing visible broke —
the page serves perfectly — but that structured version is what we need if we ever want to
*re*-render the page, so the page had quietly lost the ability to be regenerated.

That is not something my change caused; the telemetry from the run shows the writer took
exactly the same path it always did. It is a separate defect, and a slightly embarrassing
one when you see it: the save step accepts a setting that says "and here is the structured
content", and **four of the six agents that call it simply never set it**. The data is right
there in the writer's reply; it is just not passed along.

I have filed that as its own ticket, 194. I also fixed it for the rebuild agent specifically
and re-ran the page, because my own test is what caused the loss and recording damage without
repairing it is not really recording it. I deliberately did not fix the other three — two of
them are the same dormant pair from the story above, and the third runs a different flow whose
reply I have not actually read, so copying the setting across would be a guess dressed as a
fix.

One thing I considered and rejected: the old structured content *is* archived, so I could
have put it back. But that would pair yesterday's content with today's HTML, and the next
re-render would have quietly reinstated the old page. Re-running the build writes both halves
together, which is the honest repair.

## Where this leaves us

087 is done and closing. The class of bug behind it is closed too, not just the instance —
which was the point of doing it this way rather than the quick way. 194 is open, with one of
its four instances fixed and the other three written down with the evidence. And there is one
piece of tidying I deliberately left for after the next software release, recorded with the
exact change and the test, rather than done today while the ground is moving.
