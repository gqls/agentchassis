-- 617_aiao_writer_block_managed_with_guidance_carry_HOLD.sql
--
-- ai-agent-orchestration.com — opts `evidence_base` into `writer_block_managed: true`,
-- carrying the site's hand-written prohibitions through the CLM-029 `writer_block_guidance`
-- key so regeneration can never delete them again. Retires migration 611's interim
-- hand-written block: from here the block is COMPOSED from `facts[].writer_line` +
-- `writer_block_guidance` by `composeWriterBlock` (refresh_evidence_base_action.go), and
-- this file writes `writer_block` as the EXACT bytes that composer produces (proven
-- offline — see below), so the first managed refresh finds nothing to change.
--
-- ⚠ `_HOLD`: NOT for the runner. This file is inert-by-name (SIDECAR_RE excludes it) because
-- it must be applied by hand ONLY AFTER the agent-chassis binary carrying CLM-029 has ROLLED.
-- The carry commits are c17a18620 / 14ec48b89 / cbadcba71 (2026-08-25); the chassis live at
-- 2026-08-25 15:27Z was 4c996e1b5cb9b2513d88ec9fe2bae220c38fb6c2 and does NOT contain them
-- (`git merge-base --is-ancestor c17a18620 4c996e1b5` = 1). On that binary
-- `composeWriterBlock` builds from writer_lines + allowed_entities and NOTHING ELSE
-- (measured 2026-08-25 by running the live row through the real function: the output is the
-- seven number lines and no prohibition at all). Flipping the flag before the roll therefore
-- DELETES every NEVER-write ban and the whole NOT TRACKED / NEVER STATE list at the next
-- 09:06Z refresh — the defect that kept 13 of 19 writer_block sites unmanaged (bugs_open/387).
--
-- HOW TO APPLY (RUNBOOK_site_improvement.md R10 has the one-liner):
--   1. LIVE=$(psql -Atc "SELECT git_commit FROM service_binary_capabilities WHERE service='agent-chassis'
--            AND last_seen_at > now()-interval '30 minutes' GROUP BY 1 ORDER BY max(last_seen_at) DESC LIMIT 1")
--   2. git merge-base --is-ancestor c17a18620 "$LIVE"   # MUST exit 0
--   3. psql -v ON_ERROR_STOP=1 -v live_chassis="$LIVE" < this file
--   4. ./scripts/migration/run-migrations.sh --record-only 617_aiao_writer_block_managed_with_guidance_carry_HOLD.sql --note "..."
-- The CHASSIS GUARD below refuses (a) no sha, (b) the known pre-carry sha, (c) a sha that is not
-- the one heartbeating in service_binary_capabilities right now, (d) more than one distinct sha
-- heartbeating (a roll in flight). It CANNOT compute git ancestry — step 2 is yours, and the
-- guard's job is to make you type the sha you actually checked.
--
-- WHAT CHANGES IN THE DOCUMENT
--   writer_block_managed   absent -> true
--   writer_block_guidance  absent -> the prohibitive text below (NEGATIVE/PROHIBITIVE ONLY, per the
--                          CLM-029 contract; 611's list carried verbatim, plus two categories the
--                          site's own banned_claims already enforce: systems shipped for clients,
--                          client sectors served)
--   facts[]                7 -> 8: `aao-architecture`, a VALUELESS capability fact (no `value` key —
--                          EvidenceFact.Value is *float64, so a string value would break every
--                          reader's parse) whose writer_line carries 611's "Architecture:
--                          Kubernetes, Kafka, Postgres — true and stated freely" into the composed
--                          CAPABILITIES section. Positive statements may not live in guidance;
--                          this is the only channel that carries one.
--   writer_block           611's hand-written text -> the composed text (constant below)
--   banned_claims          untouched (guard-asserted)
--   the 7 existing facts   untouched — id, value, writer_line, everything (guard-asserted)
--
-- PROOF THE CONSTANT IS WHAT THE CODE WILL PRODUCE. The proposed document was built by the
-- same jsonb expression this file uses (read-only), then fed to the REAL composeWriterBlock via
-- a `go test -overlay` harness (no file on the shared tree), 2026-08-25. The 1,993-byte output is
-- pasted below unchanged. The first managed refresh (daily ~09:06Z) is the live disconfirming
-- test: expect a fresh `evidence-refresher` row whose writer_block is BYTE-IDENTICAL to this
-- constant. A different block means the prediction was wrong — read the diff, do not "fix" the
-- refresher.
--
-- DURABILITY BEYOND THE FIRST REFRESH (council 35ab8b23 r1, editquality advisory). The landmine that a
-- typed-struct round-trip through EvidenceBase DELETES unlisted keys is real, and it does not bite here:
-- CLM-029's own approval round surveyed every write path — all 9 ParseEvidenceBase callers are readers or
-- validation guards (the admin handler validates through the struct but stores the CLIENT'S bytes), and the
-- two real write paths (this refresher's raw-map marshal; write_site_spec's siteSpecDeepMerge) preserve
-- unknown keys, each pinned by a round-trip test in writer_block_guidance_387_test.go. Precedent: writer_block
-- itself is equally unlisted in the struct and no site has ever lost one. So the key survives the SECOND
-- refresh for the same reason it survives the first. R10 says so anyway: run the survival query after the
-- second ~09:06Z refresh too — that is the first pass that re-reads a refresher-WRITTEN row.
-- CHANGES NO PUBLISHED PAGE. evidence_base is read at write time; nothing re-renders.
-- SUPERSEDE, NOT MUTATE — same shape as 458/557/611/613.
-- ROLLBACK: 617_aiao_writer_block_managed_with_guidance_carry_HOLD_ROLLBACK.sql (restores the exact pre-617 row; managed goes back OFF).

\if :{?live_chassis}
\else
\set live_chassis ''
\endif

BEGIN;

-- CHASSIS_GUARD_BEGIN
SELECT set_config('aiao.live_chassis', :'live_chassis', true);
DO $$
DECLARE passed text; running text; nrun int; started timestamptz;
BEGIN
  passed := current_setting('aiao.live_chassis', true);
  IF passed IS NULL OR passed !~ '^[0-9a-f]{40}$' THEN
    RAISE EXCEPTION '617 REFUSED: pass the RUNNING agent-chassis commit with -v live_chassis=<40-hex sha>, after `git merge-base --is-ancestor c17a18620 <sha>` exits 0 (the CLM-029 carry). Got: %', coalesce(nullif(passed,''), '(unset)');
  END IF;
  IF passed = '4c996e1b5cb9b2513d88ec9fe2bae220c38fb6c2' THEN
    RAISE EXCEPTION '617 REFUSED: % is the PRE-CARRY chassis (live 2026-08-25). Its composeWriterBlock drops writer_block_guidance, so the first refresh would DELETE the NEVER-STATE list. Wait for the roll.', passed;
  END IF;
  SELECT count(DISTINCT git_commit), min(git_commit) INTO nrun, running
    FROM service_binary_capabilities
   WHERE service='agent-chassis' AND last_seen_at > now() - interval '30 minutes';
  IF nrun <> 1 THEN
    RAISE EXCEPTION '617 REFUSED: expected exactly 1 distinct agent-chassis commit heartbeating in the last 30 minutes, found % — a roll may be in flight, or the heartbeat is silent', nrun;
  END IF;
  IF running <> passed THEN
    RAISE EXCEPTION '617 REFUSED: the sha you passed (%) is not the chassis that is RUNNING (%). Re-run the merge-base check against the running one.', passed, running;
  END IF;
  -- NECESSARY, NOT SUFFICIENT (council 35ab8b23 r1, guardian + compliance advisories): a binary that
  -- contains c17a18620 was built after it was committed (2026-08-25 12:49:19Z), so its pods started
  -- after that. This refuses the likeliest mistake — applying before any post-carry roll — for ANY
  -- sha, not just the one hardcoded above. It cannot see a later restart on an OLD image, and it
  -- cannot see a roll that reverted the carry: the merge-base step in R10 is still the operator's.
  SELECT min(started_at) INTO started FROM service_binary_capabilities
   WHERE service='agent-chassis' AND git_commit = passed AND last_seen_at > now() - interval '30 minutes';
  IF started IS NULL OR started < '2026-08-25T12:49:19Z'::timestamptz THEN
    RAISE EXCEPTION '617 REFUSED: the running chassis % started at %, which is BEFORE the CLM-029 carry was committed (c17a18620, 2026-08-25 12:49:19Z) — a binary built before the carry cannot contain it. Wait for the roll.', passed, coalesce(started::text, '(no heartbeat)');
  END IF;
END $$;
-- CHASSIS_GUARD_END

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '617_aiao_writer_block_managed_with_guidance_carry', 'site_specs', ss.id::text, jsonb_build_object('data', ss.data),
       'pre-617 evidence_base for ai-agent-orchestration.com (611 block + 613 writer_lines, possibly refresher-superseded since)'
FROM site_specs ss
WHERE ss.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND ss.aspect='evidence_base' AND ss.is_current;

-- The row we are about to supersede must be the one this file was written against.
DO $$
DECLARE old jsonb; ids text;
BEGIN
  SELECT old_value->'data' INTO old FROM migration_backups WHERE migration_name='617_aiao_writer_block_managed_with_guidance_carry';
  IF old IS NULL THEN RAISE EXCEPTION '617: no backup row written (no current evidence_base?)'; END IF;
  IF old ? 'writer_block_guidance' THEN RAISE EXCEPTION '617: writer_block_guidance already present — someone opted this site in another way; read the row before applying'; END IF;
  IF coalesce((old->>'writer_block_managed')::bool, false) THEN RAISE EXCEPTION '617: writer_block_managed is already true'; END IF;
  IF jsonb_array_length(old->'facts') <> 7 THEN RAISE EXCEPTION '617: expected the 7 facts this file was written against, found %', jsonb_array_length(old->'facts'); END IF;
  SELECT string_agg(f->>'id', ',' ORDER BY f->>'id') INTO ids FROM jsonb_array_elements(old->'facts') f;
  IF ids <> 'aao-agent-definitions,aao-agent-types,aao-departments,aao-live-sites,aao-orchestrations,aao-services,aao-work-items' THEN
    RAISE EXCEPTION '617: fact ids differ from the census this file was written against: %', ids;
  END IF;
  IF (old->>'writer_block') !~ 'NOT TRACKED / DOES NOT EXIST, NEVER STATE' OR (old->>'writer_block') ~ '\mNNN\M' THEN
    RAISE EXCEPTION '617: the current writer_block is not the 611 block this file retires';
  END IF;
END $$;

UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;

INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
SELECT ss.site_id, ss.aspect,
  (ss.data
   || jsonb_build_object('writer_block_managed', true)
   || jsonb_build_object('writer_block_guidance', $G$NEVER copy a number out of a page, a template, or an older spec: the NUMBERS list above is the ONLY authority for figures about this business, and if a count is not written there, write no count at all. NEVER restate a listed floor as any other number. NEVER write letters as stand-ins for digits (a letter repeated where a number belongs): a stand-in is a defect that ships, not a placeholder. NEVER write "over 70 specialised AI agents", "70+ agents" or "30+ agent types" — true as lower bounds and understating the fleet by more than half (owner ruling 2026-07-27). NEVER frame the 8 departments as departments of external clients ("departments served"): they are the platform's OWN organisational taxonomy. NEVER state an exact daily orchestration figure, and NEVER state a total of automated work items completed. NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, "departments served", satisfaction rates, awards won, concurrent-instance counts ("thousands of concurrent instances" is not measured), uptime percentages, systems shipped for clients, client sectors served. None of these are measured; every such figure at any value is an invention.$G$::text)
   || jsonb_build_object('facts', ss.data->'facts' || jsonb_build_array(jsonb_build_object(
        'id','aao-architecture',
        'kind','capability',
        'claim','the platform runs on Kubernetes, Kafka and Postgres',
        'source', jsonb_build_object('attested_by','the platform''s own deployment manifests (deployments/kustomize) and the Go services'' Kafka and Postgres clients — a self-description of the stack, no figure'),
        'verified_at','2026-08-25',
        'writer_line','Architecture: Kubernetes, Kafka, Postgres — true and stated freely')))
   || jsonb_build_object('writer_block', $WB$NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine):
- more than 150 active agent definitions in the production registry
- more than 150 distinct agent types
- 8 departments — Strategy, Research, Content, Design, Development, Quality, Operations, Data (the platform's OWN taxonomy, never 'departments served')
- more than 20 live sites in production, built and operated end-to-end by the platform
- 17 backend services
- automated work items completed: state NO figure — the ledger is reaped, so this count falls as well as rises
- over a thousand orchestrations a day (never an exact daily figure — a rolling 24-hour window is stale within hours)

CAPABILITIES (assert without inventing numbers):
- Architecture: Kubernetes, Kafka, Postgres — true and stated freely

NEVER copy a number out of a page, a template, or an older spec: the NUMBERS list above is the ONLY authority for figures about this business, and if a count is not written there, write no count at all. NEVER restate a listed floor as any other number. NEVER write letters as stand-ins for digits (a letter repeated where a number belongs): a stand-in is a defect that ships, not a placeholder. NEVER write "over 70 specialised AI agents", "70+ agents" or "30+ agent types" — true as lower bounds and understating the fleet by more than half (owner ruling 2026-07-27). NEVER frame the 8 departments as departments of external clients ("departments served"): they are the platform's OWN organisational taxonomy. NEVER state an exact daily orchestration figure, and NEVER state a total of automated work items completed. NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, "departments served", satisfaction rates, awards won, concurrent-instance counts ("thousands of concurrent instances" is not measured), uptime percentages, systems shipped for clients, client sectors served. None of these are measured; every such figure at any value is an invention.$WB$::text)),
  ss.source, ss.source_agent,
  'superseded by 617: writer_block_managed=true with the CLM-029 writer_block_guidance carry; writer_block pre-composed to the exact bytes composeWriterBlock produces; 611''s interim block retired',
  true, '617_migration'
FROM site_specs ss
WHERE ss.site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND ss.aspect='evidence_base' AND ss.superseded_at IS NOT NULL
ORDER BY ss.superseded_at DESC LIMIT 1;

DO $$
DECLARE cur jsonb; old jsonb; n int; ph text; arch jsonb;
BEGIN
  SELECT old_value->'data' INTO old FROM migration_backups WHERE migration_name='617_aiao_writer_block_managed_with_guidance_carry';
  SELECT count(*) INTO n FROM site_specs WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;
  IF n <> 1 THEN RAISE EXCEPTION '617: expected 1 current row, found %', n; END IF;
  SELECT data INTO cur FROM site_specs WHERE site_id='2a8ebf9c-20a2-4c39-b191-840b012371da' AND aspect='evidence_base' AND is_current;

  IF NOT coalesce((cur->>'writer_block_managed')::bool, false) THEN RAISE EXCEPTION '617: writer_block_managed not true'; END IF;
  IF (cur->>'writer_block_guidance') IS DISTINCT FROM $G$NEVER copy a number out of a page, a template, or an older spec: the NUMBERS list above is the ONLY authority for figures about this business, and if a count is not written there, write no count at all. NEVER restate a listed floor as any other number. NEVER write letters as stand-ins for digits (a letter repeated where a number belongs): a stand-in is a defect that ships, not a placeholder. NEVER write "over 70 specialised AI agents", "70+ agents" or "30+ agent types" — true as lower bounds and understating the fleet by more than half (owner ruling 2026-07-27). NEVER frame the 8 departments as departments of external clients ("departments served"): they are the platform's OWN organisational taxonomy. NEVER state an exact daily orchestration figure, and NEVER state a total of automated work items completed. NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, "departments served", satisfaction rates, awards won, concurrent-instance counts ("thousands of concurrent instances" is not measured), uptime percentages, systems shipped for clients, client sectors served. None of these are measured; every such figure at any value is an invention.$G$ THEN RAISE EXCEPTION '617: writer_block_guidance is not the constant this file carries'; END IF;
  IF (cur->>'writer_block') IS DISTINCT FROM $WB$NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine):
- more than 150 active agent definitions in the production registry
- more than 150 distinct agent types
- 8 departments — Strategy, Research, Content, Design, Development, Quality, Operations, Data (the platform's OWN taxonomy, never 'departments served')
- more than 20 live sites in production, built and operated end-to-end by the platform
- 17 backend services
- automated work items completed: state NO figure — the ledger is reaped, so this count falls as well as rises
- over a thousand orchestrations a day (never an exact daily figure — a rolling 24-hour window is stale within hours)

CAPABILITIES (assert without inventing numbers):
- Architecture: Kubernetes, Kafka, Postgres — true and stated freely

NEVER copy a number out of a page, a template, or an older spec: the NUMBERS list above is the ONLY authority for figures about this business, and if a count is not written there, write no count at all. NEVER restate a listed floor as any other number. NEVER write letters as stand-ins for digits (a letter repeated where a number belongs): a stand-in is a defect that ships, not a placeholder. NEVER write "over 70 specialised AI agents", "70+ agents" or "30+ agent types" — true as lower bounds and understating the fleet by more than half (owner ruling 2026-07-27). NEVER frame the 8 departments as departments of external clients ("departments served"): they are the platform's OWN organisational taxonomy. NEVER state an exact daily orchestration figure, and NEVER state a total of automated work items completed. NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, "departments served", satisfaction rates, awards won, concurrent-instance counts ("thousands of concurrent instances" is not measured), uptime percentages, systems shipped for clients, client sectors served. None of these are measured; every such figure at any value is an invention.$WB$ THEN RAISE EXCEPTION '617: writer_block is not the pre-composed constant'; END IF;
  IF (cur->'banned_claims') IS DISTINCT FROM (old->'banned_claims') THEN RAISE EXCEPTION '617: banned_claims changed; this file must not touch enforcement'; END IF;

  -- The seven existing facts are carried byte-identically; the eighth is the only addition.
  IF jsonb_array_length(cur->'facts') <> 8 THEN RAISE EXCEPTION '617: expected 8 facts, found %', jsonb_array_length(cur->'facts'); END IF;
  IF ((cur->'facts') - 7) IS DISTINCT FROM (old->'facts') THEN RAISE EXCEPTION '617: one of the seven existing facts changed'; END IF;
  arch := cur->'facts'->7;
  IF arch->>'id' <> 'aao-architecture' OR arch ? 'value' OR arch->>'kind' <> 'capability' OR NOT (arch->'source' ? 'attested_by') THEN
    RAISE EXCEPTION '617: the architecture fact is malformed (must be valueless, kind=capability, attested): %', arch;
  END IF;

  -- No stand-in token anywhere the writer reads (the 557 defect's shape).
  IF (cur->>'writer_block') ~ '\m(NNN|XXX|NN|YYY)\M' OR (cur->>'writer_block_guidance') ~ '\m(NNN|XXX|NN|YYY)\M' THEN
    RAISE EXCEPTION '617: a letter stand-in for digits is present';
  END IF;

  -- Every prohibition 611 carried must be present in the NEW block (the whole point of the carry).
  FOREACH ph IN ARRAY ARRAY[
    'NEVER write "over 70 specialised AI agents"', '30+ agent types', 'departments served',
    'letters as stand-ins', 'exact daily orchestration', 'automated work items completed',
    'NOT TRACKED / DOES NOT EXIST, NEVER STATE', 'clients served', 'satisfaction rates', 'awards won',
    'thousands of concurrent instances', 'uptime percentages', 'Kubernetes, Kafka, Postgres',
    'more than 150 active agent definitions', 'more than 150 distinct agent types', '8 departments',
    'more than 20 live sites', '17 backend services'] LOOP
    IF position(ph IN (cur->>'writer_block')) = 0 THEN
      RAISE EXCEPTION '617: the composed writer_block lost a phrase 611 carried: %', ph;
    END IF;
  END LOOP;

  RAISE NOTICE '617 OK: writer_block_managed=true, guidance carried, 8 facts (7 untouched + valueless aao-architecture), writer_block = the pre-composed constant. NEXT: after the ~09:06Z refresh, expect a fresh evidence-refresher row with writer_block BYTE-IDENTICAL to this constant (RUNBOOK R10).';
END $$;

COMMIT;
