# 043 — generated page copy invents quantitative claims, and nothing checks them against the data

*Found 2026-07-20 by the robot-hands thread while fixing CTA pairing. **Live fabrication was being
served to the public on three pages of robot-hands.com** and has now been corrected on that site;
the platform defect is unfixed and the class is almost certainly fleet-wide.*

**Family:** the fabrication family — `020` (tool-recreation invents a dataset) and `001` (a re-plan
resurrects fabrication audited out days earlier). **Different mechanism, and that is the point.**
`020` is the tool-recreation handler having no data-dependency contract. This is ordinary
**content generation** writing specific numbers into `page_components.content_data` with no source
and no check. A site with no tools at all is still exposed.

## What was live

`robot-hands.com` publishes gripper specifications. The index holds **five** grippers
(`products`, `category='gripper'`, `site_id='00ff3af5-dad8-4770-9f70-3edc267a3c92'`). Three
separate stat blocks were serving figures invented whole:

| page | component | field | published | reality |
|---|---|---|---|---|
| `about` | `content-block-about` | Gripper Models Indexed | **1,200+** | 5 |
| `about` | `content-block-about` | Actuation Technologies | **6** | no actuation field exists in the data |
| `about` | `content-block-about` | MatchMatrix Queries Run | "Tracked & Published" | nothing tracks or publishes this |
| `gripper-detail` | `system-stats` | Gripper Models Indexed | **2,400+** | 5 |
| `gripper-detail` | `system-stats` | Manufacturers Covered | **140+** | 5 |
| `gripper-detail` | `system-stats` | Actuation Technologies | **6** | as above |
| `gripper-detail` | `system-stats` | Scored Specification Parameters | **18** | 4 are compared; 7 spec keys exist |
| `index` | `brief-explanation` | Actuation technologies benchmarked | **6** | as above |

Verified live before the fix:

```bash
curl -s https://robot-hands.com/about.html \
  | grep -oE '1,200\+|>6<|Gripper Models Indexed|Actuation Technologies'
# -> 1,200+ / >6< / both labels
```

Two pages also carried the same claim in prose — *"The catalog spans pneumatic, electric, vacuum,
magnetic, soft-robotic, and adhesive actuation technologies"* — and `gripper-detail`'s stat
descriptions asserted process the site does not perform: *"datasheet sources and last-verified
dates are displayed at model level"*, and scoring *"across 18 parameters including gripping force
(N), stroke (mm), cycle time (ms), IP rating, payload-to-weight ratio, and operating pressure
range"*. `products.specifications` holds seven keys and none of cycle time, payload-to-weight or
operating pressure.

## Three things make this worse than a bad number

**(a) The numbers are internally consistent, so they read as sourced.** 1,200+ models across 6
technologies scored on 18 parameters is a coherent story. Nothing about it looks generated. A
reviewer without the DB open has no signal at all.

**(b) The suffixes prove nobody read the output.** `gripper-detail`'s stat block carried
`stat1_suffix='%'` on "Gripper Models Indexed" and `stat2_suffix='ms'` on "Manufacturers Covered" —
unedited placeholders from a generic template. The page was rendering "2,400+%" and "140+ms" and
had been doing so live. That is a strong tell that this component was populated and deployed with
no human or machine pass over the rendered result.

**(c) A partial cleanup is worse than none.** `index/brief-explanation` had two of its three values
already blanked to an em dash (`—`) — someone recognised those had no source — while the fabricated
`6` was left in place. The block then reads as *checked*, with one surviving invention.

## Root cause

The content-generation prompt asks for a stat block and supplies page context, but **no source
binding**: there is no slot for "these are the figures you may use", and no post-generation check
that a number appearing in `content_data` traces to anything. `020` fix candidate (2) proposed
rewriting its rule 9 so it binds *data* and not just arithmetic — that rule lives in
`tool-recreation-handler`'s prompt and does not cover this path at all.

This is also why `020` fix candidate (4) — a machine-readable no-unsourced-claims site flag —
would not have caught it: that candidate is scoped to blocking *generated content deploys which
introduce records*, and these are scalar claims in a stat component, not records.

