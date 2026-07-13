# 029 — Site Plan as Declarative Artefact, Reconciler, and LLM Tiering

How the platform separates aspirational direction from realised pages, why the
current planner produces duplicates, and the staged plan to fix it without a
big-bang refactor. Threaded through is a tiering principle for routing
generative work between Opus, smaller hosted models, local 70B on Thunder, and
deterministic Go.

---

## Why this exists

Today's symptom — the gamesdesign.co.uk re-run on 2026-05-03 — produced 24 page
rows where there should have been ~14. Adoption wrote 11 pages with names like
`tool-ttk-calculator`, `guides-index`. The planner LLM later emitted page names
like `ttk-calculator`, `guides`, with different URLs (`/ttk-calculator.html` vs
`/tools/ttk-calculator/index.html`). Both surfaces called `upsertPage`, but the
`ON CONFLICT (site_id, name)` predicate never fired because the names didn't
match. Two parallel sets of work items (`adoption_page_*` and `needs_page:*`)
went into the queue, with the dispatcher unable to recognise the duplication.

The deeper cause is structural: two surfaces (adoption and the site-planner)
both write `pages` rows and queue work items, and they don't share a common
identity space, don't read each other's output, and neither one knows about
the existing realised state of the site. The current planner acts as both
*blueprint* and *queue producer*, which means re-plans always create
divergence rather than converging on a target.

The shape that fixes this is the same pattern Kubernetes uses: a declarative
artefact describing desired state, plus a reconciler that walks
desired-vs-realised and emits work for the diff. The planner stops emitting
work items. Reconciliation becomes a deterministic Go step that can't produce
duplicates by construction.

This doc is the staged plan. Phase 0 lands today and unblocks the cascade.
Phase 1 is the architectural shift. Phase 2 is the payoff — discoverers
reading the plan to judge fitness.

---

## Bugs surfaced by the 2026-05-03 gamesdesign.co.uk run

Recorded here because they aren't all in earlier handoffs and several are
addressed by Phases 0 and 1 of this plan. The run was stopped by the human
~3 hours in; rendered HTML examined.

### B-029-1 — Duplicate nav items reach the rendered page

The header `<ul>` listed `Games` twice, with hrefs `/games.html` and
`/games/index.html`. The footer's "Quick Links" had the same duplication;
"Our Services" listed `Tools / Guides / Games / Games / Guides / Tools`,
mixing both URL schemes. Same root cause as Phase 0's canonicalisation
work — two `pages` rows for each section, both picked up by the nav
builder. **Phase 0 fixes the source.** A separate guard in the nav builder
(dedup on canonical name before emitting `<li>`) is worth adding so that
even with upstream bugs, the symptom can't reach deployed HTML.

### B-029-2 — Component templates have empty/default slots that no brief fills

Examples: `<img src="">` on article cards. `<a href="/services.html"
class="btn">Open the Tools</a>` as a guides hero CTA — `/services.html`
doesn't exist on this site; it's a leaked default from a generic component
template. Empty footer brand description, empty contact email, empty
contact phone, empty meta descriptions across all pages.

This is the structural problem Phase 1's per-page brief exists to solve.
A component template is a set of named slots; the page brief enumerates
what goes in each slot. With no brief, slots fall back to component-author
defaults that bleed in from whatever vertical the component was first
designed for (in this case, what looks like a generic-services template
with a `/services.html` CTA).

**Promotes from "Phase 1 implication" to "Phase 1 acceptance test":** a
Phase 1 build of this site must produce no empty slots and no leaked
template defaults. If a slot has no brief instruction, the component
should either (a) not render that slot, or (b) emit a build error so the
problem surfaces before deploy.

### B-029-3 — Theme CSS variables never written to deployed pages

Every rendered page's `<head>` contains a placeholder block:

```html
<style>
    /* Theme-specific styles injected here:  */

</style>
```

