# 010 — Automated Action Build Pipeline

How LLM-written Go actions get from source to running containers without GitHub Actions or manual intervention.

---

## Problem

Workflows (JSON) and agent definitions (SQL) are already hot-loadable — write to the database and they're live. Go actions are the bottleneck. Every new action requires a code change, image build, and deploy. As LLMs get better at writing actions, the human in that loop becomes the slowest part.

## Approach

A compiler pod runs inside the Kubernetes cluster. It watches for new or changed action source files in git, pulls the code, compiles the full agent-chassis binary, runs tests (including LLM-generated tests), builds a container image, and pushes it to the registry. Deployment is either automatic or gated depending on the HITL setting.

Git remains the source of truth. The compiler pod replaces GitHub Actions as the build system.

---

## Pipeline Stages

```
LLM agent writes action + test
        │
        ▼
  git push to repo
  (branch: auto/<action-name> or main)
        │
        ▼
  ┌─────────────────────────┐
  │  Compiler pod detects    │
  │  new commit / job        │
  │  (poll, webhook, Kafka)  │
  └────────────┬────────────┘
               │
        ▼ git pull
               │
        ▼ go build ./...
               │  fail → status: compile_failed, error_log written
               │
        ▼ go test ./platform/orchestration/actions/...
               │  fail → status: test_failed, error_log written
               │
        ▼ LLM review of test results (optional stage)
               │  fail → status: review_failed, feedback written
               │
        ▼ Build container image (kaniko/buildah)
               │
        ▼ Push to registry with new tag
               │
        ▼ Record image_tag, set status: built
               │
        ▼ Deploy (depending on HITL setting)
               │
        ▼ Rolling restart of agent-chassis pods
               │
        ▼ status: deployed, deployed_at set
```

---

## Job Tracking Table

The compiler pod uses this table as its job queue and audit trail. Source code lives in git, not here.

```sql
CREATE TABLE action_build_jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_name     TEXT NOT NULL,
    file_path       TEXT NOT NULL,
    test_path       TEXT,
    git_commit      TEXT,
    git_branch      TEXT DEFAULT 'main',
    status          TEXT NOT NULL DEFAULT 'pending',
    error_log       TEXT,
    image_tag       TEXT,
    previous_tag    TEXT,
    created_by      TEXT,
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    deployed_at     TIMESTAMPTZ,
    rolled_back_at  TIMESTAMPTZ
);
```

Status values: `pending` → `pulling` → `compiling` → `testing` → `reviewing` → `building_image` → `pushing` → `built` → `deployed` → `failed`

On failure at any stage, `status` is set to indicate which stage failed and `error_log` captures the output. The LLM agent that created the action can read this back and attempt a fix.

---

## Compiler Pod

### What it is

A Kubernetes deployment in the `ai-persona-system` namespace. Its base image contains the full agent-chassis source tree and Go toolchain. It shares the same `go.mod` and dependency tree as the running agent-chassis image.

### How it stays in sync

The compiler pod's image is rebuilt alongside agent-chassis on manual releases. They share a Dockerfile multi-stage build:

```dockerfile
# Shared base
FROM golang:1.22 AS go-base
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Agent-chassis production image
FROM go-base AS chassis-build
RUN go build -o /agent-chassis ./cmd/
FROM gcr.io/distroless/base AS chassis
COPY --from=chassis-build /agent-chassis /agent-chassis
CMD ["/agent-chassis"]

# Compiler pod image — same base, same deps, plus build tools
FROM go-base AS compiler
RUN go install github.com/GoogleContainerTools/kaniko/cmd/executor@latest
COPY compiler-service/ ./compiler-service/
RUN go build -o /compiler-service ./compiler-service/
CMD ["/compiler-service"]
```

This guarantees ABI compatibility. The compiler always builds against the same Go version and module graph as the running chassis.

When an LLM-written action is successfully built and deployed, the compiler pod's source tree now includes that action. The next build includes all previous LLM-written actions automatically — the git repo is the accumulator.

### How it triggers

Three options (not mutually exclusive):

- **Poll**: check git for new commits every N seconds
- **Kafka**: LLM agent publishes to a `system.build.requests` topic after pushing to git
- **Webhook**: git repo sends a push webhook to the compiler pod's HTTP endpoint

Kafka fits your existing infrastructure best. The LLM agent pushes to git, then publishes a build request message containing the commit SHA and file paths.

### Image building

The compiler pod uses kaniko for in-cluster image building. Kaniko doesn't need Docker or a privileged container — it builds images in userspace. The registry credentials come from a Kubernetes secret (already configured for your existing image pushes).

```bash
/kaniko/executor \
    --context=/src \
    --dockerfile=/src/Dockerfile \
    --target=chassis \
    --destination=docker.io/aqls/agent-chassis:$NEW_TAG
```

---

## LLM Test Review Stage

