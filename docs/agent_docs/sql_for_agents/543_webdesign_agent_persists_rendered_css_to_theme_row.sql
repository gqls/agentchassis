-- 543_webdesign_agent_persists_rendered_css_to_theme_row.sql
--
-- bugs_open/198, the cause one level down. 542 stops css-patch-agent deploying an
-- unsafe base. THIS migration stops the base being unsafe in the first place.
--
-- ONE ARTEFACT, TWO WRITERS, NEITHER READING THE OTHER. `assets/css/styles.css`
-- is produced by:
--
--   webdesign-agent   generate_css (render_css_from_spec) → deploy_css (git_commit
--                     of generated_css.result). Reads the FK'd palette / layout /
--                     typography rows, the design_spec and css_snippets. It does
--                     NOT read `css_themes.css_content`, and it does not write it.
--   css-patch-agent   appends to `css_themes.css_content`, then git-commits THE
--                     WHOLE ROW over the same file.
--
-- So whichever agent ran last owns the file entirely, and the theme row is not
-- "stale" — it was never in that path. `install_site_composition_action.go:342-370`
-- inserts it as `''` on purpose ("the renderer reads composition via FKs"), and
-- nothing has ever filled it. The concept register (DES-005) states the birth
-- contract as "empty css_content — webdesign-agent fills it at render". THAT FILL
-- DOES NOT EXIST. This migration is what makes the register's own sentence true.
--
-- WHY IT IS THE DURABLE HALF. Every per-site repair done for this bug — nine rows
-- backfilled fleet-wide on 2026-08-21, plus relojistas, idea.uk, noted, dartsonline,
-- remortgagecalculator, loanzy — has an expiry equal to the OTHER writer's next
-- run: a webdesign-agent design run regenerates the file from the spec and the row
-- goes stale again, silently, with no symptom until the next patch dispatch. That
-- agent runs roughly weekly per site (dartsonline: seven times since 2026-07-06).
-- Backfills are a one-time repair; this is the mechanism that keeps the row true.
--
-- WHAT IT DOES. One `query_database` step, `persist_css_to_theme`, between
-- generate_css and deploy_css: write the just-rendered CSS into the site's linked
-- theme row, byte-for-byte. Row-first ordering is deliberate — if the deploy then
-- fails, the row still matches what the NEXT css-patch deploy will ship, so the
-- two converge in the right direction either way.
--
-- FOUR GUARDS ON THE WRITE, each closing a way this step could itself cause harm:
--
--   octet_length($2) >= 4096    Never persist a fragment. A short render is the
--                               same defect from the other end, and 4096 BYTES is
--                               the same census-derived floor 542 refuses at, so
--                               the two halves cannot disagree about what a
--                               stylesheet is. (A nil param already fails the step
--                               loudly; an empty string does not, so it is floored
--                               here in SQL.)
--   ct.origin <> 'seed'         Never overwrite a LIBRARY theme. Seed rows are
--                               shared design assets, not per-site artefacts.
--   exactly one linking site    Never push one site's CSS onto another through a
--                               shared row. `professional-dark` is one row linked
--                               by finetuning.uk AND gaswholesalers.com, which
--                               serve different files; persisting either would
--                               corrupt the other. Those sites keep whatever their
--                               row holds and are refused at the patch door by 542
--                               until a human splits them.
--   IS DISTINCT FROM $2         No version churn when the render is unchanged.
--
-- A 0-row match is therefore a NORMAL outcome (shared row, seed row, unchanged
-- content, or a run with no site), not an error — `css_persisted.count` records
-- which it was, and the workflow proceeds to deploy regardless.
--
-- FAIL-OPEN, ON PURPOSE, AND NAMED. error_step is `deploy_css`, not a failure
-- terminal. The realistic step error here is an unresolvable site id on a non-site
-- run. Failing an entire design run because a bookkeeping write did not land would
-- trade a live capability for a hygiene property, and the backstop already exists:
-- 542 refuses to patch an unsafe base, so a row that fails to persist causes a
-- REFUSAL later, never a clobber. That is the right direction to fail in.
--
-- Provenance goes in `description`, not into `css_content`. The row must stay
-- byte-identical to the deployed file — that identity is what makes "deploy the
-- whole row" safe, and it is what a future DB-vs-file drift check can compare by
-- md5 rather than by heuristic.
--
-- CONFIG IS LIVE IMMEDIATELY ON APPLY. No Go change, no roll: `query_database`,
-- its params and `output_format: object` are all long-standing live machinery.
-- Apply 542 first (or in either order — they are independent; 542 guards the
-- consumer, 543 fixes the producer).

