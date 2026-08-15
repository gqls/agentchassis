# RUNBOOK — 209 purpose-keyed source

Every command that was hard to get right, with its gotcha attached. Fix them
here, not in scrollback.

---

## 1. Which live definitions carry the action (a DB fact, never a repo fact)

A Go call-site count is not the caller count — 221's lane was corrected by the
council for exactly that. Ask the database:

```sql
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text LIKE '%deploy_image_asset%'
ORDER BY type;
```

⚠ A bare token (`deploy_image_asset`) is safe in `::text LIKE`. A **key/value**
pair is not — jsonb renders a space after the colon, so `'%"k":"v"%'` matches
nothing. Induce a non-zero result before trusting a zero.

## 2. Enumerate a definition's steps (steps is an OBJECT, not an array)

```sql
SELECT ad.type, s.key AS step, s.value->>'action' AS action,
       COALESCE(s.value->'config'->>'purpose','-')            AS cfg_purpose,
       COALESCE((s.value->'config'->'input_fields')::text,'-') AS input_fields
FROM agent_definitions ad,
     LATERAL jsonb_each(ad.default_config->'workflow'->'steps') s
WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL
  AND ad.type = '<type>'
ORDER BY s.key;
```

Use `jsonb_pretty(s.value)` for the full step when you need `next_step` /
`error_step` / conditional routing — branch exclusivity is invisible in a flat
listing, and that exclusivity is what decided this bug.

## 3. Read a run's `collected_data` shape — enumerate keys, never probe a path

A path read cannot see a shape change underneath it. Ask which keys exist:

```sql
SELECT k, count(*) AS n_runs
FROM orchestration_states o, LATERAL jsonb_object_keys(o.collected_data) k
WHERE o.owner_agent_type = 'asset-deployer'
GROUP BY k ORDER BY n_runs DESC, k;
```

⚠ The agent-type column is **`owner_agent_type`**, not `persona_id`.

## 4. Before reading anything into an ABSENCE in `orchestration_states`

The table looks like it holds a month (oldest row 2026-07-13) — it does not.
Bucket by status first:

```sql
SELECT CASE WHEN created_at > now() - interval '24 hours' THEN 'last_24h'
            WHEN created_at > now() - interval '7 days'  THEN '1-7d'
            ELSE 'older_than_7d' END AS age,
       status, count(*)
FROM orchestration_states GROUP BY 1,2 ORDER BY 1,3 DESC;
```

Measured 2026-08-08: **13** `COMPLETED` rows older than 24h, **0** older than 7
days; the old tail is `CANCELLED`/`RUNNING`/`INITIALIZED`. So for completed runs
this is a **~24-hour window** — "no rows" means "not today", nothing more.

## 5. `llm_call_log` is BLIND to these agent types — check the control first

It retains from 2026-03-25 (50,861 rows), which makes it look like the right
place to ask "has this agent ever run". It returned **0 rows for
`asset-deployer`** on a day it ran 16 times. Always run the positive control:

```sql
SELECT agent_type, count(*), max(created_at)::date
FROM llm_call_log
WHERE agent_type IN ('<the type you care about>', '<a type you KNOW ran today>')
GROUP BY 1;
```

If the known-good type returns 0, the measurement is blind — discard it, do not
report it as an absence.

## 6. Run the 209 evidence harness

```bash
go test ./platform/orchestration/actions/ \
  -run 'TestFindStorageURI_|TestExtractActionInputs_LegacyShape|TestDeployImageAsset_LegacyShape' -v
```

The `AssetIDIsNotStable` test runs the real helper 400× and prints the split. It
is **not** a flake if the numbers move — Go map order is randomised per run, so
the ratio varies (344/400 hero-wins when first measured). It *would* be a real
finding if the split ever collapsed to a single value: that means resolution
became deterministic and the hazard changed.

## 7. Applying a definition-editing migration WITHOUT taking other threads' pending files

`run-migrations.sh --apply` takes **every** pending file in `MIGRATIONS_DIR`
(default `docs/agent_docs/sql_for_agents`) — there is no per-file apply flag, and
the pending set usually contains other lanes' work. Scope with the env var:

