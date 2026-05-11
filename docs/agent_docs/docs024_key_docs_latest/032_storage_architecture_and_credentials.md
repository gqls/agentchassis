# 012 — Storage Architecture and Credentials

Operational reference for object storage in agentchassis. Read before adding any agent that uploads, downloads, or otherwise touches S3-compatible storage.

The codebase has solved storage already. This document maps what exists so new code uses it instead of reinventing.

---

## TL;DR

- The cluster runs on **Backblaze B2** (S3-compatible). Not MinIO.
- B2 credentials live in the K8s secret **`personae-platform-secrets`** as keys `B2_APPLICATION_KEY_ID` and `B2_APPLICATION_KEY` (per `011_database_and_infrastructure.md`).
- **For new actions: construct your own S3 client per-call via `storage.NewS3Client(...)`.** Don't rely on `params.StorageClient`.
- `params.StorageClient` is **deprecated for new code** (see "Deprecation note" below).
- Spawned worker pods get B2 env vars forwarded automatically via `isStorageEnabledAgent` — keep that gate maintained.

---

## Deprecation note — `params.StorageClient` (the injected S3 client)

**Status: deprecated for new code.**

The chassis used to construct an S3 client at startup and inject it into every action via `ActionParams.StorageClient`. In practice, this client is `nil` in many chassis pods because:

- The chassis init reads `IMAGE_BUCKET` from env at startup. If empty, `storage.NewS3Client` returns an error and `coordinator.storageClient` stays nil for the lifetime of that pod.
- Different operations need different buckets (image, finetuning, etc.) and may legitimately run in pods configured for one bucket but not another.
- A chassis pod that starts before its env-from-secret has propagated will permanently run with nil storage.

The existing image-handling actions (`UploadToS3Action`, `RouteStorageAction`, `storeToS3`, `storeToB2` in `platform/orchestration/actions/storage_actions.go`) all bypass `params.StorageClient` and construct their own S3 client per call. That is the canonical pattern.

**Do this:**

```go
import (
    cfgpkg "github.com/gqls/agentchassis/platform/config"
    "github.com/gqls/agentchassis/platform/storage"
)

storageConfig := cfgpkg.ObjectStorageConfig{
    Provider:        "s3",
    Bucket:          "<bucket-name>",
    AccessKeyEnvVar: "B2_APPLICATION_KEY_ID",
    SecretKeyEnvVar: "B2_APPLICATION_KEY",
}
s3Client, err := storage.NewS3Client(ctx, storageConfig, *params.Logger)
if err != nil {
    return nil, fmt.Errorf("failed to construct S3 client: %w", err)
}
uri, err := s3Client.Upload(ctx, key, contentType, body)
```

**Don't do this:**

```go
if params.StorageClient == nil {
    return nil, fmt.Errorf("storage client required")  // ← brittle
}
uri, err := params.StorageClient.Upload(ctx, ...)       // ← unreliable
```

The deprecation is at the field-usage level. The field still exists on `ActionParams` for backward compatibility with existing actions that already use it; new actions should not depend on it. Existing usages can migrate opportunistically when the action is being modified for other reasons.

Reference example of the inline pattern in a non-image agent: `platform/orchestration/actions/prepare_training_data_action.go` (training-data-preparer worker, flywheel C phase 2).

---

## What exists

There are three distinct paths through which an action ends up with an S3 client. They overlap. Knowing which one your code is hitting matters.

### Path A — Inline per-call construction (the canonical pattern)

Where: anywhere — e.g. `storage_actions.go`, `prepare_training_data_action.go`.

The action calls `storage.NewS3Client(ctx, cfg, logger)` directly and uses the returned client. The config struct (`ObjectStorageConfig`) names which env vars to read for credentials. `NewS3Client` reads them via `os.Getenv`.

Used by every existing image action and every new training-pipeline action. **This is what you should write.**

### Path B — Injected `params.StorageClient` (deprecated for new code)

Where: chassis startup → `coordinator.storageClient` → `ActionParams.StorageClient`.

Constructed once at chassis pod startup. If startup found `IMAGE_BUCKET` and B2 env vars, the client is non-nil. Otherwise nil for the pod's lifetime. See deprecation note above. Existing actions that use this still work in pods where startup succeeded; don't break them, but don't write new ones.

### Path C — Spawn-time env propagation (enabling logic)

Where: `platform/orchestration/actions/spawn_actions.go` lines 2495–2573.

