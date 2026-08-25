-- 625_aiao_owned_tool_pages_contrast.sql
--
-- ai-agent-orchestration.com — clears 9 of the 17 firm contrast failures found by the
-- FIRST full-site audit of this lane (2026-08-25: all 42 active pages, 40 measured).
-- Every figure below was measured that evening; none is carried forward.
--
-- ⚠ WHY THIS PATCHES `page_components.rendered_html` AND NOT ONLY TEMPLATES.
-- All three pages are `rebuild_policy='owned'`, and that is CORRECT — each holds its
-- whole tool (calculator markup AND its `<script>`) in a single component. The
-- ordinary route this lane has used all week — a page-scoped `template_changed`
-- rerender — is REFUSED on an owned page by `save_page_sections`, and the refusal is
-- the thing protecting the tool.
--
-- ⚠ AND THE TWO OBVIOUS WORKAROUNDS ARE BOTH WRONG:
--   · **Flipping `rebuild_policy` to 'generic'** to let the rerender through is the
--     documented tool-clobber: the composition loop commits freshly-written HTML to
--     the deploying repo ONE STEP BEFORE `save_page_sections` refuses, so the
--     calculator is replaced with prose and SHIPPED before any DB refusal saves you.
--     Calculators have already been destroyed this way (LANDMINE; migration 367 →
--     377 re-lock).
--   · **`refresh_owned_page_chrome.sh`** (the leopardess route) is safe but INERT
--     here: it re-renders in ASSEMBLE mode, which re-assembles the STORED section
--     HTML. Stored HTML is exactly what carries the stale CSS, so assemble mode
--     reproduces the defect. It fixes chrome, not section CSS.
-- So the correct operation is a surgical patch of the stored HTML: no regeneration,
-- no policy flip, no window in which a generic build could see these pages.
-- Precedent: migration `393`, which patched `page_components.rendered_html` for three
-- tool components for the same structural reason.
-- ⚠ Safe to do NOW in a way it was not before: `bugs_closed/229` (fixed and live,
-- v1.0.1276) gave `rendered_html` writes a comparison and an archive, so a
-- hand-patched row is recoverable. Before that fix this file would have been reckless.
--
-- ── THE FOUR DEFECT SHAPES, each measured and each verified against its ground ──
--
-- (a) BARE FOREGROUND `--color-primary` — migration 456's defect, on pages 456 never
--     reached. `--color-primary` is `#0D1117` here and `--color-surface` is the SAME
--     value, so the text is painted in its own background.
--       h2            #0D1117 on #080B10  = 1.04:1  ->  ink #768eb2 = 5.90:1  (need 3.0)
--       .ace-legend   #0D1117 on #0D1117  = 1.00:1  ->  ink #768eb2 = 5.66:1  (need 4.5)
--     Fix: the documented two-level fallback `var(--color-primary-ink,var(--color-primary))`.
--     ⚠ `.ace-legend`'s TEMPLATE was already repointed by 456 — only the stored HTML
--     is stale (last written 2026-05-01). This file brings the artefact into line
--     with a template that has been correct for days.
--
-- (b) A LABEL ON A PRIMARY FILL — the shape 457 fixed for `.stats-cta`, in its
--     primary (not accent) variant. `palette_specialised_slots.go:105-107` states the
--     rule: `--color-primary-text` on a primary-filled button.
--       .estimate-btn  background:var(--color-primary); color:var(--color-background)
--       .calc-btn      background:var(--color-primary); color:var(--color-on-primary, var(--color-background))
--       both           #080B10 on #0D1117 = 1.04:1  ->  #ffffff = 18.92:1  (need 4.5)
--     ⚠ `.calc-btn` already TRIES `--color-on-primary` and that token IS NOT EMITTED
--     on this site (probed live: empty string), which is precisely why it falls
--     through to `--color-background` and fails. A fallback chain is only as good as
--     its first emitted link.
--     `--color-primary-text` IS emitted (`#ffffff`, probed live) and is already the
--     idiom on sibling calculators fleet-wide (`gripper-payload-calculator`,
--     three `loanzy-uk` tools), so this is convergence, not invention.
--
-- (c) AN INLINE `style="… color: #666 …"` — not a rule at all, so no stylesheet or
--     template edit could ever have reached it.
--       #666666 on #0D1117 = 3.30:1  ->  --color-text-muted #8B949E = 6.15:1  (need 4.5)
--     Exactly 2 occurrences, both the failing labels ("Bits of Entropy",
--     "Crack Time (GPU)"). ⚠ This is why the earlier rule-shaped queries returned
--     NOTHING for this page and it read as clean — the colour was never in a `{...}`.
--
-- ── TEMPLATES ARE FIXED TOO, WHERE THEY CARRY THE DEFECT ────────────────────────
-- Patching only the artefact would leave the next regeneration to reintroduce (b).
-- So the two aiao tool templates are repointed as well. ⚠ SCOPED TO THIS SITE'S
-- COMPONENTS BY ID. `.calc-btn` exists in THIRTEEN templates fleet-wide across many
-- sites (measured 2026-08-25) and most are other sites' forks; repointing the class
-- would be a fleet change this lane has not measured. `.ace-legend`'s template needs
-- nothing — 456 already did it.
--
-- ── WHAT THIS FILE DOES NOT FIX, STATED SO THE REMAINDER IS NOT MISREAD ─────────
--   · `/contact.html` white-on-amber `.form-submit` (1) — `contact-form` is on **20
--     SITES** (2026-08-25). A fleet change; consumers must be told first
--     (owner ruling 2026-07-29 §3).
--   · `/tools/automation-savings-estimator/` (3) and `/tools/build-vs-buy-analyzer/`
--     (1) — their stored HTML ALREADY carries the ink token, so the cause is
--     something else and guessing would waste a cycle. Needs diagnosis.
--   · 2 pages that CANNOT be measured (`ai-readiness-quiz`,
--     `tool-ai-agent-roi-estimator`) — "probe produced no result", reproducible.
-- 9 fixed here + 1 already fixed on contact + 7 remaining = the 17.
--
-- NO RE-RENDER IS NEEDED OR WANTED. This writes the served artefact directly; a
-- rerender would be refused anyway. Verify over HTTP, not in the database.
--
-- ROLLBACK: 625_aiao_owned_tool_pages_contrast_ROLLBACK.sql (byte-exact from migration_backups)

