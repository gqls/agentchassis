# SUMMARY — component instance scope, 2026-08-17

Second in the series (previous: `SUMMARY_2026-08-16_…`). The milestone: the owner has ruled on the
programme's shape, the numbers under that ruling survived a hostile second look, and the machine
that executes the first three-quarters of the work is built and approved. What remains is a
different kind of work again — running it.

---

## What we're trying to do

Unchanged: make it genuinely safe to put interactive components — calculators especially — on a
page more than once. Today their element names are fixed text, so a second copy answers with the
first copy's numbers: a believable wrong answer on a consumer-credit site, not an error.

## Where we've come from

Yesterday's summary ended with the naming machinery live but inert, the council's approval in hand,
and a warning that the next phase was a different kind of commitment. Since then, three things
happened in sequence, each changing the one after it.

First, the job was measured properly. "Convert the 22 calculator templates" turned out to be **91
stored components on 94 live pages across 22 sites**. That number survived; almost nothing else
from the first sizing did.

Second, the sizing was wrong twice, and the second look you asked for caught both. A quick
pattern-match said about 30 components needed careful script work; the proper classifier said 88;
the truth is **25** — because the classifier itself had a blind spot. Our components conventionally
start their code with a documentation comment, and the checker looked at the first character to
decide "safely wrapped?", so sixty-two properly wrapped components read as dangerous. One opened
file exposed it. The checker is fixed, the fix went through review and came back **approved with no
objections at all**, and the 25 that genuinely need careful work are almost exactly the calculators
the bug was originally filed about.

Third, you ruled: the hybrid approach, loan-calculator site first, and everything run through the
platform's own work-tracking machinery rather than by hand — so every conversion is recorded,
reviewable, and reversible.

## What we've done

Built the deterministic converter as a platform action. It renames every element name a component
declares, and every reference to those names — including three kinds that break silently if missed:
form labels, style rules, and a kind we only discovered because the tests use real stored
components rather than examples I wrote: buttons that carry the *name of the element they act on*
as a data attribute, which the code reads at runtime. A composed example would never have contained
that; the real bytes did, and five copy-buttons per affected page would have quietly died.

The converter's most important property is what it refuses to do. A component whose script would
still clash after renaming is **rejected untouched** — because renaming alone produces a page that
passes every check while both calculators still run the same arithmetic. Half-converted is worse
than unconverted, so half-converted is unrepresentable through this path.

## Where we are now

Built, tested against real stored components, council-approved — and **not yet running anywhere**.
Today's redeploy restarted the machines but served them yesterday's cached software, because the
new build reused the old version number; the platform has a known trap there and we walked into it
the same day another team measured it. Nothing is lost: the code waits for the next properly
numbered release. Nothing has been converted; the original defect is still live; the bug stays open.

## Where we're going

After the next release: one canary conversion end-to-end through the framework — convert, re-render,
redeploy, diff the served page. Then the 66 mechanical components in batches, while the harder
pipeline is designed for the 25 calculators: an AI rewrites each script under the same acceptance
gate, with a size check on every result because the platform has once seen an AI truncate a
component and report success. Before any calculator converts, two bookkeeping steps you already
know about: rebaseline the byte-identical page check, and move the 170 test selectors in lockstep.

The first conversion that ships will trip a daily alarm we built for exactly this moment — that
alarm means "the architecture exception has expired, acknowledge the review", not "something broke".
