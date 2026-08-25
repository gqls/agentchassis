# Where we are — counting-fact drift (bugs_open/386)

Plain prose, append-only, newest at the bottom.

---

## 2026-08-25 — what this bug actually is, and why your ruling changed the plan

The short version of the bug: some of the numbers on our sites are live counts — "feed items
collected: 11,513". We keep those counts in a register, and a job re-reads them every night. When
the count ticks up, the register moves and the page does not. Our own honesty checker then compares
the page against the register, finds a number the register no longer holds, and reports the page
for making up a figure. Nothing false is published. The page is simply a day old. But the report is
filed at a severity that refuses a rebuild, so the page cannot be rebuilt to fix itself, and in the
review queue it looks exactly like a page that invented a number.

It re-arms every night on its own, which I could show rather than assert: when the bug was filed on
Sunday the register held 11,646; today it holds 11,828. The page still says 11,513.

The good news is that this is small and precisely bounded. Of 295 facts across the whole estate,
only 29 are the kind a nightly job moves, and only 13 of those are matched strictly enough to
convict a page. They sit on six sites — fundamentallyai, leopardess, robot-hands, vonc, and two
others. This is not a fleet-wide rework.

Then your ruling arrived, and it reframed the thing properly. I had been asking "how do we make the
checker tolerate a number that moved". You asked the better question: should the page be printing a
live counter at all? A counter's honest form is a lower bound. "We have collected 11,646 items" is
false a minute later; "at least 11,000" stays true for months and needs nobody to re-render
anything. So the default is now "at least N", or cut the number entirely where the page does not
need it.

One caution I want to flag before I act on it, because it is the kind of fix that can quietly make
things worse. Switching a fact to "at least N" tells the checker to accept any number below N that
appears near certain words. Those words are matched loosely — as fragments, not whole words — so if
the words are broad, the fact ends up vouching for almost any number on that part of the site. We
have a live example: one fact on ai-agent-orchestration.com is set to "at least" with the single
word "orchestration" attached, and it currently blesses any figure up to 7,281 sitting anywhere
near that word. Your ruling already warned about this and quoted the figure as 4,068 — it was 4,068
when someone wrote it down, and it is 7,281 now, which is the same bug biting the document that
describes the bug.

So the order of work is: measure what each proposed "at least" would start accepting, *then* change
it — never the other way round. Where a page does not need the number, I would rather delete it than
register anything, which is your stronger option and the only one with no ongoing cost.

After that there is a second, deeper fix, which your ruling explicitly left standing: teach the
register to remember the values it used to hold, so a page that printed last Tuesday's number is
recognised as having been right on Tuesday rather than accused of invention. I checked whether that
is even buildable and it is — we never actually delete the old register entries, we just mark them
superseded, so the whole history is still there: 315 old versions going back to mid-July. The
number the convicted page prints, 11,513, is sitting in that archive dated Sunday. So the history
does not need to be reconstructed by guesswork; it is already on disk.

Two smaller things I corrected on the way through, both cases of a document being more confident
than the code. The bug file proposed marking counting facts as "always goes up" — but ours don't
always: one of them falls whenever the table behind it is cleared out, which I found in our own
migration history. And two documents say we can tell a stale page from an invented number by
comparing timestamps on something called `nearest_fact_id`. That field does not exist in the code
at all; it is a label our audit model writes in its own output, so the comparison the documents
describe cannot happen where it would matter. I will correct both in place rather than only noting
it here.

Nothing has been changed on any live site yet. The only thing written so far is this lane's own
notes.
