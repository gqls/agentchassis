# SUMMARY — 2026-08-19 — silent row caps: the lane closes

*Written to be read aloud. Previous: `SUMMARY_2026-08-17_275_silent_row_caps.md`.*

## What we were trying to do

Some of our agents fetch a list from the database and hand it to a language model — "here are the tools
this site could have", "here are the pages you could link to". A few of those fetches quietly stopped
at a fixed number. The model then chose from a slice of the real options, and **nothing looked wrong**,
because a shortened list produces perfectly plausible answers. That is what makes it a *silent* cap
rather than a small one: there is no bad output to notice.

We set out to fix one instance — the tool-suggester, which was choosing from 30 of 74 tools, the first
thirty alphabetically — and to make the whole class visible wherever it occurs.

## Where we came from

The instance fix was straightforward: the tool descriptions were 80% of the payload, so we bounded
those and removed the row cap entirely, marking any description we cut so the model can see it was cut.
Both halves went live on 17 August after two council rounds.

The class fix was a detector: whenever a query returns exactly as many rows as its own limit, the
platform now warns, naming the step. That is deliberately only a *suspicion* — a result the size of its
ceiling may have been truncated — but it is the one place where every capped query in the estate passes
through.

## What we have done since

**We proved the fix at the thing itself, not at the configuration.** The old version showed the model
29 tools with nothing ranked past thirtieth alphabetically. A run this morning showed it **80 tools,
51 of them past that line**, right down to the last name in the alphabet. Those two measurements are the
strongest evidence available, because each could only have come out the way it did if the code was in
the state we claim.

**We found that the detector cannot be read.** It writes a warning to the running program's log, and
those logs are overwritten within **seconds to a minute** — I measured between 3 and 91 seconds on
machines that had been running for half an hour untouched. A warning that fires at two o'clock is gone
before anyone looks at one minute past. We tried once, properly, with a recorder attached in advance,
and still caught nothing. **You ruled that we stop chasing it, and that was right.**

**We found a better instrument, and it was already there.** Every one of those queries already writes
its results into the permanent record of the run. That record survives restarts and can be read back
over about two days. Using it, the caps stop being a suspicion and become a count: the news-feed job has
taken exactly five sites on every run for two days; the model-directory job, which has a limit of twelve
and only ever finds four things, has never hit its limit at all. The second number is what makes the
first one trustworthy.

**Asking who the caps cut turned up a defect we had explicitly written off.** We had recorded that a cap
on a work queue was harmless — take five now, the rest next time. True, but the queue is sorted
alphabetically, so the same names win every time: the five sites sorting first are **exactly on
schedule**, the four sorting last are **always late**, and the worst-affected is the one that asked to
be refreshed most often. That is now its own ticket.

**And proving our own fix broke something.** Showing the model nearly three times as many tools made its
answer longer, and the answer ran into a size limit nobody had raised — because the prompt also asked it
to justify rejecting *every* tool it was shown. That instruction alone was 37–66% of every answer. We
raised the budget and, more importantly, bounded the two lists that grow with the library. The
confirming run came back **under a third of its budget, and smaller than the pre-fix average**, despite
the much bigger menu.

## Where we are now

**The lane is closed.** The original bug and the regression it caused are both fixed, live and proven,
and have moved to the closed folder. The detector is live and its correct use — read the permanent
record, not the log — is written into the concept register, the runbook and the landmines file.

Three tickets are open and each stands on its own:

- **316** — the news-feed queue serves the alphabet, and separately is oversubscribed about two to one.
- **321** — about 72% of the tool-suggester's suggestions are silently thrown away by a collision in how
  work items are keyed. Found this morning; it is why widening the menu has produced fewer new tools
  than it should.
- **313 and 298** — the internal linker has never produced a single link in four months, because a check
  after its database fetch can never be true. Fix 313 first; 298 is meaningless until it is.

## Where we are going

Nothing here is on fire and all of it is cheap. The order I would suggest is **321 first** (it unblocks
value we have already paid for), then **313** (a capability that has never once worked), then **316**,
whose second half — how much the feed queue should cost — is a judgement about spend rather than a bug.

The habit worth keeping from this lane is smaller than any of them: **when you widen what a model is
shown, measure what it says back, and then check what happens to what it said.** Both of this lane's
follow-on defects were one query away the whole time, and neither was noticed until something failed
loudly enough to force a look.
