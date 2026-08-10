# SUMMARY — 2026-08-10 — the code index now says what it cannot see

Milestone read-out for `bugs_open/223`, phase 1 complete. Written to be read aloud.

## What we're trying to do

We keep a shared file of traps — things that will quietly mislead you the moment you touch a
particular file, table or command. A small agent checks those notes for us: it looks up the
files and symbols each note names, and records a verdict where the review council and the
next session can read it.

We want that agent to be trustworthy in a specific, narrow sense: when it says a note has
gone stale, that should mean something was checked and found to have moved. The point of the
whole mechanism is that a "stale" verdict is the signal someone uses to delete a warning.

## Where we've come from

It was not trustworthy, and the way it failed was hard to see. The agent looks things up in
an index of code symbols that contains **only Go** — 5,837 entries, every one a `.go` file.
Most of our trap notes are not about Go: they are about scripts, migrations, database
tables, config values. So the lookups came back empty, and the agent wrote that emptiness
down as *"this does not exist"*.

The clearest example, and the one that got the bug filed: a verdict declaring that three of
our own scripts and a whole category of database rows did not exist anywhere. It was
produced **by** those scripts and stored **in** that category. It disproved itself on
arrival.

The person who filed it then corrected himself in the same file, and that correction is what
shaped the fix: the blindness is perfectly consistent, but the *conclusion drawn from it*
varies run to run. Four verdicts on identical empty input ranged from a careful "cannot be
verified" to a flat assertion of non-existence. Three in four were careful. You cannot tell
which one you are holding — so asking the agent to be more careful was never going to be the
answer.

## What we've done

We fixed it one layer below the agent, in the shared lookup that four different consumers
use — including the review council's own seats and the diagnosis loop, whose source comments
had already asked for exactly this.

The lookup now reads its own contents — which file types, which kinds of declaration — and
says what it cannot see. An empty answer for something the index structurally cannot hold is
now reported as *not answerable*, in words that explicitly rule out reading it as removed,
renamed or absent. All of that wording is computed from what the index actually contains, so
when we widen the index later the warnings retire themselves rather than becoming stale
prose.

Two further protections: every stored verdict now carries a machine-written line saying how
much of that round was genuinely checkable, which the agent cannot soften or omit; and the
route through the workflow forks so that a round confirming nothing cannot even be offered
the word "stale".

Along the way we found two failure modes the original bug report did not contain. The blind
spot can produce false *confirmations*, not just false absences — a folder listing that
looks generous while every file you asked about is invisible. And a text search for a name
can be answered by completely unrelated code that merely contains those letters.

It went through the review council: approved first round, with six advisory objections. Five
named real gaps and were closed the same afternoon — including one that was right that I had
guarded two of the three false-positive modes I had myself discovered. The sixth is a
governance question we have put to the owner as `RFC_022`.

## Where we are now

Live on the current build and proven on the case that motivated it. The note that was once
declared non-existent now comes back as *"the entire footprint falls outside the Go-only
index; existence and behaviour could not be checked"* — the same absence of evidence,
reported as an absence of evidence. Four notes re-checked; all four stored verdicts carry
the machine-written summary; and on a mixed note the parts that *are* checkable were still
confirmed by name, so we did not buy the quiet by checking less.

One honest qualification, which we have written into the bug file, the register and the
notes: of the two protections, **the second one has never actually fired**. Reaching it
requires a round that confirms nothing at all, and that turns out to be nearly impossible,
because a single accidental fragment match counts as a confirmation. The protection doing
the real work is the plainer one — every answer now states what it could not see. We ranked
them in that order when we built them and the evidence agrees, but a guard that has not
fired does not get the credit.

## Where we're going

The remaining half of the bug is that the index still cannot see roughly 1,170 declarations
in our own Go code — which is also why the diagnosis loop can find every *use* of one of
them and never the declaration itself. That is specified, sized and risk-assessed, and it
gets its own change and its own review round, because it widens what every diagnosis in the
system searches.

Beyond that, one clean follow-up belongs to whoever owns the checking agent rather than to
this work: it should not ask a text-search question about a file it can already see is not
Go. That removes wasted checks and lets the standing-down guard stand up when a note really
is unverifiable.