…with the comment intact and no CSS variables defined inside. The base
stylesheet references `var(--color-text)`, `var(--color-background)`,
`var(--color-primary)` etc. which are then undefined, so the body falls
back to user-agent defaults (white background, black text). The
component-scoped `<style>` blocks (header gradient, footer cyan, hero
navy) work because they hardcode their own colour values rather than
reading vars.

This is a webdesign-agent / deploy_css bug — the resolved palette IS
written to `palettes` and `style_collections` (Apr 23 handoff confirmed
the IDs land in `sites`) but isn't reaching the per-page rendered head.
Tracked as an existing bug in 010 / the design pipeline; recording here
because this run is fresh evidence the deploy step is silently dropping
vars.

Action: not in scope for this doc, but worth a sibling investigation.
Suggested first probe: `kubectl logs` on the webdesign-agent pod that ran
`generate_css → deploy_css` and check whether the var-injection block is
populated when written.

### B-029-4 — Design intent diverges from adopted source

Adoption fingerprinted gamedesign.uk as `#121212` background, `#00bcd4`
accent, Segoe UI font stack. The planner's `design_intent` then says
`#0d1117`, `#00d4ff`, monospace headings, "professional-dark" — different
hex values, different font family, different style.

This is the planner overwriting adoption-derived facts with LLM-invented
"intent". Phase 1's reconciler architecture handles this structurally:
adoption writes `site_plan_design_direction` from the fingerprint; the
plan-builder may write *aspirational* design direction but doesn't
overwrite the adoption-derived one without explicit human direction.
Recorded so we don't lose track when designing the design partial's
ownership rules.

---

## Phase 0 — Canonicalisation and key alignment

Smallest change that stops the bleeding. Lands first; doesn't preclude any
later phase.

### What's broken precisely

| Surface | page.name produced | item_key produced |
|---|---|---|
| Adoption (`apply_adoption_plan_action.go`) | `tool-ttk-calculator`, `guides-index` | `adoption_page_<name>_<site_id>` |
| Planner (`sync_pages_to_db_action.go` writing planner LLM output) | `ttk-calculator`, `guides` | (downstream) `needs_page:<name>` from `WriteBuildItemsAction` |

Both surfaces use `upsertPage`'s `ON CONFLICT (site_id, name) DO UPDATE` —
the conflict resolution is correct. The names just diverge.

### Fix

A single canonicalisation helper in `platform/orchestration/datahelpers/`,
called from both surfaces. Input is a logical page descriptor; output is a
canonical `(name, url, page_type)` triple.

```go
// in datahelpers/page_canonical.go (new)
type PageDescriptor struct {
    Role  string // "tool" | "guide" | "game" | "section-index" | "content" | "blog-post" | "landing" | ...
    Slug  string // strategy- or adoption-derived stem
    Section string // for section indexes: "tools", "guides"
}

func CanonicalisePage(d PageDescriptor) (name, url, pageType string)
```

Naming rules (one place; both surfaces agree):

| Role | name | url |
|---|---|---|
| `section-index` (section="tools") | `tools-index` | `/tools/index.html` |
| `tool` (slug="ttk-calculator") | `tool-ttk-calculator` | `/tools/ttk-calculator/index.html` |
| `guide` (slug="rng-design") | `guide-rng-design` | `/guides/rng-design/index.html` |
| `game` (slug="jelly-invaders") | `game-jelly-invaders` | `/games/jelly-invaders/index.html` |
| `content` (slug="about") | `about` | `/about.html` |
| `landing` (slug=home or slug=index) | `index` | `/index.html` |

Implementation order:

1. Create `datahelpers/page_canonical.go` with `CanonicalisePage` and a small
   table-driven test.
