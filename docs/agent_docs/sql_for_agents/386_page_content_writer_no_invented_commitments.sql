-- 386 - page-content-writer: STRICT RULE 5 gains the invented-commitments ban
-- (COMPLIANCE_READ_2026-08-10 finding 3.4; owner APPROVED, DECISIONS_2026-08-11
-- ruling 4: "we don't want these claims").
--
-- Why: rule 5 bans invented achievements/metrics/outcomes and rule 14 bans
-- invented figures, but nothing squarely bans figure-free COMMITMENTS - a
-- money-back guarantee, "free no-obligation consultation", "we'll always pick
-- up the phone". The writer can invent a service promise without tripping any
-- existing rule. One sentence closes it, appended to rule 5 so the ban sits
-- with its family (the invention bans), not in a new rule number that would
-- renumber nothing but still change every cross-reference.
--
-- One anchored edit, derived byte-exact from the live row at the NESTED path
-- (the v4 prompt lives inside process_sections_loop's sub_workflow - the
-- 08-10 REVISE round's editquality seat confirmed the path; seed 330 pattern).
-- Do NOT regenerate from the v4 file in the brochure sql dir - it is one
-- drift behind live the moment anything else touches the row (handoff 08-11).

SELECT snapshot_agent('page-content-writer', '386_page_content_writer_no_invented_commitments.sql: pre-update');

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_386 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'page-content-writer' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $do$
DECLARE
    t text;
    c1 int;
BEGIN
    SELECT default_config#>>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO t
    FROM agent_definitions
    WHERE type = 'page-content-writer' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF t IS NULL THEN
        RAISE EXCEPTION '386: page-content-writer nested prompt_template not found';
    END IF;
    -- Dual-active-row landmine guard (council round d1e8c36e objection): four
    -- agent types carry TWO active rows and only the higher version loads.
    -- Refuse unless page-content-writer has EXACTLY ONE active non-snapshot row, so
    -- "UPDATE 1 + verify passed" cannot describe a row the loader never reads.
    IF (SELECT count(*) FROM agent_definitions
         WHERE type = 'page-content-writer' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL) <> 1 THEN
        RAISE EXCEPTION '386: page-content-writer does not have exactly one active row - resolve the duplicate before seeding';
    END IF;
    IF position('invent commitments' in t) > 0 THEN
        RAISE EXCEPTION '386: already applied - the commitments ban is present';
    END IF;

    c1 := (length(t) - length(replace(t, $a1$5. Write content that is relevant to this company's industry and services, but do NOT invent specific achievements, metrics, or outcomes that have not actually happened.$a1$, ''))) / length($a1$5. Write content that is relevant to this company's industry and services, but do NOT invent specific achievements, metrics, or outcomes that have not actually happened.$a1$);
    IF c1 <> 1 THEN
        RAISE EXCEPTION '386: rule-5 anchor count must be exactly 1, got % - the live prompt has drifted; regenerate from a fresh dump', c1;
    END IF;
END $do$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
        to_jsonb(
            replace(
                default_config#>>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}',
                $a1$5. Write content that is relevant to this company's industry and services, but do NOT invent specific achievements, metrics, or outcomes that have not actually happened.$a1$,
                $r1$5. Write content that is relevant to this company's industry and services, but do NOT invent specific achievements, metrics, or outcomes that have not actually happened. Do NOT invent commitments, guarantees, warranties, or service promises (response times, refunds, availability) not stated in this prompt.$r1$)
        )
    ),
    updated_at = NOW()
WHERE type = 'page-content-writer' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify with DO/RAISE - a verify block of SELECTs cannot stop the COMMIT.
DO $do$
DECLARE
    t text;
BEGIN
    SELECT default_config#>>'{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}'
    INTO t
    FROM agent_definitions
    WHERE type = 'page-content-writer' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF position($v1$Do NOT invent commitments, guarantees, warranties, or service promises$v1$ in t) = 0 THEN
        RAISE EXCEPTION '386: verify failed - the commitments ban is missing';
    END IF;
    IF length(t) <> 13974 THEN
        RAISE EXCEPTION '386: verify failed - post length % <> expected 13974', length(t);
    END IF;
END $do$;

COMMIT;

-- ROLLBACK recipe (hand-run):
--   UPDATE agent_definitions ad
--   SET default_config = b.default_config
--   FROM agent_definitions_bak_386 b
--   WHERE ad.id = b.id;
