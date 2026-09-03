# Who owns composition, and does it need its own loop?

**Owner question, 2026-09-03:**

> *"Do we need another loop like the experience loop that sorts out this component
> composition? we have theme kits and page composition and experience loop and visual
> designer and designer, might we extend one of their remits or another agent or
> workflow remit or is it a separate responsibility. I think whatever composition we
> decide on it should act like the components — it can have defaults and sites or
> pages can fork as they want."*

Everything below is `[MEASURED 2026-09-03]` against the live database and tree unless
marked otherwise.

---

## 1. THE FIRST ANSWER: "composition" is TWO things, at different grains

The word is doing double duty, and almost every downstream question resolves
differently depending on which is meant. Naming them apart is the single most useful
move available.

**A. WITHIN a section — parent component contains child components.**
This is `features_open/035`'s subject: a `page_components` row with
`parent_instance_id` set, rendered into its parent's `{{.slots.*}}`. It answers *"what
is inside this article — prose, then a figure, then a pull-quote?"*
**Live population: 0 of 3,035 rows.** The read path shipped in `v1.0.1355`; nothing
uses it yet.

**B. ACROSS a page — which sections, in what order.**
This already exists and is fully live: `site_plan_sections` carries
`plan_id, page_name, ordering, component_name`, plus per-section `palette_id`,
`layout_id`, `typography_set_id`, `assigned_fact_ids`, `subject`. It answers *"what is
on this page and in what sequence?"*
**Live: 54 plans across 34 sites.**

**A third thing wears the same word and is NOT either of them.** `site-design-planner`'s
remit reads *"Resolves composition (palette, layout, typography) for a site BEFORE
webdesign-agent renders"* — that is THEME composition, and
`fork_theme_composition.go` is its helper file. It composes design tokens, never
component arrangement. Anyone searching the estate for "composition" hits this first.

## 2. THE DEFAULTS-AND-FORK PRINCIPLE, measured against what exists

The owner's principle is that composition should behave like components: a default,
which a site or page may fork. **Two mechanisms already implement exactly that, and
the third does not exist.**

| concern | table | defaults | fork mechanism | live? |
|---|---|---|---|---|
| component definition | `content_components` | library = `forked_from IS NULL` | `forked_from` | **YES — 412 library, 85 forks, 88 page rows across 25 sites** |
| visual structure + default chrome | `layouts` | 18 library rows, `default_header_component_id` / `default_footer_component_id` | `forked_from_layout_id` | **NO — 18 library, 0 forked. Never once used.** |
| **arrangement (which sections, in what order)** | `site_plans` / `site_plan_sections` | **none** | **none** | n/a — `site_id` is NOT NULL, so every plan is site-specific from birth and generated from scratch |

**Three findings follow, and they are the substance of the answer.**

1. **The component fork model is real and exercised**, so "act like components" is a
   working pattern to copy rather than an aspiration. Note precisely how it works:
   the definition carries NO `site_id`. A fork is bound to a site only by a
   `page_components` row pointing at it, and every selector index carries
   `WHERE forked_from IS NULL` so forks are invisible to selection. **Defaults are
   what you select from; forks are what you reach by explicit reference.**
2. **The layout fork model is DORMANT — 0 of 18.** This estate already built one
   defaults-and-fork mechanism for structure and never drove it. That is the
   strongest single argument for caution here: 035 §6.8 says a dormant mechanism
   stays dormant without a driver, and this is a live instance of it in the very
   neighbourhood being extended.
3. **Arrangement has no defaults at all.** `site_plans.site_id` is `NOT NULL` — there
   is no library plan, no template, no fork. Every site's page arrangement is
   generated fresh. **That is the gap the owner's principle actually names**, and it
   is at grain B, not grain A.

## 3. DO WE NEED ANOTHER LOOP? — No, and the reasons are specific

**No new loop, for four reasons of descending strength.**

1. **A loop needs a driver, and this estate has the receipts on what happens without
   one.** `layouts` forking: built, never used. The experience council itself:
   80 verdicts across 13 subjects, newest **2026-08-15** — nineteen days stale at time
   of writing. A new arrangement-scoring loop would be the third mechanism in this
   neighbourhood awaiting a driver.
2. **The judgement already has a home at the right grain.** Promoting an arrangement
   IS a design decision, and `experience-approval-council` already rules on exactly
   that object class — its 13 subjects are features, experiences and design decisions
   (`D-001-free-beside-paid`, `site-chat-intake`, `tool-patent-check`). Fable reached
   this independently in the G6 draft. What it does NOT currently do is rule on
   anything editorial or on any component arrangement — but that is a matter of what
   is submitted to it, not of its remit.
