# RUNBOOK — bugfix 277 (commands, with their gotchas attached)

## Census / dry run (read-only, idempotent)

```bash
cd docs/agent_docs/docs024_key_docs_latest/bugfix_277_required_fields_repair
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -f - < CENSUS_2026-08-15_predicted_routes.sql
```
Gotcha: the census re-implements the seed's classifier readably. When either changes, change
BOTH, and re-prove the seed's own string (below) — the census passing proves the census, not
the seed.

## Prove the seed's exact embedded SQL against real rows (before apply / after any edit)

Extract the query FROM THE SEED FILE (never retype it), unescape `''`→`'`, substitute
`$1::uuid`/`$2::uuid` with a real (site_id, work_item_id), run via psql. The 2026-08-15 run of
exactly this proved all five canary candidates route as the census predicts (NOTES).

## Apply the seed (DB config, live instantly, INERT — 0 rows assigned)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/410_required_fields_missing_router.sql
```
Gotcha: the verify block RAISEs (and aborts the COMMIT) on any mis-wired branch, a retired
`spec` key, a non-triaged conversion status, or >0 rows already assigned. A re-run snapshots
first (conditional DO block) and is safe.

## Canary assignment (the first execution — treat any surprise as a stop)

```sql
UPDATE site_work_items
   SET handler_agent = 'required-fields-missing-handler',
       status = 'triaged', attempt_count = 0, updated_at = NOW()
 WHERE item_type = 'required_fields_missing'
   AND status IN ('needs_human_review','unresolved')
   AND COALESCE(handler_agent,'') = ''
   AND left(id::text,8) IN ('332bb3f6','4fa5b019','e512af8a','483fb749');
-- expect UPDATE 4
```
Gotchas: `attempt_count = 0` is load-bearing (claim gate requires attempt_count <
max_attempts). Dispatch cadence is 120s; no dispatch within ~300s of a chassis pod restart.

## Verify the canary (per arm)

```sql
-- each canary row: status + recorded route
SELECT left(id::text,8), status, result->'response'->'triage'->>'route' AS route,
       result->>'route' AS route2, left(COALESCE(error,''),60) AS err
FROM site_work_items WHERE left(id::text,8) IN ('332bb3f6','4fa5b019','e512af8a','483fb749');
-- expected: 332bb3f6 complete/stale · 4fa5b019 complete/converted (+ a new content_rewrite
-- row, source='required-fields-missing-handler', spec->>'mode'='edit_live') ·
-- e512af8a needs_human_review with the blob error message · 483fb749 needs_human_review
-- with the owned message
-- conversions:
SELECT item_type, status, item_key FROM site_work_items
WHERE source = 'required-fields-missing-handler';
-- blob dedup key still held (exactly 1 non-terminal row on the key):
SELECT count(*) FROM site_work_items
WHERE item_key = (SELECT item_key FROM site_work_items WHERE left(id::text,8)='e512af8a')
  AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled');
```
> **CORRECTED 2026-08-15, from the canary itself:** for COMPLETED rows the route is NOT on the
> row at all. The loop's mark_complete overwrites `result` with the SPAWN bookkeeping
> (role/topics/agent_id — measured on `332bb3f6`), replacing both the close arm's
> `result_fields` and any saga response. The audit trail for closed rows lives in
> `orchestration_states`:
> ```sql
> SELECT left(orchestration_id::text,8), status, current_step,
>        collected_data->'triage'->>'route' AS route
> FROM orchestration_states
> WHERE workflow_plan->>'start_step'='classify' ORDER BY created_at;
> ```
> For PARKED rows the loop's complete no-ops (guard excludes needs_human_review), so the route
> IS on the row: `result->>'route'` + the message in `error`. The canary verified all three
> executed arms this way: `0177ce18` stale · `61a71bbd` no_content_data · `8dd51e7e`
> no_plan_owned (the gas converter), each COMPLETED at `done` with correct facts.

## Fleet assignment (after the canary verifies)

Same UPDATE without the id filter. Then the after-state:

```sql
SELECT status, COALESCE(result->'response'->'triage'->>'route', result->>'route') AS route, count(*)
FROM site_work_items WHERE item_type='required_fields_missing' GROUP BY 1,2 ORDER BY 3 DESC;
```

> **SETTLED 2026-08-15 ~14:00Z — the producer change is ALREADY LIVE.** Another lane's roll to
> `v1.0.1302` carried commit `5ad81182b` (stamp `194907d5b…`, `git merge-base --is-ancestor`
> exit 0; literal probe 1 with negative control 0). **Replica coverage settled by the
> uniform-image observation**: all 25 pods running the agent-chassis image (15 Running, 10
> Succeeded job pods) carry the SAME `v1.0.1302` — one probe speaks for all, and this is the
> honest answer to the "-l app=agent-chassis is not every pod running the binary" landmine
> (enumerate by IMAGE, then check tag uniformity, before trusting any single-pod probe).
> The PBP-028 edit_live channel was probed the same way: `grep -ac 'attached current content
> for edit mode' /proc/1/exe` → 1, negative control 0.

