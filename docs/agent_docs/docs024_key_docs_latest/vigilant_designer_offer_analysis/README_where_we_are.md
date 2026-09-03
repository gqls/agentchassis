# Where we are — vigilant designer + offer analyser

Plain-prose running log. Append-only, newest at the bottom. The owner writes here too — never
edit or reorder existing entries; add dated corrections below instead.

---

## 2026-08-02 — the programme exists, and why it looks the way it does

You asked for three things: a constant vigilant designer working from whole-site down to the
finest detail, and an offer analyser (with benefit analysis inside it) that keeps asking whether
each site actually answers its target market's need in a way that pays us.

The surprising finding from digging: you already asked for about half of this on 24 July, and
the platform half-built it. There are written specifications for a screenshot-based design
critic, a "did we build what the brief asked" audit, and checks for components nobody uses —
none of them built. And the one that DID run (the brief-fidelity audit) predicted your
fundamentallyai complaints three days before you made them — but its findings sat in a queue
nothing reads, and you found the problems with your own eyes instead. That is the real lesson
this programme is built around: the platform is reasonably good at noticing things; it is bad
at doing anything about what it notices. So every detector we build ships together with the
thing that acts on it, and we prove each one by watching a real finding travel all the way to
a changed page.

On the design side: the reason all the sites look the same is now measured, not a feeling. The
page planner is only offered a brochure-shaped menu; the design AI is only allowed to change
colours (everything else is owned by shared layout templates); three shared header components
serve all fourteen sites; and five interactive components were built that no planner has ever
picked. Each of those gets its own fix.

On the offer side: each site already has a written monetisation analysis and audience
definition from build time — and nothing ever reads them again. The agent that reviews sites
doesn't load them. So the first offer work is simply wiring the premise back into the judgement,
then adding a check for the obvious mismatch ("a tools site with a consultancy CTA"), then a
proper offer analysis agent. On the live estate the recorded revenue model is mostly "none" or
"implied", so there is also a premise-refresh path to fill that in — carefully gated, because
the strategy agent currently kicks off a full rebuild chain whenever it runs.

Your four calls today: manual per-site triggers for now (the always-on switch stays off until
you say otherwise); designer first; trial both Gemini and Claude as the critic's eyes; and
broad autonomy — the designer may apply changes at every level, with two honest exceptions
(shared header/footer components get forked per-site before editing, because editing the shared
one changes every site at once; and whole-site layout swaps stay as review items because the
platform deliberately has no automatic path for those yet).

Next up: the plumbing phase — make findings flow at all, fix the counter that falsely reports
audited-out sites as "clean", and turn the browser audit's measurements into actionable work.

## 2026-08-02 (later) — the plumbing is in, and the platform now argues back usefully

Since the morning entry: the two live-config changes are applied — the sweep points at the
right lane (still switched off, that flip stays yours), and the dishonest three-audits-
and-report-clean counter is replaced by an honest gate: audit when something changed or
enough time passed, always move findings toward workers, and when three audits at an
unchanged site produce fixes that never land, say so on the roadmap instead of saying
"clean". Three code pieces are written, tested and committed but dormant until we ship a
new image: the piece that turns browser measurements into actionable work, whole-site
screenshot capture (both phone and desktop views), and the "eyes" — the ability to send
those screenshots to an AI model, built for both Gemini and Claude so your trial is a
switch, not a rebuild.

The review council looked at the screenshot work and approved it first time, with a
genuinely good catch: our capture failures were logged but not counted, so the future
design critic could have been quietly judging half a site without anyone knowing. That
counter now exists. Two more reviews are still in the queue.

Next session: ship the new images, prove one finding travels all the way from "a browser
measured it" to "the page changed", and then build the critic itself and run your
Gemini-versus-Claude comparison.

## 2026-08-03 — the two overnight objections answered, and both taught us something

The review council had sent back both remaining code pieces with one blocking objection
each. Both are now answered and resubmitted, and each objection turned out to be valuable
for a different reason than the reviewer intended.

On the findings-writer: the reviewer feared repeat findings would silently vanish after
two fix attempts. They don't — the third one is recorded and labelled "unresolved after
2 attempts" — but chasing the claim exposed two things we actually had wrong. First, our
own notes said these parked items show up in the fix loop's daily digest; they don't —
they show up on the admin dashboard's attention lists, and the notes now say so (logged
as a wrong call). Second, and bigger: the agent that's supposed to fix contrast problems
has never once been given a work item in the platform's whole history, and the prompt it
runs expects fields our findings didn't include — so our findings would have arrived as
blank forms. The revision writes the fields its prompt actually reads, plus the page
reference, so the first item that agent ever receives will be one it can act on.

On the "eyes": the reviewer said we claimed truncated AI responses fail loudly but never
built the mechanism. The mechanism has existed since 20 July — the reviewer was reading
our own concept register, which still described the bug as unfixed two weeks after it was
fixed. So the code was right and the paperwork was wrong, and wrong paperwork cost us a
review round. The register entry is corrected, a trap note now warns future readers (and
future reviewers) that register status lines go stale, and two new tests prove the
truncation alarm fires through the new image path specifically, for both AI providers.

Both revisions are committed and back in the review queue. Next: read the two verdicts,
then ship the new images and prove one finding travels browser → work item → changed page.

## 2026-08-04 — the drain works: we watched a finding travel, and the counter bug is closed

This morning the whole point of Phase 0 got proven on a real site. We hand-fired one
improvement sweep at relojistas.com (the watch-news site rebuilt at the end of July).
Before firing, we checked its queue by hand: five of its six old findings described the
site as it was before the rebuild, so we cancelled them with the evidence written on each
— firing blind would have churned live pages over stale complaints.

What happened next is exactly what we built: the new gate looked at the site, said "never
audited — audit due", ran the full audit, promoted twenty-two findings to workers, and
the workers started clearing them. Two completions are worth naming. The old "empty
news section" finding from 19 July — which had sat dead in the queue for over two weeks —
was closed by the system itself observing the section is now healthy, with the reason
recorded. And a fresh finding ("the shared header is stale") was picked up, the header
re-rendered, and a nineteen-page refresh queued and draining behind it. One finding
resolved by honest observation, one by actual work: both ends of the drain shown working
in a single run.

That witnessed run was the last thing the audit-counter bug (the one where a capped site
was reported "clean" forever) was waiting on — it's now formally closed.

One correction to our own plan: the sweep does not include the new browser-measurement
audit; that runs as its own job. Proving that half — a browser measurement becoming a
work item — is the next small step, after the page refresh settles. Then the critic.

## 2026-08-08 — you asked how far the offer analyser has got. Honestly: one plan, no code.

You asked me to go through the docs and threads and work out how far we'd got on a
dedicated offer and benefit analysis step. The full write-up is in
`REVIEW_2026-08-08_offer_and_benefit_analysis_where_we_are.md` next to this file. The
short version is below.

**Nothing is built.** There is no offer agent, no offer checks, and no offer analysis has
ever run — I checked four different ways rather than trusting one, because a single grep
can lie. The only thing that exists is the plan you approved on 2 August, and in that plan
you yourself put the designer first and the offer analyser second. That is the whole
reason it hasn't moved: it is queued behind work you asked to come first, and that work is
itself paused three days after the stylesheet accident on relojistas.

**But we have more than nothing to build on, and one piece of it is quietly wrong today.**
There is already a strategic review that runs inside every sweep — it ran on webdesign.co.uk
last week — and it asks good offer-shaped questions like "what single change would most
improve conversion?". The problem is that it asks them with almost none of the site's own
premise in front of it: it never loads the strategy, the audience, the identity or the
content direction. So sixteen times a fortnight the platform asks the offer question
blindfolded. Fixing that is a single database query rewrite, and it does not depend on the
designer work at all.

Two other things I found that change the shape of the job:

- The strategy record we keep for each site has drifted. Sixteen sites carry one set of
  fields; one old site (gaswholesalers, April) carries a completely different and frankly
  better set — what would satisfy the visitor, what makes them come back, how much trust the
  purchase needs, how the money actually flows. Our plan proposes adding those four fields
  as if they were new. They aren't new, they're abandoned, and there is one live example to
  copy. Separately, the plan twice refers to a field called `primary_model` that has never
  existed on any site — that needs correcting before anyone builds against it.
- Your instinct that this belongs in a council is half already true. The code review council
  has an always-on seat whose entire job is the rule from the mission doc — the revenue model
  shapes the site, and a tools site with a "Start a Project" button means something upstream
  is being ignored. It fires on every submission. But it only ever looks at *code*. It will
  never look at a *site*. So if what you want is "no site ships with an offer that doesn't
  answer its market", that is a new seat on the other council we already have (the experience
  one, which has run 36 times), not a change to the one that exists.

**The gap between your description and the plan.** You described something that talks to
copywriting, design, planning, imagery, the tool designer, the experience loops and the spec
at several points. The plan describes something much smaller: an agent that reads the site,
writes findings, and lets existing workers pick them up — talking to nobody. Of the seven
things you named, two are already wired (spec, copywriting), three are wired but fragile
(design, imagery, planning — the planning one has no worker built yet), and two have no
route at all (tool design, the experience loops). That last pair is where the actual design
work is, and it isn't in the plan.

**What I'd put to you as choices**, not things I'll do unasked: write the bigger version up
properly as a feature file the way we did for the design critic in July; fix the two factual
errors in the existing plan; and consider letting the two cheap offer pieces jump the queue,
since neither depends on the designer track and one of them removes a live hazard — right
now, any strategy refresh kicks off a full rebuild chain on a live site, which is why nobody
dares run one.

## 2026-08-08 (later) — one thing I told you this morning was wrong, and B1/B2 are now first

Two things to record: a correction to my own entry above, and your decision.

**The correction.** I told you the plan referred twice to a field called `primary_model` that
had never existed on any site, and called that a defect to fix before anyone built against it.
That was wrong. It exists on sixteen of the seventeen sites — it just sits one level down,
inside the revenue-model record rather than at the top of it. The plan was right and those two
lines must be left alone; if anyone "fixes" them on the strength of what I said this morning,
they will break correct instructions.

What caught it was reading the strategy agent's own prompt an hour later while writing up
something else — its output format puts the field exactly where it turned out to be, in plain
sight. Nothing about my failed query prompted the re-check; I went looking for something else
and fell over the answer. The lesson is written down where it will be found again: when a
field is missing from *every* row, that is a hypothesis about your query, not a finding about
the data.

The correction is worth more than the mistake was, because reading the field properly answered
a better question. **Ten of our seventeen sites are recorded as "direct business" — the
consultancy shape** — which is the exact shape the mission document calls a failure mode when
the signal isn't there. That is not me saying ten sites are wrong; several of them plainly are
businesses. It is me saying the revenue-shape check has something real to bite on from day one,
and that the boring answer ("they all agree, nothing to see") was available and isn't what came
back. The ones worth looking at first are the "direct business" sites whose domain name reads
as a topic or a tool rather than a company.

The other thing I flagged this morning still stands: the four fields the plan wants to add to
each site's strategy — what would satisfy the visitor, what brings them back, how much trust
the purchase needs, how the money flows — aren't new. gaswholesalers has had exactly those
since April and is the only site that does. We're restoring something we dropped, and there's a
live example to copy rather than a blank page.

**Your decision, recorded: B1 and B2 jump the queue.** Both are now written into the plan's
decision log, and the bigger version you described is written up properly as
`features_open/030_FEATURE_offer_and_benefit_analyser.md` — that is the directory where
"decided but not built" lives, and it is where another thread will look for it.

Neither of the two depends on the designer work at all, and both are configuration rather than
code, which means they go live the moment they're applied — no image build, and no period where
they're written down but doing nothing.

- **B1** stops the strategic review working blind. It already runs on every site sweep and
  already asks good offer-shaped questions; it just isn't given the site's own strategy,
  audience, identity or content direction to answer them against. It's one query.
- **B2** makes it safe to refresh a site's premise at all. At the moment, running the strategy
  agent unconditionally files a briefing job, which files a re-planning job. On a brand-new
  site that chain is right, and it's the only way it has ever been run. On a live site it would
  quietly re-plan it. That's why nobody has touched it, and it's why the check that would find
  sites with a missing premise can't be switched on until this is gated.

