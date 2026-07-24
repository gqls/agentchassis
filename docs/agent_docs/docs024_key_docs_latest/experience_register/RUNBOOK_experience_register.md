# RUNBOOK — experience register

Commands/queries proven useful so far. Update HERE when one changes.

## DB access

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Travelling-docs substrate

Current plan for a subject (exact-key only — doc_plans has NO metadata column, NO structured
search; that is why the register needs its own table):

```sql
SELECT subject_type, subject_key, is_current, created_at, left(body, 200)
FROM doc_plans WHERE subject_type='<type>' AND subject_key='<key>' AND is_current;
```

Read the LIVE subject_type CHECK (do not trust docs — 163/184 both changed it):

```sql
SELECT conname, pg_get_constraintdef(oid) FROM pg_constraint
WHERE conname IN ('doc_plans_subject_type_check','doc_notes_subject_type_check');
```

Gotcha: the DB CHECK is only one of FOUR enforcement points. The Go gate
`docResolveSubject` (platform/orchestration/actions/write_doc_plan_action.go:136) is shared
by write_doc_plan + append_doc_note + load_doc_context; `persist_diagnosis_note_action.go:78`
has its own separate allowlist (tool/pipeline only). See bugs_open/064.

## Bug filing

Next free number = max across BOTH dirs + 1 (numbering is one shared sequence, never
reassigned; several numbers are duplicated by historical accident — resolve by slug):

```
ls bugs_open/ bugs_closed/ | grep -oE '^[0-9]+' | sort -n | tail -1
```

Ownership check before routing work at an existing bug: `scripts/who-owns.py <number|slug>`.

## Later phases (do not fire yet)

- Experience-plan compose+council for one site experience (P3+, gauntlet session owns the
  vonc pilot): `092_TRIGGER_experience_plan.sh <domain> <experience_key>` — parked rule: only
  after tools-api is deployed + smoke-POSTed, liveness evidence via the 197 compose-decisions
  block.
- Council gate for the P2 platform change-set:
  `097_TRIGGER_council_review_v1.sh <submission.json>` — budget ~30 min (dispatch queues
  behind the fleet); find the run by payload correlation, not the printed id.
- Component-selector scoring reference (the shape the register's selection copies):
  `platform/orchestration/actions/component_selector.go` SelectComponentByType.
