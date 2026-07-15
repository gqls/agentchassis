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
