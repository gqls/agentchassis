# RUNBOOK — bugs_open/116, link-integrity check coverage

Every query/command that was hard to get right, with its gotcha attached.

---

## Is a link check actually running? (the query most people get wrong)

**Gotcha: the check name and the work-item type are different strings.** Checks are
plural; `item_type` values are singular, and `misdirected_cta` files under a third
name entirely. Querying the check name returns 0 rows and reads as "never ran".

```sql
-- RIGHT: query the item types
SELECT item_type, created_by, count(*), count(DISTINCT site_id) AS sites, max(created_at) AS newest
FROM site_work_items
WHERE item_type IN ('phantom_internal_link','unbuilt_internal_link','empty_internal_href',
                    'dead_control','cta_names_unknown_destination')
GROUP BY 1,2 ORDER BY 3 DESC;
```

| check `Name()` | `item_type` it writes | where the literal lives |
|---|---|---|
| `phantom_internal_links` | `phantom_internal_link`, `unbuilt_internal_link`, `empty_internal_href` | `check_phantom_internal_links.go:315,:318,:325` (computed — `ItemType: f.IssueType` at `:153`) |
| `dead_controls` | `dead_control` | `check_dead_controls.go:160` |
| `misdirected_cta` | `page_rerender` **and** `cta_names_unknown_destination` | `check_misdirected_cta.go:318,:363` |

There is **no mapping table**. Each is an independent hardcoded literal.

## Has a site ever been audited? (do NOT count findings)

**Gotcha: counting sites-with-findings cannot answer this.** A site with no
findings is either clean or unexamined, which is the bug's own point. Use the
durable stamp `improvement-loop`'s `record_audit_pass` writes (migration 291):

```sql
SELECT domain, status,
       (settings#>>'{maintenance_profile,last_audit,at}')::timestamptz AS last_audit,
       settings#>>'{maintenance_profile,last_audit,passes_at_fingerprint}' AS passes
FROM sites ORDER BY last_audit DESC NULLS LAST;
```

**Second gotcha: exclude the non-sites before you quote a denominator.** 37 rows
include 17 `pool-*.internal` and `system.internal`. The real fleet is 19.

**Third gotcha: a NULL is not proof of "never".** The key is younger than the fleet
(written by a step introduced with migration 291), so NULL means "not audited since
the field existed".

## Is anything driving the audit automatically?

```sql
-- every enabled task, and what it targets
SELECT name, interval_seconds, target_agent_type, enabled, last_triggered_at, last_completed_at
FROM scheduled_tasks WHERE enabled ORDER BY last_completed_at DESC NULLS LAST;

-- the answer as a single question
SELECT count(*) FROM scheduled_tasks
WHERE enabled AND (target_agent_type ILIKE '%discovery%' OR target_agent_type = 'improvement-loop');
```

**Gotcha — `last_completed_at` LIES on a disabled row.** `improvement-sweep` shows
`enabled=f` with a `last_completed_at` from minutes ago. `cmd/scheduler/main.go:360`
(`loadDueTasks`) selects `WHERE enabled = true`, so the scheduler is **not** firing
it. The freshness comes from `improvement-loop`'s own `notify_scheduler` step:

```sql
-- verbatim from the live agent definition
UPDATE scheduled_tasks SET last_completed_at = NOW() WHERE name = 'improvement-sweep'
```

which runs on every improvement-loop completion, however it was dispatched.
**Read `enabled` and `last_triggered_at`; ignore `last_completed_at` for this
question.** (`last_triggered_at` is 2026-05-02 — that is the honest column.)

## Reading an agent's workflow without drowning in jsonb_pretty

A whole `default_config` is ~70KB and will blow up a tool result. Get the shape
first, then one step at a time:

```sql
-- the shape: one row per step
SELECT s.key AS step, s.value->>'action' AS action,
       COALESCE(s.value->'config'->>'agent_type', s.value->>'agent_type','') AS agent,
       s.value->>'next_step' AS next_step
FROM agent_definitions a, jsonb_each(a.default_config #> '{workflow,steps}') s
WHERE a.type='improvement-loop' AND a.is_active
  AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
ORDER BY 1;

-- then one step, by name
SELECT jsonb_pretty(default_config #> '{workflow,steps,notify_scheduler}')
FROM agent_definitions WHERE type='improvement-loop' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Which checks is an agent configured to run?

```sql
SELECT jsonb_array_length(default_config #> '{workflow,steps,run_checks,config,checks}') AS n,
       default_config #> '{workflow,steps,run_checks,config,checks}' AS checks
FROM agent_definitions WHERE type='completeness-discovery-agent' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

## Finding which sessions hold which bug (before taking one)

`scripts/who-owns.py` reads **commits**, so a session mid-fix is invisible to it.
This scans live transcripts instead:

```bash
cd ~/.claude/projects/-home-ant-projects-agentchassis/
for f in $(find . -maxdepth 1 -name '*.jsonl' -mmin -240); do
  echo "$(basename $f .jsonl) :: $(tail -c 600000 "$f" \
    | grep -oE 'bugs_open/[0-9]{3}' | sort | uniq -c | sort -rn | head -4 \
    | awk '{printf "%s(%s) ", $2, $1}')"
done | grep -v ':: *$'
```

**Gotcha:** `tail -c` on some of these files panics (uutils `tail` on a growing
JSONL). The panic goes to stderr and the loop continues, so redirect `2>/dev/null`
and do not read a clean-looking run as complete coverage. Also: a bug number
appearing in a transcript may just be `MEMORY.md` being loaded — check the *count*,
and read the top hit rather than every hit.

## The audit chain, for anyone tracing a finding back

```
improvement-loop
  ensure_site_record → enrich_news_feed → load_audit_state
  → check_audit_due  (conditional: audit_state.audit_due == true)
      → spawn_quality_discovery      → call_quality_discovery
      → spawn_design_discovery       → call_design_discovery
      → spawn_completeness_discovery → call_completeness_discovery   ← the 3 link checks
      → spawn_design_audit → call_design_audit
      → spawn_site_review  → call_site_review
      → record_audit_pass  (writes maintenance_profile.last_audit)
  → triage_findings (TriageDetectedItemsAction — the ONLY detected→triaged promoter)
  → check_has_findings → … → notify_scheduler
```

`completeness-discovery-agent`'s `run_checks` step is `action: run_discovery_checks`,
scoped `site_id: site_record.site_id` — **whole-site, not per page**. All three link
checks read stored `rendered_html` (DB only, no HTTP, no LLM).
