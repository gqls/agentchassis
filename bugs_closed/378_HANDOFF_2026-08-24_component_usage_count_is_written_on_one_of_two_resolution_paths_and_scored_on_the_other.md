# 378 — `content_components.usage_count` is written on ONE of the two resolution paths and read as a quality signal on that same path, so a component's score records which route found it rather than whether it is any good

**Filed:** 2026-08-24 by the `bugs_open/351` lane, found while answering a council objection about
which path resolved a component. Not a symptom anyone reported — the numbers are simply wrong in a
direction that always favours the same kind of component.

**Severity:** latent, structural. Nothing errors, nothing is refused, no page is broken today. The
cost is a selection input that is silently biased, and a figure that reads as evidence in bug files
and hand queries when it is not.

~~**Status: OPEN, not started.** Diagnosis below is first-hand and complete; no code written.~~

> ## STATUS 2026-08-24 (updated post-roll) — **FIXED AND LIVE.** Read this block before the body.
>
> **Live on chassis `48f55f21834ac3e2d95aa43716f6e63e40ac12ee`** (pod started 18:55:21Z). Proven three
> ways, each with a control: ancestry (`5074367f7` IS an ancestor; a later commit correctly is NOT),
> the new SQL `count(DISTINCT p.site_id)` PRESENT in `/proc/1/exe`, and the old `/ 50.0` scoring term
> ABSENT with a positive control proving the grep is not blind. Council **`ca01b81a` APPROVED** round 1.
>
> ⚠ **NOT yet demand-proven, and that is why this file is still in `bugs_open/`.** `usage_count` has
> frozen — but `page_components` rows created since the roll = **0**, so nothing has exercised the path
> and the old code would have incremented nothing either. **The frozen counter is not evidence until a
> page build has run.** Recipe: `docs024_key_docs_latest/bugfix_378_usage_count_derived/HANDOFF_2026-08-24_continue_here.md` §4.
>
> **One correction to the block below:** it said the contract-row change was a risk needing post-roll
> verification. It has been verified and the result is the **opposite** of the concern. The store
> resolves what it overwrites by `function = NormaliseToKebab(section_type)`; measured over all 117
> section_types, the OLD ordering's prediction agreed with that in **88**, the NEW ordering agrees in
> **90**, and both changed types moved from disagree to agree. **The change reduced a pre-existing
> mismatch rather than creating one.** A separate finding falls out: **27 of 117** section_types still
> predict a contract the store would not enforce — pre-existing, unowned, and arguably its own bug.
>
> ## (original post-fix block, kept as written)
>
> Commit `5074367f7`, council **`ca01b81a`** (`Council-Submitted:`, **verdict not yet read**).
> Inert until the next chassis image is built and rolled, so the defect is still reproducible on
> the fleet and this file stays in `bugs_open/`. Lane:
> `docs/agent_docs/docs024_key_docs_latest/bugfix_378_usage_count_derived/`.
>
> **The body below is kept as written and is the record of what was believed when it was filed. Four
> of its structural claims are wrong or incomplete, and one of its ranked candidates is refuted.**
>
> 1. **THREE resolution paths, not two.** Path 0 (the page's stored `page_components.component_id`,
>    `bugs_open/204`) is tried *first*, before the name/function match. So the population least
>    likely to be counted is also the most settled one. The body's framing of "two ways" is wrong.
> 2. **THREE readers, not two — and the missed one matters more.**
>    `load_existing_component_action.go:170` orders by `usage_count DESC NULLS LAST` to pick the
>    **canonical contract row** — by its own comment, *"the row the store will overwrite and
>    enforce"*. That is a heavier decision than the 0.1 score, and the body does not mention it.
> 3. **The counter OVER-counts as well as under-counting, so candidate 2 is REFUTED.** The
>    increment fires inside `resolveSectionComponent` *before* `planSection` decides
>    ready/deferred/skipped and before any binding is written, and again on every re-plan — so it
>    counts *resolution attempts*, not usages. `[MEASURED 2026-08-24]` the column's two largest
>    values are both components with **zero** page bindings: `testimonials-modern` (created
>    2026-08-23, `usage_count=12`, no `page_components` row ever, checked including
>    `build_status='removed'`) and `bayesian-ranking-hero-tool_pre_037` (a retired backup copy, 20).
>    **"Call `IncrementUsageCount` on Path 1 too" would spread a definition that is already wrong in
>    both directions.**
> 4. **Both `[UNMEASURED]` items are now measured.**
>    - *Has the 0.1 term ever flipped a selection?* **No — 0 of 4,888** contested
>      `(section_type, site_type, page_type)` contexts, with a counterfactual control granting the
>      weakest candidate 50 uses that flips **52**, so the instrument had power. The cause is
>      mechanical: only **4** section_types have >1 candidate and every candidate in all four reads
>      `0`; the 12 counted components are all the sole candidate for their own type.
>    - *Do the other levels share the asymmetry?* **Worse.** `tool` is **0 of 115** — dead, exactly
>      `bugs_closed/060`'s case. `header`/`footer`/`site`/`element` all 0. Only `section` counts.
>    - The `[INFERRED]` "no third writer" **holds, and is now measured**: no DB trigger, function or
>      view touches the column; the birth INSERT at `store_generated_component_action.go:639` writes
>      the literal `0`.
>
> **What shipped, and the one decision a reviewer should go at.** The counter and its only call site
> are deleted; "how proven is this component" is derived from `page_components` (DISTINCT SITES,
> excluding `build_status='removed'`) in **one** named constant, `ComponentUsageSitesSQL`. That
> constant now orders the **contract row** (2 of 4 contested types move, both corrections:
> `about-hero`→`hero`, `archetype-taster-quiz`→`tool-archetype-taster-quiz`).
>
> **The scoring term is REMOVED, not repaired — this is candidate 3, not candidate 1.** Repairing it
> was built first and withdrawn on measurement: removing changes **0** of 4,888 winners, feeding it
> the corrected number changes **3,246** across 3 section_types. And a *working* usage term is a
> preferential-attachment loop — selected → count rises → scores higher → selected again — which is
> precisely what `bugs_open/107` ("every site gets the same homepage skeleton") is the standing
> complaint about, citing this very file. **A term that cannot be made accurate without making the
> estate more homogeneous is not worth keeping in the score.** The derived figure is still SELECTed
> and logged; nothing scores on it.
>
> ⚠ **The column `content_components.usage_count` still exists**, now written by nothing and read by
> nothing in Go. Dropping it is a follow-up migration once this code is live, so the code cannot roll
> back onto a missing column. **Until then it still reads as a maintained figure — do not quote it.**
>
> **Known contamination in the new substrate, stated not filtered:** `bugs_open/357`'s mis-bound rows
> (**22** as of 2026-08-24, all declaring `hero` while storing a tool page) become bindings. Under the
> DISTINCT SITES unit they collapse to **3 of hero's 27 sites**, so they are functionally inert — but
> 357's mint is still open (their phase 2 ships default-OFF), so treat it as a floor with a growth
> rate. Their phase 3 migration `578_HOLD` retypes the population and this self-corrects.

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

