-- 773_meta_description_backfill_message_names_the_per_page_result_keys.sql
--
-- Migration 728 fixed WHICH reasons the backfiller's completion message names.
-- This fixes WHERE it sends the reader to find them, which 728 got wrong in a
-- way nobody could have seen without reading a real multi-page run.
--
-- The message says "Read each save_result". There is a key called exactly that,
-- it is populated, it is the right shape — and on a multi-page run it holds ONLY
-- THE LAST PAGE. The per-page truth is in a numbered series beside it.
--
-- Source: bugs_open/442 §11d.
--
-- ══ WHAT WAS MEASURED ════════════════════════════════════════════════════════
-- [MEASURED 2026-09-04, orchestration_states, whole 25 h retention window]
-- A step inside a loop's sub_workflow writes its output_field TWICE: the
-- per-iteration series (save_result_0, save_result_1, …) AND a bare key holding
-- the last iteration. On this lane's only multi-page run (09-03 14:03, 2 pages):
--     save_result_0 -> page 2bcf3e28…  updated true
--     save_result_1 -> page 34d8d807…  updated true
--     save_result   -> page 34d8d807…  == save_result_1
-- Not this workflow's quirk: on page-content-writer, bare copy_gate equalled
-- copy_gate_<max N> on 20 of 20 runs and never the first. NOT uniform either —
-- section_output in that same loop has no bare form — so a reader cannot learn
-- the rule from one example, which is exactly why the message has to say it.
-- 1 of the 4 runs in the window covered more than one page, so the misreading
-- this prevents is not hypothetical arithmetic.
--
-- ⚠ [UNTRACED] which code writes the bare key. makeIterationOutputField rewrites
-- each injected step's OutputField to {field}_{N}, so the injected steps are
-- demonstrably NOT the writer. This migration therefore describes the OBSERVED
-- shape and asserts nothing about the mechanism. LANDMINES.md carries the same
-- caveat. Do not let a later edit quietly upgrade the observation to a cause.
--
-- ══ WHY THIS IS WORTH A MIGRATION AND NOT JUST A BUG NOTE ════════════════════
-- 442 exists because the ONE human-facing surface was misleading (§4: it named
-- four of seven reasons). This is the same defect class on the same surface: the
-- instruction is followable, the key it names exists, and following it produces
-- a false all-clear. A refusal on page 0 followed by a write on page 1 reads as
-- "updated": true at the obvious key.
--
-- Nothing is LOST when it misleads — §10b's work item files regardless, which is
-- the property that makes the shipped fix worth having — so this is a surface
-- correction, not a data-loss fix. Said plainly in the message itself.
--
-- ══ WHAT CHANGES, AND WHAT DOES NOT ══════════════════════════════════════════
-- One string, on one row, in one step, APPENDED to rather than rewritten: the
-- update concatenates onto the row's own current value, so 728's text cannot be
-- mis-transcribed here and cannot drift from what is live. Config — live on
-- apply, no image, no roll. Nothing branches on this message.
--
-- Reversible: 773_..._ROLLBACK.sql strips exactly the appended block.

BEGIN;

SELECT snapshot_agent('meta-description-backfiller',
                      '773_meta_description_backfill_message_names_the_per_page_result_keys.sql: pre-update');

-- DRIFT GUARD. The row must be the one 728 left behind, and must not already
-- carry this change. Both re-application and another session's rewrite abort.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
      FROM agent_definitions
     WHERE type = 'meta-description-backfiller'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       -- it is the 728-era message: all seven reasons plus the anti-rot pointer
       AND default_config->'workflow'->'steps'->'complete'->'config'->>'result_message'
           LIKE '%metaDescriptionFailsCopyGates%'
       AND default_config->'workflow'->'steps'->'complete'->'config'->>'result_message'
           LIKE '%voice_gate_unreadable%'
       -- and it has NOT already been amended by this migration
       AND default_config->'workflow'->'steps'->'complete'->'config'->>'result_message'
           NOT LIKE '%save_result_0%';
    IF n <> 1 THEN
        RAISE EXCEPTION
            'ABORT: expected exactly 1 live meta-description-backfiller carrying the '
            '728 result_message and not yet carrying this amendment, found %. Either '
            'this migration has ALREADY applied, or another session has edited this '
            'workflow. Re-read the live row before re-running.', n;
    END IF;
END $$;

