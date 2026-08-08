# SUMMARY — 2026-08-08 · the repair agent that never repaired anything, and what we found when it finally ran

Written to be read aloud. Figures re-measured against the live system this afternoon, not carried
forward from the 08-06 summary.

*(Second summary in this lane. The first, `SUMMARY_2026-08-06`, was written the moment the fix went
live and said plainly that it was unproven. That is the difference this one records — and the
difference turned out to matter more than expected.)*

## What we're trying to do

When one of our automated checks finds a small defect on a live page — markdown symbols showing up
as literal asterisks, a fabricated phone number, a page with no content — it files a repair job and
names the agent that should carry it out. We are trying to make those repairs actually happen, and
to make the system tell the truth about whether they did.

## Where we've come from

This began as a loose end on a different bug and turned into two distinct faults in the same
pipeline.

The first was that the repair jobs died on arrival. Twelve were attempted; eleven failed outright
and the twelfth reported success having changed nothing. The cause was not what the error message
said: the writer agent, called directly, expects to be told which parts of the page to work on, and
the checks never sent that. It looked at an empty list, concluded there was nothing to write, and
gave up. Fixed by pointing those checks at the page *build* handler, which fetches that list from
the site's own plan instead of trusting the caller — completing a migration two other checks had
already made.

The second was worse and quieter: when a repair reported success, nothing checked whether it had
written anything. The system took the agent's word for it.

## What we've done

**Both faults are now fixed, reviewed, deployed and demonstrated working in production.**

The first was proven a week ago by watching a fresh job get filed with the corrected routing.

The second needed a guard at the moment a job is marked done. We already had the machinery for
that — it just had never been switched on for this kind of job — so this was filling a declared gap
rather than building anything.

**The review board rejected my first attempt, and it was right to.** I had made the guard *refuse
to answer* when it couldn't tell whether a page had been repaired or emptied. That reads as the
cautious choice. It is the opposite: the system treats "couldn't check" as permission to proceed, so
my refusal quietly waved the job through — on precisely the case I had written it to catch. I had
even noticed that behaviour and written a note about it, then shipped it anyway. Noting a fault is
not controlling it. The guard now blocks instead, and a test makes the old version fail the build.

**A reviewer also caught me leaning on bad evidence.** The example I had been citing all week as
proof of the second fault sits on a page that is a known, deliberately-quarantined special case, so
it cannot prove anything. I have withdrawn that claim rather than defend it — and flagged that the
original bug report leans on the same page, which is for its next reader to weigh.

**Then we ran it for real, on a live site, with your go-ahead.** The first attempt failed for an
unrelated infrastructure reason — a known handshake fault that fails about half the time — and,
importantly, it failed *before* reaching the guard. It would have been easy to report "the job
wasn't marked done, so it works". It wasn't marked done because the run crashed. I recorded it as
inconclusive, which is the only reason we tried again.

**The second attempt is the result.** The repair ran properly, rebuilt the page — and the guard
refused to mark it done, naming the exact field and the exact text still wrong. Before this change,
that job would have been recorded as a success.

## Where we are now

**Both faults fixed and proven. The page we tested on is still broken, and that is the interesting
part.**

The repair genuinely ran: every part of that page was rewritten. The defect survived it. The writer
put markdown syntax straight back into the very field it had been sent to clean — eighteen
instances after a complete regeneration.

So the honest reading of the week: **routing was never the whole problem.** The detection is
correct, the dispatch now works, and the thing doing the writing is the remaining fault. That
belongs to a different bug, which is now unblocked and whose failure has moved somewhere visible.
There is a pending database change on that bug's own path that looks like the real lever, and I have
pointed it there.

The job is now marked failed after three attempts and routed to a human, which is the right
destination. Nothing is silently claiming success.

Measured today: the guard is still in the running system, the scheduler bypass that would have
walked around it is still closed, and the earlier bug's no-regression count is still zero.

## Where we're going

One thing genuinely waits on you. The review board pointed out — twice, from two independent
reviewers — that my fix protects *this* guard and leaves the general rule alone: every other guard
of its kind can still fail to check and have its job waved through. I have written that up with four
options costed, and I deliberately did not recommend the obvious one, because reversing the rule
would strand work whenever a guard has an off day, and we have been badly bitten by that shape
before. The decision needs one number nobody has measured — how often these guards actually fail in
practice. I can measure it.

Beyond that: the other bug now owns the writer problem, a separate thread owns the page-rebuild
fault we found in passing, and the site rebuild you asked about is scoped and waiting.

The pattern worth carrying out of this week: **three times, the thing that caught the error was a
check I had written down and then not applied to myself** — a deployment test that couldn't fail, a
verification that couldn't refuse, and evidence I hadn't checked the provenance of. All three were
caught, none of them by me first.
