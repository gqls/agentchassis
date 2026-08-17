# Where we are — the whole-page shrink floor's axis (bugs_open/293)

Plain prose, append-only, newest at the bottom.

## 2026-08-17, first session

**What the bug is, in plain terms.** When the pipeline rebuilds a page, something has to notice if
the rebuild has quietly deleted most of the writing. There is a guard that does exactly that: it
compares each section's text before and after, and refuses the whole save if a section loses more
than half of it. The problem is *what it counts as text*. It strips the HTML tags out and measures
what is left — but a page's styling instructions (CSS) and its interactive code (JavaScript) sit
*between* tags, not inside them, so all of that counts as "text" too.

The consequence is the one you would guess. If a rebuild replaces an article with a stylesheet, the
guard sees the character count go **up** and waves it through. That is not hypothetical: it is what
happened to a webdesign.co.uk article on 14 August, which served an empty page for about 23 hours.
The lane that fixed that incident fixed the *section editor* — one of the two doors this write can
come through — and filed this bug for the other, bigger door: whole-page rebuilds.

**Why they stopped rather than fixing both.** They could prove their fix was safe, because the
database keeps a copy of every section it overwrites, so they could replay 117 real edits through
the new rule and check it did not start refusing good work. Whole-page rebuilds do not leave that
trail — they delete every section and write fresh ones, so the "after" side appeared to be missing.
Changing a safety rule fleet-wide on evidence that does not cover the path is how a guard starts
blocking legitimate work, and then gets switched off. So they wrote down what evidence was needed
and left it.

**The evidence turned out to be there.** The "after" copy isn't missing — it is the row that is
*live right now*. And each live row records when it was created, which is independent proof that it
was written by the rebuild that had just deleted its predecessor. That gives 1,079 exactly-paired
rebuild writes — nine times the evidence the sibling fix had — with a check that could have gone
wrong and didn't: not one live row is older than the deletion it supposedly replaced. As a second
check I ran the same method over the *other* path and it reproduced the other lane's three known
findings to the character, plus one they had missed.

**What it says.** Two things, and they point the same way.

Switching the guard to count only the words a reader sees would have refused **none** of those 1,079
rebuilds. Across a wider, less reliable set of pairs it would have refused exactly **one** write in
eight days, and I read that one by hand: a genuine tightening of some prose on robot-hands.com,
which the operator could have let through with a config setting that already exists.

And the thing the current guard is missing is not small. Rather than wait for another incident, I
took all 1,079 real sections and *constructed* the failure — deleted every word, left the styling and
code exactly as they were, which is precisely the shape of the August incident — and asked the
guard. **It allows the total deletion of the prose on 724 of the 1,060 sections it looks at.** The
proposed measure allows none.

**Two things I did not expect to find.**

First, the guard's own cut-off is now wrong. It ignores sections under 500 characters, on the
grounds that short things shrink legitimately. But 500 was chosen when the count included all that
CSS, so on a real reading of the prose it excludes over half the sections on the page. Lowering it
roughly doubles the protected surface and — measured at every step down — does not refuse a single
additional write.

Second, there is a **third** copy of the same mistake, and it is the worst of them: an older,
page-wide version of this check that refuses a save if the whole page loses three-quarters of its
text. On the same measurement it would allow a whole-page prose wipe on 337 of 366 pages. So this is
not really a bug at one call site; it is one judgement that three places each decided how to measure
for themselves. Fixing only the one this bug names would leave the same trap for the fourth.

**And one thing I got wrong.** My first run reported a textbook hollowing on a leopardessconsulting
page and I nearly wrote it up as a finding. It wasn't real. Section names can repeat on one page,
and my query had paired one copy's "before" with a different copy's "after". Hand-checking the single
result is what caught it — the number looked completely plausible until I opened the page's history.
Worth the five minutes twice over, because the *same* wrong assumption is in the shipped guard: on
those pages it compares an arbitrary copy against an arbitrary copy, and which one wins depends on
the order the database happens to return rows in. A production defect found by tripping over it in
my own measuring instrument.

**Where this goes next.** The evidence and the measuring harness are committed — and the harness is
in the repo on purpose, because the instruction "re-run this calibration before changing it" was
sitting in the code with no way for anyone to do it. Next is the fix itself, which needs a decision
about how wide to go, and then the council.