## Post-roll (after the producer Go change ships in a chassis image)

```bash
# whose commit is the service running? (startup line scrolls; binary probe is durable)
kubectl -n ai-persona-system exec <chassis-pod> -- \
  sh -c "grep -oa 'buildinfo.GitCommit=[0-9a-f-]*' /proc/1/exe | head -1"
git merge-base --is-ancestor <producer-commit> <stamp>   # exit 0 = shipped
# belt-and-braces literal probe: 'required-fields-missing-handler' enters the BINARY
# only via the Go const (the seed is DB-side), so on EVERY replica:
kubectl -n ai-persona-system exec <pod> -- sh -c \
  "grep -ac 'required-fields-missing-handler' /proc/1/exe"        # expect >=1
kubectl -n ai-persona-system exec <pod> -- sh -c \
  "grep -ac 'zzq-negative-control-not-a-handler' /proc/1/exe"     # expect 0
```
Then re-run the fleet assignment UPDATE once — items the OLD producer filed between
assignment and roll are born parked/unassigned and need sweeping in.

## Churn guard (+7 days)

```sql
SELECT count(*) FROM site_work_items
WHERE item_type='required_fields_missing' AND status='unresolved'
  AND created_at > '<fleet-assignment-time>';
-- expect ~0; anything else means a close arm is churning against the producer
```

## Council

Submission corr `7b0e2833-715f-4a9a-897b-efd913073582`. Verdict:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='7b0e2833-715f-4a9a-897b-efd913073582' AND kind='council_report'
ORDER BY created_at;
```
Gotcha: publish→run start measured at 29 min under normal load — a missing orchestration row
is latency, not a dropped dispatch; find the run by payload, never re-trigger on absence.

## Rollback

`docs/agent_docs/sql_for_agents/410_required_fields_missing_router_ROLLBACK.sql` — refuses
while non-terminal rows still route at the handler (un-assign first; header has the UPDATE).
If the producer Go change has shipped, roll that back too or items go 'blocked' at claim
(bugs_closed/077 shape).

## Apply ONE migration when other lanes have pending files (the only safe route here)

`./scripts/migration/run-migrations.sh --apply` takes **every** pending file in
`docs/agent_docs/sql_for_agents/`, and on a tree this many sessions share that is routinely four or
five other lanes' work (2026-08-18: `462_fixer_rerenders_skip_owned_pages`, `467`, `468`, `470`).
There is no `--only <file>` flag. So apply the single file yourself, then register it:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/<NNN_name>.sql

./scripts/migration/run-migrations.sh --record-only <NNN_name>.sql \
  --note 'applied out of band by single-file psql; <what the controls proved>'
```
`--record-only` takes a **bare filename**, not a path, and is mutually exclusive with `--apply`.
Gotchas: the file must carry its own `BEGIN`/`COMMIT` (psql only wraps `-f` in a transaction if the
file says so), and its guard `DO` block must `RAISE` rather than `SELECT` — `ON_ERROR_STOP` ignores a
non-empty result set, so a verify block made of `SELECT`s cannot stop the `COMMIT`.

## Prove an edit to a live `pre_query` still parses, WITHOUT running it

`pre_query` bodies here are `UPDATE`-in-CTE statements, so "run it to see if it parses" mutates rows.

```sql
EXECUTE 'EXPLAIN ' || new_q;   -- inside the migration's DO block
```
**EXPLAIN plans without executing.** It catches the realistic failure — an apostrophe left undoubled
inside prose that is nested in a SQL string literal — and mutates nothing. Pair it with an occurrence
count on the anchor you are replacing (`(length(q)-length(replace(q,a,'')))/length(a) = 1`) so the
edit cannot silently land on text another session has since changed.

## Apply exactly ONE migration — a better way than the section above (2026-08-18)

