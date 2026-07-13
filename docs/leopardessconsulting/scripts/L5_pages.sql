-- ============================================================================
-- L5_pages.sql — remove the remaining fabrications and fix the card counts.
--
-- Fabrications removed (AUDIT U1, U3, U6, U7, U8, U9):
--   /about.html       leadership-team   — listed TWO AI AGENTS as team members,
--                                         with photo filenames that 404. Deleted;
--                                         the founder story moves into the prose
--                                         block, where a one-person practice
--                                         honestly belongs.
--   /how-we-work.html departments-grid  — "Seventy-plus agents across eight
--                                         functional departments" (no department
--                                         concept exists), plus Playwright,
--                                         anti-bot navigation and proxy pools,
--                                         none of which are in the codebase.
--   /who-we-help.html case-studies-grid — invented client names, 404 images,
--                                         hard-wired to 5 cards.
--
-- Card counts -> multiples of three (brief: 3 per row, no orphan panels, and
-- fewer panels repeating each other). Content fix, not CSS: the grid components
-- are shared across 5 sites.
--   services-grid  7 -> 6      services info-card-grid 4 -> 3
--   who-we-help info-card-grid 4 -> 3   (also: rewritten out of negative framing)
-- ============================================================================
\set ON_ERROR_STOP on
\set S '4851f6fc-71cf-4160-a270-e03d6d3e0732'
BEGIN;

-- ── /about.html : honest founder prose, real numbers ────────────────────────
UPDATE page_components pc SET content_data = pc.content_data || '{
  "eyebrow_text": "About",
  "heading": "One engineer, and a platform that does the repetitive part",
  "body_text": "<p>Leopardess Consulting is a small practice. The work is done by its founder, an engineer with thirty years of building software, most of it on systems that had to keep running whether or not anyone was awake. He built and ran worldsoccernews.com, which at its peak drew around twelve million unique users a month — covered at the time in a media trade magazine, and used in Microsoft''s own advertising. For a period it may have been among the three busiest sports sites there were. That particular ranking was never independently measured, so treat it as recollection rather than fact: it was bigger than the BBC''s football coverage at the time, and smaller than ESPN''s Soccernet.</p><p>A long stretch in financial systems followed, where being approximately right is the same as being wrong. That is where the habit of checking things against a source of record comes from, and it is why the pipeline we are proudest of stops and asks a person when it is genuinely unsure rather than picking the likeliest answer.</p><p>The platform described on this site was designed and built here, from first principles: workflow-driven agents on Kubernetes and Kafka, with state kept in Postgres so that when a step fails the work resumes from that step instead of starting again. Eight websites run on it, including this one. It writes their content, generates their images, builds their tools, and then keeps checking its own work.</p><p>We have no client engagements to show you, and we would rather say so plainly than dress our own systems up as somebody else''s. What we can show you is a platform that has been running for months, and the record of every decision it made while doing it.</p>",
  "stat_1_value": "30", "stat_1_label": "Years building software",
  "stat_2_value": "8",  "stat_2_label": "Live sites on the platform",
  "stat_3_value": "2,767", "stat_3_label": "Records verified against Companies House",
  "cta_label": "Tell us what the job is",
  "cta_url": "/contact.html"
}'::jsonb, updated_at = now()
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/about.html' AND cc.name='content-block-about';

DELETE FROM page_components pc USING pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/about.html' AND cc.name='leadership-team';

-- ── /how-we-work.html : drop the invented departments ───────────────────────
DELETE FROM page_components pc USING pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/how-we-work.html' AND cc.name='departments-grid';

-- ── /who-we-help.html : drop invented case studies; positive-framed cards ───
DELETE FROM page_components pc USING pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/who-we-help.html' AND cc.name='case-studies-grid';

