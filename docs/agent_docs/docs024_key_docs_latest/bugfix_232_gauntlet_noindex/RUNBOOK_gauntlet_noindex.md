# RUNBOOK — `bugs_open/232` gauntlet noindex

Only commands that were hard to get right, each with its gotcha attached.

## 1. Prove the bug is still real (do this before anything, it is the whole premise)

```bash
curl -s "https://vonc.com/tools/gauntlet/round.html?cb=$(date +%s)" | grep -io '<meta name="robots"[^>]*>'
curl -sI "https://vonc.com/tools/gauntlet/round.html" | grep -i 'x-robots\|cf-cache-status'
```

> ⚠ **`-I` alone cannot answer this question.** The fix is a **meta tag in the body**,
> so a header-only check reads "no robots anything" both before and after. Use the body
> grep for the page and `-I` only for the API.
>
> ⚠ **Cache-bust every time.** Cloudflare fronts the origin. A cached copy makes a
> landed fix look missing and a missing fix look landed, depending on when you asked.

## 2. Identify the hosting before proposing a header (this is what ruled out fix candidate 1)

```bash
curl -sI https://vonc.com/ | grep -iE 'server|x-amz'
#  server: cloudflare
#  x-amz-request-id / x-amz-version-id   <-- B2 objects. NO server-side path we control.
```

> The presence of `x-amz-*` under `server: cloudflare` is the tell: static objects,
> edge-cached. Do not spend time designing a response header for the HTML page.
> Grep for Caddy first and read **what it fronts** — the Caddyfiles in
> `gauntlet_dead_cta/infra/` front **tools-api**, not the site.

## 3. Which code path actually renders the page (getting this wrong wastes the whole fix)

```bash
grep -n "assemblePage(" platform/orchestration/actions/rerender_single_page_action.go
grep -rn '"assemble_page"' platform/ --include=*.go | grep -v _test.go
```
```sql
-- who dispatches the OTHER producer. DO NOT assume it is dead: a test-file comment says it is.
SELECT type FROM agent_definitions
 WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
   AND default_config::text ILIKE '%"action": "assemble_page"%';
--  site-work-orchestrator · pageflow-builder · page-rebuild      [MEASURED 2026-08-09]
```

> ⚠ **`inject_canonical_link_test.go`'s header asserts "no live agent uses
> assemble_page". It is false.** A source comment is not a census. Run the query.

Trace for the record: `rerender_page_vonc.sh` → `page-rerender` agent →
`RerenderSinglePageAction` → `assemblePage()` → the four head injections.

## 4. Apply the migration — induce the guard FIRST

```bash
# INDUCE: same file, mangled uuid. The DO/RAISE must abort and roll the ALTER back too.
sed "s/4629451e-e4f2-4fe2-b258-35107b5cb51e/00000000-0000-0000-0000-000000000000/" \
  docs/agent_docs/sql_for_agents/352_pages_noindex_flag.sql > /tmp/induce.sql
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < /tmp/induce.sql
#  expect: ERROR ... found 0     AND THEN:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc \
  "SELECT count(*) FROM information_schema.columns WHERE table_name='pages' AND column_name='noindex';"
#  expect 0 — proves the whole transaction rolled back, ALTER included

# then the real apply
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < docs/agent_docs/sql_for_agents/352_pages_noindex_flag.sql
```

> ⚠ **A verify block of bare `SELECT`s cannot stop a `COMMIT`** — `ON_ERROR_STOP`
> ignores a non-empty result. Use `DO` / `RAISE EXCEPTION`, and **induce it**, or you
> have written decoration.

Verify by **row identity**, never by count alone:
```sql
SELECT p.id, s.domain, p.url, p.noindex FROM pages p JOIN sites s ON s.id=p.site_id WHERE p.noindex;
SELECT count(*) FILTER (WHERE noindex) AS is_true, count(*) FILTER (WHERE NOT noindex) AS is_false,
       count(*) FILTER (WHERE noindex IS NULL) AS is_null, count(*) AS total FROM pages;   -- 1|629|0|630
```

## 5. Before adding a column to `pages`, prove nothing reads it positionally

