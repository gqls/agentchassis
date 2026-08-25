# RUNBOOK — the `deferred` work-item park

Every query that was hard to get right, with its gotcha attached.

## The population, split by whether anyone can tell who parked it

**This is the query that matters.** The naive `WHERE status='deferred'` mixes three unrelated
populations and reads as one problem.

```sql
SELECT COALESCE(spec->>'parked_by','(NO parked_by)')            AS parked_by,
       count(*)                                                  AS rows,
       count(*) FILTER (WHERE spec ? 'parked_reason')            AS has_reason,
       count(*) FILTER (WHERE spec ? 'not_dispatchable')         AS declares_undispatchable,
       count(*) FILTER (WHERE COALESCE(handler_agent,'') <> '')  AS names_a_handler,
       count(DISTINCT item_type)                                 AS types
FROM site_work_items WHERE status='deferred'
GROUP BY 1 ORDER BY 2 DESC;
```

⚠ **`deferred` + EMPTY `handler_agent` is CORRECT and is not the defect.** It is the estate's
deliberate roadmap convention with a live consumer and a live drain. Counting it as damage inflates
the finding roughly three-fold and puts you in an argument with six well-commented code sites.

⚠ **`parked_by = 'migration_389'` rows are `bugs_open/296` and are OWNED** by the
`bugfix_131_contrast_ratio_check` lane. Exclude them and say you have.

The suspicious population is the third: named handler, no `parked_by`, no `not_dispatchable`.

```sql
SELECT item_type, source, count(*), min(created_at)::date, max(updated_at)::date
FROM site_work_items
WHERE status='deferred' AND COALESCE(handler_agent,'')<>''
  AND NOT (spec ? 'parked_by') AND NOT (spec ? 'not_dispatchable')
GROUP BY 1,2 ORDER BY 3 DESC;
```

## Is a given row dispatchable? — read the CONSUMERS, never the row

```sql
SELECT status, count(*), max(attempt_count), min(created_at)
FROM site_work_items WHERE site_id='<sid>' GROUP BY 1;
```

⚠ **Treat anything outside `('triaged','approved','claimed')` as NOT GOING TO HAPPEN without a hand
promotion.** A `deferred` row looks *more* live than a `detected` one — real handler, real
priority, real `item_key`, no `error`. `attempt_count = 0` is the only hint and it reads as
"not started yet" rather than "will never start".

The three predicates, none of which is in the row:

| what | where | selects |
|---|---|---|
| dispatch | `claim_work_item_action.go:102` | `status IN ('triaged','approved')` |
| promotion | `triage_detect_items_action.go` | `status='detected'` |
| dedup slot held | `idx_swi_dedup` (read from `pg_indexes`) | `status <> ALL (complete, verified, rejected, wont_fix, failed, unresolved, cancelled)` |

`deferred` is absent from the first two and from the third's exclusion list — undispatchable,
un-promotable, and still holding its slot.

## Before dispatching ANY work item, check the slot

```sql
SELECT s.domain, p.url,
       COALESCE((SELECT w.status FROM site_work_items w
                 WHERE w.site_id=p.site_id
                   AND w.item_key='page_rerender_'||p.name||'_'||p.site_id||'_assemble'
                   AND w.status <> ALL (ARRAY['complete','verified','rejected','wont_fix',
                                              'failed','unresolved','cancelled'])
                 LIMIT 1),'FREE') AS slot
FROM pages p JOIN sites s ON s.id=p.site_id WHERE (s.domain,p.url) IN ( ... );
```

⚠ **A taken slot fails your INSERT with 23505, which reads as "already queued".** Re-arm the
existing row (`SET status='triaged'`) rather than inserting a duplicate, and **leave its
`source`/`created_by` alone** — that is another producer's provenance.

## ⚠ The test that CANNOT tell you what you want to know

```sql
-- DOES NOT WORK: "was this row born deferred, or moved there later?"
SELECT count(*) FILTER (WHERE updated_at - created_at < interval '5 seconds') AS born_deferred ...
```

`trg_site_work_items_updated_at` is BEFORE UPDATE FOR EACH ROW and bumps `updated_at` on **every**
write. So a row born `deferred` and later touched by anything is **indistinguishable** from one
created in another status and deferred later. This query returned "0 of 205 born deferred" and that
number means nothing. `site_work_items` keeps no status history.

**What IS evidence:** `spec->>'parked_by'` (migration 389 stamps it), and the absence of any Go
path that does `UPDATE … SET status='deferred'`.

## Who writes `deferred` in Go — read them, do not grep them

```bash
grep -rn "deferred" --include=*.go platform/ internal/ pkg/ cmd/ | grep -iE "status|Status"
```

⚠ **`plan_sections_action.go`'s four hits are a DIFFERENT `deferred`** — a section-plan status
(`"ready" | "deferred" | "skipped"`, declared at `:906`), not a work-item status. Counting them as
work-item writers is the obvious mistake and inverts the conclusion, because it makes it look as
though a Go path does produce the shape.

The six real writers — `remit.go:202`, `write_audit_findings_action.go:427`/`:584`,
`load_work_item_actions.go:279`, `check_palette_contrast.go:138`, `check_content_duplication.go:251`
— **all pair it with `HandlerAgent: ""`**.

## Parking something on purpose? Copy migration 389

`docs/agent_docs/sql_for_agents/389_park_contrast_failures_and_reenable_improvement_sweep.sql` is
the model, and the only bulk park in the estate that can be audited afterwards:

- a precondition `SELECT … INTO n_before` that `RAISE EXCEPTION`s if the premise is already gone;
- `spec.parked_from_status`, `spec.parked_reason` (naming the bug and the restore condition), and
  `spec.parked_by`;
- `GET DIAGNOSTICS n_parked = ROW_COUNT` asserted against `n_before` — aborts on a race;
- a negative control proving nothing else moved.

⚠ A verify block of bare `SELECT`s **cannot** stop the `COMMIT` (`ON_ERROR_STOP` ignores a non-empty
result). Use `DO` / `RAISE`, as 389 does, and induce the failure once to prove the guard fires.