BEGIN;

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '625_aiao_owned_tool_pages_contrast', 'page_components', pc.id::text,
       jsonb_build_object('rendered_html', pc.rendered_html),
       'pre-625 rendered_html for ' || p.name
FROM page_components pc JOIN pages p ON p.id = pc.page_id
WHERE pc.id IN ('b2b7acbd-f9b2-420f-97c7-3713726d8a57',
                '34484400-c657-4b15-be95-7746a2370aa0',
                'e073baa8-6b83-48e1-8db5-3dc9c93eb52f');

INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '625_aiao_owned_tool_pages_contrast', 'content_components', cc.id::text,
       jsonb_build_object('html_template', cc.html_template),
       'pre-625 html_template for ' || cc.name
FROM content_components cc
WHERE cc.name IN ('tool-agent-complexity-estimator-leopardessconsulting-co-uk-ai-agent-orchestration-com',
                  'tool-llm-cost-calculator-ai-agent-orchestration-com');

-- (a) + (b) on the stored artefact of the two ink-defect tool pages.
UPDATE page_components
SET rendered_html = replace(replace(replace(
      rendered_html,
      'color: var(--color-primary);',      'color: var(--color-primary-ink, var(--color-primary));'),
      'color: var(--color-background);',   'color: var(--color-primary-text, var(--color-background));'),
      'color: var(--color-on-primary, var(--color-background));',
                                           'color: var(--color-primary-text, var(--color-background));'),
    updated_at = now()
WHERE id IN ('b2b7acbd-f9b2-420f-97c7-3713726d8a57','e073baa8-6b83-48e1-8db5-3dc9c93eb52f');

-- (c) the two inline greys on password-entropy.
UPDATE page_components
SET rendered_html = replace(rendered_html, 'color: #666;', 'color: var(--color-text-muted, #666);'),
    updated_at = now()
WHERE id = '34484400-c657-4b15-be95-7746a2370aa0';

-- (b) in the templates, so a regeneration cannot reintroduce it. By NAME, this site only.
UPDATE content_components
SET html_template = replace(replace(
      html_template,
      'color: var(--color-background);',   'color: var(--color-primary-text, var(--color-background));'),
      'color: var(--color-on-primary, var(--color-background));',
                                           'color: var(--color-primary-text, var(--color-background));'),
    updated_at = now()
WHERE name IN ('tool-agent-complexity-estimator-leopardessconsulting-co-uk-ai-agent-orchestration-com',
               'tool-llm-cost-calculator-ai-agent-orchestration-com');

