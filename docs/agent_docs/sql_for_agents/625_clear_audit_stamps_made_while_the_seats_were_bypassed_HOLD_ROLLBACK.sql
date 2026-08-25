-- 625_..._ROLLBACK.sql — deliberately a GUARDED NO-OP with the pointer.
-- Restoring these stamps would re-assert "audited" over sites nothing judged —
-- the exact record 625 exists to remove. If a restoration is ever genuinely
-- wanted, the verbatim stamps are preserved in the doc_notes row
-- (source='migration-625', subject_key='improvement-loop'); restore by hand
-- from that JSON, per site, saying why.
DO $$
DECLARE v_id uuid;
BEGIN
    SELECT id INTO v_id FROM doc_notes WHERE source='migration-625' ORDER BY created_at DESC LIMIT 1;
    IF v_id IS NULL THEN RAISE EXCEPTION '625 ROLLBACK: 625 was never applied'; END IF;
    RAISE NOTICE '625 ROLLBACK: no-op by design. The cleared stamps are preserved verbatim in doc_notes row % — restore by hand only with a stated reason.', v_id;
END $$;
