# SUMMARY — 2026-08-15. Prompt caching is proven at the artefact, and it has a second adopter

Written to be read aloud. Supersedes nothing; the previous read-out is
`SUMMARY_2026-08-14_config_beats_default_is_live_and_the_review_that_improved_it.md`.

## What we're trying to do

Cut what we spend on the language model without changing what any of it does. The lever is
caching: most of what we send is a long, unchanging block of instructions and context, and the
provider will store that block and charge about a tenth as much each time we send it again.
The catch is that the stored copy expires. Everything turns on whether it survives long enough
for our next request to reuse it.

## Where we've come from

We switched caching on for the review council five days ago and it worked well — that one
change now accounts for the large majority of our caching benefit. But it exposed a problem we
then sat on. The stored copies were surviving only five minutes, because a reviewer had
removed the setting that asks for an hour: at the time, that longer setting appeared to need a
special permission we didn't have, and getting it wrong would have returned an error rather
than a saving. Since the review council checks every change we make, an error there would have
taken out our whole review process. That was the right call on the facts available.

Yesterday and this morning we checked the facts again and they had expired: there is no
special permission to obtain, and the longer setting simply works. So the hour-long setting
went into the release that shipped this morning. What we did **not** have was proof it was
actually taking effect in production — and the release had been justified almost entirely on
that claim.

## What we've done

Two things, in a particular order that turned out to matter.

First, we proved the hour-long setting is real. The proof had to be a request that reused a
stored copy *more than five minutes* after it was stored, because under the old setting that
copy would have been gone. We got one: two requests **8 minutes and 46 seconds apart**, and the
second read the stored copy rather than storing it again. To be sure that check could have
failed, we ran the reverse against three days of earlier history — 29 times a request came back
after more than five minutes, and 29 times out of 29 it had to store a fresh copy. So the
question was genuinely open, and the answer is yes.

Second, we switched caching on for a second agent — the one that works out what content a site
is missing. That took a detour. All our evidence about the hour-long setting came from the
review council, and every seat on that council runs one particular model; the content agent ran
an older one. We had confirmed the older model *accepts* the setting without complaining, but
never that it *honours* it. Switching caching on there would have staked the entire saving on
something nobody had checked, and the failure would have been silent — we'd have paid the
storage fee repeatedly and read nothing back. The owner's call was to move the agent onto the
proven model first. That move had a side effect worth knowing: the newer model thinks before
answering by default, where the old one didn't, and the budget for its answer covers the
thinking too — so we had to raise that budget or risk answers being cut off mid-structure.

We also got something wrong and fixed it. The raised budget went into a settings field that
nothing reads, and the safety check written to catch exactly that asked whether the number just
written was big enough — a question that can only answer yes. For about nine minutes the agent
ran in the state we'd written a page explaining how to avoid. No work came through in that
window, which is luck rather than good practice. It is corrected, and logged, because the shape
of it generalises: a check that reads back your own change cannot fail, and it looks exactly
like diligence.

## Where we are now

Caching is live and independently proven on two agents. The content agent's first request
stored just under 5,000 units — about 78% of everything it sends — and its next request read
them back. No errors, no cut-off answers, nothing needing reversal. The setting stays.

One useful discovery: we had been watching the wrong agent for proof. The review council is
busy enough that gaps longer than five minutes essentially never occur, so it can almost never
demonstrate the thing we needed demonstrated. The content agent is asked to work every ten
minutes or so, which is precisely the interval that discriminates — it answered the question
within nine minutes of being switched on. It is now the place to check this.

One caveat we should not lose. The newer model counts the same text as roughly a third more
units, so on its own the move costs more, not less; caching more than covers that. It is also
on introductory pricing until the end of this month, so any figure quoted today flatters it.
The number to trust is a September one.

## Where we're going

Two things remain, and the larger one has nobody on it.

The smaller: a daily automated check we designed but have not built, which would catch a class
of configuration drift we currently only find by hand.

The larger: our single biggest consumer — the agent that writes page content — accounts for
roughly 38% of everything we send and cannot be cached as it stands. Its shared material sits at
the *end* of what it sends, and caching only works from the front. Fixing that means
restructuring how its instructions are assembled, not flipping a setting. It has not been
costed and nobody owns it. On today's evidence that restructure is the largest remaining saving
available to us.
