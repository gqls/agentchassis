# NOTES — direct-caller LLM observability

*Technical log. Append-only, newest at the bottom. Missteps included — they are the point.*

---

## 2026-09-04 — the census this lane was spun out on

Run by the `bugfix_257_token_budget_at_the_client` lane, not by this one. Recorded here so the first
session has a starting point AND knows exactly how it was produced, so it can re-run it rather than
quote it.

**Who writes the instrument:**
```bash
grep -rln "llm_call_log" --include=*.go platform/ internal/ pkg/ | grep -v resulttext
```
→ every hit is under `platform/orchestration/actions/`. One package writes the table.

**Who calls a model without going through it:**
```bash
grep -rn "GenerateText(\|GenerateWithImages(" --include=*.go internal/ platform/ pkg/ cmd/ \
  | grep -v resulttext | grep -v _test.go \
  | grep -vE "platform/orchestration/actions/|platform/aiservice/"
```
→ `internal/tools-api/handlers/defend.go:101`, `position.go:90`, `gripper.go:161`,
`internal/agents/contentcreator/agent.go:305`, `internal/agents/reasoning/agent.go:127`.

⚠ **That second grep is the one that has been wrong four times before.** It is keyed on the CLIENT
INTERFACE, so it cannot see a provider called over raw HTTP — which is exactly how
`feed_actions.go`'s hardcoded `"max_tokens": 4096` survived four censuses of bug 257 across three
weeks. Census the CONCEPT as well:
```bash
grep -rnE '"(max_tokens|max_output_tokens|maxOutputTokens|num_predict|max_completion_tokens)"' \
  --include=*.go . | grep -v _test.go
```
which returned **51 hits as of 2026-09-03** against 12 from the interface grep. Full entry in
`LANDMINES.md`.

**One thing that is already better than it looks:** `internal/agents/reasoning/agent.go:127` passes
`nil` options. Since bug 257 Path A that is not a defect — a nil options map now inherits
`ai_service.max_tokens` from the config the client was constructed with. It is unlogged, which is
this lane's problem, but its budget is correct.

**And one that is worth reading before designing anything:**
`internal/tools-api/handlers/ailog.go` already reports truncation distinctly from other failures,
with the reasoning written out ("it is NOT an upstream fault but our own configured cap, and it needs
a different fix"). It writes to stdout. So the tools-api half of this lane is not "nobody thought
about it" — it is "the thought went somewhere no fleet query can read".
