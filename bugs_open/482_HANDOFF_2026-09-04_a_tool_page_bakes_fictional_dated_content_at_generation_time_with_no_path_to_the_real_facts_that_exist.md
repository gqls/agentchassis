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
