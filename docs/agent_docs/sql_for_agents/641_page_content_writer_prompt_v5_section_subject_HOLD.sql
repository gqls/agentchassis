-- 641 — page-content-writer prompt v5: render the section's SUBJECT when the
-- plan assigned one.
--
-- ⚠⚠ _HOLD, TWO gates, both must clear before hand-applying:
--   1. the image carrying sectionPlanItem.Subject has rolled, and 639 is
--      applied (else current_section.subject is never set and the block is
--      dead text);
--   2. **THE OWNER HAS READ THE INSERTED BLOCK BELOW.** The v4 prompt's
--      owner approval (2026-08-09, RFC_016 §5.2) "attaches to the committed
--      text … any later edit voids the approval and needs a fresh read".
--      This edit is that later edit. The delta is ONLY the block between the
--      INSERTED TEXT markers — nothing else in the prompt changes, and the
--      verify block proves it (anchors + em-dash census unchanged at 5).
--
-- !!! GATE 2 OUTCOME, 2026-09-02: THE READ CAME BACK AS A REDRAFT DIRECTION,
--     NOT AN APPROVAL -- DO NOT APPLY THIS FILE AS WRITTEN. The inserted block
--     below is now a historical draft: negative-instruction prompting in
--     production vocabulary, which is what the owner's directive rejects
--     (positive prompting only; written in the language expected back; the
--     prompt's own prose is the demonstration, no specimen answer). The
--     replacement is being drafted with him:
--     docs/agent_docs/docs024_key_docs_latest/finetuning_uk_service/DRAFT_2026-09-02_641_positive_prompt_candidates.md
--     (three framing candidates, his pick pending; finetuning lane test-renders,
--     the apis.uk lane writes the final SQL, and his approval attaches to the
--     EXACT final words -- RFC_016 s5.2, voids on any later edit).
--     NEW PRE-APPLY OBLIGATION: all three candidates enumerate the SIBLING
--     subjects, and that range render is UNTESTED -- prove it against real
--     loop CollectedData before the redraft ships (empty/absent subjects must
--     drop out of the enumeration, not render blank; missingkey=zero hides
--     the failure). Record in apis.uk NOTES 2026-09-02.
--
-- Placement: immediately BEFORE the Verified Facts block, so the writer reads
-- "what this section is about" before "which facts it may state". Renders
-- ONLY when the plan assigned a subject ({{if .current_section.subject}});
-- every unassigned section's prompt is byte-identical to v4.
--
-- ── INSERTED TEXT (the whole delta; plain hyphens, no em dashes) ──────────
-- {{if .current_section.subject}}## This section's subject
-- {{.current_section.subject}}
--
-- Write THIS section about that subject specifically. Sibling sections on
-- this page carry their own subjects - do not restate theirs, and do not
-- widen this section into a general treatment of the page's topic.
--
-- {{end}}
-- ── END INSERTED TEXT ─────────────────────────────────────────────────────

--
-- ── VERIFY THE ROLL FIRST (council 4bd35ed8 r2, debug_historian: "post-roll" is
--    an assumption until the RUNNING BINARY says so; same-tag rebuilds have
--    shipped nothing before). Two probes, both against the pod, controls in the
--    same breath:
--      kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--        -> git merge-base --is-ancestor 35905c547 <stamp sha>   (must exit 0)
--      # fallback if the startup line has scrolled (capability probe + controls):
--      P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
--      kubectl -n ai-persona-system exec $P -- grep -ac 'section_subjects' /proc/1/exe   # >0 = shipped
--      kubectl -n ai-persona-system exec $P -- grep -ac 'section_facts' /proc/1/exe      # positive control, must be >0
--
-- ── AFTER HAND-APPLYING: _HOLD files never reach the migration ledger, so the
--    record is THIS FILE — append one line directly below this block and commit
--    it (pathspec):  -- APPLIED <date> by <session>; roll verified at <stamp sha>
--
-- ── RECORD THE OWNER'S READ (gate 2): a dated line in the apis lane's NOTES
--    quoting his words, and name it in this file's APPLIED line — an unrecorded
--    read is indistinguishable from a skipped one.

SELECT snapshot_agent('page-content-writer', '641_page_content_writer_prompt_v5_section_subject_HOLD.sql: pre-update');

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_641 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'page-content-writer' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    t text;
    c1 int; nrows int;
BEGIN
    -- Pre-flight (council 4bd35ed8 round 1): SELECT INTO takes the FIRST of N
    -- rows silently; a duplicate active row would half-apply. Count first.
    SELECT count(*) INTO nrows FROM agent_definitions
    WHERE type = 'page-content-writer' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF nrows <> 1 THEN
        RAISE EXCEPTION '641: expected exactly 1 active page-content-writer row BEFORE writing, found %', nrows;
    END IF;
    SELECT default_config->'workflow'->'steps'->'process_sections_loop'->'config'
           ->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'page-content-writer' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF t IS NULL THEN
        RAISE EXCEPTION '641: page-content-writer generate_content prompt_template not found';
    END IF;
    IF position('{{if .current_section.subject}}' in t) > 0 THEN
        RAISE EXCEPTION '641: already applied — subject block present';
    END IF;
    c1 := (length(t) - length(replace(t, '{{if .current_section.facts_scoped}}', ''))) / length('{{if .current_section.facts_scoped}}');
    IF c1 <> 1 THEN
        RAISE EXCEPTION '641: live prompt has drifted (facts_scoped anchor count % not 1) — re-derive from the live row', c1;
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(
            replace(
                default_config->'workflow'->'steps'->'process_sections_loop'->'config'
                    ->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template',
                '{{if .current_section.facts_scoped}}',
                E'{{if .current_section.subject}}## This section''s subject\n{{.current_section.subject}}\n\nWrite THIS section about that subject specifically. Sibling sections on this page carry their own subjects - do not restate theirs, and do not widen this section into a general treatment of the page''s topic.\n\n{{end}}{{if .current_section.facts_scoped}}'
            )
        )
    ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    t text;
    dashes int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'process_sections_loop'->'config'
           ->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
    WHERE type = 'page-content-writer' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('{{if .current_section.subject}}' in t) = 0
       OR position('## This section''s subject' in t) = 0
       OR position('{{if .current_section.facts_scoped}}' in t) = 0
       OR position('assigned to THIS section' in t) = 0
       OR position('{{.voice_style}}' in t) = 0 THEN
        RAISE EXCEPTION '641: verify failed — the subject block, the facts block it must precede, or a preserved landmark is missing';
    END IF;
    dashes := length(t) - length(replace(t, '—', ''));
    IF dashes <> 5 THEN
        RAISE EXCEPTION '641: verify failed — em-dash census is % not 5; the insertion touched more than the one anchor', dashes;
    END IF;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run): restore from agent_definitions_bak_641, or:
--   the inserted block is exactly the text between the INSERTED TEXT markers
--   above — replace it with '' in the same field, same jsonb_set shape.
