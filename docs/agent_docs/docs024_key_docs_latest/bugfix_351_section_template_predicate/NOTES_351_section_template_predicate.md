# NOTES — `bugfix_351_section_template_predicate`

Running technical record. **Append-only, newest at the bottom.** Evidence, commands, what the
system actually said, and every misstep. The missteps are not an appendix — they are the point.

Created 2026-08-23, late: this lane ran from 2026-08-21 to 08-22 with only a `HANDOFF`, which is
not the standing five. The entries below therefore start at 08-23; the 08-21/22 record lives in
`bugs_open/351` itself and in the handoff, and is not restated here (CLAUDE.md: point at bugs,
do not fork a second account that drifts).

---

## 2026-08-23 — picking the lane back up: is the bug still real?

**It is half real, and the half that is fixed had not been recorded as fixed.** Working through it
in order.

### The fix had rolled and both docs still said "inert"

`bugs_open/351`'s "IMPLEMENTED" section and the handoff's §3 both said the Go change was inert
until the next chassis roll. It had rolled. Verified at the binary rather than at git or at a tag:

```sql
SELECT DISTINCT service, git_commit, max(last_seen_at) OVER (PARTITION BY service, git_commit)
FROM service_binary_capabilities WHERE kind='build' ORDER BY 1,3 DESC;
--  agent-chassis | f5eaabe3342a906b0392f3cb0d77a67765da6955 | 2026-08-23 17:40:25Z
```

```bash
git merge-base --is-ancestor 97c337371 f5eaabe33   # → yes
git merge-base --is-ancestor <a later commit> f5eaabe33   # → no   (the control)
```

The `build provenance` startup line was **not** in `--tail=3000` on either pod, exactly as
CLAUDE.md warns; the RFC_040 table is what answered it.

### MISSTEP 1 — I started to date a PAST event from a table with a two-hour retention window

A `needs_new_component:loans-credit-health-check` row was filed at **12:08:49Z**, and I wanted to
know whether the fix was live when it was. The capability table showed the older of its two commits
with `min(started_at) = 10:53Z` — before 12:08 — and that commit contains the fix. I was one step
from concluding *"the fix was live and the defect recurred anyway"*, which would have reopened a
bug that is in fact working.

What stopped it: `kubectl get rs` shows a **rollout at 11:51:18Z** whose pods appear **nowhere** in
the capability table. `RetentionWindow` is `2 hours` and ephemeral job pods are *designed* to age
out, so the table is a **window, not a history** — it cannot speak to 12:08 at all, whatever
`started_at` says. Filed as a `LANDMINES.md` entry (no symptom; the rows are all real and correctly
dated, which is what makes it dangerous) and referenced from the bug file.

The right resolution of the 12:08 row turned out to be simpler and is below: the same page bound the
incumbent at 14:23, so the item was superseded, not evidence of failure.

### MISSTEP 2 — a line grep counted a CATEGORY as a LEVEL, and the wrong number was plausible

Counting the exported corpus by a sentinel header `@@@BEGIN:<id>|<level>|<category>|<function>`:

```bash
grep -c '^@@@BEGIN:.*|section|'   # 148
grep -c '^@@@BEGIN:.*|tool|'      # 132   ← wrong, the answer is 129
grep -c '^@@@BEGIN:'              # 277
```

148 + 132 = 280 ≠ 277. Three section-level components have `category='tool'` and were counted
twice. Caught only because I had printed the total in the same breath. `awk -F'|' '$2=="section"'`
asks the question I meant and cannot make the mistake. Logged in `WRONG_CALLS.md` with the general
form: *a measurement whose wrong value sits in the same range as its right value cannot be checked
by reading it.*

### Re-calibration against the live corpus — the flip SET held

Isolated tree (`git archive HEAD`, **on disk, not `/tmp`** — another lane found that tmpfs at 100%
the same day), throwaway `zz_tmp_351_calib_test.go` inside `platform/orchestration/actions` because
the predicates are unexported, deleted with the tree:

```
read: section=148  tool=129  calculators=22     (all asserted non-zero — the vacuity guard)
sectionTemplateValid   rescued=22  regressed=0
endsCleanly flips = 2   SET = {3f946437 case-studies-grid, 6c41404d about-commercial-block}
calculators FAILING the live predicate = 0      calculators with unbalanced markup = 0
```

Mutation controls in the same harness, so the assertion could have failed: a real mid-tag cut, a cut
immediately after a complete mid-template action, and a bare `}}` suffix are all still refused;
nested trailing `{{end}}` wrappers are accepted.

**The corpus moved again** — 150/124 on 08-22, 148/129 on 08-23. Re-running was not ceremonial.

### DEMAND PROOF — the bug file's own closing condition, met

The condition was *"a site planning a calculator section that RESOLVES to a library incumbent with
no `needs_new_component` item filed at all"*.

| bound (UTC) | site | page | component | `section_type` |
|---|---|---|---|---|
| 13:57:41 | loanzy.uk | `tool-is-a-loan-right-for-me` | `loans-damage-checker` | NULL |
| 14:07:15 | loanzy.uk | `tool-eligibility-checker` | `loans-credit-health-check` | NULL |
| 14:23:29 | loanzy.uk | `tool-credit-health-check` | `loans-credit-health-check` | NULL |

All three incumbents were born on `loanandmortgagecalculator.co.uk`, so this is cross-site reuse of
the library — the thing that was impossible before. No `needs_new_component` was filed for any of
them. **Attribution checked rather than assumed:** `824e3309`'s `html_template` was last written
**2026-08-20**, so the data did not move under us; the code did.

