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

---

## 2026-09-04 — the good news nobody had read, and a signature that found a fault in what it was signing

Picking this up cold, the first thing I found was that the reviewers had **approved** our work on
the 2nd of September, at 16:33 in the afternoon — nine minutes after the last person here stopped
working — and **nobody had gone back to look.** Four rounds of review, and every single time the
reviewers sent it back it was because of how we'd *described* the change, never the change itself.
The design was never once objected to.

That unread approval had a cost beyond tidiness. Another team's work was **held waiting on us.**
They had built a related change, had it approved first time, and then stopped — because the owner
had ruled that we had to counter-sign it, since it edits a set of declarations that belong to this
piece of work. They wrote to us on the 2nd asking. Nobody was here to answer, and their own notes
had (fairly) written us off as dormant, then corrected themselves the next day when they checked
properly. **So a two-minute reading job had a built, approved, safe change parked behind it for
two days.** That is the thing worth remembering from today.

**I signed it — and reading it properly turned up a real fault, which is rather the point of
asking someone to counter-sign.**

Here is the fault in plain terms. We have two kinds of automatic check watching the live system.
One asks "is this exact wording still there?" The other asks "are there still exactly five items in
this list?" They catch different things: the first notices if something is *removed* or *altered*,
and the second is the only one that notices if something is *added*. We wrote that distinction down
ourselves, in a comment, back in August.

The other team read our comment, applied it to our own checks, and found a genuine gap — good work,
and they were right. Then, in the very next line of their instructions, they specified the
"exact wording" kind of check for a new list. **Which is the same blind spot again, one step
along.** So I built their check as written, added a sixth item to a copy of the live list, and
watched it report everything fine. Then I *removed* an item, and it complained immediately — which
is the important half: it proves the check was switched on and working, so the silence about the
addition was a real blind spot and not a broken test. The fix is one extra check of the counting
kind, and I've given them the exact line, including the one number that is easy to get wrong.

**The moral, and it is uncomfortable because it is about us:** we wrote the principle down, in a
comment, right next to the code. The other team read it, applied it correctly to our work, and
then walked into it themselves within a paragraph. **Writing a rule down does not enforce it.** The
only thing that actually revealed this was building the check and deliberately breaking things in
front of it. That is the third time this small workstream has learned the same lesson in three
sittings.

**I also checked, rather than assumed, that our own fix is genuinely running.** The reviewers had
made exactly that objection — we'd claimed the code was live based on a version number, which is
not evidence. So I asked the two live servers directly whether they contain the new code, with a
control in each direction: a phrase that must be there, and a nonsense phrase that must not.
Both servers, both controls, correct. The database changes are live too, and the check that runs
every morning is genuinely looking at all three of our declarations — I verified that by matching
what it reported against what the code actually holds, rather than trusting a clean result.

One last thing, less about this bug than about how we work. A shared piece of code has been
**broken for nine days** because of an unrelated change, and *four* separate workstreams noticed,
wrote it down, and correctly said "not ours to fix". None of them told the people whose file it
is. Everyone was being careful; nobody was being useful. I've told them today, with the fix the
error message itself suggests. Worth watching for — "I recorded it" can feel like "I dealt with
it", and here it plainly wasn't.

**The bug is closed.** The original fault — a template fix that would report success and change
nothing — is fixed and running. What remains is deliberately somebody else's: an unrecognised
instruction still completes quietly rather than being refused, and that is precisely the change I
counter-signed today.

### Same day, an hour later — I have to take some of that back

The fault I described above is real, and I proved it properly. But I told several people about it,
and wrote it into two sets of instructions, **before checking whether we already had a guard against
it. We do.** There is a test in exactly that area whose entire job is to refuse the shape I found;
it names the problem better than I did, and I confirmed it by deliberately writing the bad version
and watching the test throw it out.

So the honest version is smaller and more useful. Not *"there is a hole here you would fall
through"* — you would be stopped. It is *"when the test stops you, there are two ways out and one of
them is wrong."* You can either add the counting check, or write a short justification for skipping
it, and for this particular list skipping it would be the wrong call. Plus one number that is easy
to get wrong and that no test will catch for you.

**Why I am writing this down rather than quietly editing.** The thing I got wrong is not "I didn't
know about the test". It is that **proving something true felt like finishing the job.** The proof
was about our code; the claim I then made was about our organisation — that this would slip through
us — and I had done no work at all on that second thing. The check that would have caught it takes
about ten seconds, in a file we keep specifically for "things that will mislead you here", and the
answer was in it, naming the test.

Nothing was shipped on the strength of the overstatement and the correction went out the same hour,
to the same places. But it cost an hour of other people's attention, and I have logged it in the
file we keep for exactly this.
