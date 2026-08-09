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

---

## 2026-08-09 — you've made H the default, and here's exactly what that turns on

You said: go with H as the default for now. Written up in full as
`SUMMARY_2026-08-09_h_becomes_the_default.md`; this is the short version.

What it decides: H stops being the finance pool's voice and becomes the house voice in the
writer's base prompt — every vertical, every future site. That finishes the "wide rather
than contained" choice you made on the 5th, which had been sitting undone.

The one place the old default and H actually contradict each other is the opening sentence.
The current default says "start with the fact". H says don't open cold with a bare
assertion, and vary how sections open. H wins that. Both of them already ban opening on a
negative twist, so that half changes nothing.

Worth knowing how little this disturbs what's live: the house voice only applies where a
site hasn't got a voice of its own, and twenty of your twenty-one deployed sites have one.
Only **cookly.uk** has none at all. But all seventeen sites waiting in the pool have no
voice spec, and neither will anything we build from now on — so this is a decision about
future copy almost entirely, which is why it's cheap to make and worth making now.

Two things I've deliberately not decided, because they're real choices rather than details.
How it physically ships: the "base prompt" turns out to be seven separate prompts that have
already drifted apart from each other, so it's either seven edits that will drift again, or
one shared copy they all read. That goes to the review council with both options written
out. And what happens to the worked examples — which on this week's evidence is probably the
part that decides whether the change does anything at all, because we proved that a writer
follows the examples and treats the rule as commentary.

One thing I want on the record rather than buried, since it points the other way: when we
tested the homepage opening two ways, the plain default produced the *softer* claim and H
did not. The over-claim you objected to survived H and was removed by the default. That's
one comparison on one page and it was about claim strength, not openings — but it's the
reason I've written this up as "for now" and given it two specific triggers for revisiting,
rather than filing it as settled.

---

**2026-08-09, later — we compared our voice against a real published standard, and it was
more useful than I expected.**

You pointed me at a Simplified Technical English skill — ASD-STE100, the standard aerospace
and defence use for maintenance manuals, where the whole design goal is that a tired
engineer reading in a second language cannot possibly misread a procedure. The question was
what our content writer would look like under those rules instead of ours.

I extracted it into the same shape as our house voice block, so the two are genuinely
comparable rather than described at each other, then measured our finished loancalculator
site against it. All twenty-six pages, five hundred and forty-two sentences.

**Fifty-six percent of the site breaks at least one of its rules.** Mostly three things:
sentences over its word limit, contractions, and phrasal verbs — "pay off", "take out a
loan", "hand the car back". Those aren't accidents in our copy. They're the register we
chose on purpose, and H states the contraction rule outright. The standard also mandates
American spelling, which is a non-starter for a UK lending site.

So the answer to "should we adopt it" is no, and I'd say that fairly confidently. Its
uniformity is exactly right for a manual and exactly wrong for copy whose job is to keep a
nervous borrower reading.

But three things came out of it that I think are worth having.

The first is that it caught something in our own spec. H says sentences should average
eighteen to twenty-five words. An average isn't a check — you can hit it perfectly while
writing a fifty-word sentence, and we do: our mean is exactly eighteen, sitting on the floor
of our own band, while our longest live sentence is forty-nine words. STE uses a hard
per-sentence ceiling instead. That's enforceable and ours isn't.

The second is that it makes the writer decide, before it writes anything, whether a section
is instructions or explanation, and applies different limits to each. Our spec has one
register for the whole site. Our site plainly has both — the tool pages instruct and the
guides explain — and when I rewrote samples of each, the same rules helped one and flattened
the other.

The third is the one I didn't expect. On warnings, STE is simply better than us. It requires
the command first and the risk second. Our live overpayment warning opens "there's a catch
worth checking before you commit to it" and makes the reader dig the instruction out of it;
the STE form opens "examine your loan agreement before you overpay", then gives the ten
percent, then the charge. On the handful of sentences where being misread actually costs the
reader money, theirs is the better shape. H has a rule about callout boxes but says nothing
about ordering. I'd take that one narrowly and leave the rest.

There's also a structural point for the "how does it ship" question you've got open. The
skill splits into three files — rules, dictionary, worked examples — pulled in separately.
That's the shared-carrier option already worked out. Though note it loads its examples last
and only "when unsure", which given what we proved this week about examples beating rules is
the wrong way round. If we copy the structure we should invert that.

Two corrections on myself, since they're the useful part. My first measurement said sixty
percent and I nearly wrote it down; it was counting "Financial Conduct Authority" as a style
violation because "conduct" is on the banned word list, and scoring "promise" and "otherwise"
as British spellings. Reading every hit rather than the total caught it. And I had a
confident argument ready that the standard's ban on "should" would be dangerous for
regulated finance copy — until I checked, and our site uses "should" zero times in five
hundred and forty-two sentences. Good argument, false premise.

Nothing was changed. No agent edited, no copy touched, no test fired at the live site. The
candidate block, the comparison and the measuring script are all in this directory.