---

## 2026-08-23 — two findings the bug file does not contain

### A. `usage_count` counts ONE of the two resolution paths, and the selector scores on it

This started as "why do all 22 incumbents read `usage_count = 0` when three are bound to live
pages?" and ended somewhere more general.

`IncrementUsageCount` has **exactly one non-test caller** as of **2026-08-23**
(`grep -rn "IncrementUsageCount" platform/ internal/ pkg/ cmd/`):
`plan_sections_action.go:1957`, inside `resolveSectionComponent` — which is **Path 2 only**, the
`section_type` selector. A component resolved by **Path 1** (direct `function`/`name` match,
`plan_sections_action.go:1258`) is bound to the page and **never counted**.

And `usage_count` is a scoring input on the path that does count it —
`component_selector.go:181` and `:235`, both queries:

```sql
+ LEAST(COALESCE(usage_count, 0)::float / 50.0, 1.0) * 0.1
```

with the file header calling it *"battle-tested components score higher, with diminishing returns"*.

`[MEASURED 2026-08-23]` over active, non-forked, `component_level='section'` rows:

| | count |
|---|---|
| have any `usage_count` at all | **12** of 149 |
| `usage_count = 0` **and** ≥1 live `page_components` binding | **96** of 149 |
| page bindings invisible to the scoring term | **1,802** |

So the "battle-tested" term sees 12 components' worth of history and is blind to 1,802 bindings. It
is not merely noisy — it is **systematically** biased towards whatever was resolved by
`section_type`, because that is the only thing it can see. A component's score reflects the *path*
it was reached by, not its merit.

**Why this matters for the residual and not just as trivia:** it is a direct argument against the
"just backfill `section_type`" candidate. A backfill would put the incumbents (`usage_count = 0`,
because Path 1 is all they have ever had) into Path 2 scoring **against their own diverted twins**,
which carry 1–4 — a number that measures nothing but which path resolved them. The selector would
prefer the site-suffixed twin over the generic incumbent, on evidence that is an artefact.

`[UNMEASURED]` whether the 0.1 weight is large enough to flip a real pairing. The other terms
(site-type relevance, page-type relevance, quality, specificity) may dominate. **That is the query
to run before anyone acts on this**, and it is not run yet.

### B. A `section_type = function` backfill appears to be a NO-OP for selection

Reasoned from the code, and the reason it is worth writing down is that the bug file's open question
assumes the opposite.

`plan_sections_action.go:1258–1300` runs the paths in a fixed order: Path 1 (`components[sectionName]`,
built by `loadComponentSchemas` → `loadSectionComponents`, which matches **`name` then `function`**,
each against the **raw and kebab-normalised** form, `v3_site_actions.go:4958`), then Path 2
(`resolveSectionComponent` → `SelectComponentByType`, `WHERE section_type = $1`), then Path 3.

If a backfill sets `section_type := function`, then the only key by which the incumbent becomes a
Path-2 candidate is a string that Path 1 **already resolves** — and Path 1 runs first. So the
incumbent can never actually be *reached* through the new key. The backfill adds a row to a query
whose answer is never consulted for that string.

**Consequence:** the ordering question the bug file left open ("which wins depends on path order")
may be the wrong question. The real one is whether these components should answer to a **more
generic vocabulary term** than their own function name — which is a taxonomy decision, not a
mechanical backfill.

`[INFERRED]` — this is read off the code, not demonstrated by running it. The disconfirming
experiment is stated in the plan; do not quote this paragraph as measured.

---

## 2026-08-23 (later) — putting a number on the damage: 25 of 30 filings were avoidable

Over the **whole history** of `needs_new_component` (first row 2026-08-05, last 2026-08-23),
`[MEASURED 2026-08-23]`:

| | count |
|---|---|
| items ever filed | **30** |
| whose `section_type` exactly matched a live component's `function` | **27** |
| …and that component carried `section_type IS NULL` | **25** |

Each of those 25 is a paid LLM generation for a section the library already owned and could have
resolved by name on Path 1 — the platform commissioning a second copy of its own work.

**The obvious way this measurement could have lied, and the control that rules it out.** The
matching component might have been created *by* the item it appears to indict, which would make the
join circular and the number meaningless. Re-run with `AND c.created_at < w.created_at`:

```
30 | 27 | 25      -- identical
```

Identical, so every match genuinely predated its item. (The twins do **not** contaminate this: their
`function` values are site-suffixed — `loans-credit-health-check-loancalculator-co-uk` — so they
cannot satisfy `lower(function) = section_type`.)

`[CAVEAT]` `is_active` is evaluated as of today, not as of the item's date, so a component that was
inactive when the item was filed and activated later would be counted wrongly. Not checked — it
would move the number by at most a couple either way and does not change the shape.

**This makes a disconfirmable prediction, which is the point of recording it rather than admiring
it:** if the predicate fix is doing what we think, the *rate* of `needs_new_component` filings whose
`section_type` matches a live function should fall to near zero from 2026-08-23 onward. It has not
been long enough to test. **Whoever next picks this up should run the query above windowed on
`created_at > '2026-08-23'` — a continued high rate refutes the fix, and that is the cheapest
available way to be wrong about it.**

It also sharpens finding B above: 27 of 30 filings named an **exact function match**, so Path 1 is
the route that actually matters for this population and Path 2's `section_type` key is close to
irrelevant to it. A `section_type := function` backfill would be adding a key to the path that was
never the bottleneck.
