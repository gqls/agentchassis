-- SQL_2026-08-18g — reshape pageflow-builder's briefing_questionnaire so it works
-- for ANY sort of site, and so it asks the things that actually decide whether a
-- site is good. Owner instruction 2026-08-18, task 2.
--
-- ── WHY, MEASURED ───────────────────────────────────────────────────────────
-- The questionnaire is the ONLY one in the fleet: exactly one agent_definitions
-- row has a non-empty briefing_questionnaire, and pageflow-builder is the
-- recommended_builder on 20 of the 21 sites that have a briefing spec. So this
-- one object shapes the brief for effectively every site we build.
--
-- Its eleven fields were a corporate brochure intake form: company_name*,
-- about_us*, tagline, services*, leadership_team, case_studies, contact_email*,
-- contact_phone, headquarters, has_blog, has_careers  (* = required).
--
-- Two independent reasons that is wrong:
--   1. SITE TYPE. Fleet-wide, classification says interactive-platform 12,
--      brochure 6, interactive 2, ecommerce 1, editorial 1, hub 1. Only SIX of
--      twenty-three are brochures, yet every site is asked for its services,
--      leadership team and case studies. site_type is decided BEFORE this
--      questionnaire (domain-research-classifier writes the classification
--      aspect), so the questions arrive already knowing they do not fit.
--   2. THE OWNER'S RULING. webdesign.uk now attests `any_site_type`: "builds any
--      sort of site, not just business sites". Requiring company_name and
--      services contradicts that at the first question, and the chat bot was
--      fixed in two places on 2026-08-17 precisely to stop assuming a business.
--
-- ── WHAT IT NOW ASKS, AND WHY THOSE FIVE ────────────────────────────────────
-- site_purpose / audience / site_jobs / voice / avoid are the five things that
-- MISSION_2026-08-04_webdesign_uk.txt shows a good brief for this system carries,
-- and they are exactly what the site chat's prompt-maker draws out of a visitor
-- (facts.go promptConduct, 2026-08-18). That is the correlation: what the chat
-- collects now has somewhere to land, field for field.
--
-- ── SAFETY: WHAT MAY NOT BE RENAMED ─────────────────────────────────────────
-- `company_name` STAYS, spelled exactly that, because it is load-bearing well
-- beyond this questionnaire: 18 live agent_definitions reference it and it is a
-- column on `sites`. Only its LABEL changes, so it reads for a club or a person.
-- The four business-only names being retired (about_us, leadership_team,
-- case_studies, has_careers) are referenced by NO agent config -- checked, all
-- four came back false across every live row.
-- Downstream is safe for the rest: build-site-planner interpolates the WHOLE
-- briefing blob into its prompt as {{.site_specs.specs.briefing}} and never names
-- an individual field, so the consumer is an LLM reading JSON, not a parser.
--
-- `has_careers` is subsumed by `extra_sections`, which is strictly more
-- expressive (careers, gallery, events, downloads, whatever the type needs).
-- `contact_email` stays REQUIRED, unchanged: making it optional is a separate
-- decision and dropping it risks sites with no contact route.

BEGIN;

-- Backup first, following this repo's convention for agent_definitions surgery.
CREATE TABLE IF NOT EXISTS agent_def_pageflow_questionnaire_backup_20260818 AS
SELECT id, type, briefing_questionnaire, now() AS backed_up_at
  FROM agent_definitions
 WHERE type='pageflow-builder' AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;

