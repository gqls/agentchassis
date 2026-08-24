-- 587 — withdraw contrast_failure rows whose selector was INVENTED by the render
-- audit's class fallback (bugs_open/352).
--
-- WHAT IS WRONG WITH THESE ROWS. The render audit recorded a class-less
-- element's TAG NAME in a field called `Class`, so the orchestrator composed
-- `H3.H3` / `P.P` / `A.A`. As CSS that selects elements carrying class="H3",
-- of which there are none. css-patch-agent faithfully writes a rule against the
-- selector it is given, deploys it, and the item completes honestly — a write
-- really did happen. The text stays unreadable.
--
-- WHY WITHDRAW RATHER THAN REKEY. There is no truthful rekey available in SQL:
-- the correct new key is DOM-derived (the nearest ancestor carrying a class or
-- id), which this database cannot compute. And a bare-tag rekey — `#P` instead
-- of `#P.P` — would hand a parked fixer exactly the site-wide selector the code
-- fix exists to refuse: css-patch-agent appends to the ONE site stylesheet, so
-- `p { color: … }` recolours every paragraph on the site.
--
-- `cancelled` asserts WITHDRAWAL, NOT RESOLUTION. It is in
-- workItemClosedStatuses, so no retraction path can ever touch these rows
-- again and mint a false "no longer failing"; and by migration 157 it is also
-- terminal for dedup, so it FREES the idx_swi_dedup slot and the still-failing
-- pairings are re-filed under verified selectors by the next render audit.
--
-- ⚠ ORDERING — THIS IS WHY THE FILE IS `_HOLD` (council acadbe8b, guardian seat,
-- medium). Freeing the slot only helps once BOTH images carry the new selector
-- composition. Apply this ONLY after confirming both, per service, at the
-- artefact — never at git, never at a tag:
--
--   kubectl -n ai-persona-system logs -l app=browser-runner-adapter --tail=300 | grep -m1 'build provenance'
--   kubectl -n ai-persona-system logs -l app=agent-chassis          --tail=300 | grep -m1 'build provenance'
--   git merge-base --is-ancestor ffa6e1c3d <each stamp>
--
-- (An empty grep means "scrolled out of range", not "unstamped" — it is a
-- startup line. Fall back to the binary probe with a control in the same breath.)
-- Applied early, the freed slot is simply refilled with the same invented
-- selector by the old code on the next sweep: churn, not corruption. Applied
-- late, nothing is lost. So this is a tidiness gate, not a safety one — but the
-- plan claimed re-filing would be correct, and early application makes that
-- claim false, which is the objection.
--
-- HOW LONG BEFORE THE STILL-FAILING ONES COME BACK — MEASURED, not assumed
-- (council acadbe8b, prior_art_librarian seat, medium; it was right to ask).
-- [MEASURED 2026-08-24] Of the 13 affected sites, ALL 13 had a contrast_failure
-- filed within the last 14 days, but only 3 within 7 days and 2 within 3 days;
-- the oldest last-audit is 2026-08-10. So the honest window is **up to a
-- fortnight, not a week**. Earlier drafts of this change said "the next weekly
-- audit" — that was an overstatement and is corrected here and in the plan.
-- Nothing is lost in the interval: the defect remains on the page and is
-- re-detected, and the withdrawn row records its own pre-cancellation status.
--
-- ⚠ THE ONE THEORETICAL FALSE POSITIVE, and how it was ruled out. A site
-- genuinely using class="H3" on an <h3> would match this predicate and would be
-- a real, fixable finding (bugs_open/352's own candidate-4 caution: do not guess
-- at intent from a lossy string). [MEASURED 2026-08-24] Across every affected
-- site's page_components + site_components rendered_html, ZERO carry a class
-- token equal to the tag name, for all 13 distinct tokens (A, P, H3, H2, LEGEND,
-- STRONG, H1, LABEL, BUTTON, SPAN, EM, CODE, H4). POSITIVE CONTROL for that
-- query, because a zero with no control could not have come out otherwise: the
-- same predicate run against REAL class tokens from non-TAG.TAG findings found
-- 154 of 161. The _VERIFY file re-runs both arms; run it and read the distinct
-- keys before applying.
--
-- ⚠ REGEX DIVERGENCE (council acadbe8b, debug_historian, low). The predicate
-- below is a hand-written mirror of the producer's composition, not derived from
-- it. It keys on `item_key`, and the uppercase backreference `([A-Z][A-Z0-9]*)`
-- matching itself cannot match a real kebab-case class name. If the producer's
-- composition changes again, re-derive this rather than trusting it.
--
-- Re-runnable by design: run once more about a fortnight after both rolls to
-- sweep any rows the skew window let through. The premise guard below turns the
-- second run into a loud no-op rather than a silent one.

