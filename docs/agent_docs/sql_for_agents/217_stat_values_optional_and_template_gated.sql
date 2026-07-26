-- 217_stat_values_optional_and_template_gated.sql
-- bugs_open/043 (SLUG: generated_page_copy_invents_quantitative_claims) fix
-- candidate 1, config-only half — and the fix for bugs_open/073.
-- DB-only; effective immediately, no image roll.
--
-- ── WHY ────────────────────────────────────────────────────────────────────
-- Migration 201 (live 2026-07-24) rewrote page-content-writer's rule 14: a
-- required numeric field is NOT permission to invent a figure, and the honest
-- answer when no given figure fits is an empty string. The writer obeyed.
--
-- But `missingRequiredLLMFields` (platform/orchestration/actions/json_envelope.go:204)
-- counts an empty string as a MISSING required field (isEmptyContentValue,
-- :231), and RenderComponentAction refuses the section on that basis
-- (v3_site_actions.go:1719-1732). So the honest answer became a hard page-build
-- failure. That is bugs_open/073, and it is not theoretical: measured
-- 2026-07-26, ai-agent-orchestration.com/index.html cannot be rebuilt by the
-- writer at all (dies at process_sections_loop_iter_4_render_section on
-- case-studies-grid), and is therefore STILL SERVING the pre-201 case-study
-- metrics that migration 201 existed to stop. The anti-fabrication fix froze
-- the fabrications in place.
--
-- 201 fixed the RULE. This fixes the three things that contradicted it:
--   (1) the schema demanded a value           -> required:false + on_missing:skip_field
--   (2) the template rendered it unguarded    -> {{if .field}}...{{end}}
--   (3) the field DESCRIPTION demanded one,
--       with example shapes to copy           -> guidance rewritten, seeds stripped
-- 043 recorded the writer copying system-stats' own "e.g. '2.4M'" exemplar to
-- produce a fabricated "2,400+". content-block-about went further and asked for
-- a "memorable, credibility-building number"; gauntlet-cta for a "compelling"
-- one. Those are instructions to persuade, not to report.
--
-- And (4): component-creator is told to author image fields as optional+gated
-- but has no equivalent rule for numeric fields, so the next generated
-- component re-seeds the whole class. Without that patch this migration is a
-- one-time cleanup rather than a fix. It is included below.
--
-- ── SAFETY: WHY NO CURRENTLY-LIVE PAGE CAN CHANGE ──────────────────────────
-- Four independent legs.
--
-- (1) For a source:"llm" field, `required` never reaches anything that emits
--     markup. plan_sections_action.go:1463 is `if source == "llm" { ...append
--     spec...; continue }` — it returns before handleMissingField is reachable.
--     Exhaustive grep gives four readers of the per-field flag
--     (plan_sections_action.go:1348, json_envelope.go:217,
--     discovery_checks/check_required_fields_missing.go:187,
--     datahelpers/component_schema_fields.go:116). Their only effects are the
--     prompt's ", required" suffix, the render gate, and a post-deploy work
--     item. None renders HTML.
--
-- (2) The render gate has NEVER let an empty required value through, so nothing
--     published depends on one. Verified: zero empty-or-absent required llm
--     fields across every live page_components row of these ten components.
--     Relaxing the gate therefore cannot alter anything already deployed — it
--     only changes what a FUTURE writer output is allowed to contain.
--
-- (3) {{if .x}}...{{end}} emits the inner text byte-for-byte when x is present,
--     and every live stat value on every placement is non-empty. The gates are
--     written glued to the opening and closing tags on the same physical lines,
--     with NO trim markers ({{- / -}}) — a trim marker would strip the existing
--     indentation and change live bytes.
--
-- (4) Schema and template land in ONE transaction. Relaxing `required` without
--     the template gate is precisely how you get a live <strong></strong>
--     (073's own warning about its candidate 1).
--
-- ── WHAT IS DELIBERATELY NOT TOUCHED ───────────────────────────────────────
-- * Fields are ENUMERATED, never pattern-matched. A `%stat%` predicate also
--   catches product-details.availability_status, provocations-archive-list.
--   empty_state_label and tool-loot-probability-calculator.empty_state_message
--   — unrelated components whose truncation tripwires would be blown open.
-- * tool-guide-intro.audience_value / .skill_level_value hold prose
--   ("Beginner", "Marketers & SEO professionals"), not figures. Left required.
-- * stat-band is already the reference implementation ({{range .stats}}{{if
--   .value}}) and its guidance already says "if you do not have a verified
--   number, do not add the stat". Its `stats` array is its ONLY required field,
--   so relaxing it would leave the component with no tripwire at all — that
--   needs a decision (min_items + skip_section?), not a blind flip. Recorded as
--   a residual in bugs_open/043.
-- * gauntlet-interface.reward_value — 043 records this component as inert
--   residue owned by the gauntlet_dead_cta thread.
-- * system-stats-leo / system-stats-leopardess are inactive forks (0
--   placements) still on the pre-211 template. Out of scope; noted in 043.
--
-- ── THE TRUNCATION TRIPWIRE STAYS ARMED ────────────────────────────────────
-- The gate exists because of bug 026: a truncated LLM response silently blanked
-- nine live article bodies. Every one of the ten components below KEEPS its
-- required source:"llm" prose fields — section_headline, section_intro,
-- body_text, card titles/excerpts/client names, spec category labels, column
-- headers, archetype name/description. A truncated or unparseable response
-- still blanks those and still hard-fails the render. Nothing here is
-- fail-open, and the post-condition block asserts it (>= 3 required prose
-- fields per component).
--
-- ROLLBACK: bak_043_stat_components_20260726 holds the pre-change rows, and a
-- component_versions row per component carries the pre-edit html_template +
-- input_schema. Restore from either.
--
-- Idempotent: re-running is a genuine no-op. The template block uses a
-- three-way guard (gate present -> NOTICE and skip; ungated needle -> replace;
-- neither -> EXCEPTION, meaning another session edited the template and the
-- needle must be re-derived). Schema writes set false to false and rewrite the
-- same guidance. Backup is CREATE TABLE IF NOT EXISTS, so a replay preserves
-- the ORIGINAL snapshot. Ledger insert is ON CONFLICT DO NOTHING.

\set ON_ERROR_STOP on

BEGIN;

-- Snapshots must open the transaction (standing rule, migration 201).
SELECT snapshot_agent('page-content-writer', '217_stat_values_optional_and_template_gated.sql: pre-update');
SELECT snapshot_agent('component-creator',   '217_stat_values_optional_and_template_gated.sql: pre-update');

-- ============================================================================
-- 0. BACKUP — the house idiom for a library-wide component edit (cf. 211, 185).
--    IF NOT EXISTS so a replay keeps the original pre-change snapshot.
-- ============================================================================

CREATE TABLE IF NOT EXISTS bak_043_stat_components_20260726 AS
SELECT * FROM content_components
WHERE name IN ('system-stats','case-studies-grid','content-block-about','gauntlet-cta',
               'archetype-result-card','product-hero_pre_037',
               'bayesian-ranking-hero-tool_pre_037','product-specs',
               'platform-comparison','tool-guide-intro');

-- Machine-readable trail the regeneration paths consult (cf. 170).
INSERT INTO component_versions (component_id, version_number, html_template, input_schema,
                                change_description, changed_by, change_source)
SELECT c.id,
       COALESCE((SELECT MAX(cv.version_number) FROM component_versions cv WHERE cv.component_id = c.id), 0) + 1,
       c.html_template, c.input_schema,
       '217: pre-edit snapshot before making stat value fields optional and gating their markup (bugs_open/043 + 073)',
       '217_stat_values_optional_and_template_gated', 'migration'
FROM content_components c
WHERE c.name IN ('system-stats','case-studies-grid','content-block-about','gauntlet-cta',
                 'archetype-result-card','product-hero_pre_037',
                 'bayesian-ranking-hero-tool_pre_037','product-specs',
                 'platform-comparison','tool-guide-intro')
  AND NOT EXISTS (
        SELECT 1 FROM component_versions cv
        WHERE cv.component_id = c.id
          AND cv.changed_by = '217_stat_values_optional_and_template_gated');

-- ============================================================================
-- 1. SCHEMA — 80 fields become optional, and their guidance stops demanding a
--    number. Path-scoped jsonb_set with a merge patch:
--      * path-scoped, NOT jsonb_object_agg over {fields} (211 part C's shape),
--        so a concurrent session's edit to a SIBLING field is not lost;
--      * `|| jsonb_build_object(...)` merges, preserving type/source/fallback
--        and any key a future migration adds;
--      * `IS NOT NULL` guard is MANDATORY, not defensive: NULL || jsonb is
--        NULL, jsonb_set(x, path, NULL) returns NULL, and the row's ENTIRE
--        input_schema would be nulled silently. create_missing is false too.
--      * `false` is a real jsonb boolean, never the string "false" — the render
--        gate does def["required"].(bool) while the discovery check's
--        fieldFlagTrue also accepts "true", so a string would leave the two
--        readers disagreeing.
-- ============================================================================

DO $schema$
DECLARE
    r           record;
    guidance    text;
    n_rows      int;
    n_applied   int := 0;

    -- The bounded rule that replaces every "e.g. '2.4M'" exemplar. Names the
    -- four given-context sources by the exact labels the writer prompt uses.
    g_value CONSTANT text :=
        'State a number here ONLY if that exact number appears in the context you have been given — Verified Facts, Research Findings, the Admin Content Brief, or Existing Content. '
        'If no given figure fits this slot, return an empty string: the stat is then not rendered at all, which is the correct and honest outcome. '
        'Never estimate, round up, extrapolate, or carry a number across from a different meaning. Do not invent a figure to fill the field.';

    g_label CONSTANT text :=
        'Short label naming what the figure counts (1-4 words). Leave empty if the matching value is empty — the stat is then not rendered.';

    g_desc CONSTANT text :=
        'One sentence saying what the figure measures and over what period. Describe only what the given context supports; leave empty if the matching value is empty.';

    -- Table rows: the identity cell decides whether the row renders at all.
    g_rowkey CONSTANT text :=
        'The name of this row. Leave empty to omit the whole row rather than inventing an entry — an omitted row is correct when you have nothing true to put in it.';

    g_cell CONSTANT text :=
        'The value for this cell. State it ONLY if it appears in the context you have been given (Verified Facts, Research Findings, Admin Content Brief, or Existing Content). Never invent, estimate or extrapolate a figure. Leave empty if you have none.';
BEGIN
    FOR r IN
        SELECT * FROM (VALUES
            -- system-stats: value/label/description are one visual unit.
            ('system-stats','stat1_value','value_suffixed'), ('system-stats','stat1_label','label'), ('system-stats','stat1_description','desc'),
            ('system-stats','stat2_value','value_suffixed'), ('system-stats','stat2_label','label'), ('system-stats','stat2_description','desc'),
            ('system-stats','stat3_value','value_suffixed'), ('system-stats','stat3_label','label'), ('system-stats','stat3_description','desc'),
            ('system-stats','stat4_value','value_suffixed'), ('system-stats','stat4_label','label'), ('system-stats','stat4_description','desc'),
            -- case-studies-grid: the per-card outcome metric (073's failure site).
            ('case-studies-grid','card1_stat_value','value'), ('case-studies-grid','card1_stat_label','label'),
            ('case-studies-grid','card2_stat_value','value'), ('case-studies-grid','card2_stat_label','label'),
            ('case-studies-grid','card3_stat_value','value'), ('case-studies-grid','card3_stat_label','label'),
            ('case-studies-grid','card4_stat_value','value'), ('case-studies-grid','card4_stat_label','label'),
            ('case-studies-grid','card5_stat_value','value'), ('case-studies-grid','card5_stat_label','label'),
            -- content-block-about: labels are already optional here.
            ('content-block-about','stat_1_value','value'),
            ('content-block-about','stat_2_value','value'),
            ('content-block-about','stat_3_value','value'),
            -- gauntlet-cta
            ('gauntlet-cta','stat_1_value','value'),
            ('gauntlet-cta','stat_2_value','value'),
            ('gauntlet-cta','stat_3_value','value'),
            -- archetype-result-card
            ('archetype-result-card','stat_1_value','value'), ('archetype-result-card','stat_1_label','label'),
            ('archetype-result-card','stat_2_value','value'), ('archetype-result-card','stat_2_label','label'),
            ('archetype-result-card','stat_3_value','value'), ('archetype-result-card','stat_3_label','label'),
            -- product-hero
            ('product-hero_pre_037','stat_one_value','value'),
            ('product-hero_pre_037','stat_two_value','value'),
            ('product-hero_pre_037','stat_three_value','value'),
            -- bayesian-ranking-hero-tool
            ('bayesian-ranking-hero-tool_pre_037','stat_one_value','value'),   ('bayesian-ranking-hero-tool_pre_037','stat_one_label','label'),
            ('bayesian-ranking-hero-tool_pre_037','stat_two_value','value'),   ('bayesian-ranking-hero-tool_pre_037','stat_two_label','label'),
            ('bayesian-ranking-hero-tool_pre_037','stat_three_value','value'), ('bayesian-ranking-hero-tool_pre_037','stat_three_label','label'),
            -- product-specs: name is the row identity, value is the cell.
            ('product-specs','spec_1_name','rowkey'), ('product-specs','spec_1_value','cell'),
            ('product-specs','spec_2_name','rowkey'), ('product-specs','spec_2_value','cell'),
            ('product-specs','spec_3_name','rowkey'), ('product-specs','spec_3_value','cell'),
            ('product-specs','spec_4_name','rowkey'), ('product-specs','spec_4_value','cell'),
            ('product-specs','spec_5_name','rowkey'), ('product-specs','spec_5_value','cell'),
            ('product-specs','spec_6_name','rowkey'), ('product-specs','spec_6_value','cell'),
            ('product-specs','spec_7_name','rowkey'), ('product-specs','spec_7_value','cell'),
            ('product-specs','spec_8_name','rowkey'), ('product-specs','spec_8_value','cell'),
            -- platform-comparison: feature is the row identity, three cells each.
            ('platform-comparison','row1_feature','rowkey'), ('platform-comparison','row1_platform1_value','cell'), ('platform-comparison','row1_platform2_value','cell'), ('platform-comparison','row1_spark_value','cell'),
            ('platform-comparison','row2_feature','rowkey'), ('platform-comparison','row2_platform1_value','cell'), ('platform-comparison','row2_platform2_value','cell'), ('platform-comparison','row2_spark_value','cell'),
            ('platform-comparison','row3_feature','rowkey'), ('platform-comparison','row3_platform1_value','cell'), ('platform-comparison','row3_platform2_value','cell'), ('platform-comparison','row3_spark_value','cell'),
            ('platform-comparison','row4_feature','rowkey'), ('platform-comparison','row4_platform1_value','cell'), ('platform-comparison','row4_platform2_value','cell'), ('platform-comparison','row4_spark_value','cell'),
            ('platform-comparison','row5_feature','rowkey'), ('platform-comparison','row5_platform1_value','cell'), ('platform-comparison','row5_platform2_value','cell'), ('platform-comparison','row5_spark_value','cell'),
            -- tool-guide-intro: only the number-shaped one.
            ('tool-guide-intro','read_time_value','value')
        ) AS t(cname, fname, role)
    LOOP
        guidance := CASE r.role
            WHEN 'value_suffixed' THEN g_value || ' Do not include units or symbols here — put those in '
                                                || replace(r.fname, '_value', '_suffix') || '.'
            WHEN 'value'  THEN g_value
            WHEN 'label'  THEN g_label
            WHEN 'desc'   THEN g_desc
            WHEN 'rowkey' THEN g_rowkey
            WHEN 'cell'   THEN g_cell
        END;

        UPDATE content_components c
           SET input_schema = jsonb_set(
                   c.input_schema,
                   ARRAY['fields', r.fname],
                   (c.input_schema #> ARRAY['fields', r.fname])
                     || jsonb_build_object('required',     false,
                                           'on_missing',   'skip_field',
                                           'llm_guidance', guidance),
                   false),
               updated_at = now()
         WHERE c.name = r.cname
           AND c.input_schema #> ARRAY['fields', r.fname] IS NOT NULL;

        GET DIAGNOSTICS n_rows = ROW_COUNT;
        IF n_rows = 0 THEN
            RAISE EXCEPTION '217: component % has no field % — schema drifted, re-derive before applying', r.cname, r.fname;
        END IF;
        n_applied := n_applied + n_rows;
    END LOOP;

    IF n_applied <> 80 THEN
        RAISE EXCEPTION '217: expected 80 field updates, applied % — inspect', n_applied;
    END IF;
    RAISE NOTICE '217: % stat fields are now optional with the invention seed removed', n_applied;
END $schema$;

-- ============================================================================
-- 2. TEMPLATES — gate the markup so an absent stat renders NOTHING.
--    Three-way guard per needle makes the whole block re-runnable.
--    Rules applied:
--      * stat card  -> gate the card on its own value field;
--      * bordered container -> ALSO gate the wrapper on `or` of its values,
--        else a fully-empty block leaves rules over emptiness (.about-stats
--        has border-top AND border-bottom + 1.5rem padding; .gauntlet-stats
--        and .arc-stats have border-top + padding; .stats-grid has a 3rem
--        margin-bottom). .hero-stats and .brht-trust-row are bare flex with no
--        border, so they get no container gate;
--      * TABLE -> gate the <tr> on the row's identity field, NEVER a <td>.
--        Hiding one cell under a fixed <thead> shifts every later column left.
-- ============================================================================

DO $tpl$
DECLARE
    r          record;
    tmpl       text;
    n_changed  int := 0;
    n_skipped  int := 0;
BEGIN
    FOR r IN
        SELECT cname, needle, repl FROM (

            -- ---- system-stats: four stat cards -------------------------------
            SELECT 'system-stats' AS cname,
                   format(E'      <article class="stat-card">\n        <div class="stat-value">{{.stat%1$s_value}}<span class="stat-suffix">{{.stat%1$s_suffix}}</span></div>\n        <div class="stat-label">{{.stat%1$s_label}}</div>\n        <p class="stat-description">{{.stat%1$s_description}}</p>\n      </article>', i) AS needle,
                   format(E'      {{if .stat%1$s_value}}<article class="stat-card">\n        <div class="stat-value">{{.stat%1$s_value}}<span class="stat-suffix">{{.stat%1$s_suffix}}</span></div>\n        <div class="stat-label">{{.stat%1$s_label}}</div>\n        <p class="stat-description">{{.stat%1$s_description}}</p>\n      </article>{{end}}', i) AS repl,
                   i AS ord
            FROM generate_series(1,4) AS i

            -- ---- system-stats: the grid wrapper -----------------------------
            UNION ALL SELECT 'system-stats',
                   E'    <div class="stats-grid">',
                   E'    {{if or .stat1_value .stat2_value .stat3_value .stat4_value}}<div class="stats-grid">', 10
            UNION ALL SELECT 'system-stats',
                   E'    </div>\n    <footer class="stats-footer">',
                   E'    </div>{{end}}\n    <footer class="stats-footer">', 11

            -- ---- case-studies-grid: five card metrics ------------------------
            UNION ALL
            SELECT 'case-studies-grid',
                   format(E'            <span class="csg-card-stat"><strong>{{.card%1$s_stat_value}}</strong> {{.card%1$s_stat_label}}</span>', i),
                   format(E'            {{if .card%1$s_stat_value}}<span class="csg-card-stat"><strong>{{.card%1$s_stat_value}}</strong> {{.card%1$s_stat_label}}</span>{{end}}', i),
                   20 + i
            FROM generate_series(1,5) AS i

            -- ---- content-block-about: three stat items + bordered wrapper ----
            UNION ALL
            SELECT 'content-block-about',
                   format(E'        <div class="stat-item">\n          <span class="stat-value">{{.stat_%1$s_value}}</span>\n          <span class="stat-label">{{.stat_%1$s_label}}</span>\n        </div>', i),
                   format(E'        {{if .stat_%1$s_value}}<div class="stat-item">\n          <span class="stat-value">{{.stat_%1$s_value}}</span>\n          <span class="stat-label">{{.stat_%1$s_label}}</span>\n        </div>{{end}}', i),
                   30 + i
            FROM generate_series(1,3) AS i
            UNION ALL SELECT 'content-block-about',
                   E'      <div class="about-stats">',
                   E'      {{if or .stat_1_value .stat_2_value .stat_3_value}}<div class="about-stats">', 40
            UNION ALL SELECT 'content-block-about',
                   E'      </div>\n      {{if .cta_url}}<a href="{{.cta_url}}"',
                   E'      </div>{{end}}\n      {{if .cta_url}}<a href="{{.cta_url}}"', 41

            -- ---- gauntlet-cta: three stats + bordered wrapper ----------------
            UNION ALL
            SELECT 'gauntlet-cta',
                   format(E'        <div class="gauntlet-stat">\n          <span class="gauntlet-stat-value">{{.stat_%1$s_value}}</span>\n          <span class="gauntlet-stat-label">{{.stat_%1$s_label}}</span>\n        </div>', i),
                   format(E'        {{if .stat_%1$s_value}}<div class="gauntlet-stat">\n          <span class="gauntlet-stat-value">{{.stat_%1$s_value}}</span>\n          <span class="gauntlet-stat-label">{{.stat_%1$s_label}}</span>\n        </div>{{end}}', i),
                   50 + i
            FROM generate_series(1,3) AS i
            UNION ALL SELECT 'gauntlet-cta',
                   E'      <div class="gauntlet-stats">',
                   E'      {{if or .stat_1_value .stat_2_value .stat_3_value}}<div class="gauntlet-stats">', 60
            UNION ALL SELECT 'gauntlet-cta',
                   E'      </div>\n    </div>\n    <div class="gauntlet-panel">',
                   E'      </div>{{end}}\n    </div>\n    <div class="gauntlet-panel">', 61

            -- ---- archetype-result-card: three stats + bordered wrapper -------
            UNION ALL
            SELECT 'archetype-result-card',
                   format(E'        <div class="arc-stat" role="listitem">\n          <span class="arc-stat-value">{{.stat_%1$s_value}}</span>\n          <span class="arc-stat-label">{{.stat_%1$s_label}}</span>\n        </div>', i),
                   format(E'        {{if .stat_%1$s_value}}<div class="arc-stat" role="listitem">\n          <span class="arc-stat-value">{{.stat_%1$s_value}}</span>\n          <span class="arc-stat-label">{{.stat_%1$s_label}}</span>\n        </div>{{end}}', i),
                   70 + i
            FROM generate_series(1,3) AS i
            UNION ALL SELECT 'archetype-result-card',
                   E'      <div class="arc-stats" role="list" aria-label="{{.stats_aria_label}}">',
                   E'      {{if or .stat_1_value .stat_2_value .stat_3_value}}<div class="arc-stats" role="list" aria-label="{{.stats_aria_label}}">', 80
            UNION ALL SELECT 'archetype-result-card',
                   E'      </div>\n    </article>\n    <div class="arc-actions">',
                   E'      </div>{{end}}\n    </article>\n    <div class="arc-actions">', 81

            -- ---- product-hero: three stats, no container gate needed ---------
            UNION ALL
            SELECT 'product-hero_pre_037',
                   format(E'        <div class="hero-stat">\n          <span class="hero-stat-value">{{.stat_%1$s_value}}</span>\n          <span class="hero-stat-label">{{.stat_%1$s_label}}</span>\n        </div>', w),
                   format(E'        {{if .stat_%1$s_value}}<div class="hero-stat">\n          <span class="hero-stat-value">{{.stat_%1$s_value}}</span>\n          <span class="hero-stat-label">{{.stat_%1$s_label}}</span>\n        </div>{{end}}', w),
                   90
            FROM unnest(ARRAY['one','two','three']) AS w

            -- ---- bayesian-ranking-hero-tool: minified, one line --------------
            UNION ALL
            SELECT 'bayesian-ranking-hero-tool_pre_037',
                   format('<div class="brht-trust-item"><span class="brht-trust-value">{{.stat_%1$s_value}}</span><span class="brht-trust-label">{{.stat_%1$s_label}}</span></div>', w),
                   format('{{if .stat_%1$s_value}}<div class="brht-trust-item"><span class="brht-trust-value">{{.stat_%1$s_value}}</span><span class="brht-trust-label">{{.stat_%1$s_label}}</span></div>{{end}}', w),
                   100
            FROM unnest(ARRAY['one','two','three']) AS w

            -- ---- product-specs: gate the <tr>, never the <td> ----------------
            UNION ALL
            SELECT 'product-specs',
                   format(E'            <tr>\n              <th scope="row">{{.spec_%1$s_name}}</th>\n              <td>{{.spec_%1$s_value}}</td>\n            </tr>', i),
                   format(E'            {{if .spec_%1$s_name}}<tr>\n              <th scope="row">{{.spec_%1$s_name}}</th>\n              <td>{{.spec_%1$s_value}}</td>\n            </tr>{{end}}', i),
                   110 + i
            FROM generate_series(1,8) AS i

            -- ---- platform-comparison: gate the <tr> on the feature name ------
            UNION ALL
            SELECT 'platform-comparison',
                   format(E'          <tr>\n            <td class="pc-td-feature">{{.row%1$s_feature}}</td>\n            <td>{{.row%1$s_platform1_value}}</td>\n            <td>{{.row%1$s_platform2_value}}</td>\n            <td class="pc-td-spark">{{.row%1$s_spark_value}}</td>\n          </tr>', i),
                   format(E'          {{if .row%1$s_feature}}<tr>\n            <td class="pc-td-feature">{{.row%1$s_feature}}</td>\n            <td>{{.row%1$s_platform1_value}}</td>\n            <td>{{.row%1$s_platform2_value}}</td>\n            <td class="pc-td-spark">{{.row%1$s_spark_value}}</td>\n          </tr>{{end}}', i),
                   130 + i
            FROM generate_series(1,5) AS i

            -- ---- tool-guide-intro: gate the whole meta item ------------------
            --      Gating only the value span would leave the "Read time"
            --      label and its icon pointing at nothing.
            UNION ALL SELECT 'tool-guide-intro',
                   E'        <div class="tgi-meta-item">\n          <svg class="tgi-meta-icon" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>\n          <span class="tgi-meta-label">{{.read_time_label}}</span>\n          <span>{{.read_time_value}}</span>\n        </div>',
                   E'        {{if .read_time_value}}<div class="tgi-meta-item">\n          <svg class="tgi-meta-icon" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>\n          <span class="tgi-meta-label">{{.read_time_label}}</span>\n          <span>{{.read_time_value}}</span>\n        </div>{{end}}', 150
        ) AS needles
        ORDER BY ord, needle
    LOOP
        SELECT html_template INTO tmpl FROM content_components WHERE name = r.cname;
        IF tmpl IS NULL THEN
            RAISE EXCEPTION '217: component % not found', r.cname;
        END IF;

        IF position(r.repl in tmpl) > 0 THEN
            -- Already gated: idempotent replay.
            n_skipped := n_skipped + 1;
        ELSIF position(r.needle in tmpl) > 0 THEN
            UPDATE content_components
               SET html_template = replace(html_template, r.needle, r.repl),
                   updated_at = now()
             WHERE name = r.cname;
            n_changed := n_changed + 1;
        ELSE
            RAISE EXCEPTION '217: % — neither the gated nor the ungated form of a needle was found; the template has drifted (another session edited it). Re-derive the needle from the live row before applying. Needle head: %',
                r.cname, left(r.needle, 90);
        END IF;
    END LOOP;

    -- 46 = system-stats 4 cards + 2 container
    --    + case-studies-grid 5
    --    + content-block-about 3 + 2 container
    --    + gauntlet-cta 3 + 2 container
    --    + archetype-result-card 3 + 2 container
    --    + product-hero 3 + bayesian 3
    --    + product-specs 8 + platform-comparison 5
    --    + tool-guide-intro 1
    IF n_changed + n_skipped <> 46 THEN
        RAISE EXCEPTION '217: expected 46 template needles, saw % changed + % skipped', n_changed, n_skipped;
    END IF;
    RAISE NOTICE '217: templates — % gated, % already gated', n_changed, n_skipped;
END $tpl$;

-- ============================================================================
-- 3. component-creator — the NUMERIC FIELDS RULE.
--    Without this, 217 is a one-time cleanup: the next generated component
--    ships required numeric fields with "e.g. '2.4M'" guidance and the class
--    re-seeds. Modelled verbatim on the IMAGE FIELDS RULE already in its
--    prompt, which established exactly this optional+gated pattern.
-- ============================================================================

DO $cc$
DECLARE
    n int;
    anchor CONSTANT text := 'IMAGE FIELDS RULE:';
    newrule CONSTANT text :=
E'NUMERIC FIELDS RULE: any field holding a statistic, count, total, price, rating or coverage figure (names like stat_1_value, cardN_stat_value, metric_value) MUST be "required": false with "on_missing": "skip_field", and the template MUST gate the whole stat''s markup with a Go-template conditional on that field ({{if .stat_1_value}}...{{end}}) — including the wrapper element when it carries a border or its own spacing. A figure is only ever available if a site''s verified evidence supplies it, so the section must render cleanly without one. In llm_guidance for such a field, NEVER give example figures ("e.g. ''2.4M''") and NEVER ask for a "compelling" or "credibility-building" number: a required numeric field with example shapes and no data is an instruction to fabricate, which is bugs_open/043. Ask for the figure only if the given context supplies it, and say that an empty string is the correct answer otherwise.

IMAGE FIELDS RULE:';
BEGIN
    SELECT count(*) INTO n
    FROM agent_definitions
    WHERE type = 'component-creator' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
      AND default_config->>'prompt_template' LIKE '%NUMERIC FIELDS RULE:%';
    IF n > 0 THEN
        RAISE NOTICE '217: component-creator already carries the NUMERIC FIELDS RULE — skipping';
    ELSE
        SELECT count(*) INTO n
        FROM agent_definitions
        WHERE type = 'component-creator' AND is_active
          AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
          AND default_config->>'prompt_template' LIKE '%' || anchor || '%';
        IF n <> 1 THEN
            RAISE EXCEPTION '217: expected exactly 1 live component-creator carrying the IMAGE FIELDS RULE anchor, found %', n;
        END IF;

        UPDATE agent_definitions
           SET default_config = jsonb_set(
                   default_config, '{prompt_template}',
                   to_jsonb(replace(default_config->>'prompt_template', anchor, newrule))),
               updated_at = now()
         WHERE type = 'component-creator' AND is_active
           AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
        RAISE NOTICE '217: component-creator gained the NUMERIC FIELDS RULE';
    END IF;
END $cc$;

-- ============================================================================
-- 4. page-content-writer — make "optional" read as permission, not absence.
--    The "What To Write" block prints ", required" for required fields and
--    NOTHING for optional ones. Migration 201's post-mortem: "a prohibition
--    without a legal alternative loses to a structural demand". Silence is not
--    a legal alternative; say it out loud.
--    llmFieldSpec.Required is `json:"required,omitempty"` so the key is ABSENT
--    for optional fields, which {{if}} already handles — this is the else arm
--    of a conditional the template has always had.
-- ============================================================================

DO $pw$
DECLARE
    n int;
    old_line CONSTANT text := '{{if .required}}, required{{end}}';
    new_line CONSTANT text := '{{if .required}}, required{{else}}, optional — return "" if you have nothing true to put here{{end}}';
    path CONSTANT text[] := ARRAY['workflow','steps','process_sections_loop','config','sub_workflow','steps','generate_content','config','prompt_template'];
    tmpl text;
BEGIN
    SELECT default_config #>> path INTO tmpl
    FROM agent_definitions
    WHERE type = 'page-content-writer' AND is_active
      AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    IF tmpl IS NULL THEN
        RAISE EXCEPTION '217: page-content-writer section prompt_template not found at the expected path';
    END IF;

    IF position(new_line in tmpl) > 0 THEN
        RAISE NOTICE '217: page-content-writer already carries the optional-field marker — skipping';
    ELSE
        n := (length(tmpl) - length(replace(tmpl, old_line, ''))) / length(old_line);
        IF n <> 1 THEN
            RAISE EXCEPTION '217: expected the required-marker line exactly once in the writer prompt, found % — prompt drifted', n;
        END IF;

        UPDATE agent_definitions
           SET default_config = jsonb_set(default_config, path,
                   to_jsonb(replace(tmpl, old_line, new_line))),
               updated_at = now()
         WHERE type = 'page-content-writer' AND is_active
           AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
        RAISE NOTICE '217: page-content-writer now names the optional case explicitly';
    END IF;
END $pw$;

-- ============================================================================
-- 5. POST-CONDITIONS — fail the whole transaction if any leg did not land.
-- ============================================================================

DO $post$
DECLARE
    n int;
BEGIN
    -- (a) input_schema survived intact on every touched component.
    SELECT count(*) INTO n FROM content_components
    WHERE name IN ('system-stats','case-studies-grid','content-block-about','gauntlet-cta',
                   'archetype-result-card','product-hero_pre_037',
                   'bayesian-ranking-hero-tool_pre_037','product-specs',
                   'platform-comparison','tool-guide-intro')
      AND (input_schema IS NULL OR jsonb_typeof(input_schema->'fields') <> 'object');
    IF n <> 0 THEN
        RAISE EXCEPTION '217: % component(s) lost their input_schema — the NULL||jsonb hazard fired; ROLLING BACK', n;
    END IF;

    -- (b) no targeted field is still required, and `required` is a real boolean.
    --     The pattern below is deliberately EXACT, not a loose `%value%`. A
    --     looser one first written here flagged archetype-result-card's
    --     `archetype_description` — a prose field this migration deliberately
    --     leaves required — which is the same over-broad-predicate mistake the
    --     header warns about for the WHERE clause. Anchored alternatives only.
    SELECT count(*) INTO n
    FROM content_components c, LATERAL jsonb_each(c.input_schema->'fields') e(k,v)
    WHERE c.name IN ('system-stats','case-studies-grid','content-block-about','gauntlet-cta',
                     'archetype-result-card','product-hero_pre_037',
                     'bayesian-ranking-hero-tool_pre_037','product-specs','platform-comparison')
      AND (v->>'source') = 'llm'
      AND (v->>'required')::bool
      AND e.k ~ ('^(stat[0-9]+_(value|label|description)'
               || '|card[0-9]+_stat_(value|label)'
               || '|stat_[0-9]+_(value|label)'
               || '|stat_(one|two|three)_(value|label)'
               || '|spec_[0-9]+_(name|value)'
               || '|row[0-9]+_(feature|platform1_value|platform2_value|spark_value))$');
    IF n <> 0 THEN
        RAISE EXCEPTION '217: % stat field(s) are still required', n;
    END IF;

    SELECT count(*) INTO n
    FROM content_components c, LATERAL jsonb_each(c.input_schema->'fields') e(k,v)
    WHERE c.name IN ('system-stats','case-studies-grid','content-block-about','gauntlet-cta',
                     'archetype-result-card','product-hero_pre_037',
                     'bayesian-ranking-hero-tool_pre_037','product-specs',
                     'platform-comparison','tool-guide-intro')
      AND v ? 'required' AND jsonb_typeof(v->'required') <> 'boolean';
    IF n <> 0 THEN
        RAISE EXCEPTION '217: % field(s) carry a non-boolean `required` — the two Go readers would disagree', n;
    END IF;

    -- (c) the invention seed is gone from every touched field.
    SELECT count(*) INTO n
    FROM content_components c, LATERAL jsonb_each(c.input_schema->'fields') e(k,v)
    WHERE c.name IN ('system-stats','case-studies-grid','content-block-about','gauntlet-cta',
                     'archetype-result-card','product-hero_pre_037',
                     'bayesian-ranking-hero-tool_pre_037','product-specs','platform-comparison')
      AND e.k ~ ('^(stat[0-9]+_value'
               || '|card[0-9]+_stat_value'
               || '|stat_[0-9]+_value'
               || '|stat_(one|two|three)_value'
               || '|spec_[0-9]+_(name|value)'
               || '|row[0-9]+_(feature|platform1_value|platform2_value|spark_value))$')
      AND v->>'llm_guidance' ~ 'e\.g\.[^.]*[0-9]';
    IF n <> 0 THEN
        RAISE EXCEPTION '217: % field(s) still carry a numeric example shape in their guidance', n;
    END IF;

    -- (d) THE BUG-026 TRIPWIRE. Every touched component must keep required
    --     source:llm prose, or a truncated writer response could silently blank
    --     a section again.
    SELECT count(*) INTO n FROM (
        SELECT c.name, count(*) AS req_prose
        FROM content_components c, LATERAL jsonb_each(c.input_schema->'fields') e(k,v)
        WHERE c.name IN ('system-stats','case-studies-grid','content-block-about','gauntlet-cta',
                         'archetype-result-card','product-hero_pre_037',
                         'bayesian-ranking-hero-tool_pre_037','product-specs',
                         'platform-comparison','tool-guide-intro')
          AND (v->>'source') = 'llm' AND (v->>'required')::bool
        GROUP BY c.name HAVING count(*) >= 3
    ) t;
    IF n <> 10 THEN
        RAISE EXCEPTION '217: only % of 10 components still hold >= 3 required llm prose fields — the bug-026 truncation tripwire has been weakened', n;
    END IF;

    -- (e) no ungated stat markup survives.
    SELECT count(*) INTO n FROM content_components
    WHERE (name = 'system-stats'          AND html_template LIKE '%' || E'\n      <article class="stat-card">' || '%')
       OR (name = 'case-studies-grid'     AND html_template LIKE '%' || E'\n            <span class="csg-card-stat">' || '%')
       OR (name = 'content-block-about'   AND html_template LIKE '%' || E'\n        <div class="stat-item">' || '%')
       OR (name = 'gauntlet-cta'          AND html_template LIKE '%' || E'\n        <div class="gauntlet-stat">' || '%')
       OR (name = 'archetype-result-card' AND html_template LIKE '%' || E'\n        <div class="arc-stat" role="listitem">' || '%')
       OR (name = 'product-hero_pre_037'  AND html_template LIKE '%' || E'\n        <div class="hero-stat">' || '%')
       OR (name = 'bayesian-ranking-hero-tool_pre_037' AND html_template LIKE '%><div class="brht-trust-item">%')
       OR (name = 'product-specs'         AND html_template LIKE '%' || E'\n            <tr>\n              <th scope="row">' || '%')
       OR (name = 'platform-comparison'   AND html_template LIKE '%' || E'\n          <tr>\n            <td class="pc-td-feature">' || '%');
    IF n <> 0 THEN
        RAISE EXCEPTION '217: % component(s) still render an ungated stat block', n;
    END IF;

    RAISE NOTICE '217: all post-conditions passed';
END $post$;

-- ============================================================================
-- 6. Ledger. ON CONFLICT is required — without it a replay dies here (211's
--    omission). The runner writes its own row afterwards, also DO NOTHING, so
--    this descriptive note survives.
-- ============================================================================

INSERT INTO schema_migrations (filename, notes)
VALUES ('217_stat_values_optional_and_template_gated.sql',
        'bugs_open/043 candidate 1 + bugs_open/073: 80 stat fields across 10 components become optional with on_missing:skip_field, their markup is gated with {{if}} (tables gate the <tr>, bordered wrappers gate the container), the "e.g. 2.4M" invention seeds are stripped from their guidance, component-creator gains a NUMERIC FIELDS RULE so the class cannot re-seed, and the writer prompt names the optional case out loud. Config-only, live immediately. Backup: bak_043_stat_components_20260726 + component_versions.')
ON CONFLICT (filename) DO NOTHING;

COMMIT;
