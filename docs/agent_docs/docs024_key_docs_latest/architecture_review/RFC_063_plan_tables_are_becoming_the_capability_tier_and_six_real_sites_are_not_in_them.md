# RFC_063 — The plan tables are becoming the capability tier, and six real sites are not in them

**Status: ~~OPEN — owner decision requested~~ DECIDED 2026-09-02 (same evening) — OPTION B,
hand-insert permitted for this closed backfill.** See the OWNER RULING appendix (recorded by
the finetuning lane, `01a3b96ac`) for his verbatim words, the scope of the framework-ruling
exception, and what the ruling does NOT waive: the one-site reconciler-skip proof before any
insert, and the IMG-078 lane's cookly-pairing + seed-from-assets caveats. **Execution, as of
2026-09-02 late night: the IMAGERY SEEDING step is CLAIMED by the `bugfix_114_imagery_wiring`
lane** (their PLAN, `448671d18`; scope: per converted site, `site_plan_imagery` page-scope hero
rows seeded from the `assets` table under the ContentHeroKey convention — ai-agent-orchestration
17, finetuning 14, loancash 9, lampenkap 6, gaswholesalers/cookly zero; ships as a
council-reviewed migration; sequenced strictly AFTER the reconciler-skip proof and after the
composition half creates each `site_plans` row; finetuning.uk stays excluded from IMG-078
acceptance per their 664/649 overlap). **The COMPOSITION half and the reconciler-skip
MEASUREMENT remain UNASSIGNED** — the measurement is the first act, and the 114 lane offers
joint design (their RUNBOOK census queries; the needs_rebuild-17 split in their 412 §11b
addendum). Filed
2026-09-02 by session "bugs_open/443" (lane `bugfix_443_fallback_tier_subjects`), spun out of
`bugs_open/443` §4's closing observation, at the finetuning lane's and the council
architecture seat's prompting (corr `b7c59309`, two advisory objections on exactly this).

## The pattern, stated once

New per-section and per-site capability keeps landing in the `site_plans` table family and
nowhere else. As of 2026-09-02:

| capability | store | reachable without a current `site_plans` row? |
|---|---|---|
| per-section fact scoping (151 cand 1) | `site_plan_sections.assigned_fact_ids` | since today: `pages.section_facts` (PBP-051) |
| per-section subjects (PBP-049) | `site_plan_sections.subject` | since today: `pages.section_subjects` (PBP-051) |
| route-1 hero imagery (IMG-078) | `site_plan_imagery.plan_id → site_plans` | **NO** |
| tier-4 sibling layout synthesis | `site_plan_pages` roles | NO (returns nothing) |
| plan-version rebuild stamping | `pages.built_from_plan_version` | NO (no version to stamp) |

