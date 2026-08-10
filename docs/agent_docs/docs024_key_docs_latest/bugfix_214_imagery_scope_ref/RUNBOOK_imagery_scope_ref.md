# RUNBOOK — bug 214, imagery scope_ref

Every command that was hard to get right, with its gotcha attached. Change it
**here**, not in your scrollback.

---

## R1 — THE census: rows invisible to consumers (the one that matters)

⚠ **Resolve against `pages.name`, NOT `site_plan_pages.name`.** Every consumer joins
the deployed `pages` table. Measured both ways 2026-08-10: plan table says 22, `pages`
says **10**, and all 12 of the difference work fine. Using the plan table over-counts
and, if you build a repair on it, breaks working rows.

```sql
WITH cur AS (
    SELECT sp.id AS plan_id, sp.site_id, si.domain
      FROM site_plans sp JOIN sites si ON si.id = sp.site_id
     WHERE sp.is_current
)
SELECT count(*) AS total_page_and_section,
       count(*) FILTER (WHERE NOT EXISTS (
         SELECT 1 FROM pages p
          WHERE p.site_id = cur.site_id
            AND p.name = split_part(spi.scope_ref, ':', 1))) AS invisible_to_consumers
  FROM site_plan_imagery spi JOIN cur ON cur.plan_id = spi.plan_id
 WHERE spi.scope IN ('page','section');
-- 2026-08-10: 176 | 10
```

`split_part(x, ':', 1)` is safe for **both** scopes: a page-scope ref contains no colon
(`chk_scope_ref_consistency` forbids it), so `split_part` returns it whole.

## R2 — list the invisible rows, with what would rescue each

Use this before and after any repair; the `plan_page_candidate` column is what decides
whether a row is rescuable at all.

```sql
WITH cur AS (
    SELECT sp.id AS plan_id, sp.site_id, si.domain
      FROM site_plans sp JOIN sites si ON si.id = sp.site_id WHERE sp.is_current
)
SELECT cur.domain, spi.scope, spi.scope_ref, spi.key, spi.kind,
       (SELECT string_agg(spp.name, ',') FROM site_plan_pages spp
         WHERE spp.plan_id = spi.plan_id
           AND spp.name LIKE split_part(spi.scope_ref, ':', 1) || '%') AS plan_page_candidate,
       EXISTS(SELECT 1 FROM assets a
               WHERE a.site_id = cur.site_id AND a.asset_key = spi.key
                 AND a.status = 'active') AS asset_paid_for
  FROM site_plan_imagery spi JOIN cur ON cur.plan_id = spi.plan_id
 WHERE spi.scope IN ('page','section')
   AND NOT EXISTS (SELECT 1 FROM pages p
                    WHERE p.site_id = cur.site_id
                      AND p.name = split_part(spi.scope_ref, ':', 1))
 ORDER BY cur.domain, spi.scope_ref;
```

⚠ **`asset_paid_for` is the column that makes this a priority rather than a tidy-up.**
8 of the 10 were `t` on 2026-08-10.

## R3 — the WRITE-path check (post-roll): is the fix actually running?

This one resolves against `site_plan_pages` **deliberately** — it asks whether the
plan is internally consistent, which is what the code now guarantees. R1 asks whether
consumers can see it. They are different questions; do not merge them.

```sql
SELECT si.domain, spi.scope, spi.scope_ref, spi.key
  FROM site_plan_imagery spi
  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current
  JOIN sites si ON si.id = sp.site_id
 WHERE spi.scope IN ('page','section')
   AND split_part(COALESCE(spi.scope_ref,''), ':', 1) NOT IN
       (SELECT name FROM site_plan_pages WHERE plan_id = spi.plan_id);
```

**Disconfirming result:** a returned row whose page part is a *raw alias* of a page
that same plan DOES contain (e.g. `about` while `site_plan_pages` holds `about-index`
on the same `plan_id`) means the rewrite is not running in production. Rows naming a
page that exists nowhere are **legitimate survivors** — those are the ones kept
verbatim on purpose.

## R4 — did the durable record fire?

Every row R3 returns must have a same-day log row. **An unresolved ref with no log row
disproves the durable half of the fix.**

