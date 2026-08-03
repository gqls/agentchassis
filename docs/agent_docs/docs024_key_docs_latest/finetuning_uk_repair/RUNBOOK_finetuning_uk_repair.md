# RUNBOOK — finetuning.uk repair / framework audit

Every command here was needed and had to be got right. Gotchas attached.

## Site identity

```
finetuning.uk  →  site_id 1368e337-dd1d-4799-bbb3-8221a1b79bcc
```

## Find broken images the DB can see, fleet-wide

**THE TRAP THAT COST THE FIRST RUN: Postgres regex does not understand `\b`.**
In POSIX regex `\b` is a BACKSPACE, not a word boundary — the word-boundary
escape is `\y`. A pattern written `<img\b[^>]*\bsrc…` returns **zero rows with no
error**, which reads exactly like "the fleet is clean". It is not; it is a
silently mis-spelled query. Use `[[:space:]]` or `\y`, never `\b`.

```sql
-- Every <img src> that is a bare word: no slash, no dot. One row per (site, token).
WITH src AS (
  SELECT s.domain,
         (regexp_matches(pc.rendered_html,
            '<img[^>]*[[:space:]]src[[:space:]]*=[[:space:]]*"([^"]*)"', 'gi'))[1] AS v
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN sites s ON s.id = p.site_id
  WHERE pc.rendered_html IS NOT NULL
  UNION ALL
  SELECT s.domain,
         (regexp_matches(sc.rendered_html,
            '<img[^>]*[[:space:]]src[[:space:]]*=[[:space:]]*"([^"]*)"', 'gi'))[1]
  FROM site_components sc
  JOIN sites s ON s.id = sc.site_id
  WHERE sc.rendered_html IS NOT NULL      -- chrome: one bad path here is EVERY page
)
SELECT domain, v AS bare_token_src, count(*) AS occ
FROM src
WHERE v !~ '[/.]' AND v !~ '^[[:space:]]*$' AND v <> '#' AND v NOT LIKE 'data:%'
GROUP BY 1,2 ORDER BY 1, 3 DESC;
```

Attribute them to a component — this is what turns 31 findings into one fix:

```sql
SELECT s.domain, COALESCE(c.name,'(none)') AS component, p.url,
       (regexp_matches(pc.rendered_html,
          '<img[^>]*[[:space:]]src[[:space:]]*=[[:space:]]*"([^"/.]+)"', 'g'))[1] AS token
FROM page_components pc
JOIN pages p ON p.id = pc.page_id
JOIN sites s ON s.id = p.site_id
LEFT JOIN content_components c ON c.id = pc.component_id   -- LEFT: component_id can be NULL
WHERE pc.rendered_html ~ '<img[^>]*[[:space:]]src[[:space:]]*=[[:space:]]*"[^"/.]+"'
ORDER BY s.domain, p.url;
```

Gotcha: the components table is **`content_components`**, not `components`
(which does not exist). `\dt | grep component` returns ~40 backup tables — the
live ones are `content_components`, `page_components`, `site_components`.

## Confirm a "broken image" is actually broken

The DB says the src is a bare word; only HTTP says it 404s. Both, always:

```bash
for u in /cpu /network /database /assets/images/case-study-facilities.jpg; do
  printf '%s  %s\n' "$(curl -sS -o /dev/null -w '%{http_code}' -L "https://finetuning.uk$u")" "$u"
done
```

## Is the icon library even loaded?

Before replacing `<img src>` with `<i data-lucide>`, prove the page can render
one. A fix that renders into a page with no icon library is a different bug:

```bash
for u in "https://finetuning.uk/" "https://finetuning.uk/about.html" \
         "https://ai-agent-orchestration.com/" "https://ai-agent-orchestration.com/about.html"; do
  echo "lucide=$(curl -sS -L "$u" | grep -c 'lucide.min.js')  $u"
done
```

## Why work items are not being worked

The whole answer is two facts. The dispatcher's claim query:

```sql
-- platform/orchestration/actions/load_work_item_actions.go, LoadWorkItemsAction
WHERE wi.status IN ('triaged', 'approved')
  AND wi.attempt_count < wi.max_attempts
```

and the fleet's actual distribution:

```sql
SELECT status, count(*) AS items, count(DISTINCT site_id) AS sites
FROM site_work_items
WHERE status IN ('detected','triaged','approved','unresolved')
GROUP BY 1 ORDER BY 2 DESC;
-- 2026-08-03: unresolved 235/8, detected 204/10, triaged 2/1
```

`detected` is not claimable. The only promoter is `triage_findings` inside the
improvement-loop; its only schedule is `improvement-sweep`, off since
2026-05-02:

```sql
SELECT name, enabled, target_agent_type, target_topic, last_triggered_at
FROM scheduled_tasks WHERE name IN ('improvement-sweep','build-pipeline-trigger');
```

**`attempt_count = 0` is the tell.** It distinguishes "the handler tried and
failed" from "nothing ever picked this up". Do not read a summary prefixed
`[unresolved after 2 attempts]` as current: those strings are historical and the
counter has since been reset.

## Run the framework against one site

```bash
./docs/agent_docs/docs024_key_docs_latest/finetuning_uk_repair/294_TRIGGER_improvement_loop_v1.sh \
  1368e337-dd1d-4799-bbb3-8221a1b79bcc finetuning.uk
```

Runs discovery + audit → `triage_findings` → `call_dispatch`. Two pre-flights are
enforced by the script, both of which have bitten this estate before:

- **Nothing may be dispatched within ~300s of a chassis pod (re)start** — the
  spawn is silently dropped and looks exactly like queue latency.
- **Check the queue, not just the pod** — another session may have work in flight
  on this site.

`FORCE=1` overrides both, after reading why they fired.

Watch it. **A missing row for the first minutes is LATENCY, not a drop** — do not
re-fire on that evidence:

```sql
SELECT current_step, status, updated_at FROM orchestration_states
WHERE orchestration_id = '<the printed ORCH_ID>';
```

## Did the deploy actually ship your Go change?

Never trust git, the tag, or a roll. Grep the running binary, on **every**
replica, with a **positive control** — a string that must already be there. A
control is what proves the pipeline works; without it, `0` is ambiguous between
"not shipped" and "my grep is wrong".

```bash
for p in $(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name); do
  echo "--- $p ---"
  kubectl exec -n ai-persona-system ${p#pod/} -- sh -c '
    echo -n "NEW  : "; strings /app/agent-chassis | grep -c "image_url_404:bare-token-src"
    echo -n "CTRL : "; strings /app/agent-chassis | grep -c "image_url_404:empty-src"'
done
# 2026-08-03 10:12 — NEW 0 / CTRL 1 on both replicas: the fix is committed and NOT live.
```

## Apply a component-template change

`sql_for_agents/` files are hand-run (`platform/database/migrations/` is the
auto-applied dir — do not put a one-off config change there).

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
  -f - < docs/agent_docs/sql_for_agents/293_departments_grid_renders_an_icon_not_a_photo.sql
```

**A verify block made of `SELECT`s cannot stop a `COMMIT`** — `ON_ERROR_STOP`
does not fire on a non-empty result set. Verify in a `DO $$ … RAISE EXCEPTION`
block, which aborts the transaction. 293 does this; copy its shape.

## Run the checker's tests

```bash
go test ./platform/orchestration/actions/discovery_checks/ -run TestImageURL404 -count=1 -v
```

The negative-control test is the load-bearing one. The cheap way to silence a
false positive is to widen the pattern until it reports nothing, and a check that
reports nothing looks exactly like a check that found nothing.
