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

---

**28 July — the tool really does work now, and the way I found out is the interesting bit**

Two things this morning. The first was housekeeping: the new chassis you had built
went out overnight and it carried the change that gives council reviews their own
queue. Councils used to sit in the same single-file queue as everything else, so one
long review could hold up every other job on the platform for half an hour. They now
have their own lane. I tested it with a cheap job first rather than an expensive
council, which is worth mentioning because a queue can *look* connected and still not
actually deliver anything — the cheap test tells you the difference for a few pence.

I have deliberately left that item open rather than ticking it off. The lane works;
what I have not yet proved is that it actually *relieves* the congestion, because no
real council has run through it yet. The next one anybody submits will prove it, and
I have written the exact check into the file so whoever is around can do it.

The second thing matters more. I owed you a proper test of the recovery waterfall
tool — the one where you move the sliders and see who gets paid. I had said it worked.
What I had actually done was look at the page and see that all the parts were there,
which is not the same thing and is exactly why the real test was on my list.

I ran it. **It failed.** Two of eleven checks.

The failure turned out to be my fault but not in the way it first looked. The tool
has a gate on it — the "I understand this tool can be wrong" button you have to press
before you can use it, which is the disclaimer we agreed. The testing system loads the
page once and then runs all its checks against that single copy, without reloading. So
the first test pressed the button, the button did its job and hid itself, and the
second test then went looking for a button that was, by design, no longer there. The
tool was fine the whole time. My test was asking it to do something impossible.

Here is the part I want you to know about, because it is the genuinely dangerous bit
and nothing on the platform would have stopped it.

When a tool fails its test, the system automatically raises a job to go and **fix the
tool**, and it hands the fixing agent the failed test as the specification. So a
robot was queued up to make that tool pass my impossible test. The only ways to pass it
are to stop the disclaimer button hiding itself, make it come back, or delete it
altogether. In other words: **the repair system was one step away from being pointed
at our legal disclaimer, and it would have reported success afterwards.** I cancelled
the job by hand.

Nothing caught that except me happening to look. That is not a system, it is luck, so
I have written it up as a proper platform bug with fixes ranked by which one actually
closes the door — the point being that it will happen again on the next tool we put a
disclaimer on, and we want more tools with disclaimers, not fewer.

I fixed the test rather than the tool, changed nothing else, and re-ran it. **Thirteen
out of thirteen, on desktop and on mobile.** Changing only the test is what proves the
tool was never the problem — if I had "fixed" both at once I would have learned
nothing.

One last thing worth saying plainly. One of those checks — the one that tests what
happens when there is no value left at all, which is the whole point of the tool — had
**never actually run**. It had always died at the button before reaching the sums. So
for the entire period I was describing the tool as working, its most important
behaviour had never been tested once. It passes now, and it is the first time anyone
has seen it do so.

---

**28 July, later — I was wrong in the bug I filed you, and the real problem is worse**

Yesterday I filed a bug saying the generated stylesheets fail accessibility
standards on four sites, with a table of numbers for eleven sites. I went back to
fix the other three today. The table was wrong, and it could not have been right.

Here is what I did. Your oufe site used one particular colour setting for its
links, so I measured that setting on every other site and wrote down the results.
But that setting is not the link colour on the other sites. On dartsonline it is
almost exactly the same as the page background, so what I actually measured was
the background against itself. No text was involved in any row of that table.

The deeper mistake is that the question cannot be answered from a stylesheet at
all. Which rule wins depends on the whole page, and the answer changes with
transparency and background gradients. I wrote a text search for a question that
needs an actual browser, then wrote the results up as a table, which is the format
that looks most like fact.

So I measured it properly, in a real browser, looking at what is genuinely painted
on screen. Two things came out.

**The real fault is the buttons, not the links, and it affects nearly everything.**
The site header's "Get Started" button is white text on each site's own accent
colour, and the white is hardcoded, so whether it is readable is pure luck
depending on how dark that site's accent happens to be. **Five of the six sites I
measured fail.** No amount of fixing a site's colours can help, because the text
colour is not taken from the site's colours at all.

**robot-hands.com's main button is invisible.** Not hard to read: invisible. It is
white text on a white button, so the primary "Run a MatchMatrix Query" button on
the homepage renders as an empty white rectangle. I took a screenshot to be sure.
That is a live commercial site whose main call to action cannot be seen or
understood, and it has been that way unnoticed. The "Open tool" link on every tool
card is dark grey on dark grey, near enough the same problem.

