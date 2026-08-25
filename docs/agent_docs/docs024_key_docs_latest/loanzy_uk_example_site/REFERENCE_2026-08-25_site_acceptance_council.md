# REFERENCE — the site acceptance council: what it is, what already exists, and the route

**Written 2026-08-25** by the `loanzy_uk_example_site` lane at the owner's request, after a day
spent measuring `homegarden.uk` (the authorised greenfield canary) against his review of it.
**Status: DIRECTION APPROVED by the owner — *"I like your suggestion"* — with two placements he
set. RFC not yet written; this is the reference it lifts from.** Every figure carries its date;
every claim not measured says so.

---

## 1. In one paragraph

The build path has no closing check. A site reports `complete` when its last work item closes, and
nothing then asks whether it delivered a site. Six broken routes were found on one site today —
research never ran, the offer analyser was off, the visual designer has no dispatch path, the news
feed cannot bootstrap, the listing renders nothing, the classifier cannot name a directory — and
**every one was found by one owner review**, because that review is the only closing check the
estate has. The remedy is not another rule; it is **a reviewer at the end**: several seats, each with
one lens, judging the **served artefact**, producing a verdict that **routes work to the agent that
owns the fix**. The estate has proven this pattern three times (code, plans, register entries). It
has never pointed it at a site.

> **⚠ And then, checking before writing this: it HAS. The improvement loop is that reviewer, it
> ran for nine days in August, and it has been switched off since 17 August.** See §3. This changes
> the route from "build a council" to "switch on, and add the four seats it lacks".

---

## 2. The owner's rulings, in his words, all 2026-08-25

| ruling | words | consequence |
|---|---|---|
| **Placement** | *"The checkers will check after the fact (improvement loop)"* | The council does **not** hold a build. The site deploys; the council reviews the served result and files improvement work. **No stall risk, and a new site and one of the 51 existing sites are judged by the identical operation** — the fleet census and the per-build check are the same thing. |
| **Routing** | *"we have other agents like the component planner and so forth too"* | A REVISE verdict routes to the **existing** agent that owns the stage — planner, writer, imagery, component planner — not to a new fixer. |
| **Direction** | *"I like your suggestion"* | Approved to write up; RFC is the next step (shared mechanism → architecture-scope). |
| **The floor** | *"implement at least 6 different structures unless the site doesn't warrant it or refuses it"* | Becomes the structure seat's rule. **N = 6 is his number**, the way N = 10 was for the optional-key budget (RFC_022). Refusal becomes the seat's **recorded** verdict, not the planner's silent option. |
| **Switch-on** (earlier today) | *"Switch on all those agents and we'll need to fix or further develop them as necessary"* | The `improvement-sweep` carrier is in that set. Routed to `vigilant_designer_offer_analysis`, who measured it and will enable it. |

---

## 3. THE FINDING: the council already exists as `improvement-loop`, and it is OFF

`[MEASURED 2026-08-25 15:3xZ, live non-snapshot `agent_definitions` + `scheduled_tasks` + `llm_call_log`]`

**`improvement-loop`** is a multi-seat, post-hoc site reviewer. Its steps call, in sequence:

| step | seat agent | kind | what it judges |
|---|---|---|---|
| `call_brief_fidelity` | `brief-fidelity-auditor` | model | is the built site faithful to **its own brief** (mission, design intent, content direction) |
| `call_completeness_discovery` | `completeness-discovery-agent` | **algorithmic** | empty sections, cross-site name contamination, unrendered template syntax |
| `call_quality_discovery` | `quality-discovery-agent` | **algorithmic** | broken nav links, placeholder/fabricated contact info, generic unthemed CSS |
| `call_design_audit` | `design-audit-agent` → `visual-design-auditor` + `content-quality-auditor` | model | CSS/layout/colour; tone/gaps/CTA |
| `call_offer_analyser` | `offer-analyser` | model | is the site's offer clear |
| `call_site_review` | role `site_reviewer` | `[UNVERIFIED — role not resolved to an agent type here]` | — |
| **`enrich_news_feed`** | action `evaluate_news_feed` | — | **the owner's news-feed ask, already a step** |
| **`enrich_directory_features`** | action `evaluate_directory_features` | — | **the owner's directory ask, already a step** |
| `check_not_converging` → `record_not_converging` | — | — | 3 passes at one fingerprint with findings still open → files a `capability_gap` item |
| `insert_rerender_item`, `triage_findings` | — | — | **routing**: findings become work |