When a parent agent spawns a child via `spawn_agent`, this gate decides whether the child Job pod gets storage env vars in its environment:

```go
if isStorageEnabledAgent(agentDef.Type) ||
   agentDef.Category == "orchestrator" ||
   agentDef.Category == "code-driven" {
    // forward credentials from spawning pod's env to spawned pod's env
    b2KeyId := os.Getenv("B2_APPLICATION_KEY_ID")
    // ... forward B2 keys + S3_ENDPOINT/S3_REGION/IMAGE_BUCKET/ASSETS_BUCKET
}
```

This is **enabling logic**. It doesn't construct an S3 client itself. It makes sure spawned pods have the env vars so Path A (inline construction) inside that pod's actions can read them via `os.Getenv`.

The gate forwards from the **spawning pod's env**. Keep the chassis Deployment manifest's `envFrom: secretRef: name: personae-platform-secrets` mounted, otherwise nothing gets forwarded.

---

## Where credentials live

Per `011_database_and_infrastructure.md` §Credentials:

> All database passwords are stored in the `personae-platform-secrets` K8s secret.
>
> | `B2_APPLICATION_KEY_ID` | Backblaze B2 access key |
> | `B2_APPLICATION_KEY`    | Backblaze B2 secret key |
>
> Managed by Terraform in `047-base-configs` with values in `terraform.tfvars.secret`.

---

## Pattern for a new storage-using agent

For an agent of type X that needs to upload/download:

### 1. The agent's Go action constructs its own S3 client

```go
storageConfig := cfgpkg.ObjectStorageConfig{
    Provider:        "s3",
    Bucket:          "<bucket-name>",   // or read from step config
    AccessKeyEnvVar: "B2_APPLICATION_KEY_ID",
    SecretKeyEnvVar: "B2_APPLICATION_KEY",
}
s3Client, err := storage.NewS3Client(ctx, storageConfig, *params.Logger)
```

The bucket can be a step-config knob if multiple agents share the action and use different buckets. Hardcoded is fine when the agent has one job.

### 2. Add the agent type to `isStorageEnabledAgent`

In `platform/orchestration/actions/spawn_actions.go`, the storageAgents list. This authorises spawn-time env forwarding for the agent type. Even though Path A doesn't strictly need `params.StorageClient`, it does need `os.Getenv("B2_APPLICATION_KEY_ID")` to return a value — which requires the spawn-time forwarding (Path C).

### 3. The bucket must exist on B2

Created out-of-band in the Backblaze dashboard. Not Terraformed.

### 4. Spawn-and-call, not direct trigger

Worker agents that do storage work belong in spawned Job pods, not the chassis pod. Trigger via the parent orchestrator (e.g. `model-trainer`), not via direct trigger to `system.agent.generic.requests`.

Direct trigger runs the workflow in-chassis, which:
- Skips Path C's spawn-time env injection (different code path)
- Picks up whatever env the chassis pod has, which may or may not match the agent's needs
- Doesn't get the per-agent `IMAGE_BUCKET` override

This is fine for orchestrators (whose only job is to spawn children); not for substantive workers.

---

## Verification checklist for a new storage-using agent

- [ ] Agent type appears in `isStorageEnabledAgent` (spawn_actions.go)
- [ ] Bucket exists on Backblaze (`b2 ls` from a workstation)
- [ ] After spawn, pod env contains `B2_APPLICATION_KEY_ID` and `B2_APPLICATION_KEY` non-empty:
  ```bash
  kubectl -n ai-persona-system exec <spawned-pod> -- env | grep -E "B2_|IMAGE_BUCKET"
  ```
- [ ] Action constructs its own S3 client via `storage.NewS3Client` (NOT `params.StorageClient`)
- [ ] Action's first storage operation succeeds without `NoSuchBucket` or `InvalidAccessKeyId`

---

## What NOT to do

- Don't rely on `params.StorageClient` being non-nil — it's deprecated and unreliable.
- Don't add a new method like `UploadToBucket` to the storage Client interface. Bucket-per-agent is handled by passing different `cfg.Bucket` values to `NewS3Client`.
- Don't put credentials in the workflow JSON or `agent_config.storage_config` for new agents. Hardcode the env-var names in your action; the env vars themselves come from the secret.
- Don't add training-pipeline-specific env-var forwarding to spawn_actions.go. The existing block already handles any agent in the storageAgents list.
- Don't construct two S3 clients in one action invocation. Build once, reuse for the operations.
