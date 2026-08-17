# SUMMARY — 2026-08-17b — it fires, and we have watched it fire

*(Supersedes nothing — the earlier summary today, "approved and live, never fired", was
accurate when written and is the record of where this stood eight hours ago.)*

## What we're trying to do

Stop a calculator on one of our sites quietly working out a number using a rule the law
changed years ago — which is what happened, for sixteen months, on stamp duty.

## Where we've come from

We keep a register of verified facts per site — every tax band and threshold, each with
its government source, re-checked every morning. It governed what a page could **say**. It
never governed what a calculator **works out**. This work connected the two.

This morning the machinery was approved and running, and had **never once fired**. That is
a genuinely uncomfortable state: a check nobody has watched work is indistinguishable from
a check that cannot work, and this estate has been caught by that before.

## What we've done

The team that owns mortgagecalculator.co.uk switched it on — you chose that over the
cautious option, and they declared all thirteen stamp-duty facts rather than the two I
offered, on the good argument that the tool genuinely depends on all thirteen.

Then it was tested properly, in three parts:

**It fires, and it reads the register at the moment it checks.** Two runs differing only
in the register's value came back reporting that difference. It also correctly found a
calculator that our acceptance machinery structurally cannot see — the exact case the
design was built to reach.

**It goes quiet once it has asked.** The first real run filed thirteen "please confirm this
tool matches" items. The identical run afterwards produced **nothing**. Same query, same
site, same figure — opposite result, and the only thing that changed was that the questions
had been asked. That matters because the alternative was thirteen items every morning
forever.

**It notices a moved figure.** With the tool's recorded value nudged away from the
register's, it produced exactly one finding naming the fact, the tool, the old value and
the new — and the other twelve facts stayed silent.

**Nothing was changed to prove any of this.** No live tax figure was touched, not even for
ninety seconds: the comparison works in both directions, so we moved our own bookkeeping
rather than the published number. Everything was restored and checked as identical
afterwards.

## Where we are now

The mechanism is finished and demonstrated. If a stamp-duty threshold moves, the register
notices overnight — as it always did — and now the calculator that encodes it gets named
the same day, to a person, on a page quoting tax law.

Three things I got wrong along the way, all in **instructions I wrote for other people**,
all found by the team using them: a dispatch script using a form that silently loses four
sends in five; a test recipe predicting a result its own code cannot produce; and, twice
corrected and still wrong, the reason the system would give. The code was right every
time. That asymmetry is the useful finding — my errors live in documentation, which is the
part other people actually execute.

## Where we're going

Nothing is owed on this mechanism. What is left is deliberately out of scope and written
down: the same blindness in ordinary page text rather than calculator code; the automatic
repair path, which exists but has never run and may be unreachable on today's sites; and
the one that matters most — this tells us a figure has **moved**, never that a figure is
**right**. If the register and the calculator are wrong the same way, they agree, and it
says nothing. That needs its own review before anyone builds it.
