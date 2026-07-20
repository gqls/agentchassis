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
