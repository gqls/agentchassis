# Where we are — architecture review

Plain-prose running log, newest at the bottom. Append; never rewrite or reorder.

---

## 2026-07-27 (late afternoon)

The thing that was owed is done. The design-stage council — the one that reviews
a *feature* before it gets built — had been left the worst-equipped of the three
councils, which was an accident of plumbing: there is a script that copies
reviewer changes from one council to another, and it only spans two of them. The
design council isn't on that path, so improvements kept landing everywhere except
the place we'd argued mattered most. It now has the same three things the others
got: the council can read its own past verdicts, its safety reviewer is told to
check whether it has already pushed this same problem upstairs before doing it
again, and its bug historian has an index of our own case files.

I checked a lot before writing that, because config changes go live instantly
here and there's no undo. The short version: the change touched three blocks of
reviewer instructions and nothing else — no steps added or removed, no change to
who holds the veto, and the rollback file was verified to match what was actually
live rather than what I assumed was live.

**We have our first real evidence, and it's encouraging.** A council ran at
14:18, the first one after the change went in. Three things happened that
wouldn't have before:

- One reviewer caught the plan claiming "that's the complete set, no sixth case
  exists" and rejected it by pointing at **two of our own logged mistakes, by
  date** — both times we'd previously claimed something didn't exist without
  actually searching for it. That's the whole idea of this workstream working
  exactly as intended: our written-down errors being used to catch a new one.
- Another cited three specific case files by number when explaining that the fix
  was treating a symptom rather than the mechanism.
- The safety reviewer thought about whether to push the fix upstairs, and
  reasoned its way to *not* doing so — saying the repetition was evidence of a
  genuinely scattered bug rather than evidence the fix belonged at a higher
  layer. That's the exact judgement it previously had no information to make.

**One honest caveat about my own scorecard.** My report counted that last one as
a miss, because it measures whether the reviewer *quoted a past verdict*, and
this one reasoned about repetition without quoting anything. So the number is
harsher than the behaviour deserves. It's one data point either way — I'm not
claiming a result yet — but I'd rather flag that the metric undercounts good
behaviour now than have the numbers look tidier later than they are.

**The new forward-looking reviewer still hasn't spoken, and I now understand why
that will take longer than I said.** It only exists on the design council, and
that council refuses to run on anything without an owner-approved spec. There are
five such specs in total, two approved, and **both are already being worked by
other threads.** So it isn't a matter of waiting a few hours — the first review
will come when one of those threads runs its next round, or when you approve a
new spec. Nothing is broken; I'd just been wrong about the timescale, and the
next person reading the report shouldn't see a zero and assume a fault.

**A near miss worth telling you about, because I nearly spent your money on it.**
Wanting to see the new reviewer actually speak, I went looking for a spec to run
it against. One looked completely free: its status said "deferred", and searching
every project document for its subject turned up nothing at all — no owner, no
notes. Then I opened the spec itself, and it contains three rounds of review
history and a fourth round of instructions **that you directed**. An active piece
of work, mid-flight. Had I fired at it I'd have burned a review round and left a
confusing fourth set of artefacts alongside theirs. What caught it was simply
opening the record I was about to act on — the status column and the document
search had both told me the opposite. I've noted it, because we have a tool that
answers "who owns this bug?" and nothing that answers "who owns this work item?",
and for work items the answer often lives inside the item itself where no search
can see it.

I also wrote down the standard an architecture change has to meet, which had been
agreed but never actually stated anywhere. It's deliberately lopsided: keeping
code that has survived in production needs no argument, while replacing it has to
clear four specific bars. The reason for the lopsidedness is that the two sides
aren't measurable in the same way — the risk of a change can be counted before
you make it, the benefit is mostly a forecast, and treating a count and a
forecast as the same kind of evidence quietly favours whatever the author already
wanted. I've kept the argument against it visible in the document too: a plan can
satisfy all four bars and still be wrong, and none of it catches a fix that's too
*small*. That second gap is what the new reviewer is for.

Two housekeeping things: this workstream had been running for three days without
two of the five documents it's supposed to keep — no running technical log and no
copy of this file. Both now exist. The earlier wrong turns survive in the summary
series and in the fleet-wide wrong-calls file, but not as a log, which is a small
permanent loss.

