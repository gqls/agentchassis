-- ============================================================================
-- 274 — oufe.com second tool: "The relevant alternative: who still has a vote"
--
-- HOW TO LOAD THE HTML. The component body lives in
-- docs/agent_docs/docs024_key_docs_latest/oufe/tool/tool-relevant-alternative.html,
-- NOT inline here: a 20KB template pasted into SQL is unreviewable and every edit
-- would have to be made twice.
--
-- *** DO NOT USE `\set tmpl \`cat ...\`` WITH `kubectl exec`. IT FAILS SILENTLY. ***
-- Tool 1's file (PREPARED_tool_insert.sql) documents that form, and it only works
-- when psql runs on a machine that HAS the repo. Piped through
-- `kubectl exec -i postgres-clients-0 -- psql`, the backticks execute INSIDE THE
-- POD, which has no repo: `cat` prints "No such file or directory" to stderr, the
-- variable resolves to EMPTY, and the INSERT still COMMITS. Hit on 2026-07-31:
-- the component was created with `html_template = ''` (length 0, not NULL) and
-- every statement reported success. `\set ON_ERROR_STOP on` does not catch it —
-- the shell command failing is not a psql error.
--
-- Build the statement LOCALLY instead, with the template in a dollar-quoted
-- literal, then pipe the finished SQL in:
--
--   python3 - <<'EOF'
--   html = open("docs/agent_docs/docs024_key_docs_latest/oufe/tool/tool-relevant-alternative.html").read()
--   assert "$ratmpl$" not in html
--   open("/tmp/fix.sql","w").write("UPDATE content_components SET html_template = $ratmpl$"+html+"$ratmpl$ WHERE function='tool-relevant-alternative';")
--   EOF
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db < /tmp/fix.sql
--
-- Then ASSERT THE LENGTH, do not assume it: compare `length(html_template)` to
-- `wc -c` on the local file. They must match exactly.
--
-- THE FOUR CONSTRAINTS THIS FILE HONOURS (all learned the hard way on tool 1)
--   1. JS is INLINE in html_template and js_content stays NULL. The js_content
--      lane publishes /tools/assets/<function>.js and then injects no <script>
--      tag — published-but-inert, the bugs_open/041 class.
--   2. IDENTITY MUST AGREE across three places or acceptance resolves the wrong
--      URL and silently tests nothing (landmine G6):
--        pages.name = content_components.function = doc_plans.subject_key
--        = 'tool-relevant-alternative'
--   3. slot_name must equal the component FUNCTION and must ALSO appear in
--      pages.sections. 'main' is wrong and fails SILENTLY (bugs_open/095).
--   4. Tool pages are rebuild_policy='owned' so the generic page builder never
--      rewrites them (save_page_sections_action.go hard-refuses).
--
--   5. NEW, 2026-07-30: rebuild_policy='owned' now also matters for a second
--      reason. save_page_sections carries a CLAIMS FLOOR (CLM-018) that can
--      REFUSE a save outright. An owned page never reaches it, but if this tool
--      is ever converted to a generic page, its copy has to survive that gate.
--      Nothing in this template trips it: checked with cmd/claimscan against
--      oufe's own register before shipping, 0 findings.
-- ============================================================================

\set ON_ERROR_STOP on
-- \set tmpl `cat docs/agent_docs/docs024_key_docs_latest/oufe/tool/tool-relevant-alternative.html`

BEGIN;

-- 1. the component -----------------------------------------------------------
INSERT INTO content_components (
  name, function, display_name, description, category,
  component_level, render_mode, html_template, js_content,
  input_schema, is_active, suitable_page_types
)
VALUES (
  'tool-relevant-alternative-oufe-com',
  'tool-relevant-alternative',
  'The relevant alternative: who still has a vote',
  'Client-side model of the s.901G cross-class cram down test. The reader sets a relevant-alternative value, a plan value, four class claims and each class''s vote, and sees which classes would receive anything in the counterfactual, which of them therefore has a genuine economic interest, and whether Conditions A and B are met. Moving the relevant alternative gives classes a veto and takes it away. Deterministic arithmetic, no back end, nothing transmitted.',
  'finance',
  'tool',
  'template',
  :'tmpl',
  NULL,            -- deliberate: see constraint 1 in the header
  NULL,
  true,
  '["tool"]'::jsonb
);