> **CORRECTION to "there is no `--only <file>` flag", above.** True as written, and the
> hand-apply-then-`--record-only` recipe still works — but it makes recording a **separate human
> act** that is easy to forget, and an applied-but-unrecorded migration reads as pending to the next
> session's dry run. There is a scoped path that records automatically.

`MIGRATIONS_DIR` is the runner's own env override, and it can point anywhere. Give it a directory
holding **only your file** and `--apply` cannot reach another lane's work, because the runner never
sees it:

```bash
S=$(mktemp -d)                        # or the session scratchpad
cp docs/agent_docs/sql_for_agents/480_owned_page_refusal_is_not_a_handler_failure.sql "$S/"

MIGRATIONS_DIR="$S" ./scripts/migration/run-migrations.sh          # dry run: must list exactly 1
MIGRATIONS_DIR="$S" ./scripts/migration/run-migrations.sh --apply  # applies AND records it
```

You get the probe, the apply, and the `schema_migrations` row in one step, with the ledger keyed on
the real filename.

**Gotchas, all of them load-bearing:**

* **The assignment must be on the SAME line as the command.** `MIGRATIONS_DIR=…` on its own line is
  an ordinary shell assignment that the script's `${MIGRATIONS_DIR:-…}` default may not pick up, and
  the run then covers the whole repo directory — this is a `LANDMINES.md` entry in its own right.
* **Copy, do not move.** The file must stay in `docs/agent_docs/sql_for_agents/` for everyone else.
* **Sidecars are excluded anyway.** `_ROLLBACK.sql` / `_VERIFY.sql` match `SIDECAR_RE` and are never
  run by the runner, so copying only the migration is enough.
* **`--record-only` still takes a bare filename**, if you ever do need the hand-apply route.

Measured 2026-08-18: the unscoped dry run listed **15 pending files** from other lanes, two of them
probing *inconclusive* on live drift (`467`, `468`). Scoped, it listed one.

## Exercise a migration THREE ways before applying it

A dry run proves the SQL runs. It does not prove the guards can fire, and a guard that cannot fire
is not a guard. All three inside transactions that roll back:

```bash
# 1. the whole file, COMMIT swapped for ROLLBACK — expect: guard passes, UPDATE 1, verify NOTICE
sed 's/^COMMIT;$/ROLLBACK;/' <file>.sql | kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1

# 2. PRE-SET the state the guard refuses, then run the file — expect: the guard ABORTS
{ echo "BEGIN;"; echo "<UPDATE that sets the key/state>"; sed 's/^COMMIT;$/ROLLBACK;/' <file>.sql; echo "ROLLBACK;"; } \
  | kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1

# 3. PLANT the leak the negative control looks for, then run the file — expect: VERIFY ABORTS
#    (without this, a positive assertion passes identically on an UPDATE with no WHERE clause)
```

Both probes must print `ERROR:` **naming your own message**. If probe 2 or 3 succeeds, the guard is
decorative. Run them before the apply, not after — after, you cannot roll back.

## Tell an ownership REFUSAL from a genuine save FAILURE, post-roll

The Tier 1 change (`480` + `6aee22b00`) makes them different rows. Both controls, in one query — a
result with only the first line is equally consistent with the status write being broken:

```sql
SELECT status, count(*), bool_or(result ? 'owned_page_refusal') AS stamped
FROM site_work_items
WHERE handler_agent = 'page-build-handler'
  AND updated_at > '<the roll>'
  AND error LIKE '%OWNED_PAGE_GUARD%'
GROUP BY 1
UNION ALL
SELECT 'control: real save failures', count(*), bool_or(result ? 'owned_page_refusal')
FROM site_work_items
WHERE handler_agent = 'page-build-handler'
  AND updated_at > '<the roll>'
  AND status = 'failed' AND COALESCE(error,'') NOT LIKE '%OWNED_PAGE_GUARD%';
```
Expected: refusals `wont_fix` with `stamped = t`; the control non-zero, `failed`, `stamped = f`.
**A zero control means no genuine failures happened in the window, not that the split works** — widen
the window rather than reading it as a pass.

## The rendered_html_transform canary (post-roll, CQ-028 — run ONCE, it opens the promoter door)

> **✅ RUN 2026-08-21 13:21:42Z — DONE, do not run again.** Row `ecd947c2…` (tool-cubic-bezier)
> completed 13:25:03Z, verifier `verified`, proven at the served bytes (§"Prove one page's repair"
> below). The promoter released the other six on its very next tick (13:27), unaided. The pair
> `literal_markdown → section-editor` is known-good from here on, so **step 2 is history** — a
> future row of this pair is promoted automatically and needs no hand-promote.

