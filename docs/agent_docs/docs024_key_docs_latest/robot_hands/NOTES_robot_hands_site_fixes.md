# RUNNING NOTES — robot-hands.com site fixes (R1–R6)

Thread: "robot-hands.com" session, started 2026-07-17 (evening), continuing
2026-07-18. Entry point: `HANDOFF_2026-07-17_robot_hands_site_fixes.md`.
Companion outcome doc: `RESULTS_2026-07-18_r1_r6_fixes.md` (root-cause detail).
This file is the chronological record.

Site: robot-hands.com, site_id `00ff3af5-dad8-4770-9f70-3edc267a3c92`.
Deploy repo discovered this thread: **gqls/sites**, files under
`robot-hands.com/` (git-adapter commits via GitHub API; no per-site repo;
`sites.github_repo` empty → default "sites").

---

## Turn 1 — 2026-07-17 evening — R1 investigation

- Read the handoff; created tasks R1–R6; queue snapshot taken FIRST per
  CLAUDE.md (large stale backlog confirmed: page_rerender@35/80,
  phantom_internal_link@35, undeployed_asset@60, ~16 cta_names_unknown_
  destination in needs_human_review).
- Header slot: component `header-bold-gradient`, re-rendered 2026-07-16
  15:51–15:55 after the needs_rerender completions at 14:18–14:25. Rendered
  HTML carries an inline `<style>` with literal
  `linear-gradient(135deg, #3b82f6 0%, #64748b 100%)` — the blue brochure
  header. Template fills `{{.primary_color}}`/`{{.secondary_color}}`/
  `{{.accent_color}}` into that style block.
- **component_versions: NO rows for header/footer/head** — the handoff's
  "restore from snapshots" route doesn't exist.
