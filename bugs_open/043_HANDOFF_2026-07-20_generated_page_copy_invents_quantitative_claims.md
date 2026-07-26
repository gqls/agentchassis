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

### Wave-2d (same day, hours later): the POISONED GIVEN — a fourth mechanism, observed live

Migration 201 + evidence_base did NOT stop a 15:10 full-writer re-render of
ai-agent-orchestration/index rewriting the grounded stats back to
"70+ / 8 / 30+ / 1000s". Cause: the site's own `content_direction` / `site_plan`
/ `briefing` specs contain, verbatim, *"the system-stats component showing real
numbers — 70+ agents, 8 departments, 30+ agent types, thousands of concurrent
instances"* — rendered in the prompt under **"follow this closely"**, above the
Verified Facts block. The writer was *ordered* to write those figures and told
they are real. Rule 14 v2 permits given figures; a spec is a given. **The
candidate-3 rule cannot beat a poisoned spec — the givens themselves must
trace.**

Truth audit of the poisoned spec (agent-generated, Apr–May): three of its four
claims were TRUE-but-conservative ("over 70 agents" — registry holds 171;
"8 departments" — the platform's real self-taxonomy, which wave-1's treatment
had wrongly classed as fabricated and banned; "30+ agent types" — 165). Only
"thousands of concurrent agent instances" was untrue — and its honest
neighbour exists: **1,699 orchestrations in the 24h to 2026-07-24**. Fixed
(`SQL_2026-07-24_wave2d_…`, versioned supersede+insert on all four aspects):
untrue clause → "over a thousand orchestrations a day"; hardcoded figure list →
"real registry counts, taken only from the Verified Facts list"; evidence_base
refreshed (taxonomy + volume added, over-broad departments ban narrowed to the
"departments served" misframing); index stats restored (now 171 / 8 / 14 /
1,284 — the counts moved during the session because they are live).

**Implication for candidates 1/2 and for any future 043 work: audit the SPEC
aspects (`briefing`, `identity`, `content_direction`, `site_plan`) for numeric
claims, not just content_data.** A number in a spec is an instruction, and it
outranks every writer-side rule.

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

## Update 2026-07-26 — candidate 1 SHIPPED and LIVE (migration 217); candidate 2+4 BUILT (inert); the checkers were found to be a NO-OP on the three sites that fabricated

Artefacts: `docs/agent_docs/sql_for_agents/217_stat_values_optional_and_template_gated.sql`,
`docs/agent_docs/sql_for_agents/218_evidence_facts_for_043_sites.sql`,
`platform/orchestration/datahelpers/claims_stats.go`,
`platform/orchestration/actions/validate_page_content_stats.go` (+ tests for both).

### The state this session found: migration 201's remedy had frozen the fabrications in place

`bugs_open/073` is the direct consequence of candidate 3 and it had a live cost nobody
had measured. Rule 14 v2 tells the writer to return an **empty string** when no given
figure fits. `missingRequiredLLMFields` (`json_envelope.go:204`) counts an empty string
as a **missing required field** (`isEmptyContentValue`, `:231`), so the render gate
refuses the section (`v3_site_actions.go:1719`). Measured 2026-07-26:
**`ai-agent-orchestration.com/index.html` could not be rebuilt by the writer at all, and
was therefore still serving the pre-201 case-study metrics that 201 existed to correct.**
The anti-fabrication fix had made the copy unfixable rather than truthful.

So 043 and 073 are one fix, and it is the *structural demand*, not the rule, that had to
go — exactly what wave-1's post-mortem predicted ("a prohibition without a legal
alternative loses to a structural demand").

### Migration 217 — candidate 1, config-only, LIVE

Five parts, one transaction:

1. **80 stat fields across 10 components** → `required:false` + `on_missing:skip_field`
   (`system-stats`, `case-studies-grid`, `content-block-about`, `gauntlet-cta`,
   `archetype-result-card`, `product-hero`, `bayesian-ranking-hero-tool`,
   `product-specs`, `platform-comparison`, `tool-guide-intro.read_time_value`).
2. **46 template gates.** Stat cards gate on their own value; **tables gate the `<tr>`,
   never the `<td>`** (a hidden cell under a fixed `<thead>` shifts every later column
   left); the four wrappers carrying a border or margin (`.about-stats`,
   `.gauntlet-stats`, `.arc-stats`, `.stats-grid`) also gate the container, or a
   fully-empty block renders as rules over nothing.
