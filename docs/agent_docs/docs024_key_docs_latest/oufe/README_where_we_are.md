# Where we are — oufe.com and oxenunity.com

Plain-prose log for the owner. Append only, newest at the bottom.

---

## 2026-07-25 — opening entry

You worked the proposition out with Gemini and then couldn't get the conversation
back out of it — it errored three times on the export and the fourth attempt
handed back a document that had quietly dropped the strategy reasoning and printed
its own Python at you. So the first job here was to rescue that conversation. It's
now in `PLAN_2026-07-25_oufe.md` in this folder, and I've split it deliberately:
section 1 is what **you** decided, section 2 is where I think parts of it need
challenging. I kept your reasoning in your own words wherever it was load-bearing,
because the *why* behind a decision is the bit that gets lost first.

Nobody else in the estate has touched this. I checked the whole repo, the memory
files and about two hundred past sessions for any mention of oufe or Oxen Unity —
nothing. So we're starting clean.

**Three things I found that change the picture from what Gemini was told.** The
capabilities document you pasted into that chat is a bit out of date, and Gemini
built parts of its architecture on the stale bits.

First, the citation-verification system — the one that goes and re-reads a source
to check a quote is really there — is described as "designed, not built". It was
actually built and switched on five days ago. What's true is that it's never
successfully completed a full run since the bug blocking it was fixed. That makes
oufe its first real test, which is fitting, because this is the site where that
capability matters most.

Second, the chart library Gemini specified for the waterfall diagrams doesn't
exist. That line in the capabilities doc was corrected the day after it was
written. It doesn't matter much — a chart driven by a slider should be drawn in
the browser anyway — but it would have been an odd surprise mid-build.

Third, and this one is important: our automatic "did the AI invent this number?"
checker is close to useless on financial prose. It only looks at numbers sitting
near business words like *clients*, *awards*, *customers*. It has no idea what a
creditor or a recovery rate is, and it deliberately skips currency amounts
entirely. So on a site made almost entirely of pound figures, that check will
report "clean" and mean almost nothing. We're not relying on it. Instead the
writer only gets a whitelist of numbers it's allowed to use, everything comes with
a source, and I look at it before it ships.

**Where I disagree with you, and it's worth saying plainly.** You said start with
direction three, the automatic radar that scans for companies in trouble, because
it's the lowest risk. I think it's the highest risk thing we could do first — and
not because of the idea, which is good, but because of what we'd have to build it
on. We have no market data anywhere in the platform: no bond prices, no yields, no
maturity schedules. UK court listings have no feed you can subscribe to. And a
distress signal is a statement about a named real company — which is precisely the
shape of the worst mistake this platform has made (the vet site, where we
published invented prices against three thousand real practices).

The genuinely low-risk start is the thing you separately called the primary
magnet: the Thames Water dossier, done properly, with one excellent interactive
tool alongside it. The radar comes back later in narrow slices we can actually
source — maturity walls read out of filed accounts, for instance, which is free
and citable.

**A warning about the numbers in that Gemini transcript.** The sixteen billion of
Class A, the three billion of new money, the one billion of Class B — those are
recollections, not facts, and the case has moved since whenever Gemini learned
them. None of them go anywhere near the site until we've fetched them from a court
judgment or an Ofwat document and stored the quote. This isn't fussiness: a
fortnight ago on another site, invented figures got written *into that site's own
setup*, and a routine rebuild then wrote them back over the correct numbers — with
both our safety systems switched on and working. A number written into a site's
configuration is treated as a given, and a given beats every rule we have. So the
safest place for an unverified figure is nowhere at all.

**What I'm building now.** The docs you're reading, then the oxenunity.com page,
then the skeleton of oufe.com.

oxenunity.com is one hand-written page: the Oxen Unity wordmark, one neutral line
about what it is, and a link through to oufe.com. You chose to make no claims
about the entity at all rather than explain that there isn't a company yet, and I
think that's right — a page that claims nothing can't say anything untrue. It also
means no cookies, no forms, no tracking, so it doesn't even need a privacy page to
be honest.

oufe.com goes through the normal build pipeline but on a short leash: a small,
fixed page list (home, about, cases, the Thames Water hub, contact, plus legal),
no news feed, and a fact register attached before anything is written so the
writer starts with an empty whitelist and can't reach for a plausible number.

**Two things I need from you**, neither of which blocks me building:

The Cloudflare wiring for both domains — the zone, the nameservers at the
registrar, and crucially binding the worker route. That last step is what left
fundamentallyai.com dark after a perfectly successful build, so it's worth doing
consciously rather than discovering later.

And the disclaimer wording. This site needs to be clearly educational analysis
rather than investment advice, and the disclaimer needs to sit *with* the content,
not just in the footer — for paid research the real exposure isn't a regulator,
it's someone saying they relied on us and lost money. I'll draft it for you to
approve rather than invent your legal position for you.

**One thought to leave you with.** Everything in this field is stated with total
confidence, and a good deal of it is wrong. We have machinery that goes and checks
its own claims against the source document, and a track record of catching and
publishing our own errors. On a site about who is telling the truth about a
balance sheet, "every figure here links to the document it came from" might be the
actual product, not the hygiene. Worth thinking about as positioning.
