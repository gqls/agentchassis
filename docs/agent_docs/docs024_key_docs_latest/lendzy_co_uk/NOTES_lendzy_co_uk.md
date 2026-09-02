# NOTES — lendzy.co.uk lane

Running record, append-only, **newest at the bottom**. Missteps are the point, not an appendix.

Site: `lendzy.co.uk` = `8ff093d5-1f19-453b-9439-a10379bbcd76`.
**Counts carry the date they were counted** (owner ruling 2026-08-22).

---

## 2026-09-02 (a) — the lane is created, and what it inherited

Created today at the owner's instruction ("this is lendzy's own lane now"). Until now lendzy
had **no lane**: it was built by `portfolio_positioning` as the framework's first end-to-end
shadow build (seeded 2026-08-02), and four lanes have since worked pieces of it and handed
them off:

- `bugfix_414_planted_marker_as_claim` — the planted FCA marker. **CLOSED 08-31** (`de99599fb`),
  fixed and live and verified. Re-checked today: **0** components carry the phrase.
- `copy_quality_two_stage` — holds lendzy's copy. Their standing read: the adversarial-adjacent
  frame is EARNED, leave it. Two residuals of theirs: `key_differentiators` shared verbatim with
  `loancash.co.uk` (a copy-paste at adoption), and 5 `voice_tells` items filed 09-01, unreviewed.
- `dispatch_throughput` — used lendzy as the fleet's worst-starvation baseline (55 eligible,
  oldest 10.6h, pinned rank 44). Migrations 657/658 fixed the ordering; lendzy drained 46 → 15.
  Measurement subject, not a defect. Closed for us.
- `architecture_review/RFC_060` — filed by the 414 lane 09-02, owner-decided the same morning.
  Lendzy is the worked example of the top `relied_upon` rung.

## 2026-09-02 (b) — three tool pages serve 200 and are recorded as never built

Found while grounding the lane. `[MEASURED 2026-09-02]`

| page | serves | `build_status` | `deployed_at` |
|---|---|---|---|
| `tool-price-cap-checker` | **200**, 65,356 B, 3 `<input>` | `needs_rebuild` | **NULL** |
| `tool-true-cost-calculator` | **200**, 63,999 B, 1 `<input>` | `needs_rebuild` | **NULL** |
| `tool-complaint-deadline-calculator` | **200**, 63,997 B, 2 `<input>` | `needs_rebuild` | **NULL** |

Six sibling tool pages are `deployed` with stamps written 09-01 10:58–11:03Z. All nine serve;
all five probed carry a correct `rel=canonical`. **Invented-URL control 404s (9 B)**, so the
200s are real pages and not a parked-domain catch-all.

Two measured consequences, both from `deployed_at IS NULL`:
- **Missing from the sitemap.** 30 active pages, **27** `<loc>`. The three missing are exactly
  these. `render_sitemap_action.go:144` filters `deployed_at IS NOT NULL`.
- **47 `unbuilt_internal_link` items** at `needs_human_review`, all filed 2026-09-01, every one
  naming one of these three URLs, every one reading "points at a page that has never been
  deployed". Nothing is queued to rebuild them, so the queue cannot drain itself.

## 2026-09-02 (c) — the root cause, and the two hypotheses that died first

**Dead hypothesis 1: the deploy-skip guard refused the stamp.** `refuseDeployStampOnSkip`
(`page_build_failure_guard.go:78`) flips a page to `needs_rebuild` when the deploy step reports
`skipped`, which fitted the symptom well. **Refuted two ways.** (i) `agent_error_log` holds
**zero** `DEPLOY_STAMP_REFUSED_ON_SKIP` rows for this site — lendzy's only rows since 09-01 are
three `CTA_LABEL_MISMATCH`. (ii) The guard is opt-in on `deploy_result_field` and its own header
says the unarmed path is "every live step today". It never ran. *The lesson I nearly took: the
code that best fits a symptom is not evidence that it executed. Ask the log whether it fired.*

**Dead hypothesis 2: NULL `component_id` is the discriminator, fleet-wide.** True on lendzy
(3/3 stuck pages have one, 0/6 working ones do) and **false as stated**: the fleet carries
**16** such rows across **7** sites and **10** pages, and outside lendzy **none** is stuck.
Correlation inside one site is not a mechanism. The corrected predicate is the one below, and
it was chosen because it could have come out otherwise.

**The cause, and it is exact.** The discriminating property is not "has a NULL `component_id`"
but "**every** component row on the page has one — so there is nothing resolvable to build
from". Fleet-wide, active pages where no component row carries a `component_id`:

```
 lendzy.co.uk | tool-complaint-deadline-calculator | needs_rebuild | unstamped
 lendzy.co.uk | tool-price-cap-checker             | needs_rebuild | unstamped
 lendzy.co.uk | tool-true-cost-calculator          | needs_rebuild | unstamped
(3 rows)          [MEASURED 2026-09-02]
```

