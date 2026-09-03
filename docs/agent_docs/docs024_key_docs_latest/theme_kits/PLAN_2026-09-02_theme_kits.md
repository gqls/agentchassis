# PLAN — theme kits, Phase 1 (created late: 2026-09-03, for work done 2026-09-02)

> **This document was created a day late and says so.** CLAUDE.md requires the standing
> five at the START of a workstream. This lane ran instead through an approved plan file
> at `/home/ant/.claude/plans/please-think-hard-about-starry-locket.md`, which carries
> the full original design plus corrections **C1–C10** and remains the authoritative
> source for the parts not restated here. This file migrates the design and the
> decisions-with-reasons into the workstream directory, where the next session will
> look for them. **Nothing here is a new decision.**

## What the owner asked for

A system of **themes**: reusable, named bundles of design defaults — CSS, components,
page structure, nav, typography, copy style — that a site may *optionally* start from
and then freely diverge from, creatable from example sites as well as by hand.

## The research finding that reshaped the brief

This was not greenfield. Migration 025 (April 2026) already built a composable-theme
system: `palettes` / `layouts` / `typography_sets` composed via `css_themes` →
`style_collections`, resolved by `site-design-planner`, rendered by `webdesign-agent`.
It is live, it is what the company pitch deck calls "Composable themes", and the word
"theme" is already overloaded across the tree (`css_themes`, `theme_id`,
`needs_theme_review`, `forked_from_theme_id`).

**So Phase 1 extends that system with a higher-level named bundle. It does not replace
it.**

## Decisions, and why

- **The table is `theme_kits`, not `themes`.** Three independent research passes hit the
  same collision: a second table called `themes` would make every `theme_id` in the tree
  ambiguous. "Theme" stays the user-facing word; `theme_kits` is the internal noun.
- **Applying a kit MATERIALISES defaults into the site's own rows. It is never a live FK
  the site stays bound to.** This continues the fork idiom the platform already uses
  (component fork = row copy; CSS fork = new per-site rows), and both independent design
  passes converged on it unprompted. A site can edit any component, palette or section
  afterwards exactly as before, and **nothing checks "is this site themed" on any edit
  path.**
- **Lineage lives in `site_specs` (aspect `theme_kit_adoption`), not a
  `sites.theme_kit_id` column.** A column reads as a live binding and invites some
  future resolver to join on it at render time.
