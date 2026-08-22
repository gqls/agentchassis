# RUNBOOK — `bugs_open/357` component identity

Every query/command that was hard to get right, with its gotcha attached. Change it HERE, not in
scrollback.

---

## The population (357's own query, unrestricted)

```sql
SELECT s.domain, p.name AS page, pc.created_at::date, pc.slot_name,
       length(pc.rendered_html) AS html_len
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
WHERE cc.name = 'hero'
  AND position(left(cc.html_template, position('{{' in cc.html_template) - 1) in pc.rendered_html) = 0
ORDER BY pc.created_at;
```

> ⚠ **This returns 22, while `bugs_open/357`'s own population table lists 9.** The nine are the
> subset that ALSO had a parked `required_fields_missing` work item (all single-component pages).
> Do not read the difference as 13 new rows overnight — check `created_at` before concluding
> anything about rate.

> ⚠ **`left(tmpl, position('{{' in tmpl) - 1)` is empty when a template STARTS with `{{`**, and
> `position('' in anything)` is 1, not 0 — so such a component can never be flagged by this test.
> Silent false-negative, not an error.

## The narrow predicate — use THIS for a guard, not the one above

```sql
WITH x AS (
  SELECT pc.id, cc.name AS comp, s.domain, p.name AS page,
     substring(cc.html_template  from 'data-component="([^"{]+)"') AS tmpl_attr,
     substring(pc.rendered_html  from 'data-component="([^"]+)"')  AS html_attr
  FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
  JOIN pages p ON p.id = pc.page_id JOIN sites s ON s.id = p.site_id
)
SELECT * FROM x WHERE tmpl_attr IS NOT NULL AND html_attr IS DISTINCT FROM tmpl_attr;
```

> ⚠ **`[^"{]` in the template pattern is load-bearing.** Without excluding `{`, a template whose
> attribute is itself interpolated (`data-component="{{.kind}}"`) captures the Go template
> expression and then "disagrees" with every row it ever rendered. Excluding `{` makes the test
> skip those components instead of convicting them.

> ⚠ **`IS DISTINCT FROM`, not `<>`.** With `<>`, a NULL `html_attr` — which is the whole
> pathological class here (stored HTML carries no attribute at all) — yields NULL and the row is
> silently dropped from the result. Using `<>` returns **0 rows** and reads as "nothing is wrong".

**Always print the agreement count in the same breath** — it is the demand control. A guard census
that reports "0 disagreements" proves nothing unless the same query shows the ~1,550 rows where the
comparison ran and agreed.

```sql
SELECT count(*) FILTER (WHERE tmpl_attr IS NOT NULL AND html_attr = tmpl_attr)          AS agree,
       count(*) FILTER (WHERE tmpl_attr IS NOT NULL AND html_attr IS DISTINCT FROM tmpl_attr) AS flagged,
       count(*) FILTER (WHERE tmpl_attr IS NULL)                                        AS not_testable
FROM x;
```

## Which writer wrote a `page_components` row (fingerprints)

No writer stamps itself, so read the marks it leaves:

| writer | `position` | `content_brief` | `build_status` |
|---|---|---|---|
| `save_page_sections_action.go` | `i+1` | **written** (`"{slot} section"`) | `deployed` |
| `deploy_tool_action.go` | `2` | never | `deployed` |
| `create_tool_component_action.go` | `2` | never | `deployed` |
| `adopt_verbatim.go` | `0` | never | `approved` |
| `create_report_page_action.go` / `rebuild_blog_listing_action.go` | — | check before relying | — |

> ⚠ **`rendered_html_digest` is NOT a writer fingerprint.** The INSERT writes `md5($3)`
> unconditionally; the column postdates older rows (`bugs_open/229` / IMP-052). Rows split by
> digest-present/absent split by **DATE**. I nearly filed that as evidence of a second writer.

## Is a page's tool still being served? (the only proof that counts)

```bash
curl -s "https://<domain>/<page>.html?cb=$(date +%s)" -o /tmp/p.html -w 'http=%{http_code} bytes=%{size_download}\n'
grep -c 'class="tool-page"'      /tmp/p.html   # the tool is present
grep -c 'data-component="hero"'  /tmp/p.html   # 0 here means NO hero rendered at all
```
`bugs_closed/287`: a `complete` work item is not a repaired artefact. Assert the tool's own markup,
not the item status.

## Diagnosis loop for this lane

Intake `f7aedef7-0bee-4c68-8cde-c86ac552e3e2` → **`RUN_CORRELATION_ID=e580b34a-d284-4f80-ac96-81af1c4adaba`**
(the run id is the one the dispatch loop mints and stamps back, NOT the intake id the script prints
first — artifacts are written under the run id).

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<RUN_CORRELATION_ID>';
```
Budget ~30 minutes: the council/diagnosis itself takes 2–5 min, the dispatch queues behind the fleet.
A missing row is latency, not a dropped dispatch — do not retry on that evidence.
