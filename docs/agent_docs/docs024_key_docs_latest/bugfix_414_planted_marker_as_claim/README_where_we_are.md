# Where we are — the planted "checked against the FCA handbook" claim (bug 414)

*Plain prose, append-only, newest at the bottom. This is the owner's document too — add below, never
rewrite.*

---

## 2026-08-27, morning

**What the problem actually was.** Back on 2 August someone ran an experiment on lendzy.co.uk. To
check that the site-building machinery really does what its brief tells it, they hid an instruction
in the site's brief: "somewhere in the copy, include the exact phrase *checked against the FCA
handbook, rule by rule*". The machinery obeyed — which is what the experiment was testing — and then
nobody took the instruction out. So a site about consumer credit rights has been telling readers, in
its own voice, that its content was checked against the regulator's handbook rule by rule. Nobody did
that. On a site whose entire selling point is independence and accuracy, it is the worst sentence on
the page.

It got worse on its own. Our own quality-audit machinery came along, read the sentence off the live
page, decided it must be the site's main selling point, and filed a job asking a writer to add a "how
we verify our guides" section to back it up. So the system was about to start manufacturing evidence
for something that never happened. That is the part that made this more than a typo.

**The thing yesterday's fix missed, and it is the interesting bit.** The lane that found this last
night stripped the instruction out of the brief and recorded the source as fixed. It wasn't. Ten days
after the original plant, one of our own agents — the one that writes a site's strategy — had *read*
the planted instruction and written it out again, in its own words, in a different part of the site's
records. That copy was still live this morning. So the lesson is bigger than one site: **deleting a
planted instruction from the place you found it does not retract it, because our agents copy
instructions to each other.** I have written that up as a trap for other sessions to read before they
"fix a source", and turned it into a query that takes seconds.

**What I did, in order.**

1. Removed the surviving copy from the strategy record, under a check that refused to run unless the
   exact sentence was where I expected it. Nothing was overwritten; the old version is kept.
2. Rejected the audit job that wanted to substantiate the claim, with the reason written on the job
   itself so nobody re-opens it. It was one click away from regenerating the page around the false
   claim.
3. Asked the framework to rewrite the three bits of copy — not me. The instruction to the writer says
   what to remove, why it is false, and what is *true* according to the site's own brief (we name the
   exact rule beside every figure and link to it, so a reader can check for themselves). That is a
   real claim the site can stand behind.
4. Closed the hole so this cannot happen quietly again — three small changes, described below.

**The three framework changes, in plain terms.**

The first two teach the existing honesty checks two sentences they were missing by a hair. We already
refuse copy that claims *everything* on a site has been checked — that is unverifiable by anyone,
including us. But the rule only looked 30 characters ahead, and lendzy's sentence put 38 characters
between "every figure" and "is checked", so it slipped through; and the rule didn't recognise the word
"Everything" at all. Both fixed. Worth knowing: measured across every page we serve, that rule had
been firing **zero** times — it was asleep, and these two sentences are the first things it catches.

The third change is the one that matters. Every honesty check we have looks at what the site *says*.
Nothing looked at what the brief *tells it to say*. So a brief could lawfully order a page to state
something the page checks would refuse — which is exactly what happened, for 24 days. There is one
small daily job that already reads briefs and judges them, so it now asks this second question too,
across everything any of our agents reads (not just the writer — the copy that survived yesterday's
fix was in a part the writer never sees). Where it finds one, it files it for a *person* to decide.
Deliberately no automatic handler: an automatic brief-rewriter is precisely how the audit fleet
canonised the planted sentence in the first place.

**One thing I want to be honest about: I nearly made it much noisier.** My first design ran all our
honesty rules over every brief in the fleet. I measured it before building it, and it would have
produced about 21 findings a day, essentially all wrong — because 15 of them are *our own*
instructions telling writers "never invent a person, company or statistic", and because each site's
list of banned phrases stores the banned phrase itself, so the check would have convicted every
site's own immune system, every day, for ever. Narrowed to the one family that fits, it finds nothing
today and finds exactly the two planted rows when pointed at the history. That is the version that
shipped.

