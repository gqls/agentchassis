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

## Inducing the floor live (site A recipe — both branches, 2026-07-31)

**A green run proves nothing: the floor is inert on healthy input by design.** The
only proof is to induce the fault, watch the refusal, confirm nothing was deleted,
clear the induction, and watch a normal run pass.

### Prove the code is actually in the pod FIRST

```bash
kubectl -n ai-persona-system exec <pod> -- sh -c '
  strings /app/agent-chassis | grep -c "returned too few sections to replace what is stored"   # mine
  strings /app/agent-chassis | grep -c "wanted to %s locked section"'                          # control
```

Both replicas, one exec each, marker AND a control invariant under your own diff.

**Gotcha — a fix with no string literal cannot be pod-grepped.** The
`recurrenceExpected` fix was a struct field and a control-flow change: nothing
lands in rodata. Date it instead against a neighbouring commit that DID add a
string — if a literal from a commit made *after* yours is present, and the build
comes from committed HEAD, yours is an ancestor and is in.

### Induce by inflating the PLAN, not by adding rows

```sql
-- save the exact baseline first; you will restore it verbatim
SELECT sections::text FROM pages WHERE id='<page>';
UPDATE pages SET sections = sections || '["i1","i2",...]'::jsonb WHERE id='<page>';
```

**Gotcha — the obvious induction does not work.** Adding synthetic
`page_components` rows to inflate the *stored* side fails, because
`rerender_page_sections` loads ALL rows for the page and regenerates from them:
the synthetic rows inflate the numerator too and the ratio stays 1.0. Inflating
`pages.sections` is both effective and safer — one jsonb column, no content.

### Getting the trigger to actually REACH save_page_sections

Four gates sit in front of it. All four cost a firing:

1. `input_data.spec.reason` must be `image_landed`, `section_data_resolved` or
   `cta_links_stale` — otherwise `check_rerender_mode` routes to `render_page` and
   the section path never runs.
2. `input_data.spec.page_name` must be set — page-rerender's `save_sections`
   config reads `page_name_field: input_data.spec.page_name`, and without it the
   action returns `{"skipped": true, "reason": "no page name"}` **before every
   guard**, which looks exactly like a pass.
3. The page must not escalate. `rerender_page_sections` escalates the whole page
   if any section's `content_data` is missing a schema-required `source:"llm"`
   field — **not** merely if it is NULL. Find a page that passes by mirroring
   `missingRequiredLLMFields` in SQL:

```sql
WITH fld AS (
  SELECT pc.page_id, f.key AS field, f.value->>'source' AS src,
         (f.value->>'required')::boolean AS req, pc.content_data -> f.key AS val
  FROM page_components pc JOIN content_components c ON c.id = pc.component_id,
  LATERAL jsonb_each(COALESCE(c.input_schema->'fields', c.input_schema)) AS f
  WHERE jsonb_typeof(COALESCE(c.input_schema->'fields', c.input_schema))='object')
SELECT DISTINCT page_id FROM fld
WHERE src='llm' AND req IS TRUE
  AND (val IS NULL OR val='null'::jsonb OR val::text IN ('""','[]','{}'));  -- EXCLUDE these
```

4. `rebuild_policy='owned'` pages are refused ~370 lines earlier.

### Publish so the message actually goes

Payload in the container COMMAND with a `PUBLISH_OK` marker — `kubectl run -i |
kcat -P` loses ~4 of 5 at exit 0:

```bash
kubectl -n kafka run "kcat-$(date +%s)-$RANDOM" --rm --restart=Never \
  --image=edenhill/kcat:1.7.1 --attach=true --quiet --command -- sh -c \
  "printf '%s' '<JSON>' | kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
   -t system.agent.generic.requests -H correlation_id=<uuid> ... && echo PUBLISH_OK"
```

### Read the result

```sql
SELECT status, current_step FROM orchestration_states WHERE orchestration_id='<orch>';
SELECT jsonb_pretty(collected_data->'save_sections') FROM orchestration_states WHERE orchestration_id='<orch>';
```

Refusal → `FAILED @ save_sections`, and the reason is in the chassis log
(`REFUSED for page`), NOT in `__step_error`. Pass → `completeness_status: passed`
with both cohorts in `completeness_cohorts`.

**Then prove nothing was deleted** — compare md5s to the baseline you took:

```sql
SELECT slot_name, position, md5(COALESCE(content_data::text,'')) AS cd_md5, updated_at
FROM page_components WHERE page_id='<page>' ORDER BY position;
```

### Clean up, and check you did

```sql
UPDATE pages SET sections='<exact baseline json>'::jsonb WHERE id='<page>';
SELECT count(*) FROM pages WHERE sections::text ~ '"i1"|"induced-';  -- must be 0
```

**Gotcha:** restore the plan on EVERY page you touched, including ones you
abandoned mid-way. Two of my four attempts were dead ends on other pages and both
were left inflated until I swept for the marker.

## R-B1 — replay the nav membership rule in SQL, to size a cohort's false positives

The one query that decided site B's cohorts. It mirrors `classifyPagesForNav` and
answers "what would a rebuild produce, per site and per group, against what is
stored". **Gotchas, both of which bit:** (a) legal pages bypass the
`in_header`/`in_footer` check entirely, so a flag-only denominator matches only
14 of 16 sites — include the legal branch and it matches 16 of 16; (b) join the
items to the GROUP (`i.group_id = g.id`), never both children off `sites`, or you
get items × groups.

