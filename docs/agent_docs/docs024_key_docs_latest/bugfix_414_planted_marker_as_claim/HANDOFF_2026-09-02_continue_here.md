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
| 1 | **Populate registers for the 5 register-less finance sites** — `lendzy.co.uk`, `loanzy.uk`, `loancash.co.uk`, `loancalculator.co.uk`, `farmerinsurance.uk` | the **site lanes** (content work, no RFC, no code) | no — but it is the **highest-value remaining action in the whole thread** |
| 2 | RFC_060 Decisions A and B | the **owner** | A blocks any build of the tier |
| 3 | If A is answered yes → a **new `compliance_tier` delivery lane** | not yet created | — |

**On (3): do not start a lane yet.** RFC_060 lives in `architecture_review/`, which is where the
decision belongs and is active. A delivery lane is only worth creating once Decision A is made,
because the work spans the claims layer, a new spec field and a build-gate change and no existing lane
owns that span — `claims_verification` is a register/docs lane (quiet since 08-24) and
`copy_quality_two_stage` owns copy quality, not the severity model.

**So: this lane can be closed once (1) is handed over and RFC_060 is with the owner.** Its own code
work is finished and verified. It does not need to stay open waiting for a decision it does not own.

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
