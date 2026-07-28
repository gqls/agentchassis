-- 258 — travelling doc_notes for the three actions changed by bugs_open/100 + bugs_closed/101
--
-- WHY THIS FILE EXISTS: the council gate's `tooling_provenance` seat raised the
-- GATING objection on round 1 (corr f4cf0aab) — the fix touched four subjects that
-- carry travelling PLAN+NOTES (doc_plans/doc_notes keyed by subject_type+subject_key)
-- and wrote to none of them. The archaeology was real (code greps, live-DB sampling,
-- bugs_open/*, WRONG_CALLS.md) but it was recorded in a PARALLEL, self-built trail —
-- a workstream directory and commit messages — rather than in the platform's own
-- mechanism, where the next person editing these actions would actually meet it.
--
-- The irony the seat did not have to point out: the same submission argued at length
-- that extending the existing inert ActionInputSpec registry beat building a second
-- one. That argument was made about Go code while the identical error was committed
-- about documentation.
--
-- Each note below records the NON-OBVIOUS thing — the bit the next reader would
-- otherwise re-derive from scratch — not a summary of the fix.

BEGIN;

-- ---------------------------------------------------------------------------
-- scrape_web — the config-key contract, and the fifth key nobody had found
-- ---------------------------------------------------------------------------
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('action', 'scrape_web',
'2026-07-28 (bugs_closed/101, live on chassis v1.0.1192): this action advertised FIVE config keys that no Go code read. Four were in the bug file (max_pages, follow_links, extract_mode, fallback_url_field); the fifth, add_protocol, was found by scripts/audit-config-keys.sh on its first run and is a NEAR-MISS TYPO — the only Go code with that intent reads "add_protocol_if_missing" and belongs to a DIFFERENT action (the URL-validation one). A bare domain reaching the adapter is a failed fetch, so it was not cosmetic. All five are now read.

LANDMINE FOR THE NEXT EDITOR: this action now DECLARES its config contract via ActionInputSpec.ConfigKeys, so the workflow validator reports any step-config key not in that list instead of silently ignoring it. If you teach WebscrapeAction a new config key you MUST add it to WebscrapeInputSpec, or it becomes invisible again — there is a test (TestScrapeWebDeclaresEveryKeyItReads) that fails if you forget.

LANDMINE 2 — do not read "UNKNOWN KEYS: none" from the audit as "no step misdescribes itself". max_pages/follow_links are DECLARED (so the audit is quiet about them) but can only take effect on a CRAWL: Firecrawl /scrape fetches exactly one page. vet-practice-verifier/scrape_website and domain-research-classifier/scrape_site still advertise a three-page crawl and will only ever fetch one. They now WARN at runtime rather than doing it silently. Switching them to action:"crawl" is a deliberate behaviour change to two other owners agents (domain-research-classifier has NO owner) and was left undone on purpose.

StrictConfig is deliberately FALSE for this action: flipping it makes an unknown key a hard validation failure, which would break those two running agents to make a point about their config. Clean the definitions first.',
'["bugs_closed-101", "config-contract", "landmine", "council-gate", "scrape"]'::jsonb,
'bugs_closed/101 + docs024_key_docs_latest/bugfix_100_101_scrape_provenance/',
'bugsearch-thread');

-- ---------------------------------------------------------------------------
-- firecrawl_scrape — omission is an instruction
-- ---------------------------------------------------------------------------
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('action', 'firecrawl_scrape',
'2026-07-28 (bugs_closed/101, live on web-scrape-adapter v1.0.1192): until this date, "only_main_content: false" was INEXPRESSIBLE on the /scrape path. FirecrawlScrapingProvider.Scrape read the key into a bool and then added it to the payload ONLY WHEN TRUE, so false and unset produced an identical request — and they are not identical to Firecrawl, whose documented default is onlyMainContent=true (it strips headers, navs and footers). Every caller explicitly asking for the whole page received the exact opposite, silently, for as long as the code existed.

THREE LIVE STEPS were asking for false and getting main-content-only: site-scraper/scrape_site, site-adoption-agent/fetch_primary_css, website-capture-firecrawl/scrape_main_page. (A fourth, site-adoption-agent/crawl_site, sets false too but is firecrawl_crawl and always took the CORRECT path — the /crawl payload builder in the same file has always presence-checked. Two paths, one file, opposite semantics. That is why this survived: anyone checking "do we support this key?" found a correct implementation twenty lines away and stopped.)

THE TRANSFERABLE RULE (016b section 9, "Omitting a key is not neutral"): for any option forwarded to something with its own defaults, guard the send on PRESENCE, never on the value — "if ok", not "if v". And assert on the REQUEST you build, not the response: no assertion on scraped content can distinguish "we sent false" from "Firecrawl kept the footer anyway". That is why payload construction was extracted into buildScrapePayload, which is also the pod-grep marker for this fix, since the change itself added no new string literal.

WATCH: those three steps now receive FULL pages, so their responses are larger. bugs_closed/062 was a Kafka "Message Size Too Large" failure rooted in this same provider file, and it is SILENT to the caller (~12 min of timeout retries). Grep the ADAPTER log, not the workflow. Worst exposure is site-scraper/scrape_site, which sets no formats override. Mitigation is config-only, no roll: set scrape_config.formats.',
'["bugs_closed-101", "bugs_closed-062", "landmine", "council-gate", "scrape", "adapter"]'::jsonb,
'bugs_closed/101 + docs024_key_docs_latest/bugfix_100_101_scrape_provenance/',
'bugsearch-thread');

-- ---------------------------------------------------------------------------
-- store_business_verification — provenance is never a model claim
-- ---------------------------------------------------------------------------
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('action', 'store_business_verification',
'2026-07-28 (bugs_open/100, live on chassis v1.0.1192): this action used to take source_url / source_type / source_name from verification_result — the LLM OUTPUT OBJECT. The prompt never asked for those fields, so all 2,970 rows in business_intel.data_observations were stored with empty provenance, from the table s creation until now.

DO NOT "FIX" THIS BY ASKING THE MODEL FOR THE URL. That makes provenance an assertion generated by the same call that generated the facts, with nothing to check it against — the class this estate was remediated for in July (bugs_closed/043, bugs_closed/061). It is listed as rejected candidate 4 in bugs_open/100 precisely so nobody re-proposes it. The three model reads are DELETED, not demoted to a fallback: a fallback would have restored the old behaviour the moment a model volunteered a plausible-looking URL. A model-supplied source_url is now logged as IGNORED, so prompt drift toward self-reported provenance becomes visible instead of taking effect.

WHERE PROVENANCE COMES FROM NOW: the fetch record the scrape provider already wrote and nobody read — every webscrape provider result carries "url" and "captured_at", set beside the HTTP call. datahelpers.ExtractFetchProvenance reads it; scraped_data is appended to this action s input fields UNCONDITIONALLY rather than left to each definition s input_fields list, because making provenance depend on every caller remembering a config key is what produced 2,970 silent rows.

LANDMINE: the shape it reads (collected_data.scraped_data = {data:{url, captured_at}}) was TRACED THROUGH CODE, not observed — no run carrying scraped_data survives, collection has been off since 2026-03-18 and orchestration_states is on a retention clock. The reader accepts six shapes and logs "no fetch provenance available" when none matches, naming the field it looked in. If provenance comes back empty, read that log line before assuming absence.

CONSTRAINT: migration 257 added data_observations_provenance_not_empty (CHECK, NOT VALID) — an observation that cannot say where it came from can no longer be stored. It is a CHECK and not NOT NULL because source_type was ALREADY NOT NULL and never fired once: the old read produced an empty STRING, not a NULL. NOT VALID leaves the 2,970 historical rows as they are — unsourced and unpublishable, refused by the publishing rule rather than back-filled with invented provenance.

THE CLOSING TEST for bugs_open/100 is two columns, not one: source_url non-empty AND raw_data ? source_url still FALSE. A populated column alone proves the column was written, never BY WHAT.',
'["bugs_open-100", "provenance", "landmine", "council-gate", "business-intel"]'::jsonb,
'bugs_open/100 + docs024_key_docs_latest/bugfix_100_101_scrape_provenance/',
'bugsearch-thread');

COMMIT;

-- VERIFY (expect 3 rows, one per action):
--   SELECT subject_key, created_by, left(body,60) FROM doc_notes
--   WHERE created_by='bugsearch-thread' ORDER BY subject_key;
