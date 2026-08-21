# RUNBOOK — bugs_open/198 (the two-writer stylesheet clobber)

Commands that were hard to get right, each with its gotcha attached. Change them HERE.

## 1. The one-query tell: is any site armed right now?

The check the whole bug turns on — a theme row smaller than the file it would be deployed
over. **`octet_length`, not `length`**: a file's size is BYTES, `length()` counts CHARACTERS,
and a UTF-8 stylesheet differs by hundreds (dartsonline: 26,458 chars vs 26,917 bytes — a
148-byte "gap" that was an encoding artefact, not a discrepancy).

```sql
SELECT s.domain, ct.name AS theme, octet_length(ct.css_content) AS row_bytes, ct.version,
       ct.origin,
       (SELECT count(*) FROM sites s2 JOIN style_collections sc2 ON s2.style_collection_id = sc2.id
         WHERE sc2.css_theme_id = ct.id) AS site_count
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
JOIN css_themes ct ON ct.id = sc.css_theme_id
ORDER BY row_bytes;
```

Compare against what the site actually serves:

```bash
for d in <domains>; do printf "%s %s\n" "$d" \
  "$(curl -s -o /dev/null -w '%{http_code} %{size_download}' --max-time 12 \
     "https://$d/assets/css/styles.css")"; done
```

> ⚠ **A bare `curl` without `-L` reads a redirect page as a gutted stylesheet.**
> webdesign.uk 302s to webdesign.co.uk and returns 143 bytes of Cloudflare redirect HTML,
> which looks exactly like a clobbered file. Check the status code, not just the size — an
> earlier session lost several minutes to this and recorded it in the bug file.

## 2. Is the gate live, and does it say what I think?

```sql
SELECT concat_ws(' | ',
  default_config #>> '{workflow,steps,check_base_integrity,config,condition}',
  default_config #>> '{workflow,steps,check_has_css,config,then_step}',
  default_config #>> '{workflow,steps,deploy_css,config,file_shrink_floor}',
  default_config #>> '{workflow,steps,mark_base_unsafe,config,result_fields,parked_by}')
FROM agent_definitions
WHERE type='css-patch-agent' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

Expect: `current_css.css_len >= 4096 AND current_css.site_count <= 1 | check_base_integrity | 0.5 | css_base_integrity_guard_198`.

> ⚠ **`#>>` needs a `jsonb` left operand.** If `default_config` is ever `text` in a context
> you are querying, `#>>` errors with "operator does not exist: text #>> unknown". Cast it.
>
> ⚠ **Use `concat_ws`, not `||`, to join several probes into one row.** A single NULL in a
> `||` chain makes the ENTIRE result NULL — so one missing step silently reports every probe
> as absent, which reads as "the migration did not apply".

## 3. Predict the gate's verdict for the whole fleet before trusting it

This is the query that made the floor defensible — it must show a clean split with nothing
near the boundary:

```sql
SELECT CASE WHEN octet_length(ct.css_content) >= 4096
             AND (SELECT count(*) FROM sites s2 JOIN style_collections sc2
                    ON s2.style_collection_id = sc2.id
                   WHERE sc2.css_theme_id = ct.id) <= 1
            THEN 'PASS -> plan_css_fix' ELSE 'REFUSE -> mark_base_unsafe' END AS gate,
       count(*) AS sites,
       min(octet_length(ct.css_content)) AS min_bytes,
       max(octet_length(ct.css_content)) AS max_bytes
FROM sites s
JOIN style_collections sc ON sc.id = s.style_collection_id
JOIN css_themes ct ON ct.id = sc.css_theme_id
GROUP BY 1;
```

2026-08-21: **19 PASS (13,650–26,917 B) / 3 REFUSE (0–1,649 B)**. If a future run shows
anything between 2,381 and 13,650, the floor needs re-deriving, not overriding.

## 4. Unpark items the gate refused (after a base is repaired)

A parked item **holds its dedup key** — `idx_swi_dedup` does not exclude
`needs_human_review` — so the same finding cannot re-file while parked. This sweep is the
only route back:

