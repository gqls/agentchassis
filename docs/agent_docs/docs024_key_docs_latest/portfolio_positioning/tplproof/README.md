# tplproof — render proof for migration 764 (bugs_open/453), through the FLEET'S OWN renderer

Council round 1 (`888e7319`, editquality) objected that a harness with its own `text/template` +
funcmap proves the template text, not the production injection path. This version imports the real
package and calls **`datahelpers.RenderPromptTemplate`** — what `ai_actions.go:328` calls for
`execute_llm_prompt`, the action that `classify_and_extract`, `plan_site` AND `domain-strategist`'s
`analyze_strategy` all run under — and **`datahelpers.ScanMissingValues`**, PRC-003's own attribution
scan. So the hole and its disappearance are measured by the fleet's instrument.

```sh
cd docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/tplproof
for pair in "build-site-planner:plan_site" "domain-research-classifier:classify_and_extract"; do t=${pair%%:*}; s=${pair##*:}
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A \
    -c "SELECT default_config #>> '{workflow,steps,$s,config,prompt_template}' FROM agent_definitions WHERE type='$t' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > "$t.tpl"; done
go test -tags tplproof . -run TestBrief764 -v
```

`go.mod` here is a per-lane module with `replace github.com/gqls/agentchassis => ../../../../..`,
because `docs/go.mod` deliberately carves the docs tree out of the main module. The build tag keeps
the test out of every `./...`. `*.tpl` are live-data pulls and are git-ignored.

**What it asserts, per template, three contexts (brief with `text` / brief object without / none):**
`ScanMissingValues` occurrences delta **0 / −1 / 0** between the unmodified and the edited template;
in the object case the ORIGINAL report's `Fields` names `site_specs.specs.mission_brief.text` and the
original render does NOT show the object (control: the defect reproduces); the sentinel appears in the
fixed render; the prose case is byte-identical orig vs fixed; no Mission block with no brief.

**Result 2026-09-03 ~20:5xZ: PASS.** Object case: planner scan 3→2, classifier 4→3, both attributing
`site_specs.specs.mission_brief.text` in the original and not in the fixed. The other attributed
paths (`available_components`, `search_results`, `layout_taxonomy.*`…) are harness-context artefacts
present in BOTH renders — which is why every assertion is a delta.

**Two wrong turns, kept so nobody repeats them.** (1) The first harness asserted absolute zero
`<no value>` inside the Mission block; the classifier's block also carries research inputs a thin
context leaves empty. (2) This version's first run scanned `RenderPromptTemplate`'s OUTPUT — which
PRC-003 has already STRIPPED — and read zero holes everywhere. The scan must run on the raw execution
parsed with `datahelpers.PromptTemplateFuncs()`, exactly as `data_helpers.go:1204` does before it scans.

This proves the render path, not the fleet: after applying 764, make it run once (migration header).
