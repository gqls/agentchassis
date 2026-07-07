# PLAN — Autonomous section composition: static vs dynamic (runtime-fill) + data feeds

**Opened:** 2026-07-04  •  **Status:** DESIGN (answers "how should the framework decide
these things, for new builds and for maintenance")
**Related:** PLAN_lobby-grid.md, NOTES_provocation-card.md, RUNBOOK_section_assembly_drop.md,
the tool-generator + component-creator + site-planner + quality-auditor agent definitions.

*(Rebuilt 2026-07-07 after an accidental clobber — see RUNNING_NOTES_vonc_v2 for the incident.)*

---

## The question
The framework should build the whole site from the domain/site-spec. The decisions I was
about to ask a human to make — which sections a page has, each section's role (to avoid
overlap), whether a section is static or filled at runtime from a live feed, and which feed —
should be made by the framework from the spec, for BOTH new builds and maintenance, so the
plan/runbook can be followed autonomously.

## Where these decisions belong: the PLANNING layer, encoded in the spec/plan
The site-spec is the source of truth (aspects incl. content_direction, design_intent,
identity, mission, resolved_composition). The site-planner already produces the concrete plan
(site_plan_sections: page_name, ordering, component_name). What's missing is that the plan
does not capture, per section:
- **role** — what the section is for (needed so two sections don't own the same content;
  provocation-card's mini-lobby vs lobby-grid is exactly this overlap, uncoordinated).
- **kind** — `static` (fixed brand/explanatory content, filled at build) vs `dynamic`
  (runtime-filled from a feed; content changes faster than the build cadence).
- **data_feed** — for dynamic sections, the named JSON feed the loader fetches.

Because the plan doesn't carry `kind`, the framework didn't know provocation-card/lobby-grid
are runtime-filled — which is why the assembler dropped them and we hand-added the
data-runtime-fill marker.

### Decision rules (the planner applies, reasoning from the spec)
- **static vs dynamic:** a section is DYNAMIC if its content changes faster than a rebuild
  (today's provocation, the live arena grid, live stats, news) or depends on real-time/user
  data; STATIC if it's fixed brand/explanatory content (hero, how-it-works, about). Derived
  from the site's content model in the spec (content_direction/mission), not guessed per
  component.
- **data_feed:** the spec carries a small feed catalogue (the site's live JSON feeds + their
  shapes, e.g. provocations.json `{today, arena}`); each dynamic section maps to a feed key.
- **role / overlap:** the planner assigns each section a distinct role and coordinates so no
  two own the same content. (e.g. lobby-grid owns the arena grid; provocation-card owns the
  single headline provocation + stats + CTAs, no mini-lobby.)

### Encoding
A per-section descriptor `{component_name, role, kind, data_feed}` on the plan — written by
the planner (and surfaced in the resolved_composition spec aspect). This is the record the
build + maintenance follow autonomously. (Where the descriptor physically lives follows the
029 Q1 plan-storage decision — see SUPERSEDED note at the end.)

## New-build flow (autonomous)
domain → site-spec → **site-planner** emits composition + per-section descriptors → for each
section:
- static → **component-creator** builds a fillable template → **page-content-writer** fills at
  build.
- dynamic → **component-creator** builds a SHELL template carrying `data-runtime-fill` + a
  declared DOM contract (the selectors the loader targets) → **loader-builder** (see gap 3)
  builds the client-side loader for the feed → the **data pipeline** emits the feed.
→ assembler keeps runtime-fill sections (exemption already shipped) → deploy.

## Maintenance / improvement flow (autonomous)
**quality-auditor** scans the built site + plan and detects, then queues work items that
build-dispatch-loop routes to handlers:
- a planned section absent from the deploy → if `kind=dynamic`, verify marker + loader exist
  → queue needs_loader / needs_component_regeneration.
- two sections with the same role/content (overlap) → queue a composition/spec fix.
- a deferred section (required field's data source unresolved) → queue a spec/data fix
  (populate site_specs.* / register asset / relax the field).
- a dynamic section rendering empty at runtime → check its feed (data pipeline).
The site-plan is the "plan"; the build pipeline is the "runbook"; both run without a human in
the loop.

## Structural gaps to add (priority order)
1. **Section descriptor** (`kind`, `data_feed`, `role`) on the plan/spec + the planner
   produces it. Without this the framework can't tell static from dynamic. (Smallest,
   highest-leverage — unblocks everything else.)
2. **component-creator runtime-feed tier ("Tier E")** — for `kind=dynamic`, generate a marked
   shell + a declared DOM contract (selector → feed-field map) instead of a build-time-filled
   template.
3. **loader-builder agent** — mirror `tool-generator` (LLM generates the JS, saved + wired)
   but for FETCH-AND-FILL loaders. NB tool-generator explicitly forbids fetch, so this is a
   SIBLING agent, not a reuse of tool-generator. Input: the section's DOM contract + the feed
   shape; output: an IIFE loader (fetch the feed, fill the selectors, fail gracefully),
   installed as a js_snippet (applies_to the component) + bundled by site-asset-renderer.
4. **quality-auditor rules** for the maintenance detections above.
5. **Data pipeline** (Phase 3) — an orchestrator + scheduled refresh emitting the named feeds
   as JSON (provocations.json `{today, arena, archive}`), modelled on the news pipeline.

## Reuse (before building new)
- `tool-generator` = the model for loader-builder (LLM-generates + saves + wires); loader-builder
  is a fetch-and-fill sibling.
- `component-creator` = section templates (extend with the runtime-feed tier, not a new agent).
- `site-planner` = composition (extend to emit descriptors).
- `quality-auditor` = maintenance detection (extend its rules).
- `js_snippets` + `site-asset-renderer` = loader delivery (already how provocation-card ships).
- `data-runtime-fill` marker + the assembler exemption = already shipped.

## Immediate application to lobby-grid (recommended decisions, framework-consistent)
Rather than a pure one-off: (a) encode the descriptors now — lobby-grid `{role: arena grid,
kind: dynamic, data_feed: arena}`, provocation-card `{role: headline provocation + stats +
CTAs, kind: dynamic, data_feed: today}` (mini-lobby trimmed); (b) hand-build lobby-grid's
loader by reusing provocation_card_loader — this hand-build IS the reference spec for the
loader-builder agent (gap 3); (c) trim provocation-card's mini-lobby. So the manual work
doubles as the template for the autonomous capability.
*(2026-07-06 addendum: the provocations-archive-list build is the SECOND hand-built reference —
generation-time marker, no inline script, clone-template list — see SPEC_provocations-archive-list.)*

---

## DECISION (2026-07-06) — Plan storage: the site_plans family is authoritative
*(SUPERSEDED same day — retained for provenance; see the final section.)*

**Context.** The build path's stated authority (`site_specs` aspect `site_plan`, read first by
`load_page_sections_from_spec`) is EMPTY for vonc — every successful build has been served by
the fallback, `pages.sections`. A third store, `site_plan_sections`, is written by the planner
but not read by the build. Three sources with unclear precedence produced ten silent no-op
builds and two fixes landing in the wrong store.

**Decision (as made, pre-supersession).** `site_plans` + children become the single source of
truth: only versioned store (unique is_current per site, history, superseded_at, source_agent);
only constrained store (FK cascade; UNIQUE(plan_id, page_name, ordering); per-section
component_version_id / palette_id / layout_id / typography_set_id); already load-bearing
(resolver reads site_plan_imagery); natural home for the section descriptor; site_plan_pages
implies plan → pages materialisation as the intended derivation direction.

**Changes (as drafted):** load_page_sections_from_spec altered to read site_plan_sections
first (fallback pages.sections; spec-aspect branch retained for the aspect-generation sites);
planner writes the plan family only, pages.sections a single-writer materialised cache;
guards (≥1-section invariant; auditor drift rule; complete_error fails loudly).

### Gate results + alteration (2026-07-06)
- **Gate A:** the spec `site_plan` aspect is GENERATIONAL, not dead — vonc has no such aspect
  (11 aspects; `roadmap` is a different thing), but FIVE sites carry a current one (1368e337 —
  20 superseded + 1 current — 5fe15466, 00ff3af5/robot-hands, 2a8ebf9c, 4851f6fc). Two planner
  generations coexist; the aspect branch STAYS.
- **Gate B:** `site_plan_pages` holds ALL EIGHT vonc pages incl. provocations-index (full row:
  role section-index, slug, url, nav_order 5, title) created at the same instant as the rest.
  The June planner emitted PAGES completely and skipped SECTIONS for exactly the two
  non-standard roles: blog-post (legitimate — the blog pipeline builds those) and
  section-index (the defect). Invariant refined: every planned page whose ROLE is built by
  page-build-handler must have ≥1 section; the role→pipeline mapping must be explicit.
- **Gate C:** vonc drift baseline is CLEAN — all 8 pages agree exactly between pages.sections
  and the plan tables. The vonc-generation planner already dual-writes table→cache.
- **Alteration applied then WITHDRAWN** (see below): new Step 1 reading site_plan_sections +
  sync to pages.sections; aspect → Step 2; new source value "site_plan_tables"; strictly
  additive apart from one flagged condition edit.
- **Pre-deploy gate (moot after withdrawal):** the five aspect sites' table_pages under a
  current plan.

---

## SUPERSEDED (2026-07-06, same day) — decision deferred to 029 Q1; alteration WITHDRAWN
Reading 029 ("Site Plan as Declarative Artefact, Reconciler, and LLM Tiering") after the user's
challenge: plan storage is 029's OPEN Q1 — "site_specs aspects vs new table" — with a stated
LEAN to partitioned aspects (`site_plan_structure`, `site_plan_page_<name>` carrying section
lists), read by a Phase-1 `reconcile_site_plan` Go action. The tables are Q1's ALTERNATIVE,
implemented for the vonc generation despite the lean. 027 is actually 026 (design composition
planner) and doesn't bear on this. Observed reality = THREE shapes: legacy singular `site_plan`
blob aspect (five sites; what the loader reads); 029's partitioned aspects (apparently
unimplemented — confirm via `SELECT DISTINCT aspect FROM site_specs WHERE aspect LIKE
'site_plan%'` + repo grep for `reconcile_site_plan`); the tables (vonc generation).
=> The table-first loader alteration is WITHDRAWN (outputs:
load_page_sections_from_spec_action.WITHDRAWN.go stub; do not deploy). The unmodified action
serves all generations; vonc is unblocked via pages.sections (which the newer planner
dual-writes). The authoritative-store decision is made ONCE at the planner/reconciler design
level (029 Q1), with this thread's evidence contributed: the table path now exists in
production for new sites, post-dating the lean.
SURVIVES store-agnostically: role-aware ≥1-section invariant (pages of roles
page-build-handler owns must have sections wherever the plan lives — the June defect);
complete_error fails loudly, never reports success.
