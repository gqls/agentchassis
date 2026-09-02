# PLAN — 2026-09-02 — `bugs_open/436` — the CTA eligibility lever (candidate 1 + candidate 4)

## What we are trying to do

Build the platform lever the owner approved as **decision 3** on 2026-08-25 (recorded in
`bugs_closed/391` §OWNER DECISIONS and `bugfix_389_cta_relevance/PLAN_2026-08-25` §decisions):
an explicit **`eligible_as_cta_target` opt-out** on `pages`, read at the **ranking** and by the
**label universe**, **paired with a detector** for the anomalous-`nav_order` (fossil) shape. This
closes the *cause* that `bugs_closed/391` cleaned the damage of: `chooseCTATargets` picks every
site's primary button by `COALESCE(nav_order,100)` then `name`, with no semantic input, and the
wrong pick label-locks itself in.

## Why no fresh 090 diagnosis run (declared substitute, per the 2026-07-31 owner ruling)

The root-cause claim in `bugs_open/436` was **adversarially reviewed and CONFIRMED on 2026-08-25**
(391 §THE FEEDBACK LOOP — the review found the label-lock loop the original diagnosis sat inside),
and **re-verified first-hand at HEAD today**: every cited function re-read
(`chooseCTATargets` :651, `stampCTADestinationGuidance` :371, `setCTAField` :399, the header
fallback `render_site_components_action.go:181-196`, `LoadCTALabelUniverse`), and the consumer
enumeration re-run (see below). The live queue independently corroborates the mechanism
`[MEASURED 2026-09-02]`: two `needs_content_planning` deferred verdicts describe a 63-tool site
whose hero "singles out two arbitrary tools … as default fallback CTAs on unrelated guide and
news pages", plus ~10 `page_rerender` items unresolved after 2 attempts on misdirected CTAs.

## Consumer enumeration (RFC_022 / owner ruling 2026-07-29 §3) — `[MEASURED 2026-09-02]`

By `grep -rn 'chooseCTATargets|loadContentHubs|loadInteractivePages|LoadCTALabelUniverse'`
over `platform/ internal/ pkg/ cmd/`, non-test:

**`chooseCTATargets` + the two positional loaders — 3 callers:**
1. `resolve_internal_links_action.go:162` — build-time CTA resolution (persisted).
2. `rerender_page_sections_action.go:691` — repair-time recompute `applyCTARecompute` (persisted).
3. `render_site_components_action.go:190` — **site header CTA fallback (never persisted;
   `site_components` holds 0 `cta_url` keys)** — the reason the flag binds at the ranking.

**`LoadCTALabelUniverse` — 4 callers:**
1. `resolve_internal_links_action.go:173` (writer, build).
2. `rerender_page_sections_action.go:705` (writer, repair).
3. `discovery_checks/check_misdirected_cta.go:402` (detector).
4. `cta_label_audit.go:310` (audit).

**`BestLabelMatchForPage` — 3 non-test callers:** the two writers plus
`JudgeCTALabel` (`cta_label_agreement.go:169`), which itself serves `check_misdirected_cta` and
`check_cta_nonpage`. The judge is `bugs_open/399`'s seam — they are told via CONTRIB (see
Communications).

No new action-input key is added, so the RFC_022 optional-key budget (WFA-013) is untouched:
the only config key on this path, `stamp_cta_destination_guidance`, already exists.

## Design (the decisions, with reasons)

1. **Column, not JSONB key:** `pages.eligible_as_cta_target boolean NOT NULL DEFAULT true`
   (migration `714`). Default true = today's behaviour on every existing row, byte for byte. The
   new authority (excluding a page) defaults OFF — the 2026-08-02 §2 shape. `rebuild_policy` is
   the precedent for a policy column on `pages`.
2. **Positional binding at the RANKING (constraint 1).** The loaders SELECT the column and carry
   it on the candidate struct; `rank()` drops ineligible candidates. All three callers inherit
   through `chooseCTATargets` — the header fallback deliberately included and separately tested,
   because its output is never persisted and no `content_data` diff can show it.
