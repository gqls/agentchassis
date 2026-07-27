-- 228_evidence_facts_for_vonc.sql
-- Owed item (b) of the fabricated-stats lane: give vonc.com a machine-readable
-- facts[] register. DB-only, effective immediately — no image roll needed.
--
-- ── WHY VONC AND NOT THE OTHER FOUR ────────────────────────────────────────
-- The handoff listed webdesign.co.uk (98 pages), finetuning.uk (42),
-- gaswholesalers.com (31) and dartsonline.com (24) as owing registers. A survey
-- run 2026-07-27 with the REAL scanner (ExtractAssertionText +
-- ScanUnregisteredNumbers against an empty register, i.e. every business-shaped
-- number surfaces) says most of that work should NOT be done:
--
--   site                 components   number claims found
--   gaswholesalers.com          102   0
--   dartsonline.com              17   1   (a 30-day returns window — a policy term)
--   finetuning.uk               139   5   (4 are audience descriptors + a privacy
--                                          age limit; ONE is a real claim, below)
--   webdesign.co.uk             101   15  (ALL 15 on page_type='guide' pages)
--   vonc.com                     49   0 in prose — but 14 in stat FIELDS
--
-- webdesign.co.uk is the instructive one: every flagged number is a worked
-- example inside teaching content — "10,000 monthly visitors", "100 concurrent
-- users", "random 502 Bad Gateway errors", "Try rating Sci-Fi '5'" — on
-- learn-algorithms-bayesian-theory, learn-operations-scaling and four more
-- guides. A register there would manufacture 15 false positives and evidence
-- nothing, because the claims layer has no notion of page_type and cannot tell
-- a guide's hypothetical from an about page's assertion.
--
-- So vonc is the one site in that list whose figures are BOTH real claims and
-- verifiable. Its exposure is entirely in stat fields, which is exactly the
-- surface bugs_open/093 just added a second call site for.
--
-- ── THE HARD RAIL, UNCHANGED FROM 218 ──────────────────────────────────────
-- A fact enters the register only from a live query or an explicit owner
-- attestation, NEVER by transcribing site copy — that copy is what may be
-- fabricated. Every value below is a count of rows in `pages`, re-derived
-- 2026-07-27, and each fact carries the query that proves it:
--
--   SELECT page_type, count(*) FROM pages p JOIN sites s ON s.id=p.site_id
--   WHERE s.domain='vonc.com' AND p.status='active' GROUP BY 1;
--     entity-page 8 | content 3 | tool 3 | blog-post 2 | landing 1 | section-index 1
--   -- 18 active pages in total
--
-- ── WHAT THIS REGISTER DOES TO THE LIVE FINDINGS ───────────────────────────
-- Measured BEFORE applying, by running the shipping extractor + ScanStatClaims
-- over vonc's live content_data with this exact candidate register:
--
--   before (facts=0):  14 findings, ALL `low` — i.e. "could not check any of it"
--   after  (facts=4):   2 findings, both `medium` — a machine checked and they failed
--
-- The two survivors are the real defect, and they are both on ONE component:
--
--   page  position  component            stat_1               stat_2
--   about        2  content-block-about  Archetypes    = 3    Tools Live         = 8   <-- SWAPPED
--   about        6  gauntlet-cta         Tools live    = 3    Archetypes in play = 8
--   index        3  gauntlet-cta         Archetypes    = 8    Tools Live         = 3
--   index        4  brief-explanation    Interactive tools live = 3  Written guides live = 2
--
-- Three components agree with the database; `content-block-about` has the two
-- values transposed. That correction is 229, deliberately a SEPARATE file: this
-- one adds evidence, that one changes published copy, and they should be
-- revertable independently.
--
-- ── NOTE ON `pages live` TOLERANCE ─────────────────────────────────────────
-- 18 is registered `gte` with context terms, not `exact`, because a page count
-- moves and understating it is harmless — the live copy says "Pages Live 17",
-- which is supported (17 <= 18) and stays supported as pages are added. The
-- three set-counts are `exact`: there are eight named archetypes, and "7
-- archetypes" would be wrong, not modest. Per the EvidenceFact contract a
-- non-exact tolerance REQUIRES context_terms, or a gte fact blanket-supports
-- every smaller number on the site.

BEGIN;

