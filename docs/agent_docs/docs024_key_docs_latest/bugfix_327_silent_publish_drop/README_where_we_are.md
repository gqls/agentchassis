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
