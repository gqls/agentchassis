# RUNBOOK — `bugs_open/469`, the section-source drift closer

Every query and command that was hard to get right, with its gotcha attached. When one
changes, change it HERE.

---

## 1. Is there live section-source drift right now? (tier 1)

⚠ **Always print the demand control.** A bare `0` here has two causes with opposite
meanings — no drift, or a join that matched nothing. My first cut printed `0|0|0` and was
unreadable.

```sql
WITH tier1 AS (
  SELECT sp.site_id, sps.page_name,
         jsonb_agg(sps.component_name ORDER BY sps.ordering) AS auth
  FROM site_plan_sections sps JOIN site_plans sp ON sp.id = sps.plan_id
  WHERE sp.is_current GROUP BY 1,2),
cache AS (
  SELECT p.site_id, p.name AS page_name, p.sections AS cache FROM pages p
  WHERE COALESCE(p.status,'') <> 'deleted'
    AND jsonb_typeof(p.sections) = 'array' AND jsonb_array_length(p.sections) > 0)
SELECT count(*) AS compared,                       -- THE CONTROL. Never omit.
       count(*) FILTER (WHERE t.auth = c.cache) AS agree
FROM cache c JOIN tier1 t ON t.site_id = c.site_id AND t.page_name = c.page_name;
```

**Direction of the error, so you can read the result honestly:** this does NOT apply the
loader's locked-row merge (`datahelpers.MergeLockedPageSlots`), which the real check applies
to both sides. The merge only ever inserts locked rows into BOTH lists, so it can turn a raw
mismatch into agreement and never the reverse. **A raw zero therefore implies a merged
zero**; a raw non-zero may be a locked-row false positive and must be checked per page.

## 2. Same, for tier 2 (the `site_specs.site_plan` aspect)

⚠ Two typed guards are load-bearing. Without `jsonb_typeof(pg) = 'object'` and
`jsonb_typeof(pg->'sections') = 'array'` this dies with
`ERROR: cannot extract elements from a scalar` — some sites' aspect `pages` arrays carry
non-object entries.

```sql
WITH tier1 AS (
  SELECT sp.site_id, sps.page_name FROM site_plan_sections sps
  JOIN site_plans sp ON sp.id = sps.plan_id WHERE sp.is_current GROUP BY 1,2),
aspect AS (
  SELECT ss.site_id, pg->>'name' AS page_name,
         (SELECT jsonb_agg(s) FROM jsonb_array_elements_text(pg->'sections') s) AS auth
  FROM site_specs ss, jsonb_array_elements(ss.data->'pages') pg
  WHERE ss.aspect = 'site_plan' AND ss.is_current
    AND jsonb_typeof(ss.data->'pages') = 'array'
    AND jsonb_typeof(pg) = 'object' AND jsonb_typeof(pg->'sections') = 'array'),
cache AS (
  SELECT p.site_id, p.name AS page_name, p.sections AS cache FROM pages p
  WHERE COALESCE(p.status,'') <> 'deleted'
    AND jsonb_typeof(p.sections) = 'array' AND jsonb_array_length(p.sections) > 0)
SELECT count(*) AS compared, count(*) FILTER (WHERE a.auth IS DISTINCT FROM c.cache) AS drift
FROM cache c JOIN aspect a ON a.site_id = c.site_id AND a.page_name = c.page_name
LEFT JOIN tier1 t ON t.site_id = c.site_id AND t.page_name = c.page_name
WHERE t.page_name IS NULL;          -- tier 2 only speaks where tier 1 is silent
```

## 3. The three stores plus the live page, for ONE page

The four-way read. `page_components` is the one that says whether this is a stale cache or a
genuine loss — 469's whole finding rests on it.