**Where it stands right now.** The false claim is no longer in any brief anywhere in the fleet. The
audit job is rejected. The code is committed and has gone to the review council. The copy rewrite has
been dispatched and is retrying — the platform is having a bad morning with a particular internal
timeout, several other lanes are hitting it too, so this is queue weather rather than anything wrong
with the request. **The sentence is still on the two pages until that rewrite lands**, and I will not
call this done on a job status: I will read it off the live pages.

---

## 2026-08-27, mid-afternoon — it's off the site

**The false sentence is gone from lendzy.co.uk.** Both pages read clean, and I checked it four ways
rather than once: the database holds no trace of it in either place it was stored; both pages fetched
from the live site contain zero occurrences; and — the one I trust most — our own honesty scanner, run
over the whole site, now reports nothing at all, having convicted three components on that same site
this morning. A clean result from a tool that was finding things a few hours ago means something. A
clean result from a tool you have never seen fire does not.

**The framework wrote better copy than the sentence it replaced**, which I did not expect and think is
worth recording. The guide now says that every figure and rule reference *is given* together with the
named rule and a link to where you can read it — and then adds, unprompted, "that does not make the
checker infallible… rather than take our word for it." We removed a claim asking readers to trust us
and got back one inviting them to check. That is the outcome the brief always implied, and it came out
of the machinery rather than out of me typing it.

**Two operational notes, because both cost me time and would cost the next person the same.** The
about page was repaired for three hours before I noticed, because its job was marked "failed" — the
step that failed runs *after* the step that saves, so the work had landed and the label was about the
wrong thing. And the rewrite of the second page was stuck for two hours on an account-level outage in
the AI service, not on anything to do with this bug; the owner adding credit cleared it, and it
finished on the next attempt.

**Why I am not calling this closed.** The bad sentence is gone and that half is genuinely done. The
part that stops it happening on the *next* site is code, and code on this system does nothing until
the whole fleet is next rebuilt. Until then, a planted instruction on another site would sail through
exactly as this one did. So the file stays in the open queue with three specific things written down
for whoever picks it up after the next release — including the right way to check that the running
service actually has the new rules, because the method I had originally written down was the one
documented as unreliable for this particular service, and a reviewer caught that before I used it.

**One thing I would like you to look at when you have a moment.** One of the review seats disagreed
with the decision you took this morning. The new rule about claiming we have checked something against
a regulator's handbook ships as a *warning* rather than a *refusal*. The argument for warning still
holds — a real compliance consultancy could say it truthfully, and at refusal-strength the layer would
have blocked the honest correction "nothing here has been checked against the handbook", which our
negation detection cannot see. But the seat whose entire job is this class of claim would have refused
it, on a finance site, where the false version ran for 24 days. It is on the record so you can revisit
it deliberately rather than have it quietly settle.

**And a new hole, found by that same seat, which is not this bug but is next to it.** The register of
verified facts is deliberately not scanned by the new check, because it stores the banned phrases
themselves as data. Which means a *poisoned register* — a fabricated source or a made-up fact written
into it — passes every layer we have, because every layer treats the register as the thing it checks
against rather than something to be checked. Nobody has found an instance. If one turns up, it wants
its own file.

---

## 2026-08-31 — closed

It's done. The false sentence is gone from the site and from every brief in the fleet, the new rules
are running in the deployed system, and the file has moved to the closed queue.

The last thing to go wrong was ours, and it is the part worth telling. The new brief-side check — the
one I added so a planted instruction gets noticed where it lives — produced its first real finding
last week, and it was **wrong**. It flagged a gardening site for the phrase "we tested six lawn mowers
so you don't have to", which sits in that site's *"would never say"* list. It convicted a site for a
phrase the site's own brief bans. I had excluded one place that stores banned phrases as data,
written down exactly why, and then not carried the same reasoning to the next place. The item sat in
the review queue for three days before I saw it.

