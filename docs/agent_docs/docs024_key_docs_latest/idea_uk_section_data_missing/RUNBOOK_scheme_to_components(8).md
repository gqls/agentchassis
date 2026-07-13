# RUNBOOK — scheme reaches components (commands)

**The problem, for someone new to this file.** The chassis resolves each site to a light or dark design scheme, and the scheme itself travels correctly: idea.uk resolved light, its layout (`tool-portal-light`) emits light values for the page, surface, header and footer CSS variables, and the renderer only adds white-text (`--section-*`) overrides when a palette is dark, which idea.uk's is not. The failure sits in the component library that supplies the HTML. Components carry their own inline CSS, and much of the library was written assuming a dark site: the active hero and call-to-action paint their bands from palette variables but pair them with literal white text and an unconditional declaration of the dark-text context; the active footer consumes the correct light footer-background variable yet still declares white text over it; 37 of 84 active section components self-declare the dark context and 15 carry literal hex backgrounds. On a light site these assumptions override or bypass the correct variables, so sections render dark — or, in the footer's case, light-with-white-text, unreadable. idea.uk's deployed pages additionally appear to predate the current library (the active header component is already fully variable-driven and would render light today), so part of what is visible is stale rendered HTML. Underneath it all is one design question the library currently answers three conflicting ways: when a section's band is dark by design, who supplies the matching text colours — the component, the renderer, or the layout. This runbook is the operational companion to `PLAN_scheme_to_components.md` and `running_notes_scheme_to_components.md`; everything here is read-only investigation against `clients_db` via kubectl — no code or data changes are made from this file. The live £29 idea.uk tool is a separate Go binary on a Hetzner VM and is untouched — the chassis build deploys to Backblaze B2 and DNS still points at the VM, so chassis changes are invisible to the live site.

**▶ CURRENT POSITION (2026-07-02, post Check 3):** Checks 1–3 run. Scheme→variable pipeline verified correct; all 18 layouts carry the four chrome vars. **Staleness refuted** — idea.uk's pages were built 2026-07-01, two months after the good chrome went active, so what the compile injected is an open question (Check 4a grep on the deployed HTML decides; `site_components` still holds stale dark renders pointing at inactive components). The 37 self-declarers split ~18 hazard-class (declare dark over surface/nothing — white-on-light bugs today, incl. the footer and the five hero-* page variants) vs ~19 band-class (dark band + white text — coherent but block "fully light"). **The user answered the gating question: maximum flexibility — a light scheme must be able to render fully light, and may carry dark hero bands.** That requirement selects the paired-variable direction (Alt C: layout-curated bg+text pairs, palette-overridable per site, components consume pairs and never declare `--section-*`; renderer luminance defaults stay as the base; `is_dark_section` demoted to metadata — 6 of 37 declarers contradict their own flag; Phase 4.5 deferred). **Next: (1) Check 4a deployed-HTML grep + 4b/4c pair-curation queries; (2) draft the fix specification against the paired-variable direction (hazard-class bug fixes, band-class pair conversion, fallback chrome, creator prompt + 003 + fixer re-aim, idea.uk rebuild).**

---

## Conventions (carried, strict)

- DB: `PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"`. Run every `.sql` as a FILE (`$PSQL < file`), never pasted into an interactive psql — pasting mangles `\set`/`\echo`/blank lines and can leave an open transaction.
- k8s namespaces: `-n ai-persona-system` (app pods) and `-n kafka`. Kafka cluster `personae-kafka-cluster`.
- A `0 rows` result is not decisive until the query and the live state have been checked.
- `agent_definitions`: keyed by `type` (+ `version`); `processing_mode` is nested in `default_config`, not a top-level column.
- The chassis deploys via github → GH Actions → Backblaze B2. Components/templates are shared, so any template change affects every site that uses it — back up and keep dark sites dark.

---

## Bundle command for this thread (corrected)

Two fixes over the bundle-1 invocation: the `-doc` paths did not resolve (the `docs024` path was wrong — confirm the real filenames first), and `-step framing` produced signatures only — use `-step debug` (or `implementation`) to get function bodies. Confirm the doc paths before running:

```
ls docs/agent_docs/docs024_key_docs_latest/ | grep -E '^00[1-3]_'
```

```
go run ./cmd/bundle \
  -analysis /tmp/analysis_repo.json -root ~/projects/agentchassis \
  -constitution thin_slice_constitution.md -step debug \
  -task "<one sentence for the bundle>" \
  -scope platform/orchestration/actions/render_css_from_spec_action.go \
  -scope platform/orchestration/actions/render_css_composition_loader.go \
  -scope platform/orchestration/actions/render_css_composition_helpers.go \
  -include platform/orchestration/actions/registry.go \
  -doc docs/agent_docs/docs024_key_docs_latest/003_contracts_and_standards_7_.md \
  -doc docs/agent_docs/docs024_key_docs_latest/002_system_architecture_4_.md \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema-tables content_components,layouts,style_collections,site_components,page_components,css_themes,css_snippets,palettes,typography_sets \
  -runtime-site idea.uk -runtime-page index \
  -out /tmp/bundle_<name>.md
```