```bash
grep -rn "SELECT \* FROM pages" --include=*.go platform/ internal/       # 0 hits
grep -rn "INSERT INTO pages[[:space:]]*(" --include=*.go platform/ internal/ | wc -l   # 8, all with column lists
```
```sql
\d site_snapshots     -- pages_snapshot is jsonb, not a fixed column shape
```

## 6. Prove the tests are not vacuous (mutation, not just green)

```bash
cp platform/orchestration/actions/rerender_single_page_action.go /tmp/r.bak
# mutation 1: make the injection a no-op  -> 5 cases MUST fail
# mutation 2: swap the exact-tag marker for `<meta name="robots"` -> the COEXISTENCE case MUST fail
go test ./platform/orchestration/actions/ -run 'TestInjectRobotsNoindex'
cp /tmp/r.bak platform/orchestration/actions/rerender_single_page_action.go     # restore, re-run, expect ok
```

> Mutation 2 is the one that matters: it is the only thing distinguishing my
> idempotency-marker judgement from the other reasonable one.

## 7. Compile against **committed HEAD**, not your tree

```bash
d=$(mktemp -d); git archive HEAD | tar -x -C "$d"
(cd "$d" && go build ./platform/orchestration/actions/... ./internal/tools-api/... \
         && go test ./platform/orchestration/actions/ -run 'TestInjectRobotsNoindex')
```

> Note: `TestEveryActionInputSpecHasARegistryEntry` and
> `TestEveryCheckProducedItemTypeIsClassified` **fail at clean HEAD** and are nothing to
> do with this work. Run your own test by name; do not read a red package as your fault.

## 8. After the roll — the order matters and step 2 is silently droppable

```bash
# a) every replica. Nothing was REMOVED here, so there is no negative-string control;
#    use a known-present symbol in the SAME exec as a pipeline positive control.
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
  'strings /app/agent-chassis | grep -c injectRobotsNoindex; strings /app/agent-chassis | grep -c injectCanonicalLink'

# b) WAIT >=300s after any chassis pod (re)start, or the dispatch is silently dropped
docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/scripts/rerender_page_vonc.sh \
  4629451e-e4f2-4fe2-b258-35107b5cb51e

# c) at the served artefact, cache-busted
curl -s "https://vonc.com/tools/gauntlet/round.html?cb=$(date +%s)" | grep -io '<meta name="robots"[^>]*>'

# d) THE NO-OP CONTROL, on the new binary: re-render another vonc page, assert ABSENCE.
#    A page rendered BEFORE the roll proves nothing about the gate.
```

> Only the **rerender + verify** steps of `gauntlet_dead_cta` RUNBOOK §18 apply. The
> component is untouched, so skip `build_record_page_sql.py` / `deliver_record_component.py`
> entirely — running them would rewrite a component this fix never touched.

## 9. tools-api is a DIFFERENT deploy (it is not in this cluster)

```bash
kubectl get deploy -n ai-persona-system | grep -i tools     # no rows — it is NOT here
```
It runs on the **island VM** under docker compose: rebuild + `docker save|load` +
`compose up -d` (RUNBOOK `gauntlet_dead_cta` §5). Verify from outside:
```bash
curl -sI 'https://tools.apis.uk/api/v1/tools/gauntlet/round/<REAL-published-slug>' \
  -H 'Origin: https://vonc.com' | grep -i x-robots
```
> ⚠ **Use a real published slug.** A 404 returns no header either, which reads
> identically to a fix that is not there.

## 10. Reading the council verdict

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '1139cbbe-3173-4886-846b-c25daeeda93c';
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='1139cbbe-3173-4886-846b-c25daeeda93c' AND kind='council_report' ORDER BY created_at;
```
> ⚠ CLAUDE.md's `doc_notes ... ORDER BY created_at DESC LIMIT 1` returns **whichever
> lane's verdict is newest**, not yours (noted by the 168 lane, 2026-08-09). Key on the
> correlation. Objections live in `diagnosis_artifacts.body`, not in `metadata`.
> A missing orchestration row is **latency, not a dropped dispatch** — budget ~30 min.
