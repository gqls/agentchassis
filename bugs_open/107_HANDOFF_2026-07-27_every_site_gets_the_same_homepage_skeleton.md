# 107 — every site gets the same homepage skeleton, whatever it is for

**Filed** 2026-07-27 from the oufe.com workstream, after the owner said the new
site looks like "the standard looking site that it has produced before".
**Severity** medium-high — nothing is broken, and that is the problem. Every site
builds successfully and they all look like each other.
**Status** OPEN.

## The measurement

The palette is not the issue. oufe.com has its own style collection
(`collection-oufe-com`), as do vonc, robot-hands, idea.uk, vetcomparison and
dartsonline. The sameness lives one level up, in which sections a page is built
from:

| site | homepage composition |
|---|---|
| ai-agent-orchestration.com | hero › system-stats › features › differentiators-section › case-studies-grid › departments-grid › latest-news › call-to-action |
| finetuning.uk | hero › features › differentiators › case-studies-grid › departments-grid › call-to-action |
| fundamentallyai.com | hero › stat-band › evidence-chart › differentiators › features › info-card-grid › portfolio-showcase › call-to-action |
| robot-hands.com | hero › features › brief-explanation › tool-list › latest-news › call-to-action |
| **oufe.com** | hero › brief-explanation › info-card-grid › call-to-action |

Five sites, five different subjects, one shape: **hero first, call-to-action
last, and a run of interchangeable card-grid furniture in between.** A gripper
manufacturer, an AI consultancy, a fine-tuning service and a restructuring
publication all arrive at the same page.

oufe is the thinnest instance because its roadmap brief deliberately constrained
the page list, so it got the skeleton with fewer panels rather than a different
skeleton.

## Why this happens

The planner picks section names from a menu of available components and is asked
for a plan. Nothing in that loop represents *what kind of publication this is* as
a constraint on shape. The brief influences copy and palette. It does not
influence structure, so the structure defaults to the commonest arrangement in
the component library, which is a marketing brochure.

That default is right for a brochure site and wrong for a reference publication,
a directory, a tool site or a case library. oufe wants a reading order (mechanism
first, cases second, tools alongside) and got a conversion funnel.

Two smaller symptoms of the same cause, both fixed by hand on oufe this week:

- a **"Get Started"** header button and an **"Our Services"** footer group on a
  site that sells nothing, because the chrome template supplies them by default;
- six homepage cards linking to pages that were never built, because a card grid
  wants six cards (`bugs_open/097`).

## Why the existing machinery does not catch it

- The build gate checks claims, meta-commentary and placeholders. It has no
  opinion on structure.
- `check_voice_tells` judges register, not layout.
- The council reviews changes that are submitted to it. A first build is not.
- `features_open/017` proposes `check_interactivity_promised` — a brief that
  promises carousels against pages that use none. That is the same axis
  (brief-versus-built) and would not catch this, because oufe's brief never named
  a shape to check against.

## Fix candidates, ordered by what closes the door

1. **Give the planner an archetype that constrains shape, not just palette.** A
   publication, a directory, a tool site and a brochure have different required
   and forbidden sections. The estate already has the vocabulary: `classification`
   carries `site_type`, and `features_open/013`'s three-tier funnel and the
   `per_site_ai` archetype×pattern grid both describe archetypes already. Wire
   one of those into the planner prompt as a structural constraint, so
   "publication" cannot emit `case-studies-grid › departments-grid`.
2. **Let the roadmap brief specify structure and honour it.** The brief is
   already authoritative for the page list ("build ONLY these pages"). Extend
   that authority to section order for named pages. Cheaper than 1, and it puts
   the decision with whoever writes the brief — but it only helps sites whose
   brief bothers to say, so new sites keep defaulting to the brochure.
3. **A sameness detector.** Compare a new site's composition against the fleet
   and flag when it matches an existing site above some threshold. Diagnostic
   only, and it tells you after the fact, but it makes the drift visible the way
   `bugs_open/106`'s coverage sensor does.

Recommend 1, with 3 as the watcher that proves it worked. 2 is worth doing anyway
because it is nearly free.

## How to verify a fix

Submit two sites with deliberately different archetypes — a reference publication
and a brochure — with briefs that do not mention layout. Their homepage
compositions must differ structurally, not just in palette. **Compare the section
lists, not screenshots**, and check that the publication has no conversion
furniture (`call-to-action`, `departments-grid`) unless its brief asked for it.

A single well-shaped site proves nothing here: the defect is that every site
converges, so the test needs two at once.

## Related

- `bugs_open/097` — card grids advertising unbuilt pages, same default-furniture
  cause.
- `features_open/017` — brief-promised interactivity versus built pages.
- `features_open/015` — the site maturity ladder, which assumes sites differ by
  how developed they are. They also need to differ by kind.
- `docs024_key_docs_latest/per_site_ai/` — the archetype × pattern grid that
  candidate 1 would draw on rather than invent.