**Three of three, no counter-examples anywhere in the estate.** The query could have returned
healthy pages, or pages on other sites; it returned neither.

The chain, read at the deciding arm rather than inferred:

1. Each page has **one** `page_components` row, written **2026-08-02** (the original shadow
   build), with `component_id = NULL` and `slot_name = 'section'`.
2. `rerender_page_sections_action.go:361` `resolveComponent` resolves a section by
   `component_id` first, then by `slot_name` against a component name/function map. The id is
   empty (`COALESCE(component_id::text,'')`, line 924) and **no component is named, functioned
   or typed `section`** — verified, **0 rows** in `content_components`. Neither route resolves.
3. The slot lands in `UnresolvedSlots`. `rerenderResolution.fatal()` (line 650) counts that as
   fatal, and line 600 returns an error: *"N of N section(s) could not resolve a component"*.
4. So the page never reaches `build_status='deployed'`, and `UpdatePageStatusAction`
   (`v3_site_actions.go:1082`) only stamps `deployed_at` inside its `newStatus == "deployed"`
   branch. No status, no stamp.
5. The 2026-08-02 artefact is still in the bucket, so the URL serves 200 for ever. Every check
   that asks the **artefact** says healthy; every check that asks the **record** says never
   built. Both are reading correctly.
6. It cannot self-heal: `needs_rebuild` re-selects the page, the render fails identically, and
   it re-files. Six `page_rerender` items for these pages since 08-25 — **all `complete`**.

This is the residue of `bugs_closed/182`, not a regression of it: 182's fix made a silent
carry loud, and the loudness has no consumer for this shape.

**Filed to the diagnosis loop before asserting it durably** (owner ruling 2026-07-31, and this
is a cross-cutting mechanism claim): intake `1ff4c475-6977-4631-b641-993735429186`, run
correlation `89a84ad3-5668-44b3-a089-f9d6c0df7cbb`. Verdict to be recorded here when it lands
— **including if it refutes me**, which is the cheapest place to be wrong.

## 2026-09-02 (d) — the FCA ask, and the one thing measured so far

Owner: make the retracted claim TRUE — check all financial facts against the FCA Handbook line
by line, with a local mirrored copy kept current AND live re-checking ("probably as well").

Measured today, nothing else assumed: `https://www.handbook.fca.org.uk/handbook/CONC/5A/` 301s
to `https://handbook.fca.org.uk/handbook/conc5a` and returns **200, 477,729 B**, title
*"FCA Handbook - CONC 5A Cost cap for high-cost short-term credit"*, with rule identifiers in
the markup down to `CONC 5A.1.1`. No auth. So per-rule citation with a verbatim quote looks
mechanisable. `[MEASURED 2026-09-02]` **Everything else about the corpus is `[UNMEASURED]`** —
licensing/terms, rate limits, whether an instrument or release feed exists for change
detection, and how the page markup keys rules to text.

Lendzy has **no `evidence_base` at all** — one of the five register-less finance sites named in
`RFC_060` §1, so its numeric scan never arms. RFC_060 §4: *"If only one thing happens as a
result of this RFC"*, it should be populating those registers.

Wrote to the `claims verification` session (peer `d02867`) before designing anything, since the
owner named them as responsible here: asked what they own, whether the `evidence-refresher` is
theirs and is the right substrate at handbook scale, whether anyone already mirrors an external
regulatory corpus, and where they want the boundary between per-site register work and platform
verification. **Design held until they reply** — building a second spelling of their mechanism
is the failure mode to avoid.

## 2026-09-02 (e) — the claims-verification thread's answers, and what they change

Replied within the hour. Their answers, attributed, because the design rests on them:

- **They own the register/scan/audit layer** (CLM-025…CLM-030,
  `docs026_concept_register/register/claims-verification.md`) and filed `RFC_060`.
- **The daily refresher is theirs** — `refresh_evidence_base_action.go` + `evidence_citations.go`
  (CLM-007/008). One fact = one URL + one verbatim quote, re-fetched daily, quote re-checked against
  visible text, 403/5xx classified *unknown* and never drift. **This is the owner's "check with
  their online version each time", already built.**
- **`fad209b92` already excludes regulatory citations** (CONC, MCOB…) from the business-number
  scan, so lendzy's "0.8% per day under CONC 5A" shape needs nothing new and nothing waited on.
- **Where it breaks at handbook scale:** `fetchCitationDocument` is one unthrottled
  `http.DefaultClient.Do` per fact — no delay, no caching, no dedup across facts sharing a URL.
  Fine at the fleet's **39** citation facts (their measurement, 2026-09-02); not fine when one site
  cites dozens of handbook rules daily. **Pace it before scaling, not after being blocked.**
