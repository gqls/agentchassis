# RUNBOOK — rfc012_await_findings

Commands that were hard to get right, with the gotcha attached. Update HERE.

## The (d) census — run it, and read the ratchet

```bash
bash scripts/audit-shared-output-fields.sh          # raw: exit 1 while ANY pair exists
# the ratchet form (what the standing job runs) — exit 0 while only acked pairs reproduce:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow'))
FROM agent_definitions
WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active
  AND default_config ? 'workflow';" \
 | go run ./cmd/config-key-audit --shared-output-fields --ack scripts/shared_output_fields_ack.txt
```

2026-08-06 baseline: **176 agents, 0 new / 2 acked / 0 stale, exit 0.**

Gotcha 1: `--ack` must be argv[2]/[3] (`--shared-output-fields --ack <file>`), not later.
Gotcha 2: a clean run here proves the FLEET, not the detector — the detector's ability to
fire is proven by `go test ./cmd/config-key-audit/ -run TestSharedOutputs` (it fires on
192's own shape and LOSES the finding when the `config.then_step` edge is severed). Never
cite a green live run alone.
Gotcha 3: empty slices are initialised, not nil — a consumer doing `len(json null)` crashes.
That bit once already.

## Re-deriving the 13 routing keys (do this before trusting the literal)

The RFC addendum's "13 keys" arithmetic is loose — it names 11 config keys after
`then_step`/`else_step`. Measured against the live fleet there are **13 config keys**, and
the one the addendum omits is a **config-level `error_step`** (158 occurrences, distinct
from the top-level `Step.ErrorStep` field):

```sql
WITH cfg AS (
  SELECT s.value->'config' AS c
  FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s
  WHERE ad.is_active AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL)
SELECT k.key, count(*) FROM cfg, jsonb_object_keys(cfg.c) k(key)
WHERE cfg.c IS NOT NULL AND k.key LIKE '%\_step' GROUP BY 1 ORDER BY 2 DESC;
```
(This query reads TOP-LEVEL steps only — fine for *discovering* key names, wrong for the
census itself, which must descend into sub-workflows. That is why the Go side walks.)

## The B helpers — how to use them from an action

```go
// one durable row
LogActionError(ctx, params, siteID, domain, "my_action", "MY_CODE", "warning", msg, payload, logger)

// findings that must survive an AWAIT — call BEFORE the dispatch
attempted, recorded := LogActionFindings(ctx, params, siteID, domain, "my_action", findings, logger)
audit["conditions_recorded"] = recorded
if attempted != recorded { audit["conditions_lost"] = attempted - recorded }
```

Gotcha: a nil `Context` map marshals to JSON **`null`**, not `{}` — `json.Marshal` on a nil
map returns `null`, so the historical writer's `contextJSON == nil` guard never fired. The
behaviour is preserved deliberately (byte-compatibility); pin `"null"` in tests, not `"{}"`.

## Not-yet-built (see the handoff)

The online CronJob for (d); the 18 remaining hand-copied INSERT conversions; the reader
census artefact.