`css_snippets`, `palettes` and `typography_sets` are added to `-schema-tables` because the render path reads all three and only their columns-via-queries have been seen so far, not their full `\d`. `buildSectionDefaults` and `isDarkHex` are referenced by `render_css_from_spec_action.go` but defined elsewhere — locate and add their file with `grep -rnE -- "func buildSectionDefaults|func isDarkHex" platform/orchestration/actions/`.

---

## Bundle 2 — investigation C (library inventory)

Run as files. The first is the report's inventory; the rest extend it to the two newly-confirmed surfaces (`css_snippets`) and to the uniqueness question.

```sql
-- C1: every active section + site component against scheme
SELECT function, component_level, section_type, is_dark_section, suitable_site_types
FROM content_components
WHERE is_active = true
ORDER BY component_level, function;
```

```sql
-- C2: scan each active component's inline <style> (html_template) for
--     hardcoded hex, self-set --section-* vars, and legacy var names.
SELECT function, component_level, is_dark_section,
       (html_template ~ '#[0-9a-fA-F]{3,6}')                 AS has_hardcoded_hex,
       (html_template ILIKE '%--section-%')                  AS sets_section_vars,
       (html_template ILIKE '%--accent-color%'
         OR html_template ILIKE '%--primary-color%'
         OR html_template ILIKE '%--color-white%')           AS legacy_var_names
FROM content_components
WHERE is_active = true AND forked_from IS NULL
ORDER BY component_level, function;
```

```sql
-- C3: the css_snippets surface — does component CSS live here too, and
--     does it hardcode dark / set --section-*? (columns seen: name,
--     css_content, applies_to; \d css_snippets in the bundle to confirm.)
SELECT name, applies_to,
       length(css_content)                       AS css_bytes,
       (css_content ILIKE '%--section-%')        AS sets_section_vars,
       (css_content ~ '#[0-9a-fA-F]{3,6}')       AS has_hardcoded_hex
FROM css_snippets
ORDER BY name;
```

```sql
-- C4: is "one active component per function" enforced for sections, or
--     just a data convention? (Schema shows a UNIQUE index only for
--     component_level='tool'.) Any row here = a function with >1 active.
SELECT function, count(*) AS active_count
FROM content_components
WHERE is_active = true AND forked_from IS NULL AND component_level = 'section'
GROUP BY function
HAVING count(*) > 1
ORDER BY function;
```

```sql
-- C5: the active hero / call-to-action / footer templates in full, to read
--     the dark treatment and the classes they emit (feeds D).
SELECT function, is_dark_section, html_template
FROM content_components
WHERE function IN ('hero','call-to-action')
   OR (function LIKE 'footer-%' AND is_active = true)
ORDER BY function, is_active;
```

---

## Bundle 2 — investigation D (class-name contract, gates Q4)

```sql
-- D1: every active layout's css_template — to see what its SectionStyles
--     block emits (rules keyed on .{function}-section, with {{if .IsDark}}
--     dark treatment?) and whether it carries section STRUCTURE or only
--     context. tool-portal-light / tool-portal-dark are the key pair.
SELECT name, scheme, css_template
FROM layouts
WHERE is_active = true
ORDER BY name;
```

The analysis step (by eye, on `tool-portal-light` + the active `hero`/`call-to-action`/footer/header templates): compare the classes the components emit against the renderer's `{function}-section` contract. If the layout's `css_template` emits `.{{.ClassName}}` structure + context from `{{range .SectionStyles}}`, then Q4 option (a) — components adopt `{function}-section` — gives them structure and adaptive context for free. If the layout only emits context (not structure) for those classes, weigh (a) against (b) accordingly.

**Resolved (D1): option (a).** The premise above (that the layout emits `{{range .SectionStyles}}`) turned out false — no active layout consumes `SectionStyles`; each layout hardcodes the `.{function}-section` structural rules and the five surface classes directly, and `buildSectionDefaults` is the live `--section-*` emitter. So a component that adopts `{function}-section` and drops its self-declared dark CSS inherits structure from the layout and adaptive context from an extended `buildSectionDefaults`. See `REFERENCE_styling_render_pipeline.md` §5–§8.

---

## Verification queries (read-only)

```sql
-- The scheme recovery path the fix would add to the render context.
-- Confirms idea.uk resolves to tool-portal-light / scheme=light, and is
-- the exact join sites → style_collection → css_theme → layout.scheme.
SELECT s.domain, l.name AS layout_name, l.scheme
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
JOIN css_themes        t  ON t.id  = sc.css_theme_id
JOIN layouts           l  ON l.id  = t.layout_id
WHERE s.domain = 'idea.uk';
```

