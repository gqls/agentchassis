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

## chassis-drafts/analyser-adapter (staging → agent-chassis destinations)

```
chassis-drafts/analyser-adapter/
├── adapter.go                      → internal/adapters/analyser/adapter.go
├── analyse_action.go               → internal/adapters/analyser/analyse_action.go
├── github_source.go                → internal/adapters/analyser/github_source.go
├── cmd/analyser-adapter/main.go    → cmd/analyser-adapter/main.go
├── configs/analyser-adapter.yaml   → configs/analyser-adapter.yaml
├── analyser-adapter.yaml           → SPLIT into the kustomize base (below);
│                                     the Secret block is applied via your
│                                     secret manager, never committed with a
│                                     real token
├── code_symbols_actions.go         → alongside rag_actions.go (package actions)
│                                     + 3 registry entries: index_code_symbols,
│                                     lookup_code_symbols, request_repo_analysis
├── analyser_request_action.go      → alongside webscrape_actions.go (package actions)
└── NNN_create_code_indexer_agent.sql → migrations (your numbering; after
                                      \d agent_definitions confirm)
```

Also in the workspace root:

```
NNN_create_code_symbols_index.sql   → migrations (ALREADY APPLIED: code_symbols)
035_adapter_guide.md                → docs (canonical adapter guide; FOCUS_adapter_design retired)
```

## Still to create (per the Makefile/kustomize conventions)

```
build/docker/backend/analyser-adapter.dockerfile      canonical two-stage pattern
deployments/kustomize/services/analyser-adapter/
├── base/
│   ├── deployment.yaml             from the staged manifest (Deployment only)
│   ├── service.yaml                ClusterIP, port 8080 (health)
│   ├── kustomization.yaml          + configMapGenerator from configs/analyser-adapter.yaml
│   └── analyser-adapter.yaml       the service config (generator source)
└── overlays/production/uk_001/
    └── kustomization.yaml          images: name analyser-adapter,
                                    newName docker.io/aqls/analyser-adapter,
                                    newTag vX.Y.Z   (deploy-agents seds newTag)
```

Makefile — four insertions, same shape as thunder's:
1. `build-analyser-adapter` target (`-f build/docker/backend/analyser-adapter.dockerfile`)
2. add to the `build-adapters` aggregate list
3. `docker push $(REGISTRY)/analyser-adapter:$(IMAGE_TAG)` in `push-backend`
4. guarded sed+apply block in `deploy-agents` for
   `$(KUSTOMIZE_DIR)/services/analyser-adapter/overlays/$(OVERLAY_PATH)`

Cluster prerequisites before first deploy: the `analyser-github-read` Secret
(fine-grained, repo-scoped, read-only) and the KafkaTopic CRD for
`system.adapter.analyser.requests` (auto-create is off — without it the
consumer hangs on every fetch).
