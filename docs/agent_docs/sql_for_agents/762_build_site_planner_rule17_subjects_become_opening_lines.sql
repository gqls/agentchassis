-- 762: rule 17's description of a "subject" becomes the OPENING-LINE spec
--      (owner decision 2026-09-03: A4 block + fuller-sentence subjects)
--
-- Authority: SPEC_2026-09-03_section_subject_phrasing.md (framework_prompts_
-- positive_voice; copy in apis_uk_bees_homepage/CONTRIB_2026-09-03b...). The
-- 641 block now prints {{.current_section.subject}} VERBATIM as the section's
-- opening line, so the planner's subjects become copy: today it writes
-- capitalised noun phrases with em dashes (real, gamedesign.uk 2026-09-02),
-- which would print as a section's first sentence. This swaps ONLY:
--   (a) the two subject-description sentences 640 inserted (the 380 sentence
--       "When no Verified Facts are listed..." is NOT ours and is untouched);
--   (b) the two example subject strings -- new ones deliberately generic and
--       carrying NO NUMBERS (an example in a prompt ships verbatim into live
--       pages; a copied figure would be a fabricated claim).
--
-- CONTRACTS HONOURED, verified before writing:
--   * 640's anchor / 729's pinned literal `may also carry a "subject"` is
--     PRESERVED VERBATIM in the new text (640's header EXTERNAL READERS note;
--     the 450 lane is notified with the exact length delta -- their 729 is
--     committed-unapplied with a length-delta guard and they re-anchor on top).
--   * Zero code/script/migration readers of the old example strings exist
--     (grepped platform/ scripts/ cmd/ sql_for_agents/ 2026-09-03; only prose
--     docs describing this very swap).
--   * No absolute positions/lengths recorded anywhere (the prompt moved twice
--     today under other lanes' edits); the verify is BYTE-EXACT equality
--     against an expected value computed from the live text in the same
--     transaction, which cannot go stale.
--
-- Config is live on apply. No ordering constraint: subjects are inert to the
-- WRITER until 641 (A4) applies, and the new shape is the desired one for any
-- plan written from now on.
--
-- Verify by hand after applying:
--   SELECT substring(default_config#>>'{workflow,steps,plan_site,config,prompt_template}'
--          FROM position('Any object entry may also carry' IN
--               default_config#>>'{workflow,steps,plan_site,config,prompt_template}') FOR 300)
--     FROM agent_definitions WHERE type='build-site-planner' AND is_active
--      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- expect: "...the line this section opens on, written to the reader..."

SELECT snapshot_agent('build-site-planner', '762_build_site_planner_rule17_subjects_become_opening_lines.sql: pre-update');

BEGIN;

DO $mig$
DECLARE
    t        text;
    t2       text;
    expected text;
    v_n      int;
    old_rule text := $r1$Any object entry may also carry a "subject": one line saying what THIS section specifically covers, distinct from every sibling section's subject on the page. A "subject" is REQUIRED on every entry whose component name appears more than once on the same page — repeated components without subjects all receive the same brief, and the writer then produces the same section several times.$r1$;
    new_rule text := $r2$Any object entry may also carry a "subject": the line this section opens on, written to the reader in the site's own voice, as a sentence or a short phrase that reads as one. It says what the reader gets from this section rather than naming its topic. Keep it to one sentence, because it is shown to every other section on the page as well and a long one crowds their briefs. Give every section on a page a different one. A "subject" is REQUIRED on every entry whose component name appears more than once on the same page, because repeated components given the same brief write the same section.$r2$;
    old_ex1  text := $e1$"subject": "What the platform does"$e1$;
    new_ex1  text := $e2$"subject": "Here is what the service does day to day."$e2$;
    old_ex2  text := $e3$"subject": "How a team adopts it"$e3$;
    new_ex2  text := $e4$"subject": "Here is how a team starts using it."$e4$;
BEGIN
    SELECT count(*) INTO v_n FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '762 ABORT: expected exactly 1 active build-site-planner row, found %', v_n;
    END IF;
    PERFORM 1 FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
      AND version = (SELECT max(version) FROM agent_definitions
                     WHERE type = 'build-site-planner' AND deleted_at IS NULL);
    IF NOT FOUND THEN
        RAISE EXCEPTION '762 ABORT: the active row is not the max version - a higher-version row would shadow this edit';
    END IF;

    SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}' INTO t
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF t IS NULL THEN
        RAISE EXCEPTION '762 ABORT: plan_site prompt_template not found';
    END IF;

    IF (length(t) - length(replace(t, old_rule, ''))) / length(old_rule) <> 1 THEN
        RAISE EXCEPTION '762 ABORT: the rule-17 sentence pair is not present exactly once - the prompt has drifted; re-derive from the live row';
    END IF;
    IF position(new_rule in t) > 0 THEN
        RAISE EXCEPTION '762 ABORT: already applied - the opening-line wording is present';
    END IF;
    IF (length(t) - length(replace(t, old_ex1, ''))) / length(old_ex1) <> 1
       OR (length(t) - length(replace(t, old_ex2, ''))) / length(old_ex2) <> 1 THEN
        RAISE EXCEPTION '762 ABORT: the 640 example strings are not each present exactly once';
    END IF;

    expected := replace(replace(replace(t, old_rule, new_rule), old_ex1, new_ex1), old_ex2, new_ex2);

    UPDATE agent_definitions
    SET default_config = jsonb_set(
            default_config,
            '{workflow,steps,plan_site,config,prompt_template}',
            to_jsonb(expected)),
        updated_at = NOW()
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    GET DIAGNOSTICS v_n = ROW_COUNT;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '762 ABORT: UPDATE touched % rows, not 1', v_n;
    END IF;

    SELECT default_config#>>'{workflow,steps,plan_site,config,prompt_template}' INTO t2
    FROM agent_definitions
    WHERE type = 'build-site-planner' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF t2 IS DISTINCT FROM expected THEN
        RAISE EXCEPTION '762 VERIFY FAILED: stored text is not byte-identical to the computed expectation';
    END IF;
    IF (length(t2) - length(replace(t2, 'may also carry a "subject"', ''))) / length('may also carry a "subject"') <> 1 THEN
        RAISE EXCEPTION '762 VERIFY FAILED: the shared anchor (640 probe / 729 pin) is not present exactly once';
    END IF;
    IF position('When no Verified Facts are listed' in t2) = 0 THEN
        RAISE EXCEPTION '762 VERIFY FAILED: the bugs_open/380 sentence has been disturbed';
    END IF;
    RAISE NOTICE '762 applied: rule 17 subjects are opening lines; length % -> % (delta %); anchor x1; 380 sentence intact.',
        length(t), length(t2), length(t2) - length(t);
END
$mig$;

COMMIT;

-- ROLLBACK recipe (hand-run): the same three replaces in reverse (new->old
-- literals above), same byte-exact shape; or restore from the snapshot this
-- file took. The 380 sentence needs no restoring - it is never touched.
