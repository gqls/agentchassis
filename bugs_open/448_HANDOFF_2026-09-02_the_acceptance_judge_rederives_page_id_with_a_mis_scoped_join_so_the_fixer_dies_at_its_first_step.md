# BUG 448 — the acceptance judge RE-DERIVES `page_id` with a mis-scoped join instead of reading the one it was handed, so every `improve_tool` it files for a multi-row tool function dies at the fixer's first step

**Filed 2026-09-02** by the `mortgagecalculator_couk_adoption` lane, found while working the owner's
"verify the tools". **Status: OPEN. Nothing changed, nothing dispatched.**

Distinct from — but downstream of — `bugs_open/441` (stale acceptance selectors). 441 makes the
acceptance fail spuriously; **448 is why the repair it queues cannot even start.** Two tools on this
site are stuck behind both at once.

---

## 1. The symptom, observed

`site_work_items` `1d07d778` (`tool-deposit-tracker`) and `cc1f9dd2` (`tool-remortgage-savings`),
both `status='failed'`, `attempt_count=3` of `max_attempts=3` — retried to exhaustion 2026-08-27:

```
error: workflow failed: step load_tool failed: failed to execute action query_database:
       query param path 'input_data.spec.page_id' resolved to nil
```

`spec` keys on both rows, read live: `check`, `issue`, `screenshots`, `component_id`,
`failing_checks`, `acceptance_test`, `failing_instances`, `original_pipeline`. **No `page_id`.**

## 2. The mechanism — the reliable path exists, 300 lines below, and the wrong one is used

`platform/orchestration/actions/tool_acceptance_actions.go` files acceptance failures by two routes.

**The route for PORTED tools — a human queue with no automated fixer —
`routePortedAcceptanceFailure`, :1213-1218 — reads the value it was handed:**

```go
if pageID := datahelpers.ExtractNestedFieldString(params.CollectedData, "input_data.spec.page_id"); pageID != "" {
	spec["page_id"] = pageID
}
```

That is sound: the incoming `acceptance_run` item carries `page_id` in its own spec — the file says
so at :1039 (*"The run item's own spec names the instance (`check_tool_acceptance_due` writes
`component_id` and `page_id`)"*).

**The route for REAL tools — the one whose item goes to `tool-improver` —
`JudgeAcceptanceResultsAction`, :867-874 — throws that away and re-derives it:**

```sql
SELECT cc.id::text, COALESCE(p.id::text, '')
FROM content_components cc
LEFT JOIN page_components pc ON pc.component_id = cc.id
LEFT JOIN pages p ON p.id = pc.page_id AND p.site_id = $2::uuid
WHERE cc.function = $1 AND cc.is_active
LIMIT 1
```

Two defects in six lines:

1. **The site filter is on the LEFT JOIN, not in the WHERE.** So a `content_components` row for this
   function that is placed on a DIFFERENT site still satisfies the query — `p.id` simply comes back
   NULL, and `COALESCE` turns that into `''`.
2. **`LIMIT 1` with no `ORDER BY`.** Which of several rows sharing a `function` is returned is
   unspecified, so the outcome is not even stable between runs.

Then, :994-996, the write is conditional:

```go
if pageID != "" { spec["page_id"] = pageID }
```

— and the `INSERT` at :998-1005 lists
`site_id, source, pipeline, item_type, severity, summary, priority, handler_agent, status,
created_by, spec, item_key, batch_id`. **Neither `page_id` NOR `component_id` is in the column
list**, so there is no second store to fall back on: when the derivation misses, the value is
nowhere on the item, and `tool-improver`'s `load_tool` step — which reads
`input_data.spec.page_id` — has nothing to resolve.

**Why it bit these two tools specifically:** `tool-deposit-tracker` has **two** active
`content_components` rows under one function (measured 2026-09-02, live DB):
`tool-deposit-tracker-mortgagecalculator-co-uk-loanandmortgagecalculator-co-uk` and
`tool-deposit-tracker-mortgagecalculator-co-uk`. A function with rows placed on more than one site
is exactly the shape the mis-scoped join mishandles.

## 3. Verification status — READ THIS

**The `090` diagnosis loop was NOT run for this bug, and the reason is not judgement:** the
kubeconfig token expired mid-session (fleet-wide `Unauthorized`, the known 3-day expiry — the owner
refreshes it), so no cluster dispatch was possible. Per the owner ruling of 2026-07-31 this file
states the substitute plainly:

- **First-hand, and not affected by the outage:** the two failing rows, their exact `error` string
  and their full `spec` key list were read from the live DB **earlier in the same session, before
  the token expired**; the code in §2 is quoted from the working tree at HEAD; the two-row
  `content_components` fact for `tool-deposit-tracker` was likewise measured live.
