# RUNBOOK — component-regen clobber → recovery → F8 contamination
_Action document. History/detail: NOTES_component_regen_clobber.md (§1–9bd) and HANDOFF. Updated 2026-07-06._

## YOU ARE HERE → Part D, Step D2a (patch the deterministic CSS template — needs render_css_from_spec source + coordination)
Part C contamination CLEARED (sweep passes; matchmatrix DECIDED: sections=0 planning gap on a legacy page — parked
as hygiene). Part C residue: live eyeballs + the two backup DROPs. **Part D is now the active user-visible
problem**: the freshly rebuilt gripper-selection-guide confirms R6f — fresh templates consume even more undefined
vocabulary (`--hero-ink` ×12 joins the list), so every rebuild renders worse until the gap closes.

## Constants & conventions
- psql: `kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db`
- Capture to a LOCAL file (`\o` writes inside the pod):
  `kubectl -n ai-persona-system exec postgres-clients-0 -- psql -U clients_user -d clients_db -c "SQL…" > out.txt`
- `system-stats` component: `fdd92ad4-521a-4602-89cf-7ee1a66c10f1`. `brief-explanation`: resolve via
  `function='brief-explanation' AND forked_from IS NULL` (never hardcode — one live row).
- Sites: leopardess `4851f6fc…` · ai-agent-orch `2a8ebf9c…` · gamesdesign (test) `e33263f4…`
- Optimistic-lock pattern for co-managed writes: `WHERE updated_at = '<last-known>'` — **UPDATE 0 = stop, re-read, coordinate.**
- Snapshot-before-change: manual `component_versions` INSERT (columns as in Part C Step 6).
- Stale system-stats fingerprint (historic): md5 `462c7c4501b8137d06a0f4c3e7592772`, len 7369.
- Fleet: v1.0.1094 since 2026-07-04 19:21 (parallel chats) — re-read workflows before relying on cached knowledge.

---

## Part A — CLOSED: system-stats regen clobber + recovery (verified 2026-07-03)
Root cause: pre-guard regen renamed the shared field contract → five dependents rendered empty.
Fixes deployed + proven: **F1** store guard (rejects renames) · **F1-prompt** loader + field-name rule + function pin ·
**F3** scoped reason-stamped rerenders. Recovery: re-key (20 mappings) + real `cta` site_specs (`/contact.html` ×3,
verified) → all five pages re-rendered, distinct md5s, needles true; leopardess confirmed live. vonc.com/index = healthy
sixth dependent. Standing verify (any time):
```sql
SELECT s.domain, p.name, md5(pc.rendered_html) AS md5, pc.updated_at,
       strpos(pc.rendered_html, coalesce(nullif(pc.content_data->>'stat1_label',''),chr(1)))>0 AS has_stat1_label,
       strpos(pc.rendered_html, coalesce(nullif(pc.content_data->>'cta_url',''),chr(1)))>0    AS cta_url_baked
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.component_id='fdd92ad4-521a-4602-89cf-7ee1a66c10f1' ORDER BY 1,2;
-- PASS: no md5 = 462c7c45…, needles true
```

## Part B — CLOSED: R6c/R6d gripper-detail "blank page"
One artifact all along (md5-identical fetches; earlier mis-assembly/stale-cache readings were grep-metric artifacts).
DB + assembly + deploy healthy. The blank is **theming** → Part D (R6f). Assembly facts (from code): membership =
`page_components` by `position` with a >10-visible-chars filter; head/header/footer from `site_components`;
`pages.sections` jsonb is metadata, not membership.

---

## Part C — ACTIVE: R6e / F8 — brief-explanation contamination
Mechanism (proven): the 2026-07-01 12:46 pre-guard regen baked vonc's product ("Spark", the daily Gauntlet) into the
shared component via THREE carriers: (1) static-field fallback values, (2) values merged into dependents'
content_data, (3) per-field **llm_guidance** instructing every writer pass. The F1 guard checks names only → F8.

- [x] **Step 1 — snapshot v2** (`INSERT 0 1`, also backfills the unversioned 07-03 13:22 change).
- [x] **Step 2 — neutralize fallbacks** (stats → `llm` optional; CTAs → "Get Started"/"Learn More"). `UPDATE 1`.
  CHECK (should still hold):
  ```sql
  SELECT key, value->>'source' AS source, left(value->>'fallback',40) AS fallback
  FROM content_components, jsonb_each(input_schema->'fields')
  WHERE function='brief-explanation' AND forked_from IS NULL ORDER BY 1;
  -- PASS: stats source=llm no fallback; cta_* static neutral
  ```