2. **Adoption side** (`apply_adoption_plan_action.go`):
   - Replace ad-hoc URL/name construction (lines 438–460) with a call to
     `CanonicalisePage`.
   - Change item_key from
     `fmt.Sprintf("adoption_page_%s_%s", page.Name, siteID)` (line 653)
     to `fmt.Sprintf("needs_page:%s", page.Name)`. After this, the dedup
     index `idx_swi_dedup` catches collisions natively.
3. **Planner side** (`SyncPagesToDBAction` in `site_db_actions.go` and
   `extractPagesFromPagePlanMap`):
   - After extracting the LLM's pages, run a Go normalisation pass that maps
     each LLM-emitted page through `CanonicalisePage` using the LLM's
     `page_type` field as the role. Drop or merge any planner page whose
     canonical name already exists in `pages` (queried first).
4. Clear the test site fully (DB + git + library rows), re-adopt
   gamesdesign.co.uk, and verify only one set of page rows and one set of
   `needs_page:*` work items exists.

### What this doesn't fix

- Planner can still emit pages that are conceptually identical to adopted
  pages but with a *different role* (e.g. planner emits `content/tools` while
  adoption emitted `section-index/tools`). Ambiguity at the role level isn't
  caught by name canonicalisation alone. Phase 1 is where the planner stops
  being free to invent role assignments.
- Validator false-positives on adopted content (Bug 7 from 2026-04-23 handoff)
  still bites. Separate fix.
- Planner inventing hero images that ignore `site_archetype.design.imagery`
  (Bug 4) still bites. Phase 1 territory.

### Cost / risk

- Code change: ~200 lines, mostly the helper.
- No prompt changes. No schema changes.
- Risk: if `page_type` from LLM doesn't map cleanly to canonical roles, some
  pages get pushed to `content` as fallback. Acceptable; surface in logs.

---

## Phase 1 — Plan as declarative artefact, reconciler emits work

The structural shift. Land after Phase 0 has been live long enough to
verify dedup catches all overlaps in practice.

### The shape

Three artefacts replace today's "planner emits pages and work items":

1. **`site_specs/strategy`** and **`site_specs/briefing`** — already exist,
   already produced by their respective agents. Aspirational direction.
   Slow-changing. Per-aspect rows in `site_specs`. No change to these.

2. **`site_specs/site_plan_*`** — new aspects describing the full desired
   site. Partitioned (see below). Declarative. Created by a renamed
   `plan-builder` agent (today's site-planner, with work-item emission
   removed). Versioned via the existing `is_current` / `superseded_at`
   pattern that `site_specs` already supports.

3. **`pages` rows + `site_work_items`** — realised state. Created by the
   reconciler (Go action), not by an agent.

### Plan partitioning

A monolithic plan blob is awkward for partial updates. Use multiple `site_specs`
aspects, leveraging the existing per-(site_id, aspect) versioning:

| aspect | content | regeneration trigger | LLM tier |
|---|---|---|---|
| `site_plan_structure` | Page list, navigation order, section_type per page, page_type per page, in_header / in_footer flags. No copy. | Strategy or briefing change, or initial plan build. | Sonnet / local-70B once stable. Mostly templated from strategy + classification. |
| `site_plan_design_direction` | Style direction across the site, palette mood, typography mood, imagery rules. | Strategy change, design HITL. | Sonnet / local-70B. |
| `site_plan_content_strategy` | Cross-page voice rules, avoid-phrases, tone, social-proof handling. | Strategy or briefing change. | Sonnet / local-70B. |
| `site_plan_page_<name>` | Per-page: section list, per-section brief (3–8 bullet points of what the section contains, what it claims, what data it references). | **Lazy** — generated only when reconciler decides to promote that page to build. | Sonnet / local-70B for slot-fill against a page_type template; Opus only if the page is high-stakes (homepage, key landing). |

The lazy generation of per-page partials is the cost discipline. A 30-page
plan doesn't pay 30 LLM calls upfront. The structure partial is fully eager
(cheap, mostly algorithmic). Per-page partials accrue as pages enter the
build queue. If a page is never promoted, its brief is never generated.