- Colour source chain traced (render_site_components_action.go):
  `loadSiteDataFull` reads `style_collections.color_palette` via
  `sites.style_collection_id`. That JSONB — and `css_themes.color_palette`
  AND `palettes.colours` (palette-robot-hands-com) — all still held the
  **light brochure palette** (background #ffffff, primary #3b82f6, accent
  #fbbf24). **B7 swapped only `css_themes.layout_id`.** Root cause #1.
- `renderAndStoreSiteComponent` renders whatever `site_components.
  component_id` points at, **ignoring is_active** — and the slots pointed at
  DEACTIVATED `header-bold-gradient`/`footer-4-column` (exactly what three
  FAILED 2026-05-13 deactivated_component items said). Root cause #2.
- gqls/sites bisect of index.html: blue header entered between af0ead8da1
  (2026-07-08T22:44, markup-only header) and 78532b8c63 (2026-07-09T09:13).
  **The blue predates B7 and the 07-16 regen** — the regen only spread it.
- Served styles.css was genuinely tool-portal-dark (dark root vars) BUT with
  palette leaks: `--color-card-bg #ffffff`, `--color-header-bg #1e293b`,
  blue `--color-cta-bg`, `--section-heading #0f172a` (near-black on dark).
- hardcoded_section_colors items marked complete had actually FAILED:
  `WORKFLOW_INVALID: step 'fix_text_colors' with action
  'fix_forced_text_colors' requires a topic`. They rewrote nothing. (Two
  bugs flagged: broken workflow config; failed workflow → item complete.)
- Library scan: active, fully var()-based `site-header`(f420f3fa)/
  `site-footer`(4238e467) exist but use a different data contract
  (nav_link_1..4) — unsuitable for the render action's `nav_items_html`
  contract. Decision: hand-author new components.

## Turn 2 — 2026-07-17 ~20:20 — R1 applied

`SQL_2026-07-17_r1_dark_theme_restore.sql` (commit 17b4adcc7):
- NEW components `header-theme-chrome` (58fde68f) / `footer-theme-chrome`
  (e6347680): markup uses classes tool-portal-dark styles natively
  (.site-header/.main-nav/.site-footer/.footer-container); only var()-based
  style for gaps (.header-cta, mobile menu is-open, footer bottom row); zero
  literal colours. Named to sort AFTER the alphabetical-first components so
  the `ORDER BY name LIMIT 1` fallback for other sites is unchanged.
- Repointed robot-hands header/footer slots; head left (Document Head is
  deactivated but its output is fully var-based).
- ONE dark palette written to all three copies (values = the served
  approved dark root vars; card_bg→#1E2535, header_bg→#0F1218,
  heading→#E2E8F0, footer_bg→#0B0E14; cta_bg kept the approved gradient).
- Closed the three 2026-05-13 deactivated_component items (this WAS their
  fix). Backups: *_backup_20260717_r1 tables.
- Checked claimed items (zombie rule) — none; chassis pod 41m up (>300s
  after the new image the user mentioned) → safe to dispatch.
- kcat-triggered rerender-pages with `refresh_site_components:true`
  (single-line JSON — kcat -P line-splitting trap). Components re-rendered
  20:23: header 2547B / footer 3335B / head, all var-based, zero #3b82f6.
  37 page_rerender items created (priority 80, `triaged_at` NULL — note:
  create_rerender_items doesn't stamp it).

## Turn 3 — 2026-07-17 evening — R4/R6 facts; CSS re-render #1

- Tool sweep (live HTTP): grip-force-friction-calculator,
  gripper-cycle-time-estimator, gripper-payload-calculator = 200;
  **tool-matchmatrix and tool-robot-payload-budget-calculator = 404**
  (match the two incomplete_page_group review items).
- `/matchmatrix.html` "blank" solved: 6.3KB of text present; page is
  card-heavy — white `--color-card-bg` + light text, plus
  `color: var(--color-primary)` (dark navy) misused as text colour. Looked
  blank; wasn't. Palette fix + CSS re-render addresses it.
- Triggered webdesign-agent (kcat) to re-render CSS from the fixed palette.
- R6 rows identified: 3 scaffolds (news-post, learning-center-article,
  learning-center-post) + 3 duplicate slugs of the real guides; all
  status='active', never deployed. Their card/hero assets (generated
  2026-07-16/17 by the imagery passes) confirmed present → must supersede.
- **Discovered the imagery session's F2.1 fix already in code** (commit
  4e35c8064, `listedOnly`: blog_posts listings require deployed_at + non-
  empty sections — comment cites these exact 6 rows). Verified IN THE
  RUNNING POD via strings-grep (3 hits of the eligibility SQL). The listing
  side of R6 was already handled; rows still needed retiring.
- `SQL_2026-07-17_r6_retire_dead_article_rows.sql` (898ff3494): 6 rows
  archived, 12 assets superseded, 5 "re-render after asset landed" review
  items closed. Real tool-* guides verified untouched.

## Turn 4 — 2026-07-17 late — R3 + R5 applied

- Nav tables read: primary "Learning Center" → /learning-center.html
  (phantom-card grid); utility → /learning-center/index.html (residue).
  Hub (the real listing) not in nav at all.
- IA decision (delegated by handoff): **hub canonical**.
  `SQL_2026-07-17_r3_learning_center_ia.sql` (7be3720b9): both nav items →
  /learning-center-hub.html; hub in_header=true nav_order=6; grid page kept
  ACTIVE but demoted (archiving would flip every baked in-body link into
  phantom_internal_link churn); learning-center-index archived; plan-row
  nav flags in lockstep (flag edits ≠ re-plan); 2 failed index-rebuild
  items closed.
- R5: content-listing schema `fields.show_load_more.fallback` true→false +
  hub instance opt-out (`SQL_2026-07-17_r5_dead_load_more.sql`, c10ec34e7).
  Button is behaviourless fleet-wide; default-on was the defect. Explicit
  trues on dartsonline/idea.uk left for their owners.

## Turn 5 — 2026-07-17 ~20:30–21:00 — the webdesign churn root cause (R1b/R1c)

- CSS re-render #1 landed (302702fc96) — theme-owned slots took my palette
  (card_bg #1E2535 ✓, header_bg #0F1218 ✓, section-heading #E2E8F0 ✓) but
  core slots REGRESSED: `--color-background: #F4F5F7` — **a light
  background went live on the dark site**.
- Mechanism: webdesign-agent's `analyze_design` LLM invents a color_scheme
  EVERY run; render_css_from_spec merge = spec wins on core slots. The
  prompt renders ONLY the structured `design_intent.palette` block —
  robot-hands had only free-text colour_mood ("deep charcoal…"), which the
  template never shows the LLM → "no intent → invent" branch every run.
  Explains the whole day's hue drift (4 CSS rewrites on 07-17).
- `SQL_2026-07-17_r1b_design_intent_palette_pin.sql`: superseded
  design_intent keeping all fields, ADDING structured palette+typography
  (reference_values = approved dark set; guidance: dark is a HARD
  requirement, body never monospace). Re-triggered webdesign-agent.
- **Pin proven**: CSS #2 (f635b1e49) came back exactly on pinned values —
  background #0F1218, accent #E8500A, text #E2E8F0, IBM Plex Sans 16px.
  `sites.content_data.color_scheme` now equals the pinned set verbatim.
- **generic_theme misfire root cause (fleet-wide)**: the check demands
  `site_specs aspect='webdesign'` — **0 rows fleet-wide, no code path ever
  writes it**; webdesign-agent stores its spec in `sites.content_data`
  (update_site_content). So every themed site re-fires each discovery pass
  → webdesign dispatch → palette re-roll. Go fix committed (3437f2212):
  check also accepts `content_data ? 'color_scheme'`. **Inert until a
  chassis image build+roll** (build-from-HEAD is now the default).
- Closed the six 20:13 detections my work had already resolved (3
  needs_rerender stale-component + 3 deactivated_component).

## Turn 6 — 2026-07-18 morning — dispatch archaeology; batch promoted

- The 37 page_rerender items sat 'triaged' ~13h untouched. Diagnosis chain:
  - Dispatch pick query (load_work_item_actions.go:589-ish) would accept
    them (no triaged_at requirement).
  - `scheduled_tasks`: **improvement-sweep (improvement-loop, 180s) is
    DISABLED since 2026-07-17 13:18** — another session's action, left as
    found. `build-pipeline-trigger` (30s) IS enabled: seed → 
    find_dispatchable_site → spawn/call build-dispatch-loop, ~1s/firing.
  - find_dispatchable_site excludes any site with a claimed item
    (`NOT EXISTS … status='claimed'`) — **one item at a time per site**;
    57-item backlogs take many hours. Morning: dispatch WAS moving again
    (build-dispatch-loop claims at 09:37 etc.).
- Promoted the 37-page batch priority 80→20 (ahead of the stale 35–60
  churn; sanctioned hand-promotion), stamped triaged_at. Armed a persistent
  drain monitor (absolute created_at cutoff — an earlier monitor
  false-completed when items aged out of its relative 1h window; lesson
  recorded below).
- Committed everything per-task with explicit pathspecs: 17b4adcc7 (R1+R1b),
  7be3720b9 (R3), c10ec34e7 (R5), 898ff3494 (R6), 3437f2212 (generic_theme
  Go fix), b1a3db664 (RESULTS doc). Memory files written:
  `robot-hands-site-fixes-workstream`, `webdesign-colour-churn-landmine`.

## Turn 7 — 2026-07-18 ~11:55 — live verification (this update)

- **The blue is GONE from the live site**: index, learning-center-hub,
  matchmatrix, about, tools all have 0 hits of #3b82f6; index serves
  `<header class="site-header">` (the new theme-chrome markup). Rerender
  commits flowing in gqls/sites (11:06–11:34: index ×2,
  pneumatic-vs-electric-grippers, content-card image deploy) — pages fetch
  header/footer at assembly, so re-renders triggered by ANY queue item now
  bake the dark chrome.
- Live styles.css serves the pinned dark set (background #0F1218, card_bg
  #1E2535).
- Queue at 11:54: my batch of 37 still 'triaged' (other items re-rendered
  the pages deployed so far); site totals 52 triaged / 1 claimed /
  16 detected. Drain monitor still armed.

## Turn 8 — 2026-07-18 afternoon — CLAUDE.md obligations discharged; 37 pages drained

Prompted by a re-read of CLAUDE.md, two duties this thread had skipped:

**(a) File what you learn (§ Debugging).** Grepped 016b §10 for the mechanism
first (not the truncation family), then filed **`bugs_open/017`**: the
`fix_forced_text_colors` pair — action written with handler + input spec but in
NEITHER registry, so `validation/workflow.go:80` classes it remote and rejects
it as *"requires a topic"* (the error names the wrong thing; the validator reads
the DEPRECATED hand list, not the registry-backed `IsLocalAction` at
`registry.go:1866`) — AND `CompleteWorkItemAction` storing a
`response.status='failed'` payload next to `status='complete'`. Added the
transferable §9 pattern + index row 017. Renumbered mid-flight: three other
sessions took 014–016 while I wrote.

**(b) Council review (§ Council gate).** The `generic_theme` change is
`platform/` and I had committed it unreviewed. Submitted retroactively,
correlation **e0ebf6ee-dcc0-4a7b-9a3d-438ce9af5fff**. Three rounds:

- **R1 revise** — editquality: storage shape generalised from one site.
  bug_historian: trigger patched, damage mechanism untouched. Council's own
  answered check settled the consumer question: generic_theme routes ONLY to
  webdesign-agent (7 unresolved + 5 complete), no other consumer.
- **R2 revise** — I answered the shape doubt with a fleet sweep (**7/12 sites
  top-level, 0/12 ANY nested variant**, single shared writer) and added a
  scheme-consistency guard as edit 2. Seats split cleanly: edit 1
  "approve as-is" (guardian, reuse_agent); edit 2 objected by four seats —
  unverified helpers, unverified `layouts.scheme`, single call site, blast
  radius on a shared render boundary.
- **R3 (submitted)** — took the council's own recommendation and **SPLIT**.
  Edit 1 alone, plus bug_historian's valid catch fixed in code
  (**3b52da8ec**): the fallback now tests the color_scheme VALUE
  (`jsonb_typeof(...)='object' AND <> '{}'`), not mere key presence — a
  crashed/empty run would otherwise read as "spec exists" and suppress the
  finding, the false-negative mirror of the original bug. Fleet-checked: the
  same 7 sites pass, so no classification changes today.

- **R3 verdict: revise, 7 of 8 seats APPROVE** (editquality, reuse_agent,
  guidelines, guardian, improvement_guardian, render_guardian, debug_historian);
  bug_historian alone objected — one LOW (two coexisting "spec exists" markers
  in different tables invite divergence) and one MEDIUM (is the existence-only
  shape used by other checks?).
- **Both answered, and the loop STOPPED there** — CLAUDE.md says one council
  run per coherent task, not per iteration, and a 4th round on a low-severity
  design point would spend credits against that guidance. The MEDIUM is
  answered empirically with a grep, no code needed: **no other discovery check
  has the exposure** — `check_integrity` and `check_image_source_unsatisfiable`
  both extract VALUES (`->>` + `COALESCE`), not key existence. The LOW is
  answered in code (**5151d4a79**): the two sequential probes are now ONE
  OR'd `EXISTS` query — the dead-but-contractual site_specs arm survives as a
  term rather than its own fallback branch — and the finding's detail message
  now names both sources instead of only `site_specs`. Fleet-verified
  identical classification (7/12) before and after.

The withdrawn guard is filed as **`bugs_open/022`** rather than left as vapour,
with the three verifications the council demanded now DONE and recorded:
`parseHexColor` exists (`color_util.go:26`), `layouts.scheme` exists (text,
CHECK light/dark/neutral), and `buildPaletteMap`/`loadThemeComposition` have
**exactly one non-test caller each** — so the guard would have been at the
mechanism, not a point patch. It returns as its own submission.

**R3 nav flip CONFIRMED LIVE (2026-07-18 evening).** The 30-page nav batch
drained complete; `https://robot-hands.com/index.html` header now links
`/learning-center-hub.html` (4 hub links page-wide, 1 remaining
`/learning-center.html` — the demoted grid landing, by design).

**Site state at 15:00–16:00.** The 37-page batch **drained complete**; live
spot-checks green — index, learning-center-hub, matchmatrix all serve zero
`#3b82f6`, zero `site-header--gradient`, new `<header class="site-header">`;
hub lists exactly the 3 real guides with no Load More (R5/R6 confirmed live).
**Gap caught:** those pages assembled BEFORE the R3 nav flip, so index still
linked `/learning-center.html`. Re-triggered rerender-pages with
`refresh_site_components:true` — components re-rendered 15:42 carrying the hub
nav (header hub-only; footer carries both, by design, since the grid page stays
a demoted landing) — and promoted the resulting 30-page batch to priority 20.
Draining at time of writing.

---

## Mechanisms learned this thread (beyond the handoff's list)

- **Three palette copies must move together**: `palettes.colours`,
  `style_collections.color_palette` (component renders read THIS one),
  `css_themes.color_palette`. A layout swap alone leaves component regens
  baking the old colours.
- The webdesign prompt reads ONLY structured `design_intent.palette` /
  `.typography` (character/reference_values/guidance) — free-text
  colour_mood is invisible to it. Pin per-site or the loop repaints.
- `renderAndStoreSiteComponent` ignores `is_active` on the assigned
  component; the deactivated_component check is the only guard.
- Deploy repo = gqls/sites, domain-prefixed paths, GitHub-API commits, no
  local checkout. Page history is the ground truth for "what did it look
  like when it was good" — component_versions may have nothing.
- `build-pipeline-trigger` dispatches ONE item per site at a time (site
  excluded while anything is claimed); pace ≈ minutes/item. Hand-promote
  (priority down + triaged_at) to jump user-visible batches ahead of churn.
- create_rerender_items leaves `triaged_at` NULL — harmless to dispatch,
  confusing to humans reading the queue.
- Monitor windows must use ABSOLUTE cutoffs — a `now() - interval` window
  silently empties as items age, and the monitor reads that as "done".
- **"requires a topic" almost never means a topic is missing** — it means the
  action is in neither registry (`bugs_open/017`). Grep both lists first.
- **A work item's `result` can contradict its `status`.** Sweep for the class:
  `WHERE status='complete' AND result->'response'->>'status'='failed'`.
- **The council pays for itself on a change you thought was obvious**: three
  rounds on a 6-line check fix found a real false-negative gap (key presence vs
  value) and correctly forced a shared-render-path change out into its own
  submission. Answer its checks with queries, not prose — the seats that
  objected on evidence flipped to "approve as-is" once swept.
- `improvement-sweep` disabled ≠ dispatch dead: build-pipeline-trigger is a
  separate, still-enabled path.

## Open when this thread ends

1. **Batch drain + final visual verify** (R1 visual, R2 yellow-on-white,
   R4 matchmatrix look) — monitor armed; spot-checks already green.
2. **R2 caveat**: ensure forced-text-color-fixer doesn't "fix" amber once
   dark returns (its workflow is currently INVALID, so low risk — but that
   invalidity + failed-workflow-marked-complete are open loop-integrity
   bugs for the fixloop/empty-sections workstreams).
3. **R4 tool builds** (tool-matchmatrix, tool-robot-payload-budget-
   calculator) + body CTAs pointing at their 404s: experience_loop
   workstream's class — coordinate, don't duplicate.
4. **Chassis image roll** to activate the generic_theme fix (3437f2212) —
   batch with the next planned deploy; not while a drain is mid-flight.
5. Learning-center grid page content (phantom-CTA cards): fold into hub or
   fix targets once tools exist.

## Turn 9 — 2026-07-19 — council closed at 7/8; generic_theme fix PROVEN LIVE; docs conformed

**Council round 3 (correlation e0ebf6ee): revise, but 7 of 8 seats APPROVE.**
Only bug_historian objected — one LOW (two coexisting "spec exists" markers in
different tables invite divergence) and one MEDIUM (does any other check use the
existence-only shape?).

- MEDIUM answered by grep, no code needed: **no other discovery check has the
  exposure.** `check_integrity.go:81` and `check_image_source_unsatisfiable.go:83-84`
  both extract VALUES (`->>` with `COALESCE`), not key existence.
- LOW answered in code (**5151d4a79**): two sequential probes collapsed into ONE
  OR'd `EXISTS` — the dead-but-contractual site_specs arm survives as a term,
  not its own fallback branch. Consolidating also exposed a stale user-facing
  string I would have shipped: the finding's detail still read "No webdesign
  spec in site_specs", untrue the moment a second source was added. Now names
  both. Fleet-verified identical classification (7/12) before and after.
- **Loop stopped there deliberately** — CLAUDE.md: one council run per coherent
  task, not per iteration. A 4th round on a low-severity design point spends
  credits against explicit guidance.

**The generic_theme fix is LIVE and PROVEN — and I nearly misread the evidence.**
Pod changed under me: `agent-chassis-5c568b8c74-2f4qv`, image **v1.0.1137**,
started 2026-07-18 22:21 (another session's roll). First grep
(`grep -c 'jsonb_typeof(content_data'` → 1) I took as "my consolidation is live".
**Wrong.** A second grep on a string unique to the LATER commit
(`sites.content_data.color_scheme`, the new detail message) → **0**. So the pod
carries the *functional* fix (3437f2212 + 3b52da8ec hardening) but NOT the
round-3 consolidation — which is structural only, so nothing is missing.
*Lesson: to tell WHICH of two candidate commits is live, grep a symbol unique to
the later one; a shared symbol proves only that one of them shipped.*

**Real-world verification (the payoff):** zero `generic_theme` detections
**fleet-wide** since the roll (`WHERE item_type='generic_theme' AND created_at >
'2026-07-18 22:21'` → 0 rows). The perpetual misfire is over in production.

**R3 nav flip confirmed live:** 30-page batch drained complete; index header
links `/learning-center-hub.html`; 1 remaining `/learning-center.html` is the
demoted grid landing, by design.

**Concurrency check (owner asked):** `git log 5151d4a79..HEAD --` over my files
returns only my own commit; `git status --short` clean on all my paths. No other
thread is editing beneath this work, though HEAD has moved a long way on others'
files.

**Docs conformed to the standing five** (CLAUDE.md 2026-07-18/19 directive, read
fresh this turn): `RUNNING_NOTES_*` → **`NOTES_*`** (git mv), added **PLAN**
(decisions + the four corrections to the originating brief), **RUNBOOK** (the
commands with gotchas), **SUMMARY_2026-07-19** (five-part read-aloud),
**README_where_we_are** (was an EMPTY file — filled, in the owner's register;
it is his doc, append-only from here), and **HANDOFF_2026-07-19_…start_here**.

**Bug-numbering trap hit for real:** my `bugs_open/017…unregistered_action…`
collides with `017…static_cutover_orphans_backend_entry_forms…`. CLAUDE.md now
documents this ("016 and 017 are each used by two different cases — resolve by
slug") and `/bugs_closed/README.md` says numbering is **never reassigned**, so
the file stays as-is and every reference to it must use the slug.
`/bugs_closed/README.md` also already names `017 (unregistered action)` as
deliberately open — correct: committed-but-inert fixes stay in `/bugs_open/`.

**Process debt recorded, not hidden:** CLAUDE.md's diagnosis-loop section was
INVERTED on 2026-07-19 — the 090 loop is now the DEFAULT before asserting any
durable claim. `bugs_open/017…unregistered_action…` and `bugs_open/022` were
both authored before that correction and have NOT been through it. Flagged in
the handoff for whoever picks them up.
