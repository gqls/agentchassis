# SUMMARY — 2026-08-20: four of the five steps are live, and the fifth turned into a list

## What we're trying to do

Stop the system guessing. When one part of the pipeline needs a piece of information — the name of
the page being built, which code change deployed a job, which pages a new tool relates to — it asks
a shared lookup for it. That lookup has a last resort: if it cannot find the value where it was
told to look, it searches the entire job's data for anything with a matching name and uses whatever
it finds. When it finds two different answers, it picks one. The aim of this workstream is to make
it refuse instead, on the principle the owner has already ruled on: **no value at all is better
than a wrong one.**

## Where we've come from

We could not simply switch the guessing off, because nobody knew what depended on it. So we built
an instrument first: every time the search finds conflicting answers, it records the field, every
place it looked, and which one it picked. Then we worked down the list the instrument produced,
removing causes at source rather than papering over them — a wasteful search that ran for pages
nobody had asked about, an undeclared tie-break that decided winners by accident, a retry that
echoed itself, and finally a genuine name collision where a page's *record* and a page's *name*
were being filed under the same key.

## What we've done

All four of those are now built, reviewed, shipped and proven live. The last of them landed on the
build that rolled overnight, and we verified it the hard way: the service's own startup line
naming its code had already expired, so we asked the running binary directly, on both machines,
with a control check either side to prove the probe could tell the difference. The warning that
change was built to eliminate went from about three occurrences per job to zero.

Along the way the instrument earned its keep by catching something nobody was looking for. When
the system builds a tool for a site and nobody has said which pages relate to it, the correct
answer is "none". Instead the search finds the page list belonging to a *different* tool and uses
that — so on one site, nine unrelated tools were all queued to be linked from the same two
statistics pages. That is now filed, independently confirmed by the diagnosis loop, and none of
those wrong links reached a live page.

## Where we are now

One step remains: the flip itself. We had expected it to be quick. It is not, and finding out why
is the main result of the last day.

The flip is only safe once every place still relying on the guess has been told explicitly where
to look. We checked how many such places there are: **thirteen**. Most of them look dormant — their
last warning was a day or two ago — but that reading is a trap, and we can now prove it. One of
them belongs to a component that ran nearly six hundred times in a day without tripping. They are
not fixed; they are waiting for the right data to come past.

One of the thirteen has a real, named cost. Every finished job records which code change deployed
it, and nobody ever configured where that value comes from — so it is being supplied by the guess,
apparently landing on the right answer by luck. Switch the guess off and that field quietly stops
being recorded, on something another workstream's page-publishing fix depends on. We have written
to that workstream with the evidence and asked them for the correct source, rather than picking one
ourselves, since picking one is exactly what we are trying to abolish.

We also corrected our own plan. It justified deleting a piece of compatibility code on the grounds
that old records expire within a day. They do not, and the code is used on a path where the data
never expires at all. The conclusion held up for two better reasons, which are now written down as
queries anyone can re-run — but the reason we had was wrong, and nobody had checked it.

And one honest limit worth stating out loud: the instrument can only see conflicts where the
candidate answers *disagree*. Where the search finds a single wrong answer, it substitutes it
silently, with nothing recorded. So "the warnings have stopped" can never by itself prove the
search is safe. That is a permanent property, not something more waiting will fix.

## Where we're going

The remaining work is a list rather than a mystery, which is the real change since yesterday.
Thirteen items, reducing to four repeating shapes — so most will be fixed in groups, not one at a
time. The largest group is a single mechanism with a case already diagnosed and confirmed, and it
is the one to start on, because the item with the highest urgency is blocked on another
workstream's answer.

Nothing waits on the owner. The next session picks up the list at
`docs/agent_docs/docs024_key_docs_latest/staged_component_build/HANDOFF_2026-08-20_continue_here.md`.
