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
~~**Status:** OPEN, unowned. **Do not "fix" this in a bug patch** — see § Scope.~~

---

> # STATUS 2026-08-02 — **CLOSED: FIXED, LIVE on `v1.0.1229`, POD-VERIFIED ON BOTH REPLICAS.**
>
> Taken by the **`bugfix_168_deployed_asset_path`** lane. Council `abd9b119`: **APPROVED at
> round 3** ("with 2 advisory objection(s) — none high-severity"; 11 seats approve).
>
> **The live proof, with a negative control — a positive one cannot tell you your binary
> shipped** (`bugs_open/153`). Same exec, both replicas, after `rollout status` settled:
>
> | control | `79479769b9-g7fbt` | `79479769b9-n8nbj` | meaning |
> |---|---|---|---|
> | positive: `deploy_image_asset` | 5 | 5 | the grep pipeline works |
> | **negative: `Phase 2E: derived variant deploy path`** | **0** | **0** | the string this change **removed** is gone ⇒ it is MY binary, not a stale one |
> | marker: `derived asset deploy path` | 1 | 1 | the unified derivation shipped |
> | marker: `is a brand-head artefact published at` | 1 | 1 | the refusal guard shipped |
>
> **No regression:** all **24** brand-head artefacts (12 sites × `favicon.png` + `og-card.png`)
> serve HTTP 200 after the roll — the criterion this file's own § *Verify a fix* asked for.
>
> The image was built and rolled by another session; it carries this work because
> `make build-*` builds from committed HEAD. That is the mechanism working as designed, not
> luck — but note it also means the *first* commit went live before the council finished, which
> is the shared-HEAD reality the 2026-07-29 ruling describes rather than a process failure.
>
> **Working docs:** `docs/agent_docs/docs024_key_docs_latest/bugfix_168_deployed_asset_path/`
> · **Register:** IMG-067 · **Council gate:** corr `abd9b119-d274-43bf-a03f-cf45bfb6b881`
> · **Diagnosis loop:** corr `ae9404bd-dab7-4606-ade3-c439ebda93af` (**REFUTED** — read § below)
>
> **What shipped:** `storage.DeployedAssetPath` is now THE derivation, resolved through by
> the writer (`deploy_image_asset_action`) *and* all six readers, with `BrandHeadAssetPaths`
> folded in as its **input**. This went beyond fix candidate 1, which fixes only the readers:
> the derivation had been implemented **twice** — once in the deployer's Phase 2E branch,
> once in the helper — kept in step by a doc comment claiming to "mirror" the other. They did
> agree; the defect was that *nothing made them*. `check_image_url_404`'s
> `IsBrandHeadPurpose` branch is removed as redundant. `IsBrandHeadPurpose` itself is kept:
> it stays correct for the different question *"which table holds the evidence it is
> deployed?"*, which is what `check_undeployed_assets` uses it for, unchanged.
>
> **To verify after the next roll** (both replicas — `bugs_open/153`): positive control
> `grep -c "derived asset deploy path"`; **negative control**
> `grep -c "Phase 2E: derived variant deploy path"`, expect **0**. Then re-run the 128
> lane's whole-fleet assertion below. Only then move this file to `bugs_closed/`.
>
> ## CORRECTION to this file's own § "The mechanism, in four lines of the helper"
>
> **The stated mechanism is too broad, and fix candidate 2 below would have made things
> worse.** This file says an underscore purpose stored with `asset_key = purpose` "yields a
> path with an underscore where the deployed file has a hyphen". For a file published by
> `deploy_image_asset` that is **false**: `deploy_image_asset_action.go:185` branches on the
> *identical* condition (`if assetKey != "" && assetKey != purpose`), so the deployed name
> **also** carries the underscore. Helper and deployer agree.
>
> So **candidate 2 ("apply the `_`→`-` swap unconditionally") is a trap** — it would make the
> helper disagree with the actual writer for every future underscore purpose, manufacturing
> the drift this bug exists to remove. It is struck through below rather than deleted.
>
> The true mechanism is one level up, and it is what makes this framework-scope: **there are
> TWO writers**, and `(asset_key, purpose)` cannot express which one published a row —
> `deploy_image_asset_action` (purpose/asset_key derived) and `derive_brand_head_assets_action`
> (fixed literals `favicon.png` / `og-card.png`). Every call site therefore had to learn the
> brand-head exception separately: 016b §9 case 7.
>
> ## The diagnosis loop returned REFUTED, and that is recorded rather than buried
>
> Filed before asserting a structural cause, per the owner ruling of 2026-07-31. **Its
> correct and useful point:** `injectBrandHeadTags` hardcodes both literals and never calls
> `DeployedWebPath` nor reads `assets.url`, so there is **no render-time failure today** —
> independently corroborating this file's own "Low today, latent" severity. That is why the
> fix is framed as removing a drift mechanism, not repairing an outage.
>
> **Where the loop was wrong, and why REFUTED was not read as "no defect":** it asserted
> *"DeployedWebPath's only found call site (queryresolve.go)"*. There are **six**, and one of
> them — `check_image_url_404` — is exactly where the two derivations meet, which is why the
> 128 lane had to add a brand-head branch there to avoid reporting a 404 for the og card and
> favicon of every site in the fleet. Its refutation rested on an incomplete census.
>
> **Genuine side-finding from the loop, NOT fixed here:** several active `favicon`/`og_card`
> rows carry `assets.url = '/assets/images/input-data.asset-key.jpg'`, an unresolved template
> literal. Already documented in `check_undeployed_assets.go` and owned by `bugs_open/152`.

---

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

> **DONE 2026-08-02 — candidate 1, extended to fold the WRITER in as well.** Candidate 1
> alone fixes the readers and leaves the deployer as a second implementation of the same
> rule. Candidate 2 is **struck out as actively harmful** (see the correction banner above).
> Candidate 3 is not needed once the helper is total. Candidate 4 was the status quo.

1. **Teach `DeployedWebPath` the brand-head map** — return `BrandHeadAssetPaths[purpose]`
   when `IsBrandHeadPurpose(purpose)`. Makes the helper correct for every input, so no
   caller can get it wrong. ⚠ It **inverts** the existing
   `TestDeployedWebPathCannotExpressBrandHeadPaths`, which pins the current behaviour
   deliberately *"so the duplication cannot outlive its reason"* — that test is the 142
   lane's, and its inversion is a conversation with them, not a silent edit.
2. ~~**Make the inexpressible case unrepresentable**: apply the `_`→`-` swap
   unconditionally in `DeployedWebPath` rather than only on the else branch, so purpose
   and asset_key are normalised identically. Smaller, but it changes the path for any
   *future* underscore purpose without telling its owner, which is the same class of
   surprise in the other direction.~~
   **STRUCK 2026-08-02 — this is worse than the bug.** The deployer skips the swap on the
   same condition, so doing it unconditionally here makes the helper disagree with the file
   that is actually committed. It does not "change the path for a future underscore purpose
   without telling its owner"; it makes the path **wrong**. See the correction banner.
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
