-- 318_css_patch_agent_appends_never_round_trips.sql
--
-- bugs_open/198 fix candidate 1 (+3): css-patch-agent's maiden run (2026-08-04)
-- persisted an LLM fragment as the WHOLE stylesheet — css_themes 25,816 → 149 chars,
-- four `CSS fix: <no value>` commits at vm-sites, relojistas.com served unstyled —
-- because the workflow round-tripped the complete document through the model
-- (`patched_css`) and BOTH writers persisted whatever came back, unchecked.
--
-- THE FIX MAKES THE FRAGMENT UNREPRESENTABLE rather than guarding against it:
--
--   1. `plan_css_fix` now returns ONLY `css_added` — the new/overriding rules.
--      The model never carries the stylesheet, so it can never destroy it. This
--      also dissolves the structural conflict the bug file names: max_tokens 8000
--      could never hold a ~26KB "complete stylesheet" even from a compliant model
--      (the bugs_open/012 truncation route to the same clobber).
--   2. `save_css_to_db` APPENDS server-side: `css_content = css_content || ...`.
--      SQL concatenation is monotonic — a shrink is not expressible on this path,
--      which is candidate 2's guard delivered by construction rather than by
--      threshold. A size guard (1..8192 chars) refuses an over-large "patch" (the
--      model echoing a whole document) by matching zero rows; the appended block
--      carries a dated provenance comment naming the finding category.
--   3. NEW step `check_saved` fails LOUD (complete_error) when the guarded UPDATE
--      took no row — without it a refused append would ride the existing
--      "no files to commit → skipped, Success:true" path and read as complete
--      (the a-complete-work-item-is-not-a-repaired-artefact trap).
--   4. `deploy_css` commits `css_saved.css_content` — the DB row AS APPENDED,
--      returned by the UPDATE itself. The repo and the DB can no longer diverge
--      through this workflow; the git write equals the durable row by identity.
--   5. Commit message: the old template named `{{.input_data.spec.category}}`,
--      which GitCommitAction's fixed template context {domain, file_count,
--      filename} can never resolve → `CSS fix: <no value>` on all four incident
--      commits. The message is now composed IN the UPDATE's RETURNING (where
--      params actually resolve) as `commit_msg`, and `commit_message_field`
--      points git_commit at it. That field is a new opt-in on GitCommitAction
--      shipped alongside this migration (same task, registered as DGH-007);
--      UNTIL THAT BINARY ROLLS the running action ignores the unknown key and
--      falls back to the also-updated static template — honest, just less
--      specific. Both orders are safe; no ordering constraint is claimed
--      (owner ruling 2026-07-29 retired condition (1)).
--
-- CONFIG IS LIVE IMMEDIATELY on apply (DB-read workflow). The four contrast fixes
-- from the incident are already preserved in css_themes v6; nothing routes to
-- css-patch-agent between incidents, so applying this closes the door before the
-- next render-audit sweep dispatches anything at it.

-- Probe guard: tell the runner when this is already applied.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'css-patch-agent'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #>> '{workflow,steps,save_css_to_db,config,params,1}' = 'css_fix.result.css_added'
    ) THEN
        RAISE EXCEPTION '198/318: already applied — save_css_to_db already appends css_added';
    END IF;
END $$;

BEGIN;

SELECT snapshot_agent('css-patch-agent',
    'pre-update: bugs_open/198 — stop round-tripping the whole stylesheet through the model');

