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
