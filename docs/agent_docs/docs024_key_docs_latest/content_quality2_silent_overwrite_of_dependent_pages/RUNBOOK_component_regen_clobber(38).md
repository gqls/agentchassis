# RUNBOOK — component regeneration silently empties dependent pages

## The problem (start here)

A shared section component (one `content_components` row) is reused by many pages across several sites, and
each page stores its own `content_data` whose keys must match the component template's `{{.field}}` placeholders
exactly — `RenderTemplate` (component_library.go) fills a placeholder only when `content_data` has a key of the
same name, and silently leaves it empty otherwise (it runs Go `text/template` and then strips the resulting
`<no value>` tokens to empty string, logging only a warning, no error). For the `system-stats` component
(`fdd92ad4`), R3 confirmed the five live instances render an **identical content-free shell** (one md5, 7369
bytes each) because their `content_data` holds **old** key names (`eyebrow`, `heading`, `stat_1_number`, …) that
the current template no longer reads. The shared component itself is healthy (22 fields, template/schema synced).

When component-creator regenerates such a component, `StoreGeneratedComponentAction` overwrites its
`html_template` + `input_schema` **in place** (same `component_id`, so every dependent now points at the new
contract), snapshots the old version, marks dependents `build_status='pending'`, and raises a `needs_rerender`
work item per site — but it does **not** migrate the dependents' `content_data`, and that `needs_rerender` carries
no `reason`, so the rerender it triggers is **assemble-only** (re-ships the existing `rendered_html`) and never
re-renders the sections. R3a confirmed this is exactly what happened on 2026-06-24 15:06: the pre-regen schema
snapshot is the **old** key set (matching `content_data` field-for-field) and the live schema is the **new** set,
so the regen renamed the fields and left `content_data` unmigrated. (It was a regeneration rename — not an
out-of-band edit or a standing writer-vs-schema mismatch.)

Because `markPagesPendingRebuild` only sets `build_status`/`updated_at` (the `15:06:12.956` stamp) and the
triggered rerender is assemble-only, the rows were not repaired in place. The content-presence check (R3a)
confirmed the current state: `content_data` is **per-page** (distinct headings/stats) yet the stored render
contains none of it and all five are byte-identical — so the rows hold the **new-template render with empty
content slots**. The five are **live-broken** (empty system-stats band), not latent. gamesdesign was a sixth,
since dropped. `content_data` is intact and per-page, so recovery is a key re-map plus a section re-render that
**must carry `reason=section_data_resolved`** (a bare `needs_rerender` is assemble-only and just re-ships the
shell). The remaining steps fix the cause permanently, test it, recover the five, and verify.

## Where we are now (2026-07-03)

**Fix set deployed + proven; the five broken pages are RECOVERED (R6b pass 2026-07-03).** F1 guard live and
observed rejecting; F1-prompt (loader + rule + function pin) live and observed preserving; F3 scoped reason
propagation proven end-to-end; readiness unblocked via real per-site `cta` specs; all five dependents now render
per-page content with the CTA baked. Remaining: R6 (eyeball the live deploys — URLs in the checklist), then the
flagged follow-ups: F4 (fork advisory), F5 (guard extension for regen-ADDED required fields), F6 (guard/index
status-list alignment + `itemsCreated` overcount), stale `unresolved` item hygiene (40 rows, 2026-05-01), and the
loose dispatch item-status semantics (five sightings). Backup tables can be dropped after R6:
`DROP TABLE page_components_bak_sysstats_20260702;` (keep a few days first if preferred).

## Progress at a glance

Tick these as you go (edit the file: `[ ]` → `[x]`).

