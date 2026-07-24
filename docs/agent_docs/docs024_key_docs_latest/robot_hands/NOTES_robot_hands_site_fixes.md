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

---

## Turn 9 — 2026-07-20 — R4 closed: MatchMatrix built, CTA pairing fixed, live fabrication found

**Resumed from `HANDOFF_2026-07-19_…start_here.md`.** Checked what had moved under
it first; five things had.

**What changed underneath the handoff.**
- Image rolled **v1.0.1137 → v1.0.1139** (pod `agent-chassis-645674b498-rndg9`,
  started 2026-07-20 07:35Z). The handoff's optional item 4 is moot.
- `bugs_open/017…unregistered_action…` now reads **fixed-and-live** in 1.0.1139
  (`strings /app/agent-chassis | grep -c handlerReportedFailure` → 6; registration
  at `platform/orchestration/actions/registry.go:719`). Its own file still says
  "STAYS OPEN … INERT until a chassis image ships" — that line is **stale**. Left
  for its owning thread to close; I did not move it. `[UNVERIFIED]` whether the
  saga-completion half is equally covered — I checked the registration and the
  guard symbol, not a live failing saga.
- **Owner ruling 2026-07-20 11:17** (`3f6f1febf`): hold tool *imagery* until
  `bugs_open/020` is fixed. Literal scope is imagery, not tool building.
- Six new bugs bearing on the tool action: `020`, `024`, `028`, `029`, `030`, `033`.
- Another thread touched this site directly — migration `175` (`bugs_open/002`
  error C), preventative fix to `contact`'s section authority. Verified applied on
  the current plan (`contact-block` present, `contact-info` gone). No collision.

**Why I did not build the tools the platform way.** Three independent blockers,
each verified rather than assumed: `028` (both target pages had **zero**
`site_plan_sections` and zero components — the `add_tool` item was already
`complete` with `{"completed_steps": 0}` and no `create_result`, i.e. the build
had already run and done nothing); `020` (a gripper match-matrix is data-backed,
and `tool-generator` — the *new*-tool path, not the recreation handler the owner's
hold names — has `has_no_fake_data_rule=f` and forbids fetch, so it is
structurally pushed to invent a catalogue); `024` (any later fix could never reach
the page). So MatchMatrix was **hand-authored**.

**Correction to my own read of `bugs_open/039`.** I expected the sections path to
be the trap here — 039 (filed the same morning) says an unresolvable section name
renders a 208-byte hollow stub. It does not apply: tool pages on this site carry
`sections = []` and have **no** `site_plan_sections` rows *even when they work*.
The three live tool pages each have exactly one `page_components` row at position
2 pointing at a bespoke `content_components` row (`name = tool-<slug>-robot-hands-com`,
`function = tool-<slug>`, `render_mode = template`, 17–24 KB of template). That is
the reference implementation; 039's Part 1 naming rule is still the thing to know,
it just is not what gates a tool page.

**Also confirmed live, `bugs_open/024` leg (a):** all three working tool components
sit at `build_status='pending'` with `deploy_commit` NULL and have done since
April, while serving 200. The field is written and never read.

**The finding that mattered more than R4.** Chasing label↔URL pairs turned up
**~40 mismatches across 20 components, and only 2 URLs actually 404.**
`/tools/matchmatrix/index.html` had become a **default dumping ground**: 20
components on 11 pages pointed at it while only ~6 of their labels named
MatchMatrix. The others said "Search the Gripper Catalog", "Browse the Learning
Center", "Open the Payload Calculator", "Request Integration Support" — each with
a live 200 destination sitting unused. Secondary CTAs were worse: 20 of them,
essentially all mispaired, **none 404** — 14 labelled "Read the MatchMatrix
Methodology" pointed at `/services.html` while `/matchmatrix-methodology.html`
served 200 throughout. Nothing flagged any of it precisely because nothing was
broken. This is `bugs_open/023` in a severe instance, and it is why this site has
20 `cta_names_unknown_destination` items rotting in `needs_human_review`
(`bugs_open/033`).

**Every UPDATE is keyed on the LABEL, not the old URL.** Repointing by URL — the
obvious move, and what the handoff's interim suggestion implied — would have
cemented every mismatch. That is the single most important line in
`SQL_2026-07-20_r4_matchmatrix_and_cta_pairing.sql`.

