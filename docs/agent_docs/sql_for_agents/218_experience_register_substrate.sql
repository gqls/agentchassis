-- 218_experience_register_substrate.sql
--
-- The experience register's substrate: a library of small reusable user
-- experiences (component contracts and micro-journeys), the invariants they are
-- bound by, and the per-site forks that bind a base entry to real pages and
-- selectors. Design + owner rulings:
-- docs/agent_docs/docs024_key_docs_latest/experience_register/
--   PLAN_2026-07-24_experience_register.md (§2 rulings, §3 design)
--   harvest/HARVEST_01_2026-07-26_vonc_provocations.md   (§3 — 10 corrections)
--   harvest/HARVEST_02_2026-07-26_brochure_components.md (§2 — invariants)
--
-- WHY the shape is what it is: every column below the taxonomy axes exists
-- because a LIVE implementation needed it and the 2026-07-24 design could not
-- express it. `states` because the most valuable clause in the harvested code is
-- "this must NOT be a control"; `degraded_states` because a failure path is
-- where dishonesty enters a page; `automatic_triggers` because four of the nine
-- harvested behaviours are triggered by the viewport or a system preference
-- rather than by a visitor; `requires_invariant` because ONE rule turned up in
-- six independent implementations across two sites and five authors, and
-- copying it into each entry would industrialise the drift instead of ending it.
--
-- ORDERING — this migration must NOT be applied before the image that carries
-- the Go half (validDocSubjectTypes gaining 'experience-pattern'). Applied
-- early, the widened CHECK recreates exactly the split contract of migration
-- 184 (bugs_closed/064): the DB would accept a subject type every doc action
-- still rejects. `TestValidDocSubjectTypes_LockstepWithMigrationCheck` fails the
-- build if this file and the Go vocabulary disagree, so the pair must travel
-- together — but the ORDER is still on the operator: image first, then this.
--
-- Idempotent: constraint re-adds are guarded; CREATE TABLE IF NOT EXISTS; the
-- invariant seeds use ON CONFLICT DO NOTHING.

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. Travelling docs for register entries: subject_type 'experience-pattern'.
--    Each entry carries a doc_plans PLAN (direction, provenance, decisions) and
--    doc_notes (council trail) keyed by the pattern name — the content_components
--    + doc_plans 'tool' precedent, which is what the owner ruled on 2026-07-24.
-- ---------------------------------------------------------------------------
ALTER TABLE doc_plans DROP CONSTRAINT IF EXISTS doc_plans_subject_type_check;
ALTER TABLE doc_plans ADD CONSTRAINT doc_plans_subject_type_check
  CHECK (subject_type = ANY (ARRAY['tool'::text,'pipeline'::text,'experience'::text,'action'::text,'experience-pattern'::text]));

ALTER TABLE doc_notes DROP CONSTRAINT IF EXISTS doc_notes_subject_type_check;
ALTER TABLE doc_notes ADD CONSTRAINT doc_notes_subject_type_check
  CHECK (subject_type = ANY (ARRAY['tool'::text,'pipeline'::text,'experience'::text,'action'::text,'experience-pattern'::text]));

