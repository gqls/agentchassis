# RUNBOOK — CTA destination relevance (`bugs_open/389`)

Every command here was needed to get an answer right, with its gotcha attached.

## The ranking the resolver will actually use (the core query)
```sql
SELECT s.domain, p.name, COALESCE(p.nav_order,100) AS nav_order,
       row_number() OVER (PARTITION BY s.id ORDER BY COALESCE(p.nav_order,100), p.name) AS rank
FROM pages p JOIN sites s ON s.id=p.site_id
WHERE p.page_type IN ('tool','game') AND p.status IN ('active','deployed')
ORDER BY s.domain, rank;
```
⚠ **Mirror the code exactly or the simulation proves nothing**: `COALESCE(nav_order,100)`, then
`name` (**not** `url`) as the tiebreak, and the `('tool','game')` + `('active','deployed')`
filters. Those are `chooseCTATargets` / `loadInteractivePages`
(`platform/orchestration/actions/resolve_internal_links_action.go:651,918`).

## Split CTA urls by who wrote them — minted vs authored
```sql
SELECT (pc.content_data->'__cta_minted') ? kv.key AS minted, count(*),
       min(pc.updated_at)::date, max(pc.updated_at)::date
FROM page_components pc
CROSS JOIN LATERAL jsonb_each(pc.content_data) kv
WHERE kv.key LIKE '%\_url' AND kv.value #>> '{}' LIKE '%<your-target>%'
GROUP BY 1;
```
⚠ **Three states, not two.** `t` = the resolver minted *this* url; `f` = a stamp exists for the
field but names a *different* url, so the current value reads authored; **NULL = no stamp at all**
and means *"not recorded"*, never *"authored"* — the stamp only shipped **2026-08-22**
(`datahelpers/cta_provenance.go`, LNK-035) and there is **no backfill by design**. Treating NULL
as authored will halve your live count.
⚠ The key is `%\_url` — escape the underscore or `_` is a single-char wildcard.

## Is a CTA defect LIVE or historical?
Read `max(updated_at)` on the **minted=t** rows only. That is the whole test: a stamped mint dated
after the stamp mechanism shipped is the resolver acting now. Ours read **today**.

## Confirm at the served bytes — with the control
```bash
curl -s "https://<domain>/<page>.html?cb=$(date +%s)" \
  | grep -o '<a[^>]*<your-target>[^>]*>[^<]*</a>' | sed 's/<[^>]*>//g'
```
⚠ **The obvious grep passes while the bug is untouched.** These sites link the tool legitimately
from `/tools.html` and the nav, so `grep -c '<target>'` on a page is non-zero either way. Assert on
the **anchor in a `hero`/`call-to-action` slot**, or read the stored `cta_url`/`primary_cta_url`
field, which is unambiguous.
⚠ Probe the page that actually holds the field — the home page was clean on two of three sites
while `/services.html` carried it.

## Before reading a flag as a human judgement
```sql
SELECT in_header, count(*) FROM pages
WHERE page_type IN ('tool','game') AND status IN ('active','deployed') GROUP BY 1;
```
⚠ Run this **first**. `in_header=false` is 62.7% of tool pages — the majority state. It refuted a
13-site claim of mine that was one edit away from being filed.

## Chassis capability probe (never infer from ancestry)
```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=3000 | grep -m1 'build provenance'
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1 | cut -d/ -f2)
kubectl -n ai-persona-system exec $POD -- grep -ac "<symbol>" /proc/1/exe
```
⚠ Always run a control symbol that must be **absent** in the same breath, and remember `grep -c`
exits 1 on zero matches, so `|| echo ERR` fires on a legitimately-zero control.