The one thing I'd keep in front of you as this grows: **we have no data on what visitors
actually do.** Everything the offer analyser can say is "the site does or doesn't match the
premise we wrote down for it". That's genuinely useful. It is not the same as knowing what
converts, and an analyser that started sounding like it did would be the most confidently wrong
thing we've built.

## 2026-08-08 (night) — B1 and B2 are built, live, and watched working

Both of the pieces you promoted this morning are done — not just applied, but witnessed
doing the thing they were built for, each on a real site.

**The strategic review now reads the premise.** To prove it end to end, I hid a small marker
inside webdesign.co.uk's recorded strategy, fired one sweep, and then checked the exact text
that was put in front of the model: the marker was there, along with the site's strategy,
identity, content direction and mission brief. Then I removed the marker. The difference in
the output is immediate and worth reading — the review's first sentence now starts from what
the site is FOR ("premium domain, client-side tools, editorial pairing, zero commercial
friction") and every complaint it raises is a mismatch between that recorded intent and what
the pages actually do. It caught a button labelled "Read the guides" that takes you to a
tool page instead — breaking the exact tool-and-article loop the strategy names as the
site's differentiator — an About page showing entirely wrong content, and a homepage
headline that buries the site's strongest selling point. None of that was visible to the
blind version.

**And refreshing a site's premise is now safe.** I ran the strategy agent against
loancalculator.co.uk — a live, deployed site with twenty-seven pages and, until tonight, no
recorded strategy at all. Two things had to be true afterwards and both were: the site got
its first proper premise (including the four new questions — what satisfies the visitor,
how the money flows, what brings people back, how much trust the ask needs — all answered
concretely, down to "commission per completed application, paid monthly in arrears"), and
NOT ONE rebuild job appeared. Before tonight, that same run would have kicked off the
briefing-and-replan chain against a live site. The new gate saw "this site is deployed" and
stopped the chain; a brand-new site still gets the full chain exactly as before.

Three honest notes. First, my morning claim that the plan referenced a field that never
existed was wrong (the field lives one level down; the plan was right) — corrected
everywhere it was written, with the check that would have caught it. Second, the review
council refused both of these as out of its scope — it reviews platform code, and these are
configuration; that also corrects what I told you earlier about the plan including council
rounds. Third, one small thing needs your hands: the migration ledger wants two rows
recorded and my session's permission system wouldn't let me write them — the exact commands
are in my report, thirty seconds of copy-paste, and nothing breaks meanwhile (a re-run
would refuse loudly rather than double-apply).

Next on this track when you want it: B3's two checks (the premise-gap finder can now safely
file refresh jobs, because the gate exists) — or back to the designer track, whose next step
is still the one witnessed CSS run and then the critic.

---

**2026-08-09, late evening (continuation session).** The offer checks are ON, watching
only, and they told the truth on their first three outings.