```sql
-- idea.uk's current header/footer wiring (both expected dark, is_active=false).
SELECT sc.slot_name, sc.component_id, cc.function, cc.is_dark_section, cc.is_active
FROM site_components sc
LEFT JOIN content_components cc ON cc.id = sc.component_id
WHERE sc.site_id = '1244516d-014d-421c-88c6-090bb1e9552a'
ORDER BY sc.slot_name;
```

```sql
-- Do any layouts declare default header/footer? (Expected: none yet — the F gap.)
SELECT name, scheme, default_header_component_id, default_footer_component_id
FROM layouts
WHERE is_active = true
ORDER BY name;
```

---

## Bundle 3 — investigations E + F (last design unknowns before the fix)

E = the section-contrast model: where a per-section dark/contrast intent is stored, how `is_dark_section` is set, and whether `buildSectionDefaults` should set each dark section's background as well as its `--section-*` text. F = header/footer: untangle the four overlapping default stores (`style_collections.*_component_id` — the one `RenderHeader` actually reads — vs the `site_components` slots vs `sites.default_components` JSONB vs the all-NULL `layouts.default_*_component_id`), confirm whether the composition workflow runs `update_site_defaults`, and make the fallback + chrome scheme-aware. Confirm the doc paths first (`ls docs/agent_docs/docs024_key_docs_latest/ | grep -E '^00[1-3]_'`); `-step debug` for bodies.

```
go run ./cmd/bundle \
  -analysis /tmp/analysis_repo.json -root ~/projects/agentchassis \
  -constitution thin_slice_constitution.md -step debug \
  -task "Settle the section-contrast model (E) and header/footer scheme wiring (F): where a per-section dark/contrast intent is stored and how is_dark_section is set; whether buildSectionDefaults should set each dark section's background as well as its --section-* text; and how the header/footer component is chosen across the overlapping stores (style_collections, site_components, sites.default_components, layouts.default_*) so a light site gets light chrome." \
  -scope platform/orchestration/actions/color_util.go \
  -scope platform/orchestration/actions/render_css_from_spec_action.go:RenderCSSFromSpecAction \
  -scope platform/orchestration/actions/plan_sections_action.go \
  -scope platform/orchestration/actions/write_site_plan_action.go \
  -scope platform/orchestration/actions/install_site_composition_action.go:InstallSiteCompositionAction \
  -scope platform/orchestration/actions/v3_site_actions.go:UpdateSiteDefaultsAction \
  -scope platform/orchestration/actions/v3_site_actions.go:CompilePageSectionsAction \
  -scope platform/orchestration/actions/component_library.go:RenderHeader \
  -scope platform/orchestration/actions/component_library.go:RenderFooter \
  -scope platform/orchestration/actions/component_library.go:RenderFallbackHeader \
  -scope platform/orchestration/actions/component_library.go:RenderFallbackFooter \
  -scope platform/orchestration/actions/component_library.go:InjectHeader \
  -scope platform/orchestration/actions/component_library.go:InjectHead \
  -scope platform/orchestration/actions/component_library.go:GetStyleCollectionForSite \
  -scope platform/orchestration/actions/component_library.go:GetComponentByFunction \
  -include platform/orchestration/actions/registry.go \
  -doc docs/agent_docs/docs024_key_docs_latest/003_contracts_and_standards_7_.md \
  -doc docs/agent_docs/docs024_key_docs_latest/002_system_architecture_4_.md \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema-tables content_components,site_components,style_collections,sites,layouts,css_themes,site_specs \
  -runtime-site idea.uk -runtime-page index \
  -out /tmp/bundle_e_f.md
```

`Inject*` and the `Render*Header/Footer` symbols are all confirmed in `component_library.go` (no separate `inject_header_footer.go`). `plan_sections_action.go` and `write_site_plan_action.go` are large — narrow to symbols if the bundle is too big. Two things can't be named yet, so locate them with greps and add the file the first one finds as a `-scope`:

```bash
# E — where is_dark_section is SET (component-creation / classifier path):
grep -rnE "is_dark_section" platform/ --include=*.go | grep -iE "insert|update|set |:=|bool" | grep -v _test
grep -rln "classify_and_extract\|component-creator\|needs_new_component" platform/ --include=*.go

# F — does the composition workflow run update_site_defaults? (workflows live in SQL/JSON, not Go):
grep -rn "update_site_defaults\|UpdateSiteDefaults" platform/orchestration/ --include=*.go
grep -rn "update_site_defaults" platform/ --include=*.sql
```

### E data (run as files)

