# 233 — B2 application key pair and CLIENTS_DB_PASSWORD logged in plaintext at INFO

**Status: FIX IS LIVE on the v1.0.1274 fleet, verified at the pod 2026-08-09 —
with ONE pod still leaking and a rotation-ordering consequence. Stays OPEN.**

> **VERIFIED LIVE 2026-08-09, v1.0.1274.** Positive string the fix ADDED and
> negative string it REMOVED, same exec, every replica:
>
> | pod | `access_key_present` | `B2_APPLICATION_KEY from env` |
> |---|---|---|
> | agent-chassis ×2 (both replicas) | 1 | 0 |
> | thunder-adapter (the original sighting) | 1 | 0 |
>
> `CLIENTS_DB_PASSWORD_present` = 1 on both chassis replicas. A fourth check
> (`strings … grep "^CLIENTS_DB_PASSWORD$"` = 0) is NOT cited as evidence: Go
> merges string literals into one blob, so `strings` line boundaries are an
> artefact, and that check would read 0 whether or not the literal survived.
>
> **STILL LEAKING: `render-audit-adapter`**, which runs
> `browser-runner-adapter:v1.0.1194` — 80 tags behind, pinned since the pod was
> created (`0143a693e`). Verified at the binary with `grep -a` (that image has
> no `strings`): leak string = **1**, `access_key_present` = **0**, positive
> control `NewS3Client` = 3 — so the grep mechanism works and the absence is
> real. **The credential is in its retained log buffer right now**: 1 matching
> line out of 35 total retained (counted, never printed). Root cause is a
> release-coverage gap, filed separately as **`bugs_open/237`**.
>
> **Exposure now bounded to that one pod.** Ten `cmd/` targets link
> `platform/storage`; of those, only `component-render-check` is also pinned
> pre-fix (v1.0.1258), and it **never constructs a client** — no `NewS3Client`
> call site in `cmd/component-render-check/`, and its last CronJob run logged
> 0 leak lines out of 3 total (a non-empty log, so the zero is meaningful).
> Every other storage-linking service is on v1.0.1274.
>
> **⚠ ORDERING CONSTRAINT FOR THE OWNER'S KEY ROTATION (this is the part that
> bites).** `render-audit-adapter` reads the credential from its env at
> **startup**. Leaving it as-is, it holds the OLD key and logs it — bad but
> static. **Rotating the B2 keys while that pod is still on v1.0.1194 means the
> NEXT restart writes the NEW key into its log in plaintext**, re-creating the
> exposure the rotation was meant to end. So: **roll `render-audit-adapter` to
> a ≥v1.0.1274 image BEFORE rotating**, and do not simply `kubectl delete pod`
> it beforehand — its overlay pins the old tag, so it would come back on the
> same leaking image and emit a FRESH credential line. `bugs_open/237` has the
> one-line overlay fix.

Original status line, kept for the trail: *FIXED IN TREE 2026-08-09 — inert
until the next fleet roll* (Go change; the leak keeps firing on every running
pod built before the fix ships).

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
(commit `43c1801d6` carries it as `Council-Submitted:`).
**Verdict: APPROVED, round 1, all reviewers** (council_report row
2026-08-09 11:44 UTC; the trailer is credited by 098 at report time —
forward-only forbids amending it to `Council-Reviewed:`). Query:

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

---

## CLOSED IN SUBSTANCE 2026-08-10 — the last leaking pod is gone; ROTATION IS UNBLOCKED

`render-audit-adapter`, the sole remaining leaker, rolled to
`browser-runner-adapter:v1.0.1280` at 15:45Z (its first roll ever — see
`bugs_open/237`, whose makefile fix put it in the release path). Verified at that
pod, the checks this file specified:

| check | want | got |
|---|---|---|
| `access_key_present` in the binary | ≥1 | **1** |
| `B2_APPLICATION_KEY from env` in the binary | 0 | **0** |
| positive control `NewS3Client` | >0 | **3** |
| credential lines in the live log buffer | 0 | **0** |
| total log lines (control — proves the zero is meaningful) | >0 | 11 |

`grep -a` rather than `strings`: that image has no `strings` binary.

**The rotation-ordering constraint recorded above is now SATISFIED and no longer
applies.** No pod in the fleet runs a binary that logs these values, so rotating
the B2 application key pair and `CLIENTS_DB_PASSWORD` can happen whenever the
owner chooses, with no restart hazard.

**What is NOT closed by this:** the credentials were exposed in pod logs from
2025-10-28 to 2026-08-10 and should still be presumed disclosed to anyone with
`kubectl logs` access in that window. Rotation remains an owner decision; this
entry only removes the timing constraint on it.

Kept in `bugs_open/` per the owner ruling of 2026-08-06 (a finished bug stays).