```sql
UPDATE site_work_items
   SET status = 'detected', updated_at = NOW()
 WHERE status = 'needs_human_review'
   AND result->>'parked_by' = 'css_base_integrity_guard_198'
   AND site_id IN (<sites whose base now passes §3>);
```

Run §3 first and only unpark sites on the PASS side, or they park again immediately.

## 5. Applying a migration when the pending set belongs to other threads

**Do not run `run-migrations.sh --apply` on this tree.** It applies EVERY pending file, and
the pending set today carries a large backlog from other lanes — several of which are
"applied by hand and never recorded", i.e. the replay hazard the script's own lint warns
about. Apply your own file, then record it:

```bash
kubectl exec -i -n ai-persona-system postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/NNN_name.sql

./scripts/migration/run-migrations.sh --record-only NNN_name.sql --note "<why out-of-band>"
```

> ⚠ **Recording is not optional bookkeeping.** A migration carrying a probe guard that
> `RAISE`s on re-application will, if left unrecorded, be picked up by the next `--apply`,
> raise, and **stop the run** — blocking every later migration in the queue, including other
> lanes'. Apply-by-hand and record-only are one operation, not two.
>
> `_HOLD.sql` and `_ROLLBACK.sql` are excluded automatically (`SIDECAR_RE='_[A-Z][A-Z0-9_]*\.sql$'`),
> so a held migration is safe from the runner — but only because of that naming.

## 6. Checking a migration number is free

```sql
SELECT filename FROM schema_migrations ORDER BY filename DESC LIMIT 8;
```

plus `ls docs/agent_docs/sql_for_agents/ | grep -E "^5[0-9][0-9]"` **including untracked
files** (`git status` the dir — other sessions leave uncommitted migrations).

> ⚠ **`LIKE '54[23]%'` is not a character class.** SQL `LIKE` has no `[...]`; that pattern
> matches the literal string `54[23]` and returns zero rows whether or not the numbers are
> taken — a check that cannot come out otherwise. Use `LIKE '542%' OR LIKE '543%'`, or `~
> '^54[23]'`. Full write-up in `WRONG_CALLS.md`, 2026-08-21.

## 7. Proving the shrink floor after the images roll

Both halves must be new — the guard exists only when the chassis sends the field AND the
adapter enforces it. Probe each, **with a negative control in the same breath**:

```bash
for p in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | cut -d/ -f2); do
  kubectl -n ai-persona-system exec "$p" -- grep -ac 'file_shrink_floor' /proc/1/exe   # expect >= 1
  kubectl -n ai-persona-system exec "$p" -- grep -ac 'file_shrink_floor_NOTREAL' /proc/1/exe  # expect 0
done
# then the same two greps against the git-adapter pod
```

> ⚠ Never `strings` (absent from the debian-slim images; behind `2>/dev/null` its failure is
> indistinguishable from "not stamped"). A control that comes out PRESENT means the probe
> matches everything and proves nothing.

**The pod-grep above is the check that settles it.** There is also an end-to-end log line,
but treat it as an opportunistic bonus, never as the verification:

```bash
kubectl -n ai-persona-system logs -l app=git-adapter --tail=500 \
  | grep 'file_shrink_floor: commit passed the shrink floor'
```

> ⚠ **CORRECTED 2026-08-21 by the council's `debug_historian` seat** — this section originally
> called the log line "the honest one". It is not, and the reasoning was backwards. Pod log
> history here is short (~90 seconds is the recorded figure), `kubectl logs -l app=<x>` can
> return zero lines for a live pod, and **this guard only emits on a css-patch dispatch**, so
> an absent line means "not in range", never "not shipped". What the line does prove, IF you
> happen to catch one, is three things at once — the field arrived, the guard measured, and a
> healthy commit passed — which is why it is worth grepping after a real deploy. It just
> cannot answer "did it ship?", and only the binary probe can.

## 8. Restoring a clobbered site (the recipe, from three lanes' experience)

1. Find the pre-clobber blob: the commit **before** the first `CSS fix:` commit on
   `<domain>/assets/css/styles.css` in the deploy repo.
2. **`git pull` before assuming the local repo copy is safe** — the remote may already carry
   the clobber, and an automatic merge of damage against clean content resolves to a third
   thing that becomes the baseline every later render diffs against (cookly.uk: 875 bytes of
   neither).
