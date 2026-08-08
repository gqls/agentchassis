# RUNBOOK — 117 chrome staleness reference

Every command here was needed and had to be got right. Gotchas attached.

## R1 — the decisive cross-tab: does the live detector agree with the real signal?

This is the query the case rests on. It populates all four cells, so it can come
out otherwise.

```sql
WITH real_stale AS (
  SELECT sc.site_id, sc.slot_name, (cc.updated_at > sc.updated_at) AS truly_stale
  FROM site_components sc JOIN content_components cc ON cc.id = sc.component_id
  WHERE sc.rendered_html IS NOT NULL AND sc.rendered_html <> ''
),
detector AS (
  SELECT sc.site_id, sc.slot_name,
         (sc.updated_at < mx.latest - INTERVAL '24 hours') AS detector_fires
  FROM site_components sc
  CROSS JOIN LATERAL (
    SELECT MAX(pc.updated_at) AS latest FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
    WHERE p.site_id = sc.site_id AND pc.rendered_html IS NOT NULL) mx
  WHERE sc.rendered_html IS NOT NULL AND sc.rendered_html <> '' AND mx.latest IS NOT NULL
)
SELECT r.truly_stale, d.detector_fires, count(*) AS rows
FROM real_stale r JOIN detector d USING (site_id, slot_name)
GROUP BY 1,2 ORDER BY 1 DESC, 2 DESC;
```

**Gotcha:** the `detector` CTE must reproduce the check's predicate *exactly*,
including the `- INTERVAL '24 hours'` and the `pc.rendered_html IS NOT NULL`
filter. Drop either and you are cross-tabbing against a predicate that is not
the one in production, and the disagreement you measure is your own.

**Gotcha:** join `content_components` with an inner join here (not `LEFT`), or
the 3 rows with `component_id IS NULL` silently become `truly_stale = NULL` and
vanish from a `GROUP BY` reading. Count them separately — they are a finding in
their own right, not noise.

## R2 — the census that frames it

```sql
SELECT s.domain, sc.slot_name, sc.updated_at AS chrome_at,
       cc.name AS component, cc.is_active, cc.updated_at AS component_at,
       (cc.updated_at > sc.updated_at) AS stale
FROM site_components sc
JOIN sites s ON s.id = sc.site_id
LEFT JOIN content_components cc ON cc.id = sc.component_id
ORDER BY stale DESC NULLS FIRST, s.domain, sc.slot_name;
```
Reads 57 rows (19 sites × 3 slots). `LEFT` join **here** deliberately, so the
null-provenance rows show up at the top.

## R3 — the work-item history: is the check dormant or firing?

```sql
SELECT item_key, status, handler_agent, count(*) AS n,
       min(created_at)::date AS first, max(created_at)::date AS last
FROM site_work_items
WHERE item_key LIKE 'stale\_sc\_%' OR item_key LIKE 'deactivated\_%'
GROUP BY 1,2,3 ORDER BY last DESC NULLS LAST;
```
**Gotcha:** escape the `_` in `LIKE` (`stale\_sc\_%`). Unescaped, `_` is a
single-character wildcard and the pattern quietly widens.

**Why it matters:** a detector that fires and drains is *worse* than a dormant
one when its predicate is wrong, because its output is believed and it consumes
real rebuild capacity. Check this before assuming "0 findings" means anything.

## R4 — do NOT do this: the wider-timestamp reference

Recorded so the next session does not re-derive it.

```sql
-- REJECTED. Marks ~every row stale; sites.updated_at churns for unrelated reasons.
GREATEST(cc.updated_at, nav.mx, s.updated_at) > sc.updated_at
```
Run any candidate predicate against live data before proposing it. ~100% or ~0%
means you are measuring the wrong thing.

## R5 — finding a template-version discriminator (and why the obvious one lies)

To test whether stored chrome differs from what the current template produces,
you need a literal that is **unconditional** in the template.

