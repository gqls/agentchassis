# Go batch — slice 1: plan_sections_action.go

Two edits + a deploy of behaviour that already exists in the working copy. Base: the
file as read in this thread (the optional/skip_field branch present). Before applying,
confirm the working copy matches deployed history expectations:
`git log -1 -- platform/orchestration/actions/plan_sections_action.go` — the resolved
fork says the deployed pod PREDATES this file; these edits go on top of the working copy.

## Edit A — the required-branch skip_field case (the smell found in review)
A required field with `on_missing: skip_field` currently falls to `default:` → defers the
whole section. Honour the declared intent instead. Anchor is exact; `logger` is the
identifier already used in this function (e.g. the empty-schema warn earlier in it) —
verify the enclosing func's parameter name before applying and do NOT rename anything.

OLD (exact):
```go
			// Required field missing — apply on_missing
			switch onMissing {
			case "use_fallback":
```
NEW:
```go
			// Required field missing — apply on_missing
			switch onMissing {
			case "skip_field":
				// Required-but-skippable: honour the schema's declared intent and
				// omit the field instead of deferring the section (mirrors the
				// optional branch; templates gate on the field).
				logger.Info("plan_sections: required field missing with on_missing=skip_field — omitting field",
					zap.String("field", fieldName),
					zap.String("source", source))
			case "use_fallback":
```

## Edit B — ensureAssets: the section-scope imagery query (W7c)
Modelled on the per-page hero block directly above the insertion point: same join
(spi → current plan → active asset by key), section scope for this page, mapped BOTH
by key (per-key schema paths, e.g. the icon sets) AND by kind alias first-wins (the
generic `site_assets.illustration` path brief-explanation declares). Insert BEFORE the
`// Site logo.` line.

OLD (exact, the insertion marker):
```go
	// Site logo.
```
NEW:
```go
	// Per-page section imagery: illustrations / icons / infographics requested at
	// section scope for this page (scope_ref = "<page>:<ordinal>"), joined to the
	// deployed asset row. Mapped by KEY (per-key schema paths, e.g. icon sets) and
	// aliased by KIND first-wins (generic paths like site_assets.illustration),
	// mirroring the hero mapping above. Skipped when pageName is empty.
	if r.pageName != "" {
		rows, err := r.db.QueryContext(ctx, `
			SELECT spi.kind, spi.key, a.url
			  FROM site_plan_imagery spi
			  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
			  JOIN assets a ON a.site_id = sp.site_id
			               AND a.asset_key = spi.key
			               AND a.status = 'active'
			 WHERE sp.site_id = $1
			   AND spi.scope = 'section'
			   AND spi.scope_ref LIKE $2 || ':%'
			   AND spi.kind IN ('illustration', 'icon', 'infographic')
			 ORDER BY spi.kind, spi.ordering
		`, r.siteID, r.pageName)
		if err != nil {
			r.logger.Warn("plan_sections: section imagery lookup failed",
				zap.String("page", r.pageName), zap.Error(err))
		} else {
			defer rows.Close()
			for rows.Next() {
				var kind, key, url string
				if err := rows.Scan(&kind, &key, &url); err != nil {
					continue
				}
				if url == "" {
					continue
				}
				r.assets[key] = url
				if _, exists := r.assets[kind]; !exists {
					r.assets[kind] = url
				}
			}
			if err := rows.Err(); err != nil {
				r.logger.Warn("plan_sections: section imagery rows error",
					zap.String("page", r.pageName), zap.Error(err))
			}
		}
	}

	// Site logo.
```

## Build, deploy, verify
1. `go build ./...` (no other files change in this slice).
2. Deploy via the usual chassis image/rollout for the pod that executes page-build-handler
   workflows (the build dispatch/handler deployment — plan_sections runs inside it).
3. Post-deploy: `$PSQL < w8_01_post_deploy_rebuild.sql` (two needs_page items, index +
   tools; previous keys are complete so dedup passes).
4. Expect: brief-explanation returns to both pages WITH the illustration
   (`page_components` gains the row; the section's `<img>` src points at the B2 asset;
   the gate renders the wrapper because the field now resolves). The two
   `needs_section_data` items self-close via `closeResolvedDataRequest`.
5. Greps on the fresh index: `data-component="brief-explanation"` ≥ 1;
   `illustration_home` present in the img src.
