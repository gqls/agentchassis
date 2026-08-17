# SUMMARY — the page-save content guards, 2026-08-17

Current state only. Written to be read aloud.

## What we're trying to do

Stop the site pipeline from quietly deleting a page's writing. When it rebuilds a page it replaces
every section, and something has to notice if the replacement has thrown the words away. Three guards
exist to do that. We're making sure they measure the right thing, and that a fourth guard written
later can't measure the wrong thing again.

## Where we've come from

On 14 August a webdesign.co.uk article was replaced by a stylesheet and an empty container, and the
site served that empty page for about twenty-three hours. Two guards were watching the exact write it
came through, and both let it past — because of *what they counted as text*. They stripped the HTML
tags and measured what was left, and a page's styling instructions and its interactive code sit
between tags rather than inside them. So all of that counted as writing. The article was replaced by a
stylesheet, the character count went **up**, and both guards read the change as growth.

Another thread fixed one of the two guards that morning and stopped there, deliberately and for a good
reason: it could prove its fix was safe, because the database keeps a copy of every section that gets
overwritten, so it could replay 117 real edits through the new rule and check it wasn't about to start
blocking legitimate work. Whole-page rebuilds don't leave that trail — they delete everything and write
fresh — so the "after" side looked as though it simply didn't exist. Changing a safety rule on evidence
that doesn't cover the path it protects is how a guard starts refusing good work and then gets switched
off. So they wrote down what evidence was needed and handed it on. That's the bug we picked up.

## What we've done

**Found the evidence, and it was there all along.** The "after" copy isn't missing — it's the row
that's live right now, and each live row records when it was created, which proves it was written by
the rebuild that had just deleted its predecessor. That gives 1,079 exactly-matched rebuild writes,
nine times what the earlier fix had, with a check that could have gone wrong and didn't. Running the
same method over the *other* path reproduced the other thread's three known findings to the character,
and turned up a fourth they'd missed.

**Made the argument the evidence actually supports, which wasn't the one expected.** Switching the
measure would have refused none of those 1,079 rebuilds — but it would have *caught* none either,
because no rebuild in the recorded window happened to gut a page. So "it would have caught X" was not
available, and saying it anyway would have been dishonest. Instead we constructed the failure on all
1,079 real sections — deleted every word, left the styling and code exactly as they were, which is
precisely the shape of the August incident — and asked each guard. **The current one allows the total
deletion of the prose on 724 of the 1,060 sections it looks at.** A second, completely different
measurement agrees within one percent.

**Fixed all three guards, not the one the bug named.** A third copy of the same mistake turned up on
the way, buried as an unnamed block of code inside a much longer function, and it was the worst of
them: it would allow a *whole-page* wipe on 337 of 366 pages. It now has a name, a test, and a switch
an operator can turn off without rebuilding the software — which it had never had.

**Corrected the cut-off, which nobody had noticed was wrong.** The guards ignore sections below a size
threshold, because short things shrink legitimately. That threshold was set when the count included
all the styling, so on an honest reading of the prose it was excluding more than half the sections on
a page. Lowering it roughly doubles what's protected and, measured at every step down, doesn't refuse
a single additional real write.

**Found and fixed a second defect by tripping over it.** Section names can legitimately repeat on one
page, and both guards compared an arbitrary one of them against an arbitrary other, decided by the
order the database happened to return rows in. On one of the two that isn't just a blind spot — it can
**block a legitimate rebuild outright**.

**Put the measuring instrument in the repository.** The code said "re-run this calibration before
changing it" and there was no way for anyone to do that: the run behind the earlier fix lived in one
session's scratch directory and survived only as prose. It's now a command, it measures with the real
functions rather than an approximation, and it can size a guard's blind spot *before* an incident
rather than after.

**Put it through the council twice.** Round one came back REVISE on one question: does the page-wide
guard's filter actually match any rows? There's a known trap where the equivalent column on a
neighbouring table never holds the value we were filtering on — and if that applied here, the guard we
had just carefully extracted, named and tested would never run, with the test suite confirming it
worked. One query: it doesn't apply. Round two approved it, with three comments, and two of them
independently pointed at the same thing and were right — so the design got simpler rather than just
defended. Nine deliberate mutations of the code confirm every new check actually bites.

## Where we are now

Everything is committed and council-approved. It is Go code, so **none of it is doing anything yet** —
it starts working when the next chassis image is built and rolled.

Checking that turned up something worth knowing: **the sibling fix from this morning, which everyone
believes is live, isn't.** The running image was built yesterday and contains neither half. So both
halves of the correction will start working at the same moment, on the next roll — which is tidier
than it sounds, because they share a threshold and it's better they arrive together.

The bug stays open until then, because our own bar is "fixed **and** live", and until it rolls the
defect is still reproducible on every page rebuild.

## Where we're going

Three things, in order. Roll a chassis image and confirm at the running binary rather than at git that
the fix is in it. Then deliberately trigger a refusal on a page we're willing to leave alone, and check
the artefact — that the page's stored sections are byte-identical afterwards — plus the opposite arm,
that an ordinary legitimate save of the same page still succeeds, because a guard that refuses
everything also "passes" a test that only looks for refusals. Then watch a week, alongside a control
that confirms rebuilds are actually running, since a count of zero refusals means nothing if nothing
was being guarded.

We expect roughly one refusal a week, and we know what the first false one will look like, because we
already found it in the history and read it: a genuine tightening of some prose on robot-hands.com. If
that shape starts appearing often, the ratio is wrong for this path rather than the measure, and an
operator can change it in the database without a rebuild.

Two things we've left undone and written down rather than quietly dropped: about a tenth of sections
are still too small for the text guards to judge, with only the layout guard covering them; and the
platform's older coverage test still can't see this write path at all, because it looks for a kind of
database update this path doesn't use.