**It ran.** `offer-analyser` 9 calls (08-14 → 08-22), `brief-fidelity-auditor` 6 (08-13 → 08-22);
**10 work items across 7 sites, last 08-22.** The algorithmic seats leave no `llm_call_log` rows,
so their firing cannot be read from that table — `[UNMEASURED]` whether they ran.

**It stopped.** Its only carrier, `improvement-sweep` (15-minute interval), is **`enabled=false`,
last triggered 2026-08-17 12:30:44Z.** Nine days of running, then off. The owner's switch-on ruling
above re-enables it.

> ⚠ **Two traps in the loop as it stands, both measured from its config:**
> 1. **`enrich_news_feed` has `error_step: enrich_directory_features`, and `enrich_directory_features`
>    has `error_step: load_audit_state`.** So if either of the two steps that ARE the owner's asks
>    fails, the loop **swallows it and continues**. A site with no news and no directory passes the
>    loop with no record of why. This is the same fail-open shape as `bugs_open/380`'s auditor.
> 2. **`brief-fidelity-auditor` judges against the BRIEF, and the brief can be the problem.**
>    `homegarden.uk`'s brief was anti-commercial; a faithful build of it has no directory. Fidelity
>    to the brief is a real seat, but it is **not** the owner's *"happy user"* lens — that lens judges
>    against what a reader of *this kind of site* wants, which a brief may not have anticipated.
>    The two seats must both exist, and they will sometimes disagree, which is the point.

---

## 4. The seats: what exists, what is missing

