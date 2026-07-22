# RUNBOOK — model directory pipeline

Commands that were hard to get right, with the gotcha attached. Updated the
moment a command changes — not from scrollback later.

## DB access

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Site under pilot

`ai-agent-orchestration.com` = site_id `2a8ebf9c-20a2-4c39-b191-840b012371da`.

## Checking migration numbering before filing a new one

Another session may have claimed the next number since this doc was last
updated — always re-check IMMEDIATELY BEFORE APPLYING, not just when drafting
the file (a collision was caught this way on 2026-07-22: `191` was taken
concurrently by an unrelated session between draft and apply — ours is `192`):

```
ls docs/agent_docs/sql_for_agents/ | grep -oE '^[0-9]+' | sort -n | tail -5
git status --short docs/agent_docs/sql_for_agents/   # catches untracked concurrent files too
git log --oneline -5 -- docs/agent_docs/sql_for_agents/
```

## Applying a migration

```
./scripts/migration/run-migrations.sh            # dry run, lists pending
./scripts/migration/run-migrations.sh --apply    # applies + records, in order
```
This applies ALL pending files >= baseline in order, not just yours — check
the dry-run output first to see whether anything else is pending that isn't
yours (another session's filed-but-unapplied migration).

## Verifying the site_specs opt-in write took

```sql
SELECT aspect, is_current, jsonb_pretty(data->'content_features'->'model_directory')
FROM site_specs
WHERE site_id = '2a8ebf9c-20a2-4c39-b191-840b012371da'
  AND aspect = 'classification' AND is_current;
```

## Verifying a citation is still live (manual spot-check, mirrors verifyCitationLive)

```
curl -sL '<citation.url>' | grep -F '<citation.quote>'
```
Empty output = the quote is gone; the automated freshness sweep should have
flipped that claim's `status` to `citation_lost`, not left it `found`.

## Checking directory_claims state

```sql
SELECT de.slug, dc.field, dc.value, dc.status, dc.verified_at
FROM directory_claims dc JOIN directory_entities de ON de.id = dc.entity_id
WHERE dc.is_current
ORDER BY de.slug, dc.field;
```