```sql
-- E1: is_dark_section distribution by level — how many sections are dark by intent.
SELECT component_level, is_dark_section, count(*)
FROM content_components WHERE is_active = true
GROUP BY component_level, is_dark_section
ORDER BY component_level, is_dark_section;
```

```sql
-- E2a: site_specs is keyed by `aspect` + `data` (NOT spec_type/spec_data — the
--      write_site_spec INPUT field is spec_data, but the column is data). What
--      aspects does idea.uk hold, and how big? (is_current rows only.)
SELECT aspect, length(data::text) AS bytes, created_at
FROM site_specs
WHERE site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND is_current = true
ORDER BY aspect;
-- E2b: per-section intent does NOT live in site_specs or pages.sections (names
--      only — check_unresolved_sections uses jsonb_array_elements_text). The
--      normalised plan is site_plans → site_plan_pages → site_plan_sections.
--      Inspect a page's section rows for any per-section style/contrast/scheme col
--      (\d site_plan_sections first to see if such a column already exists):
SELECT sps.*
FROM site_plans sp
JOIN site_plan_pages    spp ON spp.plan_id = sp.id
JOIN site_plan_sections sps ON sps.plan_id = sp.id AND sps.page_name = spp.name
WHERE sp.site_id = '1244516d-014d-421c-88c6-090bb1e9552a' AND spp.name = 'index';
```

### F data (run as files)

```sql
-- F1: what header/footer will RenderHeader actually pick for idea.uk?
--     (RenderHeader reads style_collections.*_component_id, then falls back.)
SELECT sc.name, sc.css_theme_id,
       sc.header_component_id, sc.header_home_component_id, sc.footer_component_id
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
WHERE s.domain = 'idea.uk';
```

```sql
-- F2: resolve those IDs + the active chrome candidates to function /
--     is_dark_section / is_active, to see light vs dark.
SELECT id, function, component_level, is_dark_section, is_active
FROM content_components
WHERE function IN ('site-header','site-footer','head','header-minimal-tool')
ORDER BY function, is_active;
```

```sql
-- F3: the other default store for idea.uk — is sites.default_components set, to what?
--     (layouts.default_* already covered by the verification query above; expected NULL.)
SELECT default_components FROM sites WHERE domain = 'idea.uk';
```

```sql
-- F4: the active site-header / site-footer / head templates in full — read the dark
--     treatment + classes emitted (the chrome equivalent of C5).
SELECT function, is_active, is_dark_section, html_template
FROM content_components
WHERE function IN ('site-header','site-footer','head')
ORDER BY function, is_active;
```

---

## Migration / backfill (gated — not actionable until the design is settled)

Sketch only, to be specified after Q4/E/F/H. Changing a shared component template fans out to every site using it, so the order matters: change templates → identify affected sites → re-render. Reuse the existing trigger — `flag_page_image_rebuild` emits `needs_page` (priority 99) to `page-build-handler`, which re-runs `plan_sections` and re-renders the page — rather than inventing a new one. A known chassis hazard applies if idea.uk gains interactive tools before its rebuild: a content rebuild can de-tool a tool page (`page-content-writer` regenerates from `plan_sections`, which does not know the interactive tool). Keep existing dark sites dark and prove it on a known-good dark reference site (identify one) before any backfill.

## Rollback / safety

No mutating changes yet, so nothing to roll back in this thread. When the fix lands: back up any table before a migration (`CREATE TABLE bak_… AS SELECT * FROM …`), keep the change reversible, and verify a dark reference site still renders dark after each step. idea.uk's chassis build is staging only (DNS → VM); the live £29 binary keeps running throughout.

