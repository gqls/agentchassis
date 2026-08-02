# RFC 009 — one derivation for a deployed asset's path, and what its guarantee now is

**Filed:** 2026-08-02 by the `bugfix_168_deployed_asset_path` lane.
**Status:** open. **The change it describes is COMMITTED (`4035455ae`) and NOT LIVE** — Go,
so inert until the next chassis image is built and rolled. Written *because* the council's
architecture seat asked for it by name, at medium severity, twice in one round
(`abd9b119`), and because by this repo's own 2026-07-29 ruling it qualifies.

## The ask, in the architecture seat's own words

> `DeployedAssetPath` is introduced as the canonical contract for a writer plus 6 readers
> spanning 2 top-level packages, redefining what the mechanism guarantees (brand-head is now
> folded in). This is the exported-symbol trigger from the review-track test; it should land
> with a companion RFC entry recording blast radius/rollback as a standing artifact, not only
> in this plan's prose.

And, on the writer edit:

> The writer (`deploy_image_asset_action.go`) stops implementing its own path logic and calls
> the shared derivation — a coordinated writer+reader contract change. Same trigger.

The seat also recorded the part that matters most for precedent: *"The author's own framing
… is the correct self-assessment — **but per the standing rule, declaring it doesn't
relocate it.**"* That sentence is the reason this file exists. The lane self-declared
architecture scope, brought the change to the gate deliberately, and still owed a citable
artifact. **Self-declaration is not the artifact.**

## Does it actually meet the 2026-07-29 RFC trigger? Yes, and here is the test applied

The owner's ruling narrowed the trigger: an addition to a shared vocabulary needs an RFC
**only when it changes what the shared mechanism GUARANTEES**, not merely because the
vocabulary is shared. Additive-and-inert goes through the normal council gate.

This is not additive-and-inert. The guarantee moved:

| | before | after |
|---|---|---|
| what `DeployedWebPath(assetKey, purpose)` answers | the path **`deploy_image_asset` would derive** | the path the asset **is served from, whichever writer published it** |
| who must know the brand-head exception | **every caller**, separately | the derivation, once |
| how writer and readers are kept in step | a doc comment saying one "mirrors" the other | **one function** |

The first row is the trigger. A caller that asked the old question and got the old answer was
correct; the same call now answers a *different, larger* question. Nothing breaks — the
answers coincide for every input in the fleet except the two brand-head purposes, where the
new answer is the right one — but that is a fact about today's data, not about the contract.

## Blast radius, measured

**Six readers**, `grep` 2026-08-02: `plan_sections_action.go` (×5),
`render_site_components_action.go`, `emit_sprite_css_action.go`, `derive_card_asset_action.go`,
`queryresolve/queryresolve.go` (×2), `discovery_checks/check_image_url_404.go`.
**One writer**: `deploy_image_asset_action.go`. Two top-level packages
(`platform/storage`, `platform/orchestration`) — which is why the seat's *other* trigger
(3+ packages) did not fire.

Three of the six pass a **literal** purpose (`"card"` ×2, `"sprite_sheet"`). The other three
pass a variable scanned from a query, and all three resolve it through `site_plan_imagery`:

```sql
SELECT a.purpose, count(*), count(*) FILTER (WHERE a.asset_key = a.purpose) AS takes_skip_branch
  FROM site_plan_imagery spi
  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
  JOIN assets a ON a.site_id = sp.site_id AND a.asset_key = spi.key AND a.status='active'
 GROUP BY 1;
```
```
 hero | 82 | 0     icon | 77 | 0     illustration | 3 | 0     sprite_sheet | 1 | 0
```

**No brand-head purpose is reachable there, and not one of the 163 reachable rows takes the
branch that changed** — every one carries an `asset_key` distinct from its `purpose`. The
denominator was run: the same join unfiltered returns those 163 rows across 4 purposes, so
the zero is a real zero and not an empty-join artefact.

