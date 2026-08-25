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

---

## 2026-08-24, later — what "the remaining publishers" actually is, and when this can close

**On the fresh chassis build: it changes nothing here.** Everything this lane produced is shell
and Python, which is live the moment it's committed. We shipped no Go at all, so there was
nothing waiting on a roll. I checked rather than assumed.

**Now the number, because "156 remaining" was misleading and it was my figure.**

There are 155 scripts that could still publish the unsafe way. But that number is doing almost
no useful work:

- **86 of them live in lane folders under `docs/`** — one-off scripts recording what somebody
  did once. Rewriting those falsifies the record, and worse, several are triggers that *fire
  something* every time they run, so editing and testing one risks setting off a real dispatch.
- **102 haven't been touched in a month.** Dormant.
- **Six are literal duplicate copies** — `(1)`, `(2)`, `(4)` — download artefacts.
- **Eighteen aren't publishers at all.** They only match because they contain a *warning* about
  this hazard. That includes every file I've fixed, because fixing one adds a comment explaining
  what was removed.

**The real remaining work is eleven files** — the shipped tools under `scripts/` that people
still actually run. They're listed in the handoff.

**So yes, this can close, and the bar is those eleven, not 155.** My recommendation: migrate
them, leave the commit-time detector in place so no new ones appear, and write into the bug file
that the dormant lane scripts are explicitly out of scope. Then close it. Keeping it open against
the raw number would mean it never closes while nothing actually improves.

**Today I migrated five more** — the `fire-*` operator tools. These are worth a mention because
of what they contained: `fire-brief-writer.sh` carried the line *"⚠ kcat -P EXITS 0 HAVING SENT
NOTHING. The publish is therefore not evidence"* in its own header — and then published the
unsafe way with both output streams thrown away, so there was no evidence in either direction.
Someone understood the problem exactly, wrote it down for the next person, and shipped the bug
anyway.

**And the thing I'd most want you to know:** two more lanes have picked up the shared publisher
without anyone asking them to — one committed, one in progress right now. The safe method had
been written down for a month and had two users. It became something you could *call* yesterday,
and strangers started using it today. That's the argument against ever answering this kind of
problem by writing the warning down more clearly.

**Handoff for a fresh chat:**
`docs/agent_docs/docs024_key_docs_latest/bugfix_327_silent_publish_drop/HANDOFF_2026-08-24_continue_here.md`

---

## 2026-08-24, evening — the migration is done

All eleven are migrated. The check that finds live scripts still publishing the unsafe way now
returns **nothing**, and twenty-one scripts use the shared publisher. Every one of them was
tested by pointing it at a broken address and confirming it stops with an error naming what will
not happen — none of them can now exit cleanly having sent nothing.

Three judgements I made file by file rather than by rule, because getting them wrong would have
been worse than leaving the scripts alone:

**Some of these publish from inside a loop that processes many items.** There, failing to send
one message now fails *that item* and hands the decision back, rather than tearing down a sweep
that may already have done real work on earlier ones.

**Several had the message written inline in the script.** Lifting that out into a variable is
routine, except for one detail: the quoting style controls whether `${DOMAIN}` becomes the actual
domain or the literal text. Get it wrong and the script publishes placeholders and looks
perfectly healthy doing it.

**Four more printed "save this reference number" before sending anything** — the same defect as
the original bug. Those prints now come after confirmation, and I checked they don't appear on
the failing path.

**On closing.** You told me to keep the bug open and track the class, and I then wrote down what
would close it: the live scripts migrated, the commit-time check left in place, and the dormant
one-off scripts declared out of scope in writing. **All three now hold.** I have recorded that in
the bug file but have not closed it, because keeping it open was your instruction and reversing
that is your call, not mine.

What would remain after closing is genuinely not work: about 55 scripts under `scripts/` that
still contain the old pattern but haven't been touched in over a month, plus the lane one-offs
that should never be rewritten. If any of them is ever picked up again, the commit-time check
will flag it then — which is the right moment, rather than us rewriting a hundred dormant files
today on the chance that one of them gets used.

---

## 2026-08-25 — nothing left but three decisions

Checked the state fresh this morning rather than trusting yesterday's: no live scripts still
publish the unsafe way, twenty-one use the shared publisher, and nobody has touched my files
overnight.

**On the new chassis build — it makes no difference here, and I re-checked rather than repeating
myself.** Every change this lane made is shell or Python, which goes live the moment it's
committed. Not one of them touches the Go code that a build produces. So there was never anything
waiting on a roll, and there is nothing to re-verify after one. If anyone asks whether this has
shipped, the honest answer is that the question doesn't apply to this work.

**The automatic checker looked at my two hazard entries and returned "needs human review".** That
sounds worrying and isn't: it means the tool *couldn't* check, because its index only covers Go
files and everything here is shell scripts. I did the check by hand instead and wrote the results
into both entries, so they no longer sit there looking unverified.

**Doing that turned up something I want to flag, because it's the third time.** I checked that the
council trigger now prints its reference number *after* sending rather than before — and my check
said it didn't. It was wrong. It had found the phrase inside a *comment* I'd written explaining
the old behaviour, not the actual line, which is further down and correct.

That is the same mistake three times in four days: **in a piece of work whose subject is a trap,
my own writing about the trap keeps fooling the tools I build to find it.** First the count of
remaining work, which didn't move as the work got done because fixing a file adds a comment
mentioning the problem. Then the completion check, which flagged two files I'd just fixed. Now
this. Each time I built a fresh one-line check and didn't think of it as a measurement at all.

I've written it up as a tally rather than three separate notes, because the repetition is the
point — and the fix is a habit in how the check is written, not another thing to know.

