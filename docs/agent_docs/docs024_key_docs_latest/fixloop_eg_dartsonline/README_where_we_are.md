# Where we are — the fix-loop family

*The owner's running plain-prose log for the three workstreams that share this
directory: the **fix loop** itself, the **feature builder**, and the **council
gate**. Append-only, newest at the bottom. Never rewrite or reorder what is
already here, and never edit someone else's words — add a dated correction
underneath instead. If it reads like a pasted chat transcript, that's because it
is one, and that's fine.*

*(Started 2026-07-19 by the council-gate thread, because CLAUDE.md's standing-five
directive asks for this file and this directory didn't have one. The fix loop's
older prose log is `README_so_far.md` — left exactly as it is; this is not a
replacement for it.)*

---

## 2026-07-19 — council gate: live, adopted by three threads, and it has started catching things

The council gate is the fix loop's reviewer council opened up as a service, so
any working session can put a change through it before committing. It went live
on the 17th and it is now being used by threads other than the one that built it.

**Where it stands.** Thirteen reviewers. Two of them always sit — one checking
the edits are real and minimal, one guarding the rest of the platform, with the
only power to block. The other eleven only wake up when a change actually
touches their territory, which is what keeps this affordable at the pace people
work. A thread submits a change with its reasoning, gets a verdict in about two
minutes, and if approved commits with a one-line trailer so a report can later
say who was reviewed and who wasn't.

**Adoption happened without anyone being told to do it.** Within a few hours,
three separate threads had used it: the imagery thread submitted, got asked to
revise, revised, and resubmitted on the same correlation — exactly the intended
loop. The feature-builder thread submitted a stage-loop controller. A third
carried the trailer on a commit. That matters because the gate is advisory: it
cannot make anyone use it, so use is the only real measure of whether it earns
its place.

**It has already caught a real thing, on its first run.** Our own test
submission proposed a change that rested on an assumption about how records are
tagged. The reviewers are allowed to query the live database before deciding,
and their queries showed the assumption was simply false — the change would have
produced a permanently empty section, silently, with no error. That is the
platform's most persistent failure shape, and it was caught before a line of
code existed.

**Three things about the tooling turned out to be wrong, and I'd rather record
them than the successes.** First, the coverage report was quietly lying: a
plumbing mistake meant it stopped counting at the first reviewed commit, so it
was reporting four commits when there were forty-one, and looking perfectly
healthy while doing it. Another thread found that. Second, the same report
accused an honest commit of faking its review, because it only recognised one of
the two kinds of identifier a thread might legitimately cite. Third — and this is
the one worth remembering — the verdict that commit pointed at was *deleted*
between two runs of the report, by a documented practice that told people to
clear old council records before a fresh run. So a properly reviewed commit
became indistinguishable from a false claim. That advice is now retired (it was
also obsolete), and the report distinguishes "we can't find your evidence" from
"you didn't have any", because those are very different accusations.

**A related bug, flagged by the owner and audited by the reasoning-dataset
thread.** The fix loop's reviser — the step that rewrites a plan after the
council objects — had been receiving *blank* reviews for some time, through a
subtle templating fault. It was revising against nothing while looking like it
was working. That's fixed and now proven in live traffic. The same audit showed
a second, quieter version of the problem: the reviser only ever saw six of the
thirteen reviewers, because each new seat had to be threaded into the prompt by
hand and nobody had. Rather than list all thirteen — which would break again on
the fourteenth — the reviser now reads the council's report as a single
document, so new seats reach it automatically. That fix is applied but has not
yet had a chance to run, and I'm watching for the first one rather than claiming
it works.

**What I'd like a decision on, when you're ready.** The gate is advisory by
design and can't intercept a commit. The open question is whether it ever should
— platform changes riding branches, with only approved work merging. That
changes how every session works, so it waits for you and for evidence from the
advisory period. The coverage number is the input to that decision, and it is
currently very low: most platform commits still go unreviewed, which is exactly
what you'd expect a day in.

---

## 2026-07-20 — the new build is good, the council is being used properly, and I got something wrong

**The new chassis build checks out.** I verified it against the running pod
rather than trusting the version tag, and the two fixes that mattered to the
council are genuinely in the binary. The important one: until today, a single
reviewer writing an answer in slightly the wrong format could throw away an
entire council round — every other reviewer's work included — after all of them
had been paid for. That can't happen now; a malformed answer costs that one
seat's opinion.

**The council has grown to sixteen reviewers**, and the two copies of it are in
step without me touching anything, because the mirroring is mechanical now.
Fourteen of the sixteen only wake when a change touches their area — a real
submission today woke ten.

**It's being used, and it's earning its keep.** I put a genuine change through
it and got a "revise" back. The part worth telling: three different reviewers
independently spotted the same real flaw — a piece of SQL that would break if it
ever met a badly-formed record — and none of them was the reviewer you'd expect
to catch it. That's the argument for having a wide panel rather than one careful
reviewer.

**Now the part I got wrong, because it's the more useful half.** After
submitting, nothing happened. Thirteen minutes later there was still no sign of
my submission running, so I concluded the message had been lost, sent it again,
and spent about twenty-five minutes investigating the messaging plumbing —
checking whether the message ever left, reading the queue, looking for a size
limit. Nothing was broken. The queue was simply busy: my submission started
twenty-nine minutes after I sent it, and then ran perfectly.

What stings is that my own runbook warns about this in plain terms, and I had
quoted that warning to you earlier the same day. Knowing a rule and reaching for
it at the moment it applies are different things. The cost was real but
contained: a duplicate submission, the wasted investigation, and ten more
minutes lost to a silly bug in my own polling loop. I've logged it in the
fleet's wrong-calls ledger, which now shows this is the fourth time someone here
has mistaken "hasn't happened yet" for "didn't happen".

The practical upshot for everyone: **budget half an hour for a council
submission, not two minutes**, and don't resubmit when it seems quiet. I've
corrected that in the shared instructions, where it said two minutes.

**Where that leaves things for a fresh start.** I've written a handoff so a new
chat can pick this up cleanly. The one genuinely unfinished item is a fix to the
fix loop's reviser: it's applied and correct in configuration, but the fix loop
hasn't run since, so it has never actually executed. I'd rather say that plainly
than let it read as done.
