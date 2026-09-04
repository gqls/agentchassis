# 482 — a `component_level='tool'` page bakes plausible-sounding FICTIONAL dated content directly into its template at generation time, with no path to the real facts that exist for the same site

Filed 2026-09-04, by the `calendar` session, from a report relayed by `boxingonline.com`
(owner-review thread) and independently re-verified and corrected before filing — the
report's headline claim ("counts to a date two days in the past") does not match what the
stored artefact actually contains, and the real shape is worse. **Status: OPEN, unowned.**

**Severity: MEDIUM-HIGH.** Live on a paid site, actively misleading real visitors with
fictional content presented as real, every day it stays unfixed.

## 0. First-hand verification, stated per CLAUDE.md's 2026-07-31 ruling

Structural claim, no `090` run. Substituting: every figure below was queried live against
the DB in this session today (2026-09-04), the component/schema facts are a direct read of
`content_components`/`page_components` rows, and the received report's own headline claim
was **not** taken on trust — re-derived, found to disagree, and the disagreement is
recorded in §2 rather than silently corrected. I could not curl the live served page
directly (no outbound internet from this session's sandbox); everything here is read from
the stored artefact (`page_components.rendered_html`/`content_data`), which is what the
platform believes it composed. A served-vs-stored divergence is possible and unchecked —
see §6.

## 1. The symptom, boxingonline.com

**`/tools/fight-countdown/index.html`**, component `tool-fight-countdown`
(`page_components.rendered_html`, 13,475 bytes) `[MEASURED 2026-09-04]`: six fights
hardcoded directly in the component's own JS, verbatim:

```js
var fights = {
  fight1: { title: 'Usyk vs Fury II, Undisputed Heavyweight Title', ... year: 2025, month: 5, day: 14, ... },
  fight2: { title: 'Canelo vs Benavidez, Super Middleweight Unification', ... year: 2025, month: 6, day: 28, ... },
  fight3: { title: 'Inoue vs Nery, Bantamweight Title', ... year: 2025, month: 5, day: 5, ... },
  fight4: { title: 'Davis vs Garcia II, Lightweight Title', ... year: 2025, month: 8, day: 11, ... },
  fight5: { title: 'Joshua vs Wilder II, Heavyweight Clash', ... year: 2025, month: 9, day: 6, ... },
  fight6: { title: 'Taylor vs Serrano III, Lightweight Trilogy', ... year: 2025, month: 6, day: 20, ... }
};
```

**Every one of the six is dated `year: 2025`** — today is 2026-09-04, so every option is
over a year stale, not "two days" as first reported (see §2 for the correction). The tool's
own `updateCountdown()` logic checks `diff <= 0` and, when true, replaces the countdown with
*"This fight has started or concluded. Check the fight calendar for results."* — so
**selecting any of the six built-in options triggers that message immediately.** The tool
is not partially stale; it is non-functional for every option it offers.

**None of the six is the real, currently-scheduled fight.** `evidence_base` for this site
holds, among 8 facts `[MEASURED 2026-09-04]`, a genuine forward-looking dated fixture:

```json
{
  "claim": "Canelo Alvarez is scheduled to fight Christian Mbilli on October 31",
  "event_date": "2026-10-31",
  "participants": ["Canelo Alvarez", "Christian Mbilli"],
  "source": {"citation": {"publisher": "Boxing News Online", ...}}
}
```

written by `evidence-researcher` (2026-09-02) and re-verified by `evidence-refresher`
(2026-09-03) — real, dated, cited, current. The countdown tool has no connection to it at
all: **the real fixture that exists is not one of the fictional ones offered.**

**`/tools/fighter-comparator/index.html`**, component `tool-fighter-comparator`
`[MEASURED 2026-09-04]`: **0 `<option>` elements, 0 `<select>` elements** in 22,126 bytes of
rendered HTML — no fighter list to compare at all.

## 2. Correction to the report this bug was filed from

The relayed report (from `boxingonline.com`, quoting the owner) said: *"the only ISO date
anywhere in the page is `2026-09-02`"* and characterised the defect as the countdown
running two days stale. **That figure does not appear anywhere in the stored component**
— `grep -oE '202[0-9]-[0-9]{2}-[0-9]{2}'` over the full 13,475-byte `rendered_html` returns
nothing; the dates are JS object literals (`year: 2025, month: 5, day: 14`), not ISO
strings, and every one reads 2025, not any single day in September 2026. Recorded rather
than silently used, because propagating an unverified figure into a bug file is exactly
what this estate's own practice warns against, and the true shape (six dead options, all a
year-plus stale, none real) is materially different from the reported one (one countdown,
two days late) in a way that changes how bad this actually is. **Do not carry the "2026-09-02
/ two days" figure forward.**