3. **The invention seeds removed** from those fields' `llm_guidance` — every
   `e.g. '99.99', '2.4M', '150'` exemplar (043 recorded the writer copying that exact
   `2.4M` shape to produce `2,400+`), plus `gauntlet-cta`'s "compelling" and
   `content-block-about`'s "memorable, credibility-building number", which ask the writer
   to persuade rather than report.
4. **`component-creator` gained a NUMERIC FIELDS RULE**, modelled on its existing IMAGE
   FIELDS RULE. **Without this, 217 is a one-time cleanup** — the next generated
   component ships required numeric fields with example shapes and re-seeds the class.
5. **The writer prompt names the optional case out loud** (`{{else}}, optional — return
   "" if you have nothing true to put here`) instead of leaving it as silence.

**The bug-026 truncation tripwire stays armed** and a post-condition asserts it: all ten
components keep ≥3 required `source:llm` prose fields, so a truncated response still
hard-fails the render. Only the numeric fields were relaxed, enumerated explicitly —
never pattern-matched, because `%stat%` also catches `availability_status` and
`empty_state_label`/`empty_state_message` on unrelated components.

**Safety argument**: a `required:true` field can never currently hold an empty value
*because the gate refuses to render the section*, so no deployed page depends on the
relaxed constraint. The migration's own guards caught two of my errors during dry-run — a
miscounted needle total, and an over-broad post-condition regex that flagged
`archetype_description`, a prose field the migration deliberately leaves required.

> **CORRECTED 2026-07-26, same session, ~40 minutes after applying.** I wrote above, and
> in 217's header and commit message, that this was *"verified rather than asserted —
> checked across every live placement of the ten components: **zero** empty-or-absent
> required llm fields"*. **That is false, and the query I ran did not show what I said it
> showed.** Re-running the identical query against the pre-217 schemas — which survive in
> the migration's own backup table — returns **one** row:
>
> ```sql
> SELECT bak.name, p.name, f.k, pc.content_data->>f.k
> FROM page_components pc
> JOIN bak_043_stat_components_20260726 bak ON bak.id = pc.component_id
> JOIN pages p ON p.id = pc.page_id
> JOIN LATERAL jsonb_each(bak.input_schema->'fields') f(k,v)
>   ON (f.v->>'source')='llm' AND (f.v->>'required')::bool
> WHERE (pc.content_data->>f.k IS NULL OR btrim(pc.content_data->>f.k)='');
> -- case-studies-grid | enterprise-reference-deployment | card3_stat_value | ''
> ```
>
> The same query against the live (post-217) schemas returns NONE, because the field is
> now optional — which is the only reading consistent with the evidence: **my "before"
> check actually ran against the "after" state.** I believed I had ordered it correctly
> and did not confirm that against the clock, so what I published as a verified safety leg
> was a tautology: I asked whether any *required* field was empty, of a schema in which
> those fields were no longer required.
>
> **The conclusion survives, and is stronger than the claim I made.** That page's rendered
> HTML dates from 2026-05-01 and a config change does not re-render it, so nothing served
> changed. And the direction of 217's effect on it is the opposite of a regression:
> *pre*-217 that stored empty required field meant any re-render escalated to the writer
> and any rebuild died at the gate — the page was stuck in exactly the way `073` describes,
> a second instance nobody had found. *Post*-217 it renders with card 3's stat hidden. So
> the honest sentence is not "no page had an empty required field" but "**one page did, and
> it was frozen by it; 217 unfreezes it**".
>
> **What caught it:** noticing an empty `card3_stat_value` while reading that page's stored
> content for an unrelated reason, and being able to check it because the migration had
> written a backup table. The artefact that made the change reversible is what made the
> error detectable — an argument for taking the backup even when the change looks safe.
>
> **The cheap check:** when a verification is meaningful only *before* a change, capture
> its output with a timestamp in the same command that applies the change, or run it
> against the backup afterwards. "I ran it first" is a memory, not evidence.

### Proven against the live config with 073's own recorded input

The fleet build pipeline was down at verification time (see the blocker below), so the
fix was exercised directly: the **live** `case-studies-grid` schema and template, the
deployed `missingRequiredLLMFields`, and the **exact** writer output 073 recorded from
orchestration `55be2497` (`card1_stat_value: "0"`, cards 2–5 empty):