```sql
SELECT occurred_at, error_code, error_message, context->>'scope_ref_raw',
       context->>'scope_ref_written', context->>'plan_id'
  FROM agent_error_log
 WHERE error_code IN ('IMAGERY_SCOPE_REF_UNRESOLVED','IMAGERY_SCOPE_REF_ORDINAL_ANOMALY')
 ORDER BY occurred_at DESC LIMIT 20;
```

## R5 — pod-verify after the roll (never verify at git or at the tag)

```bash
POD=$(kubectl get pods -n ai-persona-system -l app=agent-chassis -o name | head -1)
kubectl exec -n ai-persona-system ${POD#pod/} -- sh -c \
  'strings /app/agent-chassis | grep -c "imagery scope_ref canonicalised"'   # expect >=1
kubectl exec -n ai-persona-system ${POD#pod/} -- sh -c \
  'strings /app/agent-chassis | grep -c "imagery scope_ref pineapple"'        # negative control, expect 0
```

⚠ **Do BOTH replicas** — a label greps 2 pods of 34, and one replica can be older.
⚠ **The negative control is not optional.** A positive-only grep proves the pipeline,
never your spelling.

## R6 — apply the backfill (AFTER the roll, not before)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db \
  < docs/agent_docs/sql_for_agents/373_bugfix_214_canonicalise_orphaned_imagery_scope_refs.sql
```

⚠ **Applying before the code is live buys exactly one plan generation** — the next
replan re-mints the same refs.
⚠ The script's guard is a `DO $$ ... RAISE EXCEPTION`, **not** a `SELECT`.
`ON_ERROR_STOP` does not fire on a non-empty result set, so a verify block of `SELECT`s
cannot stop a `COMMIT`. It aborts unless **exactly 1** unresolvable row remains, which
fails in both directions: 10 means inert, 0 means it repaired the row it was told to
leave (mortgagecalculator's `tools-index`, which has no canonical variant).

## R7 — the mutation check (do this to ANY test you claim guards a call site)

The fifteen unit tests all pass with the fix deleted. Only the wiring suite catches it.

```bash
# 1. save, 2. delete the resolution block from WriteSitePlanAction, 3. run, 4. restore
go test ./platform/orchestration/actions/ -run 'TestWriteSitePlan_'
#   with the block deleted -> 2 FAIL   (correct)
#   with it restored       -> ok
```

⚠ **Comment the call out; do not delete a symbol** — a deletion that breaks the build
is not evidence the *test* catches anything.

## R8 — ownership check before taking any bug (three ways, not one)

`who-owns.py` reads commits and is blind to a session mid-fix. The transcript grep is
the one that finds those.

```bash
scripts/who-owns.py 214
git log --since="36 hours ago" --format="%s" | grep -oE '\b2[0-4][0-9]\b' | sort -u

cd ~/.claude/projects/-home-ant-projects-agentchassis/
find . -name '*.jsonl' -newermt "$(date +%Y-%m-%d)T11:00:00" -size +50k > /tmp/live.txt
for f in $(cat /tmp/live.txt); do
  c=$(grep -c '<bug-file-slug>' "$f"); [ "$c" != "0" ] && echo "$(basename $f): $c"
done
```

⚠ **A count of `1` is noise** — that is the file appearing in an `ls bugs_open` listing.
⚠ `find -newermt` needs an **ISO timestamp**; `-newermt '-3 hours'` errors out on this
box's `find` (it is `bfs`).

## R9 — register index bookkeeping

```bash
cd docs/agent_docs/docs026_concept_register/register/
grep -cE '^\| [A-Z]{2,4}-[0-9]{3} \|' 000_concept_index.md            # the documented headline command
comm -13 <(grep -oE '^\| [A-Z]{2,4}-[0-9]{3} \|' 000_concept_index.md | tr -d '| ' | LC_ALL=C sort -u) \
         <(grep -rhoE '^### [A-Z]{2,4}-[0-9]{3}' *.md | sed 's/### //' | LC_ALL=C sort -u)
```

⚠ **`LC_ALL=C` on both sides.** Without it `comm` and `sort` disagree on collation and
you get phantom orphans (a truncated `IMG`) plus a "not in sorted order" warning that
is easy to skim past.
⚠ Re-grep the count immediately before AND after your edit — concurrent lanes land rows
in the gap (two did, in this session's window).
