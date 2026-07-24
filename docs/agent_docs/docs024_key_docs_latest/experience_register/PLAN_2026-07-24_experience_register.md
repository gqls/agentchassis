# PLAN — experience register

**Status: DESIGN DECIDED (owner rulings 2026-07-24). Build NOT started.**
Phase 1 (this doc set + draft artifacts + bug filing) done this session; Phase 2 (platform
code) gated on owner go; harvest gated on the vonc pilot (owner ruling: wait for tools-api).

Session: "experience register". Related workstreams: `experience_loop/` (per-site plans —
level 4 of this design, unchanged), `travelling_docs/` (the substrate we build on),
`gauntlet_dead_cta/` (owns the vonc pilot + tools-api — coordinate, do not fork).

## 1. Problem

Interaction behaviour is invented afresh per site per build. `content_components` records
the artifact (`html_template`, `js_content`, `input_schema`) with **no field for what a click
does or where a link should go**; link destinations are schema fields regenerated per build.
`bugs_open/023`: *"a component declares a button's label and its URL as two unrelated schema
fields, and nothing anywhere expresses 'a label implies a destination' as a constraint."* The
CTA resolver manufactures destinations label-blind by nav-order. The travelling-docs
workstream named the same defect at tool level ("lost intent"). The experience loop owns
journeys but writes each EXPERIENCE_PLAN from a blank page (~14KB authored fresh; 8 council
runs to converge once; exactly one plan exists).

**The register**: a library of small reusable experiences — e.g. *carousel card: "read more"
expands the summary in place; clicking the card/image navigates to the detail/article page;
that page links onward to related info and tools* — held as base patterns, forked/bound per
site. Links become checkable ("where they probably should go" is declared, not guessed) and
acceptance tests are derived from the pattern instead of hand-written.

## 2. Owner rulings (2026-07-24 — decisions, not open questions)

1. **Substrate**: `experience_patterns` register table (machine-facing contract + selectable
   taxonomy columns) **plus** a travelling doc per entry (new `doc_plans` subject_type) for
   provenance/direction/notes. Mirrors the tool precedent: `content_components` row +
   `doc_plans` 'tool' doc.
2. **Site-plan hook**: a site's selected patterns compile into an **`experience_brief`
   site_specs aspect** consumed by `plan_site` with `roadmap_brief`-style authority. Per-site
   bindings (destination roles → real page ids) are recorded after `sync_pages`;
   `reconcile_site_plan` gains a guard so a re-plan cannot silently drop a page a bound
   experience requires.
3. **Taxonomy**: layered, our own kebab-case enums with industry `aka` names, **seeded from
   harvest only** (tens of entries, bottom-up — not a playbook-first authored catalogue).
4. **vonc pilot**: **wait for tools-api** — the full debate-gauntlet path, no static-cut MVP.
   Harvest pattern #1 (provocations teaser→detail→related) once the pilot is live end to end.
5. **Approval is per experience** via council; the acceptance-test side is formalised (§3.5).
   Storage is DB-based. Travelling docs is the ancestor to build on.

> **CORRECTIONS to the round-1 proposal (recorded per the working-docs rule):** the first
> discussion round proposed a concept-register-style flat-markdown register and
> approval-by-use. The owner corrected both: travelling docs (which documents a tool's
> provenance and direction) is the closer ancestor, storage is DB-based, and approval is per
> experience. Round-2 research confirmed the owner's instinct: `doc_plans` carries no
> `site_id` (already library-level), and the criteria fence's read-time-only parsing is a
> live failure class that per-experience formalised approval addresses.

## 3. Design

### 3.1 Units — two register-entry kinds

- **`component-contract`** — what one approved component itself does (the carousel: swipe
  advances; "read more" expands in place). Attached to a component family via
  `section_types`.
- **`micro-journey`** — a cross-page fragment (card → entity-detail page → related
  links/tools). References component roles and **destination roles**, never concrete pages.

Level-4 site journeys/funnels remain the experience loop's per-site EXPERIENCE_PLAN unit —
composed *from* register entries once the register exists, with the council reviewing the
binding and glue rather than re-litigating invented behaviour.

### 3.2 Taxonomy — four levels, two of them register entries

1. **Interaction primitives** — controlled verb vocabulary: `reveal`, `navigate`, `submit`,
   `filter`, `sort`, `step`, `play`, `dismiss` (see `design/taxonomy_seed.md`). Vocabulary,
   not entries.
2. **Component contracts** (entries, kind=`component-contract`).
3. **Micro-journeys** (entries, kind=`micro-journey`; industry names as `aka`:
   master-detail, progressive-disclosure, hub-and-spoke, teaser→article→related,
   search→results→detail, wizard).
