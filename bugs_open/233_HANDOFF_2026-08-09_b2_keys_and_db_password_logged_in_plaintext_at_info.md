# 233 — B2 application key pair and CLIENTS_DB_PASSWORD logged in plaintext at INFO

**Status: FIXED IN TREE 2026-08-09 — inert until the next fleet roll** (Go change;
the leak keeps firing on every running pod built before the fix ships).

## Mechanism

Two debug log lines emit live secret VALUES at INFO level:

1. `platform/storage/s3.go:36-37` (pre-fix numbering) — `NewS3Client` logged
   `os.Getenv("B2_APPLICATION_KEY_ID")` and `os.Getenv("B2_APPLICATION_KEY")`
   as plain `zap.String` fields on every client construction. This is the
   shared constructor: 8 call sites — `platform/agentbase/agent.go:318`
   (chassis startup), `platform/orchestration/actions/storage_actions.go:95,612`
   (per-action client build, so per invocation, not per startup),
   `prepare_training_data_action.go:94`, and the thunder / imagegenerator /
   browserrunner / webscrape adapters. The finetuning handoff recorded it as
   "thunder-adapter logs B2 credentials on startup" — true but an undercount;
   the leak fires in **every service that touches object storage**.
2. `platform/orchestration/actions/spawn_actions.go:2734` (pre-fix numbering) —
   the `DEBUGaa` env-dump before building a spawned agent's k8s Job logged
   `os.Getenv("CLIENTS_DB_PASSWORD")` in plaintext, firing on **every
   dynamic-agent spawn**. Found by this fix's own detection grep (below), not
   previously recorded anywhere.

## Exposure window

`git log -S 'B2_APPLICATION_KEY from env' -- platform/storage/s3.go` →
introduced `9260b86ed` 2025-10-28 ("fixing s3 env vars"). So both B2 values
have been landing in pod logs for over nine months, readable by anything with
`kubectl logs` access on `ai-persona-system` (and by any log aggregation, if
one is ever attached).

**Rotation is an OWNER decision, deliberately not taken here**: the B2
application key pair and the clients-DB password should be presumed
disclosed to anyone who has had log access since 2025-10-28. Rotating the B2
key touches `personae-default-secrets` (and the GitHub-secrets copy used by
the B2 CLI per LANDMINES); rotating the DB password touches every service's
DB wiring. Neither is bundled into this fix.

## The fix (both sites, one class: values → presence booleans)

- `s3.go` — the debug block replaced with a startup log of the env var NAMES
  (`cfg.AccessKeyEnvVar`/`cfg.SecretKeyEnvVar`), presence booleans for both
  credentials, endpoint and bucket. Also fixes the pre-existing label bug
  (the old line 34 logged `AccessKeyEnvVar` under the `SecretKeyEnvVar` label).
- `spawn_actions.go` — the `CLIENTS_DB_PASSWORD` string field replaced with a
  presence boolean; the rest of the env-dump (hosts, ports, users, db names —
  non-secret) left as it was.

Behaviour change is log-output only; no control flow touched.
`go build ./platform/storage/ ./platform/orchestration/actions/` clean.

## Class bounding — the spellings tried

A grep proves absence only for the spelling it searches, so recording them:

- `zap.String(...os.Getenv(...KEY|SECRET|TOKEN|PASSWORD` → the two sites above,
  nothing else.
- `zap.String("...", accessKey|secretKey|apiKey|password|token|secret)` →
  one hit, `thunder_ssh_exec_dispatch.go:315` — **false positive**, `token`
  there is a command-template token NAME (e.g. `scripts_url`), not a credential.
- `Printf|Println|Sprintf(...os.Getenv(...KEY|SECRET|TOKEN|PASSWORD` → zero.
- Presence-boolean logging (`has_b2`, `has_aws` in `spawn_actions.go:2563-2566`)
  already existed and is the pattern the fix converges on.

Not swept: secrets reaching logs via struct dumps (`zap.Any` of a config
object). A quick check of `ObjectStorageConfig` shows it carries env var
*names*, not values, so the constructor path is clean; a fleet-wide `zap.Any`
audit is a bigger job than this fix and is NOT claimed here.

## Why no 090 run (declared per the 2026-07-31 owner ruling)

The defect is local and self-evidencing: the log line prints the secret, and
removing it stops that — watch-it-fail/watch-it-pass at the pod log. No
cross-cutting root-cause claim is being made beyond what the recorded greps
show; the one structural observation (shared constructor ⇒ every storage
service leaks, not just thunder-adapter) is direct from the caller list above.

## How to verify once an image rolls

Per the fleet rule (a roll is not evidence the fix shipped), grep the pod for
a string the change ADDED and one it REMOVED, same exec:

```
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/<binary> | grep -c "access_key_present"'   # expect ≥1
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/<binary> | grep -c "B2_APPLICATION_KEY from env"'  # expect 0
```

Then the live check: a fresh startup/spawn log line shows
`access_key_present: true` / `CLIENTS_DB_PASSWORD_present: true` and no key
material.

## Council trail

Submitted 2026-08-09, correlation `7490388d-c945-42c0-b3c4-c452741a10cd`
(commit `43c1801d6` carries it as `Council-Submitted:`). Verdict:

```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='7490388d-c945-42c0-b3c4-c452741a10cd'
  AND kind='council_report' ORDER BY created_at;
```

## Provenance

Recorded as a trap in
`docs/agent_docs/docs024_key_docs_latest/finetuning/HANDOFF_2026-08-08_continue_here.md`
§4 (found incidentally 2026-08-04 by the finetuning lane, deliberately left
unbundled). Fixed by this lane 2026-08-09; the handoff carries a dated
correction pointing here.
