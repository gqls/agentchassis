# Decisions register — oufe.com / oxenunity.com, 2026-07-26

Everything currently waiting on the owner, plus what he has already settled, so
the open list can be read in one sitting. Companion to
`SUMMARY_2026-07-26_oufe.md` (current state) — this file is *choices*, not status.

---

## Part 1 — Already decided (for reference, no action)

| # | Decision | When |
|---|---|---|
| D1 | UK focus; Thames Water as the flagship case | Gemini thread |
| D2 | Audience is the mid-market professional, not the large funds | Gemini thread |
| D3 | Free first; deliverables later; subscriptions only once regular value is proved | Gemini thread |
| D4 | oxenunity.com makes **no entity claims at all** | 07-25 |
| D5 | First slice = docs + oxenunity live + oufe P1 skeleton | 07-25 |
| D6 | Drop "every figure links to its source" — say instead that **we cite everything so you can check us, and we can still be wrong** | 07-26 |
| D7 | Lead with mechanism; real cases are clearly-marked illustration — "a possibly inaccurate case study" | 07-26 |
| D8 | Tools must say they can give a wrong answer; acknowledgement is a condition of use | 07-26 |
| D9 | Disclaimer sections A–F approved | 07-26 |
| D10 | Paid products: liability capped at the refund | 07-26 (wording drafted, §G) |

---

## Part 2 — Open decisions

Ordered by what blocks the most work.

### O1 — The audience question ⟵ blocks content direction

The owner asked whether targeting **students** would be safer. Recommendation
(PLAN §7): **take the safety posture, do not narrow the audience.**

The reasoning in one line: the risk lives in asserting live facts about named
companies, the value lives in explaining mechanism, and those are separable — so
you get the safety from *how* you write, not from *who reads it*. Professionals
already expect "check this yourself"; narrowing to students removes the ability
to charge for anything, because students have no budget.

- **(a) Recommended** — audience is "anyone learning how this works": students,
  trainees, early-career professionals, adjacent practitioners. Keeps the paying
  reader, keeps the posture.
- **(b)** Students proper. Safest-feeling, but it is a decision to make the site
  free indefinitely — worth making deliberately rather than as a by-product.
- **(c)** Keep the mid-market professional as primary and adopt only the honesty
  posture.

**Cost of delay:** the mission and roadmap briefs need revising either way, and
pages are being written now. This is the cheapest it will ever be.

### O2 — Section G liability wording ⟵ new text, unread

Drafted on the owner's instruction but not yet seen by him. Three things in it he
should register: the cap is only as strong as the refund actually being honoured;
**it protects nothing on the free content, which is currently the whole site**;
and the "cannot be excluded" carve-out is what keeps the rest enforceable-looking.
The statutory footing is marked `[UNVERIFIED — for the solicitor]` deliberately —
it would be incoherent to refuse to state law from memory on the site and then do
it in the site's own terms.

### O3 — Solicitor review: before or after launch

Two precedents in-house, both defensible. idea.uk: a ~£200–500 fixed-fee UK review
(its own terms are still flagged as drafts pending one). vetcomparison: proceed
ahead of review on defined narrowing conditions, provided the decision and its
conditions are recorded contemporaneously.

Recommendation: **after** for the free site as it stands, **before** the first
paid sale — that is the point at which G starts to matter and at which a customer
relationship exists. Drifting into a sale without choosing is the bad outcome.

### O4 — Where the promise-keeping agents go ⟵ the owner's question, answered

*"Do we need to put agents into the committee or into the workflow or both?"*

**Both — and neither is the missing piece.** There are three moments where a
promise can go wrong, and we currently cover two:

| moment | control | state |
|---|---|---|
| **Generation** — the writer invents a promise | writer rules (mig 223) | **shipped** |
| **Review** — a proposed change introduces or weakens one | council compliance seat (mig 223) | **shipped** |
| **Live content** — what is deployed right now | *nothing* | **GAP** |

The gap is not theoretical: the promise we caught this week **was never reviewed
by any council** — site copy is not submitted to the gate — and it **pre-dated the
writer rule**. It was caught by a person reading the page. That is not a control:
it does not run on a schedule and it does not scale past the site someone happens
to be looking at.

