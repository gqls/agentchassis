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

## 2026-08-17, later the same day — the fix is written, and the council found something

**The fix, in plain terms.** All three guards now count only the words a reader would actually see.
The cut-off below which they don't bother judging a section moved from 500 characters to 200, because
500 was chosen back when the count included all the styling instructions — on an honest reading of the
prose it was excluding more than half the sections on a page. And the oldest of the three guards, the
page-wide one that was buried as an unnamed block of code in the middle of a much longer function, now
has a name, a test, and a switch an operator can turn off without rebuilding the software. It had
never had one.

**I asked Fable to plan it before writing anything, and that was worth doing.** It found a fourth
place the same mistake lives — one I had missed — and that find changed the design. I had intended to
simply lower the 500 to 200 where it was defined. But something else uses that number for a completely
different purpose: deciding whether two sections are similar enough to be paired up. Lowering it in
place would have quietly changed that behaviour with no evidence behind the change. So the number
became a setting passed in by each guard, and the old one stayed exactly where it was.

It also talked me out of what I thought was the tidier fix. I wanted to make the shared decision
measure the text itself, so no caller could get it wrong. That would have broken the measuring
harness, which works precisely because it can feed the real decision *either* measure and compare. So
the check is now a test that fails the build if any caller measures the wrong way — a test can ask
"was this measured properly?"; a type cannot.

**Then the council reviewed it and sent it back, and the reviewer was right.** The blocking question
was one line long: *does the page-wide guard's filter actually match any rows?* It selects sections
marked "deployed" — and there is a known trap recorded in our own notes where the equivalent column on
a neighbouring table never holds that value at all. If the same were true here, the guard I had just
carefully extracted, named, tested and documented would never run on any page, and the test suite
would have cheerfully confirmed it worked, because tests hand the code whatever rows you tell them to.

It was one query. "Deployed" is by far the commonest value — 1,575 sections across 617 pages, 85% of
all pages — so the trap doesn't apply here and nothing needed changing.

**But the lesson is the part I want to record, because I got it wrong in a way that felt careful.** I
had kept that filter *deliberately* unchanged, and said so in the code, on the grounds that quietly
changing which sections a guard looks at is exactly the sort of unreviewed behaviour change that
causes incidents. That reasoning is sound. What I never did was check whether the filter matched
anything in the first place. "I didn't change it" protects you from *introducing* a fault and does
nothing whatsoever about *inheriting* one — and moving a line of code into a function with a proper
name, a paragraph of explanation and a test suite makes it look far more trustworthy than it did as an
anonymous fragment, without adding a single piece of evidence. Every figure I had published about that
guard rested on a number I had not counted.

**Three other things the reviewers asked for, all now done.** The page-wide guard, when it can't read
the page at all, lets the save through rather than blocking it — inherited behaviour I chose not to
change. A reviewer pointed out that it did so *silently*, which means a future content loss would be
diagnosed as "the guard should have caught that" when in fact it never ran. It now leaves a record,
under a label that says plainly nothing was blocked. Another asked whether our idea of "visible text"
matches the one the page assembler uses when it decides a section is empty and can be dropped — a fair
worry, because if they disagree, a save could pass this guard and have its content thrown away
downstream anyway. They *do* differ in how they're built; I measured 6,585 real sections and none of
them could be dropped that way, and there's now a test that keeps it so. The third was that I had
oversold my own test as "closing" a gap when it only narrowed one; I've corrected the wording and made
the test do the broader thing it was claiming.

**Where this stands.** Everything is committed. It's Go code, so none of it does anything until the
next chassis image is built and rolled — and a thing worth knowing came out of checking that: the
sibling fix from this morning, which everyone believes went live, has *not*. The running image was
built yesterday. So both halves of this correction will start working at the same moment, on the next
roll. Round two is with the council now.
