# RUNBOOK — durable-write completeness guard (bugs_open/021 INSTANCE 1)

Every query/command worth reusing, with its gotcha attached. DB access:

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Where the guard lives / the simulation harness (from component_write_guard.go)

The guard was calibrated against `component_versions` transitions. To re-run that
simulation before changing any threshold (compare consecutive versions of each
component; a hard shrink that ends cleanly is a legitimate rewrite, one ending
mid-token is the shape to refuse):

```sql
WITH v AS (SELECT component_id, version_number, html_template AS cur,
       lead(html_template) OVER (PARTITION BY component_id ORDER BY version_number) AS nxt
     FROM component_versions)
SELECT c.name, length(cur), length(nxt),
       round(100.0*length(nxt)/length(cur)) AS pct,
       (right(rtrim(nxt),1)='>') AS ends_cleanly
FROM v JOIN content_components c ON c.id=v.component_id
WHERE nxt IS NOT NULL AND length(nxt) < length(cur)
ORDER BY 1.0*length(nxt)/length(cur) ASC;
```

## Recovery-table check (which durable fields have a history table)

```sql
-- history/version/snapshot tables
SELECT tablename FROM pg_tables WHERE schemaname='public'
  AND (tablename LIKE '%version%' OR tablename LIKE '%history%'
       OR tablename LIKE '%snapshot%') ORDER BY 1;
```
Findings 2026-07-21:
- `component_versions` — snapshots `html_template` (the 012 recovery table). GUARDED source.
- `page_component_history` — snapshots **`content_data` only**, NOT `rendered_html`.
  9,933 rows, latest 2026-07-21. So rendered_html has NO direct history table;
  recovery is by re-render from html_template.
- `site_snapshots.pages_snapshot` — captures pages.rendered_header/footer/head in
  JSONB, but the columns are always NULL so it captures/restores NULL.

## Target A verification — are pages.rendered_* really never written?

Grep proved no Go writer. Confirm against live data that they are empty
(a non-empty value would mean a writer exists somewhere I missed):

```sql
SELECT
  count(*) FILTER (WHERE rendered_header IS NOT NULL AND rendered_header<>'') AS hdr,
  count(*) FILTER (WHERE rendered_footer IS NOT NULL AND rendered_footer<>'') AS ftr,
  count(*) FILTER (WHERE rendered_head   IS NOT NULL AND rendered_head  <>'') AS head,
  count(*) AS total_pages
FROM pages;
```
Expect hdr=ftr=head=0. [RESULT: pending — run before final sign-off.]

## Target B verification — is rendered_html ever NOT reproducible from html_template?

The design says rendered_html is derived. The counter-claim (save_page_sections
:276-285) is that interactive tools "exist ONLY as rendered_html". Find any live
tool page_component whose rendered_html is materially larger than / diverges from
its component's html_template (i.e. rendered_html carries bytes the template does
not):

```sql
-- tool components: compare rendered_html length on the page vs the durable template
SELECT pc.page_id, cc.name, cc.function,
       length(pc.rendered_html) AS rh_len, length(cc.html_template) AS tmpl_len,
       round(100.0*length(pc.rendered_html)/NULLIF(length(cc.html_template),0)) AS pct
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
WHERE cc.function IN ('interactive_tool','tool','game')   -- refine to real tool functions
  AND pc.rendered_html IS NOT NULL
ORDER BY rh_len DESC
LIMIT 40;
```
Interpretation: if rh_len ≈ tmpl_len (+ the fixed tool-doc header) the render is
an identity over the template ⇒ recoverable ⇒ no separate guard needed. If rh_len
>> tmpl_len for real tools, rendered_html holds durable content the template does
not, and IS a genuine unguarded durable source. [RESULT: pending.]

## site_components.rendered_html — the chrome writers (open question)

If the chrome (header/footer/head, keyed site_id+slot_name) is written from a
whole LLM artifact, it shares the 012 shape and is a real INSTANCE-1 target. Find
the writers:
```
grep -rn "site_components" platform/ --include=*.go | grep -i "INSERT\|UPDATE\|rendered_html"
```
[RESULT: pending — being classified by agent 1.]

## Build / deploy (from CLAUDE.md — for when/if code ships)

- `make build-agent-chassis` builds from committed HEAD. Commit the task first.
- Bump `IMAGE_TAG` (makefile ~line 16) every build.
- Verify against the running pod, never git:
  `kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "<symbol I CREATED>"'`
  Grep a literal the change CREATED, plus a positive control — not one it merely uses.
