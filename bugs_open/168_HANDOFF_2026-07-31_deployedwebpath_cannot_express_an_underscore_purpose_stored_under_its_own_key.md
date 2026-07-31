# 168 — `storage.DeployedWebPath` cannot express an underscore purpose stored under its own key, and there is no guard stopping one being created

**Filed:** 2026-07-31 by the `bugfix_128_image_url_404` lane, discharging an advisory
objection from the council's `bug_historian` seat (`99dca96a`, round 2, APPROVED): *"the
residual DeployedWebPath drift … is measured to zero today and pinned by a test, but the
durable fix is explicitly deferred as 'its own item' rather than filed. This is exactly
the shape of case 7 (one call site guarded, root mechanism …)."* The seat was right that
saying "its own item" and not creating one is how a deferral becomes a disappearance.

**Severity:** Low **today**, and that is the whole point of filing it — the defect is
latent, the risk set is empty, and nothing is broken. It becomes Medium the moment
somebody adds one asset purpose.
**Class:** a shared helper with an inexpressible case, guarded at each call site
individually rather than at the helper.
**Status:** OPEN, unowned. **Do not "fix" this in a bug patch** — see § Scope.

---

## The mechanism, in four lines of the helper

`platform/storage/url_helpers.go`:

```go
func DeployedWebPath(assetKey, purpose string) string {
	_, _, _, ext := GetImageConfig(purpose)
	base := BuildAssetPaths(purpose, ext)
	if assetKey == "" || assetKey == purpose {
		return base.RelativeURL          // <- purpose + ext, VERBATIM
	}
	// ... else AssetKeyFilename(assetKey, ext), which swaps `_` -> `-`
}
```

The `_`→`-` swap lives in `AssetKeyFilename` and is reached **only** on the else branch.
So a purpose that **contains an underscore**, stored on an asset whose `asset_key` is
empty or equal to that purpose, yields a path with an underscore where the deployed file
has a hyphen.

## Why nothing is broken today — measured, not assumed (2026-07-31)

Over all **267** active asset rows fleet-wide, the rows that take the skip branch are:

| purpose | asset_key | rows | underscore? |
|---|---|---|---|
| `favicon` | `favicon` | 12 | no |
| `og_card` | `og_card` | 12 | **yes** |
| `hero` | `hero` | 5 | no |
| `logo` | `logo` | 4 | no |

`og_card` is the one live instance, and it is **already handled** — by
`storage.BrandHeadAssetPaths` / `IsBrandHeadPurpose`, added by the `bugs_open/142` lane
and pinned by `TestDeployedWebPathCannotExpressBrandHeadPaths`. Every other underscore
purpose is stored with a distinct key (`content_hero` ×31, `sprite_sheet` ×1) and
therefore takes the swap correctly.

**So the risk set outside brand-head is EMPTY.** What that measurement cannot do is prove
the negative for assets that do not exist yet, and nothing in the schema or the code
prevents the next one: no constraint forbids an underscore purpose, and no guard fires
when an asset is stored with `asset_key = purpose`.

## Why it matters more than "one wrong string"

`DeployedWebPath` is consumed by **six** call sites now — `plan_sections_action` (×5),
`render_site_components_action`, `emit_sprite_css_action`, `derive_card_asset_action`,
`queryresolve`, and since 2026-07-31 `check_image_url_404`. Each has had to learn the
brand-head exception **separately**, or not learn it at all. The 128 lane nearly shipped a
check reporting a 404 for the og card and favicon of every site in the fleet by trusting
the helper's own doc comment, which calls itself the single source of truth. The pattern —
one call site guarded, the root mechanism left generic — is 016b §9's case 7.

## Fix candidates, ordered by what closes the door

1. **Teach `DeployedWebPath` the brand-head map** — return `BrandHeadAssetPaths[purpose]`
   when `IsBrandHeadPurpose(purpose)`. Makes the helper correct for every input, so no
   caller can get it wrong. ⚠ It **inverts** the existing
   `TestDeployedWebPathCannotExpressBrandHeadPaths`, which pins the current behaviour
   deliberately *"so the duplication cannot outlive its reason"* — that test is the 142
   lane's, and its inversion is a conversation with them, not a silent edit.
2. **Make the inexpressible case unrepresentable**: apply the `_`→`-` swap
   unconditionally in `DeployedWebPath` rather than only on the else branch, so purpose
   and asset_key are normalised identically. Smaller, but it changes the path for any
   *future* underscore purpose without telling its owner, which is the same class of
   surprise in the other direction.
3. **Refuse the input**: a check (or a DB constraint) forbidding an asset stored with
   `asset_key = purpose` where the purpose contains `_`. Does not fix the helper, but
   turns a silent wrong path into a loud refusal at the only moment anyone is watching.
4. Leave it, and require each caller to branch on `IsBrandHeadPurpose`. **This is the
   status quo and it has already produced one near-miss.**

## Scope — read before starting

This is a **shared helper with six consumers**, so a change here is architecture-scope by
the owner ruling of 2026-07-28/29: it alters what a shared mechanism GUARANTEES for every
one of them. `bugs_closed/124` was vetoed by the guardian seat for exactly the shape of
change this would be if it arrived inside a bug patch. Route it through the council gate
on its own merits, name the six consumers in the submission, and **tell them** — the
useful message is what changed about their guarantee, not a list of new keys.

## Verify a fix

- All six call sites still resolve the paths they resolve today. The cheapest whole-fleet
  assertion is the 128 lane's: compute `DeployedWebPath` over every active asset row and
  intersect with the distinct `/assets/images/*` paths in deployed `page_components` and
  `site_components`. On 2026-07-31 that was **109 of 127 matched, all 109 serving HTTP
  200** — any regression moves that number.
- `favicon.png` and `og-card.png` must still resolve for all 12 sites that own them.
- `TestImageURL404_BrandHeadPathsResolveThroughTheirOwnMap` and
  `TestDeployedWebPathCannotExpressBrandHeadPaths` must both be considered — a fix that
  satisfies one by breaking the other has not resolved the duplication, it has moved it.

## Related

`bugs_open/142` (the lane that found the brand-head drift and wrote the landmine),
`bugs_closed/128` (the near-miss and the audit), `bugs_closed/124` (the scope precedent),
`LANDMINES.md` § *"`storage.DeployedWebPath` … is silently WRONG for `og_card`"* and
§ *"The `assets` table records NO served path"*.
