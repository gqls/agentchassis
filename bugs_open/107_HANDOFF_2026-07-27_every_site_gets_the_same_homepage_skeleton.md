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

---

## 2026-08-04 — STILL VALID, OWNED BY vigilant_designer Phase 4, and the mechanism is now mapped

Contributed by the bug-sweep session that briefly claimed this bug the same
evening (`0a24f1e06`) before finding the ownership trail — recorded here so the
next reader does not repeat either the claim or the research.

**Ownership, in order of authority:**
- The **owner parked this bug 2026-07-27** — "not a blocker for now"
  (`oufe/HANDOFF_2026-07-27_continue_here.md:78-82`). No recorded un-parking.
- The **owner then approved a programme that carries the fix**:
  `docs024_key_docs_latest/vigilant_designer_offer_analysis/PLAN_2026-08-02…`
  **Phase 4.1** is candidate 1 verbatim (per-archetype required AND forbidden
  sections in the build-site-planner prompt, anti-sameness instruction,
  dormant-component steering), **Phase 4.2** is candidate 3
  (`check_composition_convergence`), and Phase 3.1's recompose handler carries
  "archetype constraints" too. That lane is ACTIVE (Phase 0 proven 08-04) and
  its plan flags 4.1 as council/owner-gated. **So: contribute here, do not
  fix this out from under that lane.**

**Re-validation (the bug is not stale):** the newest framework-built site,
lendzy.co.uk (2026-08-02), classified `hub` in `site_specs.classification`,
was built `hero > brief-explanation > info-card-grid > mechanism-flow >
call-to-action`. The census SQL lives in
`bugfix_107_homepage_skeleton/RUNBOOK_homepage_skeleton.md` §1.

**Mechanism map for the Phase-4 implementer** (full detail with citations in
`bugfix_107_homepage_skeleton/NOTES_homepage_skeleton.md`; verify against the
tree at implementation time — this is a snapshot):

1. The composition is born in ONE place: `build-site-planner`'s `plan_site`
   LLM step (prompt in `agent_definitions.default_config`; in-repo mirror
   `docs/agent_docs/sql_for_agents/053_build_site_planner.sql:1978-2228`).
   `plan_sections_action.go` only triages a list it is handed.
2. `load_components` offers the WHOLE library — no `suitable_site_types`
   predicate — and the prompt's one-shot example is itself the brochure
   skeleton (`"sections": ["hero","features","testimonials","call-to-action"]`).
   Rule 11 (news_feed → latest-news) is the only site-data-conditional
   section rule; nothing forbids a section or constrains order.
3. The LLM already emits `site_type` in its plan JSON and
   `WriteSitePlanAction` discards it (`site_plans` has no column).
4. `component_selector.go:176` already scores `suitable_site_types` (+0.35)
   but NO live `plan_sections` step config passes `site_type` — the hook is
   wired and starved (065/246/309 configs pass only site_id/sections/
   page_name). `content_components.suitable_site_types` is populated on
   86/178 active components.
5. Two post-planner fallbacks re-impose the skeleton and must be constrained
   too or the prompt fix leaks: `defaultSectionsForPage`
   (`apply_gap_plan_action.go:953-978`, hardcoded `hero > … > call-to-action`,
   site_type not a parameter) and the modal same-role sibling synthesis
   (`load_page_sections_from_spec_action.go:281-370`), which persists the
   borrowed skeleton into `pages.sections`.
6. `validate_components: true` silently drops unresolvable section names
   (`v3_site_actions.go:3055-3062`) — contradicting the roadmap block's
   promise that unknown section_types reach the selector. Candidate 2 dies
   here unless this is fixed alongside.
7. Three incompatible `site_type` vocabularies are in flight (classifier,
   strategist, old site-classifier — SPEC-010), and the
   `suitable_site_types` values match none of them exactly. Phase 4.1's
   "site_type vocabulary extension" flag is the right instinct.
8. Traps recorded for this exact surface: `pages.sections` is a CACHE of
   `site_plan_sections` (LANDMINES :244); `sections=[]` on a deployed page is
   a positive statement — do not attach furniture to tool/blog pages
   (WRONG_CALLS :1303, `bugs_open/001`); built compositions are
   force-preserved on re-plan (`reconcilePlanWithRealised`; release valve
   `recompose_pages`).