## CHECK 1 — do the layouts expose scheme-flipping chrome vars? (run on cluster; paste output)
Mechanism confirmed from code/docs: SECTION contrast flips via buildSectionDefaults on merged-palette background/surface luminance (idea.uk's #EFE7D6 / #E8DFCC are light → no dark overrides → var-reading sections render light already). CHROME uses --color-header-bg/--color-footer-bg/--color-cta-bg, which are SPECIALISED palette slots (theme-owned) with a layout-supplied fallback (`palette "key" "fallback"`). Open data facts:

```sql
-- 1a: does idea.uk's palette define specialised chrome slots, or only the 8 core?
--     (confirm colour column: \d palettes — 'colours' vs 'colors')
SELECT pal.name AS palette_name, pal.colours
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
JOIN css_themes ct        ON ct.id = sc.css_theme_id
JOIN palettes  pal        ON pal.id = ct.palette_id
WHERE s.domain = 'idea.uk';
-- look for header_bg / header_text / footer_bg / footer_text / cta_bg

-- 1b: what do the light vs dark layouts set for chrome + CTA vars?
SELECT name, scheme,
       substring(css_template from '--color-header-bg[^;]*;') AS header_bg_rule,
       substring(css_template from '--color-footer-bg[^;]*;') AS footer_bg_rule,
       substring(css_template from '--color-cta-bg[^;]*;')    AS cta_bg_rule
FROM layouts
WHERE name IN ('tool-portal-light','tool-portal-dark')
ORDER BY name;
-- light shows light (or light fallback) + dark shows dark → de-hardcoded chrome consumes var(--color-header-bg).
-- light shows null/dark → chrome should consume var(--color-surface)/var(--color-background) instead.
-- (null substring → dump full css_template for both, grep .site-header/.site-footer.)
```
DECIDES: whether the chrome fix is "consume --color-header-bg" or "consume surface/background" — and, with the lane choice (re-aim fixers toward Phase 4.5 vs make fixers scheme-aware), unblocks the fix spec.

## CHECK 2 — size the de-hardcode (run on cluster; paste output)
Context: header_component_id is a dead column (never populated) → RenderHeader uses GetComponentByFunction("site-header") → else dark fallback (hardcodes ctx.PrimaryColor + white text). Chrome fix = de-hardcode the one active site-header/site-footer + the fallback to consume var(--color-header-bg/-text) etc. Need the templates + blast radius:

```sql
-- 2a: the specific component templates (active AND idea.uk's own, incl. inactive site-header/footer).
SELECT function, is_active, is_dark_section, created_from,
       length(html_template) AS len, html_template
FROM content_components
WHERE function IN ('hero','features','call-to-action','site-header','site-footer')
  AND forked_from IS NULL
ORDER BY function, is_active DESC;

-- 2b: does the light layout set the chrome TEXT vars (not just bg from 1b)?
SELECT name, scheme,
       substring(css_template from '--color-header-text[^;]*;') AS header_text_rule,
       substring(css_template from '--color-footer-text[^;]*;') AS footer_text_rule
FROM layouts WHERE name IN ('tool-portal-light','tool-portal-dark') ORDER BY name;

-- 2c: blast radius — active section/header/footer components hardcoding colour or declaring --section-*.
SELECT component_level,
       count(*) FILTER (WHERE html_template ~* 'background(-color)?:\s*#[0-9a-fA-F]{3,8}') AS hardcoded_bg,
       count(*) FILTER (WHERE html_template ~* '--section-[a-z]+\s*:')                     AS declares_section_vars,
       count(*) AS total
FROM content_components
WHERE is_active = true AND forked_from IS NULL
  AND component_level IN ('section','header','footer')
GROUP BY component_level ORDER BY component_level;
```
DECIDES: 2a → the hardcoding patterns to fix + whether an active de-hardcoded site-header/footer exists or must be created/activated; 2b → whether the fallback needs the nested --color-text fallback; 2c → the count that settles S-min vs S-full (Phase 4.5).

## CHECK 2 RESULTS (run 2026-07-02) — the templates + blast radius
- **Active `site-header` (generated) is already correct**: `background-color: var(--color-header-bg, var(--color-background)); color: var(--color-header-text, var(--color-text))` throughout → renders light on idea.uk. No component fix needed.
- **Active `site-footer` (generated, is_dark_section=t) is HALF-migrated**: `background: var(--color-footer-bg, #1a1a2e)` (flips → `#f1ede4` light on idea.uk) but self-declares `--section-text: rgba(255,255,255,.9)` etc. → white text on cream, unreadable; newsletter box uses white-alpha surface/border → invisible. A live library bug, worse than "renders dark".
- **Active `hero` + `call-to-action`: no hex backgrounds** — they consume palette vars (`--color-primary/secondary/accent`) but ASSUME dark: literal `#fff` text + unconditional dark `--section-*` block. On idea.uk primary = `#1A1816` → near-black bands with white text: readable, deliberate-looking dark bands on a light site. Bug vs design choice = the gating question.
- `features` is the model citizen ("Layout only - colors inherited") — already correct. Inactive manual headers include a light variant (`header-minimal-light`, hardcodes `#ffffff`) and dark/gradient ones — historic, all inactive.
- **2b**: chrome TEXT vars flip too (light: header/footer text `#1a1a1a`; dark: `#e0e0e0`/rgba). Full chrome var set is scheme-correct. Only `--color-cta-bg` missing on the light layout; no `--color-cta-text` pair exists anywhere.
- **2c blast radius** (active, forked_from IS NULL): section 84 total — 15 hex bg, 37 self-declare `--section-*`; header 4 total — 0 hex, 2 self-declare (NB: 4 active header-level components but only one active `site-header` function → other header functions exist, list in 3d); footer 1 — the half-migrated one.
- Corollary: the generated footer proves the **component-creator prompt currently emits half-migrated components** (consumes chrome bg var, self-declares dark text) — whatever architecture is chosen must be encoded in the creator prompt + 003 checklist or drift continues.

## CHECK 3 — staleness, provenance, self-declarer split (run on cluster; paste output)
Purpose: do NOT conclude why deployed idea.uk is dark until these are read. The current active header would render light, so the deployed dark chrome is probably stale build output — verify, don't assume.

```sql
-- 3a: staleness — when did the good chrome components appear vs idea.uk's last build?
SELECT function, is_active, created_from, created_at, updated_at
FROM content_components
WHERE function IN ('site-header','site-footer') AND forked_from IS NULL
ORDER BY function, is_active DESC, created_at;
SELECT domain, last_built_at, last_deployed_at FROM sites WHERE domain='idea.uk';
SELECT p.name, p.build_status, p.updated_at
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain='idea.uk' ORDER BY p.updated_at DESC LIMIT 10;

-- 3b: provenance — what chrome is actually stored for idea.uk (and, separately, fetch one
--     deployed page from B2/git and grep its header class:
--     .site-header-section = current generated; .site-header--dark = old manual; bare
--     .site-header + stacked-nav/search = RenderFallbackHeader).
SELECT sc.slot_name, sc.component_id, cc.function, cc.is_active,
       left(sc.rendered_html, 160) AS rendered_head
FROM site_components sc
LEFT JOIN content_components cc ON cc.id = sc.component_id
WHERE sc.site_id = (SELECT id FROM sites WHERE domain='idea.uk');

-- 3c: split the 37 self-declarers — self-painted bands (declaration load-bearing today)
--     vs renderer/layout-painted surfaces (invisible-text risk on light sites):
SELECT function, is_dark_section,
       (html_template ~* 'background(-color)?:\s*var\(--color-(primary|secondary|accent)') AS paints_band_from_palette,
       (html_template ~* 'background(-color)?:\s*var\(--color-(surface|background|header-bg|footer-bg)') AS paints_surface_var,
       (html_template ~* 'background(-color)?:\s*#[0-9a-fA-F]{3,8}') AS paints_hex
FROM content_components
WHERE is_active=true AND forked_from IS NULL AND component_level='section'
  AND html_template ~* '--section-[a-z-]+\s*:'
ORDER BY function;

-- 3d: the 4 active header-level components — RenderHeader only looks up 'site-header';
--     what are the others and are they reachable?
SELECT function, name, is_active, is_dark_section
FROM content_components
WHERE component_level IN ('header','footer') AND is_active=true AND forked_from IS NULL;

-- 3e: do ALL active layouts set the four chrome vars (or only tool-portal-*)?
SELECT name, scheme,
       (css_template LIKE '%--color-header-bg%')  AS has_header_bg,
       (css_template LIKE '%--color-header-text%') AS has_header_text,
       (css_template LIKE '%--color-footer-bg%')  AS has_footer_bg,
       (css_template LIKE '%--color-footer-text%') AS has_footer_text
FROM layouts WHERE is_active=true ORDER BY scheme, name;

-- 3f: the 15 hex-background section components (candidates for fix_hardcoded_colors / review):
SELECT function, substring(html_template from 'background(-color)?:\s*#[0-9a-fA-F]{3,8}') AS first_hex
FROM content_components
WHERE is_active=true AND forked_from IS NULL AND component_level='section'
  AND html_template ~* 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'
ORDER BY function;
```
DECIDES: 3a/3b → whether deployed dark chrome is stale output (rebuild suffices for chrome) or something still selects dark at render; 3c → how many of the 37 are load-bearing band declarations vs hazardous surface declarations (sizes every architecture); 3d → whether non-`site-header` header functions are dead weight; 3e → whether chrome-var consumption is safe across all layouts or needs fallback care; 3f → the hex list for the existing fixer.

## CHECK 3 RESULTS (run 2026-07-02)
- **3a — STALENESS REFUTED.** The good generated chrome went active 2026-05-06 (header 07:30, footer 10:56). idea.uk's pages were built and deployed 2026-07-01 12:52–12:58 — nearly two months LATER. So "deployed dark chrome is stale build output" is wrong as stated: the 07-01 compile ran while the variable-driven `site-header` was the active `site-header`. Either the deployed pages already carry the light header (and the dark-chrome observation predates the 07-01 build), or the compile path injected something other than `RenderHeader`'s function lookup — decided by Check 4 (grep the deployed HTML). `sites.last_built_at` is NULL (only `last_deployed_at` is written — vestigial, matches the 2026-05-26 handoff note on `build_status`).
- **3b — site_components holds stale DARK chrome pointing at INACTIVE components**: header slot → `site-header--gradient` (manual, is_active=f, paints `{{.primary_color}}→{{.secondary_color}}` gradient = dark); footer slot → `footer-4-column` (manual, is_active=f, `background: {{.primary_color}}` + white text); head slot → inactive `head`. `render_site_components` has not re-run since those components were deactivated. Whether these rows reach deployed pages = Check 4.
- **3c — the 37 self-declarers split into two classes** (flags are per-TEMPLATE — any element can match — so per-element review is needed at fix time):
  - **Hazard class (~18): declare dark `--section-*` while painting surface/background vars or nothing.** 13 paint `--color-surface/background/header-bg/footer-bg` somewhere (archetype-grid, archetype-taster-quiz, brief-explanation, contact-block, game-master-explanation, gripper-payload-calculator, header-docs, lobby-grid, platform-comparison, provocation-feed, site-footer, site-head, tool-agent-complexity-estimator); 5 paint NOTHING matching any pattern (hero-about/-case-studies/-contact/-services/-use-cases — likely image+overlay inline styles; if the image is absent they are transparent-with-white-text). On light sites these render white-on-light. Bugs today regardless of any architecture choice.
  - **Band class (~19): paint from `--color-primary/secondary/accent` + declare white text** (call-to-action, hero, social-proof, testimonials, system-stats, product-hero, case-studies-grid, gauntlet-cta/-interface, tool-cta, several tool-* calculators…). Coherent today; these are what block "fully light".
  - Flag hygiene evidence: 6 declarers have `is_dark_section=f` (archetype-grid, archetype-taster-quiz, game-master-explanation, lobby-grid, provocation-card, tool-cta) — the LLM-authored flag disagrees with the template's own styling → do not key styling on this flag.
- **3d — the other 5 active header/footer functions** (header-with-cart-or-nav, header-minimal-tool, header-with-categories, header-with-search, footer-with-disclaimer, all `*_pre_037`) are unreachable on the compile path (`RenderHeader` looks up only `site-header`; `header_component_id` is dead). Park; future variety, not this fix.
- **3e — all 18 active layouts set all four chrome vars** → chrome-var consumption is safe library-wide. NOTE: only 3 layouts have `scheme` set (tool-portal-dark=dark; tool-portal-light + soft-editorial=light); the 15 seeded layouts have scheme EMPTY — scheme metadata is sparsely curated, and each layout's chrome-var FALLBACK VALUES are uncurated/unknown for those 16 (pair-curation query in Check 4).
- **3f — my SQL display bug (owned):** `substring(... from 'background(-color)?:...')` returns the FIRST CAPTURE GROUP in Postgres, so `first_hex` came back empty/'-color'. The 15-row LIST is right (about-content, departments-grid, filtered-result-grid, leadership-team, product-card-with-cta, product-grid, + tool-* calculators); the displayed hex is not. Re-run without capture groups if the values are wanted.

## CHECK 4 — deployed-HTML provenance + layout pair curation (run when convenient)
```bash
# 4a: which header did the 2026-07-01 build actually deploy? From the site's git repo clone
#     (the repo GitHub Actions deploys to B2), or fetch index.html from B2, then:
grep -o 'site-header[a-z-]*\|site-footer[a-z-]*\|mobile-menu-toggle' index.html | sort | uniq -c
# site-header-section  → current generated header (light on idea.uk) — compile path used RenderHeader as mapped
# site-header--gradient → the stale site_components render reached the page — path mapping needs correcting
# site-header + mobile-menu-toggle, no logo img → RenderFallbackHeader fired
# Also check the footer: site-footer-section → generated (the white-on-cream bug would be LIVE on light bg).
```
```sql
-- 4b: chrome/pair fallback curation across all 18 layouts (no capture groups this time):
SELECT name, scheme,
       substring(css_template from '--color-header-bg:[^;]+')  AS header_bg,
       substring(css_template from '--color-footer-bg:[^;]+')  AS footer_bg,
       substring(css_template from '--color-cta-bg:[^;]+')     AS cta_bg
FROM layouts WHERE is_active=true ORDER BY scheme NULLS LAST, name;
-- 4c: does any layout already carry pair/on-colour text vars?
SELECT name,
       (css_template LIKE '%--color-cta-text%')      AS has_cta_text,
       (css_template LIKE '%--color-primary-text%')  AS has_primary_text
FROM layouts WHERE is_active=true ORDER BY name;
```
DECIDES: 4a → what the 07-01 build injected (closes the provenance question and says whether the footer bug is live); 4b/4c → which layouts need pair values added vs already have them (sizes the layout-side work).

## CHECK 4 RESULTS (run 2026-07-02) — provenance answered
- **4a — the deployed pages are RERENDER output carrying stale stored renders, not build output.** Deployed index.html chrome byte-matches the stale `site_components.rendered_html` rows from 3b: header = `site-header--gradient` with `#1A1816 → #4A4540` baked in (the inactive manual gradient template rendered with idea.uk's palette), footer = `footer-4-column` with `background:#1A1816` + white text, head = the old static base-CSS head. The git commits say `Rerender: index.html` etc. (~2 weeks ago) — the rerender handler reassembles stored `page_components` and injects `site_components.rendered_html`, exactly as 016 described. The build path's `InjectHeader → RenderHeader` mapping stands, but that path never ran for these pages.
- **Sections are stale old-template renders too**: the deployed hero consumes legacy `var(--accent-color, #0f3460)` — the variable naming of the now-INACTIVE hero — proving the section HTML predates the active hero (which uses `--color-accent`). idea.uk has been living on reassemblies of early renders while the library advanced. A full page-build-handler rebuild is required for sections; `needs_rerender` would re-fossilise them.
- **Timestamp discrepancy (open, non-blocking):** `pages.updated_at`/`build_status=deployed` say 2026-07-01 12:52–12:58 and `sites.last_deployed_at` 12:49, but the git HTML commits are ~2 weeks old and styles.css ("Update stylesheet via webdesign-agent") ~1 week old. Likely a no-diff rerender or deploy-only touch on 07-01 that advanced DB timestamps without new commits. Identify what ran on 07-01 when convenient.
- **Confirmed live bugs on the deployed light site:** dark gradient header, dark footer (both stale chrome), dark CTA band (band-class component, `--color-primary` = `#1A1816`), hero dark-over-image (acceptable per design answer). The tool-list section is fully variable-driven and renders correctly — generated components CAN be model citizens.
- **4b — every layout defines all three chrome/cta fallbacks except tool-portal-light (missing `--color-cta-bg`).** The fallback values are deliberate per-layout design: several light-page layouts pair a DARK footer band (affiliate-hub `#18181b`, brochure-formal `#1a365d`, comparison-aggregator/industry-hub `#0f172a`, ecommerce `#111827`, magazine-grid `#1a1a1a`, social-lobby `#18181b`) — the "light site, dark band by choice" model is already curated in the library. cta_bg values are mostly accent/brand bands (orange/blue/violet/red), not merely dark.
- **4c — the paired-variable convention is ALREADY the library standard**: 18/18 layouts define `--color-primary-text`, 17/18 define `--color-cta-text` — the single gap is tool-portal-light (no cta pair). Further: the deployed page's baked base `<head>` CSS already consumes `--color-cta-bg`/`--color-cta-text` (`.section--cta`) and defines hero pair vars (`--color-hero-title`, `--color-hero-subtitle`) — the pair naming precedent exists; REUSE these names rather than inventing new ones. The chosen direction is therefore COMPLETION of existing architecture, not a restructure: one layout to patch, components to bring into line.

**▶ POSITION AFTER CHECK 4:** All unknowns closed except three small open items (what ran 07-01; InjectHead's selection logic; the planned-but-unbuilt pages linked in nav). Fix specification written: `SPEC_scheme_to_components.md`. **W1 prepared** (§W1 EXECUTION below: backup command + w1_00–w1_02 + rollback) — awaiting execution and paste-back (incl. the 18-layout sweep for the contrast check).

## W1 EXECUTION — add the CTA pair to tool-portal-light (first mutating change of the thread)
Files (in outputs): `w1_00_before.sql`, `w1_01_add_cta_pair.sql`, `w1_02_verify.sql`, `w1_rollback.sql`.
Run order (SQL as files, per convention):
```bash
PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"

# 0.5 — backup the full current template BEFORE the update (streams to a local file):
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -Atc \
  "SELECT css_template FROM layouts WHERE name='tool-portal-light' AND is_active=true" \
  > tool_portal_light_css_template_$(date +%F).bak
wc -c tool_portal_light_css_template_$(date +%F).bak   # sanity: non-trivial size

$PSQL < w1_00_before.sql        # \d layouts + anchor present + pair absent + the 18-layout sweep
$PSQL < w1_01_add_cta_pair.sql  # expect UPDATE 1 + both inserted_* columns shown
$PSQL < w1_02_verify.sql        # expect occurrences 1 / 1 + the insertion region
```
Notes: the update is anchored on the `--color-footer-text` line via a `\1` backreference (no whitespace guessing), guarded `NOT LIKE '%--color-cta-bg%'` (idempotent), first-occurrence-only. Values chosen: neutral light band `#e9e2d3` + ink `#1a1a1a` (contrast ≈ 13.5), mirroring tool-portal-dark's neutral `#1e1e1e` CTA band; accent alternative (`#9b4020` + `#faf8f3`, ≈ 6.1) is a two-value swap in w1_01. `updated_at` deliberately untouched (add if `\d` shows the column and it matters). **W1 is inert until W3**: nothing consumes `--color-cta-bg` yet (the CTA component still paints `--color-primary`; the legacy head CSS's `.section--cta` rule consumes it but no component carries that class) — zero visual change expected anywhere. styles.css re-render deferred to W6 step 1 for the same reason. Rollback: `w1_rollback.sql` (value-agnostic) or restore from the .bak file.