## 3. Root cause: this is generation-time fabrication, not a consumption gap

`bugs_open/427` diagnosed and fixed "nothing writes dated, cited facts" (the writer side)
and separately wired one component (`event-list`, via `query.upcoming_events`) to consume
them (the reader side, now live pending its own roll — see 427 §17-18). **This bug is a
different, sibling defect, not a new instance of 427's.**

`[MEASURED 2026-09-04]`:
```sql
SELECT function, component_level, length(html_template), input_schema IS NOT NULL
FROM content_components WHERE function IN ('tool-fight-countdown','tool-fighter-comparator');
```
→ `component_level='tool'` for both, **`input_schema` is NULL for both**, and the only Go
template variable in either `html_template` is `{{.InstanceID}}`. There is no
`content_data` path here at all — unlike `event-list` (a section-level component templated
from writer-authored `content_data`, which is what 427's fix populates), **a `tool`-level
component's content is baked into its own static template/JS at the moment the tool is
generated**, with no field, schema or seam through which `evidence_base` — or any other
live data — could ever reach it after the fact.

**So whatever generated `tool-fight-countdown` invented six plausible-sounding boxing
matchups with plausible-sounding dates, wrote them as literals into the tool's own
JavaScript, and had no step that checked whether any of that corresponded to something
real** — the identical shape to the already-known "planner logo exemplar licenses a
wordmark it never names, so the image model invents a brand" defect
(`bugs_open/417`/its closed history) and to the original boxingonline review's own §6
finding, *"the tool-suggester's own recorded reasoning… never once asks whether we hold
data that would let the tool answer anything."* This is a third sighting of the same
family: **a generator asked to produce something specific, given nothing real to draw on,
fabricates something that reads as real.**

`[UNMEASURED, left to the fixing thread]`: which action/agent actually generated these two
tools, and whether it had `evidence_base` available to it at generation time and ignored
it, or never had access to it at all. That is a call-graph/prompt-trace question this
filing did not do — establishing it decides whether the fix is "make the generator consult
`evidence_base`" or "make the generator refuse to fabricate specific real-world dated
content when it has nothing to draw on", which are different fixes.

## 4. This was partially detected already, and deliberately not repaired — but not precisely

Checked the `filing_mode: record` mechanism (RFC_056) the report named, rather than taking
the citation on trust `[MEASURED 2026-09-04]`:

```sql
SELECT count(*), count(DISTINCT site_id) FROM site_work_items
WHERE status='deferred' AND spec->>'filing_mode'='record';
-- 3,184 | 39
```
(the report's own figure, 3,181/39, is within population drift of this — the mechanism and
scale are both real, confirmed independently.)

All five cited verdict rows exist and are genuinely `status='deferred'`,
`filing_mode='record'`, `not_dispatchable` by design (`bugs_closed/077`'s convention). But
read individually, not just counted:

- **The four calendar-related rows** (`3acce370`, `70e4ed6c`, `dd07d0c7`, `9877c9a0`) all
  describe the calendar as *"mentioned in the homepage article grid subtitle… not surfaced
  as a standalone navigable section"* — this **predates** `/tools/fight-calendar/` existing
  as a real, standalone page (built as part of `bugs_open/427`'s work). They are stale
  findings about a state that no longer holds, not current descriptions of this bug.
- **The one comparator row** (`2c38eec5`) is about the tool having no explanatory heading
  ("no heading indicating how to use it, what data it draws on, or what it will show") —
  real, but a different complaint from "zero options, zero selects". Both point at the same
  underlying emptiness; neither names it as precisely as §1 does here.
- **Nothing in any of the five names the countdown's fabricated-date defect specifically.**
  That appears to be genuinely new, not a case of "five seats already caught this and it
  was parked" — the record-mode mechanism worked as designed for the calendar's earlier
  state and for the comparator's missing explanation, but this specific failure mode slipped
  past all of them.

## 5. What the owner is asking for, separately from the fix

Relayed, and worth recording precisely because it names a different, structural ask: *"The
tool provenance should be recording all this and we shouldn't be making these mistakes."*
Two candidate mechanical checks, neither needing an LLM:
- a hardcoded/baked date in a tool's own content should never resolve to `< today()` at
  serve time (a `max(target_date)` assertion against whatever dates a tool template embeds);
- an interactive tool with a `<select>`-driven interface should never ship with zero
  `<option>` elements.

**Not this lane's build.** Where this belongs is a real open question, not a settled one:
it could be `experience_loop`'s acceptance/audit machinery (an existing check-running
mechanism, per §4's own record-mode findings already running against this exact site), or
it could belong with whatever action generates `component_level='tool'` pages in the first
place (a build-time refusal, matching `period-calendar`'s own precedent of refusing a
date-shaped field by design rather than allowing one that can go stale — VIZ-017,
`605_period_calendar_component.sql` rule 1). Naming both rather than picking one; the
fixing thread should read both before choosing.

## 6. What's NOT established

- Whether the live served page matches this stored artefact — could not check, no outbound
  internet from this session. If it diverges, that is a separate, real finding (a
  publish/cache gap), not this bug's root cause.
- Which action generated these two tools (§3's open question).
- Whether other sites' `component_level='tool'` pages carry the same class of fabricated,
  now-stale, real-world-shaped content — this filing checked exactly two tools on one site,
  found by report, not by a fleet census. A fleet-wide census (any tool template embedding a
  literal date, checked against `now()`) is the natural next measurement and was not run
  here — flagged rather than silently assumed to be a two-tool, one-site problem.

## 7. Fix candidates — named, not decided

1. **Immediate content correction on this site**: rewrite the two tools' baked-in data using
   the real `evidence_base` facts (a content/component rewrite, not a code fix) — closes the
   symptom on boxingonline specifically, fixes nothing structurally.
2. **Structural, generation-time**: whoever builds a `tool`-level page with real-world dated
   content should consult `evidence_base` at generation time — populate genuinely from it,
   or decline to fabricate a specific real-world matchup/date when there is nothing to draw
   on (mirroring `period-calendar`'s own refusal precedent).
3. **Structural, serve-time backstop**: the owner's own ask — a cheap, mechanical
   provenance/acceptance check (§5) — belonging either to `experience_loop`'s existing audit
   machinery or to the tool-generation pipeline itself.
4. **Fleet census**: before scoping any of the above, measure whether this is a two-tool
   anomaly or a class — grep `content_components` for `component_level='tool'` templates
   embedding a literal `year:`/date pattern, cross-reference against `now()`.

None of these is `calendar_component`'s build — tool generation is a different pipeline
than the content components this lane owns, and the serve-time check is audit/acceptance
territory. Filed here because "why is the thing that should show a real fight showing a
fictional one" is exactly this lane's question to answer, the same standing this lane took
on `bugs_open/427`.

## 8. Cross-references

- **`bugs_open/427` §22-23 — read this FIRST, not just for cross-reference.** 427 hit the
  identical mechanism the same day, on a *different* tool: dispatching `tool-generator` for
  the fight-calendar page itself produced 12 fabricated, stale fixtures (Canelo/Charlo,
  Fury/Usyk, Wilder/Joshua…), caught before deploy, and §23 designs a three-layer checker
  plan for it — **not built, awaiting the owner's go-ahead.** Added a correction there
  (§24) after finding this bug: **none of the three proposed layers would have caught
  `tool-fight-countdown`'s violation** — layer 2's shape is ISO-date-string keys, this
  tool's dates are `year`/`month`/`day` numeric triplets; layer 3 validates `data-fact-id`
  attributes, this tool has none. Read §24 before scoping any fix for either bug — the
  detection shape needs widening, and a census over already-built tools is needed
  regardless (birth-time refusal only stops new fabrication).
- `bugs_open/417` (closed) — the wordmark-invention defect; same family (a generator with
  nothing real to draw on fabricates something that reads as real).
- `docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/OWNER_REVIEW_2026-08-31_boxingonline_what_he_found_and_what_each_finding_actually_is.md`
  §6 — "the tools make the reader supply the data", the first sighting of this family on
  this site.
- `docs/agent_docs/docs026_concept_register/register/visualisation-and-charts.md` VIZ-017 —
  `period-calendar`'s design precedent for refusing date-shaped fields by design.
- `docs/agent_docs/docs024_key_docs_latest/calendar_component/` — this lane's own docs.

---

## 9. Status update, 2026-09-04 — RESUMED as the fixing lane, and the bug is not the one that was filed

Picked up by the `bugfix_482_tool_fabrication_fence` lane. It was filed unowned (§0 header), §7
disclaims it for `calendar_component`, and the two citing lanes disclaim it too; `ListAgents`
showed no session working it. The `calendar` lane confirmed by message: *"No objection, it's
yours."* Working docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_482_tool_fabrication_fence/`.

### 9.1 Still valid, re-derived not trusted

`[MEASURED 2026-09-04]` both components `component_level='tool'`, `is_active=t`, `input_schema`
NULL, 13,279 B / 21,426 B, born 2026-08-31. All six invented fights still present verbatim.
**Live and unchanged.**

### 9.2 §3's `[UNMEASURED]` is closed: the generator never had `evidence_base`

Birth path is `tool-generator` → `create_tool_component_action.go`. Workflow steps
`[MEASURED 2026-09-04]`: `ensure_site_record, load_brand_context, load_site_page_names,
compose_plan, write_plan, index_plan, generate_tool_html, suggest_related_pages, save_tool,
enqueue_rerender, complete`. **No step loads or consults `evidence_base`.** The
`generate_tool_html.config.prompt_template` is 5,118 chars, 22 numbered rules about ids, colours,
IIFEs and readouts, with **zero** occurrences of fabricate / invent / evidence_base / provenance /
verify.

So §3's fork resolves to **"never had access"**, not "had it and ignored it" — which selects
§7's fix candidate 2 (make the generator unable to fabricate) over "make the generator consult
`evidence_base`", and independently confirms `427` §23.1's finding about the brief inviting the
fabrication.

### 9.3 THE FINDING: a fabrication gate already exists, is live, and DISCARDS ITS OWN VERDICT

This reframes the bug, and it is why 482 deserved to be filed separately from 427 rather than
folded into it.

`platform/orchestration/actions/check_tool_fabrication_action.go` — 459 lines, built for
`bugs_open/020`, council-reviewed, negation-aware via the shared `datahelpers.NegationGuard`,
tiered A (declaration) / B (corpus signature + corroboration). It is **live**, and
`[MEASURED 2026-09-04]` wired in the `default_config` of exactly **one** active agent definition:
`tool-recreation-handler`. **The birth path does not consult it at all.**

Probed via its exported pure core `DetectToolFabrication`, against real `html_template` bytes,
**with a positive control in the same run** (recipe and its gotchas in the lane's RUNBOOK):

| component | `Fabricated` | `Signals` |
|---|---|---|
| `tool-fight-countdown` | false | *(none)* |
| `tool-fighter-comparator` | false | *(none)* |
| `tool-budget-kit-builder-garden-tools-uk` | **false** | large literal record array (~20 entity objects) |
| `tool-sfi26-revenue-stacker-agritec-uk` | **false** | large literal record array (~24 entity objects) |
| `tool-vet-comparison-vetcomparison-uk` | **false** | large literal record array (~30 entity objects) |
| CONTROL — the `bugs_open/020` vetcomparison shape | **true** (`tier="declaration"`) | declared synthetic data; `makePostcode` introduced |

The control convicts, so the zeros are real zeros and not a broken harness.

**Read the middle three rows carefully. The gate is NOT blind to them.** It computes the corpus
signature — 20, 24 and 30 entity-record objects, all over `fabLiteralRecordThreshold = 15` — and
returns `Fabricated=false` anyway, because Tier B gates on `dataBacked && !preserved`, and
`dataBacked` derives from an `original` that a **born** tool has never had. The signals are
returned "for observability" and gated on nothing.

> **⚠ CORRECTION to this lane's own earlier claim, made the same day.** I told two lanes that
> *"Tier B is UNREACHABLE at birth"*. **Wrong word.** The *signature* is reachable and reached;
> the **conviction** is unreachable. The distinction is load-bearing: it means the birth arm is
> **not new detection work**. The evidence is already computed, already correct, and discarded.
> Caught by feeding the detector more inputs than the two tools in this bug — i.e. by `427`'s
> census, not by my own reasoning.

**The same shape appears independently in `427`'s half**: `projectUpcomingEvents`
(`queryresolve/upcoming_events.go` ~246) emits `fact_id` into every item map, and the `event-list`
template renders `.date/.title/.venue/.broadcaster/.source_url` and **never `.fact_id`** — so a
fact-backed fixture and a fabricated one are byte-indistinguishable at the served artefact. Two
independent instances of *"the estate already knows and throws it away"* on one root cause. This
also means **`427` §23.2 Layer 3 is not merely blind on this bug's shape — it is structurally
unreachable today**, because its subject is the served artefact and `data-fact-id` appears **zero**
times there even for the component that is doing everything right (measured independently by the
`boxingonline.com` lane).

**Bitter footnote:** `bugs_open/020`, the bug the fabrication gate was built for, was filed about
**vetcomparison.uk**. The gate was written for that site. That site is still fabricating, in a
shape the gate was never taught to convict.

### 9.4 §6's census IS a class — and this lane's first attempt at it was wrong

⚠ **Correcting §6's implied scope and this lane's own first measurement.** Keyed on **date
shapes** over all 335 active tool components, the census returns **1** — and I reported that to
two lanes as *"a first occurrence rather than a class"*. The `427` lane censused the same rows on
an **entity+attribute** axis (a record that identifies a real-world thing and attributes a
checkable property to it) and found five candidates. Verified at first hand here, all
`component_level='tool'`, all `is_active=t`:

| component | records | what is invented |
|---|---|---|
| `tool-vet-comparison-vetcomparison-uk` | **30** | UK veterinary practices + postcodes + websites |
| `tool-sfi26-revenue-stacker-agritec-uk` | 24 | UK government scheme payment rates, real scheme codes |
| `tool-budget-kit-builder-garden-tools-uk` | 20 | product price bands |
| `tool-fight-countdown` (this bug) | 6 | boxing fixtures |
| `tool-loot-table-balancer-gamesdesign-co-uk` | 3 | *(probably legitimate game vocabulary, not a real-world claim)* |

**`tool-vet-comparison-vetcomparison-uk` contains no date at all**, so no widening of a date
predicate could ever have reached it. Full misstep entry in `WRONG_CALLS.md` (2026-09-04): *I
censused the shape I already knew about.*

### 9.5 The most exposed instance is not on boxingonline

`tool-vet-comparison-vetcomparison-uk`, 16,944 B, born 2026-09-02. Placement
`[MEASURED 2026-09-04]`: `vetcomparison.uk` **`/index.html`** — the homepage —
`page_components.build_status='deployed'`, `pages.deployed_at = 2026-09-03 21:19:33+00`. **30**
postcode-bearing records, **30** `https://example-vet-*.co.uk` hostnames, e.g.
`{ name: 'Willow Tree Veterinary Surgery', location: 'Chester', postcode: 'CH1 3AB', website: 'https://example-vet-chester.co.uk' }`.

Three things found by re-deriving the report rather than accepting it, none of which is in the
countdown's shape:
- `:298` — `// Bundled, verified sample of practices. Self-contained — no fetch().` The component
  **asserts its own verification**.
- `:290-291`, tool-doc header — *"Never seeds or fabricates practice records beyond the bundled
  list; if this list needs to grow, it must be replaced with a verified set."* The prohibition and
  the violation are in the same file.
- `:40`, **served to the public** — *"Practice details shown here are a fixed sample bundled with
  this tool… please confirm anything important directly with the practice. The RCVS maintains the
  official register of accredited practices."* It invites a member of the public to confirm the
  details of a practice that does not exist, citing the regulator's register while doing so.

⚠ **For the fence's calibration:** `:290`'s *"Never … fabricates …"* is precisely the shape
`fabNegationGuard` was built (`bugs_open/222`) to suppress, because a conscientious model echoing
the prompt's prohibition was the common false-positive case. **This is the first observed artefact
where the denial and the act are in the same file.** The guard is still right and should not be
weakened on this evidence alone — but any birth arm must be calibrated knowing this case exists.

**Dispatched nothing at that site.** The `vetcomparison` lane has the evidence; remediation of a
live commercial site is that lane's and the owner's call.

### 9.6 Scope, settled with the `427` lane by message

- **427** = the mechanism (dated correctable `evidence_base` facts), Layer 1
  (`check_event_fixture_completeness`), and the **provenance rail** — making a fact-backed fixture
  say so in the served markup, so that *absence of a declaration* becomes a signal instead of the
  default state of the estate.
- **482 (this lane)** = the **fence** (route every tool-writing path through the existing gate,
  make it convictable at birth, ratchet the coverage), the **census**, and **remediation**.
- Both lanes agree **not** to widen `ExtractAssertionText` to script bodies, and not to file that
  RFC: if the rail lands and the fence consumes it, widening the claims perimeter becomes
  *duplicative* rather than deferred. `427` is recording that in its §23.2.
- Neither lane will ship a fabrication-shape enumeration. Four detectors have now been shown blind
  or self-discarding on this class (427's three proposed layers plus the one already deployed);
  the shape space is unbounded and each widening is specified against the examples in hand.

### 9.7 The tool-template write paths, censused at HEAD

`[MEASURED 2026-09-04]`:

| path | file | fabrication gate? |
|---|---|---|
| birth **and** regeneration | `create_tool_component_action.go` (`regenerateToolComponentInPlace` called from `:307`, i.e. after the `:127` / `:160` gates) | **no** |
| tool-improver | `update_component_html_action.go` (calls `sharedComponentWriteCheck`) | **no** |
| tool-recreation | `tool-recreation-handler` workflow | **yes** — the only one |
| propagation to another site | `deploy_tool_action.go` | n/a — but it forks a row to a second site without re-inspecting it |

The remedy idiom already exists on this estate for exactly this failure class
(`bugs_closed/021`, *"durable write guard covers one path only"*):
`component_template_writer_coverage_test.go`, a source-scanning coverage ratchet that fails the
build when a writer of `content_components.html_template` neither calls the shared fence nor is
listed with a written reason. Its header states the argument better than I can: *"A header census
is true on the day it is written; this test is true on every build."* The fabrication gate has no
such ratchet, which is how it came to cover one path of several without anyone noticing.

### 9.8 Not yet decided, and deliberately not decided by this lane

- **Remediation of the live tools.** boxingonline: with the owner via the `boxingonline.com` lane.
  vetcomparison: with the `vetcomparison` lane. ⚠ **The two boxingonline tools want different
  answers** — `evidence_base` for that site holds 8 facts, 7 dated, **2** with
  `event_date >= today` (measured by the `boxingonline.com` lane), so the countdown can genuinely
  be *repaired*; the comparator needs fighter attribute data for which **no source exists on the
  estate**, so its only honest options are *withdraw* or *ship visibly empty*. A single
  repair-or-withdraw decision gets one of them wrong.
- **Whether a birth-time refusal needs a third outcome** besides build-it and refuse-it. The
  boxingonline lane's argument, which this lane accepts: a provenance requirement the fact supply
  cannot satisfy converts "fabricated tool" into "unbuildable tool" for a whole category, and a
  gate that refuses good work gets switched off, after which it protects nothing.

**Verification contract for any remediation**, agreed with the `boxingonline.com` lane, both
controls in the same run — because a page that has merely been blanked scores identically to a
page that has been repaired:
- *negative*: `Usyk|Fury|Joshua|Wilder|Benavidez|Nery|Serrano|Garcia` → 0, `year: 2025` → 0;
- *positive (anti-blanking)*: `<body>` present, non-trivial bytes against the current
  13,279 / 21,426, the tool's own heading present, and for the comparator `<select>`/`<option>`
  counts **NON-ZERO** — since "0 options" is the defect and an emptied page also scores 0.
