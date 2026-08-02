# PLAN — bugs_open/168: one derivation for a deployed asset's path

**Lane opened:** 2026-08-02. **Bug:** `bugs_open/168_HANDOFF_2026-07-31_deployedwebpath_
cannot_express_an_underscore_purpose_stored_under_its_own_key.md` (filed by the
`bugfix_128_image_url_404` lane, unowned until now).

## What the defect actually is — and how the filed framing had to be corrected

The bug file states the mechanism as: *a purpose containing an underscore, stored on an
asset whose `asset_key` is empty or equal to that purpose, yields a path with an
underscore where the deployed file has a hyphen.*

> **CORRECTED 2026-08-02 — that framing is too broad, and acting on it literally would
> have made things worse.** `deploy_image_asset_action.go:185` branches on the *identical*
> condition the helper does (`if assetKey != "" && assetKey != purpose`). So for a file
> the deployer publishes, the deployed name **also** carries the underscore: helper and
> deployer agree. Fix candidate 2 in the bug file — "apply the `_`→`-` swap
> unconditionally" — would therefore have *introduced* the very drift it claims to fix,
> by making the helper disagree with the actual writer for any future underscore purpose.
> Caught by reading `DownloadOptimizeAndPrepare` and the deployer's Phase 2E branch before
> writing any code.

The real defect is one level up, and it is what makes this framework-scope:

**`DeployedWebPath` is documented and used as "the path this asset is served from" — a
question about the ARTEFACT — but it can only answer "the path `deploy_image_asset` would
derive" — a question about ONE WRITER.** There are two writers:

| writer | how it names the file |
|---|---|
| `deploy_image_asset_action` | purpose fixes directory + extension; `asset_key` (when it differs from purpose) fixes the filename, `_`→`-` via `AssetKeyFilename` |
| `derive_brand_head_assets_action` | fixed literals: `favicon.png`, `og-card.png` — committed directly, recorded by `recordDerivedAsset` |

A caller holding `(asset_key, purpose)` cannot tell which writer published the row. So the
brand-head exception had to be learned **separately at every call site** — 016b §9 case 7,
one call site guarded and the root mechanism left generic. It has six consumers.

## Decisions, and why

1. **Candidate 1, strengthened — and the strengthening is the point.** The bug file's
   candidate 1 ("teach `DeployedWebPath` the brand-head map") fixes the *readers*. It does
   nothing about the deeper problem: the derivation was **implemented twice** — once in the
   deployer, once in the helper — and kept in step by a doc comment saying it "mirrors" the
   other. They did agree. The defect was that *nothing made them*. So the fix is one
   function, `storage.DeployedAssetPath`, that the writer and all six readers resolve
   through.
2. **The brand-head map stays declared, and becomes an INPUT.** `og-card.png` is not
   derivable from `og_card` — it has to be declared somewhere. The change is that it stops
   being a *rival answer* to the same question and becomes what the shared derivation
   consults. `BrandHeadAssetPaths` keeps its name, type and values, so its external
   consumers (`check_undeployed_assets`) are untouched.
3. **The 142 lane's pinning test is inverted, at its own written instruction.** Its failure
   message reads: *"Collapse `BrandHeadAssetPaths` into it and delete this test, or the
   platform has two answers to one question."* It was written as a tripwire that names its
   own remedy, not as a veto. Kept (not deleted) with the assertion flipped, so it now
   guards the reverse regression.
4. **`IsBrandHeadPurpose` is NOT deleted.** It stays correct for a *different* question —
   "which table holds the evidence this is deployed?" — which is what
   `check_undeployed_assets` uses it for, and which is unchanged. Only the "where is it
   served from?" use goes away.
5. **The guard is structural, so the test reads source.** The value-level assertion
   "deployer path == reader path" is now a tautology (one function) and would pass even if
   the function were wrong. What failed before was a structure, so the sensor asserts a
   structure.

## Scope declaration (owner rulings 2026-07-28 / 2026-07-29)

