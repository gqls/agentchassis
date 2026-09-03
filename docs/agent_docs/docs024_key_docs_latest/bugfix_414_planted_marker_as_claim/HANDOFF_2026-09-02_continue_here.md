> **⚠ SUPERSEDED 2026-09-03 — read `HANDOFF_2026-09-03_continue_here.md` instead.** This file is
> kept for the trajectory, not for state: the detector it describes as pending is now live and
> verified, and RFC_060's remaining questions have all been ruled.

# HANDOFF — 2026-09-02 — **the lane's code work is DONE and verified live. What remains is TWO OWNER DECISIONS and one piece of content work in other lanes.**

Supersedes `HANDOFF_2026-08-31_continue_here.md` (which closed `bugs_open/414` itself and is now
history — read it only for the 414 record).

Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_414_planted_marker_as_claim/`
Bug: `bugs_closed/414_HANDOFF_2026-08-26_…` — **CLOSED 2026-08-31** (`de99599fb`).
RFC: `docs/agent_docs/docs024_key_docs_latest/architecture_review/RFC_060_compliance_tier_…md` — **DRAFT, with the owner.**

**Counts carry the date they were counted.** Everything below re-measured 2026-09-02 ~16:2xZ against
chassis **`v1.0.1354`**, nothing inherited.

---

# 0. STATE IN ONE PARAGRAPH

414 closed on 2026-08-31. Since then this lane answered the owner's follow-up question — *"what can I
do about the poisoned register hole, and shouldn't compliance be strong for finance/legal/insurance?"*
— by **measuring** rather than designing, which moved the target: the poisoned-register hole is real
but narrow and mostly on **our own** sites, while the finance sites' actual exposure is the inverse —
**5 of 9 have no evidence register at all**, so the numeric claims scan never arms on them. Arming it
naively would have produced noise, measured: 5 findings, all false. **The regulatory half of that is
fixed, council-APPROVED and live in the running chassis** (`fad209b92` + `ad4824e73`, council
`1dd3d298`). The rest is not code: two decisions sit with the owner in RFC_060, and the highest-value
next step needs no RFC and no code at all.

---

# 1. THE DECISIONS THE OWNER NEEDS TO MAKE

Two, and only the first blocks anything.

### DECISION A — confirm the tier's vocabulary (RFC_060 §3a)

You chose a **named tier** on a **semantic** axis rather than a sector one. I agreed, and proposed a
three-rung ladder, each rung adding exactly one enforcement:

| rung | means | requires |
|---|---|---|
| `standard` (absent — default) | claims are about the site's own offering | today's behaviour |
| `sourced` | the site asserts **external** facts, rules or figures as true | a scannable register |
| `relied_upon` | a reader may act to their **financial, legal, medical or safety detriment** | the above, plus a fact-quality floor and raised practice severity |

**What you are actually deciding:** whether those three names and boundaries are right. A ladder was
chosen over independent booleans because a site like lendzy is *simultaneously* "quotes external
rules" and "readers act on it", and a single name has to hold both. If you would rather have two
booleans, say so — it is a cleaner model and a worse label.

### DECISION B — is the drift risk worth a Phase 2? (RFC_060 §3a, last block)

A declared posture is a **judgement, and judgements drift**: a site that starts `standard` and grows a
rates table is now `relied_upon` and nothing notices. The record (who declared, when, why) is a weak
mitigation. The honest instrument is a **mismatch detector** — a `standard` site whose copy quotes
rulebook citations or regulatory figures is a candidate for re-posturing — which is cheap on machinery
that already exists. **Not proposed for now**; recorded so it is a decision rather than an oversight.
You can defer this one indefinitely without blocking anything.

### Already decided, recorded in RFC_060 §3a — no action needed

- **Q1** — requiring registers is the right instrument, with the content work leading.
- **Q2** — order: register-required, then fact-quality floor, then severity.
- **Q4** — the tier is a **record** (who declared it, when, on what basis), not a flag.

---

# 2. WHAT IS LEFT ON THIS LANE

**Nothing in code.** Three items, none of which this lane can close alone:

| # | item | owner | blocking? |
|---|---|---|---|
| 1 | ~~Populate registers for the 5 register-less finance sites~~ **HANDED OVER 2026-09-02 — see §2a** | the **site lanes** + `claims_verification` | no — off this lane |
| 2 | RFC_060 Decisions A and B | the **owner** | A blocks any build of the tier |
| 3 | If A is answered yes → a **new `compliance_tier` delivery lane** | not yet created | — |

### 2a. The hand-over, 2026-09-02 — what was sent and what came out of it

Four messages, each carrying the measurement that makes the ask cheap rather than a chore:

- **`lendzy`, `loanzy.uk`, `loancalculator`** (live lanes) — each told: your site has no register, so the
  numeric check has never armed on it; **and the cost is near zero, measured — your site asserts ZERO
  unbacked business numbers today**, so this is switching a check on, not auditing copy. The value is
  in `banned_claims`, not facts, and `adversecreditmortgage.co.uk` already carries a well-argued
  six-pattern consumer-credit set to adapt rather than reinvent. Each was warned off the two failure
  modes: do **not** add `attested_by`-only facts (that is the hole `RFC_060` exists to close), and do
  **not** arm and walk away.
- **`claims_verification`** (the register lane) — given the fleet picture and **the two orphans I
  cannot route**: `loancash.co.uk` (lane dirs exist, no live session) and **`farmerinsurance.uk` (no
  lane at all** — and the only one of the five with live findings: 3 unbacked numbers, all third-party
  survey figures in a news listing, which is `RFC_053`'s shape and which a register will not and
  should not silence).

### 2b. OUTCOME, measured 2026-09-02 ~17:0xZ — the hand-over worked, and it moved fast

Re-measured at the DB rather than taken from the lanes' reports:

| site | register | state |
|---|---|---|
| `loancalculator.co.uk` | **12 facts** | live today 15:30 |
| `loanzy.uk` | **3 facts** | live today 15:27 (migration 697) + `banned_claims` via 702 |
| `farmerinsurance.uk` | **7 facts** | live today 15:27 (migration 698) — **no longer an orphan**; my "no lane at all" was stale within hours |
| `lendzy.co.uk` | **none live** | migration 695 written, 8 citation facts, council round 2 — **killed twice by today's rolling chassis deploys**; resubmits when the fleet calms. Written and reviewed, NOT applied |
| `loancash.co.uk` | **none** | the genuine remaining gap; the owner has been informed |

The lanes went further than the ask: **full citation-backed registers**, not the banned-claims-only
minimum I proposed, verified against the actual FCA Handbook / legislation.gov.uk text — and in doing
so found **5 wrong live claims in one day** across three sites (two mis-attributed CONC rules, MaPS
mis-described as FCA-authorised, a wrong settlement period, an invented ERC threshold).

### 2c. ⚠ A STRUCTURAL FINDING THAT CAME OUT OF THE HAND-OVER, and it is the useful part

I pointed three lanes at `adversecreditmortgage.co.uk`'s six-pattern set. **Both lanes that adopted it
diverged on the same two of six, independently, each with a measurement:**

- `\bno (credit )?checks?\b` — lendzy: adjacent UI text ("Yes"/"No" buttons + a "Check for a breach"
  label) concatenates in the visible body to *"No Check"*. loanzy: its calculator truthfully says
  *"There's no credit check involved"* **about its own tool**. Two sites, two unrelated false-positive
  routes, same arm.
- `\b[0-9]+(\.[0-9]+)?% (apr|apcr|rate)\b` — **both omitted it.** lendzy measured 3 hits, every one
  the credit-union cap quoted beside its named rule. loanzy: *"on your sibling a literal rate is a
  price promotion; here it is pedagogy."*

**And the rate arm has a structural problem, verified in the code:** `fad209b92` exempts a regulatory
figure quoted beside its named rule from the **number** scan — but **`scanBannedClaims` consults no
regulatory-citation exemption at all** (my exclusions sit only in `isExcludedNumber` and
`ScanUnregisteredNumbers`). So a curated sector set containing a rate pattern would **re-convict, at
BLOCKER severity in the refusing union, exactly the content the number scan now exempts at error
severity** — on the sites whose purpose is quoting capped rates beside their rules. The two layers
would disagree about the same sentence and the stricter one cannot see the citation.

**So, for RFC_060: if a sector set is ever curated, the citation exemption must reach the
banned-claims layer FIRST, or such a set cannot safely contain a figure pattern.** Sent to
`claims_verification` rather than edited into the RFC, at their request. Not proposed as work here —
it widens a refusing union, which is an owner/RFC decision, not a side effect of a number-scan fix.

**A design question surfaced by doing this, and it belongs in RFC_060 rather than here:** five sites
are about to hand-roll near-identical consumer-credit banned-claim patterns. `globalBannedClaims` is
fleet-wide and these are **sector-wide** — there is no middle tier today, and copying is how five
copies drift. RFC_060 currently defers exactly this as "sector's one real job, additive, later". If
`claims_verification` has a view, that is the file to put it in.

**On (3): do not start a lane yet.** RFC_060 lives in `architecture_review/`, which is where the
decision belongs and is active. A delivery lane is only worth creating once Decision A is made,
because the work spans the claims layer, a new spec field and a build-gate change and no existing lane
owns that span — `claims_verification` is a register/docs lane (quiet since 08-24) and
`copy_quality_two_stage` owns copy quality, not the severity model.

**So: this lane is CLOSEABLE NOW.** (1) is handed over (§2a), RFC_060 is with the owner, and the
lane's own code work is finished, council-approved and verified live. It does not need to stay open
waiting for a decision it does not own or content work it cannot do. **The only reason to reopen it
is if RFC_060's Decision A comes back yes** — and that starts a `compliance_tier` lane, not this one.

---

# 3. WHAT WAS DONE SINCE THE LAST HANDOFF, AND THE EVIDENCE

| | evidence (2026-09-02) |
|---|---|
| **Regulatory citations no longer read as business numbers** | `fad209b92`; council `1dd3d298` **APPROVED**, 3 advisories, none high |
| **…and it is LIVE** | chassis **`v1.0.1354`**, four-arm binary probe: `CONC\|MCOB\|ICOBS` PRESENT, `everything (on this site` PRESENT, a pre-existing pattern PRESENT, a never-written string ABSENT |
| **The three advisories answered** | `ad4824e73` — all three asked the same thing: does the documented forced `(?i)` on banned-claims reach these patterns? **It does not**, verified three ways (compile path, no case-folding on the scan path, and behaviourally) and now written at the claim, with `TestRegulatoryCitationPatternsAreCaseSensitive` as the tripwire |
| **Effect, measured** | the 5 register-less finance sites: 5 findings → **3** with the scan armed; the 3 remaining are third-party survey figures in a news listing (`RFC_053`'s component-grain question, deliberately not fixed here). Fleet-wide over 2,715 components: **unchanged** |
| **The spec detector's demand control is deployed** | the live report now reads `0 of 39 sites … from 7198 scanned field(s)` — the zero proves itself |
| **RFC_060 filed and owner decisions recorded** | `32fdc4840`, `29cda9e55`, `f2e8fe5c0` |

---

# 4. TRAPS — read before touching any of this

- **⚠ THE REPORT HAS TWO `N of M sites` LINES AND THEY LEGITIMATELY DIFFER.** The negation half
  (`bugs_open/305`) reads the **writer-visible** surface — `11 of 34` today. The spec-claims half
  (`bugs_closed/414`, CLM-030) reads the **union of every live agent's** surface — `0 of 39`. Always
  name which. My 08-31 handoff said "read N of M" without saying which, and the
  `copy_quality_two_stage` lane hit it; that handoff now carries the correction in place.
- **`kubectl logs` on `agent-chassis` returns ZERO lines**, so the build-provenance stamp route is
  unavailable. Use the four-arm binary probe (a known-new string, a known-pre-existing string, and a
  never-written string that must be ABSENT). An empty grep there means *out of range*, never
  *unstamped*.
- **The rulebook code list is CASE-SENSITIVE and that is load-bearing** — it is the only reason short
  codes (`SUP`, `MAR`, `DISP`) can be in the list at all. Three council seats asked whether the forced
  `(?i)` on banned claims reaches it; it does not, and
  `TestRegulatoryCitationPatternsAreCaseSensitive` will go red if a refactor ever routes these through
  a shared compiler.
- **The window test requires a code FOLLOWED BY A DIGIT.** An earlier draft matched bare `FCA`/`PRA`,
  which on a consumer-credit site appears in nearly every paragraph and would have disabled the
  numeric scan for the entire sector. A code plus a digit is a citation; "the FCA" is a subject.
- **Do NOT arm the numeric scan on a finance site before its register exists.** Armed against an empty
  fact list it convicts every business-context figure, and `claims.go`'s own standard is that "a
  scanner that always reports something is one people stop reading". Order: fix the sector's
  false-positive shapes → populate the register → then the check means something.
- **Do NOT key a compliance control off `site_archetype`.** Its `constraints` array already reads like
  a semantic compliance layer (`"Never make calculators appear to give regulated financial advice"`)
  — but it is **agent-written prose**, and a guarantee conditional on an LLM classifier inherits that
  classifier's gaps. It is evidence the semantic axis is natural, not a foundation to enforce on.
- **To size an exposure without shipping anything**: export the components pod-side (RUNBOOK §4,
  counted three ways) and run `cmd/claimscan -evidence <a facts-free register>`. That arms the numeric
  scan locally and is how the 5-findings-all-false result was obtained.

---

# 5. THE STANDING FIVE

- `PLAN_2026-08-27_planted_marker_as_claim.md` — the six decisions and their reasons.
- `RUNBOOK_planted_marker_as_claim.md` — the retraction sweep, the guarded spec strip, the corpus
  export that took three attempts, the council schema, the artefact proof.
- `NOTES_planted_marker_as_claim.md` — the technical log, every misstep, newest at the bottom.
- `README_where_we_are.md` — the owner's plain-prose log. **Append; never rewrite.**
- `SUMMARY_2026-08-31_closed_the_instruction_is_read_too.md` — the closing read-out. A third summary is
  owed only if the tier is built.

**And the file to read before repeating any figure from this lane:** eight of its own wrong claims are
logged in `WRONG_CALLS.md` (2026-08-27 → 09-02). They share one shape — each was a number or an
assertion made *in passing*, to support a point that was not itself about it. The measured work held;
the prose around it did not, and prose is what travels.
