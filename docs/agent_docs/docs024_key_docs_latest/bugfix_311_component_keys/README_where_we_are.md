# README — where we are (bugfix 311: the missing calculators)

Append-only, newest at the bottom. Plain prose for the owner.

## 2026-08-19 — picked up on your instruction; the diagnosis got sharper and the fix got smaller

You asked for the missing-tools bug (311) to be fixed properly, framework-wide. Here is
where it stands after the research pass.

The bug is still live — the same seven calculator sections failed again through last
night, and nobody else is fixing it (the lane that found it deliberately waited for your
say-so and a council round; this is that).

One important correction to the bug file's story. It said the components can't be reused
because the selector looks them up by a column (`section_type`) they don't have. That's
only half right. The build actually tries the right lookup first — by the component's
function name, which these rows DO have — and finds them. It then **throws them away
because they look broken to it**: they are hand-written calculator templates shaped like
tools, and the build's "is this template whole?" check expects section-shaped templates,
so it reads them as truncated and moves on. That's why it then asks for a brand-new
component, and why the new component collides with the old one and gets refused, forever.
I've put this refinement through the diagnosis loop to be independently checked before
building on it.

Why this matters: the bug file's second fix idea — "fill in the missing section_type
column" — would not have fixed anything, and for these particular rows would have made
the failure quieter and worse. So the fix is the first idea from the file, and only that:
**when a site's build wants to write a component whose name is already taken by a
component other sites depend on, it must stop trying to overwrite it and instead create
its own, under a name scoped to that site.** The new component is created in a way the
library can find and reuse — so the next site that wants the same calculator reuses it
instead of failing. One site can then never block another site's build again, which is
the property that actually scales to the 50-site programme.

What I'm deliberately NOT doing in this round, and why: (a) not touching the three
hand-written calculator components — they belong to loanandmortgagecalculator.co.uk and
repairing their shape is that lane's work, through the framework; (b) not building the
"refuse to deploy a page with missing sections" gate — that's a real gap (it's how this
shipped silently) but it's a separate change to a different part of the pipeline and
deserves its own round rather than riding this one.

Next: council round on the plan, then the code, tests, and a commit so it rides the next
chassis build. Verification will re-run one of the failed calculator sections on
loanzy.uk and check both halves: the new site gets its calculator, AND the old site's
component is byte-for-byte untouched.
