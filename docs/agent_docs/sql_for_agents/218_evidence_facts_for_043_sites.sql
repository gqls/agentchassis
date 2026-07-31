-- 218_evidence_facts_for_043_sites.sql
-- bugs_open/043: give the three sites that actually fabricated a MACHINE-READABLE
-- register, so the claims machinery stops being a silent no-op on exactly them.
-- DB-only; effective immediately for the existing checkers, and it is the
-- opt-in signal the new content_data stat audit keys on once that image rolls.
--
-- ── THE FINDING THIS FIXES ─────────────────────────────────────────────────
-- ParseEvidenceBase (datahelpers/claims.go:109) returns (nil, nil) when a row
-- carries no facts[] AND no banned_claims[]. The rows seeded on 2026-07-24 for
-- 043's four treated sites carry ONLY a writer_block. So:
--
--   validate_page_content check 8  -> `if eb := loadEvidenceBase(...); eb != nil`
--   discovery check_unverified_claims -> same predicate
--
-- ...have both been NO-OPS on robot-hands, gamesdesign and ai-agent-orchestration
-- since the day they were "protected". The writer_block half worked (it is read
-- straight out of site_specs by the prompt template, not through
-- ParseEvidenceBase), which is why the writer stopped inventing while the
-- CHECKERS stayed blind. Verified 2026-07-26:
--
--   SELECT s.domain, jsonb_array_length(COALESCE(ss.data->'facts','[]'::jsonb))
--   FROM site_specs ss JOIN sites s ON s.id=ss.site_id
--   WHERE ss.aspect='evidence_base' AND ss.is_current;
--   -- robot-hands 0, gamesdesign 0, ai-agent-orchestration 0, vonc 0
--
-- ── EVERY FIGURE RE-DERIVED, AND MOST OF THEM HAD MOVED ────────────────────
-- The hard rail for this lane: a fact enters the register only from a live
-- query or an explicit owner attestation, NEVER by transcribing site copy —
-- that copy is what may be fabricated. So each writer_block figure was re-run
-- against the live database on 2026-07-26 before being written as a fact.
-- Almost none of them still held (all measured, not inferred):
--
--   claim                          writer_block (07-24)   live (07-26)
--   active agent definitions       170                    175
--   distinct agent types           165                    174
--   live sites                     13                     14
--   work items completed           1,267                  1,051   <-- WENT DOWN
--   orchestrations per day         1,699                  1,834
--   robot-hands spec figures       39                     59
--
-- Two things follow, and both are recorded in bugs_open/043:
--
--   (a) A frozen snapshot is not evidence, it is a fact with an expiry nobody
--       wrote down. Every fact below therefore carries `source.sql` — the query
--       that DEFINES its meaning — so refresh_evidence_base can re-run it and
--       keep the value current. The number in this file is a starting value.
--
--   (b) "work items completed" is NOT MONOTONIC: 1,267 on 07-24, 1,051 today,
--       because the ledger is reaped. A cumulative-sounding achievement stat
--       that can go DOWN is misleading whatever value it holds, and
--       ai-agent-orchestration.com/index currently publishes the stale 1,267.
--       It is registered here with tolerance `gte` so the audit FLAGS the
--       overstatement rather than blessing it. Correcting the page needs a
--       re-render, which bugs_open/029's dispatch stall is currently blocking.
--
-- ── TOLERANCES ────────────────────────────────────────────────────────────
--   gte   — a growing platform count. numberSupported treats val <= fact as
--           supported, which is exactly "a dated snapshot up to the live count
--           is fine" and still catches the OVERSTATEMENT that 043 is about.
--   exact — a catalogue or design constant, where the precise figure is the
--           claim (10 grippers, 6 technologies, 4 tuner inputs).
-- NOTE both need context_terms: numberSupported degrades a non-exact tolerance
-- with no context terms back to exact, deliberately, so it can never become
-- blanket support (claims.go:513).
--
-- ── TWO TRAPS, BOTH DELIBERATE CHOICES ────────────────────────────────────
-- 1. MERGE, never supersede-with-a-fresh-object. Recorded lesson from the
--    07-24 seed, which superseded migration 166's vonc row unread and silently
--    dropped nine banned_claims patterns another thread's checkers depended on.
--    Below, `data` is read and the facts key is ADDED to it.
-- 2. `writer_block_managed` is deliberately NOT set. composeWriterBlock
--    (refresh_evidence_base_action.go:566) rebuilds the block from each fact's
--    writer_line and emits NUMBERS / CAPABILITIES / NAMED ENTITIES only — it
--    has no NEVER-STATE section. Turning management on would silently delete
--    the "NOT TRACKED, NEVER STATE" lists, which are the half of these blocks
--    that stops the writer inventing a whole new category of figure. The blocks
--    stay hand-managed; the stale numbers in their prose are corrected here too.
--
-- Idempotent: facts are keyed by id and replaced wholesale, writer_block text
-- is set rather than patched, and re-running changes nothing.

