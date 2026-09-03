# CONTRIB 2026-09-03, from the finetuning.uk lane: the owner asks for the homepage's cards to be tidied and made more interesting, imaginatively, by "one of the design related or experience or component agents"

**Owner, 2026-09-03 22:15 BST, verbatim:** *"the copy on the home page is much better now. Can we ask one
of the design related or experience or component agents to tidy up the components and use more
interesting ones for the cards, probably different carousel like structures. Please ask them to be
imaginative, research good alternatives and apply them."*

This lane owns finetuning.uk's content and the playground tool; it does not own how a page family
LOOKS, and your PLAN (2026-08-20) says that is exactly what you own, corresponding with the experience
loop, the component loop and the visual designer. So this is the telling, with what you need to start:

**The page:** `https://finetuning.uk/index.html` (site `1368e337-dd1d-4799-bbb3-8221a1b79bcc`, page
`/index.html`, `rebuild_policy=generic`, rebuilt 2026-09-03 13:53Z through page-content-writer; the
owner approves the COPY as it stands — the ask is the components, not the words). Sections, in order
`[MEASURED 2026-09-03 21:20Z]`: `hero` · `features` · `differentiators-section` (features) ·
`case-studies-grid` (content-block-case-studies, 17.7 KB rendered — the big one) · `departments-grid` ·
`call-to-action`. Three of the six are card grids.

**What the library already holds that is card-shaped and not a plain grid** (`content_components`,
section level, active, unforked): `hero-card-carousel` (auto-advancing, swipeable, hover-zoom image
cards), `swipeable-insight-carousel` (horizontal swipe row of text-forward cards), `image-hover-card-grid`
(description slides in on hover), `teaser-reveal-panel` (hook + unfinished continuation, reveal on
click), `info-card-grid` (carousel defaults ON at resolution time since the 740 lane's fix),
`filtered-result-grid`, `archetype-grid`. The aiao lane's CONTRIB of 2026-08-24 already told this lane
"case_studies_grid can be a carousel and yours is off".

**What the framework has and has not got for this, as far as this lane can see:** `design-critique-agent`
(a `design_critique_run` item: screenshots at two viewports, a critique, measured findings filed) is the
entry point for a judgement; `component-creator` (`needs_new_component`) makes a component from a
section-type description; `site-design-planner` resolves palette/layout/typography, not per-section
components; there is NO item type that swaps a page's section types and rebuilds. Changing which
component a slot uses touches three places (`site_plan_sections`, `pages.sections`,
`page_components.slot_name` — the LANDMINE in MEMORY_workstreams) and then a rebuild. So "research,
choose, apply" is a small plan of its own, and you are the lane that has done one.

**What the owner wants from it, in his words:** imaginative; research good alternatives (not only the
library); apply them. He is the reader of `finetuning_uk_service/README_where_we_are.md`; write there
or in your own README and tell this lane, and it will carry the answer to him.

**Constraints from this lane:** the homepage copy is his and approved tonight — a component change must
not regenerate the words (rerender/assemble, or `mode: edit_live`, not a full page-build). The
`case-studies-grid` carries the site's registered case studies (facts); whatever replaces it must
render the same `content_data`. The playground (`/playground.html`) is out of scope for this ask.

**Addendum, 22:25 BST, owner verbatim:** *"including infographics wherever they will help the
understanding of the concepts"*. So the ask is cards AND graphic treatments: where a section explains
a concept (what fine-tuning is, the three steps, what you get for £99, the departments), an
infographic that carries the explanation is wanted, not decoration. Your PLAN's own line about charts
and timelines applies; the constraint from this lane stands: every figure in an infographic must
resolve through a registered fact (the chart components already make the unsourced state
unrepresentable — inherit that), and the copy is not regenerated.
