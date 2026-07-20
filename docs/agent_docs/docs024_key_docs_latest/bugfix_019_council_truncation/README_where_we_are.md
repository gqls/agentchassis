# Where we are — council truncation bug (019)

Plain-prose running log, append-only, newest at the bottom.

---

**2026-07-19/20 — found the real cause, built the fix, one session.**

The bug as filed: when one council reviewer writes past its output limit, the
whole review round is thrown away — every other reviewer's finished work
included, with nothing to show for the credits spent. Four different threads had
hit it and documented it carefully.

The first real finding was that the case file, careful as it is, pointed at the
wrong place. Everyone believed the crash happened in the code that *adds up the
verdicts*. Counting ten days of failures showed nine out of eleven never got
that far: the round died the moment the overrunning reviewer's own call came
back, because the low-level code that talks to the AI provider throws away
everything the model wrote and reports a hard failure. The reviewers run one
after another, and the one that overruns most often happens to run *first* — so
most dead rounds contain no opinions at all. There was also a queued request for
an automated diagnosis of this bug, written before we looked, and it pointed at
the wrong place too. We cancelled it rather than spend credits confirming a
mistake.

We ran the automated diagnosis loop on the corrected theory first, as the house
rules ask. It came back "cannot verify" — for an interesting reason: the
councils record their activity under the generic label rather than their own
names, so the loop's evidence searches found nothing. Its scepticism was correct
behaviour on the evidence it could see, and it flushed out a mistake of mine (I
had claimed the failing runs were council runs without having actually checked —
they were, but I only knew that after the loop made me look). It also surfaced
something genuinely new: a different council, absent from the case file, hit the
same wall at a limit *four times higher*. That settles a standing argument —
raising the limit does not fix this class of failure, it just moves it.

The fix, in one sentence: keep what the model managed to write, let the round
carry on past the cut-off reviewer, and treat that reviewer as "could not be
read" — loudly recorded, and never allowed to be the difference between "approve"
and "send it back". If salvaging the cut-off review recovers a usable verdict, we
use it, marked as damaged goods. What we deliberately did not do: raise the
limit (the evidence above), or make this behaviour automatic everywhere (owner
chose narrow — councils only; other parts of the platform genuinely cannot cope
with a half-written answer, and for them a loud failure is right).

Everything is committed and the configuration is applied live. The code does
nothing until the next image is built and rolled out — the currently deployed
build predates it. So the bug file stays open: the defect is still reproducible
in production today.

We put the fix through the council gate itself, which has a certain irony — the
gate reviewing the fix for its own worst habit, through the code that still has
the bug. The verdict is pending as this is written; it is advisory, and the
decision to commit first is recorded with reasons in the commit message. If the
round voids, that is a fifth documented reproduction, and it costs nothing we
have not already accounted for.

Nothing needs a decision from you right now. The one open question you already
own — raise the reviewer output limit or not — got harder evidence this session:
a 32,000-token limit truncated the same week, so we recommend not raising as a
fix, only as breathing room sized against the longest writer, if at all.

---

**2026-07-20, later — the same failure, a second cause (bug 036).**

The council lost another round today, and it looked exactly like the bug we just
fixed: run finishes, says COMPLETED, error field blank, no verdict anywhere, every
reviewer already run and paid for. It was not the same bug. Last time the reviewer
got cut off mid-sentence and we salvaged what it had written. This time the
reviewer wrote a complete, perfectly formed answer — and we threw it away, along
with everyone else's, over a single word.

The question we ask each reviewer is "which edit are you objecting to?", and we
expected a number. Three reviewers instead answered in words: "plan-level (deploy
verification)", "risks note on the 54 mis-stamped rows", "risks/summary (item 5)".
Our code could only read a number, so it gave up, and giving up meant discarding
the entire round.

The thing worth telling you is that those reviewers were not being difficult. They
were saying "my objection is about the plan as a whole, not about any one edit" —
and our own format already has a way to say exactly that. We just never accepted it
in words. So this was not a reviewer breaking the rules; it was us refusing a
sensible answer to a question we had asked badly. That changed what we built: not a
grudging bit of leniency, but actually understanding the answer. It also killed the
obvious quick fix — the one the bug file suggested and I would have reached for —
which would have handled "3" in quotes but none of the three real answers. Reading
what the reviewers actually said, rather than what I assumed they said, was the
whole difference.

Two changes went in. We now accept "which edit" in any reasonable form, and we keep
the reviewer's own words in the record rather than quietly replacing them with a
number, so anyone auditing a council decision sees what was really said. And more
importantly, we finished the job we started with the last bug: one reviewer's
unusable answer now costs that one reviewer, never the whole round. That principle
now covers every way a reviewer's answer can be unusable, not just the two we had
already been bitten by.

The safety rule is unchanged and still the point: if we could not read a reviewer,
the plan cannot be approved — it goes back for revision. The change can only ever
make us more cautious, never less.

Same caveat as last time: this is code, so it does nothing until the next image is
built and rolled out. The bug stays open until then, because it is still
reproducible in production today.

One thing you may want to know about, which is not ours: while this was running,
three jobs were sitting hung — two for over an hour, one for over three — and they
were clogging the queue everything else waits in. Both of this session's own jobs
sat behind them. That is an already-filed bug (029), not a new one, but today it
was slowing real work down rather than just existing on a list.

Nothing needs a decision from you.

---

**2026-07-20 evening — done, and proven done.**

You rolled the new build just before seven. We then did the one thing the whole
file had been building towards: made the bug happen on purpose, on the live
system, and watched the fix catch it.

The set-up: a throwaway copy of the review council with one reviewer's output
limit set absurdly low, so it was guaranteed to get cut off. Under the old code
that guaranteed a dead round — everything discarded, no verdict. What actually
happened: the cut-off reviewer was recorded as "could not be read", the other
nine reviewers ran and were heard, and the round returned a proper verdict with
the unreadable seat named in the record. The log line for the cut call even says,
in so many words, that the round carried on without it. The throwaway council was
deleted afterwards.

So the bug that four different threads documented across two days is now closed,
verified by the exact test those threads wrote into the file. The case has moved
to the closed pile.

One hand-off note: the second, unrelated way a round could die — the "answered in
words instead of a number" bug found earlier today — shipped its fix in this same
build. Our test didn't happen to exercise it, so confirming that one stays with
the thread that fixed it; the evidence that it's in the build is written down for
them.

Nothing needs a decision from you. The only 019-adjacent question still on your
desk is the old one — whether to raise the reviewers' output limit for breathing
room — and that is now genuinely optional rather than urgent, because running out
of room no longer costs a round.
