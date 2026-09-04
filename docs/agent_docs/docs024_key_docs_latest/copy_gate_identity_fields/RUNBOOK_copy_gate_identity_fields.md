# RUNBOOK — copy gate: `name` fields and the identity/display split

Every command here had a gotcha attached when it was first got right. Change them HERE.

---

## 1. THE CENSUS — is `name` an identity or display copy?

The load-bearing measurement of this lane. Walks **every object at any depth**, not just top-level
arrays.

```sql
WITH items AS (
  SELECT it.value AS item
  FROM page_components pc,
       LATERAL jsonb_path_query(pc.content_data, '$.**') AS arr,
       LATERAL jsonb_array_elements(arr) AS it(value)
  WHERE pc.content_data IS NOT NULL
    AND jsonb_typeof(arr) = 'array'
    AND jsonb_typeof(it.value) = 'object'
    AND it.value ? 'name'
)
SELECT CASE
         WHEN NOT (item ? 'url')       THEN '1. no url key at all'
         WHEN item->>'url' IS NULL     THEN '2. url key, JSON null'
         WHEN btrim(item->>'url') = '' THEN '3. url key, EMPTY string'
         ELSE                               '4. url key, non-empty'
       END AS url_sibling_state,
       count(*) AS items,
       count(*) FILTER (WHERE (item->>'name') LIKE '% %') AS name_prose_shaped
FROM items GROUP BY 1 ORDER BY 1;
```

⚠ **`$.**` recursion is the whole point.** A first pass over top-level arrays only returned 795
where this returns 825 — the 30 missing items are nested deeper and are disproportionately the odd
shapes. The components lane and I got different numbers for exactly this reason and the recursive
walk reconciled them.

⚠ **Split the url state four ways, not two.** `item ? 'url'` is key PRESENCE, so a `"url": ""`
would sit in the identity bucket while behaving like display copy. It does not occur today — but
that is a measured fact, not a safe assumption, and it was the copy lane's caution that prompted
checking it.

⚠ **The population grows by ADDITION while you watch.** It read 898 → 904 → 908 within 40 minutes
on 2026-09-03. Date any figure you record.

## 2. THE LIVE DEFECT COUNT — how many `name` values carry a gate shape

```sql
WITH kv AS (
  SELECT pc.page_id, m[1] AS k, m[2] AS v
  FROM page_components pc,
       LATERAL regexp_matches(pc.content_data::text,
                              '"([A-Za-z0-9_]*[Nn]ame)"\s*:\s*"((?:[^"\\]|\\.)*)"', 'g') AS m
  WHERE pc.content_data IS NOT NULL
)
SELECT s.domain, count(*) AS name_tells, count(DISTINCT kv.page_id) AS pages
FROM kv JOIN pages p ON p.id = kv.page_id JOIN sites s ON s.id = p.site_id
WHERE kv.v ~ '(?i)[[:alnum:])"''’](,|—|–|-)[[:space:]]+(not|never)[[:space:]]+'
   OR kv.v ~* '\mrather than\M' OR kv.v ~* '\minstead of\M'
   OR kv.v ~* '\mnot (just|only|merely)\M'
GROUP BY 1 ORDER BY 2 DESC;
```

Baseline `[MEASURED 2026-09-03]`: **37** values / **15** domains / **23** pages, all `deployed`.
**This number should FALL as pages rebuild.** It will not fall all at once and it will not fall
from a rerender (see §5).

⚠ **Do NOT add `em_dash` to that predicate.** `em_dash` lives in the *neighbour* set
(`negationtells.go`), used only for `displaced_*` rejection checks — it is **not** one of the seven
gate shapes. Including it inflates the count by 52 and none of those are repair targets.

⚠ The regex-over-`::text` approach catches `company_name` etc. too. Filter `m[1] = 'name'` when you
want bare `name` only.

## 3. WILL A TRUNCATION SURVIVE THE JUDGE'S WORD FLOOR?

The question that produced the heading-floor ruling. Counts the words left after cutting at the
construction:

```sql
-- ... same kv CTE as §2 ...
), tells AS (
  SELECT regexp_replace(v, '(?i)[[:space:]]*[,—–-][[:space:]]+(not|never)[[:space:]].*$', '')
         AS surviving_clause
  FROM kv WHERE v ~ '(?i)[[:alnum:])"''’](,|—|–|-)[[:space:]]+(not|never)[[:space:]]+'
)
SELECT array_length(regexp_split_to_array(btrim(surviving_clause), '[[:space:]]+'), 1) AS words_left,
       count(*) AS headings
FROM tells GROUP BY 1 ORDER BY 1;
```

