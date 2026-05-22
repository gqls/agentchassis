# PLAN — Phase 2G: imagery in the site plan

Replaces the earlier proposal sketched in `STATUS_imagery_2026-05-12.md`
which incorrectly put structured imagery requirements in `site_specs`.
The corrected design lives in the `site_plan` domain alongside text and
design directives. See `FOCUS_site_spec_vs_site_plan.md` for the layer
distinction.

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

This shape has three structural problems:

1. **Hero/logo only.** Names like `hero_about` are baked into discovery
   and routing. Illustrations, icons, infographics, product images have
   no place to live.
2. **Implicit kind, no parameters.** A prompt is a string. Downstream
   handlers can't vary generation parameters because they don't know what
   kind of image is being requested.
3. **No scope.** All prompts live at site level. A page-specific
   illustration or section-specific imagery requirement has nowhere to
   sit that's structurally aware of where it applies.

Text and design direction don't have these problems because they live in
`site_plan_directives` with `scope` / `scope_ref` / `category` / `subject`
columns and a cascade renderer. Phase 2G brings imagery into the same
pattern.

---

## What Phase 2G delivers

Five things, in dependency order:

1. **New table `site_plan_imagery`** — sibling to `site_plan_directives`.
   Structured columns for image generation; same scope/scope_ref pattern;
   same lock-transfer mechanism.
2. **Planner output extension** — the plan-builder LLM call also produces
   an `imagery` block with site/page/section entries.
3. **`write_site_plan` extension** — writes imagery rows alongside
   directives in the same transaction; transfers locks the same way.
4. **New discovery check `check_unfulfilled_imagery_plan`** — walks
   `site_plan_imagery` rows, emits work items for missing assets.
5. **`image-build-handler` accepts `kind`** — passes through to the
   spec; routes are the same as today, but the field is available for
   Phase 2H to differentiate generation parameters.

The legacy `image_prompts` dictionary continues working during transition.
Both checks fire; dedup catches overlaps. Cutover when we're confident.

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

**Design choices in this schema:**

- **`kind` is enumerated and checked.** v1 enum:
  `logo, hero, illustration, icon, infographic`. New kinds via migration.
  This keeps downstream routing deterministic. `product` is intentionally
  not here — product images come from the affiliate_products resolver
  (sibling doc), not from the planner. The planner doesn't know which
  products exist.
- **`prompt` is required.** Even a logo gets a prompt — the existing
  `image_prompts.logo` produces one today.
- **`style_hints` and `constraints` are JSONB and optional.** Loose shape
  for v1, tightenable later if we find conventions worth enforcing.
- **`ordering` for multi-imagery within the same scope.** Two section
  illustrations on the same section get distinct keys (`illustration_1`,
  `illustration_2`), but `ordering` preserves declared order for
  deterministic generation.
- **Lock columns mirror `site_plan_directives`.** Same lock-transfer
  treatment in `write_site_plan`.
- **Unique on `(plan_id, scope, scope_ref, key)`.** Prevents duplicate
  imagery rows for the same target. `COALESCE(scope_ref, '')` handles the
  NULL case in the unique index (Postgres treats NULLs as distinct in
  unique indexes without this).

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
  attached to a specific section on a specific page. Mostly empty in v1;
  populated as components evolve.
- Each entry must have `key` (the asset_key), `kind` (from the enum),
  and `prompt`. `style_hints` and `constraints` are optional.

The existing `image_prompts` dictionary continues to be emitted by the
planner during transition. `write_site_plan` writes both shapes (the new
imagery rows and the old `site_specs.site_plan.image_prompts` aspect).

---

## `write_site_plan` extension

The action gains a new step: after writing pages, sections, and directives,
it walks the `imagery` block and writes rows to `site_plan_imagery`.

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

**Lock transfer** for imagery rows follows the same pattern as directives
— after writing the new rows, walk the previous current plan's locked
imagery rows, match on `(scope, scope_ref, key)`, copy `locked_at`,
`locked_by`, and (if HITL-edited prompt differs) the prompt text. Same
"locked HITL versions win on rebuild" semantics.

