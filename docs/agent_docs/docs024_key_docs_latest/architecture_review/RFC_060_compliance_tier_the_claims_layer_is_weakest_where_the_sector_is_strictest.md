# RFC_060 — a COMPLIANCE TIER: the claims layer is weakest exactly where the sector is strictest

**Status: FULLY OWNER-DECIDED — NOTHING ON THIS RFC IS OPEN, AND Q5/Q6/Q7 ARE ALL BUILT.**
Q1–Q4 ruled 2026-09-02; Q5/Q6/Q7 ruled 2026-09-03 (§3f) and built the same day: **Q6 APPROVED**
(`ac670badf`, council `57a9939f`, 2 advisory objections both answered — one by a follow-up fix,
`3c1e1b61c`), **Q7 facts half APPROVED** (`6ec879212`, council `17fb9105`, all nine reviewers
approve), **Q5 council-submitted, verdict pending** (`939593e4c`, council `9b11752c`). Q7's
`banned_claims` half was already live from the day before (`e5b1a0f01`, confirmed running, first
pass filed nothing but that zero is uninformative by construction — see §3e's caveat). **None of
Q5/Q6/Q7 is deployed yet** — committed, awaiting a roll (Q5's council verdict still pending doesn't
block that — `Council-Submitted` was used, not `Council-Reviewed`). What remains after a roll is the
tier MECHANISM itself (the posture-ladder field + the register-required gate) — none of §2's design
is code yet; everything built today is upstream of it, per §3c's own track order. Historical
statement of the questions follows.
~~Open: **Q5** (§3b)~~ — citation-code recognition is finance-only, doesn't
generalise to other regulated sectors. **Q6** (§3d) — a citation can be substantively true and still
name the wrong rule; CONFIRMED STRUCTURAL — the FCA Handbook has no rule-level URL, so a fact
registered through the fully-verified path can still name the wrong rule, permanently. Fix sketched
(span-match within already-fetched text, no new fetch), owner says build it. **Q7** (§3e, new) — the
register's write-time verification guarantee is enforced by only ONE of its two write paths; today's
four hand-written registers (§1d) all bypassed it. **§1d (2026-09-02): day one of the register
programme found five wrong live claims in one day**, across three independent lanes, just from
reading the cited source. ~~Nothing built yet.~~ **CORRECTED 2026-09-02 evening — the TIER is
unbuilt, but Q7's `banned_claims` half IS built, council-APPROVED and committed (`e5b1a0f01`,
§3e); it is not yet RUNNING (needs a chassis roll, then the next daily `evidence-freshness`
pass).** A header reading "nothing built" over a shipped, approved detector is the stale-status
trap this estate keeps filing — the state now lives in §3e and is stated there in full.
**2026-09-02 21:30 — CONFIRMED RUNNING.** Owner refreshed the kubeconfig; verified properly this
time (a plain binary grep for the commit SHA is a KNOWN-BROKEN method — `buildinfo.GitCommit` is
one string, not an ancestry, so it returns ABSENT even for a commit that certainly built the
binary; confirmed even an 8-day-old ancestor SHA grepped absent, which is what exposed the method
as broken rather than the deploy). Used `service_binary_capabilities` instead (built for exactly
this): every current `agent-chassis` pod reports `git_commit=0d2feee2` (recorded 21:24–21:26Z,
after the pod restarts), and `git merge-base --is-ancestor e5b1a0f01 0d2feee2` returns true —
`e5b1a0f01` IS in the running build. **Gate 1 (deploy) is CLEAR. Gate 2 (the daily
`evidence-freshness` tick) is NOT** — `last_completed_at` is still 09:08:57 this morning, unchanged;
next tick ~09:09 tomorrow (2026-09-03). The check is live but has not yet run against real data.
See §3a, §3b, §3d, §3e.

Filed 2026-09-02 by the `bugfix_414_planted_marker_as_claim` lane, out of the owner's question
*"what can I do about the poisoned register hole, and shouldn't compliance be strong for sites that
require compliance strongly — finance, legal, insurance?"* The instinct is right and the measurement
moves the target, so the answer is routed here rather than built: it changes what the claims layer
GUARANTEES for a class of sites, which is the 2026-07-29 §1 trigger.

**Every figure below carries the date it was counted** (owner ruling 2026-08-22) and every one is
re-derivable from the commands in §6.

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

### 1d. Day one of the register programme (2026-09-02): four site lanes acted on §4's own recommendation, and every one that read a cited rule found it wrong

