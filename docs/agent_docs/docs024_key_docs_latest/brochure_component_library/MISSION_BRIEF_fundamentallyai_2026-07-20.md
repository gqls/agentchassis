# Mission brief for `082_submit_domain_unified.sh fundamentallyai.com`

Drafted 2026-07-20 from the resolved research/decisions in this workstream's
PLAN/NOTES. This is the `--mission-file` text for the FRESH onboarding trigger —
it seeds the domain-research-classifier's site-level archetype/design_intent, not
final page copy (that's a later, per-page content-writer brief).

---

FundamentallyAI is our own AI/software consultancy brand — a live proving ground
for a real, working AI development and orchestration platform, not a
hypothetical vendor. Audience: commercially sharp technology and business
decision-makers evaluating AI vendors, who have heard the hype and are
sceptical of it. Keep promises small, deliverable, and evidenced — never
inflate.

Visual direction: a best-in-class professional-services brochure site in the
register of Bain, BCG and McKinsey — an auto-advancing hero carousel of a few
cards (each with a title, one short teaser line, and a "read more" link, no
more copy than that), images that gently enlarge on hover instead of using
video, and further down the page a variety of section types including
swipeable card carousels on mobile. Vary the design across linked pages rather
than repeating one template. All auto-advancing/carousel components must be
genuinely accessible (pausable, keyboard-safe) and must degrade gracefully —
the first card of any carousel should carry the complete message alone. Any
numbers or statistics shown must be rendered as real, code-generated charts
from true figures, never as an AI-generated picture of a chart.

House imagery style: line illustration for any people or human figures — never
photography, and never a diffusion-generated image depicting a specific
individual. Apply one consistent tint/treatment across illustrations for
visual cohesion (in the spirit of how McKinsey applies one uniform
photographic treatment across all of its imagery).

Positioning: this consultancy markets the real, verifiable capabilities of the
platform it runs on, as service lines and case studies — not generic
consultancy filler. Emphasise, in order of strength: (1) a live multi-agent
review council that independently reviews every substantial change before it
ships, with a real decision record; (2) rapid site/tool delivery — stand up a
fully designed, working site in well under a day; (3) fine-tuning models on
real usage data, evaluated honestly against alternatives; (4) real backend
engineering including production payment integrations; (5) reviving/repairing
real digital assets (e.g. an abandoned domain's dead subscriber feed, restored
to serving correctly). A genuinely differentiating trust story: this platform
verifies its own AI-generated content against evidence and has publicly caught
and corrected its own mistakes — including a past case where our own sibling
site, leopardessconsulting.co.uk, briefly published invented details before our
verification system caught it. Name that site directly as the worked example;
do not repeat the specific invented details themselves as if they were facts.

The "private in-house search via embeddings" idea should be framed as a
capability we can build for a client (the hard technical groundwork already
exists and works), not as an existing, already-delivered guarantee.

Absolute rule, platform-wide and non-negotiable: never invent a person, a
client, a case study, or a statistic. Every claim needs a real source. If we
haven't done something for a client yet, say so plainly rather than implying
otherwise.

---

**Proposed trigger command** (not yet run):
```
./082_submit_domain_unified.sh fundamentallyai.com \
  --email fundamentallyai@contactforsales.com \
  --mission-file docs/agent_docs/docs024_key_docs_latest/brochure_component_library/MISSION_BRIEF_fundamentallyai_2026-07-20.md
```
Contact email follows the existing sibling-site convention
(`idea-uk@leopardess.uk`, `leopardess@contactforsales.com`, etc.) — trivially
changeable later via the `identity` spec if the owner wants a different address.
