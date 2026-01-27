## Image Generation Changes Summary

### Files to modify:

1. **platform/orchestration/actions/generate_image_actions.go**
    - Add `StoreGeneratedImageAction` function (from patch_store_generated_image.go)
    - Add `updateSiteContentField` helper function

2. **platform/orchestration/actions/action_registry.go**
    - Add to actionHandlers map:
      ```go
      "store_generated_image": StoreGeneratedImageAction,
      ```

3. **platform/orchestration/actions/v3_site_actions.go**
    - Add imports: `"bytes"`, `"image"`, `"image/jpeg"`, `_ "image/png"`, `"io"`, `"github.com/nfnt/resize"`
    - Add `processAndAddSiteImages`, `downloadFromS3`, `optimizeForWeb` functions (from patch_assemble_site_images.go)
    - In `AssembleMultipageSiteAction`, before final return add:
      ```go
      if err := processAndAddSiteImages(ctx, params, siteFiles, logger); err != nil {
          logger.Warn("Failed to process some images", zap.Error(err))
      }
      ```
    - In `mergeIntoRenderContextEnhanced`, after phone/email extraction add:
      ```go
      // Extract image URLs
      for _, field := range []string{"hero_home_url", "hero_about_url", "logo_url"} {
          if v, ok := data[field].(string); ok && v != "" {
              if ctx.ContentData == nil {
                  ctx.ContentData = make(map[string]interface{})
              }
              ctx.ContentData[field] = v
          }
      }
      ```
    - In `renderCtxToMap`, add image URLs to result map:
      ```go
      if ctx.ContentData != nil {
          for _, field := range []string{"hero_home_url", "hero_about_url", "logo_url"} {
              if url, ok := ctx.ContentData[field].(string); ok && url != "" {
                  result[field] = url
              }
          }
      }
      ```

4. **go.mod**
    - Add dependency: `go get github.com/nfnt/resize`

5. **Database (run image_generation_sql.sql)**
    - Updates multipage-website-builder workflow
    - Updates hero template

### Flow after changes:
```
sync_pages_to_db
    ↓
spawn_image_generator
    ↓
generate_hero_image (calls image-generator agent)
    ↓
store_hero_image (stores S3 URI + relative URL)
    ↓
generate_pages_loop (hero_home_url available in render context)
    ↓
assemble_site (downloads from S3, optimizes, adds to site_files)
    ↓
deploy (image at /assets/images/hero_home.jpg)
```

### Template output:
```html
<section class="hero" style="background: linear-gradient(...), url('/assets/images/hero_home.jpg') center/cover no-repeat;">
```