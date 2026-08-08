-- 340_site_review_agent_loads_the_premise.sql
--
-- vigilant_designer_offer_analysis programme, Phase B1 (PLAN 2026-08-02; owner
-- decision 2026-08-08: B1+B2 jump the queue ahead of the rest of Programme B —
-- see the lane PLAN's decision log and features_open/030).
--
-- THE DEFECT: site-review-agent's run_strategic_review asks offer-shaped
-- questions ("what single change would most improve conversion?") ~16x/fortnight
-- (llm_call_log, 14d to 2026-08-08) while load_strategic_context selects ONLY
-- domain, company, dream_spec, site_plan and two counts. The site's recorded
-- premise — strategy (17 sites), audience (29), identity (20), content_direction
-- (20), mission_brief (8) — is never loaded. The platform asks the offer
-- question blind, and the aspect census shows the answers sitting unread.
--
-- THE CHANGE, config only, live on apply:
--   1. load_strategic_context gains five COALESCE'd subselects (the query's own
--      existing pattern). strategy and audience load FULL; mission_brief loads
--      its {text} wrap; identity and content_direction are CAPPED at 4000 chars
--      WITH THE CAP IN THE COLUMN NAME (identity_head_4k,
--      content_direction_head_4k) — max sizes are 26.8KB and 50KB (measured
--      2026-08-08) and would dilute a 5-finding review; a silent cap is the
--      a-filtered-count-can-ship-inside-a-denominator trap, so the filter ships
--      in the name. Every subselect COALESCEs to '{}' / '' — a NULL reaching a
--      Go template renders `<no value>` (the css-patch commit-message defect
--      class, bugs_open/198).
--   2. run_strategic_review's prompt gains the premise blocks, STRATEGIC
--      QUESTION 7 (does the site match its OWN recorded
--      revenue_models.primary_model — doc 028: the revenue model shapes the
--      site, not the other way round), and an honesty constraint: the platform
--      has NO visitor behaviour data, so findings are premise-vs-artefact
--      mismatches, never claims about users.
--   3. The finding vocabulary is UNCHANGED (same 5 work_item_types, cap 5,
--      max_tokens 4000). No new item_type ships in B1 — that is B3/B4 territory.
--
-- ROLLBACK: no restore_agent_snapshot() exists. Restore from the backup row
-- (two-arg snapshot_agent writes agent_definitions_backup — NOT an is_snapshot
-- row; LANDMINES 2026-07-30):
--   UPDATE agent_definitions a SET default_config = b.default_config
--   FROM agent_definitions_backup b
--   WHERE a.type='site-review-agent' AND a.is_active
--     AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
--     AND b.type='site-review-agent'
--     AND b.snapshot_reason LIKE 'pre-update: B1%'
--   ORDER-BY-PROOF: pick b by snapshot_taken_at DESC LIMIT 1 in a subselect.
--
-- VERIFY (after apply): the B1 witness is a planted marker in a site's audience
-- aspect reaching llm_call_log.prompt_rendered on a real sweep — "applied" is
-- not "seen". See the lane PLAN_2026-08-08_B1_B2_premise_first.md.

-- Probe guard: refuse a second application.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM agent_definitions
        WHERE type = 'site-review-agent'
          AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
          AND default_config #>> '{workflow,steps,load_strategic_context,config,query}'
              LIKE '%content_direction_head_4k%'
    ) THEN
        RAISE EXCEPTION 'B1/340: already applied — load_strategic_context already loads the premise';
    END IF;
END $$;

-- Drift guard: this migration was composed against the exact texts fetched
-- 2026-08-08 ~17:00Z (md5s below). If another session has since changed either
-- step, REFUSE rather than clobber their work — recompose from the live row.
DO $$
DECLARE
    q_md5 text;
    p_md5 text;
