-- 557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate.sql
--
-- ai-agent-orchestration.com. Unblocks the `pricing` rebuild, which has failed twice
-- on a claims error, and fixes the reason at source rather than at the page.
--
-- OWNER INSTRUCTION (2026-08-19): "Update the claim to 196 agents at source."
--
-- ⚠ EXECUTED IN SUBSTANCE, NOT LITERALLY, AND HERE IS WHY — writing "196" would have
-- reproduced the bug with a different number. Three measurements, all taken before
-- this file was written:
--
--   1. THE FACTS ARE LIVE SQL AND HAVE ALREADY MOVED. `aao-agent-definitions` and
--      `aao-agent-types` are `SELECT count(*) FROM agent_definitions …`, re-run on
--      every evidence refresh. They read 175/174 on 2026-07-26, 196 when this lane
--      reported to the owner on 2026-08-19, and **199 on 2026-08-22**. Any literal
--      committed to a spec is wrong within days. That is the whole defect, and
--      hardcoding 196 would re-arm it.
--
--   2. "170+" WAS NEVER FALSE. Both facts carry `tolerance: "gte"`, and
--      `datahelpers.numberSupported` accepts any asserted value at or below the
--      registered one (`claims.go:1007`). 170 ≤ 199, so the NUMBER passed.
--
--   3. THE REJECTION WAS THE CONTEXT GATE, NOT THE NUMBER. `numberSupported` skips a
--      fact entirely unless one of its `context_terms` appears in the window around
--      the number (`claims.go:990-1001`). The generated sentence was
--      "…drawing on a registry of 170+ agents already built and running in
--      production." The fact's terms are "agent definition", "specialised ai agent",
--      "agents in the registry", "ai agents" — and "a registry of 170+ agents" is
--      none of them. So no fact was eligible to vouch for a number both facts would
--      have accepted, and the page was refused.
--
-- THE ROOT CAUSE, in one line: **this document instructs the writer to produce a
-- phrase that its own facts cannot validate.** `writer_block` says, twice, in as many
-- words, Write "170+ agents" / Write "170+ agent types" — while the fact that must
-- support those sentences does not list either wording as a context term. Every page
-- that obeys the instruction is refused by the checker; the only way to pass was to
-- disobey it.
--
-- WHY THE BLOCK IS STALE AT ALL, and this is the part worth carrying: `writer_block`
-- is HAND-WRITTEN here and nothing regenerates it. `refresh_evidence_base_action`
-- rebuilds it from each fact's `writer_line` with `{value}` interpolated live, but
-- **only where the site sets `writer_block_managed: true`** (`:474`). This site never
-- opted in, so the facts have refreshed 40+ times while the prose the writer actually
-- reads has been frozen since 2026-07-27 — still announcing "175 as of 2026-07-26"
-- and "14 live sites" against live values of 199 and 25.
--
-- ⚠ AND OPTING IN IS NOT SAFE YET — DELIBERATELY NOT DONE HERE. `composeWriterBlock`
-- (`:996`) builds the block from `writer_line`s and `allowed_entities` and NOTHING
-- ELSE. Setting the flag today would silently delete, from what the writer sees:
--   · the two NEVER-write bans (70+ agents, 30+ agent types — owner ruling 07-27);
--   · the whole NOT TRACKED / NEVER STATE list (clients served, departments served,
--     satisfaction rates, awards, uptime, concurrent instances);
--   · "DO NOT state an exact daily figure" for orchestrations (rolling 24h window);
--   · "DO NOT state a figure" for work items (a reaped ledger that FALLS as well as
--     rises — 1,051 on 07-26, 6,918 on 08-22).
-- The bans survive as `banned_claims` regexes, so ENFORCEMENT would hold; what would
-- be lost is PREVENTION, which turns a silent success into a refusal loop. The two
-- "do not state" cautions are not enforced anywhere and would simply vanish — and
-- both facts carry a `writer_line` that would then actively invite the figure.
-- Closing that gap is `composeWriterBlock` learning to carry negative guidance;
-- until it does, this block stays hand-written and this file keeps it correct.
--
-- WHAT THIS MIGRATION DOES — two changes, both at source, neither weakening a control:
--
--   A. Rewrites `writer_block` so it contains NO literal count. Every figure is
--      delegated to the facts list (the only thing that refreshes), and the two
--      agent lines now mandate a wording the checker recognises. All bans and all
--      NEVER-STATE guidance are carried forward verbatim in meaning.
--
--   B. Adds four natural paraphrases to `aao-agent-definitions.context_terms` and one
--      to `aao-agent-types`, so the checker recognises how a writer actually phrases
--      this claim instead of only the four wordings someone happened to list.
--
-- ⚠ (B) IS DELIBERATELY PHRASE-BASED, NOT THE BARE WORD "agents". Adding "agents"
-- would make every "N agents" sentence eligible for a `gte` fact of 199 — including a
-- client-delivery claim ("we built 40 agents for a customer"), which is a DIFFERENT
-- claim this site does not measure and which the number checker would then wave
-- through. The added terms all describe the platform's own registry.
--
-- DOES NOT CHANGE ANY PUBLISHED PAGE. `evidence_base` is an instruction set consumed
-- at write time. Nothing re-renders because of this file; the next page build reads
-- the new block. The live `about` page's own stale "170+ agent types" sentence is NOT
-- touched — it predates the checker and closes when that page is next rebuilt.
--
-- SUPERSEDE, DO NOT MUTATE — `site_specs` carries `is_current`/`superseded_at` and a
-- partial unique index on (site_id, aspect) WHERE is_current. Same shape as 458.
--
-- ROLLBACK: 557_..._ROLLBACK.sql (restores the exact prior row from migration_backups)