```sql
-- 0. Preconditions: chassis stamp is a descendant of the build commit (ask the POD, not git),
--    and the config half reads at the column:
SELECT default_config #>> '{workflow,steps,apply_edit,config,allow_rendered_html_transform}' AS flag,
       default_config #> '{workflow,steps,apply_edit,config,input_fields}' @> '["transform_name"]'::jsonb AS whitelisted
FROM agent_definitions WHERE type='section-editor' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;   -- expect true | t

-- 1. Wait for the detector's sweep to file a NEW-SHAPE item (spec carries edit_type):
SELECT id, status, spec->>'page_name' AS page, spec->>'edit_type' AS edit_type
FROM site_work_items
WHERE item_type='literal_markdown' AND handler_agent='section-editor'
ORDER BY created_at DESC;
-- GOTCHA: until the roll, the sweep keeps filing OLD-shape items (handler page-rerender).
-- An old-shape open row HOLDS the dedup slot; it must reach a terminal status (the wont_fix
-- loop does this on its own) before a new-shape row can exist for that page.

-- 2. Promote exactly ONE (the promoter would refuse: 0 lifetime completes on the pair):
UPDATE site_work_items SET status='triaged', triaged_at=now(),
       spec=jsonb_set(COALESCE(spec,'{}'::jsonb),'{original_pipeline}',to_jsonb(pipeline)),
       pipeline='build', updated_at=now()
WHERE id='<the one row>' AND status='detected';
-- (that is the promoter's own UPDATE, copied from 444's pre_query — keep it in lockstep)

-- 3. Watch: item → complete, and the verifier note in result->'_verification'.
```

```bash
# 4. THE PROOF IS THE SERVED BYTES, never the status (cache-busted):
curl -s "https://webdesign.co.uk/<page_url>?cb=$(date +%s)" | grep -o '<code>[^<]*</code>'   # expect the token
curl -s "https://webdesign.co.uk/<page_url>?cb=$(date +%s)" | grep -c '`'                    # script literals MAY remain — compare against the pre-repair count MINUS 2 per converted span, not against zero
```

- GOTCHA: `transform_converted` in the item's step result is the count of spans EDITED, not findings
  CLOSED — the whole-page verifier is what closes the item, and it re-scans both surfaces.
- GOTCHA: the pair stays held until THIS canary's `complete` lands; do not promote a second row to
  "help" — one completion is the door, more is just risk.

### Post-roll liveness probe (run BEFORE the canary — the config DO/RAISE proves nothing about the binary)

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- grep -ac "rendered_html_transform" /proc/1/exe        # expect >=1
kubectl -n ai-persona-system exec "$POD" -- grep -ac "code_span_to_code_tag" /proc/1/exe          # expect >=1
kubectl -n ai-persona-system exec "$POD" -- grep -ac "OWNED_PAGE_GUARD" /proc/1/exe               # positive control, expect >=1
kubectl -n ai-persona-system exec "$POD" -- grep -ac "ZZQQ_NEEDLE_THAT_MUST_NOT_EXIST" /proc/1/exe # negative control, expect 0
```
- GOTCHA (standing): never `strings` (absent from the images), always run BOTH controls in the same
  breath, and remember a binary hit count is not a call-site count (Go dedupes string constants).
- And `git merge-base --is-ancestor af0f00bb5 <the pod's build-provenance stamp>` answers "did my
  commit ship" as a query, per BLD-019.

## Force a discovery sweep for ONE site (when the 7-day rotation would not reach it)

**Why you need this:** a discovery check (`literal_markdown`, `placeholder_contact`, …) is only ever
run by its rotation task, and `site-discovery-rotation-quality`'s `pre_query` selects
`LIMIT 1` site with `last_selected_at < now() - interval '7 days'`. When every site is inside its
7 days the rotation is **IDLE, not slow** — and `last_triggered_at` keeps advancing every 3h the
whole time, so the task reads healthy while examining nothing.

```sql
-- When would the rotation reach my site on its own? (the answer is often days)
SELECT s.domain, r.last_selected_at,
       r.last_selected_at + interval '7 days' AS eligible_at,
       now() - r.last_selected_at AS age
FROM site_discovery_rotation r JOIN sites s ON s.id = r.site_id
WHERE r.agent_type = 'quality-discovery-agent'
ORDER BY r.last_selected_at;   -- top row = the next site the rotation will take
```