-- Probe guard: tell the runner when this is already applied.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'webdesign-agent'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #> '{workflow,steps}' ? 'persist_css_to_theme'
    ) THEN
        RAISE EXCEPTION '198/543: already applied — persist_css_to_theme already exists';
    END IF;
END $$;

BEGIN;

SELECT snapshot_agent('webdesign-agent',
  '543_webdesign_agent_persists_rendered_css_to_theme_row: pre-update');

-- ── DRIFT GUARD ────────────────────────────────────────────────────────────────
DO $$
DECLARE
    v_steps jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps}'
      INTO v_steps
      FROM agent_definitions
     WHERE type = 'webdesign-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF v_steps IS NULL THEN
        RAISE EXCEPTION '198/543: no live webdesign-agent row found';
    END IF;

    -- the producer step, asserted by action AND output field: the new step reads
    -- `generated_css.result`, so both must be what this migration expects.
    IF v_steps #>> '{generate_css,action}' <> 'render_css_from_spec' THEN
        RAISE EXCEPTION '198/543 drift: generate_css.action is %, expected render_css_from_spec',
            v_steps #>> '{generate_css,action}';
    END IF;
    IF v_steps #>> '{generate_css,output_field}' <> 'generated_css' THEN
        RAISE EXCEPTION '198/543 drift: generate_css.output_field is %, expected generated_css',
            v_steps #>> '{generate_css,output_field}';
    END IF;

    -- the edge being spliced
    IF v_steps #>> '{generate_css,next_step}' <> 'deploy_css' THEN
        RAISE EXCEPTION '198/543 drift: generate_css.next_step is %, expected deploy_css',
            v_steps #>> '{generate_css,next_step}';
    END IF;

    -- the consumer of the same value, asserted so that "the row equals the file"
    -- is actually true after this lands: deploy_css must still commit the very
    -- string this step persists.
    IF v_steps #>> '{deploy_css,action}' <> 'git_commit' THEN
        RAISE EXCEPTION '198/543 drift: deploy_css.action is %, expected git_commit',
            v_steps #>> '{deploy_css,action}';
    END IF;
    IF v_steps #>> '{deploy_css,config,content_field}' <> 'generated_css.result' THEN
        RAISE EXCEPTION '198/543 drift: deploy_css.content_field is %, expected generated_css.result — the persist step would no longer write what the deploy ships',
            v_steps #>> '{deploy_css,config,content_field}';
    END IF;
END $$;

