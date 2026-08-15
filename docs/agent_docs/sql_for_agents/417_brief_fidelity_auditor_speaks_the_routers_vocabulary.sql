-- 417_brief_fidelity_auditor_speaks_the_routers_vocabulary.sql
--
-- OWNER DECISION 2026-08-15 (bugs_open/279 candidate 3): make brief-fidelity-auditor
-- a real working check. This is half one — the vocabulary; 418_HOLD is half two (the
-- improvement-loop wiring, held until the routing fix rolls).
--
-- WHY THE VOCABULARY AND NOT A NEW ROUTE. The auditor hardcoded
-- "category":"brief_fidelity" — a value in NO routing set, so 100% of its output
-- took write_audit_findings' minting fallback and died (bugs_open/115/279). The
-- first fix sketch was a Go route mapping brief_fidelity → content_rewrite. The
-- auditor's own four real findings from 2026-08-13 refuted that: three of four were
-- DESIGN-intent violations ("Animations beyond hover states", "Rounded corners
-- beyond 12px"), which a content rebuild cannot repair. A single bespoke category
-- cannot say what the FIX is, and classifyFinding routes on exactly that.
--
-- So instead the auditor speaks the router's existing vocabulary — the same
-- discipline offer-analyser's prompt already carries ("MUST be exactly one of these
-- seven strings … Any other value routes the finding to a bucket no handler reads").
-- Single responsibility holds on both sides: the CATEGORY names the repair shape
-- (who fixes it), the AUDIT_SOURCE ('brief-fidelity-audit', unchanged, set by the
-- set_audit_source step per the bugs_closed/264 fix) names the axis it was found on
-- (artefact vs brief). No Go change, no new item_type, no closed-set edit: every
-- category offered below already routes to a live handler (verified against
-- classifyFinding's sets at commit d6d56e540).
--
-- The escape hatch is stated in the prompt: a finding fitting no category may use
-- its own snake_case label, which — once commit d6d56e540 rolls — files as a
-- capability_gap roadmap row instead of being minted-and-lost. Until that roll the
-- auditor has no caller (418 is held), so the residual is theoretical.
--
-- ROLLBACK: 417_..._ROLLBACK.sql restores the single brief_fidelity category.

SELECT snapshot_agent('brief-fidelity-auditor',
                      '417_brief_fidelity_auditor_speaks_the_routers_vocabulary.sql: pre-update');

BEGIN;

DO $$
DECLARE
    n_defs integer;
    occ_cat integer;
    occ_anchor integer;
BEGIN
    SELECT count(*) INTO n_defs FROM agent_definitions
    WHERE type='brief-fidelity-auditor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n_defs <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 417: expected exactly 1 live brief-fidelity-auditor, found %', n_defs;
    END IF;

    SELECT (length(default_config::text) - length(replace(default_config::text,'brief_fidelity','')))/length('brief_fidelity'),
           (length(default_config::text) - length(replace(default_config::text,'Respond with ONLY a JSON array','')))/length('Respond with ONLY a JSON array')
      INTO occ_cat, occ_anchor
    FROM agent_definitions
    WHERE type='brief-fidelity-auditor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF occ_cat = 0 THEN
        RAISE EXCEPTION 'MIGRATION 417: brief_fidelity already absent — already applied';
    END IF;
    IF occ_cat <> 1 OR occ_anchor <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 417: expected 1 occurrence of each anchor (category=%, respond-anchor=%) — prompt has drifted, re-derive the literals', occ_cat, occ_anchor;
    END IF;
END $$;

-- Two replacements in one pass: the category schema fragment gains the routable
-- vocabulary, and a "choose by repair shape" section lands just above the response
-- format. Literals operate on the jsonb::text rendering (inner quotes \" and
-- newlines \n).
UPDATE agent_definitions
SET default_config = replace(replace(default_config::text,
        '\"category\":\"brief_fidelity\"',
        '\"category\":\"colour|spacing|typography|responsive|cta|nav_restructure|gap|content|tone|structure|differentiation\"'),
        'Respond with ONLY a JSON array',
        '## Category — choose by REPAIR SHAPE\nThe category field decides WHO repairs the finding, so pick by what the fix IS, not by why you found it (your brief-fidelity identity travels in audit_source automatically):\n- colour / spacing / typography / responsive → design and CSS fixers (visual promises: styling, layout polish, breakpoints)\n- cta / nav_restructure → component fixer (calls-to-action, navigation promises)\n- gap → missing promised content (a page or section the brief implies and the site lacks)\n- content / tone / structure / differentiation → content rebuild or planning (register, copy substance, page composition promises)\nUse EXACTLY one of those strings. If a broken promise genuinely fits none of them, use a short snake_case label of your own: it will surface as a capability-gap roadmap entry for a human rather than being dispatched. The page field must be an EXACT page name from the inventory above, or site-wide.\n\nRespond with ONLY a JSON array')::jsonb,
    version    = version + 1,
    updated_at = now()
WHERE type='brief-fidelity-auditor' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
    occ_old integer;
    occ_vocab integer;
    n_steps integer;
BEGIN
    SELECT (length(default_config::text) - length(replace(default_config::text,'brief_fidelity','')))/length('brief_fidelity'),
           (length(default_config::text) - length(replace(default_config::text,'choose by REPAIR SHAPE','')))/length('choose by REPAIR SHAPE')
      INTO occ_old, occ_vocab
    FROM agent_definitions
    WHERE type='brief-fidelity-auditor' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF occ_old <> 0 THEN
        RAISE EXCEPTION 'MIGRATION 417: brief_fidelity still present (%) after the replace', occ_old;
    END IF;
    IF occ_vocab <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 417: the repair-shape section did not land (found %)', occ_vocab;
    END IF;

    SELECT count(*) INTO n_steps
    FROM jsonb_object_keys((SELECT default_config->'workflow'->'steps'
                            FROM agent_definitions WHERE type='brief-fidelity-auditor'
                              AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL)) k;
    IF n_steps <> 6 THEN
        RAISE EXCEPTION 'MIGRATION 417: workflow steps count changed (%) — the text surgery damaged the config', n_steps;
    END IF;

    RAISE NOTICE 'migration 417 OK: brief-fidelity-auditor now emits the router vocabulary; 6 steps intact';
END $$;

COMMIT;
