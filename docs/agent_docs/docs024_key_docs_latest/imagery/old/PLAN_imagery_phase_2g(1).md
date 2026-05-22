# PLAN — Phase 2G: imagery in the site plan

Replaces the earlier sketch in `STATUS_imagery_2026-05-12.md` (which
incorrectly put structured imagery requirements in `site_specs`). The
corrected design lives in the `site_plan` domain alongside text and
design directives.

Decisions in this plan are locked. See "Decisions taken" section.
For the broader site_spec / site_plan architecture, see
`FOCUS_site_spec_vs_site_plan.md`.

---

## Decisions taken (2026-05-12)

| Question | Decision |
|---|---|
| Directives or sibling table? | **Sibling table** `site_plan_imagery`. Cleaner structural columns than overloading the directive `directive` text. |
| Should cascade apply? | **Yes** — site-scope `imagery_direction` directive flows through to per-imagery generation via the brief renderer pattern. |
| Update pageflow-builder? | **No** — pageflow-builder is being left behind. New flows use triaged work items + image-build-handler + plan-builder. Backup of current state saved at `pageflow-builder_2026-05-12.txt`. |
| `kind` enum scope | `logo`, `hero`, `illustration`, `icon`, `infographic`. **No `product`** (products come from the affiliate_products resolver, not from the planner). |
| Style hints — cascade or override? | **Cascade.** Site-scope `imagery_direction` is the default; row `style_hints` extends/overrides at generation time. |
| Migration strategy | **Age-out**, with an active detection check (`check_legacy_image_prompts_aspect`) that emits a fix work item when sites are still using the legacy shape. |

Phase 1 plan-domain confirmed live as of 2026-05-12 (`site_plans` and
`site_plan_directives` populated for at least one site). Phase 2G can
proceed independently.

---

## Why this matters

Today, the planner emits a flat dictionary of image prompts:

```json
"image_prompts": {
  "logo": "...",
  "hero_home": "...",
  "hero_about": "..."
}
```

Three structural problems:

1. **Hero/logo only.** Names like `hero_about` are baked into discovery
   and routing. Illustrations, icons, infographics, product images have
   no place to live.
2. **Implicit kind, no parameters.** A prompt is a string. Downstream
   handlers can't vary generation parameters because they don't know
   what kind of image is being requested.
3. **No scope.** All prompts live at site level. A page-specific
   illustration or section-specific imagery requirement has nowhere to
   sit that's structurally aware of where it applies.

Text and design direction don't have these problems because they live
in `site_plan_directives` with `scope` / `scope_ref` / `category` /
`subject` columns and a cascade renderer. Phase 2G brings imagery into
the same pattern, with structured columns appropriate for image
generation rather than free-text directives.

---

## What Phase 2G delivers

Six things, in dependency order:

1. **New table `site_plan_imagery`** — sibling to `site_plan_directives`.
   Structured columns for image generation; same scope/scope_ref pattern;
   same lock-transfer mechanism.
2. **`write_site_plan` extension** — writes imagery rows alongside
   directives in the same transaction; transfers locks the same way.
3. **Planner prompt extension** — `build-site-planner` / `plan-builder`
   teaches the LLM to emit an `imagery` block alongside `pages`,
   `design_direction`, `content_strategy`.
4. **New discovery check `check_unfulfilled_imagery_plan`** — walks
   `site_plan_imagery` rows, emits work items for missing assets.
5. **`image-build-handler` extension** — accepts the new spec shape
   with `kind` field; passes through to image-generator. Routes through
   the existing variant chain.
6. **Age-out check `check_legacy_image_prompts_aspect`** — detects
   sites still using `site_specs.site_plan.image_prompts` and emits
   `needs_replan` work items to trigger a fresh plan-builder run.

The legacy `image_prompts` dictionary continues working during
transition. Both discovery checks fire; dedup catches overlaps. The
age-out check actively migrates sites forward as they're noticed.

---

## On leaving pageflow-builder behind

Decision: Phase 2G does not extend `pageflow-builder` with the new
imagery shape. Rationale:

- pageflow-builder writes `site_specs.site_plan` directly (legacy
  aspect), bypassing the four plan-domain tables.
