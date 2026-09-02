# RUNBOOK — site-design-planner

## 1. Find open work in this mechanism's territory

```sql
SELECT wi.id, wi.item_type, wi.status, wi.item_key, s.domain, wi.created_at, wi.updated_at
FROM site_work_items wi LEFT JOIN sites s ON s.id=wi.site_id
WHERE wi.item_type IN ('needs_composition','needs_new_layout_candidate','needs_theme_review')
  AND wi.status NOT IN ('complete','cancelled','rejected')
ORDER BY wi.created_at;
```
Full spec/result for one row:
```sql
SELECT id, spec, result, error FROM site_work_items WHERE id='<uuid>';
```

## 2. Trace a site's actual composition (the chain that resolves, not `site_specs`)

⚠ `palettes.source_domain` is stamped only on a per-site *fork* — a site riding a
**shared** seed palette resolves through this chain instead and `source_domain`
lookups return false negatives (bug 113's own trap, hit twice in one file).
```sql
SELECT sc.id AS collection_id, t.name AS theme, p.name AS palette, p.origin,
       p.source_domain, l.name AS layout, l.scheme
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
JOIN css_themes t ON t.id = sc.css_theme_id
LEFT JOIN palettes p ON p.id = t.palette_id AND p.is_active
LEFT JOIN layouts l ON l.id = t.layout_id AND l.is_active
WHERE s.domain = '<domain>';
```

## 3. Re-compose ONE site without touching every site (post-113 mechanism)

Two work items, in this order — the composition install alone does not
regenerate `styles.css` (that's the second one's job):
```sql
-- 1. needs_composition, per-request opt-in re-resolve
INSERT INTO site_work_items (site_id, source, item_type, item_key, handler_agent,
    priority, spec)
VALUES ('<site_id>', 'manual', 'needs_composition', 'needs_composition',
    'site-design-planner', 8, '{"allow_reinstall": true}'::jsonb);
-- 2. needs_design — required, not optional; renders the new composition to styles.css
INSERT INTO site_work_items (site_id, source, item_type, item_key, handler_agent,
    priority, spec)
VALUES ('<site_id>', 'manual', 'needs_design', 'needs_design',
    'webdesign-agent', 8, '{}'::jsonb);
```
Then promote `detected` → `triaged` by hand if dispatch doesn't pick it up
(`claim_work_item_action.go` only claims `status IN ('triaged','approved')`,
and promotion out of `detected` for `needs_design`/`needs_composition` is not
automatic — see bug 113 §"Dispatch note").

**Verify at the served artefact, never at `result`/`complete`** (113's own
`f7ceba19` reported `complete` in 2 minutes and changed nothing):
```bash
curl -s https://<domain>/assets/css/styles.css | grep -m1 'color-card-bg'
```
Compare `sites.style_collection_id` and the served `--color-card-bg` before and
after; a value equal to the served `--color-surface` on a palette defining no
`card_bg` is `fillDarkSchemeSpecialisedSlots`' signature (derived, not curated).

## 4. Check whether a site's layout-matcher fallback is still warranted

```sql
SELECT s.domain, ds.aspect, ds.spec_data->'industry_tags', ds.spec_data->'style_direction'
FROM sites s JOIN site_specs ds ON ds.site_id = s.id
WHERE s.domain = '<domain>' AND ds.aspect = 'classification'
ORDER BY ds.created_at DESC LIMIT 1;
```
Empty/absent `industry_tags` reproduces the fallback-to-`brochure-formal` path
(DES-036/037). If tags are present now but weren't at original resolution time
(check `needs_new_layout_candidate`'s `spec->>'site_tags'`), the classifier may
have since been fixed and the site is a candidate for re-resolution (§3) rather
than a genuinely missing layout.

## 5. Verify a chassis build actually carries a composition-pipeline fix

```bash
kubectl -n ai-persona-system exec <pod> -- \
  strings /app/agent-chassis | grep -c '<symbol the fix introduced>'
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
  "SELECT git_commit FROM service_binary_capabilities WHERE service='agent-chassis' AND kind='build' ORDER BY last_seen_at DESC LIMIT 1;"
# then: git merge-base --is-ancestor <your-commit> <that sha>
```
Do **not** trust `agent_definitions.updated_at` as evidence of a recent config
change to this agent — it is degenerate (measured 2026-09-02: 212 of ~215 live
rows share one microsecond, `2026-09-01 20:59:18.364509+00`, from a bulk
unrelated write). Verify a config change by grepping the live text for your
own needle instead.