BEGIN;

-- The premise, asserted so it can FAIL. A verify block made of bare SELECTs
-- cannot stop a COMMIT — ON_ERROR_STOP ignores a non-empty result set — so this
-- is a DO block that RAISEs.
DO $$
DECLARE n integer;
BEGIN
    SELECT count(*) INTO n
      FROM site_work_items
     WHERE item_type = 'contrast_failure'
       AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled')
       AND item_key ~ '#([A-Z][A-Z0-9]*)\.\1$';

    IF n = 0 THEN
        RAISE EXCEPTION '587: no OPEN invented-selector rows found. Either this migration has already been applied, or the premise has changed. STOP and re-measure before forcing it.';
    END IF;

    RAISE NOTICE '587: withdrawing % open invented-selector contrast_failure row(s)', n;
END $$;

UPDATE site_work_items
   SET status     = 'cancelled',
       updated_at = now(),
       result     = COALESCE(result, '{}'::jsonb) || jsonb_build_object(
           'cancelled_by',    'migration_587',
           'pre_352_status',  status,
           'cancelled_at',    now(),
           'reason',
           'The item_key selector is TAG.TAG — invented by the render audit''s class '
        || 'fallback for a class-less element (bugs_open/352). It matches nothing on any '
        || 'page, so no fix aimed at it could ever apply, and any completion recorded '
        || 'against it is false. WITHDRAWN, NOT RESOLVED: the underlying contrast failure '
        || 'may well still be on the page. It is re-detected and re-filed under an '
        || 'in-page-verified selector by the site''s next render audit (measured '
        || '2026-08-24: every affected site was audited within 14 days).')
 WHERE item_type = 'contrast_failure'
   AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled')
   AND item_key ~ '#([A-Z][A-Z0-9]*)\.\1$';

-- The estate's conventional pipeline record, so the withdrawal is visible to
-- anyone counting these rows over time rather than appearing as an unexplained
-- drop in the durable backlog (the brochure_component_library lane counts them).
-- ⚠ doc_notes.categories is JSONB (not text[]), and subject_type/subject_key are
-- NOT NULL — an ARRAY[...] literal here fails at runtime. Shape copied from the
-- live rows, not from memory.
INSERT INTO doc_notes (subject_type, subject_key, categories, source, body)
SELECT 'action', 'write_render_audit_findings',
       '["migration","bugs_open/352","contrast_failure","render-audit"]'::jsonb,
       'bugfix_352_invented_selector lane',
       format(
         '587 (bugs_open/352): withdrew %s open contrast_failure row(s) whose item_key named an '
      || 'invented TAG.TAG selector. Status cancelled = withdrawn, not resolved; prior status is '
      || 'preserved in result.pre_352_status and _ROLLBACK restores it. The dedup slot is freed, so '
      || 'still-failing pairings return under verified selectors within ~14 days (measured). If you '
      || 'are tracking the durable contrast backlog, this is the explanation for the step change.',
         (SELECT count(*) FROM site_work_items
           WHERE item_type='contrast_failure' AND status='cancelled'
             AND result->>'cancelled_by' = 'migration_587'));

COMMIT;