### Reconciler

New action: `reconcile_site_plan` (Go, no LLM).

Inputs:
- Current `site_plan_structure` aspect.
- All current `pages` rows for the site.
- All open `site_work_items` for the site.
- A `cycle_budget` config (max new work items per run; default e.g. 8 for
  routine reconciliation, higher for first-build).
- A `preference_weights` config (adopted-and-in-plan vs adopted-but-deprecated
  vs aspirational-only, with weights derived from site age and direction
  pinning).

Outputs:
- New `site_work_items` for divergences only.
- A `reconcile_report` jsonb summarising what was emitted, what was held back
  by budget, what was held back by missing dependencies.

Algorithm (sketch):

```
desired_pages = canonical names from site_plan_structure
realised_pages = canonical names from pages where build_status != 'archived'
in_flight = canonical names from site_work_items where item_key like 'needs_page:%' and status not terminal

# Categories
to_create   = desired_pages - realised_pages - in_flight
to_rebuild  = realised_pages where pages.built_from_plan_version < current_plan_version
to_archive  = realised_pages - desired_pages   # plan dropped them
already_ok  = realised_pages ∩ desired_pages where build_status='deployed' and version current

# Apply preference weights to score to_create + to_rebuild
# Apply cycle_budget to cap output
# For each scored item, ensure deps are satisfied (composition before pages, etc)
# Emit work items, idempotent via item_key='needs_page:<name>'
```

By construction, the reconciler can't double-emit: every emission is gated by
"no open item with this item_key exists". Duplicates become a thing the
schema enforces, not a thing we hope the agents respect.

### Page version tracking

Add to `pages`:

```sql
ALTER TABLE pages ADD COLUMN built_from_plan_version uuid;
-- references the site_specs row id of the site_plan_page_<name> aspect
-- that was current when the page was built
```

Drift becomes a join: `pages` whose `built_from_plan_version` doesn't equal
the current `id` of `site_specs` for the matching `site_plan_page_<name>`
aspect → drift-eligible.

### Per-page brief generation (lazy)

When the reconciler emits `needs_page:<name>` and a `site_plan_page_<name>`
spec doesn't yet exist (or is stale), the page-build-handler workflow runs
a `build_page_brief` step *before* `write_content`. That step:

1. Loads `site_plan_structure` (which page_type, which section_types).
2. Loads `site_plan_design_direction` and `site_plan_content_strategy`.
3. Loads `site_plan_page_<name>` if current; otherwise generates it via a
   slot-fill LLM call against a `page_type` template.
4. Writes/updates `site_specs` row with aspect `site_plan_page_<name>`.
5. Records the spec id on the in-flight page-build context as
   `built_from_plan_version`.

Templates per page_type live in DB (a new `page_brief_templates` table) or
in Go const. Slots are `{ slot_name, prompt_hint, max_tokens }`. The LLM
fills slots; Go assembles the full brief.

### LLM tier per call site

A cross-cutting concern. Every action that calls an LLM declares its tier.
The chassis routes:

| Tier | Routes to (today) | Routes to (post-Thunder) | Examples |
|---|---|---|---|
| `large` | Opus | Opus (kept; rare) | Strategy spec. First plan build for high-stakes site. Briefing. |
| `medium` | Sonnet | Local-70B | Plan partials. Audit judgments. Cluster framing. |
| `small` | Haiku | Local-70B (same; small batched) | Slot-fills. Per-product variations. Single-section content. |
| `none` | n/a | n/a | Reconciler. Canonicalisation. Validation. Nav ordering. |

The action-level annotation looks like:

```go
// in the action's StepConfig:
"llm_tier": "medium",  // or "small", "large"
```

The chassis maps `llm_tier` → endpoint per a config that can flip without
code changes. Once Thunder is healthy and a 70B path is stable, flip
`medium` from Sonnet to local. No action code touched.

