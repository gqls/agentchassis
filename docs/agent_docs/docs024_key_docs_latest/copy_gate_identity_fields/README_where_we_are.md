# Where we are — the copy gate and card headings

Plain prose, append-only, newest at the bottom. The owner maintains this too.

---

## 2026-09-04

**The problem, in ordinary terms.** We have a check that catches a particular writing tic — saying
what something *isn't* in order to say what it is. "A fixed margin, not a moving target." The check
reads the copy a page writer produces and rewrites the worst of it before the page ships.

It was skipping the headings on cards. Not by accident of judgement — it never looked at them at
all. The field a card's heading lives in is called `name`, and `name` was on a list of fields the
checker treats as "never prose", along with things like `url` and `price`. So the tic shipped on
the boldest line of the card while the paragraph underneath it, written in the same breath by the
same model, got cleaned up. That is on 37 headings across 15 live sites right now.

**The fix that was written down was not safe, and that turned out to be the interesting part.**
The obvious move is to take `name` off the list. I nearly did it. Two other sessions working
nearby stopped me, each for a different reason.

The first pointed out that the list was doing a second job nobody had noticed. It reads like a
performance filter — don't waste a model call on a URL — but it was also the only thing stopping
the rewriter being handed things it must never change. Take `name` off and you have not just
widened what gets *read*; you have widened what gets *overwritten*.

The second was sharper. It turns out `name` means two completely different things depending on
which part of the system wrote the card. On a blog listing it is the page's identity — the same
string the page is filed under, and the string its own web address is built from. Rewrite that and
the card quietly stops pointing at the right page, while still looking perfectly fine. On a
directory or feature card it is exactly what it appears to be: a heading, meant to be read.

So neither "always skip it" nor "always check it" is right, and the system had no way of telling
the two apart.

**What we did about it.** There is a way to tell them apart, and it costs nothing. An identity
`name` always sits next to a web address; a heading `name` never does. I checked that against
every card on every site — 1,729 of them — and it holds without a single exception in either
direction. Two of us ran that count separately, with different queries, and got the same answer.

That is a better test than the alternatives because it reads something about the *shape* of the
data rather than something about the *habits* of whoever wrote it. The habits change; the shape is
what the producer actually meant.

We also added the missing safety catch underneath. The rewriter already refused to *invent* a name
that wasn't in the original — but it would happily *lose* one. That gap did not matter while the
list was keeping names away from it entirely, and it mattered a great deal the moment we stopped
doing that. It now refuses both ways, and the refusal lives in the part of the system every
rewriter has to go through, rather than in the part you could walk around.

**One decision was yours and you made it.** The check has a rule that a rewrite must leave at least
five words behind, so it can't reduce a sentence to a stub. That was your ruling from yesterday and
it was measured on body sentences. Headings are much shorter. "A fixed margin, not a moving
target" becomes "A fixed margin" — three words, refused. Twenty-five of the thirty-six would have
been refused that way, which would have meant the fix made the problem visible and then declined to
do anything about it. You chose a separate two-word floor for headings, leaving the sentence rule
untouched.

**Something I got wrong, and you should know because you acted on it.** When I asked how to handle
the 37 headings already out there, one of the options I offered was to re-render those 23 pages.
You picked it. Then I checked what a re-render actually does, and it would not have fixed a single
one — it rebuilds the page's HTML from the text we already have, and the bad text is exactly what
it would faithfully rebuild. Only a full content rebuild regenerates those headings. I put the
corrected choice back to you and you went with leaving them to heal naturally as pages get rebuilt.
No sweep. It cost us one round trip and nothing else, but I should have checked before offering it,
and I've written that up where the estate keeps that kind of thing.

**Where it stands.** The code is committed and the review board has it. Nothing changes on any live
site until the next time the system is rebuilt and deployed — this is a code change, so it sits
inert until then. When it does go out, the count of 37 should start falling as pages get rebuilt
for other reasons. I've left the exact queries so anyone can check that themselves rather than
taking my word for it.

**One thing left deliberately alone.** The same "one list doing two jobs" problem exists in another
part of the system, and there the mutating side doesn't rewrite a section — it deletes one. I
haven't touched it, but I've written it down where someone will meet it before they touch that code,
rather than after.
