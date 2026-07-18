# HANDOFF — FIX: fix-implementer commits un-`gofmt`'d LLM output, so the build gate rejects trivially-unformatted code and burns a whole implementer run

**Filed:** 2026-07-18, from the concept-register / council thread. Diagnosed directly
(visible bug — grep + read, per CLAUDE.md; NOT routed to the diagnosis loop). This
handoff is the FIX only. Small, contained, structural.

**Severity:** Medium. Not a correctness bug — nothing wrong ships. It is a **loop-yield
bug**: a fix-implementer run whose generated code is logically correct but cosmetically
un-`gofmt`'d produces NO PR — the run is spent (LLM whole-file generation + git push +
a k8s clone/build Job) and a human must hand-finish. LLMs routinely misalign adjacent
struct fields after inserting a new one, so this is not a rare tail; it hit BUG A's very
first implementer run (below).

## Working rules (hold these)
Go, not Python. British English. Read the function before changing it; structural fixes
over patches; reuse existing machinery. Go changes are inert until a chassis image ships
(`make build-agent-chassis` from committed HEAD; bump IMAGE_TAG; verify against the
RUNNING pod, never git). Commit per task with an explicit pathspec on `git commit`.
DB: `PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"`.

## 1. The mechanism (confirmed, with the real incident)

The fix-implementer produces whole-file bodies from an LLM, commits+pushes them to a
`fix/*` branch, then a k8s Job runs the build gate. The gate's **first** step is
`gofmt -l <changed .go files>` and it fails the gate (exit 1, no PR) if any file is
unformatted:

- `platform/orchestration/actions/diagnose_build_gate_action.go:234-238` —
  `buildGateScript`: `UNFORMATTED=$(gofmt -l …); if [ -n "$UNFORMATTED" ]; then echo
  "gofmt FAILED for: $UNFORMATTED"; exit 1; fi`.

**The gate is correct — leave it.** "No PRs for broken code" is its whole charter
(`registry.go:1192`); `gofmt -l` as a read-only verifier is right. The defect is
**upstream**: the commit-prep step commits the LLM's bytes verbatim without formatting
them first.

- `platform/orchestration/actions/diagnose_prepare_fix_commit_action.go` —
  `validateImplementation` returns the `files` map (path → whole-file content) at **L150**;
  payload assembly (branch name, commit data) begins at **L166**. Between those two points
  the `.go` file contents are never passed through `gofmt`. Whatever the LLM emitted is
  what gets committed and gated.

**Real incident (the proof):** BUG A implementer run `70680566` (2026-07-18) generated
logically-correct guards in BOTH `platform/aiservice/anthropic.go` and `ollama.go`
(matching the council-APPROVED plan byte-for-byte in intent), pushed `fix/e505f70f`, and
the gate failed with `gofmt FAILED for: platform/aiservice/anthropic.go`. The sole cause:
the LLM inserted a new `StopReason string` field but did not re-align the sibling
`Usage struct` field, and left a trailing blank line — two whitespace edits `gofmt -w`
fixes instantly. `ollama.go` in the same run was already `gofmt`-clean. Run spent, no PR;
hand-finished onto branch 085 as commit `f32b208e5`.

## 2. The fix (sketch — the fixing thread owns the final shape)

Format the generated Go in-memory at write-time, so the gate becomes defence-in-depth
rather than the first line to catch trivia.

In `diagnose_prepare_fix_commit_action.go`, immediately after `validateImplementation`
(L150) and before payload assembly (L166), run each `.go` file's content through
`go/format`:

```go
import "go/format"
// …
for path, body := range files {
    if strings.HasSuffix(path, ".go") {
        formatted, ferr := format.Source([]byte(body))
        if ferr != nil {
            // gofmt-unparseable == the LLM emitted broken Go (often a max_tokens
            // truncation). Fail LOUD here with the path — do NOT commit it and do
            // NOT silently fall back to the raw body.
            return nil, fmt.Errorf("generated %s is not valid Go (cannot format — likely truncated): %w", path, ferr)
        }
        files[path] = string(formatted)
    }
}
```

Notes for the fixing thread:
- `format.Source` is exactly what `gofmt` runs; this makes the committed bytes identical
  to what the gate's `gofmt -l` expects, so the whitespace class can never fail the gate.
- The parse-error branch is a **bonus catch**: a truncated/garbled whole-file body now
  fails at commit-prep with a precise message instead of at `go build` in the Job (or,
  worse, passing gofmt and failing later) — cheaper and clearer. Note `prepare_fix_commit`
  already has a sibling truncation guard on the JSON envelope (L132-135); this extends the
  same "fail loud on truncation" posture to the file *bodies*.
- Keep the gate's `gofmt -l` step unchanged — belt and braces.
- **Unit test** (`diagnose_prepare_fix_commit_test.go` has precedent): feed a valid-but-
  unformatted `.go` body (e.g. misaligned struct fields), assert the returned `files`
  entry is `gofmt`-clean; feed a truncated body, assert the loud error.

## 3. Ship it
Go change → inert until a chassis image ships. Commit per task; build `agent-chassis`
from the committed ref; verify against the running pod (e.g. grep the binary for a symbol
from your change). No behavioural DB/config change needed.

## 4. Related (do not conflate)
- **008 / BUG A** (`stop_reason` undecoded) is the bug this incident was implementing the
  fix FOR — that fix is committed (`f32b208e5` on branch 085), not yet deployed. THIS item
  is about the implementer *machinery* that tripped while shipping it, not about BUG A.
- The truncation-detection theme recurs (008, 012, and the parse-error branch above): a
  whole-file body that won't `gofmt` is frequently a `max_tokens` truncation — the same
  root class 008 closes at the LLM client boundary. This item catches the *consequence* one
  layer up, in the implementer.
