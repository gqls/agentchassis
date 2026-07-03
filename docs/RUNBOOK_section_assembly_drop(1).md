# RUNBOOK — Sections dropped between page_components and the deployed page

**Created:** 2026-07-03  •  **Updated:** 2026-07-03
**Plan:** PLAN_section_assembly_drop.md
**Scope:** vonc.com index — provocation-card(pos 2) + lobby-grid(pos 5) present in
`page_components` (deployed) but absent from the deployed `index.html`.

> YOU ARE HERE: Phase 0 (diagnosis). Cause NOT confirmed. Do not apply a fix until D1–D4
> identify where the two sections are dropped.

Constants:
- site_id = 9ec3b9ee-5b08-461b-b4f8-9e1e03579c74
- index page_id = b4d24f8e-fccd-49df-9dad-aa56a0b20a68
- provocation-card component_id = 6163ff14-9f94-4962-aa19-d2718eabdeb1
- lobby-grid component_id = 9304f14d-e19b-4ce1-b3fd-f6a315aec6ed

---

## Phase 0 — Diagnosis (confirm the cause)

### D1 — Compare the 6 page_components across filter-candidate columns  [run first]
```sql
SELECT pc.position, cc.function,
       pc.build_status, pc.slot_name, pc.parent_instance_id, pc.schema_mode,
       pc.deploy_commit,
       (pc.content_hash IS NOT NULL) AS has_hash,
       (pc.locked_at IS NOT NULL)    AS locked,
       LENGTH(COALESCE(pc.rendered_html,'')) AS rlen,
       (pc.rendered_html LIKE '%<no value>%')     AS has_no_value,
       (pc.rendered_html LIKE '%data-component%') AS has_data_component
FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.page_id = 'b4d24f8e-fccd-49df-9dad-aa56a0b20a68'::uuid
ORDER BY pc.position;
```
Read: what do positions 2 (provocation-card) and 5 (lobby-grid) have in common that
1/3/4/6 do NOT? Prime suspects: `has_no_value = t` on 2+5 only; a non-null `slot_name`
or `parent_instance_id`; a distinct `schema_mode`; a null `deploy_commit` on 2+5 while
1/3/4/6 have a commit (⇒ the deploy only committed 4). Whichever column splits 2+5 from
the rest points at the filter.

### D2 — What the deploy step reported
```sql
SELECT item_key, status, completed_at,
       jsonb_pretty(result #> '{response,deploy_result}') AS deploy_result
FROM site_work_items
WHERE site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid
  AND item_key LIKE 'manual-rebuild-index-%'
ORDER BY created_at DESC LIMIT 1;
```
Read: does `deploy_result.rendered_page.html` contain `provocation-card-section` /
`lobby-grid-section`? If NO, the drop happened at or before page-rerender's assembly
(rules toward H1/H3/H4). Look for any section-count or skip fields.

### D3 — page-rerender workflow (find the assemble action)
```sql
SELECT type, jsonb_pretty(default_config->'workflow'->'steps')
FROM agent_definitions
WHERE type = 'page-rerender' AND is_active = true;
```
Read: identify the step whose action assembles the page (likely `assemble_page` /
`compile_page_sections` / `render_page` / similar) and note its config (what it reads —
`page_id`/`site_id` → page_components — and any filter fields).

### D4 — Read the assemble action's Go code
Pull the action file named by D3 (e.g. `assemble_page_action.go` or the compile action).
Look specifically for:
- the SELECT against `page_components` — its WHERE/ORDER (does it filter on
  `build_status`, `rendered_html <> ''`, `<no value>`, `parent_instance_id IS NULL`,
  `slot_name`, `schema_mode`, a "skip interactive/empty" branch?).
- any per-section skip/continue (empty content, invalid HTML, dedupe).
This is the decisive artifact — it says exactly why 2 rows are excluded.

### D5 — Confirm Mode-B marker on the dropped rows (supports D1)
```sql
SELECT cc.function,
       (pc.rendered_html LIKE '%<no value>%') AS rendered_has_no_value,
       (cc.html_template LIKE '%<no value>%')  AS template_has_no_value
FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.page_id='b4d24f8e-fccd-49df-9dad-aa56a0b20a68'::uuid
  AND cc.function IN ('provocation-card','lobby-grid');
```
If `rendered_has_no_value = t` and the assemble action (D4) skips `<no value>` sections,
cause = confirmed (assembly drops Mode-B shells).

### Decision gate
- D1/D4 show an assemble-time filter (Mode-B / `<no value>` / empty / slot / interactive)
  → **H1 confirmed** → Phase 1A.
- D2 shows the content-writer `page_html` already had 4 and deploy used it verbatim
  → **H2 confirmed** → Phase 1B.
- D2 shows page-rerender assembled from page_components but still 4, and D4 shows a skip
  → back to H1 (1A). If D4 shows NO skip yet 4 assembled → widen: inspect the query's
  ORDER/joins and whether the 2 rows are even returned (add logging / a COUNT).

