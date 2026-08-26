-- 649 — make `hero-tool` and `case-studies-hero` IMAGE-CAPABLE, fleet-wide
--
-- OWNER INSTRUCTION 2026-08-26, verbatim: "please make the components image capable, it can update
-- across the fleet, all forty pages and more, there is a lack of images everywhere and this will
-- help."
--
-- WHY. `bugs_open/412` §7: these two components have NO image branch, so a page-scoped
-- needs_imagery item for a page using them is accepted, generated, deployed and closed `complete`
-- while the asset is an orphan the moment it lands. Three of this lane's nine canary pages were in
-- that state. The fix is to give the components the branch their five siblings already have.
--
-- BLAST RADIUS [MEASURED 2026-08-26]:
--   hero-tool          43 live instances across 13 sites — 0 currently carry an image value
--   case-studies-hero   5 live instances across  4 sites — 1 DOES carry one it cannot display
-- So 47 of 48 gain a LATENT capability and change nothing visually today; exactly ONE page gains a
-- hero image the moment it re-renders. That single page is the change's own verification.
--
-- SHAPE COPIED FROM THE WORKING SIBLINGS, not invented: `use-cases-hero` / `services-hero` /
-- `about-hero` / `contact-hero` all use `{{if or .hero_url .background_image}}` on the class and
-- an inline dark-scrim background. Same gradient values, same cover/center.
--   * case-studies-hero also gets `--hero-ink: #fff` because its non-imaged branch inks from the
--     palette; over a dark scrim the ink must be white, exactly as its siblings do it.
--   * hero-tool does NOT need it: its band is already dark and its `--section-text` already
--     resolves to `--color-primary-text`, so the ink is light on both branches.
--
-- ⚠ NO `--imaged` CSS RULE IS ADDED, deliberately, and this differs from the siblings. Theirs carry
-- `.hero-X--imaged { background-image: inherit; }`, which cannot be what makes the image appear —
-- the INLINE style does that, and `inherit` would pull from the parent. The class marker is added
-- for CSS hooks and parity; the vestigial rule is not copied forward.
--
-- ⚠ NO FAN-OUT WITH THIS. A template change ships nothing until a page re-renders (bugs_open/283
-- §13) — but re-rendering 48 instances now would be pure churn, because 47 of them have no image
-- to show. They pick the capability up on their next natural render. The one that CAN show an
-- image is worth a single targeted re-render; see the verify note below.
--
-- Rollback: 649_..._ROLLBACK.sql

BEGIN;

UPDATE content_components
   SET html_template = replace(html_template, '<section class="hero hero-case-studies" data-component="hero-case-studies">', '<section class="hero hero-case-studies{{if or .hero_url .background_image}} hero-case-studies--imaged{{end}}" data-component="hero-case-studies"{{if or .hero_url .background_image}} style="background-image: linear-gradient(rgba(6,11,20,0.62), rgba(6,11,20,0.72)), url(''{{or .hero_url .background_image}}''); background-size: cover; background-position: center; --hero-ink: #fff;"{{end}}>'),
       updated_at = now()
 WHERE name = 'case-studies-hero' AND html_template LIKE '%<section class="hero hero-case-studies" data-component="hero-case-studies">%';

UPDATE content_components
   SET html_template = replace(html_template, '<section class="hero-tool-section" data-component="hero-tool">', '<section class="hero-tool-section{{if or .hero_url .background_image}} hero-tool-section--imaged{{end}}" data-component="hero-tool"{{if or .hero_url .background_image}} style="background-image: linear-gradient(rgba(6,11,20,0.62), rgba(6,11,20,0.72)), url(''{{or .hero_url .background_image}}''); background-size: cover; background-position: center;"{{end}}>'),
       updated_at = now()
 WHERE name = 'hero-tool' AND html_template LIKE '%<section class="hero-tool-section" data-component="hero-tool">%';

DO $$
DECLARE n_capable int; n_others int;
BEGIN
    SELECT count(*) INTO n_capable FROM content_components
     WHERE name IN ('hero-tool','case-studies-hero')
       AND html_template LIKE '%hero_url%' AND html_template LIKE '%background_image%';
    IF n_capable <> 2 THEN
        RAISE EXCEPTION '649: % of 2 components are image-capable — the needle did not match, check the section tag verbatim', n_capable;
    END IF;

    -- The five siblings must be untouched: this migration names two components and must not have
    -- reached a third by a loose needle.
    SELECT count(*) INTO n_others FROM content_components
     WHERE name IN ('hero','services-hero','about-hero','contact-hero','use-cases-hero')
       AND updated_at > now() - interval '1 minute';
    IF n_others <> 0 THEN
        RAISE EXCEPTION '649: % sibling component(s) were also modified — the replace was too broad', n_others;
    END IF;

    RAISE NOTICE '649 OK: hero-tool and case-studies-hero are image-capable; 5 siblings untouched';
END $$;

COMMIT;
