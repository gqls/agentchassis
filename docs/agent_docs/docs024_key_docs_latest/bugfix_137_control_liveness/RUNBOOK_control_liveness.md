# RUNBOOK — control-liveness / runtime-fill scope (`bugs_open/137`)

Every command here was needed to get something right. Gotchas attached.

## Is the bug's premise still live?

```bash
curl -s https://vonc.com/provocations/index.html | grep -o '<a[^>]*href="#"[^>]*>'
# 2026-07-31: exactly one hit, the data-archive-template row. Unchanged since filing.
```

## Find every call site of the exemption

**Grep the marker, not the helper.** The predicate was inlined at each site, so
searching for a function name finds nothing.

```bash
grep -rn 'strings.Contains(html, "data-runtime-fill")\|Contains(renderedHTML, "data-runtime-fill")\|runtimeFillMarker' \
  --include=*.go platform/ internal/ pkg/ | grep -v _test.go
```

**Then read each caller's INPUT scope** — that is the whole question, and grep
cannot answer it. `pc.rendered_html` in a row loop is section-scoped and fine;
an assembled page or a fetched URL is not. There is also a **SQL-side** copy
(`html_template LIKE '%data-runtime-fill%'` in `check_component_standards`,
`check_component_template_corrupted`, `check_required_fields_missing`) which the
Go grep does not see.

## Measure the masked population — the honest version

**GOTCHA, and it cost a wrong number.** Do not join pages by `name`: page names
are **not unique across sites**, so `page IN (SELECT page FROM ... )` returns
every page called `index` in the fleet. Join on `page_id`.

```sql
WITH c AS (
  SELECT s.domain, p.id AS page_id, p.name AS page, COALESCE(pc.slot_name,'') AS slot,
         COALESCE(pc.rendered_html,'') AS html,
         COALESCE(pc.rendered_html,'') LIKE '%data-runtime-fill%' AS is_shell
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
  WHERE p.status='active' AND pc.rendered_html IS NOT NULL AND pc.rendered_html<>''
), pg AS (
  SELECT page_id, bool_or(is_shell) AS page_has_shell FROM c GROUP BY page_id
)
SELECT c.domain, c.page, c.slot,
       (SELECT count(*) FROM regexp_matches(c.html,'href\s*=\s*["'']\s*["'']','g')) AS empty_hrefs
FROM c JOIN pg USING (page_id)
WHERE pg.page_has_shell AND NOT c.is_shell
  AND c.html ~ 'href\s*=\s*["'']\s*["'']'
ORDER BY 1,2,3;
-- 2026-07-31: exactly one row — vonc.com / index / gauntlet-cta, 2 empty hrefs.
```

**Measure the right class for the consumer you mean.** `RepairPageLinks` only
touches `LinkScopePage` and `LinkScopeEmpty`; `href="#"` is `LinkScopeAnchor`
and it never touched those anyway. Counting `#` hrefs answers the
*dead-controls* question, not the *repair* question — they are different
populations and I conflated them once (see NOTES).

## Assemble a real page to test against

The checks read either a whole served page or one component. To exercise the
assembled-page path without deploying:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At <<'SQL' > /tmp/page.html
SELECT string_agg(pc.rendered_html, E'\n' ORDER BY pc.position)
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE s.domain='vonc.com' AND p.name='index' AND pc.rendered_html IS NOT NULL;
SQL
```

**And fetch the REAL page-URL set for the index** — a toy index makes
`RepairPageLinks` unlink links that resolve perfectly well in production, which
reads as a finding and is not one:

```sql
SELECT p.url FROM pages p JOIN sites s ON s.id=p.site_id
WHERE s.domain='vonc.com' AND p.url IS NOT NULL AND p.url<>''
  AND COALESCE(p.status,'') NOT IN ('deleted','archived');
-- 18 rows for vonc.com on 2026-07-31
```

## Prove the tests are load-bearing (both directions)

A suite asserting only "the neighbour is now judged" passes against a change
that simply **deletes** the exemption. Mutate the predicate each way and check
which tests fall over:

```bash
# MUTANT 1 — restore the old whole-document behaviour
#   insert at the top of RuntimeFillSpans:
#   if strings.Contains(html, RuntimeFillMarker) { return ByteSpanSet{{0, len(html)}} }
# expect: 8 failures, including both consumers' "neighbour" cases.

# MUTANT 2 — delete the exemption
#   insert at the top of RuntimeFillSpans:  if true { return nil }
# expect: 10 failures, including the PRE-EXISTING TestRepairPageLinks_RuntimeFillShellIsExempt.

go test ./platform/orchestration/datahelpers/ 2>&1 | grep -E '^\s+---|FAIL|ok '
```

Restore from a copy afterwards — keep the backup **outside** the repo
(scratchpad), not as a `.bak` beside the source, or a sweeping session commits it.

## Build and test

```bash
go build ./platform/... ./internal/... ./pkg/...   # NOT ./... — docs/ has a
                                                   # two-package dir that breaks it
go test ./platform/orchestration/...
```

**`go build ./...` fails on this tree for reasons unrelated to you**:
`docs/agent_docs/.../traffic_probe/deploy_setup/working_dir` holds two packages
in one directory. Separately, **`cmd/reasoningset` does not compile at HEAD**
(committed and clean, three `declared and not used` at `main.go:504`) — that is
a pre-existing HEAD breakage, not your change. Build the trees you touched.

## Verify after the roll

```bash
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -c RuntimeFillSpans; strings /app/agent-chassis | grep -c DeadControlAnchors'
# first = your change (0 before the roll), second = the positive control (non-zero either way).
```
Both numbers in **one exec**, per the fleet rule that a roll is not evidence a
fix shipped (`bugs_open/153`).
