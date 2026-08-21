# SUMMARY — 2026-08-21: the search stops guessing

## What we're trying to do

When one part of the pipeline needs a value — the page being built, the code change that deployed a
job, the pages a new tool relates to — it asks a shared resolver. That resolver has a last resort:
if the value is not where it was told to look, it searches the entire job's data for anything with a
matching name. When it found two different answers, it picked one. **The aim of this workstream was
to make it refuse instead**, on the owner's ruling of 2026-08-15: no value at all is better than a
wrong one.

## Where we've come from

We could not simply switch the guessing off, because nobody knew what depended on it. So the work
was staged: build an instrument that records every conflict, run it, then remove the causes one at a
time until the remainder is understood — and only then flip.

Four causes came out at source: a search that ran for pages nobody had asked about; an undeclared
tie-break deciding winners by accident of insertion order; a retry payload echoing stale inputs; and
a genuine name collision where a page's *record* and a page's *name* were filed under one key. Each
was built, reviewed, shipped and proven live.

## What we've done

**All five steps are complete.** The instrument saw **19** distinct field/caller pairs over five
days. Eleven were closed by the first four steps. Two more turned out to be already wired. Three were
recorded as decisions that *absence is the correct answer*. The last three were live defects, each
fixed by a migration and **proven at the artefact, not inferred**: the page-type wire (515), the
tool-rerender declaration (512), and the deploy-commit wire (537, built jointly with another lane
that supplied the path rather than letting us guess it).

Then the flip itself: **committed `5fe010ada`, approved at council round 3.** Thirteen tests across
five files were deliberately changed and mutation-proved — reverting the one line fails all thirteen.

Along the way the instrument earned its keep twice more, catching bugs nobody was looking for:
`bugs_open/330` (an absent field replaced by an unrelated tool's data, on nine tools at one site) and
`bugs_open/350` (the resolver descending into a job's own configuration).

## Where we are now

**The flip is inert until the next build rolls** — it is program code, not configuration.

Three things are honestly open, and none is claimed as done:

- **The downstream blanking is untouched.** This change swaps a silently-wrong value for a silently-
  absent one, and an absent value still renders as blank with no error at fourteen of fifteen call
  sites. That is `bugs_open/342`, and it is not ours to close here. What the flip adds is that every
  refusal is *recorded* — field, caller, every candidate — the first time it happens.
- **The residual population.** 137 steps across 71 agents can reach this search; only 19 pairs ever
  conflicted. The rest have never conflicted in five days, which is not the same as never will.
- **One verification has no path.** Migration 512 is applied but unverifiable, because the work queue
  that would exercise it is drained. Recorded as *applied and explained, not demonstrated*.

## Where we're going

A **48-hour post-roll monitoring gate**, with its terms fixed in advance so it cannot be graded
loosely — demand control first, rows attributed to jobs by their start time rather than by the clock
(a job already running keeps its old behaviour for minutes), and any newly-conflicting pair traced to
its consumer rather than dismissed. Step 5 is not closed until that has run and been recorded either
way.

Then the small companion cleanup, now owned by a second session on this lane.

**What this workstream leaves behind**, beyond the change itself: an instrument that records what the
resolver does, a census of everything that can reach it, and a habit — visible in the review record —
of stating the check you did not do.
