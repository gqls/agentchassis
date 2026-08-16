-- Record that Pieces 2/3 of PLAN_2026-08-09_facts_into_tool_acceptance.md landed,
-- against BOTH SDLT tool subject keys, so the next session touching either fence
-- finds the design decisions on the travelling-docs trail instead of re-deriving
-- them from the diff.
--
-- Raised by the council's tooling_provenance seat (2026-08-16, severity low):
-- the change reads the travelling PLAN and the tool's own criteria fence before
-- acting, but wrote nothing back. This is the other half of that convention.
--
-- Idempotent: the guard refuses to insert a second copy for the same subject key.
-- Safe to apply before or after the chassis roll — it is a note, not config.

BEGIN;

DO $$
DECLARE
  v_subject text;
  v_body    text;
BEGIN
  v_body :=
'PIECES 2+3 of PLAN_2026-08-09_facts_into_tool_acceptance.md LANDED 2026-08-16 (commits 989addb1c, 3022c1dfe; council cff364b8; register CLM-022 + TL-045; class file bugs_open/288).

WHAT THIS FENCE CAN NOW DECLARE: a fence-level "facts": ["<fact_id>", ...] naming which evidence-register facts this tool encodes. IDS ONLY, never values — doc_plans has no site_id, so values are resolved from the driven site''s register at check time (PBP-037). Validator rule P11 refuses a malformed list where it is written. NEITHER acceptance tier reads the key: a green run on a fence carrying facts does NOT mean the figures were compared.

WHAT HAPPENS WHEN A DECLARED FACT MOVES: the daily evidence-freshness sweep files one work item per (fact, tool). improve_tool (tool-improver) ONLY when a tool-level component owns this page''s code AND this fence has no no_auto_fix; otherwise fact_drift_review, handler-less, for a human. Evidence-only drift (lost citation, changed artefact) is ALWAYS human — a 404 at GOV.UK is not evidence a figure moved. A fetch error files nothing; the citation arm escalates it on staleness_days instead.

THREE BANDS: value_drift high/30, evidence_drift medium/35, unreconciled_declaration low/60.

FIRST DECLARATION FILES ONE ITEM PER FACT. A fence declaring 13 facts produces 13 low-severity reconciliation items on its first sweep, then goes quiet (each item records the value, which becomes the baseline). This exists because the council found the mechanism would otherwise be blind to its own motivating bug: a tool stale on the day it opts in, against a fact that has not moved since, produced no baseline difference and emitted nothing — which is bugs_closed/225 exactly.

MEASURED 2026-08-16, and it changed the design: the fan-out does NOT key on the acceptance ladder''s toolEligibilityWhere. That predicate admits NEITHER SDLT tool (both are multi-component pages), so an eligibility-keyed check could never have fired on the tools that motivated it and would have read green for ever. Encoding a fact and being acceptance-eligible are different questions. The page predicate is the AUDIT one (skip only archived AND never deployed), never status=''active''.

STILL DEFERRED, stated rather than silent: Piece 4 (an oracle computing expectations from the register) is behind its own RFC. This mechanism answers "did the figure MOVE", never "is the figure RIGHT" — a tool and a register wrong in the same direction agree, and it is silent.';

  FOREACH v_subject IN ARRAY ARRAY['stamp-duty', 'mortgages-stamp-duty'] LOOP
    IF EXISTS (SELECT 1 FROM doc_notes
                WHERE subject_type = 'tool' AND subject_key = v_subject
                  AND categories ? 'fact-drift' AND body LIKE 'PIECES 2+3%') THEN
      RAISE NOTICE 'doc_note already present for %, skipping', v_subject;
    ELSE
      INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
      VALUES ('tool', v_subject, v_body,
              '["fact-drift","claims-verification","tool-lifecycle"]'::jsonb,
              'manual', 'register_guards_code_phase_b:bugs_open/288');
      RAISE NOTICE 'doc_note written for %', v_subject;
    END IF;
  END LOOP;
END $$;

-- Verify inside the transaction: RAISE on a wrong count, because a verify block of
-- bare SELECTs cannot stop a COMMIT (ON_ERROR_STOP ignores a non-empty result).
DO $$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM doc_notes
   WHERE subject_type='tool' AND subject_key IN ('stamp-duty','mortgages-stamp-duty')
     AND categories ? 'fact-drift';
  IF n <> 2 THEN
    RAISE EXCEPTION 'expected exactly 2 fact-drift doc_notes, found %', n;
  END IF;
END $$;

COMMIT;
