# Go batch — slice 3a: deploy_asset records the local URL (prevents recurrence)

The backfill (w9_04) fixes existing rows; this change makes the deployer record the
repo-local path at every future deploy, for ALL kinds — heroes included, which matters
because on any site without the legacy site-level `content_data.hero_url`, the hero's
`background_image` resolves straight from `assets.url` and would otherwise render a
presigned URL that dies in seven days.

## Edit F — deploy_image_asset_action.go (EXACT anchors, from the uploaded file)
Insertion point: immediately after the git commit succeeds and the result fields are
set. `processed.Paths.RelativeURL` / `.Filename` are the local URL and filename the
action just committed; `params.DB` (nil-checked pattern exists at :118), `inputs`,
`uuid` and `zap` are all in scope/imported. Four NEW locals are introduced deliberately
(`assetIDStr`, `assetUUID`, `parseErr`, `execErr`), scoped to the new block — no
existing identifiers change. When a caller omits `asset_id` the record is skipped
(best-effort by design; if real runs log the skip, extend matching by site+asset_key).

OLD (exact):
```go
	// Add image URL to result
	result["image_url"] = processed.Paths.RelativeURL
	result["output_path"] = processed.Paths.FilePath
	result["size_bytes"] = len(processed.Data)
```
NEW:
```go
	// Add image URL to result
	result["image_url"] = processed.Paths.RelativeURL
	result["output_path"] = processed.Paths.FilePath
	result["size_bytes"] = len(processed.Data)

	// Record the deployed local URL on the asset row so resolvers (plan_sections'
	// site_assets source) serve the site-local path instead of the expiring
	// presigned storage URL. The pre-update url's unsigned object path is
	// preserved into storage_path (COALESCE = only when empty), mirroring the
	// w9_04 backfill. Best-effort: a failure here must not fail the deploy.
	if params.DB != nil {
		if assetIDStr := inputs.Get("asset_id"); assetIDStr != "" {
			if assetUUID, parseErr := uuid.Parse(assetIDStr); parseErr == nil {
				if _, execErr := params.DB.ExecContext(ctx, `
					UPDATE assets
					SET storage_path     = COALESCE(storage_path, split_part(url, '?', 1)),
					    storage_provider = COALESCE(storage_provider, 'backblaze'),
					    filename         = $2,
					    url              = $3
					WHERE id = $1
				`, assetUUID, processed.Paths.Filename, processed.Paths.RelativeURL); execErr != nil {
					logger.Warn("deploy_image_asset: failed to record local url on asset",
						zap.String("asset_id", assetIDStr),
						zap.Error(execErr))
				} else {
					logger.Info("deploy_image_asset: recorded local url on asset",
						zap.String("asset_id", assetIDStr),
						zap.String("url", processed.Paths.RelativeURL))
				}
			} else {
				logger.Warn("deploy_image_asset: invalid asset_id for local url record",
					zap.String("asset_id", assetIDStr))
			}
		}
	}
```

## Effect
- Resolvers (Edit B, the hero's `background_image`) serve local paths from the first
  build after any deploy — no expiry class remains.
- `storage_path` + `storage_provider` retain where the original bytes live.
- Render-neutral for pages shadowed by legacy `hero_url`; corrective everywhere else.

## Sequencing
w9_04 (backfill) → w9_05 (rebuild + verify) close the live exposure now, before the
07-10 expiry, independent of this deploy. This slice rides the next chassis image with
slice 2 (component_library) and the remaining batch items.
