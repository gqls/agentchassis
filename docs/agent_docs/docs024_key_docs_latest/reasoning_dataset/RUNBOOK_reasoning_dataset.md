# RUNBOOK — reasoning-dataset extraction

*Operational commands for this workstream. Keep current: when a query changes,
change it HERE, not in your scrollback. Every query below is read-only.*

DB access (CLAUDE.md convention):

```bash
PSQL='kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db'
```

> **`-i` is required** when SQL arrives on **stdin** (heredoc / pipe), and
> unnecessary with `-c`. A psql that receives no input **exits 0 and prints
> nothing** — it fails silently, in exactly the way this workstream's bugs do.
> Verify the output, never the exit code. (Cost a silent no-op fix on 016.)

> **`kubectl exec -i` consumes the enclosing loop's stdin.** Any `for` loop over
> correlation ids that calls it needs `< /dev/null` on the kubectl call, or the
> loop stops dead after one iteration (`098_REPORT_unreviewed_commits_v1.sh:63`).

---

## 1. Corpus census — run this first, every session

Volumes move. Re-run before trusting any figure in PLAN or NOTES.

```bash
$PSQL -c "
SELECT kind, count(*) AS rows, count(DISTINCT correlation_id) AS corrs,
       min(created_at)::date AS first, max(created_at)::date AS last
FROM diagnosis_artifacts GROUP BY kind ORDER BY 2 DESC;"
```

```bash
$PSQL -c "
SELECT step_name, count(*) AS rows,
  count(*) FILTER (WHERE success) AS ok,
  count(*) FILTER (WHERE output_tokens IS NOT NULL AND max_tokens IS NOT NULL
                     AND output_tokens >= max_tokens) AS trunc_old,
  count(*) FILTER (WHERE NOT success AND error_message LIKE 'response truncated%') AS trunc_new,
  count(DISTINCT model) AS models
FROM llm_call_log
WHERE step_name ~ '^(verdict|review_|propose|repropose|reframe|council|escalat)'
   OR agent_type IN ('diagnose-agent','fix-proposer')
GROUP BY step_name ORDER BY 2 DESC;"
```

Trajectory count and reasoning depth:

```bash
$PSQL -c "
SELECT count(*) AS orchs_with_verdict,
  sum(jsonb_array_length(COALESCE(collected_data->'route'->'diagnose_state'->'trail','[]'))) AS trail_steps
FROM orchestration_states WHERE collected_data ? 'verdict';"
```

Baseline as of 2026-07-18: 38 bundles / 43 fix_plans / 43 council_reports /
5 escalations across 13 correlations; 445 reasoning rows; 26 orchestrations;
**79 trail steps**; 6 truncated rows.

---

## 2. Corpus health checks (the three that gate quality)

### 2a. `<no value>` injection — the poisoned-input check

```bash
$PSQL -c "
SELECT step_name, count(*) AS total,
       count(*) FILTER (WHERE prompt_rendered LIKE '%<no value>%') AS blanked
FROM llm_call_log
WHERE step_name ~ '^(verdict|review_|propose|repropose|reframe)'
GROUP BY 1 ORDER BY 3 DESC;"
```

Baseline: `repropose` 19/19, `review_debug_historian` 13/13, `reframe` 2/2,
`propose` 1/16, `verdict` 1/89. **Any row with a blank is quarantined**
(`input_complete: false`, `exclude_reason: "no_value_injection"`).

To see *which* sections blanked in a given call:

```bash
$PSQL -At -c "
WITH p AS (SELECT prompt_rendered pr, orchestration_id oid FROM llm_call_log
           WHERE step_name='repropose' ORDER BY created_at DESC LIMIT 1)
SELECT oid || ' ||| ' || l FROM p, unnest(string_to_array(pr, chr(10))) AS l
WHERE l LIKE '## %' OR l = '<no value>';"
```

### 2b. Pre/post-fix boundary — **join to the RUN, not the step**

The 016 render fix landed at **2026-07-18 13:15:11Z**. A run that started before
it carries pre-fix config no matter when its later steps logged.

```bash
$PSQL -c "
SELECT l.created_at AS step_at, o.created_at AS run_started,
       (o.created_at > '2026-07-18 13:15:11+00') AS post_fix, l.orchestration_id
FROM llm_call_log l
LEFT JOIN orchestration_states o ON o.orchestration_id::text = l.orchestration_id
WHERE l.step_name='repropose' ORDER BY l.created_at DESC LIMIT 10;"
```

> **This is the trap that fooled us once.** Steps at 13:17Z and 13:24Z post-date
> the fix and are still pre-fix work — their run started 13:11Z. Never grade a
> config fix on the step timestamp.

As of 2026-07-18 14:15Z: **no repropose run has started post-fix.** The first one
that does is the proof of the 016 fix — check it.

### 2c. Live seat coverage — is the reviser still half-blind?

