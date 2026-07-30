-- SQL_2026-07-30t_planner_drop_hardcoded_icon_ground.sql
--
-- THE SECOND HALF OF THE IMAGERY-CAUSE FIX.
-- Written 2026-07-29 as OWED. **APPLIED 2026-07-30 11:23Z**, precondition met.
--
-- ============================================================================
-- PRECONDITION — SATISFIED, and here is the evidence rather than the claim
-- ============================================================================
--
-- The rule was: a chassis carrying bd9ebfec6 (+ follow-up 88dee2a8d) must be
-- ROLLED and POD-VERIFIED first. Removing the hardcoded colour before the
-- palette derivation is live would leave flat-vector kinds (icon, sprite_sheet,
-- content_hero) with NO colour direction at all — worse than a wrong one,
-- because the model then invents a colour per image and it looks deliberate.
--
-- Verified on v1.0.1207, BOTH replicas (agent-chassis-6c448c66d6-fjpd7 and
-- -kmm9b), three markers rather than one:
--
--   "no other background colour"                -> 1, 1   (the new clause: ADDED)
--   "composedPaletteDirection: query failed"    -> 0, 0   (first draft's string: DELETED)
--   "composed palette unavailable"              -> 1, 1   (its replacement: ADDED)
--
-- The middle one is the load-bearing check. A positive control alone would prove
-- only that SOME build after bd9ebfec6 shipped; the delete-marker proves the
-- binary also carries 88dee2a8d, because that commit is what replaced the string.
-- A marker your change DELETED is the strongest kind of deploy evidence.
--
-- A COMMIT BEING IN HEAD IS NOT EVIDENCE THE BINARY HAS IT, which is why the
-- greps and not the timeline settled this. The timeline was merely favourable
-- and remains [INFERRED]: bd9ebfec6 landed 07-29 18:19, 88dee2a8d at 18:38, and
-- v1.0.1205/1206 rolled at 21:39/21:46. A RETAG IS NOT A REBUILD — two tags have
-- shared one image id on this fleet before, built 56 minutes before the fix they
-- were said to carry.
--
-- ============================================================================
-- WHAT THIS CHANGES, AND WHAT IT DELIBERATELY DOES NOT
-- ============================================================================
--
-- build-site-planner's prompt_template pins every icon prompt to a LIGHT ground:
--
--   'a darker grey (#4A4A4A) line on a flat solid light grey (#EEEEEE) background'
--
-- unconditionally, for every site whatever its colour scheme. Measured
-- 2026-07-29: 92 site_plan_imagery rows across 14 plans carry it; 62 on 9 sites'
-- current plans. Six of the ten sites with a resolved composition are dark. On
-- dartsonline it produced 17 icons that are all correct renderings of a wrong
-- instruction, and all unusable.
--
-- THE FLATNESS GUARDS STAY, VERBATIM. They are the reason the literal was added
-- in the first place (053_build_site_planner.sql:2300, snapshot label "icon
-- background: transparent/plain -> flat selectable grey (embrace the chip)") and
-- they are load-bearing: without them the model produces transparency,
-- checkerboard artefacts and gradients. Only the COLOUR is removed.
--
-- The prompt now defers the colour to the site instead of naming one, because
-- generate_image prepends a palette-derived clause at generation time — either
-- from the site's imagery_style_guide, or (as of bd9ebfec6) from the composed
-- palette for a dark site that has no guide.
--
-- THIS DOES NOT REWRITE THE 92 EXISTING ROWS. They keep their #EEEEEE. That is
-- deliberate and is why the derived clause ends with "and no other background
-- colour" — it has to beat a literal still sitting in the subject prompt it is
-- prepended to. Rewriting the existing rows is a separate decision: it would
-- change 9 sites' plans at once, and 4 of them are light sites where the literal
-- is harmless.
--
-- ============================================================================

-- BEGIN;   -- uncomment together with the COMMIT at the foot

SELECT snapshot_agent('build-site-planner',
  'icon ground: drop the hardcoded #EEEEEE literal; the palette supplies the colour at generation time (bd9ebfec6)');

-- 1. Sanity: confirm BOTH target fragments are present in the LIVE row.
--    Expect (t, t). If either is f, STOP — the template has moved since
--    2026-07-29 and these replacements will silently match nothing.
SELECT
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%specify a flat solid light grey (#EEEEEE) background%' AS has_guidance_fragment_expect_t,
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%a darker grey (#4A4A4A) line on a flat solid light grey (#EEEEEE) background%' AS has_worked_example_fragment_expect_t
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 2. Apply both replacements.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,prompt_template}',
        to_jsonb(
            replace(
                replace(
                    default_config #>> '{workflow,steps,plan_site,config,prompt_template}',

                    -- (1) the guidance bullet
                    '; specify a flat solid light grey (#EEEEEE) background with a darker grey (#4A4A4A) line icon — one single uniform background colour, no gradients, no shadows, no checkerboard pattern, no transparency, no photorealism. Icons are placed inside a styled container ("chip") on the page, so an opaque flat light-grey background is correct and expected — do NOT request or imply transparency',
                    '; specify a flat solid single-colour background and single-weight line icon — one single uniform background colour, no gradients, no shadows, no checkerboard pattern, no transparency, no photorealism. Icons are placed inside a styled container ("chip") on the page, so an OPAQUE flat background is correct and expected — do NOT request or imply transparency. DO NOT NAME A COLOUR OR A HEX VALUE: the site''s own palette is prepended to this prompt at generation time, and a colour written here would contradict it'
                ),

                -- (2) the worked-example substring (matches all three icon entries)
                'a darker grey (#4A4A4A) line on a flat solid light grey (#EEEEEE) background, one single uniform background colour, no gradients, no shadows, no checkerboard, no transparency, no photorealism',
                'single-weight linework on a flat solid single-colour background, one single uniform background colour, no gradients, no shadows, no checkerboard, no transparency, no photorealism, and no named colour or hex value — the site palette supplies the colours'
            )
        )
    )
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 3. Verify. The FIRST two are the ones that matter — a hex surviving anywhere
--    in the icon guidance means a replacement missed.
SELECT
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%#EEEEEE%'             AS still_has_grey_hex_expect_f,
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%#4A4A4A%'             AS still_has_line_hex_expect_f,
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%no checkerboard%'     AS kept_checkerboard_guard_expect_t,
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%do NOT request or imply transparency%' AS kept_transparency_guard_expect_t,
    (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
        LIKE '%the site palette supplies the colours%' AS has_new_deferral_expect_t
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 4. Read the edited region rather than trusting the booleans. A replace() that
--    matched nothing returns the original string and reports success.
SELECT substring(
    default_config #>> '{workflow,steps,plan_site,config,prompt_template}'
    FROM 'single-weight linework.{0,220}'
) AS edited_region_preview
FROM agent_definitions
WHERE type = 'build-site-planner' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;
-- ROLLBACK;   -- was the safe default; applied 2026-07-30 after a clean dry run

-- ============================================================================
-- APPLIED 2026-07-30 11:23Z — what actually happened
-- ============================================================================
--
-- Dry run first (BEGIN … ROLLBACK), then the same file with COMMIT. Both reads:
--
--   has_guidance_fragment_expect_t | has_worked_example_fragment_expect_t
--   t                              | t
--   UPDATE 1
--   still_has_grey_hex_expect_f | still_has_line_hex_expect_f | kept_checkerboard_guard_expect_t | kept_transparency_guard_expect_t | has_new_deferral_expect_t
--   f                           | f                           | t                                | t                                | t
--
-- Then re-read on a FRESH CONNECTION, because the verify block above runs inside
-- the same transaction as the UPDATE and can only ever agree with itself:
--   hex_gone f · line_hex_gone f · guard_kept t · deferral t
--
-- A TRAP IN THIS FILE'S OWN SHAPE, worth naming. As originally written, `BEGIN;`
-- was commented out and `ROLLBACK;` was live. Run like that, every statement
-- autocommits and the trailing ROLLBACK is a no-op with a warning — so the
-- "safe default" would have COMMITTED the change while appearing not to. The
-- dry run only worked because BEGIN was uncommented for it. **A rollback at the
-- foot of a script is only safe if the BEGIN at the head is real.**
--
-- SNAPSHOT: TWO OVERLOADS, TWO DESTINATIONS, and checking the wrong one makes a
-- successful snapshot look like a silent no-op. I called
-- `snapshot_agent(type, reason)`; it prints "Snapshot captured" and writes to
-- **agent_definitions_backup**. The one-argument `snapshot_agent(type)` instead
-- inserts an `is_snapshot = true` row into `agent_definitions`. I checked the
-- latter, found 0 rows fleet-wide for this agent, and was one step from
-- reporting that the snapshot had silently failed. It had not:
--
--   SELECT snapshot_taken_at, snapshot_reason,
--          (default_config #>> '{workflow,steps,plan_site,config,prompt_template}')
--            LIKE '%#EEEEEE%' AS backup_has_old_hex
--   FROM agent_definitions_backup WHERE type='build-site-planner'
--   ORDER BY snapshot_taken_at DESC LIMIT 1;
--   -- 2026-07-30 11:23:47Z | icon ground: drop the hardcoded #EEEEEE… | t
--
-- `backup_has_old_hex = t` is the check that matters: it proves the backup holds
-- the PRE-change text and is therefore restorable. A snapshot row that exists but
-- carries the post-change config would restore nothing.

-- ============================================================================
-- AFTER APPLYING
-- ============================================================================
--
-- There is nothing to see until a site is REPLANNED — this changes what future
-- plan rows say, not any existing row. So do not go looking for a change in
-- site_plan_imagery afterwards; there will not be one, and its absence is not
-- evidence the edit failed.
--
-- The first real evidence is the next icon generated on a site with no
-- imagery_style_guide. fundamentallyai.com and vonc.com are the two whose
-- behaviour the Go half changes (measured: 32 sites -> 10 with a resolved
-- composition -> 6 dark -> 4 of those already have a guide). Neither is this
-- workstream's to drive; contribute the observation rather than starting work
-- on them.
--
-- And READ THE PNG. The whole reason this file exists is that seventeen images
-- were generated, marked active, and nobody looked.