```bash
SCRATCH=<scratchpad>/migNNN && mkdir -p $SCRATCH
cp docs/agent_docs/sql_for_agents/NNN_my_migration.sql $SCRATCH/
MIGRATIONS_DIR=$SCRATCH ./scripts/migration/run-migrations.sh            # dry-run (doomed txn)
MIGRATIONS_DIR=$SCRATCH ./scripts/migration/run-migrations.sh --apply   # apply + record
```

⚠ Before the apply, **induce the post-verify**: run the migration's final DO
block standalone against the *unmigrated* rows — it must RAISE (0/N). A verify
that has never failed proves nothing (`SELECT`s cannot stop a COMMIT; only
DO/RAISE can, and only if it CAN fire). 348's induction raised "0 of 4" — that
zero is what made the later "4 of 4" evidence.

## 8. Post-roll config re-verification (the deploy-time stamp lies)

Every deploy re-stamps `updated_at` on ~all active `agent_definitions` (175 rows
at 08:49:01 on 08-09) **without changing content** — measured control: migration
341's `gate_next_item` step survived that stamp. So after any roll:

- Never conclude "changed" or "unchanged" from `updated_at`.
- Re-read the four deploy steps **by content** (§2 query) and diff against the
  348 shape: dotted `{p}_stored.*` paths, `domain: site_record.domain`,
  `input_fields: ["purpose","domain","asset_id"]`, **no `uri_field`**.

## 9. Ownership re-check (the `who-owns.py` false positive)

```bash
./scripts/who-owns.py 209
```

⚠ It reads **commits**, so a lane that merely *cites* a bug in a handoff reads as
owning it, and a session mid-fix is invisible. Sweep live transcripts too:

```bash
ls -t ~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl | head -16 | while read f; do
  c=$(grep -c "findStorageURI\|209_HANDOFF" "$f" 2>/dev/null); [ "$c" != "0" ] && echo "$c $(basename $f)"; done
```

Then check each hit's *last* entries for what it is actually doing — most hits
are incidental (a `git status` listing the filename, or a neighbouring lane
reading the same package).

## 10. Running the behavioural proof (a real build on a sacrificial domain)

`fire_209_proof.sh pageflow|swo` and `verify_209_proof.sh`, both in this directory.

```bash
# 0. Seed the sacrificial site row FIRST (the workflow needs a brief to plan from).
#    sites has NO site_id column — the PK is `id`. content_data is an object here;
#    on some older sites it is an ARRAY, so jsonb_object_keys() errors on those.
#    Leave github_repo NULL: resolveGitRepoNameDB then defaults to 'sites'.

# 1. Attach the log capture BEFORE dispatching. The per-agent pod keeps ~11 SECONDS.
#    Only agent-chassis is a Deployment; agent-pageflow-builder-* is spawned per run,
#    so poll for it rather than naming it:
until p=$(kubectl -n ai-persona-system get pods --no-headers -o name 2>/dev/null \
          | grep -m1 agent-pageflow-builder); [ -n "$p" ]; do sleep 1; done
kubectl -n ai-persona-system logs -f "$p" --tail=0 > pfb.log &

# 2. Dispatch. No PUBLISH_OK in the output => nothing was published; re-fire.
./fire_209_proof.sh pageflow

# 3. Follow it. Sub-agents share the correlation, so one query shows the whole tree:
#    SELECT owner_agent_type, current_step, status FROM orchestration_states
#    WHERE correlation_id='<CORR>' ORDER BY created_at;
```

⚠ **Assert the right thing.** "hero.* and logo.* differ in bytes" is NOT
disconfirmable — the deploy re-encodes per purpose (hero→jpg, logo→png), so they
differ even when the wrong source is fetched. Assert instead:

- the **downloaded object key** matches that step's own `{p}_stored.s3_uri`
  (`grep 'Downloading image from S3' pfb.log`), and
- the **asset row stamped** is that step's own `asset_id`
  (`grep 'recorded local url on asset' pfb.log`), and
- for the logo specifically, the artefact is a **400×400 PNG** — a `.jpg`, or any
  other size, means purpose resolved to "hero" (bugs_open/231).