### Affiliate / product listings — same pattern, applied at scale

The cost question for product listings (10k+ entries) needs the same
structure but at higher volume:

1. Facts come from feeds / APIs (Go, no LLM).
2. **Cluster products by feature profile** — algorithmic. K-means or
   threshold-based grouping on product attributes. Output: ~20–50 clusters
   for 10k products.
3. **One LLM call per cluster** for editorial framing — Sonnet / local-70B.
4. **Slot-fill per product** within cluster — small LLM call (or string
   template if attribute coverage is high).

The same `llm_tier` annotation applies: cluster framing is `medium`,
slot-fill is `small`. A 10k-product run becomes ~50 medium calls + ~10k
small calls, instead of 10k large ones.

This pattern lives outside the immediate site-plan work but should share
the tiering infrastructure from day one so it doesn't get retrofitted
later.

---

## Phase 2 — Discoverers and fitness checks read the plan

After Phase 1 has stabilised. The plan being a real artefact unlocks
sharper auditing.

### Today

Discoverers and auditors judge "is this page good?" against vague
heuristics (length, presence of certain keywords, tone classifiers).
False-positives are common. Bug 7 (validator flagging adopted content
because it mentions the source domain) is a symptom of fuzzy criteria.

### After Phase 1

Discoverers read `site_plan_page_<name>` and ask precise questions:

| Discoverer | Reads | Asks |
|---|---|---|
| Content auditor | `pages.<page>.sections` + `site_plan_page_<name>.section_briefs` | Does this section contain the bullet points the brief said it should? |
| Voice auditor | `pages.<page>` content + `site_plan_content_strategy` | Does the copy match the voice rules? Are avoid-phrases absent? |
| Nav auditor | `site_nav_items` + `site_plan_structure.navigation` | Does the live nav match the plan's nav order? |
| Drift auditor | `pages.built_from_plan_version` + current plan version | Which pages are out of date? |

Each auditor emits work items only for genuine drift. Each work item is
reconciled through the same pipeline as planner-emitted ones — uses the
same `needs_page:<name>` key, gets deduplicated by the index, gets
reconciler-budgeted on the next reconcile pass.

### What "the site improves over time" means concretely

- The plan is updated by spec changes (strategy, briefing, HITL direction).
- The reconciler picks up new desired pages and emits build items at the
  configured cycle budget.
- Auditors flag drift between plan and realised; flags become work items;
  reconciler picks them up.
- Aspirational pages that were initially deferred (because the cycle budget
  preferred adopted pages) get emitted in later cycles as the budget weight
  shifts.
- Old pages that the new plan no longer wants get archived (not deleted —
  archived, with their `pages` row staying for audit history).

The site is always converging on the current plan. The plan is always
converging on the current spec. Direction at the top, deterministic
reconciliation at the bottom, LLMs only where genuine generation is needed.

---

## Phase 0 deliverables — concrete

| File | Change | Status |
|---|---|---|
| `platform/orchestration/datahelpers/page_canonical.go` | NEW. `CanonicalisePage` + tests. | TO DO |
| `platform/orchestration/actions/apply_adoption_plan_action.go` | Use `CanonicalisePage`; change item_key shape. | TO DO |
| `platform/orchestration/actions/site_db_actions.go` | In `SyncPagesToDBAction`: normalise LLM-emitted pages through `CanonicalisePage` before upsert. | TO DO |
| Test re-run | Wipe gamesdesign.co.uk fully (DB + git + library rows), re-adopt, verify single set of pages and items. | TO DO |

## Phase 1 deliverables — sketch (to expand when Phase 0 is verified)

- New action `reconcile_site_plan` (Go).
- New site_specs aspects: `site_plan_structure`, `site_plan_design_direction`,
  `site_plan_content_strategy`, `site_plan_page_<name>`.
- Renamed agent: `site-planner` → `plan-builder`. Its workflow ends at
  writing the plan partials. No work-item emission.
