# RUNBOOK — bugs_open/165 completeness floors

Every query here had to be got right once. The gotcha is attached to each.

## Choosing cohorts for a new site (do this BEFORE writing any Go)

### 1. Is the candidate partition degenerate?

A per-class cohort is only meaningful if a class holds several rows. Check first —
this is what killed the per-`slot_name` shape the bug file proposed:

```sql
WITH s AS (SELECT page_id, slot_name, count(*) n FROM page_components GROUP BY 1,2)
SELECT n AS rows_per_slot, count(*) AS slot_groups FROM s GROUP BY 1 ORDER BY 1;
```

**Gotcha:** a partition where nearly every cell holds ONE row is not a partition —
it is a rule that refuses any single deletion. 998 of 1,009 here.

### 2. What would the floor have refused historically?

`page_component_history` snapshots the pre-overwrite state, so consecutive events
per page give a before→after series:

```sql
WITH ev AS (
  SELECT page_id, date_trunc('second', created_at) AS t, count(*) AS n
  FROM page_component_history WHERE source='save_page_sections_overwrite'
  GROUP BY 1,2),
pair AS (
  SELECT page_id, n AS before_n,
         lead(n) OVER (PARTITION BY page_id ORDER BY t) AS after_n
  FROM ev)
SELECT count(*) AS transitions,
       count(*) FILTER (WHERE after_n < before_n) AS shrank,
       count(*) FILTER (WHERE after_n::float/before_n < 0.5) AS would_refuse_at_0_50
FROM pair WHERE after_n IS NOT NULL;
```

**Gotcha — mark this [PROXY], do not quote it as the written count.** The snapshot
excludes rows with empty `rendered_html` and INCLUDES locked rows, and another
action may touch the page between two overwrites. It characterises shrinkage; it
does not reproduce what the writer supplied.

### 3. Simulate the cohort against every page it could reach

The one that matters, and the one I got wrong first:

```sql
WITH x AS (
  SELECT p.id, s.domain, p.url, COALESCE(p.rebuild_policy,'generic') AS policy,
     GREATEST(
       (CASE WHEN jsonb_typeof(p.sections)='array' THEN jsonb_array_length(p.sections) ELSE 0 END)
     - (CASE WHEN jsonb_typeof(p.suppressed_sections)='array' THEN jsonb_array_length(p.suppressed_sections) ELSE 0 END)
     - (SELECT count(*) FROM page_components pc WHERE pc.page_id=p.id
          AND NOT (pc.locked_at IS NULL OR (pc.lock_type='timed' AND pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at<NOW())))
     , 0) AS planned_writable,
     (SELECT count(*) FROM page_components pc WHERE pc.page_id=p.id
        AND (pc.locked_at IS NULL OR (pc.lock_type='timed' AND pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at<NOW()))) AS live_writable
  FROM pages p JOIN sites s ON s.id=p.site_id)
SELECT policy, count(*) FILTER (WHERE planned_writable>0 AND live_writable>0) AS scorable,
       count(*) FILTER (WHERE planned_writable>0 AND live_writable>0
                          AND live_writable::float/planned_writable < 0.5) AS trips
FROM x GROUP BY 1;
```

**Three gotchas, all of which bit:**

1. **Subtract the locked rows from the plan.** Without the `NOT (...)` subquery
   the denominator counts slots the save may not write, and a perfect rebuild of a
   curated page is refused. Before: 3 trips. After: 0 on reachable pages.
2. **Split by `rebuild_policy`.** `owned` pages are refused ~370 lines earlier
   (`save_page_sections_action.go:150`), so counting them inflates the
   false-positive rate with pages the code never reaches. Both remaining trips are
   `owned`.
3. **Do not identify a page by `url` alone.** `/index.html` exists once per site —
   a `WHERE url='/index.html'` returns 14 different pages and reads as duplicates.
   Join `sites` and select the domain, or key on `pages.id`.

### 4. Name the consumers whose guarantee you are changing

Owner ruling 2026-07-29 §3: measuring that nothing breaks is not the same as the
other pipeline's owners agreeing. Get the list right:

```sql
SELECT a.type,
       (length(a.default_config::text) - length(replace(a.default_config::text, '"action": "save_page_sections"', '')))
       / length('"action": "save_page_sections"') AS step_instances
FROM agent_definitions a
WHERE a.is_active AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND a.default_config::text LIKE '%"action": "save_page_sections"%'
ORDER BY 1;
```

**Gotcha — the step-level census UNDERCOUNTS, badly.** The obvious query,
`LATERAL jsonb_each(default_config->'workflow'->'steps') WHERE step->>'action_type'`,
returns **0** (the key is `action`, not `action_type`) and even corrected returns
**3 of 6**, because `pageflow-builder` and `page-rebuild` nest the step inside a
loop. That is the `bugs_open/086-087` landmine — "handler census counts step-level
ONLY; diff the population". Match the exact key text instead.

## Verifying the Go when the shared tree will not build

`go build ./platform/...` failing is usually another session's uncommitted work,
not yours (it was, twice for the 135 lane and once here —
`discovery_checks/check_empty_sections.go` had an undefined `datahelpers`).

```bash
S=<your scratchpad>/hf
rm -rf "$S"; mkdir -p "$S"
git archive HEAD go.mod go.sum platform internal pkg | tar -x -C "$S"
cp platform/orchestration/actions/<your files> "$S/platform/orchestration/actions/"
cd "$S" && go build ./platform/orchestration/actions/
```

**Gotcha:** archive only the Go trees. A full `git archive HEAD` is ~350MB+ and
`/tmp` is a shared 16G tmpfs that runs at 95%+ — a whole-repo extract fails
half-way with ENOSPC and leaves a partial tree. `go.mod go.sum platform internal
pkg` is **12MB** and is all the compiler needs.

## Mutation-testing a guard (the only proof a green test is worth anything)

Do it in the isolated tree above, never in the shared working tree — a broken file
sitting in the working tree can be swept into another session's `git add -A`.

```bash
ORIG=<scratchpad>/floor.orig
cp <real file> "$ORIG"; cp "$ORIG" "$S/platform/orchestration/actions/<file>"
for m in M1 M2 M3; do
  cp "$ORIG" "$F"          # <-- fresh baseline EVERY time
  ...apply mutation $m...
  go test ./platform/orchestration/actions/ -run '<the test that guards it>'
done
cp "$ORIG" "$F"
```

**Gotcha, and it silently invalidated a whole first pass:** if the restore path is
wrong, `cp` fails, mutations ACCUMULATE, and M3 is measured on top of M1+M2 — the
tests still fail, so it looks like it worked. Attribution is the entire point.
Check the restore actually happened (`grep -c` the mutated string), and re-run the
baseline at the end: it must come back green.
