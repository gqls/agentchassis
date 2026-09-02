# RUNBOOK — improvement loop

Every command here had a gotcha attached when I first ran it. The gotcha is the point.

`PSQL` below means:
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

---

## Is the loop running at all?

```sql
SELECT name, enabled, interval_seconds, last_triggered_at, last_completed_at
FROM scheduled_tasks WHERE name = 'improvement-sweep';
```

⚠ **Do not answer this from any document.** Several standing docs and one auto-memory
entry still carry the 2026-07-29 owner ruling that the sweep is off deliberately.
Migration `389` re-enabled it. The row is the fact; the ruling is history.

## What the loop did, and whether it did anything

Orchestration rows are purged in roughly a day, so **this window is all the evidence
there is** — a question you ask on Monday about Friday has no answer.

```sql
SELECT date_trunc('day', created_at)::date d, status, current_step, count(*)
FROM orchestration_states
WHERE owner_agent_type = 'improvement-loop' AND created_at > now() - interval '2 days'
GROUP BY 1,2,3 ORDER BY 1 DESC, 4 DESC;
```

⚠ **The column is `owner_agent_type`, not `agent_type`** — there is no `agent_type` on
this table and the query errors rather than returning zero, which is the good case.

⚠ **`execution_path` is EMPTY on these rows.** It looks like the natural way to ask
"which steps ran", and it will silently tell you "none of them" for every run. Read
`collected_data`'s keys instead:
`SELECT jsonb_object_keys(collected_data) FROM orchestration_states WHERE orchestration_id = '…';`

## Did the audit half run, or was the site skipped?

```sql
SELECT collected_data->'audit_state'->>'audit_due'      AS audit_due,
       collected_data->'audit_state'->>'not_converging' AS not_converging,
       count(*)
FROM orchestration_states
WHERE owner_agent_type = 'improvement-loop' AND created_at > now() - interval '2 days'
GROUP BY 1,2;
```

⚠ **`complete_clean` does NOT mean "site is clean".** It is also the terminus for a
site whose fingerprint has not changed (audit skipped, migration 291) and for a site
whose entire finding pile was held back as unroutable. Three very different states,
one step name. `audit_state` is what separates them — that separability was
`bugs_open/171`'s explicit requirement, so use it.

## How much work is held back, and of what kind

Per run:

```sql
SELECT collected_data->'triage_result'->>'promoted'       AS promoted,
       collected_data->'triage_result'->>'not_promotable' AS held,
       collected_data->'triage_result'->'not_promotable_by_type'
FROM orchestration_states
WHERE owner_agent_type = 'improvement-loop' ORDER BY created_at DESC LIMIT 5;
```

⚠ **Do not sum `not_promotable` across runs.** It is a per-run count of the site's
standing pile, so the same rows are counted once per visit. Summing gave me 3,866 for a
backlog whose true size is 1,385. For the standing figure, count the rows:

```sql
SELECT item_type, count(*), count(DISTINCT site_id) sites, min(created_at)::date oldest
FROM site_work_items
WHERE status = 'detected' AND (handler_agent IS NULL OR handler_agent = '')
GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **`handler_agent IS NULL` alone is not the predicate.** Migration `217` normalised
NULL to `''` and made the column NOT NULL; every live reader spells it
`(handler_agent IS NULL OR handler_agent = '')`. Ask only about NULL and you get zero,
which reads exactly like a clean estate.

⚠ **`site_work_items` is a rolling window** — `work-item-archiver` runs daily. Any
claim about a *population* over time must `UNION site_work_items_archive`. The query
above is deliberately about the *standing* pile, which is live-table-only by definition.

## Is a finding TRUE? Probe the page, never the row

```bash
for u in <recorded page_url> <invented-url-on-the-same-domain>; do
  code=$(curl -s -o /tmp/p.html -w '%{http_code}' -m 20 "$u")
  echo "$u | http=$code | $(grep -o -i -m1 '<title>[^<]*</title>' /tmp/p.html || echo NO-TITLE)"\
       "| footer=$(grep -c -i '<footer' /tmp/p.html) | bytes=$(wc -c </tmp/p.html)"
done
```

⚠ **The invented URL is not optional.** A parked domain answers 200 to every path —
boxingonline.com returns a 114-byte redirect stub for `/`, `/about.html` and anything
else you type. Without the control you cannot tell "our page is broken" from "this
domain is not ours to serve".

⚠ **`spec.missing` on an existing row is of unknown age.** `insertWorkItem` writes with
`dropOnConflict`, so a re-run of the check drops the fresh row and leaves the original
spec in place; and `head_essentials_missing` only retracts when *all* essentials are
present. A row whose skip link is still missing keeps its first-ever missing-list for
ever, however much of it has since been repaired. **Read the page, then the row.**

## Fire one loop run by hand

The sweep picks its own site (`ORDER BY s.updated_at ASC NULLS FIRST LIMIT 1`), so you
cannot choose one by waiting. Dispatch directly at the agent instead — and read
`scripts/kafka-publish-lib.sh` first rather than hand-rolling a `kcat -P`, which exits 0
having sent nothing.