3. **The parts already have a scorer.** `compute_component_quality` scores components;
   126 of 381 carry a score. An arrangement's parts are components.
   ⚠ **But see §5 — that scorer currently mis-reads composites, and it must be fixed
   before anything gates on it.**
4. **A new loop would need its own corpus to be any good.** There are zero composed
   pages. A scorer with nothing to score is a mechanism that learns nothing and
   reports confidently.

## 4. WHOSE REMIT — the recommendation, split by grain

**Grain A (within a section) → the COMPONENT LIBRARY's remit. No new owner.**

This is the recommendation because it requires *nothing new at all*. A composite is
already a `content_components` row whose `input_schema` declares `slots` (035 D3).
Therefore it already inherits, by construction:

- defaults — it is a library row, `forked_from IS NULL`, and selectable;
- forking — `forked_from`, the live mechanism, 85 forks deep;
- versioning — `component_versions`, which already snapshots every template edit;
- quality — membership in `compute_component_quality`'s sweep, which filters on
  `is_active` only.

**The owner's principle is therefore already satisfied at grain A**, and satisfied by
the existing mechanism rather than by a new one. That is worth stating plainly because
it means grain A needs a decision *not* to build anything.

**Grain B (across a page) → `site-design-planner`'s remit, EXTENDED. This is the real
proposal.**

`site-design-planner` is the right host and the argument is that it already does the
analogous job one layer down. Its own remit line is *"resolves composition … for a
site BEFORE webdesign-agent renders"*, and its helper already implements a
defaults-and-fork discipline in behaviour: layouts **matched from the library**,
typography **matched-or-new**, palettes **always site-specific** — with a HITL work
item that surfaces *"we picked X because Y"* to a reviewer.

That is precisely the shape an arrangement library needs: match a default, fork when
the site differs, and put the choice in front of a human with its reasoning.

**What it would need that it does not have:** a library of arrangements to match
against — i.e. `site_plan_sections` rows not owned by a site. Concretely, either
`site_plans.site_id` becomes nullable (a library plan) or a sibling
`plan_templates`/`arrangement_patterns` table holds them. **That is the one genuine
schema question**, and it is grain B's, not grain A's.

**Who should NOT own it, and why — stated so the alternatives are visibly considered:**

- **`experience-planner`** composes journeys, promise ledgers and data needs. Its
  output is what a visitor should be able to DO, not what a page is made of.
- **`visual-designer`** is images, logos and visual assets. Wrong medium.
- **`brand-designer`** selects and generates themes. Wrong layer — it decides how
  things look, not what is there.
- **`feature-designer`** turns an owner-approved capability spec into a feature. Its
  trigger is an owner approval; arrangement has no such trigger.
- **`design-critique-agent`** captures screenshots and critiques rendered pages. It
  judges the result; it does not choose the arrangement.
- **`page-content-writer`'s `plan_sections` step** is where arrangement is produced
  TODAY, as a side effect of content planning. Leaving it there is the do-nothing
  option and it is defensible — but it is why arrangement has no defaults: a step that
  plans one page's content has no reason to consult a library of shapes.

## 5. ⚠ ONE THING MUST BE FIXED BEFORE ANY OF IT

`extractTemplateVariables` (`compute_component_quality.go:352`) returns **deduplicated
field roots**. Probed 2026-09-02:

```
{{.slots.lead}}{{.slots.quote}}  ->  [slots]          (one name for two slots)
{{.headline}}{{.slots.lead}}     ->  [headline slots]
```

So **every composite mis-scores**, silently, as a low score rather than an error —
because its slots advertise one undeclared-looking field. Any promotion gate that
requires a quality score on a composite would gate on a number computed wrongly.
Recorded as 035 hazard 10; it is P1/P2's to fix.

## 6. WHAT THIS DOES NOT ANSWER

- **Whether a pattern is a promoted COMPONENT (grain A) or a promoted PLAN FRAGMENT
  (grain B).** Fable's G6 draft answers *component*, which is right for grain A and
  is the cheaper half. The owner's steer — *"store patterns of such combinations"* —
  reads more like grain B, because a combination of sections across a page is what a
  human would call a pattern. **These may be two features rather than one**, and G6
  §7.1 should be re-read with that possibility in view.
- Whether the experience council should start receiving editorial subjects at all. It
  has never had one.
- Anything about the routing gap: nothing reads `render_mode`
  (`HANDOFF_2026-09-03` §3), which is a separate defect and not an architecture
  question.
