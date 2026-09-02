# RFC_060 — a COMPLIANCE TIER: the claims layer is weakest exactly where the sector is strictest

**Status: OWNER-DECIDED 2026-09-02 on Q1, Q2 and Q4; Q3's AXIS decided (semantic, not sector) with the
vocabulary proposed below for confirmation. Nothing is built yet. See §3a.**

Filed 2026-09-02 by the `bugfix_414_planted_marker_as_claim` lane, out of the owner's question
*"what can I do about the poisoned register hole, and shouldn't compliance be strong for sites that
require compliance strongly — finance, legal, insurance?"* The instinct is right and the measurement
moves the target, so the answer is routed here rather than built: it changes what the claims layer
GUARANTEES for a class of sites, which is the 2026-07-29 §1 trigger.

**Every figure below carries the date it was counted** (owner ruling 2026-08-22) and every one is
re-derivable from the commands in §7.

---

## 1. Problem + evidence

### 1a. The hole the question started from is real, narrow, and NOT on the client sites

`bugs_closed/414` closed with a residual: **the evidence register is the one thing no check inspects,
because it is what every other check checks *against*.** A fabricated fact written into it would be
laundered into "verified" by every layer downstream.

Measured 2026-09-02 — **319 facts across 17 sites**, and most of the register already re-checks
itself daily:

| fact source | facts | re-checkable? |
|---|---|---|
| `citation` (alone or with others) | **~192** | yes — `evidence-refresher` re-fetches the URL and re-checks the verbatim quote **daily**, classifying honestly (200-but-quote-gone = drift; 403 = *unknown*, never drift) |
| `sql` | **30** | yes — the query is re-run daily |
| `artifact` | **27** | partially |
| **`attested_by` ONLY** | **61** | **no. A name and a date, with nothing to re-check** |

And the 61 are not where the risk narrative assumed: **50 of them are on our own sites** —
`webdesign.uk` 25, `webdesign.co.uk` 15, `finetuning.uk` 10. The client finance sites are barely in
this table at all, for the reason §1b gives.

One more measured fact that shapes the threat model: **no live agent writes `evidence_base`.** Of the
18 current rows, 11 were written by the scheduled `evidence-refresher` and 7 by human/session hands
(`manual`, `seed`, `owner-ruling`, `operator`). **The poisoning vector is a person or a session, not
an LLM** — which is exactly how 414's marker got in (`source='manual'`).

### 1b. The bigger exposure is the inverse: on the compliance sites the check is OFF

`ScanUnregisteredNumbers` is **opt-in on register presence** (`HasScannableRegister()`). Measured
2026-09-02 across the nine finance/insurance sites:

| | sites |
|---|---|
| **no `evidence_base` at all** → numeric scan never arms | **5** — `lendzy.co.uk`, `loanzy.uk`, `loancash.co.uk`, `loancalculator.co.uk`, `farmerinsurance.uk` |
| register present but **zero facts** | 2 — `adversecreditmortgage.co.uk`, `remortgagecalculator.uk` (6 banned claims each, no facts) |
| register with facts | 2 — `loanandmortgagecalculator.co.uk`, `mortgagecalculator.co.uk` (13 each) |
| carrying a `regulated` attestation | **0** — correct; none should claim FCA authorisation |

**So the sector that most needs "every number is backed by a registered fact" is the sector where
that check is largely switched off.** That is a larger and more measurable exposure than a poisoned
fact, and it is the thing this RFC is really about.

### 1c. ⚠ But switching it on naively makes it worse — and that is measured, not feared

Before proposing that those registers be populated, the scan was armed locally against those five
sites (474 components, nothing written). **It produced 5 findings and all five were false:**

- **2 regulatory** — `loancash.co.uk`'s flagged number was the **`5` of `CONC 5A`**, the rule's own
  name; `lendzy.co.uk`'s was **`0.8% per day under CONC 5A`**, a regulatory figure quoted beside its
  rule — *which that site's recorded brief REQUIRES of every regulatory figure*. The scan was
  convicting the site for doing the right thing.
- **3 third-party** — industry survey figures in a news listing on `farmerinsurance.uk`
  (*"84% expecting growth, compared to 65% of personal lines P/C companies"*).

