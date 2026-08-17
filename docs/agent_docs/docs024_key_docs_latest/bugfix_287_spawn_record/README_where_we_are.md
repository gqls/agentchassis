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