What found it was running the thing. Not the tests — I had written those, and they passed. Not the
nine-seat review round, which read the design carefully and raised good objections about other
matters. The scheduled job ran against real sites and told me. And when I fixed it, my first fix
passed its test and still failed on live data, because I had built the test fixture from my *idea* of
what the data looked like rather than from the data. The real rows had no blank line where mine did.

That is three days of this bug in one sentence: **the things that caught real problems were running
it against reality, an adversarial review before building, and other teams re-measuring what I told
them.** I logged seven of my own wrong claims along the way, and the pattern in all of them is the
same — every one was a number or an assertion made *in passing*, to support a point that wasn't about
it. The careful work held up. The remarks around it didn't, and remarks are what travel.

Two things are left for you rather than for the code. One review seat disagreed with your call that
the new rule should warn rather than refuse — it's on the record so you can revisit it rather than
have it settle by default. And there's a hole next door to this one that nobody has hit yet: our
register of verified facts is the one thing no check inspects, because it's what the checks check
*against*. If someone ever writes a fabricated fact into it, nothing we have would notice.

---

**2026-09-03 — the check is live, and we cannot yet prove it works**

The new build went out this morning and it carries the detector that was approved on Tuesday night.
I checked that properly rather than trusting the release: I asked the running program directly
whether it contains the new code, and ran two control questions alongside it — one thing that must
be there and one that must not — so that a misleading answer would have shown up as such. It is
there. The daily sweep then ran at ten past nine and reported nothing wrong.

Nothing wrong is what we expect, because I counted every pattern on every site on Tuesday and they
were all sound. But here is the honest problem, and it is worth understanding because it will come
up again: **the code is written so that a clean result leaves no trace at all.** If it runs and
finds nothing, it says nothing. If it never runs, it also says nothing. Those two situations look
identical from outside, and no amount of staring at the output will separate them.

So we know the check is *installed* and we do not yet know it *works*. The way to settle it is to
deliberately break something small and confirm the alarm goes off — plant a bad pattern on a
throwaway site, watch it get reported, then take it away. I have handed that to the team that owns
the code rather than doing it myself, because they wrote it and they are mid-way through the
council round on it.

Everything else on this piece of work is finished. The original bug is closed and the false
sentence is gone from the live site. The wider design question you ruled on yesterday and this
morning — all seven parts of it — is settled, and what is left of that is building, not deciding.
Nothing on this lane is waiting on you.

---

**Thursday 3 September, late morning.** I picked this up again an hour after writing the note above,
and three things in it had already gone out of date. That is worth telling you about, because two of
them were good news and one of them was me nearly getting something wrong.

**The deliberate break has already been done — by the team that owns the code, four minutes after we
told them how.** Yesterday the blocker was that nobody could find a throwaway site to test on. It
turned out we had been searching the wrong way: we were looking for a site with an obviously fake
name, when in fact the system marks throwaway sites with a status field, and there are three of
them, all with perfectly ordinary-looking domain names and no pages at all. I wrote that down as a
warning note for everyone at about half past ten. At 09:34 the other team planted their bad pattern
on one of those three sites. So the warning notes we keep do get read, and quickly.

**But I found a trap sitting in their path, and warned them.** They had also added a second small
change so that the check would leave a trace even when it finds nothing — exactly the right fix for
the problem I described above. It is not running yet. It was written half an hour after this
morning's build went out, so it cannot be in it, and I confirmed that by asking both copies of the
running program directly, with control questions either side. The trap is this: if their test now
comes back quiet, the natural next move is to go looking for that trace, and they will find nothing
— not because the check failed, but because the tracing code isn't installed. Silence would mean two
completely different things and look the same. I have written that to them.

