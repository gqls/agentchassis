-- 641 — page-content-writer prompt v5: render the section's SUBJECT when the
-- plan assigned one. REDRAFTED TWICE to the owner's positive-prompting
-- directive of 2026-09-02. The block below is the second redraft, decided by
-- him on 2026-09-03 after a blind read of four candidates: the subject is
-- PRINTED VERBATIM on its own line, and the sibling list is unchanged from his
-- 2026-09-02 pick "C". Test-rendered against eight real-row fixtures BEFORE
-- this SQL was written (framework_prompts_positive_voice/render_test/ -
-- OUTPUT.txt is the evidence, candidate A4; the finetuning lane's
-- render_test_641/ holds the round-1 record). The template bytes here are
-- copied verbatim from that harness, anchor swapped.
--
-- WHY C'S FIRST SENTENCE WENT (2026-09-03, and it is a data decision, not a
-- taste one): C opened "You'll want to know {{.current_section.subject}}.",
-- a fixed frame with a hole, so only a short lower-case fragment could fill
-- it. Measured on the harness: the REAL gamedesign.uk planner subjects render
-- "You'll want to know Brief description of the sister-site relationship with
-- gamesdesign.co.uk and what each site covers.", and the owner's own example
-- opening line renders with a doubled full stop. The frame and prose-shaped
-- subjects are mutually exclusive; he chose the prose. Subjects are now
-- authored as the section's OPENING LINE, addressed to the reader in the
-- site's voice - a phrasing spec goes to the finetuning and apis.uk lanes,
-- and a build-site-planner nudge follows as its own migration.
--
-- !! THE DELTA IS NOW TWO CHANGES IN ONE TRANSACTION, not one:
--   (a) the subject block inserted before the Verified Facts block;
--   (b) "sections_for_render" appended to generate_content.config.input_fields.
--   (b) is load-bearing: the prompt renders against
--   ExtractFields(CollectedData, input_fields) - a SUBSET the step names
--   (ai_actions.go -> unified_extractor.go) - and without (b) the sibling
--   range sees nil and renders an EMPTY list, silently (fixture D: "also
--   covers, each in its own section:" followed by NOTHING, no error). The
--   verify block asserts BOTH; a verify on the template alone passes on a
--   config that renders an empty list.
--
-- __ _HOLD, TWO gates, both must clear before hand-applying:
--   1. CLEARED 2026-09-02: the image carrying sectionPlanItem.Subject rolled
--      (pod-verified, apis lane NOTES) and 639 is applied (live row verified).
--   2. **THE OWNER HAS READ THE INSERTED BLOCK BELOW, AS WRITTEN HERE.**
--      RFC_016 s5.2: approval attaches to the committed text; any later edit
--      voids it. History: the v5 first cut FAILED this gate 2026-09-02 - the
--      read returned a REDRAFT direction (positive prompting only, response
--      register, no specimen answer), the framing candidates went to him, he
--      picked C's words. On 2026-09-03 he read four candidates BLIND (letters
--      randomised, key sealed before presentation, C included as the control),
--      leaned to C for its warmth, and on being shown what its frame does to
--      real planner subjects and to his own example line, chose to drop the
--      frame and keep the rest: that is the block below. HIS READ OF THESE
--      EXACT BYTES IS STILL OWED - what he has approved is the shape and the
--      rendered example, and what applies is this template. Change nothing in
--      the block without saying so to him, to the apis.uk lane (whose file
--      this is) and to the finetuning lane (whose subject data it renders).
--
-- Placement: immediately BEFORE the Verified Facts block. Renders ONLY when
-- the plan assigned a subject ({{if .current_section.subject}}); every
-- unassigned section's prompt is byte-identical to v4 (fixture B). Sibling
-- exclusion is by SUBJECT, not name - names repeat (generic-text-block x3 on
-- the real playground row); a subjectless sibling drops out cleanly, Go's
-- "and" short-circuits (fixture E). The tier-1 planner-subject question that
-- stood here as an OWNER QUESTION is ANSWERED by this redraft: with no frame,
-- any subject shape renders grammatically, so the planner nudge is no longer
-- a correctness fix. It is still owed as a REGISTER fix (the planner writes
-- capitalised noun phrases with em dashes; those now print verbatim as a
-- section's opening line), and it lands as its own migration on the apis.uk
-- lane's rule-17 anchor. Two fixtures added for this cut: F, two sections
-- sharing one subject (subject-equality exclusion drops both from the list -
-- a data defect rule 17 forbids, visible not silent); G, subjects written as
-- full sentences, which is the shape the owner asked for.
--
-- __ INSERTED TEXT (delta (a); plain hyphens, no em dashes) _______________
-- {{if .current_section.subject}}## This section
--
-- {{.current_section.subject}}
--
-- {{.current_page.title}} also covers, each in its own section:
-- {{range $s := .sections_for_render.sections_ready}}{{if and $s.subject (ne $s.subject $.current_section.subject)}}- {{$s.subject}}
-- {{end}}{{end}}
-- {{end}}
-- __ END INSERTED TEXT ____________________________________________________
--
-- __ VERIFY THE ROLL FIRST (gate 1 - cleared 2026-09-02, but re-verify if
--    applying later; same-tag rebuilds have shipped nothing before):
--      kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
--        -> git merge-base --is-ancestor 35905c547 <stamp sha>   (must exit 0)
--      # fallback if the startup line has scrolled (capability probe + controls):
--      P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
--      kubectl -n ai-persona-system exec $P -- grep -ac 'section_subjects' /proc/1/exe   # >0 = shipped
--      kubectl -n ai-persona-system exec $P -- grep -ac 'section_facts' /proc/1/exe      # positive control, must be >0
--
-- __ AFTER HAND-APPLYING: _HOLD files never reach the migration ledger, so the
--    record is THIS FILE - append one line directly below this block and commit
--    it (pathspec):  -- APPLIED <date> by <session>; roll verified at <stamp sha>
--
-- __ RECORD THE OWNER'S READ (gate 2): a dated line in the apis lane's NOTES
--    quoting his words, and name it in this file's APPLIED line - an unrecorded
--    read is indistinguishable from a skipped one.

-- APPLIED 2026-09-03 19:3xZ by the framework_prompts_positive_voice session (which
--   holds this apply by the apis.uk lane's handover of the same day). Gate 2 CLEARED
--   the same evening: the owner was shown these exact INSERTED TEXT bytes plus one
--   filled-in render and answered "yes"; his read is recorded verbatim, with the bytes
--   and with what changed since round 1, in
--   docs/agent_docs/docs024_key_docs_latest/apis_uk_bees_homepage/NOTES_apis_uk_bees_homepage.md
--   (entry dated 2026-09-03 19:2xZ). Gate 1 re-verified immediately before applying, on
--   BOTH replicas, three-way: section_subjects 3 (capability), section_facts 3 (positive
--   control), zzz_absent_zzz 0 (absent control, so the probe discriminates).
--   Apply output: "NOTICE: 641 applied: block + input_fields in one transaction;
--   em-dash census 10 (unchanged)". Live row verified after COMMIT: block present, the
--   old frame absent, sibling range present, block precedes the Verified Facts block,
--   input_fields carries sections_for_render, em dashes 10, template 14,914 chars.
--   Council: round 1 APPROVED (corr 6c92d154); round 2 resubmitted on the same
--   correlation for the A4 redraft and still open at apply time - the gate is advisory
--   and the owner's approval of the exact words is the binding one.

SELECT snapshot_agent('page-content-writer', '641_page_content_writer_prompt_v5_section_subject_HOLD.sql: pre-update (redraft C)');

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_641 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'page-content-writer' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ONE plpgsql block for pre-flight, update and verify, so the pre-update
-- em-dash count carries in a VARIABLE and the census asserts EQUALITY.
-- The first cut asserted a literal (5) measured 2026-08-26; the live count
-- was 9 by 2026-09-03 (rules 9/10 text, mig 595, and the row was updated
-- 2026-09-03 08:56Z with NO agent_snapshots row) - a literal census fails
-- the apply for someone else's edit, while the real invariant is only that
-- THIS insertion adds no em dashes.
DO $$
DECLARE
    t text;
    t2 text;
    ifields jsonb;
    pre_dashes int; post_dashes int;
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
    -- Guardian advisory (council 6c92d154): the estate has documented cases of a
    -- type carrying two active rows where only the higher version loads. Today
    -- page-content-writer has ONE row total (v2, measured 2026-09-03); this
    -- assertion keeps the apply refusing if that ever changes underneath.
    PERFORM 1 FROM agent_definitions
    WHERE type = 'page-content-writer' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
      AND version = (SELECT max(version) FROM agent_definitions
                     WHERE type = 'page-content-writer' AND deleted_at IS NULL);
    IF NOT FOUND THEN
        RAISE EXCEPTION '641: the active page-content-writer row is not the max version - a higher-version row would shadow this edit; investigate before applying';
    END IF;
    SELECT default_config->'workflow'->'steps'->'process_sections_loop'->'config'
           ->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template',
           default_config#>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,input_fields}'
    INTO t, ifields
    FROM agent_definitions
    WHERE type = 'page-content-writer' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF t IS NULL THEN
        RAISE EXCEPTION '641: page-content-writer generate_content prompt_template not found';
    END IF;
    IF ifields IS NULL OR jsonb_typeof(ifields) <> 'array' THEN
        RAISE EXCEPTION '641: generate_content input_fields not found or not an array';
    END IF;
    IF position('{{if .current_section.subject}}' in t) > 0 THEN
        RAISE EXCEPTION '641: already applied - subject block present';
    END IF;
    IF ifields ? 'sections_for_render' THEN
        RAISE EXCEPTION '641: input_fields already contains sections_for_render - half-applied state or double run; investigate before proceeding';
    END IF;
    c1 := (length(t) - length(replace(t, '{{if .current_section.facts_scoped}}', ''))) / length('{{if .current_section.facts_scoped}}');
    IF c1 <> 1 THEN
        RAISE EXCEPTION '641: live prompt has drifted (facts_scoped anchor count % not 1) - re-derive from the live row', c1;
    END IF;
    pre_dashes := length(t) - length(replace(t, '—', ''));

    -- The delta: (a) the subject block before the Verified Facts block,
    -- (b) sections_for_render appended to input_fields. Both or neither.
    UPDATE agent_definitions
    SET default_config = jsonb_set(
            jsonb_set(
                default_config,
                '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
                to_jsonb(
                    replace(
                        default_config->'workflow'->'steps'->'process_sections_loop'->'config'
                            ->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template',
                        '{{if .current_section.facts_scoped}}',
                        E'{{if .current_section.subject}}## This section\n\n{{.current_section.subject}}\n\n{{.current_page.title}} also covers, each in its own section:\n{{range $s := .sections_for_render.sections_ready}}{{if and $s.subject (ne $s.subject $.current_section.subject)}}- {{$s.subject}}\n{{end}}{{end}}\n{{end}}{{if .current_section.facts_scoped}}'
                    )
                )
            ),
            '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,input_fields}',
            (default_config#>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,input_fields}')
                || '["sections_for_render"]'::jsonb
        ),
        updated_at = NOW()
    WHERE type = 'page-content-writer' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- Verify, same transaction.
    SELECT default_config->'workflow'->'steps'->'process_sections_loop'->'config'
           ->'sub_workflow'->'steps'->'generate_content'->'config'->>'prompt_template',
           default_config#>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,input_fields}'
    INTO t2, ifields
    FROM agent_definitions
    WHERE type = 'page-content-writer' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position('{{if .current_section.subject}}' in t2) = 0
       OR position('## This section' in t2) = 0
       OR position('{{range $s := .sections_for_render.sections_ready}}' in t2) = 0
       OR position('{{if .current_section.facts_scoped}}' in t2) = 0
       OR position('assigned to THIS section' in t2) = 0
       OR position('{{.voice_style}}' in t2) = 0 THEN
        RAISE EXCEPTION '641: verify failed - the subject block, the sibling range, the facts block it must precede, or a preserved landmark is missing';
    END IF;
    IF position('{{if .current_section.subject}}' in t2) >= position('{{if .current_section.facts_scoped}}' in t2) THEN
        RAISE EXCEPTION '641: verify failed - subject block does not precede the Verified Facts block';
    END IF;
    -- BOTH halves or neither: a template that ranges a key the extractor never
    -- copies renders an empty list with no error (fixture D).
    IF NOT (ifields ? 'sections_for_render') THEN
        RAISE EXCEPTION '641: verify failed - input_fields does not contain sections_for_render; the sibling range would render EMPTY silently';
    END IF;
    post_dashes := length(t2) - length(replace(t2, '—', ''));
    IF post_dashes <> pre_dashes THEN
        RAISE EXCEPTION '641: verify failed - em-dash count changed % -> %; the insertion touched more than the one anchor', pre_dashes, post_dashes;
    END IF;
    RAISE NOTICE '641 applied: block + input_fields in one transaction; em-dash census % (unchanged)', post_dashes;
END $$;

COMMIT;

-- ROLLBACK recipe (hand-run): restore from agent_definitions_bak_641, or:
--   (a) the inserted block is exactly the text between the INSERTED TEXT
--       markers above - replace it with '' in the same jsonb_set shape;
--   (b) remove the appended element:
--       input_fields := (SELECT jsonb_agg(e) FROM jsonb_array_elements(input_fields) e
--                        WHERE e <> '"sections_for_render"'::jsonb)
