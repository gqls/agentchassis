# RUNBOOK — the unpublish primitive (`bugs_open/098`)

Every command that was hard to get right, with its gotcha attached.

---

## Verify the bug still reproduces

```sh
curl -sS -o /dev/null -w '%{http_code}\n' https://robot-hands.com/learning-center/index.html   # 200 = still live
```
```sql
SELECT count(*) FROM pages WHERE status='archived' AND deployed_at IS NOT NULL;   -- 13 on 2026-08-03
```

**Gotcha:** `deployed_at IS NOT NULL` means *a deploy happened once*, NOT *the
page is fetchable now*. It is the load-bearing half of the shared eligibility
predicates and nothing clears it on archive.

## Is the page actually FROZEN, or is something re-publishing it?

Do this before retracting anything — a retraction of a page something still
re-renders is undone within hours.

```sh
gh api "repos/gqls/sites/commits?path=<domain>/<path>&per_page=5" \
  --jq '.[] | .commit.author.date + "  " + (.commit.message|split("\n")[0])'
```
A twice-daily cadence means a scheduler or a work-item emitter owns it. Trace it:

```sql
SELECT jsonb_pretty(collected_data->'input_data')
  FROM orchestration_states
 WHERE collected_data->'input_data'->>'page_name' = '<name>'
 ORDER BY created_at DESC LIMIT 1;    -- input_data.source names the emitter
```

**Gotcha:** `orchestration_states` keeps terminal rows ~24h, so this only works
for a recent run. The git history does not expire and is the better first look.

## Probe GitHub's deletion semantics without touching the repo

```sh
BASE=$(gh api repos/gqls/sites/commits/master --jq '.commit.tree.sha')
gh api repos/gqls/sites/git/trees -X POST --input - <<EOF
{"base_tree":"$BASE","tree":[{"path":"<path>","mode":"100644","type":"blob","sha":null}]}
EOF
```
`201` if the path exists, **`422 GitRPC::BadObjectState`** if it does not.

**Why it is safe:** POST `/git/trees` creates an *unreferenced* tree object. No
commit, no ref moves, no workflow fires, and git GCs it. Do NOT follow it with
`/git/commits` + `PATCH /git/refs` unless you mean it.

## Derive a page's deploy path — never by hand

```sh
# feed "domain<TAB>name<TAB>url" on stdin; uses the REAL shared helper
go run ./<scratch>/pathcheck/main.go < candidates.tsv
```
The program is 20 lines and calls `datahelpers.PageFilePathFromURL`.

**Gotcha, and it is the whole point:** re-implementing the derivation in bash or
SQL to "check" the Go one proves nothing — a second implementation is exactly
the drift `bugs_closed/125` was about. Call the real function.

**Gotcha 2:** check each page against ITS OWN deploy repo.
`resolveGitRepoNameDB` resolves `sites.github_repo` per site (`vm-sites` for
VM-hosted sites, default `sites`). A sweep that hardcodes `gqls/sites` will
report a stale leftover as the live artefact — that happened here with
relojistas.

## Exercise the primitive without touching production

`unknown-domain/` is in the sites repo and does **not** match the deploy
workflow's changed-domain regex (`^[^/]+\.[^/]+/` — it has no dot), so it syncs
to nothing.

```sh
CORR=$(uuidgen); REQ=$(uuidgen)
PAYLOAD=$(jq -c -n --arg c "$CORR" --arg r "$REQ" '{
  headers:{correlation_id:$c,orchestration_id:$c,request_id:$r,client_id:"demo_client",
           responses_topic:"system.agent.generic.responses",step_name:"probe"},
  body:{action:"delete_file",data:{repo_name:"sites",domain:"unknown-domain",
        paths:["098-retraction-probe.html"],commit_message:"probe"}}}')
printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm "kcat-$(date +%s)" \
  --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.adapter.git.requests -H "correlation_id=$CORR" -H "message_type=request" \
  -H "action=delete_file" -H "responses_topic=system.agent.generic.responses"
```

**Gotcha:** the payload MUST be one line. `kcat -P` splits stdin on newlines into
separate messages, so a pretty-printed JSON becomes N broken messages, each of
which fails silently. `jq -c`, always.

## Verify a deploy — the chain that actually proves it

In-pod `strings`/`grep` does not work for these adapters: the binary is at
`/root/<name>` and the container runs as uid 1000, so it is `Permission denied`.
Use the digest chain instead.

