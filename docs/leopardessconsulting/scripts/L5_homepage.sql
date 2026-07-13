-- ============================================================================
-- L5_homepage.sql — rewrite leopardessconsulting.co.uk homepage content.
--
-- Removes every fabrication and rewrites the copy for the sceptical business
-- buyer (A2), in the voice spec's register: positive framing, no strawmen,
-- small exact claims, numbers traceable to AUDIT_verified_facts.md.
--
-- Layout: card grids are 3-up (no orphan row), per the brief. That is a CONTENT
-- fix, not a CSS one — the grid components are shared across 5 sites and must
-- not be edited.
--
-- Sections before:  hero | system-stats | features(8) | differentiators(6)
--                   | case-studies-grid(5, INVENTED CLIENTS, 404 images) | cta
-- Sections after:   hero | system-stats(4 real) | features(3, what we built)
--                   | differentiators(3, what we might build) | cta
--
-- rerender_page_sections re-renders each section from stored content_data via
-- RenderTemplate, so editing content_data is sufficient; rendered_html is
-- regenerated at L9. build_status stays 'deployed' (discovery checks filter on it).
-- ============================================================================
\set ON_ERROR_STOP on
\set S '4851f6fc-71cf-4160-a270-e03d6d3e0732'
BEGIN;

-- ── HERO ────────────────────────────────────────────────────────────────────
UPDATE page_components pc SET content_data = '{
  "headline": "AI systems that do one defined job, and keep doing it.",
  "subheadline": "Most of what we build is unglamorous, and that is the point. A pipeline that checks scraped business records against Companies House, and stops to ask a person when it is genuinely unsure. A system that reads across news sources and scores what is worth trusting. A website that keeps itself current. Each one runs without anybody watching it, and every decision it made is written down where you can read it back afterwards.",
  "primary_cta": "Tell us what the job is",
  "primary_cta_url": "/contact.html",
  "secondary_cta": "See what we have built",
  "secondary_cta_url": "/how-it-works.html",
  "cta_url": "/contact.html",
  "background_image": "/assets/images/hero.jpg"
}'::jsonb, updated_at = now()
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/index.html' AND cc.name='hero';

-- ── SYSTEM STATS ────────────────────────────────────────────────────────────
-- Every figure is a real count (AUDIT C1,C2,C4,C6). The old section rendered
-- nonsense: suffixes were misaligned ("70%" agents, "3ms" orchestration model,
-- "99.9x" uptime) and the uptime target was an unsupportable claim. Suffixes are
-- now empty strings so value+suffix concatenation cannot produce garbage.
UPDATE page_components pc SET content_data = '{
  "eyebrow_label": "From our own database",
  "section_headline": "Numbers from the systems we run",
  "section_intro": "Every figure here is a count taken from the database of the platform we built. That same platform built this website.",
  "stat1_value": "2,767", "stat1_suffix": "", "stat1_label": "Business records verified",
  "stat1_description": "Scraped UK business records reconciled against Companies House by a matching cascade that copes with name variations, and refers the genuinely ambiguous cases to a person rather than guessing.",
  "stat2_value": "4,672", "stat2_suffix": "", "stat2_label": "News items scored",
  "stat2_description": "Collected from RSS feeds, news search and live web search, then read and scored for how relevant each item is and how much its source deserves to be trusted.",
  "stat3_value": "8", "stat3_suffix": "", "stat3_label": "Live websites maintained",
  "stat3_description": "Planned, written, illustrated, deployed and kept current by the platform. This site is one of them.",
  "stat4_value": "75,061", "stat4_suffix": "", "stat4_label": "Decisions recorded",
  "stat4_description": "Every step every agent has taken, written to Postgres. When something fails it resumes from the step rather than the beginning, and afterwards you can read what happened.",
  "footnote_text": "Queried from the production database on 10 July 2026. These are our own systems. The eight sites are ours, and they demonstrate the platform rather than represent client engagements.",
  "cta_label": "How it works",
  "cta_url": "/how-it-works.html"
}'::jsonb, updated_at = now()
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/index.html' AND cc.name='system-stats';

