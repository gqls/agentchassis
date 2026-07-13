# 030 — Phase 1 Plan: Plan as Declarative Artefact, Reconciler, Path Resolver

Companion to doc 029. This doc records the answers to 029's four open
questions, the structural shifts we're committing to that diverge from
029's leans, and the concrete ordering of work for Phase 1.

Phase 0 (canonicalisation helper, item_key alignment) is done. The
gamesdesign.co.uk re-adopt run on 2026-05-04 confirmed dedup works:
zero `adoption_page_*` keys, no duplicated `needs_page:*` rows, 12
canonical pages from a 14-item adoption emission.

What Phase 0 doesn't fix is role assignment. The 2026-05-04 run wrote
`tools` / `guides` / `games` rows with role `content` and url
`/<slug>.html`, because the LLM labelled them `content` despite the URL
pattern being clearly `section-index`. Phase 1 makes role assignment a
deterministic concern, not an LLM one.

---

## Decisions on doc 029's open questions

### Q1. Plan storage — separate `site_plans` schema, not `site_specs` aspects

Doc 029 leaned `site_specs` aspects on the basis that the existing
versioning machinery is already there. We're rejecting that lean. Two
reasons:

1. **Conceptual.** `site_specs` is the strategic/aspirational layer
   produced by agents reasoning about the site (strategy, briefing,
   identity, voice, classification, design_intent). Plans are
   operational artefacts derived from specs by the plan-builder. A page
   list with roles and nav order is not a "spec" in the same sense as
   "the brand voice is technical and direct." Conflating them in one
   table forces every reader to filter by aspect prefix.

2. **Scale.** Sites today already carry up to 20 current `site_specs`
   rows. Per-page partials in `site_specs` would multiply this with
   page count. Anticipated scale is 1000+ pages per site (eventually
   10k+ for product-listing sites). A monolithic structure-partial
   JSONB blob at that size is awkward to update incrementally — every
   single-page change rewrites the whole blob. Normalised storage is
   the right shape from the start.

The schema is four plan-domain tables, all row-shaped for scale:

```
site_plans
  id uuid pk
  site_id uuid fk → sites
  is_current boolean default true
  created_at, superseded_at, created_by, source_agent, notes
  -- one current row per site; older rows kept for audit
  UNIQUE (site_id) WHERE is_current = true

site_plan_pages
  id uuid pk
  plan_id uuid fk → site_plans ON DELETE CASCADE
  name text                      -- canonical name (post-canonicalisation)
  role text                      -- 'tool' | 'guide' | 'section-index' | ...
  slug text                      -- raw stem before canonicalisation
  url text                       -- canonical url from CanonicalisePage
  parent_section text            -- nullable; structural parent for nested pages
  title, meta_description, nav_label text
  in_header boolean default true
  in_footer boolean default true
  nav_order int
  UNIQUE (plan_id, name)

site_plan_sections
  id uuid pk
  plan_id uuid fk → site_plans ON DELETE CASCADE
  page_name text                 -- denormalised for join-free reads
  ordering int                   -- 0-based position in the page's section list
  component_name text            -- 'hero' etc — for data-component attribute
  -- resolved provenance IDs, nullable, filled by downstream resolver
  component_version_id uuid
  palette_id uuid
  layout_id uuid
  typography_set_id uuid
  UNIQUE (plan_id, page_name, ordering)

site_plan_directives
  id uuid pk
  plan_id uuid fk → site_plans ON DELETE CASCADE
  scope text                     -- 'site' | 'page' | 'section'
  scope_ref text                 -- NULL | page_name | '<page_name>:<ordering>'
  category text                  -- 'design' | 'content' | 'voice' | 'style' | 'structural'
  subject text                   -- 'palette.character', 'voice.register', 'writing_rules', ...
  directive text                 -- the actual guidance (paragraph or bullet)
  ordering int default 0         -- preserves LLM order for multi-cardinality subjects
  source text                    -- 'llm' | 'classifier' | 'adoption' | 'manual'
  locked_at timestamptz          -- HITL lock (pattern from page_components / site_components)
  locked_by text
  -- index on (plan_id, scope, scope_ref) for cascade reads
  -- partial index on (plan_id) WHERE locked_at IS NOT NULL for lock-transfer queries
```

`site_plan_pages` and `site_plan_sections` are the row-per-thing
structural tables. `site_plan_directives` is the storage form of all
guidance — design / content / voice / style / structural — at any of
three scopes. JSONB blobs were considered and rejected because at
anticipated scale (10k+ products, thousands of directives per plan)
loading whole blobs to read one slice is wasteful, surgical HITL edits
become hard, and lock transfer at meaningful granularity is impossible.