- It uses inline `deploy_image_asset` calls (the same architecture
  mismatch that motivated Phase 2F's refactor of image-build-handler).
- It has hardcoded `generate_logo`/`generate_hero_image` steps — no
  variant support, no extensibility for new image kinds.
- Sequential page-building rather than triaged work item dispatch.

The architecture is converging on:
- **Plan production**: build-site-planner / plan-builder writing to
  plan-domain tables.
- **Page work**: triaged `needs_page:<name>` items handled by
  page-build-handler.
- **Image work**: triaged `unfulfilled_*` items handled by
  image-build-handler (now Phase 2F refactored to spawn asset-deployer).

pageflow-builder is left running for sites it has built historically.
Those sites' imagery continues to work via the legacy
`check_unfulfilled_image_prompt` check. The age-out mechanism (component
6 above) is the path forward for migrating them when needed.

**Backup of pageflow-builder's current state** is saved as
`pageflow-builder_2026-05-12.txt` with explanatory notes in
`pageflow-builder_2026-05-12_NOTES.sql`. If the decision needs reversing,
the snapshot is the recovery reference point.

The `recommended_builder` field in the classifier prompt still defaults
to `pageflow-builder`. That's a separate concern — when classifier output
shifts to recommending `build-site-planner` as the entry, or when the
intake-orchestrator routes around pageflow-builder structurally, the
field can update. Not in scope for 2G.

---

## Schema

```sql
CREATE TABLE site_plan_imagery (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id       uuid NOT NULL REFERENCES site_plans(id) ON DELETE CASCADE,
    scope         text NOT NULL,                 -- 'site' | 'page' | 'section'
    scope_ref     text,                          -- NULL | page_name | '<page_name>:<ordering>'
    key           text NOT NULL,                 -- asset_key (e.g. 'logo', 'hero_about', 'illustration_team')
    kind          text NOT NULL,                 -- 'logo' | 'hero' | 'illustration' | 'icon' | 'infographic'
    prompt        text NOT NULL,
    style_hints   jsonb,                         -- {medium: 'line drawing', mood: 'warm', ...}
    constraints   jsonb,                         -- {aspect: '1:1', transparent_background: true, ...}
    ordering      int NOT NULL DEFAULT 0,
    source        text NOT NULL DEFAULT 'llm',   -- 'llm' | 'classifier' | 'manual'
    locked_at     timestamptz,
    locked_by     text,
    created_at    timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT chk_scope CHECK (scope IN ('site', 'page', 'section')),
    CONSTRAINT chk_kind CHECK (kind IN ('logo', 'hero', 'illustration', 'icon', 'infographic')),
    CONSTRAINT chk_scope_ref_consistency CHECK (
        (scope = 'site' AND scope_ref IS NULL)
        OR (scope = 'page' AND scope_ref IS NOT NULL)
        OR (scope = 'section' AND scope_ref IS NOT NULL AND scope_ref LIKE '%:%')
    )
);

CREATE UNIQUE INDEX idx_site_plan_imagery_unique
    ON site_plan_imagery (plan_id, scope, COALESCE(scope_ref, ''), key);

CREATE INDEX idx_site_plan_imagery_plan
    ON site_plan_imagery (plan_id, scope, scope_ref);

CREATE INDEX idx_site_plan_imagery_locks
    ON site_plan_imagery (plan_id)
    WHERE locked_at IS NOT NULL;
```

Design rationale:

- **`kind` enum is checked.** Five values: `logo`, `hero`, `illustration`,
  `icon`, `infographic`. New kinds added via migration. Keeps downstream
  routing deterministic.
- **`product` deliberately excluded.** Product images come from the
  `query.affiliate_products` resolver, not the planner. The planner
  doesn't know which products exist at plan time.
- **`prompt` is required.** Even logos get a prompt today; consistency
  matters.
- **`style_hints` and `constraints` are JSONB.** Loose shape for v1,
  tightenable later if conventions emerge.
- **`ordering` for multi-imagery in same scope.** Two illustrations on
  the same section get distinct keys; `ordering` preserves declared
  order for deterministic generation.
- **Lock columns mirror `site_plan_directives`.** Same lock-transfer
  treatment.
- **Unique on `(plan_id, scope, scope_ref, key)`** with `COALESCE` for
  the NULL case. Prevents duplicate imagery rows for the same target.

---

## Planner output shape

The plan-builder LLM call adds an `imagery` key to its JSON output:

```json
{
  "pages": [...],
  "design_direction": {...},
  "content_strategy": {...},
  "imagery": {
    "site": [
      {
        "key": "logo",
        "kind": "logo",
        "prompt": "A precise, technical logomark for Robot-Hands.com..."
      },
      {
        "key": "hero_canonical",
        "kind": "hero",
        "prompt": "A dramatic, high-contrast close-up..."
      }
    ],
    "pages": {
      "about": [
        {
          "key": "hero_about",
          "kind": "hero",
          "prompt": "A wide-angle view of a modern automated production line..."
        },
        {
          "key": "illustration_team_values",
          "kind": "illustration",
          "prompt": "Stylised group of engineers collaborating around a workbench...",
          "style_hints": {"medium": "line drawing", "mood": "collaborative"}
        }
      ],
      "tools": [
        {
          "key": "hero_tools",
          "kind": "hero",
          "prompt": "An engineering workspace abstraction..."
        }
      ]
    },
    "sections": {
      "home:2": [
        {
          "key": "icon_precision",
          "kind": "icon",
          "prompt": "Geometric icon representing precision engineering",
          "constraints": {"aspect": "1:1", "transparent_background": true}
        }
      ]
    }
  }
}
```

The planner prompt extension teaches the LLM:

- Site-scope imagery is for things appearing across pages (logo, brand
  hero). Two or three entries is typical.
- Page-scope imagery is for page-specific heroes (`hero_about`,
  `hero_tools`) and decorative imagery the page calls for.
- Section-scope imagery is for icons, illustrations, or infographics
  attached to a specific section on a specific page. Mostly empty in
  v1; populated as components evolve.
- Each entry must have `key`, `kind` (from the enum), `prompt`.
  `style_hints` and `constraints` are optional.

The existing `image_prompts` dictionary continues to be emitted by the
planner during transition (so legacy discovery still works for sites
that don't yet have the new shape). `write_site_plan` writes both:
the new `site_plan_imagery` rows AND continues writing the legacy
`site_specs.site_plan.image_prompts` aspect (for now).

---

## `write_site_plan` extension

The action gains a new step: after writing pages, sections, and
directives, it walks the `imagery` block and writes rows to
`site_plan_imagery`.

```go
// Pseudocode for the imagery-write step within write_site_plan

func writeImagery(tx *sql.Tx, planID uuid.UUID, imagery map[string]interface{}) error {
    // Site scope
    if siteEntries, ok := imagery["site"].([]interface{}); ok {
        for ord, raw := range siteEntries {
            entry := raw.(map[string]interface{})
            if err := insertImageryRow(tx, planID, "site", nil, entry, ord); err != nil {
                return err
            }
        }
    }

    // Page scope
    if pageMap, ok := imagery["pages"].(map[string]interface{}); ok {
        for pageName, raw := range pageMap {
            entries := raw.([]interface{})
            for ord, e := range entries {
                entry := e.(map[string]interface{})
                if err := insertImageryRow(tx, planID, "page", &pageName, entry, ord); err != nil {
                    return err
                }
            }
        }
    }

    // Section scope — scope_ref is "page_name:ordering"
    if sectionMap, ok := imagery["sections"].(map[string]interface{}); ok {
        for sectionRef, raw := range sectionMap {
            entries := raw.([]interface{})
            for ord, e := range entries {
                entry := e.(map[string]interface{})
                if err := insertImageryRow(tx, planID, "section", &sectionRef, entry, ord); err != nil {
                    return err
                }
            }
        }
    }

    return nil
}
```

`insertImageryRow` enforces the kind enum, requires `key` and `prompt`,
passes `style_hints` / `constraints` as JSONB.

**Lock transfer** for imagery rows follows the same pattern as
directives. After writing the new rows, walk the previous current
plan's locked imagery rows, match on `(scope, scope_ref, key)`, copy
`locked_at`, `locked_by`, and (if HITL-edited prompt differs) the
prompt text. Locked HITL versions win on rebuild.

---

## Cascade at generation time

`style_hints` per imagery row are **additive to** the site-scope
`imagery_direction` directive, not a replacement.

At generation time, image-build-handler (or whatever assembles the
image-generator prompt) reads:

1. **Site-scope `imagery_direction`** from `site_plan_directives`
   (existing — already in `siteScopeDesignDirectiveSpecs`).
2. **Page-scope `imagery_direction`** if any (future addition; not in
   2G).
3. **Per-imagery `style_hints`** from the `site_plan_imagery` row.

The composed prompt has the format the Phase 0.1 work established:

```
Style direction: <imagery_direction> [+ <style_hints flattened>]

Subject: <imagery row's prompt>
```

The brief renderer (`datahelpers/page_brief.go`) already handles the
site → page cascade for text-style directives. Extending it to also
provide imagery-relevant context is a small addition that benefits the
image-build-handler's prompt composition step.

Practically: existing `getImageryDirectionForSite` helper continues to
provide the site-scope direction. A new helper
`getImageryHintsForRow(planID, scope, scopeRef, key)` reads the row's
`style_hints` JSONB and returns the cascade-merged result. Both are
called by the prompt composer.

---

## Discovery check 1: `check_unfulfilled_imagery_plan`

New file `check_unfulfilled_imagery_plan.go` in the discovery checks
directory, following the existing pattern (~80 lines).

```go
// Pseudocode

func (c *UnfulfilledImageryPlanCheck) Run(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]WorkItemSpec, error) {
    planID, err := getCurrentPlanID(ctx, db, siteID)
    if err != nil || planID == uuid.Nil {
        return nil, err
    }

    rows, err := db.QueryContext(ctx, `
        SELECT scope, scope_ref, key, kind, prompt, style_hints, constraints
        FROM site_plan_imagery
        WHERE plan_id = $1
    `, planID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var items []WorkItemSpec
    for rows.Next() {
        var scope, scopeRef, key, kind, prompt sql.NullString
        var styleHints, constraints sql.NullString
        // ... scan ...

        assetKey := computeAssetKey(scope.String, scopeRef.String, key.String)

        exists, err := hasActiveAssetForAssetKey(ctx, db, siteID, assetKey)
        if err != nil {
            return nil, err
        }
        if exists {
            continue
        }

        items = append(items, WorkItemSpec{
            ItemType:     "unfulfilled_imagery_plan",
            ItemKey:      "imagery_plan:" + assetKey,
            HandlerAgent: "image-build-handler",
            Spec: map[string]interface{}{
                "purpose":      kind.String,    // for image-build-handler routing
                "asset_key":    assetKey,
                "prompt":       prompt.String,
                "kind":         kind.String,
                "style_hints":  parseJSON(styleHints.String),
                "constraints":  parseJSON(constraints.String),
            },
        })
    }

    return items, nil
}
```

**`computeAssetKey`** namespaces the key by scope:

- `scope=site, key='logo'` → `logo`
- `scope=site, key='hero_canonical'` → `hero_canonical`
- `scope=page, scope_ref='about', key='hero_about'` → `hero_about`
- `scope=page, scope_ref='about', key='illustration_team_values'`
  → `page.about.illustration_team_values`
- `scope=section, scope_ref='home:2', key='icon_precision'`
  → `section.home.2.icon_precision`

The site-scope and obvious page-scope hero/logo names stay simple
(backward compatible with existing deploy paths). Deeper-nesting keys
namespace explicitly to avoid collisions. The deploy path translator
extends naturally: `_` → `-`, `.` → `/`, producing
`assets/images/page/about/illustration-team-values.jpg`.

---

## Discovery check 2: `check_legacy_image_prompts_aspect` (the age-out mechanism)

New file. Detects sites still using `site_specs.site_plan.image_prompts`
instead of `site_plan_imagery` rows. Emits `needs_replan` work items.

```go
// Pseudocode

func (c *LegacyImagePromptsAspectCheck) Run(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]WorkItemSpec, error) {
    // Does the site have a current plan with imagery rows?
    var imageryRowCount int
    err := db.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM site_plan_imagery spi
        JOIN site_plans sp ON sp.id = spi.plan_id
        WHERE sp.site_id = $1 AND sp.is_current = true
    `, siteID).Scan(&imageryRowCount)
    if err != nil {
        return nil, err
    }

    if imageryRowCount > 0 {
        return nil, nil // already on new shape, nothing to do
    }

    // Does the site have legacy image_prompts in site_specs?
    var legacyExists bool
    err = db.QueryRowContext(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM site_specs
            WHERE site_id = $1
              AND aspect = 'site_plan'
              AND is_current = true
              AND data ? 'image_prompts'
        )
    `, siteID).Scan(&legacyExists)
    if err != nil {
        return nil, err
    }

    if !legacyExists {
        return nil, nil // nothing to migrate
    }

    // Emit a replan work item
    return []WorkItemSpec{{
        ItemType:     "needs_replan",
        ItemKey:      "needs_replan:legacy_imagery",
        HandlerAgent: "build-site-planner",  // or plan-builder once renamed
        Priority:     30,                     // not urgent
        Severity:     "low",
        Summary:      "Site uses legacy image_prompts aspect; replan to populate site_plan_imagery",
        Spec: map[string]interface{}{
            "reason":          "legacy_imagery_aspect",
            "migration_hint":  "site_plan.image_prompts → site_plan_imagery",
        },
    }}, nil
}
```

The fix is "trigger build-site-planner to run again on this site",
which (once 2G's planner extension lands) will produce
`site_plan_imagery` rows and the legacy aspect becomes superseded.

**Important sequencing detail.** The age-out check has no effect until
the planner extension (step 3 in the work order below) is deployed. If
it fires before then, build-site-planner runs but doesn't produce the
new shape, and we churn. So this check is the LAST thing to enable.
Add it to discovery-checks code in step 6, but leave its registration
out of `design-discovery-agent` until plan-builder is reliably emitting
imagery.

---

## image-build-handler extension

Three changes to the workflow:

1. **Accept `unfulfilled_imagery_plan` in `check_item_type`.** Add a
   new branch alongside `unfulfilled_hero_variant`. Both branches go to
   the variant chain (existing `spawn_image_gen_variant` →
   `call_variant_gen` → `store_variant_asset` → `deploy_variant`).
2. **Pass `kind`, `style_hints`, `constraints`** through the workflow
   into `image-generator` spawn. Stored but not yet differentiating
   generation in this phase.
3. **Image-generator's prompt composer reads `style_hints`** when
   present. Cascade with `imagery_direction` per the rules above.

No changes to `asset-deployer`. The deploy filename derives from
`asset_key` as today.

The `purpose` field in the work item spec stays mapped to `kind` —
existing routing logic doesn't change. For `kind=illustration`,
`purpose=illustration`. The variant chain doesn't care what purpose
says, only what `asset_key` is for filename derivation.

---

## Transition: old `image_prompts` and new `site_plan_imagery`

Both run during transition:

- **`check_unfulfilled_image_prompt`** (existing) reads
  `site_specs.site_plan.image_prompts`. Emits work items for `logo`,
  `hero_home`, etc.
- **`check_unfulfilled_imagery_plan`** (new) reads `site_plan_imagery`
  rows. Emits work items with namespaced asset_keys and `kind`.
- **`check_legacy_image_prompts_aspect`** (new, age-out) emits
  `needs_replan` for sites still using only the legacy shape.

The `idx_swi_dedup` index catches overlapping `item_key` values,
swallowing duplicates. Sites built post-2G have rows in the new table
and the new check processes them. Sites built pre-2G stay on the
legacy path until either:

- They naturally replan (direction change, HITL trigger), or
- The age-out check fires `needs_replan` and the planner picks it up.

**Cutover criteria** (in some weeks' time, not as part of 2G):

- Every site that runs a fresh plan-builder cycle has
  `site_plan_imagery` rows.
- The age-out check has fired across all eligible sites.
- A SQL audit shows no sites with `image_prompts` in
  `site_specs.site_plan` but no `site_plan_imagery` rows.

Then the legacy check can be retired and the old `image_prompts` aspect
backfill check can be disabled.

---

## Sequencing

In dependency order, sized for individual deploys:

1. **Schema migration** — `site_plan_imagery` table with the constraints
   and indexes above. Pure DDL, no behaviour change. Apply alongside
   the Phase 1 plan-domain migration if it hasn't already landed.
2. **`write_site_plan` extension** — `insertImageryRow` helper + lock
   transfer for imagery rows. Go change. Plan-builder won't emit
   imagery yet, so this is dormant until step 3.
3. **Planner prompt extension** — teach plan-builder to emit the
   `imagery` block. Migration of the agent_definition's prompt
   template. First behaviour change: new plans have imagery rows.
4. **`check_unfulfilled_imagery_plan` discovery check** — new Go file.
   Wired into `design-discovery-agent`. Walks `site_plan_imagery` and
   emits work items.
5. **image-build-handler accepts `kind`** — workflow migration to add
   the third branch for `unfulfilled_imagery_plan`. Spec accepts `kind`,
   `style_hints`, `constraints`. Routes through variant chain. Prompt
   composer cascades style_hints with imagery_direction.
6. **`check_legacy_image_prompts_aspect`** — written but NOT registered
   in `design-discovery-agent` yet. Add the Go file in step 6, then
   add the registration only after steps 1-5 are stable in production
   (otherwise replans churn without producing new shape).

Each step shippable on its own. Step 1 is harmless additive DDL. Step
2 is dormant until step 3 fires. Steps 3-5 can land in any order —
they won't fully connect until all three are deployed, but partial
deployment is safe (new imagery rows sit unread; new discovery check
finds nothing to process; image-build-handler branch never fires).
Step 6 is gated separately.

The transition strategy means **the old path keeps working** through
all of this. Heroes and logos continue to generate via the legacy
check + image-build-handler's existing logo/hero branches throughout
the transition.

---

## What 2G doesn't include

Deferred to Phase 2H or later:

- **Generation parameter differentiation by `kind`.** 2G makes `kind`
  available and cascades `style_hints`; 2H makes generation parameters
  (style_preset, cfg_scale, negative_prompt) vary by kind.
- **True SVG icon generation.** `kind=icon` in 2G produces a small
  raster image. Real SVG output is a different generator.
- **Infographic generation.** `kind=infographic` is in the enum but
  behaviour is the same as `illustration` for now. Real infographics
  need chart/data viz generators.
- **Multi-provider routing.** Stays Stability-only.
- **Product imagery.** Comes from the `query.affiliate_products`
  resolver (sibling doc), not the planner.
- **Auditor awareness of imagery rows.** Visual auditor extension
  (Phase 4) reads these once it ships.
- **Updating pageflow-builder.** Decision logged above; pageflow-builder
  continues unchanged. Sites built via it use legacy path until they
  age out.

---

## What changes for downstream after 2G ships

For consumers of the imagery plan (image-build-handler, image-generator,
asset-deployer):

- A new work item type `unfulfilled_imagery_plan` carrying `kind`,
  namespaced `asset_key`, `prompt`, optional `style_hints`,
  `constraints`.
- Asset_keys may now be hierarchical
  (`page.about.illustration_team_values`) rather than flat. Deploy
  paths translate naturally.
- The `assets` table picks up rows with these richer asset_keys. No
  schema change to `assets`.
- The legacy `image_prompts` aspect still works. No retirement before
  cutover.

For consumers that look at the strategic spec (auditors,
content-quality-auditor):

- No change at the `site_specs` level. Phase 2G touches `site_plan`
  tables only.

For the planner itself:

- One additional block in the JSON output (`imagery`).
- One additional step in the workflow (the `write_site_plan` action
  picks it up automatically once the Go side is in place).

For HITL operators:

- Imagery rows are lockable like directives. Lock a row, the prompt
  text survives plan rebuilds. The asset that's already in `assets`
  doesn't get regenerated (existing asset-locking behaviour from
  Phase 2A).

For sites on the legacy path:

- They continue to work. The age-out check, once enabled, eventually
  produces `needs_replan` for them. Until then, no change.
