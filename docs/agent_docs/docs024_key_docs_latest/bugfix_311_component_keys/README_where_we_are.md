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

## 2026-08-19, evening — both halves approved; the precondition is code-complete

Two developments since the last entry. First, an owner ruling arrived via the portfolio
lane: the calculator fix must cover BOTH writers — the section one already done, and the
tool-level one described in RFC_036 — as one logical change, and the ~50-site build wave
does not start until it lands. Second, both halves are now through: the tool-level fix is
built (a native rebuild of a tool the library already offers is now born as a site copy of
it, which is what the fork field means everywhere else, so the unique-name gate no longer
kills the build), tested the same way as the first half including deliberately breaking it
to watch the test fail, and approved by the council on the first round — as was the first
half this morning.

Where each half stands: the section fix is LIVE in the fleet (proven at the running binary
with controls, not inferred from tags) but has had zero real exercises — the loanzy lane's
next clean-domain build is the agreed real test, with the old site's checksums pinned
beforehand so we can prove nothing was overwritten. The tool fix rides the next chassis
image. One follow-up is tracked in RFC_036 rather than left as folklore: the two different
ways a site can acquire a tool copy don't yet recognise each other's copies, which today
fails loudly rather than silently if it ever happens.

One process slip recorded in WRONG_CALLS: a commit message claimed a document edit that had
actually failed; the next commit corrected it and says so. The bug stays open until the
real-world test passes and a roll carries the tool half.

## 2026-08-19, evening — the real-world test ran, and the fix did exactly what it was built to do

Nobody else had run the real test (the loanzy team had no new domain lined up; the portfolio
sites are frozen under your halt), so I drove it from here on loanzy.uk, the site that first
showed the bug. I picked the car-finance calculator: it had failed three times at the exact
wall this fix removes. (The credit-health one fails for a different reason — the AI's answer is
too long for the limit — so it could not have tested anything here.)

What happened, in order. The build asked for a car-finance calculator. The AI chose the same
name as the existing calculator that belongs to loanandmortgagecalculator.co.uk — the
collision we were waiting for. Instead of failing, the platform noticed the name was taken by a
component another live site depends on, created loanzy's own copy under a site-specific name,
filed it in a way the library can find and reuse, and wrote a record saying exactly what it did
and why. The work item completed first time, no error. And the other site's calculator — all
eight of its calculators, in fact, which I fingerprinted before starting — is byte-for-byte
untouched.

One thing I got wrong earlier and have corrected in the notes: I had said the parked failures
would "converge by themselves" once the fix was live, because a background check keeps
re-flagging the pages. The flag is real, but nothing acts on it — a page only gets rebuilt when
something actually files a rebuild job for it. So the fix makes the next build succeed; it does
not start one. For loanzy that means each of the seven calculator pages needs a rebuild job
filed (or the loanzy lane's full rebuild, which files them all). I have filed one now, for the
car-finance page, to prove the last step: that the page really picks up the new calculator and
serves it. The page as served right now has zero input fields — that is the "before".

## 2026-08-19, later that evening — the page picked up its calculator; the test is passed end to end

The rebuild I filed for loanzy's car-finance page ran through cleanly. The page now carries the
new calculator as its second section, and the live page at loanzy.uk/tools/car-finance-calculator
has gone from a page with no controls at all to one with four — price, deposit, term and
interest-rate — with its own script loading. No stray template code on the page, and the other
site's eight calculators are still exactly as they were before I started.

So for the section half of this bug: fixed, live, and now proven on the real case it was built
for, with the "before" measured so nobody has to take the "after" on trust.

Two things keep the bug open, and neither is a fault in the fix. First, your ruling: the
calculator fix is a pair, and the second half (the tool-level writer) is written and approved but
is still waiting for the next chassis image — nothing to do until that rolls. Second, loanzy's
other six calculator pages are still hollow; each needs its own rebuild filed (the recipe is
written down), or the loanzy lane's full rebuild does them all. I stopped at one on purpose —
that is their site and their planned run — but it is a ten-minute job per page if you'd rather
it just happened. One of the six (credit-health-check) will still fail for an unrelated reason:
the AI's answer is longer than the limit allows, which is a separate small bug to raise.