-- ── the persist step ───────────────────────────────────────────────────────────
UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config,
           '{workflow,steps,persist_css_to_theme}',
           jsonb_build_object(
             'action', 'query_database',
             'description',
               'bugs_open/198: persist the rendered stylesheet into the site''s css_themes row so the '
               || 'row tracks the deployed file. Without this the row is never written by this producer '
               || '(install_site_composition births it empty) and css-patch-agent later deploys that '
               || 'empty row wholesale over this file. Guards: >= 4096 BYTES, not a seed theme, not a '
               || 'row shared by more than one site, and no write when unchanged. A 0-row match is a '
               || 'normal outcome, not an error.',
             'next_step',  'deploy_css',
             'error_step', 'deploy_css',
             'output_field', 'css_persisted',
             'config', jsonb_build_object(
               'query',
                 'UPDATE css_themes ct SET css_content = $2, version = version + 1, updated_at = NOW(), '
                 || 'description = ''persisted-at-render by webdesign-agent (bugs_open/198): tracks the deployed assets/css/styles.css'' '
                 || 'FROM style_collections sc, sites s '
                 || 'WHERE sc.css_theme_id = ct.id AND s.style_collection_id = sc.id AND s.id = $1::uuid '
                 || 'AND octet_length($2) >= 4096 '
                 || 'AND ct.origin <> ''seed'' '
                 || 'AND ct.css_content IS DISTINCT FROM $2 '
                 || 'AND (SELECT count(*) FROM sites s3 JOIN style_collections sc3 ON s3.style_collection_id = sc3.id WHERE sc3.css_theme_id = ct.id) = 1 '
                 || 'RETURNING ct.id::text AS theme_id, ct.version, octet_length(ct.css_content) AS css_len',
               'params', jsonb_build_array('site_context.site_id', 'generated_css.result'),
               'output_format', 'object'
             )
           )
         ),
         '{workflow,steps,generate_css,next_step}',
         to_jsonb('persist_css_to_theme'::text)
       ),
       updated_at = NOW()
 WHERE type = 'webdesign-agent'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── VERIFY ─────────────────────────────────────────────────────────────────────
DO $$
DECLARE
    v_steps jsonb;
    v_query text;
BEGIN
    SELECT default_config #> '{workflow,steps}'
      INTO v_steps
      FROM agent_definitions
     WHERE type = 'webdesign-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF NOT (v_steps ? 'persist_css_to_theme') THEN
        RAISE EXCEPTION '198/543 verify: persist_css_to_theme missing';
    END IF;
    IF v_steps #>> '{generate_css,next_step}' <> 'persist_css_to_theme' THEN
        RAISE EXCEPTION '198/543 verify: generate_css.next_step not rewired — the step is orphaned';
    END IF;
    IF v_steps #>> '{persist_css_to_theme,next_step}' <> 'deploy_css' THEN
        RAISE EXCEPTION '198/543 verify: persist step does not reach deploy_css — the deploy is orphaned';
    END IF;
    IF v_steps #>> '{persist_css_to_theme,error_step}' <> 'deploy_css' THEN
        RAISE EXCEPTION '198/543 verify: persist step must fail OPEN to deploy_css';
    END IF;

    -- params, in order: the UPDATE binds $1 to the site and $2 to the rendered CSS
    IF v_steps #>> '{persist_css_to_theme,config,params,0}' <> 'site_context.site_id'
       OR v_steps #>> '{persist_css_to_theme,config,params,1}' <> 'generated_css.result' THEN
        RAISE EXCEPTION '198/543 verify: params wrong — got %',
            v_steps #> '{persist_css_to_theme,config,params}';
    END IF;

    -- every guard present: a missing one is a silent way to cause the harm this
    -- step exists to prevent, so each is asserted individually.
    v_query := v_steps #>> '{persist_css_to_theme,config,query}';
    IF position('octet_length($2) >= 4096' in v_query) = 0 THEN
        RAISE EXCEPTION '198/543 verify: size floor missing — a fragment could be persisted';
    END IF;
    IF position('ct.origin <> ''seed''' in v_query) = 0 THEN
        RAISE EXCEPTION '198/543 verify: seed guard missing — a library theme could be overwritten';
    END IF;
    IF position('IS DISTINCT FROM $2' in v_query) = 0 THEN
        RAISE EXCEPTION '198/543 verify: no-churn guard missing';
    END IF;
    IF position(') = 1' in v_query) = 0 OR position('sc3.css_theme_id = ct.id' in v_query) = 0 THEN
        RAISE EXCEPTION '198/543 verify: single-site guard missing — a shared row could take one site''s CSS';
    END IF;

    -- and the identity that makes the whole thing safe: what is persisted is what
    -- is deployed
    IF v_steps #>> '{deploy_css,config,content_field}' <> 'generated_css.result' THEN
        RAISE EXCEPTION '198/543 verify: deploy_css no longer commits the persisted value';
    END IF;

    RAISE NOTICE '198/543: verified — the row now tracks the file at every render, guarded four ways';
END $$;

COMMIT;
