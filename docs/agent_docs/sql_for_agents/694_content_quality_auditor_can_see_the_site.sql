-- 694 — content-quality-auditor: give the seat eyes before anyone routes it
--
-- WHY THIS EXISTS. The owner asked (2026-09-02) that content-quality-auditor be
-- put into the NEW BUILD path. The experience_loop handoff of the same date sets
-- an explicit gate before that wiring: read what a run actually produces, and if
-- it does not name what the owner complained about on boxingonline, "the job is
-- bigger and the prompt is the work — say so plainly rather than shipping a route
-- to a seat that will stay quiet."
--
-- THE GATE FAILS, and this migration is the fix that has to land first. All
-- figures [MEASURED 2026-09-02] against live config and the live DB.
--
--  1. SIGHT. `load_page_content` hardcodes
--       WHERE p.site_id = $1 AND p.name IN ('index','about','services','contact')
--     Four page names. On boxingonline (site d2aa5206) that is 3 of 22 pages;
--     fleet-wide it is 92 of 1,196 pages across 36 sites — 7.7%, averaging 2.56
--     pages seen per site. 'services' exists on only 7 of 36 sites, so a quarter
--     of the allow-list usually matches nothing.
--     Every page the owner complained about — the four /guides/tool-*-guide.html
--     explainers, the /articles/index.html manifesto, the /tools/fighter-comparator/
--     form — is OUTSIDE those four names. No prompt wording can review a page that
--     was never placed in the model's context, which is why this is a query change
--     and not a prompt change.
--
--  2. DEPTH. LEFT(..., 1000) per page. Index pages average 28,180 chars
--     fleet-wide, so the landing page was sampled at 4.5% (about 11.0%,
--     contact 14.2%, services 13.6%).
--
--  3. CSS. rendered_html carries <style> blocks: 42.8% of rendered_html
--     fleet-wide over 2,851 components. On boxingonline's index the <style> block
--     starts at CHARACTER 1 and does not close inside the window, so 999 of the
--     1,000 characters reaching the model were stylesheet. about and contact lost
--     426 and 417 chars to the same shape. Stripping style/script BEFORE the
--     truncation (not after) is load-bearing: strip-after cannot match an unclosed
--     block, which is also how a naive "prose chars" metric scores the worst page
--     best — see the NOTES misstep for 2026-09-02.
--     ⚠ ROUND 1 CORRECTION (council REVISE, editquality HIGH + bug_historian MEDIUM,
--     both right). This file's first cut wrote the strip as
--     `regexp_replace(..., '<style[^>]*>.*?</style>', '', 'gs')`. PostgreSQL takes
--     the greediness of the FIRST quantifier, so `[^>]*` makes the WHOLE expression
--     greedy and `.*?` does not save it: on a component with two style blocks it
--     matches from the first `<style` to the LAST `</style>` and DELETES the prose
--     between them — the exact content-loss defect this file exists to prevent,
--     reintroduced as a silent one. Demonstrated:
--       regexp_replace('<style>.a{}</style>KEEP THIS PROSE<style>.b{}</style>AND THIS',
--                      '<style[^>]*>.*?</style>','','gs')  ->  'AND THIS'
--     [MEASURED 2026-09-02] 7 of 2,871 components carry 2+ style blocks, and the
--     greedy form destroys content on ALL SEVEN — avg 2,528 chars, worst 9,076.
--     Now uses the pattern already proven for this exact purpose in migration 601
--     (`claims-auditor.load_page_text`): `<style[^>]*?>.*?</style>` with `'gi'`, the
--     lazy FIRST quantifier being what actually makes the expression non-greedy.
--     The replacement is a SPACE, not '': removing a block with '' joins the words
--     either side of it ('ALPHA'+'BETA' -> 'ALPHABETA'), which 601 also gets right
--     and this file's first cut did not. Stripping stays PER COMPONENT inside the
--     aggregate, which is migration 518's rule (a regex must never see two
--     components at once) and which the first cut already satisfied.
--
--  4. ORDER. string_agg(pc.rendered_html, ' ') carried NO ORDER BY, so which
--     1,000 chars were sampled drifted with physical row order. Observed, not
--     theoretical: the 12:35 run's stored page_samples contains CTA prose that
--     sits at char 15,154 of 18,553 — outside any window taken later that day —
--     while all three index components have updated_at = created_at = 06:37.
--     Same rows, same bytes, different leading component. pc.position is NOT NULL
--     and is the intended order.
--
--  5. ENUM. The prompt's fifth REVIEW dimension is AUDIENCE; the declared category
--     enum offers 'content' instead and has no 'audience'. Across all stored
--     audits, 210 findings: gap 64, content 45, differentiation 45, cta 43,
--     tone 3, and audience 10 — the last being outside the declared enum entirely,
--     i.e. the model improvising a category no consumer knows. This adds the value
--     the prompt already asks for rather than removing the dimension.
--
-- WHAT THIS DOES NOT CHANGE, deliberately:
--   * filing_mode stays 'record'. The owner ruled on 2026-09-02, asked directly:
--     findings are recorded for human approval, nothing auto-regenerates. That
--     keeps the 2026-08-25 ruling ("switch off the evolutionary aspect … causing
--     too many bad / unexpected renders") intact. The human reader already exists
--     and is wired end to end — admin dashboard `recordVerdictsOnly` filter
--     (frontends/admin-dashboard/src/App.tsx:474 -> GET /admin/work-items?filing_mode=record)
--     and the per-finding Release button (App.tsx:737 -> POST
--     /admin/work-items/:item_id/release -> HandleReleaseRecordVerdict,
--     internal/core-manager/api/server.go:455). Nothing new is needed for it.
--   * No routing into site-work-orchestrator. That is the NEXT change, and it is
--     deliberately not bundled here: routing a blind seat is the defect this file
--     exists to prevent, so sight lands and is observed first.
--   * No new agent, action, contract or shared key. This edits one seat's own
--     query and prompt.
--
-- COST. Bounded and measured, not estimated: 4 pages per page_type (17 types
-- exist; a site has 5.7 on average, 9 at most) at 1,200 chars each gives, across
-- 35 sites with renderable components, avg 14.1 pages / 16,884 chars and worst
-- case 24 pages / 28,800 chars ~= 7,200 input tokens. Current live average input
-- is 1,744 tokens, so this is roughly a 2.4x input increase on a seat whose calls
-- cost fractions of a penny. The query runs in well under the 300s step timeout
-- on the largest site in the estate (155 pages): 18 pages sampled, 21,600 chars.
--
-- NOT a _HOLD: no Go change accompanies this. The query, the prompt and the enum
-- are all read from default_config by the already-deployed chassis, so this is
-- live on apply with no image-roll ordering constraint.
--
-- Council-Submitted: <fill in after 097_TRIGGER>

SELECT snapshot_agent('content-quality-auditor', '694_content_quality_auditor_can_see_the_site.sql: pre-update');

BEGIN;

CREATE TABLE IF NOT EXISTS agent_definitions_bak_694 AS
SELECT id, type, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'content-quality-auditor' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    nrows int; q text; p text;
    c_enum int; c_aud int; c_top5 int;
BEGIN
    -- 640's council-raised lesson: SELECT INTO takes the FIRST of N rows
    -- silently, so a duplicate active row would half-apply this migration.
    SELECT count(*) INTO nrows FROM agent_definitions
    WHERE type = 'content-quality-auditor' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    IF nrows <> 1 THEN
        RAISE EXCEPTION '694: expected exactly 1 active content-quality-auditor row BEFORE writing, found %', nrows;
    END IF;

    SELECT default_config->'workflow'->'steps'->'load_page_content'->'config'->>'query',
           default_config->'workflow'->'steps'->'run_content_llm_audit'->'config'->>'prompt'
    INTO q, p
    FROM agent_definitions
    WHERE type = 'content-quality-auditor' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    -- NULL-SAFETY FIRST (council round 2, editquality MEDIUM — a real fails-open).
    -- `position(<literal> in NULL)` is NULL, `NULL = 0` is NULL, and plpgsql treats
    -- IF NULL as FALSE — so every anchor guard below would silently NOT raise if a
    -- jsonb path resolved to NULL (step renamed, key missing, config drifted), and
    -- the migration would proceed to replace() over a NULL prompt. Demonstrated in
    -- psql: the bare form reports `fired = f` on a NULL input. Refuse explicitly.
    IF q IS NULL OR p IS NULL THEN
        RAISE EXCEPTION '694: load_page_content.query or run_content_llm_audit.prompt resolved to NULL on the live row (q IS NULL: %, p IS NULL: %) — the step names or config keys have moved; re-derive this seed from the live row', (q IS NULL), (p IS NULL);
    END IF;

    -- Anchor on the exact defects this file is written against. If any is
    -- already gone, the live row is not the row these measurements describe.
    IF position($a$p.name IN ('index', 'about', 'services', 'contact')$a$ in q) = 0 THEN
        RAISE EXCEPTION '694: the four-name allow-list is not in the live load_page_content query — re-derive this seed from the live row';
    END IF;
    IF position($a$LEFT(string_agg(pc.rendered_html, ' '), 1000)$a$ in q) = 0 THEN
        RAISE EXCEPTION '694: the unordered 1000-char sample is not in the live query — re-derive this seed from the live row';
    END IF;
    IF position($a$"tone|gap|cta|differentiation|content"$a$ in p) = 0 THEN
        RAISE EXCEPTION '694: the live prompt does not carry the expected category enum — re-derive this seed from the live row';
    END IF;
    IF position($a$5. AUDIENCE:$a$ in p) = 0 THEN
        RAISE EXCEPTION '694: the live prompt does not carry the AUDIENCE review dimension — re-derive this seed from the live row';
    END IF;

    -- Occurrence COUNTS, not mere presence (council round 1, debug_historian).
    -- replace() rewrites EVERY occurrence, so a needle that appears twice would
    -- be edited twice and the presence checks above could not tell. Derive the
    -- count mechanically rather than asserting a remembered number.
    c_enum := (length(p) - length(replace(p, $a$"category":"tone|gap|cta|differentiation|content"$a$, ''))) /
              length($a$"category":"tone|gap|cta|differentiation|content"$a$);
    c_aud  := (length(p) - length(replace(p, $a$5. AUDIENCE:$a$, ''))) / length($a$5. AUDIENCE:$a$);
    c_top5 := (length(p) - length(replace(p, $a$IMPORTANT: Report ONLY the TOP 5 most impactful content issues.$a$, ''))) /
              length($a$IMPORTANT: Report ONLY the TOP 5 most impactful content issues.$a$);
    IF c_enum <> 1 OR c_aud <> 1 OR c_top5 <> 1 THEN
        RAISE EXCEPTION '694: expected each prompt needle exactly once (enum=%, audience=%, top5=%) — the live prompt has drifted, re-derive this seed', c_enum, c_aud, c_top5;
    END IF;
END $$;

-- (1) SIGHT + DEPTH + CSS + ORDER: replace the query outright.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,load_page_content,config,query}',
        to_jsonb($q$WITH cleaned AS (
  SELECT p.id, p.name, p.page_type, p.url,
         string_agg(
           regexp_replace(
             regexp_replace(pc.rendered_html, '<style[^>]*?>.*?</style>', ' ', 'gi'),
             '<script[^>]*?>.*?</script>', ' ', 'gi'),
           ' ' ORDER BY pc.position) AS body
  FROM pages p
  JOIN page_components pc ON pc.page_id = p.id
  WHERE p.site_id = $1
    AND pc.rendered_html IS NOT NULL
    AND pc.rendered_html <> ''
    AND pc.locked_at IS NULL
  GROUP BY p.id, p.name, p.page_type, p.url
), ranked AS (
  SELECT *, ROW_NUMBER() OVER (
             PARTITION BY page_type
             ORDER BY (page_type = 'landing') DESC, length(body) DESC, name) AS rn
  FROM cleaned
)
SELECT name, page_type, url, LEFT(body, 1200) AS content_sample
FROM ranked
WHERE rn <= 4
ORDER BY (page_type = 'landing') DESC, page_type, name$q$::text)
    ),
    updated_at = NOW()