`[MEASURED 2026-09-02]` **6 real sites / 203 deployed pages** have no current plan row:
finetuning.uk (57), ai-agent-orchestration.com (47), gaswholesalers.com (41), loancash.co.uk
(30), cookly.uk (15), lampenkap.com (13). (19 further `pool-*.internal` rows are plan-less with
zero deployed pages — excluded, counting them triples the apparent radius.) All six carry a
current `site_specs.site_plan` ASPECT (the older generation's store), so they are not
unplanned sites — they are sites whose planning history predates the tables.

The bug that surfaced this: 11 of those pages repeat a component type, 11 of 11 served real
near-duplicate sections (≥8 verbatim h2 pairs), because subjects were structurally unreachable
at their serving tier. Fixed for subjects+facts by PBP-051 — but that fix is **the pattern
this RFC asks about**: a per-capability fallback arm, added capability by capability, forever.

## The two options, costed

**A. Keep adding fallback arms (the PBP-051 pattern).** Each new plan-table capability grows a
same-row/same-object sibling on the fallback tiers, with alignment guards and a detector.
+ Cheap per capability (~200 lines + tests this time); no birth-path risk; ships with the
  capability.
− Never fixes table-keyed capability (`site_plan_imagery` cannot be reached this way without
  duplicating the whole table one tier down); the guard surface grows with every arm; the
  architecture seat's objection on record: *"the fix compensates for uncoordinated writers
  rather than resolving the coordination gap"*.

**B. Converge the six sites into the plan tables** — materialise a minimal current
`site_plans` row + `site_plan_pages`/`site_plan_sections` derived from live `pages` (an
identity conversion: the plan says exactly what the pages already say).
+ Fixes the CLASS: subjects, facts, imagery, sibling synthesis and version stamping all become
  reachable; the estate converges on one capability tier instead of two.
− It is a PROGRAMME, not a patch (RFC_034's own ruling for the analogous conversion):
  `reconcile_site_plan` emits `needs_page` rebuild items against plan pages, so a materialised
  plan touches the rebuild path of 203 deployed pages. The reconciler's decision 1 skips
  "deployed at current version, or older version with unchanged composition" — an identity
  conversion should mostly skip, **but this is asserted, not measured**, and pages with
  `build_status='needs_rebuild'` (17 in this cohort) and the 37 empty-`sections` pages have
  non-obvious outcomes (an empty-sections page currently no-ops at `mark_no_ready_sections`;
  on a converged site tier-4 synthesis could give it a borrowed layout instead — a behaviour
  change). **21 files consume current-plan existence as of 2026-09-02** — each is a blast
  radius to measure per site before any conversion dispatches.

## What is asked of the owner

1. Is B the destination (with A as interim, which is what today's state already is), or is A
   the accepted steady state? The 2026-08-17 RFC_034 ruling pattern ("hybrid shape, run
   through the framework, one cohort first") fits B: convert ONE site (suggest cookly.uk, 15
   pages, zero repeat-layouts, smallest blast radius), measure the reconciler's actual
   behaviour against it, then decide the remaining five.
   > **Qualified by the IMG-078 lane's consumer input below (2026-09-02, `b587f116e`):**
   > cookly.uk holds ZERO content-hero assets, so it proves plan creation but nothing about
   > imagery delivery — pair it with one of the four asset-holding sites before concluding
   > anything about IMG-078. And whoever seeds imagery rows must seed from the `assets`
   > table, never page enumeration: `check_unfulfilled_imagery_plan` files GENERATION items
   > (real spend) for any current-plan imagery row whose key lacks an active asset. Their
   > section also honestly DOWN-RANKS imagery as an argument for B — read it before weighing
   > the table above.
2. If B: whether the conversion runs as a new framework action (a `materialise_site_plan`
   work-item type with the identity-conversion property asserted by its own guard) — per the
   2026-08-04 every-site-through-the-framework ruling, a hand-INSERT of plan rows is not on
   the table.

**What this RFC does NOT ask:** permission for PBP-051, which shipped under the estate's
after-the-fact review posture (owner ruling 2026-07-29 §2) with council approval `b7c59309`
and is correct under either answer here — the sibling columns remain valid for any page a
plan does not name.

Consumers told (2026-07-29 §3): `bugs_open/114` lane (IMG-078 is the capability B unblocks
for these sites), apis.uk lane (PBP-049), finetuning lane (largest affected site),
copy_quality_two_stage (subjects precondition their experiments).

Evidence trail: `bugs_open/443` §§4–8, `bugfix_443_fallback_tier_subjects/PLAN_2026-09-02` +
`NOTES`, PBP-049/PBP-051 register entries.

---

## Consumer input from the IMG-078 lane (bugfix_114_imagery_wiring, 2026-09-02 night) — what imagery delivery actually needs from option B, one interaction that could make a naive materialisation SPEND MONEY, and an honest ranking of our own claim

Contributed on request rather than waiting for the ruling. Everything below is from
reading the resolver at HEAD this session (`plan_sections_action.go`, `ensureAssets`),
not from the register.

**1. The minimal materialisation imagery needs is TWO rows, not a reconciled plan.**
Route 1 reads exactly: a `site_plans` row with `is_current = true`, joined to
`site_plan_imagery` rows (`scope='page'`, `scope_ref=<page name>`, `kind='hero'`,
`key=<asset_key>`, any `ordering` — it only orders among siblings), joined to an ACTIVE
`assets` row under that key. It reads **nothing else** from the plan tier — no
`site_plan_sections`, no subjects, no composition state. The 46 active content-hero
assets on 4 of the 6 sites already satisfy the third leg. So for THIS capability,
option B's cost is one plan row per site plus one imagery row per page-with-an-asset —
if the wider blast radius (the 21 current-plan-existence consumers) is the concern,
imagery does not add to it beyond the plan row's existence itself.

**2. ⚠ The interaction that could turn materialisation into generation SPEND:**
`check_unfulfilled_imagery_plan` sweeps the CURRENT plan's imagery rows and files a
`needs_imagery` GENERATION item for every row whose `key` has no active asset. A
materialised plan that carries imagery rows beyond the existing-asset set — or plan
rows on the two sites with zero content-hero assets (gaswholesalers.com, cookly.uk) —
enters that check's population the moment `is_current` flips true. Rows keyed only to
EXISTING active assets are inert to it. Whoever executes option B should seed imagery
rows from the `assets` table, never from page enumeration, and re-read that check
first. (cookly.uk as the first small site is fine for the composition half; for the
imagery half it is the site with nothing to wire, so it proves plan creation, not
imagery delivery — pair it with one of the four asset-holding sites before concluding
anything about IMG-078.)

**3. The honest ranking, stated so the RFC does not over-claim on our behalf: imagery
is option B's WEAKEST argument among the capabilities named.** Content heroes have a
plan-independent fallback — route 2 (Lane B) resolves `ContentHeroKey` straight from
`assets` and delivers today on plan-less sites; 443's subjects had no fallback at all
until `dbb218a41`. What route 1 adds over route 2 is priority and durability (it wins
the merge on every render), which matters most for the GAP-4 cohort — whose mechanism
is still undiagnosed (the canary protocol in `bugs_open/412` §11a runs first). If the
owner rules A, imagery loses little today; if B, seed per §2 above and IMG-078's
round 2 gains its plan-less arm for free.

— bugfix_114_imagery_wiring (session `bugs_open/114`)

---

## OWNER RULING 2026-09-02 (evening, via the finetuning lane) — DECIDED: option B, and the hand-insert is permitted for this one-off

His words, verbatim from chat:

> "yes migrate the sites to handle more complexity in the tables. I think a handwritten insert
> is ok for this one because no new sites will use the other single page option."

So, on the two questions asked:

1. **B is the destination.** Migrate the six plan-less sites into the plan tables. A (the
   PBP-051 fallback-arm pattern) is the interim state we are already in, not the steady state.
2. **Question 2 is answered in the negative — a hand-written insert IS acceptable here**, on his
   stated ground: this is a closed backfill of six known sites, and no new site will ever take
   the old single-aspect path, so there is no ongoing door to guard with a framework action.
   This is a **scoped exception to the 2026-08-04 every-site-through-the-framework ruling for
   this conversion only**, granted by the same authority that made that ruling — not a precedent
   for hand-inserting anything else.

**What the ruling does NOT waive (engineering duty, not decision):** the identity-conversion
skip property is still ASSERTED, NOT MEASURED (§ options, B's minus column). Before any insert
against a live site, prove on ONE site that materialising the plan does not put its deployed
pages onto the rebuild path — the reconciler's decision-1 skip must be demonstrated against
real rows, including the `needs_rebuild` (17) and empty-`sections` (37) cohorts whose outcomes
the RFC itself calls non-obvious. Staging (one small site first, then the five) remains the
recommended shape; the owner ruled on destination and mechanism, not on batch size. The
IMG-078 lane's caveat stands: pair cookly.uk with an asset-holding site before concluding
anything about imagery delivery, and seed imagery rows from `assets`, never page enumeration.

Recorded by the finetuning lane, which carried the question to him; the `bugs_open/443` session
(this RFC's author) has been told directly.
