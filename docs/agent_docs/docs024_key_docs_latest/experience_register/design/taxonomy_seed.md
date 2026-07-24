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

## Level 2 — component-contract candidates (HARVEST-PENDING)

From the brochure component library (`behaviour.js` exists per component — the contract is
extracted, not invented):
- `hero-card-carousel` — step (swipe/arrows), navigate (card → destination role per card)
- `swipeable-insight-carousel` — step, reveal
- `image-hover-card-grid` — reveal (hover state), navigate (`link_url` per card — today
  `source: llm`, regenerated per build; the binding replaces the guess)
- `people-feature-block` — reveal / navigate (profile)
- (`stat-band` has no interaction — a contract recording "none" is still worth an entry:
  it makes "this component is inert by design" checkable, the anti-hollow rule inverted)

## Level 3 — micro-journey candidates (HARVEST-PENDING)

| candidate name | aka | shape | source to harvest from |
|---|---|---|---|
| `teaser-detail-related` | master-detail, content-scent trail | card/teaser → `reveal` summary → `navigate` entity-page → related links (navigate: content/tool) | **pattern #1**: vonc provocations journey, after the pilot is live end to end (owner ruling: wait for tools-api) |
| `hub-spoke-index` | hub-and-spoke | section-index page → navigate child pages → navigate back/sideways | live `parent_section` hubs |
| `search-results-detail` | search / faceted find | submit query → filter/sort results → navigate detail | news/feed sites (relojistas-class) |
| `stepped-funnel` | wizard | step through stages → submit → navigate paid/deliverable page | idea.uk report funnel (NOTE: that site is another session's workstream — harvest = read-only observation of the live pattern, coordinate before deeper extraction) |
| `tool-participation-loop` | — | submit input → play round → reveal verdict → navigate onward | vonc arena/gauntlet, post-pilot |

## Level 4 — site journeys/funnels

NOT register entries. The experience loop's per-site EXPERIENCE_PLAN unit, composed from
level-3 entries once the register exists. `funnel_stage` column vocabulary
(`awareness|consideration|conversion`) adopted from the superseded `site_flows`/`flow_pages`
schema (PLAN §4).
