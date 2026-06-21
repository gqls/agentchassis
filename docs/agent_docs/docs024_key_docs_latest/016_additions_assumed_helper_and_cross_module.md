# 016 additions — assumed-helper / cross-module copy build failures (2026-06-17)

Paste these two pieces into your `016_debugging_guide_v2_*.md`. The first is a
one-line addition to **§0 (Assumption Checklist)**; the second is a new entry for
**§9 (Specific Failure Patterns)**, sitting naturally beside the existing "New JSON
walker silently returns nothing" / "Action reads a complex/array workflow input but
gets nothing" entries.

---

## ADD to §0 — Before You Change Anything (Assumption Checklist)

- **Authoring a new action? Do NOT assume a `datahelpers` (or any shared-package)
  function exists or is named what you'd expect.** `grep -rn 'func <Name>'
  platform/orchestration/datahelpers/` BEFORE you call it. The package has
  `ExtractNestedField` (path→interface{}), `ExtractNestedFieldString/Map/Int`,
  `GetStringField`, `GetIntField`, and `ExtractStringListHelper`/`ToStringSlice`
  for slices — but NOT an `ExtractStringSlice`. A `go build` of the whole binary is
  the only thing that proves the call resolves; a per-package build of the new
  code is faster feedback (`go build ./platform/orchestration/actions/`).

---

## ADD to §9 — Specific Failure Patterns

### `undefined: datahelpers.<X>` — a new action calls a helper that doesn't exist

**Symptom.** The binary build fails (not the package the helper lives in) with e.g.:

```
platform/orchestration/actions/diagnose_assemble_bundle_action.go:122:23: undefined: datahelpers.ExtractStringSlice
```

The line number is in the NEW action, and the undefined symbol is a `datahelpers`
(or other shared-package) function. The action compiled fine in whatever scratch
context it was drafted in, because nothing there checked the call against the real
package.

**Cause.** The action was written against an *assumed* helper API. `datahelpers`
has a specific, finite set of extractors; a plausible-sounding name
(`ExtractStringSlice`, `GetStringSlice`, …) may simply not be one of them. The real
ones, confirmed by grep:

- value at a dotted path → `ExtractNestedField(data, path) interface{}`
- …as string / map / int → `ExtractNestedFieldString` / `…Map` / `…Int`
- a config string with default → `GetStringField(m, key, default)`,
  `GetIntField(m, key, default)`
- coerce a value to `[]string` → `ExtractStringListHelper(val interface{}) []string`
  (handles `[]interface{}` AND `[]string`, returns nil otherwise) or
  `ToStringSlice(items []interface{}) []string`

There is no single "extract path → []string" helper; you COMPOSE:

```go
// WRONG (assumed): datahelpers.ExtractStringSlice(collected, path)
// RIGHT (real):
scope := datahelpers.ExtractStringListHelper(
    datahelpers.ExtractNestedField(collected, path),
)
```

`ExtractStringListHelper(nil)` returns nil, so a `len(scope) == 0` fallback still
fires correctly.

**Diagnose / fix.**

```bash
# 1. confirm what actually exists for the job
grep -rnE 'func (ExtractNestedField|ExtractStringListHelper|ToStringSlice|GetStringField|GetIntField)' \
  platform/orchestration/datahelpers/

# 2. read the chosen helper's BODY before using it (don't assume its nil/type behaviour)
awk '/^func ExtractStringListHelper/,/^}/' platform/orchestration/datahelpers/data_helpers.go

# 3. swap the assumed call for the real composition, then prove it resolves
go build ./platform/orchestration/actions/
```

**Prevention.** When adding several actions at once, grep EVERY `datahelpers.*` call
they make against the real package in one pass, before building — the misses are
cheap to find together and expensive to find one build at a time:

```bash
# every shared-package call the new actions make:
grep -rhoE 'datahelpers\.[A-Za-z]+' platform/orchestration/actions/diagnose_*.go | sort -u
# then eyeball each against `grep 'func <Name>' .../datahelpers/`
```

### Related: cross-module copy leaves a package un-buildable (import path, missing file, stale sibling)

When a Go package is copied across a module boundary (e.g. a tested package moved
from a standalone module into the chassis under `pkg/<name>/`), the build can fail
in THREE distinct ways, each surfacing only on the first build in the *target*
module — never in the source, where the file exists and the path is right:

1. **Wrong import path on a copied file.** `package contextkit/internal/analysis is
   not in std` — a file still imports the source module's path. Fix: rewrite to the
   target module path (`grep '^module ' go.mod` for the prefix), e.g.
   `sed -i 's#"contextkit/internal/analysis"#"github.com/gqls/agentchassis/internal/analysis"#' <file>`.
2. **A file omitted entirely.** `undefined: DecideStep / StepInput / NewCallGraphFromJSON`
   — these were all in `step.go`, which wasn't copied. Fix: copy the missing file
   (with the same import-path rewrite if it imports the moved package).
3. **A stale version of a sibling that IS present.** `undefined: bestEffortConclusion`
   — `loop.go` was copied at a pre-refactor version that predates the helper a
   newer `step.go` calls. Fix: bring the sibling up to the version the new files
   were built against.

**Prevention.** Copy the WHOLE package directory as one unit from the validated
source, then apply the single import-path rewrite — don't cherry-pick files across
versions. And `ls` BOTH directories and diff the file LIST, not just spot-check
imports:

```bash
diff <(ls <source>/internal/<pkg>) <(ls <target>/pkg/<pkg>)
go build ./pkg/<pkg>/   # per-package build = fast confirmation before the full binary
go test  ./pkg/<pkg>/   # then prove behaviour came across, not just compilation
```