\set ON_ERROR_STOP on

BEGIN;

DO $seed$
DECLARE
    r        record;
    existing jsonb;
    n        int;
BEGIN
    FOR r IN
        SELECT * FROM (VALUES

        -- ── robot-hands.com ────────────────────────────────────────────────
        ('00ff3af5-dad8-4770-9f70-3edc267a3c92'::uuid, $facts$[
          {"id":"rh-grippers","claim":"gripper models indexed","value":10,"kind":"count",
           "tolerance":"exact","context_terms":["gripper model","models indexed","grippers indexed"],
           "source":{"sql":"SELECT count(*) FROM products WHERE category='gripper' AND site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND status='active'"},
           "verified_at":"2026-07-26",
           "writer_line":"{value} gripper models indexed (products, category='gripper', status='active')"},
          {"id":"rh-manufacturers","claim":"manufacturers covered","value":6,"kind":"count",
           "tolerance":"exact","context_terms":["manufacturer"],
           "source":{"sql":"SELECT count(DISTINCT split_part(name,' ',1)) FROM products WHERE category='gripper' AND site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND status='active'"},
           "verified_at":"2026-07-26",
           "writer_line":"{value} manufacturers covered: Schunk, OnRobot, Robotiq, Zimmer Group, Festo, Schmalz"},
          {"id":"rh-actuation","claim":"actuation technologies represented","value":6,"kind":"count",
           "tolerance":"exact","context_terms":["actuation","technolog"],
           "source":{"attested_by":"robot-hands R7 catalogue expansion, 2026-07-22: electric parallel-jaw (5), pneumatic parallel-jaw, vacuum, magnetic, soft-robotic, adhesive — one datasheet-sourced model each"},
           "verified_at":"2026-07-26",
           "writer_line":"{value} actuation technologies represented in the index: electric parallel-jaw (5 models), and one each of pneumatic parallel-jaw, vacuum, magnetic, soft-robotic, adhesive"},
          {"id":"rh-parameters","claim":"parameters MatchMatrix tests","value":4,"kind":"count",
           "tolerance":"exact","context_terms":["parameter","matchmatrix","scored","tests"],
           "source":{"artifact":"MatchMatrix scoring implementation: gripping/holding force, jaw travel or grip range, rated payload, IP rating"},
           "verified_at":"2026-07-26",
           "writer_line":"{value} parameters MatchMatrix tests an application against: gripping/holding force, jaw travel or grip range, rated payload, IP rating"},
          {"id":"rh-spec-figures","claim":"published specification figures","value":59,"kind":"count",
           "tolerance":"gte","context_terms":["specification figure","spec figure","published figure"],
           "source":{"sql":"SELECT sum((SELECT count(*) FROM jsonb_each_text(specifications))) FROM products WHERE category='gripper' AND site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92' AND status='active'"},
           "verified_at":"2026-07-26",
           "writer_line":"{value} published specification figures held across the index"}
        ]$facts$::jsonb,
        -- writer_block: only the stale 39 changes.
        $rh$NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine — recount from the products table before changing one):
- 10 gripper models indexed (products, category='gripper', status='active'; as of 2026-07-26)
- 6 manufacturers covered: Schunk, OnRobot, Robotiq, Zimmer Group, Festo, Schmalz
- 6 actuation technologies represented in the index: electric parallel-jaw (5 models), and one each of pneumatic parallel-jaw, vacuum, magnetic, soft-robotic, adhesive
- 4 parameters MatchMatrix tests an application against: gripping/holding force, jaw travel or grip range, rated payload, IP rating
- 59 published specification figures held across the index (as of 2026-07-26; was 39 on 07-24 — recount, do not carry it forward)
NOT TRACKED, NEVER STATE: MatchMatrix query/usage counts, visitor numbers, customer counts, satisfaction rates. Nothing on this site measures them at any value.$rh$),

        -- ── gamesdesign.co.uk ──────────────────────────────────────────────
        ('e33263f4-74f8-494f-b191-546845dbbddf'::uuid, $facts$[
          {"id":"gd-tools","claim":"interactive design tools live","value":11,"kind":"count",
           "tolerance":"exact","context_terms":["tool"],
           "source":{"attested_by":"gamesdesign 043 wave-1 audit 2026-07-24: 11 client-side tools live, counted from the pages register"},
           "verified_at":"2026-07-24",
           "writer_line":"{value} interactive design tools live, all client-side and free"},
          -- CORRECTED 2026-07-31 (bugs_open/161). This fact was FALSE AS SEEDED, and
          -- because the register is both the writer whitelist and the gate authority it
          -- caused the claim and then vouched for it — 10 live components asserted it
          -- and cmd/claimscan reported 0 findings, correctly.
          --
          -- It read: claim "Monte Carlo trials per query", context_terms
          -- ["trial","monte carlo","simulation"], source.artifact "the figure is
          -- hard-coded in the shipped drop-rate tool JavaScript", writer_line
          -- "{value} Monte Carlo trials per query in the drop-rate tools".
          --
          -- NEITHER drop-rate tool performs Monte Carlo simulation, and neither
          -- contains ANY randomness: Math.random count is 0 in both. The tuner is
          -- closed-form Math.pow(1-p,k) plus a CDF array ("Cumulative distribution
          -- modelled via geometric distribution with optional hard pity cap", its own
          -- doc comment); the simulator computes exact binomial probability. The only
          -- real 10000 is `return Math.min(val, 10000)`, an INPUT CLAMP on attempts.
          -- The simulator component was last written 2026-06-05 — seven weeks before
          -- this seed — so it was false on arrival, not stale since.
          --
          -- "trial" is deliberately NOT kept in context_terms: it would leave
          -- "10,000 Monte Carlo trials" inside a matching window, so the engine would
          -- go on treating the false sentence as supported, which is the whole defect.
          {"id":"gd-trials","claim":"maximum attempts modelled per query","value":10000,"kind":"metric",
           "tolerance":"exact","context_terms":["attempt","modelled","max"],
           "source":{"artifact":"return Math.min(val, 10000) — the input clamp on attempts in tool-drop-rate-simulator; the tuner's CDF is built to the same bound. NOT a trial count: neither tool samples (Math.random count 0 in both, 2026-07-31)"},
           "verified_at":"2026-07-31",
           "writer_line":"{value} maximum attempts modelled per query in the drop-rate tools, using exact probability rather than sampling"},
          {"id":"gd-tuner-inputs","claim":"configurable inputs in the drop-rate tuner","value":4,"kind":"count",
           "tolerance":"exact","context_terms":["input","parameter","tuner"],
           "source":{"artifact":"drop-rate tuner UI: drop chance, kills per hour, pity timer, target hours"},
           "verified_at":"2026-07-24",
           "writer_line":"{value} configurable inputs in the drop-rate tuner: drop chance, kills per hour, pity timer, target hours"},
          {"id":"gd-articles","claim":"guides and articles live","value":10,"kind":"count",
           "tolerance":"gte","context_terms":["guide","article","post"],
           "source":{"attested_by":"gamesdesign 043 wave-1 audit 2026-07-24: 5 blog posts + 5 guides"},
           "verified_at":"2026-07-24",
           "writer_line":"{value} guides and articles live"}
        ]$facts$::jsonb,
        NULL),

        -- ── ai-agent-orchestration.com ─────────────────────────────────────
        ('2a8ebf9c-20a2-4c39-b191-840b012371da'::uuid, $facts$[
          {"id":"aao-agent-definitions","claim":"active agent definitions in the registry","value":175,"kind":"count",
           "tolerance":"gte","context_terms":["agent definition","specialised ai agent","agents in the registry","ai agents"],
           "source":{"sql":"SELECT count(*) FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL"},
           "verified_at":"2026-07-26",
           "writer_line":"more than 150 active agent definitions in the production registry ({value} as of 2026-07-26)"},
          {"id":"aao-agent-types","claim":"distinct agent types","value":174,"kind":"count",
           "tolerance":"gte","context_terms":["agent type"],
           "source":{"sql":"SELECT count(DISTINCT type) FROM agent_definitions WHERE COALESCE(is_snapshot,false)=false AND deleted_at IS NULL"},
           "verified_at":"2026-07-26",
           "writer_line":"more than 150 distinct agent types ({value} as of 2026-07-26)"},
          {"id":"aao-departments","claim":"departments in the platform's own taxonomy","value":8,"kind":"count",
           "tolerance":"exact","context_terms":["department"],
           "source":{"attested_by":"the platform's own organisational taxonomy: Strategy, Research, Content, Design, Development, Quality, Operations, Data. A SELF-description — never 'departments served'"},
           "verified_at":"2026-07-26",
           "writer_line":"{value} departments — Strategy, Research, Content, Design, Development, Quality, Operations, Data (the platform's OWN taxonomy, never 'departments served')"},
          {"id":"aao-live-sites","claim":"live sites in production","value":14,"kind":"count",
           "tolerance":"gte","context_terms":["live site","sites in production","websites"],
           "source":{"sql":"SELECT count(*) FROM sites s WHERE EXISTS (SELECT 1 FROM pages p WHERE p.site_id=s.id AND p.build_status='deployed')"},
           "verified_at":"2026-07-26",
           "writer_line":"{value} live sites in production, built and operated end-to-end by the platform"},
          {"id":"aao-services","claim":"backend services","value":17,"kind":"count",
           "tolerance":"exact","context_terms":["backend service","microservice","services"],
           "source":{"attested_by":"the service manifests under deployments/kustomize/services (17 backend services; the directory also holds frontend and infra overlays, so a bare directory count is NOT this figure)"},
           "verified_at":"2026-07-24",
           "writer_line":"{value} backend services"},
          {"id":"aao-work-items","claim":"automated work items completed","value":1051,"kind":"count",
           "tolerance":"gte","context_terms":["work item"],
           "source":{"sql":"SELECT count(*) FROM site_work_items WHERE status IN ('complete','verified')"},
           "verified_at":"2026-07-26",
           "writer_line":"{value} automated work items completed (the platform's work-item ledger — NOT cumulative, the ledger is reaped, so recount before restating)"},
          {"id":"aao-orchestrations","claim":"orchestrations per day","value":1834,"kind":"metric",
           "tolerance":"gte","context_terms":["orchestration"],
           "source":{"sql":"SELECT count(*) FROM orchestration_states WHERE created_at > now() - interval '24 hours'"},
           "verified_at":"2026-07-26",
           "writer_line":"over a thousand orchestrations a day ({value} in the 24 hours to 2026-07-26)"}
        ]$facts$::jsonb,
        $ao$NUMBERS (state only these, with their listed meaning; dated snapshots up to a listed live count are fine — these are live registry counts taken 2026-07-26, recount before restating):
- more than 150 active agent definitions in the production registry (175 as of 2026-07-26); "over 70 specialised AI agents" is also true and may stand where already written
- 8 departments — Strategy, Research, Content, Design, Development, Quality, Operations, Data. This is the platform's OWN organisational taxonomy (a self-description); never frame it as departments of external clients ("departments served")
- more than 150 distinct agent types (174 as of 2026-07-26); "30+ agent types" is also true and may stand
- 14 live sites in production, built and operated end-to-end by the platform
- 17 backend services (the service manifests under deployments/kustomize/services)
- 1,051 automated work items completed (the platform's work-item ledger). CAUTION: this figure is NOT cumulative — the ledger is reaped, and it fell from 1,267 on 07-24 to 1,051 on 07-26. Never present it as a growing total; recount before stating it.
- over a thousand orchestrations a day (1,834 in the 24 hours to 2026-07-26)
- Architecture: Kubernetes, Kafka, Postgres — true and stated freely
NOT TRACKED / DOES NOT EXIST, NEVER STATE: clients served, "departments served", satisfaction rates, awards won, concurrent-instance counts ("thousands of concurrent instances" is not measured), uptime percentages. None of these are measured; every such figure at any value is an invention.$ao$)

        ) AS t(site_id, facts, writer_block)
    LOOP
        -- Read before writing. MERGE onto whatever the row already holds, so a
        -- banned_claims list or an allowed_entities list put there by another
        -- thread survives (the 2026-07-24 lesson).
        SELECT data INTO existing FROM site_specs
        WHERE site_id = r.site_id AND aspect = 'evidence_base' AND is_current;

        IF existing IS NULL THEN
            RAISE EXCEPTION '218: no current evidence_base row for site % — expected one to merge onto', r.site_id;
        END IF;

        existing := existing || jsonb_build_object('facts', r.facts);
        IF r.writer_block IS NOT NULL THEN
            existing := existing || jsonb_build_object('writer_block', r.writer_block);
        END IF;
        -- Never turn on managed regeneration: composeWriterBlock has no
        -- NEVER-STATE section and would silently drop these blocks' bans.
        existing := existing - 'writer_block_managed';

        UPDATE site_specs SET data = existing, updated_at = now()
        WHERE site_id = r.site_id AND aspect = 'evidence_base' AND is_current;

        GET DIAGNOSTICS n = ROW_COUNT;
        IF n <> 1 THEN
            RAISE EXCEPTION '218: expected 1 evidence_base row for site %, updated %', r.site_id, n;
        END IF;
    END LOOP;