Between `testing` and `building_image`, an LLM reviews the results. This isn't the same LLM that wrote the action — it's a reviewer role with a different prompt.

The reviewer receives:

- The action source code
- The test source code
- The test output (pass/fail, coverage, any warnings)
- The `error_log` if tests failed
- The InputSpec and any relevant schema context

Its job:

- If tests passed: sanity check that the tests are meaningful (not just `assert true`). Check that edge cases are covered. Check that the action follows project conventions (uses `datahelpers.ExtractActionInputs`, has proper error handling, doesn't break variable naming contracts). Approve or reject with feedback.
- If tests failed: diagnose the failure, write a fix suggestion. The originating LLM agent can pick this up and iterate.

The reviewer writes its verdict back to `action_build_jobs.status` and `error_log`. On approval, the pipeline continues. On rejection, the originating agent gets feedback via Kafka and can push a corrected version.

This creates a loop:

```
LLM writes action → compile → test → LLM reviews
     ↑                                    │
     │         feedback if rejected        │
     └─────────────────────────────────────┘
```

The loop has a max iteration count (3-5 attempts) to prevent infinite cycles.

---

## Deployment and Rollback

### HITL dial

The HITL level is a configuration value on the compiler pod:

| Level | Behaviour |
|-------|-----------|
| `manual` | Pipeline stops at `built`. Human runs deploy command. |
| `staging` | Auto-deploys to staging namespace. Human promotes to production. |
| `auto` | Auto-deploys to production after LLM review passes. Human notified. |

Start with `manual`. Move to `staging` once confidence grows. `auto` is the end state.

### Rolling restart

```bash
kubectl -n ai-persona-system set image deployment/agent-chassis \
    agent-chassis=docker.io/aqls/agent-chassis:$NEW_TAG
```

Kubernetes handles the rolling update — new pods come up, old pods drain. Zero downtime.

### Rollback

Every `action_build_jobs` row records `previous_tag` — the image tag that was running before this deployment. Rollback:

```bash
kubectl -n ai-persona-system set image deployment/agent-chassis \
    agent-chassis=docker.io/aqls/agent-chassis:$PREVIOUS_TAG
```

Set `rolled_back_at = NOW()` on the job row. The rolled-back action's source stays in git (it's not deleted), but the running image no longer includes it. The originating LLM agent can be notified to investigate.

For automated rollback: monitor agent-chassis pod restart counts and error rates after deployment. If restarts spike within 5 minutes, auto-rollback and mark the job as failed.

---

## LLM Agent Interface

The LLM agent that writes actions needs:

- **Git push credentials** — a deploy key (SSH) or personal access token scoped to the repo, stored as a Kubernetes secret and mounted into the agent's pod.
- **Knowledge of the action template** — the development guide (001d) plus exemplar actions in its prompt context.
- **Feedback channel** — reads `action_build_jobs` status and `error_log` to iterate on failures.

The agent's workflow for writing a new action:

1. Receive a request (from an orchestrator, or from a discovery check that identifies missing capability)
2. Search existing actions in the registry to avoid duplication (Step Zero from the dev guide)
3. Write the action `.go` file and `_test.go` file
4. `git push` to the repo
5. Publish build request to Kafka
6. Poll `action_build_jobs` for status
7. If failed: read `error_log`, fix, push again (up to max iterations)
8. If deployed: register the action in the agent definition workflow that needs it

Step 8 is the interesting bit — the LLM also writes/updates the workflow JSON in `agent_definitions` to reference the new action. Since workflows are database-native, this is immediate. The full cycle: need identified → action written → compiled → tested → deployed → wired into a workflow → running.

---

## What This Doesn't Cover (Yet)

- **Action removal** — deleting an action requires ensuring no workflow references it. The compiler pod could check for references before excluding a file from the build.
- **Dependency management** — if an LLM-written action needs a new Go module (e.g. a parsing library), `go.mod` needs updating. The compiler pod can handle `go mod tidy` but adding entirely new dependencies needs care.
- **Multi-action builds** — if several LLM agents push actions simultaneously, the compiler pod needs to serialise builds or batch them. A simple queue (the `pending` status) handles this.
- **Adapter changes** — the agent-chassis isn't the only deployable. Adapter code changes would need a similar pipeline. Same pattern, different Dockerfile target.

---

## Priority

This is a medium-term investment. The immediate value is the modular discovery check pattern (the `init()` registry approach) which reduces the friction of adding checks manually. The automated pipeline builds on that foundation — once checks are self-registering files, the step to "LLM writes the file and a machine compiles it" is smaller.

The order would be:

1. Migrate discovery checks to the modular pattern (current work)
2. Build the compiler pod as a basic git-pull → build → push service
3. Add the `action_build_jobs` table and status tracking
4. Add the LLM review stage
5. Connect an LLM agent with git push credentials
6. Adjust the HITL dial as confidence grows