# Roadmap brief for `agritec.uk` — Phase 1

Drafted 2026-08-21. This is the `roadmap_brief` text for the Tier-3 submission. The planner
treats it as **authoritative**: it builds only the pages named here and does not invent additional
ones.

**No figures appear here, for the same reason as the mission brief.**

One deliberate exception, stated so it is not mistaken for an oversight: the brief says *six*
explainers. A count of pages we are instructing the planner to build is a **build instruction**,
not a claim about the world — it cannot become a false statement on a page about lighting or
scheme payments, which is the leak the no-figures rule exists to stop. Counts describing the site
being replaced were removed on the same test and live in `SUBJECT_LEDGER.md` instead, where they
are evidence rather than instruction.

---

Phase 1 rebuilds the agronomy and controlled-environment half of the site. It is a replacement for
a site that already exists, so the subjects are known rather than speculative — but the writing is
new throughout, and the standard is higher than what it replaces.

## The pages to build

**index** — the home page. What agritec.uk is, who it is for, and what it does that the trade
press and the equipment vendors do not: it takes the mechanism apart and lets the reader test it
against their own operation. Point clearly at both halves — the calculators and the explainers —
and make the relationship between them explicit, because each is much less useful alone. Do not
list achievements, user counts, or any measure of scale: this site has none, and omitting them is
correct rather than modest.

**about** — what this publication is and the standard it holds itself to. The genuinely useful
content here is the method: every figure traces to a document we can name and date; where a figure
is unverified we say so rather than estimating; a calculator models the scenario the reader
described and does not know their land or their agreement. State plainly that this is an
independent technical publication, and make no claim about staff, history, clients, accreditation
or credentials.

**tools** — the index of calculators. It must list **every** calculator on the site, without
exception. This is not a formality: on the site being replaced, several calculators are
missing from this page and are reachable only from the home page. Build this page with an explicit
list section that resolves against the tool pages themselves, so the listing cannot drift away
from what exists.

**guides** — the index of explainers. The same requirement, and the same reason: on the site being
replaced, one explainer is reachable from no index at all. This page must list every explainer,
through a list section rather than hand-written links.

**Six explainers.** Each is a standalone technical article, not an appendix to its calculator. Each
should teach the mechanism well enough that a reader who never opens the calculator still
understands the subject, and should include the governing relationship expressed properly as an
equation, at least one diagram or infographic doing real explanatory work, and — where the subject
has one — a comparison table. Every figure in any of them must come from the evidence register.
Where the register has nothing yet, write the mechanism and say plainly that the quantity is not
yet verified.

- **The physics of horticultural lighting.** Why units built for the human eye misdescribe what a
  plant receives; the distinction between instantaneous photon flux and the daily integral; how
  the two relate through the photoperiod; and why fixture efficacy imposes a cost twice over, once
  at the meter and again at the cooling plant.
- **Vapour pressure deficit and transpiration.** Why relative humidity alone is the wrong control
  variable; what deficit actually measures and why it drives transpiration; how leaf temperature
  departs from air temperature and why that offset changes the answer.
- **Hydroponic solution chemistry.** Why parts-per-million is a fiction that meters cannot
  measure; the relationship between electrical conductivity and dissolved solids and why the
  conversion scale matters; and the reason concentrated stock must be split between tanks rather
  than mixed.
- **Insect bioconversion.** The larva as a bioreactor; what bioconversion rate means and what it
  does not; frass as a second product rather than a waste stream; and why metabolic heat, not
  substrate, sets the ceiling on rearing density.
- **Seaweed and the carbon question.** The distinction between carbon cycled and carbon
  sequestered, which is where most claims in this area fail; the stoichiometry that connects
  harvested mass to carbon; and how end use determines which of the two has occurred.
- **Stacking agricultural scheme actions.** How England moved from area payments to paying for
  outcomes; how scheme actions layer on the same land and where they conflict; and what
  determines whether a given action is compatible with a holding's existing obligations. This one
  must be especially careful: it explains how the mechanism works and never tells a reader what
  they are entitled to.

## Sourcing, and how it must appear on the page

Every figure carries a visible link to its source. This is a build requirement, not a stylistic
preference, and it applies to prose, tables and calculators alike.

- **The link is an HTML anchor**, inside a field that carries markup. It is never written as
  markdown: nothing in this platform renders markdown, and the literal-markdown check treats a
  bracket-and-parenthesis link in a text field as a defect to strip — so a markdown citation is
  silently deleted and the figure is left looking sourced when it is not.
- **Name the publisher and the capture date** next to the figure or in the link text. Every fact
  in the evidence register already carries both, along with the URL.
- **A calculator's rate table is the natural place for this**, one link per rate. It is also the
  only way a figure inside a calculator becomes checkable at all: the publishing checks read the
  visible words on a page and never the code behind them.
- Where a figure is not verified, say so in place of citing it. An unsourced number is not
  published on this site in any form.

State the posture honestly and do not overclaim on it. A citation shows where a number came from;
it does not show that we read the source correctly or that the source is right. The site cites
everything so a reader can check it — not so the reader can stop checking.

**contact** — a plain contact page. No form fields beyond what is genuinely handled, and no
promises about response times.

## What is not built by the planner

**The six calculators are not planner pages.** They arrive through the tool pipeline, which
creates the tool page, its nav declaration and its cross-links itself. The planner must not invent
tool pages, and must not write placeholder pages standing in for them.

**Legal pages** — privacy and terms — are hand-written separately and are not part of the
generated build.

**There is no news feed in Phase 1.** The classifier will read this site as agriculture and will
want to seed generic farming-news keywords. That is the wrong instinct for a calculator-led
technical publication, it spends credits on every fetch, and it must be turned off deliberately
after classification rather than left on by default.

There is no pricing page, no subscription offer, no paid product and no sign-up. Nothing on this
site may offer or imply a purchase, because there is nothing to sell.

## A known wrinkle to reconcile, not to design around

The tool pipeline automatically creates a companion guide page for every tool, titled in the form
"Understanding *the tool's name*". That is a tool-centred framing, and it is a weaker artefact
than the six subject-led explainers above — "The physics of horticultural lighting" is a better
page than "Understanding the Vertical Energy Estimator", and it is the one that stands on its own.

The two do not collide, because the names differ — which means the site would end up with a
redundant stub alongside every real explainer. Reconcile this deliberately after the tools are built:
either retire the auto-created stubs, or repoint each tool's call-to-action at the real explainer
and let the stub go. Do not solve it by writing the explainers as companion guides.

## Later phases — direction only, explicitly not to be built now

The machine-vision and edge-computing half of the existing site: the engineering series on
distributed optical crop monitoring, and its supporting calculators. Everything on the site being
replaced migrates eventually; the ledger in this directory tracks it, and the ledger is the place
for counts.

After that, and only once the material above is genuinely good and current: a news capability
aimed properly at this vertical rather than at agriculture in general, editorial features
combining sourced series with analysis, and a directory of suppliers or equipment — which is
new-vertical work rather than a switch to flip, and should be planned as such.
