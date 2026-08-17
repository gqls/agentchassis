# Where we are — bugfix 291 (plain prose, append-only, newest at the bottom)

## 2026-08-17 — picked up, understood, and a plan that had to change once already

The tool auditor is doing good work: it inspects the little interactive tools on our sites
and writes up genuine problems (missing labels on inputs, scripts that can never load, layout
that breaks on narrow phones). But when it decides a finding needs a human to look at it, it
addresses that finding to a reviewer called "hitl-review" — and no such reviewer has ever
existed. It was an idea written down in April that nobody ever built. The system tries to
deliver the item, finds nobody home, and stamps it "blocked" forever. Fourteen findings are
stuck like that, the pile grows daily, and — worse — while a finding for a page is stuck, any
NEW findings for that same page are silently thrown away.

What we found underneath: the auditor's instructions never say what state a review item
should start in, and the system's default is "ready to dispatch". One missing line of
config. The sibling route (findings the auditor can fix itself) works fine.

The fix has three parts, and the order matters because of a trap we nearly walked into. The
obvious tidy-up — "stop naming the imaginary reviewer" — would today make the software refuse
to file review items AT ALL, and the errors would be swallowed silently. So: first a small
config change that parks review items in the "a human should look at this" state (that works
today and stops the bleed); then a code change so the platform catches this whole class at
the moment of writing (any item addressed to a nonexistent agent gets parked as blocked
immediately, with a clear note, and it un-parks itself automatically if that agent is ever
built); and only after that code is live, the cosmetic rename. The fourteen stuck findings
get repaired — they are real findings and there is already a working button in the admin
that turns a confirmed review item into a fix task.

Also worth saying: the bug write-up we inherited had two small errors in it (about what the
existing parked items look like, and about a second producer being about to blow up the same
way). We measured, corrected them in the file, and the corrections do not change the fix.

A formal diagnosis run is in flight to double-check our reading before the code ships, and
the code change will go through the review council as usual.