DO $$
DECLARE b int; n int;
BEGIN
  SELECT count(*) INTO b FROM migration_backups WHERE migration_name='625_aiao_owned_tool_pages_contrast';
  IF b <> 5 THEN RAISE EXCEPTION '625: expected 5 backup rows (3 artefacts + 2 templates), wrote %', b; END IF;

  -- (a) no bare foreground primary may remain in the two patched artefacts.
  SELECT count(*) INTO n FROM page_components
   WHERE id IN ('b2b7acbd-f9b2-420f-97c7-3713726d8a57','e073baa8-6b83-48e1-8db5-3dc9c93eb52f')
     AND rendered_html ~ 'color:\s*var\(--color-primary\)\s*;';
  IF n <> 0 THEN RAISE EXCEPTION '625: % artefact(s) still carry a bare foreground --color-primary', n; END IF;

  -- (b) no primary-filled label may still resolve to the page background.
  SELECT count(*) INTO n FROM page_components
   WHERE id IN ('b2b7acbd-f9b2-420f-97c7-3713726d8a57','e073baa8-6b83-48e1-8db5-3dc9c93eb52f')
     AND rendered_html ~ 'color:\s*var\(--color-(background|on-primary)';
  IF n <> 0 THEN RAISE EXCEPTION '625: % artefact(s) still paint a button label with the page background', n; END IF;

  -- (c) both inline greys tokenised, and NO bare #666 left.
  SELECT (SELECT count(*) FROM regexp_matches(rendered_html,'var\(--color-text-muted, #666\)','g')) INTO n
    FROM page_components WHERE id='34484400-c657-4b15-be95-7746a2370aa0';
  IF n <> 2 THEN RAISE EXCEPTION '625: expected 2 tokenised greys on password-entropy, found %', n; END IF;
  SELECT count(*) INTO n FROM page_components
   WHERE id='34484400-c657-4b15-be95-7746a2370aa0' AND rendered_html ~ 'color:\s*#666\s*;';
  IF n <> 0 THEN RAISE EXCEPTION '625: a bare inline #666 survives on password-entropy'; END IF;

  -- templates repointed.
  SELECT count(*) INTO n FROM content_components
   WHERE name IN ('tool-agent-complexity-estimator-leopardessconsulting-co-uk-ai-agent-orchestration-com',
                  'tool-llm-cost-calculator-ai-agent-orchestration-com')
     AND html_template LIKE '%var(--color-primary-text, var(--color-background))%';
  IF n <> 2 THEN RAISE EXCEPTION '625: expected both templates repointed, % qualify', n; END IF;

  -- ⚠ THE SCOPE GUARD: no OTHER site's component may have been MODIFIED by this file.
  -- `.calc-btn` lives in 13 templates fleet-wide and this file must reach exactly two.
  -- ⚠ The first version of this guard asserted that no other component CARRIES the new
  -- literal, and it fired on 8 rows that already had it — `var(--color-primary-text,
  -- var(--color-background))` is an existing fleet idiom, which is part of why it was
  -- chosen. That guard tested the wrong proposition: presence, not authorship. Same
  -- mistake as 469's first bare-var guard. The right test is "did I write to it".
  SELECT count(*) INTO n FROM content_components
   WHERE updated_at >= now() - interval '1 second'
     AND name NOT IN ('tool-agent-complexity-estimator-leopardessconsulting-co-uk-ai-agent-orchestration-com',
                      'tool-llm-cost-calculator-ai-agent-orchestration-com');
  IF n <> 0 THEN RAISE EXCEPTION '625: % component(s) outside this file''s two names were written in this transaction; scope escaped', n; END IF;

  SELECT count(*) INTO n FROM page_components
   WHERE updated_at >= now() - interval '1 second'
     AND id NOT IN ('b2b7acbd-f9b2-420f-97c7-3713726d8a57','34484400-c657-4b15-be95-7746a2370aa0',
                    'e073baa8-6b83-48e1-8db5-3dc9c93eb52f');
  IF n <> 0 THEN RAISE EXCEPTION '625: % page_component(s) outside the three targets were written; scope escaped', n; END IF;

  -- The tool must still be there. A calculator without its script is the damage this
  -- whole file is shaped to avoid, so assert it rather than trust the regex.
  SELECT count(*) INTO n FROM page_components
   WHERE id IN ('b2b7acbd-f9b2-420f-97c7-3713726d8a57','34484400-c657-4b15-be95-7746a2370aa0',
                'e073baa8-6b83-48e1-8db5-3dc9c93eb52f')
     AND rendered_html ~ '<script' AND length(rendered_html) > 3000;
  IF n <> 3 THEN RAISE EXCEPTION '625: only % of 3 tool artefacts still carry a <script> and real length — the tool may have been damaged', n; END IF;

  RAISE NOTICE '625 OK: 3 owned tool artefacts patched in place (no regeneration, no policy flip), 2 templates repointed, scope guard clean, all 3 tools still carry their script. Verify over HTTP.';
END $$;

COMMIT;
