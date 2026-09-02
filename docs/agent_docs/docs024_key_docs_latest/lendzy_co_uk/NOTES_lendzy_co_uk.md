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
