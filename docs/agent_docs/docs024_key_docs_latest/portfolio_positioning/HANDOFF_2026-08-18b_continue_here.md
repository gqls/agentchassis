# HANDOFF — ⛔ BUILDS ARE HALTED pending two coupled owner decisions — 2026-08-18 evening, continue here

> **⚠ SUPERSEDED 2026-08-18 late evening by `HANDOFF_2026-08-18c_continue_here.md`.**
> Accurate on its own history and §1–§2 (the halt, the two decisions) still hold — but read
> 18c first: the missing-tools cause is found and CONFIRMED (`bugs_open/311`), the directory
> is 25 lenders not 2 (migration 471), the flow decision was narrowed by owner ruling P11,
> and §4's "www resolves NOWHERE across all 36 zones" is **corrected** there — 8 of 39 zones
> carry a www record, in four different states.


Supersedes `HANDOFF_2026-08-18_continue_here.md` (accurate on its own history; this file carries
everything a fresh chat needs). **Phase C is SIGNED OFF.** Read §1 before touching anything.

## 1. ⛔ THE HALT — what it is, and how to lift it

**Owner, 2026-08-18:** *"Stop the builds until we sort out the classifier and which builder flow
we are using."*

Implemented with the platform's **own** pause switch, not by editing work items:
`sites.locked_at` is exactly what `build-pipeline-trigger.find_dispatchable_site` excludes on
(`WHERE s.locked_at IS NULL`). Nothing auto-clears it (verified — the `locked_at` code paths in
Go are all `site_components`/`pages`), so it is durable, not a lease.

| site | locked | queued work HELD | HITL | failed |
|---|---|---|---|---|
| `adversecreditmortgage.co.uk` (build #1) | ✅ | **41** | 1 | 0 |
| `remortgagecalculator.uk` (pilot) | ✅ | 0 | 15 | 6 |
| `loanzy.uk` | ❌ not ours — owner's example-site thread | 1 | 11 | 0 |

**Queued work is preserved, not cancelled.** To resume:
```sql
UPDATE sites SET locked_at = NULL, locked_by = NULL
WHERE domain IN ('adversecreditmortgage.co.uk','remortgagecalculator.uk');
```
Build #1 then continues from its 18 `needs_page` + 20 `needs_imagery` items. **Do not lift the
halt without the owner's decisions in §2.**

## 2. THE TWO DECISIONS — coupled; decide the flow FIRST

Write-up: `DECISION_2026-08-18_two_builder_flows_side_by_side.md`. RFC:
`architecture_review/RFC_037_the_classifier_cannot_see_its_siblings_….md`.

**(a) Which builder flow.** Both flows use the same script and agent graph and cost the same
(~$3.81 text + imagery). They differ only in what is seeded first — and that decides which
guards exist:

- **Flow A (seeded + hand-written mission)** — pilot and build #1. Safe, proven twice,
  **45–60 min/domain ≈ 100 hours across the fleet.**
- **Flow B (prompt only)** — `loanzy.uk`. ~2 min/domain, and **measured: 20 pages built with
  `evidence_base` that has NEVER existed** (0 rows all-history), no email, no
  `imagery_style_guide`. `loadEvidenceBase` returns nil ⇒ **every claims lane silently
  no-ops**, so nothing has checked a single assertion on those 20 pages and no `banned_claims`
  pattern exists. The missing email additionally makes the hallucinated-email check **fail
  open** (`bugs_open/063`).
- **RECOMMENDED — flow B + an automatic seed.** The gap is a *seeding* gap, not a flow gap.
  Both hand-written seeds were boilerplate apart from the compliance specifics. Move the seed
  into the pipeline and you get B's cost with A's guards.

**(b) RFC_037 — the classifier reads the register.** Owner already chose **option 2** (feed the
register entry in). Filed not built: it adds an input to a shared seam every fleet site passes
through. **Measured case:** 7 finance sites → 2 distinct classifications, `industry` **null on
all 7**. Open questions in the RFC: where the register data lives (it is markdown; no agent can
read it), and whether the input is advisory or a binding collision check.

**The coupling:** under flow A the mission carries differentiation and RFC_037 is
belt-and-braces. Under flow B **RFC_037 is the only thing standing between 140 domains and
convergence** — a precondition, not an improvement.

## 3. What is DONE since the last handoff

- **PHASE C SIGNED OFF** (owner). The gate before Phase E is passed.
- **First 50 build order APPROVED** — `PLAN_2026-08-18_first_50_build_order_FOR_APPROVAL.md`.
  Wave 1 is one-at-a-time and supervised.
- **Build #1 dispatched (then halted): `adversecreditmortgage.co.uk`.** It planned **19 pages**
  — an eligibility tool plus a page per adverse-credit type (CCJ, default, DMP, IVA,
  bankruptcy, missed payments, repossession), which is exactly what the mission asked for.
  **Both directory proof points passed again:** flag written at classification, and
  `mortgage-lenders` page (exact name+type) composed `hero → mortgage-lender-directory-listing
  → call-to-action`, plus the homepage `mortgage-lender-directory` panel.
- **Its seed fixed the pilot's escaping bug** and proved the guards in **Go** (production's own
  compile path): 6/6 must-catch, 4/4 must-allow. The pilot's six patterns were inert while its
  verify passed by *counting* them.
