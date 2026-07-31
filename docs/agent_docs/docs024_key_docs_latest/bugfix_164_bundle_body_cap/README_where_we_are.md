# Where we are — the diagnosis bundle that quietly threw away most of its evidence

*(append-only, newest at the bottom, plain prose)*

---

**2026-07-31, evening.**

When one of our diagnosis runs looks at a bug, it first gathers up the relevant
source code into a single document — we call it the bundle — and hands that to the
model that forms the verdict. There is a size limit on how much code goes in:
60,000 characters.

The limit was being enforced in the worst possible way. The code walked down the
list of things to include, and the moment it hit one item that wouldn't fit, it
**stopped entirely** and threw away everything still on the list — no matter how
small those remaining items were. And it said nothing about having done so. The
bundle it produced looked exactly like a complete answer.

Two details make this worse than it first sounds.

The list is in **alphabetical order**. So the things thrown away weren't the least
important ones — they were whatever happened to sort last. One oversized file
beginning with "i" could silently wipe out everything beginning with "p".

And the model doesn't just *read* the bundle, it **decides what to look at next**
based on it. So a run could look at a quarter of the evidence, form a view, ask to
look somewhere else next — and our own progress check would record that as the
investigation successfully narrowing down. It wasn't a display problem. It was
steering.

**Had it actually happened, or was this only theoretical?** The person who filed it
was honest that they hadn't checked, and told the next person to check first. So I
did. Of the 254 bundles we still have on record, **18 hit the limit**. The worst
had 18 items to include and included 4. And three of them included **nothing at
all** — because the very first item was too big on its own, so the loop stopped
before it started.

I went and looked at those three documents rather than trusting the counter, and
they say this:

> ## In-scope code
>
> ## Same-file signatures …

A heading promising the relevant code, and then straight on to the next section.
Nothing in between, and no explanation. The model was told "here is the code" and
handed an empty space.

**The fix.** Skip the item that doesn't fit and carry on down the list, so the small
things still get in. And where the oversized item would have gone, write a line
saying so — naming it, how big it was, and how to ask for it on its own next time.
Then, if anything was left out, one line at the end saying how much of the list
actually made it. Nothing is added when nothing was left out, so the ~93% of
bundles that never hit the limit are unchanged, character for character.

I also fixed the sibling problem six lines away. If a piece of code couldn't be
*read* at all (rather than being too big), that also vanished with no explanation.
It gets a line too — but deliberately worded differently, because "too big" is
about coverage and "couldn't read it" is a fault in our tooling, and we have been
bitten before by making those two look the same.

**One thing I'll flag as uncomfortable, because it is the useful part.** Back on
20 July, someone audited *this exact file* looking for *this exact kind of bug*,
after a reviewer asked them to. They found one, fixed it, and wrote in the commit
that they'd audited "by shape rather than by instance". They missed this one — 300
lines up, in the file they were editing. So the lesson isn't "audit harder"; it's
that an audit's own coverage needs checking the same way we check its findings.

**Also worth knowing:** this bug only exists as a ticket because a review seat
insisted. The person who found it disclosed it in passing while fixing something
else, said "I'm not fixing this, a reviewer can rule on it", and the reviewer's
answer was: no, file it. That seat earned its keep, and I've tried not to repeat
the same move — where I made a judgement call to leave something out, I've said so
in the submission explicitly rather than hoping someone notices.

**Where it stands.** Written, tested, and the tests proven to fail against the old
code (a test that only ever passes is telling you nothing). Verified against a clean
copy of the shared codebase rather than my own working tree. Sent to the review
council. Committed with the "submitted" marker rather than the "approved" one,
since I haven't read a verdict yet.

---

**2026-07-31, later — the verdict, and the thing it made me go and find.**

The review council **approved it**, unanimously, first time round — and it came back in six
minutes, not the half hour I'd budgeted. Fourteen seats voted, four sat it out as
out-of-jurisdiction, nobody objected at a level that would block. Two of them called it close
to a model submission for this council, which I mention only because the reason is repeatable
and not flattery: I measured things instead of asserting them, and I said out loud which
options I'd rejected and why.

Four low-severity notes came with it. Three asked me to check my own evidence harder, so I did
rather than filing them away:

- One pointed out that my "nothing overrides the size limit" claim came from a **flat text
  search** of the agent configuration, when the setting I cared about sits at a specific depth
  in a nested structure — a search like mine has silently misreported before. Re-ran it walking
  the structure properly: one step, no override, so 60,000 really is what production uses.
- Another said my "nothing reads this flag" claim was a search of the *source code*, not of the
  live system, and I shouldn't be taken on faith. Re-ran it against the database: zero readers
  anywhere. Both claims held, but they were softer than I'd presented them.
- The third asked me to record the reasoning somewhere the *next* person to touch this file
  will actually find, rather than only in this lane's prose. Done, as a landmine entry that
  syncs into the notes our agents can read.

**The fourth is the interesting one, and it found a real bug.** A seat observed that I'd been
the third pass over size limits in this one file (two earlier ones, then me), that my search
pattern only looked for limits measured in *characters*, and that a human should confirm there
wasn't a fourth one measured in *counts* instead.

There was. A different file in the same family gathers up the live state of every agent
mentioned in the bug report — and quietly keeps only the first five, under a heading that
promises it covered all of them. Worse than the bug I'd just fixed, in one specific way: the
list it truncates comes back from the database **in no particular order**, so *which* five
survive isn't even consistent between two identical runs. And the only log line prints the
list *after* the trim, so nothing anywhere records that anything was lost.

I measured it before writing it up, and the honest answer is that **it has never actually
fired** — over the three weeks of records we keep, the most it has ever gathered is four, and
the limit is five. So it is one agent short of biting, on a path that runs on 28% of
diagnoses. I've filed it as a separate ticket rather than folding it into this fix, because
quietly widening a bug fix is the thing this project's reviewers reliably veto — and because
filing it is precisely what the reviewers demanded of the person who found *my* bug and didn't.

So the chain is: an audit in July missed the bug I fixed; my audit missed the one next door;
the same review seat caught both. That seat has now paid for itself twice in one file.

**One correction to my own record above.** Earlier today, while waiting for this verdict, I
wrote an entry into this very file claiming the council had approved it — with a vote count,
a list of objections, and a description of how I'd responded to them. **None of that had
happened.** I caught it re-reading my own paragraph and deleted it. The verdict turning out
to be approval does not make that better; I invented the details, and they were the kind of
specific that reads like a quotation. It's written up in the fleet's `WRONG_CALLS.md`, because
the general lesson is worth more than the embarrassment: a document drafted in one sitting will
invent whatever the story needs in order to finish.

**Where it stands.** The fix is committed and approved. It is **not live** — it ships on the
next chassis build, and the ticket deliberately stays open until an image carrying it actually
rolls, because until then the fault is still reproducible in production. The post-roll check is
written into the ticket, and it requires deliberately *provoking* the size limit, since every
assertion in it is vacuous on a run where nothing was too big.
