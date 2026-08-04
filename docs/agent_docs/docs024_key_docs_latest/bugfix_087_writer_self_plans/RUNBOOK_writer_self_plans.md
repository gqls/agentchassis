# RUNBOOK — `bugs_open/087`, the self-planning writer

Every command here had a gotcha. The gotcha is attached to the command; if one
changes, change it **here**, not in your scrollback.

## Census: which agents call `page-content-writer`, and which of them plan?

**The obvious query is wrong.** `default_config::text LIKE '%plan_sections%'` is not a
test for that step: in SQL `LIKE`, `_` is a **single-character wildcard**, so the pattern
also matches the substring `plan.sections` inside `section_plan.sections_ready`. That
makes `page-content-writer` — which had no planning step at all — read as a positive.
The wrong answer looks exactly like the right one.

```sql
-- Read the STEP KEYS, not the serialised text.
SELECT ad.type,
       bool_or(s.v->>'action' = 'plan_sections')                       AS plans,
       bool_or(s.v::text LIKE '%page-content-writer%')                 AS calls_writer,
       bool_or(s.v #>> '{config,input_mapping,section_plan}' IS NOT NULL) AS maps_plan
FROM agent_definitions ad,
     LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s(k,v)
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
GROUP BY 1 HAVING bool_or(s.v::text LIKE '%page-content-writer%');
```

Note this still only sees **top-level** steps. `pageflow-builder` and
`site-work-orchestrator` call the writer from inside a `loop`'s `sub_workflow`, so for
those you must descend — or read the whole step blob:

```sql
SELECT jsonb_pretty(jsonb_object_agg(k, v))
FROM agent_definitions ad, LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s(k,v)
WHERE ad.type='<agent>' AND ad.is_active AND COALESCE(ad.is_snapshot,false)=false
  AND ad.deleted_at IS NULL AND v::text LIKE '%page-content-writer%';
```

## A verify block whose comparison can never fire

`#>>` on a missing jsonb path yields **NULL**, and `NULL <> 'x'` is NULL — **not TRUE**.
So `IF cfg #>> '{...,items_field}' <> 'expected' THEN RAISE` sits permanently green
against a key that does not exist. That is exactly what happened while writing seed 309:
the loop's key is `iterate_over`, not `items_field`, and the assertion passed against
NULL. **Use `IS DISTINCT FROM` for every string comparison in a migration verify block**,
and `COALESCE(<containment>, false)` for every `@>`.

And prove the block can fail before trusting it — run the `DO` block **alone** against
the unmodified row and require an exception:

```bash
awk '/^DO \$\$/,/^END \$\$;/' docs/agent_docs/sql_for_agents/309_*.sql > /tmp/verify_only.sql
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < /tmp/verify_only.sql
# expect: ERROR: 087/309: one of the four new steps is missing
```

A verify block made only of `SELECT`s cannot stop a `COMMIT` — `ON_ERROR_STOP` ignores a
non-empty result set. Use `DO` / `RAISE EXCEPTION`.

## Applying the seed

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
  < docs/agent_docs/sql_for_agents/309_page_content_writer_plans_its_own_sections.sql
```

Read `updated_at` on the row **immediately before** applying. ~30 sessions share this
cluster and `page-content-writer` was edited by the `192` lane at 09:01:35Z the same
morning; a surprise timestamp means read the row before you write it.

## Choosing an acceptance target — bounding the blast radius

`page-rebuild`'s `get_pages_to_rebuild` is `get_pages_to_build` with
`build_statuses: ["needs_rebuild"]` and `include_all: false`, and its loop runs to
`max_iterations: 20`. **It rebuilds every armed page on the site, not the one you had in
mind.** So pick a site with **zero** currently armed pages, then arm exactly one:

```sql
SELECT s.domain,
       count(*) FILTER (WHERE p.build_status='needs_rebuild') AS armed,
       count(*) FILTER (WHERE p.rebuild_policy='generic' AND p.build_status='deployed') AS generic_deployed
FROM sites s JOIN pages p ON p.site_id=s.id
GROUP BY 1 HAVING count(*) FILTER (WHERE p.build_status='needs_rebuild')=0
              AND count(*) FILTER (WHERE p.rebuild_policy='generic' AND p.build_status='deployed')>0;
```

Then arm **inside a transaction that aborts if the count is not 1** — checking after the
fact is not the same thing, because another session can arm a page between your check
and your dispatch:

```sql
BEGIN;
UPDATE pages SET build_status='needs_rebuild' WHERE id='<page>'::uuid AND build_status='deployed';
DO $$ DECLARE armed int; BEGIN
  SELECT count(*) INTO armed FROM pages p JOIN sites s ON s.id=p.site_id
  WHERE s.domain='<domain>' AND p.build_status='needs_rebuild';
  IF armed <> 1 THEN RAISE EXCEPTION 'expected exactly 1 armed page, found % — ABORT', armed; END IF;
END $$;
COMMIT;
```

Two more filters that matter:
- **`rebuild_policy` must not be `owned`.** `save_page_sections` refuses an owned page
  ("a generic section save would clobber it") — that is the guard that blocked 087's
  2026-07-28 attempt, and it fires *after* `deploy_page` has already run.
- **Pick a page whose `name` does not equal its `url` stem**, or the `bugs_closed/125`
  path assertion is vacuous. `tool-cma-obligation-checker-guide` →
  `/guides/tool-cma-obligation-checker-guide.html` gives a real negative control at
  `/tool-cma-obligation-checker-guide.html`, which must stay 404.
- **Grep the live session transcripts for the page name** before touching it —
  `who-owns.py` reads commits and cannot see a session mid-edit.

## Dispatching

```bash
SEND=1 DOMAIN=<domain> SITE_ID=<uuid> \
  ./docs/agent_docs/docs024_key_docs_latest/about_page_commercial/p1_trigger_rebuild.sh
```

`SEND=1` must be a **same-line prefix**; `SEND=1; ./script` does not reach the process.
Dry-run is the default. Do not dispatch within ~300s of a chassis pod (re)start — the
spawn is silently dropped. Save the printed `CORRELATION_ID`; track by correlation,
never by `created_at`.

## Tracking, and the retention trap

```sql
SELECT owner_agent_type, status, current_step,
       EXTRACT(EPOCH FROM (NOW()-last_activity))::int AS since_s
FROM orchestration_states
WHERE correlation_id='<corr>'::uuid
   OR collected_data->'input_data'->>'parent_correlation_id'='<corr>'
ORDER BY created_at;
```

`orchestration_states` keeps **terminal rows for ~24 hours**, not the ~20 days a
whole-table `min(created_at)` suggests (unreaped `CANCELLED`/`RUNNING` rows set that
floor). Capture what you need from a run **the same day**, and bound retention per
status if you ever quote it:
`SELECT status, count(*), min(created_at) FROM orchestration_states GROUP BY status;`

## Asserting the outcome

```sql
-- components actually rewritten (compare md5 + updated_at against the before-state)
SELECT pc.slot_name, md5(pc.content_data::text), pc.updated_at
FROM page_components pc WHERE pc.page_id='<page>'::uuid ORDER BY pc.slot_name;
```

```bash
# the path assertion: canonical must serve, name-derived must STAY 404
curl -s -o /dev/null -w "%{http_code} %{size_download}\n" https://<domain>/<url>
curl -s -o /dev/null -w "%{http_code} %{size_download}\n" https://<domain>/<name>.html
```

Restore `build_status` if the run does not complete:
`UPDATE pages SET build_status='deployed' WHERE id='<page>'::uuid;`