```
RENDER GATE: stat fields reported missing = []          (was: [card2..card5_stat_value] -> page build dead)
RENDER:      empty <strong></strong> = 0  |  csg-card-stat spans emitted = 1
```

and on live `system-stats`: all four empty → gate passes, the bordered grid does not
render at all, the headline survives; two of four → exactly two cards; all four present →
four cards, no template residue (the no-op case).

### Candidate 2 + 4 BUILT — and why the existing scanner never caught any of this

**The claims engine has scanned for unregistered numbers since 2026-07-16 and never
caught one of these fabrications.** Two independent reasons, both now pinned by tests:

1. **`extractAssertions` splits a stat card in half.** `div` is in
   `assertionBlockElements` (`claims.go:147`), and a stat card puts the value and its
   label in *sibling divs*. The number's block is the bare string `"170"`, so
   `claimWindow` holds no label text, `businessClaimContextRe` cannot match, and the
   candidate is dropped before `numberSupported` is ever consulted.
   **The corpus hid this**: `TestCorpusB3`'s stat fixture uses `<span>`, which is inline,
   so its value and label land in one block and the scan appears to work.
2. **`businessClaimContextRe` is a prose gate** and does not match "Gripper Models
   Indexed", "Actuation Technologies", "Takes Filed Today" or "PRD Accuracy Gap" — about
   half of this bug's documented labels. So the new scan deliberately does **not** apply
   it: a field named `stat1_value` is a published quantitative claim by construction.

Candidate 4 is only implementable on `content_data` at all — from rendered HTML
`2,400+%` is one string, and nothing can tell whether the `%` was authored or leaked from
the schema's static `fallback`.

Shape: two new files reusing `claims.go`'s helpers with **zero edits to it**, plus 13
lines in `validate_page_content.go` (check 9, two toggles). No new action, no new
`item_type`, so no `verifier_coverage_test.go` obligation, and the
`error → mark_needs_review → needs_human_review` routing is inherited whole. Pairing
refuses rather than guesses. Severity rule: **`error` must mean "a machine checked this
and it failed", never "we could not check this"** — so no-facts and unpaired both warn.
**Go half is INERT until the next image roll.**

### The finding that mattered most: the checkers were switched off on the sites that fabricated

`ParseEvidenceBase` returns nil when a row has no `facts[]` **and** no `banned_claims[]`.
The rows seeded on 07-24 for this bug's own four sites carry **only a `writer_block`**. So
both consumers — `validate_page_content` check 8 and `check_unverified_claims`, each
gated on `eb != nil` — have been **silent no-ops on robot-hands, gamesdesign and
ai-agent-orchestration since the day they were "protected"**. The writer_block half
worked (the prompt template reads it straight from `site_specs`, not via
`ParseEvidenceBase`), which is exactly why the writer stopped inventing while the
checkers stayed blind. vonc and oufe were never affected — their migration-166-era
`banned_claims` make them parse non-nil.

Migration **218** seeds real `facts[]` for the three. Every figure was re-derived live
first, per this lane's hard rail (a fact comes from a query or an owner attestation,
never from transcribing site copy — that copy is what may be fabricated). **Almost none
still held:**

| claim | writer_block (07-24) | live (07-26) |
|---|---|---|
| active agent definitions | 170 | **175** |
| distinct agent types | 165 | **174** |
| live sites | 13 | **14** |
| work items completed | 1,267 | **1,051** ← *went DOWN* |
| orchestrations per day | 1,699 | **1,834** |
| robot-hands spec figures | 39 | **59** |

Two consequences worth carrying forward:

- **A frozen snapshot is not evidence — it is a fact with an expiry nobody wrote down.**
  Every fact in 218 carries `source.sql`, the query that *defines* its meaning, so
  `refresh_evidence_base` can keep it current.
- **"Work items completed" is not monotonic.** The ledger is reaped, so a
  cumulative-sounding achievement stat can fall. Registered with tolerance `gte` so the
  audit flags the overstatement `aao/index` publishes today rather than blessing it.

Trap recorded: `writer_block_managed` was deliberately **left off**. `composeWriterBlock`
(`refresh_evidence_base_action.go:566`) emits only NUMBERS / CAPABILITIES / NAMED
ENTITIES — it has **no NEVER-STATE section**, so turning management on would silently
delete the "NOT TRACKED, NEVER STATE" lists, which are the half of these blocks that
stops the writer inventing a whole new *category* of figure.

