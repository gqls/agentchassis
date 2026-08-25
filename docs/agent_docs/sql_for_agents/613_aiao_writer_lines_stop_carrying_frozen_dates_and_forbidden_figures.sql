-- 613_aiao_writer_lines_stop_carrying_frozen_dates_and_forbidden_figures.sql
--
-- ai-agent-orchestration.com — repairs the `writer_line` half of `evidence_base`,
-- which migration `611` deliberately did not touch and which the `bugs_open/387`
-- session flagged back to this lane as ours to fix ("only you know the intended
-- phrasing"). Latent today, because `writer_block_managed` is NOT set on this site;
-- this file is a precondition for ever setting it, not a live-defect fix.
--
-- ⚠ THIS LANE OWNS THE DEFECT THAT MADE IT NECESSARY. Migration `557` (mine) put the
-- exemplar 'Phrase it as "NNN+ AI agents"' into `writer_block` and instructed the
-- writer to "take the live value from the fact" — from a list the unscoped writer
-- prompt never contains. **`NNN+` reached the public**: live on
-- `model-directory.html` on 2026-08-25, *"…against the NNN+ agent types already
-- running in production…"*, verified by curl before this file was written. Measured
-- by the 387 session: 137 instructed writer calls since 08-22, **14** copied `NNN`
-- verbatim, **0** wrote the value. `611` fixed the block (floors, and an explicit ban
-- on letter stand-ins). This file fixes the sibling field before the same shape can
-- reach the public by the other route. WRONG_CALLS logged.
--
-- ── WHAT `writer_line` IS, AND WHY A BAD ONE IS NOT INERT FOR EVER ──────────────
-- `refresh_evidence_base_action.composeWriterBlock` (`:996`) REGENERATES
-- `writer_block` from each fact's `writer_line`, substituting `{value}` from the live
-- fact — but only where the site sets `writer_block_managed: true` (`:474`). This site
-- has never opted in, which is why these lines have been wrong without doing harm.
-- **The moment anyone flips that flag, every line below is published.**
--
-- ── THREE DEFECTS, AND THE CENSUS FOUND MORE THAN WAS REPORTED ──────────────────
-- The 387 CONTRIB named two frozen dates. Censusing all 7 writer_lines found a third,
-- and a second defect class it did not name:
--
--   (a) FROZEN DATE BESIDE A LIVE VALUE — `{value}` is substituted at regeneration
--       time, the date is literal text. Managed mode would publish a TRUE number
--       under a FALSE date, which is worse than either error alone because it reads
--       as provenance.
--         aao-agent-definitions   "… ({value} as of 2026-07-26)"
--         aao-agent-types         "… ({value} as of 2026-07-26)"
--         aao-orchestrations      "… ({value} in the 24 hours to 2026-07-26)"   <- NOT reported
--       The value has moved 175 → 196 → 199 → **200** in a month; the date has not.
--
--   (b) A LINE THAT WOULD PUBLISH WHAT THE BLOCK FORBIDS. `writer_block` says, in as
--       many words, *"orchestrations per day: DO NOT state an exact daily figure"* and
--       *"automated work items completed: DO NOT state a figure"*. Both facts carry a
--       `writer_line` that substitutes exactly that figure. **Under managed mode the
--       block would contradict its own standing instruction**, and the contradiction
--       would be generated, so no author would ever see themselves write it.
--       `aao-work-items` is the sharper case: the ledger is REAPED, so the count falls
--       as well as rises (1,051 on 07-26; 6,918 on 08-22) — a published total that can
--       go DOWN.
--
--   (c) A LIVE COUNT WHERE THE BLOCK NOW CHOOSES A FLOOR. `611` moved this site's
--       stated counts to rounded floors, on the reasoning that a floor stays true as
--       the value rises and so cannot go stale between writing and publication. Three
--       writer_lines still emit live counts for facts the block now floors. Managed
--       mode would silently undo `611`'s decision.
--
-- ── WHAT THIS FILE DOES: five lines, aligned to the block `611` installed ────────
--   aao-agent-definitions  -> "more than 150 active agent definitions in the production registry"
--   aao-agent-types        -> "more than 150 distinct agent types"
--   aao-live-sites         -> "more than 20 live sites in production, built and operated end-to-end by the platform"
--   aao-orchestrations     -> "over a thousand orchestrations a day (never an exact daily figure — a rolling 24-hour window is stale within hours)"
--   aao-work-items         -> "automated work items completed: state NO figure — the ledger is reaped, so this count falls as well as rises"
--
-- ⚠ `{value}` IS REMOVED FROM ALL FIVE, DELIBERATELY. Keeping it and only fixing the
-- date would leave a live count in prose, which is the failure this site has now been
-- bitten by twice (a literal "14 live sites" stood in the block from 07-26 to 08-22
-- while the real figure reached 25). A floor needs no date and cannot go stale upward.
--
-- ── DELIBERATELY NOT TOUCHED ────────────────────────────────────────────────────
--   aao-departments  "{value} departments — Strategy, Research, …"  — 8 is a
--     SELF-DESCRIPTION of a fixed taxonomy that is enumerated in the same sentence, so
--     the number cannot drift away from the list beside it.
--   aao-services     "{value} backend services"                     — the block states
--     it as an exact figure with its own provenance note, and `611` left that as-is.
--   Changing either would be re-litigating `611`'s judgement on facts it decided to
--   leave exact. If they later drift, that is a separate, evidenced change.
--
-- ⚠ DOES NOT ENABLE MANAGED MODE, AND MUST NOT BE READ AS CLEARING THE WAY FOR IT.
-- The blocker recorded in `557` and this lane's NOTES stands: `composeWriterBlock`
-- builds the block from `writer_line`s and `allowed_entities` and NOTHING ELSE, so
-- flipping the flag still deletes both NEVER-write bans and the whole NOT TRACKED /
-- NEVER STATE list. The 387 session has proposed the missing `writer_block_guidance`
-- carry to the `bugs_open/288` lane, which owns that file. **This file removes ONE of
-- several preconditions.**
--
-- CHANGES NO PUBLISHED PAGE and files nothing. `evidence_base` is read at write time;
-- nothing re-renders because of this migration.
--
-- SUPERSEDE, NOT MUTATE — `site_specs` carries `is_current`/`superseded_at` and a
-- partial unique index on (site_id, aspect) WHERE is_current. Same shape as 458/557/611.
--
-- ⚠ CARRIES `611` FORWARD UNTOUCHED. `611` landed at 2026-08-25 11:20:26Z, ~45 minutes
-- before this file. It is the active incident fix and this migration must not disturb
-- it: `writer_block` and `banned_claims` are copied verbatim and guard-asserted
-- byte-identical below.
--
-- ROLLBACK: 613_..._ROLLBACK.sql (restores the exact prior row from migration_backups)

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '613_aiao_writer_lines_stop_carrying_frozen_dates_and_forbidden_figures',
       'site_specs', ss.id::text, jsonb_build_object('data', ss.data),
       'pre-613 evidence_base (the 611 row) for ai-agent-orchestration.com'
FROM site_specs ss
WHERE ss.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND ss.aspect='evidence_base' AND ss.is_current;

UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT ss.site_id, ss.aspect,
  jsonb_set(ss.data, '{facts}', (
    SELECT jsonb_agg(
      CASE f->>'id'
        WHEN 'aao-agent-definitions' THEN jsonb_set(f,'{writer_line}',
          to_jsonb('more than 150 active agent definitions in the production registry'::text))
        WHEN 'aao-agent-types' THEN jsonb_set(f,'{writer_line}',
          to_jsonb('more than 150 distinct agent types'::text))
        WHEN 'aao-live-sites' THEN jsonb_set(f,'{writer_line}',
          to_jsonb('more than 20 live sites in production, built and operated end-to-end by the platform'::text))
        WHEN 'aao-orchestrations' THEN jsonb_set(f,'{writer_line}',
          to_jsonb('over a thousand orchestrations a day (never an exact daily figure — a rolling 24-hour window is stale within hours)'::text))
        WHEN 'aao-work-items' THEN jsonb_set(f,'{writer_line}',
          to_jsonb('automated work items completed: state NO figure — the ledger is reaped, so this count falls as well as rises'::text))
        ELSE f
      END ORDER BY ord)
    FROM jsonb_array_elements(ss.data->'facts') WITH ORDINALITY AS t(f, ord)
  )),
  ss.source, ss.source_agent,
  'superseded by 613: writer_lines lose frozen dates and stop emitting figures the writer_block forbids; aligned to 611''s floors',
  true, '613_migration'
FROM site_specs ss
WHERE ss.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND ss.aspect='evidence_base' AND ss.superseded_at IS NOT NULL
ORDER BY ss.superseded_at DESC LIMIT 1;

DO $$
DECLARE
  cur jsonb; old jsonb; n int; bad int;
BEGIN
  SELECT old_value->'data' INTO old FROM migration_backups
   WHERE migration_name='613_aiao_writer_lines_stop_carrying_frozen_dates_and_forbidden_figures';
  IF old IS NULL THEN RAISE EXCEPTION '613: no backup row written'; END IF;

  SELECT count(*) INTO n FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION '613: expected 1 current row, found %', n; END IF;

  SELECT data INTO cur FROM site_specs
   WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;

  -- 611 must be carried forward byte-identically. This file fixes the sibling field.
  IF (cur->>'writer_block') IS DISTINCT FROM (old->>'writer_block') THEN
    RAISE EXCEPTION '613: writer_block changed; 611 is the active incident fix and must be carried forward untouched';
  END IF;
  IF (cur->'banned_claims') IS DISTINCT FROM (old->'banned_claims') THEN
    RAISE EXCEPTION '613: banned_claims changed; this file must not touch enforcement';
  END IF;

  -- No fact lost, none duplicated, no VALUE written by this file.
  IF jsonb_array_length(cur->'facts') <> jsonb_array_length(old->'facts') THEN
    RAISE EXCEPTION '613: fact count changed % -> %',
      jsonb_array_length(old->'facts'), jsonb_array_length(cur->'facts');
  END IF;
  SELECT count(*) INTO bad
    FROM jsonb_array_elements(cur->'facts') c
    JOIN jsonb_array_elements(old->'facts') o ON o->>'id' = c->>'id'
   WHERE (c->>'value') IS DISTINCT FROM (o->>'value');
  IF bad <> 0 THEN
    RAISE EXCEPTION '613: % fact value(s) changed; values belong to refresh_evidence_base_action, never to a migration', bad;
  END IF;

  -- THE POINT: no writer_line may carry a frozen date, and none of the five may
  -- substitute a value any more.
  SELECT count(*) INTO bad FROM jsonb_array_elements(cur->'facts') f
   WHERE f->>'writer_line' ~ '20[0-9]{2}-[0-9]{2}-[0-9]{2}';
  IF bad <> 0 THEN
    RAISE EXCEPTION '613: % writer_line(s) still carry a frozen date', bad;
  END IF;

  SELECT count(*) INTO bad FROM jsonb_array_elements(cur->'facts') f
   WHERE f->>'id' IN ('aao-agent-definitions','aao-agent-types','aao-live-sites',
                      'aao-orchestrations','aao-work-items')
     AND f->>'writer_line' ~ '\{value\}';
  IF bad <> 0 THEN
    RAISE EXCEPTION '613: % of the five repaired writer_line(s) still substitute {value}', bad;
  END IF;

  -- The two deliberately-left-alone lines must be UNCHANGED — a silent widening of
  -- this file's scope is exactly what the comment above disclaims.
  SELECT count(*) INTO bad
    FROM jsonb_array_elements(cur->'facts') c
    JOIN jsonb_array_elements(old->'facts') o ON o->>'id' = c->>'id'
   WHERE c->>'id' IN ('aao-departments','aao-services')
     AND (c->>'writer_line') IS DISTINCT FROM (o->>'writer_line');
  IF bad <> 0 THEN
    RAISE EXCEPTION '613: % writer_line(s) outside the declared scope were changed', bad;
  END IF;

  -- And no stand-in token may be introduced by this file — the 557 defect's shape.
  SELECT count(*) INTO bad FROM jsonb_array_elements(cur->'facts') f
   WHERE f->>'writer_line' ~ '\m(NNN|XXX|NN|YYY)\M';
  IF bad <> 0 THEN
    RAISE EXCEPTION '613: % writer_line(s) contain a letter stand-in for digits — the exact defect 557 shipped', bad;
  END IF;

  RAISE NOTICE '613 OK: 5 writer_lines repaired (3 frozen dates, 2 forbidden figures), 611 carried forward byte-identical, values untouched. Managed mode is STILL NOT SAFE — composeWriterBlock must carry negative guidance first.';
END $$;

COMMIT;