---

## Phase 1 — Fix (apply ONLY after the gate; provisional)

### 1A — Assemble filter wrongly drops valid sections (expected most-likely)
- In the assemble action, change the section selection so a section present in
  `page_components` for the page is NOT dropped for being a Mode-B/empty/`<no value>`
  shell or for being interactive. A shell with a `data-component` attribute + selectors
  is a valid, intentional section (runtime-filled). ALTER the existing action; keep the
  page-rerender workflow unchanged. Note any variable/signature change explicitly.
- If the filter exists to hide genuinely-broken `<no value>` output, the correct move is
  to make provocation-card + lobby-grid emit clean shells (no literal `<no value>` in
  rendered_html) — i.e. fix render_component for Mode-B, or regenerate them out of Mode-B
  — rather than teaching the assembler to ship `<no value>`. Decide based on D4/D5.

### 1B — Deploy used the wrong (4-section) source
- Point `deploy_page` / page-rerender at the stored `page_components` set (authoritative),
  matching its documented behaviour. Verify it re-assembles all rows for the page.

### After the fix
- Re-assemble + redeploy the index WITHOUT a full content rebuild (raise `needs_rerender`
  for the index so page-rerender re-runs assembly+deploy on the existing page_components —
  avoids re-triggering the content pipeline). If the fix is in the action, redeploy the
  action image first.

---

## Verification
```sql
-- both sections present in the freshly-assembled deploy_result
SELECT item_key, completed_at,
       (result::text LIKE '%provocation-card-section%') AS has_prov,
       (result::text LIKE '%lobby-grid-section%')       AS has_lobby
FROM site_work_items
WHERE site_id='9ec3b9ee-5b08-461b-b4f8-9e1e03579c74'::uuid
  AND (item_key LIKE 'manual-rebuild-index-%' OR item_key LIKE '%rerender%')
ORDER BY created_at DESC LIMIT 3;
```
Then the live check on vonc.com/index.html after propagation: six sections present, in
plan order (hero, provocation-card, gauntlet-cta, brief-explanation, lobby-grid,
system-stats); provocation-card fills via the loader; footer intact.

## Do NOT
- Apply 1A/1B before D1–D4 confirm the drop point.
- Teach the assembler to ship literal `<no value>` — fix the shell rendering instead.
- Run a full content rebuild to "fix" this — it re-invokes the content pipeline and its
  own deferral/clobber behaviours; a rerender (assemble-only) is the right redeploy.


---

## Phase 0 progress — D1 result (2026-07-03)
D1 returned FLAT: all 6 index page_components identical across build_status(deployed),
slot_name(=function), parent_instance_id(null), schema_mode(null), deploy_commit(null),
has_hash(f), locked(f), has_data_component(t), and has_no_value(f for ALL, incl
provocation-card + lobby-grid). => RULES OUT the <no value>/empty-content/column filters.
Nothing at the row level distinguishes the 2 dropped from the 4 kept.

NARROWED HYPOTHESIS: the only content distinguisher is a RAW INLINE `<script>` in
rendered_html. provocation-card + lobby-grid are the only two with an inline `<script>`
(hover/interaction) still in rendered_html; the four that shipped are pure CSS+markup.
This is the SAME pair with the extraction bug (js_content=0 → script never separated to
`/tools/assets/{fn}.js`, stays inline). So the drop is likely a DOWNSTREAM SYMPTOM of the
extraction bug: the assembler drops/strips sections that still carry a raw inline `<script>`.
CONFIRM at the assemble action (D3/D4) before asserting — correlation only.

Next: (a) confirm the inline-script correlation is exact (rendered LIKE '%<script%' + js_content_len);
(b) D2 deploy_result — is the drop inside page-rerender's assembly or downstream (GitHub/
Actions/B2)?; (c) D3 page-rerender workflow → the assemble action; (d) D4 read that action's
inline-script handling.

If confirmed (assembler drops raw-inline-script sections): the STRUCTURAL fix connects to
the extraction-bug fix — regenerate provocation-card + lobby-grid so separateInlineJS
extracts their scripts to js_content (rendered_html then has NO raw inline script → assembler
keeps them, and the scripts ship as /tools/assets/{fn}.js). Fixing the assembler to ship raw
inline scripts would contradict the healthy-component design (has_raw_inline=f) — prefer the
extraction fix. Decide on D4.


---

## Phase 0 progress — inline-script correlation + D2 + D3 (2026-07-03)
CONFIRMED (inline-script correlation exact): only provocation-card(2) + lobby-grid(5) have
rendered_has_inline_script=t; hero/gauntlet-cta/brief-explanation/system-stats=f. ALL six
js_content_len=0. => the 4 kept have NO script; the 2 dropped carry a RAW inline <script>
in rendered_html with nothing in js_content.

