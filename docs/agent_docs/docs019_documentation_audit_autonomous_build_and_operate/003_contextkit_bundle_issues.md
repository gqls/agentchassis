# contextkit `cmd/bundle` — issues found in use

**From:** the agent-chassis debugging work. We use `go run ./cmd/bundle` to assemble a single
context bundle (`.md`) — constitution + runtime + schema + scoped source + docs — to hand to a fresh
debugging chat.
**Date:** 2026-06-24
**Status of the tool:** it has produced good bundles before (e.g. `bundle_gamesdesign.md`), so these
are input/robustness/UX problems, not "the tool is broken." This session we attempted two real runs and
**neither produced output** — details below, worst-first.

---

## Summary

| # | What happened | Root cause | Tool defect? | Fix in one line |
|---|---|---|---|---|
| A | Tool gathered schema/capabilities/runtime, then died: `load analysis: unexpected end of JSON input` | `-analysis` file was empty/truncated, and it is loaded **after** the slow gather phases | **Yes** — late validation + cryptic, stacked error | Validate `-analysis` first; name the file + size + how to regenerate |
| B | `bash: syntax error near unexpected token '('` ×3, then `Command '-psql' not found`; no output | Unquoted parentheses in `-doc` filenames broke shell parsing — `go run` never executed | **No** — shell quoting on the invocation | Quote paths; offer a manifest-file input so a bundle isn't a 20-line shell command |
| C | (not reached, but latent) a wrong/typo'd `-doc`/`-scope`/`-include` path | unknown — does the tool fail loudly or silently skip the file? | **TBD** | Fail loudly, naming the missing path |

---

## Issue A — `-analysis` is validated too late, and the error is unhelpful  *(real bug)*

**Observed** (the "clobber" bundle run):

```
gathered schema -> /tmp/bundle-725606384/schema.md
gathered capabilities -> /tmp/bundle-725606384/dbfacts.md
gathered runtime -> /tmp/bundle-725606384/runtime.md
load analysis: unexpected end of JSON input
exit status 1
assembler failed: exit status 1
exit status 1
```

**What went wrong**

- The file passed via `-analysis` (`/tmp/analysis_repo.json`) was empty (0 bytes) / truncated, so the
  JSON decode returned `unexpected end of JSON input`.
- That file is the *cheapest* input to validate (a local read + unmarshal), yet it is loaded **after** the
  schema, capabilities, and runtime gather phases — which shell out to `kubectl exec … psql` and are the
  slow, network-bound, side-effecting part of the run. A trivial precondition failure therefore wastes all
  the expensive work, on every attempt.
- The message doesn't say *which* file, whether it was empty vs malformed, or how to regenerate it. The user
  has to already know `-analysis` → `/tmp/analysis_repo.json`, and that the file is produced by a separate
  analysis step (`analyse.go` / `callgraph.go`).
- The failure then prints as three stacked lines (`load analysis…` → `assembler failed: exit status 1` →
  `exit status 1`), which buries the one line that matters. (The "assembler" appears to be a subprocess, so
  its non-zero exit is being re-wrapped on the way out.)

**Suggested fixes**

1. **Check `-analysis` before any gather.** `os.Stat` it (exists, non-zero), read + `Unmarshal` it, and
   fail fast if bad. Generally: order preconditions cheapest-first, so DB-hitting work only starts once the
   local inputs are known good.
2. **Make the message actionable**, e.g.
   `` -analysis "/tmp/analysis_repo.json" is empty (0 bytes); regenerate it with `<analyse command>` and re-run. ``
   Wrap the json error with the file path and the `os.Stat` size, and print the regeneration command.
3. **Collapse the nested exit-status wrapping** into a single clear failure line (don't echo `exit status 1`
   from each subprocess layer).
4. *Optional:* **freshness.** If the analysis JSON is older than the repo's tracked source (or `HEAD`), warn —
   or support `-refresh-analysis` / regenerate automatically. The implicit rule "you must regenerate the
   analysis after changing code" is a recurring footgun; the tool could own it.

---

## Issue B — unquoted parentheses in `-doc` paths break the shell  *(invocation pitfall, not a tool defect)*

**Observed** (the "phantom-CTA" bundle run):

