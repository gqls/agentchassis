# Where we are — bug 327, the build trigger that can publish nothing

Plain-prose log, append-only, newest at the bottom.

---

## 2026-08-23 — picking the bug up

The bug is this. There is one command a person runs to start a whole site build for a new
domain — the thing webdesign.uk actually sells. Sometimes you run it, it prints all the
right identifiers, it exits cleanly, and **nothing whatsoever happens.** No build starts.
Nothing is queued. There is no error anywhere. It looks exactly like the successful case,
because the identifiers it prints are ones the script made up itself a moment earlier — it
prints them before it has sent anything, and it never checks whether the send worked.

The lane that filed it saw one submission in three disappear this way.

### First job: is it still true?

Yes, and I could establish that without touching the cluster. The script hasn't been edited
since 30 July — three weeks before the bug was even filed — and it still does the exact
thing that causes the problem. Nobody is working on it either: the bug file has one commit
in its whole life, the one that created it.

A caution for anyone reading the git history: **there are two bugs numbered 327.** The
other one was closed today by a different lane, and almost every commit mentioning "327"
belongs to that one. They are unrelated. I've noted this everywhere it could mislead.

### Why it happens

It's a known trap, already written down. The script sends its message by starting a
throwaway container and feeding the message in on the container's input stream. Those two
things race each other. If the container gets going before the input is connected, it sees
an empty input, decides there's nothing to send, and exits successfully. The container is
then deleted, taking the evidence with it. When this was first measured last month, **four
sends in five were being lost.**

### Something I nearly got wrong, and it's worth recording

My first instinct was to re-check the original evidence in the database. I did, and it
agreed with the bug file — nothing there for that submission. I was about to write "still
confirmed."

It would have been meaningless. That table only keeps about **two days** of history, and
the incident was five days ago. There is nothing in it for that entire date, for anything —
successful builds included. The query gives the same answer whether the message arrived or
vanished. I caught it by asking the table what it still remembers, which took about fifteen
seconds, and I've written it up properly in the missteps log.

The same habit paid off immediately afterwards. There's a rival explanation for this kind
of disappearance — the message *was* sent, but the system rejected it on arrival — and
another lane got caught out by exactly that three days ago, blaming the sending step for
something that had been recorded as a rejection all along. So I checked the rejection
records. Nothing about our submission. This time the silence means something, because that
particular log keeps a **month**, and on the day in question it recorded 3,761 entries
including a genuine rejection of the very type I was looking for. The recorder was awake
and it saw nothing. **So the message really was never sent.**

That contrast is the whole lesson: two identical-looking zeroes, one worthless and one
decisive, and the only way to tell them apart is to check whether the instrument could have
said anything else.

### The bit that turns a bug into a project

I went looking for how widespread this is, and the numbers are the reason I want to fix the
framework rather than the one script.

There are **218** scripts in this repo that send messages this way. **200** of them use the
racing form that can silently send nothing. Twenty-five have had the documented fix applied
— they ask the sender to shout "PUBLISH_OK" when it really has sent something.

**But only two of those twenty-five actually check for that shout.** The other twenty-three
print it and carry on regardless. So the receipt exists, scrolls past on screen, and the
script still exits as though all is well. The fix was applied in letter and not in effect,
twenty-three times. That's not carelessness by twenty-three people — it's a sign that the
remedy was written as advice to copy rather than as a thing you can call.

There's also a nasty asymmetry in what we can find out afterwards. If you want to know
"was my message rejected?", you can still ask a month later. If you want to know "did my
message ever arrive?", the answer is gone in about two days. So this class of failure
becomes permanently undiagnosable almost immediately — which is a strong argument for the
sender proving itself at the time, rather than us getting better at investigating later.

And it is current, not historical: **yesterday** another lane lost a council dispatch to
the identical fault, waited ninety minutes for a run that was never going to start, then
re-sent it and got a result in three seconds.

### Where I've got to

Evidence is gathered and the diagnosis is solid. I've asked for a design that fixes the
framework — a single shared way to send a message that cannot fail silently — rather than
patching the one script the bug happens to name. Two things have to be told apart that
currently can't be: *the message never left* (re-send at once) and *the message left but
nothing picked it up* (re-sending will only make a duplicate). Today both look identical,
and the advice we give sessions for one is the opposite of the advice for the other.

One constraint worth flagging early: the review council only looks at certain parts of the
codebase, and plain scripts aren't among them. So if the fix is only a script, **it cannot
be put through the council at all** — that's a fact about the gate's reach, not a reason to
avoid review, and I'll say plainly which parts of what we do end up reviewable.

Next entry will record the plan and what I decided to do about that.

---

## 2026-08-23, later — built, proven, and three things that went differently than expected

