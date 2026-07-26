-- bugs_open/043 — fix a drift I introduced myself in migration 218, caught the same
-- evening by the gate it was meant to arm.
--
-- ── WHAT HAPPENED, IN ORDER ────────────────────────────────────────────────
-- 218 seeded aao's evidence_base `facts[]`, each carrying `source.sql` so
-- refresh_evidence_base can keep the value current, and deliberately left
-- `writer_block` HAND-managed (composeWriterBlock has no NEVER-STATE section, so
-- managed regeneration would silently delete the ban list).
--
-- 18:19  218 applied. Fact aao-orchestrations = 1834; writer_block prose says
--        "over a thousand orchestrations a day (1,834 in the 24 hours to 2026-07-26)".
-- 18:34  evidence-refresher re-ran the fact's SQL. The 24-hour window had rolled and
--        rows had been pruned, so the fact fell to 1790. The writer_block was NOT
--        touched — by my own design decision.
-- 19:07  A full-writer rebuild of /index took "1,834" from the writer_block, exactly as
--        instructed, and wrote it into two components.
-- 19:10  validate_page_content rejected the page: 3 errors, all the same figure —
--        one `unregistered_number` (check 8, HTML) and two `unregistered_stat`
--        (check 9, content_data). `numberSupported(1834)` against a fact of 1790 with
--        tolerance `gte` is false, because 1834 > 1790. **The gate was right.**
--
-- So: leaving the block hand-managed while the facts auto-refresh guarantees drift for
-- any fact whose value can FALL. The writer is told one number and the gate checks
-- another. That is a defect I built, and it would have fired on every refresh cycle.
--
-- ── THE DISCRIMINATOR: WHICH FIGURES CAN FALL ─────────────────────────────
-- Not all of them, and the fix should not be blunter than the problem.
--   * MONOTONIC-ish counts (active agent definitions, distinct agent types, live sites):
--     these only grow, so a dated snapshot in the prose stays <= the live fact and
--     tolerance `gte` keeps supporting it. Those lines are SAFE and stay as they are —
--     they are also what produced the good outcome at 19:07, where the writer used
--     175 / 14 / 17 from this block and left the fifth stat honestly empty.
--   * WINDOWED or REAPED metrics: orchestrations-per-24h rolls; work-items-completed is
--     reaped (it already fell 1,267 -> 1,051 -> lower). Any absolute snapshot of these in
--     the prose will eventually exceed the fact and be correctly rejected.
--
-- So only the two volatile lines lose their absolute figure. They keep a qualitative
-- form the writer can still use ("over a thousand orchestrations a day"), which is true
-- across the whole plausible range and cannot drift out of support.
--
-- Idempotent: sets the block outright.

\set ON_ERROR_STOP on

BEGIN;

DO $wb$
DECLARE
    n int;
    new_block CONSTANT text :=
$ao$NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine — these are live registry counts taken 2026-07-26, recount before restating):
- more than 150 active agent definitions in the production registry (175 as of 2026-07-26); "over 70 specialised AI agents" is also true and may stand where already written
- 8 departments — Strategy, Research, Content, Design, Development, Quality, Operations, Data. This is the platform's OWN organisational taxonomy (a self-description); never frame it as departments of external clients ("departments served")
- more than 150 distinct agent types (174 as of 2026-07-26); "30+ agent types" is also true and may stand
- 14 live sites in production, built and operated end-to-end by the platform
- 17 backend services (the service manifests under deployments/kustomize/services)
- orchestrations per day: "over a thousand orchestrations a day" is true and may be stated. DO NOT state an exact daily figure — it is a ROLLING 24-hour window, so any number you write is stale within hours and will be rejected as unsupported.
- automated work items completed: DO NOT state a figure. The ledger is reaped, so this count FALLS as well as rises (1,267 on 07-24, 1,051 on 07-26); a total that can go down is misleading however it is phrased.
- Architecture: Kubernetes, Kafka, Postgres — true and stated freely
NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, "departments served", satisfaction rates, awards won, concurrent-instance counts ("thousands of concurrent instances" is not measured), uptime percentages. None of these are measured; every such figure at any value is an invention.$ao$;
BEGIN
    UPDATE site_specs
       SET data = data || jsonb_build_object('writer_block', new_block),
           updated_at = now()
     WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
       AND aspect = 'evidence_base' AND is_current;

    GET DIAGNOSTICS n = ROW_COUNT;
    IF n <> 1 THEN
        RAISE EXCEPTION '043: expected 1 aao evidence_base row, updated %', n;
    END IF;

    -- The ban list is the half that stops a whole CATEGORY of invention. Assert it
    -- survived, because losing it is the exact failure this file's design avoids.
    SELECT count(*) INTO n FROM site_specs
    WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current
      AND data->>'writer_block' LIKE '%NEVER STATE%'
      AND jsonb_array_length(COALESCE(data->'facts','[]'::jsonb)) = 7;
    IF n <> 1 THEN
        RAISE EXCEPTION '043: the NEVER-STATE list or the facts array did not survive';
    END IF;

    RAISE NOTICE '043: aao writer_block — volatile figures made qualitative; monotonic snapshots kept; bans intact';
END $wb$;

COMMIT;
