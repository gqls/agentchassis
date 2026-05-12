# FOCUS — `site_specs` vs `site_plan`: what lives where and why

A short reference for the two-layer architecture, written after I conflated
them in the imagery Phase 2G design and the user (correctly) pulled me back.

This isn't new architecture — it's documented across docs 021, 028, 029, 030,
and the adoption pipeline patches. This focus doc gathers the answers in
one place so future imagery / planner / agent design work doesn't make the
same mistake.

---

## The two layers

Two distinct storage layers serve distinct purposes:

### `site_specs` — strategic, brand-level, slow-changing

Per-aspect rows, one current per (site_id, aspect), versioned via
`is_current` + `superseded_at`. Stored as JSONB on each row.

Aspects: `identity`, `classification`, `content_direction`, `design_intent`,
`strategy`, `seo`, `maintenance`, plus adoption's `site_archetype` and
`design_reference`.

**The contract:** what the site IS, brand-defining direction.

### `site_plan` (four tables) — build-time, concrete, rebuilt per plan

`site_plans` (one current per site), `site_plan_pages` (one per page),
`site_plan_sections` (one per section per page, ordered),
`site_plan_directives` (guidance with scope and category).

Row-shaped for scale (up to 10k+ pages anticipated per site).

**The contract:** what we BUILD this time, and in what shape.

The two layers exist because they have different ownership, different
lifecycles, and different scale characteristics.

---

## Ownership (doc 028's contract)

Each spec aspect has exactly one owning agent under normal flow:

| Aspect | Owner |
|---|---|
| `identity`, `classification`, `content_direction`, `design_intent`, `seo`, `maintenance` | **classifier** (`domain-research-classifier`) |
| `site_archetype`, `design_reference` | **adoption** (`site-adoption-agent`) — nobody else writes these |
| `strategy` | **strategist** (`domain-strategist`) |
| `briefing` | **briefing-agent** (`build-briefing-agent`) |
| Site-plan tables (all four) | **planner** (`build-site-planner` / `plan-builder` post-Phase-1) |

Doc 028's three rules, paraphrased:

1. An agent that **reads info it didn't put in the spec** is a bug (someone
   else might not see that info).
2. An agent that **overwrites a spec aspect another agent owns** is a
   category error (breaks the ownership model).
3. An agent that **produces the right output but doesn't write it to the
   spec** is unhelpful (downstream reads the spec, not work-item payloads).

The classifier's "read-and-extend" pattern is the carve-out: when adoption
has written `identity`, `content_direction`, `design_intent`, the
classifier reads them as ground truth and re-emits versions that preserve
the adopted dimensions while adding strategic contributions.

---

## The pipeline cascade

How a domain travels from intake to deployed site (post-Phase-1 target
state, with notes where current state differs):

```
intake-orchestrator
  │  adoption mode? ───┐
  ▼                    ▼
classifier        site-adoption-agent
                  ├── writes site_archetype, design_reference
                  ├── seeds identity, content_direction, design_intent
                  └── emits needs_domain_research → classifier picks up
classifier
  ├── writes identity, classification, content_direction, design_intent, seo, maintenance
  └── emits needs_strategy
domain-strategist
  ├── writes site_specs/strategy
  └── emits needs_briefing
build-briefing-agent
  ├── writes site_specs/briefing
  └── emits needs_site_plan
build-site-planner  (renamed plan-builder post-Phase-1)
  ├── ONE LLM call producing pages + design_direction + content_strategy
  └── write_site_plan action:
       ├── writes site_plans row (new current)
       ├── writes site_plan_pages (per page)
       ├── writes site_plan_sections (per section)
       ├── writes site_plan_directives (site/page/section scope)
       ├── transfers HITL locks from previous current plan
       └── emits no work items
reconciler  (Go action, no LLM)
  ├── reads site_plan_pages vs pages table
  ├── reads open site_work_items
  └── emits needs_page:<name> for the diff (cycle-budgeted)
page-build-handler → ... → page deployed
```

In Phase 1 territory but already partly live:

- `write_site_plan` action exists. The four plan tables exist.
- `build-site-planner` workflow writes both shapes during transition
  (old `site_specs/site_plan` aspect AND new plan tables) — pageflow-builder
  (the older path) still reads the old shape.
- Reconciler is documented in doc 030 but the chassis-side implementation
  has been landing in stages.

---

## Why directives are the right shape for cross-cutting guidance

`site_plan_directives` carries fields like `voice.register`, `palette.character`,
`writing_rules`, `imagery_direction`, with three columns that together
locate guidance:

```
scope     'site' | 'page' | 'section'
scope_ref NULL | page_name | '<page_name>:<ordering>'
category  'design' | 'content' | 'voice' | 'style' | 'structural'
subject   the canonical name of the guidance type
directive the actual text
```

Plus `locked_at` / `locked_by` for HITL, mirroring the page_components /
site_components lock pattern.

Consumers don't read directive rows directly. A Go helper
(`datahelpers/page_brief.go`) **cascades** site → page → section scope and
**applies cardinality semantics** (single-valued subjects override at the
narrower scope; multi-valued subjects accumulate). The output is readable
text for an LLM prompt — short bullet-point briefs.

