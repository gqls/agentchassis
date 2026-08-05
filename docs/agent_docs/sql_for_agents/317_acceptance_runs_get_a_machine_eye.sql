-- 317_acceptance_runs_get_a_machine_eye.sql
--
-- TL-035 (e) — wire the machine eye. Owner decision 2026-08-03: close "nobody
-- looks at the renders" with a VISION check rather than a human surface, so the
-- looking happens without anyone having to remember to look.
--
-- WHAT THIS DOES, in one sentence: after the acceptance verdict is already
-- written, send the run's landing-state renders to a vision-capable model, and
-- append its critique as its OWN doc_note.
--
--   ensure_site_record -> load_docs -> request_run -> judge -> look -> record_look -> complete
--                                          |            |        |          |
--                                          +--> complete_error <--+          |
--                                                       |                    |
--                                       look/record_look error_step -> complete_no_look
--
-- FOUR SAFETY PROPERTIES, each deliberate:
--
--  1. THE VERDICT CANNOT BE LOST. `look` runs AFTER `judge`, which is what
--     writes the acceptance-run note and raises the improve_tool item. By the
--     time the camera-reader runs, the valuable output is already persisted.
--     A vision failure can therefore cost the critique and nothing else.
--
--  2. A FAILED LOOK IS NOT A FAILED ACCEPTANCE RUN. Both new steps route
--     error_step to `complete_no_look` — a SUCCESS terminal, not
--     `complete_error`. The acceptance run genuinely succeeded; only the
--     optional look failed. This matters more than it looks: execute_vision_prompt
--     is deliberately fail-loud in v1 (no tolerate_truncation, and it errors
--     when there are zero renders), and zero renders is a NORMAL outcome for a
--     run whose profiles all failed. Routing that to complete_error would have
--     turned a working acceptance run into a reported failure.
--
--  3. A SEPARATE STEP, so `complete` vs `complete_no_look` on the orchestration
--     row tells you which happened without digging into __step_error. A silent
--     no-op is the failure mode this estate keeps paying for.
--
--  4. A SEPARATE NOTE CATEGORY. The critique files under `render-critique`,
--     NOT `acceptance-run`. The lane's re-check query and contact_sheet.py both
--     read `acceptance-run`; polluting it would break both, and would also
--     re-create the exact two-producers-one-category trap already filed in
--     LANDMINES.md against this very category.
--
-- WHAT THIS DELIBERATELY DOES NOT DO:
--   * It raises NO work items. The findings -> work-item drain belongs to
--     vigilant_designer_offer_analysis ("this lane owns: the drain, the critic"),
--     and their A2 critic + Gemini-vs-Claude trial is still pending. This wires
--     the eye to OUR acceptance runs only; it does not seed a rival general
--     critic and does not touch render-audit-agent.
--   * It asserts nothing. A render is a look, never a verdict (TL-035's own
--     landmine). The prompt forbids the model from claiming the page failed.
--
-- ORDERING: config only. Every action used (execute_vision_prompt,
-- append_doc_note, complete_workflow) is already registered and live in the
-- chassis binary — verified at both replicas before this was written, and the
-- 188 landing-state fix is live on v1.0.1251+. So there is no image-before-config
-- constraint here: there is no new Go.
--
-- PRE-VERIFIED AT THE POD (2026-08-05, both chassis replicas):
--   grep -acF "landing state"   /app/agent-chassis -> 1  (188 fix live)
--   ANTHROPIC_API_KEY                              -> SET
--   AWS_ACCESS_KEY_ID / B2_APPLICATION_KEY         -> SET  (storage client)
-- and tool-acceptance-agent has NO ai_service block at root OR step level, so
-- MDL-039 (root SHADOWS step-level) does not bite: the step block below applies.

BEGIN;