```sh
# 1. extract the binary from the image you built and grep it, WITH a negative control
CID=$(docker create docker.io/aqls/git-adapter:v1.0.1235)
docker cp "$CID":/root/git-adapter ./ga && docker rm "$CID"
grep -ac 'no paths in delete_file data'  ./ga    # POS, expect >0
grep -ac 'repo_name and files are required' ./ga # NEG, expect 0 (string REMOVED by the change)
grep -ac 'fast forward'                   ./ga   # CONTROL, expect >0

# 2. tie that exact image to what is running
kubectl -n ai-persona-system get pods -l app=git-adapter \
  -o jsonpath='{range .items[*]}{.metadata.name}{"  "}{.status.containerStatuses[0].imageID}{"\n"}{end}'
# must equal the digest `docker push` printed
```

**Gotcha:** a positive control proves the PIPELINE, never the pattern. My first
attempt printed `0` for every check *including the control*, because the binary
path was wrong and `strings` was reading stdin. Without the control that reads as
"the fix did not ship".

## Validate SQL before shipping it

`go build` cannot parse SQL. `PREPARE` it against the live schema:

```sql
PREPARE q (uuid, text[]) AS <your query>;
EXECUTE q('<site-id>', ARRAY['/some/url.html']);
```

This caught `site_components.component_type` — a column that does not exist (it
is `slot_name`).

**Gotcha:** `[]string` is NOT a supported parameter type in this codebase
(database/sql over pgx stdlib, with neither `pq.Array` nor pgx array types
imported). Use `datahelpers.PGTextArrayLiteral(xs)` with an explicit `::text[]`
cast in the SQL, which is the existing convention.

## Audit a retraction's graph before firing it

```sql
-- editorial inbound (BLOCKS the retraction)
SELECT DISTINCT p.name, pc.slot_name FROM page_components pc JOIN pages p ON p.id=pc.page_id
 WHERE p.site_id=$1 AND p.status='active' AND pc.rendered_html LIKE '%href="'||$2||'"%';
SELECT DISTINCT sc.slot_name FROM site_components sc
 WHERE sc.site_id=$1 AND sc.rendered_html LIKE '%href="'||$2||'"%';
-- nav inbound (DEACTIVATED by the retraction)
SELECT i.id, i.label FROM site_nav_items i JOIN site_nav_groups g ON g.id=i.group_id
 WHERE g.site_id=$1 AND i.status='active' AND (i.url=$2 OR i.page_id=$3);
```

**Gotcha, and this one bit:** a census that reads only `page_components` will
report most of the site as unreferenced, because the nav and the chrome are
where the links actually are. All three sources, every time.

**Gotcha 2:** all-zero results are also what a *broken* query returns. Always
re-run against a page you know IS linked, as a positive control — on
robot-hands, `/contact.html` gives 4 body / 2 chrome / 1 nav.

## Build and roll without disturbing other sessions

```sh
make build-git-adapter  IMAGE_TAG=v1.0.1235      # IMAGE_TAG is `?=`, so CLI wins
make build-agent-chassis IMAGE_TAG=v1.0.1235
```

**Gotcha:** do NOT edit `IMAGE_TAG` in the makefile — it is shared, usually
dirty from another session, and overriding on the command line avoids the
same-file passenger entirely.

**Gotcha 2:** `make deploy-agents` rolls agent-chassis *and* everything else. If
a council run or a long orchestration is in flight, that kills it. Roll one
service by applying its overlay alone:

```sh
kubectl apply -k deployments/kustomize/services/git-adapter/overlays/production/uk_001
kubectl -n ai-persona-system rollout status deployment/git-adapter --timeout=180s
```

## Archiving a page? Retraction is the SECOND HALF of the procedure (decision 2026-08-04)

Archiving is `pages.status='archived'` by hand — and by owner-delegated decision it does
NOT trigger retraction automatically (no code seam exists; an automatic file-deleter on a
hand-set flag is unguarded destructive authority). So the procedure is TWO steps, always:

```sh
# 1. archive (whatever SQL you were going to run), then:
SITE_ID=<uuid> PAGE_IDS='["<page-uuid>"]' ./docs/agent_docs/sql_for_agents/216_TRIGGER_page_retraction.sh
# 2. acceptance is TWO-PART: curl 404 now, AND still-404 after the next ~08:0x/20:0x
#    refresh + zero new page_rerender rows for the page.
```

**Gotcha:** the audit you'll want afterwards is NOT in `collected_data.retraction_audit`
— the await park discards it (coordinator.go:2052, RFC_012 addendum 2). Refusals ARE
durable in `agent_error_log` (`action='retract_page_deployment'`); until debt 5b ships,
a clean run's full audit lives only in pod logs, so read the refusal table and the curls.
