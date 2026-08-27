# PLAN — bugs_open/414: a planted acceptance marker served as a compliance claim

**Opened 2026-08-27** by a session resuming the bug the `portfolio_positioning` lane filed the
previous evening and handed on. Owner decisions taken at the start of the work are recorded in §2.

## 1. What we are trying to do

A 2026-08-02 shadow experiment planted a tripwire in lendzy.co.uk's `content_direction`: *"include
the exact phrase: checked against the FCA handbook, rule by rule."* The writer obeyed — that is what
the tripwire tested — and the sentence was **served on a finance site as an unverifiable claim of
regulatory diligence**. Then the estate's own audit fleet read the served copy back and filed a
`content_rewrite` calling it "the site's core differentiator", asking for a *"How we verify our
guides"* methodology section: the improvement machinery was generating work to manufacture evidence
for a claim with nothing behind it.

Two jobs, and the second is the point: get the false claim off a live finance site, and close the
class fleet-wide so no spec can order a page to say what the page gate would refuse.

## 2. Decisions, and their reasons

**D1 — the repair is authored by the FRAMEWORK, not by this session (owner, 2026-08-27).** Two of
the three sentences needed a whole sentence rewritten, not a clause deleted, so the choice was real:
`apply_section_edit` with text I compose, or a `content_rewrite` at `spec.mode='edit_live'` whose
copy the writer produces. Owner chose the framework writer. The `rewrite_guidance` therefore says
what to remove, why it is false, and **what is true according to the site's own recorded brief**
(name the rule beside the figure and link it) — so the replacement is grounded in the brief rather
than in a session's judgement. `section-editor` is the fallback if the writer fails twice.

**D2 — the two claim shapes are split by the refusing set's own bar (owner, 2026-08-27).**
Completeness-of-verification ("Everything on this site is checked") joins `globalBannedClaims` at
blocker: it is false-by-construction for every site we will ever run, which is that family's stated
bar, and the family already carried the pattern. Diligence-performed ("checked against the FCA
handbook, rule by rule") joins `claims_practice.go` at **warning**, attestation-exemptible, because a
compliance-services client could truthfully say it. The deciding evidence for the second is a
sentence: `negationCueRe` has no bare `nothing` cue, so *"Nothing here has been checked against the
FCA handbook, rule by rule"* — a correcting disclosure — reads as un-negated and at blocker would be
**refused**. A layer that refuses the disclosure it exists to encourage is worse than a warning.

**D3 — the spec scan uses the practice family only, and this is measured, not preferred.** The first
design ran the whole engine over spec text. A skeptic pass measured it: the fleet-wide + regulated
set over 522 current spec rows gives **21** hits, effectively all false — 15 are the estate's own
honesty instructions ("Never invent a person, company, scheme…"), and `evidence_base` rows store
each site's `banned_claims` **as data**, quoting the sentences they forbid. A generic spec scan
convicts every site's own immune system, daily. The practice family over the same text: 0 of 532
current, 2 of 2,782 all-history, both true positives.

**D4 — the repair goes BEFORE the detector, against the instinct.** ⚠ With the pattern live and the
phrase still in `content_data`, a `page_rerender` regenerates HTML carrying it, the persistence floor
refuses the save, and the OLD `rendered_html` keeps serving: the item lands `unresolved` and nothing
a visitor sees changes. Shipping the gate first would have converted a useless rerender into a
stranded one. Data and dispatch take effect immediately; the Go is inert until a roll anyway.

**D5 — no guard on `WriteSiteSpecAction`.** It would not have caught this: the plant arrived as a
**manual** row (`source='manual'`, 2026-08-02 18:41) and never passed through that action. The agent
door was the second hop, not the first.

**D6 — no widening of the idiom list to chase paraphrase.** The audit fleet had already restated the
claim as "FCA-rule-level accuracy checked **guide by guide**" — no `against`, no rulebook noun. A
literal family cannot win that race and should not pretend to; the durable control for the
canonisation loop is the `model_opinion` origin door that already shipped (migration 629), and the
residual is the pre-door backlog of items that still have a one-click Retry.

## 3. Corrections to the originating brief, marked as corrections

> **The bug file's "spec source FIXED live … regeneration can no longer re-plant the phrase" was
> FALSE, and it is why the bug was still live.** Only `content_direction` had been stripped;
> `domain-strategist` had already restated the instruction, in prose, in the **current `strategy`**
> aspect on 2026-08-12 — an aspect the writer never reads and `build-site-planner` does. Corrected in
> place in the bug file (§"What is FIXED" carries a REFUTED block) and generalised into
> `LANDMINES.md` and `016b` §9.

> **The population was larger than the census showed.** `page_component_history` holds **14** archived
> rows carrying the phrase, and the guide's `article-body` re-emitted it on **4 separate
> regenerations** between 08-15 and 08-24. The census counted what was live; the history says the
> spec was re-planting it every time the page rebuilt.

## 4. Phasing

| phase | what | state |
|---|---|---|
| P1 | strip the propagated marker from `strategy` under a tail-assert guard | **done** 2026-08-27 08:30Z, row `0326a892`, history intact |
| P2 | reject the inverted audit item `052d01b0` with its reason on the row | **done** 08:33Z |
| P3 | dispatch the copy repair through the framework (`edit_live`) | dispatched 08:36Z, retrying through a fleet-wide `resolve_links` flake |
| P4 | the framework fix: 2 completeness patterns + P6 + the spec detector | **committed** `fc588e445`, council `f4c144ad` submitted |
| P5 | verify at the artefact after the roll; re-run `claimscan` and the spec scan | owed |
| P6 | tell the four lanes whose machinery this touches | owed |

## 5. What would make this wrong

- If the framework writer cannot produce clean copy in two attempts, D1's premise fails and the
  fallback (section-editor with text computed in the same run) is the honest answer — recorded here
  so the switch is a decision, not a drift.
- If the council objects to P6's placement in the practice family (a reviewer may argue the
  false-by-construction bar IS met), the migration path is a separate family with its own
  attestation, exactly as `claims_regulated.go` is separate. Cheap then, unfalsifiable now.
- If `spec_supplies_claim` fires on anything other than a real planted instruction on its first live
  run, D3's measurement was too narrow and the surface — not the severity — is what to change.
