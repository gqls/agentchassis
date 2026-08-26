# RUNBOOK — render audit rotation cursor (bugs_open/394)

The commands, each with the gotcha attached. When one changes, change it HERE.

---

## §1 The truncation rows — the whole signal, in one query

```sql
SELECT occurred_at, agent_type, step_name, domain, jsonb_pretty(context)
FROM agent_error_log
WHERE error_code = 'RENDER_AUDIT_TRUNCATED'
ORDER BY occurred_at;
```

**Read `context`, not the message.** The message is prose and truncates in most viewers; the
context carries `max_pages`, `pages_total`, `pages_audited` exactly, and `agent_type` is what
tells you WHICH caller truncated. That is how the bug's `[UNEXPLAINED]` "5 of 26" row was
resolved — the message alone could not distinguish an agent from an override.

**Gotcha:** `agent_error_log` rows expire (365 d declared, 14 d once a consumer resolves them).
Extract before resolving anything.

**Gotcha:** the originating `orchestration_states` row is a **rolling window** and will usually
be gone. Do not plan on recovering a dispatch's config from it after a few days.

## §2 Where the cap actually bites — the ordering, made visible

The action orders by `COALESCE(nav_order, 999), name` and takes the first `max_pages`. To see
the exact cut for a site, reproduce that ordering and look at the boundary:

```sql
WITH live AS (
  SELECT name, url, nav_order,
         row_number() OVER (ORDER BY COALESCE(nav_order,999), name) AS rn
  FROM pages
  WHERE site_id = '<SITE_ID>'::uuid
    AND status = 'active'
    AND NOT (deployed_at IS NULL AND COALESCE(build_status,'') <> 'deployed')
    AND COALESCE(url,'') <> '')
SELECT rn, nav_order, name FROM live WHERE rn BETWEEN 55 AND 66 ORDER BY rn;
```

**Gotcha — copy the predicate, do not retype it.** The two arms are
`datahelpers.PageWantedLivePredicateFor("")` and `datahelpers.PageHasShippedPredicateFor("")`,
and the second is deliberately NOT `build_status='deployed'` — a `needs_rebuild` page that once
deployed is still serving its previous artefact, which is exactly what a RENDER audit should
photograph (bugs_open/185 tranche 2 measured 36 such pages across 8 sites). A hand-spelled
`build_status='deployed'` here silently audits a different population from the one shipping.

## §3 The class-shaped tail — the query that decided the fix

```sql
SELECT nav_order, count(*), min(name), max(name)
FROM (
  SELECT name, nav_order,
         row_number() OVER (ORDER BY COALESCE(nav_order,999), name) AS rn
  FROM pages WHERE site_id = '<SITE_ID>'::uuid AND status='active'
    AND NOT (deployed_at IS NULL AND COALESCE(build_status,'')<>'deployed')
    AND COALESCE(url,'') <> ''
) t
WHERE rn > <CAP>
GROUP BY 1 ORDER BY 1;
```

This is the one that turns "86 pages are unaudited" into "**all 45 `tool-*-guide` pages at
`nav_order` 200 are unaudited, and no cap below 98 reaches them**". Group the tail by
`nav_order` before proposing any cap — a tail that is one nav band is a coverage bug; a tail
that is scattered is a sampling bug, and they want different fixes.

## §4 Fleet exposure at a candidate cap

```sql
SELECT s.domain, count(*) AS live_pages, greatest(count(*) - <CAP>, 0) AS tail
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.status='active'
  AND NOT (p.deployed_at IS NULL AND COALESCE(p.build_status,'') <> 'deployed')
  AND COALESCE(p.url,'') <> ''
GROUP BY 1 HAVING count(*) > <CAP> ORDER BY 2 DESC;
```

Run it at **every** live cap, not just the one you came for. `[MEASURED 2026-08-26]` at 60 it
returns one site; at 8 it returns twenty-five. Answering only for your own caller is how half a
problem gets called a whole one.

## §5 Who calls the action, and with what cap

```sql
SELECT type, s.key AS step, s.value->'config'->>'max_pages'
FROM agent_definitions a, jsonb_each(a.default_config->'workflow'->'steps') s
WHERE a.deleted_at IS NULL AND COALESCE(a.is_snapshot,false)=false AND a.is_active
  AND s.value->>'action' = 'request_render_audit';
```

**Gotcha:** this returns the STANDING config only. A dispatch can override `max_pages`, and one
did (the 08-11 `5 of 26`). So this query bounds who *can* call it, never what a given run
actually used — for that, read the `context` in §1.

**Gotcha (LANDMINES):** do not reach for `default_config::text LIKE '%request_render_audit%'` —
that matches prompt text as well as steps, and two agents "contain" action names they do not
run. `jsonb_each` over the steps object is the honest form.

## §6 The driver's cadence

```sql
SELECT name, target_agent_type, interval_seconds, enabled, pre_query
FROM scheduled_tasks WHERE name = 'site-render-audit-rotation';
```

**Gotcha:** `interval_seconds = 3600` is how often the TASK fires, not how often a site is
audited. The per-site cadence is the `now() - interval '3 days'` clause inside `pre_query`, and
the task takes `LIMIT 1`. Reading only the column misreports the cadence by a factor of 72.

## §7 Ownership — re-check at every phase boundary, and check the DIRECTORY

```bash
python3 scripts/who-owns.py 394
git status --short -- docs/agent_docs/docs024_key_docs_latest/bugfix_394_render_audit_rotation_cursor/ \
                      platform/orchestration/actions/request_render_audit_action.go
git log --oneline -- docs/agent_docs/docs024_key_docs_latest/bugfix_394_render_audit_rotation_cursor/
```

**The `git status` on your own lane directory is the load-bearing one and it is new advice.**
`who-owns.py` reads commits, so two sessions opening the same lane within minutes both pass it
— which is exactly what happened on `bugs_open/359` this morning (WRONG_CALLS, 2026-08-26). A
second session's files appear in `git status` within seconds of being written, and in nothing
else. Run it right after you create the directory, and again before the first code commit.

## §8 Deploy ordering, and proving it shipped

Go changes are inert until an image rolls; DB config is live immediately. So:

```
commit → make build-agent-chassis → bump IMAGE_TAG → push → roll → verify at the binary → apply migration
```

Never the other way round: the discovery/agent runner reads config the moment it is applied, and
a config that names a capability the binary lacks fails at run time.

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <your-commit> <the stamp>   # "did my fix ship?" is a query
```

**Gotcha:** the provenance line is a STARTUP line and scrolls. An empty grep means "not in
range", never "unstamped" — fall back to the binary probe, and always run a control string that
must be present alongside the one that must be absent.
