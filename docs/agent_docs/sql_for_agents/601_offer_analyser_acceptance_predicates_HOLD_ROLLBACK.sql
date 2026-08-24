-- 601_offer_analyser_acceptance_predicates_HOLD_ROLLBACK.sql
--
-- Undoes 601: removes the gate step, restores write_offer_ordering ->
-- write_offer_findings, points the findings write back at the raw model output,
-- and strips the predicate contract from the prompt. Restores today's behaviour
-- exactly — i.e. findings carry prose acceptance tests and nothing else, and
-- features_open/030 v2(d) reopens.
--
-- ⚠ ORDER MATTERS IF YOU ARE ALSO ROLLING THE IMAGE BACK. Revert the CONFIG
-- FIRST and the image second: a workflow naming verify_acceptance_predicates
-- against a binary that does not register it is rejected whole ("requires a
-- topic"), which is the same trap the forward file's _HOLD banner exists for,
-- just walked backwards.
--
-- ⚠ THE PROMPT ARM IS A BYTE-EXACT replace() AND IT IS A SILENT NO-OP IF IT
-- MISSES. The block below is lifted verbatim from the forward migration; if
-- another session has since edited the prompt inside that paragraph, the revert
-- will report success and leave the contract in the prompt. The verify block at
-- the end asserts the paragraph is GONE, so a miss fails the transaction rather
-- than reading as done.

BEGIN;

SELECT snapshot_agent('offer-analyser', '601_ROLLBACK: pre-revert');

-- Same duplicate-active-row guard as the forward migration: only the highest
-- version is ever loaded, so a second active row means an UPDATE here would
-- rewrite a row nobody reads.
DO $$
DECLARE total_active int;
BEGIN
    SELECT count(*) INTO total_active FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type = 'offer-analyser';
    IF total_active <> 1 THEN
        RAISE EXCEPTION
          '601 ROLLBACK: offer-analyser has % active definition rows, expected 1 — decide which row is real before reverting', total_active;
    END IF;
END $$;

UPDATE agent_definitions
   SET default_config = jsonb_set(
         jsonb_set(
           default_config #- '{workflow,steps,verify_finding_predicates}',
           '{workflow,steps,write_offer_ordering,next_step}',
           '"write_offer_findings"'::jsonb,
           false),
         '{workflow,steps,write_offer_findings,config,findings_field}',
         '"offer_analysis.result.findings"'::jsonb,
         false),
       updated_at = now()
 WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
   AND type = 'offer-analyser';

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,run_offer_analysis,config,prompt}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt',
           $rule$Bad: "The homepage should be more benefit-led".

ACCEPTANCE PREDICATE — optional, per finding, and usually omitted. If a NECESSARY condition of your acceptance_test can be stated mechanically over ONE page's meta description or title, add "acceptance_predicate" to that finding. If it cannot, OMIT THE KEY ENTIRELY. Most findings will omit it, and that is the right answer: a predicate over a clause that needs judgement ("reads as a generic contact-us button", "consistent with the one-founder model") is worse than no predicate at all, because it grades green while the judgement clause is still unmet.

Three rules. (1) A predicate can only ever REFUTE. Satisfying it never means the test is met — your prose acceptance_test stays the authority. (2) It must be FALSE OF THE PAGE AS IT IS TODAY: your finding says this page is wrong now, so the condition you write must fail now. One that already holds is removed mechanically after you answer, with a note saying why. (3) Only these three shapes, only these two fields, and no other keys. Text is matched case-insensitively at word boundaries.

{"type": "text_absent", "page": "about", "field": "meta_description", "values": ["curated", "hand-picked"]} — refuted if any listed phrase is present.
{"type": "text_present", "page": "index", "field": "title", "values": ["Kubernetes", "Kafka", "Postgres"], "min": 2} — refuted if fewer than "min" of them are present ("min" defaults to 1).
{"type": "text_order", "page": "index", "field": "meta_description", "before": ["no account", "no upload"], "after": ["$cardinal"]} — refuted if none of "before" is present, or if anything in "after" appears earlier than all of them. The reserved word "$cardinal", valid only inside "after", means the first quantity of any kind, in digits or in words — "63" and "Sixty-three" both count.

"page" defaults to the finding's own page. You are shown every page's title and meta description in THE OFFER SURFACE above: write predicates over those only, never over body copy you have not been shown.$rule$,
           'Bad: "The homepage should be more benefit-led".')),
         false),
       updated_at = now()
 WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
   AND type = 'offer-analyser';

DO $$
DECLARE cfg jsonb; prompt text;
BEGIN
    SELECT default_config INTO cfg FROM agent_definitions
     WHERE is_active AND NOT COALESCE(is_snapshot,false) AND deleted_at IS NULL
       AND type='offer-analyser';
    prompt := cfg->'workflow'->'steps'->'run_offer_analysis'->'config'->>'prompt';

    IF cfg->'workflow'->'steps' ? 'verify_finding_predicates' THEN
        RAISE EXCEPTION '601 ROLLBACK verify: the gate step is still present';
    END IF;
    IF cfg->'workflow'->'steps'->'write_offer_ordering'->>'next_step' <> 'write_offer_findings' THEN
        RAISE EXCEPTION '601 ROLLBACK verify: write_offer_ordering still points at the removed step — the workflow is now broken, not reverted';
    END IF;
    IF cfg->'workflow'->'steps'->'write_offer_findings'->'config'->>'findings_field' <> 'offer_analysis.result.findings' THEN
        RAISE EXCEPTION '601 ROLLBACK verify: the findings write still reads predicates_checked.findings, which no step now produces';
    END IF;
    IF prompt LIKE '%ACCEPTANCE PREDICATE — optional, per finding%' THEN
        RAISE EXCEPTION '601 ROLLBACK verify: the prompt contract is still present — the replace() missed, most likely because the paragraph has been edited since 601 was applied';
    END IF;
    IF prompt NOT LIKE '%Bad: "The homepage should be more benefit-led".%' THEN
        RAISE EXCEPTION '601 ROLLBACK verify: the anchor sentence was removed along with the paragraph';
    END IF;

    RAISE NOTICE '601 ROLLBACK OK: step gone, chain restored to write_offer_ordering -> write_offer_findings, prompt contract stripped';
END $$;

COMMIT;
