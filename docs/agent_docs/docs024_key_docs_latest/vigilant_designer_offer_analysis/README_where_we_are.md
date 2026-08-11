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
