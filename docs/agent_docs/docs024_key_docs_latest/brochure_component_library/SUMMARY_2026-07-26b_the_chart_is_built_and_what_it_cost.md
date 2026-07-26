# Summary — 2026-07-26b: the chart component is built and live, and the build paid for itself in findings

*Read-out for the owner, written to be read aloud. Previous:
`SUMMARY_2026-07-26_link_sound_verified_and_the_platform_findings.md`, which ended
with the chart as the sharpest remaining gap in the brief and two decisions
outstanding. This one differs because the chart now exists, one of those decisions
was taken, and the build turned up four things about the platform that matter more
than the component does.*

## What we're trying to do

Build genuinely good consultancy-brochure components — the interactive kind the big
firms use — **through the framework**, so any site we run can have them. And run
fundamentallyai.com as a site that markets this platform's real, verified
capabilities, where every claim has a source.

## Where we've come from

The brief has asked from the start for numbers rendered as real, code-generated
charts from true figures, never a picture of a chart. Nothing in the fleet could do
that: **zero chart components existed**, and the one thing that renders verified
numbers renders them as text. A design had been scoped twice — once in the
leopardess plan, once as a feature spec — and built neither time. Yesterday it was
recorded as the top task, with terms: one shared component, values from the
verified-facts register, code-rendered, and coordinate rather than start fresh.

## What we've done

**Built it, and it is live on the home page.** Three charts: a dormant site's legacy
feed going from nothing working to about 97% working within a day of relaunch; what
the news pipeline collects against what carries a credibility assessment; and — the
owner's call, and he said yes — what our own review council actually decides, 108
rounds sent back against 37 approved and 9 rejected outright. Verified on the
served page, not the database: every figure matches its fact row, and each bar is
drawn from that same number.

**Made the guarantee structural rather than instructed.** A chart definition names
which facts to draw and never repeats a value, so the only place a figure exists is
the register. The framework resolves those values itself and overwrites anything
the model wrote. There is no route by which a model-supplied number can reach a
bar — not a rule we are trusting it to follow.

**It needed no code release.** That was the surprise. The framework could already
feed a component straight from the verified-facts register; nobody had used it that
way. So the component went live the day it was written instead of waiting on the
next image build, and it needed no code review, because it is configuration.

**Leopardess gets it free**, which was the point of the "one shared component"
ruling. Their side now needs only a short list of which of their existing facts to
plot. Recorded in their own documents, with two corrections: the charting library
their plan works around is not in our codebase, and a chart-drawing routine written
for another job already is.

## Where we are now

The home page carries the charts. The capabilities page does not, and that is the
one part of the owner's answer we could not deliver: **nothing tells a section
which page it is being drawn on**, so every chart appears on every page carrying
the section. Rather than publish the same three charts twice, the section is on the
home page only. The "which page" information is already in the data, correct and
unused; a one-line fix at the next image build switches it on.

Three other things the build turned up, none of which were visible from reading
code:

- **A near miss on the dark site.** The chart card used a colour variable that
  reads as standard and that no theme actually defines — on leopardess it would
  have been a white card on a black page. Two existing components have been
  relying on the same non-existent names for weeks.
- **Rebuilding a page to add the chart reintroduced six broken links** on a page
  crawled clean the day before, four of them to pages that do not exist. The gate
  saw all six and deployed anyway, because they are classed as warnings. The
  repair we made yesterday was per-page, so the rebuild simply undid it.
- **A documented command in our own runbook fails every time it is used as
  documented** — the way to re-render a page after a template change omits a field
  the platform requires. Another session hit the identical failure on a different
  site an hour earlier. Corrected, and handed to the workstream that owns it.

**The em-dash question is finally measured**, because these were the first pages
written since the third attempt at fixing it: down by nearly half on one page,
unchanged on the other, with two components producing most of what remains.

## Where we're going

**Nothing here is blocked, and two things want a decision.** The voice fix half
worked; rather than a fourth attempt at the site-wide instruction, the cheaper
routes are a mechanical pass over finished text or fixing the two specific
components. And the decision-record page is still the owner's editorial call —
though the council chart now puts a version of that argument on the site already,
in aggregate, which may be enough.

Next in the agreed order: the component-adoption check, then the specialist design
critic, then sweep enrolment. Ahead of all three sits a one-line platform fix that
gives every component the page it is on — small, and it unlocks per-page charts
plus anything else that wants to vary by page.
