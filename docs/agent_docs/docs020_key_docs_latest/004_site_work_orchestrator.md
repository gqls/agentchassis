# Dispatch Loop: Dynamic Work Item Routing

Implementation notes for the fix items dispatch loop in site-work-orchestrator, and the self-contained asset-deployer enhancement.

---

## Problem

Discovery agents scan sites and write findings to `site_work_items` with different `handler_agent` values — `webdesign-agent` for missing CSS, `asset-deployer` for undeployed images. The orchestrator had no mechanism to process these items. The existing `build_items_loop` is hardcoded to filter `handler_agent: "page-content-writer"` and `item_domain: "build"`. CSS and asset items were written to the DB but never acted on.

For finetuning.uk, three items sat in `detected` status:

| Item | Handler | Priority | Spec |
|---|---|---|---|
| missing_css | webdesign-agent | 50 | `{"check": "missing_css"}` |
| undeployed_asset (hero) | asset-deployer | 60 | `{"purpose": "hero", "asset_id": "714d..."}` |
| undeployed_asset (logo) | asset-deployer | 60 | `{"purpose": "logo", "asset_id": "9c9d..."}` |

---

## Design Decisions

### Standard spawn→call, not custom routing

The initial approach was to bypass spawning entirely and construct Kafka topic names directly (a "universal topic fallback"), since we wouldn't know which handler types to spawn until runtime. This was wrong — `spawn_agent` already supports `agent_type_field` for dynamic type resolution from collected_data. The dispatch loop uses the same spawn→call pattern as everywhere else in the system.

### Fixed role, dynamic type

Each loop iteration spawns the handler dynamically and finds it by a fixed role:

- `spawn_handler`: `role: "fix_handler"`, `agent_type_field: "current_fix_item.handler_agent"`
- `call_handler`: `target_role: "fix_handler"`

`findAgentByRole` scans all collected_data for matching roles — no `spawn_`-prefix filtering. This is how page-content-writer's research agent works (fixed role `"researcher"`, static type `"research-agent"`). The dispatch loop adds dynamic type resolution on top of the same mechanism.

### Handlers don't know about work items

The orchestrator maps spec fields to handler input_data via `input_mapping` with `?` suffixes for optional fields. Handlers receive raw identifiers (`site_id`, `domain`, `asset_id`, `purpose`) and load their own context. Status tracking (marking in_progress/complete) stays in the orchestrator loop.

The webdesign-agent can be tested with just `{"site_id": "...", "domain": "..."}`. The asset-deployer can be tested with `{"asset_id": "...", "purpose": "hero", "domain": "..."}`. Neither knows it was triggered by a work item.

### Asset-deployer resolves its own URI

Discovery agents store presigned HTTPS URLs in work item specs, but `deploy_image_asset` needs s3:// URIs. Rather than patching specs or having the orchestrator resolve URIs (leaking handler knowledge upward), the asset-deployer resolves the URI itself from the `asset_id` via DB lookup.

---

## Changes

### SQL: site-work-orchestrator workflow

Three existing step transitions changed, four new steps added.

**Modified transitions:**

| Step | Field | Was | Now |
|---|---|---|---|
| ensure_site_record | next_step | call_site_planner | check_mode |
| check_has_items | else_step | apply_site_design | load_fix_items |
| build_items_loop | next_step | apply_site_design | load_fix_items |

**New steps:**

| Step | Action | Routes to |
|---|---|---|
| check_mode | conditional | maintenance → select_style_collection, build → call_site_planner |
| load_fix_items | load_work_items | All triaged items, no handler filter → check_has_fix_items |
| check_has_fix_items | conditional | has items → fix_items_loop, else → apply_site_design |
| fix_items_loop | loop | spawn_handler → call_handler → mark_complete per item |

**Build path flow:**

```
spawn chain → ensure_site_record → check_mode [build] → call_site_planner
→ ... planning and asset generation ...
→ build_items_loop (content pages)
→ load_fix_items → check_has_fix_items → fix_items_loop
→ apply_site_design → complete
```

**Maintenance path flow:**

```
spawn chain → ensure_site_record → check_mode [maintenance]
→ select_style_collection → set_defaults → render_components
→ load_work_items (content, likely empty) → check_has_items [else]
→ load_fix_items → check_has_fix_items → fix_items_loop
→ apply_site_design → complete
```

### SQL: finetuning.uk triage

Triaged the three detected items to `triaged` status. No spec patching — the asset-deployer resolves URIs itself.

### Go: PresignedURLToS3URI

Added to `platform/storage/uri_helpers.go`, alongside existing `ParseS3URI`, `IsS3URI`, `BuildS3URI`.

Converts presigned S3/B2 HTTPS URLs to s3:// URIs by stripping query parameters and extracting bucket + key from the path. Uses `BuildS3URI` internally.

### Go: resolveStorageURIFromAsset

Added to `deploy_image_asset_action.go` as a new fallback after the existing `findStorageURI` call. Also added `"asset_id"` to the action's optional inputs in `DeployImageAssetInputSpec`.

Resolution priority when `s3_uri` is empty:

