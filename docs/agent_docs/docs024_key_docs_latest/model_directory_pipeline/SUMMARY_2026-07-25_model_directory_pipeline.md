# SUMMARY — model directory pipeline — 2026-07-25

*(Series: second entry, after SUMMARY_2026-07-24. New file per the standing
rule; the previous one records what we believed at the end-to-end-proof
milestone, and parts of it turned out wrong in ways worth keeping.)*

## What we're trying to do

Give any site in the fleet, by a single opt-in flag, a continuously-updated,
citation-verified directory of AI models — every fact carrying the exact
sentence from its source, re-checked on a schedule — and then, on the same
machinery, a tracker of which real organisations are deploying AI agents,
with what claimed results and how those results were measured, plus the
protocols (MCP and its rivals) the agent world is standardising on. Nothing
publishable unless a machine re-fetched the source and found the quote.

## Where we've come from

Yesterday's summary recorded the model directory proven end to end on
ai-agent-orchestration.com: register → researcher → verification → page →
published data file, with 27 models across seven owners. It also recorded
two beliefs that did not survive today: that a homepage teaser section was
live (it had failed to build, three times, and the failure was hidden behind
a parent item's green status), and that every claim was re-checked daily (the
re-checking job had never run — see below).

## What we've done (today)

- **Shipped the second half of the brief with no deployment.** The adoption
  tracker's research lane went live as pure configuration, because the
  claim-verification core was built kind-agnostic in Phase A. The register
  now holds **16 companies** (Klarna, JPMorganChase, Uber, Siemens, DHL, BNY,
  Deutsche Telekom, Swisscom, Shopify, GitHub, OpenTable, Port Newark) and
  **4 protocols** (MCP, A2A, ACP, ANP) alongside the 27 models. The honesty
  discipline is visible in the rows: a claimed result and its measurement
  method are separate facts, so "stated, method not given" is recorded as
  exactly that.
- **Generalised the read/publish/discovery legs** from model-only to
  kind-driven (one profile table per leg, not three copies of three files).
  Council-approved first round, ~13 minutes, four advisory objections — every
  one a risk I had flagged myself and left unguarded; all four now closed
  with tests, and one council seat's demand for a citation exposed that my
  justification was right but by the wrong route. Inert until an image roll.
- **Found that the freshness sweep had never run — not once** (bugs 074).
  A scheduled task carrying its workflow inline targets an agent whose own
  workflow is a no-op; the task fired daily, reported success, and did
  nothing, invisibly, because the broken path is identical to the healthy one
  at every observable except the work itself. Found only by deliberately
  corrupting a stored quote and watching for a rejection that never came.
  Fixed the same day, and proven by running the sabotage again: the repaired
  sweep caught it first time. The same wiring exists in another workstream's
  evidence sweep; written up for them, not touched.
- **Put the directory in the site's navigation**, which it had never been in:
  the owner ruled Model Directory into the header, Pricing down to the
  footer's first slot. The nav tables are rebuilt; the new header is
  propagating through ~35 page re-renders. Ranked directory pages as
  navigation hubs in code so the promotion cannot silently undo itself.
- **Fixed the page rendering bare** (bugs 072): the directory components
  shipped markup with no CSS anywhere. They now carry their own styles, live
  and verified on the page.
- **Filed bugs 073**: the homepage cannot currently be rebuilt at all,
  because the anti-fabrication rule (correctly) returns an empty statistic
  and a component (incorrectly) demands one. Two right mechanisms, deadlocked.
- **Had a diagnosis refuted, usefully**: the 090 loop disproved my claim that
  nav repair "requires a human" — no human gate exists; what stalls is the
  detection-to-triage promotion, off by deliberate owner ruling.

## Where we are now

Both registers are live and verified: 47 entities, 82 current found claims,
all citation-checked at registration, and — for the first time actually true —
re-checked on schedule by a sweep that has been proven against an induced
fault, in both directions. The model directory is a styled, populated page
about to appear in the site header. The adoption tracker has real data and no
visible surface: its components, resolvers and publish profiles are written,
approved and waiting on an image roll. The homepage teaser stays blocked on
bugs 073, which belongs to the fabricated-stats lane.

## Where we're going

The next image roll activates the adoption tracker's publish and page legs;
its pilot then follows the model directory's proven path (opt-in flag →
discovery → page → publish). The weekly sweeps take over acquisition for both
registers. Open questions for the owner are unchanged — whether finetuning.uk
opts in, and pricing-claim staleness cadence — plus one new one from 073:
whether case-study cards should show no metric rather than demand one.