```sql
WITH p AS (
  SELECT pg.site_id, s.domain, pg.name, pg.url, pg.page_type,
         COALESCE(pg.in_header,false) AS ih, COALESCE(pg.in_footer,false) AS if_, lower(pg.name) AS nm
  FROM pages pg JOIN sites s ON s.id=pg.site_id
  WHERE pg.status IN ('active','deployed','pending')      -- == navPageScopeSQL
), c AS (
  SELECT *, (nm IN ('404','sitemap','robots')) AS is_system,
    (nm LIKE 'privacy%' OR nm LIKE 'terms%' OR nm LIKE 'cookie%'
      OR nm LIKE 'disclaimer%' OR nm LIKE 'legal%') AS is_legal,
    (page_type IN ('blog-post','tool','entity-page')
      OR ((lower(url) LIKE '/tools/%' OR lower(url) LIKE '/blog/%' OR lower(url) LIKE '/guides/%'
         OR lower(url) LIKE '/articles/%' OR lower(url) LIKE '/case-studies/%' OR lower(url) LIKE '/news/%'
         OR lower(url) LIKE '/resources/%' OR lower(url) LIKE '/insights/%')
        AND page_type NOT IN ('blog-index','entity-directory','section-index','news-index'))) AS never_primary
  FROM p
), exp AS (
  SELECT site_id, domain,
    CASE WHEN is_system THEN NULL WHEN is_legal THEN 'legal'
         WHEN never_primary THEN (CASE WHEN ih OR if_ THEN 'utility' END)
         WHEN NOT ih THEN (CASE WHEN if_ THEN 'utility' END)
         ELSE 'primary' END AS grp
  FROM c
), e AS (SELECT site_id, domain, grp, count(*) AS expected FROM exp WHERE grp IS NOT NULL GROUP BY 1,2,3),
   st AS (SELECT g.site_id, g.group_key, count(i.id) AS stored
          FROM site_nav_groups g LEFT JOIN site_nav_items i ON i.group_id = g.id   -- NOT off sites
          GROUP BY 1,2)
SELECT COALESCE(e.domain,(SELECT domain FROM sites WHERE id=st.site_id)) AS domain,
       COALESCE(e.grp, st.group_key) AS grp,
       COALESCE(e.expected,0) AS expected, COALESCE(st.stored,0) AS stored,
       CASE WHEN COALESCE(st.stored,0)=0 THEN 'n/a (empty)'
            WHEN COALESCE(e.expected,0)::float/st.stored < 0.5 THEN '*** WOULD REFUSE ***' ELSE 'ok' END AS verdict
FROM e FULL OUTER JOIN st ON st.site_id=e.site_id AND st.group_key=e.grp
ORDER BY verdict DESC, domain, grp;
```

**How to read it.** Per-site totals matching stored exactly (they did, 16 of 16)
means the whole-site cohort has no false positives today. Any `WOULD REFUSE` row
at group level is the argument AGAINST a per-group cohort, not for one — a
re-homed page produces exactly that signature. Drop the `FULL OUTER JOIN` half
and you lose the stored-only groups, which is where the finding was.

**Re-run it before trusting the 0-of-16 figure** — it is a snapshot, and
`classifyPagesForNav` is actively edited by the `bugfix_149_nav_membership` lane.
If the classifier's membership rule changes, this replay must change with it or it
silently measures the old rule.

## R-B2 — prove a guard FIRES, without breaking the shared tree

A green test run says nothing about a guard that is inert by design. Break it in a
sandbox built from committed `HEAD` plus your working files, never in the tree —
another session may commit your broken file.

```bash
SB=<scratchpad>/negctl
rm -rf "$SB" && mkdir -p "$SB"
git rev-parse --short HEAD > "$SB/.head"     # record what you tested against
git archive HEAD | tar -x -C "$SB"
for f in prune_floor.go nav_prune_floor.go nav_prune_floor_test.go \
         link_registry_prune_floor.go populate_nav_tables_action.go site_db_actions.go; do
  cp platform/orchestration/actions/$f "$SB/platform/orchestration/actions/$f"
done
cd "$SB" && go test ./platform/orchestration/actions/ -count=1     # baseline: green
# then neuter ONE thing and re-run — e.g. evaluatePruneFloor(floor, nil)
```

The four controls run for B and C, and what each must produce:

| break | expected |
|---|---|
| `evaluatePruneFloor(floor, nil)` in `nav_prune_floor.go` | exactly the 4 nav refusal tests fail; **no allow test fails** |
| add a `{Label: "nav group: tools", Confirmed: 0, Stored: 1}` cohort | `TestNavFloorAllowsAPageReHomedBetweenGroups` fails |
| `evaluatePruneFloor(floor, nil)` in `link_registry_prune_floor.go` | its refusal test fails |
| replace `navPageScopeSQL` in the loader with a literal | `TestLoadPagesForNavUsesTheSharedScopePredicate` fails |

**The second row is the one that matters** — it is what stops someone re-adding
the per-group cohort the bug file asked for. **"No allow test fails" in row one is
half the control**: a neutering that fails everything proves only that you broke
compilation.

## R-C1 — the query that decides C's partition, once `link_registry` is non-empty

C ships one unpartitioned cohort because there is no distribution to read
(0 rows all-history). Run this before adding a partition:

```sql
SELECT link_type, scope, count(*) AS rows, count(DISTINCT source_page_id) AS pages,
       round(avg(n),2) AS avg_rows_per_page
FROM (SELECT link_type, scope, source_page_id,
             count(*) OVER (PARTITION BY source_page_id, link_type) AS n
      FROM link_registry) x
GROUP BY 1,2 ORDER BY rows DESC;
```

**The decision rule:** if most `(page, link_type)` groups hold one row, a
per-`link_type` cohort is site A's rejected per-`slot_name` shape again — every
cohort is 1 stored, so any legitimate single-link removal scores 0% and refuses.
Do not add it.