1. `inputs.Get("s3_uri")` — direct input (existing)
2. `findStorageURI(...)` — scans collected_data for pageflow-builder patterns (existing)
3. `resolveStorageURIFromAsset(...)` — DB lookup from `asset_id` **(new)**
    - Priority 3a: `SELECT content_data->>'{purpose}_uri' FROM sites JOIN assets` — the s3:// URI written by StoreAssetAction
    - Priority 3b: `SELECT url FROM assets` → `PresignedURLToS3URI()` — converts the presigned URL

---

## Dispatch Flow: finetuning.uk Example

Triggered with: `{"domain": "finetuning.uk", "mode": "maintenance"}`

### Setup phase

```
spawn_planner, spawn_content_writer, spawn_reviewer → existing spawn chain
ensure_site_record    → loads site record (site_id, domain, content_data)
check_mode            → mode == maintenance → true → skips planning
select_style_collection → resolves "professional-dark"
set_defaults          → writes default_components
render_components     → renders header/footer chrome
load_work_items       → filters for content items → empty
check_has_items       → false → load_fix_items
```

### Fix items loaded

`load_fix_items` returns all three triaged items, sorted by priority (CSS at 50 first, assets at 60).

### Iteration 1: missing_css → webdesign-agent

```
spawn_handler       → agent_type_field resolves "webdesign-agent", role: "fix_handler"
call_handler        → target_role "fix_handler", sends:
                      {"site_id": "uuid", "domain": "finetuning.uk", "check": "missing_css"}
                      (asset_id?, purpose? silently skipped — not in CSS item spec)

  webdesign-agent:
    check_site_context  → no site_context provided → load_site_context
    load_site_context   → queries DB: site, pages, style collection, colors
    analyze_design      → LLM design spec
    generate_css        → LLM CSS generation
    deploy_css          → git commit → GitHub Actions → S3 → live
    complete            → responds to orchestrator

mark_complete       → item status = 'complete', commit_sha stored
```

### Iteration 2: undeployed_asset (hero) → asset-deployer

```
spawn_handler       → agent_type_field resolves "asset-deployer", role: "fix_handler"
                      (overwrites previous spawn_handler — same output_field, new agent)
call_handler        → target_role "fix_handler", sends:
                      {"site_id": "uuid", "domain": "finetuning.uk",
                       "asset_id": "714dbd9c...", "purpose": "hero"}

  asset-deployer (deploy_image_asset):
    inputs.Get("s3_uri")                  → "" (not sent)
    findStorageURI(...)                   → "" (no pageflow collected_data)
    resolveStorageURIFromAsset("714d...", "hero"):
      content_data->>'hero_uri'           → "s3://personae-prod-uk001-images/.../f400cbe1...png"
    DownloadOptimizeAndPrepare()          → downloads from S3, resizes 1600x900
    git commit /assets/images/hero.jpg    → GitHub Actions → S3 → live
    complete                              → responds to orchestrator

mark_complete       → item status = 'complete'
```

### Iteration 3: undeployed_asset (logo) → asset-deployer

Same as iteration 2:
- `asset_id: "9c9de5a0..."`, `purpose: "logo"`
- Resolves via `content_data->>'logo_uri'`
- Resizes to 400x400 (logo config from ImagePurposes)
- Commits `/assets/images/logo.png`

### Completion

```
fix_items_loop      → all 3 items processed
apply_site_design   → final design pass (existing step, runs regardless)
update_site_status  → complete
```

### End state

```
site_work_items:
  missing_css      | complete | result: {commit_sha: "abc123"}
  undeployed_asset | complete | result: {commit_sha: "def456"} (hero)
  undeployed_asset | complete | result: {commit_sha: "ghi789"} (logo)

finetuning.uk files deployed:
  /assets/css/styles.css  ← professional-dark theme
  /assets/images/hero.jpg ← 1600x900 optimized
  /assets/images/logo.png ← 400x400 optimized
```

Each fix is an individual git commit. GitHub Actions fires per commit, writes to S3. The site accumulates fixes incrementally.

---

## Adding New Handler Agents

The dispatch loop is generic. Adding a new handler requires no changes to the orchestrator.

1. Create the agent definition with its own workflow (load own data → do work → complete)
2. Discovery agents write items with `handler_agent: "your-new-agent"`
3. The dispatch loop spawns and calls it automatically

If a group of fixes tend to run together (e.g. CSS regeneration always followed by asset redeployment), a wrapper agent can own that sequence. The orchestrator dispatches one item to the wrapper, the wrapper handles internal coordination.

---

## Files Changed

| File | Change |
|---|---|
| `platform/storage/uri_helpers.go` | Added `PresignedURLToS3URI` |
| `platform/orchestration/actions/deploy_image_asset_action.go` | Added `resolveStorageURIFromAsset` fallback, `"asset_id"` to optional inputs |
| `agent_definitions` (site-work-orchestrator) | 3 modified transitions, 4 new steps |
| `agent_definitions` (asset-deployer) | No changes needed — existing workflow handles new inputs |
| `site_work_items` (finetuning.uk) | Triaged 3 detected items |
| `001_development_guide_new_agents.md` | Expanded spawn→call docs, dynamic dispatch pattern, findAgentByRole trap |
| `002_system_architecture.md` | Added dispatch pattern section, resolved decisions 16-17 |