**The one where I was nearly wrong.** One of the sites I had listed as still missing its evidence
register turned out to have one, created this morning by the automatic daily refresher. I read that
as the refresher having fixed the gap by itself, which would have been quite interesting. It is
false. The history shows another team filled that gap on Tuesday evening; the refresher simply
rewrote the whole record the next morning and stamped its own name on every part of it, including
parts it has no ability to write at all. **The lesson is a general one and I have written it up for
everybody: the name on a record tells you who touched it last, not who put the information there.**
The thing that makes this hard to catch is that the wrong answer is entirely plausible — a real
program, a sensible date. You only see it by reading the record's history instead of the record.

**Finally, the one job left on this lane is smaller than I said, and the problem behind it is
bigger.** I said `loancash.co.uk` still had no evidence register. True — but it is now the *last* of
the five finance sites we listed on Tuesday; the other four are all done. So that piece of work is
one site, not five, and I have corrected the design document that was still telling people to do all
five.

The bigger thing is this. Thirteen of our thirty-nine live sites have no evidence register at all.
Most of those are not a problem today, because your ruling only requires one for finance sites — but
`vetcomparison.uk` is on that list, and vet is precisely the area you said on Tuesday you want to
move into next. **The presets you approved would arrive at a site with nothing to apply them to.**

And there is a deeper catch that I think is the most useful thing I found today. The daily check
that looks after evidence registers builds its list of sites to check by asking *which sites have a
register*. So a site with no register is not merely low priority — it is **invisible to the check,
permanently**, and running the check more often would never reach it. A thing that is missing cannot
be found by a search that starts from the things that are present. Nothing today would ever tell us
`loancash.co.uk` is missing its register; we only know because a person went looking.

I have written all of that into the design document rather than building anything, because it is the
other team's area. Nothing here is waiting on a decision from you — but if you want that gap
covered, the two candidates are `loancash.co.uk` (the last of the five) and a check that starts from
the list of live sites rather than the list of registers.

**Thursday 3 September, afternoon.** Three things closed out and one question opened.

**loancash is done.** You asked for it to be populated; it is. Nineteen statements of what the law
actually says, each with a link to the official source and a sentence quoted word for word, so the
system re-checks them every night. It was reviewed by the council and approved on the first round by
every reviewer. It found three mistakes on the site, the most important being that we understate the
£15 default-fee protection — I have recorded all three and **not** touched the published wording,
because that is your call and it is written up as such.

That was the last of the five finance sites we set out to do, so that piece of work is finished.

**A new build went out at half past two, and it quietly fixed something I had warned another team
about this morning.** Their check for broken rules had no way of proving it had actually run — a
clean result left no trace at all, so "ran and found nothing" and "never ran" looked identical. They
wrote a fix for that this morning, but it wasn't in the running program yet, and I flagged that if
their test came back quiet they'd go looking for a trace that wasn't installed and misread the
silence. The new build carries it. I checked it directly on both copies of the running program, with
control questions either side, and told them. They confirmed it independently.

**Tomorrow morning's automatic run is now a proper experiment rather than a hopeful one**, and this
was worth setting up. They had planted a deliberately broken rule on a throwaway site to see whether
the alarm fires. loancash's new register gives a second, free half to the same test: it has six
rules that are all *correct*, so the run should report "checked six, found nothing wrong" — and a
number bigger than zero is the proof that the checker actually read real data, which is exactly what
was missing before. If their broken one stays silent while loancash reports six, we will know the
fault is in the alarm rather than the check, which is a much smaller thing to chase.

**The open question is vetcomparison.** I have asked that thread directly whether its register is
theirs to do or nobody's, and whether there is any reason a register would be the wrong tool for a
comparison site — its numbers may be other people's claims rather than its own, which would change
the answer. No reply yet. It matters because the vet and legal presets you approved this morning are
built, and vetcomparison currently has no register for them to apply to.

