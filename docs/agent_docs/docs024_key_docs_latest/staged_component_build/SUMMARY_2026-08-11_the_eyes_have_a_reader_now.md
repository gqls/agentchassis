# SUMMARY — 2026-08-11: the eyes have a reader now, and getting there caught a false claim

*The read-out on `bugs_open/243` candidate 3, written to be said aloud. Follows on from
`SUMMARY_2026-08-10b_the_acceptance_tests_lost_their_eyes.md`, which covers how we
discovered the seeing half of tool testing wasn't running at all. This one covers what
happened once we gave it somewhere to report to.*

## What we're trying to do

Make sure that when the automatic vision check on a tool actually spots something wrong —
illegible text, a broken layout, anything a person would notice and a mechanical click-test
wouldn't — a human actually gets told, reliably, every time, with nothing lost in between.

## Where we've come from

Two summaries ago we found that the seeing half of tool testing hadn't been running at all
(a missing storage setting), and fixed that. Then, the very first time it DID run for real,
it immediately proved its worth: it spotted genuinely illegible text on one of our tools —
and the run still reported a clean pass, because nothing was reading what the vision check
wrote down. It filed its opinion in a notebook nobody ever opened. So the eyes were open,
but there was still nobody home.

## What we've done

Built the missing piece: when the vision check reports a real problem, it now raises a
proper item in the same review queue a human already checks for other things — deduplicated,
so the same tool re-failing every night doesn't spawn a pile of duplicates, just refreshes
one.

That went through our review process, which exists specifically to catch problems like this
before they ship — and it worked exactly as intended, three times over. First pass: a
reviewer pointed out that if the "tell a human" step itself failed to save, we'd be back to
square one with the failure just moved one step later, so we added a proper record of that
too. Second pass: a different reviewer noticed the record we'd added was a bespoke,
one-off solution when the platform already has a single proper mechanism for "something
failed durably" — so we swapped to the existing one instead of keeping a second
lookalike around. That review round also caught something we'd got factually wrong: we'd
written that nothing ever automatically re-checks this kind of stuck item, to justify not
worrying about it. That's not true — there IS a system that automatically re-checks stuck
items every day. We'd made an unforced, false claim, and a reviewer caught it by simply
checking. Third pass, with that corrected and the fix applied: approved.

## Where we are now

The feature is live and signed off. We've watched it correctly stay quiet on three separate
clean runs, including a re-check of the very tool that started this whole thread — the
vision check now looks at it and confirms it's fine, three independent times, which is
reassuring in its own right. Nobody has yet seen it actually raise a real alert in
production, simply because nothing has gone wrong again yet — that half is tested but
unproven live, and we're comfortable leaving it that way rather than manufacturing a fake
defect just to watch it fire.

The correction the reviewer forced turned out to matter more than the wording fix it looked
like at first: yes, there's a daily automatic check for stuck items — but checking closer,
it only knows how to re-check six specific kinds of problem, and this new "a vision model
didn't like what it saw" kind isn't one of them. So the daily check runs, faithfully, and
still walks straight past this new kind of item every single day. Once someone files one of
these, today, nothing will ever automatically notice if the underlying page gets fixed — it
will sit waiting for a person to remember to go and close it.

## Where we're going

That's a real, separate decision, not a defect in what we shipped — the review board was
satisfied that "we haven't built an automatic re-check for this yet" is an honest, acceptable
answer for now, not a thing that had to be fixed before approving. But it shouldn't sit
undecided forever, and it's a genuinely different kind of problem to solve (the six things
the daily check already knows how to re-verify are all yes/no questions a computer can
answer cheaply; "does this still look ugly to a vision model" is neither cheap nor quite
that certain). So rather than decide it on the spot, we've written up everything needed to
start planning it properly — the mechanism, the options, the open questions, the traps — as
its own document, ready for a fresh look whenever you want to pick it up.
