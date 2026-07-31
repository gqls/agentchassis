# Where we are — bug 154, the work item whose own data was invisible

Plain-prose log, append-only, newest at the bottom.

---

## 2026-07-31, evening — what this bug actually was

The platform keeps a queue of jobs, one row per job. Different agents put jobs
on the queue; a dispatcher picks them up and hands each one to whichever agent
is meant to do it.

One of those agents, `tool-auditor`, files jobs that say "this tool on this site
needs fixing". Every single one of them died instantly, at the first step of the
agent meant to fix it, with an error saying it had not been told *which* tool.
Jobs filed by three other agents went through fine.

The odd part — and the previous session spotted this and wrote it down — is that
the jobs which failed were exactly the ones that **did** say which tool. The
column was filled in. The error said it was empty.

Both things were true, because there are two places a job can record the tool,
and only one of them was being read. The queue table has a proper column for it.
There is also a free-form blob attached to each job, and by convention most
agents drop a copy in there. The dispatcher was only ever looking in the blob.

The reason is a single line of code. The dispatcher builds its picture of a job
from a database query, and that query never asked for the column. It asked for a
neighbouring one — `page_id` — and stopped. So three real columns
(`component_id`, `entity_id`, `affected_url`) were invisible to every agent in
the system, no matter what was in them.

**`tool-auditor` was the only agent doing it properly, and that is precisely why
it was the only one broken.** Everyone else duplicated the value into the blob
and got away with it. Nothing warned about any of this: the dispatcher marks that
field "optional", and when it cannot find an optional field it quietly moves on.
The complaint surfaces two agents later, phrased as though the data was never
there.

## What I changed

The dispatcher now reads all three columns, and for each one it takes the column
if it has a value and falls back to the blob if it does not. That means a single
lookup path is correct for both kinds of job — the 235 existing ones that use the
blob, and the ones that use the column — so nothing has to be rewritten and the
next agent that fills in the column just works.

I checked this actually helps rather than just changing the error message: all
four stuck jobs point at tools that really exist, are live, and sit on a real
page — which is exactly what the fixing agent looks for. So they should now run.

## Two things I deliberately did not do

There was a tempting shortcut: copy the column's value into the blob, and then
nothing else needs changing at all. I nearly took it. It turns out another part
of the system reads that blob to decide whether a rebuild should cover the whole
site or just one component — so quietly putting a value in there could have
changed how sites get rebuilt, for reasons nobody would ever connect back to
this. I left the blob alone entirely.

The second: there is a fourth column, `page_id`, that is handled slightly
differently from the three I fixed, and it would be tidier to make it match. But
218 existing jobs would suddenly start carrying a value they do not carry today,
and the agents receiving them have not been written with that in mind. Tidiness
is not a good enough reason to change what several hundred jobs hand to their
handlers. I have written down the measurement so someone can revisit it with
evidence rather than instinct.

## Where it stands

The code is written, tested, and committed. I proved the test can actually fail
by deleting the fix and watching it go red — a test you have never seen fail
tells you nothing.

Two reviews are running: the automated diagnosis loop, which re-derives the cause
independently, and the reviewer council on the code itself. Neither had returned
when I committed, which is normal and expected here — the tree is shared, so
holding work back is not actually an option, and the commit is tagged with the
review's id so it gets credited automatically when the verdict lands.

The remaining step is that the fix is in the **program**, and the line of
configuration that uses it lives in the **database**. The database is live
instantly; the program only changes when a new image is built and rolled. So the
config change has to go second, and I have written the exact ordering down. Until
then this is inert — correct, committed, and not yet doing anything.

One more thing the previous session found and I have left alone: `tool-auditor`
names its jobs after the *site*, not the tool, so two jobs for two different
tools on one site are indistinguishable. That is a real problem but it is a
different one, in a different place, and fixing it would change how jobs are
de-duplicated across the whole fleet. It stays on the ticket rather than riding
along in a change about something else.