---

# UPDATE 2026-09-03 — the code leak is CLOSED FLEET-WIDE, and this file's own banner has been blocking the remediation for 15 days

Verified today by the `bugfix_329_takeover_claim` lane, which picked this up as an unowned bug
(no lane directory, no live session, untouched 23 d). **Nothing here was inferred from a roll —
every claim below is a probe or a census with a control.**

## 1. The "STILL LEAKING" pod is fixed — probed at the binary, three ways

`render-audit-adapter` now runs `browser-runner-adapter:**v1.0.1359**` (overlay
`kustomization.yaml:19` pins that tag), not the `v1.0.1194` this file records.

| probe on `/proc/1/exe` | result | reads as |
|---|---|---|
| `NewS3Client` (unrelated to the fix) | PRESENT | **control** — the `grep -a` mechanism works |
| `access_key_present` (added by the fix) | **PRESENT** | the fix **shipped** |
| `B2_APPLICATION_KEY from env` (removed by the fix) | **ABSENT** | the leak is **gone** |

The control is what makes this a measurement: a broken probe returns ABSENT for all three and reads
as "still leaking".

## 2. Fleet census, with the control that stops it being vacuous `[MEASURED 2026-09-03 ~16:2xZ]`

Every Deployment **and** CronJob image:

- **35** images whose version tag parsed — **min tag = max tag = `v1.0.1359`**. The whole estate is
  on one tag, 85 releases past the `v1.0.1274` fix.
- **0** parsed images older than `v1.0.1274`.
- **19** images the version regex could not parse — all third-party (`postgres:16-alpine`,
  `ollama/ollama:latest`, `edoburu/pgbouncer`, `linuxserver/wireguard`, `bitnami/kubectl`). **None is
  our Go code, so none can carry our leak string.** Recorded as *unjudged* rather than folded into
  the zero.

⚠ Without the min/max control, "0 images older than v1.0.1274" would read identically if the regex
had matched nothing at all.

**So no running workload can write a new leak line.**

## 3. ⚠ THIS FILE'S BANNER HAS BEEN OBSOLETE FOR 15 DAYS, AND IT GATES AN OWNER ACTION

The banner says *"STILL LEAKING: `render-audit-adapter`, which runs `browser-runner-adapter:v1.0.1194`
— 80 tags behind, pinned since the pod was created"*, and roots the pin in **`bugs_open/237`**.

**`237` was CLOSED on 2026-08-19** — *"fixed, LIVE and verified on `v1.0.1314`. Owner ruled to
close."* `render-audit-adapter` was unfrozen 2026-08-10 and the class fix (one declaration per set +
`check-release-coverage`, **BLD-022**) shipped 08-17.

**So the blocker was discharged fifteen days ago and this file never said so.** This is the
`a-closed-blocker-keeps-being-obeyed` class, second recorded instance: a stale status line does not
merely mislead, it **prevents the correct action**, because the action it gates looks premature.

## 4. THE ROTATION-ORDERING CONSTRAINT IS DISCHARGED — rotating is now safe

This file told the owner:

> *"Rotating the B2 keys while that pod is still on v1.0.1194 means the NEXT restart writes the NEW
> key into its log in plaintext … So: roll `render-audit-adapter` to a ≥v1.0.1274 image BEFORE
> rotating."*

**That roll has happened** (§1). The named reason to wait is gone. **A rotation performed today
cannot be re-leaked by any pod in the fleet**, because no image contains the emitting code.

## 5. What is NOT established, stated plainly

- **Whether the exposed credentials have been rotated: `[UNVERIFIED]`, and I cannot verify it.** The
  secret `personae-storage-secrets` carries `creationTimestamp 2025-08-02` and no rotation
  annotation — but ⚠ **`creationTimestamp` survives an in-place update**, so it cannot distinguish
  "never rotated" from "rotated in place". Per the owner ruling of 2026-08-23 I did not read any key
  value. **This is the owner's to answer, and it is the only thing standing between this file and
  closure.**
- **Whether any retained log buffer still holds the credential: NOT READABLE, not "clean".** The
  three pre-fix `component-render-check` pods (25–27 d old) return **0** leak lines — out of
  **0 total lines**. ⚠ **A zero from an empty log is vacuous**, which is a caveat this file's own
  earlier check was careful about ("a non-empty log, so the zero is meaningful"). Their logs have
  expired; the question cannot now be answered for them. The `render-audit-adapter` pod that *did*
  hold one has been replaced, so that buffer is gone with it.

## 6. Disposition