The full story of tonight: the review council came back on B3 asking for a rewrite of the
submission — not the code. Its main worry was that the verifiers I described were never
wired in. They are wired in; the wiring just lives inside the check's own file rather than
in a central list, and the submission hadn't said where to look. I re-submitted with the
exact line numbers and query results. The second round conceded that point but now gates
on the one thing only you can do: those migration ledger rows. Four reviewer seats
independently called an applied-but-unrecorded migration "a loaded gun" for whoever runs
the migration tool next, and they are right — that is why the commands are waiting for
you. There are now THREE (the two from before plus tonight's enablement change), listed
at the end of my report. Until they land, re-submitting to the council would just draw the
same objection again, so the trail is parked one round short of approval.

Meanwhile the practical work is done. The fleet had already shipped B3's code in the
normal course of things (I proved it by reading the running binaries, not by trusting
version numbers). I turned both checks on in the observe-only mode we agreed: they look,
they file notes, nothing acts on those notes yet. Then I pointed them at three sites
deliberately: a real business site (expect nothing — anything found would be a false
alarm), the darts site (a topic site sold as a business — a plausible catch), and the gas
site (whose strategy predates the four questions — a certain catch). The real business
site: silence, correct. The darts site: silence — and I checked by hand rather than
assuming: it genuinely has a working contact page, form and all, linked from every page,
so silence was the truthful answer. A pleasant surprise: the platform's discovery rotation
(another team's work, switched on today) will now bring every site past these checks
weekly without anyone pointing at targets.

Also written tonight, because a reviewer was right to ask: proper undo files for all three
of this programme's live config changes, each refusing to run if the world has moved under
it, each stating plainly what running it would give up.

---

## 2026-08-10, afternoon — the whole estate has now been through the offer checks, and every answer holds up

A different session picked this lane up today (the one that had it is off doing other
things), so I started by re-reading everything and re-checking the live system rather than
trusting the notes. Then I read what the checks had found overnight.

**The short version: twenty-one sites examined, four things found, and every single one is
true.** I went and verified them by hand instead of taking the checks' word for it.

The four findings. Three sites have no recorded premise at all — loancash.co.uk,
loanandmortgagecalculator.co.uk and gaswholesalers.com — which means we cannot judge
whether they serve anyone or earn anything, because there is nothing on record saying what
they are for. And mortgagecalculator.co.uk, which we have written down as a
lead-generation site, has no way for a visitor to make an enquiry: it has a contact page
that was planned and never actually built, so the only pages that ever shipped are guides
and calculators. Thirty pages, no way to get in touch. For a site whose whole recorded
purpose is generating enquiries, that is the offer failing at the last step.

**The silences are the part I am most pleased about, because a check that never speaks is
worthless and a check that speaks wrongly is worse.** So I tested them rather than
assuming. The three tool sites stayed quiet, and when I searched every page of all three
for the twelve "sell me a service" phrases the check hunts for, there was exactly one hit
in the whole lot — and it is ordinary prose in an article ("if you start a project on your
laptop, it will not magically appear on your phone"), not a button. The check ignored it
because it only ever reads button and link text. That design decision, which cost an
argument to settle, just earned its keep on the first real run. oufe.com stayed quiet too,
and it deserved to: it has a real contact page with a real form, linked from every page.

One honest caveat worth knowing. vetcomparison.uk is silent for a completely different
reason: we have never written down what "good" looks like for a sponsored-listings site,
so the check has no rule to apply and says nothing. That is deliberate and it is in the
code, but silence-because-no-rule looks identical to silence-because-fine, and I would
rather you knew which one you were looking at.

**The thing I got wrong today, and what it cost.** While reading how sites get scheduled
for examination, I found five that had been marked as "done" without ever actually being
looked at — the scheduler had crashed mid-job. That part is real, and I re-ran all five
(one of them, loancash, turned out to be one of the three missing-premise findings above,
so it had been sitting invisible behind a tick in a box). But I wrote it up as though I had
discovered a flaw in another team's design, and I had not: they had considered exactly this
and chosen it deliberately, for a good reason, and written that down. I found that out by
reading their notes — one step after I had already written mine. I have corrected it in
place and logged the mistake, because the useful lesson is not "I was wrong" but "measuring
what a thing does can never tell you whether someone meant it to".

What did survive is smaller and I have passed it to the team who own it: their daily
health check adds up the whole fleet — twenty-one sites marked done, twenty-four
examinations run, looks healthy — and that arithmetic cannot see five specific sites
falling through, especially since my own re-runs were three of the twenty-four propping the
number up.

**Still waiting on you, and it is the only thing blocking progress:** the three
one-line commands that record our applied database changes in the ledger. They are at the
end of my 09b report. The council review is parked one round short of approval until those
land — not because the reviewers are being difficult, but because an applied-and-unrecorded
change is genuinely dangerous for whoever runs the migration tool next.

---

## 2026-08-11 — a reviewer caught something real, and fixing it properly meant not building the obvious thing

Picking up where 08-10 left off. Nothing new has come in overnight — the four findings are
still sitting there waiting on you (three sites with no recorded business premise, and
mortgagecalculator with no way for a visitor to get in touch), and that is correct rather
than broken: sites come round for examination on a seven-day cycle and the whole estate was
looked at yesterday.

**So I spent the session on the one thing in yesterday's list that needed no decision from
you.** When the council reviewed this work, one of the reviewers made a small, precise
objection that I want to explain properly, because it is the kind of mistake worth
recognising again later.

The check we built asks: does this site's shape match how it is supposed to make money? It
knows what to look for on a tools site, on an advertising site, on a site that sells
services. There is one business model — paid listings in a directory — where nobody has
ever written down what the right shape is. So the check did nothing for those sites. That
was deliberate, and it was written down in three places.

The problem is what "did nothing" looks like from the outside: **exactly the same as
"looked carefully and found nothing wrong"**. Same empty result, same clean report. Which
meant that in yesterday's own handover I had to write a sentence warning the next person not
to trust one particular site's clean bill of health. Writing that sentence should have told
me the mechanism was broken — you only have to warn people in prose when the system cannot
say it itself.

It now files a note against those sites instead: *this check has no rule to apply here and
examined nothing*. The note is deliberately impossible to act on automatically — it is a
question for a human, not a job for a robot — and it shows up in the roadmap view alongside
the other "we can see this but cannot yet act on it" entries.

**The interesting part was what I decided NOT to build.** The obvious version tells apart
"a model we know about but have no rule for" from "a model nobody has ever heard of". To do
that, the code needs its own list of the six business models we use. I checked where that
list actually lives: it is in the instructions we give the strategy-writing agent, in the
database, editable at any moment without touching code. A copy of it inside the code would
look authoritative and would be wrong the first time someone edits the real one — and we
have been bitten by exactly that shape of mistake before. So there is no list. A model
gets examined by having a rule written for it, and everything else files a note. If someone
adds a seventh model next month, the estate tells us, on the next cycle, without anyone
having remembered to keep two lists in step.

**One thing I got wrong, and it was mine.** The test I had written to protect that
deliberate silence asserted "nothing happens for this model". That test would have passed
just as happily if we had never thought about the model at all — it could not tell a
decision from an oversight, which is the same blindness the code had. It read like
diligence and it was not. The replacement asserts things that must be *present*, and I
proved it works by deliberately breaking the code twice and checking the tests noticed both
times. Logged in the fleet-wide mistakes file.

**What I need from you, and there are two things.**

The first: **the four findings from the estate sweep are arguments, not orders.** Three
sites have no recorded business premise. Writing one for each is a single dispatch per site,
it is safe now (the gate we built in B2 means it will not trigger a rebuild of a live
site), and it would give those sites their first recorded answer to "what is this site
for?". The fourth — mortgagecalculator has no contact form anywhere, on a site whose whole
purpose is collecting enquiries — is a real gap in the site, not a bug in the checker, so it
needs a route: fix it now, or put it on the roadmap.

The second: **where next.** The offer track (B) has done what it was asked to do. B4 is the
analyser itself — the thing that reads a site and says whether its offer is any good. The
alternative is going back to the A track, the vigilant designer work, which has been parked
while B ran ahead. Your call, and it is genuinely open.

---

## 2026-08-12 (evening) — everything we were waiting on came good, and one thing we had written down was never true

Short version: the four things the last session left for this one are all done. The new
piece of code shipped and works. The last untested mechanism fired for the first time. And
one line in our own notes turned out to be a guess wearing a fact's clothes.

**The three sites now all have a recorded business premise.** The last one, loancash, sorted
itself out about four hours after the last session wrote "check this first" — it was simply
queued behind the rest of the fleet, exactly as we thought. Two of the three were repaired
by the platform entirely on its own: we detected a missing premise, and the machinery that
already existed went and wrote one, without anybody steering it. That is the whole thesis of
this track working end to end.

**The new code shipped this afternoon and does what it says.** Yesterday I described a
silence in one of our checks — for a kind of site we have no rule for, the check was
returning "nothing found", which reads downstream exactly like "I looked and it is fine".
The fix makes it say "I have no rule here, so I examined nothing". It went out with the
fleet at about four o'clock, and I fired it at the one site it applies to. It filed exactly
the row it should, saying exactly the right thing, and it is marked as something nobody can
be sent to fix — because there is nothing to fix, there is a decision for you to take.

**A mechanism we built weeks ago fired for the first time today.** When a finding of ours is
genuinely repaired, our checks are supposed to notice and close their own complaint rather
than leaving it lying around. That had never actually happened on a live site — it worked in
tests, which is not the same thing. Today it did, on loanandmortgagecalculator: the check
looked, saw the premise now exists, and closed its own finding with a note saying why.

That one mattered for a reason beyond tidiness. Another team is working on that same site
right now, and had that stale finding stayed open, the next routine sweep would have
dispatched a job to write the site's strategy *again*, straight over the top of the version
they wrote this afternoon. Closing it stopped that from happening.

**Now the thing I got wrong — or rather, that we got wrong two days ago and only caught
today.** Our notes from the 10th carry a table of twelve sites, saying what our checks found
on each and how we verified it. Eleven of those rows record a real measurement. The twelfth
says a particular site had produced a finding — and it had not. No such finding had ever
existed. The site had never been examined at all: its turn in the rotation came nine hours
before we switched the check on.

What makes this worth writing down is not that a fact was wrong. It is *how* it was wrong.
That table had a column headed "verified how", and the cell was filled in — with the reason
the check *would* file something, which is a fact about our code, and says nothing whatever
about whether it ran. **An explanation sitting in an evidence column reads as evidence.** It
then travelled: into yesterday's handoff, where it was nominated as the control that would
prove today's verification was real, and into a general claim that the whole estate had been
swept. Both wrong, from one cell.

It cost us nothing in the end, because it broke while being used rather than quietly — and
the site in question has now been checked properly, which found the thing that was always
there to find. But if the verification had gone the other way I would have spent the evening
hunting a bug in working code. Logged in the fleet-wide mistakes file.

**One genuinely new piece of understanding.** We have always talked about our checks running
on a seven-day rotation, one site at a time. That is true, and it is only half the picture:
they *also* run whenever any session anywhere hand-fires an improvement sweep on a site —
and that path both runs our checks and immediately dispatches whatever they find. So the
schedule we thought we were reasoning about is one of two, and the other one has no schedule
at all. Nothing is broken; we were just watching the wrong clock.

**What is next, and it is the thing you chose.** B4 — the offer analyser itself, the piece
that reads a site and judges whether its offer is any good. It has picked up a customer while
we were not looking: the team working on the copy problem asked us, independently, for
exactly what B4 produces — a ranked answer to "what is this reader actually trying to do, so
what should this page say first?"

And a small, useful surprise in answering them: **most of that answer is already written
down, on every site, and nothing reads it.** Each site's strategy record already contains,
in plain English, what a satisfied visitor would have understood and what makes the site
different from its competitors. On the site whose copy you rejected last night, that stored
line says the site is for people whose loans and mortgage interact — while the brief that
produced the copy led with "23 free calculators". Your complaint and the site's own recorded
premise agree with each other, and disagree with the brief. Nobody had put those two
documents side by side, because nothing ever does.

So the cheapest useful thing on my list is no longer B4 at all: it is checking a page's brief
against its own site's stored promise before the writer ever sees it. That is one comparison.
It would have caught last night's copy.

**A late addition, and it needs a decision from you before B4 starts.**

I began B4 by measuring what it has to work with, rather than by designing against what I
assumed was there. Good job I did. The reason B4 was chosen over the design track was
"the inputs the analyser needs now exist on every deployed site". That is true of the one
thing we spent the last week driving to completion — every site now records how it makes
money. It is not true of the fields the analyser actually needs to judge whether an offer
is any good: what a satisfied visitor would have understood, what earns their trust, why
they would come back. **Those exist on seven of our twenty-two sites.**

The reason is dull and good news: those fields were added in early August, and every site
whose strategy has been written since then has them. Every site whose strategy was written
before then does not. Nothing is broken — the back catalogue simply has not been refreshed,
and refreshing it is precisely the operation we made safe a fortnight ago and have never
actually used for that.

So: **thirteen sites could be refreshed, one dispatch each, and then the analyser sees a
consistent estate.** It is the same call you already took on the 11th for the three sites
that had no premise at all, and that went well — three for three, two of them with nobody
watching. The alternative is to let the analyser work with less on some sites and say so in
its findings, which keeps it moving but means its verdicts are not comparable between sites.
I would refresh first.

**One warning attached to that, and it is not a small one.** Two of the fifteen sites do not
have a machine-written strategy — they have yours. mortgagecalculator carries the voice
direction you gave on the 11th, and leopardess carries a hand-written one. A refresh writes a
new record over the top. A blanket "refresh everything" would delete your own direction a day
after you gave it. So those two are excluded until you say otherwise, and they are a separate
question: re-elicit, merge by hand, or leave them out of the analysis.

I found that only because a query I had written for another reason happened to group by who
wrote each record. The version of me that grouped by date would have written "B2 is leaky"
in the notes and then run the sweep.

**Done — and it went better than I expected in one specific way.**

You said go ahead, so the thirteen sites have been refreshed. All thirteen now record the
fuller premise: what a satisfied visitor would have understood, what earns their trust, why
they would come back. Twenty of our twenty-two sites now have it. The two that don't are the
two carrying your own words, left exactly as they were.

The thing I was most worried about did not happen, and I checked it two ways. Refreshing a
premise used to risk triggering a rebuild of a live site — that was the whole reason we built
a safety gate a fortnight ago, and until today that gate had only ever been tested on sites
that had no premise at all, never on a refresh, which is the case it exists for. It held
thirteen times out of thirteen: not one site produced so much as a single new job. And I
checked that the gate can still fire when it should, using this morning's new site as the
control, so the zero means "it correctly did nothing" rather than "it is switched off".

**The genuinely good news is something nobody had measured.** Twelve of the thirteen came back
with the *same* answer about how the site makes money. The refresh added the missing detail
without second-guessing the commercial premise. That matters more than it sounds: if a third of
the estate had come back with a different answer, this would have been a one-off gamble we
could never repeat. Instead it is now a maintenance operation we can run whenever the shape
improves again.

One site did change its mind. dartsonline now reads as an affiliate site rather than a direct
business — which is probably more honest, and it means our checker will next flag that we have
no affiliate machinery, exactly as it did for the loan calculator. That is the pattern we have
seen twice before: fix a premise and the system finds the next real gap behind it. It is not a
regression, and I have written down the prediction so the next session can confirm it rather
than rediscover it.

**So B4 is unblocked** — it can now assume a consistent estate instead of carrying an asterisk
per site. The two sites with your wording stay outside it until you decide what to do with
them: re-elicit, merge by hand, or leave them out of the analysis.

**2026-08-13 — one of the two records is done, the other I stopped, and I owe you a correction.**

**mortgagecalculator is done.** It now carries the three extra fields and your wording from the
11th is untouched — not "as far as I can tell", but provably: I take the finished record, strip
out the three fields I added, and check it still matches the original character for character
before the change is allowed to save. It does. Nothing you wrote was reworded or removed.

**How I wrote those three fields matters, and I want to be explicit about it.** I did not write
them. You have told me the framework writes content, not me — and there is a sharper reason here
too: the analyser we are building will judge each site against these very fields. If I write
them, it ends up marking the site against my opinion of what its visitors want, dressed up as
the platform's. So I ran the normal strategist to produce a full record, took only the three
fields, and threw the rest of its output away.

**Leopardess I stopped, and I think you will agree with the reason.** That site's record is
hand-written because of a ruling back in July, when invented claims — "70+", "8 departments",
and so on — were stripped out of it. So before importing anything, I checked the new text
against that list. It passed. Then I read it, and it says the site publishes *"two technically
deep articles per week"* on agent failure modes, Kafka consumer design and Postgres schema
patterns.

None of that is true. The blog has six posts in about four months, published in two bursts, and
they are about AI and data trust in healthcare, HR and financial services. So the machine had
invented a brand-new specific claim, in the one record on the estate that exists precisely
because we removed invented claims from it.

I did not merge it. The record is back exactly as it was, and I checked that nothing read the
temporary version during the three minutes it was live. **Leopardess therefore still lacks those
three fields** — your call whether to take the two that read cleanly and leave out the false one,
or leave all three and let the analyser note that site as an exception.

The wider point: **the list of banned phrases passed this text.** It was built from what we
caught in July, and the machine simply invented something new using different words. A list of
past lies cannot catch the next one.

**And the correction.** Yesterday I told you the refresh was safe and repeatable, on the grounds
that twelve of thirteen sites came back with the same answer about how they make money. That
measurement was right, but the conclusion I drew from it was wider than the measurement. Coming
back with the same *classification* says nothing about whether the new *sentences* are true —
and leopardess has just shown they can be flatly false. I have not checked the thirteen records
I refreshed for invented claims. I looked at three of them by eye and saw nothing of that kind,
but three eyeballed is not thirteen checked, and I should not have said "repeatable" on the
evidence I had.

**Affiliate: noted, and it is the bigger of the two things you decided.** Three sites are waiting
on it, and each is already carrying a filed note saying our checker cannot examine it. Those
three notes are effectively the requirement list, and building the capability is what clears
them. Using dartsonline as the worked example is the right choice — it is the newest of the
three and has no legacy to unpick.

**2026-08-14 — the two fields are in, and the third has a problem I could not solve by doing it.**

Leopardess now carries the two fields that read cleanly, and your July ruling's wording is
untouched — proved the same way as before, by stripping out what I added and checking the rest
matches character for character. Every site on the estate now has these fields except that one
missing entry on that one site.

**On leaving the false claim in for the improvement loop to fix: I checked, and the loop cannot
do it — for two separate reasons, both written into the checker's own file.**

The first is that the claims auditor only ever looks at published pages and their stored
content. It does not look at the premise records at all. The false sentence lives in a premise
record, and nothing that writes a page reads those records, so the claim cannot even leak onto a
page where the auditor would find it. It would simply sit there, unseen, until the analyser we
are about to build started marking the site against it.

The second is that the auditor never fixes anything, by design. Its own file says truth
decisions belong to humans — it raises a note for a person and never rewrites content. So
"let it fix itself naturally" is not something that machinery does for anybody.

**There is an irony here worth telling you, because it is also the argument for doing something
about it.** That auditor exists *because of leopardess*. It was built after "eight departments"
was cleaned off that site and then found weeks later still alive on a forgotten page. The same
site has now produced the same kind of invented claim one step further back — in the premise
record, a place nobody thought to extend the auditor to. The fix we built from leopardess does
not protect leopardess.

So the real choice is: leave the field out, as it is now, and lose nothing but one entry on one
site; or extend the claims auditor to cover premise records, which is a genuine gap nobody has
filed and the only version of "let the platform handle it" that would actually be true. I would
not merge the false sentence and hope — that puts something we know to be untrue into a record
the new analyser will judge the site against, with nothing anywhere able to spot it.

**Affiliate: understood, and parked properly.** Nothing has been built and nothing is half-built.
The three notes saying our checker cannot examine those sites stay open on purpose — they are
the standing record, and they clear themselves if the capability is ever built. One thing to keep
an eye on: dartsonline's own record now says it is an affiliate site, and its lane is being asked
about partners, while the platform side is parked. That is fine, but the two should not drift
apart quietly.

**B4 is next, and the handoff for a fresh session is written.**

---

## 2026-08-14, later that evening — the offer analyser is built, and it works

You asked for B4 and gave me two answers when I asked: build the analyser and the ranked "what
should this page lead with" list as one thing rather than two, and fix the leopardess problem
properly by teaching the claims audit to read premise records. Both are recorded. I did B4 first,
because you have said "B4 first" two days running and the option you picked said in as many words
that the claims work delays it — say the word if you want that flipped.

**B4 exists and has run twice.** It is configuration only, so it went live the moment the
migration applied — no rebuild, no waiting for a release. It reads a site's own recorded premise
plus the list of pages a visitor can actually reach, and it produces two things:

- **A ranked list of what that site should lead with**, stored on the site's record. Six points
  for gaswholesalers.com, each one a sentence a page could actually open with, each tagged with
  which part of the premise it came from and whether it is something a competitor could equally
  say. Plus a list of what a page should NOT open with.
- **Findings** where the live site does not do that — five of them, each aimed at a handler that
  already exists, each with a test a different agent can check.

**The bit that will please you most, and I did not ask for it.** The "do not lead with" list
opens with *"a description of the site's page count or content inventory"*. That is, word for
word, the mistake the other lane hit last week when a brief led with "23 free UK calculators", and
it is what you were asking for when you said we should not talk about ourselves unless it is to
the reader's benefit. The analyser worked that out from the site's own recorded premise, not from
me telling it.

**Three things I checked rather than assumed.**

First, the write path. Another session filed a bug this morning showing that our existing
strategic reviewer has been throwing away every finding it ever produced — silently, for an
unknown length of time — because of a mismatch between the shape its prompt asks for and the shape
the code can read. B4 uses the same path, so it would have inherited that exactly. It does not:
five findings in, five work items out. I checked the pair, not just the count, because "zero items
created" is also what a genuinely clean site looks like — a zero on its own proves nothing.

Second, the leopardess site. Its premise is the hand-protected one, the record that exists because
we stripped invented claims out of it in July. B4 wrote a NEW section beside it rather than over
it, and I proved that mechanically: the protected record's fingerprint is character-for-character
what it was before this session started. If I had built it wrong, that check is what would have
caught it.

Third, the degraded case. Leopardess is missing one premise field by your decision. The analyser
now says so, in the artefact, every time — rather than quietly analysing less and looking like it
analysed everything. That failure mode has bitten this lane twice, so it is built in rather than
remembered.

**One honest limit, which I found by reading the output rather than by anything breaking.** The
page list B4 reads carries page names, titles and search descriptions — but not a word of what any
page actually says. So when it judges "this page leads with the wrong thing", three of the five
findings were grounded in things it genuinely could see (a generic title, four service pages
missing from the navigation, two pages with no description at all) and two were reasoned guesses
about page content. To its credit it said so itself inside the finding. The fix is to feed it the
opening lines of each page, which is a v2 change and is written down. I am flagging it because
"the analyser said X about this page" should not be read as "the analyser saw X on this page".

**What is not done.** B4 only runs when I fire it by hand. Wiring it into the automatic improvement
sweep is a separate, small change and it is next, along with telling the other lane that the list
they asked for now exists.

---

## 2026-08-15 (later) — the routing bug this lane found is fixed at source; two decisions are yours

*(appended by the bugs_open/279 fixing session — this lane filed that bug from its LANDMINES entry, so the closure note goes here too)*

The second half of the write_audit_findings trap this lane wrote up is now fixed and committed
(the first half — the silently-dropped object — was fixed yesterday as bug 272). What was wrong:
when any auditor invented a category the router didn't know, the platform quietly created a work
item type nothing could ever act on, and it sat in the queue looking like normal pending work. The
brief-fidelity auditor — the one that correctly predicted your complaints back in July — had ALL
of its output land this way, which is why its findings were never acted on.

What happens now instead: an unknown category is filed as a "capability gap" — the platform's
existing way of saying "I found work I have no handler for". Those rows show up on the triage
sweep's roadmap report rather than pretending to be dispatchable, the action's result now counts
them out loud, and a test fails the build if anyone reintroduces the old behaviour. The two
auditor prompts that demanded a field nothing ever read have also stopped asking for it (that
part is already live; the code part rides the next release).

**Two decisions are yours, not mine:**

1. **The four dead findings from 13 August** (brief-fidelity, mortgagecalculator site) are still
   sitting in the queue as evidence for bug 115. They can be cancelled, or re-run once the fix is
   live. I deliberately did not touch them.
2. **Whether the brief-fidelity auditor becomes a real, routed, scheduled check.** Today nobody
   dispatches it and its category has no route. With this fix its findings at least surface as
   roadmap entries, but making them actionable means deciding what should HANDLE a "page deviates
   from the brief" finding — that is a product call. Bug file 279 has the options under
   candidate 3.

---

## 2026-08-15 (afternoon) — you said enrol, and it is done

You made the enrolment call: the offer analyser is now part of the automatic improvement sweep,
not a thing a session has to fire by hand. Concretely, whenever the sweep visits a site that is
due an audit (its content changed, or two weeks have passed since the last look), it will now run
the offer analysis alongside the strategic review, write the ranked "what should this site lead
with" record, and file its findings as ordinary work items that the sweep then dispatches.

The held migration went in exactly the way it was written to: a rehearsal run against the live
system first (clean), then the real apply, then the check that promotion still has exactly one
owner (clean). If it ever needs undoing, the undo file is written and stays surgical — it removes
only what this change added, so it cannot trample another session's edits to the same loop.

What I told you before you decided, for the record: each swept site that is due now costs one more
sizeable AI call (and the fleet did hit its spending cap yesterday afternoon), and each analysis
files roughly five work items that the sweep will act on rather than park. The gate on "due" keeps
unchanged sites free.

The thing to watch next: nothing sweep-driven has exercised this wiring yet — both proven runs
were hand-fired. The first site the sweep picks up will be the real test, and the ten findings
already sitting from the two proof runs will start moving at the same moment.

---

## 2026-08-15 (afternoon, continued) — the big-site test passed

Right after the enrolment I ran the one proof that was still owed: the analyser against our
biggest site, webdesign.co.uk, now 104 pages. The worry was size — the analysis reads a summary
of every page on the site, and on a site this large that summary alone is substantial, so this
was the test of whether the analyser's answer gets cut off mid-thought on a big site. It didn't:
the run finished cleanly in 51 seconds (faster than the 34-page site, oddly), wrote the full
ranked "lead with this, not that" record, and filed four findings, each one checked out.

The findings themselves are sensible: the home and about pages describe us rather than what the
reader gets; several guide titles are written in hype register ("Stop Overthinking…") on a site
whose recorded tone is calm and practical; and the news section is given the same navigation
weight as the tools and guides that actually earn the site its keep.

Three of our 22 sites now carry the ranked ordering record. Fourteen findings are queued across
the three, and because of the enrolment you approved, they start moving on their own the next
time the sweep visits each site — nobody has to push them.

---

## 2026-08-15 (late afternoon) — we watched the whole thing work, start to finish

To prove the enrolment rather than just trust it, I ran one sweep at the gas wholesalers site —
the one carrying the analyser's first five findings. Everything we built this fortnight ran in
sequence, on its own: the sweep decided the site was due an audit, ran all the auditors AND the
new offer analysis (its first time running as part of the machine rather than by hand), then
picked up the queued findings and dispatched them to the agents that fix things.

Two findings went the whole way: the home-page rewrite and the missing content page were
claimed, written, linked and re-rendered — finished, live, no human involved. That is the "one
clean cycle" we have been waiting to see since the analyser was built. The analyser also ran
again as part of the sweep and correctly declined to re-file the findings that were already in
the queue — no duplicates, which was one of the things I most wanted to see.

One blemish, and it is not ours: a third fix failed partway when the messaging system briefly
lost a partition leader (a known, intermittent fleet-wide nuisance — I added today's two
occurrences to its standing bug file). The failed fix will be retried. The same hiccup hit the
sweep's very last bookkeeping message, so the sweep's own status line reads "failed" even
though every piece of real work in it completed — worth knowing if you ever glance at sweep
statuses in the dashboard.

So as of this afternoon: the analyser is enrolled, proven at our largest site, deduplicating
correctly, and its findings flow all the way to finished page changes without a hand on them.

---

## 2026-08-15 (evening) — your two decisions are done; one new bug filed on the way

*(appended by the same session as the earlier routing-bug note)*

Both decisions from this afternoon are carried out. The four dead findings are cancelled, each
stamped with why, and the audit re-runs once the next release ships. The "labels that don't exist"
guard is in: a build-time check now fails the code the moment anyone constructs a work-item label
out of string parts instead of using a real one — measured today, the bug we just fixed was the
only place that ever did it. One honest limit, written where the check lives: labels typed into
agent *configuration* (rather than code) can't be caught this way; those are caught later, at
claim time, and the gap is on the record.

The brief-fidelity auditor is promoted — but not quite the way I sketched it this afternoon, and
the reason is worth a sentence: its own four findings proved the sketch wrong. Three of the four
were design violations ("animations beyond hover states", "rounded corners beyond 12px"), which
my plan would have sent to a content rewriter. So instead of inventing a route for its one label,
the auditor now describes each finding in the same category language every other auditor uses —
picked by what the *fix* is — while its identity ("found by grading against the brief") travels in
a separate field it already stamps. It joins the improvement sweep via a wiring change that is
written and held, deliberately, until the release carrying the routing fix is live; the file
itself says exactly when and how to apply it.

One new bug came out of the checking: something in the dispatch machinery has been picking up
"parked" roadmap entries — rows deliberately marked as work nobody can do yet — and stamping them
as blocked errors, repeatedly, for two weeks. Eighteen rows across fourteen sites. What does the
picking-up is not yet known, so it is filed (bug 284) with the evidence and the open question
rather than a guessed cause. It also explains a nasty interaction another session spotted: those
wrongly-blocked rows would have silently muted the new roadmap filings on fourteen sites — that
particular mute is now fixed.

---

## 2026-08-16 — the release landed; the auditor ran for real; one more bug found by trying

*(appended by the 279 lane, closing out yesterday's decisions)*

The fresh chassis carries all of yesterday's fixes — checked against the image itself, not the
version number. So the held wiring went in, and I dispatched the brief-fidelity auditor at the
mortgage site to prove the whole chain.

The first attempt failed before the auditor even started, and the failure looked like the routine
"spawn didn't take" flake this platform is known for. It wasn't. The auditor's original seed had
never given it a description, the database allows that, and every piece of code that loads an
agent choked on the empty field. **This agent had never been startable in the normal way** — which
is a second, independent reason "nobody ran it" was true, on top of the wiring. Worse: the loop
is designed to carry on when one auditor fails, so had I not been watching this exact run, every
sweep would have reported success with the auditor quietly skipped. Filed as bug 290; the one-row
fix is live and the code fix (make all five loaders tolerate the empty field, plus a build check
so a sixth can't reappear) is committed for the next release. Whether the column should be made
mandatory outright is a schema change I've left as your call.

Second attempt: it worked, end to end. Eight findings, every one routed to a real handler — the
"cookie-cutter" and "animations beyond hover" complaints from July are back in the queue,
this time reachable. Bugs 279 and 115 are closed. Still open and owned: 284 (what keeps
grabbing parked roadmap rows) and 287's code half awaiting its release.

---

## 2026-08-16 — the findings all landed; reading the pages afterwards is where it gets interesting

Yesterday ended with the analyser enrolled and its first findings starting to move. Overnight they
all finished moving. Seventeen findings across three sites: thirteen finished, three failed, one
correctly parked for you. So the machine works end to end, unattended — that part is settled.

Then I did the thing we keep telling ourselves to do, and read the actual pages rather than the
status column. Two things came out of it that the statuses do not say.

**First, the good news, and it is real.** The web design home page now opens: *"Sixty-three browser
tools for front-end work. No account, no upload, nothing stored. Everything runs client-side, so
nothing you type or drop into a tool leaves your machine."* That replaced "Tools and guides for
people who build websites". The gas wholesalers home page now opens by saying every load is priced
against the Platts daily benchmark plus a single fixed margin — a checkable operational fact where
there used to be a category label. I proved the first one was our work and not a coincidence: the
page was rewritten thirteen seconds before the finding was marked done, by the job the finding
spawned.

**Second, the catch — and it is the same mistake you complained about in the first place.** Each
finding is filed with a test attached: a plain sentence saying what "fixed" would look like. The
web design one said the new copy must mention at least two of *no account, runs in the browser,
nothing stored* **before any count of tools or articles**. The rewrite mentions all three. It also
leads with "Sixty-three". That is a count of our own inventory — the exact thing the analyser's own
"do not lead with" list puts at number one, and the exact thing that made you reject "23 free UK
calculators" last week. The finding was marked complete anyway.

The reason is structural rather than anyone being careless: **nothing reads the test.** The agent
that writes the fix is not shown it, and the agent that marks the item done does not check it. We
built a falsifiable test into every finding and then never asked the question. On the gas
wholesalers page the same gap shows up differently — the finding was specifically about the page
*title*, the title still contains the exact phrase it objected to, and the item is marked complete
because the hero underneath it got rewritten.

**Third, two of the three failures are not what yesterday assumed.** I had put them down to the
known messaging fault. Two of them are something else entirely, and it is worth knowing about: a
quarter of the pages on the estate (172 of 704) are marked as owned by a tool or widget, which
means a generic "rewrite this page" is refused outright to stop it destroying the tool. The refusal
is correct and the error message even names the right alternative — but nothing acts on it, so the
finding just dies. This is not our problem alone: the design auditor and the automated checkers
have items sitting dead on owned pages too. The route that works on those pages exists and is used
successfully eighteen times elsewhere; no content finding is ever put onto it.

So: the analyser is producing findings that make sites genuinely better, and the weakest link has
moved. It is no longer "can we find the right thing" — it is "does the fix actually satisfy what we
asked for, and does anyone ever check". I have not fixed either of these yet; I want your steer on
which to take first.

---

## 2026-08-17 — a correction to what I told you on Saturday, and one thing filed

I came back to this a day later and re-checked my own figures before adding to them. Nothing had
moved: still three sites with the ranked record, still the same seventeen findings in the same
states. That turned out to be the interesting part.

**The correction first, because it is mine.** When you approved enrolling the analyser into the
automatic sweep, I told you its findings would "start moving on their own the next time the sweep
visits each site — nobody has to push them". That is true only if a sweep happens, and **the sweep
is switched off**. It is switched off on purpose, by you: back on the 11th you said *"lets reenable
improvement-sweep for the rerenders for a short while - it will be expensive so I am wary of
costs"*, and it has been off again since the 14th. So it runs in short windows you open, not
continuously. Everything else about the enrolment stands — it is wired in correctly and we watched
it work inside a real sweep — but it is a precondition, not a self-winding clock. Growing this past
three sites needs you to open a window, or a session to fire it deliberately. I should have said
that plainly at the time; the fact was written in our own notes and did not make it into the
sentence you read.

To be clear about what is *not* broken: the rest of the scheduling estate is running normally —
twenty-two other scheduled jobs fired within the hour I checked. This one is specifically off.

**The thing I filed.** Yesterday I told you two of the three failures were a guard correctly
refusing to overwrite a page owned by a tool. I have now traced why that refusal goes nowhere.
There are three places in the platform that stop a generic rewrite from clobbering a tool page.
Two of them, having refused, file a note for a human saying "this page was refused, here is the
right way to edit it". The third — the only one on the path our findings actually take — just
throws an error, so the work item dies and the reason disappears within a day when the run record
ages out. That is filed as bug 295, with the preferred fix being about six lines: make the third
one file the same note the other two already file. It is worth doing beyond our own lane, because
the design auditor and the automated checkers have items dying the same way on the same pages, and
a quarter of all pages on the estate (172 of 704) are tool-owned.

I checked one thing before writing that up, and it changed the wording: the guard is not wrong to
refuse, and the session that built it had already thought carefully about this area. The fault is
only that the refusal leaves no trace. Reading their commit rather than just the code is what
stopped me filing it as something it is not.

**Where that leaves us.** Two real gaps are now on the table and neither is fixed: nothing checks
whether a fix actually satisfied the test the finding was filed with, and a refusal on a tool-owned
page vanishes. Plus the two tracks you already approved — the claims audit over our written
premises, and the analyser's v2. I would like your steer on the order.

---

## 2026-08-17 (afternoon) — both your decisions are done, and I got three things wrong along the way

**The bug you told me to fix first is fixed and reviewed.** The refusal that goes nowhere — where a
tool-owned page correctly refuses a rewrite and then the reason vanishes — now leaves the same note
for a human that the platform's two other refusals already leave. It is about six lines. I proved
the test works by deliberately breaking the fix twice and watching it fail both times, then putting
it back. The review council passed it first time, all thirteen reviewers.

Two of their comments were worth more than the approval. One reviewer pointed out that the advice
my note gives — "use the targeted section editor instead" — is right for rewriting something that
already exists and is a **dead end** for adding something new, because that editor cannot add. That
is true, I checked, and I have left it alone deliberately: the wording belongs to a shared piece of
machinery used by four different callers, and changing it inside a small fix is exactly the sort of
scope creep four other reviewers had just praised the change for avoiding. It is written down as a
follow-up with the correct route named. The other reviewer noticed I had claimed "three places in
the code do this" without ever having run the search. I ran it. It is two, not three — the third
gets there by a different path. The conclusion held; my count did not.

**The sweep window ran for twenty-four minutes and I closed it again.** In that time it visited two
sites on its own — a games design site and a robotics one — and wrote the ranked "lead with this"
record for both. That takes us from three sites to five. Nobody touched anything: the sweep chose
the sites, ran the analysis, filed the findings and started dispatching them.

**The number you need, if you want to decide on a longer window:** it is about one site every
fifteen minutes, and *every* site is currently due an audit, so each visit is the expensive kind —
a full audit plus the offer analysis plus dispatching what it finds. Eighteen sites still have no
ranked record, so finishing the estate is roughly four and a half hours of open window. Say the
word and I will open one; it stays shut until you do.

**One thing that is genuinely pleasing.** On the games design site, the analyser's "do not lead
with" list opens with *"the number of tools or guides currently on the platform"*. That is the
fourth site where it has independently put that first — and it is exactly what our own writer then
did to the web design home page yesterday, which now opens "Sixty-three browser tools". The
analyser knows the rule and the writer breaks it. That is a solvable problem and I have written up
how.

**Now the three things I got wrong, because they are the useful part.**

The first two were the same mistake twice: our database reports times in UTC and my terminal
reports them in British summer time, an hour apart. So a sweep that had been working for two
minutes looked like it had been stuck for seventy-two. I caught that one before it left my screen,
wrote up the trap so nobody repeats it — and then twenty minutes later told you the sweeps were
running slower than their schedule, which was the *same error* wearing a different hat. They were
exactly on schedule. What I take from it: writing down a rule does not protect you from it, because
by then the wrong number had become an explanation, and I went looking for evidence to support the
explanation instead of rechecking the number.

The third is subtler and worse. I reported that one site was missing a piece of its commercial
record, and used that as the reason to make a change to the analyser. Seven minutes later the site
had it. The site had been **created that morning and had no pages yet** — another thread was
building it while I measured, and I had caught it half-finished and written that instant down as a
fact about the estate. My own output had "created today" printed in the column next to it. The
underlying gap in the analyser is real but now has no live example, so I have downgraded it and
struck the claim through rather than quietly deleting it.

All three are logged where the next person will trip over them.

---

## 2026-08-17 (late afternoon) — the fresh build did not ship any of today's code, and that is not just my fix

You told me a fresh chassis had been deployed, so I went to check my fix was live rather than
assume it. **It is not — and neither is anything else anyone committed today.**

The pods genuinely did restart, at 14:42, on a new deployment. But the version label was never
changed: the fleet was on `v1.0.1305` before and is on `v1.0.1305` now, and when the label does not
change, the machines are entitled to reuse the copy they already had. So new pods came up running
**yesterday evening's code**. I confirmed it by asking the binary itself which commit built it —
it answers with one from the 16th — and there are **222 commits sitting in our repository that are
not in the thing actually running.**

**This is not about my change.** I tested for a distinctive piece of another thread's work,
committed at lunchtime today, and it is absent too. Three markers from yesterday are all present.
So the line is clean: everything committed on the 16th is live, everything committed today is not,
across several threads who each believe their work has shipped. One of them independently found the
same thing this afternoon and wrote it down, which is how I know it is not a quirk of my method.

**One thing that makes this genuinely confusing, and worth you knowing:** database and
configuration changes go live immediately and are completely unaffected. Only compiled code is
stuck. So a thread that changed configuration sees its work behaving correctly, and a thread that
changed code sees nothing happen — and both look equally "deployed" from the outside.

**What is needed is one line:** the version label in the makefile needs bumping before the rebuild,
then a normal release. Re-running the deploy at the current label cannot help, because the machines
will keep serving the copy they have. That is your call and your command, so I have not touched it.

**The silver lining, and it is a real one.** Because the fix is definitively not live, the bug it
fixes is still reproducible — and it reproduced this afternoon exactly as I predicted it would,
unprompted, on a tool page on the games design site. The sweep filed the finding, the system tried
to rewrite the page, the guard correctly refused, the task died, and no note was left for anyone.
That gives us a clean "before" measurement to compare against once the fix does ship: the same site,
the same page, and a count that is currently, provably, zero.

---

## 2026-08-17 (evening) — the second build was real, and the fix is now proven working

The version label was bumped this time, so the machines actually picked up new code. I checked the
binary rather than trusting the deployment, and the fix is in it.

Then I waited for the thing that actually matters, which is not "is it installed" but "does it do
anything". **It does.** Over the evening, eight tool-owned pages across four of our sites were
put through a generic rewrite, correctly refused — and each refusal now leaves a note for a human
saying which page, why, and what to do instead. Before this fix there had been **zero** such notes
in the entire history of the system; the refusals simply vanished.

**The check I care about most is the one that could have embarrassed me.** Eight notes appearing
proves nothing on its own — a broken version of this fix that filed a note every single time would
produce exactly the same eight, plus a great many more nobody would notice for weeks. So I checked
the opposite: in the same period, six ordinary pages were rewritten successfully and **produced no
notes at all**. That is the result that separates "it works" from "it fires at everything", and
it is the only reason I am willing to call this proven rather than probably fine.

Two smaller things held up as designed. The same page was refused twice within five minutes and
produced one note, not two. And the failed tasks stayed marked as failed rather than being quietly
marked done — which matters, because a task falsely marked done is a worse problem than the one we
just fixed.

I have closed the bug. Two things deliberately did **not** close with it, and both are written
down where the next person will find them. The note we now leave tells a human to use the targeted
section editor — which is right if they are changing something already on the page, and a dead end
if they need to *add* something, because that editor cannot add. And more importantly: this makes
the refusal visible, it does not make the page get fixed. The tool-owned pages that need content
changes still need a route that works on them, and that is a separate piece of work.

**One number worth your attention.** While measuring, I found **26** content tasks that have died
on tool-owned pages over the past week, from **six different parts of the system** — the automated
checkers, the design auditor, the cross-linker, the offer analyser and two others. I can only prove
three of those were this exact fault; the rest are past the point where the evidence expires. But
that is the scale of what has been disappearing quietly, and from tonight it is all visible.

---

## 2026-08-18 — the fix's first day turned up something more expensive than the thing it fixed

Yesterday's fix has been live for a day. It is working: **fifty-nine** refusals recorded across five
sites, each one a tool-owned page the system tried to rewrite generically and correctly declined.
Every one of those would previously have vanished without trace.

But the shape of them is the story. Forty-nine are on the web design site — **half of all its
tool-owned pages** — and they arrived in a burst between three and four in the morning. So I
followed the burst, and found this:

**In two and a half hours the platform wrote content for thirty-nine pages using the AI writer,
resolved all their internal links, and then threw the lot away** — because the check for "is this
page owned by a tool?" happens at the *last* step of the process, not the first. The information
needed to skip the whole thing is available at step two. It is consulted at step twelve.

So every night we have been paying for AI writing that could never be saved, and until yesterday
there was no record of it happening at all. I have filed it (bug 301) with the fix, which is small:
move the check to the front, keep the one at the back as a safety net.

**The wider point, and I think it is the useful one.** The fix I shipped yesterday was deliberately
unambitious — it did not repair anything, it just made a silent refusal leave a note. It looked
like the boring option. Within a day that note-leaving surfaced a cost problem considerably larger
than the missing notes were. I would not have found this by looking for it; I found it because the
system finally said what it was doing.

**One honest caution about my own number.** The "thirty-nine" comes from three separate counts over
the same period that all agree — thirty-nine writer runs, thirty-nine link resolutions, thirty-nine
failed builds. That is strong, but I have not traced one individual build end to end to prove they
are the same thirty-nine, and the records expire within a day so it has to be done on a fresh night.
I have written that limitation into the bug rather than let the number stand as firmer than it is.

**Where that leaves things.** The analyser is on five of twenty-three sites and will not reach the
rest without you opening the sweep again — about four and a half hours of running to finish the
estate, and it stays shut until you say. The next piece of building work is making the analyser's
own "what would fixed look like" tests machine-checkable, which is where we found it quietly
reintroducing the very fault it had flagged.

---

## 2026-08-20 — the analyser wrote something that isn't true, and someone else caught it

I started today by checking what had changed rather than carrying on, because a lot happens here
overnight. Nearly three hundred commits had landed. Two of them closed out the tool-owned-page work
— another thread took the fix I filed, shipped it and proved it, and then filed the remaining piece
as its own job. That is all good and none of it needs anything from us.

**The thing that does need us is ours, and it is not comfortable.**

Yesterday the analyser was run again against the Leopardess site. Its number-one recommendation —
the single thing it says that site should lead with — contains the phrase *"the same stack that runs
eight live sites"*. **We run twenty-three.** The figure is wrong, and it is the kind of wrong that
would have gone straight into a headline.

Worse, it is tagged as having come from the site's recorded strategy. It did not. I checked every
strategy record we hold and none of them contains that phrase. **Two of the site's own web pages
contain it, in their search descriptions** — and the analyser reads those descriptions as part of
what it looks at. So it picked a stale claim off a page, promoted it into the strategic record, and
labelled it as though the strategy had said it.

**We did not catch this. The Leopardess thread did.** They held all five of the analyser's findings
and wrote down exactly why, one step before any of it reached a writer. I want to be plain about
what that means: every check we built passed. The record had all its fields, correctly flagged the
one piece of missing input, ranked its points properly and explained its reasoning coherently — and
carried a false number at the top. **Our checks confirm the shape of the answer, and nothing we had
was looking at whether the facts in it were true.**

I have written it up as a bug against our own analyser, with a fix that makes the mistake hard to
make rather than asking the model to be careful: if a recommendation contains a number, that number
has to appear in the record it claims to have come from, or it does not get written. There is a
trap in the obvious version of that fix — ban numbers outright and the recommendations become
useless — so the test includes a site where the numbers *are* properly sourced and must survive.

**One connection worth drawing.** Back on the 14th you approved checking our written strategy
records for invented claims. That job is still not started, and today makes the case for it
stronger — but it would have caught this *after* the false sentence was written, not stopped it
being imported. Both are needed. This one is the tap; that one is the mop.

**Where we are otherwise:** five of twenty-three sites carry the ranked record, and the sweep stays
shut until you open it — about four and a half hours of running to finish the estate.

---

**2026-08-21 — the false sentence can no longer be written, though it is not switched on yet**

Yesterday's problem was that our own analyser put a number into the most important line on a site —
"the same stack that runs eight live sites" — and labelled it as coming from the site's own strategy
record. The number was wrong (it is twenty-three), and the strategy record it pointed at contains no
number at all. The label is the part that stings: it is the very thing we built so a reader could
check a claim, and it vouched for something the record never said.

It is now fixed, in two pieces. The first is a check that runs automatically between the moment the
model answers and the moment we save: if a sentence states a quantity, that quantity has to actually
appear in the record the sentence says it came from, or the sentence does not get saved. The second
is a line in the instructions telling the model the rule up front. We want both — the instruction
makes it rarer, the check makes it impossible.

**Neither is live yet, and that is deliberate.** The check is new program code, and program code on
this platform only starts working after the next fleet release. The configuration change that
switches it on is being deliberately held back until that happens, because switching it on early
would not merely fail to work — it would break the whole analyser, including the parts that are fine
today. The file that turns it on carries the instructions for doing that safely.

**Two things I got wrong, both worth you knowing.**

The first is mine from yesterday. I wrote that we should test the fix against gaswholesalers, to make
sure a rule against invented numbers did not also strip out legitimate ones. When I actually looked at
every sentence on every site, gaswholesalers turned out to contain **no numbers at all** — so it would
have passed no matter how badly the rule was written. It was a test that could not fail, which is the
same as no test. The sites that genuinely test it are webdesign ("sixty-three tools") and robot-hands
("six actuation types", "2–3 articles a month"), where the numbers *are* in the record and must
survive. Those are now built into the code's own tests.

The second is subtler and nearly shipped a check that would have missed the very thing it was for. We
already had something similar elsewhere on the platform, and I reused its approach — but that one only
looks for numbers written as digits, and our bad sentence said "eight", in letters. It would have run,
found nothing, and reported everything as fine. Two of the three legitimate cases are also written in
letters, so the lazy alternative — banning spelled-out numbers outright — would have fixed the bug and
ruined the artefact at the same time.

**What is still owed.** The change went to the review council and I have not yet read the verdict. And
when the release does happen, switching it on does **not** repair leopardess by itself: the wrong
sentence is already saved, and only a fresh run replaces it. The automatic sweep has been off since the
17th on your cost instruction, so nothing will re-run that site unprompted — and the leopardess lane is
holding our findings there pending a design report for you, so we should talk to them before firing
anything at that site.

**Where we are otherwise:** unchanged — five of twenty-three sites carry the ranked record, and the
sweep stays shut until you open it.

---

**2026-08-22 — the check is switched on, the false sentence is gone, and I want to be careful about what that does and does not prove**

The fleet released overnight, so this morning the program code from yesterday was finally running.
I checked that properly rather than assuming: I asked the running service directly whether it
contains the new check, and I asked it two control questions as well — one thing that should be
missing (it was) and one that should be there (it was). Only then did I switch the configuration on.

Then, with your go-ahead, I ran the analyser twice: once on webdesign, once on leopardess. Both
completed. **The "eight live sites" sentence is gone from leopardess.**

**But I do not want to let that stand as "the check works", because it isn't quite what happened.**
The check removed nothing on either run — its own record of removals is empty both times. The false
sentence is gone because the *instruction* we added stopped the model writing it in the first place,
not because the check caught it on the way out. Those are two different halves of the fix, and only
the first has been shown working on real data. The catching half is proven by tests built from the
actual live sentences, but it has still never fired in production. I would rather say that plainly
now than have someone later discover the proof was thinner than it sounded.

**Two things came out of those runs that are worth your attention.**

The first is a genuine mistake of mine that the runs exposed. The check counts numbers, and it was
counting the "2" inside "B2B" as if someone had claimed a quantity of two. Same for "S3", "IPv6",
"Web3". That means a perfectly ordinary sentence about serving "UK B2B SaaS teams" could have been
deleted for asserting a number it never asserted. On leopardess it survived purely by luck — the
site's own strategy record happens to contain "B2B" as well, so the "2" was considered accounted
for. I have fixed it, but the fix is program code, so it will not be live until the next release.
Until then I would not run the analyser by hand without accepting that small risk.

The second is a cost I did not anticipate. Both new versions contain no spelled-out numbers at all,
where the previous versions each had one. Webdesign's opening line used to say "any of the
sixty-three tools" — which was true, and was drawn straight from the site's own record. It is gone
now. So the instruction may be making the model shy of numbers in general rather than shy of
*unsupported* numbers, and we would be trading away exactly the kind of specific, checkable detail
these records are supposed to supply. Two runs is not enough to be sure; it is enough to watch.

**One thing I stopped myself doing.** The normal way this lane runs the analyser fires the whole
improvement loop, and that loop promotes every queued item on the site into live page edits — 111
of them on webdesign, 37 on leopardess, including work other people had queued. To test one check,
that would have been wildly disproportionate, and it was not what you agreed to. I wrote a small
dispatcher that runs only the analyser instead.

---

**2026-08-24 — I was wrong about the numbers disappearing, and the answer was in front of me**

Two days ago I told you the new instruction might be making the analyser shy of numbers in general,
and that webdesign had lost "any of the sixty-three tools" as a result. I said it was worth watching.
I have now watched it, and **I was wrong** — the instruction is not doing that.

Two things settled it. First, I ran the analyser on robot-hands, and it **kept** its number: the
fourth line still reads "across six actuation types", which is drawn straight from that site's own
strategy record. Something that suppressed numbers generally could not have left that alone.

Second — and this is the part I should have checked before saying anything — every version of these
records also contains a short list of things the site should *not* open with. Webdesign's list has
said "a count of how many tools or articles the site contains" in **all three** versions, including
the one from before my change. So the old version that led with "sixty-three tools" was contradicting
its own advice on the same page; the new ones simply stop doing that. If anything it is tidier now,
not poorer.

The two sites divide exactly along that line: robot-hands avoids an *inventory* count and kept a
*categorical* one ("six actuation types" — you cannot say the sentence without the number); webdesign
avoids an inventory count and dropped one. Nothing was lost that the site had not already said it did
not want.

**Why I am making a point of this rather than quietly moving on.** The claim reached four documents,
including this one, and it read like a finding because I had attached a number to it. Adding "only
two runs, worth watching" made it feel careful without making it true. The check that would have
caught it was one query against data I already had open. I have written it up in the fleet-wide log
of wrong calls, because the useful part is the pattern: **when my own change is the obvious
explanation for something that changed, I should spend one query looking for an instruction that was
already there before me.**

**Where that leaves the actual job.** The false "eight live sites" sentence is still gone. The check
is live and behaving. The one thing still not shown is the check *catching* something — it has now
run four times across three sites and removed nothing, because the instruction is stopping the bad
sentences before they are written. That is the good outcome, but it does mean the safety net itself
is still only proven by tests rather than in the wild.

---

**2026-08-24 (later) — the job is done and the bug is closed**

The fresh build you deployed carries both halves of the fix, and I checked that at the running
service rather than assuming it, with two control questions either side.

I then ran the analyser across **all five** sites that carry these records, so the evidence is the
whole estate rather than the one site that had the problem. Every run completed, every one shows the
new check actually executed, and **not one of them produced an unsupported number**. The false
"eight live sites" sentence is still gone from leopardess — and worth noting, the pages on that site
*still* carry the phrase in their metadata, so the thing that caused the problem is still sitting
there and the analyser no longer picks it up.

**The test I had been missing finally happened.** Until today every clean result was clean because
the analyser happened to write no numbers at all — which proves nothing, since a rule about numbers
is trivially satisfied by prose that has none. Today robot-hands **kept** its number: "across six
actuation types", which is drawn word-for-word from that site's own strategy record. That is the
first time the check has been shown leaving a legitimate figure alone on real data.

I have closed the bug.

**Three things I have written down rather than quietly dropped**, because they are conditions on
whoever touches this next rather than problems today:

1. **The check has still never actually caught anything** — it has removed nothing, ever, because the
   instruction stops the bad sentences being written in the first place. That is the outcome we want,
   but it means the safety net itself is proven by tests rather than in the wild. Nobody should point
   at a clean run and call that proof the net works.
2. **When the check does remove something, it writes that into the record but nothing reads it.**
   Fine while nothing else consumes these records — which is true today — but that must be fixed
   before anything starts to.
3. A small known gap in how numbers are recognised: "GPT-4" is still read as the number 4. I have
   pinned it with a test that will fail if anyone fixes it, so the note cannot go stale.

**And I corrected myself again.** Earlier today I told you the numbers disappearing was possibly the
new instruction being too cautious. It was not, and the answer was in the same records I was already
reading. I have written that up properly in the fleet-wide log of wrong calls, because the pattern is
worth more than the incident: when my own change is the obvious explanation for something, I should
spend one query looking for an instruction that was already there before me.

---

## 2026-08-24, later the same day — the analyser can now write a test a machine can check, and I found three pages we had already called finished

Picking up where the handoff pointed: the next job on the list was to let each finding carry a small,
checkable version of its own "how would we know this was fixed" test, instead of only a sentence.

**First I went to see whether it was worth doing, and the answer was uncomfortable.** Every finding
the analyser writes comes with a test in plain English — *"the meta description must state the
no-account promise before any mention of how many tools there are"*, that sort of thing. Nothing has
ever checked one. So I read all thirty-seven of them we have, and then went and looked at the actual
pages.

Three of them are sitting on jobs we have marked **done**. The page was rebuilt, deployed, the job
closed. And the thing the job itself said would make it right is still not true — checkable in one
line, against the exact words on the page.

The clearest is our own webdesign.co.uk front page. The test asks for the reader benefit *before* any
count of tools. The description we are serving right now opens **"Sixty-three browser tools for web
design and development. No account, no upload…"**. The count comes first. That page has been rebuilt
twice since the job was closed and it still reads that way. Another is a robot-hands page whose test
says it should appear in the header menu; it does not appear there at all.

Nobody was careless. There is simply nothing in the platform that reads the test back after the work
is done, so "the handler finished" and "the page now passes" are the same word — *complete*.

**What I have built is the first half of fixing that.** A finding may now carry, next to the English
sentence, a small structured condition: *this phrase must be absent*, *at least two of these three
must appear*, *this must come before that*. Only over a page's title and description, only the three
shapes, and it is entirely optional — most findings will carry none, which is the right answer.

**Two rules make it safe, and they are the whole design.**

The first is that a condition can only ever say **no**. Passing it means "not caught", never "this is
fixed" — the English sentence stays the authority. That matters because two thirds of these tests
bolt a checkable clause onto a matter of judgement, and a tick against the easy half would be worse
than no tick at all.

The second is that the condition has to **already fail today**. The finding says the page is wrong
now, so a condition that genuinely captures it must fail now. It is the only thing we can check at
the moment the condition is written, and it throws out the useless case — a rule about a word that
appears nowhere, which would pass for ever and look like verification. When one is thrown out we
write down why, so the cost of that rule is visible rather than silent.

**I nearly shipped the exact mistake this is designed to prevent.** My first version also checked the
header menu, and it "found" a fourth broken job: leopardess's test says the menu should have no more
than seven items, and our database says thirteen. Decisive-looking. Then I loaded the actual page —
the menu shows seven. The database column is not the menu; the site quietly leaves four of those out.
The test was passing all along, and I had a check ready to declare it broken with an air of
arithmetic. I have dropped menus from this feature entirely and written the trap down where the next
person will hit it, because the obvious alternative source turned out to be an empty column.

**Where it stands.** The code is written, tested twenty-six ways, and proven to build against the
committed shared branch rather than just my own copy. It is **not yet switched on**: the database
change that turns it on has to wait for a rebuilt service, or it would break the analyser outright.
I have written that change and its undo, and left the switch-on instructions in the file itself.

**Two things I could not finish, and why.** The review-council submission is written and passed every
local check, but the cluster login expired part way through the afternoon (everything comes back
"Unauthorized"), so it has not been sent — I would rather leave it queued than fire it down a dead
connection and be told a review is running when it is not. And you mentioned a fresh service being
built and deployed: that build takes the last *committed* code, and this work was not committed yet
when it started, so **that deployment will not contain it** and the switch must not be flipped on the
strength of it.

**One honest limit.** Nothing automatic reads these conditions yet, so today this prevents no false
"done" — it turns a sentence into something a machine *can* check. The piece that would actually stop
a job closing while its own test fails is a separate change at the closing step, and it changes
behaviour for handlers other lanes own, so it deserves its own review rather than riding in on this
one. I have named it, in the place the next person will look.

---

## 2026-08-25 — it is switched on, and the first thing it did was catch us

Your rebuild carried the code, so I switched it on. Probed both copies of the service first, with a
control that was capable of coming back negative, then applied the database change by hand and read
it back independently.

**Then I fired one run at webdesign.co.uk to see whether the analyser would actually use it.** It did,
without being told to — the option is described in the instructions but deliberately left out of the
required output, so it had to reach for it. Four findings, three carrying a checkable condition, one
left bare. All three conditions correctly say "this page fails right now". The bare one is the right
answer for a finding that turns on judgement, and I am glad it happened on the first run rather than
being something I had to argue for.

**And then the first run caught us.** One of those findings — the front page description — was picked
up, the page was rebuilt, deployed, and the job closed as **done**. The condition stored on that job
still says the page fails. I checked it with the platform's own checker rather than by eye, and the
page has since been rebuilt a second time and still reads *"Sixty-three browser tools…"*, with the
count first, which is exactly what the job said had to change.

So the thing I described yesterday as a suspicion is now a measurement with a machine behind it. I
have written it up as a bug in its own right — `bugs_open/395` — because it is not really about our
lane: **"done" means the handler finished its job, and nothing anywhere reads the criterion the job
was created with.** The handler did rebuild the page; it just did not do the thing that was asked. No
part of the platform is in a position to notice the difference.

**One mistake of mine, and it is worth you knowing.** Before firing that run I checked that the
automatic sweep was switched off and concluded a single run could not change any live page. It did:
another loop, which is always on, picked the findings up thirty-one seconds later and rebuilt the
page. The outcome was harmless — our own site, our own findings, the pipeline working as designed —
but the reasoning was wrong in a way that would not have been harmless on the leopardess site, whose
lane is holding five of our findings and waiting for a decision from you. I checked whether the one
mechanism I had thought of was off, and reported that as "nothing will happen". The correct check is
the opposite question — *what is switched on?* — and it is one query.

**Where that leaves us.** The producer half is live and behaving. The half that would actually stop a
job closing while its own test fails is not built, and it changes behaviour for handlers other lanes
own, so it wants its own review rather than being slipped in. That is written up as the next job, with
one caveat I want to flag: every condition we have so far says "fails". Until one of them says "passes"
after a real repair, a gate built on this cannot be told apart from a gate that refuses everything.
That control has to be manufactured deliberately.

Everything is committed, and the next session starts from
`docs/agent_docs/docs024_key_docs_latest/vigilant_designer_offer_analysis/HANDOFF_2026-08-25_continue_here.md`.

---

## 2026-08-25, afternoon — the loop closes: the system now reads its own homework back

Yesterday we got the analyser to write down, for each thing it says is wrong with a page, a small
test a machine can run. Three of those went live on webdesign.co.uk. And then one of them proved the
thing we had suspected for weeks: an item was marked **done** — page rebuilt, deployed, commit
recorded — while the test it had set itself was still failing.

Nothing read those tests. That was the whole point of yesterday's handover: build the half that
reads them.

**That half is now built, reviewed and committed.** It is a check that runs in the moment between
"the agent says it finished" and "the system writes down that it's done". If the item set itself a
machine-checkable test, the check re-runs that test against the page as it stands right now.

### What it does today, and what it deliberately does not do

It **records**. It does not refuse.

That is the one decision worth explaining, because refusing sounds obviously better.

The problem is that every single test we have ever had is currently **failing**. All three. We have
never once seen this check say "yes, that's fixed" about a real page. So if I switched on refusing
today, and it started blocking things, I would have no way to tell the difference between a check
that is working and a check that simply says no to everything. That is not a hypothetical worry —
it is the exact mistake this lane has made before, and the bug file says so in as many words.

So it watches, and it writes down its verdict on every item it looks at — **including the ones it
passes**. That last part matters more than it sounds. If it only recorded failures, then a check that
had quietly stopped working would look identical to a check that was passing everything. Recording
the passes is what makes the difference visible.

### And it will not be allowed to just sit there watching for ever

The honest risk with "we'll only watch for now" is that watching becomes permanent because nobody
remembers it was meant to be temporary. So the promotion is wired to **break the build**.

There is a separate, older mechanism that can quietly mark items done when they time out, going
round every check we have. The moment anyone flips this from watching to refusing, the build fails
until they have also closed that side door. It is not a note in a file that someone has to read. The
code will not compile.

### Two things I got wrong, both worth telling you

**First, I nearly built it for the wrong reason.** Yesterday's handover explained *why* the check had
to go in one particular place, and I was about to copy that explanation into the new code. It was
wrong — it was an argument about a different check, answering a different question, and it happened
to be sitting nearby. The conclusion was right; the reasoning was borrowed. I had also written that
same wrong reasoning into the main debugging guide the day before, so I have struck it out there
with a note saying what it should say.

The lesson I'd take from that: when a handover gives you a decision *and* its justification, the
justification is the part to re-check. It was written about whatever was in front of that author.

**Second, and this one nearly shipped something useless.** When the analyser writes one of these
tests, the system stamps a little record onto it — "checked on this date, verdict: failing". Perfectly
sensible. But the checker that reads these tests is strict about what it will accept, and it does not
recognise that stamp. So if you hand it back a test straight out of the database, it says "I can't
read this" — **every time, for every test we have.**

The check would have run, reported no problems, and been completely blind. And the message it gives
reads like the analyser wrote something malformed, so the natural reaction would be to go looking in
entirely the wrong place.

What makes this worth telling you rather than just fixing: **no test would have caught it.** The one
test we have that uses real live data types the tests out by hand, without the stamp — so it was
testing a shape that does not exist in the database. I found it by reading the live record next to
the rules, not by anything failing.

### The review

Fourteen reviewers, approved first time round, with five advisory notes and nothing serious.

One of them was worth the whole exercise. A reviewer asked whether there was a *second* route by
which items get marked done that would go round my new check entirely. There is such a route, it is
a known trap, and I had not checked it. I went and looked: it turns out that route uses the same
door I am standing in, so the check does cover it — and 1,600 of the 1,638 items of this kind in our
whole history went through exactly that door, including the one this bug is about. But 38 of them,
about 2%, spread over five months, went through something else. Nothing is lost today, because none
of those 38 ever had a test attached. It is written down as the thing to deal with before switching
refusing on.

Two other reviewers, independently, made the same broader point: this is now the **fourth** check
bolted onto the same piece of machinery, each one hand-wired the same way. Neither blocked it, but
one asked that it be *named* rather than quietly absorbed. So I have written it up as a proposal for
you rather than burying it in a commit message — with a recommendation that we do the small version
(share the one piece that genuinely repeats) rather than build a framework around four examples.

### Where this leaves us

The producing half and the reading half now both exist. Nothing is prevented yet — that is
deliberate, and it is waiting on one thing: **the first time this check sees a page that has
genuinely been fixed and says so.** After the next fleet update I want to run the analyser again and
watch for that. Until it happens, we have a check that has only ever seen failure, and I am not
willing to give that authority to block anything.

---

## 2026-08-25, later the same evening — the job could never have been done, and that was already written down

*Added by the session working `bugs_open/395`, picking the bug up from this lane's handoff.*

The entry above ends by saying we are waiting for one thing: the first time the new check sees a page
that has genuinely been fixed. **That was never going to happen, and here is why.**

Before writing any code I asked one question the bug file had never asked: *who actually changes the
thing the check is checking?* All three of the tests we have written so far are about a page's
**description** — the sentence that appears under the page title in a Google result. So I went and
found every piece of code that can write one.

There are four. Three of them will only ever *fill in a blank* — if a description already exists,
they leave it alone, deliberately. The fourth can genuinely replace one, but only one agent in the
whole system can reach it, and that agent is set up to visit only pages whose description is
**empty**.

And the agent we actually sent to fix the page — the one that rebuilt it, deployed it, and reported
success — **has no ability to touch a description at all.** Not "did the wrong thing". Could not have
done the right thing, whatever we had told it.

So the item we filed was work nobody in the system could do. The agent went, rebuilt the page
honestly, came back and said it was done, and we wrote down `complete`. The check we built yesterday
is right to say the criterion is still false. It will go on saying so for ever.

### The part that stings, and is the more useful finding

**We already knew.** On the 19th, another thread investigating why more than half our pages had no
description at all wrote up the same page, the same field, the same agent, the same
green-and-did-nothing outcome — and finished with a plain instruction: *do not file these items,
because they complete and achieve nothing.*

Five days later a different part of the system filed exactly that item.

Nobody was careless. That instruction is clear, well-evidenced, and sits in a file about *missing*
descriptions, while this thread was searching for anything about *checking finished work*. Two
threads looking at the same rows, describing them in words that share nothing, so neither finds the
other. The instruction was a sentence in a document, and **a sentence cannot stop a piece of software
doing something.**

### What I have done about it

I have turned that sentence into code. From now on, when the system writes down a complaint about a
page, it checks whether the agent it is about to send can actually change the thing being complained
about. If it cannot, the complaint is still recorded — filed as *"we can see this and currently have
no way to fix it"*, which is a list two other parts of the system already read — but it is no longer
handed to an agent who will come back and say it is done.

It is written so that this cannot quietly stop working: I deliberately broke it three separate ways
and confirmed each break made a test fail. It is with the review council now.

### The one thing that needs you, and I have not done it

Making these complaints actually *fixable* means letting an automated process rewrite the published
description of a live page. **You withheld exactly that permission on the 21st** — you allowed a
one-off pass over 681 pages and were explicit that the standing machinery must not have it.

So the choice is yours, and it is a real one: leave these findings on the "seen, can't fix" list, or
give something the authority to rewrite a published description when a finding asks for it. I have
not taken that decision and neither has the other thread.

Until then the check we built yesterday will keep watching, and it still cannot be allowed to block
anything — not because it is unproven, but because the work it would be blocking is work nobody can
currently do.

---

## 26 August — the pictures job turned out to be two jobs, and the second one was quietly about to break a page

You asked for pictures between the paragraphs on built pages. Here is where that got to today.

**The good news first: we did not need to build anything.** There is already a section type that does
exactly this — a heading, an optional picture, then the prose. It is well made: the picture is
optional, so if a site has no suitable image the section just renders as text with no gap or broken
frame. Somebody built it properly a while ago.

**Nothing was choosing it, and the reason was almost silly.** When the system plans a page, it is
shown a menu of the available section types with a short description of what each one can do. That
description is generated automatically. It could say "this one does text", "this one does lists",
"this one does tables" — and it had **no word for pictures at all**. So the illustrated section and
the plain text section were described to the planner in identical terms, and it picked the plain one
essentially every time: 208 plain ones across 23 sites, against 6 illustrated ones on a single site.

So the fix was to teach it the word. That part was small, and it is done.

**Then I read the part nobody had read, and it changed the job.** The illustrated section asks for
its picture from the site's own image library. I went to check *which* picture it actually gets, and
the answer is: **the one already at the top of the page.** There is a translation table buried in
another part of the system that quietly turns the request for "an image" into a request for "the
page's headline image". It always has. Nothing in the section's own definition says so.

You can see the result on pages we have already built. On one consultancy site's About page, the big
picture at the top and the picture inside the block below it are the same file. Across the estate, 20
of the 52 places this happens are showing a duplicate of something already on the page.

**So shipping just the small fix would have made things worse, not better.** We would have taught the
planner to reach for the illustrated section, and it would have filled it with the same photograph
that was already at the top. That is not what you asked for and it looks careless.

I put that to you and you said to fix the source as well before shipping. That is what I did.

**It also turned out to be about to destroy something.** One site — apis.uk, the bee one — has six of
these illustrated sections with six genuinely different pictures, hand-placed. Because of the
translation table, the next time that page was rebuilt, all six would have been replaced with six
copies of the page's top image. I flagged it to the person working on that site, and they came back
saying that page was already queued for a rebuild imminently. So the fix landed a few hours ahead of
the thing that would have wiped it. That was luck as much as judgement.

**One smaller thing worth knowing:** the descriptive text that screen readers use for these pictures
was being filled in with the image's file path. So a blind visitor would have heard a web address
read out instead of a description. That is fixed too — it now gets written properly.

**What I have NOT fixed, and it is the bigger half.** Making the section available does not create any
pictures. The system is good at generating *headline* images — 206 of them across 28 sites — and
almost never generates the kind that sits inside an article: 26 across 5 sites. So on most sites this
new capability will correctly and quietly do nothing, because there is nothing to put in it.

That is the real answer to "why don't the pages have pictures in them". It was never mainly that the
system couldn't place them. It is that it isn't making them. **Nobody currently owns that**, and it is
a bigger piece of work than today's. It is the thing I would point you at next.

**One judgement call I made that you might want to overrule.** In deciding which section types count
as "can show a picture", I included the headline/banner ones and excluded only the logo. A reviewer
could reasonably say the banners should be excluded too. It is in the change under review, so it can
still be argued.

---

**2026-08-31, evening.** The benefit-priorities work you approved this morning had a hole in it, and
tonight it is closed — or rather, the closing is built and waiting for the next deployment.

Here is the plain version. This morning a repair ran across the estate and fixed 41 benefit
statements that used the phrasings you rejected. That worked. But the thing that WRITES those
statements never learned anything, and it kept writing. By late afternoon it had produced 75 fresh
statements and 18 of them carried the same phrasings — the same one-in-four rate as before the
repair. Fifteen are live right now across eight sites, and two of them are the FIRST statement on
their site: mortgagecalculator.co.uk, written at 16:25, and finetuning.uk, written at 14:54 — the
site you rejected in the first place. The first statement is the one a headline gets written from.

So the repair was never going to win. You cannot clean a well faster than it refills. The fix has to
be at the point where the statements are written, and that is what I built.

**You chose the behaviour and it was the right call.** When it catches a bad statement it now asks
for it to be rewritten, rather than deleting it or throwing away the whole analysis. The two
alternatives both destroyed something real: deleting loses about one benefit in four, some of them
the site's best one, and refusing the analysis outright leaves the site sitting on its OLD
statements, which are just as bad.

**The part worth telling you about is where I got it wrong**, because it is the same shape as the
mistake we keep making. Every rewrite is vetted before it is accepted. I built the vetting rule out
of this morning's numbers: those 41 repairs shortened the statements by 29% on average, and ten had
to be rescued by hand, so I set a rule that a rewrite may not cut a distinctive statement by more
than 40%. Then I tested it against the actual repair that caused the problem — the one where
"delivers systems in days, not months" became "delivers systems in days" — and my rule let it
straight through. That repair only removed twelve characters. The distinctiveness was in those
twelve characters: "in days" is a duration, "in days, not months" is a claim against the
competition.

The lesson is that I built a rule out of an AVERAGE and then expected it to catch an individual
case. Averages do not do that. The rule I have now is not a measurement at all, it is a
statement about the language: when a statement distinguishes you, the thing it distinguishes you
FROM is the point, so any repair that simply stops the sentence early has removed the point, however
few characters it took. That version cannot be fooled by arithmetic.

I only found it because I tested against the real sentence from the real site rather than one I made
up. A made-up example would have passed, I would have shipped, and the guard would have been blind
to exactly the case it exists for.

**Where this leaves things.** The code is written, reviewed by the council (verdict still coming in
as I write), and does nothing until the next deployment — deliberately, because switching it on
before the code ships would break the analysis outright. Two things I want to be straight about:
nothing has actually run through it yet, so I am not claiming it works, only that it is built and
tested; and it is deliberately strict, so it will sometimes refuse a rewrite and keep the original
bad statement rather than accept a repair that guts it. Every one of those refusals gets written
down, which is how we will find out whether the strictness is right.

**One thing I found on the way that is worth a moment.** The instructions for how our own workflows
are wired live in two places: a file in the repository, and the live database. For the piece I
needed, the file was out of date by one change — a safety check had been inserted since and nobody
updated the file. If I had trusted the file, my change would have quietly disconnected that safety
check while reporting success, and every test would still have passed. I caught it only because a
number did not add up. It is written into the traps file now, because that one will bite somebody
else.

---

## 2026-09-03, later — the question hierarchy landed, and it found something specific

The thing we built last week went live and has now run on eighteen sites. It works: for each site
it writes down the five or six questions a visitor arrives with, ranked, and then says which of our
selling points answers each one.

I had predicted, in writing, that most of those questions would come back with no answer — that our
sites talk about themselves and don't answer what people actually want to know. **That prediction
was wrong.** Ninety-two per cent came back answered.

So rather than repeat the number I read the pairs by hand, thirty-six of them across seven sites:
the question, next to the words of the point claimed to answer it. Most of the joins are honest. On
two sites in particular every question has a real, matching answer. Our sites are better at this
than I said they were.

**But there is one thing they are consistently bad at, and the tool found it cleanly.** Of the
questions about money — what does this cost, why would I pay when the alternative is free, what does
paying actually unlock — **five of seven have no answer anywhere in our copy.** Every other kind of
question comes back with a clean zero unanswered: trust, credibility, what you get, whether it's
worth coming back. It is only price.

And the model ranks the price question **last**, on average. For idea.uk — which sells a £29 report
and whose competitor is free AI — the question *"why would I pay £29 for this?"* is sitting at rank
five out of five. I don't accept that ordering, and I've said so to the lane that owns ordering.

That is a much more useful finding than the one I expected, and it is the first thing this machinery
has produced that changes what someone should write.

## The carousel switch, and the check that would have made a correct change look broken

You ruled that the card grids should default to carousels. Writing that turned out to be four lines;
the useful part was everything around it.

**The instruction I had left for whoever did this work was wrong, and would have caused them to undo
a change that had just worked.** I had written down a "control" — a number that must not move when
the switch is flipped, as proof nothing unexpected happened. It reads 2 on a page with the carousel
on and 2 on a page with it off, which looks exactly like a control should look. But the two 2s had
nothing to do with each other; they were unrelated bits of styling on two different pages, and the
number is one the carousel itself adds. So a correct flip moves it from 2 to 3, and my own note
would have told the next person that meant something was wrong.

I only caught it because the job came back to me and I read the template before running my own test.
It is corrected, and written up as a trap, because the general form is worth having: **a count added
up over a whole page is not a control unless you know what each occurrence actually is, and two
pages agreeing on a total is not agreement.**

## The review caught something real, and it was not in my change

The change went to the reviewer council. Nine of eleven seats approved; one blocked it, and it was
right to.

Its objection: my whole safety argument rested on a claim about how the system behaves, and **our own
traps file contains an entry saying the exact opposite about the exact same piece of code.** I had
never looked. If that entry were still true, this change could have overwritten someone's deliberate
choice rather than merely doing nothing.

It turns out the entry is out of date. It was written on 3 August; the code it describes was fixed on
14 August, eleven days later. I checked it in the live data too, not just the history — there are
eleven places on the live sites right now where someone's own wording has survived exactly where that
entry says it would have been wiped.

**The wider point is the one I've written into the file.** A trap note describes a defect, and a
defect is precisely the thing most likely to get fixed — so these notes go stale in the one way their
own advice can't catch, and they go on reading like a live warning indefinitely. This one was right
for eleven days and misleading for twenty. It has now cost one review round; left alone it would
eventually have licensed a genuinely wrong change.

One more seat objected that I had not read the existing notes attached to this component. Fair, I
hadn't — and doing so found something none of the eleven reviewers had spotted: **the component has
its own automated acceptance test, and one of its checks is "nothing overflows sideways".** A
carousel is a thing that overflows sideways. Reading the checker's code says it should still pass,
because it deliberately exempts anything you can scroll. But that test has never actually been run
against a carousel. So it is now written down as owed before this ships, rather than assumed.

## boxingonline — asked one thing, found another

The lane holding your first paid build asked me for a design decision on its colours, and whether the
site name should sit next to the logo.

**Neither question needed me, and on the second one I was wrong.** You had already ruled on both, on
2 September. On the colours I found your ruling and told them; on the header I searched for the wrong
words, concluded you had never ruled, and told them so — and they found your words within the hour in
a third lane's notes. That is the fourth time in a week I have announced that something doesn't exist
when it does, and every one of those was caught by somebody else. It is logged.

**What I did find, and it stands:** the logo shares no colour at all with the site. The site is deep
red, gold and near-black. The logo is 52% blue and 45% grey, with exactly one red pixel out of
sixteen thousand — nothing anywhere near either of the site's own colours. It's a raised fist inside
a diamond, which is a protest symbol rather than a boxing one, and at the size it actually appears in
the header it's an unreadable smudge losing nearly half of itself against the dark background behind
it.

The other lane checked my numbers on a different copy of the file and got the same answer, then found
the better half: **the instruction we sent asked for "a stylised boxing glove or ring ropes".** So
this isn't a taste disagreement or a colour problem — the picture-making step was told the right
thing and ignored it. They're bringing it to you as a new question.
