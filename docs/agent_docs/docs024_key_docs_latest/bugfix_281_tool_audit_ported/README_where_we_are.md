# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-15 — what this is about

The owner looked at webdesign.co.uk by hand and found the Mind Map Studio tool with text you
could barely read and nonsense in it, and nothing in the system had ever noticed. The reason
turned out to be simple: the check that looks at tools only looked at tools built by the
framework's own tool generator. Sixty-three of webdesign's sixty-seven tools were brought in
from the old site as ready-made pages, all hanging off one shared "ported page" component,
so the check never saw them.

Worse, the two agents that act on tools (the auditor and the improver) find a tool by its
component. Point either at the shared component and it picks one of the 115 ported pages at
random. And the improver's fix goes back into the shared component's template — so a fix
meant for one page lands on every ported page on the estate. That is not hypothetical: it
happened on 5 August and again on 14 August. The second time it was triggered by a different
check (the acceptance one) that had already been widened to see ported tools without anyone
noticing that this made the improver dangerous.

## 2026-08-15 — what we did

Every tool is now identified by its page as well as its component, everywhere the audit
machinery touches it. The health check now sees all sixty-six auditable tools on webdesign.
For the ported ones it files a finding for a human to look at, deliberately, because there is
no safe automated way to fix a ported tool yet (the only fixer writes the shared template).
The auditor and improver now load exactly the page instance they were sent. And the write
that caused both incidents is now refused unless someone explicitly turns it on for a genuine
fleet-wide template change.

We also found, while checking, that the shared ported-page template is currently sitting in
the state the 14 August incident left it — the improver's output for one tool, with all 115
pages flagged for a rebuild. It has NOT reached any live page yet. Putting it back to the
plain pass-through is a decision for the lane that owns those pages; it is written up in the
bug file and the register so it does not get lost.

On the owner's other question — should we decompose all sixty-three ported tools into proper
framework components — the answer is "yes, eventually, and not yet". The machinery to do it
exists and worked this week on another site, but a decomposed page currently cannot be
rebuilt (open bug 204), and the same machinery, as run this week, actually made those tools
LESS visible to the audits because it did not mark them as tools. That is written up as a
proposal with the preconditions, for the owner to decide.

Next: council review of the code, commit, apply the two config changes, and the Go rides the
next fleet build. Verification queries are in the RUNBOOK.

## 2026-08-15 (later) — the review came back approved

The council approved the change (round two — round one failed on a formatting slip of mine in
the submission, not on the work). It raised six advisory points, all reasonable, and I acted
on them rather than argued: the new safety check now lives in the shared "component write
guard" file where the other such checks live; it only hard-stops for shared non-tool
components (a normal tool fix that hits a database hiccup carries on as before); I listed
every piece of code that can rewrite a component template and found one more that could in
principle be pointed at a shared component from a single page — recorded, not changed, because
it is also the thing that repaired the last incident; and the register entry now says exactly
what would have to be built for the "human review" items to stop being a permanent pile.

Where this leaves us: the two config changes are live now; the code changes ride the next
fleet build. Once that has rolled, one discovery pass on webdesign should show roughly
sixty-six tools examined instead of four, with the Mind Map's colour problem among the
findings — the queries to check that are in the runbook. Two things remain yours to decide:
whether to put the shared ported-page template back to its plain pass-through (it is still
carrying the 14 August rewrite, safely un-propagated), and whether to start the lane that
decomposes the sixty-three ported tools properly (the proposal lists what has to be true
first).
