# SUMMARY — 2026-07-26 — bug 074 closed: the task that reported success and did nothing

## What we're trying to do

Make it impossible for a scheduled job on this platform to look like it is working when it is not.
The specific case: a task could carry its own programme of work in a shape the system accepted,
ignored, and never mentioned again — firing on time, reporting success, advancing every timestamp
a health check reads, and doing nothing at all.

## Where we've come from

The bug was filed on 25 July by another thread, and only because that thread deliberately broke
something to see whether a check would catch it. It didn't, and chasing why revealed that the
check had never run in its life. Three tasks were written in the bad shape; that thread repaired
its own two by hand and left the general defect, plus one task belonging to the
claims-verification work, for whoever came next.

Two closed cases stood behind it. One had already ruled that a task's data column is the payload
only and that the scheduler must not go rummaging in it — a ruling that turned out to decide this
fix. The other was the sibling defect, where a task's data was written one level too deep and an
export silently refused for weeks.

## What we've done

Read the code before the report, which changed the answer. The system *does* support a programme
of work sent in a message — other parts of the platform use that route constantly. What it cannot
do is send one *from a scheduled task*, because the scheduler builds the message envelope out of
the task's own columns and buries anything the author wrote a level below the only code that reads
it. So the honest choices were "teach the scheduler to dig it out" or "refuse the shape". The
prior ruling said don't dig; the owner agreed, with the alternative explicitly on the table.

Three things went in:

- **A database rule that rejects the shape at the moment it is written** — live on application, no
  deployment, and it fails the person making the mistake rather than the run that inherits it.
  Verified the failing way round (the bad shape is refused, naming the rule) and with a control (an
  ordinary task still saves), so the refusal isn't passing for some unrelated reason.
- **The last broken task repaired**, its programme moved to where the system genuinely reads it,
  copying the pattern proven the day before. Staged behind a dry run because it rewrites the
  approved-numbers list on two sites other threads are actively working on.
- **The scheduler taught to refuse and say so**, rather than manufacture a successful-looking job
  for work it never sent. Committed; **inert until the scheduler image is next built**, and said
  plainly rather than implied to have shipped.

Then we broke a figure on purpose to prove the repaired check actually catches things — it did,
with the right explanation, and put the real number back itself.

## Where we are now

074 is closed: the shape cannot be authored, no task carries it, and the last casualty runs. That
casualty is the claims-verification freshness sweep, and its first run in existence checked 24
published figures across four sites, re-synced 13, and raised three for a human to rule on where
published copy has drifted past what its wording allows.

The deliberate fault also exposed a smaller defect, filed as 091 and not fixed here: when a
"figure has drifted" notice is already open for a site and a different figure drifts, the new
finding is dropped and the run still reports that it raised it. Recoverable, but while the notice
is open it describes the wrong problem — and the fix touches machinery every detector shares, so
it belongs to that machinery's owners.

One correction of our own is on the record: a count written from a query minutes old, already
stale because another session had rewritten the rows in between. The sweep's own report caught it.

## Where we're going

Three things are owed, all named where they will be looked for:

1. **The scheduler build.** Until it rolls, the runtime refusal is inert; after it, one pod-grep
   confirms it — a string the change created, with a control.
2. **091**, for the work-item and claims-verification owners: the contained half is one line (stop
   reporting a write that did not happen); the fleet-wide half needs care.
3. **Two additions to the debugging guide's existing entry for this case** — that the receiving end
   did support the field while the sending end could not express it, and that refusing at
   authorship beats warning at use. They are written and waiting, unpasted only because another
   session has uncommitted work in that file and a same-file passenger cannot be left out of a
   commit.
