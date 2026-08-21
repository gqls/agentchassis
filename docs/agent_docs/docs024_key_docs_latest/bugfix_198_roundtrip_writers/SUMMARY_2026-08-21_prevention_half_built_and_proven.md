# SUMMARY 2026-08-21 — bug 198: the prevention half, built, approved and proven live

Written to be read aloud. First summary for this lane, at the point the prevention work was
finished and demonstrated rather than argued.

## What we're trying to do

Stop sites losing their stylesheets. A site's `styles.css` had been replaced by a few hundred
bytes of leftover patch rules on nine live sites across three separate waves, each time
silently, each time with every internal signal reporting success. Other lanes had spent the day
repairing the damage. **This lane's job was to make it stop happening** — not to restore
another site, but to close the doors.

## Where we've come from

The bug was filed on 4 August, when a fix agent's very first run replaced a 25,816-character
stylesheet with 149 characters. That was diagnosed as the model returning a fragment instead of
the whole document, and fixed properly on 5 August: the model stopped carrying the stylesheet
at all, and the database write became an append, which cannot shrink.

That fix worked exactly as built — and the sites kept being destroyed. Twice more, on 18–19
August and again on the morning of the 21st. The reason took three lanes and two weeks to see:
**the file has two authors who never read each other.** The design agent builds it from the
site's palette and layout records and never touches the stored copy. The patch agent appends to
the stored copy and then publishes that copy over the file. And new sites are created with the
stored copy deliberately empty. So on most sites the patch agent was appending to nothing and
publishing the result — an append that cannot shrink, converging on the wrong answer.

Worse, it fed itself: the contrast checker measured the ruined page, filed more faults, and the
system routed them straight back to the agent that caused the damage. One site absorbed eleven
in eight minutes, every one recorded as a success.

## What we've done

Four changes, in the order of how much they close:

1. **The patch agent refuses to work from a bad starting point.** The old check asked "is there
   a stored copy?" — and an *empty* copy passed, which is the hole all three waves went through.
   It now requires a plausible size and sole ownership. The size threshold was derived from a
   fleet census rather than chosen: healthy copies measure 13,650–26,917 bytes, every damaged one
   ever seen is under 2,400, and the line sits in the empty gap between.
2. **Refusals and failures stopped being recorded as successes.** Every ending in this workflow
   was labelled a success, so the system marked the work "complete" whatever had happened. All
   seven non-success exits now record the truth first. This was the quiet one: it means every
   count of outstanding faults taken before today was a floor, not a total.
3. **The design agent now writes the stored copy** it never wrote. This is the change that fixes
   the cause. It also matters retrospectively: every repair anyone made for this bug had a hidden
   expiry — the next design run on that site would silently undo it, roughly weekly.
4. **A size guard at the final publishing step**, available to any agent, switched on for this
   one. It lives in the component that can actually see the file being replaced.

Along the way the two sites that shared one stylesheet record — the case no repair could fix,
because they serve different files — were given a record each on the owner's ruling, carrying
their design and their page furniture across unchanged.

The council reviewed the platform changes and approved them at the first round. Its most useful
objection corrected our reasoning rather than our code: we had used one rule to answer a
different rule, and the correction is recorded where the reasoning was, not just where the fix
was.

## Where we are now

**Every site in the fleet — 22 of 22 — is now in a state where this bug cannot fire**, up from
19 of 22 this morning. There are no empty and no shared stylesheet records left.

**The fix is proven, not asserted.** We watched it work on a live site: a real job, through the
real queue, into the live agent, on a site that genuinely had the fault. It refused, recorded
the refusal honestly, and — the only check that really matters — **touched nothing**: the stored
copy was unchanged afterwards and the published file was byte-for-byte identical with no new
commit. We chose a site whose stylesheet is served to nobody, so that a failure of the guard
could not have reached a visitor, and we held a copy of the file throughout.

One half is written and tested but **not yet running**: the publishing-step size guard needs the
next software release. Everything else took effect the moment it was applied.

## Where we're going

Three things are outstanding and named rather than folded into the good news:

- **Confirm the publishing guard after the next release** — by asking the running programs what
  they contain, not by reading a log line, which for this guard would be close to unobservable.
- **A separate defect we did not touch**: the agent sometimes writes a fix aimed at something
  that does not exist on the page. Three sites' worth of evidence, its own job.
- **An inventory owed since 5 August**: which other workflows send a whole document through a
  language model into an unguarded writer. Nobody has that list, and this work did not produce
  it — the guard we built is for one seam, not for the class.

The honest summary of the state: **the damage is repaired, the cause is fixed at its source, the
guard is proven, and the remaining risk is a known list rather than an unknown one.**
