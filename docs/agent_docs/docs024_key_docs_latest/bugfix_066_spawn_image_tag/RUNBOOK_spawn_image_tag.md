# RUNBOOK — 066 spawn image tags

Every command here was run against production on 2026-07-27 unless marked otherwise.

## The one command that answers "are we exposed"

```bash
scripts/check-agent-image-drift.sh      # or: make check-agent-image-drift
```
Prints three separate answers — what the Deployment runs, what the rows say, what spawned
pods are actually running — and exits `0` no drift / `1` drift in the record / `2` could not
determine. **Gotcha:** drift in the record is not exposure once the fix is live; the script
says so and gives the pod-grep that decides which world you are in.

## Is the fix in the running binary?

```bash
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "bugs_open/066: agent_definitions.image_tag trails"'
```
**Gotcha:** grep a string the fix *created* and that survives into the binary. A comment does
not (`bugfix-079` lost a round to exactly that), and a typed constant can be optimised away
(`cta_link_integrity`). This one is a live `logger.Warn` message, so it is in the binary.
Negative control: the same grep on a pre-v1.0.117x pod must return `0`.

## Census of the rows (what the old RUNBOOKs call "the 066 census")

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT COALESCE(image_tag,'(null)') tag, count(*), max(updated_at) AS newest_update
FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
GROUP BY 1 ORDER BY 2 DESC;"
```
**Gotcha:** `max(updated_at)` is the part worth having. On 2026-07-27 it read `13:44:56`
while the chassis pod's `startTime` was `13:45:31` — the row *led* the Deployment by 35 s.
Without the timestamp the census looks clean and hides that the two records move separately.

## Repair the record (does not touch pods)

```bash
DRY_RUN=1 scripts/deploy/update-agent-images.sh   # read-only preview first
scripts/deploy/update-agent-images.sh
```
Reads the tag from the **live Deployment**, not `$(IMAGE_TAG)`. **Gotcha:** it deliberately
refuses a digest-pinned Deployment rather than writing a fabricated tag, and it will not
touch `is_snapshot` / soft-deleted / `pin_image_tag` rows.

## Prove the scoped predicate WITHOUT writing (do this before trusting any sync)

```sql
SELECT
  count(*)                                                                            AS all_rows,
  count(*) FILTER (WHERE deleted_at IS NOT NULL)                                      AS excluded_soft_deleted,
  count(*) FILTER (WHERE COALESCE(is_snapshot,false))                                 AS excluded_snapshots,
  count(*) FILTER (WHERE COALESCE(default_config->'pin_image_tag','false'::jsonb) = 'true'::jsonb) AS excluded_pinned,
  count(*) FILTER (WHERE deleted_at IS NULL
                     AND NOT COALESCE(is_snapshot,false)
                     AND image_repository = 'docker.io/aqls/agent-chassis'
                     AND COALESCE(default_config->'pin_image_tag','false'::jsonb) <> 'true'::jsonb) AS in_scope
FROM agent_definitions;
```
Result 2026-07-27: `183 / 2 / 1 / 0 / 180`. **Gotcha:** the old `WHERE 1=1` sync hit all
**183** — including the one `is_snapshot` row, which is a `021_model_swap_and_rollback.sql`
rollback copy whose tag is exactly what a rollback is supposed to restore.

## Prove a check's FAILING branch, not just its passing one

```sql
-- the drift script's own query, against a tag nothing is on
SELECT count(*) FROM agent_definitions
WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false) = false
  AND image_repository = 'docker.io/aqls/agent-chassis'
  AND COALESCE(image_tag,'') <> 'v1.0.9999'
  AND COALESCE(default_config->'pin_image_tag','false'::jsonb) <> 'true'::jsonb;
```
Must return non-zero (returned `180`). **Gotcha:** without this, a zero from the real query
is indistinguishable from a query that is silently broken — which is not hypothetical: the
first run of `check-agent-image-drift.sh` printed an empty census because the SQL had an
aggregate in `GROUP BY` and `psql_q` was swallowing stderr with `2>/dev/null`. The redirect
is now gone, with a comment saying why.

## RBAC — what the service account can and cannot do

```bash
kubectl auth can-i get pods        -n ai-persona-system --as=system:serviceaccount:ai-persona-system:ai-persona-app   # yes
kubectl auth can-i get deployments -n ai-persona-system --as=system:serviceaccount:ai-persona-system:ai-persona-app   # no
```
**Gotcha:** this is why the resolver reads its own pod. "Read the Deployment" is the obvious
design and it needs an RBAC change nobody would notice was missing until a spawn failed.

## Unit tests

```bash
go test ./platform/orchestration/actions/ -run 'TestChooseAgentImage|TestParseImageRef|TestSameRepository' -v
```
**Gotcha:** `go build ./...` fails in this tree on `cmd/reasoningset` — another session's
uncommitted WIP, not this change. Build `./platform/... ./internal/... ./pkg/...` to get a
meaningful answer. Likewise `gofmt -l` lists ~11 pre-existing unformatted files in the
actions package; check `gofmt -l` on *your* files only, and confirm HEAD was clean first:
`git show HEAD:<file> > /tmp/x.go && gofmt -l /tmp/x.go`.

## Post-roll verification (OWED — see bugs_open/066 § How to verify)

Induce the failing branch: snapshot a non-critical agent row, set its `image_tag` stale, fire
it, and confirm the spawned pod carries the **chassis's** tag plus the drift warning in the
chassis log. A green happy path proves deployment, not correctness.