The fix is in and working. The build trigger now refuses to pretend. Point it at a broken
address and it stops, tells you nothing went out, and hands you a retry command that actually
works. It no longer prints "save this reference number" before it has tried to send anything —
that line used to appear three lines *above* the send, which is why the failing case read
exactly like the succeeding one.

Underneath it is a shared publisher any script can call. That was the real decision. The fix
for this has been written down for a month; twenty-five scripts copied it and only two made it
work. When guidance is followed and still doesn't take, the guidance isn't the fix — so we
made it something you call rather than something you copy.

**Three things I want to say plainly, because they went against me.**

The first is that **I could not reproduce the original failure.** Ten out of ten old-style
sends arrived on the day I tested. That kills the four-in-five loss rate seen last month, but
it only proves today's rate is under about a quarter — it doesn't prove the new way is better
in a head-to-head. What I can say is that the new way cannot have the fault at all, because it
doesn't use the mechanism that breaks. That is a stronger claim than winning a race, but it is
a different claim, and I'd rather say which one I have.

The second is that **the old method sent one message twice** in that same test. Nobody
predicted that. A duplicate on the real system means two builds competing over one site. One
observation, not a rate, and written down as such.

The third is that **I nearly recorded a meaningless result as a verification.** My first move
was to re-check the original evidence in the database; it agreed with the bug report, and I was
a keystroke from writing "confirmed". That table only keeps two days of history and the
incident was five days ago — it holds nothing at all for that date, for successes or failures
alike. Fifteen seconds of checking caught it. The same habit paid off straight afterwards when
it let me rule out a rival explanation properly, so the lesson isn't "distrust zeros" — it's
that a control is the only thing that tells you which kind of zero you're holding.

**Two things I found that weren't what I went looking for.**

A handoff written today, steering a live build, said this bug was already fixed. It wasn't —
it had been confused with a different bug that happens to share the number 327. I've added a
dated correction where the claim sits, without touching the lane's own words. Their practical
advice was already sound; only the status line was wrong.

And the tool that checks our own landmine documentation publishes using the exact pattern we
were fixing, then reports "0 failed to publish" — a number computed from the one signal that is
always missing when a message is silently lost. It is the next thing to fix.

**On review:** I put the change to the council and it declined, correctly, because the code
lives in a directory the council doesn't cover. I've recorded that rather than overriding it —
overriding spends fleet credits against a rule you set, which is your call and not mine. Worth
knowing: the file holding our 22 commit-time checks is invisible to the council for the same
reason, which is the identical gap you closed for a different file today.

**What's left.** One script of about 178 is migrated. I'm deliberately not sweeping the rest:
bulk-editing old one-off scripts risks firing live triggers, and about two dozen of them turn
out to be notes files that cannot run at all. There's a better end state we've costed but not
built — sending from inside the cluster, where the message broker itself confirms receipt.
That would make a silent loss impossible rather than merely visible.

---

## 2026-08-24 — your four decisions, done

**The council now covers `pattern-check.py`.** That's the file holding the 22 checks that run
on every commit in every session. The measurement that justifies it: 2,058 lines in that one
file, against 2,220 lines for every other audit script in the repo put together — so it is
about half our checking machinery, and the half that runs most often. It is now reviewable;
the rest of `scripts/` deliberately isn't, so this didn't loosen a whole directory.

Proof it took, rather than my word for it: this lane's own submission was refused yesterday
and is accepted today, and a commit of mine that was previously in *no* bucket of the coverage
report now shows up as unreviewed.

**One thing here is worth knowing about**, because it nearly bit me. Widening the scope means
editing two files, not one — the coverage report has to list candidate commits before it can
judge them, and it keeps its own separate copy of the scope for that. Get only one of them and
commits don't appear as unreviewed, they simply vanish from the report. That happened on the
23rd and hid 22 commits across four lanes.

I only avoided it because the lane that got caught wrote the warning next to the *scope
definition* rather than in the report — reasoning that a warning in the report gets read by
whoever edits the report, who isn't the person with the problem. I was the person with the
problem, one day later. That's the system working.

**The landmine verifier is fixed.** This is the tool that checks our own hazard documentation,
and it was reporting "0 failed to publish" using the one signal that's always missing when a
message is lost. It now exits with an error and says "no verdict will ever arrive for this
run". Its caller needed no changes — it was already counting failures correctly, it just never
had a real failure reported to it.

I proved it on a live dispatch that was also useful: it armed the verification for the entry I
edited yesterday, which is the thing that silently didn't happen.

**I did not run the council**, as you asked. Worth noting the widening changes the nature of
that choice: this lane's submission is now legitimately in scope rather than needing an
override, so it's a straightforward question of whether the credits are worth it, whenever you
want to decide. Nothing is waiting on it.

**The bug stays open and now says so explicitly** — it tracks the class, not the one incident.
Two of about 178 scripts are migrated. Next is the council trigger itself, which needs a little
care because its exit codes 1 and 2 already mean specific things.