```sql
WITH lits AS (
  SELECT DISTINCT (regexp_matches(html_template, '(class="[a-z0-9_-]{4,40}")', 'g'))[1] AS lit
  FROM content_components WHERE name='footer-theme-chrome'
), f AS (
  SELECT sc.rendered_html h, (sc.updated_at > '<template-change-ts>'::timestamptz) AS newr
  FROM site_components sc
  WHERE sc.slot_name='footer'
    AND sc.component_id=(SELECT id FROM content_components WHERE name='footer-theme-chrome')
)
SELECT l.lit,
       count(*) FILTER (WHERE f.newr AND f.h LIKE '%'||l.lit||'%') AS new_has,
       count(*) FILTER (WHERE f.newr) AS new_total,
       count(*) FILTER (WHERE NOT f.newr AND f.h LIKE '%'||l.lit||'%') AS old_has,
       count(*) FILTER (WHERE NOT f.newr) AS old_total
FROM lits l CROSS JOIN f GROUP BY l.lit;
```

**Gotcha (this cost a wrong conclusion):** a token that appears only in the new
renders may be **data-gated, not version-gated**. `class="footer-compliance"`
sits inside `{{if .compliance_lines}}`. Before trusting any discriminator, read
its template context:
```sql
SELECT substring(html_template,
                 greatest(1, position('<token>' in html_template) - 400), 800)
FROM content_components WHERE name='<component>';
```
If it is inside a `{{if …}}`, it discriminates **site data**, not template
version. Discard it.

**Gotcha:** `psql` alias-in-ORDER-BY. `ORDER BY (new_has::float/…)` fails with
`column "new_has" does not exist` — output-column aliases are not visible in an
expression in `ORDER BY`. Wrap the aggregate in a CTE and order outside it.

## R6 — filing the 090 diagnosis

```bash
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/090_TRIGGER_needs_diagnosis_v1.sh "<symptom>"
```
**Gotcha:** the script prints two correlations. The intake one is *not* the key
the artifacts are written under — `RUN_CORRELATION_ID` is. Save both.

Retrieve the run's evidence:
```sql
SELECT iteration, kind, source_agent, length(body), metadata
FROM diagnosis_artifacts WHERE correlation_id='<RUN_CORRELATION_ID>' ORDER BY iteration;
```
**Gotcha:** the columns are `kind` and `body` — not `artifact_type`/`content`,
and there is no `verdict` column. `\d diagnosis_artifacts` before guessing.

**Gotcha, and it is the important one:** the bundles hold the *evidence
gathered*, not a verdict. On this run no `doc_notes` row joined to either
correlation and no bundle contained a `VERDICT` line — the known
"verdict computed then thrown away" defect (commit `0252b3cae`).
**A `complete` work item is not a verdict.** Do not report a diagnosis as
confirming anything you have not read.

## R7 — shell traps that bit

- **Single-quote grep patterns containing `$`.** `grep -n "… = \$1 WHERE id = \$2"`
  returned zero matches and read as "the code changed under me". It had not.
- `find -newermt '-4 hours'` is rejected by this box's `bfs`-backed `find`.
  Use an absolute stamp: `CUT=$(date -d '-5 hours' '+%Y-%m-%dT%H:%M:%S')`.
- `tail -c N` on a live-appended `.jsonl` can panic (`InvalidInput`). Tolerate
  it in a loop (`2>/dev/null`) rather than letting it kill the sweep.

## R8 — checking a bug is genuinely unowned

Three checks, all of them lagging in different ways, so do all three:
```bash
scripts/who-owns.py <number>                 # reads COMMITS — blind to live sessions
CUT=$(date -d '-5 hours' '+%Y-%m-%dT%H:%M:%S')
find /home/ant/.claude/projects/-home-ant-projects-agentchassis -name '*.jsonl' -newermt "$CUT"
# then grep each for the bug number AND for the symbols you intend to edit
```
**Gotcha:** grep the **symbols**, not just the bug number. Four sessions
mentioned `site_components` without mentioning 117; none touched
`check_integrity`. The symbol grep is what told me the code was free.

**Gotcha:** `bugs_open/` contains **finished** bugs by owner ruling (2026-08-06).
Presence there is not evidence a bug is unfixed — 2 of my 8 candidates (126, 181)
were already closed and live. Read the file's status section.
