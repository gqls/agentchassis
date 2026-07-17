# The concept register — where we've been, where we've got to, where we're going

*2026-07-17 (evening). The calm, read-aloud version — a companion to the
morning's `SUMMARY_where_we_are_2026-07-17.md`, picking up after the
bug-historian went live. For the technical detail see `PLAN_concept_register.md`,
`DESIGN_relevance_filter.md`, and the turn-by-turn `RUNNING_NOTES_concept_register.md`.*

---

## What we're doing, in one breath

We built a map of everything this platform has ever been described as doing,
checked that map against the real code, and are now using it to grow the panel
of reviewers that inspects every proposed bug fix before it becomes real. The
morning had one new reviewer live. By this evening there are three, plus the
machinery that will let us add many more without slowing the panel down.

## Where we've been

The short recap, for anyone joining here: this platform's history was scattered
across about four thousand documents. We distilled it into a register of
sixteen hundred-odd concepts, checked every one against the running code — about
one in thirteen turned out to be wrong, and we corrected those, with a second
independent check trying to disprove each correction before we trusted it. The
whole point was never the register itself; it was to grow the fix-loop's review
panel from its original two members into a proper council, choosing each new
member with evidence rather than a hunch.

## Where we've got to

Three new reviewers are now live, each chosen because the platform's own history
kept teaching the same lesson to people who didn't know it had been learned
before:

- A **bug-historian**, who recognises a bug shape the platform has been bitten
  by repeatedly — content silently vanishing during a rebuild, which we found
  had happened at least seven separate times in different corners of the system.
- A **reuse checker**, whose whole job is "have we already built this?" — born
  from a real incident where someone rebuilt a piece of machinery that already
  existed.
- A **guidelines reviewer**, who checks a fix against the platform's own written
  rules — and, cleverly, tells the difference between a fix that *breaks* a good
  rule and a fix that's correct but *exposes* a bad rule. The second kind isn't
  blocked; it's approved with a note that the rule needs fixing. We'd just found
  a live example of exactly that: a documented rule about model settings that
  turned out to be written backwards.

Along the way, a neighbouring team reviewing our work caught a genuine gap — the
new reviewers could ask for a fact to be checked, but those checks weren't
actually being run. We fixed it the same day. That's the council idea working on
itself: one set of careful eyes catching what another missed.

The bigger piece we finished today is the groundwork for scale. Adding reviewers
one after another has a cost: every reviewer looks at every fix, so ten more
reviewers would mean ten more inspections on even the smallest change. Most of
those inspections would be irrelevant — a typo fix doesn't need the model-settings
specialist. So we built a **relevance filter**: a small, fast, no-guesswork step
that looks at which files a fix actually touches and wakes only the reviewers who
have something to say about it. The engine for that is built, tested, and in
place, waiting on the next routine platform update to switch on — an update
another team is already leading, so ours simply rides along with it. Once it's
on, we can add specialist reviewers freely, and each one costs nothing until a
fix actually enters its area.

## Where we're going

The path from here is clear and mostly downhill. The relevance filter switches
on with the next platform update. After that, we add the remaining specialist
reviewers — a diagnosis-machinery guardian, a model-reliability specialist, a
compliance eye, and others — each one now cheap to add because it only ever
wakes when it's needed. And underneath all of it, the same quiet discipline
continues: keep the register honest as the platform keeps moving, so the map
never drifts out of date the way it had before we started.

The one-line state: the council has grown from two reviewers to five, the
machinery to grow it much further without slowing it down is built and ready,
and the next reviewers are lined up. Nothing is blocked; the next step happens
on its own when the next platform update ships.
