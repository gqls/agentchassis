# Summary — tool-suggester phantom tool links (bugs_open/029 → bugs_closed)

**2026-07-26.** First summary in this workstream; written at the milestone where the fix is built,
the damaging behaviour is switched off in production, and the case is closed.

## What we're trying to do

Stop the platform putting links on customers' web pages that lead nowhere. The specific case: when
the system decides a site could use a new interactive tool — a calculator, an estimator — it also
writes instructions telling the copywriter to mention that tool on a few existing pages and link
to it. Those links were going to pages that did not exist. A visitor to
leopardessconsulting.co.uk clicked one and got a blank 404.

## Where we've come from

The damage was first blamed on a different bug entirely (the re-plan clobber, `001`) and had to be
re-attributed. Then, on 21 July, another look corrected the story again: it was not that the
copywriter had invented a plausible-looking address. **The system itself invented the address** —
built it from the tool's name to a fixed pattern — wrote it into the instruction, and then wrote
an acceptance test demanding that exact address appear on the page. The copywriter was following
orders, and the test it was marked against required the broken link.

The finding that settled the shape of the fix: tool pages do not have one address pattern. They
have three, depending on which of three internal routes built them. So the guess was wrong even
for tools that had been built successfully. Across the whole fleet, **not one** of these generated
links pointed at a real page.

That is why this could not be fixed where it was. An address that three different parts of the
system write in three different ways cannot be reconstructed from a name; it has to be read from
the record of the page itself — and that record does not exist until the page has been created.

## What we've done

**Moved the job to where the answer is known.** The instructions are now written by the part of
the system that *builds* the tool, at the moment it builds it, using the address it has just
created. Nothing constructs an address any more; the code refuses to write an instruction at all
if it is not handed a real one.

**Added a hold-back the original plan did not have.** A tool's page is created before its content
is written, so for a while it exists but is not published — and on the current fleet a lot of
those content jobs are parked waiting for a person. Left alone, that would have recreated the same
dead link by a new route. So each "mention this tool" instruction is now held until the tool's page
is actually live, using the dependency mechanism the platform already has. If the page never goes
live, the mention is never written. We also confirmed those held-back instructions cannot pile up
forever: an existing cleanup job clears them after 48 hours.

**Switched the old behaviour off in production immediately.** A configuration change (migration
211) deleted the step that was writing these instructions at suggestion time. That took effect the
moment it was applied, independently of any software release — so no new broken links can be
created, whatever happens next.

**Made silence impossible.** Every time the new code declines to write an instruction, it records
why, in a table the diagnosis system already reads. Previously this kind of decision left nothing
behind but a log line.

**Put it through the reviewer panel five times.** Twelve of fourteen reviewers now approve. The
panel earned its keep: it caught that the fix left the platform's real last line of defence
untouched (filed separately as bug 079), that we had copied a database rule instead of calling the
shared code that owns it (now fixed), and — indirectly — that the rollback recipe shipped with the
migration would have restored nothing at all, because it looked in the wrong table.

## Where we are now

- **The bug cannot recur.** Nothing emits these instructions at suggestion time any more, and none
  of the 27 already-emitted instructions can run again. That half is live.
- **The replacement is built but not yet running.** The new code is committed and compiled into an
  image; it takes effect at the next release. We verified against the running server — not the
  version label — that it is still on the old code today.
- **The existing damage is still there:** nine dead tool links on eight published pages across
  three sites. We measured them precisely and handed the list to the team that owns the broader
  broken-links work (bug 049). Most are repairable by correcting the address, because the tools
  themselves do exist — just at a different address than the guess assumed.
- **No formal approval stamp.** Three of the five review rounds were failed by a reviewer whose
  answer was lost in transit rather than by anything in the plan, so the run never converted to a
  formal APPROVED. We have not claimed the approval marker on any commit, because it would be a
  false record.

## Where we're going

1. **Next release.** When the chassis image rolls, check the running server carries the new code,
   then trigger a tool build and confirm the resulting instruction carries the tool page's real
   address. The exact commands are in the runbook.
2. **Bug 079** — the real backstop. The platform already spots dead links before publishing and
   publishes anyway. Deciding what it should do instead needs measuring first: making it block
   could stop pages shipping for reasons nobody has counted.
3. **The nine live dead links** — with bug 049's owners, not as a competing effort.
4. **Worth raising with whoever owns the reviewer panel:** at fourteen reviewers, one lost answer
   fails the whole round, and it happened three times in five. That is a compounding problem, not
   bad luck.