```
bash: syntax error near unexpected token `('
bash: syntax error near unexpected token `('
bash: syntax error near unexpected token `('
Command '-psql' not found, did you mean:
  command 'psql' from deb postgresql-client-common (282ubuntu1)
cp: cannot stat '/tmp/bundle_phantom_cta.md': No such file or directory
```

The three `-doc` paths contained parentheses (browser-style duplicate-download names):

```
-doc docs/.../016_debugging_guide_v2_56(1).md
-doc docs/.../005_tool_pipeline(1).md
-doc docs/.../003_contracts_and_standards(7).md
```

Bash treats `(` and `)` as metacharacters, so each of those lines is a syntax error. The broken backslash
line-continuation then left the next physical line (`-psql '…'`) to be parsed as its own command, hence
`Command '-psql' not found`. `go run` never executed, so there was no output and the trailing `cp` also
failed.

**This is a shell-quoting problem (quoting the paths fixes it), not a bug in the Go tool** — the tool
received no arguments. But it recurs, because (a) real filenames here contain `()` and sometimes spaces, and
(b) a bundle invocation is a ~20-line, backslash-continued shell command, which is inherently fragile.
Tool-side mitigations, strongest first:

1. **Accept a manifest / config file** (JSON/YAML/TOML) listing `scopes`, `includes`, `docs`,
   `schema_tables`, `runtime`, `psql`, etc., invoked as `-config bundle.yaml`. This removes the entire class
   of shell-quoting hazards, and makes a bundle reproducible and reviewable (you can commit it). This is the
   highest-leverage change.
2. **Quote every path in the README / example invocations**, with a one-line note: *"paths may contain `()`
   or spaces — quote them."* Most users copy the examples.
3. *Optional:* a `--print-config` / dry-run that echoes the parsed flags and resolved file list and exits,
   so a malformed invocation is caught instantly (and doubles as documentation of what got included).

---

## Issue C — behaviour on a missing / typo'd `-doc` / `-include` / `-scope` path  *(please confirm)*

We didn't reach this directly, but it's important for how the bundle is used: the bundle is frequently the
*entire* context handed to a fresh debugging chat. If a target path is wrong, the dangerous outcome is the
tool **silently omitting** that file — the downstream chat then reasons from incomplete context, and nobody
notices.

Concretely, our two invocations disagreed on doc filenames (`016_debugging_guide_v2_56.md` vs
`…_v2_56(1).md`; `003_contracts_and_standards_7_.md` vs `…(7).md`), so a wrong name is a realistic event.

**Request:** confirm that a missing/unreadable `-doc` / `-include` / `-scope` target **fails loudly, naming
the path**, rather than being skipped. If it currently skips, please make it fail — or at minimum print a
prominent `OMITTED: <path>` summary at the end of a run so the gap is visible.

---

## Lower-priority notes

- **Temp intermediates** (`/tmp/bundle-<n>/schema.md`, `dbfacts.md`, `runtime.md`) are left behind on
  failure. The path is printed, which is helpful; cleanup-on-success (or a `-keep-temp` flag) would tidy up.
- **Verbose paths.** The long prefix `platform/orchestration/actions/…` repeats on every `-scope` /
  `-include`. A base-dir flag or glob would shorten invocations and pairs naturally with the manifest idea.

---

## Appendix — the two failing invocations (for reproduction)

Run 1 — "phantom-CTA" (failed at the shell, before `go run`):

```bash
go run ./cmd/bundle \
  -analysis /tmp/analysis_repo.json -root ~/projects/agentchassis \
  -constitution thin_slice_constitution.md -step debug \
  -task "<one-line task, omitted>" \
  -scope platform/orchestration/actions/resolve_internal_links_action.go:ResolveInternalLinksAction \
  -scope platform/orchestration/actions/extract_fields_action.go:ExtractFieldsAction \
  -scope platform/orchestration/actions/call_agent_action.go:CallAgentAction \
  -scope platform/orchestration/actions/complete_workflow_action.go:CompleteWorkflowAction \
  -include platform/orchestration/actions/compile_page_sections_action.go \
  -include platform/orchestration/actions/render_component_action.go \
  -include platform/orchestration/actions/registry.go \
  -doc docs/agent_docs/docs024_key_docs_latest/016_debugging_guide_v2_56(1).md \
  -doc docs/agent_docs/docs024_key_docs_latest/005_tool_pipeline(1).md \
  -doc docs/agent_docs/docs024_key_docs_latest/003_contracts_and_standards(7).md \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema-tables orchestration_states,agent_definitions,pages,page_components,page_component_history,site_work_items \
  -runtime-site gamesdesign.co.uk -runtime-page guide-economy-basics \
  -capabilities -df-filter snapshot \
  -out /tmp/bundle_phantom_cta.md
```
→ unquoted `()` in the three `-doc` paths ⇒ `syntax error near unexpected token '('` ⇒ `go run` never ran.
**Fix:** single-quote the parenthesised paths (or use a `-config` file).

Run 2 — "clobber" (parsed fine; tool ran and failed on `-analysis`):

```bash
go run ./cmd/bundle \
  -analysis /tmp/analysis_repo.json -root ~/projects/agentchassis \
  -constitution thin_slice_constitution.md -step debug \
  -task "<one-line task, omitted>" \
  -scope platform/orchestration/actions/save_page_sections_action.go:SavePageSectionsAction \
  -scope platform/orchestration/actions/plan_sections_action.go \
  -scope platform/orchestration/actions/load_page_sections_from_spec_action.go \
  -include platform/orchestration/actions/registry.go \
  -doc docs/agent_docs/docs024_key_docs_latest/016_debugging_guide_v2_56.md \
  -doc docs/agent_docs/docs024_key_docs_latest/026_component_regeneration_flow_1_.md \
  -doc docs/agent_docs/docs024_key_docs_latest/003_contracts_and_standards_7_.md \
  -doc docs/agent_docs/docs024_key_docs_latest/020_tool_lifecycle_2_.md \
  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema-tables page_components,pages,page_component_history,site_work_items \
  -runtime-site gamesdesign.co.uk -runtime-page game-pathfinding \
  -capabilities -df-filter snapshot \
  -out /tmp/bundle_clobber.md
```
→ `gathered schema/capabilities/runtime`, then `load analysis: unexpected end of JSON input`.
**Fix (user side):** regenerate `/tmp/analysis_repo.json` (the analyse step) and confirm `wc -c` > 0 before
re-running. **Fix (tool side):** Issue A above.
