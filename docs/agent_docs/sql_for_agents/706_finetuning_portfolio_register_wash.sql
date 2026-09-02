-- 706_finetuning_portfolio_register_wash.sql
--
-- Owner ruling 2026-09-02 (verbatim, copy lane ledger): "wash it. We don't have to use AI
-- slop - if we use those words we cheapen our offering." The finetuning.uk `portfolio` spec
-- feeds use_cases[] into pages AT BUILD TIME, bypassing the writer and the negation gate
-- (measured 2026-08-31: byte-identical tells through a full rebuild), so the spec store is
-- the only wash surface. Surgery offline, deletion-first with per-needle exactly-once asserts
-- (two word substitutions: Honest->Clear in a title, honest->clear in one results line); the
-- brief's DELIBERATE example utterance "AI isn't the right fit yet" is KEPT — it is content
-- the voice may emit, not register. No differentiated points in this corpus (marketing
-- prose), so the isTruncationOf prefix test does not bind here; stated, not skipped.
-- Inherits the 674 round's class requirements: aspect predicate on the flip, in-verify
-- recursive shape preserved (asserted OFFLINE at build — key sets identical — and the
-- verify below asserts structure counts), schema_migrations INSERT inside the transaction.
-- Predecessor id censused 2026-09-02: 6a22f018-3148-4da9-aca6-0a9ee28ba60d.
-- Apply: psql -v mig_checksum=<md5 of this file> -f <this file>. ROLLBACK alongside.

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '706_finetuning_portfolio_register_wash', 'site_specs', ss.id::text,
       jsonb_build_object('data', ss.data), 'pre-706 portfolio for finetuning.uk'
FROM site_specs ss WHERE ss.id='6a22f018-3148-4da9-aca6-0a9ee28ba60d';

DO $m$
DECLARE n int; sid uuid;
BEGIN
  SELECT count(*) INTO n FROM migration_backups WHERE migration_name='706_finetuning_portfolio_register_wash';
  IF n <> 1 THEN RAISE EXCEPTION '706: % backup rows, want 1', n; END IF;
  SELECT site_id INTO sid FROM site_specs
   WHERE id='6a22f018-3148-4da9-aca6-0a9ee28ba60d' AND is_current AND aspect='portfolio';
  IF sid IS NULL THEN RAISE EXCEPTION '706: predecessor 6a22f018 no longer the current portfolio row — regenerated since the 2026-09-02 census; re-base'; END IF;
  UPDATE site_specs SET is_current=false, superseded_at=now()
   WHERE id='6a22f018-3148-4da9-aca6-0a9ee28ba60d' AND is_current AND aspect='portfolio';
  GET DIAGNOSTICS n = ROW_COUNT;
  IF n <> 1 THEN RAISE EXCEPTION '706: flip touched % rows, want 1', n; END IF;
  INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
  VALUES (sid, 'portfolio', $j706${"use_cases": [{"title": "A Professional Website That Looks After Itself", "client": "Service Businesses (5-50 employees)", "results": "A professional, maintained website that stays current. You control what matters without doing the legwork.", "summary": "You need a website that actually represents your business. Our AI researches your industry, writes content that speaks to your customers, generates useful tools like calculators or estimators, and keeps everything updated. You can review and approve anything before it goes live, or let it run hands-off. Your call, page by page."}, {"title": "Continuous Site Improvement Without the Overhead", "client": "Businesses with Websites That Go Stale", "results": "A website that gets better over time. Routine maintenance handled automatically, significant changes flagged for your approval.", "summary": "Your website was fine when it launched, but now the content is dated, the tools are broken, and nobody has time to fix it. Our improvement pipeline runs continuously — auditing content quality, checking tools work properly, identifying gaps, and fixing what it finds. Every change can be reviewed before it goes live, or you can let the system handle routine maintenance autonomously while you approve the bigger changes."}, {"title": "Automated Data Collection You Can Trust", "client": "Businesses Drowning in Manual Research", "results": "Structured, verified data on your schedule. You review the exceptions.", "summary": "You spend hours every week checking competitor prices, tracking supplier details, or updating a spreadsheet of contacts. Our agents do that collection automatically — finding sources, extracting the data, and checking it against official records. Anything the system isn't sure about gets flagged for you to confirm."}, {"title": "Industry News Grounded in Credible Sources", "client": "Businesses That Need Fresh Content", "results": "Credible, well-sourced industry content on your site. AI does the research and heavy lifting, humans ensure the quality.", "summary": "Your website looks static because nobody has time to curate industry news. Our pipeline collects relevant news from multiple sources and checks the credibility of each item — is this from a recognised publication or an unverified social media post? AI handles the research, source-checking, and first drafts. A human reviews and finalises the quality before anything goes live on your site. The result is authoritative content."}, {"title": "Clear Assessment, Then Practical Implementation", "client": "Business Owners Not Sure Where AI Fits", "results": "A clear answer on where AI helps your business, followed by implementation that pays for itself — or a clear 'not yet' that saves you money.", "summary": "You keep hearing about AI but you're not sure what's real and what's hype for your business. I research your specific operations and constraints, then propose something concrete — sometimes that's automation, sometimes it's a better workflow, sometimes the answer is that AI isn't the right fit yet. If we build something, you have approval at every step. No black boxes."}], "case_studies": [{"title": "AI-Powered Website Generation for SMEs", "client": "Industry Portal Network", "results": "Multiple production sites running autonomously. Each site receives industry-specific content, relevant interactive tools, and ongoing quality improvements without manual intervention. SME owners get a professional web presence without the typical agency timeline or cost.", "summary": "Built an AI system that takes a domain name and autonomously creates a complete business website — researching the industry, writing targeted content, designing appropriate visuals, generating interactive tools, and deploying to production. The system understands different industries and adapts its output accordingly, from gas wholesalers to consulting firms."}, {"title": "Automated Data Collection and Business Verification", "client": "Veterinary Industry", "results": "Thousands of verified practice records with financial enrichment data. What would take a research team weeks runs continuously and autonomously, with built-in quality checks at every stage.", "summary": "Deployed AI agents to collect, verify, and enrich veterinary practice data across the UK. The system discovers practices by area, extracts structured information, then cross-references against Companies House to verify legitimacy — catching dissolved businesses and incomplete registrations automatically."}, {"title": "Intelligent Tool Suggestion and Generation", "client": "Fuel Distribution Sector", "results": "Industry-specific tools deployed across multiple sites with companion guides and automatic cross-referencing from related content pages. The system handles everything from tool concept through to live deployment.", "summary": "The AI evaluates what interactive tools would genuinely help a website's visitors based on their industry and needs — then builds and deploys those tools automatically. A gas wholesaler gets fuel cost calculators and unit converters; a consulting firm gets ROI estimators. No irrelevant suggestions, no manual development."}, {"title": "Automated News Collection with Credibility Filtering", "client": "News-Driven Businesses", "results": "Automated news sections running on production sites with six-hour refresh cycles. Source credibility tracking means businesses only display verified, relevant industry news — building authority without editorial overhead.", "summary": "Built a multi-source news pipeline that collects from RSS, web search, and AI-powered search simultaneously. An AI triage layer scores every item for relevance to the business and credibility of the source — distinguishing Reuters from anonymous social media posts before anything reaches the live site."}]}$j706$::jsonb, 'owner_review', 'copy_quality_two_stage',
          'Supersedes 6a22f018 per the owner ruling of 2026-09-02 ("we don''t have to use AI slop — if we use those words we cheapen our offering"): register constructions washed from the spec-fed use_cases copy; predecessor kept.',
          true, 'copy_quality_two_stage');