Every seat is **mechanical** (a query or a fetch, free, runs every time) or **model** (relevance-
gated, costs credits, like the code council's). The design principle from today: **prefer
mechanical where a query can answer**, because a model seat can be wrong in the confident register
and a query cannot.

### 4a. Exists in the loop today

| lens | seat | kind | note |
|---|---|---|---|
| Offer | `offer-analyser` | model | live in the loop; carriers off |
| Fidelity to brief | `brief-fidelity-auditor` | model | see trap 2 above |
| Completeness | `completeness-discovery-agent` | mechanical | covers **empty sections** — one of the four prerequisites |
| Integrity | `quality-discovery-agent` | mechanical | broken nav, placeholder contact |
| Design | `design-audit-agent` | model | tone/gaps/CTA via `content-quality-auditor` — partial reader lens |
| News / directory | `enrich_*` steps | — | present; **fail-open** |

### 4b. Missing — the four to add

| lens | rule | kind | evidence it is needed (2026-08-25) |
|---|---|---|---|
| **Prerequisites** | research ran · feed sources seeded · evidence base populated **by research** (not minted — `380` D1 stands) | mechanical, 3 queries | `research-agent`: **0** `llm_call_log` rows in its life; `content_sources` for homegarden **0** (9 of 51 sites enrolled); `evidence_base` absent on 33 of 48 sites |
| **Promise** | every page delivers what its own headings say | mechanical, at the served page | `after_test.sh` PROMISE-vs-DELIVERY: fires on `garden-tools` seasonal-planner (3 months under "month by month"), on 3 homegarden index pages; silent on 18 |
| **Structure** | **≥ N delivered** structures across the served site, **or a recorded refusal per shortfall** | mechanical count + recorded verdict | homegarden: 21 pages, **17 identical** (`hero`, `generic-text-block`, `content-listing`) |
| **Freshness** | current period, live links, dated content | mechanical | *"Check what's due in the garden this **April**"* — in August |
| **Depth** | research behind every page | mechanical (rows) + model (does the page use it) | zero research behind any page on any site built this way |
| **Reader** | *would a person on this kind of site want this page?* | model | about.html: **14 of 17** headings about the site's own methodology; "Get Started" → contact on a gardening site |

**Reader is the seat the owner has asked for three times in three words** — "user experience agent",
"happy user", "adversarially increase the quality". It is distinct from fidelity-to-brief (§3 trap 2)
and from `content-quality-auditor` (which checks tone, not purpose).

### 4c. How the structure seat works — the count, tightened

*Structure* is defined by **what the reader gets**, not by component name: list, table, calendar,
checklist, directory, feed, tool, guide, comparison. Hero, nav, footer and CTA do not count.
*Delivered* means the promise test passes for that structure — a calendar rendering twelve months
counts; a comparison index with no table does not. Below N, the seat does not ask the planner whether
it minds: it **records why** ("5 not 6: brief is anti-commercial, directory ruled out"). Refusals are
then auditable — rare and reasonable means the floor works; common and thin means the prompt is wrong.

**Why a count beats a checklist:** it is vertical-agnostic (no collision with the classifier's
taxonomy), it is one number (the RFC_022 governance shape, already accepted), and it forces the
planner to *explore* a 47-component vocabulary that most sites use three of. **Why it is not
enough alone:** it measures breadth. Depth is the seat beside it. Neither touches copy or imagery.

---

## 5. Verdicts and routing

Same three verdicts as the code council, same deterministic decision action
(`experience-approval-council` already reuses it). **After the fact, so:**

| verdict | effect |
|---|---|
| **APPROVED** | nothing filed; `record_audit_pass` |
| **REVISE** | one improvement item **per seat finding**, routed to the agent that owns the stage — planner (structure, freshness), writer (depth, reader), imagery, component planner |
| **REJECTED** | site is live but judged unfit; escalates to the owner rather than to a fixer |

The loop's existing convergence rule stands: three passes at one fingerprint with findings still
open → `capability_gap`, which is the honest name for "the fixer cannot fix this".

---

## 6. What this absorbs, so the two earlier proposals do not survive as separate things

| proposal | becomes |
|---|---|
| **benchmark checklist** (owner, first framing) | the mechanical seats — each item a query, enforced at the artefact, not a prompt clause. **BLD-006** (register: *"coverage baseline: guides, tools, news, curated top-N"*, status **aspirational**, enforcement *"the strategist/planner prompts"*) is superseded, not joined |
| **"at least 6 structures unless it refuses"** (owner, second framing) | the structure seat, §4c |
| **"a user experience agent"** | the reader seat |
| **"the offer and benefit analysis agent"** | already a seat; carrier off |
| **"let the visual designer know it hasn't done its job"** | it has no dispatch path (`[MEASURED]` by the vigilant lane, verified here); routing is authorised |
| **"point them at the experience loop to adversarially increase quality"** | this IS that, with the loop's name |

---

## 7. Why this route — now that most of it exists

1. **It is the estate's most proven pattern**, applied where it was missing. Code council: ~80%
   approval, real rejections. Experience loop: eight escalations, every one correct. Improvement
   loop: **ran for nine days and filed real work** before it was switched off.
2. **Many lenses, not one.** Fifteen wrong calls across two lanes today, **none caught by its
   author**. A single "happy user" agent is a single author.
3. **It surfaces broken routes instead of needing them fixed first.** A prerequisite seat finding
   zero research files that *against the writer*. A structure seat finding no directory type files
   that *against the taxonomy*. Today's six were found by hand; the loop finds them every pass.
4. **It converts silent absences into named failures.** Every defect today was an absence — no
   error, `complete` status, nothing to grep. A seat verdict is a row.
5. **Cost is bounded the way the code council's is** — relevance-gated model seats, free mechanical
   ones, and the loop's convergence cap stops it re-judging the same fingerprint for ever.
6. **After-the-fact placement (the owner's) removes the stall risk** that a gate would carry, and
   makes the 51-site census the same operation as a new build's review.

---

## 8. What it collides with, and why the loop surfaces rather than blocks

| collision | how it resolves |
|---|---|
| `bugs_open/380` ruling D1 — no shell evidence registers | the prerequisite seat asks for research that **populates** a register; it never mints one |
| the classifier's taxonomy has no `directory` / `news` type (15 categories, `[MEASURED]`) | a structure seat that cannot find one files a finding against the taxonomy; the seat is not blocked |
| the 51 existing sites fail on day one (`[MEASURED]` 13 of 27 with zero inline images; 42 of 51 with no news) | after-the-fact placement means the first fleet pass **is a census** — verdicts are the estate's score, filed as findings, not 51 bugs |
| the unaided-route measurement ends | the loop IS the measurement, on every build, dated |
| "unless determined otherwise" needs an override | the recorded refusal *is* the override, and it is auditable — no default-OFF switch (which the 07-29 ruling warns rots) |
| shared mechanism → council-gate scope | yes: **the seat definitions are the benchmark**, and the architecture seat should rule on them. RFC. |

---

## 9. What this does NOT do

- **It does not fix the six broken routes.** It finds them, every pass. Fixing research dispatch,
  the news bootstrap, the visual designer's path, the listing invalidation and the taxonomy gap are
  each their own work, owned elsewhere (`bugs_open/376`, `384`, `316`; the vigilant lane; this lane).
- **It does not judge copy quality.** The reader seat judges *purpose*; the copy audit the owner
  instructed today (`copy_quality_two_stage`: deep refresh, then audit every prompt) judges *prose*.
- **It does not replace the brief.** Fidelity-to-brief stays a seat. The reader seat is the second
  opinion the brief cannot give about itself.

---

## 10. The route, in order

1. **`improvement-sweep` back on** — authorised, routed to `vigilant_designer_offer_analysis`, who
   hold the measurement of what it selects and will stage it. *Expected to reveal defects; that is
   what "fix or further develop as necessary" licensed.*
2. **Close the two fail-open `error_step`s** on `enrich_news_feed` / `enrich_directory_features`
   so a failed enrichment leaves a row. Config-only; one migration; council-gated.
3. **The RFC** — seats, N, verdict routing, the `380`/taxonomy collisions, BLD-006 superseded.
4. **The two cheapest new seats first**, both mechanical and both existing as code today: the
   prerequisite queries and the promise harness
   (`loanzy_uk_example_site/after_test.sh`, PROMISE-vs-DELIVERY section, chrome-stripped, control-gated).
5. **Then structure (N = 6) and freshness**, mechanical.
6. **Then the reader seat**, model, relevance-gated.
7. **Then depth**, once `research-agent` is actually driven — a depth seat before that files the same
   finding on every site, which is a census, not a review.

---

## 11. Falsifiers — what would show this reference is wrong

- `improvement-loop` switched on and **producing nothing routable** — then the loop is not the
  reviewer it looks like, and the council is a new build after all.
- A structure count of **6** being met trivially by chrome — then §4c's definition of "structure"
  is too loose and must name the shapes.
- The reader seat agreeing with `brief-fidelity-auditor` on every site — then one of them is
  redundant, or the reader prompt is judging fidelity rather than purpose.
- Refusals filed on **most** sites — then N is too high for the vocabulary, or the planner has
  learned that refusing is cheaper than composing.
- A mechanical seat passing a page the owner rejects — the finetuning lane's finding today, one
  level up: *the owner's tell class is wider than the gate's*. The reader seat exists for that
  case; if it is also fooled, this document's premise is wrong.
