# Where we are — fleet copy quality

The owner's running plain-prose log. Append-only, newest at the bottom.

---

## 2026-08-06 — I checked the copy three ways, found nothing, and that turned out to be the answer

You said copywriting is an art and subtler than I was treating it, and asked how we'd
handle it properly. Here's what I did and why, and what it turned up.

Before proposing anything I wanted to know one thing: **is the problem something a
machine can see?** That matters more than it sounds. If a fault shows up in a
measurement, we can detect it, put a gate on it and prove a fix worked. If it doesn't,
then no rule and no automatic check will ever hold it, and every hour spent writing more
rules is wasted. So I took the three most likely mechanical explanations and tried hard
to prove each one true, against real copy pulled off six live sites.

I set each check up so it could come out against me. That's the part that matters — a
check that can only give one answer tells you nothing.

**First: are all the sentences the same length?** That was my best guess, because our
rules tell writers to keep one idea per sentence. My first look seemed to confirm it
beautifully: every sentence on the page between six and fourteen words. Then I widened
from two pages to nearly nine hundred sentences and it evaporated. The variety is
normal, human, fine. The monotony was in my sample, not in the copy — and it happened to
be exactly the answer I was hoping for, which is the most dangerous kind of result.

**Second: does the copy keep announcing its limits?** That was your specific complaint.
Searched as phrases, it's essentially absent — well under one instance per two thousand
words on every site. The copy is also already talking to the reader rather than about
itself; on one site "you" outnumbers "we" eighteen to one.

**Third: do the sections repeat each other?** There's a real structural reason they
might — each section is written by a separate pass that can't see any of the others. But
measured properly, they don't. I'd predicted they would, and I was wrong.

So: three checks, three misses. Everything countable about our copy looks healthy.

**Then I stopped counting and just read it, and there it was on the first two pages I
opened.** One site tells you, in three paragraphs running, that it won't say whether
your idea will succeed, that it doesn't give verdicts, and that it can still be wrong
and it's your call anyway. Another leads with "No preferred platforms. No black boxes."
and then, a paragraph later, "We don't have a large org chart or a department for every
service. What we have is…".

That last sentence is the whole thing in miniature. **It is precisely the construction
your own style prompt banned two revisions ago, sitting on a live site right now, with
the rule still loaded in the writer's instructions.**

Which is why the three misses are the answer rather than a dead end. The fault is real —
you spotted it without any of this — but it doesn't live anywhere a rule or a checker can
reach. Every search I wrote missed it, because it's a *move*, not a phrase.

Two things seem to be going on, and the second one is uncomfortable. A rule can only
name a shape, and what's going wrong is a habit: ban "it isn't X, it's Y" and it comes
back as "Nothing here is X", ban that and it comes back as "We don't have X, what we
have is Y". Same instinct, three costumes, two patches, still shipping. And underneath
that — some rules are easy to obey and easy to check, like not using dashes, while the
ones we actually care about, like knowing which ideas deserve explaining, are neither.
Under pressure the easy ones win. So every round of tuning adds more easy rules and
buries the ones that matter. **The work we've been doing on this has been making it
slightly worse.**

What I'd suggest, and I'd rather talk it through than just do it: stop adding rules and
add a reader instead — a second pass that reads the finished page the way a person would
and repairs what's weak, briefed with what we're trying to achieve rather than a
checklist. That's not a new idea here; it's exactly how your own style prompt got built
in the first place, by writing it both ways and picking blind, and it's how the review
council works. A reader can catch a move. A rule can only catch a shape.

I've kept the measuring script. It's found nothing three times, which is now its job: it
can't tell us the copy is good, but it will tell us if we make it mechanically worse,
and I'd rather have that in place before anyone starts editing prompts.

Three things I can't work out from the code, which is why I asked rather than guessed.
Whose voice is this meant to be — yours, which is what your style prompt was built from,
or a service talking to a stranger with a job to do? What should decide whether we make
an offer strongly or just hint at it? And is there a page anywhere on the estate you'd
point at and say that one's close — because a real example beats any amount of
describing, and that's how we got the style prompt right last time.
