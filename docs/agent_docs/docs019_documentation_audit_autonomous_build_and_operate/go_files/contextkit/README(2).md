# Directory map

Two trees share this workspace: the **contextkit** module (compiles standalone)
and the **chassis-drafts** staging area (files destined for the agent-chassis
repo; they do not compile here). Destinations below follow the conventions
confirmed from the repo Makefile (2026-06-11): per-service images
`$(REGISTRY)/<name>:$(IMAGE_TAG)` with `REGISTRY=docker.io/aqls`, dockerfiles at
`build/docker/backend/`, kustomize at
`deployments/kustomize/services/<name>/{base,overlays/$(OVERLAY_PATH)}` where
`OVERLAY_PATH=production/uk_001` for production.

## contextkit (module `contextkit`)

```
contextkit/
├── go.mod                          module contextkit
├── internal/
│   ├── analysis/
│   │   ├── types.go                analyser output contract — defined ONCE
│   │   └── analyse.go              the layer-1 AST walk: Analyse(root) (Output, error)
│   │                               (extracted from cmd/analyser; SHARED with the
│   │                               chassis — see destination note below)
│   └── candidates/
│       └── types.go                ranked-candidate contract — defined ONCE
└── cmd/
    ├── analyser/main.go            thin wrapper over internal/analysis → analysis JSON
    ├── assembler/main.go           builds the paste-ready bundle
    ├── embed/main.go               build | query the semantic index
    ├── dbcontext/main.go           schema / rows / runtime evidence via psql
    ├── resolve_targets/main.go     lexical target proposal (-json)
    ├── fuse/main.go                RRF merge of candidate lists
    └── eval_targets/main.go        recall@N / MRR vs ground truth
```

```
tar -xzf contextkit.tar.gz
cd contextkit
go build ./...        # compiles all seven commands
go run ./cmd/analyser /path/to/your/repo > analysis.json
```

**Shared-library note:** `internal/analysis/` is copied (not moved) into the
agent-chassis repo as `internal/analysis/` — a chassis-root internal package,
NOT under `internal/adapters/analyser/`. The adapter imports
`github.com/gqls/agentchassis/internal/analysis`; keeping it at the root lets
future consumers (the JS parser, indexer tests) import it without importing the
adapter. contextkit keeps its own copy for the standalone CLI.

## Where everything lands in the agentchassis tree

Anchored to the real repo tree (tree -d, 2026-06-11). Markers: (existing) =
directory already in the tree; (NEW) = directory to create; (EDIT) = existing
file to modify.

```
agentchassis/
├── cmd/
│   └── analyser-adapter/                          (NEW — sibling of thunder-adapter/)
│       └── main.go                                ← chassis-drafts/…/cmd/analyser-adapter/main.go
│
├── internal/
│   ├── adapters/
│   │   └── analyser/                              (NEW — sibling of git/, thunder/, webscrape/)
│   │       ├── adapter.go                         ← chassis-drafts/…/adapter.go
│   │       ├── analyse_action.go                  ← chassis-drafts/…/analyse_action.go
│   │       └── github_source.go                   ← chassis-drafts/…/github_source.go
│   │
│   └── analysis/                                  (existing — ALREADY in your tree)
│       ├── analyse.go                             ← contextkit/internal/analysis/analyse.go
│       └── types.go                               ← contextkit/internal/analysis/types.go
│                                                  verify the repo copies match the
│                                                  contextkit versions (same files)
│
├── platform/
│   ├── orchestration/
│   │   └── actions/                               (existing — registry.go, rag_actions.go,
│   │       │                                       webscrape_actions.go live here)
│   │       ├── code_symbols_actions.go            ← chassis-drafts/…/code_symbols_actions.go
│   │       ├── analyser_request_action.go         ← chassis-drafts/…/analyser_request_action.go
│   │       └── registry.go                        (EDIT — paste the block from
│   │                                               registry_insertions.md after
│   │                                               training_data_export)
│   └── database/
│       └── migrations/                            (existing)
│           ├── 0NN_create_code_symbols_index.sql  ← workspace root (ALREADY APPLIED —
│           │                                       commit for the record, your numbering)
│           └── 0NN_create_code_indexer_agent.sql  ← chassis-drafts/… (run after
│                                                   \d agent_definitions confirm)
│
├── configs/                                       (existing, flat)
│   └── analyser-adapter.yaml                      ← chassis-drafts/…/configs/analyser-adapter.yaml
│
├── build/docker/backend/                          (existing)
│   └── analyser-adapter.dockerfile                (NEW file — copy the two-stage
│                                                   pattern from thunder-adapter.dockerfile)
│
├── deployments/kustomize/services/
│   └── analyser-adapter/                          (NEW — sibling of thunder-adapter/)
│       ├── base/
│       │   ├── deployment.yaml                    ← Deployment block of
│       │   │                                       chassis-drafts/…/analyser-adapter.yaml
│       │   ├── service.yaml                       (NEW — ClusterIP, port 8080)
│       │   ├── kustomization.yaml                 (NEW — + configMapGenerator
│       │   │                                       analyser-adapter-config)
│       │   └── analyser-adapter.yaml              (the service config — generator
│       │                                           source; mirror where thunder's
│       │                                           base keeps its config)
│       └── overlays/production/uk_001/            (matches OVERLAY_PATH)
│           └── kustomization.yaml                 (NEW — images: newName
│                                                   docker.io/aqls/analyser-adapter,
│                                                   newTag; deploy-agents seds newTag)
│
├── Makefile                                       (EDIT — 4 insertions: build target,
│                                                   build-adapters list, push-backend
│                                                   line, guarded deploy-agents block)
│
└── docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/
    └── go_files/contextkit/                       (existing — contextkit's in-repo home;
                                                    sync with contextkit.tar.gz incl. this
                                                    README + internal/analysis/analyse.go)
```

Not placeable from a directory-only tree — your call:

- `035_adapter_guide.md` → wherever `001_development_guide` / `003_contracts_and_standards` live.
- The KafkaTopic CRD for `system.adapter.analyser.requests` → wherever
  `system.adapter.thunder.requests` is declared: candidates are
  `deployments/kustomize/jobs/kafka-topics/` or
  `deployments/terraform/environments/production/uk001/080-kafka-topics/`.
- The `analyser-github-read` Secret → applied via your secret manager, like
  `docker-hub-creds`; never committed with a real token.