## Fix candidates

1. **Bind stat components to a source.** A `system-stats` / stat-block component should require
   each value to carry a provenance field (a query, a table, or an explicit `unverified` marker),
   and render nothing where provenance is absent. Structural, and it makes the next invention
   impossible rather than detectable.
2. **Post-generation numeric audit.** Beside the existing `check_completeness`, add a check that
   extracts numeric claims from generated `content_data` and fails the item to
   `needs_human_review` when a figure cannot be tied to a countable source. Cheap net; catches the
   next variant of this and of `020`.
3. **Extend the fake-data prohibition to every generative prompt that can emit a number**, not just
   the tool prompts — the wording `020` proposes ("never generate, synthesise, seed or hard-code
   example records") needs a scalar sibling: *never state a count, total or coverage figure you
   have not been given*.
4. **Lint the placeholder suffixes.** `%` on a model count and `ms` on a manufacturer count are
   mechanically detectable and would have flagged this block on the day it deployed.

Prefer (1)+(3). (4) is nearly free and is the only one that would have caught this specific
instance at deploy time.

## How to verify a fix

- Generate a stat block for a site whose data is thin, and assert it either renders sourced figures
  or renders nothing — never a plausible invention.
- Site-wide sweep, which is how the second and third blocks here were found:
  ```sql
  SELECT p.name, e.k, e.v
  FROM page_components pc JOIN pages p ON p.id=pc.page_id,
  LATERAL jsonb_each_text(pc.content_data) AS e(k,v)
  WHERE e.k ~ 'stat[_]?[0-9]+_?value' ORDER BY p.name, e.k;
  ```
  Run it per site; every value should be traceable. **Check the rendered page too** — the suffix
  bug is invisible in `content_data` alone.
- Grep rendered pages for the suffix tell: `curl <page> | grep -oE '[0-9,]+\+(%|ms|x)'`.

## Containment applied (robot-hands only)

`docs/agent_docs/docs024_key_docs_latest/robot_hands/SQL_2026-07-20_r4b_fabricated_about_stats.sql`
and `..._r4c_fabricated_stats_sitewide.sql`. All nine stat values now trace to a query recorded in
the file header (5 models / 5 manufacturers / 4 parameters compared / 24 published figures).
Re-render queued as `source='robot-hands-r4-cta-pairing'`.

**Not yet done on robot-hands:** the same six-technology claim survives in **42 further
`content_data` fields** (body prose, `features`, `subheadline`, FAQ `questions`, `cards`) across
that one site. Those are substantive copy, not stat fields, and rewriting them is a content
decision rather than a containment step — left for the owner. That ratio (9 stat fields fixed vs
42 prose fields carrying the same unsupported claim) is worth noting for anyone sizing the
fleet-wide version of this: **the stat blocks are the visible tip.**

## Fleet-wide exposure — UNMEASURED

I checked one site. The `system-stats` component and the generic stat-block shape are not
site-specific, so the class very probably reaches other sites, but **I have not run the sweep
fleet-wide and no figure here should be read as if I had.** The verify query above is the sweep;
running it across all sites is the first thing the fixing thread should do, before choosing
between candidates (1)–(4).

## Update 2026-07-22 — robot-hands resolved the actuation claim by EXPANDING the index (owner decision)

The `43 prose fields` note above framed the six-technology claim as a copy decision.
Put to the owner as four options (rewrite prose / expand catalogue / minimal strip /
leave), the owner chose **expand the catalogue**: keep the positioning, make it true.

Key finding that sharpened the decision — the claim was not merely *unsourced*, it was
**contradicted by the index**. All five grippers are parallel-jaw (three explicitly
`24 V DC` → electric); the specs contain **no** vacuum, magnetic, soft-robotic, adhesive
or even pneumatic attribute. So "indexes grippers across six actuation technologies:
pneumatic, electric, vacuum, magnetic, soft-robotic, adhesive" advertised **four
technologies with zero grippers**. This is the vetcomparison class (published claim vs
data), not a soft overstatement.

Fix applied (`SQL_2026-07-22_r7_expand_catalogue_actuation_types.sql`): one **real,
datasheet-sourced** gripper added per missing technology — Festo DHPS-10-A (pneumatic),
OnRobot VG10 (vacuum), OnRobot Soft Gripper SG (soft-robotic), OnRobot Gecko SP5
(adhesive), Schmalz SGM-HP 50 (magnetic). Each carries `content_data.source_url` +
`verified_date` and only figures read off the manufacturer/distributor page — **no
invented specs** (that being the whole point of the family). The index now holds **10
grippers across 6 actuation technologies**, and the stat blocks were re-pointed to
**computed subqueries over `products`** (fix candidate 1 in miniature: a stat value that
traces to a query and cannot drift). about/gripper-detail now compute 10 models / 6
manufacturers / 39 figures; the `index` page's "6 actuation technologies" — flagged above
as fabricated — is now backed (`count(distinct actuation)=6`). Re-render queued
(`robot-hands-r7-catalogue-expand`); inert until the build queue drains (bug 029 stall).

**Residual, honestly (NOT fixed by expansion, flagged to owner):**
- **MatchMatrix scope.** The tool scores only the parallel-jaw grippers on clamping
  force; it does not evaluate vacuum/magnetic/adhesive/soft grippers (different physics).
  Prose claiming the *tool* "evaluates across six technologies" is still ahead of the tool.
  Expanding the index fixed the **catalogue** claim, not the **tool-scope** claim.
- **No browsable listing.** The gripper-catalog page is static prose (0 gripper names
  rendered live 2026-07-22); the new grippers back the claim + the counts but do not
  auto-generate catalog rows or detail pages. Pre-existing rendering gap.
- **Learning-center depth.** "One guide per actuation technology" (6) vs 3 real guides.

**This does not close the platform bug.** Candidates 1–4 are still unbuilt and the
fleet-wide sweep is still unrun — expansion is a per-site containment for robot-hands,
the same status the stat-number fix had. The lesson it adds: for a data-backed site the
honest fix can be to *make the data real*, not to soften the copy — but only where real
sources exist, and only with the specs actually cited.

## Update 2026-07-22 (b) — point (b)'s ROOT CAUSE located: the suffix placeholders live in a shared component's input_schema `fallback`s, and it is fleet-wide

Point (b) above ("unedited placeholders from a generic template", `%`/`ms`) is not a
per-instance content mistake — it is **structural**, in the `system-stats` content
component (`content_components.id = fdd92ad4-521a-4602-89cf-7ee1a66c10f1`). Its
`input_schema.fields` declare the four suffixes as:

```
stat1_suffix: source=static, fallback="%"
stat2_suffix: source=static, fallback="ms"
stat3_suffix: source=static, fallback="+"
stat4_suffix: source=static, fallback="x"
```

When a page leaves a suffix empty, the render resolves it to the fallback **and persists
that back into `page_components.content_data`**. So the junk is not typed by anyone — the
schema hands it out as the default. `stat*_value/label/description` are `source=llm` with
no fallback, which is why hand-edits to *those* hold and edits to the suffix do not (proven
on robot-hands: R7's values and R7b's descriptions held; R7b's `suffix=""` reverted to `%`
within one render).

**Fleet-wide — every consumer renders the junk** (5 pages, 4 sites; none set a real unit):

| site | page | renders |
|---|---|---|
| ai-agent-orchestration.com | index | `70+%`, `1,000sms`, … |
| ai-agent-orchestration.com | case-study-kafka-… | `70%`, `8ms`, … |
| gamesdesign.co.uk | index | `36.6%`, `4ms`, … |
| vonc.com | index | `14,203%`, `61ms`, … |
| robot-hands.com | gripper-detail | `10%`, `6ms`, `4+`, `39x` |

`1,000sms` and `14,203%` show these are not intended units — they are the schema
placeholder leaking through.

**Structural fix (applied to the root + robot-hands 2026-07-22,
`SQL_2026-07-22_r7c_…`):** set the four `fallback`s to `""` — a stat has no unit unless one
is specified. Because the field is `source=static` (deterministic, not LLM) and the other
four pages carry their suffix **persisted non-empty**, this change has **zero effect on any
current live page** — it only changes the default for a genuinely-empty suffix. robot-hands
gripper-detail is then fixed by clearing its four persisted `%/ms/+/x` back to empty and
re-rendering, so the static resolve lands on the new empty fallback.

**Residual (NOT fixed here — other owners' sites):** the four rows above other than
robot-hands still carry the junk suffix **persisted** in their `content_data`; the schema
fix does not retro-clear a persisted value. Each needs its suffixes cleared (or set to the
real unit its owner intends) and a re-render. Listed here for the 043 fixing thread; not
touched from the robot-hands lane because the intended unit is the site owner's call
(gamesdesign's `36.6` may genuinely want `%`).

## Update 2026-07-24 — fleet sweep RUN; live recurrence caught; candidate 3 SHIPPED (migration 201 + evidence_base); all found instances fixed

Artefacts: `docs/agent_docs/docs024_key_docs_latest/fabricated_stats_043/` (per-site
SQL, wave-2, evidence_base seeds); migration
`docs/agent_docs/sql_for_agents/201_content_writers_never_invent_numbers.sql`.

### The sweep (finally run, per "How to verify a fix" above)

All `stat[_]?N_value` fields fleet-wide + rendered-page suffix grep + a sibling
sweep of `input_schema` static fallbacks. Classification:

| site / page | was serving | verdict → action |
|---|---|---|
| robot-hands/index (brief-explanation) | **"2,400+ Gripper Models Indexed"** — RE-INVENTED by a routine re-render 2026-07-24 10:54, four days after R4c corrected it | **LIVE RECURRENCE** → fixed (computed 10/6/39); the motivating case for candidate 3 |
| vonc/index + about (system-stats, gauntlet-cta ×2, content-block, brief-explanation) | "14,203 Takes Filed Today", "Happy Customers 14,203", "Avg. Rating: 6 Archetypes", "Setup Time: 4h 12m", "Players Scored 10K+", "updated every 15 minutes" — NO takes/activity tables exist; wrong about its own content (8 archetypes, not 9; no "Contrarian") | fabricated → replaced with register counts (8/3/2/17), computed |
| gamesdesign/index | "PRD Accuracy Gap 36.6%" (no tool implements PRD), "Economy Model Types 6" (no presets exist), pity description naming parameters the tuner doesn't have | fabricated → replaced with traced figures (11 tools / 4 real tuner inputs / 10,000 trials **kept — traces to shipped JS** / 10 articles) |
| ai-agent-orch/case-study + index + about | "70 Deployed Agents / 8 Departments / 30 Types / 1000 Concurrent"; about: "Satisfaction Rate: **30+**", "Awards Won: **30 yrs**" (template mad-libs) | fabricated (and UNDERSTATED reality) → grounded in the platform's own DB: 170 agents / 13 sites / 17 services / 1,267 work items |
| leopardessconsulting/about | "Agent Definitions 150+" | **VERIFIED TRUE** (registry holds 170) — the audited-claims discipline works |
| gamesdesign/about-index, idea.uk | "Free/100%/Growing"; "£29/8 tools/5 working days" | qualitative or owner-set product facts — left |
| vonc archetype pages | "Longest/Widest/Sharpest…" | superlative brand voice, no figures — not this class |
| **finetuning.uk/about** | "Clients Served 11+ / Satisfaction Rate 100% / Awards Won 0" | **fabricated, NOT FIXED** — no honest replacement derivable from the DB; needs its owner's real story. **OPEN residual.** |
| sibling components | `system-stats-leopardess` had the same junk suffix fallbacks (0 consumers) → cleared; `gauntlet-interface` still holds 12,847/94,210/7 persisted + as schema fallbacks but the template no longer references them — INERT residue, gauntlet_dead_cta thread's territory | |

### The recurrence, mechanically (why rule 14 lost)

The 10:54 regression came through `needs_page` → page-build-handler →
page-content-writer. The writer's prompt **already had** rule 14 ("NEVER invent
specific statistics"). But the "What To Write" block lists each llm field with
its schema description — and a stat field's says *"stat_1_value (required): The
numeric value… e.g. '99.99', '2.4M', '150'"*. A REQUIRED field demanding a
number, example shapes to copy, no data anywhere in the prompt: the model
resolves the conflict by inventing ("2,400+" is literally the '2.4M' example
shape). **A prohibition without a legal alternative loses to a structural
demand.**

### Candidate 3 SHIPPED (config-only, LIVE 2026-07-24)

- **Migration 201**: rule 14 rewritten to name the conflict and the alternative —
  a figure not given in the prompt (Verified Facts / Research / Admin Brief /
  Existing Content) must not be stated; for numeric stat fields, required-ness
  is NOT permission; return an **empty string**. Same rule added to
  content-writer's Guidelines. Snapshots taken; anchors verified; ledger-recorded.
- **evidence_base seeded for the four fixed sites** (robot-hands, vonc,
  gamesdesign, ai-agent-orch): the writer_block lists each site's true countables
  with meanings + explicit NEVER-STATE lists — so 201 has figures to *allow* and
  a full-writer re-render keeps the corrected stats instead of blanking them.
  (vonc's row is a MERGE preserving migration 166's `banned_claims` checker
  patterns — see the correction note in the SQL file.)

### Wave-2c (same day): the prose-beyond-stat-fields tail, on vonc

Chasing the last grep hit ("4h 12m") after the stat blocks were clean surfaced
the class 043 predicted: fabrication in ordinary prose fields. Fixed
(`SQL_2026-07-24_wave2c_…`): fabricated countdowns ("Gauntlet closes in 4h 12m"
/ "3h 44m" — no clock exists), liveness theatre ("watch the split happen in
real time. The clock is live. The takes are stacking."; "Your Archetype updated
in real time" — no server, no persistence), and — the deepest cut — the
archetypes page's `archetype-combinations` component built entirely on **six
invented archetypes** (Contrarian, Analyst, Idealist, Provocateur, Realist,
Sage) that are not the site's documented eight; rewritten to real pairs.

**Recorded, deliberately NOT touched (experience-loop / vonc-spark thread's
call):** the present-tense product-VISION copy — the arena guide article
("Every day, a new Provocation drops… watch the distribution shift in real
time") and the conceptual differentiators ("The Gauntlet Has a Clock", "The
World reads your pattern"). Their migration-166 banned_claims deliberately
routes such copy to review rather than banning the concept; whether Spark's
vision may be described in the present tense is a positioning decision in their
lane, not a number to correct in this one.

### Still open on this bug

- **Candidate 1** (provenance-bound stat fields in component schemas — the
  `stat_1_value` llm_guidance with its "2.4M" example shapes is the field-level
  seed of the invention; binding those fields to `query.*`/evidence sources kills
  the class structurally). **Candidate 2** (post-generation numeric audit →
  needs_human_review). Both Go/schema work, unbuilt.
- finetuning.uk/about (owner story needed).
- vonc present-tense vision copy (above) — experience-loop thread's positioning
  call.
- Sites beyond this sweep's stat-field shape: prose numbers not in stat fields
  are NOT covered by the stat-value sweep (robot-hands' 42-field experience says
  the stat blocks are the tip, and wave-2c proved it on vonc); 201 now guards
  the writer for all of them going forward, but existing prose was not audited
  fleet-wide.

## Related

- `/bugs_open/020` — the tool-recreation fabrication. Same family, different path. Fix both.
- `/bugs_open/001` — re-plan resurrecting audited-out fabrication.
- `/bugs_open/023`, `/bugs_open/033` — the delivery gap. robot-hands had 20
  `cta_names_unknown_destination` items in `needs_human_review` and none of them were ever read;
  a detector for this class would rot the same way unless routing is fixed too.
- `docs/agent_docs/docs024_key_docs_latest/vetcomparison/LEGAL_2026-07-15_*` — what published
  unsourced figures cost the last time.