4. **Site journeys/funnels** — NOT register entries; the experience loop's unit. Funnel-stage
   vocabulary (`awareness|consideration|conversion`) is adopted from the dormant
   `site_flows`/`flow_pages` schema (§4 supersession note).

Queryable axes are **columns**, deliberately shaped so the component selector's scoring
pattern transfers: `kind`, `primitives[]`, `section_types[]`, `destination_roles[]`,
`funnel_stage`, `suitable_site_types[]`.

### 3.3 Destination roles + per-site bindings (the "probably should go" mechanism)

Base entries name destination **roles**, built on the existing `page_type`/`role` vocabulary
(`index, content, landing, entity-directory, entity-page, tool, guide, section-index,
blog-index, blog-post, game…`) plus an entity/topic parameter — e.g.
`entity-page(card.entity)`, `tool(topic)`. **A base entry never carries a concrete URL**
(bug 045 lesson: static fallbacks re-apply on every render and cannot be overridden —
bindings are per-site data, the base is invariant).

A per-site **binding** resolves each role to a real page id after `sync_pages`. Bindings give:
- `resolve_internal_links` a lookup instead of a label-blind nav-order guess (bug 023's root
  cause; the intent-matching upgrade LNK-011 reserved its agent boundary for);
- `misdirected_cta`/`dead_controls` a **declared** expected destination to check against,
  instead of reverse-engineering intent from label text;
- `link_registry` its missing intended-target counterpart (today it is post-hoc extraction
  only; planning-time `target_page_id` was never populated — LNK-019).

### 3.4 Contract shape — every trigger names an observable outcome

