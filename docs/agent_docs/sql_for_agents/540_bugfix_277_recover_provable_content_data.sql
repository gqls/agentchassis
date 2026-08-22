-- 540 — bugs_open/277: recover content_data for the three parked
-- `no_content_data` components whose recovery can be PROVEN, and only those.
--
-- WHAT THIS WRITES. Three page_components rows gain a content_data recovered
-- from their own rendered_html by inverting their component template
-- (cmd/content-data-recover). Nothing else changes; no HTML is touched.
--
-- ⚠ WHY ONLY THREE, WHEN 27 ARE PARKED — the number is the finding, not a
-- shortfall. Measured 2026-08-22 over all 27 candidates:
--     3   recovered, round-trip byte-identical      <- this migration
--     7   blocked: template drift, inside <style> only
--     8   blocked: template drift, in MARKUP
--     9   MUST NOT be backfilled: the stored HTML is not that component's
--         output at all (a whole tool page held in a slot pointing at `hero`)
-- The 15 drifted rows were rendered by template versions that no longer exist:
-- component_versions holds 367 rows across 202 components and ZERO for any of
-- the nine components involved, so the original cannot be recovered and the
-- proof cannot be met. They stay parked, with the facts.
--
-- WHY THE PROOF IS THE WHOLE SAFETY ARGUMENT. datahelpers.ContentDataCanFillTemplate
-- returns true when content_data holds ANY ONE of a template's top-level fields.
-- So writing even a single field flips a component from "cannot regenerate" to
-- "can regenerate", and the next regeneration renders the template with that one
-- field and blanks the rest under missingkey=zero — the 004/007 blanking family.
-- A PARTIAL backfill would convert a page that is safe because it is unfillable
-- into one that rebuilds nearly empty. Each row below re-renders, through the
-- same text/template execution the estate renders components with
-- (missingkey=zero, actions.executeGoTemplate's FuncMap), to the stored bytes
-- EXACTLY — so making these three fillable is safe by construction: what a
-- regeneration would produce is what is being served today.
--
-- GUARDS, on every statement rather than as pre-checks, so a row that moved
-- since the export updates 0 rows instead of being written wrongly:
--   content_data IS NULL          — refuses to overwrite anything since written
--   md5(rendered_html) = <digest> — the proof is about THOSE bytes; if the
--                                   component has been re-rendered the proof
--                                   does not transfer and it must be re-proven
--
-- BACKUP: page_components_bak_20260822_277_recover holds the pre-image of the
-- three rows (the estate's page_components_bak_* convention), so the reverse is
-- a restore rather than a guess. ROLLBACK file does exactly that.
--
-- Tooling: cmd/content-data-recover (+ tests; the round-trip gate is
-- mutation-proven — disabling it fails exactly one test, which exists because
-- every other control is caught earlier by the matcher, a guard in series).

\set ON_ERROR_STOP on

BEGIN;

CREATE TABLE IF NOT EXISTS page_components_bak_20260822_277_recover AS
SELECT * FROM page_components WHERE 1=0;

INSERT INTO page_components_bak_20260822_277_recover
SELECT * FROM page_components
 WHERE id IN ('e50a9dbc-569c-41c5-ac01-bc564dc9a53a',
              'bd1f5219-c230-4143-93d7-7ece0f4d8e9f',
              '2b9d24d7-9e04-401b-a0b5-0e16e7731895');

-- case-studies-list  gaswholesalers.com/client-case-studies  (e50a9dbc-569c-41c5-ac01-bc564dc9a53a)  fields: subheadline, headline, case_studies
UPDATE page_components SET content_data = $cd${"case_studies":[{"client":"Independent Fuel Retailers","results":"Operators typically gain greater pricing visibility, reduced administrative burden around procurement, and a dependable supply relationship they can plan around.","summary":"Independent forecourt operators and fuel retailers often face inconsistent supply terms, unpredictable pricing, and suppliers who prioritise larger accounts. We work directly with independent station owners to establish reliable wholesale supply arrangements with clear, transparent pricing and consistent delivery schedules — so operators can focus on running their forecourts rather than chasing their fuel supply.","title":"Fuel Retail Network Supply"},{"client":"Fleet Operators and Logistics Businesses","results":"Fleet operators benefit from consolidated procurement, competitive wholesale rates appropriate to their volume, and supply continuity that keeps vehicles on the road.","summary":"Businesses running commercial vehicle fleets — from haulage companies to service contractors — need bulk fuel supply they can depend on without the overhead of managing multiple supplier relationships. We supply directly to fleet operators, providing wholesale pricing structures suited to high-volume, ongoing requirements and straightforward account management.","title":"Commercial Fleet Fuelling"},{"client":"Industrial Facilities and Process Operations","results":"Facilities gain a supply partner that understands operational continuity matters, with agreements structured around their actual usage patterns rather than off-the-shelf contracts.","summary":"Industrial sites and manufacturing operations require gas supply at scale, often with specific scheduling requirements tied to production cycles. Interruptions are costly. We work with facilities managers and procurement teams to structure supply agreements that align with operational demand, with clear terms and no unnecessary complexity.","title":"Industrial and Manufacturing Facility Supply"},{"client":"Multi-Site Convenience and Fuel Retail Chains","results":"Multi-site operators typically see simplified procurement administration, consistent supply terms across locations, and a single point of contact for all wholesale gas requirements.","summary":"Convenience store groups operating across multiple forecourt sites need supply arrangements that work consistently across locations without requiring site-by-site negotiation. We work with multi-site operators to consolidate their wholesale gas procurement under a single, manageable supply relationship — reducing the time spent managing supplier accounts and improving pricing consistency across the estate.","title":"Convenience Store and Forecourt Groups"}],"headline":"How We Work With Our Clients","subheadline":"Real supply challenges require straightforward solutions. Here is the kind of work we do and what businesses can expect when they work with Gas Wholesalers."}$cd$::jsonb, updated_at = now()
 WHERE id = 'e50a9dbc-569c-41c5-ac01-bc564dc9a53a'
   AND content_data IS NULL
   AND md5(rendered_html) = '7336eb0c529c5d060324936e7668740f';

-- use-cases-list  finetuning.uk/use-cases  (bd1f5219-c230-4143-93d7-7ece0f4d8e9f)  fields: subheadline, headline, use_cases
UPDATE page_components SET content_data = $cd${"headline":"What We Help Businesses Do","subheadline":"Real problems. Practical AI. No jargon. Here are the kinds of challenges we help UK businesses tackle — and what that looks like in practice.","use_cases":[{"client":"Professional Services","results":"Teams get back hours each week previously spent on manual triage. Response times improve. No genuine enquiry gets missed because someone was busy.","summary":"Many service businesses spend hours each week reading incoming enquiries, deciding which are worth pursuing, and routing them to the right person. We build AI systems that read, categorise, and triage incoming requests automatically — flagging high-priority leads, filtering out time-wasters, and drafting initial responses for review. Staff approve what goes out. The system handles the volume.","title":"Automating Client Intake and Request Screening"},{"client":"Operations and Admin Teams","results":"Staff stop digging through folders and start getting answers in seconds. Onboarding new team members becomes faster. Institutional knowledge stops walking out the door.","summary":"When your knowledge lives across emails, shared drives, PDFs, and old reports, finding the right information takes too long. We deploy RAG systems that let your team ask plain-English questions and get accurate answers drawn from your own documents — contracts, policies, past proposals, financial records, whatever you have. The AI cites its sources so staff can verify what they're reading.","title":"Searching and Summarising Internal Documents"},{"client":"Trades and Service Businesses","results":"Quote turnaround drops from days to minutes. Estimators spend their time on complex jobs rather than routine paperwork. Win rates improve when responses go out faster.","summary":"Putting together quotes is repetitive work that pulls skilled people away from the jobs that actually make money. We build automated quoting pipelines that take structured inputs — job type, size, materials, location — and produce accurate, formatted proposals ready for review and sending. Staff check the output and approve it. The system does the drafting.","title":"Generating Quotes and Proposals at Scale"},{"client":"Retail and Marketing Teams","results":"Businesses stay informed without the manual effort. Pricing decisions and marketing responses happen faster because the data is already there.","summary":"Keeping track of what competitors are doing — pricing changes, new product lines, promotional activity — is time-consuming to do manually and easy to let slip. We build data collection pipelines that monitor thousands of web sources automatically, structure the information, and surface what's relevant. You get a clear picture of the market without anyone having to spend hours on research.","title":"Monitoring Competitor and Market Data"},{"client":"Finance, Compliance and Due Diligence Teams","results":"Due diligence that used to take days can be completed in a fraction of the time. Teams get consistent, structured data rather than inconsistent manual extracts.","summary":"Pulling structured information from Companies House filings, regulatory documents, and public financial records is slow, error-prone work when done by hand. We build pipelines that collect, parse, and organise this data automatically — at scale, across thousands of companies if needed. The result is clean, structured data ready for analysis rather than a stack of PDFs to wade through.","title":"Extracting Data from Financial and Company Records"},{"client":"Healthcare Admin, Legal and Financial Services","results":"Businesses in regulated sectors can use AI without compromising on data security. Staff get powerful tools. Clients and regulators stay reassured.","summary":"Some businesses can't send client data to third-party AI providers — whether for regulatory reasons, client confidentiality, or simple data governance. We deploy private AI systems that run locally or within your own infrastructure. No data leaves your environment. You get the productivity benefits of AI without the compliance risk.","title":"Running AI on Sensitive Data Without Sending It Externally"},{"client":"Operations Leaders and Growth Teams","results":"Complex multi-step processes that previously required significant human coordination can run largely automatically, with people stepping in to approve rather than to execute.","summary":"Some tasks are too complex for a single AI model — they require research, decision-making, verification, and action across multiple steps. We design and deploy networks of AI agents that work together to handle these workflows: gathering information, checking facts, making decisions, and passing results between each other. Human review gates are built in wherever approval matters.","title":"Building and Deploying Agent Systems for Complex Workflows"},{"client":"Businesses with Specialist Knowledge or Terminology","results":"AI outputs that fit your business from day one rather than needing constant correction. Models that understand your terminology, tone, and context without having to be told every time.","summary":"General AI models don't know your industry's language, your product catalogue, your pricing logic, or your way of doing things. We fine-tune models on your specific data using efficient training techniques — so the AI you use actually understands your business. This isn't prompt engineering. It's a model that has genuinely learned from your content.","title":"Training AI Models on Your Own Business Data"}]}$cd$::jsonb, updated_at = now()
 WHERE id = 'bd1f5219-c230-4143-93d7-7ece0f4d8e9f'
   AND content_data IS NULL
   AND md5(rendered_html) = '7fee29b712c792e2d5c0a913efad2882';

-- features  finetuning.uk/our-position-on-ai  (2b9d24d7-9e04-401b-a0b5-0e16e7731895)  fields: subheadline, features, headline
UPDATE page_components SET content_data = $cd${"features":[{"description":"We remove manual, repetitive work from your business processes. That means things like screening incoming requests, generating quotes, analysing documents, and routing work to the right place — without a human having to touch it each time.","icon":"settings","name":"AI Automation Systems"},{"description":"Networks of AI agents that can tackle complex tasks collaboratively — analysing information, making decisions, and working through problems at a scale no single tool can match. Built for businesses with serious, multi-step challenges.","icon":"cpu","name":"Intelligent Agent Systems"},{"description":"Off-the-shelf AI doesn't know your industry, your terminology, or your way of working. We train models specifically on your data using efficient methods so they behave the way your business actually needs.","icon":"sliders","name":"Custom AI Model Training"},{"description":"Give your team an AI that can search and reason over your own documents, emails, financial records, and knowledge bases. Ask it a question and get an answer drawn from your actual data — not a generic guess.","icon":"database","name":"RAG Systems for Company Data"},{"description":"Not comfortable sending company data to third-party servers? Neither are we, sometimes. We deploy AI systems that run privately — locally or in-browser — so your data stays where it belongs and your costs stay predictable.","icon":"lock","name":"Private AI Deployments"},{"description":"Large-scale, structured data extraction from thousands of websites — including business registries, regulatory filings, and financial sources. If the data is publicly available, we can collect it systematically and deliver it in a format you can actually use.","icon":"download","name":"Bulk Data Collection"},{"description":"Automated collection and analysis of financial data from public sources. Useful for due diligence, market research, competitor tracking, or any workflow that currently involves someone manually pulling figures from websites.","icon":"bar-chart-2","name":"Financial and Company Data Pipelines"},{"description":"High-quality static websites — for company identity, marketing, landing pages, or documentation — built quickly without sacrificing quality. Useful when you need something professional up fast without a drawn-out agency process.","icon":"globe","name":"Rapid Website Generation"},{"description":"Not sure where to start? We work with you to identify where AI can genuinely save time or money in your business, then help you get there. Practical guidance from people who build these systems — not a slide deck and a handshake.","icon":"map","name":"AI Strategy \u0026 Implementation"}],"headline":"What We Actually Do","subheadline":"No mystery, no jargon. Here are the specific things we build and deploy for businesses like yours."}$cd$::jsonb, updated_at = now()
 WHERE id = '2b9d24d7-9e04-401b-a0b5-0e16e7731895'
   AND content_data IS NULL
   AND md5(rendered_html) = '584d9550a1992c5d9b43533adbaa9191';


DO $verify$
DECLARE n int; digests int;
BEGIN
  SELECT count(*) INTO n FROM page_components
   WHERE id IN ('e50a9dbc-569c-41c5-ac01-bc564dc9a53a',
                'bd1f5219-c230-4143-93d7-7ece0f4d8e9f',
                '2b9d24d7-9e04-401b-a0b5-0e16e7731895')
     AND content_data IS NOT NULL AND content_data::text NOT IN ('{}','null');
  IF n <> 3 THEN
    RAISE EXCEPTION '540: % of 3 target rows carry content_data, expected 3 (a guard refused — re-export and re-prove)', n;
  END IF;

  -- The HTML must be untouched: this migration writes data, never markup.
  SELECT count(*) INTO digests
    FROM page_components pc
    JOIN page_components_bak_20260822_277_recover b ON b.id = pc.id
   WHERE md5(pc.rendered_html) IS DISTINCT FROM md5(b.rendered_html);
  IF digests <> 0 THEN
    RAISE EXCEPTION '540: % rows had rendered_html change — this migration must not touch markup', digests;
  END IF;

  -- The backup must hold the pre-image, or the rollback is a guess.
  SELECT count(*) INTO n FROM page_components_bak_20260822_277_recover
   WHERE content_data IS NULL;
  IF n <> 3 THEN
    RAISE EXCEPTION '540: backup holds % NULL-content_data pre-images, expected 3', n;
  END IF;

  RAISE NOTICE '540 OK: 3 components recovered (round-trip proven), rendered_html untouched, pre-image backed up.';
END $verify$;

COMMIT;
