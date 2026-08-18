# SUMMARY 2026-08-18 — the instrument we were told to build caught its first real bug, and it was somebody else's

*Written to be read aloud. Previous summary: `SUMMARY_2026-08-15_the_backlog_starts_to_drain.md`,
which covers a different sub-thread of this lane (the placeholder image backlog) and is not
superseded by this one.*

## What we're trying to do

Deep inside the platform there is a piece of plumbing that decides, for every step of every
pipeline, where a piece of information comes from. Most of the time an author has said
explicitly: "this field comes from there." When nobody has said, the plumbing goes looking — it
searches the whole pile of data the run has accumulated for anything with a matching name, and
takes what it finds first.

That search is convenient and it is dangerous. If two different things share a name, it picks one
of them, and until this month *which* one was effectively a coin flip on every run. It had
already caused two production incidents. The owner ruled on it: the search may keep working when
every candidate agrees, but **it may never guess** — no field at all is better than a wrong
field. Our job was to carry that ruling out, in two deliberate stages: first make the thing
observable, then make it refuse.

## Where we've come from

The first stage shipped, and then the review council told us it was not good enough — for a
reason we had missed and they were right about. We had made the search log a warning whenever it
met disagreeing candidates. But a warning in this system's logs survives about ninety seconds
before it scrolls away, and the whole plan depended on watching those warnings for a week before
deciding anything. We had built a smoke detector with no memory.

So we rebuilt that half: every such event is now written permanently to a database table, one row
per occurrence, with the field name and every candidate the search considered. The council
approved it on the second round. Alongside it we shipped an opt-in mark an author can put on a
wiring line meaning "this exact source, or fail loudly — never guess", and switched the first
pipeline over to it.

## What we've done

Three things, in the ten days since the ruling.

**We made the invisible countable.** The recorder went live and immediately started writing. In
its first twenty-four hours it logged about sixteen hundred events — sixteen hundred moments when
the platform was choosing between look-alike fields, none of which anyone could previously have
seen.

**It found that eighty-six per cent of them were one already-known bug** — not ours, in the
dispatch loop that hands work to other agents. We did not start a competing investigation. We
took our measurements to that lane's own bug file, where they answered the one question their
diagnosis had left open: which of the look-alike values the search was actually picking. We also
found *why* the wrong one always won, which nobody had spotted: when a loop runs, each pass
leaves a numbered copy of every field behind, and the search sorts alphabetically, so the plain
un-numbered name — holding the previous pass's data — always sorts first. It wasn't bad luck; it
was arithmetic.

**Their fix shipped, and our recorder graded it.** The specific failure went from about eight
hundred occurrences to **zero**, and has stayed at zero through a hundred and ninety-three runs of
real traffic. Their fix used our new "or fail loudly" mark — in the safe order we had warned was
essential, because putting it on before the underlying fix would have broken every dispatch in
the fleet. And the strict mark's own first pipeline is now proven too: twenty-six runs, all
twenty-six passing the right value through, none of them tripping the failure mode we were
watching for.

We got one thing wrong and have corrected it in both places it appeared, including in the other
lane's file: we described their fix as a seventy-three per cent improvement, measured over eleven
runs in eighty minutes. Measured properly the next morning, over a hundred and ninety-three runs,
it is fifty-three per cent. Still large, still real — but we published a number our sample could
not support, and it sat in someone else's bug record for a day.

## Where we are now

Stage one is complete and both of its promises are discharged: the observation window can
actually be read after the fact, and the strict mark is proven on a live pipeline. The council
approved the work; the migration is applied; everything is verified against the running binary
rather than against what we hoped had been deployed.

What is left is smaller and named. The search still guesses in about half a dozen identified
places — two of them busy, the rest occasional. For each one the question is the same and simple:
is the value it picks the value the step actually needs? If yes, write the wiring down explicitly
and then mark it strict. If no, that pipeline has been living on the search and needs the wiring
even more urgently.

There is one honest loose end in the middle of that. Every reference in the busiest offender
already *looks* explicitly wired, so something is reaching the search that shouldn't be, and we
have not yet found what. We have ranked the two candidate explanations and the likelier one is a
single search of the code away. We also confirmed that the dangerous consumer — the step that
decides which piece of work gets marked finished — is already protected, so the remaining noise
is untidy rather than urgent.

## Where we're going

Finish the list. That means answering the loose end above, then walking the half-dozen places one
at a time and writing down explicitly where each value comes from. When that list reaches zero,
the precondition the owner set for stage two is met, and stage two — the search stops guessing
altogether, and returns nothing when candidates disagree — goes to the review council as its own
piece of work.

Two things we are watching rather than doing. The strict mark's failure branch has still never
fired in production, which is the correct and expected state, but it means that path is proven by
reasoning and tests rather than by having happened. And while proving the image pipeline's mark,
we noticed that fourteen of its twenty-six recent deployments failed — for a reason that has
nothing to do with our work, well after our part of the job, on an error about fetching a
repository branch. That is a fifty-four per cent failure rate on asset deployments and it does not
appear to belong to anyone at present. It is written down where the next person will see it.
