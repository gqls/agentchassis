-- FILE: docs/agent_docs/sql_for_agents/525_bugfix_238_contact_block_reads_identity_email.sql
--
-- bugs_open/238 §11.11 — point `contact-block`'s `contact_email` field at the
-- spelling the resolver can actually reach.
--
-- THE DEFECT, in one sentence: the component asks for `site_specs.contact.email`
-- and the platform stores that fact as `site_specs.identity.email` / the
-- `sites.email` column, and the resolver's bridge between the two is hard-gated
-- to ONE aspect name, so the value is refused even when it is sitting there.
--
-- The gate, `plan_sections_action.go` `resolveSpecAlias` step 2:
--
--     // 2. The canonical sites row.
--     if aspect != "identity" { return nil, false }
--     col, mapped := siteRowIdentityColumns[leaf]   // "email" -> "email"
--
-- So `site_specs.identity.email` resolves from `sites.email`, and
-- `site_specs.contact.email` — the same fact, different spelling — returns
-- not-found on every site, on every build, and always has.
--
-- MEASURED 2026-08-21, and the blast radius is exactly the damage:
--   * ONE active component declares `contact.email` (this one, `contact-block`,
--     id 4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4) and ONE declares `identity.email`.
--     The first resolves nowhere; the second works. Two spellings of one fact.
--   * `contact_email` on `contact-block`: **6 deployed rows across 3 sites, all 6
--     missing the key** — which is precisely 6 of the 9 unqueued (page, slot)
--     pairs the STRUCTURAL_KEY_CARRY_MISS findings name.
--   * `function='contact-block'` has exactly ONE active row, so there is no fork
--     to skip (RFC_034: convert by `content_components.id`, never by function —
--     a function-keyed change silently misses forks).
--
-- WHY THIS SPELLING AND NOT A RESOLVER CHANGE. Widening the alias so any aspect
-- may read the sites row is a shared-seam change to the resolver every build
-- flows through: bigger blast radius, RFC/council-shaped, and it would take the
-- publish-this-address question (below) fleet-wide in one step. Repointing one
-- field is one row, no Go, no roll, reversible, and it aligns this component
-- with the sibling that already works. Reuse the working spelling; do not widen
-- the mechanism to accommodate the broken one.
--
-- ⚠ OWNER DECISION THIS RESTS ON, recorded because the code cannot justify it.
-- Every address this will now publish is `<site>@contactforsales.com`, a pattern
-- on 15 of 44 live sites — systematic enough to look platform-assigned rather
-- than a client's own inbox. Publishing a reachable address merely BECAUSE it is
-- reachable is `bugs_open/140`'s defect exactly (a contact component served an
-- invented `info@example.com` on eight live sites). So it was put to the owner
-- rather than assumed, and the ruling on 2026-08-21 was:
--     "yes, that email should appear on contact pages"
-- Without that ruling this file must not be applied.
--
-- WHAT IT WILL AND WILL NOT FIX, stated per site so a partial result is not read
-- as a partial failure. Measured against `is_current = true` rows only, because
-- `ensureSpecs` loads `WHERE is_current = true` and takes the LAST row per aspect
-- with no ORDER BY — checked here: exactly ONE current `identity` row per site,
-- so resolution is deterministic (the 29 / 11 / 3 totals are historical versions).
--   * leopardessconsulting.co.uk — current `identity` row HAS `email`, so it
--     resolves at the LITERAL path. 4 pairs fixed.
--   * robot-hands.com — no `identity.email` leaf, but `sites.email` is populated,
--     so it resolves via alias step 2. 1 pair fixed.
--   * gamesdesign.co.uk — no email in the identity aspect AND an empty
--     `sites.email`. It will STILL not resolve, and that is the correct outcome:
--     the site genuinely has no contact address. 1 pair stays unresolved and
--     becomes an honest data gap for a human, not a resolver bug.
-- Expected: 5 of 6 rows resolve on their next section resolve; 1 does not.
--
-- ⚠ CONFIG IS LIVE ON APPLY, BUT THE PAGES ARE NOT. This changes what the NEXT
-- resolve computes; it does not backfill `page_components.content_data`. The six
-- pages keep serving what they serve until each is re-rendered
-- (`page_rerender` with `spec.reason='section_data_resolved'` — the no-LLM,
-- MERGING path). Applying this alone is safe and inert at the artefact.
--
-- NOT A CARRY REGRESSION. The stored rows do not hold `contact_email` at all
-- (that is the damage), so there is no stored value for the PBP-039 carry to
-- prefer over the newly-resolving source. Live resolution wins anyway — the
-- carry runs only after the literal path and every alias have missed.
--
-- Rollback: 525_bugfix_238_contact_block_reads_identity_email_ROLLBACK.sql
-- (sets the source back to `site_specs.contact.email`).

\set ON_ERROR_STOP on

BEGIN;

