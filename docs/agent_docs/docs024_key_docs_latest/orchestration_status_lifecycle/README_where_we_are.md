# Where we are — orchestration status lifecycle

Plain-prose log for the owner. Append only; never rewrite or reorder.

---

**19 August — what this was, in one page.**

Three days ago a bug was filed saying that a job which stopped in a particular internal state would
never be cleaned up by anything, ever. The oldest one found had been sitting there nineteen days,
through several fleet restarts, and each one was quietly holding on to two Kafka topics it would
never release. That mattered beyond tidiness: there is a separate open bug about the message system
running out of memory because there are too many topics, and this was feeding it.

The fix itself was small — a cleanup job goes round every three minutes failing jobs that have
obviously died, and it worked from a list of states it knew about. One state was missing from the
list. I added it.

**The interesting part was the check the bug file told me to run first.** It said, in bold, to
re-run a particular measurement because that measurement was what made the four-hour cutoff safe.
I ran it, and it had gone useless: it now gives the same answer whether things are healthy or
broken, because the evidence it counted had been cleared the week before. Following the instruction
to the letter would have given me a green light that could not possibly have come out red. So I
went and read the code instead and found something better and permanent — the state is only ever
set on one line, by one caller, and the very next thing that code does is move the job out of it
again. It is meant to last milliseconds. Four hours is not close to risky.

**Then the same gap turned up one state over**, and it turned out another session was working on it
at the same moment. We produced character-for-character identical fixes without knowing about each
other. Theirs was retired as the duplicate. The one difference between us is worth recording: they
reused my reasoning about the first state and were wrong by a factor of about a thousand, because
the second state genuinely does sit and wait for a message, so its duration depends on how busy the
system is. I measured it rather than reasoning about it — five and a half thousand cases, longest
ever six seconds.

**That is when it became clear we were fixing instances of a defect rather than the defect.** The
cleaner worked from a *list*, so any state nobody thought to list was invisible to it for ever. So
the list is gone. It now works from a rule — anything that is not finished, is not deliberately
waiting, is waiting for nothing, and has not moved in four hours. I proved it by inventing a state
that has never existed, planting a job in it, and watching the real cleaner reap it while leaving a
healthy one alone. A second cleanup job had the same blind spot and now follows the same rule.

**And today, the last layer.** Even after all that, the question "which states count as finished"
was still written out by hand in two places. So there is now a single table that says what every
state means, and the database itself refuses to write a state that is not in it. I tested that both
ways: an invented state is rejected with a clear message, and the ordinary states carry on
untouched. Forty-five jobs ran through in the ten minutes after, all fine.

**One thing that nearly went badly, because it is the kind of mistake worth telling you about.**
Only five states exist in the table right now, because the earlier fixes cleaned all the others out.
If I had built the new table from what was *in* the database, rather than from what the *code can
write*, then the next job the system created would have been rejected — every new job, immediately.
I caught it because I went looking for the writers in the code rather than trusting the data in
front of me, and I put a guard in the migration that refuses to run if anything is uncovered.

**On the "pause for human approval" feature you asked me to deal with:** it does not exist. It is
declared in five different places, implemented in none of them, and the two halves of it disagree
about its own name — so the four guards written to protect a paused job could never have matched
one. Nothing has ever used it. I have removed the dead declarations and left a note saying exactly
how to do it properly if you ever want it: one row in the new table, and the cleaner will
automatically leave such jobs alone without anyone changing the cleaner.

**Separately, I re-checked the memory bug you asked about.** Its symptom is gone — topics are down
from twenty-five thousand to five, memory is at 35Mb against a 256Mb ceiling, and it has not
restarted. But I have deliberately *not* closed it, because that is a judgement about whether the
underlying cause was fixed, and the file itself records that the main candidate fix was tried and
refuted. What I would suggest is that its headline stops announcing a live emergency that ended
nine days ago.

**What I would still like a steer on** is in the summary of the chat above — mainly whether the
"one table per status column" idea should spread to the other eight tables that have the same
problem in a milder form, or stay where it is.

**No summary document yet, deliberately.** The house rule is that summaries are milestones, not a
diary, and writing one this morning and another this afternoon would bury the one that matters. The
right moment is once the council has ruled on today's change. The other four working documents now
exist and I have written up the plan, the runbook and the running notes — including, honestly, the
half-dozen places I got something wrong and what caught each one.
