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
