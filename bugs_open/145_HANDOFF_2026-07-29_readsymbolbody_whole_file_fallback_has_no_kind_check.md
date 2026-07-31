# 145 — `ReadSymbolBody`'s whole-file fallback reads ANY path off disk, with no kind check

> ## STATUS 2026-07-31 (session "bugfix 12") — FIXED, COUNCIL-APPROVED, COMMITTED. **Stays OPEN for one reason: not yet live.**
>
> **Fix:** `691c1817a` — the analyser `Output` is now the READ BOUNDARY, not just the span
> source: an unconditional `findFile` runs BEFORE `os.ReadFile`, for the whole-file branch as
> well as the symbol branch. Council gate **APPROVED at round 1** (corr
> `bce4caab-17b6-4bbb-ba6b-d1e18f196156`, 2 low advisories, none high). Register **CTXK-002**
> carries the seam, its landmine and two answered verify-laters. Workstream:
> `docs024_key_docs_latest/bugfix_145_symbolbody_boundary/`.
>
> **WHY IT IS STILL OPEN — the defect is still reproducible in production.** A Go change is
> inert until an image is rebuilt and rolled. The running chassis is `v1.0.1215`, pods started
> **2026-07-31 15:00:20Z**, i.e. **two hours before the fix commit**. Per CLAUDE.md's closing
> bar ("fixed AND live"), that keeps this OPEN. **Close it — and move it to `bugs_closed/` —
> when this passes on the running pod** (a positive control in the same exec, because a roll is
> not evidence your fix shipped, `bugs_open/153`):
>
> ```bash
> P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name | head -1)
> kubectl -n ai-persona-system exec $P -- sh -c \
>   'strings /app/agent-chassis | grep -c "bodies are read only for files the analyser parsed"; \
>    strings /app/agent-chassis | grep -c "not in analysis (no FileInfo for that path)"'
> ```
> Expect **≥1** for the first (the string this fix ADDED — the new suffix on the refusal) and
> **≥1** for the second (the positive control: the pre-existing part of the same message, which
> proves the grep and the binary are both sound). A `0` then `≥1` means the image predates the
> fix; `0` then `0` means the probe is broken, not the fix.
>
> **Not built/rolled by this session, deliberately:** another session had `IMAGE_TAG`
> `v1.0.1215` bumped uncommitted (mid-build), and at the time of writing two council review
> steps belonging to another round were EXECUTING — a roll kills an in-flight council. The
> commit is in `HEAD`, so the next `make build-agent-chassis` by anyone carries it.
>
> **Spun out, do not re-file:** `bugs_open/164` (the same loop's `maxBodyChars` `break`
> silently dropping the rest of scope — filed at the council's request) and `bugs_open/163`
> (the landmine verifier, found while filing this bug's landmine).
>
> **Two claims in the filing below are WRONG — corrected in the body, but read this first:**
> (1) the bare-path branch is **NOT** "currently unreachable" — `route.scope` carries the LLM
> verdict's `next_scope` verbatim and the bundle *advertises* bare paths at `:597`, which
> **kills fix candidate 1**; (2) the leak was **not** bounded by the repo — a file outside the
> checkout came back. Both measured.

**Filed 2026-07-29. OPEN, UNOWNED. Pre-existing — reachable today, independent of the
change that surfaced it.**

Found by the council gate's **`architecture` seat** (medium) while reviewing D11
layer 1b round 10 (corr `7ba5b8c4`). Its objection, verbatim:

> "The fix filters kind at the two lookup queries but leaves `ReadSymbolBody`'s
> whole-file fallback (empty ':Symbol' part -> entire file read from disk, no kind
> check) intact as a **generic hazard reachable by any future scope-entry producer**,
> not just this one."

It is right, and this is the seat working exactly as designed: the layer-1b fix
closed *its own* instance at the query source; the seat asked what happens to the
*next* producer.

## The mechanism

`internal/analysis/symbolbody.go:44-59`:

```go
func ReadSymbolBody(root string, out Output, symbol string) (string, error) {
    pathPart, namePart := splitSymbol(symbol)
    if pathPart == "" { return "", fmt.Errorf(...) }

    abs := filepath.Join(root, filepath.FromSlash(pathPart))
    src, err := os.ReadFile(abs)            // ← DISK, not the analyser Output
    if err != nil { return "", ... }

    // Whole-file scope entry.
    if namePart == "" {
        return string(src), nil             // ← the ENTIRE FILE, unbounded, unchecked
    }
    ...
}
```

Two properties combine badly:

1. **It reads from DISK, not from the analyser `Output`.** The Output is Go-only
   (`internal/analysis/analyse.go:91` — `if !strings.HasSuffix(path, ".go")`), but
   the checkout on disk is the **whole tarball with no path filter**
   (`internal/reposource/github_source.go:196-245` writes every `tar.TypeReg`).
   So `bugs_open/*.md`, `.env`-shaped files, fixtures, anything committed, is
   readable through this function.
2. **A scope entry with no `:Symbol` part returns the whole file**, with no size
   bound of its own.

The consumer renders whatever comes back as Go:

```go
// diagnose_assemble_bundle_action.go
fmt.Fprintf(&b, "### %s\n```go\n%s\n```\n\n", sym, body)   // under "## In-scope code"
```

**So any producer that can put a bare path into `scope` can get an arbitrary
repository file rendered to the verdicter inside a ```go fence, labelled in-scope
code.** The only current brake is `maxBodyChars` (default 60,000) truncating the
bundle *after* the read.

## Why it is not currently firing

Scope entries today come from `diagnose_route` (path:Symbol handles),
a caller-supplied seed, or `scopeFromCodeResults`
(`diagnose_assemble_bundle_action.go:596`), which only emits a bare path when a row
has a `path` and an empty `symbol`:

```go
if path != "" && symbol != "" { out = append(out, path+":"+symbol) } else if path != "" { out = append(out, path) }
```

`code_symbols.symbol` is `NOT NULL` and the indexer never writes an empty one, so
the bare-path branch is currently unreachable **by accident of the writer, not by
construction of the boundary**. That is the defect: the safety property lives in a
producer, not in the function that would be harmed.

## Why layer 1b did NOT fix it here

Layer 1b closed its own exposure at the query source (`lookup_code_symbols` now
returns only code kinds, reusing the D12 guard's allow-list). That is correct and
sufficient **for that producer**. Fixing the generic fallback changes the shared
helper itself, for every future producer — a different blast radius, a different
review. **Splitting it out is legitimate precisely because it is pre-existing and
not layer 1b's to create** (contrast the D12 citation hazard, which layer 1b *would*
have created and therefore could not be split).

> **CORRECTED 2026-07-29, before this file was an hour old.** The first version of
> this section said the helper is "used by `cmd/assembler` and the bundle action
> alike". **`cmd/assembler` does not exist in this repo.** The only live caller of
> `ReadSymbolBody` is `diagnose_assemble_bundle_action.go:201`; the assembler lives
> in archived reference code at
> `docs/agent_docs/docs019_…/go_files/contextkit/cmd/assembler/main.go`.
>
> **So the blast radius is ONE call site, not two — the fix is cheaper than this
> file first claimed.** I took the claim from the code comment at
> `diagnose_assemble_bundle_action.go:641` ("so `cmd/assembler` and this action
> slice IDENTICALLY"), **which is itself stale** and should be corrected by whoever
> takes this bug.
>
> Caught by `scripts/pattern-check.py`'s `new-capability-surface` check, which
> flagged that the file "proposes `cmd/assembler/`, which does not exist" — a check
> written for a different purpose (spotting a second copy of a service) that
> happened to catch a stale citation. **Worth recording as a fire that earned its
> keep**, and as the third instance today of the same root error: repeating a claim
> from a doc or comment without running the one command that checks it.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Refuse the whole-file branch unless the caller opts in.** `ReadSymbolBody` is
   named for reading a *symbol*; returning an entire file on an empty name is a
   surprising overload. Either drop the branch or move it to an explicit
   `ReadWholeFile` the caller must choose. **Makes the bad state unrepresentable.**
2. **Bound it by extension.** Only slice/return files the analyser could have
   parsed (`.go`), since the Output is the thing giving the spans meaning. Cheap,
   and it aligns the function with its only real input.
3. **Bound it by size** at the read, not at the bundle. Weakest — it caps the damage
   without removing the class, and "the operator must remember `maxBodyChars`" is
   the shape this repo already calls a defect.

## How to verify a fix

- A scope entry of a bare non-Go path (e.g. `bugs_open/145_….md`) must NOT appear
  under `## In-scope code`, and must not be rendered in a ```go fence.
- A normal `path.go:Symbol` entry must still slice identically — diff a bundle
  before/after on a real diagnosis.
- Negative control: a bare `.go` path must behave as the caller currently expects,
  or the change is a silent behaviour break for `cmd/assembler`.

## Related

- `bugs_closed/108` — the same file's body-slicing, from the freshness angle.
- D12 / corr `da1f9c81` — the citation guard that makes `[doc]` rows distinguishable
  once they exist; it does not cover this path, because this path never consults
  `kind` at all.
- `architecture_review/NOTES_architecture_seat.md`, 2026-07-29 — the trace that
  found it, including why the round-9 HIGH's stated mechanism was refuted while its
  instinct was right.
