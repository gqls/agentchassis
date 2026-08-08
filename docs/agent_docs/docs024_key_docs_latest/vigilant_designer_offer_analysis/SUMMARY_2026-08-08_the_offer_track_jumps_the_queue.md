# SUMMARY 2026-08-08 — the offer track jumps the queue

*Read-out for the owner. Current state only; the chronology is in NOTES and
README_where_we_are. Previous: `SUMMARY_2026-08-04_phase0_proven_drain_works.md`.*

---

## What we're trying to do

Two standing faculties, built to the same rule.

A **vigilant designer** that looks at every site from the whole page down to the finest
detail, and an **offer and benefit analyser** that keeps asking a different question: does
this site actually answer its target market's need, in a way that pays us? Not "is it well
built" and not "is it true" — is the *offer* right, and is the benefit to the visitor
visible?

The rule both are built to, and the reason the programme exists at all: **every detector
ships with the thing that acts on it.** This platform is reasonably good at noticing and bad
at consuming. The worked case is a brief-fidelity audit that predicted the owner's own
complaints about a live site three days early and then sat unread in a queue nothing
consumes. Noticing more, without draining it, would make that worse rather than better.

## Where we've come from

The programme opened on 2 August with four owner decisions: manual per-site triggers for now,
**designer first and offer analyser second**, trial both Gemini and Claude as the critic's
eyes, and broad autonomy to apply fixes with two honest exceptions.

Through 3–4 August the plumbing went in and was proven, not assumed. The dishonest counter
that reported audited-out sites as "clean" was replaced with a real gate. The browser audit
learned to turn its measurements into work. Screenshot capture and the ability to send images
to a model — neither of which existed anywhere in the chassis before — were built, reviewed
and shipped. Then we watched a finding travel the whole way on a real site: twenty-two
findings promoted, a stale one closed by the system observing the problem had gone, a fresh
one worked through to a re-rendered header and a nineteen-page refresh that all deployed. The
counter bug closed on that run.

Then the first dispatch to the CSS-fixing agent produced **correct** colour fixes and
destroyed the stylesheet — the model returned only its new rule and nothing between it and
the two writers checked the size, so a live site served an unstyled page until it was
restored. That is `bugs_open/198`. The fix is in and council-approved: the save is now an
append, so shrinking is not expressible.

## What we've done since

Answered a direct question — how far has the offer analyser actually got — and then acted on
the answer.

The honest finding is **one approved plan and no code**. No offer agent, no offer checks, no
offer analysis has ever run. That was checked four ways rather than one, because a single
grep only ever proves absence for the spelling it searched.

But looking properly turned up four things that change the job:

- **The platform already asks the offer question, blind.** The strategic review inside every
  sweep asks "what single change would most improve conversion?" — with no strategy, no
  audience, no identity and no content direction in front of it. Sixteen times a fortnight.
- **The better strategy shape is the older one.** One site, written in April, records what
  would satisfy the visitor, what brings them back, how much trust the purchase needs and how
  the money flows. Sixteen sites carry a later shape without them. What the plan proposed as
  four new fields is a restoration with a live worked example.
- **Ten of seventeen sites are recorded as the consultancy shape** — the one the mission
  document names as a failure mode when the signal is absent. That does not mean ten are
  wrong; several genuinely are businesses. It means the revenue-shape check has a real
  population and a real question on day one, and that the boring answer was available and is
  not what came back.
- **Half of "put it on the council" is already true.** An always-on council seat enforces the
  revenue-shape doctrine on every platform change — but only on code. It will never look at a
  site.

One of our own claims in that review was wrong and is corrected in place: we said a field the
plan relies on had never been written, and called two correct lines of the plan defective. It
exists on sixteen of seventeen sites, one level down in the record. Reading the top level of
a nested structure and calling that a census is the error; it is logged with the check that
catches it.

## Where we are now

Programme A is two phases done and proven live, three not started, and **blocked on one
witnessed run** to close `bugs_open/198`. The taste critic — the thing the whole designer
track is for — has not been seeded.

Programme B is unstarted, and the owner has now **partially reversed the build order**: the
two cheapest pieces jump the queue, because neither depends on the designer work and each
fixes something already wrong today.

- **B1 — give the strategic review its own premise.** One query rewrite, so that a review
  which already runs stops running blind.
- **B2 — make a premise refresh safe.** Today, refreshing a site's strategy unconditionally
  files a briefing job, which files a re-plan job. On a greenfield build that chain is
  correct and it is the only way it has ever run. On a live site it would re-plan it as a
  side effect. That is why nobody has dared refresh one.

Both are configuration, so both are live the moment they apply — no image build, and no
window where they are written but inert.

The wider version the owner described — a full agent with its own checker and handler,
possibly a council seat, corresponding with copywriting, design, planning, imagery, tool
design, the experience loops and the spec — is now written up as
`features_open/030_FEATURE_offer_and_benefit_analyser.md`, which is where "designed but not
built" is supposed to live. Of the seven counterparts named there, two are wired today, three
are wired but fragile, and **two have no route at all** — tool design and the experience
loops. That pair is the real design work, and it was not in the plan.

## Where we're going

Immediately: B1 and B2, in that order, each verified by a planted marker reaching the
assembled prompt rather than by a query returning rows — the distinction that separates
"the config changed" from "the agent sees it".

Then the designer track resumes where it stopped: one CSS dispatch to discharge 198, then
seed the critic and run the Gemini-versus-Claude comparison that has been owed since 2 August.

Five questions are open for the owner and are listed in the feature file. The one that
shapes the most work is whether offer judgement belongs on a council at all — a reviewer
gates a proposal, an auditor judges a live artefact, and most of what was described is
auditor work.

One constraint holds across all of it and should be said plainly: **we have no outcome data.**
The analyser can grade a site against its own stated premise. It cannot grade either against
what visitors actually did. An offer analyser that sounded like it knew what converts, while
reading only our own specs, would be the most confidently wrong instrument we have built.
