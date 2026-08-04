# RUNBOOK — bugfix 107 homepage skeleton

## §1 Fleet homepage-composition census

The measurement that re-validated the bug. Gotcha: filter
`parent_instance_id IS NULL` or nested children inflate the list; order by
`position`, not created_at.

```sql
SELECT s.domain, s.created_at::date,
       (SELECT string_agg(cc.function, ' > ' ORDER BY pc.position)
        FROM page_components pc JOIN content_components cc ON cc.id = pc.component_id
        WHERE pc.page_id = p.id AND pc.parent_instance_id IS NULL) AS composition
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.name IN ('index','index.html') AND p.status='active'
ORDER BY s.created_at;
```

Run via: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

## §2 Competing-work checks (do all three, they see different things)

```bash
scripts/who-owns.py 107            # commits only — LAGGING, blind to live sessions
# live sessions (hits 1-3 are usually just `ls bugs_open` listings; read context before concluding):
cd ~/.claude/projects/-home-ant-projects-agentchassis && \
  for f in $(find . -maxdepth 1 -name '*.jsonl' -mmin -2880); do \
    c=$(grep -c "107_HANDOFF" "$f"); [ "$c" -gt 1 ] && echo "$f : $c"; done
```

Queue: `SELECT item_type, summary, status FROM site_work_items WHERE status NOT IN ('complete','cancelled','rejected') AND summary ILIKE '%planner%' ...`