- [x] **Step 3 — backup + strip merged keys** (`page_components_bak_briefexp_20260703` = 3; strip `UPDATE 3`).
- [x] **Step 4 — scoped rerenders** (`f8_rerender_briefexp:<id>` ×2, `INSERT 0 2`). Outcome: robot-hands index +
  how-it-works clean (9181B neutral shells); auto-escalations created `needs_page` for matchmatrix +
  gripper-selection-guide; two child errors were parent-topic/pod-death noise.
- [x] **Step 5 — residual diagnosis** (A–D + writer config + guidance dump): idea.uk ×2 and the 18:04
  gripper-selection-guide rebuild carry the pitch in **generated llm copy**; carrier = `llm_guidance` (all 11 fields).

- [x] **Step 6 — snapshot v3** ✔ 2026-07-06 (INSERT 0 1; versions 1/2/3 verified):
  ```sql
  INSERT INTO component_versions
    (component_id, version_number, html_template, input_schema,
     change_description, changed_by, change_source, created_at)
  SELECT cc.id,
         (SELECT COALESCE(MAX(version_number),0)+1 FROM component_versions v WHERE v.component_id = cc.id),
         cc.html_template, cc.input_schema,
         'F8 carrier 3: neutralize vonc/Spark product spec in llm_guidance (all 11 llm fields)',
         'manual-recovery:f8', 'manual', NOW()
  FROM content_components cc
  WHERE cc.function='brief-explanation' AND cc.forked_from IS NULL;
  -- expect INSERT 0 1
  ```
  CHECK:
  ```sql
  SELECT cv.version_number, cv.changed_by, cv.created_at
  FROM component_versions cv JOIN content_components cc ON cc.id=cv.component_id
  WHERE cc.function='brief-explanation' ORDER BY 1;
  -- PASS: rows 1,2,3 present; v3 = manual-recovery:f8, just now
  ```

- [x] **Step 7 — neutralize llm_guidance ×11** ✔ 2026-07-06 (UPDATE 1 — lock held across 3 idle days; verify f|f|f + 11 neutral strings):
  ```sql
  WITH new_guidance(field, guidance) AS (VALUES
    ('badge_label',      'Short badge text overlaid on the illustration image (2–5 words) that feels current and active, e.g. ''Live Now'' or ''Now Available''.'),
    ('description',      'One or two sentences (max 40 words) that tell a first-time visitor what the product or service is and how it works at a glance. Confident and direct, no filler.'),
    ('eyebrow',          'A short uppercase eyebrow label (2–4 words) that frames what this section is, e.g. ''How It Works''. Sets the section''s context before the heading.'),
    ('heading',          'A punchy, confident heading (6–10 words) with exactly ONE word or short phrase wrapped in <em> tags to receive the accent-colour emphasis. Should communicate the company''s core offering in the site''s tone.'),
    ('illustration_alt', 'Descriptive alt text for the site illustration image (10–15 words). Should describe what the illustration depicts in the context of the company''s offering.'),
    ('step_1_sentence',  'One sentence (max 18 words) expanding on step 1 — how the process begins for the visitor or customer.'),
    ('step_1_title',     'Bold title for step 1 of the process (3–5 words) describing how it starts.'),
    ('step_2_sentence',  'One sentence (max 18 words) expanding on step 2 — what the visitor or customer does next.'),
    ('step_2_title',     'Bold title for step 2 (3–5 words) describing the middle step.'),
    ('step_3_sentence',  'One sentence (max 18 words) expanding on step 3 — the outcome or result at the end of the process.'),
    ('step_3_title',     'Bold title for step 3 (3–5 words) describing the outcome.')
  ),
  patched AS (
    SELECT cc.id,
           jsonb_set(cc.input_schema, '{fields}',
             (SELECT jsonb_object_agg(
                       f.key,
                       CASE WHEN ng.field IS NOT NULL
                            THEN f.value || jsonb_build_object('llm_guidance', ng.guidance)
                            ELSE f.value END)
              FROM jsonb_each(cc.input_schema->'fields') f
              LEFT JOIN new_guidance ng ON ng.field = f.key)) AS new_schema
    FROM content_components cc
    WHERE cc.function='brief-explanation' AND cc.forked_from IS NULL
  )
  UPDATE content_components cc
  SET input_schema = p.new_schema, updated_at = NOW()
  FROM patched p
  WHERE cc.id = p.id
    AND cc.updated_at = '2026-07-03 17:17:38.725719+00';
  -- expect UPDATE 1  (0 = moved under us — stop, re-read, coordinate)
  ```
  (How it works: `jsonb_each` explodes fields → one row each; `||` overwrites just `llm_guidance` where a
  replacement exists; `jsonb_object_agg(key, value)` reassembles the rows into one object; `jsonb_set` puts it back.)
  CHECK:
  ```sql
  SELECT strpos(lower(input_schema::text),'gauntlet')>0 AS gauntlet,
         strpos(lower(input_schema::text),'spark')>0     AS spark,
         strpos(lower(input_schema::text),'provocation')>0 AS provocation
  FROM content_components WHERE function='brief-explanation' AND forked_from IS NULL;
  -- PASS: f | f | f
  SELECT f.key, left(f.value->>'llm_guidance',60) FROM content_components,
    LATERAL jsonb_each(input_schema->'fields') f
  WHERE function='brief-explanation' AND forked_from IS NULL AND f.value ? 'llm_guidance' ORDER BY 1;
  -- PASS: the 11 neutral strings
  ```

