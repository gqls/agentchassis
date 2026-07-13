-- ============================================================================
-- L5_faq_hero.sql — the last fabrications, on /who-we-help.html.
--
-- The FAQ carried "70+ agents across eight functional departments" (AUDIT U1/U2)
-- plus a stack of specifics I could not verify anywhere in the codebase:
-- per-agent token budgets, circuit breakers, routing to a cheaper model tier,
-- Helm charts, deployment into AWS/GCP/Azure, per-agent least-privilege IAM.
-- None of that ships. It is replaced with six questions the actual reader — a
-- commercially sharp non-specialist who has heard the hype — would ask, answered
-- only from AUDIT_verified_facts.md.
--
-- The hero was CTO copy in the banned register ("Not retrofitted. Not promised.
-- Deployed.") — a strawman and a staccato LLM tell.
-- ============================================================================
\set ON_ERROR_STOP on
\set S '4851f6fc-71cf-4160-a270-e03d6d3e0732'
BEGIN;

UPDATE page_components pc SET content_data = '{
  "headline": "Work worth handing to a machine",
  "subheadline": "If a job in your business is well defined, happens often, and a mistake in it would be visible, there is a good chance a system can take it on. This page describes the sort of work that suits that, and the sort that does not.",
  "primary_cta": "Tell us what the job is",
  "primary_cta_url": "/contact.html",
  "secondary_cta": "See what we have built",
  "secondary_cta_url": "/how-it-works.html",
  "cta_url": "/contact.html",
  "background_image": "/assets/images/hero.jpg"
}'::jsonb, updated_at = now()
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/who-we-help.html' AND cc.name='hero';

UPDATE page_components pc SET content_data = '{
  "section_title": "Questions people actually ask us",
  "questions": [
    {
      "question": "I have heard a great deal about AI and seen very little of it working. What can it actually do for my business?",
      "answer": "Less than the loudest people claim, and more than the sceptics expect. It is genuinely good at a narrow class of job: reading a lot of material and sorting it, comparing one list against another, filling in a form from a document, deciding which of five categories a message belongs to. It is poor at anything requiring judgement it cannot check, and it is confidently wrong when it is wrong. The useful question is not whether AI works, it is whether the specific job you have in mind is one of the jobs it is good at. That is a question we can usually answer in a conversation."
    },
    {
      "question": "How do I know it will not simply make things up?",
      "answer": "Because for anything that matters, you do not let it decide on its own. Our Companies House pipeline is the clearest example: it does not ask a language model whether two company names refer to the same business. It compares the names, confirms against geography, and reads the registration number out of the website footer where one is published. A language model is only used where its answer can be checked. Where a case falls into the genuinely ambiguous band, the system stops and asks a person rather than picking the likeliest answer. That design choice is the difference between a system you can trust and one you cannot."
    },
    {
      "question": "What happens when it gets something wrong?",
      "answer": "Two things. First, the work resumes rather than restarts — every step writes its state to a database, so a failure costs you that step, not the whole run. There are 75,061 of those state records from our own systems. Second, you can find out what happened. Every decision an agent took is written down, including what it was looking at when it took it. That is not a feature we added for auditors; it is how we debug our own work, and it is why we can tell you honestly what a system did rather than guessing."
    },
    {
      "question": "Can you keep our data away from a third-party AI provider?",
      "answer": "In part, and we would rather explain the boundary than overpromise. Each step in a workflow can be pointed at a different model, and one of those options is a model running entirely inside a cluster we control, which never calls out to anyone. So a workflow can be built where the step that touches your sensitive material stays inside that boundary while other steps use a commercial model. That mechanism is built and working. What we have not done is run it for a client, and we should be plain that our own platform today keeps every site''s data in one shared database — real separation would mean standing up a system dedicated to you, which is a conversation to have at the start rather than a box to tick afterwards."
    },
    {
      "question": "What would it cost, and how would we start?",
      "answer": "Start small. Pick one job with a clear edge to it, and we will scope it and agree a fixed price for that piece of work. You get something running that you can judge on its own merits, and nobody has committed to a long programme on the strength of a conversation. What happens after that depends on what the pilot taught us both — sometimes a licence and an installation on your own infrastructure, sometimes a few days a month, and sometimes nothing further, because the thing runs and that was the point."
    },
    {
      "question": "What if AI turns out to be the wrong answer for our problem?",
      "answer": "Then we will say so. Some of the jobs people bring to us do not need a language model; they need a better form, a fixed spreadsheet, or a rule that somebody should have written down years ago. Recommending that costs us an engagement and it is still the right thing to do, partly because a system nobody needed is a liability, and partly because we would rather be the people who told you the truth the first time."
    }
  ]
}'::jsonb, updated_at = now()
FROM pages p, content_components cc
WHERE pc.page_id=p.id AND cc.id=pc.component_id
  AND p.site_id=:'S' AND p.url='/who-we-help.html' AND cc.name='FAQ Section';

COMMIT;
