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

### O4 — Where the promise-keeping agents go ⟵ REWRITTEN 2026-07-27 after research

> **CORRECTED.** The first version of this decision recommended building a
> live-content sweep and a new promise register, on the premise that this class
> was invisible to every check we own. **The premise was false** and the owner
> caught it by refusing the answer and telling me to look harder at what exists.
> The corrected recommendation is almost entirely reuse. Original reasoning left
> visible below the line; full entry in `WRONG_CALLS.md`.

**The answer to "committee or workflow or both" is: both, and they are already
built. What is missing is not a layer — it is REACH.**

| moment | control | state |
|---|---|---|
| Generation | writer rules (mig 223) | shipped, fleet-wide |
| Review of a proposed change | compliance seat (mig 223, corrected by 227) | shipped, both rosters |
| **Build gate** | `ScanBannedClaims` → severity **blocker** | **existed all along** |
| **Live deployed pages** | `check_unverified_claims` → **high** + HITL item | **existed all along** |

The last two are the discovery. The scanner is an unrestricted case-insensitive
regex over prose — it catches whatever patterns a site is given, numeric or not.
It would have caught all four oufe phrases. Nobody had ever written a pattern for
this class, on any site.

**So the work is coverage, in three steps, mostly config:**

1. **Arm the patterns.** DONE for oufe (mig 226): 10 patterns, tested both ways —
   10/10 fabrication shapes blocked including all four live phrases, 13/13
   legitimate sentences pass. One UPDATE, no image roll. The line they draw is
   worth keeping: **you may describe what you DO; you may not claim what that
   GUARANTEES.** "We cite every figure and date it" passes; "a claim without a
   source does not appear here" does not.
2. **Make it reach the fleet.** Today only **5 of 15 live sites** carry a single
   banned pattern — the ten without include **vetcomparison.uk**, the site of our
   worst fabrication incident, and **idea.uk**, which takes real money. There is
   no mechanism to define a set once. **This exact question is already filed and
   deferred**: `SPEC_claims_verification.md:250-252` asks whether `banned_claims`
   should be fleet-shareable and answers *"per-site only until two sites have
   evidence bases"*. That precondition lapsed at n=8. **The decision is due, not
   new.** The implementation is precedented one directory away:
   `globalTellPhrases()` (`voicetells.go:121-137`) is a hardcoded fleet list
   unioned with the per-site list at `:109` — mirror that into
   `ParseEvidenceBase`. Small, and it is how the sibling engine already works.
3. **Give the post-deploy sweep a cadence.** `check_unverified_claims` runs only
   under `quality-discovery-agent`, which runs only under `improvement-loop`,
   whose `improvement-sweep` task has been **disabled since 2026-05-02**
   (`bugs_open/083`). So the live-content check exists and effectively never
   runs. That is an owner decision, because re-enabling also re-arms fleet-wide
   auto-fixing.

**On the promise question specifically — also mostly reuse.** `evidence_base`
already defines `kind: metric | capability | entity | attestation` and
`source: sql | artifact | attested_by`, and V4 already re-runs sql-sourced facts
on a daily fleet sweep and raises `stale_evidence` on drift. **`Kind` is declared
in the struct and never read anywhere in the platform** — the slot for "a
capability claim, backed by a query, an artifact, or a named human's attestation"
was cut and never used. A promise is exactly that: `kind: capability`, source =
the mechanism that keeps it.

Two other things already exist and should not be reinvented: the EXPERIENCE_PLAN's
§2 is *literally called a promise ledger* (CTA copy → the state the destination
must deliver), and `check_contact_form_undeliverable` already refuses to accept a
synthesised `info@own-domain` as an inbox, on the stated ground that *"a mailto
nobody reads makes the form look repaired while still losing the message"* — the
nearest thing we have to a monitored-inbox check, stopping one step short by
design.

- **(a) Recommended** — steps 1–3 above; then extend `evidence_base` to consume
  `kind: capability` for promises rather than building a new register.
- **(b)** Steps 1 and 3 only; leave promises to human reading.
- **(c)** Step 1 only (per-site arming, forever). Cheapest, and leaves every new
  site unprotected by default — which is the state that produced this bug.

<details><summary>Original reasoning, superseded — kept because the error is the point</summary>

The first version asserted a three-moment table with "Live content — nothing —
GAP", concluded that a promise is invisible to every scanner, and proposed a new
sweep plus a new promise register. The sweep already exists (it just has no
cadence); the scanner already covers the class (it just has no patterns); and the
register has an unused slot for capability claims. The proposal was new machinery
standing in for unread source.

</details>

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
