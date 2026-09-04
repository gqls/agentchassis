-- REPAIR_479 — re-attach the component reference to orphaned tool slots.
--
-- WHAT IT REPAIRS: page_components rows holding a working tool's rendered_html with
-- component_id NULL, written by Layer 2's re-append arm (bugs_open/479). The bytes are
-- CORRECT and are already serving; only the reference is missing. No re-render is needed
-- or wanted — see RUNBOOK §8b, do not verify with a re-render.
--
-- HOW A ROW IS IDENTIFIED, and why one signal is not enough. Two independent conditions
-- must BOTH hold, and the query below is the whole rule:
--   (1) an ACTIVE content_components row whose `function` equals the slot name;
--   (2) that component is not already bound to a page on a DIFFERENT site.
-- (1) alone is unsound: [MEASURED 2026-09-04] 3 of the 5 tool orphans have 2-3 active
-- same-function components, because tools are FORKED across sites under compounded names
-- (`…-advertise-co-uk-websitepromotion-co-uk`). (2) is what names the owner.
--
-- ⚠ A BYTE-EXACT RENDER CANNOT NAME THE OWNER, and it was tried. Rendering each candidate
-- VERSION with the instance token (`c-` || slot_name) and comparing md5 to the stored bytes
-- reproduces them EXACTLY — for TWO different components at once, because a fork is a
-- literal copy of the template. It proves the template text and says nothing about which
-- component the page holds. Kept here so nobody re-walks it: for
-- advertise.co.uk/tool-ad-budget-calculator, versions ae5b412f (advertise) AND 205cc5a8 /
-- bc53c2b2 (the websitepromotion fork) all md5-match the stored 16,962 bytes.
--
-- ⚠ THE ASSERTION TRAP THIS AVOIDS (paid for by the portfolio_positioning lane, 2026-09-03):
-- do NOT assert digest integrity across a POPULATION. `rendered_html_digest IS DISTINCT FROM
-- md5(rendered_html)` convicts a NULL digest, and a NULL digest is a normal state — 206 of
-- 3,220 rows fleet-wide. Compare the rows you actually TOUCH, before against after.
--
-- ⚠ And a verify block of bare SELECTs cannot stop a COMMIT (ON_ERROR_STOP ignores a
-- non-empty result). Every check below is DO/RAISE.
--
-- REHEARSE FIRST: run with the final line as ROLLBACK, read the NOTICEs, then re-run with
-- COMMIT. Rehearsed 2026-09-04 — see the lane NOTES for the output.

BEGIN;

CREATE TEMP TABLE repair_479_map ON COMMIT DROP AS
WITH orphan AS (
  SELECT pc.id AS orphan_id, pc.slot_name, pc.rendered_html, p.site_id, s.domain, p.name AS page
    FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
    JOIN sites s ON s.id = p.site_id
   WHERE pc.component_id IS NULL
     AND pc.build_status <> 'removed'
     AND pc.slot_name LIKE 'tool-%'
), cand AS (
  SELECT o.*, cc.id AS cid, cc.name AS cname
    FROM orphan o
    JOIN content_components cc
      ON cc.function = o.slot_name AND cc.is_active
   WHERE NOT EXISTS (
           SELECT 1 FROM page_components x
             JOIN pages xp ON xp.id = x.page_id
            WHERE x.component_id = cc.id
              AND x.build_status <> 'removed'
              AND xp.site_id <> o.site_id)
)
SELECT orphan_id, domain, page, slot_name, cid, cname,
       md5(rendered_html) AS before_md5,
       length(rendered_html) AS before_len,
       count(*) OVER (PARTITION BY orphan_id) AS n_candidates
  FROM cand;

-- GUARD 1: refuse the whole repair if ANY row is ambiguous. Binding an arbitrary fork is
-- precisely the silent substitution this bug is about; refusing is the safe direction.
DO $$
DECLARE amb int; tot int;
BEGIN
  SELECT count(*) FILTER (WHERE n_candidates <> 1), count(DISTINCT orphan_id)
    INTO amb, tot FROM repair_479_map;
  IF amb > 0 THEN
    RAISE EXCEPTION 'REFUSED: % row(s) resolve to more than one free candidate — identify them by hand, do not guess', amb;
  END IF;
  IF tot = 0 THEN
    RAISE EXCEPTION 'REFUSED: nothing to repair — either already done, or the rule no longer matches (re-read bugs_open/479 §6)';
  END IF;
  RAISE NOTICE 'repair_479: % row(s) resolved, each to exactly one candidate', tot;
END $$;

UPDATE page_components pc
   SET component_id = m.cid
  FROM repair_479_map m
 WHERE pc.id = m.orphan_id
   AND pc.component_id IS NULL;          -- re-assert the precondition at write time

-- GUARD 2: the bytes must be untouched, compared ROW BY ROW against the snapshot — never
-- as a population property.
DO $$
DECLARE changed int; unbound int;
BEGIN
  SELECT count(*) INTO changed
    FROM repair_479_map m JOIN page_components pc ON pc.id = m.orphan_id
   WHERE md5(pc.rendered_html) <> m.before_md5
      OR length(pc.rendered_html) <> m.before_len;
  IF changed > 0 THEN
    RAISE EXCEPTION 'ABORT: % row(s) had their bytes altered — this repair must touch component_id only', changed;
  END IF;

  SELECT count(*) INTO unbound
    FROM repair_479_map m JOIN page_components pc ON pc.id = m.orphan_id
   WHERE pc.component_id IS DISTINCT FROM m.cid;
  IF unbound > 0 THEN
    RAISE EXCEPTION 'ABORT: % row(s) did not take the intended component', unbound;
  END IF;

  RAISE NOTICE 'repair_479: bytes identical on every touched row; every row now bound as intended';
END $$;

SELECT domain, page, slot_name, cname AS bound_to, before_len AS bytes FROM repair_479_map ORDER BY 1,2;

ROLLBACK;   -- ⚠ REHEARSAL. Change to COMMIT to apply.
