# SUMMARY — ai-agent-orchestration.com, 2026-08-25: the fix that shipped a placeholder

Second summary for this lane, one day after the first. **Written not because a day passed — that is
explicitly not the trigger — but because the read-out genuinely changed.** Yesterday's summary said
the work was finished. It was, and still is, for the thing it measured. What it could not say is
that a change this lane made on Friday had been publishing a placeholder to the live site for three
days, and that somebody else found it.

## What we're trying to do

Unchanged: make ai-agent-orchestration.com a site that demonstrates the platform rather than
argues against it.

## Where we've come from

Yesterday: 44 unreadable elements reduced to zero across four pages, carousels built, ten broken
pictures replaced. That is all still true and none of it has regressed — re-checked today.

## What we've done since

**Found and fixed a defect we caused.** On Friday, migration `557` rewrote the site's instruction
sheet so it carried no hard-coded counts — the counts are a live database figure and any number
typed into a document is wrong within days (it has moved 175 → 196 → 199 → 200 in a month). In
place of a number, the sheet told the writer: *phrase it as "NNN+ AI agents", and take the live
value from the facts list.*

Both halves were wrong in the same way. **`NNN` is a stand-in with nothing behind it to substitute
it**, and **the writer is never shown the facts list** — its prompt contains the instruction sheet
and not the data. So it did the only thing it could: it printed what it was shown. Since Friday,
14 of 137 attempts copied `NNN` verbatim and **none** wrote the real figure. One of them reached
the public: *"…against the NNN+ agent types already running in production…"*, live on the model
directory page.

**We did not catch it.** Another team did, while at this site for an unrelated reason. That matters
more than the bug: this lane re-read that migration several times and verified it at the live page,
but verified *the thing it had changed* — that the sheet no longer carried a stale number — and
never asked what the writer would do with the replacement. The check that would have caught it was
one query against the log of what the writer is actually sent.

Two repairs are in: theirs replaced the stand-ins with plain lower bounds ("more than 150 active
agent definitions") and banned letter stand-ins outright; ours repaired the sibling field they
flagged back to us. **Censusing that field found more than was reported** — three frozen dates
rather than two, plus two lines that would have published figures the sheet itself forbids.

## Where we are now

The readability work remains finished and the carousels and pictures are intact. The placeholder is
out of the source and the affected page regenerates on a six-hourly cycle. Nothing else on the site
carries it.

The lesson we are taking is not "be more careful with placeholders". It is that **prose written
into a prompt is input, not documentation.** A style guide can say "NNN" or "take the value from X"
to a human, who resolves it against context they already have. A generative reader has only the
bytes in front of it, and its failure mode is to render a dangling pointer rather than notice that
it dangles.

## Where we're going

Nothing is blocked. The remaining items are small and none is urgent: a batch of old readability
items that only an automatic audit can clear (it has not run here since 10 August); and the
automation that would keep the instruction sheet current, which still cannot be switched on because
as built it would delete the site's "never claim this" rules. Our repair removed one precondition
for that; it is not the last one.