### Live residuals found this session, NOT fixed (both need a re-render)

- **`ai-agent-orchestration.com/enterprise-reference-deployment.html`** (HTTP 200, live)
  still serves the poisoned-spec figures wave-2d corrected on `index` only: *"70+ agents
  in concurrent production / 8 departments running isolated agent groups / 30+ agent
  types under full audit coverage"*. Component last written **2026-05-01**; the page was
  never in the 07-24 sweep.
- **`aao/index`** publishes `1,267` work items against a live `1,051`.

### Also confirmed clear

**`finetuning.uk/about` — the one named open per-site residual — is already fixed**
(2026-07-24 16:33). "Clients Served 11+ / Satisfaction Rate 100% / Awards Won 0" is now
"UK-Based / Services Offered 11 / Vendor-Neutral". No action needed.

### Spec-aspect drift, swept fleet-wide (wave-2d's parting lesson)

Hard figures still live as *instructions* in `briefing`/`identity`/`site_plan`: aao
("over 70 specialised AI agents organised into 8 departments", "30+ agent types" — both
true-but-conservative) and **leopardess `identity`: "It holds 143 agent definitions, of
which 56 are active"** — live count today is **175 active**, so that spec is stale by 32
and the writer is told to follow it. Nothing refreshes a spec;
`refresh_evidence_base` refreshes only the evidence base. Unfixed.

### Blocker on end-to-end verification (NOT this bug)

The full-writer rebuild of `aao/index` could not be run. Since ~18:02 UTC on 2026-07-26
**every** `build-pipeline-trigger` hangs at `spawn_dispatch`/`AWAITING_RESPONSES` and
never spawns a child — `bugs_open/029`'s signature, fleet-wide, affecting all sites and
all sessions. A direct kcat fire of `page-build-handler`, bypassing the dispatcher
entirely, also produced no orchestration row. Other lanes (council-gate,
endpoint-health-checker, evidence-freshness) complete normally throughout, which is what
makes it easy to misread as "my page is stuck". **The build pipeline is down, not this
fix.** Work item `54734027-a910-4d86-9cc1-336f0619fe47` is parked `triaged` for whoever
picks this up; correlation `8085c770-5011-49c4-a7e4-14035a6ba753` is the direct fire.

> **SUPERSEDED the same evening — the stall cleared and the rebuild ran.** See the update
> immediately below: correlation `81efa9cb-a501-4bbb-b27a-a12d2aa68089` at 19:06–19:10
> passed iteration 4. The direct fire above (`8085c770…`) did eventually start at 18:44 and
> then died at `spawn_content_writer` with "timed out after 3 retries" — a postgres restart
> at ~18:52, not this fix.

### Update, same evening (19:07–19:20) — the Go half went LIVE, the rebuild PASSED iteration 4, and check 9 caught a drift I had just created

Three things landed after the sections above were written. All measured.

**1. The image was rolled to `v1.0.1170` by another session, so candidates 2+4 are LIVE,
not inert.** Every line above saying "inert until the next image roll" is superseded.

**2. The end-to-end rebuild ran and got PAST iteration 4** — the verification both this
file and `073` asked for, now observed rather than argued. Correlation
`81efa9cb-a501-4bbb-b27a-a12d2aa68089`, 19:06–19:10. The writer produced all eight sections
(`generated_content_0` … `generated_content_7`), including **`generated_content_4`, the
`case-studies-grid` iteration that killed the build deterministically on 07-24 and 07-25**.
What it wrote into the five previously-fatal slots:

| slot | value | label |
|---|---|---|
| card1 | `1,834` | Orchestrations in 24 hrs |
| card2 | `14` | Live sites in production |
| card3 | *(empty)* | *(empty)* |
| card4 | `175` | Active agent definitions |
| card5 | `17` | Backend services |

Against what that component was serving before — `4 days` / `<10 min` / `8+` / `1,267` /
`100%`, invented per-case-study outcome metrics — **four of the five are now figures
registered in migration 218 hours earlier, and the fifth is honestly blank.** That is
candidates 1 and 3 working together as designed: a legal way to say "I have no figure",
and a register of true ones to reach for instead. First live exercise of 218's registers,
consumed within the hour.

