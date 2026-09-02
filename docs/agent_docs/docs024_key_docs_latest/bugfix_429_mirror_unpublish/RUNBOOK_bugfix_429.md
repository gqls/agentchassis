# RUNBOOK — bugfix 429

## Probe the orphan with controls (the recipe that re-validated the bug)

```bash
for u in "https://boxingonline.ugg2.com/contact.html" \
         "https://boxingonline.ugg2.com/index.html" \
         "https://boxingonline.ugg2.com/zz-invented-control-9481.html"; do
  echo -n "$u -> "; curl -s -o /dev/null -w "%{http_code}\n" --max-time 15 "$u?cb=$RANDOM$$"
done
```
Gotchas: the sibling-200 and invented-404 controls are MANDATORY (a parked domain
200s every path); cache-bust with a query string (the worker ignores it when
building the object key, the CDN cache keys on it); status codes only — a
byte-compare here walks into the CDN-adds-bytes landmine.

## Who does the reconciler service next? (read BEFORE any thought of forcing)

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT s.domain, c.last_checked_at FROM sites s
  LEFT JOIN site_publish_checks c ON c.site_id=s.id
  WHERE s.publish_target IS NOT NULL ORDER BY c.last_checked_at ASC NULLS FIRST;"
```
Gotcha: forcing stamps the forced site to the BACK (LANDMINES). The th2 rollout
needs no forcing — do not touch `last_triggered_at`.

## Watch the th2 convergence post-roll

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT domain, left(published_hash,12) AS hash, published_at FROM sites WHERE publish_target IS NOT NULL;"
```
Done when both rows show `th2:` and `published_at` has advanced past the roll.
Gotcha: `site_publish_checks.last_checked_at` moving is NOT evidence of a
publish ("checked" ≠ "published") — read `published_hash`/`published_at`.

## Tests for this change

```bash
go test ./platform/publish/ ./platform/orchestration/actions/ -run 'B2Worker|PublishSite|TreeHash' -count=1
go test ./cmd/config-key-audit/ -run 'BudgetCron' -count=1   # parity: check.py literal vs registry
./scripts/audit-optional-key-budget.sh                        # budget itself (N=10)
```
Gotcha: after editing `check.py`, the CLUSTER keeps the old literal until the
kustomize overlay is re-applied (CLAUDE.md RFC_022 note).