```sql
-- Fire one sweep now. Precedent shape: oneshot-quality-discovery-wdcouk-20260810.
-- NO pre_query, so it does NOT consume the site's rotation stamp — the natural slot survives.
INSERT INTO scheduled_tasks
  (name, description, target_agent_type, target_topic, interval_seconds,
   input_data, concurrency_group, max_concurrent, timeout_seconds, fire_message, enabled)
VALUES
  ('oneshot-quality-discovery-<slug>-<yyyymmdd>',
   'ONE-SHOT (<why>): ... DISABLE IMMEDIATELY AFTER IT FIRES.',
   'quality-discovery-agent', 'system.agent.scheduled.requests', 86400,
   '{"domain": "<domain>", "site_id": "<uuid>"}'::jsonb,
   NULL, 1, 600, true, true);

-- it fires within seconds; then IMMEDIATELY:
UPDATE scheduled_tasks SET enabled=false, updated_at=now() WHERE name='<the name>' AND enabled;
```

- **Check the rotation's own courtesy gate first** — it refuses a site with work in flight, so you
  should too: `SELECT count(*) FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
  WHERE s.domain='<domain>' AND wi.status='claimed' AND wi.pipeline='build';` → expect 0.
- **⚠ A TRIGGER STAMP IS NOT A RUN** (`bugfix_230_discovery_driver`'s five-stamps CONTRIB). Verify:
  `SELECT orchestration_id, status, created_at, last_activity FROM orchestration_states
   WHERE owner_agent_type='quality-discovery-agent' ORDER BY created_at DESC LIMIT 3;`
- **A one-shot fires the agent's WHOLE check list**, not the one check you care about (quality =
  9 checks as of 2026-08-21). Dedup usually holds the rest — on 2026-08-21 the sweep filed 8 rows
  and every one was the check being tested — but say so before firing, do not assume it.
- Leave the row `enabled=false` afterwards rather than deleting it; the disabled one-shots are the
  estate's record of who forced what, and they are where you copy the shape from next time.

## Prove one page's repair at the served bytes (before → after, with the control that matters)

```bash
D=/tmp/proof; mkdir -p $D
U="https://<domain>/<page_url>"
curl -s "$U?cb=$(date +%s)" -o $D/before.html        # BEFORE promoting the item
# … run the repair …
curl -s "$U?cb=$(date +%s)" -o $D/after.html
for f in before after; do
  echo "$f: backticks=$(grep -o '`' $D/$f.html | wc -l)" \
       "code_tags=$(grep -o '<code>' $D/$f.html | wc -l)" \
       "script_backticks=$(awk '/<script/,/<\/script>/' $D/$f.html | grep -o '`' | wc -l)"
done
diff <(fold -w120 $D/before.html) <(fold -w120 $D/after.html)   # expect ONLY the intended span
```

- The **script-backtick count is the control that carries the risk**: a tool page's own JS uses
  template literals, and a transform that touched them would still show a falling total. Equal
  before and after is the pass; a falling total alone is not.
- Arithmetic worth checking because it is free: bytes should move by exactly
  `+11` per converted span (`<code>` + `</code>` = 13, minus the two backticks removed).

## ⚠ This lane's migrations are NOT in `schema_migrations`, and two of them ABORT on re-run

[MEASURED 2026-08-22] `schema_migrations` (primary key: **filename**) contains none of this lane's
migrations — `497`, `498`, `499`, `513_section_editor…`, `530`, `531`, `540` — because every one was
applied by piping the file to `psql` rather than through `run-migrations.sh`. Only runner-applied
files are recorded.

**So `run-migrations.sh --apply` over `docs/agent_docs/sql_for_agents/` treats them all as pending.**

| file | what a re-run does |
|---|---|
| `540_bugfix_277_recover_provable_content_data.sql` | **harmless** — its `content_data IS NULL` + `md5(rendered_html)` guards update 0 rows, and the verify checks FINAL state, so it passes |
| `530`, `531` | **ABORT** — their anchor-count guards require the pre-change text to be present exactly once, and after application it is gone. Failing loudly is correct, but it fails the batch |

Before running the runner over this directory, scope it or expect those two to stop it. Record a
hand-applied migration deliberately if you want the runner to skip it:
`INSERT INTO schema_migrations (filename, applied_by, notes) VALUES ('<file>', 'hand-applied', '<why>');`

**Also: check the number you intend to use, not the range around it.**
`ls docs/agent_docs/sql_for_agents/ | grep "^540_"` — `540` was already claimed by another lane when
this lane took it (cosmetic, since the ledger keys on filename, but avoidable in one command).