UPDATE page_components pc SET content_data = '{
  "section_title": "Where an agent tends to earn its keep",
  "section_subtitle": "The jobs worth handing over have a shape in common: the task is well defined, a mistake is visible, and a person can stay in the loop at the step where judgement actually matters.",
  "cards": [
    {
      "title": "A list that has to be checked against another list",
      "description": "Somebody reconciles your records against a register you do not control — a company register, a licensing body, an accreditation list, or simply your CRM against your billing system. The work is repetitive until it suddenly is not, and the interesting cases are the ambiguous ones. A system can do the matching and hand you only the rows it could not settle."
    },
    {
      "title": "A report someone assembles by hand every month",
      "description": "The data lives in four places, and pulling it together takes two days that the person doing it would rather spend elsewhere. This is the most common thing we are asked about, and it is usually worth doing, because the inputs are stable and the output is checkable."
    },
    {
      "title": "More reading than one person can keep up with",
      "description": "You need to know what is being said about a market, a competitor, or a regulation, and nobody owns that job. A system can read broadly, score what deserves attention, and record why it judged each source the way it did, so you can disagree with it when it is wrong."
    }
  ]
}'::jsonb, updated_at = now()
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/who-we-help.html' AND cc.name='info-card-grid';

-- ── /services.html : 7 -> 6 services, 4 -> 3 engagement cards ───────────────
UPDATE page_components pc SET content_data = '{
  "headline": "What we do",
  "subheadline": "Six things, described plainly. If your problem is not on this list, it may still be a good fit — the underlying work is usually the same shape.",
  "features": [
    {"icon":"badge-check","name":"Collecting data, and checking it against a source of record",
     "description":"We gather information at volume and then verify it against something authoritative, rather than trusting the place it came from. Where a match is genuinely ambiguous, the system refers it to a person instead of guessing."},
    {"icon":"newspaper","name":"Monitoring what is published about a subject",
     "description":"Agents read across feeds, news search and the live web, then score each item for relevance and for how much its source deserves to be trusted, keeping a note of where it came from and why it scored the way it did."},
    {"icon":"globe","name":"Building and maintaining a website",
     "description":"The platform plans a site, writes and reviews its content, generates its images, deploys it, and then keeps auditing its own work — finding stale pages, broken tools and missing images, and fixing the routine ones."},
    {"icon":"calculator","name":"Interactive tools for your visitors",
     "description":"Calculators, estimators and assessments that run entirely in the visitor''s browser. We work out which ones would genuinely help a particular audience, build them, and publish a written guide beside each one."},
    {"icon":"server","name":"Agent orchestration on infrastructure you own",
     "description":"The same architecture can be built on your own Kubernetes cluster. Agent behaviour is stored as data rather than compiled in, so a workflow can be changed while the system is running. Where you want a person to approve a step first, that is a setting, made per stage."},
    {"icon":"shield","name":"Keeping sensitive steps on infrastructure you control",
     "description":"Each step in a workflow can be pointed at a different model, including one running entirely inside our own cluster with no third party involved. If part of your work cannot leave the building, we can architect only that step to run that way. We have built the mechanism; we have not yet run it for a client."}
  ]
}'::jsonb, updated_at = now()
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/services.html' AND cc.name='services-grid';

UPDATE page_components pc SET content_data = '{
  "section_title": "How an engagement usually starts",
  "section_subtitle": "We have no client track record to point at, so we would rather start small and let the work make the case than ask you to take a large step on trust.",
  "cards": [
    {
      "title": "A bounded pilot, priced as a project",
      "description": "Pick one job with a clear edge to it. We scope it, agree a fixed price, and build it. You end up with something running that you can judge on its own terms, and we end up having earned the next conversation. This is the step we would recommend to almost anyone."
    },
    {
      "title": "Then whichever shape the pilot suggests",
      "description": "Some work wants a licence and an installation on your own infrastructure. Some wants a few days a month keeping things healthy. Some wants nothing further at all, because the system runs and that was the point. We would rather decide that after the pilot has told us something true than guess at it beforehand."
    },
    {
      "title": "An honest answer about whether to bother",
      "description": "Sometimes the right recommendation is that a job should not be automated, or that it needs a better form rather than a language model. Saying so costs us an engagement and earns us a reputation, and we would rather have the second one."
    }
  ]
}'::jsonb, updated_at = now()
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/services.html' AND cc.name='info-card-grid';

COMMIT;

-- verify: card counts are all multiples of 3, and the fabricated sections are gone
SELECT p.url, cc.name,
       COALESCE(jsonb_array_length(pc.content_data->'features'),
                jsonb_array_length(pc.content_data->'cards')) AS cards
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN content_components cc ON cc.id=pc.component_id
WHERE p.site_id=:'S'
  AND (cc.name IN ('features','info-card-grid','services-grid','differentiators-section')
       OR cc.name IN ('leadership-team','departments-grid','case-studies-grid'))
ORDER BY p.url, pc.position;
