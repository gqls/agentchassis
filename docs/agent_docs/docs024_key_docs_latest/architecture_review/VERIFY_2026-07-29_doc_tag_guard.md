# VERIFY — D12 option 1b guard (`doc` rows are not static-tier evidence)

Council `da1f9c81-0b4b-41ff-9b2c-bc0057ad3cf8` — **APPROVED**, 12 reviewers, 0
unreadable. `diagnosis_guardian`, the seat that raised D12, approved with **zero
objections**.

**Only 2 of the 3 edits are load-bearing** (conceding `editquality`, low): this file
changes no runtime behaviour and is not counted as covering any mechanism.

---

## 1. The binary — nothing else can stand in for it

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis \
      -o jsonpath='{.items[0].metadata.name}')
for s in 'They are CODE' 'docTag' 'answerCodeCheck'; do
  printf '%-20s %s\n' "$s" \
    "$(kubectl -n ai-persona-system exec $POD -- sh -c "strings /app/agent-chassis | grep -c '$s'")"
done
```

**MEASURED PRE-CHANGE (2026-07-29, live pod, before the roll):**

| string | before | after must be | why |
|---|---|---|---|
| `They are CODE` | **1** | **0** | the string this change DELETES |
| `docTag` | **0** | **>0** | the symbol it ADDS |
| `answerCodeCheck` | **8** | **>0** | positive control — proves the grep works |

**A deleted string going to 0 cannot be faked by a stale cached layer**, the way a
missing added string is indistinguishable from a bad grep. Re-run after ANY roll you
did not perform yourself — this fleet rolls hourly.

**Pre-merge check demanded by `guardian` (low), RUN AND CLEAR:** nothing outside the
edited line asserts on the literal `They are CODE` —
`grep -rn "They are CODE" --include=*.go .` returns only
`diagnose_assemble_bundle_action.go:268` itself. No test, no third renderer.

## 2. Code rows are unchanged — the negative control

```sql
SELECT kind, count(*) FROM code_symbols GROUP BY kind;   -- expect ZERO 'doc'
```

With no doc rows, every tagging branch is dead and a code answer must render
**byte-identically** to before. `go test ./platform/orchestration/actions/...
./pkg/diagnose/...` passes (4 packages, 2026-07-29).

**The `ls` trap, caught during implementation and worth keeping:** adding `kind` to
`SELECT DISTINCT path, commit_sha` would have returned **one row per (path, kind)**
and multiplied every Go file in an `ls` listing. It is a `GROUP BY` with
`bool_or(kind = ANY(string_to_array($4,',')))` instead — a path is code if ANY row
under it is code. The kind list has **one source of truth** (`codeKindList` builds
both the Go map and the bound CSV), so the two cannot drift.

## 3. The tagging is an ALLOW-LIST, not a denylist of one

Conceding `bug_historian` (medium), which was right: `if kind == "doc"` would tag
doc and silently default **every other value — including a kind nobody has invented
yet — to "render as code, untagged"**. That is the dispatch-table-`default:` shape
that recurs in this codebase. Inverted:

```go
var codeKindList = []string{"func","method","struct","interface","alias","type","var","const"}
func docTag(kind string) string { if codeKinds[kind] { return "" }; return " [doc]" }
```

An unknown kind now renders as NOT-code — wrong in the direction that costs a
reviewer one wasted lookup, never a CONFIRMED verdict resting on prose.

## 4. What this file does NOT prove — read this before citing it as coverage

**It proves the guard is PRESENT and HARMLESS, not that it WORKS.** Nothing can
exercise it until `kind='doc'` is representable, which needs the markdown plan's
migration. That proof belongs to corr `7ba5b8c4`'s VERIFY, and per the
`architecture` seat (low) it must include a **NEGATIVE CONTROL, not just a presence
check**:

```sql
-- 4a. presence: doc rows arrive and carry the tag
--     (grep the rendered bundle for "[doc]" on a known markdown section)
-- 4b. NEGATIVE CONTROL, the one the seat asked for:
--     NO kind='doc' row may render outside the prose block or without [doc].
SELECT count(*) FROM code_symbols WHERE kind='doc';       -- > 0 by then
-- then, in a real diagnosis bundle:
--   every line matching a doc path MUST carry [doc]      -- 0 exceptions
--   every such line MUST fall after the docBlockHeader   -- 0 exceptions
```

A presence check passes if ONE doc row is tagged. The negative control is what
fails if any row is missed — which is the failure this whole guard exists to stop.

## 5. Noted, not actioned

Three seats (`guardian`, `prior_art_librarian` ×2) independently reported they
**cannot verify `code_symbols` at all** — it is absent from the schema hint the
council is given, so the plan's central safety claim ("every branch is dead on
arrival") had to be taken on the submitter's word. That is a real gap in council
tooling, not in this change, and it is the same class of limit the `architecture`
seat named about its own tier. Worth a schema-hint addition; filed nowhere yet.

`tooling_provenance` (low) noted no `doc_notes` breadcrumb is written for the
subjects `diagnose_assemble_bundle_action` / `diagnose_code_lookup_action`, so the
next fixer has only this repo markdown. Accepted for now — the OWNER RULING + this
directory are the provenance store for architecture-scope decisions — but the
equivalence is asserted, not established.