⚠ The **hero** step is the one worth reading: it is the only deploy that runs with
both assets in `collected_data`. The logo deploys before the hero exists.

## 8. Migration 380 — blast-radius queries for a dispatch-mapping change, and the re-drive

Who binds the top-level key you are about to inject (run BEFORE adding any
`input_mapping` line — the whole blast radius is definitions that read it):

```sql
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text LIKE '%input_data.purpose%';
-- 2026-08-11: exactly asset-deployer, image-build-handler
```

Which item types even carry the spec field (the population the mapping
touches; `?` mapping skips the rest silently):

```sql
SELECT DISTINCT item_type, handler_agent FROM site_work_items WHERE spec ? 'purpose';
```

Gotcha: `workflow.steps` is a jsonb OBJECT keyed by step name, not an array —
`jsonb_array_elements` fails with "cannot extract elements from an object";
use `jsonb_each` or a `#>` path.

Re-drive a completed undeployed_asset item (the webdesign.uk-proven path —
the build-pipeline-trigger picks it up within ~2 min):

```sql
UPDATE site_work_items SET status='triaged', updated_at=now()
WHERE id='<item>' AND status='complete';
```

Verify at the artefact, in this order: work item `result` →
`commit_message` must name the right purpose ("Deploy logo image", not
"Deploy hero image") · asset row `url`/`filename` restamped · served object
(`curl -sI`) content-type + `file` on the bytes (the purpose's extension AND
dimension class AND alpha channel for png).

## 11. The 231 census — run, read, and the trap in reading it (added 2026-08-11)

```bash
# The whole census, live fleet, human-readable (exit 1 = dead mismatched exist):
./scripts/audit-default-shadowed-keys.sh
# Machine-readable, for joins:
./scripts/audit-default-shadowed-keys.sh --json > census.json
# Arm 3 (every spec's Defaults, from the BINARY — a source grep hits 223's
# var-blindness):
go run ./cmd/config-key-audit --specs | jq 'with_entries(select(.value.defaults))'
```

Gotchas that cost time on 2026-08-11:

- **A dead-mismatched finding is NOT yet damage.** 20 of the first run's 24
  were honoured anyway, because the action reads `config[...]` directly in its
  body — the finding only asserts the ExtractActionInputs path. Before
  asserting damage, grep the ACTION for the key and find the read line:
  `grep -n '"<key>"' platform/orchestration/actions/<action file>` — a
  `GetIntField(config, …)` / `config["k"].(type)` read means honoured; an
  `inputs.Get("k")` read means the shadow is real (that is how the four
  `audit_source` findings were separated from the other 20).
- The read-path table in `bugs_open/231` (2026-08-11 section) is a
  **point-in-time census** — re-verify at the read line before reusing it.
- `dotted_conditional` findings on `*_field` keys (append_doc_note,
  write_doc_plan, the diagnose family) are extractor-IRRELEVANT — those
  actions read the `*_field` key from config directly and resolve the path
  themselves. Do not read them as arm-2 exposure.

---

## Prompt caching (LCO-008) — the four commands worth keeping (2026-08-15)

### 1. Read an agent's EFFECTIVE `max_tokens` — never the key you are about to write

The resolver is `ai_actions.go`: `agentConfig["max_tokens"]` (**the TOP LEVEL of
`default_config`** — `agentConfig = agentDef.DefaultConfig`, *not* the step's `config`, despite
the name) → merged `ai_service` (root overlaid by step `config.ai_service`) → the client's
hardcoded `2048`. **`...steps.<step>.config.max_tokens` is INERT** and a migration that writes
it will pass a guard that reads it back. Show all four so a future divergence cannot hide:

