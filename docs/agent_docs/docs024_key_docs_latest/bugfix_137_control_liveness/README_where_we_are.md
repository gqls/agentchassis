# Where we are — the two judges of "is this control alive" (bug 137)

*(Append below, newest at the bottom. Plain prose.)*

---

**2026-07-31, evening.**

Bug 137 was an odd one to pick up because it doesn't claim anything is broken.
It says: we have two bits of code that both answer the question "is this button
actually a button, or is it a dead thing pretending to be one" — they sit in the
same function, they run in the same pass, and on one real element on vonc.com
they give opposite answers. One says fine, one says broken. Nobody had decided
which was right, and the person who filed it said so honestly, including the
uncomfortable observation that the reading which made their own red result go
away was suspiciously convenient. It was filed because a council reviewer asked
for it to be, rather than deferred to "the register, later".

The first thing I did was check it was still true. It is — that page still
serves exactly one dead-looking link, the hidden template row the page's own
JavaScript clones to build the archive list.

**What I found when I read the code is that the disagreement was a symptom.**
Both mechanisms share an escape hatch: a section can be marked "this gets filled
in by the browser later", and anything in it is excused, because its links are
meant to be placeholders. Perfectly sensible. The problem is *how* the excuse was
tested — with a plain "does this text appear anywhere in what I was given". So
the excuse's reach depends entirely on how much HTML the caller handed over. Hand
it one section, and it means "is this section a placeholder", which is right.
Hand it a whole page, and it means "does this page have a placeholder section
anywhere", which excuses **everything else on the page too**.

That line appears in eight places. Nobody wrote down the second behaviour and no
test caught it, because at each individual spot it reads as obviously correct.
And somebody had already half-noticed: one file works around it by deliberately
feeding the function one section at a time, with a comment explaining that this
stops a placeholder section's neighbours "riding on its exemption". They got the
right answer and fixed it in their own corner, which left everyone else to
rediscover it.

**So the fix is to make the excuse element-shaped instead of page-shaped.** A
control is excused only if it is *itself* inside a placeholder section — not
because something else on the page is. That's one small idea, and it resolves the
original disagreement for free, because once both mechanisms ask the
element-shaped question they agree about the element they were arguing over.

I measured it on the real pages rather than reasoning about it. On vonc.com's
home page the old test excuses 100% of the page; the new one excuses 12.6%. What
comes into view as a result is two dead buttons ("Get Started" and "Learn More")
and two broken links — and those two broken links are "Enter the Gauntlet" and
"Find Your Archetype", **which are the exact buttons the dead-control checker
names in its own header as the reason it was written**. They'd been sitting
behind the page-wide excuse. On the provocations page, the one the bug is about,
the new rule excuses the template row correctly and finds nothing else, and the
page comes out byte-for-byte unchanged.

**One judgement call worth flagging to you.** For the stricter of the two
mechanisms — the one where a person writes down "this element must not have a
link" — I did *not* make it silently say "fine" for elements inside a
placeholder. It says "skipped, and here's why", which counts as neither pass nor
fail. That matters because a "pass" would mean the system vouching for markup
nobody actually checked, and that vacuous-pass problem is the whole reason that
file exists. Elements outside a placeholder are still judged exactly as strictly
as before.

I also had to change a test that asserted the *old* contradictory behaviour — it
had the disagreement baked in as an expectation. I've flagged that to the council
explicitly rather than let a reviewer stumble on it, because if anyone thinks the
reconciliation went the wrong way, that test is where the decision lives.

Three things I got wrong along the way, all caught before they reached anything
durable, all written up in the notes: I nearly reported a masking effect against
the wrong mechanism; I wrote a database query joining pages by name when names
aren't unique across sites, and got 75 rows of every site's home page; and my
first live measurement used a hand-made list of two page URLs, which made the
tool "repair" a link that is perfectly fine in production. The last one is the
same trap I've seen logged repeatedly — the test fixture I invented decided the
answer.

It's committed and submitted to the council. The code is inert until the next
chassis build, so the bug stays open until then.