---

# §CLOSED 2026-08-25 — fixed, live, and DEMAND-PROVEN

**Live on chassis `635f2d32f5bbe3789867a978284c9c125d718eb0`** (pod started 2026-08-25 08:49:31Z);
first shipped on `48f55f21…` at 2026-08-24 18:55:21Z. Commit `5074367f7`. Council **`ca01b81a`
APPROVED** round 1.

The bar this file was held to for an extra day was **demand-proof** — the counter had frozen, but
nothing had been built, so the freeze was consistent with the fix and equally consistent with nothing
having happened. That gap is now closed.

## The evidence, each arm with its control

**1. The removed code is not in the binary** (proof of a removal is absence, and absence needs a
two-sided probe):
```
grep -aq 'count(DISTINCT p.site_id)'          /proc/1/exe   -> PRESENT   (the new derivation)
grep -aq 'usage_count, 0)::float / 50.0'      /proc/1/exe   -> ABSENT    (the old scoring term)
grep -aq 'component_selector: selected'       /proc/1/exe   -> PRESENT   (control: the probe is not blind)
git merge-base --is-ancestor 5074367f7 635f2d32f5bb…        -> YES
git merge-base --is-ancestor bba8a892d 635f2d32f5bb…        -> NO        (control: a later commit)
```