So: **behaviour-identical at all six readers.** The single behaviour change is at the
**writer**, for brand-head purposes only, which nothing routes to it (`check_undeployed_assets`
excludes those purposes from its generic half and files `needs_brand_head_assets` instead).
That change is tracked as `bugs_open/179`, because it moved the failure mode from "publishes
an orphan nobody references" to "overwrites the deriver's artefact".

## Rollback

Cheap and total, in the sense that matters: **`DeployedWebPath` kept its exact signature**,
so nothing downstream was rewritten and there is no call-site migration to unwind. Reverting
means restoring the deployer's inline branch and the brand-head branch in
`check_image_url_404` — two hunks — and the two tests that pin the new contract. No schema,
no migration, no config, no seed, no agent definition. Nothing persisted takes the new shape:
the derivation is computed at read time and at commit time, never stored, which is *itself*
the property that makes this cheap to undo and is worth noticing as a design point.

The one asymmetry: any brand-head file published through `deploy_image_asset` while the new
code is live would have been written to `og-card.png`, and a revert would not move it back.
Zero such publishes are possible today (above), so the asymmetry is theoretical — but it is
the direction in which this is *not* a pure code change.

## The general question this raises, which is the reason to file a paper and not a note

**Where should a "which writer produced this artefact?" fact live?**

This is the third open bug with the same root — `bugs_open/152`, `bugs_open/155`, and now
`179` — and the register entry for the map (IMG-066) names it too. In every case the platform
**reconstructs an artefact's identity from its metadata** instead of **reading what the writer
recorded**. The `assets` table already has `storage_path` and `filename` columns; they are
populated on **78 and 76 of 267** active rows respectively, so they are neither reliable nor
unused. `assets.url` is worse than unreliable — it is *polymorphic across writers*: an
expiring presigned S3 URL for generated rows, a site-relative web path for the 24 brand-head
rows written by `recordDerivedAsset`.

168 makes the reconstruction **correct** for both writers we have. It does not make it
**unnecessary**, and a third writer would reopen the whole class. The candidate worth
designing — deliberately *not* attempted inside a bug fix — is:

> the writer records the path it committed; readers read it, and derive only as a fallback
> for rows written before the recording existed.

That subsumes 152, 155, 179's finding A, and this RFC's residual. It needs a decision about
backfill (189 of 267 rows have no `storage_path`), about which column is canonical, and about
what a reader does when the recorded and derived paths disagree — which is a question about
trust, not about code. **It should be designed with those lanes, not by this one**, and that
is the recommendation this RFC makes rather than smuggling a schema opinion into a path
helper.

## What was NOT deferred, so this file is not a promise instead of work

- The derivation is one function; the writer resolves through it. Committed.
- Every guard proven by **mutation**, not by a passing run — four of them, including one run
  *after* the fix to prove that removing `check_image_url_404`'s local brand-head branch did
  not evaporate the protection it provided (mutating the helper reproduces
  `image_url_404:og-card.png`, the fleet-wide false positive, on demand).
- The residual escape hatch (`deploy_path`) and the writer's new clobber path are **filed as
  `bugs_open/179`**, not left in prose — the `bug_historian` seat's objection, discharged.
- The `LANDMINES.md` entry for this symbol is updated with a HEAD-vs-running-fleet banner,
  because the fix is committed and not live and the old trap is still true of the binary
  actually running.

## Related

`bugs_open/168` (the case), `bugs_open/179` (the residual), `bugs_closed/142` (the map and
its tripwire, inverted here at its own written instruction), `bugs_closed/128` (the near-miss
that filed 168), `bugs_open/152` + `bugs_open/155` (the same root, the wider question above),
IMG-067 + IMG-066 (concept register), council `abd9b119` round 1,
`RFC_007` (a structurally identical case: a shared vocabulary re-typed at each consumer
because the import direction forbids sharing — worth reading alongside, since this RFC is the
outcome that one is asking for).