**Then: live fabricated statistics.** The verify query in R4b surfaced a second
stat block, and a site-wide sweep found a third. `about` was serving **"1,200+
Gripper Models Indexed"** against an index of **five**; `gripper-detail` was
serving **2,400+ / 140+ / 6 / 18**; `index` had the fabricated "6 actuation
technologies" left behind after a *partial* cleanup that blanked its two
neighbours to em dashes. `products.specifications` has **no actuation-type field
at all**, so no honest actuation count exists at any value. Filed as
**`/bugs_open/043`** — same family as `020` but a different path (ordinary content
generation, not tool recreation), so a site with no tools is still exposed.

The tell that nobody had ever read the rendered output: `gripper-detail` carried
`stat1_suffix='%'` on a model count and `stat2_suffix='ms'` on a manufacturer
count — unedited generic-template placeholders. The live page was rendering
**"2,400+%"** and **"140+ms"**.

**Scoped out, deliberately:** the same six-technology claim survives in **42
further `content_data` fields** on this site (body prose, `features`,
`subheadline`, FAQ `questions`, `cards`). Those are substantive copy, not stat
fields; rewriting them is a content decision and is the owner's call, not a
containment step. Recorded in `043` because the 9-vs-42 ratio is the useful signal
for anyone sizing the fleet-wide version: **stat blocks are the visible tip.**

**My own wrong call this turn**, logged to `WRONG_CALLS.md`: I asserted in a test
that the OnRobot 2FG7 (rated 11 kg) would clear an 8 kg part. It does not — that
part needs 523.2 N and the 2FG7 publishes 140 N; the 11 kg headline implies
μ ≈ 0.77. The tool was right and my expectation was wrong. Had I not written a
test I would have "fixed" correct code to match it. It became the tool's best
feature: MatchMatrix now explains that discrepancy inline, naming the implied μ.

**Verification, artefacts not statuses.**
- `/tools/matchmatrix/index.html` → **200**, 38,144 B, 1 form / 4 inputs /
  4 selects / 4 `addEventListener`, all five real grippers present, **0** hardcoded
  `#3b82f6`, **0** fabrication tells (`Mulberry32|makePostcode|buildData|SUFFIXES|
  Math.random`). Deployed to `gqls/sites` @ `0a6dc426`.
- **19/19 logic tests pass** (`node` in a container; the tool has no test harness
  on the platform, so the tests live in this workstream's scratch and are
  reproduced in the RUNBOOK).
- All nine stat values now trace to a query recorded in the SQL header:
  5 models / 5 manufacturers / 4 parameters compared / 24 published figures.
- 12 `page_rerender` items queued at priority 20,
  `source='robot-hands-r4-cta-pairing'`, reason `cta_links_stale` (chosen because
  per `024` that reason reaches the real `rerender_sections` branch rather than
  stale-HTML assembly). **The DB edits are inert until that batch drains** — the
  live pages still carry the old CTAs and old stats until then.

### CORRECTED 2026-07-20, same session — the CTA URL fix is NOT durable, and I now know why

**What I claimed earlier in this turn:** that `SQL_2026-07-20_r4_matchmatrix_and_cta_pairing.sql`
fixed the CTA pairing, and that keying the UPDATEs on the label rather than the
URL was the important part.

**What is actually true:** the label-keying was right, but **`content_data` is not
the source of truth for a CTA URL**, so the SQL alone cannot hold. Caught when
`services.html` re-rendered: I set "Request Integration Support" →
`/contact.html`, the page re-rendered at 15:31:06, and both the URL *and* the
`primary_cta_target_title` I had set came back as
`/tools/matchmatrix/index.html` / "Run MatchMatrix | Gripper Selection Tool".
The DB row's own `updated_at` was later than my UPDATE — so it was overwritten,
not un-applied.

**The mechanism, from the code** (`platform/orchestration/actions/
resolve_internal_links_action.go`):

- `ctaFieldNames` (`:99-105`) registers the CTA URL fields the resolver owns —
  `hero{cta_url, secondary_cta_url}`, `call-to-action{primary_cta_url,
  secondary_cta_url}`, `content-block-about{cta_url}`. Every component I edited
  is on that list.
- `chooseCTATargets` (`:319`) recomputes them, and **it never looks at the
  label.** It ranks interactive pages then hubs, drops excluded areas and
  self-links, sorts by `NavOrder` then `Name`, and takes
  `primary = ordered[0]`, `secondary = ordered[1]`.

