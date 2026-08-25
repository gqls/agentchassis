-- 626_bug386_arm_fact_history_on_fundamentallyai_counters.sql
--
-- bugs_open/386 Phase 2: arm fact-value HISTORY (CLM-028) on the five
-- fundamentallyai.com counting facts that are the bug's motivating case.
--
-- CONTEXT. The daily `evidence-freshness` sweep overwrites a counting fact's
-- value and discards the reading it just held
-- (refresh_evidence_base_action.go:534-544). Every already-deployed page still
-- rendering yesterday's number is then convicted by numberSupported as an
-- `unregistered_number` at ERROR severity (validate_page_content.go:1324-1326;
-- error is failing at :420 and returns without saving at :449) — so a page
-- whose only fault is being a day old cannot be rebuilt to fix itself, and in
-- the bugs_open/033 queue it is indistinguishable from a real fabrication.
--
-- CLM-028 shipped the memory: an opt-in `retain_history` flag plus a capped
-- `history` array, consulted by numberSupported as an EXACT match inside the
-- existing context-term gate. The Go is live — chassis v1.0.1339, verified at
-- BOTH replicas' /proc/1/exe as a capability probe for `retain_history`, with a
-- present-control (`context_terms`, found) and an absent-control (found on
-- neither). It is armed on nothing: zero facts fleet-wide carry the key
-- [MEASURED 2026-08-25]. This file arms the first five.
--
-- WHY THESE FIVE, AND WHY THE OWNER'S "at least N" RULING DOES NOT REACH THEM.
-- The owner ruled 2026-08-25 (bug file §4b) that a counting fact should be
-- expressed as "at least N" or the claim cancelled — and §4b explicitly
-- preserves this mechanism for the case the ruling cannot reach. F9-F13 have
-- NO `writer_line`, so they contribute nothing to composeWriterBlock: their
-- numbers reach the page through an `evidence-chart` component rendering the
-- register into content_data, not through prose. There is no sentence to
-- rephrase, only a chart — and the chart stamps its own "verified 2026-08-23",
-- so "Feed items collected 11513 verified 2026-08-23" is TRUE for ever. The
-- register simply cannot currently agree with it.
--
-- THE HISTORY IS DERIVED, NOT TYPED. Every value written below is computed in
-- this file from the site's own SUPERSEDED site_specs rows. No number is
-- hardcoded, so this migration structurally CANNOT seed a value the register
-- never held — which is the honesty property the whole mechanism rests on. The
-- derivation applies exactly the rules recordFactHistory applies at runtime:
-- consecutive duplicates collapsed (2026-08-16 carries two refreshes), the
-- fact's CURRENT value excluded (numberSupported checks Value first), oldest
-- first, capped at 90 (datahelpers.FactHistoryMaxEntries).
--
-- Expected result, measured 2026-08-25 by running the derivation read-only:
--   F9-feed-items-collected     30 entries   7321 - 11646
--   F10-feed-items-scored       30 entries   6166 - 10416
--   F11-council-rounds-revise   31 entries    108 - 437
--   F12-council-rounds-approved 30 entries     37 - 503
--   F13-council-rounds-rejected 12 entries      9 - 23
--
-- THE PRE-ARMING CONTROL, run before writing this file (Phase 2's gate, and
-- the council's `compliance` seat asked for the sharper version of it):
--   * fresh export asserted 91 rows against 91 in the DB, stderr empty;
--   * BASELINE (live register): 5 findings, all capabilities/evidence-chart —
--     11513, 10194, 428, 483, 23;
--   * ARMED (this file's derivation applied offline): 0 findings;
--   * disappeared: exactly those 5, EVERY ONE on the register-rendering
--     component, ZERO in free-text prose. A disappearing PROSE finding would
--     have been the accidental-support signal and would have stopped this file;
--   * appeared: none.
--
-- AND THE NEGATIVE CONTROL, which is what makes the above mean anything. Zero
-- findings after arming makes "nothing else disappeared" trivially true — the
-- site had nothing else to lose — so the diff alone cannot distinguish "spares
-- stale renders" from "spares everything". A value the register has NEVER held
-- was injected into the same component and rescanned against the armed
-- register: 11513 -> 11514 (the archive runs ... 11373, 11513, 11646, 11828).
-- STILL FLAGGED. One away from a genuinely-held value, identical context
-- window, caught. Arming vouches for the values the register held and nothing
-- adjacent to them.
--
-- WHAT HAPPENS NEXT (observables, so verification is a query and not a hope):
--   * The next `evidence-freshness` tick that MOVES one of these five values
--     will append the outgoing reading itself (recordFactHistory) and report
--     it in siteRefreshResult.FactsHistoryRecorded. That is the demand control
--     for the WRITER half, which this file does not exercise: this file seeds
--     history, it does not prove the code writes it.
--   * The five findings above should be absent from the next claimscan / the
--     next `check_unverified_claims` pass on fundamentallyai.
--   * A rebuild of `capabilities` should no longer be refused for those five.
--
-- SCOPE, deliberately narrow (council `guardian` advisory, corr 18dba069):
-- arming is a one-fact-at-a-time reviewed operator act, never a bulk toggle.
-- This file arms FIVE facts on ONE site, each named, on the strength of the
-- controls above. It does not arm the fleet and must not be copied to do so.
--
-- ORDERING. None. The reading code has been live since v1.0.1339; data is live
-- the moment this applies. Rollback is supersede-and-reinsert without the keys
-- (no binary depends on their presence).
--
-- Supersede-then-insert rather than UPDATE, following 585/270 and
-- writeRefreshedEvidenceBase, because idx_site_specs_current is UNIQUE on
-- (site_id, aspect) WHERE is_current — the two statements must be one
-- transaction, in this order. The superseded row is kept as history.

BEGIN;

-- Guard 1: refuse if already applied (message matches the runner probe's
-- /already/i detection).
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
    FROM site_specs, jsonb_array_elements(data->'facts') f
    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
      AND aspect = 'evidence_base' AND is_current
      AND (f ? 'retain_history' OR f ? 'history');
    IF n <> 0 THEN
        RAISE EXCEPTION 'bug386 arming: % fact(s) already carry retain_history/history — this file is already applied.', n;
    END IF;
END $$;

-- Guard 2: the register must be in the state this file was written against.
DO $$
DECLARE nfacts int; narmed int;
BEGIN
    SELECT jsonb_array_length(data->'facts') INTO nfacts
    FROM site_specs
    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
      AND aspect = 'evidence_base' AND is_current;
    IF nfacts IS DISTINCT FROM 16 THEN
        RAISE EXCEPTION 'bug386 arming: expected 16 facts in the register, found %. The register has moved — re-run the pre-arming control before applying.', coalesce(nfacts::text, '(null)');
    END IF;

    SELECT count(*) INTO narmed
    FROM site_specs, jsonb_array_elements(data->'facts') f
    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
      AND aspect = 'evidence_base' AND is_current
      AND f->>'id' IN ('F9-feed-items-collected','F10-feed-items-scored',
                       'F11-council-rounds-revise','F12-council-rounds-approved',
                       'F13-council-rounds-rejected')
      AND f ? 'value';
    IF narmed <> 5 THEN
        RAISE EXCEPTION 'bug386 arming: expected the 5 named number-bearing facts, found %. Re-read before applying.', narmed;
    END IF;
END $$;

-- Guard 3: every one of the five must HAVE derivable history, and none may
-- exceed the cap. A fact with zero derived entries would be armed to no
-- purpose and signals the superseded archive has been pruned since the
-- pre-arming control ran; an over-cap fact would mean the derivation's LIMIT
-- has drifted from datahelpers.FactHistoryMaxEntries.
DO $$
DECLARE bad text;
BEGIN
    WITH base AS (
        SELECT id, created_at, data FROM site_specs
         WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
           AND aspect = 'evidence_base' AND is_current
    ),
    cur AS (
        SELECT f->>'id' AS fid, (f->>'value')::numeric AS curval
          FROM base, jsonb_array_elements(base.data->'facts') f
         WHERE f ? 'value'
    ),
    raw AS (
        SELECT f->>'id' AS fid, (f->>'value')::numeric AS v, ss.created_at,
               lag((f->>'value')::numeric) OVER (PARTITION BY f->>'id' ORDER BY ss.created_at) AS prev
          FROM site_specs ss, jsonb_array_elements(ss.data->'facts') f
         WHERE ss.site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
           AND ss.aspect = 'evidence_base' AND NOT ss.is_current AND f ? 'value'
    ),
    kept AS (
        SELECT r.fid, count(*) AS n
          FROM raw r JOIN cur c ON c.fid = r.fid
         WHERE (r.prev IS NULL OR r.v <> r.prev) AND r.v <> c.curval
         GROUP BY r.fid
    )
    SELECT string_agg(t.fid || '=' || coalesce(k.n, 0)::text, ', ')
      INTO bad
      FROM (VALUES ('F9-feed-items-collected'),('F10-feed-items-scored'),
                   ('F11-council-rounds-revise'),('F12-council-rounds-approved'),
                   ('F13-council-rounds-rejected')) AS t(fid)
      LEFT JOIN kept k ON k.fid = t.fid
     WHERE coalesce(k.n, 0) = 0 OR k.n > 90;

    IF bad IS NOT NULL THEN
        RAISE EXCEPTION 'bug386 arming: derived history is empty or over the 90 cap for: %. Re-run the pre-arming control before applying.', bad;
    END IF;
END $$;

-- 1. supersede the current row (kept as history)
UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
  AND aspect = 'evidence_base' AND is_current;

-- 2. reinsert with retain_history + derived history on the five named facts.
--    Everything else — writer_block, the other eleven facts, every unknown
--    key — carried verbatim. Fact ORDER preserved via WITH ORDINALITY.
WITH armed(fid) AS (
    -- The named five, and ONLY these. Without this filter the derivation below
    -- happily produces history for every fact whose value has ever moved —
    -- F1-live-sites and F2-council-seats among them — and the INSERT would arm
    -- seven facts on a file that says five. Caught by this file's own VERIFY
    -- block during the runner's doomed-transaction probe, before any write.
    -- Arming is a one-fact-at-a-time reviewed act (council `guardian`, corr
    -- 18dba069); a derivation that silently widens the set is exactly the bulk
    -- toggle that advisory warns against.
    VALUES ('F9-feed-items-collected'), ('F10-feed-items-scored'),
           ('F11-council-rounds-revise'), ('F12-council-rounds-approved'),
           ('F13-council-rounds-rejected')
),
base AS (
    SELECT * FROM site_specs
     WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
       AND aspect = 'evidence_base'
       AND is_current = false
       AND superseded_at IS NOT NULL
     ORDER BY superseded_at DESC
     LIMIT 1
),
cur AS (
    SELECT f->>'id' AS fid, (f->>'value')::numeric AS curval
      FROM base, jsonb_array_elements(base.data->'facts') f
     WHERE f ? 'value'
),
raw AS (
    SELECT f->>'id' AS fid, (f->>'value')::numeric AS v,
           COALESCE(f->>'verified_at', ss.created_at::date::text) AS va,
           ss.created_at,
           lag((f->>'value')::numeric) OVER (PARTITION BY f->>'id' ORDER BY ss.created_at) AS prev
      FROM site_specs ss, jsonb_array_elements(ss.data->'facts') f, base b
     WHERE ss.site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
       AND ss.aspect = 'evidence_base'
       AND NOT ss.is_current
       AND ss.id <> b.id                    -- exclude the row we just superseded
       AND f ? 'value'
       AND f->>'id' IN (SELECT fid FROM armed)   -- the named five, and only these
),
kept AS (
    SELECT r.fid, r.v, r.va, r.created_at,
           row_number() OVER (PARTITION BY r.fid ORDER BY r.created_at DESC) AS rn_desc
      FROM raw r JOIN cur c ON c.fid = r.fid
     WHERE (r.prev IS NULL OR r.v <> r.prev)   -- collapse consecutive duplicates
       AND r.v <> c.curval                     -- Value is checked first; don't duplicate it
),
hist AS (
    SELECT fid,
           jsonb_agg(jsonb_build_object('value', v, 'verified_at', va)
                     ORDER BY created_at) AS entries
      FROM kept
     WHERE rn_desc <= 90                       -- FactHistoryMaxEntries, newest kept
     GROUP BY fid
)
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, pinned, notes)
SELECT
    b.site_id,
    'evidence_base',
    jsonb_set(
        b.data,
        '{facts}',
        (
            SELECT jsonb_agg(
                CASE WHEN h.entries IS NOT NULL THEN
                    jsonb_set(
                        jsonb_set(f, '{retain_history}', 'true'::jsonb),
                        '{history}', h.entries
                    )
                ELSE f END
                ORDER BY ord
            )
            FROM jsonb_array_elements(b.data->'facts') WITH ORDINALITY AS t(f, ord)
            LEFT JOIN hist h ON h.fid = f->>'id'
        )
    ),
    'manual',
    NULL,
    'session-2026-08-25-bug386-arm-fact-history',
    b.pinned,
    'bugs_open/386 Phase 2 / CLM-028: arms retain_history on the five fundamentallyai counting facts (F9-F13) whose stale renders the daily refresh was convicting at error severity. History DERIVED from this site''s own superseded evidence_base rows, so no value here is one the register never held. Pre-arming control: baseline 5 findings all on the register-rendering component, armed 0, zero prose findings lost; negative control 11513->11514 still flagged. Chassis v1.0.1339 carries the reader (capability-probed at both replicas with present- and absent-controls).'
FROM base b;

-- Verify INSIDE the transaction, as a DO block that can actually abort. A
-- verify block made of SELECTs cannot stop a COMMIT: ON_ERROR_STOP does not
-- fire on a non-empty result set (LANDMINES, RFC_006).
DO $$
DECLARE narmed int; nhist int; nfacts int; leaked int;
BEGIN
    SELECT jsonb_array_length(data->'facts') INTO nfacts
    FROM site_specs
    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
      AND aspect = 'evidence_base' AND is_current;
    IF nfacts <> 16 THEN
        RAISE EXCEPTION 'bug386 arming VERIFY: fact count changed to % — the rewrite dropped or duplicated facts.', nfacts;
    END IF;

    SELECT count(*) INTO narmed
    FROM site_specs, jsonb_array_elements(data->'facts') f
    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
      AND aspect = 'evidence_base' AND is_current
      AND (f->>'retain_history')::boolean IS TRUE;
    IF narmed <> 5 THEN
        RAISE EXCEPTION 'bug386 arming VERIFY: expected 5 armed facts, found %.', narmed;
    END IF;

    SELECT count(*) INTO nhist
    FROM site_specs, jsonb_array_elements(data->'facts') f
    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
      AND aspect = 'evidence_base' AND is_current
      AND jsonb_typeof(f->'history') = 'array'
      AND jsonb_array_length(f->'history') BETWEEN 1 AND 90;
    IF nhist <> 5 THEN
        RAISE EXCEPTION 'bug386 arming VERIFY: expected 5 facts with a 1..90-entry history, found %.', nhist;
    END IF;

    -- No fact outside the named five may have been armed.
    SELECT count(*) INTO leaked
    FROM site_specs, jsonb_array_elements(data->'facts') f
    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
      AND aspect = 'evidence_base' AND is_current
      AND (f ? 'retain_history' OR f ? 'history')
      AND f->>'id' NOT IN ('F9-feed-items-collected','F10-feed-items-scored',
                           'F11-council-rounds-revise','F12-council-rounds-approved',
                           'F13-council-rounds-rejected');
    IF leaked <> 0 THEN
        RAISE EXCEPTION 'bug386 arming VERIFY: % fact(s) outside the named five were armed.', leaked;
    END IF;

    -- The five values the bug is about must now be present in history.
    SELECT count(*) INTO nhist
    FROM site_specs, jsonb_array_elements(data->'facts') f, jsonb_array_elements(f->'history') h
    WHERE site_id = '199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
      AND aspect = 'evidence_base' AND is_current
      AND (f->>'id', (h->>'value')::numeric) IN (
            ('F9-feed-items-collected', 11513), ('F10-feed-items-scored', 10194),
            ('F11-council-rounds-revise', 428), ('F12-council-rounds-approved', 483),
            ('F13-council-rounds-rejected', 23));
    IF nhist <> 5 THEN
        RAISE EXCEPTION 'bug386 arming VERIFY: expected the 5 convicted values to be present in history, found %. The page''s figures are not the ones now vouched for.', nhist;
    END IF;
END $$;

COMMIT;