**The regulatory half is already fixed** (`fad209b92`, council `1dd3d298`): 5 → 3, fleet-wide
regression nil. The third-party half is deliberately **not** fixed — it is `RFC_053`'s component-grain
question (a news listing wrapped in a page's own first-person surface), and a second mechanism for it
would be the duplication this estate keeps filing bugs about.

**The order matters and it is the practical heart of this RFC:** arm the check before the sector's
false-positive shapes are handled and it produces noise, and *"a scanner that always reports
something is one people stop reading"* — `claims.go`'s own words.

---

## 2. What is being proposed

A **compliance tier**: a declared, per-site risk class that raises what the claims layer guarantees.

### 2a. It must be DECLARED, not derived — there is nothing to derive it from

Checked 2026-09-02: there is **no sector or vertical field**. `classification.site_type` describes the
site's *shape* — `interactive`, `content`, `interactive-platform` — and every finance site above
carries one of those, indistinguishable from a games site. Deriving a compliance class from the
domain name or the strategy prose would be guessing about the one thing that must not be guessed.

Per the owner's 2026-08-02 §2 ruling, it therefore ships as **an opt-in field whose unsafe side is
the default**: absent ⇒ today's behaviour exactly. Proposed home: `evidence_base.compliance_tier`,
beside `regulated` and `operating_history`, because it is a property of the site's evidence posture
and that is where the other two attestations already live.

### 2b. What the tier would buy, in increasing strength

Presented as three separable steps so the owner can take one, two or all three.

**(i) Severity.** The practice family's default severity for a tiered site rises from `warning` to
`error`. Cheap, reversible, and it makes "we checked X" claims visible rather than merely recorded.
*(Note: a fleet-wide flip is RFC_003 §8 Q1; this is a per-class flip, which is why it is here.)*

**(ii) A register becomes REQUIRED.** The build gate refuses a page on a tiered site with no
scannable register — turning today's opt-in into a precondition for that class. **This is the step
that closes the 5-of-9 gap**, and it is the one with real teeth: it means a regulated-sector site
cannot ship copy until someone has written down what it is allowed to assert. ⚠ It must not land
before §1c's false-positive shapes are handled, or it refuses honest pages.

**(iii) A fact-quality floor.** On a tiered site an `attested_by`-only fact **cannot license a numeric
claim** in copy: it needs a `citation`, `sql` or `artifact` — or an attestation carrying recorded
`evidence`, exactly as `RegulatedAttestation` already demands (firm, FRN, attester, date, evidence).
**This is the direct answer to the poisoned-register question**: it does not try to detect a
fabricated fact, which is undecidable; it raises the *evidence class* a fact must have before the
platform will let copy lean on it.

### 2c. What is NOT proposed

- **Not** scanning `evidence_base` for banned claims. It stores the banned phrases themselves as
  data; scanning it convicts every site's own immune system (measured: 21 hits over 522 spec rows,
  effectively all false — `bugs_closed/414` §7c).
- **Not** an automated fixer. Findings stay HITL, for the reason 414 exists: the audit fleet's own
  attempt to "substantiate" a claim is what turned a leak into an identity.
- **Not** a fleet-wide severity change. The whole point of a tier is that it is a class, not the fleet.

---

## 3. The questions for the owner

1. **Is a declared compliance tier the right instrument at all**, versus simply requiring registers
   on finance sites as a matter of practice and leaving the code alone? The practice route needs no
   RFC and no code; it also has nothing that notices when it is not followed.
2. **Which of (i)/(ii)/(iii) do you want**, and in what order? My recommendation is **(ii) then
   (iii), with (i) last** — (ii) closes the measured gap, (iii) answers the question you asked, and
   (i) is the noisiest for the least structural gain.
3. **What is the tier's vocabulary?** A boolean (`regulated: true`) is cheapest; a named tier
   (`finance` / `legal` / `insurance` / `health`) allows sector-specific banned-claim sets later but
   invites arguments about edge cases. I lean boolean now, named later if a sector-specific rule ever
   actually differs.
