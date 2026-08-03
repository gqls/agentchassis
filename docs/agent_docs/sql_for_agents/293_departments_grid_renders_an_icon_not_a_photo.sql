-- 293 — departments-grid renders its icon as an ICON, not as an <img src>
--
-- WHAT: rewrites content_components.html_template for `departments-grid` so the
-- per-department `icon` field is emitted as <i data-lucide="{{.icon}}"></i>
-- inside a styled badge, instead of <img src="{{.icon}}" class="member-photo">.
-- The .member-photo rule is replaced by .member-icon, same 120px circle, so the
-- grid's geometry is unchanged.
--
-- WHY. departments-grid was forked from a team/staff component — the markup
-- still says team-section / team-member / member-photo / member-title /
-- member-bio — and repurposed to show DEPARTMENTS, whose input_schema declares
--
--     "departments": { "items": { "icon": "string", "name": "string", ... } }
--
-- an icon NAME, not a photo URL. The template kept the photo's <img src>, so
-- every department rendered as
--
--     <img src="cpu" alt="Automation & Workflow Department" class="member-photo">
--
-- and the browser painted a broken-image icon. Confirmed live before this file
-- was written: https://finetuning.uk/cpu, /network, /database all return 404.
--
-- This is what the owner meant by "the finetuning site is looking terrible":
-- eight broken images down the middle of the homepage and eleven more on
-- /about.html, each one 120px across.
--
-- THE CENSUS, so the blast radius is measured rather than assumed. Over every
-- rendered surface in the fleet (page_components + site_components) on
-- 2026-08-03, `<img src="…">` with no '/' and no '.' in the value:
--
--     ai-agent-orchestration.com   16 occurrences   departments-grid
--     finetuning.uk                15 occurrences   departments-grid
--
-- 31 occurrences, 2 sites, 1 component, 4 page_components rows. Every one is a
-- broken image, and this file fixes all of them at the source.
--
-- WHY <i data-lucide> IS THE RIGHT TARGET rather than a guess. It is what this
-- fleet already does: the `features` component renders
--
--     {{if .icon}}<div class="feature-icon"><i data-lucide="{{.icon}}"></i></div>{{end}}
--
-- and on finetuning.uk it renders on the SAME PAGE as the broken one, working,
-- with lucide.min.js already loaded and lucide.createIcons() already called.
-- Verified 2026-08-03 that all four affected pages load lucide:
--
--     finetuning.uk/  finetuning.uk/about.html
--     ai-agent-orchestration.com/  ai-agent-orchestration.com/about.html
--
-- so the replacement cannot render into a page that has no icon library.
--
-- WHAT THIS FILE DOES NOT FIX, stated so nobody reads it as a full repair.
-- ai-agent-orchestration.com/index.html supplies eight values that are not
-- lucide icon names at all — strategy, research, content, design, development,
-- quality, operations, data. lucide leaves an unknown name untouched, so those
-- eight go from a broken-image icon to an empty badge. That is an improvement
-- and it is not a fix: the CONTENT is wrong there, which is a content_data
-- repair on that site, not a template one. finetuning.uk's fifteen are all real
-- lucide names (cpu, network, database, shield, search, sliders, map, layout,
-- workflow, globe, layers, lock, settings, bar-chart-2, download-cloud) and all
-- fifteen render.
--
-- ORDER: this is DB config, so it is live the moment it commits. The rendered
-- pages do NOT change until the four page_components rows are re-rendered —
-- that is the framework's job (rerender-pages), not this file's.
--
-- APPLY:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 293_….sql

BEGIN;

-- Backup, per fleet convention: one table, dated, this component only.
CREATE TABLE IF NOT EXISTS content_components_bak_20260803_deptgrid AS
SELECT * FROM content_components WHERE name = 'departments-grid';

UPDATE content_components
SET html_template = replace(
      replace(
        html_template,
        '{{if .icon}}<img src="{{.icon}}" alt="{{.name}} Department" class="member-photo">{{end}}',
        '{{if .icon}}<div class="member-icon"><i data-lucide="{{.icon}}"></i></div>{{end}}'
      ),
      -- The photo rule becomes the badge rule. Same 120px circle and same
      -- margins, so the grid does not reflow; object-fit goes (nothing to crop)
      -- and flex centring arrives (an SVG must be centred in the badge).
      '.member-photo {
    width: 120px;
    height: 120px;
    border-radius: 50%;
    object-fit: cover;
    margin: 0 auto 1.5rem;
    background: #e0e0e0;
}',
      '.member-icon {
    width: 120px;
    height: 120px;
    border-radius: 50%;
    margin: 0 auto 1.5rem;
    background: #e0e0e0;
    display: flex;
    align-items: center;
    justify-content: center;
}
.member-icon svg {
    width: 48px;
    height: 48px;
    stroke-width: 1.5;
}'
    ),
    updated_at = now()
WHERE name = 'departments-grid';

-- VERIFY, as DO/RAISE rather than a SELECT: a verify block made of SELECTs
-- cannot stop the COMMIT, because ON_ERROR_STOP does not fire on a non-empty
-- result. Every one of these must hold or the transaction aborts.
DO $$
DECLARE
    tpl text;
BEGIN
    SELECT html_template INTO tpl FROM content_components WHERE name = 'departments-grid';

    IF tpl IS NULL THEN
        RAISE EXCEPTION '293: departments-grid not found';
    END IF;

    -- The defect is gone.
    IF tpl LIKE '%<img src="{{.icon}}"%' THEN
        RAISE EXCEPTION '293: the <img src="{{.icon}}"> is still present — replace() did not match';
    END IF;
    IF tpl LIKE '%member-photo%' THEN
        RAISE EXCEPTION '293: .member-photo survives — the CSS replace() did not match';
    END IF;

    -- The replacement is present, and is the shape the fleet already uses.
    IF tpl NOT LIKE '%<i data-lucide="{{.icon}}"></i>%' THEN
        RAISE EXCEPTION '293: the lucide element was not written';
    END IF;
    IF tpl NOT LIKE '%.member-icon svg%' THEN
        RAISE EXCEPTION '293: the badge CSS was not written';
    END IF;

    -- Nothing else moved: the surrounding markup is untouched.
    IF tpl NOT LIKE '%{{if .description}}<p class="member-bio">{{.description}}</p>{{end}}%' THEN
        RAISE EXCEPTION '293: surrounding markup changed — replace() over-matched';
    END IF;

    RAISE NOTICE '293 OK: departments-grid renders <i data-lucide> in a .member-icon badge';
END $$;

COMMIT;
