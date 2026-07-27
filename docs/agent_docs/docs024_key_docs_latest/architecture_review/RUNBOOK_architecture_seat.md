# Runbook — architecture seat / council memory

Commands that were hard to get right, with the gotcha attached. Append here when
one changes; do not leave it in scrollback.

## Read the live council roster (any of the four council-bearing agents)

There is **not one council**. Four agents carry `review_*` steps, at three
lifecycle points. Filtering to `type='council-gate'` answers a question about a
small world and tells you nothing about the other three — this cost a wrong claim
to the owner on 2026-07-26 (`WRONG_CALLS.md`).

```sql
SELECT d.type, key
FROM agent_definitions d, jsonb_object_keys(d.default_config->'workflow'->'steps') key
WHERE d.type IN ('council-gate','fix-proposer','feature-designer','experience-planner')
  AND d.is_active AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND key LIKE 'review_%'
ORDER BY d.type, key;
```

## Read the council's own minutes (259 verdicts, full text)

`diagnosis_artifacts` has always been in the seats' schema hint; nobody had told
them. `kind='council_report'`, `body` is the full verdict JSON.

```sql
SELECT kind, count(*) FROM diagnosis_artifacts GROUP BY 1;   -- council_report / fix_plan / bundle / escalation
```

Guardian verdict distribution, and objections, need two `jsonb_array_elements`
unnests — the reviews array, then each review's objections:

```sql
WITH g AS (
  SELECT created_at, correlation_id, jsonb_array_elements(body::jsonb->'reviews') AS r
  FROM diagnosis_artifacts WHERE kind='council_report'
)
SELECT r->>'verdict', count(*) FROM g WHERE r->>'reviewer'='guardian' GROUP BY 1;
```

**Gotcha:** `body` is `text`, not `jsonb` — cast it (`body::jsonb`) or the arrow
operators fail.

## Re-run the D5 ossification measurement

Recurrence per deflected core site. Site tagging is `ILIKE` on objection text, so
the counts are **floors** — it undercounts anything phrased without the symbol name.

```sql
WITH g AS (
  SELECT created_at, correlation_id, jsonb_array_elements(body::jsonb->'reviews') AS r
  FROM diagnosis_artifacts WHERE kind='council_report'
), gg AS (
  SELECT created_at, correlation_id, jsonb_array_elements(r->'objections')->>'problem' AS problem
  FROM g WHERE r->>'reviewer'='guardian'
)
SELECT count(DISTINCT correlation_id)
FROM gg
WHERE (problem ILIKE '%higher%layer%' OR problem ILIKE '%less-foundational%'
       OR problem ILIKE '%battle-tested%' OR problem ILIKE '%foundational%')
  AND problem ILIKE '%ProcessResponse%';
```

Churn side, from git — the split is the whole point, so always take both:

```bash
git log --since="60 days ago" --oneline -- platform/orchestration/ | wc -l                     # 366
git log --since="60 days ago" --oneline -- platform/orchestration/actions/ | wc -l             # 348
git log --since="60 days ago" --oneline -- platform/orchestration/ \
        ':(exclude)platform/orchestration/actions/' | wc -l                                    # 55
```

**Gotcha:** the headline 366 reads as alarming churn and is not — it is a plug-in
registry growing. Quote the split or the number misleads.

## Change a seat prompt (D8a′, applied 2026-07-27)

**Never hand-patch `council-gate`.** Patch `fix-proposer`, then mirror.

```bash
/tmp/acm/APPLY_council_memory.sh                       # patches fix-proposer; prints the 5 seats
python3 .../fixloop_eg_dartsonline/099_SYNC_gate_roster.py            # DRY RUN first
python3 .../fixloop_eg_dartsonline/099_SYNC_gate_roster.py --apply    # snapshots, then writes
RESTORE=1 /tmp/acm/APPLY_council_memory.sh             # rollback fix-proposer, then re-mirror
```

Three gotchas, all confirmed the hard way:

1. **The mirror copies `review_*` and `gate_*` steps only — NOT `load_schema_hint`**
   (099 line 117 carries non-review steps over from the *gate's own* copy). So a
   **prompt** change rides the mirror to two agents; a **schema-hint** change is a
   four-place edit across all four council-bearing agents. Prefer the prompt route.
2. **Push the JSON as base64**, not as a quoted SQL literal — the prompts contain
   quotes, backslashes and `$`, and any of them will mangle a heredoc:
   `convert_from(decode('<b64>','base64'),'UTF8')::jsonb`.
3. **Verify the diff is prompt-text-only before applying** — assert the step set is
   unchanged and that no config key other than `prompt_template` differs. A step
   accidentally added or renamed breaks routing for every concurrent session.

Verify after (expect 5 seats × 2 agents = 10 rows):

```sql
SELECT d.type, s.key
FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
WHERE d.type IN ('council-gate','fix-proposer') AND d.is_active
  AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
  AND s.key LIKE 'review_%' AND (s.value->'config'->>'prompt_template') ILIKE '%council_report%'
ORDER BY 1,2;
```

**Config is live immediately — no image, no roll.** Which also means a mistake is
live immediately; the dry run is not optional.

## Size the written corpus (before proposing to inline any of it)

```bash
wc -c docs/agent_docs/docs024_key_docs_latest/016b_debugging_guide_8_consolidated.md \
      docs/agent_docs/docs024_key_docs_latest/016_debugging_guide_v2_58_consolidated.md \
      docs/agent_docs/docs024_key_docs_latest/WRONG_CALLS.md
cat bugs_open/*.md | wc -c ; cat bugs_closed/*.md | wc -c
grep -c '^### ' docs/agent_docs/docs024_key_docs_latest/016b_debugging_guide_8_consolidated.md
```

~3.3 MB across 124 files against `max_tokens: 8000` — un-inlinable. But `016b` §9
is one-line dated headings, so the **heading index** is the promptable subset.