UPDATE agent_definitions SET
  briefing_questionnaire = $q${
    "sections": [
      {"name":"site","title":"The site",
       "questions":[
         {"field":"company_name","label":"Name of the site, business, group or person it is for","type":"text","required":true},
         {"field":"site_purpose","label":"What this site is for, in plain words: what it is, or what it does","type":"textarea","required":true},
         {"field":"tagline","label":"A single line describing it, if there is one","type":"text","required":false}
       ]},
      {"name":"audience","title":"Who it is for",
       "questions":[
         {"field":"audience","label":"Who will read or use this, and what that person is like. A person, not a market segment.","type":"textarea","required":true}
       ]},
      {"name":"jobs","title":"What the site has to do",
       "questions":[
         {"field":"site_jobs","label":"What the site must achieve for that reader (list of {job, detail}). For a tool or platform this is what a visitor can DO; for a brochure it is what they must understand before getting in touch.","type":"json_array","required":true},
         {"field":"offerings","label":"Anything on offer: services, products, tools, activities, events (list of {name, description}). Leave empty if the site offers nothing.","type":"json_array","required":false}
       ]},
      {"name":"voice","title":"How it should sound",
       "questions":[
         {"field":"voice","label":"How the writing should sound, and whose voice it is","type":"textarea","required":true},
         {"field":"avoid","label":"Anything to avoid: words, claims, comparisons, visual cliches, or a style they are sick of","type":"textarea","required":false}
       ]},
      {"name":"contact","title":"Getting in touch",
       "questions":[
         {"field":"contact_email","label":"Email","type":"text","required":true},
         {"field":"contact_phone","label":"Phone","type":"text","required":false},
         {"field":"location","label":"Where it is based, if that matters to the reader","type":"text","required":false}
       ]},
      {"name":"extras","title":"Extra sections",
       "questions":[
         {"field":"people","label":"People the site should name, if any (list of {name, role, bio}). A committee, a founder, a sole trader, or nobody.","type":"json_array","required":false},
         {"field":"examples","label":"Work, projects, events or results worth showing (list of {title, description}). Only real ones.","type":"json_array","required":false},
         {"field":"has_blog","label":"Does it need a news, blog or updates section?","type":"boolean","required":false},
         {"field":"extra_sections","label":"Any other section this site needs (list of names), e.g. careers, gallery, events, downloads, opening times","type":"json_array","required":false}
       ]}
    ]
  }$q$::jsonb
WHERE type='pageflow-builder' AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;

DO $$
DECLARE q jsonb; nf int; nreq int; missing text := '';
BEGIN
  SELECT briefing_questionnaire INTO q FROM agent_definitions
   WHERE type='pageflow-builder' AND deleted_at IS NULL AND COALESCE(is_snapshot,false)=false;
  IF q IS NULL OR q = '{}'::jsonb THEN RAISE EXCEPTION 'questionnaire is empty after the write'; END IF;

  SELECT count(*) INTO nf FROM jsonb_array_elements(q->'sections') s, jsonb_array_elements(s->'questions') x;
  IF nf <> 15 THEN RAISE EXCEPTION 'expected 15 questions, found %', nf; END IF;

  -- company_name must survive EXACTLY, and stay required: 18 agents and sites.company_name depend on it
  SELECT count(*) INTO nreq FROM jsonb_array_elements(q->'sections') s, jsonb_array_elements(s->'questions') x
   WHERE x->>'field'='company_name' AND (x->>'required')::boolean;
  IF nreq <> 1 THEN RAISE EXCEPTION 'company_name must be present exactly once and required'; END IF;

  -- the five brief-shaping fields the chat prompt-maker feeds
  SELECT string_agg(k, ', ') INTO missing FROM unnest(ARRAY['site_purpose','audience','site_jobs','voice','avoid']) k
   WHERE NOT EXISTS (SELECT 1 FROM jsonb_array_elements(q->'sections') s, jsonb_array_elements(s->'questions') x
                      WHERE x->>'field'=k);
  IF missing IS NOT NULL THEN RAISE EXCEPTION 'brief-shaping field(s) missing: %', missing; END IF;

  -- the business-only fields must be GONE
  SELECT string_agg(k, ', ') INTO missing FROM unnest(ARRAY['about_us','services','leadership_team','case_studies','headquarters','has_careers']) k
   WHERE EXISTS (SELECT 1 FROM jsonb_array_elements(q->'sections') s, jsonb_array_elements(s->'questions') x
                  WHERE x->>'field'=k);
  IF missing IS NOT NULL THEN RAISE EXCEPTION 'business-only field(s) survive: %', missing; END IF;

  -- the backup must hold the OLD shape, or we have nothing to go back to
  PERFORM 1 FROM agent_def_pageflow_questionnaire_backup_20260818
    WHERE briefing_questionnaire::text ~ 'leadership_team';
  IF NOT FOUND THEN RAISE EXCEPTION 'backup does not contain the pre-change questionnaire'; END IF;
END $$;

COMMIT;