- **`theme_kits` is a registry other mechanisms CONSULT, not a new authority that
  absorbs their write paths.** This follows the "Choice B" precedent: when
  `site-design-planner` was first designed, an early draft gave it ownership of
  navigation and layout outright, and that was walked back within 24 hours and restated
  twice as policy. Nothing about applying a kit changes who owns
  `site_nav_groups`, `content_direction` or `page_components`.
  > **CORRECTED 2026-09-03, and the correction was PREDICTED by the reviewer who
  > asked for it.** This read *"`site-design-planner` stays the ONE writer of
  > `sites.style_collection_id`"*, and I put the same phrase in the council
  > submission. **There are TWO Go writers** —
  > `install_site_composition_action.go` (site-design-planner's, the one I meant)
  > and **`SelectStyleCollectionAction` in `v3_site_actions.go`**, which persists
  > the column so downstream agents can find it by DB lookup, treating a failure
  > as non-fatal. The check is one command:
  > ```bash
  > grep -rln "UPDATE sites SET[^;]*style_collection_id" platform/ internal/ --include=*.go
  > ```
  > The council's `prior_art_librarian` seat flagged it as an uncited
  > existence/precedent claim **and named why it distrusted it: this same
  > submission had already produced two false absence claims in earlier rounds.**
  > It was right on all three counts. **What survives is the narrower claim that
  > was the actual point:** applying a kit installs nothing — it queues
  > `needs_composition` and lets site-design-planner do the install. That is what
  > the Choice-B precedent governs, and it is unaffected.
- **`page_archetypes` replaces the hardcoded `defaultSectionsForPage` Go switch**, and
  keeps it as a logged last-resort fallback rather than deleting it. Three-way scoped
  (site row > theme-kit row > fleet row, CHECK-enforced) so a site can declare its own
  durable structure default **without adopting any kit** — "sites don't necessarily have
  to be created from a theme".
- **`UNIQUE NULLS NOT DISTINCT` on `page_archetypes` is load-bearing, not style.** Every
  row has a NULL in the key and Postgres NULLs are distinct by default, so under a plain
  UNIQUE the constraint could never fire and the seed's `ON CONFLICT` was dead code.
  Proven by induced fault: re-running the seed returns `INSERT 0 0`, where the old
  version would have returned `INSERT 0 1`.

## The owner's ruling that governs the semantics (2026-09-02)

> *"I think the classifier can be given the choice. I think by default it can start with
> a theme and change it if it wishes, but it must have full authority to ignore our set
> of themes if it chooses."*

Consequences, all implemented:

- **Default mode is `start`**: it WRITES the kit's palette and typography, superseding
  what is there. It first shipped defaulting to `fill_gaps`, which deferred to any
  existing `design_intent` and was therefore a **no-op on the 33 of 57 sites the
  classifier had already touched** — a theme that never started anything.
- Written values carry `reference_source: "theme_kit:<name>"` and
  `reference_is_default: true`, so a later reader can tell a kit's default from a
  decision.
- `fill_gaps` remains as an explicit conservative mode; `reapply` also replaces an
  installed composition.
- **The ONE thing no mode overwrites is `design_intent.<dim>.locked: true`** — a
  deliberate human pin that nothing sets automatically. It is an in-data key and **not
  `site_specs.pinned`**, because `pinned` is a per-ROW flag and `design_intent` is
  superseded-then-inserted on every write with nothing carrying it forward, so a pin set
  that way survives exactly until the next write of any kind. Two of the four rows ever
  pinned are already superseded.
- **The corollary matters as much: nothing freezes a themed site's values against the
  classifier or the render overlay.** RFC_059 proposed exactly that and was WITHDRAWN
  under this ruling.

## The typography asymmetry, and why it needed its own guard

The two dimensions are not symmetric, and this is a pre-existing property of the
codebase, not something this work introduced:

| dimension | rung 1 | rung 2 | so writing `design_intent`… |
|---|---|---|---|
| palette | `mission` | `design_intent` | cannot displace a human's mission hint |
| typography | `design_intent` | `mission` | **WOULD silently outrank it** |

Hence `missionPrefersTypography()` feeds `typoLocked`, so a site carrying
`mission.preferred_typography` keeps it and the kit stays a default. Palette needs no
equivalent. **This is the guard a council reviewer correctly objected they could not see
in round 1's sketch** — the claim was in the rationale and the code, and not in the
plan.

## Phasing

- **Phase 1 (BUILT, LIVE, ADOPTED BY NOTHING):** the registry, `apply_theme_kit`,
  `page_archetypes`, and the one resolver change layout needs.
- **Layout is the only dimension needing a resolver rung**, because layout resolution
  never consults `design_intent` at all, whereas palette and typography already read
  `design_intent.<dim>.reference_values` and therefore needed **zero** resolver changes.
- **Deferred by design, sketched so the shape is ready:** nav patterns
  (`theme_kits.extras.nav_pattern` → the existing `site_specs.site_config.chrome.*`
  fields) and voice/copy-style presets (`theme_kits.extras.voice_preset` →
  `content_direction`, injected at `ai_actions.go`'s `injectPlatformBlocks`). Neither
  needs schema work: `extras jsonb` already has room, so building them later is additive.
- **"Create a kit from an example site" — designed, NOT built.** Two directions landing
  through one materialisation action so "promote a site we built" and "learn from a site
  we found" are one code path. From our own site: mirror
  `fork_theme_from_site_action.go`'s transaction shape. From an external reference:
  clone the site-adoption pipeline's crawl → fingerprint → classify stages, plus one new
  LLM step `derive_page_archetypes`. **Do not resurrect the VLM/screenshot-to-code
  approach** — designed once and abandoned on cost grounds (`adoption-pipeline.md`
  ADO-028). **HITL gate for both: reuse `status='needs_human_review'` +
  `handler_agent=''` (`bugs_open/291`)**, never `fork_theme_from_site`'s phantom
  `theme-review-handler`, which both research passes independently flagged as broken.

## Corrections to the originating plan

C1–C10 live in the plan file and are not restated here. The two that changed the design
rather than the code: **C10** is the owner ruling above (it answers C7 and C8), and
**C8** concluded RFC_059 should not be ratified as drafted — it was subsequently
withdrawn.

**Corrections made after the plan file closed** are in
`NOTES_theme_kits.md` and, where they change what a reader elsewhere would believe, in
the concept-register entry DES-085 itself.

## What Phase 1 turned out to be worth — read this before planning Phase 2

Measured after shipping, and it deflates the lane. **Of the four dimensions a kit
bundles, three cannot change what a site looks like:**

1. **Palette cannot reach the stylesheet.** `render_css_from_spec` is spec-wins on all 8
   core slots and `analyze_design` reads `design_intent`, never the composed palette row.
   Measured at the artefact: gamedesign.uk resolved a deliberately hand-chosen palette
   (`palette_source=mission_hint`, the first time that rung ever fired in fleet history)
   and served **none** of its eight core colours. **This is the owner's ruling working,
   not a defect** — the lever on served colour is the BRIEF.
2. **`page_archetypes` governs at most 1 live page in 18.** 1,022 of 1,083 live pages
   (94.4%) match no exact `defaultSectionsForPage` output, and 5.6% is an UPPER bound
   because a planner can choose those lists unaided. **The structure lever is the
   planner's prompt.**
3. **Chrome is a no-op as seeded.** All four kits pin header/footer components that
   resolve to `site-header`/`site-footer`, which is exactly what `ChromeSlotFunction()`
   hardcodes for a site with no pin at all.

**Layout is the only dimension where adopting a kit changes anything** — and two of the
four seeded kits pick a layout the tag matcher would have picked anyway. So Phase 2
should not be "more kit dimensions". The honest next question is whether a kit is the
right vehicle at all, or whether the value is entirely in **layout reachability**, which
`bugs_open/445`'s fleet scorer is about to make measurable.