---

## Discovery check

A new file `check_unfulfilled_imagery_plan.go` in the discovery checks
directory, following the existing pattern (~80 lines).

```go
// Pseudocode

func (c *UnfulfilledImageryPlanCheck) Run(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]WorkItemSpec, error) {
    // Get the current plan for this site
    planID, err := getCurrentPlanID(ctx, db, siteID)
    if err != nil {
        return nil, err
    }
    if planID == uuid.Nil {
        return nil, nil // no plan yet
    }

    // Walk every site_plan_imagery row for the current plan
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

        // Compute the namespaced asset_key
        assetKey := computeAssetKey(scope.String, scopeRef.String, key.String)

        // Does the asset exist?
        exists, err := hasActiveAssetForAssetKey(ctx, db, siteID, assetKey)
        if err != nil {
            return nil, err
        }
        if exists {
            continue
        }

        // Emit a work item
        items = append(items, WorkItemSpec{
            ItemType:     "unfulfilled_imagery_plan",
            ItemKey:      "imagery_plan:" + assetKey,
            HandlerAgent: "image-build-handler",
            Spec: map[string]interface{}{
                "purpose":      kind.String,
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
namespace explicitly to avoid collisions. The deploy path translator can
extend naturally: `_` → `-`, `.` → `/`, producing
`assets/images/page/about/illustration-team-values.jpg`.

---

## image-build-handler extension

Three changes to the workflow:

1. **Accept `kind` in the spec.** The `check_item_type` branch
   recognises `unfulfilled_imagery_plan` (in addition to the existing
   `unfulfilled_image_prompt` and `unfulfilled_hero_variant`).
2. **Route through the variant chain.** For `unfulfilled_imagery_plan`,
   the workflow uses the existing variant branch (already does the right
   thing structurally: stores asset with `purpose + asset_key`, deploys
   per-variant filename). The `kind` field is passed to
   `image-generator` for future use (Phase 2H) but doesn't change
   generation parameters in 2G.
3. **Pass `kind`, `style_hints`, `constraints`** to the image-generator
   spawn. Stored but not yet differentiating generation in this phase.

No changes to `asset-deployer`. The deploy filename derives from
`asset_key` as today.

---

## Transition: old `image_prompts` and new `site_plan_imagery`

Both run during transition:

- **Old check `check_unfulfilled_image_prompt`** reads `site_specs.site_plan.image_prompts`
  (legacy aspect). Emits work items for `logo`, `hero_home`, `hero_about`,
  etc.
- **New check `check_unfulfilled_imagery_plan`** reads `site_plan_imagery`
  rows. Emits work items with namespaced asset_keys and `kind`.

The `idx_swi_dedup` index catches overlapping `item_key` values, swallowing
duplicates. Sites whose plan was built post-2G have rows in the new table
and the new check processes them. Sites built pre-2G have entries in the
legacy aspect and the old check processes them. Both deliver the same
end-state: assets in git.

**Cutover criteria** (in some weeks' time, not as part of 2G):

- Every site that runs a fresh plan-builder cycle has `site_plan_imagery`
  rows.
- No new sites are using the legacy path.
- A migration backfills any remaining legacy `image_prompts` into
  `site_plan_imagery` rows.

Then the old check can be retired.

---

## Sequencing

In dependency order, sized for individual deploys:

1. **Schema migration** — `site_plan_imagery` table with the constraints
   and indexes above. Pure DDL, no behaviour change. Apply alongside the
   Phase 1 plan-domain migration if it hasn't already landed.
2. **`write_site_plan` extension** — `insertImageryRow` helper + lock
   transfer for imagery rows. Go change. Plan-builder won't emit imagery
   yet, so this is dormant until step 3.
3. **Planner prompt extension** — teach plan-builder to emit the
   `imagery` block. Migration of the agent_definition's prompt template.
   First behaviour change: new plans have imagery rows.
4. **`check_unfulfilled_imagery_plan` discovery check** — new Go file.
   Wired into `design-discovery-agent`. Walks `site_plan_imagery` and
   emits work items.
5. **image-build-handler accepts `kind`** — workflow migration to add a
   third branch for `unfulfilled_imagery_plan`. Spec accepts `kind`,
   `style_hints`, `constraints`. Routes through variant chain.

Each step shippable on its own. Step 1 is harmless additive DDL. Step 2
is dormant until step 3 fires. Steps 3-5 can land in any order — they
won't fully connect until all three are deployed, but partial deployment
is safe (new imagery rows sit unread; new discovery check finds nothing
to process; image-build-handler branch never fires).

The transition strategy means **the old path keeps working** through all
of this. We never have a moment where heroes aren't getting generated.

---

## What 2G doesn't include

Deferred to Phase 2H or later:

- **Generation parameter differentiation by `kind`.** 2G makes `kind`
  available; 2H makes it matter (style_preset, cfg_scale,
  negative_prompt vary by kind).
- **True SVG icon generation.** `kind=icon` in 2G produces a small
  raster image. Real SVG output is a different generator entirely.
- **Infographic generation.** `kind=infographic` is in the enum but
  the generator's behaviour is the same as `illustration` for now.
  Real infographics need chart/data viz generators.
- **Multi-provider routing.** Stays Stability-only.
- **Product imagery.** Comes from the `query.affiliate_products`
  resolver (sibling doc), not from the planner.
- **Auditor awareness of imagery rows.** Visual auditor extension
  (Phase 4) reads these once it ships.

---

## Open questions

Three things worth flagging that the plan as written assumes one way but
could be done differently:

1. **`kind` enum scope.** Five values (`logo, hero, illustration, icon,
   infographic`). Should `product` be in here for parity with affiliate
   products? My read is no — product imagery doesn't go through the
   planner; it comes from the resolver. But if you'd prefer one enum
   covering everything, easy to add.

2. **Should `style_hints` propagate from `site_plan_directives`?** The
   site-scope `imagery_direction` directive carries style guidance. The
   per-imagery `style_hints` lets the planner override or extend it for
   a specific image. At generation time, does the prompt include both
   (cascade), or only the row's hints? My read is cascade: site-scope
   imagery_direction is the default, row hints override. The brief
   renderer pattern already handles this for directives — reusable.

3. **Backfill or just-let-it-age?** Existing sites have `image_prompts`
   in `site_specs.site_plan` and won't have `site_plan_imagery` rows
   until they re-run plan-builder. Option A: backfill migration that
   converts `image_prompts` keys into `site_plan_imagery` rows for
   existing plans. Option B: do nothing; sites get migrated when they
   next replan; old check handles them in the meantime. My read is
   option B — backfilling means parsing `image_prompts` keys to guess
   the kind, which the LLM does better at planning time. Run both
   checks, retire the old one when the legacy aspect is empty.

Worth your view on these before I write any code.

---

## What changes for downstream after 2G ships

For consumers of the imagery plan (image-build-handler, image-generator,
asset-deployer, the auditor whenever it ships):

- A new work item type `unfulfilled_imagery_plan` carrying `kind`,
  namespaced `asset_key`, `prompt`, optional `style_hints`,
  `constraints`.
- Asset_keys may now be hierarchical (`page.about.illustration_team_values`)
  rather than flat. Deploy paths translate naturally.
- The `assets` table picks up rows with these richer asset_keys. No
  schema change to `assets`.
- The legacy `image_prompts` aspect still works. No retirement before
  cutover.

For consumers that look at the spec (auditors, content-quality-auditor):

- No change at the `site_specs` level. Phase 2G touches `site_plan`
  tables only.

For the planner itself:

- One additional block in the JSON output (`imagery`).
- One additional step in the workflow (the `write_site_plan` action
  picks it up automatically once the Go side is in place).

For HITL operators:

- Imagery rows are lockable like directives. Lock a row, the prompt
  text survives plan rebuilds. The asset that's already in `assets`
  doesn't get regenerated (existing asset-locking behaviour
  from Phase 2A).