3. **Label binding by REFUSAL, not removal (constraint 2).** `CTALabelUniverseSQL` selects the
   column onto `LabelMatchCandidate`; `BestLabelMatchForPage` returns `!ok` when the best match is
   ineligible. **The page stays in the pool** — removing it would let a weak-token runner-up win,
   the exact failure measured for self-links (cta_label_universe.go:144-155: 10 of 35 wrote
   somewhere else, "most of those were wrong"). Refusal mirrors the proven self-link design:
   no opinion, keeps decide, positional pick (which now also refuses the page) decides last.
4. **The judge gains a reason, not a second question.** `JudgeCTALabel` reports
   `SilenceNamesIneligiblePage` for copy naming an opted-out page — the Silence seam exists for
   exactly this (cta_label_agreement.go:110-115). Detectors keep a first-class signal instead of
   a fold into `names_nothing`; no finding suggests a repair the writers now refuse to perform
   (the 308 defect shape: 188 findings naming a destination no writer could write).
5. **Shared supply + ranking move to `datahelpers/cta_positional.go`** (LNK-036's own argument,
   one seam over): the detector must mirror the ranking exactly — "mirror the code exactly or the
   simulation proves nothing" (391's runbook) — and `discovery_checks` cannot import `actions`.
   SQL constants + candidate struct + `RankCTAPositionalCandidates` are single-sourced; the
   actions-side loaders/`chooseCTATargets` become thin wrappers with unchanged signatures, so all
   three callers compile untouched and behaviour is byte-identical apart from the new filter.
6. **Candidate 4 = `cta_rank_anomaly` discovery check.** Fires when the site-level rank-1 CTA
   target (`pageName=""`, the header form) is an interactive page holding a **unique minimum**
   `nav_order` that is **below the default (100)** and **≥50 below the runner-up**, among ≥3
   eligible interactive candidates. Fires on 391's fossil (1 vs 100), silent on all-default sites
   (webdesign's arbitrary-but-not-anomalous shape is candidate 3's business, deliberately out of
   scope), silent on a curated ladder (10/20/30), silent once the page is demoted or opted out —
   correct silencing: the condition is gone, not blinded. `needs_human_review`, one item per site,
   deduped; positively `Resolved` (AllOfType) when a healthy rank-1 is observed.
7. **Check enablement is a separate `_HOLD` migration** (`715_…_HOLD.sql`): the checks array in
   `completeness-discovery-agent` config is live the moment it applies, and the Go check only
   exists after an image roll — image first, then seeds. `_HOLD` is the sanctioned ordering
   mechanism (a banner cannot hold a file; the runner's guard checks drift, not order).

## What this deliberately does NOT do

- Does not touch the keep branches (248/308 semantics intact — authored and minted-utility
  destinations survive exactly as today).
- Does not change nav, listings, or linkability — this is CTA-*choice* policy only.
- Does not set the flag on any existing page. Data decisions stay with the owner.
- Does not unlock label-locked fields on the RECOMPUTE path: `applyCTARecompute` KEEP #2 holds
  any valid stored page, by design. A full rebuild does replace an opted-out destination. Stated
  honestly: this lever prevents the class on NEW sites and new builds; existing locked fields
  need 391's repair recipe (rewrite + relink) or retirement.

## Phasing

- **Phase 1 (this session):** implement + tests; register the seam in the concept register in the
  SAME commit (2026-07-28 condition (2)); council submission before/alongside the commit
  (architecture-scope: the ranking's guarantee changes — 2026-07-29 §1); commit narrowly by
  pathspec with `Council-Submitted:`; `verify-head-builds.sh`.
- **Phase 2 (roll-bound):** after the next chassis roll, apply `715_HOLD` by hand, watch the
  check's first fleet pass, verify at a canary site both directions (induce, don't wait).
- **Phase 3 (owner):** whether to opt out any live page today (e.g. the demoted password-entropy
  trio are already harmless at nav_order 900) — raised, not assumed.

## Communications

- `bugs_open/399` (label agreement): their judge gains a silence reason → CONTRIB into
  `bugfix_399_cta_label_agreement/`. No live session (checked ListAgents 2026-09-02 19:09).
- `cta_target_content_pass` (decision-5 commission, re-scope pending): the lever landing changes
  their population arithmetic → CONTRIB into their dir.
- `bugs_open/384` (stale listings): **not affected** — listing arrays don't pass through CTA
  resolution; no message sent, reason recorded here.
- `bugs_open/248`: closed-adjacent; keeps untouched — noted in the register entry, no thread.