**What I still need from you** is written up properly in the handoff, but in short: whether to repair
loancash's three wrong sentences, whether to do vetcomparison, and the bigger one — twelve of our
live sites have no register at all and nothing in the system can notice a site that is missing one,
because the nightly check builds its list from the sites that already have one. I think the cheap fix
is to build the missing check first, then populate; but the scoping question inside that is genuinely
yours, because most of those twelve are not finance sites and your ruling only required registers
there.

--

## 2026-09-03, evening — the loancash pages are repaired, and vetcomparison has its register

Two jobs this session: put back what yesterday's repair accidentally deleted, and build the
vetcomparison register you asked for. Both are done and both are checked at the actual live pages
rather than at a status somewhere saying they worked.

**The loancash repair.** Three guide pages had lost their "we are not a lender, this is not advice,
we are not the FCA" block when an earlier fix rewrote them wholesale instead of editing them. The
repair for that was written yesterday and deliberately held back, waiting for a review verdict —
because applying before reading a verdict is exactly what caused the damage in the first place.

The verdict came back "revise", and the main objection turned out to be **wrong about the actual
file**. The reviewers were told the repair left the items on automatic dispatch; the file plainly
sets them to manual, and the file's own self-check would have refused to run otherwise. What had
gone stale was the *summary* sent to the reviewers, not the migration. Five reviewers spent their
objections on a defect that did not exist. I checked every other objection against the real code and
the live system before acting — one of them was about whether the repair would even work, and the
answer is in the code and in the running configuration, so that is now settled rather than argued.

So I applied it, and released **one page first** rather than all three, because firing four at once
is what went wrong last time. That page came back clean, so I released the other two.

