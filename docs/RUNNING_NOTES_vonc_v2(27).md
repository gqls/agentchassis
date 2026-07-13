# RUNNING NOTES — vonc.com / Spark (v2)

Continued from RUNNING_NOTES_vonc.md (archived 2026-07-03; full history there). This file
is the primary chronological log going forward. Per-tool provenance is in PLAN_/NOTES_
files; framework/debug patterns in 016b_debugging_guide_6_.md; the guidelines are
001/002/003 in the repo. Newest entries at the bottom.

---

## CARRIED-FORWARD STATE (compact — as of 2026-07-03)

**Site:** vonc.com / "Spark" — AI daily-provocation platform (world is the content; your
take is the game). Dark violet/magenta brand.
- site_id `9ec3b9ee-5b08-461b-b4f8-9e1e03579c74`
- index page_id `b4d24f8e-fccd-49df-9dad-aa56a0b20a68` (/index.html); plan_id
  `77493277-f510-47ea-aa27-8fca415743d6`
- Deploy: github ('sites' repo) → GitHub Actions → Backblaze B2. k8s -n ai-persona-system
  (app=agent-chassis) / -n kafka (cluster personae-kafka-cluster).

**Index — 6 planned sections.** Live now (5): hero, provocation-card (restored),
gauntlet-cta, brief-explanation, system-stats. lobby-grid: ABSENT (Mode-B empty shell,
correctly filtered) — its build is the NEXT task.
Component ids: hero 23f95f00 · provocation-card 6163ff14 · gauntlet-cta (agent, q100, 20f) ·
brief-explanation 58363894 (agent, q100, 20f) · lobby-grid 9304f14d · system-stats (agent, q100, 22f).
Per-tool state: see NOTES_provocation-card / NOTES_brief-explanation / NOTES_lobby-grid.

**Framework fixes:**
- PATCH_section_visible_content.go — sectionHasVisibleContent exempts `data-runtime-fill`
  sections (runtime-filled shells first-class). DEPLOYED + verified (provocation-card back).
- PATCH_validate_input_contract.go — ValidateInputContract accepts required fields from
  top-level OR input_data.spec.*. WRITTEN, deploy PENDING.