**2. There was real demand, `[MEASURED 2026-08-25]`** — this is the arm that was missing yesterday:

| signal, since the fix went live (2026-08-24 18:55:21Z) | count |
|---|---|
| `page_components` rows created | **403** across **125** distinct pages |
| `page-build-handler` orchestrations (runs the section loop) | **73** |
| `page-rerender` orchestrations | **880** |
| `component-creator` orchestrations (the ONLY consumer of the changed contract reader) | **17** |
| components actually born (creator succeeding end to end) | **5** |

**3. The counter did not move.** All 12 components carrying a value read **byte-identically** to the
pre-fix snapshot of 2026-08-24 13:30Z — `20, 19, 12, 7, 4, 2, 1×6`, same names, same order. Under the
old code, any Path-2 resolution across those 73 page builds would have incremented.

**4. Nothing broke.** `needs_new_component` filings ran at **5** since the fix against **6** in the
preceding 24 hours, and `needs_section_data` went **3 → 0**. A broken selector query would have
failed every Path-2 resolution and spiked both. It did not.

⚠ **One check I tried and correctly discarded:** grepping the chassis logs for selector errors
returned `0` — and so did the control asking whether those logs contain *any* plan/selector lines at
all. The work runs in dynamically-named pods (the build stamp itself came from
`agent-page-content-writer-…`), not the two pods labelled `app=agent-chassis`. **That zero was
vacuous and is recorded here so nobody quotes it as evidence.**

## What was actually wrong with the bug as filed

Four structural claims corrected (three paths not two; a third and heavier reader; the counter
over-counts as well as under-counts, which refutes its own candidate 2; both `[UNMEASURED]` items
measured). **And the ranked fix was wrong**: the file ranked "count at the binding" first. What
shipped is candidate **3** — *stop scoring on it* — because removing the term changes **0** of 4,888
contested selections while repairing it changes **3,246**, and an accurate "prefer what is already
used" term is the preferential-attachment loop `bugs_open/107` exists to complain about.

## The one concern that inverted

The council's `bug_historian` seat objected (medium) that repointing the contract-row `ORDER BY`
could silently re-shape the enforced schema. Right family, backwards direction: the store resolves
what it overwrites by `function = NormaliseToKebab(section_type)`, and measured across all **117**
section_types the OLD ordering agreed with that in **88**, the NEW ordering agrees in **90** — both
changed types moving from disagree to agree. **The change reduced a pre-existing mismatch.**

## Residuals, carried out of this bug rather than buried in it

- **27 of 117 section_types still predict a contract the store would not enforce.** Pre-existing, NOT
  caused by this change, and nobody owns it. Filed separately — see `bugs_open/` for the
  contract-prediction mismatch.
- **`content_components.usage_count` still exists**, written by nothing and read by nothing in Go.
  Deliberately left so the code could not roll back onto a missing column; that reason has now
  expired. A `COMMENT ON COLUMN` (or drop) migration is owed and is council-scope in its own right.
  ⚠ **Until it runs, the column still reads as a maintained figure — do not quote it.**
- `bugs_open/357`'s 22 mis-bound rows remain a stated contamination of the new substrate: under the
  DISTINCT SITES unit they collapse to **3 of hero's 27 sites**, and their phase 3 migration retypes
  them, at which point this self-corrects.

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_378_usage_count_derived/`
