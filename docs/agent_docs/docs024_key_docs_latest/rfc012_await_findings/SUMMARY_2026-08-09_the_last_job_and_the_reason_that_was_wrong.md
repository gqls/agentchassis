# SUMMARY — 2026-08-09 — the last job is done, the reason for it was wrong, and the gate said no

*A new file, never an edit of an earlier one. The series is the record.*
*Previous: `SUMMARY_2026-08-08c_the_seam_is_strict_now.md`.*

---

## What we're trying to do

When anything goes wrong anywhere in this system, we write a row into one table — a durable note
saying what failed, where, and **which agent was responsible**. That last column is the one every
investigation starts from; the table is indexed on it precisely because it is the first thing
anyone types.

This lane's job has been to make that record trustworthy. Not more detailed — trustworthy. A
record that is wrong in a way nobody can see is worse than no record, because it gets believed.

## Where we've come from

Three things, in order, over the last four days.

**One writer.** The code that wrote those rows had been copy-pasted nineteen times, and the copies
had drifted — different columns, different handling of missing values, and nine of them omitting
the very field that lets you connect an error to the job that produced it. That was consolidated
into a single writer everything now goes through.

**One door, and it refuses to guess.** The shared writer would helpfully fill in a missing
"which agent" from whatever happened to be running. Helpful when you left it out on purpose;
silently wrong when you simply forgot. That is now split: the bookkeeping fields are still filled
in automatically (nobody can get those wrong), but the *identity* fields are never inherited
unless the caller says so in as many words. Forget, and the row lands marked `unattributed` — a
word that appears nowhere else, so it is a standing alarm rather than a plausible answer.

**One ladder.** That is this week's piece, and the subject of this summary.

## What we've done

The system asks "which agent is running right now?" in two places, and the two places asked
differently. The part that stamps the owner on a whole job asked properly. The part that stamps a
name on each error row asked a shorter version and often got back the placeholder `generic` — so
errors were filed under a word meaning "somebody", while the job record for the very same run
named the real agent.

There is now **one** way to ask, it lives on the thing being asked about, and both places call it.
Not two copies kept in step by discipline — one implementation, in the only place both halves of
the code can see.

Two things about how it was built are worth stating, because they are the reason to trust it:

- **It was proved by breaking it, not by passing tests.** Every existing test in that area happily
  ignored the column in question — which is exactly why the bug survived so long. So the change was
  proved by deliberately re-introducing the fault six different ways and checking that precisely
  the right tests failed each time. One of those runs revealed that the *other* half of the system
  had never had a single test in its entire life; it does now.
- **The limits are written down.** There is one situation — a job that pauses and resumes — where
  this may not take effect. I could not test it: doing so would need records from a table that
  only keeps a day, and every affected record is weeks old. So it is stated plainly, with the exact
  count to run after the next rebuild and the number it has to beat.

## Where we are now

**Three things are true and they pull in different directions.**

**The work is done and committed.** Nothing is half-finished. Like all our Go changes it does
nothing until the next fleet rebuild, and the notes say so rather than claiming otherwise.

**The reason we did it turned out to be wrong.** The brief said this was costing us 559 bad
records. It was a real count, honestly taken, and it had already been re-checked once when
reviewers asked. It was still misleading: **499 of those 559 were written before we fixed the
other half of the same problem, three weeks ago.** Most of the "damage" was a photograph of a
wound that had healed. The live cost is about 36 records over 13 days.

That is not carelessness, and I want to be careful not to let it read as such. The count was
correct. It was re-run when challenged — which is exactly what we ask for — and **re-running the
same question later reproduces the same blind spot**, because the blind spot is in the shape of
the question. A table that keeps one month cannot tell you whether a problem is raging or was
fixed a fortnight ago: both look like the same number. The cure was one extra column, and it is
now written into three places so nobody pays for it again.

So the case for the change is weaker and truer than it was: not "this is costing us hundreds of
records", but "one question should not have two answers in two places that cannot see each other".
I have said exactly that in the paper, and listed *do nothing* as a genuine option, because the
owner should weigh the real number rather than the impressive one.

**The review board said no — and the board disagrees with itself.** The automated review rejected
it. Ten of the twelve seats approved, several of them warmly; one seat has a veto and used it, on
the grounds that the fix adds a shared piece of code that three parts of the system can see, and
its job is to be conservative about exactly that. Its suggested safer alternative is to **copy the
two lines into the second place instead** — which is, precisely, the duplication the change exists
to remove. The architecture seat said so in the same session, unprompted: the contained fix *"would
have re-created the drift risk the author is trying to close … a third site would have been next
… I'd rather see this land than not."*

I also got the scope judgement wrong myself, and have said so in the paper rather than quietly
adjusting it. I argued this belonged to the ordinary review, and offered to be proved wrong by
anyone who could name something automated that depends on the column. Nobody challenged that —
the rule that fired was a different one I had already noticed and then talked past.

**So the position is: not resubmitted, not reverted, and waiting on a person.** Our own standing
rule is that a rejection about *scope* is not answered by producing better measurements — it is a
judgement about how a change reached production, and when the reviewers contradict each other, a
human breaks the tie. The vetoing seat says the same thing itself: this decision belongs to the
architecture paper, not to the gate.

## Where we're going

Three things, and only one of them is work.

1. **An owner decision, not a task.** Which design ships: one shared ladder, or two local copies?
   The two review seats that disagree are both applying rules we wrote, and both correctly. The
   paper (`RFC_019`) lays out the disagreement with both sides quoted, and asks for a ruling.
2. **One measurement after the next rebuild**, which will settle whether the fix reaches the
   pause-and-resume case. The query and the number to beat are written down. If it does not move,
   the follow-on is a single line of code and I have named where it goes.
3. **A smaller loose end worth mentioning** because it keeps recurring: our list of architecture
   paper numbers had gone stale again — eight papers filed in a row without claiming a number,
   against a note still saying "next free number is 11" when we were on 19. It has now happened
   twice, and both times the fix was to stop relying on memory and read the directory. I have
   written that one-line command into the file and said plainly that the real answer is an
   automatic check nobody has built. A list that is only right when everyone remembers is not a
   list.

Everything else this lane owed — the three owner rulings from the RFC's second sitting — was
delivered and proved live before today.
