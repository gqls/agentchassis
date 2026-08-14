# Summary — the index-answerability lane, as of 2026-08-14

Third in the series (08-10, 08-11, this). Written at a real milestone: both bugs closed
and live, the governance question ruled, the counter built and enforcing.

## What we're trying to do

Make the code index's answers honest about what they can and cannot see, so the
automated reviewers that depend on them stop inventing explanations for empty results —
and settle the governance question that work surfaced about how shared machinery is
allowed to grow.

## Where we've come from

It started with a verifier confidently declaring three of our own scripts non-existent.
The real cause: the code index holds Go symbols only, and nothing in its answers said
so — a reviewer handed "0 rows" filled the silence with a story. Phase 1 made every
answer state what the index cannot represent; phase 2 removed two of those blind spots
outright by indexing package-level declarations. Then, verifying our own paperwork, we
found the same disease one level up: the index is only as fresh as the last push, the
staleness warning lived in a header the reviewing model demonstrably read and ignored,
and saved verdicts couldn't be dated against the code at all. That became its own bug.
Along the way, the architecture reviewer kept flagging exactly the careful, opt-in way
we had been told to ship changes — so following the rule and breaking it drew the same
warning. That contradiction became RFC 022.

## What we've done

The staleness bug is fixed by placement, not by another warning: every empty answer now
says, in the same breath, which commit the index describes and that anything newer
cannot appear in it; every saved verdict now ends by naming that commit. The reviewers
approved it unanimously; it is live in the current build and we watched a real run
persist the dated verdict. The diagnosis loop deserves credit: it refuted our own first
theory (that no warning existed) before we could file it as fact, and the honest
correction is recorded in the wrong-calls log.

RFC 022 is closed in three steps. First the interim ruling: an opt-in field, off by
default, named by no live consumer, is not architecture business. Then the counter that
ruling demanded: a one-command census of how many optional fields each shared action has
accumulated, checked against a budget. Then the budget itself, ruled at ten — with a
correction that is now written into the reviewers' own instructions: sharing is the
estate's founding design, agents are meant to be reusable across workflows, and what
gets reviewed when the budget trips is the accumulated pile of switches, never the
reuse. The reviewer prompts were updated surgically four times through this arc, each
time without disturbing the prompt-caching layout that carries most of the fleet's
review spend.

## Where we are now

Nothing in this lane is broken, pending, or waiting on a build. Both bug files are
closed with their evidence inside. The counter runs with the ruled budget of ten as its
default and today flags exactly three actions — the repo analyser (12 optional fields),
the note-writer (11, used by eight agents), and the fix-commit preparer (11) — each of
which owes one ordinary review of its accumulated switches, after which its level is the
accepted baseline. The owner has ruled the existing Python mirror inside the scheduled
RFC 006 check stays as it is; no Go rewrite. Two small watch items are written down with
their checks: the first live rendering of the new empty-answer note (none has occurred
yet — a zero means no empty answer has happened, not that the note is missing), and
whether wrong kind-shaped explanations actually stop appearing now the commit is stated
beside every empty answer.

## Where we're going

Little remains, and none of it is urgent. Whether the counter runs on a daily clock or
stays run-on-demand is the one open choice; the three flagged actions need their one-off
reviews scheduled by whoever picks them up; and bug 223's own file could move to the
closed pile under the restored filing rule. After that this lane is complete, and what
it leaves behind is machinery rather than advice: an index that states its own limits,
verdicts that carry their own dates, and a counter that notices the tenth switch nobody
would otherwise have watched for.
