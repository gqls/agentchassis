# PATCH — lift the GitHub fetcher into `internal/reposource`, register `analyse_repo_local`

Three small edits in the agent-chassis tree, alongside the two NEW files
(`internal/reposource/github_source.go`, `platform/orchestration/actions/analyse_repo_local_action.go`).
Apply in your env and build there (none of this compiles in the contextkit container).

---

## 1. Move the fetcher — delete the adapter-local copy

The fetcher is now `internal/reposource/github_source.go` (lifted verbatim; only
the package name + FILE path changed, plus a `Fetcher` interface added). Delete
the old copy so there is ONE implementation:

```bash
git rm internal/adapters/analyser/github_source.go
```

Nothing else in the analyser package defines `GitHubSource` / `NewGitHubSource` /
`extractTarGz`, so this is a clean removal once edit #2 lands.

`analyse_action.go` needs **no change**: it depends on its own local `repoSource`
interface, which `*reposource.GitHubSource` still satisfies structurally.

---

## 2. `internal/adapters/analyser/adapter.go` — import the lifted package + repoint the constructor

**Add the import** (with the other `github.com/gqls/agentchassis/...` imports):

```go
	"github.com/gqls/agentchassis/internal/reposource"
```

**Change the one construction site** in `NewAdapter` (the only reference):

```go
	// BEFORE:
	source, err := NewGitHubSource(os.Getenv("GITHUB_READ_TOKEN"), os.Getenv("GITHUB_API_BASE"), logger)

	// AFTER:
	source, err := reposource.NewGitHubSource(os.Getenv("GITHUB_READ_TOKEN"), os.Getenv("GITHUB_API_BASE"), logger)
```

Signature, args, env vars, and behaviour are identical — only the package
qualifier changes. `NewAnalyseAction(source, logger)` is unchanged (`source` still
satisfies the local `repoSource` interface).

---

## 3. `platform/orchestration/actions/registry.go` — register the new action

In the **DOCUMENT AND CODE ANALYSIS** block, add `analyse_repo_local` right after
the `request_repo_analysis` entry:

```go
	"request_repo_analysis": {
		Handler:     RequestRepoAnalysisAction,
		Category:    "code",
		Description: "Ask the analyser adapter to parse a repo at ref; awaits the symbol output",
		IsLocal:     true,
	},
	"analyse_repo_local": {
		Handler:     AnalyseRepoLocalAction,
		Category:    "code",
		Description: "Fetch a repo at ref to a local temp dir (read-only tarball) and analyse it in-process; returns the analyser Output with a real local root + commit_sha (for the diagnose loop's body reads). Read-only.",
		IsLocal:     true,
	},
```

Optionally update the DIAGNOSE block's flow comment so it documents the swap:

```go
	//   gather: analyse_repo_local → lookup_code_symbols → diagnose_load_runtime
	//           → diagnose_assemble_bundle
```

(`request_repo_analysis` stays registered — the separate **code-indexer** agent
still uses it.)

---

## Build / deploy order

1. Apply edits #1–#3 + drop in the two new files.
2. `go build ./...` (the lift + the new action compile against existing deps:
   `internal/analysis`, `datahelpers`, `zap`, `net/http`).
3. Push → GitHub Actions → Backblaze; confirm the new image tag.
4. THEN run `NNN_swap_analyse_repo_to_local.sql` (it needs the new action present).
5. Re-trigger the §6F diagnosis run; confirm by correlation_id (not LIMIT 1).

The diagnose-agent pod already gets `GITHUB_READ_TOKEN` via `spawn_actions.go`
(`isRepoCloningAgent`), so no secret/RBAC change is needed.
