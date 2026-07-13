# Go batch — slice 3a: deploy_asset records the local URL (prevents recurrence)

The backfill (w9_04) fixes existing rows; this change makes the deployer record the
repo-local path at every future deploy, for ALL kinds — heroes included, which matters
because on any site without the legacy site-level `content_data.hero_url`, the hero's
`background_image` resolves straight from `assets.url` and would otherwise render a
presigned URL that dies in seven days.

## The change (locate the `deploy_asset` action's Go file; upload it and I will anchor
## the edit exactly — the contract below is complete either way)
At the point where the git commit has SUCCEEDED and the final committed filename is
known (the action already holds `domain`, `asset_key`, `purpose`, and it computed the
optimised filename it wrote), execute:

```sql
UPDATE assets
SET storage_path     = COALESCE(storage_path, split_part(url, '?', 1)),
    storage_provider = COALESCE(storage_provider, 'backblaze'),
    filename         = $2,                      -- the committed filename, e.g. illustration-home.jpg
    url              = '/assets/images/' || $2
WHERE id = $1
```
(single statement — the right-hand `url` reads the pre-update value, so the S3 object
path is preserved into storage_path exactly as in the backfill). Log with `logger.Info`
("deploy_asset: recorded local url", asset_key, filename). No variable renames; reuse
the action's existing db handle and asset id.

## Effect
- Resolvers (Edit B, the hero's `background_image`) serve local paths from the first
  build after any deploy — no expiry class remains.
- `storage_path` + `storage_provider` retain where the original bytes live.
- Render-neutral for pages shadowed by legacy `hero_url`; corrective everywhere else.

## Sequencing
w9_04 (backfill) → w9_05 (rebuild + verify) close the live exposure now, before the
07-10 expiry, independent of this deploy. This slice rides the next chassis image with
slice 2 (component_library) and the remaining batch items.