**3. Check 9 fired on real traffic and was right — about a defect I had introduced that
afternoon.** The build stopped at `validate_content`, 0 blockers / 3 errors, all the same
figure:

```
unregistered_number  1,834  claims       "…Platform 1,834 Orchestrations in 24 hrs"
unregistered_stat    1834   stat_claims  system-stats.stat2_value
unregistered_stat    1,834  stat_claims  case-studies-grid.card1_stat_value
```

The two `unregistered_stat` rows are check 9 — the new content_data lane — reporting
`component.field` locations on its first live run. The `unregistered_number` is check 8,
active on this site for the first time because 218 finally gave it facts to compare against.

**Why it fired is the useful part.** 218 gave each fact a `source.sql` so
`refresh_evidence_base` keeps it current, and deliberately left `writer_block` hand-managed
(managed regeneration would delete the NEVER-STATE list). At 18:34 the refresher re-ran
`aao-orchestrations`' SQL; the rolling 24-hour window had moved and rows had been pruned,
so the **fact fell 1834 → 1790**. The prose block still said "1,834". At 19:07 the writer
did exactly as instructed and wrote 1,834; at 19:10 the gate, reading the fresh fact,
rejected it — `1834 > 1790` fails tolerance `gte`. **The gate was right, and the page was
correctly stopped from deploying.**

But the cause is mine: **hand-managed prose plus auto-refreshed facts guarantees drift for
any fact whose value can FALL**, and it would have fired on every refresh cycle thereafter.
Fixed in `SQL_2026-07-26b_writer_block_volatile_figures.sql`, deliberately narrower than the
problem:

- **Monotonic-ish counts keep their dated snapshot** (active agent definitions, distinct
  agent types, live sites, backend services). They only grow, so a snapshot stays ≤ the
  live fact and `gte` keeps supporting it — and they are what produced the good result above.
- **Windowed and reaped metrics lose their absolute figure**: orchestrations-per-day rolls,
  work-items-completed is reaped (1,267 → 1,051 → lower). Both now carry a qualitative form
  the writer can still use ("over a thousand orchestrations a day") plus an explicit
  instruction not to state an exact figure.

**Transferable rule: a figure may live in the writer's prose block only if it cannot fall.**
Anything windowed, reaped or otherwise non-monotonic must be qualitative there and numeric
only in the register.

### Still open on this bug

- ~~**The Go audit (candidates 2+4) is committed but INERT until the next image roll.**~~
  **Superseded — LIVE in `v1.0.1170` and exercised on real traffic (above).**
  Verify with `strings /app/agent-chassis | grep -c stat_unit_impossible` against the
  running pod, never git.
- **The two live residuals above**, both blocked on a re-render.
- **Evidence registers for the remaining publishing sites** (owner's instruction, this
  session): webdesign.co.uk (98 live pages), finetuning.uk (41), gaswholesalers.com (28),
  idea.uk (20), vetcomparison.uk (5), dartsonline.com (4), plus `facts[]` for vonc and
  oufe. The 17 `pool-*.internal` rows have zero deployed pages and need none. Three want
  their owning thread consulted rather than a unilateral seed: **oufe** already runs a
  "no figure outside the evidence register" rail, **vetcomparison** carries a legal record
  from published fabricated prices, **idea.uk** is owned by another session.
- **Spec-aspect numeric drift** (above) — a number in a spec is an instruction, and
  nothing keeps it fresh.
- **043's point (c), partial blanking**, is deliberately NOT implemented: it cannot be
  detected from the extracted claim list because blank sentinels are dropped by design,
  and a check that can never fire is worse than none — it reads as coverage. It needs its
  own function over the raw `content_data`.
- Prose numbers outside stat fields were still not audited fleet-wide.

## Related

- `/bugs_open/020` — the tool-recreation fabrication. Same family, different path. Fix both.
- `/bugs_open/001` — re-plan resurrecting audited-out fabrication.
- `/bugs_open/023`, `/bugs_open/033` — the delivery gap. robot-hands had 20
  `cta_names_unknown_destination` items in `needs_human_review` and none of them were ever read;
  a detector for this class would rot the same way unless routing is fixed too.
- `docs/agent_docs/docs024_key_docs_latest/vetcomparison/LEGAL_2026-07-15_*` — what published
  unsourced figures cost the last time.