So the first recommendation is a **sweep**: a discovery check that scans deployed
pages for reliability overclaims, in the same shape as the existing
`check_unverified_claims`. That is the layer that would have caught oufe without
me, and it is the cheapest of the three to build because the pattern exists.

**But the owner's phrasing points at something none of the three do.** He asked
how we ensure our promises are *met* — which is a different question from whether
we *overclaim*. Detecting a false promise and keeping a true one are not the same
job:

> "Tell us and we will correct it" needs a monitored inbox.
> "14-day refund" needs someone able to issue refunds.
> "We name the document behind each figure" needs the register to be populated.

Every one of those is a claim about our own future behaviour, and **nothing in the
estate checks that a promise has a mechanism behind it.** The platform already
understands this shape for buttons — a label implies a real destination, and a
button with no destination does not render. **A promise implies a real
mechanism**, and we have no equivalent.

- **(a) Recommended** — build the live-content sweep now (small, known pattern),
  and open a **promise register**: each outward promise recorded with the
  mechanism that keeps it and a check that the mechanism exists. Same idea as
  CTA integrity, one level up.
- **(b)** Sweep only. Catches overclaims, does not tell you whether a true promise
  is being kept.
- **(c)** Neither; rely on the two shipped layers and human reading. Honest
  position, but it is the position that already failed once this week.

### O5 — Say plainly that Oxen Unity is not a company?

oxenunity.com is resolved (it claims nothing). oufe.com is not: a publication
invites "who are you?", and the honest answer today is "an independent research
project, not an incorporated firm". Recommendation: **say exactly that** on the
about page and in the disclaimer. It costs a little authority and buys the thing
the site is actually selling.

### O6 — The radar ordering (still unanswered from 07-25)

The owner ruled "direction 3 first, it is lowest risk". This workstream argued the
opposite and built the dossier-plus-tool path instead (PLAN §C1): no market data
exists in the platform, UK dockets have no feed, and a distress signal is a
factual claim about a named real company. **We proceeded on our own reading**, so
this needs either ratifying or reversing rather than being left implicit.

### O7 — News feed: confirm it stays off

Deliberately disabled (PLAN §C7). The classifier reads the site as `finance` and
would seed generic market-news keywords — the opposite of a specialist
restructuring publication, and it spends credits per fetch. Recommendation:
**stays off**. Reversal trigger: a genuine restructuring vertical, which is a
fleet-wide Go change.

### O8 — Contact address

`oufe@contactforsales.com` is seeded and live in the footer. Confirm or replace.
Note it becomes load-bearing the moment F (correction and removal) publishes —
that promise needs a monitored inbox behind it, which is O4's point in miniature.

### O9 — What the first paid product actually is

Not urgent, but it shapes P2. The Gemini plan assumed deal packs at £400–£1,000.
§7.6 raises an alternative that fits the teaching posture better and needs no live
data: **training material for trainees and graduates inside law and advisory
firms**. `[UNMEASURED]` — no demand evidence, no pricing research, no
conversations. Flagged as a direction to test, deliberately not as a plan.

### O10 — Widen the finance vocabulary in the number scan?

The deterministic scan is near-inert on this site: no debt/creditor/recovery
vocabulary, and currency amounts excluded outright (PLAN §C2c). Widening it is a
**fleet-wide Go change** affecting every site, so it belongs in front of the
council. Recommendation: **defer** — the writer whitelist and banned patterns
carry oufe today, and a fleet change to serve one site is the wrong trade until a
second site needs it.

---

## Part 3 — One consequence for another workstream

`features_open/014` records that idea.uk's stages 6–9 (patents, copyright,
funding) are **hand-authored rather than generated**, explicitly because
claims-verification V5 was inert and `bugs_open/043` was live.

**Both of those conditions have now changed.** V5 completed end to end for the
first time today (14 citations verified from legislation.gov.uk), 043 is closed,
and there is now a lane purpose-built for exactly that content class — research
first, verbatim quotes, machine-verified, human-gated, cannot publish itself.

That does not mean the constraint should be lifted, and it is not this
workstream's call. But it was written against conditions that no longer hold, and
whoever owns 014 should know that.
