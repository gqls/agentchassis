# RUNBOOK — chrome selection on the page-build path (`bugs_closed/167`)

Every command here had to be got right once. The gotcha is attached to each.

## Which component actually serves a chrome slot, per predicate

The two predicates can agree for months and then diverge on one `UPDATE`. Run
**both**, never one — the whole of 167 is that they were assumed to differ and did
not, and the whole of the *original* bug is that they can.

```sql
-- GetComponentByFunction: the SECTION-shaped lookup (no component_level filter)
SELECT name, component_level FROM content_components
WHERE function = 'site-header' AND is_active = true AND forked_from IS NULL
ORDER BY name LIMIT 1;

-- ResolveChromeComponent: the CHROME predicate, eligible-first
SELECT name, component_level,
       (is_active AND forked_from IS NULL
        AND component_level IN ('site','header','footer','head')) AS eligible
FROM content_components
WHERE function = 'site-header'
ORDER BY (is_active AND forked_from IS NULL
          AND component_level IN ('site','header','footer','head')) DESC, name
LIMIT 1;
```

**Gotcha:** repeat for `site-footer` and `head`. `head` is the one that behaves
differently — the first query returns **no row** (both candidates are
`is_active=false`) while the second returns `Document Head`, a
`component_level='section'` component, with `eligible=false`. A caller that
mistakes "a row came back" for "this is chrome" renders an 8.5KB page section as
`<head>`.

## The whole candidate pool at a glance

```sql
SELECT function, name, component_level, is_active, forked_from IS NOT NULL AS is_fork,
       length(html_template) AS len,
       (is_active AND forked_from IS NULL
        AND component_level IN ('site','header','footer','head')) AS chrome_eligible
FROM content_components
WHERE function IN ('site-header','site-footer','head')
ORDER BY function, chrome_eligible DESC, name;
```

**Gotcha:** `ORDER BY name` is the live tie-break in the *old* code path, and it is
what makes today's answer correct by accident — `header-theme-chrome` sorts before
`site-header`, `footer-theme-chrome` before `site-footer`. Sort the output by
`name` when you want to see what the old predicate would pick.

## Which sites take which branch

`RenderHeader`/`RenderFooter` consult the style collection **first**. A site whose
collection pins chrome never reaches the by-function branch at all.

```sql
SELECT s.domain, s.status, sc.name AS collection,
       sc.header_component_id IS NOT NULL AS pinned_header,
       h.name AS pinned_to, h.is_active, h.component_level,
       h.forked_from IS NOT NULL AS is_fork
FROM sites s
LEFT JOIN style_collections sc ON s.style_collection_id = sc.id
LEFT JOIN content_components h ON h.id = sc.header_component_id
ORDER BY s.domain;
```

**Gotcha:** a `LEFT JOIN` here, not an inner one. Every per-site `collection-*` row
has `header_component_id IS NULL`, and an inner join silently drops exactly the
sites that take the branch you are investigating — the 10 that matter. This is the
query that found `bugs_open/170`: the four sites that *do* pin are pinned to a
deactivated component or a fork.

## When the two chrome components last moved

```sql
SELECT name, component_level, is_active, updated_at::timestamp(0)
FROM content_components
WHERE function IN ('site-header','site-footer') AND forked_from IS NULL
ORDER BY updated_at DESC;
```

