# 142 — the undeployed_asset detector cannot see a missing artefact, and mis-fires on present ones

Filed 2026-07-29 by the bugfix_131_og_card lane (session relojistas-5), which hit both halves
live. Workstream docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_131_og_card/`.

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