**The CODE defect is fixed and live everywhere** — that half meets the `bugs_closed/` bar. **Left
OPEN deliberately**, because moving this file is exactly how the outstanding remediation gets
forgotten, and the remediation is the part that actually ends the exposure: **the keys were emitted
in plaintext and should be treated as compromised until rotated.** Close it when the owner confirms
the rotation, or rules that it is not needed.

---

# UPDATE 2026-09-04 — the `zap.Any` residual this file explicitly did NOT claim is now swept and CLEAN

The original fix bounded its class by grep and said so honestly:

> *"Not swept: secrets reaching logs via struct dumps (`zap.Any` of a config object) … a fleet-wide
> `zap.Any` audit is a bigger job than this fix and is NOT claimed here."*

Swept 2026-09-04. **Nothing found — and every step below carries the control that stops its zero
being vacuous.**

## 1. Do any live structs HOLD a secret value, and are they logged?

Structs with a `string` field named `*Password|Secret|Token|Credential|ApiKey|AccessKey|SecretKey`
(excluding `*EnvVar`, which holds a NAME), non-test, live code only:

- `internal/tools-api/config/config.go:17` — `AnthropicAPIKey string`
- `internal/auth-service/…` — `NewPassword`, `CurrentPassword`, `AccessToken`, `RefreshToken`
  (request/response DTOs)

**Neither family is ever logged.** `zap.Any(<cfg|conf|Config>)` and `%+v` of a config: **0 hits** in
`internal/tools-api/`; request-struct logging: **0 hits** in `internal/auth-service/`.

⚠ **CONTROL, because a zero from a broken pattern reads identically:** the same expression run
repo-wide returns **15** hits and correctly finds the known `zap.Any("config", params.StepConfig)`
dumps at `spawn_actions.go:624/630/638`. The pattern can find a positive; these are real absences.
(Matches under `docs/` — the `noted-engine` and `idea.uk` bundles — are documentation copies, not
built code, and are excluded rather than silently folded into the zero.)

## 2. The 15 sites that DO dump a config — can a step config carry a secret VALUE?

Those sites dump `params.StepConfig`, i.e. `agent_definitions.default_config`. So the question is
whether any live agent config holds an inline credential.

First pass, text regex over every active agent's config for a secret-shaped key with a ≥12-char
string value: **81 agents** `[MEASURED 2026-09-04]`, against a control of **67** agents carrying a
`prompt_template` — so the mechanism works and 81 is not a pattern failure.

⚠ **81 is entirely FALSE POSITIVES, and the count alone would have been alarming and wrong.**
Extracting the matching **key names** (never the values — owner ruling 2026-08-23) gives, exhaustively:

| key | occurrences |
|---|---|
| `api_key_env_var` | 160 |
| `secret_key_env_var` | 2 |
| `access_key_env_var` | 2 |

**Every match is an `*_env_var` key holding an environment variable NAME.** That is precisely the
convention this bug's own fix converged on ("values → presence booleans", env var *names* logged).
**Zero inline secret values in live agent configs**, so the config-dumping sites cannot leak one.

⚠ This is the same false-positive class the original sweep already hit once
(`thunder_ssh_exec_dispatch.go:315`, where `token` was a command-template token NAME). Two sightings
now: **a credential-shaped KEY NAME is not a credential**, and the discriminator is to extract names
rather than count matches.

## 3. So what remains before this file can close

**Nothing technical. One owner decision.**

- **Code: FIXED AND LIVE**, re-verified at the binary on **`v1.0.1360`** 2026-09-04 on both
  `render-audit-adapter` and `agent-chassis` — control PRESENT / fix `access_key_present` PRESENT /
  leak `B2_APPLICATION_KEY from env` ABSENT. Fleet: **36 images, min tag = max tag = v1.0.1360**, zero
  pre-fix.
- **Class: BOUNDED.** The grep spellings in the original file, plus §1–§2 above. The one stated gap is
  now closed.
- **⛔ REMAINING — ROTATION, and it is the whole of it.** `git log -S` dates the leak to `9260b86ed`,
  **2025-10-28**. The B2 application key pair and `CLIENTS_DB_PASSWORD` have therefore been landing in
  pod logs for **over ten months** and must be presumed disclosed to anyone with `kubectl logs` access
  in that window. The original file deferred this to the owner deliberately; the ordering constraint
  that later blocked it is **discharged** (see the 2026-09-03 update). B2 rotation touches
  `personae-default-secrets` and the GitHub-secrets copy used by the B2 CLI; DB rotation touches every
  service's DB wiring.
  **[UNVERIFIED] whether either has been rotated, and this session cannot check it** — the secret's
  `creationTimestamp` survives in-place updates, and reading a key value is forbidden.

**Close this file when the owner confirms the rotation, or rules it unnecessary. Not before, and not
on the strength of the code fix alone** — the code stopped the emission; it did not un-emit ten
months of it.