DO $seed$
DECLARE
    v_site   uuid;
    v_facts  jsonb;
    n        int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'vonc.com';
    IF v_site IS NULL THEN
        RAISE EXCEPTION '228: no site row for vonc.com';
    END IF;

    -- Pre-condition: the row must already exist and already parse as opted-in
    -- via banned_claims. If it does not, this file is being run against a
    -- different world than the one it was written for.
    SELECT count(*) INTO n
    FROM site_specs
    WHERE site_id = v_site AND aspect = 'evidence_base' AND is_current
      AND jsonb_array_length(COALESCE(data->'banned_claims','[]'::jsonb)) = 9;
    IF n <> 1 THEN
        RAISE EXCEPTION '228: expected exactly 1 current evidence_base row for vonc with 9 banned_claims, found %', n;
    END IF;

    v_facts := $f$[
      {"id":"vonc-archetypes","kind":"count","claim":"archetypes in the World","value":8,
       "source":{"sql":"SELECT count(*) FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='vonc.com' AND p.page_type='entity-page' AND p.status='active'"},
       "tolerance":"exact","verified_at":"2026-07-27",
       "writer_line":"{value} archetypes: Catalyst, Judge, Maker, Mentor, Oracle, Scout, Surgeon, Wildcard (one entity-page each)",
       "context_terms":["archetype"]},
      {"id":"vonc-tools","kind":"count","claim":"interactive tools live","value":3,
       "source":{"sql":"SELECT count(*) FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='vonc.com' AND p.page_type='tool' AND p.status='active'"},
       "tolerance":"exact","verified_at":"2026-07-27",
       "writer_line":"{value} interactive tools live (pages.page_type='tool', status='active')",
       "context_terms":["tool"]},
      {"id":"vonc-guides","kind":"count","claim":"written guides live","value":2,
       "source":{"sql":"SELECT count(*) FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='vonc.com' AND p.page_type='blog-post' AND p.status='active'"},
       "tolerance":"exact","verified_at":"2026-07-27",
       "writer_line":"{value} written guides live (pages.page_type='blog-post', status='active')",
       "context_terms":["guide"]},
      {"id":"vonc-pages","kind":"count","claim":"pages live","value":18,
       "source":{"sql":"SELECT count(*) FROM pages p JOIN sites s ON s.id=p.site_id WHERE s.domain='vonc.com' AND p.status='active'"},
       "tolerance":"gte","verified_at":"2026-07-27",
       "writer_line":"{value} pages live in total (pages.status='active')",
       "context_terms":["pages live","pages built","page live"]}
    ]$f$::jsonb;

    -- jsonb_set, NOT `data || '{"facts":…}'`: the landmine here is that
    -- `NULL || jsonb` is NULL and jsonb_set(x, path, NULL) returns NULL, either
    -- of which would null the whole register. The `data ? 'facts'` guard makes
    -- the path's existence a precondition rather than an assumption.
    UPDATE site_specs
    SET data = jsonb_set(data, '{facts}', v_facts), updated_at = now()
    WHERE site_id = v_site AND aspect = 'evidence_base' AND is_current
      AND data ? 'facts';

    GET DIAGNOSTICS n = ROW_COUNT;
    IF n <> 1 THEN
        RAISE EXCEPTION '228: expected to update 1 evidence_base row, updated %', n;
    END IF;
END $seed$;

-- ── Post-conditions ────────────────────────────────────────────────────────
DO $post$
DECLARE
    v_facts int;
    v_banned int;
    v_keys  int;
BEGIN
    SELECT jsonb_array_length(ss.data->'facts'),
           jsonb_array_length(ss.data->'banned_claims'),
           (SELECT count(*) FROM jsonb_object_keys(ss.data))
      INTO v_facts, v_banned, v_keys
    FROM site_specs ss JOIN sites s ON s.id = ss.site_id
    WHERE ss.aspect='evidence_base' AND ss.is_current AND s.domain='vonc.com';

    IF v_facts <> 4 THEN
        RAISE EXCEPTION '228: expected 4 facts after seeding, found %', v_facts;
    END IF;
    -- The whole point of jsonb_set over a rewrite: nothing else may be lost.
    IF v_banned <> 9 THEN
        RAISE EXCEPTION '228: banned_claims went from 9 to % — the merge clobbered them', v_banned;
    END IF;
    IF v_keys <> 7 THEN
        RAISE EXCEPTION '228: evidence_base went from 7 top-level keys to % (writer_block/allowed_entities/governing_rule/schema_notes/audit_doc must survive)', v_keys;
    END IF;
    RAISE NOTICE '228 OK: vonc.com evidence_base now has % facts, % banned_claims, % keys', v_facts, v_banned, v_keys;
END $post$;

COMMIT;