- [x] **R0** — Orient (read this + the handoff)
- [x] **R1** — Inspect the component (schema fields, template placeholders, version history)
- [x] **R2** — Enumerate affected pages + resolve site domains
- [x] **R3** — Confirm the bind failure on one page (keys don't intersect)
- [x] **R3a** — When/why the binding broke — RESOLVED: 15:06 regen rename, `content_data` not migrated; five are **live-broken** (content-presence check)
- [x] **F1** — Permanent fix: field-contract guard in `StoreGeneratedComponentAction` — deployed (same image as the load action); its live firing is proven by F2 Tier 3b
- [x] **F1-prompt 1a** — `load_existing_component_action.go` deployed + registered (`IsLocal: true`); validator passes, loader step executes on live pods
- [x] **F1-prompt 1b+2** — migration applied (verify row: step wired, `existing_component` in input_fields, rule present); dead `existing_field_names` block cleaned (snapshot-first)
- [x] **F2** — Test: Tier 1 unit ✔ · Tier 3a keep ✔ (two regens, `{body,eyebrow,heading}` preserved, templates genuinely changed — md5) · Tier 3b reject ✔ (2026-07-02 16:39 — guard fired live, named `zzz_alpha/beta/gamma`, fixture untouched, 0 version rows; rejection visible at 3 log levels)
- [x] **F2-cleanup** — done 2026-07-02 (item _3 + fixture removed; all counts reconciled)
- [x] **F1-prompt pin** — `F1prompt2_pin_function.sql` applied (`pin_present=t`); its live rendering piggybacks on the next active-path regen (mechanism identical to the field-names injection observed in 3a)
- [ ] **F4 (flagged)** — store-side advisory when the function lookup misses but an active same-`section_type` component exists (parallel-duplicate visibility) — follow-up, not built
- [x] **F3** — COMPLETE (2026-07-02): code applied in repo ✔ (byte-verified) · F3c config applied ✔ (all 5 keys) · fleet on **v1.0.1088** (rerender-pages, component-creator, page-rerender). Note: F3c was re-run once from the unfiltered SQL — idempotent, one extra snapshot, config unchanged
- [ ] **R4** — Back up before any write
- [x] **R5 (C0–C2)** — Backup ✔, re-key ✔ (UPDATE 5, 20 new keys), triggers ✔ (all five scoped items created + processed — **F3 proven end-to-end**); renders still carried (see R5b)
- [x] **R5b** — RECOVERED 5/5, verified 2026-07-03 (R6b pass: distinct md5s, needles true, cta baked; the two stalled children self-resolved — ai-agent-orch ran 13:42–13:45, gripper-detail 13:44). vonc.com/index = healthy 6th dependent
- [ ] **F6 (flagged)** — align the store guard's NOT EXISTS status list with the dedup index predicate (`unresolved` is index-terminal but guard-blocking); pair with the `itemsCreated`-not-gated-on-RowsAffected overcount fix in `create_rerender_items`
- [ ] **R6e — F8 Steps 1–3 EXECUTED 2026-07-03** (snapshot v2 ✔ · neutralize ✔ · verify ✔ stats=llm/CTAs neutral · backup 3 ✔ · strip UPDATE 3 ✔ · scoped triggers INSERT 0 2 ✔ · other chat informed). A–D READ 2026-07-04: matchmatrix was NOT skipped — fanned out and ERRORED (error text to fetch); auto-escalations
  fired (needs_page for matchmatrix + gripper-selection-guide); writer rebuilt gripper-selection-guide 18:04 with
  the IDENTICAL vonc pitch as idea.uk ("Every day a new Gauntlet begins" etc.) — AFTER fallback neutralization ⇒
  suspected THIRD F8 carrier: generation guidance (llm-field descriptions + component description, untouched by the
  neutralization). Decider query issued; if confirmed → snapshot v3 → neutralize descriptions → writer re-passes
  (gripper-selection-guide trivial; idea.uk coordinated; matchmatrix rebuild unproven — verify). History:  (snapshot v2 ✔, neutralize ✔ UPDATE 1, lock held). Step 3 in flight: backup+strip 3 rows → 2 scoped rerenders → gripper-selection-guide auto-escalates to needs_page (clone its shape for how-it-works) → post-verify gauntletised=f except vonc. Original:  (fallbacks carry the exact gauntlet strings; v1 = old-architecture, not restorable). EXECUTING: snapshot v2 → neutralize 8 fields (stats→llm optional, CTAs→neutral statics; optimistic-lock on updated_at 13:22:44) → strips + scoped rerenders (vonc untouched; idea.uk coordinate-gated; robot-hands index strip+rerender, other two pages → needs_page). Original line:  Live template is CLEAN
  (`template_has_gauntlet=f`); gauntlet strings hypothesized in static-field FALLBACK VALUES — merged into
  dependents' content_data by the stored⊕resolved merge (robot-hands index gained stat/cta keys at 12:59) and
  consumed by NEW builds (idea.uk, TODAY, gauntletised). Component updated 13:22 with NO version row (provenance
  open). Decider: the jsonb_each fallback query. Harm-stop: manual snapshot → neutralize fallbacks (coordinate —
  vonc's copy already in its own content_data) → scoped F3 re-renders (robot-hands index + latent how-it-works/
  gripper-selection-guide + idea.uk ×2). Old line for history:  Shared `brief-explanation` gauntlet-ified
  2026-07-01 12:46 (pre-guard, vonc-work window; static gauntlet copy + href="#") — our 12:59 all-sections pass
  stamped it onto robot-hands index (F3 blast radius; same bug family as system-stats). Playbook: check
  `component_versions` for a pre-07-01 snapshot → restore/repair the shared template (vonc's gauntlet copy belongs
  in vonc content_data or a fork) → scoped `needs_rerender` per affected site. **Coordinate with the other chat.**
- [ ] **R6f — vocabulary drift; OWNER LOCATED: webdesign-agent renders styles.css (storage_actions writes it; fix_harcoded_colours = post-pass precedent).** Missing consumed vars: --section-text/-muted/-surface/-border, --spacing-section, --border-radius, --color-heading, --color-white, --container-max-width (vs defined --container-max). Structural fix in the design generation or a palette-mapped post-pass. Original line:  styles.css defines the vars (:root ×2, --color- ×60);
  current-generator heads are rootless by design (vonc's too — R6d refresh idea dead). Diff the :root-defined names
  against the product-* templates' `var(--…)` names; the gap is why gripper-detail renders dark-on-dark while index
  bands (lighter fallbacks) stay visible. Fix lands wherever the gap is: styles.css/theme generation or template
  var-name conventions.
- [ ] **F8 (flagged)** — the field-contract guard checks NAMES, not fallback CONTENT: site-specific copy baked
  into a shared schema's fallback values sails through and contaminates every dependent via the resolved-merge.
  Candidate mitigations: prompt rule (fallbacks must be site-neutral), and/or store-time lint on fallback strings.
- [ ] **F7 (re-scoped)** — snapshot INSERT in `update_component_html` is ALREADY FIXED in current code (versioned, correct columns). Hero's 16:43 zero-version update likely = `IncrementUsageCount` (verify: usage_count/tpl-md5 query). Residual: the action swaps templates without placeholder⇄schema sync validation — narrower follow-up
- [ ] **F5 (flagged)** — extend the guard: a regen ADDING a required, fallback-less field is also a contract change (reject / force optional)
- [ ] **R6 (last recovery step)** — Eyeball the system-stats band on the LIVE deploys (deploy already ran inside page-rerender). ✔ leopardessconsulting.co.uk/index.html CONFIRMED 2026-07-03 (eyebrow "BY THE NUMBERS", headline, 4 stat cards, footnote, "View full report" → /contact.html). Remaining four — get the REAL paths from pages.url (do not guess paths):
  ```sql
  SELECT s.domain, p.name, p.url FROM pages p JOIN sites s ON s.id = p.site_id
  WHERE (s.domain, p.name) IN (('robot-hands.com','index'),('robot-hands.com','gripper-detail'),
        ('ai-agent-orchestration.com','index'),('ai-agent-orchestration.com','case-study-kafka-consumer-group-remediation'))
  ORDER BY 1, 2;
  ```
  then load `https://<domain><url>` for each — same band, per-site numbers (robot-hands: 2400 / "Gripper Models Indexed"). A "band" = one full-width page section = one `page_components` row's render
  LEDGER 2026-07-03: leopardess/index ✔ · robot-hands/index stats band ✔ (page carries an UNRELATED gauntlet band
  belonging to vonc.com — other chat's active area; evidence query below, handed off, not chased here) ·
  ai-agent/index mostly ok per user (case-study images render as alt-text + one mis-sized component — image
  pipeline territory, parked) · ai-agent/case-study page still to eyeball · gripper-detail → R6c.
  ```sql
  -- evidence for the gauntlet-on-robot-hands contamination (hand to the other chat)
  SELECT cc.function, pc.slot_name, pc.build_status, pc.updated_at, length(pc.rendered_html) AS len
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
  WHERE s.domain = 'robot-hands.com' AND p.name = 'index' ORDER BY pc.updated_at DESC;
  ```
- [ ] **R6c — gripper-detail: DB whole, ARTIFACT mis-assembled (pinned 2026-07-03).** All 8 sections rendered
  healthy in `page_components` (13:44:31, ~49KB). The 62KB live artifact contains only **4 of 8 bands**
  (product-specs duplicated; hero/details/card/text-block MISSING), **`:root` = 0** (no theme variables → dark
  theme renders blank), and the pages row shows `sections_listed = 0`, `rendered_header`/`rendered_footer` NULL
  (`deployed_at` set, `last_built_at` not). ⇒ the defect is the ASSEMBLY layer building from an empty `sections`
  list with no head/header/footer. RELATED: robot-hands **index** has NO gauntlet row in `page_components` (seven
  clean sections) yet the live page shows a gauntlet band → the foreign band is stitched at assembly/deploy time —
  same defect class, NOT a page_components contamination (earlier hand-off-to-other-chat framing revised; both
  artifacts came from our rerender-triggered assemblies at 12:59/13:44 with vonc building concurrently at 13:15).
  Next reads: comparative pages rows incl. `rendered_head` (healthy vs broken: is `sections_listed>0` + head
  non-null the discriminator?); artifact `data-component` inventory + `:root` for robot-hands index and the healthy
  leopardess control; then the assemble action source
  (`grep -rn "rerender_single_page|\"render_page\"|rendered_head|rendered_header" platform/orchestration/actions/`)
  MECHANISM MAP (from code, 2026-07-03): assembly membership = page_components by position with a >10-visible-chars
  filter (pages.sections jsonb is NOT used for membership — metadata only); head/header/footer = site_components,
  else a 5-line default head linking /assets/css/styles.css; the gripper-detail artifact's inline base-styles head
  matches the OLD assemble_from_library builder ⇒ likely a STALE deploy (check repo file + Actions for the 13:45
  commit + cache-busted curl). robot-hands gauntlet-in-hero candidate: shared 'hero' template altered during the
  vonc work (update_component_html has no field guard) + our 12:59 all-sections re-render (F3 blast radius) —
  check strpos(hero,'Gauntlet') + hero component_versions. Theme blank: robot-hands has no site head row
  (leopardess does) — verify site_components + that /assets/css/styles.css exists live. index 6/7: info-card-grid
  possibly dropped by the visible-text filter — eyeball its stripped text.
  UPDATE 2026-07-03: **gripper-detail = STALE CACHE** — cb-curl returned the new 8-band artifact (product-hero ×57)
  while the plain URL served the old 4-band one ⇒ deploy shipped; close on hard refresh. Hero-contamination and
  missing-site-head theories FALSIFIED (hero row clean; robot-hands has an 8KB stored head). Gauntlet: two
  cache-testable candidates — stale cached index vs JS-injected UI (script-mounted panels don't appear in
  data-component greps); cb-fetch both pages + script-src listing + does the stored head contain :root.
  UPDATE 2026-07-03 (later): md5-identical fetches ⇒ ONE artifact; the mis-assembly and stale-cache readings were
  metric artifacts (data-component attrs only on some templates). Blank = robot-hands head lacks `:root`
  (leopardess' has it) ⇒ per-component fallback colors, dark-on-dark. **R6d fix path** (pending vonc-head sample +
  styles.css grep): CTAS-backup robot-hands' 3 `site_components` rows → `needs_rerender` with
  `refresh_site_components: true` → head regenerated by the current generator → pages re-assemble themed.
  New F4 evidence: TWO non-forked `hero` rows (2026-03-09 + 2026-07-02 16:43, synced=f, no version row).
  **F7 flagged:** shared `hero` updated 2026-07-02 16:43 with ZERO version rows — the `update_component_html`
  signature (no field guard; snapshot INSERT silently fails on the missing `version_note` column). Unguarded,
  unversioned write path to shared components — repair the snapshot + consider extending the guard there.

**R6a — scoped items, each carrying the reason:**
```sql
SELECT s.domain, w.status, w.spec->>'page_name' AS page, w.spec->>'reason' AS reason
FROM site_work_items w JOIN sites s ON s.id = w.site_id
WHERE w.item_type='page_rerender' AND w.spec->>'reason'='section_data_resolved'
ORDER BY s.domain, page;
-- expect exactly the dependent pages per site, no others
```

**R6b — renders regained content (distinctive needles only):**
```sql
SELECT s.domain, p.name, pc.build_status, length(pc.rendered_html) AS len, md5(pc.rendered_html) AS md5, pc.updated_at,
       strpos(pc.rendered_html, coalesce(nullif(pc.content_data->>'stat1_label',''),chr(1)))>0      AS has_stat1_label,
       strpos(pc.rendered_html, coalesce(nullif(pc.content_data->>'section_headline',''),chr(1)))>0 AS has_headline,
       strpos(pc.rendered_html, 'href=""')>0 AS cta_href_empty,
       strpos(pc.rendered_html, coalesce(nullif(pc.content_data->>'cta_url',''),chr(1)))>0 AS cta_url_baked
FROM page_components pc JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
WHERE pc.component_id='fdd92ad4-521a-4602-89cf-7ee1a66c10f1'
ORDER BY s.domain, p.name;
-- pass: has_stat1_label + has_headline TRUE on all rows, DISTINCT md5 per page, md5 ≠ the stale render,
-- build_status transitioned, cta_href_empty FALSE (TRUE → Tier-C follow-up, not a recovery failure)
```

Then confirm the bands on the live deploys (git/Backblaze).

## Step R7 — Freshness / concurrency check (immediately before any write)

Another chat co-manages components. Just before R4/R5 (and any structural fix), re-check that nothing moved since
R1/R2:

```sql
SELECT id, updated_at, schema_template_synced, schema_field_count
FROM content_components WHERE id = 'fdd92ad4-521a-4602-89cf-7ee1a66c10f1';

SELECT max(updated_at) AS latest_pagecomp_update, count(*) AS instances
FROM page_components WHERE component_id = 'fdd92ad4-521a-4602-89cf-7ee1a66c10f1';

SELECT version_number, changed_by, created_at
FROM component_versions WHERE component_id = 'fdd92ad4-521a-4602-89cf-7ee1a66c10f1'
ORDER BY version_number DESC LIMIT 3;
```

If the component's `updated_at`, the instance count, or the version history differ from what you recorded in
R1/R2, stop and re-inspect — the component may have been changed underneath you. No blind writes.