This is more upfront cost than reusing `site_specs`. The cost is a
schema migration and a few new helpers. The payoff is a clean data
model that handles the scale case without retrofit, and a directive
table that supports per-row HITL granularity natively.

### Q2. Plan-builder LLM call shape — one call producing structure + design_direction + content_strategy together

Earlier draft of this doc proposed three sequential LLM calls. Looking
at the existing `build-site-planner` agent, that lean was wrong.

The current planner produces `pages`, `design_intent`, and
`content_direction` together as one JSON shape. There are good reasons
to keep that:

1. **Coherence.** The LLM gets to make design and content guidance
   *consistent with* the page list it just decided. Three independent
   calls would either need to chain (each call sees the previous
   output, adding complexity) or run independently (risking design /
   content guidance that doesn't match the structure).

2. **Latency is not the priority** (per your direction). The single
   call is no slower than three sequential. Per-partial retry
   granularity is achievable by re-running the same call rather than
   making it three steps.

3. **No evidence of need.** The existing single-call pattern is in
   production and has not surfaced retry-granularity problems. Three
   calls solves a problem we don't have.

The change for Phase 1 is therefore not "split the call" but "redirect
where the call's outputs are written":

- `pages` array → `site_plan_pages` rows (via canonicaliser + role validator)
- nested `design_direction` / `content_strategy` JSON fields → flattened into `site_plan_directives` rows (via the JSON-to-directive mapping)
- nothing written to `site_specs.design_intent` or `site_specs.content_direction` by the planner anymore (those slots stay strategic — see naming note below)

### Strategic vs plan-time naming

The existing system writes `site_specs.design_intent` from three
agents: `domain-research-classifier` (strategic, brand-level),
`site-adoption-agent` (captured from existing CSS), and
`build-site-planner` (plan-time direction informed by strategy +
briefing). The third overwrites the first two. That conflation is what
Phase 1 fixes.

The clean split:

| Concept | Authored by | Lives in |
|---|---|---|
| Strategic design_intent (brand-level identity) | classifier or adoption | `site_specs.design_intent` |
| Plan-time design_direction | plan-builder | `site_plan_directives` (category=design) |
| Strategic content_direction (brand voice) | classifier or adoption | `site_specs.content_direction` |
| Plan-time content_strategy | plan-builder | `site_plan_directives` (category=content) |

Plan-builder reads strategic design_intent / content_direction from
`site_specs` as input, then emits its own plan-time direction into
`site_plan_directives`. No more overwriting. Downstream agents that
read strategic guidance keep reading `site_specs`; agents that want
plan-time direction read directives.

### Q3. URL paths — `CanonicalisePage` (Phase 0 helper, extended); link/nav agents own drift

URL paths sit on a clear boundary that this section makes explicit.

**Decision at plan-write time** is owned by `CanonicalisePage` in
`datahelpers/page_canonical.go` — the same helper introduced in
Phase 0 for name canonicalisation, now extended to honour an optional
`ParentSection` field for nested-URL synthesis. Plan-builder's
`write_site_plan` action calls it once per page and writes the
returned `(name, url, page_type)` triple into `site_plan_pages`.

The earlier draft of this doc proposed a separate `BuildPageURL` helper
sitting alongside `CanonicalisePage`, on the basis that the Phase 0
helper shouldn't be touched. That argument was overly cautious: the
Phase 1 change is purely additive (a new optional field on
`PageDescriptor`, used only when non-empty), adoption call sites pass
empty `ParentSection` and get identical Phase 0 behaviour, and keeping
two helpers that both produce URLs created exactly the kind of dual-
contract surface the development guide warns against. Consolidated.

**Drift after pages are realised** is owned by the existing link/nav
layer (doc 024). When a later plan version changes a URL, `pages.url`
diverges from `link_registry.target_url`. This is what
`link-validator`, `nav-link-fixer`, `internal-linker`, and the
`broken_internal_link` work item exist to handle. We do not duplicate
that machinery here.

The boundary in one sentence: `CanonicalisePage` produces URLs at
plan-write time; the link/nav agents reconcile URLs that have drifted
between plan and realised state. They do not overlap.

**What we considered and rejected**: a `path-policy-agent` that would
own URL conventions site-wide and be the only writer of URLs. Rejected
because URL synthesis from `(role, slug, parent_section)` is
deterministic. Wrapping a deterministic function in an agent is
ceremony. If URL conventions need to change site-wide, change the
helper and trigger a plan rebuild — same effect, no extra agent
identity.

**Role assignment** is the part that does need care, because the LLM
mislabels (gamesdesign run: `tools` page came back as `role: content`
when it's structurally a section index). The role validator is a
separate helper next to `page_canonical.go`, but its logic — URL-shape
inference, parent-section detection — overlaps with what
`extract_and_sync_links` and `populate_nav_tables` already do for
realised state. Keeping it in `datahelpers/` keeps it reachable from
both plan-write code and link/nav code without circular packaging.

The responsibility split:

| Component | Decides | Determinism |
|---|---|---|
| plan-builder (LLM) | Which pages exist; role per page; nav membership/order | LLM |
| `CanonicalisePage` (Go, `datahelpers/`) | Canonical `(name, url, page_type)` for `(role, slug, parent_section)` | Deterministic |
| role validator (Go, `datahelpers/`) | Correct role when LLM mislabels | Deterministic |
| reconciler (Go) | Work items to emit from desired-vs-realised diff | Deterministic |
| link/nav agents | Reconcile URLs that drift between plan and realised state | Existing per doc 024 |
| navigation builder | What appears in header/footer/sidebar from canonical set | Deterministic |

This is the structural fix for the gamesdesign mislabel. The plan-builder
LLM is allowed to be sloppy about role; the role validator catches and
corrects before anything is persisted to `site_plan_pages`. Same
principle as Phase 0's canonicalisation helper, scaled up.

The role validator's logic, sketched:

```
if role == 'content' or empty:
    if slug names a section that has child pages with parent_section = slug:
        → role := 'section-index'
    elif url_hint is /<slug>/index.html:
        → role := 'section-index'
    elif url_hint is /<dir>/<slug>/index.html (well-formed nested):
        → role := infer from <dir>  ('guide' if dir='guides' etc)

if role declared but conflicts with URL hint:
    → trust URL hint when well-formed; log conflict
```

The path resolver lives in `datahelpers/` next to `page_canonical.go`.
The role validator either lives there too or as a sibling file. Both
consume the LLM's raw output and produce canonical fields.

### Q4. Reconciler cadence — both, scheduled is the workhorse

Event-driven trigger fires when plan-builder writes a partial or when
adoption completes. Scheduled tick (5 min initial) catches missed
events and applies cycle-budget logic.

Scheduled is more important. The event-driven side is an optimisation
for responsiveness on direction changes. If we have to ship one,
scheduled.

The scheduled tick needs `sites.last_reconciled_at timestamptz` to
avoid re-scanning sites that just reconciled. A tick walks sites where
`last_reconciled_at` is older than the tick interval, runs reconcile,
updates the column.

---

## Directive cascade and brief assembly

`site_plan_directives` is the storage form of guidance. Consumers do
not read it directly; they read assembled briefs via a Go helper
(`datahelpers/page_brief.go`). The helper does two things:

1. **Walks the cascade.** For a section brief: site-scope directives
   first, then page-scope for the section's page, then section-scope.
   For a page brief: site-scope, then page-scope. For a site brief:
   just site-scope.

2. **Applies cardinality semantics.** Some directive subjects are
   single-valued (`voice.register`, `palette.character`) — the closer
   scope overrides the broader scope on the same `(category, subject)`.
   Some are multi-valued (`writing_rules`, `things_to_avoid`,
   `key_terms`) — directives at all scopes accumulate in declaration
   order.

   Cardinality is decided by a small lookup table in the renderer, not
   stored on the directive row. New subjects default to single-valued
   unless added to the multi list. Code-side decision keeps the data
   simple; if subject vocabulary grows organically beyond the renderer
   developer's knowledge, promoting to a row column is a one-migration
   change.

The brief renderer's output shape is readable text suitable for an LLM
prompt. Example for the hero section of `tool-ttk-calculator`:

```
## Section: hero (tool page)

### Style
- Dark scheme; near-black background, cyan accent
- Typography: Segoe UI / Roboto; monospace for numerical output

### Voice
- Terse, technical, imperative
- No "sign up" CTAs; primary CTA = link directly to the tool

### Content goals
- One-line value proposition specific to this tool
- One sentence on why guessing is costly
```

The text is short enough to keep in a prompt. The LLM can expand each
bullet into prose if it chooses, or keep them as bullets — the
directives leave room for either.

---

## Lock transfer across plan rebuilds

HITL is supported via `locked_at` / `locked_by` columns on
`site_plan_directives`, mirroring the established pattern from
`page_components` and `site_components`. A locked directive is one a
human has approved or edited; subsequent plan rebuilds must not silently
overwrite it.

Lock transfer happens **only in `write_site_plan`**. No other reader
needs to know about the mechanism. The procedure on plan rebuild:

1. Write the new `site_plans` row and all fresh page / section /
   directive rows from the LLM's output.
2. Query the previous current plan's directives where
   `locked_at IS NOT NULL`.
3. For each locked previous directive, find the matching new directive
   by composite key `(scope, scope_ref, category, subject, ordering)`.
4. If a match exists: copy `locked_at`, `locked_by`, and (if the
   client-agreed text was different from the LLM's new text) the
   directive text itself onto the new row. The HITL-approved version
   wins.
5. If no match exists: the locked directive's slot has gone away. Log
   it. The lock is dropped because there is no row for it on the new
   plan.

For section scope, `scope_ref` is the composite `<page_name>:<ordering>`
form. This is stable across plan rebuilds when the LLM produces the
same page with sections in the same order — the common case for the
same brief. If the LLM reorders sections, locks can transfer to the
wrong row; this is a known imperfection accepted in exchange for the
simplicity of not maintaining a separate override table that every
reader would need to know about.

The wider durability story for client-agreed plans is: keep the
previous plan as a history row (which we already do via
`is_current = false` + `superseded_at`), so any directive that didn't
transfer can be retrieved manually if needed. The lock-transfer
mechanism handles the routine case; the history row handles the audit
case.

---

## Plan-builder cascade (replaces today's site-planner emit-and-queue)

The current `build-site-planner` does both planning and work-item
emission, and writes plan-time direction into `site_specs`.
Phase 1 splits these. The new flow:

```
adoption (or HITL direction change) →
  emits needs_strategy →
  domain-strategist writes site_specs/strategy →
    emits needs_briefing →
    briefing-agent writes site_specs/briefing →
      emits needs_site_plan →
      plan-builder runs ONE LLM call producing pages + design_direction +
        content_strategy in one coherent JSON shape, then via write_site_plan:
          - canonicalise + role-validate pages → site_plan_pages rows
          - resolve sections from page section lists → site_plan_sections rows
          - flatten design_direction / content_strategy nested JSON →
            site_plan_directives rows (site / page / section scopes)
          - transfer locks from previous plan's locked directives
          - mark new site_plans row current
        → emits no work items
        ↓
        reconciler runs (event-driven on plan-current change OR scheduled)
        reads site_plan_pages + pages + open site_work_items
        emits needs_page:<name> for diff
```

Per doc 029's "What NOT to do" list, we keep the agent_type
`build-site-planner` (and the older `site-planner` used by other
pipelines). Renaming is a deferred migration. The behaviour change is
internal: the workflow stops calling `WriteBuildItemsAction` and stops
writing `site_specs/design_intent` + `site_specs/content_direction`,
and instead calls `write_site_plan` which routes outputs into the
plan-domain tables.

---

## Phase 1 work order

| Step | Item | Notes |
|---|---|---|
| 1 | **Schema migration**: `site_plans`, `site_plan_pages`, `site_plan_sections`, `site_plan_directives` tables; `pages.built_from_plan_version uuid NULL`; `sites.last_reconciled_at timestamptz NULL`. | Single migration file. Nullable additions are safe. |
| 2 | **Extend `CanonicalisePage` with `ParentSection` + role validator** in `datahelpers/`. Pure Go, table-driven tests. The Phase 0 helper is patched additively; role validator is a new sibling file. | First code change. Independent of any other Phase 1 piece. |
| 3 | **`write_site_plan` Go action**. Takes plan-builder LLM output (one call producing pages + design_direction + content_strategy together), runs role validator and canonicaliser per page, flattens nested design / content JSON into directive rows, writes one `site_plans` row + N `site_plan_pages` rows + M `site_plan_sections` rows + K `site_plan_directives` rows in one tx. Transfers locks from previous current plan. | Replaces `WriteBuildItemsAction`'s role and the planner's writes to `site_specs.design_intent` / `site_specs.content_direction`. |
| 4 | **`build-site-planner` workflow change**. Single LLM call (existing prompt, mostly unchanged — outputs are unchanged, only the destination changes). Terminal step is `write_site_plan` instead of `write_build_items`. Removes `write_design_intent`, `write_content_direction`, `write_plan_spec` steps. | SQL change in `agent_definitions` for `build-site-planner`. The older `site-planner` (used by `pageflow-builder`) is left alone. |
| 5 | **`reconcile_site_plan` Go action**. No LLM. Reads `site_plan_pages` + `pages` + open `site_work_items`. Emits `needs_page:<name>` for diff. Updates `sites.last_reconciled_at`. | Cycle budget config TBD; v1 has no budget cap. |
| 6 | **Reconciler triggers**. Event-driven via Kafka (plan-builder completion). Scheduled via existing kafka-scheduler with a 5-min tick. | Both. |
| 7 | **Brief renderer helper** in `datahelpers/page_brief.go`. Reads directive cascade for site / page / section consumers, applies cardinality rules, emits prompt-shaped text. Pure Go, no LLM. | Becomes the read path for content-writer / webdesign-agent prompts that want plan-time guidance. |
| 8 | **Wire brief renderer into page-build-handler**. Replace today's reads of `site_specs.design_intent` (plan-time portions) with calls to the brief renderer for the relevant page / section. Strategic site_specs reads are unchanged. | Step 7 must be live first. |
| 9 | **Drift detection**. Pages where `built_from_plan_version` != current plan's `id` are flagged for rebuild. | Reconciler enhancement; same emission path. |

### Note on the `target_site_id` input field name

`write_site_plan` declares its required input as `target_site_id`, not `site_id`. This is deliberate.

`ExtractActionInputs` runs a nested-source loop late in its resolution chain (after explicit dot-paths, top-level lookups, and deprecated `*_field` configs) that walks `current_page`, `rerender_pages`, `site_record`, `input_data` for any unresolved field. The loop fires for both required and optional fields, not just optional. If a caller's `input_mapping` doesn't explicitly map `site_id` AND the cascade has populated `site_record.site_id` in CollectedData, an action with a `site_id` field can silently pick up that nested value.

In the normal cascade flow this resolves to the same UUID the caller intended — the latent collision rarely bites — but the failure mode is silent and hard to debug when it does. New code should use field names that don't collide with database columns or with the nested-source list (`current_page`, `site_record`, `input_data`). `target_site_id` satisfies both criteria and reads as "the target of this write" at call sites:

```json
"input_mapping": {
  "target_site_id": "site_record.site_id"
}
```

Older actions in the codebase use `site_id` Required and ship in production. Renaming them would be a much larger change than the latent risk warrants. The convention here is: **leave existing code alone, but write new code with collision-free names**.

---

Steps 1–2 land first because they unblock everything. Step 3 needs
step 2. Steps 4 and 5 can be developed in parallel once step 3 is in.
Step 6 needs step 5. Step 7 lands when convenient — first per-partial
LLM call is its trigger. Steps 8 and 9 are downstream, after the
reconciler is observably working.

---

## What stays in `site_specs`

Strategy. Briefing. Identity. Classification. Voice. Tone. Design
intent. Anything produced by an agent reasoning about *the site as a
whole*. These remain `site_specs` aspects with the existing versioning.

What moves out: the page list, design direction (the LLM's plan-time
view of design, distinct from `design_intent` which is strategy-time),
content strategy (cross-page voice rules at plan time, distinct from
`voice` which is strategy-time), per-page briefs.

The naming distinction matters. `site_specs/design_intent` is "what
character should the site project" (strategy). `site_plan_partials/
design_direction` is "given that intent and the realised structure,
how should pages be designed" (plan). One feeds the other.

---

## Risks and what to watch

- **Role validator misfires.** If the validator overrides a correctly-
  labelled `content` page to `section-index` because of accidental URL
  similarity, it produces wrong pages. Mitigation: only override when
  the LLM-supplied URL clearly matches the section pattern AND the
  slug appears as a `parent_section` for at least one other page in
  the same plan. Both signals required.

- **Plan version churn.** Every plan-builder run writes a new
  `site_plans` row, supersedes the old. If the reconciler reacts to
  every supersession, it could spam work items. Mitigation: reconciler
  diffs current plan against last-realised state, not against last
  plan; supersession alone doesn't trigger work.

- **Scale not yet stress-tested.** 1000+ page sites are anticipated,
  not realised. A 12-page test site won't surface row-volume issues.
  Worth running a synthetic 1000-page plan through reconcile once step
  5 is live.

- **`site-planner` rename deferred.** Workflow definitions reference
  `site-planner` by type name. Renaming is its own migration. Phase 1
  keeps the type name; behaviour changes internally.

---

## What this doc is not

Not a sequencing of when within Phase 1 each step ships to production.
Steps 1–6 are the minimum for the new flow to work end-to-end. Steps
7–9 are quality-of-life and observability improvements that ride after
the core flow is proven.

Not a commitment on Phase 2 (auditors reading the plan). Phase 2 is
out of scope for this doc.
