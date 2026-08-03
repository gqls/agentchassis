# Where we are — the link-checking coverage bug (116)

Plain-prose log, append-only, newest at the bottom.

---

## 2026-08-03, evening

I picked up bug 116 because it was the next open bug that no other session was
holding. There are a lot of sessions running on this repo tonight — I checked all
of them, and eight of the fifty-odd open bugs had nobody near them. Of those eight,
116 was the one whose fix would be a change to the framework rather than a patch to
one site, which is what you've asked for.

The bug says this: we have three checks that look for broken links on our sites —
links pointing at pages that don't exist, buttons that go nowhere, "Get in touch"
buttons that land somewhere unexpected. The checks are well written. The complaint
was that nothing ever ran them, so every site we've declared healthy has never
actually had its links looked at. The person who filed it in July put it well: a
site with no reported problems looks exactly the same whether it's clean or whether
nobody ever checked.

**The first thing I found is that the headline is no longer true.** The checks did
run — today, in fact, about an hour before I started, on three sites, and they
found real problems. What misled the original filing is a naming quirk: the checks
are named in the plural and the problems they file are recorded in the singular, and
one of the three files its findings under a completely different name. So if you
search for what the checks are called, you get nothing back, and nothing looks
exactly like never. Another session fell into the identical hole on this same bug
earlier today, which is a fair sign it's a trap rather than carelessness.

**The second thing I found is more useful, and it's why I've deliberately not
fixed this.** The proposed fix — and it was your own suggestion in July, that the
checkers should run after every build or change — would work mechanically. But the
findings these checks produce go into a queue at a stage called "detected", and the
only thing in the whole platform that moves a finding from "detected" to "somebody
will act on this" lives inside the improvement loop, which you stopped on purpose
on the 29th of July. Right now there are 204 findings parked across ten sites, and
two that have made it through.

So if I'd built the fix, I'd have built something that finds more problems and puts
them on a pile nobody is emptying. That isn't me being cautious — it's written down
in three separate places as the thing not to do, including in the very piece of
code the bug file suggested I hang the fix off. Someone faced this exact decision at
this exact spot before and wrote down why they declined: filing a work item would
"promise a repair that nothing performs".

The other three options in the bug file are all closed too: two of them amount to
restarting the loop you deliberately stopped, and the third is something the team
working on a neighbouring bug has explicitly asked people not to do yet, because it
would muddle a measurement they're in the middle of.

**So the honest answer is that this bug is waiting on a decision, not on code.** The
decision is the one you're already being asked elsewhere: what happens to those 204
parked findings, and whether the improvement loop comes back. Until that's settled,
widening the net makes things worse, not better. I've written all of this into the
bug file with the evidence, so the next person who picks it up doesn't spend an hour
rediscovering that the title is out of date and then build the thing we shouldn't
build.

I also nearly filed a second bug tonight and I'm glad I didn't. I'd noticed that a
scheduled job which is switched off nevertheless shows a "last completed" timestamp
from minutes ago, and I built a neat theory about how that could cause the system to
run the same sweep twice over. Then I read the function that writes that timestamp
and the theory evaporated — the scheduler already handles it, and the misleading
timestamp is only misleading to a human reading the table, not to the machine. It's
recorded as a trap for the next person rather than as a defect.

Two things I got wrong along the way are written up properly: I measured "how many
sites have been link-checked" by counting sites that had findings, which is the
flattering version of the question and can't distinguish a clean site from an
unexamined one — the very mistake the bug file warns about. There is a proper record
of when each site was last audited and I've documented where it lives.

Nothing has been committed to the running system. No code changed.