CONFIRMED (D2 — drop is INSIDE page-rerender assembly): deploy_result assembled HTML has
gauntlet-cta-section=t but provocation-card-section=f AND lobby-grid-section=f. So the two
are already absent in page-rerender's OWN output — the loss is in assembly, BEFORE
git/Actions/Backblaze. (item manual-rebuild-index-05251582, completed 13:17:56.)

LOCATED (D3 — the assembler): page-rerender workflow. Assemble step = `render_page`
(action `rerender_single_page`, "Assemble page from stored components") → output
rendered_page.files → `deploy_page` (action `git_commit`, repo 'sites') ships those files.
rerender_single_page is the same action that carries collectJSAssets (extraction-bug work).
Two entry paths in the workflow: reason image_landed/section_data_resolved →
rerender_sections → … → render_page; else → render_page directly. Both assemble via
render_page.

=> The drop is in rerender_single_page_action.go: it assembles from page_components and
excludes the whole section for components with a raw inline <script> + empty js_content
(the un-migrated/broken extraction-bug shape). CONFIRM by reading the action (D4).

### D4 — read rerender_single_page_action.go (DECISIVE)
Look for: the per-component loop and any skip/continue; whether the skip keys on
js_content empty, on `<script` present in rendered_html, or a has_raw_inline check; how
collectJSAssets treats a component whose inline script was never separated.

### Fix direction (provisional, confirm on D4)
Most likely = the EXTRACTION-BUG fix already scoped: regenerate provocation-card + lobby-grid
so separateInlineJS lifts their inline <script> into js_content → rendered_html clean (no raw
inline script) → assembler keeps them → behaviour ships as /tools/assets/{fn}.js. A narrower
assembler change (keep sections but strip/relocate the raw inline script) is the alternative
if regeneration is blocked — decide on D4. Reuse/alter the existing action; do not add raw
inline scripts against the healthy-component design (has_raw_inline=f).
Redeploy after fix via needs_rerender (assemble-only), not a full content rebuild.


---

## D4 RESULT — CAUSE CONFIRMED (2026-07-03): sectionHasVisibleContent, NOT the extraction bug
Read rerender_single_page_action.go. assemblePage → getPageSections → for each section
calls `sectionHasVisibleContent(html)`: strips <style> + <script> blocks, tags, entities,
whitespace, and KEEPS the section only if >10 chars of visible text remain; else it is
SKIPPED (logged "skipping section with no visible content"). provocation-card + lobby-grid
are Mode-B shells with EMPTY build-time fields → <10 visible chars → dropped.
=> CORRECTION: the earlier "extraction-bug / raw inline <script>" hypothesis is WRONG. The
   filter strips <script> BEFORE measuring, so the script is irrelevant. This is a
   visible-TEXT filter. The extraction bug (js_content=0) is a SEPARATE, unrelated issue.
   (Value of reading the action before asserting.)

ASYMMETRY:
 - provocation-card: legitimately runtime-filled (Path-2 loader fills it in-browser,
   proven filling in Phase 2). Empty-at-build is BY DESIGN → the drop is WRONG.
 - lobby-grid: genuinely empty (Mode-B, NO loader, hover script doesn't fill). Would ship
   blank → the drop is arguably CORRECT. It should stay dropped until built (content or a
   loader). Its build is a SEPARATE task.

## DECISION — fix
CHOSEN (structural): make the assembler treat runtime-filled sections as first-class.
 - PATCH_section_visible_content.go: sectionHasVisibleContent gains an early exemption —
   a section carrying `data-runtime-fill` is kept regardless of build-time text. One added
   package regexp (reRuntimeFill) + one early return. Signature unchanged, no renames,
   strictly additive (unmarked sections behave exactly as before).
 - Marker SQL: add data-runtime-fill="true" to provocation-card's <section> in
   html_template + the current index page_components.rendered_html (guarded NOT LIKE
   '%data-runtime-fill%').
 - lobby-grid: NOT marked → stays correctly filtered until built.
 - Order: (1) deploy the code change (agent-chassis image rebuild); (2) run marker SQL;
   (3) raise needs_rerender for the index → page-rerender re-assembles + redeploys (NO full
   content rebuild). The marker is only honoured by the fixed assembler, hence the order.
ALTERNATIVE (data-only, no code deploy): give provocation-card build-time fallback text
 (e.g. "Loading today's provocation…") that the loader overwrites → passes the existing
 filter + graceful no-JS fallback. Per-component (less framework-robust). Stopgap only.

Expected after Option A: index deploys FIVE sections (hero, provocation-card, gauntlet-cta,
brief-explanation, system-stats); provocation-card shell present → loader fills it live.
lobby-grid remains absent (correct) pending its own build.
Verify: deploy_result assembled HTML contains provocation-card-section; live index shows
the provocation card (filled) + footer; lobby-grid still absent (expected).