WHERE type = 'content-quality-auditor' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- (2) ENUM + the promise-keeping questions the owner actually asked for.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,run_content_llm_audit,config,prompt}',
        to_jsonb(
          replace(replace(replace(
            default_config->'workflow'->'steps'->'run_content_llm_audit'->'config'->>'prompt',
            $a1$IMPORTANT: Report ONLY the TOP 5 most impactful content issues.$a1$,
            $r1$Each page sample below carries its name, its page_type and its url, and you are now shown a sample of EVERY kind of page on the site (up to four per page_type), not just the home page. Use page_type: an "index" or "-index" page is a page whose job is to LIST things, a "tool" page is a page whose job is to DO something for the reader, and a "guide" page only explains a tool that exists elsewhere.

IMPORTANT: Report ONLY the TOP 8 most impactful content issues.$r1$),
            $a2$5. AUDIENCE: Does the content speak to the target audience specifically?$a2$,
            $r2$5. AUDIENCE: Does the content speak to the target audience specifically?
6. PROMISE: Does a listing or index contain the CLASS of thing its own heading promises? A block headed "Latest news" or "Latest articles" whose items are guides, tool explainers or static pages is a broken promise even though every item is a valid page. Quote the heading and the item titles.
7. EMPTY INDEX: Does an index page actually list its own items, or does it write ABOUT itself instead? A section index that serves headed prose describing its editorial standards while listing zero items is a journey with no end, and it looks healthy because the page is well written.
8. TOOL DATA: For each "tool" page, separate what the READER must type in from what the SITE supplies. A tool that asks the reader for every value it needs — figures they would have to look up elsewhere first — is a form, not a tool, and does not keep the promise its own card makes.
9. GUIDE PROMINENCE: If a "guide" page explaining a tool is more prominent than the tool itself (listed higher, linked from the home page while the tool is not, or reachable when the tool is not), say so. The thing the reader can USE should lead the thing that EXPLAINS it.$r2$),
            $a3${"category":"tone|gap|cta|differentiation|content"$a3$,
            $r3${"category":"tone|gap|cta|differentiation|content|audience"$r3$)
        )
    ),
    updated_at = NOW()
WHERE type = 'content-quality-auditor' AND is_active
  AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE q text; p text; v_site uuid; v_rows int; v_types int;
BEGIN
    SELECT default_config->'workflow'->'steps'->'load_page_content'->'config'->>'query',
           default_config->'workflow'->'steps'->'run_content_llm_audit'->'config'->>'prompt'
    INTO q, p
    FROM agent_definitions
    WHERE type = 'content-quality-auditor' AND is_active
      AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF q IS NULL OR p IS NULL THEN
        RAISE EXCEPTION '694: verify failed — the query or prompt resolved to NULL after the update';
    END IF;

    -- The four-name allow-list must be GONE, not merely joined by new text.
    IF position($a$p.name IN ($a$ in q) > 0 THEN
        RAISE EXCEPTION '694: verify failed — a page-name allow-list survives in load_page_content';
    END IF;
    IF position('ORDER BY pc.position' in q) = 0 THEN
        RAISE EXCEPTION '694: verify failed — the aggregate is still unordered';
    END IF;
    IF position('<style[^>]*?>.*?</style>' in q) = 0
       OR position('<script[^>]*?>.*?</script>' in q) = 0 THEN
        RAISE EXCEPTION '694: verify failed — style/script blocks are not being stripped with the non-greedy 601 pattern';
    END IF;
    -- The greedy form is the council's HIGH objection and must never come back.
    IF position('<style[^>]*>' in q) > 0 THEN
        RAISE EXCEPTION '694: verify failed — the GREEDY strip pattern is present; PostgreSQL takes the greediness of the FIRST quantifier, so [^>]* makes the whole expression greedy and it eats prose between two style blocks';
    END IF;
    IF position('LEFT(body, 1200)' in q) = 0 OR position('p.page_type' in q) = 0 THEN
        RAISE EXCEPTION '694: verify failed — the sample budget or the page_type column is missing';
    END IF;
    IF position('differentiation|content|audience' in p) = 0 THEN
        RAISE EXCEPTION '694: verify failed — audience is not in the category enum';
    END IF;
    IF position('6. PROMISE:' in p) = 0 OR position('7. EMPTY INDEX:' in p) = 0
       OR position('8. TOOL DATA:' in p) = 0 OR position('9. GUIDE PROMINENCE:' in p) = 0 THEN
        RAISE EXCEPTION '694: verify failed — a promise-keeping review dimension is missing';
    END IF;
    IF position('TOP 5 most impactful' in p) > 0 THEN
        RAISE EXCEPTION '694: verify failed — the old TOP 5 cap survived';
    END IF;

    -- EXECUTE THE EMBEDDED SQL (council round 2, debug_historian MEDIUM).
    -- Substring checks prove the text was written, never that it PARSES or runs:
    -- SQL inside a step config is DATA to this migration and only becomes code
    -- when the chassis runs the step, which is far too late to find a typo. So
    -- run it here, against the real site with the most renderable components, and
    -- assert it returns rows across more than one page_type (which is the whole
    -- point of the change — the old query could only ever return the four names).
    SELECT p2.site_id INTO v_site
    FROM pages p2
    JOIN page_components pc2 ON pc2.page_id = p2.id
    WHERE pc2.rendered_html IS NOT NULL AND pc2.rendered_html <> '' AND pc2.locked_at IS NULL
    GROUP BY p2.site_id
    ORDER BY count(DISTINCT p2.page_type) DESC, count(*) DESC
    LIMIT 1;

    IF v_site IS NULL THEN
        RAISE EXCEPTION '694: verify failed — no site with renderable components to execute the new query against, so the execution check would be vacuous';
    END IF;

    EXECUTE 'SELECT count(*), count(DISTINCT page_type) FROM (' || q || ') AS s'
      INTO v_rows, v_types USING v_site;

    IF v_rows IS NULL OR v_rows = 0 THEN
        RAISE EXCEPTION '694: verify failed — the new load_page_content query parses but returned 0 rows for site % (it must sample something)', v_site;
    END IF;
    IF v_types < 2 THEN
        RAISE EXCEPTION '694: verify failed — the new query returned % row(s) spanning only % page_type(s) for site %; the change exists to sample MORE than one class of page', v_rows, v_types, v_site;
    END IF;
    RAISE NOTICE '694 verify: new query executed on site % -> % rows across % page_types', v_site, v_rows, v_types;
END $$;

COMMIT;

-- ROLLBACK: docs/agent_docs/sql_for_agents/694_content_quality_auditor_can_see_the_site_ROLLBACK.sql
-- restores default_config from agent_definitions_bak_694 (or the snapshot_agent
-- row this file took first).