-- 2. the page ----------------------------------------------------------------
INSERT INTO pages (
  site_id, name, url, title, page_type, status, build_status,
  nav_label, nav_order, in_header, in_footer, rebuild_policy, meta_description
)
SELECT
  s.id,
  'tool-relevant-alternative',
  '/tools/tool-relevant-alternative.html',
  'The relevant alternative: who still has a vote | OUFE',
  'tool',
  'active',
  'pending',
  'Relevant alternative',
  41,
  false,   -- header stays short; reached from the footer Explore slot and the dossier
  true,
  'owned',
  'An interactive model of the cross-class cram down test: set what would happen if a restructuring plan were refused, and see which classes keep the ability to block it.'
FROM sites s WHERE s.domain = 'oufe.com';

-- 3. bind component to page --------------------------------------------------
UPDATE pages SET sections = '["tool-relevant-alternative"]'::jsonb
 WHERE site_id = (SELECT id FROM sites WHERE domain='oufe.com')
   AND name = 'tool-relevant-alternative';

INSERT INTO page_components (page_id, component_id, position, slot_name, build_status)
SELECT p.id, c.id, 1, 'tool-relevant-alternative', 'pending'
FROM pages p
JOIN sites s ON s.id = p.site_id
CROSS JOIN content_components c
WHERE s.domain = 'oufe.com'
  AND p.name = 'tool-relevant-alternative'
  AND c.function = 'tool-relevant-alternative'
  AND c.name = 'tool-relevant-alternative-oufe-com';

-- 4. travelling PLAN with acceptance criteria --------------------------------
UPDATE doc_plans SET is_current = false, superseded_at = now()
WHERE subject_type = 'tool' AND subject_key = 'tool-relevant-alternative' AND is_current;

INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by, is_current, pinned)
VALUES (
  'tool', 'tool-relevant-alternative',
  $plan$# PLAN — The relevant alternative: who still has a vote (oufe.com)

## What it is for
A restructuring plan can be forced on a class that votes against it. Whether it
can turns almost entirely on one assumption: what the court thinks would happen
if the plan were refused. A reader should be able to move that assumption with
one slider and watch classes gain and lose the ability to block. The point being
taught is that **the counterfactual, not the plan, decides who matters.**

## What it computes
Two strict-priority waterfalls over the same capital structure: one on the
relevant-alternative value, one on the plan value. Then, per class:

- **Genuine economic interest** — does the class recover anything at all in the
  relevant alternative? Only such a class can support Condition B.
- **Dissent** — is the class below 75% in value in favour?
- **Condition A** — is any DISSENTING class worse off under the plan than in the
  relevant alternative?
- **Condition B** — has at least one class that both approved and is in the money
  in the relevant alternative?

Verdict: cram down available only if both conditions hold and there is a
dissenting class at all.

## The behaviour that carries the teaching (verified before shipping)
With the shipped defaults (plan 1400; claims 800/500/400/0; votes 95/80/10/0):

| relevant alternative | classes in the money | Cond A | Cond B | outcome |
|---|---|---|---|---|
| 0 | 0 | Y | **N** | not available — nobody is in the money, so no approving class can satisfy Condition B |
| 600 (default) | 1 | Y | Y | **available** — the subordinated class dissents but does better under the plan than without it |
| 1500 | 3 | **N** | Y | not available — the subordinated class is now in the money in the counterfactual, recovers less under the plan, and has a veto |

That arc is the whole tool. Raising the relevant alternative hands the
subordinated class a blocking position it did not have.

## Honesty contract
Every number is computed from the reader's own inputs. Nothing is fetched,
nothing is assumed about any real company, no real capital structure is
pre-loaded, and there is no LLM in the render path. The statutory threshold and
both conditions are quoted from Part 26A of the Companies Act 2006; the site's
evidence register holds them as CIT-709026d97df11645 (s.901G / 75%),
CIT-917cae2baf2a9069 (Condition A), CIT-81d532ab22dcb359 (Condition B) and
CIT-f729a1b60a30481 (relevant alternative).

**The READING of those provisions applied here is ours and is a simplification,
and the tool says so in its own words rather than in the site footer.** It names
what it does not model: timing, risk, the form of consideration, non-financial
rights, class composition, security, guarantees, structural subordination,
intercompany claims, contingent liabilities, and the court's discretion. It
models "genuine economic interest" as "recovers something in the counterfactual",
which is a simplification and is disclosed as one.

## Privacy
Everything runs in the reader's browser. No input is transmitted or persisted,
including the acknowledgement, which is why the gate reappears on each visit.

## Delivery
Inline — all CSS and JS ship inside html_template, render_mode='template',
js_content NULL. Extracting the JS would silently kill the tool (bugs_open/041).

## Acceptance criteria

```criteria
{ "profiles": ["desktop","mobile"],
  "container": ".tool-container",
  "checks": [
    {"id":"gate-exists","type":"selector_exists","selector":"#ra-gate"},
    {"id":"gate-button-exists","type":"selector_exists","selector":"#ra-accept"},
    {"id":"boots","type":"selector_exists","selector":".tool-container"},
    {"id":"console","type":"no_console_errors"},
    {"id":"status","type":"page_status_ok"},
    {"id":"mobile-fit","type":"no_horizontal_overflow","profiles":["mobile"]},
    {"id":"default-allows-cram-down","type":"interaction",
     "steps":[{"action":"click","selector":"#ra-accept"}],
     "expect":{"selector":"#ra-verdict-head","text_matches":"Cram down available"}},
    {"id":"high-alternative-restores-the-veto","type":"interaction",
     "steps":[{"action":"click","selector":"#ra-accept"},
              {"action":"fill","selector":"#ra-alt","value":"1500"}],
     "expect":{"selector":"#ra-verdict-head","text_matches":"Cram down not available"}}
  ]
}
```

**`gate-exists` and `gate-button-exists` come FIRST, and that placement is the
point — see `bugs_open/126`.** A gated tool cannot pass a two-click acceptance
run, and the automated repair loop's response was to raise an `improve_tool`
spec satisfiable only by DELETING the disclaimer. Giving the gate its own
acceptance checks makes removing it a FAILURE rather than a fix. Any future
repair that "solves" an acceptance problem by removing `#ra-gate` now breaks two
named checks instead of quietly improving the score.

The two interaction checks assert the arithmetic actually runs, at opposite ends
of the range, and they are not redundant: the first proves `recalc()` fires on
accept, the second proves the input handler recomputes. A tool whose JS failed to
execute would keep its initial verdict text ("Set a value to begin") and fail
both. `"Cram down available"` is not a substring of `"Cram down not available"`,
so the two expectations genuinely discriminate.

Deliberately NO `asset_loads` check — that criterion asserts a JS-extraction path
that was designed and never built, and failed every tool on its first sweep
(TL-016).

Selectors are copied from the shipped template, not invented.
$plan$,
  'manual', 'oufe-workstream-2026-07-31', true, true
);

COMMIT;

-- ---------------------------------------------------------------------------
-- VERIFY (run after apply; every one of these should return exactly 1 row)
-- ---------------------------------------------------------------------------
-- SELECT function, component_level, render_mode, js_content IS NULL AS js_inline,
--        length(html_template) AS tmpl_len
--   FROM content_components WHERE function='tool-relevant-alternative';
--
-- SELECT name, url, page_type, rebuild_policy, sections, in_footer
--   FROM pages WHERE name='tool-relevant-alternative';
--
-- SELECT pc.slot_name, pc.build_status
--   FROM page_components pc JOIN pages p ON p.id=pc.page_id
--  WHERE p.name='tool-relevant-alternative';
--
-- SELECT subject_key, is_current, pinned, body LIKE '%```criteria%' AS has_criteria
--   FROM doc_plans WHERE subject_key='tool-relevant-alternative' AND is_current;
--
-- IDENTITY CHECK — all three must return the SAME string or acceptance tests nothing:
-- SELECT (SELECT name FROM pages WHERE name='tool-relevant-alternative') AS page,
--        (SELECT function FROM content_components WHERE function='tool-relevant-alternative') AS comp,
--        (SELECT subject_key FROM doc_plans WHERE subject_key='tool-relevant-alternative' AND is_current) AS plan;