- New action `build_page_brief` invoked lazily by page-build-handler.
- Schema change: `pages.built_from_plan_version uuid`.
- Schema change: `page_brief_templates` table or const-table for slot
  templates per page_type.
- LLM tier annotation infrastructure in the chassis.

## Phase 2 deliverables — outline only

- Auditor agents (content, voice, nav, drift) updated to read the plan.
- Improvement loop wired to the reconciler.
- Cycle budget tuning per site age / direction.

---

## Open questions

These are decisions to make before Phase 1, not Phase 0.

### Q1. Plan storage: site_specs aspects vs new table

`site_specs` versioning is well-tested but uses `aspect` as a free-form
text key. With per-page partials, the aspect set grows as pages grow:
`site_plan_page_index`, `site_plan_page_about`, `site_plan_page_tools-index`,
etc. The unique `(site_id, aspect)` index handles this fine, but the row
count per site grows linearly with page count. For a 50-page site that's
~55 site_specs rows from the plan alone.

Alternative: a new `site_plans` table with a structured schema and
explicit per-partial versioning. More upfront cost; cleaner queries.

**Lean:** site_specs aspects, because the existing versioning machinery and
the per-aspect history index `idx_site_specs_history` already give us what
we need. The row-count concern is theoretical at our scale.

### Q2. Plan-builder LLM tier

Today's site-planner uses a single LLM call to produce a half-formed plan.
The new `plan-builder` produces multiple partials. Should it run them as:

(a) One big LLM call that produces all eager partials (structure + design +
content strategy) and Go splits the response into rows. Cheaper per token,
risks LLM omitting partials.

(b) Three sequential LLM calls, one per partial. Each is bounded and
focused. More tokens total but each call is easier to validate and
re-run if it fails.

**Lean:** (b), per-partial calls. Better quality per partial, easier
recovery, and each can be tier-annotated separately if it turns out
structure-partial wants `medium` while design-partial wants `large`.

### Q3. When is the plan re-built vs when is it incrementally updated

- A strategy or briefing change is the loud signal: rebuild the structure
  partial. Per-page partials whose page is still in the new structure stay;
  those whose page is dropped get archived.
- A HITL direction tweak might change just one partial. Should the
  plan-builder accept a "scope" parameter telling it which partials to
  refresh? Or should it always rebuild structure + downstream partials
  that depend on it?

**Lean:** scoped rebuilds. The chassis already supports targeted action
invocation. Saves tokens and avoids unnecessary churn on stable partials.

### Q4. Reconciler cadence

Today the dispatcher fires on a 30s scheduler tick. The reconciler doesn't
need to fire that often — most of the time there's no new plan to react to.
Trigger options:

- Reconcile on every plan-builder completion (event-driven).
- Reconcile on a longer scheduled tick (e.g. 5 min) as a safety net.
- Both.

**Lean:** both. Event-driven for responsiveness, scheduled for robustness.
The scheduled run becomes the natural place to apply cycle-budget logic
("this site has had no work in 24h, advance one batch from the
aspirational queue").

---

## What NOT to do, even if tempting

- **Don't add a "fixed pages" lock flag on `pages` rows.** Pages aren't
  fixed; they're realisations of the plan. If the plan changes, they
  change. The right lock is on the plan partial (`pinned` already exists
  on `site_specs`).
- **Don't keep emitting work items from the planner "for safety" while
  introducing the reconciler.** That dual-write produces exactly the
  duplicates the reconciler exists to prevent. Cut over cleanly.
- **Don't generate per-page briefs upfront for the whole site.** Token
  cost has to scale with what gets built, not what's planned.
- **Don't hardcode model names in actions.** Tier annotations now mean we
  can flip backends without code.
- **Don't rename `site-planner` to `plan-builder` until Phase 1 ships.**
  Workflow definitions reference the type name; renaming is its own
  migration.
