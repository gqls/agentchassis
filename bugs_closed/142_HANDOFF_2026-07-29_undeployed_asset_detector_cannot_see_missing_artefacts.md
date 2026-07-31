# 142 — the undeployed_asset detector cannot see a missing artefact, and mis-fires on present ones

Filed 2026-07-29 by the bugfix_131_og_card lane (session relojistas-5), which hit both halves
live. Workstream docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_131_og_card/`.

> # CLOSED 2026-07-31 — LIVE on v1.0.1219 and PROVEN BY A REAL SWEEP, not by a pod-grep
>
> **Both mechanisms verified in production, by the platform running the check itself.**
> `design-discovery-agent` — the only agent that runs this check — executed on the
> shipped binary at 19:17:48Z (**vonc.com**) and 19:18:04Z (**idea.uk**). Pods started
> 19:09Z; both runs `COMPLETED`, `created_by='generic'` (the chassis worker), not
> fired by this session.
>
> | | result |
> |---|---|
> | **the blind half — absence is now visible** | **idea.uk raised 2 × `needs_brand_head_assets`**, `item_key` `…:favicon` / `…:og_card`, carrying this fix's own summary wording verbatim. idea.uk is one of the two sites serving 404, holds **0** active brand-head rows, and had produced **zero** findings in the detector's entire history. |
> | **the mis-firing half — false positives are gone** | **vonc.com raised 0 `undeployed_asset` items of any purpose.** It holds 2 active brand-head rows (favicon + og_card, both serving 200), so under the old predicate it would have produced **2 permanent false positives** — no `page_component` can ever reference a head-injected asset. Fleet-wide since the roll: `item_type='undeployed_asset' AND spec->>'purpose' IN ('favicon','og_card')` → **0**. |
>
> > **CORRECTED within the hour, before this file was committed — the first version of
> > this block said "16 would have fired, 0 did" across "9 swept sites", and the
> > denominator was wrong.** Nine sites received work items in that window, but only
> > **two** ran `design-discovery-agent`; the other seven were served by `nav-updater`,
> > `internal-link-resolver`, `page-build-handler` and others, which never run this
> > check. I had counted "sites with any item in the window" as "sites that ran the
> > check". The true counterfactual is **2 avoided, not 16** — a real proof, on n=1
> > site, and it must not be quoted as a fleet figure. *What caught it:* checking which
> > agent actually owned the orchestrations before trusting a count I had already
> > written down. This is the wrong-denominator trap this repo has logged repeatedly.
> >
> > **Also corrected: I briefly inferred that `design-discovery-agent` must now have a
> > cadence.** It does not follow. **Every** `scheduled_tasks` row targeting any
> > discovery or improvement agent is `enabled=false`, and orchestration naming does
> > **not** discriminate scheduled from hand-fired (`build-pipeline-trigger`, a genuinely
> > scheduled task, uses the identical `<agent>-orchestrate-MMDD-HHMM` form). So the
> > trigger for these two runs is **`[UNVERIFIED]`** — most likely another session
> > firing by hand, which is exactly what the fleet's standing note says happens. The
> > claim in `bugs_open/083`/`093` that nothing schedules the checker layer is **not**
> > contradicted by this, and none of it affects the verification: what matters is that
> > the check RAN on the shipped binary and produced the right rows. |
>
> Pod-grep on **both** replicas of `v1.0.1219`, with two positive controls in the same
> exec: `brand_head_provenance_url_unexpected` 1/1, `No asset record publishes` 1/1,
> `best-effort provenance row was lost` 1/1, `needs_brand_head_assets:` 1/1; controls
> `generated but not deployed to site` 1/1 and `assets/images/og-card.png` 3/3.
>
> Commits `3b812161b`, `d671fb2b2`, `6f5e85886` (code), `c617c9eb6` (docs).
> Council **APPROVED at round 1**, correlation `35d88a60-ec1c-4cd3-b69c-f2813c3e837f`
> — 7 advisory objections, none high; the two actionable ones are answered in
> `6f5e85886`. Workstream:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_142_undeployed_asset_detector/`.
>
> **Two branches are NOT exercised live, stated rather than glossed:** the
> provenance-observation branch (gamesdesign.co.uk, robot-hands.com) and the second
> GAP site (webdesign.co.uk) were not in this sweep's 9 sites. Both are covered by
> mutation-verified unit tests, and the observation branch by construction files
> nothing, so there is no row to look for. If you want them live, they need a sweep
> that includes those sites.
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
