# Where we are — bug 168, the asset path helper

Plain prose, append-only, newest at the bottom.

---

## 2026-08-02, morning

I picked up bug 168 off the open pile. It had been filed on 31 July by the lane that fixed
128, and nobody had touched it since. I checked that twice — once with the ownership script,
and once by reading the live transcripts of the 27 other Claude sessions currently working
this repo, because the script only knows about work that has already been committed and half
these sessions are mid-fix. None of them was near this one.

The bug is about a small function with a big job. When the platform generates an image — a
hero, a logo, an icon, a social card — it commits the file into the site's git repo and then
some page has to point at it. The function `DeployedWebPath` is what everything asks for that
path. Six different places call it.

The complaint in the bug file is that this function gets one case wrong: the social card. It
answers `og_card.png`, with an underscore, when the actual file on disk is `og-card.png`,
with a hyphen. That is true. But when I went to fix it I found the filed explanation of *why*
was wrong, and wrong in a way that mattered — one of the four suggested fixes would have made
things worse rather than better. That is the first thing worth saying.

**What was actually going on.** There are two different bits of code that publish these image
files, and they name files differently. The main one derives the name from the image's
purpose. The other one — the bit that makes the favicon and the social card — just writes
`favicon.png` and `og-card.png` as fixed names, because those are what browsers and Facebook
expect. `DeployedWebPath` only ever knew about the first one. And here is the thing: from
what a caller passes in, there is no way to tell which of the two published any given image.
So every single place that used this function had to separately remember "…except for the
favicon and the social card". Some did. One nearly didn't, and if it had shipped it would
have reported a broken social card and a broken favicon on every site we run.

The suggested fix I nearly took would have swapped underscores for hyphens everywhere. That
sounds harmless. It isn't: the main publisher *also* uses underscores in that situation, so
the two agree today, and the "fix" would have pulled them apart. I only caught it by going
and reading the publisher instead of trusting the comment on the function, which claims to
mirror it. A comment claiming two things match is not the same as two things matching.

**What I did instead.** I made it one function. Where before there were two copies of the
naming rule — one in the publisher, one in the lookup — kept in step by a comment saying they
matched, now both call the same code. They can't drift, because there's nothing to drift
from. And the special case for the favicon and the social card lives inside that one
function, so none of the six callers has to remember it any more. The one that had been
carefully remembering it can now stop.

**On being wrong, deliberately checked.** Before I asserted any of this I put it through the
diagnosis loop, which is the thing that reads the real code and the real database and tells
you whether your theory holds. It came back **REFUTED**. That is worth explaining rather than
hiding, because it was useful in both directions. It was right about the thing that matters
for how alarmed anyone should be: the code that writes the social card tag into the page
doesn't use this function at all, it writes the correct name directly — so nothing is broken
on any live site today, and this is a trap waiting to be sprung rather than a fire. That
matches what the bug file itself said, and it's why I've described the change as removing a
hazard rather than fixing an outage.

It was also wrong about something, and I've written that down rather than just accepting the
verdict. It claimed the function had only one caller. It has six. Its conclusion rested on
having missed five of them, including the one place where the problem genuinely bites. So I
took the useful half and recorded the rest as an error of the tool's, with the evidence.

**Where it stands.** The code is written, builds clean, and all the tests pass. I don't
trust "the tests pass" on its own, so I broke the code three separate ways on purpose to
check that each new test actually notices — it did, all three times, and then I put it back.
It's gone to the review council, which takes about half an hour to come back. It's committed,
because on this repo holding code back isn't actually available: everyone shares one branch,
and the next person who builds ships whatever is there.

The one thing I want to be plain about: this is **not live yet**. The change is Go code, and
Go code does nothing until someone builds a new image and rolls it out. Until that happens
the old behaviour is still what's running. So the bug ticket stays open — the house rule is
that a bug is only closed when the fix is actually running in production, not when it's
written — and I'd rather leave it honestly open than tidy it away into the closed pile.
