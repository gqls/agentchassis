# RUNBOOK — `bugs_open/302` design-repair completion verification

Every command that was hard to get right, with its gotcha attached. When one changes, change it
HERE.

DB access: `kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "..."`

---

## R1 — did gate 1b's unreadable-payload arm actually fire? (the whole premise)

```sql
SELECT occurred_at, work_item_id, context->>'item_type' AS item_type
FROM agent_error_log WHERE error_code='NO_CHANGE_GATE_UNREADABLE_RESULT'
ORDER BY occurred_at DESC;
```

⚠ **The column is `occurred_at`, NOT `created_at`** — `agent_error_log` has no `created_at` and
the query errors out rather than returning a wrong answer, which is the harmless direction.

⚠ **This row is written best-effort AFTER the completion is already allowed**
(`recordUnknownNoChangeShape`), so a missing row is not proof the arm did not fire — it is proof
the arm did not fire *or* the insert failed. Corroborate against the item's own payload (R2).

## R2 — the payload shapes behind those rows (this is what refuted the filing's account)

```sql
SELECT id, status, updated_at,
       (SELECT string_agg(k,',' ORDER BY k) FROM jsonb_object_keys(COALESCE(result,'{}')) k) AS result_keys
FROM site_work_items
WHERE item_type='dark_section_audit' AND status='complete'
  AND (result #> '{response,fix_result,total_fixed}') IS NULL
ORDER BY updated_at DESC;
```

⚠ **`jsonb_object_keys` is a set-returning function** — it needs the scalar subquery + `string_agg`
shown, not a bare call in the select list, or one row per key comes back and every count is wrong.

⚠ **A payload keyed `agent_id,agent_type,role,topics` is a SPAWN RECORD, not a handler reply** —
`bugs_closed/287`. Attributing it to the handler is the mistake the 302 filing made. Check the
timestamp against the `v1.0.1307` roll (2026-08-17 17:05Z) before drawing any conclusion.

## R3 — the demand control, which is the one that stops the fix being oversold

```sql
SELECT count(*) FROM site_work_items
WHERE item_type='dark_section_audit' AND updated_at > '<the roll>';    -- traffic on the opted-in type
SELECT count(*) FROM site_work_items
WHERE status='complete' AND updated_at > '<the roll>';                 -- traffic on the fleet
```

⚠ **Run BOTH.** The first alone cannot distinguish "the fix worked" from "nothing arrived"; the
second establishes the fleet was busy in the same window. On 2026-08-18 this read **0 against
1,862** — no demand, so no post-roll rate is measurable in either direction.

## R4 — the producer split, MANDATORY before registering any verifier

```sql
SELECT item_type, count(DISTINCT COALESCE(spec->>'audit_source','<none>')) AS producers,
       string_agg(DISTINCT COALESCE(spec->>'audit_source','<none>'), ' | ') AS which, count(*) n
FROM site_work_items WHERE item_type IN ('dark_section_audit','hardcoded_section_colors',
  'needs_design_review','contrast_failure','spacing_fix','responsive_fix','generic_theme','forced_text_colors')
GROUP BY 1 ORDER BY producers DESC, n DESC;
```

⚠ **Never ask this with `GROUP BY item_type` alone** — it shows one population and hides the
split. ⚠ **Never let `created_by` corroborate it**: it is `config[source]` falling back to
`params.AgentType`, bottoming out at the literal `generic`. A `2` or more here means a verifier
registered for that type would grade another producer's items (`bugs_closed/213`).

## R5 — is a registered verifier actually gating, in BOTH directions?

```sql
SELECT status, result->'_verification'->>'status' AS verif, count(*) n, max(updated_at) latest
FROM site_work_items WHERE item_type='<type>' GROUP BY 1,2 ORDER BY n DESC;
```

⚠ **A pass-only reading is not evidence.** Require a `defect_persists` (refused) AND a `verified`
(certified) row before believing the verifier discriminates. On `literal_markdown` 2026-08-18:
15 refused, 1 certified.

⚠ **A `complete` row with a NULL `_verification` is not automatically a bypassed gate** — check
`result->>'resolved_by'` first. The discovery-check retraction seam (WII-009) closes items on the
detector's own re-scan and correctly never runs the completion gate.

## R6 — which types are opted into which gate (read the code, not a doc)

```bash
grep -n '^\t"' platform/orchestration/actions/complete_work_item_no_change.go   # noChangeGates roster
grep -rn 'RegisterVerifier(' platform/ --include=*.go | grep -v _test.go        # gate 2 roster
grep -n '^\t"' platform/orchestration/actions/write_audit_findings_retraction.go # silenceRetractionGates
```

⚠ A concept-register STATUS line is a snapshot that outlives its truth (LANDMINES). WII-011 /
WII-013 / WII-017 / WII-018 describe these rosters; read the roster.
