# SUMMARY 2026-08-19 — the loop engine, and the states nothing was watching

*First summary for this lane, written at its completion. Current state only — the chronology is in
`NOTES_tool_audit_by_instance.md`, the plain-English history in `README_where_we_are.md`.*

---

## What we're trying to do

Make the tool-audit machinery actually work on this estate, and fix whatever it turns out to be
sitting on. The lane began as a narrow job — the audit was blind to ported page tools, 63 of 67 on
one site — and the interesting part is what that exposed underneath it, which had nothing to do
with tool audits and affected every job the platform runs.

## Where we've come from

The audit machinery was blind, so we ported it. Once it could see, it started running — and almost
never finished. A `tool-auditor` job completed **once in sixty-three attempts**. Chasing that led
away from the audit entirely and into the shared loop engine that every repeating job uses.

The cause was a name collision in the workflow model. A loop's final step and the final step of
each individual lap are both declared with the same action, `loop_complete`. The engine could not
tell them apart, so it inserted the whole-loop summariser once **per lap** — and each lap then
swallowed every earlier lap's summary. The data a job carried doubled every time round: two laps,
four, eight. Real jobs reached 22 MB and died of their own weight, and the more findings a run had,
the more certainly it killed itself.

Then, clearing up the wreckage, we found the second problem. The dead jobs were not just dead —
they were **immortal**. Nothing in the platform would ever look at them again, and each one held
two messaging channels open permanently.

## What we've done

Five bugs opened by this lane; **all five are closed, fixed and live.**

- **281** — the audit's blindness to ported tools. Closed.
- **289** — the doubling. Fixed by teaching the engine to mark each lap's final step explicitly at
  the moment it creates it, rather than guessing from the action name. Live since `v1.0.1307`.
- **291** — the auditor filing work for a reviewer that had never existed. Closed.
- **294** — jobs stuck in the "running" state were reachable by no recovery path at all. Fixed by
  giving the tidy-up job an arm for that state. Live, and it also released the pinned channels.
- **310** — the identical hole one state along, for jobs stuck at "just created". Fixed the same
  way. Live since 18:43 on 18 August; three minutes later the tidy-up job cleared both stranded
  jobs, including one that had sat there for thirty-six days.

Two things beyond the fixes are worth recording. We wrote down a **latent trap**: the code names the
paused-job state one thing and five separate production queries look for a slightly different
spelling, so whoever switches on pause-for-human will silently lose their job's channels. And we
wrote down a **measurement trap** on the same table: a status count there is a one-day window for
some states and an all-time count for others, so any rate built from it is wrong in one direction
or the other unless you check which is which.

## Where we are now

Verified this morning against the current build, `v1.0.1314` — read from the image's own commit
label rather than from the tag, because a rebuild at the same tag can ship no new code at all:

- The doubling is gone. Each lap now carries **77 bytes, flat**, across every multi-lap job since
  the roll. Demand control: 123 jobs ran in that window, so the flatness is a result and not an
  idle fleet.
- **No job anywhere is stuck for more than four hours.** Both recovery arms are live and the
  tidy-up job is ticking normally.
- The loop engine's one remaining cleanup item was **ruled on and deliberately not taken** — see
  "Where we're going", item 1. Nothing on this lane is now waiting on anyone.

The single most useful thing we learned is not in any of the fixes. Twice this lane argued that a
state was safe because of what the code does, and the second time that was **wrong**. One of those
states is brief because the code makes it brief — nothing can lengthen it. The other only *looks*
like it, but it is genuinely waiting for a message to arrive, so how long it sits there is a fact
about how busy the system is, not about the code. Reading the code told us what *writes* a state,
never how long anything *sits* in it. A concurrent session caught it by measuring: the figure we
had derived was out by about a thousand times. That is now written into the bug file, the wrong-call
log and memory, because the reasoning is far more transferable than the fix.

## Where we're going

The lane is effectively complete. Four things outlive it, and only the first is ours:

1. **DECIDED 2026-08-19 — nothing is left here.** The engine still keeps a second, older way of
   recognising a lap's final step, kept for jobs that were already running when the fix shipped.
   Those have all finished — measured, twice — so it could now be deleted, and a reviewer had asked
   for that. **The owner ruled that it stays.** It has quietly become a second independent guard
   against the exact failure that caused all this, it costs two lines, and deleting it would make
   correctness depend on a single line being right for ever. The ruling is recorded at the code
   itself, so someone arriving to simplify that function meets it before they cut.
2. **The class of bug is still open.** We have now patched two states that nothing was watching,
   one at a time. The platform lists these states by hand in six different places and **no two
   lists agree**, so forgetting one is the normal outcome rather than an accident. The general fix
   is a single rule instead of a list; the owner has taken it as separate work. It is blocked on
   settling the paused-state spelling above, or the general rule would inherit the wrong spelling
   and start killing paused jobs while looking correct.
3. **The audit machinery's yield is unresolved and belongs to the neighbouring lane.** A run can
   produce eighteen findings and file at most one review row per page. Chasing the artefact this
   morning, it is not in the live table or the archive. It is recorded only inside closed bug
   files, which is where a residual goes to be forgotten.
4. **Two sessions built the same fix for 310 in parallel, letter for letter identical.** Nothing
   available today can show you a session that is working but has not yet written a file. That is
   a coordination cost we paid in full and should expect to pay again.
