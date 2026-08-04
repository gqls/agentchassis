# Where we are — bug 195, the classifier that read error messages like prose

Plain prose, append-only, newest at the bottom.

---

## 2026-08-04

**What was broken.** When the platform is handed a workflow that is simply *wrong* — a
configuration mistake, the kind that will never work no matter how many times you try it —
it is supposed to do two things: stop trying, and write down what happened. It was doing
neither. The job vanished. No record in any table, no error report, and one line in a log
that had rotated away within eight hours. From the outside it was indistinguishable from the
system just being slow, which is exactly what the last team to hit it assumed — and they
lost three attempts and a quarter of an hour to that assumption.

**Why.** The code decides "is this permanently broken, or just a hiccup?" by **searching the
error message for words**. It looks for four: *is required*, *validation*, *invalid*,
*missing*. The actual error says:

> `WORKFLOW_INVALID: Invalid workflow configuration (caused by: step 'done' … requires a topic)`

Read those against the list. *"is required"* — the message says "**requires** a topic", which
is the same thing in different words. *"invalid"* — the message says "**I**nvalid", with a
capital I, and the search is case-sensitive. So the single most common permanent
configuration error on the platform matched **none** of them, and was filed as a hiccup worth
retrying.

The maddening part is that the error already carried an exact, machine-readable label
(`WORKFLOW_INVALID`). The code threw that away and went looking for words instead.

**The deeper problem, which is the bit worth remembering.** An earlier fix had promised
"a dropped error always leaves a record". That promise was real, and it was tested, and it
shipped. But the record gets written *inside the branch this word-search chooses* — so the
promise was really "always leaves a record, **provided the word-search recognises it**", and
nobody ever checked whether it recognised the commonest case. **A guarantee that depends on a
classifier quietly inherits every gap in that classifier.**

**What I changed.** Two things. It now matches on the exact machine-readable label instead of
hunting for words — which cannot be defeated by rephrasing or by a capital letter. And,
separately, it now writes the record for **every** failure, not only the ones it classified.
That second part is the one I'd defend hardest: it means being *wrong* about the
classification costs us some wasted retries, but can never again cost us the only evidence
that anything happened at all.

**Two things I found while in there.**

The report I was working from said the failed job was being *retried*. It isn't — the phrase
it counted appears twice per single attempt, so one attempt looks like two. That matters more
than it sounds, because the report also proposed checking for exactly that phrase to confirm a
fix, and that check would have reported failure against a fix that was working perfectly.

And the report's open question — "does the waiting parent job hang?" — has an answer, and it's
worse than hanging. **The parent is told the step succeeded.** The failure is sent back in an
envelope stamped "complete", and the parent reads the stamp, not the contents. So it marks
the work done and moves on, carrying an error message where its results should be. That's not
this bug and I haven't fixed it, but I've written it down prominently because it is the kind
of thing that looks fine in every dashboard.

**Honestly: I made a mess in the middle of this.** To prove my new test could actually detect
the bug, I deliberately broke the code, ran the test, and restored it. The restore relied on a
backup copy that — unknown to me — was never made, because the scratch folder had been
cleared. The script didn't stop when the copy failed. So for a few minutes a shared source
file sat on the branch in a deliberately broken state. One final check caught it. I've written
up what went wrong, and corrected the recipe I'd published three hours earlier, which was
right about what to do and silent about what had to be true for it to be safe.

**Where it stands.** Fixed, tested, committed, and with the review panel now. It is **not
live**: this is program code, which does nothing until a new build is rolled out. Until then
the fault is still there on the running system. The ticket stays open and says so.