-- ---------------------------------------------------------------------------
-- 2. experience_invariants — clauses that hold across many patterns, held ONCE.
--    An invariant earns its row by having been implemented independently more
--    than once; `sightings` carries that evidence, because the count IS the
--    argument.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS experience_invariants (
    name        text PRIMARY KEY,
    clause      text NOT NULL,
    rationale   text NOT NULL,
    sightings   jsonb NOT NULL DEFAULT '[]'::jsonb,
    status      text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','approved')),
    created_by  text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------------
-- 3. experience_patterns — the register itself. Site-agnostic by construction:
--    no URLs, no site-specific selectors (bugs_closed/045 lesson — a static
--    fallback that re-applies on every render cannot be overridden; all
--    concrete values live in the per-site fork, which is data).
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS experience_patterns (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name          text NOT NULL UNIQUE,
    kind          text NOT NULL CHECK (kind IN ('component-contract','micro-journey')),
    display_name  text NOT NULL,
    aka           jsonb NOT NULL DEFAULT '[]'::jsonb,
    description   text,

    -- Selection axes (the component-selector scoring shape transfers)
    primitives          jsonb NOT NULL DEFAULT '[]'::jsonb,
    section_types       jsonb NOT NULL DEFAULT '[]'::jsonb,
    destination_roles   jsonb NOT NULL DEFAULT '[]'::jsonb,
    funnel_stage        text CHECK (funnel_stage IN ('awareness','consideration','conversion')),
    suitable_site_types jsonb NOT NULL DEFAULT '[]'::jsonb,

    -- The contract: every trigger names an OBSERVABLE outcome; navigating
    -- triggers name a destination ROLE, never a URL.
    contract jsonb NOT NULL,

    -- Clauses the trigger list cannot hold (HARVEST_01 §§3.1/3.6-3.9, HARVEST_02 §3.1)
    states                      jsonb NOT NULL DEFAULT '[]'::jsonb,
    automatic_triggers          jsonb NOT NULL DEFAULT '[]'::jsonb,
    data_contract               jsonb NOT NULL DEFAULT '{}'::jsonb,
    degraded_states             jsonb NOT NULL DEFAULT '[]'::jsonb,
    entry_points                jsonb NOT NULL DEFAULT '[]'::jsonb,
    requires_component_contract jsonb NOT NULL DEFAULT '[]'::jsonb,
    requires_invariant          jsonb NOT NULL DEFAULT '[]'::jsonb,

    -- Acceptance side: template + the placeholders a fork must bind.
    binding_schema    jsonb NOT NULL DEFAULT '{}'::jsonb,
    criteria_template jsonb,

    status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','approved','proven')),
        -- draft = visible, UNSELECTABLE; approved = pattern-council pass;
        -- proven = first live green run of bound criteria (evidence, never review)

    source         text,
    source_agent   text,
    harvested_from text,
    created_by     text,
    created_at     timestamptz NOT NULL DEFAULT now(),
    updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_exp_patterns_kind          ON experience_patterns (kind) WHERE status <> 'draft';
CREATE INDEX IF NOT EXISTS idx_exp_patterns_section_types ON experience_patterns USING gin (section_types);
CREATE INDEX IF NOT EXISTS idx_exp_patterns_dest_roles    ON experience_patterns USING gin (destination_roles);
CREATE INDEX IF NOT EXISTS idx_exp_patterns_site_types    ON experience_patterns USING gin (suitable_site_types);

-- ---------------------------------------------------------------------------
-- 4. site_experiences — per-site forks. The base entry is invariant; every
--    concrete selector, page and value lives here.
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS site_experiences (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    site_id      uuid NOT NULL,
    pattern_name text NOT NULL REFERENCES experience_patterns(name),
    instance_key text NOT NULL DEFAULT 'default',
    bindings     jsonb NOT NULL DEFAULT '{}'::jsonb,
    status       text NOT NULL DEFAULT 'proposed'
                 CHECK (status IN ('proposed','bound','verified','broken')),
    last_checked_at        timestamptz,
    last_check_result      jsonb,
    built_from_plan_version uuid,
    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (site_id, pattern_name, instance_key)
);

CREATE INDEX IF NOT EXISTS idx_site_experiences_site ON site_experiences (site_id, status);

-- ---------------------------------------------------------------------------
-- 5. Seed the two invariants harvest has evidence for. Entries themselves are
--    NOT seeded here — they are written through the validating write path, so
--    that the first rows in the register are ones the validator accepted (a
--    register seeded by raw INSERT would have no proof its own contract holds).
-- ---------------------------------------------------------------------------
INSERT INTO experience_invariants (name, clause, rationale, sightings, created_by)
VALUES
 ('no-inert-control',
  'A control that cannot do anything must not be presented as a control: no placeholder href, no tab stop, no rendered button.',
  'Six independent implementations across two sites and five authors, none referring to each other. The inverse — presenting a control that does nothing — is the defect class of bugs_open/023 and of the 2026-07-22 gauntlet correction, where wiring a dead control to an invisible effect fixed the letter and missed the point.',
  '[{"where":"vonc provocations archive","mechanism":"loader strips the template href and tabindex from a row with no case","evidence":"archive_loader_new.js --static branch; live check B.3"},
    {"where":"vonc provocation-card / lobby-grid","mechanism":"the CTA is not rendered when the record carries no destination","evidence":"provocations.json today.primary_cta"},
    {"where":"fundamentallyai hero-card-carousel","mechanism":"arrows and pause hidden when fewer than two cards","evidence":"behaviour.js slides.length < 2"},
    {"where":"fundamentallyai hero-card-carousel","mechanism":"the pause control is never emitted when auto-advance is off","evidence":"live /capabilities.html ships zero [data-hcc-pause]"},
    {"where":"fundamentallyai swipeable-insight-carousel","mechanism":"template conditional emits no anchor without a link_url","evidence":"template.html {{if .link_url}}; live page links 2 cards of N"},
    {"where":"fundamentallyai people-feature-block","mechanism":"same template conditional","evidence":"template.html; live /about.html renders one anchor"}]'::jsonb,
  'experience-register-p2'),
 ('pointer-behaviour-has-a-keyboard-equal',
  'Any state a pointer can reach — a hover reveal, a drag, a swipe — is reachable by keyboard, producing the same observable outcome.',
  'The first clause dropped when behaviour is re-invented, and invisible to every check we currently run. Two independent implementations honour it today.',
  '[{"where":"fundamentallyai image-hover-card-grid","mechanism":":hover and :focus-visible written as a pair on every reveal rule","evidence":"shipped page CSS, 4 focus-visible rules"},
    {"where":"fundamentallyai hero-card-carousel","mechanism":"ArrowLeft/ArrowRight do exactly what the arrow controls do, with the same live-region announcement","evidence":"behaviour.js keydown handler"}]'::jsonb,
  'experience-register-p2')
ON CONFLICT (name) DO NOTHING;

COMMIT;

-- Verify
SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint
WHERE conname IN ('doc_plans_subject_type_check','doc_notes_subject_type_check');
SELECT name, status, jsonb_array_length(sightings) AS sightings FROM experience_invariants ORDER BY name;
SELECT to_regclass('experience_patterns') AS patterns, to_regclass('site_experiences') AS forks;
