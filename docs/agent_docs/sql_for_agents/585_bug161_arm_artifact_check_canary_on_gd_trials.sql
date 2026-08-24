-- 585_bug161_arm_artifact_check_canary_on_gd_trials.sql
--
-- bugs_open/161 close-out + RFC_025 §5.3 step: attach the first real
-- `artifact_check` to the fact that motivated the whole mechanism.
--
-- CONTEXT. RFC_025 (RATIFIED 2026-08-12, council APPROVED round 2, corr
-- 9fd94852-ff79-496b-96b5-78a8d3619162) shipped `source.artifact_check`: an
-- opt-in key on an `artifact`-sourced evidence fact that the daily
-- evidence-freshness sweep re-proves against the named stored artefact
-- (refreshArtifactCheckFact, refresh_evidence_base_action.go). The code has
-- been live on the fleet since chassis v1.0.1295 (2026-08-13). As of
-- 2026-08-24, ZERO facts fleet-wide carry the key — the mechanism has never
-- been exercised on real data. RFC_025's own staged plan (§5.3) names the
-- canary: "retype the `gd-trials` fact itself ... a fact that would have
-- caught the ORIGINAL false claim, proving the mechanism on the motivating
-- case rather than a synthetic one." This file is that step, and it is the
-- last unmet condition for RFC_025's IMPLEMENTED status.
--
-- THE FACT. gamesdesign.co.uk's `gd-trials` ("maximum attempts modelled per
-- query" = 10000) was corrected 2026-07-31 (migration 270) after being
-- registered false ("Monte Carlo trials per query") — the register both
-- instructed the writer to state the falsehood and vouched for it against
-- every claims gate (bugs_open/161). The corrected fact cites
-- `Math.min(val, 10000)`, the input clamp in tool-drop-rate-simulator's
-- shipped JS. Today that citation is still free text nothing re-checks.
--
-- WHAT THIS FILE DOES. Adds to the fact's source:
--
--   "artifact_check": {
--     "component_id":   "15f1f798-51fb-41d0-8a07-18148b39a293",
--     "pattern":        "Math\.min\(val,\s*10000\)",
--     "must_be_present": true
--   }
--
-- The component is tool-drop-rate-simulator's hero component — the row that
-- actually contains the clamp, untouched since 2026-06-05 20:12:52 (i.e. the
-- same bytes the false fact was registered against, and the same bytes the
-- 2026-07-31 diagnosis read). Measured 2026-08-24 before writing this file:
-- rendered_html ~ 'Math\.min\(val,\s*10000\)' → true; 'Math\.random' → false
-- (the bug's decisive negative still holds).
--
-- WHAT HAPPENS NEXT (the observables, so verification is a query, not a hope):
--   * The daily `evidence-freshness` scheduled task (86400s, last fired
--     2026-08-24 09:05Z) sweeps EVERY site with a current evidence_base —
--     resolveEvidenceSites has no sql-facts filter — so gamesdesign is
--     covered even though its register has zero sql-sourced facts.
--   * On a PASS the sweep bumps the fact's `verified_at` to the sweep date.
--     This file deliberately leaves `verified_at` at '2026-07-31' so that
--     bump is a demand-control observable: verified_at still '2026-07-31'
--     after the next sweep = the check did NOT run; bumped = it ran and
--     passed.
--   * On a DRIFT (someone rewrites the tool and the clamp changes) the sweep
--     raises a `stale_evidence` item — shouldRaiseStaleEvidence's
--     ArtifactCheckDrifted disjunct gates this independently of `changed`.
--   * The fail direction is proven at unit level
--     (TestArtifactCheck_MismatchedPatternFlagsDrift is exactly this fact's
--     shape) — this file does NOT induce a live fault, because a deliberate
--     drift would raise a needs_human_review item for the owner to triage.
--
-- CHURN NOTE. A passing check sets changed=true, so gamesdesign's register
-- will now gain one superseded row per day, like every sql-sourced register
-- already does (measured 2026-08-24: 11-14 rows/10d on the sql-fact sites).
-- Not new behaviour, just this register joining it.
--
-- ORDERING. None. The reading code has been live since 2026-08-13; data is
-- live the moment this applies. Safe to apply at any time; the rollback is
-- to supersede-and-reinsert without the key (no binary depends on it).
--
-- Supersede-then-insert rather than UPDATE, following migration 270 and
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
    FROM site_specs
    WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
      AND aspect = 'evidence_base' AND is_current
      AND data->'facts' @> '[{"id":"gd-trials","source":{"artifact_check":{}}}]'::jsonb;
    IF n <> 0 THEN
        RAISE EXCEPTION 'bug161/RFC_025 canary: gd-trials already carries artifact_check — this file is already applied.';
    END IF;
END $$;

-- Guard 2: the register must be in the state this file was written against.
DO $$
DECLARE n int; nfacts int;
BEGIN
    SELECT count(*) INTO n
    FROM site_specs
    WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
      AND aspect = 'evidence_base' AND is_current
      AND data->'facts' @> '[{"id":"gd-trials","claim":"maximum attempts modelled per query"}]'::jsonb;
    IF n <> 1 THEN
        RAISE EXCEPTION 'bug161/RFC_025 canary: expected exactly 1 current evidence_base row carrying the CORRECTED gd-trials claim, found %. The register has moved — re-read before applying.', n;
    END IF;

    SELECT jsonb_array_length(data->'facts') INTO nfacts
    FROM site_specs
    WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
      AND aspect = 'evidence_base' AND is_current;
    IF nfacts <> 4 THEN
        RAISE EXCEPTION 'bug161/RFC_025 canary: expected 4 facts in the register, found %. Re-read before applying.', nfacts;
    END IF;
END $$;

-- Guard 3: the canary must be born passing — the named component must exist
-- on THIS site (the same join refreshArtifactCheckFact uses) and must match
-- the pattern today. If this fails, the artefact has changed since
-- 2026-08-24 and the pattern needs re-deriving, not forcing.
DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n
    FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
    WHERE pc.id = '15f1f798-51fb-41d0-8a07-18148b39a293'
      AND p.site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
      AND pc.rendered_html ~ 'Math\.min\(val,\s*10000\)';
    IF n <> 1 THEN
        RAISE EXCEPTION 'bug161/RFC_025 canary: component 15f1f798 does not exist on gamesdesign or no longer matches Math\.min\(val,\s*10000\) — the artefact has changed; re-derive the pattern before applying.';
    END IF;
END $$;

-- 1. supersede the current row (kept as history)
UPDATE site_specs
SET is_current = false, superseded_at = now()
WHERE site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
  AND aspect = 'evidence_base' AND is_current;

-- 2. insert the register with artifact_check added to gd-trials' source —
--    everything else byte-identical (writer_block, the other three facts,
--    any unknown keys all carried verbatim).
INSERT INTO site_specs (site_id, aspect, data, source, source_agent, created_by, pinned, notes)
SELECT
    s.site_id,
    'evidence_base',
    jsonb_set(
        s.data,
        '{facts}',
        (
            SELECT jsonb_agg(
                CASE WHEN f->>'id' = 'gd-trials' THEN
                    jsonb_set(
                        f,
                        '{source,artifact_check}',
                        jsonb_build_object(
                            'component_id',    '15f1f798-51fb-41d0-8a07-18148b39a293',
                            'pattern',         'Math\.min\(val,\s*10000\)',
                            'must_be_present', true
                        )
                    )
                ELSE f END
                ORDER BY ord
            )
            FROM jsonb_array_elements(s.data->'facts') WITH ORDINALITY AS t(f, ord)
        )
    ),
    'manual',
    NULL,
    'session-2026-08-24-bug161-artifact-check-canary',
    s.pinned,
    'bugs_open/161 close-out / RFC_025 §5.3: first real artifact_check fleet-wide, attached to the fact that motivated the mechanism. gd-trials now re-proves its cited Math.min(val, 10000) clamp against tool-drop-rate-simulator''s hero component (15f1f798) on every daily evidence-freshness sweep. A check of this shape, present on 2026-07-24, would have refused the original "Monte Carlo trials" registration the day it was written. verified_at deliberately left at 2026-07-31 so the first sweep''s bump proves the check ran.'
FROM site_specs s
WHERE s.site_id = 'e33263f4-74f8-494f-b191-546845dbbddf'
  AND s.aspect = 'evidence_base'
  AND s.is_current = false
  AND s.superseded_at IS NOT NULL
ORDER BY s.superseded_at DESC
LIMIT 1;

-- 3. assert the outcome before committing (DO/RAISE — a SELECT-only verify
--    block cannot stop the COMMIT).
DO $$
DECLARE
    n_current int; n_facts int; pat text; comp text; mbp text; vat text;
    wb_old text; wb_new text;
BEGIN
    SELECT count(*) INTO n_current FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND aspect='evidence_base' AND is_current;
    IF n_current <> 1 THEN
        RAISE EXCEPTION 'bug161/RFC_025 canary: expected exactly 1 current row after the write, found %', n_current;
    END IF;

    SELECT jsonb_array_length(data->'facts'),
           (SELECT f->'source'->'artifact_check'->>'pattern'
              FROM jsonb_array_elements(data->'facts') f WHERE f->>'id'='gd-trials'),
           (SELECT f->'source'->'artifact_check'->>'component_id'
              FROM jsonb_array_elements(data->'facts') f WHERE f->>'id'='gd-trials'),
           (SELECT f->'source'->'artifact_check'->>'must_be_present'
              FROM jsonb_array_elements(data->'facts') f WHERE f->>'id'='gd-trials'),
           (SELECT f->>'verified_at'
              FROM jsonb_array_elements(data->'facts') f WHERE f->>'id'='gd-trials')
    INTO n_facts, pat, comp, mbp, vat
    FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND aspect='evidence_base' AND is_current;

    IF n_facts <> 4 THEN
        RAISE EXCEPTION 'bug161/RFC_025 canary: fact count changed to % — this file adds a key to one fact, never adds or drops a fact', n_facts;
    END IF;
    IF pat IS DISTINCT FROM 'Math\.min\(val,\s*10000\)' THEN
        RAISE EXCEPTION 'bug161/RFC_025 canary: pattern did not land as written, got %', coalesce(pat, '(null)');
    END IF;
    IF comp IS DISTINCT FROM '15f1f798-51fb-41d0-8a07-18148b39a293' THEN
        RAISE EXCEPTION 'bug161/RFC_025 canary: component_id did not land, got %', coalesce(comp, '(null)');
    END IF;
    IF mbp IS DISTINCT FROM 'true' THEN
        RAISE EXCEPTION 'bug161/RFC_025 canary: must_be_present did not land as true, got %', coalesce(mbp, '(null)');
    END IF;
    IF vat IS DISTINCT FROM '2026-07-31' THEN
        RAISE EXCEPTION 'bug161/RFC_025 canary: verified_at must remain 2026-07-31 (the sweep''s bump is the proof the check ran), got %', coalesce(vat, '(null)');
    END IF;

    -- the writer_block must be byte-identical to the superseded row's
    SELECT data->>'writer_block' INTO wb_new FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND aspect='evidence_base' AND is_current;
    SELECT data->>'writer_block' INTO wb_old FROM site_specs
    WHERE site_id='e33263f4-74f8-494f-b191-546845dbbddf' AND aspect='evidence_base'
      AND is_current = false AND superseded_at IS NOT NULL
    ORDER BY superseded_at DESC LIMIT 1;
    IF wb_new IS DISTINCT FROM wb_old THEN
        RAISE EXCEPTION 'bug161/RFC_025 canary: writer_block changed — this file must not touch it';
    END IF;
END $$;

COMMIT;
