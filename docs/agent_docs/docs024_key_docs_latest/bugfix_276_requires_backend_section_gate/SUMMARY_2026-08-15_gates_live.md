# SUMMARY — requires-backend section gate (bugs_open/276), 2026-08-15

**What we're trying to do.** A handful of website components need a real server behind the
site to work — a form that posts somewhere, for instance. One of these was already stopped
from being offered to sites that don't have a server, but only in its "suggest me a tool"
form. The equivalent everyday "plan me a page section" form had no such check anywhere, so a
plain static website could still, in principle, be planned this component and end up with a
form that silently goes nowhere for visitors.

**Where we've come from.** The gap was noticed and filed as a bug the same day the tool-side
check went live, at the reviewing council's own insistence — they'd approved the tool-side fix
but flagged that the sibling gap "still needed a concrete follow-up." The bug and the
platform's own design notes both pointed at one specific agent as the thing to fix.

**What we've done.** Before writing any fix, we checked properly rather than trusting the
one-call-site framing: we listed every place in the platform that could plausibly offer this
kind of component to a page plan, and found three, not one. The one everybody had named turned
out to be used only a couple of times a month; the one nobody had mentioned is used well over a
hundred times a month, including that same morning. Fixing only the named one would have looked
like a fix while leaving the real exposure untouched. We also checked: nothing has actually
broken yet — the one place this component is currently placed happens to be on a site that does
have a server, so no visitor has hit a broken page over this. We fixed all three call sites:
two get a proper "does this site have a server" check; the third, which turns out to be dead
code that has never once run, gets the component removed outright rather than a guessed check,
because there was no safe way to prove what a check on it would even be checking against.

**Where we are now.** All three fixes are small, self-contained database changes — no
software rebuild needed, live the instant they're applied. Each was proven correct against
real site data before being applied (comparing what the old and new versions would offer, for
both a server-having site and a plain static one), applied, and then proven correct again
against the live system afterwards. They've also gone to the platform's routine advisory
review process. The bug is closed; the platform's design notes have been updated to record
that the fix ended up covering more ground than originally scoped, and why.

**Where we're going.** One piece of the original design is still deliberately left for later:
a periodic check that would catch any case where this kind of mismatch slips in through some
route other than normal planning. We're not building that now — there's currently nothing for
it to find, and it's a different kind of change (it needs a software rebuild, not just a
database update). It stays on the record as the next piece if and when it's needed.