- **Owner rulings on B8/B9/I10** — all three HOLDs lifted, recorded in the register **with
  shapes**: B8 app-shaped content (name spent knowingly); B9 standalone trade site (we sell no
  equipment — and its directory would be a NEW non-finance kind = a seven-place DIR-001
  addition needing its own decision); I10 differentiate-or-leave-unbuilt.
- **L9 reassigned:** `loanzy.uk` is an example site, no register entry, built from the
  webdesign.uk prompt — owner's separate thread. Resolves the Phase D conflict.
- **DNS: `remortgagecalculator.uk` zone CREATED and configured** (zone
  `c7ef25edb1221fb4ffc4d4dade271781`, proxied apex A → `192.0.2.1`, worker route →
  `portfolio-sites-router`). **Blocked on the owner setting `alexis.ns.cloudflare.com` /
  `leah.ns.cloudflare.com` at Nominet.** Token: `~/.config/cloudflare/portfoliotoken` (All
  zones — Zone/DNS/Workers Routes Edit, no expiry). Recipe: `RUNBOOK_dns_pointing_a_domain_…md`.
- **Copy defect handed to `copy_quality_two_stage`** —
  `CONTRIB_2026-08-18_the_negative_default_survives_a_POSITIVE_identity_spec…md`. Their own
  08-12 root cause (negativity inherited from `identity.key_differentiators`) does **not** fit
  this case: I read the site's differentiators and they are entirely positive. Unfixed, so any
  new build still gets the negative CTA voice.
- **Mortgage-lender researcher fired** (12:33Z). Still **2 lenders** — but **4 items are in the
  `directory_citation_unverified` HITL queue**. Working that queue is what turns them into
  published lenders.

## 4. Traps learned the hard way (do not re-pay these)

- **A parked domain returns 200 on EVERY path.** Verify DNS by reading the **body**, never the
  status code.
- **`www.` resolves NOWHERE across all 36 zones** (measured on the live working site). Not
  fixed — fixing one zone diverges it from 35. **Owner decision**, options in the runbook.
- **Never verify a fix by grepping the binary for its commit sha** — it carries ONE stamp (the
  commit it was built FROM), so ABSENT is normal on a healthy build. Use `git merge-base
  --is-ancestor <fix> <tag-bump commit>`. All three failed probes: `WRONG_CALLS.md` 08-17.
- **Seeded `banned_claims` fail silently when double-escaped** — valid regex, matches nothing,
  and a count-based verify passes. **Probe, in Go.** `LANDMINES.md` + the 08-18 seed's verify
  block show the working shape.
- **Cost measurements taken while a build is running are low** — the pilot's figure was ~70%
  under because `collected_data` fills in as runs progress. Re-run until two agree.
- **I ruled a market category out of existence** (`sportsreviewinsurance`) **without reading the
  adjacent register entry that named it** (I9 lists `journalistinsurance` — "PI for journalists
  — libel/defamation"). A `[REASONED]` marker cannot flag evidence you never looked for.
  `WRONG_CALLS.md` 08-18.

## 5. Cost baseline (measured; images assumed)

**Text: $3.81/domain today · $4.83 from 2026-09-01** (Sonnet 5 intro rate ends). 73 calls,
663,759 in, 184,596 out, three agreeing runs. **Images are NOT measured** — all from a Google
model (`banana/gemini-3-pro-image-preview`), and no cost column exists anywhere in the platform.
At 30 images/site the total is **$5.01–$11.31/domain**; **fleet of 140 ≈ $700–$1,600**. Above
~$0.08/image, imagery overtakes text as the dominant cost.

## 6. Files of record

`PLAN_2026-08-12_fleet_buildout.md` (phases) · `PLAN_2026-08-18_first_50_build_order_FOR_APPROVAL.md`
(approved) · `DECISION_2026-08-18_two_builder_flows_side_by_side.md` ·
`RUNBOOK_dns_pointing_a_domain_at_the_serving_worker.md` · `REGISTER_positioning.md` (B8/B9/I10
rulings) · `NOTES_portfolio_positioning.md` (evidence, newest at bottom) · `README_where_we_are.md`
(owner's log) · `SUMMARY_2026-08-17_…md` · seeds + missions for the pilot and build #1.
Architecture: `RFC_037` (classifier), `RFC_031` (enrichment splices).
Register: `docs026_concept_register/register/directory-pipeline.md` (DIR-001). Closed:
`bugs_closed/292`.
