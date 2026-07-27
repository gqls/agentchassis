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

## Pre-flight before ANY live config push (used for APPLY_gap.sh, 2026-07-27)

Config is live instantly and there is no undo, so prove three things first. All
three are cheap; skipping the third is what makes a rollback file a guess.

```bash
# 1. Has another session written this row since your payload was generated?
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT type, updated_at, length(default_config::text) FROM agent_definitions
WHERE type IN ('council-gate','fix-proposer','feature-designer') AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL ORDER BY type;"

# 2. Dump live, then diff STRUCTURALLY against your payload — not with `diff`,
#    which reports jsonb key-order noise. Assert the difference set is exactly
#    the prompt strings you intended and nothing else.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT default_config FROM agent_definitions WHERE type='feature-designer' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > /tmp/live.json
# then a recursive walk comparing live vs payload, printing (path, kind, len->len)
```

**Gotcha, and the load-bearing one:** also assert `live == <your ROLLBACK file>`
before applying. `ROLLBACK=1` restores a *file*, not a snapshot of what was
actually there — if live has drifted from that file, the "rollback" silently
writes a third state. It held on 2026-07-27 (`live == SEATED.json` → `True`) but
only because it was checked.

**Then check nothing is mid-run on the agent you are about to rewrite** — a
prompt change lands on the next step of an in-flight council:

```sql
SELECT correlation_id, owner_agent_type, current_step, status, created_at
FROM orchestration_states WHERE status='EXECUTING_STEP' ORDER BY updated_at DESC;
```

**Gotcha:** `current_step='review_editquality'` does NOT identify the agent —
that seat exists on all three councils. Use `owner_agent_type`. And read
`now()` from the DB: it is **UTC**, local file timestamps are BST, so a run that
looks an hour stale may have started 38 seconds ago.

## Is a seat actually REACHABLE? (not just present in the steps map)

A seat can exist, be correctly configured, and never run. Walking `next_step`
alone stops dead at the first `conditional` and reports a false orphan — traverse
`then_step`/`else_step`/`error_step` too, breadth-first from `workflow.start_step`.
On 2026-07-27 that gave 24 of 24 steps reachable, no orphans, and the chain
`review_guidelines → review_architecture → review_guardian → council_decide`.

Also check for a relevance gate: the 16-seat gate's seats fire only when edited
paths match a footprint, but `review_architecture` on `feature-designer` is an
**unconditional step**, so it runs on every design run that clears the approval
gate. A seat with 0 reviews is a rate limit or a gate — distinguish them before
concluding it is broken.

## Who owns a WORK ITEM? (`who-owns.py` does not cover this)

`scripts/who-owns.py` resolves a bug number or slug. There is no equivalent for a
`site_work_items` row, and for work items the ownership evidence is often **inside
the `spec` jsonb** where no repo grep reaches. `status='deferred'` is not
abandonment — it is where a lane parks an item between council rounds.

```sql
-- read the row itself, not its status, not a docs grep
SELECT jsonb_pretty(spec) FROM site_work_items WHERE id='<uuid>';
-- and the artefact trail for any correlation the spec names
SELECT correlation_id, kind, created_at, metadata->>'decision'
FROM diagnosis_artifacts WHERE correlation_id::text LIKE '<prefix>%' ORDER BY created_at;
```

Look for `ROUND N` / `REVISION REQUIRED` markers and a `prior_round` key. Which
`capability_gap` specs could even reach a design council:

```sql
SELECT id, status, left(summary,60),
       spec ? 'owner_approval' AS appr, spec ? 'code_pointers' AS ptrs
FROM site_work_items WHERE item_type='capability_gap' ORDER BY created_at DESC;
```

**Gotcha:** the `feature-designer` trigger
(`fixloop_eg_dartsonline/0NN_TRIGGER_feature_designer_v1.sh`) publishes via
`kubectl run -i --rm ... kcat -P <<JSON`, which is the **known silent-drop
shape** — it can exit 0 having published nothing. Verify by DB row
(`orchestration_states` for the printed `ORCHESTRATION_ID`), never by exit code.

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
