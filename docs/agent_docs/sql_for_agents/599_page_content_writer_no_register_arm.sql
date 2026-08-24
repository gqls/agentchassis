-- 599_page_content_writer_no_register_arm.sql (was _HOLD until the owner read the v5 plaintext)
--
-- *** OWNER APPROVED the v5 plaintext 2026-08-24 evening ("approved") — released from _HOLD and applied the same evening. ***
-- RFC_016 §5.2: the owner's 2026-08-09 approval of the v4 writer prompt attaches to that
-- text; any later edit voids it and needs a fresh read (compliance seat's round-1 ask).
-- The full post-598 plaintext is committed for that read at:
--   docs/agent_docs/docs024_key_docs_latest/brochure_component_library/sql/
--   page_content_writer_prompt_v5_2026-08-24.txt
-- To apply after approval: rename away the _HOLD suffix, psql -f, then --record-only.
--
-- bugs_open/380 slice S1 (writer half). Three defects in the live v4 prompt on a site with
-- no evidence register:
--   1. The scoped-empty arm says "other sections on this site carry them" — FALSE on a
--      register-less site (nothing carries them anywhere), and the false clause invites the
--      writer to allude to facts that exist nowhere.
--   2. There is NO no-register arm at all: with no evidence_base the whole Verified Facts
--      block renders empty and the writer is told nothing.
--   3. The STRICT accuracy rule says "If the section calls for a statement about method,
--      say what we DO" — on a methodology page of a business with no operating history that
--      is a licence for present-tense practice prose ("we buy the tool at the same price a
--      reader would pay"). Owner ruling 2026-08-24: aspirations must not be stated as
--      present-tense practice; say only what is sourced.
--
-- A prompt instruction is not a control on output (house rule) — the controls are 597's
-- cold audit and the Go practice-claims family. This is the prevention half: stop INVITING
-- the defect. Structure: each new conditional is nested one level at a time because
-- text/template under missingkey=zero ERRORS on a dotted path through an absent
-- intermediate map ({{if .a.b}} with no .a crashes the render) — the same reason 330
-- guards each level. `.site_specs.specs.evidence_base.operating_history` is a plain map
-- lookup on an aspect the template already receives, NOT a new input_fields variable.
--
-- Anchors verified against the live template 2026-08-24 (each exactly once) at path
-- {workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,
-- prompt_template}. Em-dash census and {{.}}-render count asserted unchanged; if/else/end
-- balance asserted +3/+3/+3.

BEGIN;

DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN
    RAISE EXCEPTION '599 REFUSED: expected exactly 1 active page-content-writer row, found %', n;
  END IF;
  PERFORM snapshot_agent('page-content-writer',
                         '599_page_content_writer_no_register_arm.sql: pre-update');
END $$;

DO $do$
DECLARE
  tpl text; newtpl text; n int;
  ifs int; ends int; elses int; renders int; emdash int;
  anchor1 text := $A$State NO business numbers, counts, statistics, or named-entity relationship claims in this section - other sections on this site carry them. Write this section's purpose without site-specific figures.$A$;
  repl1 text := $R$State NO business numbers, counts, statistics, or named-entity relationship claims in this section{{if .site_specs.specs.evidence_base}} - other sections on this site carry them{{else}} - this site has NO verified-facts register, so nothing carries them anywhere{{end}}. Write this section's purpose without site-specific figures.$R$;
  anchor2 text := E'{{end}}{{end}}{{end}}\n\n{{if .rewrite_guidance}}';
  block text := $B$## Operating history: NONE RECORDED
This business has no recorded operating history. Do not state, in any tense, that we test, trial, buy, purchase, weigh, measure, record, inspect, use, receive or are sent products or samples, visit, interview or survey anyone, or garden, cook or build ourselves. Do not describe a review or assessment method as something we do. Where the brief asks for method, say what the site does WITH SOURCES: name the manufacturer specification, published standard or retailer listing a figure comes from, and date it. An intention may be stated as an intention only when the brief supplies it; otherwise leave it out. This applies to FAQ answers exactly as it applies to body copy.$B$;
  repl2 text;
  anchor3 text := 'say what we DO -- we name our sources and their dates';
  repl3 text := 'say what we DO -- and with no recorded operating history that means ONLY how the content is sourced: we name our sources and their dates';
BEGIN
  repl2 := E'{{end}}{{end}}{{end}}\n{{if .site_specs.specs.evidence_base}}{{if .site_specs.specs.evidence_base.operating_history}}{{else}}\n' || block || E'\n{{end}}{{else}}\n' || block || E'\n{{end}}\n\n{{if .rewrite_guidance}}';

  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO tpl FROM agent_definitions
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF tpl IS NULL THEN RAISE EXCEPTION '599: writer prompt_template not found at the loop path'; END IF;

  n := (length(tpl) - length(replace(tpl, anchor1, ''))) / length(anchor1);
  IF n <> 1 THEN RAISE EXCEPTION '599: anchor1 (scoped-empty arm) found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor2, ''))) / length(anchor2);
  IF n <> 1 THEN RAISE EXCEPTION '599: anchor2 (block close) found % times, expected 1', n; END IF;
  n := (length(tpl) - length(replace(tpl, anchor3, ''))) / length(anchor3);
  IF n <> 1 THEN RAISE EXCEPTION '599: anchor3 (say what we DO) found % times, expected 1', n; END IF;

  ifs    := (length(tpl) - length(replace(tpl, '{{if ', ''))) / length('{{if ');
  ends   := (length(tpl) - length(replace(tpl, '{{end}}', ''))) / length('{{end}}');
  elses  := (length(tpl) - length(replace(tpl, '{{else}}', ''))) / length('{{else}}');
  renders := length(tpl) - length(replace(tpl, '{{.', ''));
  emdash  := length(tpl) - length(replace(tpl, '—', ''));

  newtpl := replace(tpl, anchor1, repl1);
  newtpl := replace(newtpl, anchor2, repl2);
  newtpl := replace(newtpl, anchor3, repl3);

  IF length(newtpl) <> length(tpl)
       + (length(repl1) - length(anchor1))
       + (length(repl2) - length(anchor2))
       + (length(repl3) - length(anchor3)) THEN
    RAISE EXCEPTION '599: unexpected length delta';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{if ', ''))) / length('{{if ') <> ifs + 3
     OR (length(newtpl) - length(replace(newtpl, '{{end}}', ''))) / length('{{end}}') <> ends + 3
     OR (length(newtpl) - length(replace(newtpl, '{{else}}', ''))) / length('{{else}}') <> elses + 3 THEN
    RAISE EXCEPTION '599: if/else/end balance not +3/+3/+3';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '{{.', ''))) <> renders THEN
    RAISE EXCEPTION '599: a {{.}} render was introduced — would need input_fields';
  END IF;
  IF (length(newtpl) - length(replace(newtpl, '—', ''))) <> emdash THEN
    RAISE EXCEPTION '599: em-dash census changed (330 pins it)';
  END IF;

  UPDATE agent_definitions
     SET default_config = jsonb_set(default_config,
           '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
           to_jsonb(newtpl), false),
         updated_at = now()
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '599: updated % rows, expected exactly 1', n; END IF;
END $do$;

DO $$
DECLARE tpl text;
BEGIN
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO tpl FROM agent_definitions
   WHERE type='page-content-writer' AND is_active
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF position('## Operating history: NONE RECORDED' in tpl) = 0 THEN
    RAISE EXCEPTION '599 VERIFY: operating-history block missing';
  END IF;
  IF position('so nothing carries them anywhere' in tpl) = 0 THEN
    RAISE EXCEPTION '599 VERIFY: scoped-empty conditional clause missing';
  END IF;
  IF position('ONLY how the content is sourced' in tpl) = 0 THEN
    RAISE EXCEPTION '599 VERIFY: method-rule qualification missing';
  END IF;
  RAISE NOTICE '599 OK: writer no-register arm live.';
END $$;

COMMIT;