BEGIN;

-- 0. Byte-exact backup of the row being superseded.
INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate',
       'site_specs', ss.id::text, jsonb_build_object('data', ss.data),
       'pre-557 evidence_base for ai-agent-orchestration.com'
FROM site_specs ss
WHERE ss.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND ss.aspect='evidence_base' AND ss.is_current;

-- 1. Close the current row.
UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND aspect='evidence_base' AND is_current;

-- 2. Insert the successor: new writer_block, widened context terms, facts otherwise
--    untouched (their VALUES are owned by the refresh action, never by this file).
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT
  ss.site_id,
  ss.aspect,
  jsonb_set(
    jsonb_set(
      ss.data,
      '{writer_block}',
      to_jsonb(
        'NUMBERS — the facts list in this document is the ONLY authority for figures about this ' ||
        'business, and every count in it is a LIVE query re-run on each refresh. Never copy a ' ||
        'number out of a page, a template, or an older spec, and never repeat a number written ' ||
        'here in prose: read the current value from the facts list. State it as a lower bound ' ||
        '("NNN+") so it cannot go stale between writing and publication.' || E'\n' ||
        '- active agent definitions in the production registry: take the live value from the ' ||
        '`aao-agent-definitions` fact. Phrase it as "NNN+ AI agents" or "NNN+ agents in the ' ||
        'registry" — those wordings are the ones the evidence checker recognises for this fact. ' ||
        'NEVER write "over 70 specialised AI agents" or "70+ agents": true as a lower bound but ' ||
        'understating the fleet by more than half (owner ruling 2026-07-27).' || E'\n' ||
        '- distinct agent types: take the live value from the `aao-agent-types` fact, phrased as ' ||
        '"NNN+ agent types". NEVER write "30+ agent types" — understates by more than five times ' ||
        '(owner ruling 2026-07-27).' || E'\n' ||
        '- 8 departments — Strategy, Research, Content, Design, Development, Quality, Operations, ' ||
        'Data. This is the platform''s OWN organisational taxonomy (a self-description); never ' ||
        'frame it as departments of external clients ("departments served").' || E'\n' ||
        '- live sites in production, built and operated end-to-end by the platform: take the live ' ||
        'value from the `aao-live-sites` fact. (A literal "14" stood here from 2026-07-26 until ' ||
        '2026-08-22 while the real figure reached 25. That is why no count is written in this ' ||
        'block any more.)' || E'\n' ||
        '- 17 backend services (the service manifests under deployments/kustomize/services; the ' ||
        'directory also holds frontend and infra overlays, so a bare directory count is NOT this ' ||
        'figure).' || E'\n' ||
        '- orchestrations per day: "over a thousand orchestrations a day" is true and may be ' ||
        'stated. DO NOT state an exact daily figure — it is a ROLLING 24-hour window, so any ' ||
        'number you write is stale within hours and will be rejected as unsupported.' || E'\n' ||
        '- automated work items completed: DO NOT state a figure. The ledger is reaped, so this ' ||
        'count FALLS as well as rises (1,051 on 2026-07-26; 6,918 on 2026-08-22); a total that ' ||
        'can go down is misleading however it is phrased.' || E'\n' ||
        '- Architecture: Kubernetes, Kafka, Postgres — true and stated freely.' || E'\n' ||
        'NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, "departments served", ' ||
        'satisfaction rates, awards won, concurrent-instance counts ("thousands of concurrent ' ||
        'instances" is not measured), uptime percentages. None of these are measured; every such ' ||
        'figure at any value is an invention.'
      )
    ),
    '{facts}',
    (
      SELECT jsonb_agg(
        CASE
          WHEN f->>'id' = 'aao-agent-definitions' THEN
            jsonb_set(f, '{context_terms}',
              (f->'context_terms') ||
              '["agents already built","agents in production","agents organised into","registry of"]'::jsonb)
          WHEN f->>'id' = 'aao-agent-types' THEN
            jsonb_set(f, '{context_terms}',
              (f->'context_terms') || '["agent types in production"]'::jsonb)
          ELSE f
        END
        ORDER BY ord
      )
      FROM jsonb_array_elements(ss.data->'facts') WITH ORDINALITY AS t(f, ord)
    )
  ),
  ss.source, ss.source_agent,
  'superseded by migration 557: writer_block delegated to the facts list (no literal counts) and '
    || 'agent context_terms widened so the checker recognises the wording the block mandates',
  true,
  '557_migration'