- **No FCA mirror exists**, but the Companies House enrichment pipeline
  (`017_companies_house_enrichment.md`) is the pattern: bulk-collect an external authority into a
  local table, paginated, **deliberately throttled to ~7% of the published cap**, scheduled, then
  queried locally. Build the mirror as its sibling, **not** inside `evidence_base`.
- **The boundary** is already drawn: "a fact cites a rule" is register content (no code); "the
  platform re-verifies the citation" is CLM-008. Genuinely new: *discovering* which rules changed
  (nothing does this — the refresher only re-checks quotes a human already chose to cite), and
  pacing.
- They folded the sector-generalisation question into `RFC_060` §3b as Q5, and asked that design
  docs route through their thread or cite RFC_060 so the estate does not grow two accounts of one
  register mechanism.

## 2026-09-02 (f) — the machinery WORKS against the FCA Handbook, proven with the production matcher

The question that had to be answered before writing a single fact: **would a quote we store
actually match what the refresher extracts?** A quote that does not match is classified
`citation_lost` — drift — **every day, for ever**, and a false alarm is indistinguishable from a
real one.

I did **not** answer it with my own regex extraction, because a mirror of the extraction passes
happily while production disagrees. `cmd/fcaquotecheck` calls the real
`datahelpers.VisibleTextFromHTML` (i.e. `ExtractAssertionText`) and `datahelpers.QuoteFoundInText`.

`[MEASURED 2026-09-02]` against `https://handbook.fca.org.uk/handbook/conc5a` — HTTP 200,
477,729 raw bytes reducing to **44,121 visible characters**:

| quote | matched |
|---|---|
| `exceed or are capable of exceeding 0.8% of the amount of credit provided under the agreement calculated per day` | **true** |
| `exceed or are capable of exceeding the amount of credit provided under the agreement` | **true** |
| `cumulatively in relation to multiple breaches of the agreement) exceed or are capable of exceeding £15` | **true** |
| a deliberately absent control string | **false** |

Positive and negative controls in the same run. **So the FCA Handbook is compatible with the
existing citation machinery as it stands, and Phase B-i needs no code at all.**

The chapter also parses cleanly into individual rules — `[MEASURED]` **78** rules in CONC 5A,
each with its id, effective date and R/G type, by splitting visible text on
`CONC 5A.\d+.\d+ dd/mm/yyyy [RG]`. That is what makes "rule by rule" mechanical rather than
aspirational.

