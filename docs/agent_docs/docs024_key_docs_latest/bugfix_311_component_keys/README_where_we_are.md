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

## 2026-08-19, later — the fix is written, tested and committed; the council is reviewing it now

The code went in as one commit (`17d883333`). What it does, in one sentence: when a site's
build tries to write a component whose name is already held by a component other sites
depend on, it now creates its own site-named copy that the library can find and reuse,
instead of failing forever — and the old site's component is never touched.

The tests prove both directions: the colliding case creates the new copy (and we checked
the test really bites by deliberately breaking the fix and watching the test fail), and
the normal case — a site regenerating its own component — behaves exactly as before.

Two hiccups worth knowing about. The independent check I sent through the diagnosis loop
failed because the fleet's AI budget cap was hit mid-morning; the cap lifted within the
hour, and I re-armed the check to run again on its own. And while I worked, two other
sessions were editing the same shared record files — their finished entries rode along in
my commit, declared in the commit message, which is how this tree handles that.

The council review of the plan is running as I write this. If it approves, nothing more
is needed until the next chassis image rolls — then the verification recipe in the
RUNBOOK proves both halves on the real loanzy.uk case. If it asks for revisions, that's
the next piece of work.

## 2026-08-19, end of session — the council APPROVED it, first time through

Twelve reviewers looked at the plan; it was approved on the first round with four advisory
notes and nothing severe. Most of the notes were "you asserted X, show it" — and each X has
now been measured and written down: the naming helper exists (the code compiles and the
tests pass against it), exactly one workflow in the fleet calls this action, the race two
simultaneous builds could cause resolves itself loudly rather than silently, and the reason
this fix differs in shape from the neighbouring tool-level proposal is now demonstrated
from the live database schema, recorded in that proposal's own file so the two are tracked
together.

The independent diagnosis check I re-armed earlier is running now. Nothing further is
needed from anyone until the next chassis image rolls; then the RUNBOOK's recipe proves the
fix on the real case. The bug file's status line now says exactly this.