`[MEASURED 2026-09-03]`: 1 word ×1, 2 ×10, 3 ×6, 4 ×8, 5 ×6, 6 ×3, 7 ×2 — so a 5-word floor
refuses **25 of 36**.

## 4. IS THE GATE EVEN LIVE ON THIS AGENT?

```sql
SELECT type,
       (default_config::text ~ '"copy_gate_annotate":\s*true') AS annotate_on,
       (default_config::text ~ 'rewrite_negations')            AS has_repair_step
FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND type IN ('page-content-writer','page-rerender','section-editor');
```

`page-content-writer` is `t/t`. **`page-rerender` is `f/f`** — which is §5.

## 5. ⚠ A RERENDER DOES NOT REPAIR THIS. THE ONE-LINE CHECK.

```sql
SELECT default_config::text ~ 'rewrite_negations' FROM agent_definitions
WHERE type='page-rerender' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

`rerender_page_sections_action.go:3-9` says it re-renders stored `content_data` **"WITHOUT invoking
the content writer (no LLM)"**. The defect lives in `content_data.name`, so a rerender faithfully
re-renders it and reports success. Only a `page-content-writer` rebuild regenerates the value.
**I offered a rerender to the owner before running this query; he chose it. One query, and I had
not run it** (`WRONG_CALLS.md` 2026-09-04).

## 6. VERIFYING THE CHANGE — build, mutations, HEAD

```bash
go test ./platform/orchestration/datahelpers/
go test ./platform/orchestration/actions/ -run 'CardName|IdentityNameIsNever|PlanClassifies|HeadlineHit'

# HEAD + only your files — this is the one that matters on a shared tree
./scripts/verify-head-builds.sh --test \
  --with platform/orchestration/datahelpers/negation_content.go \
  --with platform/orchestration/datahelpers/negationtells.go \
  ./platform/orchestration/datahelpers/...
```

⚠ **`go test ./platform/orchestration/actions/` is RED and it is not you.** Two tests fail at plain
HEAD (`TestFindingCodeScanEveryWriteIsRegistered`, `TestTemplateExecutorsAreDeclared`, both about
`renderFailWorkItemMessage`). A third, `TestBudgetLadderPrefersTheMostSpecificLevel`, fails only in
the **working tree** because `llm_budget_ladder_test.go` is another session's **untracked** file.
**Run the no-overlay control before blaming yourself:**
`./scripts/verify-head-builds.sh --test ./platform/orchestration/actions/`

⚠ **`go build ./platform/...` was OOM-KILLED (exit 137)** on this box. Scope it to the package.

**Running a mutation** (the only thing that proves a test can fail): back the file up to the
scratchpad, patch with `python3`, run the single named test, restore with a `trap`. All 14 for this
change are listed in the two test-file headers. `git stash` is **FORBIDDEN** and mechanically
blocked — never reach for it to get a clean tree.

## 7. POST-ROLL: PROVE IT AT THE BINARY, NOT THE TAG

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
# MUST be present (the new capability):
kubectl -n ai-persona-system exec "$POD" -- grep -aq "identityContentField" /proc/1/exe && echo PRESENT
# CONTROL 1 — must ALSO be present (proves the probe works at all):
kubectl -n ai-persona-system exec "$POD" -- grep -aq "AcceptNegationRewrite" /proc/1/exe && echo "control ok"
# CONTROL 2 — must be ABSENT (proves the probe can say no):
kubectl -n ai-persona-system exec "$POD" -- grep -aq "identityContentFieldXYZZY" /proc/1/exe && echo "BAD: matches anything"
```

⚠ **Never `strings`** — absent from the debian-slim images, and behind the customary `2>/dev/null`
its failure is indistinguishable from "not stamped". ⚠ **Never a discovery grep for "some 40-hex
string"** — it matches Go's internal digit table and returns the same wrong answer on every
service. ⚠ **Always run both controls in the same breath.**

The `build provenance` log line is a STARTUP line and scrolls out of reach within minutes on a busy
service — an empty result there means "not in range", not "unstamped".

## 8. COUNCIL

```bash
# free admission test, nothing dispatched, no credits
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <file.json>
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <file.json>
```

⚠ **≤8 edits, enforced server-side.** This submission used exactly 8.
⚠ **Budget ~30 minutes, not 2** — the council itself takes 2–5 but the dispatch queues behind the
fleet. A missing orchestration row is latency; do not retry on that evidence.
Find the run by payload, not by the printed id:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '3e9e8ce8-fb9b-4f5b-a610-016b57427a27';
```
