-- 417_brief_fidelity_auditor_speaks_the_routers_vocabulary_ROLLBACK.sql
-- Hand-run sidecar (uppercase suffix: excluded from run-migrations.sh --apply).
--
-- Restores the single hardcoded brief_fidelity category and removes the
-- repair-shape section 417 inserted. Before running, be sure: with the old
-- category restored, 100% of the auditor's output routes to the unknown-category
-- path again (capability_gap post-d6d56e540; MINTED unrouteable rows on any
-- older binary — bugs_open/279). If 418 has been applied, roll IT back first or
-- every sweep files roadmap noise.

BEGIN;

DO $$
DECLARE
    occ_vocab integer;
    occ_old integer;
BEGIN
    SELECT (length(default_config::text) - length(replace(default_config::text,'choose by REPAIR SHAPE','')))/length('choose by REPAIR SHAPE'),
           (length(default_config::text) - length(replace(default_config::text,'brief_fidelity','')))/length('brief_fidelity')
      INTO occ_vocab, occ_old
    FROM agent_definitions
    WHERE type='brief-fidelity-auditor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF occ_old <> 0 OR occ_vocab <> 1 THEN
        RAISE EXCEPTION 'ROLLBACK 417: unexpected state (brief_fidelity=%, repair-shape=%) — 417 is not cleanly applied; reconcile by hand', occ_old, occ_vocab;
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = replace(replace(default_config::text,
        '\"category\":\"colour|spacing|typography|responsive|cta|nav_restructure|gap|content|tone|structure|differentiation\"',
        '\"category\":\"brief_fidelity\"'),
        '## Category — choose by REPAIR SHAPE\nThe category field decides WHO repairs the finding, so pick by what the fix IS, not by why you found it (your brief-fidelity identity travels in audit_source automatically):\n- colour / spacing / typography / responsive → design and CSS fixers (visual promises: styling, layout polish, breakpoints)\n- cta / nav_restructure → component fixer (calls-to-action, navigation promises)\n- gap → missing promised content (a page or section the brief implies and the site lacks)\n- content / tone / structure / differentiation → content rebuild or planning (register, copy substance, page composition promises)\nUse EXACTLY one of those strings. If a broken promise genuinely fits none of them, use a short snake_case label of your own: it will surface as a capability-gap roadmap entry for a human rather than being dispatched. The page field must be an EXACT page name from the inventory above, or site-wide.\n\nRespond with ONLY a JSON array',
        'Respond with ONLY a JSON array')::jsonb,
    version    = version + 1,
    updated_at = now()
WHERE type='brief-fidelity-auditor' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
    occ_old integer;
BEGIN
    SELECT (length(default_config::text) - length(replace(default_config::text,'brief_fidelity','')))/length('brief_fidelity')
      INTO occ_old
    FROM agent_definitions
    WHERE type='brief-fidelity-auditor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF occ_old <> 1 THEN
        RAISE EXCEPTION 'ROLLBACK 417: expected brief_fidelity restored exactly once, found %', occ_old;
    END IF;
    RAISE NOTICE 'rollback 417 OK: single brief_fidelity category restored';
END $$;

COMMIT;
