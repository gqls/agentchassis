-- 232_fix_231_cascading_replace_and_missed_variants.sql
--
-- CORRECTS A DEFECT I INTRODUCED IN 231, ~20 minutes earlier, same session.
-- Kept as its own file rather than folded into 231 so the mistake stays on the
-- record: 231 was applied to production and this is what it did wrong.
--
-- ══ DEFECT 1 — A CASCADING replace() CHAIN ═════════════════════════════════
-- 231 normalised the agent figure with nested replace() calls:
--
--   replace(replace(replace(X, '70+ Agents','170+ Agents'),
--                              '70+ agents','170+ agents'),
--                              '70+ agent', '170+ agent')
--
-- The third call reads the SECOND one's OUTPUT. "170+ agents" CONTAINS the
-- substring "70+ agent" at offset 1, so it was rewritten again:
--
--   "70+ agents"  ->  "170+ agents"  ->  "1170+ agents"     <-- shipped
--
-- Six live sentences were left claiming **1170+ agents** — a figure ~6.7x the
-- real fleet, produced by a fix for understating it. The general rule:
-- **a chain of replace() calls where one pattern is a substring of another's
-- REPLACEMENT is self-feeding.** Anchor with a word boundary, or do one pass.
--
-- ══ DEFECT 2 — TWO VARIANTS NEVER MATCHED ══════════════════════════════════
-- 231's literals were taken from a LIKE census that I wrote from the variants I
-- had already seen, so it could only find what I had already thought of:
--
--   "70 Agents in Production. Yours Could Be Next."   (no "+")
--   "more than 30 distinct agent types"               ("distinct" infix)
--   "30+ distinct agent types"                        ("distinct" infix)
--
-- ══ HOW BOTH WERE CAUGHT — AND WHY 231's OWN POST-CONDITIONS DID NOT ═══════
-- 231 asserted `content_data !~ '(^|[^1])70\+\s*[Aa]gent'`. That predicate is
-- blind to BOTH defects by construction: `[^1]` excuses "1170+", and it only
-- looks for the exact shapes 231 already knew about. **A post-condition written
-- from the same mental model as the change cannot falsify the change.**
--
-- What caught them was running the REAL scanner (ParseEvidenceBase +
-- ScanBannedClaims) over the stored content with the site's own new register
-- and asking a question the migration could not answer for itself:
--
--   does the CORRECTED copy trip the site's OWN bans?   -> 4 hits.
--
-- That check is now the post-condition below, expressed in SQL as the same
-- regexes the Go patterns use.

BEGIN;

DO $fix$
DECLARE
    v_site uuid;
    n int;
BEGIN
    SELECT id INTO v_site FROM sites WHERE domain = 'ai-agent-orchestration.com';
    IF v_site IS NULL THEN
        RAISE EXCEPTION '232: no site row for ai-agent-orchestration.com';
    END IF;

    -- One ordered pass of ANCHORED regexes. `\m` is Postgres's word-start, so
    -- "70" inside "170"/"1170" can never match and no step can feed the next.
    -- (Postgres syntax is correct HERE. The banned_claims PATTERNS are compiled
    -- by Go and must use \b — see 231's note.)
    WITH upd AS (
        UPDATE page_components pc
        SET content_data =
            regexp_replace(
            regexp_replace(
            regexp_replace(
            regexp_replace(
            regexp_replace(
              regexp_replace(pc.content_data::text, '\m1170\s*\+?', '170+', 'g'),
              '\mover 70 specialised AI agents', '170+ agents', 'gi'),
              '\m70\s*\+?\s*Agents', '170+ Agents', 'g'),
              '\m70\s*\+?\s*agents', '170+ agents', 'g'),
              '\m70\s*\+?\s*agent\M', '170+ agent', 'g'),
              '\m(30|thirty)\s*\+?\s*(distinct )?agent types', '170+ agent types', 'gi')::jsonb,
            updated_at = now()
        FROM pages p
        WHERE p.id = pc.page_id AND p.site_id = v_site
          AND (pc.content_data::text ~ '1170'
            OR pc.content_data::text ~* '\m(70|seventy)\s*\+?\s*(specialised |specialized )?(ai )?agents?\M'
            OR pc.content_data::text ~* '\m(30|thirty)\s*\+?\s*(distinct )?agent types')
        RETURNING 1
    )
    SELECT count(*) INTO n FROM upd;
    RAISE NOTICE '232: repaired % components', n;
END $fix$;

-- ── Post-conditions: the check 231 should have had ─────────────────────────
-- These are the SAME regexes as the seeded banned_claims, so this asserts
-- exactly "the corrected copy does not trip the site's own bans".
DO $post$
DECLARE
    v_cascade int; v_70 int; v_30 int; v_conc int;
BEGIN
    SELECT
      count(*) FILTER (WHERE pc.content_data::text ~ '1170'),
      count(*) FILTER (WHERE pc.content_data::text ~* '\m(70|seventy)\s*\+?\s*(specialised |specialized )?(ai )?agents?\M'),
      count(*) FILTER (WHERE pc.content_data::text ~* '\m(30|thirty)\s*\+?\s*(distinct )?agent types'),
      count(*) FILTER (WHERE pc.content_data::text ~* 'thousands of concurrent')
      INTO v_cascade, v_70, v_30, v_conc
    FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
    WHERE s.domain='ai-agent-orchestration.com';

    IF v_cascade <> 0 THEN RAISE EXCEPTION '232: % component(s) still carry the 1170+ cascade artefact', v_cascade; END IF;
    IF v_70 <> 0 THEN RAISE EXCEPTION '232: % component(s) still match the banned 70-agents pattern', v_70; END IF;
    IF v_30 <> 0 THEN RAISE EXCEPTION '232: % component(s) still match the banned 30-agent-types pattern', v_30; END IF;
    IF v_conc <> 0 THEN RAISE EXCEPTION '232: % component(s) still carry the forbidden concurrency claim', v_conc; END IF;

    RAISE NOTICE '232 OK: stored content trips ZERO of the site''s own banned_claims';
END $post$;

COMMIT;
