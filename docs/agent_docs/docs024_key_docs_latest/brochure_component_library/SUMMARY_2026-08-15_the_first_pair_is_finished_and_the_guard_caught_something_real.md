# SUMMARY 2026-08-15 — the first duplicate pair is finished, and the protection we built caught something on its own

*A new file, not an edit of `SUMMARY_2026-08-14b_the_cleanup_is_half_the_job_we_thought.md`. The
series is the record. This one is written because two things changed that yesterday's read-out
could not have said: a pair went all the way through, and a fix that had never fired, fired.*

## What we are trying to do

Some pages on our sites exist **twice**, under two different names and two different addresses —
a plain one like `/gripper-payload-calculator.html` and a `tool-` one like
`/tools/gripper-payload-calculator/index.html`. Both are real, both are published, and the
machinery has been maintaining both copies. There are seven such pairs across four sites.

Two halves to the job. **Stop new ones being made** — that is built, approved and live. **Clean up
the seven that already exist** — that needs a human decision per pair about which side survives,
because it is a content and search-visibility question, not something a machine should pick. The
owner made all seven decisions on 13 August and revised two of them on the 14th.

## Where we have come from

The clean-up looked mechanical: eight steps per pair, mostly database edits. It has been resized
twice, both times by trying it.

**Pair 1 taught us the procedure was wrong about its own safety.** Two things fell over. There is
**no redirect mechanism anywhere in this platform** — the step that said "write a redirect so the
old address forwards to the new one" writes to a table nothing reads. Retiring an address 404s it,
full stop. That finding sent one of the owner's decisions back to him, and he reversed it. Then
the platform **refused** to retire pair 1's page, correctly, because three live things still
linked to it — and the check we had been relying on to say "nothing links here" turned out to be a
table that is **empty across the entire estate**, so it had been answering "nothing links here"
about every page in the world.

**Then the size of the remaining job got measured properly** — read-only, by reproducing the
platform's own link census in SQL rather than mutating six live pages to ask a question. Three of
the six remaining pairs need no content work at all. That falsified the previous day's gloomy
extrapolation from a single example.

## What we have done

**Pair 5 — robot-hands' gripper payload calculator — is complete, and it is the first one to
finish.** All eight steps: taken out of the site's plan so the machinery cannot re-create it, its
nine queued jobs cancelled, the row retired, the file deleted, and verified at the live site.
The three database changes went in as one all-or-nothing transaction with its own assertions, and
an exact undo recorded before anything was touched.

**The acceptance test has two halves and we passed both.** The first — the address returns "not
found" immediately — is weak, because it passes even when nothing works. The second is the real
one: it must **still** be gone after the site republishes itself, which it does twice a day, since
retired pages coming back is the entire reason this bug exists. It is still gone. And we can show
the republish actually happened: two unrelated pages on that site each grew by exactly 123 bytes
overnight and the surviving page has a fresh publish stamp. **A durability test that cannot
demonstrate the event occurred is passed by a system that did nothing.**

**Separately, the protection built last week fired for the first time — twenty times, unprompted.**
It refused attempts to republish three retired robot-hands pages, from two different parts of the
system, at both of the two points it was designed to cover. Until yesterday it was correct code
that nothing had ever tested. Nobody staged this; it was found while re-checking counters after a
routine upgrade.

**And that exposed a second, unrelated problem.** Two of the refused attempts were the routine
tidy-up sweep trying to fix stray formatting **on retired pages**. It tried three times each,
doing real work each time, then recorded the failure as *"the fix didn't work"* — when the fix had
worked and the protection had correctly refused to publish it. Anyone reading those records would
conclude the formatting fixer is broken. It is not. Written up where the bug lives, with the fix
options costed, and left for whoever owns that sweep.

## Where we are now

**One pair of seven is finished. The procedure is proven end to end and needs no more design.**

The other six are all blocked on the same class of thing rather than on mechanics. One needs two
pieces of body text repaired before its page can be retired. One is held behind an unrelated bug.
Two belong to another workstream working the same site, and have been handed over with the
measurements done. One needs four links repaired, including both the site header and footer. One
needs roughly 1,700 words merged from the page being retired into the page being kept.

**That last point is the shape of the whole remainder: it is writing, and the framework writes
content, not a session.** That is a standing ruling here and it is the right one — hand-authored
content silently opts out of every check the pipeline applies.

The archived-page protection now meets every bar we have for being called finished, and has been
**deliberately left open**, because two routes remain by which the original problem can still
happen. Closing it would retire a ticket whose defect two open doors still admit. That is recorded
as a decision so the next person does not read it as an oversight.

## Where we are going

Nothing here needs a decision from the owner. The next moves, in cost order:

1. **The two fundamentally.ai pairs** — decided, measured, one of them with zero blockers, already
   routed to the workstream that owns that site's execution.
2. **Pair 1's two editorial repairs** — a site footer, which is a stored artefact with its own
   known trap, and one article body.
3. **Pair 7's content merge**, then its retirement. The most valuable and the slowest, because the
   text has to be written properly.
4. **Pair 6**, the most expensive: four links including both header and footer.
5. **Pair 2** stays held until its blocking bug is fixed.

The thing worth carrying out of this week is not any single fix. It is that **three separate
"everything is fine" answers this month came from instruments that could not have said anything
else** — an empty table, a durability check with no proof the event happened, and a counter with
no demand behind it. Each was caught by asking the same question first: *could this have come out
differently?*