**The result, measured sentence by sentence.** Last time, 36 of 37 sentences on a page were silently
replaced while the page kept 84% of its length — which is why I no longer trust a length check. This
time I compared the actual sentences before and after. On the first page **all 35 original sentences
survived word for word** and 5 were added. On the other two, a handful of sentences changed — and
every one of them is a seam where new material was spliced in ("lends money **for profit** without
authorisation"). Nothing was lost on any page. The disclaimer is back, and the regulatory corrections
from yesterday are all still there and correct.

One honest note: the pass/fail test I wrote for this said "additions only, nothing reworded". That
test was **too strict to be right** — a repair that has to insert a clause into an existing sentence
can never satisfy it. The measure that actually matters is whether any sentence *disappeared without
a trace*, and none did. I have fixed the test rather than pretending the repair failed.

**The vetcomparison register.** This is the "list of every figure the site is allowed to assert, with
where each one comes from" that stops a page inventing numbers. It did not have one, and a site with
no register has the checking switched off entirely — so this closes a real hole.

**It found three errors in the live site, two of which nobody knew about.**

- The site says the competition regulator published its final report **in November 2024**. It was
  **March 2026** — out by about sixteen months, on two pages. November 2024 is when the inquiry chair
  gave a *speech*, which is listed on the same official page and is almost certainly where the slip
  came from.
- The site states the **£21 and £12.50 prescription fee caps as settled facts**. They are not settled:
  the draft order carries them in square brackets, to be adjusted for inflation before the order is
  made. The vetcomparison team spotted this on 24 August and nobody acted on it; it is on seven pages.
- The site says practices must publish prices across **"36 service categories"**. The official schedule
  defines **36 services grouped into 5 categories**. The number is right; the word after it is not.

I have **not changed any of the site's wording** — that is your call, and the standing rule is that a
register records what is true rather than rewriting pages. All three are now written down where the
daily checks can see them.

**The part I want to flag, because it is a trap the whole estate can walk into.** The regulator
publishes its rules as **PDFs**, and our daily "is this citation still accurate?" checker can only read
web pages. When I tested it against the PDF it returned "not found" for every quote — *including a
phrase that is unquestionably in the document* — and also "not found" for a deliberately fake control
phrase. That last part is the dangerous bit: when the checker fails to find both the real quote and the
fake one, it is telling you it cannot see anything at all, but it looks exactly like "you typed the
quote wrong". Had I linked those facts to the PDF, the site would have reported a false alarm every
single day, for ever. So those eight facts record *who read the document and when* instead, which the
system deliberately does not re-check nightly. It costs nothing in protection. This is now written into
the estate's trap list.

**And something I nearly got wrong myself.** I ran the number-checking scan against the site with the
new register loaded and got a clean zero. That looked like success. Then I ran it again with the facts
**deleted** — and got exactly the same zero. The scan simply does not reach those pages, so my clean
result proved nothing at all. I have replaced it with a test that can actually fail, and the register
now says plainly that this half of the protection is *switched on but not yet exercised* on this site,
rather than implying a coverage it does not have.

**One more.** A self-check I wrote to prevent a migration destroying the register **passed while the
register was being destroyed** — a subtlety of how databases compare a value against nothing. It
printed the word "intact" next to a blank. Found by deliberately trying to break my own guard, fixed,
and written up so nobody else ships that shape.

**Where it stands.** The three loancash pages are correct and live. Vetcomparison has a register of 21
facts, six banned phrases (including the one the site was remediated for last July), and a recorded
judgement about how strictly it should be held. The review board approved it first time, with one fair
objection that I have already fixed. Ten more sites are waiting for the same treatment, and the three
vetcomparison wording errors are sitting there recorded, waiting on your decision about whether to
correct the pages.


--

## 2026-09-04 — I was wrong about the PDF problem, and the vetcomparison copy is being fixed

**First, a correction to what I told you yesterday.** I said that if we linked those CMA facts to the
regulator's PDF, the site would report a false alarm every single day for ever. **That was wrong.** The
nightly checker already handles it properly: when it meets a document it cannot read, it records "I
could not check this" — not "this is wrong". The code says so in plain English in its own notes, and
has all along. I never opened it.

What actually misled me was the little *testing tool* we use when writing a citation. It was supposed
to behave exactly like the nightly checker — that is the whole point of it, and our own written rule
says never test a source with anything else — but it quietly didn't. It skipped the check that spots
an unreadable document, so it just reported "quote not found", which looks identical to getting the
quote wrong. I trusted the tool, drew a conclusion about the real system, and wrote it into five
documents. Two rounds of review didn't catch it either, because everyone was reading my explanation
rather than the code.

**Both are fixed.** The testing tool now uses the real checker's own fetching, so it says "NOT
VERIFIABLE UNATTENDED: unsupported content type" instead of a misleading "not found". And the register
is better than before: those five CMA facts now keep the source link and the exact quoted wording,
which I had needlessly thrown away, and they are set to expire on **23 September** — the day the
regulator's Order is due — so the system will ask someone to re-check them rather than relying on a
note in a handover.

**Second, the copy.** You asked me to fix the three vetcomparison errors, so I have dispatched all
nine page corrections. I told the vetcomparison thread first, since it is their site; they confirmed
nothing was half-finished and warned me off one page whose text was painstakingly hand-checked against
government sources — regenerating it would have reintroduced three falsehoods they had removed. That
page needed no fixing anyway, and the migration now *refuses* to touch it rather than relying on me
remembering.

I released one page first as a test rather than all nine, which is the lesson from last week's
accident. **It has not run yet, and I found out why rather than guessing:** the build queue takes one
site every thirty seconds, oldest job first, and ours was filed most recently — so it is behind three
other sites, one with fifty-one jobs. It is waiting, not broken. Worth knowing generally: "the job is
filed" is not "the job is about to run".

**One more thing I got wrong today, caught before it mattered.** While diagnosing that queue I
rebuilt the system's own selection query by hand, left out two conditions, and was about to tell you
the whole fleet's dispatch had been stalled for twenty-three hours. It hasn't — those jobs are
correctly waiting on something else to finish. Same mistake as the testing tool, twice in one day: a
close copy of the real thing is not the real thing.