-- ── FEATURES → "What we have built" (8 cards → 3) ────────────────────────────
UPDATE page_components pc SET content_data = '{
  "headline": "What we have built",
  "subheadline": "Three systems that are running now. Each one took over a job that used to need somebody watching it.",
  "features": [
    {
      "icon": "badge-check",
      "name": "Checking records against an official register",
      "description": "The web version of a company name is rarely the registered one, so a list of businesses gathered from the open web cannot simply be trusted. Our pipeline discovers an area, collects the businesses in it, and then tries to match each one to a real Companies House record. It compares names, confirms them against geography, and reads registration numbers straight out of a website footer where one is published. When a match falls into the genuinely uncertain band, the system stops and asks for a decision instead of guessing. So far it has verified 2,767 records and enriched 937 of them with filed accounts."
    },
    {
      "icon": "newspaper",
      "name": "Reading widely, and judging what deserves trust",
      "description": "Agents read across RSS feeds, news search APIs and live web search. A language model then scores two separate things for each item: how relevant it is to the subject, and how much weight the source deserves. It records why it gave the score it did, and it keeps the trail from the original publisher through whichever channel we happened to find the story on, because those are different facts and running them together is how a feed fills up with laundered rumour. It has collected 5,652 items and scored 4,672 of them."
    },
    {
      "icon": "globe",
      "name": "Websites that look after themselves",
      "description": "The platform plans a site, researches and writes the content, generates the images, deploys it, and then keeps checking its own work. It notices stale pages, broken tools and missing images, and fixes the routine ones on its own. It decides which interactive tools would genuinely help a particular audience, builds them to run in the visitor’s browser, and writes a guide to sit alongside each one. Eight sites run this way today, and you are reading one of them."
    }
  ]
}'::jsonb, updated_at = now()
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/index.html' AND cc.name='features';

-- ── DIFFERENTIATORS → "What we might build with you" (6 → 3) ─────────────────
-- Honestly labelled possibilities, per the brief: we may say what we might be
-- able to do, provided it is not pie in the sky and is marked as not-yet-done.
UPDATE page_components pc SET content_data = '{
  "headline": "What we might build with you",
  "features": [
    {
      "name": "Reconciling your records against a register you do not control",
      "description": "The Companies House pipeline is not really about Companies House. It is about the general problem of holding a messy internal list beside an authoritative external one, and working out row by row which things are the same thing. Charity registers, licensing bodies, professional accreditation lists, and your own CRM against your own billing system are all that same shape. We have built this once, against one register. Pointing it at another would be real work, and we would rather scope it honestly than promise it."
    },
    {
      "name": "Watching a subject, and telling you only what changed",
      "description": "The reading and the scoring already exist. What we have not built is the part that stays quiet until something actually matters: a regulation moves, a competitor changes their pricing, a supplier turns up in a story they would rather not be in. Knowing when to interrupt somebody is a harder problem than collecting the material, and we would want to understand your threshold for that before building it."
    },
    {
      "name": "Taking a repetitive process off a person",
      "description": "Somebody spends two days a month assembling a report from four systems. Somebody checks one list against another. Somebody reads an inbox and decides which of five things each message is. These are the jobs where an agent earns its keep, because the task is well defined, a mistake is visible, and a person can stay in the loop at exactly the step where judgement is needed. We would want to watch the process before saying whether it is worth automating. Sometimes the honest answer is that it is not, or that it needs a better form rather than a language model."
    }
  ]
}'::jsonb, updated_at = now()
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/index.html' AND cc.name='differentiators-section';

-- ── CALL TO ACTION ──────────────────────────────────────────────────────────
UPDATE page_components pc SET content_data = '{
  "headline": "Tell us what the job is",
  "subheadline": "The useful first conversation is a specific one. Describe the task you would like to hand over: the report somebody assembles by hand every month, the list that has to be checked against another list, the inbox that needs sorting before anyone can act on it. We will tell you what it would take, roughly what it would cost, and when the honest answer is that it is not worth automating.",
  "primary_cta": "Start that conversation",
  "primary_cta_url": "/contact.html",
  "secondary_cta": "Read how we work",
  "secondary_cta_url": "/how-it-works.html"
}'::jsonb, updated_at = now()
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/index.html' AND cc.name='call-to-action';

-- ── DELETE the fabricated case-studies grid (invented clients, 404 images) ───
-- Hard-wired to 5 cards (card1..card5), which cannot be 3-up without an orphan,
-- and every client name on it is invented (AUDIT U3). It references
-- /assets/images/case-study-0N.jpg, all of which return 404.
DELETE FROM page_components pc
USING pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/index.html' AND cc.name='case-studies-grid';

COMMIT;

SELECT pc.position, cc.name, jsonb_array_length(COALESCE(pc.content_data->'features', '[]'::jsonb)) AS cards
FROM page_components pc JOIN pages p ON p.id=pc.page_id LEFT JOIN content_components cc ON cc.id=pc.component_id
WHERE p.site_id=:'S' AND p.url='/index.html' ORDER BY pc.position;