This changes what a shared mechanism **guarantees** for six consumers, so it is
architecture-scope and went to the council gate on its own merits rather than riding inside
a bug patch — the `bugs_closed/124` precedent. Per the 2026-07-29 ruling I do **not** claim
an ordering constraint: HEAD is shared, `make build-*` builds from committed HEAD, and any
other session's roll ships this. Review here is after the fact, by design. Condition (2)
stands and is met: the seam is registered in the concept register in the same commit.

**The six consumers, named and told** (`grep`, 2026-08-02): `plan_sections_action.go` (×5),
`render_site_components_action.go`, `emit_sprite_css_action.go`, `derive_card_asset_action.go`,
`queryresolve/queryresolve.go` (×2), `discovery_checks/check_image_url_404.go`. What changed
about their guarantee: `DeployedWebPath` now returns the path the asset is served from
**whichever writer published it**, so none of them needs `IsBrandHeadPurpose` to ask that
question. Five pass non-brand-head purposes and are behaviour-identical; the sixth already
special-cased brand-head to the same answer, so its branch is removed here.

## What is deliberately NOT closed

`deploy_image_asset`'s `deploy_path` input overrides everything and is **invisible** from
`(asset_key, purpose)`. No Go code sets it and it appears in zero orchestrations in history
(audited by the 128 lane, 2026-07-31), so it is an unused passthrough — but a caller that
starts setting it has left this contract. Stated in the helper's doc comment rather than
implied away.

## Phasing

- [x] Verify the bug is still live and unowned; census re-run against the live DB
- [x] File the diagnosis loop **before** asserting a structural cause (corr `ae9404bd`, REFUTED
      — reported, and its own incomplete census corrected)
- [x] One derivation + brand-head as input; deployer folded in; redundant branch removed
- [x] Guards proven by mutation, not by a passing run — **five**, including one run *after*
      the fix to prove the removed local guard transferred rather than evaporated
- [x] Council gate round 1 (corr `abd9b119`) → **REVISE**, gated by `guardian`
- [x] Verdict read and every objection dispositioned (full record in NOTES)
- [x] The one real code defect fixed: `brandHeadAssetPathsFor` took the map's value whole
      instead of reconstructing it under `DefaultAssetBasePath` (`editquality`'s catch)
- [x] `bugs_open/179` filed — the `deploy_path` escape hatch and the writer's new clobber
      path, *tracked* rather than disclosed in prose (`bug_historian`'s objection)
- [x] `RFC_009` filed — the standing artifact (`architecture`'s objection: "declaring it
      doesn't relocate it")
- [x] Register entry IMG-067; IMG-066's two now-false sentences corrected and its open
      `verify-later` answered; LANDMINES entry updated + synced; committed
- [x] Pod-verified the fix is **NOT** live (3 controls on `v1.0.1228`) — which is *why* the
      bug stays open
- [x] Council gate round 2 (same corr, `RESUBMIT_CORR`) — every objection answered with
      evidence rather than argument
- [x] Round-2 verdict read and acted on — **REVISE, gated HIGH**: the clobber I twice called
      unreachable *was* reachable (11 queued items). Guard shipped rather than filed
- [x] Council round 3 — **APPROVED**, 2 advisory objections, none high. Four of them were
      checkable and were checked, not filed; `reuse_agent`'s caught a convention I had
      asserted without looking for, and `bug_historian`'s caught a silent path my own round-1
      fix had introduced
- [x] **LIVE on `v1.0.1229`**, pod-verified on BOTH replicas with a NEGATIVE control; all 24
      brand-head artefacts across 12 sites serve 200
- [x] `bugs_open/168` → `bugs_closed/168`, register + landmine updated to match
- [ ] **Owner call, not code:** whether the 11 stale queued items should be re-pointed at
      `mode=brand_head` (the repair they actually want) rather than merely refused
- [ ] **`bugs_open/179` finding A** — the `deploy_path` escape hatch, measured empty across
      three populations, still unguarded. Two seats pressed it at medium
- [ ] **`RFC_009`'s wider question** — the platform reconstructs an artefact's identity from
      metadata instead of reading what the writer recorded (shared root of 152 / 155 / 179).
      Wants designing *with* those lanes
