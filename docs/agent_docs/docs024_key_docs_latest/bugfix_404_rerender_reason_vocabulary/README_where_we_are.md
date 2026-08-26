# Where we are — a word that means two different things

Plain prose, append-only, newest at the bottom. The owner maintains this too — add below,
never rewrite.

---

## 2026-08-26 (evening) — the silent way for a fix to do nothing

When we re-publish a page we can do it two ways. The cheap way re-ships the page exactly as it
was last built. The expensive way rebuilds each section from the stored content, picking up any
template or content fix. Which one happens is decided by a single word attached to the job — a
"reason".

The system keeps a list of the reasons that mean "do it the expensive way". **The trouble is that
the list was written down in three places**, and on the 18th of August two different people each
added a word to one copy and not the others. So one part of the system knew five words and
another knew three.

**Here is why that matters more than it sounds.** If a part of the system meets a word it doesn't
recognise, it quietly falls back to the cheap way. Not an error — the job completes, the status
goes green, and the page is re-published exactly as it was. So a template fix routed through the
wrong door does nothing at all, and reports success. If the fallback had gone the *other* way
we'd have noticed within a day, because we'd be doing far too much work. Falling back to "do
nothing" is invisible by design.

**What I've done is make the list exist once**, in the place this codebase keeps declarations
about live settings, with a check that runs every morning comparing what we've declared against
what the live system actually does. Every part of the system that reads that list now refers to
the one copy — so if someone retires a word, the build breaks in every place that used it,
instead of one of them silently changing behaviour.

I also fixed something adjacent I found along the way: the component-fix tool was filing
re-publish jobs for **retired** pages, which would put them back on the internet. That's the same
class of problem I fixed at a different seam earlier today. 31 retired pages were exposed to it.

**Three things I got wrong today, and they're all the same shape.**

Test files here carry a table saying "break this line and that test will fail" — the evidence that
each safety catch really works. I've written three of those tables today and **every one had rows
that were false when I wrote them.** Each for a different reason: the first pair were tripping over
an unrelated problem; the second pair used examples that couldn't produce the failure at all; and
today's couldn't *read* the thing it was checking — it scanned our database-change files for a
word, and matched only the explanatory comments, never the actual instructions, because the two
are written with different quote marks. It reported "12 things checked" and passed, and all 12
were comments.

The common thread is that none of them was visible by reading. All three were only found by
actually breaking the code and watching whether the test noticed. I've written that up, because
one instance is carelessness and three in a day is a habit worth naming: **a passing test proves
the current code is acceptable to that test, and nothing else.**

**Lastly, the reviewers.** This one came back needing revision — the first time today — and both
of the things they caught were real. One pointed out I'd chosen a database filter without first
checking what values that column actually holds, which is a rule this place has for good reason;
I'd assumed and they were right to insist. The other was sharper: I'd left two parts of the system
still using their own private copy of the list, arguing it was out of scope. Two reviewers
independently pointed out that "fix one copy properly and leave its sibling alone" is *exactly the
problem I was fixing*, one level up. That was the plan's own argument used against it, and they
were right. It's done properly now.
