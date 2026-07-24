-- 201_content_writers_never_invent_numbers.sql — bug_open/043 fix candidate 3:
-- the scalar no-unsourced-figures rule, bound where the fabrication actually
-- happens. DB-only; effective immediately, no image roll.
--
-- BUG 043 (robot-hands.com and fleet-wide): ordinary content generation writes
-- specific numbers into content_data with no source ("1,200+ Gripper Models
-- Indexed" against an index of 5; "14,203 Takes Filed Today" with no takes
-- table; "70 Deployed Agents" nothing counts). The 2026-07-24 sweep found the
-- class live on 5 sites and — decisively — caught a LIVE RECURRENCE: a routine
-- needs_page re-render of robot-hands/index that morning RE-INVENTED
-- "2,400+ Gripper Models Indexed" in a stat block that had been hand-corrected
-- to sourced values four days earlier.
--
-- WHY THE EXISTING RULE LOSES. page-content-writer's section prompt already
-- says (rule 14) "NEVER invent specific statistics...". But the "What To Write"
-- block lists each llm field with its schema description, and a stat component's
-- fields say "stat_1_value (text, required): The numeric value for the first
-- stat. Use short-form where appropriate, e.g. '99.99', '2.4M', '150'". Faced
-- with a REQUIRED field whose description demands a number and example shapes
-- to copy, and no data anywhere in the prompt, the model resolves the conflict
-- by inventing — the "2,400+" it wrote is literally the '2.4M' example shape.
-- Rule 14 forbids the invention but offers no legal alternative for a required
-- numeric field, so it loses. The fix names the conflict and gives the
-- alternative: return an empty string. (An empty stat renders as nothing —
-- suffix fallbacks were cleared fleet-wide 2026-07-22; an invented stat
-- publishes a false claim.)
--
-- Two agents patched:
--   page-content-writer (sub-workflow section prompt, rule 14 rewritten) — the
--     path that produced every confirmed instance, including today's recurrence.
--   content-writer (template-placeholder filler, Guidelines block) — same
--     exposure via content_requirements placeholders; one rule line added.
--
-- Candidates 1 (provenance-bound stat fields) and 2 (post-generation numeric
-- audit) remain open in bugs_open/043 — this is the prompt-side stop.
--
-- Anchors verified unique against the LIVE rows 2026-07-24 (rule-14 sentence:
-- exactly 1 occurrence; content-writer Guidelines head: 1 occurrence).
-- Standing rule: snapshot_agent opens the transaction; the DO block fails the
-- whole transaction if any anchor did not match.

BEGIN;

SELECT snapshot_agent('page-content-writer', '201_content_writers_never_invent_numbers.sql: pre-update');
SELECT snapshot_agent('content-writer',      '201_content_writers_never_invent_numbers.sql: pre-update');

-- (1) page-content-writer: rewrite rule 14 in the section-generation prompt.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        '14. NEVER invent specific statistics, percentages, or metrics. Describe the type of outcome without fabricating numbers.',
        $r14$14. NEVER invent specific statistics, percentages, or metrics. Never state a count, total, price, rating, or coverage figure you have not been given in THIS prompt (Verified Facts, Research Findings, Admin Content Brief, or Existing Content). This applies with full force to numeric stat fields (field names like stat_1_value): a field being required, or its description asking for a numeric value with example shapes, is NOT permission to invent one — if no given figure fits, return an empty string for that field. An empty stat renders as nothing; an invented stat publishes a false claim on a live site. Describe the type of outcome without fabricating numbers.$r14$
      ))
    ),
    updated_at = now()
WHERE type = 'page-content-writer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- (2) content-writer: add the scalar rule at the head of its Guidelines block.
UPDATE agent_definitions
SET default_config = jsonb_set(
      default_config,
      '{workflow,steps,generate_content,config,prompt_template}',
      to_jsonb(replace(
        default_config #>> '{workflow,steps,generate_content,config,prompt_template}',
        'Guidelines:
- Write compelling, conversion-focused copy',
        $cwg$Guidelines:
- NEVER invent counts, statistics, percentages, prices, or metrics: state a figure only if it appears in the details, strategy, or requirements you were given; otherwise write without a number (leave a numeric placeholder empty rather than inventing a value)
- Write compelling, conversion-focused copy$cwg$
      ))
    ),
    updated_at = now()
WHERE type = 'content-writer' AND is_active
  AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

-- Verify: one live row each, new text present exactly once, old text gone.
DO $$
DECLARE pt text; ct text; n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
      WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n <> 1 THEN RAISE EXCEPTION 'page-content-writer: expected exactly one live row, found %', n; END IF;
    SELECT count(*) INTO n FROM agent_definitions
      WHERE type='content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n <> 1 THEN RAISE EXCEPTION 'content-writer: expected exactly one live row, found %', n; END IF;

    SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
      INTO pt FROM agent_definitions
      WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF position('is NOT permission to invent one' in pt) = 0 THEN
        RAISE EXCEPTION 'page-content-writer: new rule 14 missing (anchor did not match)';
    END IF;
    IF (length(pt)-length(replace(pt,'is NOT permission to invent one','')))/length('is NOT permission to invent one') <> 1 THEN
        RAISE EXCEPTION 'page-content-writer: new rule 14 not present exactly once';
    END IF;
    IF position('or metrics. Describe the type of outcome' in pt) <> 0 THEN
        RAISE EXCEPTION 'page-content-writer: old rule 14 sentence still present';
    END IF;

    SELECT default_config #>> '{workflow,steps,generate_content,config,prompt_template}'
      INTO ct FROM agent_definitions
      WHERE type='content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF position('leave a numeric placeholder empty rather than inventing a value' in ct) = 0 THEN
        RAISE EXCEPTION 'content-writer: scalar rule missing (Guidelines anchor did not match)';
    END IF;

    RAISE NOTICE '201 applied: scalar no-unsourced-figures rule live in page-content-writer (rule 14) and content-writer (Guidelines)';
END $$;

COMMIT;