This is the pattern the user pointed at when they said "model images on
what's already been decided for text, design." Text and design direction
already cascade through directives. Imagery should follow the same shape.

---

## Lock transfer (doc 030)

Plan rebuilds happen on direction changes. HITL approvals can't be lost.

The lock-transfer mechanism, run inside `write_site_plan`:

1. Write the new `site_plans` row and all fresh rows from the LLM output.
2. Query the previous current plan's directives where `locked_at IS NOT NULL`.
3. For each locked previous directive, find the matching new directive by
   composite key `(scope, scope_ref, category, subject, ordering)`.
4. If a match exists: copy `locked_at`, `locked_by`, and (if HITL-edited
   text differs from LLM new text) the directive text itself. HITL wins.
5. If no match: log it, lock drops. The previous plan stays as history
   row (`is_current = false`) for manual recovery.

Any new sibling table that wants HITL on it should adopt the same pattern
(same column shape, same composite key, same handling in `write_site_plan`).

---

## Where imagery currently lives — and where it should

Today:

- **`site_specs.design_intent.imagery_direction`** — site-wide strategic
  guidance (e.g. "warm hand-drawn illustration"). Owned by classifier.
  Read by image-generator's prompt composer (Phase 0.1).
- **`site_plan_directives` scope=site, category=design, subject=imagery_direction**
  — same value, build-time view. Survives plan rebuilds.
- **`site_specs.site_plan.image_prompts`** (a JSONB dictionary inside the
  legacy `site_plan` aspect) — flat dictionary like
  `{logo: "...", hero_home: "...", hero_about: "..."}`. Written by the old
  planner. Read by discovery and image-build-handler.
- **The legacy `image_prompts` is the only place per-image asks live.**
  There is no per-page or per-section imagery in the new plan tables today.

The natural successor (per `ASSESSMENT_imagery_phase_0_1_vs_phase_1_architecture.md`):

> Per-page imagery direction (scope=`page`, category=`design`) becomes
> meaningful for hero variants beyond `hero_home`... straightforward to
> implement once `site_plan_directives` exists.

The structural extension proposed in Phase 2G is a sibling table
`site_plan_imagery` (per the plan doc), mirroring `site_plan_directives`'
scope+scope_ref+locking pattern but with structured columns appropriate
for image generation (kind, asset_key, prompt, style hints).

---

## Quick decision rules

When designing where a new piece of data lives:

- **"What the site IS"** (brand, identity, voice, regulatory constraints,
  industry character, strategic positioning) → `site_specs`.
- **"What we BUILD this time"** (pages we'll generate, sections within
  them, prompts, directives shaping those sections) → `site_plan` tables.
- **Slow-changing, owned by one agent, brand-defining** → `site_specs`
  aspect.
- **Per-build, rebuildable, scoped to a level of the hierarchy** →
  `site_plan_directives` row (free text) or a new sibling table (structured
  data).
- **Cross-cutting guidance that should cascade** → `site_plan_directives`
  + the brief renderer.
- **Per-instance structured data** (image prompts, product cards, tool
  specs) → a sibling plan table mirroring the directive pattern.

---

## Anti-patterns

Things to avoid (collected from mistakes including the one that motivated
this doc):

1. **Putting build-time structured data in `site_specs`.** It violates
   ownership (build-time decisions don't belong to the classifier's aspect
   set) and doesn't get the cascade/lock benefits.
2. **Putting site-wide strategic direction inside per-page plan rows.**
   Forces every reader to walk pages to find what should be one site-scope
   directive.
3. **Free-text directives carrying structured data.** If downstream needs
   to parse the directive text to get `kind` and `asset_key`, the row
   should have those as columns instead.
4. **Reading the spec to learn what work to do.** Work items carry the
   work; specs carry the contract. An agent reads its work item to know
   what to do; reads the spec to know what the answer should look like.
5. **Two writers for the same aspect.** The ownership rule means at most
   one agent owns each aspect under normal flow. The classifier's
   read-and-extend is the carve-out, designed deliberately and bounded
   to the adoption-classifier handover.

---

## Where to read next

- **`021_site_spec_and_classifier.md`** — the unified spec idea, classifier
  vs planner responsibilities, per-aspect versioning.
- **`028_platform_mission_and_pipeline_direction(2).md`** — ownership rules,
  doc 028 failure modes, fidelity dial.
- **`029_site_plan_and_reconciler(1).md`** — why the plan domain is
  separate from the spec domain.
- **`030_phase1_plan_and_reconciler(4).md`** — the four plan tables, the
  directive cascade, the lock-transfer mechanism, the work order for
  Phase 1.
- **`007_adoption_pipeline_v4.md` / `.patch`** — adoption-classifier
  handover, why adoption writes specs but doesn't shortcut the classifier.
- **`ASSESSMENT_imagery_phase_0_1_vs_phase_1_architecture.md`** — already
  flagged that per-page imagery is the natural extension once
  `site_plan_directives` exists.