END $seed$;

-- ── Post-conditions ────────────────────────────────────────────────────────
DO $post$
DECLARE
    n int;
BEGIN
    -- Every targeted site now parses as opted-in (facts[] non-empty).
    SELECT count(*) INTO n
    FROM site_specs ss
    WHERE ss.aspect='evidence_base' AND ss.is_current
      AND ss.site_id IN ('00ff3af5-dad8-4770-9f70-3edc267a3c92',
                         'e33263f4-74f8-494f-b191-546845dbbddf',
                         '2a8ebf9c-20a2-4c39-b191-840b012371da')
      AND jsonb_array_length(COALESCE(ss.data->'facts','[]'::jsonb)) > 0;
    IF n <> 3 THEN
        RAISE EXCEPTION '218: expected 3 sites with a non-empty facts[], got %', n;
    END IF;

    -- The writer_block survived on all three (it is what stops the writer
    -- inventing; dropping it would be a silent regression).
    SELECT count(*) INTO n
    FROM site_specs ss
    WHERE ss.aspect='evidence_base' AND ss.is_current
      AND ss.site_id IN ('00ff3af5-dad8-4770-9f70-3edc267a3c92',
                         'e33263f4-74f8-494f-b191-546845dbbddf',
                         '2a8ebf9c-20a2-4c39-b191-840b012371da')
      AND length(COALESCE(ss.data->>'writer_block','')) > 200;
    IF n <> 3 THEN
        RAISE EXCEPTION '218: a writer_block was lost — % of 3 survive', n;
    END IF;

    -- Every fact carries a source and a writer_line, or the register is not
    -- machine-usable and composeWriterBlock would omit it.
    SELECT count(*) INTO n
    FROM site_specs ss, LATERAL jsonb_array_elements(ss.data->'facts') f
    WHERE ss.aspect='evidence_base' AND ss.is_current
      AND ss.site_id IN ('00ff3af5-dad8-4770-9f70-3edc267a3c92',
                         'e33263f4-74f8-494f-b191-546845dbbddf',
                         '2a8ebf9c-20a2-4c39-b191-840b012371da')
      AND (NOT (f ? 'source') OR NOT (f ? 'writer_line') OR NOT (f ? 'context_terms'));
    IF n <> 0 THEN
        RAISE EXCEPTION '218: % fact(s) lack a source, writer_line or context_terms', n;
    END IF;

    RAISE NOTICE '218: three registers seeded — the claims checkers are no longer a no-op on the sites that fabricated';
END $post$;

INSERT INTO schema_migrations (filename, notes)
VALUES ('218_evidence_facts_for_043_sites.sql',
        'bugs_open/043: seed machine-readable facts[] for robot-hands, gamesdesign and ai-agent-orchestration. Their evidence_base rows were writer_block-only, so ParseEvidenceBase returned nil and BOTH claims checkers were silent no-ops on the three sites that had actually fabricated. Every figure re-derived live 2026-07-26 first: most had moved, and work-items-completed had gone DOWN (1,267 -> 1,051, the ledger is reaped). Facts carry source.sql so refresh_evidence_base keeps them current. writer_block_managed deliberately left off — composeWriterBlock has no NEVER-STATE section.')
ON CONFLICT (filename) DO NOTHING;

COMMIT;
