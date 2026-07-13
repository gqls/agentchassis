# PATCH — shared owner/repo label resolver (index + lookup symmetry)

FILE: `platform/orchestration/actions/code_symbols_actions.go` (chassis; builds in your env).

## Why

`index_code_symbols` and `lookup_code_symbols` must key on the SAME `code_symbols.repo`
label or the lookup finds nothing. Today they diverge:

- `index_code_symbols` (lines ~145–157) resolves `repo_field`/`config.repo`, and **when
  empty composes** `repo_analysis.owner + "/" + repo_analysis.repo` → `gqls/agentchassis`.
- `lookup_code_symbols` (line ~59) is `repo := resolveRAGConfigField(config, "repo_field",
  "repo", ...)` with **no composition**. The diagnose workflow set `lookup_symbols.repo_field
  = "repo_analysis.repo"`, the BARE name, so the lookup queried `WHERE repo='agentchassis'`
  against rows stored under `gqls/agentchassis` → 0 hits → empty `code_results` →
  `assemble_bundle: no scope`.

Fix the asymmetry at the source: one resolver, used by both, so they can never drift.

## 1) Add the shared resolver (one definition)

```go
// resolveCodeRepoLabel resolves the owner/repo label that BOTH index_code_symbols and
// lookup_code_symbols key on, so the writer and reader can never diverge:
//   1. config.repo / config.repo_field — explicit override (non-git corpora);
//   2. COMPOSE owner/repo from the analyser reply (the default — matches what was
//      fetched and stored: repo_analysis.owner + "/" + repo_analysis.repo);
//   3. input_data.repo — last-resort fallback.
func resolveCodeRepoLabel(config map[string]interface{}, collected map[string]interface{}) string {
	repo := resolveRAGConfigField(config, "repo_field", "repo", collected)
	if repo == "" {
		ownerPath := datahelpers.GetStringField(config, "owner_field", "repo_analysis.owner")
		namePath := datahelpers.GetStringField(config, "repo_name_field", "repo_analysis.repo")
		owner := datahelpers.ExtractNestedFieldString(collected, ownerPath)
		name := datahelpers.ExtractNestedFieldString(collected, namePath)
		if owner != "" && name != "" {
			repo = owner + "/" + name
		}
	}
	if repo == "" {
		repo = datahelpers.ExtractNestedFieldString(collected, "input_data.repo")
	}
	return repo
}
```

## 2) Use it in BOTH actions

`LookupCodeSymbolsAction` — replace line ~59:
```go
	repo := resolveCodeRepoLabel(config, params.CollectedData)
```

`IndexCodeSymbolsAction` — replace the inline block at lines ~145–157 (everything from
`repo := resolveRAGConfigField(...)` through the `input_data.repo` fallback, but KEEP the
`if repo == "" { return ...error... }` guard at ~158) with:
```go
	repo := resolveCodeRepoLabel(config, params.CollectedData)
```
This is behaviour-preserving for index (same resolution order) and gives lookup the same
composition. No variable names change beyond removing the now-inlined locals.

## 3) Workflow config — let the lookup compose (diagnose-agent)

With the resolver, `lookup_symbols` must NOT pin `repo_field` to the bare name, or step 1
returns it and composition never runs. End state for the step config:
`{ "top_k": 12, "query_field": "input_data.symptom" }` — NO `repo_field`, NO `repo`.
The `owner_field`/`repo_name_field` defaults (`repo_analysis.owner` / `repo_analysis.repo`)
then compose `gqls/agentchassis` for ANY repo. That is exactly the REVERT block in
`NNN_fix_lookup_repo_label_workaround.sql` — run it once this image is deployed.

## Order

1. Apply this patch, build + deploy (GitHub Actions → Backblaze).
2. Run the workaround's REVERT block (drops the hard-coded literal + bare `repo_field`).
3. Re-trigger §6F — the lookup now seeds iteration 1 from the symptom for any repo.

Until the image ships, the workaround migration (literal `config.repo`) keeps the eval
unblocked.
