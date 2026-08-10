# SUMMARY — where we are, 2026-08-10

*Concept register. Written to be read aloud. Previous milestone:
`SUMMARY_where_we_are_2026-08-09.md`; the series is the record, so this is a new
file rather than an edit of that one. Written one day after the last, which is
unusually soon and deliberate: yesterday's read-out ended on an open question
about whether to build a gate, and that question is now answered, built and
measured.*

## What we're trying to do

Keep one place that answers **"does this already exist, and is it real?"** for
every mechanism, contract, agent, tool and idea in the platform. Not
documentation — documentation says how a thing works. The register says *that* it
exists, who built it, what it is for, what will mislead you about it, and whether
it is genuinely live or merely written down.

Two audiences, and the second is why accuracy is not housekeeping: a session about
to build something, who needs to discover it was built in June; and the council
seats that review our changes, whose prompts are seeded from these entries. A wrong
entry is not misleading prose — it is false evidence inside a machine review.

## Where we've come from

Built in July by sweeping ~4,100 files under `docs/`: 2,185 raw concepts extracted,
merged to 1,627, then every one checked against live code and database. Since 20
July it has had no dedicated thread; it grows because each session registers what
it builds, watched by a coverage check.

On **4 August** we found that model was not working. Thirty-four concepts had an
entry and no row in the index — the index being the thing people actually search,
so those read as *does not exist*. The same week a daily watcher was built and
deployed, and on **9 August** every stored count in the register was retired at the
owner's direction, because four different commands counted the index and all four
answers were individually correct.

Yesterday's read-out ended with the register complete, consistent and
self-monitoring — and still leaking. Concepts kept arriving with no index row,
roughly one every day and a half. The open question was whether to keep watching
or to intervene, and the reason given for waiting was good: *we now have a
mechanism that will tell us the rate instead of us guessing it.*

## What we've done

**We let it tell us the rate, and then we built the gate.**

`check_register_entry_without_row` is live in the pre-commit hook. When a commit
adds a concept entry without its index row, the commit says so — to the person
writing it, before it lands. A second arm catches an ID number another lane has
already claimed, which happened between two sessions three hours apart on the 8th
and cost a permanent explanatory note in the register. It is advisory, like every
check in that file; it never blocks. Registered as **OPP-006**, entry and index row
in the same commit — its own rule applied to itself, and the gate named it when the
row was briefly missing.

**It was measured against real history before being wired in**, which is that
file's standing bar. Over the fourteen days to 10 August: 398 commits touched the
register, 159 of them added at least one concept, **133 of those (84%) did it
correctly and the check is silent on them**, and 26 (16%) shipped an entry with no
row — 34 findings, no false positives. Against the whole commit stream that is
0.8%, quieter than checks already running.

**But the hit rate is not what settles it, and this is the finding worth
remembering.** A count of firings cannot tell a real leak from a two-commit habit
that self-corrects five minutes later. The waiting time can: **the median concept
sat 93 hours — nearly four days — before its row appeared.** Twenty-three of
thirty-two took over a day. Twenty-one were finally closed not by their authors but
in one sweep by another session's clean-up. That is a cost somebody else pays.

**We also corrected a number we had been quoting, including in yesterday's
summary.** The leak is about one a day, not one every day and a half. Nothing was
measured carelessly — **the daily watcher can only report what is still missing at
06:50**, so a row backfilled the same afternoon never enters its count at all. Two
such cases in the last week were invisible to it. A report's figure is bounded by
its sampling interval, and nothing in the report said so.

**And one mistake of ours, which is the more useful story.** The ID-collision arm
was *dead* when first written — not erroring, dead in the way that looks exactly
like working. A git command had its arguments in an order git rejects, and because
the tool reads a command's output and ignores its complaints, the refusal arrived
as "nothing found": no matches, no findings, no error, which is also what a clean
register looks like. It passed a sweep over four hundred commits of real history,
because that sweep exercised the one mode where the argument order happens to be
legal — and not the mode that runs when you actually commit. What caught it was
staging a case that *had* to produce a finding and refusing to accept silence. It
is now filed as a fleet-wide trap: seven of our twenty-two maintenance scripts
listen to output and ignore complaints the same way.

**Two loose ends from yesterday, checked rather than assumed.** The CronJob
re-deploy recorded as owed had already happened — the running job is byte-for-byte
the current script. And the one file still carrying a stored count is still another
session's half-finished work, last touched on the 8th, so it is live work rather
than abandoned; tidying our one line out of it would sweep their changes into our
commit. Left alone deliberately, for the third time.

## Where we are now

**1,817 entries, 1,817 index rows — they agree exactly**, and the two concepts the
watcher named this morning have rows written from their entries, neither claiming
more than the entry does. No ID is used twice. The drift check's only remaining
finding is that one stored count, which is owed rather than forgotten.

The class that kept pulling the two halves apart is now caught **at the moment of
writing** rather than reported the next morning. The daily watcher stays exactly as
it was — it is the instrument that proved the leak existed and measured its rate,
and it remains the only thing that can see drift arriving from outside a commit.

One honest caveat about today specifically: the branch has 65 commits unpushed
across all sessions, and the watcher reads the *pushed* branch — so tomorrow's
06:50 run may still name concepts whose rows are already committed here. That is
the watcher working as designed, not a regression.

## Where we're going

**Staleness, and now it really is the only open flank.** Coverage answers "is it
here?", drift answers "does it agree with itself?", the new gate answers "did both
halves ship together?" — and **nothing answers "is it still true?"** Verification
ran once, in July. A register status is already a known trap: a snapshot that
outlives its truth, quoted by council seats as ground truth. The building blocks
exist — `covers-through` stamps, the landmine-verifier, the bugs-open staleness
sweep — and pointing them at register entries is a design question that deserves
its own session rather than a corner of this one.

**The test we should hold the gate to.** Its own entry names it: does the daily
watcher's missing-row count actually fall to zero now? If it does not, the gate is
being ignored rather than working, and that is a different problem with a different
fix. Worth reading the daily row for a week before declaring this closed.

**And the thread running through the last three days.** Yesterday's lesson was that
removing a drifting artefact beats watching it. Today's is narrower and sharper:
**a measurement is only as honest as the thing that could have made it come out
differently.** The stored counts had three plausible rival answers. The watcher's
rate could only ever see the slow half of the leak. The collision check could only
ever return "clean". None of those were carelessness — each was an instrument
answering a smaller question than the one being asked, in the same confident voice.