Per trigger: `control role → interaction primitive → OBSERVABLE outcome → destination role
(if navigating)`. The observable-outcome requirement is load-bearing: the gauntlet correction
(owner verdict, 2026-07-22: wiring a dead control to an invisible effect "fixed the dead-link
LETTER and missed the POINT") showed removing a dead control is not giving it real behaviour.
The experience council's journeys critic already enforces this per plan; the register
enforces it per entry, so a hollow binding cannot be recorded as an experience at all.

### 3.5 Criteria templates — the formalised acceptance side (owner ruling)

Criteria are parameterised templates in the same vocabulary as the existing schema-v0 checks
(`selector_exists`, `selector_count`, `interaction`, `asset_loads`, `page_status_ok`;
Tier-4-only types beyond). Placeholders reference the entry's `binding_schema`. Three
validation moments (full spec: `design/criteria_template_schema_v1.md`):

- **Write time** — template must parse against the versioned schema AND every placeholder
  used must be declared in `binding_schema` (closure check). This is new discipline: today
  the criteria fence is parsed at read time only, and stale criteria / unfilled `-EDIT`
  placeholders / silently-empty unclosed fences are all live, repeatedly-sighted failure
  classes (travelling_docs/README_where_we_are.md).
- **Bind time** — every placeholder bound; bound selectors anchor-resolvable against the
  site's actual components (the check_tool_acceptance anchor rule).
- **Run time** — Tier-2 static immediately; Tier-4 journey execution when the browser-runner
  T5.1 navigate extension lands (not started; owned by the experience-loop/travelling-docs
  threads; G9: today each URL is a fresh browser).

### 3.6 Substrate

- **`experience_patterns`** table: draft DDL in `design/ddl_experience_patterns.sql`.
- **`site_experiences`** bindings table: site_id, pattern, bindings jsonb, status,
  built_from_plan_version (so binding staleness is detectable the same way page staleness
  is).
- **Travelling doc per entry**: new subject_type **`experience-pattern`**, subject_key = the
  pattern name. Carries direction, provenance, decisions, council trail (doc_notes).
- **The 184 lesson (bugs_open/064)**: the subject_type contract now has FOUR enforcement
  points (DB CHECK on doc_plans + doc_notes, `docResolveSubject` shared by all three doc
  actions, `persist_diagnosis_note`'s separate allowlist) and every addition so far has
  missed at least one (163 missed persist_diagnosis_note; 184 missed the Go gate entirely,
  leaving 'action' rows unreachable through the doc actions). The change-set that adds
  `experience-pattern` must move ALL of them together, and should fix 064 in passing —
  enumerated in `design/subject_type_addition.md`.

### 3.7 Site-plan integration

The page set is decided in one place — build-site-planner's `plan_site` step, reading
`site_specs` aspects. Integration follows the existing hooks, no new seams invented:

1. **Plan time**: selected patterns compile into an **`experience_brief`** aspect. Authority
   semantics copy `roadmap_brief` ("ROADMAP OVERRIDES THE COMPONENT LIST… the roadmap is the
   authority", 053_build_site_planner.sql): required pages by role (+ parent_section where
   the pattern is hub-shaped) and required section_types. Unknown section_types already fall
   through to `needs_new_component` downstream.
2. **Bind time**: after `sync_pages`, roles resolve to page ids → `site_experiences` rows.
   Deterministic where possible (role = page_type + parent_section + entity match), flagged
   for review where ambiguous.
3. **Re-plan safety**: `reconcile_site_plan` gains a bound-experience guard alongside the
   existing owned-page guard (`rebuild_policy='owned'` → `owned_page_review`): a re-plan that
   would drop a page a bound experience requires surfaces instead of silently orphaning the
   journey.
4. **Selection**: which patterns a site gets is proposed from `classification`/archetype +
   `suitable_site_types` scoring (the component-selector shape), recorded in the
   `experience_brief`. First consumer is wired in the SAME change that creates the register —
   a register nothing selects from is dead stock (prior-art librarian's DORMANT-MACHINERY
   class; the brochure library's prompt-seam landmine).

### 3.8 Approval lifecycle (owner ruling: per experience)

`draft → approved → proven`:
- **draft** — harvested or authored; visible, **unselectable**.
- **approved** — a pattern-scoped council pass: journeys/observable-outcome critic,
  contracts critic (criteria template provably agrees with the declared contract), honesty
  seat (hard veto, inherited from the experience council), prior-art/reuse seat (near-
  duplicates merge rather than accumulate). Reuses the experience-council machinery
  (167-family seeds), pattern-scoped seat set; verdicts land in doc_notes under the entry's
  travelling doc.
- **proven** — evidence upgrade, not a review outcome: first live green run of the entry's
  bound criteria on a real site.

## 4. Prior art — reuse and explicit supersessions

**Reused**: travelling docs (substrate; the tool precedent of register-row + travelling doc);
`roadmap_brief` (authority-override precedent); component-selector scoring shape (selection);
experience loop's journey vocabulary + council machinery (levels 4 + approval); the criteria
fence schema v0 + anchor rule (checks); `page_type`/`role` vocabulary (destination roles).

**Superseded — `site_flows` / `flow_pages`** (dormant tables: `stage_in_narrative`
awareness/consideration/conversion, `sequence_order`; **zero references in `platform/`**).
This design supersedes them as the cross-page-journey substrate. We adopt their funnel-stage
vocabulary as the `funnel_stage` column and DO NOT build on the tables (they predate the
site_plan machinery and nothing reads them). If Phase 2 archaeology finds live value, revisit
— recorded here so a second account of "pages in a journey" is not left ambiguous.

**Adjacent, not dependencies**: `features_open/013`'s archetype × pattern grid and
`per_site_ai`'s pool blueprints are the same base+fork shape on the product/funnel axis; the
register is that shape on the interaction axis. `misdirected_cta` remains as a backstop for
unbound links; bindings make its job mechanical where they exist.

## 5. Phases and gates

- **P1 (done this session)**: this doc set; `design/` draft artifacts (nothing applied);
  bugs_open/064 filed; memory + commits.
- **P2 (owner go required)**: the real change-set, ONE coherent council-gate submission:
  migration (both tables + subject_type, per `design/subject_type_addition.md` — fixes 064 in
  passing), write-time criteria validation, `experience_brief` consumption in `plan_site`,
  bindings after `sync_pages`, reconcile guard, selection wiring.
- **P3 (gated on the vonc pilot — owner ruling: wait for tools-api)**: harvest. Pattern #1 =
  the provocations teaser→detail→related journey, extracted from the proven live pilot;
  then the brochure components' `behaviour.js` contracts. Pilot critical path (owned by
  gauntlet_dead_cta session): implementer emits valid tools-api code → owner merges PR →
  owner's 4 infra tasks (subdomain, bastion VM, WireGuard peering, Cloudflare tunnel) →
  smoke-POST → re-fire 092 → session-driven T4 build.
- **P4**: first bound site end to end; Tier-2 intent checks live. Tier-4 journey acceptance
  arrives with T5.1 (separate image, separate thread).

## 6. Landmines

- **Subject_type split contract** — bugs_open/064; four enforcement points move together.
- **Dead stock** — wire the first consumer in the creating change (§3.7.4).
- **Base/binding blur** — base entries carry no concrete values (bug 045 class).
- **Criteria drift** — write-time validation (§3.5) is the counter to the stale-criteria and
  `-EDIT` classes; rekey coverage must extend to the new subject_type (today
  RekeyTravellingDocs is wired for 'tool' renames only).
- **Ownership** — the vonc pilot and tools-api belong to the gauntlet_dead_cta session;
  this workstream consumes its output (pattern #1), it does not drive the pilot.