```bash
# seats seeded
$PSQL -At -c "
SELECT count(*) FILTER (WHERE k LIKE 'review_%'), string_agg(k, ', ') FILTER (WHERE k LIKE 'review_%')
FROM agent_definitions ad, jsonb_object_keys(ad.default_config->'workflow'->'steps') k
WHERE ad.type='fix-proposer' AND ad.deleted_at IS NULL;"

# seats the repropose prompt actually references
$PSQL -At -c "
SELECT DISTINCT m[1] FROM agent_definitions ad,
  LATERAL regexp_matches(ad.default_config::text, '(\{\{\.review_[a-z_]*[^}]*\}\})', 'g') AS m
WHERE ad.type='fix-proposer' AND ad.deleted_at IS NULL ORDER BY 1;"
```

Baseline 2026-07-18: **13 seeded, 6 referenced, 7 invisible.** If the second
query returns anything ending `.result}}`, the 016 render bug has been
**re-seeded** — say so in NOTES and tell the fixloop thread immediately.

Note `default_config->'workflow'->'steps'` is a **JSON object keyed by step name**,
not an array. `jsonb_array_elements` on it errors with *"cannot extract elements
from an object"* — use `jsonb_object_keys` / `->'<step name>'`.

---

## 3. Extraction (Phase 1 — not yet built)

Runs **outside** the cluster. Do not run it as a pod:
`platform/orchestration/actions/training_data_export.go:3-8` records that
file-writing "landed on ephemeral chassis pods" and was retired for it.

```bash
# intended shape once cmd/reasoningset exists
$PSQL -At -f cmd/reasoningset/extract.sql | go run ./cmd/reasoningset \
  --labels docs/agent_docs/docs024_key_docs_latest/reasoning_dataset/LABELS_benchmark.json \
  --out reasoning_v1.jsonl
```

Fetch one artifact body by hand (the documented route,
`090_TRIGGER_needs_diagnosis_v1.sh:364-371`) — used to byte-check `input_state`:

```bash
$PSQL -t -A -c "SELECT body FROM diagnosis_artifacts
  WHERE correlation_id='<CORR>' AND kind='bundle' AND iteration=1" > bundle_iter1.md
```

---

## 4. Verifying an extraction run

1. **Counts reconcile** — JSONL records per task/model must equal §1. A mismatch
   means a join dropped rows, most likely a uuid/text cast:
   `orchestration_states.correlation_id` is `uuid`,
   `diagnosis_artifacts.correlation_id` is `text`,
   `llm_call_log.correlation_id` is `varchar(255)`.
2. **Two trajectories hand-checked** end-to-end, one CONFIRMED and one
   guard-degraded: emitted `input_state` must byte-match the bundle body from §3.
3. **Guard block** — find a verdict where raw ≠ coerced; `guard.diagnostic` must
   match the sentence prepended to `NeededEvidence` in the trail, and
   `guard.tripped` must be false wherever raw == coerced.
4. **Blinding** — `grep -c 'fixloop_eg_dartsonline\|RUBRIC_\|NOTES_running' reasoning_v1.jsonl`
   must be **0**. The fixloop docs are excluded from the loop's own input so the
   benchmark stays honest; leaking them into an eval set leaks the answers.
5. **Exclusion audit** — every `input_complete: false` row carries a non-null
   `exclude_reason`, and the `no_value_injection` count equals §2a (36).

---

## 5. Gotchas (each one cost someone real time)

- **`orchestration_states.last_activity` is `timestamp WITHOUT time zone`** while
  `created_at` is `timestamptz`, on a BST host. Interval arithmetic across them is
  silently wrong by the UTC offset. `llm_call_log.created_at` and
  `diagnosis_artifacts.created_at` are both `timestamptz` — stay in those two.
- **`LogLLMCall` is fire-and-forget** with a 5s timeout
  (`platform/orchestration/actions/llm_call_logger.go:34`). Rows can be missing.
  Never assume 1:1 with workflow steps; a missing row is absent evidence, not zero.
- **Trail entries have no JSON tags** (`pkg/diagnose/loop.go:79-153`) — Go field
  names and **integer enums**: `Outcome` 0=Unverifiable/1=Confirmed/2=Refuted,
  `Tier` 0=static/1=state/2=runtime. `collected_data.verdict.result` is the
  snake_case wire form. Both shapes must be handled; the trail is the **coerced**
  verdict, `verdict.result` the **raw** one.
- **`diagnosis_artifacts.correlation_id` collapses a diagnosis run and its
  fix-proposer run onto one key.** 13 correlations, 43 council_reports — use
  `orchestration_id` to disambiguate re-runs.
- **Council-gate submissions mint their own `SUBMISSION_CORR` with no diagnosis
  behind them.** They appear in `diagnosis_artifacts` with no bundle. Tag them as
  a different task type; they are not broken trajectories.
- **`iteration_note` is a declared `kind` that no Go code ever writes.** Empty
  partition, not missing data.
