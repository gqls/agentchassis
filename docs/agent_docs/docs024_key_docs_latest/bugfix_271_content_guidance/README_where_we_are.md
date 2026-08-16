# Where we are — the rewrite briefs nothing reads (bug 271)

Plain-prose running log, append-only, newest at the bottom.

---

**2026-08-15, evening.** Picked up bug 271. The owner suggested 279, but 279 is
already someone else's and is essentially finished — its fix is committed and
just waiting for the next chassis roll, so there was nothing to take. I checked
which bugs the other live sessions are actually working on right now (about
forty-five of them are open on this repo at the moment) and 271 was the newest
one with nobody on it.

**What the bug says.** When you tell the platform to rewrite a page — "state
this, remove that, add the other" — you put those instructions in a field on the
work item called `content_guidance`. Nothing reads it. The rewrite happens
anyway, steered only by the site's general facts and whatever is already on the
page, and it reports success. So the instructions are honoured only when they
happen to coincide with what the writer would have done regardless, which is why
this survived so long: it looks intermittent rather than broken.

**What I found that the bug file did not.** There is already a working channel
for exactly this, under a different name. A field called `suggestion` on the same
work item travels all the way through to the writer's prompt, where it appears
under a heading that says "Rewrite Guidance (IMPORTANT: incorporate this into the
content)". I traced that end to end and confirmed each hop against the live
system. So the real shape of the problem is not "a field with no reader" — it is
**two names for one thing, and only one of them is plumbed in.** The unplumbed
name is the one the platform's own gap-planner uses.

**How much is affected.** Roughly ninety live work items are carrying written
instructions in the dead field with nothing in the live one. Every other kind of
work item on the platform already uses the live name, so this is a minority
spelling that got copied between four bits of code and never wired up.

**Why that changes the fix.** The bug file's suggested fix was to plumb the dead
field into a different part of the pipeline. That would work, but it adds a
second channel rather than removing the confusion, and it means editing a file
another session is currently working in. The cheaper and more general fix is to
make the platform accept both names at the single point every work item passes
through on its way to a handler — which also rescues the ninety items already
sitting there. I have asked a second model to design that properly and to argue
against me where I have assumed something; it is working now.

**One thing I want to flag before it lands.** The moment this ships, about ninety
work items that currently do nothing with their instructions will start acting on
them. That is the point of the fix, but it is a real behaviour change to items
already in the queue, so I have asked for that to be stated plainly in the plan
and I will put it to the council review before committing.

---

**2026-08-15, later that evening.** The fix is written, reviewed and committed. It
is not live yet — it is Go, and Go does nothing until a new chassis image is
built and rolled, so the last step belongs to whoever sees the next roll. There
is a short checklist at the end of the bug file for exactly that.

**What we did, in plain terms.** Rather than build a new pipe for the ignored
instructions, we connected them to the pipe that already works. There is now one
place — the point every piece of queued work passes through on its way to being
done — where a brief written under the old name is quietly read as if it had been
written under the working name. Nothing is rewritten in the database; this
happens in memory, on the way past. The four bits of code that were writing the
dead name now write the live one, and there is a test that fails the build if
anyone reintroduces the dead name.

**The second model earned its keep.** I asked it to design the fix and to argue
with me rather than agree. It corrected me on three points, and one mattered: I
had claimed only one thing in the whole system reads the working field, and there
are three. My query could only see one *kind* of reader. Since the entire safety
argument is "everything that reads this field just puts text into a prompt —
nothing makes a decision from it", an enumeration that could only see part of the
picture was not good enough to support it. I re-checked it myself before
believing either of us. That is now written up as a mistake to learn from.

**The review board approved it first time, and the useful part was the
complaints.** One reviewer said something I think was exactly right: I had
written the safety rule as a comment in the code — "this is the one permitted
exception" — and a comment is not a rule, it is a suggestion that the next person
may read as permission. So I replaced it with a test: the queue loader must hand
on each item's data exactly as stored, plus the one agreed addition, and nothing
else. I then deliberately broke that rule to confirm the test catches it. Another
reviewer made the fair point that we have fixed the one known wrong spelling, not
the underlying reason a wrong spelling can go unnoticed at all — that is now
written down as unfinished business rather than quietly filed as solved.

**One thing worth your attention — and a correction to what I told you earlier.**
Higher up this page I said that when this ships, ninety work items "will start
acting on their instructions", and that this was a real change to things already
in the queue. **That was wrong, and I only found out because I went and counted.**
Around ninety pieces of queued work are
carrying instructions in the dead field. None of them is currently active — they
are all finished, cancelled, or parked awaiting a human — so nothing changes
behaviour the moment this ships. But about twenty-five are parked in
"needs review" or "failed", and if someone releases one of those back into the
queue after this is live, it will act on its instructions for the first time.
That is what you would want to happen, but it is a real change and I would rather
flag it now than have it surprise someone later.

---

**2026-08-16, afternoon.** It is live and it works. Your release put the fix on
the fleet (chassis v1.0.1304), and I checked it on the running system rather than
taking the release as proof.

**How I proved it.** I made up a phrase that appears nowhere in any of our sites —
"heliotrope kettledrum" — and filed a rewrite instruction containing it, using
only the broken field name, on an internal spare site nobody visits. The phrase
turned up in both of the writer's prompts and in the page content it produced. I
then filed a second item with no instructions at all, and it correctly produced no
guidance section — which is the bit that matters, because otherwise I would only
have shown that the heading is always there.

**Something worth knowing about checking deploys.** My first attempt to work out
which version was running gave me the wrong answer, confidently. The startup line
that states the version had already scrolled out of the logs, so I searched the
logs for anything that looked like a version stamp — and found one. It belonged to
a different system whose report our software happens to print. That made it look
as though your release did *not* contain the fix, and I was a sentence away from
telling you so. What settled it properly was asking the running program directly
whether it contained the new function, with two deliberate control questions whose
answers I already knew. I have written that up, because it is the kind of mistake
that would otherwise be repeated.

**The twenty-five items.** You asked me to put all of them back in the queue and I
have. Ten had previously reported a failure at the publishing step — those had
probably published fine and merely recorded failure, so the real gain is that they
now get redone *with* their instructions. Fifteen were sitting waiting for a human
to look at them, some since April; putting those back skips that wait, which is
what you decided when I flagged it. Each item now carries a note saying what its
previous state was and why it was re-queued, so nothing is lost, and I saved the
before-picture to a file in case you want any of them put back as they were.

**One courtesy I owe someone.** One of the twenty-five was another workstream's own
test item, which they had left in place deliberately. It has now been re-run. No
harm done, but they should hear it from us rather than notice it.

**What is left.** We fixed the one wrong field name; we have not fixed the fact
that a wrong field name fails silently in the first place. If some future code
invents a third name for the same thing, this bug comes straight back and nothing
would warn us. That needs a proper list of which fields mean something, which does
not exist yet and is already noted against another open bug. Nothing is broken
today because of it — I just don't want it recorded as "solved" when it is
"solved for the case we found".