4. **Who declares it, and on what evidence?** `RegulatedAttestation`'s precedent is that a claim
   someone must stand behind is a *record* (who, when, what they saw), not a flag. A compliance tier
   is a smaller claim — "this site operates in a regulated sector" — but the same logic argues for at
   least an attester and a date.

---

## 3a. OWNER DECISIONS, 2026-09-02 — and where the axis moved

**Q1 — DECIDED: requiring registers is the right choice.** The tier proceeds, and the register
requirement (§2b(ii)) is its point rather than a side effect. §4's content work — populating the five
missing registers — leads, because it is what makes the requirement satisfiable.

**Q2 — DECIDED as proposed: (ii), then (iii), then (i).**

**Q4 — DECIDED as proposed:** the tier is a RECORD (who declared it, when, on what basis), not a flag,
mirroring `RegulatedAttestation`. A boolean nobody signed is unfalsifiable six months later.

**Q3 — the owner chose a NAMED tier and questioned the boundaries: *"maybe it should have a semantic
decision layer rather than sector specific"*. I agree, and the estate has already reached for the same
thing twice without anyone joining them up.**

### Why sector is the wrong key for THESE three enforcements

Read what (i), (ii) and (iii) actually respond to. None of them is about an industry:

| enforcement | what actually drives it |
|---|---|
| register REQUIRED | does this site assert **external facts** — rules, figures, data — as true? |
| fact-quality floor | what is the **cost of a wrong number** here? |
| practice severity | does anyone **rely** on "we checked X"? |

Sector is a *proxy* for those, and a leaky one in both directions. A mortgage calculator that quotes
no figures needs less than a foraging site that says which mushrooms are safe; `vetcomparison.uk`
carries animal-health consequence and is not "finance"; a "know your rights" content site is
legal-adjacent without being a law firm. And sector boundaries have no ground truth, so every edge
case becomes an argument — whereas *"does this site quote external rules as fact?"* has a yes or no.

**The estate already agrees with the owner, in two places, and neither is joined up:**

1. **The claims layer's existing controls are claim-shaped, not sector-shaped.**
   `RegulatedAttestation` is not "this is a finance site" — it is "this firm is authorised, here is the
   FRN". `OperatingHistoryAttestation` is not "this is a reviews site" — it is "we really do test
   things, here is who attested". A sector tag would be the odd one out among its own siblings.