```sql
SELECT default_config->>'max_tokens'                                                    AS top_level,
       default_config->'ai_service'->>'max_tokens'                                      AS root_ai,
       default_config->'workflow'->'steps'->'plan_gaps'->'config'->'ai_service'->>'max_tokens' AS step_ai,
       default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'max_tokens'        AS step_cfg_INERT
FROM agent_definitions
WHERE type='content-gap-planner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

⚠ **Gotcha:** a pre-condition NOTICE reading `max_tokens=(unset)` is the tell that you are
looking at the wrong path — not that the value is absent estate-wide.

### 2. Is the TTL discriminator even able to fire? (the demand control)

A zero from the >5min-gap query is meaningless without this. It answers "could this query have
returned a row at all yet", which is a different question from "did it".

```sql
WITH c AS (
  SELECT agent_type, created_at, cache_read_input_tokens,
         lag(created_at) OVER (PARTITION BY agent_type, md5(left(prompt_rendered,4000))
                               ORDER BY created_at) AS prev
  FROM llm_call_log
  WHERE created_at > '<cutoff>' AND prompt_rendered IS NOT NULL
)
SELECT count(*) AS calls, count(prev) AS repeat_pairs,
       count(*) FILTER (WHERE created_at - prev > interval '5 minutes') AS pairs_over_5min,
       max(created_at - prev) AS widest_gap
FROM c;
```

⚠ **Scope it to agents that actually carry a marker.** Unmarked agents cannot produce a cache
read, so including them guarantees zeros that say nothing about the TTL:
`SELECT agent_type, sum(coalesce(cache_read_input_tokens,0)) FROM llm_call_log WHERE created_at > … GROUP BY 1;`

**And validate the discriminator against pre-roll history before trusting it** — under the 5m
TTL a >5min gap must force a WRITE and yield NO read. Measured 2026-08-12 → 08-15 on
`council-gate`: 29 such gaps, **0 reads, 28 writes**. If that ever comes back with reads, the
check is broken and every conclusion drawn from it is void.

### 3. Find the shared/varying boundary before placing a marker

Do not reason about it from the template — the template is ~4k chars and renders to ~15k, so
the boundary is where the *rendered* text stops being shared. `distinct_prefixes_at_boundary`
must be **1** per group, which is the exact precondition prefix caching needs:

```sql
SELECT strpos(prompt_rendered,'<ANCHOR>') AS boundary, length(prompt_rendered) AS total,
       count(*) AS calls,
       count(DISTINCT left(prompt_rendered, strpos(prompt_rendered,'<ANCHOR>')-1)) AS distinct_prefixes
FROM llm_call_log
WHERE agent_type='<agent>' AND prompt_rendered IS NOT NULL
  AND created_at > now() - interval '3 days' AND strpos(prompt_rendered,'<ANCHOR>') > 0
GROUP BY 1,2 ORDER BY calls DESC;
```

### 4. Applying a migration when the runner cannot reach it

`run-migrations.sh --apply` walks **every** pending file in order and a failure stops the run.
As of 2026-08-15 there are 17 pending from other lanes and it halts at `324`, whose guard
**refuses by design** unless a specific `-c` setting is prepended — so it never reaches the 400s.
Apply out-of-band, then register:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/<file>.sql

./scripts/migration/run-migrations.sh --record-only <file>.sql --note '<what you verified>'
```

⚠ Feed the file on **stdin** (or `-f`), never paste — pasting mangles comments and
dollar-quoted `DO $$` bodies. `-v ON_ERROR_STOP=1` is what makes a failed guard abort rather
than commit the rest.

### 5. Verify a marker fired — and know which zero is which

```sql
SELECT created_at, model, input_tokens, output_tokens,
       coalesce(cache_creation_input_tokens,0) AS writes,
       coalesce(cache_read_input_tokens,0)     AS reads
FROM llm_call_log
WHERE agent_type='<agent>' AND created_at > '<when the marker went in>'
ORDER BY created_at;
```

- `writes > 0, reads = 0` on the **first** call is correct — there was nothing to read.
- `writes > 0, reads = 0` **persisting** across calls is the silent failure: every call is
  paying the write premium and reading nothing, which is worse than no caching.
- `output_tokens` at or near the configured cap means the completion was **CUT**.
- ⚠ `llm_call_log` stores only **totals** — there is no 5m/1h bucket breakdown, so the log can
  never tell you which bucket a write landed in. The only log-based TTL proof is behavioural:
  a read at a gap wider than 5 minutes.
