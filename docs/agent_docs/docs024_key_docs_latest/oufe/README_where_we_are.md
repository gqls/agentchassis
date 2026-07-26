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

**Update, same day — good news on one domain, a job for you on the other.**

The oxenunity.com page is written and deployed. It's live in storage — I've
confirmed the file is there — and the deploy pipeline ran clean in twenty-one
seconds.

But I was wrong about the Cloudflare job being one task covering both domains, and
the difference matters. I went and looked rather than assuming, and the two
domains are in opposite states.

**oufe.com is already fully wired up.** Its nameservers are ours, and when I asked
it for a page it answered with our own system's error message saying it looked in
the right place and found nothing there — which is exactly right, because we
haven't built anything yet. So the thing I flagged as a risk, the step that left
fundamentallyai.com invisible after a perfectly good build, simply cannot happen
here. The moment content exists, oufe.com serves it. Nothing for you to do.

**oxenunity.com is not with us at all.** It's still sitting at the registrar and
currently redirects to a parking page. So the page I've built for it is finished
and sitting in storage, and it will stay unreachable at its own address until the
domain is moved onto Cloudflare and pointed at our system. That's now the only
infrastructure job in this whole workstream, and it's a small one.

**One thought to leave you with.** Everything in this field is stated with total
confidence, and a good deal of it is wrong. We have machinery that goes and checks
its own claims against the source document, and a track record of catching and
publishing our own errors. On a site about who is telling the truth about a
balance sheet, "every figure here links to the document it came from" might be the
actual product, not the hygiene. Worth thinking about as positioning.

---

## 2026-07-26 — oxenunity is live, and you were right to kill my slogan

**oxenunity.com is up.** Your nameserver change went through and the page is
serving properly. If you look at it and still see the old parking redirect, that's
your own machine holding a cached answer for a few hours — I checked it against
the real thing and it's live.

**On the line I suggested yesterday — you were right and I was wrong.** I proposed
"every figure here links to the document it came from" as the positioning. That's
a promise about our reliability, and it doesn't survive being looked at properly.
A citation tells you where something came from. It doesn't tell you we read the
source correctly, or picked the right bit of it, or that the source itself is
right — and it does nothing at all about the model writing a convincing sentence
around a real link, which is exactly the failure we've actually had. It's a claim
we can't keep.

What replaces it is the opposite, and I think it's stronger: **we make mistakes,
the tools can be wrong, the sources can be wrong, and our reading of them can be
wrong — and we cite everything anyway, so you can check us.** The citation stops
being a warrant that we're right and becomes the instrument you catch us with.
That's a claim we can keep every day, and in a field where everyone sounds certain,
being the one publication that says plainly where it can fail is a genuine
position rather than an apology.

**The tool now makes you agree before you can use it.** You asked whether the
disclaimer could be a condition of use — it can, and it now is. You get a panel
saying the model can give you a wrong answer, that it's a simplification, and that
every result should be treated as a possibly inaccurate worked example. The tool
itself doesn't appear until you click through it. It asks again on every visit,
because I deliberately don't remember your answer — no cookie, nothing stored,
which also keeps the "nothing you type leaves your browser" claim literally true.

One thing I'd push back on gently: you floated putting warnings at several stages
through the tool. I'd keep it to the one gate, because repeated warnings get
click-blind within about two uses, and a warning nobody reads is weaker in
practice — and probably in law — than a single one someone had to act on. What I
did instead is put a short caveat line *inside the results*, so if someone
screenshots the output into a deck, the caveat travels with it. That felt like the
gap that actually mattered.

**On aiming at students — the safety instinct is dead right, but I don't think
you need to narrow the audience to get it, and narrowing costs a lot.**

Here's the thing I found when I worked it through. Almost all the *risk* on this
site comes from one specific act: stating live facts about real named companies.
Almost all the *value* comes from something different: explaining the mechanism —
how new money leapfrogs existing lenders, how a majority binds a dissenting
minority, what a court has to be satisfied of.

Those two are separable, and that's the useful bit. You can teach the entire
mechanism using numbers that are openly made up, and made-up numbers can't be
wrong about anybody. "Suppose senior debt of six billion" is completely accurate,
teaches the whole lesson, and carries no exposure whatsoever. The danger lives in
the illustration, not in the teaching.

So the safest version of this site isn't "aim at students" — it's **lead with
mechanism, and treat real live cases as clearly-marked illustration.**

And the "treat this as a possibly inaccurate case study" posture you want costs
nothing with a professional audience, because professionals already expect it.
Every legal update and broker note says check this before you act. A restructuring
lawyer wouldn't dream of relying on someone else's summary of a judgment without
opening the judgment. So we can take your whole honesty posture without moving the
audience an inch.

What narrowing to students would cost is the money. The plan to sell deal packs at
a few hundred pounds rests entirely on a reader with an expense account. Students
have no budget and no procurement. Choosing students as *the* audience is choosing
to make this free indefinitely — which is a perfectly reasonable choice, but it
should be a decision you make, not a side effect of a safety decision.

**What I'd recommend instead:** take the honesty posture completely; reframe the
site as teaching material rather than a research product, which is what makes that
honesty coherent instead of apologetic (a case study is *supposed* to be a thing
you interrogate; a research report saying "this might be wrong" is just a bad
research report); and define the audience as anyone learning how this works —
students, trainees, early-career professionals, and people in adjacent seats like
the corporate lawyer trying to understand what the credit team is doing. Bigger
than students, same needs, and it contains people who can pay.

It also quietly improves the Thames Water piece. It stops trying to be the
definitive account of Thames Water — which we can't win against services with real
reporters, and which is the risky claim anyway — and becomes a live case to test
your understanding against. Here's the mechanism, here's a real fight where it's
being argued, here's what the documents say, here's our reading, check it. That's
a claim we can actually keep.

One possible source of money that survives all of this: training material for
trainees and graduates inside law firms and advisory shops. Same content, and the
buyer has a budget. I want to flag clearly that I have no evidence for that — no
conversations, no pricing, nothing. It's a direction worth testing, not a plan,
and I'd rather say so than have it harden into a fact by being written down
confidently.

**So the decision I need from you** is which of those three you want: my version
above, or students proper as the target, or keep the mid-market professional as
primary. The site is mid-build right now, so this is the cheapest moment it will
ever be to change — the briefs would need a revision either way, and pages haven't
been written yet.
