# 378 — `content_components.usage_count` is written on ONE of the two resolution paths and read as a quality signal on that same path, so a component's score records which route found it rather than whether it is any good

**Filed:** 2026-08-24 by the `bugs_open/351` lane, found while answering a council objection about
which path resolved a component. Not a symptom anyone reported — the numbers are simply wrong in a
direction that always favours the same kind of component.

**Severity:** latent, structural. Nothing errors, nothing is refused, no page is broken today. The
cost is a selection input that is silently biased, and a figure that reads as evidence in bug files
and hand queries when it is not.

**Status: OPEN, not started.** Diagnosis below is first-hand and complete; no code written.

## The finding in one paragraph

A section component can be resolved two ways. **Path 1** matches the planner's section name against
`content_components.name` then `function` (`plan_sections_action.go:1258` → `loadComponentSchemas`
→ `loadSectionComponents`, `v3_site_actions.go:4958`). **Path 2** is the selector, which matches
`section_type` and *scores* the candidates (`component_selector.go`). **`IncrementUsageCount` is
called on Path 2 only** — one non-test call site, `plan_sections_action.go:1957`, inside
`resolveSectionComponent`. A component resolved by Path 1 is bound to a live page and never counted.
And `usage_count` is then **read as a scoring input on Path 2**, under a file header calling it
*"battle-tested components score higher"*. So the term does not measure how proven a component is;
it measures how often it was reached by the route that happens to do the counting.

## Evidence

### The writer: one call site, and it is inside the scoring path

```
$ grep -rn "IncrementUsageCount" --include="*.go" platform/ internal/ pkg/ cmd/ | grep -v _test.go
platform/orchestration/actions/plan_sections_action.go:1957      # inside resolveSectionComponent = Path 2
platform/orchestration/actions/component_selector.go:131,133     # the definition itself
```

`resolveSectionComponent` (`plan_sections_action.go:1926`) is reached only from Path 2 of the section
loop. Path 1 (`:1258`) returns `comp` straight from the `components` map and calls nothing.

### The reader: the same column, weighted, in both selector queries

`component_selector.go:181` and `:235`, identically:

```sql
+ LEAST(COALESCE(usage_count, 0)::float / 50.0, 1.0) * 0.1
```

### The gap, measured

`[MEASURED 2026-08-23]`, active, non-forked, `component_level='section'`:

| | count |
|---|---|
| have any `usage_count` at all | **12** of 149 |
| `usage_count = 0` **and** ≥1 live `page_components` binding | **96** of 149 |
| page bindings invisible to the scoring term | **1,802** |

```sql
WITH b AS (SELECT component_id, count(*) AS bindings FROM page_components
           WHERE component_id IS NOT NULL GROUP BY 1)
SELECT count(*) FILTER (WHERE cc.usage_count = 0 AND b.bindings >= 1),
       count(*) FILTER (WHERE cc.usage_count > 0),
       count(*), sum(b.bindings) FILTER (WHERE cc.usage_count = 0)
FROM content_components cc LEFT JOIN b ON b.component_id = cc.id
WHERE cc.is_active AND cc.component_level='section' AND cc.forked_from IS NULL;
```

### The correlation that shows it is mechanical, not noise

If `usage_count` were a genuine popularity measure it would be independent of `section_type`. It is
not — a NULL `section_type` makes Path 2 structurally impossible (`WHERE section_type = $1`, and
`NULL = anything` is never true), so those rows can never be counted:

```sql
SELECT count(*) FROM content_components WHERE section_type IS NULL AND usage_count > 0;  -- 0
```

**Zero, fleet-wide as of 2026-08-23.** The two facts are perfectly correlated across every row in
the table, which is what the mechanism predicts and chance does not.

Three concrete rows: `loans-credit-health-check` (`824e3309`) and `loans-damage-checker`
(`966e97dd`) were bound to live pages on `loanzy.uk` on 2026-08-23 (13:57Z, 14:07Z, 14:23Z) and
still read `usage_count = 0`. Meanwhile `bayesian-ranking-hero-tool` reads 20 and
`case-studies-grid` 19.

## Why it matters, in order

1. **The bias is systematic and always in the same direction.** A component reachable only by
   `function` scores permanently as unproven; one carrying a `section_type` accrues. So the selector
   prefers whichever component was already selectable, and the metric that is supposed to reward a
   proven component instead rewards a *discoverable* one.
2. **It has already distorted a decision.** `bugs_open/351`'s ruling of 2026-08-23 declined a
   `section_type` backfill partly on this ground: a backfill would have entered 21 library
   incumbents (`usage_count = 0`, because Path 1 is all they have ever had) into scoring against
   their own site-specific twins (1–4), on a number measuring nothing but route. The right call was
   still made, but only because the defect was noticed first.