- [x] **Step 8 — capture the `needs_page` shape** ✔ (handler `page-build-handler`; key `needs_page:<page>`; spec `{reason: content_data_backfill, page_name}`):
  ```sql
  SELECT w.handler_agent, w.item_key, w.status, jsonb_pretty(w.spec) AS spec
  FROM site_work_items w JOIN sites s ON s.id=w.site_id
  WHERE s.domain='robot-hands.com' AND w.item_type='needs_page'
  ORDER BY w.updated_at DESC LIMIT 2;
  ```

- [x] **Step 9 — re-triage + clone the rebuilds** ✔ (gripper-selection-guide 12:59 f · how-it-works 13:01 f; matchmatrix no-op'd twice → reclassified, see Step 12). matchmatrix's rebuild never ran (band still
  2026-07-01 12:46, item `complete` = dispatch stamp) and gripper-selection-guide's ran under contaminated
  guidance — re-triage both; how-it-works gets a fresh item mirrored from gripper's row (no guessed columns).
  **9a:**
  ```sql
  UPDATE site_work_items w
  SET status = 'triaged', updated_at = NOW()
  FROM sites s
  WHERE s.id = w.site_id AND s.domain = 'robot-hands.com'
    AND w.item_type = 'needs_page' AND w.status = 'complete'
    AND w.spec->>'page_name' IN ('matchmatrix','gripper-selection-guide');
  -- expect UPDATE 2
  ```
  **9b:**
  ```sql
  INSERT INTO site_work_items (site_id, source, pipeline, item_type, severity, summary, priority,
                               handler_agent, status, created_by, spec, item_key)
  SELECT w.site_id, w.source, w.pipeline, w.item_type, w.severity,
         'F8: rebuild how-it-works content under the cleaned brief-explanation contract',
         w.priority, w.handler_agent, 'triaged', 'manual-recovery:f8',
         jsonb_build_object('reason','content_data_backfill','page_name','how-it-works'),
         'needs_page:how-it-works'
  FROM site_work_items w JOIN sites s ON s.id = w.site_id
  WHERE s.domain = 'robot-hands.com' AND w.item_key = 'needs_page:gripper-selection-guide'
    AND NOT EXISTS (SELECT 1 FROM site_work_items x
                    WHERE x.site_id = w.site_id AND x.item_key = 'needs_page:how-it-works'
                      AND x.status NOT IN ('complete','verified','rejected','wont_fix','failed'))
  LIMIT 1;
  -- expect INSERT 0 1
  ```
  9a ✔ UPDATE 2 · 9b ✔ INSERT 0 1 (2026-07-06).
  PROGRESS (re-run as often as you like; `\watch 120` works):
  ```sql
  SELECT p.name,
         w.status AS item_status,
         (w.error IS NOT NULL AND w.error <> '') AS has_err,
         w.updated_at AS item_stamp,
         pc.updated_at AS band_stamp,
         (pc.updated_at > '2026-07-06 11:50') AS band_rebuilt,
         strpos(lower(pc.rendered_html),'gauntlet') > 0 AS gauntletised
  FROM pages p
  JOIN sites s ON s.id = p.site_id
  LEFT JOIN LATERAL (
    SELECT * FROM site_work_items w
    WHERE w.site_id = s.id AND w.item_type = 'needs_page' AND w.spec->>'page_name' = p.name
    ORDER BY w.updated_at DESC LIMIT 1
  ) w ON true
  LEFT JOIN page_components pc
    ON pc.page_id = p.id
   AND pc.component_id = (SELECT id FROM content_components
                          WHERE function = 'brief-explanation' AND forked_from IS NULL)
  WHERE s.domain = 'robot-hands.com'
    AND p.name IN ('matchmatrix','gripper-selection-guide','how-it-works')
  ORDER BY p.name;
  ```
  Reading: `triaged` = waiting for the loop · `claimed` = writer running (if claimed >30 min with no band movement,
  it's stuck — re-triage per the Step-9a pattern) · `complete` with `band_rebuilt = f` = dispatch stamp only, the
  rebuild hasn't actually happened — keep waiting or investigate · **PASS = all three rows `band_rebuilt = t` and
  `gauntletised = f`** (band stamps after 2026-07-06 11:50).
  PROGRESS 2026-07-06 13:15: gripper-selection-guide ✔ 12:59 f · how-it-works ✔ 13:01 f (items lag in
  'claimed' — band stamp is the proof). matchmatrix = SECOND no-op (complete again, band still 07-01 12:46, no
  error; both re-triages claimed in one batch sweep). Suspect: `page_type` (MatchMatrix is a tool page — the
  handler may branch/skip by design). Check `p.page_type` ×3 + matchmatrix band cd_keys; if tool-ish, dump the
  page-build-handler config and DECIDE leave-benign-old-band vs tool-page flow — matchmatrix is NOT contaminated,
  so no forced third re-triage.
  CHECK (after the loop runs — give it time; writer rebuilds are slow):
  ```sql
  SELECT p.name, pc.updated_at AS band_stamp, strpos(lower(pc.rendered_html),'gauntlet')>0 AS gauntletised
  FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
  JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  WHERE s.domain='robot-hands.com' AND cc.function='brief-explanation'
    AND p.name IN ('matchmatrix','gripper-selection-guide','how-it-works')
  ORDER BY 1;
  -- PASS: all three band_stamps fresh, gauntletised = f
  ```

- [x] **Step 10 — idea.uk re-pass** ✔ — regenerated 2026-07-06 13:10/13:21 (other chat's pass / their pipeline); both pages gauntletised = f.

- [x] **Step 11 — matchmatrix verified never-rebuilt** (band 12:46, item complete-at-dispatch — 7th semantics sighting). Folded into Step 9a.

- [ ] **Step 12 ← NEXT — closeout**. The contamination sweep PASSES (f except vonc). Remaining:
  (a) matchmatrix — DECIDED 2026-07-06: sections_listed=0 (siblings 9/5) ⇒ planning-data gap on a legacy page
  (repo: last touched 2 months ago) — PARKED as hygiene (later: planner reconcile/adopt, or retire). Discriminator
  kept for reference:
  ```sql
  SELECT p.name, jsonb_array_length(p.sections) AS sections_listed, p.build_status, p.last_built_at, p.deployed_at
  FROM pages p JOIN sites s ON s.id = p.site_id
  WHERE s.domain='robot-hands.com' AND p.name IN ('matchmatrix','gripper-selection-guide','how-it-works');
  -- matchmatrix sections_listed=0 while siblings >0 ⇒ planning-data gap (handler had nothing to write) — hygiene
  ```
  plus eyeball robot-hands.com/matchmatrix.html live.
  (b) live eyeballs: robot-hands index + the two rebuilt pages; idea.uk; the outstanding ai-agent case-study page
  (Part A ledger).
  (c) drop the backups when comfortable:
  `DROP TABLE page_components_bak_sysstats_20260702; DROP TABLE page_components_bak_briefexp_20260703;`
  Standing sweep (re-run any time):
  ```sql
  SELECT s.domain, p.name, pc.updated_at, strpos(lower(pc.rendered_html),'gauntlet')>0 AS gauntletised
  FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
  JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  WHERE cc.function='brief-explanation' ORDER BY 1,2;
  -- PASS: f everywhere except vonc.com (its own legitimate content)
  ```
  Then live eyeballs (robot-hands index; idea.uk after their re-pass), the outstanding ai-agent case-study page
  eyeball from Part A, and drop the backups when comfortable:
  `DROP TABLE page_components_bak_sysstats_20260702; DROP TABLE page_components_bak_briefexp_20260703;`

---

## Part D — ACTIVE: R6f theming vocabulary drift (structural design settled)

CORRECTED DIFF (proper defined-list capture — my earlier grep was lossy; --section-* IS defined): the real gap is
11 names in two patterns — SYNONYMS (--border-radius vs --radius/-sm/-lg · --shadow vs --shadow-sm/md/lg ·
--spacing-section vs --section-pad-y · --container-max-width vs --container-max · legacy --primary-color/
--secondary-color/--accent-color vs --color-*) and ORPHANS (--hero-ink, --color-heading, --color-white,
--color-error). Structural reading: two generation surfaces without a shared token contract — except it turns out
only ONE is an LLM: **styles.css is rendered by a deterministic Go template** (webdesign-agent: analyze_design LLM
→ design spec JSON → `render_css_from_spec` "Go template, no LLM" → git_commit). The vocabulary lives in ONE file.
(storage_actions.go's styles.css writes = the OLD builder extract paths — NOT this flow; don't patch there.)

- [ ] **D2a ← NEXT — extend the Go CSS template in `render_css_from_spec`** with a compatibility section emitted
  for every site:
  ```css
  /* --- component-vocabulary compatibility (canonical aliases) --- */
  --border-radius: var(--radius);
  --shadow: var(--shadow-md);
  --spacing-section: var(--section-pad-y);
  --container-max-width: var(--container-max);
  --primary-color: var(--color-primary);
  --secondary-color: var(--color-secondary);
  --accent-color: var(--color-accent);
  --color-heading: var(--color-text);
  --color-white: #ffffff;
  --color-error: #d64545;
  --hero-ink: var(--color-text);
  ```
  NEEDED: the action source — `grep -rln "render_css_from_spec" platform/orchestration/actions/ --include=*.go`
  → upload. **COORDINATE FIRST**: the webdesign-agent row was updated 2026-07-06 13:22 (v1.0.1096) — the parallel
  chats are actively on this agent; re-read its workflow at patch time.
- [ ] **D2b — prevention**: component-creator prompt gains a canonical-token rule (template CSS uses ONLY the
  canonical list + sanctioned aliases — stops new orphans like --hero-ink); optional store-time WARN lint on
  unknown var(--…) names (F8's sibling).
- [ ] **D2c — rollout**: sites pick the fix up on their next webdesign pass. CAUTION: re-running the agent re-runs
  analyze_design (fresh LLM spec — palette can shift unless pinned via design_intent/design_reference). Low-churn
  interim for robot-hands (+vonc): the manual bridge commit (same CSS block, palette preserved) until their next
  natural pass.

## Part E — Flags & hygiene
- [ ] **F8 mitigations**: store-time lint + prompt rule — site-neutrality for fallbacks, merged statics, and
  especially `llm_guidance` (the strongest carrier: it instructs every future writer pass).
- [ ] **F7 (residual)**: `update_component_html` swaps templates without placeholder⇄schema sync validation
  (its snapshot INSERT is already fixed). Hero's 07-02 16:43 zero-version update unexplained but benign-looking.
- [ ] **F6**: align store guard's NOT EXISTS status list with `idx_swi_dedup` (`unresolved` differs); gate
  `itemsCreated++` on RowsAffected.
- [ ] **F5**: guard extension — regen ADDING a required fallback-less field strands renderability.
- [ ] **F4 (softened)**: fork-instead-of-match advisory (evidence: the F2-3b fork).
- [ ] Hygiene: 40 stale `unresolved` page_rerender items (2026-05-01); loose dispatch item-status semantics
  (7 sightings incl. errors-in-complete and parent-topic-vanished noise).
