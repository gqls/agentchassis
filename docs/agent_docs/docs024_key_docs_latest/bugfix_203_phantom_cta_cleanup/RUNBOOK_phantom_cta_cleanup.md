# RUNBOOK — bugfix 203 cleanup

Commands that were hard to get right, with their gotchas.

## Council verdict + full report for the source fix

```sql
-- verdict summary (doc_notes):
SELECT body FROM doc_notes WHERE body LIKE '%42eda9a5%' ORDER BY created_at DESC LIMIT 1;
-- full seat-by-seat report:
SELECT body FROM diagnosis_artifacts
WHERE kind='council_report' AND correlation_id='42eda9a5-6188-4e89-a11a-adb1dcbb135f'
ORDER BY created_at DESC LIMIT 1;
```
Gotcha: `diagnosis_artifacts` has no `content` column — the text is `body`. Schema-first.
Gotcha: `orchestration_states` lookup by
`collected_data->'input_data'->>'fix_correlation_id'` returned nothing for this corr —
the run had long completed; the durable artefacts are doc_notes + diagnosis_artifacts.

## Liveness of a Go fix without a pod-grep (ancestry route)

```bash
git merge-base --is-ancestor 880a405a6 1e349d046 && echo carried
# 1e349d046 = fix(197), pod-proven live on v1.0.1259 by that lane, real traffic.
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.metadata.name} {.spec.containers[0].image}{"\n"}{end}'
```
Gotcha: only valid because builds are `git archive` from committed HEAD on a
forward-only tree AND the anchor commit was proven at the pod by someone. An image tag
alone proves nothing (bugs_open/153).

## Census of shipped phantom-CTA instances (stored artefact)

```sql
SELECT s.domain, p.url, pc.content_data->>'cta_text' AS cta_text
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=pc.site_id
WHERE pc.content_data ? 'cta_text' AND NOT (pc.content_data ? 'cta_url')
  AND pc.rendered_html ~ 'href="[^"]*"[^>]*>[^<]*</a>';
```
Gotcha: reads STORED html. Served pages can differ in both directions (LANDMINE:
"phantom check reads STORED html, not served"). Verify per-row at the live URL before
and after cleanup.