- ~~**NOT yet measured, and marked so — do not quote a size for this bug:** `[UNMEASURED]` how many
  `improve_tool` rows fleet-wide are `failed` with this error, and `[UNMEASURED]` how many tool
  functions have active rows on more than one site.~~ **MEASURED 2026-09-02, once the token was
  refreshed — see §5, which now carries the numbers.**

This is close to the self-evidencing case the CLAUDE.md diagnosis rule exempts — the error message
names the exact path, the only writer of that path is six lines away and conditional, and the
condition's input is a query that demonstrably returns empty for a real shape — but it is filed as
UNVERIFIED-AT-SCALE rather than dressed up as complete.

## 4. Fix candidates, ordered by what closes the door

1. **Read the handed-down value first (recommended).** In `JudgeAcceptanceResultsAction`, take
   `page_id` (and `page_name`) from `input_data.spec.*` exactly as `routePortedAcceptanceFailure`
   already does, and use the query only as a fallback. It is the same file, the same params, and the
   pattern is already written and in use — this is deleting a divergence, not inventing a mechanism.
2. **Fix the query as well, not instead.** Move `p.site_id = $2` into a `WHERE` with an inner join
   (or `AND p.id IS NOT NULL`), and give `LIMIT 1` an `ORDER BY`. ⚠ **Do this even after (1)** — the
   same query also yields `componentID`, and a non-deterministic `LIMIT 1` there picks which
   component a fixer will rewrite. That is `bugs_closed/285`'s territory (*tool-improver rewrites a
   shared template from a single tool finding*) and deserves its own look.
3. **Make the item carry the columns.** `site_work_items` HAS `page_id` and `component_id` columns
   and this INSERT populates neither. Filling them gives every downstream reader a second store and
   is what `bugs_closed/154` established the framework can map from. Cheapest durable belt-and-braces.

**Do not** fix this by making `load_tool` tolerate a missing `page_id` — a fixer that proceeds
without knowing which page it is repairing is `bugs_closed/285` waiting to happen.

## 5. Blast radius — MEASURED 2026-09-02

| measure | value |
|---|---|
| `improve_tool` rows `failed` with this exact error | **16**, across **12 sites**, 2026-08-26 → 2026-09-02 |
| **all** `failed` `improve_tool` rows fleet-wide (the control) | **26**, across 14 sites |
| **so this defect is** | **62% of every failed `improve_tool` on the estate** |
| tool functions placed on more than one site (the population the join mishandles) | **42** |

The control is what makes the 16 mean something: failure is not the normal state of `improve_tool`,
and this one error accounts for nearly two thirds of it. The window opening on 2026-08-26 is worth a
look by whoever fixes this — it is one day after the Tier-4 due-sweep widened, so the defect is
plausibly older than the evidence and simply had no traffic before then. **`[INFERRED]` — I have not
checked what changed on 08-26.**

### The queries, for re-running



```sql
-- (a) fleet-wide blast radius of the observed failure
SELECT count(*) AS failed_items, count(DISTINCT site_id) AS sites
  FROM site_work_items
 WHERE item_type = 'improve_tool' AND status = 'failed'
   AND error LIKE '%input_data.spec.page_id%resolved to nil%';

-- (b) the population the mis-scoped join can mishandle
SELECT count(*) FROM (
  SELECT cc.function
    FROM content_components cc
    JOIN page_components pc ON pc.component_id = cc.id
    JOIN pages p ON p.id = pc.page_id
   WHERE cc.is_active AND cc.component_level = 'tool'
   GROUP BY cc.function HAVING count(DISTINCT p.site_id) > 1) q;
```

## 6. How to verify a fix

- Re-file one of the two failed items (or a fresh Tier-4 failure on `tool-deposit-tracker`) and
  watch `load_tool` clear. ⚠ **Check `collected_data->'__step_error'` is empty, not just that the
  orchestration says COMPLETED** — the `bugs_open/099` landmine: a failed step hides under
  COMPLETED with `error` NULL.
- **Induce the red:** a genuinely page-less function must still fail loudly rather than silently
  filing an unusable item. A fix that only makes the happy path work is indistinguishable from one
  that swallowed the error.
- ⚠ **Both of this site's cases are ALSO blocked by `bugs_open/441`** (their acceptance failed on
  stale selectors). Fixing 448 alone will let the fixer start and then have it "repair" a tool that
  is not broken. **Fix 441 first, or verify 448 on a tool whose acceptance failure is genuine.**

## 7. Related

- `bugs_closed/154` — `tool-improver` dying at `load_tool` on `input_data.component_id resolved to
  nil`. Same step, same shape, different field and different cause (there: the framework dropped a
  value between the column and spec stores; here: the producer never derives it). **154's fix does
  not cover this**, and the recurrence is the argument for candidate 3.
- `bugs_closed/285` — tool-improver rewriting a shared template from a single finding; the
  unordered `LIMIT 1` in §2 is a live path back into it.
- `bugs_open/441` — the stale-selector defect that generated these two failures in the first place.