FROM site_specs ss
WHERE ss.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND ss.aspect='evidence_base' AND ss.superseded_at IS NOT NULL
ORDER BY ss.superseded_at DESC
LIMIT 1;

-- 3. Guard the exact post-condition. A verify block of bare SELECTs cannot stop a
--    COMMIT (ON_ERROR_STOP ignores a non-empty result), so this must RAISE.
DO $$
DECLARE
  cur          jsonb;
  n_current    int;
  n_facts_old  int;
  n_facts_new  int;
  defs_terms   int;
  types_terms  int;
  backed_up    int;
BEGIN
  SELECT count(*) INTO backed_up FROM migration_backups
   WHERE migration_name='557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate';
  IF backed_up <> 1 THEN
    RAISE EXCEPTION '557: expected 1 backup row, wrote % — refusing without a restore path', backed_up;
  END IF;

  SELECT count(*) INTO n_current FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
  IF n_current <> 1 THEN
    RAISE EXCEPTION '557: expected exactly 1 current evidence_base row, found %', n_current;
  END IF;

  SELECT data INTO cur FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;

  -- No fact may be lost or duplicated by the rebuild of the array.
  SELECT jsonb_array_length(old_value->'data'->'facts') INTO n_facts_old
    FROM migration_backups
   WHERE migration_name='557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate';
  n_facts_new := jsonb_array_length(cur->'facts');
  IF n_facts_new <> n_facts_old THEN
    RAISE EXCEPTION '557: fact count changed % -> % — the array rebuild dropped or duplicated a fact', n_facts_old, n_facts_new;
  END IF;

  -- THE POINT OF THE FILE: no literal count may remain in the writer_block. These are
  -- the three that were there (175, 174, 14) plus the two mandated phrasings.
  IF cur->>'writer_block' ~ '\m(175|174)\M' THEN
    RAISE EXCEPTION '557: writer_block still carries a stale hardcoded registry count';
  END IF;
  IF cur->>'writer_block' ~ 'Write "170\+' THEN
    RAISE EXCEPTION '557: writer_block still mandates the 170+ wording the facts cannot validate';
  END IF;
  IF cur->>'writer_block' !~ 'NNN\+ AI agents' THEN
    RAISE EXCEPTION '557: writer_block does not name a wording the checker recognises';
  END IF;

  -- Every ban carried forward. These are the protective half and losing them silently
  -- is the exact hazard that keeps writer_block_managed switched off here.
  IF cur->>'writer_block' !~ 'NEVER write "over 70' THEN
    RAISE EXCEPTION '557: the 70+ agents ban was lost from writer_block';
  END IF;
  IF cur->>'writer_block' !~ 'NEVER write "30\+ agent types"' THEN
    RAISE EXCEPTION '557: the 30+ agent types ban was lost from writer_block';
  END IF;
  IF cur->>'writer_block' !~ 'NOT TRACKED / DOES NOT EXIST, NEVER STATE' THEN
    RAISE EXCEPTION '557: the NEVER-STATE list was lost from writer_block';
  END IF;
  IF cur->>'writer_block' !~ 'DO NOT state an exact daily figure' THEN
    RAISE EXCEPTION '557: the orchestrations-per-day caution was lost from writer_block';
  END IF;
  IF cur->>'writer_block' !~ 'automated work items completed: DO NOT state a figure' THEN
    RAISE EXCEPTION '557: the work-items caution was lost from writer_block';
  END IF;

  -- banned_claims untouched.
  IF jsonb_array_length(cur->'banned_claims') <>
     (SELECT jsonb_array_length(old_value->'data'->'banned_claims') FROM migration_backups
       WHERE migration_name='557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate') THEN
    RAISE EXCEPTION '557: banned_claims count changed; this file must not touch enforcement';
  END IF;

  -- The context terms the rejection turned on.
  SELECT count(*) INTO defs_terms FROM jsonb_array_elements_text(
    (SELECT f->'context_terms' FROM jsonb_array_elements(cur->'facts') f WHERE f->>'id'='aao-agent-definitions')) t
   WHERE t IN ('agents already built','agents in production','agents organised into','registry of');
  IF defs_terms <> 4 THEN
    RAISE EXCEPTION '557: expected 4 new context terms on aao-agent-definitions, found %', defs_terms;
  END IF;

  SELECT count(*) INTO types_terms FROM jsonb_array_elements_text(
    (SELECT f->'context_terms' FROM jsonb_array_elements(cur->'facts') f WHERE f->>'id'='aao-agent-types')) t
   WHERE t = 'agent types in production';
  IF types_terms <> 1 THEN
    RAISE EXCEPTION '557: expected the new context term on aao-agent-types, found %', types_terms;
  END IF;

  -- Values are the refresh action's to own, not this file's.
  IF (SELECT f->>'value' FROM jsonb_array_elements(cur->'facts') f WHERE f->>'id'='aao-agent-definitions')
     IS DISTINCT FROM
     (SELECT f->>'value' FROM jsonb_array_elements(
        (SELECT old_value->'data'->'facts' FROM migration_backups
          WHERE migration_name='557_aiao_evidence_base_stops_mandating_a_phrase_its_own_facts_cannot_validate')) f
       WHERE f->>'id'='aao-agent-definitions') THEN
    RAISE EXCEPTION '557: a fact VALUE changed; this file must never write a count';
  END IF;

  RAISE NOTICE '557 OK: writer_block carries no literal count and mandates a recognised wording; 5 context terms added; facts, values and banned_claims untouched. Nothing re-renders — the next page build reads it.';
END $$;

COMMIT;