BEGIN
    SELECT md5(default_config #>> '{workflow,steps,load_strategic_context,config,query}'),
           md5(default_config #>> '{workflow,steps,run_strategic_review,config,prompt}')
      INTO q_md5, p_md5
      FROM agent_definitions
     WHERE type = 'site-review-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF q_md5 IS DISTINCT FROM '4e4aac875befba711e0cb2fbed5ae07f'
       OR p_md5 IS DISTINCT FROM 'bedcd1cceca777c01a62aea413beb5b8' THEN
        RAISE EXCEPTION 'B1/340: DRIFT — live step text differs from what this migration was composed against (query %, prompt %). Another session changed it; recompose.', q_md5, p_md5;
    END IF;
END $$;

BEGIN;

SELECT snapshot_agent('site-review-agent',
    'pre-update: B1 (vigilant_designer_offer_analysis) — strategic review loads the premise');

UPDATE agent_definitions
SET default_config =
    jsonb_set(
    jsonb_set(
        default_config,
        '{workflow,steps,load_strategic_context,config,query}',
        to_jsonb($q$SELECT s.domain, COALESCE(s.company_name, s.domain) as company, s.content_data->>'dream_spec' as dream_spec, COALESCE(ss.data::text, '{}'::text) as site_plan, (SELECT COUNT(*) FROM pages WHERE site_id = s.id AND build_status = 'deployed') as deployed_pages, (SELECT COUNT(*) FROM site_work_items WHERE site_id = s.id AND status = 'complete') as completed_items, COALESCE((SELECT sp.data::text FROM site_specs sp WHERE sp.site_id = s.id AND sp.aspect = 'strategy' AND sp.is_current = true ORDER BY sp.created_at DESC LIMIT 1), '{}') as strategy, COALESCE((SELECT sp.data::text FROM site_specs sp WHERE sp.site_id = s.id AND sp.aspect = 'audience' AND sp.is_current = true ORDER BY sp.created_at DESC LIMIT 1), '{}') as audience, COALESCE((SELECT sp.data->>'text' FROM site_specs sp WHERE sp.site_id = s.id AND sp.aspect = 'mission_brief' AND sp.is_current = true ORDER BY sp.created_at DESC LIMIT 1), '') as mission_brief, COALESCE((SELECT left(sp.data::text, 4000) FROM site_specs sp WHERE sp.site_id = s.id AND sp.aspect = 'identity' AND sp.is_current = true ORDER BY sp.created_at DESC LIMIT 1), '{}') as identity_head_4k, COALESCE((SELECT left(sp.data::text, 4000) FROM site_specs sp WHERE sp.site_id = s.id AND sp.aspect = 'content_direction' AND sp.is_current = true ORDER BY sp.created_at DESC LIMIT 1), '{}') as content_direction_head_4k FROM sites s LEFT JOIN site_specs ss ON ss.site_id = s.id AND ss.aspect = 'site_plan' AND ss.is_current = true WHERE s.id = $1 ORDER BY ss.created_at DESC LIMIT 1$q$::text)
    ),
        '{workflow,steps,run_strategic_review,config,prompt}',
        to_jsonb($prompt$You are a website strategist. Review whether this site achieves its stated purpose.

IMPORTANT: Report ONLY the TOP 5 most impactful strategic issues. Quality over quantity. Focus on problems with the highest business impact. Skip cosmetic or minor content issues — those are handled by other auditors.

Domain: {{.strategic_context.domain}}
Company: {{.strategic_context.company}}
Deployed pages: {{.strategic_context.deployed_pages}}

Recorded strategy (the site's OWN premise — revenue model, value proposition, competitive position. Judge the site against THIS):
{{.strategic_context.strategy}}

Target audience (recorded at build time):
{{.strategic_context.audience}}

Identity (first 4000 chars):
{{.strategic_context.identity_head_4k}}

Content direction (first 4000 chars):
{{.strategic_context.content_direction_head_4k}}

Mission brief (the operator's own words; empty if none was given):
{{.strategic_context.mission_brief}}

Site plan summary:
{{.strategic_context.site_plan}}

Dream spec (aspirational goals):
{{.strategic_context.dream_spec}}

Content audit findings:
{{.content_audit_result}}

STRATEGIC QUESTIONS:
1. Is the site's overall message clear within 5 seconds of landing?
2. Does the page structure serve the business goal or is it generic?
3. What's the biggest gap between the dream spec and current reality?
4. What single change would most improve conversion?
5. Are there pages that should exist but don't?
6. Should any existing pages be restructured or merged?
7. Does the site match its OWN recorded premise? The recorded strategy above names a primary revenue model (revenue_models.primary_model). The revenue model shapes the site, not the other way round: a tools or advertising site with service-selling CTAs, a lead-generation site with no reachable enquiry path, or any mismatch between the recorded revenue shape and what the pages ask the visitor to do is a top-5 finding. If the recorded strategy above is empty ({}), say so in your summary and judge from the domain alone — do not invent a premise.

HONESTY CONSTRAINT: you have NO visitor behaviour data. Never claim to know what users want, do, or what converts. Phrase every finding as a mismatch between the recorded premise and the built artefact (or between the dream spec and reality) — never as a claim about user behaviour.

Respond with ONLY a JSON object:
{"overall_score": 1-10, "summary": "one paragraph", "findings": [UP TO 5 findings]}

Each finding MUST include ALL of these fields:
{"category":"structure|content|gap|cta|differentiation","severity":"high|medium|low","description":"what is wrong","current_value":"what is currently there or missing","suggestion":"specific fix recommendation","acceptance_test":"a concrete testable criterion that a DIFFERENT agent could verify","page":"which page (or site-wide)","work_item_type":"content_rewrite|needs_content_page|tone_shift|cta_improvement|nav_restructure","max_fix_attempts":2}

The acceptance_test must be specific enough to verify with a simple check. Good: "Homepage hero contains a clear value proposition with a single primary CTA button". Bad: "Site should convert better".

Example finding:
{"category":"gap","severity":"high","description":"No pricing page despite services-based business model","current_value":"Pricing page does not exist","suggestion":"Create pricing page with 2-3 tiered packages and clear CTAs","acceptance_test":"A page named pricing exists with at least 2 pricing tiers and a CTA per tier","page":"pricing","work_item_type":"needs_content_page","max_fix_attempts":2}$prompt$::text)
    )
WHERE type = 'site-review-agent'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verify inside the transaction: the update landed on exactly the active row.
DO $$
DECLARE
    n int;
BEGIN
    SELECT count(*) INTO n
      FROM agent_definitions
     WHERE type = 'site-review-agent'
       AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
       AND default_config #>> '{workflow,steps,load_strategic_context,config,query}'
           LIKE '%content_direction_head_4k%'
       AND default_config #>> '{workflow,steps,run_strategic_review,config,prompt}'
           LIKE '%revenue_models.primary_model%'
       AND default_config #>> '{workflow,steps,run_strategic_review,config,prompt}'
           LIKE '%HONESTY CONSTRAINT%';
    IF n <> 1 THEN
        RAISE EXCEPTION 'B1/340 VERIFY FAILED: expected exactly 1 updated active row, found %', n;
    END IF;
    -- The snapshot must hold the PRE-change text or it restores nothing
    -- (LANDMINES: two-arg snapshot_agent -> agent_definitions_backup).
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions_backup
        WHERE type = 'site-review-agent'
          AND snapshot_reason LIKE 'pre-update: B1%'
          AND default_config #>> '{workflow,steps,load_strategic_context,config,query}'
              NOT LIKE '%content_direction_head_4k%'
    ) THEN
        RAISE EXCEPTION 'B1/340 VERIFY FAILED: no backup row carrying the PRE-change query';
    END IF;
END $$;

COMMIT;
