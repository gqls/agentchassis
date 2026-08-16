# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-15 evening — picking up the "other 285"

The fix the 281 lane built for this bug is committed and approved by the council, but it is
not running yet: the chassis that is live was built six hours before that code was committed.
So until the next release rolls, the same accident can happen a third time. Nothing we can do
about that from here except not delay the roll.

Two things in the bug file turned out to be wrong, and one of them mattered. The file said no
page had ever served the bad markup. One had. When the tool-improver "fixed" the asset formatter
on 14 August, it also chose — at random, because it looked up the shared component and took the
first page it found — a learn article called "AI builders: content first", and re-rendered that
page from the poisoned template. Since then that page has been live with an empty article and a
"Related Downloads" box offering three files that do not exist. I put the original article back
(the platform had archived it automatically when it was overwritten, and the archived copy
matches the import fingerprint byte for byte) and queued a plain re-deploy. I still need to
confirm the live page is clean.

The other wrong thing: the earlier decision not to auto-fix ported tools was justified partly by
"no tool has a plan to judge it by". That was counted in the wrong table — 87 tools do have a
current plan, including 14 of the 63 ported webdesign tools. It does not change the decision (there
is still no safe way to fix one ported page without touching the others), but it does mean the
missing piece is exactly one thing: a fixer that writes to the one page instead of the shared
template. That is what I intend to build next, and whether ported findings should then be routed
to it automatically is your call.

## 2026-08-16 morning — the roll landed, and this bug is closed

The new build is running the fence. I did not take that on trust: I sent the platform a deliberate
"rewrite the shared template" request carrying exactly the bytes it already holds, so that if the
fence had somehow not been there nothing bad would have been written — and the request was refused
in under half a second, with the reason recorded ("this component is on 115 pages across 2 sites").
The learn article that had been serving an empty body and fake download links is back to its real
content on the live site; I checked the page itself, not just the job status.

I also added a permanent check that runs on every build: any piece of code that rewrites a
component template must either ask the fence or say, in writing, why its change is meant to reach
every page. I broke it on purpose two ways to make sure it actually fails, then put it back. It has
gone to the council for review (advisory).

One thing for you to decide, and I have written it up rather than built it: today a problem found on
one of the 63 imported webdesign tools goes to a human, because the only automatic fixer would write
to the shared template. The fix for that is a fixer that writes to the one page instead. It is about
a day's work, but it is only worth building if you want those findings fixed automatically — and I'd
suggest only for tools that already have a plan to be judged against (14 of the 63), starting with
one supervised run on the asset formatter. If you'd rather keep them as human-review items, nothing
needs building. The bug itself is closed either way.

Two small corrections to what you were told yesterday, both now recorded: the "recovery" on 8 August
was actually the poisoned version being archived (the recovery came from the regeneration that
followed), and that regeneration itself triggered 154 page re-renders and 73 attempted full rebuilds
of imported pages, all of which were refused by a different safety check. That is the "or worse"
scenario, and it has already happened once — we were saved by the last guard in the chain.