- PENDING: store-validation hardening (reject unclosed <script> at store time, the gap that
  let provocation-card ship truncated); resolver illustration support (ensureAssets surfaces
  only hero/logo, so site_assets.illustration can't resolve though illustration assets exist).

**Key architecture (compact):**
- page-rerender (render_page → rerender_single_page action) ASSEMBLES stored
  page_components.rendered_html + deploys (git_commit). getPageSections drops sections with
  <=10 chars visible text (strips style/script/tags/entities) UNLESS marked data-runtime-fill.
- needs_page → build-dispatch-loop → page-build-handler → page-content-writer: PLAN-DRIVEN
  content build (load sections from site_plan → plan_sections triage → writer renders ALL
  ready planned sections → save_page_sections → deploy). Not just pending sections.
- plan_sections DEFERS a section when a required schema field's data source doesn't resolve
  with no fallback (on_missing empty → skip_field → hits `default:` → defer). save_page_sections
  then drops deferred sections' instances.
- Manual agent trigger = spawn_agent + call_agent via system.agent.generic.requests.
  call_agent validates the target input_contract against TOP-LEVEL extracted fields (the
  input_mapping destination keys); handlers read input_data.spec.* (dispatch convention). For
  component-creator (required section_type) provide it top-level AND in spec — or the validator
  patch fixes the class. build-dispatch-loop's call_handler mapping does NOT flatten
  section_type/page_id.
- Resolver (plan_sections): site_specs.X → site_specs row aspect=X, data.X ; site_assets.X →
  r.assets[X], but only hero/logo populated (from site_plan_imagery kinds hero/logo).
- deriveRenderMode → 'agent' if any field source='llm' else 'template' (NOT a deferral driver).
- Two broken-template modes: Mode A (<no value>FIELD</no>, string-repairable) ; Mode B (bare
  <no value>, needs regeneration). css_themes.css_content empty post-025 (composition via FK
  chain at render).

**Docs:** RUNBOOK_section_assembly_drop, RUNBOOK_phase2_provocation_js, 016b_debugging_guide_6_,
TOOL_DOCS_convention, PLAN_/NOTES_ per tool, PATCH_validate_input_contract.go,
PATCH_section_visible_content.go, rerender_single_page_action.go (patched),
provocation_card_loader.js, provocations.sample.json.

**Standing rules (brief):** schema-before-SQL; reuse/alter before create; structural over
quick; workflows THIN, logic in Go actions; no sub-workflows (spawn sub-agents); every agent
is an orchestrator; agents respond to the CALLER's responses topic; British English; no
logger.Debug (won't show); note any variable-name changes; don't take 0-rows/empty as decisive
without checking the query; verify against deployed artifacts (curl/DB), not ephemeral pod logs.

---

## 2026-07-03 — Section-drop CLOSED (provocation-card restored); log-grep reasoning

Verified via curl: live index serves hero, provocation-card, gauntlet-cta, brief-explanation,
system-stats (lobby-grid absent, correct). provocation-card appears twice in the grep = its
<section> tag + its inline script's [data-component="provocation-card"] selector (ONE section).

Why the log greps failed (documented per request):
 1. First live grep used the literal 'skipped=1 kept=5' — WRONG: zap emits JSON
    {"msg":"getPageSections: filtered empty sections","skipped":1,"kept":5}; skipped/kept are
    separate JSON fields, not a 'field=value' substring. (My error.)
 2. `logs*` file grep had the right message string but the saved files didn't span the rerender.
 3. Live --since=120m grep with the right strings ('filtered empty sections' Info; 'skipping
    section with no visible content' Warn) found NOTHING. These lines DO emit (Info/Warn, not
    Debug). The patched getPageSections DEFINITELY ran — proven by the asymmetric outcome
    (lobby-grid dropped + provocation-card KEPT, which only the marker-aware code produces).
    => LOG-RETENTION artifact: most plausibly the pod that served the rerender (RS 6dff694bb)
    was replaced by a later rollout during active deploying; kubectl logs follow pod lifecycle
    (deleted pod → logs gone). A different sink for action logs isn't fully excluded. curl is
    the ground truth.
LESSON: during active rollouts, verify against the deployed artifact (curl/DB), not pod logs;
grep zap logs by message string + JSON field, never by 'field=value' literals.

## 2026-07-03 — Started lobby-grid builder (next queued task)
Created PLAN_lobby-grid.md. Approach: runtime-fill loader (Path 2, mirror
provocation_card_loader) + data-runtime-fill marker so lobby-grid ships + fills with today's
provocations. Two decisions to confirm with the user BEFORE building: (a) overlap — make
lobby-grid the primary "today's provocations grid" and trim provocation-card's 4-card
mini-lobby? (recommended); (b) data source — an `arena` array (~6 cards) in provocations.json
(Phase 3 pipeline emits it; interim = extend the sample). First step: inspect lobby-grid's
html_template DOM (card/header/CTA selectors, Mode-B) — SQL issued; need the full template to
design the loader accurately.

---

## 2026-07-04 — Reframed lobby-grid as a framework decision (autonomous section composition)

User: the framework should make the section decisions (composition, role, static-vs-dynamic,
data feed) from the domain/site-spec, for new builds AND maintenance, followable autonomously.
Confirmed the recommended lobby-grid decisions (lobby-grid = primary arena grid; provocation-card
= headline + stats + CTAs, mini-lobby trimmed; feed = `arena`).

Uploaded sql_result = the tool-generator agent def (NOT the lobby-grid inspection). Useful:
tool-generator builds SELF-CONTAINED tools (LLM-generated, NO fetch, own page). So the framework
has component-creator (section templates) + tool-generator (self-contained tools) but NO
runtime-fill LOADER builder — the missing piece (provocation-card/lobby-grid loaders were
hand-built).

Design captured in PLAN_dynamic_sections_and_loaders.md:
 - Decisions belong in PLANNING, encoded as a per-section descriptor {role, kind(static|dynamic),
   data_feed} on the plan/spec, produced by site-planner. (Plan currently lacks kind → framework
   didn't know provocation-card/lobby-grid are runtime-filled → assembler dropped them.)
 - New build: planner → component-creator (static template | dynamic shell+DOM-contract+marker) +
   loader-builder (dynamic) + data pipeline (feeds) → content-writer (static) → assembler keeps
   runtime-fill → deploy.
 - Maintenance: quality-auditor detects (dropped-dynamic, overlap, deferred/data-gap, empty-dynamic)
   → work items → handlers.
 - Gaps to add (priority): (1) section descriptor kind/data_feed/role on plan+planner; (2)
   component-creator runtime-feed tier (Tier E: marked shell + DOM contract); (3) loader-builder
   agent (SIBLING of tool-generator — fetch-and-fill; tool-generator forbids fetch); (4) auditor
   rules; (5) Phase-3 data pipeline.
 - Reuse: tool-generator (model for loader-builder), component-creator, site-planner, quality-auditor,
   js_snippets+site-asset-renderer, the data-runtime-fill marker+exemption (shipped).
Immediate lobby-grid: encode the two descriptors + hand-build lobby-grid loader (reuse
provocation_card_loader — doubles as the reference for the loader-builder) + trim provocation-card
mini-lobby. STILL need lobby-grid's DOM (structure-check + full template) — the earlier inspection
SQL wasn't run (tool-generator query came back instead); re-requested.

---

## 2026-07-04 — lobby-grid loader built + install pack ready

Template inspected (htmlblocklobby.txt): all selectors green, Mode-B, 15652 bytes, hover/
entrance script intact. Structure: header (eyebrow/title/subtitle) + 6 cards in DOM order
(featured, standard×4, wide; 2 accent stat-dots) + CTA (label + btn with play-SVG + trailing
text). Icon slot = svg INNER markup; stat text = the non-dot span; aria-label also a slot.
Built: lobby_grid_loader.js (node --check OK; DOM+data contract in header), provocations.
sample.json v2 (`arena` OBJECT — refinement over the bare array: header+CTA need data —
{eyebrow,title,subtitle,cta_label,cta,cards[6]}), lobby_grid_install.sql (Part A js_snippet
insert dollar-quoted; Part B data-runtime-fill marker template+instance, guarded; template
dump confirms the plain data-component literal so REPLACE applies; the template's own script
uses [data-component="lobby-grid"] so the added attribute is harmless).
RUN ORDER: insert snippet → trigger-asset-renderer-vonc.sh (bundle snippets.js) → marker SQL
→ commit /data/provocations.json (v2 sample, interim until Phase 3) → rerender-index-vonc.sh
→ verify (curl: data-component="lobby-grid" present; browser: 6 cards filled; provocation-card
unaffected). THEN step 6: trim provocation-card mini-lobby (next turn, per D-A).

---

## 2026-07-04 — File destinations for the lobby-grid pack (runbook App. H)

User asked where the three deliverables go. Recorded in RUNBOOK_phase2_provocation_js App. H:
 - lobby_grid_loader.js = SOURCE OF RECORD only (library repo, next to provocation_card_loader);
   its content is embedded in the install SQL and ships via js_snippets → site-asset-renderer →
   snippets.js. Never deployed as a file; edits go into the js_snippets row + re-trigger renderer.
 - lobby_grid_install.sql = psql against clients_db in TWO parts (A insert → renderer trigger →
   B marker); then archive with the numbered one-off scripts.
 - provocations.sample.json = REPLACES vonc.com/data/provocations.json in the sites repo (superset:
   keeps today + lobby, adds arena) → Actions → B2. Interim until Phase 3 emits it.
Run order + verification restated in App. H.

---

## 2026-07-04 ~18:15 — Part A + renderer verified; runbook markdown repaired

lobby-grid install progress: Part A snippet inserted (js_len 6743); site-asset-renderer COMPLETED
(corr 045afbf2, load_site → render_js_snippets → deploy_js_snippets → complete, ~12s). Verified the
SHIPPED bundle (latest_snippets.js): both loaders present; lobby-grid-loader fragment extracted and
node --check PASSES as a complete IIFE. 6-byte js_len drift vs source = cosmetic (paste), not
truncation — checked because of our truncation history. psql echo garbling = bracketed-paste
cosmetics; INSERT 0 1 + verify row are the facts.
Runbook markdown repaired: bare <script>-style tokens in PROSE were parsed as inline HTML by the
user's reader, swallowing everything after (same failure mode as the live page bug). Sweep: 21 bare
tags backtick-wrapped (fence-aware, span-aware), 2 backslash-terminated code spans reworded; 4
flagged leftovers all verified inside code spans (safe). Lesson for docs: never leave a bare
angle-bracket tag in markdown prose — backtick-wrap always.
NEXT: Part B marker → commit v2 provocations.json (arena) → rerender index → verify six sections +
cards filling.

---

## 2026-07-04 — Dated backups before Part B (marker UPDATE)

User asked to back up the page_components row before the Part B marker change. Provided:
 - _vonc_pc_backup_20260704 (ALL 6 index page_components rows) — DATED name because the old
   _vonc_index_pc_backup is the stale 2026-07-02 4-row pre-rebuild snapshot, and CREATE TABLE
   IF NOT EXISTS against the old name would silently no-op while looking like a fresh backup.
 - _vonc_cc_lobby_backup_20260704 (the lobby-grid content_components row) — Part B also edits
   the template, and a direct UPDATE does NOT go through store_generated_component so no
   component_versions snapshot happens automatically.
Restore = UPDATE-in-place keyed on id (rendered_html / html_template), no delete+insert.
Note for the taxonomy: dated backup tables per change; never reuse an IF NOT EXISTS backup name.

---

## 2026-07-04 — lobby-grid restored on the index; BUT marker REPLACE broke inline-script selectors

Post-rerender (corr 8b977d1c) voncindex1.html: ALL SIX sections present (hero, provocation-card,
gauntlet-cta, brief-explanation, lobby-grid, system-stats); footer intact; snippets.js referenced;
data-runtime-fill on provocation-card + lobby-grid. lobby-grid RESTORED (primary goal met).

BUG I INTRODUCED (twice): the marker REPLACE 'data-component="X"' → +marker hit BOTH the <section>
tag AND the section's inline querySelector('[data-component="X"]'), producing the malformed
selector [data-component="X" data-runtime-fill="true"] → SyntaxError → the cosmetic hover/entrance
IIFE dies on BOTH provocation-card (last turn) and lobby-grid (this turn). Loaders in snippets.js
are UNAFFECTED (REPLACE only touched rendered_html/template), so sections still FILL + lobby-grid
cards stay clickable (loader adds handlers). Cosmetic + console-error regression, not functional.
My SQL note ("does not affect it") was WRONG.
FIX: fix_marker_selector.sql — revert ONLY the in-selector copy (the one followed by ']'), keep the
<section>-tag marker; both components, template + index instance; RETURNING still_broken=f /
section_marker_kept=t. THEN re-run rerender-index-vonc.sh to redeploy corrected HTML.
LESSON (debugging guide): anchor marker REPLACE on the OPENING TAG, not bare data-component; better,
emit the marker at generation (component-creator runtime-feed tier). Recorded.

OPEN / verify next:
 - Did the v2 provocations.json (with `arena`) get committed to the sites repo? lobby-grid's loader
   needs data.arena to FILL — without it the shell renders empty. Check:
   curl -s https://vonc.com/data/provocations.json | grep -c '"arena"'  (expect 1) + browser.
 - After the selector fix + rerender: confirm no SyntaxError; hover/entrance work; 6 sections; cards fill.

---

## 2026-07-04 — Shared-component clobber check (prompted by cross-site gauntlet contamination elsewhere)

Immediate state OK: arena served (grep=1), selector fix applied (still_broken=f, section_marker_kept=t
x4), re-rerender (corr bf5c9c76) deployed corrected HTML — malformed querySelector gone from live page
(ran before the code push). lobby-grid restored.

SHARED-COMPONENT RISK — read store_generated_component_action.go:
 - Components with forked_from IS NULL are a SHARED library across sites (lookup L204
   `WHERE function=$1 AND forked_from IS NULL`); we edited shared rows.
 - BUT a deliberate GUARD (L315-335) BLOCKS a regeneration that DROPS/RENAMES an existing field on a
   shared component — references the fdd92ad4 system-stats incident (stat_1_number→stat1_value renamed
   in place silently emptied every dependent). brief-explanation was Mode-B with 0 fields → went 0→20
   (pure ADD) → guard passed (why 083 succeeded) → NO dependent content_data stranded. The worst clobber
   was prevented by design.
 - markPagesPendingRebuild (L1018-1062) enumerates affected sites via page_components→pages WHERE
   component_id=X and raises needs_rerender per site. So affected sites = sites whose page_components
   reference our edited component_ids.

CHECKS ISSUED (not yet run): (1) which sites reference brief-explanation/provocation-card/lobby-grid
(the code's own affected-sites join) — vonc-only = no cross-site reach; (2) did 083 raise needs_rerender
for other site_ids; (3) is the 081 stray general-hero (0ef52c95) deactivated — a LIVE generic component
is the classic contaminant auto-selected onto other sites (the gauntlet symptom); (4) js_snippets scope
(\d + site_id?) — if global + render bundles all active snippets, our 2 loaders would ship in EVERY
site's snippets.js → need to read render_js_snippets_for_site.
Interpretation: field guard held (no stranding). Remaining = blast radius + stray + snippet scope.
Provocation-card/lobby-grid expected vonc-only (Spark-specific); brief-explanation is the one to watch;
general-hero stray must be confirmed inactive.

---

## 2026-07-04 — Shared-component check VERDICT: no contamination; brief-explanation base is neutral

Results of the three checks on the shared brief-explanation (58363894), whose dependents are
vonc + idea.uk (1244516d) + robot-hands.com (00ff3af5), both deployed:
 (a) STATIC fallbacks: only cta_primary_label="Get Started", cta_secondary_label="Learn More" —
     fully generic, no Spark voice.
 (b) Template vocabulary: gauntlet/provocation/archetype/contrarian/take ALL false — no hardcoded
     Spark copy; the component-creator prompt kept all voice copy in per-site placeholders.
VERDICT: the 083 regen produced a NEUTRAL how-it-works base. idea.uk + robot-hands previously had
the broken Mode-B shell (q50, 0 fields) → moving them onto a working neutral template is an
accidental IMPROVEMENT, not gauntlet-style contamination. Together with: field guard held (0→20
pure add, nothing stranded), general-hero stray is_active=f, loaders inert off-vonc (self-check for
sections that only vonc has) ⇒ clobber concern CLOSED for our edits.

CAVEAT 1 — schema discrepancy to confirm (do NOT gloss): on 2026-07-02 brief-explanation had EIGHT
static fields (2 CTA labels + stat_1/2/3 label+value); today source='static' returns only the 2 CTA
labels. Most plausible: the stat fields were re-sourced to llm in a PARALLEL chat (user is clearing
things up elsewhere; this was the Finding-2 recommendation for the generic SaaS stat fallbacks) —
which would further REDUCE leak surface. Verification query issued (list all 20 fields + sources).
Not assumed until seen.

CAVEAT 2 — idea.uk + robot-hands are MID-TRANSITION, not finished: the regen set their
brief-explanation instances build_status='pending', but the raised items were needs_rerender
(ASSEMBLE-only) → their pages re-assembled STALE rendered_html from the old broken template. Nothing
has rendered the new 20-field template with their content yet. They stay pending until a content
rebuild (needs_page) per site — which will now succeed (illustration_url optional on the shared
schema benefits them too). Owned by those sites' threads; check query issued (their instances'
build_status + rlen).

DURABLE LESSON (for the autonomy design + guide): regenerating a SHARED base is acceptable only as a
neutral base improvement with pure field ADDITION (the guard enforces the latter). Site-specific
customisation must FORK (forked_from = base id), leaving the base neutral — the "deliberate
migration" the code comments prescribe. Fold into the planner/component-creator section-descriptor
work: the plan should record fork-vs-base per site per component. If vonc later wants Spark-specific
brief-explanation wording, that is a FORK, not another base regen.

---

## 2026-07-04 — Static-field discrepancy RESOLVED; shared-component concern fully closed

Ran the 20-field source list on brief-explanation (58363894): stat_1/2/3_label + _value are now
source='llm' (were static w/ generic fallbacks on 2026-07-02). Confirms the Finding-2 fix from the
parallel chat — the generic "Happy Customers"/"14,203%" stat fallbacks that showed on vonc's live page
were re-sourced to llm so each site's content-writer fills them per-site.
Full field picture on the shared base: 18 llm (per-site voice) + 2 generic static CTA labels
("Get Started"/"Learn More") + 1 illustration_url (optional). No Spark vocabulary. Leak surface even
smaller than the earlier verdict assumed.
=> SHARED-COMPONENT CLOBBER CONCERN CLOSED. idea.uk + robot-hands received a NEUTRAL how-it-works base
(improvement over their prior broken Mode-B shell), not vonc's voice. Field guard held, general-hero
stray inactive, loaders inert off-vonc.

STILL OPEN (owned by those sites' threads, not vonc): idea.uk + robot-hands brief-explanation instances
are build_status=pending with stale rendered_html — the 083 regen raised needs_rerender (assemble-only),
so they re-assembled the old broken template; they await a per-site content rebuild (needs_page) to
render the new 20-field template with their content (will now succeed — illustration optional). Check
query recorded.

NEXT (vonc thread): verify lobby-grid fills in a browser (arena served) → trim provocation-card
mini-lobby (confirmed decision) → remaining tracks: store-validation hardening, illustration wiring,
other-page audit.

---

## 2026-07-04 — lobby-grid VERIFIED live; cross-site status; store_generated_component analysis; provocations 404

**lobby-grid DONE (browser-verified).** Screenshot shows "Six rooms are live right now" + the six
arena cards (AI will never be funny / Remote work killed mentorship / Privacy is already over / Most
'data-driven' decisions aren't / Reading fiction makes you worse at facts / The four-day week — the
overnight wide card) + pulsing stat dots + "Enter the Arena". The Mode-B shell is filled at runtime
from provocations.json `arena` by lobby-grid-loader. Full Path-2 build landed end-to-end: assembler
keeps the marked shell → loader bundled (snippets.js) → data served → cards render. Confirmed the
cards ARE the lobby-grid arena cards.

**Cross-site brief-explanation status (corrects earlier "stuck/stale" caveat).** Instances on the two
other dependent sites:
  idea.uk        : 2 instances, both DEPLOYED (~10.6K) — rebuilt with their own content.
  robot-hands.com: 4 instances, 3 DEPLOYED (~9.2–9.8K) + 1 PENDING (7869 stale).
So brief-explanation is on MULTIPLE pages per site, and both sites have LARGELY rebuilt onto the
neutral template (their parallel-chat work); only 1 robot-hands instance still pending. Transition
nearly complete, not stuck. No contamination (neutral base). Owned by those sites' threads.

**Do we need to change store_generated_component_action.go? — NO (not urgently, not for our regen).**
 - It already did its defensive job: field-set guard held (0→20 pure add, no stranding — the fdd92ad4
   failure it prevents), logs affected_sites, snapshots to component_versions.
 - Gap it does NOT cover: site-specific CONTENT (static fallbacks, hardcoded template wording) leaking
   into a shared base — only the field-SET is guarded. Ours came out neutral because the
   component-creator PROMPT keeps voice in placeholders, not because the store action guarantees it.
 - PROPER fix = upstream fork-vs-base decision (section descriptor carries shared-vs-forked; planner/
   component-creator chooses). That WILL bring a store change later — to create a forked_from row on
   request — as part of the fork capability, not an isolated edit.
 - OPTIONAL cheap guard now (reuses existing affectedSiteIDs): `if len(affectedSiteIDs) > 1 &&
   !allow_shared_base_regen { reject }` → turns accidental multi-site clobber into a conscious choice.
   Design call (would also gate the auditor's auto-regen of shared components). HOLD unless wanted.
 CHOICE recorded: no change now; fork-vs-base is the structural fix, folded into the autonomy/section-
 descriptor work. Shared-base regen only for NEUTRAL improvements with pure field ADDITION; site-
 specific voice = FORK (forked_from = base), the "deliberate migration" the code comments prescribe.

**NEW ISSUE — provocations page 404 in B2.** B2 GET vonc.com/provocations/index.html → 404 NoSuchKey.
That is the destination of the arena cards AND the hero/gauntlet CTAs (site_specs.cta.primary_url =
/provocations/index.html, + arena card urls). So those links currently dead-end. provocations-index
page e4b3b195 is planned but evidently never built/deployed. Check: SELECT build_status FROM pages
WHERE id='e4b3b195-...'. If unbuilt → building/deploying the provocations page is the fix (arena's
destination); folds into the other-page audit. Did not guess the fetch trigger.

QUEUE (vonc): provocations page build (arena/CTA destination — 404) + other-page audit; store-validation
hardening; illustration wiring; provocation-card mini-lobby trim (still pending — the overlap is now
live: provocation-card's 4-card mini-lobby AND lobby-grid's 6-card arena both on the index).

---

## 2026-07-04 — Provocations page confirmed never built (the 404 cause); gated build queued

pages row e4b3b195 (provocations-index, /provocations/index.html, "Provocations Archive | Spark",
page_type section-index): build_status='planned', updated 2026-06-22 — planned in the original build,
NEVER built. Explains the B2 404 for the destination of ALL primary CTAs (hero, gauntlet-cta,
system-stats via site_specs.cta.primary_url, lobby-grid Enter-the-Arena + 6 card urls,
provocation-card primary CTA). Schema correction noted: pages.name (not page_name) — my query column
was at fault first, per the standing rule; corrected via \d pages.
GATE issued before building (runbook App. I): D1 plan sections for provocations-index; D2 those
components' state (Mode-B / deferral risk — post assembler patch, empty sections are DROPPED, so a
blind build could deploy a husk); D3 prior work items for the page (a failed item may say why the
pipeline skipped it). Provisional needs_page INSERT written (page_id authoritative, mirrors the
proven index item). If D2 shows broken/dynamic sections → fix first (index playbook: data gaps /
marker+loader / regeneration); the archive list section is likely dynamic (Phase-3 provocations feed
family).

---

## 2026-07-04 — provocations-index: zero planned sections; 7 silent no-op completes; guide merged to v5

D1: 0 plan sections for provocations-index (query shape proven on index; DISTINCT page_name +
pages.sections='[]' confirms pending). D3: SEVEN items ALL complete/no-error (needs_page 06-22 +14s,
2× manual 06-26, 4× rerender 06-23→07-01) with the pages row UNTOUCHED (build_status planned,
updated_at = creation) ⇒ handler exits on zero-sections before any page-writing step; rerender skips
deploy with no components. ROUTE: planner emitted page (nav/CTAs point at it) but NO sections (likely
systemic for section-index/archive pages — no dynamic-list vocabulary); build+rerender treat "nothing
to do" as SUCCESS ⇒ two weeks invisible. Do NOT fire the provisional INSERT (= no-op #8): plan
sections FIRST (header + archive list component, kind=dynamic, provocations feed), then build.
PREVENTION (guide entry): planner ≥1-section invariant; handler zero-PLANNED (fail/raise
needs_plan_sections) vs zero-READY (report) guard; rerender skip-warn; auditor linked+planned rule +
post-deploy URL presence check; section-index vocabulary (section-descriptor + loader-builder + the
complex-tool loop in the parallel chat).
GUIDE FORK RESOLVED: uploaded 7_2 had parallel-chat entries but lacked my 3 (deferral drop; artifact-
not-pod-logs; marker anchoring). Merged into 016b_debugging_guide_merged.md (v5 changelog line) + the
new silent-noop entry. Apply the MERGED copy to the project.
Docs: NOTES_provocations-index.md (new, provenance), cross-entries in NOTES_lobby-grid +
NOTES_provocation-card, runbook App I gate results.

---

## 2026-07-05 — Confirms interpreted; archive-list spec written; runbook tidied (full-UUID scripts)

CONFIRMS: DISTINCT page_name = DEFINITIVE — current plan covers ONLY about, archetypes, contact,
index, tool-archetype-taster-quiz, tool-gauntlet. No provocations-index under any spelling (and no
`provocation` blog page either — yet that page was built → evidence the pages.sections FALLBACK path
works; queued for the other-page audit). The second query hit the WRONG page id: ecb637c1 is
TOOL-GAUNTLET (sections ["gauntlet-interface","archetype-result-card"]) — proves the fallback
mechanism is real/populated for tool pages, but provocations-index's own pages.sections confirm is
STILL PENDING (corrected full-uuid query in the runbook Step 0a).

COMPOSITION DECISION (precedent: tool pages run 1-2 self-contained sections): provocations-index v1 =
ONE section — `provocations-archive-list`, self-contained (own header), runtime-fill. Avoids guessing
an inner-page header component; single plan row.

SPEC written (SPEC_provocations-archive-list.md) with the thread's lessons baked in AT GENERATION:
data-runtime-fill emitted in the template's section tag (no string surgery); NO inline script (loader
= external js_snippet; extraction-bug class impossible); llm header fields (no deferral risk); list
entries pure markup (nothing for the resolver to fail on); clone-template item for the variable-length
archive (vs lobby-grid's fixed 6); visible empty state so the page ships before the data lands.
Data contract: provocations.json `archive.entries[]` {date,title,teaser,stat,url}, newest-first, cap 24.

RUNBOOK tidied per request: top "CURRENT STATUS + NEXT STEPS" block with FULL-uuid SQL for Steps 0-4
(0 confirms incl. plan-id + style-id pattern check; 1 component via 084/083-pattern + verify
has_marker=t has_inline_script=f; 2 plan-section INSERT-SELECT copying style ids from index ordering 0;
3 needs_page build + 200 check; 4 loader+archive data). Elided '9ec3b9ee-...' style uuids in App I
replaced with full values. Sequence gate: component BEFORE plan row BEFORE build (a plan row naming a
nonexistent component is unknown triage territory).
NEXT: run Step 0 confirms → I generate the 084 trigger script from the spec.

---

## 2026-07-05 — Steps run out of order (plan row + build fired before the component exists); 084 ready

Step 0 confirms all good (pages.sections=[] for e4b3b195; plan 77493277 current; style ids NULL on
existing rows → INSERT copied NULLs). Step-1 verify correctly 0 rows (component not created yet — 084
hadn't been generated). BUT Steps 2+3 ran anyway: plan row e006a35e inserted (provocations-index/0/
provocations-archive-list) and needs_page FIRED at a plan whose only section names a NONEXISTENT
component. pages.build_status still 'planned' (nothing built). Issued a 60-min work-items query to see
what the early build did — the interesting branch: if it raised a needs_new_component child that
COMPLETED, the ValidateInputContract patch is live and the loop built the component (would be the
first proof of the work-item path for component-creator); if the child FAILED missing [section_type],
patch not deployed → 084 manual path.
084_create_provocations-archive-list_vonc.sh generated + validated (083 dual-placement; description
QUOTE-FREE — attribute values named in prose to avoid kcat/JSON escaping hazards; asserts: marker
instruction, no-script instruction, function pin, JSON parses). Sequence now: inspect early-build
items → 084 (unless loop built it) → Step-1 verify → RE-FIRE Step 3 (fresh gen_random_uuid item_key)
→ watch (SQL added to runbook) → Step 4 loader+data.
Runbook: newcomer task paragraph added at top; EXECUTION STATE block (steps 0/2 done, 3 early, 1 open);
watch SQL added under Step 3.

---

## 2026-07-06 — Early build confirmed silent no-op #8 (new variant); proceeding with 084

60-min work-items query: only the needs_page item itself (manual-build-provocations-f7e6429e),
complete ~2.5 min, no error, NO children. So the pipeline did NOT self-heal the missing component —
no needs_new_component raised (and therefore no evidence either way on the validator-patch deploy;
the triage never tried to dispatch component-creator). NEW VARIANT of the silent-noop family for the
guide: a plan row naming a NONEXISTENT component also passes through silently (neither builds, fails,
nor escalates). pages.build_status still 'planned'.
DECISION: run 084 (manual component creation, dual-placement) → Step-1 verify with the two
generation-time guards eyeballed (has_marker=t, has_inline_script=f — if the LLM ignored either
instruction, targeted template edit or regen BEFORE building) → re-fire Step 3 needs_page (fresh
gen_random_uuid key) → watch → curl 200 with empty state → Step 4 loader pack against the generated
template's real DOM.

---

## 2026-07-06 — 084 SUCCEEDED (guards held at generation); page build complete in 33s; verifying

084 (corr 7b5f96df): provocations-archive-list CREATED — 70d6662a-0e6f-478d-bc2e-b9e8e5eaeb37, active,
q100, 8 fields, has_marker=t, has_inline_script=f. FIRST LIVE VALIDATION of baking the guards in at
generation: marker born in the template (no Part-B string surgery — Step 4 loses that step entirely)
and no inline script (extraction-bug class impossible for this component). The dual-placement trigger
pattern worked again (no contract violation).
Step 3 re-fired (448ee097): triaged → claimed 14:01:24 → complete 14:01:57 (~33s), no error. Plausible
for a one-section page (one content-writer LLM call) — but after 8 silent no-ops on this page, NOT
trusting 'complete' without artifacts: verification queries issued (pages.build_status/deployed_at;
page_components row with marker_in_rendered/has_clone_template/has_empty_state; curl 200 + grep
data-component). Step-4 prep queries issued: the 8 fields+sources and the FULL template dump — the
generated DOM is authoritative over the spec's class names (lobby-grid precedent) before I write
provocations_archive_loader.js + the archive data + the snippet install (Part A only; NO marker part).

---

## 2026-07-06 — Silent no-op #9 (component + plan row BOTH present); deferral eliminated; suspicion → load_spec_sections

Verification after the 33s 'complete' (item 448ee097): pages.build_status STILL 'planned',
deployed_at/last_built_at NULL, page_components = 0 rows. Query shape proven → real. So the build
no-op'd AGAIN despite the plan row (e006a35e) AND the component (70d6662a) existing.

fieldsources.txt read: all 8 fields are llm or static-with-fallback (cta_url → /index.html) —
NOTHING can defer. Template confirmed to spec: marker on the section tag, hidden clone-template item
(bonus: stat text span has its own class provocations-archive__item-stat-value — cleaner loader than
lobby-grid's non-dot hunt), empty state, CTA, NO script. Component is right.
=> Deferral hypothesis ELIMINATED. The build never SAW the section — suspicion moves upstream to
load_spec_sections. Candidate mechanisms (do not conclude yet):
 (i) pages.sections precedence: tool-gauntlet carries pages.sections=[...] and the provocation blog
     page BUILT despite being absent from the plan → pages.sections is a working lookup path. Our
     page has sections=[] — if the code reads pages.sections FIRST (treating [] as present-and-empty)
     → sections_found=0 → zero path → complete. 
 (ii) the plan lookup keys on something our inserted row doesn't match (page_name vs page ref vs a
     plan-version column — NB \d pages indexes reference built_from_plan_version though the column
     paste was truncated; pages HAS that column).
DECISIVE next: (A) jsonb_pretty(result) of item 448ee097 (sections_found / ready / skip reason);
(B) does the INDEX page carry pages.sections too (precedence evidence). If A=found:0 and B shows the
index populated → pragmatic fix = set pages.sections='["provocations-archive-list"]' on e4b3b195
(reuse of the proven tool-page pattern) + re-fire; else follow A's reason; if A sparse → read the
load_spec_sections action code.
UUID note: I mistyped the page id twice in the reply drafting query B — canonical value restated:
e4b3b195-919f-45ad-854e-201d3e846ea8. (Full-uuid discipline: copy from constants, don't hand-type.)

---

## 2026-07-06 — Item 448ee097 result read: workflow stopped AT load_page_record (before sections). Hypothesis corrected.

voncdocs1.txt = jsonb_pretty(result) of the 33s 'complete' build. response contains ONLY
`site_record` (+ response_status complete). Healthy index rebuilds carried sections_found/
sections_saved/deploy_result — NONE present here. => The workflow terminated at/just after
load_page_record, BEFORE load_spec_sections. All pages.sections-vs-plan precedence theorising was
downstream of the real stop point.

HONEST CORRECTION (guide amendment pending): the planning gap (plan had no rows for this page) is a
REAL independent fact, but the SILENT-COMPLETE mechanism for the provocations builds may be the
load_page_record NOT-FOUND branch completing gracefully with only the site record — not the
check_has_ready_sections zero-path written in the guide §9 route. Both are silent-noop variants;
amend the route once the branch is confirmed. (The June no-ops may ALSO have died at page lookup —
unproven either way.)

WHY would load_page_record miss a page that exists (spec carries page_id + page_name + filename)?
Candidate in-hand: spec filename 'provocations/index.html' vs pages.url '/provocations/index.html'
(leading slash) — BUT the index used 'index.html' vs '/index.html' and WORKED, so either the chain
prefers a key that matches for index but not here, or the branch differs. load_page_record_action.go
is the guideline's own fallback-chain example (001 §1022) — requested; readable not guessable.

DIAGNOSTICS ISSUED (full uuids): (1) in-flight 986c106f status + page/pc state (still 'claimed' at
+19s — noops completed in 33-65s, so a long claim would mean THIS run is different); (2) DECISIVE
jsonb_object_keys(result->'response') healthy-index vs noop side-by-side (shows exactly which step's
outputs exist); (3) page-build-handler workflow dump (load_page_record's not-found branch +
complete_workflow output_fields); (4) B re-run with the CORRECT page uuid (my typo killed the first
attempt). No refires until the branch is understood.

---

## 2026-07-06 — load_page_record read: page IS found; the gate is pages.sections (section_count 0)

voncdocs2.txt uploaded EMPTY (0 bytes) — none of the four diagnostics arrived; re-requested compactly.
load_page_record_action.go read in full:
 - Lookup priority: page_name (site_id+name) first; page_id only if name empty/bogus (nonPageNames).
   Our spec's page_name 'provocations-index' matches pages.name exactly → LOOKUP SUCCEEDS. The
   not-found branch is NOT our path.
 - On success it returns found:true + `sections` PARSED FROM pages.sections + section_count.
   Our page: pages.sections=[] → section_count:0.
=> Coherent story matching the only-site_record result: the workflow BRANCHES on the page record's
   own sections — empty pages.sections routes to a graceful completion (output_fields = site_record
   only) BEFORE load_spec_sections ever runs. My plan-row insert targeted the right concept, wrong
   COLUMN: the gate reads pages.sections, not site_plan_sections.
Existing evidence beyond inference: the provocation BLOG page built with ZERO plan rows (absent from
the plan's six page_names) → pages.sections alone demonstrably drives builds; tool-gauntlet carries
populated pages.sections. Near-certain; query B is the confirm (never ran: my uuid typo + empty file).
CONFIRMS re-issued (B pages.sections index-vs-provocations; 986c106f status; page-build-handler
workflow dump for the branch + output_fields). FIX ready, GATED on B: UPDATE pages SET
sections='["provocations-archive-list"]' WHERE id='e4b3b195-919f-45ad-854e-201d3e846ea8' AND
sections='[]' → re-fire needs_page.
GUIDE AMENDMENT pending (after workflow dump): §9 route corrected to the page-record sections gate
(not check_has_ready_sections); prevention gains: planner must populate pages.sections at page
creation (June gap manifested in BOTH tables); the zero-sections branch should consult the plan or
fail loudly, never complete silently.

---

## 2026-07-06 — Mechanism FULLY CONFIRMED (workflow dump); pages.sections UPDATE applied; attempt #11 in flight; Step-4 pack built

voncdocs3.txt (all diagnostics arrived this time):
 - B: index pages.sections = the 6 sections; provocations was [] → GATE CONFIRMED. User's UPDATE ran:
   provocations-index now ["provocations-archive-list"].
 - 986c106f: complete 15:42:51 — no-op #10 (pre-UPDATE, expected).
 - Workflow dump closes every observation: the silent complete is the step named `complete_error` —
   a complete_workflow with output_fields [page_content, site_record] + success_message "Content
   writer skipped — page has no sections defined", reached from check_has_ready_sections
   (ready_count==0) and as error_step of earlier steps. Writer never ran → page_content absent →
   result carried ONLY site_record (the exact signature we read). Healthy `complete` outputs
   [sections_saved, deploy_result] (matches index results). STRUCTURAL SMELL: an error path
   implemented as a SUCCESSFUL completion; its message never surfaces into the item.
 - NEW DISCOVERY: load_spec_sections = load_page_sections_from_spec — reads site_specs aspect
   `site_plan` (AUTHORITATIVE) → fallback page_record.sections. The site_plan_sections TABLE is NOT
   read by the build path (planner's store). So my plan-row insert (e006a35e) targeted the right
   concept in the WRONG STORE twice over; the pages.sections UPDATE is the unblock. Row kept
   (planner-consistent). THREE section sources documented: spec site_plan aspect → pages.sections →
   site_plan_sections (not read here). Optional completing check issued: dump the site_plan aspect.
 - Attempt #11 (the INSERT in the user's message) fired AFTER the UPDATE → first run with the
   fallback populated → expect the full path (writer → validate → save → update_status → rerender →
   deploy). Watch + verify with the standing set.
GUIDE amended (merged copy): AMENDMENT 2026-07-06 on the §9 entry — confirmed route (complete_error
signature; only-site_record diagnostic), the three-source finding, prevention additions (planner
writes pages.sections; zero-sections must fail/raise, never report success).
STEP-4 PACK BUILT (ready for after the page verifies): provocations_archive_loader.js (list-only —
header/CTA are build-time here, unlike lobby-grid; clone-template per archive.entries, cap 24, hides
__empty; node --check OK), provocations.sample.json v3 (+archive, 8 entries; keys today/lobby/arena/
archive), provocations_archive_install.sql (Part A ONLY — no marker part, it was emitted at
generation). Run order in the SQL header: insert → asset renderer → commit JSON v3 → browser verify.

---

## 2026-07-06 — ARCHITECTURE DECISION: site_plans family = authoritative plan store

User's position (agreed after working it through): the authoritative plan belongs in site_plans +
children, not the site_specs aspect (EMPTY for vonc — 0 rows; query looks sound, aspect-list + global
confirm issued per the 0-rows rule; behaviour corroborates: every build served by the pages.sections
FALLBACK) and not pages.sections (unversioned, unconstrained, drift-prone — the June gap manifested
there). Reasons: only versioned store (is_current unique/history/superseded_at/source_agent); only
constrained store (FKs, UNIQUE(plan_id,page_name,ordering), per-section style/version FK columns);
already load-bearing (resolver reads site_plan_imagery); natural home for the section descriptor
(kind/data_feed/role columns); site_plan_pages implies plan→pages materialisation as the intended
derivation direction. Three-peer-sources-with-unclear-precedence is exactly what caused ten silent
no-ops + two of my fixes landing in the wrong store (site_plan_sections row; then pages.sections).
DECISION + target architecture + gates recorded in PLAN_dynamic_sections_and_loaders.md: alter
load_page_sections_from_spec to read site_plan_sections FIRST (fallback pages.sections during
migration; delete the spec-aspect branch once confirmed dead globally); planner writes plan family
only, pages.sections becomes a single-writer materialised cache; guards (≥1-section invariant vs
site_plan_sections; auditor drift rule; complete_error fails loudly).
Gate queries issued (A aspect checks; B \d site_plan_pages + plan-77493277 contents — locates whether
the June planner emitted the PAGE but skipped its SECTIONS; C vonc drift baseline). Requested the
load_page_sections_from_spec action Go file for the alteration.
Provocations thread still open in parallel: attempt #11 (post pages.sections UPDATE) watch/verify
pending; Step-4 pack ready in outputs.

---

## 2026-07-06 — Plan-storage gates read; load_page_sections_from_spec ALTERED (table-first)

GATES: (A) spec site_plan aspect is GENERATIONAL not dead — vonc has none (11 aspects; roadmap is a
different thing) but FIVE sites carry a current one (incl. robot-hands 00ff3af5; 1368e337 has 20
superseded + 1 current). Two planner generations coexist ⇒ the aspect branch STAYS as source 2, my
"dead branch / delete it" assumption CORRECTED. (B) site_plan_pages holds ALL 8 vonc pages incl.
provocations-index (full row, role section-index, created same instant) ⇒ the June planner emitted
PAGES fully and skipped SECTIONS for exactly roles blog-post (legit — blog pipeline) + section-index
(the defect). Invariant refined: ≥1 section for every planned page whose ROLE page-build-handler
builds; role→pipeline mapping explicit. (C) vonc drift baseline CLEAN — all 8 pages agree
pages.sections == plan tables ⇒ the vonc planner already dual-writes table→cache; decision formalises
existing behaviour.
ACTION READ: load_page_sections_from_spec already reads the plan TABLES in fallback 2b (sibling
synthesis) — the alteration is a PROMOTION not an introduction. ALTERED (outputs copy): Step 1 =
site_plan_sections (current plan, by page_name, ordered) + sync to pages.sections (same guarded
IS DISTINCT FROM update); aspect → Step 2 (query wrapped in len==0 guard; uses-block condition gained
`specSource == ""` — FLAGGED edit so Step-1 fills aren't relabelled/re-synced); pages fallback + 2b
unchanged; banners renumbered 1-5; new source value "site_plan_tables" (downstream may branch on
source — noted). Diff verified strictly additive apart from the flagged condition; braces 73/73; new
locals planRows/tblErr/comp clash-free. Guarded-UPDATE now ×3 in file — future tidy-up, not folded in.
PRE-DEPLOY GATE issued: the 5 aspect sites' table_pages under a current plan — 0 everywhere ⇒ zero
behaviour change for them; >0 ⇒ per-page aspect-vs-table comparison first.
STILL OPEN: provocations build attempt #11 result unreported (post pages.sections UPDATE — expected
to be the first REAL build of the page); Step-4 loader pack ready in outputs.

---

## 2026-07-07 — ALTERATION WITHDRAWN (029 Q1); plus a working-copy clobber incident + lesson

WITHDRAWAL: user challenged the table-first change ("if it's not dead code I don't want to change the
working architecture — the spec may have a planning aspect and the site plan may come from that.
Check."). Read the uploaded design docs BEFORE defending:
 - 027 is headed "026 — Design Composition & Site Design Planner" — the DESIGN planner (writes one
   aspect, resolved_composition); does not bear on page/section plan storage.
 - 029 = the deciding doc. Its Q1 is EXACTLY our question ("Plan storage: site_specs aspects vs new
   table") and is OPEN, with a stated LEAN to partitioned site_plan_* aspects (site_plan_structure,
   site_plan_page_<name> carrying section lists) + a Phase-1 reconcile_site_plan Go action. The
   TABLES are Q1's ALTERNATIVE — implemented for the vonc generation despite the lean.
 - Reality = THREE shapes: legacy singular site_plan blob aspect (5 sites; what the loader reads);
   029 partitioned aspects (apparently unimplemented — confirms issued: DISTINCT aspect LIKE
   'site_plan%' + repo grep reconcile_site_plan); the tables (vonc generation).
=> Table-first loader alteration WITHDRAWN; do not deploy. The unmodified action serves all
generations; vonc is unblocked via pages.sections (dual-written by the vonc planner). The
authoritative-store choice is 029 Q1 — decided once at the planner/reconciler level (the other
chat's loop), with this thread contributing: the table path now exists in production, post-dating
the lean. Store-agnostic preventions RETAINED: role-aware ≥1-section invariant; complete_error fails
loudly. Pre-deploy gate query now moot.

INCIDENT (operational lesson): this session started on a FRESH container — /home/claude had reset —
and the withdrawal script assumed working copies existed. Consequences: (a) the altered .go could not
be marked (file gone; outputs deployable copy rm'd, WITHDRAWN cp failed → outputs briefly had
NEITHER); (b) `cat >> PLAN_dynamic_sections_and_loaders.md` CREATED A 19-LINE FRAGMENT which the
later cp CLOBBERED over the full outputs doc. RECOVERY: full plan doc reconstructed from
conversation context (rebuilt-note added); WITHDRAWN provenance stub created (repo original is
untouched, so no code was lost). RUNNING_NOTES_vonc_v2 survived only because its cp happened to fail.
LESSON (add to guide later): /home/claude resets between sessions; /mnt/user-data/outputs persists.
At session start, RE-SEED working copies FROM outputs before any >> append or cp TO outputs; never
`cat >>` a working-copy path without confirming it exists (>> creates a fragment that then clobbers
outputs on cp); treat outputs as the durable store and /home/claude as scratch.

STILL OPEN on vonc: provocations attempt #11 result (post pages.sections UPDATE — expected first REAL
build of the page); Step-4 loader pack ready in outputs; provocation-card mini-lobby trim queued.

---

## 2026-07-07 — User's repo copy of load_page_sections_from_spec = the WITHDRAWN alteration; revert recommended, decision pending

User re-uploaded their current load_page_sections_from_spec_action.go asking "is it correct?".
Identified: it IS the withdrawn table-first alteration, applied verbatim (398 lines / 14452 bytes;
2026-07-06 header; Step-1 site_plan_sections block; flagged specSource=="" guard; source value
site_plan_tables). Tricky region verified intact in their copy (wrap + declaration + guard).
ANSWER given on two axes: (1) as CODE it is sound — strictly additive, resilient fall-through,
behaviour-preserving for the five aspect sites IF the pre-deploy gate holds (table_pages=0), and
consistent for the vonc generation (stores agree). (2) as ARCHITECTURE it is the withdrawn change —
keeping it pre-empts 029 Q1 (makes tables the de-facto build authority before the design thread
decides) → RECOMMEND REVERT.
Revert artifacts: git restore is the byte-exact path (git diff first); fallback
load_page_sections_from_spec_action.ORIGINAL.go in outputs = faithful reconstruction by reversing the
four known edits on the uploaded copy (verified: site_plan_tables absent; braces 64/64 = 9 pairs
fewer; 119-line delta matching the alteration shape; honest note: 6 bytes off the first upload's
size — trivial whitespace suspected; git authoritative). DEPLOY NOTE: if the altered chassis was
already deployed (pushes on the 4th/6th), reverting source requires redeploy or pods keep table-first
behaviour.
KEEP option stated with obligations: pre-deploy gate query (five aspect sites table_pages=0 required)
+ record in the 029/Q1 thread that the loader already implements the table-first alternative.
DECISION PENDING (user to call revert vs keep) → record in PLAN_dynamic_sections_and_loaders +
here once called.
STILL OPEN on vonc: attempt #11 build result (post pages.sections UPDATE) unreported; Step-4 archive
loader pack ready in outputs; provocation-card mini-lobby trim queued.

---

## 2026-07-07 — Revert decision CLOSED; attempt #11 (eb3cdf42) COMPLETE in ~5 min; one gate left; runbook you-are-here consolidated

DECISION: user reverts the loader action with load_page_sections_from_spec_action.ORIGINAL.go
(6-byte drift accepted; git restore = byte-exact alternative). Recorded in the plan doc. Redeploy
note given: pods keep table-first behaviour until the next chassis build/push.

DELIVERED (the thing I'd been waiting on, now explained precisely): the post-UPDATE build item
manual-build-provocations-eb3cdf42 — claimed 16:33:21 → complete 16:38:15 (~5 min), no error. The
ten silent no-ops all finished in 33–65s; five minutes has the shape of the FIRST REAL BUILD
(content-writer LLM call + validate + save + deploy handoff).

ONE GATE REMAINS before Step 4 (stated as three copy-paste artifacts in the runbook YOU ARE HERE):
(1) pages row build_status/deployed_at/last_built_at off 'planned'; (2) the page_components row for
provocations-archive-list (rendered_len, marker_in_rendered, has_clone_template, has_empty_state all
t); (3) curl 200 + data-component grep on https://vonc.com/provocations/index.html. Rationale
restated: after ten complete-with-nothing items, item status is not proof; artifacts are.

SIDE-EVIDENCE (user's extra query, parallel track): gauntlet-interface / latest-news /
tool-archetype-taster-quiz all have js_content extracted + <script src= refs, no raw inline — the
extraction pattern is LIVE for tool components. provocation-card + lobby-grid cosmetic inline
scripts are candidates for the same treatment later (cosmetic track, not critical path).

RUNBOOK: consolidated to a SINGLE ">>> YOU ARE HERE (2026-07-07) <<<" block at the top (decisions,
state, the three-artifact gate, Step-4 run order, queue). Stale markers demoted: the 2026-07-05
"START HERE" heading → "History: ... (superseded)"; the 2026-06-29 ">>> YOU ARE HERE" → historical
position marker; Appendix F's "(YOU ARE HERE — 2026-07-02)" → historical marker. Newcomer paragraph
now points at the new block.

---

## 2026-07-07 — GATE PASSED: /provocations/index.html is LIVE (first real build after 10 no-ops); Step 4 current

All three artifacts delivered and green: (1) pages row build_status=deployed, deployed_at 2026-07-06
16:38:00 (inside eb3cdf42's window); last_built_at EMPTY — minor bookkeeping inconsistency (build
sets deployed_at only) → audit list. (2) page_components: position 1, provocations-archive-list,
deployed, rendered_len 7455, marker_in_rendered=t, has_clone_template=t, has_empty_state=t.
(3) curl 200 + data-component="provocations-archive-list"; git commit "Rerender:
provocations/index.html" (9900517). The 404 destination of every primary CTA is GONE.

Deployed-page HTML read (index_5_.html): snippets.js referenced in THIS page's head (loader delivery
live — checked because a missing ref would make Step 4 silently not fill); content-writer header copy
in voice: eyebrow "THE RECORD", title "Every <em>provocation</em> on record.", subtitle "...The
positions are permanent. The splits don't lie."; empty state "Nothing filed yet. The first
provocation drops soon."; CTA href=/index.html "Enter today's Arena" (the static fallback working);
clone template ×1; zero <no value>; page whole. The generation-time-guards build worked through the
real pipeline end to end.

STEP 4 = CURRENT (pack in outputs): install SQL Part A (provocations-archive-loader; NO marker part)
→ trigger-asset-renderer-vonc.sh (expect 3 active snippets in the bundle header) → commit
provocations.sample.json v3 (archive, 8 entries) as vonc.com/data/provocations.json → browser: 8 rows
fill, empty state hides, rows navigate; index unaffected.
QUEUE after Step 4: provocation-card mini-lobby trim; framework preventions (role-aware invariant;
complete_error loud; last_built_at bookkeeping); Q1 evidence hand-off; inline-script extraction for
provocation-card/lobby-grid cosmetics (pattern proven live on the three tool components).

---

## 2026-07-07 — Step 4 rewritten as explicit 4.1–4.4 (full SQL inlined); tool docs updated; heredoc incident

User asked for a much clearer Step 4 (the terse 2026-07-05 sketch confused; also my chat shorthand).
DONE: runbook YOU ARE HERE now carries the explicit procedure — 4.1 psql install (FULL
provocations_archive_install.sql inlined in the runbook, dollar-quoted body and all; expected
INSERT 0 1 + verify row js_len ~3800-4200), 4.2 asset-renderer trigger (script path; orchestration
COMPLETED check; curl the bundle: header "3 active snippet(s)" + grep provocations-archive-loader),
4.3 replace vonc.com/data/provocations.json with the v3 sample (adds archive[8]; keeps
today/lobby/arena) + curl grep '"archive"'=1, 4.4 browser: 8 rows fill, empty state hides, index
unaffected; ordering note (4.1 before 4.2; page fills only when bundle AND data are live). The
historical Step-4 sketch is marked superseded pointing at 4.1–4.4. Same procedure given verbatim in
chat.
TOOL DOCS updated: NOTES_provocations-index.md new dated STATE (page LIVE, unblock history one-liner,
last_built_at bookkeeping observation, categories updated); SPEC_provocations-archive-list status
line → COMPONENT CREATED + PAGE LIVE + Step 4 pending.
INCIDENT #2 this session (file handling): wrote the runbook block via an UNQUOTED bash heredoc — the
shell executed every backticked markdown span as a command (substituting empty strings; the fences
and inlined SQL vanished; it even ATTEMPTED my example curl — caught by the egress allowlist, whose
error text got substituted INTO the document). Damage confined to the new block; rewritten via a
QUOTED heredoc and verified (spans/fences/dollar-quoted body present; proxy text purged).
LESSON: always quote heredoc delimiters (<< 'EOF') when the body carries markdown/backticks/$ — add
to the operational lessons alongside the container-reset/fragment-clobber rule.

---

## 2026-07-07 — STEP 4 DONE (browser-verified); ghost-row defect fix ready; arena question answered

4.1 js_len 3281 vs source 3287 (±6 paste drift, same as lobby-grid; bundle+browser authoritative —
my 3800-4200 runbook estimate corrected). 4.2 renderer COMPLETED d6b54136; bundle "3 active
snippet(s)". 4.3 archive served. 4.4 SCREENSHOT: 8 rows render, empty state hidden. Archive WORKS.
DEFECT (from screenshot): empty dot-row above "5 Jul" = the hidden clone template rendering — author
`.provocations-archive__item{display:grid}` beats the hidden attribute's UA display:none. FIX:
fix_archive_template_display.sql (anchored on the base-rule opener, guarded; template + instance) →
085 rerender (make_085_rerender_provocations.sh seds the proven index trigger: page id, filename,
labels). Guide entry added (css-specificity pattern) + prevention: component-creator must emit
[data-…-template]{display:none} alongside hidden for clone templates.
ARENA QUESTION: no arena page exists; the INTERACTIVE surface does — /tools/gauntlet/index.html
(deployed, gauntlet-interface js_content 3909 live). CTA graph CIRCULAR: hero→archive (primary_url we
set pre-page), archive→home (spec'd fallback), gauntlet-cta→archive; only nav/footer reach the tool.
OPTIONS given: A (recommended, minimal) retarget archive CTA → /tools/gauntlet/index.html (schema
fallback jsonb_set + anchored instance REPLACE — SQL provided; rides the same 085 rerender); B leave
until the real take-filing arena exists. Broader CTA-map pass QUEUED as deliberate work (URLs baked
into rendered sections → proper refresh = section rebuilds). AWAITING user's A/B call + fix results.

---

## 2026-07-07 ~16:30 — Option B recorded; 085 works; fix NOT yet applied (rendered_len evidence)

CTA DECISION: B — leave the graph until the real arena exists (retarget SQL parked; broader CTA-map
pass stays queued). 085 script correct (sed swap visible in the echo) and RAN (corr 0ea43016):
deployed_at bumped to 16:20:17 ⇒ the page-level rerender path is proven for provocations-index.
DISCREPANCY CAUGHT: rendered_len STILL 7455 (pre-fix value; the fix adds ~110 chars → ~7565) and no
template_fixed/instance_fixed output in the paste ⇒ fix_archive_template_display.sql was not run (or
0-rowed unnoticed) before the rerender ⇒ the redeploy shipped the ghost dot-row again. Gave the
two-second live check (curl grep for the display:none rule → expect 0 now) + the remaining sequence:
fix SQL (expect t/t; 0 rows = anchor mismatch, stop) → 085 → hard refresh / curl grep = 1 /
rendered_len ~7565.
last_built_at: NULL through BOTH build and rerender ⇒ column dead; audit-list item confirmed.
QUEUE unchanged after this lands: provocation-card mini-lobby trim → framework preventions
(role-aware invariant; complete_error loud; last_built_at bookkeeping) → CTA-map pass (parked with B).

---

## 2026-07-08 — provocations-index CLOSED; next = provocation-card mini-lobby trim

Ghost-row fix verified end-to-end: fix SQL both UPDATE 1 (template_fixed=t, instance_fixed=t); 085
rerender corr cd5f553e, deployed_at 2026-07-08 16:30:43; rendered_len 7455 → 7671. The +216 delta =
the hiding rule inserted into BOTH the base `.provocations-archive__item {` and its mobile
media-query copy (selector appears twice → REPLACE fired twice → correct; both instances needed the
rule). DB side proven; live confirm = curl grep for 'data-archive-template] { display: none;' → 2 (was
0) after Actions push + hard refresh (dot-row gone).
PROVOCATIONS-INDEX THREAD DONE. Full arc for the record: planned 2026-06-22, never built (10 silent
no-op completes) → root cause complete_error zero-sections path + pages.sections gate (load_page_record
returns pages.sections; load_page_sections_from_spec reads spec aspect → pages.sections; site_plan_sections
table not read by this path) → unblocked via pages.sections UPDATE → component 70d6662a built with
generation-time guards (marker + no inline script — first live validation) → page live (build eb3cdf42
~5min) → archive loader (Step 4) fills 8 rows → ghost-row css-specificity fix → CLOSED.
NEXT: provocation-card mini-lobby trim (confirmed) — remove .pc-card-grid from template 6163ff14 +
data.lobby fill from provocation_card_loader.js; keep headline+stats+CTAs; then bundle re-render +
index rerender. THEN framework preventions (role-aware ≥1-section invariant; complete_error loud;
last_built_at dead-column bookkeeping) + the parked CTA-map pass (Option B).

---

## 2026-07-09 — Live confirm; docs updated; HANDOFF written for a fresh chat

LIVE CONFIRM: `curl -s https://vonc.com/provocations/index.html | grep -c 'data-archive-template] {
display: none;'` → 2 (was 0). The deployed page carries the hiding rule in both the base selector and
its mobile media-query copy; ghost dot-row gone. DB evidence (rendered_len 7671) and the visitor-facing
artifact agree. provocations-index thread FULLY CLOSED.

DOCS UPDATED: PLAN_provocation-card — the mini-lobby trim moves from "open decision" to CONFIRMED, and
a full "Trim plan" section is added (3 artefacts: template block, loader `data.lobby` block, the live
index instance because rerender is assemble-only; the two deliberate consequences — the `pc-container`
1fr 1fr media-query rule and the now-dead inline hover script; method = dump → edit offline → UPDATE
full text → verify by length delta; two separate deploys; acceptance criteria). NOTES_provocation-card
— dated state. RUNBOOK — YOU ARE HERE retitled and now points a fresh chat at the handoff.

HANDOFF CREATED (user request): HANDOFF_2026-07-09_vonc_spark_minilobby_trim.md, following the
travelling-docs convention (dated, stands alone, `Categories:` tags). Sections: §1 first actions;
§2 orientation (agentchassis, Spark, the runtime-fill mechanism end to end); §3 DONE state with the
full provocations arc, the complete_error signature, the three section sources, the reference
component 70d6662a, the shared-component verdict, and the withdrawn plan-storage alteration (029 Q1);
§4 the trim task — why, scope, the two consequences, the safe method, both redeploy paths, artifact
verification, acceptance; §5 data to collect (dated backups + four dumps + the shared-library check);
§6 commands/triggers table; §7 ids + hard-won schema notes (pages.name; js_snippets has no site_id;
last_built_at dead); §8 gotchas (artifact-not-status; 0-rows discipline; marker anchoring; hidden vs
author CSS; assemble-only rerenders; dual-placement trigger + quote-free description; generation-time
guards; ±6-byte js_len drift; container reset; quoted heredocs; no bare angle-bracket tags in prose);
§9 backlog (framework preventions first); §10 file inventory. Verified: zero bare HTML tags in prose
outside code fences (the reader-breaking failure mode).
`Categories:` (docs, handoff, milestone)

---

## 2026-07-09 — Bundle primed for the trim; method CORRECTED (I had drafted the rejected mechanism)

User asked to prime a cmd/bundle request for the next task. Grounding it in the real code/docs first
changed the task itself:
 - `cmd/bundle` flags confirmed from its own usage block (project knowledge), not guessed:
   required `-analysis -root -constitution -task` + ≥1 `-scope`/`-include`; repeatable
   `-scope path[:Sym]`, `-include`, `-doc`; DB gather only with `-psql` (`-schema-tables`,
   `-runtime-site/-page`, `-capabilities`, `-df-filter`); `-step`, `-neighbour`, `-out`, `-dry-run`.
   NB /mnt/project/main.go is cmd/FUSE (RRF ranking), not cmd/bundle — a flag cross-check against it
   proved nothing; caught before concluding.
 - File-header check gave exact repo paths and two findings: `fix_component_template_action.go`
   ALREADY has a `remove_element` fix type (this trim), but defers page_components content changes to
   the "section-editor workflow"; `rerender_single_page_action` is confirmed ASSEMBLE-ONLY and also
   injects `cc.js_content` per component. `check_component_standards.go` lives under
   actions/discovery_checks/ — a wrong -scope path would have silently under-scoped the bundle.
 - Docs 003/002 (project knowledge): **content_data is the source of truth; HTML patching was
   REJECTED as an edit mechanism** (patched rendered_html is lost on the next re-render). Two
   re-render paths: light `rerender_page_sections` (page_rerender item; from stored content_data via
   RenderComponentAction; no LLM; **escalates to full rebuild when content_data is NULL**) and full
   `needs_page`. So my handoff §4 method (hand-UPDATE the live instance) was the rejected mechanism.
   Corrected in HANDOFF §4 + §4.0, PLAN_provocation-card, NOTES_provocation-card. Reassurance: every
   edit we shipped touched BOTH template and instance, so nothing live is at risk — the instance edits
   were bridges, and the template carries the truth.
 - Mode-B consequence: provocation-card's content_data may be NULL → the light path would escalate.
   Added D0 probe to HANDOFF §5 as the query that decides the sequence.
DELIVERED: bundle_minilobby_trim.sh — task string framing the whole question (which action edits the
template; which re-renders instances from a changed template; Mode-B/NULL content_data; which item
type raises each path; do hand-added attributes survive a re-render; how a js_snippets change reaches
the bundle). Scopes: fix_component_template (whole), store_generated_component (3 symbols),
rerender_single_page (whole), + resolved render_component / rerender_page_sections /
render_js_snippets_for_site / section-editor(component_swap), save_page_sections; -include registry.go.
Docs 003, 002, 022, 016b, 020. `-schema-tables content_components,page_components,pages,js_snippets,
component_versions,site_work_items,sites,site_specs`; `-runtime-site vonc.com -runtime-page index`.
Step-0 RESOLVERS (grep the action's registration key, exclude registry.go, fail loudly unless exactly
one hit) because a wrong -scope silently under-scopes rather than erroring. DRY=1 supported. Trailing
comment block carries the row-state SQL the bundle cannot: the content_data probe, the schema/marker
check, the shared-library check, and the item-type history.
`Categories:` (tooling, method-correction, reuse)

---

## 2026-07-09 — Bundle resolver v1 failed on render_js_snippets_for_site; v2 rewritten (two-stage, registry-first)

Run result: resolvers 1-2 passed (render_component, rerender_page_sections pinned), resolver 3 matched
0 files → script exited, as designed (fail loudly rather than silently under-scope). ROOT CAUSE of the
resolver, not of the repo: v1 grepped for the QUOTED ACTION KEY while EXCLUDING registry.go, assuming
every action file repeats its own registration key in a header comment. Many do (load_page_record,
store_generated_component); the js-snippets renderer evidently does not — so the only occurrence was
the one place v1 excluded.
FIX (v2, tested against a fake tree before shipping): `resolve_by_key` is two-stage — find the key
ANYWHERE incl. registry.go → read the `Handler:` symbol from that entry → locate the file that DEFINES
that function (`^func <Handler>(`). Verified: pins site_asset_actions.go from a registry entry alone,
with no key in the file; and fails loudly on a bogus key. Also added: `must_exist` for the four known
paths (fail with a hint instead of under-scoping); `PIN_*` env overrides for every resolved path;
save_page_sections falls back to the registry if renamed; the section-editor is now a SOFT resolve
(grep for component_swap; if no Go file mentions it, warn that it is likely a workflow/agent_definitions
row and lean on doc 002's "Section Editor" section) so a missing file cannot block the bundle.
LESSON (generalise): resolve an action's file from the REGISTRY (key → Handler → definition), never
from a convention about header comments. Same class as "read the real code before asserting".
`Categories:` (tooling, resolver, lesson)

---

## 2026-07-09 — Bundle v3: all 8 paths resolved; cmd/bundle lives under contextkit; symbol-scope shared files

RESOLVER TABLE (now knowledge, record it):
  fix_component_template   platform/orchestration/actions/fix_component_template_action.go
  store_generated_comp     platform/orchestration/actions/store_generated_component_action.go
  rerender_single_page     platform/orchestration/actions/rerender_single_page_action.go
  render_component         platform/orchestration/actions/v3_site_actions.go        <- SHARED file
  rerender_page_sections   platform/orchestration/actions/rerender_page_sections_action.go
  render_js_snippets       platform/orchestration/actions/render_js_snippets_for_site_action.go
  save_page_sections       platform/orchestration/actions/save_page_sections_action.go
  section_editor           platform/orchestration/actions/section_editor_actions.go  <- exists as Go code, not just a workflow
Confirms v1's root cause: render_js_snippets_for_site_action.go DOES exist but never repeats its own
registration key in a header comment, so excluding registry.go left zero hits.

FAILURE 3 (mine): `go run ./cmd/bundle` — cmd/bundle lives in the CONTEXTKIT tree
(docs/agent_docs/docs019_.../go_files/contextkit/cmd/bundle), not at the repo root. RUNBOOK_thin_slice
documents the correct form (`go run ./$CK/cmd/bundle ... -constitution $CK/thin_slice_constitution.md`,
run from the repo root so -root/-scope/-doc resolve against the chassis); I had followed
example_bundle.txt's abbreviated form instead. LESSON: prefer the authored runbook's invocation over
an example's shorthand.

v3 CHANGES: CK variable (overridable) + guards that cmd/bundle and the constitution exist under it;
invocation `go run "./$CK/cmd/bundle"`; NEW scope_for() heuristic — a file dedicated to one action
(<key>_action.go) is scoped WHOLE (its helpers matter), a SHARED file is scoped BY SYMBOL, so
render_component enters as `v3_site_actions.go:RenderComponentAction` rather than dragging
UpdateSiteDefaults/CompilePageSections into the bundle (B4a: irrelevant content dilutes attention).
Heuristic tested against all four resolved paths. resolve_by_key now returns "path|Handler".
`Categories:` (tooling, resolver, lesson)

---

## 2026-07-09 — Bundle v4: contextkit is a SEPARATE GO MODULE; run bundle from inside it

FAILURE 4: `go run ./$CK/cmd/bundle` from the chassis root → "main module (github.com/gqls/agentchassis)
does not contain package .../contextkit/cmd/bundle". Cause: contextkit has its own go.mod; `go run`
only builds packages in the CURRENT module. So RUNBOOK_thin_slice's `go run ./$CK/cmd/bundle` form
cannot work as written from the chassis root either (or that tree was a single module when it was
written) — the invocation must happen INSIDE contextkit, which also makes bundle's defaults
-assembler-bin ./cmd/assembler and -dbcontext-bin ./cmd/dbcontext resolve.

PATH SEMANTICS (settled from cmd/assembler source, not guessed): the assembler does os.ReadFile on
-constitution and on each -doc — read relative to the PROCESS CWD — while -scope/-include are resolved
against -root. exec.Command inherits cwd. Therefore with cwd=contextkit:
  -root                                  = chassis root, ABSOLUTE
  -scope / -include                      = relative to the chassis root (unchanged)
  -analysis / -constitution / -doc / -out = ABSOLUTE
v4 does exactly this: resolve + grep in the chassis root, absolutise analysis/constitution/docs/out,
then `cd "$CK_ABS" && go run ./cmd/bundle`. ANSWER to "where do I run it from?": ANYWHERE — the script
cds itself.

CLAIM RETRACTED (self-corrected the same turn): I asserted that `[ -f "$CK_ABS/go.mod" ] && echo …`
would exit under `set -e`. My own demo disproved it — bash's set -e IGNORES the failure of a
non-final command in an `&&` list, so the script continues and prints nothing. Both branches printed
REACHED END, exit 0. The replacement if-block is cosmetic, NOT a bug fix; the note claiming otherwise
was wrong and is corrected here. (The real set -e traps are elsewhere: `local x=$(cmd)` masks the
status, and command substitution in a plain assignment DOES propagate failure — which is what makes
the resolver assignments abort correctly.)
LESSON about the lesson: run the demo BEFORE writing the claim, and if the demo contradicts it, change
the claim, not the code.
`Categories:` (tooling, lesson, self-correction)

---

## 2026-07-09 — Bundle v4 RAN successfully; verdict pending; handing the trim to the other chat

`bundle_minilobby_trim.sh` (v4) completed: bundle written to /tmp/bundle_minilobby_trim.md. It carries
the constitution, the task, the in-scope code (fix_component_template whole; store_generated_component
by 3 symbols; rerender_single_page whole; v3_site_actions.go:RenderComponentAction by symbol;
rerender_page_sections / render_js_snippets_for_site / save_page_sections / section_editor_actions
whole; registry.go as -include), docs 003/002/022/016b/020, the \d for
content_components,page_components,pages,js_snippets,component_versions,site_work_items,sites,site_specs,
and vonc.com/index runtime evidence — all read-only (dbcontext + pure assembler; nothing triggered).

Four failures reached that point, each recorded because each was a wrong assumption of mine rather than
a repo fault: (1) resolver excluded registry.go, assuming action files repeat their own registration key
in a header comment — render_js_snippets_for_site_action.go does not; fixed with a two-stage
key→Handler→definition resolver, tested on a fake tree. (2) `go run ./cmd/bundle` assumed cmd/bundle at
the chassis root; it lives in contextkit. (3) `go run ./$CK/cmd/bundle` from the chassis root can never
work: contextkit is a SEPARATE GO MODULE. Bundle must be invoked from inside contextkit, so
-analysis/-constitution/-doc/-out go ABSOLUTE (cmd/assembler os.ReadFile's them relative to process cwd)
while -scope/-include stay relative to -root. (4) A `set -e` bug I claimed and my own demo disproved —
retracted in place rather than left in the record.

Also learned and worth keeping: `render_component` lives in the SHARED v3_site_actions.go, so it is
scoped by SYMBOL (scope_for heuristic: `<key>_action.go` ⇒ whole file, shared file ⇒ file:Handler) to
keep the bundle focused; `section_editor_actions.go` exists, so the section-editor named by
fix_component_template's header is real Go code, not merely a workflow row.

STATE: provocations-index CLOSED and live-verified. The mini-lobby trim is the next task, blocked on the
bundle VERDICT, which must decide (a) the sanctioned way to change content_components.html_template —
fix_component_template's remove_element vs store_generated_component vs section-editor vs direct SQL;
(b) which action re-renders page_components.rendered_html from a CHANGED template — the light path
(rerender_page_sections, a page_rerender item, from stored content_data via RenderComponentAction) vs a
full needs_page rebuild vs rerender_single_page, which only ASSEMBLES pre-rendered components;
(c) what a Mode-B section with empty/NULL content_data does, since 003 says the light path then
escalates to a full rebuild; (d) whether attributes hand-added to rendered_html survive a re-render;
(e) how a changed js_snippets.js_content reaches /assets/js/snippets.js. Handing to the other chat with
HANDOFF_2026-07-09 + the bundle.
`Categories:` (tooling, milestone, handoff)
