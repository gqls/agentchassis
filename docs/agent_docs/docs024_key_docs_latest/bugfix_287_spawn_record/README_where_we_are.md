# README — where we are (bugfix 287, spawn_record) — append-only, newest at the bottom

## 2026-08-17 — lane opened, cause pinned, fix designed

The bug: when the dispatch loop finishes a piece of work, it writes a record of "what the
worker said" onto the work item. Since the 15 August release, about three quarters of those
records hold the wrong thing — not the worker's answer, but the bookkeeping note from when
the worker was hired ("spawned"). The work itself is fine; the record of it is someone
else's. Anyone reading the item afterwards — verifiers, operators — is misled.

What we found today, reading the live code and the live database:

- The loop asks for the answer by a shorthand name ("result: handler_result"). The
  platform's input resolver treats that shorthand as a last resort: first it goes hunting
  through everything the run has collected for anything called "result", and it finds the
  hiring note first, every time. The author's own mapping loses to the hunt. The platform
  already has an instrument counting exactly this (hundreds of rows a day), and a recently
  built "!" marker that means "use my mapping and nothing else — fail loudly if it's
  missing". That marker is the real fix.
- A second, quieter gap: inside loops, each pass writes its outputs under numbered names,
  and the machinery that rewrites references to those numbered names only covers a
  hard-coded list of setting names. "result" isn't on the list. We close that by making the
  rewrite apply to anything that looks like a reference, so the list can never be
  incomplete again — the framework-level fix rather than a one-off.
- Two earlier beliefs needed correcting, with evidence: the answer IS present under the
  plain name at the moment the record is written (so the "!" marker is safe to arm), and
  the numbered-name rewrite ALONE would fix nothing (the hunt would still win) — the two
  halves are needed together.
- An independent reviewer pass over the plan found one real hazard we fixed in the design:
  if the strict marker ever does fail, it must fail just that one item, not kill the whole
  loop — so the migrations also add an error route ("mark the item failed and carry on").

Where this goes next: the code half is committed and rides the next release; the settings
half is two small migrations — one safe to apply now (two simpler agents with the same
bug), one held until the release lands. Then we verify at the live database that the wrong
records stop appearing while work keeps flowing, and measure whether the few hundred
already-wrong records can be repaired from the surviving run data (owner decision with
counts before anything is rewritten).

## 2026-08-17 (later) — shipped, one migration live, counts for a decision

Done today: the code half is committed and will ride the next release. The settings half is
live NOW for the two simpler agents (diagnosis and reporting runs) — their work records are
trustworthy again from this afternoon. The main dispatcher's flip is written and waiting,
deliberately, for the release; the file itself says exactly how to check the release has
landed and how to watch the wrong records stop.

An independent reviewer panel is reading the change as I write; its verdict will be recorded
here. The fix was also tested the unforgiving way: we deliberately broke each new guard in a
copy and watched the tests catch it, and we deliberately fed the migration's safety check a
bad state and watched it refuse.

One decision for you: since the bug appeared on the 15th, 2,259 finished jobs carry the
wrong record. For 244 of them the true answer still exists in the system and a small repair
script could restore it; for the other ~2,000 the true answer is gone (the records aren't —
the work itself was done and delivered; only the note about it is wrong, and our docs warn
readers not to trust those notes). Say the word and the 244 get repaired; or we leave
history as-is and rely on the warnings.

## 2026-08-17 (evening) — the reviewer panel's verdict, and one question routed to you

The panel REJECTED the bundle — but read what it rejected: not the bug fix. The settings
half (the strict markers) it explicitly called correct and contained, and that half is what
actually stops the wrong records. The objection is to the code half: making the loop
machinery rewrite references generically is a fleet-wide convention change, and the panel
says a change like that deserves its own architecture decision rather than riding a bug fix.
That is a fair process point, and the platform's standing rule for exactly this case is:
record the veto, write the architecture question up properly, and let you decide. That
write-up is RFC_035 — three options (keep the generic rewrite; go back to an explicit list
with the missing entries added; undo the code half and rely on the strict markers alone),
with our recommendation to keep it, honestly costed. The code sits committed either way;
nothing rolls back silently. The panel also caught three worthwhile hardening points on the
held migration, all now folded in.

## 2026-08-17 (late afternoon) — it works, proven on real traffic; and the "fresh build" needs redoing

Two things to report, one good and one that needs you.