-- ── 1. the look ────────────────────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,look}',
    $json$
    {
      "action": "execute_vision_prompt",
      "description": "TL-035 (e): send the run's landing-state renders to a vision model. Best-effort, never a verdict; a failure here cannot affect the acceptance result, which judge has already written.",
      "output_field": "vision_look",
      "next_step": "record_look",
      "config": {
        "error_step": "complete_no_look",
        "images_field": "browser_run",
        "max_images": 4,
        "output_type": "text",
        "ai_service": {
          "provider": "anthropic",
          "model": "claude-sonnet-5",
          "api_key_env_var": "ANTHROPIC_API_KEY",
          "max_tokens": 4000
        },
        "prompt_template": "You are looking at full-page screenshots of a web page that has ALREADY PASSED every automated check configured for it. Those checks cover structure and behaviour. They do NOT cover spacing, alignment, overlap, contrast or anything else that only an eye catches. Your job is to look for what no assertion covered.\n\nImages, in order:\n{{ .vision_image_manifest }}\n\nReport ONLY defects a real visitor would see on the page as served. For each one: what it is, roughly where on the page, and which image (index and profile) shows it.\n\nDo NOT report any of the following. Each is a known artefact of how these pictures are taken, not a fault in the page:\n- The navigation bar appearing part-way down the image rather than at the top. These are full-page captures and a sticky header paints where it was scrolled to.\n- The image being extremely tall, or content looking small relative to the image. A full-page capture is not a viewport view.\n- Any page that looks blank, empty, half-populated, or mid-interaction. These pictures are normally taken before the checks touch the page, but a rare fallback path photographs the page after they have driven it. If what you see looks like the aftermath of a form being cleared or a control being toggled, say INCONCLUSIVE for that image and move on. Do not report it as a defect.\n\nIf you find nothing worth a human's attention, say so plainly in one line. Finding nothing is a perfectly good answer and is much better than inventing something marginal.\n\nYou are NOT judging whether the page passes or fails - that has already been decided and is not yours to revisit. You are a second pair of eyes, and your output is a note for a person to read.\n\nWrite plain prose. No JSON, no code fences, no headings, no preamble."
      }
    }
    $json$::jsonb)
WHERE type = 'tool-acceptance-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── 2. persist it, under its OWN category ──────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,record_look}',
    $json$
    {
      "action": "append_doc_note",
      "description": "TL-035 (e): append the vision critique as its own note. Category render-critique, deliberately NOT acceptance-run.",
      "output_field": "render_critique_note",
      "next_step": "complete",
      "config": {
        "error_step": "complete_no_look",
        "subject_type": "tool",
        "subject_key_field": "input_data.spec.function",
        "note_body_field": "vision_look.result",
        "note_categories": ["render-critique"],
        "note_source": "tool-acceptance-vision",
        "created_by": "tool-acceptance-agent"
      }
    }
    $json$::jsonb)
WHERE type = 'tool-acceptance-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── 3. a terminal that says "ran, did not look" ────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{workflow,steps,complete_no_look}',
    $json$
    {
      "action": "complete_workflow",
      "description": "Acceptance run completed; the optional vision look did not produce a note. NOT an error - the verdict stands.",
      "config": {
        "success_message": "Acceptance run completed (no render critique)",
        "multiple_output_fields": ["acceptance_verdict"]
      }
    }
    $json$::jsonb)
WHERE type = 'tool-acceptance-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── 4. route judge into the look ───────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config, '{workflow,steps,judge,next_step}', '"look"'::jsonb)
WHERE type = 'tool-acceptance-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ── VERIFY. DO/RAISE, because a verify block of SELECTs cannot stop a COMMIT
--    (ON_ERROR_STOP ignores a non-empty result). Both guards were INDUCED
--    before the real apply — see the lane RUNBOOK.
DO $$
DECLARE
  v_cfg     jsonb;
  v_missing text;