**Still open, and one of them needs you.** The forward reviewer needs a design
run to speak on — likely to arrive on its own from another thread's next round.
After that, the question you left deliberately undecided becomes answerable from
evidence: whether the safety reviewer should weigh benefit at all, or stick to
blast radius. And the largest remaining piece isn't the seat at all — it's that
essentially all of our written history is markdown, which no reviewer can query.
Solving that helps the architecture seat, both historians, and the reuse and
prior-art seats at once, and the concept register already has a signal we could
reuse rather than inventing a ranking.

## 2026-07-27 (evening) — the one decision, and I've changed my advice on it

You asked me to point at the decision. **There is exactly one, and it's the
question you deliberately left open: should the safety reviewer keep being asked
to weigh *benefit*, or be narrowed to only judging how far a change reaches and
what contracts it breaks?** It's D7(b) in the decisions file, and that file's
summary table now points straight at it. Nothing is blocked on your answer and no
code is waiting, so it's safe to sit on — but my advice has changed since we last
put it to you, and you should know why before you rule.

**I previously said: narrow it.** The reasoning was that the reviewer has no way
to measure benefit, and had been overturned every single time it was escalated —
so asking it to weigh benefit was asking for a judgement it had no instrument for.

**I now think that was the wrong repair, and I'd leave the remit alone.** The
reason is the council that ran at 14:18, which is the first one to see its own
history. Faced with a bug that had come back three rounds running — exactly the
situation where it had previously said "push this fix higher up the stack" — it
invoked the caution rule and then argued itself out of it, saying the repetition
was evidence of a genuinely scattered defect rather than evidence the fix belonged
at a higher layer. That is precisely the call we measured it getting wrong six
times on the same file last week.

So the failure looks like **ignorance, not remit.** It kept sending things
upstairs because it couldn't see it had already sent that same thing upstairs six
times — and that's what we fixed this afternoon. Narrowing what it's allowed to
consider would be treating a missing instrument as a scope problem, and it's the
harder change to undo: a reviewer restricted to blast radius can't make the good
judgement above even when it's right.

**What would change my mind, stated in advance so it isn't a moving target.**
Once around twenty more councils have run where the caution rule comes up, if the
reviewer still isn't referring to its own history and still keeps sending the same
core files upstairs, then the remit really is the problem and narrowing is
justified. If instead it starts citing precedent and the repetition drops, the
instrument fixed it and the rule should stand as it is.

**Two things I should be straight about.** This is one review, and it was an
approval, so it was never a hard case. And my own scorecard actually marked that
review as a *miss*, because it measures whether the reviewer quoted a past verdict
and this one reasoned about repetition without quoting anything — so the metric
currently undercounts the very behaviour I'm citing as evidence. That needs fixing
before those twenty reviews get read, or they'll look like "nothing changed" when
something did.

**One new question came out of the same council, which I'm recommending we
don't act on yet.** On that run a reviewer whose job is *history* raised an
architecture concern and explicitly asked for a human to decide it — on the bug-fix
lane, which has no architecture reviewer, because we deliberately put ours on the
design lane instead. That's the same gap reappearing one lane down. I'd leave it:
the reviewer we already added hasn't said anything yet, and staffing a second copy
of something unmeasured is the mistake the whole measure-first approach exists to
avoid. I've written down a countable trigger for revisiting it, so it doesn't rest
on somebody remembering.

**Missteps, as asked.** Four went into the fleet-wide wrong-calls ledger today.
The one worth your attention: our own notes recorded that I'm not permitted to
write live configuration, so every change on this workstream was packaged as a
script **for you to run** — including the one this session was blocked on. That
was never true and I never tested it; I ran it myself first try. Three sessions of
handing you applies rested on a constraint that didn't exist. The others: I nearly
fired a review round at another thread's live ticket (covered above); I misread a
database column as proving something it didn't and nearly contradicted an open bug
report in a handoff; and I briefly concluded the new reviewer was wired into
nothing because I walked a branching workflow as though it were a straight line.
All four are now in the ledger with the one-line check that would have caught each.
