# 142 — the undeployed_asset detector cannot see a missing artefact, and mis-fires on present ones

Filed 2026-07-29 by the bugfix_131_og_card lane (session relojistas-5), which hit both halves
live. Workstream docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_131_og_card/`.

> **STATUS 2026-07-31 — FIXED IN CODE, committed, council submitted. STILL OPEN until it is
> live and pod-verified.** Commits `3b812161b` + `d671fb2b2`. Council correlation
> `35d88a60-ec1c-4cd3-b69c-f2813c3e837f` (`Council-Submitted:` trailer — no verdict read yet,
> so no `Council-Reviewed:` claim). Workstream:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_142_undeployed_asset_detector/`.
>
> **Both defects reproduced against live production before touching anything**, and this
> file's own figures had moved, so re-measure before quoting anything below:
>
> | | filed 07-29 | measured 07-31 |
> |---|---|---|
> | detector firings | "five times, all robot-hands.com" | 7 sites; 73 complete / 35 detected / 59 unresolved |
> | what the shipping predicate would raise NOW | not stated | **96 findings across 14 of 14 sites** |
> | sites blind to absence | "the 11 broken sites" | **2** — idea.uk + webdesign.co.uk (the only two serving 404 today) |
>
> ⚠ **The `complete` rows do not mean the detector works.** The recent ones carry
> `created_by='operator'` and `'session-2026-07-31-robot-hands-carousel'` — hand-filed by
> humans. Every row the detector itself created is `unresolved` or parked at `detected`.
>
> **Fix candidates 1 and 3 are BUILT** (population that includes absence; brand-head purposes
> excluded from the rendered_html predicate). **Candidate 2 is built in a stronger form than
> proposed** — see the correction below.
>
> **CORRECTION to fix candidate 2, and it changed the design.** This file proposes that for
> favicon/og_card, "deployed" should be *"the site chrome/head containing the reference"*.
> **That does not work: the head reference is not evidence.** `injectBrandHeadTags` emits the
> `<link>`/`<meta>` **unconditionally**, so 13 of 14 heads advertise a card whether or not one
> was ever generated — idea.uk's head advertises a live 404. The evidence used instead is an
> active `assets` row **recording the published path**, because `recordDerivedAsset` runs only
> after `sendGitCommitRequest` succeeds; verified on the wire across all 15 sites with no
> exception either way. The head reference is still read, but only to say *"and the site head
> already points at it"* in the finding.
>
> **A THIRD state exists that this file does not mention.** gamesdesign.co.uk and
> robot-hands.com each hold exactly one active `favicon` and one active `og_card` row whose
> `url` is the unresolved template literal `/assets/images/input-data.asset-key.jpg` — and both
> sites serve their real files **200**. Such a row is evidence of neither deployment nor
> absence, so it is recorded in `Findings` and raises **nothing**. Filing "never generated"
> against a site serving 200 would be a false claim, which is the defect this bug is about.
> (`bugs_open/152` owns the URL-rewrite defect that produces those rows.)
>
> **The expected-noise paragraph below is now WRONG, and in the good direction.** It predicts
> one standing false positive per backfilled leopardess lock row "by defect 2". Defect 2 is
> gone, so those produce nothing — simulated fleet-wide: leopardess raises **0** brand-head
> findings.
>
> **Net effect, simulated against live data before shipping:** 96 → 72 page-asset items
> + **4** real brand-head findings (idea.uk, webdesign.co.uk — 2 each) + 4 observations.
> The 24 removed are exactly 12 sites × 2 brand-head purposes. **Zero** false positives remain.
>
> **What is still owed:** the roll, then the pod-grep in the workstream RUNBOOK §R9, then read
> the council verdict. ⚠ **A live firing is NOT available to verify with** — this check is
> reachable only via `design-discovery-agent` ← `improvement-sweep`, disabled since 2026-05-02
> (`bugs_open/083`). Verify by pod-grep with a positive control, not by waiting for a finding.

## Symptom

While 11 of 14 sites served a 404 for the `og:image` card their pages advertised
(`bugs_open/131`, og-card slug), the `undeployed_asset` detector fired **five times — every
one for robot-hands.com, the one site whose card worked**. It never once fired for any of the
11 sites actually missing the artefact.

## Root cause — two independent defects in one check

`platform/orchestration/actions/discovery_checks/check_undeployed_assets.go` (`findUndeployedAssets`):

1. **The denominator is the `assets` table.** The query starts `FROM assets a WHERE a.site_id
   = $1 AND NOT EXISTS (…)`. A site whose card was never generated has **no row**, so it is
   structurally invisible. A detector whose denominator is the artefact table cannot see a
   missing artefact — the 11 broken sites could never fire it. (Same family as the fleet's
   "check with no failing branch", one layer down: here the check has no failing *population*.)

2. **The numerator's deploy evidence is wrong for head-injected assets.** "Deployed" is
   defined as a deployed `page_components.rendered_html` containing
   `/assets/images/<purpose>.…` — but `favicon.png` / `og-card.png` are referenced from the
   HEAD (injected by `injectBrandHeadTags` in render_site_components), which is chrome, not a
   page component. So a *present, serving* og_card row can never look deployed, and fires
   forever. That is exactly robot-hands' five findings: the working site read as broken.

Compounding: the emitted work items carry `status='detected'`, which the build trigger never
dispatches (`bugs_open/083` covers that class), so even the false positives only sat in the
queue.

## Evidence

- Five historical `undeployed_asset` firings, all robot-hands (`detected` 07-24;
  `unresolved` ×3 07-18, ×1 07-19) — NOTES entry (1) of the og-card lane, measured 07-28.
- 11 sites reproduced serving 404 og-cards on the wire 07-28 21:2xZ, zero detector findings
  for any of them, ever.
- The check's query, read 2026-07-29 (`check_undeployed_assets.go:71-83`): `FROM assets` +
  `NOT EXISTS (… pc.rendered_html LIKE '%/assets/images/' || purpose || '.%' …)`.

**Expected new noise, deliberate:** on 2026-07-29 the og-card lane backfilled **locked**
`og_card` + `favicon` rows for leopardessconsulting.co.uk (protecting its hand-made artefacts
from re-derivation — see the lane's NOTES (4)/(5)). By defect 2 those rows will produce one
standing `undeployed_asset` finding each (item_key-deduped). They are false positives; do not
"fix" them by deleting the rows — the rows are load-bearing locks.

## Fix candidates, ranked by what closes the door

1. **Give the check a population that includes absence**: for brand-head artefacts, iterate
   over *sites with deployed pages* (or over `ImagePurposes` keys expected per site), not over
   `assets` rows — emit `needs_brand_head_assets` (handler `asset-deployer`, mode
   `brand_head`) when the site has an active logo and no serving card. The generator exists
   and is one work item away (the whole of 131 was "nobody ever queued it").
2. **Fix the deploy evidence for head assets**: for `favicon`/`og_card`, "deployed" should be
   the site chrome/head containing the reference (or an HTTP check of the deployed URL), not
   `page_components.rendered_html`.
3. At minimum, exclude head-injected purposes from the current rendered_html predicate so the
   perpetual false positives stop burying real findings.

## How to verify

Before: detector fires only on rows invisible from rendered_html (robot-hands, leopardess
post-backfill), never on a site with no card. After fix 1: temporarily archive a test site's
og_card row + card file → a `needs_brand_head_assets` item appears; leopardess (locked rows,
card serving) emits nothing.
