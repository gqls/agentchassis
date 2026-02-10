## Pageflow-Builder Image Generation - Changes Summary

### Overview
Add AI-generated hero images to pageflow-builder workflow with relative URLs.

### Flow After Changes
```
generate_hero_image → image-generator adapter → returns image_uri, image_url
    ↓
store_hero_asset → stores in assets table (existing)
    ↓
deploy_hero_image (NEW) → downloads from S3, optimizes, commits to git
    ↓
select_style_collection → ...build_pages_loop
    ↓
Templates use {{.hero_url}} → /assets/images/hero.jpg
```

---

### Files to Add

#### 1. platform/storage/image_processing.go (NEW FILE)
Copy `storage_image_processing.go` to the storage package.
```go
// Contains:
// - DownloadAndOptimizeImage(ctx, s3URI, purpose, logger) ([]byte, error)
// - OptimizeImageForWeb(imageData, purpose, logger) ([]byte, error)
// - ImageDimensions map for different purposes
```

#### 2. platform/orchestration/actions/deploy_image_asset_action.go (NEW FILE)
Copy `deploy_image_asset_action.go` to the actions package.
```go
// DeployImageAssetAction:
// - Downloads image from S3 using storage.DownloadAndOptimizeImage
// - Base64 encodes for git
// - Sends commit request to git-adapter
```

---

### Files to Modify

#### 3. platform/orchestration/actions/action_registry.go
Add to the actionHandlers map:
```go
"deploy_image_asset": DeployImageAssetAction,
```

#### 4. platform/orchestration/actions/v3_site_actions.go

**In StoreAssetAction** (after asset insert, before final return):
```go
// Add after "if err != nil" error handling for asset insert...

// Get purpose from config
purpose := ""
if p, ok := config["purpose"].(string); ok && p != "" {
    purpose = p
}

// If purpose is set, store S3 URI in content_data for deploy_image_asset
if purpose != "" && siteID != nil {
    // Find S3 URI from asset data
    s3URI := ""
    if assetDataMap, ok := assetData.(map[string]interface{}); ok {
        if uri, ok := assetDataMap["image_uri"].(string); ok {
            s3URI = uri
        }
    }
    
    // Store URI in content_data
    if s3URI != "" {
        updateQuery := `
            UPDATE sites 
            SET content_data = jsonb_set(
                COALESCE(content_data, '{}'::jsonb),
                $2::text[],
                to_jsonb($3::text),
                true
            ),
            updated_at = NOW()
            WHERE id = $1
        `
        jsonPath := fmt.Sprintf("{%s_uri}", purpose)
        execDB(ctx, params.DB, updateQuery, *siteID, jsonPath, s3URI)
        
        // Also store in collected_data
        params.CollectedData[purpose+"_uri"] = s3URI
    }
}
```

**In mergeIntoRenderContextEnhanced** (after phone/email extraction):
```go
// Extract image URLs
for _, field := range []string{"hero_url", "hero_home_url", "logo_url"} {
    if v, ok := data[field].(string); ok && v != "" {
        if ctx.ContentData == nil {
            ctx.ContentData = make(map[string]interface{})
        }
        ctx.ContentData[field] = v
    }
}
```

**In renderCtxToMap** (after services_html):
```go
// Add image URLs to result
if ctx.ContentData != nil {
    for _, field := range []string{"hero_url", "hero_home_url", "logo_url"} {
        if url, ok := ctx.ContentData[field].(string); ok && url != "" {
            result[field] = url
        }
    }
}
```

**In BuildRenderContextAction** (after sources loop):
```go
// Check hero_deployed output from deploy_image_asset
if heroDeployed, ok := params.CollectedData["hero_deployed"].(map[string]interface{}); ok {
    if imageURL, ok := heroDeployed["image_url"].(string); ok && imageURL != "" {
        if renderCtx.ContentData == nil {
            renderCtx.ContentData = make(map[string]interface{})
        }
        renderCtx.ContentData["hero_url"] = imageURL
    }
}
```

---

### Database Changes (run pageflow_image_workflow.sql)

1. **Add deploy_hero_image step** to pageflow-builder workflow
2. **Update store_hero_asset.next_step** → "deploy_hero_image"
3. **Update hero template** to use `{{.hero_url}}` background image

---

### Go Dependency
```bash
go get github.com/nfnt/resize
```

---

### Verification After Deploy

1. Trigger a new site build
2. Check logs for:
    - "Downloading image from S3"
    - "Image downloaded and optimized"
    - "Image deploy request sent to git adapter"
3. Verify git commit includes `assets/images/hero.jpg`
4. Verify hero section has background-image style

---

### File Locations Summary
```
platform/
├── storage/
│   ├── s3.go                    (existing)
│   └── image_processing.go      (NEW)
└── orchestration/
    └── actions/
        ├── action_registry.go         (add registration)
        ├── deploy_image_asset_action.go (NEW)
        └── v3_site_actions.go         (patches)
```