# README — where we are (bugs_open/297)

Plain prose, append-only, newest at the bottom.

## 2026-08-17 — picked up, checked, and the fix turned out simpler than its sibling

This bug is the second of three "silent row cap" bugs the 275 work uncovered. When the system
rebuilds an interactive tool (a calculator, a game) on one of our sites, it first asks the
database "what other pages does this site have?" so the AI understands how the tool fits the
site. That question was capped at 10 answers — and which 10 was decided by menu position, not
relevance. On the biggest site the AI saw 10 pages out of 107.

I checked nobody else was fixing it (two ways — the ownership script, and reading the other live
sessions' transcripts), confirmed it is still real against the live database, and then measured
before choosing a remedy. The measurement was good news: unlike the sibling bug, where each row
carried a long description that had to be trimmed before the cap could go, these rows are one
short line each — page name, type, title. Showing ALL pages on even the biggest site costs about
two thousand words of context in a prompt that already contains the entire original page. So the
cap can simply go, with nothing trimmed and nothing hidden.

While measuring I found a second, smaller defect in the same query: one page on one site has two
research records attached, and the query lists that page twice — today, on the live system. The
fix closes that door too: each page now contributes exactly one line, using its newest research
record.

The change is one database migration (453), written to the same hardened pattern the sibling's
council review settled: it takes a backup snapshot first, refuses to run if another session has
already changed the row, and verifies its own result before committing. Next: council review,
commit, apply, verify live.

## 2026-08-17 (later) — the fix is live, and the council is looking at it

The migration is applied. The system now shows the analysing model every page on the site rather
than the first ten by menu position — on our biggest site that is 107 pages instead of 10 — and
each page appears exactly once, which closes the duplicate I found earlier today. A backup of the
old configuration was taken automatically first, and a tested one-command revert sits beside the
migration if anything looks wrong.

One small thing worth recording, because it is the kind of judgement I would want to be asked
about: while writing up the risks I noticed the "newest research record" rule could misbehave if a
record ever arrived without a timestamp. None of the twenty-one existing records has that problem,
so I could have left it as a note for the reviewers. I closed it instead — it cost one word in the
query and it means the problem cannot happen rather than merely being unlikely today.

The council round is running (they take about half an hour to reach the front of the queue). I have
committed everything with a marker that says "submitted, verdict not yet read", which is the house
rule here — nobody holds work waiting for a review, because the review is designed to come after.
When the verdict lands I will read it and act on it; if the reviewers find something, that is
cheaper than finding it later, and on the sibling bug two of the rounds found real defects.

## 2026-08-17 (evening) — the council approved it, and their questions made the answer better

Approved on the first round: fourteen reviewers, three abstained, four advisory objections and none
of them serious enough to block. That is the outcome, but the useful part is what the objections
made me go and check.

The strongest one was worth its fee. One reviewer said, in effect: you have removed a limit on how
much text goes into the AI, so that text now grows forever as a site gains pages, and this system
has a history of AI replies being cut off silently when prompts get big. I had left that as "owed"
— something to confirm later — and being called out on it was fair. So I measured. In every one of
the 129 times this step has ever run, the reply has never once been cut off; and if it ever were,
this agent would fail loudly rather than quietly save half an answer. Better still, there is already
a standing check that runs every six hours across the whole fleet watching exactly this, and it
already lists this step by name as one to keep an eye on — peak usage 96.7% of its allowance, zero
truncations. So the guard the reviewer wanted turned out to exist; I just had not looked for it.
What is true is that the headroom is thin — about 265 words' worth — and the existing check is what
will tell us if that changes.

Two of the objections were straightforward misses of mine, both worth recording. I had written that
this step's output has only one consumer, but I had checked only inside this one agent and stated it
as if I had checked everywhere. When I did check everywhere, the claim held — but it held by luck,
and I found that four different agents use the same field name for two unrelated things, which is a
trap for the next person. And I had not searched our own landmine file for this agent's name before
editing it; six entries mention it. None of them contradicts what I did, but "nothing is wrong" is
only worth something once you have actually looked.

Everything is committed and the fix is live. The bug file now carries the verdict and the
measurements behind it.

## 2026-08-17 (late) — asked whether to force a live run, and decided not to

The one loose end was the final end-to-end proof: seeing the longer page list inside a real
rebuild's prompt. That needs an actual tool rebuild, so I went looking for somewhere safe to run
one, and the looking turned out to be worth more than the run would have been.

First, the free finding. Instead of firing something, I read the prompts this system has already
sent in the past. Every single one of them — eight rebuilds, back to 8 August — listed exactly ten
other pages, on a site that has thirty-two. So we now have proof from the real thing, not from
reasoning, that the limit was genuinely cutting what the AI saw. The same prompts also answer the
reviewer's worry about size in real numbers: that list was 979 characters inside a 29,000-character
prompt, and showing the whole site would make it about 2,000. Three times the coverage for about
three and a half percent more prompt.

Then the awkward part. There are only six pages in the whole estate where we still hold the
original crawled code that a faithful rebuild copies from, and all six are on one site — which
another session is working on right now, on the very page I would have chosen. Worse, all six are
marked "owned", which means a rebuild would do all the expensive AI work and then be refused at the
last step by a safety guard, leaving a failed job behind and muddling that session's experiment.
The one site where a rebuild would actually complete has lost its original source entirely, so the
AI would be rewriting a live, public mortgage and stamp-duty calculator from a summary, with nothing
to copy. This estate has already once shipped a calculator using a tax threshold that had been
withdrawn, so that is not a risk to take for a log entry.

So I put the choice to you rather than picking, and you said don't force it. That is now written
into the bug file so nobody later reads the missing proof as an oversight and decides to spend a
live page closing it. The check takes one query whenever a rebuild happens naturally, and the query
is saved in the file.

One useful thing fell out of the search: the guard that blocks those rebuilds tells whoever reads
its refusal to "use the tool pipeline for rebuilds" — but the tool pipeline uses the very step being
refused, so for the twelve pages in that state the advice points at a door that is locked. That has
gone into the relevant bug for the session that owns it.
