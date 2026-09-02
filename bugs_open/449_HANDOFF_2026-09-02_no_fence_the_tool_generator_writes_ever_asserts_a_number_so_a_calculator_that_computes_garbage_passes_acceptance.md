# BUG 449 — no acceptance fence the tool-generator writes ever asserts a NUMBER, so a calculator that computes garbage passes Tier 4

**Filed 2026-09-02** by the `mortgagecalculator_couk_adoption` lane, answering the owner's "verify
the tools" — the answer being that where the platform's verification runs at all, it is not checking
what the tools are for. **Status: OPEN. Nothing changed, nothing dispatched.**

---

## 1. The claim

The acceptance runner implements a check type, `computed_values`, that drives a tool with fixed
inputs and asserts the **exact text** of each output (`run_checks_action.go:708`, `:809`). It works —
this site's `simple` calculator passes four of them. **No agent that authors fences has ever been
told it exists**, so no generated fence uses it, and the generated fences that do exist check that
the page loads, logs no console errors, fits a phone, and produces *an element* after a click.
**A calculator that renders a confidently wrong figure passes every one of them.**

## 2. Measured 2026-09-02, live DB

| fences by author | count | assert **no expected value at all** | use `computed_values` |
|---|---|---|---|
| `tool-generator` | **170** | **107 (63%)** | **0** |
| `operator:bugfix224-session` | 16 | 0 | 16 |
| `operator:mortgagecalculator-lane-a4` | 8 | 0 | 8 |
| `operator:staged_component_build` | 8 | 0 | **6 — and 6 also carry `interaction`** |
| `webdesign_couk_thread` | 14 | 4 | 0 |

Two things that keep this honest:

- **The claim is NOT "generated fences never assert anything".** 63 of the 170 use
  `interaction.expect.text_matches`, and some of those patterns are real values (`\$1000\.00`,
  `40.0%`, `Checkout for 121`). ⚠ **My first version of this finding said "zero assert a computed
  value" and that was too strong** — it was true of the check *type* and false of the behaviour.
  The accurate figure is the middle column: **107 of 170 assert no expected value anywhere.**
- **`operator:staged_component_build` is the existence proof.** 6 of its 8 fences carry
  `computed_values` **and** `interaction` — correctness and health in one fence. So this is not a
  limitation of the format or the runner; it is a gap in what the authors are told.

**The two check families are otherwise disjoint**, which is the sharp version of the problem:

| family | asserts | blind to |
|---|---|---|
| `computed_values` (operator-authored) | the numbers are right | whether the page loads, errors, or fits a phone |
| `interaction` + `no_console_errors` + `no_horizontal_overflow` + `page_status_ok` + `selector_exists` (generated) | the tool is alive and responds | **whether any number it prints is correct** |

## 3. The cause, and it is one place

```sql
SELECT type, (default_config::text LIKE '%computed_values%') AS knows_computed_values
  FROM agent_definitions
 WHERE default_config::text LIKE '%```criteria%'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

| agent | knows `interaction` | knows `selector_exists` | knows `no_console_errors` | **knows `computed_values`** |
|---|---|---|---|---|
| `tool-generator` | yes | yes | yes | **NO** |
| `experience-planner` | yes | yes | no | **NO** |

Both fence-authoring agents were taught the liveness vocabulary and neither was taught the
correctness one. The 0-of-170 is not a modelling failure or a hard case — **the type is absent from
the prompt**, so it is never a candidate.

## 4. What this means for a "PASSED" verdict — read before quoting one

A Tier-4 PASS is only as strong as its fence, and the two families make opposite promises:

- `tool-overpayment-priority` and `tool-rate-scenarios` (this site) are **PASSING**. Their fences
  carry no value assertion of any kind. The verdict means *the page loads and something appears when
  you click* — **it says nothing about the arithmetic.**
- `simple` (this site) is **PASSING** on four `computed_values` checks and nothing else — no boot
  check, no console check, no status check, and every check pinned to `profiles: ["desktop"]`, so
  its doc-level `["desktop","mobile"]` produces a mobile pass in which **all four checks are
  skipped**. The verdict means *four sums are right on desktop.*

**Not one tool on this site is verified for both correctness and health**, and the same will be true
of any site whose fences came from `tool-generator`.

## 5. Fix candidates

1. **Teach both authoring agents the type (recommended, and small).** Add `computed_values` to the
   criteria vocabulary in `tool-generator` and `experience-planner`, with the instruction that a
   tool which computes anything must assert at least one worked example. `staged_component_build`'s
   6 fences are the template to copy.
   ⚠ **The generator must not invent expected values.** The RUNBOOK §14 rule for this site's own
   fences applies fleet-wide: *an emitted value is not an expected value*. A generated
   `computed_values` check whose expectation is whatever the tool printed at birth pins the bug
   along with the behaviour — `run_checks_action.go:775-781` says so in the code that does it. The
   generator can propose the inputs; the expectation needs a source that is not the tool.
2. **Report the gap rather than assume it.** A `needs_criteria`-style note when a fence for a
   computing tool carries no value assertion. Cheap, and it makes the 107 visible without anyone
   having to re-run this query.
3. **Backfill the 107.** Real work, tool by tool, and it needs (1) first or the next generated fence
   re-creates the gap.

## 6. How to verify a fix

- **Induce the red, which is the whole point here:** take a passing tool, change one constant in its
  JS so it computes a wrong figure, re-run acceptance. **Today it still passes.** After the fix it
  must fail. A fix that only adds checks which pass is indistinguishable from no fix.
- Re-run §2's query: `assert_no_value` for `tool-generator` must fall for **newly created** fences.
  Compare by `created_at` window, not by total — the 107 existing ones do not change themselves.

## 7. Related

- `bugs_open/441` — the fences that DO assert selectors are naming pre-conversion ids on
  instance-scoped tools. 441 is "the check cannot find the element"; 449 is "the check never asked
  whether the number was right". Independent, and both must be fixed for a PASS to mean much.
- `bugs_open/126` — gated-tool acceptance / `no_auto_fix`.
- RUNBOOK §14 of `mortgagecalculator_couk_adoption` — the emitted-vs-expected discipline, and
  `verify_criteria.py`, which re-derives every pinned value from a non-page source at three labelled
  strengths. That is the machinery candidate 1 needs and it already exists.
