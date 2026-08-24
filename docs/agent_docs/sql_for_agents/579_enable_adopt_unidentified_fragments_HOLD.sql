-- 579 — ARM `adopt_unidentified_fragments` (bugs_open/357 phase 2, RFC_046)
--
-- ⚠ _HOLD: applied BY HAND, on the owner's instruction of 2026-08-24 ("arm it").
-- Not for the runner. This is the CONFIG half of phase 2; the Go half is live
-- already (committed `e6afb4873`/`117c918c1`, council-APPROVED on trail
-- `74e4c1fd`, and probed in the running binary with both control arms). Applying
-- this makes the behaviour change real, immediately, with no image tag to roll
-- back — the rollback is 579_..._ROLLBACK.sql.
--
-- PREREQUISITE, and it is enforced below rather than trusted: migration 577 must
-- have seeded the `adopted-fragment` component first. Armed-but-unseeded is a
-- reachable state that degrades to `component_id` NULL — the honest unknown, but
-- also a new population the council's bug_historian seat objected to on
-- bugs_closed/039's precedent. Seed first, arm second.
--
-- ============================================================================
-- WHAT IT TURNS ON
-- ============================================================================
-- One key on every live `save_page_sections` step. It governs TWO coupled halves:
--
--  (a) a fallback-adopted fragment — a page whose HTML carries no <section> at
--      all, stored whole as one section, and declaring no data-component — is
--      bound to the `adopted-fragment` component AFTER a byte-identity round trip
--      proves that component reproduces its bytes, or left honestly unidentified.
--      Never bound to the positional slot name its page plan supplied, which is
--      how 22 live rows came to declare themselves the shared `hero` while storing
--      a whole interactive tool.
--
--  (b) Layer 2's carry arms let carried bytes keep their own component — narrowed
--      in council round 2 to `adopted-fragment` rows ONLY, so no legitimately
--      typed component's carry semantics change.
--
-- Neither half is useful alone: without (b) the next rebuild re-mints the plan's
-- identity over an adopted row, which is why they share one key.
--
-- ⚠ NO SLOT NAME CHANGES AND `pages.sections` IS NEVER TOUCHED. The carry-forward
-- landmine (Layer 2 matches stored to incoming rows on slot-name equality and
-- nothing else) is not managed by this change; it is never armed by it.
--
-- ============================================================================
-- THE STEPS — all six live ones, enumerated RECURSIVELY on 2026-08-24
-- ============================================================================
-- Three are nested inside `sub_workflow`s and a top-level `jsonb_each` MISSES
-- them (575's lesson, and the reason this file lists paths rather than agents):
--
--   page-build-handler       workflow,steps,save_sections                                              -> ON
--   pageflow-builder         workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections   -> ON
--   page-rebuild             workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections   -> ON
--   site-work-orchestrator   workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections   -> ON
--   tool-recreation-handler  workflow,steps,save_sections                                              -> ON
--   page-rerender            workflow,steps,save_sections                                              -> ON (carry half only; see below)
--
-- WHY ALL SIX AND NOT ONE CANARY, which is what this lane's own runbook said
-- before the surface was enumerated. The mint can only occur on the HTML-parsing
-- path, because that is where the no-<section> fallback lives — and the save falls
-- through to HTML parsing whenever the metadata path yields nothing. **FIVE of the
-- six carry an `html_field`** (measured 2026-08-24), so five can mint, not one.
-- Arming only `tool-recreation-handler` — the one that declares
-- `expects_no_sections_metadata` and is the most obvious producer — would have
-- left four minting, which is precisely the "one call site of a shared mechanism
-- gets the rigorous fix while the mechanism stays generic elsewhere" shape that
-- 575's own bug_historian objection names, and that gated this lane's phase 2
-- round 1. Measured instead of argued.
--
-- `page-rerender` has NO `html_field`, so it cannot reach the fallback and cannot
-- mint. It is armed for half (b) alone: without it an adopted row loses its
-- identity the first time a rerender carries its bytes. Named here so the next
-- reader does not have to re-derive that it was considered.
--
-- WHY THIS IS SAFE TO ARM WIDE, in one checkable sentence: adoption fires ONLY on
-- a section produced by the no-<section> fallback (`SectionData.FallbackAdopted`,
-- set nowhere else) that ALSO declares no `data-component`. An ordinary page has
-- <section> blocks, so the branch is unreachable for it — arming a whole-site
-- builder cannot change what a normal page stores.
--
-- ROLLBACK: 579_enable_adopt_unidentified_fragments_HOLD_ROLLBACK.sql — the exact
-- inverse (`#-` of the one key on each of the six paths).

BEGIN;

-- Top-level steps
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
        '{workflow,steps,save_sections,config,adopt_unidentified_fragments}', 'true'::jsonb, true)
 WHERE type IN ('page-build-handler', 'tool-recreation-handler', 'page-rerender')
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config->'workflow'->'steps'->'save_sections'->>'action' = 'save_page_sections';

-- Nested in build_pages_loop.sub_workflow
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections,config,adopt_unidentified_fragments}',
        'true'::jsonb, true)
 WHERE type IN ('pageflow-builder', 'page-rebuild')
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'save_sections'->>'action' = 'save_page_sections';

-- Nested in build_items_loop.sub_workflow
UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
        '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections,config,adopt_unidentified_fragments}',
        'true'::jsonb, true)
 WHERE type = 'site-work-orchestrator'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config->'workflow'->'steps'->'build_items_loop'->'config'->'sub_workflow'->'steps'->'save_sections'->>'action' = 'save_page_sections';

-- Refuse to commit unless the seed exists AND all six steps are armed. A verify
-- block of SELECTs cannot stop a COMMIT (ON_ERROR_STOP ignores a non-empty
-- result set), so this RAISEs — the RFC_006 trap.
DO $$
DECLARE armed int; seeded int;
BEGIN
    SELECT count(*) INTO seeded FROM content_components
     WHERE function='adopted-fragment' AND is_active AND btrim(html_template)='{{.body}}';
    IF seeded <> 1 THEN
        RAISE EXCEPTION 'migration 577 has not been applied (or the template is not the identity function): '
                        'found % seeded adopted-fragment component(s). Arming without the seed degrades every '
                        'adoption to an unidentified row.', seeded;
    END IF;

    SELECT count(*) INTO armed
      FROM agent_definitions ad,
           LATERAL jsonb_path_query(ad.default_config, 'strict $.**') AS step(value)
     WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
       AND jsonb_typeof(step.value)='object'
       AND step.value->>'action' = 'save_page_sections'
       AND (step.value->'config'->>'adopt_unidentified_fragments') = 'true';
    IF armed <> 6 THEN
        RAISE EXCEPTION 'expected 6 armed save_page_sections steps, found %. A path that did not match is a '
                        'step still minting — re-enumerate recursively before retrying.', armed;
    END IF;
END $$;

COMMIT;