-- Pre-image for the positive control: the WHOLE default_config, so the control
-- can prove the update touched exactly one path. A jsonb_set on a wrong path
-- satisfies every "the new text is present" check while rewriting the workflow.
CREATE TEMP TABLE _pre_773 ON COMMIT DROP AS
SELECT id, default_config
  FROM agent_definitions
 WHERE type = 'meta-description-backfiller'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- APPEND, referencing the row's own current value. 728's text is never restated
-- in this file, so it cannot be transcribed wrongly and cannot go stale.
UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,complete,config,result_message}',
           to_jsonb(
               (default_config->'workflow'->'steps'->'complete'->'config'->>'result_message')
               || $msg$

⚠ WHERE THIS RUN'S RESULTS ACTUALLY ARE — AND THE KEY THAT WILL MISLEAD YOU.

Each page's outcome is written to collected_data as save_result_0, save_result_1, ... — one per page, in write order. Read those.

There is ALSO a bare save_result. It holds ONLY THE LAST PAGE. So on a run covering more than one page, a refusal on an earlier page sits in save_result_0 while the bare key reads "updated": true, and a reader who takes the obvious key sees a clean run. [MEASURED 2026-09-04, bugs_open/442 §11d] the bare key equalled the last page on every run checked; the same convention holds on other loops (page-content-writer's copy_gate and generated_content) and, confusingly, NOT on every field in them — so do not infer it from one example.

Two cheap controls: the number of save_result_<N> keys should equal the number of descriptions the writer returned, and a short series means pages are MISSING, not that they passed.

Nothing is lost when this misleads you: a copy-gate refusal also files a meta_description_refused work item at meta-description-repair, and THAT is the durable record. This message is the convenient surface, not the authoritative one.$msg$::text),
           false),
       updated_at = now()
 WHERE type = 'meta-description-backfiller'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- VERIFY. DO/RAISE, not SELECTs: ON_ERROR_STOP does not fire on a non-empty
-- result set, so a block of SELECTs cannot stop the COMMIT.
DO $$
DECLARE
    msg       text;
    missing   text[] := '{}';
    r         text;
    untouched int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'complete'->'config'->>'result_message'
      INTO msg
      FROM agent_definitions
     WHERE type = 'meta-description-backfiller'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF msg IS NULL THEN
        RAISE EXCEPTION 'ABORT: the result_message path does not exist after the update — '
                        'jsonb_set created a new key instead of replacing one';
    END IF;

    -- 728's work must survive verbatim. Appending cannot drop a reason, so this
    -- asserts the concatenation took the RIGHT operand as its base and not, say,
    -- an empty string or another step's message.
    FOREACH r IN ARRAY ARRAY['empty_candidate','candidate_looks_internal','candidate_too_long',
                             'already_has_description','voice_tell','banned_claim',
                             'voice_gate_unreadable','metaDescriptionFailsCopyGates']
    LOOP
        IF position(r in msg) = 0 THEN
            missing := missing || r;
        END IF;
    END LOOP;
    IF array_length(missing, 1) > 0 THEN
        RAISE EXCEPTION 'ABORT: the amended message has LOST 728 content: %', missing;
    END IF;

    -- The new content, asserted on the two parts that carry the warning. Naming
    -- the series without saying the bare key is only the last page would leave a
    -- reader with two plausible keys and no reason to prefer either.
    IF position('save_result_0' in msg) = 0
       OR position('ONLY THE LAST PAGE' in msg) = 0 THEN
        RAISE EXCEPTION 'ABORT: the amendment did not land — the message does not name the '
                        'per-page series and the bare-key warning together';
    END IF;

    -- POSITIVE CONTROL. Reconstruct the pre-image with ONLY result_message
    -- replaced and require it to equal what is stored. Anything else the UPDATE
    -- touched — the loop, continue_on_error, the writer prompt, another step —
    -- fails here, and no "the text is present" check above could have seen it.
    SELECT count(*) INTO untouched
      FROM agent_definitions ad JOIN _pre_773 pre ON pre.id = ad.id
     WHERE ad.default_config
           = jsonb_set(pre.default_config,
                       '{workflow,steps,complete,config,result_message}',
                       to_jsonb(msg), false);
    IF untouched <> 1 THEN
        RAISE EXCEPTION 'ABORT: the update changed more than the result_message '
                        '(% of 1 rows match the pre-image with only that path replaced)', untouched;
    END IF;

    RAISE NOTICE '773: the completion message now names the per-page save_result_<N> series '
                 'and warns that the bare save_result is only the last page. 728 content '
                 'intact; everything else in default_config byte-identical to the pre-image.';
END $$;

COMMIT;
