# RUNBOOK — scheme reaches components (commands)

Operational companion to `PLAN_scheme_to_components.md` and `running_notes_scheme_to_components.md`. Read-only investigation only so far; no mutating SQL has been run in this thread. The live £29 idea.uk tool is a separate Go binary on a Hetzner VM and is untouched — the chassis build deploys to Backblaze B2 and DNS still points at the VM, so chassis changes are invisible to the live site.

**▶ CURRENT POSITION:** bundle 1 (investigations A/B/G) assembled and read. Scheme signal traced: computed at composition, used to pick the layout, then dropped at both render entry points (CSS render + component render context); recoverable only via `layouts.scheme`. The `--section-*` mechanism (`SectionStyles` + `buildSectionDefaults`) and the `{function}-section` class contract are identified; the components bypass both. **Next: bundle 2 — the C inventory + the D class-name pull — to settle the gating Q4.**

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

## Migration / backfill (gated — not actionable until the design is settled)

Sketch only, to be specified after Q4/E/F/H. Changing a shared component template fans out to every site using it, so the order matters: change templates → identify affected sites → re-render. Reuse the existing trigger — `flag_page_image_rebuild` emits `needs_page` (priority 99) to `page-build-handler`, which re-runs `plan_sections` and re-renders the page — rather than inventing a new one. A known chassis hazard applies if idea.uk gains interactive tools before its rebuild: a content rebuild can de-tool a tool page (`page-content-writer` regenerates from `plan_sections`, which does not know the interactive tool). Keep existing dark sites dark and prove it on a known-good dark reference site (identify one) before any backfill.

## Rollback / safety

No mutating changes yet, so nothing to roll back in this thread. When the fix lands: back up any table before a migration (`CREATE TABLE bak_… AS SELECT * FROM …`), keep the change reversible, and verify a dark reference site still renders dark after each step. idea.uk's chassis build is staging only (DNS → VM); the live £29 binary keeps running throughout.