```sql
SELECT s.domain, p.name, p.status, p.build_status, p.sections::text AS cache,
       (SELECT jsonb_agg(sps.component_name ORDER BY sps.ordering)
          FROM site_plan_sections sps JOIN site_plans sp ON sp.id = sps.plan_id
         WHERE sp.site_id = p.site_id AND sp.is_current AND sps.page_name = p.name)::text AS authority,
       (SELECT string_agg(COALESCE(pc.slot_name,'?'), ', ' ORDER BY pc.position)
          FROM page_components pc
         WHERE pc.page_id = p.id AND COALESCE(pc.build_status,'') <> 'removed') AS live_slots
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE s.domain = :domain AND p.name = :page;
```

## 4. Does the page actually serve? — NEVER compose the URL

⚠ **`pages.status = 'archived'` does NOT mean "not serving".** `gripper-catalog` is archived
and serves 200. Four sessions have composed a URL by hand and read the 404 as damage; use
the script, which reads `pages.url` and runs an invented-URL control and a known-good
sibling control in the same run.

```bash
./scripts/probe-page-url.sh robot-hands.com gripper-catalog gripper-catalog-index
```

An answer is only meaningful when both control lines hold (`invented=` non-200,
`sibling=` 200).

## 5. Recovering what a lost section actually WAS

The bytes survive; the list does not. `page_component_history` (migration 357's trigger
pair) archives every deleted `page_components` row with `slot_name`, `position` and
`rendered_html`.

```sql
SELECT h.created_at, p.name AS page, h.slot_name, h.position, h.op, h.source,
       h.divergence, length(h.rendered_html) AS bytes
FROM page_component_history h JOIN pages p ON p.id = h.page_id
WHERE h.slot_name = :slot ORDER BY h.created_at DESC;
```

⚠ **Every rebuild deletes every row**, so this holds a `delete` for each section on each
build. "Which one was DROPPED" is only derivable by diffing consecutive builds — the
presence of a delete row proves nothing on its own.

## 6. The drift backlog and its receipts

```sql
SELECT id, site_id, status, created_at::date, summary FROM site_work_items
WHERE item_type = 'section_source_drift' ORDER BY created_at;

-- the frozen filing-time evidence (this is the ONLY record of the destroyed list)
SELECT jsonb_pretty(spec) FROM site_work_items WHERE id = :id;
-- how it was closed, and WHICH SIDE WON
SELECT jsonb_pretty(result) FROM site_work_items WHERE id = :id;
```

## 7. Is the check actually enabled?

A dead config key looks exactly like a live one. Ask the live row, not the seed.

```sql
SELECT type,
       (default_config->'workflow'->'steps'->'run_checks'->'config'->'checks')
         @> '["section_source_drift"]' AS has_ssd
FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL
  AND default_config->'workflow'->'steps'->'run_checks' IS NOT NULL;
```
→ `completeness-discovery-agent` is the only `t`.

## 8. How often does that agent run? (so you can size the closer's latency)

`orchestration_states` has no `agent_type` column and querying its jsonb is slow enough to
time out at 120 s. Use the work items it created instead — a lower bound on runs, and
indexed.

```sql
SELECT date_trunc('day', created_at)::date AS day, count(*) items, count(DISTINCT site_id) sites
FROM site_work_items WHERE created_by = 'completeness-discovery-agent'
  AND created_at > now() - interval '10 days' GROUP BY 1 ORDER BY 1 DESC;
```

## 9. Which checks can close their own items?

```bash
cd platform/orchestration/actions/discovery_checks
ls check_*.go | grep -v _test | wc -l                       # 71 checks
grep -l '\.Resolved = append' check_*.go | grep -v _test     # 19 can retract
grep -ln 'HandlerAgent: *""' check_*.go | grep -v _test      # 18 file flag-only items
```

## 10. Building and verifying

```bash
go build ./... && go test ./platform/orchestration/actions/discovery_checks/...
scripts/verify-head-builds.sh --with <file> --test    # BEFORE committing
scripts/verify-head-builds.sh                          # AFTER committing
```
⚠ Never hand-roll `git archive HEAD | tar` — that recipe is why this machine runs out of
space. `/tmp` is a 16 GB tmpfs (RAM); a full one presents as
`link: mapping output file failed: no space left on device`, which reads like a compiler
fault and is not one.
