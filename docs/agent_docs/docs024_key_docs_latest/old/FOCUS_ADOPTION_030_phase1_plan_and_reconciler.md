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
pattern being clearly `section_index`. Phase 1 makes role assignment a
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

The schema is:

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
  role text                      -- 'tool' | 'guide' | 'section_index' | ...
  slug text                      -- raw stem before canonicalisation
  url text                       -- canonical url (path resolver output)
  parent_section text            -- nullable; for tools nested under tools-index
  in_header boolean default true
  in_footer boolean default true
  nav_order int
  page_data jsonb                -- title, sections list, meta_description, etc.
  UNIQUE (plan_id, name)

site_plan_partials
  id uuid pk
  plan_id uuid fk → site_plans ON DELETE CASCADE
  partial_type text              -- 'design_direction' | 'content_strategy' | 'page_brief:<name>'
  data jsonb
  is_current boolean default true
  -- per-partial versioning, like site_specs
  UNIQUE (plan_id, partial_type) WHERE is_current = true
```

`site_plan_pages` is the row-per-page table that scales — unique index
on `(plan_id, name)` mirrors the role `idx_site_specs_current` plays
for `site_specs`. Per-page brief lives in `site_plan_partials` with
`partial_type = 'page_brief:<name>'`, lazily generated.

This is more upfront cost than reusing `site_specs`. The cost is a
schema migration and a couple of new helpers. The payoff is a clean
data model that handles the 1000+ page case without retrofit.

### Q2. Plan-builder LLM tier — three sequential per-partial calls

Three calls: structure, design direction, content strategy. Per-page
briefs are lazy and not part of the eager build.

Latency is not the priority. Quality and re-runnability per partial
are. If structure fails, retry just structure. If design partial
fails, retry just design. Saves nothing on tokens vs one big call but
buys recovery granularity.

### Q3. URL paths — Go helper at plan-write time; link/nav agents own drift

URL paths sit on a clear boundary that this section makes explicit.

**Decision at plan-write time** is owned by a small Go helper, sibling
to `datahelpers/page_canonical.go`. It takes `(role, slug,
parent_section)` and returns the canonical url string. Plan-builder's
`write_site_plan` action calls it once per page and writes the result
into `site_plan_pages.url`. That's the entire contract. It is a pure
function, ~50 lines, no DB access, no agent identity. Calling it a
"path resolver" oversells what it is — it's a string-format helper.

**Drift after pages are realised** is owned by the existing link/nav
layer (doc 024). When a later plan version changes a URL, `pages.url`
diverges from `link_registry.target_url`. This is what
`link-validator`, `nav-link-fixer`, `internal-linker`, and the
`broken_internal_link` work item exist to handle. We do not duplicate
that machinery here.

The boundary in one sentence: the helper produces URLs at plan-write
time; the link/nav agents reconcile URLs that have drifted between
plan and realised state. They do not overlap.

**What we considered and rejected**: a `path-policy-agent` that would
own URL conventions site-wide and be the only writer of URLs. Rejected
because URL synthesis from `(role, slug, parent_section)` is
deterministic. Wrapping a deterministic function in an agent is
ceremony. If URL conventions need to change site-wide, change the
helper and trigger a plan rebuild — same effect, no extra agent
identity.

**Role assignment** is the part that does need care, because the LLM
mislabels (gamesdesign run: `tools` page came back as `role: content`
when it's structurally a section index). The role validator is also a
Go helper next to `page_canonical.go`, but its logic — URL-shape
inference, parent-section detection — overlaps with what
`extract_and_sync_links` and `populate_nav_tables` already do for
realised state. Keeping it in `datahelpers/` keeps it reachable from
both plan-write code and link/nav code without circular packaging.

The responsibility split:

| Component | Decides | Determinism |
|---|---|---|
| plan-builder (LLM) | Which pages exist; role per page; nav membership/order | LLM |
| URL helper (Go, `datahelpers/`) | Canonical `url` for `(role, slug, parent_section)` at plan-write time | Deterministic |
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
        → role := 'section_index'
    elif url_hint is /<slug>/index.html:
        → role := 'section_index'
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

## Plan-builder cascade (replaces today's site-planner emit-and-queue)

The current `site-planner` does both planning and work-item emission.
Phase 1 splits these. The new flow:

```
adoption (or HITL direction change) →
  emits needs_strategy →
  domain-strategist writes site_specs/strategy →
    emits needs_briefing →
    briefing-agent writes site_specs/briefing →
      emits needs_site_plan →
      plan-builder runs three calls:
        (a) structure → site_plan_pages rows
        (b) design_direction → site_plan_partials
        (c) content_strategy → site_plan_partials
      writes site_plans row, marks plan current
      → emits no work items
        ↓
        reconciler runs (event-driven on plan-current change OR scheduled)
        reads site_plan_pages + pages + open site_work_items
        emits needs_page:<name> for diff
```

Per doc 029's "What NOT to do" list, we keep the agent_type
`site-planner`. Renaming to `plan-builder` is a deferred migration. The
behaviour change is internal: the workflow stops calling
`WriteBuildItemsAction` and instead calls a new `write_site_plan`
action.

---

## Phase 1 work order

| Step | Item | Notes |
|---|---|---|
| 1 | **Schema migration**: `site_plans`, `site_plan_pages`, `site_plan_partials` tables; `pages.built_from_plan_version uuid NULL`; `sites.last_reconciled_at timestamptz NULL`. | Single migration file. Nullable additions are safe. |
| 2 | **URL helper + role validator** in `datahelpers/`. Pure Go, table-driven tests. Sibling to `page_canonical.go`. | First code change. Independent of any other Phase 1 piece. |
| 3 | **`write_site_plan` Go action**. Takes plan-builder LLM output, runs role validator, runs URL helper, writes one `site_plans` row + N `site_plan_pages` rows. Marks previous `site_plans` row for site as superseded. | Replaces `WriteBuildItemsAction`'s role in the planner workflow. |
| 4 | **Plan-builder workflow change**. Three sequential LLM calls (structure → design_direction → content_strategy). Last step is `write_site_plan`. No work-item emission. | SQL change in `agent_definitions` for `site-planner`. |
| 5 | **`reconcile_site_plan` Go action**. No LLM. Reads `site_plan_pages` + `pages` + open `site_work_items`. Emits `needs_page:<name>` for diff. Updates `sites.last_reconciled_at`. | Cycle budget config TBD; v1 has no budget cap. |
| 6 | **Reconciler triggers**. Event-driven via Kafka (plan-builder completion). Scheduled via existing kafka-scheduler with a 5-min tick. | Both. |
| 7 | **LLM tier annotation infrastructure**. `StepConfig.Config.llm_tier` → routes to `large` / `medium` / `small` endpoint. | Cross-cutting; lands when first per-partial action wants tier annotation. |
| 8 | **Lazy per-page brief generation**. `build_page_brief` step in page-build-handler. Reads structure partial, generates `site_plan_partials/page_brief:<name>` if missing. | Step 5 must be live first. |
| 9 | **Drift detection**. Pages where `built_from_plan_version` != current plan's `id` are flagged for rebuild. | Reconciler enhancement; same emission path. |

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
  labelled `content` page to `section_index` because of accidental URL
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