BEGIN
  SELECT default_config INTO v_cfg FROM agent_definitions
   WHERE type = 'tool-acceptance-agent'
     AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

  IF v_cfg IS NULL THEN
    RAISE EXCEPTION '317: no live tool-acceptance-agent row';
  END IF;

  -- (a) what we just wrote
  IF v_cfg #>> '{workflow,steps,look,action}' IS DISTINCT FROM 'execute_vision_prompt' THEN
    RAISE EXCEPTION '317: look step not set (found %)', v_cfg #>> '{workflow,steps,look,action}';
  END IF;
  IF v_cfg #>> '{workflow,steps,record_look,action}' IS DISTINCT FROM 'append_doc_note' THEN
    RAISE EXCEPTION '317: record_look step not set (found %)', v_cfg #>> '{workflow,steps,record_look,action}';
  END IF;
  IF v_cfg #>> '{workflow,steps,complete_no_look,action}' IS DISTINCT FROM 'complete_workflow' THEN
    RAISE EXCEPTION '317: complete_no_look step not set (found %)', v_cfg #>> '{workflow,steps,complete_no_look,action}';
  END IF;
  IF v_cfg #>> '{workflow,steps,judge,next_step}' IS DISTINCT FROM 'look' THEN
    RAISE EXCEPTION '317: judge does not route to look (found %)', v_cfg #>> '{workflow,steps,judge,next_step}';
  END IF;

  -- (b) THE SAFETY PROPERTIES, asserted rather than trusted to review
  IF v_cfg #>> '{workflow,steps,look,config,error_step}' IS DISTINCT FROM 'complete_no_look'
     OR v_cfg #>> '{workflow,steps,record_look,config,error_step}' IS DISTINCT FROM 'complete_no_look' THEN
    RAISE EXCEPTION '317: a look step routes failure somewhere other than complete_no_look — a failed look must not fail the run';
  END IF;
  IF v_cfg #> '{workflow,steps,record_look,config,note_categories}' @> '["acceptance-run"]'::jsonb THEN
    RAISE EXCEPTION '317: the critique must NOT file under acceptance-run — it would break the re-check query and contact_sheet.py';
  END IF;

  -- (c) NEIGHBOUR keys — a guard that checks only what it just wrote cannot
  --     tell a surgical jsonb_set from a write that flattened the object.
  --     147's profiles and 292's capture_renders are the canaries.
  v_missing := '';
  -- NULL-SAFE ON PURPOSE, and this was a real bug caught by inducing the mutant:
  -- a flattened config makes the #> return NULL, `NULL @> '[...]'` is NULL, and
  -- `IF NULL THEN` does not fire — so the plain @> form silently MISSED the exact
  -- write this canary exists to catch. (`IS DISTINCT FROM` below is NULL-safe,
  -- which is why capture_renders fired and this one did not.)
  IF v_cfg #> '{workflow,steps,request_run,config,profiles}' IS NULL
     OR NOT (v_cfg #> '{workflow,steps,request_run,config,profiles}' @> '["desktop","mobile"]'::jsonb) THEN
    v_missing := v_missing || ' 147-profiles';
  END IF;
  IF v_cfg #>> '{workflow,steps,request_run,config,capture_renders}' IS DISTINCT FROM 'true' THEN
    v_missing := v_missing || ' 292-capture_renders';
  END IF;
  IF v_cfg #>> '{workflow,steps,judge,action}' IS DISTINCT FROM 'judge_acceptance_results' THEN
    v_missing := v_missing || ' judge-action';
  END IF;
  IF v_cfg #>> '{workflow,start_step}' IS DISTINCT FROM 'ensure_site_record' THEN
    v_missing := v_missing || ' start_step';
  END IF;
  IF v_missing <> '' THEN
    RAISE EXCEPTION '317: neighbour key(s) did not survive the write:%', v_missing;
  END IF;

  RAISE NOTICE '317 OK: look + record_look + complete_no_look wired; judge -> look; neighbours intact';
END $$;

COMMIT;
