# Where we are — the orphan check and retired pages

Plain prose, append-only, newest at the bottom.

---

## 2026-08-22 — what I was asked to look at, and what it turned out to be

You asked me to look at bug 298. The first thing worth saying is that **298 itself is genuinely
fixed**, and I checked that against the live system rather than taking the file's word for it.
Its subject was that the internal linker chose link targets from at most fifteen candidate
pages, picked alphabetically. That cap is gone, the migration that removed it is recorded as
applied, and the linker has since made its first real link plan. So there was nothing to fix
there.

What 298 also did was record a loose end it deliberately did not chase: a large minority of the
linker's completed jobs had finished having found **no page to work on at all**. It said twice
that this was unclaimed and deserved its own ticket. Nobody had picked it up, so I did.

**The answer turned out to be short and slightly embarrassing for us.** Every single one of
those empty runs was pointed at a page we had already retired. Not most of them — all seventeen.

## How that happens

A page in our database has two separate facts attached to it, and they are genuinely different
questions. One is *has this page ever been published* — the build question. The other is *do we
still want this page live* — the lifecycle question. A page we retire keeps its publishing
history, so it still answers "yes" to the first question while answering "no" to the second.

Our own code says, in a comment on the shared helper, that anyone asking about pages has to
decide which of those two questions they mean, and that there is deliberately no combined
helper because different parts of the system legitimately need different combinations.

The check that finds unreachable pages asked only the build question. So it kept finding retired
pages, deciding they were unreachable — which they are, because we retired them — and filing
work saying *someone should link to this page from somewhere*.

## Why nothing bad visibly happened, which is the interesting part

The work went to three different handlers depending on the kind of page. **All three refused
it**, and all three were right to, because each of them checks the lifecycle question properly
before doing anything. So we were not re-publishing retired pages. We were doing something
quieter: generating work that could never be done, forever.

Each attempt completed successfully as far as the system could tell, so it looked fine. And
because our de-duplication deliberately allows a finished finding to be raised again next time
round, the same retired pages came back every rotation. Some of them have been re-detected and
re-dispatched three or four times since April.

**The genuinely expensive part is a side effect.** We have a sensible rule that if a piece of
work fails twice, we stop retrying it and park it for a human. These impossible jobs were
burning through that allowance — and when a batch gets parked, it parks *everything queued at
the time*, not just the impossible ones. Of the twenty linker jobs currently parked and
undispatchable, fifteen name pages that are perfectly live. Real work, retired as collateral
damage by fake work.

## It is not one check

I had all seventy-one checks in that part of the codebase read line by line, and then
re-verified a sample by hand in both directions, because a summary I cannot check is just
another document. **Eighteen of them can pick up a retired page and route it at a handler.**

The reason is not eighteen separate mistakes. It is that only three of them use the shared
helper for the lifecycle question; everyone else has hand-written it, in four different ways —
and **two of those four ways do nothing whatsoever**. One of them filters out pages marked
"deleted", and we have no pages marked "deleted"; that value does not exist in our system. It
looks like a filter, it reads like a filter, and it excludes precisely nothing. That is on the
highest-priority check of the eighteen.

## What I have done, and one thing I deliberately have not

I fixed the check that caused the measured damage — two lines, so that it asks both questions
the way its own three handlers already do.

For the class, I did **not** simply add the missing filter to the other seventeen. Two reasons.
The first is that each one needs a real judgement rather than a batch edit, and a sweeping
unreviewed change is close to the thing that caused this. The second matters more: **a retired
page can still be publicly visible.** Another team hit this in August and left a warning saying,
in effect, *do not let anyone "fix" the audits by hiding archived pages, because an archived
page that is still serving to the public is exactly the case you must not stop looking at.* They
are right, and one of our checks depends on that.

So instead, every check that queries pages now has to **declare** which stance it takes and
why, and that declaration is checked by the build. A new check cannot be written without making
the choice. The seventeen outstanding ones are declared as **known gaps** — which is not the
same as being waved through: they are a distinct category, each names this bug, and the test
prints the running count, so a backlog of twenty-three reads as a backlog of twenty-three rather
than as silence.

## A confession about the safeguard itself

I wrote a check to make sure a check that *claims* to filter properly actually does. **I got it
wrong twice, and both wrong versions passed cleanly.**

The first version searched the file for the right function name — and was satisfied by the
explanatory *comment* I had written next to the fix. Delete the actual code, keep the comment,
and it still said everything was fine.

The second version stripped the comments out first, but then searched for the filter without
checking *which table* it applied to. Almost every one of these checks also filters a
navigation table on the same words. So again: delete the real filter, and the safeguard was
satisfied by an unrelated line elsewhere in the file. **That is precisely the same mistake that
made this bug invisible in the first place — reproduced inside the tool I was writing to
prevent it.**

The only reason I know any of this is that I deleted the fix on purpose and checked that the
safeguard complained. It did not, twice. The third version does, with both decoys deliberately
left in place. I have written all of it down, because "the guard looked right and asserted
nothing" is a far more dangerous outcome than having no guard at all — it would have given
every future author a green light.

## Where it stands

The fix and the safeguard are committed. The safeguard is a test, so it is protecting us as soon
as it merges; the two-line fix is Go, so it takes effect at the next chassis build. I have put
the change through the review council and the verdict has not come back yet. The independent
diagnosis run died on an infrastructure hiccup — a Kafka leader election, nothing to do with the
finding — so I have re-fired it and will record what it says, **including if it disagrees with
me**.

Nothing needs a decision from you. The one thing worth flagging for later: we have **no detector
anywhere for "this page is retired but still serving to the public"**. I checked, it does not
exist. My change neither creates that gap nor closes it, but it came up twice while I was
working and it looks like a real hole.
