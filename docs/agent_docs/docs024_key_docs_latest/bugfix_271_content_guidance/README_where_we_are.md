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
