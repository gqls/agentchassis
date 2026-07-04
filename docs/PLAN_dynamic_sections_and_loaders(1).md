# PLAN — Autonomous section composition: static vs dynamic (runtime-fill) + data feeds

**Opened:** 2026-07-04  •  **Status:** DESIGN (answers "how should the framework decide
these things, for new builds and for maintenance")
**Related:** PLAN_lobby-grid.md, NOTES_provocation-card.md, RUNBOOK_section_assembly_drop.md,
the tool-generator + component-creator + site-planner + quality-auditor agent definitions.

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
A per-section descriptor `{component_name, role, kind, data_feed}` on the plan — either new
columns on `site_plan_sections` (kind text, data_feed text, role text) or a JSONB descriptor,
written by the planner (and surfaced in the resolved_composition spec aspect). This is the
record the build + maintenance follow autonomously.

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
   as JSON (provocations.json `{today, arena}`), modelled on the news pipeline.

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
```
Steps still gated on inspecting lobby-grid's actual DOM (structure-check + full template dump)
so the loader targets the real selectors — see PLAN_lobby-grid.md.
```