⚠ **The URL scheme is NOT uniform** and a collector must not assume one form. `handbook/conc5a`
works; `handbook/conc6/7` does **not** (bare title — a miss, caught by the landmine's own control);
`handbook/CONC/6/7.html` works. Confirm every fetch by title, never by status.

## 2026-09-02 (g) — TWO OF LENDZY'S RULE CITATIONS ARE WRONG, and this is the ask paying for itself

Surveyed every page's factual claims and checked each cited rule against the handbook text.
`[MEASURED 2026-09-02]`

**Correct, verified by section title and text:** `CONC 5A` (0.8%/day, the 100% total cost cap, the
£15 default cap) · `CONC 5.2A` = *Creditworthiness assessment* · `CONC 7.3` = *Treatment of
customers in or approaching arrears or in default* · `DISP 1.6` = *Complaints time limit rules* ·
`DISP 2.8` = *Was the complaint referred to the Financial Ombudsman Service in time?*

**Wrong:**

| the site says | the handbook says |
|---|---|
| rollover-rules: *"Financial Conduct Authority (FCA) rule **CONC 6.7.17** limits how many times a firm can refinance"*, and *"This is an FCA rule called CONC 6.7.17"* | **CONC 6.7.17** is the DEFINITIONS rule — *"In CONC 6.7.18 R to CONC 6.7.23 R 'refinance' means to extend…"*. The two-rollover limit is **CONC 6.7.23**: *"A firm must not refinance high-cost short-term credit (other than by exercising forbearance) on more than two occasions."* |
| cant-pay and your-rights: the two-attempt card limit *"is set by FCA rule **CONC 6.7.23**"*, and *"Trying a third time breaches FCA rule CONC 6.7.23"* | **CONC 6.7.23** is the refinance limit above. The continuous-payment-authority limit is **CONC 7.6.12**: *"a firm must not request a payment service provider to make a payment, under a continuous payment authority … if it has done so … on two previous occasions and those previous payment requests have been refused."* |

**Both substantive claims are TRUE** — two rollovers, two card attempts. Only the attributions are
wrong, and they are wrong in a shifted pattern: the rollover claim points at the definitions rule,
and the CPA claim points at the rollover rule.

Honesty about strength: the rollover one admits a generous reading, since CONC 6.7.17 *introduces*
the range "CONC 6.7.18 R to CONC 6.7.23 R" and so is a pointer to the rollover block — "imprecise"
is arguable there. **The CPA one is not arguable**: CONC 6.7.23 is about refinancing and says
nothing about card payments.

This is the whole ask demonstrated on day one. A site telling people in financial difficulty which
rule protects them, citing the wrong rule, is exactly the damage a register prevents — and note
that **no existing check could have caught it**: every claims detector asks whether a claim is
supported, and none asks whether the *rule number* is the rule that says it.

**NOT YET FIXED.** The wrong numbers are in served copy, so repairing them changes published prose,
and the estate is deliberately careful about automated rewrites of published copy. Recorded here
and put to the owner rather than fixed unilaterally.

## 2026-09-02 (h) — the 090 verdict came back UNVERIFIABLE, and it did NOT confirm my root cause

Run `89a84ad3` finished at iteration 5 with:

> `"status": "UNVERIFIABLE"` — *"Diagnosis NOT confirmed (stopped: iteration-cap). Best-effort
> trail attached for a human; no fix proposed."*

**This is not a confirmation and I am not going to write it up as one.** Nor is it a refutation.
The loop spent its iterations on pages that are not lendzy's stuck three — `llm-cost-calculator`,
`tool-ai-vendor-trust-checklist`, `learning-center-hub`, none of which is on this site — and its
own final bundle says why that trail was worthless: those page names recur across sites, every row
it found was already `deployed` with a stamp, and *"the actual current-state row for lendzy.co.uk's
own … pages was never successfully retrieved (the site-scoped query errored with `column reference
"id" is ambiguous` and returned 0 rows)"*. So it never looked at the three rows the symptom names.

**What this means for the claim, stated plainly under the 2026-07-31 owner ruling.** That ruling
requires a cross-cutting root-cause claim to go through the loop **or** for the filing session to
say why it substituted equivalent first-hand verification. I did run the loop, it returned nothing
usable, and I am therefore relying on the first-hand verification — which is recorded in (c) and
consists of:

1. the exact rows: **3/3** stuck pages carry a single `component_id IS NULL` row, **0/6** healthy
   siblings do;
2. a **disconfirmable** fleet census — pages where *no* component row carries a `component_id` —
   which could have returned healthy pages, or pages on other sites, and returned **exactly these
   three and nothing else**;
3. the code path read at the **deciding arm**, not summarised: `resolveComponent`'s two returns,
   `fatal()`'s body, and the `newStatus == "deployed"` branch that is the only caller of the stamp;
4. the missing component verified as an **absence** (`0` rows named/functioned/typed `section`),
   not assumed;
5. **two competing hypotheses actively refuted** with their own evidence (the deploy-skip guard, by
   the empty error log and its unarmed opt-in; and the naive `component_id IS NULL` predicate, by
   the 16-row/7-site census that showed it does not discriminate).

That is stronger than what the loop produced, but it is *my* verification and it carries my
fallibility. The honest status of the root cause is **first-hand verified, loop UNVERIFIABLE**, and
any document quoting it should say so.

⚠ **Do not read the UNVERIFIABLE as evidence about the platform.** It is evidence about that run:
a broken query and a wandering hypothesis. A re-file with a tighter symptom (naming the three page
names and the site id explicitly, since the loop demonstrably lost the site scope) would be a
reasonable next step for anyone who needs the independent check — and is cheap.

**Correction to my own earlier expectation, recorded because the missteps are the point:** in (c)
I wrote that the verdict would be recorded "including if it refutes me". I framed the outcomes as
confirm-or-refute and did not consider the third, which is the one that happened — the loop failing
to engage the question at all. A run that produces no verdict is not a neutral event: it costs the
time you spent waiting for it, and if you are not careful it gets quoted later as though the
absence of refutation were support.

## 2026-09-02 (i) — 693 round 2 REVISE: two real defects and a named sibling class, all answered by measurement

The council's round 2 found what round 2 deserved to be caught on:

- **The rerender INSERT was broken twice over** (prior_art_librarian, HIGH): it omitted the
  first-class `page_id` column — the live producer sets it on **all 7,481** of its rows
  `[MEASURED]` — and omitted `created_by` entirely, which is NOT NULL with no default, so the
  round-2 migration would have **errored at apply**. Both fixed in 693 AND pre-emptively in 696,
  whose own council round had not yet reported the same defect it inherited from the same author.
- **The named prior class exists**: `bugs_open/357` — whole working tools stored under a FALSE
  component identity, one row per page. Their arm: component_id points at the shared `hero`, so
  regeneration would swap a 16KB tool for a 2KB title band, and their park is load-bearing. Our
  arm: component_id NULL, so nothing resolves and the page can never deploy. Structural point that
  matters both ways: **adoption makes the declared template byte-identical to the stored tool, so
  the regeneration that is their disaster is our no-op** — and may be their fix shape too
  (CONTRIB'd to their session; their 090 run 63d4d1a7 is still pending, not prejudged).
- **bug_historian's "will it actually resolve"** was answered with the production function: all
  three adopted bodies PASS `toolTemplateValid` (the guard `loadComponentSchemasByID` applies at
  `component_level='tool'`), truncated control REJECTED, via a one-shot in-package test (created,
  run, deleted — never committed).

**Misstep, logged because it is the point: my first probe control was VACUOUS.** I passed a
31-char truncated string as the must-fail control; `toolTemplateValid` allows anything under
**100 chars** through as a deliberate stub, so the control PASSED and the run proved nothing — and
the three PASSes I wanted were sitting right there looking like a result. The probe's own
control-failed arm is what caught it. *A control must be able to fail for the reason the real case
would fail, not merely be wrong-shaped* — a control under the guard's stub floor tests the floor,
not the guard.

Also recorded: the 47 items need no hand-retraction — `unbuilt_internal_link` carries a registered
verifier (`check_phantom_internal_links.go:451`) judging by `NeverDeployedPagePredicate`, so they
resolve on revalidation once the pages stamp. Hand-closing them would blind the mechanism that
filed them.

## 2026-09-02 (j) — the register programme's first day, fleet-wide: four sites done, five wrong claims found

The owner's decision 5 propagated faster than expected — both notified lanes executed same-day:

| site | migration | facts | wrong claims found in live copy |
|---|---|---|---|
| lendzy.co.uk | 695 (council r2 in flight) | 8 | **2** — the 6.7.17/6.7.23 pair (this lane, (g)) |
| loanzy.uk | 697 (applied) | CCA s.66A, StepChange, MaPS | **1** — MaPS grouped under "FCA-authorised services"; it is the statutory guidance body, not an FCA firm |
| farmerinsurance.uk | 698 (applied; supersede-and-merge preserving existing facts) | ICOBS 8.1.1, DISP 1.6.2/3.6.6, ELCI 1998 reg 3 | 0 (insurance sources; method held unchanged) |
| loancalculator.co.uk | 699 (applied) | 12 (9 statute: CCA 1974 + SIs) | **2** — "ten working days" that is 12 (SI 1983/1564 reg 4); an invented "10%/12mo" ERC-free threshold that is £8,000/12mo (s.95A(2)(a)) |

**loancash.co.uk is the remaining gap — no session** (and it serves lendzy's propagated wrong
rollover cite, which 696 fixes as a specific error; its register remains unwritten).

New traps contributed by the peers, folded into RUNBOOK §8b/8c with attribution: Cloudflare-
challenged hosts (maps.org.uk, moneyhelper.org.uk) pass a human eyeball and fail unattended for
ever — refuse them at write time; gov.uk founding-name slugs; legislation.gov.uk 200s wrong paths
exactly as the FCA host does; near-identical 1983 SI names (1569 vs 1564 — only reading the
schedule discriminates). The host-admission rule is forwarded to claims-verification as design
input for the registration path and the mirror.

**The base rate, stated because it is now measured and it reframes the programme:** every lane
that ran the method found errors. Five wrong live claims across three sites in one day is not a
tail risk; it is what "unchecked citations" means at this fleet's scale. The Q6 checker and the
mirror are being built against that rate, not against a hypothetical.

## 2026-09-02 (k) — the roll, the apply, and the RAISE that aborted it

**The chassis rolled at ~15:39Z** (new replicaset `744cfb4bf`; the whole agent fleet cycled behind
it). My monitor armed seconds AFTER the new pods existed, so its roll-detection baseline was
post-roll and would never have fired — caught by reading the baseline timestamp in the armed
message against the pod age measured half an hour earlier. *A watcher's baseline can race the event
it watches; read the baseline as data, not as setup.*

Both in-flight council runs (693 r3, 695 r2) froze at the roll instant and stayed ~10 min stale
with their executing pods gone — **resubmitted on their existing correlations** per the estate
precedent, trail kept.

**696 applied at ~15:46Z, on the second attempt.** The first apply ABORTED at the verify block:
`RAISE EXCEPTION` with a parameter but no `%` placeholder is a **compile-time** error in PL/pgSQL,
so the single wrapped transaction rolled everything back (verified pre-state `4 | 0 | 0` before
retrying — the abort cost nothing but minutes). A mechanical audit of every RAISE across all six
lane migration files found exactly TWO mismatches, both in 696's pair, both fixed (`9e5a689b7`).
*The lesson: the DO/RAISE discipline this lane preaches has its own failure mode — a RAISE that
cannot compile fails the migration for the wrong reason at the last block; audit placeholder
counts mechanically, not by eye.*

Second apply, clean:
> `696 OK: CONC 6.7.17 extinct; 10 CPA cites now 7.6.12; spec superseded; 2 tool templates
> corrected (lendzy + the loancash fork); 11 rerenders queued` — plus `bak_696_citation_surgery`
> (9 rows) and `bak_696_component_templates` (2) kept.

Watcher armed on the 11 rerenders; the POST-APPLY artefact check runs when they are all terminal.
695/693 remain UNAPPLIED, awaiting their (resubmitted) verdicts.

## 2026-09-02 (l) — 696 VERIFIED AT THE ARTEFACT: the wrong citations are off the wire, fleet-wide

All 11 rerenders reached `complete` within ~20 minutes of apply, each with its own deploy
commit sha. The committed POST-APPLY check then ran against served bytes, with the invented-URL
404 control alongside. `[MEASURED 2026-09-02 ~16:0xZ]`

| check | result |
|---|---|
| 7 content pages, `CONC 6.7.17` | **0 on every page** |
| rollover-rules positives | `CONC 6.7.23` ×2 |
| cant-pay positives | `CONC 7.6.12` ×2 |
| both guides (at their **recorded** `pages.url`) | clean; corrected cites present |
| lendzy tool | 6717=0, 6723=1, **inputs=8 unchanged** |
| **loancash tool** (the propagated fork) | 6717=0, 6723=1, **inputs=8 unchanged** |
| invented-URL control | 404 |

Two working notes from the check itself:
- My first pass at the guides used **composed** URLs and read 404 bodies (size 9) — the exact
  bug-387 trap. The recorded `pages.url` (`/guides/…`) was the answer; RUNBOOK §2's discipline now
  says so explicitly by example.
- The CPA guide "stored 2 / served 1" alarm was my own grep's strictness: the served page carries
  **2** bare `7.6.12`, the second not "CONC "-prefixed in serve form. Count with the loosest
  pattern that is still unambiguous before treating a delta as loss.

**Owner decision 1 is DONE end-to-end**: wrong rule numbers extinct in storage AND at the served
artefact, on both sites, with working calculators unchanged. Remaining lane work rides on the
resubmitted 693/695 verdicts.

## 2026-09-02 (m) — 693 ACCEPTED at record and artefact; the lane's founding defect is closed

Applied ~16:01Z (second attempt: the doc_notes `subject_type='site'` value failed the table's
CHECK constraint — allowed values are tool/pipeline/experience/action/experience-pattern/landmine/
component/decision — first transaction aborted cleanly; fixed to `decision`, and the identical
latent bug in 695's INSERT fixed in the same commit before 695 could hit it). Three rerenders
drained by ~16:07 through the twice-rolled dispatch loop.

`[MEASURED 2026-09-02 ~16:1xZ]`

| acceptance | result |
|---|---|
| record | all three pages `build_status='deployed'` with `deployed_at` 16:06–16:07Z — **first stamp of their lives**, 31 days after they began serving |
| artefact | 200 ×3, **inputs exactly 3/1/2** (the baseline — no calculator swapped), invented-URL control 404 |
| fleet invariant | verify's own re-run: **0** active pages anywhere with no identified component on any row |
| sitemap | still 27 — CORRECT for now; the sitemap follows the deploy on its own rotation (SEO-007/642) and the three now satisfy `deployed_at IS NOT NULL`. Check after lendzy's next sitemap-refresh tick: expect **30** |
| the 47 items | still `needs_human_review` — they drain when revalidation re-judges them via `VerifyUnbuiltInternalLinkResolved` against the now-stamped pages. Deliberately NOT hand-closed |

The PBP-038 advisory from round 3 is also now double-answered: 693's own rerenders stamped through
the same path 696's did, same site, same day, fresh chassis.

**695 remains the only unapplied piece.** Its resubmitted council run froze at the second roll
(review_tooling_provenance, 15+ min stale) and the fleet was still cycling new pods at 16:1x —
resubmission waits for a stable window (monitor armed) rather than feeding a third run to a roll.

## 2026-09-02 (n) — 695 APPLIED: every migration is live, and the day's programme is complete

Approved round 2 (2 advisories + several lows), applied ~17:0xZ after acting on them: two new
verify arms assert the STORED regex form mechanically (zero double backslashes; the single-escaped
`\b` present in all 5 — the double-escape landmine where a pattern compiles and matches nothing can
now fail the migration instead of shipping), and the rollback header states the real refusal
mechanism (the refresher SUPERSEDES, so a post-refresher rollback aborts at the zero-rows check,
correctly). Read-back: **8 facts | 5 banned patterns | `\bguaranteed…` single-escaped | created_by
the lane.**

The banned set is the calibrated adaptation of adversecreditmortgage's ((n-1), RUNBOOK §8b/8e):
four adopted, `no credit checks?` narrowed on the measured UI false positive, the literal-rate
pattern deliberately absent with the measurement recorded inside the register.

**What arms from here, on schedules not ours:** the daily `evidence-freshness` tick (last ran
09:08Z) starts re-verifying all 8 quotes against the live Handbook tomorrow morning; the numeric
scan arms on lendzy for the first time — and per the 414 lane's sizing, arms CLEAN (zero unbacked
business numbers post-`fad209b92`); the sitemap picks up the three repaired tool pages on its next
rotation (27 → 30); the 47 `unbuilt_internal_link` items drain as revalidation re-judges them.
RUNBOOK's standing instruction: **read the first `claims_unverified` findings rather than assuming
silence is cleanliness** — a register's first live pass is a measurement, not a formality.

End of day one. Three migrations written, reviewed (7 council rounds between them, every REVISE
answered by measurement), applied, and verified at the artefact where an artefact exists.

## 2026-09-02 (o) — the adoption shape travelled: 357's owner ruled for it, and their drafting answered our open question

The 357 lane's migration 701 (in council, corr df6c1b41) adopts all 22 of their wrong-id tool
pages off 693's crib — with the third repoint leg their plan-driven case needed
(`site_plan_sections.component_name` + the derived `pages.sections` copy). Both of this lane's
gotchas held up in their drafting: 22/22 bodies passed `toolTemplateValid` through the real
function with a genuine >100-char must-fail control, and their rerender filing copies the
`page_id` + `created_by` shape.

Two of their measured facts banked here, with attribution, because they answer questions this lane
left open:

- **`pages.sections` is a DERIVED copy** — sync overwrites it from `site_plan_sections`, so the
  durable leg of any plan repoint is the PLAN table, never `pages.sections`. Matters to this lane
  only if lendzy ever grows populated plans (today 8 of 9 tool pages carry `[]`), but a future
  session editing `pages.sections` as though authoritative would watch its edit evaporate at the
  next sync.
- **The drift-reconciler interaction I flagged as unmeasured is now measured** (by them):
  the comparison is `built_from_plan_version` (a `site_plans.id`) vs the current plan id, and
  plan-ELEMENT edits create no plan row — so an element-level repoint causes no rebuild storm.

Their second-fork finding on `tool-equity-release` (an unplaced deploy-path copy beside the
library parent on mortgagecalculator) is their tail, disclosed to their council — noted here only
as the RFC_036 §11 shape recurring.

## 2026-09-02 (p) — the live patterns probe-fired through Go itself: the silent-degrade class is excluded

The 414 lane corrected a landmine (`4f1ca1384`): a per-site banned pattern that fails to compile
degrades **silently to a literal of its own source text** (`claims.go:348`) — no logger, no error,
counted as armed by every count-based check. My 695 verify asserted stored FORM (backslash
discipline) but never compiled or fired anything, so `[MEASURED 2026-09-02 ~18:0xZ]` the five live
patterns were pulled from the current row and run through Go's own `regexp.Compile`, each with a
must-match arm (its target phrase) and a shared must-NOT-match arm — a clean control sentence
deliberately containing this lane's two known false-positive fragments ("Yes No. Check for a
breach", "3% per month, roughly 42.6% APR"):

> `ALL 5 LIVE PATTERNS: compile, fire on their target phrase, stay silent on legitimate copy`

Fleet context from the same lane's baseline: 239 live patterns across 19 sites, 0 non-compiling —
nothing has gone wrong yet; the point is nothing would have said so if it had. This lane's ladder
for pattern verification now reads: stored form (695's verify, at apply) → Go compile + probe-fire
(this check, post-apply) → the first real `claims_unverified` findings (RUNBOOK's standing read).

Also from their census: **farmerinsurance carries 7 facts and ZERO banned_claims** — a register
that reads "done" in any facts-count check while enforcing nothing at the build gate. Relayed to
the loanzy.uk session, which holds that site's lane (correcting the 414 lane's "no lane owns it").

## 2026-09-02 (q) — register programme end-state: every owned seat filled, and the method came back improved again

farmerinsurance's banned set landed (their migration 713, corr cea2a32c at the gate): five
INSURANCE-SHAPED patterns designed as a sector set rather than the credit set transplanted — the
first deliberate sector design, which is itself a data point on RFC_060's shared-set question. The
full ladder ran in order, and step (1)'s count reconciliation caught a REAL defect on their own
site mid-calibration (the 18th active page, /claims.html, serves 404 — their bugs_open/437).

Two of their structural absences cite this lane's findings, and the second is a NEW pattern-
authoring rule worth keeping: no literal-rate pattern (§8e's no-citation-exemption), and **no
local copy of a pattern a fleet-wide check already owns** (first-person FCA-authorisation belongs
to CGV-033 — a per-site copy would SHADOW the fleet check, splitting one judgement across two
homes).

**Scoreboard at close `[MEASURED 2026-09-02 evening, per-lane reports]`:** loancalculator 12/8 ·
lendzy 8/5 · farmerinsurance 7/5 · loanzy 3/5 · **loancash absent — the fleet's only empty seat,
owned by nobody.** Four sites went from unregistered to citation-backed-and-guarded in one day,
each lane improving the shared method as it passed through.

## 2026-09-02 (r) — spec.reason is PARSED, not read: my 14 rerenders were assemble-mode no-ops, and what that does and does not undo

The components session's fleet sweep found all 14 of this lane's rerenders (696's 11, 693's 3)
carried a prose sentence in `spec.reason`. `page-rerender` branches on that field ALONE against
five literals — `image_landed · section_data_resolved · cta_links_stale · template_changed ·
literal_markdown` — and anything else takes `else_step → render_page`: simple concatenation,
re-shipping the STORED rendered_html byte for byte. All 14 completed, stamped, and re-rendered
nothing.

**What that does NOT undo, and why — measured, not argued:**
- **696 is unaffected in outcome.** The migration corrected `rendered_html` directly in the same
  transaction, so byte-for-byte re-shipping shipped the CORRECTION — this was the render_guardian
  advisory's exact belt-and-braces scenario, and the post-apply artefact check (NOTES (l)) plus
  tonight's re-confirmation (rollover-rules 6717=0/6723=2, your-rights 6717=0/7612=1) prove the
  wire is right.
- **693's stamps are real.** The assemble path still runs UpdatePageStatus's deployed branch —
  deployed_at 16:06–07 stands, and with template==stored-bytes by design, the served content is
  correct either way (inputs 3/1/2 verified).

**What it DOES leave open, and the honest correction to (m):** the adopted components have NEVER
been resolved by a live template render — the assemble path bypasses `resolveComponent` entirely,
which is the very path that used to fail. (m)'s implication that the rerender exercised the repair
was wrong; it exercised the stamp, not the resolution. **The discriminating proof is now filed:**
three rerenders with `reason='template_changed'` (the literal that routes to
`rerender_page_sections`), prose moved to `summary`, `source='manual'` per the provenance
convention. On the OLD state this exact path went fatal; if these three complete, the adopted
components resolve in production. Watching.

Also a self-correction on foreseeability: 696's approval advisory NAMED the unrecognised-reason
fallthrough and I recorded "safe by construction" — true for 696, and then I reused the same prose-
reason shape in 693 where the fallthrough silently skipped the resolution proof. A fact learned
about one migration was not carried to its sibling written the same hour.

## 2026-09-02 (s) — the components session struck their urgency claim, and handed back one caveat that upgrades the proof read

They struck the "compliance-shaped and urgent" consequence in the shared ledger (strike-through,
not quiet softening — the three-way split beneath it: 701 correct by construction, 696 correct
because the migration wrote the artefact, 693 where the finding holds). Their distilled rule is
worth carrying: **ask what WROTE the artefact before concluding a no-op cost anything.**

**The caveat that changes how I read the pending proofs:** they hold rows tonight that completed
TWICE with `reason='template_changed'` and produced pre-fix output both times (their open defect).
So a recognised reason + completion may still mean the template path silently did not run. For my
three proofs the output side cannot discriminate anyway — the adopted template IS the stored bytes,
so "rendered through and produced the same bytes" and "never rendered" are output-identical BY
DESIGN. The read when they land is therefore layered:
1. completion vs the historical fatal (primary — this path errored on the pre-693 state);
2. the item `result` payload's section counts if the action surfaces them (reRendered vs carried
   names which branch actually ran);
3. `page_components.updated_at` bumped + served inputs still 3/1/2;
and if (2) is not surfaced, the honest status is "resolution proven to NOT ERROR; proven to have
RUN only as far as the result payload discriminates" — recorded as such, not rounded up.

## 2026-09-02 (t) — the template-path proof LANDED: 693 is closed with no asterisk

The three `template_changed` rerenders completed (~21:33Z), and the layered read closed every gap:

1. **Routing**: `template_changed` is one of the five literals and routes to
   `rerender_page_sections` — the action where an unresolved component is FATAL, and which errored
   on these exact pages in their pre-693 state. **Completion is resolution.**
2. **The artefact, by HASH not proximity**: each proof-run's `files_sha256` equals the served
   page's sha256 exactly — `a76037f9…`, `2797d4ee…`, `bb0fbb93…` — so the bytes on the wire ARE
   those runs' own output. Inputs 3/1/2 throughout.
3. The result payloads carry the full `rendered_page.html`, deploy commit and per-file sha —
   which is what made (2) possible without inference.

**And the populated-plan complement arrived from production itself**: bugs_open/357 is CLOSED at
population 0 — all 22 pages, 21 adoptions + one §9.3 fork — and hours later a news wave
organically rebuilt vetcomparison/index end-to-end through plan → resolve → RENDER, reproducing
the adopted tool to within one trailing newline through a full delete-and-reinsert. The
template-equals-bytes property is now proven on both arms: our NULL-id/empty-plan case by the
directed proof above, their wrong-id/populated-plan case by an unprompted production rebuild.

Their parting caution (that 693's original rerenders likely re-shipped stored bytes) was correct
and is (r)'s finding — with the stamps unaffected (assemble also reaches the deployed branch,
measured at (m)) and now superseded by the real render above.

**Day-one ledger closes: every owner ask delivered, every claim at artefact strength.**
