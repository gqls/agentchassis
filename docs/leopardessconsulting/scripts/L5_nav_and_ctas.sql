-- ============================================================================
-- L5_nav_and_ctas.sql — declutter the header nav, and stop every CTA repeating
-- "start with an architecture conversation, not a sales pitch".
--
-- NAV: design_intent wants a minimal, functional header. It currently has ~15
-- items including a BLANK page ("For Leaders" = for-engineering-leaders, 0 sections)
-- and three overlapping "how" pages. Header nav is built from pages.in_header at
-- render time (render_site_components_action.go:550), so setting in_header=false
-- drops a page from the nav without deleting the page. Kept in header (business-buyer
-- nav): Services, How It Works, Use Cases, What we've built, ROI Estimator, LLM Cost
-- Calculator, Insights, About, Contact.
--
-- CTAs: the voice spec bans motif repetition across pages ("signals content
-- shallowness to a reader who visits more than one page"). Six pages carried a
-- variant of the same "architecture conversation, not a sales pitch" CTA, in CTO
-- register. Rewritten per page for the sceptical business buyer (A2).
-- ============================================================================
\set ON_ERROR_STOP on
\set S '4851f6fc-71cf-4160-a270-e03d6d3e0732'
BEGIN;

-- ── NAV: drop from header (overlap / clutter / blank) ───────────────────────
UPDATE pages SET in_header=false
WHERE site_id=:'S' AND name IN (
  'for-engineering-leaders',  -- blank page (0 sections), redundant with for-engineering-teams
  'how-we-work',              -- overlaps how-it-works
  'our-approach',             -- overlaps how-it-works
  'who-we-help',              -- overlaps use-cases
  'password-entropy',         -- a password tool doesn't belong in the primary nav
  'tool-agent-complexity-estimator'  -- had a blank nav_label
);

-- give the estimator tool a real label (in case it's surfaced elsewhere), and set a
-- clean nav order for the kept items
UPDATE pages SET nav_label='Complexity Estimator' WHERE site_id=:'S' AND name='tool-agent-complexity-estimator';
UPDATE pages SET nav_order=2  WHERE site_id=:'S' AND name='services';
UPDATE pages SET nav_order=3  WHERE site_id=:'S' AND name='how-it-works';
UPDATE pages SET nav_order=4  WHERE site_id=:'S' AND name='use-cases';
UPDATE pages SET nav_order=5, nav_label='What we''ve built' WHERE site_id=:'S' AND name='case-studies';
UPDATE pages SET nav_order=6  WHERE site_id=:'S' AND name='ai-agent-roi-estimator';
UPDATE pages SET nav_order=7  WHERE site_id=:'S' AND name='llm-cost-calculator';
UPDATE pages SET nav_order=8  WHERE site_id=:'S' AND name='blog';
UPDATE pages SET nav_order=9  WHERE site_id=:'S' AND name='about';
UPDATE pages SET nav_order=10 WHERE site_id=:'S' AND name='contact';

-- ── CTAs: one distinct call per page, business-buyer register ────────────────
-- how-it-works: reader now understands the mechanism → invite them to name a job
UPDATE page_components pc SET content_data = content_data || '{
  "headline":"Have a job in mind?",
  "subheadline":"Describe the task you would like taken off someone''s desk — the report assembled by hand each month, the list checked against another list, the inbox that has to be sorted before anyone can act. We will tell you plainly whether it is a good fit and roughly what it would take.",
  "primary_cta":"Describe the job","primary_cta_url":"/contact.html",
  "secondary_cta":"See what we have built","secondary_cta_url":"/case-studies.html"
}'::jsonb, updated_at=now()
FROM pages p, content_components cc WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/how-it-works.html' AND cc.name='call-to-action';

-- our-approach
UPDATE page_components pc SET content_data = content_data || '{
  "headline":"If the approach makes sense to you",
  "subheadline":"The next step is a short conversation about the specific job you have in mind, so we can tell you whether it suits this kind of system and what it would take to build.",
  "primary_cta":"Start that conversation","primary_cta_url":"/contact.html",
  "secondary_cta":"Read the engagement model","secondary_cta_url":"/engagement-model.html"
}'::jsonb, updated_at=now()
FROM pages p, content_components cc WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/our-approach.html' AND cc.name='call-to-action';

-- how-we-work
UPDATE page_components pc SET content_data = content_data || '{
  "headline":"Talk through a piece of work",
  "subheadline":"Bring the process you would like to hand over. We will walk through how it would be built, where a person stays in the loop, and what running it would actually look like.",
  "primary_cta":"Get in touch","primary_cta_url":"/contact.html",
  "secondary_cta":"How a project is priced","secondary_cta_url":"/engagement-model.html"
}'::jsonb, updated_at=now()
FROM pages p, content_components cc WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/how-we-work.html' AND cc.name='call-to-action';

-- engagement-model
UPDATE page_components pc SET content_data = content_data || '{
  "headline":"Start with one bounded job",
  "subheadline":"Pick a single task with a clear edge to it. We scope it, agree a fixed price, and build it — so you can judge the result on its own terms before anyone commits to anything larger.",
  "primary_cta":"Scope a pilot","primary_cta_url":"/contact.html",
  "secondary_cta":"See what we have built","secondary_cta_url":"/case-studies.html"
}'::jsonb, updated_at=now()
FROM pages p, content_components cc WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/engagement-model.html' AND cc.name='call-to-action';

-- for-engineering-teams (this IS the technical layer — an engineer-facing CTA is fine here)
UPDATE page_components pc SET content_data = content_data || '{
  "headline":"Want to see how it is put together?",
  "subheadline":"If you would rather read the architecture than the brochure, bring your stack and the failure modes you are worried about. We will go through where state lives, how failures recover, and where human approval sits.",
  "primary_cta":"Talk to us","primary_cta_url":"/contact.html",
  "secondary_cta":"Read the architecture","secondary_cta_url":"/technical-architecture.html"
}'::jsonb, updated_at=now()
FROM pages p, content_components cc WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/for-engineering-teams.html' AND cc.name='call-to-action';

-- technical-architecture
UPDATE page_components pc SET content_data = content_data || '{
  "headline":"Questions about the architecture?",
  "subheadline":"Ask them directly. We would rather answer a hard technical question now than have it surface halfway through a build.",
  "primary_cta":"Ask a question","primary_cta_url":"/contact.html",
  "secondary_cta":"How an engagement works","secondary_cta_url":"/engagement-model.html"
}'::jsonb, updated_at=now()
FROM pages p, content_components cc WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/technical-architecture.html' AND cc.name='call-to-action';

COMMIT;

SELECT url, nav_label, in_header, nav_order FROM pages WHERE site_id=:'S' AND in_header=true ORDER BY nav_order;