UPDATE agent_definitions
SET default_config =
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
    jsonb_set(
        default_config,
        '{workflow,steps,plan_css_fix,config,prompt_template}',
        to_jsonb($prompt$You are a CSS expert writing a targeted patch for an existing stylesheet.

## Site
Domain: {{.site_record.domain}}

## Audit Finding
Category: {{.input_data.spec.category}}
Description: {{.input_data.spec.description}}
Suggestion: {{.input_data.spec.suggestion}}
{{if .input_data.spec.affected_component}}Affected component: {{.input_data.spec.affected_component}}{{end}}
{{if .input_data.spec.page_name}}Page: {{.input_data.spec.page_name}}{{end}}

## Current Stylesheet (read-only context — do NOT return it)
```css
{{.current_css.css_content}}
```

## Instructions
Write ONLY the new or overriding CSS rules that implement the fix described in the audit finding. The platform APPENDS your rules to the END of the stylesheet above — you never return the stylesheet itself, and you cannot delete or edit existing rules.
- Rely on the cascade: an appended rule with the same or higher specificity overrides the earlier declaration. Repeat the offending selector exactly as it appears above (or more specifically) so your override wins.
- Make the minimum change needed. Do not redesign or refactor unrelated CSS.
- Only use var(--x) names that are DEFINED in the stylesheet above; otherwise use literal values. An undefined var() with no fallback invalidates the whole declaration at computed-value time.
- To neutralise a wrong declaration, override it with a corrected value (or `unset`).

Return valid JSON:
{
  "css_added": "... ONLY the new/overriding CSS rules to append ...",
  "changes_summary": "Brief description of what was changed",
  "lines_changed": 3
}$prompt$::text),
        false),
        '{workflow,steps,plan_css_fix,description}',
        to_jsonb('LLM writes override rules to append — never the whole stylesheet (bugs_open/198)'::text),
        false),
        '{workflow,steps,save_css_to_db,config,query}',
        to_jsonb($q$UPDATE css_themes SET css_content = css_content || E'\n\n' || '/* css-patch-agent ' || to_char(now(), 'YYYY-MM-DD') || ': ' || replace($3, '*/', '') || ' */' || E'\n' || $2, updated_at = NOW(), version = version + 1 WHERE id = $1::uuid AND length($2) BETWEEN 1 AND 8192 RETURNING id::text, version, css_content, 'CSS fix: ' || replace($3, E'\n', ' ') || ' (theme v' || version::text || ')' AS commit_msg$q$::text),
        false),
        '{workflow,steps,save_css_to_db,config,params}',
        '["current_css.theme_id", "css_fix.result.css_added", "input_data.spec.category"]'::jsonb,
        false),
        '{workflow,steps,save_css_to_db,next_step}',
        to_jsonb('check_saved'::text),
        false),
        '{workflow,steps,save_css_to_db,description}',
        to_jsonb('Append the patch to css_themes server-side (guarded: 1..8192 chars, shrink unrepresentable)'::text),
        false),
        '{workflow,steps,check_saved}',
        '{"action": "conditional_branch", "config": {"condition": "css_saved.count >= 1", "then_step": "deploy_css", "else_step": "complete_error"}, "description": "Refuse to deploy unless the guarded append took a row (bugs_open/198)"}'::jsonb,
        true),
        '{workflow,steps,deploy_css,config,content_field}',
        to_jsonb('css_saved.css_content'::text),
        false),
        '{workflow,steps,deploy_css,config,commit_message}',
        to_jsonb('CSS patch: {{.filename}} ({{.domain}})'::text),
        false),
    updated_at = NOW()
WHERE type = 'css-patch-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  -- Drift guard: only rewrite the shape this file was written against.
  AND default_config #>> '{workflow,steps,save_css_to_db,config,params,1}' = 'css_fix.result.patched_css'
  AND default_config #>> '{workflow,steps,deploy_css,config,content_field}' = 'css_fix.result.patched_css';

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_css,config,commit_message_field}',
        to_jsonb('css_saved.commit_msg'::text),
        true)
WHERE type = 'css-patch-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,deploy_css,config,content_field}' = 'css_saved.css_content';

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config INTO cfg FROM agent_definitions
    WHERE type = 'css-patch-agent'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '198/318: no active css-patch-agent definition found';
    END IF;

    -- IS DISTINCT FROM, not <>: a missing jsonb path is NULL and NULL <> x is NULL.
    IF cfg #>> '{workflow,steps,save_css_to_db,config,params,1}' IS DISTINCT FROM 'css_fix.result.css_added' THEN
        RAISE EXCEPTION '198/318: save_css_to_db params[1] is %, expected css_fix.result.css_added (drift guard refused the rewrite?)',
            COALESCE(cfg #>> '{workflow,steps,save_css_to_db,config,params,1}', '<NULL>');
    END IF;

    IF cfg #>> '{workflow,steps,save_css_to_db,config,query}' NOT LIKE '%css_content = css_content ||%' THEN
        RAISE EXCEPTION '198/318: save_css_to_db query does not append';
    END IF;

    IF cfg #>> '{workflow,steps,save_css_to_db,next_step}' IS DISTINCT FROM 'check_saved' THEN
        RAISE EXCEPTION '198/318: save_css_to_db next_step is %, expected check_saved',
            COALESCE(cfg #>> '{workflow,steps,save_css_to_db,next_step}', '<NULL>');
    END IF;

    IF cfg #> '{workflow,steps,check_saved}' IS NULL THEN
        RAISE EXCEPTION '198/318: check_saved step missing';
    END IF;

    IF cfg #>> '{workflow,steps,deploy_css,config,content_field}' IS DISTINCT FROM 'css_saved.css_content' THEN
        RAISE EXCEPTION '198/318: deploy_css content_field is %, expected css_saved.css_content',
            COALESCE(cfg #>> '{workflow,steps,deploy_css,config,content_field}', '<NULL>');
    END IF;

    IF cfg #>> '{workflow,steps,deploy_css,config,commit_message_field}' IS DISTINCT FROM 'css_saved.commit_msg' THEN
        RAISE EXCEPTION '198/318: deploy_css commit_message_field is %, expected css_saved.commit_msg',
            COALESCE(cfg #>> '{workflow,steps,deploy_css,config,commit_message_field}', '<NULL>');
    END IF;

    IF cfg #>> '{workflow,steps,plan_css_fix,config,prompt_template}' LIKE '%patched_css%' THEN
        RAISE EXCEPTION '198/318: prompt still asks for patched_css';
    END IF;

    IF position('css_fix.result.patched_css' IN cfg::text) > 0 THEN
        RAISE EXCEPTION '198/318: a step still reads css_fix.result.patched_css';
    END IF;
END $$;

COMMIT;