**Gotcha:** this is the query that dates a bug file's claim. Several rows sharing
one `updated_at` **to the second** is a bulk operation by another lane, not
coincidence — that is how the stale blast-radius table in 167 was explained
(`12:39:53`, 118's fleet repoint).

## Proving the guard scans can actually fail

A guard that has never been seen red is being trusted on faith.

```bash
cat > platform/orchestration/actions/zz_induce_tmp.go <<'GO'
package actions

import (
	"context"

	"go.uber.org/zap"
)

func induceTmp(ctx context.Context, db interface{}, logger *zap.Logger) (*Component, error) {
	return GetComponentByFunction(ctx, db, "site-header", logger)
}
GO
go test -run TestNoBuildPathResolvesChromeByPlainFunctionLookup ./platform/orchestration/actions/
rm -f platform/orchestration/actions/zz_induce_tmp.go
go test -run TestNoBuildPathResolvesChromeByPlainFunctionLookup ./platform/orchestration/actions/
```

Expect FAIL naming `zz_induce_tmp.go:11`, then `ok`. **Gotcha:** the file must be
valid Go and in the package, or you get a build error that looks like a caught
defect but is not.

## Testing on a shared tree that other sessions are breaking

`go test ./platform/orchestration/actions/` failed under me with nine
`undefined: uuid/context/sql/zap/json` errors in `prune_floor.go` — a file this
lane never touched, being extended by another session mid-edit.

```bash
T=$(mktemp -d)
git archive HEAD | tar -x -C "$T"
cp platform/orchestration/actions/component_library.go       "$T/platform/orchestration/actions/"
cp platform/orchestration/actions/chrome_build_path_test.go  "$T/platform/orchestration/actions/"
cd "$T" && go build ./platform/orchestration/actions/ && go test ./platform/orchestration/actions/
```

**Gotcha:** this is the only run that means anything here. A green local build is
not a green HEAD, and — the direction that catches people — **a red local build is
not necessarily your fault.** Check `git status` on the failing file before
debugging it.

## Proving the fix reached production

```bash
kubectl get pods -n ai-persona-system -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.metadata.name}{"  "}{.spec.containers[0].image}{"\n"}{end}'

kubectl exec -n ai-persona-system <pod> -- sh -c '
  echo -n "header (NOTE the capital N): ";
  strings /app/agent-chassis | grep -c "No eligible header component in the library";
  echo -n "footer:                      ";
  strings /app/agent-chassis | grep -c "No eligible footer component in the library";
  echo -n "head   (lower-case n here!): ";
  strings /app/agent-chassis | grep -c "RenderHead: no eligible head component in the library";
  echo -n "NEGATIVE control, old string (must be 0): ";
  strings /app/agent-chassis | grep -c "No header component found, using fallback";
  echo -n "POSITIVE control (118): ";
  strings /app/agent-chassis | grep -c "no component serves chrome function"'
```

Ran on `v1.0.1225`, both replicas, 2026-07-31 23:1x UTC: **1 / 1 / 1 / 0 / 1** — the fix
is LIVE.

**Gotchas, and the first one bit me on this exact command:**
- ⚠ **CASE. The three log lines are NOT spelled alike.** `RenderHeader`/`RenderFooter`
  begin their message with a capital `No eligible …`; `RenderHead` begins with
  `RenderHead: no eligible …`, lower-case, because the sentence continues after the
  prefix. My first verification grepped lower-case for all three, and got
  **`header: 0, footer: 0, head: 1`** on a binary that contained all three. A
  case-mismatched grep returns a **false absence that is indistinguishable from a
  genuine one**, and the positive control did NOT catch it because the control tested a
  *different string* — a control only proves the pipeline works, never that your pattern
  is spelled right. Use `grep -ic`, or paste the literal from the source.
- **Add a NEGATIVE control**: grep for a string your change *removed*
  (`No header component found, using fallback` → must be 0). A positive control proves
  `strings|grep` works; only a negative control proves the **old code is gone** rather
  than the new code merely being present alongside it.
- Run it on **every** replica. `logs deploy/X` and a single `exec` read one pod of N.
- The positive control must be in the **same exec**. Without it a `0` is ambiguous
  between "not shipped" and "grep is broken" (`bugs_open/153`).
- Use plain `grep -c`, never `-x` or `^`/`$`: the Go linker packs unrelated string
  constants into one contiguous rodata blob, so `strings` output does not put one
  constant per line and any anchored pattern tests where the blob happens to break.

## Council submission for this lane

```
SUBMISSION_CORR = d73a4b06-a190-426e-bdf7-18d830d06a9d
```

```sql
SELECT created_at, metadata->>'decision'
FROM diagnosis_artifacts
WHERE correlation_id = 'd73a4b06-a190-426e-bdf7-18d830d06a9d'
  AND kind = 'council_report'
ORDER BY created_at;
```

**Gotcha:** budget ~30 minutes, not ~2 — the council itself takes 2–5 minutes and
the dispatch queues behind the fleet. A missing row is latency, not a dropped
dispatch; do not retry on that evidence. Find the run by payload, never by the
printed id:

```sql
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id'
      = 'd73a4b06-a190-426e-bdf7-18d830d06a9d';
```