2. **`site_archetype.constraints` is ALREADY a semantic compliance layer** — and nothing enforces it.
   `loancalculator.co.uk` carries, in prose: *"Never make calculators appear to give regulated
   financial advice"*, *"Never add real lender recommendations or ranked product tables without FCA
   authorisation disclaimers"*, *"Never reposition the site as a lender or broker without appropriate
   regulatory framing"*. `[MEASURED 2026-09-02: 8 sites have an archetype; 5 mention regulation; all 8
   carry a `constraints` array.]`

⚠ **But `site_archetype` must NOT become the key.** It is **agent-written prose**, and a guarantee
conditional on an LLM classifier inherits that classifier's gaps. It is evidence that the semantic
axis is the natural one — not a foundation to enforce on. The tier is declared by a person.

### Proposed vocabulary: a three-rung POSTURE ladder, each rung adding exactly one enforcement

Named, as the owner asked, but named for the **claim relationship** rather than the industry. Ordered,
because a ladder is what makes "named tier" workable where overlapping booleans would not be — lendzy
is simultaneously "quotes external rules" and "readers act on it", and a single name has to hold both.

| rung | means | requires |
|---|---|---|
| **`standard`** (absent — the default) | the site's claims are about its own offering | today's behaviour, unchanged |
| **`sourced`** | the site asserts **external** facts, rules or figures as true | a scannable register (§2b(ii)) |
| **`relied_upon`** | a reader may act on those assertions to their **financial, legal, medical or safety detriment** | everything in `sourced`, plus the fact-quality floor (iii) and raised practice severity (i) |

Worked: `lendzy.co.uk`, `loancalculator.co.uk`, `vetcomparison.uk` → `relied_upon`. A reviews site
saying "we tested six mowers" → `sourced` at least. `webdesign.uk` → `standard` or `sourced`
depending on whether its figures are about itself or the world.

### Where sector still has a job — and it is a different one

Not "sector is wrong", but "sector is not the key for these three". It has exactly one real use:
**a shared banned-claim set** — "guaranteed acceptance", "no credit check", "guaranteed compensation"
are sector-specific falsehoods, and a `finance` label would let one curated set attach to every site
in that sector instead of each hand-rolling its own `banned_claims`. That is additive, needs none of
this design, and can be a separate optional field later. It should **not** be smuggled in as the key
now, because doing so would tie the register requirement to an argument about industry boundaries.

### The one thing I would push back on, stated so it is on the record

A declared posture is a **judgement**, and judgements drift — a site that starts `standard` and grows
a rates table is now `relied_upon`, and nothing notices. The Q4 record (who, when, on what basis) is
the mitigation, but it is a weak one. If this ever needs strengthening, the honest instrument is a
**detector that flags a mismatch** — a `standard` site whose copy quotes rulebook citations or
regulatory figures is a candidate for re-posturing — which is cheap to build on the machinery that
already exists and would be the natural Phase 2. **Not proposed now**; recorded so it is a decision
rather than an oversight.

## 4. What this lane has already done, so the RFC is not the whole answer

- **`fad209b92`** (council `1dd3d298`): regulatory rule citations no longer read as business numbers.
  This is a precondition for (ii) and is done.
- **`bugs_closed/414`**: the claim rules now also read the *instruction* a generator is given, not
  only its output (`CLM-030`).
- **Not done, and it is content work rather than platform work:** populating registers for the five
  register-less finance sites. That belongs to the site lanes, needs no RFC, and would deliver most
  of the benefit of (ii) without any code at all. **If only one thing happens as a result of this
  RFC, it should be that.**

---

## 5. Sources

`platform/orchestration/datahelpers/claims.go` (`HasScannableRegister`, `ScanUnregisteredNumbers`,
`isExcludedNumber`, the new regulatory-citation exclusions) · `claims_regulated.go`
(`RegulatedAttestation` — the record-not-a-flag precedent) · `claims_practice.go`
(`OperatingHistoryAttestation`, warning-severity doctrine, RFC_003 §8 Q1) ·
`refresh_evidence_base_action.go` + `evidence_citations.go` (the daily re-check that already covers
citation and sql facts) · `bugs_closed/414` §7c and §7o · register `CLM-015`, `CLM-026`, `CLM-030` ·
owner rulings 2026-07-29 §1 (guarantee change ⇒ RFC) and 2026-08-02 §2 (opt-in, unsafe default OFF).

## 6. How to re-derive every figure here

```sql
-- fact sources across the fleet (319 / 17 sites, 2026-09-02)
WITH f AS (SELECT s.domain, fact FROM site_specs ss JOIN sites s ON s.id=ss.site_id,
     LATERAL jsonb_array_elements(CASE WHEN jsonb_typeof(ss.data->'facts')='array' THEN ss.data->'facts' ELSE '[]'::jsonb END) fact
     WHERE ss.is_current AND ss.aspect='evidence_base')
SELECT jsonb_typeof(fact->'source') AS t,
       CASE WHEN jsonb_typeof(fact->'source')='object'
            THEN (SELECT string_agg(k,'+' ORDER BY k) FROM jsonb_object_keys(fact->'source') k) END AS keys,
       count(*), count(DISTINCT domain) FROM f GROUP BY 1,2 ORDER BY 3 DESC;

-- register coverage on the compliance-sensitive sites (9 sites, 2026-09-02)
SELECT s.domain, (eb.id IS NOT NULL) AS has_register,
       COALESCE(jsonb_array_length(CASE WHEN jsonb_typeof(eb.data->'facts')='array' THEN eb.data->'facts' END),0) AS facts
FROM sites s LEFT JOIN site_specs eb ON eb.site_id=s.id AND eb.aspect='evidence_base' AND eb.is_current
WHERE s.status IN ('active','deployed')
  AND s.domain ~* 'loan|mortgage|credit|lend|insur|finance|legal|law' ORDER BY 2, 1;
```

For §1c, export those sites' components (the RUNBOOK's pod-side recipe, counted three ways) and run
`cmd/claimscan` with a facts-free register — **that arms the numeric scan without shipping anything**,
and it is how the 5-findings-all-false result was obtained.
