-- 728_meta_description_backfill_result_message_names_the_copy_gates.sql
--
-- The meta-description backfiller's `complete` step carries the ONLY surface
-- that tells a person how to read the run's outcome, and it names FOUR of the
-- action's SEVEN refusal reasons. The three it omits are the three copy-gate
-- refusals — the expensive ones, the ones that leave a page blank on an hourly
-- schedule, and the ones that need a human. A reader following the message's
-- own instruction concludes the gates cannot refuse.
--
-- Source: bugs_open/442 §4 (candidate 3). This is the CONFIG-ONLY half of that
-- bug. It removes an actively misleading surface; it does NOT make a refusal
-- loud — the refusal is still a `logger.Warn` nothing asserts on (§2). Route B
-- (a durable surface) is an open owner question, deliberately not taken here.
--
-- ══ THE ASYMMETRY THAT PRODUCED THE STALE LIST ═══════════════════════════════
-- [MEASURED 2026-09-03, re-run against the tree today]
-- The four the message names are string literals in the result maps:
--     grep -nE '"reason": *"[a-z_]+"' save_page_meta_description_action.go
--       -> :161 empty_candidate  :175 candidate_looks_internal
--          :182 candidate_too_long  :233 already_has_description   = 4
-- The three it omits are returned as BARE STRINGS from the gate helper, so the
-- obvious single grep finds exactly the four already documented and reports the
-- list as complete:
--     grep -nE 'return "[a-z_]+",' save_page_meta_description_action.go
--       -> :316 voice_gate_unreadable  :334 voice_tell  :341 banned_claim = 3
-- It takes TWO greps. That is why the omission survived from 2026-08-19, when
-- the owner requirement of bugs_open/320 added the gates to the action, to today.
--
-- ══ WHY THE REPLACEMENT CARRIES ITS OWN STALENESS WARNING ════════════════════
-- Enumerating seven instead of four fixes today and rots tomorrow by exactly the
-- same route. The owner ruling of 2026-08-22 requires a COUNT to carry the date
-- it was counted; an ENUMERATION has no equivalent rule and rots identically
-- (bugs_open/442 §4 raises this as a candidate norm — the THIRD stale-by-addition
-- list found in one task). So the new message names the seven AND tells the
-- reader the list is a copy, where the authoritative set lives, and that finding
-- it takes two greps. A reader who follows that cannot be misled by an eighth.
--
-- ══ WHAT CHANGES, AND WHAT DOES NOT ══════════════════════════════════════════
-- One string, on one row, in one step. Config — live on apply, no image, no roll.
-- No behaviour changes: `complete_workflow` prints this message and nothing
-- branches on it. Every other step, including the loop and its
-- `continue_on_error`, is asserted byte-identical by the control below.
--
-- Reversible: 728_..._ROLLBACK.sql restores the four-reason message verbatim.

BEGIN;

SELECT snapshot_agent('meta-description-backfiller',
                      '728_meta_description_backfill_result_message_names_the_copy_gates.sql: pre-update');

-- DRIFT GUARD. Abort rather than clobber if the row is not in the state this
-- migration was written against — another session editing this workflow, or
-- this file having already applied, are both caught here.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
      FROM agent_definitions
     WHERE type = 'meta-description-backfiller'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config->'workflow'->'steps'->'complete'->'config'->>'result_message'
           = 'Meta description backfill finished. Read each save_result: "updated" true is a write, false carries a named reason (empty_candidate / candidate_looks_internal / candidate_too_long / already_has_description).';
    IF n <> 1 THEN
        RAISE EXCEPTION
            'ABORT: expected exactly 1 live meta-description-backfiller carrying the '
            'four-reason result_message, found %. Another session has edited this '
            'workflow, or this migration has ALREADY applied. Re-read before re-running.', n;
    END IF;
END $$;

-- Pre-image for the positive control. Captured as the WHOLE default_config, so
-- the control can prove that the update touched exactly one path and nothing
-- else — a jsonb_set on a wrong path satisfies every "the new text is present"
-- check while silently rewriting the workflow around it.
CREATE TEMP TABLE _pre_728 ON COMMIT DROP AS
SELECT id, default_config
  FROM agent_definitions
 WHERE type = 'meta-description-backfiller'
   AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

UPDATE agent_definitions
   SET default_config = jsonb_set(
           default_config,
           '{workflow,steps,complete,config,result_message}',
           to_jsonb($msg$Meta description backfill finished. Read each save_result: "updated" true is a write; false carries a named "reason" and a "detail" — and NOTHING downstream retries, escalates or records it (bugs_open/442). This message is the only surface.

The reasons split two ways, and they ask different things of you:

  NOTHING PUBLISHABLE WAS OFFERED — no action needed. empty_candidate, candidate_looks_internal, candidate_too_long, already_has_description.

  A COPY GATE REFUSED THE SENTENCE AND THE PAGE STAYS BLANK — voice_tell, banned_claim, voice_gate_unreadable. A person has to judge the copy. The hourly schedule re-offers the same page and can refuse it again indefinitely; nothing else will notice.

This list is a COPY of the action's vocabulary and rots by ADDITION: it named only the first four from 2026-08-19, when the copy gates were added, to 2026-09-03 — so the three refusals that most needed a reader were exactly the three it omitted. The authoritative set is what save_page_meta_description returns: platform/orchestration/actions/save_page_meta_description_action.go, the "reason" keys in the result maps PLUS the bare returns from metaDescriptionFailsCopyGates. That is two greps, not one. Check there before trusting this list.$msg$::text),
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

    -- The load-bearing assertion, and it is the whole point of the migration:
    -- every reason the action can return must be nameable from this message.
    FOREACH r IN ARRAY ARRAY['empty_candidate','candidate_looks_internal','candidate_too_long',
                             'already_has_description','voice_tell','banned_claim',
                             'voice_gate_unreadable']
    LOOP
        IF position(r in msg) = 0 THEN
            missing := missing || r;
        END IF;
    END LOOP;
    IF array_length(missing, 1) > 0 THEN
        RAISE EXCEPTION 'ABORT: the new message still omits %', missing;
    END IF;

    -- Anti-rot assertion. The pointer to the source of truth is the only part
    -- that survives an eighth reason being added, so it is asserted, not trusted.
    IF position('save_page_meta_description_action.go' in msg) = 0
       OR position('metaDescriptionFailsCopyGates' in msg) = 0 THEN
        RAISE EXCEPTION 'ABORT: the new message does not point at the authoritative reason list';
    END IF;

    -- POSITIVE CONTROL. Reconstruct the pre-image with ONLY the result_message
    -- replaced and require it to equal what is stored. Anything else the UPDATE
    -- touched — the loop, continue_on_error, the writer prompt, another step —
    -- fails here, and no "the text is present" check above could have seen it.
    SELECT count(*) INTO untouched
      FROM agent_definitions ad JOIN _pre_728 pre ON pre.id = ad.id
     WHERE ad.default_config
           = jsonb_set(pre.default_config,
                       '{workflow,steps,complete,config,result_message}',
                       to_jsonb(msg), false);
    IF untouched <> 1 THEN
        RAISE EXCEPTION 'ABORT: the update changed more than the result_message '
                        '(% of 1 rows match the pre-image with only that path replaced)', untouched;
    END IF;

    RAISE NOTICE '728: result_message now names all 7 refusal reasons, splits them by '
                 'what they ask of a reader, and points at the authoritative list. '
                 'Everything else in default_config is byte-identical to the pre-image.';
END $$;

COMMIT;