3. Write the blob into the theme row md5-guarded, then **the file restore is optional**: the
   next patch run appends to the truth and deploys the whole row. Seed the row always; push
   the file only when the site is visibly broken now.
4. **Check which side of 2026-08-14 the blob falls on.** A restore reinstates a point in
   time, and a pre-08-14 blob carries the old `-ink` derivation. Run the two-clause staleness
   check: `<x>-ink == --color-text AND <x>-ink != --color-<x>` — the second clause is what
   separates *substituted* from *returned unchanged*, and without it you will "correct" a
   correct token.
5. Do **not** carry forward patch rules authored against a clobbered file. They were written
   blind against an empty `current_css`, and several target selectors that match nothing
   (`H3.H3`, `p.P` — `render_audit.py` labels findings by uppercased TAG and the agent read
   the label as a class).
6. `sites.github_repo` is empty for several domains — resolve the repo the git-adapter's way
   (config `repo_name` → `site_record.github_repo` → the `sites` row → default `"sites"`),
   never by assuming vm-sites.

## 9. Validate the workflow GRAPH after any config edit — this found a real defect

Run this after **any** migration that rewires an agent's steps. It resolves every edge
(`next_step`, `error_step`, `config.then_step`, `config.else_step`) against the step map:

```sql
WITH s AS (SELECT default_config #> '{workflow,steps}' AS steps FROM agent_definitions
           WHERE type='<agent>' AND is_active
             AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL),
edges AS (
  SELECT k AS from_step, v->>'next_step' AS tgt, 'next' AS kind
    FROM s, jsonb_each(s.steps) AS e(k,v) WHERE v ? 'next_step'
  UNION ALL SELECT k, v->>'error_step', 'error'
    FROM s, jsonb_each(s.steps) AS e(k,v) WHERE v ? 'error_step'
  UNION ALL SELECT k, v->'config'->>'then_step', 'then'
    FROM s, jsonb_each(s.steps) AS e(k,v) WHERE v->'config' ? 'then_step'
  UNION ALL SELECT k, v->'config'->>'else_step', 'else'
    FROM s, jsonb_each(s.steps) AS e(k,v) WHERE v->'config' ? 'else_step')
SELECT from_step, kind, tgt,
       CASE WHEN s.steps ? tgt THEN 'ok' ELSE '*** DANGLING ***' END AS resolves
FROM edges, s WHERE tgt IS NOT NULL ORDER BY resolves DESC, from_step;
```

> ⚠ **READ THE ROWS, do not just check for DANGLING.** This query was written to catch an
> orphaned step after migration 542. Every edge resolved — and the table it printed is what
> revealed `check_saved | else | complete_error`, an arm 542 had never touched, still minting
> `complete` for a refused append (fixed by 546). **Reading the steps you edited cannot find
> the step you did not edit**, and a verify block written from your own diff can only ever
> confirm your diff. The `count(*) FILTER (WHERE NOT (steps ? tgt))` form is the assertable
> version and is now embedded in 546's verify block; the human-readable form above is the one
> that finds what you were not looking for.

## 10. Execute an installed step query VERBATIM before trusting it

A migration's `DO/RAISE` verify can only assert that a step's SQL *string* matches what you
wrote. **Step SQL is DATA to the migration — it parses only when the step RUNS**, so a
syntax or semantic error ships silently and fails live. Extract and run it:

```bash
Q=$(kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT default_config #>> '{workflow,steps,<step>,config,query}' FROM agent_definitions
WHERE type='<agent>' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;")

SID=$(… the real parameter value …)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -v ON_ERROR_STOP=1 -c "PREPARE p AS $Q; EXECUTE p('$SID'); DEALLOCATE p;"
```

`PREPARE` succeeding is itself the parse proof; the `EXECUTE` proves the columns come back.
**Run it on a row from each arm** — for the 198 gate that is a healthy site (`26917 / 1`,
pass) and a shared-theme site (`1649 / 2`, refuse), so the query is shown to discriminate
rather than merely to run.

> ⚠ `EXECUTE p((SELECT …))` fails — "cannot use subquery in EXECUTE parameter". Resolve the
> value into a shell variable first.
