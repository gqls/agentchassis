-- 669: build-site-planner's worked-example logo prompt licenses "a wordmark" it never
-- names, so the image model invents brand text ("Farm Shield Info" on farmerinsurance.uk,
-- bugs_open/417). OWNER RULING 2026-08-31 (loanzy lane, decision 1): logos carry NO WORDS.
-- This aligns the exemplar with the estate's own ruled default
-- (discovery_checks/default_brand_prompt.go: "no lettering or words" — "generated
-- wordmarks reliably produce malformed text, and this asset is used at favicon size").
-- Companion 670 rewrites the 19 already-propagated plan prompts (417 §3: this migration
-- alone stops the NEXT sites and repairs NONE of the 19 — apply both).
-- Verify census AFTERWARDS must count the wordmark LICENCE, not the prohibition (417 §2).
\set ON_ERROR_STOP on
BEGIN;
DO $$
DECLARE v_cfg text; v_n int;
BEGIN
  -- probe guard: already applied?
  SELECT count(*) INTO v_n FROM agent_definitions
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false
     AND default_config::text LIKE '%no text outside the wordmark itself%';
  IF v_n = 0 THEN
    RAISE EXCEPTION '669 probe: exemplar anchor absent — already applied or drifted; do not re-apply blind';
  END IF;
  IF v_n > 1 THEN
    RAISE EXCEPTION '669 probe: % live planner rows carry the anchor (expected 1)', v_n;
  END IF;
  -- drift guard: the anchor must occur EXACTLY once in the row
  SELECT default_config::text INTO v_cfg FROM agent_definitions
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false;
  IF (length(v_cfg) - length(replace(v_cfg, 'no text outside the wordmark itself', '')))
     / length('no text outside the wordmark itself') <> 1 THEN
    RAISE EXCEPTION '669 drift: anchor not exactly-once in build-site-planner config';
  END IF;
  PERFORM snapshot_agent('build-site-planner', '669_pre_logo_exemplar_no_words');
  UPDATE agent_definitions
     SET default_config = replace(default_config::text,
           'no text outside the wordmark itself',
           'no lettering or words of any kind (a text-free mark: the brand name is set in HTML beside the logo, never rendered in the image)')::jsonb,
         updated_at = now()
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false;
  -- verify: licence gone, replacement present exactly once
  SELECT default_config::text INTO v_cfg FROM agent_definitions
   WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false;
  IF v_cfg LIKE '%no text outside the wordmark%' THEN
    RAISE EXCEPTION '669 verify: wordmark licence still present after replace';
  END IF;
  IF (length(v_cfg) - length(replace(v_cfg, 'a text-free mark', '')))
     / length('a text-free mark') <> 1 THEN
    RAISE EXCEPTION '669 verify: replacement not exactly-once';
  END IF;
END $$;
COMMIT;