\echo '=== BEFORE ==='
SELECT id, name, jsonb_pretty(input_schema->'fields'->'contact_email') AS contact_email_field
  FROM content_components
 WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';

-- Pre-flight: assert the row, the field and the CURRENT source are what this
-- file was written against. A jsonb_set on a moved shape would create a branch
-- and report success — arming nothing while every reader says it is done.
DO $$
DECLARE
    v_rows int;
    v_src  text;
BEGIN
    SELECT count(*) INTO v_rows
      FROM content_components
     WHERE function = 'contact-block' AND COALESCE(is_active, true);
    IF v_rows <> 1 THEN
        RAISE EXCEPTION '238/525: expected exactly 1 active contact-block row, found % — a fork appeared; convert by content_components.id per RFC_034 and re-derive this file', v_rows;
    END IF;

    SELECT input_schema->'fields'->'contact_email'->>'source' INTO v_src
      FROM content_components
     WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';

    IF v_src IS DISTINCT FROM 'site_specs.contact.email' THEN
        RAISE EXCEPTION '238/525: contact_email.source is % (want site_specs.contact.email) — already applied, or the schema moved; re-read before writing', COALESCE(v_src, '(absent)');
    END IF;
END $$;

UPDATE content_components
   SET input_schema = jsonb_set(
           input_schema,
           '{fields,contact_email,source}',
           '"site_specs.identity.email"'::jsonb,
           false),                      -- create_missing = FALSE: the key must
                                        -- already exist, and the pre-flight
                                        -- proved it does. A missing key here is
                                        -- a moved schema, not something to add.
       updated_at = now()
 WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';

\echo '=== AFTER ==='
SELECT id, name, jsonb_pretty(input_schema->'fields'->'contact_email') AS contact_email_field
  FROM content_components
 WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';

-- VERIFY: the source moved, the rest of the field is untouched, and NOTHING
-- else on the estate still declares the unreachable spelling.
DO $$
DECLARE
    v_src      text;
    v_type     text;
    v_required boolean;
    v_stragglers int;
BEGIN
    SELECT input_schema->'fields'->'contact_email'->>'source',
           input_schema->'fields'->'contact_email'->>'type',
           (input_schema->'fields'->'contact_email'->>'required')::boolean
      INTO v_src, v_type, v_required
      FROM content_components
     WHERE id = '4ebdccb5-6eb0-4b03-9fe1-0fa5cc315fb4';

    IF v_src IS DISTINCT FROM 'site_specs.identity.email' THEN
        RAISE EXCEPTION '238/525 verify FAILED: source is % (want site_specs.identity.email)', COALESCE(v_src, '(absent)');
    END IF;
    -- The neighbours are asserted, not assumed: jsonb_set replaces a subtree, and
    -- a path typo would silently drop type/required while the source read fine.
    IF v_type IS DISTINCT FROM 'text' OR v_required IS NOT TRUE THEN
        RAISE EXCEPTION '238/525 verify FAILED: the field lost its neighbours (type=%, required=%) — jsonb_set hit the wrong path', COALESCE(v_type,'(absent)'), COALESCE(v_required::text,'(absent)');
    END IF;

    SELECT count(*) INTO v_stragglers
      FROM content_components
     WHERE COALESCE(is_active, true)
       AND input_schema::text LIKE '%site_specs.contact.%';
    IF v_stragglers <> 0 THEN
        RAISE NOTICE '238/525: % active component(s) still declare a site_specs.contact.* source — the aspect has never existed on any site, so those fields resolve nowhere too. Not this file''s scope; census them before assuming they are fine.', v_stragglers;
    END IF;

    RAISE NOTICE '238/525: contact-block.contact_email now reads site_specs.identity.email (type and required intact)';
END $$;

COMMIT;

-- ---------------------------------------------------------------------------
-- AFTER APPLYING — resolution is now possible, but no page has changed yet.
--
-- 1. Confirm what each affected site will resolve to (expect 2 of 3 to answer):
--
--   SELECT s.domain,
--          COALESCE((SELECT ss.data #>> '{email}' FROM site_specs ss
--                     WHERE ss.site_id = s.id AND ss.aspect = 'identity'
--                       AND ss.is_current = true LIMIT 1),
--                   NULLIF(btrim(s.email), ''),
--                   '(still nothing — honest data gap)') AS will_resolve_to
--     FROM sites s
--    WHERE s.domain IN ('leopardessconsulting.co.uk','robot-hands.com','gamesdesign.co.uk');
--
-- 2. Then re-render the six pages to carry it to the artefact — `page_rerender`
--    with `spec.reason='section_data_resolved'` (no LLM; the MERGING path, which
--    structurally cannot lose a key). Check each page's divergence stamp first:
--    a rebuild silently discards hand-patched rendered_html (bugs_open/229).
--
-- 3. Verify at the SERVED page, never at the row:
--      curl -s https://<domain>/<page>.html | grep -c '<the mailto or address>'
--    A repaired row is not a repaired page.
