-- 765 — copywriter-directory-researcher: a fourth producer for the global verified-claim
-- register (DIR-001), kind `copywriter`. Owner instruction 2026-09-04: "Add a copywriter kind
-- to the global register first then extend the business directory to a copywriting vertical."
--
-- _HOLD: config-only and it APPLIES cleanly on the current binary, but it is held so the owner
-- picks the moment — it adds a shared producer, and its first run should be supervised the way
-- the finance kinds' first runs were (DIR-001, 2026-08-15).
--
-- WHAT THIS IS, AND WHAT IT IS NOT. General-purpose research already exists and is live:
-- `evidence-researcher` (same six-step chain, prompt takes any `research_query`) and
-- `research-agent`. What is per-kind is ONLY the DIRECTORY REGISTRATION leg: the extraction
-- prompt must emit `entity_kind` plus a field vocabulary, because `directory_claims.go:177`
-- reads `entity_kind` from the candidate the model returns (default "model"). So this file is
-- a PROMPT, wrapped in the workflow its three siblings already share — not a new mechanism.
-- (A generator that writes this prompt for any kind is the owner's "research agent creator"
-- idea and is proposed separately; this file is the worked instance it would generate.)
--
-- CLONED FROM THE LIVE ROW, not from the seed file: the workflow is copied verbatim from
-- `finance-directory-researcher`'s current `default_config` and only the `extract_claims`
-- prompt is replaced, so the six steps, their wiring and their configs cannot drift from the
-- sibling that is known to work.
--
-- FIELD VOCABULARY (open, deliberately): `directory_claims.go` closes the vocabulary only for
-- the three finance kinds ("Kinds absent from this map … have no closed vocabulary and are
-- unaffected"). Copywriting is not a regulated financial promotion, so no allowlist is added
-- and NO Go CHANGE IS NEEDED for registration. The prompt still states its own field list, so
-- the register stays legible.
--
-- ORGANISATIONS ONLY — the owner's ruling of 2026-09-03 evening, and it is in the prompt as a
-- hard rule: a listed entity is a BUSINESS with a public website, never a named individual.
-- That is what makes the removal question answerable and keeps the listing off people who
-- never asked to be on it.
--
-- AFTER APPLYING: 766 retargets and re-enables the scheduled task (kept separate so the
-- research does not start the moment this row lands).

BEGIN;

DO $g$
DECLARE n int;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='finance-directory-researcher' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF n <> 1 THEN RAISE EXCEPTION '765 REFUSED: expected exactly 1 active finance-directory-researcher to clone, found %', n; END IF;
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type='copywriter-directory-researcher' AND deleted_at IS NULL;
  IF n <> 0 THEN RAISE EXCEPTION '765 REFUSED: copywriter-directory-researcher already exists (% rows)', n; END IF;
  SELECT count(*) INTO n FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s(k,v)
   WHERE ad.type='finance-directory-researcher' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
     AND s.k='extract_claims' AND s.v->'config' ? 'prompt_template';
  IF n <> 1 THEN RAISE EXCEPTION '765 REFUSED: the template row has no extract_claims prompt_template to replace'; END IF;
END $g$;

INSERT INTO agent_definitions (
    type, display_name, description, category, agent_category, status, is_active,
    image_repository, image_tag, input_contract, output_contract, default_config
)
SELECT
    'copywriter-directory-researcher',
    'Copywriter Directory Researcher',
    'Copywriting-supplier directory acquisition lane (DIR-001, kind copywriter): researches UK copywriting ORGANISATIONS — agencies, studios, consultancies and marketplaces — and extracts atomic, citable, non-price candidate claims (specialisms, sectors served, supplier type, location, established year, website) for the global register. Organisations only, never named individuals (owner ruling 2026-09-03). Cloned from finance-directory-researcher; only the extraction prompt differs.',
    src.category, src.agent_category, 'active', true,
    src.image_repository, src.image_tag,
    '{"required": ["research_query"]}'::jsonb,
    '{"produces": {"registration": "verified UK copywriting-organisation claims added to the directory register under kind copywriter; rejects raised for human review"}}'::jsonb,
    jsonb_set(
      src.default_config,
      '{workflow,steps,extract_claims,config,prompt_template}',
      to_jsonb($p$You are extracting ATOMIC, CITABLE facts about UK COPYWRITING ORGANISATIONS for a public cited directory.

Research question: {{.input_data.research_query}}

SCRAPED SOURCES (each has a url):
{{.scrape_results}}

Extract up to 15 candidate claims, one per (organisation, fact) pair. At most ONE claim per (organisation, field) pair — if several passages state the same field, emit only the single most complete enumeration: a later duplicate OVERWRITES the earlier one at registration, so a weaker duplicate destroys a stronger claim.

TWO HARD RULES. A claim that breaks either is worthless and must not be emitted.

1. AN ENTITY IS ONE NAMED ORGANISATION WITH ITS OWN PUBLIC WEBSITE — an agency, studio, consultancy or marketplace. NEVER a named individual or sole trader trading under their own personal name; never a sector, a market segment, a "type of provider", a ranking list, or an article about copywriting. If the source is a listicle ("the 10 best copywriting agencies"), extract the ORGANISATIONS it names, never the list itself. If you cannot name the organisation and point at its own website, emit nothing for it.

2. EVERY CLAIM CARRIES A VERBATIM QUOTE FROM THE SOURCE THAT STATES IT, and the url of that source. No inference, no summary in place of a quote, no fact assembled from two passages.

Each claim MUST give:
- entity_slug: lowercase-hyphenated, e.g. 'stratton-craig'; entity_name: display name, e.g. 'Stratton Craig'; entity_owner: the parent group if the source names one, else same as entity_name;
- entity_kind: EXACTLY the string "copywriter". Never any other value.
- field: ONLY from this list — specialisms (what kinds of copy: web, ads, email, product, tone-of-voice, technical, B2B), sectors_served, supplier_type (exactly one of "agency", "studio", "consultancy", "marketplace"), location (town or city in the UK), established_year, team_size, website;
- value: the fact itself, concise and self-contained;
- quote: the verbatim sentence from the source that states it;
- source_url: the url of the source that carries that quote.

NEVER extract prices, day rates, per-word rates or quotes for work — they go stale and a stale price under a named firm misrepresents that firm. NEVER extract testimonials, awards claimed without a named awarding body, or marketing superlatives ("the best", "leading", "award-winning") — they are not facts.

Return a JSON array of claim objects and nothing else. An empty array is the correct answer when the sources name no qualifying organisation.$p$::text)
    )
FROM agent_definitions src
WHERE src.type='finance-directory-researcher' AND src.is_active AND COALESCE(src.is_snapshot,false)=false AND src.deleted_at IS NULL;

DO $v$
DECLARE cfg jsonb; p text; n int;
BEGIN
  SELECT default_config INTO cfg FROM agent_definitions
   WHERE type='copywriter-directory-researcher' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
  IF cfg IS NULL THEN RAISE EXCEPTION '765 VERIFY: the new row was not created'; END IF;
  -- the six steps came across intact
  SELECT count(*) INTO n FROM jsonb_object_keys(cfg->'workflow'->'steps') k;
  IF n <> 6 THEN RAISE EXCEPTION '765 VERIFY: expected 6 steps, found %', n; END IF;
  IF (cfg->'workflow'->'steps'->'verify_and_register'->>'action') <> 'verify_and_register_directory_claims'
     OR (cfg->'workflow'->'steps'->'search_web'->>'action') <> 'web_search'
  THEN RAISE EXCEPTION '765 VERIFY: the workflow wiring did not clone'; END IF;
  -- the prompt is OURS, not the template's
  p := cfg #>> '{workflow,steps,extract_claims,config,prompt_template}';
  IF position('UK COPYWRITING ORGANISATIONS' in p) = 0 THEN RAISE EXCEPTION '765 VERIFY: copywriter prompt not installed'; END IF;
  IF position('FINANCIAL SERVICES' in p) > 0 THEN RAISE EXCEPTION '765 VERIFY: the finance prompt survived the replace'; END IF;
  IF position('"copywriter"' in p) = 0 THEN RAISE EXCEPTION '765 VERIFY: entity_kind literal missing from the prompt'; END IF;
  -- the search step still reads the task's query the way its siblings do
  IF (cfg #>> '{workflow,steps,search_web,config,query_from}') <> 'input_data.research_query'
  THEN RAISE EXCEPTION '765 VERIFY: search_web query_from is not input_data.research_query'; END IF;
END $v$;

COMMIT;