END $m$;

DO $v$
DECLARE d jsonb; n int;
BEGIN
  SELECT data INTO d FROM site_specs
   WHERE site_id=(SELECT id FROM sites WHERE domain='finetuning.uk') AND aspect='portfolio' AND is_current AND created_by='copy_quality_two_stage';
  IF d IS NULL THEN RAISE EXCEPTION '706 VERIFY: no current successor'; END IF;
  IF d::text ~* '(,\s+not\s+|rather than|instead of|not just|AI slop|plainly|honest)' THEN RAISE EXCEPTION '706 VERIFY: register battery fires on the successor'; END IF;
  IF NOT (SELECT data::text ~* '(,\s+not\s+|rather than|instead of|not just|AI slop|plainly|honest)' FROM site_specs WHERE id='6a22f018-3148-4da9-aca6-0a9ee28ba60d') THEN
    RAISE EXCEPTION '706 VERIFY CONTROL: battery silent on the predecessor — it cannot see what it checks for';
  END IF;
  SELECT jsonb_array_length(d->'use_cases') + jsonb_array_length(d->'case_studies') INTO n;
  IF n <> 9 THEN RAISE EXCEPTION '706 VERIFY: % entries, want 9 (5 use_cases + 4 case_studies) — structure lost', n; END IF;
  IF position('AI isn''t the right fit yet' IN d::text) = 0 THEN
    RAISE EXCEPTION '706 VERIFY: the deliberately KEPT voice utterance is gone — over-washed';
  END IF;
  RAISE NOTICE '706 verify: successor current, battery-clean, 9 entries, kept-utterance present; control fired.';
END $v$;

INSERT INTO schema_migrations (filename, checksum, applied_by, notes)
VALUES ('706_finetuning_portfolio_register_wash.sql', :'mig_checksum', 'copy_quality_two_stage session',
        'Owner ruling 2026-09-02: wash the spec-fed portfolio register on finetuning.uk (the surface no rebuild can reach). Predecessor 6a22f018 kept.');

COMMIT;