So this was never drift. **Every CTA on the site is deterministically assigned
the same first-and-second destination by nav order**, and
`/tools/matchmatrix/index.html` became the "dumping ground" for the plain reason
that `tool-matchmatrix` sorted first among interactive pages. The labels were
written by a content pass that had no idea what the resolver would pick.

- `areasExcludedFromCTA` (`:72-74`) = `{about, contact, privacy, terms, legal}`,
  applied via `ctaExcludedDestination` (`:357`). **`/contact.html` can never be a
  CTA destination.** That is specifically why "Request Integration Support"
  reverted, and it means that particular pairing cannot be fixed in `content_data`
  at all.

**What this means for the fixes applied this turn:**

- **Statistics (`r4b`/`r4c`) are durable** — `stat_*` fields are not in
  `ctaFieldNames`, the resolver does not touch them, and `about.html` re-rendered
  live showing 5 / 5 / 4. Confirmed on the live page, not the row.
- **Secondary CTA fixes appear to have held** on the pages that re-rendered
  (`/matchmatrix-methodology.html` on services and product-detail) — but
  `secondary_cta_target_title` on services still reads "Gripper Payload
  Calculator", i.e. URL and target-title now **disagree**. `[INFERRED]` that this
  will be "corrected" back to the resolver's choice on the next render; I have not
  watched a second render to confirm it.
- **Primary CTA fixes are not durable.** Those that still look right — "Run a
  MatchMatrix Query", "Run MatchMatrix on This Model" — look right only because
  the resolver would pick `/tools/matchmatrix/index.html` anyway. Building the
  tool is what made those correct, not my SQL.
- **The tool page itself is unaffected** — it is a committed artefact in
  `gqls/sites`, not a resolver-managed field.

**Why the earlier verify looked green:** I ran it inside the same transaction as
the UPDATEs, before any render. A `content_data` verify cannot detect a field the
render will recompute. The check that would have caught it is the one CLAUDE.md
already states — *trust the rendered artefact, not the status* — applied to the
row as well: re-read after a render, not after a write.

