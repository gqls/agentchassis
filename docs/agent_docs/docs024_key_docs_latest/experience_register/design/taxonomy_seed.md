# Taxonomy seed — experience register (P2/P3 design artifact)

Owner ruling: layered taxonomy, our own kebab-case enums, industry names as `aka`, seeded
from harvest only (tens, bottom-up). Nothing below enters the register until P3 harvest;
entries marked HARVEST-PENDING are candidates observed in the live estate, not commitments.

## Level 1 — interaction primitives (vocabulary, not entries)

| primitive | meaning | not to be confused with |
|---|---|---|
| `reveal` | show/expand content in place (read-more, accordion, tooltip) | `navigate` (page changes) |
| `navigate` | leave for another page/anchor; MUST name a destination role | `reveal` |
| `submit` | send user input to a handler (form, search box) | `step` |
| `filter` | narrow a visible set without navigation | `sort` |
| `sort` | reorder a visible set | `filter` |
| `step` | advance within a bounded sequence (carousel next, wizard stage) | `navigate` |
| `play` | start/pause a time-based behaviour (autoplay, animation, game round) | `step` |
| `dismiss` | close/hide an overlay or notice | `reveal` |

Rules: closed vocabulary — extending it is a PLAN edit, not an entry-time invention. Every
contract trigger names exactly one primitive. `navigate` triggers require a
`destination_role`; no other primitive may carry one.

## Destination-role vocabulary (parameter of level-2/3 entries)

Reuses the existing `page_type`/`site_plan_pages.role` values — canonical planner list
`index, content, landing, entity-directory, entity-page, tool, blog-index, blog-post` plus
observed `guide, section-index, game, news-index` — parameterised by what the role is *of*:
`{"role":"entity-page","of":"card.entity"}`, `{"role":"tool","of":"topic"}`. `external` is
the one addition (evidence/reference links leaving the site). No new page-type taxonomy is
invented; if a pattern needs a role no page_type expresses, that is a planner-vocabulary
conversation first.

## Level 2 — component contracts (HARVESTED 2026-07-26, all five live-verified)

From the brochure component library, extracted from the sources AND confirmed on the deployed
pages of fundamentallyai.com (one component per page) — `harvest/HARVEST_02`:

| entry | component | live page | primitives |
|---|---|---|---|
| `CC-003 arrow-and-swipe-card-carousel` | `hero-card-carousel` | /capabilities.html | step, play, navigate |
| `CC-004 hover-reveal-card-grid` | `image-hover-card-grid` | /model-fine-tuning.html | reveal, navigate |
| `CC-005 scroll-snap-card-track` | `swipeable-insight-carousel` | /multi-agent-review-council.html | step, navigate |
| `CC-006 count-up-stat-band` | `stat-band` | /index.html | play (viewport-triggered) |
| `CC-007 illustrated-statement-block` | `people-feature-block` | /about.html | navigate (at most one) |

> **CORRECTED 2026-07-26 — `stat-band` was listed here as the example of a component with NO
> interaction.** It ships a count-up animation triggered by scrolling into view, with an
> honesty rule worth registering on its own: it never invents digits, it counts only to a value
> already rendered from authored data, and the true figure stays on the accessible name so
> assistive technology never hears the intermediate fakes. The "inert by design" slot passes to
> `CC-007 people-feature-block`, which genuinely has nothing — and that contract is still worth
> an entry, because "correct and inert" is otherwise indistinguishable from "unfinished".

**Also corrected: `play` is not always a control.** The vocabulary below defines `play` as
"start/pause a time-based behaviour", which fits the carousel's pause button but not the
count-up, which has no control at all — its actor is the viewport. Triggers therefore carry an
**actor** (visitor / viewport / system preference / focus movement / page lifecycle), and the
non-visitor ones live in `automatic_triggers`, not `contract` (HARVEST_02 §3.1).

## Level 3 — micro-journey candidates (HARVEST-PENDING)

> **CORRECTED 2026-07-26 — pattern #1's name and shape were both wrong.** It was written here
> as `teaser-detail-related`, with a third leg of "related links and tools" out of the detail.
> The live implementation has **no such leg**: the detail region holds the case and a close
> control, and the onward links live in the feed's `today`/`arena` CTAs — harvested separately
> as `CC-002 feed-promised-cta`. Renamed `teaser-detail-deeplink`, and the deep-linkable URL
> (the part that was NOT anticipated here) turns out to be its defining property. This is the
> harvest-only rule paying for itself on its first use: authored top-down, the register's
> first entry would have carried a leg nobody has built. Also note the primitive: the detail
> opens **in place** (`reveal`), not by navigation — with an address change but no page load,
> a case the level-1 vocabulary had no way to express (HARVEST_01 §3.8, new `self-state`
> destination role).

| candidate name | aka | shape | source to harvest from |
|---|---|---|---|
| `teaser-detail-deeplink` **(HARVESTED 2026-07-26)** | master-detail, drill-down, progressive disclosure | teaser row → `reveal` full case in place + address gains `?<param>=<key>` → `dismiss` restores both | **pattern #1, done**: vonc provocations archive, live. Entry: `harvest/entries/MJ-001_teaser-detail-deeplink.json`; its component half is `CC-001 feed-driven-teaser-list` |
| `hub-spoke-index` | hub-and-spoke | section-index page → navigate child pages → navigate back/sideways | live `parent_section` hubs |
| `search-results-detail` | search / faceted find | submit query → filter/sort results → navigate detail | news/feed sites (relojistas-class) |
| `stepped-funnel` | wizard | step through stages → submit → navigate paid/deliverable page | idea.uk report funnel (NOTE: that site is another session's workstream — harvest = read-only observation of the live pattern, coordinate before deeper extraction) |
| `timed-remote-challenge-loop` **(HARVESTED 2026-07-26)** | challenge–response loop, turn-based tool session | `submit` → real remote response → `reveal` generated reply; clock and progress markers advance ONLY on a 200 | **done**: vonc gauntlet against tools.apis.uk, live. Entry: `harvest/entries/MJ-002_timed-remote-challenge-loop.json`. Was listed here as `tool-participation-loop` with a `play` primitive — the live loop has no `play`; the clock is a consequence of a response, not a control the visitor operates |

## Level 4 — site journeys/funnels

NOT register entries. The experience loop's per-site EXPERIENCE_PLAN unit, composed from
level-3 entries once the register exists. `funnel_stage` column vocabulary
(`awareness|consideration|conversion`) adopted from the superseded `site_flows`/`flow_pages`
schema (PLAN §4).