3. **The figure reads as evidence.** It is a plausible-looking integer in a column named
   `usage_count`. Nothing marks it as partial. `bugs_open/351` records a session (mine) checking it
   and having to be told by a control that it does not mean what it says.

## This is `bugs_closed/060`'s class, one table over

`060` — *"the platform keeps no durable record of which agents have ever run (`usage_count` is
dead)"* — found `agent_definitions.usage_count` at `0` for **every** active agent and fixed it
2026-07-26. **Same column name, same failure, different table.**

⚠ **This instance is arguably worse than 060's, and the difference is the point.** 060's counter was
**dead** — uniformly zero, so anyone querying it saw immediately that it meant nothing. This one is
**alive on one path**, so it returns a spread of plausible non-zero values (7, 19, 20) and looks
maintained. **A half-written counter is harder to catch than a dead one**, because the disconfirming
observation (a zero) is indistinguishable from a legitimately unused component.

## Fix candidates, ranked by what closes the door

1. **Count at the BINDING, not at the resolution.** The durable fact is
   `page_components` — a component bound to a page was used, whichever path found it.
   `usage_count` becomes derived (a view, or a counter maintained where the binding is written,
   `save_page_sections_action.go`'s single INSERT). Makes the biased state unrepresentable rather
   than merely corrected. ⚠ Needs a decision on whether the historical 1,802 bindings are
   backfilled; a backfill changes every selector score at once, so it is a behavioural change and
   not bookkeeping.
2. **Call `IncrementUsageCount` on Path 1 too.** Two lines, closes the gap for new resolutions,
   leaves the existing 96 rows wrong for ever and leaves the counter as something a third path can
   forget again. A patch, not a fix.
3. **Stop scoring on it until it is trustworthy** — drop the `* 0.1` term, or gate it on a
   `usage_count_is_complete` flag. Honest, reversible, and the only candidate that improves
   selection *today* rather than after a backfill.
4. **Do nothing but document it.** Defensible on cost (the term is weighted 0.1 of a ~0.16 base, so
   it rarely decides a contest on its own) — but the bias is monotone, so "rarely decides" is not
   "never decides", and nothing tells the next reader.

⚠ **Whichever is chosen, `content_components.usage_count` must stop being quotable as a usage figure
until it is fixed.** Read `page_components` instead.

## What is NOT established, marked so nobody quotes it as measured

- `[UNMEASURED]` **whether the 0.1 term has ever actually flipped a selection.** The other inputs
  (site-type relevance, page-type relevance, quality, specificity) may dominate in practice. This is
  the query that would size the problem and it has not been run — a pairing where the term is
  decisive would justify candidate 1 or 3; if none exists, candidate 4 gets much stronger.
- `[UNMEASURED]` whether `component_level='tool'` and the chrome levels have the same asymmetry.
  Only the `section` population was censused.
- `[INFERRED]` that no *third* writer exists. The grep above is exhaustive over `platform/`,
  `internal/`, `pkg/` and `cmd/` for the Go helper, but a hand-written migration could have set the
  column directly and would not appear. `component_versions`/`git log` over
  `docs/agent_docs/sql_for_agents/` would settle it.

## Diagnosis-loop note (owner ruling 2026-07-31)

`090` was **not** run. Stating the substitute plainly, as that ruling requires: the writer claim is
an exhaustive grep of all four Go trees for the only helper that mutates the column; the reader
claim is the two scoring queries quoted verbatim; and the mechanism claim is confirmed by a
fleet-wide correlation (`section_type IS NULL AND usage_count > 0` → **0** of 149) that would have
been non-zero had any other writer existed. The residual uncertainty is recorded above as
`[INFERRED]` rather than hidden. A `090` run is still worth its cost before anyone acts on candidate
1, because that one changes every selector score at once.

## Related

- `bugs_closed/060` — the same column name and the same failure on `agent_definitions`; fixed
  2026-07-26. The precedent, and the reason this one is harder to see.
- `bugs_open/351` — where this was found, and the decision it already influenced (the declined
  `section_type` backfill). Its council round `7b662d65` is where the objection came from.
- `docs024_key_docs_latest/bugfix_351_section_template_predicate/NOTES_351_section_template_predicate.md`
  — the working record, under "two findings the bug file does not contain".
- register **CLC-029** / migration `581` — the birth guard that stops new components arriving
  `section_type`-less, i.e. stops the *population* of permanently-uncountable rows growing.
