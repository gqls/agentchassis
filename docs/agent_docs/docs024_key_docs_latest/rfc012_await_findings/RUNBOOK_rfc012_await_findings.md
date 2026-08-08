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

## Proving the B conversions are still live after a roll

Run this after ANY fleet roll. The pair is discriminating on purpose: `f930de86b` reworded
one log line singular→plural, so a POS and a NEG must both hold **in the same binary** —
no stale image and no lucky substring satisfies both.

```bash
for sel in app=agent-chassis app=dynamic-agent; do
  POD=$(kubectl -n ai-persona-system get pods -l $sel --field-selector status.phase=Running \
        --no-headers -o custom-columns=:metadata.name | head -1)
  echo "== $sel -> $POD"
  kubectl -n ai-persona-system exec "$POD" -- sh -c '
    printf "POS(plural, added)  = "; strings /app/agent-chassis | grep -cF "failed to write some discovery check error records"
    printf "NEG(singular, gone) = "; strings /app/agent-chassis | grep -cF "failed to write discovery check error record"
    printf "NEG(envelope, gone) = "; strings /app/agent-chassis | grep -cF "content_data envelope: failed to write record"
    true'   # <- see gotcha 2
done
```
Expect `1 / 0 / 0` per pod. Measured on v1.0.1262, v1.0.1263 and **v1.0.1264** (2026-08-08,
both labels).

Gotcha 1 — **`-l app=agent-chassis` is 2 pods and it is NOT the fleet.** Raised by the
council's `debug_historian` seat against this lane's own evidence, and it was right:
measured 2026-08-08, **42 pods run an agent-chassis image under FOUR app labels**
(`dynamic-agent` 38, `agent-chassis` 2, `business-intel` 1, `vet-intel` 1), of which 19 are
Running — the rest are completed job pods, and exec'ing one fails with "cannot exec into a
container in a completed pod", not with a wrong answer. So "both replicas verified" is a
claim about 2 of 19. What licenses generalising is **tag uniformity**, which is a separate
query and should be run with the greps:
```bash
kubectl -n ai-persona-system get pods -o json | python3 -c "
import json,sys,collections
d=json.load(sys.stdin)
print(collections.Counter(c['image'].split(':')[-1] for p in d['items'] for c in p['spec'].get('containers',[]) if 'agent-chassis' in c['image']))"
# one tag for all of them (v1.0.1264 x42 on 2026-08-08) => the two pods grepped stand for the rest
```

Gotcha 2 — **`grep -c` prints `0` but EXITS 1, so the NEG half poisons the exec's status.**
The same seat objected that a negative control read as `=0` might be capturing an empty or
errored result. Measured: it is not — `grep -cF` genuinely prints `0` on no match
(`stdout=[0] exit=1`), so the printed number is trustworthy. **The exit code is not.** With
a NEG grep last, `kubectl exec` returns `command terminated with exit code 1` on a
*correct* result, so any wrapper branching on exec status reads a pass as a failure. Hence
the trailing `true` above. Do not "fix" this by moving a POS grep last — that hides a real
exec failure instead.

## Answering a council reuse objection with a measurement, not an argument

The `--shared-output-fields` descent was swapped onto `validation.WalkSteps` (`867037f5a`).
The pattern worth reusing is the A/B, because "my refactor changed nothing" is otherwise
just a claim:

```bash
SP=<scratch>; git archive HEAD | tar -x -C $SP/oldtree     # committed HEAD, no WIP
# ONE export, both binaries, identical bytes in:
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc "
SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow'))
FROM agent_definitions WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false
  AND is_active AND default_config ? 'workflow';" > $SP/live_agents.json
(cd $SP/oldtree && go run ./cmd/config-key-audit --shared-output-fields --ack scripts/shared_output_fields_ack.txt < $SP/live_agents.json) > $SP/old.json
go run ./cmd/config-key-audit --shared-output-fields --ack scripts/shared_output_fields_ack.txt < $SP/live_agents.json > $SP/new.json
diff <(python3 -m json.tool $SP/old.json) <(python3 -m json.tool $SP/new.json)   # empty == no-op
```

Gotcha 3 — **do NOT pipe the psql export straight into the tool.** A truncated
`kubectl exec` exits 0, so a short read arrives at the parser looking like a small fleet.
Write it to a file and assert the byte count first (1,097,081 bytes / 177 agents on
2026-08-08). The tool refuses a 0-agent decode by design, which is what caught this — but
it cannot catch a *partial* array that still parses.

## Not-yet-built (see the handoff)

The online CronJob for (d). **Everything else on the old list here is DONE**: the 18
hand-copied INSERT conversions shipped (`f930de86b`, live) and the reader census artefact is
delivered (`architecture_review/CENSUS_2026-08-07_rfc012_await_step_readers.md`).