**What's left is three decisions, and no work.** Whether to close the bug (its stated conditions
are all met, and I'd close it); whether to spend credits on a council review now that it's
eligible (I'd genuinely leave that to you — it would review the detector, not the publisher
itself, which is still outside scope); and whether to build the stronger in-cluster version (I'd
wait, unless we ever see a message lost *through* the new publisher, which would be the evidence
that the current receipt isn't enough).

Handoff, updated with today's state at the top:
`docs/agent_docs/docs024_key_docs_latest/bugfix_327_silent_publish_drop/HANDOFF_2026-08-24_continue_here.md`

---

## 2026-08-25b — you're closing it; here's what I've left you

**A closing summary is written.** I checked whether it was warranted rather than assuming: the
test is whether the five headings would say something genuinely different from the last one, and
two of them do — "where we are now" goes from a queue of eleven with an open bug to finished, and
"where we're going" from a migration plan to two optional improvements. The deciding argument is
simpler though: **the newest summary is what someone reads months later**, and yesterday's says
eleven files remain. Once they don't, that's actively misleading. The earlier two are untouched;
the wrong turns live in them and that's the part worth keeping.

**The two open questions now have their own file**, deliberately outside the bug, because a
question left inside a closed bug is exactly how one gets forgotten.

**Two things bite when you move the file, and neither is obvious.**

The smaller one: 66 files mention the path `bugs_open/327`, and all of them point at nothing the
moment it moves. My advice is to leave them alone — they're explanations of why some code looks the
way it does, not instructions anyone will follow, and a search for "327" still finds the bug in its
new home. Rewriting 66 files to fix a cosmetic path would be the worse trade.

The one that actually matters: **both bugs numbered 327 will then be in the same folder.** Right
now you can tell them apart by where they live — the open one is the dispatch drop, the closed one
is the spec-write bug. After the move that distinction is gone, and someone has already confused
the two once in a live handoff. Worth putting a line in the closing commit message saying which
327 it is, because that's where the next person will meet it.

Both points are written into the handoff as a closing checklist so you don't have to hold them in
your head.

---

## 2026-08-25c — your three answers, and one correction to something I implied

**Renamed to 327b.** Done, with `git mv` so the history follows. There was already a precedent —
`347b` — and 347 turns out to be one of **thirty-six** numbers in this repo naming two unrelated
bugs, so this is a small instance of a wide problem rather than a one-off tidy-up. I renamed ours
rather than the other one because this lane is still active to fix the references and the other is
closed and isn't mine. The number wasn't reassigned — nothing ever held 327b — so this doesn't
breach the never-renumber rule; the suffix only tells them apart.

I updated the four files that named it by its full filename and deliberately left the sixty-six
that just cite "bugs_open/327". Those are explanations of why some code looks the way it does, not
instructions anyone will follow, and a search for "327" still finds it.

**Yes — you can call the publisher onto the council whenever you want it.** `FORCE=1` overrides the
scope refusal for a single submission. So the publisher doesn't need to be permanently in scope to
be reviewable; it can go to the council the day there's a reason to. That's the better arrangement
than widening the rules, because a widening taxes every future change in that directory, whereas
this costs exactly one round on the day you choose.

**On what we're saving — I should correct something I implied.** I'd left the impression it was
mainly about time. It isn't.

It's **mostly credits.** An out-of-scope submission is refused before anything is spent, and even
an accepted one doesn't pay for all seventeen reviewers — most only wake up if the files you
touched match what they care about.

Second, and smaller, it protects **a reviewer's usefulness over time**. The argument on record for
keeping prose out wasn't cost — it was that a reviewer firing constantly on things it can't judge
becomes one people stop reading.

**Time is barely a factor**: a round averages about nine minutes.

And your other guess — too many members arguing — **isn't how it works**, which is worth saying
plainly because it's a reasonable thing to assume. Reviewers who aren't relevant don't take part
at all, so an extra submission doesn't add voices to a debate. More seats firing means wider
coverage, not more argument.

**Decision 3: waiting, as agreed**, with the trigger written down so nobody reopens it without
one — a message observed lost *through* the new publisher.

---

## 2026-08-25d — closed

`327b` is in `bugs_closed/`. Git recorded it as a rename rather than a delete-and-add, so the
history follows the file, and I checked the repository itself rather than my own working copy —
which is the mistake I made on the earlier rename, where a half-committed move left both filenames
live for everyone but me.

The closing note on the file says what was verified at the moment of closing rather than what was
believed: no live script still publishes the unsafe way, twenty-two now use the shared publisher,
and every migrated script was tested by pointing it at a broken address. It also says plainly that
what remains — about fifty-five dormant files and the lane one-offs — is data rather than a defect,
because the commit-time check catches any of them the moment someone picks one up. **The class is
held by a detector, not by us finishing a sweep**, and that distinction is what makes closing
honest rather than optimistic.

The commit message names *which* 327 it is, in the subject. That sounds fussy and isn't: both bugs
numbered 327 now sit in the same folder, someone has already confused them once, and the commit log
is where the next person will meet the pair.

One last thing I found while auditing, and it's the kind of gap worth admitting: I had written the
case file and two hazard entries, but not the entry in the shared debugging guide — the one a
person reads when they *have* the symptom and no idea why. The two I'd written serve someone about
to touch a file. Someone staring at a dispatch that vanished would have found neither. That's
written now, and it's the piece most likely to save somebody a day.

**Nothing is outstanding.** Two decisions are deliberately open, recorded with the triggers that
would reopen them, in `HANDOFF_2026-08-25_open_decisions.md`.