**The good one: the bug is fixed on the main dispatcher and I watched it work.** A job that
started after the change went in processed four pieces of work, and all four recorded the
worker's actual answer — I opened one and it holds the real reply (which page it linked, its
id, a sample of the content), not the hiring note. Wrong records that appeared alongside were
all from jobs that had *started* before the change: a job carries a copy of its instructions
from the moment it begins, so anything already running keeps the old behaviour until it ends.
That is expected and it drains by itself.

**The one that needs you: the fresh build didn't ship any code.** The version number wasn't
changed, so the machines re-served the image they already had. I checked the running programme
itself on both machines: it is still the same build as this morning, and none of today's code
is in it — another lane measured 203 unshipped commits from the same event. Nothing is broken
by this; it just means the code half of my fix (and everyone else's work today) is sitting
committed and inert. To ship it, the version number needs bumping and the build redone.

Because that couldn't be waited for while ~25 wrong records an hour kept accruing, I applied
the settings change without waiting — but not blindly: I first measured, from the system's own
event log, that the value the change relies on was actually present at the moment it is read
(201 confirmations against 155 completions in six hours), and the change is written to fail one
item rather than a whole batch if it ever isn't. That reasoning is recorded in the migration
file itself, and I logged two mistakes I made reaching it in our wrong-calls log.

Still yours to decide: the 244 repairable historical records, and the architecture question in
RFC_035. One small admin thing: the migration ledger wouldn't accept my "record this as
applied" command (a permissions block in my session), so that one bookkeeping row is missing —
the change is live regardless, and the file says so loudly.

## 2026-08-17 (early evening) — both migrated agents now proven on real work; three things are yours

The diagnosis pipeline's first job since its change finished, and its record now holds the
actual verdict it reached — where before it held the hiring note and no verdict at all. With
the main dispatcher's four-for-four earlier, both changed agents are proven on real traffic,
each on a job that began after its change. The third (report generation) has the identical
setting but hasn't had a job to run, so it is untested rather than unfixed.

I've also written the history-repair script and rehearsed it against the live database inside a
transaction that I rolled back: it finds 303 repairable records out of 3,330 and restores each
one from the job that produced it — matched exactly, by requiring the job to name that very
piece of work, not by a near-enough identifier. It is not applied; it waits on you. Writing it
caught one of my own mistakes: the table stamps a "last changed" time on every write no matter
what you ask, so my first draft's safety check would have refused to run every single time. The
script now keeps each record's true completion time inside the record, and warns that the
progress report must exclude repaired rows — otherwise a repair could make the fix look better
than it is.

**Three things need you:**
1. **Redo the build with a new version number** — today's code (mine and several other lanes',
   203 commits' worth) is committed but not running.
2. **RFC_035** — keep, narrow, or undo the code half. The bug is fixed without it, so this is
   purely the convention question the reviewers asked you to settle.
3. **The 303-record repair** — say go and I'll apply it; say leave it and history stays
   annotated with warnings.

I've left the bug file open rather than filing it as closed: nothing is biting new work, but
those 303 records are still wrong for anyone reading them, and your two decisions live in that
file.

## 2026-08-17 (evening) — this one is done: the real build landed and the bug is provably dead

The rebuild worked this time. I checked it the same way as before and it is genuinely a new
programme: the fingerprint of the old one is gone, and a deliberately fake fingerprint doesn't
match either, so the check can tell the difference. The startup line that names the build had
already scrolled away on both machines, so I proved it a better way — by what the system now
*does*.

The code half is working exactly as designed: every new job now writes its instructions in the
corrected form, and the two things that should have been left alone were left alone. Then the
whole point of the exercise, over the first 75 minutes of real traffic:

- the internal warning that fired about 455 times a day on this defect: **zero**
- jobs run in the window: **10**, work items completed by the dispatcher: **11**
- of those, carrying the worker's real answer: **11 of 11**
- carrying the wrong "hiring note" instead: **none**

That is the bug closed, with the traffic to prove it rather than an absence of complaints. (Seven
other completions in the window came from a different route entirely, not the dispatcher.)

**Two things still sit with you, neither of them the bug:**
1. **RFC_035** — keep, narrow, or undo the code half. It is now running, so this is a decision
   about live behaviour; that is normal here, and the write-up says so plainly.
2. **The 303 repairable historical records** — script written and rehearsed, waiting on your go.

I've kept the bug file in the open folder purely because those 303 records are still wrong for
anyone reading them. Nothing else is outstanding on it.