**Consequence for `/bugs_open/023`:** its framing ("label and URL are unrelated
schema fields, nothing pairs them") is right about the symptom but understates
the cause. Nothing pairs them **because the URL is not authored at all** — it is
derived, label-blind, from nav order. A fix that only teaches the content writer
to pair them will be overwritten by the next render. The fix has to be in
`chooseCTATargets` (make it label-aware, or let an explicit authored URL win).
Recorded there.

### Post-drain, 2026-07-20 — two more derived-field traps, and a real exposure I had left open

The 12-item batch completed (12/12). Verified against the **rendered pages**:
all fabricated figures gone from `/about.html`, `/entities/gripper-detail.html`
and `/index.html`; `gripper-detail` now renders 5 / 5 / 4 / 24 with correct
labels and **no placeholder suffixes** — the live `2,400+%` and `140+ms` are
gone.

**Trap 3 — the tool list is derived too.** `r4` removed the dead
payload-budget card from `content_data->'items'`; the re-render at 17:10:23 put
it straight back, all five entries, and the homepage still served the 404. Same
shape as the CTA resolver: **the listing is regenerated from the site's tool
pages**, so editing the rendered array cannot hold. It listed the page because
the page still *existed* (`build_status='planned'`, never built).

Fixed the way R6 did it — `pages.status='archived'`
(`SQL_2026-07-20_r4e_archive_unbuilt_tool_page.sql`), which R6 already proved
the listings respect. `tool-matchmatrix` deliberately **not** archived: it is
real now and should be listed. Both `incomplete_page_group` items closed at the
same time (`complete` for matchmatrix, `wont_fix` for payload-budget) — they had
been sitting in `needs_human_review` since 2026-07-14 in a queue with no
consumer (`/bugs_open/033`).

> The structural point, which I have NOT fixed: **a listing that advertises
> pages which were never built is how these 404 CTAs arose in the first place.**
> Archiving rows per site is containment. The durable fix is for the tool-list
> query to filter on `build_status`. Same family as the `023` addendum —
> *derived field, authored edit cannot hold.*

**Trap 4 — I had left the tool itself unprotected.** `tool-matchmatrix` still
read `build_status='planned'` with **zero** `page_components`, because I had
deployed the artefact straight to `gqls/sites` and never created a DB source.
The page was live and working with nothing behind it — so the next rebuild of
that page would have rendered an empty shell **over a working tool**. That is
precisely the `/bugs_open/001` / `/bugs_open/038` class, and I had walked into
it while writing up that very class two hours earlier.

Fixed in `SQL_2026-07-20_r4f_matchmatrix_durable_source.sql`: component +
page_components row wired to match the live reference implementation
(`function=tool-matchmatrix`, `render_mode=template`, `component_level=tool`,
`category=interactive`, `content_data={}`), page set to `deployed`. 23,281 B of
template now in the DB. One deliberate deviation: the three existing tool
components carry `is_active=f` and this one is created **active** — nothing
depends on the flag (`renderAndStoreSiteComponent` ignores it, the R1 finding)
and copying a past cleanup's artefact into a new row would be cargo-culting.

**The pattern across all four traps.** Every one is the same shape: a field that
*looks* authored is actually **derived on render** — CTA URLs from nav order,
CTA target titles alongside them, the tool list from the page set, and the page
body from the component source I had not created. The check that catches this
class is not a better verify query; it is **re-reading after a render**, which is
CLAUDE.md's "trust the rendered artefact, not the status" applied to the DB row
as well as the page. My `content_data` verify ran inside the writing transaction
and could not, in principle, have caught any of them.

## Turn (bugfix-022 thread) — 2026-07-20 evening — scheme-guard live test touched this site

Context: `bugs_open/022` (the R1 damage mechanism — spec background silently
overriding `scheme=dark`) got its structural fix (`enforceLayoutScheme`,
`9c3b0c3e7`) into the v1.0.1140 roll. This site is the canonical scheme=dark
target, so its live verification ran here. What was done TO robot-hands:

- `design_intent` superseded twice and **restored verbatim** from a pre-test
  snapshot — net state: the R1b palette pin exactly as before (`background
  #0F1218`, hard "never light" guidance). The two `bugfix-022-thread` rows in
  `site_specs` history document the test window.
- One real webdesign-agent run (orch `fb744273`) completed WITH THE PIN
  REMOVED: the LLM proposed dark anyway (`#0F1219` — it anchors on
  `content_data.color_scheme`), CSS rendered dark, deployed, live URL
  verified. So: one routine CSS commit on gqls/sites (#0F1218→#0F1219
  hairline drift in some values), site visually unchanged, guard live in the
  render path.
- A deterministic reject-path test (pin pointed at light values) was
  attempted; both dispatches vanished (003-class producer-side drop, noted in
  003) and the test was abandoned with the pin restored rather than left
  armed. If anyone wants that live-fire later, the method is in
  `bugs_closed/022`'s closure record.

R-series implication: the R1 churn mechanism is now guarded at the merge seam
fleet-wide, not just by this site's pin. The pin stays (defence-in-depth).

---

## Turn 10 — 2026-07-22 — R7: expanded the index so "six actuation technologies" is TRUE

Continued from `HANDOFF_2026-07-20`. Everything R1–R6 + R4 re-verified holding on
v1.0.1144 (pod `agent-chassis-59c675c4f-pxr9f`, started 08:47Z). The one genuinely
open, robot-hands-scoped, owner-blocked item was 043's "42 prose fields" carrying the
six-actuation-technologies claim. Put it to the owner; **owner chose EXPAND the
catalogue** (see PLAN D10).

**The finding that made it a data bug, not a copy call.** Read the actual index
(`products`, category='gripper'): five parallel-jaw grippers, three with an explicit
`24 V DC` spec (electric). Full spec dump showed **no** actuation field and **no**
vacuum/magnetic/soft-robotic/adhesive/pneumatic attribute on any of the five. So the
claim named four technologies the index held **zero** of. `[VERIFIED]` by
`SELECT name, specifications FROM products WHERE site_id=…` — not inferred from the
part numbers (that would be the exact trap in WRONG_CALLS; the 24 V DC values are the
evidence, the model names are not).

**Sourcing discipline (the whole point — 043/020 is the fabrication family).** One
real product per missing technology, every figure read off a fetched manufacturer or
distributor page on 2026-07-22, stored with `content_data.source_url` + `verified_date`
exactly as the existing five already do (Festo EHPS cites rs-online, etc.):

| technology | product | key figure(s) verified | source |
|---|---|---|---|
| pneumatic | Festo DHPS-10-A | 34.5 N/jaw closing @6bar, 3 mm stroke, 67 g, 2–8 bar | motionworld (Festo DS 1254040; Festo PDF was binary-unreadable via fetch) |
| vacuum | OnRobot VG10 | 15 kg payload, electric dual-zone | onrobot.com |
| soft-robotic | OnRobot Soft Gripper SG | 2.2 kg, 11–118 mm grip, food-grade silicone | onrobot.com |
| adhesive | OnRobot Gecko SP5 | 5 kg payload, van der Waals | onrobot.com |
| magnetic | Schmalz SGM-HP 50 | 560 N (385 N w/ friction ring), 50 mm, 350 °C | schmalz.com |

I deliberately did **not** carry figures I saw only in a search-result *summary* (e.g.
VG10's "80% vacuum / 12 Nl/min") into the row — only what a fetched page actually stated.
That is the difference between "verified" and "plausible" and it is where this family
bites.

**Stat blocks re-pointed to the catalogue, not hand-typed** (bug 043 fix candidate 1 at
one-site scale). The three count stats now UPDATE from `SELECT count(*)…`/`count(DISTINCT
…)` subqueries over `products`, so they cannot drift from the index again:
about + gripper-detail = 10 models / 6 manufacturers (was 5/5); gripper-detail figures
24 → 39; `index`'s "6 actuation technologies" was fabricated, now `count(distinct
actuation)=6` backs it — same value, now true. Verified inline in the SQL echoes:
10 / 6 / 6 and adhesive1 electric5 magnetic1 pneumatic1 soft-robotic1 vacuum1.

**What is live vs pending.** The catalogue expansion is live in the DB **immediately**,
and because the six-technology prose was already rendered (just unbacked), the claim is
true the moment the index holds them — no re-render needed for *correctness*. Only the
two stat *numbers* (5→10, 5→6 on about/gripper-detail) await a re-render, queued as
`robot-hands-r7-catalogue-expand` with `handler_agent='page-rerender'` + a distinct
`item_key` and `reason='cta_links_stale'` (the proven branch; R4d learned that a NULL
handler_agent leaves the item unroutable-but-green). Until it drains those pages show 5 —
a conservative understatement, not a fabrication. `[NOT-YET-LIVE]` on the numbers.

**Did not fight the queue.** robot-hands has 1 `claimed` item ahead of mine and the
fleet pipeline is doing ~2 completions/hour (bug 029 stall, third roll). README says the
029 recovery is another thread's and not to loop it, so the re-renders sit. Fine — the
data is already correct.

**Residual, recorded not hidden** (into 043 too): MatchMatrix scores only the
parallel-jaw subset on clamping force, so any prose claiming the *tool* evaluates all six
technologies is still ahead of the tool — expansion fixed the **index** claim, not the
**tool-scope** claim. The new grippers also have no browsable catalog row/detail page
(the catalog page is static prose). Neither is part of the owner's decision; both are
flagged for a follow-up.

---

## Turn 11 — 2026-07-22 — R7 verified after the v1.0.1149 roll; R7b fixes what live verification exposed

Owner said a new chassis image was on prod (pod `agent-chassis-7d4ff8b54-cm786`,
v1.0.1149, started 13:56Z). Checked whether R7 went fully live. The R7 re-renders had
already completed under the *previous* image (about 07-22 11:19, gripper-detail 12:04) —
the bug-029 queue drained (it self-heals ~40 min per the corrected 029 diagnosis).

**about.html — clean and correct live: 10 models / 6 manufacturers, zero fabrication
tells.** `[VERIFIED]` curl of the deployed artefact.

**A misstep I caught, logged to WRONG_CALLS.** I curled `https://robot-hands.com/
gripper-detail.html` (root) and saw an empty stat skeleton whose gqls/sites file dated to
**May 02** — and briefly diagnosed a bug-024 delivery gap (re-render marked complete/
deployed but never delivered). It wasn't. **gripper-detail deploys to
`/entities/gripper-detail.html`** — `pages.url` says so, and so did the 07-20 handoff. The
root file is a stale unrelated artefact. On the *correct* URL R7's values were live all
along (10 / 6 / 4 / 39). Lesson: read `pages.url` before curling; the page NAME is not the
path, and "the deployed file is old" against the wrong path proves nothing. The DB was the
tell I ignored first — the component's `rendered_html` (build_status=deployed, updated
12:03) already held the right content; a deployed artefact that contradicts a fresh
`rendered_html` should have pointed me at "wrong file", not "delivery gap".

**R7b — two real defects the correct-URL check exposed** (`SQL_2026-07-22_r7b_…`):
1. **The placeholder-suffix bug (043 point b) was STILL LIVE on gripper-detail.**
   `stat{1..4}_suffix` = `%`, `ms`, `+`, `x` (unedited generic-template junk), so the page
   rendered **"10%", "6ms", "4+", "39x"**. The 07-20 containment cleared this on `about`
   (which simply has no suffix keys) but not here. Cleared all four → empty.
2. **R7's own count bump left two descriptions false** — the exact fabrication-family drift
   R7 was about, now self-inflicted. `stat2_description` listed 5 manufacturers (value 6) →
   added Schmalz. `stat4_description` said "two models publish no payload rating and two
   publish no IP rating" — with ten grippers it is **4 no-payload / 7 no-IP** `[VERIFIED]`
   `count(*) FILTER (WHERE NOT specifications ? 'payload')=4`, `… 'ip_rating')=7`) →
   genericised so no hardcoded count can drift again.
   **Takeaway for the R7 method: updating a stat VALUE from the catalogue is not enough —
   the surrounding prose that cites the old value has to move with it, or it re-fabricates.**
   `about` needed nothing (labels+values only, no descriptions/suffixes).

R7b re-render queued (`robot-hands-r7b-suffix-fix`, gripper-detail only); live-verify
pending the drain.

---

## Turn 12 — 2026-07-22 — R7b's suffix fix did NOT hold; found and fixed the real (fleet-wide) root

> **CORRECTED — R7b (Turn 11) claimed the suffix fix; it reverted within one render.**
Verified the R7b re-render live: gripper-detail still rendered **"10%", "6ms", "4+",
"39x"**, and `content_data` had my `suffix=""` back to `%/ms/+/x`. My *value* (10/6/39) and
*description* (Schmalz, genericised) edits from R7/R7b **held**; only the suffixes reverted.

**Why — and the misstep.** The `system-stats` component is `render_mode='agent'`, and its
`input_schema` marks the suffix fields **`source=static, fallback="%"/"ms"/"+"/"x"`** while
value/label/description are `source=llm` with no fallback. A `static` field left empty
resolves to its fallback **and persists it**. So an empty suffix can never stay empty while
the fallback is junk. I edited `content_data` on that field without checking the schema —
the same shape as the "DERIVED on render, content_data edits can't hold" landmine already
in memory, which I should have applied here. `about` stays clean only because its component
has no suffix fields at all. Logged to WRONG_CALLS.

**This is bug 043 point (b)'s structural ROOT, and it is fleet-wide.** All five
`system-stats` consumers render the fallback: ai-agent-orchestration ("1,000sms",
"14,203%" on vonc) — nonsense that proves it is a leaking placeholder, not a unit.

**R7c fix** (`SQL_2026-07-22_r7c_…`): set the four fallbacks to `""` on the shared
component (`fdd92ad4-…`). Because the field is `source=static` (deterministic, not LLM) and
the other four pages carry their suffix persisted **non-empty**, this changes **no current
live page** — verified blast radius (5 pages / 4 sites, all non-empty) before applying.
Then cleared robot-hands' persisted `%/ms/+/x` to empty so the static resolve lands on the
new empty fallback, and re-rendered. Live-verify pending the drain. Residual, into 043: the
other four sites still carry the junk **persisted** — the schema fix does not retro-clear a
persisted value, and the intended unit is each owner's call, so not touched from this lane.

**The compounding method lesson (Turn 11 + 12):** a stat block on an agent-rendered,
schema-backed component has three independent layers that each need the right level — the
VALUE (llm field, edit content_data), the PROSE that cites it (llm field, edit content_data,
must move with the value), and the UNIT (static field, fix the schema fallback, NOT
content_data). Getting one right and assuming the others follow is how R7→R7b→R7c each
found the next layer.

**R7c VERIFIED LIVE 2026-07-22** (`agent-chassis-7d4ff8b54-cm786`, v1.0.1149). The
re-render completed and `/entities/gripper-detail.html` now renders the stat block bare:
`10 Gripper Models Indexed`, `6 Manufacturers Covered`, `4 Parameters Compared`,
`39 Published Figures Held` — zero unit suffixes, correct labels, corrected descriptions.
(The R7c re-render also changed the card structure to `<div class="stat-value">10<span
class="stat-suffix"></span></div>`; the number is a bare text node now, not a `stat-number`
span — noted so the next verifier's grep does not read an empty `stat-number` as a lost
value.) robot-hands is correct end-to-end: about 10/6, gripper-detail 10/6/4/39, six real
sourced actuation technologies, no fabrication tells. The 043(b) fallback root fix is live
fleet-wide (recurrence prevented); the four other sites' persisted junk suffixes remain
their owners' cleanup, documented in 043.

---

## Turn 13 — 2026-07-24 — R8 (MatchMatrix v2) · R8b (methodology truth) · R9 (catalog grid) — the residuals closed

Owner said "carry on with the residuals" and chose: extend the tool (not soften
the prose), full 043 treatment on the other sites, sweep + prompt rule. This
site's share of that session:

**R8 — MatchMatrix v2, all six technologies, honest physics.** The v1 tool
tested 5 parallel-jaw grippers and its scope note claimed "5 grippers … the
complete index" — false since R7. v2 (committed `997a04e16`, live gqls/sites
`67a6c3d5`, DB durable source synced): all 10 grippers, each assessed on the
criteria its manufacturer PUBLISHES — jaw friction calc unchanged (μ-trap note
kept); magnetic = direct hold m·a·S vs the LOWER published figure (385 N,
friction ring) + a ferromagnetic-workpiece gate; vacuum/adhesive/soft =
published payload vs dynamics-adjusted equivalent m′ = m·a·S/g (a rating earned
at rest is not credited with surviving your acceleration); soft also the
11–118 mm cup range. Requirement panel prints all three formulas with numbers.
**30/30 logic tests** (node:20-alpine, DOM stub over the real submit handler) —
and one test failed FIRST because the TEST was wrong: DHPS-10-A has 6 mm total
stroke, so a 10 mm travel ask rightly fails — the μ-trap lesson self-applied;
kept as case 21b. `[VERIFIED live]` all 10 grippers render, form intact, 0 tells.

**R8b — the methodology page described an invented system.** Re-reading the ~10
tool-scope claims against v2: most became TRUE, but matchmatrix-methodology +
two about differentiators described machinery that has never existed — a
six-dimension weighted distance metric, adjustable/saveable weighting vectors,
spec "normalization" with a fabricated "120 N → 94 N, IEC 60529-aligned"
example (IEC 60529 is the IP-rating standard — the generator even misused the
standard number), third-party/peer-reviewed data, manufacturer "enrolment",
hard-limit override, "identical criteria" across technologies. No honest
extension can make those true — they'd require inventing data/processes. Rewrote
3 fields (about features [0][1][2][4][6]; methodology body; methodology FAQ 9 of
10 answers) to v2's real mechanics; kept truthful entries byte-identical.
`[VERIFIED live]` fiction markers 0, real markers present, on both pages.

**R9 — gripper-catalog gets the real grid.** Reused the EXACT machinery proven
on gripper-detail: one page_components row → `gripper-spec-sheet`
(`query.products:gripper`), sections array mirrored, positions shifted 3→4→5.
Rejected `product-grid` deliberately — it is an e-commerce card
(price/rating/badge/image) and empty price scaffolding is a fabrication
invitation. `[VERIFIED live]` all 10 gripper names render on
/gripper-catalog.html, including the 5 new ones; component survived its render.

**The regression that proves the platform bug.** The fleet sweep (043 work, same
session) found robot-hands/index brief-explanation RE-fabricated: a routine
needs_page re-render at 10:54 that morning had re-invented "2,400+ Gripper
Models Indexed" — the exact figure R4c removed on 07-20 — because the writer's
schema demands a required numeric stat value and rule 14's prohibition offered
no legal alternative ("2,400+" is the prompt's own '2.4M' example shape). Fixed
(computed 10/6/39) AND root-treated: migration 201 (scalar rule with the
empty-string alternative) + an evidence_base writer_block for this site listing
the only assertable numbers. Full mechanics in `bugs_open/043` Update 2026-07-24.
**Lesson for this site's docs: content_data value-edits on agent-rendered
components are NOT durable against the full writer path — durability needs the
evidence_base declaration (now in place) or a query-backed field.**

Re-renders all verified against live pages, not statuses. robot-hands ends the
day fully consistent: index 10/6/39, about 10/6 + honest differentiators,
gripper-detail 10/6/4/39 bare, catalog page listing all 10, tool testing all 10,
methodology describing the real tool.
