-- 469_departments_grid_and_leadership_team_consume_site_tokens.sql
--
-- bugs_open/122 family, "family B" — the 24 remaining firm contrast failures on
-- ai-agent-orchestration.com. Two components have NO THEME SUPPORT AT ALL: their
-- entire colour surface is six hardcoded literals, in a library where the sibling
-- section component (differentiators-section) already consumes site tokens
-- correctly. On the only DARK site that places them they were never going to work.
--
-- ⚠ NUMBERING. The lane handoff of 2026-08-18 calls this "migration 459". That
-- number was taken by another lane (459_zip_deliverer_agent_HOLD.sql) before this
-- was written. This is the same change under the next free number.
--
-- THE MECHANISM, measured at the artefact (render_audit.py, 2026-08-18 ~19:14Z,
-- overImage excluded — 32 firm failures on 4 pages, of which these are 24):
--
--   .team-section { background: #f8f9fa; }   <- unthemed GROUND
--   .team-member  { background: #fff; }      <- unthemed GROUND
--   .team-section h2, .team-member h3        <- set NO colour, so they INHERIT
--                                               --color-text (#E6EDF3 here)
--
-- So the failure is not a bad foreground. It is a THEMED foreground inherited onto
-- an UNTHEMED ground: #E6EDF3 on #FFFFFF measures 1.18:1, and #E6EDF3 on #F8F9FA
-- measures 1.12:1. Every one of the 24 is one of those two pairs.
--
-- BREAKDOWN (firm, by placement):
--   index / departments-grid   9   (1 H2 + 8 department H3s)
--   about / departments-grid   9   (1 H2 + 8 department H3s)
--   about / leadership-team    6   (1 H2 + 1 stray P + 4 member H3s)
--
-- ⚠ WHY THE WHOLE BLOCK MOVES TOGETHER, AND WHY "JUST TOKENISE THE BACKGROUNDS"
-- IS WRONG. Repointing only the two grounds would leave `color: #555` (.member-bio,
-- .section-intro) and `color: #0f3460` (.member-title) painted onto a #0D1117 card
-- — a fresh set of invisible text, which is exactly the defect migration 456
-- introduced by repointing foregrounds without regard to their ground. Six
-- declarations, one edit.
--
-- ⚠ BOTH COMPONENTS MUST MOVE IN THE SAME MIGRATION, and this is not tidiness.
-- They define the SAME class names (.team-section, .team-member, .member-*), and
-- on about.html BOTH are placed. Their <style> blocks land on one page, so the
-- later block wins for both sections. Migrating one alone would restyle the other
-- through the cascade. Verified fleet-wide: exactly 2 content_components mention
-- these classes, so nothing outside this pair can be caught by the cascade either.
--
-- THE MAPPING. Every token verified live by getComputedStyle on each of the three
-- affected sites (never read from a stylesheet — this site's served stylesheet
-- states `h3 { color:#ffffff }` and that is NOT the winning declaration):
--
--   background: #f8f9fa  ->  var(--color-background, #f8f9fa)
--   background: #fff     ->  var(--color-surface,    #fff)
--   background: #e0e0e0  ->  var(--color-border,     #e0e0e0)
--   color:      #0f3460  ->  var(--color-text,       #0f3460)
--   color:      #555     ->  var(--color-text-muted, #555)
--
-- All five tokens are SET on all three sites, so no branch falls back in practice;
-- the fallbacks are there for a site that emits fewer tokens, where each degrades
-- to today's exact colour rather than dropping the declaration.
--
-- BLAST RADIUS — 3 sites, 7 placements, ENUMERATED not asserted:
--
--   ai-agent-orchestration.com  DARK   departments-grid x2, leadership-team x1
--   finetuning.uk               light  departments-grid x2, leadership-team x1
--   leopardessconsulting.co.uk  light  leadership-team  x1
--
-- SIMULATED AFTER-STATE, per site, every element in the block (ratio before ->
-- after, against the WCAG level that element needs):
--
--   ai-agent-orchestration.com (dark)
--     h2 on section ground      1.12 -> 16.68  (need 3.0)   FIXED
--     h3 on member card         1.18 -> 16.02  (need 4.5)   FIXED
--     icon stroke on icon well  1.12 -> 12.88  (need 3.0)   FIXED
--     .section-intro            7.07 ->  6.41  (need 4.5)   passes both ways
--     .member-title            12.50 -> 16.02  (need 4.5)   passes both ways
--     .member-bio               7.46 ->  6.15  (need 4.5)   passes both ways
--   finetuning.uk (light)      every row passes before AND after; lowest after is
--                              .section-intro at 5.02 (need 4.5)
--   leopardessconsulting (light) every row passes before AND after; lowest after is
--                              .section-intro at 6.79 (need 4.5)
--
-- ⚠ THE LIGHT SITES ARE NOT "UNCHANGED", and an earlier statement of this plan said
-- they were. Only `.team-member`'s ground is genuinely identical (#fff -> #FFFFFF
-- on both). `.team-section`'s ground moves #f8f9fa -> #F5F3EF / #FAF8F4, the icon
-- well moves #e0e0e0 -> #D4CFC6 / #E4DFD5, and both text literals move to the
-- sites' own text tokens. Those are the intended effect — each site gets its own
-- answer instead of one imposed on all three — but they ARE changes, and the claim
-- to make is "no element loses its contrast floor on any site", which is the table
-- above, not "nothing moves".
--
-- ⚠ NO NEW COLOUR PAIR IS INTRODUCED THAT WAS NOT ALREADY IN THE PALETTE. Verify
-- after propagation BY COLOUR PAIR: any (fg,bg) in the after-set that is absent
-- from the before-set is a regression this migration introduced. That is the check
-- 456 did not have, and it is why 456's regression survived a net improvement.
--
-- OUT OF SCOPE, deliberately, both found while writing this:
--  1. `box-shadow: 0 2px 8px rgba(0,0,0,0.06)` is a black shadow, invisible on a
--     dark card. Cosmetic, not a contrast failure, and untouched.
--  2. about/leadership-team renders `<p class="section-intro"><p>...</p></p>` —
--     content that already carries its own <p> is wrapped in another. The parser
--     auto-closes the outer tag, so the inner bare <p> is a SIBLING with no class
--     and inherits --color-text. That stray P is one of the 24 and this migration
--     does fix its contrast (it sits on .team-section), but the double-wrap itself
--     is a separate content/render defect and is NOT addressed here.
--
-- DOES NOT RE-RENDER. Placements keep the OLD html until re-rendered, exactly as
-- with 456 and 457. Propagate afterwards, then re-audit the live pages.
--
-- ROLLBACK: 469_departments_grid_and_leadership_team_consume_site_tokens_ROLLBACK.sql
--   (restores the byte-exact pre-469 html_template from migration_backups rows
--    written below — not a reverse-regex, so it cannot drift.)

BEGIN;

-- 1. Byte-exact backup of the two rows being changed, so rollback restores the
--    original rather than reconstructing it.
INSERT INTO migration_backups (migration_name, target_table, target_id, old_value, notes)
SELECT '469_departments_grid_and_leadership_team_consume_site_tokens',
       'content_components', cc.id::text,
       jsonb_build_object('html_template', cc.html_template),
       'pre-469 html_template for ' || cc.name
FROM content_components cc
WHERE cc.id IN ('be82aac8-c416-443a-a63f-0c58724a5b6b',
                'c5af72e8-73ff-4dfe-bf88-54f7fa3978e1');

-- 2. The edit. Plain literal replacement, not regex: each anchor is a complete
--    declaration that occurs exactly once (twice for `color: #555`) in each row,
--    confirmed by dry run against these same rows. Addressed BY ID rather than by
--    name so a concurrently-created same-named component cannot be swept in.
UPDATE content_components
SET html_template =
      replace(replace(replace(replace(replace(
        html_template,
        'background: #f8f9fa', 'background: var(--color-background, #f8f9fa)'),
        'background: #fff',    'background: var(--color-surface, #fff)'),
        'background: #e0e0e0', 'background: var(--color-border, #e0e0e0)'),
        'color: #0f3460',      'color: var(--color-text, #0f3460)'),
        'color: #555',         'color: var(--color-text-muted, #555)'),
    updated_at = now()
WHERE id IN ('be82aac8-c416-443a-a63f-0c58724a5b6b',
             'c5af72e8-73ff-4dfe-bf88-54f7fa3978e1');

-- 3. Guard the exact post-condition. A verify block of bare SELECTs cannot stop a
--    COMMIT (ON_ERROR_STOP ignores a non-empty result), so this must RAISE.
DO $$
DECLARE
  backed_up   int;
  targets     int;
  tokenised   int;
  bare_left   int;
  bare_var    int;
  bare_ink    int;
  class_owners int;
  r           record;
BEGIN
  SELECT count(*) INTO backed_up FROM migration_backups
   WHERE migration_name = '469_departments_grid_and_leadership_team_consume_site_tokens';
  IF backed_up <> 2 THEN
    RAISE EXCEPTION '469: expected 2 backup rows, wrote % — refusing to proceed without a byte-exact restore path', backed_up;
  END IF;

  -- The two ids must still be the two components we measured.
  SELECT count(*) INTO targets FROM content_components
   WHERE id IN ('be82aac8-c416-443a-a63f-0c58724a5b6b','c5af72e8-73ff-4dfe-bf88-54f7fa3978e1')
     AND name IN ('departments-grid','leadership-team');
  IF targets <> 2 THEN
    RAISE EXCEPTION '469: expected 2 target rows named departments-grid/leadership-team, found % — the ids have drifted, ABORTING rather than half-applying', targets;
  END IF;

  -- Each row must now carry all five tokens (6 references: --color-text-muted twice).
  FOR r IN SELECT name, html_template FROM content_components
            WHERE id IN ('be82aac8-c416-443a-a63f-0c58724a5b6b','c5af72e8-73ff-4dfe-bf88-54f7fa3978e1') LOOP
    IF r.html_template !~ 'background:\s*var\(--color-background,' THEN
      RAISE EXCEPTION '469: % is missing the --color-background ground', r.name;
    END IF;
    IF r.html_template !~ 'background:\s*var\(--color-surface,' THEN
      RAISE EXCEPTION '469: % is missing the --color-surface ground', r.name;
    END IF;
    IF r.html_template !~ 'background:\s*var\(--color-border,' THEN
      RAISE EXCEPTION '469: % is missing the --color-border ground', r.name;
    END IF;
    IF r.html_template !~ 'color:\s*var\(--color-text,' THEN
      RAISE EXCEPTION '469: % is missing the --color-text foreground', r.name;
    END IF;
    IF (SELECT count(*) FROM regexp_matches(r.html_template,'color:\s*var\(--color-text-muted,','g')) <> 2 THEN
      RAISE EXCEPTION '469: % must carry exactly 2 --color-text-muted foregrounds (.section-intro and .member-bio)', r.name;
    END IF;
  END LOOP;

  SELECT count(*) INTO tokenised FROM content_components
   WHERE id IN ('be82aac8-c416-443a-a63f-0c58724a5b6b','c5af72e8-73ff-4dfe-bf88-54f7fa3978e1')
     AND (SELECT count(*) FROM regexp_matches(html_template,'var\(--color-','g')) = 6;
  IF tokenised <> 2 THEN
    RAISE EXCEPTION '469: expected both rows to carry exactly 6 var(--color-*) references, % qualify', tokenised;
  END IF;

  -- THE LOAD-BEARING ONE: no bare colour literal may remain in either row. This is
  -- what catches "tokenised the backgrounds and left the text behind", i.e. 456's
  -- mistake in its other direction.
  SELECT count(*) INTO bare_left FROM content_components
   WHERE id IN ('be82aac8-c416-443a-a63f-0c58724a5b6b','c5af72e8-73ff-4dfe-bf88-54f7fa3978e1')
     AND html_template ~ '(background|color):\s*#[0-9a-fA-F]{3,8}';
  IF bare_left <> 0 THEN
    RAISE EXCEPTION '469: % row(s) still carry a bare hex colour declaration; the whole block must move together or the survivors land on a re-themed ground', bare_left;
  END IF;

  -- Never introduce a BARE var() reference IN THE ROWS THIS MIGRATION TOUCHES. A
  -- disabled/absent token emits nothing and a bare var() drops the whole
  -- declaration, which fails silently and looks like a styling choice.
  -- ⚠ SCOPED TO THE TWO ROWS ON PURPOSE. 145 content_components fleet-wide already
  -- carry a bare var(--color-*) (measured 2026-08-18, before this migration), so a
  -- fleet-wide assertion here would abort on pre-existing state that has nothing to
  -- do with this change. That is a real backlog, and it is not this file's to fix.
  SELECT count(*) INTO bare_var FROM content_components
   WHERE id IN ('be82aac8-c416-443a-a63f-0c58724a5b6b','c5af72e8-73ff-4dfe-bf88-54f7fa3978e1')
     AND html_template ~ 'var\(\s*--color-[a-z-]+\s*\)';
  IF bare_var <> 0 THEN
    RAISE EXCEPTION '469: % touched row(s) carry a BARE var(--color-*) reference; the two-level fallback is mandatory', bare_var;
  END IF;

  -- The fleet-wide invariant 456/457 established and that IS currently held: no
  -- bare ink reference anywhere. Kept because this file is in that family.
  SELECT count(*) INTO bare_ink FROM content_components
   WHERE html_template ~ 'var\(\s*--color-(primary|accent)-ink\s*\)';
  IF bare_ink <> 0 THEN
    RAISE EXCEPTION '469: % content_component(s) carry a BARE ink reference', bare_ink;
  END IF;

  -- The cascade assumption this migration rests on: these class names are owned by
  -- exactly these two components fleet-wide. If a third appears, the "both move
  -- together" reasoning above no longer covers the page.
  SELECT count(*) INTO class_owners FROM content_components
   WHERE html_template ~ '\.(team-section|team-member|member-icon|member-photo|member-title|member-bio)';
  IF class_owners <> 2 THEN
    RAISE EXCEPTION '469: % components define the .team-*/.member-* classes, expected 2 — a third owner changes the cascade and this migration no longer reasons about the whole page', class_owners;
  END IF;

  RAISE NOTICE '469 OK: departments-grid + leadership-team now consume site tokens (5 tokens, 6 declarations, two-level fallbacks). Placements keep the OLD html until re-rendered — propagate, then re-audit BY COLOUR PAIR.';
END $$;

COMMIT;
