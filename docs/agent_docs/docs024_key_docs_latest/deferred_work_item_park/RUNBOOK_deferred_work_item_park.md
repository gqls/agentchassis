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

## ⚠ ENUMERATE the stamps; do not check for the ones you know

This is the query the 2026-08-25 diagnosis loop effectively ran and I had not. I checked
`spec ? 'parked_by'` and `spec ? 'not_dispatchable'` — the two conventions I knew about — and
missed a **third**, `spec.deferred_reason`, carrying a full owner-sanctioned park explanation.

```sql
SELECT k, count(*)
FROM site_work_items w, LATERAL jsonb_object_keys(w.spec) k
WHERE w.status='deferred' AND COALESCE(w.handler_agent,'')<>''
GROUP BY 1 ORDER BY 2 DESC LIMIT 20;
```

⚠ **Not every key is provenance.** `spec.reason` appears on 22 rows and records why the item was
**detected** (`cta_links_stale`, `not_built`, `no_style_collection`, `post_reconcile_assembly`), not
why it was parked. Reading it as a trace shrinks the population on a false basis. Open one row per
key and decide, rather than pattern-matching the key name.

## The split, computed rather than eyeballed

Get the arithmetic from the database, not from reading a result table. Two counts were published
wrong by eye on 2026-08-25 before this rule was written down.

```sql
WITH d AS (SELECT * FROM site_work_items WHERE status='deferred')
SELECT CASE
         WHEN spec ? 'parked_by'               THEN '1. traceable: parked_by ('||(spec->>'parked_by')||')'
         WHEN spec ? 'deferred_reason'         THEN '2. traceable: deferred_reason'
         WHEN COALESCE(handler_agent,'') = ''  THEN '3. correct convention: EMPTY handler'
         ELSE                                       '4. UNTRACEABLE: named handler, no stamp'
       END AS population,
       count(*) AS rows,
       count(*) FILTER (WHERE spec ? 'not_dispatchable') AS declares_itself,
       count(DISTINCT item_type) AS types
FROM d GROUP BY 1 ORDER BY 1;
```

[MEASURED 2026-08-25 12:47Z] 303 total = 87 + 4 + 98 + **114**. Re-run before quoting.

## Ruling a candidate writer in or out — use its FINGERPRINT, and prove the fingerprint is live

`FailWorkItemAction` looked decisive: it writes a step-config string straight into `status` and
leaves `handler_agent` untouched, and the comment above it names the exact handlers on the parked
rows. It is still not the writer, because it stamps `handled_by = agentType`:

```sql
-- the test AND the control, in one statement
SELECT count(*) AS suspects,
       count(*) FILTER (WHERE COALESCE(handled_by,'')<>'') AS carry_the_fingerprint
FROM site_work_items
WHERE status='deferred' AND COALESCE(handler_agent,'')<>'' AND NOT (spec ? 'parked_by');

SELECT status, count(*), count(*) FILTER (WHERE COALESCE(handled_by,'')<>'') AS has_handled_by
FROM site_work_items GROUP BY 1 ORDER BY 2 DESC;   -- <- the CONTROL
```

⚠ **Without the second query the zero means nothing** — a column nobody writes and a column written
everywhere-but-here give the same answer. [MEASURED 2026-08-25] `handled_by` is set on **7,114 of
7,329** `complete` rows, so the zero is a real absence.

## A co-occurring actor is not a writing actor

`agent_error_log` retains to 2026-07-24 and covers every bulk-park minute, so you can ask what else
was running:

```sql
SELECT to_char(occurred_at,'MM-DD HH24:MI'), domain, agent_type, step_name, count(*)
FROM agent_error_log
WHERE occurred_at BETWEEN '<minute-1>' AND '<minute+3>'
GROUP BY 1,2,3,4 ORDER BY 1;
```

⚠ It will find something — this fleet is always busy. A discovery run completing at the same minute
on the same site is **equally consistent** with it parking the rows and with it merely *touching*
them while they were already parked. That is the same trap as the `updated_at` test above, one
layer along, and it killed two hypotheses on 2026-08-25.

## Do NOT chase `HandleUpdateWorkItem`

The 2026-08-25 diagnosis verdict names it as *"the only call site the index shows that SETs
`handler_agent` on an UPDATE"*. **The symbol does not exist**: `grep -rn "func Handle.*WorkItem"`
returns nothing repo-wide, and every `handler_agent` reference under `internal/core-manager/admin/`
is an INSERT column list. The verdict's own caveat — only a signature was indexed, not a body — is
the tell. The code index lags the working branch; verify a symbol before spending a round on it.