Live register coverage, re-measured `[MEASURED 2026-09-02, afternoon]` (was §1b's table, this morning):

| domain | register | facts |
|---|---|---|
| `lendzy.co.uk` | migration 695 written, **round 2 at council — not yet applied** (killed twice by today's rolling chassis deploys, resubmitted each time) | 0 (pending) |
| `loanzy.uk` | migration 697 **applied** | 3 |
| `loancalculator.co.uk` | migration 699 **applied** | 12 (9 via CCA 1974/SI citations, `legislation.gov.uk`) |
| `farmerinsurance.uk` | migration 698 **applied** | 7 |
| `loancash.co.uk` | **still none — no session assigned; owner informed** | 0 |

Fleet citation-sourced facts, same method as §1a: **256**, up from **~192** that same morning
`[MEASURED 2026-09-02]`.

> **RE-MEASURED `[2026-09-03 09:57 UTC]`, from the `bugfix_414` lane — four of the five are DONE, so
> §3c track 1 is now ONE site, and the wider population it sits in is thirteen.**
>
> - **`lendzy.co.uk` — register is LIVE**: 8 facts, 5 `banned_claims`. The table above still reads
>   "not yet applied … 0 (pending)", which was true on 09-02 afternoon.
> - **`farmerinsurance.uk` — 7 facts AND 5 `banned_claims`**; the banned half landed **09-02
>   18:34:47 by migration 713**, not by the 09-03 09:11 refresher whose name the *current* row now
>   carries. Read the spec history, not the current row — the refresher rewrites the whole row and
>   relabels `created_by` on every key it merely preserved, including one it has no code path to
>   author (LANDMINES: *"A refreshed spec's `created_by` names the last WRITER"*).
> - ~~**`loancash.co.uk` — still none**~~ → **DONE `[2026-09-03 12:45 UTC]`, owner-directed**: migration
>   **738** applied — **19 facts + 6 `banned_claims`**, 3 carrying `corrects_site_citation`.
>   **§3c TRACK 1 IS NOW COMPLETE — all five finance registers exist.** Council-Submitted
>   `cf7470b7`. Detail at §1e below.
>
> **The wider frame, same query:** `[MEASURED 2026-09-03]` **13 of 39 `deployed` sites hold no
> current `evidence_base`** — `advertise.co.uk`, `cookly.uk`, `cv1.co.uk`, `designblog.co.uk`,
> `garden-tools.uk`, `homegarden.uk`, `idea.uk`, `lampenkap.com`, `loancash.co.uk`, `oxenunity.com`,
> `seotools.co.uk`, `vetcomparison.uk`, `websitepromotion.co.uk`. Q1's requirement is finance-scoped,
> so most of those are not violations today. **Two are in ruled or imminent scope**: `loancash.co.uk`
> under Q1, and `vetcomparison.uk` — already flagged at §3b line 509 — in exactly the sector Q5's
> ruling names as next (*"extending to vet and legal quite soon"*). **The presets Q5 approved will
> land on a site with no register to apply them to.**
>
> ⚠ **And nothing raises the absence — this is structural, not an oversight in the backlog.**
> `resolveEvidenceSites` (`refresh_evidence_base_action.go` **:281**, its fleet query at **:291**,
> read at `f1e110a82`) builds the daily sweep's target list
> as `SELECT site_id FROM site_specs WHERE aspect='evidence_base' AND is_current` — it selects the
> sites that **have** a register. **The target set is defined by the presence of the very thing whose
> absence is the defect**, so a register-less site is invisible to the freshness sweep, to the fact
> checks, and to the new `invalid_banned_claim_pattern` detector alike, for ever and silently.
> Two weaker searches point the same way and are marked as searches, not proofs: no `site_work_items`
> row has ever been filed under any `item_type` matching `%evidence%`/`%register%`/`%claim%` other
> than `claims_unverified` (47), `stale_evidence` (10), `spec_supplies_claim` (2) and
> `stale_directory_claim` (2) — every one of which presupposes a register exists; and a name-search
> over `platform/`, `cmd/`, `internal/`, `pkg/` for four candidate identifiers
> (`missing_evidence_base`, `no_evidence_register`, `missing_register`, `evidence_base_missing`)
> returns nothing. **Q1 requires registers; no reader enforces or even reports the requirement.**
> Recorded here rather than built: it is this RFC's build, not the `414` lane's.

### 1e. The fifth register (`loancash.co.uk`, migration 738, 2026-09-03) — and the two traps it turned up

**Track 1 is finished.** All five finance sites named in §1d now hold a register. The last was also
the highest-risk, for a structural reason worth stating: **the other four calculate or compare;
loancash EXPLAINS THE RULES THEMSELVES** to people in financial difficulty. Its 30 served pages carry
**338 regulatory-shaped sentences** `[MEASURED 2026-09-03, crawled at the artefact with an
invented-URL 404 control]`, of which **three** cite a rule number. The rest state 0.8%, £15, 100%,
8 weeks, 6 months, 60 days and 3% per month in plain English, with nothing re-checking any of them.

**19 facts** — 11 FCA Handbook rules (CONC 5A.2.2/.3/.10/.14, CONC 6.7.23, CONC 7.6.12/.14,
CONC 5.2A.4, DISP 1.6.2, DISP 2.8.2(1)/(2)(a)/(2)(b)) and 8 statutory (Debt Respite Scheme Regs 2020
regs 16/24/26/32, Credit Unions Maximum Interest Rate Order 2013 art 2, FSMA 2000 ss.19/23).
**19/19 quotes verified through the production matcher**, absent control false in every run.

**The base rate holds for a fourth lane — three more wrong live claims** (lendzy 2 · loanzy 1 ·
loancalculator 2 · loancash 3). All three recorded as `corrects_site_citation`; served copy untouched:

- **CONC 5A.2.14 — the £15 default cap is CUMULATIVE across the agreement**, "whether in relation to
  one breach or cumulatively in relation to multiple breaches". Two pages frame it as *per missed
  payment*. The site **understates the protection it exists to explain**: a reader with two missed
  payments would accept a second £15 as lawful. It is not.
- **CONC 7.6.12 — the CPA limit is TWO REFUSED requests, and no £1 threshold exists anywhere in
  CONC 7.6.** One page says "cannot take more than one payment attempt of over £1". The site's *own*
  CPA page states the rule correctly — an internal contradiction, not a house view.
- **CONC 5.2A.4 — affordability is CONC 5.2A, not "CONC 5A"**, which is the price-cap chapter and
  contains no affordability rule. Cited correctly on three other pages.

**TRAP 1 — the "shared finance banned-claims set" is TWO sets, and the difference is invisible in a
coverage count.** `lendzy.co.uk` carries a bare `\bno credit checks?\b`; `loanzy.uk` and
`loancalculator.co.uk` carry a narrow variant requiring the product noun. `[MEASURED 2026-09-03]` the
**bare** variant fires on loancash's *correct* consumer advice that an employer salary advance
involves "no interest and no credit check" (**1 hit**); the narrow variant fires **0** across all 30
pages. Adopting "the shared set" without checking *which width* would have convicted the site of its
own accurate guidance — on the site whose entire product is accurate guidance. **A count of sites
carrying "the set" cannot see this.** All 6 adopted patterns were compiled with the production prefix
`regexp.Compile("(?i)"+p)` (`claims.go:468`) and fired against a positive control: 6/6 compile, 6/6
match their positive, 0/6 match anything served.

**TRAP 2 — a hand-transcribed quote fails silently and for ever.** The DISP 2.8.2(2)(b) quote was
first written with commas where the source has parentheses ("became aware, or ought reasonably to
have become aware,"). It returned **false** on the production matcher. Shipped, it would have
classified as `citation_lost` drift **every day**, a false alarm indistinguishable from a real one.
**Never hand-transcribe a citation quote — paste it, and let the matcher decide.**

**What it discharges.** `loancash_couk_fca_validation/README_where_we_are.md` (2026-08-11) verified
the three price-cap constants by hand and then wrote: *"What is still true is the second half of the
worry: nothing is checking … What actually earns its keep is something that reads the rulebook and
shouts if it disagrees with what is on our page."* That is this mechanism, arriving three weeks
later. That lane also named the complaint-deadline calculator as its highest-value unchecked item,
*because limitation periods, unlike the price cap, do move* — DISP 2.8.2(1)/(2)(a)/(2)(b) are
registered for exactly that reason.

**The number that matters more than the coverage:** three independent lanes, populating registers by
reading the cited source rather than trusting the site's existing prose, found **five wrong live
claims in one day** — not from a detector, from a human-equivalent read of the primary source:

- `lendzy.co.uk` — the two mis-attributed CONC rules (§3d/Q6).
- `loanzy.uk` — MaPS (Money and Pensions Service) grouped under "FCA-authorised services"; MaPS is the
  **statutory guidance body**, not an FCA-authorised firm.
- `loancalculator.co.uk` — a settlement-figure claim of "ten working days" where the actual period
  under SI 1983/1564 reg 4 is **twelve**; and an invented "10% per 12 months" early-repayment-charge
  threshold where the actual figure is **£8,000 per 12 months** (s.95A(2)(a)).

Every lane that checked found an error. That is the base rate the whole claims-verification layer —
and specifically §3d/Q6's checker — is being built against, on sites already carrying `relied_upon`
consequence, discovered on the FIRST pass through content nobody had previously verified against its
own cited source.

**A related write-time gap, surfaced by the same day's work (raised for a decision, not decided
here — see the addendum below):** the four registers above were all written **by hand, via migration**
— none went through `verify_and_register_citations` (V5/CLM-008's own write path), which is the one
place a citation is checked against its host BEFORE it enters the register. The loanzy lane's run hit
this directly: `maps.org.uk` and `moneyhelper.org.uk` both sit behind a Cloudflare "Just a moment…"
challenge page. A citation against either would pass a human skim and then read `citation_lost` in the
daily refresh FOR EVER — a failure of the host, not the fact. `verify_and_register_citations` would
have caught it (a challenge-page title never matches a real quote); the migration path that actually
shipped all four registers today has no equivalent gate. **A write-time verification guarantee that
only one of the register's two write paths enforces is not a guarantee** — see §3e.

### 1e. The sector-set question answered itself, 3 of 3, by sites choosing independently (2026-09-02, evening)

The proposal that finance sites share a **curated banned-claims set** is the part of this RFC most
likely to look obviously right and be wrong. It now has a measurement, and it came from the sites
rather than from me: **three finance sites independently adopted a shared set today and all three
declined the same pattern — a literal `%APR` ban — each with its own live-copy census.**

| site | migration | why it declined the literal-APR arm |
|---|---|---|
| `loanzy.uk` | 702 | declined on its own copy census |
| `adversecreditmortgage.uk` | (source set) | the arm was never in the set it contributed |
| `loancalculator.co.uk` | 707 | the pattern matches **twice** on live copy — `compare-loans`' deliberately illustrative *"a 7.9% APR loan and an 8.4% APR loan"* |

**On a loan-calculator site the illustrative APR IS the pedagogy.** That is §1c's
convict-the-site-for-doing-its-job class with a page and a sentence attached, and it is now the
majority outcome rather than a hazard I inferred. Three sites, three independent calls, same arm.

**Why this is decisive for Decision 3 rather than merely cautionary.** It compounds with the
structural finding in §3b/§3d from the opposite end of the layer: `scanBannedClaims` has **no
regulatory-citation exemption**. So a curated set containing a figure pattern would re-convict at
**blocker** severity exactly the content `cmd/regcheck`'s number scan now exempts at error severity
(`fad209b92`) — the two halves of the layer disagreeing about the same sentence, with the stricter
half winning and the sector's own honest copy losing. **A shared sector set cannot safely carry a
figure pattern until citation recognition reaches the banned-claims layer** — which is Q5, and this
is the second independent reason to sequence it first.

**A second refinement from the same day, same shape.** `loancalculator` recommends adopting
`loanzy` 702's **narrowed** no-credit-check form — banning the *lending promise*, not the phrase —
because calculator and tool sites truthfully describe their own tools as involving no credit check,
and the broad form permanently refuses that honest sentence. Same failure as the APR arm: a
sector-wide pattern meeting a site whose **archetype** makes the banned phrase true.

**What this says about the axis the owner already moved.** Both findings are archetype collisions,
not sector collisions — every site here IS finance, and the sector key predicts nothing about
either. The owner's instinct that the tier wants *"a semantic decision layer rather than sector
specific"* (§3a) is not merely tidier; on the only evidence this RFC has, **sector is the key that
would have got both of these wrong.**

⚠ **And a live authoring trap on the mechanism the whole idea rests on** (`LANDMINES`, corrected
2026-09-02, `4f1ca1384`): a per-site pattern that fails to compile degrades **silently** to a
literal of its own source text (`claims.go:348`) — no logger, no error path, and the admin door
counts patterns without ever compiling one. It is then armed, listed, counted and **inert**, and
every count-based verification passes. `[MEASURED 2026-09-02]` the fleet is clean today — **239
live per-site patterns across 19 sites, 0 non-compiling, 0 doubled-backslash**, controls firing —
but *a curated set distributed to N sites is exactly the mechanism that would propagate one typo
into N silently inert guards*, and the census is stale by ADDITION on the next seed. Any shared-set
proposal owes a compile-and-probe step at distribution time; `loancalculator` 707 is the worked
example (8/8 probe-fired, then a 0-match census over all 28 served pages so arming could not refuse
a current save).

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

**Q3 VOCABULARY — DECIDED 2026-09-02: the three-rung posture ladder below is approved as proposed**
(`standard` / `sourced` / `relied_upon`, exact names and requirements). This closes Q3 fully — the
axis (semantic, not sector) and the vocabulary are both now settled. Nothing left open on Q1–Q4; the
one remaining question is §3b's Q5.

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
| **`standard`** (absent — the default) | the site's claims are about its own offering | ~~today's behaviour, unchanged~~ → **AMENDED by the owner ruling of 2026-09-03 (§3g): a register IS required, at the ATTESTED bar — `value` + `context_terms`, no citation** |
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

**UPDATE 2026-09-02, from the `bugs_open/414,` lane — a naive shared set does not survive contact,
measured, not assumed.** `adversecreditmortgage.co.uk`'s six-pattern set was offered to lendzy and
loanzy as a starting point. Both lanes independently diverged on the **same two of six**, for
**different reasons each**:
- `\bno (credit )?checks?\b` — lendzy's false positive is adjacent UI text concatenating to "No
  Check" in the visible body (a "Yes"/"No" answer button beside a "Check for a breach" button);
  loanzy's is a TRUE first-person statement about its own tool ("There's no credit check involved").
- `\b[0-9]+(\.[0-9]+)?% (apr|apcr|rate)\b` — both omitted it. lendzy measured 3 hits, every one the
  credit-union rate cap quoted beside its named rule in educational content; loanzy's framing: *"on
  your sibling a literal rate is a price promotion; here it is pedagogy."*

**2 of 6 patterns needed a per-site exception at 100% of adopting sites (2 of 2) — that is
disagreement on contact, not drift accumulating over time.** A shared set, if built, needs per-site
override machinery from day one, not as a later refinement.

**A second, structural problem, verified at the code, not proposed as a build:** the regulatory-
citation exemption (`fad209b92`, §3b) lives ONLY in `isExcludedNumber`/`ScanUnregisteredNumbers` — the
NUMBER scan. `scanBannedClaims` (`claims.go:527`) consults no such exemption; it has no reference to
`regulatoryCitationContextRe` or any citation-awareness at all `[VERIFIED 2026-09-02: read the
function, confirmed absent]`. **So a shared sector set containing a literal-rate pattern would
re-convict — likely at BLOCKER severity in the refusing union — exactly the content the number scan
now correctly exempts at `error`:** the two layers would disagree about the same sentence, on the
sites whose whole purpose is quoting capped rates beside their rules. **If a sector set is ever
curated and contains a figure-shaped pattern, the citation exemption must reach `scanBannedClaims`
first, or the set is unsafe by construction.** Not proposed as a build — a precondition, recorded so
it is a decision rather than a surprise if this is ever picked up.

### The one thing I would push back on, stated so it is on the record

A declared posture is a **judgement**, and judgements drift — a site that starts `standard` and grows
a rates table is now `relied_upon`, and nothing notices. The Q4 record (who, when, on what basis) is
the mitigation, but it is a weak one. If this ever needs strengthening, the honest instrument is a
**detector that flags a mismatch** — a `standard` site whose copy quotes rulebook citations or
regulatory figures is a candidate for re-posturing — which is cheap to build on the machinery that
already exists and would be the natural Phase 2. **Not proposed now**; recorded so it is a decision
rather than an oversight.

## 3f. OWNER DECISIONS 2026-09-03 — Q5, Q6 and Q7 all RULED; nothing on this RFC is now undecided

Given directly, in one message, closing every open question. **Q6: "do as you suggest."
Q7: "fix as you suggest." Q5: "I will be extending to vet and legal quite soon so let's fix it
with those in mind."**

**Q6 — DECIDED: build the fix as proposed in §3d.** Span-match within the already-fetched text, no
new fetch. This confirms directly what had been relayed via the lendzy lane, so the "confirmation
pending" note above is now discharged.

**Q7 — DECIDED: build the fix as proposed in §3e.** The `banned_claims` half is already built and
council-APPROVED (`e5b1a0f01`); this ruling covers the **facts** half — the register's second write
path, the hand-written/migration one that all four of §1d's registers used, which bypasses
`VerifyAndRegisterCitationsAction` entirely. Build it into the same daily loop, per §3e's own
"one loop, two guarantees".

**Q5 — DECIDED, AND THE ANSWER INVERTED, because the owner supplied the fact the recommendation
turned on.** I recommended the plain per-site field and explicitly advised holding the sector
PRESETS back — *"add sector presets only once a second sector actually needs one, per this RFC's own
instinct not to design ahead of a second consumer"* (§3c track 3). That instinct was sound and its
premise was simply false: **a second and third consumer are imminent and now named — veterinary and
legal.** So the third bullet of §3b's proposed fix, offered as optional, is **IN**:

- `evidence_base.citation_codes: []string`, per-site declared data — as proposed;
- the hardcoded FCA list stays the always-on default and `citation_codes` unions on top — as
  proposed, so no regression and no forced migration;
- **plus** named sector presets a site opts into (`veterinary` → RCVS/VMD, `legal` → SRA, `medical`
  → GMC/MHRA/CQC). **Note `legal` was NOT in §3b's sketch** — that listed veterinary and medical —
  and it is now a first-class target rather than an "…" at the end of a list.

**Four constraints the builder inherits, none of them optional:**

1. **The matching rule does not change.** Exactly what `fad209b92` shipped: **case-sensitive**, and
   the code must be **immediately followed by a digit**. That is what stops a bare `FCA` — or a bare
   `RCVS`, `SRA`, `GMC` — exempting any number that happens to sit near it. Two-letter codes stay
   excluded. A preset that relaxes any of these is a different mechanism wearing this one's name.
2. **§1c's ordering warning applies PER SECTOR, not once** — *"arm the check before the sector's
   false-positive shapes are handled and it produces noise"*. Each preset is measured against that
   sector's live copy **before** it arms. `vetcomparison.uk` is deployed today and is this RFC's own
   `relied_upon` worked example, so veterinary is a real corpus to measure against, not a
   hypothetical one.
3. ⚠ **The failure mode INVERTS relative to `banned_claims`, and it is the more dangerous
   direction.** A broken banned-claim pattern makes a guard **inert** — it fails open on the ban.
   A wrong or over-broad `citation_codes` entry makes the numeric scan **BLIND** — it fails open on
   *detection*, silently exempting numbers that should have been caught, on exactly the sites this
   RFC exists to protect. Both are silent; this one is worse, because a disarmed check and a clean
   site produce identical output. **So probe-fire each preset in BOTH directions: codes that must be
   exempted, AND numbers that must still be CAUGHT.** An exemption list validated only on what it
   should exempt is untested.
4. **A preset is shared vocabulary distributed to N sites**, which is the same distribution
   mechanism `LANDMINES` warns propagates one author's typo into N sites at once. Compile and
   probe-fire at distribution time, as migration 707 did.

**Q5 BUILT 2026-09-03 (`939593e4c`), Council-Submitted: `9b11752c`.**
`EvidenceBase.CitationCodes` (ad hoc) + `CitationCodePresets` (named: veterinary/legal/medical)
union onto `regulatoryRulebookCodesBase`, compiled once per site at `ParseEvidenceBase` time —
constraint 1 enforced in code (same regex shape, case-sensitive, digit-adjacent, two-letter codes
silently dropped even when site-declared), asserted directly by a test comparing a bare site's
compiled pattern to the fleet default byte for byte. 13 tests; three medical fixtures were rewritten
mid-build after the file's own must-fire discipline caught them passing vacuously on the first
draft. Mutation-verified: disabling preset/code expansion fails 8 of 13 tests — every one
exercising the new mechanism — while the 5 mutation-independent tests correctly hold.
**Constraints 2–4 (measure each preset against its sector's live copy before arming; probe-fire
both directions; compile-and-probe at distribution time) are content-population-time obligations
for whoever arms a preset on a real site — this commit builds the mechanism, it does not discharge
them.** No site has `citation_codes`/`citation_code_presets` set today; behaviour is unchanged
fleet-wide until a human opts one in.

**Sequencing note, 2026-09-03:** none of Q5/Q6/Q7 is written yet, so none rides the chassis build
rolling now. What that build DOES carry is `e5b1a0f01` — after which the pattern detector fires on
the next daily `evidence-freshness` pass, and **its first findings are the thing to read**, not the
roll.

**Q6 BUILT 2026-09-03 (`ac670badf`), Council-Reviewed: `57a9939f` — APPROVED 2026-09-03 09:49Z, 2
advisory objections, none high.** `editquality` correctly flagged two things, both answered: (1) the
fix only covers the daily refresh path, not `verify_and_register_citations` (registration) —
deliberate, not fixed here; today's real mis-attributions arrived via migration, which registration
never sees anyway, so this was the load-bearing path. (2) the heading marker was hard-coded `[RG]`
and the FCA Handbook carries other provision-type letters (E, D) — **fixed** (`3c1e1b61c`): widened
to `[A-Z]`, since the marker letter was never load-bearing to the check (the date before it is), only
untested. Live-checked CONC 6.7 and COBS 2.1, both R/G-only — no counter-example in hand, fixed
anyway because the gap was real regardless. `reuse_agent` noted `verifyCitationLiveForRule`
duplicates `verifyCitationLive`'s fetch call — accepted tradeoff, documented in the code: touching
the shared function would affect 3 other call sites with no rule-attribution concept.
`datahelpers.CitationRuleSpan` (pure) + `actions.verifyCitationLiveForRule` (new function, doesn't
touch `verifyCitationLive`'s three other call sites) wired into `refreshCitationFact`, reading
`fact["rule"]` off the raw map. Ten tests total (six pure, four httptest/offline), the load-bearing
one carrying its own control (whole-page verification is asserted to PASS the same quote first,
proving the rule-scoped rejection is `CitationRuleSpan`'s doing, not a broken fixture).
Mutation-verified in an isolated worktree: disabling the span check fails exactly the tests it
should. Not yet deployed — same two gates as Q7's `banned_claims` half (build+roll, then next daily
tick).

**Q7 FACTS HALF BUILT 2026-09-03 (`6ec879212`), Council-Reviewed: `17fb9105` — APPROVED 2026-09-03
10:00Z, ALL NINE REVIEWERS APPROVE, zero objections beyond one advisory note (`reuse_agent`: no
recorded search for an existing bot-detection utility elsewhere in the codebase before writing
`botChallengeReason` — none found in this session's own earlier investigation either).**
`botChallengeReason` detects the loanzy lane's exact Cloudflare interstitial (title, the
`_cf_chl_opt` JS variable, the noscript fallback id — all three confirmed live against
`maps.org.uk` by curl before writing the detector, not guessed) from the RAW html inside
`fetchCitationDocument`, since every marker lives in `<head>`/`<script>`/`<noscript>`, all excluded
from visible-text extraction by design. Turns a challenge page into a `fetch_error` (the same
"unknown, never drift" bucket a 403 already gets) instead of the wrong `citation_lost` the loanzy
lane found live. One shared choke point protects both effective write paths: registration already
refuses any `!outcome.Found` candidate, and the daily refresh now classifies migration-written
facts correctly too — migrations are raw SQL and cannot be gated by Go at insert time, so this is
the same "one loop, two guarantees" shape as the `banned_claims` half. Mutation-verified: disabling
the detector reproduces the exact originally-reported bug. Not yet deployed.

**Q7's `banned_claims` half is separately already LIVE** (`e5b1a0f01`, confirmed running
2026-09-02 21:30 — see the top status line). The `facts` half above is new, unrelated to that
deployment, and needs its own build+roll+tick before it does anything.

> **CONFIRMED 2026-09-03, and it is a caveat rather than a green light.** The build deployed
> (replicaset `75b987cbd7`) and the detector **is in the running binary** — probed at `/proc/1/exe`
> for `invalid_banned_claim_pattern` with **both** controls (target **6**, must-be-absent **0**
> exit 1, must-be-present `stale_evidence` **6**). `evidence-freshness` then ran at **09:10:23**
> under those pods and completed. Items filed: **0**.
> ⚠ **That zero is uninformative BY CONSTRUCTION.** Both result fields are `omitempty`
> (`refresh_evidence_base_action.go:216`, `:221`), so a clean result serialises to **nothing** — of
> **23** evidence runs since 09:00, **0** mention the field, a figure identical whether the code
> ran clean or never executed, and there is no log line on the clean path to separate them.
> **"Live" is proven; "works" is not.** The outstanding step is a **demand control** — plant a
> deliberately broken pattern on a scratch site, confirm the pass files an item, remove it — owned
> by the claims-verification lane. Until that is done this must not be read as a passing check.

---

## 3b. ADDENDUM 2026-09-02 (Q5, NEW — undecided): citation recognition is finance-only and does not
generalise across sectors

Raised by the claims-verification thread, from the owner's direct question: *"vet compliance and
medical and everything else will all have to be accommodated too"* — asked about `fad209b92`, this
RFC's own stated precondition for (ii).

**The mechanism is hardcoded to one sector.** `fad209b92` (this morning) added `regulatoryRulebookCodes`
to `claims.go`:

```go
const regulatoryRulebookCodes = `(?:CONC|MCOB|ICOBS|BCOBS|COBS|DISP|SYSC|PRIN|CASS|PERG|CREDS|COLL|
DEPP|GENPRU|MIPRU|BIPRU|IPRU|FEES|SUP|MAR|DTR)`
```

A single Go constant, compiled fleet-wide, FCA Handbook sourcebooks only — what stops "CONC 5A" reading
as an unbacked business number. It has no equivalent for RCVS/VMD (veterinary), GMC/MHRA/CQC (medical),
SRA (legal), or anything else, and cannot get one without a Go change + council review + image roll
**per sector**.

**Not hypothetical for this estate.** Measured 2026-09-02:

| domain | status | evidence_base |
|---|---|---|
| `vetcomparison.uk` | **deployed** | none — 0 facts, absent from the register-coverage table in §1b |
| `pool-vet-animal.internal` | pool (unbuilt) | — |
| `pool-health-medical.internal` | pool (unbuilt) | — |

`vetcomparison.uk` is already this RFC's own `relied_upon` worked example (§3a: "carries animal-health
consequence"). The moment it gets a register and the numeric scan arms under (ii), a genuine RCVS or
VMD citation on it reproduces §1c's finding exactly — the site convicted for correctly citing its own
rulebook, on a sector this RFC exists to protect.

**The estate already has the right shape for this — one field is the odd one out.** `BannedClaims` and
`AllowedEntities` are already per-site DECLARED DATA on `evidence_base`; a site states its own
vocabulary, no Go change to onboard a new one. `regulatoryRulebookCodes` breaks that pattern — it is
the one piece of sector vocabulary compiled into the binary. §"Where sector still has a job" above
already proposes exactly this shape for banned claims ("a shared banned-claim set... additive, needs
none of this design, can be a separate optional field later"); citation-code recognition is the same
problem one layer earlier — it decides whether the scan fires at all, not what it forbids.

**Proposed fix (Q5) — data-shaped, additive only, matching logic unchanged:**
- `evidence_base.citation_codes: []string` — a site's own rulebook prefixes, matched by the SAME rule
  `fad209b92` shipped (case-sensitive, code immediately followed by a digit).
- The current hardcoded FCA list stays as the always-on default — no regression, no forced migration
  for the two finance sites already carrying facts; `citation_codes` unions on top of it.
- Optionally a small sector-keyed PRESET a site opts into by name (`veterinary` → RCVS/VMD, `medical` →
  GMC/MHRA/CQC, …) instead of typing codes by hand — the same opt-in shared-vocabulary idea already
  agreed above for banned claims, extended to cover this too.

**Why Q5 and not silently folded into (ii):** §1c's ordering warning — *"arm the check before the
sector's false-positive shapes are handled and it produces noise"* — applies per sector, not once.
**Does not block Q1/Q2/Q3/Q4** — the tier design and build order stand — it constrains WHEN (ii) is
safe to apply outside finance.

**Classification:** per the 2026-07-29 ruling (§1: a shared vocabulary needs an RFC only when it changes
what the mechanism GUARANTEES), moving this from code to site-declared data doesn't change the
guarantee — citations are still excluded from the numeric scan — only widens which strings qualify,
sourced from data instead of a constant, the same status `BannedClaims`/`AllowedEntities` already hold.
Read as a normal build, not an RFC-blocking change on its own; recorded here because it gates (ii)'s
safe scope and because the owner asked the question directly.

## 3c. Where this leaves the build, now Q1–Q4 are settled (2026-09-02)

The design is done; only Q5's shape (site-declared field vs. field + sector presets) is open, and it
does not block starting. Three independent tracks, in the owner's chosen order:

1. **Content work, unblocked, no code** — populate the five register-less finance sites named in §1b
   (§4's own top recommendation). Does the most for the least risk and needs nobody's further decision.
2. **(ii) register-required gate, scoped to FINANCE-tiered sites** — safe to build now: `fad209b92`
   already covers the FCA codes those sites will cite. Extending the gate to `vetcomparison.uk` or any
   future non-finance `sourced`/`relied_upon` site should wait for Q5, or it reproduces §1c's finding.
3. **Q5 itself** — resolve the citation-code shape (plain per-site field is the smaller change; add
   sector presets only once a second sector actually needs one, per this RFC's own instinct not to
   design ahead of a second consumer).

Then (iii) fact-quality floor and (i) severity follow, per Q2's decided order.

---

## 3d. ADDENDUM 2026-09-02 (Q6, NEW — undecided): a citation can be true, sourced, and still name the
WRONG rule — nothing checks attribution, only presence

Reported by the `lendzy_co_uk` lane, from Phase B-i of populating lendzy's register (the content work
§3c/§4 recommends). Building `cmd/fcaquotecheck` (`c904ffd5d`, calls the SAME
`datahelpers.VisibleTextFromHTML` / `QuoteFoundInText` the production refresher uses, not a
reimplementation) to verify lendzy's seven existing rule-citations before registering them as facts,
they found **five correct and two wrong** `[MEASURED 2026-09-02]`:

- served copy cites **CONC 6.7.17** for the two-rollover limit — that rule is the DEFINITIONS clause
  for the range; the substantive limit is **CONC 6.7.23**.
- served copy cites **CONC 6.7.23** (the rule above) for the two-attempt continuous-payment-authority
  limit; the substantive rule is **CONC 7.6.12**.

**In both cases the underlying business claim is TRUE and the source domain is right — only the cited
rule NUMBER is wrong.** This is a distinct failure mode from anything this layer currently checks:

| existing check | asks |
|---|---|
| `ScanUnregisteredNumbers` | is this number backed by ANY registered fact? |
| CLM-025 cold audit (LLM) | is this assertion supported by the register, wording-not-topic? |
| CLM-008 citation refresh | does the REGISTERED quote still appear at the REGISTERED url, today? |

None of them ask **"is the cited rule the rule that actually says this?"** — CLM-008 re-verifies
whatever URL a fact already carries; it has no way to notice the URL was the wrong one from the start,
because it was never told what "wrong" would look like for a citation rather than a plain business
number. On a `relied_upon` site — the rung this RFC's own ladder reserves for readers who may act on
what's claimed to their financial/legal/medical/safety detriment — an accurately-sourced-but-
mis-attributed rule is arguably worse than an unsupported one: it reads as MORE trustworthy, not less.

**RESOLVED 2026-09-02 — the gap is STRUCTURAL, not confined to legacy prose.** Registration DOES
already verify the specific cited URL: `VerifyAndRegisterCitationsAction` (`evidence_citations.go:182`)
fetches the candidate's own URL and rejects anything whose quote is absent — a fact cannot enter the
register with a quote missing from the page it names. That is not where this fails. **The FCA Handbook
has no rule-level URL** `[MEASURED 2026-09-02, lendzy lane]`: CONC 6.7 is ONE page holding 54 rules
(6.7.17 and 6.7.23 both live on `handbook/CONC/6/7.html`), and every rule-level URL variant tried
(`.../6.7.23.html`, `conc6.7.23`) 200s to the wrong page. So a quote genuinely belonging to 6.7.23
verifies perfectly against a citation LABELLED 6.7.17 — same page, same bytes, same fetch.
`verifyCitationLive` answers *"is this quote on this page"*, which is right for a news source and
insufficient for a rulebook, where the page is a chapter and the citation is a line. **A fact
registered tomorrow through the fully-verified path, naming the wrong rule, passes registration and
passes every subsequent daily re-check, for ever.**

**Proposed fix shape** (design only, not built, not decided who builds it or when — changes what a
`sourced`/`relied_upon` citation guarantees fleet-wide, so it needs the same treatment as Q5): no new
fetch is required. The visible text already retrieved partitions on the rule-heading pattern
`CONC \d+[A-Z]?\.\d+\.\d+` followed by a `dd/mm/yyyy` date and an `R`/`G` marker — the same split that
gave 78 identified rules on CONC 5A and 54 on CONC 6.7. Checking "does the quote fall within the SPAN
belonging to the rule this fact names" is a narrowing of the existing page-level check, using data
already in hand — not a new dependency. **The trap for whoever builds it:** anchor the span on the
id **plus its date and R/G marker**, never the bare id — rule text routinely cross-references other
rule ids inline (CONC 6.7.17 itself names the range "CONC 6.7.18 R to CONC 6.7.23 R"), so "find the id,
take what follows" lands inside a neighbour's cross-reference and silently spans the wrong text. The
date+marker suffix is what distinguishes a rule's own heading from a mention of it — present on all
78 and all 54 rules measured.

Full measurement table and write-up:
`docs/agent_docs/docs024_key_docs_latest/lendzy_co_uk/PLAN_2026-09-02_lendzy_co_uk.md` §B5
(`9da3363a6`).

**Not fixed unilaterally, correctly** — the two mis-citations are in served copy; rewriting published
prose on an automated finding is authority the owner withheld 2026-08-21 (`bugs_open/320` §15). Routed
to the owner with a recommendation by the lendzy lane; this addendum is the platform-layer half.

**Does not block §3c's three tracks.** Lane docs:
`docs/agent_docs/docs024_key_docs_latest/lendzy_co_uk/` (PLAN §B).

---

## 3e. ADDENDUM 2026-09-02 (Q7, NEW — undecided): the write-time verification guarantee is enforced by
only ONE of the register's two write paths

Raised by the `loanzy.uk` lane, seconded by lendzy, from the same day's register-writing work (§1d).

> **The register's OTHER half has the identical shape, and it is cheaper to close** (added
> 2026-09-02 by the bugfix_414 lane, argued by `loancalculator_couk`). Q7 is about **facts**:
> a verification guarantee that only one of two write paths enforces. The same sentence is true
> of **`banned_claims`**: a pattern that fails to compile degrades **silently** to a literal of
> its own source text (`claims.go:348` — no logger, no error path), and the only thing that
> catches it is `TestEveryGlobalPatternIsAValidRegex`, a **CI test over the Go-authored
> fleet-wide set**, which by construction cannot see a pattern arriving as **data**. The admin
> door counts patterns without compiling one; a migration cannot compile a regex Postgres never
> parses. So the guard is armed, listed, counted and **inert**, and every count-based
> verification passes — the same *guarantee enforced on one path only* that Q7 names, one field
> along. `[MEASURED 2026-09-02]` the fleet is clean — **239 live per-site patterns, 19 sites, 0
> non-compiling, 0 doubled-backslash**, broken controls firing — so this is a door standing open,
> not damage.
>
> **It needs no decision from the owner and no new surface**, which is why it is a note here and
> not a Q8: the daily `evidence-freshness` task already selects every site with a current
> `evidence_base` and loads the whole register (`refresh_evidence_base_action.go:278`, `:325`;
> enabled, 86400s, last completed 2026-09-02 09:08:58 — verified, not assumed). A compile pass
> inside that existing loop is the census→cron step RFC_006 and WFA-013 have both already shipped.
> **If Q7 is built as a sweep, it should carry this in the same pass** — one loop, two guarantees,
> and the claims-verification lane owns both. **BUILT 2026-09-02 (`e5b1a0f01`):** the banned_claims
> half only — `checkBannedClaimPatterns` (pure, re-runs claims.go:348's exact compile) wired into
> `refreshOneSiteEvidence`, filing via `createInvalidBannedClaimPatternItems`
> (`insertWorkItem`/`dropOnConflict`, keyed per pattern via fnv64a, never `refreshOnConflict` —
> `bugs_closed/213`). Six tests, mutation-verified in an isolated `git worktree` (another session's
> unrelated uncommitted WIP was breaking `go test` for the whole package at the time — the worktree
> sidestepped it without touching their files): flipping the conflict policy in a scratch copy made
> the discriminating test fail on an unexpected extra query, confirming it actually distinguishes
> DO-NOTHING from DO-UPDATE rather than passing vacuously. **Council-Reviewed: `bc3697a5` — APPROVED
> 2026-09-02 18:39:52Z**, 3 advisory objections, none high-severity, `abstained:6`. Read the verdict
> directly, not the correlation alone: `prior_art_librarian`'s absence claim ("nothing else validates
> per-site pattern compile-ability") was checked afterward — every consumer of `BannedClaims`,
> `discovery_checks/check_unverified_claims.go` included, goes through `ParseEvidenceBase`'s SAME
> silent fallback; no other check exists, the claim holds. `tooling_provenance` correctly flagged no
> NOTES entry was planned — `claims_verification/NOTES_claims_verification.md`, 2026-09-02 entry, now
> is one. `editquality`'s item_key objection read the SUBMISSION SKETCH, which under-described the
> real key format (it DOES carry an `invalid_banned_claim_pattern:` type prefix, per the sibling
> convention) — a submission-quality gap, not a code defect; the real function was correct throughout.
> ⚠ **`[MEASURED ~19:40]` was COMMITTED, NOT RUNNING** — both `agent-chassis` pods started before
> the 18:30Z commit (caught by the `bugfix_414` lane verifying at the pod, not by this thread's own
> report). **CONFIRMED LIVE 2026-09-02 21:30**, after the owner's roll: `service_binary_capabilities`
> shows every current pod on `git_commit=0d2feee2`, and `e5b1a0f01` is a confirmed ancestor
> (`git merge-base --is-ancestor`). Still waiting on the daily `evidence-freshness` tick
> (`last_completed_at` unchanged at 09:08:57 — next run ~09:09 tomorrow) before the check runs
> against real data for the first time. Q7's own half (facts /
> host admission) is NOT built — still open, still needs the owner's steer on shape. Full costing and
> the two filing traps (`ON CONFLICT DO NOTHING`; key on the finding, not the site): `LANDMINES.md`,
> the banned-claims-escaping entry, amended 2026-09-02.

**The gap.** `verify_and_register_citations` (V5/CLM-008) checks a candidate citation's host BEFORE
admitting it: fetch the URL, require the quote in the visible text, reject on failure. That is real
protection — but **all four registers written today (§1d) went in by hand, via migration**, which
never calls that function. The migration path has no equivalent gate.

**Caught live, not hypothetically.** The loanzy lane's own run tried to cite `maps.org.uk` and
`moneyhelper.org.uk`; both sit behind a Cloudflare "Just a moment…" interstitial. A citation against
either passes a human skim (the page LOOKS fine in a browser, which runs the challenge) and then
reads `citation_lost` in CLM-008's daily re-check **for ever** — a failure of the host, not of the
fact, and one that would have been silently invisible if the lane hadn't happened to run the
production extractor by hand before writing (the method both lendzy's and loanzy's lanes now follow;
see `lendzy_co_uk/RUNBOOK_lendzy_co_uk.md` §8b for the full trap).

**Proposed fix (Q7), not built, not decided:** a write-time HOST admission check — before any citation
enters the register by ANY route, run its URL through the same extractor CLM-008 uses in production
(not a reimplementation, per the `cmd/fcaquotecheck` lesson in §3d) and reject a challenge-page title
as a source. Because it must guard BOTH write paths (the LLM-mediated one and hand-authored
migrations), it belongs somewhere both share — not inside `verify_and_register_citations` alone, which
is exactly what already protects one path and not the other.

**Why this is Q7 and not silently fixed:** it is the same class of question as Q6 — what a
`sourced`/`relied_upon` citation is allowed to enter the register AT ALL — and, like Q6, changes a
shared guarantee rather than patching one call site. **Does not block §3c's three tracks or Q6's
build**; it constrains what "the register is verified" means going forward as more sites populate
theirs by hand, which §1d shows is happening THIS week, not eventually.

Evidence: `lendzy_co_uk/RUNBOOK_lendzy_co_uk.md` §8b (host traps), §8c (near-identical statutory
instrument numbers as a related mis-citation source — SI 1983/1569 vs 1564, caught by
loancalculator's own schedule read; and gov.uk keeps organisation pages at FOUNDING-name slugs, e.g.
MaPS at `single-financial-guidance-body`, so name-guessed URLs 404 rather than warn).

---

## 3g. OWNER RULINGS 2026-09-03 (afternoon) — FOUR decisions, one of which AMENDS the Q3 ladder

Given in response to the register-absence finding at §1d/§1e. Recorded verbatim in substance because
one of them changes a previously-ruled table row, and a silent change there would be re-derived
wrongly by whoever builds the twelve registers.

**D1 — `vetcomparison.uk` gets a register.** *"we'd want a register for vetcomparison."* It is this
RFC's own `relied_upon` worked example (§3a, §3b) and the sector Q5's presets name, so the full
`sourced`/`relied_upon` bar applies: primary-source citations, quotes verified through the production
matcher. **Not the attested bar below.**

**D2 — loancash.co.uk's three wrong sentences are to be REPAIRED.** *"fix the loancash wrong
sentences."* This lifts the standing hold from the 695 precedent / `bugs_open/320` §15, under which
`corrects_site_citation` recorded a finding and the served copy was deliberately left alone pending
the owner. **The hold is lifted for these three findings on this site only** — it is not a general
licence to rewrite published prose on an automated finding, and the repair goes through the framework
(the owner ruling of 2026-08-04: every site goes through the framework; and 2026-08-06: the framework
writes the content, not the session).

**D3 — build the absence detector AND populate the missing registers.** *"build the missing check and
fill the missing data for the sites."* Both halves, not one. §1d's structural finding stands as the
reason the detector is not optional: the daily sweep's target list is drawn from the sites that HAVE
a register, so absence is permanently invisible and no cadence change reaches it.

**D4 — A REGISTER FOR EVERY SITE, WITH A LOWER BAR FOR NORMAL ONES.** *"I think we should do a
register for each site to avoid AI slop but the bar can be lower for compliance for normal sites
somehow."*

This is the one that amends Q3. The ladder above gave `standard` "today's behaviour, unchanged" —
i.e. **no register**. D4 says every site gets one. The rung stays; its *requirement* changes.

### 3g(i). What the lower bar IS, derived from the code rather than invented

The owner's "somehow" has a mechanical answer, and it is better than a reduced version of the
`sourced` bar. **The two rungs' registers do different jobs, and only one of them needs a source.**

`ScanUnregisteredNumbers` (`datahelpers/claims.go:1341`) is the anti-slop mechanism D4 is reaching
for — it flags a number in the copy that no registered fact supports. Two facts about it decide the
design:

1. **It is disarmed entirely by an absent register**: `if eb == nil … return nil` (`:1342`). That is
   the hole D4 names. A site with no register can have any figure invented into its copy and nothing
   notices.
2. **It does not read citations.** `numberSupported` (`:~1395`) consults only `f.Value`,
   `f.ContextTerms` and `f.Tolerance` — **never `f.Source`**. So a fact carrying a value and no URL
   arms the scan *exactly as fully* as one carrying a verified quote.

**Therefore the `standard` bar is the ATTESTED register:**

| | `standard` (attested) | `sourced` / `relied_upon` (cited) |
|---|---|---|
| facts carry | `value` + `context_terms` (+ `tolerance`) | that, **plus** `source.citation{url, quote, title, publisher, accessed}` |
| what it arms | `ScanUnregisteredNumbers` — fully | that, **plus** the nightly quote re-check |
| nightly fetch | **none — so no `citation_lost` drift risk at all** | every citation URL re-fetched and re-matched |
| who vouches | an attestation: who stated the figure, when (the Q4 "record, not flag" shape) | the primary source |
| `banned_claims` | yes — the sector set applies at both rungs | yes |
| cost | **hours**: the figures are the site's own, there is nothing external to read | ~half a day: the expensive half is reading the primary source |

The asymmetry is not a concession, it is the correct shape: **a `standard` site's claims are about
itself, so there IS no external authority to cite.** "We have built 40 sites" cannot be verified
against a URL; it can only be attested. Requiring a citation there would either produce fabricated
sources or an empty register — and an empty register disarms the scan, which is the failure D4 exists
to prevent.

### 3g(ii). ⚠ The bound on D4's benefit, stated now rather than discovered after twelve registers

`ScanUnregisteredNumbers` is gated on `surface.ProseNumbersAreClaims()`, which returns **false** for
`editorialPageTypes` — `guide`, `blog-post`, `tool`, `game`, `news-index` — and for
`thirdPartyDataComponents`. **So an attested register does NOT arm the numeric scan over guide,
blog-post or tool bodies.** On a site that is mostly editorial, coverage falls on `content`,
`landing`, `section-index`, `entity-page`, `report` and the like.

**This bound is deliberate and measured, not a defect to route around.** Each membership was earned
by counted false positives on live copy against a real register (`blog-post` 46, `tool` 7, `game` 4,
`guide` 1+15, `news-index` 1 — `cmd/claimscan`, 2026-07-28), and that map's own comment sets the bar
for adding one: *"a measured false positive on live copy, AND a body that is never marketing. Do not
widen this from intuition."* `section-index` was measured and **rejected**; `blog-index` was never
measured and deliberately left out.

So D4 delivers: **numeric-claim coverage on the marketing surfaces of every site**, plus
`banned_claims` at both rungs, plus the absence of the disarming-by-emptiness hole. It does **not**
deliver number-checking inside guides and blog posts, and nothing in this ruling should be read as
promising that. Whether the editorial exclusion should narrow (per-component rather than per-page, as
RFC_053 Phase 2 already did for the tracker/directory three) is a **separate question with its own
measured bar** — it is not part of D4 and must not be bundled into it.

### 3g(iii). What D4 changes about the twelve

§1d's twelve register-less `deployed` sites are no longer "mostly not violations". Under D4 **all
twelve need a register** — but only the ones asserting external facts need the expensive one. First
pass at the split, `[INFERRED from domain and category — NOT yet measured per site, and the posture is
a Q4 RECORD that a human signs, not a guess a session makes]`:

- **`relied_upon` / `sourced` (cited bar):** `vetcomparison.uk` (D1, animal-health claims).
- **likely `sourced`:** `seotools.co.uk`, `cv1.co.uk`, `idea.uk`, `cookly.uk`, `lampenkap.com`,
  `garden-tools.uk`, `homegarden.uk` — each turns on whether its figures are about the world or about
  itself, which is exactly the question the ladder asks and which must be **read at the pages**, not
  assumed from the domain.
- **likely `standard` (attested bar):** `advertise.co.uk`, `designblog.co.uk`, `websitepromotion.co.uk`,
  `oxenunity.com`.

**Do not build to this list.** It is the shape of the work, not its scope: the rung is a Q4 record
carrying who declared it and on what basis, so each site's posture is a decision with a signature,
taken after reading what the site actually asserts. The absence detector (D3) is what produces the
queue; this list is only an estimate of how much of that queue is cheap.

---

## 4. What this lane has already done, so the RFC is not the whole answer

- **`fad209b92`** (council `1dd3d298`): regulatory rule citations no longer read as business numbers —
  **for the FCA Handbook only**; see §3b for why this is a precondition for (ii) on finance sites
  specifically, not sector-generally.
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

```sql
-- §3b: other regulated-sector sites the citation-code gap would hit (2026-09-02)
SELECT domain, status FROM sites
WHERE domain ~* 'vet|animal|pet|health|medical|clinic|dental|nhs|pharma|law|legal|solicitor'
ORDER BY domain;
```