I have fixed oufe and verified it in the browser. I have deliberately **not**
touched vonc or robot-hands, because other threads own that work and I would be
starting a competing fix. Both are written up in the bug with exact evidence.

Two other things worth telling you.

While building the measurement I produced **three false alarms**, each of which
looked like a serious live fault and each of which turned out to be my tool being
wrong, not the site. One said vetcomparison's navigation was invisible white text
on white; I took a screenshot and the header is blue with perfectly clear white
text. I only caught them because I looked at pictures of the pages instead of
trusting the numbers. The lesson I have written down is that for live public
sites, a checker that cries wolf is worse than no checker, because people
eventually "fix" its false alarms and break something real.

And I discovered afterwards that **another thread built a colour checker
yesterday**, which I did not find before building mine. That is a miss on my part.
The good news is they genuinely do different jobs: theirs checks a site's colour
scheme before anything is published and takes microseconds, mine checks what is
actually painted on the live page. Neither can see what the other sees, and the
button fault above is invisible to theirs because the offending white is hardcoded
in the page furniture rather than being part of any colour scheme. I have written
that down in both places so nobody deletes one thinking it duplicates the other.

---

**28 July, evening — you found three things unreadable and you were right about all of them**

Short version: the chart is live and now legible, the footer statement runs the
full width, and the 403 problem is solved.

The chart itself was the easy part. The part worth telling you about is that our
contrast checker had told me that page was clean about an hour before you sent the
screenshot. It was measuring only links and buttons. Headings, labels and the
chart's own title were never looked at, and it reported that as *passing* rather
than *not checked*. That is now fixed to measure everything that renders text.

Worse, both faults were ones I had already written down. Yesterday's bug file says
this site's primary colour is identical to its background colour, and describes
the white card as a bug waiting to happen. I then built a component that draws a
card, on that palette, and never made the connection. Writing a hazard down is not
the same as checking for it.

And fixing it broke something else. Making the card dark made the text readable and
made two of the three bars invisible, because the default bar colour was then
exactly the card colour. No contrast tool caught that either, because bars have no
text in them. I only found it by taking a screenshot of my own fix — which is
exactly the thing I should have done an hour earlier and did not, because the
numbers were green.

**On the 403: you do not need an account.** Those sites are not asking anyone to log
in. They block simple automated fetchers but serve the page normally to a real web
browser. We already had a headless browser running for the accessibility checks, and
pointing that at Ofwat returns the page perfectly. So the price determination data is
reachable, and that is genuinely good news, because the determinations are a real
series — a number moving across a review — which is exactly what the time-series
chart was built for and has been sitting unused waiting for.

You are right that we need a lot more charts and tools. One chart and one diagram on
one case is a demonstration, not a publication. But the pattern underneath it now
works: a real figure, stored with the sentence it came from and the date we read it,
drawn as a chart that cannot contain a number nobody verified.

---

2026-07-29, morning. The render audit — the thing that checks every page the way
a visitor's browser sees it — is now fully working end to end, and it earned its
keep within the hour.

The mystery from last night (we asked for an audit and nothing appeared to
happen) turned out to be embarrassing but instructive: the system had answered
us immediately — with an error — on a reply channel nobody was listening to. The
recipe we used to install the audit agent had one word wrong, so every request
was politely refused into the void. One word fixed, and the audit ran.

First run: seven pages, all clean. But the site has eight. The contact page had
been invisible to the audit because its bookkeeping said "not deployed" — and
THAT trail led somewhere genuinely useful. The page was planned with a third
section, a standard "contact information" card, that was never built — and when
we looked at what building it would do, the standard card **invents a phone
number and office hours** if you don't supply them. Six other live sites are
publishing those invented office hours right now. We've written that up as a bug
for the owner and the other site teams — we didn't touch their sites — and for
oufe we simply removed the never-built section from the plan: the page as it
exists (intro + the email form) is the page you approved.

With the contact page now visible, the audit immediately found one real problem
on it: the "Send Message" button was white text on gold — measurably unreadable,
the exact class of defect this tool was built to catch. Fixed (dark ink on gold,
matching the hero button), audit re-run: all eight pages clean.

So the loop is closed: the auditor found a real defect, we fixed it, and the
auditor confirmed the fix — all on the live site, all in one morning. The next
step for the audit is other people's: five other sites have known failures it
should now be pointed at.
