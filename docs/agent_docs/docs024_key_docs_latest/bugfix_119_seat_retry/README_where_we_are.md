# Where we are — the council seat that loses a whole round

*(Plain-prose log, append-only, newest at the bottom.)*

---

**2026-08-01, morning.**

I picked up bug 119. The short version of it: when we send a change to the review
council, thirteen or so reviewers each write their opinion as a small structured
document. If **one** of them writes that document badly enough that our code cannot read
it, we throw away the *entire* round — every other reviewer's completed work included —
and the submitter has to start again. It costs their credits and about half an hour, and
the verdict that comes back says nothing at all about the change they proposed.

Before doing anything I checked whether this was still real, because the bug was filed
five days ago. It is real and it is worse than filed. Across every council round we have
ever run — 424 of them — **23 were decided by "we couldn't read one reviewer" rather than
by anyone's judgement**, and 15 of those were in the last week alone. It has happened to
seven different reviewers, so it isn't one badly-written reviewer prompt; it's something
about the whole arrangement. The worst single case was on 31 July, where one submission
burned **three rounds in ten minutes** — same reviewer failing the same way each time,
because nothing anywhere tries again.

**Then the measurement changed my mind about what to build.** The bug says the cause is a
reviewer writing a document that's complete but has a bracket in the wrong place. I went
looking for live examples so I could fix exactly that — and there are none. Not "a few":
zero. Of the 39 cases I could find, the evidence for 36 had already been deleted by
routine cleanup, and all 3 survivors were a different failure: the reviewer ran out of
room mid-thought and produced nothing at all.

That mattered, because I was about ten minutes from building something that would have
fired zero times and reporting it as a fix for fifteen lost rounds a week. What is
actually and permanently true is the *other* half of the bug: **nothing ever asks the
reviewer to try again.** So that's what I built against.

**While reading the code I found something I wasn't looking for, and it's the better
half of the fix.** Each step in our system can declare what it needs back — "I need
JSON", say. It turns out the code reads a setting called `output_type`, and almost
nothing in the fleet writes that. What the fleet actually writes is `output_format` — a
near-identical name — and **nothing anywhere reads it**. The numbers: 6 steps use the
name the code understands; 100 use the name it ignores. Ninety of those hundred are
asking for JSON, across 32 different agents, and that includes every single council
reviewer.

The practical effect is small to describe and large in aggregate: those 90 steps were
never given the instruction sheet for producing JSON. And the very first line of that
instruction sheet is *"Ensure valid JSON syntax (proper quotes, commas, brackets)"* —
which is, word for word, the thing the bug says the reviewers keep getting wrong. I want
to be careful here, because I can't prove that instruction would have prevented the
failure; nobody can prove that. What I can say plainly is that we were asking these
agents for something and never telling them how, and now we do.

**So the fix is two things.** Make the setting the fleet actually writes mean something
(prevention). And when a step that asked for JSON gets back something unreadable, ask it
**once** more — not with the identical question, which we know just reproduces the same
failure, but with a short corrective note that depends on what went wrong: "you ran out
of room, same judgement but shorter" or "that wasn't valid JSON, same answer as one valid
document". If the second attempt fails too, everything behaves exactly as it does today.

I checked what that costs before building it rather than guessing: across 785 of these
steps in the recent record, this would have triggered **twice**. It only ever runs on a
path that was already producing something unusable.

One deliberate non-decision worth flagging: my first instinct for the "ran out of room"
case was to give the reviewer a bigger budget on the retry. I didn't, because there's a
note in our own code from an earlier round of this exact problem saying, in effect, don't
— whatever the limit is, the reviewer with the most to say will reach it, so raising it
buys a little time and no cure. Asking for the same judgement more briefly is the honest
fix.

It's committed and it's had tests written against it that I verified by deliberately
breaking the fix to check the tests actually notice. It's now in front of the review
council itself, which is a pleasing sort of circularity. **It is not live yet** — Go
changes only take effect when a new image is built and rolled — so the bug stays open
until I can prove it's running in the actual pod rather than just in git.
